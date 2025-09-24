/**
 * Global teardown for Playwright visual regression tests
 * This runs once after all tests complete
 */

async function globalTeardown(config) {
  console.log('🏁 KYB Platform Visual Regression Tests Completed');
  
  // Log test summary
  const fs = require('fs');
  const path = require('path');
  
  const resultsFile = path.join(process.cwd(), 'test-results', 'results.json');
  if (fs.existsSync(resultsFile)) {
    try {
      const results = JSON.parse(fs.readFileSync(resultsFile, 'utf8'));
      console.log('📊 Test Summary:');
      console.log(`   Total Tests: ${results.stats?.total || 'N/A'}`);
      console.log(`   Passed: ${results.stats?.passed || 'N/A'}`);
      console.log(`   Failed: ${results.stats?.failed || 'N/A'}`);
      console.log(`   Skipped: ${results.stats?.skipped || 'N/A'}`);
    } catch (error) {
      console.log('⚠️  Could not read test results summary');
    }
  }
  
  console.log('📁 Test artifacts saved to: test-results/artifacts/');
  console.log('📊 HTML report available at: playwright-report/index.html');
  console.log('✅ Global teardown completed');
}

module.exports = globalTeardown;
