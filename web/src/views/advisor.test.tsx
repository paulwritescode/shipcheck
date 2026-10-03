import { render, screen } from '@testing-library/react';
import type { Launch } from '../types';
import { Advisor } from './Advisor';

// The Advisor panel must always show the AI-generated / non-commercial-demo
// disclaimer (Req 21.1-21.2) and the focused action controls (Req 9.9).

function launch(): Launch {
  return {
    id: 'l1',
    name: 'Test',
    description: '',
    targetDate: '2026-10-16',
    owner: 'Alice',
    checklistItems: [],
    risks: [],
    assessment: { status: 'Needs Review', reasons: [], categories: [] },
  };
}

describe('Advisor panel', () => {
  it('shows the AI-generated / non-commercial demo disclaimer', () => {
    render(<Advisor launch={launch()} onChange={() => {}} />);
    expect(screen.getByText(/non-commercial demonstration/i)).toBeInTheDocument();
    expect(screen.getByText(/not a release approval/i)).toBeInTheDocument();
    expect(screen.getByText(/require human review/i)).toBeInTheDocument();
  });

  it('offers the seven focused actions', () => {
    render(<Advisor launch={launch()} onChange={() => {}} />);
    for (const label of [
      'Analyze launch',
      'Find blockers',
      'Suggest checklist',
      'Review risks',
      'Prepare review',
      'Explain readiness',
      'Re-analyze',
    ]) {
      expect(screen.getByRole('button', { name: label })).toBeInTheDocument();
    }
  });
});
