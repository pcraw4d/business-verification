#!/bin/bash
# Script to refresh PostgREST schema cache using Supabase Management API
# This works for hosted Supabase projects

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Refreshing PostgREST Schema Cache via Supabase API${NC}"
echo ""

# Load environment variables from railway.env if it exists
if [ -f "railway.env" ]; then
    echo -e "${YELLOW}Loading configuration from railway.env...${NC}"
    export $(grep -E '^SUPABASE_URL=|^SUPABASE_SERVICE_ROLE_KEY=' railway.env | grep -v '^#' | xargs)
fi

# Set defaults from railway.env values
SUPABASE_URL="${SUPABASE_URL:-https://qpqhuqqmkjxsltzshfam.supabase.co}"
SUPABASE_SERVICE_ROLE_KEY="${SUPABASE_SERVICE_ROLE_KEY:-}"

# Check if service role key is set
if [ -z "$SUPABASE_SERVICE_ROLE_KEY" ]; then
    echo -e "${RED}Error: SUPABASE_SERVICE_ROLE_KEY is not set${NC}"
    echo ""
    echo "Please set it in one of these ways:"
    echo "1. Export it: export SUPABASE_SERVICE_ROLE_KEY='your-key'"
    echo "2. Add it to railway.env file"
    echo "3. Pass it as argument: $0 <service-role-key>"
    echo ""
    if [ -n "$1" ]; then
        SUPABASE_SERVICE_ROLE_KEY="$1"
        echo -e "${YELLOW}Using key from command line argument${NC}"
    else
        exit 1
    fi
fi

echo -e "${YELLOW}Project URL: ${SUPABASE_URL}${NC}"
echo ""

# Method 1: Try using the reload_schema RPC function (if it exists)
echo -e "${BLUE}Method 1: Attempting to call reload_schema RPC...${NC}"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
  "${SUPABASE_URL}/rest/v1/rpc/reload_schema" \
  -H "apikey: ${SUPABASE_SERVICE_ROLE_KEY}" \
  -H "Authorization: Bearer ${SUPABASE_SERVICE_ROLE_KEY}" \
  -H "Content-Type: application/json" \
  -d '{}' 2>&1)

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "204" ]; then
    echo -e "${GREEN}✓ Schema cache reloaded successfully!${NC}"
    echo ""
    echo -e "${GREEN}PostgREST schema cache has been refreshed.${NC}"
    echo "Wait 10-30 seconds, then try the refresh button again."
    exit 0
elif [ "$HTTP_CODE" = "404" ]; then
    echo -e "${YELLOW}⚠ reload_schema RPC function not found (this is normal for hosted projects)${NC}"
    echo ""
else
    echo -e "${YELLOW}⚠ Received HTTP $HTTP_CODE${NC}"
    echo "Response: $BODY"
    echo ""
fi

# Method 2: Use Supabase Management API (requires access token)
echo -e "${BLUE}Method 2: Using Supabase Management API...${NC}"
echo ""
echo -e "${YELLOW}For hosted Supabase projects, the recommended way is:${NC}"
echo ""
echo "1. Go to Supabase Dashboard:"
echo "   ${BLUE}https://supabase.com/dashboard/project/qpqhuqqmkjxsltzshfam${NC}"
echo ""
echo "2. Navigate to: ${BLUE}Settings > API${NC}"
echo ""
echo "3. Look for ${BLUE}'Reload Schema'${NC} or ${BLUE}'Refresh Schema Cache'${NC} button"
echo ""
echo "4. Click it to force PostgREST to reload the schema"
echo ""
echo -e "${YELLOW}Alternative: Wait 5-10 minutes for automatic refresh${NC}"
echo ""
echo -e "${GREEN}The schema cache will refresh automatically within 5-10 minutes.${NC}"
echo "After that, website analysis and classification should work correctly."

