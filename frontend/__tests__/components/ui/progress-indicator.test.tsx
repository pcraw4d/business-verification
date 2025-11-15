import { describe, it, expect } from '@jest/globals';
import { render, screen } from '@testing-library/react';
import { ProgressIndicator } from '@/components/ui/progress-indicator';

describe('ProgressIndicator', () => {
  it('should render progress with percentage', () => {
    render(<ProgressIndicator progress={50} />);
    expect(screen.getByText(/50%/)).toBeInTheDocument();
  });

  it('should render with label', () => {
    render(<ProgressIndicator progress={75} label="Loading..." />);
    expect(screen.getByText('Loading...')).toBeInTheDocument();
    expect(screen.getByText(/75%/)).toBeInTheDocument();
  });

  it('should clamp progress to 0-100', () => {
    const { rerender } = render(<ProgressIndicator progress={150} label="Test" />);
    expect(screen.getByText(/100%/)).toBeInTheDocument();

    rerender(<ProgressIndicator progress={-10} label="Test" />);
    expect(screen.getByText(/0%/)).toBeInTheDocument();
  });

  it('should hide percentage when showPercentage is false', () => {
    render(<ProgressIndicator progress={50} showPercentage={false} />);
    expect(screen.queryByText(/50%/)).not.toBeInTheDocument();
  });
});

