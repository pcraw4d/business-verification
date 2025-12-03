#!/bin/bash

# Start Classification Service Locally
# Run this in a separate terminal window

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=== Starting Classification Service ===${NC}\n"

# Load environment variables from .env if it exists
if [ -f ".env" ]; then
    echo -e "${GREEN}✅ Loading environment from .env file${NC}"
    export $(cat .env | grep -v '^#' | xargs)
else
    echo -e "${YELLOW}⚠️  No .env file found${NC}"
fi

# Map SUPABASE_API_KEY to SUPABASE_ANON_KEY if needed
if [ -n "$SUPABASE_API_KEY" ] && [ -z "$SUPABASE_ANON_KEY" ]; then
    export SUPABASE_ANON_KEY="$SUPABASE_API_KEY"
    echo -e "${GREEN}✅ Mapped SUPABASE_API_KEY to SUPABASE_ANON_KEY${NC}"
fi

# Set defaults
export PORT="${PORT:-8081}"
export LOG_LEVEL="${LOG_LEVEL:-debug}"
export PLAYWRIGHT_SERVICE_URL="${PLAYWRIGHT_SERVICE_URL:-https://playwright-service-production-b21a.up.railway.app}"

# Verify required variables
if [ -z "$SUPABASE_URL" ] || [ -z "$SUPABASE_ANON_KEY" ] || [ -z "$SUPABASE_SERVICE_ROLE_KEY" ]; then
    echo -e "${RED}❌ Missing required environment variables${NC}"
    echo "Required:"
    echo "  SUPABASE_URL"
    echo "  SUPABASE_ANON_KEY (or SUPABASE_API_KEY)"
    echo "  SUPABASE_SERVICE_ROLE_KEY"
    exit 1
fi

echo -e "${GREEN}✅ Environment configured${NC}"
echo -e "  PORT: ${PORT}"
echo -e "  LOG_LEVEL: ${LOG_LEVEL}"
echo -e "  SUPABASE_URL: ${SUPABASE_URL}"
if [ -n "$PLAYWRIGHT_SERVICE_URL" ]; then
    echo -e "  PLAYWRIGHT_SERVICE_URL: ${PLAYWRIGHT_SERVICE_URL}"
fi
echo ""

# Change to service directory
cd "$(dirname "$0")/../services/classification-service"

echo -e "${BLUE}Starting service...${NC}"
echo -e "${YELLOW}Watch for these startup logs:${NC}"
echo -e "  🚀 Starting Classification Service"
echo -e "  ✅ Phase 1 enhanced website scraper initialized"
echo -e "  🚀 Classification Service listening on :${PORT}"
echo ""
echo -e "${BLUE}=== Service Logs ===${NC}\n"

# Start the service
go run cmd/main.go

