package security

import (
	"net/url"
	"testing"

	"pgregory.net/rapid"

	"github.com/paulwritescode/shipcheck/internal/domain"
)

// Feature: shipcheck, Property 33: External-fetch allowlist enforcement.
// Validates: Requirements 17.6, 19.1
// For any URL, the guard permits it if and only if it is an https URL whose
// host is on the allowlist.
func TestProperty33_AllowlistEnforcement(t *testing.T) {
	allow := DefaultAllowlist()

	schemeGen := rapid.SampledFrom([]string{"https", "http", "ftp", ""})
	hostGen := rapid.SampledFrom([]string{
		"github.com", "api.github.com", "raw.githubusercontent.com", "objects.githubusercontent.com", // allowed
		"evil.example.com", "github.com.evil.com", "notgithub.com", "", // not allowed
	})
	pathGen := rapid.StringMatching(`(/[a-z0-9]{0,8}){0,3}`)

	allowedHosts := map[string]bool{}
	for _, h := range DefaultAllowedHosts {
		allowedHosts[h] = true
	}

	rapid.Check(t, func(t *rapid.T) {
		scheme := schemeGen.Draw(t, "scheme")
		host := hostGen.Draw(t, "host")
		path := pathGen.Draw(t, "path")

		raw := scheme + "://" + host + path
		got := allow.Permits(raw)

		// Independently decide the expected answer.
		want := scheme == "https" && allowedHosts[host]
		// Parse to mirror the guard's host extraction (handles empty host etc.).
		if u, err := url.Parse(raw); err == nil {
			want = u.Scheme == "https" && allowedHosts[u.Hostname()]
		} else {
			want = false
		}

		if got != want {
			t.Fatalf("Permits(%q) = %v, want %v", raw, got, want)
		}
	})
}

// launchWithPrivateGen draws a launch that may contain private notes and
// private evidence, so redaction has something to strip.
func launchWithPrivateGen() *rapid.Generator[domain.Launch] {
	evGen := func(prefix string) *rapid.Generator[[]domain.Evidence] {
		return rapid.Custom(func(t *rapid.T) []domain.Evidence {
			n := rapid.IntRange(0, 3).Draw(t, "nEv")
			out := make([]domain.Evidence, n)
			for i := 0; i < n; i++ {
				out[i] = domain.Evidence{
					ID:        prefix + string(rune('a'+i)),
					Note:      rapid.StringN(0, 10, 10).Draw(t, "note"),
					IsPrivate: rapid.Bool().Draw(t, "priv"),
				}
			}
			return out
		})
	}
	return rapid.Custom(func(t *rapid.T) domain.Launch {
		nItems := rapid.IntRange(0, 4).Draw(t, "nItems")
		items := make([]domain.ChecklistItem, nItems)
		for i := 0; i < nItems; i++ {
			items[i] = domain.ChecklistItem{
				ID: "i" + string(rune('a'+i)), Title: "t", Category: domain.CategoryEngineering,
				Priority: domain.PriorityLow, CompletionState: domain.CompletionIncomplete,
				Evidence: evGen("ie"+string(rune('a'+i))).Draw(t, "iev"),
			}
		}
		nRisks := rapid.IntRange(0, 3).Draw(t, "nRisks")
		risks := make([]domain.Risk, nRisks)
		for i := 0; i < nRisks; i++ {
			risks[i] = domain.Risk{
				ID: "r" + string(rune('a'+i)), Title: "r", Severity: domain.SeverityLow,
				Likelihood: domain.LikelihoodLow, Status: domain.RiskOpen,
				Evidence: evGen("re"+string(rune('a'+i))).Draw(t, "rev"),
			}
		}
		var notes string
		if rapid.Bool().Draw(t, "hasNotes") {
			notes = rapid.StringN(1, 30, 30).Draw(t, "notes")
		}
		return domain.Launch{
			LaunchData: domain.LaunchData{
				ID: "L", Name: "n", TargetDate: "2026-10-16", Owner: "o",
				PrivateNotes: notes, ChecklistItems: items, Risks: risks,
				Share: &domain.ShareLink{Token: "secret-token", Enabled: true},
			},
		}
	})
}

// Feature: shipcheck, Property 34: Prohibited-data redaction in the model snapshot.
// Validates: Requirements 19.2
// For any launch, the redacted copy contains no private notes, no private
// evidence, and no share link — nothing prohibited can reach the model.
func TestProperty34_ProhibitedDataRedaction(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		launch := launchWithPrivateGen().Draw(t, "launch")
		redacted := RedactForModel(launch)

		if ContainsProhibited(redacted) {
			t.Fatal("redacted launch still contains prohibited data")
		}
		if redacted.PrivateNotes != "" {
			t.Fatal("private notes not stripped")
		}
		if redacted.Share != nil {
			t.Fatal("share link not stripped")
		}
		for _, it := range redacted.ChecklistItems {
			for _, e := range it.Evidence {
				if e.IsPrivate {
					t.Fatal("private item evidence survived redaction")
				}
			}
		}
		for _, r := range redacted.Risks {
			for _, e := range r.Evidence {
				if e.IsPrivate {
					t.Fatal("private risk evidence survived redaction")
				}
			}
		}

		// Redaction preserves public evidence: counts match the non-private
		// subset of the original.
		for i, it := range launch.ChecklistItems {
			want := 0
			for _, e := range it.Evidence {
				if !e.IsPrivate {
					want++
				}
			}
			if len(redacted.ChecklistItems[i].Evidence) != want {
				t.Fatalf("item %d public evidence count: got %d want %d", i, len(redacted.ChecklistItems[i].Evidence), want)
			}
		}
	})
}

// TestRedactDoesNotMutateInput confirms redaction is pure.
func TestRedactDoesNotMutateInput(t *testing.T) {
	l := domain.Launch{LaunchData: domain.LaunchData{
		ID: "L", PrivateNotes: "secret",
		ChecklistItems: []domain.ChecklistItem{{ID: "i1", Evidence: []domain.Evidence{{ID: "e1", IsPrivate: true}}}},
		Share:          &domain.ShareLink{Token: "t", Enabled: true},
	}}
	_ = RedactForModel(l)
	if l.PrivateNotes != "secret" || l.Share == nil || !l.ChecklistItems[0].Evidence[0].IsPrivate {
		t.Fatal("RedactForModel mutated its input")
	}
}
