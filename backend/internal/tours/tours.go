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

// List returns the caller's own tours plus every tour belonging to an artist
// they represent. If an "artist_id" query param is given, the list is
// narrowed to just that artist - who must be the caller or someone they
// represent, else a 404 (so as not to reveal whether that artist_id exists).
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	caller := auth.UserFromContext(r.Context())
	accessibleIDs, err := tourdates.AccessibleArtistIDs(r.Context(), h.DB, caller.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list tours")
		return
	}

	var rows pgx.Rows
	if artistIDParam := r.URL.Query().Get("artist_id"); artistIDParam != "" {
		artistID, err := uuid.Parse(artistIDParam)
		if err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "artist_id must be a valid UUID")
			return
		}
		if !tourdates.ContainsID(accessibleIDs, artistID) {
			httpx.WriteError(w, http.StatusNotFound, "artist not found")
			return
		}
		rows, err = h.DB.Query(r.Context(),
			"SELECT "+tourColumns+" FROM tours WHERE user_id = $1 ORDER BY created_at", artistID)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "failed to list tours")
			return
		}
	} else {
		rows, err = h.DB.Query(r.Context(),
			"SELECT "+tourColumns+" FROM tours WHERE user_id = ANY($1::uuid[]) ORDER BY created_at", accessibleIDs)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "failed to list tours")
			return
		}
	}
	defer rows.Close()

	tourList := []Tour{}
	for rows.Next() {
		t, err := scanTour(rows)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "failed to read tours")
			return
		}
		tourList = append(tourList, t)
	}
	if err := rows.Err(); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to read tours")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, tourList)
}

// Get 404s (rather than 403s) when the tour belongs to an artist the caller
// doesn't represent (and isn't themselves), so as not to reveal that it
// exists.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	caller := auth.UserFromContext(r.Context())
	accessibleIDs, err := tourdates.AccessibleArtistIDs(r.Context(), h.DB, caller.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to read tour")
		return
	}

	row := h.DB.QueryRow(r.Context(),
		"SELECT "+tourColumns+" FROM tours WHERE id = $1 AND user_id = ANY($2::uuid[])", id, accessibleIDs)
	t, err := scanTour(row)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusNotFound, "tour not found")
		return
	} else if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to read tour")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, t)
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

	caller := auth.UserFromContext(r.Context())
	artistID := req.ArtistID
	if artistID == uuid.Nil {
		artistID = caller.ID
	} else if artistID != caller.ID {
		accessibleIDs, err := tourdates.AccessibleArtistIDs(r.Context(), h.DB, caller.ID)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "failed to create tour")
			return
		}
		if !tourdates.ContainsID(accessibleIDs, artistID) {
			httpx.WriteError(w, http.StatusForbidden, "you do not have access to create tours for this artist")
			return
		}
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

	row := h.DB.QueryRow(r.Context(),
		"UPDATE tours SET name = $1 WHERE id = $2 AND user_id = ANY($3::uuid[]) RETURNING "+tourColumns,
		name, id, accessibleIDs)
	t, err := scanTour(row)
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
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	caller := auth.UserFromContext(r.Context())
	accessibleIDs, err := tourdates.AccessibleArtistIDs(r.Context(), h.DB, caller.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to delete tour")
		return
	}

	tag, err := h.DB.Exec(r.Context(),
		"DELETE FROM tours WHERE id = $1 AND user_id = ANY($2::uuid[])", id, accessibleIDs)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to delete tour")
		return
	}
	if tag.RowsAffected() == 0 {
		httpx.WriteError(w, http.StatusNotFound, "tour not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
