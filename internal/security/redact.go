package security

import "github.com/paulwritescode/shipcheck/internal/domain"

// RedactForModel returns a copy of the launch with prohibited data removed, so
// that nothing sensitive can reach a model provider (Req 19.1, 19.2). It is a
// pure function and never mutates the input.
//
// Redaction removes:
//   - private launch notes (LaunchData.PrivateNotes),
//   - every evidence entry marked private (Evidence.IsPrivate), from both
//     checklist items and risks.
//
// The structured launch data that remains (titles, categories, statuses,
// owners, public evidence) is what the advisor reasons over. Secrets never live
// in these fields by design — keys/tokens are server-side configuration only —
// so there are no secret-bearing fields to strip here; redaction is about the
// user-controlled private content that must not leave the workspace.
func RedactForModel(l domain.Launch) domain.Launch {
	out := l // copies scalars + the Assessment value

	// Strip private launch notes.
	out.PrivateNotes = ""

	// Copy checklist items, dropping private evidence.
	if l.ChecklistItems != nil {
		out.ChecklistItems = make([]domain.ChecklistItem, len(l.ChecklistItems))
		for i, it := range l.ChecklistItems {
			it.Evidence = publicEvidence(it.Evidence)
			out.ChecklistItems[i] = it
		}
	}

	// Copy risks, dropping private evidence.
	if l.Risks != nil {
		out.Risks = make([]domain.Risk, len(l.Risks))
		for i, r := range l.Risks {
			r.Evidence = publicEvidence(r.Evidence)
			out.Risks[i] = r
		}
	}

	// Share link is not sent to the model.
	out.Share = nil

	return out
}

// publicEvidence returns only the non-private evidence entries. Returns nil
// when there is no public evidence so the shape stays clean.
func publicEvidence(ev []domain.Evidence) []domain.Evidence {
	var out []domain.Evidence
	for _, e := range ev {
		if !e.IsPrivate {
			out = append(out, e)
		}
	}
	return out
}

// ContainsProhibited reports whether a launch still carries data that must not
// reach the model: private notes or any private evidence. Used by tests to
// assert redaction is complete.
func ContainsProhibited(l domain.Launch) bool {
	if l.PrivateNotes != "" {
		return true
	}
	for _, it := range l.ChecklistItems {
		for _, e := range it.Evidence {
			if e.IsPrivate {
				return true
			}
		}
	}
	for _, r := range l.Risks {
		for _, e := range r.Evidence {
			if e.IsPrivate {
				return true
			}
		}
	}
	return false
}
