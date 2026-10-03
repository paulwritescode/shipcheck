# Security and privacy guidance

ShipCheck is a public, non-commercial demonstration hosted on always-free AWS.
It is safe to run in the open because secrets never enter the repo and the
public surfaces leak nothing private.

## No secrets in the repository

- Provider API keys (Hugging Face, NVIDIA) live **only** in Lambda environment
  variables or AWS Secrets Manager, read server-side at runtime. They are never
  committed, never sent to client code, and never logged.
- Bedrock uses the Lambda execution role's IAM permissions
  (least-privilege `bedrock:InvokeModel`) — **no key at all**.
- The live public deployment defaults to the free deterministic
  `FallbackProvider`, so the live link needs no provider secret and incurs no
  model cost.
- The repo is hardened for public hosting: a strict `.gitignore` excludes
  credentials, `.env`, private keys, and `cdk.out/`; `scripts/secret-scan.sh`
  scans for credential patterns and runs in CI and on commit.
- Intentionally public and safe to commit: the entire `.kiro/` tree
  (specs/steering/hooks/power/agents) and the CDK app under `infra/`.

## Public share report privacy boundary

- Share tokens are cryptographically random, opaque strings. They carry no
  launch id and are not guessable.
- `GET /api/r/{token}` is a no-auth public projection. It excludes
  `PrivateNotes` and any evidence marked private, and it reports only the
  stored readiness status.
- Regenerating or disabling a token immediately returns the unavailable state
  for the old token.
- The redaction and projection logic lives in `internal/domain/report.go` and
  `internal/security/redact.go`; keep them the sole gate for what leaves the
  trust boundary.

## Model trust boundary

- `security.RedactForModel` strips prohibited fields (keys/tokens/passwords,
  private repo content, unnecessary personal/customer/health/financial/identity
  data, `PrivateNotes`, private evidence, `Share`) from the snapshot before any
  `ModelProvider` call.
- External fetches are limited to allowlisted public URLs
  (`internal/security/allowlist.go`); external content is labeled
  `Untrusted_External_Content` and treated as data, never instructions.
