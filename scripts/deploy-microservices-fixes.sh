#!/bin/bash

# KYB Platform Microservices Fixes Deployment Script
# This script deploys fixes for frontend interface, service discovery, and BI gateway

set -e

echo "🚀 Starting KYB Platform Microservices Fixes Deployment"
echo "=================================================="

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
    print_error "Railway CLI is not installed. Please install it first."
    print_status "Visit: https://docs.railway.app/develop/cli"
    exit 1
fi

# Check if user is logged in to Railway
if ! railway whoami &> /dev/null; then
    print_error "Not logged in to Railway. Please run 'railway login' first."
    exit 1
fi

print_status "Logged in to Railway as: $(railway whoami)"

# Function to deploy a service
deploy_service() {
    local service_name=$1
    local service_path=$2
    local service_description=$3
    
    print_status "Deploying $service_description..."
    
    if [ ! -d "$service_path" ]; then
        print_error "Service directory not found: $service_path"
        return 1
    fi
    
    cd "$service_path"
    
    # Check if Railway project is linked
    if [ ! -f ".railway/project.json" ]; then
        print_warning "No Railway project linked for $service_name. Please link manually:"
        print_status "cd $service_path && railway link"
        cd - > /dev/null
        return 1
    fi
    
    # Deploy the service
    print_status "Building and deploying $service_name..."
    if railway up --detach; then
        print_success "$service_description deployed successfully"
    else
        print_error "Failed to deploy $service_description"
        cd - > /dev/null
        return 1
    fi
    
    cd - > /dev/null
}

# Function to test service health
test_service_health() {
    local service_url=$1
    local service_name=$2
    
    print_status "Testing health of $service_name..."
    
    if curl -s -f "$service_url/health" > /dev/null; then
        print_success "$service_name is healthy"
        return 0
    else
        print_error "$service_name health check failed"
        return 1
    fi
}

# Main deployment process
main() {
    print_status "Starting microservices fixes deployment..."
    
    # 1. Deploy Frontend Service Fix
    print_status "=== 1. Deploying Frontend Service Fix ==="
    deploy_service "kyb-frontend-production" "services/frontend" "Frontend Service (Fixed Interface)"
    
    # 2. Deploy Business Intelligence Gateway
    print_status "=== 2. Deploying Business Intelligence Gateway ==="
    deploy_service "kyb-business-intelligence-gateway-production" "cmd/business-intelligence-gateway" "Business Intelligence Gateway"
    
    # 3. Deploy Service Discovery
    print_status "=== 3. Deploying Service Discovery ==="
    deploy_service "kyb-service-discovery-production" "cmd/service-discovery" "Service Discovery Server"
    
    # Wait for deployments to complete
    print_status "Waiting for deployments to complete..."
    sleep 30
    
    # Test all services
    print_status "=== Testing Service Health ==="
    
    services=(
        "https://kyb-frontend-production.up.railway.app:Frontend Service"
        "https://kyb-business-intelligence-gateway-production.up.railway.app:Business Intelligence Gateway"
        "https://kyb-service-discovery-production.up.railway.app:Service Discovery"
        "https://kyb-api-gateway-production.up.railway.app:API Gateway"
        "https://kyb-classification-service-production.up.railway.app:Classification Service"
        "https://kyb-merchant-service-production.up.railway.app:Merchant Service"
        "https://kyb-monitoring-production.up.railway.app:Monitoring Service"
        "https://kyb-pipeline-service-production.up.railway.app:Pipeline Service"
    )
    
    healthy_services=0
    total_services=${#services[@]}
    
    for service_info in "${services[@]}"; do
        IFS=':' read -r url name <<< "$service_info"
        if test_service_health "$url" "$name"; then
            ((healthy_services++))
        fi
    done
    
    # Summary
    print_status "=== Deployment Summary ==="
    print_status "Total Services: $total_services"
    print_status "Healthy Services: $healthy_services"
    print_status "Unhealthy Services: $((total_services - healthy_services))"
    
    if [ $healthy_services -eq $total_services ]; then
        print_success "All services are healthy! 🎉"
    else
        print_warning "Some services are not healthy. Please check the logs."
    fi
    
    # Service URLs
    print_status "=== Service URLs ==="
    echo "Frontend Service: https://kyb-frontend-production.up.railway.app"
    echo "Business Intelligence Gateway: https://kyb-business-intelligence-gateway-production.up.railway.app"
    echo "Service Discovery: https://kyb-service-discovery-production.up.railway.app"
    echo "API Gateway: https://kyb-api-gateway-production.up.railway.app"
    echo "Classification Service: https://kyb-classification-service-production.up.railway.app"
    echo "Merchant Service: https://kyb-merchant-service-production.up.railway.app"
    echo "Monitoring Service: https://kyb-monitoring-production.up.railway.app"
    echo "Pipeline Service: https://kyb-pipeline-service-production.up.railway.app"
    
    print_success "Microservices fixes deployment completed!"
}

# Run main function
main "$@"
