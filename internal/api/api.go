// Package api wires the ShipCheck HTTP routes to the domain core and the
// repository. The same handler runs locally as a plain net/http server and,
// when deployed, behind an API Gateway HTTP API via the Lambda adapter.
//
// Handlers validate input with the pure domain validators (returning HTTP 400
// with field-specific messages), then call repository methods — which persist
// the change and recompute the readiness assessment — and return the updated
// launch as JSON.
package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	httpadapter "github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	"github.com/paulwritescode/shipcheck/internal/advisor"
	"github.com/paulwritescode/shipcheck/internal/domain"
	"github.com/paulwritescode/shipcheck/internal/repo"
)

// Server holds the dependencies the handlers need.
type Server struct {
	repo    repo.LaunchRepository
	advisor *advisor.Advisor
}

// NewRouter builds the HTTP handler for ShipCheck backed by the given
// repository. It accepts the LaunchRepository interface so the in-memory and
// DynamoDB implementations are interchangeable. The advisor provider is chosen
// from configuration (free deterministic fallback by default).
func NewRouter(r repo.LaunchRepository) http.Handler {
	s := &Server{repo: r, advisor: advisor.New(advisor.SelectProvider())}
	mux := http.NewServeMux()

	// Health.
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Launches.
	mux.HandleFunc("POST /api/launches", s.handleCreateLaunch)
	mux.HandleFunc("GET /api/launches/{id}", s.handleGetLaunch)
	mux.HandleFunc("PATCH /api/launches/{id}", s.handleUpdateLaunch)
	mux.HandleFunc("DELETE /api/launches/{id}", s.handleDeleteLaunch)

	// Checklist items.
	mux.HandleFunc("POST /api/launches/{id}/items", s.handleCreateItem)
	mux.HandleFunc("PATCH /api/launches/{id}/items/{itemId}", s.handleUpdateItem)
	mux.HandleFunc("DELETE /api/launches/{id}/items/{itemId}", s.handleDeleteItem)

	// Risks.
	mux.HandleFunc("POST /api/launches/{id}/risks", s.handleCreateRisk)
	mux.HandleFunc("PATCH /api/launches/{id}/risks/{riskId}", s.handleUpdateRisk)
	mux.HandleFunc("DELETE /api/launches/{id}/risks/{riskId}", s.handleDeleteRisk)

	// Sharing (full behavior lands with the public-report task).
	mux.HandleFunc("POST /api/launches/{id}/share", s.handleShare)

	// AI advisor (read-only).
	mux.HandleFunc("POST /api/launches/{id}/analyze", s.handleAnalyze)

	// Public read-only report (no auth), resolved by share token.
	mux.HandleFunc("GET /api/r/{token}", s.handlePublicReport)

	return mux
}

// ServeLambda runs the handler under the AWS Lambda custom runtime, adapting
// API Gateway HTTP API (v2) proxy events to the net/http handler. The same
// handler serves local HTTP and Lambda, so there is one code path.
func ServeLambda(h http.Handler) {
	adapter := httpadapter.NewV2(h)
	lambda.Start(adapter.ProxyWithContext)
}

// ---- shared helpers ----

// writeJSON encodes v as JSON with the given status.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// errorBody is the shape of a non-validation error response.
type errorBody struct {
	Error string `json:"error"`
}

// writeError sends a plain error with the given status.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorBody{Error: msg})
}

// validationBody is the shape of a 400 response listing field errors.
type validationBody struct {
	Error  string                   `json:"error"`
	Fields []domain.ValidationError `json:"fields"`
}

// writeValidation sends HTTP 400 with the collected field errors.
func writeValidation(w http.ResponseWriter, fields []domain.ValidationError) {
	writeJSON(w, http.StatusBadRequest, validationBody{
		Error:  "validation failed",
		Fields: fields,
	})
}

// decodeJSON reads the request body into v, returning false (and writing a 400)
// if the body is not valid JSON.
func decodeJSON(w http.ResponseWriter, req *http.Request, v any) bool {
	if err := json.NewDecoder(req.Body).Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "request body is not valid JSON")
		return false
	}
	return true
}

// writeRepoResult maps a repository (launch, err) result to an HTTP response:
// 200 with the launch on success, 404 on ErrNotFound, 500 otherwise.
func (s *Server) writeRepoResult(w http.ResponseWriter, l domain.Launch, err error) {
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, l)
	case errors.Is(err, repo.ErrNotFound):
		writeError(w, http.StatusNotFound, "launch not found")
	default:
		writeError(w, http.StatusInternalServerError, "the change could not be saved")
	}
}
