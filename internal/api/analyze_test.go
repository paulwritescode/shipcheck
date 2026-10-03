package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/paulwritescode/shipcheck/internal/advisor"
)

// analyzeResponse mirrors advisor.Result for decoding in tests.
type analyzeResponse struct {
	Analysis struct {
		Recommendations    []advisor.Recommendation `json:"recommendations"`
		SuggestedQuestions []string                 `json:"suggested_questions"`
	} `json:"analysis"`
	Provider string `json:"provider"`
}

func TestAnalyzeReturnsRecommendations(t *testing.T) {
	h := newTestServer()
	id := createLaunch(t, h)
	// Add an incomplete critical item so the advisor has a blocker to surface.
	do(t, h, "POST", "/api/launches/"+id+"/items",
		`{"title":"migrate","category":"Engineering","priority":"High","completionState":"blocked","isCritical":true}`)

	rec := do(t, h, "POST", "/api/launches/"+id+"/analyze", `{"action":"analyze"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("analyze: want 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	var resp analyzeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Provider != "fallback" {
		t.Fatalf("default provider: want fallback, got %q", resp.Provider)
	}
	if len(resp.Analysis.Recommendations) == 0 {
		t.Fatal("expected recommendations for a launch with an incomplete critical item")
	}
	// Every recommendation references a real entity id (validated server-side).
	// We just check they are non-empty here; integrity is proven by Property 30.
	for _, r := range resp.Analysis.Recommendations {
		if r.Title == "" || r.Reason == "" {
			t.Fatal("recommendation missing title/reason")
		}
	}
}

func TestAnalyzeEmptyBodyDefaultsToAnalyze(t *testing.T) {
	h := newTestServer()
	id := createLaunch(t, h)
	rec := do(t, h, "POST", "/api/launches/"+id+"/analyze", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("analyze empty body: want 200, got %d", rec.Code)
	}
}

func TestAnalyzeMissingLaunch(t *testing.T) {
	rec := do(t, newTestServer(), "POST", "/api/launches/nope/analyze", `{"action":"analyze"}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("analyze missing launch: want 404, got %d", rec.Code)
	}
}

func TestAnalyzeDoesNotMutateLaunch(t *testing.T) {
	h := newTestServer()
	id := createLaunch(t, h)
	do(t, h, "POST", "/api/launches/"+id+"/items",
		`{"title":"x","category":"Engineering","completionState":"incomplete"}`)

	// Capture state before analysis.
	before := do(t, h, "GET", "/api/launches/"+id, "").Body.String()
	do(t, h, "POST", "/api/launches/"+id+"/analyze", `{"action":"find_blockers"}`)
	after := do(t, h, "GET", "/api/launches/"+id, "").Body.String()

	if before != after {
		t.Fatal("analyze mutated the launch")
	}
}
