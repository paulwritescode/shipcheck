// Share controls: enable, regenerate, or disable the public read-only report
// link (Req 10.1, 10.5, 10.6). The link points at the SPA's #/r/<token> route.

import { useState } from 'react';
import type { Launch } from '../types';
import { api } from '../api';

function reportUrl(token: string): string {
  // The public report is an SPA hash route served from the same origin.
  return `${window.location.origin}/#/r/${token}`;
}

export function Share({ launch, onChange }: { launch: Launch; onChange: (l: Launch) => void }) {
  const [busy, setBusy] = useState(false);
  const [copied, setCopied] = useState(false);

  const enabled = !!launch.share?.enabled;
  const token = launch.share?.token ?? '';
  const url = enabled && token ? reportUrl(token) : '';

  async function act(action: 'enable' | 'regenerate' | 'disable') {
    setBusy(true);
    setCopied(false);
    try {
      onChange(await api.share(launch.id, action));
    } finally {
      setBusy(false);
    }
  }

  async function copy() {
    if (!url) return;
    try {
      await navigator.clipboard.writeText(url);
      setCopied(true);
    } catch {
      // Clipboard may be unavailable; the URL is shown for manual copy.
    }
  }

  return (
    <div className="space-y-5">
      <p className="text-sm text-slate-600">
        Share a public, read-only report of this launch. The report excludes private notes and
        evidence marked private.
      </p>

      {enabled && url ? (
        <div className="space-y-3">
          <div className="flex flex-col gap-2 sm:flex-row">
            <input
              readOnly
              value={url}
              className="min-w-0 flex-1 rounded-md border border-slate-300 px-3 py-1.5 text-sm text-slate-700"
              aria-label="Public report URL"
            />
            <button onClick={copy} className="rounded-md bg-slate-800 px-3 py-1.5 text-sm font-medium text-white">
              {copied ? 'Copied' : 'Copy link'}
            </button>
          </div>
          <div className="flex gap-2">
            <button
              onClick={() => act('regenerate')}
              disabled={busy}
              className="rounded-md border border-slate-300 px-3 py-1.5 text-sm disabled:opacity-50"
            >
              Regenerate link
            </button>
            <button
              onClick={() => act('disable')}
              disabled={busy}
              className="rounded-md border border-red-300 px-3 py-1.5 text-sm text-red-700 disabled:opacity-50"
            >
              Disable sharing
            </button>
          </div>
          <p className="text-xs text-slate-400">
            Regenerating issues a new link and stops the old one from resolving.
          </p>
        </div>
      ) : (
        <button
          onClick={() => act('enable')}
          disabled={busy}
          className="rounded-md bg-slate-900 px-4 py-2 text-sm font-semibold text-white disabled:opacity-50"
        >
          Enable public report
        </button>
      )}
    </div>
  );
}
