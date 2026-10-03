package advisor

import (
	"context"
	"testing"

	"github.com/paulwritescode/shipcheck/internal/domain"
	"github.com/paulwritescode/shipcheck/internal/seed"
)

// seedSnapshot builds a LaunchSnapshot from the seeded Team Inbox launch with a
// freshly computed assessment (what the advisor flow passes to a provider).
func seedSnapshot() LaunchSnapshot {
	in := seed.TeamInbox()
	data := domain.LaunchData{
		ID: in.ID, Name: in.Name, Description: in.Description, TargetDate: in.TargetDate,
		Owner: in.Owner, ChecklistItems: in.ChecklistItems, Risks: in.Risks, Brief: in.Brief,
	}
	return LaunchSnapshot{Launch: domain.Launch{LaunchData: data, Assessment: domain.ComputeReadiness(data)}}
}

func TestSelectProviderDefaultsToFallback(t *testing.T) {
	t.Setenv("SHIPCHECK_MODEL_PROVIDER", "")
	p := SelectProvider()
	if p.Name() != "fallback" {
		t.Fatalf("default provider: want fallback, got %s", p.Name())
	}
}

func TestSelectProviderUnknownFallsBack(t *testing.T) {
	t.Setenv("SHIPCHECK_MODEL_PROVIDER", "wat")
	if p := SelectProvider(); p.Name() != "fallback" {
		t.Fatalf("unknown provider: want fallback, got %s", p.Name())
	}
}

func TestSelectNvidiaWithoutKeyFallsBack(t *testing.T) {
	t.Setenv("SHIPCHECK_MODEL_PROVIDER", "nvidia")
	t.Setenv("NVIDIA_API_KEY", "") // no key configured
	if p := SelectProvider(); p.Name() != "fallback" {
		t.Fatalf("nvidia w/o key: want fallback, got %s", p.Name())
	}
}

func TestSelectNvidiaWithKey(t *testing.T) {
	t.Setenv("SHIPCHECK_MODEL_PROVIDER", "nvidia")
	t.Setenv("NVIDIA_API_KEY", "test-key") // server-side only; never sent to client
	if p := SelectProvider(); p.Name() != "nvidia" {
		t.Fatalf("nvidia w/ key: want nvidia, got %s", p.Name())
	}
}

func TestSelectHuggingFaceWithoutKeyFallsBack(t *testing.T) {
	t.Setenv("SHIPCHECK_MODEL_PROVIDER", "hf")
	t.Setenv("HF_TOKEN", "")
	if p := SelectProvider(); p.Name() != "fallback" {
		t.Fatalf("hf w/o token: want fallback, got %s", p.Name())
	}
}

func TestSelectBedrockAlwaysSelectable(t *testing.T) {
	t.Setenv("SHIPCHECK_MODEL_PROVIDER", "bedrock")
	// Bedrock uses IAM-role auth (no key) and is selectable; it degrades to
	// fallback at call time if AWS is unavailable.
	if p := SelectProvider(); p.Name() != "bedrock" {
		t.Fatalf("bedrock: want bedrock, got %s", p.Name())
	}
}

func TestFallbackNeedsNoConfig(t *testing.T) {
	// The fallback produces a valid analysis with no env/config at all.
	p := NewFallbackProvider()
	got, err := p.AnalyzeLaunch(context.Background(), AdvisorInput{Action: ActionAnalyze, Snapshot: seedSnapshot()})
	if err != nil {
		t.Fatalf("fallback error: %v", err)
	}
	if len(got.Recommendations) == 0 {
		t.Fatal("fallback produced no recommendations for the seeded launch")
	}
	if len(got.SuggestedQuestions) == 0 {
		t.Fatal("fallback produced no suggested questions")
	}
}

func TestFallbackReferencesRealEntities(t *testing.T) {
	snap := seedSnapshot()
	ids := map[string]bool{snap.Launch.ID: true}
	for _, it := range snap.Launch.ChecklistItems {
		ids[it.ID] = true
	}
	for _, r := range snap.Launch.Risks {
		ids[r.ID] = true
	}
	got, _ := NewFallbackProvider().AnalyzeLaunch(context.Background(), AdvisorInput{Action: ActionAnalyze, Snapshot: snap})
	for _, rec := range got.Recommendations {
		for _, e := range rec.Evidence {
			if !ids[e.EntityID] {
				t.Fatalf("recommendation references unknown entity id %q", e.EntityID)
			}
		}
	}
}

func TestFallbackDeterministic(t *testing.T) {
	snap := seedSnapshot()
	a, _ := NewFallbackProvider().AnalyzeLaunch(context.Background(), AdvisorInput{Action: ActionAnalyze, Snapshot: snap})
	b, _ := NewFallbackProvider().AnalyzeLaunch(context.Background(), AdvisorInput{Action: ActionAnalyze, Snapshot: snap})
	if len(a.Recommendations) != len(b.Recommendations) {
		t.Fatalf("non-deterministic recommendation count: %d vs %d", len(a.Recommendations), len(b.Recommendations))
	}
	for i := range a.Recommendations {
		if a.Recommendations[i].Title != b.Recommendations[i].Title ||
			a.Recommendations[i].Type != b.Recommendations[i].Type {
			t.Fatalf("recommendation %d differs between runs", i)
		}
	}
}
