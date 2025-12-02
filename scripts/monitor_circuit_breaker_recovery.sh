#!/bin/bash
# Monitor Circuit Breaker Recovery
# Continuously monitors circuit breaker state until it closes

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

CLASSIFICATION_SERVICE_URL="${CLASSIFICATION_SERVICE_URL:-https://classification-service-production.up.railway.app}"
MAX_WAIT_TIME="${MAX_WAIT_TIME:-300}" # 5 minutes max
CHECK_INTERVAL="${CHECK_INTERVAL:-10}" # Check every 10 seconds

echo "🔍 Circuit Breaker Recovery Monitor"
echo "===================================="
echo ""
echo "Service: $CLASSIFICATION_SERVICE_URL"
echo "Max wait time: ${MAX_WAIT_TIME}s"
echo "Check interval: ${CHECK_INTERVAL}s"
echo ""

start_time=$(date +%s)
attempts=0

while true; do
    attempts=$((attempts + 1))
    elapsed=$(( $(date +%s) - start_time ))
    
    if [ $elapsed -gt $MAX_WAIT_TIME ]; then
        echo -e "${RED}❌ Timeout: Circuit breaker did not close within ${MAX_WAIT_TIME}s${NC}"
        exit 1
    fi
    
    echo "[Attempt $attempts, ${elapsed}s elapsed] Checking circuit breaker state..."
    
    # Get health status
    health_response=$(curl -s "${CLASSIFICATION_SERVICE_URL}/health" 2>/dev/null || echo "ERROR")
    
    if [[ "$health_response" == *"ERROR"* ]]; then
        echo -e "${RED}❌ Failed to connect to classification service${NC}"
        sleep $CHECK_INTERVAL
        continue
    fi
    
    # Extract circuit breaker state
    if command -v jq &> /dev/null; then
        cb_state=$(echo "$health_response" | jq -r '.ml_service_status.circuit_breaker_state // "unknown"')
        available=$(echo "$health_response" | jq -r '.ml_service_status.available // false')
        failure_count=$(echo "$health_response" | jq -r '.ml_service_status.circuit_breaker_metrics.failure_count // 0')
        success_count=$(echo "$health_response" | jq -r '.ml_service_status.circuit_breaker_metrics.success_count // 0')
    else
        # Fallback parsing without jq
        cb_state=$(echo "$health_response" | grep -o '"circuit_breaker_state":"[^"]*"' | cut -d'"' -f4 || echo "unknown")
        available=$(echo "$health_response" | grep -o '"available":[^,}]*' | grep -o 'true\|false' || echo "false")
    fi
    
    echo "   State: $cb_state"
    echo "   Available: $available"
    if [ -n "$failure_count" ] && [ "$failure_count" != "null" ]; then
        echo "   Failure Count: $failure_count"
    fi
    if [ -n "$success_count" ] && [ "$success_count" != "null" ]; then
        echo "   Success Count: $success_count"
    fi
    
    # Check if circuit breaker is closed
    if [ "$cb_state" == "CLOSED" ] || [ "$cb_state" == "closed" ]; then
        echo ""
        echo -e "${GREEN}✅ Circuit breaker is CLOSED!${NC}"
        echo "   ML service is now available for classification"
        echo ""
        
        # Test a classification to verify ML is being used
        echo "Testing classification to verify ML service usage..."
        test_response=$(curl -s -X POST "${CLASSIFICATION_SERVICE_URL}/v1/classify" \
            -H "Content-Type: application/json" \
            -d '{"business_name":"Acme Technology Corp","description":"Software development and cloud services","website_url":"https://www.acme.com"}' 2>/dev/null)
        
        if command -v jq &> /dev/null; then
            method=$(echo "$test_response" | jq -r '.classification_method // .method // "unknown"')
            industry=$(echo "$test_response" | jq -r '.industry_name // .primary_industry // "unknown"')
            confidence=$(echo "$test_response" | jq -r '.confidence_score // .confidence // 0')
            
            echo "   Classification Method: $method"
            echo "   Industry: $industry"
            echo "   Confidence: $confidence"
            
            if [[ "$method" == *"ml"* ]] || [[ "$method" == "ml_distilbart" ]]; then
                echo -e "${GREEN}✅ ML service is being used!${NC}"
            else
                echo -e "${YELLOW}⚠️  ML service not used (method: $method)${NC}"
            fi
        else
            echo "   Response received (jq not available for parsing)"
        fi
        
        exit 0
    elif [ "$cb_state" == "OPEN" ] || [ "$cb_state" == "open" ]; then
        echo -e "${YELLOW}   Circuit breaker is still OPEN${NC}"
        echo "   Waiting for recovery..."
    elif [ "$cb_state" == "HALF_OPEN" ] || [ "$cb_state" == "half_open" ]; then
        echo -e "${BLUE}   Circuit breaker is HALF_OPEN (testing recovery)${NC}"
        echo "   Attempting test request to help close circuit..."
        
        # Make a test request to help close the circuit
        curl -s -X POST "${CLASSIFICATION_SERVICE_URL}/v1/classify" \
            -H "Content-Type: application/json" \
            -d '{"business_name":"Test Recovery","description":"Testing circuit breaker recovery","website_url":"https://example.com"}' \
            > /dev/null 2>&1 || true
    else
        echo -e "${YELLOW}   Circuit breaker state: $cb_state${NC}"
    fi
    
    echo ""
    sleep $CHECK_INTERVAL
done

