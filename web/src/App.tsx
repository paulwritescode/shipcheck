// App is the SPA root. It uses a tiny hash-based router so the whole SPA is a
// single static bundle (works on S3/CloudFront with no server routing):
//   #/                 -> landing
//   #/launch/:id       -> launch workspace
//   #/r/:token         -> public read-only report (completed in the share task)

import { useEffect, useState } from 'react';
import { Landing } from './views/Landing';
import { LaunchPage } from './views/LaunchPage';
import { PublicReport } from './views/PublicReport';

function useHashRoute(): string {
  const [hash, setHash] = useState(() => window.location.hash || '#/');
  useEffect(() => {
    const onChange = () => setHash(window.location.hash || '#/');
    window.addEventListener('hashchange', onChange);
    return () => window.removeEventListener('hashchange', onChange);
  }, []);
  return hash;
}

function navigate(to: string) {
  window.location.hash = to;
}

export default function App() {
  const hash = useHashRoute();
  const route = hash.replace(/^#/, '');

  let content: React.ReactNode;
  const launchMatch = route.match(/^\/launch\/([^/]+)$/);
  const reportMatch = route.match(/^\/r\/([^/]+)$/);

  if (launchMatch) {
    content = <LaunchPage launchId={decodeURIComponent(launchMatch[1])} />;
  } else if (reportMatch) {
    content = <PublicReport token={decodeURIComponent(reportMatch[1])} />;
  } else {
    content = <Landing onExplore={(id) => navigate(`/launch/${id}`)} />;
  }

  return (
    <div className="min-h-screen bg-slate-50 text-slate-900">
      <header className="border-b border-slate-200 bg-white">
        <div className="mx-auto flex max-w-5xl items-center justify-between px-4 py-3">
          <button onClick={() => navigate('/')} className="text-lg font-semibold text-slate-900">
            ShipCheck
          </button>
          <span className="text-xs text-slate-400">launch-readiness workspace</span>
        </div>
      </header>
      <main className="mx-auto max-w-5xl px-4 py-8">{content}</main>
    </div>
  );
}
