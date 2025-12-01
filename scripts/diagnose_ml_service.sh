#!/bin/bash
# Diagnostic script for ML service integration issues
# This script helps identify why the ML service is not working with testing

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo ""
echo "🔍 ML Service Integration Diagnostic"
echo "===================================="
echo ""

# Step 1: Check environment variable
echo "📋 Step 1: Checking PYTHON_ML_SERVICE_URL environment variable..."
echo ""

if [ -z "$PYTHON_ML_SERVICE_URL" ]; then
    echo -e "${RED}❌ PYTHON_ML_SERVICE_URL is NOT set${NC}"
    echo ""
    echo "   This is why ML service is not being used in tests."
    echo ""
    echo "   To fix:"
    echo "   export PYTHON_ML_SERVICE_URL=\"http://localhost:8000\""
    echo ""
    ML_URL=""
else
    echo -e "${GREEN}✅ PYTHON_ML_SERVICE_URL is set${NC}"
    echo "   Value: $PYTHON_ML_SERVICE_URL"
    echo ""
    ML_URL="$PYTHON_ML_SERVICE_URL"
fi

# Step 2: Check if service is running
echo "📋 Step 2: Checking if Python ML service is running..."
echo ""

if [ -z "$ML_URL" ]; then
    # Try default localhost URL
    ML_URL="http://localhost:8000"
    echo "   Using default URL: $ML_URL"
fi

# Remove trailing slash
ML_URL=$(echo "$ML_URL" | sed 's|/$||')

# Test /ping endpoint
echo -n "   Testing /ping endpoint... "
PING_RESPONSE=$(curl -s -w "\n%{http_code}" -m 5 "$ML_URL/ping" 2>&1 || echo "ERROR")
HTTP_CODE=$(echo "$PING_RESPONSE" | tail -n 1)
PING_BODY=$(echo "$PING_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✅ OK${NC}"
    echo "   Response: $PING_BODY"
    PING_WORKING=true
elif [ "$HTTP_CODE" = "ERROR" ] || [ -z "$HTTP_CODE" ]; then
    echo -e "${RED}❌ Connection failed${NC}"
    echo "   Error: $PING_BODY"
    PING_WORKING=false
else
    echo -e "${YELLOW}⚠️  Unexpected status: $HTTP_CODE${NC}"
    echo "   Response: $PING_BODY"
    PING_WORKING=false
fi

# Test /health endpoint
echo -n "   Testing /health endpoint... "
HEALTH_RESPONSE=$(curl -s -w "\n%{http_code}" -m 5 "$ML_URL/health" 2>&1 || echo "ERROR")
HTTP_CODE=$(echo "$HEALTH_RESPONSE" | tail -n 1)
HEALTH_BODY=$(echo "$HEALTH_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✅ OK${NC}"
    echo "   Response: $HEALTH_BODY" | head -c 200
    echo ""
    HEALTH_WORKING=true
elif [ "$HTTP_CODE" = "ERROR" ] || [ -z "$HTTP_CODE" ]; then
    echo -e "${RED}❌ Connection failed${NC}"
    echo "   Error: $HEALTH_BODY"
    HEALTH_WORKING=false
else
    echo -e "${YELLOW}⚠️  Unexpected status: $HTTP_CODE${NC}"
    echo "   Response: $HEALTH_BODY"
    HEALTH_WORKING=false
fi

# Step 3: Check if service process is running (for localhost)
echo ""
echo "📋 Step 3: Checking for Python ML service process..."
echo ""

if echo "$ML_URL" | grep -q "localhost\|127.0.0.1"; then
    # Check for Python processes
    PYTHON_PROCESSES=$(ps aux | grep -E "python.*app\.py|uvicorn.*app" | grep -v grep || true)
    
    if [ -n "$PYTHON_PROCESSES" ]; then
        echo -e "${GREEN}✅ Python ML service process found${NC}"
        echo "$PYTHON_PROCESSES" | head -n 3 | while read line; do
            echo "   $line"
        done
    else
        echo -e "${YELLOW}⚠️  No Python ML service process found${NC}"
        echo "   Service may be running in Docker or on a remote server"
    fi
else
    echo "   Service URL is not localhost, skipping process check"
fi

# Step 4: Check test binary
echo ""
echo "📋 Step 4: Checking test binary..."
echo ""

if [ -f "bin/comprehensive_accuracy_test" ]; then
    echo -e "${GREEN}✅ Test binary exists${NC}"
    echo "   Path: bin/comprehensive_accuracy_test"
    
    # Check if it's executable
    if [ -x "bin/comprehensive_accuracy_test" ]; then
        echo -e "${GREEN}✅ Test binary is executable${NC}"
    else
        echo -e "${YELLOW}⚠️  Test binary is not executable${NC}"
        echo "   Fix: chmod +x bin/comprehensive_accuracy_test"
    fi
else
    echo -e "${YELLOW}⚠️  Test binary not found${NC}"
    echo "   Build it with: go build -o bin/comprehensive_accuracy_test ./cmd/comprehensive_accuracy_test"
fi

# Step 5: Check Python dependencies (if localhost)
echo ""
echo "📋 Step 5: Checking Python dependencies..."
echo ""

if echo "$ML_URL" | grep -q "localhost\|127.0.0.1"; then
    if [ -d "python_ml_service/venv" ]; then
        # Check if torch is installed in venv
        if python_ml_service/venv/bin/python -c "import torch" 2>/dev/null; then
            echo -e "${GREEN}✅ Python dependencies installed (torch found)${NC}"
            DEPS_INSTALLED=true
        else
            echo -e "${YELLOW}⚠️  Python dependencies not installed${NC}"
            echo "   Missing: torch (and likely other packages)"
            echo "   Fix: cd python_ml_service && source venv/bin/activate && pip install -r requirements.txt"
            DEPS_INSTALLED=false
        fi
    else
        echo "   Virtual environment not found (python_ml_service/venv)"
        DEPS_INSTALLED=false
    fi
else
    echo "   Service URL is not localhost, skipping dependency check"
    DEPS_INSTALLED=true  # Assume OK for remote services
fi

# Step 6: Summary and recommendations
echo ""
echo "📊 Summary and Recommendations"
echo "=============================="
echo ""

ISSUES=0

if [ -z "$PYTHON_ML_SERVICE_URL" ]; then
    echo -e "${RED}❌ Issue 1: PYTHON_ML_SERVICE_URL not set${NC}"
    ISSUES=$((ISSUES + 1))
fi

if [ "$PING_WORKING" = false ]; then
    echo -e "${RED}❌ Issue 2: /ping endpoint not accessible${NC}"
    ISSUES=$((ISSUES + 1))
fi

if [ "$HEALTH_WORKING" = false ]; then
    echo -e "${RED}❌ Issue 3: /health endpoint not accessible${NC}"
    ISSUES=$((ISSUES + 1))
fi

if [ "$DEPS_INSTALLED" = false ]; then
    echo -e "${RED}❌ Issue 4: Python dependencies not installed${NC}"
    ISSUES=$((ISSUES + 1))
fi

if [ $ISSUES -eq 0 ]; then
    echo -e "${GREEN}✅ All checks passed! ML service should work.${NC}"
    echo ""
    echo "   To run tests with ML:"
    echo "   ./bin/comprehensive_accuracy_test -verbose -output accuracy_report_ml.json"
else
    echo ""
    echo -e "${YELLOW}⚠️  Found $ISSUES issue(s)${NC}"
    echo ""
    echo "🔧 Recommended Fixes:"
    echo ""
    
    if [ -z "$PYTHON_ML_SERVICE_URL" ]; then
        echo "1. Set PYTHON_ML_SERVICE_URL:"
        echo "   export PYTHON_ML_SERVICE_URL=\"http://localhost:8000\""
        echo ""
    fi
    
    if [ "$DEPS_INSTALLED" = false ]; then
        echo "2. Install Python dependencies:"
        echo "   cd python_ml_service"
        echo "   source venv/bin/activate"
        echo "   pip install -r requirements.txt"
        echo "   (This may take 5-10 minutes)"
        echo ""
    fi
    
    if [ "$PING_WORKING" = false ] || [ "$HEALTH_WORKING" = false ]; then
        echo "3. Start Python ML service:"
        echo "   cd python_ml_service"
        echo "   source venv/bin/activate"
        echo "   python app.py"
        echo ""
        echo "   Or use the automated script:"
        echo "   ./scripts/run_ml_accuracy_tests.sh"
        echo ""
    fi
    
    echo "4. After fixing, re-run this diagnostic:"
    echo "   ./scripts/diagnose_ml_service.sh"
    echo ""
fi

# Step 6: Quick test (if everything looks good)
if [ $ISSUES -eq 0 ]; then
    echo "🧪 Quick Integration Test"
    echo "======================="
    echo ""
    echo "Testing Go service initialization..."
    
    # Create a simple test program
    cat > /tmp/test_ml_init.go << 'EOF'
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"
    "kyb-platform/internal/machine_learning/infrastructure"
)

func main() {
    endpoint := os.Getenv("PYTHON_ML_SERVICE_URL")
    if endpoint == "" {
        fmt.Println("❌ PYTHON_ML_SERVICE_URL not set")
        os.Exit(1)
    }
    
    logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
    service := infrastructure.NewPythonMLService(endpoint, logger)
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    if err := service.Initialize(ctx); err != nil {
        fmt.Printf("❌ Initialization failed: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Println("✅ ML service initialized successfully!")
}
EOF
    
    if go run /tmp/test_ml_init.go 2>&1 | grep -q "✅"; then
        echo -e "${GREEN}✅ Go service can initialize ML service${NC}"
    else
        echo -e "${YELLOW}⚠️  Go service initialization test failed${NC}"
        echo "   This may indicate a code issue, not just configuration"
    fi
    
    rm -f /tmp/test_ml_init.go
    echo ""
fi

echo "✅ Diagnostic complete"
echo ""

