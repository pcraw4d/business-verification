# Restoration Functionality Test Suite

This directory contains test scripts to verify all restored functionality.

## Test Scripts

### 1. `restoration_tests.sh` - Comprehensive Test Suite
Tests all restored endpoints and functionality.

**Usage:**
```bash
# Default (localhost:8080)
./test/restoration_tests.sh

# Custom base URL
BASE_URL=http://localhost:3000 ./test/restoration_tests.sh
```

**Tests:**
- Threshold CRUD operations
- Export/Import round-trip
- Risk factors and categories
- Recommendation rules CRUD
- Notification channels
- System monitoring endpoints
- Request ID extraction

### 2. `test_database_persistence.sh` - Database Persistence Test
Tests that thresholds persist in the database across server restarts.

**Usage:**
```bash
./test/test_database_persistence.sh
```

**What it does:**
1. Creates a threshold
2. Verifies it exists
3. Provides instructions for manual server restart test

### 3. `verify_persistence.sh` - Verify After Restart
Verifies that thresholds created before restart still exist.

**Usage:**
```bash
# After restarting the server
./test/verify_persistence.sh
```

### 4. `test_graceful_degradation.sh` - Graceful Degradation Test
Tests that the system works when database/Redis are unavailable.

**Usage:**
```bash
./test/test_graceful_degradation.sh
```

**What it tests:**
- Health check reports service status
- Threshold endpoints work with in-memory fallback
- Classification works without Redis caching

## Prerequisites

1. **Server Running**: The server must be running on the target URL
2. **jq**: Required for JSON parsing (install with `brew install jq` or `apt-get install jq`)
3. **curl**: Required for HTTP requests (usually pre-installed)

## Environment Variables

- `BASE_URL`: Base URL of the API server (default: `http://localhost:8080`)
- `TEST_OUTPUT_DIR`: Directory for test output files (default: `./test_output`)

## Test Output

Test results and data are saved to `test_output/` directory:
- `threshold_id.txt`: Created threshold IDs
- `thresholds_export.json`: Exported thresholds
- `persistence_threshold_id.txt`: Threshold ID for persistence testing

## Running All Tests

```bash
# Make scripts executable
chmod +x test/*.sh

# Run comprehensive test suite
./test/restoration_tests.sh

# Test database persistence
./test/test_database_persistence.sh

# Test graceful degradation
./test/test_graceful_degradation.sh
```

## Expected Results

### With Database Configured
- ✅ All CRUD operations work
- ✅ Thresholds persist across restarts
- ✅ Export/Import works correctly

### Without Database (In-Memory Fallback)
- ✅ GET endpoints return empty or in-memory data
- ✅ CREATE/UPDATE/DELETE work in-memory
- ✅ No persistence across restarts (expected)

### With Redis Configured
- ✅ Classification endpoint uses caching
- ✅ X-Cache header present in responses

### Without Redis
- ✅ Classification endpoint works without caching
- ✅ No X-Cache header (expected)

## Troubleshooting

### Tests Fail with Connection Refused
- Ensure the server is running: `go run cmd/railway-server/main.go`
- Check the BASE_URL matches your server port

### Tests Fail with jq Errors
- Install jq: `brew install jq` (macOS) or `apt-get install jq` (Linux)
- Some tests will still work without jq, but output formatting will be limited

### Database Persistence Test Fails
- Verify DATABASE_URL is set correctly
- Check database connection in health endpoint: `curl http://localhost:8080/health/detailed`
- Ensure database schema includes `risk_thresholds` table

### Threshold Not Found After Restart
- This is expected if DATABASE_URL is not set (in-memory mode)
- Check health endpoint to verify database status
- Verify database connection logs on server startup

## Manual Testing Checklist

For complete verification, also perform these manual tests:

- [ ] Create threshold → Restart server → Verify threshold still exists
- [ ] Export thresholds → Modify JSON → Import → Verify changes
- [ ] Test all notification channel types (email, SMS, Slack, webhook, etc.)
- [ ] Test error cases (invalid JSON, missing fields, non-existent IDs)
- [ ] Test with X-Request-ID header and verify it's returned
- [ ] Test system health endpoint with/without database
- [ ] Test system metrics endpoint
- [ ] Test cleanup endpoint with different data types
