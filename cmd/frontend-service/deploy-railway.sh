#!/bin/bash

# Railway Deployment Script for Frontend Service
# This script handles deployment to Railway

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVICE_NAME="frontend-service"
PROJECT_NAME="kyb-platform"
ENVIRONMENT="${ENVIRONMENT:-production}"

echo -e "${BLUE}🚀 Railway Deployment Script for Frontend Service${NC}"
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
        echo "Please install it with: npm install -g @railway/cli"
        exit 1
    fi
    
    # Check if logged in
    if ! railway whoami &> /dev/null; then
        echo -e "${YELLOW}⚠️  Not logged in to Railway${NC}"
        echo "Please run: railway login"
        exit 1
    fi
    
    # Check if we're in the right directory
    if [ ! -f "Dockerfile" ] || [ ! -f "railway.json" ]; then
        echo -e "${RED}❌ Not in the correct directory${NC}"
        echo "Please run from cmd/frontend-service/"
        exit 1
    fi
    
    echo -e "${GREEN}✅ Prerequisites check passed${NC}"
}

# Function to deploy
deploy() {
    echo -e "${YELLOW}📦 Deploying to Railway...${NC}"
    
    # Link to Railway project if not already linked
    if [ ! -f ".railway/project.json" ]; then
        echo -e "${YELLOW}🔗 Linking to Railway project...${NC}"
        railway link
    fi
    
    # Deploy
    echo -e "${YELLOW}🚀 Starting deployment...${NC}"
    railway up --detach
    
    echo -e "${GREEN}✅ Deployment initiated${NC}"
    echo -e "${BLUE}📊 Check deployment status with: railway status${NC}"
    echo -e "${BLUE}📋 View logs with: railway logs${NC}"
}

# Main execution
main() {
    check_prerequisites
    deploy
    
    echo -e "${GREEN}🎉 Deployment process completed!${NC}"
}

# Run main function
main

