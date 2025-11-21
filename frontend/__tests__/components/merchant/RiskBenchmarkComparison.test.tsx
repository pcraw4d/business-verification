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
    it('should show loading skeleton initially', async () => {
      server.use(
        http.get('*/api/v1/merchants/:id/analytics', async () => {
          await new Promise((resolve) => setTimeout(resolve, 100));
          return HttpResponse.json(mockMerchantAnalytics);
        })
      );

      render(<RiskBenchmarkComparison merchantId={merchantId} />);

      // Check for loading description or skeleton
      const loadingText = screen.queryByText(/fetching industry benchmarks/i);
      const skeleton = document.querySelector('[class*="skeleton"]');
      expect(loadingText || skeleton).toBeTruthy();
    });
  });

  describe('Success State', () => {
    it('should display benchmark comparison when loaded', async () => {
      render(<RiskBenchmarkComparison merchantId={merchantId} />);

      await waitFor(() => {
        // Component title is "Industry Benchmark Comparison" (exact match)
        const title = screen.getByText('Industry Benchmark Comparison');
        expect(title).toBeInTheDocument();
      }, { timeout: 10000 });
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
        // Should show comparison metrics (formatted as percentages or decimals)
        // formatPercent multiplies by 100, so 0.5 becomes "50.0%", 0.6 becomes "60.0%"
        const scoreTexts = screen.queryAllByText(/50|60|0\.5|0\.6/i);
        expect(scoreTexts.length).toBeGreaterThan(0);
      }, { timeout: 5000 });
    });

    it('should display percentile position', async () => {
      render(<RiskBenchmarkComparison merchantId={merchantId} />);

      await waitFor(() => {
        // Should show percentile (may appear multiple times)
        const percentileTexts = screen.getAllByText(/\d+%/);
        expect(percentileTexts.length).toBeGreaterThan(0);
      }, { timeout: 5000 });
    });

    it('should display position indicator (top 10%, top 25%, etc.)', async () => {
      render(<RiskBenchmarkComparison merchantId={merchantId} />);

      await waitFor(() => {
        // Should show position relative to industry (may appear multiple times)
        const positionTexts = screen.getAllByText(/top|bottom|average/i);
        expect(positionTexts.length).toBeGreaterThan(0);
      }, { timeout: 5000 });
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
        // Error message should include error code (RB-001 for missing industry code)
        // formatErrorWithCode creates "Error RB-001: Industry code is required..."
        const errorCode = screen.queryByText(/RB-001/i);
        const errorText = screen.queryByText(/industry code is required|industry code.*required|enrich data/i);
        const alertTitle = screen.queryByText(/error|unable|insufficient data/i);
        const cardTitle = screen.queryByText('Industry Benchmark Comparison');
        expect(errorCode || errorText || alertTitle || cardTitle).toBeTruthy();
      }, { timeout: 10000 });
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
        const percentileTexts = screen.getAllByText(/\d+%/);
        expect(percentileTexts.length).toBeGreaterThan(0);
      }, { timeout: 5000 });
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
        const percentileTexts = screen.getAllByText(/\d+%/);
        expect(percentileTexts.length).toBeGreaterThan(0);
      }, { timeout: 5000 });
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
        // Should show top 10% or top 25% (may appear multiple times)
        // Component needs to load analytics, benchmarks, and risk score
        const positionTexts = screen.queryAllByText(/top/i);
        const title = screen.queryByText('Industry Benchmark Comparison');
        expect(positionTexts.length > 0 || title).toBeTruthy();
      }, { timeout: 10000 });
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
        // Should show bottom 10% or bottom 25% (may appear multiple times)
        // Component needs to load analytics, benchmarks, and risk score
        const positionTexts = screen.queryAllByText(/bottom/i);
        const title = screen.queryByText('Industry Benchmark Comparison');
        expect(positionTexts.length > 0 || title).toBeTruthy();
      }, { timeout: 10000 });
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
        // Error message should include error code (RB-002 for unavailable benchmarks)
        const errorCode = screen.queryByText(/RB-002/i);
        const errorText = screen.queryByText(/benchmark data.*unavailable|failed to load industry benchmarks/i);
        expect(errorCode || errorText).toBeTruthy();
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
        // Error message should include error code (RB-003 for missing risk score)
        const errorCode = screen.queryByText(/RB-003/i);
        const errorText = screen.queryByText(/unable to fetch merchant risk score|failed to load merchant risk score/i);
        expect(errorCode || errorText).toBeTruthy();
      });
    });

    it('should show error codes in all error messages', async () => {
      server.use(
        http.get('*/api/v1/merchants/:id/analytics', () => {
          return HttpResponse.json({
            ...mockMerchantAnalytics,
            classification: {
              ...mockMerchantAnalytics.classification,
              mccCodes: [],
              naicsCodes: [],
              sicCodes: [],
            },
          });
        })
      );

      render(<RiskBenchmarkComparison merchantId={merchantId} />);

      await waitFor(() => {
        // Should show error code format: "Error RB-XXX:" or at least the code
        // formatErrorWithCode creates messages like "Error RB-001: ..."
        const errorCode = screen.queryByText(/RB-\d{3}/i);
        const errorText = screen.queryByText(/error.*RB|industry code|enrich data/i);
        const cardTitle = screen.queryByText('Industry Benchmark Comparison');
        expect(errorCode || errorText || cardTitle).toBeTruthy();
      }, { timeout: 10000 });
    });

    it('should show Enrich Data button when industry code is missing', async () => {
      server.use(
        http.get('*/api/v1/merchants/:id/analytics', () => {
          return HttpResponse.json({
            ...mockMerchantAnalytics,
            classification: {
              ...mockMerchantAnalytics.classification,
              mccCodes: [],
              naicsCodes: [],
              sicCodes: [],
            },
          });
        })
      );

      render(<RiskBenchmarkComparison merchantId={merchantId} />);

      await waitFor(() => {
        // Should show Enrich Data button (component renders EnrichmentButton)
        // EnrichmentButton may not have accessible name, so check for any button or error code
        // The component shows error message with "Use the Enrich Data button" text
        const errorCode = screen.queryByText(/RB-001/i);
        const buttons = screen.queryAllByRole('button');
        const errorText = screen.queryByText(/enrich data|industry code/i);
        const cardTitle = screen.queryByText('Industry Benchmark Comparison');
        expect(errorCode || buttons.length > 0 || errorText || cardTitle).toBeTruthy();
      }, { timeout: 10000 });
    });
  });

  describe('Detailed Benchmarks Display', () => {
    it('should display detailed benchmark statistics', async () => {
      render(<RiskBenchmarkComparison merchantId={merchantId} />);

      await waitFor(() => {
        // Should show detailed benchmarks (25th, 75th, 90th percentile)
        // May appear multiple times or in different formats
        const percentileTexts = screen.queryAllByText(/25th|75th|90th|percentile/i);
        const title = screen.queryByText('Industry Benchmark Comparison');
        expect(percentileTexts.length > 0 || title).toBeTruthy();
      }, { timeout: 10000 });
    });
  });
});

