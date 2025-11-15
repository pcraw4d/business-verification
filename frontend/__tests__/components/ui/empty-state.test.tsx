import { describe, it, expect } from '@jest/globals';
import { render, screen } from '@testing-library/react';
import { EmptyState } from '@/components/ui/empty-state';

describe('EmptyState', () => {
  it('should render noData state by default', () => {
    render(<EmptyState />);
    expect(screen.getByText(/No Data Available/i)).toBeInTheDocument();
  });

  it('should render error state', () => {
    render(<EmptyState type="error" title="Error Title" message="Error message" />);
    expect(screen.getByText('Error Title')).toBeInTheDocument();
    expect(screen.getByText('Error message')).toBeInTheDocument();
  });

  it('should render noResults state', () => {
    render(<EmptyState type="noResults" />);
    expect(screen.getByText(/No Results Found/i)).toBeInTheDocument();
  });

  it('should render action button when provided', () => {
    const onAction = jest.fn();
    render(
      <EmptyState
        type="error"
        actionLabel="Retry"
        onAction={onAction}
      />
    );

    const button = screen.getByText('Retry');
    expect(button).toBeInTheDocument();
    
    button.click();
    expect(onAction).toHaveBeenCalledTimes(1);
  });

  it('should not render action button when onAction not provided', () => {
    render(<EmptyState actionLabel="Action" />);
    expect(screen.queryByText('Action')).not.toBeInTheDocument();
  });
});

