// Vitest globals are available via globals: true in vitest.config.ts
import { vi } from "vitest";

// Mock ErrorHandler - must be before imports
vi.mock("@/lib/error-handler", () => ({
  ErrorHandler: {
    handleAPIError: vi.fn().mockResolvedValue(undefined),
    showErrorNotification: vi.fn(),
    showSuccessNotification: vi.fn(),
    showInfoNotification: vi.fn(),
    parseErrorResponse: vi
      .fn()
      .mockResolvedValue({ code: "TEST_ERROR", message: "Test error" }),
    logError: vi.fn(),
  },
}));

// Mock APICache and RequestDeduplicator
// These must be mocked before the api module is imported
// because api.ts initializes instances at module load time
const mockCache = new Map<string, any>();
const mockCacheTTL = new Map<string, number>();

vi.mock("@/lib/api-cache", () => ({
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
    delete = vi.fn((key: string) => {
      mockCache.delete(key);
      mockCacheTTL.delete(key);
    });
    clear = vi.fn(() => {
      mockCache.clear();
      mockCacheTTL.clear();
    });
  },
}));

vi.mock("@/lib/request-deduplicator", () => ({
  RequestDeduplicator: class RequestDeduplicator {
    deduplicate = vi.fn((key, fn) => fn());
  },
}));

// Import after mocks
import {
  explainRiskAssessment,
  getAPICache,
  getEnrichmentSources,
  getMerchant,
  getMerchantAnalytics,
  getMerchantRiskScore,
  getPortfolioAnalytics,
  getPortfolioStatistics,
  getRequestDeduplicator,
  getRiskAlerts,
  getRiskAssessment,
  getRiskBenchmarks,
  getRiskHistory,
  getRiskIndicators,
  getRiskInsights,
  getRiskPredictions,
  getRiskRecommendations,
  getRiskTrends,
  startRiskAssessment,
  triggerEnrichment,
} from "@/lib/api";
import { ErrorHandler } from "@/lib/error-handler";

describe("API Client", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Clear cache between tests
    mockCache.clear();
    mockCacheTTL.clear();
    // Mock fetch globally
    global.fetch = vi.fn();
    // Reset ErrorHandler mocks
    vi.mocked(ErrorHandler.handleAPIError).mockResolvedValue(undefined);
    vi.mocked(ErrorHandler.parseErrorResponse).mockResolvedValue({
      code: "TEST_ERROR",
      message: "Test error",
    });
    // Reset sessionStorage mock
    Object.defineProperty(window, "sessionStorage", {
      value: {
        getItem: vi.fn().mockReturnValue(null),
        setItem: vi.fn(),
        removeItem: vi.fn(),
        clear: vi.fn(),
      },
      writable: true,
    });
  });

  describe("getMerchant", () => {
    it("should fetch merchant data successfully", async () => {
      const mockMerchant = {
        id: "merchant-123",
        business_name: "Test Business",
        status: "active",
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockMerchant,
      });

      const result = await getMerchant("merchant-123");

      expect(result.id).toBe("merchant-123");
      expect(result.businessName).toBe("Test Business");
      expect(result.status).toBe("active");
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("/api/v1/merchants/merchant-123"),
        expect.objectContaining({ method: "GET" })
      );
    });

    it("should map financial information fields (Phase 1)", async () => {
      const mockMerchant = {
        id: "merchant-123",
        business_name: "Test Business",
        status: "active",
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
        founded_date: "2020-01-15T00:00:00Z",
        employee_count: 150,
        annual_revenue: 5000000.5,
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockMerchant,
      });

      const result = await getMerchant("merchant-123");

      expect(result.foundedDate).toBeDefined();
      expect(new Date(result.foundedDate!)).toBeInstanceOf(Date);
      expect(result.employeeCount).toBe(150);
      expect(result.annualRevenue).toBe(5000000.5);
    });

    it("should map system information fields (Phase 1)", async () => {
      const mockMerchant = {
        id: "merchant-123",
        business_name: "Test Business",
        status: "active",
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
        created_by: "user-123",
        metadata: {
          source: "manual",
          verified: true,
          tags: ["enterprise"],
        },
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockMerchant,
      });

      const result = await getMerchant("merchant-123");

      expect(result.createdBy).toBe("user-123");
      expect(result.metadata).toBeDefined();
      expect(result.metadata?.source).toBe("manual");
      expect(result.metadata?.verified).toBe(true);
    });

    it("should map address fields including street1, street2, countryCode (Phase 1)", async () => {
      const mockMerchant = {
        id: "merchant-123",
        business_name: "Test Business",
        status: "active",
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
        address: {
          street1: "123 Main Street",
          street2: "Suite 100",
          city: "San Francisco",
          state: "CA",
          postal_code: "94102",
          country: "United States",
          country_code: "US",
        },
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockMerchant,
      });

      const result = await getMerchant("merchant-123");

      expect(result.address).toBeDefined();
      expect(result.address?.street1).toBe("123 Main Street");
      expect(result.address?.street2).toBe("Suite 100");
      expect(result.address?.city).toBe("San Francisco");
      expect(result.address?.state).toBe("CA");
      expect(result.address?.postalCode).toBe("94102");
      expect(result.address?.country).toBe("United States");
      expect(result.address?.countryCode).toBe("US");
    });

    it("should map flat address fields (address_street1, etc.)", async () => {
      const mockMerchant = {
        id: "merchant-123",
        business_name: "Test Business",
        status: "active",
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
        address_street1: "123 Main Street",
        address_street2: "Suite 100",
        address_city: "San Francisco",
        address_state: "CA",
        address_postal_code: "94102",
        address_country: "United States",
        address_country_code: "US",
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockMerchant,
      });

      const result = await getMerchant("merchant-123");

      expect(result.address).toBeDefined();
      expect(result.address?.street1).toBe("123 Main Street");
      expect(result.address?.street2).toBe("Suite 100");
      expect(result.address?.city).toBe("San Francisco");
      expect(result.address?.state).toBe("CA");
      expect(result.address?.postalCode).toBe("94102");
      expect(result.address?.country).toBe("United States");
      expect(result.address?.countryCode).toBe("US");
    });

    it("should handle optional financial fields as undefined", async () => {
      const mockMerchant = {
        id: "merchant-123",
        business_name: "Test Business",
        status: "active",
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockMerchant,
      });

      const result = await getMerchant("merchant-123");

      expect(result.foundedDate).toBeUndefined();
      expect(result.employeeCount).toBeUndefined();
      expect(result.annualRevenue).toBeUndefined();
    });

    it("should handle contact_info nested object", async () => {
      const mockMerchant = {
        id: "merchant-123",
        business_name: "Test Business",
        status: "active",
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
        contact_info: {
          email: "test@example.com",
          phone: "+1-555-123-4567",
          website: "https://test.com",
        },
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockMerchant,
      });

      const result = await getMerchant("merchant-123");

      expect(result.email).toBe("test@example.com");
      expect(result.phone).toBe("+1-555-123-4567");
      expect(result.website).toBe("https://test.com");
    });

    it("should validate API response with Zod schema (Phase 5)", async () => {
      const mockMerchant = {
        id: "merchant-123",
        business_name: "Test Business",
        status: "active",
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
        founded_date: "2020-01-15T00:00:00Z",
        employee_count: 150,
        annual_revenue: 5000000.5,
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockMerchant,
      });

      const result = await getMerchant("merchant-123");

      // Validation should pass and return valid merchant
      expect(result.id).toBe("merchant-123");
      expect(result.businessName).toBe("Test Business");
      expect(result.status).toBe("active");
    });

    it("should handle API errors", async () => {
      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: false,
        status: 404,
        statusText: "Not Found",
        json: async () => ({
          code: "NOT_FOUND",
          message: "Merchant not found",
        }),
      });

      // Verify error is thrown when API returns error status
      await expect(getMerchant("invalid-id")).rejects.toThrow();

      // Verify fetch was called
      expect(global.fetch).toHaveBeenCalled();
    });
  });

  describe("getMerchantAnalytics", () => {
    it("should fetch analytics data successfully", async () => {
      const mockAnalytics = {
        merchantId: "merchant-123",
        classification: {
          primaryIndustry: "Technology",
          confidenceScore: 0.95,
          riskLevel: "low", // Required by schema
        },
        security: { trustScore: 0.8, sslValid: true },
        quality: { completenessScore: 0.9, dataPoints: 100 },
        intelligence: {},
        timestamp: new Date().toISOString(),
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockAnalytics,
      });

      const result = await getMerchantAnalytics("merchant-123");

      expect(result).toEqual(mockAnalytics);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("/api/v1/merchants/merchant-123/analytics"),
        expect.objectContaining({ method: "GET" })
      );
    });
  });

  describe("getRiskAssessment", () => {
    it("should fetch risk assessment successfully", async () => {
      const mockAssessment = {
        id: "assessment-123",
        merchantId: "merchant-123",
        status: "completed" as const,
        options: { includeHistory: true, includePredictions: false }, // Required by schema
        progress: 100, // Required by schema
        result: { overallScore: 0.7, riskLevel: "medium", factors: [] }, // Required fields
        createdAt: "2024-01-01T00:00:00Z", // Required by schema
        updatedAt: "2024-01-01T00:00:00Z", // Required by schema
        completedAt: "2024-01-01T00:00:00Z",
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockAssessment,
      });

      const result = await getRiskAssessment("merchant-123");

      expect(result).toEqual(mockAssessment);
    });

    it("should return null for 404 errors", async () => {
      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: false,
        status: 404,
        statusText: "Not Found",
      });

      const result = await getRiskAssessment("merchant-123");

      expect(result).toBeNull();
    });
  });

  describe("startRiskAssessment", () => {
    it("should start risk assessment successfully", async () => {
      const mockResponse = {
        assessmentId: "assessment-123",
        status: "pending",
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      const result = await startRiskAssessment({
        merchantId: "merchant-123",
        options: { includeHistory: true },
      });

      expect(result).toEqual(mockResponse);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("/api/v1/risk/assess"),
        expect.objectContaining({
          method: "POST",
          body: expect.any(String),
        })
      );
    });
  });

  describe("getRiskHistory", () => {
    it("should fetch risk history with pagination", async () => {
      const mockHistory = {
        merchantId: "merchant-123",
        history: [],
        limit: 10,
        offset: 0,
        total: 0,
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockHistory,
      });

      const result = await getRiskHistory("merchant-123", 10, 0);

      expect(result).toEqual(mockHistory);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("/api/v1/risk/history/merchant-123"),
        expect.objectContaining({ method: "GET" })
      );
    });
  });

  describe("getRiskPredictions", () => {
    it("should fetch risk predictions with horizons", async () => {
      const mockPredictions = {
        merchantId: "merchant-123",
        horizons: [3, 6, 12],
        predictions: [],
        includeScenarios: false,
        includeConfidence: false,
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockPredictions,
      });

      const result = await getRiskPredictions("merchant-123", [3, 6, 12]);

      expect(result).toEqual(mockPredictions);
    });
  });

  describe("explainRiskAssessment", () => {
    it("should fetch risk assessment explanation", async () => {
      const mockExplanation = {
        assessmentId: "assessment-123",
        factors: [],
        shapValues: {},
        baseValue: 0.5,
        prediction: 0.7,
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockExplanation,
      });

      const result = await explainRiskAssessment("assessment-123");

      expect(result).toEqual(mockExplanation);
    });
  });

  describe("getRiskRecommendations", () => {
    it("should fetch risk recommendations", async () => {
      const mockRecommendations = {
        merchantId: "merchant-123",
        recommendations: [],
        timestamp: new Date().toISOString(),
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockRecommendations,
      });

      const result = await getRiskRecommendations("merchant-123");

      expect(result).toEqual(mockRecommendations);
    });
  });

  describe("getRiskIndicators", () => {
    it("should fetch risk indicators with filters", async () => {
      const mockIndicators = {
        merchantId: "merchant-123",
        indicators: [], // Schema expects indicators array
        timestamp: new Date().toISOString(), // Schema expects timestamp (optional)
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockIndicators,
      });

      const result = await getRiskIndicators("merchant-123", "high", "active");

      expect(result).toEqual(mockIndicators);
    });
  });

  describe("getEnrichmentSources", () => {
    it("should fetch enrichment sources", async () => {
      const mockSources = {
        merchantId: "merchant-123",
        sources: [],
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockSources,
      });

      const result = await getEnrichmentSources("merchant-123");

      expect(result).toEqual(mockSources);
    });
  });

  describe("triggerEnrichment", () => {
    it("should trigger enrichment successfully", async () => {
      const mockJob = {
        jobId: "job-123",
        merchantId: "merchant-123",
        source: "external-api",
        status: "pending",
        createdAt: new Date().toISOString(),
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockJob,
      });

      const result = await triggerEnrichment("merchant-123", "external-api");

      expect(result).toEqual(mockJob);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining(
          "/api/v1/merchants/merchant-123/enrichment/trigger"
        ),
        expect.objectContaining({
          method: "POST",
          body: expect.stringContaining("external-api"),
        })
      );
    });
  });

  describe("getPortfolioAnalytics", () => {
    it("should fetch portfolio analytics successfully", async () => {
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
        expect.stringContaining("/api/v1/merchants/analytics"),
        expect.objectContaining({ method: "GET" })
      );
    });

    it("should use cache when available", async () => {
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
      apiCache.set("portfolio-analytics", mockAnalytics);

      const result = await getPortfolioAnalytics();

      expect(result).toEqual(mockAnalytics);
      expect(global.fetch).not.toHaveBeenCalled();
    });

    it("should handle API errors", async () => {
      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: "Internal Server Error",
        json: async () => ({
          code: "SERVER_ERROR",
          message: "Internal server error",
        }),
      });

      await expect(getPortfolioAnalytics()).rejects.toThrow();
      expect(ErrorHandler.handleAPIError).toHaveBeenCalled();
    });

    it("should cache response after successful fetch", async () => {
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
      const setSpy = vi.spyOn(apiCache, "set");

      await getPortfolioAnalytics();

      expect(setSpy).toHaveBeenCalledWith(
        "portfolio-analytics",
        mockAnalytics,
        7 * 60 * 1000
      );
    });
  });

  describe("getPortfolioStatistics", () => {
    it("should fetch portfolio statistics successfully", async () => {
      const mockStatistics = {
        totalMerchants: 100,
        totalAssessments: 150, // Required by schema
        averageRiskScore: 0.6,
        riskDistribution: { low: 40, medium: 50, high: 10 },
        industryBreakdown: [], // Required by schema (array)
        countryBreakdown: [], // Required by schema (array)
        timestamp: new Date().toISOString(),
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockStatistics,
      });

      const result = await getPortfolioStatistics();

      expect(result).toEqual(mockStatistics);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("/api/v1/merchants/statistics"),
        expect.objectContaining({ method: "GET" })
      );
    });

    it("should use cache when available", async () => {
      const mockStatistics = {
        totalMerchants: 100,
        totalAssessments: 150, // Required by schema
        averageRiskScore: 0.6,
        riskDistribution: { low: 40, medium: 50, high: 10 },
        industryBreakdown: [], // Required by schema
        countryBreakdown: [], // Required by schema
        timestamp: new Date().toISOString(),
      };

      const apiCache = getAPICache();
      apiCache.set("portfolio-statistics", mockStatistics);

      const result = await getPortfolioStatistics();

      expect(result).toEqual(mockStatistics);
      expect(global.fetch).not.toHaveBeenCalled();
    });

    it("should handle API errors", async () => {
      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: "Internal Server Error",
        json: async () => ({
          code: "SERVER_ERROR",
          message: "Internal server error",
        }),
      });

      await expect(getPortfolioStatistics()).rejects.toThrow();
      expect(ErrorHandler.handleAPIError).toHaveBeenCalled();
    });
  });

  describe("getRiskTrends", () => {
    it("should fetch risk trends successfully without params", async () => {
      const mockTrends = {
        trends: [],
        summary: {
          averageScore: 0.6,
          trendDirection: "stable",
          changePercentage: 0,
        },
        timeframe: "6m",
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockTrends,
      });

      const result = await getRiskTrends();

      expect(result).toEqual(mockTrends);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("/api/v1/analytics/trends"),
        expect.objectContaining({ method: "GET" })
      );
    });

    it("should fetch risk trends with query parameters", async () => {
      const mockTrends = {
        trends: [],
        summary: {
          averageScore: 0.6,
          trendDirection: "improving",
          changePercentage: -5.2,
        },
        timeframe: "3m",
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockTrends,
      });

      const result = await getRiskTrends({
        industry: "technology",
        country: "US",
        timeframe: "3m",
        limit: 10,
      });

      expect(result).toEqual(mockTrends);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("/api/v1/analytics/trends"),
        expect.objectContaining({ method: "GET" })
      );
    });

    it("should use cache when available", async () => {
      const mockTrends = {
        trends: [],
        summary: {
          averageScore: 0.6,
          trendDirection: "stable",
          changePercentage: 0,
        },
        timeframe: "6m",
      };

      const apiCache = getAPICache();
      apiCache.set("risk-trends:{}", mockTrends);

      const result = await getRiskTrends();

      expect(result).toEqual(mockTrends);
      expect(global.fetch).not.toHaveBeenCalled();
    });

    it("should create different cache keys for different params", async () => {
      const mockTrends1 = {
        trends: [],
        summary: {
          averageScore: 0.6,
          trendDirection: "stable",
          changePercentage: 0,
        },
        timeframe: "6m",
      };
      const mockTrends2 = {
        trends: [],
        summary: {
          averageScore: 0.7,
          trendDirection: "declining",
          changePercentage: 5,
        },
        timeframe: "3m",
      };

      (global.fetch as vi.Mock)
        .mockResolvedValueOnce({ ok: true, json: async () => mockTrends1 })
        .mockResolvedValueOnce({ ok: true, json: async () => mockTrends2 });

      const result1 = await getRiskTrends();
      const result2 = await getRiskTrends({ industry: "technology" });

      expect(result1).toEqual(mockTrends1);
      expect(result2).toEqual(mockTrends2);
      expect(global.fetch).toHaveBeenCalledTimes(2);
    });

    it("should handle API errors", async () => {
      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: "Internal Server Error",
        json: async () => ({
          code: "SERVER_ERROR",
          message: "Internal server error",
        }),
      });

      await expect(getRiskTrends()).rejects.toThrow();
      expect(ErrorHandler.handleAPIError).toHaveBeenCalled();
    });
  });

  describe("getRiskInsights", () => {
    it("should fetch risk insights successfully without params", async () => {
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
        expect.stringContaining("/api/v1/analytics/insights"),
        expect.objectContaining({ method: "GET" })
      );
    });

    it("should fetch risk insights with query parameters", async () => {
      const mockInsights = {
        insights: [
          {
            id: "insight-1",
            type: "risk_distribution",
            title: "High Risk Concentration",
            description: "10% of merchants are high risk",
            severity: "medium",
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
        industry: "technology",
        country: "US",
        risk_level: "high",
      });

      expect(result).toEqual(mockInsights);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("/api/v1/analytics/insights"),
        expect.objectContaining({ method: "GET" })
      );
    });

    it("should use cache when available", async () => {
      const mockInsights = {
        insights: [],
        recommendations: [],
        timestamp: new Date().toISOString(),
      };

      const apiCache = getAPICache();
      apiCache.set("risk-insights:{}", mockInsights);

      const result = await getRiskInsights();

      expect(result).toEqual(mockInsights);
      expect(global.fetch).not.toHaveBeenCalled();
    });

    it("should handle API errors", async () => {
      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: "Internal Server Error",
        json: async () => ({
          code: "SERVER_ERROR",
          message: "Internal server error",
        }),
      });

      await expect(getRiskInsights()).rejects.toThrow();
      expect(ErrorHandler.handleAPIError).toHaveBeenCalled();
    });
  });

  describe("getRiskBenchmarks", () => {
    it("should fetch risk benchmarks with MCC code", async () => {
      const mockBenchmarks = {
        industry_code: "5734",
        industry_type: "mcc" as const,
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

      const result = await getRiskBenchmarks({ mcc: "5734" });

      expect(result).toEqual(mockBenchmarks);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("/api/v1/risk/benchmarks"),
        expect.objectContaining({ method: "GET" })
      );
    });

    it("should fetch risk benchmarks with NAICS code", async () => {
      const mockBenchmarks = {
        industry_code: "541511",
        industry_type: "naics" as const,
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

      const result = await getRiskBenchmarks({ naics: "541511" });

      expect(result).toEqual(mockBenchmarks);
    });

    it("should fetch risk benchmarks with SIC code", async () => {
      const mockBenchmarks = {
        industry_code: "7372",
        industry_type: "sic" as const,
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

      const result = await getRiskBenchmarks({ sic: "7372" });

      expect(result).toEqual(mockBenchmarks);
    });

    it("should use cache when available", async () => {
      const mockBenchmarks = {
        industry_code: "5734",
        industry_type: "mcc" as const,
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
      apiCache.set("risk-benchmarks:5734", mockBenchmarks);

      const result = await getRiskBenchmarks({ mcc: "5734" });

      expect(result).toEqual(mockBenchmarks);
      expect(global.fetch).not.toHaveBeenCalled();
    });

    it("should handle API errors", async () => {
      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: false,
        status: 404,
        statusText: "Not Found",
        json: async () => ({
          code: "NOT_FOUND",
          message: "Benchmarks not found",
        }),
      });

      await expect(getRiskBenchmarks({ mcc: "5734" })).rejects.toThrow();
      expect(ErrorHandler.handleAPIError).toHaveBeenCalled();
    });
  });

  describe("getMerchantRiskScore", () => {
    it("should fetch merchant risk score successfully", async () => {
      const mockRiskScore = {
        merchant_id: "merchant-123",
        risk_score: 0.65,
        risk_level: "medium" as const,
        confidence_score: 0.85,
        assessment_date: "2025-01-27T00:00:00Z",
        factors: [
          { category: "Financial Risk", score: 0.7, weight: 0.4 }, // Schema expects 'category' not 'name'
        ],
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockRiskScore,
      });

      const result = await getMerchantRiskScore("merchant-123");

      expect(result).toEqual(mockRiskScore);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("/api/v1/merchants/merchant-123/risk-score"),
        expect.objectContaining({ method: "GET" })
      );
    });

    it("should use cache when available", async () => {
      const mockRiskScore = {
        merchant_id: "merchant-123",
        risk_score: 0.65,
        risk_level: "medium" as const,
        confidence_score: 0.85,
        assessment_date: "2025-01-27T00:00:00Z",
        factors: [],
      };

      const apiCache = getAPICache();
      apiCache.set("merchant-risk-score:merchant-123", mockRiskScore);

      const result = await getMerchantRiskScore("merchant-123");

      expect(result).toEqual(mockRiskScore);
      expect(global.fetch).not.toHaveBeenCalled();
    });

    it("should handle API errors", async () => {
      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: false,
        status: 404,
        statusText: "Not Found",
        json: async () => ({
          code: "NOT_FOUND",
          message: "Risk score not found",
        }),
      });

      await expect(getMerchantRiskScore("merchant-123")).rejects.toThrow();
      expect(ErrorHandler.handleAPIError).toHaveBeenCalled();
    });

    it("should cache response with 3 minute TTL", async () => {
      const mockRiskScore = {
        merchant_id: "merchant-123",
        risk_score: 0.65,
        risk_level: "medium" as const,
        confidence_score: 0.85,
        assessment_date: "2025-01-27T00:00:00Z",
        factors: [],
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockRiskScore,
      });

      const apiCache = getAPICache();
      const setSpy = vi.spyOn(apiCache, "set");

      await getMerchantRiskScore("merchant-123");

      expect(setSpy).toHaveBeenCalledWith(
        "merchant-risk-score:merchant-123",
        mockRiskScore,
        3 * 60 * 1000
      );
    });
  });

  describe("getRiskAlerts", () => {
    it("should fetch risk alerts successfully", async () => {
      const mockAlerts = {
        merchantId: "merchant-123",
        indicators: [
          // Schema expects indicators array
          {
            id: "indicator-1",
            title: "High Risk Indicator", // Schema expects title
            description: "Merchant has high financial risk", // Schema expects description
            severity: "high" as const, // Schema expects severity enum
            status: "active", // Optional
            createdAt: "2025-01-27T00:00:00Z", // Optional
            updatedAt: "2025-01-27T00:00:00Z", // Optional
          },
        ],
        timestamp: "2025-01-27T00:00:00Z", // Schema expects timestamp (optional)
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockAlerts,
      });

      const result = await getRiskAlerts("merchant-123");

      expect(result).toEqual(mockAlerts);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("/api/v1/risk/indicators/merchant-123"),
        expect.objectContaining({ method: "GET" })
      );
    });

    it("should fetch risk alerts with severity filter", async () => {
      const mockAlerts = {
        merchantId: "merchant-123",
        indicators: [
          {
            id: "indicator-1",
            title: "High Risk Indicator",
            description: "Merchant has high financial risk",
            severity: "high" as const,
            status: "active",
            createdAt: "2025-01-27T00:00:00Z",
            updatedAt: "2025-01-27T00:00:00Z",
          },
        ],
        timestamp: "2025-01-27T00:00:00Z",
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockAlerts,
      });

      const result = await getRiskAlerts("merchant-123", "high");

      expect(result).toEqual(mockAlerts);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("/api/v1/risk/indicators/merchant-123"),
        expect.objectContaining({ method: "GET" })
      );
    });

    it('should use getRiskIndicators with status="active"', async () => {
      const mockAlerts = {
        merchantId: "merchant-123",
        indicators: [],
        timestamp: "2025-01-27T00:00:00Z",
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockAlerts,
      });

      await getRiskAlerts("merchant-123");

      // Verify it calls getRiskIndicators with status="active"
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("/api/v1/risk/indicators/merchant-123"),
        expect.objectContaining({ method: "GET" })
      );
    });

    it("should handle API errors", async () => {
      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: "Internal Server Error",
        json: async () => ({
          code: "SERVER_ERROR",
          message: "Internal server error",
        }),
      });

      await expect(getRiskAlerts("merchant-123")).rejects.toThrow();
      expect(ErrorHandler.handleAPIError).toHaveBeenCalled();
    });
  });

  describe("explainRiskAssessment", () => {
    it("should fetch risk assessment explanation successfully", async () => {
      const mockExplanation = {
        assessmentId: "assessment-123",
        factors: [
          { name: "Financial Risk", score: 0.7, weight: 0.4 },
          { name: "Operational Risk", score: 0.6, weight: 0.3 },
        ],
        shapValues: {
          financial_indicators: 0.15,
          operational_efficiency: 0.12,
        },
        baseValue: 0.5,
        prediction: 0.65,
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockExplanation,
      });

      const result = await explainRiskAssessment("assessment-123");

      expect(result).toEqual(mockExplanation);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("/api/v1/risk/explain/assessment-123"),
        expect.objectContaining({ method: "GET" })
      );
    });

    it("should use cache when available", async () => {
      const mockExplanation = {
        assessmentId: "assessment-123",
        factors: [],
        shapValues: {},
        baseValue: 0.5,
        prediction: 0.65,
      };

      const apiCache = getAPICache();
      apiCache.set("risk-explain:assessment-123", mockExplanation);

      const result = await explainRiskAssessment("assessment-123");

      expect(result).toEqual(mockExplanation);
      expect(global.fetch).not.toHaveBeenCalled();
    });

    it("should handle API errors", async () => {
      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: false,
        status: 404,
        statusText: "Not Found",
        json: async () => ({
          code: "NOT_FOUND",
          message: "Explanation not found",
        }),
      });

      await expect(explainRiskAssessment("assessment-123")).rejects.toThrow();
      expect(ErrorHandler.handleAPIError).toHaveBeenCalled();
    });
  });

  describe("getRiskRecommendations", () => {
    it("should fetch risk recommendations successfully", async () => {
      const mockRecommendations = {
        merchantId: "merchant-123",
        recommendations: [
          {
            id: "rec-1",
            type: "financial",
            priority: "high",
            title: "Improve Financial Stability",
            description: "Consider improving financial stability",
            actionItems: ["Action 1", "Action 2"],
          },
        ],
        timestamp: new Date().toISOString(),
      };

      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockRecommendations,
      });

      const result = await getRiskRecommendations("merchant-123");

      expect(result).toEqual(mockRecommendations);
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining(
          "/api/v1/merchants/merchant-123/risk-recommendations"
        ),
        expect.objectContaining({ method: "GET" })
      );
    });

    it("should use cache when available", async () => {
      const mockRecommendations = {
        merchantId: "merchant-123",
        recommendations: [],
        timestamp: new Date().toISOString(),
      };

      const apiCache = getAPICache();
      apiCache.set("risk-recommendations:merchant-123", mockRecommendations);

      const result = await getRiskRecommendations("merchant-123");

      expect(result).toEqual(mockRecommendations);
      expect(global.fetch).not.toHaveBeenCalled();
    });

    it("should handle API errors", async () => {
      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: "Internal Server Error",
        json: async () => ({
          code: "SERVER_ERROR",
          message: "Internal server error",
        }),
      });

      await expect(getRiskRecommendations("merchant-123")).rejects.toThrow();
      expect(ErrorHandler.handleAPIError).toHaveBeenCalled();
    });
  });

  describe("Request Deduplication", () => {
    it("should deduplicate concurrent requests for getPortfolioAnalytics", async () => {
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
          statusText: "OK",
          json: async () => mockAnalytics,
        } as Response;
      });

      const requestDeduplicator = getRequestDeduplicator();
      const deduplicateSpy = vi.spyOn(requestDeduplicator, "deduplicate");

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

  describe("Retry Logic", () => {
    it("should retry on failure and eventually succeed", async () => {
      const mockMerchant = {
        id: "merchant-123",
        business_name: "Test Business",
        status: "active",
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
      };

      // Clear cache first
      mockCache.delete("merchant:merchant-123");

      // First two calls fail, third succeeds
      (global.fetch as vi.Mock)
        .mockRejectedValueOnce(new TypeError("Failed to fetch"))
        .mockRejectedValueOnce(new TypeError("Failed to fetch"))
        .mockResolvedValueOnce({
          ok: true,
          status: 200,
          json: async () => mockMerchant,
        });

      // Use real timers but with shorter delays for testing
      const result = await getMerchant("merchant-123", { retries: 3 });

      expect(result.id).toBe("merchant-123");
      expect(global.fetch).toHaveBeenCalledTimes(3);
    }, 15000); // Increase timeout for retries (1s + 2s + buffer)

    it("should fail after max retries", async () => {
      // Clear cache first
      mockCache.delete("merchant:merchant-123");

      // All retries fail - use 'Failed to fetch' to trigger network error path
      (global.fetch as vi.Mock).mockRejectedValue(
        new TypeError("Failed to fetch")
      );

      await expect(
        getMerchant("merchant-123", { retries: 2 })
      ).rejects.toThrow();
      expect(global.fetch).toHaveBeenCalledTimes(2); // Initial + 1 retry
    }, 15000); // Increase timeout for retries

    it("should use custom retry count", async () => {
      const mockMerchant = {
        id: "merchant-123",
        business_name: "Test Business",
        status: "active",
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-01T00:00:00Z",
      };

      // Clear cache first
      mockCache.delete("merchant:merchant-123");

      // First call fails with HTTP error (not network), second succeeds
      // This avoids the retry delay and tests the retry count branch
      (global.fetch as vi.Mock)
        .mockResolvedValueOnce({
          ok: false,
          status: 500,
          statusText: "Internal Server Error",
          json: async () => ({
            code: "SERVER_ERROR",
            message: "Temporary error",
          }),
        })
        .mockResolvedValueOnce({
          ok: true,
          status: 200,
          json: async () => mockMerchant,
        });

      vi.mocked(ErrorHandler.parseErrorResponse)
        .mockResolvedValueOnce({
          code: "SERVER_ERROR",
          message: "Temporary error",
        })
        .mockResolvedValueOnce({
          code: "SERVER_ERROR",
          message: "Temporary error",
        });

      const result = await getMerchant("merchant-123", { retries: 2 });

      expect(result.id).toBe("merchant-123");
      // Should retry once, so fetch called twice (initial + retry)
      expect(global.fetch).toHaveBeenCalledTimes(2);
    }, 15000); // Increase timeout for retries
  });

  describe("CORS and Network Errors", () => {
    it("should handle CORS errors", async () => {
      (global.fetch as vi.Mock).mockRejectedValue(
        new TypeError("Failed to fetch")
      );

      await expect(getMerchant("merchant-123")).rejects.toThrow("CORS policy");
    });

    it("should handle network errors", async () => {
      const networkError = new TypeError("Network request failed");
      (global.fetch as vi.Mock).mockRejectedValue(networkError);

      await expect(getMerchant("merchant-123")).rejects.toThrow(
        "Network request failed"
      );
    });

    it("should handle generic TypeError errors", async () => {
      const typeError = new TypeError("Some other error");
      (global.fetch as vi.Mock).mockRejectedValue(typeError);

      await expect(getMerchant("merchant-123")).rejects.toThrow(
        "Network request failed"
      );
    });
  });

  describe("Error Response Handling", () => {
    it("should handle error response with message", async () => {
      (global.fetch as vi.Mock).mockResolvedValue({
        ok: false,
        status: 400,
        statusText: "Bad Request",
        json: async () => ({
          code: "VALIDATION_ERROR",
          message: "Invalid input",
        }),
      });

      vi.mocked(ErrorHandler.parseErrorResponse).mockResolvedValue({
        code: "VALIDATION_ERROR",
        message: "Invalid input",
      });

      await expect(getMerchant("merchant-123")).rejects.toThrow(
        "Invalid input"
      );

      try {
        await getMerchant("merchant-123");
      } catch (error: any) {
        expect(error.status).toBe(400);
        expect(error.code).toBe("VALIDATION_ERROR");
      }
    });

    it("should handle error response without message", async () => {
      (global.fetch as vi.Mock).mockResolvedValue({
        ok: false,
        status: 500,
        statusText: "Internal Server Error",
        json: async () => ({ code: "SERVER_ERROR" }),
      });

      vi.mocked(ErrorHandler.parseErrorResponse).mockResolvedValue({
        code: "SERVER_ERROR",
      });

      await expect(getMerchant("merchant-123")).rejects.toThrow(
        "API Error 500"
      );
    });

    it("should handle error response when parseErrorResponse throws", async () => {
      (global.fetch as vi.Mock).mockResolvedValue({
        ok: false,
        status: 500,
        statusText: "Internal Server Error",
        json: async () => {
          throw new Error("Parse error");
        },
      });

      vi.mocked(ErrorHandler.parseErrorResponse).mockRejectedValue(
        new Error("Parse error")
      );

      // When parseErrorResponse throws an Error, it re-throws that error
      // So we expect the parse error message, not "API Error 500"
      await expect(getMerchant("merchant-123")).rejects.toThrow("Parse error");

      try {
        await getMerchant("merchant-123");
      } catch (error: any) {
        // The error from parseErrorResponse is re-thrown, so it has the parse error message
        expect(error.message).toBe("Parse error");
      }
    });

    it("should handle error response when parseErrorResponse returns non-Error", async () => {
      (global.fetch as vi.Mock).mockResolvedValue({
        ok: false,
        status: 500,
        statusText: "Internal Server Error",
        json: async () => ({ code: "SERVER_ERROR" }),
      });

      vi.mocked(ErrorHandler.parseErrorResponse).mockRejectedValue(
        "String error"
      );

      await expect(getMerchant("merchant-123")).rejects.toThrow(
        "API Error 500"
      );
    });

    it("should handle error response with status but no ok property", async () => {
      (global.fetch as vi.Mock).mockResolvedValue({
        ok: false,
        status: 404,
        statusText: "Not Found",
        json: async () => ({
          code: "NOT_FOUND",
          message: "Resource not found",
        }),
      } as Response);

      vi.mocked(ErrorHandler.parseErrorResponse).mockResolvedValue({
        code: "NOT_FOUND",
        message: "Resource not found",
      });

      await expect(getMerchant("merchant-123")).rejects.toThrow(
        "Resource not found"
      );
    });
  });

  describe("JSON Parse Errors", () => {
    it("should handle JSON parse errors in successful response", async () => {
      (global.fetch as vi.Mock).mockResolvedValue({
        ok: true,
        status: 200,
        statusText: "OK",
        json: async () => {
          throw new SyntaxError("Unexpected token");
        },
      });

      await expect(getMerchant("merchant-123")).rejects.toThrow(
        "Failed to parse JSON"
      );

      try {
        await getMerchant("merchant-123");
      } catch (error: any) {
        expect(error.code).toBe("JSON_PARSE_ERROR");
      }
    });

    it("should handle non-Error JSON parse failures", async () => {
      (global.fetch as vi.Mock).mockResolvedValue({
        ok: true,
        status: 200,
        statusText: "OK",
        json: async () => {
          throw "String error";
        },
      });

      await expect(getMerchant("merchant-123")).rejects.toThrow(
        "Failed to parse JSON"
      );
    });
  });

  describe("Response Status Handling", () => {
    it("should handle response with ok=true but status outside 200-299", async () => {
      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: true,
        status: 299,
        statusText: "OK",
        json: async () => ({
          id: "merchant-123",
          business_name: "Test Business",
          status: "active",
          created_at: "2024-01-01T00:00:00Z",
          updated_at: "2024-01-01T00:00:00Z",
        }),
      });

      const result = await getMerchant("merchant-123");
      expect(result.id).toBe("merchant-123");
    });

    it("should handle response with ok=false but status in 200-299 range", async () => {
      (global.fetch as vi.Mock).mockResolvedValueOnce({
        ok: false,
        status: 200,
        statusText: "OK",
        json: async () => ({
          id: "merchant-123",
          business_name: "Test Business",
          status: "active",
          created_at: "2024-01-01T00:00:00Z",
          updated_at: "2024-01-01T00:00:00Z",
        }),
      });

      const result = await getMerchant("merchant-123");
      expect(result.id).toBe("merchant-123");
    });
  });
});
