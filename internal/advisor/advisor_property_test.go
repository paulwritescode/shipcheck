package advisor

import (
	"context"
	"reflect"
	"testing"

	"pgregory.net/rapid"

	"github.com/paulwritescode/shipcheck/internal/domain"
)

// launchGen draws an arbitrary launch (with a computed assessment) for the
// advisor purity/authority properties. IDs are unique within the launch.
func launchGen() *rapid.Generator[domain.Launch] {
	return rapid.Custom(func(t *rapid.T) domain.Launch {
		nItems := rapid.IntRange(0, 5).Draw(t, "nItems")
		items := make([]domain.ChecklistItem, nItems)
		for i := 0; i < nItems; i++ {
			items[i] = domain.ChecklistItem{
				ID:              itemID("i", i),
				Title:           rapid.StringN(1, 16, 16).Draw(t, "title"),
				Category:        rapid.SampledFrom(domain.Categories).Draw(t, "cat"),
				Owner:           rapid.StringN(0, 8, 8).Draw(t, "owner"),
				Priority:        rapid.SampledFrom([]domain.Priority{domain.PriorityLow, domain.PriorityMedium, domain.PriorityHigh}).Draw(t, "pri"),
				CompletionState: rapid.SampledFrom([]domain.CompletionState{domain.CompletionComplete, domain.CompletionInProgress, domain.CompletionBlocked, domain.CompletionIncomplete}).Draw(t, "comp"),
				IsCritical:      rapid.Bool().Draw(t, "crit"),
			}
		}
		nRisks := rapid.IntRange(0, 4).Draw(t, "nRisks")
		risks := make([]domain.Risk, nRisks)
		for i := 0; i < nRisks; i++ {
			risks[i] = domain.Risk{
				ID:         itemID("r", i),
				Title:      rapid.StringN(1, 16, 16).Draw(t, "rtitle"),
				Severity:   rapid.SampledFrom([]domain.Severity{domain.SeverityLow, domain.SeverityMedium, domain.SeverityHigh}).Draw(t, "sev"),
				Likelihood: rapid.SampledFrom([]domain.Likelihood{domain.LikelihoodLow, domain.LikelihoodMedium, domain.LikelihoodHigh}).Draw(t, "lik"),
				Status:     rapid.SampledFrom([]domain.RiskStatus{domain.RiskOpen, domain.RiskMitigating, domain.RiskResolved, domain.RiskAccepted}).Draw(t, "rstatus"),
			}
		}
		owner := ""
		if rapid.Bool().Draw(t, "hasOwner") {
			owner = "owner"
		}
		data := domain.LaunchData{
			ID: "L", Name: rapid.StringN(1, 20, 20).Draw(t, "name"), TargetDate: "2026-10-16",
			Owner: owner, ChecklistItems: items, Risks: risks,
		}
		return domain.Launch{LaunchData: data, Assessment: domain.ComputeReadiness(data)}
	})
}

func itemID(prefix string, i int) string {
	return prefix + string(rune('a'+i%26))
}

// cloneLaunch deep-copies a launch, preserving nil-vs-empty slices so a
// DeepEqual comparison reflects only genuine mutations.
func cloneLaunch(l domain.Launch) domain.Launch {
	out := l // copies scalars + the Assessment value
	if l.ChecklistItems != nil {
		out.ChecklistItems = make([]domain.ChecklistItem, len(l.ChecklistItems))
		for i, it := range l.ChecklistItems {
			if it.Evidence != nil {
				it.Evidence = append([]domain.Evidence(nil), it.Evidence...)
			}
			out.ChecklistItems[i] = it
		}
	}
	if l.Risks != nil {
		out.Risks = make([]domain.Risk, len(l.Risks))
		for i, r := range l.Risks {
			if r.Evidence != nil {
				r.Evidence = append([]domain.Evidence(nil), r.Evidence...)
			}
			out.Risks[i] = r
		}
	}
	if l.Assessment.Reasons != nil {
		out.Assessment.Reasons = append([]domain.ReasonEntry(nil), l.Assessment.Reasons...)
	}
	if l.Assessment.Categories != nil {
		out.Assessment.Categories = append([]domain.CategoryCompletion(nil), l.Assessment.Categories...)
	}
	return out
}

var actionGen = rapid.SampledFrom([]Action{
	ActionAnalyze, ActionFindBlockers, ActionSuggestChecklist, ActionReviewRisks,
	ActionPrepareReview, ActionExplainReadiness, ActionReanalyze,
})

// Feature: shipcheck, Property 31: Advisor never mutates launch state.
// Validates: Requirements 9.8
// For any launch and any action, running the advisor leaves the launch value
// unchanged (the advisor is read-only with respect to launch state).
func TestProperty31_AdvisorNeverMutates(t *testing.T) {
	adv := New(NewFallbackProvider())
	rapid.Check(t, func(t *rapid.T) {
		launch := launchGen().Draw(t, "launch")
		action := actionGen.Draw(t, "action")

		// Capture a deep copy that preserves nil-vs-empty slices, so the
		// comparison reflects only real mutations, not slice normalization.
		before := cloneLaunch(launch)

		_, err := adv.Analyze(context.Background(), launch, action)
		if err != nil {
			t.Fatalf("advisor error: %v", err)
		}

		if !reflect.DeepEqual(launch, before) {
			t.Fatal("advisor mutated the launch")
		}
	})
}

// Feature: shipcheck, Property 32: Deterministic authority — advisor output cannot change the score.
// Validates: Requirements 9.7, 9.8, 15.2
// For any launch and action, the readiness status computed by the engine before
// and after the advisor runs is identical: the advisor is never the source of
// the status.
func TestProperty32_DeterministicAuthority(t *testing.T) {
	adv := New(NewFallbackProvider())
	rapid.Check(t, func(t *rapid.T) {
		launch := launchGen().Draw(t, "launch")
		action := actionGen.Draw(t, "action")

		statusBefore := domain.ComputeReadiness(launch.LaunchData).Status
		if _, err := adv.Analyze(context.Background(), launch, action); err != nil {
			t.Fatalf("advisor error: %v", err)
		}
		statusAfter := domain.ComputeReadiness(launch.LaunchData).Status

		if statusBefore != statusAfter {
			t.Fatalf("advisor changed the engine status: %s -> %s", statusBefore, statusAfter)
		}
		if statusAfter != launch.Assessment.Status {
			t.Fatalf("engine status %s drifted from stored assessment %s", statusAfter, launch.Assessment.Status)
		}
	})
}
