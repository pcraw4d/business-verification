# Integration Test Setup Guide

This guide explains how to set up and run integration tests for the merchant service.

## Prerequisites

- Go 1.24 or later
- Supabase account (for test database)
- Environment variables configured (see below)

## Test Database Setup

Integration tests require a Supabase database connection. You have two options:

### Option 1: Use Existing Supabase Project (Recommended)

If you already have a Supabase project, you can use it for testing:

1. **Set Environment Variables:**
   ```bash
   export SUPABASE_URL="https://your-project-id.supabase.co"
   export SUPABASE_SERVICE_ROLE_KEY="your-service-role-key"
   ```

2. **Verify Connection:**
   ```bash
   cd services/merchant-service
   go test ./test/integration -v -run TestDatabaseConnection
   ```

### Option 2: Create Dedicated Test Project

For isolated testing, create a separate Supabase project:

1. **Create Test Project:**
   - Go to [Supabase Dashboard](https://app.supabase.com)
   - Create a new project specifically for testing
   - Note the project URL and service role key

2. **Set Environment Variables:**
   ```bash
   export SUPABASE_URL="https://your-test-project-id.supabase.co"
   export SUPABASE_SERVICE_ROLE_KEY="your-test-service-role-key"
   ```

3. **Run Migrations:**
   ```bash
   # Apply all migrations to test database
   # Use your migration tool or Supabase CLI
   ```

## Running Integration Tests

### Run All Integration Tests

```bash
cd services/merchant-service
go test ./internal/handlers/... -v -run Integration
```

### Run Specific Test

```bash
go test ./internal/handlers/... -v -run TestMerchantCreationTriggersClassificationJob
```

### Skip Integration Tests (Unit Tests Only)

```bash
go test ./internal/handlers/... -short
```

Integration tests will automatically skip if:
- `-short` flag is used
- `SUPABASE_URL` or `SUPABASE_SERVICE_ROLE_KEY` are not set
- Database connection fails

## Test Helpers

The test suite includes helper functions in `test/integration/test_helpers.go`:

- `SetupTestDatabase(t *testing.T)` - Creates test database connection
- `TeardownTestDatabase()` - Closes database connection
- `CleanupTestData(t *testing.T, merchantIDs []string)` - Removes test data
- `CreateTestMerchant(t *testing.T, data map[string]interface{})` - Creates test merchant
- `GetTestMerchant(t *testing.T, merchantID string)` - Retrieves test merchant

## Test Data Cleanup

Tests automatically clean up created data using `defer` statements:

```go
merchantID, err := testDB.CreateTestMerchant(t, merchantData)
require.NoError(t, err)
defer testDB.CleanupTestData(t, []string{merchantID})
```

## Environment Variables

Create a `.env.test` file in the project root (optional):

```env
SUPABASE_URL=https://your-project-id.supabase.co
SUPABASE_SERVICE_ROLE_KEY=your-service-role-key
```

Load it before running tests:

```bash
export $(cat .env.test | xargs)
go test ./internal/handlers/... -v
```

## Troubleshooting

### Tests Skip Automatically

If tests are skipping, check:
1. Environment variables are set correctly
2. Supabase project is accessible
3. Service role key has proper permissions
4. Database migrations are applied

### Connection Timeout

If you see connection timeouts:
1. Verify Supabase URL is correct
2. Check network connectivity
3. Ensure service role key is valid
4. Check Supabase project status

### Test Data Not Cleaning Up

If test data persists:
1. Check that `defer testDB.CleanupTestData()` is called
2. Verify service role key has DELETE permissions
3. Manually clean up test data if needed

## Best Practices

1. **Always use test database helpers** - Don't create database connections manually
2. **Clean up test data** - Always use `defer` for cleanup
3. **Use unique test data** - Avoid conflicts with other tests
4. **Skip in short mode** - Use `testing.Short()` check
5. **Handle errors gracefully** - Tests should skip, not fail, if database unavailable

## CI/CD Integration

For CI/CD pipelines, set environment variables as secrets:

```yaml
env:
  SUPABASE_URL: ${{ secrets.SUPABASE_TEST_URL }}
  SUPABASE_SERVICE_ROLE_KEY: ${{ secrets.SUPABASE_TEST_SERVICE_ROLE_KEY }}
```

## Example Test

```go
func TestExampleIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // Setup
    testDB := integration.SetupTestDatabase(t)
    defer testDB.TeardownTestDatabase()

    // Create test data
    merchantID, err := testDB.CreateTestMerchant(t, map[string]interface{}{
        "name": "Test Merchant",
        "industry": "Technology",
    })
    require.NoError(t, err)
    defer testDB.CleanupTestData(t, []string{merchantID})

    // Run test
    // ...

    // Cleanup happens automatically via defer
}
```

