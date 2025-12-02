#!/bin/bash

# Classification Service Configuration Verification Script
# Verifies all required environment variables are set for website scraping optimizations

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "=========================================="
echo "Classification Service Configuration Verification"
echo "=========================================="
echo ""

# Required environment variables for website scraping optimizations
REQUIRED_VARS=(
    "ENABLE_FAST_PATH_SCRAPING"
    "CLASSIFICATION_MAX_CONCURRENT_PAGES"
    "CLASSIFICATION_CRAWL_DELAY_MS"
    "CLASSIFICATION_FAST_PATH_MAX_PAGES"
    "CLASSIFICATION_WEBSITE_SCRAPING_TIMEOUT"
)

# Optional but recommended variables
OPTIONAL_VARS=(
    "REDIS_ENABLED"
    "REDIS_URL"
    "ENABLE_WEBSITE_CONTENT_CACHE"
    "CACHE_ENABLED"
    "CACHE_TTL"
)

# Expected default values
EXPECTED_VALUES=(
    "ENABLE_FAST_PATH_SCRAPING=true"
    "CLASSIFICATION_MAX_CONCURRENT_PAGES=3"
    "CLASSIFICATION_CRAWL_DELAY_MS=500"
    "CLASSIFICATION_FAST_PATH_MAX_PAGES=8"
    "CLASSIFICATION_WEBSITE_SCRAPING_TIMEOUT=5s"
)

echo "Checking required environment variables..."
echo ""

# Check required variables
missing_vars=()
incorrect_values=()

for var in "${REQUIRED_VARS[@]}"; do
    value=$(printenv "$var" 2>/dev/null || echo "")
    
    if [ -z "$value" ]; then
        echo "${RED}❌ $var${NC} - NOT SET"
        missing_vars+=("$var")
    else
        echo "${GREEN}✅ $var${NC} - Set to: $value"
        
        # Check if value matches expected (for known defaults)
        case "$var" in
            "ENABLE_FAST_PATH_SCRAPING")
                if [ "$value" != "true" ] && [ "$value" != "True" ] && [ "$value" != "TRUE" ]; then
                    echo "   ${YELLOW}⚠️  Warning: Expected 'true' but got '$value'${NC}"
                    incorrect_values+=("$var")
                fi
                ;;
            "CLASSIFICATION_MAX_CONCURRENT_PAGES")
                if ! [[ "$value" =~ ^[0-9]+$ ]] || [ "$value" -lt 1 ] || [ "$value" -gt 10 ]; then
                    echo "   ${YELLOW}⚠️  Warning: Expected number between 1-10, got '$value'${NC}"
                    incorrect_values+=("$var")
                fi
                ;;
            "CLASSIFICATION_CRAWL_DELAY_MS")
                if ! [[ "$value" =~ ^[0-9]+$ ]] || [ "$value" -lt 100 ] || [ "$value" -gt 5000 ]; then
                    echo "   ${YELLOW}⚠️  Warning: Expected number between 100-5000ms, got '$value'${NC}"
                    incorrect_values+=("$var")
                fi
                ;;
            "CLASSIFICATION_FAST_PATH_MAX_PAGES")
                if ! [[ "$value" =~ ^[0-9]+$ ]] || [ "$value" -lt 1 ] || [ "$value" -gt 20 ]; then
                    echo "   ${YELLOW}⚠️  Warning: Expected number between 1-20, got '$value'${NC}"
                    incorrect_values+=("$var")
                fi
                ;;
        esac
    fi
done

echo ""
echo "Checking optional but recommended variables..."
echo ""

# Check optional variables
for var in "${OPTIONAL_VARS[@]}"; do
    value=$(printenv "$var" 2>/dev/null || echo "")
    
    if [ -z "$value" ]; then
        echo "${YELLOW}⚠️  $var${NC} - NOT SET (optional but recommended)"
    else
        echo "${GREEN}✅ $var${NC} - Set to: ${value:0:50}${#value:50:+...}"
    fi
done

echo ""
echo "=========================================="
echo "Summary"
echo "=========================================="
echo ""

if [ ${#missing_vars[@]} -eq 0 ] && [ ${#incorrect_values[@]} -eq 0 ]; then
    echo "${GREEN}✅ All required configuration variables are set correctly!${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Monitor production logs for performance metrics"
    echo "2. Run performance tests to verify improvements"
    exit 0
else
    if [ ${#missing_vars[@]} -gt 0 ]; then
        echo "${RED}❌ Missing required variables:${NC}"
        for var in "${missing_vars[@]}"; do
            echo "   - $var"
        done
        echo ""
    fi
    
    if [ ${#incorrect_values[@]} -gt 0 ]; then
        echo "${YELLOW}⚠️  Variables with unexpected values:${NC}"
        for var in "${incorrect_values[@]}"; do
            echo "   - $var"
        done
        echo ""
    fi
    
    echo "To set these variables in Railway:"
    echo "1. Go to Railway Dashboard → Classification Service → Variables"
    echo "2. Add the missing variables with recommended values:"
    echo ""
    for expected in "${EXPECTED_VALUES[@]}"; do
        echo "   $expected"
    done
    echo ""
    exit 1
fi

