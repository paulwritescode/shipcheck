// Package seed builds the pre-populated "Team Inbox 2.0" demonstration launch.
//
// The builder is a pure function so the demo is reproducible and testable: the
// same inputs always produce the same launch, and the readiness engine then
// classifies it as Not Ready in its initial state (Req 12.7). The documented
// demo edits (complete the rollback item with evidence and resolve the
// high-severity risk) move it to Conditionally Ready (Req 12.8).
package seed

import (
	"github.com/paulwritescode/shipcheck/internal/domain"
	"github.com/paulwritescode/shipcheck/internal/repo"
)

// SeedLaunchID is the fixed id of the seeded demo launch so the SPA can link to
// it directly ("explore the demo").
const SeedLaunchID = "team-inbox-2"

// Fixed ids for the two items the demo flow interacts with, so tests and the
// SPA can reference them without guessing.
const (
	RollbackItemID = "ti-ops-rollback" // incomplete critical rollback doc (Launch Operations)
	HighRiskID     = "ti-risk-scale"   // the single high-severity risk
)

// ev builds a single evidence entry.
func ev(id, url, note string) domain.Evidence {
	return domain.Evidence{ID: id, URL: url, Note: note}
}

// TeamInbox returns the seeded "Team Inbox 2.0" launch as repository input.
//
// Shape (Req 12.1-12.6):
//   - name "Team Inbox 2.0", target 2026-10-16, owned.
//   - 18 checklist items: 13 complete, 3 in progress, 2 blocked.
//   - 4 risks, exactly one High severity.
//   - at least one item with no owner.
//   - an incomplete critical rollback-documentation item in Launch Operations.
//   - a public preview URL and GitHub issue/PR evidence links.
func TeamInbox() repo.NewLaunch {
	preview := "https://preview.shipcheck.example/team-inbox-2"

	items := []domain.ChecklistItem{
		// Requirements (3 complete, all with evidence where critical).
		mk("ti-req-1", "Assignment rules specified", domain.CategoryRequirements, "Priya", domain.PriorityHigh, domain.CompletionComplete, true,
			ev("e-req-1", "https://github.com/example/team-inbox/issues/101", "Spec issue")),
		mk("ti-req-2", "Saved views specified", domain.CategoryRequirements, "Priya", domain.PriorityMedium, domain.CompletionComplete, false),
		mk("ti-req-3", "Response-time reporting specified", domain.CategoryRequirements, "Priya", domain.PriorityMedium, domain.CompletionComplete, false),

		// Engineering (3 complete, 1 in progress, 1 blocked).
		mk("ti-eng-1", "Assignment-rules engine", domain.CategoryEngineering, "Marco", domain.PriorityHigh, domain.CompletionComplete, true,
			ev("e-eng-1", "https://github.com/example/team-inbox/pull/210", "PR #210")),
		mk("ti-eng-2", "Saved-views API", domain.CategoryEngineering, "Marco", domain.PriorityMedium, domain.CompletionComplete, false),
		mk("ti-eng-3", "Reporting aggregation job", domain.CategoryEngineering, "Marco", domain.PriorityMedium, domain.CompletionComplete, false),
		mk("ti-eng-4", "Realtime inbox sync", domain.CategoryEngineering, "Lena", domain.PriorityHigh, domain.CompletionInProgress, false),
		mk("ti-eng-5", "Search index migration", domain.CategoryEngineering, "", domain.PriorityHigh, domain.CompletionBlocked, false), // unowned + blocked

		// Quality Assurance (2 complete, 1 in progress).
		mk("ti-qa-1", "Assignment-rules test suite", domain.CategoryQualityAssurance, "Dana", domain.PriorityHigh, domain.CompletionComplete, true,
			ev("e-qa-1", "https://github.com/example/team-inbox/pull/215", "Test PR")),
		mk("ti-qa-2", "Load test saved views", domain.CategoryQualityAssurance, "Dana", domain.PriorityMedium, domain.CompletionComplete, false),
		mk("ti-qa-3", "Regression pass on reporting", domain.CategoryQualityAssurance, "Dana", domain.PriorityMedium, domain.CompletionInProgress, false),

		// Documentation (2 complete).
		mk("ti-doc-1", "User help: assignment rules", domain.CategoryDocumentation, "Sam", domain.PriorityLow, domain.CompletionComplete, false),
		mk("ti-doc-2", "Changelog entry", domain.CategoryDocumentation, "Sam", domain.PriorityLow, domain.CompletionInProgress, false),

		// Ownership (1 complete).
		mk("ti-own-1", "Launch owner assigned", domain.CategoryOwnership, "Priya", domain.PriorityMedium, domain.CompletionComplete, false),

		// Launch Operations (2 complete, 1 blocked critical rollback = the top blocker).
		mk("ti-ops-1", "Feature flag wired", domain.CategoryLaunchOperations, "Lena", domain.PriorityHigh, domain.CompletionComplete, false),
		mk("ti-ops-2", "Monitoring dashboards", domain.CategoryLaunchOperations, "Lena", domain.PriorityMedium, domain.CompletionComplete, false),
		mk(RollbackItemID, "Database migration rollback plan", domain.CategoryLaunchOperations, "", domain.PriorityHigh, domain.CompletionBlocked, true), // incomplete + critical + unowned

		// Risks category checklist (1 complete) to round the count to 18.
		mk("ti-rsk-1", "Risk review completed", domain.CategoryRisks, "Priya", domain.PriorityMedium, domain.CompletionComplete, false),
	}

	risks := []domain.Risk{
		{
			ID: HighRiskID, Title: "Migration fails under production load",
			Description: "The search index migration may time out on large workspaces.",
			Severity:    domain.SeverityHigh, Likelihood: domain.LikelihoodMedium,
			Owner: "", Status: domain.RiskOpen, // high + open + unowned = blocking
			Evidence: []domain.Evidence{ev("e-risk-1", "https://github.com/example/team-inbox/issues/230", "Export fails for large workspaces")},
		},
		{
			ID: "ti-risk-perms", Title: "Permission regression in saved views",
			Severity: domain.SeverityMedium, Likelihood: domain.LikelihoodMedium,
			Owner: "Dana", Mitigation: "Added permission test matrix", Status: domain.RiskMitigating,
		},
		{
			ID: "ti-risk-notify", Title: "Notification spam from assignment rules",
			Severity: domain.SeverityMedium, Likelihood: domain.LikelihoodLow,
			Owner: "Marco", Mitigation: "Rate limiting applied", Status: domain.RiskResolved,
		},
		{
			ID: "ti-risk-copy", Title: "Help copy not localized",
			Severity: domain.SeverityLow, Likelihood: domain.LikelihoodLow,
			Owner: "Sam", Status: domain.RiskAccepted,
		},
	}

	return repo.NewLaunch{
		ID:            SeedLaunchID,
		Name:          "Team Inbox 2.0",
		Description:   "A redesigned shared inbox with assignment rules, saved views, and response-time reporting for small support teams.",
		TargetDate:    "2026-10-16",
		Owner:         "Priya",
		ProductArea:   "Support",
		RepositoryURL: "https://github.com/example/team-inbox",
		Brief: &domain.LaunchBrief{
			WhatIsReleasing:   "Team Inbox 2.0: assignment rules, saved views, response-time reporting.",
			Audience:          "Small support teams.",
			SuccessDefinition: "Faster triage and visible response-time metrics without regressions.",
			Dependencies:      "Search index migration; realtime sync service.",
		},
		ChecklistItems: withPreview(items, preview),
		Risks:          risks,
	}
}

// withPreview attaches the public preview URL as evidence on the first
// Launch Operations item, surfacing it in the demo (Req 12.6).
func withPreview(items []domain.ChecklistItem, preview string) []domain.ChecklistItem {
	for i := range items {
		if items[i].ID == "ti-ops-1" {
			items[i].Evidence = append(items[i].Evidence, ev("e-preview", preview, "Public preview URL"))
			break
		}
	}
	return items
}

// mk builds a checklist item, attaching any provided evidence.
func mk(id, title string, cat domain.Category, owner string, p domain.Priority, state domain.CompletionState, critical bool, evidence ...domain.Evidence) domain.ChecklistItem {
	return domain.ChecklistItem{
		ID:              id,
		Title:           title,
		Category:        cat,
		Owner:           owner,
		Priority:        p,
		CompletionState: state,
		IsCritical:      critical,
		Evidence:        evidence,
	}
}
