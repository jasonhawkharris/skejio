package riders_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"skejio/backend/internal/riders"
	"skejio/backend/internal/testutil"
)

func TestMain(m *testing.M) { testutil.Run(m) }

func TestCreateRider(t *testing.T) {
	testutil.TruncateTables(t)
	user, token := testutil.CreateAndLoginTestUser(t)

	rec := testutil.DoAuthRequest(http.MethodPost, "/riders", `{"name":"Club Rider","content":"2 mics, 4 monitors"}`, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var rd riders.Rider
	if err := json.Unmarshal(rec.Body.Bytes(), &rd); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if rd.Name != "Club Rider" || rd.Content != "2 mics, 4 monitors" {
		t.Fatalf("unexpected rider: %+v", rd)
	}
	if rd.UserID != user.ID {
		t.Fatalf("expected user_id %s, got %s", user.ID, rd.UserID)
	}
}

func TestCreateRider_MissingFields(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)

	cases := []struct {
		name string
		body string
	}{
		{"missing name", `{"content":"2 mics"}`},
		{"missing content", `{"name":"Club Rider"}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec := testutil.DoAuthRequest(http.MethodPost, "/riders", c.body, token)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestCreateRider_RequiresAuth(t *testing.T) {
	testutil.TruncateTables(t)

	rec := testutil.DoRequest(http.MethodPost, "/riders", `{"name":"Club Rider","content":"2 mics"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListRiders_ScopedToOwner(t *testing.T) {
	testutil.TruncateTables(t)
	_, tokenA := testutil.CreateAndLoginTestUser(t)
	_, tokenB := testutil.CreateAndLoginTestUser(t)

	testutil.CreateTestRider(t, tokenA, `{"name":"A's Rider","content":"stuff"}`)
	testutil.CreateTestRider(t, tokenB, `{"name":"B's Rider","content":"stuff"}`)

	rec := testutil.DoAuthRequest(http.MethodGet, "/riders", "", tokenA)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var rds []riders.Rider
	json.Unmarshal(rec.Body.Bytes(), &rds)
	if len(rds) != 1 || rds[0].Name != "A's Rider" {
		t.Fatalf("expected only user A's rider, got %+v", rds)
	}
}

func TestGetRider_OtherUsersRider404(t *testing.T) {
	testutil.TruncateTables(t)
	_, tokenA := testutil.CreateAndLoginTestUser(t)
	_, tokenB := testutil.CreateAndLoginTestUser(t)
	created := testutil.CreateTestRider(t, tokenA, `{"name":"A's Rider","content":"stuff"}`)

	rec := testutil.DoAuthRequest(http.MethodGet, "/riders/"+created.ID.String(), "", tokenB)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 (not 403, to avoid revealing existence), got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetRider_NotFound(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)

	rec := testutil.DoAuthRequest(http.MethodGet, "/riders/11111111-1111-1111-1111-111111111111", "", token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchRider_PartialUpdate(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	created := testutil.CreateTestRider(t, token, `{"name":"Club Rider","content":"2 mics"}`)

	rec := testutil.DoAuthRequest(http.MethodPatch, "/riders/"+created.ID.String(), `{"content":"4 mics, catering"}`, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var rd riders.Rider
	json.Unmarshal(rec.Body.Bytes(), &rd)
	if rd.Content != "4 mics, catering" {
		t.Fatalf("expected content to be updated, got %s", rd.Content)
	}
	if rd.Name != "Club Rider" {
		t.Fatalf("expected name untouched, got %s", rd.Name)
	}
}

func TestPatchRider_CannotClearRequiredFields(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	created := testutil.CreateTestRider(t, token, `{"name":"Club Rider","content":"2 mics"}`)

	rec := testutil.DoAuthRequest(http.MethodPatch, "/riders/"+created.ID.String(), `{"content":null}`, token)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchRider_OtherUsersRider404(t *testing.T) {
	testutil.TruncateTables(t)
	_, tokenA := testutil.CreateAndLoginTestUser(t)
	_, tokenB := testutil.CreateAndLoginTestUser(t)
	created := testutil.CreateTestRider(t, tokenA, `{"name":"A's Rider","content":"stuff"}`)

	rec := testutil.DoAuthRequest(http.MethodPatch, "/riders/"+created.ID.String(), `{"name":"Hacked"}`, tokenB)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteRider_SetsTourdatesRiderIDToNull(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	rider := testutil.CreateTestRider(t, token, `{"name":"Club Rider","content":"2 mics"}`)
	created := testutil.CreateTestTourDate(t, token,
		`{"date":"2026-09-15","city":"Austin","venue":"Moody Center","rider_id":"`+rider.ID.String()+`"}`)

	rec := testutil.DoAuthRequest(http.MethodDelete, "/riders/"+rider.ID.String(), "", token)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testutil.DoAuthRequest(http.MethodGet, "/tourdates/"+created.ID.String(), "", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var td struct {
		RiderID *string `json:"rider_id"`
	}
	json.Unmarshal(rec.Body.Bytes(), &td)
	if td.RiderID != nil {
		t.Fatalf("expected rider_id to be nulled out after rider delete, got %v", *td.RiderID)
	}
}

func TestDeleteRider_OtherUsersRider404(t *testing.T) {
	testutil.TruncateTables(t)
	_, tokenA := testutil.CreateAndLoginTestUser(t)
	_, tokenB := testutil.CreateAndLoginTestUser(t)
	created := testutil.CreateTestRider(t, tokenA, `{"name":"A's Rider","content":"stuff"}`)

	rec := testutil.DoAuthRequest(http.MethodDelete, "/riders/"+created.ID.String(), "", tokenB)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testutil.DoAuthRequest(http.MethodGet, "/riders/"+created.ID.String(), "", tokenA)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected rider to survive the other user's failed delete, got %d", rec.Code)
	}
}
