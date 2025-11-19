import { server } from '@/__tests__/mocks/server';
import { RiskScoreCard } from '@/components/merchant/RiskScoreCard';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { toast } from 'sonner';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('sonner');
const mockToast = vi.mocked(toast);

describe('RiskScoreCard', () => {
  const merchantId = 'merchant-123';

  const mockRiskScore = {
    merchant_id: merchantId,
    risk_score: 0.65,
    risk_level: 'medium' as const,
    confidence_score: 0.85,
    assessment_date: '2025-01-27T00:00:00Z',
    factors: [
      { category: 'Financial Risk', score: 0.7, weight: 0.4 },
      { category: 'Operational Risk', score: 0.6, weight: 0.3 },
      { category: 'Compliance Risk', score: 0.65, weight: 0.3 },
    ],
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockToast.error = vi.fn();
  });

  describe('Loading State', () => {
    it('should show loading skeleton initially', () => {
      server.use(
        http.get('*/api/v1/merchants/:id/risk-score', async () => {
          await new Promise((resolve) => setTimeout(resolve, 100));
          return HttpResponse.json(mockRiskScore);
        })
      );

      render(<RiskScoreCard merchantId={merchantId} />);

      expect(screen.getByText('Risk Score')).toBeInTheDocument();
      expect(screen.getByText('Current merchant risk assessment')).toBeInTheDocument();
      // Should show skeleton
      const skeleton = document.querySelector('[class*="skeleton"]');
      expect(skeleton).toBeInTheDocument();
    });
  });

  describe('Success State', () => {
    it('should display risk score data when loaded', async () => {
      server.use(
        http.get('*/api/v1/merchants/:id/risk-score', () => {
          return HttpResponse.json(mockRiskScore);
        })
      );

      render(<RiskScoreCard merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText('65.0')).toBeInTheDocument();
        expect(screen.getByText('Medium Risk')).toBeInTheDocument();
        expect(screen.getByText('85.0%')).toBeInTheDocument();
      });
    });

    it('should display low risk badge for low risk level', async () => {
      const lowRiskScore = { ...mockRiskScore, risk_level: 'low' as const, risk_score: 0.2 };
      server.use(
        http.get('*/api/v1/merchants/:id/risk-score', () => {
          return HttpResponse.json(lowRiskScore);
        })
      );

      render(<RiskScoreCard merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText('Low Risk')).toBeInTheDocument();
      });
    });

    it('should display high risk badge for high risk level', async () => {
      const highRiskScore = { ...mockRiskScore, risk_level: 'high' as const, risk_score: 0.9 };
      server.use(
        http.get('*/api/v1/merchants/:id/risk-score', () => {
          return HttpResponse.json(highRiskScore);
        })
      );

      render(<RiskScoreCard merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText('High Risk')).toBeInTheDocument();
      });
    });

    it('should display risk factors when available', async () => {
      server.use(
        http.get('*/api/v1/merchants/:id/risk-score', () => {
          return HttpResponse.json(mockRiskScore);
        })
      );

      render(<RiskScoreCard merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText('Key Risk Factors')).toBeInTheDocument();
        expect(screen.getByText('Financial Risk')).toBeInTheDocument();
        expect(screen.getByText('Operational Risk')).toBeInTheDocument();
        expect(screen.getByText('Compliance Risk')).toBeInTheDocument();
      });
    });

    it('should display assessment date when available', async () => {
      server.use(
        http.get('*/api/v1/merchants/:id/risk-score', () => {
          return HttpResponse.json(mockRiskScore);
        })
      );

      render(<RiskScoreCard merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText('Assessment Date')).toBeInTheDocument();
      });
    });

    it('should use correct color for risk score', async () => {
      const lowScore = { ...mockRiskScore, risk_score: 0.2 };
      server.use(
        http.get('*/api/v1/merchants/:id/risk-score', () => {
          return HttpResponse.json(lowScore);
        })
      );

      render(<RiskScoreCard merchantId={merchantId} />);

      await waitFor(() => {
        // Use more specific query - the main risk score (large text)
        const scoreElement = screen.getByText('20.0', { selector: 'p.text-3xl' });
        expect(scoreElement).toHaveClass('text-green-600');
      });
    });
  });

  describe('Error State', () => {
    it('should display error message when API call fails', async () => {
      server.use(
        http.get('*/api/v1/merchants/:id/risk-score', () => {
          return HttpResponse.json({ error: 'Not found' }, { status: 404 });
        })
      );

      render(<RiskScoreCard merchantId={merchantId} />);

      await waitFor(() => {
        // Error message may be "API Error 404" or similar
        const errorText = screen.getByText(/error|failed/i, { exact: false });
        expect(errorText).toBeInTheDocument();
        expect(mockToast.error).toHaveBeenCalled();
      });
    });

    it('should show retry button on error', async () => {
      server.use(
        http.get('*/api/v1/merchants/:id/risk-score', () => {
          return HttpResponse.json({ error: 'Not found' }, { status: 404 });
        })
      );

      render(<RiskScoreCard merchantId={merchantId} />);

      await waitFor(() => {
        const retryButton = screen.getByRole('button', { name: /retry/i });
        expect(retryButton).toBeInTheDocument();
      });
    });

    it('should retry fetching when retry button is clicked', async () => {
      let callCount = 0;
      server.use(
        http.get('*/api/v1/merchants/:id/risk-score', () => {
          callCount++;
          if (callCount === 1) {
            return HttpResponse.json({ error: 'Not found' }, { status: 404 });
          }
          return HttpResponse.json(mockRiskScore);
        })
      );

      const user = userEvent.setup();
      render(<RiskScoreCard merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /retry/i })).toBeInTheDocument();
      });

      const retryButton = screen.getByRole('button', { name: /retry/i });
      await user.click(retryButton);

      await waitFor(() => {
        // Use more specific query - the main risk score (large text)
        const scoreElement = screen.getByText('65.0', { selector: 'p.text-3xl' });
        expect(scoreElement).toBeInTheDocument();
      });
    });
  });

  describe('Edge Cases', () => {
    it('should handle missing risk factors', async () => {
      const scoreWithoutFactors = { ...mockRiskScore, factors: [] };
      server.use(
        http.get('*/api/v1/merchants/:id/risk-score', () => {
          return HttpResponse.json(scoreWithoutFactors);
        })
      );

      render(<RiskScoreCard merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText('65.0')).toBeInTheDocument();
        expect(screen.queryByText('Key Risk Factors')).not.toBeInTheDocument();
      });
    });

    it('should handle missing assessment date', async () => {
      const scoreWithoutDate = { ...mockRiskScore, assessment_date: undefined };
      server.use(
        http.get('*/api/v1/merchants/:id/risk-score', () => {
          return HttpResponse.json(scoreWithoutDate);
        })
      );

      render(<RiskScoreCard merchantId={merchantId} />);

      await waitFor(() => {
        // Use more specific query - the main risk score (large text)
        const scoreElement = screen.getByText('65.0', { selector: 'p.text-3xl' });
        expect(scoreElement).toBeInTheDocument();
        expect(screen.queryByText('Assessment Date')).not.toBeInTheDocument();
      });
    });
  });
});

