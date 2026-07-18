// Package merch manages an artist's merchandise catalog: items like shirts,
// vinyl, or posters, owned by an artist the same way tours/riders are -
// caller or an artist they represent. Per-variant (size/color/etc)
// inventory is tracked in internal/merchvariants, nested under
// /merch/{id}/variants.
package merch

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

var validTypes = map[string]bool{
	"APPAREL": true,
	"MUSIC":   true,
	"MISC":    true,
}

// Merch is one item in an artist's merchandise catalog. Price and cost are
// uniform across an item's variants - only inventory varies by variant (see
// internal/merchvariants).
type Merch struct {
	ID                uuid.UUID `json:"id"`
	UserID            uuid.UUID `json:"user_id"`
	Name              string    `json:"name"`
	Type              string    `json:"type"`
	Description       *string   `json:"description"`
	ManufacturingCost float64   `json:"manufacturing_cost"`
	SellingPrice      *float64  `json:"selling_price"`
	PhotoURL          *string   `json:"photo_url"`
	CreatedAt         time.Time `json:"created_at"`
}

type Handler struct {
	DB *pgxpool.Pool
}

const merchColumns = "id, user_id, name, type, description, manufacturing_cost, selling_price, photo_url, created_at"

func scanMerch(row pgx.Row) (Merch, error) {
	var m Merch
	err := row.Scan(&m.ID, &m.UserID, &m.Name, &m.Type, &m.Description, &m.ManufacturingCost, &m.SellingPrice, &m.PhotoURL, &m.CreatedAt)
	return m, err
}

func (h *Handler) resource() tourdates.ScopedResource[Merch] {
	return tourdates.ScopedResource[Merch]{DB: h.DB, Table: "merch", Columns: merchColumns, OrderBy: "created_at", Scan: scanMerch}
}

// AccessibleMerchID parses the "id" route param and confirms the caller may
// access that merch item (their own, or one belonging to an artist they
// represent) - 404ing otherwise, so as not to reveal whether it exists.
// Shared by any resource nested under /merch/{id}/... (variants).
func AccessibleMerchID(w http.ResponseWriter, r *http.Request, db *pgxpool.Pool) (uuid.UUID, bool) {
	merchID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id")
		return uuid.Nil, false
	}

	caller := auth.UserFromContext(r.Context())
	accessibleIDs, err := tourdates.AccessibleArtistIDs(r.Context(), db, caller.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to look up merch")
		return uuid.Nil, false
	}

	var ownerID uuid.UUID
	err = db.QueryRow(r.Context(), "SELECT user_id FROM merch WHERE id = $1", merchID).Scan(&ownerID)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusNotFound, "merch not found")
		return uuid.Nil, false
	} else if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to look up merch")
		return uuid.Nil, false
	}
	if !tourdates.ContainsID(accessibleIDs, ownerID) {
		httpx.WriteError(w, http.StatusNotFound, "merch not found")
		return uuid.Nil, false
	}

	return merchID, true
}

// List returns the caller's own merch plus every item belonging to an artist
// they represent. If an "artist_id" query param is given, the list is
// narrowed to just that artist - who must be the caller or someone they
// represent, else a 404 (so as not to reveal whether that artist_id exists).
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	tourdates.HandleList(w, r, h.resource(), "merch")
}

// Get 404s (rather than 403s) when the item belongs to an artist the caller
// doesn't represent (and isn't themselves), so as not to reveal that it
// exists.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	tourdates.HandleGet(w, r, h.resource(), "merch")
}

type createMerchRequest struct {
	Name              string    `json:"name"`
	Type              string    `json:"type"`
	Description       *string   `json:"description"`
	ManufacturingCost *float64  `json:"manufacturing_cost"`
	SellingPrice      *float64  `json:"selling_price"`
	PhotoURL          *string   `json:"photo_url"`
	ArtistID          uuid.UUID `json:"artist_id"`
}

// Create assigns ownership to artist_id if given - which must be the caller
// themselves or an artist they represent (403 otherwise) - or to the caller
// themselves if artist_id is omitted. Same pattern as tourdates/tours/riders
// Create.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createMerchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Name == "" {
		httpx.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}
	if !validTypes[req.Type] {
		httpx.WriteError(w, http.StatusBadRequest, "type must be one of APPAREL, MUSIC, MISC")
		return
	}
	if req.ManufacturingCost == nil {
		httpx.WriteError(w, http.StatusBadRequest, "manufacturing_cost is required")
		return
	}

	artistID, ok := tourdates.ResolveCreateOwner(w, r, h.DB, req.ArtistID, "merch")
	if !ok {
		return
	}

	row := h.DB.QueryRow(r.Context(),
		"INSERT INTO merch (user_id, name, type, description, manufacturing_cost, selling_price, photo_url) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING "+merchColumns,
		artistID, req.Name, req.Type, req.Description, *req.ManufacturingCost, req.SellingPrice, req.PhotoURL)
	m, err := scanMerch(row)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create merch")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, m)
}

// Patch applies a partial update. A field omitted from the JSON body is left
// unchanged; "description", "photo_url", and "selling_price" are nullable
// and an explicit null clears them, while "name", "type", and
// "manufacturing_cost" are required columns, so an explicit null for any of
// them is rejected. Ownership isn't patchable. An item belonging to an
// artist the caller doesn't represent (and isn't themselves) 404s, same as
// Get.
func (h *Handler) Patch(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	caller := auth.UserFromContext(r.Context())
	accessibleIDs, err := tourdates.AccessibleArtistIDs(r.Context(), h.DB, caller.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to update merch")
		return
	}

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	var pb httpx.PatchBuilder

	if v, ok := raw["name"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil || s == "" {
			httpx.WriteError(w, http.StatusBadRequest, "name must be a non-empty string")
			return
		}
		pb.Set("name", s)
	}
	if v, ok := raw["type"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil || !validTypes[s] {
			httpx.WriteError(w, http.StatusBadRequest, "type must be one of APPAREL, MUSIC, MISC")
			return
		}
		pb.Set("type", s)
	}
	if v, ok := raw["manufacturing_cost"]; ok {
		if string(v) == "null" {
			httpx.WriteError(w, http.StatusBadRequest, "manufacturing_cost cannot be null")
			return
		}
		var f float64
		if err := json.Unmarshal(v, &f); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "manufacturing_cost must be a number")
			return
		}
		pb.Set("manufacturing_cost", f)
	}
	if v, ok := raw["selling_price"]; ok {
		if string(v) == "null" {
			pb.Set("selling_price", nil)
		} else {
			var f float64
			if err := json.Unmarshal(v, &f); err != nil {
				httpx.WriteError(w, http.StatusBadRequest, "selling_price must be a number or null")
				return
			}
			pb.Set("selling_price", f)
		}
	}
	for _, column := range []string{"description", "photo_url"} {
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

	m, err := h.resource().Update(r.Context(), &pb, id, accessibleIDs)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusNotFound, "merch not found")
		return
	} else if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to update merch")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, m)
}

// Delete 404s for an item belonging to an artist the caller doesn't
// represent (and isn't themselves), same as Get/Patch. Its variants are
// removed along with it (ON DELETE CASCADE).
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	tourdates.HandleDelete(w, r, h.resource(), "merch")
}
