/**
 * Lighthouse Performance Testing Configuration
 * 
 * Run Lighthouse audits for performance testing
 * Usage: npx lighthouse http://localhost:3000/merchant-details/merchant-123 --config-path=./tests/performance/lighthouse.config.js
 */

module.exports = {
  extends: 'lighthouse:default',
  settings: {
    onlyCategories: ['performance'],
    throttling: {
      rttMs: 40,
      throughputKbps: 10 * 1024,
      cpuSlowdownMultiplier: 1,
    },
    throttlingMethod: 'simulate',
    screenEmulation: {
      mobile: false,
      width: 1350,
      height: 940,
      deviceScaleFactor: 1,
    },
  },
  audits: [
    'first-contentful-paint',
    'largest-contentful-paint',
    'total-blocking-time',
    'cumulative-layout-shift',
    'speed-index',
    'interactive',
  ],
  categories: {
    performance: {
      title: 'Performance',
      auditRefs: [
        { id: 'first-contentful-paint', weight: 10 },
        { id: 'largest-contentful-paint', weight: 25 },
        { id: 'total-blocking-time', weight: 30 },
        { id: 'cumulative-layout-shift', weight: 15 },
        { id: 'speed-index', weight: 10 },
        { id: 'interactive', weight: 10 },
      ],
    },
  },
};

