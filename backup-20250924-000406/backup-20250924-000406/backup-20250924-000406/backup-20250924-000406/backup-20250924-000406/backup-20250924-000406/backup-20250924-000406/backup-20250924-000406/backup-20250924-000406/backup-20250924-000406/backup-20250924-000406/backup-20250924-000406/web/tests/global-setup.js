/**
 * Global setup for Playwright visual regression tests
 * This runs once before all tests
 */

async function globalSetup(config) {
  console.log('🚀 Starting KYB Platform Visual Regression Tests');
  console.log('📁 Test directory:', config?.testDir || 'web/tests');
  console.log('🌐 Base URL:', config?.use?.baseURL || 'http://localhost:8080');
  
  // Create test results directory if it doesn't exist
  const fs = require('fs');
  const path = require('path');
  
  const resultsDir = path.join(process.cwd(), 'test-results');
  if (!fs.existsSync(resultsDir)) {
    fs.mkdirSync(resultsDir, { recursive: true });
    console.log('📁 Created test-results directory');
  }
  
  const artifactsDir = path.join(resultsDir, 'artifacts');
  if (!fs.existsSync(artifactsDir)) {
    fs.mkdirSync(artifactsDir, { recursive: true });
    console.log('📁 Created artifacts directory');
  }
  
  console.log('✅ Global setup completed');
}

module.exports = globalSetup;
