package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"skejio/backend/internal/httpx"
	"skejio/backend/internal/password"
)

const sessionDuration = 24 * time.Hour

type contextKey int

const currentUserContextKey contextKey = iota

type CurrentUser struct {
	ID       uuid.UUID
	UserType string
	Token    string
}

func UserFromContext(ctx context.Context) CurrentUser {
	u, _ := ctx.Value(currentUserContextKey).(CurrentUser)
	return u
}

func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Middleware requires a valid, unexpired "Authorization: Bearer <token>"
// session and attaches the authenticated user to the request context.
func Middleware(db *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok || token == "" {
				httpx.WriteError(w, http.StatusUnauthorized, "authentication required")
				return
			}

			var userID uuid.UUID
			var userType string
			err := db.QueryRow(r.Context(),
				`SELECT s.user_id, u.user_type FROM sessions s
				 JOIN users u ON u.id = s.user_id
				 WHERE s.token = $1 AND s.expires_at > now()`, token,
			).Scan(&userID, &userType)
			if errors.Is(err, pgx.ErrNoRows) {
				httpx.WriteError(w, http.StatusUnauthorized, "invalid or expired session")
				return
			} else if err != nil {
				httpx.WriteError(w, http.StatusInternalServerError, "failed to authenticate")
				return
			}

			ctx := context.WithValue(r.Context(), currentUserContextKey, CurrentUser{ID: userID, UserType: userType, Token: token})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type Handler struct {
	DB *pgxpool.Pool
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Email == "" || req.Password == "" {
		httpx.WriteError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	var userID uuid.UUID
	var passwordHash string
	var userType string
	var name string
	err := h.DB.QueryRow(r.Context(), "SELECT id, password_hash, user_type, name FROM users WHERE email = $1", req.Email).
		Scan(&userID, &passwordHash, &userType, &name)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusUnauthorized, "invalid email or password")
		return
	} else if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to log in")
		return
	}

	if err := password.Verify(passwordHash, req.Password); err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	token, err := generateSessionToken()
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to log in")
		return
	}
	expiresAt := time.Now().Add(sessionDuration)

	if _, err := h.DB.Exec(r.Context(),
		"INSERT INTO sessions (token, user_id, expires_at) VALUES ($1, $2, $3)",
		token, userID, expiresAt,
	); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to log in")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"token":      token,
		"expires_at": expiresAt,
		"user_type":  userType,
		"name":       name,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cu := UserFromContext(r.Context())
	if _, err := h.DB.Exec(r.Context(), "DELETE FROM sessions WHERE token = $1", cu.Token); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to log out")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
