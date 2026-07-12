package main

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
	"golang.org/x/crypto/bcrypt"
)

const sessionDuration = 24 * time.Hour

type contextKey int

const currentUserContextKey contextKey = iota

type currentUser struct {
	ID    uuid.UUID
	Token string
}

func userFromContext(ctx context.Context) currentUser {
	u, _ := ctx.Value(currentUserContextKey).(currentUser)
	return u
}

func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// authMiddleware requires a valid, unexpired "Authorization: Bearer <token>"
// session and attaches the authenticated user to the request context.
func authMiddleware(app *App) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok || token == "" {
				writeError(w, http.StatusUnauthorized, "authentication required")
				return
			}

			var userID uuid.UUID
			err := app.db.QueryRow(r.Context(),
				"SELECT user_id FROM sessions WHERE token = $1 AND expires_at > now()", token,
			).Scan(&userID)
			if errors.Is(err, pgx.ErrNoRows) {
				writeError(w, http.StatusUnauthorized, "invalid or expired session")
				return
			} else if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to authenticate")
				return
			}

			ctx := context.WithValue(r.Context(), currentUserContextKey, currentUser{ID: userID, Token: token})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *App) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	var userID uuid.UUID
	var passwordHash string
	err := a.db.QueryRow(r.Context(), "SELECT id, password_hash FROM users WHERE email = $1", req.Email).
		Scan(&userID, &passwordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to log in")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	token, err := generateSessionToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to log in")
		return
	}
	expiresAt := time.Now().Add(sessionDuration)

	if _, err := a.db.Exec(r.Context(),
		"INSERT INTO sessions (token, user_id, expires_at) VALUES ($1, $2, $3)",
		token, userID, expiresAt,
	); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to log in")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"token":      token,
		"expires_at": expiresAt,
	})
}

func (a *App) Logout(w http.ResponseWriter, r *http.Request) {
	cu := userFromContext(r.Context())
	if _, err := a.db.Exec(r.Context(), "DELETE FROM sessions WHERE token = $1", cu.Token); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to log out")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
