// Risk register: severity/likelihood/owner/mitigation/status/due date with
// high-emphasis treatment for High severity or likelihood (Req 4).

import { useState } from 'react';
import type { Launch, Risk, RiskStatus } from '../types';
import { api } from '../api';
import { SeverityBadge, hasHighEmphasis } from '../components/badges';
import { EmptyState } from '../components/states';

const STATUSES: RiskStatus[] = ['Open', 'Mitigating', 'Resolved', 'Accepted'];

export function Risks({
  launch,
  readOnly,
  onChange,
}: {
  launch: Launch;
  readOnly?: boolean;
  onChange: (l: Launch) => void;
}) {
  const [newTitle, setNewTitle] = useState('');
  const [busy, setBusy] = useState(false);
  const risks = launch.risks ?? [];

  async function mutate(fn: () => Promise<Launch>) {
    setBusy(true);
    try {
      onChange(await fn());
    } finally {
      setBusy(false);
    }
  }

  async function addRisk() {
    if (!newTitle.trim()) return;
    await mutate(() => api.addRisk(launch.id, { title: newTitle, severity: 'Medium', likelihood: 'Medium', status: 'Open' }));
    setNewTitle('');
  }

  async function setStatus(risk: Risk, status: RiskStatus) {
    await mutate(() => api.updateRisk(launch.id, { ...risk, status }));
  }

  async function remove(risk: Risk) {
    await mutate(() => api.deleteRisk(launch.id, risk.id));
  }

  return (
    <div className="space-y-6">
      {!readOnly && (
        <div className="flex flex-wrap gap-2">
          <input
            value={newTitle}
            onChange={(e) => setNewTitle((e.target as HTMLInputElement).value)}
            placeholder="New risk…"
            className="min-w-0 flex-1 rounded-md border border-slate-300 px-3 py-1.5 text-sm"
            aria-label="New risk title"
          />
          <button
            onClick={addRisk}
            disabled={busy || !newTitle.trim()}
            className="rounded-md bg-slate-800 px-3 py-1.5 text-sm font-medium text-white disabled:opacity-50"
          >
            Add risk
          </button>
        </div>
      )}

      {risks.length === 0 ? (
        <EmptyState title="No risks recorded" detail="Add a risk to track dangers to the release." />
      ) : (
        <ul className="space-y-2">
          {risks.map((r) => (
            <li
              key={r.id}
              data-testid="risk-row"
              data-high-emphasis={hasHighEmphasis(r)}
              className={`rounded-md border px-3 py-2 ${
                hasHighEmphasis(r) ? 'border-red-300 bg-red-50' : 'border-slate-200'
              }`}
            >
              <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                <div>
                  <div className="text-sm font-medium text-slate-800">{r.title}</div>
                  <div className="mt-1 flex items-center gap-2 text-xs text-slate-500">
                    <SeverityBadge severity={r.severity} />
                    <span>Likelihood: {r.likelihood}</span>
                    {r.owner ? <span>· {r.owner}</span> : <span className="text-amber-600">· no owner</span>}
                    {r.dueDate && <span>· due {r.dueDate}</span>}
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  {readOnly ? (
                    <span className="text-xs text-slate-500">{r.status}</span>
                  ) : (
                    <select
                      value={r.status}
                      onChange={(e) => setStatus(r, (e.target as HTMLSelectElement).value as RiskStatus)}
                      disabled={busy}
                      className="rounded-md border border-slate-300 px-2 py-1 text-xs"
                      aria-label={`Status for ${r.title}`}
                    >
                      {STATUSES.map((s) => (
                        <option key={s} value={s}>
                          {s}
                        </option>
                      ))}
                    </select>
                  )}
                  {!readOnly && (
                    <button onClick={() => remove(r)} disabled={busy} className="text-xs text-red-500 hover:text-red-700">
                      delete
                    </button>
                  )}
                </div>
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
