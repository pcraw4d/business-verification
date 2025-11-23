#!/bin/bash

# Migration: Add industry column to risk_assessments table
# This fixes ERROR #5 - Missing industry column causing 500 errors on /api/v1/analytics/trends

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

MIGRATION_FILE="supabase-migrations/add_industry_column_to_risk_assessments.sql"

echo -e "${BLUE}🚀 Running Migration: Add Industry Column to risk_assessments${NC}"
echo "=================================================="
echo ""

# Check if migration file exists
if [ ! -f "$MIGRATION_FILE" ]; then
    echo -e "${RED}❌ Migration file not found: $MIGRATION_FILE${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Migration file found${NC}"

# Check for DATABASE_URL
if [ -z "$DATABASE_URL" ]; then
    echo -e "${YELLOW}⚠️  DATABASE_URL not set${NC}"
    echo "Please set DATABASE_URL environment variable"
    echo ""
    echo "For Supabase, use the PostgreSQL connection string:"
    echo "  Format: postgresql://postgres:[password]@db.[project-ref].supabase.co:5432/postgres"
    echo ""
    echo "Get it from: Supabase Dashboard > Project Settings > Database > Connection String"
    echo ""
    echo "Example: export DATABASE_URL='postgresql://postgres:password@db.qpqhuqqmkjxsltzshfam.supabase.co:5432/postgres'"
    exit 1
fi

echo -e "${GREEN}✅ DATABASE_URL is set${NC}"
echo ""

# Check if psql is available
if command -v psql &> /dev/null; then
    echo -e "${BLUE}Running migration using psql...${NC}"
    echo ""
    
    if psql "$DATABASE_URL" -f "$MIGRATION_FILE"; then
        echo ""
        echo -e "${GREEN}✅ Migration completed successfully!${NC}"
        echo ""
        echo -e "${BLUE}Verifying migration...${NC}"
        
        # Verify column was added
        psql "$DATABASE_URL" -c "
            SELECT column_name, data_type, character_maximum_length
            FROM information_schema.columns 
            WHERE table_name = 'risk_assessments' 
            AND column_name = 'industry';
        " || echo -e "${YELLOW}⚠️  Could not verify column${NC}"
        
    else
        echo ""
        echo -e "${RED}❌ Migration failed${NC}"
        exit 1
    fi
else
    echo -e "${YELLOW}⚠️  psql not found${NC}"
    echo ""
    echo "Please install PostgreSQL client tools or run the migration manually:"
    echo "  psql \$DATABASE_URL -f $MIGRATION_FILE"
    exit 1
fi

echo ""
echo -e "${GREEN}🎉 Migration completed! Industry column added to risk_assessments table${NC}"

