#!/bin/bash

# Extract Phase 1 metrics from logs
# Extracts quality scores, word counts, and strategy distribution

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

LOG_FILE="${1:-/tmp/classification-logs.json}"
OUTPUT_FILE="${2:-/tmp/phase1-metrics.json}"

echo -e "${BLUE}=== Extracting Phase 1 Metrics from Logs ===${NC}\n"

# Check if log file exists
if [ ! -f "$LOG_FILE" ]; then
    echo -e "${YELLOW}Log file not found. Fetching from Docker...${NC}"
    docker compose -f docker-compose.local.yml logs classification-service --tail=10000 > "$LOG_FILE" 2>&1
fi

echo -e "Analyzing log file: ${LOG_FILE}\n"

# Extract structured JSON logs (remove Docker prefix and warnings)
grep -E '^\s*kyb-classification-local\s+\|' "$LOG_FILE" | \
    sed 's/^[^|]*| //' | \
    jq -R 'fromjson? | select(. != null)' 2>/dev/null > /tmp/structured-logs.json || {
    echo -e "${YELLOW}⚠️  Could not parse structured JSON logs${NC}"
    echo -e "Trying alternative extraction method...\n"
    
    # Alternative: extract JSON lines directly
    grep -oP '\{[^}]*"level"[^}]*\}' "$LOG_FILE" > /tmp/structured-logs.json || {
        echo -e "${RED}❌ Could not extract structured logs${NC}"
        exit 1
    }
}

# Extract metrics
echo -e "${BLUE}Extracting metrics...${NC}"

# Quality scores
quality_scores=$(jq -r 'select(.msg | contains("Strategy succeeded") or contains("quality_score")) | .quality_score // empty' /tmp/structured-logs.json 2>/dev/null | grep -v '^$' || echo "")

# Word counts
word_counts=$(jq -r 'select(.msg | contains("Strategy succeeded") or contains("word_count")) | .word_count // empty' /tmp/structured-logs.json 2>/dev/null | grep -v '^$' || echo "")

# Strategy names
strategies=$(jq -r 'select(.strategy != null) | .strategy' /tmp/structured-logs.json 2>/dev/null | grep -v '^$' || echo "")

# Count metrics
quality_count=$(echo "$quality_scores" | grep -c . || echo "0")
word_count_count=$(echo "$word_counts" | grep -c . || echo "0")
strategy_count=$(echo "$strategies" | grep -c . || echo "0")

echo -e "  Quality scores found: ${quality_count}"
echo -e "  Word counts found: ${word_count_count}"
echo -e "  Strategy entries found: ${strategy_count}\n"

# Calculate statistics if we have data
if [ "$quality_count" -gt 0 ]; then
    echo -e "${BLUE}Quality Score Analysis:${NC}"
    
    # Calculate average
    avg_quality=$(echo "$quality_scores" | awk '{sum+=$1; count++} END {if(count>0) print sum/count; else print "0"}')
    echo -e "  Average: ${avg_quality}"
    
    # Count ≥0.7
    high_quality=$(echo "$quality_scores" | awk '$1 >= 0.7 {count++} END {print count+0}')
    high_quality_pct=$(echo "scale=2; $high_quality * 100 / $quality_count" | bc)
    echo -e "  Scores ≥0.7: ${high_quality} (${high_quality_pct}%)"
    
    if [ "$(echo "$high_quality_pct >= 90" | bc)" = "1" ]; then
        echo -e "  Status: ${GREEN}✅ PASS (target: ≥90%)${NC}"
    else
        echo -e "  Status: ${RED}❌ FAIL (target: ≥90%)${NC}"
    fi
    
    echo ""
fi

if [ "$word_count_count" -gt 0 ]; then
    echo -e "${BLUE}Word Count Analysis:${NC}"
    
    # Calculate average
    avg_words=$(echo "$word_counts" | awk '{sum+=$1; count++} END {if(count>0) print sum/count; else print "0"}')
    echo -e "  Average: ${avg_words}"
    
    if [ "$(echo "$avg_words >= 200" | bc)" = "1" ]; then
        echo -e "  Status: ${GREEN}✅ PASS (target: ≥200)${NC}"
    else
        echo -e "  Status: ${RED}❌ FAIL (target: ≥200)${NC}"
    fi
    
    echo ""
fi

if [ "$strategy_count" -gt 0 ]; then
    echo -e "${BLUE}Strategy Distribution:${NC}"
    
    simple_http=$(echo "$strategies" | grep -c "simple_http" || echo "0")
    browser_headers=$(echo "$strategies" | grep -c "browser_headers" || echo "0")
    playwright=$(echo "$strategies" | grep -c "playwright" || echo "0")
    
    total=$((simple_http + browser_headers + playwright))
    
    if [ $total -gt 0 ]; then
        simple_http_pct=$(echo "scale=2; $simple_http * 100 / $total" | bc)
        browser_headers_pct=$(echo "scale=2; $browser_headers * 100 / $total" | bc)
        playwright_pct=$(echo "scale=2; $playwright * 100 / $total" | bc)
        
        echo -e "  SimpleHTTP: ${simple_http} (${simple_http_pct}%)"
        echo -e "  BrowserHeaders: ${browser_headers} (${browser_headers_pct}%)"
        echo -e "  Playwright: ${playwright} (${playwright_pct}%)"
    fi
    
    echo ""
fi

# Save results
cat > "$OUTPUT_FILE" <<EOF
{
  "quality_scores": {
    "count": $quality_count,
    "average": ${avg_quality:-0},
    "high_quality_count": ${high_quality:-0},
    "high_quality_percentage": ${high_quality_pct:-0}
  },
  "word_counts": {
    "count": $word_count_count,
    "average": ${avg_words:-0}
  },
  "strategies": {
    "total": $strategy_count,
    "simple_http": ${simple_http:-0},
    "browser_headers": ${browser_headers:-0},
    "playwright": ${playwright:-0}
  }
}
EOF

echo -e "${GREEN}✅ Metrics extracted and saved to: ${OUTPUT_FILE}${NC}"

