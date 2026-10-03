# UX guidance

The ShipCheck SPA (`web/`) is the only human surface. It must stay legible on a
phone and a laptop, and it must never leave the user staring at a blank panel.

## Four UI states on every data view

Every view that loads data from the API renders four explicit states as
first-class branches, not incidental fallbacks:

- **Loading** — a visible pending affordance while the request is in flight.
- **Empty** — the resource exists but has no data yet (e.g. a launch with no
  checklist items and no risks). Show a prompt that explains the next step.
- **No-results** — data exists but the active filter or search matched nothing
  (e.g. the "incomplete-only" or "high-priority" filter). This is distinct from
  empty and must say so.
- **Error** — the request failed. Show a human-readable error and a way to
  retry; never silently swallow the failure.

Treat these as a checklist for any new view. A view that only handles the
happy path is incomplete.

## Responsive layout

- The dashboard, checklist, risk register, brief, advisor, and public report
  all work on desktop and mobile widths.
- Prefer flow layouts that reflow over fixed multi-column grids that overflow.
- Readiness status and critical blockers must be visible without horizontal
  scrolling on a narrow viewport.

## Status presentation

- The readiness status comes from the deterministic engine. Render it exactly
  as computed — never recompute or soften it in the client.
- Advisor output is always labeled **AI-generated, requires human review**, and
  the public deployment disclaimer states ShipCheck is a non-commercial
  demonstration and that AI recommendations are not a release approval.
