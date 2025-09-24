# Merchant Playwright Tests

This directory contains comprehensive Playwright tests for the merchant-centric UI implementation of the KYB Platform.

## Test Structure

### Test Files

- **`merchant-portfolio.spec.js`** - Tests for merchant portfolio functionality
  - Merchant list display and pagination
  - Search and filtering capabilities
  - Portfolio type and risk level filtering
  - Bulk selection and operations
  - Export functionality
  - Responsive design testing

- **`merchant-detail.spec.js`** - Tests for merchant detail views
  - Merchant information display
  - Portfolio information
  - Risk assessment display
  - Compliance information
  - Transaction history
  - Audit log display
  - Edit functionality
  - Navigation and breadcrumbs

- **`merchant-bulk-operations.spec.js`** - Tests for bulk operations
  - Merchant selection and deselection
  - Bulk portfolio type updates
  - Bulk risk level updates
  - Bulk export functionality
  - Progress tracking
  - Pause/resume functionality
  - Error handling
  - Operation history

- **`merchant-comparison.spec.js`** - Tests for merchant comparison
  - Merchant selection for comparison
  - Side-by-side comparison display
  - Basic information comparison
  - Portfolio comparison
  - Risk assessment comparison
  - Compliance comparison
  - Difference highlighting
  - Export functionality
  - Responsive design

- **`merchant-hub-integration.spec.js`** - Tests for hub integration
  - Navigation integration
  - Merchant context switching
  - Dashboard content updates
  - Session management
  - Breadcrumb navigation
  - Real-time updates
  - Error handling
  - Responsive design

### Configuration Files

- **`merchant-test.config.js`** - Playwright configuration for merchant tests
- **`run-merchant-tests.js`** - Test runner script with reporting
- **`utils/merchant-test-helpers.js`** - Common test utilities and helpers

## Running Tests

### Prerequisites

1. Install Playwright:
   ```bash
   npm install @playwright/test
   npx playwright install
   ```

2. Start the web server:
   ```bash
   python3 -m http.server 8080 --directory web
   ```

### Running Individual Test Files

```bash
# Run merchant portfolio tests
npx playwright test merchant-portfolio.spec.js

# Run merchant detail tests
npx playwright test merchant-detail.spec.js

# Run bulk operations tests
npx playwright test merchant-bulk-operations.spec.js

# Run comparison tests
npx playwright test merchant-comparison.spec.js

# Run hub integration tests
npx playwright test merchant-hub-integration.spec.js
```

### Running All Merchant Tests

```bash
# Run all merchant tests
npx playwright test --config=merchant-test.config.js

# Run with custom configuration
npx playwright test --config=merchant-test.config.js --project=merchant-chromium
```

### Using the Test Runner

```bash
# Run the comprehensive test runner
node run-merchant-tests.js

# Run with specific options
node run-merchant-tests.js --environment=chromium --mode=headless
```

## Test Configuration

### Browser Support

- **Desktop Browsers**: Chrome, Firefox, Safari
- **Mobile Browsers**: Chrome (Android), Safari (iOS)
- **Viewports**: Desktop (1280x720), Tablet (768x1024), Mobile (375x667)

### Test Features

- **Parallel Execution**: Tests run in parallel for faster execution
- **Retry Logic**: Failed tests are retried up to 2 times
- **Screenshots**: Screenshots are taken on test failures
- **Videos**: Videos are recorded for failed tests
- **Traces**: Traces are collected for debugging failed tests
- **Reports**: HTML, JSON, and JUnit reports are generated

### Test Data

Tests use mock data that is automatically generated and seeded into the test environment. The mock data includes:

- 5000+ realistic merchant records
- Diverse business types and industries
- Various portfolio types (onboarded, deactivated, prospective, pending)
- Different risk levels (high, medium, low)
- Realistic business information (names, addresses, contact details)

## Test Utilities

### MerchantTestHelpers Class

The `MerchantTestHelpers` class provides common utilities for merchant tests:

```javascript
const MerchantTestHelpers = require('./utils/merchant-test-helpers');

test('example test', async ({ page }) => {
  const helpers = new MerchantTestHelpers(page);
  
  // Navigate to merchant portfolio
  await helpers.navigateToMerchantPortfolio();
  
  // Select merchants for bulk operations
  await helpers.selectMerchantsForBulkOperation(2);
  
  // Perform bulk update
  await helpers.performBulkPortfolioTypeUpdate('pending');
  
  // Verify results
  const count = await helpers.getMerchantCount();
  expect(count).toBeGreaterThan(0);
});
```

### Common Helper Methods

- **Navigation**: `navigateToMerchantPortfolio()`, `navigateToMerchantDetail()`, etc.
- **Selection**: `selectMerchant()`, `selectMerchantsForBulkOperation()`, etc.
- **Filtering**: `filterByPortfolioType()`, `filterByRiskLevel()`, etc.
- **Operations**: `performBulkPortfolioTypeUpdate()`, `exportMerchantData()`, etc.
- **Utilities**: `waitForLoadingToComplete()`, `takeScreenshot()`, etc.

## Test Reports

### Report Types

1. **HTML Report**: Interactive HTML report with test results, screenshots, and videos
2. **JSON Report**: Machine-readable JSON report for CI/CD integration
3. **JUnit Report**: XML report compatible with CI/CD systems
4. **Console Output**: Real-time test execution output

### Report Locations

- **HTML Reports**: `test-results/merchant-reports/html/`
- **JSON Reports**: `test-results/merchant-reports/results.json`
- **JUnit Reports**: `test-results/merchant-reports/results.xml`
- **Artifacts**: `test-results/merchant-artifacts/`

## Best Practices

### Test Organization

1. **Group Related Tests**: Use `test.describe()` to group related tests
2. **Clear Test Names**: Use descriptive test names that explain what is being tested
3. **Setup and Teardown**: Use `test.beforeEach()` and `test.afterEach()` for setup
4. **Data Test IDs**: Use consistent `data-testid` attributes for reliable element selection

### Test Reliability

1. **Wait for Elements**: Always wait for elements to be visible before interacting
2. **Use Timeouts**: Set appropriate timeouts for different operations
3. **Handle Async Operations**: Properly handle async operations and loading states
4. **Mock External Dependencies**: Mock external APIs and services for consistent testing

### Test Maintenance

1. **Regular Updates**: Update tests when UI changes
2. **Refactor Common Code**: Extract common test logic into helper functions
3. **Document Changes**: Document any changes to test structure or requirements
4. **Monitor Performance**: Monitor test execution time and optimize as needed

## Troubleshooting

### Common Issues

1. **Element Not Found**: Check if `data-testid` attributes are present and correct
2. **Timeout Errors**: Increase timeout values for slow operations
3. **Network Errors**: Ensure web server is running and accessible
4. **Browser Issues**: Update browser drivers and check browser compatibility

### Debug Mode

Run tests in debug mode for step-by-step execution:

```bash
npx playwright test --debug merchant-portfolio.spec.js
```

### Trace Viewer

View detailed traces of test execution:

```bash
npx playwright show-trace test-results/merchant-artifacts/trace.zip
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Merchant Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-node@v2
        with:
          node-version: '18'
      - run: npm install
      - run: npx playwright install
      - run: npx playwright test --config=merchant-test.config.js
      - uses: actions/upload-artifact@v2
        if: failure()
        with:
          name: playwright-report
          path: test-results/merchant-reports/
```

### Docker

```dockerfile
FROM mcr.microsoft.com/playwright:v1.40.0-focal
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
CMD ["npx", "playwright", "test", "--config=merchant-test.config.js"]
```

## Contributing

### Adding New Tests

1. Create new test file following naming convention: `merchant-*.spec.js`
2. Use existing test structure and helper functions
3. Add appropriate `data-testid` attributes to HTML elements
4. Update this documentation with new test information

### Test Review

1. Ensure tests cover all user interactions
2. Verify tests are reliable and don't flake
3. Check that tests follow established patterns
4. Validate that tests provide meaningful feedback on failures

## Support

For questions or issues with merchant tests:

1. Check the troubleshooting section above
2. Review existing test implementations
3. Consult Playwright documentation
4. Create an issue with detailed error information
