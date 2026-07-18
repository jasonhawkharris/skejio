package auth_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"skejio/backend/internal/testutil"
)

func TestMain(m *testing.M) { testutil.Run(m) }

func TestLogin_Success(t *testing.T) {
	testutil.TruncateTables(t)
	email := "jason@example.com"
	testutil.CreateTestUserViaAPI(t, email)

	rec := testutil.DoRequest(http.MethodPost, "/login", `{"email":"jason@example.com","password":"password123"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Token    string `json:"token"`
		UserType string `json:"user_type"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Token == "" {
		t.Fatalf("expected a session token in the response")
	}
	if resp.UserType != "ARTIST" {
		t.Fatalf("expected user_type ARTIST, got %q", resp.UserType)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	testutil.TruncateTables(t)
	testutil.CreateTestUserViaAPI(t, "jason@example.com")

	rec := testutil.DoRequest(http.MethodPost, "/login", `{"email":"jason@example.com","password":"wrongpassword"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLogin_UnknownEmail(t *testing.T) {
	testutil.TruncateTables(t)

	rec := testutil.DoRequest(http.MethodPost, "/login", `{"email":"nobody@example.com","password":"password123"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLogin_MissingFields(t *testing.T) {
	testutil.TruncateTables(t)

	rec := testutil.DoRequest(http.MethodPost, "/login", `{"email":"jason@example.com"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLogout_InvalidatesSession(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)

	rec := testutil.DoAuthRequest(http.MethodPost, "/logout", "", token)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testutil.DoAuthRequest(http.MethodGet, "/tourdates", "", token)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 after logout, got %d", rec.Code)
	}
}

func TestLogout_RequiresAuth(t *testing.T) {
	testutil.TruncateTables(t)

	rec := testutil.DoRequest(http.MethodPost, "/logout", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}
