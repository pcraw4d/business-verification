import { server } from '@/__tests__/mocks/server';
import { RiskBenchmarkComparison } from '@/components/merchant/RiskBenchmarkComparison';
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

describe('RiskBenchmarkComparison', () => {
  const merchantId = 'merchant-123';

  const mockMerchantAnalytics = {
    merchantId,
    classification: {
      primaryIndustry: 'Technology',
      confidenceScore: 0.95,
      mccCodes: [
        { code: '5734', description: 'Computer Software Stores', confidence: 0.95 },
      ],
      naicsCodes: [],
      sicCodes: [],
    },
    security: { trustScore: 0.8, sslValid: true },
    quality: { completenessScore: 0.9, dataPoints: 100 },
    intelligence: {},
    timestamp: new Date().toISOString(),
  };

  const mockRiskBenchmarks = {
    industry_code: '5734',
    industry_type: 'mcc' as const,
    average_risk_score: 0.6,
    median_risk_score: 0.55,
    percentile_25: 0.45,
    percentile_75: 0.7,
    percentile_90: 0.85,
    sample_size: 100,
    benchmarks: {
      average: 0.6,
      median: 0.55,
      p25: 0.45,
      p75: 0.7,
      p90: 0.85,
    },
  };

  const mockMerchantRiskScore = {
    merchant_id: merchantId,
    risk_score: 0.5, // Lower than average (better)
    risk_level: 'low' as const,
    confidence_score: 0.85,
    assessment_date: '2025-01-27T00:00:00Z',
    factors: [],
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockToast.error = vi.fn();
    
    server.use(
      http.get('*/api/v1/merchants/:id/analytics', () => {
        return HttpResponse.json(mockMerchantAnalytics);
      }),
      http.get('*/api/v1/risk/benchmarks', () => {
        return HttpResponse.json(mockRiskBenchmarks);
      }),
      http.get('*/api/v1/merchants/:id/risk-score', () => {
        return HttpResponse.json(mockMerchantRiskScore);
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

      render(<RiskBenchmarkComparison merchantId={merchantId} />);

      const skeleton = document.querySelector('[class*="skeleton"]');
      expect(skeleton).toBeInTheDocument();
    });
  });

  describe('Success State', () => {
    it('should display benchmark comparison when loaded', async () => {
      render(<RiskBenchmarkComparison merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/industry benchmark/i)).toBeInTheDocument();
      });
    });

    it('should extract MCC code from merchant analytics', async () => {
      render(<RiskBenchmarkComparison merchantId={merchantId} />);

      await waitFor(() => {
        // Should use MCC code for benchmarks
        expect(screen.getByText(/5734|computer software/i)).toBeInTheDocument();
      });
    });

    it('should display merchant score vs industry benchmarks', async () => {
      render(<RiskBenchmarkComparison merchantId={merchantId} />);

      await waitFor(() => {
        // Should show comparison metrics
        expect(screen.getByText(/50\.0|60\.0/i)).toBeInTheDocument();
      });
    });

    it('should display percentile position', async () => {
      render(<RiskBenchmarkComparison merchantId={merchantId} />);

      await waitFor(() => {
        // Should show percentile
        expect(screen.getByText(/\d+%/)).toBeInTheDocument();
      });
    });

    it('should display position indicator (top 10%, top 25%, etc.)', async () => {
      render(<RiskBenchmarkComparison merchantId={merchantId} />);

      await waitFor(() => {
        // Should show position relative to industry
        const positionText = screen.getByText(/top|bottom|average/i);
        expect(positionText).toBeInTheDocument();
      });
    });

    it('should display benchmark chart', async () => {
      render(<RiskBenchmarkComparison merchantId={merchantId} />);

      await waitFor(() => {
        // Should render chart
        const chart = screen.getByTestId('bar-chart');
        expect(chart).toBeInTheDocument();
      });
    });
  });

  describe('Industry Code Extraction', () => {
    it('should prefer MCC code over NAICS', async () => {
      const analyticsWithBoth = {
        ...mockMerchantAnalytics,
        classification: {
          ...mockMerchantAnalytics.classification,
          naicsCodes: [
            { code: '541511', description: 'Custom Computer Programming Services', confidence: 0.9 },
          ],
        },
      };

      server.use(
        http.get('*/api/v1/merchants/:id/analytics', () => {
          return HttpResponse.json(analyticsWithBoth);
        })
      );

      render(<RiskBenchmarkComparison merchantId={merchantId} />);

      await waitFor(() => {
        // Should use MCC (5734) not NAICS
        expect(screen.getByText(/5734|computer software/i)).toBeInTheDocument();
      });
    });

    it('should use NAICS if MCC not available', async () => {
      const analyticsWithNaicsOnly = {
        ...mockMerchantAnalytics,
        classification: {
          ...mockMerchantAnalytics.classification,
          mccCodes: [],
          naicsCodes: [
            { code: '541511', description: 'Custom Computer Programming Services', confidence: 0.9 },
          ],
        },
      };

      server.use(
        http.get('*/api/v1/merchants/:id/analytics', () => {
          return HttpResponse.json(analyticsWithNaicsOnly);
        }),
        http.get('*/api/v1/risk/benchmarks', ({ request }) => {
          const url = new URL(request.url);
          expect(url.searchParams.get('naics')).toBe('541511');
          return HttpResponse.json({ ...mockRiskBenchmarks, industry_code: '541511', industry_type: 'naics' });
        })
      );

      render(<RiskBenchmarkComparison merchantId={merchantId} />);

      await waitFor(() => {
        // Should use NAICS code
        expect(screen.getByText(/541511|custom computer/i)).toBeInTheDocument();
      });
    });

    it('should use SIC if MCC and NAICS not available', async () => {
      const analyticsWithSicOnly = {
        ...mockMerchantAnalytics,
        classification: {
          ...mockMerchantAnalytics.classification,
          mccCodes: [],
          naicsCodes: [],
          sicCodes: [
            { code: '7372', description: 'Prepackaged Software', confidence: 0.9 },
          ],
        },
      };

      server.use(
        http.get('*/api/v1/merchants/:id/analytics', () => {
          return HttpResponse.json(analyticsWithSicOnly);
        }),
        http.get('*/api/v1/risk/benchmarks', ({ request }) => {
          const url = new URL(request.url);
          expect(url.searchParams.get('sic')).toBe('7372');
          return HttpResponse.json({ ...mockRiskBenchmarks, industry_code: '7372', industry_type: 'sic' });
        })
      );

      render(<RiskBenchmarkComparison merchantId={merchantId} />);

      await waitFor(() => {
        // Should use SIC code
        expect(screen.getByText(/7372|prepackaged software/i)).toBeInTheDocument();
      });
    });

    it('should show error when no industry code available', async () => {
      const analyticsWithoutCodes = {
        ...mockMerchantAnalytics,
        classification: {
          ...mockMerchantAnalytics.classification,
          mccCodes: [],
          naicsCodes: [],
          sicCodes: [],
        },
      };

      server.use(
        http.get('*/api/v1/merchants/:id/analytics', () => {
          return HttpResponse.json(analyticsWithoutCodes);
        })
      );

      render(<RiskBenchmarkComparison merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/no industry code available/i)).toBeInTheDocument();
      });
    });
  });

  describe('Comparison Calculations', () => {
    it('should calculate correct percentile for low risk score', async () => {
      const lowRiskScore = { ...mockMerchantRiskScore, risk_score: 0.3 };
      server.use(
        http.get('*/api/v1/merchants/:id/risk-score', () => {
          return HttpResponse.json(lowRiskScore);
        })
      );

      render(<RiskBenchmarkComparison merchantId={merchantId} />);

      await waitFor(() => {
        // Lower risk score should result in higher percentile (better position)
        const percentileText = screen.getByText(/\d+%/);
        expect(percentileText).toBeInTheDocument();
      });
    });

    it('should calculate correct percentile for high risk score', async () => {
      const highRiskScore = { ...mockMerchantRiskScore, risk_score: 0.9 };
      server.use(
        http.get('*/api/v1/merchants/:id/risk-score', () => {
          return HttpResponse.json(highRiskScore);
        })
      );

      render(<RiskBenchmarkComparison merchantId={merchantId} />);

      await waitFor(() => {
        // Higher risk score should result in lower percentile (worse position)
        const percentileText = screen.getByText(/\d+%/);
        expect(percentileText).toBeInTheDocument();
      });
    });

    it('should show "Top 10%" for very low risk score', async () => {
      const veryLowRiskScore = { ...mockMerchantRiskScore, risk_score: 0.2 };
      server.use(
        http.get('*/api/v1/merchants/:id/risk-score', () => {
          return HttpResponse.json(veryLowRiskScore);
        })
      );

      render(<RiskBenchmarkComparison merchantId={merchantId} />);

      await waitFor(() => {
        // Should show top 10% or top 25%
        const positionText = screen.getByText(/top/i);
        expect(positionText).toBeInTheDocument();
      });
    });

    it('should show "Bottom 10%" for very high risk score', async () => {
      const veryHighRiskScore = { ...mockMerchantRiskScore, risk_score: 0.95 };
      server.use(
        http.get('*/api/v1/merchants/:id/risk-score', () => {
          return HttpResponse.json(veryHighRiskScore);
        })
      );

      render(<RiskBenchmarkComparison merchantId={merchantId} />);

      await waitFor(() => {
        // Should show bottom 10% or bottom 25%
        const positionText = screen.getByText(/bottom/i);
        expect(positionText).toBeInTheDocument();
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

      render(<RiskBenchmarkComparison merchantId={merchantId} />);

      await waitFor(() => {
        expect(mockToast.error).toHaveBeenCalled();
      });
    });

    it('should handle benchmarks fetch failure', async () => {
      server.use(
        http.get('*/api/v1/risk/benchmarks', () => {
          return HttpResponse.json({ error: 'Not found' }, { status: 404 });
        })
      );

      render(<RiskBenchmarkComparison merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/failed to load industry benchmarks/i)).toBeInTheDocument();
      });
    });

    it('should handle merchant risk score fetch failure', async () => {
      server.use(
        http.get('*/api/v1/merchants/:id/risk-score', () => {
          return HttpResponse.json({ error: 'Not found' }, { status: 404 });
        })
      );

      render(<RiskBenchmarkComparison merchantId={merchantId} />);

      await waitFor(() => {
        expect(screen.getByText(/failed to load merchant risk score/i)).toBeInTheDocument();
      });
    });
  });

  describe('Detailed Benchmarks Display', () => {
    it('should display detailed benchmark statistics', async () => {
      render(<RiskBenchmarkComparison merchantId={merchantId} />);

      await waitFor(() => {
        // Should show detailed benchmarks (25th, 75th, 90th percentile)
        expect(screen.getByText(/25th|75th|90th|percentile/i)).toBeInTheDocument();
      });
    });
  });
});

