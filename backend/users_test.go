package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func createTestUserViaAPI(t *testing.T, email string) User {
	t.Helper()
	rec := doRequest(http.MethodPost, "/users", fmt.Sprintf(`{"name":"Test User","email":%q,"password":"password123"}`, email))
	if rec.Code != http.StatusCreated {
		t.Fatalf("failed to create user: status %d, body %s", rec.Code, rec.Body.String())
	}
	var u User
	if err := json.Unmarshal(rec.Body.Bytes(), &u); err != nil {
		t.Fatalf("failed to decode created user: %v", err)
	}
	return u
}

func TestCreateUser(t *testing.T) {
	truncateTables(t)

	rec := doRequest(http.MethodPost, "/users", `{"name":"Jason Harris","email":"jason@example.com","password":"password123"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "password") {
		t.Fatalf("response must never include password/password_hash, got: %s", rec.Body.String())
	}
	var u User
	json.Unmarshal(rec.Body.Bytes(), &u)
	if u.Name != "Jason Harris" || u.Email != "jason@example.com" {
		t.Fatalf("unexpected user: %+v", u)
	}
	if u.ID.String() == "" {
		t.Fatalf("expected a generated id")
	}
}

func TestCreateUser_MissingFields(t *testing.T) {
	truncateTables(t)

	cases := []struct {
		name string
		body string
	}{
		{"missing name", `{"email":"jason@example.com","password":"password123"}`},
		{"missing email", `{"name":"Jason Harris","password":"password123"}`},
		{"missing password", `{"name":"Jason Harris","email":"jason@example.com"}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec := doRequest(http.MethodPost, "/users", c.body)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	truncateTables(t)
	createTestUserViaAPI(t, "dup@example.com")

	rec := doRequest(http.MethodPost, "/users", `{"name":"Someone Else","email":"dup@example.com","password":"password123"}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateUser_InvalidJSON(t *testing.T) {
	truncateTables(t)

	rec := doRequest(http.MethodPost, "/users", `not json`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListUsers(t *testing.T) {
	truncateTables(t)
	createTestUserViaAPI(t, "one@example.com")
	createTestUserViaAPI(t, "two@example.com")

	rec := doRequest(http.MethodGet, "/users", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var users []User
	json.Unmarshal(rec.Body.Bytes(), &users)
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
}

func TestGetUser(t *testing.T) {
	truncateTables(t)
	created := createTestUserViaAPI(t, "jason@example.com")

	rec := doRequest(http.MethodGet, "/users/"+created.ID.String(), "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var u User
	json.Unmarshal(rec.Body.Bytes(), &u)
	if u.ID != created.ID {
		t.Fatalf("expected id %s, got %s", created.ID, u.ID)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	truncateTables(t)

	rec := doRequest(http.MethodGet, "/users/11111111-1111-1111-1111-111111111111", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetUser_InvalidID(t *testing.T) {
	truncateTables(t)

	rec := doRequest(http.MethodGet, "/users/not-a-uuid", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchUser_PartialUpdate(t *testing.T) {
	truncateTables(t)
	created := createTestUserViaAPI(t, "jason@example.com")

	rec := doRequest(http.MethodPatch, "/users/"+created.ID.String(), `{"name":"J. Harris"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var u User
	json.Unmarshal(rec.Body.Bytes(), &u)
	if u.Name != "J. Harris" {
		t.Fatalf("expected name updated, got %s", u.Name)
	}
	if u.Email != "jason@example.com" {
		t.Fatalf("expected email untouched, got %s", u.Email)
	}
}

func TestPatchUser_Password(t *testing.T) {
	truncateTables(t)
	created := createTestUserViaAPI(t, "jason@example.com")

	rec := doRequest(http.MethodPatch, "/users/"+created.ID.String(), `{"password":"newpassword456"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var hash string
	err := testPool.QueryRow(context.Background(), "SELECT password_hash FROM users WHERE id = $1", created.ID).Scan(&hash)
	if err != nil {
		t.Fatalf("failed to read password_hash: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("newpassword456")); err != nil {
		t.Fatalf("password_hash does not match new password: %v", err)
	}
}

func TestPatchUser_DuplicateEmail(t *testing.T) {
	truncateTables(t)
	createTestUserViaAPI(t, "taken@example.com")
	created := createTestUserViaAPI(t, "jason@example.com")

	rec := doRequest(http.MethodPatch, "/users/"+created.ID.String(), `{"email":"taken@example.com"}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchUser_NotFound(t *testing.T) {
	truncateTables(t)

	rec := doRequest(http.MethodPatch, "/users/11111111-1111-1111-1111-111111111111", `{"name":"Nobody"}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchUser_NoFields(t *testing.T) {
	truncateTables(t)
	created := createTestUserViaAPI(t, "jason@example.com")

	rec := doRequest(http.MethodPatch, "/users/"+created.ID.String(), `{}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteUser(t *testing.T) {
	truncateTables(t)
	created := createTestUserViaAPI(t, "jason@example.com")

	rec := doRequest(http.MethodDelete, "/users/"+created.ID.String(), "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = doRequest(http.MethodGet, "/users/"+created.ID.String(), "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", rec.Code)
	}
}

func TestDeleteUser_NotFound(t *testing.T) {
	truncateTables(t)

	rec := doRequest(http.MethodDelete, "/users/11111111-1111-1111-1111-111111111111", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteUser_CascadesTourdates(t *testing.T) {
	truncateTables(t)
	created := createTestUserViaAPI(t, "jason@example.com")
	tourdate := createTestTourDate(t, created.ID, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)

	rec := doRequest(http.MethodDelete, "/users/"+created.ID.String(), "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = doRequest(http.MethodGet, "/tourdates/"+tourdate.ID.String(), "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected tourdate to be cascade-deleted (404), got %d", rec.Code)
	}
}
