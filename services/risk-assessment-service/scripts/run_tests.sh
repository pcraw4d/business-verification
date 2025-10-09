#!/bin/bash

# Test runner script for the risk assessment service
# This script runs all tests and generates coverage reports

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVICE_DIR="services/risk-assessment-service"
COVERAGE_DIR="coverage"
COVERAGE_THRESHOLD=95

echo -e "${BLUE}🧪 Running Risk Assessment Service Tests${NC}"
echo "=================================================="

# Create coverage directory
mkdir -p "$COVERAGE_DIR"

# Change to service directory
cd "$SERVICE_DIR"

echo -e "${YELLOW}📦 Installing test dependencies...${NC}"
go mod tidy

echo -e "${YELLOW}🔍 Running unit tests...${NC}"

# Run tests with coverage
go test -v -race -coverprofile=../$COVERAGE_DIR/coverage.out ./...

# Generate coverage report
echo -e "${YELLOW}📊 Generating coverage report...${NC}"
go tool cover -html=../$COVERAGE_DIR/coverage.out -o ../$COVERAGE_DIR/coverage.html

# Get coverage percentage
COVERAGE_PERCENT=$(go tool cover -func=../$COVERAGE_DIR/coverage.out | grep total | awk '{print $3}' | sed 's/%//')

echo -e "${BLUE}📈 Coverage Report${NC}"
echo "=================="
go tool cover -func=../$COVERAGE_DIR/coverage.out

echo ""
echo -e "${BLUE}📊 Coverage Summary${NC}"
echo "=================="
echo -e "Total Coverage: ${COVERAGE_PERCENT}%"
echo -e "Target Coverage: ${COVERAGE_THRESHOLD}%"

# Check if coverage meets threshold
if (( $(echo "$COVERAGE_PERCENT >= $COVERAGE_THRESHOLD" | bc -l) )); then
    echo -e "${GREEN}✅ Coverage target met!${NC}"
    COVERAGE_STATUS="PASS"
else
    echo -e "${RED}❌ Coverage target not met!${NC}"
    COVERAGE_STATUS="FAIL"
fi

echo ""
echo -e "${YELLOW}🧪 Running benchmark tests...${NC}"

# Run benchmark tests
go test -bench=. -benchmem ./...

echo ""
echo -e "${YELLOW}🔍 Running race condition tests...${NC}"

# Run race condition tests
go test -race ./...

echo ""
echo -e "${YELLOW}📝 Running linting...${NC}"

# Run golangci-lint if available
if command -v golangci-lint &> /dev/null; then
    golangci-lint run
else
    echo -e "${YELLOW}⚠️  golangci-lint not found, skipping linting${NC}"
fi

echo ""
echo -e "${YELLOW}🔍 Running security scan...${NC}"

# Run gosec if available
if command -v gosec &> /dev/null; then
    gosec ./...
else
    echo -e "${YELLOW}⚠️  gosec not found, skipping security scan${NC}"
fi

echo ""
echo -e "${BLUE}📋 Test Summary${NC}"
echo "==============="
echo -e "Coverage: ${COVERAGE_PERCENT}% (Target: ${COVERAGE_THRESHOLD}%)"
echo -e "Status: ${COVERAGE_STATUS}"

if [ "$COVERAGE_STATUS" = "PASS" ]; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Tests failed!${NC}"
    exit 1
fi
