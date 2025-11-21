#!/bin/bash

# Check Merchant Service Environment Variables
# This script verifies that all required environment variables are set
# Usage: ./scripts/check-merchant-service-env.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔍 Checking Merchant Service Environment Variables...${NC}"
echo ""

# Required variables
REQUIRED_VARS=(
    "SUPABASE_URL"
    "SUPABASE_ANON_KEY"
    "SUPABASE_SERVICE_ROLE_KEY"
)

# Optional variables
OPTIONAL_VARS=(
    "SUPABASE_JWT_SECRET"
    "ENVIRONMENT"
    "PORT"
)

# Check if railway.env exists
if [ ! -f "railway.env" ]; then
    echo -e "${RED}❌ railway.env file not found${NC}"
    echo -e "${YELLOW}   Please create railway.env with required environment variables${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Found railway.env${NC}"
echo ""

# Source railway.env
source railway.env

# Check required variables
MISSING_VARS=()
ALL_SET=true

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Required Variables:${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

for var in "${REQUIRED_VARS[@]}"; do
    value=$(eval echo \$$var)
    if [ -z "$value" ] || [ "$value" = "your_supabase_anon_key_here" ] || [ "$value" = "your_supabase_service_role_key_here" ] || [ "$value" = "your_supabase_jwt_secret_here" ]; then
        echo -e "${RED}✗ ${var}: NOT SET or PLACEHOLDER${NC}"
        MISSING_VARS+=("$var")
        ALL_SET=false
    else
        # Show first 20 characters for security
        display_value="${value:0:20}..."
        if [ ${#value} -le 20 ]; then
            display_value="$value"
        fi
        echo -e "${GREEN}✓ ${var}: ${display_value}${NC}"
    fi
done

echo ""

# Check optional variables
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Optional Variables:${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

for var in "${OPTIONAL_VARS[@]}"; do
    value=$(eval echo \$$var)
    if [ -z "$value" ]; then
        echo -e "${YELLOW}⚠ ${var}: NOT SET (using default)${NC}"
    else
        if [[ "$var" == *"KEY"* ]] || [[ "$var" == *"SECRET"* ]]; then
            display_value="${value:0:20}..."
            if [ ${#value} -le 20 ]; then
                display_value="$value"
            fi
            echo -e "${GREEN}✓ ${var}: ${display_value}${NC}"
        else
            echo -e "${GREEN}✓ ${var}: ${value}${NC}"
        fi
    fi
done

echo ""

# Summary
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
if [ "$ALL_SET" = true ]; then
    echo -e "${GREEN}✅ All required environment variables are set!${NC}"
    echo ""
    echo -e "${GREEN}Merchant service should be able to start successfully.${NC}"
    exit 0
else
    echo -e "${RED}❌ Missing required environment variables:${NC}"
    for var in "${MISSING_VARS[@]}"; do
        echo -e "${RED}   - ${var}${NC}"
    done
    echo ""
    echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}How to Fix:${NC}"
    echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo "1. Get your Supabase credentials from:"
    echo "   https://supabase.com/dashboard/project/qpqhuqqmkjxsltzshfam/settings/api"
    echo ""
    echo "2. Edit railway.env and replace placeholder values:"
    echo "   SUPABASE_ANON_KEY=your_actual_anon_key"
    echo "   SUPABASE_SERVICE_ROLE_KEY=your_actual_service_role_key"
    echo "   SUPABASE_JWT_SECRET=your_actual_jwt_secret"
    echo ""
    echo "3. Or set them as environment variables:"
    echo "   export SUPABASE_ANON_KEY='your_key'"
    echo "   export SUPABASE_SERVICE_ROLE_KEY='your_key'"
    echo ""
    exit 1
fi

