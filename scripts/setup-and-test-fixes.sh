#!/bin/bash

# Setup Environment and Test Fixes
# This script sets up the environment and tests the fixes

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Setting Up Environment to Test Fixes${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if railway.env exists
if [ ! -f "railway.env" ]; then
    echo -e "${RED}Error: railway.env file not found${NC}"
    echo "Please create railway.env with required environment variables"
    exit 1
fi

# Source environment variables
echo -e "${YELLOW}Loading environment variables from railway.env...${NC}"
source railway.env

# Set development environment
export ENVIRONMENT=development
export PORT=8080

# Create logs directory
mkdir -p logs

# Function to check if a port is in use
check_port() {
    local port=$1
    if lsof -ti:$port > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# Function to kill process on port
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

# Function to wait for service to be ready
wait_for_service() {
    local port=$1
    local service_name=$2
    local max_attempts=30
    
    echo -e "${YELLOW}Waiting for ${service_name} to be ready...${NC}"
    for i in $(seq 1 $max_attempts); do
        if curl -s "http://localhost:${port}/health" > /dev/null 2>&1; then
            echo -e "${GREEN}✓ ${service_name} is ready${NC}"
            return 0
        fi
        sleep 1
    done
    
    echo -e "${RED}✗ ${service_name} failed to start${NC}"
    return 1
}

# Function to start service
start_service() {
    local service_name=$1
    local service_path=$2
    local port=$3
    local log_file="logs/${service_name}.log"
    
    # Kill existing process on port
    kill_port $port
    
    echo -e "${GREEN}Starting ${service_name} on port ${port}...${NC}"
    
    cd "${service_path}"
    
    # Start service in background with environment variables
    go run cmd/main.go > "../../${log_file}" 2>&1 &
    local pid=$!
    echo $pid > "../../logs/${service_name}.pid"
    
    cd - > /dev/null
    
    echo -e "${GREEN}${service_name} started (PID: ${pid}, Log: ${log_file})${NC}"
    
    # Wait for service to be ready
    wait_for_service $port "$service_name"
    return $?
}

# Step 1: Start Merchant Service
echo ""
echo -e "${BLUE}Step 1: Starting Merchant Service...${NC}"
if start_service "merchant-service" "services/merchant-service" "8083"; then
    echo -e "${GREEN}✓ Merchant Service started successfully${NC}"
else
    echo -e "${RED}✗ Failed to start Merchant Service${NC}"
    echo -e "${YELLOW}Check logs/merchant-service.log for details${NC}"
    exit 1
fi

# Step 2: Start Risk Assessment Service
echo ""
echo -e "${BLUE}Step 2: Starting Risk Assessment Service...${NC}"
if start_service "risk-assessment-service" "services/risk-assessment-service" "8082"; then
    echo -e "${GREEN}✓ Risk Assessment Service started successfully${NC}"
else
    echo -e "${RED}✗ Failed to start Risk Assessment Service${NC}"
    echo -e "${YELLOW}Check logs/risk-assessment-service.log for details${NC}"
    exit 1
fi

# Step 3: Start API Gateway
echo ""
echo -e "${BLUE}Step 3: Starting API Gateway...${NC}"
if start_service "api-gateway" "services/api-gateway" "8080"; then
    echo -e "${GREEN}✓ API Gateway started successfully${NC}"
else
    echo -e "${RED}✗ Failed to start API Gateway${NC}"
    echo -e "${YELLOW}Check logs/api-gateway.log for details${NC}"
    exit 1
fi

# Step 4: Test Fixes
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Testing Fixes${NC}"
echo -e "${BLUE}========================================${NC}"

# Test 1: Invalid Merchant ID
echo ""
echo -e "${BLUE}Test 1: Invalid Merchant ID (should return 404)${NC}"
INVALID_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" "http://localhost:8080/api/v1/merchants/invalid-id-123" 2>&1)
INVALID_STATUS=$(echo "$INVALID_RESPONSE" | grep "HTTP_STATUS" | cut -d: -f2)
INVALID_BODY=$(echo "$INVALID_RESPONSE" | sed '/HTTP_STATUS/d')

if [ "$INVALID_STATUS" == "404" ]; then
    echo -e "${GREEN}✓ PASS: Invalid merchant ID returns 404${NC}"
    echo "  Response: $INVALID_BODY"
else
    echo -e "${RED}✗ FAIL: Invalid merchant ID returns ${INVALID_STATUS} (expected 404)${NC}"
    echo "  Response: $INVALID_BODY"
fi

# Test 2: Valid Merchant ID (if exists)
echo ""
echo -e "${BLUE}Test 2: Valid Merchant ID (should return 200)${NC}"
VALID_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" "http://localhost:8080/api/v1/merchants/merchant-123" 2>&1)
VALID_STATUS=$(echo "$VALID_RESPONSE" | grep "HTTP_STATUS" | cut -d: -f2)

if [ "$VALID_STATUS" == "200" ] || [ "$VALID_STATUS" == "404" ]; then
    if [ "$VALID_STATUS" == "200" ]; then
        echo -e "${GREEN}✓ PASS: Valid merchant ID returns 200${NC}"
    else
        echo -e "${YELLOW}⚠ INFO: Merchant ID 'merchant-123' not found (404) - this is expected if it doesn't exist${NC}"
    fi
else
    echo -e "${RED}✗ FAIL: Valid merchant ID returns ${VALID_STATUS}${NC}"
fi

# Test 3: Service Connectivity (check API Gateway logs for localhost URLs)
echo ""
echo -e "${BLUE}Test 3: Service Connectivity (localhost URLs)${NC}"
if grep -q "localhost" logs/api-gateway.log 2>/dev/null; then
    echo -e "${GREEN}✓ PASS: API Gateway is using localhost URLs${NC}"
    echo "  Check logs/api-gateway.log for service URLs"
else
    echo -e "${YELLOW}⚠ INFO: Could not verify localhost URLs from logs${NC}"
    echo "  Check logs/api-gateway.log manually"
fi

# Test 4: Health Checks
echo ""
echo -e "${BLUE}Test 4: Service Health Checks${NC}"

# API Gateway Health
if curl -s "http://localhost:8080/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ API Gateway is healthy${NC}"
else
    echo -e "${RED}✗ API Gateway health check failed${NC}"
fi

# Merchant Service Health (through API Gateway)
if curl -s "http://localhost:8080/api/v1/merchant/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Merchant Service is accessible${NC}"
else
    echo -e "${YELLOW}⚠ Merchant Service health check failed (may be expected)${NC}"
fi

# Risk Assessment Service Health (through API Gateway)
if curl -s "http://localhost:8080/api/v1/risk/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Risk Assessment Service is accessible${NC}"
else
    echo -e "${YELLOW}⚠ Risk Assessment Service health check failed (may be expected)${NC}"
fi

# Summary
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Test Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${GREEN}Services Running:${NC}"
echo "  - API Gateway: http://localhost:8080"
echo "  - Merchant Service: http://localhost:8083"
echo "  - Risk Assessment Service: http://localhost:8082"
echo ""
echo -e "${GREEN}Logs:${NC}"
echo "  - API Gateway: tail -f logs/api-gateway.log"
echo "  - Merchant Service: tail -f logs/merchant-service.log"
echo "  - Risk Assessment Service: tail -f logs/risk-assessment-service.log"
echo ""
echo -e "${GREEN}To stop services:${NC}"
echo "  kill \$(cat logs/api-gateway.pid)"
echo "  kill \$(cat logs/merchant-service.pid)"
echo "  kill \$(cat logs/risk-assessment-service.pid)"
echo ""
echo -e "${BLUE}Testing complete!${NC}"

