package main

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestLogin_Success(t *testing.T) {
	truncateTables(t)
	email := "jason@example.com"
	createTestUserViaAPI(t, email)

	rec := doRequest(http.MethodPost, "/login", `{"email":"jason@example.com","password":"password123"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Token string `json:"token"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Token == "" {
		t.Fatalf("expected a session token in the response")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	truncateTables(t)
	createTestUserViaAPI(t, "jason@example.com")

	rec := doRequest(http.MethodPost, "/login", `{"email":"jason@example.com","password":"wrongpassword"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLogin_UnknownEmail(t *testing.T) {
	truncateTables(t)

	rec := doRequest(http.MethodPost, "/login", `{"email":"nobody@example.com","password":"password123"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLogin_MissingFields(t *testing.T) {
	truncateTables(t)

	rec := doRequest(http.MethodPost, "/login", `{"email":"jason@example.com"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLogout_InvalidatesSession(t *testing.T) {
	truncateTables(t)
	_, token := createAndLoginTestUser(t)

	rec := doAuthRequest(http.MethodPost, "/logout", "", token)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = doAuthRequest(http.MethodGet, "/tourdates", "", token)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 after logout, got %d", rec.Code)
	}
}

func TestLogout_RequiresAuth(t *testing.T) {
	truncateTables(t)

	rec := doRequest(http.MethodPost, "/logout", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}
