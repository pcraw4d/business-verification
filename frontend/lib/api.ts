// API client for Merchant Details
import { APICache } from '@/lib/api-cache';
import { ApiEndpoints } from '@/lib/api-config';
import { ErrorHandler } from '@/lib/error-handler';
import { RequestDeduplicator } from '@/lib/request-deduplicator';
import type {
  BusinessIntelligenceMetrics,
  ComplianceStatus,
  DashboardMetrics,
  RiskMetrics,
  SystemMetrics,
} from '@/types/dashboard';
import type {
  AnalyticsData,
  AssessmentStatusResponse,
  EnrichmentSource,
  Merchant,
  MerchantListParams,
  MerchantListResponse,
  RiskAssessment,
  RiskAssessmentRequest,
  RiskAssessmentResponse,
  RiskIndicatorsData,
  WebsiteAnalysisData,
} from '@/types/merchant';

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
  if (typeof window === 'undefined') return null;
  return sessionStorage.getItem('authToken');
}

// Helper function to handle API errors
async function handleResponse<T>(response: Response): Promise<T> {
  // Check response status - MSW responses should have status and ok set correctly
  const status = response.status;
  const isOk = response.ok === true || (status >= 200 && status < 300);
  
  // Debug logging in test environment
  if (process.env.NODE_ENV === 'test' && !isOk) {
    console.log('[API Debug] Error response - status:', status, 'ok:', response.ok, 'statusText:', response.statusText);
  }
  
  if (!isOk) {
    try {
      const errorData = await ErrorHandler.parseErrorResponse(response);
      const errorMessage = errorData && typeof errorData === 'object' && 'message' in errorData 
        ? String(errorData.message) 
        : `API Error ${status}`;
      throw new Error(errorMessage);
    } catch {
      // If parsing error response fails, just throw with status
      throw new Error(`API Error ${status}`);
    }
  }
  
  // Safely parse JSON response
  try {
    const json = await response.json();
    return json as T;
  } catch (error) {
    // If JSON parsing fails, throw a more helpful error
    const errorMessage = error instanceof Error ? error.message : String(error);
    throw new Error(`Failed to parse JSON response: ${errorMessage}`);
  }
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

// API client functions
export async function getMerchant(merchantId: string): Promise<Merchant> {
  const cacheKey = `merchant:${merchantId}`;
  
  // Check cache first
  const cached = apiCache.get<Merchant>(cacheKey);
  if (cached) {
    if (process.env.NODE_ENV === 'test') {
      console.log('[API] getMerchant: Returning cached data for', merchantId);
    }
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  if (process.env.NODE_ENV === 'test') {
    console.log('[API] getMerchant: Making request for', merchantId);
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      if (process.env.NODE_ENV === 'test') {
        console.log('[API] getMerchant: Fetching', ApiEndpoints.merchants.get(merchantId));
      }
      const response = await fetch(ApiEndpoints.merchants.get(merchantId), {
        method: 'GET',
        headers,
      });
      if (process.env.NODE_ENV === 'test') {
        console.log('[API] getMerchant: Response received', response.status, response.ok);
      }
      const data = await handleResponse<Merchant>(response);
      if (process.env.NODE_ENV === 'test') {
        console.log('[API] getMerchant: Data parsed successfully', data);
      }
      // Cache the result
      apiCache.set(cacheKey, data);
      return data;
    } catch (error) {
      if (process.env.NODE_ENV === 'test') {
        console.error('[API] getMerchant: Error occurred', error);
      }
      await ErrorHandler.handleAPIError(error);
      throw error;
    }
  });
}

export async function getMerchantAnalytics(merchantId: string): Promise<AnalyticsData> {
  const cacheKey = `analytics:${merchantId}`;
  
  // Check cache first
  const cached = apiCache.get<AnalyticsData>(cacheKey);
  if (cached) {
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      const response = await fetch(ApiEndpoints.merchants.analytics(merchantId), {
        method: 'GET',
        headers,
      });
      const data = await handleResponse<AnalyticsData>(response);
      // Cache the result
      apiCache.set(cacheKey, data);
      return data;
    } catch (error) {
      await ErrorHandler.handleAPIError(error);
      throw error;
    }
  });
}

export async function getWebsiteAnalysis(merchantId: string): Promise<WebsiteAnalysisData> {
  const token = getAuthToken();
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  return retryWithBackoff(async () => {
    const response = await fetch(ApiEndpoints.merchants.websiteAnalysis(merchantId), {
      method: 'GET',
      headers,
    });
    return handleResponse<WebsiteAnalysisData>(response);
  });
}

export async function getRiskAssessment(merchantId: string): Promise<RiskAssessment | null> {
  const token = getAuthToken();
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  try {
    return await retryWithBackoff(async () => {
      const response = await fetch(ApiEndpoints.merchants.riskScore(merchantId), {
        method: 'GET',
        headers,
      });
      if (response.status === 404) {
        return null;
      }
      return handleResponse<RiskAssessment>(response);
    });
  } catch (error) {
    console.error('Error fetching risk assessment:', error);
    return null;
  }
}

export async function startRiskAssessment(
  request: RiskAssessmentRequest
): Promise<RiskAssessmentResponse> {
  const token = getAuthToken();
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  return retryWithBackoff(async () => {
    const response = await fetch(ApiEndpoints.risk.assess(), {
      method: 'POST',
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
    'Content-Type': 'application/json',
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  return retryWithBackoff(async () => {
    const response = await fetch(ApiEndpoints.risk.getAssessment(assessmentId), {
      method: 'GET',
      headers,
    });
    return handleResponse<AssessmentStatusResponse>(response);
  });
}

// Risk History
type RiskHistoryResponse = { merchantId: string; history: RiskAssessment[]; limit: number; offset: number; total: number };

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
    'Content-Type': 'application/json',
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      const response = await fetch(
        ApiEndpoints.risk.history(merchantId, limit, offset),
        {
          method: 'GET',
          headers,
        }
      );
      const data = await handleResponse<RiskHistoryResponse>(response);
      apiCache.set(cacheKey, data);
      return data;
    } catch (error) {
      await ErrorHandler.handleAPIError(error);
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
  const cacheKey = `risk-predictions:${merchantId}:${horizons.join(',')}:${includeScenarios}:${includeConfidence}`;
  
  const cached = apiCache.get<RiskPredictionsResponse>(cacheKey);
  if (cached) {
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      // Use first horizon for endpoint, but we'll pass all params via query string if needed
      const horizon = horizons.length > 0 ? String(horizons[0]) : undefined;
      let url = ApiEndpoints.risk.predictions(merchantId, horizon);
      
      // Add additional params if needed (horizons array, includeScenarios, includeConfidence)
      const params = new URLSearchParams();
      if (horizons.length > 1) params.append('horizons', horizons.join(','));
      if (includeScenarios) params.append('includeScenarios', String(includeScenarios));
      if (includeConfidence) params.append('includeConfidence', String(includeConfidence));
      const queryString = params.toString();
      if (queryString) url = `${url}${url.includes('?') ? '&' : '?'}${queryString}`;
      
      const response = await fetch(url, {
          method: 'GET',
          headers,
        }
      );
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

export async function explainRiskAssessment(assessmentId: string): Promise<RiskExplanationResponse> {
  const cacheKey = `risk-explain:${assessmentId}`;
  
  const cached = apiCache.get<RiskExplanationResponse>(cacheKey);
  if (cached) {
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      const response = await fetch(ApiEndpoints.risk.explain(assessmentId), {
        method: 'GET',
        headers,
      });
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

export async function getRiskRecommendations(merchantId: string): Promise<RiskRecommendationsResponse> {
  const cacheKey = `risk-recommendations:${merchantId}`;
  
  const cached = apiCache.get<RiskRecommendationsResponse>(cacheKey);
  if (cached) {
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      const response = await fetch(
        ApiEndpoints.merchants.riskRecommendations(merchantId),
        {
          method: 'GET',
          headers,
        }
      );
      const data = await handleResponse<RiskRecommendationsResponse>(response);
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
  const cacheKey = `risk-indicators:${merchantId}:${severity || ''}:${status || ''}`;
  
  const cached = apiCache.get<RiskIndicatorsData>(cacheKey);
  if (cached) {
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      const filters: Record<string, string> = {};
      if (severity) filters.severity = severity;
      if (status) filters.status = status;
      
      const response = await fetch(
        ApiEndpoints.risk.indicators(merchantId, filters),
        {
          method: 'GET',
          headers,
        }
      );
      const data = await handleResponse<RiskIndicatorsData>(response);
      apiCache.set(cacheKey, data);
      return data;
    } catch (error) {
      await ErrorHandler.handleAPIError(error);
      throw error;
    }
  });
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
    'Content-Type': 'application/json',
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      const response = await fetch(
        ApiEndpoints.merchants.enrichmentSources(merchantId),
        {
          method: 'GET',
          headers,
        }
      );
      const data = await handleResponse<{
        merchantId: string;
        sources: EnrichmentSource[];
      }>(response);
      apiCache.set(cacheKey, data);
      return data;
    } catch (error) {
      await ErrorHandler.handleAPIError(error);
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
    'Content-Type': 'application/json',
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  return retryWithBackoff(async () => {
    const response = await fetch(
      ApiEndpoints.merchants.triggerEnrichment(merchantId),
      {
        method: 'POST',
        headers,
        body: JSON.stringify({ source }),
      }
    );
    return handleResponse<EnrichmentJobResponse>(response);
  });
}

// Merchant List
export async function getMerchantsList(params?: MerchantListParams): Promise<MerchantListResponse> {
  const cacheKey = `merchants-list:${JSON.stringify(params || {})}`;
  
  // Check cache first (shorter TTL for list data)
  const cached = apiCache.get<MerchantListResponse>(cacheKey);
  if (cached) {
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      const queryParams = new URLSearchParams();
      if (params?.page) queryParams.append('page', params.page.toString());
      if (params?.pageSize) queryParams.append('page_size', params.pageSize.toString());
      if (params?.portfolioType) queryParams.append('portfolio_type', params.portfolioType);
      if (params?.riskLevel) queryParams.append('risk_level', params.riskLevel);
      if (params?.status) queryParams.append('status', params.status);
      if (params?.search) queryParams.append('search', params.search);
      if (params?.sortBy) queryParams.append('sort_by', params.sortBy);
      if (params?.sortOrder) queryParams.append('sort_order', params.sortOrder);

      const queryString = queryParams.toString();
      const url = `${ApiEndpoints.merchants.list()}${queryString ? `?${queryString}` : ''}`;
      
      const response = await fetch(url, {
        method: 'GET',
        headers,
      });
      
      const data = await handleResponse<MerchantListResponse>(response);
      
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
  const cacheKey = 'dashboard-metrics';
  
  const cached = apiCache.get<DashboardMetrics>(cacheKey);
  if (cached) {
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      // Try v3 endpoint first, fallback to v1
      let response = await fetch(ApiEndpoints.dashboard.metrics('v3'), {
        method: 'GET',
        headers,
      });

      if (!response.ok && response.status === 404) {
        // Fallback to v1 endpoint if v3 doesn't exist
        response = await fetch(ApiEndpoints.dashboard.metrics('v1'), {
          method: 'GET',
          headers,
        });
      }

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
        business?: Record<string, unknown> | {
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
        const businessData = data.business as { total_verifications?: number; revenue?: number; growth_rate?: number; analytics_score?: number } | undefined;
        metrics = {
          totalMerchants: (businessData?.total_verifications) || 
                        (data.overview?.total_requests) || 0,
          revenue: (businessData?.revenue) || 0,
          growthRate: (businessData?.growth_rate) || 0,
          analyticsScore: (data.performance?.response_time) || 
                         (data.overview?.average_response_time) || 0,
        };
      } else {
        // Fallback to business object
        const businessData = data.business as { total_verifications?: number; revenue?: number; growth_rate?: number; analytics_score?: number } | undefined;
        metrics = {
          totalMerchants: (businessData?.total_verifications) || 0,
          revenue: (businessData?.revenue) || 0,
          growthRate: (businessData?.growth_rate) || 0,
          analyticsScore: (businessData?.analytics_score) || 0,
        };
      }

      apiCache.set(cacheKey, metrics, 60 * 1000); // 1 minute cache
      return metrics;
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
  const cacheKey = 'risk-metrics';
  
  const cached = apiCache.get<RiskMetrics>(cacheKey);
  if (cached) {
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      const response = await fetch(ApiEndpoints.risk.metrics(), {
        method: 'GET',
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

      const data = await handleResponse<RiskMetrics>(response);
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
  const cacheKey = 'system-metrics';
  
  const cached = apiCache.get<SystemMetrics>(cacheKey);
  if (cached) {
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      // Try multiple possible endpoints
      let response = await fetch(ApiEndpoints.monitoring.metrics(), {
        method: 'GET',
        headers,
      });

      if (!response.ok && response.status === 404) {
        response = await fetch(ApiEndpoints.monitoring.systemMetrics(), {
          method: 'GET',
          headers,
        });
      }

      if (!response.ok && response.status === 404) {
        response = await fetch(ApiEndpoints.monitoring.generalMetrics(), {
          method: 'GET',
          headers,
        });
      }

      if (!response.ok) {
        // Return default healthy values
        return {
          systemHealth: 100,
          serverStatus: 'Online',
          databaseStatus: 'Connected',
          responseTime: 0,
        };
      }

      const data = await handleResponse<SystemMetrics | { data?: SystemMetrics }>(response);
      
      // Handle different response formats
      const metrics: SystemMetrics = ('data' in data && data.data) ? data.data as SystemMetrics : data as SystemMetrics;
      
      apiCache.set(cacheKey, metrics, 30 * 1000); // 30 second cache for system metrics
      return metrics;
    } catch {
      // Return default healthy values on error
      return {
        systemHealth: 100,
        serverStatus: 'Online',
        databaseStatus: 'Connected',
        responseTime: 0,
      };
    }
  });
}

// Compliance Status
export async function getComplianceStatus(): Promise<ComplianceStatus> {
  const cacheKey = 'compliance-status';
  
  const cached = apiCache.get<ComplianceStatus>(cacheKey);
  if (cached) {
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      const response = await fetch(ApiEndpoints.compliance.status(), {
        method: 'GET',
        headers,
      });

      if (!response.ok) {
        // Return default values if endpoint doesn't exist
        return {
          overallScore: 0,
          pendingReviews: 0,
          complianceTrend: 'Stable',
          regulatoryFrameworks: 0,
        };
      }

      const data = await handleResponse<ComplianceStatus | {
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
      }>(response);
      
      // Handle different response formats
      let status: ComplianceStatus;
      
      if ('overallScore' in data || 'compliance_score' in data || 'overall_score' in data) {
        // Enhanced format with comprehensive data
        type EnhancedComplianceData = {
          compliance_score?: number;
          overall_score?: number;
          frameworks?: Array<{ framework_id: string; framework_name: string; status: string; score: number }>;
          requirements?: Array<{ requirement_id: string; status: string }>;
          alerts?: Array<{ id: string; severity: string; status: string }>;
        };
        const enhancedData = data as EnhancedComplianceData;
        
        const frameworks = Array.isArray(enhancedData.frameworks) ? enhancedData.frameworks : [];
        const requirements = Array.isArray(enhancedData.requirements) ? enhancedData.requirements : [];
        const alerts = Array.isArray(enhancedData.alerts) ? enhancedData.alerts : [];
        
        // Calculate pending reviews from requirements with pending status
        const pendingReviews = requirements.filter((req: { status: string }) => 
          req.status === 'pending' || req.status === 'in_progress'
        ).length;
        
        // Determine compliance trend from framework scores
        let complianceTrend: 'Improving' | 'Stable' | 'Declining' = 'Stable';
        if (frameworks.length > 0) {
          const avgScore = frameworks.reduce((sum: number, f: { score: number }) => sum + f.score, 0) / frameworks.length;
          if (avgScore >= 0.9) {
            complianceTrend = 'Improving';
          } else if (avgScore < 0.7) {
            complianceTrend = 'Declining';
          }
        }
        
        status = {
          overallScore: (enhancedData.compliance_score) || 
                      (enhancedData.overall_score) || 
                      ((data as ComplianceStatus).overallScore) || 0,
          pendingReviews: pendingReviews || ((data as ComplianceStatus).pendingReviews) || 0,
          complianceTrend: complianceTrend || ((data as ComplianceStatus).complianceTrend) || 'Stable',
          regulatoryFrameworks: frameworks.length || ((data as ComplianceStatus).regulatoryFrameworks) || 0,
          violations: alerts.filter((a: { severity: string }) => a.severity === 'high' || a.severity === 'critical').length,
        };
      } else {
        // Direct ComplianceStatus format
        status = data as ComplianceStatus;
      }
      
      apiCache.set(cacheKey, status, 5 * 60 * 1000); // 5 minute cache
      return status;
    } catch {
      // Return default values on error
      return {
        overallScore: 0,
        pendingReviews: 0,
        complianceTrend: 'Stable',
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

export async function createMerchant(data: CreateMerchantRequest): Promise<CreateMerchantResponse> {
  const token = getAuthToken();
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  return retryWithBackoff(async () => {
    try {
      const response = await fetch(ApiEndpoints.merchants.create(), {
        method: 'POST',
        headers,
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        const errorData = await ErrorHandler.parseErrorResponse(response);
        const errorMessage = errorData && typeof errorData === 'object' && 'message' in errorData
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
  const cacheKey = 'business-intelligence-metrics';
  
  const cached = apiCache.get<BusinessIntelligenceMetrics>(cacheKey);
  if (cached) {
    return cached;
  }

  const token = getAuthToken();
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  return requestDeduplicator.deduplicate(cacheKey, async () => {
    try {
      const response = await fetch(ApiEndpoints.businessIntelligence.metrics(), {
        method: 'GET',
        headers,
      });

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

