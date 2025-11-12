#!/bin/bash

# Run migration 010: Add async risk assessment columns
# This script runs the database migration for async risk assessment support

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

MIGRATION_FILE="internal/database/migrations/010_add_async_risk_assessment_columns.sql"

echo -e "${BLUE}🚀 Running Migration 010: Async Risk Assessment Columns${NC}"
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
    echo "  Format: postgresql://postgres.[project-ref]:[password]@aws-0-[region].pooler.supabase.com:6543/postgres"
    echo "  Or direct: postgresql://postgres:[password]@db.[project-ref].supabase.co:5432/postgres"
    echo ""
    echo "Get it from: Supabase Dashboard > Project Settings > Database > Connection String"
    echo ""
    echo "Example: export DATABASE_URL='postgresql://postgres:password@db.xxxxx.supabase.co:5432/postgres'"
    exit 1
fi

# Check if DATABASE_URL looks like a Supabase project URL (not connection string)
if [[ "$DATABASE_URL" == https://*.supabase.co* ]] || [[ "$DATABASE_URL" == https://*.supabase.co ]]; then
    echo -e "${RED}❌ Error: DATABASE_URL appears to be a Supabase project URL, not a PostgreSQL connection string${NC}"
    echo ""
    echo "You need the PostgreSQL connection string, not the project URL."
    echo ""
    echo "To get the connection string:"
    echo "1. Go to Supabase Dashboard: https://supabase.com/dashboard"
    echo "2. Select your project"
    echo "3. Go to: Project Settings > Database"
    echo "4. Copy the 'Connection string' (URI format)"
    echo ""
    echo "It should look like:"
    echo "  postgresql://postgres:[YOUR-PASSWORD]@db.qpqhuqqmkjxsltzshfam.supabase.co:5432/postgres"
    echo ""
    echo "Or use the connection pooler:"
    echo "  postgresql://postgres.qpqhuqqmkjxsltzshfam:[YOUR-PASSWORD]@aws-0-us-east-1.pooler.supabase.com:6543/postgres"
    echo ""
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
        
        # Verify columns were added
        psql "$DATABASE_URL" -c "
            SELECT column_name, data_type 
            FROM information_schema.columns 
            WHERE table_name = 'risk_assessments' 
            AND column_name IN ('merchant_id', 'status', 'options', 'result', 'progress', 'estimated_completion', 'completed_at')
            ORDER BY column_name;
        " || echo -e "${YELLOW}⚠️  Could not verify columns (table might not exist yet)${NC}"
        
    else
        echo ""
        echo -e "${RED}❌ Migration failed${NC}"
        exit 1
    fi
else
    echo -e "${YELLOW}⚠️  psql not found${NC}"
    echo ""
    echo "Please run the migration manually:"
    echo "  psql \$DATABASE_URL -f $MIGRATION_FILE"
    echo ""
    echo "Or use the Go migration tool:"
    echo "  go run cmd/migrate/main.go"
    exit 1
fi

echo ""
echo -e "${GREEN}🎉 Migration 010 completed!${NC}"

