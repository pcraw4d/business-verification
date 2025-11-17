#!/bin/bash

# Database Schema Verification Script
# Verifies that the risk_thresholds table exists with correct structure

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if DATABASE_URL is set
if [ -z "$DATABASE_URL" ]; then
    echo -e "${RED}❌ ERROR: DATABASE_URL environment variable is not set${NC}"
    echo "Please set DATABASE_URL before running this script:"
    echo "  export DATABASE_URL='postgres://user:pass@host:port/dbname'"
    exit 1
fi

echo -e "${GREEN}🔍 Verifying database schema...${NC}"
echo ""

# Extract connection details from DATABASE_URL
# Format: postgres://user:password@host:port/database
DB_URL="$DATABASE_URL"

# Check if psql is available
if ! command -v psql &> /dev/null; then
    echo -e "${YELLOW}⚠️  WARNING: psql not found. Using alternative method...${NC}"
    USE_PSQL=false
else
    USE_PSQL=true
fi

# Function to verify table exists
verify_table_exists() {
    if [ "$USE_PSQL" = true ]; then
        TABLE_EXISTS=$(psql "$DB_URL" -tAc "SELECT EXISTS (
            SELECT FROM information_schema.tables 
            WHERE table_schema = 'public' 
            AND table_name = 'risk_thresholds'
        );" 2>/dev/null || echo "false")
    else
        # Alternative: Use Go script or API endpoint
        echo -e "${YELLOW}⚠️  Using API endpoint to verify table...${NC}"
        TABLE_EXISTS="unknown"
    fi
    
    if [ "$TABLE_EXISTS" = "t" ] || [ "$TABLE_EXISTS" = "true" ]; then
        echo -e "${GREEN}✅ Table 'risk_thresholds' exists${NC}"
        return 0
    else
        echo -e "${RED}❌ Table 'risk_thresholds' does not exist${NC}"
        return 1
    fi
}

# Function to verify table structure
verify_table_structure() {
    echo -e "${GREEN}🔍 Verifying table structure...${NC}"
    
    if [ "$USE_PSQL" = true ]; then
        # Check required columns
        REQUIRED_COLUMNS=("id" "name" "category" "risk_levels" "is_default" "is_active" "created_at" "updated_at")
        
        for COLUMN in "${REQUIRED_COLUMNS[@]}"; do
            COLUMN_EXISTS=$(psql "$DB_URL" -tAc "SELECT EXISTS (
                SELECT FROM information_schema.columns 
                WHERE table_schema = 'public' 
                AND table_name = 'risk_thresholds' 
                AND column_name = '$COLUMN'
            );" 2>/dev/null || echo "false")
            
            if [ "$COLUMN_EXISTS" = "t" ] || [ "$COLUMN_EXISTS" = "true" ]; then
                echo -e "  ${GREEN}✅ Column '$COLUMN' exists${NC}"
            else
                echo -e "  ${RED}❌ Column '$COLUMN' missing${NC}"
                return 1
            fi
        done
        
        # Check indexes
        echo -e "${GREEN}🔍 Verifying indexes...${NC}"
        INDEX_COUNT=$(psql "$DB_URL" -tAc "SELECT COUNT(*) FROM pg_indexes WHERE tablename = 'risk_thresholds';" 2>/dev/null || echo "0")
        echo -e "  ${GREEN}✅ Found $INDEX_COUNT indexes${NC}"
        
        return 0
    else
        echo -e "${YELLOW}⚠️  Cannot verify structure without psql. Please install PostgreSQL client tools.${NC}"
        return 1
    fi
}

# Function to check if migration needs to be applied
check_migration_status() {
    echo -e "${GREEN}🔍 Checking migration status...${NC}"
    
    if [ "$USE_PSQL" = true ]; then
        # Try to query the table
        QUERY_RESULT=$(psql "$DB_URL" -tAc "SELECT COUNT(*) FROM risk_thresholds;" 2>/dev/null || echo "ERROR")
        
        if [ "$QUERY_RESULT" != "ERROR" ]; then
            echo -e "  ${GREEN}✅ Table is accessible and queryable${NC}"
            echo -e "  ${GREEN}✅ Current threshold count: $QUERY_RESULT${NC}"
            return 0
        else
            echo -e "  ${RED}❌ Table exists but is not accessible${NC}"
            return 1
        fi
    else
        echo -e "${YELLOW}⚠️  Cannot check migration status without psql${NC}"
        return 1
    fi
}

# Main verification flow
echo -e "${GREEN}════════════════════════════════════════${NC}"
echo -e "${GREEN}  Database Schema Verification${NC}"
echo -e "${GREEN}════════════════════════════════════════${NC}"
echo ""

# Step 1: Verify table exists
if verify_table_exists; then
    # Step 2: Verify table structure
    if verify_table_structure; then
        # Step 3: Check migration status
        if check_migration_status; then
            echo ""
            echo -e "${GREEN}════════════════════════════════════════${NC}"
            echo -e "${GREEN}✅ All schema checks passed!${NC}"
            echo -e "${GREEN}════════════════════════════════════════${NC}"
            exit 0
        else
            echo ""
            echo -e "${YELLOW}⚠️  Table exists but may need migration${NC}"
            echo "Run the migration script:"
            echo "  psql \$DATABASE_URL -f internal/database/migrations/012_create_risk_thresholds_table.sql"
            exit 1
        fi
    else
        echo ""
        echo -e "${RED}❌ Table structure verification failed${NC}"
        echo "Please run the migration script:"
        echo "  psql \$DATABASE_URL -f internal/database/migrations/012_create_risk_thresholds_table.sql"
        exit 1
    fi
else
    echo ""
    echo -e "${RED}❌ Table does not exist${NC}"
    echo "Please run the migration script:"
    echo "  psql \$DATABASE_URL -f internal/database/migrations/012_create_risk_thresholds_table.sql"
    exit 1
fi

