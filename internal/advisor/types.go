// Package advisor implements the read-only ShipCheck Advisor and its
// provider-agnostic ModelProvider abstraction.
//
// The governing principle: the deterministic Readiness_Engine decides the
// state; the advisor only explains it and recommends next actions. The advisor
// operates on a LaunchSnapshot produced AFTER the engine computes readiness,
// never mutates launch data, and never changes the status. Its structured
// output is schema-validated before display.
package advisor

import "github.com/paulwritescode/shipcheck/internal/domain"

// Action is one of the fixed, focused advisor actions (Req 9.9) — not a
// free-form chatbot.
type Action string

const (
	ActionAnalyze          Action = "analyze"
	ActionFindBlockers     Action = "find_blockers"
	ActionSuggestChecklist Action = "suggest_checklist"
	ActionReviewRisks      Action = "review_risks"
	ActionPrepareReview    Action = "prepare_review"
	ActionExplainReadiness Action = "explain_readiness"
	ActionReanalyze        Action = "reanalyze"
)

// IsValid reports whether a is a known advisor action.
func (a Action) IsValid() bool {
	switch a {
	case ActionAnalyze, ActionFindBlockers, ActionSuggestChecklist, ActionReviewRisks,
		ActionPrepareReview, ActionExplainReadiness, ActionReanalyze:
		return true
	default:
		return false
	}
}

// RecommendationType categorizes an advisor recommendation.
type RecommendationType string

const (
	RecMissingInformation     RecommendationType = "missing_information"
	RecLikelyBlocker          RecommendationType = "likely_blocker"
	RecPossibleContradiction  RecommendationType = "possible_contradiction"
	RecUnownedWork            RecommendationType = "unowned_work"
	RecSuggestedChecklistItem RecommendationType = "suggested_checklist_item"
	RecReviewQuestion         RecommendationType = "review_question"
)

// EntityKind identifies which kind of launch entity a recommendation references.
type EntityKind string

const (
	EntityChecklistItem EntityKind = "checklistItem"
	EntityRisk          EntityKind = "risk"
	EntityEvidence      EntityKind = "evidence"
	EntityLaunch        EntityKind = "launch"
)

// EvidenceRef points at a real launch entity the recommendation is about.
type EvidenceRef struct {
	EntityKind EntityKind `json:"entityKind"`
	EntityID   string     `json:"entityId"`
}

// Recommendation is a single advisor suggestion. It never applies itself; the
// user accepts or dismisses it.
type Recommendation struct {
	Type            RecommendationType `json:"type"`
	Priority        domain.Priority    `json:"priority"`
	Title           string             `json:"title"`
	Reason          string             `json:"reason"`
	SuggestedAction string             `json:"suggested_action"`
	Evidence        []EvidenceRef      `json:"evidence"`
}

// LaunchAnalysis is the advisor's full structured result (Req 9.10).
type LaunchAnalysis struct {
	Recommendations    []Recommendation `json:"recommendations"`
	SuggestedQuestions []string         `json:"suggested_questions"`
}

// LaunchSnapshot is the immutable copy of launch data (including the computed
// assessment) passed to a provider for a single analysis request. It is built
// AFTER the engine computes readiness so the model can never be the authority
// for the status (Req 9.7).
type LaunchSnapshot struct {
	Launch domain.Launch `json:"launch"`
	// ExternalContext holds untrusted, read-only external material (e.g. MCP
	// results). It is advisory only and never feeds the engine.
	ExternalContext []ExternalContextItem `json:"externalContext,omitempty"`
}

// ExternalContextItem is a labeled piece of untrusted external content.
type ExternalContextItem struct {
	Source  string `json:"source"`  // e.g. "github-issue"
	URL     string `json:"url"`     // allowlisted public URL
	Content string `json:"content"` // treated as DATA, never instructions
}

// AdvisorInput is what a ModelProvider receives for one analysis.
type AdvisorInput struct {
	Action       Action         `json:"action"`
	Snapshot     LaunchSnapshot `json:"snapshot"`
	SystemPrompt string         `json:"systemPrompt"`
}
