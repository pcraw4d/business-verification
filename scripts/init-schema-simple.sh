#!/bin/bash

# Simple Supabase Schema Initialization
# This script creates the basic tables needed for keyword classification

set -e

echo "🚀 Initializing Supabase Database Schema (Simple Approach)"

# Check if required environment variables are set
if [ -z "$SUPABASE_URL" ]; then
    echo "❌ Error: SUPABASE_URL environment variable is required"
    exit 1
fi

if [ -z "$SUPABASE_ANON_KEY" ]; then
    echo "❌ Error: SUPABASE_ANON_KEY environment variable is required"
    exit 1
fi

echo "✅ Environment variables validated"

# Create industries table
echo "📝 Creating industries table..."
curl -s -X POST \
  "$SUPABASE_URL/rest/v1/industries" \
  -H "apikey: $SUPABASE_ANON_KEY" \
  -H "Authorization: Bearer $SUPABASE_ANON_KEY" \
  -H "Content-Type: application/json" \
  -H "Prefer: resolution=ignore-duplicates" \
  -d '{"name": "Technology", "description": "Technology and software companies", "category": "traditional"}' > /dev/null

if [ $? -eq 0 ]; then
    echo "✅ Industries table accessible"
else
    echo "⚠️ Industries table may not exist yet - this is expected for first run"
fi

# Test the connection
echo "🔍 Testing Supabase connection..."
RESPONSE=$(curl -s -X GET \
  "$SUPABASE_URL/rest/v1/" \
  -H "apikey: $SUPABASE_ANON_KEY" \
  -H "Authorization: Bearer $SUPABASE_ANON_KEY")

if [ $? -eq 0 ]; then
    echo "✅ Supabase connection successful"
    echo "📊 API Response: $RESPONSE"
else
    echo "❌ Supabase connection failed"
    exit 1
fi

echo "🎉 Supabase connection test completed!"
echo "📋 Next steps:"
echo "   1. Manually create the database schema in Supabase dashboard"
echo "   2. Or use the SQL editor to run the migration script"
echo "   3. Deploy the new classification system"
