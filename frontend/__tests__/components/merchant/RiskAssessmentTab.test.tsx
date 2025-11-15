import { describe, it, expect, beforeEach, jest } from '@jest/globals';
import { render, screen, waitFor } from '@testing-library/react';
import { RiskAssessmentTab } from '@/components/merchant/RiskAssessmentTab';

// Mock API
const mockGetRiskAssessment = jest.fn();
const mockStartRiskAssessment = jest.fn();
const mockGetAssessmentStatus = jest.fn();
jest.mock('@/lib/api', () => ({
  getRiskAssessment: (...args: any[]) => mockGetRiskAssessment(...args),
  startRiskAssessment: (...args: any[]) => mockStartRiskAssessment(...args),
  getAssessmentStatus: (...args: any[]) => mockGetAssessmentStatus(...args),
}));

// Mock ErrorHandler
const mockHandleAPIError = jest.fn().mockResolvedValue(undefined);
jest.mock('@/lib/error-handler', () => ({
  ErrorHandler: {
    handleAPIError: mockHandleAPIError,
  },
}));

// Mock toast
const mockToastInfo = jest.fn();
const mockToastSuccess = jest.fn();
const mockToastError = jest.fn();
jest.mock('sonner', () => ({
  toast: {
    info: mockToastInfo,
    success: mockToastSuccess,
    error: mockToastError,
  },
}));

describe('RiskAssessmentTab', () => {
  const mockAssessment = {
    id: 'assessment-123',
    merchantId: 'merchant-123',
    status: 'completed',
    progress: 100,
    result: {
      overallScore: 0.7,
      riskLevel: 'medium',
      factors: [],
    },
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  };

  beforeEach(() => {
    jest.clearAllMocks();
    mockGetRiskAssessment.mockClear();
    mockStartRiskAssessment.mockClear();
    mockGetAssessmentStatus.mockClear();
    mockHandleAPIError.mockClear();
    mockToastInfo.mockClear();
    mockToastSuccess.mockClear();
    mockToastError.mockClear();
  });

  it('should render loading state initially', () => {
    mockGetRiskAssessment.mockImplementation(
      () => new Promise(() => {})
    );

    render(<RiskAssessmentTab merchantId="merchant-123" />);

    expect(screen.getByRole('status', { hidden: true })).toBeInTheDocument();
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
    mockGetRiskAssessment.mockResolvedValue(null);

    render(<RiskAssessmentTab merchantId="merchant-123" />);

    await waitFor(() => {
      expect(screen.getByText(/No Risk Assessment/i)).toBeInTheDocument();
    });
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
    mockGetRiskAssessment.mockResolvedValue({
      ...mockAssessment,
      status: 'processing',
      progress: 50,
    });

    render(<RiskAssessmentTab merchantId="merchant-123" />);

    await waitFor(() => {
      expect(screen.getByText(/Processing Assessment/i)).toBeInTheDocument();
    });
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

    await waitFor(() => {
      expect(mockToastInfo).toHaveBeenCalledWith('Starting risk assessment...');
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

    await waitFor(() => {
      expect(screen.getByText(/Start Assessment/i)).toBeInTheDocument();
    });

    const button = screen.getByText(/Start Assessment/i);
    button.click();

    await waitFor(() => {
      expect(mockToastError).toHaveBeenCalledWith('Risk assessment failed');
    });
  });
});

