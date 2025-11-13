#!/bin/bash

# Start Local Microservices Development Environment
# Alternative to Docker Compose - runs services directly with Go
# Usage: ./scripts/start-local-services.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if railway.env exists
if [ ! -f "railway.env" ]; then
    echo -e "${RED}Error: railway.env file not found${NC}"
    echo "Please create railway.env with required environment variables"
    exit 1
fi

# Source environment variables
source railway.env

# Create logs directory
mkdir -p logs

# Function to start a service
start_service() {
    local service_name=$1
    local service_path=$2
    local port=$3
    local log_file="logs/${service_name}.log"
    
    echo -e "${GREEN}Starting ${service_name} on port ${port}...${NC}"
    
    cd "${service_path}"
    PORT=${port} go run cmd/main.go > "../${log_file}" 2>&1 &
    local pid=$!
    echo $pid > "../logs/${service_name}.pid"
    cd ..
    
    echo -e "${GREEN}${service_name} started (PID: ${pid}, Log: ${log_file})${NC}"
    sleep 2
}

# Function to check if a service is running
check_service() {
    local port=$1
    local service_name=$2
    
    if curl -s "http://localhost:${port}/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ ${service_name} is healthy${NC}"
        return 0
    else
        echo -e "${YELLOW}⚠ ${service_name} not yet ready${NC}"
        return 1
    fi
}

# Start Redis (if not running)
if ! pgrep -x "redis-server" > /dev/null; then
    echo -e "${GREEN}Starting Redis...${NC}"
    redis-server --port 6379 --daemonize yes
    sleep 1
else
    echo -e "${YELLOW}Redis already running${NC}"
fi

# Start services
echo -e "${GREEN}Starting KYB Platform microservices...${NC}"

start_service "classification-service" "services/classification-service" "8081"
start_service "merchant-service" "services/merchant-service" "8083"
start_service "risk-assessment-service" "services/risk-assessment-service" "8082"
start_service "api-gateway" "services/api-gateway" "8080"
start_service "frontend" "services/frontend" "8086"

# Wait for services to be ready
echo -e "${YELLOW}Waiting for services to be ready...${NC}"
sleep 5

# Check service health
echo -e "${GREEN}Checking service health...${NC}"
check_service "8081" "Classification Service" || true
check_service "8083" "Merchant Service" || true
check_service "8082" "Risk Assessment Service" || true
check_service "8080" "API Gateway" || true
check_service "8086" "Frontend" || true

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}KYB Platform Services Started${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Services:"
echo "  - Classification Service: http://localhost:8081"
echo "  - Merchant Service:        http://localhost:8083"
echo "  - Risk Assessment Service: http://localhost:8082"
echo "  - API Gateway:             http://localhost:8080"
echo "  - Frontend:                http://localhost:8086"
echo ""
echo "Logs are in the logs/ directory"
echo "To stop services, run: ./scripts/stop-local-services.sh"
echo ""

