#!/bin/bash

# Test Circuit Breaker Recovery
# Tests if the circuit breaker has recovered and ML service is working

set -e

CLASSIFICATION_API_URL="https://classification-service-production.up.railway.app"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Circuit Breaker Recovery Test${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "API URL: ${CYAN}$CLASSIFICATION_API_URL${NC}"
echo ""

# Step 1: Check initial circuit breaker state
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Step 1: Check Initial Circuit Breaker State${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

HEALTH_RESPONSE=$(curl -s "$CLASSIFICATION_API_URL/health")
CB_STATE=$(echo "$HEALTH_RESPONSE" | python3 -c "import sys, json; d=json.load(sys.stdin); print(d.get('ml_service_status', {}).get('circuit_breaker_state', 'unknown'))" 2>/dev/null || echo "unknown")
CB_FAILURES=$(echo "$HEALTH_RESPONSE" | python3 -c "import sys, json; d=json.load(sys.stdin); print(d.get('ml_service_status', {}).get('circuit_breaker_metrics', {}).get('failure_count', 0))" 2>/dev/null || echo "0")
CB_SUCCESSES=$(echo "$HEALTH_RESPONSE" | python3 -c "import sys, json; d=json.load(sys.stdin); print(d.get('ml_service_status', {}).get('circuit_breaker_metrics', {}).get('success_count', 0))" 2>/dev/null || echo "0")
CB_TOTAL=$(echo "$HEALTH_RESPONSE" | python3 -c "import sys, json; d=json.load(sys.stdin); print(d.get('ml_service_status', {}).get('circuit_breaker_metrics', {}).get('total_requests', 0))" 2>/dev/null || echo "0")
CB_REJECTED=$(echo "$HEALTH_RESPONSE" | python3 -c "import sys, json; d=json.load(sys.stdin); print(d.get('ml_service_status', {}).get('circuit_breaker_metrics', {}).get('rejected_requests', 0))" 2>/dev/null || echo "0")
PYTHON_HEALTH=$(echo "$HEALTH_RESPONSE" | python3 -c "import sys, json; d=json.load(sys.stdin); print(d.get('ml_service_status', {}).get('health_checks', {}).get('python_ml_service', {}).get('status', 'unknown'))" 2>/dev/null || echo "unknown")

echo -e "Circuit Breaker State: ${CYAN}$CB_STATE${NC}"
echo -e "Failure Count: ${CYAN}$CB_FAILURES${NC}"
echo -e "Success Count: ${CYAN}$CB_SUCCESSES${NC}"
echo -e "Total Requests: ${CYAN}$CB_TOTAL${NC}"
echo -e "Rejected Requests: ${CYAN}$CB_REJECTED${NC}"
echo -e "Python ML Service Health: ${CYAN}$PYTHON_HEALTH${NC}"
echo ""

# Step 2: Test classification request
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Step 2: Test Classification Request${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

TEST_DATA='{
  "business_name": "Tech Startup Inc",
  "description": "Software development and cloud consulting services",
  "website_url": "https://techstartup.example.com"
}'

echo -e "Test Request Data:"
echo "$TEST_DATA" | python3 -m json.tool 2>/dev/null || echo "$TEST_DATA"
echo ""

START_TIME=$(date +%s.%N)
CLASSIFY_RESPONSE=$(curl -s -X POST "$CLASSIFICATION_API_URL/v1/classify" \
  -H "Content-Type: application/json" \
  -d "$TEST_DATA" \
  --max-time 60 \
  -w "\n%{http_code}" 2>&1)
END_TIME=$(date +%s.%N)
DURATION=$(echo "$END_TIME - $START_TIME" | bc)

HTTP_CODE=$(echo "$CLASSIFY_RESPONSE" | tail -n1)
RESPONSE_BODY=$(echo "$CLASSIFY_RESPONSE" | sed '$d')

echo -e "HTTP Status: ${CYAN}$HTTP_CODE${NC}"
echo -e "Response Time: ${CYAN}${DURATION}s${NC}"
echo ""

if [ "$HTTP_CODE" = "200" ]; then
    SUCCESS=$(echo "$RESPONSE_BODY" | python3 -c "import sys, json; d=json.load(sys.stdin); print(d.get('success', False))" 2>/dev/null || echo "false")
    STATUS=$(echo "$RESPONSE_BODY" | python3 -c "import sys, json; d=json.load(sys.stdin); print(d.get('status', 'unknown'))" 2>/dev/null || echo "unknown")
    INDUSTRY=$(echo "$RESPONSE_BODY" | python3 -c "import sys, json; d=json.load(sys.stdin); print(d.get('primary_industry', 'N/A'))" 2>/dev/null || echo "N/A")
    CONFIDENCE=$(echo "$RESPONSE_BODY" | python3 -c "import sys, json; d=json.load(sys.stdin); print(d.get('confidence_score', 0))" 2>/dev/null || echo "0")
    REQUEST_ID=$(echo "$RESPONSE_BODY" | python3 -c "import sys, json; d=json.load(sys.stdin); print(d.get('request_id', 'N/A'))" 2>/dev/null || echo "N/A")
    
    echo -e "Success: ${CYAN}$SUCCESS${NC}"
    echo -e "Status: ${CYAN}$STATUS${NC}"
    echo -e "Primary Industry: ${CYAN}$INDUSTRY${NC}"
    echo -e "Confidence Score: ${CYAN}$CONFIDENCE${NC}"
    echo -e "Request ID: ${CYAN}$REQUEST_ID${NC}"
    
    if [ "$SUCCESS" = "True" ] || [ "$SUCCESS" = "true" ]; then
        echo -e "${GREEN}✅ Classification request succeeded${NC}"
        REQUEST_SUCCESS=true
    else
        echo -e "${RED}❌ Classification request failed (success=false)${NC}"
        REQUEST_SUCCESS=false
    fi
else
    echo -e "${RED}❌ Classification request failed (HTTP $HTTP_CODE)${NC}"
    echo "Response:"
    echo "$RESPONSE_BODY" | head -10
    REQUEST_SUCCESS=false
fi

echo ""

# Step 3: Check circuit breaker state after request
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Step 3: Check Circuit Breaker State After Request${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

sleep 2
HEALTH_RESPONSE_AFTER=$(curl -s "$CLASSIFICATION_API_URL/health")
CB_STATE_AFTER=$(echo "$HEALTH_RESPONSE_AFTER" | python3 -c "import sys, json; d=json.load(sys.stdin); print(d.get('ml_service_status', {}).get('circuit_breaker_state', 'unknown'))" 2>/dev/null || echo "unknown")
CB_FAILURES_AFTER=$(echo "$HEALTH_RESPONSE_AFTER" | python3 -c "import sys, json; d=json.load(sys.stdin); print(d.get('ml_service_status', {}).get('circuit_breaker_metrics', {}).get('failure_count', 0))" 2>/dev/null || echo "0")
CB_SUCCESSES_AFTER=$(echo "$HEALTH_RESPONSE_AFTER" | python3 -c "import sys, json; d=json.load(sys.stdin); print(d.get('ml_service_status', {}).get('circuit_breaker_metrics', {}).get('success_count', 0))" 2>/dev/null || echo "0")
CB_TOTAL_AFTER=$(echo "$HEALTH_RESPONSE_AFTER" | python3 -c "import sys, json; d=json.load(sys.stdin); print(d.get('ml_service_status', {}).get('circuit_breaker_metrics', {}).get('total_requests', 0))" 2>/dev/null || echo "0")
CB_REJECTED_AFTER=$(echo "$HEALTH_RESPONSE_AFTER" | python3 -c "import sys, json; d=json.load(sys.stdin); print(d.get('ml_service_status', {}).get('circuit_breaker_metrics', {}).get('rejected_requests', 0))" 2>/dev/null || echo "0")

echo -e "Circuit Breaker State: ${CYAN}$CB_STATE_AFTER${NC}"
echo -e "Failure Count: ${CYAN}$CB_FAILURES_AFTER${NC}"
echo -e "Success Count: ${CYAN}$CB_SUCCESSES_AFTER${NC}"
echo -e "Total Requests: ${CYAN}$CB_TOTAL_AFTER${NC}"
echo -e "Rejected Requests: ${CYAN}$CB_REJECTED_AFTER${NC}"
echo ""

# Step 4: Summary
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Test Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

if [ "$CB_STATE" = "open" ] && [ "$CB_STATE_AFTER" = "closed" ]; then
    echo -e "${GREEN}✅ Circuit Breaker Recovered${NC}"
    echo -e "  - State changed from OPEN to CLOSED"
    echo -e "  - Automatic recovery is working"
elif [ "$CB_STATE" = "open" ] && [ "$CB_STATE_AFTER" = "half-open" ]; then
    echo -e "${YELLOW}🔄 Circuit Breaker Transitioning${NC}"
    echo -e "  - State changed from OPEN to HALF_OPEN"
    echo -e "  - Recovery in progress"
elif [ "$CB_STATE" = "closed" ]; then
    echo -e "${GREEN}✅ Circuit Breaker is CLOSED${NC}"
    echo -e "  - Circuit breaker is healthy"
elif [ "$CB_STATE" = "half-open" ]; then
    echo -e "${YELLOW}🔄 Circuit Breaker is HALF_OPEN${NC}"
    echo -e "  - Testing recovery"
else
    echo -e "${RED}❌ Circuit Breaker is OPEN${NC}"
    echo -e "  - State: $CB_STATE"
    echo -e "  - Failures: $CB_FAILURES"
    echo -e "  - May need manual reset"
fi

echo ""

if [ "$REQUEST_SUCCESS" = "true" ]; then
    echo -e "${GREEN}✅ Classification Request Succeeded${NC}"
    echo -e "  - ML service is working"
    echo -e "  - Industry: $INDUSTRY"
    echo -e "  - Confidence: $CONFIDENCE"
else
    echo -e "${RED}❌ Classification Request Failed${NC}"
    echo -e "  - HTTP Status: $HTTP_CODE"
    echo -e "  - May be due to circuit breaker blocking requests"
fi

echo ""

# Check if manual reset is needed
if [ "$CB_STATE" = "open" ] && [ "$CB_STATE_AFTER" = "open" ] && [ "$PYTHON_HEALTH" = "pass" ]; then
    echo -e "${YELLOW}⚠️  Recommendation:${NC}"
    echo -e "  - Service is healthy but circuit breaker is still open"
    echo -e "  - Consider using manual reset endpoint:"
    echo -e "    POST /admin/circuit-breaker/reset"
    echo -e "    Header: X-Admin-Key: <admin-key>"
fi

echo ""
echo -e "${BLUE}========================================${NC}"

