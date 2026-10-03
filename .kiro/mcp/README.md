# MCP configuration — ShipCheck read-only external context

This directory holds the MCP configuration artifact for ShipCheck. To activate
it in Kiro, copy `mcp.json` here into your workspace MCP settings at
`.kiro/settings/mcp.json` (or merge it into your user-level
`~/.kiro/settings/mcp.json`). It is kept here, tracked and reviewable, because
the Kiro settings directory is intentionally not written to automatically.

## Two ways the same read-only context is used

ShipCheck's advisor can enrich its recommendations with **read-only external
context**: public repository summaries, open issues, recent pull requests,
release notes, documentation pages, and public GitHub references. There are two
delivery paths, by design:

1. **In production (app-managed, server-side).** The Go analyze Lambda owns the
   MCP connection, converts tool definitions to the model's tool format,
   executes the read-only tools, and returns results to the model **as data**.
   The model never connects to MCP directly, and external context never reaches
   the deterministic Readiness_Engine (Req 17.4, 17.5, 17.8). The tool set lives
   in `internal/mcp/tools.go`:
   - `get_public_repository_summary`
   - `get_open_issues`
   - `get_recent_pull_requests`
   - `get_release_notes`
   - `get_documentation_page`
   - `inspect_public_github_reference`

2. **In the Kiro IDE (`mcp.json` here).** This config registers a read-only
   external-context server so the same class of context is available while
   developing with Kiro. It is **disabled by default** because the live/public
   demo runs the deterministic `FallbackProvider` and needs no external context,
   and because enabling it spawns a local process (`uvx mcp-server-fetch`).

## Enabling it

Copy `mcp.json` into `.kiro/settings/mcp.json`, then set `"disabled": false`.
The server runs via `uvx`, which requires `uv`
(see https://docs.astral.sh/uv/getting-started/installation/). `uvx` downloads
and runs the server on first use — there is no separate install step.

## Trust boundary

- All tools are **read-only** (Req 17.2). Any tool that would change data would
  require explicit user confirmation (Req 17.3); none exists in v1.
- External fetches are restricted to the allowlist in
  `internal/security/allowlist.go` (public GitHub hosts over HTTPS).
- All external content is treated as `Untrusted_External_Content` — data for
  evidence, never instructions (see `.kiro/steering/prompt-injection.md`).
