#!/usr/bin/env node

/**
 * Component Test Runner
 * Runs all component tests and generates coverage reports
 */

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

// Test configuration
const TEST_CONFIG = {
    testDirectory: __dirname,
    coverageDirectory: path.join(__dirname, '..', '..', 'coverage', 'components'),
    reportDirectory: path.join(__dirname, '..', '..', 'test-results', 'components'),
    targetCoverage: 80 // 80% coverage target
};

// Component test files
const COMPONENT_TESTS = [
    'merchant-search.test.js',
    'portfolio-type-filter.test.js',
    'risk-level-indicator.test.js',
    'session-manager.test.js',
    'merchant-navigation.test.js',
    'coming-soon-banner.test.js',
    'mock-data-warning.test.js',
    'bulk-progress-tracker.test.js',
    'merchant-comparison.test.js',
    'merchant-context.test.js',
    'navigation.test.js',
    'responsive-design.test.js',
    'component-interaction.test.js'
];

// Component source files
const COMPONENT_SOURCES = [
    'merchant-search.js',
    'portfolio-type-filter.js',
    'risk-level-indicator.js',
    'session-manager.js',
    'merchant-navigation.js',
    'coming-soon-banner.js',
    'mock-data-warning.js',
    'bulk-progress-tracker.js',
    'merchant-comparison.js',
    'merchant-context.js',
    'navigation.js'
];

class ComponentTestRunner {
    constructor() {
        this.results = {
            totalTests: 0,
            passedTests: 0,
            failedTests: 0,
            skippedTests: 0,
            coverage: 0,
            testFiles: [],
            errors: []
        };
    }

    async runAllTests() {
        console.log('🧪 Starting Component Test Suite...\n');
        
        // Create directories
        this.createDirectories();
        
        // Check if Jest is available
        if (!this.checkJestAvailability()) {
            console.error('❌ Jest is not available. Please install Jest to run component tests.');
            process.exit(1);
        }
        
        // Run individual component tests
        await this.runIndividualTests();
        
        // Run integration tests
        await this.runIntegrationTests();
        
        // Generate coverage report
        await this.generateCoverageReport();
        
        // Generate test summary
        this.generateTestSummary();
        
        // Display results
        this.displayResults();
        
        return this.results;
    }

    createDirectories() {
        const dirs = [
            TEST_CONFIG.coverageDirectory,
            TEST_CONFIG.reportDirectory
        ];
        
        dirs.forEach(dir => {
            if (!fs.existsSync(dir)) {
                fs.mkdirSync(dir, { recursive: true });
            }
        });
    }

    checkJestAvailability() {
        try {
            execSync('npx jest --version', { stdio: 'pipe' });
            return true;
        } catch (error) {
            return false;
        }
    }

    async runIndividualTests() {
        console.log('📋 Running Individual Component Tests...\n');
        
        for (const testFile of COMPONENT_TESTS) {
            const testPath = path.join(TEST_CONFIG.testDirectory, testFile);
            
            if (fs.existsSync(testPath)) {
                console.log(`  🔍 Running ${testFile}...`);
                
                try {
                    const result = await this.runTestFile(testPath);
                    this.results.testFiles.push({
                        file: testFile,
                        status: result.success ? 'passed' : 'failed',
                        tests: result.tests,
                        errors: result.errors
                    });
                    
                    if (result.success) {
                        console.log(`  ✅ ${testFile} - ${result.tests} tests passed`);
                        this.results.passedTests += result.tests;
                    } else {
                        console.log(`  ❌ ${testFile} - ${result.errors.length} errors`);
                        this.results.failedTests += result.tests;
                        this.results.errors.push(...result.errors);
                    }
                    
                    this.results.totalTests += result.tests;
                } catch (error) {
                    console.log(`  ❌ ${testFile} - Failed to run: ${error.message}`);
                    this.results.errors.push({
                        file: testFile,
                        error: error.message
                    });
                }
            } else {
                console.log(`  ⚠️  ${testFile} - File not found`);
            }
        }
        
        console.log('');
    }

    async runTestFile(testPath) {
        try {
            const command = `npx jest "${testPath}" --verbose --no-coverage --json`;
            const output = execSync(command, { 
                cwd: TEST_CONFIG.testDirectory,
                encoding: 'utf8',
                stdio: 'pipe'
            });
            
            const result = JSON.parse(output);
            
            return {
                success: result.success,
                tests: result.numTotalTests,
                errors: result.testResults
                    .filter(test => test.status === 'failed')
                    .map(test => ({
                        file: test.name,
                        errors: test.failureMessages
                    }))
            };
        } catch (error) {
            // Try to parse Jest output even if it fails
            try {
                const output = error.stdout || error.stderr || '';
                const result = JSON.parse(output);
                
                return {
                    success: false,
                    tests: result.numTotalTests || 0,
                    errors: [{
                        file: testPath,
                        error: error.message
                    }]
                };
            } catch (parseError) {
                return {
                    success: false,
                    tests: 0,
                    errors: [{
                        file: testPath,
                        error: error.message
                    }]
                };
            }
        }
    }

    async runIntegrationTests() {
        console.log('🔗 Running Integration Tests...\n');
        
        const integrationTests = [
            'responsive-design.test.js',
            'component-interaction.test.js'
        ];
        
        for (const testFile of integrationTests) {
            const testPath = path.join(TEST_CONFIG.testDirectory, testFile);
            
            if (fs.existsSync(testPath)) {
                console.log(`  🔍 Running ${testFile}...`);
                
                try {
                    const result = await this.runTestFile(testPath);
                    
                    if (result.success) {
                        console.log(`  ✅ ${testFile} - ${result.tests} tests passed`);
                        this.results.passedTests += result.tests;
                    } else {
                        console.log(`  ❌ ${testFile} - ${result.errors.length} errors`);
                        this.results.failedTests += result.tests;
                        this.results.errors.push(...result.errors);
                    }
                    
                    this.results.totalTests += result.tests;
                } catch (error) {
                    console.log(`  ❌ ${testFile} - Failed to run: ${error.message}`);
                    this.results.errors.push({
                        file: testFile,
                        error: error.message
                    });
                }
            }
        }
        
        console.log('');
    }

    async generateCoverageReport() {
        console.log('📊 Generating Coverage Report...\n');
        
        try {
            // Run tests with coverage
            const coverageCommand = `npx jest "${TEST_CONFIG.testDirectory}/*.test.js" --coverage --coverageDirectory="${TEST_CONFIG.coverageDirectory}" --coverageReporters=json,text,html`;
            
            execSync(coverageCommand, { 
                cwd: TEST_CONFIG.testDirectory,
                stdio: 'pipe'
            });
            
            // Read coverage report
            const coveragePath = path.join(TEST_CONFIG.coverageDirectory, 'coverage-final.json');
            if (fs.existsSync(coveragePath)) {
                const coverageData = JSON.parse(fs.readFileSync(coveragePath, 'utf8'));
                this.results.coverage = this.calculateCoverage(coverageData);
            }
            
            console.log(`  ✅ Coverage report generated: ${this.results.coverage.toFixed(2)}%`);
        } catch (error) {
            console.log(`  ⚠️  Coverage report generation failed: ${error.message}`);
            this.results.coverage = 0;
        }
        
        console.log('');
    }

    calculateCoverage(coverageData) {
        let totalStatements = 0;
        let coveredStatements = 0;
        let totalFunctions = 0;
        let coveredFunctions = 0;
        let totalBranches = 0;
        let coveredBranches = 0;
        let totalLines = 0;
        let coveredLines = 0;
        
        Object.values(coverageData).forEach(file => {
            if (file.s) {
                totalStatements += Object.keys(file.s).length;
                coveredStatements += Object.values(file.s).filter(count => count > 0).length;
            }
            
            if (file.f) {
                totalFunctions += Object.keys(file.f).length;
                coveredFunctions += Object.values(file.f).filter(count => count > 0).length;
            }
            
            if (file.b) {
                totalBranches += Object.keys(file.b).length;
                coveredBranches += Object.values(file.b).filter(count => count > 0).length;
            }
            
            if (file.l) {
                totalLines += Object.keys(file.l).length;
                coveredLines += Object.values(file.l).filter(count => count > 0).length;
            }
        });
        
        // Calculate overall coverage percentage
        const statementCoverage = totalStatements > 0 ? (coveredStatements / totalStatements) * 100 : 0;
        const functionCoverage = totalFunctions > 0 ? (coveredFunctions / totalFunctions) * 100 : 0;
        const branchCoverage = totalBranches > 0 ? (coveredBranches / totalBranches) * 100 : 0;
        const lineCoverage = totalLines > 0 ? (coveredLines / totalLines) * 100 : 0;
        
        return (statementCoverage + functionCoverage + branchCoverage + lineCoverage) / 4;
    }

    generateTestSummary() {
        const summary = {
            timestamp: new Date().toISOString(),
            totalTests: this.results.totalTests,
            passedTests: this.results.passedTests,
            failedTests: this.results.failedTests,
            skippedTests: this.results.skippedTests,
            coverage: this.results.coverage,
            testFiles: this.results.testFiles,
            errors: this.results.errors,
            success: this.results.failedTests === 0 && this.results.coverage >= TEST_CONFIG.targetCoverage
        };
        
        const summaryPath = path.join(TEST_CONFIG.reportDirectory, 'test-summary.json');
        fs.writeFileSync(summaryPath, JSON.stringify(summary, null, 2));
        
        // Generate HTML report
        this.generateHTMLReport(summary);
    }

    generateHTMLReport(summary) {
        const html = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Component Test Results</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 30px; }
        .status { padding: 10px; border-radius: 5px; margin: 10px 0; }
        .status.success { background: #d4edda; color: #155724; border: 1px solid #c3e6cb; }
        .status.failure { background: #f8d7da; color: #721c24; border: 1px solid #f5c6cb; }
        .status.warning { background: #fff3cd; color: #856404; border: 1px solid #ffeaa7; }
        .metrics { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin: 20px 0; }
        .metric { text-align: center; padding: 15px; background: #f8f9fa; border-radius: 5px; }
        .metric-value { font-size: 2em; font-weight: bold; margin-bottom: 5px; }
        .metric-label { color: #666; }
        .test-files { margin: 20px 0; }
        .test-file { padding: 10px; margin: 5px 0; border-radius: 5px; border-left: 4px solid; }
        .test-file.passed { background: #d4edda; border-color: #28a745; }
        .test-file.failed { background: #f8d7da; border-color: #dc3545; }
        .errors { margin: 20px 0; }
        .error { padding: 10px; margin: 5px 0; background: #f8d7da; border-radius: 5px; border-left: 4px solid #dc3545; }
        .timestamp { text-align: center; color: #666; margin-top: 30px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🧪 Component Test Results</h1>
            <div class="status ${summary.success ? 'success' : 'failure'}">
                ${summary.success ? '✅ All tests passed!' : '❌ Some tests failed'}
            </div>
        </div>
        
        <div class="metrics">
            <div class="metric">
                <div class="metric-value">${summary.totalTests}</div>
                <div class="metric-label">Total Tests</div>
            </div>
            <div class="metric">
                <div class="metric-value">${summary.passedTests}</div>
                <div class="metric-label">Passed</div>
            </div>
            <div class="metric">
                <div class="metric-value">${summary.failedTests}</div>
                <div class="metric-label">Failed</div>
            </div>
            <div class="metric">
                <div class="metric-value">${summary.coverage.toFixed(1)}%</div>
                <div class="metric-label">Coverage</div>
            </div>
        </div>
        
        <div class="test-files">
            <h2>Test Files</h2>
            ${summary.testFiles.map(file => `
                <div class="test-file ${file.status}">
                    <strong>${file.file}</strong> - ${file.tests} tests - ${file.status}
                </div>
            `).join('')}
        </div>
        
        ${summary.errors.length > 0 ? `
        <div class="errors">
            <h2>Errors</h2>
            ${summary.errors.map(error => `
                <div class="error">
                    <strong>${error.file}</strong><br>
                    ${error.error || error.errors?.join('<br>') || 'Unknown error'}
                </div>
            `).join('')}
        </div>
        ` : ''}
        
        <div class="timestamp">
            Generated on ${new Date(summary.timestamp).toLocaleString()}
        </div>
    </div>
</body>
</html>
        `;
        
        const htmlPath = path.join(TEST_CONFIG.reportDirectory, 'test-results.html');
        fs.writeFileSync(htmlPath, html);
    }

    displayResults() {
        console.log('📋 Test Results Summary:');
        console.log('========================\n');
        
        console.log(`Total Tests: ${this.results.totalTests}`);
        console.log(`Passed: ${this.results.passedTests}`);
        console.log(`Failed: ${this.results.failedTests}`);
        console.log(`Skipped: ${this.results.skippedTests}`);
        console.log(`Coverage: ${this.results.coverage.toFixed(2)}%`);
        console.log(`Target Coverage: ${TEST_CONFIG.targetCoverage}%`);
        
        const success = this.results.failedTests === 0 && this.results.coverage >= TEST_CONFIG.targetCoverage;
        console.log(`\nOverall Status: ${success ? '✅ PASSED' : '❌ FAILED'}`);
        
        if (this.results.errors.length > 0) {
            console.log('\n❌ Errors:');
            this.results.errors.forEach(error => {
                console.log(`  - ${error.file}: ${error.error || 'Unknown error'}`);
            });
        }
        
        console.log(`\n📊 Reports generated in: ${TEST_CONFIG.reportDirectory}`);
        console.log(`📈 Coverage report: ${TEST_CONFIG.coverageDirectory}`);
        
        if (!success) {
            process.exit(1);
        }
    }
}

// Run tests if this script is executed directly
if (require.main === module) {
    const runner = new ComponentTestRunner();
    runner.runAllTests().catch(error => {
        console.error('❌ Test runner failed:', error);
        process.exit(1);
    });
}

module.exports = ComponentTestRunner;
