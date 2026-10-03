package repo

import (
	"context"
	"errors"
	"testing"

	"github.com/paulwritescode/shipcheck/internal/domain"
)

func mkRepo(t *testing.T) (*InMemoryLaunchRepository, context.Context, domain.Launch) {
	t.Helper()
	r := NewInMemoryLaunchRepository()
	ctx := context.Background()
	l, err := r.Create(ctx, NewLaunch{ID: "l1", Name: "Test", TargetDate: "2026-10-16", Owner: "Alice"})
	if err != nil {
		t.Fatal(err)
	}
	return r, ctx, l
}

func TestInMemoryGetNotFound(t *testing.T) {
	r := NewInMemoryLaunchRepository()
	if _, err := r.Get(context.Background(), "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestInMemoryChecklistCRUDAndRecompute(t *testing.T) {
	r, ctx, _ := mkRepo(t)
	// Add an incomplete critical item -> status becomes Not Ready.
	l, err := r.AddChecklistItem(ctx, "l1", domain.ChecklistItem{
		ID: "i1", Title: "crit", Category: domain.CategoryEngineering,
		CompletionState: domain.CompletionBlocked, IsCritical: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if l.Assessment.Status != domain.StatusNotReady {
		t.Fatalf("after add incomplete critical: want Not Ready, got %s", l.Assessment.Status)
	}
	// Complete it with evidence -> Ready.
	l, err = r.UpdateChecklistItem(ctx, "l1", domain.ChecklistItem{
		ID: "i1", Title: "crit", Category: domain.CategoryEngineering,
		CompletionState: domain.CompletionComplete, IsCritical: true,
		Evidence: []domain.Evidence{{ID: "e1", Note: "done"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if l.Assessment.Status != domain.StatusReady {
		t.Fatalf("after complete+evidence: want Ready, got %s", l.Assessment.Status)
	}
	// Delete it -> empty launch -> Needs Review.
	l, err = r.DeleteChecklistItem(ctx, "l1", "i1")
	if err != nil {
		t.Fatal(err)
	}
	if l.Assessment.Status != domain.StatusNeedsReview {
		t.Fatalf("after delete: want Needs Review, got %s", l.Assessment.Status)
	}
}

func TestInMemoryShareTokenLookup(t *testing.T) {
	r, ctx, _ := mkRepo(t)
	// No link yet.
	if _, err := r.GetByShareToken(ctx, "tok"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound before share, got %v", err)
	}
	// Enable a link.
	if _, err := r.SetShareLink(ctx, "l1", &domain.ShareLink{Token: "tok", Enabled: true}); err != nil {
		t.Fatal(err)
	}
	got, err := r.GetByShareToken(ctx, "tok")
	if err != nil {
		t.Fatalf("enabled token should resolve: %v", err)
	}
	if got.ID != "l1" {
		t.Fatalf("wrong launch: %s", got.ID)
	}
	// Disable it -> no longer resolves.
	if _, err := r.SetShareLink(ctx, "l1", &domain.ShareLink{Token: "tok", Enabled: false}); err != nil {
		t.Fatal(err)
	}
	if _, err := r.GetByShareToken(ctx, "tok"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("disabled token should not resolve, got %v", err)
	}
}

func TestInMemoryDelete(t *testing.T) {
	r, ctx, _ := mkRepo(t)
	if err := r.Delete(ctx, "l1"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Get(ctx, "l1"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound after delete, got %v", err)
	}
	if err := r.Delete(ctx, "l1"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("deleting missing: want ErrNotFound, got %v", err)
	}
}
