import { render, screen } from '@testing-library/react';
import App from './App';

describe('App', () => {
  it('renders the landing page by default', () => {
    window.location.hash = '#/';
    render(<App />);
    // The product name appears in the header and the landing hero.
    expect(screen.getAllByText(/shipcheck/i).length).toBeGreaterThan(0);
    expect(
      screen.getByRole('button', { name: /explore the demo launch/i }),
    ).toBeInTheDocument();
  });
});
