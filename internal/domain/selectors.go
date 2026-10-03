package domain

import "sort"

// This file holds the pure dashboard/presentation selectors and small edit
// helpers. Like the engine, every function here is a pure function of its
// inputs and never mutates them, so each is independently property-testable.

// Blocker is a single critical blocker surfaced on the dashboard: either an
// incomplete critical checklist item or a blocking risk.
type Blocker struct {
	Kind   string // "item" or "risk"
	ItemID string // the ChecklistItem.ID or Risk.ID
	Title  string
}

// CriticalBlockers returns the launch's critical blockers: every incomplete
// critical checklist item plus every blocking risk (Req 8.3). Items come
// first (ID-sorted), then risks (ID-sorted), so the result is deterministic.
func CriticalBlockers(l LaunchData) []Blocker {
	var out []Blocker
	for _, it := range sortedItems(l.ChecklistItems) {
		if IsCriticalItem(it) && !itemComplete(it) {
			out = append(out, Blocker{Kind: "item", ItemID: it.ID, Title: it.Title})
		}
	}
	for _, r := range sortedRisks(l.Risks) {
		if IsBlockingRisk(r) {
			out = append(out, Blocker{Kind: "risk", ItemID: r.ID, Title: r.Title})
		}
	}
	return out
}

// severityRank orders severities from most to least severe for display.
func severityRank(s Severity) int {
	switch s {
	case SeverityHigh:
		return 0
	case SeverityMedium:
		return 1
	default: // Low or unknown
		return 2
	}
}

// RisksBySeverity returns a copy of risks ordered so that higher-severity risks
// come first (Req 8.4). The sort is stable, so risks of equal severity keep
// their input order. The input slice is not mutated.
func RisksBySeverity(risks []Risk) []Risk {
	out := make([]Risk, len(risks))
	copy(out, risks)
	sort.SliceStable(out, func(i, j int) bool {
		return severityRank(out[i].Severity) < severityRank(out[j].Severity)
	})
	return out
}

// IncompleteItems returns the items whose completion state is exactly
// Incomplete (Req 3.11). Order is preserved; the input is not mutated.
func IncompleteItems(items []ChecklistItem) []ChecklistItem {
	out := make([]ChecklistItem, 0, len(items))
	for _, it := range items {
		if it.CompletionState == CompletionIncomplete {
			out = append(out, it)
		}
	}
	return out
}

// HighPriorityItems returns the items whose priority is High (Req 3.12).
func HighPriorityItems(items []ChecklistItem) []ChecklistItem {
	out := make([]ChecklistItem, 0, len(items))
	for _, it := range items {
		if it.Priority == PriorityHigh {
			out = append(out, it)
		}
	}
	return out
}

// GroupByCategory partitions items by category (Req 3.13). The returned map
// has one entry per category that has at least one item; within each group the
// items keep their input order. The input is not mutated.
func GroupByCategory(items []ChecklistItem) map[Category][]ChecklistItem {
	out := make(map[Category][]ChecklistItem)
	for _, it := range items {
		out[it.Category] = append(out[it.Category], it)
	}
	return out
}

// HasHighEmphasis reports whether a risk warrants high-emphasis display:
// High severity OR High likelihood (Req 4.6).
func HasHighEmphasis(r Risk) bool {
	return r.Severity == SeverityHigh || r.Likelihood == LikelihoodHigh
}
