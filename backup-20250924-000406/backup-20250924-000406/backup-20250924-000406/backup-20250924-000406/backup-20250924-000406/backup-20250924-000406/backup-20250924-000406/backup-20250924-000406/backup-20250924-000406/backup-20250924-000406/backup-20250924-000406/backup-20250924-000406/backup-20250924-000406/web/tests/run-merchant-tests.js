#!/usr/bin/env node

/**
 * Merchant Playwright Test Runner
 * 
 * This script runs all merchant-related Playwright tests with proper configuration
 * and reporting for the KYB Platform merchant-centric UI implementation.
 */

const { execSync } = require('child_process');
const path = require('path');
const fs = require('fs');

// Test configuration
const testConfig = {
  // Test files to run
  testFiles: [
    'merchant-portfolio.spec.js',
    'merchant-detail.spec.js',
    'merchant-bulk-operations.spec.js',
    'merchant-comparison.spec.js',
    'merchant-hub-integration.spec.js'
  ],
  
  // Test environments
  environments: [
    'chromium',
    'firefox',
    'webkit'
  ],
  
  // Test modes
  modes: [
    'headed',    // Run with browser UI
    'headless'   // Run without browser UI
  ],
  
  // Output directories
  outputDir: 'test-results/merchant-tests',
  reportDir: 'test-results/merchant-reports'
};

// Colors for console output
const colors = {
  reset: '\x1b[0m',
  bright: '\x1b[1m',
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  magenta: '\x1b[35m',
  cyan: '\x1b[36m'
};

// Utility functions
function log(message, color = colors.reset) {
  console.log(`${color}${message}${colors.reset}`);
}

function logHeader(message) {
  log(`\n${colors.bright}${colors.cyan}${'='.repeat(60)}${colors.reset}`);
  log(`${colors.bright}${colors.cyan}${message}${colors.reset}`);
  log(`${colors.bright}${colors.cyan}${'='.repeat(60)}${colors.reset}\n`);
}

function logSuccess(message) {
  log(`${colors.green}✓ ${message}${colors.reset}`);
}

function logError(message) {
  log(`${colors.red}✗ ${message}${colors.reset}`);
}

function logWarning(message) {
  log(`${colors.yellow}⚠ ${message}${colors.reset}`);
}

function logInfo(message) {
  log(`${colors.blue}ℹ ${message}${colors.reset}`);
}

// Test execution functions
function createOutputDirectories() {
  logInfo('Creating output directories...');
  
  const dirs = [
    testConfig.outputDir,
    testConfig.reportDir,
    path.join(testConfig.outputDir, 'artifacts'),
    path.join(testConfig.outputDir, 'screenshots'),
    path.join(testConfig.outputDir, 'videos'),
    path.join(testConfig.reportDir, 'html'),
    path.join(testConfig.reportDir, 'json'),
    path.join(testConfig.reportDir, 'junit')
  ];
  
  dirs.forEach(dir => {
    if (!fs.existsSync(dir)) {
      fs.mkdirSync(dir, { recursive: true });
      logSuccess(`Created directory: ${dir}`);
    }
  });
}

function checkPrerequisites() {
  logInfo('Checking prerequisites...');
  
  try {
    // Check if Playwright is installed
    execSync('npx playwright --version', { stdio: 'pipe' });
    logSuccess('Playwright is installed');
  } catch (error) {
    logError('Playwright is not installed. Please run: npm install @playwright/test');
    process.exit(1);
  }
  
  try {
    // Check if test files exist
    testConfig.testFiles.forEach(file => {
      const filePath = path.join(__dirname, file);
      if (!fs.existsSync(filePath)) {
        logError(`Test file not found: ${file}`);
        process.exit(1);
      }
    });
    logSuccess('All test files found');
  } catch (error) {
    logError('Error checking test files');
    process.exit(1);
  }
}

function runTests(environment, mode, testFile) {
  const testName = path.basename(testFile, '.spec.js');
  const outputFile = path.join(testConfig.outputDir, `${testName}-${environment}-${mode}.json`);
  const reportFile = path.join(testConfig.reportDir, `${testName}-${environment}-${mode}.html`);
  
  logInfo(`Running ${testName} on ${environment} (${mode})...`);
  
  const command = [
    'npx playwright test',
    `--project=${environment}`,
    `--reporter=json,html,junit`,
    `--output=${testConfig.outputDir}`,
    `--reporter-options=outputFile=${outputFile}`,
    `--reporter-options=htmlFile=${reportFile}`,
    mode === 'headless' ? '--headed=false' : '--headed=true',
    `--timeout=30000`,
    `--retries=2`,
    testFile
  ].join(' ');
  
  try {
    const result = execSync(command, { 
      stdio: 'pipe',
      cwd: __dirname,
      encoding: 'utf8'
    });
    
    logSuccess(`${testName} on ${environment} (${mode}) completed successfully`);
    return { success: true, output: result };
  } catch (error) {
    logError(`${testName} on ${environment} (${mode}) failed`);
    logError(error.stdout || error.message);
    return { success: false, error: error.stdout || error.message };
  }
}

function generateTestReport(results) {
  logInfo('Generating test report...');
  
  const report = {
    timestamp: new Date().toISOString(),
    summary: {
      total: 0,
      passed: 0,
      failed: 0,
      skipped: 0
    },
    results: results,
    environments: testConfig.environments,
    testFiles: testConfig.testFiles
  };
  
  // Calculate summary
  results.forEach(result => {
    report.summary.total++;
    if (result.success) {
      report.summary.passed++;
    } else {
      report.summary.failed++;
    }
  });
  
  // Write JSON report
  const jsonReportPath = path.join(testConfig.reportDir, 'test-summary.json');
  fs.writeFileSync(jsonReportPath, JSON.stringify(report, null, 2));
  logSuccess(`JSON report written to: ${jsonReportPath}`);
  
  // Write HTML report
  const htmlReport = generateHTMLReport(report);
  const htmlReportPath = path.join(testConfig.reportDir, 'test-summary.html');
  fs.writeFileSync(htmlReportPath, htmlReport);
  logSuccess(`HTML report written to: ${htmlReportPath}`);
  
  return report;
}

function generateHTMLReport(report) {
  return `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Merchant Playwright Test Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 30px; }
        .summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin-bottom: 30px; }
        .summary-card { padding: 20px; border-radius: 8px; text-align: center; }
        .summary-card.total { background-color: #e3f2fd; }
        .summary-card.passed { background-color: #e8f5e8; }
        .summary-card.failed { background-color: #ffebee; }
        .summary-card.skipped { background-color: #fff3e0; }
        .summary-card h3 { margin: 0 0 10px 0; }
        .summary-card .number { font-size: 2em; font-weight: bold; }
        .results { margin-top: 30px; }
        .result-item { padding: 15px; margin: 10px 0; border-radius: 5px; border-left: 4px solid; }
        .result-item.success { background-color: #e8f5e8; border-left-color: #4caf50; }
        .result-item.failure { background-color: #ffebee; border-left-color: #f44336; }
        .result-item h4 { margin: 0 0 10px 0; }
        .result-item .details { font-size: 0.9em; color: #666; }
        .timestamp { text-align: center; color: #666; margin-top: 30px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Merchant Playwright Test Report</h1>
            <p>KYB Platform - Merchant-Centric UI Implementation</p>
        </div>
        
        <div class="summary">
            <div class="summary-card total">
                <h3>Total Tests</h3>
                <div class="number">${report.summary.total}</div>
            </div>
            <div class="summary-card passed">
                <h3>Passed</h3>
                <div class="number">${report.summary.passed}</div>
            </div>
            <div class="summary-card failed">
                <h3>Failed</h3>
                <div class="number">${report.summary.failed}</div>
            </div>
            <div class="summary-card skipped">
                <h3>Skipped</h3>
                <div class="number">${report.summary.skipped}</div>
            </div>
        </div>
        
        <div class="results">
            <h2>Test Results</h2>
            ${report.results.map(result => `
                <div class="result-item ${result.success ? 'success' : 'failure'}">
                    <h4>${result.testName} - ${result.environment} (${result.mode})</h4>
                    <div class="details">
                        Status: ${result.success ? 'PASSED' : 'FAILED'}<br>
                        ${result.success ? 'Output: Available in artifacts' : `Error: ${result.error}`}
                    </div>
                </div>
            `).join('')}
        </div>
        
        <div class="timestamp">
            <p>Report generated on: ${new Date(report.timestamp).toLocaleString()}</p>
        </div>
    </div>
</body>
</html>`;
}

function main() {
  logHeader('Merchant Playwright Test Runner');
  logInfo('Starting merchant-centric UI test execution...');
  
  // Check prerequisites
  checkPrerequisites();
  
  // Create output directories
  createOutputDirectories();
  
  // Run tests
  const results = [];
  
  testConfig.testFiles.forEach(testFile => {
    testConfig.environments.forEach(environment => {
      testConfig.modes.forEach(mode => {
        const result = runTests(environment, mode, testFile);
        results.push({
          testName: path.basename(testFile, '.spec.js'),
          environment,
          mode,
          success: result.success,
          output: result.output,
          error: result.error
        });
      });
    });
  });
  
  // Generate report
  const report = generateTestReport(results);
  
  // Display summary
  logHeader('Test Execution Summary');
  logInfo(`Total tests: ${report.summary.total}`);
  logSuccess(`Passed: ${report.summary.passed}`);
  if (report.summary.failed > 0) {
    logError(`Failed: ${report.summary.failed}`);
  }
  logInfo(`Skipped: ${report.summary.skipped}`);
  
  // Display report locations
  logHeader('Report Locations');
  logInfo(`JSON Report: ${path.join(testConfig.reportDir, 'test-summary.json')}`);
  logInfo(`HTML Report: ${path.join(testConfig.reportDir, 'test-summary.html')}`);
  logInfo(`Artifacts: ${testConfig.outputDir}`);
  
  // Exit with appropriate code
  if (report.summary.failed > 0) {
    logError('Some tests failed. Please check the reports for details.');
    process.exit(1);
  } else {
    logSuccess('All tests passed successfully!');
    process.exit(0);
  }
}

// Run the main function
if (require.main === module) {
  main();
}

module.exports = {
  testConfig,
  runTests,
  generateTestReport
};
