#!/bin/bash

# Comprehensive Phase 1 Metrics Measurement Script
# Extracts all Phase 1 success criteria from Docker logs and generates detailed report

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
CLASSIFICATION_SERVICE="classification-service"
COMPOSE_FILE="docker-compose.local.yml"
REPORT_FILE="docs/phase1-metrics-report-$(date +%Y%m%d-%H%M%S).md"
LOG_START_TIME=$(date -u -v-10M '+%Y-%m-%dT%H:%M:%SZ' 2>/dev/null || date -u -d '10 minutes ago' '+%Y-%m-%dT%H:%M:%SZ' 2>/dev/null || echo "")

echo -e "${BLUE}=== Phase 1 Comprehensive Metrics Measurement ===${NC}\n"

# Check if Docker Compose is available
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker is not installed or not in PATH${NC}"
    exit 1
fi

# Check if classification service exists (running or stopped, we just need logs)
if ! docker compose -f "$COMPOSE_FILE" ps -a "$CLASSIFICATION_SERVICE" | grep -q "$CLASSIFICATION_SERVICE"; then
    echo -e "${RED}❌ Classification service container not found${NC}"
    echo -e "${YELLOW}Start it with: docker compose -f $COMPOSE_FILE up -d $CLASSIFICATION_SERVICE${NC}"
    exit 1
fi

# Check if service is running (preferred) but continue if stopped (may have logs)
if docker compose -f "$COMPOSE_FILE" ps "$CLASSIFICATION_SERVICE" | grep -q "Up"; then
    echo -e "${GREEN}✅ Classification service is running${NC}\n"
else
    echo -e "${YELLOW}⚠️  Classification service is not running, but extracting metrics from available logs...${NC}\n"
fi

# Extract logs
echo -e "${BLUE}Extracting Phase 1 metrics from logs...${NC}"

# Get logs since last 30 minutes (or all if LOG_START_TIME is empty)
if [ -n "$LOG_START_TIME" ]; then
    LOGS=$(docker compose -f "$COMPOSE_FILE" logs --since="$LOG_START_TIME" "$CLASSIFICATION_SERVICE" 2>&1)
else
    LOGS=$(docker compose -f "$COMPOSE_FILE" logs --tail=10000 "$CLASSIFICATION_SERVICE" 2>&1)
fi

# Extract metrics using jq
echo -e "${CYAN}Analyzing logs...${NC}"

# Count total scrape attempts
TOTAL_ATTEMPTS=$(echo "$LOGS" | jq -r 'select(.msg | contains("Starting scrape with structured content extraction"))' 2>/dev/null | wc -l | tr -d ' ')

# Count successful scrapes
SUCCESSFUL_SCRAPES=$(echo "$LOGS" | jq -r 'select(.msg | contains("Strategy succeeded"))' 2>/dev/null | wc -l | tr -d ' ')

# Count failed scrapes
FAILED_SCRAPES=$(echo "$LOGS" | jq -r 'select(.msg | contains("All scraping strategies failed"))' 2>/dev/null | wc -l | tr -d ' ')

# Extract quality scores
QUALITY_SCORES=$(echo "$LOGS" | jq -r 'select(.quality_score != null and .quality_score > 0) | .quality_score' 2>/dev/null)

# Extract word counts
WORD_COUNTS=$(echo "$LOGS" | jq -r 'select(.word_count != null and .word_count > 0) | .word_count' 2>/dev/null)

# Extract strategy distribution
SIMPLE_HTTP_COUNT=$(echo "$LOGS" | jq -r 'select(.msg | contains("Strategy succeeded") and .strategy == "simple_http")' 2>/dev/null | wc -l | tr -d ' ')
BROWSER_HEADERS_COUNT=$(echo "$LOGS" | jq -r 'select(.msg | contains("Strategy succeeded") and .strategy == "browser_headers")' 2>/dev/null | wc -l | tr -d ' ')
PLAYWRIGHT_COUNT=$(echo "$LOGS" | jq -r 'select(.msg | contains("Strategy succeeded") and .strategy == "playwright")' 2>/dev/null | wc -l | tr -d ' ')

# Calculate metrics
if [ "$TOTAL_ATTEMPTS" -gt 0 ]; then
    SUCCESS_RATE=$(echo "scale=2; $SUCCESSFUL_SCRAPES * 100 / $TOTAL_ATTEMPTS" | bc)
else
    SUCCESS_RATE=0
fi

# Calculate quality score metrics
if [ -n "$QUALITY_SCORES" ]; then
    QUALITY_COUNT=$(echo "$QUALITY_SCORES" | wc -l | tr -d ' ')
    QUALITY_ABOVE_07=$(echo "$QUALITY_SCORES" | awk '$1 >= 0.7' | wc -l | tr -d ' ')
    if [ "$QUALITY_COUNT" -gt 0 ]; then
        QUALITY_PERCENT_ABOVE_07=$(echo "scale=2; $QUALITY_ABOVE_07 * 100 / $QUALITY_COUNT" | bc)
        AVG_QUALITY=$(echo "$QUALITY_SCORES" | awk '{sum+=$1; count++} END {if(count>0) print sum/count; else print 0}')
    else
        QUALITY_PERCENT_ABOVE_07=0
        AVG_QUALITY=0
    fi
else
    QUALITY_COUNT=0
    QUALITY_ABOVE_07=0
    QUALITY_PERCENT_ABOVE_07=0
    AVG_QUALITY=0
fi

# Calculate word count metrics
if [ -n "$WORD_COUNTS" ]; then
    WORD_COUNT_TOTAL=$(echo "$WORD_COUNTS" | wc -l | tr -d ' ')
    AVG_WORD_COUNT=$(echo "$WORD_COUNTS" | awk '{sum+=$1; count++} END {if(count>0) print int(sum/count); else print 0}')
    MIN_WORD_COUNT=$(echo "$WORD_COUNTS" | sort -n | head -1)
    MAX_WORD_COUNT=$(echo "$WORD_COUNTS" | sort -n | tail -1)
else
    WORD_COUNT_TOTAL=0
    AVG_WORD_COUNT=0
    MIN_WORD_COUNT=0
    MAX_WORD_COUNT=0
fi

# Calculate strategy distribution
TOTAL_STRATEGY_USAGE=$((SIMPLE_HTTP_COUNT + BROWSER_HEADERS_COUNT + PLAYWRIGHT_COUNT))
if [ "$TOTAL_STRATEGY_USAGE" -gt 0 ]; then
    SIMPLE_HTTP_PERCENT=$(echo "scale=2; $SIMPLE_HTTP_COUNT * 100 / $TOTAL_STRATEGY_USAGE" | bc)
    BROWSER_HEADERS_PERCENT=$(echo "scale=2; $BROWSER_HEADERS_COUNT * 100 / $TOTAL_STRATEGY_USAGE" | bc)
    PLAYWRIGHT_PERCENT=$(echo "scale=2; $PLAYWRIGHT_COUNT * 100 / $TOTAL_STRATEGY_USAGE" | bc)
else
    SIMPLE_HTTP_PERCENT=0
    BROWSER_HEADERS_PERCENT=0
    PLAYWRIGHT_PERCENT=0
fi

# Generate report
echo -e "${BLUE}Generating metrics report...${NC}\n"

mkdir -p docs

cat > "$REPORT_FILE" << EOF
# Phase 1 Comprehensive Metrics Report

**Generated:** $(date)
**Service:** $CLASSIFICATION_SERVICE
**Log Period:** Last 30 minutes (or all available logs)

---

## Executive Summary

| Metric | Result | Target | Status |
|--------|--------|--------|--------|
| **Scrape Success Rate** | ${SUCCESS_RATE}% | ≥95% | $(if (( $(echo "$SUCCESS_RATE >= 95" | bc -l) )); then echo "✅ PASS"; else echo "❌ FAIL"; fi) |
| **Quality Score (≥0.7)** | ${QUALITY_PERCENT_ABOVE_07}% | ≥90% | $(if (( $(echo "$QUALITY_PERCENT_ABOVE_07 >= 90" | bc -l) )); then echo "✅ PASS"; else echo "❌ FAIL"; fi) |
| **Average Word Count** | ${AVG_WORD_COUNT} | ≥200 | $(if [ "$AVG_WORD_COUNT" -ge 200 ]; then echo "✅ PASS"; else echo "❌ FAIL"; fi) |
| **"No Output" Errors** | ${FAILED_SCRAPES} (${FAILED_SCRAPES}%) | <2% | $(if (( $(echo "$FAILED_SCRAPES * 100 / ($TOTAL_ATTEMPTS + 1) < 2" | bc -l) )); then echo "✅ PASS"; else echo "❌ FAIL"; fi) |

---

## Detailed Metrics

### Scrape Success Rate

- **Total Attempts:** $TOTAL_ATTEMPTS
- **Successful:** $SUCCESSFUL_SCRAPES
- **Failed:** $FAILED_SCRAPES
- **Success Rate:** ${SUCCESS_RATE}%
- **Target:** ≥95%
- **Status:** $(if (( $(echo "$SUCCESS_RATE >= 95" | bc -l) )); then echo "✅ PASS"; else echo "❌ FAIL"; fi)

### Content Quality Scores

- **Total Quality Scores:** $QUALITY_COUNT
- **Scores ≥0.7:** $QUALITY_ABOVE_07
- **Percentage ≥0.7:** ${QUALITY_PERCENT_ABOVE_07}%
- **Average Quality Score:** ${AVG_QUALITY}
- **Target:** ≥0.7 for 90%+ of scrapes
- **Status:** $(if (( $(echo "$QUALITY_PERCENT_ABOVE_07 >= 90" | bc -l) )); then echo "✅ PASS"; else echo "❌ FAIL"; fi)

### Word Count Metrics

- **Total Word Counts:** $WORD_COUNT_TOTAL
- **Average Word Count:** $AVG_WORD_COUNT
- **Minimum Word Count:** $MIN_WORD_COUNT
- **Maximum Word Count:** $MAX_WORD_COUNT
- **Target:** Average ≥200 words
- **Status:** $(if [ "$AVG_WORD_COUNT" -ge 200 ]; then echo "✅ PASS"; else echo "❌ FAIL"; fi)

### Strategy Distribution

- **SimpleHTTP:** $SIMPLE_HTTP_COUNT (${SIMPLE_HTTP_PERCENT}%)
- **BrowserHeaders:** $BROWSER_HEADERS_COUNT (${BROWSER_HEADERS_PERCENT}%)
- **Playwright:** $PLAYWRIGHT_COUNT (${PLAYWRIGHT_PERCENT}%)
- **Total Strategy Usage:** $TOTAL_STRATEGY_USAGE

**Expected Distribution:**
- SimpleHTTP: ~60%
- BrowserHeaders: ~20-30%
- Playwright: ~10-20%

---

## Success Criteria Assessment

### ✅ Scrape Success Rate: $(if (( $(echo "$SUCCESS_RATE >= 95" | bc -l) )); then echo "PASS"; else echo "FAIL"; fi)

$(if (( $(echo "$SUCCESS_RATE >= 95" | bc -l) )); then echo "✅ **PASS** - Success rate of ${SUCCESS_RATE}% exceeds target of ≥95%"; else echo "❌ **FAIL** - Success rate of ${SUCCESS_RATE}% is below target of ≥95%"; fi)

### ✅ Content Quality Score: $(if (( $(echo "$QUALITY_PERCENT_ABOVE_07 >= 90" | bc -l) )); then echo "PASS"; else echo "FAIL"; fi)

$(if (( $(echo "$QUALITY_PERCENT_ABOVE_07 >= 90" | bc -l) )); then echo "✅ **PASS** - ${QUALITY_PERCENT_ABOVE_07}% of scrapes have quality score ≥0.7 (target: ≥90%)"; else echo "❌ **FAIL** - ${QUALITY_PERCENT_ABOVE_07}% of scrapes have quality score ≥0.7 (target: ≥90%)"; fi)

### ✅ Average Word Count: $(if [ "$AVG_WORD_COUNT" -ge 200 ]; then echo "PASS"; else echo "FAIL"; fi)

$(if [ "$AVG_WORD_COUNT" -ge 200 ]; then echo "✅ **PASS** - Average word count of $AVG_WORD_COUNT exceeds target of ≥200"; else echo "❌ **FAIL** - Average word count of $AVG_WORD_COUNT is below target of ≥200"; fi)

### ✅ "No Output" Errors: $(if (( $(echo "$FAILED_SCRAPES * 100 / ($TOTAL_ATTEMPTS + 1) < 2" | bc -l) )); then echo "PASS"; else echo "FAIL"; fi)

$(if (( $(echo "$FAILED_SCRAPES * 100 / ($TOTAL_ATTEMPTS + 1) < 2" | bc -l) )); then echo "✅ **PASS** - Error rate of ${FAILED_SCRAPES}% is below target of <2%"; else echo "❌ **FAIL** - Error rate exceeds target of <2%"; fi)

---

## Recommendations

$(if [ "$TOTAL_ATTEMPTS" -eq 0 ]; then
    echo "- ⚠️ **No scraping attempts found in logs** - Run comprehensive test suite first"
    echo "  - Execute: ./scripts/test-phase1-comprehensive.sh"
    echo "  - Then re-run this metrics script"
elif [ "$TOTAL_ATTEMPTS" -lt 20 ]; then
    echo "- ⚠️ **Limited test data** - Only $TOTAL_ATTEMPTS attempts found"
    echo "  - Recommend running comprehensive test suite with 50-100 websites"
    echo "  - Execute: ./scripts/test-phase1-comprehensive.sh"
fi)

$(if (( $(echo "$SUCCESS_RATE < 95" | bc -l) )); then
    echo "- ❌ **Scrape success rate below target** - Investigate failed scrapes"
    echo "  - Check logs for error patterns"
    echo "  - Verify Playwright service is accessible"
    echo "  - Check network connectivity"
fi)

$(if (( $(echo "$QUALITY_PERCENT_ABOVE_07 < 90" | bc -l) )); then
    echo "- ❌ **Quality scores below target** - Review content extraction"
    echo "  - Check if structured content extraction is working"
    echo "  - Verify about section extraction"
    echo "  - Review quality score calculation"
fi)

$(if [ "$AVG_WORD_COUNT" -lt 200 ]; then
    echo "- ❌ **Average word count below target** - Review content extraction"
    echo "  - Check if text extraction is complete"
    echo "  - Verify about section extraction is working"
    echo "  - Review content validation thresholds"
fi)

---

## Next Steps

1. **If metrics are below target:**
   - Review logs for error patterns
   - Check service health
   - Verify environment variables
   - Test with individual websites

2. **If metrics meet/exceed target:**
   - ✅ Phase 1 implementation validated
   - Proceed with Phase 2 implementation
   - Document findings

---

**Report Generated:** $(date)
EOF

echo -e "${GREEN}✅ Metrics report generated: $REPORT_FILE${NC}\n"

# Display summary
echo -e "${BLUE}=== Metrics Summary ===${NC}\n"
echo -e "Scrape Success Rate: ${CYAN}${SUCCESS_RATE}%${NC} (target: ≥95%)"
echo -e "Quality Score (≥0.7): ${CYAN}${QUALITY_PERCENT_ABOVE_07}%${NC} (target: ≥90%)"
echo -e "Average Word Count: ${CYAN}${AVG_WORD_COUNT}${NC} (target: ≥200)"
echo -e "Strategy Distribution:"
echo -e "  - SimpleHTTP: ${CYAN}${SIMPLE_HTTP_PERCENT}%${NC}"
echo -e "  - BrowserHeaders: ${CYAN}${BROWSER_HEADERS_PERCENT}%${NC}"
echo -e "  - Playwright: ${CYAN}${PLAYWRIGHT_PERCENT}%${NC}"
echo ""
echo -e "${GREEN}✅ Detailed report saved to: $REPORT_FILE${NC}"

