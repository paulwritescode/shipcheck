package repo

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"

	"github.com/paulwritescode/shipcheck/internal/domain"
)

// InMemoryLaunchRepository stores launches in process memory. It requires no
// database and no AWS account, so the whole app runs locally and the tests run
// without external services. It deep-copies on every read and write so callers
// can never mutate stored state through shared slices.
type InMemoryLaunchRepository struct {
	mu       sync.RWMutex
	launches map[string]domain.LaunchData
}

// NewInMemoryLaunchRepository returns an empty in-memory repository.
func NewInMemoryLaunchRepository() *InMemoryLaunchRepository {
	return &InMemoryLaunchRepository{launches: make(map[string]domain.LaunchData)}
}

// compile-time check that the in-memory store satisfies the interface.
var _ LaunchRepository = (*InMemoryLaunchRepository)(nil)

func (r *InMemoryLaunchRepository) Create(_ context.Context, input NewLaunch) (domain.Launch, error) {
	id := input.ID
	if id == "" {
		id = newID()
	}
	data := domain.LaunchData{
		ID:               id,
		Name:             input.Name,
		Description:      input.Description,
		TargetDate:       input.TargetDate,
		Owner:            input.Owner,
		ProductArea:      input.ProductArea,
		RepositoryURL:    input.RepositoryURL,
		DocumentationURL: input.DocumentationURL,
		Brief:            input.Brief,
		ChecklistItems:   input.ChecklistItems,
		Risks:            input.Risks,
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.launches[id] = cloneLaunchData(data)
	return assess(cloneLaunchData(data)), nil
}

func (r *InMemoryLaunchRepository) Get(_ context.Context, id string) (domain.Launch, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	data, ok := r.launches[id]
	if !ok {
		return domain.Launch{}, ErrNotFound
	}
	return assess(cloneLaunchData(data)), nil
}

func (r *InMemoryLaunchRepository) GetByShareToken(_ context.Context, token string) (domain.Launch, error) {
	if token == "" {
		return domain.Launch{}, ErrNotFound
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, data := range r.launches {
		if data.Share != nil && data.Share.Enabled && data.Share.Token == token {
			return assess(cloneLaunchData(data)), nil
		}
	}
	return domain.Launch{}, ErrNotFound
}

func (r *InMemoryLaunchRepository) Update(_ context.Context, data domain.LaunchData) (domain.Launch, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.launches[data.ID]; !ok {
		return domain.Launch{}, ErrNotFound
	}
	r.launches[data.ID] = cloneLaunchData(data)
	return assess(cloneLaunchData(data)), nil
}

func (r *InMemoryLaunchRepository) AddChecklistItem(_ context.Context, launchID string, item domain.ChecklistItem) (domain.Launch, error) {
	return r.mutate(launchID, func(d *domain.LaunchData) {
		d.ChecklistItems = append(d.ChecklistItems, item)
	})
}

func (r *InMemoryLaunchRepository) UpdateChecklistItem(_ context.Context, launchID string, item domain.ChecklistItem) (domain.Launch, error) {
	return r.mutate(launchID, func(d *domain.LaunchData) {
		for i := range d.ChecklistItems {
			if d.ChecklistItems[i].ID == item.ID {
				d.ChecklistItems[i] = item
				return
			}
		}
	})
}

func (r *InMemoryLaunchRepository) DeleteChecklistItem(_ context.Context, launchID, itemID string) (domain.Launch, error) {
	return r.mutate(launchID, func(d *domain.LaunchData) {
		out := d.ChecklistItems[:0]
		for _, it := range d.ChecklistItems {
			if it.ID != itemID {
				out = append(out, it)
			}
		}
		d.ChecklistItems = out
	})
}

func (r *InMemoryLaunchRepository) AddRisk(_ context.Context, launchID string, risk domain.Risk) (domain.Launch, error) {
	return r.mutate(launchID, func(d *domain.LaunchData) {
		d.Risks = append(d.Risks, risk)
	})
}

func (r *InMemoryLaunchRepository) UpdateRisk(_ context.Context, launchID string, risk domain.Risk) (domain.Launch, error) {
	return r.mutate(launchID, func(d *domain.LaunchData) {
		for i := range d.Risks {
			if d.Risks[i].ID == risk.ID {
				d.Risks[i] = risk
				return
			}
		}
	})
}

func (r *InMemoryLaunchRepository) DeleteRisk(_ context.Context, launchID, riskID string) (domain.Launch, error) {
	return r.mutate(launchID, func(d *domain.LaunchData) {
		out := d.Risks[:0]
		for _, rk := range d.Risks {
			if rk.ID != riskID {
				out = append(out, rk)
			}
		}
		d.Risks = out
	})
}

func (r *InMemoryLaunchRepository) SetShareLink(_ context.Context, launchID string, link *domain.ShareLink) (domain.Launch, error) {
	return r.mutate(launchID, func(d *domain.LaunchData) {
		d.Share = link
	})
}

func (r *InMemoryLaunchRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.launches[id]; !ok {
		return ErrNotFound
	}
	delete(r.launches, id)
	return nil
}

// mutate loads a stored launch, applies fn to a working copy, stores the
// result, and returns the launch with a freshly computed assessment.
func (r *InMemoryLaunchRepository) mutate(launchID string, fn func(*domain.LaunchData)) (domain.Launch, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored, ok := r.launches[launchID]
	if !ok {
		return domain.Launch{}, ErrNotFound
	}
	working := cloneLaunchData(stored)
	fn(&working)
	r.launches[launchID] = cloneLaunchData(working)
	return assess(cloneLaunchData(working)), nil
}

// newID returns a short random hex id for a launch.
func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
