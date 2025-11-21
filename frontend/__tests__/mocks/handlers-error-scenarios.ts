import { http, HttpResponse } from 'msw';

// Extended MSW handlers for error scenarios and Phase 2 testing
// These handlers support testing error states, missing data, and edge cases

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';
const API_PATH = '/api/v1';

// Helper to determine if MSW should return error based on merchant ID pattern
const shouldReturnError = (merchantId: string, errorType: string): boolean => {
  return merchantId.includes(errorType);
};

// Helper to determine if merchant has missing data
const hasMissingData = (merchantId: string, dataType: string): boolean => {
  return merchantId.includes(`-no-${dataType}`);
};

export const errorHandlers = [
  // Error: Merchant not found (404)
  http.get(`${API_BASE_URL}${API_PATH}/merchants/:merchantId`, ({ params }) => {
    const merchantId = params.merchantId as string;
    
    if (shouldReturnError(merchantId, 'not-found') || merchantId === 'merchant-404') {
      return HttpResponse.json(
        { code: 'NOT_FOUND', message: 'Merchant not found' },
        { status: 404 }
      );
    }
    
    // Return null to let other handlers process
    return null;
  }),

  // Error: Server error (500)
  http.get(`${API_BASE_URL}${API_PATH}/merchants/:merchantId/risk-score`, ({ params }) => {
    const merchantId = params.merchantId as string;
    
    if (shouldReturnError(merchantId, 'server-error') || merchantId === 'merchant-500') {
      return HttpResponse.json(
        { code: 'INTERNAL_SERVER_ERROR', message: 'Internal server error' },
        { status: 500 }
      );
    }
    
    // Missing risk assessment (no data, but 200 response)
    if (hasMissingData(merchantId, 'risk') || merchantId === 'merchant-no-risk') {
      return HttpResponse.json(
        { code: 'NOT_FOUND', message: 'No risk assessment found' },
        { status: 200 }
      );
    }
    
    return null;
  }),

  // Error: Missing portfolio statistics (404)
  http.get(`${API_BASE_URL}${API_PATH}/merchants/statistics`, ({ request }) => {
    // Check if we should return 404 for portfolio stats
    const url = new URL(request.url);
    if (url.searchParams.get('mock-portfolio-404') === 'true') {
      return HttpResponse.json(
        { code: 'NOT_FOUND', message: 'Portfolio statistics not available' },
        { status: 404 }
      );
    }
    
    return null;
  }),

  // Error: Missing merchant analytics (404)
  http.get(`${API_BASE_URL}${API_PATH}/merchants/:merchantId/analytics`, ({ params }) => {
    const merchantId = params.merchantId as string;
    
    if (hasMissingData(merchantId, 'analytics') || merchantId === 'merchant-no-analytics') {
      return HttpResponse.json(
        { code: 'NOT_FOUND', message: 'Merchant analytics not found' },
        { status: 404 }
      );
    }
    
    return null;
  }),

  // Error: Missing portfolio analytics (404)
  http.get(`${API_BASE_URL}${API_PATH}/merchants/analytics`, ({ request }) => {
    const url = new URL(request.url);
    if (url.searchParams.get('mock-portfolio-analytics-404') === 'true') {
      return HttpResponse.json(
        { code: 'NOT_FOUND', message: 'Portfolio analytics not found' },
        { status: 404 }
      );
    }
    
    return null;
  }),

  // Error: Missing industry code for benchmarks
  http.get(`${API_BASE_URL}${API_PATH}/risk/benchmarks`, ({ request }) => {
    const url = new URL(request.url);
    const mcc = url.searchParams.get('mcc');
    const naics = url.searchParams.get('naics');
    const sic = url.searchParams.get('sic');
    
    // If no industry codes provided, return error
    if (!mcc && !naics && !sic) {
      return HttpResponse.json(
        { code: 'MISSING_INDUSTRY_CODE', message: 'Industry code is required' },
        { status: 400 }
      );
    }
    
    return null;
  }),

  // Error: Network timeout simulation
  http.get(`${API_BASE_URL}${API_PATH}/merchants/:merchantId`, async ({ params }) => {
    const merchantId = params.merchantId as string;
    
    if (shouldReturnError(merchantId, 'timeout') || merchantId === 'merchant-timeout') {
      // Simulate timeout by waiting longer than request timeout
      await new Promise(resolve => setTimeout(resolve, 10000));
      return HttpResponse.json({}, { status: 200 });
    }
    
    return null;
  }),

  // Error: CORS error simulation (can't be fully simulated, but can return CORS-like error)
  http.get(`${API_BASE_URL}${API_PATH}/merchants/:merchantId`, ({ params }) => {
    const merchantId = params.merchantId as string;
    
    if (shouldReturnError(merchantId, 'cors') || merchantId === 'merchant-cors') {
      // Return response without CORS headers to simulate CORS error
      return HttpResponse.json(
        { code: 'CORS_ERROR', message: 'CORS policy blocked the request' },
        { 
          status: 200,
          headers: {
            // Intentionally omit CORS headers
          }
        }
      );
    }
    
    return null;
  }),
];

// Test merchant data handlers (for specific test scenarios)
export const testMerchantHandlers = [
  // Merchant with no risk assessment
  http.get(`${API_BASE_URL}${API_PATH}/merchants/merchant-no-risk/risk-score`, () => {
    return HttpResponse.json(
      { code: 'NOT_FOUND', message: 'No risk assessment found' },
      { status: 200 }
    );
  }),

  // Merchant with no portfolio stats
  http.get(`${API_BASE_URL}${API_PATH}/merchants/statistics`, ({ request }) => {
    const url = new URL(request.url);
    if (url.searchParams.get('merchant') === 'merchant-no-portfolio-stats') {
      return HttpResponse.json(
        { code: 'NOT_FOUND', message: 'Portfolio statistics are being calculated' },
        { status: 404 }
      );
    }
    return null;
  }),

  // Merchant with no analytics
  http.get(`${API_BASE_URL}${API_PATH}/merchants/merchant-no-analytics/analytics`, () => {
    return HttpResponse.json(
      { code: 'NOT_FOUND', message: 'Merchant analytics not found' },
      { status: 404 }
    );
  }),

  // Merchant with no industry code
  http.get(`${API_BASE_URL}${API_PATH}/merchants/merchant-no-industry-code/analytics`, () => {
    return HttpResponse.json({
      merchantId: 'merchant-no-industry-code',
      classification: {
        primaryIndustry: null, // No industry code
        confidenceScore: 0.0,
      },
      security: {
        trustScore: 0.5,
      },
    });
  }),

  // Merchant with complete data (for success scenarios)
  http.get(`${API_BASE_URL}${API_PATH}/merchants/merchant-complete-123/risk-score`, () => {
    return HttpResponse.json({
      merchant_id: 'merchant-complete-123',
      risk_score: 0.35,
      risk_level: 'low',
      confidence_score: 0.95,
      assessment_date: new Date().toISOString(),
      factors: [
        { name: 'Financial Stability', score: 0.9, weight: 0.3 },
        { name: 'Business History', score: 0.8, weight: 0.25 },
      ],
    });
  }),

  http.get(`${API_BASE_URL}${API_PATH}/merchants/merchant-complete-123/analytics`, () => {
    return HttpResponse.json({
      merchantId: 'merchant-complete-123',
      classification: {
        primaryIndustry: 'Technology',
        confidenceScore: 0.95,
        riskLevel: 'low',
        mccCodes: [{ code: '5734', description: 'Computer Software Stores', confidence: 0.92 }],
        naicsCodes: [{ code: '541511', description: 'Custom Computer Programming Services', confidence: 0.93 }],
      },
      security: {
        trustScore: 0.8,
        sslValid: true,
      },
    });
  }),
];

