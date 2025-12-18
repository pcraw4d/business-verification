#!/bin/bash

# Comprehensive Classification E2E Test Runner Script
# This script runs the comprehensive classification tests and generates reports

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
TEST_DIR="$PROJECT_ROOT/test"
RESULTS_DIR="$TEST_DIR/results"

# Default to localhost, but allow override
# For Railway production, use: CLASSIFICATION_API_URL=https://classification-service-production.up.railway.app
API_URL="${CLASSIFICATION_API_URL:-http://localhost:8081}"

# Create results directory if it doesn't exist
mkdir -p "$RESULTS_DIR"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Comprehensive Classification E2E Tests${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "API URL: ${YELLOW}$API_URL${NC}"
echo -e "Results Directory: ${YELLOW}$RESULTS_DIR${NC}"
echo ""

# Check if API is accessible
echo -e "${BLUE}Checking API availability...${NC}"
if ! curl -s -f "$API_URL/health" > /dev/null 2>&1; then
    echo -e "${RED}❌ API is not accessible at $API_URL${NC}"
    echo -e "${YELLOW}Please ensure the classification service is running${NC}"
    exit 1
fi
echo -e "${GREEN}✅ API is accessible${NC}"
echo ""

# Check if test samples file exists
if [ ! -f "$TEST_DIR/data/comprehensive_test_samples.json" ]; then
    echo -e "${RED}❌ Test samples file not found: $TEST_DIR/data/comprehensive_test_samples.json${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Test samples file found${NC}"
echo ""

# Run tests
echo -e "${BLUE}Running comprehensive E2E tests...${NC}"
echo ""

cd "$PROJECT_ROOT"

# Set environment variable for API URL
export CLASSIFICATION_API_URL="$API_URL"

# Run Go tests
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
TEST_OUTPUT="$RESULTS_DIR/test_output_$TIMESTAMP.txt"

if go test -v -timeout 30m ./test/integration -run TestComprehensiveClassificationE2E 2>&1 | tee "$TEST_OUTPUT"; then
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
    
    # Generate additional reports if possible
    echo ""
    echo -e "${BLUE}Generating additional reports...${NC}"
    
    # Try to generate HTML report (if report generator is available)
    if command -v go &> /dev/null; then
        echo -e "${YELLOW}Note: HTML and Markdown reports can be generated programmatically${NC}"
    fi
    
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

