package tourdates_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"skejio/backend/internal/testutil"
	"skejio/backend/internal/tourdates"
)

func TestMain(m *testing.M) { testutil.Run(m) }

func TestCreateTourDate(t *testing.T) {
	testutil.TruncateTables(t)
	user, token := testutil.CreateAndLoginTestUser(t)

	rec := testutil.DoAuthRequest(http.MethodPost, "/tourdates", `{"date":"2026-09-15","city":"Austin","state":"TX","venue":"Moody Center"}`, token)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var td tourdates.TourDate
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
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)

	rec := testutil.DoAuthRequest(http.MethodPost, "/tourdates", `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`, token)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var td tourdates.TourDate
	json.Unmarshal(rec.Body.Bytes(), &td)
	if td.State != nil {
		t.Fatalf("expected nil state, got %v", *td.State)
	}
}

func TestCreateTourDate_MissingRequiredFields(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)

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
			rec := testutil.DoAuthRequest(http.MethodPost, "/tourdates", c.body, token)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestCreateTourDate_RequiresAuth(t *testing.T) {
	testutil.TruncateTables(t)

	rec := testutil.DoRequest(http.MethodPost, "/tourdates", `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateTourDate_InvalidJSON(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)

	rec := testutil.DoAuthRequest(http.MethodPost, "/tourdates", `not json`, token)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListTourDates_OrderedByDate(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)

	testutil.CreateTestTourDate(t, token, `{"date":"2026-10-01","city":"Dallas","venue":"American Airlines Center"}`)
	testutil.CreateTestTourDate(t, token, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)

	rec := testutil.DoAuthRequest(http.MethodGet, "/tourdates", "", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var tds []tourdates.TourDate
	if err := json.Unmarshal(rec.Body.Bytes(), &tds); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(tds) != 2 {
		t.Fatalf("expected 2 tourdates, got %d", len(tds))
	}
	if tds[0].City != "Austin" || tds[1].City != "Dallas" {
		t.Fatalf("expected results ordered by date, got %s then %s", tds[0].City, tds[1].City)
	}
}

func TestListTourDates_ScopedToOwner(t *testing.T) {
	testutil.TruncateTables(t)
	_, tokenA := testutil.CreateAndLoginTestUser(t)
	_, tokenB := testutil.CreateAndLoginTestUser(t)

	testutil.CreateTestTourDate(t, tokenA, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)
	testutil.CreateTestTourDate(t, tokenB, `{"date":"2026-09-20","city":"Dallas","venue":"American Airlines Center"}`)

	rec := testutil.DoAuthRequest(http.MethodGet, "/tourdates", "", tokenA)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var tds []tourdates.TourDate
	json.Unmarshal(rec.Body.Bytes(), &tds)
	if len(tds) != 1 || tds[0].City != "Austin" {
		t.Fatalf("expected only user A's tourdate, got %+v", tds)
	}
}

func TestGetTourDate(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	created := testutil.CreateTestTourDate(t, token, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)

	rec := testutil.DoAuthRequest(http.MethodGet, "/tourdates/"+created.ID.String(), "", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var td tourdates.TourDate
	json.Unmarshal(rec.Body.Bytes(), &td)
	if td.ID != created.ID {
		t.Fatalf("expected id %s, got %s", created.ID, td.ID)
	}
}

func TestGetTourDate_NotFound(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)

	rec := testutil.DoAuthRequest(http.MethodGet, "/tourdates/11111111-1111-1111-1111-111111111111", "", token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetTourDate_InvalidID(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)

	rec := testutil.DoAuthRequest(http.MethodGet, "/tourdates/not-a-uuid", "", token)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetTourDate_RequiresAuth(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	created := testutil.CreateTestTourDate(t, token, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)

	rec := testutil.DoRequest(http.MethodGet, "/tourdates/"+created.ID.String(), "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetTourDate_OtherUsersTourdate404(t *testing.T) {
	testutil.TruncateTables(t)
	_, tokenA := testutil.CreateAndLoginTestUser(t)
	_, tokenB := testutil.CreateAndLoginTestUser(t)
	created := testutil.CreateTestTourDate(t, tokenA, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)

	rec := testutil.DoAuthRequest(http.MethodGet, "/tourdates/"+created.ID.String(), "", tokenB)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 (not 403, to avoid revealing existence), got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchTourDate_PartialUpdate(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	created := testutil.CreateTestTourDate(t, token, `{"date":"2026-09-15","city":"Austin","state":"TX","venue":"Moody Center"}`)

	rec := testutil.DoAuthRequest(http.MethodPatch, "/tourdates/"+created.ID.String(), `{"venue":"ACL Live"}`, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var td tourdates.TourDate
	json.Unmarshal(rec.Body.Bytes(), &td)
	if td.Venue != "ACL Live" {
		t.Fatalf("expected venue to be updated, got %s", td.Venue)
	}
	if td.City != "Austin" || td.State == nil || *td.State != "TX" {
		t.Fatalf("expected other fields untouched, got %+v", td)
	}
}

func TestPatchTourDate_ClearStateToNull(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	created := testutil.CreateTestTourDate(t, token, `{"date":"2026-09-15","city":"Austin","state":"TX","venue":"Moody Center"}`)

	rec := testutil.DoAuthRequest(http.MethodPatch, "/tourdates/"+created.ID.String(), `{"state":null}`, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var td tourdates.TourDate
	json.Unmarshal(rec.Body.Bytes(), &td)
	if td.State != nil {
		t.Fatalf("expected state to be cleared, got %v", *td.State)
	}
}

func TestPatchTourDate_NotFound(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)

	rec := testutil.DoAuthRequest(http.MethodPatch, "/tourdates/11111111-1111-1111-1111-111111111111", `{"venue":"ACL Live"}`, token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchTourDate_NoFields(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	created := testutil.CreateTestTourDate(t, token, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)

	rec := testutil.DoAuthRequest(http.MethodPatch, "/tourdates/"+created.ID.String(), `{}`, token)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchTourDate_OtherUsersTourdate404(t *testing.T) {
	testutil.TruncateTables(t)
	_, tokenA := testutil.CreateAndLoginTestUser(t)
	_, tokenB := testutil.CreateAndLoginTestUser(t)
	created := testutil.CreateTestTourDate(t, tokenA, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)

	rec := testutil.DoAuthRequest(http.MethodPatch, "/tourdates/"+created.ID.String(), `{"venue":"Hacked"}`, tokenB)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteTourDate(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	created := testutil.CreateTestTourDate(t, token, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)

	rec := testutil.DoAuthRequest(http.MethodDelete, "/tourdates/"+created.ID.String(), "", token)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testutil.DoAuthRequest(http.MethodGet, "/tourdates/"+created.ID.String(), "", token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", rec.Code)
	}
}

func TestDeleteTourDate_NotFound(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)

	rec := testutil.DoAuthRequest(http.MethodDelete, "/tourdates/11111111-1111-1111-1111-111111111111", "", token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteTourDate_OtherUsersTourdate404(t *testing.T) {
	testutil.TruncateTables(t)
	_, tokenA := testutil.CreateAndLoginTestUser(t)
	_, tokenB := testutil.CreateAndLoginTestUser(t)
	created := testutil.CreateTestTourDate(t, tokenA, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)

	rec := testutil.DoAuthRequest(http.MethodDelete, "/tourdates/"+created.ID.String(), "", tokenB)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testutil.DoAuthRequest(http.MethodGet, "/tourdates/"+created.ID.String(), "", tokenA)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected tourdate to survive the other user's failed delete, got %d", rec.Code)
	}
}
