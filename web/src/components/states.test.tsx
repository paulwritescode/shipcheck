import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { EmptyState, ErrorState, LoadingState, NoResultsState } from './states';

// Component tests for the four required UI states (Req 14.2-14.5).

describe('UI states', () => {
  it('renders a loading state with a status role', () => {
    render(<LoadingState title="Loading launch…" />);
    expect(screen.getByRole('status')).toHaveTextContent(/loading launch/i);
  });

  it('renders an empty state with title and detail', () => {
    render(<EmptyState title="No launch brief yet" detail="Describe what you are releasing." />);
    expect(screen.getByText(/no launch brief yet/i)).toBeInTheDocument();
    expect(screen.getByText(/describe what you are releasing/i)).toBeInTheDocument();
  });

  it('renders a no-results state', () => {
    render(<NoResultsState detail="No items match this filter." />);
    expect(screen.getByText(/no matching items/i)).toBeInTheDocument();
    expect(screen.getByText(/no items match this filter/i)).toBeInTheDocument();
  });

  it('renders an error state with an alert role and retry', async () => {
    const onRetry = vi.fn();
    render(<ErrorState detail="Could not load the launch." onRetry={onRetry} />);
    expect(screen.getByRole('alert')).toHaveTextContent(/could not load/i);
    await userEvent.click(screen.getByRole('button', { name: /try again/i }));
    expect(onRetry).toHaveBeenCalledOnce();
  });
});
