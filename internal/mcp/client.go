package mcp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/paulwritescode/shipcheck/internal/advisor"
	"github.com/paulwritescode/shipcheck/internal/security"
)

// ErrNotAllowed is returned when a tool is asked to fetch a URL that is not on
// the public allowlist (Req 17.6).
var ErrNotAllowed = errors.New("url not on the public allowlist")

// Fetcher performs a read-only HTTP GET. It is an interface so the client is
// testable without network access.
type Fetcher interface {
	Get(ctx context.Context, url string) (status int, body string, err error)
}

// httpFetcher is the production Fetcher. It performs plain read-only GETs with
// a short timeout and a capped body size.
type httpFetcher struct {
	client  *http.Client
	maxSize int64
}

func newHTTPFetcher() *httpFetcher {
	return &httpFetcher{client: &http.Client{Timeout: 15 * time.Second}, maxSize: 64 * 1024}
}

func (f *httpFetcher) Get(ctx context.Context, rawURL string) (int, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json, text/plain, */*")
	resp, err := f.client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer func() { _ = resp.Body.Close() }()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, f.maxSize))
	return resp.StatusCode, string(b), nil
}

// Client is the app-managed MCP client. It owns the connection, enforces the
// allowlist, and executes read-only tools, returning results as labeled,
// untrusted external-context items (never fed to the engine).
type Client struct {
	allow   *security.Allowlist
	fetcher Fetcher
}

// NewClient builds a client with the default public allowlist and HTTP fetcher.
func NewClient() *Client {
	return &Client{allow: security.DefaultAllowlist(), fetcher: newHTTPFetcher()}
}

// NewClientWith builds a client with an explicit allowlist and fetcher (tests).
func NewClientWith(allow *security.Allowlist, fetcher Fetcher) *Client {
	return &Client{allow: allow, fetcher: fetcher}
}

// Execute runs a read-only tool with the given argument URL and returns the
// result as an external-context item. The URL must pass the allowlist; any
// off-allowlist URL is refused before any network call (Req 17.6, 17.7).
func (c *Client) Execute(ctx context.Context, tool ToolName, targetURL string) (advisor.ExternalContextItem, error) {
	fetchURL, source, err := resolve(tool, targetURL)
	if err != nil {
		return advisor.ExternalContextItem{}, err
	}
	if !c.allow.Permits(fetchURL) {
		return advisor.ExternalContextItem{}, ErrNotAllowed
	}

	status, body, err := c.fetcher.Get(ctx, fetchURL)
	if err != nil {
		return advisor.ExternalContextItem{}, err
	}
	if status < 200 || status >= 300 {
		return advisor.ExternalContextItem{}, fmt.Errorf("%s: fetch returned status %d", tool, status)
	}

	// The content is UNTRUSTED reference data; it is labeled as such and is
	// only ever handed to the advisor, never to the engine (Req 17.8, 18.1).
	return advisor.ExternalContextItem{Source: source, URL: fetchURL, Content: body}, nil
}

// resolve maps a tool + a user-provided repo/doc URL to the concrete public
// URL to fetch and a source label. For GitHub tools it derives the public API
// URL from the repository URL; the resulting URL is still allowlist-checked.
func resolve(tool ToolName, target string) (fetchURL, source string, err error) {
	switch tool {
	case ToolRepoSummary:
		api, e := githubAPIBase(target)
		return api, "github-repo", e
	case ToolOpenIssues:
		api, e := githubAPIBase(target)
		return api + "/issues?state=open", "github-issues", e
	case ToolRecentPRs:
		api, e := githubAPIBase(target)
		return api + "/pulls?state=all&per_page=10", "github-pulls", e
	case ToolReleaseNotes:
		api, e := githubAPIBase(target)
		return api + "/releases", "github-releases", e
	case ToolDocumentation:
		return target, "documentation", nil
	case ToolGitHubReference:
		return target, "github-reference", nil
	default:
		return "", "", fmt.Errorf("unknown tool %q", tool)
	}
}
