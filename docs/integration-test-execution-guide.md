# Integration Test Execution Guide

## Running Weeks 2-4 Integration Tests

The integration tests for Weeks 2-4 features require a database connection. Follow these steps to run them:

### Prerequisites

1. **Database Setup**: Ensure you have access to a test database (Supabase or local PostgreSQL)
2. **Environment Variables**: Set the required database connection variables

### Environment Variables

You need to set one of the following:

**Option 1: Direct Database URL**
```bash
export TEST_DATABASE_URL="postgres://user:password@host:5432/database?sslmode=require"
```

**Option 2: Supabase Connection**
```bash
export SUPABASE_URL="https://your-project.supabase.co"
export SUPABASE_SERVICE_ROLE_KEY="your-service-role-key"
```

The test will automatically construct the connection string from Supabase variables if `TEST_DATABASE_URL` is not set.

### Running the Tests

#### Method 1: Using the Test Runner Script ✅ RECOMMENDED

```bash
# From project root
export SUPABASE_URL="https://your-project.supabase.co"
export SUPABASE_SERVICE_ROLE_KEY="your-service-role-key"

bash test/integration/run_weeks_24_tests.sh
```

#### Method 2: Direct Command Execution

From the project root:

```bash
# Set environment variables first
export SUPABASE_URL="https://your-project.supabase.co"
export SUPABASE_SERVICE_ROLE_KEY="your-service-role-key"

# Run the tests using specific file paths (required due to go.work configuration)
go test -tags=integration -v -run TestWeeks24Integration \
  ./test/integration/weeks_2_4_integration_test.go \
  ./test/integration/database_setup.go
```

**Important:** We use specific file paths instead of `./test/integration/...` because of the `go.work` workspace configuration. See `docs/go-work-configuration-analysis.md` for details.

### Test Coverage

The `TestWeeks24Integration` test covers:

1. **GetMerchantAnalytics** - Tests analytics endpoint
2. **GetWebsiteAnalysis** - Tests website analysis endpoint
3. **GetRiskHistory** - Tests risk history endpoint
4. **GetRiskPredictions** - Tests risk predictions endpoint
5. **GetRiskIndicators** - Tests risk indicators endpoint
6. **GetEnrichmentSources** - Tests enrichment sources endpoint

### Expected Behavior

- Tests will skip if database is not available (with clear error message)
- Tests will skip if run with `-short` flag
- Tests verify that endpoints exist and handle requests correctly
- Some tests may return errors for non-existent merchants (expected behavior)

### Troubleshooting

**Issue: "database not available" or "password authentication failed"**
- ✅ **This is expected if credentials aren't set**
- Check that environment variables are set correctly
- Verify database connection string format
- Ensure database is accessible from your network
- For Supabase: Verify the service role key is correct

**Issue: "package not found"**
- Ensure you're running from the project root
- Use specific file paths as shown in Method 2
- Don't use `./test/integration/...` pattern

**Issue: Tests skip automatically**
- Check if `-short` flag is being used
- Verify database connection is working
- Check test output for skip reasons (should show database connection error)

### Test Database Requirements

The tests require these tables to exist:
- `merchants`
- `risk_assessments`
- `merchant_analytics` (newly created)
- `risk_indicators` (newly created)
- `enrichment_jobs` (newly created)
- `enrichment_sources` (newly created)

All tables should have been created by migration `011_create_test_tables.sql`.

### Go Workspace Configuration

This project uses `go.work` for multi-module workspace management. When running tests:
- ✅ Use specific file paths: `./test/integration/weeks_2_4_integration_test.go`
- ❌ Don't use patterns: `./test/integration/...` (causes workspace resolution issues)

See `docs/go-work-configuration-analysis.md` for detailed explanation.

### Example Output

**Successful test run (with database):**
```
=== RUN   TestWeeks24Integration
=== RUN   TestWeeks24Integration/GetMerchantAnalytics
=== RUN   TestWeeks24Integration/GetWebsiteAnalysis
...
PASS
ok      command-line-arguments    2.345s
```

**Test skipped (no database credentials):**
```
=== RUN   TestWeeks24Integration
    weeks_2_4_integration_test.go:31: Skipping integration test - database not available: ...
--- SKIP: TestWeeks24Integration (0.08s)
PASS
ok      command-line-arguments    0.587s
```

### Next Steps

After running the tests:
1. Review test output for any failures
2. Check database logs if tests fail
3. Verify table structures match expectations
4. Update tests if schema changes are needed
