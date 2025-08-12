#!/bin/bash

# Quick Beta Deployment using Railway
# This script deploys the KYB Platform to Railway for public beta testing

set -e

echo "🚀 Quick Beta Deployment using Railway..."

# Configuration
PROJECT_NAME="${PROJECT_NAME:-kyb-platform-beta}"
RAILWAY_TOKEN="${RAILWAY_TOKEN}"

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

# Check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check if Railway CLI is installed
    if ! command -v railway &> /dev/null; then
        print_error "Railway CLI is not installed."
        echo "Install it with: npm install -g @railway/cli"
        exit 1
    fi
    
    # Check if logged in to Railway
    if ! railway whoami &> /dev/null; then
        print_error "Not logged in to Railway. Please run: railway login"
        exit 1
    fi
    
    print_success "Prerequisites check completed"
}

# Create Railway configuration
create_railway_config() {
    print_status "Creating Railway configuration..."
    
    # Create railway.json
    cat > railway.json << EOF
{
  "$schema": "https://railway.app/railway.schema.json",
  "build": {
    "builder": "DOCKERFILE",
    "dockerfilePath": "Dockerfile.beta"
  },
  "deploy": {
    "startCommand": "./kyb-platform",
    "healthcheckPath": "/health",
    "healthcheckTimeout": 300,
    "restartPolicyType": "ON_FAILURE",
    "restartPolicyMaxRetries": 10
  }
}
EOF

    # Create Dockerfile for Railway
    cat > Dockerfile.beta << EOF
# Multi-stage build for KYB Platform Beta
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o kyb-platform ./cmd/api

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/kyb-platform .

# Copy configuration files
COPY --from=builder /app/configs/beta ./configs/beta

# Copy static files for web interface
COPY --from=builder /app/web/dist ./web/dist

# Expose port
EXPOSE 8080

# Run the application
CMD ["./kyb-platform"]
EOF

    print_success "Railway configuration created"
}

# Deploy to Railway
deploy_to_railway() {
    print_status "Deploying to Railway..."
    
    # Create new project if it doesn't exist
    if ! railway project list | grep -q "$PROJECT_NAME"; then
        print_status "Creating new Railway project: $PROJECT_NAME"
        railway project create "$PROJECT_NAME"
    fi
    
    # Link to the project
    railway link "$PROJECT_NAME"
    
    # Set environment variables
    railway variables set ENVIRONMENT=beta
    railway variables set BETA_MODE=true
    railway variables set ANALYTICS_ENABLED=true
    railway variables set FEEDBACK_COLLECTION=true
    
    # Deploy
    railway up
    
    print_success "Deployment to Railway completed"
}

# Get deployment URL
get_deployment_url() {
    print_status "Getting deployment URL..."
    
    # Get the deployment URL
    DEPLOYMENT_URL=$(railway status --json | jq -r '.deployment.url')
    
    if [[ -z "$DEPLOYMENT_URL" || "$DEPLOYMENT_URL" == "null" ]]; then
        print_error "Could not get deployment URL"
        return 1
    fi
    
    print_success "Deployment URL: $DEPLOYMENT_URL"
    
    # Generate shareable links
    cat > SHAREABLE_LINKS_RAILWAY.md << EOF
# KYB Platform Beta Testing - Railway Deployment

## 🌐 Public Access Links

### Web Interface (For Non-Technical Users)
- **Main Dashboard**: $DEPLOYMENT_URL
- **User Registration**: $DEPLOYMENT_URL/register
- **User Login**: $DEPLOYMENT_URL/login

### API Documentation (For Technical Users)
- **API Documentation**: $DEPLOYMENT_URL/docs
- **API Base URL**: $DEPLOYMENT_URL/api/v1

### Health Check
- **Health Status**: $DEPLOYMENT_URL/health

## 🔑 Beta Testing Credentials

### Test Users
- **Compliance Officer**: compliance@beta.kybplatform.com / password123
- **Risk Manager**: risk@beta.kybplatform.com / password123
- **Business Analyst**: analyst@beta.kybplatform.com / password123

### API Access
- **API Key**: your-api-key-here
- **Rate Limit**: 1000 requests per minute

## 📋 Beta Testing Instructions

### For Non-Technical Users:
1. Visit $DEPLOYMENT_URL
2. Click "Register" to create an account
3. Use the web interface to test business classification
4. Explore risk assessment and compliance features
5. Provide feedback through the built-in feedback system

### For Technical Users:
1. Visit $DEPLOYMENT_URL/docs for API documentation
2. Use the API endpoints for integration testing
3. Test authentication and rate limiting
4. Validate all platform features programmatically

## 🚨 Important Notes
- This is a beta environment deployed on Railway
- Data may be reset periodically
- Report any issues through the feedback system
- Beta testing period: [Start Date] to [End Date]

## 📞 Support
- **Email**: beta-support@kybplatform.com
- **Documentation**: $DEPLOYMENT_URL/docs
- **Feedback**: Use the feedback form in the web interface

## 🔧 Railway Management
- **Project Dashboard**: https://railway.app/project/[PROJECT_ID]
- **Deployment Logs**: railway logs
- **Environment Variables**: railway variables
- **Scale Resources**: railway scale
EOF

    print_success "Shareable links generated in SHAREABLE_LINKS_RAILWAY.md"
}

# Health check
health_check() {
    print_status "Performing health check..."
    
    # Wait for deployment to be ready
    sleep 60
    
    # Get deployment URL
    DEPLOYMENT_URL=$(railway status --json | jq -r '.deployment.url')
    
    if [[ -z "$DEPLOYMENT_URL" || "$DEPLOYMENT_URL" == "null" ]]; then
        print_error "Could not get deployment URL for health check"
        return 1
    fi
    
    # Check health endpoint
    if curl -f -s "$DEPLOYMENT_URL/health" > /dev/null; then
        print_success "Health check passed"
    else
        print_error "Health check failed"
        return 1
    fi
}

# Main deployment process
main() {
    echo "🚀 KYB Platform Beta Deployment on Railway"
    echo "=========================================="
    
    check_prerequisites
    create_railway_config
    deploy_to_railway
    get_deployment_url
    health_check
    
    echo ""
    echo "🎉 Railway deployment completed successfully!"
    echo ""
    echo "📋 Next Steps:"
    echo "1. Review SHAREABLE_LINKS_RAILWAY.md for access information"
    echo "2. Send the shareable links to your beta testers"
    echo "3. Monitor the deployment using Railway dashboard"
    echo "4. Collect feedback through the web interface"
    echo ""
    echo "🌐 Web Interface: $DEPLOYMENT_URL"
    echo "📊 Railway Dashboard: https://railway.app/project/[PROJECT_ID]"
    echo ""
}

# Run main function
main "$@"
