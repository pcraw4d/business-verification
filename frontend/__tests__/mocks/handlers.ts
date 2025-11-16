import { http, HttpResponse } from 'msw';

// MSW handlers for API mocking
// These intercept network requests at the fetch level, bypassing Jest module hoisting issues

// Use a pattern that matches both with and without protocol
const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';
// Extract just the path part for more flexible matching
const API_PATH = '/api/v1';

export const handlers = [
  // Get merchant by ID - MSW v2 can match by full URL or path pattern
  // Try both patterns to ensure matching works
  http.get(`${API_BASE_URL}${API_PATH}/merchants/:merchantId`, ({ params }) => {
    const merchantId = params.merchantId as string;
    
    // Return mock merchant data - ensure all values are serializable
    const mockData = {
      id: merchantId,
      businessName: 'Test Business',
      industry: 'Technology',
      status: 'active',
      email: 'test@example.com',
      phone: '+1-555-123-4567',
      website: 'https://test.com',
      address: {
        street: '123 Main St',
        city: 'San Francisco',
        state: 'CA',
        postalCode: '94102',
        country: 'USA',
      },
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };
    
    // MSW v2 HttpResponse.json() - now using Vitest per MSW FAQ recommendation
    // Vitest has none of the Node.js globals issues that Jest/JSDOM has
    // This should work correctly without the 500 error we saw with Jest
    return HttpResponse.json(mockData, { status: 200 });
  }),


  // Get merchant analytics
  http.get(`${API_BASE_URL}/api/v1/merchants/:merchantId/analytics`, () => {
    return HttpResponse.json({
      merchantId: 'merchant-123',
      classification: {
        primaryIndustry: 'Technology',
        confidenceScore: 0.95,
        riskLevel: 'low',
      },
      security: {
        trustScore: 0.8,
        sslValid: true,
      },
    });
  }),

  // Get website analysis
  http.get(`${API_BASE_URL}/api/v1/merchants/:merchantId/website-analysis`, () => {
    return HttpResponse.json({
      merchantId: 'merchant-123',
      domain: 'test.com',
      sslValid: true,
      trustScore: 0.8,
    });
  }),

  // Get risk assessment
  http.get(`${API_BASE_URL}/api/v1/merchants/:merchantId/risk-score`, () => {
    return HttpResponse.json({
      merchantId: 'merchant-123',
      riskScore: 0.3,
      riskLevel: 'low',
      factors: [],
    });
  }),

  // Start risk assessment
  http.post(`${API_BASE_URL}/api/v1/risk/assess`, () => {
    return HttpResponse.json({
      assessmentId: 'assessment-123',
      status: 'pending',
    });
  }),

  // Get assessment status
  http.get(`${API_BASE_URL}/api/v1/risk/assess/:assessmentId`, () => {
    return HttpResponse.json({
      assessmentId: 'assessment-123',
      status: 'completed',
      progress: 100,
    });
  }),

  // Get risk history
  http.get(`${API_BASE_URL}/api/v1/risk/history/:merchantId`, () => {
    return HttpResponse.json({
      merchantId: 'merchant-123',
      history: [],
      limit: 10,
      offset: 0,
      total: 0,
    });
  }),

  // Get risk predictions
  http.get(`${API_BASE_URL}/api/v1/risk/predictions/:merchantId`, () => {
    return HttpResponse.json({
      merchantId: 'merchant-123',
      predictions: [],
    });
  }),

  // Explain risk assessment
  http.get(`${API_BASE_URL}/api/v1/risk/explain/:assessmentId`, () => {
    return HttpResponse.json({
      assessmentId: 'assessment-123',
      explanation: 'Test explanation',
    });
  }),

  // Get risk recommendations
  http.get(`${API_BASE_URL}/api/v1/merchants/:merchantId/risk-recommendations`, () => {
    return HttpResponse.json({
      merchantId: 'merchant-123',
      recommendations: [],
    });
  }),

  // Get risk indicators
  http.get(`${API_BASE_URL}/api/v1/risk/indicators/:merchantId`, () => {
    return HttpResponse.json({
      merchantId: 'merchant-123',
      indicators: [],
    });
  }),

  // Get enrichment sources
  http.get(`${API_BASE_URL}/api/v1/merchants/:merchantId/enrichment/sources`, () => {
    return HttpResponse.json({
      sources: [],
    });
  }),

  // Trigger enrichment
  http.post(`${API_BASE_URL}/api/v1/merchants/:merchantId/enrichment/trigger`, () => {
    return HttpResponse.json({
      jobId: 'job-123',
      status: 'pending',
    });
  }),
];

