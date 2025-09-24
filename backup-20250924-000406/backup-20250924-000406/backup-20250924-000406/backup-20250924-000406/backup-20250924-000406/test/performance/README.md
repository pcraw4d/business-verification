# Performance Testing Suite

This directory contains comprehensive performance tests for the Merchant-Centric UI implementation of the KYB Platform.

## Overview

The performance testing suite is designed to validate the system's ability to handle:
- **5000+ merchants** in the portfolio
- **20 concurrent users** (MVP target)
- **Bulk operations** on large datasets
- **Real-time search and filtering**
- **Session management** under load

## Test Structure

### 1. Core Performance Framework (`performance_test.go`)
- `PerformanceTestConfig`: Configuration for performance tests
- `PerformanceMetrics`: Metrics collection and analysis
- `PerformanceTestRunner`: Test execution management
- `ConcurrentUserSimulator`: Simulates concurrent user behavior
- `PerformanceTestSuite`: Manages test collections

### 2. Merchant Portfolio Performance Tests (`merchant_portfolio_performance_test.go`)
- **List Performance**: Tests listing merchants with large datasets
- **Search Performance**: Tests search operations with 5000+ merchants
- **Filtering Performance**: Tests filtering by portfolio type, risk level, industry
- **Pagination Performance**: Tests pagination with large datasets
- **Detail View Performance**: Tests individual merchant detail loading

### 3. Bulk Operations Performance Tests (`bulk_operations_performance_test.go`)
- **Bulk Update**: Tests updating 1000+ merchants in batches
- **Bulk Status Change**: Tests changing status for multiple merchants
- **Bulk Export**: Tests exporting large datasets (CSV, JSON, Excel)
- **Bulk Import**: Tests importing large datasets with validation
- **Bulk Deletion**: Tests deleting multiple merchants

### 4. Concurrent User Performance Tests (`concurrent_user_performance_test.go`)
- **Portfolio Access**: Tests 20 concurrent users accessing portfolios
- **Detail Views**: Tests concurrent merchant detail views
- **Search Operations**: Tests concurrent search operations
- **Bulk Operations**: Tests concurrent bulk operations
- **Session Management**: Tests concurrent session management

### 5. Performance Reporting (`performance_reporting.go`)
- **PerformanceReport**: Comprehensive test reporting
- **HTML Reports**: Visual performance reports
- **JSON Reports**: Machine-readable performance data
- **Summary Reports**: Text-based performance summaries
- **Recommendations**: Automated performance recommendations

### 6. Test Runner (`performance_test_runner.go`)
- **Test Orchestration**: Manages test execution
- **Report Generation**: Creates comprehensive reports
- **Benchmark Execution**: Runs performance benchmarks
- **Result Analysis**: Analyzes and compares results

## Configuration

### Default Configuration
```go
config := &PerformanceTestConfig{
    MaxMerchants:      5000,
    ConcurrentUsers:   20,
    TestDuration:      5 * time.Minute,
    BulkOperationSize: 1000,
    ResponseTimeLimit: 2 * time.Second,
}
```

### Performance Targets
- **Response Time**: < 2 seconds for most operations
- **Error Rate**: < 5% under normal load
- **Throughput**: > 10 requests per second
- **Concurrent Users**: Support 20 users (MVP target)
- **Bulk Operations**: Handle 1000+ merchants per operation

## Running Tests

### Run All Performance Tests
```bash
go test ./test/performance/... -v
```

### Run Specific Test Categories
```bash
# Merchant Portfolio Tests
go test ./test/performance/ -run TestMerchantPortfolioPerformance -v

# Bulk Operations Tests
go test ./test/performance/ -run TestBulkOperationsPerformance -v

# Concurrent User Tests
go test ./test/performance/ -run TestConcurrentUserPerformance -v
```

### Run Benchmarks
```bash
go test ./test/performance/ -bench=. -benchmem
```

### Run with Custom Configuration
```bash
go test ./test/performance/ -v -timeout 30m
```

## Test Reports

### Report Locations
- **JSON Reports**: `test-results/performance/performance_report_YYYY-MM-DD_HH-MM-SS.json`
- **HTML Reports**: `test-results/performance/performance_report_YYYY-MM-DD_HH-MM-SS.html`
- **Summary Reports**: `test-results/performance/performance_summary.txt`
- **Benchmark Results**: `test-results/benchmarks/benchmark_results.json`

### Report Contents
- **Test Summary**: Pass/fail counts, success rates
- **Performance Metrics**: Response times, throughput, error rates
- **Test Details**: Individual test results and metrics
- **Recommendations**: Automated performance improvement suggestions
- **Benchmark Comparisons**: Performance comparisons between operations

## Performance Monitoring

### Key Metrics
- **Response Time**: Average, min, max response times
- **Throughput**: Requests per second
- **Error Rate**: Percentage of failed requests
- **Resource Usage**: Memory and CPU utilization
- **Concurrent Load**: Performance under concurrent user load

### Performance Thresholds
- **Green**: All metrics within acceptable ranges
- **Yellow**: Some metrics approaching limits
- **Red**: Metrics exceeding acceptable thresholds

## Troubleshooting

### Common Issues
1. **High Response Times**: Check database queries, implement caching
2. **High Error Rates**: Review error logs, check resource limits
3. **Low Throughput**: Optimize database connections, implement connection pooling
4. **Memory Issues**: Review data structures, implement pagination

### Performance Optimization Tips
1. **Database Optimization**: Add indexes, optimize queries
2. **Caching**: Implement Redis caching for frequently accessed data
3. **Connection Pooling**: Use connection pools for database access
4. **Pagination**: Implement efficient pagination for large datasets
5. **Async Processing**: Use background processing for bulk operations

## Integration with CI/CD

### Continuous Performance Testing
```yaml
# Example GitHub Actions workflow
- name: Run Performance Tests
  run: |
    go test ./test/performance/... -v -timeout 30m
    # Generate performance reports
    # Upload reports to monitoring system
```

### Performance Regression Detection
- Compare current results with baseline
- Alert on performance degradation
- Track performance trends over time

## Best Practices

### Test Design
- Use realistic test data
- Simulate real user behavior patterns
- Test edge cases and boundary conditions
- Include both happy path and error scenarios

### Test Execution
- Run tests in isolated environments
- Use consistent test configurations
- Monitor system resources during tests
- Document test results and findings

### Performance Analysis
- Focus on user-facing metrics
- Consider business impact of performance issues
- Prioritize optimization based on usage patterns
- Regular performance reviews and improvements

## Future Enhancements

### Planned Improvements
- **Load Testing**: Integration with external load testing tools
- **Real-time Monitoring**: Live performance monitoring during tests
- **Automated Optimization**: AI-driven performance optimization suggestions
- **Performance Budgets**: Automated performance budget enforcement
- **Distributed Testing**: Multi-region performance testing

### Scalability Testing
- Test with 100+ concurrent users
- Test with 50,000+ merchants
- Test with complex bulk operations
- Test with real-time data updates

## Support

For questions or issues with performance testing:
1. Check the test logs for detailed error information
2. Review the performance reports for optimization suggestions
3. Consult the troubleshooting section above
4. Create an issue with detailed test configuration and results
