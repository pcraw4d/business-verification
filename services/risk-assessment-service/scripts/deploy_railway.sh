#!/bin/bash

# Railway Deployment Script for Risk Assessment Service
# This script handles the complete deployment process to Railway

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVICE_NAME="risk-assessment-service"
PROJECT_NAME="kyb-platform"
ENVIRONMENT="${ENVIRONMENT:-production}"
RAILWAY_TOKEN="${RAILWAY_TOKEN:-}"

echo -e "${BLUE}🚀 Railway Deployment Script for Risk Assessment Service${NC}"
echo "================================================================"
echo -e "Service: ${YELLOW}$SERVICE_NAME${NC}"
echo -e "Project: ${YELLOW}$PROJECT_NAME${NC}"
echo -e "Environment: ${YELLOW}$ENVIRONMENT${NC}"
echo "================================================================"

# Function to check prerequisites
check_prerequisites() {
    echo -e "${YELLOW}🔍 Checking prerequisites...${NC}"
    
    # Check if Railway CLI is installed
    if ! command -v railway &> /dev/null; then
        echo -e "${RED}❌ Railway CLI is not installed${NC}"
        echo "Please install it from: https://docs.railway.app/develop/cli"
        exit 1
    fi
    
    # Check if Railway token is set
    if [ -z "$RAILWAY_TOKEN" ]; then
        echo -e "${YELLOW}⚠️  Railway token not set. Please login:${NC}"
        railway login
    fi
    
    # Check if we're in the right directory
    if [ ! -f "Dockerfile" ] || [ ! -f "railway.json" ]; then
        echo -e "${RED}❌ Not in the correct directory. Please run from service root.${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ Prerequisites check passed${NC}"
}

# Function to validate configuration
validate_configuration() {
    echo -e "${YELLOW}🔧 Validating configuration...${NC}"
    
    # Check if required files exist
    local required_files=("Dockerfile" "railway.json" "go.mod" "cmd/main.go")
    for file in "${required_files[@]}"; do
        if [ ! -f "$file" ]; then
            echo -e "${RED}❌ Required file missing: $file${NC}"
            exit 1
        fi
    done
    
    # Validate Go module
    if ! go mod verify &> /dev/null; then
        echo -e "${RED}❌ Go module verification failed${NC}"
        exit 1
    fi
    
    # Check if service builds locally
    echo -e "${YELLOW}🔨 Testing local build...${NC}"
    if ! go build -o /tmp/risk-assessment-service-test ./cmd/main.go; then
        echo -e "${RED}❌ Local build failed${NC}"
        exit 1
    fi
    rm -f /tmp/risk-assessment-service-test
    
    echo -e "${GREEN}✅ Configuration validation passed${NC}"
}

# Function to run tests
run_tests() {
    echo -e "${YELLOW}🧪 Running tests...${NC}"
    
    # Use the CI test script if available
    if [ -f "scripts/run_ci_tests.sh" ]; then
        echo -e "${YELLOW}🔧 Running CI test suite...${NC}"
        chmod +x scripts/run_ci_tests.sh
        
        # Set test environment variables
        export DATABASE_URL="${DATABASE_URL:-postgres://postgres:postgres@localhost:5432/test_risk_assessment?sslmode=disable}"
        export REDIS_URL="${REDIS_URL:-redis://localhost:6379}"
        export SUPABASE_URL="${SUPABASE_URL:-https://test.supabase.co}"
        export SUPABASE_ANON_KEY="${SUPABASE_ANON_KEY:-test-key}"
        export SUPABASE_SERVICE_ROLE_KEY="${SUPABASE_SERVICE_ROLE_KEY:-test-service-key}"
        export LOG_LEVEL="${LOG_LEVEL:-debug}"
        
        # Run CI tests with appropriate flags
        if ! ./scripts/run_ci_tests.sh --skip-integration --skip-performance; then
            echo -e "${RED}❌ CI tests failed${NC}"
            exit 1
        fi
    else
        # Fallback to basic tests
        echo -e "${YELLOW}🔧 Running basic tests...${NC}"
        if ! go test -v ./...; then
            echo -e "${RED}❌ Unit tests failed${NC}"
            exit 1
        fi
    fi
    
    # Run load tests if service is running locally
    if curl -s http://localhost:8080/health &> /dev/null; then
        echo -e "${YELLOW}🔄 Running load tests...${NC}"
        if ! go run ./cmd/load_test.go -url=http://localhost:8080 -duration=1m -users=5 -type=load; then
            echo -e "${YELLOW}⚠️  Load tests failed, but continuing deployment${NC}"
        fi
    else
        echo -e "${YELLOW}⚠️  Service not running locally, skipping load tests${NC}"
    fi
    
    echo -e "${GREEN}✅ Tests completed${NC}"
}

# Function to setup Railway project
setup_railway_project() {
    echo -e "${YELLOW}🚂 Setting up Railway project...${NC}"
    
    # Check if project exists
    if ! railway status &> /dev/null; then
        echo -e "${YELLOW}📁 Creating Railway project...${NC}"
        railway init "$PROJECT_NAME" || {
            echo -e "${YELLOW}⚠️  Project might already exist, continuing...${NC}"
        }
    fi
    
    # Link to existing project if needed
    if ! railway status &> /dev/null; then
        echo -e "${YELLOW}🔗 Linking to existing project...${NC}"
        railway link "$PROJECT_NAME" || {
            echo -e "${RED}❌ Failed to link to Railway project${NC}"
            exit 1
        }
    fi
    
    echo -e "${GREEN}✅ Railway project setup completed${NC}"
}

# Function to set environment variables
set_environment_variables() {
    echo -e "${YELLOW}🔐 Setting environment variables...${NC}"
    
    # Read from railway-essential.env if it exists
    if [ -f "../../railway-essential.env" ]; then
        echo -e "${YELLOW}📋 Loading environment variables from railway-essential.env...${NC}"
        
        # Set variables from the file (skip comments and empty lines)
        while IFS= read -r line; do
            if [[ ! "$line" =~ ^[[:space:]]*# ]] && [[ -n "$line" ]]; then
                local key=$(echo "$line" | cut -d'=' -f1)
                local value=$(echo "$line" | cut -d'=' -f2-)
                
                if [ -n "$key" ] && [ -n "$value" ]; then
                    echo -e "${BLUE}Setting $key${NC}"
                    railway variables set "$key=$value" || {
                        echo -e "${YELLOW}⚠️  Failed to set $key, continuing...${NC}"
                    }
                fi
            fi
        done < "../../railway-essential.env"
    else
        echo -e "${YELLOW}⚠️  railway-essential.env not found, using defaults${NC}"
    fi
    
    # Set service-specific variables
    local service_vars=(
        "SERVICE_NAME=$SERVICE_NAME"
        "SERVICE_VERSION=1.0.0"
        "SERVICE_DESCRIPTION=Enhanced Risk Assessment Service with ML-powered predictions"
        "PERFORMANCE_MONITORING_ENABLED=true"
        "METRICS_ENABLED=true"
        "LOG_LEVEL=info"
        "LOG_FORMAT=json"
        "REDIS_POOL_SIZE=50"
        "REDIS_MIN_IDLE_CONNS=10"
        "REDIS_MAX_IDLE_CONNS=20"
        "REDIS_DIAL_TIMEOUT=5s"
        "REDIS_READ_TIMEOUT=3s"
        "REDIS_WRITE_TIMEOUT=3s"
        "REDIS_POOL_TIMEOUT=4s"
        "REDIS_IDLE_TIMEOUT=5m"
        "REDIS_MAX_RETRIES=3"
        "REDIS_ENABLE_FALLBACK=true"
        "REDIS_FALLBACK_TO_MEMORY=true"
        "REDIS_KEY_PREFIX=ra:"
    )
    
    for var in "${service_vars[@]}"; do
        local key=$(echo "$var" | cut -d'=' -f1)
        local value=$(echo "$var" | cut -d'=' -f2-)
        echo -e "${BLUE}Setting $key=$value${NC}"
        railway variables set "$var" || {
            echo -e "${YELLOW}⚠️  Failed to set $key, continuing...${NC}"
        }
    done
    
    echo -e "${GREEN}✅ Environment variables set${NC}"
}

# Function to deploy to Railway
deploy_to_railway() {
    echo -e "${YELLOW}🚀 Deploying to Railway...${NC}"
    
    # Deploy the service
    if railway up --detach; then
        echo -e "${GREEN}✅ Deployment initiated successfully${NC}"
    else
        echo -e "${RED}❌ Deployment failed${NC}"
        exit 1
    fi
    
    # Wait for deployment to complete
    echo -e "${YELLOW}⏳ Waiting for deployment to complete...${NC}"
    sleep 30
    
    # Get deployment URL
    local deployment_url=$(railway domain)
    if [ -n "$deployment_url" ]; then
        echo -e "${GREEN}🌐 Service deployed at: https://$deployment_url${NC}"
    else
        echo -e "${YELLOW}⚠️  Could not get deployment URL${NC}"
    fi
}

# Function to verify deployment
verify_deployment() {
    echo -e "${YELLOW}🔍 Verifying deployment...${NC}"
    
    # Get deployment URL
    local deployment_url=$(railway domain)
    if [ -z "$deployment_url" ]; then
        echo -e "${RED}❌ Could not get deployment URL${NC}"
        return 1
    fi
    
    local service_url="https://$deployment_url"
    
    # Wait for service to be ready
    echo -e "${YELLOW}⏳ Waiting for service to be ready...${NC}"
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s "$service_url/health" &> /dev/null; then
            echo -e "${GREEN}✅ Service is healthy${NC}"
            break
        fi
        
        echo -e "${YELLOW}Attempt $attempt/$max_attempts - waiting 10 seconds...${NC}"
        sleep 10
        attempt=$((attempt + 1))
    done
    
    if [ $attempt -gt $max_attempts ]; then
        echo -e "${RED}❌ Service failed to become ready${NC}"
        return 1
    fi
    
    # Test key endpoints
    echo -e "${YELLOW}🧪 Testing key endpoints...${NC}"
    
    local endpoints=(
        "/health"
        "/api/v1/performance/health"
        "/api/v1/performance/stats"
        "/metrics"
    )
    
    for endpoint in "${endpoints[@]}"; do
        if curl -s "$service_url$endpoint" &> /dev/null; then
            echo -e "${GREEN}✅ $endpoint is working${NC}"
        else
            echo -e "${YELLOW}⚠️  $endpoint is not responding${NC}"
        fi
    done
    
    # Run a quick load test
    echo -e "${YELLOW}🚀 Running quick load test...${NC}"
    if go run ./cmd/load_test.go -url="$service_url" -duration=1m -users=5 -type=load -verbose; then
        echo -e "${GREEN}✅ Load test passed${NC}"
    else
        echo -e "${YELLOW}⚠️  Load test failed, but service is deployed${NC}"
    fi
    
    echo -e "${GREEN}🎉 Deployment verification completed${NC}"
    echo -e "${BLUE}📊 Service URL: $service_url${NC}"
    echo -e "${BLUE}📈 Performance Stats: $service_url/api/v1/performance/stats${NC}"
    echo -e "${BLUE}🔍 Health Check: $service_url/health${NC}"
    echo -e "${BLUE}📊 Metrics: $service_url/metrics${NC}"
}

# Function to show deployment status
show_deployment_status() {
    echo -e "${BLUE}📊 Deployment Status${NC}"
    echo "================================================================"
    
    # Show Railway status
    railway status
    
    # Show logs
    echo -e "${YELLOW}📋 Recent logs:${NC}"
    railway logs --tail 20
    
    # Show environment variables
    echo -e "${YELLOW}🔐 Environment variables:${NC}"
    railway variables
}

# Main deployment function
main() {
    echo -e "${BLUE}🎯 Starting Railway deployment process${NC}"
    echo ""
    
    # Run deployment steps
    check_prerequisites
    validate_configuration
    run_tests
    setup_railway_project
    set_environment_variables
    deploy_to_railway
    verify_deployment
    show_deployment_status
    
    echo ""
    echo -e "${GREEN}🎉 Railway deployment completed successfully!${NC}"
    echo "================================================================"
    echo -e "Service: ${YELLOW}$SERVICE_NAME${NC}"
    echo -e "Environment: ${YELLOW}$ENVIRONMENT${NC}"
    echo -e "Status: ${GREEN}Deployed and Verified${NC}"
    echo ""
    echo -e "${BLUE}Next steps:${NC}"
    echo "1. Monitor service health and performance"
    echo "2. Run comprehensive load tests"
    echo "3. Set up monitoring and alerting"
    echo "4. Configure external integrations"
    echo ""
    echo -e "${BLUE}Useful commands:${NC}"
    echo "- View logs: railway logs"
    echo "- Check status: railway status"
    echo "- View variables: railway variables"
    echo "- Open service: railway open"
}

# Handle script arguments
case "${1:-}" in
    "verify")
        verify_deployment
        ;;
    "status")
        show_deployment_status
        ;;
    "logs")
        railway logs --tail 100
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [verify|status|logs|help]"
        echo ""
        echo "Commands:"
        echo "  (no args)  - Full deployment process"
        echo "  verify     - Verify existing deployment"
        echo "  status     - Show deployment status"
        echo "  logs       - Show recent logs"
        echo "  help       - Show this help message"
        ;;
    *)
        main
        ;;
esac
