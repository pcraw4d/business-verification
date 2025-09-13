#!/bin/bash

# Merchant Portfolio Integration Test Runner
# This script runs comprehensive integration tests for the merchant portfolio system

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
TEST_DATABASE_URL="${TEST_DATABASE_URL:-postgres://postgres:password@localhost:5432/kyb_test?sslmode=disable}"
TEST_USER_ID="${TEST_USER_ID:-test-user-123}"
SKIP_DATABASE_TESTS="${SKIP_DATABASE_TESTS:-false}"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Merchant Portfolio Integration Tests${NC}"
echo -e "${BLUE}========================================${NC}"

# Function to print status
print_status() {
    local status=$1
    local message=$2
    
    if [ "$status" = "SUCCESS" ]; then
        echo -e "${GREEN}✓${NC} $message"
    elif [ "$status" = "ERROR" ]; then
        echo -e "${RED}✗${NC} $message"
    elif [ "$status" = "INFO" ]; then
        echo -e "${BLUE}ℹ${NC} $message"
    elif [ "$status" = "WARNING" ]; then
        echo -e "${YELLOW}⚠${NC} $message"
    fi
}

# Function to check if database is available
check_database() {
    print_status "INFO" "Checking database connectivity..."
    
    if [ "$SKIP_DATABASE_TESTS" = "true" ]; then
        print_status "WARNING" "Database tests skipped (SKIP_DATABASE_TESTS=true)"
        return 1
    fi
    
    # Try to connect to the database
    if command -v psql >/dev/null 2>&1; then
        if psql "$TEST_DATABASE_URL" -c "SELECT 1;" >/dev/null 2>&1; then
            print_status "SUCCESS" "Database connection successful"
            return 0
        else
            print_status "ERROR" "Cannot connect to database: $TEST_DATABASE_URL"
            return 1
        fi
    else
        print_status "WARNING" "psql not found, skipping database connectivity check"
        return 0
    fi
}

# Function to setup test database
setup_test_database() {
    print_status "INFO" "Setting up test database..."
    
    # Create test database if it doesn't exist
    if command -v psql >/dev/null 2>&1; then
        # Extract database name from URL
        DB_NAME=$(echo "$TEST_DATABASE_URL" | sed -n 's/.*\/\([^?]*\).*/\1/p')
        BASE_URL=$(echo "$TEST_DATABASE_URL" | sed 's/\/[^/]*$//')
        
        # Create database if it doesn't exist
        psql "$BASE_URL/postgres" -c "CREATE DATABASE $DB_NAME;" 2>/dev/null || true
        
        print_status "SUCCESS" "Test database setup complete"
    else
        print_status "WARNING" "psql not found, skipping database setup"
    fi
}

# Function to run database integration tests
run_database_tests() {
    print_status "INFO" "Running database integration tests..."
    
    export TEST_DATABASE_URL="$TEST_DATABASE_URL"
    export TEST_USER_ID="$TEST_USER_ID"
    
    if go test -v -run "TestMerchantPortfolioRepository_Integration" ./test/integration/...; then
        print_status "SUCCESS" "Database integration tests passed"
        return 0
    else
        print_status "ERROR" "Database integration tests failed"
        return 1
    fi
}

# Function to run service integration tests
run_service_tests() {
    print_status "INFO" "Running service integration tests..."
    
    export TEST_DATABASE_URL="$TEST_DATABASE_URL"
    export TEST_USER_ID="$TEST_USER_ID"
    
    if go test -v -run "TestMerchantPortfolioService_Integration" ./test/integration/...; then
        print_status "SUCCESS" "Service integration tests passed"
        return 0
    else
        print_status "ERROR" "Service integration tests failed"
        return 1
    fi
}

# Function to run API integration tests
run_api_tests() {
    print_status "INFO" "Running API integration tests..."
    
    export TEST_DATABASE_URL="$TEST_DATABASE_URL"
    export TEST_USER_ID="$TEST_USER_ID"
    
    if go test -v -run "TestMerchantPortfolioAPI_Integration" ./test/integration/...; then
        print_status "SUCCESS" "API integration tests passed"
        return 0
    else
        print_status "ERROR" "API integration tests failed"
        return 1
    fi
}

# Function to run error handling tests
run_error_handling_tests() {
    print_status "INFO" "Running error handling integration tests..."
    
    export TEST_DATABASE_URL="$TEST_DATABASE_URL"
    export TEST_USER_ID="$TEST_USER_ID"
    
    if go test -v -run "TestMerchantPortfolioErrorHandling_Integration" ./test/integration/...; then
        print_status "SUCCESS" "Error handling integration tests passed"
        return 0
    else
        print_status "ERROR" "Error handling integration tests failed"
        return 1
    fi
}

# Function to run performance tests
run_performance_tests() {
    print_status "INFO" "Running performance integration tests..."
    
    export TEST_DATABASE_URL="$TEST_DATABASE_URL"
    export TEST_USER_ID="$TEST_USER_ID"
    
    if go test -v -run "TestMerchantPortfolioPerformance_Integration" ./test/integration/...; then
        print_status "SUCCESS" "Performance integration tests passed"
        return 0
    else
        print_status "ERROR" "Performance integration tests failed"
        return 1
    fi
}

# Function to run all integration tests
run_all_tests() {
    print_status "INFO" "Running all integration tests..."
    
    export TEST_DATABASE_URL="$TEST_DATABASE_URL"
    export TEST_USER_ID="$TEST_USER_ID"
    
    if go test -v ./test/integration/...; then
        print_status "SUCCESS" "All integration tests passed"
        return 0
    else
        print_status "ERROR" "Some integration tests failed"
        return 1
    fi
}

# Function to generate test report
generate_test_report() {
    print_status "INFO" "Generating test report..."
    
    # Create reports directory if it doesn't exist
    mkdir -p test/reports
    
    # Run tests with coverage
    go test -v -coverprofile=test/reports/integration_coverage.out ./test/integration/...
    
    # Generate coverage report
    go tool cover -html=test/reports/integration_coverage.out -o test/reports/integration_coverage.html
    
    print_status "SUCCESS" "Test report generated in test/reports/"
}

# Main execution
main() {
    local test_type="${1:-all}"
    local generate_report="${2:-false}"
    
    # Check if we're in the right directory
    if [ ! -f "go.mod" ]; then
        print_status "ERROR" "Please run this script from the project root directory"
        exit 1
    fi
    
    # Check database connectivity
    if ! check_database; then
        print_status "WARNING" "Database not available, some tests may be skipped"
    else
        setup_test_database
    fi
    
    # Run tests based on type
    case "$test_type" in
        "database")
            run_database_tests
            ;;
        "service")
            run_service_tests
            ;;
        "api")
            run_api_tests
            ;;
        "error")
            run_error_handling_tests
            ;;
        "performance")
            run_performance_tests
            ;;
        "all")
            run_all_tests
            ;;
        *)
            echo "Usage: $0 [database|service|api|error|performance|all] [report]"
            echo "  database    - Run database integration tests"
            echo "  service     - Run service integration tests"
            echo "  api         - Run API integration tests"
            echo "  error       - Run error handling tests"
            echo "  performance - Run performance tests"
            echo "  all         - Run all integration tests (default)"
            echo "  report      - Generate test report (optional)"
            exit 1
            ;;
    esac
    
    # Generate report if requested
    if [ "$generate_report" = "true" ] || [ "$2" = "report" ]; then
        generate_test_report
    fi
    
    print_status "SUCCESS" "Integration test execution completed"
}

# Run main function with all arguments
main "$@"

