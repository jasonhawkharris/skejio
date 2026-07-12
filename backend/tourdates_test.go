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
	if _, err := testPool.Exec(context.Background(), "TRUNCATE TABLE tourdates, users"); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}
}

func createTestUser(t *testing.T) uuid.UUID {
	t.Helper()
	return createTestUserViaAPI(t, fmt.Sprintf("%s@example.com", uuid.NewString())).ID
}

func doRequest(method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)
	return rec
}

func createTestTourDate(t *testing.T, userID uuid.UUID, body string) TourDate {
	t.Helper()
	rec := doRequest(http.MethodPost, "/tourdates", withUserID(body, userID))
	if rec.Code != http.StatusCreated {
		t.Fatalf("failed to create tourdate: status %d, body %s", rec.Code, rec.Body.String())
	}
	var td TourDate
	if err := json.Unmarshal(rec.Body.Bytes(), &td); err != nil {
		t.Fatalf("failed to decode created tourdate: %v", err)
	}
	return td
}

// withUserID merges a "user_id" key into a JSON object literal used in tests.
func withUserID(body string, userID uuid.UUID) string {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(body), &fields); err != nil {
		panic("withUserID: invalid JSON body: " + err.Error())
	}
	idJSON, _ := json.Marshal(userID.String())
	fields["user_id"] = idJSON
	merged, err := json.Marshal(fields)
	if err != nil {
		panic("withUserID: failed to remarshal body: " + err.Error())
	}
	return string(merged)
}

func TestCreateTourDate(t *testing.T) {
	truncateTables(t)
	userID := createTestUser(t)

	rec := doRequest(http.MethodPost, "/tourdates", withUserID(`{"date":"2026-09-15","city":"Austin","state":"TX","venue":"Moody Center"}`, userID))

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
	if td.UserID != userID {
		t.Fatalf("expected user_id %s, got %s", userID, td.UserID)
	}
	if td.ID.String() == "" {
		t.Fatalf("expected a generated id")
	}
}

func TestCreateTourDate_NullableState(t *testing.T) {
	truncateTables(t)
	userID := createTestUser(t)

	rec := doRequest(http.MethodPost, "/tourdates", withUserID(`{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`, userID))

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
	userID := createTestUser(t)

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
			rec := doRequest(http.MethodPost, "/tourdates", withUserID(c.body, userID))
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestCreateTourDate_MissingUserID(t *testing.T) {
	truncateTables(t)

	rec := doRequest(http.MethodPost, "/tourdates", `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateTourDate_UnknownUserID(t *testing.T) {
	truncateTables(t)

	rec := doRequest(http.MethodPost, "/tourdates", `{"date":"2026-09-15","city":"Austin","venue":"Moody Center","user_id":"11111111-1111-1111-1111-111111111111"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateTourDate_InvalidJSON(t *testing.T) {
	truncateTables(t)

	rec := doRequest(http.MethodPost, "/tourdates", `not json`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListTourDates(t *testing.T) {
	truncateTables(t)
	userID := createTestUser(t)

	createTestTourDate(t, userID, `{"date":"2026-10-01","city":"Dallas","venue":"American Airlines Center"}`)
	createTestTourDate(t, userID, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)

	rec := doRequest(http.MethodGet, "/tourdates", "")
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

func TestGetTourDate(t *testing.T) {
	truncateTables(t)
	userID := createTestUser(t)
	created := createTestTourDate(t, userID, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)

	rec := doRequest(http.MethodGet, "/tourdates/"+created.ID.String(), "")
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

	rec := doRequest(http.MethodGet, "/tourdates/11111111-1111-1111-1111-111111111111", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetTourDate_InvalidID(t *testing.T) {
	truncateTables(t)

	rec := doRequest(http.MethodGet, "/tourdates/not-a-uuid", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchTourDate_PartialUpdate(t *testing.T) {
	truncateTables(t)
	userID := createTestUser(t)
	created := createTestTourDate(t, userID, `{"date":"2026-09-15","city":"Austin","state":"TX","venue":"Moody Center"}`)

	rec := doRequest(http.MethodPatch, "/tourdates/"+created.ID.String(), `{"venue":"ACL Live"}`)
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
	userID := createTestUser(t)
	created := createTestTourDate(t, userID, `{"date":"2026-09-15","city":"Austin","state":"TX","venue":"Moody Center"}`)

	rec := doRequest(http.MethodPatch, "/tourdates/"+created.ID.String(), `{"state":null}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var td TourDate
	json.Unmarshal(rec.Body.Bytes(), &td)
	if td.State != nil {
		t.Fatalf("expected state to be cleared, got %v", *td.State)
	}
}

func TestPatchTourDate_Reassign(t *testing.T) {
	truncateTables(t)
	userID := createTestUser(t)
	otherUserID := createTestUser(t)
	created := createTestTourDate(t, userID, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)

	rec := doRequest(http.MethodPatch, "/tourdates/"+created.ID.String(), fmt.Sprintf(`{"user_id":"%s"}`, otherUserID))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var td TourDate
	json.Unmarshal(rec.Body.Bytes(), &td)
	if td.UserID != otherUserID {
		t.Fatalf("expected user_id %s, got %s", otherUserID, td.UserID)
	}
}

func TestPatchTourDate_UnknownUserID(t *testing.T) {
	truncateTables(t)
	userID := createTestUser(t)
	created := createTestTourDate(t, userID, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)

	rec := doRequest(http.MethodPatch, "/tourdates/"+created.ID.String(), `{"user_id":"11111111-1111-1111-1111-111111111111"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchTourDate_NotFound(t *testing.T) {
	truncateTables(t)

	rec := doRequest(http.MethodPatch, "/tourdates/11111111-1111-1111-1111-111111111111", `{"venue":"ACL Live"}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchTourDate_NoFields(t *testing.T) {
	truncateTables(t)
	userID := createTestUser(t)
	created := createTestTourDate(t, userID, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)

	rec := doRequest(http.MethodPatch, "/tourdates/"+created.ID.String(), `{}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteTourDate(t *testing.T) {
	truncateTables(t)
	userID := createTestUser(t)
	created := createTestTourDate(t, userID, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)

	rec := doRequest(http.MethodDelete, "/tourdates/"+created.ID.String(), "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = doRequest(http.MethodGet, "/tourdates/"+created.ID.String(), "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", rec.Code)
	}
}

func TestDeleteTourDate_NotFound(t *testing.T) {
	truncateTables(t)

	rec := doRequest(http.MethodDelete, "/tourdates/11111111-1111-1111-1111-111111111111", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}
