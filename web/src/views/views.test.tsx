import { render, screen } from '@testing-library/react';
import type { Launch } from '../types';
import { Dashboard } from './Dashboard';
import { Risks } from './Risks';
import { hasHighEmphasis } from '../components/badges';

// View-level UI tests: dashboard empty state (Req 8.7) and risk high-emphasis
// treatment (Req 4.6). These are component/UI-state tests only — the universal
// correctness guarantees live in the Go property suite.

function baseLaunch(overrides: Partial<Launch> = {}): Launch {
  return {
    id: 'l1',
    name: 'Test Launch',
    description: '',
    targetDate: '2026-10-16',
    owner: 'Alice',
    checklistItems: [],
    risks: [],
    assessment: { status: 'Needs Review', reasons: [], categories: [] },
    ...overrides,
  };
}

describe('Dashboard', () => {
  it('shows an empty state when there are no items and no risks', () => {
    render(<Dashboard launch={baseLaunch()} />);
    expect(screen.getByText(/no data yet/i)).toBeInTheDocument();
  });

  it('shows status and progress when data exists', () => {
    const launch = baseLaunch({
      checklistItems: [
        { id: 'i1', title: 'done', category: 'Engineering', priority: 'Low', completionState: 'complete', isCritical: false },
        { id: 'i2', title: 'todo', category: 'Engineering', priority: 'Low', completionState: 'incomplete', isCritical: false },
      ],
      assessment: {
        status: 'Conditionally Ready',
        reasons: [],
        categories: [{ category: 'Engineering', completeCount: 1, totalCount: 2, state: 'partial' }],
      },
    });
    render(<Dashboard launch={launch} />);
    expect(screen.getByTestId('status-badge')).toHaveTextContent('Conditionally Ready');
    expect(screen.getByText(/50% complete/i)).toBeInTheDocument();
  });
});

describe('Risks high-emphasis', () => {
  it('applies high-emphasis to High-severity or High-likelihood risks', () => {
    const launch = baseLaunch({
      risks: [
        { id: 'r1', title: 'critical', severity: 'High', likelihood: 'Low', status: 'Open' },
        { id: 'r2', title: 'minor', severity: 'Low', likelihood: 'Low', status: 'Open' },
        { id: 'r3', title: 'likely', severity: 'Low', likelihood: 'High', status: 'Open' },
      ],
    });
    render(<Risks launch={launch} readOnly onChange={() => {}} />);
    const rows = screen.getAllByTestId('risk-row');
    const byEmphasis = rows.map((el) => el.getAttribute('data-high-emphasis'));
    // Two of the three risks (r1 High severity, r3 High likelihood) are emphasized.
    expect(byEmphasis.filter((v) => v === 'true')).toHaveLength(2);
  });

  it('hasHighEmphasis matches the backend predicate', () => {
    expect(hasHighEmphasis({ id: 'x', title: '', severity: 'High', likelihood: 'Low', status: 'Open' })).toBe(true);
    expect(hasHighEmphasis({ id: 'x', title: '', severity: 'Low', likelihood: 'High', status: 'Open' })).toBe(true);
    expect(hasHighEmphasis({ id: 'x', title: '', severity: 'Low', likelihood: 'Low', status: 'Open' })).toBe(false);
  });
});
