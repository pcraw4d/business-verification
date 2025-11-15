import { describe, it, expect, beforeEach, jest } from '@jest/globals';
import { render, screen, waitFor } from '@testing-library/react';
import { BusinessAnalyticsTab } from '@/components/merchant/BusinessAnalyticsTab';
import * as api from '@/lib/api';
import * as lazyLoader from '@/lib/lazy-loader';

// Mock API
jest.mock('@/lib/api', () => ({
  getMerchantAnalytics: jest.fn(),
  getWebsiteAnalysis: jest.fn(),
}));

jest.mock('@/lib/lazy-loader', () => ({
  deferNonCriticalDataLoad: jest.fn((fn) => fn()),
}));

describe('BusinessAnalyticsTab', () => {
  const mockAnalytics = {
    merchantId: 'merchant-123',
    classification: {
      primaryIndustry: 'Technology',
      confidenceScore: 0.95,
      riskLevel: 'low',
    },
    security: {
      trustScore: 0.8,
      sslValid: true,
    },
    quality: {
      completenessScore: 0.9,
      dataPoints: 100,
    },
    intelligence: {},
    timestamp: new Date().toISOString(),
  };

  const mockWebsiteAnalysis = {
    merchantId: 'merchant-123',
    websiteUrl: 'https://test.com',
    performance: { score: 85 },
    accessibility: { score: 0.9 },
  };

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should render loading state initially', () => {
    (api.getMerchantAnalytics as jest.Mock).mockImplementation(
      () => new Promise(() => {})
    );

    render(<BusinessAnalyticsTab merchantId="merchant-123" />);

    expect(screen.getByRole('status', { hidden: true })).toBeInTheDocument();
  });

  it('should render analytics data when loaded', async () => {
    (api.getMerchantAnalytics as jest.Mock).mockResolvedValue(mockAnalytics);
    (api.getWebsiteAnalysis as jest.Mock).mockResolvedValue(mockWebsiteAnalysis);

    render(<BusinessAnalyticsTab merchantId="merchant-123" />);

    await waitFor(() => {
      expect(screen.getByText('Technology')).toBeInTheDocument();
    });

    expect(screen.getByText(/95.0%/)).toBeInTheDocument();
    expect(screen.getByText(/80.0%/)).toBeInTheDocument();
  });

  it('should render empty state when no data', async () => {
    (api.getMerchantAnalytics as jest.Mock).mockResolvedValue(null);
    (api.getWebsiteAnalysis as jest.Mock).mockResolvedValue(null);

    render(<BusinessAnalyticsTab merchantId="merchant-123" />);

    await waitFor(() => {
      expect(screen.getByText(/No Analytics Data/i)).toBeInTheDocument();
    });
  });

  it('should use lazy loading for website analysis', async () => {
    (api.getMerchantAnalytics as jest.Mock).mockResolvedValue(mockAnalytics);
    (api.getWebsiteAnalysis as jest.Mock).mockResolvedValue(mockWebsiteAnalysis);

    render(<BusinessAnalyticsTab merchantId="merchant-123" />);

    await waitFor(() => {
      expect(lazyLoader.deferNonCriticalDataLoad).toHaveBeenCalled();
    });
  });
});

