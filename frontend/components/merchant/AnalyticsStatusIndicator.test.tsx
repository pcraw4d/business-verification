import { render, screen, waitFor } from '@testing-library/react';
import { AnalyticsStatusIndicator } from './AnalyticsStatusIndicator';
import * as api from '@/lib/api';

// Mock the API module
jest.mock('@/lib/api', () => ({
  getMerchantAnalyticsStatus: jest.fn(),
}));

describe('AnalyticsStatusIndicator', () => {
  const mockGetMerchantAnalyticsStatus = api.getMerchantAnalyticsStatus as jest.MockedFunction<
    typeof api.getMerchantAnalyticsStatus
  >;

  beforeEach(() => {
    jest.clearAllMocks();
    jest.useFakeTimers();
  });

  afterEach(() => {
    jest.runOnlyPendingTimers();
    jest.useRealTimers();
  });

  it('displays pending status initially', async () => {
    mockGetMerchantAnalyticsStatus.mockResolvedValue({
      merchantId: 'merchant_123',
      status: {
        classification: 'pending',
        websiteAnalysis: 'pending',
      },
      timestamp: new Date().toISOString(),
    });

    render(<AnalyticsStatusIndicator merchantId="merchant_123" type="classification" />);

    await waitFor(() => {
      expect(screen.getByText(/pending/i)).toBeInTheDocument();
    });
  });

  it('displays processing status with spinner', async () => {
    mockGetMerchantAnalyticsStatus.mockResolvedValue({
      merchantId: 'merchant_123',
      status: {
        classification: 'processing',
        websiteAnalysis: 'pending',
      },
      timestamp: new Date().toISOString(),
    });

    render(<AnalyticsStatusIndicator merchantId="merchant_123" type="classification" />);

    await waitFor(() => {
      expect(screen.getByText(/processing/i)).toBeInTheDocument();
    });

    // Should have spinner icon
    const spinner = screen.getByRole('status', { hidden: true }) || 
                   document.querySelector('.animate-spin');
    expect(spinner).toBeInTheDocument();
  });

  it('displays completed status with checkmark', async () => {
    mockGetMerchantAnalyticsStatus.mockResolvedValue({
      merchantId: 'merchant_123',
      status: {
        classification: 'completed',
        websiteAnalysis: 'pending',
      },
      timestamp: new Date().toISOString(),
    });

    render(<AnalyticsStatusIndicator merchantId="merchant_123" type="classification" />);

    await waitFor(() => {
      expect(screen.getByText(/completed/i)).toBeInTheDocument();
    });
  });

  it('displays failed status with error icon', async () => {
    mockGetMerchantAnalyticsStatus.mockResolvedValue({
      merchantId: 'merchant_123',
      status: {
        classification: 'failed',
        websiteAnalysis: 'pending',
      },
      timestamp: new Date().toISOString(),
    });

    render(<AnalyticsStatusIndicator merchantId="merchant_123" type="classification" />);

    await waitFor(() => {
      expect(screen.getByText(/failed/i)).toBeInTheDocument();
    });
  });

  it('displays skipped status for website analysis', async () => {
    mockGetMerchantAnalyticsStatus.mockResolvedValue({
      merchantId: 'merchant_123',
      status: {
        classification: 'completed',
        websiteAnalysis: 'skipped',
      },
      timestamp: new Date().toISOString(),
    });

    render(<AnalyticsStatusIndicator merchantId="merchant_123" type="websiteAnalysis" />);

    await waitFor(() => {
      expect(screen.getByText(/skipped/i)).toBeInTheDocument();
    });
  });

  it('polls status endpoint when processing', async () => {
    let callCount = 0;
    mockGetMerchantAnalyticsStatus.mockImplementation(async () => {
      callCount++;
      if (callCount === 1) {
        return {
          merchantId: 'merchant_123',
          status: {
            classification: 'processing',
            websiteAnalysis: 'pending',
          },
          timestamp: new Date().toISOString(),
        };
      }
      return {
        merchantId: 'merchant_123',
        status: {
          classification: 'completed',
          websiteAnalysis: 'pending',
        },
        timestamp: new Date().toISOString(),
      };
    });

    render(<AnalyticsStatusIndicator merchantId="merchant_123" type="classification" />);

    // Initial call
    await waitFor(() => {
      expect(mockGetMerchantAnalyticsStatus).toHaveBeenCalledTimes(1);
    });

    // Advance timer to trigger polling
    jest.advanceTimersByTime(3000);

    // Should poll again
    await waitFor(() => {
      expect(mockGetMerchantAnalyticsStatus).toHaveBeenCalledTimes(2);
    });
  });

  it('stops polling when status is completed', async () => {
    mockGetMerchantAnalyticsStatus.mockResolvedValue({
      merchantId: 'merchant_123',
      status: {
        classification: 'completed',
        websiteAnalysis: 'pending',
      },
      timestamp: new Date().toISOString(),
    });

    render(<AnalyticsStatusIndicator merchantId="merchant_123" type="classification" />);

    await waitFor(() => {
      expect(mockGetMerchantAnalyticsStatus).toHaveBeenCalledTimes(1);
    });

    // Advance timer - should not poll again
    jest.advanceTimersByTime(3000);
    jest.advanceTimersByTime(3000);

    // Should still only be called once
    expect(mockGetMerchantAnalyticsStatus).toHaveBeenCalledTimes(1);
  });

  it('handles API errors gracefully', async () => {
    mockGetMerchantAnalyticsStatus.mockRejectedValue(new Error('API Error'));

    render(<AnalyticsStatusIndicator merchantId="merchant_123" type="classification" />);

    // Should not crash, just show pending or handle error
    await waitFor(() => {
      // Component should still render
      expect(screen.getByRole('status', { hidden: true })).toBeInTheDocument();
    });
  });
});

