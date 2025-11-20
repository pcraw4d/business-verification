#!/usr/bin/env node

/**
 * Page Performance Measurement Script
 * 
 * Measures page load times and Web Vitals for frontend pages
 * Usage: node scripts/measure-page-performance.js [url]
 */

const { chromium } = require('playwright');

const BASE_URL = process.env.PLAYWRIGHT_TEST_BASE_URL || 'http://localhost:3000';
const PAGES = [
  { name: 'Merchant Details', path: '/merchant-details/merchant-123', target: 2.0 },
  { name: 'Business Intelligence Dashboard', path: '/dashboard', target: 3.0 },
  { name: 'Risk Dashboard', path: '/risk-dashboard', target: 3.0 },
  { name: 'Risk Indicators Dashboard', path: '/risk-indicators', target: 3.0 },
];

async function measurePagePerformance(page, url, name, target) {
  console.log(`\n📊 Measuring: ${name}`);
  console.log(`   URL: ${url}`);

  // Navigate to page
  await page.goto(url, { waitUntil: 'domcontentloaded' });

  // Wait for main content
  try {
    await page.waitForSelector('h1, [role="heading"]', { timeout: 5000 });
  } catch (e) {
    console.log(`   ⚠️  Warning: Could not find main heading`);
  }

  // Measure performance metrics
  const metrics = await page.evaluate(() => {
    const perfData = performance.getEntriesByType('navigation')[0];
    if (!perfData) return null;

    const navTiming = perfData;
    const paintEntries = performance.getEntriesByType('paint');
    const fcp = paintEntries.find(entry => entry.name === 'first-contentful-paint');
    const lcpEntries = performance.getEntriesByType('largest-contentful-paint');
    const lcp = lcpEntries.length > 0 ? lcpEntries[lcpEntries.length - 1] : null;

    return {
      // Load time metrics
      domContentLoaded: navTiming.domContentLoadedEventEnd - navTiming.fetchStart,
      loadComplete: navTiming.loadEventEnd - navTiming.fetchStart,
      
      // Web Vitals
      firstContentfulPaint: fcp ? fcp.startTime : null,
      largestContentfulPaint: lcp ? (lcp.renderTime || lcp.loadTime || lcp.startTime) : null,
      
      // Resource timing
      dns: navTiming.domainLookupEnd - navTiming.domainLookupStart,
      tcp: navTiming.connectEnd - navTiming.connectStart,
      request: navTiming.responseStart - navTiming.requestStart,
      response: navTiming.responseEnd - navTiming.responseStart,
      domProcessing: navTiming.domInteractive - navTiming.responseEnd,
    };
  });

  if (!metrics) {
    console.log(`   ❌ Could not measure performance metrics`);
    return null;
  }

  // Display results
  console.log(`   ┌─────────────────────────────────────────┐`);
  console.log(`   │ Performance Metrics                   │`);
  console.log(`   ├─────────────────────────────────────────┤`);
  console.log(`   │ DOM Content Loaded: ${(metrics.domContentLoaded / 1000).toFixed(2)}s`);
  console.log(`   │ Load Complete:      ${(metrics.loadComplete / 1000).toFixed(2)}s`);
  
  if (metrics.firstContentfulPaint) {
    console.log(`   │ First Contentful Paint: ${(metrics.firstContentfulPaint / 1000).toFixed(2)}s`);
  }
  
  if (metrics.largestContentfulPaint) {
    console.log(`   │ Largest Contentful Paint: ${(metrics.largestContentfulPaint / 1000).toFixed(2)}s`);
  }
  
  console.log(`   ├─────────────────────────────────────────┤`);
  console.log(`   │ DNS Lookup:         ${metrics.dns.toFixed(0)}ms`);
  console.log(`   │ TCP Connection:     ${metrics.tcp.toFixed(0)}ms`);
  console.log(`   │ Request:            ${metrics.request.toFixed(0)}ms`);
  console.log(`   │ Response:           ${metrics.response.toFixed(0)}ms`);
  console.log(`   │ DOM Processing:    ${metrics.domProcessing.toFixed(0)}ms`);
  console.log(`   └─────────────────────────────────────────┘`);

  // Check if target is met
  const loadTimeSeconds = metrics.loadComplete / 1000;
  if (loadTimeSeconds <= target) {
    console.log(`   ✅ Load time (${loadTimeSeconds.toFixed(2)}s) meets target (${target}s)`);
  } else {
    console.log(`   ⚠️  Load time (${loadTimeSeconds.toFixed(2)}s) exceeds target (${target}s)`);
  }

  return metrics;
}

async function runPerformanceTests() {
  console.log('⚡ Frontend Performance Testing');
  console.log('═'.repeat(50));
  console.log(`Base URL: ${BASE_URL}`);
  console.log('');

  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext();
  const page = await context.newPage();

  const results = [];

  for (const pageConfig of PAGES) {
    const url = `${BASE_URL}${pageConfig.path}`;
    const metrics = await measurePagePerformance(page, url, pageConfig.name, pageConfig.target);
    if (metrics) {
      results.push({
        name: pageConfig.name,
        path: pageConfig.path,
        loadTime: metrics.loadComplete / 1000,
        target: pageConfig.target,
        metrics,
      });
    }
  }

  await browser.close();

  // Summary
  console.log('\n📈 Performance Summary');
  console.log('═'.repeat(50));
  
  let allPassed = true;
  for (const result of results) {
    const status = result.loadTime <= result.target ? '✅' : '⚠️';
    if (result.loadTime > result.target) allPassed = false;
    
    console.log(`${status} ${result.name}: ${result.loadTime.toFixed(2)}s (target: ${result.target}s)`);
  }

  console.log('');
  if (allPassed) {
    console.log('✅ All pages meet performance targets!');
    process.exit(0);
  } else {
    console.log('⚠️  Some pages exceed performance targets');
    process.exit(1);
  }
}

// Run tests
runPerformanceTests().catch((error) => {
  console.error('❌ Error running performance tests:', error);
  process.exit(1);
});

