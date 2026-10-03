package repo

import (
	"context"
	"reflect"
	"testing"

	"pgregory.net/rapid"

	"github.com/paulwritescode/shipcheck/internal/domain"
)

// launchDataGen draws well-formed LaunchData for persistence round-trip tests.
// IDs are unique within the launch so stored/loaded aggregates compare cleanly.
func launchDataGen() *rapid.Generator[domain.LaunchData] {
	return rapid.Custom(func(t *rapid.T) domain.LaunchData {
		nItems := rapid.IntRange(0, 5).Draw(t, "nItems")
		items := make([]domain.ChecklistItem, nItems)
		for i := 0; i < nItems; i++ {
			nEv := rapid.IntRange(0, 2).Draw(t, "nEv")
			ev := make([]domain.Evidence, nEv)
			for j := 0; j < nEv; j++ {
				ev[j] = domain.Evidence{
					ID:        itemIndexID("e", i*10+j),
					URL:       rapid.StringN(0, 20, 20).Draw(t, "evURL"),
					Note:      rapid.StringN(0, 20, 20).Draw(t, "evNote"),
					IsPrivate: rapid.Bool().Draw(t, "evPriv"),
				}
			}
			items[i] = domain.ChecklistItem{
				ID:              itemIndexID("i", i),
				Title:           rapid.StringN(1, 20, 20).Draw(t, "title"),
				Category:        rapid.SampledFrom(domain.Categories).Draw(t, "cat"),
				Owner:           rapid.StringN(0, 10, 10).Draw(t, "owner"),
				Priority:        rapid.SampledFrom([]domain.Priority{domain.PriorityLow, domain.PriorityMedium, domain.PriorityHigh}).Draw(t, "pri"),
				CompletionState: rapid.SampledFrom([]domain.CompletionState{domain.CompletionComplete, domain.CompletionInProgress, domain.CompletionBlocked, domain.CompletionIncomplete}).Draw(t, "comp"),
				IsCritical:      rapid.Bool().Draw(t, "crit"),
				Evidence:        ev,
			}
		}
		nRisks := rapid.IntRange(0, 4).Draw(t, "nRisks")
		risks := make([]domain.Risk, nRisks)
		for i := 0; i < nRisks; i++ {
			risks[i] = domain.Risk{
				ID:         itemIndexID("r", i),
				Title:      rapid.StringN(1, 20, 20).Draw(t, "rtitle"),
				Severity:   rapid.SampledFrom([]domain.Severity{domain.SeverityLow, domain.SeverityMedium, domain.SeverityHigh}).Draw(t, "sev"),
				Likelihood: rapid.SampledFrom([]domain.Likelihood{domain.LikelihoodLow, domain.LikelihoodMedium, domain.LikelihoodHigh}).Draw(t, "lik"),
				Status:     rapid.SampledFrom([]domain.RiskStatus{domain.RiskOpen, domain.RiskMitigating, domain.RiskResolved, domain.RiskAccepted}).Draw(t, "rstatus"),
			}
		}
		var brief *domain.LaunchBrief
		if rapid.Bool().Draw(t, "hasBrief") {
			brief = &domain.LaunchBrief{WhatIsReleasing: rapid.StringN(0, 30, 30).Draw(t, "wir")}
		}
		return domain.LaunchData{
			ID:             "launch",
			Name:           rapid.StringN(1, 30, 30).Draw(t, "name"),
			TargetDate:     "2026-10-16",
			Owner:          rapid.StringN(0, 10, 10).Draw(t, "lowner"),
			Brief:          brief,
			ChecklistItems: items,
			Risks:          risks,
		}
	})
}

func itemIndexID(prefix string, i int) string {
	return prefix + string(rune('a'+i%26)) + string(rune('0'+i/26))
}

// Feature: shipcheck, Property 28: Persistence round-trip.
// Validates: Requirements 13.2
// For any launch data, creating it and reloading it yields a launch equal to
// the original across all persisted fields (items, risks, evidence, brief).
func TestProperty28_PersistenceRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		data := launchDataGen().Draw(t, "launch")
		r := NewInMemoryLaunchRepository()
		ctx := context.Background()

		created, err := r.Create(ctx, NewLaunch{
			ID: data.ID, Name: data.Name, Description: data.Description,
			TargetDate: data.TargetDate, Owner: data.Owner, Brief: data.Brief,
			ChecklistItems: data.ChecklistItems, Risks: data.Risks,
		})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		loaded, err := r.Get(ctx, data.ID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		// The persisted inputs round-trip identically.
		if !reflect.DeepEqual(created.LaunchData, loaded.LaunchData) {
			t.Fatalf("round-trip mismatch:\n created=%+v\n loaded=%+v", created.LaunchData, loaded.LaunchData)
		}
		// The stored assessment matches a fresh engine computation (consistency).
		if !reflect.DeepEqual(loaded.Assessment, domain.ComputeReadiness(loaded.LaunchData)) {
			t.Fatalf("stored assessment differs from recomputed")
		}
	})
}

// TestRoundTripIsolation confirms the repo deep-copies: mutating a returned
// launch's slices must not change stored state.
func TestRoundTripIsolation(t *testing.T) {
	r := NewInMemoryLaunchRepository()
	ctx := context.Background()
	created, err := r.Create(ctx, NewLaunch{
		Name: "n", TargetDate: "2026-10-16", Owner: "o",
		ChecklistItems: []domain.ChecklistItem{{ID: "i1", Title: "x", Category: domain.CategoryEngineering, CompletionState: domain.CompletionComplete}},
	})
	if err != nil {
		t.Fatal(err)
	}
	// Mutate the returned slice in place.
	created.ChecklistItems[0].Title = "HACKED"
	reloaded, err := r.Get(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.ChecklistItems[0].Title != "x" {
		t.Fatalf("stored state mutated through returned slice: %q", reloaded.ChecklistItems[0].Title)
	}
}
