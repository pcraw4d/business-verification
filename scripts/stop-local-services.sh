#!/bin/bash

# Stop Local Microservices Development Environment
# Usage: ./scripts/stop-local-services.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Stopping KYB Platform services...${NC}"

# Stop services by PID files
if [ -d "logs" ]; then
    for pid_file in logs/*.pid; do
        if [ -f "$pid_file" ]; then
            service_name=$(basename "$pid_file" .pid)
            pid=$(cat "$pid_file")
            
            if ps -p "$pid" > /dev/null 2>&1; then
                echo -e "${GREEN}Stopping ${service_name} (PID: ${pid})...${NC}"
                kill "$pid" 2>/dev/null || true
                rm "$pid_file"
            fi
        fi
    done
fi

# Kill any remaining Go processes for our services
pkill -f "services/classification-service/cmd/main.go" 2>/dev/null || true
pkill -f "services/merchant-service/cmd/main.go" 2>/dev/null || true
pkill -f "services/risk-assessment-service/cmd/main.go" 2>/dev/null || true
pkill -f "services/api-gateway/cmd/main.go" 2>/dev/null || true
pkill -f "services/frontend/cmd/main.go" 2>/dev/null || true

# Stop Redis if started by our script
if pgrep -x "redis-server" > /dev/null; then
    echo -e "${GREEN}Stopping Redis...${NC}"
    pkill -x "redis-server" || true
fi

echo -e "${GREEN}All services stopped${NC}"

