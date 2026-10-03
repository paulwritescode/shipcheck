// Package mcp provides the read-only ShipCheck MCP tools and the app-managed
// bridge that executes them. Per the design (and NVIDIA's NIM tool-calling
// model), the APPLICATION — not the model — manages the connection and runs
// tools; it converts tool definitions to the model's tool format and returns
// results to the model as DATA. External context enriches the advisor only and
// is NEVER provided to the Readiness_Engine (Req 17.4, 17.5, 17.8).
//
// Every tool is read-only in v1 (Req 17.2). Any tool that would change GitHub
// or ShipCheck data would require explicit user confirmation (Req 17.3); none
// exists in v1.
package mcp

// ToolName identifies a read-only external-context tool.
type ToolName string

const (
	ToolRepoSummary     ToolName = "get_public_repository_summary"
	ToolOpenIssues      ToolName = "get_open_issues"
	ToolRecentPRs       ToolName = "get_recent_pull_requests"
	ToolReleaseNotes    ToolName = "get_release_notes"
	ToolDocumentation   ToolName = "get_documentation_page"
	ToolGitHubReference ToolName = "inspect_public_github_reference"
)

// ToolParameter describes one input parameter of a tool, in a shape that maps
// cleanly to the model's tool-call parameter schema.
type ToolParameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // "string"
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// ToolDefinition is the app-side description of a read-only tool. It is
// converted to the model's tool format by ToModelTools; the model never sees
// the implementation, only this definition.
type ToolDefinition struct {
	Name        ToolName        `json:"name"`
	Description string          `json:"description"`
	ReadOnly    bool            `json:"readOnly"`
	Parameters  []ToolParameter `json:"parameters"`
}

// Definitions returns the fixed set of read-only tool definitions (Req 17.1).
func Definitions() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        ToolRepoSummary,
			Description: "Summarize a public GitHub repository (name, description, default branch).",
			ReadOnly:    true,
			Parameters:  []ToolParameter{{Name: "repoUrl", Type: "string", Description: "Public repository URL", Required: true}},
		},
		{
			Name:        ToolOpenIssues,
			Description: "List open issues for a public GitHub repository.",
			ReadOnly:    true,
			Parameters:  []ToolParameter{{Name: "repoUrl", Type: "string", Description: "Public repository URL", Required: true}},
		},
		{
			Name:        ToolRecentPRs,
			Description: "List recent pull requests for a public GitHub repository.",
			ReadOnly:    true,
			Parameters:  []ToolParameter{{Name: "repoUrl", Type: "string", Description: "Public repository URL", Required: true}},
		},
		{
			Name:        ToolReleaseNotes,
			Description: "Fetch published release notes for a public GitHub repository.",
			ReadOnly:    true,
			Parameters:  []ToolParameter{{Name: "repoUrl", Type: "string", Description: "Public repository URL", Required: true}},
		},
		{
			Name:        ToolDocumentation,
			Description: "Fetch an allowlisted public documentation page.",
			ReadOnly:    true,
			Parameters:  []ToolParameter{{Name: "url", Type: "string", Description: "Allowlisted public URL", Required: true}},
		},
		{
			Name:        ToolGitHubReference,
			Description: "Read a public GitHub reference (issue or pull request URL).",
			ReadOnly:    true,
			Parameters:  []ToolParameter{{Name: "url", Type: "string", Description: "Public GitHub issue/PR URL", Required: true}},
		},
	}
}

// ModelTool is a tool definition in the OpenAI-compatible "function" tool
// format the model consumes.
type ModelTool struct {
	Type     string            `json:"type"` // "function"
	Function ModelToolFunction `json:"function"`
}

// ModelToolFunction is the function portion of a model tool.
type ModelToolFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"` // JSON Schema
}

// ToModelTools converts the app-side tool definitions to the model's tool
// format (Req 17.5). The application performs this conversion; the model only
// ever receives these shapes and never connects to any external system.
func ToModelTools(defs []ToolDefinition) []ModelTool {
	out := make([]ModelTool, 0, len(defs))
	for _, d := range defs {
		props := map[string]any{}
		var required []string
		for _, p := range d.Parameters {
			props[p.Name] = map[string]any{"type": p.Type, "description": p.Description}
			if p.Required {
				required = append(required, p.Name)
			}
		}
		out = append(out, ModelTool{
			Type: "function",
			Function: ModelToolFunction{
				Name:        string(d.Name),
				Description: d.Description,
				Parameters: map[string]any{
					"type":       "object",
					"properties": props,
					"required":   required,
				},
			},
		})
	}
	return out
}
