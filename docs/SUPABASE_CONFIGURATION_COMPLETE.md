# Supabase Configuration for Integration Tests - Complete

## ✅ Configuration Complete

The Supabase integration test configuration is now fully set up and ready to use.

## What Was Configured

### 1. Environment Setup Script
**File**: `scripts/setup_test_env.sh`

- ✅ Automatically finds and loads credentials from existing `.env` files
- ✅ Checks multiple config locations (test.env, development.env, .env, etc.)
- ✅ Maps `SUPABASE_API_KEY` to `SUPABASE_ANON_KEY` for compatibility
- ✅ Validates required variables are present
- ✅ Warns about placeholder values
- ✅ Exports variables for use in tests

### 2. Test Runner Script
**File**: `scripts/run_hybrid_tests.sh`

- ✅ Automatically sources credentials using setup script
- ✅ Runs unit tests (always)
- ✅ Runs integration tests (if Supabase configured)
- ✅ Runs benchmarks (if Supabase configured)
- ✅ Provides clear feedback about test status

### 3. Integration Tests
**File**: `internal/classification/testutil/integration_full_test.go`

- ✅ Supports both `SUPABASE_ANON_KEY` and `SUPABASE_API_KEY`
- ✅ Gracefully skips if credentials not configured
- ✅ Tests hybrid code generation with real Supabase
- ✅ Tests multi-industry generation
- ✅ Tests keyword lookup

### 4. Example Configuration
**File**: `configs/test.env.example`

- ✅ Template showing required variables
- ✅ Comments explaining where to get values
- ✅ Safe to commit to version control

## How to Use

### Step 1: Update Config File

Edit `configs/test.env` with your real Supabase credentials:

```bash
nano configs/test.env
```

Update:
- `SUPABASE_URL` - Your Supabase project URL
- `SUPABASE_API_KEY` - Your anon/public key
- `SUPABASE_SERVICE_ROLE_KEY` - Your service role key (optional)

### Step 2: Verify Configuration

```bash
./scripts/setup_test_env.sh
```

You should see:
- ✅ Configuration loaded successfully
- ✅ No placeholder warnings (if using real credentials)

### Step 3: Run Tests

```bash
./scripts/run_hybrid_tests.sh
```

Or run specific tests:
```bash
# Integration tests
go test ./internal/classification/testutil -v -run TestHybridCodeGeneration_WithRealRepository

# All integration tests
go test ./internal/classification/testutil -v -run ".*WithRealRepository"
```

## Config File Priority

The setup script checks these files in order:
1. `configs/test.env` ← **Best for tests**
2. `configs/development.env`
3. `.env` (project root)
4. `railway.env`
5. `configs/production.env`

## Variable Compatibility

The system supports both variable naming conventions:
- `SUPABASE_ANON_KEY` (preferred)
- `SUPABASE_API_KEY` (automatically mapped to SUPABASE_ANON_KEY)

This ensures compatibility with existing `development.env` files.

## Security Notes

- ✅ `configs/test.env` should be in `.gitignore` (already configured)
- ✅ Never commit real credentials to version control
- ✅ Use `configs/test.env.example` as a template
- ✅ Consider using a separate Supabase project for testing

## Current Status

- ✅ Scripts created and executable
- ✅ Integration tests updated to use credentials
- ✅ Example config file created
- ⚠️ **Action Required**: Update `configs/test.env` with real Supabase credentials

## Next Steps

1. **Get Supabase credentials** from https://app.supabase.com
2. **Update `configs/test.env`** with real values
3. **Run `./scripts/setup_test_env.sh`** to verify
4. **Run `./scripts/run_hybrid_tests.sh`** to execute tests

## Documentation

- **Quick Start**: `docs/QUICK_START_SUPABASE_TESTS.md`
- **Full Guide**: `docs/SUPABASE_TEST_CONFIGURATION.md` (if it exists)
- **Test Summary**: `docs/TESTING_COMPLETE_SUMMARY.md`

## Troubleshooting

If tests fail with "no such host":
- ✅ Your `SUPABASE_URL` is still a placeholder
- ✅ Update `configs/test.env` with your real Supabase project URL

If tests are skipped:
- ✅ Credentials not loaded - ensure `configs/test.env` exists
- ✅ Run `./scripts/setup_test_env.sh` to verify configuration

