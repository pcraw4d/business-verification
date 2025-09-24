# Visual Regression Testing Guide

## Overview

This guide provides comprehensive documentation for the visual regression testing framework implemented for the KYB Platform customer-facing UI components.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Test Types](#test-types)
3. [Running Tests](#running-tests)
4. [Baseline Management](#baseline-management)
5. [CI/CD Integration](#cicd-integration)
6. [Troubleshooting](#troubleshooting)
7. [Best Practices](#best-practices)

## Getting Started

### Prerequisites

- Node.js 18+
- npm or yarn
- Playwright installed

### Installation

```bash
cd web
npm install
npx playwright install
```

### Quick Start

```bash
# Run all visual tests
npm run test:visual

# Run specific test type
npm run test:baseline
npm run test:interactive
npm run test:state-based
npm run test:cross-browser
npm run test:responsive
```

## Test Types

### 1. Baseline Screenshots
**File**: `tests/visual/baseline-screenshots.spec.js`

Captures baseline screenshots for all pages and components to establish visual regression baselines.

**Features**:
- Full-page screenshots
- Component-level screenshots
- Multiple viewport sizes
- Different risk states

### 2. Interactive Element Tests
**File**: `tests/visual/interactive-element-tests.spec.js`

Tests interactive elements like hover states, tooltips, animations, and focus states.

**Features**:
- Hover state testing
- Tooltip visibility testing
- Animation state testing
- Focus state testing
- Responsive interaction testing
- Accessibility interaction testing

### 3. State-Based Tests
**File**: `tests/visual/state-based-tests.spec.js`

Tests different application states and data scenarios.

**Features**:
- Risk level states (Low/Medium/High/Critical)
- Loading states
- Error states
- Empty data states
- Success states

### 4. Cross-Browser Tests
**File**: `tests/visual/cross-browser.spec.js`

Ensures visual consistency across different browsers.

**Features**:
- Chrome testing
- Firefox testing
- Safari testing (if available)
- Edge testing
- Browser-specific visual validation

### 5. Responsive Design Tests
**File**: `tests/visual/responsive-design.spec.js`

Validates responsive design across different screen sizes.

**Features**:
- Mobile viewport testing (375x667)
- Tablet viewport testing (768x1024)
- Desktop viewport testing (1920x1080)
- Large screen testing (2560x1440)
- Responsive layout validation

## Running Tests

### Local Development

```bash
# Run all tests
npx playwright test

# Run specific test file
npx playwright test tests/visual/baseline-screenshots.spec.js

# Run with specific browser
npx playwright test --project=chromium

# Run in headed mode (see browser)
npx playwright test --headed

# Run in debug mode
npx playwright test --debug
```

### NPM Scripts

```bash
# Visual regression tests
npm run test:visual

# Individual test types
npm run test:baseline
npm run test:interactive
npm run test:state-based
npm run test:cross-browser
npm run test:responsive

# Interactive test categories
npm run test:interactive:hover
npm run test:interactive:tooltip
npm run test:interactive:animation
npm run test:interactive:focus
npm run test:interactive:responsive
npm run test:interactive:accessibility

# Test modes
npm run test:interactive:headed
npm run test:interactive:debug
npm run test:interactive:ui
```

### Custom Test Runner

```bash
# Run interactive element tests with custom runner
node web/tests/scripts/run-interactive-element-tests.js all
node web/tests/scripts/run-interactive-element-tests.js hover
node web/tests/scripts/run-interactive-element-tests.js tooltip
```

## Baseline Management

### Updating Baselines

When UI changes are intentional and you want to update the baseline screenshots:

```bash
# Update all baselines
npx playwright test --update-snapshots

# Update specific test baselines
npx playwright test tests/visual/baseline-screenshots.spec.js --update-snapshots
```

### Baseline Update Workflow

1. **Local Development**: Update baselines locally during development
2. **Review Changes**: Review the updated screenshots to ensure they're correct
3. **Commit Changes**: Commit the updated baseline files
4. **CI/CD**: The CI/CD pipeline will use the new baselines for future tests

### Baseline File Structure

```
web/tests/visual/
├── baseline-screenshots.spec.js-snapshots/
│   ├── risk-dashboard-page-full-darwin.png
│   ├── enhanced-risk-indicators-page-full-darwin.png
│   └── ...
├── interactive-element-tests.spec.js-snapshots/
│   ├── interactive-button-hover-primary-darwin.png
│   ├── interactive-tooltip-hover-darwin.png
│   └── ...
└── ...
```

## CI/CD Integration

### GitHub Actions Workflows

#### 1. Visual Regression Tests (`visual-regression-tests.yml`)

**Triggers**:
- Push to main/develop branches
- Pull requests to main/develop branches
- Changes to web files, package.json, or Playwright config

**Features**:
- Matrix strategy for different test types
- Artifact storage for screenshots and reports
- PR comment integration for visual diffs
- Automatic baseline updates on main branch

#### 2. Update Baselines (`update-baselines.yml`)

**Triggers**:
- Manual workflow dispatch
- Allows selective baseline updates

**Features**:
- Choose specific test type to update
- Automatic commit of updated baselines
- Summary generation

### Artifact Storage

The CI/CD pipeline stores the following artifacts:

- **Playwright Reports**: HTML reports for each test run
- **Test Screenshots**: Screenshots from test execution
- **Baseline Screenshots**: Current baseline screenshots
- **Visual Diff Reports**: Generated diff analysis

### PR Integration

When a pull request is created:

1. Visual regression tests run automatically
2. Screenshots are captured and stored as artifacts
3. A PR comment is created with visual diff analysis
4. Reviewers can download artifacts to review changes

## Troubleshooting

### Common Issues

#### 1. Test Timeouts

**Problem**: Tests fail with timeout errors

**Solutions**:
- Increase timeout in test configuration
- Check if the application is responding
- Verify Railway deployment is accessible
- Check network connectivity

```javascript
// Increase timeout in test
test('my test', async ({ page }) => {
  await page.goto('/my-page', { timeout: 30000 });
});
```

#### 2. Element Not Found

**Problem**: Tests fail because elements are not found

**Solutions**:
- Check if selectors are correct
- Verify elements are visible and stable
- Add proper waits for dynamic content
- Check if elements exist in the current page state

```javascript
// Wait for element to be visible
await page.waitForSelector('.my-element', { state: 'visible' });
```

#### 3. Screenshot Mismatches

**Problem**: Screenshots don't match baselines

**Solutions**:
- Review the actual vs expected screenshots
- Check if changes are intentional
- Update baselines if changes are correct
- Investigate if there are timing issues

#### 4. Railway Connection Issues

**Problem**: Tests fail to connect to Railway deployment

**Solutions**:
- Verify Railway deployment is running
- Check if the URL is correct
- Verify network connectivity
- Check if the application is accessible

### Debug Mode

Run tests in debug mode to step through issues:

```bash
npx playwright test --debug
```

This opens the Playwright Inspector where you can:
- Step through test execution
- Inspect page state
- Debug element selection
- View console logs

### Test Reports

View detailed test reports:

```bash
npx playwright show-report
```

This opens an HTML report with:
- Test results and status
- Screenshots and videos
- Error details and stack traces
- Performance metrics

## Best Practices

### 1. Test Organization

- Group related tests in the same file
- Use descriptive test names
- Keep tests focused and atomic
- Avoid test interdependencies

### 2. Selector Strategy

- Use stable, semantic selectors
- Prefer data-testid attributes
- Avoid brittle CSS selectors
- Use role-based selectors for accessibility

### 3. Timing and Waits

- Use explicit waits instead of fixed timeouts
- Wait for network requests to complete
- Wait for animations to finish
- Use Playwright's built-in waiting mechanisms

### 4. Baseline Management

- Review baseline updates carefully
- Update baselines only for intentional changes
- Keep baseline files in version control
- Document significant baseline changes

### 5. CI/CD Integration

- Run tests on multiple browsers
- Store artifacts for review
- Provide clear feedback on failures
- Automate baseline updates when appropriate

### 6. Performance

- Run tests in parallel when possible
- Use appropriate timeouts
- Optimize test execution time
- Monitor test performance

### 7. Maintenance

- Regularly review and update tests
- Remove obsolete tests
- Keep test documentation current
- Monitor test stability

## Configuration

### Playwright Configuration

The testing framework uses multiple Playwright configuration files:

- `playwright.config.js` - Main configuration
- `web/tests/config/interactive-element.config.js` - Interactive tests
- `web/tests/config/state-based.config.js` - State-based tests
- `web/tests/config/cross-browser.config.js` - Cross-browser tests

### Environment Variables

- `BASE_URL` - Base URL for the application (defaults to Railway deployment)
- `CI` - Set to true in CI environments
- `HEADLESS` - Set to false to run in headed mode

### Test Data

Test data is stored in:
- `web/tests/fixtures/test-data.json` - Static test data
- `web/tests/utils/test-helpers.js` - Helper functions
- `web/tests/utils/interactive-helpers.js` - Interactive test helpers

## Support

For issues or questions:

1. Check this documentation
2. Review test logs and reports
3. Check GitHub Actions workflow runs
4. Review Playwright documentation
5. Create an issue in the repository

## Contributing

When adding new visual tests:

1. Follow the existing test structure
2. Add appropriate documentation
3. Update this guide if needed
4. Test your changes thoroughly
5. Update baselines if necessary
