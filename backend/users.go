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
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

const userColumns = "id, name, email, created_at"

func scanUser(row pgx.Row) (User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
	return u, err
}

func (a *App) ListUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.Query(r.Context(), "SELECT "+userColumns+" FROM users ORDER BY created_at")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list users")
		return
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read users")
			return
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read users")
		return
	}

	writeJSON(w, http.StatusOK, users)
}

func (a *App) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	row := a.db.QueryRow(r.Context(), "SELECT "+userColumns+" FROM users WHERE id = $1", id)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "user not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read user")
		return
	}

	writeJSON(w, http.StatusOK, u)
}

type createUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *App) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "email is required")
		return
	}
	if req.Password == "" {
		writeError(w, http.StatusBadRequest, "password is required")
		return
	}

	hash, err := hashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	row := a.db.QueryRow(r.Context(),
		"INSERT INTO users (name, email, password_hash) VALUES ($1, $2, $3) RETURNING "+userColumns,
		req.Name, req.Email, hash)
	u, err := scanUser(row)
	if isUniqueViolation(err) {
		writeError(w, http.StatusConflict, "email is already in use")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	writeJSON(w, http.StatusCreated, u)
}

// PatchUser applies a partial update to name, email, and/or password. A field
// omitted from the JSON body is left unchanged; name/email/password may not
// be set to null since all three columns are NOT NULL.
func (a *App) PatchUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
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

	if v, ok := raw["name"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil || s == "" {
			writeError(w, http.StatusBadRequest, "name must be a non-empty string")
			return
		}
		addSet("name", s)
	}
	if v, ok := raw["email"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil || s == "" {
			writeError(w, http.StatusBadRequest, "email must be a non-empty string")
			return
		}
		addSet("email", s)
	}
	if v, ok := raw["password"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil || s == "" {
			writeError(w, http.StatusBadRequest, "password must be a non-empty string")
			return
		}
		hash, err := hashPassword(s)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to update user")
			return
		}
		addSet("password_hash", hash)
	}

	if len(setClauses) == 0 {
		writeError(w, http.StatusBadRequest, "no updatable fields provided")
		return
	}

	args = append(args, id)
	query := fmt.Sprintf(
		"UPDATE users SET %s WHERE id = $%d RETURNING "+userColumns,
		strings.Join(setClauses, ", "), argPos,
	)

	row := a.db.QueryRow(r.Context(), query, args...)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "user not found")
		return
	} else if isUniqueViolation(err) {
		writeError(w, http.StatusConflict, "email is already in use")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	writeJSON(w, http.StatusOK, u)
}

func (a *App) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	tag, err := a.db.Exec(r.Context(), "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete user")
		return
	}
	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
