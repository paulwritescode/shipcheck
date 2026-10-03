// Package repo defines the LaunchRepository persistence boundary and its
// implementations. The interface keeps the pure domain core and the API layer
// independent of the storage engine: tests and local development use the
// in-memory implementation, while deployment uses DynamoDB.
//
// Every mutating method recomputes and stores the readiness assessment via
// domain.ComputeReadiness, so the dashboard and the public report always read
// a status consistent with the launch data (design: recompute-on-mutation).
package repo

import (
	"context"
	"errors"

	"github.com/paulwritescode/shipcheck/internal/domain"
)

// ErrNotFound is returned when a launch (or share token) does not resolve.
var ErrNotFound = errors.New("launch not found")

// NewLaunch is the input for creating a launch. The repository assigns the
// stored id (callers may supply one; empty means the repository generates it)
// and computes the initial assessment.
type NewLaunch struct {
	ID               string
	Name             string
	Description      string
	TargetDate       string
	Owner            string
	ProductArea      string
	RepositoryURL    string
	DocumentationURL string
	Brief            *domain.LaunchBrief
	// Optional pre-populated children, used by the seeded demo launch.
	ChecklistItems []domain.ChecklistItem
	Risks          []domain.Risk
}

// LaunchRepository is the persistence boundary for launches. Implementations
// must recompute and store the readiness assessment on every mutation.
type LaunchRepository interface {
	// Create persists a new launch and returns it with a computed assessment.
	Create(ctx context.Context, input NewLaunch) (domain.Launch, error)

	// Get returns the launch by id, or ErrNotFound.
	Get(ctx context.Context, id string) (domain.Launch, error)

	// GetByShareToken returns the launch whose enabled share link matches the
	// token, or ErrNotFound when no enabled link resolves.
	GetByShareToken(ctx context.Context, token string) (domain.Launch, error)

	// Update replaces the stored launch data and recomputes the assessment.
	Update(ctx context.Context, data domain.LaunchData) (domain.Launch, error)

	// AddChecklistItem appends an item and recomputes the assessment.
	AddChecklistItem(ctx context.Context, launchID string, item domain.ChecklistItem) (domain.Launch, error)
	// UpdateChecklistItem replaces an item by id and recomputes.
	UpdateChecklistItem(ctx context.Context, launchID string, item domain.ChecklistItem) (domain.Launch, error)
	// DeleteChecklistItem removes an item by id and recomputes.
	DeleteChecklistItem(ctx context.Context, launchID, itemID string) (domain.Launch, error)

	// AddRisk appends a risk and recomputes.
	AddRisk(ctx context.Context, launchID string, risk domain.Risk) (domain.Launch, error)
	// UpdateRisk replaces a risk by id and recomputes.
	UpdateRisk(ctx context.Context, launchID string, risk domain.Risk) (domain.Launch, error)
	// DeleteRisk removes a risk by id and recomputes.
	DeleteRisk(ctx context.Context, launchID, riskID string) (domain.Launch, error)

	// SetShareLink stores (or clears) the share link and recomputes.
	SetShareLink(ctx context.Context, launchID string, link *domain.ShareLink) (domain.Launch, error)

	// Delete removes a launch entirely.
	Delete(ctx context.Context, id string) error
}

// assess attaches a freshly computed assessment to launch data, producing a
// stored Launch. This is the single place the engine is invoked for
// persistence, so every implementation stays consistent.
func assess(data domain.LaunchData) domain.Launch {
	return domain.Launch{
		LaunchData: data,
		Assessment: domain.ComputeReadiness(data),
	}
}
