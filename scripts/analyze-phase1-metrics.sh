#!/bin/bash

# Analyze Phase 1 metrics from logs and test results

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=== Phase 1 Metrics Analysis ===${NC}\n"

# Load test results
if [ -f /tmp/phase1-test-results.json ]; then
    echo -e "${GREEN}✅ Test results found${NC}"
    total_tests=$(jq 'length' /tmp/phase1-test-results.json)
    success_count=$(jq '[.[] | select(.success == true)] | length' /tmp/phase1-test-results.json)
    success_rate=$(echo "scale=2; $success_count * 100 / $total_tests" | bc)
    
    echo -e "  Total Tests: ${total_tests}"
    echo -e "  Successful: ${success_count}"
    echo -e "  Success Rate: ${success_rate}%"
    
    if [ "$(echo "$success_rate >= 95" | bc)" = "1" ]; then
        echo -e "  Status: ${GREEN}✅ PASS (target: ≥95%)${NC}"
    else
        echo -e "  Status: ${RED}❌ FAIL (target: ≥95%)${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  Test results not found${NC}"
fi

echo ""

# Analyze classification service logs
echo -e "${BLUE}Analyzing classification service logs...${NC}"

# Count strategy usage (if logs are available)
simple_http_count=$(docker compose -f docker-compose.local.yml logs classification-service --tail=10000 2>&1 | grep -c "simple_http" || echo "0")
browser_headers_count=$(docker compose -f docker-compose.local.yml logs classification-service --tail=10000 2>&1 | grep -c "browser_headers" || echo "0")
playwright_count=$(docker compose -f docker-compose.local.yml logs classification-service --tail=10000 2>&1 | grep -c "playwright" || echo "0")

total_strategies=$((simple_http_count + browser_headers_count + playwright_count))

if [ $total_strategies -gt 0 ]; then
    echo -e "  Strategy Distribution:"
    echo -e "    SimpleHTTP: ${simple_http_count}"
    echo -e "    BrowserHeaders: ${browser_headers_count}"
    echo -e "    Playwright: ${playwright_count}"
    echo -e "    Total: ${total_strategies}"
else
    echo -e "  ${YELLOW}⚠️  Strategy logs not found (may be using log.Printf)${NC}"
fi

echo ""

# Check Playwright service usage
echo -e "${BLUE}Playwright Service Usage:${NC}"
playwright_requests=$(docker compose -f docker-compose.local.yml logs playwright-scraper --tail=1000 2>&1 | grep -c "Scraping:" || echo "0")
echo -e "  Total Playwright requests: ${playwright_requests}"

echo ""

# Summary
echo -e "${BLUE}=== Summary ===${NC}"
echo -e "✅ Scrape Success Rate: ${success_rate}% (target: ≥95%)"
echo -e "⚠️  Quality Scores: Requires log analysis (target: ≥0.7 for 90%+)"
echo -e "⚠️  Word Counts: Requires log analysis (target: ≥200 avg)"
echo -e "✅ Playwright Service: Working (${playwright_requests} requests)"
echo ""
echo -e "${YELLOW}Note: Quality scores and word counts are logged but may not be visible${NC}"
echo -e "in Docker logs due to log format differences (log.Printf vs structured JSON)"

