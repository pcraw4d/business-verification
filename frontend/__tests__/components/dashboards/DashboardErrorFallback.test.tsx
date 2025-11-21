import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { DashboardErrorFallback } from '@/components/dashboards/DashboardErrorFallback';

describe('DashboardErrorFallback', () => {
  it('should render error message', () => {
    const resetError = vi.fn();
    render(<DashboardErrorFallback error={null} resetError={resetError} />);

    expect(screen.getByText(/Failed to Load Dashboard/i)).toBeInTheDocument();
    expect(screen.getByText(/We couldn't load the dashboard data/i)).toBeInTheDocument();
  });

  it('should render retry button', () => {
    const resetError = vi.fn();
    render(<DashboardErrorFallback error={null} resetError={resetError} />);

    const retryButton = screen.getByRole('button', { name: /Retry/i });
    expect(retryButton).toBeInTheDocument();
  });

  it('should call resetError when retry is clicked', async () => {
    const user = userEvent.setup();
    const resetError = vi.fn();

    render(<DashboardErrorFallback error={null} resetError={resetError} />);

    const retryButton = screen.getByRole('button', { name: /Retry/i });
    await user.click(retryButton);

    expect(resetError).toHaveBeenCalled();
  });

  it('should show error details in development mode', () => {
    const originalEnv = process.env.NODE_ENV;
    process.env.NODE_ENV = 'development';
    const resetError = vi.fn();
    const testError = new Error('Test error');

    render(<DashboardErrorFallback error={testError} resetError={resetError} />);

    expect(screen.getByText(/Error details \(development only\)/i)).toBeInTheDocument();

    process.env.NODE_ENV = originalEnv;
  });

  it('should not show error details in production mode', () => {
    const originalEnv = process.env.NODE_ENV;
    process.env.NODE_ENV = 'production';
    const resetError = vi.fn();
    const testError = new Error('Test error');

    render(<DashboardErrorFallback error={testError} resetError={resetError} />);

    expect(screen.queryByText(/Error details \(development only\)/i)).not.toBeInTheDocument();

    process.env.NODE_ENV = originalEnv;
  });
});

