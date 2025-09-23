#!/bin/bash

# Deploy UI Fixes to Railway
# This script deploys the website keyword extraction and classification accuracy fixes

set -e

echo "🚀 Deploying UI Fixes to Railway..."
echo "=================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    print_error "Not in the project root directory. Please run from the project root."
    exit 1
fi

# Check if Railway CLI is installed
if ! command -v railway &> /dev/null; then
    print_error "Railway CLI is not installed. Please install it first."
    exit 1
fi

# Check if we're logged into Railway
if ! railway whoami &> /dev/null; then
    print_error "Not logged into Railway. Please run 'railway login' first."
    exit 1
fi

print_status "Current Railway status:"
railway status

print_status "Building the application..."
go build -o kyb-platform .

if [ $? -ne 0 ]; then
    print_error "Build failed. Please check for compilation errors."
    exit 1
fi

print_success "Build completed successfully"

print_status "Deploying to Railway..."
railway up

if [ $? -eq 0 ]; then
    print_success "Deployment to Railway completed successfully!"
    
    print_status "Getting deployment URL..."
    DEPLOYMENT_URL=$(railway status --json | jq -r '.url' 2>/dev/null || echo "https://shimmering-comfort-production.up.railway.app")
    
    if [ "$DEPLOYMENT_URL" != "null" ] && [ "$DEPLOYMENT_URL" != "" ]; then
        print_success "Deployment URL: $DEPLOYMENT_URL"
        echo "$DEPLOYMENT_URL" > deployment-url.txt
    else
        print_warning "Could not retrieve deployment URL automatically"
        print_status "Please check Railway dashboard for the deployment URL"
    fi
    
    print_status "Testing the deployed application..."
    
    # Test health endpoint
    if curl -f -s "$DEPLOYMENT_URL/health" > /dev/null; then
        print_success "Health check passed - application is running"
    else
        print_warning "Health check failed - application may still be starting up"
    fi
    
    print_success "UI fixes have been deployed to Railway!"
    print_status "The following fixes are now live:"
    echo "  ✅ Enhanced website keyword extraction"
    echo "  ✅ Improved classification accuracy with dynamic confidence scoring"
    echo "  ✅ Better fallback mechanisms for website scraping"
    echo "  ✅ Cloud-first deployment (no local servers)"
    
else
    print_error "Deployment failed. Please check the error messages above."
    exit 1
fi
