package merch_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"skejio/backend/internal/merch"
	"skejio/backend/internal/testutil"
)

func TestMain(m *testing.M) { testutil.Run(m) }

func TestCreateMerch(t *testing.T) {
	testutil.TruncateTables(t)
	user, token := testutil.CreateAndLoginTestUser(t)

	rec := testutil.DoAuthRequest(http.MethodPost, "/merch",
		`{"name":"Tour Tee","type":"APPAREL","description":"Black tee","manufacturing_cost":5.50,"selling_price":25}`, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var m merch.Merch
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if m.Name != "Tour Tee" || m.Type != "APPAREL" || m.ManufacturingCost != 5.50 {
		t.Fatalf("unexpected merch: %+v", m)
	}
	if m.SellingPrice == nil || *m.SellingPrice != 25 {
		t.Fatalf("expected selling_price to be set, got %v", m.SellingPrice)
	}
	if m.Description == nil || *m.Description != "Black tee" {
		t.Fatalf("expected description to be set, got %+v", m.Description)
	}
	if m.UserID != user.ID {
		t.Fatalf("expected user_id %s, got %s", user.ID, m.UserID)
	}
}

func TestCreateMerch_NullableFields(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)

	rec := testutil.DoAuthRequest(http.MethodPost, "/merch",
		`{"name":"Vinyl","type":"MUSIC","manufacturing_cost":3}`, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var m merch.Merch
	json.Unmarshal(rec.Body.Bytes(), &m)
	if m.Description != nil {
		t.Fatalf("expected nil description, got %v", *m.Description)
	}
	if m.PhotoURL != nil {
		t.Fatalf("expected nil photo_url, got %v", *m.PhotoURL)
	}
	if m.SellingPrice != nil {
		t.Fatalf("expected nil selling_price, got %v", *m.SellingPrice)
	}
}

func TestCreateMerch_InvalidType(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)

	rec := testutil.DoAuthRequest(http.MethodPost, "/merch",
		`{"name":"Tour Tee","type":"CLOTHING","manufacturing_cost":5,"selling_price":25}`, token)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateMerch_MissingFields(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)

	cases := []struct {
		name string
		body string
	}{
		{"missing name", `{"type":"APPAREL","manufacturing_cost":5,"selling_price":25}`},
		{"missing manufacturing_cost", `{"name":"Tour Tee","type":"APPAREL","selling_price":25}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec := testutil.DoAuthRequest(http.MethodPost, "/merch", c.body, token)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestCreateMerch_RequiresAuth(t *testing.T) {
	testutil.TruncateTables(t)

	rec := testutil.DoRequest(http.MethodPost, "/merch",
		`{"name":"Tour Tee","type":"APPAREL","manufacturing_cost":5,"selling_price":25}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListMerch_ScopedToOwner(t *testing.T) {
	testutil.TruncateTables(t)
	_, tokenA := testutil.CreateAndLoginTestUser(t)
	_, tokenB := testutil.CreateAndLoginTestUser(t)

	testutil.CreateTestMerch(t, tokenA, `{"name":"A's Tee","type":"APPAREL","manufacturing_cost":5,"selling_price":25}`)
	testutil.CreateTestMerch(t, tokenB, `{"name":"B's Tee","type":"APPAREL","manufacturing_cost":5,"selling_price":25}`)

	rec := testutil.DoAuthRequest(http.MethodGet, "/merch", "", tokenA)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var list []merch.Merch
	json.Unmarshal(rec.Body.Bytes(), &list)
	if len(list) != 1 || list[0].Name != "A's Tee" {
		t.Fatalf("expected only user A's merch, got %+v", list)
	}
}

func TestGetMerch_OtherUsersMerch404(t *testing.T) {
	testutil.TruncateTables(t)
	_, tokenA := testutil.CreateAndLoginTestUser(t)
	_, tokenB := testutil.CreateAndLoginTestUser(t)
	created := testutil.CreateTestMerch(t, tokenA, `{"name":"A's Tee","type":"APPAREL","manufacturing_cost":5,"selling_price":25}`)

	rec := testutil.DoAuthRequest(http.MethodGet, "/merch/"+created.ID.String(), "", tokenB)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 (not 403, to avoid revealing existence), got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetMerch_NotFound(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)

	rec := testutil.DoAuthRequest(http.MethodGet, "/merch/11111111-1111-1111-1111-111111111111", "", token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchMerch_PartialUpdate(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	created := testutil.CreateTestMerch(t, token, `{"name":"Tour Tee","type":"APPAREL","manufacturing_cost":5,"selling_price":25}`)

	rec := testutil.DoAuthRequest(http.MethodPatch, "/merch/"+created.ID.String(), `{"selling_price":30}`, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var m merch.Merch
	json.Unmarshal(rec.Body.Bytes(), &m)
	if m.SellingPrice == nil || *m.SellingPrice != 30 {
		t.Fatalf("expected selling_price to be updated, got %v", m.SellingPrice)
	}
	if m.Name != "Tour Tee" || m.ManufacturingCost != 5 {
		t.Fatalf("expected other fields untouched, got %+v", m)
	}
}

func TestPatchMerch_ClearDescriptionToNull(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	created := testutil.CreateTestMerch(t, token,
		`{"name":"Tour Tee","type":"APPAREL","description":"Black tee","manufacturing_cost":5,"selling_price":25}`)

	rec := testutil.DoAuthRequest(http.MethodPatch, "/merch/"+created.ID.String(), `{"description":null}`, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var m merch.Merch
	json.Unmarshal(rec.Body.Bytes(), &m)
	if m.Description != nil {
		t.Fatalf("expected description to be cleared, got %v", *m.Description)
	}
}

func TestPatchMerch_CannotClearRequiredFields(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	created := testutil.CreateTestMerch(t, token, `{"name":"Tour Tee","type":"APPAREL","manufacturing_cost":5,"selling_price":25}`)

	cases := []struct {
		name string
		body string
	}{
		{"name", `{"name":null}`},
		{"manufacturing_cost", `{"manufacturing_cost":null}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec := testutil.DoAuthRequest(http.MethodPatch, "/merch/"+created.ID.String(), c.body, token)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestPatchMerch_ClearSellingPriceToNull(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	created := testutil.CreateTestMerch(t, token, `{"name":"Tour Tee","type":"APPAREL","manufacturing_cost":5,"selling_price":25}`)

	rec := testutil.DoAuthRequest(http.MethodPatch, "/merch/"+created.ID.String(), `{"selling_price":null}`, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var m merch.Merch
	json.Unmarshal(rec.Body.Bytes(), &m)
	if m.SellingPrice != nil {
		t.Fatalf("expected selling_price to be cleared, got %v", *m.SellingPrice)
	}
}

func TestPatchMerch_OtherUsersMerch404(t *testing.T) {
	testutil.TruncateTables(t)
	_, tokenA := testutil.CreateAndLoginTestUser(t)
	_, tokenB := testutil.CreateAndLoginTestUser(t)
	created := testutil.CreateTestMerch(t, tokenA, `{"name":"A's Tee","type":"APPAREL","manufacturing_cost":5,"selling_price":25}`)

	rec := testutil.DoAuthRequest(http.MethodPatch, "/merch/"+created.ID.String(), `{"name":"Hacked"}`, tokenB)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteMerch(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	created := testutil.CreateTestMerch(t, token, `{"name":"Tour Tee","type":"APPAREL","manufacturing_cost":5,"selling_price":25}`)

	rec := testutil.DoAuthRequest(http.MethodDelete, "/merch/"+created.ID.String(), "", token)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testutil.DoAuthRequest(http.MethodGet, "/merch/"+created.ID.String(), "", token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected merch to be gone, got %d", rec.Code)
	}
}

func TestDeleteMerch_OtherUsersMerch404(t *testing.T) {
	testutil.TruncateTables(t)
	_, tokenA := testutil.CreateAndLoginTestUser(t)
	_, tokenB := testutil.CreateAndLoginTestUser(t)
	created := testutil.CreateTestMerch(t, tokenA, `{"name":"A's Tee","type":"APPAREL","manufacturing_cost":5,"selling_price":25}`)

	rec := testutil.DoAuthRequest(http.MethodDelete, "/merch/"+created.ID.String(), "", tokenB)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testutil.DoAuthRequest(http.MethodGet, "/merch/"+created.ID.String(), "", tokenA)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected merch to survive the other user's failed delete, got %d", rec.Code)
	}
}

func TestDeleteMerch_CascadesVariants(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	created := testutil.CreateTestMerch(t, token, `{"name":"Tour Tee","type":"APPAREL","manufacturing_cost":5,"selling_price":25}`)
	variant := testutil.CreateTestMerchVariant(t, token, created.ID.String(), `{"label":"Medium","inventory_count":10}`)

	rec := testutil.DoAuthRequest(http.MethodDelete, "/merch/"+created.ID.String(), "", token)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testutil.DoAuthRequest(http.MethodGet, "/merch/"+created.ID.String()+"/variants/"+variant.ID.String(), "", token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected variant to be gone after merch delete, got %d", rec.Code)
	}
}

func TestMerch_RepresentativeHasFullAccess(t *testing.T) {
	testutil.TruncateTables(t)
	_, artistToken := testutil.CreateTestUserOfType(t, "ARTIST")
	manager, managerToken := testutil.CreateTestUserOfType(t, "MANAGER")
	testutil.AddRepresentative(t, artistToken, manager.ID)

	rec := testutil.DoAuthRequest(http.MethodPost, "/merch",
		`{"name":"Tour Tee","type":"APPAREL","manufacturing_cost":5,"selling_price":25}`, managerToken)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected manager to create merch for their artist, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestMerch_UnrepresentedArtist404(t *testing.T) {
	testutil.TruncateTables(t)
	_, artistToken := testutil.CreateTestUserOfType(t, "ARTIST")
	_, managerToken := testutil.CreateTestUserOfType(t, "MANAGER")
	created := testutil.CreateTestMerch(t, artistToken, `{"name":"A's Tee","type":"APPAREL","manufacturing_cost":5,"selling_price":25}`)

	rec := testutil.DoAuthRequest(http.MethodGet, "/merch/"+created.ID.String(), "", managerToken)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for an artist the manager doesn't represent, got %d: %s", rec.Code, rec.Body.String())
	}
}
