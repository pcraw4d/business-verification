#!/bin/bash
# Run Weeks 2-4 Integration Tests
# This script sets up the environment and runs the integration tests

set -e

echo "=========================================="
echo "Weeks 2-4 Integration Test Runner"
echo "=========================================="

# Get the project root directory
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$PROJECT_ROOT"

echo "Project root: $PROJECT_ROOT"
echo ""

# Check for required environment variables
if [ -z "$TEST_DATABASE_URL" ] && [ -z "$SUPABASE_URL" ]; then
    echo "⚠️  Warning: Neither TEST_DATABASE_URL nor SUPABASE_URL is set"
    echo "   Tests will attempt to use default local PostgreSQL connection"
    echo ""
fi

# Load Railway credentials first (highest priority)
if [ -f ".env.railway.full" ]; then
    echo "Loading Railway credentials from .env.railway.full..."
    set -a
    source .env.railway.full
    set +a
    # Store Railway values before other configs overwrite them
    RAILWAY_SUPABASE_URL="${SUPABASE_URL:-}"
    RAILWAY_SUPABASE_KEY="${SUPABASE_SERVICE_ROLE_KEY:-}"
    RAILWAY_DATABASE_URL="${DATABASE_URL:-}"
    RAILWAY_TEST_DB_URL="${TEST_DATABASE_URL:-}"
fi

# Load test environment if it exists (lower priority)
if [ -f "configs/test.env" ]; then
    echo "Loading test environment from configs/test.env..."
    set -a
    source configs/test.env
    set +a
fi

# Export environment variables if they exist in configs (lowest priority)
if [ -f "configs/development.env" ]; then
    echo "Loading development environment variables..."
    set -a
    source configs/development.env
    set +a
fi

# Restore Railway credentials if they were loaded (they take precedence)
if [ -n "$RAILWAY_SUPABASE_URL" ]; then
    export SUPABASE_URL="$RAILWAY_SUPABASE_URL"
fi
if [ -n "$RAILWAY_SUPABASE_KEY" ]; then
    export SUPABASE_SERVICE_ROLE_KEY="$RAILWAY_SUPABASE_KEY"
fi
if [ -n "$RAILWAY_DATABASE_URL" ]; then
    export DATABASE_URL="$RAILWAY_DATABASE_URL"
fi
if [ -n "$RAILWAY_TEST_DB_URL" ]; then
    export TEST_DATABASE_URL="$RAILWAY_TEST_DB_URL"
fi

echo ""
echo "Environment check:"
echo "  TEST_DATABASE_URL: ${TEST_DATABASE_URL:-not set}"
echo "  DATABASE_URL: ${DATABASE_URL:+set (hidden)}"
echo "  SUPABASE_URL: ${SUPABASE_URL:-not set}"
echo "  SUPABASE_SERVICE_ROLE_KEY: ${SUPABASE_SERVICE_ROLE_KEY:+set (hidden)}"
echo ""

# Check if we have a valid database connection method
if [ -z "$TEST_DATABASE_URL" ] && [ -z "$DATABASE_URL" ] && [ -z "$SUPABASE_URL" ]; then
    echo "⚠️  Warning: No database connection method found!"
    echo "   Please set one of:"
    echo "   - TEST_DATABASE_URL (direct PostgreSQL connection)"
    echo "   - DATABASE_URL (from Supabase dashboard - recommended)"
    echo "   - SUPABASE_URL + SUPABASE_SERVICE_ROLE_KEY (may require additional configuration)"
    echo ""
    echo "   See docs/supabase-database-connection.md for details"
    echo ""
fi

# Change to project root (stay here)
# Don't change to test/integration directory due to go.work configuration

echo "Running integration tests..."
echo ""

# Run the tests using specific file paths (works with go.work)
# This avoids the workspace pattern resolution issue
go test -tags=integration -v -run TestWeeks24Integration \
  ./test/integration/weeks_2_4_integration_test.go \
  ./test/integration/database_setup.go

echo ""
echo "=========================================="
echo "Test execution complete"
echo "=========================================="

