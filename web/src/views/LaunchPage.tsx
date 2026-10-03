// LaunchPage loads a launch by id and hosts the tabbed views. It owns the
// loading/error states for the initial fetch (Req 14).

import { useCallback, useEffect, useState } from 'react';
import type { Launch } from '../types';
import { ApiError, api } from '../api';
import { ErrorState, LoadingState } from '../components/states';
import { Dashboard } from './Dashboard';
import { Checklist } from './Checklist';
import { Risks } from './Risks';
import { Brief } from './Brief';
import { Advisor } from './Advisor';
import { Share } from './Share';

type Tab = 'dashboard' | 'checklist' | 'risks' | 'brief' | 'advisor' | 'share';

const TABS: { key: Tab; label: string }[] = [
  { key: 'dashboard', label: 'Dashboard' },
  { key: 'checklist', label: 'Checklist' },
  { key: 'risks', label: 'Risks' },
  { key: 'brief', label: 'Brief' },
  { key: 'advisor', label: 'Advisor' },
  { key: 'share', label: 'Share' },
];

export function LaunchPage({ launchId }: { launchId: string }) {
  const [launch, setLaunch] = useState<Launch | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [tab, setTab] = useState<Tab>('dashboard');

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      setLaunch(await api.getLaunch(launchId));
    } catch (e) {
      const msg = e instanceof ApiError && e.status === 404 ? 'Launch not found.' : 'Could not load the launch.';
      setError(msg);
    } finally {
      setLoading(false);
    }
  }, [launchId]);

  useEffect(() => {
    void load();
  }, [load]);

  if (loading) return <LoadingState title="Loading launch…" />;
  if (error || !launch) return <ErrorState detail={error ?? 'Unknown error'} onRetry={load} />;

  return (
    <div className="space-y-6">
      <nav className="flex gap-1 border-b border-slate-200">
        {TABS.map((t) => (
          <button
            key={t.key}
            onClick={() => setTab(t.key)}
            className={`-mb-px border-b-2 px-4 py-2 text-sm font-medium ${
              tab === t.key
                ? 'border-slate-800 text-slate-900'
                : 'border-transparent text-slate-500 hover:text-slate-800'
            }`}
          >
            {t.label}
          </button>
        ))}
      </nav>

      {tab === 'dashboard' && <Dashboard launch={launch} />}
      {tab === 'checklist' && <Checklist launch={launch} onChange={setLaunch} />}
      {tab === 'risks' && <Risks launch={launch} onChange={setLaunch} />}
      {tab === 'brief' && <Brief launch={launch} onChange={setLaunch} />}
      {tab === 'advisor' && <Advisor launch={launch} onChange={setLaunch} />}
      {tab === 'share' && <Share launch={launch} onChange={setLaunch} />}
    </div>
  );
}
