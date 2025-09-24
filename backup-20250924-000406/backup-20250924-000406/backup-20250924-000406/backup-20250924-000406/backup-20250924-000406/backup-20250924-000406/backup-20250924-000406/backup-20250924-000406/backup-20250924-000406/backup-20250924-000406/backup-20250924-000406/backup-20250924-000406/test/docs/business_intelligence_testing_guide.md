# Business Intelligence API Testing Guide

## Overview

This document provides comprehensive guidance for testing the Business Intelligence API. The testing suite includes unit tests, integration tests, performance tests, and error handling tests to ensure the API's reliability, performance, and robustness.

## Table of Contents

1. [Test Structure](#test-structure)
2. [Running Tests](#running-tests)
3. [Test Categories](#test-categories)
4. [Test Configuration](#test-configuration)
5. [Test Data Management](#test-data-management)
6. [Performance Testing](#performance-testing)
7. [Error Handling Testing](#error-handling-testing)
8. [Integration Testing](#integration-testing)
9. [Test Reporting](#test-reporting)
10. [Troubleshooting](#troubleshooting)

## Test Structure

The test suite is organized into the following directories:

```
test/
├── unit/                           # Unit tests
│   └── business_intelligence_handler_test.go
├── integration/                    # Integration tests
│   └── business_intelligence_integration_test.go
├── performance/                    # Performance tests
│   └── business_intelligence_performance_test.go
├── error_handling/                 # Error handling tests
│   └── business_intelligence_error_handling_test.go
├── config/                         # Test configuration
│   └── business_intelligence_test_config.yaml
├── docs/                          # Test documentation
│   └── business_intelligence_testing_guide.md
└── reports/                       # Test reports (generated)
    ├── coverage.html
    ├── performance.json
    └── integration.json
```

## Running Tests

### Prerequisites

- Go 1.22 or later
- Access to the project repository
- Required dependencies installed

### Quick Start

Run all tests with a single command:

```bash
./scripts/run_business_intelligence_tests.sh
```

### Specific Test Categories

#### Unit Tests Only
```bash
./scripts/run_business_intelligence_tests.sh --unit
```

#### Integration Tests Only
```bash
./scripts/run_business_intelligence_tests.sh --integration
```

#### Performance Tests Only
```bash
./scripts/run_business_intelligence_tests.sh --performance
```

#### Error Handling Tests Only
```bash
./scripts/run_business_intelligence_tests.sh --error
```

#### Tests with Coverage
```bash
./scripts/run_business_intelligence_tests.sh --coverage
```

### Manual Test Execution

#### Unit Tests
```bash
go test -v ./internal/api/handlers/
```

#### Integration Tests
```bash
go test -v -tags=integration ./test/integration/
```

#### Performance Tests
```bash
go test -bench=. -benchmem ./test/performance/
```

#### Error Handling Tests
```bash
go test -v ./test/error_handling/
```

## Test Categories

### 1. Unit Tests

Unit tests focus on testing individual functions and methods in isolation.

**Location**: `./internal/api/handlers/business_intelligence_handler_test.go`

**Coverage**:
- Handler initialization
- Request validation
- Response generation
- Error handling
- Data processing functions

**Key Test Cases**:
- Valid request processing
- Invalid request handling
- Response format validation
- Error response generation
- Data transformation

### 2. Integration Tests

Integration tests verify that different components work together correctly.

**Location**: `./test/integration/business_intelligence_integration_test.go`

**Coverage**:
- Complete workflow testing
- Cross-service integration
- Data persistence
- Job processing
- API endpoint interactions

**Key Test Scenarios**:
- Market Analysis Workflow
- Competitive Analysis Workflow
- Growth Analytics Workflow
- Business Intelligence Aggregation Workflow
- Cross-Service Integration

### 3. Performance Tests

Performance tests measure the API's speed, throughput, and resource usage.

**Location**: `./test/performance/business_intelligence_performance_test.go`

**Coverage**:
- Response time benchmarks
- Throughput measurements
- Memory usage analysis
- Concurrent access testing
- Load testing

**Key Benchmarks**:
- Market Analysis Creation
- Competitive Analysis Creation
- Growth Analytics Creation
- Business Intelligence Aggregation Creation
- Analysis Retrieval
- Analysis Listing
- Job Creation and Status Checking

### 4. Error Handling Tests

Error handling tests ensure the API gracefully handles various error conditions.

**Location**: `./test/error_handling/business_intelligence_error_handling_test.go`

**Coverage**:
- Invalid JSON handling
- Missing required fields
- Invalid data types
- Invalid time ranges
- Invalid analysis types
- Non-existent resource access
- Concurrent access errors
- Rate limiting
- Memory exhaustion
- Network timeout simulation

## Test Configuration

### Configuration File

The test configuration is managed through `test/config/business_intelligence_test_config.yaml`.

### Key Configuration Sections

#### Test Environment Settings
```yaml
test_environment:
  name: "business_intelligence_test"
  description: "Test environment for Business Intelligence API"
  timeout: 30s
  retry_attempts: 3
  retry_delay: 1s
```

#### Performance Test Configuration
```yaml
performance_tests:
  benchmarks:
    market_analysis_creation:
      iterations: 1000
      timeout: 5s
      max_memory: 100MB
```

#### Integration Test Configuration
```yaml
integration_tests:
  scenarios:
    - name: "market_analysis_workflow"
      description: "Complete market analysis workflow"
      steps:
        - create_market_analysis
        - retrieve_market_analysis
        - list_market_analyses
```

## Test Data Management

### Sample Test Data

The test suite includes predefined sample data for consistent testing:

#### Sample Businesses
```yaml
sample_businesses:
  - id: "test-business-1"
    name: "Test Technology Company"
    industry: "Technology"
    geographic_area: "North America"
    size: "medium"
    revenue: 1000000
```

#### Sample Competitors
```yaml
sample_competitors:
  - "Competitor A"
  - "Competitor B"
  - "Competitor C"
  - "Competitor D"
  - "Competitor E"
```

#### Sample Time Ranges
```yaml
sample_time_ranges:
  - name: "last_month"
    start_date: "2024-01-01T00:00:00Z"
    end_date: "2024-01-31T23:59:59Z"
    time_zone: "UTC"
```

### Test Data Cleanup

The test suite automatically cleans up test data after execution:

```yaml
test_cleanup:
  cleanup_after_tests: true
  cleanup_test_data: true
  cleanup_temp_files: true
  cleanup_logs: false
```

## Performance Testing

### Benchmark Tests

Performance tests measure the API's performance characteristics:

#### Response Time Benchmarks
- Market Analysis Creation: < 100ms
- Competitive Analysis Creation: < 100ms
- Growth Analytics Creation: < 100ms
- Business Intelligence Aggregation Creation: < 150ms
- Analysis Retrieval: < 50ms
- Analysis Listing: < 50ms

#### Throughput Tests
- Concurrent user simulation
- Load testing with 100+ concurrent users
- Stress testing with 1000+ concurrent users

#### Memory Usage Tests
- Memory consumption during high load
- Memory leak detection
- Garbage collection analysis

### Performance Test Execution

```bash
# Run all performance tests
go test -bench=. -benchmem ./test/performance/

# Run specific benchmark
go test -bench=BenchmarkMarketAnalysisCreation ./test/performance/

# Run with memory profiling
go test -bench=. -memprofile=mem.prof ./test/performance/
```

## Error Handling Testing

### Error Scenarios

The error handling tests cover various error conditions:

#### Input Validation Errors
- Invalid JSON format
- Missing required fields
- Invalid data types
- Invalid time ranges
- Invalid analysis types

#### Resource Access Errors
- Non-existent analysis access
- Non-existent job access
- Non-existent aggregation access

#### System Errors
- Concurrent access conflicts
- Rate limiting
- Memory exhaustion
- Network timeouts

### Error Response Validation

All error responses are validated for:
- Correct HTTP status codes
- Proper error message format
- Error code consistency
- Response structure validation

## Integration Testing

### Workflow Testing

Integration tests verify complete workflows:

#### Market Analysis Workflow
1. Create market analysis
2. Retrieve market analysis
3. List market analyses
4. Create background job
5. Check job status
6. List jobs

#### Competitive Analysis Workflow
1. Create competitive analysis
2. Retrieve competitive analysis
3. List competitive analyses
4. Create background job
5. Check job status
6. List jobs

#### Growth Analytics Workflow
1. Create growth analytics
2. Retrieve growth analytics
3. List growth analytics
4. Create background job
5. Check job status
6. List jobs

#### Business Intelligence Aggregation Workflow
1. Create aggregation
2. Retrieve aggregation
3. List aggregations
4. Create background job
5. Check job status
6. List jobs

### Cross-Service Integration

Tests verify that different services work together:
- Market analysis + Competitive analysis + Growth analytics
- Aggregation of multiple analysis types
- Data consistency across services
- Job processing coordination

## Test Reporting

### Report Formats

The test suite generates reports in multiple formats:
- JSON: Machine-readable format
- HTML: Human-readable format with coverage
- XML: JUnit-compatible format
- Coverage: Go coverage format

### Report Locations

Reports are generated in the `test/reports/` directory:
- `coverage.html`: HTML coverage report
- `performance.json`: Performance test results
- `integration.json`: Integration test results
- `error_handling.json`: Error handling test results

### Coverage Reports

Coverage reports show:
- Line coverage percentage
- Function coverage percentage
- Branch coverage percentage
- Uncovered code sections

## Troubleshooting

### Common Issues

#### Test Failures

**Issue**: Tests fail with "connection refused" errors
**Solution**: Ensure the API server is running and accessible

**Issue**: Tests fail with timeout errors
**Solution**: Increase timeout values in test configuration

**Issue**: Tests fail with memory errors
**Solution**: Reduce test data size or increase system memory

#### Performance Issues

**Issue**: Benchmarks show slow performance
**Solution**: 
- Check system resources
- Verify test data size
- Review test configuration

**Issue**: Memory usage is high
**Solution**:
- Check for memory leaks
- Reduce test data size
- Optimize test execution

#### Integration Issues

**Issue**: Integration tests fail intermittently
**Solution**:
- Add retry logic
- Increase timeout values
- Check test data consistency

### Debug Mode

Enable debug mode for detailed logging:

```bash
go test -v -race -coverprofile=coverage.out -covermode=atomic ./... -args -debug
```

### Verbose Output

Enable verbose output for detailed test information:

```bash
go test -v ./test/integration/
```

### Test Isolation

Ensure tests run in isolation:

```bash
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
```

## Best Practices

### Test Development

1. **Write Clear Test Names**: Use descriptive test names that explain what is being tested
2. **Use Table-Driven Tests**: For testing multiple scenarios with similar logic
3. **Mock External Dependencies**: Use mocks for external services and databases
4. **Test Edge Cases**: Include tests for boundary conditions and error cases
5. **Maintain Test Data**: Keep test data consistent and up-to-date

### Test Execution

1. **Run Tests Regularly**: Execute tests as part of the development workflow
2. **Monitor Performance**: Track performance metrics over time
3. **Review Coverage**: Maintain high test coverage (>80%)
4. **Fix Failing Tests**: Address test failures promptly
5. **Update Tests**: Keep tests synchronized with code changes

### Test Maintenance

1. **Regular Updates**: Update tests when requirements change
2. **Cleanup**: Remove obsolete tests and test data
3. **Documentation**: Keep test documentation current
4. **Review**: Regularly review test effectiveness
5. **Optimization**: Optimize slow tests for better performance

## Conclusion

This testing guide provides comprehensive coverage of the Business Intelligence API testing suite. By following the guidelines and best practices outlined in this document, you can ensure the API's reliability, performance, and robustness.

For additional support or questions, please refer to the project documentation or contact the development team.
