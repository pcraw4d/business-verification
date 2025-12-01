# ML Service Implementation - Next Steps Guide

**Date**: 2025-01-19  
**Status**: Implementation Complete - Ready for Testing

---

## Overview

This guide provides step-by-step instructions for completing the remaining tasks from the ML Service Accuracy Implementation Review. All code changes have been completed; these tasks focus on testing, monitoring, and validation.

---

## Task 1: Run Accuracy Tests ✅ Ready

### Objective
Verify that the improvements (circuit breaker enhancements, fallback classifier improvements, website scraping optimization) meet the success criteria.

### Prerequisites
- Python ML service running (optional - tests will use fallback if unavailable)
- Database connection configured
- Test dataset populated

### Steps

#### Option A: Automated Script (Recommended)

```bash
# Run the automated script
./scripts/run_ml_accuracy_tests.sh
```

This script will:
1. Check if Python ML service is running
2. Start it if needed
3. Wait for it to be ready
4. Run accuracy tests with ML support
5. Save results to timestamped JSON file

#### Option B: Manual Execution

```bash
# 1. Start Python ML service (if available)
export PYTHON_ML_SERVICE_URL="http://localhost:8000"
cd python_ml_service
python app.py &
cd ..

# 2. Build the test binary
go build -o bin/comprehensive_accuracy_test ./cmd/comprehensive_accuracy_test

# 3. Run accuracy tests
./bin/comprehensive_accuracy_test -verbose -output accuracy_report_$(date +%Y%m%d_%H%M%S).json
```

### Expected Results

Based on the implementation improvements, you should see:

- **Circuit Breaker**: Should recover within 60 seconds if it opens
- **Fallback Classifier**: Should use keyword-based classification when ML fails (improved from 0% accuracy)
- **Website Scraping**: Should complete in < 5 seconds (reduced from 15s)
- **Cache Hit Rate**: Should see cache hits for repeated website URLs

### Success Criteria

- Industry accuracy > 50% (Week 1 target)
- Code accuracy > 40% (Week 1 target)
- ML service utilization > 80% (when available)
- Fallback classifier accuracy > 30%

### Analyzing Results

```bash
# View the generated JSON report
cat accuracy_report_*.json | jq '.summary'

# Compare with previous results
diff accuracy_report_v3.json accuracy_report_*.json
```

---

## Task 2: Monitor Circuit Breaker State 🔍 Ready

### Objective
Monitor the circuit breaker state in production to ensure it's functioning correctly and recovering from failures.

### Prerequisites
- Python ML service running
- Access to service health endpoint

### Steps

#### Option A: Using Monitoring Script

```bash
# Run the monitoring script
./scripts/monitor_circuit_breaker.sh

# For detailed metrics
./scripts/monitor_circuit_breaker.sh --detailed

# Continuous monitoring (updates every 5 seconds)
watch -n 5 ./scripts/monitor_circuit_breaker.sh
```

#### Option B: Manual Health Check

```bash
# Set service URL
export PYTHON_ML_SERVICE_URL="http://localhost:8000"

# Check health with circuit breaker info
curl -s "${PYTHON_ML_SERVICE_URL}/health" | jq '.checks.circuit_breaker'
```

#### Option C: Programmatic Access

If you have access to the classification service code, you can access circuit breaker metrics:

```go
// Get circuit breaker state
state := pythonMLService.GetCircuitBreakerState()

// Get comprehensive metrics
metrics := pythonMLService.GetCircuitBreakerMetrics()

// Health check with circuit breaker info
health, err := pythonMLService.HealthCheckWithCircuitBreaker(ctx)
```

### What to Monitor

1. **Circuit Breaker State**:
   - `CLOSED`: Normal operation (good)
   - `HALF_OPEN`: Testing recovery (transitional)
   - `OPEN`: Circuit is open, requests rejected (needs attention)

2. **Metrics**:
   - Failure count
   - Success count
   - State change time
   - Last failure time
   - Total requests vs rejected requests

3. **Recovery Time**:
   - Should recover within 60 seconds (timeout configured)
   - Should transition from OPEN → HALF_OPEN → CLOSED

### Alerting

Set up alerts for:
- Circuit breaker state = OPEN for > 2 minutes
- Failure rate > 50%
- Recovery time > 60 seconds

---

## Task 3: Performance Testing ⚡ Ready

### Objective
Verify that performance improvements (website scraping optimization, caching) meet the targets.

### Prerequisites
- Classification service running
- Python ML service running (optional)

### Steps

#### Using Performance Test Script

```bash
# Run performance test
./scripts/performance_test_classification.sh

# Customize test parameters
NUM_REQUESTS=20 CONCURRENT_REQUESTS=5 ./scripts/performance_test_classification.sh
```

#### Manual Performance Testing

```bash
# Test single request timing
time curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"business_name":"Test Business","description":"Test description","website_url":"https://example.com"}' \
  http://localhost:8080/v1/classify

# Test with website scraping (should be < 5s with caching)
time curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"business_name":"Acme Corp","description":"Technology company","website_url":"https://www.acme.com"}' \
  http://localhost:8080/v1/classify
```

### Performance Targets

- **Average processing time**: < 5 seconds ✅
- **P95 processing time**: < 8 seconds ✅
- **P99 processing time**: < 12 seconds ✅
- **Website scraping time**: < 3 seconds (95th percentile) ✅
- **Cache hit rate**: > 30% ✅

### What to Test

1. **First Request** (cold start):
   - Website scraping should complete in < 5s
   - Database queries should be fast (cached)

2. **Subsequent Requests** (warm):
   - Website content should be cached (instant)
   - Database queries should use cache
   - Overall latency should be < 2s

3. **Concurrent Requests**:
   - Should handle multiple requests in parallel
   - Cache should reduce load on external services

---

## Task 4: Observability Dashboard 📊 Future Work

### Objective
Create a dashboard to visualize metrics and monitor system health.

### Recommended Tools

1. **Prometheus + Grafana**:
   - Export metrics from classification service
   - Create dashboards for circuit breaker state, accuracy metrics, performance

2. **Custom Dashboard**:
   - Create HTTP endpoints for metrics
   - Build simple web dashboard
   - Use existing health check endpoints

### Metrics to Display

1. **Circuit Breaker Metrics**:
   - Current state (CLOSED/OPEN/HALF_OPEN)
   - Failure count
   - Success count
   - State change history

2. **Classification Metrics**:
   - Total classifications
   - Accuracy by method (ML/keyword/fallback)
   - Accuracy by industry
   - Processing time distribution

3. **Performance Metrics**:
   - Average/P95/P99 latency
   - Website scraping time
   - Cache hit rate
   - Database query time

### Implementation Steps

1. **Export Metrics** (if not already done):
   - Add Prometheus metrics exporter
   - Expose `/metrics` endpoint

2. **Create Dashboard**:
   - Set up Grafana
   - Import/create dashboard templates
   - Configure alerts

3. **Monitor**:
   - Set up alerting rules
   - Create runbooks for common issues

---

## Quick Reference

### Environment Variables

```bash
# Python ML Service URL
export PYTHON_ML_SERVICE_URL="http://localhost:8000"

# Classification Service URL
export CLASSIFICATION_SERVICE_URL="http://localhost:8080"

# Test Configuration
export NUM_REQUESTS=10
export CONCURRENT_REQUESTS=3
```

### Useful Commands

```bash
# Check circuit breaker state
./scripts/monitor_circuit_breaker.sh

# Run accuracy tests
./scripts/run_ml_accuracy_tests.sh

# Run performance tests
./scripts/performance_test_classification.sh

# Check service health
curl http://localhost:8000/health | jq
curl http://localhost:8080/health | jq
```

### Troubleshooting

**Circuit Breaker Stuck Open**:
- Check Python ML service health
- Verify service is accessible
- Check logs for initialization errors
- Use `InitializeWithRetry()` (already implemented)

**Low Accuracy**:
- Verify ML service is being used (check logs)
- Check fallback classifier is working
- Review keyword coverage in database

**Slow Performance**:
- Check website scraping timeout (should be 5s)
- Verify caching is working (check cache hit rate)
- Review database query performance

---

## Next Actions

1. ✅ **Immediate**: Run accuracy tests to verify improvements
2. ✅ **This Week**: Monitor circuit breaker state in production
3. ✅ **This Week**: Run performance tests
4. ⏳ **Next Week**: Analyze results and optimize further
5. ⏳ **Future**: Create observability dashboard

---

## Support

For issues or questions:
- Check logs: `tail -f python_ml_service.log`
- Review documentation: `docs/ml_service_accuracy_implementation_review.md`
- Check health endpoints: `/health` on both services

