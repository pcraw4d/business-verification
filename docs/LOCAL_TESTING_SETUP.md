# Local Testing Environment Setup Guide

This guide will help you set up a local testing environment for running merchant-details tests.

## Quick Start

Run the automated setup script:

```bash
./scripts/setup-local-testing.sh
```

This script will:
1. Check prerequisites (Go, Docker, Docker Compose)
2. Create `.env.test` configuration file
3. Start PostgreSQL and Redis in Docker
4. Run database migrations
5. Verify the setup

## Manual Setup

If you prefer to set up manually or the script doesn't work for your environment:

### 1. Prerequisites

Ensure you have installed:
- **Go** 1.22 or later
- **Docker** and **Docker Compose**
- **PostgreSQL client tools** (optional, for manual database access)

### 2. Start Database Services

#### Option A: Using Docker Compose (Recommended)

```bash
# Start PostgreSQL and Redis
docker-compose -f docker-compose.test.yml up -d

# Verify services are running
docker-compose -f docker-compose.test.yml ps
```

#### Option B: Using Local Services

If you have PostgreSQL and Redis installed locally:

```bash
# Start PostgreSQL (macOS with Homebrew)
brew services start postgresql

# Start Redis (macOS with Homebrew)
brew services start redis

# Or on Linux
sudo systemctl start postgresql
sudo systemctl start redis
```

### 3. Create Test Environment File

Create `.env.test` in the project root:

```bash
# Test Configuration
TEST_BASE_URL=http://localhost:8080
TEST_AUTH_TOKEN=test-token-local

# Database Configuration
DATABASE_URL=postgres://kyb_test:kyb_test_password@localhost:5432/kyb_test?sslmode=disable

# Redis Configuration
REDIS_URL=redis://localhost:6379/0

# Application Configuration
ENVIRONMENT=local
LOG_LEVEL=debug
PORT=8080
```

### 4. Run Database Migrations

```bash
# Load environment variables
source .env.test

# Or export manually
export DATABASE_URL="postgres://kyb_test:kyb_test_password@localhost:5432/kyb_test?sslmode=disable"

# Run migrations
psql $DATABASE_URL -f internal/database/migrations/001_initial_schema.sql
psql $DATABASE_URL -f internal/database/migrations/002_rbac_schema.sql
psql $DATABASE_URL -f internal/database/migrations/003_performance_indexes.sql
psql $DATABASE_URL -f internal/database/migrations/004_enhanced_classification.sql
psql $DATABASE_URL -f internal/database/migrations/005_merchant_portfolio_schema.sql
psql $DATABASE_URL -f internal/database/migrations/007_foreign_key_relationships.sql
psql $DATABASE_URL -f internal/database/migrations/008_additional_performance_indexes.sql
psql $DATABASE_URL -f internal/database/migrations/008_enhance_merchants_table.sql
psql $DATABASE_URL -f internal/database/migrations/008_unified_compliance_schema.sql
psql $DATABASE_URL -f internal/database/migrations/008_user_table_consolidation.sql
psql $DATABASE_URL -f internal/database/migrations/009_remove_redundant_profiles_table.sql
psql $DATABASE_URL -f internal/database/migrations/009_unified_audit_schema.sql
psql $DATABASE_URL -f internal/database/migrations/010_add_async_risk_assessment_columns.sql
psql $DATABASE_URL -f internal/database/migrations/011_add_updated_at_to_risk_assessments.sql
```

**Note**: Some migrations may show "already exists" errors if run multiple times. This is normal.

### 5. Verify Setup

```bash
# Test database connection
psql $DATABASE_URL -c "SELECT version();"

# Test Redis connection
redis-cli ping
# Should return: PONG
```

## Running Tests

### Load Environment Variables

```bash
# Option 1: Source the .env.test file
source .env.test

# Option 2: Export manually
export TEST_BASE_URL=http://localhost:8080
export TEST_AUTH_TOKEN=test-token-local
export DATABASE_URL=postgres://kyb_test:kyb_test_password@localhost:5432/kyb_test?sslmode=disable
```

### Run Integration Tests

```bash
# Run all integration tests
go test -v -tags=integration ./test/integration/...

# Run specific risk assessment integration tests
go test -v -tags=integration ./test/integration/risk_assessment_integration_test.go

# Run with timeout
go test -v -tags=integration -timeout=30m ./test/integration/risk_assessment_integration_test.go
```

### Run E2E Tests

```bash
# Run all E2E tests
go test -v -tags=e2e ./test/e2e/...

# Run merchant details E2E tests
go test -v -tags=e2e ./test/e2e/merchant_details_e2e_test.go

# Run merchant analytics API tests
go test -v -tags=e2e ./test/e2e/merchant_analytics_api_test.go

# Run with timeout
go test -v -tags=e2e -timeout=30m ./test/e2e/merchant_details_e2e_test.go
```

### Run Tests with JSON Output (for CI/CD)

```bash
# Generate JSON output for test reporting
go test -v -tags=integration -json ./test/integration/risk_assessment_integration_test.go | tee test-results.json
```

## Testing Against Running Services

If you want to test against a running API server:

### 1. Start the API Server

```bash
# Load environment variables
source .env.test

# Start the server (adjust path based on your project structure)
go run cmd/server/main.go
# Or
go run cmd/railway-server/main.go
```

### 2. Update TEST_BASE_URL

Make sure `TEST_BASE_URL` points to your running server:

```bash
export TEST_BASE_URL=http://localhost:8080
```

### 3. Run Tests

Tests will now hit your running server instead of using mocks.

## Troubleshooting

### Database Connection Issues

```bash
# Check if PostgreSQL is running
docker ps | grep postgres
# Or
ps aux | grep postgres

# Test connection manually
psql $DATABASE_URL -c "SELECT 1;"

# Check database exists
psql $DATABASE_URL -c "\l" | grep kyb_test
```

### Redis Connection Issues

```bash
# Check if Redis is running
docker ps | grep redis
# Or
redis-cli ping

# Test connection manually
docker exec kyb-test-redis redis-cli ping
```

### Migration Errors

If migrations fail with "already exists" errors, this is usually fine - it means the migration was already applied. You can:

1. **Check what's already applied:**
   ```bash
   psql $DATABASE_URL -c "\dt"  # List tables
   ```

2. **Reset database (WARNING: Deletes all data):**
   ```bash
   docker-compose -f docker-compose.test.yml down -v
   docker-compose -f docker-compose.test.yml up -d
   # Then run migrations again
   ```

### Port Conflicts

If ports 5432 (PostgreSQL) or 6379 (Redis) are already in use:

1. **Stop existing services:**
   ```bash
   # Stop local PostgreSQL
   brew services stop postgresql
   
   # Stop local Redis
   brew services stop redis
   ```

2. **Or change ports in docker-compose.test.yml:**
   ```yaml
   ports:
     - "5433:5432"  # Use 5433 instead of 5432
   ```
   Then update `DATABASE_URL` accordingly.

### Test Failures

1. **Check environment variables are set:**
   ```bash
   echo $TEST_BASE_URL
   echo $TEST_AUTH_TOKEN
   echo $DATABASE_URL
   ```

2. **Run tests with verbose output:**
   ```bash
   go test -v -tags=integration ./test/integration/risk_assessment_integration_test.go
   ```

3. **Check test logs for specific errors**

## Cleaning Up

### Stop Docker Services

```bash
# Stop services but keep data
docker-compose -f docker-compose.test.yml down

# Stop services and remove all data
docker-compose -f docker-compose.test.yml down -v
```

### Remove Test Environment File

```bash
rm .env.test
```

## Next Steps

After setting up the local testing environment:

1. ✅ Run integration tests to verify database operations
2. ✅ Run E2E tests to verify API endpoints
3. ✅ Review test results and fix any issues
4. ✅ Commit your changes

For more information, see:
- [Week 1 Merchant Details Tasks](../docs/WEEK1_MERCHANT_DETAILS_REMAINING_TASKS.md)
- [API Testing Guide](../tests/api/merchant-details/README.md)

