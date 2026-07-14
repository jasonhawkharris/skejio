package riders

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"skejio/backend/internal/auth"
	"skejio/backend/internal/httpx"
	"skejio/backend/internal/tourdates"
)

// Rider is a reusable document (technical/hospitality requirements, etc.)
// an artist can attach to a tourdate. One-to-many with tourdates
// (tourdates.rider_id) - a single rider can be reused across many
// tourdates, since which one applies is usually determined by the
// prominence of the venue rather than being unique per date.
type Rider struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Handler struct {
	DB *pgxpool.Pool
}

const riderColumns = "id, name, content, user_id, created_at"

func scanRider(row pgx.Row) (Rider, error) {
	var rd Rider
	err := row.Scan(&rd.ID, &rd.Name, &rd.Content, &rd.UserID, &rd.CreatedAt)
	return rd, err
}

func (h *Handler) resource() tourdates.ScopedResource[Rider] {
	return tourdates.ScopedResource[Rider]{DB: h.DB, Table: "riders", Columns: riderColumns, OrderBy: "created_at", Scan: scanRider}
}

// List returns the caller's own riders plus every rider belonging to an
// artist they represent. If an "artist_id" query param is given, the list is
// narrowed to just that artist - who must be the caller or someone they
// represent, else a 404 (so as not to reveal whether that artist_id exists).
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	tourdates.HandleList(w, r, h.resource(), "rider")
}

// Get 404s (rather than 403s) when the rider belongs to an artist the caller
// doesn't represent (and isn't themselves), so as not to reveal that it
// exists.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	tourdates.HandleGet(w, r, h.resource(), "rider")
}

type createRiderRequest struct {
	Name     string    `json:"name"`
	Content  string    `json:"content"`
	ArtistID uuid.UUID `json:"artist_id"`
}

// Create assigns ownership to artist_id if given - which must be the caller
// themselves or an artist they represent (403 otherwise) - or to the caller
// themselves if artist_id is omitted. Same pattern as tourdates.Create.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createRiderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Name == "" {
		httpx.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Content == "" {
		httpx.WriteError(w, http.StatusBadRequest, "content is required")
		return
	}

	artistID, ok := tourdates.ResolveCreateOwner(w, r, h.DB, req.ArtistID, "rider")
	if !ok {
		return
	}

	row := h.DB.QueryRow(r.Context(),
		"INSERT INTO riders (name, content, user_id) VALUES ($1, $2, $3) RETURNING "+riderColumns,
		req.Name, req.Content, artistID)
	rd, err := scanRider(row)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create rider")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, rd)
}

// Patch applies a partial update. A field omitted from the JSON body is left
// unchanged - "name" and "content" are both required columns, so an explicit
// null for either is rejected rather than silently ignored. Ownership isn't
// patchable. A rider belonging to an artist the caller doesn't represent
// (and isn't themselves) 404s, same as Get.
func (h *Handler) Patch(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	caller := auth.UserFromContext(r.Context())
	accessibleIDs, err := tourdates.AccessibleArtistIDs(r.Context(), h.DB, caller.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to update rider")
		return
	}

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	var pb httpx.PatchBuilder
	for _, column := range []string{"name", "content"} {
		v, ok := raw[column]
		if !ok {
			continue
		}
		if string(v) == "null" {
			httpx.WriteError(w, http.StatusBadRequest, column+" cannot be null")
			return
		}
		var s string
		if err := json.Unmarshal(v, &s); err != nil || s == "" {
			httpx.WriteError(w, http.StatusBadRequest, column+" must be a non-empty string")
			return
		}
		pb.Set(column, s)
	}

	if pb.Empty() {
		httpx.WriteError(w, http.StatusBadRequest, "no updatable fields provided")
		return
	}

	rd, err := h.resource().Update(r.Context(), &pb, id, accessibleIDs)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusNotFound, "rider not found")
		return
	} else if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to update rider")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, rd)
}

// Delete 404s for a rider belonging to an artist the caller doesn't
// represent (and isn't themselves), same as Get/Patch. Tourdates using it
// survive with rider_id set to null (ON DELETE SET NULL).
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	tourdates.HandleDelete(w, r, h.resource(), "rider")
}
