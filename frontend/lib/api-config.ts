/**
 * Centralized API Configuration
 * 
 * Single source of truth for API base URL and endpoint configuration.
 * Provides runtime validation, environment detection, and type-safe endpoint builders.
 */

/**
 * Get the API base URL with validation and warnings
 */
export function getApiBaseUrl(): string {
  const apiUrl = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';
  
  // Runtime validation for production
  if (typeof window !== 'undefined') {
    const isProduction = window.location.hostname !== 'localhost' && 
                        window.location.hostname !== '127.0.0.1' &&
                        !window.location.hostname.startsWith('192.168.');
    
    if (isProduction && apiUrl.includes('localhost')) {
      console.error(
        '%c[API Config Error]',
        'color: red; font-weight: bold; font-size: 14px;',
        'API base URL is set to localhost in production!',
        '\nThis will cause API calls to fail.',
        '\nPlease set NEXT_PUBLIC_API_BASE_URL environment variable in Railway.'
      );
      console.warn(
        'Current API URL:',
        apiUrl,
        '\nExpected: https://api-gateway-service-production-21fd.up.railway.app'
      );
    }
  }
  
  return apiUrl;
}

/**
 * Check if we're in development environment
 */
export function isDevelopment(): boolean {
  if (typeof window === 'undefined') {
    return process.env.NODE_ENV === 'development';
  }
  
  return window.location.hostname === 'localhost' || 
         window.location.hostname === '127.0.0.1' ||
         window.location.hostname.startsWith('192.168.');
}

/**
 * Check if we're in production environment
 */
export function isProduction(): boolean {
  return !isDevelopment();
}

/**
 * Build a full API endpoint URL
 * @param path - API path (e.g., '/api/v1/merchants' or 'api/v1/merchants')
 * @returns Full URL to the endpoint
 */
export function buildApiUrl(path: string): string {
  const baseUrl = getApiBaseUrl();
  const cleanPath = path.startsWith('/') ? path : `/${path}`;
  
  // Remove trailing slash from base URL if present
  const cleanBaseUrl = baseUrl.replace(/\/$/, '');
  
  return `${cleanBaseUrl}${cleanPath}`;
}

/**
 * Build a WebSocket URL from the API base URL
 * @param path - WebSocket path (e.g., '/api/v1/risk/ws')
 * @returns WebSocket URL (ws:// or wss://)
 */
export function buildWebSocketUrl(path: string): string {
  const baseUrl = getApiBaseUrl();
  const cleanPath = path.startsWith('/') ? path : `/${path}`;
  
  // Convert http/https to ws/wss
  const wsUrl = baseUrl
    .replace(/^http:\/\//, 'ws://')
    .replace(/^https:\/\//, 'wss://')
    .replace(/\/$/, '');
  
  return `${wsUrl}${cleanPath}`;
}

/**
 * Type-safe API endpoint builders
 */
export const ApiEndpoints = {
  // Merchant endpoints
  merchants: {
    list: () => buildApiUrl('/api/v1/merchants'),
    get: (id: string) => buildApiUrl(`/api/v1/merchants/${id}`),
    create: () => buildApiUrl('/api/v1/merchants'),
    update: (id: string) => buildApiUrl(`/api/v1/merchants/${id}`),
    delete: (id: string) => buildApiUrl(`/api/v1/merchants/${id}`),
    search: () => buildApiUrl('/api/v1/merchants/search'),
    analytics: (id: string) => buildApiUrl(`/api/v1/merchants/${id}/analytics`),
    websiteAnalysis: (id: string) => buildApiUrl(`/api/v1/merchants/${id}/website-analysis`),
    riskScore: (id: string) => buildApiUrl(`/api/v1/merchants/${id}/risk-score`),
    riskRecommendations: (id: string) => buildApiUrl(`/api/v1/merchants/${id}/risk-recommendations`),
    enrichmentSources: (id: string) => buildApiUrl(`/api/v1/merchants/${id}/enrichment/sources`),
    triggerEnrichment: (id: string) => buildApiUrl(`/api/v1/merchants/${id}/enrichment/trigger`),
    export: (id: string, format?: string) => {
      const url = buildApiUrl(`/api/v1/merchants/${id}/export`);
      return format ? `${url}?format=${format}` : url;
    },
    bulkUpdate: () => buildApiUrl('/api/v1/merchants/bulk/update'),
    statistics: () => buildApiUrl('/api/v1/merchants/statistics'),
    portfolioTypes: () => buildApiUrl('/api/v1/merchants/portfolio-types'),
    riskLevels: () => buildApiUrl('/api/v1/merchants/risk-levels'),
  },
  
  // Risk endpoints
  risk: {
    assess: () => buildApiUrl('/api/v1/risk/assess'),
    getAssessment: (id: string) => buildApiUrl(`/api/v1/risk/assess/${id}`),
    history: (merchantId: string, limit?: number, offset?: number) => {
      const params = new URLSearchParams();
      if (limit) params.append('limit', limit.toString());
      if (offset) params.append('offset', offset.toString());
      const query = params.toString();
      return buildApiUrl(`/api/v1/risk/history/${merchantId}${query ? `?${query}` : ''}`);
    },
    predictions: (merchantId: string, horizon?: string) => {
      const params = new URLSearchParams();
      if (horizon) params.append('horizon', horizon);
      const query = params.toString();
      return buildApiUrl(`/api/v1/risk/predictions/${merchantId}${query ? `?${query}` : ''}`);
    },
    explain: (assessmentId: string) => buildApiUrl(`/api/v1/risk/explain/${assessmentId}`),
    indicators: (merchantId: string, filters?: Record<string, string>) => {
      const params = new URLSearchParams(filters);
      const query = params.toString();
      return buildApiUrl(`/api/v1/risk/indicators/${merchantId}${query ? `?${query}` : ''}`);
    },
    metrics: () => buildApiUrl('/api/v1/risk/metrics'),
    ws: () => buildWebSocketUrl('/api/v1/risk/ws'),
  },
  
  // Dashboard endpoints
  dashboard: {
    metrics: (version: 'v1' | 'v3' = 'v3') => buildApiUrl(`/api/${version}/dashboard/metrics`),
  },
  
  // Compliance endpoints
  compliance: {
    status: () => buildApiUrl('/api/v1/compliance/status'),
  },
  
  // Session endpoints
  sessions: {
    list: () => buildApiUrl('/api/v1/sessions'),
  },
  
  // Auth endpoints
  auth: {
    register: () => buildApiUrl('/v1/auth/register'),
    login: () => buildApiUrl('/v1/auth/login'),
  },
  
  // Business Intelligence endpoints
  businessIntelligence: {
    metrics: () => buildApiUrl('/api/v1/business-intelligence/metrics'),
  },
  
  // Monitoring endpoints
  monitoring: {
    metrics: () => buildApiUrl('/api/v1/monitoring/metrics'),
    systemMetrics: () => buildApiUrl('/api/v1/system/metrics'),
    generalMetrics: () => buildApiUrl('/api/v1/metrics'),
  },
} as const;

/**
 * Export the base URL for backward compatibility
 * @deprecated Use getApiBaseUrl() or ApiEndpoints instead
 */
export const API_BASE_URL = getApiBaseUrl();

