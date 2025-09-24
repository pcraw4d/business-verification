// web/tests/config/cross-browser.config.js
const { defineConfig, devices } = require('@playwright/test');

/**
 * Cross-browser specific configuration for visual regression testing
 * This configuration focuses on browser compatibility and rendering consistency
 */
module.exports = defineConfig({
  testDir: './web/tests/visual',
  testMatch: '**/cross-browser.spec.js',
  
  /* Run tests in parallel for faster execution */
  fullyParallel: true,
  
  /* Fail the build on CI if you accidentally left test.only in the source code */
  forbidOnly: !!process.env.CI,
  
  /* Retry on CI only */
  retries: process.env.CI ? 2 : 0,
  
  /* Opt out of parallel tests on CI for stability */
  workers: process.env.CI ? 1 : undefined,
  
  /* Reporter configuration for cross-browser testing */
  reporter: [
    ['html', { outputFolder: 'test-results/cross-browser-report' }],
    ['json', { outputFile: 'test-results/cross-browser-results.json' }],
    ['junit', { outputFile: 'test-results/cross-browser-results.xml' }],
    ['list'] // Console output for debugging
  ],
  
  /* Shared settings for all browser projects */
  use: {
    baseURL: 'https://shimmering-comfort-production.up.railway.app',
    
    /* Enhanced tracing for cross-browser debugging */
    trace: 'on-first-retry',
    
    /* Screenshot on failure for visual debugging */
    screenshot: 'only-on-failure',
    
    /* Video recording for failed tests */
    video: 'retain-on-failure',
    
    /* Extended timeouts for cross-browser testing */
    actionTimeout: 15000,
    navigationTimeout: 45000,
    
    /* Browser-specific settings */
    ignoreHTTPSErrors: true,
    acceptDownloads: true,
  },

  /* Browser-specific project configurations */
  projects: [
    // Chrome/Chromium testing
    {
      name: 'chrome-desktop',
      use: { 
        ...devices['Desktop Chrome'],
        // Chrome-specific settings
        launchOptions: {
          args: [
            '--disable-web-security',
            '--disable-features=VizDisplayCompositor',
            '--disable-background-timer-throttling',
            '--disable-backgrounding-occluded-windows',
            '--disable-renderer-backgrounding'
          ]
        }
      },
    },
    
    {
      name: 'chrome-mobile',
      use: { 
        ...devices['Pixel 5'],
        // Mobile Chrome settings
        launchOptions: {
          args: [
            '--disable-web-security',
            '--disable-features=VizDisplayCompositor'
          ]
        }
      },
    },

    // Firefox testing
    {
      name: 'firefox-desktop',
      use: { 
        ...devices['Desktop Firefox'],
        // Firefox-specific settings
        launchOptions: {
          firefoxUserPrefs: {
            'dom.webnotifications.enabled': false,
            'dom.push.enabled': false,
            'media.navigator.streams.fake': true,
            'media.navigator.permission.disabled': true
          }
        }
      },
    },
    
    {
      name: 'firefox-mobile',
      use: { 
        ...devices['Pixel 5'],
        // Mobile Firefox settings
        launchOptions: {
          firefoxUserPrefs: {
            'dom.webnotifications.enabled': false,
            'dom.push.enabled': false
          }
        }
      },
    },

    // Safari/WebKit testing
    {
      name: 'safari-desktop',
      use: { 
        ...devices['Desktop Safari'],
        // Safari-specific settings
        launchOptions: {
          args: [
            '--disable-web-security',
            '--disable-features=VizDisplayCompositor'
          ]
        }
      },
    },
    
    {
      name: 'safari-mobile',
      use: { 
        ...devices['iPhone 12'],
        // Mobile Safari settings
        launchOptions: {
          args: [
            '--disable-web-security'
          ]
        }
      },
    },

    // Edge testing
    {
      name: 'edge-desktop',
      use: { 
        ...devices['Desktop Edge'],
        channel: 'msedge',
        // Edge-specific settings
        launchOptions: {
          args: [
            '--disable-web-security',
            '--disable-features=VizDisplayCompositor',
            '--disable-background-timer-throttling'
          ]
        }
      },
    },
    
    {
      name: 'edge-mobile',
      use: { 
        ...devices['Pixel 5'],
        channel: 'msedge',
        // Mobile Edge settings
        launchOptions: {
          args: [
            '--disable-web-security'
          ]
        }
      },
    },
  ],

  /* Web server configuration */
  webServer: {
    command: 'python3 -m http.server 8080 --directory web',
    url: 'http://localhost:8080',
    reuseExistingServer: !process.env.CI,
    timeout: 120 * 1000,
  },

  /* Extended timeouts for cross-browser testing */
  timeout: 60 * 1000, // 60 seconds
  expect: {
    timeout: 10000, // 10 seconds for assertions
  },

  /* Output directory for cross-browser test artifacts */
  outputDir: 'test-results/cross-browser-artifacts',

  /* Global setup and teardown */
  globalSetup: require.resolve('../global-setup.js'),
  globalTeardown: require.resolve('../global-teardown.js'),
});
