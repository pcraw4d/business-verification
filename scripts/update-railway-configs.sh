#!/bin/bash

# Railway Service Configuration Update Script
# This script updates Railway services to use the new monorepo structure

set -e

echo "🚀 Updating Railway Service Configurations..."
echo "============================================="

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

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Railway CLI is installed
if ! command -v railway &> /dev/null; then
    print_error "Railway CLI not found. Please install it first."
    exit 1
fi

# Function to update service configuration
update_service_config() {
    local service_name=$1
    local service_path=$2
    local dockerfile_path=$3
    local start_command=$4
    
    print_status "Updating $service_name service configuration..."
    
    # Switch to the service
    railway service $service_name
    
    # Set the root directory
    railway variables --set "RAILWAY_ROOT_DIRECTORY=$service_path"
    
    # Set the Dockerfile path
    railway variables --set "RAILWAY_DOCKERFILE_PATH=$dockerfile_path"
    
    # Set the start command
    railway variables --set "RAILWAY_START_COMMAND=$start_command"
    
    print_success "$service_name service configuration updated"
}

# Update API service configuration
print_status "Updating API service (shimmering-comfort)..."
update_service_config "shimmering-comfort" "services/api" "Dockerfile" "./server"

# Update Frontend service configuration  
print_status "Updating Frontend service (frontend-UI)..."
update_service_config "frontend-UI" "services/frontend" "Dockerfile" "./frontend-server"

print_success "All Railway service configurations updated!"
echo ""
echo "📋 Updated Configurations:"
echo "  API Service:"
echo "    - Root Directory: services/api"
echo "    - Dockerfile: Dockerfile"
echo "    - Start Command: ./server"
echo ""
echo "  Frontend Service:"
echo "    - Root Directory: services/frontend"
echo "    - Dockerfile: Dockerfile"
echo "    - Start Command: ./frontend-server"
echo ""
print_warning "Next steps:"
echo "1. Commit and push your changes to trigger new deployments"
echo "2. Monitor Railway dashboard for successful deployments"
echo "3. Test both services to ensure they're working correctly"
