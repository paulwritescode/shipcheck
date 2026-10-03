# Custom agents — ShipCheck Advisor

`shipcheck-advisor.json` configures the **ShipCheck Advisor** as a Kiro custom
agent, demonstrating the custom-agent lesson.

## Strict read-only permissions

The agent is granted a **limited read-only toolset** (`read_file`,
`list_directory`, `file_search`, `grep_search`, `read_code`) and explicitly
denies every mutating tool (`fs_write`, `fs_append`, `str_replace`,
`delete_file`, `execute_bash`, `smart_relocate`, `semantic_rename`,
`createHook`).

This enforces the core ShipCheck invariant at the tooling layer: the advisor
**explains readiness and recommends actions but never mutates launch data,
never changes the status, and never approves or denies a release** (Req 9.8,
9.9). The deterministic Readiness_Engine remains the sole authority.

## Focused-action sub-agents

The advisor's focused actions map to the launch-review sub-agents described in
the design's Kiro Lessons Mapping:

- **Planning** (Suggest Checklist) — expands a launch brief into suggested
  checklist items across the seven categories.
- **Risk-review** (Review Risks) — inspects the risk register for unowned or
  under-mitigated high-severity risks.
- **QA-review** (Analyze Launch) — checks that completed critical items carry
  supporting evidence.

The same read-only, non-authoritative contract is enforced in the application
code at `internal/advisor/` and verified by Properties 31–32 (advisor never
mutates state; advisor output cannot change the score).
