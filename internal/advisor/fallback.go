package advisor

import (
	"context"
	"sort"

	"github.com/paulwritescode/shipcheck/internal/domain"
)

// FallbackProvider is the free, deterministic, rule-based advisor and the
// default the public demo runs on. It derives recommendations directly from the
// launch data and its computed assessment — incomplete critical items, blocking
// risks, unowned work, completion claims lacking evidence, and per-category
// gaps — and references only real launch entities. It requires no external
// model and no key, and its output is schema-valid like any other provider, so
// the advisor properties (29-34) hold for it.
type FallbackProvider struct{}

// NewFallbackProvider returns the deterministic fallback provider.
func NewFallbackProvider() *FallbackProvider { return &FallbackProvider{} }

func (p *FallbackProvider) Name() string { return "fallback" }

var _ ModelProvider = (*FallbackProvider)(nil)

// AnalyzeLaunch produces recommendations from the snapshot. It is pure and
// deterministic: the same snapshot yields the same analysis, and it never
// mutates the input.
func (p *FallbackProvider) AnalyzeLaunch(_ context.Context, in AdvisorInput) (LaunchAnalysis, error) {
	l := in.Snapshot.Launch
	items := sortedItems(l.ChecklistItems)
	risks := sortedRisks(l.Risks)

	var recs []Recommendation

	// Unowned launch (launch-level).
	if l.Owner == "" {
		recs = append(recs, Recommendation{
			Type: RecUnownedWork, Priority: domain.PriorityHigh,
			Title:           "Assign a launch owner",
			Reason:          "The launch has no owner, so accountability for the release is unclear.",
			SuggestedAction: "Set a launch owner on the launch details.",
			Evidence:        []EvidenceRef{{EntityKind: EntityLaunch, EntityID: l.ID}},
		})
	}

	for _, it := range items {
		critical := it.IsCritical
		complete := it.CompletionState == domain.CompletionComplete

		// Incomplete critical item -> likely blocker.
		if critical && !complete {
			recs = append(recs, Recommendation{
				Type: RecLikelyBlocker, Priority: domain.PriorityHigh,
				Title:           "Resolve critical item: " + it.Title,
				Reason:          "Critical item \"" + it.Title + "\" is not complete and blocks readiness.",
				SuggestedAction: "Complete \"" + it.Title + "\" and attach supporting evidence.",
				Evidence:        []EvidenceRef{{EntityKind: EntityChecklistItem, EntityID: it.ID}},
			})
		}

		// Complete critical item without evidence -> contradiction.
		if critical && complete && len(it.Evidence) == 0 {
			recs = append(recs, Recommendation{
				Type: RecPossibleContradiction, Priority: domain.PriorityHigh,
				Title:           "Missing evidence for completed critical item: " + it.Title,
				Reason:          "\"" + it.Title + "\" is marked complete but has no supporting evidence.",
				SuggestedAction: "Attach evidence (a PR, test report, or doc) to \"" + it.Title + "\".",
				Evidence:        []EvidenceRef{{EntityKind: EntityChecklistItem, EntityID: it.ID}},
			})
		}

		// Unowned incomplete item -> unowned work.
		if it.Owner == "" && !complete {
			recs = append(recs, Recommendation{
				Type: RecUnownedWork, Priority: domain.PriorityMedium,
				Title:           "Assign an owner: " + it.Title,
				Reason:          "Incomplete item \"" + it.Title + "\" has no owner.",
				SuggestedAction: "Assign an owner to \"" + it.Title + "\".",
				Evidence:        []EvidenceRef{{EntityKind: EntityChecklistItem, EntityID: it.ID}},
			})
		}
	}

	for _, r := range risks {
		blocking := r.Severity == domain.SeverityHigh && r.Status != domain.RiskResolved && r.Status != domain.RiskAccepted
		if blocking {
			recs = append(recs, Recommendation{
				Type: RecLikelyBlocker, Priority: domain.PriorityHigh,
				Title:           "Resolve high-severity risk: " + r.Title,
				Reason:          "High-severity risk \"" + r.Title + "\" is unresolved and blocks readiness.",
				SuggestedAction: "Resolve or formally accept \"" + r.Title + "\".",
				Evidence:        []EvidenceRef{{EntityKind: EntityRisk, EntityID: r.ID}},
			})
		}
		// Unowned high/medium risk -> unowned work.
		if r.Owner == "" && (r.Severity == domain.SeverityHigh || r.Severity == domain.SeverityMedium) {
			recs = append(recs, Recommendation{
				Type: RecUnownedWork, Priority: domain.PriorityMedium,
				Title:           "Assign an owner to risk: " + r.Title,
				Reason:          "Risk \"" + r.Title + "\" has no owner.",
				SuggestedAction: "Assign an owner to risk \"" + r.Title + "\".",
				Evidence:        []EvidenceRef{{EntityKind: EntityRisk, EntityID: r.ID}},
			})
		}
	}

	// Empty-category gaps -> missing information (launch-level).
	for _, c := range l.Assessment.Categories {
		if c.TotalCount == 0 {
			recs = append(recs, Recommendation{
				Type: RecMissingInformation, Priority: domain.PriorityLow,
				Title:           "No items in " + string(c.Category),
				Reason:          "The " + string(c.Category) + " category has no checklist items.",
				SuggestedAction: "Add " + string(c.Category) + " items if this category is relevant.",
				Evidence:        []EvidenceRef{{EntityKind: EntityLaunch, EntityID: l.ID}},
			})
		}
	}

	return LaunchAnalysis{
		Recommendations:    recs,
		SuggestedQuestions: suggestedQuestions(l),
	}, nil
}

// suggestedQuestions returns deterministic launch-review questions tailored to
// the launch's current state.
func suggestedQuestions(l domain.Launch) []string {
	qs := []string{
		"What is the go/no-go decision owner for this launch?",
		"Is there a tested rollback plan for every risky change?",
	}
	switch l.Assessment.Status {
	case domain.StatusNotReady:
		qs = append(qs, "Which blocker, if resolved, most changes the readiness picture?")
	case domain.StatusConditionallyReady:
		qs = append(qs, "Are the remaining non-critical gaps acceptable to ship with?")
	case domain.StatusNeedsReview:
		qs = append(qs, "What minimum information is needed to make a reliable assessment?")
	}
	return qs
}

func sortedItems(items []domain.ChecklistItem) []domain.ChecklistItem {
	out := make([]domain.ChecklistItem, len(items))
	copy(out, items)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func sortedRisks(risks []domain.Risk) []domain.Risk {
	out := make([]domain.Risk, len(risks))
	copy(out, risks)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
