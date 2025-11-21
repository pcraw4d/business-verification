#!/bin/bash

# Restart Backend API Gateway Service
# This script stops the running backend and starts it again
# Usage: ./scripts/restart-backend.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

BACKEND_PORT=8080
BACKEND_DIR="services/api-gateway"
LOG_FILE="logs/api-gateway.log"
PID_FILE="logs/api-gateway.pid"

echo -e "${BLUE}🔄 Restarting API Gateway Backend...${NC}"
echo ""

# Create logs directory if it doesn't exist
mkdir -p logs

# Function to find and kill process on port 8080
kill_port_process() {
    local pid=$(lsof -ti:${BACKEND_PORT} 2>/dev/null || true)
    if [ -n "$pid" ]; then
        echo -e "${YELLOW}⚠️  Found process ${pid} on port ${BACKEND_PORT}${NC}"
        echo -e "${YELLOW}   Stopping process...${NC}"
        kill -TERM "$pid" 2>/dev/null || true
        sleep 2
        
        # Force kill if still running
        if lsof -ti:${BACKEND_PORT} >/dev/null 2>&1; then
            echo -e "${YELLOW}   Process still running, force killing...${NC}"
            kill -9 "$pid" 2>/dev/null || true
            sleep 1
        fi
        
        echo -e "${GREEN}✅ Process stopped${NC}"
    else
        echo -e "${GREEN}✅ No process found on port ${BACKEND_PORT}${NC}"
    fi
}

# Function to check if backend is running
check_backend() {
    local max_attempts=30
    local attempt=0
    
    echo -e "${YELLOW}⏳ Waiting for backend to be ready...${NC}"
    
    while [ $attempt -lt $max_attempts ]; do
        if curl -s "http://localhost:${BACKEND_PORT}/health" > /dev/null 2>&1; then
            echo -e "${GREEN}✅ Backend is healthy and ready!${NC}"
            return 0
        fi
        attempt=$((attempt + 1))
        sleep 1
        echo -n "."
    done
    
    echo ""
    echo -e "${RED}❌ Backend failed to start within ${max_attempts} seconds${NC}"
    return 1
}

# Step 1: Stop existing backend
echo -e "${BLUE}Step 1: Stopping existing backend...${NC}"
kill_port_process

# Also remove PID file if it exists
if [ -f "$PID_FILE" ]; then
    rm -f "$PID_FILE"
fi

echo ""

# Step 2: Check if railway.env exists
if [ ! -f "railway.env" ]; then
    echo -e "${YELLOW}⚠️  Warning: railway.env file not found${NC}"
    echo -e "${YELLOW}   Starting backend without environment file...${NC}"
else
    echo -e "${GREEN}✅ Found railway.env${NC}"
    source railway.env
fi

echo ""

# Step 3: Start backend
echo -e "${BLUE}Step 2: Starting backend...${NC}"
echo -e "${YELLOW}   Directory: ${BACKEND_DIR}${NC}"
echo -e "${YELLOW}   Port: ${BACKEND_PORT}${NC}"
echo -e "${YELLOW}   Log: ${LOG_FILE}${NC}"
echo ""

# Get absolute path to root directory
ROOT_DIR=$(pwd)

# Start backend in background from root directory
echo -e "${GREEN}🚀 Starting API Gateway...${NC}"
cd "${BACKEND_DIR}"

# Export environment variables so Go process can access them
export SUPABASE_URL
export SUPABASE_ANON_KEY
export SUPABASE_SERVICE_ROLE_KEY
export SUPABASE_JWT_SECRET
export ENVIRONMENT
export CORS_ALLOWED_ORIGINS
export CORS_ALLOWED_METHODS
export CORS_ALLOWED_HEADERS
export CORS_ALLOW_CREDENTIALS
export CORS_MAX_AGE

# Start the service with environment variables
env SUPABASE_URL="${SUPABASE_URL}" \
    SUPABASE_ANON_KEY="${SUPABASE_ANON_KEY}" \
    SUPABASE_SERVICE_ROLE_KEY="${SUPABASE_SERVICE_ROLE_KEY}" \
    SUPABASE_JWT_SECRET="${SUPABASE_JWT_SECRET}" \
    ENVIRONMENT="${ENVIRONMENT:-production}" \
    CORS_ALLOWED_ORIGINS="${CORS_ALLOWED_ORIGINS:-*}" \
    CORS_ALLOWED_METHODS="${CORS_ALLOWED_METHODS:-GET,POST,PUT,DELETE,OPTIONS}" \
    CORS_ALLOWED_HEADERS="${CORS_ALLOWED_HEADERS:-*}" \
    CORS_ALLOW_CREDENTIALS="${CORS_ALLOW_CREDENTIALS:-true}" \
    CORS_MAX_AGE="${CORS_MAX_AGE:-86400}" \
    go run cmd/main.go > "${ROOT_DIR}/${LOG_FILE}" 2>&1 &
BACKEND_PID=$!
cd "${ROOT_DIR}"

# Save PID
echo $BACKEND_PID > "${PID_FILE}"

cd - > /dev/null

echo -e "${GREEN}✅ Backend started (PID: ${BACKEND_PID})${NC}"
echo ""

# Step 4: Wait for backend to be ready
check_backend

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}✅ Backend Restart Complete!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo "Backend Information:"
    echo "  - URL: http://localhost:${BACKEND_PORT}"
    echo "  - Health: http://localhost:${BACKEND_PORT}/health"
    echo "  - PID: ${BACKEND_PID}"
    echo "  - Log: ${LOG_FILE}"
    echo ""
    echo -e "${BLUE}💡 To view logs: tail -f ${LOG_FILE}${NC}"
    echo -e "${BLUE}💡 To stop: kill ${BACKEND_PID} or ./scripts/stop-local-services.sh${NC}"
    echo ""
else
    echo ""
    echo -e "${RED}❌ Backend restart failed${NC}"
    echo -e "${YELLOW}   Check logs: ${LOG_FILE}${NC}"
    exit 1
fi

