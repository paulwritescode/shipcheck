package domain

// ReasonEntry explains one condition that contributed to a readiness status.
// A reason tied to a specific item/risk carries its ItemID; a launch-level
// reason leaves ItemID empty. Actionable reasons carry a RecommendedAction.
type ReasonEntry struct {
	RuleID            string `json:"ruleId"`                      // the matched Requirement-6 rule
	Message           string `json:"message"`                     // human-readable explanation
	ItemID            string `json:"itemId,omitempty"`            // set when tied to a specific entity
	RecommendedAction string `json:"recommendedAction,omitempty"` // set for actionable reasons
}

// CategoryCompletion is the per-category completion summary.
type CategoryCompletion struct {
	Category      Category      `json:"category"`
	CompleteCount int           `json:"completeCount"`
	TotalCount    int           `json:"totalCount"`
	State         CategoryState `json:"state"` // complete | partial | empty
}

// ReadinessAssessment is the full result the readiness engine produces: the
// status, the ordered reasons that justify it, and the per-category breakdown.
type ReadinessAssessment struct {
	Status     ReadinessStatus      `json:"status"`
	Reasons    []ReasonEntry        `json:"reasons"`
	Categories []CategoryCompletion `json:"categories"`
}
