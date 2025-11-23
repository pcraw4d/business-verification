// web/tests/config/state-based.config.js
const { defineConfig, devices } = require('@playwright/test');

/**
 * State-based visual testing configuration
 * Focuses on testing different application states and user interactions
 */
module.exports = defineConfig({
  testDir: './web/tests/visual',
  testMatch: '**/state-based-tests.spec.js',
  
  /* Run tests in parallel for faster execution */
  fullyParallel: true,
  
  /* Fail the build on CI if you accidentally left test.only in the source code */
  forbidOnly: !!process.env.CI,
  
  /* Retry on CI only */
  retries: process.env.CI ? 2 : 0,
  
  /* Opt out of parallel tests on CI for stability */
  workers: process.env.CI ? 1 : undefined,
  
  /* Reporter configuration for state-based testing */
  reporter: [
    ['html', { outputFolder: 'test-results/state-based-report' }],
    ['json', { outputFile: 'test-results/state-based-results.json' }],
    ['junit', { outputFile: 'test-results/state-based-results.xml' }],
    ['list'] // Console output for debugging
  ],
  
  /* Shared settings for all state-based tests */
  use: {
    baseURL: 'https://creative-determination-production.up.railway.app',
    
    /* Enhanced tracing for state debugging */
    trace: 'on-first-retry',
    
    /* Screenshot on failure for visual debugging */
    screenshot: 'only-on-failure',
    
    /* Video recording for failed tests */
    video: 'retain-on-failure',
    
    /* Extended timeouts for state transitions */
    actionTimeout: 10000,
    navigationTimeout: 30000,
    
    /* Browser settings optimized for state testing */
    ignoreHTTPSErrors: true,
    acceptDownloads: true,
  },

  /* Browser configurations for state testing */
  projects: [
    // Desktop Chrome for state testing
    {
      name: 'chrome-desktop',
      use: { 
        ...devices['Desktop Chrome'],
        // Chrome-specific settings for state testing
        launchOptions: {
          args: [
            '--disable-web-security',
            '--disable-features=VizDisplayCompositor',
            '--disable-background-timer-throttling',
            '--disable-backgrounding-occluded-windows',
            '--disable-renderer-backgrounding',
            '--enable-precise-memory-info'
          ]
        }
      },
    },
    
    // Mobile Chrome for responsive state testing
    {
      name: 'chrome-mobile',
      use: { 
        ...devices['Pixel 5'],
        // Mobile Chrome settings for state testing
        launchOptions: {
          args: [
            '--disable-web-security',
            '--disable-features=VizDisplayCompositor'
          ]
        }
      },
    },

    // Firefox for cross-browser state testing
    {
      name: 'firefox-desktop',
      use: { 
        ...devices['Desktop Firefox'],
        // Firefox-specific settings for state testing
        launchOptions: {
          firefoxUserPrefs: {
            'dom.webnotifications.enabled': false,
            'dom.push.enabled': false,
            'media.navigator.streams.fake': true,
            'media.navigator.permission.disabled': true,
            'layout.css.grid.enabled': true,
            'layout.css.flexbox.enabled': true
          }
        }
      },
    },
    
    // Safari for WebKit state testing
    {
      name: 'safari-desktop',
      use: { 
        ...devices['Desktop Safari'],
        // Safari-specific settings for state testing
        launchOptions: {
          args: [
            '--disable-web-security',
            '--disable-features=VizDisplayCompositor'
          ]
        }
      },
    },

    // Edge for Microsoft browser state testing
    {
      name: 'edge-desktop',
      use: { 
        ...devices['Desktop Edge'],
        channel: 'msedge',
        // Edge-specific settings for state testing
        launchOptions: {
          args: [
            '--disable-web-security',
            '--disable-features=VizDisplayCompositor',
            '--disable-background-timer-throttling'
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

  /* Extended timeouts for state-based testing */
  timeout: 45 * 1000, // 45 seconds for state transitions
  expect: {
    timeout: 8000, // 8 seconds for state assertions
  },

  /* Output directory for state-based test artifacts */
  outputDir: 'test-results/state-based-artifacts',

  /* Global setup and teardown */
  globalSetup: require.resolve('../global-setup.js'),
  globalTeardown: require.resolve('../global-teardown.js'),
});
