#!/bin/bash

# KYB Platform - Readiness Check Script
# Checks if all components are ready for merchant-details API endpoints

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔍 KYB Platform - Readiness Check${NC}"
echo "=================================="
echo ""

# Check if migration file exists
echo -e "${BLUE}Checking migration file...${NC}"
if [ -f "internal/database/migrations/010_add_async_risk_assessment_columns.sql" ]; then
    echo -e "${GREEN}✅ Migration file exists${NC}"
else
    echo -e "${RED}❌ Migration file not found${NC}"
    echo "   Expected: internal/database/migrations/010_add_async_risk_assessment_columns.sql"
    exit 1
fi

# Check if handlers exist
echo ""
echo -e "${BLUE}Checking handler files...${NC}"
if [ -f "internal/api/handlers/merchant_analytics_handler.go" ]; then
    echo -e "${GREEN}✅ Merchant analytics handler exists${NC}"
else
    echo -e "${RED}❌ Merchant analytics handler not found${NC}"
fi

if [ -f "internal/api/handlers/async_risk_assessment_handler.go" ]; then
    echo -e "${GREEN}✅ Async risk assessment handler exists${NC}"
else
    echo -e "${RED}❌ Async risk assessment handler not found${NC}"
fi

# Check if routes exist
echo ""
echo -e "${BLUE}Checking route files...${NC}"
if [ -f "internal/api/routes/merchant_routes.go" ]; then
    echo -e "${GREEN}✅ Merchant routes exist${NC}"
else
    echo -e "${RED}❌ Merchant routes not found${NC}"
fi

if [ -f "internal/api/routes/risk_routes.go" ]; then
    echo -e "${GREEN}✅ Risk routes exist${NC}"
else
    echo -e "${RED}❌ Risk routes not found${NC}"
fi

# Check if services exist
echo ""
echo -e "${BLUE}Checking service files...${NC}"
if [ -f "internal/services/merchant_analytics_service.go" ]; then
    echo -e "${GREEN}✅ Merchant analytics service exists${NC}"
else
    echo -e "${RED}❌ Merchant analytics service not found${NC}"
fi

if [ -f "internal/services/risk_assessment_service.go" ]; then
    echo -e "${GREEN}✅ Risk assessment service exists${NC}"
else
    echo -e "${RED}❌ Risk assessment service not found${NC}"
fi

# Check if repositories exist
echo ""
echo -e "${BLUE}Checking repository files...${NC}"
if [ -f "internal/database/merchant_analytics_repository.go" ]; then
    echo -e "${GREEN}✅ Merchant analytics repository exists${NC}"
else
    echo -e "${RED}❌ Merchant analytics repository not found${NC}"
fi

if [ -f "internal/database/risk_assessment_repository.go" ]; then
    echo -e "${GREEN}✅ Risk assessment repository exists${NC}"
else
    echo -e "${RED}❌ Risk assessment repository not found${NC}"
fi

# Check if tests exist
echo ""
echo -e "${BLUE}Checking test files...${NC}"
if [ -f "test/e2e/merchant_analytics_api_test.go" ]; then
    echo -e "${GREEN}✅ E2E tests exist${NC}"
else
    echo -e "${YELLOW}⚠️  E2E tests not found${NC}"
fi

if [ -f "test/integration/risk_assessment_integration_test.go" ]; then
    echo -e "${GREEN}✅ Integration tests exist${NC}"
else
    echo -e "${YELLOW}⚠️  Integration tests not found${NC}"
fi

# Check if API documentation exists
echo ""
echo -e "${BLUE}Checking API documentation...${NC}"
if [ -f "api/openapi/merchant-details-api-spec.yaml" ]; then
    echo -e "${GREEN}✅ OpenAPI specification exists${NC}"
else
    echo -e "${YELLOW}⚠️  OpenAPI specification not found${NC}"
fi

if [ -f "docs/async-routes-integration-guide.md" ]; then
    echo -e "${GREEN}✅ Integration guide exists${NC}"
else
    echo -e "${YELLOW}⚠️  Integration guide not found${NC}"
fi

# Try to compile handlers (if Go is available)
echo ""
echo -e "${BLUE}Checking compilation...${NC}"
if command -v go &> /dev/null; then
    if go build -o /dev/null ./internal/api/handlers/merchant_analytics_handler.go 2>/dev/null; then
        echo -e "${GREEN}✅ Merchant analytics handler compiles${NC}"
    else
        echo -e "${RED}❌ Merchant analytics handler has compilation errors${NC}"
    fi
    
    if go build -o /dev/null ./internal/api/handlers/async_risk_assessment_handler.go 2>/dev/null; then
        echo -e "${GREEN}✅ Async risk assessment handler compiles${NC}"
    else
        echo -e "${RED}❌ Async risk assessment handler has compilation errors${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  Go not found, skipping compilation check${NC}"
fi

# Summary
echo ""
echo "=================================="
echo -e "${BLUE}📋 Next Steps:${NC}"
echo ""
echo "1. ${YELLOW}Run database migration:${NC}"
echo "   psql \$DATABASE_URL -f internal/database/migrations/010_add_async_risk_assessment_columns.sql"
echo ""
echo "2. ${YELLOW}Register routes in your main server application${NC}"
echo "   See: docs/async-routes-integration-guide.md"
echo ""
echo "3. ${YELLOW}Test endpoints using Postman/Insomnia${NC}"
echo "   Collections: tests/api/merchant-details/"
echo ""
echo "4. ${YELLOW}Run integration tests:${NC}"
echo "   go test -v -tags=integration ./test/integration/risk_assessment_integration_test.go"
echo ""
echo -e "${GREEN}✅ Readiness check complete!${NC}"

