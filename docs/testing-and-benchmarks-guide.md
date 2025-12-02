# Testing and Benchmarks Guide

## Overview

This guide provides instructions for running the comprehensive test suite and benchmarks for the classification service optimizations.

## Test Files Created

### Unit Tests

1. **Keyword Extraction Accuracy Tests**
   - File: `internal/classification/smart_website_crawler_keyword_accuracy_test.go`
   - Tests: Enhanced word validation, suspicious pattern detection, n-gram validation
   - Status: ✅ Ready

2. **Keyword Gibberish Filter Tests**
   - File: `internal/classification/repository/keyword_gibberish_filter_test.go`
   - Tests: Post-processing filter, known gibberish word removal
   - Status: ✅ Ready

3. **Website Content Cache Tests**
   - File: `services/classification-service/internal/cache/website_content_cache_test.go`
   - Tests: Cache get/set operations, expiration, deletion
   - Status: ✅ Ready

4. **Website Content Service Tests**
   - File: `internal/classification/website_content_service_test.go`
   - Tests: Content sufficiency validation
   - Status: ✅ Ready

5. **Early Termination Tests**
   - File: `services/classification-service/internal/handlers/early_termination_test.go`
   - Tests: Early termination logic, threshold configuration
   - Status: ✅ Ready

6. **Parallel Classification Tests**
   - File: `services/classification-service/internal/handlers/parallel_classification_test.go`
   - Tests: Ensemble voting, consensus boost, parallel execution
   - Status: ✅ Ready

### Integration Tests

7. **Classification Optimizations Integration Tests**
   - File: `test/integration/classification_optimizations_integration_test.go`
   - Tests: End-to-end pipeline, request deduplication, Redis cache, ensemble voting, smart crawling
   - Status: ⚠️ Requires environment setup

### Performance Benchmarks

8. **Classification Optimization Benchmarks**
   - File: `test/benchmark/classification_optimization_benchmarks_test.go`
   - Benchmarks: Keyword extraction, cache operations, deduplication, parallel execution
   - Status: ✅ Ready

## Running Tests

### Unit Tests

```bash
# Keyword extraction accuracy tests
go test -v ./internal/classification -run TestIsValidEnglishWord
go test -v ./internal/classification -run TestHasSuspiciousPatterns
go test -v ./internal/classification -run TestHasValidNgramPatterns
go test -v ./internal/classification -run TestLoadCommonEnglishWords

# Keyword gibberish filter tests
go test -v ./internal/classification/repository -run TestFilterGibberishKeywords
go test -v ./internal/classification/repository -run TestHasSuspiciousPattern
go test -v ./internal/classification/repository -run TestHasValidNgramPattern

# Website content cache tests
go test -v ./services/classification-service/internal/cache -run TestWebsiteContentCache

# Website content service tests
go test -v ./internal/classification -run TestIsContentSufficient

# Early termination tests
go test -v ./services/classification-service/internal/handlers -run TestEarlyTermination
go test -v ./services/classification-service/internal/handlers -run TestContentQualityValidation

# Parallel classification tests
go test -v ./services/classification-service/internal/handlers -run TestParallelClassification
go test -v ./services/classification-service/internal/handlers -run TestEnsembleVoting
```

### Run All Unit Tests

```bash
# Run all classification tests
go test -v ./internal/classification/...

# Run all handler tests
go test -v ./services/classification-service/internal/handlers/...

# Run all cache tests
go test -v ./services/classification-service/internal/cache/...
```

### Integration Tests

```bash
# Set environment variable
export INTEGRATION_TESTS=true
export REDIS_URL=redis://localhost:6379  # Optional, for Redis tests

# Run integration tests
go test -v ./test/integration -run TestClassificationOptimizations
```

### Performance Benchmarks

```bash
# Run all benchmarks
go test -bench=. -benchmem ./test/benchmark

# Run specific benchmarks
go test -bench=BenchmarkKeywordExtraction -benchmem ./test/benchmark
go test -bench=BenchmarkWebsiteContentCache -benchmem ./test/benchmark
go test -bench=BenchmarkParallelClassification -benchmem ./test/benchmark

# Compare sequential vs parallel
go test -bench=BenchmarkParallelClassification_Speedup -benchmem ./test/benchmark
```

## Test Coverage

### Expected Coverage Targets

- **Keyword Extraction**: > 90%
- **Cache Operations**: > 85%
- **Early Termination**: > 80%
- **Parallel Classification**: > 75%
- **Overall**: > 80%

### Generate Coverage Report

```bash
# Generate coverage for all tests
go test -coverprofile=coverage.out ./internal/classification/... ./services/classification-service/...

# View coverage report
go tool cover -html=coverage.out

# Coverage by package
go test -cover ./internal/classification/...
go test -cover ./services/classification-service/internal/handlers/...
go test -cover ./services/classification-service/internal/cache/...
```

## Benchmark Results Analysis

### Key Metrics to Track

1. **Keyword Extraction Performance**
   - Time per word validation
   - Dictionary lookup speed
   - Pattern detection overhead

2. **Cache Performance**
   - Get/Set operation latency
   - Cache hit rate
   - Memory usage

3. **Deduplication Performance**
   - Concurrent request handling
   - Overhead of deduplication logic
   - Memory usage for in-flight tracking

4. **Parallel Execution Speedup**
   - Sequential vs parallel execution time
   - Expected: 1.5-2x speedup for 2 parallel operations

### Benchmark Comparison

```bash
# Run benchmarks and save results
go test -bench=. -benchmem -count=5 ./test/benchmark > benchmark_results.txt

# Compare with previous results
benchcmp old_results.txt benchmark_results.txt
```

## Test Data

### Sample Test Cases

1. **Known Gibberish Words**
   - "ivdi", "fays", "yilp", "dioy", "ukxa"
   - Should be filtered by validation

2. **Valid Business Keywords**
   - "business", "technology", "restaurant", "retail"
   - Should pass validation

3. **Edge Cases**
   - Short words (< 4 chars)
   - Words with repeated letters
   - Words with rare letter combinations

## Troubleshooting

### Common Issues

1. **Redis Connection Errors**
   - Ensure Redis is running or tests will use fallback cache
   - Set `REDIS_URL` environment variable

2. **Integration Test Failures**
   - Ensure `INTEGRATION_TESTS=true` is set
   - Check database and external service availability

3. **Benchmark Timeouts**
   - Increase timeout for slow benchmarks
   - Use `-timeout` flag: `go test -timeout=10m -bench=.`

### Debug Mode

```bash
# Run tests with verbose output
go test -v -args -test.v

# Run specific test with debug logging
go test -v -run TestIsValidEnglishWord ./internal/classification
```

## Continuous Integration

### CI Test Commands

```bash
# Fast unit tests (no external dependencies)
go test -short ./internal/classification/... ./services/classification-service/...

# Full test suite (requires Redis, database)
go test ./internal/classification/... ./services/classification-service/... ./test/integration/...

# Benchmarks (optional, can be slow)
go test -bench=. -benchtime=1s ./test/benchmark
```

## Next Steps

1. **Run Unit Tests**: Verify all optimizations work correctly
2. **Run Integration Tests**: Verify end-to-end pipeline (requires environment setup)
3. **Run Benchmarks**: Measure performance improvements
4. **Compare Results**: Compare against baseline metrics
5. **Document Findings**: Update performance documentation with results

