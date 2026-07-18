package merchvariants_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"skejio/backend/internal/merchvariants"
	"skejio/backend/internal/testutil"
)

func TestMain(m *testing.M) { testutil.Run(m) }

func createMerch(t *testing.T, token string) string {
	t.Helper()
	m := testutil.CreateTestMerch(t, token, `{"name":"Tour Tee","type":"APPAREL","manufacturing_cost":5,"selling_price":25}`)
	return m.ID.String()
}

func TestCreateVariant(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	merchID := createMerch(t, token)

	rec := testutil.DoAuthRequest(http.MethodPost, "/merch/"+merchID+"/variants", `{"label":"Medium","inventory_count":10}`, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var v merchvariants.Variant
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if v.MerchID.String() != merchID {
		t.Fatalf("expected merch_id %s, got %s", merchID, v.MerchID)
	}
	if v.Label != "Medium" || v.InventoryCount != 10 {
		t.Fatalf("unexpected variant: %+v", v)
	}
}

func TestCreateVariant_DefaultsInventoryToZero(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	merchID := createMerch(t, token)

	rec := testutil.DoAuthRequest(http.MethodPost, "/merch/"+merchID+"/variants", `{"label":"Red"}`, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var v merchvariants.Variant
	json.Unmarshal(rec.Body.Bytes(), &v)
	if v.InventoryCount != 0 {
		t.Fatalf("expected inventory_count to default to 0, got %d", v.InventoryCount)
	}
}

func TestCreateVariant_MissingLabel(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	merchID := createMerch(t, token)

	rec := testutil.DoAuthRequest(http.MethodPost, "/merch/"+merchID+"/variants", `{"inventory_count":5}`, token)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateVariant_NegativeInventoryRejected(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	merchID := createMerch(t, token)

	rec := testutil.DoAuthRequest(http.MethodPost, "/merch/"+merchID+"/variants", `{"label":"Medium","inventory_count":-1}`, token)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateVariant_DuplicateLabelRejected(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	merchID := createMerch(t, token)
	testutil.CreateTestMerchVariant(t, token, merchID, `{"label":"Medium","inventory_count":10}`)

	rec := testutil.DoAuthRequest(http.MethodPost, "/merch/"+merchID+"/variants", `{"label":"Medium","inventory_count":5}`, token)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateVariant_MerchNotFound(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)

	rec := testutil.DoAuthRequest(http.MethodPost, "/merch/11111111-1111-1111-1111-111111111111/variants", `{"label":"Medium"}`, token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateVariant_OtherUsersMerch404(t *testing.T) {
	testutil.TruncateTables(t)
	_, tokenA := testutil.CreateAndLoginTestUser(t)
	_, tokenB := testutil.CreateAndLoginTestUser(t)
	merchID := createMerch(t, tokenA)

	rec := testutil.DoAuthRequest(http.MethodPost, "/merch/"+merchID+"/variants", `{"label":"Medium"}`, tokenB)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 (not 403, to avoid revealing existence), got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateVariant_RequiresAuth(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	merchID := createMerch(t, token)

	rec := testutil.DoRequest(http.MethodPost, "/merch/"+merchID+"/variants", `{"label":"Medium"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListVariants_OrderedByCreatedAt(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	merchID := createMerch(t, token)

	testutil.CreateTestMerchVariant(t, token, merchID, `{"label":"Small"}`)
	testutil.CreateTestMerchVariant(t, token, merchID, `{"label":"Medium"}`)

	rec := testutil.DoAuthRequest(http.MethodGet, "/merch/"+merchID+"/variants", "", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var list []merchvariants.Variant
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 variants, got %d", len(list))
	}
	if list[0].Label != "Small" || list[1].Label != "Medium" {
		t.Fatalf("expected results ordered by created_at, got %s then %s", list[0].Label, list[1].Label)
	}
}

func TestListVariants_ScopedToMerch(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	merchA := createMerch(t, token)
	merchB := createMerch(t, token)

	testutil.CreateTestMerchVariant(t, token, merchA, `{"label":"Small"}`)
	testutil.CreateTestMerchVariant(t, token, merchB, `{"label":"Red"}`)

	rec := testutil.DoAuthRequest(http.MethodGet, "/merch/"+merchA+"/variants", "", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var list []merchvariants.Variant
	json.Unmarshal(rec.Body.Bytes(), &list)
	if len(list) != 1 || list[0].Label != "Small" {
		t.Fatalf("expected only merch A's variant, got %+v", list)
	}
}

func TestGetVariant(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	merchID := createMerch(t, token)
	created := testutil.CreateTestMerchVariant(t, token, merchID, `{"label":"Medium","inventory_count":10}`)

	rec := testutil.DoAuthRequest(http.MethodGet, "/merch/"+merchID+"/variants/"+created.ID.String(), "", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var v merchvariants.Variant
	json.Unmarshal(rec.Body.Bytes(), &v)
	if v.ID != created.ID {
		t.Fatalf("expected id %s, got %s", created.ID, v.ID)
	}
}

func TestGetVariant_NotFound(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	merchID := createMerch(t, token)

	rec := testutil.DoAuthRequest(http.MethodGet, "/merch/"+merchID+"/variants/11111111-1111-1111-1111-111111111111", "", token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetVariant_WrongMerch404(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	merchA := createMerch(t, token)
	merchB := createMerch(t, token)
	created := testutil.CreateTestMerchVariant(t, token, merchA, `{"label":"Medium"}`)

	rec := testutil.DoAuthRequest(http.MethodGet, "/merch/"+merchB+"/variants/"+created.ID.String(), "", token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for a variant that belongs to a different merch item, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchVariant_PartialUpdate(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	merchID := createMerch(t, token)
	created := testutil.CreateTestMerchVariant(t, token, merchID, `{"label":"Medium","inventory_count":10}`)

	rec := testutil.DoAuthRequest(http.MethodPatch, "/merch/"+merchID+"/variants/"+created.ID.String(), `{"inventory_count":25}`, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var v merchvariants.Variant
	json.Unmarshal(rec.Body.Bytes(), &v)
	if v.InventoryCount != 25 {
		t.Fatalf("expected inventory_count to be updated to 25, got %d", v.InventoryCount)
	}
	if v.Label != "Medium" {
		t.Fatalf("expected label untouched, got %s", v.Label)
	}
}

func TestPatchVariant_NullLabelRejected(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	merchID := createMerch(t, token)
	created := testutil.CreateTestMerchVariant(t, token, merchID, `{"label":"Medium"}`)

	rec := testutil.DoAuthRequest(http.MethodPatch, "/merch/"+merchID+"/variants/"+created.ID.String(), `{"label":null}`, token)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchVariant_NegativeInventoryRejected(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	merchID := createMerch(t, token)
	created := testutil.CreateTestMerchVariant(t, token, merchID, `{"label":"Medium","inventory_count":10}`)

	rec := testutil.DoAuthRequest(http.MethodPatch, "/merch/"+merchID+"/variants/"+created.ID.String(), `{"inventory_count":-5}`, token)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchVariant_DuplicateLabelRejected(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	merchID := createMerch(t, token)
	testutil.CreateTestMerchVariant(t, token, merchID, `{"label":"Small"}`)
	created := testutil.CreateTestMerchVariant(t, token, merchID, `{"label":"Medium"}`)

	rec := testutil.DoAuthRequest(http.MethodPatch, "/merch/"+merchID+"/variants/"+created.ID.String(), `{"label":"Small"}`, token)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchVariant_NotFound(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	merchID := createMerch(t, token)

	rec := testutil.DoAuthRequest(http.MethodPatch, "/merch/"+merchID+"/variants/11111111-1111-1111-1111-111111111111", `{"inventory_count":5}`, token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchVariant_NoFields(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	merchID := createMerch(t, token)
	created := testutil.CreateTestMerchVariant(t, token, merchID, `{"label":"Medium"}`)

	rec := testutil.DoAuthRequest(http.MethodPatch, "/merch/"+merchID+"/variants/"+created.ID.String(), `{}`, token)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteVariant(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	merchID := createMerch(t, token)
	created := testutil.CreateTestMerchVariant(t, token, merchID, `{"label":"Medium"}`)

	rec := testutil.DoAuthRequest(http.MethodDelete, "/merch/"+merchID+"/variants/"+created.ID.String(), "", token)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testutil.DoAuthRequest(http.MethodGet, "/merch/"+merchID+"/variants/"+created.ID.String(), "", token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected variant to be gone, got %d", rec.Code)
	}
}

func TestDeleteVariant_NotFound(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	merchID := createMerch(t, token)

	rec := testutil.DoAuthRequest(http.MethodDelete, "/merch/"+merchID+"/variants/11111111-1111-1111-1111-111111111111", "", token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestVariants_RepresentativeHasFullAccess(t *testing.T) {
	testutil.TruncateTables(t)
	_, artistToken := testutil.CreateTestUserOfType(t, "ARTIST")
	manager, managerToken := testutil.CreateTestUserOfType(t, "MANAGER")
	testutil.AddRepresentative(t, artistToken, manager.ID)
	merchID := createMerch(t, artistToken)

	rec := testutil.DoAuthRequest(http.MethodPost, "/merch/"+merchID+"/variants", `{"label":"Medium"}`, managerToken)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected manager to create a variant for their artist's merch, got %d: %s", rec.Code, rec.Body.String())
	}
}
