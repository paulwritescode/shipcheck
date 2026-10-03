// Landing page: describes the product and provides a one-click entry into the
// seeded demo launch, with no account required (Req 11).

import { SEED_LAUNCH_ID } from '../types';

export function Landing({ onExplore }: { onExplore: (launchId: string) => void }) {
  return (
    <div className="mx-auto max-w-2xl px-4 py-16 text-center">
      <h1 className="text-4xl font-bold tracking-tight text-slate-900">ShipCheck</h1>
      <p className="mt-4 text-lg text-slate-600">
        An evidence-based launch-readiness workspace. Find release blockers, assign the remaining
        work, and share an explainable go/no-go report.
      </p>
      <div className="mt-8 flex flex-col items-center gap-3 sm:flex-row sm:justify-center">
        <button
          onClick={() => onExplore(SEED_LAUNCH_ID)}
          className="rounded-md bg-slate-900 px-5 py-2.5 text-sm font-semibold text-white hover:bg-slate-800"
        >
          Explore the demo launch
        </button>
      </div>
      <p className="mt-4 text-sm text-slate-500">No account needed — explore the seeded “Team Inbox 2.0” launch.</p>

      <div className="mt-12 grid grid-cols-1 gap-4 text-left sm:grid-cols-3">
        {[
          ['Evidence first', 'Back completion claims with links and notes.'],
          ['Explainable status', 'Every readiness decision shows why.'],
          ['Shareable report', 'Give stakeholders a clean read-only view.'],
        ].map(([title, body]) => (
          <div key={title} className="rounded-lg border border-slate-200 p-4">
            <h3 className="text-sm font-semibold text-slate-900">{title}</h3>
            <p className="mt-1 text-sm text-slate-600">{body}</p>
          </div>
        ))}
      </div>
    </div>
  );
}
