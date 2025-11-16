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
vi.mock('@/lib/api-cache', () => ({
  APICache: class APICache {
    constructor(ttl?: number) {}
    get = vi.fn().mockReturnValue(null);
    set = vi.fn();
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
} from '@/lib/api';
import { ErrorHandler } from '@/lib/error-handler';

describe('API Client', () => {
  beforeEach(() => {
    vi.clearAllMocks();
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
        expect.stringContaining('/api/v1/risk/history/merchant-123?limit=10&offset=0'),
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
});

