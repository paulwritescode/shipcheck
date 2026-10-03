package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/paulwritescode/shipcheck/internal/domain"
	"github.com/paulwritescode/shipcheck/internal/repo"
)

func newTestServer() http.Handler {
	return NewRouter(repo.NewInMemoryLaunchRepository())
}

// do performs a request against the handler and returns the recorder.
func do(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, path, nil)
	} else {
		r = httptest.NewRequest(method, path, bytes.NewBufferString(body))
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	return rec
}

// createLaunch creates a launch and returns its id.
func createLaunch(t *testing.T, h http.Handler) string {
	t.Helper()
	rec := do(t, h, "POST", "/api/launches", `{"name":"Team Inbox","targetDate":"2026-10-16","owner":"Alice"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create launch: want 201, got %d (%s)", rec.Code, rec.Body.String())
	}
	var l domain.Launch
	if err := json.Unmarshal(rec.Body.Bytes(), &l); err != nil {
		t.Fatalf("decode launch: %v", err)
	}
	if l.ID == "" {
		t.Fatal("created launch has no id")
	}
	return l.ID
}

func TestHealth(t *testing.T) {
	rec := do(t, newTestServer(), "GET", "/api/health", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("health: want 200, got %d", rec.Code)
	}
}

func TestCreateLaunchValidation(t *testing.T) {
	h := newTestServer()
	// Empty name + bad date -> 400 with field errors.
	rec := do(t, h, "POST", "/api/launches", `{"name":"  ","targetDate":"2026-02-30"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d (%s)", rec.Code, rec.Body.String())
	}
	var vb validationBody
	if err := json.Unmarshal(rec.Body.Bytes(), &vb); err != nil {
		t.Fatal(err)
	}
	gotFields := map[string]bool{}
	for _, f := range vb.Fields {
		gotFields[f.Field] = true
	}
	if !gotFields["name"] || !gotFields["targetDate"] {
		t.Fatalf("expected name+targetDate errors, got %+v", vb.Fields)
	}
}

func TestCreateLaunchBadURL(t *testing.T) {
	rec := do(t, newTestServer(), "POST", "/api/launches",
		`{"name":"n","targetDate":"2026-10-16","owner":"a","repositoryUrl":"not a url"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad url: want 400, got %d", rec.Code)
	}
}

func TestChecklistLifecycleAndRecompute(t *testing.T) {
	h := newTestServer()
	id := createLaunch(t, h)

	// Add an incomplete critical item -> Not Ready.
	rec := do(t, h, "POST", "/api/launches/"+id+"/items",
		`{"title":"migrate","category":"Engineering","priority":"High","completionState":"blocked","isCritical":true}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("add item: want 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	var l domain.Launch
	_ = json.Unmarshal(rec.Body.Bytes(), &l)
	if l.Assessment.Status != domain.StatusNotReady {
		t.Fatalf("after add incomplete critical: want Not Ready, got %s", l.Assessment.Status)
	}
	itemID := l.ChecklistItems[0].ID

	// Complete it with evidence -> Ready.
	rec = do(t, h, "PATCH", "/api/launches/"+id+"/items/"+itemID,
		`{"title":"migrate","category":"Engineering","priority":"High","completionState":"complete","isCritical":true,"evidence":[{"id":"e1","url":"https://example.com/pr/1"}]}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("update item: want 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &l)
	if l.Assessment.Status != domain.StatusReady {
		t.Fatalf("after complete+evidence: want Ready, got %s", l.Assessment.Status)
	}

	// Delete it -> empty launch -> Needs Review.
	rec = do(t, h, "DELETE", "/api/launches/"+id+"/items/"+itemID, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("delete item: want 200, got %d", rec.Code)
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &l)
	if l.Assessment.Status != domain.StatusNeedsReview {
		t.Fatalf("after delete: want Needs Review, got %s", l.Assessment.Status)
	}
}

func TestItemValidation(t *testing.T) {
	h := newTestServer()
	id := createLaunch(t, h)
	// Bad category -> 400.
	rec := do(t, h, "POST", "/api/launches/"+id+"/items", `{"title":"x","category":"Nonsense"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad category: want 400, got %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestRiskLifecycle(t *testing.T) {
	h := newTestServer()
	id := createLaunch(t, h)
	rec := do(t, h, "POST", "/api/launches/"+id+"/risks",
		`{"title":"scale","severity":"High","likelihood":"High","status":"Open"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("add risk: want 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	var l domain.Launch
	_ = json.Unmarshal(rec.Body.Bytes(), &l)
	// A blocking risk forces Not Ready.
	if l.Assessment.Status != domain.StatusNotReady {
		t.Fatalf("blocking risk: want Not Ready, got %s", l.Assessment.Status)
	}
	riskID := l.Risks[0].ID
	// Resolve it -> no longer blocking.
	rec = do(t, h, "PATCH", "/api/launches/"+id+"/risks/"+riskID,
		`{"title":"scale","severity":"High","likelihood":"High","status":"Resolved"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("update risk: want 200, got %d", rec.Code)
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &l)
	if l.Assessment.Status == domain.StatusNotReady {
		t.Fatalf("after resolve: should no longer be Not Ready from this risk")
	}
}

func TestUpdateLaunchOwner(t *testing.T) {
	h := newTestServer()
	id := createLaunch(t, h)
	// Clear the owner -> Needs Review (empty + no owner both apply).
	rec := do(t, h, "PATCH", "/api/launches/"+id, `{"owner":""}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("patch owner: want 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	var l domain.Launch
	_ = json.Unmarshal(rec.Body.Bytes(), &l)
	if l.Assessment.Status != domain.StatusNeedsReview {
		t.Fatalf("cleared owner: want Needs Review, got %s", l.Assessment.Status)
	}
}

func TestGetAndDeleteLaunch(t *testing.T) {
	h := newTestServer()
	id := createLaunch(t, h)
	if rec := do(t, h, "GET", "/api/launches/"+id, ""); rec.Code != http.StatusOK {
		t.Fatalf("get: want 200, got %d", rec.Code)
	}
	if rec := do(t, h, "GET", "/api/launches/missing", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("get missing: want 404, got %d", rec.Code)
	}
	if rec := do(t, h, "DELETE", "/api/launches/"+id, ""); rec.Code != http.StatusNoContent {
		t.Fatalf("delete: want 204, got %d", rec.Code)
	}
	if rec := do(t, h, "GET", "/api/launches/"+id, ""); rec.Code != http.StatusNotFound {
		t.Fatalf("get after delete: want 404, got %d", rec.Code)
	}
}

func TestShareEnableRegenerateDisable(t *testing.T) {
	h := newTestServer()
	id := createLaunch(t, h)

	rec := do(t, h, "POST", "/api/launches/"+id+"/share", `{"action":"enable"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("enable share: want 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	var l domain.Launch
	_ = json.Unmarshal(rec.Body.Bytes(), &l)
	if l.Share == nil || !l.Share.Enabled || l.Share.Token == "" {
		t.Fatalf("share not enabled: %+v", l.Share)
	}
	first := l.Share.Token

	// Regenerate -> new token.
	rec = do(t, h, "POST", "/api/launches/"+id+"/share", `{"action":"regenerate"}`)
	_ = json.Unmarshal(rec.Body.Bytes(), &l)
	if l.Share.Token == first {
		t.Fatal("regenerate did not change the token")
	}

	// Disable.
	rec = do(t, h, "POST", "/api/launches/"+id+"/share", `{"action":"disable"}`)
	_ = json.Unmarshal(rec.Body.Bytes(), &l)
	if l.Share.Enabled {
		t.Fatal("disable did not disable the link")
	}

	// Bad action -> 400.
	if rec := do(t, h, "POST", "/api/launches/"+id+"/share", `{"action":"bogus"}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("bad action: want 400, got %d", rec.Code)
	}
}

func TestBadJSON(t *testing.T) {
	h := newTestServer()
	id := createLaunch(t, h)
	rec := do(t, h, "POST", "/api/launches/"+id+"/items", `{not json`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad json: want 400, got %d", rec.Code)
	}
}

func TestSeedNotWired(t *testing.T) {
	// Until the seed task wires seedFn, requesting a seeded launch is 501.
	if seedFn != nil {
		t.Skip("seedFn is wired; seed behavior tested elsewhere")
	}
	rec := do(t, newTestServer(), "POST", "/api/launches", `{"seed":true}`)
	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("seed not wired: want 501, got %d", rec.Code)
	}
}

// sanity: validation body is valid JSON with the documented shape.
func TestValidationBodyShape(t *testing.T) {
	rec := do(t, newTestServer(), "POST", "/api/launches", `{"name":""}`)
	if !strings.Contains(rec.Body.String(), `"fields"`) {
		t.Fatalf("validation body missing fields: %s", rec.Body.String())
	}
}
