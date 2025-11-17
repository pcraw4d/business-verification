# Testing Setup Guide

## Prerequisites

Before running E2E tests, you need to install Playwright browsers:

```bash
# Install all browsers (recommended for full test coverage)
npx playwright install

# Or install only Chromium for faster setup
npx playwright install chromium
```

## Running Tests

```bash
# Run all E2E tests
npm run test:e2e

# Run specific test file
npx playwright test tests/e2e/risk-assessment.spec.ts

# Run tests in headed mode (see browser)
npx playwright test --headed

# Run tests in specific browser
npx playwright test --project=chromium
```

## Troubleshooting

### Error: Executable doesn't exist
If you see an error about missing browser executables, run:
```bash
npx playwright install
```

### Tests fail with timeout
- Ensure the dev server is running or will auto-start
- Check that the base URL is correct (default: http://localhost:3000)
- Increase timeout in test if needed

### Browser-specific issues
- Firefox/WebKit tests are commented out by default
- Uncomment in `playwright.config.ts` after installing those browsers
- Run `npx playwright install firefox` or `npx playwright install webkit`

