#!/bin/bash

# Railway Issue Diagnostic Script
# This script helps identify the specific cause of Railway deployment failures

set -e

echo "🔍 Railway Issue Diagnostic Script"
echo "=================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Function to check Railway CLI
check_railway_cli() {
    print_status "Checking Railway CLI..."
    
    if command -v railway &> /dev/null; then
        print_success "Railway CLI is installed"
        railway_version=$(railway --version 2>/dev/null || echo "unknown")
        echo "Version: $railway_version"
    else
        print_error "Railway CLI is not installed"
        echo "Install with: npm install -g @railway/cli"
        return 1
    fi
}

# Function to check Railway login
check_railway_login() {
    print_status "Checking Railway login status..."
    
    if railway whoami &> /dev/null; then
        print_success "Logged into Railway"
        user=$(railway whoami 2>/dev/null || echo "unknown")
        echo "User: $user"
    else
        print_error "Not logged into Railway"
        echo "Login with: railway login"
        return 1
    fi
}

# Function to check project status
check_project_status() {
    print_status "Checking project status..."
    
    if railway status &> /dev/null; then
        print_success "Project is accessible"
        
        # Get deployment status
        status_output=$(railway status 2>/dev/null || echo "")
        echo "Status: $status_output"
    else
        print_error "Cannot access project"
        echo "Check if you're in the correct project directory"
        return 1
    fi
}

# Function to check environment variables
check_environment_variables() {
    print_status "Checking environment variables..."
    
    variables=$(railway variables list 2>/dev/null || echo "")
    
    if [ -z "$variables" ]; then
        print_error "Cannot retrieve environment variables"
        return 1
    fi
    
    # Check critical variables
    critical_vars=("JWT_SECRET" "DATABASE_URL" "PORT" "HOST")
    
    for var in "${critical_vars[@]}"; do
        if echo "$variables" | grep -q "$var"; then
            print_success "$var is set"
        else
            print_error "$var is missing"
        fi
    done
    
    # Check for empty values
    echo ""
    print_status "Checking for empty values..."
    
    if echo "$variables" | grep -q "JWT_SECRET="; then
        jwt_value=$(echo "$variables" | grep "JWT_SECRET=" | cut -d'=' -f2)
        if [ -z "$jwt_value" ]; then
            print_error "JWT_SECRET is empty"
        else
            print_success "JWT_SECRET has a value"
        fi
    fi
    
    if echo "$variables" | grep -q "DATABASE_URL="; then
        db_value=$(echo "$variables" | grep "DATABASE_URL=" | cut -d'=' -f2)
        if [ -z "$db_value" ]; then
            print_error "DATABASE_URL is empty"
        else
            print_success "DATABASE_URL has a value"
        fi
    fi
}

# Function to check PostgreSQL service
check_postgresql_service() {
    print_status "Checking PostgreSQL service..."
    
    services=$(railway service list 2>/dev/null || echo "")
    
    if echo "$services" | grep -q postgresql; then
        print_success "PostgreSQL service is added"
    else
        print_error "PostgreSQL service is not added"
        echo "Add it with: railway service add postgresql"
    fi
}

# Function to check recent logs
check_recent_logs() {
    print_status "Checking recent logs..."
    
    # Get last 20 lines of logs
    recent_logs=$(railway logs --tail 20 2>/dev/null || echo "Cannot retrieve logs")
    
    if [ "$recent_logs" = "Cannot retrieve logs" ]; then
        print_error "Cannot retrieve logs"
        return 1
    fi
    
    echo "Recent logs:"
    echo "$recent_logs"
    
    # Check for common error patterns
    echo ""
    print_status "Analyzing logs for common issues..."
    
    if echo "$recent_logs" | grep -q "JWT secret is required"; then
        print_error "Found: JWT secret is required"
        echo "Solution: Set JWT_SECRET environment variable"
    fi
    
    if echo "$recent_logs" | grep -q "Failed to connect to database"; then
        print_error "Found: Database connection failed"
        echo "Solution: Add PostgreSQL service and check DATABASE_URL"
    fi
    
    if echo "$recent_logs" | grep -q "Port already in use"; then
        print_error "Found: Port already in use"
        echo "Solution: Check PORT environment variable"
    fi
    
    if echo "$recent_logs" | grep -q "Permission denied"; then
        print_error "Found: Permission denied"
        echo "Solution: Check file permissions in Dockerfile"
    fi
    
    if echo "$recent_logs" | grep -q "Healthcheck failed"; then
        print_error "Found: Healthcheck failed"
        echo "Solution: Check if application is starting properly"
    fi
}

# Function to check health endpoint
check_health_endpoint() {
    print_status "Checking health endpoint..."
    
    # Get app URL
    app_url=$(railway status --json 2>/dev/null | jq -r '.services[0].url' 2>/dev/null || echo "")
    
    if [ -z "$app_url" ]; then
        print_warning "Cannot determine app URL"
        return 1
    fi
    
    echo "App URL: $app_url"
    
    # Test health endpoint
    health_url="$app_url/health"
    echo "Testing: $health_url"
    
    response=$(curl -s -o /dev/null -w "%{http_code}" "$health_url" 2>/dev/null || echo "000")
    
    if [ "$response" = "200" ]; then
        print_success "Health endpoint is responding (HTTP 200)"
    else
        print_error "Health endpoint failed (HTTP $response)"
        echo "This is likely the cause of the healthcheck failure"
    fi
}

# Function to provide recommendations
provide_recommendations() {
    echo ""
    print_status "Recommendations:"
    echo "=================="
    
    echo "1. Check Railway logs for specific errors:"
    echo "   railway logs --follow"
    echo ""
    
    echo "2. Verify environment variables are set:"
    echo "   railway variables list"
    echo ""
    
    echo "3. Add PostgreSQL service if missing:"
    echo "   railway service add postgresql"
    echo ""
    
    echo "4. Set required environment variables:"
    echo "   railway variables set JWT_SECRET=your-secret"
    echo "   railway variables set PORT=8080"
    echo "   railway variables set HOST=0.0.0.0"
    echo ""
    
    echo "5. Redeploy after fixing issues:"
    echo "   railway up"
    echo ""
    
    echo "6. Test health endpoint:"
    echo "   curl https://your-app.railway.app/health"
    echo ""
}

# Main execution
main() {
    echo "Starting Railway issue diagnosis..."
    echo ""
    
    # Run all checks
    check_railway_cli
    echo ""
    
    check_railway_login
    echo ""
    
    check_project_status
    echo ""
    
    check_environment_variables
    echo ""
    
    check_postgresql_service
    echo ""
    
    check_recent_logs
    echo ""
    
    check_health_endpoint
    echo ""
    
    provide_recommendations
}

# Run main function
main "$@"
