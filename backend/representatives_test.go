package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

// createTestUserOfType creates and logs in a user of the given user_type,
// returning the user and their session token.
func createTestUserOfType(t *testing.T, userType string) (User, string) {
	t.Helper()
	email := fmt.Sprintf("%s@example.com", uuid.NewString())
	rec := doRequest(http.MethodPost, "/users", fmt.Sprintf(
		`{"name":"Test User","email":%q,"password":%q,"user_type":%q}`, email, testUserPassword, userType))
	if rec.Code != http.StatusCreated {
		t.Fatalf("failed to create %s user: status %d, body %s", userType, rec.Code, rec.Body.String())
	}
	var u User
	if err := json.Unmarshal(rec.Body.Bytes(), &u); err != nil {
		t.Fatalf("failed to decode created user: %v", err)
	}
	token := loginTestUser(t, email, testUserPassword)
	return u, token
}

func addRepresentative(t *testing.T, artistToken string, representativeID uuid.UUID) RepresentedUser {
	t.Helper()
	rec := doAuthRequest(http.MethodPost, "/representatives", fmt.Sprintf(`{"representative_id":%q}`, representativeID), artistToken)
	if rec.Code != http.StatusCreated {
		t.Fatalf("failed to create representative: status %d, body %s", rec.Code, rec.Body.String())
	}
	var ru RepresentedUser
	if err := json.Unmarshal(rec.Body.Bytes(), &ru); err != nil {
		t.Fatalf("failed to decode representative: %v", err)
	}
	return ru
}

func TestCreateRepresentative_Success(t *testing.T) {
	truncateTables(t)
	_, artistToken := createTestUserOfType(t, "ARTIST")
	manager, _ := createTestUserOfType(t, "MANAGER")

	rec := doAuthRequest(http.MethodPost, "/representatives", fmt.Sprintf(`{"representative_id":%q}`, manager.ID), artistToken)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var ru RepresentedUser
	json.Unmarshal(rec.Body.Bytes(), &ru)
	if ru.UserID != manager.ID || ru.UserType != "MANAGER" {
		t.Fatalf("unexpected representative: %+v", ru)
	}
}

func TestCreateRepresentative_NonArtistForbidden(t *testing.T) {
	truncateTables(t)
	_, managerToken := createTestUserOfType(t, "MANAGER")
	otherManager, _ := createTestUserOfType(t, "MANAGER")

	rec := doAuthRequest(http.MethodPost, "/representatives", fmt.Sprintf(`{"representative_id":%q}`, otherManager.ID), managerToken)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateRepresentative_InvalidRepresentativeType(t *testing.T) {
	truncateTables(t)
	_, artistToken := createTestUserOfType(t, "ARTIST")
	otherArtist, _ := createTestUserOfType(t, "ARTIST")

	rec := doAuthRequest(http.MethodPost, "/representatives", fmt.Sprintf(`{"representative_id":%q}`, otherArtist.ID), artistToken)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateRepresentative_UnknownRepresentative(t *testing.T) {
	truncateTables(t)
	_, artistToken := createTestUserOfType(t, "ARTIST")

	rec := doAuthRequest(http.MethodPost, "/representatives", `{"representative_id":"11111111-1111-1111-1111-111111111111"}`, artistToken)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateRepresentative_Self(t *testing.T) {
	truncateTables(t)
	artist, artistToken := createTestUserOfType(t, "ARTIST")

	rec := doAuthRequest(http.MethodPost, "/representatives", fmt.Sprintf(`{"representative_id":%q}`, artist.ID), artistToken)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateRepresentative_Duplicate(t *testing.T) {
	truncateTables(t)
	_, artistToken := createTestUserOfType(t, "ARTIST")
	manager, _ := createTestUserOfType(t, "MANAGER")
	addRepresentative(t, artistToken, manager.ID)

	rec := doAuthRequest(http.MethodPost, "/representatives", fmt.Sprintf(`{"representative_id":%q}`, manager.ID), artistToken)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListMyRepresentatives(t *testing.T) {
	truncateTables(t)
	_, artistToken := createTestUserOfType(t, "ARTIST")
	manager, _ := createTestUserOfType(t, "MANAGER")
	agent, _ := createTestUserOfType(t, "AGENT")
	addRepresentative(t, artistToken, manager.ID)
	addRepresentative(t, artistToken, agent.ID)

	rec := doAuthRequest(http.MethodGet, "/representatives", "", artistToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var reps []RepresentedUser
	json.Unmarshal(rec.Body.Bytes(), &reps)
	if len(reps) != 2 {
		t.Fatalf("expected 2 representatives, got %d", len(reps))
	}
}

func TestListRepresentedArtists(t *testing.T) {
	truncateTables(t)
	artistA, artistTokenA := createTestUserOfType(t, "ARTIST")
	artistB, artistTokenB := createTestUserOfType(t, "ARTIST")
	manager, managerToken := createTestUserOfType(t, "MANAGER")
	addRepresentative(t, artistTokenA, manager.ID)
	addRepresentative(t, artistTokenB, manager.ID)

	rec := doAuthRequest(http.MethodGet, "/represented-artists", "", managerToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var artists []RepresentedUser
	json.Unmarshal(rec.Body.Bytes(), &artists)
	if len(artists) != 2 {
		t.Fatalf("expected 2 represented artists, got %d", len(artists))
	}
	ids := map[uuid.UUID]bool{artists[0].UserID: true, artists[1].UserID: true}
	if !ids[artistA.ID] || !ids[artistB.ID] {
		t.Fatalf("expected both artists in list, got %+v", artists)
	}
}

func TestDeleteRepresentative_ByArtist(t *testing.T) {
	truncateTables(t)
	_, artistToken := createTestUserOfType(t, "ARTIST")
	manager, _ := createTestUserOfType(t, "MANAGER")
	ru := addRepresentative(t, artistToken, manager.ID)

	rec := doAuthRequest(http.MethodDelete, "/representatives/"+ru.RelationshipID.String(), "", artistToken)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteRepresentative_ByRepresentative(t *testing.T) {
	truncateTables(t)
	_, artistToken := createTestUserOfType(t, "ARTIST")
	manager, managerToken := createTestUserOfType(t, "MANAGER")
	ru := addRepresentative(t, artistToken, manager.ID)

	rec := doAuthRequest(http.MethodDelete, "/representatives/"+ru.RelationshipID.String(), "", managerToken)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteRepresentative_UnrelatedUser404(t *testing.T) {
	truncateTables(t)
	_, artistToken := createTestUserOfType(t, "ARTIST")
	manager, _ := createTestUserOfType(t, "MANAGER")
	ru := addRepresentative(t, artistToken, manager.ID)
	_, strangerToken := createTestUserOfType(t, "AGENT")

	rec := doAuthRequest(http.MethodDelete, "/representatives/"+ru.RelationshipID.String(), "", strangerToken)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

// --- tourdates access extended to representatives ---

func TestManagerCanReadArtistTourdates(t *testing.T) {
	truncateTables(t)
	artist, artistToken := createTestUserOfType(t, "ARTIST")
	manager, managerToken := createTestUserOfType(t, "MANAGER")
	addRepresentative(t, artistToken, manager.ID)
	created := createTestTourDate(t, artistToken, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)

	rec := doAuthRequest(http.MethodGet, "/tourdates/"+created.ID.String(), "", managerToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var td TourDate
	json.Unmarshal(rec.Body.Bytes(), &td)
	if td.UserID != artist.ID {
		t.Fatalf("expected tourdate owned by artist %s, got %s", artist.ID, td.UserID)
	}
}

func TestManagerCanWriteArtistTourdates(t *testing.T) {
	truncateTables(t)
	_, artistToken := createTestUserOfType(t, "ARTIST")
	manager, managerToken := createTestUserOfType(t, "MANAGER")
	addRepresentative(t, artistToken, manager.ID)
	created := createTestTourDate(t, artistToken, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)

	rec := doAuthRequest(http.MethodPatch, "/tourdates/"+created.ID.String(), `{"venue":"ACL Live"}`, managerToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var td TourDate
	json.Unmarshal(rec.Body.Bytes(), &td)
	if td.Venue != "ACL Live" {
		t.Fatalf("expected venue updated, got %s", td.Venue)
	}

	rec = doAuthRequest(http.MethodDelete, "/tourdates/"+created.ID.String(), "", managerToken)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestManagerCanCreateTourdateForArtist(t *testing.T) {
	truncateTables(t)
	artist, artistToken := createTestUserOfType(t, "ARTIST")
	manager, managerToken := createTestUserOfType(t, "MANAGER")
	addRepresentative(t, artistToken, manager.ID)

	rec := doAuthRequest(http.MethodPost, "/tourdates",
		fmt.Sprintf(`{"date":"2026-09-15","city":"Austin","venue":"Moody Center","artist_id":%q}`, artist.ID), managerToken)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var td TourDate
	json.Unmarshal(rec.Body.Bytes(), &td)
	if td.UserID != artist.ID {
		t.Fatalf("expected tourdate owned by artist %s, got %s", artist.ID, td.UserID)
	}
}

func TestManagerCannotAccessUnrepresentedArtist(t *testing.T) {
	truncateTables(t)
	_, artistToken := createTestUserOfType(t, "ARTIST")
	_, managerToken := createTestUserOfType(t, "MANAGER") // not added as a representative
	created := createTestTourDate(t, artistToken, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)

	rec := doAuthRequest(http.MethodGet, "/tourdates/"+created.ID.String(), "", managerToken)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestManagerCannotCreateForUnrepresentedArtist(t *testing.T) {
	truncateTables(t)
	artist, _ := createTestUserOfType(t, "ARTIST")
	_, managerToken := createTestUserOfType(t, "MANAGER")

	rec := doAuthRequest(http.MethodPost, "/tourdates",
		fmt.Sprintf(`{"date":"2026-09-15","city":"Austin","venue":"Moody Center","artist_id":%q}`, artist.ID), managerToken)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListTourDates_MergedAcrossRepresentedArtists(t *testing.T) {
	truncateTables(t)
	_, artistTokenA := createTestUserOfType(t, "ARTIST")
	_, artistTokenB := createTestUserOfType(t, "ARTIST")
	managerUser, managerToken := createTestUserOfType(t, "MANAGER")
	addRepresentative(t, artistTokenA, managerUser.ID)
	addRepresentative(t, artistTokenB, managerUser.ID)
	createTestTourDate(t, artistTokenA, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)
	createTestTourDate(t, artistTokenB, `{"date":"2026-09-20","city":"Dallas","venue":"American Airlines Center"}`)
	createTestTourDate(t, managerToken, `{"date":"2026-09-10","city":"Houston","venue":"Toyota Center"}`)

	rec := doAuthRequest(http.MethodGet, "/tourdates", "", managerToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var tourdates []TourDate
	json.Unmarshal(rec.Body.Bytes(), &tourdates)
	if len(tourdates) != 3 {
		t.Fatalf("expected 3 merged tourdates (2 artists + manager's own), got %d: %+v", len(tourdates), tourdates)
	}
}

func TestListTourDates_FilteredByArtistID(t *testing.T) {
	truncateTables(t)
	artistA, artistTokenA := createTestUserOfType(t, "ARTIST")
	_, artistTokenB := createTestUserOfType(t, "ARTIST")
	managerUser, managerToken := createTestUserOfType(t, "MANAGER")
	addRepresentative(t, artistTokenA, managerUser.ID)
	addRepresentative(t, artistTokenB, managerUser.ID)
	createTestTourDate(t, artistTokenA, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)
	createTestTourDate(t, artistTokenB, `{"date":"2026-09-20","city":"Dallas","venue":"American Airlines Center"}`)

	rec := doAuthRequest(http.MethodGet, "/tourdates?artist_id="+artistA.ID.String(), "", managerToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var tourdates []TourDate
	json.Unmarshal(rec.Body.Bytes(), &tourdates)
	if len(tourdates) != 1 || tourdates[0].City != "Austin" {
		t.Fatalf("expected only artist A's tourdate, got %+v", tourdates)
	}
}

func TestListTourDates_FilteredByUnrepresentedArtistID404(t *testing.T) {
	truncateTables(t)
	strangerArtist, _ := createTestUserOfType(t, "ARTIST")
	_, managerToken := createTestUserOfType(t, "MANAGER")

	rec := doAuthRequest(http.MethodGet, "/tourdates?artist_id="+strangerArtist.ID.String(), "", managerToken)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}
