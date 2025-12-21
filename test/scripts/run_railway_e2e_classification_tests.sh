#!/bin/bash

# Railway Comprehensive E2E Classification Test Runner
# This script runs comprehensive end-to-end tests against Railway production
# covering web scraping, crawling strategies, classification accuracy, and code/explanation generation

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
RAILWAY_API_URL="${RAILWAY_API_URL:-https://classification-service-production.up.railway.app}"
TIMEOUT="${TEST_TIMEOUT:-90m}"
VERBOSE="${VERBOSE:-true}"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Railway Comprehensive E2E Classification Tests${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${YELLOW}Configuration:${NC}"
echo "  API URL: $RAILWAY_API_URL"
echo "  Timeout: $TIMEOUT"
echo ""

# Warning about production testing
echo -e "${YELLOW}⚠️  WARNING: This will test against Railway PRODUCTION environment${NC}"
echo -e "${YELLOW}   - Real API calls will be made${NC}"
echo -e "${YELLOW}   - May incur costs for scraping and API usage${NC}"
echo -e "${YELLOW}   - May impact production metrics and logs${NC}"
echo ""
read -p "Continue? (yes/no): " confirm

if [[ "$confirm" != "yes" ]]; then
    echo -e "${RED}Test cancelled${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}🚀 Starting comprehensive E2E tests...${NC}"
echo ""

# Create results directory if it doesn't exist
mkdir -p test/results

# Run the tests
export RAILWAY_API_URL="$RAILWAY_API_URL"

if [ "$VERBOSE" = "true" ]; then
    go test -v -timeout "$TIMEOUT" -tags e2e_railway ./test/integration -run TestRailwayComprehensiveE2EClassification 2>&1 | tee "test/results/railway_e2e_test_output_$(date +%Y%m%d_%H%M%S).txt"
else
    go test -timeout "$TIMEOUT" -tags e2e_railway ./test/integration -run TestRailwayComprehensiveE2EClassification 2>&1 | tee "test/results/railway_e2e_test_output_$(date +%Y%m%d_%H%M%S).txt"
fi

TEST_EXIT_CODE=$?

echo ""
if [ $TEST_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✅ Tests completed successfully!${NC}"
    echo ""
    echo -e "${BLUE}📊 Results saved to:${NC}"
    echo "  - test/results/railway_e2e_classification_*.json"
    echo "  - test/results/railway_e2e_analysis_*.json"
    echo "  - test/results/railway_e2e_test_output_*.txt"
else
    echo -e "${RED}❌ Tests failed with exit code: $TEST_EXIT_CODE${NC}"
    echo ""
    echo -e "${YELLOW}Check the output log for details${NC}"
fi

echo ""
echo -e "${BLUE}========================================${NC}"

exit $TEST_EXIT_CODE

