package mcp

import (
	"fmt"
	"net/url"
	"strings"
)

// githubAPIBase converts a public GitHub repository URL
// (https://github.com/owner/repo[/...]) into the public REST API base
// (https://api.github.com/repos/owner/repo). The result is still subject to the
// allowlist check in Client.Execute.
func githubAPIBase(repoURL string) (string, error) {
	u, err := url.Parse(repoURL)
	if err != nil {
		return "", err
	}
	if strings.ToLower(u.Hostname()) != "github.com" {
		return "", fmt.Errorf("not a github.com repository URL: %q", repoURL)
	}
	parts := strings.FieldsFunc(u.Path, func(r rune) bool { return r == '/' })
	if len(parts) < 2 {
		return "", fmt.Errorf("repository URL must include owner and repo: %q", repoURL)
	}
	owner, repo := parts[0], strings.TrimSuffix(parts[1], ".git")
	return "https://api.github.com/repos/" + owner + "/" + repo, nil
}
