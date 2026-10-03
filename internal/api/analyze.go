package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/paulwritescode/shipcheck/internal/advisor"
)

// analyzeRequest selects which focused advisor action to run.
type analyzeRequest struct {
	Action string `json:"action"`
}

// handleAnalyze runs the read-only advisor against a launch and returns its
// validated recommendations. It never mutates the launch: it loads the launch
// (which already carries its engine-computed assessment), builds a snapshot,
// and runs the advisor. Accepting a suggested item is a separate, explicit
// mutation the client makes through the normal checklist-create path.
func (s *Server) handleAnalyze(w http.ResponseWriter, req *http.Request) {
	launch, err := s.repo.Get(req.Context(), req.PathValue("id"))
	if err != nil {
		s.writeRepoResult(w, launch, err)
		return
	}

	var body analyzeRequest
	// An empty body is allowed; default to the full "analyze" action.
	_ = decodeOptionalJSON(req, &body)
	action := advisor.Action(body.Action)
	if action == "" {
		action = advisor.ActionAnalyze
	}

	result, err := s.advisor.Analyze(req.Context(), launch, action)
	if err != nil {
		// Advisor failure leaves launch data unchanged (Req 9.5).
		writeError(w, http.StatusBadGateway, "the analysis did not complete")
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// decodeOptionalJSON decodes the body into v if present, tolerating an empty
// body (EOF). Returns an error only on malformed non-empty JSON.
func decodeOptionalJSON(req *http.Request, v any) error {
	if req.Body == nil {
		return nil
	}
	err := json.NewDecoder(req.Body).Decode(v)
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}
