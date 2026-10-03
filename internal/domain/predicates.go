package domain

// This file holds the small, pure scoring predicates the readiness engine is
// built from. Each is a total function with no side effects, so each is
// independently property-testable.

// IsCompleteForScoring reports whether a completion state counts as complete
// for readiness scoring. Only CompletionComplete counts; in progress, blocked,
// and incomplete all count as "not complete" (Req 6.2).
func IsCompleteForScoring(state CompletionState) bool {
	return state == CompletionComplete
}

// IsCriticalItem reports whether a checklist item is a Critical_Item.
func IsCriticalItem(item ChecklistItem) bool {
	return item.IsCritical
}

// IsBlockingRisk reports whether a risk blocks the launch: High severity with a
// status that is neither Resolved nor Accepted (glossary + Req 6.8).
func IsBlockingRisk(risk Risk) bool {
	return risk.Severity == SeverityHigh &&
		risk.Status != RiskResolved &&
		risk.Status != RiskAccepted
}

// itemComplete reports whether a checklist item is complete for scoring.
func itemComplete(item ChecklistItem) bool {
	return IsCompleteForScoring(item.CompletionState)
}

// hasEvidence reports whether a checklist item carries at least one evidence
// entry.
func hasEvidence(item ChecklistItem) bool {
	return len(item.Evidence) > 0
}

// hasOwner reports whether the launch has a (non-empty) owner.
func hasOwner(l LaunchData) bool {
	return l.Owner != ""
}
