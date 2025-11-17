#!/bin/bash

# Database Connection Test Script
# Tests database connectivity and connection pooling

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if DATABASE_URL is set
if [ -z "$DATABASE_URL" ]; then
    echo -e "${RED}❌ ERROR: DATABASE_URL environment variable is not set${NC}"
    echo "Please set DATABASE_URL before running this script:"
    echo "  export DATABASE_URL='postgres://user:pass@host:port/dbname'"
    exit 1
fi

echo -e "${BLUE}════════════════════════════════════════${NC}"
echo -e "${BLUE}  Database Connection Test${NC}"
echo -e "${BLUE}════════════════════════════════════════${NC}"
echo ""

DB_URL="$DATABASE_URL"

# Test 1: Basic connection
echo -e "${GREEN}Test 1: Basic Connection${NC}"
if command -v psql &> /dev/null; then
    if psql "$DB_URL" -c "SELECT version();" > /dev/null 2>&1; then
        VERSION=$(psql "$DB_URL" -tAc "SELECT version();" 2>/dev/null)
        echo -e "  ${GREEN}✅ Connection successful${NC}"
        echo -e "  ${GREEN}   Database: ${VERSION:0:50}...${NC}"
    else
        echo -e "  ${RED}❌ Connection failed${NC}"
        exit 1
    fi
else
    echo -e "  ${YELLOW}⚠️  psql not available, skipping direct connection test${NC}"
fi

# Test 2: Test via API endpoint (if server is running)
echo ""
echo -e "${GREEN}Test 2: API Endpoint Connection${NC}"
API_URL="${API_URL:-http://localhost:8080}"

if curl -s -f "${API_URL}/health/detailed" > /dev/null 2>&1; then
    HEALTH_RESPONSE=$(curl -s "${API_URL}/health/detailed")
    if echo "$HEALTH_RESPONSE" | grep -q "database" || echo "$HEALTH_RESPONSE" | grep -q "postgres"; then
        echo -e "  ${GREEN}✅ API endpoint accessible${NC}"
        
        # Check database status in health response
        if echo "$HEALTH_RESPONSE" | grep -qi "database.*ok\|database.*healthy\|postgres.*ok"; then
            echo -e "  ${GREEN}✅ Database health check passed${NC}"
        else
            echo -e "  ${YELLOW}⚠️  Database status unclear in health response${NC}"
        fi
    else
        echo -e "  ${YELLOW}⚠️  API accessible but database status not found${NC}"
    fi
else
    echo -e "  ${YELLOW}⚠️  API endpoint not accessible (server may not be running)${NC}"
    echo "  Start server with: go run cmd/railway-server/main.go"
fi

# Test 3: Connection pool settings (if psql available)
if command -v psql &> /dev/null; then
    echo ""
    echo -e "${GREEN}Test 3: Connection Pool Information${NC}"
    
    MAX_CONNECTIONS=$(psql "$DB_URL" -tAc "SHOW max_connections;" 2>/dev/null || echo "unknown")
    CURRENT_CONNECTIONS=$(psql "$DB_URL" -tAc "SELECT count(*) FROM pg_stat_activity WHERE datname = current_database();" 2>/dev/null || echo "unknown")
    
    echo -e "  ${GREEN}✅ Max connections: $MAX_CONNECTIONS${NC}"
    echo -e "  ${GREEN}✅ Current connections: $CURRENT_CONNECTIONS${NC}"
    
    if [ "$MAX_CONNECTIONS" != "unknown" ] && [ "$CURRENT_CONNECTIONS" != "unknown" ]; then
        USAGE_PERCENT=$((CURRENT_CONNECTIONS * 100 / MAX_CONNECTIONS))
        if [ "$USAGE_PERCENT" -lt 50 ]; then
            echo -e "  ${GREEN}✅ Connection pool usage: ${USAGE_PERCENT}% (healthy)${NC}"
        elif [ "$USAGE_PERCENT" -lt 80 ]; then
            echo -e "  ${YELLOW}⚠️  Connection pool usage: ${USAGE_PERCENT}% (moderate)${NC}"
        else
            echo -e "  ${RED}⚠️  Connection pool usage: ${USAGE_PERCENT}% (high)${NC}"
        fi
    fi
fi

# Test 4: Query performance
echo ""
echo -e "${GREEN}Test 4: Query Performance${NC}"
if command -v psql &> /dev/null; then
    START_TIME=$(date +%s%N)
    QUERY_RESULT=$(psql "$DB_URL" -tAc "SELECT COUNT(*) FROM risk_thresholds;" 2>/dev/null || echo "ERROR")
    END_TIME=$(date +%s%N)
    
    if [ "$QUERY_RESULT" != "ERROR" ]; then
        DURATION_MS=$(( (END_TIME - START_TIME) / 1000000 ))
        echo -e "  ${GREEN}✅ Query successful${NC}"
        echo -e "  ${GREEN}✅ Threshold count: $QUERY_RESULT${NC}"
        echo -e "  ${GREEN}✅ Query time: ${DURATION_MS}ms${NC}"
        
        if [ "$DURATION_MS" -lt 100 ]; then
            echo -e "  ${GREEN}✅ Performance: Excellent (< 100ms)${NC}"
        elif [ "$DURATION_MS" -lt 500 ]; then
            echo -e "  ${GREEN}✅ Performance: Good (< 500ms)${NC}"
        else
            echo -e "  ${YELLOW}⚠️  Performance: Slow (> 500ms)${NC}"
        fi
    else
        echo -e "  ${YELLOW}⚠️  Query failed (table may not exist)${NC}"
        echo "  Run migration: psql \$DATABASE_URL -f internal/database/migrations/012_create_risk_thresholds_table.sql"
    fi
fi

# Summary
echo ""
echo -e "${BLUE}════════════════════════════════════════${NC}"
echo -e "${GREEN}✅ Connection tests completed${NC}"
echo -e "${BLUE}════════════════════════════════════════${NC}"

