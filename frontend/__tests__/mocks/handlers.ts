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
  http.get(`${API_BASE_URL}/api/v1/merchants/:merchantId/analytics`, ({ params }) => {
    const merchantId = params.merchantId as string;
    return HttpResponse.json({
      merchantId: merchantId,
      classification: {
        primaryIndustry: 'Technology',
        confidenceScore: 0.95,
        riskLevel: 'low',
        mccCodes: [],
        sicCodes: [],
        naicsCodes: [],
      },
      security: {
        trustScore: 0.8,
        sslValid: true,
        sslExpiryDate: new Date(Date.now() + 90 * 24 * 60 * 60 * 1000).toISOString(),
        securityHeaders: [],
      },
      quality: {
        completenessScore: 0.85,
        dataPoints: 12,
        missingFields: [],
      },
      intelligence: {
        businessAge: 5,
        employeeCount: 50,
        annualRevenue: 1000000,
      },
      timestamp: new Date().toISOString(),
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

  // Get merchant risk score / risk assessment
  // Note: This endpoint is used by both getRiskAssessment and getMerchantRiskScore
  // We check for a query parameter to determine which format to return
  http.get(`${API_BASE_URL}/api/v1/merchants/:merchantId/risk-score`, ({ params, request }) => {
    const merchantId = params.merchantId as string;
    const url = new URL(request.url);
    const format = url.searchParams.get('format');
    
    // If format=assessment, return RiskAssessmentSchema format
    if (format === 'assessment') {
      return HttpResponse.json({
        id: `assessment-${merchantId}`,
        merchantId: merchantId,
        status: 'completed',
        options: {
          includeHistory: false,
          includePredictions: false,
        },
        result: {
          overallScore: 0.3,
          riskLevel: 'low',
          factors: [
            {
              name: 'Business Age',
              score: 0.2,
              weight: 0.3,
            },
            {
              name: 'Financial Stability',
              score: 0.4,
              weight: 0.4,
            },
          ],
        },
        progress: 100,
        estimatedCompletion: new Date().toISOString(),
        createdAt: new Date(Date.now() - 86400000).toISOString(),
        updatedAt: new Date().toISOString(),
        completedAt: new Date().toISOString(),
      });
    }
    
    // Default: return MerchantRiskScoreSchema format
    return HttpResponse.json({
      merchant_id: merchantId,
      risk_score: 0.3,
      risk_level: 'low',
      confidence_score: 0.85,
      assessment_date: new Date().toISOString(),
      factors: [
        {
          category: 'Business Age',
          score: 0.2,
          weight: 0.3,
        },
        {
          category: 'Financial Stability',
          score: 0.4,
          weight: 0.4,
        },
      ],
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
  http.get(`${API_BASE_URL}/api/v1/risk/assess/:assessmentId`, ({ params }) => {
    const assessmentId = params.assessmentId as string;
    return HttpResponse.json({
      assessmentId: assessmentId,
      merchantId: 'merchant-123',
      status: 'completed',
      progress: 100,
      estimatedCompletion: new Date().toISOString(),
      result: {
        overallScore: 0.3,
        riskLevel: 'low',
        factors: [
          {
            name: 'Business Age',
            score: 0.2,
            weight: 0.3,
          },
        ],
      },
      completedAt: new Date().toISOString(),
    });
  }),

  // Export merchant data - match with query parameter (format is in query string, not path)
  http.get(`${API_BASE_URL}${API_PATH}/merchants/:merchantId/export`, ({ request, params }) => {
    const url = new URL(request.url);
    const format = url.searchParams.get('format') || 'csv';
    const merchantId = params.merchantId as string;
    
    // Return appropriate blob based on format
    if (format === 'csv') {
      return HttpResponse.text(`id,name,status\n${merchantId},Test Business,active`, {
        headers: { 'Content-Type': 'text/csv' },
      });
    } else if (format === 'pdf') {
      // Return a minimal PDF-like blob
      const pdfContent = '%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n>>\nendobj\nxref\n0 1\ntrailer\n<<\n/Root 1 0 R\n>>\n%%EOF';
      return HttpResponse.text(pdfContent, {
        headers: { 'Content-Type': 'application/pdf' },
      });
    } else if (format === 'excel') {
      // Return text instead of ArrayBuffer to match type
      return HttpResponse.text('', {
        headers: { 'Content-Type': 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' },
      });
    }
    
    return HttpResponse.text('', { status: 400 });
  }),
  
  // Also match the pattern used in tests (with format as path param - for backward compatibility)
  http.get(`${API_BASE_URL}${API_PATH}/merchants/:merchantId/export/:format`, ({ params }) => {
    const format = params.format as string;
    const merchantId = params.merchantId as string;
    
    if (format === 'csv') {
      return HttpResponse.text(`id,name,status\n${merchantId},Test Business,active`, {
        headers: { 'Content-Type': 'text/csv' },
      });
    } else if (format === 'pdf') {
      const pdfContent = '%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n>>\nendobj\nxref\n0 1\ntrailer\n<<\n/Root 1 0 R\n>>\n%%EOF';
      return HttpResponse.text(pdfContent, {
        headers: { 'Content-Type': 'application/pdf' },
      });
    }
    
    return HttpResponse.text('', { status: 400 });
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

  // Create merchant
  http.post(`${API_BASE_URL}${API_PATH}/merchants`, () => {
    return HttpResponse.json({
      id: 'merchant-123',
      name: 'Test Business',
      status: 'active',
      createdAt: new Date().toISOString(),
    });
  }),

  // Get portfolio statistics
  http.get(`${API_BASE_URL}${API_PATH}/merchants/statistics`, () => {
    return HttpResponse.json({
      totalMerchants: 100,
      totalAssessments: 95,
      averageRiskScore: 0.45,
      riskDistribution: {
        low: 40,
        medium: 45,
        high: 15,
      },
      industryBreakdown: [
        {
          industry: 'Technology',
          count: 30,
          averageRiskScore: 0.35,
        },
        {
          industry: 'Finance',
          count: 25,
          averageRiskScore: 0.50,
        },
        {
          industry: 'Retail',
          count: 20,
          averageRiskScore: 0.55,
        },
        {
          industry: 'Other',
          count: 25,
          averageRiskScore: 0.45,
        },
      ],
      countryBreakdown: [
        {
          country: 'USA',
          count: 60,
          averageRiskScore: 0.40,
        },
        {
          country: 'Canada',
          count: 20,
          averageRiskScore: 0.50,
        },
        {
          country: 'UK',
          count: 20,
          averageRiskScore: 0.55,
        },
      ],
      timestamp: new Date().toISOString(),
    });
  }),

  // Get portfolio analytics
  http.get(`${API_BASE_URL}${API_PATH}/merchants/analytics`, () => {
    return HttpResponse.json({
      portfolioAnalytics: {
        averageRiskScore: 0.45,
        totalMerchants: 100,
        industryDistribution: {
          Technology: 30,
          Finance: 25,
          Retail: 20,
          Other: 25,
        },
      },
    });
  }),

  // Get risk benchmarks
  http.get(`${API_BASE_URL}${API_PATH}/risk/benchmarks`, ({ request }) => {
    const url = new URL(request.url);
    const mcc = url.searchParams.get('mcc');
    const naics = url.searchParams.get('naics');
    const sic = url.searchParams.get('sic');
    
    return HttpResponse.json({
      industry_code: mcc || naics || sic || '541511',
      average_risk_score: 0.42,
      median_risk_score: 0.40,
      p25_risk_score: 0.30,
      p75_risk_score: 0.55,
      p90_risk_score: 0.65,
      sample_size: 150,
    });
  }),
];

