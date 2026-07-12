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
	rec := doRequest(http.MethodPost, "/users", fmt.Sprintf(`{"name":"Test User","email":%q,"password":%q,"user_type":"ARTIST"}`, email, testUserPassword))
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

	rec := doRequest(http.MethodPost, "/users", `{"name":"Jason Harris","email":"jason@example.com","password":"password123","user_type":"MANAGER"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "password") {
		t.Fatalf("response must never include password/password_hash, got: %s", rec.Body.String())
	}
	var u User
	json.Unmarshal(rec.Body.Bytes(), &u)
	if u.Name != "Jason Harris" || u.Email != "jason@example.com" || u.UserType != "MANAGER" {
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
		{"missing name", `{"email":"jason@example.com","password":"password123","user_type":"ARTIST"}`},
		{"missing email", `{"name":"Jason Harris","password":"password123","user_type":"ARTIST"}`},
		{"missing password", `{"name":"Jason Harris","email":"jason@example.com","user_type":"ARTIST"}`},
		{"missing user_type", `{"name":"Jason Harris","email":"jason@example.com","password":"password123"}`},
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

func TestCreateUser_InvalidUserType(t *testing.T) {
	truncateTables(t)

	rec := doRequest(http.MethodPost, "/users", `{"name":"Jason Harris","email":"jason@example.com","password":"password123","user_type":"WIZARD"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	truncateTables(t)
	createTestUserViaAPI(t, "dup@example.com")

	rec := doRequest(http.MethodPost, "/users", `{"name":"Someone Else","email":"dup@example.com","password":"password123","user_type":"ARTIST"}`)
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

func TestGetUser(t *testing.T) {
	truncateTables(t)
	email := "jason@example.com"
	created := createTestUserViaAPI(t, email)
	token := loginTestUser(t, email, testUserPassword)

	rec := doAuthRequest(http.MethodGet, "/users/"+created.ID.String(), "", token)
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
	_, token := createAndLoginTestUser(t)

	rec := doAuthRequest(http.MethodGet, "/users/11111111-1111-1111-1111-111111111111", "", token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetUser_InvalidID(t *testing.T) {
	truncateTables(t)
	_, token := createAndLoginTestUser(t)

	rec := doAuthRequest(http.MethodGet, "/users/not-a-uuid", "", token)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetUser_RequiresAuth(t *testing.T) {
	truncateTables(t)
	created := createTestUserViaAPI(t, "jason@example.com")

	rec := doRequest(http.MethodGet, "/users/"+created.ID.String(), "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetUser_OtherUsersAccount404(t *testing.T) {
	truncateTables(t)
	_, tokenA := createAndLoginTestUser(t)
	userB, _ := createAndLoginTestUser(t)

	rec := doAuthRequest(http.MethodGet, "/users/"+userB.ID.String(), "", tokenA)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 (not 403, to avoid revealing existence), got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchUser_PartialUpdate(t *testing.T) {
	truncateTables(t)
	email := "jason@example.com"
	created := createTestUserViaAPI(t, email)
	token := loginTestUser(t, email, testUserPassword)

	rec := doAuthRequest(http.MethodPatch, "/users/"+created.ID.String(), `{"name":"J. Harris"}`, token)
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
	email := "jason@example.com"
	created := createTestUserViaAPI(t, email)
	token := loginTestUser(t, email, testUserPassword)

	rec := doAuthRequest(http.MethodPatch, "/users/"+created.ID.String(), `{"password":"newpassword456"}`, token)
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

func TestPatchUser_UserType(t *testing.T) {
	truncateTables(t)
	email := "jason@example.com"
	created := createTestUserViaAPI(t, email)
	token := loginTestUser(t, email, testUserPassword)

	rec := doAuthRequest(http.MethodPatch, "/users/"+created.ID.String(), `{"user_type":"LABEL"}`, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var u User
	json.Unmarshal(rec.Body.Bytes(), &u)
	if u.UserType != "LABEL" {
		t.Fatalf("expected user_type updated to LABEL, got %s", u.UserType)
	}
}

func TestPatchUser_InvalidUserType(t *testing.T) {
	truncateTables(t)
	email := "jason@example.com"
	created := createTestUserViaAPI(t, email)
	token := loginTestUser(t, email, testUserPassword)

	rec := doAuthRequest(http.MethodPatch, "/users/"+created.ID.String(), `{"user_type":"WIZARD"}`, token)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchUser_DuplicateEmail(t *testing.T) {
	truncateTables(t)
	createTestUserViaAPI(t, "taken@example.com")
	email := "jason@example.com"
	created := createTestUserViaAPI(t, email)
	token := loginTestUser(t, email, testUserPassword)

	rec := doAuthRequest(http.MethodPatch, "/users/"+created.ID.String(), `{"email":"taken@example.com"}`, token)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchUser_NotFound(t *testing.T) {
	truncateTables(t)
	_, token := createAndLoginTestUser(t)

	rec := doAuthRequest(http.MethodPatch, "/users/11111111-1111-1111-1111-111111111111", `{"name":"Nobody"}`, token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchUser_NoFields(t *testing.T) {
	truncateTables(t)
	email := "jason@example.com"
	created := createTestUserViaAPI(t, email)
	token := loginTestUser(t, email, testUserPassword)

	rec := doAuthRequest(http.MethodPatch, "/users/"+created.ID.String(), `{}`, token)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchUser_RequiresAuth(t *testing.T) {
	truncateTables(t)
	created := createTestUserViaAPI(t, "jason@example.com")

	rec := doRequest(http.MethodPatch, "/users/"+created.ID.String(), `{"name":"Nobody"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchUser_OtherUsersAccount404(t *testing.T) {
	truncateTables(t)
	_, tokenA := createAndLoginTestUser(t)
	userB, _ := createAndLoginTestUser(t)

	rec := doAuthRequest(http.MethodPatch, "/users/"+userB.ID.String(), `{"name":"Hacked"}`, tokenA)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}

	var name string
	err := testPool.QueryRow(context.Background(), "SELECT name FROM users WHERE id = $1", userB.ID).Scan(&name)
	if err != nil {
		t.Fatalf("failed to read user: %v", err)
	}
	if name == "Hacked" {
		t.Fatalf("expected user B's name to be untouched, got %q", name)
	}
}

func TestDeleteUser(t *testing.T) {
	truncateTables(t)
	email := "jason@example.com"
	created := createTestUserViaAPI(t, email)
	token := loginTestUser(t, email, testUserPassword)

	rec := doAuthRequest(http.MethodDelete, "/users/"+created.ID.String(), "", token)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	var count int
	err := testPool.QueryRow(context.Background(), "SELECT count(*) FROM users WHERE id = $1", created.ID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query users: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected user to be deleted, but it still exists")
	}
}

func TestDeleteUser_NotFound(t *testing.T) {
	truncateTables(t)
	_, token := createAndLoginTestUser(t)

	rec := doAuthRequest(http.MethodDelete, "/users/11111111-1111-1111-1111-111111111111", "", token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteUser_RequiresAuth(t *testing.T) {
	truncateTables(t)
	created := createTestUserViaAPI(t, "jason@example.com")

	rec := doRequest(http.MethodDelete, "/users/"+created.ID.String(), "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteUser_OtherUsersAccount404(t *testing.T) {
	truncateTables(t)
	_, tokenA := createAndLoginTestUser(t)
	userB, _ := createAndLoginTestUser(t)

	rec := doAuthRequest(http.MethodDelete, "/users/"+userB.ID.String(), "", tokenA)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}

	var count int
	err := testPool.QueryRow(context.Background(), "SELECT count(*) FROM users WHERE id = $1", userB.ID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query users: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected user B to survive user A's failed delete attempt")
	}
}

func TestDeleteUser_CascadesTourdates(t *testing.T) {
	truncateTables(t)
	email := "jason@example.com"
	created := createTestUserViaAPI(t, email)
	token := loginTestUser(t, email, testUserPassword)
	tourdate := createTestTourDate(t, token, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)

	rec := doAuthRequest(http.MethodDelete, "/users/"+created.ID.String(), "", token)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	// The user's own session was cascade-deleted along with the user, so the
	// old token can no longer authenticate - check the row is gone directly.
	var count int
	err := testPool.QueryRow(context.Background(), "SELECT count(*) FROM tourdates WHERE id = $1", tourdate.ID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query tourdates: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected tourdate to be cascade-deleted, but it still exists")
	}
}
