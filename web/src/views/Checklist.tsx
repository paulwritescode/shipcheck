// Checklist view: items grouped by category, filters, and create/edit/
// complete/flag/delete with evidence (Req 3, 5). Mutations call the API and
// hand the returned launch back to the parent via onChange.

import { useState } from 'react';
import type { Category, ChecklistItem, CompletionState, Launch } from '../types';
import { CATEGORIES } from '../types';
import { api } from '../api';
import { CompletionBadge, PriorityBadge } from '../components/badges';
import { NoResultsState } from '../components/states';

type Filter = 'all' | 'incomplete' | 'highPriority';

export function Checklist({
  launch,
  readOnly,
  onChange,
}: {
  launch: Launch;
  readOnly?: boolean;
  onChange: (l: Launch) => void;
}) {
  const [filter, setFilter] = useState<Filter>('all');
  const [newTitle, setNewTitle] = useState('');
  const [newCategory, setNewCategory] = useState<Category>('Engineering');
  const [busy, setBusy] = useState(false);

  const items = launch.checklistItems ?? [];
  const filtered = items.filter((i) => {
    if (filter === 'incomplete') return i.completionState === 'incomplete';
    if (filter === 'highPriority') return i.priority === 'High';
    return true;
  });

  async function mutate(fn: () => Promise<Launch>) {
    setBusy(true);
    try {
      onChange(await fn());
    } finally {
      setBusy(false);
    }
  }

  async function addItem() {
    if (!newTitle.trim()) return;
    await mutate(() => api.addItem(launch.id, { title: newTitle, category: newCategory }));
    setNewTitle('');
  }

  async function cycleCompletion(item: ChecklistItem) {
    const order: CompletionState[] = ['incomplete', 'in progress', 'blocked', 'complete'];
    const next = order[(order.indexOf(item.completionState) + 1) % order.length];
    await mutate(() => api.updateItem(launch.id, { ...item, completionState: next }));
  }

  async function toggleCritical(item: ChecklistItem) {
    await mutate(() => api.updateItem(launch.id, { ...item, isCritical: !item.isCritical }));
  }

  async function remove(item: ChecklistItem) {
    await mutate(() => api.deleteItem(launch.id, item.id));
  }

  const grouped = CATEGORIES.map((cat) => ({
    cat,
    items: filtered.filter((i) => i.category === cat),
  })).filter((g) => g.items.length > 0);

  return (
    <div className="space-y-6">
      {/* Filters */}
      <div className="flex flex-wrap items-center gap-2">
        {(['all', 'incomplete', 'highPriority'] as Filter[]).map((f) => (
          <button
            key={f}
            onClick={() => setFilter(f)}
            className={`rounded-full px-3 py-1 text-sm ${
              filter === f ? 'bg-slate-800 text-white' : 'bg-slate-100 text-slate-700 hover:bg-slate-200'
            }`}
          >
            {f === 'all' ? 'All' : f === 'incomplete' ? 'Incomplete' : 'High priority'}
          </button>
        ))}
      </div>

      {/* Add item */}
      {!readOnly && (
        <div className="flex flex-wrap gap-2">
          <input
            value={newTitle}
            onChange={(e) => setNewTitle((e.target as HTMLInputElement).value)}
            placeholder="New checklist item…"
            className="min-w-0 flex-1 rounded-md border border-slate-300 px-3 py-1.5 text-sm"
            aria-label="New checklist item title"
          />
          <select
            value={newCategory}
            onChange={(e) => setNewCategory((e.target as HTMLSelectElement).value as Category)}
            className="rounded-md border border-slate-300 px-2 py-1.5 text-sm"
            aria-label="New item category"
          >
            {CATEGORIES.map((c) => (
              <option key={c} value={c}>
                {c}
              </option>
            ))}
          </select>
          <button
            onClick={addItem}
            disabled={busy || !newTitle.trim()}
            className="rounded-md bg-slate-800 px-3 py-1.5 text-sm font-medium text-white disabled:opacity-50"
          >
            Add
          </button>
        </div>
      )}

      {/* Grouped items */}
      {filtered.length === 0 ? (
        <NoResultsState detail="No items match this filter." />
      ) : (
        grouped.map((g) => (
          <section key={g.cat}>
            <h3 className="mb-2 text-sm font-semibold uppercase tracking-wide text-slate-500">{g.cat}</h3>
            <ul className="space-y-2">
              {g.items.map((item) => (
                <li
                  key={item.id}
                  className="flex flex-col gap-2 rounded-md border border-slate-200 px-3 py-2 sm:flex-row sm:items-center sm:justify-between"
                >
                  <div className="flex items-center gap-2">
                    <span className="text-sm text-slate-800">{item.title}</span>
                    {item.isCritical && (
                      <span className="rounded bg-purple-100 px-1.5 py-0.5 text-xs font-medium text-purple-700">
                        Critical
                      </span>
                    )}
                    {item.owner ? (
                      <span className="text-xs text-slate-500">· {item.owner}</span>
                    ) : (
                      <span className="text-xs text-amber-600">· no owner</span>
                    )}
                  </div>
                  <div className="flex items-center gap-2">
                    <PriorityBadge priority={item.priority} />
                    <button onClick={() => !readOnly && cycleCompletion(item)} disabled={readOnly || busy} title="Cycle status">
                      <CompletionBadge state={item.completionState} />
                    </button>
                    {!readOnly && (
                      <>
                        <button
                          onClick={() => toggleCritical(item)}
                          disabled={busy}
                          className="text-xs text-slate-500 hover:text-slate-800"
                        >
                          flag
                        </button>
                        <button
                          onClick={() => remove(item)}
                          disabled={busy}
                          className="text-xs text-red-500 hover:text-red-700"
                        >
                          delete
                        </button>
                      </>
                    )}
                  </div>
                </li>
              ))}
            </ul>
          </section>
        ))
      )}
    </div>
  );
}
