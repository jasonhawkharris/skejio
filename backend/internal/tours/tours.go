package tours

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

// Tour groups a set of tourdates under one name. One-to-many with tourdates
// (tourdates.tour_id), owned by an artist the same way tourdates are.
type Tour struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Handler struct {
	DB *pgxpool.Pool
}

const tourColumns = "id, name, user_id, created_at"

func scanTour(row pgx.Row) (Tour, error) {
	var t Tour
	err := row.Scan(&t.ID, &t.Name, &t.UserID, &t.CreatedAt)
	return t, err
}

func (h *Handler) resource() tourdates.ScopedResource[Tour] {
	return tourdates.ScopedResource[Tour]{DB: h.DB, Table: "tours", Columns: tourColumns, OrderBy: "created_at", Scan: scanTour}
}

// List returns the caller's own tours plus every tour belonging to an artist
// they represent. If an "artist_id" query param is given, the list is
// narrowed to just that artist - who must be the caller or someone they
// represent, else a 404 (so as not to reveal whether that artist_id exists).
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	tourdates.HandleList(w, r, h.resource(), "tour")
}

// Get 404s (rather than 403s) when the tour belongs to an artist the caller
// doesn't represent (and isn't themselves), so as not to reveal that it
// exists.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	tourdates.HandleGet(w, r, h.resource(), "tour")
}

type createTourRequest struct {
	Name     string    `json:"name"`
	ArtistID uuid.UUID `json:"artist_id"`
}

// Create assigns ownership to artist_id if given - which must be the caller
// themselves or an artist they represent (403 otherwise) - or to the caller
// themselves if artist_id is omitted. Same pattern as tourdates.Create.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createTourRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Name == "" {
		httpx.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}

	artistID, ok := tourdates.ResolveCreateOwner(w, r, h.DB, req.ArtistID, "tour")
	if !ok {
		return
	}

	row := h.DB.QueryRow(r.Context(),
		"INSERT INTO tours (name, user_id) VALUES ($1, $2) RETURNING "+tourColumns,
		req.Name, artistID)
	t, err := scanTour(row)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create tour")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, t)
}

// Patch applies a partial update. Only "name" is patchable - ownership isn't.
// A tour belonging to an artist the caller doesn't represent (and isn't
// themselves) 404s, same as Get.
func (h *Handler) Patch(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	caller := auth.UserFromContext(r.Context())
	accessibleIDs, err := tourdates.AccessibleArtistIDs(r.Context(), h.DB, caller.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to update tour")
		return
	}

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	v, ok := raw["name"]
	if !ok {
		httpx.WriteError(w, http.StatusBadRequest, "no updatable fields provided")
		return
	}
	var name string
	if err := json.Unmarshal(v, &name); err != nil || name == "" {
		httpx.WriteError(w, http.StatusBadRequest, "name must be a non-empty string")
		return
	}

	var pb httpx.PatchBuilder
	pb.Set("name", name)

	t, err := h.resource().Update(r.Context(), &pb, id, accessibleIDs)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusNotFound, "tour not found")
		return
	} else if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to update tour")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, t)
}

// Delete 404s for a tour belonging to an artist the caller doesn't represent
// (and isn't themselves), same as Get/Patch. Its tourdates survive with
// tour_id set to null (ON DELETE SET NULL).
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	tourdates.HandleDelete(w, r, h.resource(), "tour")
}
