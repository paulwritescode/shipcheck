// Launch brief editor with an empty-state prompt when no brief content exists
// yet (Req 2).

import { useState } from 'react';
import type { Launch, LaunchBrief } from '../types';
import { api } from '../api';
import { EmptyState } from '../components/states';

const FIELDS: { key: keyof LaunchBrief; label: string }[] = [
  { key: 'whatIsReleasing', label: 'What is being released' },
  { key: 'audience', label: 'Audience' },
  { key: 'successDefinition', label: 'Definition of success' },
  { key: 'requirements', label: 'Requirements' },
  { key: 'constraints', label: 'Known constraints' },
  { key: 'dependencies', label: 'Dependencies' },
];

export function Brief({
  launch,
  readOnly,
  onChange,
}: {
  launch: Launch;
  readOnly?: boolean;
  onChange: (l: Launch) => void;
}) {
  const [draft, setDraft] = useState<LaunchBrief>(launch.brief ?? {});
  const [busy, setBusy] = useState(false);
  const [editing, setEditing] = useState(false);

  const hasContent = Object.values(launch.brief ?? {}).some((v) => (v ?? '').trim() !== '');

  async function save() {
    setBusy(true);
    try {
      onChange(await api.updateLaunch(launch.id, { brief: draft }));
      setEditing(false);
    } finally {
      setBusy(false);
    }
  }

  if (!hasContent && !editing) {
    return readOnly ? (
      <EmptyState title="No launch brief" detail="The brief has not been filled in." />
    ) : (
      <div>
        <EmptyState title="No launch brief yet" detail="Describe what you are releasing and what success means." />
        <div className="mt-4 text-center">
          <button onClick={() => setEditing(true)} className="rounded-md bg-slate-800 px-3 py-1.5 text-sm font-medium text-white">
            Add brief
          </button>
        </div>
      </div>
    );
  }

  if (readOnly || !editing) {
    return (
      <div className="space-y-4">
        {FIELDS.map(({ key, label }) => {
          const value = (launch.brief ?? {})[key];
          if (!value) return null;
          return (
            <div key={key}>
              <h4 className="text-xs font-semibold uppercase tracking-wide text-slate-500">{label}</h4>
              <p className="mt-1 whitespace-pre-wrap text-sm text-slate-700">{value}</p>
            </div>
          );
        })}
        {!readOnly && (
          <button onClick={() => setEditing(true)} className="rounded-md border border-slate-300 px-3 py-1.5 text-sm">
            Edit brief
          </button>
        )}
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {FIELDS.map(({ key, label }) => (
        <div key={key}>
          <label className="text-xs font-semibold uppercase tracking-wide text-slate-500">{label}</label>
          <textarea
            value={draft[key] ?? ''}
            onChange={(e) => setDraft({ ...draft, [key]: (e.target as HTMLTextAreaElement).value })}
            rows={2}
            className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm"
          />
        </div>
      ))}
      <div className="flex gap-2">
        <button onClick={save} disabled={busy} className="rounded-md bg-slate-800 px-3 py-1.5 text-sm font-medium text-white disabled:opacity-50">
          Save
        </button>
        <button onClick={() => setEditing(false)} className="rounded-md border border-slate-300 px-3 py-1.5 text-sm">
          Cancel
        </button>
      </div>
    </div>
  );
}
