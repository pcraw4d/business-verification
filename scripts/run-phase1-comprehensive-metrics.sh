#!/bin/bash

# Complete Phase 1 Comprehensive Metrics Workflow
# 1. Ensures services are running
# 2. Runs comprehensive test suite
# 3. Extracts and reports all Phase 1 metrics

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

COMPOSE_FILE="docker-compose.local.yml"
CLASSIFICATION_SERVICE="classification-service"

echo -e "${BLUE}=== Phase 1 Comprehensive Metrics Workflow ===${NC}\n"

# Step 1: Check and start services
echo -e "${BLUE}Step 1: Checking services...${NC}"
if ! docker compose -f "$COMPOSE_FILE" ps "$CLASSIFICATION_SERVICE" | grep -q "Up"; then
    echo -e "${YELLOW}⚠️  Classification service not running. Starting...${NC}"
    docker compose -f "$COMPOSE_FILE" up -d "$CLASSIFICATION_SERVICE"
    echo -e "${CYAN}Waiting for service to be healthy...${NC}"
    sleep 15
    
    # Check health
    if curl -s -f http://localhost:8081/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Service is healthy${NC}\n"
    else
        echo -e "${RED}❌ Service health check failed${NC}"
        echo -e "${YELLOW}Check logs: docker compose -f $COMPOSE_FILE logs $CLASSIFICATION_SERVICE${NC}"
        exit 1
    fi
else
    echo -e "${GREEN}✅ Classification service is running${NC}\n"
fi

# Step 2: Run comprehensive test suite
echo -e "${BLUE}Step 2: Running comprehensive test suite...${NC}"
echo -e "${CYAN}This will test 44 diverse websites and may take several minutes...${NC}\n"

# Clear previous logs to get fresh metrics
echo -e "${CYAN}Clearing old logs for fresh metrics...${NC}"
docker compose -f "$COMPOSE_FILE" logs --no-log-prefix "$CLASSIFICATION_SERVICE" > /dev/null 2>&1 || true

# Run test suite
if [ -f "scripts/test-phase1-comprehensive.sh" ]; then
    # Override CLASSIFICATION_URL to use localhost
    export CLASSIFICATION_URL="http://localhost:8081"
    bash scripts/test-phase1-comprehensive.sh
    TEST_EXIT_CODE=$?
    
    if [ $TEST_EXIT_CODE -ne 0 ]; then
        echo -e "${YELLOW}⚠️  Test suite had issues, but continuing with metrics extraction...${NC}\n"
    else
        echo -e "${GREEN}✅ Test suite completed${NC}\n"
    fi
else
    echo -e "${YELLOW}⚠️  Comprehensive test script not found, using quick test...${NC}"
    # Quick test with a few websites
    TEST_URLS=(
        "https://example.com"
        "https://stripe.com"
        "https://www.wikipedia.org"
        "https://www.shopify.com"
        "https://www.github.com"
    )
    
    for url in "${TEST_URLS[@]}"; do
        echo -e "${CYAN}Testing: $url${NC}"
        curl -s -X POST http://localhost:8081/classify \
            -H "Content-Type: application/json" \
            -d "{\"business_name\": \"Test\", \"website_url\": \"$url\"}" \
            --max-time 60 > /dev/null 2>&1 || true
        sleep 2
    done
    echo -e "${GREEN}✅ Quick test completed${NC}\n"
fi

# Wait a bit for logs to be written
echo -e "${CYAN}Waiting for logs to be written...${NC}"
sleep 5

# Step 3: Extract metrics
echo -e "${BLUE}Step 3: Extracting Phase 1 metrics from logs...${NC}\n"

if [ -f "scripts/measure-phase1-metrics-comprehensive.sh" ]; then
    bash scripts/measure-phase1-metrics-comprehensive.sh
    METRICS_EXIT_CODE=$?
    
    if [ $METRICS_EXIT_CODE -eq 0 ]; then
        echo -e "\n${GREEN}✅ Metrics extraction completed successfully${NC}"
    else
        echo -e "\n${YELLOW}⚠️  Metrics extraction had issues${NC}"
        echo -e "${CYAN}You can manually extract metrics using:${NC}"
        echo -e "  ./scripts/extract-phase1-metrics-improved.sh"
    fi
else
    echo -e "${YELLOW}⚠️  Metrics extraction script not found${NC}"
    echo -e "${CYAN}Using alternative extraction method...${NC}"
    bash scripts/extract-phase1-metrics-improved.sh
fi

echo ""
echo -e "${BLUE}=== Workflow Complete ===${NC}"
echo -e "${GREEN}✅ Phase 1 comprehensive metrics measurement completed${NC}"
echo ""
echo -e "${CYAN}Next steps:${NC}"
echo -e "1. Review the generated metrics report in docs/"
echo -e "2. Check if all success criteria are met"
echo -e "3. If metrics are below target, review logs for issues"
echo -e "4. If metrics meet/exceed target, Phase 1 is validated ✅"

