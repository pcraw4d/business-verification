#!/bin/bash

# Railway Deployment Fix Script
# This script helps fix common Railway deployment issues

set -e

echo "🔧 Railway Deployment Fix Script"
echo "================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Function to check if Railway CLI is installed
check_railway_cli() {
    if ! command -v railway &> /dev/null; then
        print_error "Railway CLI is not installed"
        echo "Install it with: npm install -g @railway/cli"
        exit 1
    fi
    print_success "Railway CLI is installed"
}

# Function to check Railway login status
check_railway_login() {
    if ! railway whoami &> /dev/null; then
        print_error "Not logged into Railway"
        echo "Login with: railway login"
        exit 1
    fi
    print_success "Logged into Railway"
}

# Function to generate secure secrets
generate_secrets() {
    print_status "Generating secure secrets..."
    
    # Generate JWT secret
    JWT_SECRET=$(openssl rand -base64 32)
    print_success "Generated JWT_SECRET"
    
    # Generate encryption key
    ENCRYPTION_KEY=$(openssl rand -base64 32)
    print_success "Generated ENCRYPTION_KEY"
    
    # Generate API secret
    API_SECRET=$(openssl rand -base64 18)
    print_success "Generated API_SECRET"
    
    echo ""
    print_status "Generated secrets (add these to Railway variables):"
    echo "JWT_SECRET=$JWT_SECRET"
    echo "ENCRYPTION_KEY=$ENCRYPTION_KEY"
    echo "API_SECRET=$API_SECRET"
    echo ""
}

# Function to set Railway variables
set_railway_variables() {
    print_status "Setting Railway environment variables..."
    
    # Set server configuration
    railway variables set PORT=8080
    railway variables set HOST=0.0.0.0
    
    print_success "Set server configuration"
    
    # Check if secrets are provided
    if [ -n "$1" ] && [ -n "$2" ] && [ -n "$3" ]; then
        railway variables set JWT_SECRET="$1"
        railway variables set ENCRYPTION_KEY="$2"
        railway variables set API_SECRET="$3"
        print_success "Set security secrets"
    else
        print_warning "Secrets not provided, you'll need to set them manually"
        echo "Use the generated secrets above or run:"
        echo "railway variables set JWT_SECRET=your-secret"
        echo "railway variables set ENCRYPTION_KEY=your-key"
        echo "railway variables set API_SECRET=your-api-secret"
    fi
}

# Function to check PostgreSQL service
check_postgresql_service() {
    print_status "Checking PostgreSQL service..."
    
    if railway service list | grep -q postgresql; then
        print_success "PostgreSQL service found"
    else
        print_warning "PostgreSQL service not found"
        echo "Add it with: railway service add postgresql"
    fi
}

# Function to check environment variables
check_environment_variables() {
    print_status "Checking environment variables..."
    
    variables=$(railway variables list)
    
    # Check required variables
    required_vars=("JWT_SECRET" "DATABASE_URL" "PORT" "HOST")
    
    for var in "${required_vars[@]}"; do
        if echo "$variables" | grep -q "$var"; then
            print_success "$var is set"
        else
            print_error "$var is missing"
        fi
    done
}

# Function to redeploy
redeploy() {
    print_status "Redeploying application..."
    
    railway up
    
    print_success "Deployment initiated"
    echo "Check status with: railway status"
    echo "View logs with: railway logs --follow"
}

# Function to show logs
show_logs() {
    print_status "Showing recent logs..."
    railway logs --follow
}

# Function to show status
show_status() {
    print_status "Showing deployment status..."
    railway status
}

# Function to test health endpoint
test_health() {
    print_status "Testing health endpoint..."
    
    # Get the app URL
    APP_URL=$(railway status --json | jq -r '.services[0].url' 2>/dev/null || echo "")
    
    if [ -z "$APP_URL" ]; then
        print_error "Could not determine app URL"
        return 1
    fi
    
    echo "Testing: $APP_URL/health"
    
    # Test health endpoint
    response=$(curl -s -o /dev/null -w "%{http_code}" "$APP_URL/health" 2>/dev/null || echo "000")
    
    if [ "$response" = "200" ]; then
        print_success "Health check passed!"
        echo "Application is running at: $APP_URL"
    else
        print_error "Health check failed (HTTP $response)"
        echo "Check logs with: railway logs --follow"
    fi
}

# Function to show help
show_help() {
    echo "Railway Deployment Fix Script"
    echo ""
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  check       - Check Railway configuration and environment"
    echo "  generate    - Generate secure secrets"
    echo "  fix         - Fix common deployment issues"
    echo "  deploy      - Redeploy the application"
    echo "  logs        - Show application logs"
    echo "  status      - Show deployment status"
    echo "  health      - Test health endpoint"
    echo "  help        - Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 check                    # Check current configuration"
    echo "  $0 generate                 # Generate new secrets"
    echo "  $0 fix                      # Fix deployment issues"
    echo "  $0 deploy                   # Redeploy application"
    echo ""
}

# Main execution
main() {
    case "${1:-help}" in
        check)
            check_railway_cli
            check_railway_login
            check_postgresql_service
            check_environment_variables
            ;;
        generate)
            generate_secrets
            ;;
        fix)
            check_railway_cli
            check_railway_login
            check_postgresql_service
            set_railway_variables
            print_status "Fix complete. Run 'railway up' to redeploy."
            ;;
        deploy)
            check_railway_cli
            check_railway_login
            redeploy
            ;;
        logs)
            check_railway_cli
            check_railway_login
            show_logs
            ;;
        status)
            check_railway_cli
            check_railway_login
            show_status
            ;;
        health)
            check_railway_cli
            check_railway_login
            test_health
            ;;
        help|*)
            show_help
            ;;
    esac
}

# Run main function
main "$@"
