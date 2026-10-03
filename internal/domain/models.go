package domain

// Evidence is a stored link or note supporting a checklist item or risk.
// Evidence marked private is excluded from the public report.
type Evidence struct {
	ID        string `json:"id"`
	URL       string `json:"url,omitempty"`  // syntactically valid URL when present
	Note      string `json:"note,omitempty"` // free-text note
	IsPrivate bool   `json:"isPrivate"`      // excluded from the public report when true
}

// ChecklistItem is a unit of launch work in one of the seven categories.
type ChecklistItem struct {
	ID              string          `json:"id"`
	Title           string          `json:"title"`
	Category        Category        `json:"category"`
	Owner           string          `json:"owner,omitempty"`
	Priority        Priority        `json:"priority"`
	CompletionState CompletionState `json:"completionState"`
	IsCritical      bool            `json:"isCritical"` // a Critical_Item when true
	Evidence        []Evidence      `json:"evidence,omitempty"`
}

// Risk is a risk-register entry.
type Risk struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Severity    Severity   `json:"severity"`
	Likelihood  Likelihood `json:"likelihood"`
	Owner       string     `json:"owner,omitempty"`
	Mitigation  string     `json:"mitigation,omitempty"`
	Status      RiskStatus `json:"status"`
	DueDate     string     `json:"dueDate,omitempty"` // optional ISO date
	Evidence    []Evidence `json:"evidence,omitempty"`
}

// ShareLink is the opaque, revocable token that exposes a launch's public
// report. Absent/disabled means the report is not publicly reachable.
type ShareLink struct {
	Token   string `json:"token"`   // cryptographically random, opaque
	Enabled bool   `json:"enabled"` // regenerate/disable toggles this and the token
}

// LaunchBrief captures the launch context for the readiness review.
type LaunchBrief struct {
	WhatIsReleasing   string `json:"whatIsReleasing,omitempty"`
	Audience          string `json:"audience,omitempty"`
	SuccessDefinition string `json:"successDefinition,omitempty"`
	Requirements      string `json:"requirements,omitempty"`
	Constraints       string `json:"constraints,omitempty"`
	Dependencies      string `json:"dependencies,omitempty"`
}

// LaunchData is the canonical, storage-agnostic shape of a launch that the
// readiness engine consumes. The engine never mutates it.
type LaunchData struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`        // 1–200 chars
	Description      string          `json:"description"` // 0–5000 chars
	TargetDate       string          `json:"targetDate"`  // valid calendar date
	Owner            string          `json:"owner,omitempty"`
	ProductArea      string          `json:"productArea,omitempty"`
	RepositoryURL    string          `json:"repositoryUrl,omitempty"`
	DocumentationURL string          `json:"documentationUrl,omitempty"`
	Brief            *LaunchBrief    `json:"brief,omitempty"`
	PrivateNotes     string          `json:"privateNotes,omitempty"` // excluded from public report
	ChecklistItems   []ChecklistItem `json:"checklistItems"`
	Risks            []Risk          `json:"risks"`
	Share            *ShareLink      `json:"share,omitempty"`
}

// Launch is a LaunchData together with its most recently computed readiness
// assessment. The assessment is recomputed and stored on every mutation.
type Launch struct {
	LaunchData
	Assessment ReadinessAssessment `json:"assessment"`
}
