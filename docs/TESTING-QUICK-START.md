# Quick Start - Testing & Debugging

## Run All Tests

```bash
./scripts/run-all-tests.sh
```

This single command will:
1. ✅ Test all API endpoints
2. ✅ Detect placeholder/mock data
3. ✅ Find unused features

## Individual Tests

### Test API Endpoints
```bash
./scripts/comprehensive-api-test.sh
```

### Find Placeholder Data
```bash
./scripts/detect-placeholder-data.sh
```

### Find Unused Features
```bash
node scripts/analyze-unused-features.js
```

## What Gets Tested

### API Testing
- All health endpoints
- Merchant CRUD operations
- Risk assessment endpoints
- Business Intelligence endpoints
- Classification endpoints
- Validates responses are not placeholder data
- Checks for valid JSON
- Verifies HTTP status codes

### Placeholder Detection
Scans for:
- "Sample Merchant"
- "Mock" / "mock"
- "TODO.*return"
- "placeholder"
- "test-"
- "dummy"
- "fake"
- "example"
- "For now"
- "temporary"
- "fallback"

### Unused Features
Identifies:
- Backend API endpoints not called from frontend
- Frontend components not instantiated
- Data services not used
- Utility functions not referenced

## Test Results

All results saved to: `test-results/`

- `api-test-report-{timestamp}.json` - API test details
- `unused-features-analysis.json` - Unused features list

## Configuration

Set environment variables before running:

```bash
export API_BASE_URL="https://api-gateway-service-production-21fd.up.railway.app"
export TEST_MERCHANT_ID="your-merchant-id"
./scripts/comprehensive-api-test.sh
```

## Next Steps

1. **Review test results** in `test-results/` directory
2. **Fix any placeholder data** found by detection script
3. **Implement UI for unused endpoints** identified by analysis
4. **Run tests regularly** before deploying

For detailed documentation, see: `docs/TESTING-AND-DEBUGGING-GUIDE.md`

