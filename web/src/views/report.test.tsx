import { render, screen, waitFor } from '@testing-library/react';
import { PublicReport } from './PublicReport';
import { ApiError, api } from '../api';
import type { PublicReport as Report } from '../types';

// The public report must render a report for a valid token and an "unavailable"
// message when the token is disabled/regenerated/unknown (Req 10.7).

describe('PublicReport', () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('renders the report for a valid token', async () => {
    const report: Report = {
      name: 'Team Inbox 2.0',
      targetDate: '2026-10-16',
      status: 'Not Ready',
      categories: [],
      completedWork: [{ title: 'done thing', category: 'Engineering', completionState: 'complete', critical: false }],
      incompleteWork: [],
      highPriorityRisks: [],
      openQuestions: [],
      lastUpdated: '2026-10-16T00:00:00Z',
    };
    vi.spyOn(api, 'getPublicReport').mockResolvedValue(report);

    render(<PublicReport token="valid" />);
    await waitFor(() => expect(screen.getByText('Team Inbox 2.0')).toBeInTheDocument());
    expect(screen.getByTestId('status-badge')).toHaveTextContent('Not Ready');
  });

  it('shows an unavailable message on a 404', async () => {
    vi.spyOn(api, 'getPublicReport').mockRejectedValue(new ApiError(404, 'unavailable'));

    render(<PublicReport token="gone" />);
    await waitFor(() => expect(screen.getByText(/report is unavailable/i)).toBeInTheDocument());
  });
});
