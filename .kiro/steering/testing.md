# Testing guidance

Property-based testing is the primary and highest-value test strategy for
ShipCheck. The readiness engine is a pure function over a large structured
input space with strong universal invariants, which is exactly where PBT finds
the edge cases hand-written examples miss.

## Mandatory property-based testing

- Use **`pgregory.net/rapid`** (integrated with `go test`, automatic shrinking).
  Do not hand-roll a PBT framework.
- The 34 design properties live in `*_property_test.go` next to the pure code
  they verify:
  - `internal/domain/engine_property_test.go` — Properties 1–17
  - `internal/domain/selectors_property_test.go` — Properties 18–26
  - `internal/domain/report_property_test.go` — Property 27
  - `internal/repo/repo_property_test.go` — Property 28
  - `internal/advisor/validate_property_test.go` — Properties 29–30
  - `internal/advisor/advisor_property_test.go` — Properties 31–32
  - `internal/security/security_property_test.go` — Properties 33–34
- Every property test carries the comment tag
  **`Feature: shipcheck, Property {n}: {text}`**, maps to exactly one design
  property, and references the requirement clause(s) it validates.
- Each property runs a **minimum of 100 checks**. CI and release verification
  run them harder (`-rapid.checks=1000`).
- Test the pure core **without mocks**. Custom generators produce well-formed
  `LaunchData`, including tier-targeting generators (empty, incomplete-critical,
  blocking-risk, complete-without-evidence, all-complete) for precedence.

## What is tested by example instead of PBT

CRUD handlers, SPA UI states and responsive rendering, persistence wiring, the
DynamoDB repository (DynamoDB Local / real table), provider selection, and
prompt-injection *resistance* (which depends on external model behavior) are
covered by example, integration, smoke, or component tests — never by the
property suite. SPA tests use Vitest + React Testing Library for UI behavior
only.

## Scope and running

- Scope Go tooling to `./cmd/... ./internal/...` (`infra/` and `web/` are
  separate toolchains).
- `make ci` = gofmt check + vet + build + test + secret-scan.
- Property focus: `go test ./internal/domain/ -run Property -rapid.checks=1000`.
- SPA: `cd web && npm test && npm run typecheck && npm run lint && npm run build`.

Add property tests when adding new pure logic to the domain core; add example
tests when fixing a CRUD or UI bug.
