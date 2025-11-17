import { Progress } from '@/components/ui/progress';
import { ProgressIndicator } from '@/components/ui/progress-indicator';
import { Skeleton } from '@/components/ui/skeleton';
import { render, screen } from '@testing-library/react';

describe('Loading States', () => {
  describe('Skeleton Loader', () => {
    it('should render skeleton component', () => {
      render(<Skeleton className="h-4 w-full" />);
      
      const skeleton = document.querySelector('[data-slot="skeleton"]');
      expect(skeleton).toBeInTheDocument();
    });

    it('should apply custom className', () => {
      render(<Skeleton className="h-8 w-32" />);
      
      const skeleton = document.querySelector('[data-slot="skeleton"]');
      expect(skeleton).toHaveClass('h-8', 'w-32');
    });

    it('should have pulse animation', () => {
      render(<Skeleton />);
      
      const skeleton = document.querySelector('[data-slot="skeleton"]');
      expect(skeleton).toHaveClass('animate-pulse');
    });

    it('should render multiple skeletons', () => {
      render(
        <div>
          <Skeleton className="h-4 w-full" />
          <Skeleton className="h-4 w-full" />
          <Skeleton className="h-4 w-full" />
        </div>
      );
      
      const skeletons = document.querySelectorAll('[data-slot="skeleton"]');
      expect(skeletons.length).toBe(3);
    });
  });

  describe('Progress Indicator', () => {
    it('should render progress indicator', () => {
      render(<ProgressIndicator progress={50} />);
      
      const progress = document.querySelector('[data-slot="progress"]');
      expect(progress).toBeInTheDocument();
    });

    it('should display progress percentage', () => {
      render(<ProgressIndicator progress={75} showPercentage={true} />);
      
      expect(screen.getByText('75%')).toBeInTheDocument();
    });

    it('should hide percentage when showPercentage is false', () => {
      render(<ProgressIndicator progress={50} showPercentage={false} />);
      
      expect(screen.queryByText('50%')).not.toBeInTheDocument();
    });

    it('should display label when provided', () => {
      render(<ProgressIndicator progress={50} label="Loading..." />);
      
      expect(screen.getByText('Loading...')).toBeInTheDocument();
    });

    it('should clamp progress to 0-100 range', () => {
      const { rerender } = render(<ProgressIndicator progress={150} />);
      
      expect(screen.getByText('100%')).toBeInTheDocument();
      
      rerender(<ProgressIndicator progress={-10} />);
      
      expect(screen.getByText('0%')).toBeInTheDocument();
    });

    it('should update progress value', () => {
      const { rerender } = render(<ProgressIndicator progress={25} />);
      
      expect(screen.getByText('25%')).toBeInTheDocument();
      
      rerender(<ProgressIndicator progress={75} />);
      
      expect(screen.getByText('75%')).toBeInTheDocument();
    });
  });

  describe('Progress Component', () => {
    it('should render progress bar', () => {
      render(<Progress value={50} />);
      
      const progress = document.querySelector('[data-slot="progress"]');
      expect(progress).toBeInTheDocument();
    });

    it('should display progress value', () => {
      render(<Progress value={75} />);
      
      const indicator = document.querySelector('[data-slot="progress-indicator"]');
      expect(indicator).toHaveStyle({ transform: 'translateX(-25%)' });
    });

    it('should handle 0% progress', () => {
      render(<Progress value={0} />);
      
      const indicator = document.querySelector('[data-slot="progress-indicator"]');
      expect(indicator).toHaveStyle({ transform: 'translateX(-100%)' });
    });

    it('should handle 100% progress', () => {
      render(<Progress value={100} />);
      
      const indicator = document.querySelector('[data-slot="progress-indicator"]');
      expect(indicator).toHaveStyle({ transform: 'translateX(-0%)' });
    });

    it('should handle undefined value', () => {
      render(<Progress value={undefined} />);
      
      const indicator = document.querySelector('[data-slot="progress-indicator"]');
      expect(indicator).toHaveStyle({ transform: 'translateX(-100%)' });
    });
  });

  describe('Progressive Loading', () => {
    it('should show skeleton while loading', () => {
      const isLoading = true;
      
      render(
        <div>
          {isLoading ? (
            <Skeleton className="h-64 w-full" />
          ) : (
            <div>Content loaded</div>
          )}
        </div>
      );
      
      const skeleton = document.querySelector('[data-slot="skeleton"]');
      expect(skeleton).toBeInTheDocument();
      expect(screen.queryByText('Content loaded')).not.toBeInTheDocument();
    });

    it('should show content when loaded', () => {
      const isLoading = false;
      
      render(
        <div>
          {isLoading ? (
            <Skeleton className="h-64 w-full" />
          ) : (
            <div>Content loaded</div>
          )}
        </div>
      );
      
      expect(screen.getByText('Content loaded')).toBeInTheDocument();
      expect(document.querySelector('[data-slot="skeleton"]')).not.toBeInTheDocument();
    });

    it('should show progress during data loading', () => {
      const progress = 60;
      
      render(
        <div>
          <ProgressIndicator progress={progress} label="Loading data..." />
        </div>
      );
      
      expect(screen.getByText('Loading data...')).toBeInTheDocument();
      expect(screen.getByText('60%')).toBeInTheDocument();
    });
  });
});

