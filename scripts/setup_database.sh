#!/bin/bash

# Database Setup Script
# Guides user through database configuration and migration

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${BLUE}════════════════════════════════════════${NC}"
echo -e "${BLUE}  Database Configuration Setup${NC}"
echo -e "${BLUE}════════════════════════════════════════${NC}"
echo ""

# Check if DATABASE_URL is already set
if [ -n "$DATABASE_URL" ]; then
    echo -e "${GREEN}✅ DATABASE_URL is already set${NC}"
    echo -e "   Current value: ${CYAN}${DATABASE_URL:0:50}...${NC}"
    echo ""
    read -p "Do you want to use this existing DATABASE_URL? (y/n): " use_existing
    if [ "$use_existing" != "y" ] && [ "$use_existing" != "Y" ]; then
        unset DATABASE_URL
    fi
fi

# If DATABASE_URL is not set, help user configure it
if [ -z "$DATABASE_URL" ]; then
    echo -e "${YELLOW}📋 Database Configuration Required${NC}"
    echo ""
    echo "You need to provide a PostgreSQL connection string."
    echo ""
    echo "Options:"
    echo "  1. Supabase Database (recommended)"
    echo "  2. Custom PostgreSQL Database"
    echo "  3. Use existing environment variable"
    echo ""
    read -p "Choose option (1-3): " db_option
    
    case $db_option in
        1)
            echo ""
            echo -e "${CYAN}Supabase Database Setup${NC}"
            echo ""
            echo "To get your Supabase database URL:"
            echo "  1. Go to Supabase Dashboard: https://app.supabase.com"
            echo "  2. Select your project"
            echo "  3. Go to Settings → Database"
            echo "  4. Copy the 'Connection string' under 'Connection pooling'"
            echo ""
            echo "Format: postgresql://postgres:[YOUR-PASSWORD]@db.[PROJECT-REF].supabase.co:5432/postgres"
            echo ""
            read -p "Enter your Supabase DATABASE_URL: " db_url
            export DATABASE_URL="$db_url"
            ;;
        2)
            echo ""
            echo -e "${CYAN}Custom PostgreSQL Database Setup${NC}"
            echo ""
            read -p "Enter PostgreSQL host: " db_host
            read -p "Enter PostgreSQL port (default 5432): " db_port
            db_port=${db_port:-5432}
            read -p "Enter database name: " db_name
            read -p "Enter username: " db_user
            read -s -p "Enter password: " db_pass
            echo ""
            read -p "Use SSL? (y/n, default y): " use_ssl
            if [ "$use_ssl" != "n" ] && [ "$use_ssl" != "N" ]; then
                ssl_mode="sslmode=require"
            else
                ssl_mode="sslmode=disable"
            fi
            export DATABASE_URL="postgres://${db_user}:${db_pass}@${db_host}:${db_port}/${db_name}?${ssl_mode}"
            ;;
        3)
            echo ""
            echo "Please set DATABASE_URL environment variable:"
            echo "  export DATABASE_URL='postgres://user:pass@host:port/dbname'"
            echo ""
            echo "Then run this script again."
            exit 0
            ;;
        *)
            echo -e "${RED}Invalid option${NC}"
            exit 1
            ;;
    esac
fi

echo ""
echo -e "${GREEN}🔍 Testing database connection...${NC}"

# Test connection
if [ -f "./scripts/test_database_connection.sh" ]; then
    if ./scripts/test_database_connection.sh; then
        echo -e "${GREEN}✅ Database connection successful!${NC}"
    else
        echo -e "${RED}❌ Database connection failed${NC}"
        echo "Please check your DATABASE_URL and try again."
        exit 1
    fi
else
    # Basic connection test with psql if available
    if command -v psql &> /dev/null; then
        if psql "$DATABASE_URL" -c "SELECT version();" > /dev/null 2>&1; then
            echo -e "${GREEN}✅ Database connection successful!${NC}"
        else
            echo -e "${RED}❌ Database connection failed${NC}"
            exit 1
        fi
    else
        echo -e "${YELLOW}⚠️  Cannot test connection (psql not available)${NC}"
        echo "Please verify your DATABASE_URL is correct."
    fi
fi

echo ""
echo -e "${GREEN}📋 Checking database schema...${NC}"

# Check if table exists
if command -v psql &> /dev/null; then
    TABLE_EXISTS=$(psql "$DATABASE_URL" -tAc "SELECT EXISTS (
        SELECT FROM information_schema.tables 
        WHERE table_schema = 'public' 
        AND table_name = 'risk_thresholds'
    );" 2>/dev/null || echo "false")
    
    if [ "$TABLE_EXISTS" = "t" ] || [ "$TABLE_EXISTS" = "true" ]; then
        echo -e "${GREEN}✅ risk_thresholds table already exists${NC}"
        echo ""
        read -p "Do you want to run the migration anyway? (y/n): " run_migration
        if [ "$run_migration" != "y" ] && [ "$run_migration" != "Y" ]; then
            echo -e "${GREEN}✅ Skipping migration${NC}"
            MIGRATION_NEEDED=false
        else
            MIGRATION_NEEDED=true
        fi
    else
        echo -e "${YELLOW}⚠️  risk_thresholds table does not exist${NC}"
        MIGRATION_NEEDED=true
    fi
else
    echo -e "${YELLOW}⚠️  Cannot check schema (psql not available)${NC}"
    MIGRATION_NEEDED=true
fi

# Run migration if needed
if [ "$MIGRATION_NEEDED" = true ]; then
    echo ""
    echo -e "${GREEN}🚀 Running database migration...${NC}"
    
    MIGRATION_FILE="internal/database/migrations/012_create_risk_thresholds_table.sql"
    
    if [ ! -f "$MIGRATION_FILE" ]; then
        echo -e "${RED}❌ Migration file not found: $MIGRATION_FILE${NC}"
        exit 1
    fi
    
    if command -v psql &> /dev/null; then
        if psql "$DATABASE_URL" -f "$MIGRATION_FILE"; then
            echo -e "${GREEN}✅ Migration completed successfully!${NC}"
        else
            echo -e "${RED}❌ Migration failed${NC}"
            exit 1
        fi
    else
        echo -e "${YELLOW}⚠️  psql not available. Please run migration manually:${NC}"
        echo "  psql \$DATABASE_URL -f $MIGRATION_FILE"
        read -p "Press Enter after running the migration..."
    fi
fi

# Verify schema
echo ""
echo -e "${GREEN}🔍 Verifying database schema...${NC}"
if [ -f "./scripts/verify_database_schema.sh" ]; then
    if ./scripts/verify_database_schema.sh; then
        echo -e "${GREEN}✅ Schema verification passed!${NC}"
    else
        echo -e "${YELLOW}⚠️  Schema verification had warnings${NC}"
    fi
fi

# Save to .env file
echo ""
echo -e "${GREEN}💾 Saving configuration...${NC}"
read -p "Save DATABASE_URL to .env file? (y/n): " save_env

if [ "$save_env" = "y" ] || [ "$save_env" = "Y" ]; then
    # Check if .env exists
    if [ -f ".env" ]; then
        # Check if DATABASE_URL already exists
        if grep -q "^DATABASE_URL=" .env; then
            read -p "DATABASE_URL already exists in .env. Overwrite? (y/n): " overwrite
            if [ "$overwrite" = "y" ] || [ "$overwrite" = "Y" ]; then
                # Remove old DATABASE_URL line
                sed -i.bak '/^DATABASE_URL=/d' .env
                echo "DATABASE_URL=$DATABASE_URL" >> .env
                echo -e "${GREEN}✅ Updated .env file${NC}"
            fi
        else
            echo "DATABASE_URL=$DATABASE_URL" >> .env
            echo -e "${GREEN}✅ Added DATABASE_URL to .env file${NC}"
        fi
    else
        echo "DATABASE_URL=$DATABASE_URL" > .env
        echo -e "${GREEN}✅ Created .env file with DATABASE_URL${NC}"
    fi
fi

# Summary
echo ""
echo -e "${BLUE}════════════════════════════════════════${NC}"
echo -e "${GREEN}✅ Database setup complete!${NC}"
echo -e "${BLUE}════════════════════════════════════════${NC}"
echo ""
echo -e "${CYAN}Next steps:${NC}"
echo "  1. Test the connection: ./scripts/test_database_connection.sh"
echo "  2. Verify schema: ./scripts/verify_database_schema.sh"
echo "  3. Start server: go run cmd/railway-server/main.go"
echo "  4. Test endpoints: ./test/restoration_tests.sh"
echo ""
echo -e "${YELLOW}Note:${NC} Make sure to set DATABASE_URL in your production environment!"
echo ""

