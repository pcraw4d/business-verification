#!/bin/bash

# Seed Test Merchants for Phase 2 Testing
# This script seeds the Supabase database with test merchants for various scenarios
# Usage: ./scripts/seed-test-merchants.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

ENV_FILE="railway.env"
SQL_FILE="test/sql/test_merchant_data.sql"

echo -e "${BLUE}🌱 Seeding Test Merchants...${NC}"
echo ""

# Check if SQL file exists
if [ ! -f "${SQL_FILE}" ]; then
    echo -e "${RED}❌ SQL file not found: ${SQL_FILE}${NC}"
    exit 1
fi

# Source environment variables
if [ -f "${ENV_FILE}" ]; then
    source "${ENV_FILE}"
    echo -e "${GREEN}✅ Found ${ENV_FILE}${NC}"
else
    echo -e "${RED}Error: ${ENV_FILE} file not found${NC}"
    echo "Please create ${ENV_FILE} with required environment variables"
    exit 1
fi

# Check for required environment variables
if [ -z "${SUPABASE_URL}" ] || [ -z "${SUPABASE_SERVICE_ROLE_KEY}" ]; then
    echo -e "${RED}❌ Missing required environment variables: SUPABASE_URL or SUPABASE_SERVICE_ROLE_KEY${NC}"
    exit 1
fi

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Database Information${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo "Supabase URL: ${SUPABASE_URL}"
echo "SQL File: ${SQL_FILE}"
echo ""

# Extract database connection details from Supabase URL
# Supabase URL format: https://<project-ref>.supabase.co
# We need to connect to: postgresql://postgres:[password]@db.<project-ref>.supabase.co:5432/postgres

# For now, we'll use psql if available, or provide instructions
if command -v psql &> /dev/null; then
    echo -e "${YELLOW}⚠️  Direct psql connection requires database password${NC}"
    echo -e "${YELLOW}   Please run the SQL file manually using Supabase Dashboard or psql${NC}"
    echo ""
    echo -e "${BLUE}Alternative: Use Supabase MCP or Dashboard${NC}"
    echo ""
    echo -e "${GREEN}SQL file location: ${SQL_FILE}${NC}"
    echo ""
    echo -e "${BLUE}To execute manually:${NC}"
    echo "1. Open Supabase Dashboard: https://supabase.com/dashboard"
    echo "2. Go to SQL Editor"
    echo "3. Copy contents of ${SQL_FILE}"
    echo "4. Execute the SQL"
    echo ""
    echo -e "${BLUE}Or use Supabase CLI:${NC}"
    echo "supabase db execute --file ${SQL_FILE}"
    exit 0
else
    echo -e "${YELLOW}⚠️  psql not found. Please install PostgreSQL client or use Supabase Dashboard${NC}"
    echo ""
    echo -e "${GREEN}SQL file location: ${SQL_FILE}${NC}"
    echo ""
    echo -e "${BLUE}To execute manually:${NC}"
    echo "1. Open Supabase Dashboard: https://supabase.com/dashboard"
    echo "2. Go to SQL Editor"
    echo "3. Copy contents of ${SQL_FILE}"
    echo "4. Execute the SQL"
    exit 0
fi

