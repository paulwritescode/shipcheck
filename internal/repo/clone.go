package repo

import "github.com/paulwritescode/shipcheck/internal/domain"

// cloneLaunchData returns a deep copy of launch data so stored state and
// returned values never share mutable slices or pointers. Keeping clones at
// the repository boundary is what lets callers edit returned values freely and
// guarantees the persistence round-trip property.
func cloneLaunchData(d domain.LaunchData) domain.LaunchData {
	out := d // copies scalar fields

	out.ChecklistItems = cloneItems(d.ChecklistItems)
	out.Risks = cloneRisks(d.Risks)

	if d.Brief != nil {
		b := *d.Brief
		out.Brief = &b
	}
	if d.Share != nil {
		s := *d.Share
		out.Share = &s
	}
	return out
}

func cloneItems(items []domain.ChecklistItem) []domain.ChecklistItem {
	if items == nil {
		return nil
	}
	out := make([]domain.ChecklistItem, len(items))
	for i, it := range items {
		it.Evidence = cloneEvidence(it.Evidence)
		out[i] = it
	}
	return out
}

func cloneRisks(risks []domain.Risk) []domain.Risk {
	if risks == nil {
		return nil
	}
	out := make([]domain.Risk, len(risks))
	for i, r := range risks {
		r.Evidence = cloneEvidence(r.Evidence)
		out[i] = r
	}
	return out
}

func cloneEvidence(ev []domain.Evidence) []domain.Evidence {
	if ev == nil {
		return nil
	}
	out := make([]domain.Evidence, len(ev))
	copy(out, ev)
	return out
}
