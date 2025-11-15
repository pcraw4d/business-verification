# Testing Guide

This guide provides comprehensive information about testing in the KYB Platform, including how to run tests, test structure, writing new tests, best practices, and debugging.

## Table of Contents

1. [Running Tests](#running-tests)
2. [Test Structure](#test-structure)
3. [Writing New Tests](#writing-new-tests)
4. [Test Best Practices](#test-best-practices)
5. [Debugging Test Failures](#debugging-test-failures)
6. [CI/CD Test Execution](#cicd-test-execution)

## Running Tests

### Frontend Tests

#### Run All Tests
```bash
cd frontend
npm test
```

#### Run Tests in Watch Mode
```bash
npm test -- --watch
```

#### Run Tests with Coverage
```bash
npm test -- --coverage
```

#### Run Specific Test File
```bash
npm test -- --testPathPatterns="api.test"
```

#### Run Tests Matching Pattern
```bash
npm test -- --testNamePattern="should fetch"
```

### Backend Tests

#### Run All Tests
```bash
go test ./...
```

#### Run Tests with Verbose Output
```bash
go test ./... -v
```

#### Run Tests with Coverage
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

#### Run Specific Package Tests
```bash
go test ./internal/api/handlers -v
```

#### Run Integration Tests
```bash
go test ./test/integration -v
```

#### Skip Integration Tests
```bash
go test ./... -short
```

### E2E Tests

#### Run All E2E Tests
```bash
cd frontend
npx playwright test
```

#### Run E2E Tests in UI Mode
```bash
npx playwright test --ui
```

#### Run Specific E2E Test
```bash
npx playwright test merchant-details
```

#### Run E2E Tests in Debug Mode
```bash
npx playwright test --debug
```

### Performance Tests

#### Run Frontend Performance Tests
```bash
cd frontend
npm test -- --testPathPatterns="performance"
```

#### Run Backend Performance Tests
```bash
go test ./test/performance -v
```

## Test Structure

### Frontend Test Structure

```
frontend/
├── __tests__/
│   ├── lib/
│   │   ├── api.test.ts
│   │   ├── api-cache.test.ts
│   │   ├── error-handler.test.ts
│   │   ├── lazy-loader.test.ts
│   │   └── request-deduplicator.test.ts
│   ├── components/
│   │   ├── merchant/
│   │   │   ├── BusinessAnalyticsTab.test.tsx
│   │   │   ├── MerchantDetailsLayout.test.tsx
│   │   │   ├── MerchantOverviewTab.test.tsx
│   │   │   ├── RiskAssessmentTab.test.tsx
│   │   │   └── RiskIndicatorsTab.test.tsx
│   │   └── ui/
│   │       ├── empty-state.test.tsx
│   │       └── progress-indicator.test.tsx
│   └── performance/
│       ├── cache.test.ts
│       ├── deduplication.test.ts
│       └── lazy-loading.test.ts
├── tests/
│   └── e2e/
│       ├── analytics.spec.ts
│       ├── merchant-details.spec.ts
│       └── risk-assessment.spec.ts
├── jest.config.js
├── jest.setup.js
└── playwright.config.ts
```

### Backend Test Structure

```
.
├── internal/
│   ├── api/
│   │   └── handlers/
│   │       └── *_test.go
│   ├── services/
│   │   └── *_test.go
│   └── database/
│       └── *_test.go
├── test/
│   ├── integration/
│   │   ├── database_setup.go
│   │   ├── weeks_2_4_integration_test.go
│   │   └── merchant_portfolio_integration_test.go
│   └── performance/
│       ├── api_load_test.go
│       ├── cache_performance_test.go
│       └── parallel_fetch_test.go
```

## Writing New Tests

### Frontend Unit Test Example

```typescript
import { describe, it, expect, beforeEach } from '@jest/globals';
import { render, screen } from '@testing-library/react';
import { MyComponent } from '@/components/MyComponent';

describe('MyComponent', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should render correctly', () => {
    render(<MyComponent prop="value" />);
    expect(screen.getByText('Expected Text')).toBeInTheDocument();
  });

  it('should handle user interaction', async () => {
    render(<MyComponent prop="value" />);
    const button = screen.getByRole('button');
    button.click();
    await waitFor(() => {
      expect(screen.getByText('Updated Text')).toBeInTheDocument();
    });
  });
});
```

### Backend Unit Test Example

```go
package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMyHandler(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful request",
			request:        httptest.NewRequest("GET", "/api/v1/test", nil),
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"ok"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			handler := NewMyHandler()
			handler.ServeHTTP(w, tt.request)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
```

### E2E Test Example

```typescript
import { test, expect } from '@playwright/test';

test.describe('My Feature', () => {
  test('should work end-to-end', async ({ page }) => {
    await page.goto('/my-page');
    await expect(page.getByText('Welcome')).toBeVisible();
    
    await page.getByRole('button', { name: 'Click Me' }).click();
    await expect(page.getByText('Success')).toBeVisible();
  });
});
```

## Test Best Practices

### Frontend Testing

1. **Use React Testing Library**: Prefer user-centric queries over implementation details
2. **Mock External Dependencies**: Mock API calls, browser APIs, and third-party libraries
3. **Test User Interactions**: Test what users see and do, not implementation details
4. **Use Descriptive Test Names**: Test names should describe what is being tested
5. **Keep Tests Isolated**: Each test should be independent and not rely on other tests
6. **Clean Up**: Use `beforeEach` and `afterEach` to set up and clean up test state

### Backend Testing

1. **Table-Driven Tests**: Use table-driven tests for multiple test cases
2. **Test Edge Cases**: Test error conditions, boundary values, and edge cases
3. **Mock External Services**: Mock external APIs and databases when appropriate
4. **Use Test Helpers**: Create helper functions for common test setup
5. **Test Error Handling**: Ensure error handling is properly tested
6. **Integration Tests**: Use integration tests for testing with real databases

### E2E Testing

1. **Test Critical Paths**: Focus on user journeys that matter most
2. **Use Page Objects**: Create page object models for maintainable tests
3. **Wait for Elements**: Always wait for elements to be visible before interacting
4. **Test Across Browsers**: Test in multiple browsers for compatibility
5. **Keep Tests Fast**: Optimize tests to run quickly
6. **Use Fixtures**: Use test fixtures for consistent test data

## Debugging Test Failures

### Frontend Test Debugging

1. **Check Console Output**: Look for error messages in test output
2. **Use `screen.debug()`**: Print the rendered component to see what's available
3. **Check Mock Setup**: Verify mocks are set up correctly
4. **Verify Async Operations**: Ensure async operations are properly awaited
5. **Check Test Isolation**: Ensure tests aren't affecting each other

### Backend Test Debugging

1. **Use Verbose Output**: Run tests with `-v` flag for detailed output
2. **Check Error Messages**: Read error messages carefully for clues
3. **Use Debugger**: Use `delve` or IDE debugger for complex issues
4. **Check Test Data**: Verify test data is set up correctly
5. **Review Test Logs**: Check test logs for additional context

### E2E Test Debugging

1. **Use Playwright Inspector**: Run tests with `--debug` flag
2. **Take Screenshots**: Use `page.screenshot()` to see what's happening
3. **Check Network Requests**: Use `page.route()` to inspect API calls
4. **Slow Down Tests**: Use `page.setDefaultTimeout()` to see what's happening
5. **Check Browser Console**: Look for JavaScript errors in browser console

## CI/CD Test Execution

### GitHub Actions

Tests run automatically on:
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop` branches

### Test Workflow

1. **Frontend Tests**: Run Jest tests with coverage
2. **Backend Tests**: Run Go tests with race detection
3. **E2E Tests**: Run Playwright tests in CI environment
4. **Coverage Upload**: Upload coverage reports to Codecov

### Local CI Simulation

```bash
# Run all tests as CI would
cd frontend && npm ci && npm test -- --ci
cd .. && go test ./... -race -coverprofile=coverage.out
cd frontend && npx playwright test
```

## Additional Resources

- [Jest Documentation](https://jestjs.io/docs/getting-started)
- [React Testing Library](https://testing-library.com/react)
- [Playwright Documentation](https://playwright.dev)
- [Go Testing Package](https://pkg.go.dev/testing)
- [Test Database Setup Guide](./test-database-setup.md)
