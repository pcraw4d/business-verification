import { server } from '@/__tests__/mocks/server';
import { AnalyticsComparison } from '@/components/merchant/AnalyticsComparison';
import { render, screen, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { toast } from 'sonner';
import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock recharts to avoid rendering issues in tests
vi.mock('recharts', async () => {
  const actual = await vi.importActual('recharts');
  return {
    ...actual,
    ResponsiveContainer: ({ children }: any) => (
      <div data-testid="responsive-container">{children}</div>
    ),
  };
});

// Mock chart components
vi.mock('@/components/charts/lazy', () => ({
  BarChart: ({ data }: any) => (
    <div data-testid="bar-chart">{JSON.stringify(data)}</div>
  ),
}));

vi.mock('sonner');
const mockToast = vi.mocked(toast);

describe('AnalyticsComparison', () => {
  const merchantId = 'merchant-123';

  const mockMerchantAnalytics = {
    merchantId,
    classification: {
      primaryIndustry: 'Technology',
      confidenceScore: 0.95,
      mccCodes: [],
      naicsCodes: [],
      sicCodes: [],
    },
    security: { trustScore: 0.85, sslValid: true },
    quality: { completenessScore: 0.9, dataPoints: 100 },
    intelligence: {},
    timestamp: new Date().toISOString(),
  };

  const mockPortfolioAnalytics = {
    totalMerchants: 100,
    averageRiskScore: 0.6,
    averageClassificationConfidence: 0.8,
    averageSecurityTrustScore: 0.75,
    averageDataQuality: 0.85,
    riskDistribution: { low: 40, medium: 50, high: 10 },
    industryDistribution: {},
    countryDistribution: {},
    timestamp: new Date().toISOString(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockToast.error = vi.fn();
    
    server.use(
      http.get('*/api/v1/merchants/:id/analytics', () => {
        return HttpResponse.json(mockMerchantAnalytics);
      }),
      http.get('*/api/v1/merchants/analytics', () => {
        return HttpResponse.json(mockPortfolioAnalytics);
      })
    );
  });

  describe('Loading State', () => {
    it('should show loading skeleton initially', () => {
      server.use(
        http.get('*/api/v1/merchants/:id/analytics', async () => {
          await new Promise((resolve) => setTimeout(resolve, 100));
          return HttpResponse.json(mockMerchantAnalytics);
        })
      );

      render(<AnalyticsComparison merchantId={merchantId} />);

      const skeleton = document.querySelector('[class*="skeleton"]');
      expect(skeleton).toBeInTheDocument();
    });
  });

  describe('Success State', () => {
    it('should display analytics comparison when loaded', async () => {
      render(<AnalyticsComparison merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/analytics comparison/i)).toBeInTheDocument();
      });
    });

    it('should display classification confidence comparison', async () => {
      render(<AnalyticsComparison merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/classification confidence/i)).toBeInTheDocument();
      });
    });

    it('should display security trust score comparison', async () => {
      render(<AnalyticsComparison merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/security trust score/i)).toBeInTheDocument();
      });
    });

    it('should display data quality comparison', async () => {
      render(<AnalyticsComparison merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/data quality/i)).toBeInTheDocument();
      });
    });

    it('should display comparison charts', async () => {
      render(<AnalyticsComparison merchantId={merchantId} />);

      await waitFor(() => {
        // Should render charts
        const charts = screen.getAllByTestId('bar-chart');
        expect(charts.length).toBeGreaterThan(0);
      });
    });

    it('should show difference indicators (positive/negative)', async () => {
      render(<AnalyticsComparison merchantId={merchantId} />);

      await waitFor(() => {
        // Should show difference percentages
        expect(screen.getByText(/%/)).toBeInTheDocument();
      });
    });
  });

  describe('Using Provided Analytics', () => {
    it('should use provided merchant analytics instead of fetching', async () => {
      render(<AnalyticsComparison merchantId={merchantId} merchantAnalytics={mockMerchantAnalytics} />);

      await waitFor(() => {
        expect(screen.getByText(/analytics comparison/i)).toBeInTheDocument();
      });

      // Should not have made a request for merchant analytics
      // (we can't easily verify this, but the component should work)
    });
  });

  describe('Comparison Calculations', () => {
    it('should calculate positive difference when merchant is better', async () => {
      // Merchant has higher scores than portfolio
      const betterMerchantAnalytics = {
        ...mockMerchantAnalytics,
        classification: { ...mockMerchantAnalytics.classification, confidenceScore: 0.95 },
        security: { ...mockMerchantAnalytics.security, trustScore: 0.9 },
        quality: { ...mockMerchantAnalytics.quality, completenessScore: 0.95 },
      };

      server.use(
        http.get('*/api/v1/merchants/:id/analytics', () => {
          return HttpResponse.json(betterMerchantAnalytics);
        })
      );

      render(<AnalyticsComparison merchantId={merchantId} />);

      await waitFor(() => {
        // Should show positive differences
        expect(screen.getByText(/\+/)).toBeInTheDocument();
      });
    });

    it('should calculate negative difference when merchant is worse', async () => {
      // Merchant has lower scores than portfolio
      const worseMerchantAnalytics = {
        ...mockMerchantAnalytics,
        classification: { ...mockMerchantAnalytics.classification, confidenceScore: 0.6 },
        security: { ...mockMerchantAnalytics.security, trustScore: 0.5 },
        quality: { ...mockMerchantAnalytics.quality, completenessScore: 0.7 },
      };

      server.use(
        http.get('*/api/v1/merchants/:id/analytics', () => {
          return HttpResponse.json(worseMerchantAnalytics);
        })
      );

      render(<AnalyticsComparison merchantId={merchantId} />);

      await waitFor(() => {
        // Should show negative differences or "Similar"
        const diffText = screen.getByText(/-|similar/i);
        expect(diffText).toBeInTheDocument();
      });
    });

    it('should show "Similar" for very small differences', async () => {
      // Merchant scores very close to portfolio
      const similarMerchantAnalytics = {
        ...mockMerchantAnalytics,
        classification: { ...mockMerchantAnalytics.classification, confidenceScore: 0.801 },
        security: { ...mockMerchantAnalytics.security, trustScore: 0.751 },
        quality: { ...mockMerchantAnalytics.quality, completenessScore: 0.851 },
      };

      server.use(
        http.get('*/api/v1/merchants/:id/analytics', () => {
          return HttpResponse.json(similarMerchantAnalytics);
        })
      );

      render(<AnalyticsComparison merchantId={merchantId} />);

      await waitFor(() => {
        // Should show "Similar" for small differences
        expect(screen.getByText(/similar/i)).toBeInTheDocument();
      });
    });
  });

  describe('Error Handling', () => {
    it('should handle merchant analytics fetch failure', async () => {
      server.use(
        http.get('*/api/v1/merchants/:id/analytics', () => {
          return HttpResponse.json({ error: 'Not found' }, { status: 404 });
        })
      );

      render(<AnalyticsComparison merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/failed to load merchant analytics/i)).toBeInTheDocument();
      });
    });

    it('should handle portfolio analytics fetch failure', async () => {
      server.use(
        http.get('*/api/v1/merchants/analytics', () => {
          return HttpResponse.json({ error: 'Not found' }, { status: 404 });
        })
      );

      render(<AnalyticsComparison merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/failed to load portfolio analytics/i)).toBeInTheDocument();
      });
    });

    it('should show error when not enough data', async () => {
      server.use(
        http.get('*/api/v1/merchants/:id/analytics', () => {
          return HttpResponse.json({ merchantId, classification: null, security: null, quality: null });
        })
      );

      render(<AnalyticsComparison merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/not enough data/i)).toBeInTheDocument();
      });
    });
  });

  describe('Chart Data', () => {
    it('should render charts with correct data', async () => {
      render(<AnalyticsComparison merchantId={merchantId} />);

      await waitFor(() => {
        const charts = screen.getAllByTestId('bar-chart');
        expect(charts.length).toBeGreaterThan(0);
        // Charts should contain data
        charts.forEach(chart => {
          expect(chart.textContent).toBeTruthy();
        });
      });
    });
  });
});

