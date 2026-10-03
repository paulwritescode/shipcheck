package api

import (
	"net/http"

	"github.com/paulwritescode/shipcheck/internal/domain"
)

// riskRequest is the body for creating or updating a risk.
type riskRequest struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Severity    domain.Severity   `json:"severity"`
	Likelihood  domain.Likelihood `json:"likelihood"`
	Owner       string            `json:"owner"`
	Mitigation  string            `json:"mitigation"`
	Status      domain.RiskStatus `json:"status"`
	DueDate     string            `json:"dueDate"`
	Evidence    []domain.Evidence `json:"evidence"`
}

func (b riskRequest) toRisk(id string) domain.Risk {
	sev := b.Severity
	if sev == "" {
		sev = domain.SeverityLow
	}
	lik := b.Likelihood
	if lik == "" {
		lik = domain.LikelihoodLow
	}
	status := b.Status
	if status == "" {
		status = domain.RiskOpen
	}
	return domain.Risk{
		ID:          id,
		Title:       b.Title,
		Description: b.Description,
		Severity:    sev,
		Likelihood:  lik,
		Owner:       b.Owner,
		Mitigation:  b.Mitigation,
		Status:      status,
		DueDate:     b.DueDate,
		Evidence:    b.Evidence,
	}
}

func validateRisk(r domain.Risk) []domain.ValidationError {
	var fields []domain.ValidationError
	if r.Title == "" {
		fields = append(fields, domain.ValidationError{Field: "title", Message: "Title is required."})
	}
	if !r.Severity.IsValid() {
		fields = append(fields, domain.ValidationError{Field: "severity", Message: "Severity must be Low, Medium, or High."})
	}
	if !r.Likelihood.IsValid() {
		fields = append(fields, domain.ValidationError{Field: "likelihood", Message: "Likelihood must be Low, Medium, or High."})
	}
	if msg := domain.ValidateRiskStatus(r.Status); msg != "" {
		fields = append(fields, domain.ValidationError{Field: "status", Message: msg})
	}
	if r.DueDate != "" && !domain.IsValidCalendarDate(r.DueDate) {
		fields = append(fields, domain.ValidationError{Field: "dueDate", Message: "Due date must be a valid calendar date (YYYY-MM-DD)."})
	}
	fields = append(fields, validateEvidence(r.Evidence)...)
	return fields
}

func (s *Server) handleCreateRisk(w http.ResponseWriter, req *http.Request) {
	var body riskRequest
	if !decodeJSON(w, req, &body) {
		return
	}
	risk := body.toRisk(newChildID("r"))
	if fields := validateRisk(risk); len(fields) > 0 {
		writeValidation(w, fields)
		return
	}
	l, err := s.repo.AddRisk(req.Context(), req.PathValue("id"), risk)
	s.writeRepoResult(w, l, err)
}

func (s *Server) handleUpdateRisk(w http.ResponseWriter, req *http.Request) {
	var body riskRequest
	if !decodeJSON(w, req, &body) {
		return
	}
	risk := body.toRisk(req.PathValue("riskId"))
	if fields := validateRisk(risk); len(fields) > 0 {
		writeValidation(w, fields)
		return
	}
	l, err := s.repo.UpdateRisk(req.Context(), req.PathValue("id"), risk)
	s.writeRepoResult(w, l, err)
}

func (s *Server) handleDeleteRisk(w http.ResponseWriter, req *http.Request) {
	l, err := s.repo.DeleteRisk(req.Context(), req.PathValue("id"), req.PathValue("riskId"))
	s.writeRepoResult(w, l, err)
}
