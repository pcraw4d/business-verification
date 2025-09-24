# Feature Functionality Testing

This directory contains comprehensive feature functionality tests for the KYB Platform. These tests validate all critical features including business classification, risk assessment, compliance checking, and merchant management.

## Overview

The feature functionality testing suite provides comprehensive validation of all platform features to ensure they meet quality standards and performance requirements. The tests are designed to be:

- **Comprehensive**: Cover all major features and functionality
- **Reliable**: Provide consistent and repeatable results
- **Fast**: Execute efficiently with parallel processing
- **Maintainable**: Easy to update and extend
- **Informative**: Provide detailed reports and insights

## Test Structure

```
test/
├── feature_functionality_test.go      # Main test suite
├── business_classification_test.go    # Business classification tests
├── risk_assessment_test.go           # Risk assessment tests
├── compliance_checking_test.go       # Compliance checking tests
├── merchant_management_test.go       # Merchant management tests
├── test_runner.go                    # Test execution framework
├── test_config.yaml                  # Test configuration
├── run_feature_tests.sh              # Test execution script
└── README.md                         # This documentation
```

## Test Categories

### 1. Business Classification Tests

Tests the multi-method business classification system:

- **MultiMethodClassification**: Tests the ensemble approach combining multiple classification methods
- **KeywordBasedClassification**: Tests keyword-based industry classification
- **MLBasedClassification**: Tests machine learning-based classification
- **EnsembleClassification**: Tests ensemble classification combining all methods
- **ConfidenceScoring**: Tests confidence scoring and validation

**Expected Results:**
- Accuracy: ≥95%
- Response Time: ≤1 second
- Confidence Score: 0.8-1.0 for high-quality classifications

### 2. Risk Assessment Tests

Tests the comprehensive risk assessment system:

- **ComprehensiveRiskAssessment**: Tests end-to-end risk assessment workflow
- **SecurityAnalysis**: Tests website security analysis
- **DomainAnalysis**: Tests domain reputation and age analysis
- **ReputationAnalysis**: Tests business reputation analysis
- **ComplianceAnalysis**: Tests regulatory compliance analysis
- **FinancialAnalysis**: Tests financial health analysis
- **RiskScoring**: Tests risk scoring algorithms

**Expected Results:**
- Accuracy: ≥90%
- Response Time: ≤3 seconds
- Risk Score Range: 0.0-1.0

### 3. Compliance Checking Tests

Tests various compliance frameworks:

- **AMLCompliance**: Tests Anti-Money Laundering compliance
- **KYCCompliance**: Tests Know Your Customer compliance
- **KYBCompliance**: Tests Know Your Business compliance
- **GDPRCompliance**: Tests General Data Protection Regulation compliance
- **PCICompliance**: Tests Payment Card Industry compliance
- **SOC2Compliance**: Tests Service Organization Control 2 compliance

**Expected Results:**
- Accuracy: ≥95%
- Response Time: ≤2 seconds
- Compliance Status: Valid status for each framework

### 4. Merchant Management Tests

Tests merchant portfolio management functionality:

- **CreateMerchant**: Tests merchant creation
- **GetMerchant**: Tests merchant retrieval
- **UpdateMerchant**: Tests merchant updates
- **DeleteMerchant**: Tests merchant deletion
- **SearchMerchants**: Tests merchant search and filtering
- **BulkOperations**: Tests bulk update operations
- **PortfolioManagement**: Tests portfolio management features

**Expected Results:**
- Accuracy: 100%
- Response Time: ≤500ms
- Data Integrity: Complete and consistent

## Running Tests

### Quick Start

```bash
# Run all tests
./test/run_feature_tests.sh

# Run with verbose output
./test/run_feature_tests.sh --verbose

# Run specific test category
./test/run_feature_tests.sh --business-classification
./test/run_feature_tests.sh --risk-assessment
./test/run_feature_tests.sh --compliance-checking
./test/run_feature_tests.sh --merchant-management
```

### Advanced Usage

```bash
# Run with custom configuration
./test/run_feature_tests.sh --config custom_config.yaml

# Run with custom timeout
./test/run_feature_tests.sh --timeout 10m

# Run benchmark tests
./test/run_feature_tests.sh --benchmark

# Run load tests
./test/run_feature_tests.sh --load-test --timeout 5m

# Clean and run
./test/run_feature_tests.sh --clean --verbose

# Run without cleanup
./test/run_feature_tests.sh --no-cleanup
```

### Using Go Test Directly

```bash
# Run all feature tests
go test -v ./test/...

# Run specific test
go test -v -run TestFeatureFunctionality ./test/...

# Run with timeout
go test -v -timeout 30m ./test/...

# Run benchmark tests
go test -bench=. -benchmem ./test/...

# Run with coverage
go test -v -cover ./test/...
```

## Configuration

### Test Configuration File

The test configuration is defined in `test_config.yaml`:

```yaml
# Test execution settings
test_execution:
  timeout: "30m"
  parallel_tests: true
  verbose_output: true
  generate_report: true
  report_format: "json"

# Service configuration
services:
  classification:
    enabled: true
    timeout: "30s"
    confidence_threshold: 0.8
    
  risk_assessment:
    enabled: true
    timeout: "60s"
    risk_threshold: 0.7

# Test data configuration
test_data:
  path: "./testdata"
  mock_data_enabled: true
  real_data_enabled: false
```

### Environment Variables

- `TEST_CONFIG_FILE`: Path to test configuration file
- `TEST_REPORT_DIR`: Directory for test reports
- `TEST_LOG_DIR`: Directory for test logs
- `TEST_VERBOSE`: Enable verbose output
- `TEST_PARALLEL`: Enable parallel execution
- `TEST_TIMEOUT`: Test timeout duration

## Test Data

### Mock Data

The tests use comprehensive mock data to ensure consistent and reliable results:

- **Business Classification Data**: Sample businesses across various industries
- **Risk Assessment Data**: Test cases for different risk levels
- **Compliance Data**: Sample compliance scenarios
- **Merchant Data**: Test merchant profiles and portfolios

### Real Data (Optional)

For integration testing, real data can be enabled:

```yaml
test_data:
  real_data_enabled: true
  data_sources:
    - name: "production_data"
      path: "./testdata/production.json"
      type: "real"
```

## Test Reports

### Report Formats

Tests generate comprehensive reports in multiple formats:

- **JSON**: Machine-readable format for CI/CD integration
- **HTML**: Human-readable format with visualizations
- **XML**: Standard format for test result aggregation

### Report Contents

- **Test Summary**: Overall test results and statistics
- **Detailed Results**: Individual test results and assertions
- **Performance Metrics**: Response times and throughput
- **Error Analysis**: Failed tests and error details
- **Recommendations**: Suggestions for improvements

### Report Location

Reports are saved to the `test-reports` directory:

```
test-reports/
├── feature_functionality_report.json
├── feature_functionality_report.html
├── performance_metrics.json
└── error_analysis.json
```

## Performance Testing

### Benchmark Tests

Benchmark tests measure performance characteristics:

```bash
# Run benchmark tests
./test/run_feature_tests.sh --benchmark

# Benchmark specific functionality
go test -bench=BenchmarkBusinessClassification ./test/...
go test -bench=BenchmarkRiskAssessment ./test/...
```

### Load Tests

Load tests validate system performance under stress:

```bash
# Run load tests
./test/run_feature_tests.sh --load-test --timeout 10m

# Load test with custom concurrency
go test -run TestLoad -test.timeout=10m ./test/...
```

### Performance Targets

- **Business Classification**: <1 second response time
- **Risk Assessment**: <3 seconds response time
- **Compliance Checking**: <2 seconds response time
- **Merchant Management**: <500ms response time

## Continuous Integration

### GitHub Actions Integration

```yaml
name: Feature Functionality Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.22
      - name: Run Feature Tests
        run: ./test/run_feature_tests.sh --verbose
      - name: Upload Test Reports
        uses: actions/upload-artifact@v2
        with:
          name: test-reports
          path: test/test-reports/
```

### Jenkins Integration

```groovy
pipeline {
    agent any
    stages {
        stage('Feature Tests') {
            steps {
                sh './test/run_feature_tests.sh --verbose'
            }
            post {
                always {
                    publishTestResults testResultsPattern: 'test/test-reports/*.xml'
                    publishHTML([
                        allowMissing: false,
                        alwaysLinkToLastBuild: true,
                        keepAll: true,
                        reportDir: 'test/test-reports',
                        reportFiles: 'feature_functionality_report.html',
                        reportName: 'Feature Test Report'
                    ])
                }
            }
        }
    }
}
```

## Troubleshooting

### Common Issues

1. **Test Timeout**: Increase timeout in configuration
2. **Memory Issues**: Reduce parallel test execution
3. **Service Unavailable**: Check service configuration
4. **Data Issues**: Verify test data setup

### Debug Mode

```bash
# Run with debug logging
TEST_VERBOSE=true ./test/run_feature_tests.sh --verbose

# Run single test with debug
go test -v -run TestBusinessClassification ./test/...
```

### Log Analysis

Test logs are saved to the `test-logs` directory:

```bash
# View test logs
tail -f test/test-logs/test.log

# Search for errors
grep -i error test/test-logs/test.log
```

## Contributing

### Adding New Tests

1. Create test file in appropriate category
2. Follow naming conventions: `*_test.go`
3. Use table-driven tests for multiple scenarios
4. Include comprehensive assertions
5. Update documentation

### Test Guidelines

- **Naming**: Use descriptive test names
- **Structure**: Follow Arrange-Act-Assert pattern
- **Assertions**: Use specific assertions with clear messages
- **Data**: Use consistent test data
- **Cleanup**: Clean up resources after tests

### Code Review

- Verify test coverage
- Check assertion quality
- Validate test data
- Review performance impact
- Ensure maintainability

## Support

For questions or issues with feature functionality testing:

1. Check this documentation
2. Review test logs and reports
3. Consult the main project documentation
4. Create an issue in the project repository

## License

This testing suite is part of the KYB Platform project and follows the same license terms.
