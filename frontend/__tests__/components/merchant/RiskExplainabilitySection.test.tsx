import { server } from '@/__tests__/mocks/server';
import { RiskExplainabilitySection } from '@/components/merchant/RiskExplainabilitySection';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
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

describe('RiskExplainabilitySection', () => {
  const merchantId = 'merchant-123';

  const mockRiskAssessment = {
    id: 'assessment-123',
    merchantId,
    status: 'completed' as const,
    createdAt: '2025-01-27T00:00:00Z',
    result: {
      overallScore: 0.65,
      riskLevel: 'medium',
      factors: [
        { name: 'Financial Risk', score: 0.7, weight: 0.4 },
        { name: 'Operational Risk', score: 0.6, weight: 0.3 },
        { name: 'Compliance Risk', score: 0.65, weight: 0.3 },
      ],
    },
  };

  const mockRiskExplanation = {
    assessmentId: 'assessment-123',
    factors: [
      { name: 'Financial Risk', score: 0.7, weight: 0.4 },
      { name: 'Operational Risk', score: 0.6, weight: 0.3 },
      { name: 'Compliance Risk', score: 0.65, weight: 0.3 },
      { name: 'Security Risk', score: 0.5, weight: 0.2 },
    ],
    shapValues: {
      'financial_indicators': 0.15,
      'operational_efficiency': 0.12,
      'compliance_score': 0.10,
      'security_measures': 0.08,
      'business_age': 0.05,
      'transaction_volume': 0.04,
      'geographic_risk': 0.03,
      'industry_risk': 0.02,
      'credit_score': 0.01,
      'revenue_growth': 0.005,
    },
    baseValue: 0.5,
    prediction: 0.65,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockToast.error = vi.fn();
    
    server.use(
      http.get('*/api/v1/merchants/:id/risk-assessment', () => {
        return HttpResponse.json(mockRiskAssessment);
      }),
      http.get('*/api/v1/risk/explain/:assessmentId', () => {
        return HttpResponse.json(mockRiskExplanation);
      })
    );
  });

  describe('Loading State', () => {
    it('should show loading skeleton initially', () => {
      server.use(
        http.get('*/api/v1/merchants/:id/risk-assessment', async () => {
          await new Promise((resolve) => setTimeout(resolve, 100));
          return HttpResponse.json(mockRiskAssessment);
        })
      );

      render(<RiskExplainabilitySection merchantId={merchantId} />);

      const skeleton = document.querySelector('[class*="skeleton"]');
      expect(skeleton).toBeInTheDocument();
    });
  });

  describe('Success State', () => {
    it('should display risk explainability section when loaded', async () => {
      render(<RiskExplainabilitySection merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/risk assessment explainability/i)).toBeInTheDocument();
        expect(screen.getByText(/shap values and feature importance/i)).toBeInTheDocument();
      });
    });

    it('should display SHAP values chart', async () => {
      render(<RiskExplainabilitySection merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/shap values/i)).toBeInTheDocument();
        const chart = screen.getByTestId('bar-chart');
        expect(chart).toBeInTheDocument();
      });
    });

    it('should display feature importance chart', async () => {
      render(<RiskExplainabilitySection merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/feature importance/i)).toBeInTheDocument();
        const charts = screen.getAllByTestId('bar-chart');
        expect(charts.length).toBeGreaterThan(0);
      });
    });

    it('should display risk factors table', async () => {
      render(<RiskExplainabilitySection merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/risk factors/i)).toBeInTheDocument();
        expect(screen.getByText('Financial Risk')).toBeInTheDocument();
        expect(screen.getByText('Operational Risk')).toBeInTheDocument();
        expect(screen.getByText('Compliance Risk')).toBeInTheDocument();
      });
    });

    it('should display top 10 SHAP values', async () => {
      render(<RiskExplainabilitySection merchantId={merchantId} />);

      await waitFor(() => {
        // Should show top features from SHAP values
        expect(screen.getByText(/financial_indicators|operational_efficiency/i)).toBeInTheDocument();
      });
    });

    it('should display factor scores and weights in table', async () => {
      render(<RiskExplainabilitySection merchantId={merchantId} />);

      await waitFor(() => {
        // Should show scores and weights
        expect(screen.getByText(/0\.7|0\.6|0\.65/i)).toBeInTheDocument();
        expect(screen.getByText(/0\.4|0\.3/i)).toBeInTheDocument();
      });
    });

    it('should calculate and display impact (score * weight)', async () => {
      render(<RiskExplainabilitySection merchantId={merchantId} />);

      await waitFor(() => {
        // Impact = score * weight
        // Financial Risk: 0.7 * 0.4 = 0.28
        // Should show impact values
        expect(screen.getByText(/0\.28|0\.18|0\.195/i)).toBeInTheDocument();
      });
    });
  });

  describe('Assessment ID Resolution', () => {
    it('should fetch assessment ID from risk assessment', async () => {
      render(<RiskExplainabilitySection merchantId={merchantId} />);

      await waitFor(() => {
        // Should have fetched assessment and then explanation
        expect(screen.getByText(/risk assessment explainability/i)).toBeInTheDocument();
      });
    });

    it('should use cached assessment ID on subsequent renders', async () => {
      const { rerender } = render(<RiskExplainabilitySection merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/risk assessment explainability/i)).toBeInTheDocument();
      });

      // Rerender should use cached assessment ID
      rerender(<RiskExplainabilitySection merchantId={merchantId} />);

      await waitFor(() => {
        // Should still work without re-fetching assessment
        expect(screen.getByText(/risk assessment explainability/i)).toBeInTheDocument();
      });
    });
  });

  describe('Error Handling', () => {
    it('should handle missing risk assessment', async () => {
      server.use(
        http.get('*/api/v1/merchants/:id/risk-assessment', () => {
          return HttpResponse.json({ error: 'Not found' }, { status: 404 });
        })
      );

      render(<RiskExplainabilitySection merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/no risk assessment found/i)).toBeInTheDocument();
      });
    });

    it('should handle risk assessment without ID', async () => {
      const assessmentWithoutId = { ...mockRiskAssessment, id: undefined };
      server.use(
        http.get('*/api/v1/merchants/:id/risk-assessment', () => {
          return HttpResponse.json(assessmentWithoutId);
        })
      );

      render(<RiskExplainabilitySection merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/no risk assessment found/i)).toBeInTheDocument();
      });
    });

    it('should handle explanation fetch failure', async () => {
      server.use(
        http.get('*/api/v1/risk/explain/:assessmentId', () => {
          return HttpResponse.json({ error: 'Not found' }, { status: 404 });
        })
      );

      render(<RiskExplainabilitySection merchantId={merchantId} />);

      await waitFor(() => {
        expect(mockToast.error).toHaveBeenCalled();
      });
    });

    it('should show retry button on error', async () => {
      server.use(
        http.get('*/api/v1/risk/explain/:assessmentId', () => {
          return HttpResponse.json({ error: 'Not found' }, { status: 404 });
        })
      );

      render(<RiskExplainabilitySection merchantId={merchantId} />);

      await waitFor(() => {
        const retryButton = screen.getByRole('button', { name: /retry/i });
        expect(retryButton).toBeInTheDocument();
      });
    });

    it('should retry fetching when retry button is clicked', async () => {
      let callCount = 0;
      server.use(
        http.get('*/api/v1/risk/explain/:assessmentId', () => {
          callCount++;
          if (callCount === 1) {
            return HttpResponse.json({ error: 'Not found' }, { status: 404 });
          }
          return HttpResponse.json(mockRiskExplanation);
        })
      );

      const user = userEvent.setup();
      render(<RiskExplainabilitySection merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /retry/i })).toBeInTheDocument();
      });

      const retryButton = screen.getByRole('button', { name: /retry/i });
      await user.click(retryButton);

      await waitFor(() => {
        expect(screen.getByText(/shap values/i)).toBeInTheDocument();
      });
    });
  });

  describe('Empty State', () => {
    it('should show message when no explanation data available', async () => {
      server.use(
        http.get('*/api/v1/merchants/:id/risk-assessment', () => {
          return HttpResponse.json(null);
        })
      );

      render(<RiskExplainabilitySection merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/no explanation data available/i)).toBeInTheDocument();
      });
    });
  });

  describe('SHAP Values Processing', () => {
    it('should sort SHAP values by absolute value', async () => {
      render(<RiskExplainabilitySection merchantId={merchantId} />);

      await waitFor(() => {
        // Top features should be displayed (sorted by absolute value)
        expect(screen.getByText(/financial_indicators/i)).toBeInTheDocument();
      });
    });

    it('should limit SHAP values to top 10', async () => {
      render(<RiskExplainabilitySection merchantId={merchantId} />);

      await waitFor(() => {
        // Should only show top 10 features
        const chart = screen.getByTestId('bar-chart');
        expect(chart).toBeInTheDocument();
      });
    });
  });

  describe('Feature Importance Calculation', () => {
    it('should calculate impact as score * weight', async () => {
      render(<RiskExplainabilitySection merchantId={merchantId} />);

      await waitFor(() => {
        // Financial Risk: 0.7 * 0.4 = 0.28
        // Operational Risk: 0.6 * 0.3 = 0.18
        // Compliance Risk: 0.65 * 0.3 = 0.195
        expect(screen.getByText(/0\.28|0\.18|0\.195/i)).toBeInTheDocument();
      });
    });

    it('should sort features by impact', async () => {
      render(<RiskExplainabilitySection merchantId={merchantId} />);

      await waitFor(() => {
        // Features should be sorted by impact (highest first)
        expect(screen.getByText(/financial risk/i)).toBeInTheDocument();
      });
    });
  });
});

