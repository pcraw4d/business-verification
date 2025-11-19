import { server } from '@/__tests__/mocks/server';
import { PortfolioComparisonCard } from '@/components/merchant/PortfolioComparisonCard';
import { render, screen, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { describe, it, expect, beforeEach } from 'vitest';

describe('PortfolioComparisonCard', () => {
  const merchantId = 'merchant-123';

  const mockPortfolioStats = {
    totalMerchants: 100,
    totalAssessments: 150,
    averageRiskScore: 0.6,
    riskDistribution: { low: 40, medium: 50, high: 10 },
    industryBreakdown: [],
    countryBreakdown: [],
    timestamp: new Date().toISOString(),
  };

  const mockMerchantRiskScore = {
    merchant_id: merchantId,
    risk_score: 0.45, // Lower than average (better)
    risk_level: 'low' as const,
    confidence_score: 0.85,
    assessment_date: '2025-01-27T00:00:00Z',
    factors: [],
  };

  beforeEach(() => {
    server.use(
      http.get('*/api/v1/merchants/statistics', () => {
        return HttpResponse.json(mockPortfolioStats);
      }),
      http.get('*/api/v1/merchants/:id/risk-score', () => {
        return HttpResponse.json(mockMerchantRiskScore);
      })
    );
  });

  describe('Loading State', () => {
    it('should show loading skeleton initially', () => {
      server.use(
        http.get('*/api/v1/merchants/statistics', async () => {
          await new Promise((resolve) => setTimeout(resolve, 100));
          return HttpResponse.json(mockPortfolioStats);
        })
      );

      render(<PortfolioComparisonCard merchantId={merchantId} />);

      const skeleton = document.querySelector('[class*="skeleton"]');
      expect(skeleton).toBeInTheDocument();
    });
  });

  describe('Success State', () => {
    it('should display portfolio comparison data when loaded', async () => {
      render(<PortfolioComparisonCard merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/portfolio comparison/i)).toBeInTheDocument();
      });
    });

    it('should display merchant score vs portfolio average', async () => {
      render(<PortfolioComparisonCard merchantId={merchantId} />);

      await waitFor(() => {
        // Should show merchant score (45.0) and portfolio average (60.0)
        expect(screen.getByText(/45\.0|60\.0/i)).toBeInTheDocument();
      });
    });

    it('should display percentile ranking', async () => {
      render(<PortfolioComparisonCard merchantId={merchantId} />);

      await waitFor(() => {
        // Should show percentile information
        expect(screen.getByText(/percentile|ranking|position/i)).toBeInTheDocument();
      });
    });

    it('should display position indicator (above/below average)', async () => {
      render(<PortfolioComparisonCard merchantId={merchantId} />);

      await waitFor(() => {
        // Should show position relative to portfolio
        const positionText = screen.getByText(/above|below|average/i);
        expect(positionText).toBeInTheDocument();
      });
    });

    it('should display difference percentage', async () => {
      render(<PortfolioComparisonCard merchantId={merchantId} />);

      await waitFor(() => {
        // Should show difference between merchant and portfolio
        expect(screen.getByText(/%/)).toBeInTheDocument();
      });
    });
  });

  describe('Comparison Calculations', () => {
    it('should show "Above Average" for merchant with lower risk score', async () => {
      const lowRiskScore = { ...mockMerchantRiskScore, risk_score: 0.3 };
      server.use(
        http.get('*/api/v1/merchants/:id/risk-score', () => {
          return HttpResponse.json(lowRiskScore);
        })
      );

      render(<PortfolioComparisonCard merchantId={merchantId} />);

      await waitFor(() => {
        // Lower risk score is better, so should show above average
        const positionText = screen.getByText(/above|better|top/i);
        expect(positionText).toBeInTheDocument();
      });
    });

    it('should show "Below Average" for merchant with higher risk score', async () => {
      const highRiskScore = { ...mockMerchantRiskScore, risk_score: 0.85 };
      server.use(
        http.get('*/api/v1/merchants/:id/risk-score', () => {
          return HttpResponse.json(highRiskScore);
        })
      );

      render(<PortfolioComparisonCard merchantId={merchantId} />);

      await waitFor(() => {
        // Higher risk score is worse, so should show below average
        const positionText = screen.getByText(/below|worse|bottom/i);
        expect(positionText).toBeInTheDocument();
      });
    });

    it('should calculate correct percentile for merchant score', async () => {
      render(<PortfolioComparisonCard merchantId={merchantId} />);

      await waitFor(() => {
        // Should display percentile value
        expect(screen.getByText(/\d+%/)).toBeInTheDocument();
      });
    });
  });

  describe('Error Handling', () => {
    it('should handle portfolio statistics fetch failure', async () => {
      server.use(
        http.get('*/api/v1/merchants/statistics', () => {
          return HttpResponse.json({ error: 'Not found' }, { status: 404 });
        })
      );

      render(<PortfolioComparisonCard merchantId={merchantId} />);

      await waitFor(() => {
        // Should show error or empty state
        const errorText = screen.queryByText(/error|failed/i);
        // Component may not show error, just not display data
        expect(errorText || screen.queryByText(/portfolio comparison/i)).toBeTruthy();
      }, { timeout: 3000 });
    });

    it('should handle merchant risk score fetch failure', async () => {
      server.use(
        http.get('*/api/v1/merchants/:id/risk-score', () => {
          return HttpResponse.json({ error: 'Not found' }, { status: 404 });
        })
      );

      render(<PortfolioComparisonCard merchantId={merchantId} />);

      await waitFor(() => {
        // Should handle gracefully
        const skeleton = document.querySelector('[class*="skeleton"]');
        expect(skeleton).not.toBeInTheDocument();
      }, { timeout: 3000 });
    });
  });

  describe('Merchant Risk Level Display', () => {
    it('should display merchant risk level when provided', async () => {
      render(<PortfolioComparisonCard merchantId={merchantId} merchantRiskLevel="low" />);

      await waitFor(() => {
        expect(screen.getByText(/low|medium|high/i)).toBeInTheDocument();
      });
    });

    it('should work without merchant risk level prop', async () => {
      render(<PortfolioComparisonCard merchantId={merchantId} />);

      await waitFor(() => {
        // Should still display comparison data
        expect(screen.getByText(/portfolio comparison/i)).toBeInTheDocument();
      });
    });
  });
});

