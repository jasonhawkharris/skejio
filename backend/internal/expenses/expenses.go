package expenses

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
	"skejio/backend/internal/httpx"
	"skejio/backend/internal/tourdates"
)

var validCategories = map[string]bool{
	"TRAVEL":  true,
	"LODGING": true,
	"CREW":    true,
	"GEAR":    true,
	"OTHER":   true,
}

// Expense is one line item of spend against a tourdate. Unlike financials
// (one-to-one), a tourdate can have any number of expenses.
type Expense struct {
	ID          uuid.UUID `json:"id"`
	TourDateID  uuid.UUID `json:"tourdate_id"`
	Category    string    `json:"category"`
	Amount      float64   `json:"amount"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type Handler struct {
	DB *pgxpool.Pool
}

const expenseColumns = "id, tourdate_id, category, amount, description, created_at"

func scanExpense(row pgx.Row) (Expense, error) {
	var e Expense
	err := row.Scan(&e.ID, &e.TourDateID, &e.Category, &e.Amount, &e.Description, &e.CreatedAt)
	return e, err
}

// List returns every expense recorded against a tourdate, oldest first.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	tourDateID, ok := tourdates.AccessibleTourDateID(w, r, h.DB)
	if !ok {
		return
	}

	rows, err := h.DB.Query(r.Context(),
		"SELECT "+expenseColumns+" FROM expenses WHERE tourdate_id = $1 ORDER BY created_at", tourDateID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list expenses")
		return
	}
	expenseList, err := db.ScanAll(rows, scanExpense)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to read expenses")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, expenseList)
}

// Get returns a single expense. 404s if the tourdate isn't accessible, or
// the expense doesn't belong to it.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	tourDateID, ok := tourdates.AccessibleTourDateID(w, r, h.DB)
	if !ok {
		return
	}
	expenseID, err := uuid.Parse(mux.Vars(r)["expenseID"])
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	row := h.DB.QueryRow(r.Context(),
		"SELECT "+expenseColumns+" FROM expenses WHERE id = $1 AND tourdate_id = $2", expenseID, tourDateID)
	e, err := scanExpense(row)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusNotFound, "expense not found")
		return
	} else if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to read expense")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, e)
}

type createExpenseRequest struct {
	Category    string   `json:"category"`
	Amount      *float64 `json:"amount"`
	Description *string  `json:"description"`
}

// Create adds a new expense line item to a tourdate.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	tourDateID, ok := tourdates.AccessibleTourDateID(w, r, h.DB)
	if !ok {
		return
	}

	var req createExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if !validCategories[req.Category] {
		httpx.WriteError(w, http.StatusBadRequest, "category must be one of TRAVEL, LODGING, CREW, GEAR, OTHER")
		return
	}
	if req.Amount == nil {
		httpx.WriteError(w, http.StatusBadRequest, "amount is required")
		return
	}

	row := h.DB.QueryRow(r.Context(),
		"INSERT INTO expenses (tourdate_id, category, amount, description) VALUES ($1, $2, $3, $4) RETURNING "+expenseColumns,
		tourDateID, req.Category, *req.Amount, req.Description)
	e, err := scanExpense(row)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create expense")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, e)
}

// Patch applies a partial update to an expense. A field omitted from the
// JSON body is left unchanged. "description" is nullable and an explicit
// null clears it; "category" and "amount" are required columns, so an
// explicit null for either is rejected rather than silently ignored.
func (h *Handler) Patch(w http.ResponseWriter, r *http.Request) {
	tourDateID, ok := tourdates.AccessibleTourDateID(w, r, h.DB)
	if !ok {
		return
	}
	expenseID, err := uuid.Parse(mux.Vars(r)["expenseID"])
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

	if v, ok := raw["category"]; ok {
		if string(v) == "null" {
			httpx.WriteError(w, http.StatusBadRequest, "category cannot be null")
			return
		}
		var s string
		if err := json.Unmarshal(v, &s); err != nil || !validCategories[s] {
			httpx.WriteError(w, http.StatusBadRequest, "category must be one of TRAVEL, LODGING, CREW, GEAR, OTHER")
			return
		}
		pb.Set("category", s)
	}
	if v, ok := raw["amount"]; ok {
		if string(v) == "null" {
			httpx.WriteError(w, http.StatusBadRequest, "amount cannot be null")
			return
		}
		var f float64
		if err := json.Unmarshal(v, &f); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "amount must be a number")
			return
		}
		pb.Set("amount", f)
	}
	if v, ok := raw["description"]; ok {
		if string(v) == "null" {
			pb.Set("description", nil)
		} else {
			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				httpx.WriteError(w, http.StatusBadRequest, "description must be a string or null")
				return
			}
			pb.Set("description", s)
		}
	}

	if pb.Empty() {
		httpx.WriteError(w, http.StatusBadRequest, "no updatable fields provided")
		return
	}

	where := fmt.Sprintf("id = $%d AND tourdate_id = $%d", pb.NextArg(), pb.NextArg()+1)
	query, args := pb.Build("expenses", where, expenseColumns, expenseID, tourDateID)

	row := h.DB.QueryRow(r.Context(), query, args...)
	e, err := scanExpense(row)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusNotFound, "expense not found")
		return
	} else if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to update expense")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, e)
}

// Delete removes an expense. 404s if the tourdate isn't accessible, or the
// expense doesn't belong to it.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	tourDateID, ok := tourdates.AccessibleTourDateID(w, r, h.DB)
	if !ok {
		return
	}
	expenseID, err := uuid.Parse(mux.Vars(r)["expenseID"])
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	tag, err := h.DB.Exec(r.Context(), "DELETE FROM expenses WHERE id = $1 AND tourdate_id = $2", expenseID, tourDateID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to delete expense")
		return
	}
	if tag.RowsAffected() == 0 {
		httpx.WriteError(w, http.StatusNotFound, "expense not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
