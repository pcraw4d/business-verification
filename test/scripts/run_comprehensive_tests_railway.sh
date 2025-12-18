#!/bin/bash

# Comprehensive Classification E2E Test Runner Script for Railway Production
# This script runs the comprehensive classification tests against Railway production environment

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Railway Production URLs
CLASSIFICATION_SERVICE_URL="https://classification-service-production.up.railway.app"
API_GATEWAY_URL="https://api-gateway-service-production-21fd.up.railway.app"

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
TEST_DIR="$PROJECT_ROOT/test"
RESULTS_DIR="$TEST_DIR/results"

# Allow override via environment variable
CLASSIFICATION_API_URL="${CLASSIFICATION_API_URL:-$CLASSIFICATION_SERVICE_URL}"
USE_API_GATEWAY="${USE_API_GATEWAY:-false}"

# If USE_API_GATEWAY is true, use API Gateway URL instead
if [ "$USE_API_GATEWAY" = "true" ]; then
    CLASSIFICATION_API_URL="$API_GATEWAY_URL/api/v1"
    echo -e "${CYAN}Using API Gateway endpoint${NC}"
else
    echo -e "${CYAN}Using direct Classification Service endpoint${NC}"
fi

# Create results directory if it doesn't exist
mkdir -p "$RESULTS_DIR"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Comprehensive Classification E2E Tests${NC}"
echo -e "${BLUE}Railway Production Environment${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "Environment: ${YELLOW}Railway Production${NC}"
echo -e "API URL: ${YELLOW}$CLASSIFICATION_API_URL${NC}"
echo -e "Results Directory: ${YELLOW}$RESULTS_DIR${NC}"
echo ""

# Check if API is accessible
echo -e "${BLUE}Checking API availability...${NC}"
HEALTH_ENDPOINT="$CLASSIFICATION_API_URL/health"
if [ "$USE_API_GATEWAY" = "true" ]; then
    HEALTH_ENDPOINT="$API_GATEWAY_URL/health"
fi

if ! curl -s -f -k --max-time 10 "$HEALTH_ENDPOINT" > /dev/null 2>&1; then
    echo -e "${RED}❌ API is not accessible at $HEALTH_ENDPOINT${NC}"
    echo -e "${YELLOW}Please verify:${NC}"
    echo -e "  1. Railway service is deployed and running"
    echo -e "  2. Service URL is correct: $CLASSIFICATION_API_URL"
    echo -e "  3. Network connectivity to Railway"
    exit 1
fi
echo -e "${GREEN}✅ API is accessible${NC}"

# Get service status
echo -e "${BLUE}Fetching service status...${NC}"
STATUS_RESPONSE=$(curl -s -k --max-time 10 "$HEALTH_ENDPOINT" 2>/dev/null || echo "")
if [ -n "$STATUS_RESPONSE" ]; then
    echo -e "${GREEN}✅ Service is healthy${NC}"
    echo -e "${CYAN}Status: $STATUS_RESPONSE${NC}" | head -c 100
    echo ""
else
    echo -e "${YELLOW}⚠ Could not fetch detailed status${NC}"
fi
echo ""

# Check if test samples file exists
if [ ! -f "$TEST_DIR/data/comprehensive_test_samples.json" ]; then
    echo -e "${RED}❌ Test samples file not found: $TEST_DIR/data/comprehensive_test_samples.json${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Test samples file found${NC}"
echo ""

# Warn about production testing
echo -e "${YELLOW}⚠ WARNING: Running tests against PRODUCTION environment${NC}"
echo -e "${YELLOW}This will make real API calls to Railway production services${NC}"
echo ""
read -p "Continue? (yes/no): " -r
if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
    echo -e "${RED}Test cancelled${NC}"
    exit 0
fi
echo ""

# Run tests
echo -e "${BLUE}Running comprehensive E2E tests...${NC}"
echo -e "${CYAN}This may take 15-30 minutes for 100 samples${NC}"
echo ""

cd "$PROJECT_ROOT"

# Set environment variable for API URL
export CLASSIFICATION_API_URL="$CLASSIFICATION_API_URL"

# Run Go tests with extended timeout for production
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
TEST_OUTPUT="$RESULTS_DIR/test_output_railway_$TIMESTAMP.txt"

echo -e "${BLUE}Starting test execution...${NC}"
echo -e "${CYAN}Output will be saved to: $TEST_OUTPUT${NC}"
echo ""

if go test -v -timeout 60m ./test/integration -run TestComprehensiveClassificationE2E 2>&1 | tee "$TEST_OUTPUT"; then
    echo ""
    echo -e "${GREEN}✅ Tests completed successfully${NC}"
else
    echo ""
    echo -e "${RED}❌ Tests failed${NC}"
    echo -e "Check output: ${YELLOW}$TEST_OUTPUT${NC}"
    exit 1
fi

# Check if results file was generated
RESULTS_FILE="$RESULTS_DIR/comprehensive_test_results.json"
if [ -f "$RESULTS_FILE" ]; then
    echo ""
    echo -e "${GREEN}✅ Test results saved to: $RESULTS_FILE${NC}"
    
    # Generate summary
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${GREEN}Test Summary${NC}"
    echo -e "${BLUE}========================================${NC}"
    
    # Extract summary from JSON if jq is available
    if command -v jq &> /dev/null; then
        echo ""
        echo "Test Summary:"
        jq -r '.test_summary | "Total: \(.total_samples) | Successful: \(.successful_tests) | Failed: \(.failed_tests) | Accuracy: \(.overall_accuracy * 100 | floor)%"' "$RESULTS_FILE" 2>/dev/null || echo "Could not parse summary"
        
        echo ""
        echo "Performance Metrics:"
        jq -r '.performance_metrics | "Avg Latency: \(.average_latency_ms)ms | P95: \(.p95_latency_ms)ms | Throughput: \(.throughput_rps | floor) req/s"' "$RESULTS_FILE" 2>/dev/null || echo "Could not parse performance metrics"
        
        echo ""
        echo "Strategy Distribution:"
        jq -r '.strategy_distribution.percentages | to_entries[] | "\(.key): \(.value | floor)%"' "$RESULTS_FILE" 2>/dev/null || echo "Could not parse strategy distribution"
        
        echo ""
        echo "Frontend Compatibility:"
        jq -r '.frontend_compatibility | "All Fields: \(.all_fields_present * 100 | floor)% | Industry: \(.industry_present * 100 | floor)% | Codes: \(.codes_present * 100 | floor)% | Explanation: \(.explanation_present * 100 | floor)%"' "$RESULTS_FILE" 2>/dev/null || echo "Could not parse frontend compatibility"
    else
        echo -e "${YELLOW}Install 'jq' for better summary output${NC}"
        echo "Results available in: $RESULTS_FILE"
    fi
else
    echo -e "${YELLOW}⚠ Results file not found: $RESULTS_FILE${NC}"
fi

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Test execution complete!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo "Results:"
echo "  - JSON Report: $RESULTS_FILE"
echo "  - Test Output: $TEST_OUTPUT"
echo ""
echo -e "${CYAN}Note: These tests ran against Railway production environment${NC}"
echo ""

