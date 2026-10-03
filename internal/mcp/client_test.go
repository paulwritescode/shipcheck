package mcp

import (
	"context"
	"errors"
	"testing"

	"github.com/paulwritescode/shipcheck/internal/security"
)

// fakeFetcher records the URL it was asked to fetch and returns a canned
// response, so tests never touch the network and can assert the allowlist
// blocks a fetch before it happens.
type fakeFetcher struct {
	called  bool
	lastURL string
	status  int
	body    string
	err     error
}

func (f *fakeFetcher) Get(_ context.Context, url string) (int, string, error) {
	f.called = true
	f.lastURL = url
	if f.err != nil {
		return 0, "", f.err
	}
	status := f.status
	if status == 0 {
		status = 200
	}
	return status, f.body, nil
}

func TestToModelToolsConversion(t *testing.T) {
	tools := ToModelTools(Definitions())
	if len(tools) != len(Definitions()) {
		t.Fatalf("want %d model tools, got %d", len(Definitions()), len(tools))
	}
	for _, mt := range tools {
		if mt.Type != "function" {
			t.Errorf("tool type: want function, got %q", mt.Type)
		}
		if mt.Function.Name == "" {
			t.Error("tool function has no name")
		}
		if _, ok := mt.Function.Parameters["type"]; !ok {
			t.Error("tool parameters missing JSON-schema type")
		}
	}
}

func TestAllToolsAreReadOnly(t *testing.T) {
	for _, d := range Definitions() {
		if !d.ReadOnly {
			t.Errorf("tool %q is not read-only", d.Name)
		}
	}
}

func TestExecuteRefusesOffAllowlistURL(t *testing.T) {
	f := &fakeFetcher{body: "secret"}
	c := NewClientWith(security.DefaultAllowlist(), f)
	// evil.example.com is not on the allowlist; the fetch must be refused
	// BEFORE any network call.
	_, err := c.Execute(context.Background(), ToolDocumentation, "https://evil.example.com/x")
	if !errors.Is(err, ErrNotAllowed) {
		t.Fatalf("want ErrNotAllowed, got %v", err)
	}
	if f.called {
		t.Fatal("fetcher was called for an off-allowlist URL")
	}
}

func TestExecuteRefusesNonHTTPS(t *testing.T) {
	f := &fakeFetcher{}
	c := NewClientWith(security.DefaultAllowlist(), f)
	_, err := c.Execute(context.Background(), ToolGitHubReference, "http://github.com/x/y/issues/1")
	if !errors.Is(err, ErrNotAllowed) {
		t.Fatalf("non-https: want ErrNotAllowed, got %v", err)
	}
	if f.called {
		t.Fatal("fetcher called for non-https URL")
	}
}

func TestExecuteRepoSummaryFetchesAllowlistedAPI(t *testing.T) {
	f := &fakeFetcher{body: `{"name":"team-inbox","description":"demo"}`}
	c := NewClientWith(security.DefaultAllowlist(), f)
	item, err := c.Execute(context.Background(), ToolRepoSummary, "https://github.com/example/team-inbox")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if f.lastURL != "https://api.github.com/repos/example/team-inbox" {
		t.Fatalf("derived API URL wrong: %q", f.lastURL)
	}
	if item.Source != "github-repo" || item.Content == "" {
		t.Fatalf("external context item malformed: %+v", item)
	}
	// Content is returned as data (untrusted reference material).
	if item.URL != f.lastURL {
		t.Fatalf("item URL mismatch: %q", item.URL)
	}
}

func TestGithubAPIBase(t *testing.T) {
	cases := map[string]string{
		"https://github.com/example/team-inbox":        "https://api.github.com/repos/example/team-inbox",
		"https://github.com/example/team-inbox.git":    "https://api.github.com/repos/example/team-inbox",
		"https://github.com/example/team-inbox/issues": "https://api.github.com/repos/example/team-inbox",
	}
	for in, want := range cases {
		got, err := githubAPIBase(in)
		if err != nil {
			t.Fatalf("%s: %v", in, err)
		}
		if got != want {
			t.Fatalf("%s: got %q want %q", in, got, want)
		}
	}
	if _, err := githubAPIBase("https://gitlab.com/x/y"); err == nil {
		t.Fatal("non-github URL should error")
	}
}

func TestAllowlistPermits(t *testing.T) {
	a := security.DefaultAllowlist()
	if !a.Permits("https://github.com/x/y") {
		t.Error("github.com should be permitted")
	}
	if a.Permits("https://evil.example.com") {
		t.Error("evil.example.com should be refused")
	}
	if a.Permits("http://github.com/x") {
		t.Error("non-https should be refused")
	}
	if a.Permits("not a url") {
		t.Error("non-URL should be refused")
	}
}
