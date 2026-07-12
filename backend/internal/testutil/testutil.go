// Package testutil is the shared test harness for every internal/<domain>
// package's tests: it builds the same router production uses (via
// internal/api), connects it to the skejio_test database, and exposes
// request/fixture helpers so each domain's tests don't duplicate this setup.
//
// All domain packages share one physical skejio_test database, so their test
// binaries must not run concurrently - `go test ./...` runs different
// packages' binaries in parallel by default, which races truncates/inserts
// across packages. Always run the suite with `go test -p 1 ./...`.
package testutil

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"

	"skejio/backend/internal/api"
	"skejio/backend/internal/auth"
	"skejio/backend/internal/representatives"
	"skejio/backend/internal/tourdates"
	"skejio/backend/internal/users"
)

const defaultTestDatabaseURL = "postgres://jasonharris@localhost:5432/skejio_test?sslmode=disable"

const TestUserPassword = "password123"

var (
	Pool   *pgxpool.Pool
	Router *mux.Router
)

// Run connects to the test database, builds the production router, and
// executes the package's tests. Call it as the entire body of a package's
// TestMain: func TestMain(m *testing.M) { testutil.Run(m) }
func Run(m *testing.M) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = defaultTestDatabaseURL
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		panic("unable to connect to test database: " + err.Error())
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		panic("test database unreachable: " + err.Error())
	}

	Pool = pool
	Router = api.NewRouter(
		&auth.Handler{DB: pool},
		&tourdates.Handler{DB: pool},
		&users.Handler{DB: pool},
		&representatives.Handler{DB: pool},
	)

	os.Exit(m.Run())
}

func TruncateTables(t *testing.T) {
	t.Helper()
	if _, err := Pool.Exec(context.Background(), "TRUNCATE TABLE tourdates, users, sessions, artist_representatives"); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}
}

// DoRequest issues an unauthenticated request (for /login, /users, /test, /).
func DoRequest(method, path, body string) *httptest.ResponseRecorder {
	return DoAuthRequest(method, path, body, "")
}

// DoAuthRequest issues a request with an optional Bearer token.
func DoAuthRequest(method, path, body, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	Router.ServeHTTP(rec, req)
	return rec
}

func CreateTestUserViaAPI(t *testing.T, email string) users.User {
	t.Helper()
	rec := DoRequest(http.MethodPost, "/users", fmt.Sprintf(
		`{"name":"Test User","email":%q,"password":%q,"user_type":"ARTIST"}`, email, TestUserPassword))
	if rec.Code != http.StatusCreated {
		t.Fatalf("failed to create user: status %d, body %s", rec.Code, rec.Body.String())
	}
	var u users.User
	if err := json.Unmarshal(rec.Body.Bytes(), &u); err != nil {
		t.Fatalf("failed to decode created user: %v", err)
	}
	return u
}

// CreateTestUserOfType creates and logs in a user of the given user_type,
// returning the user and their session token.
func CreateTestUserOfType(t *testing.T, userType string) (users.User, string) {
	t.Helper()
	email := fmt.Sprintf("%s@example.com", uuid.NewString())
	rec := DoRequest(http.MethodPost, "/users", fmt.Sprintf(
		`{"name":"Test User","email":%q,"password":%q,"user_type":%q}`, email, TestUserPassword, userType))
	if rec.Code != http.StatusCreated {
		t.Fatalf("failed to create %s user: status %d, body %s", userType, rec.Code, rec.Body.String())
	}
	var u users.User
	if err := json.Unmarshal(rec.Body.Bytes(), &u); err != nil {
		t.Fatalf("failed to decode created user: %v", err)
	}
	token := LoginTestUser(t, email, TestUserPassword)
	return u, token
}

func LoginTestUser(t *testing.T, email, password string) string {
	t.Helper()
	rec := DoRequest(http.MethodPost, "/login", fmt.Sprintf(`{"email":%q,"password":%q}`, email, password))
	if rec.Code != http.StatusOK {
		t.Fatalf("failed to log in: status %d, body %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}
	return resp.Token
}

// CreateAndLoginTestUser creates a user via the real /users endpoint and logs
// them in via the real /login endpoint, returning the user and their session
// token for use in Authorization headers.
func CreateAndLoginTestUser(t *testing.T) (users.User, string) {
	t.Helper()
	email := fmt.Sprintf("%s@example.com", uuid.NewString())
	user := CreateTestUserViaAPI(t, email)
	token := LoginTestUser(t, email, TestUserPassword)
	return user, token
}

func CreateTestTourDate(t *testing.T, token, body string) tourdates.TourDate {
	t.Helper()
	rec := DoAuthRequest(http.MethodPost, "/tourdates", body, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("failed to create tourdate: status %d, body %s", rec.Code, rec.Body.String())
	}
	var td tourdates.TourDate
	if err := json.Unmarshal(rec.Body.Bytes(), &td); err != nil {
		t.Fatalf("failed to decode created tourdate: %v", err)
	}
	return td
}

func AddRepresentative(t *testing.T, artistToken string, representativeID uuid.UUID) representatives.RepresentedUser {
	t.Helper()
	rec := DoAuthRequest(http.MethodPost, "/representatives", fmt.Sprintf(`{"representative_id":%q}`, representativeID), artistToken)
	if rec.Code != http.StatusCreated {
		t.Fatalf("failed to create representative: status %d, body %s", rec.Code, rec.Body.String())
	}
	var ru representatives.RepresentedUser
	if err := json.Unmarshal(rec.Body.Bytes(), &ru); err != nil {
		t.Fatalf("failed to decode representative: %v", err)
	}
	return ru
}
