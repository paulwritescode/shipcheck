package domain

import "sort"

// Rule identifiers attached to reason entries so each reason names the matched
// Requirement-6 rule (Req 7.1). Kept as stable strings for the API/UI.
const (
	RuleEmptyLaunch        = "6.4-empty-launch"
	RuleNoOwner            = "6.5-no-owner"
	RuleCompleteNoEvidence = "6.6-complete-without-evidence"
	RuleIncompleteCritical = "6.7-incomplete-critical-item"
	RuleBlockingRisk       = "6.8-blocking-risk"
	RuleNonCriticalGap     = "6.10-non-critical-gap"
	RuleAcceptedRisk       = "6.11-accepted-risk"
	RuleAllComplete        = "6.12-ready"
	RuleDefaultReady       = "6.13-default-ready"
)

// ComputeReadiness is the public entrypoint of the readiness engine. It is a
// pure, total function of the launch data: the same launch state always yields
// the same assessment (Req 6.15), it never panics on well-formed input, and it
// assigns exactly one status (Req 6.14).
func ComputeReadiness(l LaunchData) ReadinessAssessment {
	status, reasons := EvaluateRules(l)
	return ReadinessAssessment{
		Status:     status,
		Reasons:    reasons,
		Categories: ComputeCategoryStates(l),
	}
}

// EvaluateRules applies the fixed precedence from Requirement 6 — Needs Review,
// then Not Ready, then Conditionally Ready, then Ready — and returns the first
// matching status together with the reason entries that justify it (Req 6.3).
//
// Reason ordering is deterministic: items and risks are traversed in
// identifier-sorted order, so the reason set depends only on launch state, not
// on input slice ordering (Req 6.15).
func EvaluateRules(l LaunchData) (ReadinessStatus, []ReasonEntry) {
	items := sortedItems(l.ChecklistItems)
	risks := sortedRisks(l.Risks)

	// ---- Tier 1: Needs Review (Req 6.4, 6.5, 6.6) ----
	var needsReview []ReasonEntry

	// 6.4 empty launch (launch-level reason, no item id).
	if len(l.ChecklistItems) == 0 && len(l.Risks) == 0 {
		needsReview = append(needsReview, ReasonEntry{
			RuleID:            RuleEmptyLaunch,
			Message:           "Launch has no checklist items and no risks, so readiness cannot be assessed.",
			RecommendedAction: "Add checklist items and risks, or start from the seeded demo launch.",
		})
	}
	// 6.5 absent owner (launch-level reason).
	if !hasOwner(l) {
		needsReview = append(needsReview, ReasonEntry{
			RuleID:            RuleNoOwner,
			Message:           "Launch has no owner.",
			RecommendedAction: "Assign a launch owner.",
		})
	}
	// 6.6 critical item marked complete with zero evidence (contradiction).
	for _, item := range items {
		if IsCriticalItem(item) && itemComplete(item) && !hasEvidence(item) {
			needsReview = append(needsReview, ReasonEntry{
				RuleID:            RuleCompleteNoEvidence,
				Message:           "Critical item \"" + item.Title + "\" is marked complete but has no supporting evidence.",
				ItemID:            item.ID,
				RecommendedAction: "Attach evidence to critical item \"" + item.Title + "\", or change its status.",
			})
		}
	}
	if len(needsReview) > 0 {
		return StatusNeedsReview, needsReview
	}

	// ---- Tier 2: Not Ready (Req 6.7, 6.8, 6.9) ----
	var notReady []ReasonEntry

	// 6.7 (+6.9, which it subsumes): each critical item not complete for scoring.
	for _, item := range items {
		if IsCriticalItem(item) && !itemComplete(item) {
			action := "Complete critical item \"" + item.Title + "\"."
			if item.Category == CategoryLaunchOperations {
				action = "Complete the critical launch-operations item \"" + item.Title +
					"\" (e.g. the rollback/contingency plan)."
			}
			notReady = append(notReady, ReasonEntry{
				RuleID:            RuleIncompleteCritical,
				Message:           "Critical item \"" + item.Title + "\" is not complete.",
				ItemID:            item.ID,
				RecommendedAction: action,
			})
		}
	}
	// 6.8 each blocking risk.
	for _, risk := range risks {
		if IsBlockingRisk(risk) {
			notReady = append(notReady, ReasonEntry{
				RuleID:            RuleBlockingRisk,
				Message:           "High-severity risk \"" + risk.Title + "\" is unresolved.",
				ItemID:            risk.ID,
				RecommendedAction: "Resolve or accept high-severity risk \"" + risk.Title + "\".",
			})
		}
	}
	if len(notReady) > 0 {
		return StatusNotReady, notReady
	}

	// Past this point: every critical item is complete for scoring and no
	// blocking risk exists. (Tier 1 already ruled out any complete-critical
	// item lacking evidence, so every critical item has evidence here.)

	// ---- Tier 3: Conditionally Ready (Req 6.10, 6.11) ----
	var conditional []ReasonEntry

	// 6.10 at least one non-critical item not complete for scoring.
	for _, item := range items {
		if !IsCriticalItem(item) && !itemComplete(item) {
			conditional = append(conditional, ReasonEntry{
				RuleID:            RuleNonCriticalGap,
				Message:           "Non-critical item \"" + item.Title + "\" is not complete.",
				ItemID:            item.ID,
				RecommendedAction: "Complete non-critical item \"" + item.Title + "\" to reach Ready.",
			})
		}
	}
	// 6.11 at least one accepted risk.
	for _, risk := range risks {
		if risk.Status == RiskAccepted {
			conditional = append(conditional, ReasonEntry{
				RuleID:  RuleAcceptedRisk,
				Message: "Risk \"" + risk.Title + "\" is accepted rather than resolved.",
				ItemID:  risk.ID,
			})
		}
	}
	if len(conditional) > 0 {
		return StatusConditionallyReady, conditional
	}

	// ---- Tier 4: Ready (Req 6.12 / default 6.13) ----
	// Every checklist item is complete for scoring, every critical item has
	// evidence, no blocking risk exists, and the launch has an owner.
	ruleID := RuleAllComplete
	if len(l.ChecklistItems) == 0 {
		// No items but at least one risk (empty-launch was ruled out in tier 1)
		// and none Accepted/blocking: fall through to the default-Ready rule.
		ruleID = RuleDefaultReady
	}
	return StatusReady, []ReasonEntry{{
		RuleID:  ruleID,
		Message: "All critical work is complete with evidence, no high-severity risks are unresolved, and the launch has an owner.",
	}}
}

// sortedItems returns a copy of items ordered by ID so reason ordering is
// deterministic. The input slice is never mutated (purity).
func sortedItems(items []ChecklistItem) []ChecklistItem {
	out := make([]ChecklistItem, len(items))
	copy(out, items)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// sortedRisks returns a copy of risks ordered by ID. The input is not mutated.
func sortedRisks(risks []Risk) []Risk {
	out := make([]Risk, len(risks))
	copy(out, risks)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
