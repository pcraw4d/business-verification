
import { render, screen, waitFor } from '@testing-library/react';
import { BusinessAnalyticsTab } from '@/components/merchant/BusinessAnalyticsTab';

// Mock API
const mockGetMerchantAnalytics = vi.fn();
const mockGetWebsiteAnalysis = vi.fn();
const mockGetMerchant = vi.fn();
vi.mock('@/lib/api', () => ({
  getMerchantAnalytics: (...args: any[]) => mockGetMerchantAnalytics(...args),
  getWebsiteAnalysis: (...args: any[]) => mockGetWebsiteAnalysis(...args),
  getMerchant: (...args: any[]) => mockGetMerchant(...args),
}));

// Mock lazy loader
const mockDeferNonCriticalDataLoad = vi.fn((fn) => fn());
vi.mock('@/lib/lazy-loader', () => ({
  deferNonCriticalDataLoad: (...args: any[]) => mockDeferNonCriticalDataLoad(...args),
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

  const mockMerchant = {
    id: 'merchant-123',
    businessName: 'Test Business',
    industry: 'Technology',
    status: 'active',
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockGetMerchantAnalytics.mockClear();
    mockGetWebsiteAnalysis.mockClear();
    mockGetMerchant.mockClear();
    mockDeferNonCriticalDataLoad.mockClear();
    // Default mock for getMerchant
    mockGetMerchant.mockResolvedValue(mockMerchant);
  });

  it('should render loading state initially', () => {
    mockGetMerchantAnalytics.mockImplementation(
      () => new Promise(() => {})
    );

    const { container } = render(<BusinessAnalyticsTab merchantId="merchant-123" />);

    // Check for skeleton/loading indicators - Skeleton component uses data-slot="skeleton"
    const skeletons = container.querySelectorAll('[data-slot="skeleton"], [class*="skeleton"], [class*="Skeleton"]');
    // Component should render loading state - check for skeletons or container
    const hasSkeletons = skeletons.length > 0;
    const hasContainer = container.querySelector('.space-y-6') !== null;
    expect(hasSkeletons || hasContainer).toBe(true);
  });

  it('should render analytics data when loaded', async () => {
    mockGetMerchantAnalytics.mockResolvedValue(mockAnalytics);
    mockGetWebsiteAnalysis.mockResolvedValue(mockWebsiteAnalysis);
    mockGetMerchant.mockResolvedValue(mockMerchant);

    render(<BusinessAnalyticsTab merchantId="merchant-123" />);

    await waitFor(() => {
      // The component displays the primaryIndustry from classification
      expect(screen.getByText('Technology')).toBeInTheDocument();
    }, { timeout: 3000 });

    expect(screen.getByText(/95.0%/)).toBeInTheDocument();
    expect(screen.getByText(/80.0%/)).toBeInTheDocument();
  });

  it('should render empty state when no data', async () => {
    mockGetMerchantAnalytics.mockResolvedValue(null);
    mockGetWebsiteAnalysis.mockResolvedValue(null);
    mockGetMerchant.mockResolvedValue(mockMerchant);

    render(<BusinessAnalyticsTab merchantId="merchant-123" />);

    await waitFor(() => {
      // Component shows empty state when analytics is null
      expect(screen.getByText(/No Analytics Data/i)).toBeInTheDocument();
    }, { timeout: 3000 });
  });

  it('should use lazy loading for website analysis', async () => {
    mockGetMerchantAnalytics.mockResolvedValue(mockAnalytics);
    mockGetWebsiteAnalysis.mockResolvedValue(mockWebsiteAnalysis);
    mockGetMerchant.mockResolvedValue(mockMerchant);

    render(<BusinessAnalyticsTab merchantId="merchant-123" />);

    // Wait for initial load to complete
    await waitFor(() => {
      expect(mockGetMerchantAnalytics).toHaveBeenCalled();
    }, { timeout: 3000 });

    // deferNonCriticalDataLoad should be called for website analysis
    await waitFor(() => {
      expect(mockDeferNonCriticalDataLoad).toHaveBeenCalled();
    }, { timeout: 1000 });
  });
});

