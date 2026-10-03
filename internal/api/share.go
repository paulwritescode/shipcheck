package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/paulwritescode/shipcheck/internal/domain"
)

// shareRequest controls the share link. Action is one of "enable",
// "regenerate", or "disable". The public read-only report that the token
// resolves to is completed in the public-report task.
type shareRequest struct {
	Action string `json:"action"`
}

func (s *Server) handleShare(w http.ResponseWriter, req *http.Request) {
	var body shareRequest
	if !decodeJSON(w, req, &body) {
		return
	}
	id := req.PathValue("id")

	var link *domain.ShareLink
	switch body.Action {
	case "enable", "regenerate":
		// A fresh token on both enable and regenerate makes the previous link
		// stop resolving (Req 10.1, 10.5).
		link = &domain.ShareLink{Token: newShareToken(), Enabled: true}
	case "disable":
		link = &domain.ShareLink{Token: newShareToken(), Enabled: false}
	default:
		writeError(w, http.StatusBadRequest, `action must be "enable", "regenerate", or "disable"`)
		return
	}

	l, err := s.repo.SetShareLink(req.Context(), id, link)
	s.writeRepoResult(w, l, err)
}

// newShareToken returns a cryptographically random, opaque, URL-safe token.
func newShareToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// newChildID returns a short random id for a checklist item or risk, prefixed
// so ids are easy to tell apart in logs and the UI.
func newChildID(prefix string) string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return prefix + "_" + hex.EncodeToString(b)
}
