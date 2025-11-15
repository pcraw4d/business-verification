# Quick Start Testing Guide

## Prerequisites

1. **Start the server:**
   ```bash
   go run cmd/railway-server/main.go
   ```

2. **Install jq (for JSON parsing):**
   ```bash
   # macOS
   brew install jq
   
   # Linux
   sudo apt-get install jq
   ```

## Quick Test Run

### 1. Run All Tests (Recommended)
```bash
./test/restoration_tests.sh
```

This comprehensive test suite covers:
- ✅ Threshold CRUD operations
- ✅ Export/Import functionality
- ✅ Risk factors and categories
- ✅ Recommendation rules
- ✅ Notification channels
- ✅ System monitoring
- ✅ Request ID handling

### 2. Test Database Persistence
```bash
# Step 1: Create a threshold
./test/test_database_persistence.sh

# Step 2: Restart your server (Ctrl+C, then restart)

# Step 3: Verify persistence
./test/verify_persistence.sh
```

### 3. Test Graceful Degradation
```bash
# Test with/without database and Redis
./test/test_graceful_degradation.sh
```

## Expected Output

### Successful Test Run
```
==========================================
Restoration Functionality Test Suite
==========================================
Base URL: http://localhost:8080

Testing: GET all thresholds
  HTTP Status: 200
  ✓ PASSED

Testing: CREATE threshold
  HTTP Status: 201
  ✓ PASSED

...

==========================================
Test Summary
==========================================
Tests Passed: 23
Tests Failed: 0
Total Tests: 23

✓ All tests passed!
```

## Troubleshooting

### Server Not Running
```bash
# Check if server is running
curl http://localhost:8080/health

# If not, start it
go run cmd/railway-server/main.go
```

### Port Already in Use
```bash
# Use a different port
PORT=3000 go run cmd/railway-server/main.go

# Update test script
BASE_URL=http://localhost:3000 ./test/restoration_tests.sh
```

### Database Connection Issues
```bash
# Check health endpoint
curl http://localhost:8080/health/detailed | jq '.checks.postgres'

# Verify DATABASE_URL is set
echo $DATABASE_URL
```

## Next Steps

After running automated tests, perform these manual tests:

1. **Database Persistence:**
   - Create threshold → Restart server → Verify it still exists

2. **Export/Import Round-trip:**
   - Export → Modify JSON → Import → Verify changes

3. **Error Handling:**
   - Test invalid JSON, missing fields, non-existent IDs

4. **Performance:**
   - Test with multiple concurrent requests
   - Verify connection pooling works

