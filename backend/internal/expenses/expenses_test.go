package expenses_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"skejio/backend/internal/expenses"
	"skejio/backend/internal/testutil"
)

func TestMain(m *testing.M) { testutil.Run(m) }

func createTourDate(t *testing.T, token string) string {
	t.Helper()
	td := testutil.CreateTestTourDate(t, token, `{"date":"2026-09-15","city":"Austin","venue":"Moody Center"}`)
	return td.ID.String()
}

func TestCreateExpense(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, token)

	rec := testutil.DoAuthRequest(http.MethodPost, "/tourdates/"+tourDateID+"/expenses",
		`{"category":"TRAVEL","amount":250.75,"description":"Flights"}`, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var e expenses.Expense
	if err := json.Unmarshal(rec.Body.Bytes(), &e); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if e.TourDateID.String() != tourDateID {
		t.Fatalf("expected tourdate_id %s, got %s", tourDateID, e.TourDateID)
	}
	if e.Category != "TRAVEL" || e.Amount != 250.75 {
		t.Fatalf("unexpected expense: %+v", e)
	}
	if e.Description == nil || *e.Description != "Flights" {
		t.Fatalf("expected description to be set, got %+v", e.Description)
	}
}

func TestCreateExpense_NullableDescription(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, token)

	rec := testutil.DoAuthRequest(http.MethodPost, "/tourdates/"+tourDateID+"/expenses", `{"category":"OTHER","amount":10}`, token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var e expenses.Expense
	json.Unmarshal(rec.Body.Bytes(), &e)
	if e.Description != nil {
		t.Fatalf("expected nil description, got %v", *e.Description)
	}
}

func TestCreateExpense_InvalidCategory(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, token)

	rec := testutil.DoAuthRequest(http.MethodPost, "/tourdates/"+tourDateID+"/expenses", `{"category":"SNACKS","amount":10}`, token)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateExpense_MissingAmount(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, token)

	rec := testutil.DoAuthRequest(http.MethodPost, "/tourdates/"+tourDateID+"/expenses", `{"category":"OTHER"}`, token)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateExpense_TourdateNotFound(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)

	rec := testutil.DoAuthRequest(http.MethodPost, "/tourdates/11111111-1111-1111-1111-111111111111/expenses", `{"category":"OTHER","amount":10}`, token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateExpense_OtherUsersTourdate404(t *testing.T) {
	testutil.TruncateTables(t)
	_, tokenA := testutil.CreateAndLoginTestUser(t)
	_, tokenB := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, tokenA)

	rec := testutil.DoAuthRequest(http.MethodPost, "/tourdates/"+tourDateID+"/expenses", `{"category":"OTHER","amount":10}`, tokenB)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 (not 403, to avoid revealing existence), got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateExpense_RequiresAuth(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, token)

	rec := testutil.DoRequest(http.MethodPost, "/tourdates/"+tourDateID+"/expenses", `{"category":"OTHER","amount":10}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListExpenses_OrderedByCreatedAt(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, token)

	testutil.CreateTestExpense(t, token, tourDateID, `{"category":"TRAVEL","amount":100}`)
	testutil.CreateTestExpense(t, token, tourDateID, `{"category":"LODGING","amount":200}`)

	rec := testutil.DoAuthRequest(http.MethodGet, "/tourdates/"+tourDateID+"/expenses", "", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var list []expenses.Expense
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 expenses, got %d", len(list))
	}
	if list[0].Category != "TRAVEL" || list[1].Category != "LODGING" {
		t.Fatalf("expected results ordered by created_at, got %s then %s", list[0].Category, list[1].Category)
	}
}

func TestListExpenses_ScopedToTourdate(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateA := createTourDate(t, token)
	tourDateB := createTourDate(t, token)

	testutil.CreateTestExpense(t, token, tourDateA, `{"category":"TRAVEL","amount":100}`)
	testutil.CreateTestExpense(t, token, tourDateB, `{"category":"LODGING","amount":200}`)

	rec := testutil.DoAuthRequest(http.MethodGet, "/tourdates/"+tourDateA+"/expenses", "", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var list []expenses.Expense
	json.Unmarshal(rec.Body.Bytes(), &list)
	if len(list) != 1 || list[0].Category != "TRAVEL" {
		t.Fatalf("expected only tourdate A's expense, got %+v", list)
	}
}

func TestGetExpense(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, token)
	created := testutil.CreateTestExpense(t, token, tourDateID, `{"category":"CREW","amount":500}`)

	rec := testutil.DoAuthRequest(http.MethodGet, "/tourdates/"+tourDateID+"/expenses/"+created.ID.String(), "", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var e expenses.Expense
	json.Unmarshal(rec.Body.Bytes(), &e)
	if e.ID != created.ID {
		t.Fatalf("expected id %s, got %s", created.ID, e.ID)
	}
}

func TestGetExpense_NotFound(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, token)

	rec := testutil.DoAuthRequest(http.MethodGet, "/tourdates/"+tourDateID+"/expenses/11111111-1111-1111-1111-111111111111", "", token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetExpense_WrongTourdate404(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateA := createTourDate(t, token)
	tourDateB := createTourDate(t, token)
	created := testutil.CreateTestExpense(t, token, tourDateA, `{"category":"CREW","amount":500}`)

	rec := testutil.DoAuthRequest(http.MethodGet, "/tourdates/"+tourDateB+"/expenses/"+created.ID.String(), "", token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for an expense that belongs to a different tourdate, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchExpense_PartialUpdate(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, token)
	created := testutil.CreateTestExpense(t, token, tourDateID, `{"category":"GEAR","amount":300,"description":"Backline rental"}`)

	rec := testutil.DoAuthRequest(http.MethodPatch, "/tourdates/"+tourDateID+"/expenses/"+created.ID.String(), `{"amount":350}`, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var e expenses.Expense
	json.Unmarshal(rec.Body.Bytes(), &e)
	if e.Amount != 350 {
		t.Fatalf("expected amount to be updated to 350, got %v", e.Amount)
	}
	if e.Category != "GEAR" || e.Description == nil || *e.Description != "Backline rental" {
		t.Fatalf("expected other fields untouched, got %+v", e)
	}
}

func TestPatchExpense_ClearDescriptionToNull(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, token)
	created := testutil.CreateTestExpense(t, token, tourDateID, `{"category":"GEAR","amount":300,"description":"Backline rental"}`)

	rec := testutil.DoAuthRequest(http.MethodPatch, "/tourdates/"+tourDateID+"/expenses/"+created.ID.String(), `{"description":null}`, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var e expenses.Expense
	json.Unmarshal(rec.Body.Bytes(), &e)
	if e.Description != nil {
		t.Fatalf("expected description to be cleared, got %v", *e.Description)
	}
}

func TestPatchExpense_NullCategoryRejected(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, token)
	created := testutil.CreateTestExpense(t, token, tourDateID, `{"category":"GEAR","amount":300}`)

	rec := testutil.DoAuthRequest(http.MethodPatch, "/tourdates/"+tourDateID+"/expenses/"+created.ID.String(), `{"category":null}`, token)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchExpense_NotFound(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, token)

	rec := testutil.DoAuthRequest(http.MethodPatch, "/tourdates/"+tourDateID+"/expenses/11111111-1111-1111-1111-111111111111", `{"amount":10}`, token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPatchExpense_NoFields(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, token)
	created := testutil.CreateTestExpense(t, token, tourDateID, `{"category":"GEAR","amount":300}`)

	rec := testutil.DoAuthRequest(http.MethodPatch, "/tourdates/"+tourDateID+"/expenses/"+created.ID.String(), `{}`, token)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteExpense(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, token)
	created := testutil.CreateTestExpense(t, token, tourDateID, `{"category":"GEAR","amount":300}`)

	rec := testutil.DoAuthRequest(http.MethodDelete, "/tourdates/"+tourDateID+"/expenses/"+created.ID.String(), "", token)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = testutil.DoAuthRequest(http.MethodGet, "/tourdates/"+tourDateID+"/expenses/"+created.ID.String(), "", token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected expense to be gone, got %d", rec.Code)
	}
}

func TestDeleteExpense_NotFound(t *testing.T) {
	testutil.TruncateTables(t)
	_, token := testutil.CreateAndLoginTestUser(t)
	tourDateID := createTourDate(t, token)

	rec := testutil.DoAuthRequest(http.MethodDelete, "/tourdates/"+tourDateID+"/expenses/11111111-1111-1111-1111-111111111111", "", token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestExpenses_RepresentativeHasFullAccess(t *testing.T) {
	testutil.TruncateTables(t)
	_, artistToken := testutil.CreateTestUserOfType(t, "ARTIST")
	manager, managerToken := testutil.CreateTestUserOfType(t, "MANAGER")
	testutil.AddRepresentative(t, artistToken, manager.ID)
	tourDateID := createTourDate(t, artistToken)

	rec := testutil.DoAuthRequest(http.MethodPost, "/tourdates/"+tourDateID+"/expenses", `{"category":"CREW","amount":500}`, managerToken)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected manager to create an expense for their artist, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestExpenses_UnrepresentedArtist404(t *testing.T) {
	testutil.TruncateTables(t)
	_, artistToken := testutil.CreateTestUserOfType(t, "ARTIST")
	_, managerToken := testutil.CreateTestUserOfType(t, "MANAGER")
	tourDateID := createTourDate(t, artistToken)

	rec := testutil.DoAuthRequest(http.MethodGet, "/tourdates/"+tourDateID+"/expenses", "", managerToken)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for an artist the manager doesn't represent, got %d: %s", rec.Code, rec.Body.String())
	}
}
