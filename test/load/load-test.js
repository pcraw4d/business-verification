import { check, sleep } from 'k6';
import http from 'k6/http';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
export const errorRate = new Rate('errors');
export const responseTime = new Trend('response_time');

// Test configuration
export const options = {
  stages: [
    { duration: '2m', target: 10 },   // Ramp up to 10 users
    { duration: '5m', target: 10 },   // Stay at 10 users
    { duration: '2m', target: 20 },   // Ramp up to 20 users
    { duration: '5m', target: 20 },   // Stay at 20 users
    { duration: '2m', target: 50 },   // Ramp up to 50 users
    { duration: '5m', target: 50 },   // Stay at 50 users
    { duration: '2m', target: 0 },    // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000'], // 95% of requests must complete below 1s
    http_req_failed: ['rate<0.1'],     // Error rate must be below 10%
    errors: ['rate<0.1'],              // Custom error rate must be below 10%
  },
};

// Test data
const testBusinesses = [
  {
    name: "Test Restaurant Corp",
    description: "A fine dining restaurant specializing in Italian cuisine with full bar service",
    address: "123 Main Street, New York, NY 10001",
    industry: "Restaurant",
    website: "https://testrestaurant.com",
    phone: "+1-555-123-4567",
    email: "info@testrestaurant.com"
  },
  {
    name: "Tech Startup Inc",
    description: "A technology startup developing AI-powered business solutions",
    address: "456 Tech Avenue, San Francisco, CA 94105",
    industry: "Technology",
    website: "https://techstartup.com",
    phone: "+1-555-987-6543",
    email: "contact@techstartup.com"
  },
  {
    name: "Retail Store LLC",
    description: "A retail store selling clothing and accessories",
    address: "789 Retail Boulevard, Chicago, IL 60601",
    industry: "Retail",
    website: "https://retailstore.com",
    phone: "+1-555-456-7890",
    email: "sales@retailstore.com"
  },
  {
    name: "Healthcare Services",
    description: "A healthcare provider offering medical services and consultations",
    address: "321 Health Street, Boston, MA 02101",
    industry: "Healthcare",
    website: "https://healthcareservices.com",
    phone: "+1-555-321-0987",
    email: "info@healthcareservices.com"
  },
  {
    name: "Financial Advisory",
    description: "A financial advisory firm providing investment and wealth management services",
    address: "654 Finance Plaza, Miami, FL 33101",
    industry: "Financial Services",
    website: "https://financialadvisory.com",
    phone: "+1-555-654-3210",
    email: "advisors@financialadvisory.com"
  }
];

// Base URL from environment
const BASE_URL = __ENV.K6_BASE_URL || 'https://kyb-api-gateway-production.up.railway.app';
const API_KEY = __ENV.K6_API_KEY || 'test-api-key';

// Headers
const headers = {
  'Content-Type': 'application/json',
  'Accept': 'application/json',
  'Authorization': `Bearer ${API_KEY}`,
};

export default function() {
  // Test 1: Health Check
  testHealthCheck();
  
  // Test 2: Business Verification
  testBusinessVerification();
  
  // Test 3: Merchant Management
  testMerchantManagement();
  
  // Test 4: Monitoring
  testMonitoring();
  
  // Test 5: Business Intelligence
  testBusinessIntelligence();
  
  // Sleep between requests
  sleep(1);
}

function testHealthCheck() {
  const response = http.get(`${BASE_URL}/health`, { headers });
  
  const success = check(response, {
    'health check status is 200': (r) => r.status === 200,
    'health check response time < 500ms': (r) => r.timings.duration < 500,
    'health check response contains status': (r) => r.json('status') === 'healthy',
  });
  
  errorRate.add(!success);
  responseTime.add(response.timings.duration);
  
  if (!success) {
    console.error(`Health check failed: ${response.status} - ${response.body}`);
  }
}

function testBusinessVerification() {
  // Select random test business
  const testBusiness = testBusinesses[Math.floor(Math.random() * testBusinesses.length)];
  
  // Add some randomization to avoid caching
  testBusiness.name = `${testBusiness.name} ${Date.now()}`;
  
  const response = http.post(`${BASE_URL}/verify`, JSON.stringify(testBusiness), { headers });
  
  const success = check(response, {
    'business verification status is 200': (r) => r.status === 200,
    'business verification response time < 2000ms': (r) => r.timings.duration < 2000,
    'business verification response has ID': (r) => r.json('id') !== undefined,
    'business verification response has status': (r) => r.json('status') !== undefined,
    'business verification response has score': (r) => r.json('score') !== undefined,
  });
  
  errorRate.add(!success);
  responseTime.add(response.timings.duration);
  
  if (!success) {
    console.error(`Business verification failed: ${response.status} - ${response.body}`);
  }
}

function testMerchantManagement() {
  // First, create a test business
  const testBusiness = testBusinesses[Math.floor(Math.random() * testBusinesses.length)];
  testBusiness.name = `${testBusiness.name} ${Date.now()}`;
  
  const createResponse = http.post(`${BASE_URL}/verify`, JSON.stringify(testBusiness), { headers });
  
  if (createResponse.status === 200) {
    const businessId = createResponse.json('id');
    
    // Test merchant retrieval
    const getResponse = http.get(`${BASE_URL}/merchants/${businessId}`, { headers });
    
    const success = check(getResponse, {
      'merchant retrieval status is 200': (r) => r.status === 200,
      'merchant retrieval response time < 1000ms': (r) => r.timings.duration < 1000,
      'merchant retrieval response has ID': (r) => r.json('id') === businessId,
    });
    
    errorRate.add(!success);
    responseTime.add(getResponse.timings.duration);
    
    if (!success) {
      console.error(`Merchant retrieval failed: ${getResponse.status} - ${getResponse.body}`);
    }
  } else {
    errorRate.add(true);
    console.error(`Merchant creation failed: ${createResponse.status} - ${createResponse.body}`);
  }
}

function testMonitoring() {
  const response = http.get(`${BASE_URL}/metrics`, { headers });
  
  const success = check(response, {
    'monitoring status is 200': (r) => r.status === 200,
    'monitoring response time < 1000ms': (r) => r.timings.duration < 1000,
    'monitoring response has metrics': (r) => r.json() !== null,
  });
  
  errorRate.add(!success);
  responseTime.add(response.timings.duration);
  
  if (!success) {
    console.error(`Monitoring failed: ${response.status} - ${response.body}`);
  }
}

function testBusinessIntelligence() {
  const response = http.get(`${BASE_URL}/dashboard/executive`, { headers });
  
  const success = check(response, {
    'BI dashboard status is 200': (r) => r.status === 200,
    'BI dashboard response time < 2000ms': (r) => r.timings.duration < 2000,
    'BI dashboard response has data': (r) => r.json() !== null,
  });
  
  errorRate.add(!success);
  responseTime.add(response.timings.duration);
  
  if (!success) {
    console.error(`BI dashboard failed: ${response.status} - ${response.body}`);
  }
}

// Setup function (runs once at the beginning)
export function setup() {
  console.log('Starting load test against:', BASE_URL);
  
  // Verify the service is accessible
  const healthResponse = http.get(`${BASE_URL}/health`);
  if (healthResponse.status !== 200) {
    throw new Error(`Service is not accessible: ${healthResponse.status}`);
  }
  
  console.log('Service is accessible, starting load test...');
  return { baseUrl: BASE_URL };
}

// Teardown function (runs once at the end)
export function teardown(data) {
  console.log('Load test completed for:', data.baseUrl);
}
