#!/bin/bash

# Enhanced Railway Deployment Script
# This script deploys the KYB Platform to Railway with proper Supabase integration

set -e

echo "🚀 Starting Enhanced Railway Deployment..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Railway CLI is installed
if ! command -v railway &> /dev/null; then
    print_error "Railway CLI is not installed. Please install it first:"
    echo "npm install -g @railway/cli"
    exit 1
fi

# Check if user is logged in to Railway
if ! railway whoami &> /dev/null; then
    print_error "Not logged in to Railway. Please login first:"
    echo "railway login"
    exit 1
fi

print_status "Checking Railway project status..."

# Get current project
PROJECT_ID=$(railway status --json | jq -r '.project.id' 2>/dev/null || echo "")
if [ -z "$PROJECT_ID" ]; then
    print_error "No Railway project found. Please link to a project first:"
    echo "railway link"
    exit 1
fi

print_success "Connected to Railway project: $PROJECT_ID"

# Check if environment variables are set
print_status "Checking environment variables..."

# Required Supabase environment variables
REQUIRED_VARS=(
    "SUPABASE_URL"
    "SUPABASE_API_KEY"
    "SUPABASE_SERVICE_ROLE_KEY"
    "SUPABASE_JWT_SECRET"
)

MISSING_VARS=()

for var in "${REQUIRED_VARS[@]}"; do
    if [ -z "${!var}" ]; then
        MISSING_VARS+=("$var")
    fi
done

if [ ${#MISSING_VARS[@]} -gt 0 ]; then
    print_warning "Missing required environment variables:"
    for var in "${MISSING_VARS[@]}"; do
        echo "  - $var"
    done
    echo ""
    print_status "You can set these in Railway dashboard or use the following commands:"
    for var in "${MISSING_VARS[@]}"; do
        echo "  railway variables set $var=your_value_here"
    done
    echo ""
    read -p "Do you want to continue with fallback mode? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_error "Deployment cancelled. Please set the required environment variables."
        exit 1
    fi
    print_warning "Continuing with fallback mode (mock data will be used)"
else
    print_success "All required environment variables are set"
fi

# Set Railway environment variables if they're available locally
print_status "Setting Railway environment variables..."

# Set basic configuration
railway variables set ENV=production
railway variables set PORT=8080
railway variables set HOST=0.0.0.0
railway variables set LOG_LEVEL=info
railway variables set LOG_FORMAT=json

# Set CORS configuration
railway variables set CORS_ALLOWED_ORIGINS="*"
railway variables set CORS_ALLOWED_METHODS="GET,POST,PUT,DELETE,OPTIONS"
railway variables set CORS_ALLOWED_HEADERS="*"
railway variables set CORS_ALLOW_CREDENTIALS=true

# Set rate limiting
railway variables set RATE_LIMIT_ENABLED=true
railway variables set RATE_LIMIT_REQUESTS_PER=100
railway variables set RATE_LIMIT_WINDOW_SIZE=60

# Set feature flags
railway variables set FEATURE_BUSINESS_CLASSIFICATION=true
railway variables set FEATURE_RISK_ASSESSMENT=true
railway variables set FEATURE_COMPLIANCE_FRAMEWORK=true

# Set Supabase variables if available
if [ -n "$SUPABASE_URL" ]; then
    railway variables set SUPABASE_URL="$SUPABASE_URL"
    print_success "Set SUPABASE_URL"
fi

if [ -n "$SUPABASE_API_KEY" ]; then
    railway variables set SUPABASE_API_KEY="$SUPABASE_API_KEY"
    print_success "Set SUPABASE_API_KEY"
fi

if [ -n "$SUPABASE_SERVICE_ROLE_KEY" ]; then
    railway variables set SUPABASE_SERVICE_ROLE_KEY="$SUPABASE_SERVICE_ROLE_KEY"
    print_success "Set SUPABASE_SERVICE_ROLE_KEY"
fi

if [ -n "$SUPABASE_JWT_SECRET" ]; then
    railway variables set SUPABASE_JWT_SECRET="$SUPABASE_JWT_SECRET"
    print_success "Set SUPABASE_JWT_SECRET"
fi

# Set JWT secret if not already set
if [ -z "$JWT_SECRET" ]; then
    JWT_SECRET=$(openssl rand -base64 32 2>/dev/null || echo "default_jwt_secret_change_in_production")
fi
railway variables set JWT_SECRET="$JWT_SECRET"
print_success "Set JWT_SECRET"

# Build and deploy
print_status "Building and deploying to Railway..."

# Use the production Dockerfile
railway up --dockerfile Dockerfile.production

print_success "Deployment initiated!"

# Wait for deployment to complete
print_status "Waiting for deployment to complete..."
sleep 30

# Get deployment URL
DEPLOYMENT_URL=$(railway status --json | jq -r '.deployment.url' 2>/dev/null || echo "")
if [ -n "$DEPLOYMENT_URL" ]; then
    print_success "Deployment URL: $DEPLOYMENT_URL"
    
    # Test the deployment
    print_status "Testing deployment..."
    
    # Test health endpoint
    if curl -s "$DEPLOYMENT_URL/health" > /dev/null; then
        print_success "Health check passed"
        
        # Test classification endpoint
        if curl -s -X POST "$DEPLOYMENT_URL/v1/classify" \
            -H "Content-Type: application/json" \
            -d '{"business_name":"Test Company","description":"A technology company"}' > /dev/null; then
            print_success "Classification endpoint working"
        else
            print_warning "Classification endpoint test failed"
        fi
        
        # Test merchants endpoint
        if curl -s "$DEPLOYMENT_URL/api/v1/merchants" > /dev/null; then
            print_success "Merchants endpoint working"
        else
            print_warning "Merchants endpoint test failed"
        fi
        
    else
        print_error "Health check failed"
    fi
    
    echo ""
    print_success "🎉 Deployment completed successfully!"
    echo ""
    echo "📊 Platform URLs:"
    echo "  - Main Platform: $DEPLOYMENT_URL"
    echo "  - Health Check: $DEPLOYMENT_URL/health"
    echo "  - Business Intelligence: $DEPLOYMENT_URL/business-intelligence.html"
    echo "  - Merchant Hub: $DEPLOYMENT_URL/merchant-hub.html"
    echo "  - Merchant Portfolio: $DEPLOYMENT_URL/merchant-portfolio.html"
    echo ""
    echo "🔧 Next Steps:"
    echo "  1. Test all UI pages to ensure they're working"
    echo "  2. Configure Supabase environment variables if not already done"
    echo "  3. Set up monitoring and alerting"
    echo "  4. Configure authentication if needed"
    echo ""
    echo "📝 To view logs: railway logs"
    echo "📝 To view variables: railway variables"
    echo "📝 To redeploy: railway up"
    
else
    print_warning "Could not determine deployment URL. Check Railway dashboard for the URL."
fi

print_success "Enhanced Railway deployment script completed!"
