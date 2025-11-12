#!/bin/bash

# Setup environment variables for KYB Platform
# This script helps you set up the DATABASE_URL and other required environment variables

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔧 KYB Platform Environment Setup${NC}"
echo "=================================="
echo ""

# Check if .env file exists
if [ -f .env ]; then
    echo -e "${YELLOW}⚠️  .env file already exists${NC}"
    read -p "Do you want to overwrite it? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Keeping existing .env file"
        exit 0
    fi
fi

echo "Please provide the following information:"
echo ""

# Get DATABASE_URL
echo -e "${BLUE}1. Database Connection String${NC}"
echo "   Get this from Supabase Dashboard > Project Settings > Database > Connection String"
echo "   Format: postgresql://postgres:[password]@db.[project-ref].supabase.co:5432/postgres"
read -p "   DATABASE_URL: " DATABASE_URL

# Get Supabase URL
echo ""
echo -e "${BLUE}2. Supabase Project URL${NC}"
echo "   Format: https://[project-ref].supabase.co"
read -p "   SUPABASE_URL: " SUPABASE_URL

# Get Supabase Anon Key
echo ""
echo -e "${BLUE}3. Supabase Anon Key${NC}"
echo "   Get this from Supabase Dashboard > Project Settings > API"
read -p "   SUPABASE_ANON_KEY: " SUPABASE_ANON_KEY

# Get Port (optional)
echo ""
echo -e "${BLUE}4. Server Port (optional, default: 8080)${NC}"
read -p "   PORT [8080]: " PORT
PORT=${PORT:-8080}

# Create .env file
cat > .env << ENVFILE
# Database Configuration
DATABASE_URL=${DATABASE_URL}

# Supabase Configuration
SUPABASE_URL=${SUPABASE_URL}
SUPABASE_ANON_KEY=${SUPABASE_ANON_KEY}

# Server Configuration
PORT=${PORT}
SERVICE_NAME=kyb-platform-v4-complete
ENVFILE

echo ""
echo -e "${GREEN}✅ .env file created successfully!${NC}"
echo ""
echo "To use these environment variables, run:"
echo "  source .env"
echo ""
echo "Or use a tool like 'direnv' or 'dotenv' to load them automatically"
