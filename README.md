# ShipCheck

**ShipCheck is a launch-readiness workspace for small product and engineering
teams.** It answers one question before you release a product, feature, or major
update: *are we actually ready to ship?*

Readiness is not a vibe. In ShipCheck it is a **pure, deterministic function of
launch data** — your checklist and your risk register — evaluated by a rule
engine that is the sole authority for the verdict. An AI advisor sits alongside
to explain the status and recommend next steps, but it never decides, approves,
or changes anything.

> **Non-commercial demonstration.** ShipCheck is a portfolio/learning project
> built for the Kiro University Challenge. It is not a commercial product, and
> its AI recommendations are **not a release approval**. See
> [Model & AI disclosure](#model--ai-disclosure) and [Privacy](#privacy).

Repository: https://github.com/paulwritescode/shipcheck

---

## What it does

- **Deterministic readiness engine.** A pure Go core computes one of four
  statuses — *Needs Review*, *Not Ready*, *Conditionally Ready*, *Ready* — using
  a fixed precedence and produces a reason for every blocker.
- **Checklist across seven categories** with owners, priorities, critical flags,
  and evidence; **risk register** with severity/likelihood/mitigation/status.
- **AI advisor** that explains the status and recommends actions — read-only and
  non-authoritative by construction.
- **Public read-only share report** behind an opaque token, excluding private
  notes and private evidence.
- **Seeded "Team Inbox 2.0" demo launch** so you can explore without an account.
- Runs **entirely locally with no AWS connection**, and deploys to
  **always-free AWS** serverless when you want a live link.

## Architecture

| Layer | Tech | Notes |
| --- | --- | --- |
| Pure core | Go (`internal/domain`) | Readiness engine, selectors, validators, seed. **No I/O, clock, or randomness.** |
| Persistence | `LaunchRepository` interface (`internal/repo`) | In-memory for local/tests; DynamoDB for deploy. Recomputes the assessment on every mutation. |
| API | Go `net/http` + Lambda (`internal/api`, `cmd/`) | Same handlers local and on Lambda. |
| Advisor / MCP / security | `internal/advisor`, `internal/mcp`, `internal/security` | AI advisor, read-only external context, trust-boundary guards. |
| Frontend | React 18 + TypeScript + Vite + Tailwind (`web/`) | SPA; four UI states on every data view; responsive. |
| Infra | AWS CDK in Go (`infra/`, separate module) | Lambda `provided.al2023` arm64 + API Gateway HTTP API + DynamoDB + S3/CloudFront. |

The guiding rule — the deterministic engine decides, the AI explains — is
documented in `.kiro/steering/architecture.md` and `.kiro/steering/advisor.md`.

## Run it locally

```bash
# Backend (Go 1.27+). Serves the API and pre-seeds the demo launch.
go run ./cmd/shipcheck
# -> listens on :8080 (override with SHIPCHECK_ADDR)

# Frontend (Node 18+), in a second terminal:
cd web && npm install && npm run dev
```

No AWS account or key is required to run or demo ShipCheck locally — the default
AI provider is the free, deterministic `FallbackProvider`.

## Verify it

```bash
make ci          # gofmt check + go vet + build + test + secret-scan
make test-pbt    # the property-based test deliverable (verbose)

# Run the 34 correctness properties hard:
go test ./internal/domain/ -run Property -rapid.checks=1000

# Frontend checks:
cd web && npm test && npm run typecheck && npm run lint && npm run build
```

## Deploy to always-free AWS (optional)

```bash
make lambda      # cross-compile the Lambda bootstrap (linux/arm64, provided.al2023)
make deploy      # build SPA + cdk deploy (needs AWS creds + CDK CLI)
```

Sized for the AWS always-free tier (Lambda, API Gateway HTTP API, DynamoDB 25 GB,
S3 + CloudFront). The live deployment defaults to the `FallbackProvider`, so the
public link needs **no provider key and incurs no model cost**.

---

## Kiro University lessons mapping

ShipCheck was built spec-first in Kiro and deliberately produces a concrete,
reviewable artifact for each of the seven lessons.

| # | Lesson | Where it lives | What it demonstrates |
| --- | --- | --- | --- |
| 1 | **Spec-driven development** | `.kiro/specs/shipcheck/` — [requirements.md](.kiro/specs/shipcheck/requirements.md), [design.md](.kiro/specs/shipcheck/design.md), [tasks.md](.kiro/specs/shipcheck/tasks.md) | The whole feature flows requirements → design → tasks. The 34 correctness properties are defined in the design and traced to requirements. |
| 2 | **Steering** | [`.kiro/steering/`](.kiro/steering/) — architecture, ux, security, testing, advisor, prompt-injection | Persistent guidance the agent follows across the build (pure-core boundary, four UI states, no-secrets policy, mandatory PBT, read-only advisor, untrusted-content handling). |
| 3 | **Hooks** | [`.kiro/hooks/`](.kiro/hooks/) | Go vet/fmt/build/test on save; readiness property tests on domain-core change (`-rapid.checks=1000`); SPA lint + type-check on save; secret-scan on save; spec-consistency reminder. |
| 4 | **Property-based testing** | `internal/**/..._property_test.go` with [`pgregory.net/rapid`](https://pkg.go.dev/pgregory.net/rapid) | The **34 correctness properties** against the pure engine, selectors, repo round-trip, advisor, and security guards. Each test is tagged `Feature: shipcheck, Property N` and runs ≥100 checks. This is the highest-value lesson. |
| 5 | **Powers** | [`.kiro/powers/shipcheck/`](.kiro/powers/shipcheck/) | A reusable ShipCheck power bundling readiness-review steering + a launch-review workflow guide for other repositories. |
| 6 | **MCP** | [`.kiro/mcp/`](.kiro/mcp/) config + `internal/mcp/` | Read-only external context (repo summaries, issues, PRs, release notes, docs) for the advisor only — never the engine. In production the Go analyze Lambda manages the connection and the model never connects to MCP directly. |
| 7 | **Custom agents** | [`.kiro/agents/shipcheck-advisor.json`](.kiro/agents/shipcheck-advisor.json) | The ShipCheck Advisor as a custom agent with a limited read-only toolset and strict permissions — it explains and recommends but never mutates launch data or changes status. Realizes the planning, risk-review, and QA-review sub-actions. |

### Property-based testing map

| Properties | File |
| --- | --- |
| 1–17 (engine) | `internal/domain/engine_property_test.go` |
| 18–26 (selectors/validators) | `internal/domain/selectors_property_test.go` |
| 27 (public-report projection) | `internal/domain/report_property_test.go` |
| 28 (persistence round-trip) | `internal/repo/repo_property_test.go` |
| 29–30 (advisor schema / evidence integrity) | `internal/advisor/validate_property_test.go` |
| 31–32 (advisor purity / deterministic authority) | `internal/advisor/advisor_property_test.go` |
| 33–34 (allowlist / redaction) | `internal/security/security_property_test.go` |

---

## Model & AI disclosure

Per the challenge's disclosure requirement, here is the full AI provider picture.
All model providers sit behind the `ModelProvider` interface
(`internal/advisor/`) and are selectable by configuration
(`SHIPCHECK_MODEL_PROVIDER`) without code changes. Keys are read **server-side
only** (Lambda env / AWS Secrets Manager) and never exposed to client code.

| Provider | Model | Auth | Terms / caveats |
| --- | --- | --- | --- |
| `FallbackProvider` **(default)** | none — deterministic, rule-based, derived from the engine assessment | none | Free. No external model, no key, no cost. **This is what the live/public demo runs.** |
| `HuggingFaceModelProvider` (opt-in) | a small hosted SLM (e.g. Llama 3.2 1B / Qwen2.5 1.5B) via the HF Inference API | server-side API key | HF free serverless allowance is **rate-limited and non-production**. |
| `BedrockModelProvider` (opt-in) | Amazon Nova Micro via AWS SDK for Go v2 | **IAM role** (least-privilege `bedrock:InvokeModel`) — no key | **Credit-limited; no free-inference tier.** |
| `NvidiaModelProvider` (opt-in) | NVIDIA model catalog (build.nvidia.com, OpenAI-compatible NIM), model name configurable | server-side API key | NVIDIA API trial is **dev/testing only, not production**. Each catalog model's own license (e.g. DeepSeek, Llama) additionally applies if self-hosted. |
| `LocalModelProvider` (opt-in) | self-hosted / local model | local | For local development. |

**For any live demonstration using a non-default provider, the specific model,
provider, license, and access terms must be recorded alongside the demo.** The
public deployment itself runs the free deterministic `FallbackProvider`.

All advisor output in the UI is labeled **AI-generated and requires human
review**, and the public deployment shows a disclaimer that ShipCheck is a
non-commercial demonstration and that AI recommendations are not a release
approval.

## Privacy

- **No secrets in this repo.** Provider keys live only in Lambda env vars / AWS
  Secrets Manager; Bedrock uses IAM-role auth with no key. A hardened
  `.gitignore` and `scripts/secret-scan.sh` (run in CI and on save) keep
  credentials out of every commit.
- **Public share reports** are served behind cryptographically random opaque
  tokens and **exclude private notes and private evidence**; regenerating or
  disabling a token immediately invalidates the old link.
- **Before any model call**, `security.RedactForModel` strips prohibited data
  (keys/tokens/passwords, private repo content, private notes/evidence, share
  links, and unnecessary personal/customer/health/financial/identity data) from
  the snapshot. External fetches are limited to an allowlist of public hosts and
  treated as untrusted data, never instructions.
- **Data sent to a model provider:** when a non-default (opt-in) provider is
  configured, the redacted launch snapshot is sent to that provider for the
  requested analysis. With the default `FallbackProvider`, **no data leaves the
  application** — the analysis is computed locally from the engine assessment.

## License

See [LICENSE](LICENSE).
