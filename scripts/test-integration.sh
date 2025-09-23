#!/bin/bash

# KYB Platform Integration Test Script
# Comprehensive testing before deployment

set -e  # Exit on any error

echo "🧪 KYB Platform Integration Test Suite"
echo "======================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test results tracking
TESTS_PASSED=0
TESTS_FAILED=0

# Function to print test results
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ $2${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}❌ $2${NC}"
        ((TESTS_FAILED++))
    fi
}

# Function to run a test
run_test() {
    local test_name="$1"
    local test_command="$2"
    
    echo -e "${BLUE}Running: $test_name${NC}"
    if eval "$test_command"; then
        print_result 0 "$test_name"
    else
        print_result 1 "$test_name"
    fi
    echo ""
}

# Check prerequisites
echo "🔍 Checking Prerequisites..."
echo "============================="

# Check Go installation
if command -v go &> /dev/null; then
    print_result 0 "Go is installed ($(go version))"
else
    print_result 1 "Go is not installed"
    exit 1
fi

# Check Node.js installation
if command -v node &> /dev/null; then
    print_result 0 "Node.js is installed ($(node --version))"
else
    print_result 1 "Node.js is not installed"
    exit 1
fi

# Check npm installation
if command -v npm &> /dev/null; then
    print_result 0 "npm is installed ($(npm --version))"
else
    print_result 1 "npm is not installed"
    exit 1
fi

echo ""

# Backend Tests
echo "🔧 Backend Tests"
echo "================"

# Go module validation
run_test "Go module validation" "go mod verify"

# Go formatting check
run_test "Go code formatting" "go fmt ./..."

# Go vet check
run_test "Go vet analysis" "go vet ./..."

# Go tests
run_test "Go unit tests" "go test -v ./..."

# Go build test
run_test "Go build test" "go build -o kyb-platform ./cmd/railway-server"

# Clean up build artifact
rm -f kyb-platform

echo ""

# Frontend Tests
echo "🎨 Frontend Tests"
echo "================="

# Check if web directory exists
if [ -d "web" ]; then
    cd web
    
    # npm install
    run_test "npm install" "npm ci"
    
    # npm linting
    run_test "npm linting" "npm run lint"
    
    # npm tests
    run_test "npm tests" "npm test"
    
    # npm build
    run_test "npm build" "npm run build"
    
    # Check build artifacts
    run_test "Build artifacts check" "test -f dist/real-data-integration.*.js && test -f dist/merchant-dashboard.*.js"
    
    cd ..
else
    print_result 1 "Web directory not found"
fi

echo ""

# Integration Tests
echo "🔗 Integration Tests"
echo "===================="

# Test API endpoints
echo "Testing API endpoints..."

# Health check
if curl -f -s https://shimmering-comfort-production.up.railway.app/health > /dev/null; then
    print_result 0 "Health endpoint"
else
    print_result 1 "Health endpoint"
fi

# Classification endpoint
if curl -f -s -X POST https://shimmering-comfort-production.up.railway.app/v1/classify \
    -H "Content-Type: application/json" \
    -d '{"business_name": "Test Company", "description": "Test"}' > /dev/null; then
    print_result 0 "Classification endpoint"
else
    print_result 1 "Classification endpoint"
fi

# Merchants endpoint
if curl -f -s https://shimmering-comfort-production.up.railway.app/api/v1/merchants > /dev/null; then
    print_result 0 "Merchants endpoint"
else
    print_result 1 "Merchants endpoint"
fi

# Analytics endpoint
if curl -f -s https://shimmering-comfort-production.up.railway.app/api/v1/merchants/analytics > /dev/null; then
    print_result 0 "Analytics endpoint"
else
    print_result 1 "Analytics endpoint"
fi

echo ""

# Database Tests
echo "🗄️ Database Tests"
echo "=================="

# Check if migration files exist
run_test "Migration files exist" "test -f supabase-full-integration-migration.sql"

# Test database connection via API
echo "Testing database connection..."
HEALTH_RESPONSE=$(curl -s https://shimmering-comfort-production.up.railway.app/health)
if echo "$HEALTH_RESPONSE" | grep -q '"connected":true'; then
    print_result 0 "Database connection"
else
    print_result 1 "Database connection"
fi

echo ""

# Security Tests
echo "🔒 Security Tests"
echo "================="

# Go security scan
run_test "Go security scan" "go list -json -deps ./... | nancy sleuth"

# npm audit
if [ -d "web" ]; then
    cd web
    run_test "npm security audit" "npm audit --audit-level=moderate"
    cd ..
fi

echo ""

# File Structure Tests
echo "📁 File Structure Tests"
echo "======================="

# Check for required files
run_test "Go module file" "test -f go.mod"
run_test "Railway config" "test -f railway.json"
run_test "Main application" "test -f cmd/railway-server/main.go"
run_test "Real data integration" "test -f web/components/real-data-integration.js"
run_test "Merchant dashboard" "test -f web/merchant-dashboard-real-data.js"
run_test "Monitoring dashboard" "test -f web/monitoring-dashboard-real-data.js"
run_test "Bulk operations" "test -f web/merchant-bulk-operations-real-data.js"
run_test "Main dashboard" "test -f web/dashboard-real-data.js"

echo ""

# Performance Tests
echo "⚡ Performance Tests"
echo "===================="

# Test API response times
echo "Testing API response times..."

# Health endpoint response time
HEALTH_TIME=$(curl -w "%{time_total}" -o /dev/null -s https://shimmering-comfort-production.up.railway.app/health)
if (( $(echo "$HEALTH_TIME < 2.0" | bc -l) )); then
    print_result 0 "Health endpoint response time (${HEALTH_TIME}s)"
else
    print_result 1 "Health endpoint response time (${HEALTH_TIME}s)"
fi

# Merchants endpoint response time
MERCHANTS_TIME=$(curl -w "%{time_total}" -o /dev/null -s https://shimmering-comfort-production.up.railway.app/api/v1/merchants)
if (( $(echo "$MERCHANTS_TIME < 3.0" | bc -l) )); then
    print_result 0 "Merchants endpoint response time (${MERCHANTS_TIME}s)"
else
    print_result 1 "Merchants endpoint response time (${MERCHANTS_TIME}s)"
fi

echo ""

# Final Results
echo "📊 Test Results Summary"
echo "======================="
echo -e "${GREEN}Tests Passed: $TESTS_PASSED${NC}"
echo -e "${RED}Tests Failed: $TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}🎉 All tests passed! Ready for deployment.${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Commit your changes: git add . && git commit -m 'feat: implement real data integration'"
    echo "2. Push to main branch: git push origin main"
    echo "3. Monitor deployment in GitHub Actions"
    exit 0
else
    echo -e "${RED}❌ Some tests failed. Please fix the issues before deploying.${NC}"
    echo ""
    echo "Failed tests need to be addressed before deployment."
    exit 1
fi
