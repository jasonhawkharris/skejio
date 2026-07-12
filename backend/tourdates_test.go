package main

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
)

const defaultTestDatabaseURL = "postgres://jasonharris@localhost:5432/skejio_test?sslmode=disable"

var (
	testPool   *pgxpool.Pool
	testRouter *mux.Router
)

func TestMain(m *testing.M) {
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

	testPool = pool
	testRouter = newRouter(&App{db: pool})

	os.Exit(m.Run())
}

func truncateTables(t *testing.T) {
	t.Helper()
	if _, err := testPool.Exec(context.Background(), "TRUNCATE TABLE tourdates, users, sessions, artist_representatives"); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}
}

const testUserPassword = "password123"

// createAndLoginTestUser creates a user via the real /users endpoint and logs
// them in via the real /login endpoint, returning the user and their session
// token for use in Authorization headers.
func createAndLoginTestUser(t *testing.T) (User, string) {
	t.Helper()
	email := fmt.Sprintf("%s@example.com", uuid.NewString())
	user := createTestUserViaAPI(t, email)
	token := loginTestUser(t, email, testUserPassword)
	return user, token
}

func loginTestUser(t *testing.T, email, password string) string {
	t.Helper()
	rec := doRequest(http.MethodPost, "/login", fmt.Sprintf(`{"email":%q,"password":%q}`, email, password))
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

// doRequest issues an unauthenticated request (for /login, /users, /test, /).
func doRequest(method, path, body string) *httptest.ResponseRecorder {
	return doAuthRequest(method, path, body, "")
}

// doAuthRequest issues a request with an optional Bearer token.
func doAuthRequest(method, path, body, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)
	return rec
}

func createTestTourDate(t *testing.T, token, body string) TourDate {
	t.Helper()
	rec := doAuthRequest(http.MethodPost, "/tourdates", body, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("failed to create tourdate: status %d, body %s", rec.Code, rec.Body.String())
	}
	var td TourDate
	if err := json.Unmarshal(rec.Body.Bytes(), &td); err != nil {
		t.Fatalf("failed to decode created tourdate: %v", err)
	}
	return td
}

func TestCreateTourDate(t *testing.T) {
	truncateTables(t)
	user, token := createAndLoginTestUser(t)

	rec := doAuthRequest(http.MethodPost, "/tourdates", `{"date":"2026-09-15","city":"Austin","state":"TX","venue":"Moody Center"}`, token)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var td TourDate
	if err := json.Unmarshal(rec.Body.Bytes(), &td); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if td.City != "Austin" || td.Venue != "Moody Center" || td.State == nil || *td.State != "TX" {
		t.Fatalf("unexpected tourdate: %+v", td)
	}
	if td.UserID != user.ID {
		t.Fatalf("expected user_id %s, got %s", user.ID, td.UserID)
	}
	if td.ID.String() == "" {
		t.Fatalf("expected a generated id")
	}
}

func TestCreateTourDate_NullableState(t *testing.T) {
	truncateTables(t)
	_, token := createAndLoginTestUser(t)

	rec := doAuthRequest(http.MethodPost, "/tourdates", `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`, token)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var td TourDate
	json.Unmarshal(rec.Body.Bytes(), &td)
	if td.State != nil {
		t.Fatalf("expected nil state, got %v", *td.State)
	}
}

func TestCreateTourDate_MissingRequiredFields(t *testing.T) {
	truncateTables(t)
	_, token := createAndLoginTestUser(t)

	cases := []struct {
		name string
		body string
	}{
		{"missing date", `{"city":"Austin","venue":"Moody Center"}`},
		{"missing city", `{"date":"2026-09-15","venue":"Moody Center"}`},
		{"missing venue", `{"date":"2026-09-15","city":"Austin"}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec := doAuthRequest(http.MethodPost, "/tourdates", c.body, token)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestCreateTourDate_RequiresAuth(t *testing.T) {
	truncateTables(t)

	rec := doRequest(http.MethodPost, "/tourdates", `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateTourDate_InvalidJSON(t *testing.T) {
	truncateTables(t)
	_, token := createAndLoginTestUser(t)

	rec := doAuthRequest(http.MethodPost, "/tourdates", `not json`, token)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListTourDates_OrderedByDate(t *testing.T) {
	truncateTables(t)
	_, token := createAndLoginTestUser(t)

	createTestTourDate(t, token, `{"date":"2026-10-01","city":"Dallas","venue":"American Airlines Center"}`)
	createTestTourDate(t, token, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)

	rec := doAuthRequest(http.MethodGet, "/tourdates", "", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var tourdates []TourDate
	if err := json.Unmarshal(rec.Body.Bytes(), &tourdates); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(tourdates) != 2 {
		t.Fatalf("expected 2 tourdates, got %d", len(tourdates))
	}
	if tourdates[0].City != "Austin" || tourdates[1].City != "Dallas" {
		t.Fatalf("expected results ordered by date, got %s then %s", tourdates[0].City, tourdates[1].City)
	}
}

func TestListTourDates_ScopedToOwner(t *testing.T) {
	truncateTables(t)
	_, tokenA := createAndLoginTestUser(t)
	_, tokenB := createAndLoginTestUser(t)

	createTestTourDate(t, tokenA, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)
	createTestTourDate(t, tokenB, `{"date":"2026-09-20","city":"Dallas","venue":"American Airlines Center"}`)

	rec := doAuthRequest(http.MethodGet, "/tourdates", "", tokenA)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var tourdates []TourDate
	json.Unmarshal(rec.Body.Bytes(), &tourdates)
	if len(tourdates) != 1 || tourdates[0].City != "Austin" {
		t.Fatalf("expected only user A's tourdate, got %+v", tourdates)
	}
}

func TestGetTourDate(t *testing.T) {
	truncateTables(t)
	_, token := createAndLoginTestUser(t)
	created := createTestTourDate(t, token, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)

	rec := doAuthRequest(http.MethodGet, "/tourdates/"+created.ID.String(), "", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var td TourDate
	json.Unmarshal(rec.Body.Bytes(), &td)
	if td.ID != created.ID {
		t.Fatalf("expected id %s, got %s", created.ID, td.ID)
	}
}

func TestGetTourDate_NotFound(t *testing.T) {
	truncateTables(t)
	_, token := createAndLoginTestUser(t)

	rec := doAuthRequest(http.MethodGet, "/tourdates/11111111-1111-1111-1111-111111111111", "", token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetTourDate_InvalidID(t *testing.T) {
	truncateTables(t)
	_, token := createAndLoginTestUser(t)

	rec := doAuthRequest(http.MethodGet, "/tourdates/not-a-uuid", "", token)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetTourDate_RequiresAuth(t *testing.T) {
	truncateTables(t)
	_, token := createAndLoginTestUser(t)
	created := createTestTourDate(t, token, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)

	rec := doRequest(http.MethodGet, "/tourdates/"+created.ID.String(), "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetTourDate_OtherUsersTourdate404(t *testing.T) {
	truncateTables(t)
	_, tokenA := createAndLoginTestUser(t)
	_, tokenB := createAndLoginTestUser(t)
	created := createTestTourDate(t, tokenA, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)

	rec := doAuthRequest(http.MethodGet, "/tourdates/"+created.ID.String(), "", tokenB)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 (not 403, to avoid revealing existence), got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchTourDate_PartialUpdate(t *testing.T) {
	truncateTables(t)
	_, token := createAndLoginTestUser(t)
	created := createTestTourDate(t, token, `{"date":"2026-09-15","city":"Austin","state":"TX","venue":"Moody Center"}`)

	rec := doAuthRequest(http.MethodPatch, "/tourdates/"+created.ID.String(), `{"venue":"ACL Live"}`, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var td TourDate
	json.Unmarshal(rec.Body.Bytes(), &td)
	if td.Venue != "ACL Live" {
		t.Fatalf("expected venue to be updated, got %s", td.Venue)
	}
	if td.City != "Austin" || td.State == nil || *td.State != "TX" {
		t.Fatalf("expected other fields untouched, got %+v", td)
	}
}

func TestPatchTourDate_ClearStateToNull(t *testing.T) {
	truncateTables(t)
	_, token := createAndLoginTestUser(t)
	created := createTestTourDate(t, token, `{"date":"2026-09-15","city":"Austin","state":"TX","venue":"Moody Center"}`)

	rec := doAuthRequest(http.MethodPatch, "/tourdates/"+created.ID.String(), `{"state":null}`, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var td TourDate
	json.Unmarshal(rec.Body.Bytes(), &td)
	if td.State != nil {
		t.Fatalf("expected state to be cleared, got %v", *td.State)
	}
}

func TestPatchTourDate_NotFound(t *testing.T) {
	truncateTables(t)
	_, token := createAndLoginTestUser(t)

	rec := doAuthRequest(http.MethodPatch, "/tourdates/11111111-1111-1111-1111-111111111111", `{"venue":"ACL Live"}`, token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchTourDate_NoFields(t *testing.T) {
	truncateTables(t)
	_, token := createAndLoginTestUser(t)
	created := createTestTourDate(t, token, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)

	rec := doAuthRequest(http.MethodPatch, "/tourdates/"+created.ID.String(), `{}`, token)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchTourDate_OtherUsersTourdate404(t *testing.T) {
	truncateTables(t)
	_, tokenA := createAndLoginTestUser(t)
	_, tokenB := createAndLoginTestUser(t)
	created := createTestTourDate(t, tokenA, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)

	rec := doAuthRequest(http.MethodPatch, "/tourdates/"+created.ID.String(), `{"venue":"Hacked"}`, tokenB)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteTourDate(t *testing.T) {
	truncateTables(t)
	_, token := createAndLoginTestUser(t)
	created := createTestTourDate(t, token, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)

	rec := doAuthRequest(http.MethodDelete, "/tourdates/"+created.ID.String(), "", token)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = doAuthRequest(http.MethodGet, "/tourdates/"+created.ID.String(), "", token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", rec.Code)
	}
}

func TestDeleteTourDate_NotFound(t *testing.T) {
	truncateTables(t)
	_, token := createAndLoginTestUser(t)

	rec := doAuthRequest(http.MethodDelete, "/tourdates/11111111-1111-1111-1111-111111111111", "", token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteTourDate_OtherUsersTourdate404(t *testing.T) {
	truncateTables(t)
	_, tokenA := createAndLoginTestUser(t)
	_, tokenB := createAndLoginTestUser(t)
	created := createTestTourDate(t, tokenA, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)

	rec := doAuthRequest(http.MethodDelete, "/tourdates/"+created.ID.String(), "", tokenB)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = doAuthRequest(http.MethodGet, "/tourdates/"+created.ID.String(), "", tokenA)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected tourdate to survive the other user's failed delete, got %d", rec.Code)
	}
}
