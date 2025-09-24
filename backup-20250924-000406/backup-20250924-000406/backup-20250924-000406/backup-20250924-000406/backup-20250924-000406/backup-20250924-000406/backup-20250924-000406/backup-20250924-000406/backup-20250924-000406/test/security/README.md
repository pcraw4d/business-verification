# Security Testing Framework

## Overview

This security testing framework provides comprehensive security testing for the KYB Platform as part of **Subtask 4.2.4: Security Testing** from the Supabase Table Improvement Implementation Plan. The framework tests authentication flows, authorization controls, data access restrictions, and audit logging to ensure the platform meets security requirements.

## Features

### 🔐 Authentication Testing
- JWT token validation and expiration
- API key authentication
- Invalid token rejection
- Missing authentication handling
- Session management

### 🛡️ Authorization Testing
- Role-based access control (RBAC)
- Permission validation
- Admin vs user access controls
- API key role validation

### 🔒 Data Access Testing
- User data isolation
- Sensitive data protection
- Data filtering and privacy controls
- Cross-user data access prevention

### 📋 Audit Logging Testing
- Authentication event logging
- Authorization event logging
- Security event tracking
- Compliance logging

### 🔍 Input Validation Testing
- SQL injection prevention
- XSS attack prevention
- Path traversal protection
- Command injection prevention

### ⚡ Rate Limiting Testing
- Request throttling
- Abuse prevention
- Rate limit enforcement
- Burst handling

### 🔒 Security Headers Testing
- Security header validation
- Missing header detection
- Header configuration verification

## Quick Start

### Running Security Tests

```bash
# Run all security tests
go test ./test/security/... -v

# Run specific test categories
go test ./test/security/... -run TestAuthenticationFlows -v
go test ./test/security/... -run TestAuthorizationControls -v
go test ./test/security/... -run TestDataAccessRestrictions -v
go test ./test/security/... -run TestAuditLogging -v

# Run comprehensive security testing
go test ./test/security/... -run TestComprehensiveSecurityTesting -v
```

### Test Reports

After running tests, reports are generated in `test/reports/security/`:

- `security_test_results.json` - Detailed JSON report
- `security_test_report.md` - Comprehensive markdown report
- `security_summary.md` - Quick summary report

## Test Categories

### 1. Authentication Flows (`AUTHENTICATION`)

Tests the authentication system to ensure:
- Valid tokens are accepted
- Invalid tokens are rejected
- Expired tokens are handled properly
- Missing authentication is detected
- API keys work correctly

**Critical Tests:**
- Valid JWT Token Authentication
- Valid API Key Authentication
- Invalid Token Rejection
- Expired Token Rejection
- Missing Authentication Header

### 2. Authorization Controls (`AUTHORIZATION`)

Tests role-based access control to ensure:
- Admin users can access admin endpoints
- Regular users are denied admin access
- API keys have appropriate permissions
- Role validation works correctly

**Critical Tests:**
- Admin Role Access
- User Role Access Denied
- API Key Role Validation

### 3. Data Access Restrictions (`DATA_ACCESS`)

Tests data isolation and privacy to ensure:
- Users can only access their own data
- Sensitive data is not exposed
- Cross-user data access is prevented
- Data filtering works correctly

**Critical Tests:**
- User Data Isolation
- Sensitive Data Protection

### 4. Audit Logging (`AUDIT_LOGGING`)

Tests audit logging functionality to ensure:
- Authentication events are logged
- Authorization events are tracked
- Security events are recorded
- Audit logs are accessible

**Critical Tests:**
- Authentication Event Logging
- Authorization Event Logging

### 5. Input Validation (`INPUT_VALIDATION`)

Tests input sanitization to prevent:
- SQL injection attacks
- XSS attacks
- Path traversal attacks
- Command injection attacks

**Critical Tests:**
- SQL Injection Prevention
- XSS Prevention

### 6. Rate Limiting (`RATE_LIMITING`)

Tests rate limiting functionality to ensure:
- Request throttling works
- Abuse is prevented
- Rate limits are enforced
- Burst handling is correct

**Tests:**
- Rate Limiting Enforcement

### 7. Security Headers (`SECURITY_HEADERS`)

Tests security headers to ensure:
- Required headers are present
- Header values are correct
- Security policies are enforced

**Tests:**
- Security Headers Implementation

## Configuration

### Security Test Configuration

The framework uses `SecurityTestSuiteConfig` for configuration:

```go
config := DefaultSecurityTestSuiteConfig()

// Customize configuration
config.Config.Timeout = 60 * time.Second
config.Config.RateLimitEnabled = true
config.Config.RequestsPerMinute = 100

// Enable/disable test categories
config.Categories.Authentication = true
config.Categories.Authorization = true
config.Categories.DataAccess = true
config.Categories.AuditLogging = true
config.Categories.InputValidation = true
config.Categories.RateLimiting = true
config.Categories.SecurityHeaders = true
```

### Test Environment

Configure the test environment:

```go
config.Environment.BaseURL = "http://localhost:8080"
config.Environment.DatabaseURL = "postgres://test:test@localhost:5432/kyb_test"
config.Environment.Environment = "test"
```

### Security Thresholds

Set security test thresholds:

```go
config.Thresholds.MinPassRate = 0.80          // 80% pass rate required
config.Thresholds.MaxCriticalFailures = 0     // No critical failures allowed
config.Thresholds.MinSecurityScore = 70       // Minimum security score of 70/100
```

## Test Results

### Result Status

Each test can have one of three statuses:

- **PASS** ✅ - Test passed successfully
- **FAIL** ❌ - Test failed (security issue found)
- **WARN** ⚠️ - Test passed with warnings (potential improvement)

### Security Score

The framework calculates an overall security score (0-100) based on:
- Passed tests: +10 points
- Warning tests: +5 points
- Failed tests: -5 points
- Critical failures: -20 points

### Critical Failures

Critical failures are security issues that must be addressed immediately:
- Authentication failures
- Authorization bypasses
- Input validation failures
- Data access violations

## Integration with CI/CD

### GitHub Actions

Add security testing to your CI/CD pipeline:

```yaml
name: Security Tests
on: [push, pull_request]

jobs:
  security-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.22'
      
      - name: Run Security Tests
        run: go test ./test/security/... -v
      
      - name: Upload Security Reports
        uses: actions/upload-artifact@v3
        with:
          name: security-reports
          path: test/reports/security/
```

### Pre-commit Hooks

Add security testing to pre-commit hooks:

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running security tests..."
go test ./test/security/... -v

if [ $? -ne 0 ]; then
    echo "Security tests failed. Commit rejected."
    exit 1
fi

echo "Security tests passed. Proceeding with commit."
```

## Troubleshooting

### Common Issues

#### Test Failures

1. **Authentication Failures**
   - Check JWT secret configuration
   - Verify token expiration settings
   - Ensure API keys are properly configured

2. **Authorization Failures**
   - Verify role-based access control implementation
   - Check permission configurations
   - Ensure admin endpoints are properly protected

3. **Input Validation Failures**
   - Review input sanitization logic
   - Check SQL injection prevention
   - Verify XSS protection implementation

#### Performance Issues

1. **Slow Test Execution**
   - Reduce test timeout settings
   - Disable parallel execution if needed
   - Check database connection performance

2. **Memory Issues**
   - Increase memory limits
   - Check for memory leaks in test code
   - Optimize test data generation

### Debug Mode

Enable debug mode for detailed troubleshooting:

```bash
export SECURITY_TEST_DEBUG=true
go test ./test/security/... -v
```

## Best Practices

### Security Testing

1. **Regular Testing**
   - Run security tests on every commit
   - Schedule regular security audits
   - Monitor security test results

2. **Test Coverage**
   - Ensure all security-critical code is tested
   - Test both positive and negative scenarios
   - Include edge cases and boundary conditions

3. **Test Data**
   - Use realistic test data
   - Avoid production data in tests
   - Clean up test data after tests

### Maintenance

1. **Keep Tests Updated**
   - Update tests when security requirements change
   - Add new tests for new security features
   - Remove obsolete tests

2. **Monitor Results**
   - Track security test trends over time
   - Investigate test failures immediately
   - Document security improvements

## Contributing

### Adding New Tests

1. **Create Test Function**
   ```go
   func TestNewSecurityFeature(t *testing.T) {
       // Test implementation
   }
   ```

2. **Add to Test Suite**
   ```go
   func (sts *SecurityTestSuite) TestNewSecurityFeature(t *testing.T) []SecurityTestResult {
       // Test implementation
   }
   ```

3. **Update Documentation**
   - Add test description to README
   - Update test categories if needed
   - Document test requirements

### Test Standards

1. **Naming Convention**
   - Use descriptive test names
   - Include test category in name
   - Follow Go testing conventions

2. **Test Structure**
   - Use table-driven tests when appropriate
   - Include setup and teardown
   - Clean up resources properly

3. **Assertions**
   - Use clear, specific assertions
   - Include helpful error messages
   - Test both success and failure cases

## Support

For questions or issues with the security testing framework:

1. Check the troubleshooting section
2. Review test logs and reports
3. Consult the main project documentation
4. Create an issue in the project repository

## License

This security testing framework is part of the KYB Platform project and follows the same license terms.
