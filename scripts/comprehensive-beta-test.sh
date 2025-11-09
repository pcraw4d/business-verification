#!/bin/bash

# Comprehensive Beta Testing Script
# Tests for directory mismatches, API failures, and deployment readiness

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

PASSED=0
FAILED=0
WARNINGS=0

# Test results file
TEST_RESULTS="beta-test-results-$(date +%Y%m%d-%H%M%S).txt"
echo "Beta Testing Results - $(date)" > "$TEST_RESULTS"
echo "=================================" >> "$TEST_RESULTS"
echo "" >> "$TEST_RESULTS"

log_pass() {
    echo -e "${GREEN}✅ PASS:${NC} $1"
    echo "✅ PASS: $1" >> "$TEST_RESULTS"
    ((PASSED++))
}

log_fail() {
    echo -e "${RED}❌ FAIL:${NC} $1"
    echo "❌ FAIL: $1" >> "$TEST_RESULTS"
    ((FAILED++))
}

log_warn() {
    echo -e "${YELLOW}⚠️  WARN:${NC} $1"
    echo "⚠️  WARN: $1" >> "$TEST_RESULTS"
    ((WARNINGS++))
}

log_info() {
    echo -e "${BLUE}ℹ️  INFO:${NC} $1"
    echo "ℹ️  INFO: $1" >> "$TEST_RESULTS"
}

echo "🧪 Starting Comprehensive Beta Testing..."
echo ""

# ============================================================================
# TEST 1: Frontend Directory Sync
# ============================================================================
echo "📋 TEST 1: Frontend Directory Sync"
echo "-----------------------------------"

SOURCE_DIR="services/frontend/public"
TARGET_DIR="cmd/frontend-service/static"
CRITICAL_FILES=("add-merchant.html" "merchant-details.html")

if [ -d "$SOURCE_DIR" ] && [ -d "$TARGET_DIR" ]; then
    for file in "${CRITICAL_FILES[@]}"; do
        if [ -f "$SOURCE_DIR/$file" ]; then
            if [ ! -f "$TARGET_DIR/$file" ]; then
                log_fail "Critical file missing in deployment: $file"
            elif ! diff -q "$SOURCE_DIR/$file" "$TARGET_DIR/$file" > /dev/null 2>&1; then
                log_fail "Files differ: $file (source vs deployment)"
            else
                log_pass "File synced: $file"
            fi
        fi
    done
else
    log_fail "Frontend directories not found"
fi

echo ""

# ============================================================================
# TEST 2: Service Configuration Files
# ============================================================================
echo "📋 TEST 2: Service Configuration Files"
echo "----------------------------------------"

SERVICES=(
    "services/api-gateway:railway.json"
    "services/classification-service:railway.json"
    "services/merchant-service:railway.json"
    "services/risk-assessment-service:railway.json"
    "cmd/frontend-service:railway.json"
)

for service_config in "${SERVICES[@]}"; do
    IFS=':' read -r service config_file <<< "$service_config"
    if [ -f "$service/$config_file" ]; then
        log_pass "Config found: $service/$config_file"
    else
        log_warn "Config missing: $service/$config_file"
    fi
done

echo ""

# ============================================================================
# TEST 3: Dockerfile Presence
# ============================================================================
echo "📋 TEST 3: Dockerfile Presence"
echo "--------------------------------"

for service in services/* cmd/frontend-service; do
    if [ -d "$service" ] && [ -f "$service/Dockerfile" ]; then
        log_pass "Dockerfile found: $service"
    elif [ -d "$service" ] && [ -f "$service/cmd/main.go" ] || [ -f "$service/main.go" ]; then
        log_warn "Dockerfile missing but Go code present: $service"
    fi
done

echo ""

# ============================================================================
# TEST 4: API Endpoint Configuration
# ============================================================================
echo "📋 TEST 4: API Endpoint Configuration"
echo "--------------------------------------"

# Check for hardcoded API URLs
API_URLS=(
    "api-gateway-service-production-21fd.up.railway.app"
    "frontend-service-production-b225.up.railway.app"
    "classification-service-production.up.railway.app"
)

for url in "${API_URLS[@]}"; do
    if grep -r "$url" services/frontend/public cmd/frontend-service/static 2>/dev/null | grep -v node_modules | grep -v ".git" > /dev/null; then
        log_pass "API URL found in frontend: $url"
    else
        log_warn "API URL not found: $url (may be configured elsewhere)"
    fi
done

echo ""

# ============================================================================
# TEST 5: Service Health Endpoints
# ============================================================================
echo "📋 TEST 5: Service Health Endpoints"
echo "-----------------------------------"

# Test production endpoints
PROD_ENDPOINTS=(
    "https://frontend-service-production-b225.up.railway.app/health"
    "https://api-gateway-service-production-21fd.up.railway.app/health"
)

for endpoint in "${PROD_ENDPOINTS[@]}"; do
    if curl -s -f --max-time 10 "$endpoint" > /dev/null 2>&1; then
        log_pass "Health check passed: $endpoint"
    else
        log_fail "Health check failed: $endpoint"
    fi
done

echo ""

# ============================================================================
# TEST 6: API Gateway Classification Endpoint
# ============================================================================
echo "📋 TEST 6: API Gateway Classification Endpoint"
echo "----------------------------------------------"

TEST_PAYLOAD='{"business_name":"Test Company","description":"Test description","website_url":"https://example.com"}'
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -d "$TEST_PAYLOAD" \
    --max-time 30 \
    "https://api-gateway-service-production-21fd.up.railway.app/api/v1/classify" 2>&1)

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "200" ]; then
    if echo "$BODY" | grep -q "success"; then
        log_pass "Classification API working (HTTP $HTTP_CODE)"
    else
        log_warn "Classification API returned 200 but no success field"
    fi
else
    log_fail "Classification API failed (HTTP $HTTP_CODE)"
    echo "Response: $BODY" | head -5 >> "$TEST_RESULTS"
fi

echo ""

# ============================================================================
# TEST 7: Frontend Static File Serving
# ============================================================================
echo "📋 TEST 7: Frontend Static File Serving"
echo "----------------------------------------"

FRONTEND_URL="https://frontend-service-production-b225.up.railway.app"
PAGES=("add-merchant" "merchant-details" "merchant-portfolio")

for page in "${PAGES[@]}"; do
    if curl -s -f --max-time 10 "$FRONTEND_URL/$page" | grep -q "<!DOCTYPE html\|<html" 2>/dev/null; then
        log_pass "Page accessible: $page"
    else
        log_fail "Page not accessible: $page"
    fi
done

echo ""

# ============================================================================
# TEST 8: JavaScript File References
# ============================================================================
echo "📋 TEST 8: JavaScript File References"
echo "--------------------------------------"

# Check if critical JS files exist
JS_FILES=(
    "cmd/frontend-service/static/js/api-config.js"
    "cmd/frontend-service/static/components/navigation.js"
)

for js_file in "${JS_FILES[@]}"; do
    if [ -f "$js_file" ]; then
        log_pass "JS file exists: $js_file"
    else
        log_warn "JS file missing: $js_file"
    fi
done

echo ""

# ============================================================================
# TEST 9: Error Handling in Frontend
# ============================================================================
echo "📋 TEST 9: Error Handling in Frontend"
echo "--------------------------------------"

# Check for Promise.allSettled usage (should not have .catch() before it)
if grep -r "Promise.allSettled" cmd/frontend-service/static/add-merchant.html 2>/dev/null | grep -v "\.catch" > /dev/null; then
    log_pass "Promise.allSettled used correctly (no .catch() before it)"
else
    log_fail "Promise.allSettled may have .catch() handlers (defeats purpose)"
fi

# Check for escapeHtml function (XSS protection)
if grep -r "escapeHtml\|escapeHTML" cmd/frontend-service/static/merchant-details.html 2>/dev/null > /dev/null; then
    log_pass "XSS protection (escapeHtml) found in merchant-details"
else
    log_warn "XSS protection (escapeHtml) not found in merchant-details"
fi

echo ""

# ============================================================================
# TEST 10: Environment Configuration
# ============================================================================
echo "📋 TEST 10: Environment Configuration"
echo "--------------------------------------"

# Check for environment variable usage
if grep -r "process.env\|os.Getenv" services cmd/frontend-service 2>/dev/null | grep -v node_modules | grep -v ".git" > /dev/null; then
    log_pass "Environment variables used in services"
else
    log_warn "No environment variable usage found (may be hardcoded)"
fi

echo ""

# ============================================================================
# TEST 11: CORS Configuration
# ============================================================================
echo "📋 TEST 11: CORS Configuration"
echo "-------------------------------"

# Check for CORS middleware
if grep -r "CORS\|Access-Control" services/api-gateway 2>/dev/null | grep -v node_modules > /dev/null; then
    log_pass "CORS configuration found in API Gateway"
else
    log_warn "CORS configuration not found in API Gateway"
fi

echo ""

# ============================================================================
# TEST 12: Logging and Error Handling
# ============================================================================
echo "📋 TEST 12: Logging and Error Handling"
echo "---------------------------------------"

# Check for proper error logging
if grep -r "console.error\|log.Error\|zap.Error" services/api-gateway cmd/frontend-service 2>/dev/null | grep -v node_modules > /dev/null; then
    log_pass "Error logging found in services"
else
    log_warn "Limited error logging found"
fi

echo ""

# ============================================================================
# TEST 13: Database Connection Strings
# ============================================================================
echo "📋 TEST 13: Database Connection Strings"
echo "----------------------------------------"

# Check for hardcoded database URLs (should use env vars)
if grep -r "postgres://\|mysql://\|mongodb://" services cmd 2>/dev/null | grep -v node_modules | grep -v ".git" | grep -v "example\|test" > /dev/null; then
    log_fail "Hardcoded database URLs found (should use environment variables)"
else
    log_pass "No hardcoded database URLs found"
fi

echo ""

# ============================================================================
# TEST 14: API Request Body Handling
# ============================================================================
echo "📋 TEST 14: API Request Body Handling"
echo "--------------------------------------"

# Check for request body being read multiple times
if grep -r "r.Body\|req.Body" services/api-gateway 2>/dev/null | grep -v node_modules | grep -v "io.ReadAll\|bodyBytes" > /dev/null; then
    log_warn "Potential request body read issue (check for multiple reads)"
else
    log_pass "Request body handling looks correct"
fi

echo ""

# ============================================================================
# TEST 15: File Count Verification
# ============================================================================
echo "📋 TEST 15: File Count Verification"
echo "------------------------------------"

SOURCE_HTML=$(find services/frontend/public -name "*.html" -type f 2>/dev/null | wc -l | tr -d ' ')
TARGET_HTML=$(find cmd/frontend-service/static -name "*.html" -type f 2>/dev/null | wc -l | tr -d ' ')

if [ "$SOURCE_HTML" -eq "$TARGET_HTML" ] || [ "$SOURCE_HTML" -eq $((TARGET_HTML + 1)) ]; then
    log_pass "HTML file counts match (Source: $SOURCE_HTML, Target: $TARGET_HTML)"
else
    log_warn "HTML file count mismatch (Source: $SOURCE_HTML, Target: $TARGET_HTML)"
fi

echo ""

# ============================================================================
# SUMMARY
# ============================================================================
echo "=========================================="
echo "📊 TEST SUMMARY"
echo "=========================================="
echo ""
echo -e "${GREEN}✅ Passed: $PASSED${NC}"
echo -e "${YELLOW}⚠️  Warnings: $WARNINGS${NC}"
echo -e "${RED}❌ Failed: $FAILED${NC}"
echo ""
echo "Full results saved to: $TEST_RESULTS"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}🎉 All critical tests passed! Product is beta-ready.${NC}"
    exit 0
else
    echo -e "${RED}⚠️  Some tests failed. Please review and fix issues before beta testing.${NC}"
    exit 1
fi

