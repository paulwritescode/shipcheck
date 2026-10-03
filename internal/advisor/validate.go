package advisor

import (
	"fmt"

	"github.com/paulwritescode/shipcheck/internal/domain"
)

// DiscardedRecommendation records why a recommendation was dropped during
// validation (Req 9.12), so the backend can log/surface the issue without
// showing an invalid recommendation to the user.
type DiscardedRecommendation struct {
	Index  int    `json:"index"`
	Reason string `json:"reason"`
}

// ValidateAnalysis checks an advisor result against the schema and the launch,
// returning a cleaned analysis plus the list of discarded recommendations
// (Req 9.11, 9.12).
//
// A recommendation is DISCARDED when:
//   - its type, priority, or any evidence entityKind is not a known value, or
//   - its title or reason is empty, or
//   - any evidence reference points at a launch entity that does not exist.
//
// Retained recommendations are therefore always structurally valid and
// reference only real launch entities (Properties 29 and 30). Suggested
// questions are passed through unchanged (dropping empty strings).
func ValidateAnalysis(a LaunchAnalysis, l domain.Launch) (LaunchAnalysis, []DiscardedRecommendation) {
	exists := entityIndex(l)

	clean := LaunchAnalysis{}
	var discarded []DiscardedRecommendation

	for i, rec := range a.Recommendations {
		if reason := recProblem(rec, exists); reason != "" {
			discarded = append(discarded, DiscardedRecommendation{Index: i, Reason: reason})
			continue
		}
		clean.Recommendations = append(clean.Recommendations, rec)
	}

	for _, q := range a.SuggestedQuestions {
		if q != "" {
			clean.SuggestedQuestions = append(clean.SuggestedQuestions, q)
		}
	}

	return clean, discarded
}

// recProblem returns a non-empty reason when the recommendation is invalid, or
// "" when it is valid and references only real entities.
func recProblem(rec Recommendation, exists map[entityKey]bool) string {
	if !validRecommendationType(rec.Type) {
		return fmt.Sprintf("unknown recommendation type %q", rec.Type)
	}
	if !rec.Priority.IsValid() {
		return fmt.Sprintf("invalid priority %q", rec.Priority)
	}
	if rec.Title == "" {
		return "empty title"
	}
	if rec.Reason == "" {
		return "empty reason"
	}
	for _, e := range rec.Evidence {
		if !validEntityKind(e.EntityKind) {
			return fmt.Sprintf("unknown entity kind %q", e.EntityKind)
		}
		if !exists[entityKey{e.EntityKind, e.EntityID}] {
			return fmt.Sprintf("references nonexistent %s %q", e.EntityKind, e.EntityID)
		}
	}
	return ""
}

// entityKey identifies a specific launch entity for existence checks.
type entityKey struct {
	kind EntityKind
	id   string
}

// entityIndex builds the set of entities that exist on the launch: the launch
// itself, its checklist items, its risks, and their evidence.
func entityIndex(l domain.Launch) map[entityKey]bool {
	m := map[entityKey]bool{
		{EntityLaunch, l.ID}: true,
	}
	for _, it := range l.ChecklistItems {
		m[entityKey{EntityChecklistItem, it.ID}] = true
		for _, e := range it.Evidence {
			m[entityKey{EntityEvidence, e.ID}] = true
		}
	}
	for _, r := range l.Risks {
		m[entityKey{EntityRisk, r.ID}] = true
		for _, e := range r.Evidence {
			m[entityKey{EntityEvidence, e.ID}] = true
		}
	}
	return m
}

func validRecommendationType(t RecommendationType) bool {
	switch t {
	case RecMissingInformation, RecLikelyBlocker, RecPossibleContradiction,
		RecUnownedWork, RecSuggestedChecklistItem, RecReviewQuestion:
		return true
	default:
		return false
	}
}

func validEntityKind(k EntityKind) bool {
	switch k {
	case EntityChecklistItem, EntityRisk, EntityEvidence, EntityLaunch:
		return true
	default:
		return false
	}
}
