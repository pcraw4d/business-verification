#!/usr/bin/env node
// web/tests/scripts/run-interactive-element-tests.js

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

/**
 * Interactive element test runner script
 * Provides easy execution of interactive element visual regression tests
 */

const TestTypes = {
  HOVER: 'hover',
  TOOLTIP: 'tooltip',
  ANIMATION: 'animation',
  FOCUS: 'focus',
  RESPONSIVE: 'responsive',
  ACCESSIBILITY: 'accessibility',
  ALL: 'all'
};

const Viewports = {
  MOBILE: 'mobile',
  TABLET: 'tablet',
  DESKTOP: 'desktop',
  ALL: 'all'
};

class InteractiveElementTestRunner {
  constructor() {
    this.projectRoot = path.resolve(__dirname, '../../..');
    this.testResultsDir = path.join(this.projectRoot, 'test-results');
    this.artifactsDir = path.join(this.testResultsDir, 'interactive-element-artifacts');
    
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
   * Run interactive element tests for specific test type
   * @param {string} testType - Type of interactive tests to run
   * @param {string} viewport - Viewport size to test
   * @param {Object} options - Additional options
   */
  async runInteractiveTests(testType = TestTypes.ALL, viewport = Viewports.ALL, options = {}) {
    console.log(`\n🧪 Running ${testType} interactive tests for ${viewport} viewport...`);
    
    const configFile = path.join(this.projectRoot, 'web/tests/config/interactive-element.config.js');
    const testFile = path.join(this.projectRoot, 'web/tests/visual/interactive-element-tests.spec.js');
    
    let command = `npx playwright test "${testFile}" --config="${configFile}"`;
    
    // Add test type filtering
    if (testType !== TestTypes.ALL) {
      const testTypeMap = {
        [TestTypes.HOVER]: 'Hover State Tests',
        [TestTypes.TOOLTIP]: 'Tooltip Tests',
        [TestTypes.ANIMATION]: 'Animation State Tests',
        [TestTypes.FOCUS]: 'Focus State Tests',
        [TestTypes.RESPONSIVE]: 'Responsive Interactive Tests',
        [TestTypes.ACCESSIBILITY]: 'Accessibility Interactive Tests'
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
      
      console.log(`✅ ${testType} interactive tests completed successfully`);
      return true;
    } catch (error) {
      console.error(`❌ ${testType} interactive tests failed:`, error.message);
      return false;
    }
  }

  /**
   * Run all interactive element tests
   * @param {Object} options - Test options
   */
  async runAllInteractiveTests(options = {}) {
    console.log('\n🚀 Starting comprehensive interactive element testing...');
    
    const testTypes = [
      TestTypes.HOVER,
      TestTypes.TOOLTIP,
      TestTypes.ANIMATION,
      TestTypes.FOCUS,
      TestTypes.RESPONSIVE,
      TestTypes.ACCESSIBILITY
    ];
    
    const results = {
      total: 0,
      passed: 0,
      failed: 0,
      testTypes: {}
    };
    
    for (const testType of testTypes) {
      console.log(`\n📱 Testing ${testType.toUpperCase()}...`);
      
      const success = await this.runInteractiveTests(testType, Viewports.ALL, options);
      results.testTypes[testType] = success;
      results.total++;
      
      if (success) {
        results.passed++;
      } else {
        results.failed++;
      }
    }
    
    this.generateInteractiveTestReport(results);
    return results;
  }

  /**
   * Run hover state tests
   * @param {Object} options - Test options
   */
  async runHoverTests(options = {}) {
    console.log('\n🖱️ Running hover state tests...');
    
    const hoverTypes = ['button', 'card', 'navigation', 'form', 'consistency'];
    const results = {
      total: 0,
      passed: 0,
      failed: 0,
      hoverTypes: {}
    };
    
    for (const type of hoverTypes) {
      console.log(`\n🔍 Testing ${type.toUpperCase()} hover state...`);
      
      const success = await this.runInteractiveTests(TestTypes.HOVER, Viewports.ALL, {
        ...options,
        grep: type
      });
      
      results.hoverTypes[type] = success;
      results.total++;
      
      if (success) {
        results.passed++;
      } else {
        results.failed++;
      }
    }
    
    this.generateHoverTestReport(results);
    return results;
  }

  /**
   * Run tooltip tests
   * @param {Object} options - Test options
   */
  async runTooltipTests(options = {}) {
    console.log('\n💬 Running tooltip tests...');
    
    const tooltipTypes = ['risk-indicators', 'form-elements', 'navigation', 'positioning', 'responsive'];
    const results = {
      total: 0,
      passed: 0,
      failed: 0,
      tooltipTypes: {}
    };
    
    for (const type of tooltipTypes) {
      console.log(`\n🔍 Testing ${type.toUpperCase()} tooltip...`);
      
      const success = await this.runInteractiveTests(TestTypes.TOOLTIP, Viewports.ALL, {
        ...options,
        grep: type
      });
      
      results.tooltipTypes[type] = success;
      results.total++;
      
      if (success) {
        results.passed++;
      } else {
        results.failed++;
      }
    }
    
    this.generateTooltipTestReport(results);
    return results;
  }

  /**
   * Run animation tests
   * @param {Object} options - Test options
   */
  async runAnimationTests(options = {}) {
    console.log('\n🎬 Running animation tests...');
    
    const animationTypes = ['risk-transitions', 'loading', 'button-click', 'card-hover', 'form-validation'];
    const results = {
      total: 0,
      passed: 0,
      failed: 0,
      animationTypes: {}
    };
    
    for (const type of animationTypes) {
      console.log(`\n🔍 Testing ${type.toUpperCase()} animation...`);
      
      const success = await this.runInteractiveTests(TestTypes.ANIMATION, Viewports.ALL, {
        ...options,
        grep: type
      });
      
      results.animationTypes[type] = success;
      results.total++;
      
      if (success) {
        results.passed++;
      } else {
        results.failed++;
      }
    }
    
    this.generateAnimationTestReport(results);
    return results;
  }

  /**
   * Run focus tests
   * @param {Object} options - Test options
   */
  async runFocusTests(options = {}) {
    console.log('\n⌨️ Running focus tests...');
    
    const focusTypes = ['form-inputs', 'buttons', 'navigation', 'consistency', 'keyboard-nav', 'focus-trap'];
    const results = {
      total: 0,
      passed: 0,
      failed: 0,
      focusTypes: {}
    };
    
    for (const type of focusTypes) {
      console.log(`\n🔍 Testing ${type.toUpperCase()} focus...`);
      
      const success = await this.runInteractiveTests(TestTypes.FOCUS, Viewports.ALL, {
        ...options,
        grep: type
      });
      
      results.focusTypes[type] = success;
      results.total++;
      
      if (success) {
        results.passed++;
      } else {
        results.failed++;
      }
    }
    
    this.generateFocusTestReport(results);
    return results;
  }

  /**
   * Run responsive interaction tests
   * @param {Object} options - Test options
   */
  async runResponsiveInteractionTests(options = {}) {
    console.log('\n📱 Running responsive interaction tests...');
    
    const viewports = [Viewports.MOBILE, Viewports.TABLET, Viewports.DESKTOP];
    const results = {
      total: 0,
      passed: 0,
      failed: 0,
      viewports: {}
    };
    
    for (const viewport of viewports) {
      console.log(`\n🖥️ Testing ${viewport.toUpperCase()} interactions...`);
      
      const success = await this.runInteractiveTests(TestTypes.RESPONSIVE, viewport, options);
      results.viewports[viewport] = success;
      results.total++;
      
      if (success) {
        results.passed++;
      } else {
        results.failed++;
      }
    }
    
    this.generateResponsiveInteractionReport(results);
    return results;
  }

  /**
   * Run accessibility interaction tests
   * @param {Object} options - Test options
   */
  async runAccessibilityTests(options = {}) {
    console.log('\n♿ Running accessibility interaction tests...');
    
    const a11yTypes = ['keyboard-nav', 'screen-reader', 'high-contrast'];
    const results = {
      total: 0,
      passed: 0,
      failed: 0,
      a11yTypes: {}
    };
    
    for (const type of a11yTypes) {
      console.log(`\n🔍 Testing ${type.toUpperCase()} accessibility...`);
      
      const success = await this.runInteractiveTests(TestTypes.ACCESSIBILITY, Viewports.ALL, {
        ...options,
        grep: type
      });
      
      results.a11yTypes[type] = success;
      results.total++;
      
      if (success) {
        results.passed++;
      } else {
        results.failed++;
      }
    }
    
    this.generateAccessibilityTestReport(results);
    return results;
  }

  /**
   * Generate comprehensive interactive test report
   * @param {Object} results - Test results
   */
  generateInteractiveTestReport(results) {
    const reportPath = path.join(this.testResultsDir, 'interactive-element-test-report.json');
    
    const report = {
      timestamp: new Date().toISOString(),
      type: 'interactive-element-testing',
      summary: {
        total: results.total,
        passed: results.passed,
        failed: results.failed,
        successRate: results.total > 0 ? (results.passed / results.total * 100).toFixed(2) : 0
      },
      testTypes: results.testTypes,
      artifacts: {
        screenshots: this.artifactsDir,
        reports: path.join(this.testResultsDir, 'interactive-element-report')
      }
    };
    
    fs.writeFileSync(reportPath, JSON.stringify(report, null, 2));
    
    console.log('\n📊 Interactive Element Test Report:');
    console.log(`Total Test Types: ${results.total}`);
    console.log(`Passed: ${results.passed}`);
    console.log(`Failed: ${results.failed}`);
    console.log(`Success Rate: ${report.summary.successRate}%`);
    console.log(`\nReport saved to: ${reportPath}`);
  }

  /**
   * Generate hover test report
   * @param {Object} results - Test results
   */
  generateHoverTestReport(results) {
    const reportPath = path.join(this.testResultsDir, 'hover-test-report.json');
    
    const report = {
      timestamp: new Date().toISOString(),
      type: 'hover-testing',
      summary: {
        total: results.total,
        passed: results.passed,
        failed: results.failed,
        successRate: results.total > 0 ? (results.passed / results.total * 100).toFixed(2) : 0
      },
      hoverTypes: results.hoverTypes,
      artifacts: {
        screenshots: this.artifactsDir,
        reports: path.join(this.testResultsDir, 'interactive-element-report')
      }
    };
    
    fs.writeFileSync(reportPath, JSON.stringify(report, null, 2));
    
    console.log('\n🖱️ Hover Test Report:');
    console.log(`Total Hover Types: ${results.total}`);
    console.log(`Passed: ${results.passed}`);
    console.log(`Failed: ${results.failed}`);
    console.log(`Success Rate: ${report.summary.successRate}%`);
    console.log(`\nHover test report saved to: ${reportPath}`);
  }

  /**
   * Generate tooltip test report
   * @param {Object} results - Test results
   */
  generateTooltipTestReport(results) {
    const reportPath = path.join(this.testResultsDir, 'tooltip-test-report.json');
    
    const report = {
      timestamp: new Date().toISOString(),
      type: 'tooltip-testing',
      summary: {
        total: results.total,
        passed: results.passed,
        failed: results.failed,
        successRate: results.total > 0 ? (results.passed / results.total * 100).toFixed(2) : 0
      },
      tooltipTypes: results.tooltipTypes,
      artifacts: {
        screenshots: this.artifactsDir,
        reports: path.join(this.testResultsDir, 'interactive-element-report')
      }
    };
    
    fs.writeFileSync(reportPath, JSON.stringify(report, null, 2));
    
    console.log('\n💬 Tooltip Test Report:');
    console.log(`Total Tooltip Types: ${results.total}`);
    console.log(`Passed: ${results.passed}`);
    console.log(`Failed: ${results.failed}`);
    console.log(`Success Rate: ${report.summary.successRate}%`);
    console.log(`\nTooltip test report saved to: ${reportPath}`);
  }

  /**
   * Generate animation test report
   * @param {Object} results - Test results
   */
  generateAnimationTestReport(results) {
    const reportPath = path.join(this.testResultsDir, 'animation-test-report.json');
    
    const report = {
      timestamp: new Date().toISOString(),
      type: 'animation-testing',
      summary: {
        total: results.total,
        passed: results.passed,
        failed: results.failed,
        successRate: results.total > 0 ? (results.passed / results.total * 100).toFixed(2) : 0
      },
      animationTypes: results.animationTypes,
      artifacts: {
        screenshots: this.artifactsDir,
        reports: path.join(this.testResultsDir, 'interactive-element-report')
      }
    };
    
    fs.writeFileSync(reportPath, JSON.stringify(report, null, 2));
    
    console.log('\n🎬 Animation Test Report:');
    console.log(`Total Animation Types: ${results.total}`);
    console.log(`Passed: ${results.passed}`);
    console.log(`Failed: ${results.failed}`);
    console.log(`Success Rate: ${report.summary.successRate}%`);
    console.log(`\nAnimation test report saved to: ${reportPath}`);
  }

  /**
   * Generate focus test report
   * @param {Object} results - Test results
   */
  generateFocusTestReport(results) {
    const reportPath = path.join(this.testResultsDir, 'focus-test-report.json');
    
    const report = {
      timestamp: new Date().toISOString(),
      type: 'focus-testing',
      summary: {
        total: results.total,
        passed: results.passed,
        failed: results.failed,
        successRate: results.total > 0 ? (results.passed / results.total * 100).toFixed(2) : 0
      },
      focusTypes: results.focusTypes,
      artifacts: {
        screenshots: this.artifactsDir,
        reports: path.join(this.testResultsDir, 'interactive-element-report')
      }
    };
    
    fs.writeFileSync(reportPath, JSON.stringify(report, null, 2));
    
    console.log('\n⌨️ Focus Test Report:');
    console.log(`Total Focus Types: ${results.total}`);
    console.log(`Passed: ${results.passed}`);
    console.log(`Failed: ${results.failed}`);
    console.log(`Success Rate: ${report.summary.successRate}%`);
    console.log(`\nFocus test report saved to: ${reportPath}`);
  }

  /**
   * Generate responsive interaction test report
   * @param {Object} results - Test results
   */
  generateResponsiveInteractionReport(results) {
    const reportPath = path.join(this.testResultsDir, 'responsive-interaction-test-report.json');
    
    const report = {
      timestamp: new Date().toISOString(),
      type: 'responsive-interaction-testing',
      summary: {
        total: results.total,
        passed: results.passed,
        failed: results.failed,
        successRate: results.total > 0 ? (results.passed / results.total * 100).toFixed(2) : 0
      },
      viewports: results.viewports,
      artifacts: {
        screenshots: this.artifactsDir,
        reports: path.join(this.testResultsDir, 'interactive-element-report')
      }
    };
    
    fs.writeFileSync(reportPath, JSON.stringify(report, null, 2));
    
    console.log('\n📱 Responsive Interaction Test Report:');
    console.log(`Total Viewports: ${results.total}`);
    console.log(`Passed: ${results.passed}`);
    console.log(`Failed: ${results.failed}`);
    console.log(`Success Rate: ${report.summary.successRate}%`);
    console.log(`\nResponsive interaction test report saved to: ${reportPath}`);
  }

  /**
   * Generate accessibility test report
   * @param {Object} results - Test results
   */
  generateAccessibilityTestReport(results) {
    const reportPath = path.join(this.testResultsDir, 'accessibility-test-report.json');
    
    const report = {
      timestamp: new Date().toISOString(),
      type: 'accessibility-testing',
      summary: {
        total: results.total,
        passed: results.passed,
        failed: results.failed,
        successRate: results.total > 0 ? (results.passed / results.total * 100).toFixed(2) : 0
      },
      a11yTypes: results.a11yTypes,
      artifacts: {
        screenshots: this.artifactsDir,
        reports: path.join(this.testResultsDir, 'interactive-element-report')
      }
    };
    
    fs.writeFileSync(reportPath, JSON.stringify(report, null, 2));
    
    console.log('\n♿ Accessibility Test Report:');
    console.log(`Total A11y Types: ${results.total}`);
    console.log(`Passed: ${results.passed}`);
    console.log(`Failed: ${results.failed}`);
    console.log(`Success Rate: ${report.summary.successRate}%`);
    console.log(`\nAccessibility test report saved to: ${reportPath}`);
  }

  /**
   * Clean test artifacts
   */
  cleanArtifacts() {
    console.log('\n🧹 Cleaning interactive element test artifacts...');
    
    try {
      if (fs.existsSync(this.testResultsDir)) {
        fs.rmSync(this.testResultsDir, { recursive: true, force: true });
      }
      this.ensureDirectories();
      console.log('✅ Interactive element test artifacts cleaned');
    } catch (error) {
      console.error('❌ Failed to clean artifacts:', error.message);
    }
  }

  /**
   * Show help information
   */
  showHelp() {
    console.log(`
🧪 Interactive Element Test Runner

Usage: node run-interactive-element-tests.js [command] [options]

Commands:
  all                    Run all interactive element tests
  hover                  Run hover state tests
  tooltip                Run tooltip tests
  animation              Run animation state tests
  focus                  Run focus state tests
  responsive             Run responsive interaction tests
  accessibility          Run accessibility interaction tests
  clean                  Clean test artifacts
  help                   Show this help message

Options:
  --headed               Run tests in headed mode (visible browser)
  --debug                Run tests in debug mode
  --ui                   Run tests with UI mode
  --ci                   Run in CI mode
  --reporter <type>      Specify reporter type

Examples:
  node run-interactive-element-tests.js all
  node run-interactive-element-tests.js hover --headed
  node run-interactive-element-tests.js tooltip --debug
  node run-interactive-element-tests.js animation --ui
  node run-interactive-element-tests.js focus --headed
  node run-interactive-element-tests.js clean
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
  
  const runner = new InteractiveElementTestRunner();
  
  try {
    switch (command) {
      case 'all':
        await runner.runAllInteractiveTests(options);
        break;
      case 'hover':
        await runner.runHoverTests(options);
        break;
      case 'tooltip':
        await runner.runTooltipTests(options);
        break;
      case 'animation':
        await runner.runAnimationTests(options);
        break;
      case 'focus':
        await runner.runFocusTests(options);
        break;
      case 'responsive':
        await runner.runResponsiveInteractionTests(options);
        break;
      case 'accessibility':
        await runner.runAccessibilityTests(options);
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

module.exports = { InteractiveElementTestRunner, TestTypes, Viewports };
