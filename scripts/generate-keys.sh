#!/bin/bash

# KYB Platform - Secure Key Generation Script
# This script generates secure encryption keys and JWT secrets

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "🔐 KYB Platform - Secure Key Generation"
echo "======================================="

# Function to generate random string
generate_random_string() {
    local length=$1
    openssl rand -base64 $((length * 3 / 4)) | tr -d "=+/" | cut -c1-${length}
}

# Function to generate secure key
generate_secure_key() {
    local length=$1
    local name=$2
    
    echo -e "${BLUE}Generating ${name}...${NC}"
    local key=$(generate_random_string $length)
    echo -e "${GREEN}✅ ${name}:${NC} $key"
    echo ""
    echo "$key"
}

echo -e "${YELLOW}⚠️  IMPORTANT: Keep these keys secure and never share them!${NC}"
echo ""

# Generate ENCRYPTION_KEY (32 characters)
echo "🔑 ENCRYPTION_KEY (32 characters):"
ENCRYPTION_KEY=$(generate_secure_key 32 "ENCRYPTION_KEY")

echo ""

# Generate JWT_SECRET (32 characters)
echo "🔑 JWT_SECRET (32 characters):"
JWT_SECRET=$(generate_secure_key 32 "JWT_SECRET")

echo ""

# Generate additional security keys
echo "🔑 API_SECRET (24 characters):"
API_SECRET=$(generate_secure_key 24 "API_SECRET")

echo ""

# Create environment file with generated keys
echo -e "${BLUE}📝 Creating .env.railway.generated with your keys...${NC}"

cat > .env.railway.generated << EOF
# KYB Platform - Generated Environment Variables
# Generated on: $(date)
# ⚠️  KEEP THESE KEYS SECURE - NEVER COMMIT TO GIT

# Application
ENVIRONMENT=beta
BETA_MODE=true
PORT=8080

# Security Keys (GENERATED - REPLACE WITH YOUR VALUES)
JWT_SECRET=${JWT_SECRET}
ENCRYPTION_KEY=${ENCRYPTION_KEY}
API_SECRET=${API_SECRET}

# Database (Railway will set this automatically)
DATABASE_URL=postgresql://postgres:password@localhost:5432/kyb_beta
DATABASE_TYPE=postgresql

# Supabase Integration (FILL IN YOUR VALUES)
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your-anon-key
SUPABASE_SERVICE_ROLE_KEY=your-service-role-key

# Redis (Railway will set this if you add Redis)
REDIS_URL=redis://localhost:6379

# External APIs (FILL IN YOUR VALUES)
BUSINESS_DATA_API_KEY=your-api-key
BUSINESS_DATA_API_URL=https://api.example.com

# Monitoring
ANALYTICS_ENABLED=true
FEEDBACK_COLLECTION=true
LOG_LEVEL=info

# CORS (Railway will set this automatically)
CORS_ORIGIN=https://your-app.railway.app
EOF

echo -e "${GREEN}✅ Created .env.railway.generated${NC}"

echo ""
echo -e "${YELLOW}📋 NEXT STEPS:${NC}"
echo "1. Copy the keys above"
echo "2. Go to Railway dashboard → Variables tab"
echo "3. Add these environment variables:"
echo "   - JWT_SECRET: $JWT_SECRET"
echo "   - ENCRYPTION_KEY: $ENCRYPTION_KEY"
echo "   - API_SECRET: $API_SECRET"
echo ""
echo -e "${YELLOW}🔒 SECURITY NOTES:${NC}"
echo "• These keys are unique to your deployment"
echo "• Never share or commit them to version control"
echo "• Store them securely (password manager recommended)"
echo "• You can regenerate them anytime with this script"
echo ""
echo -e "${GREEN}🚀 Ready to deploy with secure keys!${NC}"
