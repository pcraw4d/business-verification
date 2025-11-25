# Quick Start: Supabase Integration Tests

## Setup (One-Time)

1. **Get your Supabase credentials** from https://app.supabase.com:
   - Go to your project → Settings → API
   - Copy:
     - **Project URL** → `SUPABASE_URL`
     - **anon public key** → `SUPABASE_ANON_KEY` (or `SUPABASE_API_KEY`)
     - **service_role secret key** → `SUPABASE_SERVICE_ROLE_KEY`

2. **Update `configs/test.env`** with your real credentials:
   ```bash
   # Edit the file
   nano configs/test.env
   
   # Update these lines:
   SUPABASE_URL=https://your-actual-project-ref.supabase.co
   SUPABASE_API_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.your-actual-anon-key
   SUPABASE_SERVICE_ROLE_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.your-actual-service-role-key
   ```

3. **Verify configuration**:
   ```bash
   ./scripts/setup_test_env.sh
   ```
   
   You should see:
   - ✅ Configuration loaded successfully!
   - ✅ No placeholder warnings

## Running Tests

### Run All Tests (Recommended)
```bash
./scripts/run_hybrid_tests.sh
```

This will:
- ✅ Automatically load credentials from `configs/test.env`
- ✅ Run unit tests
- ✅ Run integration tests (if Supabase is configured)
- ✅ Run benchmarks

### Run Specific Tests

```bash
# Unit tests only
go test ./internal/classification/testutil -v -run TestHybridCodeGeneration_Integration

# Integration tests only
go test ./internal/classification/testutil -v -run TestHybridCodeGeneration_WithRealRepository

# All integration tests
go test ./internal/classification/testutil -v -run ".*WithRealRepository"
```

## Troubleshooting

### "No such host" error
- ✅ Your `SUPABASE_URL` is still a placeholder
- ✅ Update `configs/test.env` with your real Supabase project URL

### "Skipping test - Supabase not configured"
- ✅ Credentials not loaded - run `./scripts/setup_test_env.sh` first
- ✅ Or ensure `configs/test.env` exists and has valid credentials

### Connection errors
- ✅ Verify your Supabase project is active
- ✅ Check that your API keys are correct
- ✅ Ensure your network can reach Supabase

## Config File Locations

The setup script checks these files in order:
1. `configs/test.env` ← **Recommended for tests**
2. `configs/development.env`
3. `.env` (project root)
4. `railway.env`
5. `configs/production.env`

## Variable Name Compatibility

The script automatically maps:
- `SUPABASE_API_KEY` → `SUPABASE_ANON_KEY` (for compatibility with `development.env`)

Both variable names work, but `SUPABASE_ANON_KEY` is preferred.

## Example Config File

```bash
# configs/test.env
ENV=test
PORT=8080

# Supabase Configuration
SUPABASE_URL=https://abcdefghijklmnop.supabase.co
SUPABASE_API_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImFiY2RlZmdoaWprbG1ub3AiLCJyb2xlIjoiYW5vbiIsImlhdCI6MTYxNjIzOTAyMiwiZXhwIjoxOTMxODE1MDIyfQ.actual-key-here
SUPABASE_SERVICE_ROLE_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImFiY2RlZmdoaWprbG1ub3AiLCJyb2xlIjoic2VydmljZV9yb2xlIiwiaWF0IjoxNjE2MjM5MDIyLCJleHAiOjE5MzE4MTUwMjJ9.actual-key-here
```

## Next Steps

Once configured:
1. ✅ Run `./scripts/run_hybrid_tests.sh` to verify everything works
2. ✅ Check test output for any failures
3. ✅ Review benchmark results for performance characteristics

