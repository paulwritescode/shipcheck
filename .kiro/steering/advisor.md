# Advisor guidance

ShipCheck splits authority cleanly: **the deterministic Readiness_Engine decides
the state; the ShipCheck_Advisor (AI) explains the state and recommends what to
do next.** The advisor is never the authority.

## Non-negotiable rules

- The advisor is **read-only** with respect to launch state. It never changes a
  readiness status, never approves or denies a release, never deletes or mutates
  launch data. (Verified by Properties 31 and 32.)
- The advisor operates on an **immutable `LaunchSnapshot` taken after** the
  engine computes readiness. The model therefore cannot alter the deterministic
  score — it only ever sees a result it cannot change.
- Every recommendation requires **explicit user action to apply**. "Accept"
  routes a suggestion through the normal checklist-create path; "dismiss" makes
  no change.
- Advisor output is **schema-validated structured JSON** (`LaunchAnalysis` /
  `Recommendation` / `SuggestedQuestions`). Any recommendation referencing a
  nonexistent launch entity is discarded and recorded as an error.
  (Verified by Properties 29–30.)

## Behavior

- Seven focused actions: Analyze Launch, Find Blockers, Suggest Checklist,
  Review Risks, Prepare Launch Review, Explain Readiness, Re-analyze After
  Changes.
- The provider lineup is behind the `ModelProvider` interface and selectable by
  config. The default is the free deterministic `FallbackProvider` (no external
  model, no key) that the live/public demo runs on. Opt-in providers
  (Hugging Face SLM, Bedrock Nova Micro, NVIDIA catalog, Local) read keys
  server-side only or use IAM-role auth.
- On provider failure the advisor returns an error state and leaves launch data
  unchanged; a loading state renders while running.
- All advisor output is labeled **AI-generated and requiring human review**.

Relevant code: `internal/advisor/` (advisor.go, provider.go, fallback.go,
validate.go, prompt.go, types.go).
