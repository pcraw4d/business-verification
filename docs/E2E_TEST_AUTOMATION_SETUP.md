# E2E Test Automation Setup

**Date**: 2025-01-17  
**Status**: ✅ **COMPLETE**

## Overview

This document describes the E2E test automation setup, including reporting, notifications, and scheduled test execution.

## Test Configuration

### Playwright Configuration

The Playwright configuration (`frontend/playwright.config.ts`) is set up with:

- **Multiple Reporters** (CI environments):
  - HTML reporter for detailed visual reports
  - JSON reporter for programmatic analysis
  - JUnit reporter for CI/CD integration
  - List reporter for console output

- **Test Execution**:
  - Parallel execution enabled
  - Retries on CI (2 retries)
  - Screenshots on failure
  - Traces on first retry

- **Browser Coverage**:
  - Chromium (Desktop Chrome)
  - Firefox (Desktop Firefox)
  - WebKit (Desktop Safari)
  - Mobile Chrome (Pixel 5)
  - Mobile Safari (iPhone 12)

## CI/CD Integration

### GitHub Actions Workflow

The CI/CD pipeline (`.github/workflows/frontend-ci.yml`) includes:

1. **E2E Tests (Local)**:
   - Runs against local development server
   - Tests all E2E test suites
   - Uploads HTML report as artifact

2. **E2E Tests (Railway)**:
   - Runs against deployed Railway environment
   - Tests production-like environment
   - Uploads HTML report as artifact
   - Generates test summary in GitHub Actions

3. **Test Summaries**:
   - Automatic test result summaries in GitHub Actions
   - Test statistics (total, passed, failed, skipped)
   - Links to detailed reports

### Scheduled Test Execution

A separate workflow (`.github/workflows/e2e-scheduled.yml`) runs:

- **Schedule**: Daily at 2 AM UTC
- **Manual Trigger**: Available via `workflow_dispatch`
- **Notifications**: Creates GitHub Issue on failure
- **Artifact Retention**: 90 days (longer than regular CI runs)

## Test Reports

### Report Types

1. **HTML Report**:
   - Visual test results
   - Screenshots of failures
   - Traces for debugging
   - Accessible via GitHub Actions artifacts

2. **JSON Report**:
   - Machine-readable format
   - Used for programmatic analysis
   - Includes test statistics

3. **JUnit Report**:
   - Standard CI/CD format
   - Compatible with test result parsers
   - Used for test result aggregation

### Accessing Reports

**Local Development**:
```bash
# Run tests and view report
npm run test:e2e
npm run test:e2e:report
```

**CI/CD**:
- Reports are automatically uploaded as GitHub Actions artifacts
- Access via: GitHub Actions → Workflow Run → Artifacts
- Test summaries appear in workflow summary

## Notifications

### Failure Notifications

When scheduled E2E tests fail:
- **GitHub Issue Created**: Automatic issue with failure details
- **Labels Applied**: `bug`, `e2e-tests`, `automated`
- **Issue Includes**: Workflow run link, date, failure context

### Success Notifications

- Test summaries in GitHub Actions workflow summary
- All test results visible in workflow run
- Artifacts available for download

## Test Execution

### Local Execution

```bash
# Run all E2E tests locally
npm run test:e2e

# Run against Railway
npm run test:e2e:railway

# Run with UI
npm run test:e2e:ui

# Run in headed mode (see browser)
npm run test:e2e:headed

# View last test report
npm run test:e2e:report
```

### CI/CD Execution

Tests run automatically on:
- **Push to main**: All E2E tests (local and Railway)
- **Pull Requests**: All E2E tests (local and Railway)
- **Scheduled**: Daily Railway tests at 2 AM UTC

## Test Coverage

### Critical User Journeys

- ✅ Complete merchant onboarding flow
- ✅ Merchant discovery and analysis flow
- ✅ Compliance monitoring flow
- ✅ Bulk operations workflow

### Error Scenarios

- ✅ Network failure handling
- ✅ API timeout handling
- ✅ Partial API failure handling
- ✅ Invalid API response format handling

### Mobile Responsiveness

- ✅ Mobile navigation
- ✅ Mobile forms
- ✅ Mobile tables

### Existing Test Suites

- ✅ Forms (validation, submission, errors)
- ✅ Navigation (all routes)
- ✅ Data loading (dashboard, compliance, sessions, portfolio)
- ✅ Analytics (data loading, lazy loading, empty states)
- ✅ Risk assessment (start, display, polling)
- ✅ Export (CSV, JSON, risk assessment tab)
- ✅ Bulk operations
- ✅ Merchant details (loading, tabs, information display, errors)

## Best Practices

### Writing E2E Tests

1. **Use Multiple Selectors**: Try role-based, value-based, and text-based selectors
2. **Graceful Skipping**: Skip tests when elements aren't found (use `test.skip()`)
3. **Wait for Elements**: Use `waitForLoadState` and explicit waits
4. **Error Handling**: Use `.catch()` for optional elements
5. **Mobile Support**: Test with mobile viewports when relevant

### Test Maintenance

1. **Regular Execution**: Tests run daily via scheduled workflow
2. **Monitor Failures**: Check GitHub Issues for automated failure reports
3. **Update Selectors**: Keep selectors updated with UI changes
4. **Review Reports**: Regularly review HTML reports for flaky tests

## Troubleshooting

### Tests Failing in CI

1. **Check Artifacts**: Download Playwright report from GitHub Actions
2. **Review Screenshots**: Check failure screenshots in HTML report
3. **Check Traces**: Use Playwright trace viewer for detailed debugging
4. **Review Logs**: Check workflow logs for environment issues

### Tests Timing Out

1. **Increase Timeouts**: Adjust timeouts in test files if needed
2. **Check Environment**: Verify Railway environment is accessible
3. **Review Network**: Check for network issues in test environment

### Flaky Tests

1. **Add Retries**: Tests automatically retry 2 times in CI
2. **Improve Selectors**: Use more specific, stable selectors
3. **Add Waits**: Ensure proper waiting for dynamic content
4. **Review Timing**: Check for race conditions

## Future Enhancements

### Potential Improvements

1. **Slack/Discord Notifications**: Add webhook notifications for test failures
2. **Test Result Dashboard**: Create dashboard for test result trends
3. **Performance Metrics**: Track test execution time trends
4. **Test Coverage Metrics**: Track E2E test coverage over time
5. **Visual Regression**: Add visual regression testing
6. **Accessibility Testing**: Integrate accessibility checks in E2E tests

## Related Documentation

- **CI/CD Guide**: `docs/CI_CD_GUIDE.md`
- **Testing Documentation**: `docs/COMPREHENSIVE_TESTING_DOCUMENTATION.md`
- **Playwright Docs**: https://playwright.dev/docs/intro

---

**Last Updated**: 2025-01-17

