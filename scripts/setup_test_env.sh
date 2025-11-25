#!/bin/bash

# Script to source Supabase credentials from existing .env files
# This script checks multiple config locations and sources the first available

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "🔧 Setting up test environment for Supabase integration tests"
echo "=============================================================="
echo ""

# Function to source env file and export variables
source_env_file() {
    local env_file=$1
    if [ -f "$env_file" ]; then
        echo -e "${GREEN}✓${NC} Found config file: $env_file"
        
        # Source the file
        set -a
        source "$env_file"
        set +a
        
        # Map SUPABASE_API_KEY to SUPABASE_ANON_KEY if needed
        if [ -n "$SUPABASE_API_KEY" ] && [ -z "$SUPABASE_ANON_KEY" ]; then
            export SUPABASE_ANON_KEY="$SUPABASE_API_KEY"
            echo -e "${YELLOW}⚠${NC}  Mapped SUPABASE_API_KEY to SUPABASE_ANON_KEY"
        fi
        
        return 0
    fi
    return 1
}

# Try to source from various config locations (in order of preference)
CONFIG_FOUND=false

# 1. Check for test.env in configs directory
if source_env_file "configs/test.env"; then
    CONFIG_FOUND=true
# 2. Check for development.env in configs directory
elif source_env_file "configs/development.env"; then
    CONFIG_FOUND=true
# 3. Check for .env in project root
elif source_env_file ".env"; then
    CONFIG_FOUND=true
# 4. Check for railway.env
elif source_env_file "railway.env"; then
    CONFIG_FOUND=true
# 5. Check for production.env in configs
elif source_env_file "configs/production.env"; then
    CONFIG_FOUND=true
fi

if [ "$CONFIG_FOUND" = false ]; then
    echo -e "${RED}✗${NC} No config file found in expected locations:"
    echo "  - configs/test.env"
    echo "  - configs/development.env"
    echo "  - .env"
    echo "  - railway.env"
    echo "  - configs/production.env"
    echo ""
    echo "Please create one of these files with your Supabase credentials:"
    echo ""
    echo "  SUPABASE_URL=https://your-project.supabase.co"
    echo "  SUPABASE_ANON_KEY=your_anon_key"
    echo "  SUPABASE_SERVICE_ROLE_KEY=your_service_role_key"
    echo ""
    exit 1
fi

# Validate required variables
MISSING_VARS=()

if [ -z "$SUPABASE_URL" ]; then
    MISSING_VARS+=("SUPABASE_URL")
fi

if [ -z "$SUPABASE_ANON_KEY" ]; then
    MISSING_VARS+=("SUPABASE_ANON_KEY")
fi

if [ -n "${MISSING_VARS[*]}" ]; then
    echo -e "${RED}✗${NC} Missing required environment variables:"
    for var in "${MISSING_VARS[@]}"; do
        echo "  - $var"
    done
    echo ""
    echo "Please add these to your config file."
    exit 1
fi

# Check if values are placeholders
if [[ "$SUPABASE_URL" == *"your-project"* ]] || [[ "$SUPABASE_URL" == *"test-project"* ]]; then
    echo -e "${YELLOW}⚠${NC}  SUPABASE_URL appears to be a placeholder value"
    echo "  Current value: $SUPABASE_URL"
    echo "  Please update with your actual Supabase project URL"
fi

if [[ "$SUPABASE_ANON_KEY" == *"your"* ]] || [[ "$SUPABASE_ANON_KEY" == *"test"* ]]; then
    echo -e "${YELLOW}⚠${NC}  SUPABASE_ANON_KEY appears to be a placeholder value"
    echo "  Please update with your actual Supabase anon key"
fi

# Display configuration (mask sensitive values)
echo ""
echo "✅ Configuration loaded successfully!"
echo ""
echo "Supabase Configuration:"
echo "  URL: $SUPABASE_URL"
if [ -n "$SUPABASE_ANON_KEY" ]; then
    ANON_KEY_PREVIEW="${SUPABASE_ANON_KEY:0:20}..."
    echo "  Anon Key: $ANON_KEY_PREVIEW"
fi
if [ -n "$SUPABASE_SERVICE_ROLE_KEY" ]; then
    SERVICE_KEY_PREVIEW="${SUPABASE_SERVICE_ROLE_KEY:0:20}..."
    echo "  Service Role Key: $SERVICE_KEY_PREVIEW"
else
    echo -e "${YELLOW}⚠${NC}  SUPABASE_SERVICE_ROLE_KEY not set (optional for some tests)"
fi
echo ""

# Export variables for use in child processes
export SUPABASE_URL
export SUPABASE_ANON_KEY
export SUPABASE_SERVICE_ROLE_KEY

echo "Environment variables exported. You can now run:"
echo "  ./scripts/run_hybrid_tests.sh"
echo "  or"
echo "  go test ./internal/classification/testutil -v -run TestHybridCodeGeneration_WithRealRepository"
echo ""

