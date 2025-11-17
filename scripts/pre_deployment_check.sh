#!/bin/bash

# Pre-Deployment Verification Script
# Runs all checks before deployment to ensure readiness

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

ERRORS=0
WARNINGS=0

echo -e "${BLUE}════════════════════════════════════════${NC}"
echo -e "${BLUE}  Pre-Deployment Verification${NC}"
echo -e "${BLUE}════════════════════════════════════════${NC}"
echo ""

# Check 1: Configuration Validation
echo -e "${GREEN}Check 1: Configuration Validation${NC}"
if [ -f "./scripts/validate_config.sh" ]; then
    if ./scripts/validate_config.sh; then
        echo -e "  ${GREEN}✅ Configuration valid${NC}"
    else
        echo -e "  ${RED}❌ Configuration validation failed${NC}"
        ((ERRORS++))
    fi
else
    echo -e "  ${YELLOW}⚠️  Validation script not found${NC}"
    ((WARNINGS++))
fi

# Check 2: Database Connection (if DATABASE_URL is set)
echo ""
echo -e "${GREEN}Check 2: Database Connection${NC}"
if [ -n "$DATABASE_URL" ]; then
    if [ -f "./scripts/test_database_connection.sh" ]; then
        if ./scripts/test_database_connection.sh > /dev/null 2>&1; then
            echo -e "  ${GREEN}✅ Database connection successful${NC}"
        else
            echo -e "  ${RED}❌ Database connection failed${NC}"
            ((ERRORS++))
        fi
    else
        echo -e "  ${YELLOW}⚠️  Connection test script not found${NC}"
        ((WARNINGS++))
    fi
    
    # Check schema
    if [ -f "./scripts/verify_database_schema.sh" ]; then
        if ./scripts/verify_database_schema.sh > /dev/null 2>&1; then
            echo -e "  ${GREEN}✅ Database schema verified${NC}"
        else
            echo -e "  ${YELLOW}⚠️  Database schema verification failed (may need migration)${NC}"
            ((WARNINGS++))
        fi
    fi
else
    echo -e "  ${YELLOW}⚠️  DATABASE_URL not set (using in-memory mode)${NC}"
    ((WARNINGS++))
fi

# Check 3: Code Quality
echo ""
echo -e "${GREEN}Check 3: Code Quality${NC}"

# Check if Go code compiles
if command -v go &> /dev/null; then
    if go build ./cmd/railway-server/... > /dev/null 2>&1; then
        echo -e "  ${GREEN}✅ Code compiles successfully${NC}"
    else
        echo -e "  ${RED}❌ Code compilation failed${NC}"
        ((ERRORS++))
    fi
else
    echo -e "  ${YELLOW}⚠️  Go not found, skipping compilation check${NC}"
    ((WARNINGS++))
fi

# Check 4: Test Suite
echo ""
echo -e "${GREEN}Check 4: Test Suite${NC}"
if [ -f "./test/restoration_tests.sh" ]; then
    echo -e "  ${GREEN}✅ Test suite available${NC}"
    echo -e "  ${YELLOW}⚠️  Run tests manually: ./test/restoration_tests.sh${NC}"
else
    echo -e "  ${YELLOW}⚠️  Test suite not found${NC}"
    ((WARNINGS++))
fi

# Check 5: Documentation
echo ""
echo -e "${GREEN}Check 5: Documentation${NC}"
DOCS=(
    "docs/PRODUCTION_ENV_SETUP.md"
    "docs/DEPLOYMENT_CHECKLIST.md"
    "docs/RESTORED_ENDPOINTS_DOCUMENTATION.md"
)

MISSING_DOCS=0
for doc in "${DOCS[@]}"; do
    if [ -f "$doc" ]; then
        echo -e "  ${GREEN}✅ $(basename $doc) exists${NC}"
    else
        echo -e "  ${YELLOW}⚠️  $(basename $doc) missing${NC}"
        ((MISSING_DOCS++))
    fi
done

if [ $MISSING_DOCS -gt 0 ]; then
    ((WARNINGS++))
fi

# Check 6: Migration Files
echo ""
echo -e "${GREEN}Check 6: Migration Files${NC}"
if [ -f "internal/database/migrations/012_create_risk_thresholds_table.sql" ]; then
    echo -e "  ${GREEN}✅ Migration file exists${NC}"
else
    echo -e "  ${YELLOW}⚠️  Migration file not found${NC}"
    ((WARNINGS++))
fi

# Check 7: Git Status
echo ""
echo -e "${GREEN}Check 7: Git Status${NC}"
if command -v git &> /dev/null; then
    if [ -d ".git" ]; then
        UNCOMMITTED=$(git status --porcelain | wc -l)
        if [ "$UNCOMMITTED" -eq 0 ]; then
            echo -e "  ${GREEN}✅ No uncommitted changes${NC}"
        else
            echo -e "  ${YELLOW}⚠️  $UNCOMMITTED uncommitted file(s)${NC}"
            ((WARNINGS++))
        fi
        
        # Check if on main branch
        BRANCH=$(git rev-parse --abbrev-ref HEAD)
        if [ "$BRANCH" = "main" ] || [ "$BRANCH" = "master" ]; then
            echo -e "  ${GREEN}✅ On $BRANCH branch${NC}"
        else
            echo -e "  ${YELLOW}⚠️  On $BRANCH branch (not main/master)${NC}"
            ((WARNINGS++))
        fi
    else
        echo -e "  ${YELLOW}⚠️  Not a git repository${NC}"
    fi
else
    echo -e "  ${YELLOW}⚠️  Git not found${NC}"
fi

# Summary
echo ""
echo -e "${BLUE}════════════════════════════════════════${NC}"
if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo -e "${GREEN}✅ All checks passed! Ready for deployment.${NC}"
    echo -e "${BLUE}════════════════════════════════════════${NC}"
    exit 0
elif [ $ERRORS -eq 0 ]; then
    echo -e "${GREEN}✅ Critical checks passed${NC}"
    echo -e "${YELLOW}⚠️  $WARNINGS warning(s) - Review before deployment${NC}"
    echo -e "${BLUE}════════════════════════════════════════${NC}"
    exit 0
else
    echo -e "${RED}❌ Deployment blocked: $ERRORS error(s), $WARNINGS warning(s)${NC}"
    echo -e "${BLUE}════════════════════════════════════════${NC}"
    exit 1
fi

