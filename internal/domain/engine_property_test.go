package domain

import (
	"reflect"
	"testing"

	"pgregory.net/rapid"
)

// Property-based tests for the Readiness_Engine (design.md Properties 1-17).
// Each test is tagged "Feature: shipcheck, Property N" and runs rapid's
// default minimum of 100 checks. The pure engine is exercised directly with
// no mocks.

// validStatuses is the exhaustive set of readiness statuses.
var validStatuses = map[ReadinessStatus]bool{
	StatusReady: true, StatusConditionallyReady: true,
	StatusNotReady: true, StatusNeedsReview: true,
}

// Feature: shipcheck, Property 1: Totality — exactly one status for any launch.
// Validates: Requirements 6.13, 6.14
func TestProperty1_Totality(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		l := launchGen().Draw(t, "launch")
		a := ComputeReadiness(l)
		if !validStatuses[a.Status] {
			t.Fatalf("status %q is not one of the four valid statuses", a.Status)
		}
	})
}

// Feature: shipcheck, Property 2: Determinism and idempotence of assessment.
// Validates: Requirements 6.15
func TestProperty2_Determinism(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		l := launchGen().Draw(t, "launch")
		a1 := ComputeReadiness(l)
		a2 := ComputeReadiness(l)
		if !reflect.DeepEqual(a1, a2) {
			t.Fatalf("non-deterministic: %+v != %+v", a1, a2)
		}
		// Structural clone yields the identical assessment too.
		clone := l
		clone.ChecklistItems = append([]ChecklistItem(nil), l.ChecklistItems...)
		clone.Risks = append([]Risk(nil), l.Risks...)
		a3 := ComputeReadiness(clone)
		if !reflect.DeepEqual(a1, a3) {
			t.Fatalf("clone differs: %+v != %+v", a1, a3)
		}
	})
}

// Feature: shipcheck, Property 3: Completion-state scoring mapping.
// Validates: Requirements 6.2
func TestProperty3_CompletionScoringMapping(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		s := completionGen.Draw(t, "state")
		got := IsCompleteForScoring(s)
		want := s == CompletionComplete
		if got != want {
			t.Fatalf("IsCompleteForScoring(%q) = %v, want %v", s, got, want)
		}
	})
}

// Feature: shipcheck, Property 4: Precedence — the highest matching tier wins.
// Validates: Requirements 6.3
// Strategy: construct a launch that satisfies BOTH a Needs-Review trigger
// (complete critical without evidence) AND a Not-Ready trigger (a blocking
// risk). The status must be the higher-precedence Needs Review.
func TestProperty4_Precedence(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		l := ownedNonEmptyGen().Draw(t, "base")
		l.ChecklistItems = append(l.ChecklistItems, ChecklistItem{
			ID: "crit-ne", Title: "claimed", Category: CategoryEngineering,
			Priority: PriorityHigh, CompletionState: CompletionComplete, IsCritical: true,
		})
		l.Risks = append(l.Risks, Risk{
			ID: "block", Title: "boom", Severity: SeverityHigh, Status: RiskOpen,
		})
		if got := ComputeReadiness(l).Status; got != StatusNeedsReview {
			t.Fatalf("precedence: want Needs Review, got %s", got)
		}
	})
}

// Feature: shipcheck, Property 5: Needs Review on an empty launch.
// Validates: Requirements 6.4, 1.9
func TestProperty5_EmptyLaunchNeedsReview(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		l := emptyLaunchGen().Draw(t, "empty")
		if got := ComputeReadiness(l).Status; got != StatusNeedsReview {
			t.Fatalf("empty launch: want Needs Review, got %s", got)
		}
	})
}

// Feature: shipcheck, Property 6: Needs Review on absent owner.
// Validates: Requirements 6.5
func TestProperty6_NoOwnerNeedsReview(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		l := launchGen().Draw(t, "launch")
		l.Owner = ""
		// Ensure not the empty-launch case so we isolate the owner rule;
		// either way the result is Needs Review, which is what we assert.
		if got := ComputeReadiness(l).Status; got != StatusNeedsReview {
			t.Fatalf("no owner: want Needs Review, got %s", got)
		}
	})
}

// Feature: shipcheck, Property 7: Needs Review on a completion claim without evidence.
// Validates: Requirements 6.6
func TestProperty7_CompleteCriticalNoEvidence(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		l := completeCriticalNoEvidenceGen().Draw(t, "launch")
		if got := ComputeReadiness(l).Status; got != StatusNeedsReview {
			t.Fatalf("complete critical w/o evidence: want Needs Review, got %s", got)
		}
	})
}

// Feature: shipcheck, Property 8: Not Ready on an incomplete critical item.
// Validates: Requirements 6.7, 6.9
func TestProperty8_IncompleteCriticalNotReady(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		l := incompleteCriticalGen().Draw(t, "launch")
		if got := ComputeReadiness(l).Status; got != StatusNotReady {
			t.Fatalf("incomplete critical: want Not Ready, got %s", got)
		}
	})
}

// Feature: shipcheck, Property 9: Not Ready on a Blocking_Risk, and the Blocking_Risk definition.
// Validates: Requirements 6.8
func TestProperty9_BlockingRisk(t *testing.T) {
	// Part A: the predicate itself.
	rapid.Check(t, func(t *rapid.T) {
		r := riskGen("r").Draw(t, "risk")
		want := r.Severity == SeverityHigh && r.Status != RiskResolved && r.Status != RiskAccepted
		if IsBlockingRisk(r) != want {
			t.Fatalf("IsBlockingRisk(%+v) = %v, want %v", r, IsBlockingRisk(r), want)
		}
	})
	// Part B: a launch with a blocking risk (and no higher-tier trigger) is Not Ready.
	rapid.Check(t, func(t *rapid.T) {
		l := blockingRiskGen().Draw(t, "launch")
		if got := ComputeReadiness(l).Status; got != StatusNotReady {
			t.Fatalf("blocking risk: want Not Ready, got %s", got)
		}
	})
}

// Feature: shipcheck, Property 10: Conditionally Ready when criticals satisfied
// but a non-critical gap or accepted risk remains.
// Validates: Requirements 6.10, 6.11
func TestProperty10_ConditionallyReady(t *testing.T) {
	// Non-critical gap variant.
	rapid.Check(t, func(t *rapid.T) {
		l := allCompleteGen().Draw(t, "base")
		l.ChecklistItems = append(l.ChecklistItems, ChecklistItem{
			ID: "nc-gap", Title: "later", Category: CategoryDocumentation,
			Priority: PriorityLow, CompletionState: CompletionInProgress, IsCritical: false,
		})
		if got := ComputeReadiness(l).Status; got != StatusConditionallyReady {
			t.Fatalf("non-critical gap: want Conditionally Ready, got %s", got)
		}
	})
	// Accepted-risk variant.
	rapid.Check(t, func(t *rapid.T) {
		l := allCompleteGen().Draw(t, "base")
		l.Risks = append(l.Risks, Risk{
			ID: "acc", Title: "known", Severity: SeverityHigh, Status: RiskAccepted,
		})
		if got := ComputeReadiness(l).Status; got != StatusConditionallyReady {
			t.Fatalf("accepted risk: want Conditionally Ready, got %s", got)
		}
	})
}

// Feature: shipcheck, Property 11: Ready when everything is complete and evidenced.
// Validates: Requirements 6.12
func TestProperty11_Ready(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		l := allCompleteGen().Draw(t, "launch")
		if got := ComputeReadiness(l).Status; got != StatusReady {
			t.Fatalf("all complete: want Ready, got %s", got)
		}
	})
}

// Feature: shipcheck, Property 12: Every assessment carries at least one rule-identified reason.
// Validates: Requirements 7.1
func TestProperty12_HasRuleIdentifiedReason(t *testing.T) {
	knownRule := map[string]bool{
		RuleEmptyLaunch: true, RuleNoOwner: true, RuleCompleteNoEvidence: true,
		RuleIncompleteCritical: true, RuleBlockingRisk: true, RuleNonCriticalGap: true,
		RuleAcceptedRisk: true, RuleAllComplete: true, RuleDefaultReady: true,
	}
	rapid.Check(t, func(t *rapid.T) {
		a := ComputeReadiness(launchGen().Draw(t, "launch"))
		if len(a.Reasons) == 0 {
			t.Fatal("assessment has no reasons")
		}
		for _, r := range a.Reasons {
			if !knownRule[r.RuleID] {
				t.Fatalf("reason has unknown ruleId %q", r.RuleID)
			}
		}
	})
}

// Feature: shipcheck, Property 13: Reason item-id association is correct.
// Validates: Requirements 7.2
// Launch-level rules (empty launch, no owner, all-complete, default-ready)
// carry no item id; item-triggered rules carry an id that exists on the launch.
func TestProperty13_ReasonItemIDAssociation(t *testing.T) {
	launchLevel := map[string]bool{
		RuleEmptyLaunch: true, RuleNoOwner: true, RuleAllComplete: true, RuleDefaultReady: true,
	}
	rapid.Check(t, func(t *rapid.T) {
		l := launchGen().Draw(t, "launch")
		ids := map[string]bool{}
		for _, it := range l.ChecklistItems {
			ids[it.ID] = true
		}
		for _, r := range l.Risks {
			ids[r.ID] = true
		}
		for _, reason := range ComputeReadiness(l).Reasons {
			if launchLevel[reason.RuleID] {
				if reason.ItemID != "" {
					t.Fatalf("launch-level rule %q carries item id %q", reason.RuleID, reason.ItemID)
				}
				continue
			}
			if reason.ItemID == "" {
				t.Fatalf("item-triggered rule %q has no item id", reason.RuleID)
			}
			if !ids[reason.ItemID] {
				t.Fatalf("reason references unknown item id %q", reason.ItemID)
			}
		}
	})
}

// Feature: shipcheck, Property 14: Not Ready reasons correspond one-to-one to their causes.
// Validates: Requirements 7.3
func TestProperty14_NotReadyReasonCounts(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		l := ownedNonEmptyGen().Draw(t, "base")
		// Force Not Ready by guaranteeing at least one incomplete critical.
		l.ChecklistItems = append(l.ChecklistItems, ChecklistItem{
			ID: "nr-crit", Title: "x", Category: CategoryEngineering,
			Priority: PriorityHigh, CompletionState: CompletionBlocked, IsCritical: true,
		})
		a := ComputeReadiness(l)
		if a.Status != StatusNotReady {
			return // generator produced a higher-tier launch; nothing to assert here
		}
		wantIncompleteCrit := 0
		for _, it := range l.ChecklistItems {
			if IsCriticalItem(it) && !itemComplete(it) {
				wantIncompleteCrit++
			}
		}
		wantBlocking := 0
		for _, r := range l.Risks {
			if IsBlockingRisk(r) {
				wantBlocking++
			}
		}
		gotIncompleteCrit, gotBlocking := 0, 0
		for _, reason := range a.Reasons {
			switch reason.RuleID {
			case RuleIncompleteCritical:
				gotIncompleteCrit++
			case RuleBlockingRisk:
				gotBlocking++
			default:
				t.Fatalf("unexpected Not-Ready reason rule %q", reason.RuleID)
			}
		}
		if gotIncompleteCrit != wantIncompleteCrit {
			t.Fatalf("incomplete-critical reasons: got %d want %d", gotIncompleteCrit, wantIncompleteCrit)
		}
		if gotBlocking != wantBlocking {
			t.Fatalf("blocking-risk reasons: got %d want %d", gotBlocking, wantBlocking)
		}
	})
}

// Feature: shipcheck, Property 15: Needs Review reasons cover every triggering condition.
// Validates: Requirements 7.4
func TestProperty15_NeedsReviewReasonCoverage(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		l := launchGen().Draw(t, "launch")
		a := ComputeReadiness(l)
		if a.Status != StatusNeedsReview {
			return
		}
		emptyLaunch := len(l.ChecklistItems) == 0 && len(l.Risks) == 0
		noOwner := l.Owner == ""
		wantCompleteNoEv := 0
		for _, it := range l.ChecklistItems {
			if IsCriticalItem(it) && itemComplete(it) && !hasEvidence(it) {
				wantCompleteNoEv++
			}
		}
		gotEmpty, gotNoOwner, gotCompleteNoEv := 0, 0, 0
		for _, reason := range a.Reasons {
			switch reason.RuleID {
			case RuleEmptyLaunch:
				gotEmpty++
			case RuleNoOwner:
				gotNoOwner++
			case RuleCompleteNoEvidence:
				gotCompleteNoEv++
			default:
				t.Fatalf("unexpected Needs-Review reason rule %q", reason.RuleID)
			}
		}
		if emptyLaunch && gotEmpty != 1 {
			t.Fatalf("empty-launch reason: got %d want 1", gotEmpty)
		}
		if !emptyLaunch && gotEmpty != 0 {
			t.Fatalf("unexpected empty-launch reason")
		}
		if noOwner && gotNoOwner != 1 {
			t.Fatalf("no-owner reason: got %d want 1", gotNoOwner)
		}
		if gotCompleteNoEv != wantCompleteNoEv {
			t.Fatalf("complete-no-evidence reasons: got %d want %d", gotCompleteNoEv, wantCompleteNoEv)
		}
	})
}

// Feature: shipcheck, Property 16: Actionable reasons carry a targeted recommended action.
// Validates: Requirements 7.5
// Only incomplete-critical, blocking-risk, absent-owner, and
// complete-without-evidence reasons must carry a recommended action.
func TestProperty16_ActionableReasonsHaveActions(t *testing.T) {
	actionable := map[string]bool{
		RuleIncompleteCritical: true, RuleBlockingRisk: true,
		RuleNoOwner: true, RuleCompleteNoEvidence: true,
	}
	rapid.Check(t, func(t *rapid.T) {
		a := ComputeReadiness(launchGen().Draw(t, "launch"))
		for _, r := range a.Reasons {
			if actionable[r.RuleID] && r.RecommendedAction == "" {
				t.Fatalf("actionable rule %q missing recommended action", r.RuleID)
			}
		}
	})
}

// Feature: shipcheck, Property 17: Per-category completion counts and classification are correct.
// Validates: Requirements 7.6
func TestProperty17_CategoryCompletion(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		l := launchGen().Draw(t, "launch")
		cats := ComputeCategoryStates(l)
		if len(cats) != len(Categories) {
			t.Fatalf("want %d categories, got %d", len(Categories), len(cats))
		}
		// Independently recompute expected counts.
		type tally struct{ complete, total int }
		exp := map[Category]tally{}
		for _, it := range l.ChecklistItems {
			tl := exp[it.Category]
			tl.total++
			if itemComplete(it) {
				tl.complete++
			}
			exp[it.Category] = tl
		}
		for i, c := range cats {
			if c.Category != Categories[i] {
				t.Fatalf("category order mismatch at %d: %s", i, c.Category)
			}
			want := exp[c.Category]
			if c.CompleteCount != want.complete || c.TotalCount != want.total {
				t.Fatalf("%s counts: got (%d/%d) want (%d/%d)", c.Category, c.CompleteCount, c.TotalCount, want.complete, want.total)
			}
			var wantState CategoryState
			switch {
			case want.total == 0:
				wantState = CategoryEmpty
			case want.complete == want.total:
				wantState = CategoryComplete
			default:
				wantState = CategoryPartial
			}
			if c.State != wantState {
				t.Fatalf("%s state: got %s want %s", c.Category, c.State, wantState)
			}
		}
	})
}
