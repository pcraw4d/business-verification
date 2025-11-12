#!/bin/bash

# Interactive script to update DATABASE_URL in railway.env
# This script prompts for individual components and constructs the connection string

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

RAILWAY_ENV_FILE="railway.env"

echo -e "${BLUE}🔧 Interactive DATABASE_URL Update${NC}"
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

echo "We'll construct the DATABASE_URL from individual components."
echo "You can press Enter to use defaults (from existing railway.env values)."
echo ""

# Read existing values as defaults
DEFAULT_HOST=$(grep "^DB_HOST=" "$RAILWAY_ENV_FILE" 2>/dev/null | cut -d'=' -f2 | tr -d '"' || echo "db.qpqhuqqmkjxsltzshfam.supabase.co")
DEFAULT_PORT=$(grep "^DB_PORT=" "$RAILWAY_ENV_FILE" 2>/dev/null | cut -d'=' -f2 | tr -d '"' || echo "5432")
DEFAULT_USER=$(grep "^DB_USERNAME=" "$RAILWAY_ENV_FILE" 2>/dev/null | cut -d'=' -f2 | tr -d '"' || echo "postgres")
DEFAULT_DB=$(grep "^DB_DATABASE=" "$RAILWAY_ENV_FILE" 2>/dev/null | cut -d'=' -f2 | tr -d '"' || echo "postgres")

read -p "Database Host [$DEFAULT_HOST]: " DB_HOST
DB_HOST=${DB_HOST:-$DEFAULT_HOST}

read -p "Database Port [$DEFAULT_PORT]: " DB_PORT
DB_PORT=${DB_PORT:-$DEFAULT_PORT}

read -p "Database Username [$DEFAULT_USER]: " DB_USER
DB_USER=${DB_USER:-$DEFAULT_USER}

read -p "Database Name [$DEFAULT_DB]: " DB_NAME
DB_NAME=${DB_NAME:-$DEFAULT_DB}

echo ""
echo -e "${YELLOW}⚠️  Database Password (required)${NC}"
read -sp "Database Password: " DB_PASSWORD
echo ""

if [ -z "$DB_PASSWORD" ]; then
    echo -e "${RED}❌ Error: Database password cannot be empty${NC}"
    exit 1
fi

# Construct DATABASE_URL
DATABASE_URL="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require"

echo ""
echo -e "${BLUE}Constructed DATABASE_URL:${NC}"
echo "postgresql://${DB_USER}:***@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require"
echo ""
read -p "Is this correct? (y/N): " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Aborted."
    exit 0
fi

# Update the file
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

