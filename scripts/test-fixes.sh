#!/bin/bash

# Test Fixes Script
# Restarts services with fixes and tests them

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Testing Fixes: Service Restart & Verification${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if services are running
check_port() {
    local port=$1
    if lsof -ti:$port > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# Kill process on port
kill_port() {
    local port=$1
    local pid=$(lsof -ti:$port 2>/dev/null)
    if [ -n "$pid" ]; then
        echo -e "${YELLOW}Stopping process on port ${port} (PID: ${pid})...${NC}"
        kill $pid 2>/dev/null || true
        sleep 2
        # Force kill if still running
        if kill -0 $pid 2>/dev/null; then
            kill -9 $pid 2>/dev/null || true
        fi
    fi
}

# Start service
start_service() {
    local service_name=$1
    local service_path=$2
    local port=$3
    local env_vars=$4
    
    echo -e "${GREEN}Starting ${service_name} on port ${port}...${NC}"
    
    cd "${service_path}"
    
    # Set environment variables
    export ENVIRONMENT=development
    export PORT=${port}
    eval "$env_vars"
    
    # Start service in background
    go run cmd/main.go > "/tmp/${service_name}.log" 2>&1 &
    local pid=$!
    echo $pid > "/tmp/${service_name}.pid"
    
    cd - > /dev/null
    
    echo -e "${GREEN}${service_name} started (PID: ${pid})${NC}"
    
    # Wait for service to be ready
    echo -e "${YELLOW}Waiting for ${service_name} to be ready...${NC}"
    for i in {1..30}; do
        if curl -s "http://localhost:${port}/health" > /dev/null 2>&1; then
            echo -e "${GREEN}✓ ${service_name} is ready${NC}"
            return 0
        fi
        sleep 1
    done
    
    echo -e "${RED}✗ ${service_name} failed to start${NC}"
    return 1
}

# Test invalid merchant ID
test_invalid_merchant_id() {
    echo ""
    echo -e "${BLUE}--- Testing Invalid Merchant ID Fix ---${NC}"
    
    local response=$(curl -s -w "\n%{http_code}" http://localhost:8080/api/v1/merchants/invalid-id-123 2>&1)
    local status_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    if [ "$status_code" == "404" ]; then
        echo -e "${GREEN}✓ Invalid merchant ID returns 404 (FIXED)${NC}"
        return 0
    else
        echo -e "${RED}✗ Invalid merchant ID returns ${status_code} (expected 404)${NC}"
        echo "Response: $body"
        return 1
    fi
}

# Test service connectivity
test_service_connectivity() {
    echo ""
    echo -e "${BLUE}--- Testing Service Connectivity ---${NC}"
    
    # Check API Gateway config
    local response=$(curl -s http://localhost:8080/health 2>&1)
    if echo "$response" | grep -q "healthy"; then
        echo -e "${GREEN}✓ API Gateway is running${NC}"
    else
        echo -e "${RED}✗ API Gateway is not responding${NC}"
        return 1
    fi
    
    # Test merchant service connectivity
    local merchant_health=$(curl -s http://localhost:8080/api/v1/merchant/health 2>&1)
    if echo "$merchant_health" | grep -q "healthy"; then
        echo -e "${GREEN}✓ Merchant service is accessible${NC}"
    else
        echo -e "${YELLOW}⚠ Merchant service may not be running locally${NC}"
    fi
    
    # Test risk assessment service connectivity
    local risk_health=$(curl -s http://localhost:8080/api/v1/risk/health 2>&1)
    if echo "$risk_health" | grep -q "healthy"; then
        echo -e "${GREEN}✓ Risk Assessment service is accessible${NC}"
    else
        echo -e "${YELLOW}⚠ Risk Assessment service may not be running locally${NC}"
    fi
    
    return 0
}

# Main execution
main() {
    echo -e "${BLUE}Step 1: Stopping existing services...${NC}"
    kill_port 8080  # API Gateway
    kill_port 8082  # Risk Assessment Service
    kill_port 8083  # Merchant Service
    kill_port 8084  # Risk Assessment Service (alternative port)
    sleep 2
    
    echo ""
    echo -e "${BLUE}Step 2: Starting services with fixes...${NC}"
    
    # Start Merchant Service (port 8083 as per start-local-services.sh)
    if ! check_port 8083; then
        start_service "merchant-service" "services/merchant-service" "8083" "export ENVIRONMENT=development"
    else
        echo -e "${YELLOW}Merchant service already running on port 8083${NC}"
    fi
    
    # Start Risk Assessment Service (port 8082 as per start-local-services.sh)
    if ! check_port 8082; then
        start_service "risk-assessment-service" "services/risk-assessment-service" "8082" "export ENVIRONMENT=development"
    else
        echo -e "${YELLOW}Risk Assessment service already running on port 8082${NC}"
    fi
    
    # Start API Gateway (port 8080)
    if ! check_port 8080; then
        start_service "api-gateway" "services/api-gateway" "8080" "export ENVIRONMENT=development"
    else
        echo -e "${YELLOW}API Gateway already running on port 8080${NC}"
    fi
    
    echo ""
    echo -e "${BLUE}Step 3: Testing fixes...${NC}"
    
    # Wait a bit for services to fully initialize
    sleep 3
    
    # Test invalid merchant ID
    test_invalid_merchant_id
    local invalid_id_test=$?
    
    # Test service connectivity
    test_service_connectivity
    local connectivity_test=$?
    
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}Test Results${NC}"
    echo -e "${BLUE}========================================${NC}"
    
    if [ $invalid_id_test -eq 0 ]; then
        echo -e "${GREEN}✓ Invalid Merchant ID Fix: PASSED${NC}"
    else
        echo -e "${RED}✗ Invalid Merchant ID Fix: FAILED${NC}"
    fi
    
    if [ $connectivity_test -eq 0 ]; then
        echo -e "${GREEN}✓ Service Connectivity: PASSED${NC}"
    else
        echo -e "${YELLOW}⚠ Service Connectivity: PARTIAL (some services may not be running)${NC}"
    fi
    
    echo ""
    echo -e "${BLUE}Service Logs:${NC}"
    echo "  Merchant Service: tail -f /tmp/merchant-service.log"
    echo "  Risk Assessment Service: tail -f /tmp/risk-assessment-service.log"
    echo "  API Gateway: tail -f /tmp/api-gateway.log"
    echo ""
    echo -e "${BLUE}To stop services:${NC}"
    echo "  kill \$(cat /tmp/merchant-service.pid)"
    echo "  kill \$(cat /tmp/risk-assessment-service.pid)"
    echo "  kill \$(cat /tmp/api-gateway.pid)"
}

main "$@"

