package domain

import (
	"fmt"

	"pgregory.net/rapid"
)

// Custom rapid generators that produce well-formed LaunchData, plus
// tier-targeting generators that force the engine into a specific readiness
// tier. These back the property-based suite.
//
// Generated checklist-item and risk IDs are unique within a launch (derived
// from their index) because the engine keys reasons by ID and its determinism
// guarantee assumes distinct identifiers.

var (
	categoryGen   = rapid.SampledFrom(Categories)
	priorityGen   = rapid.SampledFrom([]Priority{PriorityLow, PriorityMedium, PriorityHigh})
	severityGen   = rapid.SampledFrom([]Severity{SeverityLow, SeverityMedium, SeverityHigh})
	likelihoodGen = rapid.SampledFrom([]Likelihood{LikelihoodLow, LikelihoodMedium, LikelihoodHigh})
	riskStatusGen = rapid.SampledFrom([]RiskStatus{RiskOpen, RiskMitigating, RiskResolved, RiskAccepted})
	completionGen = rapid.SampledFrom([]CompletionState{
		CompletionComplete, CompletionInProgress, CompletionBlocked, CompletionIncomplete,
	})
)

// evidenceSliceGen draws 0..3 evidence entries with unique ids.
func evidenceSliceGen(itemID string) *rapid.Generator[[]Evidence] {
	return rapid.Custom(func(t *rapid.T) []Evidence {
		n := rapid.IntRange(0, 3).Draw(t, "nEvidence")
		out := make([]Evidence, n)
		for i := 0; i < n; i++ {
			out[i] = Evidence{
				ID:        fmt.Sprintf("%s-e%d", itemID, i),
				Note:      rapid.StringN(0, 20, 20).Draw(t, "evNote"),
				IsPrivate: rapid.Bool().Draw(t, "evPrivate"),
			}
		}
		return out
	})
}

// checklistItemGen draws a single checklist item with the given id.
func checklistItemGen(id string) *rapid.Generator[ChecklistItem] {
	return rapid.Custom(func(t *rapid.T) ChecklistItem {
		return ChecklistItem{
			ID:              id,
			Title:           rapid.StringN(1, 24, 24).Draw(t, "title"),
			Category:        categoryGen.Draw(t, "category"),
			Owner:           rapid.StringN(0, 12, 12).Draw(t, "owner"),
			Priority:        priorityGen.Draw(t, "priority"),
			CompletionState: completionGen.Draw(t, "completion"),
			IsCritical:      rapid.Bool().Draw(t, "critical"),
			Evidence:        evidenceSliceGen(id).Draw(t, "evidence"),
		}
	})
}

// riskGen draws a single risk with the given id.
func riskGen(id string) *rapid.Generator[Risk] {
	return rapid.Custom(func(t *rapid.T) Risk {
		return Risk{
			ID:         id,
			Title:      rapid.StringN(1, 24, 24).Draw(t, "rtitle"),
			Severity:   severityGen.Draw(t, "severity"),
			Likelihood: likelihoodGen.Draw(t, "likelihood"),
			Owner:      rapid.StringN(0, 12, 12).Draw(t, "rowner"),
			Status:     riskStatusGen.Draw(t, "rstatus"),
		}
	})
}

// launchGen produces an arbitrary well-formed LaunchData spanning all tiers.
func launchGen() *rapid.Generator[LaunchData] {
	return rapid.Custom(func(t *rapid.T) LaunchData {
		nItems := rapid.IntRange(0, 6).Draw(t, "nItems")
		items := make([]ChecklistItem, nItems)
		for i := 0; i < nItems; i++ {
			items[i] = checklistItemGen(fmt.Sprintf("i%d", i)).Draw(t, fmt.Sprintf("item%d", i))
		}
		nRisks := rapid.IntRange(0, 4).Draw(t, "nRisks")
		risks := make([]Risk, nRisks)
		for i := 0; i < nRisks; i++ {
			risks[i] = riskGen(fmt.Sprintf("r%d", i)).Draw(t, fmt.Sprintf("risk%d", i))
		}
		// Owner present most of the time, sometimes absent.
		owner := ""
		if rapid.Bool().Draw(t, "hasOwner") {
			owner = "owner-" + rapid.StringN(1, 8, 8).Draw(t, "ownerName")
		}
		return LaunchData{
			ID:             "launch",
			Name:           rapid.StringN(1, 40, 40).Draw(t, "name"),
			TargetDate:     "2026-10-16",
			Owner:          owner,
			ChecklistItems: items,
			Risks:          risks,
		}
	})
}

// ---- Tier-targeting generators ----

// emptyLaunchGen: zero items and zero risks (Needs Review via 6.4).
func emptyLaunchGen() *rapid.Generator[LaunchData] {
	return rapid.Custom(func(t *rapid.T) LaunchData {
		owner := ""
		if rapid.Bool().Draw(t, "hasOwner") {
			owner = "owner"
		}
		return LaunchData{ID: "launch", Name: "n", TargetDate: "2026-10-16", Owner: owner}
	})
}

// ownedNonEmptyGen produces a launch that always has an owner and at least one
// item or risk, so the empty-launch and no-owner Needs-Review rules never fire.
// Used as the base for lower-tier targeting generators.
func ownedNonEmptyGen() *rapid.Generator[LaunchData] {
	return rapid.Custom(func(t *rapid.T) LaunchData {
		l := launchGen().Draw(t, "base")
		l.Owner = "owner"
		if len(l.ChecklistItems) == 0 && len(l.Risks) == 0 {
			l.ChecklistItems = []ChecklistItem{{
				ID: "i0", Title: "seed", Category: CategoryEngineering,
				Priority: PriorityLow, CompletionState: CompletionComplete,
			}}
		}
		// Remove any complete-critical-without-evidence contradiction so this
		// base never lands in Needs Review: give every complete critical item
		// evidence.
		for idx := range l.ChecklistItems {
			it := &l.ChecklistItems[idx]
			if it.IsCritical && itemComplete(*it) && !hasEvidence(*it) {
				it.Evidence = []Evidence{{ID: it.ID + "-e", Note: "ok"}}
			}
		}
		return l
	})
}

// incompleteCriticalGen guarantees at least one incomplete critical item and no
// Needs-Review trigger (Not Ready via 6.7).
func incompleteCriticalGen() *rapid.Generator[LaunchData] {
	return rapid.Custom(func(t *rapid.T) LaunchData {
		l := ownedNonEmptyGen().Draw(t, "base")
		l.ChecklistItems = append(l.ChecklistItems, ChecklistItem{
			ID: "crit-x", Title: "blocker", Category: CategoryEngineering,
			Priority: PriorityHigh, CompletionState: CompletionBlocked, IsCritical: true,
		})
		return l
	})
}

// blockingRiskGen guarantees at least one blocking risk and no Needs-Review or
// incomplete-critical trigger (Not Ready via 6.8). All items are complete; any
// complete critical item gets evidence.
func blockingRiskGen() *rapid.Generator[LaunchData] {
	return rapid.Custom(func(t *rapid.T) LaunchData {
		l := allCompleteGen().Draw(t, "base")
		l.Risks = append(l.Risks, Risk{
			ID: "block-x", Title: "boom", Severity: SeverityHigh,
			Likelihood: LikelihoodHigh, Status: RiskOpen,
		})
		return l
	})
}

// completeCriticalNoEvidenceGen guarantees a complete critical item with no
// evidence (Needs Review via 6.6), owner present.
func completeCriticalNoEvidenceGen() *rapid.Generator[LaunchData] {
	return rapid.Custom(func(t *rapid.T) LaunchData {
		l := ownedNonEmptyGen().Draw(t, "base")
		l.ChecklistItems = append(l.ChecklistItems, ChecklistItem{
			ID: "crit-ne", Title: "claimed", Category: CategoryEngineering,
			Priority: PriorityHigh, CompletionState: CompletionComplete, IsCritical: true,
			// deliberately no evidence
		})
		return l
	})
}

// allCompleteGen produces a launch with an owner where every item is complete
// for scoring, every critical item has evidence, and no blocking risk exists
// (Ready via 6.12). At least one item is present.
func allCompleteGen() *rapid.Generator[LaunchData] {
	return rapid.Custom(func(t *rapid.T) LaunchData {
		nItems := rapid.IntRange(1, 5).Draw(t, "nItems")
		items := make([]ChecklistItem, nItems)
		for i := 0; i < nItems; i++ {
			crit := rapid.Bool().Draw(t, fmt.Sprintf("crit%d", i))
			it := ChecklistItem{
				ID: fmt.Sprintf("i%d", i), Title: "done", Category: categoryGen.Draw(t, fmt.Sprintf("cat%d", i)),
				Priority: priorityGen.Draw(t, fmt.Sprintf("pri%d", i)), CompletionState: CompletionComplete, IsCritical: crit,
			}
			if crit {
				it.Evidence = []Evidence{{ID: it.ID + "-e", Note: "evidence"}}
			}
			items[i] = it
		}
		// Only non-blocking risks (not High, or High-but-Resolved/Accepted).
		nRisks := rapid.IntRange(0, 3).Draw(t, "nRisks")
		risks := make([]Risk, 0, nRisks)
		for i := 0; i < nRisks; i++ {
			r := Risk{ID: fmt.Sprintf("r%d", i), Title: "r", Severity: SeverityLow,
				Likelihood: LikelihoodLow, Status: RiskOpen}
			risks = append(risks, r)
		}
		return LaunchData{ID: "launch", Name: "n", TargetDate: "2026-10-16", Owner: "owner",
			ChecklistItems: items, Risks: risks}
	})
}
