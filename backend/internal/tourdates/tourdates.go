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

// List returns the caller's own tourdates plus every tourdate belonging to
// an artist they represent, merged into one list ordered by date. If an
// "artist_id" query param is given, the list is narrowed to just that
// artist - who must be the caller or someone they represent, else a 404 (so
// as not to reveal whether that artist_id exists at all).
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	caller := auth.UserFromContext(r.Context())
	accessibleIDs, err := AccessibleArtistIDs(r.Context(), h.DB, caller.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list tourdates")
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

	var rows pgx.Rows
	if artistIDParam := r.URL.Query().Get("artist_id"); artistIDParam != "" {
		artistID, err := uuid.Parse(artistIDParam)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "artist_id must be a valid UUID")
			return
		}
		if !ContainsID(accessibleIDs, artistID) {
			httpx.WriteError(w, http.StatusNotFound, "artist not found")
			return
		}
		rows, err = h.DB.Query(r.Context(),
			"SELECT "+tourDateColumns+" FROM tourdates WHERE user_id = $1 AND ($2::uuid IS NULL OR tour_id = $2) AND ($3::uuid IS NULL OR rider_id = $3) ORDER BY date",
			artistID, tourIDFilter, riderIDFilter)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "failed to list tourdates")
			return
		}
	} else {
		rows, err = h.DB.Query(r.Context(),
			"SELECT "+tourDateColumns+" FROM tourdates WHERE user_id = ANY($1::uuid[]) AND ($2::uuid IS NULL OR tour_id = $2) AND ($3::uuid IS NULL OR rider_id = $3) ORDER BY date",
			accessibleIDs, tourIDFilter, riderIDFilter)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "failed to list tourdates")
			return
		}
	}
	defer rows.Close()

	tourdates := []TourDate{}
	for rows.Next() {
		td, err := scanTourDate(rows)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "failed to read tourdates")
			return
		}
		tourdates = append(tourdates, td)
	}
	if err := rows.Err(); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to read tourdates")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, tourdates)
}

// Get 404s (rather than 403s) when the tourdate belongs to an artist the
// caller doesn't represent (and isn't themselves), so as not to reveal that
// it exists.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	caller := auth.UserFromContext(r.Context())
	accessibleIDs, err := AccessibleArtistIDs(r.Context(), h.DB, caller.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to read tourdate")
		return
	}

	row := h.DB.QueryRow(r.Context(),
		"SELECT "+tourDateColumns+" FROM tourdates WHERE id = $1 AND user_id = ANY($2::uuid[])", id, accessibleIDs)
	td, err := scanTourDate(row)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusNotFound, "tourdate not found")
		return
	} else if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to read tourdate")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, td)
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

	caller := auth.UserFromContext(r.Context())
	artistID := req.ArtistID
	if artistID == uuid.Nil {
		artistID = caller.ID
	} else if artistID != caller.ID {
		accessibleIDs, err := AccessibleArtistIDs(r.Context(), h.DB, caller.ID)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "failed to create tourdate")
			return
		}
		if !ContainsID(accessibleIDs, artistID) {
			httpx.WriteError(w, http.StatusForbidden, "you do not have access to create tourdates for this artist")
			return
		}
	}

	var tourID *uuid.UUID
	if req.TourID != uuid.Nil {
		var tourOwnerID uuid.UUID
		err := h.DB.QueryRow(r.Context(), "SELECT user_id FROM tours WHERE id = $1", req.TourID).Scan(&tourOwnerID)
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.WriteError(w, http.StatusBadRequest, "tour_id references a nonexistent tour")
			return
		} else if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "failed to create tourdate")
			return
		}
		if tourOwnerID != artistID {
			httpx.WriteError(w, http.StatusBadRequest, "tour_id must belong to the same artist as this tourdate")
			return
		}
		tourID = &req.TourID
	}

	var riderID *uuid.UUID
	if req.RiderID != uuid.Nil {
		var riderOwnerID uuid.UUID
		err := h.DB.QueryRow(r.Context(), "SELECT user_id FROM riders WHERE id = $1", req.RiderID).Scan(&riderOwnerID)
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.WriteError(w, http.StatusBadRequest, "rider_id references a nonexistent rider")
			return
		} else if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "failed to create tourdate")
			return
		}
		if riderOwnerID != artistID {
			httpx.WriteError(w, http.StatusBadRequest, "rider_id must belong to the same artist as this tourdate")
			return
		}
		riderID = &req.RiderID
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

	var setClauses []string
	var args []any
	argPos := 1

	addSet := func(column string, value any) {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", column, argPos))
		args = append(args, value)
		argPos++
	}

	if v, ok := raw["date"]; ok {
		var d Date
		if err := json.Unmarshal(v, &d); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "date must be in YYYY-MM-DD format")
			return
		}
		addSet("date", time.Time(d))
	}
	if v, ok := raw["city"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil || s == "" {
			httpx.WriteError(w, http.StatusBadRequest, "city must be a non-empty string")
			return
		}
		addSet("city", s)
	}
	if v, ok := raw["venue"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil || s == "" {
			httpx.WriteError(w, http.StatusBadRequest, "venue must be a non-empty string")
			return
		}
		addSet("venue", s)
	}
	if v, ok := raw["tour_id"]; ok {
		if string(v) == "null" {
			addSet("tour_id", nil)
		} else {
			var tourID uuid.UUID
			if err := json.Unmarshal(v, &tourID); err != nil {
				httpx.WriteError(w, http.StatusBadRequest, "tour_id must be a valid UUID or null")
				return
			}

			// Look up the tourdate's current owner through the same
			// accessibleIDs filter as the final UPDATE, so an inaccessible
			// tourdate 404s here rather than leaking its existence via a
			// 400 further down.
			var currentOwnerID uuid.UUID
			err := h.DB.QueryRow(r.Context(),
				"SELECT user_id FROM tourdates WHERE id = $1 AND user_id = ANY($2::uuid[])", id, accessibleIDs,
			).Scan(&currentOwnerID)
			if errors.Is(err, pgx.ErrNoRows) {
				httpx.WriteError(w, http.StatusNotFound, "tourdate not found")
				return
			} else if err != nil {
				httpx.WriteError(w, http.StatusInternalServerError, "failed to update tourdate")
				return
			}

			var tourOwnerID uuid.UUID
			err = h.DB.QueryRow(r.Context(), "SELECT user_id FROM tours WHERE id = $1", tourID).Scan(&tourOwnerID)
			if errors.Is(err, pgx.ErrNoRows) {
				httpx.WriteError(w, http.StatusBadRequest, "tour_id references a nonexistent tour")
				return
			} else if err != nil {
				httpx.WriteError(w, http.StatusInternalServerError, "failed to update tourdate")
				return
			}
			if tourOwnerID != currentOwnerID {
				httpx.WriteError(w, http.StatusBadRequest, "tour_id must belong to the same artist as this tourdate")
				return
			}
			addSet("tour_id", tourID)
		}
	}
	if v, ok := raw["rider_id"]; ok {
		if string(v) == "null" {
			addSet("rider_id", nil)
		} else {
			var riderID uuid.UUID
			if err := json.Unmarshal(v, &riderID); err != nil {
				httpx.WriteError(w, http.StatusBadRequest, "rider_id must be a valid UUID or null")
				return
			}

			// Same accessibleIDs-filtered lookup as tour_id above, so an
			// inaccessible tourdate 404s here rather than leaking its
			// existence via a 400 further down.
			var currentOwnerID uuid.UUID
			err := h.DB.QueryRow(r.Context(),
				"SELECT user_id FROM tourdates WHERE id = $1 AND user_id = ANY($2::uuid[])", id, accessibleIDs,
			).Scan(&currentOwnerID)
			if errors.Is(err, pgx.ErrNoRows) {
				httpx.WriteError(w, http.StatusNotFound, "tourdate not found")
				return
			} else if err != nil {
				httpx.WriteError(w, http.StatusInternalServerError, "failed to update tourdate")
				return
			}

			var riderOwnerID uuid.UUID
			err = h.DB.QueryRow(r.Context(), "SELECT user_id FROM riders WHERE id = $1", riderID).Scan(&riderOwnerID)
			if errors.Is(err, pgx.ErrNoRows) {
				httpx.WriteError(w, http.StatusBadRequest, "rider_id references a nonexistent rider")
				return
			} else if err != nil {
				httpx.WriteError(w, http.StatusInternalServerError, "failed to update tourdate")
				return
			}
			if riderOwnerID != currentOwnerID {
				httpx.WriteError(w, http.StatusBadRequest, "rider_id must belong to the same artist as this tourdate")
				return
			}
			addSet("rider_id", riderID)
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
			addSet(column, nil)
			continue
		}
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, column+" must be a string or null")
			return
		}
		addSet(column, s)
	}

	if len(setClauses) == 0 {
		httpx.WriteError(w, http.StatusBadRequest, "no updatable fields provided")
		return
	}

	args = append(args, id, accessibleIDs)
	query := fmt.Sprintf(
		"UPDATE tourdates SET %s WHERE id = $%d AND user_id = ANY($%d::uuid[]) RETURNING "+tourDateColumns,
		strings.Join(setClauses, ", "), argPos, argPos+1,
	)

	row := h.DB.QueryRow(r.Context(), query, args...)
	td, err := scanTourDate(row)
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
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	caller := auth.UserFromContext(r.Context())
	accessibleIDs, err := AccessibleArtistIDs(r.Context(), h.DB, caller.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to delete tourdate")
		return
	}

	tag, err := h.DB.Exec(r.Context(),
		"DELETE FROM tourdates WHERE id = $1 AND user_id = ANY($2::uuid[])", id, accessibleIDs)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to delete tourdate")
		return
	}
	if tag.RowsAffected() == 0 {
		httpx.WriteError(w, http.StatusNotFound, "tourdate not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
