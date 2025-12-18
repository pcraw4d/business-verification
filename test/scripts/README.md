# Test Scripts

This directory contains scripts for running comprehensive classification tests.

## Scripts

### `run_comprehensive_tests.sh`
Runs tests against localhost (default) or custom URL.

**Usage:**
```bash
# Default (localhost:8081)
./test/scripts/run_comprehensive_tests.sh

# Custom URL
CLASSIFICATION_API_URL=http://localhost:8081 ./test/scripts/run_comprehensive_tests.sh
```

### `run_comprehensive_tests_railway.sh`
Runs tests against Railway production environment.

**Usage:**
```bash
# Direct Classification Service endpoint
./test/scripts/run_comprehensive_tests_railway.sh

# Via API Gateway
USE_API_GATEWAY=true ./test/scripts/run_comprehensive_tests_railway.sh
```

**Features:**
- Automatically uses Railway production URLs
- Verifies service health before running
- Warns about production testing
- Extended timeout (60 minutes)
- Production-specific error handling

## Environment Variables

- `CLASSIFICATION_API_URL` - Override the API URL (defaults vary by script)
- `USE_API_GATEWAY` - Use API Gateway endpoint instead of direct service (Railway script only)

## Railway Production URLs

- **Classification Service**: `https://classification-service-production.up.railway.app`
- **API Gateway**: `https://api-gateway-service-production-21fd.up.railway.app`

## Output

Both scripts generate:
- `test/results/comprehensive_test_results.json` - Detailed JSON report
- `test/results/test_output_*.txt` - Full test output log

