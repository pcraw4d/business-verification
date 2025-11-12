#!/bin/bash

# Setup script to help integrate new merchant analytics and async risk assessment routes
# This script helps identify where to add route registration in your server

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🔍 KYB Platform - Route Integration Helper${NC}"
echo "=============================================="
echo ""

echo -e "${BLUE}This script helps you identify where to add route registration.${NC}"
echo ""

# Find main.go files that might be the main server
echo -e "${YELLOW}Found potential main server files:${NC}"
find cmd -name "main.go" -type f 2>/dev/null | head -5

echo ""
echo -e "${BLUE}Next steps:${NC}"
echo ""
echo "1. Identify your main API server file (usually in cmd/)"
echo "2. Look for where routes are registered (search for 'HandleFunc' or 'Handle')"
echo "3. Add the route registration code from:"
echo "   - docs/async-routes-integration-guide.md"
echo "   - internal/api/routes/integration_example.go"
echo ""
echo -e "${GREEN}Example locations to check:${NC}"
echo "  - cmd/railway-server/main.go (setupRoutes function)"
echo "  - cmd/frontend-service/main.go (setupRoutes function)"
echo "  - services/api-gateway/cmd/main.go (router setup)"
echo ""
echo -e "${YELLOW}For detailed integration instructions, see:${NC}"
echo "  docs/async-routes-integration-guide.md"
echo "  docs/next-steps-operational-checklist.md"

