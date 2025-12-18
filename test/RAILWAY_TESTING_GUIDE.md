# Railway Production Testing Guide

## Quick Start

### Run Tests Against Railway Production

```bash
# Simple - Direct Classification Service
./test/scripts/run_comprehensive_tests_railway.sh

# Via API Gateway (tests full routing)
USE_API_GATEWAY=true ./test/scripts/run_comprehensive_tests_railway.sh
```

## Railway Production URLs

- **Classification Service**: `https://classification-service-production.up.railway.app`
- **API Gateway**: `https://api-gateway-service-production-21fd.up.railway.app`

## What the Script Does

1. ✅ **Health Check**: Verifies the service is accessible before running tests
2. ⚠️ **Production Warning**: Prompts you to confirm you want to test production
3. 🚀 **Extended Timeout**: Uses 60-minute timeout (vs 30 minutes for local)
4. 📊 **Results**: Generates comprehensive JSON report and test output log

## Test Endpoints

### Direct Classification Service
- **Health**: `https://classification-service-production.up.railway.app/health`
- **Classify**: `https://classification-service-production.up.railway.app/v1/classify`

### Via API Gateway
- **Health**: `https://api-gateway-service-production-21fd.up.railway.app/health`
- **Classify**: `https://api-gateway-service-production-21fd.up.railway.app/api/v1/classify`

## Important Notes

⚠️ **Production Testing Considerations**:

1. **Real API Calls**: Tests make actual requests to production services
2. **Rate Limiting**: Be aware of production rate limits
3. **Cost**: May incur costs for API calls and scraping
4. **Cache Impact**: Production cache may affect test results
5. **Network Latency**: Expect higher latency than localhost
6. **Service Impact**: Tests may impact production metrics and logs

## Expected Results

### Performance (Railway Production)
- **Average Latency**: 1.5-3s (higher than localhost due to network)
- **P95 Latency**: 3-6s
- **Throughput**: 5-15 req/s (network dependent)

### Success Criteria
- ✅ Overall accuracy ≥ 95%
- ✅ All frontend fields present
- ✅ hrequests usage: 60-70%
- ✅ Cache hit rate: 60-70% (if running multiple times)

## Troubleshooting

### Service Not Accessible

```bash
# Test health endpoint manually
curl https://classification-service-production.up.railway.app/health

# Check Railway dashboard for service status
```

### Timeout Issues

The script uses 60-minute timeout. If tests still timeout:
1. Check Railway service logs
2. Verify service is not under heavy load
3. Consider running tests in smaller batches

### Rate Limiting

If you hit rate limits:
1. Wait a few minutes between test runs
2. Check Railway service rate limit settings
3. Consider running tests during off-peak hours

## Output Files

After running tests, check:
- `test/results/comprehensive_test_results.json` - Detailed results
- `test/results/test_output_railway_*.txt` - Full test log

## Alternative: Manual Testing

If you prefer to run manually:

```bash
# Set Railway URL
export CLASSIFICATION_API_URL=https://classification-service-production.up.railway.app

# Run tests
go test -v -timeout 60m ./test/integration -run TestComprehensiveClassificationE2E
```

## CI/CD Integration

For automated Railway testing in CI/CD:

```yaml
# Example GitHub Actions
- name: Run Railway Production Tests
  env:
    CLASSIFICATION_API_URL: https://classification-service-production.up.railway.app
  run: |
    go test -v -timeout 60m ./test/integration -run TestComprehensiveClassificationE2E
```

## Support

For issues:
1. Check Railway service logs
2. Verify service health endpoint
3. Review test output logs
4. Check Railway dashboard for service status

