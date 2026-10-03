# Architecture guidance

ShipCheck is organized around one rule: the readiness logic is a **pure,
deterministic function of launch data**, isolated from storage, HTTP, the AWS
SDK, clocks, and randomness.

## Layers

- `internal/domain` — the pure core. The readiness engine, selectors,
  validators, and the seed builder live here. **No I/O, no clock, no
  randomness** in this package. This purity is what makes the 34 correctness
  properties verifiable with property-based tests.
- `internal/repo` — persistence behind the `LaunchRepository` interface
  (in-memory for tests/local, DynamoDB for deploy). The repository recomputes
  and stores the readiness assessment on every mutation.
- `internal/api` — HTTP handlers (same code for local `net/http` and Lambda).
  Validates input, calls the repository, returns JSON.
- `internal/advisor`, `internal/mcp`, `internal/security` — the AI advisor,
  read-only external context, and the trust-boundary guards.
- `web/` — the React SPA. `infra/` — the AWS CDK app (separate Go module).

## Rules the agent must follow

- Never introduce I/O, time, or randomness into `internal/domain`. If a value
  depends on the clock (e.g. "overdue"), compute it at the boundary and pass it
  in; the engine must stay referentially transparent.
- The deterministic engine is the **sole authority** for the readiness status.
  The AI advisor explains and recommends but never decides or changes status.
- Keep `LaunchRepository` the only persistence seam. The domain core and API
  depend on the interface, never on DynamoDB directly.
- Scope Go tooling to `./cmd/... ./internal/...`; `infra/` and `web/` are
  separate modules/toolchains.
