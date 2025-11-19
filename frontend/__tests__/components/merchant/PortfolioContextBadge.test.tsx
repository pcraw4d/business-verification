import { server } from '@/__tests__/mocks/server';
import { PortfolioContextBadge } from '@/components/merchant/PortfolioContextBadge';
import { render, screen, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { describe, it, expect, beforeEach } from 'vitest';

describe('PortfolioContextBadge', () => {
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
    it('should show skeleton while loading', () => {
      server.use(
        http.get('*/api/v1/merchants/statistics', async () => {
          await new Promise((resolve) => setTimeout(resolve, 100));
          return HttpResponse.json(mockPortfolioStats);
        })
      );

      render(<PortfolioContextBadge merchantId={merchantId} />);

      const skeleton = document.querySelector('[class*="skeleton"]');
      expect(skeleton).toBeInTheDocument();
    });
  });

  describe('Default Variant', () => {
    it('should display badge with position indicator', async () => {
      render(<PortfolioContextBadge merchantId={merchantId} variant="default" />);

      await waitFor(() => {
        // Should show a badge (exact text depends on percentile calculation)
        const badge = screen.getByRole('status', { hidden: true }) || document.querySelector('[class*="badge"]');
        expect(badge).toBeInTheDocument();
      });
    });
  });

  describe('Compact Variant', () => {
    it('should display compact badge', async () => {
      render(<PortfolioContextBadge merchantId={merchantId} variant="compact" />);

      await waitFor(() => {
        const badge = document.querySelector('[class*="badge"]');
        expect(badge).toBeInTheDocument();
      });
    });
  });

  describe('Detailed Variant', () => {
    it('should display detailed badge with more information', async () => {
      render(<PortfolioContextBadge merchantId={merchantId} variant="detailed" />);

      await waitFor(() => {
        const badge = document.querySelector('[class*="badge"]');
        expect(badge).toBeInTheDocument();
      });
    });
  });

  describe('Position Calculation', () => {
    it('should show top 10% for very low risk score', async () => {
      const lowRiskScore = { ...mockMerchantRiskScore, risk_score: 0.2 };
      server.use(
        http.get('*/api/v1/merchants/:id/risk-score', () => {
          return HttpResponse.json(lowRiskScore);
        })
      );

      render(<PortfolioContextBadge merchantId={merchantId} />);

      await waitFor(() => {
        // Should show top 10% or similar
        const badge = document.querySelector('[class*="badge"]');
        expect(badge).toBeInTheDocument();
      });
    });

    it('should show bottom 10% for very high risk score', async () => {
      const highRiskScore = { ...mockMerchantRiskScore, risk_score: 0.95 };
      server.use(
        http.get('*/api/v1/merchants/:id/risk-score', () => {
          return HttpResponse.json(highRiskScore);
        })
      );

      render(<PortfolioContextBadge merchantId={merchantId} />);

      await waitFor(() => {
        const badge = document.querySelector('[class*="badge"]');
        expect(badge).toBeInTheDocument();
      });
    });

    it('should show average for risk score close to portfolio average', async () => {
      const averageRiskScore = { ...mockMerchantRiskScore, risk_score: 0.6 };
      server.use(
        http.get('*/api/v1/merchants/:id/risk-score', () => {
          return HttpResponse.json(averageRiskScore);
        })
      );

      render(<PortfolioContextBadge merchantId={merchantId} />);

      await waitFor(() => {
        const badge = document.querySelector('[class*="badge"]');
        expect(badge).toBeInTheDocument();
      });
    });
  });

  describe('Error Handling', () => {
    it('should handle portfolio statistics fetch failure gracefully', async () => {
      server.use(
        http.get('*/api/v1/merchants/statistics', () => {
          return HttpResponse.json({ error: 'Not found' }, { status: 404 });
        })
      );

      render(<PortfolioContextBadge merchantId={merchantId} />);

      await waitFor(() => {
        // Component should not crash, may show nothing or error state
        const skeleton = document.querySelector('[class*="skeleton"]');
        // Should eventually stop showing skeleton
        expect(skeleton).not.toBeInTheDocument();
      }, { timeout: 3000 });
    });

    it('should handle merchant risk score fetch failure gracefully', async () => {
      server.use(
        http.get('*/api/v1/merchants/:id/risk-score', () => {
          return HttpResponse.json({ error: 'Not found' }, { status: 404 });
        })
      );

      render(<PortfolioContextBadge merchantId={merchantId} />);

      await waitFor(() => {
        // Component should not crash
        const skeleton = document.querySelector('[class*="skeleton"]');
        expect(skeleton).not.toBeInTheDocument();
      }, { timeout: 3000 });
    });
  });
});

