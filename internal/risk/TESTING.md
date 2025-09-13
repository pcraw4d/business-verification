# Enhanced Risk Services Testing Guide

This document provides comprehensive guidance for testing the enhanced risk services in the KYB Platform.

## Overview

The enhanced risk services testing suite includes:
- Unit tests for individual components
- Integration tests for complete service workflows
- Performance benchmarks
- Concurrency and thread safety tests
- Memory usage and garbage collection tests
- Error handling and edge case tests

## Test Structure

### Test Files

- `enhanced_risk_test.go` - Unit tests for API handlers
- `enhanced_risk_integration_test.go` - Integration tests for complete workflows
- `enhanced_risk_performance_test.go` - Performance benchmarks
- `test_config.go` - Test configuration and mock data generation
- `test_runner.go` - Test orchestration and execution
- `enhanced_risk_test_suite.go` - Comprehensive test suite with multiple test types

### Test Types

#### 1. Unit Tests
Test individual components in isolation:
- `EnhancedRiskFactorCalculator`
- `RecommendationEngine`
- `TrendAnalysisService`
- `CorrelationAnalyzer`
- `ConfidenceCalibrator`
- `RiskAlertSystem`

#### 2. Integration Tests
Test complete workflows:
- End-to-end risk assessment
- Service factory creation
- Component interaction

#### 3. Performance Tests
Benchmark performance:
- Service response times
- Memory usage patterns
- Throughput measurements

#### 4. Concurrency Tests
Test thread safety:
- Concurrent access patterns
- Race condition detection
- Deadlock prevention

#### 5. Memory Tests
Test memory management:
- Memory usage patterns
- Garbage collection behavior
- Memory leak detection

## Running Tests

### Basic Test Execution

```bash
# Run all tests
go test ./internal/risk/...

# Run specific test file
go test ./internal/risk/enhanced_risk_test.go

# Run specific test function
go test -run TestEnhancedRiskServicesSmoke ./internal/risk/
```

### Test Suite Execution

```bash
# Run complete test suite
go test -run TestEnhancedRiskServicesTestSuite ./internal/risk/

# Run quick test suite (unit + integration only)
go test -run TestEnhancedRiskServicesQuick ./internal/risk/

# Run performance test suite
go test -run TestEnhancedRiskServicesPerformance ./internal/risk/

# Run stress test suite
go test -run TestEnhancedRiskServicesStress ./internal/risk/
```

### Test with Flags

```bash
# Run with custom configuration
go test -run TestEnhancedRiskServicesMainWithFlags ./internal/risk/ \
  -unit=true \
  -integration=true \
  -performance=false \
  -benchmark-iterations=500 \
  -test-timeout=60s \
  -concurrent-goroutines=20 \
  -memory-assessments=200 \
  -log-level=info \
  -log-format=console
```

### Test with Environment Variables

```bash
# Set environment variables and run tests
ENHANCED_RISK_UNIT_TESTS=true \
ENHANCED_RISK_INTEGRATION_TESTS=true \
ENHANCED_RISK_PERFORMANCE_TESTS=false \
go test -run TestEnhancedRiskServicesMainWithEnv ./internal/risk/
```

### Benchmark Execution

```bash
# Run all benchmarks
go test -bench=. ./internal/risk/

# Run specific benchmark
go test -bench=BenchmarkEnhancedRiskService ./internal/risk/

# Run benchmark with custom iterations
go test -bench=BenchmarkEnhancedRiskService -benchtime=10s ./internal/risk/

# Run benchmark with memory profiling
go test -bench=BenchmarkEnhancedRiskService -benchmem ./internal/risk/
```

## Test Configuration

### Default Configuration

```go
type TestSuiteConfig struct {
    RunUnitTests         bool          // Run unit tests
    RunIntegrationTests  bool          // Run integration tests
    RunPerformanceTests  bool          // Run performance tests
    RunConcurrencyTests  bool          // Run concurrency tests
    RunMemoryTests       bool          // Run memory tests
    BenchmarkIterations  int           // Number of benchmark iterations
    TestTimeout          time.Duration // Test timeout duration
    ConcurrentGoroutines int           // Number of concurrent goroutines
    MemoryTestAssessments int          // Number of memory test assessments
    LogLevel             string        // Log level (debug, info, warn, error)
    LogFormat            string        // Log format (json, console)
}
```

### Custom Configuration

You can create custom test configurations by modifying the `TestSuiteConfig` struct or using the provided configuration methods.

## Mock Data Generation

The test suite includes comprehensive mock data generation for:
- Risk factor inputs
- Historical risk data
- Recommendations
- Alerts
- Time ranges

### Example Mock Data Usage

```go
config := DefaultTestConfig()
dataGenerator := NewTestDataGenerator(config)

// Generate mock risk factors
factors := dataGenerator.GenerateMockRiskFactors()

// Generate mock historical data
historicalData := dataGenerator.GenerateMockHistoricalData("test-business")

// Generate mock recommendations
recommendations := dataGenerator.GenerateMockRecommendations()

// Generate mock alerts
alerts := dataGenerator.GenerateMockAlerts("test-business")
```

## Test Coverage

### Coverage Requirements

- Unit tests: 90%+ coverage
- Integration tests: 80%+ coverage
- Performance tests: All critical paths
- Error handling: 100% coverage

### Running Coverage Tests

```bash
# Run tests with coverage
go test -cover ./internal/risk/...

# Run tests with detailed coverage
go test -coverprofile=coverage.out ./internal/risk/
go tool cover -html=coverage.out

# Run tests with coverage and benchmarks
go test -cover -bench=. ./internal/risk/
```

## Performance Testing

### Benchmark Targets

- Risk assessment: < 100ms average
- Factor calculation: < 10ms average
- Recommendation generation: < 50ms average
- Trend analysis: < 200ms average
- Alert checking: < 20ms average

### Performance Test Execution

```bash
# Run performance benchmarks
go test -bench=BenchmarkEnhancedRiskService ./internal/risk/

# Run with memory profiling
go test -bench=BenchmarkEnhancedRiskService -benchmem ./internal/risk/

# Run with CPU profiling
go test -bench=BenchmarkEnhancedRiskService -cpuprofile=cpu.prof ./internal/risk/
```

## Concurrency Testing

### Concurrency Test Execution

```bash
# Run concurrency tests
go test -run TestEnhancedRiskServicesConcurrentAccess ./internal/risk/

# Run with race detection
go test -race ./internal/risk/

# Run with race detection and benchmarks
go test -race -bench=. ./internal/risk/
```

## Memory Testing

### Memory Test Execution

```bash
# Run memory tests
go test -run TestEnhancedRiskServicesMemoryUsage ./internal/risk/

# Run with memory profiling
go test -run TestEnhancedRiskServicesMemoryUsage -memprofile=mem.prof ./internal/risk/

# Run with heap profiling
go test -run TestEnhancedRiskServicesMemoryUsage -heap=heap.prof ./internal/risk/
```

## Error Handling Testing

### Error Test Execution

```bash
# Run error handling tests
go test -run TestEnhancedRiskServicesErrorHandling ./internal/risk/

# Run edge case tests
go test -run TestEnhancedRiskServicesEdgeCases ./internal/risk/
```

## Continuous Integration

### CI Test Execution

```bash
# Run CI test suite
go test -run TestEnhancedRiskServicesTestSuite ./internal/risk/ \
  -unit=true \
  -integration=true \
  -performance=true \
  -concurrency=true \
  -memory=true \
  -benchmark-iterations=1000 \
  -test-timeout=300s \
  -concurrent-goroutines=10 \
  -memory-assessments=100 \
  -log-level=info \
  -log-format=json
```

## Test Data Management

### Test Data Cleanup

The test suite automatically cleans up test data after each test run. No manual cleanup is required.

### Test Data Isolation

Each test run uses isolated test data to prevent interference between tests.

## Debugging Tests

### Debug Mode

```bash
# Run tests in debug mode
go test -run TestEnhancedRiskServicesTestSuite ./internal/risk/ \
  -log-level=debug \
  -log-format=console
```

### Verbose Output

```bash
# Run tests with verbose output
go test -v ./internal/risk/

# Run specific test with verbose output
go test -v -run TestEnhancedRiskServicesSmoke ./internal/risk/
```

## Test Maintenance

### Adding New Tests

1. Create test functions following the naming convention `Test*`
2. Add test cases to the appropriate test file
3. Update test configuration if needed
4. Add test documentation

### Updating Existing Tests

1. Modify test functions as needed
2. Update test data if required
3. Update test documentation
4. Run full test suite to ensure no regressions

### Test Data Updates

1. Update mock data generation in `test_config.go`
2. Update test expectations if needed
3. Run tests to ensure data changes work correctly

## Troubleshooting

### Common Issues

1. **Test Timeouts**: Increase `TestTimeout` in configuration
2. **Memory Issues**: Reduce `MemoryTestAssessments` or `ConcurrentGoroutines`
3. **Performance Issues**: Reduce `BenchmarkIterations`
4. **Race Conditions**: Run tests with `-race` flag

### Test Failures

1. Check test logs for detailed error information
2. Verify test configuration
3. Check for environment-specific issues
4. Review test data generation

### Performance Issues

1. Profile tests with `-cpuprofile` and `-memprofile`
2. Check for memory leaks
3. Optimize test data generation
4. Review test configuration

## Best Practices

### Test Writing

1. Write clear, descriptive test names
2. Use table-driven tests where appropriate
3. Test both success and failure cases
4. Include edge cases and boundary conditions
5. Use meaningful assertions

### Test Organization

1. Group related tests together
2. Use subtests for complex test scenarios
3. Keep tests focused and single-purpose
4. Use consistent naming conventions

### Test Data

1. Use realistic test data
2. Generate test data programmatically
3. Ensure test data isolation
4. Clean up test data after tests

### Performance Testing

1. Set realistic performance targets
2. Test under various load conditions
3. Monitor memory usage
4. Profile performance bottlenecks

### Concurrency Testing

1. Test with various concurrency levels
2. Use race detection
3. Test for deadlocks
4. Verify thread safety

## Conclusion

This testing guide provides comprehensive coverage of the enhanced risk services testing suite. Follow the guidelines and best practices to ensure robust, reliable, and performant risk assessment services.

For additional support or questions, refer to the main project documentation or contact the development team.
