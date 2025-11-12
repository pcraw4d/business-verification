#!/bin/bash

# Script to update DATABASE_URL in railway.env
# This script helps you update the DATABASE_URL without manually editing the file

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

RAILWAY_ENV_FILE="railway.env"

echo -e "${BLUE}🔧 Update DATABASE_URL in railway.env${NC}"
echo "=========================================="
echo ""

# Check if railway.env exists
if [ ! -f "$RAILWAY_ENV_FILE" ]; then
    echo -e "${RED}❌ Error: $RAILWAY_ENV_FILE not found${NC}"
    exit 1
fi

# Check if file is writable
if [ ! -w "$RAILWAY_ENV_FILE" ]; then
    echo -e "${YELLOW}⚠️  File is not writable. Attempting to fix permissions...${NC}"
    chmod u+w "$RAILWAY_ENV_FILE" 2>/dev/null || {
        echo -e "${RED}❌ Error: Cannot make file writable. Please check permissions.${NC}"
        exit 1
    }
fi

echo "Please provide your database connection string:"
echo ""
echo "Format: postgresql://postgres:[password]@db.[project-ref].supabase.co:5432/postgres?sslmode=require"
echo ""
echo "Or get it from: Supabase Dashboard > Project Settings > Database > Connection String (URI format)"
echo ""
read -p "DATABASE_URL: " DATABASE_URL

if [ -z "$DATABASE_URL" ]; then
    echo -e "${RED}❌ Error: DATABASE_URL cannot be empty${NC}"
    exit 1
fi

# Check if DATABASE_URL line exists
if grep -q "^DATABASE_URL=" "$RAILWAY_ENV_FILE"; then
    # Update existing DATABASE_URL
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        sed -i '' "s|^DATABASE_URL=.*|DATABASE_URL=$DATABASE_URL|" "$RAILWAY_ENV_FILE"
    else
        # Linux
        sed -i "s|^DATABASE_URL=.*|DATABASE_URL=$DATABASE_URL|" "$RAILWAY_ENV_FILE"
    fi
    echo -e "${GREEN}✅ Updated existing DATABASE_URL in $RAILWAY_ENV_FILE${NC}"
else
    # Add new DATABASE_URL after DB_AUTO_MIGRATE line
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        sed -i '' "/^DB_AUTO_MIGRATE=/a\\
\\
# Database URL (PostgreSQL connection string)\\
# Format: postgresql://[username]:[password]@[host]:[port]/[database]?sslmode=[ssl_mode]\\
DATABASE_URL=$DATABASE_URL" "$RAILWAY_ENV_FILE"
    else
        # Linux
        sed -i "/^DB_AUTO_MIGRATE=/a\\
\\
# Database URL (PostgreSQL connection string)\\
# Format: postgresql://[username]:[password]@[host]:[port]/[database]?sslmode=[ssl_mode]\\
DATABASE_URL=$DATABASE_URL" "$RAILWAY_ENV_FILE"
    fi
    echo -e "${GREEN}✅ Added DATABASE_URL to $RAILWAY_ENV_FILE${NC}"
fi

echo ""
echo -e "${GREEN}✅ Successfully updated $RAILWAY_ENV_FILE${NC}"
echo ""
echo "To use the updated environment variables, run:"
echo "  source $RAILWAY_ENV_FILE"
echo ""
echo "Or verify the update:"
echo "  grep DATABASE_URL $RAILWAY_ENV_FILE"

