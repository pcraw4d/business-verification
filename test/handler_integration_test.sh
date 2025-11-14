#!/bin/bash
# Test script for merchant portfolio handler integration

set -e

echo "🧪 Testing Merchant Portfolio Handler Integration"
echo "=================================================="

# Load test environment
source .env.test

# Check if database is accessible
echo ""
echo "1. Checking database connection..."
psql "$DATABASE_URL" -c "SELECT 1;" > /dev/null 2>&1 && echo "✅ Database connection OK" || { echo "❌ Database connection failed"; exit 1; }

# Setup test data
echo ""
echo "2. Setting up test data..."
go test -v ./test/database/repository_methods_test.go ./test/database/database_operations_test.go -run TestRepositoryMethodsWithData > /dev/null 2>&1
echo "✅ Test data setup complete"

# Check if server is running
echo ""
echo "3. Checking if server is running on port 8080..."
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "✅ Server is running"
else
    echo "⚠️  Server is not running. Please start it with:"
    echo "   PORT=8080 DATABASE_URL=\"\$DATABASE_URL\" go run cmd/railway-server/main.go"
    echo ""
    echo "   Or in background:"
    echo "   PORT=8080 DATABASE_URL=\"\$DATABASE_URL\" go run cmd/railway-server/main.go &"
    exit 1
fi

# Test the analytics endpoint
echo ""
echo "4. Testing GET /api/v1/merchants/analytics endpoint..."
RESPONSE=$(curl -s -X GET "http://localhost:8080/api/v1/merchants/analytics" \
    -H "Authorization: Bearer test-token-local")

# Check if response contains expected fields
if echo "$RESPONSE" | grep -q "total_merchants"; then
    echo "✅ Response contains 'total_merchants'"
else
    echo "❌ Response does not contain 'total_merchants'"
    echo "Response: $RESPONSE"
    exit 1
fi

if echo "$RESPONSE" | grep -q "portfolio_distribution"; then
    echo "✅ Response contains 'portfolio_distribution'"
else
    echo "❌ Response does not contain 'portfolio_distribution'"
    echo "Response: $RESPONSE"
    exit 1
fi

if echo "$RESPONSE" | grep -q "risk_distribution"; then
    echo "✅ Response contains 'risk_distribution'"
else
    echo "❌ Response does not contain 'risk_distribution'"
    echo "Response: $RESPONSE"
    exit 1
fi

# Pretty print the response
echo ""
echo "5. Analytics Response:"
echo "$RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE"

echo ""
echo "✅ All tests passed!"
echo ""
echo "Summary:"
echo "- Database connection: ✅"
echo "- Test data setup: ✅"
echo "- Server running: ✅"
echo "- Analytics endpoint: ✅"
echo "- Response format: ✅"

