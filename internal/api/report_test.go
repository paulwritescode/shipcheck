package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/paulwritescode/shipcheck/internal/domain"
)

// enableShare enables sharing for a launch and returns the fresh token.
func enableShare(t *testing.T, h http.Handler, id string) string {
	t.Helper()
	rec := do(t, h, "POST", "/api/launches/"+id+"/share", `{"action":"enable"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("enable share: want 200, got %d", rec.Code)
	}
	var l domain.Launch
	_ = json.Unmarshal(rec.Body.Bytes(), &l)
	if l.Share == nil || l.Share.Token == "" {
		t.Fatal("no share token returned")
	}
	return l.Share.Token
}

func TestPublicReportResolvesEnabledToken(t *testing.T) {
	h := newTestServer()
	id := createLaunch(t, h)
	do(t, h, "POST", "/api/launches/"+id+"/items",
		`{"title":"migrate","category":"Engineering","completionState":"complete"}`)
	token := enableShare(t, h, id)

	rec := do(t, h, "GET", "/api/r/"+token, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("public report: want 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	var report domain.PublicReport
	if err := json.Unmarshal(rec.Body.Bytes(), &report); err != nil {
		t.Fatalf("decode report: %v", err)
	}
	if report.Name == "" || report.Status == "" {
		t.Fatalf("report missing name/status: %+v", report)
	}
	if report.LastUpdated == "" {
		t.Fatal("report missing lastUpdated")
	}
}

func TestPublicReportUnknownTokenUnavailable(t *testing.T) {
	rec := do(t, newTestServer(), "GET", "/api/r/nonexistent", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unknown token: want 404, got %d", rec.Code)
	}
	if !jsonHasKey(rec.Body.Bytes(), "error") {
		t.Fatal("unavailable response should carry an error message")
	}
}

func TestPublicReportDisabledTokenUnavailable(t *testing.T) {
	h := newTestServer()
	id := createLaunch(t, h)
	token := enableShare(t, h, id)

	// Resolves while enabled.
	if rec := do(t, h, "GET", "/api/r/"+token, ""); rec.Code != http.StatusOK {
		t.Fatalf("enabled token should resolve, got %d", rec.Code)
	}
	// Disable sharing.
	do(t, h, "POST", "/api/launches/"+id+"/share", `{"action":"disable"}`)
	// Now unavailable.
	if rec := do(t, h, "GET", "/api/r/"+token, ""); rec.Code != http.StatusNotFound {
		t.Fatalf("disabled token: want 404, got %d", rec.Code)
	}
}

func TestPublicReportRegeneratedTokenUnavailable(t *testing.T) {
	h := newTestServer()
	id := createLaunch(t, h)
	first := enableShare(t, h, id)

	// Regenerate -> old token must stop resolving.
	do(t, h, "POST", "/api/launches/"+id+"/share", `{"action":"regenerate"}`)
	if rec := do(t, h, "GET", "/api/r/"+first, ""); rec.Code != http.StatusNotFound {
		t.Fatalf("regenerated: old token should 404, got %d", rec.Code)
	}
}

func jsonHasKey(b []byte, key string) bool {
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return false
	}
	_, ok := m[key]
	return ok
}
