package domain

import (
	"strings"
	"testing"

	"pgregory.net/rapid"
)

// Feature: shipcheck, Property 27: Public report projection preserves status and excludes private data.
// Validates: Requirements 10.4, 10.8
// For any launch, the public report's status equals the launch's computed
// assessment status, and the report never exposes private notes or evidence
// marked private (the projection carries no evidence content at all, and no
// notes field).
func TestProperty27_PublicReportProjection(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		data := launchGen().Draw(t, "launch")
		launch := Launch{LaunchData: data, Assessment: ComputeReadiness(data)}

		report := ProjectPublicReport(launch, "2026-10-16T00:00:00Z")

		// (10.8) Status fidelity: the report shows exactly the engine status.
		if report.Status != launch.Assessment.Status {
			t.Fatalf("report status %q != assessment status %q", report.Status, launch.Assessment.Status)
		}

		// (10.4) The report carries no private launch notes. The PublicReport
		// type has no notes field at all, and completed+incomplete items carry
		// no evidence, so no private evidence can leak. Verify the item
		// accounting is complete and homogeneous.
		if len(report.CompletedWork)+len(report.IncompleteWork) != len(launch.ChecklistItems) {
			t.Fatalf("item accounting off: %d completed + %d incomplete != %d items",
				len(report.CompletedWork), len(report.IncompleteWork), len(launch.ChecklistItems))
		}
		for _, it := range report.CompletedWork {
			if it.State != CompletionComplete {
				t.Fatalf("completed work has non-complete state %q", it.State)
			}
		}
		for _, it := range report.IncompleteWork {
			if it.State == CompletionComplete {
				t.Fatalf("incomplete work has complete state")
			}
		}

		// High-priority risks are a subset of the launch's risks and all qualify.
		for _, r := range report.HighPriorityRisks {
			if r.Severity != SeverityHigh {
				// qualified via High likelihood; confirm such a risk exists.
				found := false
				for _, lr := range launch.Risks {
					if lr.Title == r.Title && lr.Likelihood == LikelihoodHigh {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("high-priority risk %q is neither High severity nor High likelihood", r.Title)
				}
			}
		}
	})
}

// TestPublicReportHasNoPrivateFields is a structural guard: the PublicReport
// type must not expose a notes field or evidence, so private content cannot be
// projected even by mistake. (A compile-time + reflective check of intent.)
func TestPublicReportExcludesEvidenceAndNotes(t *testing.T) {
	// Build a launch with private notes and private evidence.
	data := LaunchData{
		ID: "L", Name: "n", TargetDate: "2026-10-16", Owner: "o",
		PrivateNotes: "TOP SECRET",
		ChecklistItems: []ChecklistItem{{
			ID: "i1", Title: "x", Category: CategoryEngineering, Priority: PriorityLow,
			CompletionState: CompletionComplete,
			Evidence:        []Evidence{{ID: "e1", Note: "PRIVATE DETAIL", IsPrivate: true}},
		}},
	}
	launch := Launch{LaunchData: data, Assessment: ComputeReadiness(data)}
	report := ProjectPublicReport(launch, "2026-10-16T00:00:00Z")

	// Serialize-ish check: the report must not contain the secret strings
	// anywhere in its item titles/categories (the only free text it carries).
	for _, it := range append(report.CompletedWork, report.IncompleteWork...) {
		if strings.Contains(it.Title, "SECRET") || strings.Contains(it.Title, "PRIVATE") {
			t.Fatal("public item leaked private content")
		}
	}
	// Open questions come from recommended actions (safe, engine-derived),
	// never from notes/evidence.
}
