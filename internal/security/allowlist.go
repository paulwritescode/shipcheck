// Package security holds the trust-boundary guards: the external-fetch
// allowlist and the snapshot redaction that strips prohibited data before it
// can reach a model. These are pure, testable functions.
package security

import (
	"net/url"
	"strings"
)

// DefaultAllowedHosts is the v1 allowlist of public hosts the advisor may fetch
// external context from (Req 17.6, 19). Only public, read-only sources.
var DefaultAllowedHosts = []string{
	"github.com",
	"api.github.com",
	"raw.githubusercontent.com",
	"objects.githubusercontent.com",
}

// Allowlist decides whether an external URL may be fetched.
type Allowlist struct {
	hosts map[string]bool
}

// NewAllowlist builds an allowlist from the given hosts (lower-cased).
func NewAllowlist(hosts []string) *Allowlist {
	m := make(map[string]bool, len(hosts))
	for _, h := range hosts {
		m[strings.ToLower(h)] = true
	}
	return &Allowlist{hosts: m}
}

// DefaultAllowlist returns the v1 allowlist.
func DefaultAllowlist() *Allowlist { return NewAllowlist(DefaultAllowedHosts) }

// Permits reports whether raw is an https URL whose host is on the allowlist.
// Non-URLs, non-https schemes, and off-allowlist hosts are rejected. This is
// the single gate every external fetch must pass (Req 17.6, 19.1).
func (a *Allowlist) Permits(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	if u.Scheme != "https" {
		return false
	}
	host := strings.ToLower(u.Hostname())
	return a.hosts[host]
}
