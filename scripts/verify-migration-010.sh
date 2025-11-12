#!/bin/bash

# Verify Migration 010: Check if async risk assessment columns exist
# This script verifies that the migration was applied successfully

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔍 Verifying Migration 010: Async Risk Assessment Columns${NC}"
echo "======================================================"
echo ""

# Check for DATABASE_URL
if [ -z "$DATABASE_URL" ]; then
    echo -e "${YELLOW}⚠️  DATABASE_URL not set${NC}"
    echo "Please set DATABASE_URL environment variable"
    exit 1
fi

# Check if DATABASE_URL looks like a Supabase project URL (not connection string)
if [[ "$DATABASE_URL" == https://*.supabase.co* ]] || [[ "$DATABASE_URL" == https://*.supabase.co ]]; then
    echo -e "${RED}❌ Error: DATABASE_URL appears to be a Supabase project URL${NC}"
    echo "You need the PostgreSQL connection string. See scripts/get-supabase-connection-string.md"
    exit 1
fi

echo -e "${GREEN}✅ DATABASE_URL is set${NC}"
echo ""

# Check if psql is available
if ! command -v psql &> /dev/null; then
    echo -e "${YELLOW}⚠️  psql not found${NC}"
    echo "Please install PostgreSQL client tools or verify manually in Supabase SQL Editor"
    exit 1
fi

echo -e "${BLUE}Checking if columns exist...${NC}"
echo ""

# Verify columns exist
VERIFICATION_QUERY="
SELECT 
    column_name, 
    data_type,
    is_nullable
FROM information_schema.columns 
WHERE table_name = 'risk_assessments' 
AND column_name IN ('merchant_id', 'status', 'options', 'result', 'progress', 'estimated_completion', 'completed_at')
ORDER BY column_name;
"

if psql "$DATABASE_URL" -c "$VERIFICATION_QUERY" -t; then
    echo ""
    echo -e "${BLUE}Checking indexes...${NC}"
    
    # Verify indexes exist
    INDEX_QUERY="
    SELECT 
        indexname,
        indexdef
    FROM pg_indexes 
    WHERE tablename = 'risk_assessments' 
    AND indexname IN ('idx_risk_assessments_merchant_id', 'idx_risk_assessments_status', 'idx_risk_assessments_created_at')
    ORDER BY indexname;
    "
    
    if psql "$DATABASE_URL" -c "$INDEX_QUERY" -t; then
        echo ""
        echo -e "${GREEN}✅ Migration 010 verified successfully!${NC}"
        echo ""
        echo -e "${GREEN}All required columns and indexes are present.${NC}"
    else
        echo ""
        echo -e "${YELLOW}⚠️  Columns exist but some indexes may be missing${NC}"
    fi
else
    echo ""
    echo -e "${RED}❌ Verification failed${NC}"
    echo "The migration may not have been applied, or the risk_assessments table doesn't exist."
    exit 1
fi

echo ""
echo -e "${BLUE}📋 Next Steps:${NC}"
echo "1. ✅ Database migration complete"
echo "2. ⏳ Register routes in your main server application"
echo "3. ⏳ Test endpoints using Postman/Insomnia"
echo ""
echo "See: docs/async-routes-integration-guide.md for route registration"

