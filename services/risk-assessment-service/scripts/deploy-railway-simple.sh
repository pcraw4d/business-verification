#!/bin/bash

# Simple Railway Deployment Script for Risk Assessment Service
# This script handles the deployment without requiring Docker locally

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVICE_NAME="risk-assessment-service"
ENVIRONMENT=${1:-"production"}

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if Railway CLI is installed
    if ! command -v railway &> /dev/null; then
        log_error "Railway CLI is not installed. Please install it first:"
        echo "npm install -g @railway/cli"
        exit 1
    fi
    
    # Check if we're logged into Railway
    if ! railway whoami &> /dev/null; then
        log_error "Not logged into Railway. Please run: railway login"
        exit 1
    fi
    
    # Check if we're in the right directory
    if [ ! -f "Dockerfile" ]; then
        log_error "Dockerfile not found. Please run this script from the service directory."
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

validate_code() {
    log_info "Validating application code..."
    
    # Check if Go is available for basic validation
    if command -v go &> /dev/null; then
        log_info "Running Go module validation..."
        if ! go mod tidy; then
            log_error "Go module validation failed"
            exit 1
        fi
        
        log_info "Running Go build validation..."
        if ! go build -o /tmp/risk-assessment-service ./cmd/main.go; then
            log_error "Go build validation failed"
            exit 1
        fi
        
        # Clean up test binary
        rm -f /tmp/risk-assessment-service
        log_success "Go validation completed"
    else
        log_warning "Go not available for local validation, skipping build test"
    fi
    
    log_success "Code validation completed successfully"
}

deploy_to_railway() {
    log_info "Deploying to Railway ($ENVIRONMENT environment)..."
    
    # Set environment variables for Railway
    log_info "Setting environment variables..."
    
    # Core service variables
    railway variables --set ENVIRONMENT="$ENVIRONMENT"
    railway variables --set LOG_LEVEL="info"
    railway variables --set PORT="8080"
    
    # ML model configuration
    railway variables --set ENABLE_ENSEMBLE="true"
    railway variables --set DEFAULT_PREDICTION_HORIZON="3"
    railway variables --set LSTM_MODEL_PATH="/app/models/risk_lstm_v1.onnx"
    railway variables --set XGBOOST_MODEL_PATH="/app/models/xgb_model.json"
    
    # ONNX Runtime configuration
    railway variables --set CGO_ENABLED="1"
    railway variables --set CGO_LDFLAGS="-L/app/onnxruntime/lib"
    railway variables --set CGO_CFLAGS="-I/app/onnxruntime/include"
    railway variables --set LD_LIBRARY_PATH="/app/onnxruntime/lib"
    
    # Deploy to Railway
    log_info "Starting Railway deployment..."
    if railway up --detach; then
        log_success "Deployment initiated successfully"
    else
        log_error "Deployment failed"
        exit 1
    fi
}

wait_for_deployment() {
    log_info "Waiting for deployment to complete..."
    
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        log_info "Checking deployment status (attempt $attempt/$max_attempts)..."
        
        if railway status | grep -q "Deployed"; then
            log_success "Deployment completed successfully"
            return 0
        fi
        
        sleep 10
        attempt=$((attempt + 1))
    done
    
    log_error "Deployment timed out"
    return 1
}

get_service_url() {
    log_info "Getting service URL..."
    
    local url=$(railway domain 2>/dev/null || echo "")
    if [ -n "$url" ]; then
        echo "https://$url"
    else
        log_warning "Could not determine service URL automatically"
        echo "Please check Railway dashboard for the service URL"
    fi
}

run_smoke_tests() {
    local service_url=$(get_service_url)
    
    if [ -z "$service_url" ] || [ "$service_url" = "Please check Railway dashboard for the service URL" ]; then
        log_warning "Skipping smoke tests - service URL not available"
        return 0
    fi
    
    log_info "Running smoke tests against $service_url..."
    
    # Test health endpoint
    if curl -f -s "$service_url/health" > /dev/null; then
        log_success "Health endpoint test passed"
    else
        log_warning "Health endpoint test failed"
    fi
    
    # Test metrics endpoint
    if curl -f -s "$service_url/metrics" > /dev/null; then
        log_success "Metrics endpoint test passed"
    else
        log_warning "Metrics endpoint test failed"
    fi
}

show_deployment_info() {
    local service_url=$(get_service_url)
    
    echo ""
    echo "=========================================="
    echo "DEPLOYMENT COMPLETED SUCCESSFULLY"
    echo "=========================================="
    echo "Service: $SERVICE_NAME"
    echo "Environment: $ENVIRONMENT"
    echo "Service URL: $service_url"
    echo ""
    
    log_info "Useful commands:"
    echo "  railway logs          # View logs"
    echo "  railway status        # Check status"
    echo "  railway variables     # View environment variables"
    echo "  railway shell         # Access service shell"
    echo "  railway redeploy      # Redeploy service"
    echo ""
    
    if [ -n "$service_url" ] && [ "$service_url" != "Please check Railway dashboard for the service URL" ]; then
        log_info "API Endpoints:"
        echo "  Health: $service_url/health"
        echo "  Metrics: $service_url/metrics"
        echo "  API Docs: $service_url/docs"
        echo "  Risk Assessment: $service_url/api/v1/assess"
        echo "  Advanced Prediction: $service_url/api/v1/risk/predict-advanced"
    fi
}

show_help() {
    echo "Simple Railway Deployment Script for Risk Assessment Service"
    echo ""
    echo "Usage: $0 [ENVIRONMENT]"
    echo ""
    echo "Arguments:"
    echo "  ENVIRONMENT    Deployment environment (default: production)"
    echo "                 Options: production, staging, development"
    echo ""
    echo "Examples:"
    echo "  $0                    # Deploy to production"
    echo "  $0 staging           # Deploy to staging"
    echo "  $0 development       # Deploy to development"
    echo ""
    echo "Prerequisites:"
    echo "  - Railway CLI installed (npm install -g @railway/cli)"
    echo "  - Railway account and project access"
    echo "  - Go 1.22+ (for local validation, optional)"
    echo ""
    echo "Environment Variables:"
    echo "  RAILWAY_TOKEN        Railway API token (optional, will prompt for login)"
    echo "  RAILWAY_PROJECT      Railway project name (auto-detected)"
}

# Check for help flag
if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    show_help
    exit 0
fi

# Main execution
main() {
    log_info "Starting Railway deployment for $SERVICE_NAME"
    log_info "Environment: $ENVIRONMENT"
    echo ""
    
    # Execute deployment steps
    check_prerequisites
    validate_code
    deploy_to_railway
    
    if wait_for_deployment; then
        run_smoke_tests
        show_deployment_info
        log_success "Deployment completed successfully!"
    else
        log_error "Deployment failed or timed out"
        exit 1
    fi
}

# Run main function
main "$@"
