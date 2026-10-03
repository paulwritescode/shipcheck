package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/paulwritescode/shipcheck/internal/domain"
	"github.com/paulwritescode/shipcheck/internal/repo"
)

// handlePublicReport serves the no-auth public read-only report for a share
// token (Req 10.2). Only an enabled share link resolves; a disabled,
// regenerated, or unknown token yields an "unavailable" 404 (Req 10.7), never
// leaking launch data.
func (s *Server) handlePublicReport(w http.ResponseWriter, req *http.Request) {
	token := req.PathValue("token")
	launch, err := s.repo.GetByShareToken(req.Context(), token)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": "This report is unavailable. The link may have been disabled or regenerated.",
			})
			return
		}
		writeError(w, http.StatusInternalServerError, "could not load the report")
		return
	}

	report := domain.ProjectPublicReport(launch, time.Now().UTC().Format(time.RFC3339))
	writeJSON(w, http.StatusOK, report)
}
