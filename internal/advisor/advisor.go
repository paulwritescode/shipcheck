package advisor

import (
	"context"

	"github.com/paulwritescode/shipcheck/internal/domain"
	"github.com/paulwritescode/shipcheck/internal/security"
)

// Advisor orchestrates a single read-only analysis: it builds a snapshot from a
// launch that already carries its engine-computed assessment, runs the chosen
// ModelProvider with an injection-hardened system prompt, and validates the
// result. It never mutates the launch and never changes the status — the
// deterministic engine remains the sole authority (Req 9.7, 9.8).
type Advisor struct {
	provider ModelProvider
}

// New returns an advisor backed by the given provider. Pass SelectProvider()
// in production; pass a stub/fallback in tests.
func New(p ModelProvider) *Advisor { return &Advisor{provider: p} }

// Result is what the advisor returns to the API layer: the validated analysis
// plus any discarded recommendations and the provider that produced it (for
// disclosure, Req 21.3).
type Result struct {
	Analysis  LaunchAnalysis            `json:"analysis"`
	Discarded []DiscardedRecommendation `json:"discarded,omitempty"`
	Provider  string                    `json:"provider"`
}

// Analyze runs one focused advisor action against the launch. The launch is
// passed by value and never written back, so it cannot be mutated here.
func (a *Advisor) Analyze(ctx context.Context, launch domain.Launch, action Action) (Result, error) {
	if !action.IsValid() {
		action = ActionAnalyze
	}
	snapshot := BuildSnapshot(launch)
	in := AdvisorInput{
		Action:       action,
		Snapshot:     snapshot,
		SystemPrompt: SystemPrompt(action),
	}

	raw, err := a.provider.AnalyzeLaunch(ctx, in)
	if err != nil {
		return Result{}, err
	}

	analysis, discarded := ValidateAnalysis(raw, launch)
	analysis = filterForAction(action, analysis)

	return Result{Analysis: analysis, Discarded: discarded, Provider: a.provider.Name()}, nil
}

// BuildSnapshot builds the immutable snapshot the advisor reasons over. It
// redacts prohibited data (private notes, private evidence) before anything can
// reach a model provider (Req 19.1, 19.2), and the redacted copy means the
// advisor path can never reach back into the caller's launch value.
func BuildSnapshot(launch domain.Launch) LaunchSnapshot {
	return LaunchSnapshot{Launch: security.RedactForModel(launch)}
}

// filterForAction narrows the analysis to what a focused action should show.
// The underlying provider may return a full analysis; the action shapes which
// recommendations and questions surface (Req 9.9). Analyze/Explain/Prepare keep
// everything; the others filter to relevant recommendation types.
func filterForAction(action Action, a LaunchAnalysis) LaunchAnalysis {
	keep := func(pred func(RecommendationType) bool) LaunchAnalysis {
		out := LaunchAnalysis{SuggestedQuestions: a.SuggestedQuestions}
		for _, r := range a.Recommendations {
			if pred(r.Type) {
				out.Recommendations = append(out.Recommendations, r)
			}
		}
		return out
	}

	switch action {
	case ActionFindBlockers:
		return keep(func(t RecommendationType) bool { return t == RecLikelyBlocker })
	case ActionSuggestChecklist:
		return keep(func(t RecommendationType) bool {
			return t == RecSuggestedChecklistItem || t == RecMissingInformation
		})
	case ActionReviewRisks:
		return keep(func(t RecommendationType) bool {
			return t == RecLikelyBlocker || t == RecUnownedWork
		})
	default:
		// analyze, prepare_review, explain_readiness, reanalyze: full analysis.
		return a
	}
}
