#!/bin/bash

# Verify Railway Configuration
# Checks that all services have proper configuration files

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASSED=0
FAILED=0

check_file() {
    if [ -f "$1" ]; then
        echo -e "${GREEN}✅${NC} $1"
        ((PASSED++))
        return 0
    else
        echo -e "${RED}❌${NC} $1 (MISSING)"
        ((FAILED++))
        return 1
    fi
}

check_dockerfile() {
    if [ -f "$1" ]; then
        # Check if Dockerfile has wget for health checks
        if grep -q "wget" "$1" 2>/dev/null; then
            echo -e "${GREEN}✅${NC} $1 (has wget)"
        else
            echo -e "${YELLOW}⚠️${NC} $1 (missing wget - may cause health check failures)"
        fi
        ((PASSED++))
        return 0
    else
        echo -e "${RED}❌${NC} $1 (MISSING)"
        ((FAILED++))
        return 1
    fi
}

echo "🔍 Verifying Railway Configuration..."
echo ""

echo "📋 Railway Configuration Files"
echo "-------------------------------"
check_file "cmd/frontend-service/railway.json"
check_file "services/api-gateway/railway.json"
check_file "services/classification-service/railway.json"
check_file "services/merchant-service/railway.json"
check_file "services/risk-assessment-service/railway.json"
check_file "cmd/business-intelligence-gateway/railway.json"
check_file "cmd/pipeline-service/railway.json"
check_file "cmd/service-discovery/railway.json"
echo ""

echo "🐳 Dockerfiles"
echo "--------------"
check_dockerfile "cmd/frontend-service/Dockerfile"
check_dockerfile "services/api-gateway/Dockerfile"
check_dockerfile "services/classification-service/Dockerfile"
check_dockerfile "services/merchant-service/Dockerfile"
check_dockerfile "services/risk-assessment-service/Dockerfile"
check_dockerfile "cmd/business-intelligence-gateway/Dockerfile"
check_dockerfile "cmd/pipeline-service/Dockerfile"
check_dockerfile "cmd/service-discovery/Dockerfile"
echo ""

echo "📦 Go Modules"
echo "-------------"
check_file "cmd/frontend-service/go.mod"
check_file "services/api-gateway/go.mod"
check_file "services/classification-service/go.mod"
check_file "services/merchant-service/go.mod"
check_file "services/risk-assessment-service/go.mod"
check_file "cmd/business-intelligence-gateway/go.mod"
check_file "cmd/pipeline-service/go.mod"
check_file "cmd/service-discovery/go.mod"
echo ""

echo "📁 Critical Directories"
echo "----------------------"
if [ -d "cmd/frontend-service/static" ]; then
    echo -e "${GREEN}✅${NC} cmd/frontend-service/static (exists)"
    ((PASSED++))
else
    echo -e "${RED}❌${NC} cmd/frontend-service/static (MISSING)"
    ((FAILED++))
fi

echo ""
echo "=========================================="
echo "📊 Summary"
echo "=========================================="
echo -e "${GREEN}✅ Passed: $PASSED${NC}"
echo -e "${RED}❌ Failed: $FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All configuration files present!${NC}"
    exit 0
else
    echo -e "${RED}⚠️  Some configuration files are missing${NC}"
    exit 1
fi

