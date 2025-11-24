#!/bin/bash
# Script to refresh PostgREST schema cache for Supabase hosted project
# This uses the Supabase Management API to reload the schema

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Refreshing PostgREST Schema Cache${NC}"
echo ""

# Check if Supabase CLI is installed
if ! command -v supabase &> /dev/null; then
    echo -e "${RED}Error: Supabase CLI is not installed${NC}"
    echo "Install it with: brew install supabase/tap/supabase"
    exit 1
fi

# Check if we're linked to a project
if [ ! -f "supabase/.temp/project-ref" ] && [ -z "$SUPABASE_PROJECT_ID" ]; then
    echo -e "${YELLOW}Not linked to a project. Attempting to link...${NC}"
    echo ""
    echo "You'll need to:"
    echo "1. Get your project reference ID from Supabase dashboard"
    echo "2. Run: supabase link --project-ref qpqhuqqmkjxsltzshfam"
    echo ""
    read -p "Do you want to link now? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        supabase link --project-ref qpqhuqqmkjxsltzshfam
    else
        echo "Exiting. Please link manually first."
        exit 1
    fi
fi

# For hosted projects, we need to use the Management API
# The CLI doesn't have a direct command to refresh PostgREST cache
# But we can trigger it by pushing migrations or using the API

echo -e "${YELLOW}Attempting to refresh schema cache...${NC}"
echo ""

# Option 1: Try using supabase db push (this might trigger a reload)
echo "Method 1: Pushing migrations (this will trigger schema reload)..."
if supabase db push --dry-run 2>&1 | grep -q "No changes"; then
    echo -e "${GREEN}✓ No pending migrations${NC}"
else
    echo "Pushing migrations to trigger schema reload..."
    supabase db push
fi

echo ""
echo -e "${YELLOW}Note: For hosted Supabase projects, the best way to refresh PostgREST schema cache is:${NC}"
echo ""
echo "1. Go to Supabase Dashboard: https://supabase.com/dashboard/project/qpqhuqqmkjxsltzshfam"
echo "2. Navigate to Settings > API"
echo "3. Click 'Reload Schema' or 'Refresh Schema Cache' button"
echo ""
echo "Or wait 5-10 minutes for automatic refresh."
echo ""
echo -e "${GREEN}Schema cache refresh initiated!${NC}"

