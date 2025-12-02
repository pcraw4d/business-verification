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
    # Use set -a to automatically export all variables
    set -a
    source railway.env
    set +a
    # Explicitly export key variables to ensure they're available
    export SUPABASE_URL
    export SUPABASE_ANON_KEY
    export SUPABASE_SERVICE_ROLE_KEY
    export DATABASE_URL
    export PYTHON_ML_SERVICE_URL
    echo -e "${GREEN}✅ Environment variables loaded and exported${NC}"
elif [ -f "railway-essential.env" ]; then
    echo "📋 Loading Railway environment variables from railway-essential.env..."
    set -a
    source railway-essential.env
    set +a
    export SUPABASE_URL
    export SUPABASE_ANON_KEY
    export SUPABASE_SERVICE_ROLE_KEY
    export DATABASE_URL
    export PYTHON_ML_SERVICE_URL
    echo -e "${GREEN}✅ Environment variables loaded and exported${NC}"
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

# Check if DATABASE_URL has placeholder and try to get real one from Railway
if [ -n "$DATABASE_URL" ] && echo "$DATABASE_URL" | grep -q "\[YOUR_PASSWORD_HERE\]\|\[YOUR_PASSWORD\]\|placeholder\|your_password"; then
    echo -e "${YELLOW}⚠️  DATABASE_URL contains placeholder${NC}"
    echo "   Attempting to get real DATABASE_URL from Railway..."
    
    # Try to get from Railway CLI
    if command -v railway &> /dev/null; then
        RAILWAY_DB_URL=$(railway variables 2>/dev/null | grep -i "DATABASE_URL" | head -1 | awk -F'=' '{print $2}' | tr -d ' ' || echo "")
        if [ -n "$RAILWAY_DB_URL" ] && ! echo "$RAILWAY_DB_URL" | grep -q "\[YOUR_PASSWORD_HERE\]\|\[YOUR_PASSWORD\]\|placeholder"; then
            DATABASE_URL="$RAILWAY_DB_URL"
            echo -e "${GREEN}✅ Got DATABASE_URL from Railway${NC}"
        else
            echo "   Railway DATABASE_URL also has placeholder or not found"
            DATABASE_URL=""
        fi
    else
        DATABASE_URL=""
    fi
fi

if [ -z "$DATABASE_URL" ] && [ -z "$TEST_DATABASE_URL" ]; then
    echo -e "${YELLOW}⚠️  DATABASE_URL or TEST_DATABASE_URL not set${NC}"
    echo "   Attempting to construct from Supabase connection..."
    
    if [ -n "$SUPABASE_URL" ]; then
        # Extract project ref from Supabase URL
        PROJECT_REF=$(echo "$SUPABASE_URL" | sed 's|https://||' | sed 's|\.supabase\.co||')
        
        # Try to get password from Railway or environment
        # For Supabase, we can use the service role key or connection pooling
        # Use Supabase connection pooling URL format
        if [ -n "$SUPABASE_SERVICE_ROLE_KEY" ]; then
            # Use Supabase connection pooling (port 6543) or direct connection
            # Note: This requires the actual database password
            echo "   ⚠️  Cannot auto-construct DATABASE_URL without password"
            echo "   Please set DATABASE_URL or TEST_DATABASE_URL manually"
            echo "   Format: postgresql://postgres:[PASSWORD]@db.$PROJECT_REF.supabase.co:5432/postgres"
        fi
    fi
else
    if [ -n "$DATABASE_URL" ] && ! echo "$DATABASE_URL" | grep -q "\[YOUR_PASSWORD_HERE\]\|placeholder"; then
        echo -e "${GREEN}✅ DATABASE_URL: [REDACTED]${NC}"
    elif [ -n "$TEST_DATABASE_URL" ]; then
        echo -e "${GREEN}✅ TEST_DATABASE_URL: [REDACTED]${NC}"
        DATABASE_URL="$TEST_DATABASE_URL"
    else
        echo -e "${RED}❌ DATABASE_URL contains invalid placeholder${NC}"
        MISSING_VARS=$((MISSING_VARS + 1))
    fi
fi

if [ $MISSING_VARS -gt 0 ]; then
    echo ""
    echo -e "${YELLOW}⚠️  Some environment variables have issues${NC}"
    echo ""
    
    if [ -z "$DATABASE_URL" ] || echo "$DATABASE_URL" | grep -q "\[YOUR_PASSWORD_HERE\]\|\[YOUR_PASSWORD\]\|placeholder"; then
        echo "DATABASE_URL issue detected. The test may still work using Supabase API."
        echo "If tests fail, you can:"
        echo "  1. Set DATABASE_URL manually: export DATABASE_URL='postgresql://postgres:[PASSWORD]@db.qpqhuqqmkjxsltzshfam.supabase.co:5432/postgres'"
        echo "  2. Or get it from Railway: railway variables | grep DATABASE_URL"
        echo ""
        read -p "Continue with tests anyway? (y/n): " CONTINUE_DB
        if [ "$CONTINUE_DB" != "y" ]; then
            exit 1
        fi
    else
        echo "Please set the following variables:"
        echo "  export SUPABASE_URL='https://your-project.supabase.co'"
        echo "  export SUPABASE_ANON_KEY='your_anon_key'"
        echo ""
        exit 1
    fi
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
            # Try known production URL (from documentation)
            echo "   Using known production URL from documentation..."
            PYTHON_ML_SERVICE_URL="https://python-ml-service-production-a6b8.up.railway.app"
            echo "   Using: $PYTHON_ML_SERVICE_URL"
            export PYTHON_ML_SERVICE_URL
        fi
        else
            # Try to get from railway.env file
            if [ -f "railway.env" ]; then
                ENV_URL=$(grep "^PYTHON_ML_SERVICE_URL=" railway.env | cut -d'=' -f2- | tr -d '"' | tr -d "'" || echo "")
                if [ -n "$ENV_URL" ]; then
                    PYTHON_ML_SERVICE_URL="$ENV_URL"
                    echo "   Found in railway.env: $PYTHON_ML_SERVICE_URL"
                    export PYTHON_ML_SERVICE_URL
                else
                    # Use known production URL
                    PYTHON_ML_SERVICE_URL="https://python-ml-service-production-a6b8.up.railway.app"
                    echo "   Using known production URL: $PYTHON_ML_SERVICE_URL"
                    export PYTHON_ML_SERVICE_URL
                fi
            else
                # Use known production URL
                PYTHON_ML_SERVICE_URL="https://python-ml-service-production-a6b8.up.railway.app"
                echo "   Using known production URL: $PYTHON_ML_SERVICE_URL"
                echo "   (You can override with: export PYTHON_ML_SERVICE_URL='your-url')"
                export PYTHON_ML_SERVICE_URL
            fi
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
    echo -e "${YELLOW}⚠️  Continuing anyway (tests will use fallback classifier)${NC}"
    # Auto-continue for automated runs
    CONTINUE="y"
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

