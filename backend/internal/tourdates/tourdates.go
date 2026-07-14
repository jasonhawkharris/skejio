package tourdates

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"skejio/backend/internal/auth"
	"skejio/backend/internal/httpx"
)

const dateLayout = "2006-01-02"

// Date wraps time.Time to marshal/unmarshal as a plain YYYY-MM-DD string,
// matching the Postgres DATE column (no time-of-day or timezone).
type Date time.Time

func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(d).Format(dateLayout))
}

func (d *Date) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return fmt.Errorf("date must be in YYYY-MM-DD format: %w", err)
	}
	*d = Date(t)
	return nil
}

type TourDate struct {
	ID             uuid.UUID  `json:"id"`
	Date           Date       `json:"date"`
	City           string     `json:"city"`
	State          *string    `json:"state"`
	Venue          string     `json:"venue"`
	POCName        *string    `json:"poc_name"`
	POCNumber      *string    `json:"poc_number"`
	POCEmail       *string    `json:"poc_email"`
	PromoterName   *string    `json:"promoter_name"`
	PromoterNumber *string    `json:"promoter_number"`
	PromoterEmail  *string    `json:"promoter_email"`
	UserID         uuid.UUID  `json:"user_id"`
	TourID         *uuid.UUID `json:"tour_id"`
	RiderID        *uuid.UUID `json:"rider_id"`
	CreatedAt      time.Time  `json:"created_at"`
}

type Handler struct {
	DB *pgxpool.Pool
}

func scanTourDate(row pgx.Row) (TourDate, error) {
	var td TourDate
	var d time.Time
	err := row.Scan(
		&td.ID, &d, &td.City, &td.State, &td.Venue,
		&td.POCName, &td.POCNumber, &td.POCEmail,
		&td.PromoterName, &td.PromoterNumber, &td.PromoterEmail,
		&td.UserID, &td.TourID, &td.RiderID, &td.CreatedAt,
	)
	td.Date = Date(d)
	return td, err
}

const tourDateColumns = "id, date, city, state, venue, poc_name, poc_number, poc_email, promoter_name, promoter_number, promoter_email, user_id, tour_id, rider_id, created_at"

// AccessibleArtistIDs returns the set of user ids whose tourdates the caller
// may access: their own id, plus every artist they represent (if any).
// Returned as strings so it can be bound directly to a ::uuid[] parameter.
func AccessibleArtistIDs(ctx context.Context, db *pgxpool.Pool, callerID uuid.UUID) ([]string, error) {
	ids := []string{callerID.String()}

	rows, err := db.Query(ctx, "SELECT artist_id FROM artist_representatives WHERE representative_id = $1", callerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var artistID uuid.UUID
		if err := rows.Scan(&artistID); err != nil {
			return nil, err
		}
		ids = append(ids, artistID.String())
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}

func ContainsID(ids []string, id uuid.UUID) bool {
	target := id.String()
	for _, existing := range ids {
		if existing == target {
			return true
		}
	}
	return false
}

// AccessibleTourDateID parses the "id" route param and confirms the caller
// may access that tourdate (their own, or one they represent) - 404ing
// otherwise, so as not to reveal whether the tourdate exists. Shared by any
// resource nested under /tourdates/{id}/... (financials, expenses).
func AccessibleTourDateID(w http.ResponseWriter, r *http.Request, db *pgxpool.Pool) (uuid.UUID, bool) {
	tourDateID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id")
		return uuid.Nil, false
	}

	caller := auth.UserFromContext(r.Context())
	accessibleIDs, err := AccessibleArtistIDs(r.Context(), db, caller.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to look up tourdate")
		return uuid.Nil, false
	}

	var ownerID uuid.UUID
	err = db.QueryRow(r.Context(), "SELECT user_id FROM tourdates WHERE id = $1", tourDateID).Scan(&ownerID)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusNotFound, "tourdate not found")
		return uuid.Nil, false
	} else if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to look up tourdate")
		return uuid.Nil, false
	}
	if !ContainsID(accessibleIDs, ownerID) {
		httpx.WriteError(w, http.StatusNotFound, "tourdate not found")
		return uuid.Nil, false
	}

	return tourDateID, true
}

// resolveOwnedFK validates a nullable owned-FK request field for Create:
// field is the JSON field name (e.g. "tour_id"), table is the referenced
// table (e.g. "tours") - which must have a user_id column. id of uuid.Nil
// means the field was omitted, returning (nil, true). Otherwise id must
// reference a row in table owned by ownerID.
func (h *Handler) resolveOwnedFK(ctx context.Context, w http.ResponseWriter, field, table string, id, ownerID uuid.UUID) (*uuid.UUID, bool) {
	if id == uuid.Nil {
		return nil, true
	}

	var actualOwnerID uuid.UUID
	err := h.DB.QueryRow(ctx, "SELECT user_id FROM "+table+" WHERE id = $1", id).Scan(&actualOwnerID)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusBadRequest, field+" references a nonexistent "+strings.TrimSuffix(field, "_id"))
		return nil, false
	} else if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create tourdate")
		return nil, false
	}
	if actualOwnerID != ownerID {
		httpx.WriteError(w, http.StatusBadRequest, field+" must belong to the same artist as this tourdate")
		return nil, false
	}

	return &id, true
}

// applyOwnedFKPatch handles a nullable owned-FK field (e.g. "tour_id",
// "rider_id") in a PATCH body via addSet: raw is that field's JSON value -
// literal null clears the column; otherwise it must be a UUID referencing a
// row in table (which must have a user_id column) owned by the same artist
// as the tourdate being patched. The tourdate's current owner is looked up
// through the same accessibleIDs filter as the final UPDATE, so an
// inaccessible tourdate 404s here rather than leaking its existence via a
// 400 further down.
func (h *Handler) applyOwnedFKPatch(
	ctx context.Context, w http.ResponseWriter, raw json.RawMessage, field, table string,
	tourdateID uuid.UUID, accessibleIDs []string, addSet func(string, any),
) bool {
	if string(raw) == "null" {
		addSet(field, nil)
		return true
	}

	var id uuid.UUID
	if err := json.Unmarshal(raw, &id); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, field+" must be a valid UUID or null")
		return false
	}

	var currentOwnerID uuid.UUID
	err := h.DB.QueryRow(ctx,
		"SELECT user_id FROM tourdates WHERE id = $1 AND user_id = ANY($2::uuid[])", tourdateID, accessibleIDs,
	).Scan(&currentOwnerID)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusNotFound, "tourdate not found")
		return false
	} else if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to update tourdate")
		return false
	}

	var actualOwnerID uuid.UUID
	err = h.DB.QueryRow(ctx, "SELECT user_id FROM "+table+" WHERE id = $1", id).Scan(&actualOwnerID)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusBadRequest, field+" references a nonexistent "+strings.TrimSuffix(field, "_id"))
		return false
	} else if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to update tourdate")
		return false
	}
	if actualOwnerID != currentOwnerID {
		httpx.WriteError(w, http.StatusBadRequest, field+" must belong to the same artist as this tourdate")
		return false
	}

	addSet(field, id)
	return true
}

func (h *Handler) resource() ScopedResource[TourDate] {
	return ScopedResource[TourDate]{DB: h.DB, Table: "tourdates", Columns: tourDateColumns, OrderBy: "date", Scan: scanTourDate}
}

// List returns the caller's own tourdates plus every tourdate belonging to
// an artist they represent, merged into one list ordered by date. If an
// "artist_id" query param is given, the list is narrowed to just that
// artist - who must be the caller or someone they represent, else a 404 (so
// as not to reveal whether that artist_id exists at all). "tour_id" and
// "rider_id" query params further narrow the list, if given.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	accessibleIDs, artistID, ok := ResolveListScope(w, r, h.DB, "tourdate")
	if !ok {
		return
	}

	var tourIDFilter *uuid.UUID
	if tourIDParam := r.URL.Query().Get("tour_id"); tourIDParam != "" {
		parsed, err := uuid.Parse(tourIDParam)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "tour_id must be a valid UUID")
			return
		}
		tourIDFilter = &parsed
	}

	var riderIDFilter *uuid.UUID
	if riderIDParam := r.URL.Query().Get("rider_id"); riderIDParam != "" {
		parsed, err := uuid.Parse(riderIDParam)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "rider_id must be a valid UUID")
			return
		}
		riderIDFilter = &parsed
	}

	tourdateList, err := h.resource().List(r.Context(), accessibleIDs, artistID,
		" AND ($2::uuid IS NULL OR tour_id = $2) AND ($3::uuid IS NULL OR rider_id = $3)",
		tourIDFilter, riderIDFilter)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list tourdates")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, tourdateList)
}

// Get 404s (rather than 403s) when the tourdate belongs to an artist the
// caller doesn't represent (and isn't themselves), so as not to reveal that
// it exists.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	HandleGet(w, r, h.resource(), "tourdate")
}

type createTourDateRequest struct {
	Date           Date      `json:"date"`
	City           string    `json:"city"`
	State          *string   `json:"state"`
	Venue          string    `json:"venue"`
	POCName        *string   `json:"poc_name"`
	POCNumber      *string   `json:"poc_number"`
	POCEmail       *string   `json:"poc_email"`
	PromoterName   *string   `json:"promoter_name"`
	PromoterNumber *string   `json:"promoter_number"`
	PromoterEmail  *string   `json:"promoter_email"`
	ArtistID       uuid.UUID `json:"artist_id"`
	TourID         uuid.UUID `json:"tour_id"`
	RiderID        uuid.UUID `json:"rider_id"`
}

// Create assigns ownership to artist_id if given - which must be the caller
// themselves or an artist they represent (403 otherwise) - or to the caller
// themselves if artist_id is omitted.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createTourDateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if time.Time(req.Date).IsZero() {
		httpx.WriteError(w, http.StatusBadRequest, "date is required")
		return
	}
	if req.City == "" {
		httpx.WriteError(w, http.StatusBadRequest, "city is required")
		return
	}
	if req.Venue == "" {
		httpx.WriteError(w, http.StatusBadRequest, "venue is required")
		return
	}

	artistID, ok := ResolveCreateOwner(w, r, h.DB, req.ArtistID, "tourdate")
	if !ok {
		return
	}

	tourID, ok := h.resolveOwnedFK(r.Context(), w, "tour_id", "tours", req.TourID, artistID)
	if !ok {
		return
	}
	riderID, ok := h.resolveOwnedFK(r.Context(), w, "rider_id", "riders", req.RiderID, artistID)
	if !ok {
		return
	}

	row := h.DB.QueryRow(r.Context(),
		"INSERT INTO tourdates (date, city, state, venue, poc_name, poc_number, poc_email, promoter_name, promoter_number, promoter_email, user_id, tour_id, rider_id) "+
			"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING "+tourDateColumns,
		time.Time(req.Date), req.City, req.State, req.Venue,
		req.POCName, req.POCNumber, req.POCEmail,
		req.PromoterName, req.PromoterNumber, req.PromoterEmail,
		artistID, tourID, riderID)
	td, err := scanTourDate(row)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create tourdate")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, td)
}

// Patch applies a partial update. A field omitted from the JSON body is left
// unchanged; a field present but set to null clears it (only valid for the
// nullable "state", "poc_*", and "promoter_*" columns). Ownership is not
// patchable. A tourdate belonging to an artist the caller doesn't represent
// (and isn't themselves) 404s, same as Get.
func (h *Handler) Patch(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	caller := auth.UserFromContext(r.Context())
	accessibleIDs, err := AccessibleArtistIDs(r.Context(), h.DB, caller.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to update tourdate")
		return
	}

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	var pb httpx.PatchBuilder

	if v, ok := raw["date"]; ok {
		var d Date
		if err := json.Unmarshal(v, &d); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "date must be in YYYY-MM-DD format")
			return
		}
		pb.Set("date", time.Time(d))
	}
	if v, ok := raw["city"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil || s == "" {
			httpx.WriteError(w, http.StatusBadRequest, "city must be a non-empty string")
			return
		}
		pb.Set("city", s)
	}
	if v, ok := raw["venue"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil || s == "" {
			httpx.WriteError(w, http.StatusBadRequest, "venue must be a non-empty string")
			return
		}
		pb.Set("venue", s)
	}
	if v, ok := raw["tour_id"]; ok {
		if ok := h.applyOwnedFKPatch(r.Context(), w, v, "tour_id", "tours", id, accessibleIDs, pb.Set); !ok {
			return
		}
	}
	if v, ok := raw["rider_id"]; ok {
		if ok := h.applyOwnedFKPatch(r.Context(), w, v, "rider_id", "riders", id, accessibleIDs, pb.Set); !ok {
			return
		}
	}

	// state and the POC/promoter contact fields are all nullable TEXT
	// columns with the same partial-update semantics: omitted leaves the
	// column unchanged, explicit null clears it.
	nullableStringColumns := []string{
		"state", "poc_name", "poc_number", "poc_email",
		"promoter_name", "promoter_number", "promoter_email",
	}
	for _, column := range nullableStringColumns {
		v, ok := raw[column]
		if !ok {
			continue
		}
		if string(v) == "null" {
			pb.Set(column, nil)
			continue
		}
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, column+" must be a string or null")
			return
		}
		pb.Set(column, s)
	}

	if pb.Empty() {
		httpx.WriteError(w, http.StatusBadRequest, "no updatable fields provided")
		return
	}

	td, err := h.resource().Update(r.Context(), &pb, id, accessibleIDs)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusNotFound, "tourdate not found")
		return
	} else if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to update tourdate")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, td)
}

// Delete 404s for a tourdate belonging to an artist the caller doesn't
// represent (and isn't themselves), same as Get/Patch.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	HandleDelete(w, r, h.resource(), "tourdate")
}
