package api

import (
	"net/http"

	"github.com/paulwritescode/shipcheck/internal/domain"
	"github.com/paulwritescode/shipcheck/internal/repo"
)

// createLaunchRequest is the body for POST /api/launches. When Seed is true the
// handler builds the seeded "Team Inbox 2.0" launch (wired in the seed task)
// and the other fields are ignored.
type createLaunchRequest struct {
	Seed             bool                `json:"seed"`
	Name             string              `json:"name"`
	Description      string              `json:"description"`
	TargetDate       string              `json:"targetDate"`
	Owner            string              `json:"owner"`
	ProductArea      string              `json:"productArea"`
	RepositoryURL    string              `json:"repositoryUrl"`
	DocumentationURL string              `json:"documentationUrl"`
	Brief            *domain.LaunchBrief `json:"brief"`
}

func (s *Server) handleCreateLaunch(w http.ResponseWriter, req *http.Request) {
	var body createLaunchRequest
	if !decodeJSON(w, req, &body) {
		return
	}

	// The seeded demo launch is created by the seed builder (seed task). Until
	// then, return 501 so the contract is explicit rather than silently wrong.
	if body.Seed {
		if seedFn == nil {
			writeError(w, http.StatusNotImplemented, "seeded launch is not available yet")
			return
		}
		l, err := s.repo.Create(req.Context(), seedFn())
		s.writeRepoResult(w, l, err)
		return
	}

	if fields := validateCreate(body); len(fields) > 0 {
		writeValidation(w, fields)
		return
	}

	l, err := s.repo.Create(req.Context(), repo.NewLaunch{
		Name:             body.Name,
		Description:      body.Description,
		TargetDate:       body.TargetDate,
		Owner:            body.Owner,
		ProductArea:      body.ProductArea,
		RepositoryURL:    body.RepositoryURL,
		DocumentationURL: body.DocumentationURL,
		Brief:            body.Brief,
	})
	// A created launch starts at Needs Review until data is added (Req 1.9),
	// which the engine already yields for an empty, owner-only launch.
	if err != nil {
		writeError(w, http.StatusInternalServerError, "the launch could not be created")
		return
	}
	writeJSON(w, http.StatusCreated, l)
}

func validateCreate(body createLaunchRequest) []domain.ValidationError {
	var fields []domain.ValidationError
	if msg := domain.ValidateLaunchName(body.Name); msg != "" {
		fields = append(fields, domain.ValidationError{Field: "name", Message: msg})
	}
	if msg := domain.ValidateDescription(body.Description); msg != "" {
		fields = append(fields, domain.ValidationError{Field: "description", Message: msg})
	}
	if msg := domain.ValidateTargetDate(body.TargetDate); msg != "" {
		fields = append(fields, domain.ValidationError{Field: "targetDate", Message: msg})
	}
	if msg := domain.ValidateURL(body.RepositoryURL); msg != "" {
		fields = append(fields, domain.ValidationError{Field: "repositoryUrl", Message: msg})
	}
	if msg := domain.ValidateURL(body.DocumentationURL); msg != "" {
		fields = append(fields, domain.ValidationError{Field: "documentationUrl", Message: msg})
	}
	return fields
}

func (s *Server) handleGetLaunch(w http.ResponseWriter, req *http.Request) {
	l, err := s.repo.Get(req.Context(), req.PathValue("id"))
	s.writeRepoResult(w, l, err)
}

// updateLaunchRequest patches launch-level fields. Nil pointers mean "leave
// unchanged"; non-nil pointers (including empty strings) are applied.
type updateLaunchRequest struct {
	Name             *string             `json:"name"`
	Description      *string             `json:"description"`
	TargetDate       *string             `json:"targetDate"`
	Owner            *string             `json:"owner"`
	ProductArea      *string             `json:"productArea"`
	RepositoryURL    *string             `json:"repositoryUrl"`
	DocumentationURL *string             `json:"documentationUrl"`
	Brief            *domain.LaunchBrief `json:"brief"`
	PrivateNotes     *string             `json:"privateNotes"`
}

func (s *Server) handleUpdateLaunch(w http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	current, err := s.repo.Get(req.Context(), id)
	if err != nil {
		s.writeRepoResult(w, domain.Launch{}, err)
		return
	}

	var body updateLaunchRequest
	if !decodeJSON(w, req, &body) {
		return
	}

	data := current.LaunchData
	if body.Name != nil {
		data.Name = *body.Name
	}
	if body.Description != nil {
		data.Description = *body.Description
	}
	if body.TargetDate != nil {
		data.TargetDate = *body.TargetDate
	}
	if body.Owner != nil {
		data.Owner = *body.Owner
	}
	if body.ProductArea != nil {
		data.ProductArea = *body.ProductArea
	}
	if body.RepositoryURL != nil {
		data.RepositoryURL = *body.RepositoryURL
	}
	if body.DocumentationURL != nil {
		data.DocumentationURL = *body.DocumentationURL
	}
	if body.Brief != nil {
		data.Brief = body.Brief
	}
	if body.PrivateNotes != nil {
		data.PrivateNotes = *body.PrivateNotes
	}

	// Validate the resulting state's constrained fields.
	var fields []domain.ValidationError
	if msg := domain.ValidateLaunchName(data.Name); msg != "" {
		fields = append(fields, domain.ValidationError{Field: "name", Message: msg})
	}
	if msg := domain.ValidateDescription(data.Description); msg != "" {
		fields = append(fields, domain.ValidationError{Field: "description", Message: msg})
	}
	if msg := domain.ValidateTargetDate(data.TargetDate); msg != "" {
		fields = append(fields, domain.ValidationError{Field: "targetDate", Message: msg})
	}
	if msg := domain.ValidateURL(data.RepositoryURL); msg != "" {
		fields = append(fields, domain.ValidationError{Field: "repositoryUrl", Message: msg})
	}
	if msg := domain.ValidateURL(data.DocumentationURL); msg != "" {
		fields = append(fields, domain.ValidationError{Field: "documentationUrl", Message: msg})
	}
	if len(fields) > 0 {
		writeValidation(w, fields)
		return
	}

	l, err := s.repo.Update(req.Context(), data)
	s.writeRepoResult(w, l, err)
}

func (s *Server) handleDeleteLaunch(w http.ResponseWriter, req *http.Request) {
	err := s.repo.Delete(req.Context(), req.PathValue("id"))
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	default:
		s.writeRepoResult(w, domain.Launch{}, err)
	}
}

// seedFn builds the seeded demo launch. It is injected via SetSeedBuilder so
// the API package does not depend on the seed implementation directly.
var seedFn func() repo.NewLaunch

// SetSeedBuilder registers the function that builds the seeded demo launch,
// enabling POST /api/launches with {"seed": true}.
func SetSeedBuilder(fn func() repo.NewLaunch) {
	seedFn = fn
}
