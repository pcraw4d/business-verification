#!/bin/bash

# Extract Performance Metrics from Classification Service Logs
# Analyzes profiling data to measure optimization improvements

set -e

COMPOSE_FILE="docker-compose.local.yml"
SERVICE_NAME="classification-service"
OUTPUT_FILE="docs/performance-metrics-analysis-$(date +%Y%m%d-%H%M%S).md"

echo "=== Extracting Performance Metrics ==="

# Get recent logs
LOGS=$(docker compose -f "$COMPOSE_FILE" logs --tail=5000 "$SERVICE_NAME" 2>&1)

# Extract ClassifyBusinessByContextualKeywords metrics
echo "Extracting ClassifyBusinessByContextualKeywords metrics..."
CLASSIFY_ENTRIES=$(echo "$LOGS" | grep -o "ClassifyBusinessByContextualKeywords entry.*time remaining: [^,}]*" | grep -o "time remaining: [^,}]*" | sed 's/time remaining: //' | sed 's/s$//')

# Extract duration breakdowns
echo "Extracting duration breakdowns..."
DURATION_BREAKDOWNS=$(echo "$LOGS" | grep "Duration breakdown" | tail -20)

# Extract CalculateEnhancedScore durations
echo "Extracting CalculateEnhancedScore metrics..."
ENHANCED_SCORE_DURATIONS=$(echo "$LOGS" | grep -o "calculate_enhanced_score_duration: [^,}]*" | sed 's/calculate_enhanced_score_duration: //' | sed 's/s$//')

# Extract total classification durations
echo "Extracting total classification durations..."
TOTAL_DURATIONS=$(echo "$LOGS" | grep -o "classify_duration: [^,}]*" | sed 's/classify_duration: //' | sed 's/s$//')

# Extract time remaining violations
echo "Counting deadline violations..."
DEADLINE_VIOLATIONS=$(echo "$LOGS" | grep "time remaining: -" | wc -l | tr -d ' ')

# Extract parallel query metrics
echo "Extracting parallel query metrics..."
PARALLEL_QUERIES=$(echo "$LOGS" | grep -E "parallel queries|GetIndustryByID|GetCachedClassificationCodes" | tail -20)

# Generate report
cat > "$OUTPUT_FILE" <<EOF
# Performance Metrics Analysis

**Date:** $(date)  
**Source:** Classification Service Logs

---

## Summary

### Key Metrics

- **Total ClassifyBusinessByContextualKeywords Calls:** $(echo "$CLASSIFY_ENTRIES" | wc -l | tr -d ' ')
- **Context Deadline Violations:** $DEADLINE_VIOLATIONS
- **Average Classification Duration:** $(echo "$TOTAL_DURATIONS" | awk '{sum+=$1; count++} END {if(count>0) printf "%.2f", sum/count; else print "N/A"}')s
- **Average Enhanced Score Duration:** $(echo "$ENHANCED_SCORE_DURATIONS" | awk '{sum+=$1; count++} END {if(count>0) printf "%.2f", sum/count; else print "N/A"}')s

---

## Detailed Metrics

### ClassifyBusinessByContextualKeywords Entry Times

\`\`\`
$(echo "$CLASSIFY_ENTRIES" | head -20)
\`\`\`

### Duration Breakdowns

\`\`\`
$DURATION_BREAKDOWNS
\`\`\`

### Parallel Query Metrics

\`\`\`
$PARALLEL_QUERIES
\`\`\`

---

## Analysis

### Performance Improvements

1. **Classification Duration:** 
   - Target: < 10 seconds
   - Current: See metrics above

2. **Context Deadline Violations:**
   - Target: < 5%
   - Current: $DEADLINE_VIOLATIONS violations found

3. **Enhanced Score Duration:**
   - Target: < 8 seconds
   - Current: See metrics above

---

## Recommendations

Based on the metrics above, review:
1. extractKeywords duration (may be the bottleneck)
2. Context deadline management
3. Cache hit rates
4. Parallel query effectiveness

EOF

echo "✅ Metrics report generated: $OUTPUT_FILE"
cat "$OUTPUT_FILE"

