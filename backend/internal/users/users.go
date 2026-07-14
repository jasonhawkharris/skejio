package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"skejio/backend/internal/auth"
	"skejio/backend/internal/dberr"
	"skejio/backend/internal/httpx"
	"skejio/backend/internal/password"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	UserType     string    `json:"user_type"`
	CreatedAt    time.Time `json:"created_at"`
}

var validUserTypes = map[string]bool{
	"ARTIST":  true,
	"MANAGER": true,
	"AGENT":   true,
	"CREW":    true,
	"LABEL":   true,
}

type Handler struct {
	DB *pgxpool.Pool
}

const userColumns = "id, name, email, user_type, created_at"

func scanUser(row pgx.Row) (User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.UserType, &u.CreatedAt)
	return u, err
}

// Get 404s (rather than 403s) for an id other than the caller's own, so as
// not to reveal whether that account exists.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if id != auth.UserFromContext(r.Context()).ID {
		httpx.WriteError(w, http.StatusNotFound, "user not found")
		return
	}

	row := h.DB.QueryRow(r.Context(), "SELECT "+userColumns+" FROM users WHERE id = $1", id)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusNotFound, "user not found")
		return
	} else if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to read user")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, u)
}

type createUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	UserType string `json:"user_type"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Name == "" {
		httpx.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Email == "" {
		httpx.WriteError(w, http.StatusBadRequest, "email is required")
		return
	}
	if req.Password == "" {
		httpx.WriteError(w, http.StatusBadRequest, "password is required")
		return
	}
	if !validUserTypes[req.UserType] {
		httpx.WriteError(w, http.StatusBadRequest, "user_type must be one of ARTIST, MANAGER, AGENT, CREW, LABEL")
		return
	}

	hash, err := password.Hash(req.Password)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	row := h.DB.QueryRow(r.Context(),
		"INSERT INTO users (name, email, password_hash, user_type) VALUES ($1, $2, $3, $4) RETURNING "+userColumns,
		req.Name, req.Email, hash, req.UserType)
	u, err := scanUser(row)
	if dberr.IsUniqueViolation(err) {
		httpx.WriteError(w, http.StatusConflict, "email is already in use")
		return
	} else if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, u)
}

// Patch applies a partial update to name, email, password, and/or user_type.
// A field omitted from the JSON body is left unchanged; none of these may be
// set to null since all of the underlying columns are NOT NULL. An id other
// than the caller's own 404s, same as Get.
func (h *Handler) Patch(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if id != auth.UserFromContext(r.Context()).ID {
		httpx.WriteError(w, http.StatusNotFound, "user not found")
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
	if v, ok := raw["email"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil || s == "" {
			httpx.WriteError(w, http.StatusBadRequest, "email must be a non-empty string")
			return
		}
		pb.Set("email", s)
	}
	if v, ok := raw["password"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil || s == "" {
			httpx.WriteError(w, http.StatusBadRequest, "password must be a non-empty string")
			return
		}
		hash, err := password.Hash(s)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "failed to update user")
			return
		}
		pb.Set("password_hash", hash)
	}
	if v, ok := raw["user_type"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err != nil || !validUserTypes[s] {
			httpx.WriteError(w, http.StatusBadRequest, "user_type must be one of ARTIST, MANAGER, AGENT, CREW, LABEL")
			return
		}
		pb.Set("user_type", s)
	}

	if pb.Empty() {
		httpx.WriteError(w, http.StatusBadRequest, "no updatable fields provided")
		return
	}

	query, args := pb.Build("users", fmt.Sprintf("id = $%d", pb.NextArg()), userColumns, id)

	row := h.DB.QueryRow(r.Context(), query, args...)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusNotFound, "user not found")
		return
	} else if dberr.IsUniqueViolation(err) {
		httpx.WriteError(w, http.StatusConflict, "email is already in use")
		return
	} else if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, u)
}

// Delete 404s for an id other than the caller's own, same as Get.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if id != auth.UserFromContext(r.Context()).ID {
		httpx.WriteError(w, http.StatusNotFound, "user not found")
		return
	}

	tag, err := h.DB.Exec(r.Context(), "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to delete user")
		return
	}
	if tag.RowsAffected() == 0 {
		httpx.WriteError(w, http.StatusNotFound, "user not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
