package financials_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"skejio/backend/internal/financials"
	"skejio/backend/internal/testutil"
)

func TestMain(m *testing.M) { testutil.Run(m) }

func createTourDate(t *testing.T, token string) string {
	t.Helper()
	td := testutil.CreateTestTourDate(t, token, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)
	return td.ID.String()
}

func TestCreateFinancials(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, token)

	rec := testutil.DoAuthRequest(http.MethodPost, "/tourdates/"+tourDateID+"/financials", `{"fee":1500,"tips":200.50}`, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var f financials.Financials
	if err := json.Unmarshal(rec.Body.Bytes(), &f); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if f.TourDateID.String() != tourDateID {
		t.Fatalf("expected tourdate_id %s, got %s", tourDateID, f.TourDateID)
	}
	if f.Fee == nil || *f.Fee != 1500 {
		t.Fatalf("expected fee 1500, got %+v", f.Fee)
	}
	if f.Tips == nil || *f.Tips != 200.50 {
		t.Fatalf("expected tips 200.50, got %+v", f.Tips)
	}
}

func TestCreateFinancials_NullableFields(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, token)

	rec := testutil.DoAuthRequest(http.MethodPost, "/tourdates/"+tourDateID+"/financials", `{}`, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var f financials.Financials
	json.Unmarshal(rec.Body.Bytes(), &f)
	if f.Fee != nil || f.Tips != nil {
		t.Fatalf("expected nil fee/tips, got %+v", f)
	}
}

func TestCreateFinancials_Duplicate(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, token)

	testutil.DoAuthRequest(http.MethodPost, "/tourdates/"+tourDateID+"/financials", `{"fee":100}`, token)
	rec := testutil.DoAuthRequest(http.MethodPost, "/tourdates/"+tourDateID+"/financials", `{"fee":200}`, token)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateFinancials_TourdateNotFound(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)

	rec := testutil.DoAuthRequest(http.MethodPost, "/tourdates/11111111-1111-1111-1111-111111111111/financials", `{"fee":100}`, token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateFinancials_OtherUsersTourdate404(t *testing.T) {
	testutil.TruncateTables(t)
	_, tokenA := testutil.CreateAndLoginTestUser(t)
	_, tokenB := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, tokenA)

	rec := testutil.DoAuthRequest(http.MethodPost, "/tourdates/"+tourDateID+"/financials", `{"fee":100}`, tokenB)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 (not 403, to avoid revealing existence), got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateFinancials_RequiresAuth(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, token)

	rec := testutil.DoRequest(http.MethodPost, "/tourdates/"+tourDateID+"/financials", `{"fee":100}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetFinancials(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, token)
	testutil.DoAuthRequest(http.MethodPost, "/tourdates/"+tourDateID+"/financials", `{"fee":1500,"tips":200}`, token)

	rec := testutil.DoAuthRequest(http.MethodGet, "/tourdates/"+tourDateID+"/financials", "", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var f financials.Financials
	json.Unmarshal(rec.Body.Bytes(), &f)
	if f.Fee == nil || *f.Fee != 1500 {
		t.Fatalf("expected fee 1500, got %+v", f.Fee)
	}
}

func TestGetFinancials_NotFound(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, token)

	rec := testutil.DoAuthRequest(http.MethodGet, "/tourdates/"+tourDateID+"/financials", "", token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchFinancials_PartialUpdate(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, token)
	testutil.DoAuthRequest(http.MethodPost, "/tourdates/"+tourDateID+"/financials", `{"fee":1500,"tips":200}`, token)

	rec := testutil.DoAuthRequest(http.MethodPatch, "/tourdates/"+tourDateID+"/financials", `{"fee":1750}`, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var f financials.Financials
	json.Unmarshal(rec.Body.Bytes(), &f)
	if f.Fee == nil || *f.Fee != 1750 {
		t.Fatalf("expected fee to be updated to 1750, got %+v", f.Fee)
	}
	if f.Tips == nil || *f.Tips != 200 {
		t.Fatalf("expected tips untouched, got %+v", f.Tips)
	}
}

func TestPatchFinancials_ClearToNull(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, token)
	testutil.DoAuthRequest(http.MethodPost, "/tourdates/"+tourDateID+"/financials", `{"fee":1500,"tips":200}`, token)

	rec := testutil.DoAuthRequest(http.MethodPatch, "/tourdates/"+tourDateID+"/financials", `{"tips":null}`, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var f financials.Financials
	json.Unmarshal(rec.Body.Bytes(), &f)
	if f.Tips != nil {
		t.Fatalf("expected tips to be cleared, got %v", *f.Tips)
	}
}

func TestPatchFinancials_NotFound(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, token)

	rec := testutil.DoAuthRequest(http.MethodPatch, "/tourdates/"+tourDateID+"/financials", `{"fee":100}`, token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchFinancials_NoFields(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, token)
	testutil.DoAuthRequest(http.MethodPost, "/tourdates/"+tourDateID+"/financials", `{"fee":100}`, token)

	rec := testutil.DoAuthRequest(http.MethodPatch, "/tourdates/"+tourDateID+"/financials", `{}`, token)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteFinancials(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, token)
	testutil.DoAuthRequest(http.MethodPost, "/tourdates/"+tourDateID+"/financials", `{"fee":100}`, token)

	rec := testutil.DoAuthRequest(http.MethodDelete, "/tourdates/"+tourDateID+"/financials", "", token)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testutil.DoAuthRequest(http.MethodGet, "/tourdates/"+tourDateID+"/financials", "", token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected financials to be gone, got %d", rec.Code)
	}
}

func TestDeleteFinancials_NotFound(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, token)

	rec := testutil.DoAuthRequest(http.MethodDelete, "/tourdates/"+tourDateID+"/financials", "", token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestFinancials_RepresentativeHasFullAccess(t *testing.T) {
	testutil.TruncateTables(t)
	_, artistToken := testutil.CreateTestUserOfType(t, "ARTIST")
	manager, managerToken := testutil.CreateTestUserOfType(t, "MANAGER")
	testutil.AddRepresentative(t, artistToken, manager.ID)
	tourDateID := createTourDate(t, artistToken)

	rec := testutil.DoAuthRequest(http.MethodPost, "/tourdates/"+tourDateID+"/financials", `{"fee":1500}`, managerToken)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected manager to create financials for their artist, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testutil.DoAuthRequest(http.MethodPatch, "/tourdates/"+tourDateID+"/financials", `{"tips":50}`, managerToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected manager to patch financials for their artist, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestFinancials_UnrepresentedArtist404(t *testing.T) {
	testutil.TruncateTables(t)
	_, artistToken := testutil.CreateTestUserOfType(t, "ARTIST")
	_, managerToken := testutil.CreateTestUserOfType(t, "MANAGER")
	tourDateID := createTourDate(t, artistToken)

	rec := testutil.DoAuthRequest(http.MethodGet, "/tourdates/"+tourDateID+"/financials", "", managerToken)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for an artist the manager doesn't represent, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSummary_FeeTipsAndExpenses(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, token)
	testutil.DoAuthRequest(http.MethodPost, "/tourdates/"+tourDateID+"/financials", `{"fee":1500,"tips":200}`, token)
	testutil.CreateTestExpense(t, token, tourDateID, `{"category":"TRAVEL","amount":300}`)
	testutil.CreateTestExpense(t, token, tourDateID, `{"category":"LODGING","amount":150}`)

	rec := testutil.DoAuthRequest(http.MethodGet, "/tourdates/"+tourDateID+"/summary", "", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var s financials.Summary
	if err := json.Unmarshal(rec.Body.Bytes(), &s); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if s.Fee == nil || *s.Fee != 1500 {
		t.Fatalf("expected fee 1500, got %+v", s.Fee)
	}
	if s.Tips == nil || *s.Tips != 200 {
		t.Fatalf("expected tips 200, got %+v", s.Tips)
	}
	if s.TotalExpenses != 450 {
		t.Fatalf("expected total_expenses 450, got %v", s.TotalExpenses)
	}
	if s.Net != 1250 {
		t.Fatalf("expected net 1250 (1500+200-450), got %v", s.Net)
	}
}

func TestSummary_NoFinancialsYet(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, token)
	testutil.CreateTestExpense(t, token, tourDateID, `{"category":"GEAR","amount":100}`)

	rec := testutil.DoAuthRequest(http.MethodGet, "/tourdates/"+tourDateID+"/summary", "", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var s financials.Summary
	json.Unmarshal(rec.Body.Bytes(), &s)
	if s.Fee != nil || s.Tips != nil {
		t.Fatalf("expected nil fee/tips with no financials row, got %+v", s)
	}
	if s.TotalExpenses != 100 {
		t.Fatalf("expected total_expenses 100, got %v", s.TotalExpenses)
	}
	if s.Net != -100 {
		t.Fatalf("expected net -100 (0-100), got %v", s.Net)
	}
}

func TestSummary_NoExpenses(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, token)
	testutil.DoAuthRequest(http.MethodPost, "/tourdates/"+tourDateID+"/financials", `{"fee":1000}`, token)

	rec := testutil.DoAuthRequest(http.MethodGet, "/tourdates/"+tourDateID+"/summary", "", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var s financials.Summary
	json.Unmarshal(rec.Body.Bytes(), &s)
	if s.TotalExpenses != 0 {
		t.Fatalf("expected total_expenses 0, got %v", s.TotalExpenses)
	}
	if s.Net != 1000 {
		t.Fatalf("expected net 1000, got %v", s.Net)
	}
}

func TestSummary_TourdateNotFound(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)

	rec := testutil.DoAuthRequest(http.MethodGet, "/tourdates/11111111-1111-1111-1111-111111111111/summary", "", token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}
