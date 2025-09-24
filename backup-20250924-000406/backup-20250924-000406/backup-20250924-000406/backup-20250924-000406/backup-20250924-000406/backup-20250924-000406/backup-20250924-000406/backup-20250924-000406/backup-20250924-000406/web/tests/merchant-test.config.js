// @ts-check
const { defineConfig, devices } = require('@playwright/test');

/**
 * Merchant Test Configuration
 * 
 * This configuration is specifically for testing the merchant-centric UI
 * components and functionality of the KYB Platform.
 */

module.exports = defineConfig({
  testDir: './web/tests',
  
  // Test file patterns
  testMatch: [
    'merchant-*.spec.js'
  ],
  
  // Test timeout
  timeout: 30 * 1000,
  
  // Expect timeout
  expect: {
    timeout: 10 * 1000,
  },
  
  // Fail the build on CI if you accidentally left test.only in the source code
  forbidOnly: !!process.env.CI,
  
  // Retry on CI only
  retries: process.env.CI ? 2 : 0,
  
  // Opt out of parallel tests on CI
  workers: process.env.CI ? 1 : undefined,
  
  // Reporter configuration
  reporter: [
    ['html', { outputFolder: 'test-results/merchant-reports/html' }],
    ['json', { outputFile: 'test-results/merchant-reports/results.json' }],
    ['junit', { outputFile: 'test-results/merchant-reports/results.xml' }],
    ['list']
  ],
  
  // Global test options
  use: {
    // Base URL for merchant tests
    baseURL: 'http://localhost:8080',
    
    // Browser context options
    viewport: { width: 1280, height: 720 },
    ignoreHTTPSErrors: true,
    
    // Trace and screenshot options
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    
    // Timeouts
    actionTimeout: 10 * 1000,
    navigationTimeout: 30 * 1000,
    
    // Additional context options
    locale: 'en-US',
    timezoneId: 'America/New_York',
  },
  
  // Test projects for different browsers
  projects: [
    {
      name: 'merchant-chromium',
      use: { 
        ...devices['Desktop Chrome'],
        // Additional options for merchant tests
        launchOptions: {
          args: ['--disable-web-security', '--disable-features=VizDisplayCompositor']
        }
      },
    },
    
    {
      name: 'merchant-firefox',
      use: { 
        ...devices['Desktop Firefox'],
        // Additional options for merchant tests
        launchOptions: {
          firefoxUserPrefs: {
            'dom.webnotifications.enabled': false,
            'dom.push.enabled': false
          }
        }
      },
    },
    
    {
      name: 'merchant-webkit',
      use: { 
        ...devices['Desktop Safari'],
        // Additional options for merchant tests
        launchOptions: {
          args: ['--disable-web-security']
        }
      },
    },
    
    // Mobile testing
    {
      name: 'merchant-mobile-chrome',
      use: { 
        ...devices['Pixel 5'],
        // Additional options for mobile merchant tests
        viewport: { width: 375, height: 667 },
      },
    },
    
    {
      name: 'merchant-mobile-safari',
      use: { 
        ...devices['iPhone 12'],
        // Additional options for mobile merchant tests
        viewport: { width: 390, height: 844 },
      },
    },
  ],
  
  // Web server configuration
  webServer: {
    command: 'python3 -m http.server 8080 --directory web',
    url: 'http://localhost:8080',
    reuseExistingServer: !process.env.CI,
    timeout: 120 * 1000,
    // Additional server options
    env: {
      NODE_ENV: 'test',
      TEST_MODE: 'true'
    }
  },
  
  // Global setup and teardown
  globalSetup: require.resolve('./global-setup.js'),
  globalTeardown: require.resolve('./global-teardown.js'),
  
  // Output directory for test artifacts
  outputDir: 'test-results/merchant-artifacts',
  
  // Test metadata
  metadata: {
    testType: 'merchant-ui',
    version: '1.0.0',
    description: 'Merchant-centric UI tests for KYB Platform'
  }
});
