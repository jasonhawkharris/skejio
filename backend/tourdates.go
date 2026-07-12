package main

import (
	"context"
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

// accessibleArtistIDs returns the set of user ids whose tourdates the caller
// may access: their own id, plus every artist they represent (if any).
// Returned as strings so it can be bound directly to a ::uuid[] parameter.
func accessibleArtistIDs(ctx context.Context, db *pgxpool.Pool, callerID uuid.UUID) ([]string, error) {
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

func containsID(ids []string, id uuid.UUID) bool {
	target := id.String()
	for _, existing := range ids {
		if existing == target {
			return true
		}
	}
	return false
}

// ListTourDates returns the caller's own tourdates plus every tourdate
// belonging to an artist they represent, merged into one list ordered by
// date. If an "artist_id" query param is given, the list is narrowed to
// just that artist - who must be the caller or someone they represent, else
// a 404 (so as not to reveal whether that artist_id exists at all).
func (a *App) ListTourDates(w http.ResponseWriter, r *http.Request) {
	caller := userFromContext(r.Context())
	accessibleIDs, err := accessibleArtistIDs(r.Context(), a.db, caller.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list tourdates")
		return
	}

	var rows pgx.Rows
	if artistIDParam := r.URL.Query().Get("artist_id"); artistIDParam != "" {
		artistID, err := uuid.Parse(artistIDParam)
		if err != nil {
			writeError(w, http.StatusBadRequest, "artist_id must be a valid UUID")
			return
		}
		if !containsID(accessibleIDs, artistID) {
			writeError(w, http.StatusNotFound, "artist not found")
			return
		}
		rows, err = a.db.Query(r.Context(),
			"SELECT "+tourDateColumns+" FROM tourdates WHERE user_id = $1 ORDER BY date", artistID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to list tourdates")
			return
		}
	} else {
		rows, err = a.db.Query(r.Context(),
			"SELECT "+tourDateColumns+" FROM tourdates WHERE user_id = ANY($1::uuid[]) ORDER BY date", accessibleIDs)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to list tourdates")
			return
		}
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

// GetTourDate 404s (rather than 403s) when the tourdate belongs to an artist
// the caller doesn't represent (and isn't themselves), so as not to reveal
// that it exists.
func (a *App) GetTourDate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	caller := userFromContext(r.Context())
	accessibleIDs, err := accessibleArtistIDs(r.Context(), a.db, caller.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read tourdate")
		return
	}

	row := a.db.QueryRow(r.Context(),
		"SELECT "+tourDateColumns+" FROM tourdates WHERE id = $1 AND user_id = ANY($2::uuid[])", id, accessibleIDs)
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
	Date     Date      `json:"date"`
	City     string    `json:"city"`
	State    *string   `json:"state"`
	Venue    string    `json:"venue"`
	ArtistID uuid.UUID `json:"artist_id"`
}

// CreateTourDate assigns ownership to artist_id if given - which must be the
// caller themselves or an artist they represent (403 otherwise) - or to the
// caller themselves if artist_id is omitted.
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

	caller := userFromContext(r.Context())
	artistID := req.ArtistID
	if artistID == uuid.Nil {
		artistID = caller.ID
	} else if artistID != caller.ID {
		accessibleIDs, err := accessibleArtistIDs(r.Context(), a.db, caller.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create tourdate")
			return
		}
		if !containsID(accessibleIDs, artistID) {
			writeError(w, http.StatusForbidden, "you do not have access to create tourdates for this artist")
			return
		}
	}

	row := a.db.QueryRow(r.Context(),
		"INSERT INTO tourdates (date, city, state, venue, user_id) VALUES ($1, $2, $3, $4, $5) RETURNING "+tourDateColumns,
		time.Time(req.Date), req.City, req.State, req.Venue, artistID)
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
// belonging to an artist the caller doesn't represent (and isn't themselves)
// 404s, same as GetTourDate.
func (a *App) PatchTourDate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	caller := userFromContext(r.Context())
	accessibleIDs, err := accessibleArtistIDs(r.Context(), a.db, caller.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update tourdate")
		return
	}

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

	args = append(args, id, accessibleIDs)
	query := fmt.Sprintf(
		"UPDATE tourdates SET %s WHERE id = $%d AND user_id = ANY($%d::uuid[]) RETURNING "+tourDateColumns,
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

// DeleteTourDate 404s for a tourdate belonging to an artist the caller
// doesn't represent (and isn't themselves), same as GetTourDate/PatchTourDate.
func (a *App) DeleteTourDate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	caller := userFromContext(r.Context())
	accessibleIDs, err := accessibleArtistIDs(r.Context(), a.db, caller.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete tourdate")
		return
	}

	tag, err := a.db.Exec(r.Context(),
		"DELETE FROM tourdates WHERE id = $1 AND user_id = ANY($2::uuid[])", id, accessibleIDs)
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
