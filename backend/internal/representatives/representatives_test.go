package representatives_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"skejio/backend/internal/representatives"
	"skejio/backend/internal/testutil"
	"skejio/backend/internal/tourdates"
)

func TestMain(m *testing.M) { testutil.Run(m) }

func TestCreateRepresentative_Success(t *testing.T) {
	testutil.TruncateTables(t)
	_, artistToken := testutil.CreateTestUserOfType(t, "ARTIST")
	manager, _ := testutil.CreateTestUserOfType(t, "MANAGER")

	rec := testutil.DoAuthRequest(http.MethodPost, "/representatives", fmt.Sprintf(`{"representative_id":%q}`, manager.ID), artistToken)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var ru representatives.RepresentedUser
	json.Unmarshal(rec.Body.Bytes(), &ru)
	if ru.UserID != manager.ID || ru.UserType != "MANAGER" {
		t.Fatalf("unexpected representative: %+v", ru)
	}
}

func TestCreateRepresentative_NonArtistForbidden(t *testing.T) {
	testutil.TruncateTables(t)
	_, managerToken := testutil.CreateTestUserOfType(t, "MANAGER")
	otherManager, _ := testutil.CreateTestUserOfType(t, "MANAGER")

	rec := testutil.DoAuthRequest(http.MethodPost, "/representatives", fmt.Sprintf(`{"representative_id":%q}`, otherManager.ID), managerToken)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateRepresentative_InvalidRepresentativeType(t *testing.T) {
	testutil.TruncateTables(t)
	_, artistToken := testutil.CreateTestUserOfType(t, "ARTIST")
	otherArtist, _ := testutil.CreateTestUserOfType(t, "ARTIST")

	rec := testutil.DoAuthRequest(http.MethodPost, "/representatives", fmt.Sprintf(`{"representative_id":%q}`, otherArtist.ID), artistToken)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateRepresentative_UnknownRepresentative(t *testing.T) {
	testutil.TruncateTables(t)
	_, artistToken := testutil.CreateTestUserOfType(t, "ARTIST")

	rec := testutil.DoAuthRequest(http.MethodPost, "/representatives", `{"representative_id":"11111111-1111-1111-1111-111111111111"}`, artistToken)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateRepresentative_Self(t *testing.T) {
	testutil.TruncateTables(t)
	artist, artistToken := testutil.CreateTestUserOfType(t, "ARTIST")

	rec := testutil.DoAuthRequest(http.MethodPost, "/representatives", fmt.Sprintf(`{"representative_id":%q}`, artist.ID), artistToken)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateRepresentative_Duplicate(t *testing.T) {
	testutil.TruncateTables(t)
	_, artistToken := testutil.CreateTestUserOfType(t, "ARTIST")
	manager, _ := testutil.CreateTestUserOfType(t, "MANAGER")
	testutil.AddRepresentative(t, artistToken, manager.ID)

	rec := testutil.DoAuthRequest(http.MethodPost, "/representatives", fmt.Sprintf(`{"representative_id":%q}`, manager.ID), artistToken)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListMyRepresentatives(t *testing.T) {
	testutil.TruncateTables(t)
	_, artistToken := testutil.CreateTestUserOfType(t, "ARTIST")
	manager, _ := testutil.CreateTestUserOfType(t, "MANAGER")
	agent, _ := testutil.CreateTestUserOfType(t, "AGENT")
	testutil.AddRepresentative(t, artistToken, manager.ID)
	testutil.AddRepresentative(t, artistToken, agent.ID)

	rec := testutil.DoAuthRequest(http.MethodGet, "/representatives", "", artistToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var reps []representatives.RepresentedUser
	json.Unmarshal(rec.Body.Bytes(), &reps)
	if len(reps) != 2 {
		t.Fatalf("expected 2 representatives, got %d", len(reps))
	}
}

func TestListRepresentedArtists(t *testing.T) {
	testutil.TruncateTables(t)
	artistA, artistTokenA := testutil.CreateTestUserOfType(t, "ARTIST")
	artistB, artistTokenB := testutil.CreateTestUserOfType(t, "ARTIST")
	manager, managerToken := testutil.CreateTestUserOfType(t, "MANAGER")
	testutil.AddRepresentative(t, artistTokenA, manager.ID)
	testutil.AddRepresentative(t, artistTokenB, manager.ID)

	rec := testutil.DoAuthRequest(http.MethodGet, "/represented-artists", "", managerToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var artists []representatives.RepresentedUser
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
	testutil.TruncateTables(t)
	_, artistToken := testutil.CreateTestUserOfType(t, "ARTIST")
	manager, _ := testutil.CreateTestUserOfType(t, "MANAGER")
	ru := testutil.AddRepresentative(t, artistToken, manager.ID)

	rec := testutil.DoAuthRequest(http.MethodDelete, "/representatives/"+ru.RelationshipID.String(), "", artistToken)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteRepresentative_ByRepresentative(t *testing.T) {
	testutil.TruncateTables(t)
	_, artistToken := testutil.CreateTestUserOfType(t, "ARTIST")
	manager, managerToken := testutil.CreateTestUserOfType(t, "MANAGER")
	ru := testutil.AddRepresentative(t, artistToken, manager.ID)

	rec := testutil.DoAuthRequest(http.MethodDelete, "/representatives/"+ru.RelationshipID.String(), "", managerToken)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteRepresentative_UnrelatedUser404(t *testing.T) {
	testutil.TruncateTables(t)
	_, artistToken := testutil.CreateTestUserOfType(t, "ARTIST")
	manager, _ := testutil.CreateTestUserOfType(t, "MANAGER")
	ru := testutil.AddRepresentative(t, artistToken, manager.ID)
	_, strangerToken := testutil.CreateTestUserOfType(t, "AGENT")

	rec := testutil.DoAuthRequest(http.MethodDelete, "/representatives/"+ru.RelationshipID.String(), "", strangerToken)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

// --- tourdates access extended to representatives ---

func TestManagerCanReadArtistTourdates(t *testing.T) {
	testutil.TruncateTables(t)
	artist, artistToken := testutil.CreateTestUserOfType(t, "ARTIST")
	manager, managerToken := testutil.CreateTestUserOfType(t, "MANAGER")
	testutil.AddRepresentative(t, artistToken, manager.ID)
	created := testutil.CreateTestTourDate(t, artistToken, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)

	rec := testutil.DoAuthRequest(http.MethodGet, "/tourdates/"+created.ID.String(), "", managerToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var td tourdates.TourDate
	json.Unmarshal(rec.Body.Bytes(), &td)
	if td.UserID != artist.ID {
		t.Fatalf("expected tourdate owned by artist %s, got %s", artist.ID, td.UserID)
	}
}

func TestManagerCanWriteArtistTourdates(t *testing.T) {
	testutil.TruncateTables(t)
	_, artistToken := testutil.CreateTestUserOfType(t, "ARTIST")
	manager, managerToken := testutil.CreateTestUserOfType(t, "MANAGER")
	testutil.AddRepresentative(t, artistToken, manager.ID)
	created := testutil.CreateTestTourDate(t, artistToken, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)

	rec := testutil.DoAuthRequest(http.MethodPatch, "/tourdates/"+created.ID.String(), `{"venue":"ACL Live"}`, managerToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var td tourdates.TourDate
	json.Unmarshal(rec.Body.Bytes(), &td)
	if td.Venue != "ACL Live" {
		t.Fatalf("expected venue updated, got %s", td.Venue)
	}

	rec = testutil.DoAuthRequest(http.MethodDelete, "/tourdates/"+created.ID.String(), "", managerToken)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestManagerCanCreateTourdateForArtist(t *testing.T) {
	testutil.TruncateTables(t)
	artist, artistToken := testutil.CreateTestUserOfType(t, "ARTIST")
	manager, managerToken := testutil.CreateTestUserOfType(t, "MANAGER")
	testutil.AddRepresentative(t, artistToken, manager.ID)

	rec := testutil.DoAuthRequest(http.MethodPost, "/tourdates",
		fmt.Sprintf(`{"date":"2026-09-15","city":"Austin","venue":"Moody Center","artist_id":%q}`, artist.ID), managerToken)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var td tourdates.TourDate
	json.Unmarshal(rec.Body.Bytes(), &td)
	if td.UserID != artist.ID {
		t.Fatalf("expected tourdate owned by artist %s, got %s", artist.ID, td.UserID)
	}
}

func TestManagerCannotAccessUnrepresentedArtist(t *testing.T) {
	testutil.TruncateTables(t)
	_, artistToken := testutil.CreateTestUserOfType(t, "ARTIST")
	_, managerToken := testutil.CreateTestUserOfType(t, "MANAGER") // not added as a representative
	created := testutil.CreateTestTourDate(t, artistToken, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)

	rec := testutil.DoAuthRequest(http.MethodGet, "/tourdates/"+created.ID.String(), "", managerToken)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestManagerCannotCreateForUnrepresentedArtist(t *testing.T) {
	testutil.TruncateTables(t)
	artist, _ := testutil.CreateTestUserOfType(t, "ARTIST")
	_, managerToken := testutil.CreateTestUserOfType(t, "MANAGER")

	rec := testutil.DoAuthRequest(http.MethodPost, "/tourdates",
		fmt.Sprintf(`{"date":"2026-09-15","city":"Austin","venue":"Moody Center","artist_id":%q}`, artist.ID), managerToken)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListTourDates_MergedAcrossRepresentedArtists(t *testing.T) {
	testutil.TruncateTables(t)
	_, artistTokenA := testutil.CreateTestUserOfType(t, "ARTIST")
	_, artistTokenB := testutil.CreateTestUserOfType(t, "ARTIST")
	managerUser, managerToken := testutil.CreateTestUserOfType(t, "MANAGER")
	testutil.AddRepresentative(t, artistTokenA, managerUser.ID)
	testutil.AddRepresentative(t, artistTokenB, managerUser.ID)
	testutil.CreateTestTourDate(t, artistTokenA, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)
	testutil.CreateTestTourDate(t, artistTokenB, `{"date":"2026-09-20","city":"Dallas","venue":"American Airlines Center"}`)
	testutil.CreateTestTourDate(t, managerToken, `{"date":"2026-09-10","city":"Houston","venue":"Toyota Center"}`)

	rec := testutil.DoAuthRequest(http.MethodGet, "/tourdates", "", managerToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var tds []tourdates.TourDate
	json.Unmarshal(rec.Body.Bytes(), &tds)
	if len(tds) != 3 {
		t.Fatalf("expected 3 merged tourdates (2 artists + manager's own), got %d: %+v", len(tds), tds)
	}
}

func TestListTourDates_FilteredByArtistID(t *testing.T) {
	testutil.TruncateTables(t)
	artistA, artistTokenA := testutil.CreateTestUserOfType(t, "ARTIST")
	_, artistTokenB := testutil.CreateTestUserOfType(t, "ARTIST")
	managerUser, managerToken := testutil.CreateTestUserOfType(t, "MANAGER")
	testutil.AddRepresentative(t, artistTokenA, managerUser.ID)
	testutil.AddRepresentative(t, artistTokenB, managerUser.ID)
	testutil.CreateTestTourDate(t, artistTokenA, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)
	testutil.CreateTestTourDate(t, artistTokenB, `{"date":"2026-09-20","city":"Dallas","venue":"American Airlines Center"}`)

	rec := testutil.DoAuthRequest(http.MethodGet, "/tourdates?artist_id="+artistA.ID.String(), "", managerToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var tds []tourdates.TourDate
	json.Unmarshal(rec.Body.Bytes(), &tds)
	if len(tds) != 1 || tds[0].City != "Austin" {
		t.Fatalf("expected only artist A's tourdate, got %+v", tds)
	}
}

func TestListTourDates_FilteredByUnrepresentedArtistID404(t *testing.T) {
	testutil.TruncateTables(t)
	strangerArtist, _ := testutil.CreateTestUserOfType(t, "ARTIST")
	_, managerToken := testutil.CreateTestUserOfType(t, "MANAGER")

	rec := testutil.DoAuthRequest(http.MethodGet, "/tourdates?artist_id="+strangerArtist.ID.String(), "", managerToken)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}
