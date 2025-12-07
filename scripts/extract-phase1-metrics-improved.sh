#!/bin/bash

# Improved Phase 1 Metrics Extraction
# Handles Docker log format and extracts metrics from multiple sources

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

OUTPUT_FILE="${1:-/tmp/phase1-metrics-final.json}"

echo -e "${BLUE}=== Phase 1 Metrics Extraction (Improved) ===${NC}\n"

# Step 1: Get fresh logs
echo -e "${BLUE}Step 1: Fetching logs from Docker...${NC}"
docker compose -f docker-compose.local.yml logs classification-service --tail=10000 > /tmp/classification-logs-raw.txt 2>&1
echo -e "${GREEN}✅ Logs fetched${NC}\n"

# Step 2: Extract structured JSON
echo -e "${BLUE}Step 2: Parsing structured JSON logs...${NC}"
grep -E '^\s*kyb-classification-local\s+\|' /tmp/classification-logs-raw.txt | \
    sed 's/^[^|]*| //' | \
    jq -R 'fromjson? | select(. != null)' 2>/dev/null > /tmp/structured-logs.json || {
    echo -e "${YELLOW}⚠️  Could not parse all logs as JSON, trying alternative...${NC}"
    # Alternative: extract JSON objects directly
    grep -oE '\{[^{}]*"level"[^{}]*\}' /tmp/classification-logs-raw.txt | \
        jq -R 'fromjson? | select(. != null)' 2>/dev/null > /tmp/structured-logs.json || true
}

structured_count=$(jq -s 'length' /tmp/structured-logs.json 2>/dev/null || echo "0")
echo -e "${GREEN}✅ Parsed ${structured_count} structured log entries${NC}\n"

# Step 3: Extract quality scores
echo -e "${BLUE}Step 3: Extracting quality scores...${NC}"
quality_scores=$(jq -r 'select(.quality_score != null and (.quality_score | type == "number")) | .quality_score' /tmp/structured-logs.json 2>/dev/null | grep -v '^$' || echo "")

if [ -n "$quality_scores" ]; then
    quality_count=$(echo "$quality_scores" | grep -c . || echo "0")
    echo -e "${GREEN}✅ Found ${quality_count} quality scores${NC}"
    
    # Calculate statistics
    avg_quality=$(echo "$quality_scores" | awk '{sum+=$1; count++} END {if(count>0) printf "%.2f", sum/count; else print "0"}')
    high_quality=$(echo "$quality_scores" | awk '$1 >= 0.7 {count++} END {print count+0}')
    high_quality_pct=$(echo "scale=2; $high_quality * 100 / $quality_count" | bc 2>/dev/null || echo "0")
    
    echo -e "  Average: ${avg_quality}"
    echo -e "  Scores ≥0.7: ${high_quality} (${high_quality_pct}%)"
    
    if [ "$(echo "$high_quality_pct >= 90" | bc 2>/dev/null || echo "0")" = "1" ]; then
        echo -e "  Status: ${GREEN}✅ PASS${NC}"
    else
        echo -e "  Status: ${RED}❌ FAIL (target: ≥90%)${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  No quality scores found in structured logs${NC}"
    quality_count=0
    avg_quality=0
    high_quality=0
    high_quality_pct=0
fi
echo ""

# Step 4: Extract word counts
echo -e "${BLUE}Step 4: Extracting word counts...${NC}"
word_counts=$(jq -r 'select(.word_count != null and (.word_count | type == "number")) | .word_count' /tmp/structured-logs.json 2>/dev/null | grep -v '^$' || echo "")

if [ -n "$word_counts" ]; then
    word_count_count=$(echo "$word_counts" | grep -c . || echo "0")
    echo -e "${GREEN}✅ Found ${word_count_count} word counts${NC}"
    
    # Calculate statistics
    avg_words=$(echo "$word_counts" | awk '{sum+=$1; count++} END {if(count>0) printf "%.0f", sum/count; else print "0"}')
    median_words=$(echo "$word_counts" | sort -n | awk '{
        count[NR] = $1
    }
    END {
        if (NR % 2) {
            print count[(NR + 1) / 2]
        } else {
            print (count[NR / 2] + count[NR / 2 + 1]) / 2
        }
    }')
    
    echo -e "  Average: ${avg_words}"
    echo -e "  Median: ${median_words}"
    
    if [ "$(echo "$avg_words >= 200" | bc 2>/dev/null || echo "0")" = "1" ]; then
        echo -e "  Status: ${GREEN}✅ PASS${NC}"
    else
        echo -e "  Status: ${RED}❌ FAIL (target: ≥200)${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  No word counts found in structured logs${NC}"
    word_count_count=0
    avg_words=0
    median_words=0
fi
echo ""

# Step 5: Extract strategy distribution
echo -e "${BLUE}Step 5: Extracting strategy distribution...${NC}"
strategies=$(jq -r 'select(.strategy != null) | .strategy' /tmp/structured-logs.json 2>/dev/null | grep -v '^$' || echo "")

if [ -n "$strategies" ]; then
    strategy_count=$(echo "$strategies" | grep -c . || echo "0")
    echo -e "${GREEN}✅ Found ${strategy_count} strategy entries${NC}"
    
    simple_http=$(echo "$strategies" | grep -c "simple_http" || echo "0")
    browser_headers=$(echo "$strategies" | grep -c "browser_headers" || echo "0")
    playwright=$(echo "$strategies" | grep -c "playwright" || echo "0")
    
    total=$((simple_http + browser_headers + playwright))
    
    if [ $total -gt 0 ]; then
        simple_http_pct=$(echo "scale=2; $simple_http * 100 / $total" | bc 2>/dev/null || echo "0")
        browser_headers_pct=$(echo "scale=2; $browser_headers * 100 / $total" | bc 2>/dev/null || echo "0")
        playwright_pct=$(echo "scale=2; $playwright * 100 / $total" | bc 2>/dev/null || echo "0")
        
        echo -e "  SimpleHTTP: ${simple_http} (${simple_http_pct}%)"
        echo -e "  BrowserHeaders: ${browser_headers} (${browser_headers_pct}%)"
        echo -e "  Playwright: ${playwright} (${playwright_pct}%)"
    fi
else
    echo -e "${YELLOW}⚠️  No strategy entries found${NC}"
    strategy_count=0
    simple_http=0
    browser_headers=0
    playwright=0
fi
echo ""

# Step 6: Check for Phase 1 messages
echo -e "${BLUE}Step 6: Checking for Phase 1 log messages...${NC}"
phase1_messages=$(jq -r 'select(.msg | contains("Phase1") or contains("Phase 1")) | .msg' /tmp/structured-logs.json 2>/dev/null | head -10 || echo "")
if [ -n "$phase1_messages" ]; then
    phase1_count=$(echo "$phase1_messages" | grep -c . || echo "0")
    echo -e "${GREEN}✅ Found ${phase1_count} Phase 1 messages${NC}"
    echo "$phase1_messages" | head -5 | sed 's/^/  /'
else
    echo -e "${YELLOW}⚠️  No Phase 1 messages found in structured logs${NC}"
fi
echo ""

# Step 7: Save results
echo -e "${BLUE}Step 7: Saving results...${NC}"
cat > "$OUTPUT_FILE" <<EOF
{
  "extraction_timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "log_analysis": {
    "total_structured_logs": $structured_count,
    "quality_scores_found": $quality_count,
    "word_counts_found": $word_count_count,
    "strategy_entries_found": $strategy_count,
    "phase1_messages_found": ${phase1_count:-0}
  },
  "quality_scores": {
    "count": $quality_count,
    "average": ${avg_quality:-0},
    "high_quality_count": ${high_quality:-0},
    "high_quality_percentage": ${high_quality_pct:-0},
    "target_met": $(if [ "$(echo "${high_quality_pct:-0} >= 90" | bc 2>/dev/null || echo "0")" = "1" ]; then echo "true"; else echo "false"; fi)
  },
  "word_counts": {
    "count": $word_count_count,
    "average": ${avg_words:-0},
    "median": ${median_words:-0},
    "target_met": $(if [ "$(echo "${avg_words:-0} >= 200" | bc 2>/dev/null || echo "0")" = "1" ]; then echo "true"; else echo "false"; fi)
  },
  "strategies": {
    "total": ${strategy_count:-0},
    "simple_http": ${simple_http:-0},
    "browser_headers": ${browser_headers:-0},
    "playwright": ${playwright:-0}
  },
  "notes": {
    "phase1_logs_visible": $(if [ -n "$phase1_messages" ]; then echo "true"; else echo "false"; fi),
    "metrics_available": $(if [ $quality_count -gt 0 ] || [ $word_count_count -gt 0 ]; then echo "true"; else echo "false"; fi)
  }
}
EOF

echo -e "${GREEN}✅ Results saved to: ${OUTPUT_FILE}${NC}\n"

# Step 8: Summary
echo -e "${BLUE}=== Summary ===${NC}\n"
echo -e "Quality Scores: ${quality_count} found"
if [ $quality_count -gt 0 ]; then
    echo -e "  Average: ${avg_quality}"
    echo -e "  ≥0.7: ${high_quality} (${high_quality_pct}%)"
fi
echo ""
echo -e "Word Counts: ${word_count_count} found"
if [ $word_count_count -gt 0 ]; then
    echo -e "  Average: ${avg_words}"
fi
echo ""
echo -e "Strategies: ${strategy_count} found"
echo ""
if [ $quality_count -eq 0 ] && [ $word_count_count -eq 0 ]; then
    echo -e "${YELLOW}⚠️  No Phase 1 metrics found in logs${NC}"
    echo -e "Possible reasons:"
    echo -e "  1. Phase 1 scraper not being called (falling back to legacy)"
    echo -e "  2. Logs using different format (log.Printf vs structured)"
    echo -e "  3. Metrics logged but not captured in Docker logs"
fi

