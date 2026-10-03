// Small presentational badges for readiness status, priority, and risk
// severity/likelihood, with the color treatment the dashboard and registers use.

import type {
  CompletionState,
  Priority,
  ReadinessStatus,
  Risk,
  Severity,
} from '../types';

const statusClasses: Record<ReadinessStatus, string> = {
  Ready: 'bg-green-100 text-green-800 border-green-300',
  'Conditionally Ready': 'bg-amber-100 text-amber-800 border-amber-300',
  'Not Ready': 'bg-red-100 text-red-800 border-red-300',
  'Needs Review': 'bg-slate-100 text-slate-700 border-slate-300',
};

export function StatusBadge({ status }: { status: ReadinessStatus }) {
  return (
    <span
      data-testid="status-badge"
      className={`inline-flex items-center rounded-full border px-3 py-1 text-sm font-semibold ${statusClasses[status]}`}
    >
      {status}
    </span>
  );
}

const priorityClasses: Record<Priority, string> = {
  High: 'bg-red-50 text-red-700 border-red-200',
  Medium: 'bg-amber-50 text-amber-700 border-amber-200',
  Low: 'bg-slate-50 text-slate-600 border-slate-200',
};

export function PriorityBadge({ priority }: { priority: Priority }) {
  return (
    <span className={`inline-flex rounded border px-2 py-0.5 text-xs font-medium ${priorityClasses[priority]}`}>
      {priority}
    </span>
  );
}

const sevClasses: Record<Severity, string> = {
  High: 'bg-red-100 text-red-800',
  Medium: 'bg-amber-100 text-amber-800',
  Low: 'bg-slate-100 text-slate-700',
};

export function SeverityBadge({ severity }: { severity: Severity }) {
  return <span className={`inline-flex rounded px-2 py-0.5 text-xs font-medium ${sevClasses[severity]}`}>{severity}</span>;
}

const completionLabels: Record<CompletionState, string> = {
  complete: 'Complete',
  'in progress': 'In progress',
  blocked: 'Blocked',
  incomplete: 'Incomplete',
};

const completionClasses: Record<CompletionState, string> = {
  complete: 'bg-green-100 text-green-800',
  'in progress': 'bg-blue-100 text-blue-800',
  blocked: 'bg-red-100 text-red-800',
  incomplete: 'bg-slate-100 text-slate-700',
};

export function CompletionBadge({ state }: { state: CompletionState }) {
  return (
    <span className={`inline-flex rounded px-2 py-0.5 text-xs font-medium ${completionClasses[state]}`}>
      {completionLabels[state]}
    </span>
  );
}

/** hasHighEmphasis mirrors the backend predicate (Req 4.6). */
export function hasHighEmphasis(r: Risk): boolean {
  return r.severity === 'High' || r.likelihood === 'High';
}
