package domain

import "testing"

// These example-based sanity checks exercise each readiness tier with concrete
// launches. The exhaustive universal guarantees are covered by the
// property-based suite (added in the next task).

func ownedBase() LaunchData {
	return LaunchData{ID: "l1", Name: "Test", TargetDate: "2026-10-16", Owner: "Alice"}
}

func TestEmptyLaunchNeedsReview(t *testing.T) {
	got := ComputeReadiness(LaunchData{ID: "l1", Name: "T", TargetDate: "2026-10-16"})
	if got.Status != StatusNeedsReview {
		t.Fatalf("empty launch: want Needs Review, got %s", got.Status)
	}
}

func TestNoOwnerNeedsReview(t *testing.T) {
	l := ownedBase()
	l.Owner = ""
	l.ChecklistItems = []ChecklistItem{{ID: "i1", Title: "x", Category: CategoryEngineering, CompletionState: CompletionComplete}}
	got := ComputeReadiness(l)
	if got.Status != StatusNeedsReview {
		t.Fatalf("no owner: want Needs Review, got %s", got.Status)
	}
}

func TestCompleteCriticalWithoutEvidenceNeedsReview(t *testing.T) {
	l := ownedBase()
	l.ChecklistItems = []ChecklistItem{{
		ID: "i1", Title: "crit", Category: CategoryEngineering,
		CompletionState: CompletionComplete, IsCritical: true, // no evidence
	}}
	got := ComputeReadiness(l)
	if got.Status != StatusNeedsReview {
		t.Fatalf("complete critical w/o evidence: want Needs Review, got %s", got.Status)
	}
}

func TestIncompleteCriticalNotReady(t *testing.T) {
	l := ownedBase()
	l.ChecklistItems = []ChecklistItem{{
		ID: "i1", Title: "crit", Category: CategoryEngineering,
		CompletionState: CompletionBlocked, IsCritical: true,
	}}
	got := ComputeReadiness(l)
	if got.Status != StatusNotReady {
		t.Fatalf("incomplete critical: want Not Ready, got %s", got.Status)
	}
	if len(got.Reasons) != 1 || got.Reasons[0].RuleID != RuleIncompleteCritical {
		t.Fatalf("expected one incomplete-critical reason, got %+v", got.Reasons)
	}
}

func TestBlockingRiskNotReady(t *testing.T) {
	l := ownedBase()
	l.ChecklistItems = []ChecklistItem{{ID: "i1", Title: "x", Category: CategoryEngineering, CompletionState: CompletionComplete}}
	l.Risks = []Risk{{ID: "r1", Title: "boom", Severity: SeverityHigh, Status: RiskOpen}}
	got := ComputeReadiness(l)
	if got.Status != StatusNotReady {
		t.Fatalf("blocking risk: want Not Ready, got %s", got.Status)
	}
}

func TestConditionallyReadyNonCriticalGap(t *testing.T) {
	l := ownedBase()
	l.ChecklistItems = []ChecklistItem{
		{ID: "i1", Title: "crit", Category: CategoryEngineering, CompletionState: CompletionComplete, IsCritical: true, Evidence: []Evidence{{ID: "e1", Note: "done"}}},
		{ID: "i2", Title: "nice", Category: CategoryDocumentation, CompletionState: CompletionInProgress},
	}
	got := ComputeReadiness(l)
	if got.Status != StatusConditionallyReady {
		t.Fatalf("non-critical gap: want Conditionally Ready, got %s", got.Status)
	}
}

func TestReadyAllComplete(t *testing.T) {
	l := ownedBase()
	l.ChecklistItems = []ChecklistItem{
		{ID: "i1", Title: "crit", Category: CategoryEngineering, CompletionState: CompletionComplete, IsCritical: true, Evidence: []Evidence{{ID: "e1", Note: "done"}}},
		{ID: "i2", Title: "nice", Category: CategoryDocumentation, CompletionState: CompletionComplete},
	}
	l.Risks = []Risk{{ID: "r1", Title: "minor", Severity: SeverityLow, Status: RiskOpen}}
	got := ComputeReadiness(l)
	if got.Status != StatusReady {
		t.Fatalf("all complete: want Ready, got %s", got.Status)
	}
}

func TestCategoryStatesCountsAndClassification(t *testing.T) {
	l := ownedBase()
	l.ChecklistItems = []ChecklistItem{
		{ID: "i1", Title: "a", Category: CategoryEngineering, CompletionState: CompletionComplete},
		{ID: "i2", Title: "b", Category: CategoryEngineering, CompletionState: CompletionIncomplete},
		{ID: "i3", Title: "c", Category: CategoryDocumentation, CompletionState: CompletionComplete},
	}
	cats := ComputeCategoryStates(l)
	if len(cats) != len(Categories) {
		t.Fatalf("want %d categories, got %d", len(Categories), len(cats))
	}
	for _, c := range cats {
		switch c.Category {
		case CategoryEngineering:
			if c.CompleteCount != 1 || c.TotalCount != 2 || c.State != CategoryPartial {
				t.Errorf("engineering: got %+v", c)
			}
		case CategoryDocumentation:
			if c.CompleteCount != 1 || c.TotalCount != 1 || c.State != CategoryComplete {
				t.Errorf("documentation: got %+v", c)
			}
		default:
			if c.TotalCount != 0 || c.State != CategoryEmpty {
				t.Errorf("%s: want empty, got %+v", c.Category, c)
			}
		}
	}
}
