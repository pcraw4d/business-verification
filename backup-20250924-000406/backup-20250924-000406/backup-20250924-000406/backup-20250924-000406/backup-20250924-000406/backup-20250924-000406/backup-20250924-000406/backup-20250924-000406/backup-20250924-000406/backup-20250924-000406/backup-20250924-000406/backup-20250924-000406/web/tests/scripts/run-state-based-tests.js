#!/usr/bin/env node
// web/tests/scripts/run-state-based-tests.js

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

/**
 * State-based test runner script
 * Provides easy execution of state-based visual regression tests
 */

const TestTypes = {
  RISK_LEVELS: 'risk-levels',
  LOADING: 'loading',
  ERROR: 'error',
  EMPTY_DATA: 'empty-data',
  RESPONSIVE: 'responsive',
  ALL: 'all'
};

const Viewports = {
  MOBILE: 'mobile',
  TABLET: 'tablet',
  DESKTOP: 'desktop',
  LARGE: 'large',
  ALL: 'all'
};

class StateBasedTestRunner {
  constructor() {
    this.projectRoot = path.resolve(__dirname, '../../..');
    this.testResultsDir = path.join(this.projectRoot, 'test-results');
    this.artifactsDir = path.join(this.testResultsDir, 'state-based-artifacts');
    
    this.ensureDirectories();
  }

  ensureDirectories() {
    if (!fs.existsSync(this.testResultsDir)) {
      fs.mkdirSync(this.testResultsDir, { recursive: true });
    }
    if (!fs.existsSync(this.artifactsDir)) {
      fs.mkdirSync(this.artifactsDir, { recursive: true });
    }
  }

  /**
   * Run state-based tests for specific test type
   * @param {string} testType - Type of state tests to run
   * @param {string} viewport - Viewport size to test
   * @param {Object} options - Additional options
   */
  async runStateTests(testType = TestTypes.ALL, viewport = Viewports.ALL, options = {}) {
    console.log(`\n🧪 Running ${testType} state tests for ${viewport} viewport...`);
    
    const configFile = path.join(this.projectRoot, 'web/tests/config/state-based.config.js');
    const testFile = path.join(this.projectRoot, 'web/tests/visual/state-based-tests.spec.js');
    
    let command = `npx playwright test "${testFile}" --config="${configFile}"`;
    
    // Add test type filtering
    if (testType !== TestTypes.ALL) {
      const testTypeMap = {
        [TestTypes.RISK_LEVELS]: 'Risk Level State Tests',
        [TestTypes.LOADING]: 'Loading State Tests',
        [TestTypes.ERROR]: 'Error State Tests',
        [TestTypes.EMPTY_DATA]: 'Empty Data State Tests',
        [TestTypes.RESPONSIVE]: 'Responsive State Tests'
      };
      
      const grepPattern = testTypeMap[testType];
      if (grepPattern) {
        command += ` --grep="${grepPattern}"`;
      }
    }
    
    // Add viewport filtering
    if (viewport !== Viewports.ALL) {
      command += ` --grep="${viewport}"`;
    }
    
    // Add additional options
    if (options.headed) {
      command += ' --headed';
    }
    
    if (options.debug) {
      command += ' --debug';
    }
    
    if (options.ui) {
      command += ' --ui';
    }
    
    if (options.reporter) {
      command += ` --reporter=${options.reporter}`;
    }
    
    try {
      console.log(`Executing: ${command}`);
      execSync(command, { 
        stdio: 'inherit',
        cwd: this.projectRoot,
        env: { ...process.env, CI: options.ci ? 'true' : 'false' }
      });
      
      console.log(`✅ ${testType} state tests completed successfully`);
      return true;
    } catch (error) {
      console.error(`❌ ${testType} state tests failed:`, error.message);
      return false;
    }
  }

  /**
   * Run all state-based tests
   * @param {Object} options - Test options
   */
  async runAllStateTests(options = {}) {
    console.log('\n🚀 Starting comprehensive state-based testing...');
    
    const testTypes = [
      TestTypes.RISK_LEVELS,
      TestTypes.LOADING,
      TestTypes.ERROR,
      TestTypes.EMPTY_DATA,
      TestTypes.RESPONSIVE
    ];
    
    const results = {
      total: 0,
      passed: 0,
      failed: 0,
      testTypes: {}
    };
    
    for (const testType of testTypes) {
      console.log(`\n📱 Testing ${testType.toUpperCase()}...`);
      
      const success = await this.runStateTests(testType, Viewports.ALL, options);
      results.testTypes[testType] = success;
      results.total++;
      
      if (success) {
        results.passed++;
      } else {
        results.failed++;
      }
    }
    
    this.generateStateTestReport(results);
    return results;
  }

  /**
   * Run risk level state tests
   * @param {Object} options - Test options
   */
  async runRiskLevelTests(options = {}) {
    console.log('\n🎯 Running risk level state tests...');
    
    const riskLevels = ['low', 'medium', 'high', 'critical'];
    const results = {
      total: 0,
      passed: 0,
      failed: 0,
      riskLevels: {}
    };
    
    for (const level of riskLevels) {
      console.log(`\n🔍 Testing ${level.toUpperCase()} risk level...`);
      
      const success = await this.runStateTests(TestTypes.RISK_LEVELS, Viewports.ALL, {
        ...options,
        grep: level
      });
      
      results.riskLevels[level] = success;
      results.total++;
      
      if (success) {
        results.passed++;
      } else {
        results.failed++;
      }
    }
    
    this.generateRiskLevelReport(results);
    return results;
  }

  /**
   * Run loading state tests
   * @param {Object} options - Test options
   */
  async runLoadingStateTests(options = {}) {
    console.log('\n⏳ Running loading state tests...');
    
    const loadingTypes = ['overlay', 'skeleton', 'spinner'];
    const results = {
      total: 0,
      passed: 0,
      failed: 0,
      loadingTypes: {}
    };
    
    for (const type of loadingTypes) {
      console.log(`\n🔄 Testing ${type.toUpperCase()} loading state...`);
      
      const success = await this.runStateTests(TestTypes.LOADING, Viewports.ALL, {
        ...options,
        grep: type
      });
      
      results.loadingTypes[type] = success;
      results.total++;
      
      if (success) {
        results.passed++;
      } else {
        results.failed++;
      }
    }
    
    this.generateLoadingStateReport(results);
    return results;
  }

  /**
   * Run error state tests
   * @param {Object} options - Test options
   */
  async runErrorStateTests(options = {}) {
    console.log('\n❌ Running error state tests...');
    
    const errorTypes = ['overlay', 'card', 'form'];
    const results = {
      total: 0,
      passed: 0,
      failed: 0,
      errorTypes: {}
    };
    
    for (const type of errorTypes) {
      console.log(`\n🚨 Testing ${type.toUpperCase()} error state...`);
      
      const success = await this.runStateTests(TestTypes.ERROR, Viewports.ALL, {
        ...options,
        grep: type
      });
      
      results.errorTypes[type] = success;
      results.total++;
      
      if (success) {
        results.passed++;
      } else {
        results.failed++;
      }
    }
    
    this.generateErrorStateReport(results);
    return results;
  }

  /**
   * Run empty data state tests
   * @param {Object} options - Test options
   */
  async runEmptyDataTests(options = {}) {
    console.log('\n📭 Running empty data state tests...');
    
    const emptyTypes = ['dashboard', 'cards', 'charts'];
    const results = {
      total: 0,
      passed: 0,
      failed: 0,
      emptyTypes: {}
    };
    
    for (const type of emptyTypes) {
      console.log(`\n📊 Testing ${type.toUpperCase()} empty state...`);
      
      const success = await this.runStateTests(TestTypes.EMPTY_DATA, Viewports.ALL, {
        ...options,
        grep: type
      });
      
      results.emptyTypes[type] = success;
      results.total++;
      
      if (success) {
        results.passed++;
      } else {
        results.failed++;
      }
    }
    
    this.generateEmptyDataReport(results);
    return results;
  }

  /**
   * Run responsive state tests
   * @param {Object} options - Test options
   */
  async runResponsiveStateTests(options = {}) {
    console.log('\n📱 Running responsive state tests...');
    
    const viewports = [Viewports.MOBILE, Viewports.TABLET, Viewports.DESKTOP];
    const results = {
      total: 0,
      passed: 0,
      failed: 0,
      viewports: {}
    };
    
    for (const viewport of viewports) {
      console.log(`\n🖥️  Testing ${viewport.toUpperCase()} viewport...`);
      
      const success = await this.runStateTests(TestTypes.RESPONSIVE, viewport, options);
      results.viewports[viewport] = success;
      results.total++;
      
      if (success) {
        results.passed++;
      } else {
        results.failed++;
      }
    }
    
    this.generateResponsiveStateReport(results);
    return results;
  }

  /**
   * Generate comprehensive state test report
   * @param {Object} results - Test results
   */
  generateStateTestReport(results) {
    const reportPath = path.join(this.testResultsDir, 'state-based-test-report.json');
    
    const report = {
      timestamp: new Date().toISOString(),
      type: 'state-based-testing',
      summary: {
        total: results.total,
        passed: results.passed,
        failed: results.failed,
        successRate: results.total > 0 ? (results.passed / results.total * 100).toFixed(2) : 0
      },
      testTypes: results.testTypes,
      artifacts: {
        screenshots: this.artifactsDir,
        reports: path.join(this.testResultsDir, 'state-based-report')
      }
    };
    
    fs.writeFileSync(reportPath, JSON.stringify(report, null, 2));
    
    console.log('\n📊 State-Based Test Report:');
    console.log(`Total Test Types: ${results.total}`);
    console.log(`Passed: ${results.passed}`);
    console.log(`Failed: ${results.failed}`);
    console.log(`Success Rate: ${report.summary.successRate}%`);
    console.log(`\nReport saved to: ${reportPath}`);
  }

  /**
   * Generate risk level test report
   * @param {Object} results - Test results
   */
  generateRiskLevelReport(results) {
    const reportPath = path.join(this.testResultsDir, 'risk-level-test-report.json');
    
    const report = {
      timestamp: new Date().toISOString(),
      type: 'risk-level-testing',
      summary: {
        total: results.total,
        passed: results.passed,
        failed: results.failed,
        successRate: results.total > 0 ? (results.passed / results.total * 100).toFixed(2) : 0
      },
      riskLevels: results.riskLevels,
      artifacts: {
        screenshots: this.artifactsDir,
        reports: path.join(this.testResultsDir, 'state-based-report')
      }
    };
    
    fs.writeFileSync(reportPath, JSON.stringify(report, null, 2));
    
    console.log('\n🎯 Risk Level Test Report:');
    console.log(`Total Risk Levels: ${results.total}`);
    console.log(`Passed: ${results.passed}`);
    console.log(`Failed: ${results.failed}`);
    console.log(`Success Rate: ${report.summary.successRate}%`);
    console.log(`\nRisk level report saved to: ${reportPath}`);
  }

  /**
   * Generate loading state test report
   * @param {Object} results - Test results
   */
  generateLoadingStateReport(results) {
    const reportPath = path.join(this.testResultsDir, 'loading-state-test-report.json');
    
    const report = {
      timestamp: new Date().toISOString(),
      type: 'loading-state-testing',
      summary: {
        total: results.total,
        passed: results.passed,
        failed: results.failed,
        successRate: results.total > 0 ? (results.passed / results.total * 100).toFixed(2) : 0
      },
      loadingTypes: results.loadingTypes,
      artifacts: {
        screenshots: this.artifactsDir,
        reports: path.join(this.testResultsDir, 'state-based-report')
      }
    };
    
    fs.writeFileSync(reportPath, JSON.stringify(report, null, 2));
    
    console.log('\n⏳ Loading State Test Report:');
    console.log(`Total Loading Types: ${results.total}`);
    console.log(`Passed: ${results.passed}`);
    console.log(`Failed: ${results.failed}`);
    console.log(`Success Rate: ${report.summary.successRate}%`);
    console.log(`\nLoading state report saved to: ${reportPath}`);
  }

  /**
   * Generate error state test report
   * @param {Object} results - Test results
   */
  generateErrorStateReport(results) {
    const reportPath = path.join(this.testResultsDir, 'error-state-test-report.json');
    
    const report = {
      timestamp: new Date().toISOString(),
      type: 'error-state-testing',
      summary: {
        total: results.total,
        passed: results.passed,
        failed: results.failed,
        successRate: results.total > 0 ? (results.passed / results.total * 100).toFixed(2) : 0
      },
      errorTypes: results.errorTypes,
      artifacts: {
        screenshots: this.artifactsDir,
        reports: path.join(this.testResultsDir, 'state-based-report')
      }
    };
    
    fs.writeFileSync(reportPath, JSON.stringify(report, null, 2));
    
    console.log('\n❌ Error State Test Report:');
    console.log(`Total Error Types: ${results.total}`);
    console.log(`Passed: ${results.passed}`);
    console.log(`Failed: ${results.failed}`);
    console.log(`Success Rate: ${report.summary.successRate}%`);
    console.log(`\nError state report saved to: ${reportPath}`);
  }

  /**
   * Generate empty data test report
   * @param {Object} results - Test results
   */
  generateEmptyDataReport(results) {
    const reportPath = path.join(this.testResultsDir, 'empty-data-test-report.json');
    
    const report = {
      timestamp: new Date().toISOString(),
      type: 'empty-data-testing',
      summary: {
        total: results.total,
        passed: results.passed,
        failed: results.failed,
        successRate: results.total > 0 ? (results.passed / results.total * 100).toFixed(2) : 0
      },
      emptyTypes: results.emptyTypes,
      artifacts: {
        screenshots: this.artifactsDir,
        reports: path.join(this.testResultsDir, 'state-based-report')
      }
    };
    
    fs.writeFileSync(reportPath, JSON.stringify(report, null, 2));
    
    console.log('\n📭 Empty Data Test Report:');
    console.log(`Total Empty Types: ${results.total}`);
    console.log(`Passed: ${results.passed}`);
    console.log(`Failed: ${results.failed}`);
    console.log(`Success Rate: ${report.summary.successRate}%`);
    console.log(`\nEmpty data report saved to: ${reportPath}`);
  }

  /**
   * Generate responsive state test report
   * @param {Object} results - Test results
   */
  generateResponsiveStateReport(results) {
    const reportPath = path.join(this.testResultsDir, 'responsive-state-test-report.json');
    
    const report = {
      timestamp: new Date().toISOString(),
      type: 'responsive-state-testing',
      summary: {
        total: results.total,
        passed: results.passed,
        failed: results.failed,
        successRate: results.total > 0 ? (results.passed / results.total * 100).toFixed(2) : 0
      },
      viewports: results.viewports,
      artifacts: {
        screenshots: this.artifactsDir,
        reports: path.join(this.testResultsDir, 'state-based-report')
      }
    };
    
    fs.writeFileSync(reportPath, JSON.stringify(report, null, 2));
    
    console.log('\n📱 Responsive State Test Report:');
    console.log(`Total Viewports: ${results.total}`);
    console.log(`Passed: ${results.passed}`);
    console.log(`Failed: ${results.failed}`);
    console.log(`Success Rate: ${report.summary.successRate}%`);
    console.log(`\nResponsive state report saved to: ${reportPath}`);
  }

  /**
   * Clean test artifacts
   */
  cleanArtifacts() {
    console.log('\n🧹 Cleaning state-based test artifacts...');
    
    try {
      if (fs.existsSync(this.testResultsDir)) {
        fs.rmSync(this.testResultsDir, { recursive: true, force: true });
      }
      this.ensureDirectories();
      console.log('✅ State-based test artifacts cleaned');
    } catch (error) {
      console.error('❌ Failed to clean artifacts:', error.message);
    }
  }

  /**
   * Show help information
   */
  showHelp() {
    console.log(`
🧪 State-Based Test Runner

Usage: node run-state-based-tests.js [command] [options]

Commands:
  all                    Run all state-based tests
  risk-levels            Run risk level state tests
  loading                Run loading state tests
  error                  Run error state tests
  empty-data             Run empty data state tests
  responsive             Run responsive state tests
  clean                  Clean test artifacts
  help                   Show this help message

Options:
  --headed               Run tests in headed mode (visible browser)
  --debug                Run tests in debug mode
  --ui                   Run tests with UI mode
  --ci                   Run in CI mode
  --reporter <type>      Specify reporter type

Examples:
  node run-state-based-tests.js all
  node run-state-based-tests.js risk-levels --headed
  node run-state-based-tests.js loading --debug
  node run-state-based-tests.js error --ui
  node run-state-based-tests.js clean
    `);
  }
}

// CLI interface
async function main() {
  const args = process.argv.slice(2);
  const command = args[0];
  const options = {};
  
  // Parse options
  for (let i = 1; i < args.length; i++) {
    const arg = args[i];
    switch (arg) {
      case '--headed':
        options.headed = true;
        break;
      case '--debug':
        options.debug = true;
        break;
      case '--ui':
        options.ui = true;
        break;
      case '--ci':
        options.ci = true;
        break;
      case '--reporter':
        options.reporter = args[++i];
        break;
    }
  }
  
  const runner = new StateBasedTestRunner();
  
  try {
    switch (command) {
      case 'all':
        await runner.runAllStateTests(options);
        break;
      case 'risk-levels':
        await runner.runRiskLevelTests(options);
        break;
      case 'loading':
        await runner.runLoadingStateTests(options);
        break;
      case 'error':
        await runner.runErrorStateTests(options);
        break;
      case 'empty-data':
        await runner.runEmptyDataTests(options);
        break;
      case 'responsive':
        await runner.runResponsiveStateTests(options);
        break;
      case 'clean':
        runner.cleanArtifacts();
        break;
      case 'help':
      case '--help':
      case '-h':
        runner.showHelp();
        break;
      default:
        console.error(`❌ Unknown command: ${command}`);
        runner.showHelp();
        process.exit(1);
    }
  } catch (error) {
    console.error('❌ Test execution failed:', error.message);
    process.exit(1);
  }
}

if (require.main === module) {
  main();
}

module.exports = { StateBasedTestRunner, TestTypes, Viewports };
