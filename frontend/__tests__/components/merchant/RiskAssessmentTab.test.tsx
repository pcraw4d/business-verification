
import { render, screen, waitFor, act } from '@testing-library/react';
import { RiskAssessmentTab } from '@/components/merchant/RiskAssessmentTab';

// Mock API
const mockGetRiskAssessment = vi.fn();
const mockStartRiskAssessment = vi.fn();
const mockGetAssessmentStatus = vi.fn();
vi.mock('@/lib/api', () => ({
  getRiskAssessment: (...args: any[]) => mockGetRiskAssessment(...args),
  startRiskAssessment: (...args: any[]) => mockStartRiskAssessment(...args),
  getAssessmentStatus: (...args: any[]) => mockGetAssessmentStatus(...args),
}));

// Mock ErrorHandler
vi.mock('@/lib/error-handler', () => ({
  ErrorHandler: {
    handleAPIError: vi.fn().mockResolvedValue(undefined),
    showErrorNotification: vi.fn(),
    showSuccessNotification: vi.fn(),
    showInfoNotification: vi.fn(),
    parseErrorResponse: vi.fn(),
    logError: vi.fn(),
  },
}));

// Mock toast
vi.mock('sonner', () => ({
  toast: {
    info: vi.fn(),
    success: vi.fn(),
    error: vi.fn(),
  },
}));

describe('RiskAssessmentTab', () => {
  const mockAssessment = {
    id: 'assessment-123',
    merchantId: 'merchant-123',
    status: 'completed' as const,
    progress: 100,
    options: {
      includeHistory: true,
      includePredictions: true,
    },
    result: {
      overallScore: 0.7,
      riskLevel: 'medium',
      factors: [],
    },
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  };

  beforeEach(async () => {
    vi.clearAllMocks();
    mockGetRiskAssessment.mockClear();
    mockStartRiskAssessment.mockClear();
    mockGetAssessmentStatus.mockClear();
    
    // Clear API cache and request deduplicator to ensure test isolation
    // RiskAssessmentTab uses mocked API functions, but we still need to clear
    // any real API cache/deduplicator that might be used by other components
    try {
      const { getAPICache, getRequestDeduplicator } = await import('@/lib/api');
      const apiCache = getAPICache();
      const requestDeduplicator = getRequestDeduplicator();
      
      if (apiCache && typeof apiCache.clear === 'function') {
        apiCache.clear();
      }
      if (requestDeduplicator && typeof requestDeduplicator.clear === 'function') {
        requestDeduplicator.clear();
      }
    } catch (error) {
      // Ignore errors if module not loaded yet
    }
  });

  it('should render loading state initially', () => {
    mockGetRiskAssessment.mockImplementation(
      () => new Promise(() => {})
    );

    const { container } = render(<RiskAssessmentTab merchantId="merchant-123" />);

    // Check for loading indicators - Skeleton component uses data-slot="skeleton"
    const skeletons = container.querySelectorAll('[data-slot="skeleton"], [class*="skeleton"], [class*="Skeleton"]');
    // Component should render something while loading
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it('should render assessment data when loaded', async () => {
    mockGetRiskAssessment.mockResolvedValue(mockAssessment);

    render(<RiskAssessmentTab merchantId="merchant-123" />);

    await waitFor(() => {
      expect(screen.getByText(/completed/i)).toBeInTheDocument();
    });

    expect(screen.getByText(/medium/i)).toBeInTheDocument();
  });

  it('should render empty state when no assessment', async () => {
    // Ensure mock resolves immediately
    mockGetRiskAssessment.mockResolvedValue(null);

    render(<RiskAssessmentTab merchantId="merchant-123" />);

    // Wait for loading to complete first
    await act(async () => {
      await waitFor(() => {
        const loadingElements = document.querySelectorAll('[class*="skeleton"], [class*="Skeleton"]');
        if (loadingElements.length > 0) {
          throw new Error('Still loading');
        }
      }, { timeout: 2000 });
    });

    // Wait for the empty state to appear - component should show empty state when assessment is null
    // The component checks: if (!assessment && !processing) -> show empty state
    // EmptyState component renders the title in an h3 tag, so use findByRole
    const emptyStateHeading = await screen.findByRole('heading', { name: /No Risk Assessment/i }, { timeout: 5000 });
    expect(emptyStateHeading).toBeInTheDocument();
    
    // Verify the API was called
    expect(mockGetRiskAssessment).toHaveBeenCalledWith('merchant-123');
  });

  it('should render error state on API error', async () => {
    mockGetRiskAssessment.mockRejectedValue(new Error('Failed to load'));

    render(<RiskAssessmentTab merchantId="merchant-123" />);

    await waitFor(() => {
      expect(screen.getByText(/Failed to load/i)).toBeInTheDocument();
    });
  });

  it('should start assessment when button clicked', async () => {
    mockGetRiskAssessment.mockResolvedValue(null);
    mockStartRiskAssessment.mockResolvedValue({
      assessmentId: 'assessment-123',
      status: 'pending',
    });
    mockGetAssessmentStatus.mockResolvedValue({
      assessmentId: 'assessment-123',
      status: 'completed',
      progress: 100,
    });
    mockGetRiskAssessment.mockResolvedValueOnce(null).mockResolvedValueOnce(mockAssessment);

    render(<RiskAssessmentTab merchantId="merchant-123" />);

    await waitFor(() => {
      expect(screen.getByText(/Start Assessment/i)).toBeInTheDocument();
    });

    const button = screen.getByText(/Start Assessment/i);
    button.click();

    await waitFor(() => {
      expect(mockStartRiskAssessment).toHaveBeenCalledWith(
        expect.objectContaining({
          merchantId: 'merchant-123',
        })
      );
    });
  });

  it('should show progress indicator when processing', async () => {
    // Reset the mock to ensure clean state
    mockGetRiskAssessment.mockReset();
    mockGetRiskAssessment.mockResolvedValue({
      ...mockAssessment,
      status: 'processing' as const,
      progress: 50,
      result: undefined, // No result when processing
    });

    render(<RiskAssessmentTab merchantId="merchant-123" />);

    // Wait for loading to complete first - the component shows Skeleton while loading
    await act(async () => {
      await waitFor(() => {
        const loadingElements = document.querySelectorAll('[class*="skeleton"], [class*="Skeleton"]');
        if (loadingElements.length > 0) {
          throw new Error('Still loading');
        }
      }, { timeout: 3000 });
    });

    // Wait for the assessment card to appear - it should show when assessment is loaded
    // The component renders the assessment card when assessment exists
    // The status badge shows "processing" and progress shows "50%"
    // First wait for the "Status" label to appear, then look for the status badge
    // The component renders "Status" as a label, then the Badge with the status text
    const statusLabel = await screen.findByText(/Status/i, {}, { timeout: 10000 });
    expect(statusLabel).toBeInTheDocument();
    
    // Then verify the status badge shows "processing"
    const statusText = screen.getByText(/processing/i);
    expect(statusText).toBeInTheDocument();
    
    // Also verify progress is shown - progress is rendered as "{progress}%"
    const progressText = screen.getByText(/50%/i);
    expect(progressText).toBeInTheDocument();
  });

  it('should display toast notifications for assessment lifecycle', async () => {
    mockGetRiskAssessment.mockResolvedValue(null);
    mockStartRiskAssessment.mockResolvedValue({
      assessmentId: 'assessment-123',
      status: 'pending',
    });

    render(<RiskAssessmentTab merchantId="merchant-123" />);

    await waitFor(() => {
      expect(screen.getByText(/Start Assessment/i)).toBeInTheDocument();
    });

    const button = screen.getByText(/Start Assessment/i);
    button.click();

    // Import toast to access the mock
    const { toast } = await import('sonner');
    await waitFor(() => {
      expect(vi.mocked(toast.info)).toHaveBeenCalledWith('Starting risk assessment...');
    });
  });

  it('should handle assessment status polling', async () => {
    mockGetRiskAssessment.mockResolvedValue(null);
    mockStartRiskAssessment.mockResolvedValue({
      assessmentId: 'assessment-123',
      status: 'pending',
    });
    mockGetAssessmentStatus
      .mockResolvedValueOnce({ status: 'processing', progress: 50 })
      .mockResolvedValueOnce({ status: 'completed', progress: 100 });

    render(<RiskAssessmentTab merchantId="merchant-123" />);

    await waitFor(() => {
      expect(screen.getByText(/Start Assessment/i)).toBeInTheDocument();
    });

    const button = screen.getByText(/Start Assessment/i);
    button.click();

    // Wait for polling to complete
    await waitFor(() => {
      expect(mockGetAssessmentStatus).toHaveBeenCalled();
    }, { timeout: 5000 });
  });

  it('should handle assessment failure', async () => {
    mockGetRiskAssessment.mockResolvedValue(null);
    mockStartRiskAssessment.mockResolvedValue({
      assessmentId: 'assessment-123',
      status: 'pending',
    });
    mockGetAssessmentStatus.mockResolvedValue({
      assessmentId: 'assessment-123',
      status: 'failed',
      progress: 0,
    });

    render(<RiskAssessmentTab merchantId="merchant-123" />);

    // Wait for empty state to appear first
    await act(async () => {
      await waitFor(() => {
        const loadingElements = document.querySelectorAll('[class*="skeleton"], [class*="Skeleton"]');
        if (loadingElements.length > 0) {
          throw new Error('Still loading');
        }
      }, { timeout: 2000 });
    });

    const button = await screen.findByRole('button', { name: /Start.*Assessment/i }, { timeout: 5000 });
    
    // Click the button and wait for the polling to start
    await act(async () => {
      button.click();
    });

    // Wait for the polling interval to run and detect the failed status
    // The component polls every 2 seconds, so we need to wait for that
    await waitFor(() => {
      expect(mockStartRiskAssessment).toHaveBeenCalled();
    }, { timeout: 3000 });

    // Wait for the status polling to detect the failure
    // The component calls getAssessmentStatus in a polling interval (every 2 seconds)
    await waitFor(() => {
      expect(mockGetAssessmentStatus).toHaveBeenCalled();
    }, { timeout: 5000 });

    // Import toast to access the mock
    const { toast } = await import('sonner');
    // The toast.error should be called when status is 'failed'
    await waitFor(() => {
      expect(vi.mocked(toast.error)).toHaveBeenCalledWith('Risk assessment failed');
    }, { timeout: 5000 });
  });
});

