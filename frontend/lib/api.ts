// API client for Merchant Details
import { APICache } from "@/lib/api-cache";
import { ApiEndpoints } from "@/lib/api-config";
import {
  AnalyticsDataSchema,
  AssessmentStatusResponseSchema,
  ComplianceStatusSchema,
  DashboardMetricsSchema,
  MerchantListResponseSchema,
  MerchantRiskScoreSchema,
  MerchantSchema,
  PortfolioStatisticsSchema,
  RiskAssessmentSchema,
  RiskBenchmarksSchema,
  RiskHistoryResponseSchema,
  RiskIndicatorsDataSchema,
  RiskMetricsSchema,
  RiskRecommendationsResponseSchema,
  SystemMetricsSchema,
  validateAPIResponse,
} from "@/lib/api-validation";
import { ErrorHandler } from "@/lib/error-handler";
import { RequestDeduplicator } from "@/lib/request-deduplicator";
import type {
  BusinessIntelligenceMetrics,
  ComplianceStatus,
  DashboardMetrics,
  RiskMetrics,
  SystemMetrics,
} from "@/types/dashboard";
import type {
  AnalyticsData,
  AssessmentStatusResponse,
  EnrichmentSource,
  Merchant,
  MerchantListParams,
  MerchantListResponse,
  MerchantRiskScore,
  PortfolioAnalytics,
  PortfolioStatistics,
  RiskAssessment,
  RiskAssessmentRequest,
  RiskAssessmentResponse,
  RiskBenchmarks,
  RiskIndicatorsData,
  RiskInsights,
  RiskTrends,
  WebsiteAnalysisData,
} from "@/types/merchant";

// API_BASE_URL is now accessed via ApiEndpoints - keeping for backward compatibility if needed
// const API_BASE_URL = getApiBaseUrl();

// Initialize optimization utilities
const apiCache = new APICache(5 * 60 * 1000); // 5 minutes default TTL
const requestDeduplicator = new RequestDeduplicator();

// Export for test cleanup - allows clearing cache/deduplicator between tests
export const getAPICache = () => apiCache;
export const getRequestDeduplicator = () => requestDeduplicator;

// Helper function to get auth token
function getAuthToken(): string | null {
  if (typeof window === "undefined") return null;
  return sessionStorage.getItem("authToken");
}

// Helper function to wrap fetch and handle CORS/network errors
async function safeFetch(
  url: string,
  options?: RequestInit
): Promise<Response> {
  try {
    const response = await fetch(url, options);
    return response;
  } catch (error) {
    // Handle CORS and network errors
    if (
      error instanceof TypeError &&
      error.message.includes("Failed to fetch")
    ) {
      // Check if it's a CORS error
      const corsError = new Error(
        "CORS policy blocked the request. Please check server configuration."
      );
      (corsError as any).code = "CORS_ERROR";
      (corsError as any).isCORS = true;
      throw corsError;
    }
    // Handle other network errors
    if (error instanceof TypeError) {
      const networkError = new Error(
        "Network request failed. Please check your connection."
      );
      (networkError as any).code = "NETWORK_ERROR";
      throw networkError;
    }
    throw error;
  }
}

// Helper function to handle API errors
async function handleResponse<T>(response: Response): Promise<T> {
  // Check response status - MSW responses should have status and ok set correctly
  const status = response.status;
  const isOk = response.ok === true || (status >= 200 && status < 300);

  // Debug logging in test environment
  if (process.env.NODE_ENV === "test" && !isOk) {
    console.log(
      "[API Debug] Error response - status:",
      status,
      "ok:",
      response.ok,
      "statusText:",
      response.statusText
    );
  }

  if (!isOk) {
    try {
      const errorData = await ErrorHandler.parseErrorResponse(response);
      const errorMessage =
        errorData && typeof errorData === "object" && "message" in errorData
          ? String(errorData.message)
          : `API Error ${status}: ${response.statusText}`;

      // Create error with status code for better error handling
      const error = new Error(errorMessage);
      (error as any).status = status;
      (error as any).code =
        errorData && typeof errorData === "object" && "code" in errorData
          ? String(errorData.code)
          : `HTTP_${status}`;
      throw error;
    } catch (err) {
      // If parsing error response fails, create error with status
      if (err instanceof Error) {
        throw err;
      }
      const error = new Error(`API Error ${status}: ${response.statusText}`);
      (error as any).status = status;
      (error as any).code = `HTTP_${status}`;
      throw error;
    }
  }

  // Safely parse JSON response
  try {
    const json = await response.json();
    return json as T;
  } catch (error) {
    // If JSON parsing fails, throw a more helpful error
    const errorMessage = error instanceof Error ? error.message : String(error);
    const parseError = new Error(
      `Failed to parse JSON response: ${errorMessage}`
    );
    (parseError as any).code = "JSON_PARSE_ERROR";
    throw parseError;
  }
}

// Helper function to map address from backend (handles both nested map and flat fields)
function mapAddress(
  addressData: unknown,
  rawData: Record<string, unknown>
): Merchant["address"] | undefined {
  const getString = (obj: unknown, key: string): string | undefined => {
    if (obj && typeof obj === "object" && key in obj) {
      const value = (obj as Record<string, unknown>)[key];
      return typeof value === "string" ? value : undefined;
    }
    return undefined;
  };

  // If address is a map/object, extract fields from it
  if (addressData && typeof addressData === "object" && addressData !== null) {
    const addr = addressData as Record<string, unknown>;
    return {
      street: getString(addr, "street") || getString(addr, "street1"),
      street1: getString(addr, "street1"),
      street2: getString(addr, "street2"),
      city: getString(addr, "city"),
      state: getString(addr, "state"),
      postalCode:
        getString(addr, "postal_code") || getString(addr, "postalCode"),
      country: getString(addr, "country"),
      countryCode:
        getString(addr, "country_code") || getString(addr, "countryCode"),
    };
  }

  // If address fields are flat in rawData, extract them
  if (rawData.address_street1 || rawData.address_city) {
    return {
      street:
        typeof rawData.address_street1 === "string"
          ? rawData.address_street1
          : undefined,
      street1:
        typeof rawData.address_street1 === "string"
          ? rawData.address_street1
          : undefined,
      street2:
        typeof rawData.address_street2 === "string"
          ? rawData.address_street2
          : undefined,
      city:
        typeof rawData.address_city === "string"
          ? rawData.address_city
          : undefined,
      state:
        typeof rawData.address_state === "string"
          ? rawData.address_state
          : undefined,
      postalCode:
        typeof rawData.address_postal_code === "string"
          ? rawData.address_postal_code
          : undefined,
      country:
        typeof rawData.address_country === "string"
          ? rawData.address_country
          : undefined,
      countryCode:
        typeof rawData.address_country_code === "string"
          ? rawData.address_country_code
          : undefined,
    };
  }

  return undefined;
}

// Retry logic with exponential backoff
async function retryWithBackoff<T>(
  fn: () => Promise<T>,
  maxRetries = 3,
  initialDelay = 1000
): Promise<T> {
  let lastError: Error;
  let delay = initialDelay;

  for (let i = 0; i < maxRetries; i++) {
    try {
      return await fn();
    } catch (error) {
      lastError = error as Error;
      if (i < maxRetries - 1) {
        await new Promise((resolve) => setTimeout(resolve, delay));
        delay *= 2; // Exponential backoff
      }
    }
  }

  throw lastError!;
}

// API client functions with enhanced retry and cancellation support
export async function getMerchant(
  merchantId: string,
  options?: {
    signal?: AbortSignal;
    retries?: number;
  }
): Promise<Merchant> {
  const cacheKey = `merchant:${merchantId}`;
  const maxRetries = options?.retries ?? 3;

  // Check cache first
  const cached = apiCache.get<Merchant>(cacheKey);
  if (cached) {
    if (process.env.NODE_ENV === "test") {
      console.log("[API] getMerchant: Returning cached data for", merchantId);
    }
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  if (process.env.NODE_ENV === "test") {
    console.log("[API] getMerchant: Making request for", merchantId);
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    // Use retryWithBackoff for automatic retries with exponential backoff
    return retryWithBackoff(
      async () => {
        try {
          if (process.env.NODE_ENV === "test") {
            console.log(
              "[API] getMerchant: Fetching",
              ApiEndpoints.merchants.get(merchantId)
            );
          }

          // Create AbortController if signal provided, or use provided signal
          const controller = options?.signal
            ? undefined
            : new AbortController();
          const signal = options?.signal || controller?.signal;

          const response = await safeFetch(
            ApiEndpoints.merchants.get(merchantId),
            {
              method: "GET",
              headers,
              signal, // Support request cancellation
            }
          );

          if (process.env.NODE_ENV === "test" && response) {
            console.log(
              "[API] getMerchant: Response received",
              response.status,
              response.ok
            );
          }

          // Get raw response data (may have snake_case fields)
          const rawData = await handleResponse<Record<string, unknown>>(
            response
          );

          if (process.env.NODE_ENV === "test") {
            console.log("[API] getMerchant: Data parsed successfully", rawData);
          }

          // Map backend fields to frontend types
          const getString = (key: string): string | undefined => {
            const value = rawData[key];
            return typeof value === "string" ? value : undefined;
          };

          const contactInfo =
            rawData.contact_info &&
            typeof rawData.contact_info === "object" &&
            rawData.contact_info !== null
              ? (rawData.contact_info as Record<string, unknown>)
              : undefined;

          const data: Merchant = {
            id: getString("id") || "",
            businessName:
              getString("business_name") ||
              getString("name") ||
              getString("businessName") ||
              "",
            name: getString("name"),
            legalName: getString("legal_name"),
            registrationNumber: getString("registration_number"),
            taxId: getString("tax_id"),
            industry: getString("industry"),
            industryCode: getString("industry_code"),
            businessType: getString("business_type"),
            description: getString("description"),
            status: getString("status") || "",
            website:
              getString("website") ||
              (contactInfo && typeof contactInfo.website === "string"
                ? contactInfo.website
                : undefined),
            email:
              getString("email") ||
              (contactInfo && typeof contactInfo.email === "string"
                ? contactInfo.email
                : undefined),
            phone:
              getString("phone") ||
              (contactInfo && typeof contactInfo.phone === "string"
                ? contactInfo.phone
                : undefined),
            // Map address - handle both nested map and flat fields
            address: mapAddress(rawData.address, rawData),
            portfolioType: getString("portfolio_type"),
            riskLevel: getString("risk_level"),
            complianceStatus: getString("compliance_status"),
            // Map financial information
            foundedDate:
              rawData.founded_date && typeof rawData.founded_date === "string"
                ? new Date(rawData.founded_date).toISOString()
                : undefined,
            employeeCount:
              typeof rawData.employee_count === "number"
                ? rawData.employee_count
                : undefined,
            annualRevenue:
              typeof rawData.annual_revenue === "number"
                ? rawData.annual_revenue
                : undefined,
            // Map system information
            createdBy: getString("created_by"),
            metadata:
              rawData.metadata &&
              typeof rawData.metadata === "object" &&
              rawData.metadata !== null &&
              !Array.isArray(rawData.metadata)
                ? (rawData.metadata as Record<string, unknown>)
                : undefined,
            createdAt: getString("created_at") || getString("createdAt") || "",
            updatedAt: getString("updated_at") || getString("updatedAt") || "",
          };

          // Validate with Zod schema
          const validatedData = validateAPIResponse(
            MerchantSchema,
            data,
            `getMerchant(${merchantId})`
          );

          // Development logging
          if (process.env.NODE_ENV === "development") {
            const hasFinancialData = !!(
              validatedData.foundedDate ||
              validatedData.employeeCount ||
              validatedData.annualRevenue
            );
            console.log("[API] Mapped merchant fields:", {
              id: validatedData.id,
              hasFinancialData,
              hasAddress: !!validatedData.address,
              hasMetadata: !!validatedData.metadata,
              hasCreatedBy: !!validatedData.createdBy,
            });

            if (
              !validatedData.foundedDate &&
              !validatedData.employeeCount &&
              !validatedData.annualRevenue
            ) {
              console.warn(
                "[API] Merchant missing financial data:",
                validatedData.id
              );
            }
          }

          // Cache the result
          apiCache.set(cacheKey, validatedData);
          return validatedData;
        } catch (error) {
          // Don't retry on abort (user cancellation)
          if (error instanceof Error && error.name === "AbortError") {
            throw error;
          }

          // Don't retry on 4xx errors (client errors)
          if (error instanceof Error && "status" in error) {
            const status = (error as Error & { status?: number }).status;
            if (status !== undefined && status >= 400 && status < 500) {
              throw error;
            }
          }

          if (process.env.NODE_ENV === "test") {
            console.error("[API] getMerchant: Error occurred", error);
          }

          // Re-throw to let retryWithBackoff handle it
          throw error;
        }
      },
      maxRetries,
      1000 // Initial delay of 1 second
    ).catch(async (error) => {
      // Final error handling after all retries exhausted
      await ErrorHandler.handleAPIError(error);
      throw error;
    });
  });
}

export async function getMerchantAnalytics(
  merchantId: string
): Promise<AnalyticsData> {
  const cacheKey = `analytics:${merchantId}`;

  // Check cache first
  const cached = apiCache.get<AnalyticsData>(cacheKey);
  if (cached) {
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      const response = await safeFetch(
        ApiEndpoints.merchants.analytics(merchantId),
        {
          method: "GET",
          headers,
        }
      );
      const rawData = await handleResponse<unknown>(response);
      // Validate with Zod schema
      const data = validateAPIResponse(
        AnalyticsDataSchema,
        rawData,
        `getMerchantAnalytics(${merchantId})`
      );
      // Cache the result
      apiCache.set(cacheKey, data);
      return data;
    } catch (error) {
      await ErrorHandler.handleAPIError(error);
      throw error;
    }
  });
}

export async function getWebsiteAnalysis(
  merchantId: string
): Promise<WebsiteAnalysisData> {
  const token = getAuthToken();
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return retryWithBackoff(async () => {
    const response = await safeFetch(
      ApiEndpoints.merchants.websiteAnalysis(merchantId),
      {
        method: "GET",
        headers,
      }
    );
    return handleResponse<WebsiteAnalysisData>(response);
  });
}

export interface MerchantAnalyticsStatus {
  merchantId: string;
  status: {
    classification: "pending" | "processing" | "completed" | "failed";
    websiteAnalysis: "pending" | "processing" | "completed" | "failed" | "skipped";
    classificationUpdatedAt?: string;
    websiteAnalysisUpdatedAt?: string;
  };
  timestamp: string;
}

export async function getMerchantAnalyticsStatus(
  merchantId: string
): Promise<MerchantAnalyticsStatus> {
  const token = getAuthToken();
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return retryWithBackoff(async () => {
    const response = await safeFetch(
      ApiEndpoints.merchants.analyticsStatus(merchantId),
      {
        method: "GET",
        headers,
      }
    );
    return handleResponse<MerchantAnalyticsStatus>(response);
  });
}

export async function getRiskAssessment(
  merchantId: string
): Promise<RiskAssessment | null> {
  const token = getAuthToken();
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  try {
    return await retryWithBackoff(async () => {
      // Add format=assessment query parameter to distinguish from getMerchantRiskScore
      const url = `${ApiEndpoints.merchants.riskScore(merchantId)}?format=assessment`;
      const response = await safeFetch(
        url,
        {
          method: "GET",
          headers,
        }
      );
      if (response.status === 404) {
        return null;
      }
      const rawData = await handleResponse<unknown>(response);
      // Validate with Zod schema
      return validateAPIResponse(
        RiskAssessmentSchema,
        rawData,
        `getRiskAssessment(${merchantId})`
      );
    });
  } catch (error) {
    console.error("Error fetching risk assessment:", error);
    return null;
  }
}

export async function startRiskAssessment(
  request: RiskAssessmentRequest
): Promise<RiskAssessmentResponse> {
  const token = getAuthToken();
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return retryWithBackoff(async () => {
    const response = await safeFetch(ApiEndpoints.risk.assess(), {
      method: "POST",
      headers,
      body: JSON.stringify(request),
    });
    return handleResponse<RiskAssessmentResponse>(response);
  });
}

export async function getAssessmentStatus(
  assessmentId: string
): Promise<AssessmentStatusResponse> {
  const token = getAuthToken();
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return retryWithBackoff(async () => {
    const response = await safeFetch(
      ApiEndpoints.risk.getAssessment(assessmentId),
      {
        method: "GET",
        headers,
      }
    );
    const rawData = await handleResponse<unknown>(response);
    // Validate with Zod schema
    return validateAPIResponse(
      AssessmentStatusResponseSchema,
      rawData,
      `getAssessmentStatus(${assessmentId})`
    );
  });
}

// Risk History
type RiskHistoryResponse = {
  merchantId: string;
  history: RiskAssessment[];
  limit: number;
  offset: number;
  total: number;
};

export async function getRiskHistory(
  merchantId: string,
  limit = 10,
  offset = 0
): Promise<RiskHistoryResponse> {
  const cacheKey = `risk-history:${merchantId}:${limit}:${offset}`;

  const cached = apiCache.get<RiskHistoryResponse>(cacheKey);
  if (cached) {
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      const response = await safeFetch(
        ApiEndpoints.risk.history(merchantId, limit, offset),
        {
          method: "GET",
          headers,
        }
      );

      // Handle 404 gracefully - endpoint may not be implemented yet
      if (response.status === 404) {
        // Return empty history instead of throwing
        return {
          merchantId,
          history: [],
          limit,
          offset,
          total: 0,
        };
      }

      const rawData = await handleResponse<unknown>(response);
      // Validate with Zod schema
      const data = validateAPIResponse(
        RiskHistoryResponseSchema,
        rawData,
        `getRiskHistory(${merchantId}, ${limit}, ${offset})`
      );
      apiCache.set(cacheKey, data);
      return data;
    } catch (error) {
      // Don't show error notifications for 404s on optional endpoints
      const is404 = error instanceof Error && error.message.includes("404");
      if (!is404) {
        await ErrorHandler.handleAPIError(error);
      }
      throw error;
    }
  });
}

// Risk Predictions
type RiskPredictionsResponse = {
  merchantId: string;
  horizons: number[];
  predictions: Array<{
    horizon: number;
    months: number;
    predictedScore?: number;
    riskLevel?: string;
    confidence?: number;
    scenarios?: Array<{ name: string; probability: number; score: number }>;
  }>;
  includeScenarios: boolean;
  includeConfidence: boolean;
};

export async function getRiskPredictions(
  merchantId: string,
  horizons: number[] = [3, 6, 12],
  includeScenarios = false,
  includeConfidence = false
): Promise<RiskPredictionsResponse> {
  const cacheKey = `risk-predictions:${merchantId}:${horizons.join(
    ","
  )}:${includeScenarios}:${includeConfidence}`;

  const cached = apiCache.get<RiskPredictionsResponse>(cacheKey);
  if (cached) {
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      // Use first horizon for endpoint, but we'll pass all params via query string if needed
      const horizon = horizons.length > 0 ? String(horizons[0]) : undefined;
      let url = ApiEndpoints.risk.predictions(merchantId, horizon);

      // Add additional params if needed (horizons array, includeScenarios, includeConfidence)
      const params = new URLSearchParams();
      if (horizons.length > 1) params.append("horizons", horizons.join(","));
      if (includeScenarios)
        params.append("includeScenarios", String(includeScenarios));
      if (includeConfidence)
        params.append("includeConfidence", String(includeConfidence));
      const queryString = params.toString();
      if (queryString)
        url = `${url}${url.includes("?") ? "&" : "?"}${queryString}`;

      const response = await safeFetch(url, {
        method: "GET",
        headers,
      });
      const data = await handleResponse<RiskPredictionsResponse>(response);
      apiCache.set(cacheKey, data);
      return data;
    } catch (error) {
      await ErrorHandler.handleAPIError(error);
      throw error;
    }
  });
}

// Explain Risk Assessment
type RiskExplanationResponse = {
  assessmentId: string;
  factors: Array<{ name: string; score: number; weight: number }>;
  shapValues: Record<string, number>;
  baseValue: number;
  prediction: number;
};

export async function explainRiskAssessment(
  assessmentId: string
): Promise<RiskExplanationResponse> {
  const cacheKey = `risk-explain:${assessmentId}`;

  const cached = apiCache.get<RiskExplanationResponse>(cacheKey);
  if (cached) {
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      const response = await safeFetch(
        ApiEndpoints.risk.explain(assessmentId),
        {
          method: "GET",
          headers,
        }
      );
      const data = await handleResponse<RiskExplanationResponse>(response);
      apiCache.set(cacheKey, data);
      return data;
    } catch (error) {
      await ErrorHandler.handleAPIError(error);
      throw error;
    }
  });
}

// Risk Recommendations
type RiskRecommendationsResponse = {
  merchantId: string;
  recommendations: Array<{
    id: string;
    type: string;
    priority: string;
    title: string;
    description: string;
    actionItems: string[];
  }>;
  timestamp: string;
};

export async function getRiskRecommendations(
  merchantId: string
): Promise<RiskRecommendationsResponse> {
  const cacheKey = `risk-recommendations:${merchantId}`;

  const cached = apiCache.get<RiskRecommendationsResponse>(cacheKey);
  if (cached) {
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      const response = await safeFetch(
        ApiEndpoints.merchants.riskRecommendations(merchantId),
        {
          method: "GET",
          headers,
        }
      );
      const rawData = await handleResponse<unknown>(response);
      // Validate with Zod schema
      const data = validateAPIResponse(
        RiskRecommendationsResponseSchema,
        rawData,
        `getRiskRecommendations(${merchantId})`
      );
      apiCache.set(cacheKey, data);
      return data;
    } catch (error) {
      await ErrorHandler.handleAPIError(error);
      throw error;
    }
  });
}

// Risk Indicators
export async function getRiskIndicators(
  merchantId: string,
  severity?: string,
  status?: string
): Promise<RiskIndicatorsData> {
  const cacheKey = `risk-indicators:${merchantId}:${severity || ""}:${
    status || ""
  }`;

  const cached = apiCache.get<RiskIndicatorsData>(cacheKey);
  if (cached) {
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      const filters: Record<string, string> = {};
      if (severity) filters.severity = severity;
      if (status) filters.status = status;

      const response = await safeFetch(
        ApiEndpoints.risk.indicators(merchantId, filters),
        {
          method: "GET",
          headers,
        }
      );

      // Handle 404 gracefully - endpoint may not be implemented yet
      if (response.status === 404) {
        // Return empty indicators instead of throwing
        return {
          merchantId,
          indicators: [],
        };
      }

      const rawData = await handleResponse<unknown>(response);
      // Validate with Zod schema
      const data = validateAPIResponse(
        RiskIndicatorsDataSchema,
        rawData,
        `getRiskIndicators(${merchantId}, ${severity || ""}, ${status || ""})`
      );
      apiCache.set(cacheKey, data);
      return data;
    } catch (error) {
      // Don't show error notifications for 404s on optional endpoints
      const is404 = error instanceof Error && error.message.includes("404");
      if (!is404) {
        await ErrorHandler.handleAPIError(error);
      }
      throw error;
    }
  });
}

// Risk Alerts (active risk indicators)
export async function getRiskAlerts(
  merchantId: string,
  severity?: string
): Promise<RiskIndicatorsData> {
  // Alerts are active indicators, so use getRiskIndicators with status="active"
  return getRiskIndicators(merchantId, severity, "active");
}

// Enrichment Sources
export async function getEnrichmentSources(merchantId: string): Promise<{
  merchantId: string;
  sources: EnrichmentSource[];
}> {
  const cacheKey = `enrichment-sources:${merchantId}`;

  const cached = apiCache.get<{
    merchantId: string;
    sources: EnrichmentSource[];
  }>(cacheKey);
  if (cached) {
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      const response = await safeFetch(
        ApiEndpoints.merchants.enrichmentSources(merchantId),
        {
          method: "GET",
          headers,
        }
      );

      // Handle 404 gracefully - endpoint may not be implemented yet
      if (response.status === 404) {
        // Return empty sources instead of throwing
        return {
          merchantId,
          sources: [],
        };
      }

      const data = await handleResponse<{
        merchantId: string;
        sources: EnrichmentSource[];
      }>(response);
      apiCache.set(cacheKey, data);
      return data;
    } catch (error) {
      // Don't show error notifications for 404s on optional endpoints
      const is404 = error instanceof Error && error.message.includes("404");
      if (!is404) {
        await ErrorHandler.handleAPIError(error);
      }
      throw error;
    }
  });
}

// Trigger Enrichment
type EnrichmentJobResponse = {
  jobId: string;
  merchantId: string;
  source: string;
  status: string;
  createdAt: string;
};

export async function triggerEnrichment(
  merchantId: string,
  source: string
): Promise<EnrichmentJobResponse> {
  const token = getAuthToken();
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return retryWithBackoff(async () => {
    const response = await safeFetch(
      ApiEndpoints.merchants.triggerEnrichment(merchantId),
      {
        method: "POST",
        headers,
        body: JSON.stringify({ source }),
      }
    );
    return handleResponse<EnrichmentJobResponse>(response);
  });
}

// Merchant List
export async function getMerchantsList(
  params?: MerchantListParams
): Promise<MerchantListResponse> {
  const cacheKey = `merchants-list:${JSON.stringify(params || {})}`;

  // Check cache first (shorter TTL for list data)
  const cached = apiCache.get<MerchantListResponse>(cacheKey);
  if (cached) {
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      const queryParams = new URLSearchParams();
      if (params?.page) queryParams.append("page", params.page.toString());
      if (params?.pageSize)
        queryParams.append("page_size", params.pageSize.toString());
      if (params?.portfolioType)
        queryParams.append("portfolio_type", params.portfolioType);
      if (params?.riskLevel) queryParams.append("risk_level", params.riskLevel);
      if (params?.status) queryParams.append("status", params.status);
      if (params?.search) queryParams.append("search", params.search);
      if (params?.sortBy) queryParams.append("sort_by", params.sortBy);
      if (params?.sortOrder) queryParams.append("sort_order", params.sortOrder);

      const queryString = queryParams.toString();
      const url = `${ApiEndpoints.merchants.list()}${
        queryString ? `?${queryString}` : ""
      }`;

      const response = await safeFetch(url, {
        method: "GET",
        headers,
      });

      const rawData = await handleResponse<unknown>(response);
      // Validate with Zod schema (validates snake_case from backend)
      // Note: MerchantListItem interface uses snake_case (legal_name, created_at, updated_at, etc.)
      // which matches the backend response. This is different from getMerchant() which maps to camelCase.
      // For full consistency, we could map these fields to camelCase here, but that would require
      // updating the MerchantListItem interface and all usages. Current implementation is validated and working.
      const data = validateAPIResponse(
        MerchantListResponseSchema,
        rawData,
        `getMerchantsList(${JSON.stringify(params || {})})`
      );

      // Cache with shorter TTL for list data (1 minute)
      apiCache.set(cacheKey, data, 60 * 1000);

      return data;
    } catch (error) {
      await ErrorHandler.handleAPIError(error);
      throw error;
    }
  });
}

// Dashboard Metrics
export async function getDashboardMetrics(): Promise<DashboardMetrics> {
  const cacheKey = "dashboard-metrics";

  const cached = apiCache.get<DashboardMetrics>(cacheKey);
  if (cached) {
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      // Use v3 endpoint (v1 deprecated and removed)
      const response = await safeFetch(ApiEndpoints.dashboard.metrics("v3"), {
        method: "GET",
        headers,
      });

      if (!response.ok) {
        // Return default values if endpoint doesn't exist
        return {
          totalMerchants: 0,
          revenue: 0,
          growthRate: 0,
          analyticsScore: 0,
        };
      }

      const data = await handleResponse<{
        data?: DashboardMetrics;
        business?:
          | Record<string, unknown>
          | {
              active_users?: number;
              total_verifications?: number;
              revenue?: number;
              growth_rate?: number;
            };
        overview?: {
          total_requests?: number;
          active_users?: number;
          success_rate?: number;
          average_response_time?: number;
        };
        performance?: {
          response_time?: number;
          throughput?: number;
          error_rate?: number;
        };
      }>(response);

      // Handle different response formats
      // v3 comprehensive format: { overview: {...}, performance: {...}, business: {...} }
      // v1 basic format: { data: {...} } or { business: {...} }
      let metrics: DashboardMetrics;

      if (data.data) {
        // Direct data format
        metrics = data.data;
      } else if (data.overview || data.performance || data.business) {
        // v3 comprehensive format - map to DashboardMetrics
        const businessData = data.business as
          | {
              total_verifications?: number;
              revenue?: number;
              growth_rate?: number;
              analytics_score?: number;
            }
          | undefined;
        metrics = {
          totalMerchants:
            businessData?.total_verifications ||
            data.overview?.total_requests ||
            0,
          revenue: businessData?.revenue || 0,
          growthRate: businessData?.growth_rate || 0,
          analyticsScore:
            data.performance?.response_time ||
            data.overview?.average_response_time ||
            0,
        };
      } else {
        // Fallback to business object
        const businessData = data.business as
          | {
              total_verifications?: number;
              revenue?: number;
              growth_rate?: number;
              analytics_score?: number;
            }
          | undefined;
        metrics = {
          totalMerchants: businessData?.total_verifications || 0,
          revenue: businessData?.revenue || 0,
          growthRate: businessData?.growth_rate || 0,
          analyticsScore: businessData?.analytics_score || 0,
        };
      }

      // Validate with Zod schema
      const validatedMetrics = validateAPIResponse(
        DashboardMetricsSchema,
        metrics,
        "getDashboardMetrics()"
      );

      apiCache.set(cacheKey, validatedMetrics, 60 * 1000); // 1 minute cache
      return validatedMetrics;
    } catch {
      // Return default values on error
      return {
        totalMerchants: 0,
        revenue: 0,
        growthRate: 0,
        analyticsScore: 0,
      };
    }
  });
}

// Risk Metrics
export async function getRiskMetrics(): Promise<RiskMetrics> {
  const cacheKey = "risk-metrics";

  const cached = apiCache.get<RiskMetrics>(cacheKey);
  if (cached) {
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      const response = await safeFetch(ApiEndpoints.risk.metrics(), {
        method: "GET",
        headers,
      });

      if (!response.ok) {
        // Return default values if endpoint doesn't exist
        return {
          overallRiskScore: 0,
          highRiskMerchants: 0,
          riskAssessments: 0,
          riskTrend: 0,
        };
      }

      const rawData = await handleResponse<unknown>(response);
      // Validate with Zod schema
      const data = validateAPIResponse(
        RiskMetricsSchema,
        rawData,
        "getRiskMetrics()"
      );
      apiCache.set(cacheKey, data, 60 * 1000); // 1 minute cache
      return data;
    } catch {
      // Return default values on error
      return {
        overallRiskScore: 0,
        highRiskMerchants: 0,
        riskAssessments: 0,
        riskTrend: 0,
      };
    }
  });
}

// System Metrics
export async function getSystemMetrics(): Promise<SystemMetrics> {
  const cacheKey = "system-metrics";

  const cached = apiCache.get<SystemMetrics>(cacheKey);
  if (cached) {
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      // Try multiple possible endpoints
      let response = await safeFetch(ApiEndpoints.monitoring.metrics(), {
        method: "GET",
        headers,
      });

      if (!response.ok && response.status === 404) {
        response = await safeFetch(ApiEndpoints.monitoring.systemMetrics(), {
          method: "GET",
          headers,
        });
      }

      if (!response.ok && response.status === 404) {
        response = await safeFetch(ApiEndpoints.monitoring.generalMetrics(), {
          method: "GET",
          headers,
        });
      }

      if (!response.ok) {
        // Return default healthy values
        return {
          systemHealth: 100,
          serverStatus: "Online",
          databaseStatus: "Connected",
          responseTime: 0,
        };
      }

      const rawData = await handleResponse<unknown>(response);

      // Handle different response formats
      const data =
        typeof rawData === "object" &&
        rawData !== null &&
        "data" in rawData &&
        rawData.data
          ? rawData.data
          : rawData;

      // Validate with Zod schema
      const metrics = validateAPIResponse(
        SystemMetricsSchema,
        data,
        "getSystemMetrics()"
      );

      apiCache.set(cacheKey, metrics, 30 * 1000); // 30 second cache for system metrics
      return metrics;
    } catch {
      // Return default healthy values on error
      return {
        systemHealth: 100,
        serverStatus: "Online",
        databaseStatus: "Connected",
        responseTime: 0,
      };
    }
  });
}

// Compliance Status
export async function getComplianceStatus(): Promise<ComplianceStatus> {
  const cacheKey = "compliance-status";

  const cached = apiCache.get<ComplianceStatus>(cacheKey);
  if (cached) {
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      const response = await safeFetch(ApiEndpoints.compliance.status(), {
        method: "GET",
        headers,
      });

      if (!response.ok) {
        // Return default values if endpoint doesn't exist
        return {
          overallScore: 0,
          pendingReviews: 0,
          complianceTrend: "Stable",
          regulatoryFrameworks: 0,
        };
      }

      const data = await handleResponse<
        | ComplianceStatus
        | {
            overall_status?: string;
            compliance_score?: number;
            overall_score?: number;
            frameworks?: Array<{
              framework_id: string;
              framework_name: string;
              status: string;
              score: number;
            }>;
            requirements?: Array<{
              requirement_id: string;
              status: string;
            }>;
            alerts?: Array<{
              id: string;
              severity: string;
              status: string;
            }>;
          }
      >(response);

      // Handle different response formats
      let status: ComplianceStatus;

      if (
        "overallScore" in data ||
        "compliance_score" in data ||
        "overall_score" in data
      ) {
        // Enhanced format with comprehensive data
        type EnhancedComplianceData = {
          compliance_score?: number;
          overall_score?: number;
          frameworks?: Array<{
            framework_id: string;
            framework_name: string;
            status: string;
            score: number;
          }>;
          requirements?: Array<{ requirement_id: string; status: string }>;
          alerts?: Array<{ id: string; severity: string; status: string }>;
        };
        const enhancedData = data as EnhancedComplianceData;

        const frameworks = Array.isArray(enhancedData.frameworks)
          ? enhancedData.frameworks
          : [];
        const requirements = Array.isArray(enhancedData.requirements)
          ? enhancedData.requirements
          : [];
        const alerts = Array.isArray(enhancedData.alerts)
          ? enhancedData.alerts
          : [];

        // Calculate pending reviews from requirements with pending status
        const pendingReviews = requirements.filter(
          (req: { status: string }) =>
            req.status === "pending" || req.status === "in_progress"
        ).length;

        // Determine compliance trend from framework scores
        let complianceTrend: "Improving" | "Stable" | "Declining" = "Stable";
        if (frameworks.length > 0) {
          const avgScore =
            frameworks.reduce(
              (sum: number, f: { score: number }) => sum + f.score,
              0
            ) / frameworks.length;
          if (avgScore >= 0.9) {
            complianceTrend = "Improving";
          } else if (avgScore < 0.7) {
            complianceTrend = "Declining";
          }
        }

        status = {
          overallScore:
            enhancedData.compliance_score ||
            enhancedData.overall_score ||
            (data as ComplianceStatus).overallScore ||
            0,
          pendingReviews:
            pendingReviews || (data as ComplianceStatus).pendingReviews || 0,
          complianceTrend:
            complianceTrend ||
            (data as ComplianceStatus).complianceTrend ||
            "Stable",
          regulatoryFrameworks:
            frameworks.length ||
            (data as ComplianceStatus).regulatoryFrameworks ||
            0,
          violations: alerts.filter(
            (a: { severity: string }) =>
              a.severity === "high" || a.severity === "critical"
          ).length,
        };
      } else {
        // Direct ComplianceStatus format
        status = data as ComplianceStatus;
      }

      // Validate with Zod schema
      const validatedStatus = validateAPIResponse(
        ComplianceStatusSchema,
        status,
        "getComplianceStatus()"
      );

      apiCache.set(cacheKey, validatedStatus, 5 * 60 * 1000); // 5 minute cache
      return validatedStatus;
    } catch {
      // Return default values on error
      return {
        overallScore: 0,
        pendingReviews: 0,
        complianceTrend: "Stable",
        regulatoryFrameworks: 0,
      };
    }
  });
}

// Create Merchant
export interface CreateMerchantRequest {
  name: string;
  legal_name?: string;
  website?: string;
  address?: {
    street?: string;
    city?: string;
    state?: string;
    postal_code?: string;
    country?: string;
  };
  contact_info?: {
    phone?: string;
    email?: string;
  };
  registration_number?: string;
  tax_id?: string;
  industry?: string;
  country: string;
  analysis_type?: string;
  assessment_type?: string;
}

export interface CreateMerchantResponse {
  id: string;
  name: string;
  status: string;
  created_at: string;
}

export async function createMerchant(
  data: CreateMerchantRequest
): Promise<CreateMerchantResponse> {
  const token = getAuthToken();
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return retryWithBackoff(async () => {
    try {
      const response = await safeFetch(ApiEndpoints.merchants.create(), {
        method: "POST",
        headers,
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        const errorData = await ErrorHandler.parseErrorResponse(response);
        const errorMessage =
          errorData && typeof errorData === "object" && "message" in errorData
            ? String(errorData.message)
            : `Failed to create merchant: ${response.status}`;
        throw new Error(errorMessage);
      }

      const result = await handleResponse<CreateMerchantResponse>(response);

      // Clear cache for merchant list
      apiCache.clear();

      return result;
    } catch (error) {
      await ErrorHandler.handleAPIError(error);
      throw error;
    }
  });
}

// Business Intelligence Metrics
export async function getBusinessIntelligenceMetrics(): Promise<BusinessIntelligenceMetrics> {
  const cacheKey = "business-intelligence-metrics";

  const cached = apiCache.get<BusinessIntelligenceMetrics>(cacheKey);
  if (cached) {
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      const response = await safeFetch(
        ApiEndpoints.businessIntelligence.metrics(),
        {
          method: "GET",
          headers,
        }
      );

      if (!response.ok) {
        // Return default values if endpoint doesn't exist
        return {
          revenueGrowth: 0,
          marketShare: 0,
          performanceScore: 0,
          analyticsScore: 0,
        };
      }

      const data = await handleResponse<BusinessIntelligenceMetrics>(response);
      apiCache.set(cacheKey, data, 60 * 1000); // 1 minute cache
      return data;
    } catch {
      // Return default values on error
      return {
        revenueGrowth: 0,
        marketShare: 0,
        performanceScore: 0,
        analyticsScore: 0,
      };
    }
  });
}

// Portfolio-level analytics functions

/**
 * Get portfolio-wide analytics (all merchants)
 * Uses caching with 5-10 minute TTL
 */
export async function getPortfolioAnalytics(): Promise<PortfolioAnalytics> {
  const cacheKey = "portfolio-analytics";

  // Check cache first
  const cached = apiCache.get<PortfolioAnalytics>(cacheKey);
  if (cached) {
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      const response = await safeFetch(
        ApiEndpoints.merchants.portfolioAnalytics(),
        {
          method: "GET",
          headers,
        }
      );
      const data = await handleResponse<PortfolioAnalytics>(response);
      // Cache for 5-10 minutes (using 7 minutes as middle ground)
      apiCache.set(cacheKey, data, 7 * 60 * 1000);
      return data;
    } catch (error) {
      await ErrorHandler.handleAPIError(error);
      throw error;
    }
  });
}

/**
 * Get portfolio-wide statistics
 * Uses caching with 5-10 minute TTL
 */
export async function getPortfolioStatistics(): Promise<PortfolioStatistics> {
  const cacheKey = "portfolio-statistics";

  // Check cache first
  const cached = apiCache.get<PortfolioStatistics>(cacheKey);
  if (cached) {
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      const response = await safeFetch(ApiEndpoints.merchants.statistics(), {
        method: "GET",
        headers,
      });
      const rawData = await handleResponse<unknown>(response);
      // Validate with Zod schema
      const data = validateAPIResponse(
        PortfolioStatisticsSchema,
        rawData,
        "getPortfolioStatistics()"
      );
      // Cache for 5-10 minutes (using 7 minutes as middle ground)
      apiCache.set(cacheKey, data, 7 * 60 * 1000);
      return data;
    } catch (error) {
      await ErrorHandler.handleAPIError(error);
      throw error;
    }
  });
}

/**
 * Get risk trends analytics
 * Uses caching with deduplication
 */
export async function getRiskTrends(params?: {
  industry?: string;
  country?: string;
  timeframe?: string;
  limit?: number;
}): Promise<RiskTrends> {
  const cacheKey = `risk-trends:${JSON.stringify(params || {})}`;

  // Check cache first
  const cached = apiCache.get<RiskTrends>(cacheKey);
  if (cached) {
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      const response = await safeFetch(ApiEndpoints.analytics.trends(params), {
        method: "GET",
        headers,
      });
      const data = await handleResponse<RiskTrends>(response);
      // Cache for 5 minutes
      apiCache.set(cacheKey, data, 5 * 60 * 1000);
      return data;
    } catch (error) {
      await ErrorHandler.handleAPIError(error);
      throw error;
    }
  });
}

/**
 * Get risk insights analytics
 * Uses caching with deduplication
 */
export async function getRiskInsights(params?: {
  industry?: string;
  country?: string;
  risk_level?: string;
}): Promise<RiskInsights> {
  const cacheKey = `risk-insights:${JSON.stringify(params || {})}`;

  // Check cache first
  const cached = apiCache.get<RiskInsights>(cacheKey);
  if (cached) {
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      const response = await safeFetch(
        ApiEndpoints.analytics.insights(params),
        {
          method: "GET",
          headers,
        }
      );
      const data = await handleResponse<RiskInsights>(response);
      // Cache for 5 minutes
      apiCache.set(cacheKey, data, 5 * 60 * 1000);
      return data;
    } catch (error) {
      await ErrorHandler.handleAPIError(error);
      throw error;
    }
  });
}

/**
 * Get risk benchmarks for an industry
 * Uses caching with 10-15 minute TTL
 */
export async function getRiskBenchmarks(params: {
  mcc?: string;
  naics?: string;
  sic?: string;
}): Promise<RiskBenchmarks> {
  const cacheKey = `risk-benchmarks:${
    params.mcc || params.naics || params.sic
  }`;

  // Check cache first
  const cached = apiCache.get<RiskBenchmarks>(cacheKey);
  if (cached) {
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      const response = await safeFetch(ApiEndpoints.risk.benchmarks(params), {
        method: "GET",
        headers,
      });
      const rawData = await handleResponse<unknown>(response);
      // Validate with Zod schema
      const data = validateAPIResponse(
        RiskBenchmarksSchema,
        rawData,
        `getRiskBenchmarks(${JSON.stringify(params)})`
      );
      // Cache for 10-15 minutes (using 12 minutes as middle ground)
      apiCache.set(cacheKey, data, 12 * 60 * 1000);
      return data;
    } catch (error) {
      await ErrorHandler.handleAPIError(error);
      throw error;
    }
  });
}

/**
 * Get merchant risk score
 * Uses caching with 2-5 minute TTL
 */
export async function getMerchantRiskScore(
  merchantId: string
): Promise<MerchantRiskScore> {
  const cacheKey = `merchant-risk-score:${merchantId}`;

  // Check cache first
  const cached = apiCache.get<MerchantRiskScore>(cacheKey);
  if (cached) {
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      const response = await safeFetch(
        ApiEndpoints.merchants.riskScore(merchantId),
        {
          method: "GET",
          headers,
        }
      );
      const rawData = await handleResponse<unknown>(response);
      // Validate with Zod schema
      const data = validateAPIResponse(
        MerchantRiskScoreSchema,
        rawData,
        `getMerchantRiskScore(${merchantId})`
      );
      // Cache for 2-5 minutes (using 3 minutes as middle ground)
      apiCache.set(cacheKey, data, 3 * 60 * 1000);
      return data;
    } catch (error) {
      await ErrorHandler.handleAPIError(error);
      throw error;
    }
  });
}
