#!/bin/bash
# Monitor Circuit Breaker State for Python ML Service
# This script checks the circuit breaker state and metrics

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "🔍 Python ML Service Circuit Breaker Monitor"
echo "=============================================="
echo ""

# Check if Python ML service URL is set
PYTHON_ML_SERVICE_URL="${PYTHON_ML_SERVICE_URL:-http://localhost:8000}"

echo "📡 Checking Python ML Service at: $PYTHON_ML_SERVICE_URL"
echo ""

# Function to check health with circuit breaker info
check_circuit_breaker() {
    local url="$1"
    
    echo "🔄 Fetching circuit breaker status..."
    
    # Try to get health check with circuit breaker info
    response=$(curl -s -w "\n%{http_code}" "${url}/health" 2>/dev/null || echo "ERROR")
    
    if [[ "$response" == *"ERROR"* ]]; then
        echo -e "${RED}❌ Failed to connect to Python ML Service${NC}"
        echo "   Make sure the service is running at: $url"
        return 1
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" != "200" ]; then
        echo -e "${YELLOW}⚠️  Health check returned HTTP $http_code${NC}"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
        return 1
    fi
    
    # Parse JSON response
    if command -v jq &> /dev/null; then
        status=$(echo "$body" | jq -r '.status // "unknown"')
        checks=$(echo "$body" | jq -r '.checks // {}')
        
        echo -e "${GREEN}✅ Service Status: $status${NC}"
        echo ""
        
        # Check for circuit breaker info
        if echo "$checks" | jq -e '.circuit_breaker' > /dev/null 2>&1; then
            cb_status=$(echo "$checks" | jq -r '.circuit_breaker.status // "unknown"')
            cb_message=$(echo "$checks" | jq -r '.circuit_breaker.message // ""')
            
            echo "🔌 Circuit Breaker Status:"
            case "$cb_status" in
                "pass")
                    echo -e "   ${GREEN}State: CLOSED (operational)${NC}"
                    ;;
                "warn")
                    echo -e "   ${YELLOW}State: HALF-OPEN (testing recovery)${NC}"
                    ;;
                "fail")
                    echo -e "   ${RED}State: OPEN (circuit is open, requests rejected)${NC}"
                    ;;
                *)
                    echo -e "   ${YELLOW}State: $cb_status${NC}"
                    ;;
            esac
            
            if [ -n "$cb_message" ]; then
                echo "   Message: $cb_message"
            fi
        else
            echo -e "${YELLOW}⚠️  Circuit breaker information not available in health check${NC}"
            echo "   Response:"
            echo "$body" | jq '.' 2>/dev/null || echo "$body"
        fi
    else
        echo "📄 Health Check Response:"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
        echo ""
        echo -e "${YELLOW}💡 Tip: Install 'jq' for better JSON parsing: brew install jq${NC}"
    fi
}

# Function to get detailed metrics (if endpoint exists)
get_detailed_metrics() {
    local url="$1"
    
    echo ""
    echo "📊 Attempting to fetch detailed metrics..."
    
    # Try various possible endpoints
    for endpoint in "/metrics" "/health/detailed" "/circuit-breaker/metrics"; do
        response=$(curl -s -w "\n%{http_code}" "${url}${endpoint}" 2>/dev/null || echo "ERROR")
        
        if [[ "$response" != *"ERROR"* ]]; then
            http_code=$(echo "$response" | tail -n1)
            if [ "$http_code" == "200" ]; then
                body=$(echo "$response" | sed '$d')
                echo "✅ Found metrics at: ${endpoint}"
                echo "$body" | jq '.' 2>/dev/null || echo "$body"
                return 0
            fi
        fi
    done
    
    echo -e "${YELLOW}⚠️  Detailed metrics endpoint not found${NC}"
    echo "   The service may not expose a dedicated metrics endpoint"
}

# Main execution
check_circuit_breaker "$PYTHON_ML_SERVICE_URL"

# Optionally try to get detailed metrics
if [ "${1:-}" == "--detailed" ]; then
    get_detailed_metrics "$PYTHON_ML_SERVICE_URL"
fi

echo ""
echo "💡 Usage Tips:"
echo "   - Set PYTHON_ML_SERVICE_URL environment variable to change service URL"
echo "   - Use --detailed flag to attempt fetching detailed metrics"
echo "   - Monitor regularly: watch -n 5 $0"
echo ""

