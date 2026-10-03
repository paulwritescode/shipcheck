package seed

import (
	"testing"

	"github.com/paulwritescode/shipcheck/internal/domain"
)

// toLaunchData assembles a domain.LaunchData from the seed input so the engine
// can be run against it directly (mirrors what the repository stores).
func toLaunchData() domain.LaunchData {
	in := TeamInbox()
	return domain.LaunchData{
		ID: in.ID, Name: in.Name, Description: in.Description, TargetDate: in.TargetDate,
		Owner: in.Owner, ProductArea: in.ProductArea, RepositoryURL: in.RepositoryURL,
		Brief: in.Brief, ChecklistItems: in.ChecklistItems, Risks: in.Risks,
	}
}

// Req 12.1-12.6: fixed shape of the seeded launch.
func TestSeedShape(t *testing.T) {
	in := TeamInbox()

	if in.Name != "Team Inbox 2.0" {
		t.Errorf("name: got %q", in.Name)
	}
	if in.TargetDate != "2026-10-16" {
		t.Errorf("target date: got %q", in.TargetDate)
	}

	// 18 checklist items.
	if len(in.ChecklistItems) != 18 {
		t.Fatalf("want 18 checklist items, got %d", len(in.ChecklistItems))
	}
	var complete, inProgress, blocked int
	var unowned int
	for _, it := range in.ChecklistItems {
		switch it.CompletionState {
		case domain.CompletionComplete:
			complete++
		case domain.CompletionInProgress:
			inProgress++
		case domain.CompletionBlocked:
			blocked++
		}
		if it.Owner == "" {
			unowned++
		}
	}
	if complete != 13 || inProgress != 3 || blocked != 2 {
		t.Fatalf("completion mix: got complete=%d inProgress=%d blocked=%d (want 13/3/2)", complete, inProgress, blocked)
	}
	if unowned < 1 {
		t.Fatalf("want at least one unowned item, got %d", unowned)
	}

	// 4 risks, exactly one High severity.
	if len(in.Risks) != 4 {
		t.Fatalf("want 4 risks, got %d", len(in.Risks))
	}
	high := 0
	for _, r := range in.Risks {
		if r.Severity == domain.SeverityHigh {
			high++
		}
	}
	if high != 1 {
		t.Fatalf("want exactly one High-severity risk, got %d", high)
	}

	// Incomplete critical rollback item in Launch Operations.
	var foundRollback bool
	for _, it := range in.ChecklistItems {
		if it.ID == RollbackItemID {
			foundRollback = true
			if it.Category != domain.CategoryLaunchOperations {
				t.Errorf("rollback item category: got %q", it.Category)
			}
			if !it.IsCritical {
				t.Error("rollback item should be critical")
			}
			if domain.IsCompleteForScoring(it.CompletionState) {
				t.Error("rollback item should be incomplete initially")
			}
		}
	}
	if !foundRollback {
		t.Fatalf("rollback item %q not found", RollbackItemID)
	}

	// Public preview URL + GitHub-style evidence present somewhere.
	var sawPreview, sawGitHub bool
	for _, it := range in.ChecklistItems {
		for _, e := range it.Evidence {
			if e.Note == "Public preview URL" {
				sawPreview = true
			}
			if contains(e.URL, "github.com") {
				sawGitHub = true
			}
		}
	}
	for _, r := range in.Risks {
		for _, e := range r.Evidence {
			if contains(e.URL, "github.com") {
				sawGitHub = true
			}
		}
	}
	if !sawPreview {
		t.Error("expected a public preview URL in evidence")
	}
	if !sawGitHub {
		t.Error("expected GitHub-style evidence links")
	}
}

// Req 12.7: initial readiness is Not Ready.
func TestSeedInitialNotReady(t *testing.T) {
	a := domain.ComputeReadiness(toLaunchData())
	if a.Status != domain.StatusNotReady {
		t.Fatalf("initial seed status: want Not Ready, got %s\nreasons: %+v", a.Status, a.Reasons)
	}
}

// Req 12.8: the documented demo edits move the launch to Conditionally Ready.
// Edits: complete the critical rollback item (with evidence + owner) and
// resolve the high-severity blocking risk.
func TestSeedDemoTransitionToConditionallyReady(t *testing.T) {
	data := toLaunchData()

	for i := range data.ChecklistItems {
		if data.ChecklistItems[i].ID == RollbackItemID {
			data.ChecklistItems[i].CompletionState = domain.CompletionComplete
			data.ChecklistItems[i].Owner = "Lena"
			data.ChecklistItems[i].Evidence = []domain.Evidence{{ID: "e-rollback", URL: "https://github.com/example/team-inbox/pull/240", Note: "Rollback runbook"}}
		}
	}
	for i := range data.Risks {
		if data.Risks[i].ID == HighRiskID {
			data.Risks[i].Status = domain.RiskResolved
		}
	}

	a := domain.ComputeReadiness(data)
	if a.Status != domain.StatusConditionallyReady {
		t.Fatalf("after demo edits: want Conditionally Ready, got %s\nreasons: %+v", a.Status, a.Reasons)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && indexOf(s, sub) >= 0
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
