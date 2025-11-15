import { describe, it, expect, beforeEach, jest } from '@jest/globals';
import { render, screen, waitFor } from '@testing-library/react';
import { RiskIndicatorsTab } from '@/components/merchant/RiskIndicatorsTab';

// Mock API
const mockGetRiskIndicators = jest.fn();
jest.mock('@/lib/api', () => ({
  getRiskIndicators: (...args: any[]) => mockGetRiskIndicators(...args),
}));

// Mock lazy loader
const mockDeferNonCriticalDataLoad = jest.fn((fn) => fn());
jest.mock('@/lib/lazy-loader', () => ({
  deferNonCriticalDataLoad: (...args: any[]) => mockDeferNonCriticalDataLoad(...args),
}));

describe('RiskIndicatorsTab', () => {
  const mockIndicators = {
    merchantId: 'merchant-123',
    overallScore: 0.7,
    indicators: [
      {
        id: 'indicator-1',
        title: 'High Risk Factor',
        description: 'This is a high risk indicator',
        severity: 'high',
        status: 'active',
        category: 'financial',
        detectedAt: new Date().toISOString(),
      },
      {
        id: 'indicator-2',
        title: 'Critical Risk Factor',
        description: 'This is a critical risk indicator',
        severity: 'critical',
        status: 'active',
        category: 'compliance',
        detectedAt: new Date().toISOString(),
      },
    ],
    lastUpdated: new Date().toISOString(),
  };

  beforeEach(() => {
    jest.clearAllMocks();
    mockGetRiskIndicators.mockClear();
    mockDeferNonCriticalDataLoad.mockClear();
  });

  it('should render loading state initially', () => {
    mockGetRiskIndicators.mockImplementation(
      () => new Promise(() => {})
    );

    render(<RiskIndicatorsTab merchantId="merchant-123" />);

    expect(screen.getByRole('status', { hidden: true })).toBeInTheDocument();
  });

  it('should render indicators when loaded', async () => {
    mockGetRiskIndicators.mockResolvedValue(mockIndicators);

    render(<RiskIndicatorsTab merchantId="merchant-123" />);

    await waitFor(() => {
      expect(screen.getByText('High Risk Factor')).toBeInTheDocument();
    });

    expect(screen.getByText('Critical Risk Factor')).toBeInTheDocument();
  });

  it('should render empty state when no indicators', async () => {
    mockGetRiskIndicators.mockResolvedValue({
      merchantId: 'merchant-123',
      overallScore: 0.5,
      indicators: [],
      lastUpdated: new Date().toISOString(),
    });

    render(<RiskIndicatorsTab merchantId="merchant-123" />);

    await waitFor(() => {
      expect(screen.getByText(/No Active Indicators/i)).toBeInTheDocument();
    });
  });

  it('should render error state on API error', async () => {
    mockGetRiskIndicators.mockRejectedValue(new Error('Failed to load'));

    render(<RiskIndicatorsTab merchantId="merchant-123" />);

    await waitFor(() => {
      expect(screen.getByText(/Failed to load/i)).toBeInTheDocument();
    });
  });

  it('should use lazy loading for indicators', async () => {
    mockGetRiskIndicators.mockResolvedValue(mockIndicators);

    render(<RiskIndicatorsTab merchantId="merchant-123" />);

    await waitFor(() => {
      expect(mockDeferNonCriticalDataLoad).toHaveBeenCalled();
    });
  });

  it('should display severity badges correctly', async () => {
    mockGetRiskIndicators.mockResolvedValue(mockIndicators);

    render(<RiskIndicatorsTab merchantId="merchant-123" />);

    await waitFor(() => {
      expect(screen.getByText('high')).toBeInTheDocument();
      expect(screen.getByText('critical')).toBeInTheDocument();
    });
  });

  it('should allow retry on error', async () => {
    mockGetRiskIndicators.mockRejectedValueOnce(new Error('Failed to load'))
      .mockResolvedValueOnce(mockIndicators);

    render(<RiskIndicatorsTab merchantId="merchant-123" />);

    await waitFor(() => {
      expect(screen.getByText(/Retry/i)).toBeInTheDocument();
    });

    const retryButton = screen.getByText(/Retry/i);
    retryButton.click();

    await waitFor(() => {
      expect(mockGetRiskIndicators).toHaveBeenCalledTimes(2);
    });
  });
});

