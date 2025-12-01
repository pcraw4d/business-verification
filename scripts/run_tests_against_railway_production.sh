#!/bin/bash
# Run accuracy tests against Railway production services
# This script configures environment to use Railway production endpoints

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo ""
echo "🚂 Running Tests Against Railway Production"
echo "==========================================="
echo ""

# Load Railway environment variables
if [ -f "railway.env" ]; then
    echo "📋 Loading Railway environment variables from railway.env..."
    set -a
    source railway.env
    set +a
    echo -e "${GREEN}✅ Environment variables loaded${NC}"
elif [ -f "railway-essential.env" ]; then
    echo "📋 Loading Railway environment variables from railway-essential.env..."
    set -a
    source railway-essential.env
    set +a
    echo -e "${GREEN}✅ Environment variables loaded${NC}"
else
    echo -e "${YELLOW}⚠️  No railway.env or railway-essential.env found${NC}"
    echo "   Using environment variables from current shell"
fi

# Verify required variables
echo ""
echo "🔍 Verifying Required Environment Variables..."
echo ""

MISSING_VARS=0

if [ -z "$SUPABASE_URL" ]; then
    echo -e "${RED}❌ SUPABASE_URL not set${NC}"
    MISSING_VARS=$((MISSING_VARS + 1))
else
    echo -e "${GREEN}✅ SUPABASE_URL: $SUPABASE_URL${NC}"
fi

if [ -z "$SUPABASE_ANON_KEY" ]; then
    echo -e "${RED}❌ SUPABASE_ANON_KEY not set${NC}"
    MISSING_VARS=$((MISSING_VARS + 1))
else
    echo -e "${GREEN}✅ SUPABASE_ANON_KEY: [REDACTED]${NC}"
fi

if [ -z "$DATABASE_URL" ] && [ -z "$TEST_DATABASE_URL" ]; then
    echo -e "${YELLOW}⚠️  DATABASE_URL or TEST_DATABASE_URL not set${NC}"
    echo "   Will try to construct from SUPABASE_URL"
    if [ -n "$SUPABASE_URL" ]; then
        # Extract project ref from Supabase URL
        PROJECT_REF=$(echo "$SUPABASE_URL" | sed 's|https://||' | sed 's|\.supabase\.co||')
        if [ -n "$SUPABASE_SERVICE_ROLE_KEY" ]; then
            # Try to construct DATABASE_URL (user needs to provide password)
            echo "   You may need to set DATABASE_URL manually"
        fi
    fi
else
    if [ -n "$DATABASE_URL" ]; then
        echo -e "${GREEN}✅ DATABASE_URL: [REDACTED]${NC}"
    fi
    if [ -n "$TEST_DATABASE_URL" ]; then
        echo -e "${GREEN}✅ TEST_DATABASE_URL: [REDACTED]${NC}"
    fi
fi

if [ $MISSING_VARS -gt 0 ]; then
    echo ""
    echo -e "${RED}❌ Missing required environment variables${NC}"
    echo ""
    echo "Please set the following variables:"
    echo "  export SUPABASE_URL='https://your-project.supabase.co'"
    echo "  export SUPABASE_ANON_KEY='your_anon_key'"
    echo "  export DATABASE_URL='postgresql://...' (or TEST_DATABASE_URL)"
    echo ""
    exit 1
fi

# Check for Python ML service URL
echo ""
echo "🔍 Checking Python ML Service Configuration..."
echo ""

# Try to get Python ML service URL from Railway
if [ -z "$PYTHON_ML_SERVICE_URL" ]; then
    echo -e "${YELLOW}⚠️  PYTHON_ML_SERVICE_URL not set${NC}"
    echo ""
    echo "Checking Railway for Python ML service..."
    
    # Try to get from Railway CLI if available
    if command -v railway &> /dev/null; then
        echo "   Using Railway CLI to find Python ML service..."
        PYTHON_ML_URL=$(railway variables --service python-ml-service 2>/dev/null | grep "RAILWAY_PUBLIC_DOMAIN" | awk -F'=' '{print $2}' | tr -d ' ' || echo "")
        
        if [ -n "$PYTHON_ML_URL" ]; then
            PYTHON_ML_SERVICE_URL="https://$PYTHON_ML_URL"
            echo -e "${GREEN}✅ Found Python ML service: $PYTHON_ML_SERVICE_URL${NC}"
            export PYTHON_ML_SERVICE_URL
        else
            # Try common Railway URL pattern
            echo "   Trying common Railway URL pattern..."
            PYTHON_ML_SERVICE_URL="https://python-ml-service-production.up.railway.app"
            echo "   Using: $PYTHON_ML_SERVICE_URL"
            export PYTHON_ML_SERVICE_URL
        fi
    else
        # Use default Railway URL pattern
        PYTHON_ML_SERVICE_URL="https://python-ml-service-production.up.railway.app"
        echo "   Using default: $PYTHON_ML_SERVICE_URL"
        echo "   (You can override with: export PYTHON_ML_SERVICE_URL='your-url')"
        export PYTHON_ML_SERVICE_URL
    fi
else
    echo -e "${GREEN}✅ PYTHON_ML_SERVICE_URL: $PYTHON_ML_SERVICE_URL${NC}"
fi

# Verify Python ML service is accessible
echo ""
echo "🔍 Verifying Python ML Service is Accessible..."
echo ""

if curl -s -f -m 5 "$PYTHON_ML_SERVICE_URL/ping" > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Python ML service is accessible${NC}"
    PING_RESPONSE=$(curl -s "$PYTHON_ML_SERVICE_URL/ping")
    echo "   Response: $PING_RESPONSE"
else
    echo -e "${RED}❌ Python ML service is NOT accessible${NC}"
    echo "   URL: $PYTHON_ML_SERVICE_URL"
    echo "   This may indicate the service is not deployed or not running"
    echo ""
    read -p "Continue anyway? (y/n): " CONTINUE
    if [ "$CONTINUE" != "y" ]; then
        exit 1
    fi
fi

# Verify health endpoint
if curl -s -f -m 5 "$PYTHON_ML_SERVICE_URL/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Python ML service health check passed${NC}"
else
    echo -e "${YELLOW}⚠️  Python ML service health check failed${NC}"
fi

# Build test binary if needed
echo ""
echo "📦 Building Test Binary..."
echo ""

if [ ! -f "bin/comprehensive_accuracy_test" ]; then
    echo "   Building comprehensive_accuracy_test..."
    go build -o bin/comprehensive_accuracy_test ./cmd/comprehensive_accuracy_test
    echo -e "${GREEN}✅ Test binary built${NC}"
else
    echo -e "${GREEN}✅ Test binary already exists${NC}"
fi

# Run tests
echo ""
echo "🧪 Running Accuracy Tests Against Railway Production..."
echo "======================================================"
echo ""
echo "Configuration:"
echo "  - Supabase URL: $SUPABASE_URL"
echo "  - Python ML Service: $PYTHON_ML_SERVICE_URL"
echo "  - Environment: Production (Railway)"
echo ""

# Generate output filename with timestamp
OUTPUT_FILE="accuracy_report_railway_production_$(date +%Y%m%d_%H%M%S).json"

# Run the tests
echo "Running tests..."
echo ""

./bin/comprehensive_accuracy_test \
    -verbose \
    -output "$OUTPUT_FILE" \
    -supabase-url "$SUPABASE_URL" \
    -supabase-key "$SUPABASE_ANON_KEY" \
    ${SUPABASE_SERVICE_ROLE_KEY:+-supabase-service-key "$SUPABASE_SERVICE_ROLE_KEY"} \
    ${DATABASE_URL:+-database-url "$DATABASE_URL"} \
    ${TEST_DATABASE_URL:+-database-url "$TEST_DATABASE_URL"}

TEST_EXIT_CODE=$?

echo ""
echo "======================================================"

if [ $TEST_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✅ Tests completed successfully!${NC}"
    echo ""
    echo "📊 Results saved to: $OUTPUT_FILE"
    echo ""
    echo "To view results:"
    echo "  cat $OUTPUT_FILE | jq"
    echo ""
else
    echo -e "${RED}❌ Tests failed with exit code: $TEST_EXIT_CODE${NC}"
    echo ""
    echo "Check the output above for details"
    exit $TEST_EXIT_CODE
fi

