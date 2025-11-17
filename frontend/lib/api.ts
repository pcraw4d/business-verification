// API client for Merchant Details
import { APICache } from '@/lib/api-cache';
import { ErrorHandler } from '@/lib/error-handler';
import { RequestDeduplicator } from '@/lib/request-deduplicator';
import type {
  AnalyticsData,
  AssessmentStatusResponse,
  EnrichmentSource,
  Merchant,
  RiskAssessment,
  RiskAssessmentRequest,
  RiskAssessmentResponse,
  RiskIndicatorsData,
  WebsiteAnalysisData,
} from '@/types/merchant';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';

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
    } catch (parseError) {
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
        console.log('[API] getMerchant: Fetching', `${API_BASE_URL}/api/v1/merchants/${merchantId}`);
      }
      const response = await fetch(`${API_BASE_URL}/api/v1/merchants/${merchantId}`, {
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
      const response = await fetch(`${API_BASE_URL}/api/v1/merchants/${merchantId}/analytics`, {
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
    const response = await fetch(`${API_BASE_URL}/api/v1/merchants/${merchantId}/website-analysis`, {
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
      const response = await fetch(`${API_BASE_URL}/api/v1/merchants/${merchantId}/risk-score`, {
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
    const response = await fetch(`${API_BASE_URL}/api/v1/risk/assess`, {
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
    const response = await fetch(`${API_BASE_URL}/api/v1/risk/assess/${assessmentId}`, {
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
        `${API_BASE_URL}/api/v1/risk/history/${merchantId}?limit=${limit}&offset=${offset}`,
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
      const params = new URLSearchParams({
        horizons: horizons.join(','),
        includeScenarios: String(includeScenarios),
        includeConfidence: String(includeConfidence),
      });
      const response = await fetch(
        `${API_BASE_URL}/api/v1/risk/predictions/${merchantId}?${params}`,
        {
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
      const response = await fetch(`${API_BASE_URL}/api/v1/risk/explain/${assessmentId}`, {
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
        `${API_BASE_URL}/api/v1/merchants/${merchantId}/risk-recommendations`,
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
      const params = new URLSearchParams();
      if (severity) params.append('severity', severity);
      if (status) params.append('status', status);
      
      const response = await fetch(
        `${API_BASE_URL}/api/v1/risk/indicators/${merchantId}?${params}`,
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
        `${API_BASE_URL}/api/v1/merchants/${merchantId}/enrichment/sources`,
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
      `${API_BASE_URL}/api/v1/merchants/${merchantId}/enrichment/trigger`,
      {
        method: 'POST',
        headers,
        body: JSON.stringify({ source }),
      }
    );
    return handleResponse<EnrichmentJobResponse>(response);
  });
}

