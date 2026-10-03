// Package domain holds the pure ShipCheck domain core: the data models, the
// readiness engine, and the supporting pure functions. This package has no
// knowledge of storage, HTTP, the AWS SDK, clocks, or randomness — every
// function here is a pure function of its inputs. That purity is what makes
// the readiness engine directly verifiable with property-based tests.
package domain

// Category is one of the seven fixed launch-readiness groupings.
type Category string

const (
	CategoryRequirements     Category = "Requirements"
	CategoryEngineering      Category = "Engineering"
	CategoryQualityAssurance Category = "Quality Assurance"
	CategoryDocumentation    Category = "Documentation"
	CategoryOwnership        Category = "Ownership"
	CategoryLaunchOperations Category = "Launch Operations"
	CategoryRisks            Category = "Risks"
)

// Categories lists every Category in display order. The readiness engine and
// the dashboard iterate this fixed set so output ordering is deterministic.
var Categories = []Category{
	CategoryRequirements,
	CategoryEngineering,
	CategoryQualityAssurance,
	CategoryDocumentation,
	CategoryOwnership,
	CategoryLaunchOperations,
	CategoryRisks,
}

// IsValid reports whether c is one of the seven known categories.
func (c Category) IsValid() bool {
	switch c {
	case CategoryRequirements, CategoryEngineering, CategoryQualityAssurance,
		CategoryDocumentation, CategoryOwnership, CategoryLaunchOperations,
		CategoryRisks:
		return true
	default:
		return false
	}
}

// CompletionState is the state of a checklist item. For readiness scoring only
// CompletionComplete counts as complete; InProgress, Blocked, and Incomplete
// all count as "not complete" (see IsCompleteForScoring).
type CompletionState string

const (
	CompletionComplete   CompletionState = "complete"
	CompletionInProgress CompletionState = "in progress"
	CompletionBlocked    CompletionState = "blocked"
	CompletionIncomplete CompletionState = "incomplete"
)

// IsValid reports whether s is a known completion state.
func (s CompletionState) IsValid() bool {
	switch s {
	case CompletionComplete, CompletionInProgress, CompletionBlocked, CompletionIncomplete:
		return true
	default:
		return false
	}
}

// Priority ranks a checklist item.
type Priority string

const (
	PriorityLow    Priority = "Low"
	PriorityMedium Priority = "Medium"
	PriorityHigh   Priority = "High"
)

// IsValid reports whether p is a known priority.
func (p Priority) IsValid() bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh:
		return true
	default:
		return false
	}
}

// Severity is a risk's impact level.
type Severity string

const (
	SeverityLow    Severity = "Low"
	SeverityMedium Severity = "Medium"
	SeverityHigh   Severity = "High"
)

// IsValid reports whether s is a known severity.
func (s Severity) IsValid() bool {
	switch s {
	case SeverityLow, SeverityMedium, SeverityHigh:
		return true
	default:
		return false
	}
}

// Likelihood is a risk's probability level.
type Likelihood string

const (
	LikelihoodLow    Likelihood = "Low"
	LikelihoodMedium Likelihood = "Medium"
	LikelihoodHigh   Likelihood = "High"
)

// IsValid reports whether l is a known likelihood.
func (l Likelihood) IsValid() bool {
	switch l {
	case LikelihoodLow, LikelihoodMedium, LikelihoodHigh:
		return true
	default:
		return false
	}
}

// RiskStatus is the lifecycle state of a risk.
type RiskStatus string

const (
	RiskOpen       RiskStatus = "Open"
	RiskMitigating RiskStatus = "Mitigating"
	RiskResolved   RiskStatus = "Resolved"
	RiskAccepted   RiskStatus = "Accepted"
)

// IsValid reports whether s is a known risk status.
func (s RiskStatus) IsValid() bool {
	switch s {
	case RiskOpen, RiskMitigating, RiskResolved, RiskAccepted:
		return true
	default:
		return false
	}
}

// ReadinessStatus is the computed launch state. Exactly one is assigned to a
// launch for a given launch state (see the readiness engine).
type ReadinessStatus string

const (
	StatusReady              ReadinessStatus = "Ready"
	StatusConditionallyReady ReadinessStatus = "Conditionally Ready"
	StatusNotReady           ReadinessStatus = "Not Ready"
	StatusNeedsReview        ReadinessStatus = "Needs Review"
)

// IsValid reports whether s is a known readiness status.
func (s ReadinessStatus) IsValid() bool {
	switch s {
	case StatusReady, StatusConditionallyReady, StatusNotReady, StatusNeedsReview:
		return true
	default:
		return false
	}
}

// CategoryState classifies a category's completion in the assessment.
type CategoryState string

const (
	CategoryComplete CategoryState = "complete"
	CategoryPartial  CategoryState = "partial"
	CategoryEmpty    CategoryState = "empty"
)
