# Implementation Plan: ShipCheck

## Overview

This plan converts the ShipCheck design into incremental, test-driven coding steps in **Go (latest stable)** on **AWS serverless** — a Go backend on **AWS Lambda** behind **Amazon API Gateway (HTTP API)**, persistence on **Amazon DynamoDB** (AWS SDK for Go v2), a **React SPA** (TypeScript + Vite + Tailwind) served from **Amazon S3 via CloudFront**, and infrastructure provisioned with **AWS CDK authored in Go**.

The build front-loads the pure Go domain core so the **Readiness_Engine** and its **34 correctness properties** (design.md, Properties 1–34) can be verified early with **`pgregory.net/rapid`** — this is the highest-value deliverable and the primary property-based-testing demonstration. Each property test carries the tag `Feature: shipcheck, Property {n}: {text}`, runs a minimum of 100 checks, maps to exactly one design property, and references the requirement clause(s) it validates.

Order: scaffold (Go module + tooling + repo layout) + public-repo hardening (hardened `.gitignore` + secret scanning) → domain types → pure engine + PBT (Props 1–17) → checkpoint → selectors/validators + PBT (Props 18–26) → persistence (DynamoDB repo + in-memory + Property 28) → API handlers (Lambda) with recompute → seeded "Team Inbox 2.0" → checkpoint → React SPA (landing/guest, dashboard, checklist/risk/brief/evidence, four UI states, responsive) → checkpoint → ModelProvider abstraction → advisor schema/validation (Props 29–30) → ShipCheck_Advisor + focused actions + injection prompt (Props 31–32) → MCP read-only external context → safe data-flow guards + redaction/allowlist (Props 33–34) → public share report (Property 27) → checkpoint → CDK deployment/IaC → Kiro University lesson artifacts → final checkpoint.

The repository is a **single repo (not a monorepo)**: a top-level Go module holds the backend and pure core, `web/` holds the SPA, `infra/` holds the CDK app, and `.kiro/` holds specs/steering/hooks/power/MCP/agent config.

Tasks marked with `*` are optional test sub-tasks and can be skipped for a faster MVP; they are the property-based-testing deliverable and should be implemented for full traceability. Core implementation tasks are never optional.

## Tasks

- [ ] 1. Scaffold the repository and harden it for public hosting
  - [ ] 1.1 Scaffold the repository, Go module, and toolchain
    - Initialize a **single repository** (not a monorepo): a top-level **Go module** for the backend and pure core, a `web/` directory for the React SPA, an `infra/` directory for the CDK app, and the `.kiro/` tree for specs/steering/hooks
    - Create the Go package layout: `internal/domain/` (pure core), `internal/repo/` (LaunchRepository + implementations), `internal/api/` (Lambda handlers), `internal/advisor/` (ModelProvider + advisor), `internal/mcp/` (external context), `internal/security/` (guards/redaction), `cmd/` (Lambda `bootstrap` entrypoint)
    - Add `pgregory.net/rapid` and the AWS SDK for Go v2 as module dependencies; add `gofmt`/`go vet`/`go build`/`go test` workflow scripts
    - Scaffold the SPA under `web/` with Vite + TypeScript + Tailwind and Vitest + React Testing Library; add `lint`, `typecheck`, and single-run `test` scripts (no watch mode)
    - _Requirements: 15.3; Design: Architecture → High-Level Structure, Deployment, Repo layout_

  - [ ] 1.2 Create a hardened `.gitignore` for a public repository
    - Exclude AWS credentials, CDK-generated files (`cdk.out/`, `cdk.context.json`), `.env` files, `*.pem` / private keys, and Go/Node build artifacts
    - Confirm `cdk.json`, the CDK app code, and the entire `.kiro/` tree remain tracked (safe/required to commit)
    - _Requirements: 16.2, 16.3, 19.8; Design: Security, Privacy, and Trust Boundaries → Public repository hygiene_

  - [ ] 1.3 Set up secret scanning so the repo is safe to be public from the first commit
    - Add a pre-commit/secret-scan mechanism and/or CI check that scans for credential patterns so no credential is introduced in any commit
    - _Requirements: 16.2, 16.3, 19.8; Design: Security, Privacy, and Trust Boundaries → Public repository hygiene_

- [ ] 2. Define the pure domain data models and enums (Go)
  - [ ] 2.1 Create canonical domain types and enums as Go structs/constants
    - Define `Category`, `CompletionState`, `Priority`, `Severity`, `Likelihood`, `RiskStatus`, `ReadinessStatus`, `CategoryState`
    - Define `Evidence`, `ChecklistItem`, `Risk`, `ShareLink`, `LaunchBrief`, `LaunchData`, `Launch`
    - Define `ReasonEntry`, `CategoryCompletion`, `ReadinessAssessment`
    - Keep this package free of I/O, clock, and randomness (pure core boundary)
    - _Requirements: 3.2, 4.5, 6.2; Design: Data Models_

- [ ] 3. Implement the pure Readiness_Engine (Go)
  - [ ] 3.1 Implement scoring predicates
    - `IsCompleteForScoring(state)` true only for `CompletionComplete`
    - `IsCriticalItem(item)` and `IsBlockingRisk(risk)` (Severity High AND status neither Resolved nor Accepted)
    - _Requirements: 6.2, 6.8; Design: Blocking_Risk and scoring definitions_

  - [ ] 3.2 Implement the per-category completion calculator
    - `ComputeCategoryStates(launch)` returning CompleteCount, TotalCount, and State (complete/partial/empty) for each of the seven categories
    - _Requirements: 7.6_

  - [ ] 3.3 Implement the rule-precedence evaluator
    - `EvaluateRules(launch)` applies the fixed precedence Needs Review > Not Ready > Conditionally Ready > Ready and returns the first matching status with justifying reason entries
    - Encode every Requirement-6 rule (6.4–6.13) and derive stable reason ordering from an identifier-sorted traversal
    - _Requirements: 6.3, 6.4, 6.5, 6.6, 6.7, 6.8, 6.9, 6.10, 6.11, 6.12, 6.13, 6.14, 6.15_

  - [ ] 3.4 Implement reason entries and recommended actions
    - Attach `RuleID`, `Message`, optional `ItemID`, and `RecommendedAction` per the reason rules
    - Ensure Not Ready produces one reason per incomplete critical item and per Blocking_Risk; Needs Review covers absent owner, empty launch, and each complete-without-evidence critical item
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5_

  - [ ] 3.5 Assemble `ComputeReadiness(launch)`
    - Compose predicates, evaluator, reasons, and category states into a total `ReadinessAssessment`; never panic on well-formed input
    - _Requirements: 6.1, 6.13, 6.14; Design: Readiness_Engine_

- [ ] 4. Build custom rapid generators for launch data
  - [ ] 4.1 Build well-formed `LaunchData` generators and tier-targeting generators
    - Generic `LaunchData` rapid generator plus generators that force each readiness tier (empty launch, incomplete-critical, blocking-risk, complete-without-evidence, all-complete) and multi-tier overlap for precedence
    - _Requirements: 6.3; Design: Testing Strategy → Generators_

- [ ] 5. Property-based tests for the Readiness_Engine with rapid (primary PBT deliverable)
  - [ ]* 5.1 Totality, determinism, and scoring
    - **Property 1: Totality — exactly one status for any launch** — Validates: Requirements 6.13, 6.14
    - **Property 2: Determinism and idempotence of assessment** — Validates: Requirements 6.15
    - **Property 3: Completion-state scoring mapping** — Validates: Requirements 6.2
    - Tag each `Feature: shipcheck, Property {n}: {text}`; run rapid with a minimum of 100 checks
  - [ ]* 5.2 Precedence and tier properties
    - **Property 4: Precedence — the highest matching tier wins** — Validates: Requirements 6.3
    - **Property 5: Needs Review on an empty launch** — Validates: Requirements 6.4, 1.9
    - **Property 6: Needs Review on absent owner** — Validates: Requirements 6.5
    - **Property 7: Needs Review on a completion claim without evidence** — Validates: Requirements 6.6
    - **Property 8: Not Ready on an incomplete critical item** — Validates: Requirements 6.7, 6.9
    - **Property 9: Not Ready on a Blocking_Risk, and the Blocking_Risk definition** — Validates: Requirements 6.8
    - **Property 10: Conditionally Ready when criticals satisfied but a gap/accepted risk remains** — Validates: Requirements 6.10, 6.11
    - **Property 11: Ready when everything is complete and evidenced** — Validates: Requirements 6.12
  - [ ]* 5.3 Reason and category properties
    - **Property 12: Every assessment carries at least one rule-identified reason** — Validates: Requirements 7.1
    - **Property 13: Reason item-id association is correct** — Validates: Requirements 7.2
    - **Property 14: Not Ready reasons correspond one-to-one to their causes** — Validates: Requirements 7.3
    - **Property 15: Needs Review reasons cover every triggering condition** — Validates: Requirements 7.4
    - **Property 16: Actionable reasons carry a targeted recommended action** — Validates: Requirements 7.5
    - **Property 17: Per-category completion counts and classification are correct** — Validates: Requirements 7.6

- [ ] 6. Checkpoint - Ensure engine and property tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 7. Implement dashboard/selector pure functions and their properties (Go)
  - [ ] 7.1 Implement blocker selector, risk ordering, filters, grouping, and edit helpers
    - Critical-blocker selector (incomplete criticals + Blocking_Risks), High-severity-first risk ordering, incomplete-only and high-priority filters, group-by-category, single-field edit preserving all other fields, high-emphasis risk predicate
    - _Requirements: 8.3, 8.4, 3.11, 3.12, 3.13, 3.9, 4.7, 4.6_
  - [ ]* 7.2 Property tests (rapid) for selectors and helpers
    - **Property 18: Blocker selection equals incomplete criticals plus blocking risks** — Validates: Requirements 8.3
    - **Property 19: Risk ordering places High severity first** — Validates: Requirements 8.4
    - **Property 20: Filter correctness** — Validates: Requirements 3.11, 3.12
    - **Property 21: Category grouping partitions without loss and each group is homogeneous** — Validates: Requirements 3.13
    - **Property 22: Editing one field preserves all other fields** — Validates: Requirements 3.9, 4.7
    - **Property 26: High-emphasis predicate** — Validates: Requirements 4.6

- [ ] 8. Implement the Go validation package
  - [ ] 8.1 Implement validators with field-specific messages
    - Launch name (1–200, required/whitespace/too-long), description length, target-date calendar validity, URL syntax (repository/documentation/evidence), category enum, priority enum, risk-status enum
    - _Requirements: 1.2, 1.3, 1.4, 1.5, 1.8, 3.2, 3.4, 4.2, 4.3, 4.5, 5.3_
  - [ ]* 8.2 Property tests (rapid) for validators
    - **Property 23: URL validation accepts valid URLs and rejects invalid ones** — Validates: Requirements 1.8, 5.3
    - **Property 24: Name and enum validation** — Validates: Requirements 1.2, 1.3, 3.2, 4.5
    - **Property 25: Calendar-date validation** — Validates: Requirements 1.5

- [ ] 9. Implement persistence (DynamoDB + LaunchRepository)
  - [ ] 9.1 Define the `LaunchRepository` Go interface and `InMemoryLaunchRepository`
    - Interface: `Create`, `Get`, `GetByShareToken`, `Update`, `AddChecklistItem`, and analogous risk/evidence methods; persist the computed assessment on every mutation; map to/from domain types
    - In-memory implementation used by unit, integration, and property tests
    - _Requirements: 13.1, 13.2, 10.8_
  - [ ] 9.2 Implement `DynamoDBLaunchRepository` (AWS SDK for Go v2)
    - Single-table design keyed by launch id storing the launch aggregate (checklist items, risks, evidence, brief, share link, stored assessment) via attribute-value marshaling; a GSI maps share token → launch id for `GetByShareToken`; no relational migrations
    - _Requirements: 13.1, 13.2, 10.8; Design: Launch Repository → DynamoDB table design_
  - [ ]* 9.3 Property test (rapid) for persistence round-trip against the in-memory repo
    - **Property 28: Persistence round-trip** — Validates: Requirements 13.2
  - [ ]* 9.4 Integration test for the DynamoDB repository and failure handling
    - Exercise `DynamoDBLaunchRepository` against DynamoDB Local (or a real table); assert round-trip and that a persistence failure surfaces an error state and discards the in-memory mutation
    - _Requirements: 13.1, 13.2, 13.3_

- [ ] 10. Implement API route handlers on Lambda with recompute-on-mutation (Go)
  - [ ] 10.1 Launch create/edit handlers
    - `POST /api/launches` (create or seed) and `PATCH /api/launches/:id` (brief/owner/details); validate input, persist, then call `ComputeReadiness` and return the launch with a fresh assessment; new launches start at Needs Review
    - _Requirements: 1.1, 1.6, 1.7, 1.9, 2.1, 2.3, 6.1_
  - [ ] 10.2 Checklist item handlers
    - `POST /api/launches/:id/items`, `PATCH .../items/:itemId`, `DELETE .../items/:itemId` for create/edit/complete/flag/delete, plus evidence attach; recompute after each mutation
    - _Requirements: 3.1, 3.3, 3.5, 3.6, 3.7, 3.8, 3.9, 3.10, 6.1_
  - [ ] 10.3 Risk and evidence handlers
    - `POST /api/launches/:id/risks`, `PATCH/DELETE .../risks/:riskId`, and evidence attach/remove routes; recompute after each mutation
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.8, 5.1, 5.2, 5.4, 5.5, 6.1_
  - [ ] 10.4 Wire the Lambda `bootstrap` entrypoint and API Gateway routing
    - Single Go Lambda handler dispatching the HTTP API routes to the handlers above; JSON request/response; structured HTTP 400 validation-error bodies with field messages
    - _Requirements: 1.1, 6.1; Design: API Surface, Deployment_
  - [ ]* 10.5 Unit/integration tests for CRUD and recompute
    - Assert create/edit/delete behavior, validation-error responses (HTTP 400 with field messages), and that mutations trigger recompute
    - _Requirements: 1.1–1.8, 2, 3.1–3.10, 4.1–4.8, 5.1–5.5, 6.1_

- [ ] 11. Implement the Seeded "Team Inbox 2.0" launch builder (pure Go)
  - [ ] 11.1 Build the seed as a pure function
    - Name "Team Inbox 2.0", target 2026-10-16; 18 items (13 complete, 3 in progress, 2 blocked); 4 risks with exactly one High severity; at least one unowned item; an incomplete critical rollback-documentation item in Launch Operations; a public preview URL; evidence referencing example GitHub issues/PRs
    - Wire seed creation into `POST /api/launches`
    - _Requirements: 12.1, 12.2, 12.3, 12.4, 12.5, 12.6, 1.10_
  - [ ]* 11.2 Example tests for seed shape and readiness transitions
    - Assert the fixed shape, that the initial state evaluates to Not Ready, and that the documented demo edits move it to Conditionally Ready
    - _Requirements: 12.2–12.6, 12.7, 12.8_

- [ ] 12. Checkpoint - Ensure persistence, API, and seed tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 13. Implement the landing page and guest exploration (React SPA)
  - [ ] 13.1 Build the landing page and demo entry point
    - Describe product purpose; provide an entry point that opens the Seeded_Launch without account creation; allow guest read access to dashboard, checklist, risk register, and readiness assessment; SPA calls the API Gateway endpoints over HTTPS
    - _Requirements: 11.1, 11.2, 11.3_

- [ ] 14. Implement the readiness dashboard (React SPA)
  - [ ] 14.1 Build the dashboard view
    - Show title, target date, overall status, and overall progress; per-category breakdown; current critical blockers; risks ordered High-severity-first; recommended next actions; blocker detail (reason + recommended action) on open; empty-state prompt when no items and no risks
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5, 8.6, 8.7_

- [ ] 15. Implement checklist, risk register, brief, and evidence views (React SPA)
  - [ ] 15.1 Build the checklist view
    - Grouped by category; create/edit/complete/flag/delete; owner, priority, critical flag, evidence; incomplete-only and high-priority filters
    - _Requirements: 3.1, 3.3, 3.4, 3.5, 3.6, 3.7, 3.8, 3.9, 3.10, 3.11, 3.12, 3.13_
  - [ ] 15.2 Build the risk register and launch brief views
    - Risk register with severity/likelihood/owner/mitigation/status/due date and high-emphasis treatment; launch brief editor with empty-state prompt
    - _Requirements: 4.1–4.8, 2.1, 2.2, 2.3, 2.4_
  - [ ] 15.3 Build the evidence UI
    - Attach/remove URL and note evidence on items and risks with URL validation messaging
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_

- [ ] 16. Implement responsive layout and the four UI states (React SPA)
  - [ ] 16.1 Add loading, empty, no-results, and error states across views; ensure responsive desktop/mobile layout
    - _Requirements: 14.1, 14.2, 14.3, 14.4, 14.5_
  - [ ]* 16.2 Component tests (Vitest + React Testing Library) for UI states and responsive rendering
    - Component/UI-state tests only — not the property suite
    - _Requirements: 14.1–14.5_

- [ ] 17. Checkpoint - Ensure SPA views and UI-state tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 18. Implement the ModelProvider abstraction (Go)
  - [ ] 18.1 Define `ModelProvider` and concrete providers
    - `ModelProvider` Go interface (`AnalyzeLaunch`); the finalized provider lineup, all behind the interface and selectable by config:
      - **`FallbackProvider`** (default) — free, deterministic, rule-based, derived from the engine assessment; **no external model, no paid key**; the **always-free default the live/public demo runs on**
      - **`HuggingFaceModelProvider`** (opt-in for the demo video) — a small hosted SLM such as **Llama 3.2 1B** or **Qwen2.5 1.5B** via the **Hugging Face Inference API** using a **server-side API key** never exposed to client code; note the HF free serverless allowance is **rate-limited / non-production**
      - **`BedrockModelProvider`** (alternative opt-in) — **Amazon Nova Micro** via the **AWS SDK for Go v2**, authenticated by the **Lambda execution role's IAM permissions (no API key)** with least-privilege `bedrock:InvokeModel`; **credit-limited (no free-inference tier)**
      - **`NvidiaModelProvider`** (alternative opt-in) — NVIDIA model **catalog** at build.nvidia.com over the OpenAI-compatible NIM endpoint, model selectable by a configurable model-name field
      - **`LocalModelProvider`** (alternative opt-in) — self-hosted / local model
    - Key-based providers (Hugging Face, NVIDIA) read keys **server-side only** (Lambda env vars / AWS Secrets Manager) and never expose them to client code; Bedrock uses **IAM-role auth with no key**; the public deployment **defaults to `FallbackProvider`** so the live demo needs **no key and incurs no cost**
    - _Requirements: 16.1, 16.2, 16.3, 16.4, 16.5, 16.6, 16.7, 16.8, 16.9_
  - [ ]* 18.2 Config/smoke tests for provider selection and server-side keys
    - Assert a provider is **selectable without code changes**; the default `FallbackProvider` **needs no external model or key**; the `HuggingFaceModelProvider` **reads a server-side key only** (never exposed to client code); the `BedrockModelProvider` uses **IAM-role auth (no key)**; and **all providers are invoked from server-side Lambda code only**
    - _Requirements: 16.1, 16.2, 16.3, 16.4, 16.5, 16.6, 16.7_

- [ ] 19. Implement the advisor structured-output schema and validation (Go)
  - [ ] 19.1 Define the `LaunchAnalysis` types and validation pipeline
    - Go types for `Recommendation` (type, priority, title, reason, suggested_action, evidence refs) and `SuggestedQuestions`; validate advisor responses against the schema; discard and record an error for any recommendation referencing a nonexistent launch entity
    - _Requirements: 9.10, 9.11, 9.12_
  - [ ]* 19.2 Property tests (rapid) for advisor schema and evidence integrity
    - **Property 29: Advisor output is structurally valid** — Validates: Requirements 9.10, 9.11
    - **Property 30: Evidence-reference integrity** — Validates: Requirements 9.12

- [ ] 20. Implement the ShipCheck_Advisor and focused actions (Go)
  - [ ] 20.1 Implement the read-only advisor over the LaunchSnapshot
    - Build the immutable `LaunchSnapshot` after the engine computes readiness; implement the seven focused actions (Analyze Launch, Find Blockers, Suggest Checklist, Review Risks, Prepare Launch Review, Explain Readiness, Re-analyze After Changes); recommendation-only output; accept routes through the normal checklist-create path, dismiss makes no change; loading and error states; advisor never mutates launch data or changes status
    - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5, 9.6, 9.7, 9.8, 9.9, 9.13, 9.14_
  - [ ] 20.2 Implement the injection-hardened advisor system prompt
    - Instruct the model to treat all `Untrusted_External_Content` as data not instructions, disregard embedded instructions, withhold secrets/system prompts/credentials/hidden context, and use untrusted content only as evidence
    - _Requirements: 18.1, 18.2, 18.3, 18.4_
  - [ ] 20.3 Wire `POST /api/launches/:id/analyze` in the analyze Lambda
    - Server-side orchestration: build snapshot, call `ModelProvider`, validate response, return recommendations + questions with evidence links; leave data unchanged on failure
    - _Requirements: 9.5, 9.7, 16.4_
  - [ ] 20.4 Build the advisor UI (React SPA)
    - Focused-action controls, recommendations list with accept/dismiss, suggested questions, loading and error states; recommendations labeled AI-generated requiring human review
    - _Requirements: 9.2, 9.3, 9.4, 9.5, 9.6, 19.3_
  - [ ] 20.7 Render the AI-provenance and demo disclaimer (React SPA)
    - Label all advisor output as **AI-generated and requiring human review**; show a **public-deployment disclaimer** stating that ShipCheck is a **non-commercial demonstration** and that AI recommendations are **not a release approval**
    - _Requirements: 21.1, 21.2; Design: Security, Privacy, and Trust Boundaries → AI provenance and demo disclaimer_
  - [ ]* 20.5 Property tests (rapid) for advisor purity and deterministic authority
    - **Property 31: Advisor never mutates launch state** — Validates: Requirements 9.8
    - **Property 32: Deterministic authority — advisor output cannot change the score** — Validates: Requirements 9.7, 9.8, 15.2
  - [ ]* 20.6 Stub/fallback-provider unit tests for advisor behavior
    - Recommendation-only and non-mutating (9.2–9.4, 9.8); failure error state (9.5); loading state (9.6); per-action result shapes (9.9, 9.13, 9.14); system-prompt directives and untrusted-content tagging (18.1–18.4)
    - _Requirements: 9.2, 9.3, 9.4, 9.5, 9.6, 9.8, 9.9, 9.13, 9.14, 18.1, 18.2, 18.3, 18.4_

- [ ] 21. Implement MCP read-only external context (Go, server-side)
  - [ ] 21.1 Implement the read-only ShipCheck MCP tools
    - Expose `get_public_repository_summary`, `get_open_issues`, `get_recent_pull_requests`, `get_release_notes`, `get_documentation_page`, `inspect_public_github_reference`; all read-only; user confirmation required before any tool that would change data
    - _Requirements: 17.1, 17.2, 17.3_
  - [ ] 21.2 Implement app-managed MCP connection and tool bridging in the analyze Lambda
    - The Go application manages the connection and executes tools; convert MCP tool definitions to the model tool format and return results to the model as data; never provide external context to the Readiness_Engine
    - _Requirements: 17.4, 17.5, 17.8_

- [ ] 22. Implement safe data-flow guards and redaction (Go)
  - [ ] 22.1 Implement the analysis-request pipeline, allowlist guard, and snapshot redaction
    - Load only the selected launch; strip unnecessary PII; prefer structured data; external fetches limited to allowlisted public URLs; label external content Untrusted_External_Content; redact prohibited fields (keys/tokens/passwords, private repo content, unnecessary personal/customer/health/financial/identity info) from the snapshot; require user confirmation before saving changes; rate limiting, abuse-reporting path, log retention limit, launch-data deletion
    - _Requirements: 19.1, 19.2, 19.4, 19.5, 19.6, 19.8, 19.9, 19.10, 19.11, 17.6, 17.7_
  - [ ]* 22.2 Property tests (rapid) for allowlist and redaction
    - **Property 33: External-fetch allowlist enforcement** — Validates: Requirements 17.6, 19.1
    - **Property 34: Prohibited-data redaction in the model snapshot** — Validates: Requirements 19.2
  - [ ]* 22.3 Integration test for MCP tool bridging
    - Assert the app executes a read-only MCP tool, converts tool definitions to model format, and returns results, with the model never connecting to MCP directly
    - _Requirements: 17.4, 17.5_

- [ ] 23. Implement the public read-only share report
  - [ ] 23.1 Implement share tokens and the public projection endpoint (Go)
    - `POST /api/launches/:id/share` to enable/regenerate/disable with a cryptographically random opaque token; `GET /api/r/:token` returns a no-auth public projection resolved via the DynamoDB GSI; projection excludes private notes and evidence marked private and reports the stored assessment status; regenerated/disabled tokens return an unavailable state
    - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5, 10.6, 10.7, 10.8_
  - [ ] 23.2 Build the public report SPA route (React)
    - `GET /r/:token` client route that fetches `GET /api/r/:token` and renders the public report, including the unavailable-message state
    - _Requirements: 10.2, 10.3, 10.7, 11.1_
  - [ ]* 23.3 Property test (rapid) for public-report projection
    - **Property 27: Public report projection preserves status and excludes private data** — Validates: Requirements 10.4, 10.8
  - [ ]* 23.4 Integration test for the share flow
    - Enable, resolve, regenerate, disable, unavailable-message paths; guest access to the seeded launch
    - _Requirements: 10.1, 10.2, 10.3, 10.5, 10.6, 10.7, 11.1, 11.2, 11.3_

- [ ] 24. Checkpoint - Ensure advisor, MCP, security, and share tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 25. Implement deployment and infrastructure as code (AWS CDK in Go)
  - [ ] 25.1 Author the CDK stack in `infra/` (Go)
    - Define the Go Lambda(s) on the `provided.al2023` custom runtime with a `bootstrap` entrypoint (pre-built binary, no Docker; `go build` for `linux/arm64` or `linux/amd64`), the API Gateway HTTP API, the DynamoDB table + GSI, the S3 bucket, and the CloudFront distribution (default certificate serves HTTPS) as one CDK stack
    - Grant the Lambda execution role **least-privilege `bedrock:InvokeModel` ONLY when the `BedrockModelProvider` is enabled**; store any **Hugging Face / NVIDIA key as a Lambda env var / AWS Secrets Manager secret** (never committed to the repo); the **public deployment defaults to the `FallbackProvider`** so **no provider secret is required for the live link**
    - _Requirements: 19.7, 19.8, 16.2, 16.7, 16.8; Design: Deployment; Security, Privacy, and Trust Boundaries → Public repository hygiene_
  - [ ] 25.2 Add the build-and-deploy scripts
    - Scripts to `go build` the Lambda binary, build the SPA static assets, and run `cdk deploy` (CDK CLI via Node.js as a build-time-only dependency); note always-free-tier sizing
    - _Requirements: 15.3; Design: Deployment_

- [ ] 26. Deliver Kiro University lesson artifacts
  - [ ] 26.1 Author steering files (`.kiro/steering/`)
    - `architecture.md` (pure Go core / I/O-boundary separation; engine has no I/O, clock, or randomness), `ux.md` (four UI states + responsive SPA), `security.md` (public-report privacy, opaque tokens, keys server-side in Lambda env / Secrets Manager; **no secrets in the repo — all provider keys only in Lambda env vars / AWS Secrets Manager; hardened `.gitignore` and secret scanning; the `.kiro/` tree and CDK code are intentionally public**), `testing.md` (mandatory PBT with `pgregory.net/rapid` tagged to design numbers, min 100 checks), `advisor.md` (read-only, non-authoritative, post-assessment snapshot, schema-validated JSON), `prompt-injection.md` (untrusted content as data, redaction) — reflecting Go/AWS conventions
    - _Design: Kiro Lessons Mapping → Planned steering files; Security, Privacy, and Trust Boundaries → Public repository hygiene_
  - [ ] 26.2 Author hooks (`.kiro/hooks/`)
    - `go vet`/`gofmt`/`go build` + `go test` on save for `*.go`; run readiness property tests on domain-core / property-test change; SPA lint & type-check on save for `*.ts`/`*.tsx`; spec-consistency reminder on edits to `requirements.md` or `design.md`; **a secret-scanning hook that scans staged files on save/commit for credential patterns**
    - _Design: Kiro Lessons Mapping → Planned hooks; Security, Privacy, and Trust Boundaries → Public repository hygiene_
  - [ ] 26.3 Package the reusable ShipCheck power
    - Bundle readiness-review steering plus a launch-review workflow guide for reuse across repositories
    - _Design: Kiro Lessons Mapping → ShipCheck power_
  - [ ] 26.4 Add MCP server configuration
    - Configuration registering the read-only ShipCheck MCP server for advisor external context (app-managed, server-side in the Go analyze Lambda)
    - _Design: Kiro Lessons Mapping → MCP servers; Requirements 17.1, 17.4, 17.5_
  - [ ] 26.5 Add the custom-agent configuration for the ShipCheck_Advisor
    - Configure the advisor as a custom agent with a limited read-only toolset and strict permissions (never mutate launch data, never change status, never approve/deny), realizing the planning/risk-review/QA-review sub-actions
    - _Design: Kiro Lessons Mapping → Custom agents; Requirements 9.8, 9.9_
  - [ ] 26.6 Write the README lessons section and model disclosure
    - Document how each of the seven lessons is demonstrated with file references
    - Document the **full provider set** — `FallbackProvider` (default), `HuggingFaceModelProvider`, `BedrockModelProvider`, `NvidiaModelProvider`, `LocalModelProvider` — and, per Req 21.3, **disclose the specific model and provider used for any demonstration** plus its applicable non-production/trial/credit terms: **HF free serverless = rate-limited, non-production**; **Bedrock = credit-limited, no free-inference tier**; **NVIDIA API trial = dev/testing only, not production** (and each catalog model's own license — e.g. DeepSeek, Llama — additionally applies if self-hosted)
    - State that the **live public deployment runs the free deterministic `FallbackProvider`** (no external model, no key, no cost); record the exact model name, provider, license, and access terms in the repository; add a privacy notice for data sent to the model provider
    - _Requirements: 20.4, 20.5, 20.3, 21.3_

- [ ] 27. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional test sub-tasks and can be skipped for a faster MVP; they are the property-based-testing deliverable and should be implemented for full traceability.
- Each task references specific requirement clauses and/or design properties for traceability.
- Property tests each map to exactly one design property, run a minimum of 100 checks with `pgregory.net/rapid` (automatic shrinking), and carry the tag `Feature: shipcheck, Property {n}: {text}`.
- Checkpoints ensure incremental validation; property tests validate universal correctness while unit/integration tests cover CRUD, UI states, persistence, and sharing. SPA tests use Vitest + React Testing Library for component/UI-state coverage only, never the property suite.
- The Readiness_Engine stays pure Go (no I/O, clock, or randomness); all security/privacy concerns live at the application boundary (the Lambda handlers, provider path, and external fetches).
- Property 28 targets the `LaunchRepository` interface via the `InMemoryLaunchRepository`; the `DynamoDBLaunchRepository` is validated by an integration test (DynamoDB Local or a real table).

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1"] },
    { "id": 1, "tasks": ["1.2", "1.3", "2.1"] },
    { "id": 2, "tasks": ["3.1", "3.2"] },
    { "id": 3, "tasks": ["3.3"] },
    { "id": 4, "tasks": ["3.4", "3.5"] },
    { "id": 5, "tasks": ["4.1"] },
    { "id": 6, "tasks": ["5.1", "5.2", "5.3", "7.1", "8.1"] },
    { "id": 7, "tasks": ["7.2", "8.2", "9.1"] },
    { "id": 8, "tasks": ["9.2", "9.3"] },
    { "id": 9, "tasks": ["9.4", "10.1", "10.2", "10.3"] },
    { "id": 10, "tasks": ["10.4", "11.1"] },
    { "id": 11, "tasks": ["10.5", "11.2", "13.1", "14.1", "15.1", "15.2", "15.3"] },
    { "id": 12, "tasks": ["16.1"] },
    { "id": 13, "tasks": ["16.2", "18.1"] },
    { "id": 14, "tasks": ["18.2", "19.1"] },
    { "id": 15, "tasks": ["19.2", "20.1", "20.2"] },
    { "id": 16, "tasks": ["20.3", "20.4", "20.7", "21.1"] },
    { "id": 17, "tasks": ["20.5", "20.6", "21.2", "22.1"] },
    { "id": 18, "tasks": ["22.2", "22.3", "23.1"] },
    { "id": 19, "tasks": ["23.2", "23.3", "23.4", "25.1"] },
    { "id": 20, "tasks": ["25.2", "26.1", "26.2", "26.3", "26.4", "26.5"] },
    { "id": 21, "tasks": ["26.6"] }
  ]
}
```
