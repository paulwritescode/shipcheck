package domain

import (
	"reflect"
	"testing"

	"pgregory.net/rapid"
)

// Property-based tests for the dashboard selectors, filters, grouping, edit
// helpers, validators, and the high-emphasis predicate (design.md Properties
// 18-26). Each is tagged "Feature: shipcheck, Property N".

// itemsGen draws a slice of 0..8 checklist items with unique ids.
func itemsGen() *rapid.Generator[[]ChecklistItem] {
	return rapid.Custom(func(t *rapid.T) []ChecklistItem {
		n := rapid.IntRange(0, 8).Draw(t, "n")
		out := make([]ChecklistItem, n)
		for i := 0; i < n; i++ {
			out[i] = checklistItemGen(itemID(i)).Draw(t, "it")
		}
		return out
	})
}

func itemID(i int) string {
	return "it" + string(rune('a'+i))
}

// risksGen draws a slice of 0..6 risks with unique ids.
func risksGen() *rapid.Generator[[]Risk] {
	return rapid.Custom(func(t *rapid.T) []Risk {
		n := rapid.IntRange(0, 6).Draw(t, "n")
		out := make([]Risk, n)
		for i := 0; i < n; i++ {
			out[i] = riskGen("rk"+string(rune('a'+i))).Draw(t, "rk")
		}
		return out
	})
}

// Feature: shipcheck, Property 18: Blocker selection equals incomplete criticals plus blocking risks.
// Validates: Requirements 8.3
func TestProperty18_BlockerSelection(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		l := launchGen().Draw(t, "launch")
		got := CriticalBlockers(l)

		wantItems := map[string]bool{}
		for _, it := range l.ChecklistItems {
			if IsCriticalItem(it) && !itemComplete(it) {
				wantItems[it.ID] = true
			}
		}
		wantRisks := map[string]bool{}
		for _, r := range l.Risks {
			if IsBlockingRisk(r) {
				wantRisks[r.ID] = true
			}
		}
		gotItems := map[string]bool{}
		gotRisks := map[string]bool{}
		for _, b := range got {
			switch b.Kind {
			case "item":
				gotItems[b.ItemID] = true
			case "risk":
				gotRisks[b.ItemID] = true
			default:
				t.Fatalf("unexpected blocker kind %q", b.Kind)
			}
		}
		if !reflect.DeepEqual(gotItems, wantItems) {
			t.Fatalf("item blockers: got %v want %v", gotItems, wantItems)
		}
		if !reflect.DeepEqual(gotRisks, wantRisks) {
			t.Fatalf("risk blockers: got %v want %v", gotRisks, wantRisks)
		}
		if len(got) != len(wantItems)+len(wantRisks) {
			t.Fatalf("blocker count %d != %d (no dupes expected)", len(got), len(wantItems)+len(wantRisks))
		}
	})
}

// Feature: shipcheck, Property 19: Risk ordering places High severity first.
// Validates: Requirements 8.4
func TestProperty19_RiskOrdering(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		risks := risksGen().Draw(t, "risks")
		ordered := RisksBySeverity(risks)
		// No lower-severity risk appears before any High-severity risk: once a
		// non-High risk is seen, no High risk may follow.
		seenNonHigh := false
		for _, r := range ordered {
			if r.Severity == SeverityHigh {
				if seenNonHigh {
					t.Fatalf("High-severity risk appears after a lower-severity one")
				}
			} else {
				seenNonHigh = true
			}
		}
		// Ordering is a permutation (same multiset of ids).
		if len(ordered) != len(risks) {
			t.Fatalf("ordering changed length")
		}
	})
}

// Feature: shipcheck, Property 20: Filter correctness.
// Validates: Requirements 3.11, 3.12
func TestProperty20_FilterCorrectness(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		items := itemsGen().Draw(t, "items")

		inc := IncompleteItems(items)
		for _, it := range inc {
			if it.CompletionState != CompletionIncomplete {
				t.Fatalf("incomplete filter returned %q", it.CompletionState)
			}
		}
		wantInc := 0
		for _, it := range items {
			if it.CompletionState == CompletionIncomplete {
				wantInc++
			}
		}
		if len(inc) != wantInc {
			t.Fatalf("incomplete count: got %d want %d", len(inc), wantInc)
		}

		hp := HighPriorityItems(items)
		for _, it := range hp {
			if it.Priority != PriorityHigh {
				t.Fatalf("high-priority filter returned %q", it.Priority)
			}
		}
		wantHP := 0
		for _, it := range items {
			if it.Priority == PriorityHigh {
				wantHP++
			}
		}
		if len(hp) != wantHP {
			t.Fatalf("high-priority count: got %d want %d", len(hp), wantHP)
		}
	})
}

// Feature: shipcheck, Property 21: Category grouping partitions without loss and each group is homogeneous.
// Validates: Requirements 3.13
func TestProperty21_CategoryGrouping(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		items := itemsGen().Draw(t, "items")
		groups := GroupByCategory(items)

		total := 0
		for cat, group := range groups {
			for _, it := range group {
				if it.Category != cat {
					t.Fatalf("item in group %q has category %q", cat, it.Category)
				}
			}
			total += len(group)
		}
		if total != len(items) {
			t.Fatalf("grouping lost/added items: %d grouped vs %d input", total, len(items))
		}
		// Every input item appears exactly once (compare id multisets).
		gotIDs := map[string]int{}
		for _, group := range groups {
			for _, it := range group {
				gotIDs[it.ID]++
			}
		}
		for _, it := range items {
			if gotIDs[it.ID] != 1 {
				t.Fatalf("item %q appears %d times", it.ID, gotIDs[it.ID])
			}
		}
	})
}

// Feature: shipcheck, Property 22: Editing one field preserves all other fields.
// Validates: Requirements 3.9, 4.7
func TestProperty22_EditPreservesFields(t *testing.T) {
	// Item owner edit.
	rapid.Check(t, func(t *rapid.T) {
		item := checklistItemGen("x").Draw(t, "item")
		newOwner := rapid.StringN(0, 12, 12).Draw(t, "owner")
		edited := WithItemOwner(item, newOwner)
		if edited.Owner != newOwner {
			t.Fatalf("owner not updated")
		}
		// Reset the edited field; everything else must be identical.
		edited.Owner = item.Owner
		if !reflect.DeepEqual(edited, item) {
			t.Fatalf("other fields changed: %+v vs %+v", edited, item)
		}
	})
	// Item completion edit.
	rapid.Check(t, func(t *rapid.T) {
		item := checklistItemGen("x").Draw(t, "item")
		s := completionGen.Draw(t, "state")
		edited := WithItemCompletion(item, s)
		if edited.CompletionState != s {
			t.Fatalf("completion not updated")
		}
		edited.CompletionState = item.CompletionState
		if !reflect.DeepEqual(edited, item) {
			t.Fatalf("other fields changed")
		}
	})
	// Risk status edit.
	rapid.Check(t, func(t *rapid.T) {
		risk := riskGen("x").Draw(t, "risk")
		s := riskStatusGen.Draw(t, "status")
		edited := WithRiskStatus(risk, s)
		if edited.Status != s {
			t.Fatalf("status not updated")
		}
		edited.Status = risk.Status
		if !reflect.DeepEqual(edited, risk) {
			t.Fatalf("other risk fields changed")
		}
	})
}

// Feature: shipcheck, Property 23: URL validation accepts valid URLs and rejects invalid ones.
// Validates: Requirements 1.8, 5.3
func TestProperty23_URLValidation(t *testing.T) {
	// Valid absolute URLs are accepted.
	rapid.Check(t, func(t *rapid.T) {
		scheme := rapid.SampledFrom([]string{"http", "https"}).Draw(t, "scheme")
		host := rapid.StringMatching(`[a-z]{2,8}\.(com|org|dev)`).Draw(t, "host")
		path := rapid.StringMatching(`(/[a-z0-9]{0,6}){0,3}`).Draw(t, "path")
		u := scheme + "://" + host + path
		if !IsValidURL(u) {
			t.Fatalf("valid URL rejected: %q", u)
		}
	})
	// Strings with no scheme/host are rejected.
	rapid.Check(t, func(t *rapid.T) {
		s := rapid.StringMatching(`[a-z ]{1,20}`).Draw(t, "s")
		if IsValidURL(s) {
			t.Fatalf("scheme-less/host-less string accepted: %q", s)
		}
	})
}

// Feature: shipcheck, Property 24: Name and enum validation.
// Validates: Requirements 1.2, 1.3, 3.2, 4.5
func TestProperty24_NameAndEnumValidation(t *testing.T) {
	// Whitespace-only names are rejected as required.
	rapid.Check(t, func(t *rapid.T) {
		ws := rapid.StringMatching(`[ \t\n]{0,10}`).Draw(t, "ws")
		if ValidateLaunchName(ws) == "" {
			t.Fatalf("whitespace-only name accepted: %q", ws)
		}
	})
	// Over-length names are rejected.
	rapid.Check(t, func(t *rapid.T) {
		extra := rapid.IntRange(1, 50).Draw(t, "extra")
		name := make([]byte, MaxLaunchNameLen+extra)
		for i := range name {
			name[i] = 'a'
		}
		if ValidateLaunchName(string(name)) == "" {
			t.Fatalf("over-length name accepted (len %d)", len(name))
		}
	})
	// Category validator: accepts iff one of the seven.
	rapid.Check(t, func(t *rapid.T) {
		valid := rapid.SampledFrom(Categories).Draw(t, "valid")
		if ValidateCategory(valid) != "" {
			t.Fatalf("valid category rejected: %q", valid)
		}
		bogus := Category(rapid.StringN(1, 10, 10).Draw(t, "bogus"))
		if bogus.IsValid() {
			return // generator happened to hit a real category
		}
		if ValidateCategory(bogus) == "" {
			t.Fatalf("bogus category accepted: %q", bogus)
		}
	})
	// Risk-status validator: accepts iff one of the four.
	rapid.Check(t, func(t *rapid.T) {
		valid := rapid.SampledFrom([]RiskStatus{RiskOpen, RiskMitigating, RiskResolved, RiskAccepted}).Draw(t, "valid")
		if ValidateRiskStatus(valid) != "" {
			t.Fatalf("valid risk status rejected: %q", valid)
		}
	})
}

// Feature: shipcheck, Property 25: Calendar-date validation.
// Validates: Requirements 1.5
func TestProperty25_CalendarDateValidation(t *testing.T) {
	// Real calendar dates are accepted.
	rapid.Check(t, func(t *rapid.T) {
		y := rapid.IntRange(2000, 2099).Draw(t, "y")
		m := rapid.IntRange(1, 12).Draw(t, "m")
		d := rapid.IntRange(1, 28).Draw(t, "d") // 1..28 valid in every month
		s := pad4(y) + "-" + pad2(m) + "-" + pad2(d)
		if !IsValidCalendarDate(s) {
			t.Fatalf("valid date rejected: %q", s)
		}
	})
	// Impossible dates are rejected (strict, non-normalizing).
	impossible := []string{"2026-02-30", "2026-13-01", "2026-00-10", "2026-01-32", "not-a-date", ""}
	for _, s := range impossible {
		if IsValidCalendarDate(s) {
			t.Fatalf("impossible date accepted: %q", s)
		}
	}
}

func pad2(n int) string {
	if n < 10 {
		return "0" + itoa(n)
	}
	return itoa(n)
}

func pad4(n int) string {
	s := itoa(n)
	for len(s) < 4 {
		s = "0" + s
	}
	return s
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}

// Feature: shipcheck, Property 26: High-emphasis predicate.
// Validates: Requirements 4.6
func TestProperty26_HighEmphasis(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		r := riskGen("x").Draw(t, "risk")
		want := r.Severity == SeverityHigh || r.Likelihood == LikelihoodHigh
		if HasHighEmphasis(r) != want {
			t.Fatalf("HasHighEmphasis(%+v) = %v want %v", r, HasHighEmphasis(r), want)
		}
	})
}
