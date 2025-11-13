#!/bin/bash

# Setup Local Environment File
# Creates .env.local from railway.env with local development settings

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Setting up .env.local for local development...${NC}"

# Check if railway.env exists
if [ ! -f "railway.env" ]; then
    echo -e "${YELLOW}Warning: railway.env not found${NC}"
    echo "Creating .env.local from template..."
    cat > .env.local << 'EOF'
# Local Development Environment Variables
# Fill in your actual values

SUPABASE_URL=https://qpqhuqqmkjxsltzshfam.supabase.co
SUPABASE_ANON_KEY=your_supabase_anon_key_here
SUPABASE_SERVICE_ROLE_KEY=your_supabase_service_role_key_here
SUPABASE_JWT_SECRET=your_supabase_jwt_secret_here

DATABASE_URL=postgresql://postgres:[YOUR_PASSWORD]@db.qpqhuqqmkjxsltzshfam.supabase.co:5432/postgres

ENV=local
ENVIRONMENT=local
LOG_LEVEL=debug
EOF
    echo -e "${YELLOW}Please edit .env.local with your actual credentials${NC}"
    exit 0
fi

# Copy railway.env to .env.local
cp railway.env .env.local

# Update environment settings for local development
echo -e "${GREEN}Updating environment settings for local development...${NC}"

# Use sed to update values (works on both macOS and Linux)
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    sed -i '' 's/^ENV=production/ENV=local/' .env.local
    sed -i '' 's/^ENVIRONMENT=production/ENVIRONMENT=local/' .env.local
    sed -i '' 's/^LOG_LEVEL=info/LOG_LEVEL=debug/' .env.local
else
    # Linux
    sed -i 's/^ENV=production/ENV=local/' .env.local
    sed -i 's/^ENVIRONMENT=production/ENVIRONMENT=local/' .env.local
    sed -i 's/^LOG_LEVEL=info/LOG_LEVEL=debug/' .env.local
fi

# Add header comment
cat > .env.local.tmp << 'EOF'
# Local Development Environment Variables
# Generated from railway.env for local development
# Service URLs are auto-configured by docker-compose.local.yml
# DO NOT COMMIT THIS FILE - it contains sensitive credentials

EOF

cat .env.local >> .env.local.tmp
mv .env.local.tmp .env.local

echo -e "${GREEN}✅ .env.local created successfully!${NC}"
echo ""
echo "Next steps:"
echo "1. Review .env.local and ensure all credentials are correct"
echo "2. Start Docker Desktop (if using microservices mode)"
echo "3. Run: make start-local"
echo ""

