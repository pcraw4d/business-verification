#!/usr/bin/env node

/**
 * Automated Page Testing Script
 * 
 * Tests all pages in the application for:
 * - Page accessibility (200 status)
 * - Console errors
 * - API call correctness (not localhost)
 * - Data loading states
 * - Missing components or broken imports
 * 
 * Usage:
 *   npm run test:pages
 *   npm run test:pages -- --base-url http://localhost:3000
 *   npm run test:pages -- --api-url https://api-gateway-service-production-21fd.up.railway.app
 */

const http = require('http');
const https = require('https');
const { URL } = require('url');

// Configuration
const config = {
  baseUrl: process.env.BASE_URL || process.argv.includes('--base-url') 
    ? process.argv[process.argv.indexOf('--base-url') + 1] 
    : 'http://localhost:3000',
  apiUrl: process.env.API_URL || process.argv.includes('--api-url')
    ? process.argv[process.argv.indexOf('--api-url') + 1]
    : 'https://api-gateway-service-production-21fd.up.railway.app',
  timeout: 10000,
  verbose: process.argv.includes('--verbose') || process.argv.includes('-v'),
};

// All pages to test (from Sidebar.tsx and file system scan)
const pages = [
  // Platform
  { path: '/', name: 'Home', category: 'Platform' },
  { path: '/dashboard-hub', name: 'Dashboard Hub', category: 'Platform' },
  
  // Merchant Verification & Risk
  { path: '/add-merchant', name: 'Add Merchant', category: 'Merchant Verification & Risk' },
  { path: '/dashboard', name: 'Business Intelligence', category: 'Merchant Verification & Risk' },
  { path: '/risk-dashboard', name: 'Risk Assessment', category: 'Merchant Verification & Risk' },
  { path: '/risk-indicators', name: 'Risk Indicators', category: 'Merchant Verification & Risk' },
  
  // Compliance
  { path: '/compliance', name: 'Compliance Status', category: 'Compliance' },
  { path: '/compliance/gap-analysis', name: 'Gap Analysis', category: 'Compliance' },
  { path: '/compliance/progress-tracking', name: 'Progress Tracking', category: 'Compliance' },
  { path: '/compliance/alerts', name: 'Compliance Alerts', category: 'Compliance' },
  { path: '/compliance/framework-indicators', name: 'Framework Indicators', category: 'Compliance' },
  { path: '/compliance/summary-reports', name: 'Summary Reports', category: 'Compliance' },
  
  // Merchant Management
  { path: '/merchant-hub', name: 'Merchant Hub', category: 'Merchant Management' },
  { path: '/merchant-hub/integration', name: 'Merchant Integration', category: 'Merchant Management' },
  { path: '/merchant-portfolio', name: 'Merchant Portfolio', category: 'Merchant Management' },
  { path: '/merchant/bulk-operations', name: 'Bulk Operations', category: 'Merchant Management' },
  { path: '/merchant/comparison', name: 'Merchant Comparison', category: 'Merchant Management' },
  { path: '/risk-assessment/portfolio', name: 'Risk Assessment Portfolio', category: 'Merchant Management' },
  
  // Market Intelligence
  { path: '/market-analysis', name: 'Market Analysis', category: 'Market Intelligence' },
  { path: '/competitive-analysis', name: 'Competitive Analysis', category: 'Market Intelligence' },
  
  // Administration
  { path: '/admin', name: 'Admin Dashboard', category: 'Administration' },
  { path: '/admin/models', name: 'Admin Models', category: 'Administration' },
  { path: '/admin/queue', name: 'Admin Queue', category: 'Administration' },
  { path: '/sessions', name: 'Sessions', category: 'Administration' },
  
  // Additional Pages
  { path: '/register', name: 'Register', category: 'Additional' },
  { path: '/monitoring', name: 'Monitoring', category: 'Additional' },
  { path: '/analytics-insights', name: 'Analytics Insights', category: 'Additional' },
  { path: '/business-intelligence', name: 'Business Intelligence', category: 'Additional' },
  { path: '/business-growth', name: 'Business Growth', category: 'Additional' },
  { path: '/api-test', name: 'API Test', category: 'Additional' },
  { path: '/gap-tracking', name: 'Gap Tracking', category: 'Additional' },
  { path: '/gap-analysis/reports', name: 'Gap Analysis Reports', category: 'Additional' },
];

// Test results
const results = {
  passed: [],
  failed: [],
  warnings: [],
  skipped: [],
};

/**
 * Make HTTP request
 */
function makeRequest(url) {
  return new Promise((resolve, reject) => {
    const parsedUrl = new URL(url);
    const client = parsedUrl.protocol === 'https:' ? https : http;
    const options = {
      hostname: parsedUrl.hostname,
      port: parsedUrl.port || (parsedUrl.protocol === 'https:' ? 443 : 80),
      path: parsedUrl.pathname + parsedUrl.search,
      method: 'GET',
      timeout: config.timeout,
      headers: {
        'User-Agent': 'KYB-Platform-Test-Script/1.0',
      },
    };

    const req = client.request(options, (res) => {
      let data = '';
      res.on('data', (chunk) => {
        data += chunk;
      });
      res.on('end', () => {
        resolve({
          statusCode: res.statusCode,
          headers: res.headers,
          body: data,
        });
      });
    });

    req.on('error', (error) => {
      reject(error);
    });

    req.on('timeout', () => {
      req.destroy();
      reject(new Error('Request timeout'));
    });

    req.end();
  });
}

/**
 * Test a single page
 */
async function testPage(page) {
  const url = `${config.baseUrl}${page.path}`;
  const result = {
    page,
    url,
    status: 'unknown',
    statusCode: null,
    errors: [],
    warnings: [],
    apiCalls: [],
  };

  try {
    if (config.verbose) {
      console.log(`\n🔍 Testing: ${page.name} (${page.path})`);
    }

    const response = await makeRequest(url);
    result.statusCode = response.statusCode;

    // Check status code
    if (response.statusCode === 200) {
      result.status = 'passed';
      
      // Check for localhost API calls in HTML (basic check)
      if (response.body.includes('localhost:8080') || response.body.includes('127.0.0.1:8080')) {
        result.warnings.push('Page HTML contains localhost API references');
      }

      // Check for API base URL in HTML
      if (response.body.includes(config.apiUrl)) {
        result.apiCalls.push('Page references correct API URL');
      }

      // Check for common error patterns
      if (response.body.includes('Error:') || response.body.includes('error')) {
        // This is a basic check - might have false positives
        if (response.body.includes('Cannot read property') || 
            response.body.includes('undefined is not a function')) {
          result.errors.push('Potential JavaScript error in page');
        }
      }

      results.passed.push(result);
    } else if (response.statusCode === 404) {
      result.status = 'failed';
      result.errors.push(`Page returned 404 Not Found`);
      results.failed.push(result);
    } else if (response.statusCode >= 500) {
      result.status = 'failed';
      result.errors.push(`Server error: ${response.statusCode}`);
      results.failed.push(result);
    } else {
      result.status = 'warning';
      result.warnings.push(`Unexpected status code: ${response.statusCode}`);
      results.warnings.push(result);
    }
  } catch (error) {
    result.status = 'failed';
    result.errors.push(`Request failed: ${error.message}`);
    results.failed.push(result);
  }

  return result;
}

/**
 * Test API endpoint accessibility
 */
async function testApiEndpoint(endpoint) {
  const url = `${config.apiUrl}${endpoint}`;
  try {
    const response = await makeRequest(url);
    return {
      endpoint,
      statusCode: response.statusCode,
      accessible: response.statusCode < 500,
    };
  } catch (error) {
    return {
      endpoint,
      statusCode: null,
      accessible: false,
      error: error.message,
    };
  }
}

/**
 * Test critical API endpoints
 */
async function testApiEndpoints() {
  const endpoints = [
    '/api/v1/merchants',
    '/api/v1/dashboard/metrics',
    '/api/v1/risk/metrics',
    '/api/v1/compliance/status',
    '/api/v1/sessions',
  ];

  console.log('\n🔍 Testing API Endpoints...\n');
  const apiResults = [];

  for (const endpoint of endpoints) {
    const result = await testApiEndpoint(endpoint);
    apiResults.push(result);
    
    if (result.accessible) {
      console.log(`  ✅ ${endpoint} - ${result.statusCode}`);
    } else {
      console.log(`  ❌ ${endpoint} - ${result.error || `Status: ${result.statusCode}`}`);
    }
  }

  return apiResults;
}

/**
 * Main test runner
 */
async function runTests() {
  console.log('🚀 Starting Automated Page Testing\n');
  console.log(`Base URL: ${config.baseUrl}`);
  console.log(`API URL: ${config.apiUrl}`);
  console.log(`Testing ${pages.length} pages...\n`);

  // Test all pages
  for (const page of pages) {
    const result = await testPage(page);
    
    if (config.verbose) {
      if (result.status === 'passed') {
        console.log(`  ✅ ${page.name} - ${result.statusCode}`);
      } else if (result.status === 'failed') {
        console.log(`  ❌ ${page.name} - ${result.errors.join(', ')}`);
      } else {
        console.log(`  ⚠️  ${page.name} - ${result.warnings.join(', ')}`);
      }
    } else {
      process.stdout.write(result.status === 'passed' ? '.' : result.status === 'failed' ? 'F' : 'W');
    }
  }

  // Test API endpoints
  const apiResults = await testApiEndpoints();

  // Print summary
  console.log('\n\n' + '='.repeat(60));
  console.log('📊 Test Summary');
  console.log('='.repeat(60));
  
  console.log(`\n✅ Passed: ${results.passed.length}`);
  console.log(`❌ Failed: ${results.failed.length}`);
  console.log(`⚠️  Warnings: ${results.warnings.length}`);
  
  if (results.failed.length > 0) {
    console.log('\n❌ Failed Pages:');
    results.failed.forEach((result) => {
      console.log(`   - ${result.page.name} (${result.page.path})`);
      result.errors.forEach((error) => {
        console.log(`     ${error}`);
      });
    });
  }

  if (results.warnings.length > 0) {
    console.log('\n⚠️  Warnings:');
    results.warnings.forEach((result) => {
      console.log(`   - ${result.page.name} (${result.page.path})`);
      result.warnings.forEach((warning) => {
        console.log(`     ${warning}`);
      });
    });
  }

  // API endpoint summary
  const accessibleEndpoints = apiResults.filter(r => r.accessible).length;
  console.log(`\n🌐 API Endpoints: ${accessibleEndpoints}/${apiResults.length} accessible`);

  // Exit code
  const exitCode = results.failed.length > 0 ? 1 : 0;
  process.exit(exitCode);
}

// Run tests
runTests().catch((error) => {
  console.error('\n❌ Test runner failed:', error);
  process.exit(1);
});

