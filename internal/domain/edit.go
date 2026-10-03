package domain

// Pure single-field edit helpers. Each returns a copy of the entity with one
// field changed and every other field preserved (Req 3.9, 4.7). The original
// is never mutated. Nested slices (Evidence) are copied so the result shares
// no mutable state with the input.

// cloneEvidence returns an independent copy of an evidence slice.
func cloneEvidence(ev []Evidence) []Evidence {
	if ev == nil {
		return nil
	}
	out := make([]Evidence, len(ev))
	copy(out, ev)
	return out
}

// WithItemOwner returns item with its Owner set to owner.
func WithItemOwner(item ChecklistItem, owner string) ChecklistItem {
	item.Evidence = cloneEvidence(item.Evidence)
	item.Owner = owner
	return item
}

// WithItemPriority returns item with its Priority set to p.
func WithItemPriority(item ChecklistItem, p Priority) ChecklistItem {
	item.Evidence = cloneEvidence(item.Evidence)
	item.Priority = p
	return item
}

// WithItemCompletion returns item with its CompletionState set to s.
func WithItemCompletion(item ChecklistItem, s CompletionState) ChecklistItem {
	item.Evidence = cloneEvidence(item.Evidence)
	item.CompletionState = s
	return item
}

// WithItemCritical returns item with its IsCritical flag set to critical.
func WithItemCritical(item ChecklistItem, critical bool) ChecklistItem {
	item.Evidence = cloneEvidence(item.Evidence)
	item.IsCritical = critical
	return item
}

// WithRiskStatus returns risk with its Status set to s.
func WithRiskStatus(risk Risk, s RiskStatus) Risk {
	risk.Evidence = cloneEvidence(risk.Evidence)
	risk.Status = s
	return risk
}

// WithRiskOwner returns risk with its Owner set to owner.
func WithRiskOwner(risk Risk, owner string) Risk {
	risk.Evidence = cloneEvidence(risk.Evidence)
	risk.Owner = owner
	return risk
}
