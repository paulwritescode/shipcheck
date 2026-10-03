# Launch-review workflow guide (ShipCheck power)

A repeatable procedure for running a launch review. Follow it top to bottom to
produce an evidence-backed go/no-go for any release.

## When to use

Run this before shipping a product, a significant feature, or a major update —
and again after any material change to the launch's checklist or risks.

## Procedure

### 1. Establish the launch brief
Capture what is shipping, the target date, and the owner. If there is no owner,
the launch is Needs Review until one is assigned.

### 2. Build the checklist across the seven categories
Product & Scope, Engineering & Quality, Launch Operations, Security & Privacy,
Legal & Compliance, Communications & Enablement, Post-Launch & Monitoring. For
each item set: owner, priority, critical flag, and (when complete) evidence.

### 3. Build the risk register
For each risk record severity, likelihood, owner, mitigation, status, and due
date. A High-severity risk that is neither Resolved nor Accepted is a
**blocking risk**.

### 4. Compute readiness (deterministic)
Apply the tier precedence — Needs Review > Not Ready > Conditionally Ready >
Ready — and take the first match. Produce one reason per cause: per incomplete
critical item, per blocking risk, per unevidenced completion claim.

### 5. Advise, don't decide (optional AI step)
Run the advisor's focused actions against a snapshot taken **after** the status
is computed:
- **Find Blockers** — list what stands between here and Ready.
- **Review Risks** — surface unowned or under-mitigated high-severity risks.
- **Suggest Checklist** — expand the brief into missing items.
- **Prepare Launch Review** — assemble a review summary.
- **Explain Readiness** — justify the current status in plain language.
Every suggestion is advisory and requires explicit human action to apply.

### 6. Close the gaps and re-review
Attach evidence, complete or re-scope critical items, mitigate or accept risks,
then recompute. Repeat until the status is Ready or the team accepts a
Conditionally Ready launch with eyes open.

### 7. Share the result
Produce a read-only report (status + reasons) for stakeholders. Exclude private
notes and private evidence from anything shared outside the team.

## Definition of done

The review is complete when the computed status is Ready, or when the team has
explicitly accepted every remaining non-critical gap and risk and recorded that
acceptance as evidence.
