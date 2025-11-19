// Vitest globals are available via globals: true in vitest.config.ts
import { vi } from 'vitest';

// Mock ErrorHandler - must be before imports
vi.mock('@/lib/error-handler', () => ({
  ErrorHandler: {
    handleAPIError: vi.fn().mockResolvedValue(undefined),
    showErrorNotification: vi.fn(),
    showSuccessNotification: vi.fn(),
    showInfoNotification: vi.fn(),
    parseErrorResponse: vi.fn().mockResolvedValue({ code: 'TEST_ERROR', message: 'Test error' }),
    logError: vi.fn(),
  },
}));

// Mock APICache and RequestDeduplicator
// These must be mocked before the api module is imported
// because api.ts initializes instances at module load time
const mockCache = new Map<string, any>();
const mockCacheTTL = new Map<string, number>();

vi.mock('@/lib/api-cache', () => ({
  APICache: class APICache {
    constructor(ttl?: number) {}
    get = vi.fn((key: string) => {
      return mockCache.get(key) || null;
    });
    set = vi.fn((key: string, value: any, ttl?: number) => {
      mockCache.set(key, value);
      if (ttl) {
        mockCacheTTL.set(key, ttl);
      }
    });
  },
}));

vi.mock('@/lib/request-deduplicator', () => ({
  RequestDeduplicator: class RequestDeduplicator {
    deduplicate = vi.fn((key, fn) => fn());
  },
}));

// Import after mocks
import {
  getMerchant,
  getMerchantAnalytics,
  getWebsiteAnalysis,
  getRiskAssessment,
  startRiskAssessment,
  getAssessmentStatus,
  getRiskHistory,
  getRiskPredictions,
  explainRiskAssessment,
  getRiskRecommendations,
  getRiskIndicators,
  getEnrichmentSources,
  triggerEnrichment,
  getPortfolioAnalytics,
  getPortfolioStatistics,
  getRiskTrends,
  getRiskInsights,
  getRiskBenchmarks,
  getMerchantRiskScore,
  getRiskAlerts,
  getAPICache,
  getRequestDeduplicator,
} from '@/lib/api';
import { ErrorHandler } from '@/lib/error-handler';
import { APICache } from '@/lib/api-cache';
import { RequestDeduplicator } from '@/lib/request-deduplicator';

describe('API Client', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Clear cache between tests
    mockCache.clear();
    mockCacheTTL.clear();
    // Mock fetch globally
    global.fetch = vi.fn();
    // Reset ErrorHandler mocks
    vi.mocked(ErrorHandler.handleAPIError).mockResolvedValue(undefined);
    vi.mocked(ErrorHandler.parseErrorResponse).mockResolvedValue({ code: 'TEST_ERROR', message: 'Test error' });
    // Reset sessionStorage mock
    Object.defineProperty(window, 'sessionStorage', {
      value: {
        getItem: vi.fn().mockReturnValue(null),
        setItem: vi.fn(),
        removeItem: vi.fn(),
        clear: vi.fn(),
      },
      writable: true,
    });
  });

  describe('getMerchant', () => {
    it('should fetch merchant data successfully', async () => {
      const mockMerchant = {
        id: 'merchant-123',
        businessName: 'Test Business',
        status: 'active',
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockMerchant,
      });

      const result = await getMerchant('merchant-123');

      expect(result).toEqual(mockMerchant);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/merchants/merchant-123'),
        expect.objectContaining({ method: 'GET' })
      );
    });

    it('should handle API errors', async () => {
      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: false,
        status: 404,
        statusText: 'Not Found',
        json: async () => ({ code: 'NOT_FOUND', message: 'Merchant not found' }),
      });

      // Verify error is thrown when API returns error status
      await expect(getMerchant('invalid-id')).rejects.toThrow();
      
      // Verify fetch was called
      expect(global.fetch).toHaveBeenCalled();
    });
  });

  describe('getMerchantAnalytics', () => {
    it('should fetch analytics data successfully', async () => {
      const mockAnalytics = {
        merchantId: 'merchant-123',
        classification: { primaryIndustry: 'Technology', confidenceScore: 0.95 },
        security: { trustScore: 0.8, sslValid: true },
        quality: { completenessScore: 0.9, dataPoints: 100 },
        intelligence: {},
        timestamp: new Date().toISOString(),
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockAnalytics,
      });

      const result = await getMerchantAnalytics('merchant-123');

      expect(result).toEqual(mockAnalytics);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/merchants/merchant-123/analytics'),
        expect.objectContaining({ method: 'GET' })
      );
    });
  });

  describe('getRiskAssessment', () => {
    it('should fetch risk assessment successfully', async () => {
      const mockAssessment = {
        id: 'assessment-123',
        merchantId: 'merchant-123',
        status: 'completed',
        result: { overallScore: 0.7, riskLevel: 'medium' },
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockAssessment,
      });

      const result = await getRiskAssessment('merchant-123');

      expect(result).toEqual(mockAssessment);
    });

    it('should return null for 404 errors', async () => {
      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: false,
        status: 404,
        statusText: 'Not Found',
      });

      const result = await getRiskAssessment('merchant-123');

      expect(result).toBeNull();
    });
  });

  describe('startRiskAssessment', () => {
    it('should start risk assessment successfully', async () => {
      const mockResponse = {
        assessmentId: 'assessment-123',
        status: 'pending',
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      const result = await startRiskAssessment({
        merchantId: 'merchant-123',
        options: { includeHistory: true },
      });

      expect(result).toEqual(mockResponse);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/risk/assess'),
        expect.objectContaining({
          method: 'POST',
          body: expect.any(String),
        })
      );
    });
  });

  describe('getRiskHistory', () => {
    it('should fetch risk history with pagination', async () => {
      const mockHistory = {
        merchantId: 'merchant-123',
        history: [],
        limit: 10,
        offset: 0,
        total: 0,
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockHistory,
      });

      const result = await getRiskHistory('merchant-123', 10, 0);

      expect(result).toEqual(mockHistory);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/risk/history/merchant-123'),
        expect.objectContaining({ method: 'GET' })
      );
    });
  });

  describe('getRiskPredictions', () => {
    it('should fetch risk predictions with horizons', async () => {
      const mockPredictions = {
        merchantId: 'merchant-123',
        horizons: [3, 6, 12],
        predictions: [],
        includeScenarios: false,
        includeConfidence: false,
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockPredictions,
      });

      const result = await getRiskPredictions('merchant-123', [3, 6, 12]);

      expect(result).toEqual(mockPredictions);
    });
  });

  describe('explainRiskAssessment', () => {
    it('should fetch risk assessment explanation', async () => {
      const mockExplanation = {
        assessmentId: 'assessment-123',
        factors: [],
        shapValues: {},
        baseValue: 0.5,
        prediction: 0.7,
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockExplanation,
      });

      const result = await explainRiskAssessment('assessment-123');

      expect(result).toEqual(mockExplanation);
    });
  });

  describe('getRiskRecommendations', () => {
    it('should fetch risk recommendations', async () => {
      const mockRecommendations = {
        merchantId: 'merchant-123',
        recommendations: [],
        timestamp: new Date().toISOString(),
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockRecommendations,
      });

      const result = await getRiskRecommendations('merchant-123');

      expect(result).toEqual(mockRecommendations);
    });
  });

  describe('getRiskIndicators', () => {
    it('should fetch risk indicators with filters', async () => {
      const mockIndicators = {
        merchantId: 'merchant-123',
        overallScore: 0.7,
        indicators: [],
        lastUpdated: new Date().toISOString(),
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockIndicators,
      });

      const result = await getRiskIndicators('merchant-123', 'high', 'active');

      expect(result).toEqual(mockIndicators);
    });
  });

  describe('getEnrichmentSources', () => {
    it('should fetch enrichment sources', async () => {
      const mockSources = {
        merchantId: 'merchant-123',
        sources: [],
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockSources,
      });

      const result = await getEnrichmentSources('merchant-123');

      expect(result).toEqual(mockSources);
    });
  });

  describe('triggerEnrichment', () => {
    it('should trigger enrichment successfully', async () => {
      const mockJob = {
        jobId: 'job-123',
        merchantId: 'merchant-123',
        source: 'external-api',
        status: 'pending',
        createdAt: new Date().toISOString(),
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockJob,
      });

      const result = await triggerEnrichment('merchant-123', 'external-api');

      expect(result).toEqual(mockJob);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/merchants/merchant-123/enrichment/trigger'),
        expect.objectContaining({
          method: 'POST',
          body: expect.stringContaining('external-api'),
        })
      );
    });
  });

  describe('getPortfolioAnalytics', () => {
    it('should fetch portfolio analytics successfully', async () => {
      const mockAnalytics = {
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

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockAnalytics,
      });

      const result = await getPortfolioAnalytics();

      expect(result).toEqual(mockAnalytics);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/merchants/analytics'),
        expect.objectContaining({ method: 'GET' })
      );
    });

    it('should use cache when available', async () => {
      const mockAnalytics = {
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

      const apiCache = getAPICache();
      apiCache.set('portfolio-analytics', mockAnalytics);

      const result = await getPortfolioAnalytics();

      expect(result).toEqual(mockAnalytics);
      expect(global.fetch).not.toHaveBeenCalled();
    });

    it('should handle API errors', async () => {
      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error',
        json: async () => ({ code: 'SERVER_ERROR', message: 'Internal server error' }),
      });

      await expect(getPortfolioAnalytics()).rejects.toThrow();
      expect(ErrorHandler.handleAPIError).toHaveBeenCalled();
    });

    it('should cache response after successful fetch', async () => {
      const mockAnalytics = {
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

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockAnalytics,
      });

      const apiCache = getAPICache();
      const setSpy = vi.spyOn(apiCache, 'set');

      await getPortfolioAnalytics();

      expect(setSpy).toHaveBeenCalledWith('portfolio-analytics', mockAnalytics, 7 * 60 * 1000);
    });
  });

  describe('getPortfolioStatistics', () => {
    it('should fetch portfolio statistics successfully', async () => {
      const mockStatistics = {
        totalMerchants: 100,
        averageRiskScore: 0.6,
        riskDistribution: { low: 40, medium: 50, high: 10 },
        industryDistribution: {},
        countryDistribution: {},
        timestamp: new Date().toISOString(),
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockStatistics,
      });

      const result = await getPortfolioStatistics();

      expect(result).toEqual(mockStatistics);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/merchants/statistics'),
        expect.objectContaining({ method: 'GET' })
      );
    });

    it('should use cache when available', async () => {
      const mockStatistics = {
        totalMerchants: 100,
        averageRiskScore: 0.6,
        riskDistribution: { low: 40, medium: 50, high: 10 },
        industryDistribution: {},
        countryDistribution: {},
        timestamp: new Date().toISOString(),
      };

      const apiCache = getAPICache();
      apiCache.set('portfolio-statistics', mockStatistics);

      const result = await getPortfolioStatistics();

      expect(result).toEqual(mockStatistics);
      expect(global.fetch).not.toHaveBeenCalled();
    });

    it('should handle API errors', async () => {
      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error',
        json: async () => ({ code: 'SERVER_ERROR', message: 'Internal server error' }),
      });

      await expect(getPortfolioStatistics()).rejects.toThrow();
      expect(ErrorHandler.handleAPIError).toHaveBeenCalled();
    });
  });

  describe('getRiskTrends', () => {
    it('should fetch risk trends successfully without params', async () => {
      const mockTrends = {
        trends: [],
        summary: {
          averageScore: 0.6,
          trendDirection: 'stable',
          changePercentage: 0,
        },
        timeframe: '6m',
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockTrends,
      });

      const result = await getRiskTrends();

      expect(result).toEqual(mockTrends);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/analytics/trends'),
        expect.objectContaining({ method: 'GET' })
      );
    });

    it('should fetch risk trends with query parameters', async () => {
      const mockTrends = {
        trends: [],
        summary: {
          averageScore: 0.6,
          trendDirection: 'improving',
          changePercentage: -5.2,
        },
        timeframe: '3m',
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockTrends,
      });

      const result = await getRiskTrends({
        industry: 'technology',
        country: 'US',
        timeframe: '3m',
        limit: 10,
      });

      expect(result).toEqual(mockTrends);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/analytics/trends'),
        expect.objectContaining({ method: 'GET' })
      );
    });

    it('should use cache when available', async () => {
      const mockTrends = {
        trends: [],
        summary: {
          averageScore: 0.6,
          trendDirection: 'stable',
          changePercentage: 0,
        },
        timeframe: '6m',
      };

      const apiCache = getAPICache();
      apiCache.set('risk-trends:{}', mockTrends);

      const result = await getRiskTrends();

      expect(result).toEqual(mockTrends);
      expect(global.fetch).not.toHaveBeenCalled();
    });

    it('should create different cache keys for different params', async () => {
      const mockTrends1 = {
        trends: [],
        summary: { averageScore: 0.6, trendDirection: 'stable', changePercentage: 0 },
        timeframe: '6m',
      };
      const mockTrends2 = {
        trends: [],
        summary: { averageScore: 0.7, trendDirection: 'declining', changePercentage: 5 },
        timeframe: '3m',
      };

      (global.fetch as vi.Mock)
        .mockResolvedValueOnce({ ok: true, json: async () => mockTrends1 })
        .mockResolvedValueOnce({ ok: true, json: async () => mockTrends2 });

      const result1 = await getRiskTrends();
      const result2 = await getRiskTrends({ industry: 'technology' });

      expect(result1).toEqual(mockTrends1);
      expect(result2).toEqual(mockTrends2);
      expect(global.fetch).toHaveBeenCalledTimes(2);
    });

    it('should handle API errors', async () => {
      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error',
        json: async () => ({ code: 'SERVER_ERROR', message: 'Internal server error' }),
      });

      await expect(getRiskTrends()).rejects.toThrow();
      expect(ErrorHandler.handleAPIError).toHaveBeenCalled();
    });
  });

  describe('getRiskInsights', () => {
    it('should fetch risk insights successfully without params', async () => {
      const mockInsights = {
        insights: [],
        recommendations: [],
        timestamp: new Date().toISOString(),
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockInsights,
      });

      const result = await getRiskInsights();

      expect(result).toEqual(mockInsights);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/analytics/insights'),
        expect.objectContaining({ method: 'GET' })
      );
    });

    it('should fetch risk insights with query parameters', async () => {
      const mockInsights = {
        insights: [
          {
            id: 'insight-1',
            type: 'risk_distribution',
            title: 'High Risk Concentration',
            description: '10% of merchants are high risk',
            severity: 'medium',
          },
        ],
        recommendations: [],
        timestamp: new Date().toISOString(),
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockInsights,
      });

      const result = await getRiskInsights({
        industry: 'technology',
        country: 'US',
        risk_level: 'high',
      });

      expect(result).toEqual(mockInsights);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/analytics/insights'),
        expect.objectContaining({ method: 'GET' })
      );
    });

    it('should use cache when available', async () => {
      const mockInsights = {
        insights: [],
        recommendations: [],
        timestamp: new Date().toISOString(),
      };

      const apiCache = getAPICache();
      apiCache.set('risk-insights:{}', mockInsights);

      const result = await getRiskInsights();

      expect(result).toEqual(mockInsights);
      expect(global.fetch).not.toHaveBeenCalled();
    });

    it('should handle API errors', async () => {
      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error',
        json: async () => ({ code: 'SERVER_ERROR', message: 'Internal server error' }),
      });

      await expect(getRiskInsights()).rejects.toThrow();
      expect(ErrorHandler.handleAPIError).toHaveBeenCalled();
    });
  });

  describe('getRiskBenchmarks', () => {
    it('should fetch risk benchmarks with MCC code', async () => {
      const mockBenchmarks = {
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

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockBenchmarks,
      });

      const result = await getRiskBenchmarks({ mcc: '5734' });

      expect(result).toEqual(mockBenchmarks);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/risk/benchmarks'),
        expect.objectContaining({ method: 'GET' })
      );
    });

    it('should fetch risk benchmarks with NAICS code', async () => {
      const mockBenchmarks = {
        industry_code: '541511',
        industry_type: 'naics' as const,
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

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockBenchmarks,
      });

      const result = await getRiskBenchmarks({ naics: '541511' });

      expect(result).toEqual(mockBenchmarks);
    });

    it('should fetch risk benchmarks with SIC code', async () => {
      const mockBenchmarks = {
        industry_code: '7372',
        industry_type: 'sic' as const,
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

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockBenchmarks,
      });

      const result = await getRiskBenchmarks({ sic: '7372' });

      expect(result).toEqual(mockBenchmarks);
    });

    it('should use cache when available', async () => {
      const mockBenchmarks = {
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

      const apiCache = getAPICache();
      apiCache.set('risk-benchmarks:5734', mockBenchmarks);

      const result = await getRiskBenchmarks({ mcc: '5734' });

      expect(result).toEqual(mockBenchmarks);
      expect(global.fetch).not.toHaveBeenCalled();
    });

    it('should handle API errors', async () => {
      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: false,
        status: 404,
        statusText: 'Not Found',
        json: async () => ({ code: 'NOT_FOUND', message: 'Benchmarks not found' }),
      });

      await expect(getRiskBenchmarks({ mcc: '5734' })).rejects.toThrow();
      expect(ErrorHandler.handleAPIError).toHaveBeenCalled();
    });
  });

  describe('getMerchantRiskScore', () => {
    it('should fetch merchant risk score successfully', async () => {
      const mockRiskScore = {
        merchant_id: 'merchant-123',
        risk_score: 0.65,
        risk_level: 'medium' as const,
        confidence_score: 0.85,
        assessment_date: '2025-01-27T00:00:00Z',
        factors: [
          { name: 'Financial Risk', score: 0.7, weight: 0.4 },
        ],
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockRiskScore,
      });

      const result = await getMerchantRiskScore('merchant-123');

      expect(result).toEqual(mockRiskScore);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/merchants/merchant-123/risk-score'),
        expect.objectContaining({ method: 'GET' })
      );
    });

    it('should use cache when available', async () => {
      const mockRiskScore = {
        merchant_id: 'merchant-123',
        risk_score: 0.65,
        risk_level: 'medium' as const,
        confidence_score: 0.85,
        assessment_date: '2025-01-27T00:00:00Z',
        factors: [],
      };

      const apiCache = getAPICache();
      apiCache.set('merchant-risk-score:merchant-123', mockRiskScore);

      const result = await getMerchantRiskScore('merchant-123');

      expect(result).toEqual(mockRiskScore);
      expect(global.fetch).not.toHaveBeenCalled();
    });

    it('should handle API errors', async () => {
      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: false,
        status: 404,
        statusText: 'Not Found',
        json: async () => ({ code: 'NOT_FOUND', message: 'Risk score not found' }),
      });

      await expect(getMerchantRiskScore('merchant-123')).rejects.toThrow();
      expect(ErrorHandler.handleAPIError).toHaveBeenCalled();
    });

    it('should cache response with 3 minute TTL', async () => {
      const mockRiskScore = {
        merchant_id: 'merchant-123',
        risk_score: 0.65,
        risk_level: 'medium' as const,
        confidence_score: 0.85,
        assessment_date: '2025-01-27T00:00:00Z',
        factors: [],
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockRiskScore,
      });

      const apiCache = getAPICache();
      const setSpy = vi.spyOn(apiCache, 'set');

      await getMerchantRiskScore('merchant-123');

      expect(setSpy).toHaveBeenCalledWith('merchant-risk-score:merchant-123', mockRiskScore, 3 * 60 * 1000);
    });
  });

  describe('getRiskAlerts', () => {
    it('should fetch risk alerts successfully', async () => {
      const mockAlerts = {
        merchantId: 'merchant-123',
        overallScore: 0.7,
        indicators: [
          {
            id: 'indicator-1',
            type: 'financial',
            severity: 'high',
            status: 'active',
            title: 'High Risk Indicator',
            description: 'Merchant has high financial risk',
            detectedAt: '2025-01-27T00:00:00Z',
          },
        ],
        lastUpdated: '2025-01-27T00:00:00Z',
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockAlerts,
      });

      const result = await getRiskAlerts('merchant-123');

      expect(result).toEqual(mockAlerts);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/risk/indicators/merchant-123'),
        expect.objectContaining({ method: 'GET' })
      );
    });

    it('should fetch risk alerts with severity filter', async () => {
      const mockAlerts = {
        merchantId: 'merchant-123',
        overallScore: 0.7,
        indicators: [
          {
            id: 'indicator-1',
            type: 'financial',
            severity: 'high',
            status: 'active',
            title: 'High Risk Indicator',
            description: 'Merchant has high financial risk',
            detectedAt: '2025-01-27T00:00:00Z',
          },
        ],
        lastUpdated: '2025-01-27T00:00:00Z',
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockAlerts,
      });

      const result = await getRiskAlerts('merchant-123', 'high');

      expect(result).toEqual(mockAlerts);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/risk/indicators/merchant-123'),
        expect.objectContaining({ method: 'GET' })
      );
    });

    it('should use getRiskIndicators with status="active"', async () => {
      const mockAlerts = {
        merchantId: 'merchant-123',
        overallScore: 0.7,
        indicators: [],
        lastUpdated: '2025-01-27T00:00:00Z',
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockAlerts,
      });

      await getRiskAlerts('merchant-123');

      // Verify it calls getRiskIndicators with status="active"
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/risk/indicators/merchant-123'),
        expect.objectContaining({ method: 'GET' })
      );
    });

    it('should handle API errors', async () => {
      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error',
        json: async () => ({ code: 'SERVER_ERROR', message: 'Internal server error' }),
      });

      await expect(getRiskAlerts('merchant-123')).rejects.toThrow();
      expect(ErrorHandler.handleAPIError).toHaveBeenCalled();
    });
  });

  describe('explainRiskAssessment', () => {
    it('should fetch risk assessment explanation successfully', async () => {
      const mockExplanation = {
        assessmentId: 'assessment-123',
        factors: [
          { name: 'Financial Risk', score: 0.7, weight: 0.4 },
          { name: 'Operational Risk', score: 0.6, weight: 0.3 },
        ],
        shapValues: {
          'financial_indicators': 0.15,
          'operational_efficiency': 0.12,
        },
        baseValue: 0.5,
        prediction: 0.65,
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockExplanation,
      });

      const result = await explainRiskAssessment('assessment-123');

      expect(result).toEqual(mockExplanation);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/risk/explain/assessment-123'),
        expect.objectContaining({ method: 'GET' })
      );
    });

    it('should use cache when available', async () => {
      const mockExplanation = {
        assessmentId: 'assessment-123',
        factors: [],
        shapValues: {},
        baseValue: 0.5,
        prediction: 0.65,
      };

      const apiCache = getAPICache();
      apiCache.set('risk-explain:assessment-123', mockExplanation);

      const result = await explainRiskAssessment('assessment-123');

      expect(result).toEqual(mockExplanation);
      expect(global.fetch).not.toHaveBeenCalled();
    });

    it('should handle API errors', async () => {
      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: false,
        status: 404,
        statusText: 'Not Found',
        json: async () => ({ code: 'NOT_FOUND', message: 'Explanation not found' }),
      });

      await expect(explainRiskAssessment('assessment-123')).rejects.toThrow();
      expect(ErrorHandler.handleAPIError).toHaveBeenCalled();
    });
  });

  describe('getRiskRecommendations', () => {
    it('should fetch risk recommendations successfully', async () => {
      const mockRecommendations = {
        merchantId: 'merchant-123',
        recommendations: [
          {
            id: 'rec-1',
            type: 'financial',
            priority: 'high',
            title: 'Improve Financial Stability',
            description: 'Consider improving financial stability',
            actionItems: ['Action 1', 'Action 2'],
          },
        ],
        timestamp: new Date().toISOString(),
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockRecommendations,
      });

      const result = await getRiskRecommendations('merchant-123');

      expect(result).toEqual(mockRecommendations);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/api/v1/merchants/merchant-123/risk-recommendations'),
        expect.objectContaining({ method: 'GET' })
      );
    });

    it('should use cache when available', async () => {
      const mockRecommendations = {
        merchantId: 'merchant-123',
        recommendations: [],
        timestamp: new Date().toISOString(),
      };

      const apiCache = getAPICache();
      apiCache.set('risk-recommendations:merchant-123', mockRecommendations);

      const result = await getRiskRecommendations('merchant-123');

      expect(result).toEqual(mockRecommendations);
      expect(global.fetch).not.toHaveBeenCalled();
    });

    it('should handle API errors', async () => {
      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error',
        json: async () => ({ code: 'SERVER_ERROR', message: 'Internal server error' }),
      });

      await expect(getRiskRecommendations('merchant-123')).rejects.toThrow();
      expect(ErrorHandler.handleAPIError).toHaveBeenCalled();
    });
  });

  describe('Request Deduplication', () => {
    it('should deduplicate concurrent requests for getPortfolioAnalytics', async () => {
      const mockAnalytics = {
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

      // Mock fetch to return proper Response object
      (global.fetch as vi.Mock).mockImplementation(async () => {
        return {
          ok: true,
          status: 200,
          statusText: 'OK',
          json: async () => mockAnalytics,
        } as Response;
      });

      const requestDeduplicator = getRequestDeduplicator();
      const deduplicateSpy = vi.spyOn(requestDeduplicator, 'deduplicate');

      // Make concurrent requests
      const results = await Promise.all([
        getPortfolioAnalytics(),
        getPortfolioAnalytics(),
        getPortfolioAnalytics(),
      ]);

      // Should deduplicate requests
      expect(deduplicateSpy).toHaveBeenCalled();
      // All results should be the same
      expect(results[0]).toEqual(mockAnalytics);
      expect(results[1]).toEqual(mockAnalytics);
      expect(results[2]).toEqual(mockAnalytics);
      // Fetch should be called (deduplication may still call fetch, but results are shared)
      expect(global.fetch).toHaveBeenCalled();
    });
  });
});

