# Test Automation Framework for Risk Assessment Service

## Overview

This comprehensive test automation framework provides a complete testing solution for the Risk Assessment Service, including unit tests, integration tests, performance tests, security tests, end-to-end tests, and ML model validation tests.

## Features

### 🧪 **Comprehensive Test Coverage**
- **Unit Tests**: >95% code coverage with race detection
- **Integration Tests**: API endpoints, external integrations, database operations
- **Performance Tests**: Load testing with Locust, stress testing, spike testing
- **Security Tests**: Vulnerability scanning, input validation, authentication
- **End-to-End Tests**: Complete risk assessment workflows
- **ML Model Tests**: Cross-validation, model performance, robustness testing

### 🔧 **Test Automation Framework**
- **TestExecutor**: Handles test execution across all test types
- **TestReporter**: Generates comprehensive reports in multiple formats
- **TestMonitor**: Real-time monitoring and metrics collection
- **TestCleanup**: Automated cleanup operations
- **TestConfig**: Flexible configuration management

### 📊 **Reporting & Analytics**
- **Multiple Formats**: JSON, HTML, XML, Markdown
- **Coverage Reports**: Detailed code coverage analysis
- **Performance Metrics**: Response time, throughput, error rates
- **Security Reports**: Vulnerability summaries and recommendations
- **Trend Analysis**: Historical test performance tracking

### 🚀 **CI/CD Integration**
- **GitHub Actions**: Automated testing on push/PR
- **Jenkins**: Pipeline integration
- **GitLab CI**: GitLab pipeline support
- **Azure DevOps**: Azure pipeline integration
- **Quality Gates**: Automated deployment blocking

## Test Types

### 1. Unit Tests
- **Coverage**: >95% code coverage target
- **Race Detection**: Enabled for concurrent code testing
- **Build Tags**: Standard Go unit tests
- **Timeout**: 5 minutes
- **Output**: Coverage reports in HTML and text formats

### 2. Integration Tests
- **Scope**: API endpoints, external integrations, database operations
- **Build Tags**: `integration`
- **Dependencies**: Redis, PostgreSQL, Supabase, external APIs
- **Timeout**: 10 minutes
- **Mocking**: External API mocking for reliable testing

### 3. Performance Tests
- **Tool**: Locust for load testing
- **Scenarios**: Standard load, high volume, batch processing, stress testing
- **Thresholds**: 
  - Response time P95 < 1 second
  - Response time P99 < 2 seconds
  - Error rate < 1%
  - Throughput > 1000 req/min
- **Timeout**: 30 minutes

### 4. Security Tests
- **Input Validation**: SQL injection, XSS, command injection, path traversal
- **Authentication**: Token validation, privilege escalation
- **Rate Limiting**: Brute force, DDoS protection
- **CORS**: Origin validation, method restrictions
- **Vulnerability Scanning**: gosec, trivy, nancy, golangci-lint
- **Timeout**: 15 minutes

### 5. End-to-End Tests
- **Workflows**: Complete risk assessment workflows
- **Scenarios**: Single assessment, batch processing, scenario analysis
- **Error Handling**: Invalid requests, timeout handling
- **Performance**: Response time validation
- **Build Tags**: `e2e`
- **Timeout**: 20 minutes

### 6. ML Model Tests
- **Validation**: Cross-validation with 5-10 folds
- **Performance**: Accuracy, precision, recall, F1 score
- **Robustness**: Different data distributions, edge cases
- **Inference**: Single and batch inference performance
- **Build Tags**: `ml`
- **Timeout**: 30 minutes

## Configuration

### Test Configuration (`automation_config.yaml`)
```yaml
# Test Environment
environment:
  name: "test"
  host: "http://localhost:8080"
  port: 8080
  timeout: 30

# Test Types
test_types:
  unit_tests:
    enabled: true
    coverage_threshold: 95.0
    race_detection: true
    
  integration_tests:
    enabled: true
    build_tags: ["integration"]
    
  performance_tests:
    enabled: true
    locust_config:
      users: 100
      spawn_rate: 10
      run_time: "5m"
      
  security_tests:
    enabled: true
    vulnerability_scanning:
      tools: ["gosec", "trivy", "nancy", "golangci-lint"]
      
  e2e_tests:
    enabled: true
    build_tags: ["e2e"]
    
  ml_tests:
    enabled: true
    build_tags: ["ml"]
    cross_validation_folds: 5
```

### Performance Configuration (`performance_config.yaml`)
```yaml
# Load Testing Scenarios
load_tests:
  - name: "standard_load"
    users: 100
    spawn_rate: 10
    run_time: "5m"
    
  - name: "high_volume"
    users: 200
    spawn_rate: 20
    run_time: "10m"

# Performance Thresholds
thresholds:
  response_time:
    p95: 1000
    p99: 2000
  error_rate:
    max: 0.01
  throughput:
    min: 1000
```

### Security Configuration (`security_config.yaml`)
```yaml
# Security Test Categories
security_tests:
  input_validation:
    enabled: true
    tests:
      - name: "sql_injection"
        severity: "critical"
      - name: "xss_attack"
        severity: "high"
        
  authentication:
    enabled: true
    tests:
      - name: "missing_auth"
        severity: "high"
        
  vulnerability_scanning:
    enabled: true
    tools:
      - name: "gosec"
        severity_threshold: "medium"
      - name: "trivy"
        severity_threshold: "medium"
```

## Usage

### Command Line Interface

#### Run All Tests
```bash
./test/automation/run_automation.sh
```

#### Run Specific Test Types
```bash
# Unit tests only
./test/automation/run_automation.sh -t unit

# Integration tests only
./test/automation/run_automation.sh -t integration

# Performance tests only
./test/automation/run_automation.sh -t performance

# Security tests only
./test/automation/run_automation.sh -t security

# End-to-end tests only
./test/automation/run_automation.sh -t e2e

# ML model tests only
./test/automation/run_automation.sh -t ml
```

#### Advanced Options
```bash
# Run with verbose output
./test/automation/run_automation.sh -v

# Run tests in parallel
./test/automation/run_automation.sh -p

# Use custom configuration
./test/automation/run_automation.sh -c custom_config.yaml

# Custom output directory
./test/automation/run_automation.sh -o ./custom_reports

# Skip cleanup operations
./test/automation/run_automation.sh --skip-cleanup
```

### Performance Testing

#### Run Performance Tests
```bash
./test/performance/run_performance_tests.sh
```

#### Performance Test Options
```bash
# Load testing with 200 users for 10 minutes
./test/performance/run_performance_tests.sh -t load -u 200 -d 10m

# Stress testing in headless mode
./test/performance/run_performance_tests.sh -t stress -u 500 -H

# Spike testing with custom host
./test/performance/run_performance_tests.sh -t spike -h https://staging.example.com -u 1000
```

### Security Testing

#### Run Security Tests
```bash
./test/security/run_security_tests.sh
```

#### Security Test Options
```bash
# Run only security tests (skip scanning)
./test/security/run_security_tests.sh --skip-scanning

# Run only vulnerability scanning
./test/security/run_security_tests.sh --skip-tests

# Run with custom host
./test/security/run_security_tests.sh -h https://staging.example.com
```

## Test Execution

### Local Development
```bash
# Run all tests locally
make test

# Run specific test types
make test-unit
make test-integration
make test-performance
make test-security
make test-e2e
make test-ml

# Run with coverage
make test-coverage

# Run with race detection
make test-race
```

### CI/CD Pipeline
```yaml
# GitHub Actions example
name: Test Automation
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.22'
      
      - name: Run Test Automation
        run: ./test/automation/run_automation.sh
      
      - name: Upload Test Reports
        uses: actions/upload-artifact@v3
        with:
          name: test-reports
          path: test/automation/reports/
```

## Test Data Management

### Test Data Types
- **Risk Assessment Requests**: Various business scenarios
- **Business Data**: Company information, addresses, industries
- **External API Responses**: Mocked external service responses
- **ML Training Data**: Generated training and validation data
- **Performance Test Data**: Load testing scenarios

### Test Data Generation
```go
// Generate test data for various scenarios
func GenerateTestData() map[string]interface{} {
    return map[string]interface{}{
        "valid_business": CreateTestRiskAssessmentRequest(),
        "invalid_business": &models.RiskAssessmentRequest{
            BusinessName: "", // Invalid: empty name
        },
        "high_risk_business": &models.RiskAssessmentRequest{
            BusinessName: "High Risk Company",
            Industry: "financial_services",
        },
    }
}
```

## Monitoring & Metrics

### Test Metrics
- **Test Execution Time**: Duration of test runs
- **Test Success Rate**: Percentage of passing tests
- **Test Failure Rate**: Percentage of failing tests
- **Coverage Percentage**: Code coverage metrics
- **Performance Metrics**: Response time, throughput, error rates
- **Security Vulnerabilities**: Number and severity of vulnerabilities

### Real-time Monitoring
```yaml
monitoring:
  enabled: true
  metrics_endpoint: "http://localhost:9090/metrics"
  real_time:
    enabled: true
    update_interval: 10
    alerts:
      - condition: "test_failure_rate > 0.1"
        message: "Test failure rate exceeds 10%"
        severity: "warning"
```

## Quality Gates

### Quality Criteria
- **Test Coverage**: ≥95%
- **Test Success Rate**: ≥95%
- **Security Vulnerabilities**: 0 critical, 0 high
- **Performance Response Time P95**: ≤1000ms
- **Performance Error Rate**: ≤1%

### Gate Behavior
```yaml
quality_gates:
  enabled: true
  criteria:
    - name: "test_coverage"
      threshold: 95.0
      operator: ">="
    - name: "test_success_rate"
      threshold: 95.0
      operator: ">="
  behavior:
    fail_on_gate_failure: true
    block_deployment: true
    send_notifications: true
```

## Reporting

### Report Formats
- **JSON**: Machine-readable test results
- **HTML**: Interactive test reports with charts
- **XML**: JUnit-compatible test results
- **Markdown**: Human-readable test summaries

### Report Content
- **Test Results**: Pass/fail status, duration, errors
- **Coverage Report**: Code coverage analysis
- **Performance Metrics**: Response time, throughput, error rates
- **Security Scan Results**: Vulnerability summaries
- **Error Details**: Detailed error information
- **Recommendations**: Actionable improvement suggestions

### Sample Report Structure
```
test_reports/
├── test_automation_report_20241219_143022.md
├── coverage.html
├── coverage.txt
├── performance_report.html
├── gosec_report.json
├── trivy_report.json
├── nancy_report.json
└── golangci_report.json
```

## Best Practices

### Test Organization
- **Unit Tests**: Place alongside source code (`*_test.go`)
- **Integration Tests**: Use `integration` build tag
- **Performance Tests**: Use `performance` build tag
- **Security Tests**: Use `security` build tag
- **E2E Tests**: Use `e2e` build tag
- **ML Tests**: Use `ml` build tag

### Test Naming
- **Descriptive Names**: Clear test purpose and scenario
- **Table-driven Tests**: Use for multiple test cases
- **Consistent Structure**: Arrange, Act, Assert pattern
- **Error Messages**: Include context in assertions

### Test Data
- **Isolation**: Each test should be independent
- **Cleanup**: Remove test data after tests
- **Realistic Data**: Use realistic test scenarios
- **Edge Cases**: Include boundary conditions

### Performance Testing
- **Baseline Metrics**: Establish performance baselines
- **Gradual Load**: Start with low load and increase
- **Realistic Scenarios**: Use production-like data
- **Monitor Resources**: Track CPU, memory, disk usage

### Security Testing
- **Comprehensive Coverage**: Test all security aspects
- **Regular Scanning**: Run vulnerability scans frequently
- **Dependency Updates**: Keep dependencies updated
- **Security Headers**: Validate security headers

## Troubleshooting

### Common Issues

#### Test Failures
```bash
# Check test logs
tail -f test.log

# Run tests with verbose output
go test -v ./...

# Run specific test
go test -v -run TestSpecificFunction ./...
```

#### Performance Issues
```bash
# Check system resources
top
htop
iostat

# Profile Go application
go tool pprof http://localhost:8080/debug/pprof/profile
```

#### Security Issues
```bash
# Check security scan results
cat test_reports/gosec_report.json | jq '.Issues'

# Update dependencies
go mod tidy
go mod vendor
```

### Debug Mode
```bash
# Enable debug logging
export LOG_LEVEL=debug

# Run tests with debug output
./test/automation/run_automation.sh -v
```

## Contributing

### Adding New Tests
1. **Create Test File**: Follow naming convention (`*_test.go`)
2. **Add Build Tags**: Use appropriate build tags
3. **Update Configuration**: Add test type to config
4. **Update Documentation**: Document new test type
5. **Run Tests**: Verify tests pass

### Test Framework Extensions
1. **Add New Test Type**: Extend `TestExecutor`
2. **Add New Reporter**: Extend `TestReporter`
3. **Add New Monitor**: Extend `TestMonitor`
4. **Update Configuration**: Add new config options
5. **Update Documentation**: Document new features

## Support

### Documentation
- **API Documentation**: `/docs/api/`
- **Architecture Documentation**: `/docs/architecture/`
- **Performance Documentation**: `/docs/performance/`
- **Security Documentation**: `/docs/security/`

### Contact
- **Issues**: GitHub Issues
- **Discussions**: GitHub Discussions
- **Email**: team@kyb-platform.com

## License

This test automation framework is part of the KYB Platform and is licensed under the MIT License.
