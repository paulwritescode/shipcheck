// Dashboard summarizes the launch: status, progress, per-category breakdown,
// critical blockers, high-severity-first risks, and recommended next actions
// (Req 8).

import type { Launch, Risk } from '../types';
import { CATEGORIES } from '../types';
import { EmptyState } from '../components/states';
import { SeverityBadge, StatusBadge, hasHighEmphasis } from '../components/badges';

function severityRank(s: Risk['severity']): number {
  return s === 'High' ? 0 : s === 'Medium' ? 1 : 2;
}

export function Dashboard({ launch }: { launch: Launch }) {
  const items = launch.checklistItems ?? [];
  const risks = launch.risks ?? [];

  if (items.length === 0 && risks.length === 0) {
    return (
      <EmptyState
        title="This launch has no data yet"
        detail="Add checklist items and risks to see a readiness assessment."
      />
    );
  }

  const completeCount = items.filter((i) => i.completionState === 'complete').length;
  const progress = items.length > 0 ? Math.round((completeCount / items.length) * 100) : 0;

  const blockers = [
    ...items
      .filter((i) => i.isCritical && i.completionState !== 'complete')
      .map((i) => ({ id: i.id, kind: 'Item', title: i.title })),
    ...risks
      .filter((r) => r.severity === 'High' && r.status !== 'Resolved' && r.status !== 'Accepted')
      .map((r) => ({ id: r.id, kind: 'Risk', title: r.title })),
  ];

  const orderedRisks = [...risks].sort((a, b) => severityRank(a.severity) - severityRank(b.severity));
  const catByName = new Map(launch.assessment.categories.map((c) => [c.category, c]));
  const actions = launch.assessment.reasons.filter((r) => r.recommendedAction);

  return (
    <div className="space-y-8">
      {/* Header: title, target date, status, progress */}
      <section className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 className="text-2xl font-semibold text-slate-900">{launch.name}</h2>
          <p className="text-sm text-slate-500">Target: {launch.targetDate}</p>
        </div>
        <div className="flex items-center gap-4">
          <StatusBadge status={launch.assessment.status} />
          <div className="text-right">
            <div className="text-sm font-medium text-slate-700">{progress}% complete</div>
            <div className="mt-1 h-2 w-32 overflow-hidden rounded bg-slate-200">
              <div className="h-full bg-slate-600" style={{ width: `${progress}%` }} />
            </div>
          </div>
        </div>
      </section>

      {/* Per-category breakdown */}
      <section>
        <h3 className="mb-3 text-sm font-semibold uppercase tracking-wide text-slate-500">Categories</h3>
        <div className="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3">
          {CATEGORIES.map((cat) => {
            const c = catByName.get(cat);
            const total = c?.totalCount ?? 0;
            const done = c?.completeCount ?? 0;
            return (
              <div key={cat} className="flex items-center justify-between rounded-md border border-slate-200 px-3 py-2">
                <span className="text-sm text-slate-700">{cat}</span>
                <span className="text-xs font-medium text-slate-500">
                  {total === 0 ? 'empty' : `${done}/${total}`}
                </span>
              </div>
            );
          })}
        </div>
      </section>

      {/* Critical blockers */}
      <section>
        <h3 className="mb-3 text-sm font-semibold uppercase tracking-wide text-slate-500">Critical blockers</h3>
        {blockers.length === 0 ? (
          <p className="text-sm text-slate-500">No critical blockers.</p>
        ) : (
          <ul className="space-y-2">
            {blockers.map((b) => (
              <li key={b.id} className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
                <span className="font-medium">{b.kind}:</span> {b.title}
              </li>
            ))}
          </ul>
        )}
      </section>

      {/* Recommended next actions */}
      {actions.length > 0 && (
        <section>
          <h3 className="mb-3 text-sm font-semibold uppercase tracking-wide text-slate-500">Recommended next actions</h3>
          <ul className="space-y-2">
            {actions.map((r, idx) => (
              <li key={`${r.ruleId}-${idx}`} className="rounded-md border border-slate-200 px-3 py-2 text-sm text-slate-700">
                {r.recommendedAction}
              </li>
            ))}
          </ul>
        </section>
      )}

      {/* Risks (high severity first) */}
      <section>
        <h3 className="mb-3 text-sm font-semibold uppercase tracking-wide text-slate-500">Risks</h3>
        {orderedRisks.length === 0 ? (
          <p className="text-sm text-slate-500">No risks recorded.</p>
        ) : (
          <ul className="space-y-2">
            {orderedRisks.map((r) => (
              <li
                key={r.id}
                className={`flex items-center justify-between rounded-md border px-3 py-2 text-sm ${
                  hasHighEmphasis(r) ? 'border-red-300 bg-red-50' : 'border-slate-200'
                }`}
              >
                <span className="text-slate-800">{r.title}</span>
                <span className="flex items-center gap-2">
                  <SeverityBadge severity={r.severity} />
                  <span className="text-xs text-slate-500">{r.status}</span>
                </span>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  );
}
