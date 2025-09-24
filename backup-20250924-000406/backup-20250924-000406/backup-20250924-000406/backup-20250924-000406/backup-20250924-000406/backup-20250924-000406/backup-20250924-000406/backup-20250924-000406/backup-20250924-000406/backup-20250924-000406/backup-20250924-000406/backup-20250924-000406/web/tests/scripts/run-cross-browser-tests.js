#!/usr/bin/env node
// web/tests/scripts/run-cross-browser-tests.js

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

/**
 * Cross-browser test runner script
 * Provides easy execution of cross-browser tests with different configurations
 */

const Browsers = {
  CHROME: 'chrome',
  FIREFOX: 'firefox',
  SAFARI: 'safari',
  EDGE: 'edge',
  ALL: 'all'
};

const TestTypes = {
  VISUAL: 'visual',
  FUNCTIONAL: 'functional',
  PERFORMANCE: 'performance',
  ALL: 'all'
};

class CrossBrowserTestRunner {
  constructor() {
    this.projectRoot = path.resolve(__dirname, '../../..');
    this.testResultsDir = path.join(this.projectRoot, 'test-results');
    this.artifactsDir = path.join(this.testResultsDir, 'cross-browser-artifacts');
    
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
   * Run cross-browser tests for specific browser
   * @param {string} browser - Browser to test
   * @param {string} testType - Type of tests to run
   * @param {Object} options - Additional options
   */
  async runBrowserTests(browser, testType = TestTypes.ALL, options = {}) {
    console.log(`\n🧪 Running ${testType} tests for ${browser}...`);
    
    const configFile = path.join(this.projectRoot, 'web/tests/config/cross-browser.config.js');
    const testFile = path.join(this.projectRoot, 'web/tests/visual/cross-browser.spec.js');
    
    let command = `npx playwright test "${testFile}" --config="${configFile}"`;
    
    // Add browser-specific project
    if (browser !== Browsers.ALL) {
      const projectMap = {
        [Browsers.CHROME]: 'chrome-desktop',
        [Browsers.FIREFOX]: 'firefox-desktop',
        [Browsers.SAFARI]: 'safari-desktop',
        [Browsers.EDGE]: 'edge-desktop'
      };
      
      const project = projectMap[browser];
      if (project) {
        command += ` --project="${project}"`;
      }
    }
    
    // Add test type filtering
    if (testType !== TestTypes.ALL) {
      const testTypeMap = {
        [TestTypes.VISUAL]: 'visual',
        [TestTypes.FUNCTIONAL]: 'functional',
        [TestTypes.PERFORMANCE]: 'performance'
      };
      
      const grepPattern = testTypeMap[testType];
      if (grepPattern) {
        command += ` --grep="${grepPattern}"`;
      }
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
      
      console.log(`✅ ${browser} ${testType} tests completed successfully`);
      return true;
    } catch (error) {
      console.error(`❌ ${browser} ${testType} tests failed:`, error.message);
      return false;
    }
  }

  /**
   * Run all cross-browser tests
   * @param {Object} options - Test options
   */
  async runAllTests(options = {}) {
    console.log('\n🚀 Starting comprehensive cross-browser testing...');
    
    const browsers = [Browsers.CHROME, Browsers.FIREFOX, Browsers.SAFARI, Browsers.EDGE];
    const results = {
      total: 0,
      passed: 0,
      failed: 0,
      browsers: {}
    };
    
    for (const browser of browsers) {
      console.log(`\n📱 Testing ${browser.toUpperCase()}...`);
      
      const success = await this.runBrowserTests(browser, TestTypes.ALL, options);
      results.browsers[browser] = success;
      results.total++;
      
      if (success) {
        results.passed++;
      } else {
        results.failed++;
      }
    }
    
    this.generateTestReport(results);
    return results;
  }

  /**
   * Run visual regression tests across all browsers
   * @param {Object} options - Test options
   */
  async runVisualRegressionTests(options = {}) {
    console.log('\n🎨 Running visual regression tests across all browsers...');
    
    const browsers = [Browsers.CHROME, Browsers.FIREFOX, Browsers.SAFARI, Browsers.EDGE];
    const results = {
      total: 0,
      passed: 0,
      failed: 0,
      browsers: {}
    };
    
    for (const browser of browsers) {
      console.log(`\n🖼️  Visual testing ${browser.toUpperCase()}...`);
      
      const success = await this.runBrowserTests(browser, TestTypes.VISUAL, options);
      results.browsers[browser] = success;
      results.total++;
      
      if (success) {
        results.passed++;
      } else {
        results.failed++;
      }
    }
    
    this.generateVisualReport(results);
    return results;
  }

  /**
   * Generate comprehensive test report
   * @param {Object} results - Test results
   */
  generateTestReport(results) {
    const reportPath = path.join(this.testResultsDir, 'cross-browser-test-report.json');
    
    const report = {
      timestamp: new Date().toISOString(),
      summary: {
        total: results.total,
        passed: results.passed,
        failed: results.failed,
        successRate: results.total > 0 ? (results.passed / results.total * 100).toFixed(2) : 0
      },
      browsers: results.browsers,
      artifacts: {
        screenshots: this.artifactsDir,
        reports: path.join(this.testResultsDir, 'cross-browser-report')
      }
    };
    
    fs.writeFileSync(reportPath, JSON.stringify(report, null, 2));
    
    console.log('\n📊 Cross-Browser Test Report:');
    console.log(`Total Tests: ${results.total}`);
    console.log(`Passed: ${results.passed}`);
    console.log(`Failed: ${results.failed}`);
    console.log(`Success Rate: ${report.summary.successRate}%`);
    console.log(`\nReport saved to: ${reportPath}`);
  }

  /**
   * Generate visual regression test report
   * @param {Object} results - Test results
   */
  generateVisualReport(results) {
    const reportPath = path.join(this.testResultsDir, 'visual-regression-report.json');
    
    const report = {
      timestamp: new Date().toISOString(),
      type: 'visual-regression',
      summary: {
        total: results.total,
        passed: results.passed,
        failed: results.failed,
        successRate: results.total > 0 ? (results.passed / results.total * 100).toFixed(2) : 0
      },
      browsers: results.browsers,
      visualArtifacts: {
        screenshots: this.artifactsDir,
        baseline: path.join(this.projectRoot, 'web/tests/visual'),
        reports: path.join(this.testResultsDir, 'cross-browser-report')
      }
    };
    
    fs.writeFileSync(reportPath, JSON.stringify(report, null, 2));
    
    console.log('\n🎨 Visual Regression Test Report:');
    console.log(`Total Visual Tests: ${results.total}`);
    console.log(`Passed: ${results.passed}`);
    console.log(`Failed: ${results.failed}`);
    console.log(`Success Rate: ${report.summary.successRate}%`);
    console.log(`\nVisual report saved to: ${reportPath}`);
  }

  /**
   * Install browser dependencies
   */
  installBrowsers() {
    console.log('\n📦 Installing browser dependencies...');
    
    try {
      execSync('npx playwright install', { 
        stdio: 'inherit',
        cwd: this.projectRoot 
      });
      console.log('✅ Browser dependencies installed successfully');
    } catch (error) {
      console.error('❌ Failed to install browser dependencies:', error.message);
      throw error;
    }
  }

  /**
   * Clean test artifacts
   */
  cleanArtifacts() {
    console.log('\n🧹 Cleaning test artifacts...');
    
    try {
      if (fs.existsSync(this.testResultsDir)) {
        fs.rmSync(this.testResultsDir, { recursive: true, force: true });
      }
      this.ensureDirectories();
      console.log('✅ Test artifacts cleaned');
    } catch (error) {
      console.error('❌ Failed to clean artifacts:', error.message);
    }
  }

  /**
   * Show help information
   */
  showHelp() {
    console.log(`
🧪 Cross-Browser Test Runner

Usage: node run-cross-browser-tests.js [command] [options]

Commands:
  all                    Run all cross-browser tests
  visual                 Run visual regression tests only
  browser <name>         Run tests for specific browser (chrome, firefox, safari, edge)
  install                Install browser dependencies
  clean                  Clean test artifacts
  help                   Show this help message

Options:
  --headed               Run tests in headed mode (visible browser)
  --debug                Run tests in debug mode
  --ui                   Run tests with UI mode
  --ci                   Run in CI mode
  --reporter <type>      Specify reporter type

Examples:
  node run-cross-browser-tests.js all
  node run-cross-browser-tests.js visual --headed
  node run-cross-browser-tests.js browser chrome --debug
  node run-cross-browser-tests.js install
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
  
  const runner = new CrossBrowserTestRunner();
  
  try {
    switch (command) {
      case 'all':
        await runner.runAllTests(options);
        break;
      case 'visual':
        await runner.runVisualRegressionTests(options);
        break;
      case 'browser':
        const browser = args[1];
        if (!browser) {
          console.error('❌ Browser name required');
          process.exit(1);
        }
        await runner.runBrowserTests(browser, TestTypes.ALL, options);
        break;
      case 'install':
        runner.installBrowsers();
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

module.exports = { CrossBrowserTestRunner, Browsers, TestTypes };
