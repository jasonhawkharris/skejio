package tours_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"skejio/backend/internal/testutil"
	"skejio/backend/internal/tours"
)

func TestMain(m *testing.M) { testutil.Run(m) }

func TestCreateTour(t *testing.T) {
	testutil.TruncateTables(t)
	user, token := testutil.CreateAndLoginTestUser(t)

	rec := testutil.DoAuthRequest(http.MethodPost, "/tours", `{"name":"Fall Tour"}`, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var tr tours.Tour
	if err := json.Unmarshal(rec.Body.Bytes(), &tr); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if tr.Name != "Fall Tour" {
		t.Fatalf("expected name Fall Tour, got %s", tr.Name)
	}
	if tr.UserID != user.ID {
		t.Fatalf("expected user_id %s, got %s", user.ID, tr.UserID)
	}
}

func TestCreateTour_MissingName(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)

	rec := testutil.DoAuthRequest(http.MethodPost, "/tours", `{}`, token)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateTour_RequiresAuth(t *testing.T) {
	testutil.TruncateTables(t)

	rec := testutil.DoRequest(http.MethodPost, "/tours", `{"name":"Fall Tour"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListTours_ScopedToOwner(t *testing.T) {
	testutil.TruncateTables(t)
	_, tokenA := testutil.CreateAndLoginTestUser(t)
	_, tokenB := testutil.CreateAndLoginTestUser(t)

	testutil.CreateTestTour(t, tokenA, `{"name":"A's Tour"}`)
	testutil.CreateTestTour(t, tokenB, `{"name":"B's Tour"}`)

	rec := testutil.DoAuthRequest(http.MethodGet, "/tours", "", tokenA)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var trs []tours.Tour
	json.Unmarshal(rec.Body.Bytes(), &trs)
	if len(trs) != 1 || trs[0].Name != "A's Tour" {
		t.Fatalf("expected only user A's tour, got %+v", trs)
	}
}

func TestGetTour_OtherUsersTour404(t *testing.T) {
	testutil.TruncateTables(t)
	_, tokenA := testutil.CreateAndLoginTestUser(t)
	_, tokenB := testutil.CreateAndLoginTestUser(t)
	created := testutil.CreateTestTour(t, tokenA, `{"name":"A's Tour"}`)

	rec := testutil.DoAuthRequest(http.MethodGet, "/tours/"+created.ID.String(), "", tokenB)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 (not 403, to avoid revealing existence), got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetTour_NotFound(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)

	rec := testutil.DoAuthRequest(http.MethodGet, "/tours/11111111-1111-1111-1111-111111111111", "", token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchTour_PartialUpdate(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	created := testutil.CreateTestTour(t, token, `{"name":"Fall Tour"}`)

	rec := testutil.DoAuthRequest(http.MethodPatch, "/tours/"+created.ID.String(), `{"name":"Winter Tour"}`, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var tr tours.Tour
	json.Unmarshal(rec.Body.Bytes(), &tr)
	if tr.Name != "Winter Tour" {
		t.Fatalf("expected name to be updated, got %s", tr.Name)
	}
}

func TestPatchTour_OtherUsersTour404(t *testing.T) {
	testutil.TruncateTables(t)
	_, tokenA := testutil.CreateAndLoginTestUser(t)
	_, tokenB := testutil.CreateAndLoginTestUser(t)
	created := testutil.CreateTestTour(t, tokenA, `{"name":"A's Tour"}`)

	rec := testutil.DoAuthRequest(http.MethodPatch, "/tours/"+created.ID.String(), `{"name":"Hacked"}`, tokenB)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteTour_SetsTourdatesTourIDToNull(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tour := testutil.CreateTestTour(t, token, `{"name":"Fall Tour"}`)
	created := testutil.CreateTestTourDate(t, token,
		`{"date":"2026-09-15","city":"Austin","venue":"Moody Center","tour_id":"`+tour.ID.String()+`"}`)

	rec := testutil.DoAuthRequest(http.MethodDelete, "/tours/"+tour.ID.String(), "", token)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testutil.DoAuthRequest(http.MethodGet, "/tourdates/"+created.ID.String(), "", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var td struct {
		TourID *string `json:"tour_id"`
	}
	json.Unmarshal(rec.Body.Bytes(), &td)
	if td.TourID != nil {
		t.Fatalf("expected tour_id to be nulled out after tour delete, got %v", *td.TourID)
	}
}

func TestDeleteTour_OtherUsersTour404(t *testing.T) {
	testutil.TruncateTables(t)
	_, tokenA := testutil.CreateAndLoginTestUser(t)
	_, tokenB := testutil.CreateAndLoginTestUser(t)
	created := testutil.CreateTestTour(t, tokenA, `{"name":"A's Tour"}`)

	rec := testutil.DoAuthRequest(http.MethodDelete, "/tours/"+created.ID.String(), "", tokenB)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testutil.DoAuthRequest(http.MethodGet, "/tours/"+created.ID.String(), "", tokenA)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected tour to survive the other user's failed delete, got %d", rec.Code)
	}
}
