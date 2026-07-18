// Package merchvariants tracks per-variant inventory (size for apparel,
// color for vinyl, etc - the label's meaning depends on the parent merch
// item's type) for merch items owned by internal/merch. One-to-many with
// merch, nested under /merch/{id}/variants the same way internal/expenses
// nests under /tourdates/{id}/expenses.
package merchvariants

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"skejio/backend/internal/db"
	"skejio/backend/internal/dberr"
	"skejio/backend/internal/httpx"
	"skejio/backend/internal/merch"
)

// Variant is one purchasable size/color/etc of a merch item, with its own
// inventory count. Price and manufacturing cost live on the parent merch
// item, since they're uniform across variants.
type Variant struct {
	ID             uuid.UUID `json:"id"`
	MerchID        uuid.UUID `json:"merch_id"`
	Label          string    `json:"label"`
	InventoryCount int       `json:"inventory_count"`
	CreatedAt      time.Time `json:"created_at"`
}

type Handler struct {
	DB *pgxpool.Pool
}

const variantColumns = "id, merch_id, label, inventory_count, created_at"

func scanVariant(row pgx.Row) (Variant, error) {
	var v Variant
	err := row.Scan(&v.ID, &v.MerchID, &v.Label, &v.InventoryCount, &v.CreatedAt)
	return v, err
}

// List returns every variant of a merch item, oldest first.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	merchID, ok := merch.AccessibleMerchID(w, r, h.DB)
	if !ok {
		return
	}

	rows, err := h.DB.Query(r.Context(),
		"SELECT "+variantColumns+" FROM merch_variants WHERE merch_id = $1 ORDER BY created_at", merchID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list variants")
		return
	}
	variantList, err := db.ScanAll(rows, scanVariant)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to read variants")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, variantList)
}

// Get returns a single variant. 404s if the merch item isn't accessible, or
// the variant doesn't belong to it.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	merchID, ok := merch.AccessibleMerchID(w, r, h.DB)
	if !ok {
		return
	}
	variantID, err := uuid.Parse(mux.Vars(r)["variantID"])
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	row := h.DB.QueryRow(r.Context(),
		"SELECT "+variantColumns+" FROM merch_variants WHERE id = $1 AND merch_id = $2", variantID, merchID)
	v, err := scanVariant(row)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusNotFound, "variant not found")
		return
	} else if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to read variant")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, v)
}

type createVariantRequest struct {
	Label          string `json:"label"`
	InventoryCount *int   `json:"inventory_count"`
}

// Create adds a new variant to a merch item. inventory_count defaults to 0
// if omitted. 409s if this item already has a variant with the same label.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	merchID, ok := merch.AccessibleMerchID(w, r, h.DB)
	if !ok {
		return
	}

	var req createVariantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Label == "" {
		httpx.WriteError(w, http.StatusBadRequest, "label is required")
		return
	}
	inventoryCount := 0
	if req.InventoryCount != nil {
		if *req.InventoryCount < 0 {
			httpx.WriteError(w, http.StatusBadRequest, "inventory_count must be >= 0")
			return
		}
		inventoryCount = *req.InventoryCount
	}

	row := h.DB.QueryRow(r.Context(),
		"INSERT INTO merch_variants (merch_id, label, inventory_count) VALUES ($1, $2, $3) RETURNING "+variantColumns,
		merchID, req.Label, inventoryCount)
	v, err := scanVariant(row)
	if dberr.IsUniqueViolation(err) {
		httpx.WriteError(w, http.StatusConflict, "a variant with this label already exists for this item")
		return
	} else if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create variant")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, v)
}

// Patch applies a partial update to a variant. A field omitted from the JSON
// body is left unchanged; both "label" and "inventory_count" are required
// columns, so an explicit null for either is rejected rather than silently
// ignored. 409s if changing the label would collide with another variant of
// the same item.
func (h *Handler) Patch(w http.ResponseWriter, r *http.Request) {
	merchID, ok := merch.AccessibleMerchID(w, r, h.DB)
	if !ok {
		return
	}
	variantID, err := uuid.Parse(mux.Vars(r)["variantID"])
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	var pb httpx.PatchBuilder

	if v, ok := raw["label"]; ok {
		if string(v) == "null" {
			httpx.WriteError(w, http.StatusBadRequest, "label cannot be null")
			return
		}
		var s string
		if err := json.Unmarshal(v, &s); err != nil || s == "" {
			httpx.WriteError(w, http.StatusBadRequest, "label must be a non-empty string")
			return
		}
		pb.Set("label", s)
	}
	if v, ok := raw["inventory_count"]; ok {
		if string(v) == "null" {
			httpx.WriteError(w, http.StatusBadRequest, "inventory_count cannot be null")
			return
		}
		var i int
		if err := json.Unmarshal(v, &i); err != nil || i < 0 {
			httpx.WriteError(w, http.StatusBadRequest, "inventory_count must be an integer >= 0")
			return
		}
		pb.Set("inventory_count", i)
	}

	if pb.Empty() {
		httpx.WriteError(w, http.StatusBadRequest, "no updatable fields provided")
		return
	}

	where := fmt.Sprintf("id = $%d AND merch_id = $%d", pb.NextArg(), pb.NextArg()+1)
	query, args := pb.Build("merch_variants", where, variantColumns, variantID, merchID)

	row := h.DB.QueryRow(r.Context(), query, args...)
	v, err := scanVariant(row)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusNotFound, "variant not found")
		return
	} else if dberr.IsUniqueViolation(err) {
		httpx.WriteError(w, http.StatusConflict, "a variant with this label already exists for this item")
		return
	} else if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to update variant")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, v)
}

// Delete removes a variant. 404s if the merch item isn't accessible, or the
// variant doesn't belong to it.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	merchID, ok := merch.AccessibleMerchID(w, r, h.DB)
	if !ok {
		return
	}
	variantID, err := uuid.Parse(mux.Vars(r)["variantID"])
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	tag, err := h.DB.Exec(r.Context(), "DELETE FROM merch_variants WHERE id = $1 AND merch_id = $2", variantID, merchID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to delete variant")
		return
	}
	if tag.RowsAffected() == 0 {
		httpx.WriteError(w, http.StatusNotFound, "variant not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
