# Test Database Setup Guide

This guide explains how to set up a test database for running integration tests in the KYB Platform.

## Overview

Integration tests require a PostgreSQL database to test database interactions, repositories, and services. The test framework supports multiple database setup options:

1. **Supabase Test Project** (Recommended for cloud-based testing)
2. **Local Docker PostgreSQL** (Recommended for local development)
3. **Local PostgreSQL Installation** (Alternative for local development)

## Option 1: Supabase Test Project (Recommended)

### Prerequisites

- Supabase account
- Supabase project created

### Setup Steps

1. **Create a Supabase Test Project:**
   - Go to [Supabase Dashboard](https://app.supabase.com)
   - Create a new project specifically for testing
   - Note the project URL and service role key

2. **Set Environment Variables:**
   ```bash
   export SUPABASE_URL="https://your-project-id.supabase.co"
   export SUPABASE_SERVICE_ROLE_KEY="your-service-role-key"
   export TEST_DATABASE_URL="postgres://postgres:[PASSWORD]@db.[PROJECT-REF].supabase.co:5432/postgres?sslmode=require"
   ```

   Or create a `.env.test` file:
   ```env
   SUPABASE_URL=https://your-project-id.supabase.co
   SUPABASE_SERVICE_ROLE_KEY=your-service-role-key
   TEST_DATABASE_URL=postgres://postgres:[PASSWORD]@db.[PROJECT-REF].supabase.co:5432/postgres?sslmode=require
   ```

3. **Run Database Migrations:**
   ```bash
   # Apply migrations to test database
   # (Use your migration tool of choice)
   ```

4. **Verify Connection:**
   ```bash
   go test ./test/integration -v -run TestVerifyDatabase
   ```

## Option 2: Local Docker PostgreSQL

### Prerequisites

- Docker installed
- Docker Compose (optional)

### Setup Steps

1. **Create `docker-compose.test.yml`:**
   ```yaml
   version: '3.8'
   services:
     test-db:
       image: postgres:15-alpine
       container_name: kyb-test-db
       environment:
         POSTGRES_USER: postgres
         POSTGRES_PASSWORD: password
         POSTGRES_DB: kyb_test
       ports:
         - "5433:5432"
       volumes:
         - test-db-data:/var/lib/postgresql/data
       healthcheck:
         test: ["CMD-SHELL", "pg_isready -U postgres"]
         interval: 5s
         timeout: 5s
         retries: 5

   volumes:
     test-db-data:
   ```

2. **Start the Test Database:**
   ```bash
   docker-compose -f docker-compose.test.yml up -d
   ```

3. **Set Environment Variable:**
   ```bash
   export TEST_DATABASE_URL="postgres://postgres:password@localhost:5433/kyb_test?sslmode=disable"
   ```

4. **Run Database Migrations:**
   ```bash
   # Apply migrations to test database
   ```

5. **Verify Connection:**
   ```bash
   go test ./test/integration -v -run TestVerifyDatabase
   ```

6. **Stop the Test Database (when done):**
   ```bash
   docker-compose -f docker-compose.test.yml down
   ```

## Option 3: Local PostgreSQL Installation

### Prerequisites

- PostgreSQL 12+ installed locally

### Setup Steps

1. **Create Test Database:**
   ```bash
   createdb kyb_test
   ```

2. **Set Environment Variable:**
   ```bash
   export TEST_DATABASE_URL="postgres://postgres:password@localhost:5432/kyb_test?sslmode=disable"
   ```

3. **Run Database Migrations:**
   ```bash
   # Apply migrations to test database
   ```

4. **Verify Connection:**
   ```bash
   go test ./test/integration -v -run TestVerifyDatabase
   ```

## Environment Variables

The test framework looks for the following environment variables in order:

1. `TEST_DATABASE_URL` - Direct database connection string (highest priority)
2. `SUPABASE_URL` + `SUPABASE_SERVICE_ROLE_KEY` - Supabase project credentials
3. Default: `postgres://postgres:password@localhost:5432/kyb_test?sslmode=disable`

## Running Tests

### Run All Integration Tests

```bash
go test ./test/integration -v
```

### Run Specific Test

```bash
go test ./test/integration -v -run TestName
```

### Skip Database Tests

If you want to skip database-dependent tests:

```bash
export SKIP_DATABASE_TESTS=true
go test ./test/integration -v
```

### Run Tests in Short Mode

```bash
go test ./test/integration -short
```

## Test Database Helpers

The test framework provides helper functions in `test/integration/database_setup.go`:

- `SetupTestDatabase()` - Initialize test database connection
- `CleanupTestDatabase()` - Close database connection
- `ResetTestDatabase(ctx)` - Truncate all test tables
- `SeedTestData(ctx)` - Insert test data fixtures
- `VerifyTestDatabase()` - Check if database is accessible

## Example Usage

```go
func TestExample(t *testing.T) {
    // Setup test database
    testDB, err := SetupTestDatabase()
    if err != nil {
        t.Fatalf("Failed to setup test database: %v", err)
    }
    defer testDB.CleanupTestDatabase()

    // Reset database before test
    ctx := context.Background()
    if err := testDB.ResetTestDatabase(ctx); err != nil {
        t.Fatalf("Failed to reset test database: %v", err)
    }

    // Seed test data
    if err := testDB.SeedTestData(ctx); err != nil {
        t.Fatalf("Failed to seed test data: %v", err)
    }

    // Run your test...
    db := testDB.GetDB()
    // Use db for testing
}
```

## Troubleshooting

### Connection Refused

- Check if PostgreSQL is running
- Verify port number (default: 5432, Docker: 5433)
- Check firewall settings

### Authentication Failed

- Verify username and password
- Check PostgreSQL authentication configuration
- For Supabase, ensure service role key is correct

### Database Does Not Exist

- Create the database: `createdb kyb_test`
- Verify database name in connection string

### SSL Mode Errors

- For local testing, use `sslmode=disable`
- For Supabase, use `sslmode=require`

## Best Practices

1. **Isolate Test Data:** Always reset the database before each test run
2. **Use Transactions:** Wrap tests in transactions when possible for faster cleanup
3. **Clean Up:** Always clean up test data after tests complete
4. **Separate Test Database:** Never use production database for testing
5. **Environment Variables:** Use `.env.test` file for local development
6. **CI/CD:** Configure test database in CI/CD pipeline

## CI/CD Configuration

For CI/CD pipelines, configure the test database as a service:

```yaml
# Example GitHub Actions
services:
  postgres:
    image: postgres:15-alpine
    env:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
      POSTGRES_DB: kyb_test
    options: >-
      --health-cmd pg_isready
      --health-interval 10s
      --health-timeout 5s
      --health-retries 5
    ports:
      - 5432:5432

env:
  TEST_DATABASE_URL: postgres://postgres:password@localhost:5432/kyb_test?sslmode=disable
```

## Additional Resources

- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Supabase Documentation](https://supabase.com/docs)
- [Docker PostgreSQL Image](https://hub.docker.com/_/postgres)

