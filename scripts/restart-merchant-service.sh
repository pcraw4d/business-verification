#!/bin/bash

# Restart Merchant Service
# This script stops the running merchant service and starts it again
# Usage: ./scripts/restart-merchant-service.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

SERVICE_PORT=8083
SERVICE_DIR="services/merchant-service"
LOG_FILE="logs/merchant-service.log"
PID_FILE="logs/merchant-service.pid"

echo -e "${BLUE}🔄 Restarting Merchant Service...${NC}"
echo ""

# Create logs directory if it doesn't exist
mkdir -p logs

# Function to find and kill process on port 8083
kill_port_process() {
    local pid=$(lsof -ti:${SERVICE_PORT} 2>/dev/null || true)
    if [ -n "$pid" ]; then
        echo -e "${YELLOW}⚠️  Found process ${pid} on port ${SERVICE_PORT}${NC}"
        echo -e "${YELLOW}   Stopping process...${NC}"
        kill -TERM "$pid" 2>/dev/null || true
        sleep 2
        
        # Force kill if still running
        if lsof -ti:${SERVICE_PORT} >/dev/null 2>&1; then
            echo -e "${YELLOW}   Process still running, force killing...${NC}"
            kill -9 "$pid" 2>/dev/null || true
            sleep 1
        fi
        
        echo -e "${GREEN}✅ Process stopped${NC}"
    else
        echo -e "${GREEN}✅ No process found on port ${SERVICE_PORT}${NC}"
    fi
}

# Function to check if service is running
check_service() {
    local max_attempts=30
    local attempt=0
    
    echo -e "${YELLOW}⏳ Waiting for service to be ready...${NC}"
    
    while [ $attempt -lt $max_attempts ]; do
        if curl -s "http://localhost:${SERVICE_PORT}/health" > /dev/null 2>&1; then
            echo -e "${GREEN}✅ Service is healthy and ready!${NC}"
            return 0
        fi
        attempt=$((attempt + 1))
        sleep 1
        echo -n "."
    done
    
    echo ""
    echo -e "${RED}❌ Service failed to start within ${max_attempts} seconds${NC}"
    return 1
}

# Step 1: Stop existing service
echo -e "${BLUE}Step 1: Stopping existing service...${NC}"
kill_port_process

# Also remove PID file if it exists
if [ -f "$PID_FILE" ]; then
    rm -f "$PID_FILE"
fi

echo ""

# Step 2: Check if railway.env exists
if [ ! -f "railway.env" ]; then
    echo -e "${YELLOW}⚠️  Warning: railway.env file not found${NC}"
    echo -e "${YELLOW}   Starting service without environment file...${NC}"
else
    echo -e "${GREEN}✅ Found railway.env${NC}"
    source railway.env
fi

echo ""

# Step 3: Start service
echo -e "${BLUE}Step 2: Starting service...${NC}"
echo -e "${YELLOW}   Directory: ${SERVICE_DIR}${NC}"
echo -e "${YELLOW}   Port: ${SERVICE_PORT}${NC}"
echo -e "${YELLOW}   Log: ${LOG_FILE}${NC}"
echo ""

# Get absolute path to root directory
ROOT_DIR=$(pwd)

# Start service in background from root directory
echo -e "${GREEN}🚀 Starting Merchant Service...${NC}"
cd "${SERVICE_DIR}"

# Export environment variables so Go process can access them
export PORT=${SERVICE_PORT}
export SUPABASE_URL
export SUPABASE_ANON_KEY
export SUPABASE_SERVICE_ROLE_KEY
export SUPABASE_JWT_SECRET
export ENVIRONMENT

# Start the service with environment variables
env PORT=${SERVICE_PORT} \
    SUPABASE_URL="${SUPABASE_URL}" \
    SUPABASE_ANON_KEY="${SUPABASE_ANON_KEY}" \
    SUPABASE_SERVICE_ROLE_KEY="${SUPABASE_SERVICE_ROLE_KEY}" \
    SUPABASE_JWT_SECRET="${SUPABASE_JWT_SECRET}" \
    ENVIRONMENT="${ENVIRONMENT:-development}" \
    go run cmd/main.go > "${ROOT_DIR}/${LOG_FILE}" 2>&1 &
SERVICE_PID=$!
cd "${ROOT_DIR}"

# Save PID
echo $SERVICE_PID > "${PID_FILE}"

echo -e "${GREEN}✅ Service started (PID: ${SERVICE_PID})${NC}"
echo ""

# Step 4: Wait for service to be ready
check_service

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}✅ Merchant Service Restart Complete!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo "Service Information:"
    echo "  - URL: http://localhost:${SERVICE_PORT}"
    echo "  - Health: http://localhost:${SERVICE_PORT}/health"
    echo "  - PID: ${SERVICE_PID}"
    echo "  - Log: ${LOG_FILE}"
    echo ""
    echo -e "${BLUE}💡 To view logs: tail -f ${LOG_FILE}${NC}"
    echo -e "${BLUE}💡 To stop: kill ${SERVICE_PID} or ./scripts/stop-local-services.sh${NC}"
    echo ""
else
    echo ""
    echo -e "${RED}❌ Service restart failed${NC}"
    echo -e "${YELLOW}   Check logs: ${LOG_FILE}${NC}"
    exit 1
fi

