package api

import (
	"net/http"

	"github.com/paulwritescode/shipcheck/internal/domain"
)

// itemRequest is the body for creating or updating a checklist item. Evidence
// is managed as part of the item (attach/remove by sending the desired list).
type itemRequest struct {
	ID              string                 `json:"id"`
	Title           string                 `json:"title"`
	Category        domain.Category        `json:"category"`
	Owner           string                 `json:"owner"`
	Priority        domain.Priority        `json:"priority"`
	CompletionState domain.CompletionState `json:"completionState"`
	IsCritical      bool                   `json:"isCritical"`
	Evidence        []domain.Evidence      `json:"evidence"`
}

func (b itemRequest) toItem(id string) domain.ChecklistItem {
	priority := b.Priority
	if priority == "" {
		priority = domain.PriorityMedium
	}
	state := b.CompletionState
	if state == "" {
		state = domain.CompletionIncomplete
	}
	return domain.ChecklistItem{
		ID:              id,
		Title:           b.Title,
		Category:        b.Category,
		Owner:           b.Owner,
		Priority:        priority,
		CompletionState: state,
		IsCritical:      b.IsCritical,
		Evidence:        b.Evidence,
	}
}

// validateItem checks an item's constrained fields and evidence URLs.
func validateItem(it domain.ChecklistItem) []domain.ValidationError {
	var fields []domain.ValidationError
	if it.Title == "" {
		fields = append(fields, domain.ValidationError{Field: "title", Message: "Title is required."})
	}
	if msg := domain.ValidateCategory(it.Category); msg != "" {
		fields = append(fields, domain.ValidationError{Field: "category", Message: msg})
	}
	if msg := domain.ValidatePriority(it.Priority); msg != "" {
		fields = append(fields, domain.ValidationError{Field: "priority", Message: msg})
	}
	if !it.CompletionState.IsValid() {
		fields = append(fields, domain.ValidationError{Field: "completionState", Message: "Completion state is not valid."})
	}
	fields = append(fields, validateEvidence(it.Evidence)...)
	return fields
}

// validateEvidence checks each evidence entry's URL syntax.
func validateEvidence(ev []domain.Evidence) []domain.ValidationError {
	var fields []domain.ValidationError
	for i, e := range ev {
		if msg := domain.ValidateURL(e.URL); msg != "" {
			fields = append(fields, domain.ValidationError{
				Field:   "evidence",
				Message: msg,
			})
			_ = i
		}
	}
	return fields
}

func (s *Server) handleCreateItem(w http.ResponseWriter, req *http.Request) {
	var body itemRequest
	if !decodeJSON(w, req, &body) {
		return
	}
	item := body.toItem(newChildID("i"))
	if fields := validateItem(item); len(fields) > 0 {
		writeValidation(w, fields)
		return
	}
	l, err := s.repo.AddChecklistItem(req.Context(), req.PathValue("id"), item)
	s.writeRepoResult(w, l, err)
}

func (s *Server) handleUpdateItem(w http.ResponseWriter, req *http.Request) {
	var body itemRequest
	if !decodeJSON(w, req, &body) {
		return
	}
	item := body.toItem(req.PathValue("itemId"))
	if fields := validateItem(item); len(fields) > 0 {
		writeValidation(w, fields)
		return
	}
	l, err := s.repo.UpdateChecklistItem(req.Context(), req.PathValue("id"), item)
	s.writeRepoResult(w, l, err)
}

func (s *Server) handleDeleteItem(w http.ResponseWriter, req *http.Request) {
	l, err := s.repo.DeleteChecklistItem(req.Context(), req.PathValue("id"), req.PathValue("itemId"))
	s.writeRepoResult(w, l, err)
}
