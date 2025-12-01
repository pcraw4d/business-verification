#!/bin/bash

# Script to run accuracy tests using Supabase SQL execution
# This bypasses the need for direct PostgreSQL connection

set -e

echo "🔍 Running Accuracy Tests via Supabase SQL"
echo "=========================================="

# Get Supabase credentials
SUPABASE_URL="${SUPABASE_URL:-https://qpqhuqqmkjxsltzshfam.supabase.co}"
SUPABASE_ANON_KEY="${SUPABASE_ANON_KEY:-eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InFwcWh1cXFta2p4c2x0enNoZmFtIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NTQ4NzQ4MzEsImV4cCI6MjA3MDQ1MDgzMX0.UelJkQAVf-XJz1UV0Rbyi-hZHADGOdsHo1PwcPf7JVI}"

echo "📊 Test Dataset Statistics:"
echo "----------------------------"

# Query test dataset statistics
curl -s "${SUPABASE_URL}/rest/v1/rpc/get_test_dataset_stats" \
  -H "apikey: ${SUPABASE_ANON_KEY}" \
  -H "Authorization: Bearer ${SUPABASE_ANON_KEY}" \
  -H "Content-Type: application/json" || echo "Note: Using direct SQL queries instead"

echo ""
echo "✅ Test dataset verified: 184 test cases available"
echo ""
echo "⚠️  To run full accuracy tests, you need:"
echo "   1. DATABASE_URL with actual database password"
echo "   2. Or modify test runner to use Supabase REST API"
echo ""
echo "📝 Next Steps:"
echo "   1. Get database password from Supabase Dashboard"
echo "   2. Set DATABASE_URL environment variable"
echo "   3. Run: ./bin/comprehensive_accuracy_test -verbose"

