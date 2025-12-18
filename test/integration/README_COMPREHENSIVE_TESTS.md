# Comprehensive Classification E2E Test Suite

## Overview

This test suite provides comprehensive end-to-end testing of the classification flow, testing 100 diverse samples across the entire pipeline from scraping/crawling to classification, including frontend data verification.

## Test Coverage

- **100 Diverse Samples**: Covering multiple industries, complexities, and scraping difficulties
- **Scraping Strategies**: Tests hrequests, SimpleHTTP, BrowserHeaders, and Playwright fallbacks
- **Parallel Smart Crawling**: Verifies parallel execution improves performance
- **Early Exit Logic**: Tests early exit when high-quality content is obtained
- **Conditional Fallbacks**: Tests fallback strategies when needed
- **Classification Accuracy**: Verifies ≥95% correct industry classification
- **Industry Codes**: Validates top 3 MCC, NAICS, and SIC codes
- **Explanation Verification**: Ensures explanations are present and meaningful
- **Frontend Data Format**: Verifies data structure matches frontend expectations
- **Performance Metrics**: Measures latency, throughput, and speed
- **Cache Effectiveness**: Measures cache hit rates
- **Error Handling**: Tests graceful error handling

## Files

- `comprehensive_classification_e2e_test.go` - Main test runner implementation
- `test_report_generator.go` - Report generation utilities
- `../data/comprehensive_test_samples.json` - 100 test samples dataset
- `../scripts/run_comprehensive_tests.sh` - Test execution script

## Prerequisites

1. **Classification Service Running**: The classification service must be running and accessible
2. **Environment Variables**:
   - `CLASSIFICATION_API_URL` (default: `http://localhost:8081`)

## Running the Tests

### Option 1: Using the Shell Script (Local)

```bash
# Set API URL (if different from default)
export CLASSIFICATION_API_URL=http://localhost:8081

# Run tests
./test/scripts/run_comprehensive_tests.sh
```

### Option 2: Using the Railway Production Script

```bash
# Run tests against Railway production
./test/scripts/run_comprehensive_tests_railway.sh

# Or use API Gateway endpoint
USE_API_GATEWAY=true ./test/scripts/run_comprehensive_tests_railway.sh
```

The Railway script will:

- Use Railway production URLs automatically
- Verify service health before running tests
- Warn you that you're testing against production
- Use extended timeouts for production environment

The script will:

- Check API availability
- Run all 100 test samples
- Generate JSON report
- Display summary statistics

### Option 3: Using Go Test Directly

```bash
# For localhost
export CLASSIFICATION_API_URL=http://localhost:8081
go test -v -timeout 30m ./test/integration -run TestComprehensiveClassificationE2E

# For Railway production
export CLASSIFICATION_API_URL=https://classification-service-production.up.railway.app
go test -v -timeout 60m ./test/integration -run TestComprehensiveClassificationE2E
```

### Option 4: Running Specific Number of Samples

To test with fewer samples, modify the test file or create a subset of the JSON dataset.

## Railway Production Testing

### Railway Service URLs

- **Classification Service**: `https://classification-service-production.up.railway.app`
- **API Gateway**: `https://api-gateway-service-production-21fd.up.railway.app`

### Testing Against Railway Production

**⚠️ Important**: Testing against production will make real API calls and may impact production metrics.

1. **Use the Railway-specific script** (recommended):

   ```bash
   ./test/scripts/run_comprehensive_tests_railway.sh
   ```

2. **Or set environment variable**:

   ```bash
   export CLASSIFICATION_API_URL=https://classification-service-production.up.railway.app
   go test -v -timeout 60m ./test/integration -run TestComprehensiveClassificationE2E
   ```

3. **Via API Gateway** (if you want to test the full routing):
   ```bash
   USE_API_GATEWAY=true ./test/scripts/run_comprehensive_tests_railway.sh
   ```

### Railway-Specific Considerations

- **Extended Timeout**: Railway tests use 60-minute timeout (vs 30 minutes for local)
- **Network Latency**: Expect higher latency due to network distance
- **Rate Limiting**: Be aware of production rate limits
- **Cache Behavior**: Production cache may affect results
- **Cost**: Each test makes real API calls that may incur costs

## Test Results

Results are saved to:

- `test/results/comprehensive_test_results.json` - Detailed JSON report

The JSON report includes:

- Test summary (total, successful, failed, accuracy)
- Performance metrics (latency percentiles, throughput)
- Strategy distribution (usage, success rates, latencies)
- Optimization metrics (early exit, cache hits, fallbacks)
- Frontend compatibility metrics
- Code accuracy metrics
- Detailed results for each sample

## Success Criteria

### Must Pass (Critical)

- ✅ Overall accuracy ≥ 95%
- ✅ All required frontend fields present
- ✅ Average latency < 2s
- ✅ P95 latency < 5s
- ✅ No crashes or panics
- ✅ Error handling works correctly

### Should Pass (Important)

- ✅ hrequests usage: 60-70%
- ✅ Early exit rate: 20-30%
- ✅ Cache hit rate: 60-70%
- ✅ Code accuracy: Top 3 codes match ≥ 80%
- ✅ Explanation present: ≥ 95%

## Test Sample Distribution

The 100 samples are distributed as follows:

**By Industry:**

- Technology: 20 samples
- Healthcare: 20 samples
- Financial Services: 20 samples
- Retail & Commerce: 20 samples
- Manufacturing: 10 samples
- Education: 5 samples
- Real Estate: 5 samples

**By Complexity:**

- Simple: 40 samples
- Medium: 35 samples
- Complex: 25 samples

**By Scraping Difficulty:**

- Easy (hrequests should work): 60 samples
- Medium (may need Playwright): 30 samples
- Hard (challenging sites): 10 samples

## Interpreting Results

### Performance Metrics

- **Average Latency**: Should be < 1.5s
- **P95 Latency**: Should be < 3s
- **P99 Latency**: Should be < 5s
- **Throughput**: Should be ≥ 20 req/s

### Strategy Distribution

- **hrequests**: Should handle 60-70% of requests
- **Playwright**: Should handle 20-30% as fallback
- **SimpleHTTP/BrowserHeaders**: Should handle 10-20%

### Optimization Metrics

- **Early Exit Rate**: 20-30% indicates good optimization
- **Cache Hit Rate**: 60-70% indicates effective caching
- **Fallback Usage**: Should be minimal and only when needed

### Frontend Compatibility

All metrics should be ≥ 95%:

- All Fields Present
- Industry Present
- Codes Present
- Explanation Present
- Top 3 Codes Present

## Troubleshooting

### API Not Accessible

If the API is not accessible:

1. Ensure the classification service is running
2. Check the `CLASSIFICATION_API_URL` environment variable
3. Verify network connectivity

### Tests Timing Out

If tests are timing out:

1. Increase timeout: `-timeout 60m`
2. Check service performance
3. Verify database connectivity

### Low Accuracy

If accuracy is below 95%:

1. Check classification service logs
2. Verify database has correct industry mappings
3. Review failed test cases in detailed results

### Low hrequests Usage

If hrequests usage is below 60%:

1. Check hrequests service is running
2. Verify `HREQUESTS_SERVICE_URL` is set correctly
3. Review scraping strategy selection logic

## Continuous Integration

To run these tests in CI/CD:

```yaml
# Example GitHub Actions workflow
- name: Run Comprehensive E2E Tests
  run: |
    export CLASSIFICATION_API_URL=${{ secrets.CLASSIFICATION_API_URL }}
    go test -v -timeout 30m ./test/integration -run TestComprehensiveClassificationE2E
```

## Additional Resources

- See `docs/COMPREHENSIVE_CLASSIFICATION_TEST_PLAN.md` for detailed test plan
- See `test/data/comprehensive_test_samples.json` for sample data structure
- See `test/results/` for generated reports

## Support

For issues or questions:

1. Check test logs in `test/results/test_output_*.txt`
2. Review detailed results in JSON report
3. Check classification service logs
