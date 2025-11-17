#!/bin/bash

# Configuration Validation Script
# Validates all required environment variables and configuration

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

ERRORS=0
WARNINGS=0

echo -e "${BLUE}════════════════════════════════════════${NC}"
echo -e "${BLUE}  Configuration Validation${NC}"
echo -e "${BLUE}════════════════════════════════════════${NC}"
echo ""

# Required variables
REQUIRED_VARS=(
    "SUPABASE_URL"
    "SUPABASE_ANON_KEY"
)

# Optional but recommended variables
OPTIONAL_VARS=(
    "DATABASE_URL"
    "REDIS_URL"
    "PORT"
    "SERVICE_NAME"
)

# Function to validate URL format
validate_url() {
    local url=$1
    if [[ $url =~ ^https?:// ]] || [[ $url =~ ^postgres:// ]] || [[ $url =~ ^redis:// ]]; then
        return 0
    else
        return 1
    fi
}

# Function to check required variable
check_required() {
    local var_name=$1
    if [ -z "${!var_name}" ]; then
        echo -e "  ${RED}❌ $var_name: NOT SET (required)${NC}"
        ((ERRORS++))
        return 1
    else
        if validate_url "${!var_name}"; then
            echo -e "  ${GREEN}✅ $var_name: SET (valid format)${NC}"
        else
            echo -e "  ${GREEN}✅ $var_name: SET${NC}"
        fi
        return 0
    fi
}

# Function to check optional variable
check_optional() {
    local var_name=$1
    if [ -z "${!var_name}" ]; then
        echo -e "  ${YELLOW}⚠️  $var_name: NOT SET (optional)${NC}"
        ((WARNINGS++))
        return 1
    else
        if validate_url "${!var_name}"; then
            echo -e "  ${GREEN}✅ $var_name: SET (valid format)${NC}"
        else
            echo -e "  ${GREEN}✅ $var_name: SET${NC}"
        fi
        return 0
    fi
}

# Check required variables
echo -e "${GREEN}Required Variables:${NC}"
for var in "${REQUIRED_VARS[@]}"; do
    check_required "$var"
done

echo ""
echo -e "${GREEN}Optional Variables:${NC}"
for var in "${OPTIONAL_VARS[@]}"; do
    check_optional "$var"
done

# Special validations
echo ""
echo -e "${GREEN}Special Validations:${NC}"

# Validate PORT if set
if [ -n "$PORT" ]; then
    if [[ "$PORT" =~ ^[0-9]+$ ]] && [ "$PORT" -ge 1 ] && [ "$PORT" -le 65535 ]; then
        echo -e "  ${GREEN}✅ PORT: Valid port number ($PORT)${NC}"
    else
        echo -e "  ${RED}❌ PORT: Invalid port number ($PORT)${NC}"
        ((ERRORS++))
    fi
else
    echo -e "  ${YELLOW}⚠️  PORT: Using default (8080)${NC}"
fi

# Validate DATABASE_URL format if set
if [ -n "$DATABASE_URL" ]; then
    if [[ "$DATABASE_URL" =~ ^postgres:// ]] || [[ "$DATABASE_URL" =~ ^postgresql:// ]]; then
        echo -e "  ${GREEN}✅ DATABASE_URL: Valid PostgreSQL connection string${NC}"
    else
        echo -e "  ${YELLOW}⚠️  DATABASE_URL: May not be a valid PostgreSQL connection string${NC}"
        ((WARNINGS++))
    fi
fi

# Validate REDIS_URL format if set
if [ -n "$REDIS_URL" ]; then
    if [[ "$REDIS_URL" =~ ^redis:// ]] || [[ "$REDIS_URL" =~ ^rediss:// ]]; then
        echo -e "  ${GREEN}✅ REDIS_URL: Valid Redis connection string${NC}"
    else
        echo -e "  ${YELLOW}⚠️  REDIS_URL: May not be a valid Redis connection string${NC}"
        ((WARNINGS++))
    fi
fi

# Summary
echo ""
echo -e "${BLUE}════════════════════════════════════════${NC}"
if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo -e "${GREEN}✅ All validations passed!${NC}"
    echo -e "${BLUE}════════════════════════════════════════${NC}"
    exit 0
elif [ $ERRORS -eq 0 ]; then
    echo -e "${GREEN}✅ Required validations passed${NC}"
    echo -e "${YELLOW}⚠️  $WARNINGS warning(s) (optional variables)${NC}"
    echo -e "${BLUE}════════════════════════════════════════${NC}"
    exit 0
else
    echo -e "${RED}❌ Validation failed: $ERRORS error(s), $WARNINGS warning(s)${NC}"
    echo -e "${BLUE}════════════════════════════════════════${NC}"
    exit 1
fi

