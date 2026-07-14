package financials

import (
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
	"skejio/backend/internal/dberr"
	"skejio/backend/internal/httpx"
	"skejio/backend/internal/tourdates"
)

// Financials holds the money side of a tourdate, kept in its own table
// (rather than more nullable columns on tourdates) since it's a distinct
// concern that's likely to grow its own fields over time. One-to-one with
// tourdates - enforced by a UNIQUE constraint on tourdate_id.
type Financials struct {
	ID         uuid.UUID `json:"id"`
	TourDateID uuid.UUID `json:"tourdate_id"`
	Fee        *float64  `json:"fee"`
	Tips       *float64  `json:"tips"`
	CreatedAt  time.Time `json:"created_at"`
}

type Handler struct {
	DB *pgxpool.Pool
}

const financialsColumns = "id, tourdate_id, fee, tips, created_at"

func scanFinancials(row pgx.Row) (Financials, error) {
	var f Financials
	err := row.Scan(&f.ID, &f.TourDateID, &f.Fee, &f.Tips, &f.CreatedAt)
	return f, err
}

// accessibleTourDateID parses the "id" route param and confirms the caller
// may access that tourdate (their own, or one they represent) - 404ing
// otherwise, same as tourdates itself, so as not to reveal whether the
// tourdate exists.
func accessibleTourDateID(w http.ResponseWriter, r *http.Request, db *pgxpool.Pool) (uuid.UUID, bool) {
	tourDateID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id")
		return uuid.Nil, false
	}

	caller := auth.UserFromContext(r.Context())
	accessibleIDs, err := tourdates.AccessibleArtistIDs(r.Context(), db, caller.ID)
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
	if !tourdates.ContainsID(accessibleIDs, ownerID) {
		httpx.WriteError(w, http.StatusNotFound, "tourdate not found")
		return uuid.Nil, false
	}

	return tourDateID, true
}

// Get returns the financials row for a tourdate. 404s if the tourdate isn't
// accessible, or if it's accessible but has no financials row yet.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	tourDateID, ok := accessibleTourDateID(w, r, h.DB)
	if !ok {
		return
	}

	row := h.DB.QueryRow(r.Context(), "SELECT "+financialsColumns+" FROM financials WHERE tourdate_id = $1", tourDateID)
	f, err := scanFinancials(row)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusNotFound, "financials not found")
		return
	} else if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to read financials")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, f)
}

// Summary is the computed money picture of a tourdate: financials plus the
// running total of its expenses. Computed on read rather than stored, so it
// never drifts out of sync with the underlying financials/expenses rows.
type Summary struct {
	TourDateID    uuid.UUID `json:"tourdate_id"`
	Fee           *float64  `json:"fee"`
	Tips          *float64  `json:"tips"`
	TotalExpenses float64   `json:"total_expenses"`
	Net           float64   `json:"net"`
}

// Summary returns fee, tips, total expenses, and net (fee + tips -
// total_expenses) for a tourdate. fee/tips are nil if no financials row
// exists yet; total_expenses is 0 (not nil) if there are no expenses.
func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	tourDateID, ok := accessibleTourDateID(w, r, h.DB)
	if !ok {
		return
	}

	var fee, tips *float64
	err := h.DB.QueryRow(r.Context(), "SELECT fee, tips FROM financials WHERE tourdate_id = $1", tourDateID).Scan(&fee, &tips)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to compute summary")
		return
	}

	var totalExpenses float64
	err = h.DB.QueryRow(r.Context(), "SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE tourdate_id = $1", tourDateID).Scan(&totalExpenses)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to compute summary")
		return
	}

	net := -totalExpenses
	if fee != nil {
		net += *fee
	}
	if tips != nil {
		net += *tips
	}

	httpx.WriteJSON(w, http.StatusOK, Summary{
		TourDateID:    tourDateID,
		Fee:           fee,
		Tips:          tips,
		TotalExpenses: totalExpenses,
		Net:           net,
	})
}

type createFinancialsRequest struct {
	Fee  *float64 `json:"fee"`
	Tips *float64 `json:"tips"`
}

// Create adds the financials row for a tourdate. 409s if one already exists
// (tourdate_id is UNIQUE) - use Patch to update it instead.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	tourDateID, ok := accessibleTourDateID(w, r, h.DB)
	if !ok {
		return
	}

	var req createFinancialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	row := h.DB.QueryRow(r.Context(),
		"INSERT INTO financials (tourdate_id, fee, tips) VALUES ($1, $2, $3) RETURNING "+financialsColumns,
		tourDateID, req.Fee, req.Tips)
	f, err := scanFinancials(row)
	if dberr.IsUniqueViolation(err) {
		httpx.WriteError(w, http.StatusConflict, "financials already exist for this tourdate")
		return
	} else if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create financials")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, f)
}

// Patch applies a partial update to a tourdate's financials. A field omitted
// from the JSON body is left unchanged; a field present but set to null
// clears it. 404s if the tourdate isn't accessible, or has no financials row
// yet (use Create first).
func (h *Handler) Patch(w http.ResponseWriter, r *http.Request) {
	tourDateID, ok := accessibleTourDateID(w, r, h.DB)
	if !ok {
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

	for _, column := range []string{"fee", "tips"} {
		v, ok := raw[column]
		if !ok {
			continue
		}
		if string(v) == "null" {
			addSet(column, nil)
			continue
		}
		var f float64
		if err := json.Unmarshal(v, &f); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, column+" must be a number or null")
			return
		}
		addSet(column, f)
	}

	if len(setClauses) == 0 {
		httpx.WriteError(w, http.StatusBadRequest, "no updatable fields provided")
		return
	}

	args = append(args, tourDateID)
	query := fmt.Sprintf(
		"UPDATE financials SET %s WHERE tourdate_id = $%d RETURNING "+financialsColumns,
		strings.Join(setClauses, ", "), argPos,
	)

	row := h.DB.QueryRow(r.Context(), query, args...)
	f, err := scanFinancials(row)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusNotFound, "financials not found")
		return
	} else if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to update financials")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, f)
}

// Delete removes a tourdate's financials row. 404s if the tourdate isn't
// accessible, or has no financials row.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	tourDateID, ok := accessibleTourDateID(w, r, h.DB)
	if !ok {
		return
	}

	tag, err := h.DB.Exec(r.Context(), "DELETE FROM financials WHERE tourdate_id = $1", tourDateID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to delete financials")
		return
	}
	if tag.RowsAffected() == 0 {
		httpx.WriteError(w, http.StatusNotFound, "financials not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
