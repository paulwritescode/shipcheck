package advisor

import (
	"testing"

	"pgregory.net/rapid"

	"github.com/paulwritescode/shipcheck/internal/domain"
)

// A small launch with known entity ids, used to decide which generated
// references are "real" vs "bogus".
func fixtureLaunch() domain.Launch {
	data := domain.LaunchData{
		ID: "L1", Name: "Fixture", TargetDate: "2026-10-16", Owner: "Alice",
		ChecklistItems: []domain.ChecklistItem{
			{ID: "I1", Title: "a", Category: domain.CategoryEngineering, Priority: domain.PriorityLow, CompletionState: domain.CompletionComplete, Evidence: []domain.Evidence{{ID: "E1"}}},
			{ID: "I2", Title: "b", Category: domain.CategoryDocumentation, Priority: domain.PriorityLow, CompletionState: domain.CompletionIncomplete},
		},
		Risks: []domain.Risk{{ID: "R1", Title: "r", Severity: domain.SeverityLow, Likelihood: domain.LikelihoodLow, Status: domain.RiskOpen}},
	}
	return domain.Launch{LaunchData: data, Assessment: domain.ComputeReadiness(data)}
}

// realEntityKeys returns the set of (kind,id) that exist on the fixture.
func realEntityKeys() map[entityKey]bool { return entityIndex(fixtureLaunch()) }

var (
	recTypeGen = rapid.SampledFrom([]RecommendationType{
		RecMissingInformation, RecLikelyBlocker, RecPossibleContradiction,
		RecUnownedWork, RecSuggestedChecklistItem, RecReviewQuestion,
		"bogus_type", // sometimes invalid
	})
	entityKindGen = rapid.SampledFrom([]EntityKind{
		EntityChecklistItem, EntityRisk, EntityEvidence, EntityLaunch, "bogus_kind",
	})
	priorityGen = rapid.SampledFrom([]domain.Priority{
		domain.PriorityLow, domain.PriorityMedium, domain.PriorityHigh, "bogus_priority",
	})
	entityIDGen = rapid.SampledFrom([]string{
		"L1", "I1", "I2", "R1", "E1", // real
		"NOPE", "ghost", "", // bogus
	})
)

// analysisGen draws arbitrary advisor output, including malformed and
// bogus-reference recommendations.
func analysisGen() *rapid.Generator[LaunchAnalysis] {
	return rapid.Custom(func(t *rapid.T) LaunchAnalysis {
		n := rapid.IntRange(0, 6).Draw(t, "nRecs")
		recs := make([]Recommendation, n)
		for i := 0; i < n; i++ {
			nEv := rapid.IntRange(0, 3).Draw(t, "nEv")
			ev := make([]EvidenceRef, nEv)
			for j := 0; j < nEv; j++ {
				ev[j] = EvidenceRef{
					EntityKind: entityKindGen.Draw(t, "kind"),
					EntityID:   entityIDGen.Draw(t, "eid"),
				}
			}
			recs[i] = Recommendation{
				Type:            recTypeGen.Draw(t, "type"),
				Priority:        priorityGen.Draw(t, "pri"),
				Title:           rapid.StringN(0, 20, 20).Draw(t, "title"),
				Reason:          rapid.StringN(0, 20, 20).Draw(t, "reason"),
				SuggestedAction: rapid.StringN(0, 20, 20).Draw(t, "action"),
				Evidence:        ev,
			}
		}
		nq := rapid.IntRange(0, 4).Draw(t, "nq")
		qs := make([]string, nq)
		for i := 0; i < nq; i++ {
			qs[i] = rapid.StringN(0, 20, 20).Draw(t, "q")
		}
		return LaunchAnalysis{Recommendations: recs, SuggestedQuestions: qs}
	})
}

// Feature: shipcheck, Property 29: Advisor output is structurally valid after validation.
// Validates: Requirements 9.10, 9.11
// Every recommendation retained by the validator has a known type, a valid
// priority, a non-empty title and reason, and only known evidence kinds.
func TestProperty29_AdvisorOutputStructurallyValid(t *testing.T) {
	l := fixtureLaunch()
	rapid.Check(t, func(t *rapid.T) {
		raw := analysisGen().Draw(t, "analysis")
		clean, _ := ValidateAnalysis(raw, l)
		for _, rec := range clean.Recommendations {
			if !validRecommendationType(rec.Type) {
				t.Fatalf("retained rec has invalid type %q", rec.Type)
			}
			if !rec.Priority.IsValid() {
				t.Fatalf("retained rec has invalid priority %q", rec.Priority)
			}
			if rec.Title == "" || rec.Reason == "" {
				t.Fatalf("retained rec has empty title/reason")
			}
			for _, e := range rec.Evidence {
				if !validEntityKind(e.EntityKind) {
					t.Fatalf("retained rec has invalid entity kind %q", e.EntityKind)
				}
			}
		}
		// Suggested questions retained are all non-empty.
		for _, q := range clean.SuggestedQuestions {
			if q == "" {
				t.Fatal("retained an empty suggested question")
			}
		}
	})
}

// Feature: shipcheck, Property 30: Evidence-reference integrity.
// Validates: Requirements 9.12
// Every retained recommendation references only entities that exist on the
// launch; any recommendation referencing a nonexistent entity is discarded.
func TestProperty30_EvidenceReferenceIntegrity(t *testing.T) {
	l := fixtureLaunch()
	real := realEntityKeys()
	rapid.Check(t, func(t *rapid.T) {
		raw := analysisGen().Draw(t, "analysis")
		clean, discarded := ValidateAnalysis(raw, l)

		// Retained recs reference only real entities.
		for _, rec := range clean.Recommendations {
			for _, e := range rec.Evidence {
				if !real[entityKey{e.EntityKind, e.EntityID}] {
					t.Fatalf("retained rec references nonexistent %s %q", e.EntityKind, e.EntityID)
				}
			}
		}

		// Conservation: retained + discarded == total input.
		if len(clean.Recommendations)+len(discarded) != len(raw.Recommendations) {
			t.Fatalf("rec accounting off: %d retained + %d discarded != %d input",
				len(clean.Recommendations), len(discarded), len(raw.Recommendations))
		}

		// Any input rec that references a nonexistent entity (with otherwise
		// valid fields) must have been discarded — none survive in clean.
		for _, rec := range clean.Recommendations {
			if recProblem(rec, real) != "" {
				t.Fatalf("retained rec still has a problem: %s", recProblem(rec, real))
			}
		}
	})
}
