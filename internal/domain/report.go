package domain

// PublicReport is the read-only, externally shareable projection of a launch
// (Req 10.3). It deliberately excludes private launch notes and any evidence
// marked private (Req 10.4), and it carries the same readiness status the
// engine computed for the launch (Req 10.8).
type PublicReport struct {
	Name              string               `json:"name"`
	TargetDate        string               `json:"targetDate"`
	Status            ReadinessStatus      `json:"status"`
	Categories        []CategoryCompletion `json:"categories"`
	CompletedWork     []PublicItem         `json:"completedWork"`
	IncompleteWork    []PublicItem         `json:"incompleteWork"`
	HighPriorityRisks []PublicRisk         `json:"highPriorityRisks"`
	OpenQuestions     []string             `json:"openQuestions"`
	LastUpdated       string               `json:"lastUpdated"`
}

// PublicItem is a checklist item as shown on the public report: no private
// evidence, no owner PII beyond the (team-chosen) owner label.
type PublicItem struct {
	Title    string          `json:"title"`
	Category Category        `json:"category"`
	State    CompletionState `json:"completionState"`
	Critical bool            `json:"critical"`
}

// PublicRisk is a risk as shown on the public report.
type PublicRisk struct {
	Title    string     `json:"title"`
	Severity Severity   `json:"severity"`
	Status   RiskStatus `json:"status"`
}

// ProjectPublicReport builds the public report for a launch. lastUpdated is an
// RFC3339 timestamp supplied by the caller (the engine/domain stays clock-free;
// the API layer passes the time). It never includes private notes or private
// evidence, and its status equals the launch's computed assessment status.
func ProjectPublicReport(l Launch, lastUpdated string) PublicReport {
	var completed, incomplete []PublicItem
	for _, it := range l.ChecklistItems {
		pi := PublicItem{Title: it.Title, Category: it.Category, State: it.CompletionState, Critical: it.IsCritical}
		if IsCompleteForScoring(it.CompletionState) {
			completed = append(completed, pi)
		} else {
			incomplete = append(incomplete, pi)
		}
	}

	// High-priority risks: High severity or High likelihood, ordered High-first.
	var highRisks []PublicRisk
	for _, r := range RisksBySeverity(l.Risks) {
		if r.Severity == SeverityHigh || r.Likelihood == LikelihoodHigh {
			highRisks = append(highRisks, PublicRisk{Title: r.Title, Severity: r.Severity, Status: r.Status})
		}
	}

	// Open questions: the recommended next actions for unresolved blockers/gaps.
	var questions []string
	for _, reason := range l.Assessment.Reasons {
		if reason.RecommendedAction != "" {
			questions = append(questions, reason.RecommendedAction)
		}
	}

	return PublicReport{
		Name:              l.Name,
		TargetDate:        l.TargetDate,
		Status:            l.Assessment.Status,
		Categories:        l.Assessment.Categories,
		CompletedWork:     completed,
		IncompleteWork:    incomplete,
		HighPriorityRisks: highRisks,
		OpenQuestions:     questions,
		LastUpdated:       lastUpdated,
	}
}
