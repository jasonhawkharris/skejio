package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
	ID        uuid.UUID `json:"id"`
	Date      Date      `json:"date"`
	City      string    `json:"city"`
	State     *string   `json:"state"`
	Venue     string    `json:"venue"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type App struct {
	db *pgxpool.Pool
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func scanTourDate(row pgx.Row) (TourDate, error) {
	var td TourDate
	var d time.Time
	err := row.Scan(&td.ID, &d, &td.City, &td.State, &td.Venue, &td.UserID, &td.CreatedAt)
	td.Date = Date(d)
	return td, err
}

const tourDateColumns = "id, date, city, state, venue, user_id, created_at"

// ListTourDates returns only the tourdates owned by the authenticated user.
func (a *App) ListTourDates(w http.ResponseWriter, r *http.Request) {
	userID := userFromContext(r.Context()).ID

	rows, err := a.db.Query(r.Context(),
		"SELECT "+tourDateColumns+" FROM tourdates WHERE user_id = $1 ORDER BY date", userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list tourdates")
		return
	}
	defer rows.Close()

	tourdates := []TourDate{}
	for rows.Next() {
		td, err := scanTourDate(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read tourdates")
			return
		}
		tourdates = append(tourdates, td)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read tourdates")
		return
	}

	writeJSON(w, http.StatusOK, tourdates)
}

// GetTourDate 404s (rather than 403s) when the tourdate belongs to another
// user, so as not to reveal that it exists.
func (a *App) GetTourDate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	userID := userFromContext(r.Context()).ID

	row := a.db.QueryRow(r.Context(),
		"SELECT "+tourDateColumns+" FROM tourdates WHERE id = $1 AND user_id = $2", id, userID)
	td, err := scanTourDate(row)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "tourdate not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read tourdate")
		return
	}

	writeJSON(w, http.StatusOK, td)
}

type createTourDateRequest struct {
	Date  Date    `json:"date"`
	City  string  `json:"city"`
	State *string `json:"state"`
	Venue string  `json:"venue"`
}

// CreateTourDate always assigns ownership to the authenticated user; the
// client cannot specify a different owner.
func (a *App) CreateTourDate(w http.ResponseWriter, r *http.Request) {
	var req createTourDateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if time.Time(req.Date).IsZero() {
		writeError(w, http.StatusBadRequest, "date is required")
		return
	}
	if req.City == "" {
		writeError(w, http.StatusBadRequest, "city is required")
		return
	}
	if req.Venue == "" {
		writeError(w, http.StatusBadRequest, "venue is required")
		return
	}
	userID := userFromContext(r.Context()).ID

	row := a.db.QueryRow(r.Context(),
		"INSERT INTO tourdates (date, city, state, venue, user_id) VALUES ($1, $2, $3, $4, $5) RETURNING "+tourDateColumns,
		time.Time(req.Date), req.City, req.State, req.Venue, userID)
	td, err := scanTourDate(row)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create tourdate")
		return
	}

	writeJSON(w, http.StatusCreated, td)
}

// PatchTourDate applies a partial update. A field omitted from the JSON body
// is left unchanged; a field present but set to null clears it (only valid
// for the nullable "state" column). Ownership is not patchable. A tourdate
// owned by another user 404s, same as GetTourDate.
func (a *App) PatchTourDate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	userID := userFromContext(r.Context()).ID

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
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
			writeError(w, http.StatusBadRequest, "date must be in YYYY-MM-DD format")
			return
		}
		addSet("date", time.Time(d))
	}
	if v, ok := raw["city"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil || s == "" {
			writeError(w, http.StatusBadRequest, "city must be a non-empty string")
			return
		}
		addSet("city", s)
	}
	if v, ok := raw["state"]; ok {
		if string(v) == "null" {
			addSet("state", nil)
		} else {
			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				writeError(w, http.StatusBadRequest, "state must be a string or null")
				return
			}
			addSet("state", s)
		}
	}
	if v, ok := raw["venue"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil || s == "" {
			writeError(w, http.StatusBadRequest, "venue must be a non-empty string")
			return
		}
		addSet("venue", s)
	}

	if len(setClauses) == 0 {
		writeError(w, http.StatusBadRequest, "no updatable fields provided")
		return
	}

	args = append(args, id, userID)
	query := fmt.Sprintf(
		"UPDATE tourdates SET %s WHERE id = $%d AND user_id = $%d RETURNING "+tourDateColumns,
		strings.Join(setClauses, ", "), argPos, argPos+1,
	)

	row := a.db.QueryRow(r.Context(), query, args...)
	td, err := scanTourDate(row)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "tourdate not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update tourdate")
		return
	}

	writeJSON(w, http.StatusOK, td)
}

// DeleteTourDate 404s for a tourdate owned by another user, same as
// GetTourDate/PatchTourDate.
func (a *App) DeleteTourDate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	userID := userFromContext(r.Context()).ID

	tag, err := a.db.Exec(r.Context(), "DELETE FROM tourdates WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete tourdate")
		return
	}
	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "tourdate not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
