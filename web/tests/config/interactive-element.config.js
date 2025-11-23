// web/tests/config/interactive-element.config.js
const { defineConfig, devices } = require('@playwright/test');

/**
 * Interactive element testing configuration
 * Focuses on testing user interactions, animations, and accessibility
 */
module.exports = defineConfig({
  testDir: './web/tests/visual',
  testMatch: '**/interactive-element-tests.spec.js',
  
  /* Run tests in parallel for faster execution */
  fullyParallel: true,
  
  /* Fail the build on CI if you accidentally left test.only in the source code */
  forbidOnly: !!process.env.CI,
  
  /* Retry on CI only */
  retries: process.env.CI ? 2 : 0,
  
  /* Opt out of parallel tests on CI for stability */
  workers: process.env.CI ? 1 : undefined,
  
  /* Reporter configuration for interactive element testing */
  reporter: [
    ['html', { outputFolder: 'test-results/interactive-element-report' }],
    ['json', { outputFile: 'test-results/interactive-element-results.json' }],
    ['junit', { outputFile: 'test-results/interactive-element-results.xml' }],
    ['list'] // Console output for debugging
  ],
  
  /* Shared settings for all interactive element tests */
  use: {
    baseURL: 'https://creative-determination-production.up.railway.app',
    
    /* Enhanced tracing for interaction debugging */
    trace: 'on-first-retry',
    
    /* Screenshot on failure for visual debugging */
    screenshot: 'only-on-failure',
    
    /* Video recording for failed tests */
    video: 'retain-on-failure',
    
    /* Extended timeouts for interactions and animations */
    actionTimeout: 10000,
    navigationTimeout: 30000,
    
    /* Browser settings optimized for interaction testing */
    ignoreHTTPSErrors: true,
    acceptDownloads: true,
  },

  /* Browser configurations for interactive element testing */
  projects: [
    // Desktop Chrome for interaction testing
    {
      name: 'chrome-desktop',
      use: { 
        ...devices['Desktop Chrome'],
        // Chrome-specific settings for interaction testing
        launchOptions: {
          args: [
            '--disable-web-security',
            '--disable-features=VizDisplayCompositor',
            '--disable-background-timer-throttling',
            '--disable-backgrounding-occluded-windows',
            '--disable-renderer-backgrounding',
            '--enable-precise-memory-info',
            '--enable-gpu-rasterization'
          ]
        }
      },
    },
    
    // Mobile Chrome for touch interaction testing
    {
      name: 'chrome-mobile',
      use: { 
        ...devices['Pixel 5'],
        // Mobile Chrome settings for touch testing
        launchOptions: {
          args: [
            '--disable-web-security',
            '--disable-features=VizDisplayCompositor',
            '--enable-touch-events'
          ]
        }
      },
    },

    // Firefox for cross-browser interaction testing
    {
      name: 'firefox-desktop',
      use: { 
        ...devices['Desktop Firefox'],
        // Firefox-specific settings for interaction testing
        launchOptions: {
          firefoxUserPrefs: {
            'dom.webnotifications.enabled': false,
            'dom.push.enabled': false,
            'media.navigator.streams.fake': true,
            'media.navigator.permission.disabled': true,
            'layout.css.grid.enabled': true,
            'layout.css.flexbox.enabled': true,
            'gfx.webrender.all': true,
            'gfx.webrender.enabled': true
          }
        }
      },
    },
    
    // Safari for WebKit interaction testing
    {
      name: 'safari-desktop',
      use: { 
        ...devices['Desktop Safari'],
        // Safari-specific settings for interaction testing
        launchOptions: {
          args: [
            '--disable-web-security',
            '--disable-features=VizDisplayCompositor',
            '--enable-touch-events'
          ]
        }
      },
    },

    // Edge for Microsoft browser interaction testing
    {
      name: 'edge-desktop',
      use: { 
        ...devices['Desktop Edge'],
        channel: 'msedge',
        // Edge-specific settings for interaction testing
        launchOptions: {
          args: [
            '--disable-web-security',
            '--disable-features=VizDisplayCompositor',
            '--disable-background-timer-throttling',
            '--enable-gpu-rasterization'
          ]
        }
      },
    },
  ],

  /* Web server configuration - Using Railway deployment */
  // webServer: {
  //   command: 'python3 -m http.server 8080 --directory web',
  //   url: 'http://localhost:8080',
  //   reuseExistingServer: !process.env.CI,
  //   timeout: 120 * 1000,
  // },

  /* Extended timeouts for interactive element testing */
  timeout: 60 * 1000, // 60 seconds for complex interactions
  expect: {
    timeout: 10000, // 10 seconds for interaction assertions
  },

  /* Output directory for interactive element test artifacts */
  outputDir: 'test-results/interactive-element-artifacts',

  /* Global setup and teardown */
  globalSetup: require.resolve('../global-setup.js'),
  globalTeardown: require.resolve('../global-teardown.js'),
});
