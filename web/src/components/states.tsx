// The four required UI states (Req 14.2-14.5), as small reusable components.
// Keeping them in one place makes them easy to reuse across every data view
// and easy to test in isolation.

interface StateProps {
  title: string;
  detail?: string;
}

export function LoadingState({ title = 'Loading…', detail }: Partial<StateProps>) {
  return (
    <div role="status" aria-live="polite" className="py-12 text-center text-slate-500">
      <div className="mx-auto mb-3 h-6 w-6 animate-spin rounded-full border-2 border-slate-300 border-t-slate-600" />
      <p className="font-medium">{title}</p>
      {detail && <p className="mt-1 text-sm">{detail}</p>}
    </div>
  );
}

export function EmptyState({ title, detail }: StateProps) {
  return (
    <div className="rounded-lg border border-dashed border-slate-300 py-12 text-center text-slate-500">
      <p className="font-medium text-slate-700">{title}</p>
      {detail && <p className="mt-1 text-sm">{detail}</p>}
    </div>
  );
}

export function NoResultsState({ title = 'No matching items', detail }: Partial<StateProps>) {
  return (
    <div className="py-10 text-center text-slate-500">
      <p className="font-medium">{title}</p>
      {detail && <p className="mt-1 text-sm">{detail}</p>}
    </div>
  );
}

export function ErrorState({
  title = 'Something went wrong',
  detail,
  onRetry,
}: Partial<StateProps> & { onRetry?: () => void }) {
  return (
    <div role="alert" className="rounded-lg border border-red-200 bg-red-50 py-10 text-center text-red-700">
      <p className="font-medium">{title}</p>
      {detail && <p className="mt-1 text-sm">{detail}</p>}
      {onRetry && (
        <button
          onClick={onRetry}
          className="mt-4 rounded-md bg-red-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-red-700"
        >
          Try again
        </button>
      )}
    </div>
  );
}
