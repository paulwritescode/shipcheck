package domain

import (
	"net/url"
	"strings"
	"time"
)

// Pure validation functions with field-specific messages. They are reused by
// the API layer (to reject bad input with HTTP 400) and by domain construction.
// Each returns an empty string when valid, or a human-readable message
// identifying the problem.

// Field-length bounds (Req 1).
const (
	MaxLaunchNameLen  = 200
	MaxDescriptionLen = 5000
)

// ValidationError names a field and the reason it was rejected.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidateLaunchName checks a launch name (Req 1.1-1.3). Returns "" when valid.
func ValidateLaunchName(name string) string {
	if strings.TrimSpace(name) == "" {
		return "Launch name is required."
	}
	if len(name) > MaxLaunchNameLen {
		return "Launch name must be at most 200 characters."
	}
	return ""
}

// ValidateDescription checks the optional description length (Req 1.1).
func ValidateDescription(desc string) string {
	if len(desc) > MaxDescriptionLen {
		return "Description must be at most 5000 characters."
	}
	return ""
}

// ValidateTargetDate checks that the date is present and a valid calendar date
// in YYYY-MM-DD form (Req 1.4, 1.5). time.Parse rejects impossible dates such
// as 2026-02-30 because it uses strict (non-normalizing) parsing here.
func ValidateTargetDate(date string) string {
	if strings.TrimSpace(date) == "" {
		return "Target date is required."
	}
	if !IsValidCalendarDate(date) {
		return "Target date must be a valid calendar date (YYYY-MM-DD)."
	}
	return ""
}

// IsValidCalendarDate reports whether s is a valid YYYY-MM-DD calendar date.
// It round-trips the parse so normalized dates (e.g. 2026-02-30 -> 2026-03-02)
// are rejected rather than silently accepted.
func IsValidCalendarDate(s string) bool {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return false
	}
	return t.Format("2006-01-02") == s
}

// ValidateURL checks that a URL is syntactically valid (Req 1.8, 5.3). An
// empty string is treated as "not provided" and is valid; callers that require
// a URL should check for emptiness separately.
func ValidateURL(raw string) string {
	if raw == "" {
		return ""
	}
	if !IsValidURL(raw) {
		return "URL is not valid."
	}
	return ""
}

// IsValidURL reports whether raw is a syntactically valid absolute URL with a
// scheme and host (e.g. https://example.com/x). Relative or scheme-less
// strings are rejected.
func IsValidURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	return u.Scheme != "" && u.Host != ""
}

// ValidateCategory reports a message when c is not one of the seven categories.
func ValidateCategory(c Category) string {
	if !c.IsValid() {
		return "Category must be one of the seven known categories."
	}
	return ""
}

// ValidatePriority reports a message when p is not a known priority.
func ValidatePriority(p Priority) string {
	if !p.IsValid() {
		return "Priority must be Low, Medium, or High."
	}
	return ""
}

// ValidateRiskStatus reports a message when s is not a known risk status.
func ValidateRiskStatus(s RiskStatus) string {
	if !s.IsValid() {
		return "Risk status must be Open, Mitigating, Resolved, or Accepted."
	}
	return ""
}
