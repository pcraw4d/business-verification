#!/bin/bash

# Load environment variables from .env.railway.full
echo "Loading Supabase environment variables..."
source .env.railway.full

# Export the variables to make them available to child processes
export SUPABASE_URL
export SUPABASE_ANON_KEY
export SUPABASE_SERVICE_ROLE_KEY
export SUPABASE_ENABLED

# Verify environment variables are loaded
echo "Verifying Supabase configuration..."
echo "SUPABASE_URL: $SUPABASE_URL"
echo "SUPABASE_ANON_KEY: ${SUPABASE_ANON_KEY:0:20}..."
echo "SUPABASE_SERVICE_ROLE_KEY: ${SUPABASE_SERVICE_ROLE_KEY:0:20}..."

# Check if all required variables are set
if [ -z "$SUPABASE_URL" ] || [ -z "$SUPABASE_ANON_KEY" ] || [ -z "$SUPABASE_SERVICE_ROLE_KEY" ]; then
    echo "❌ Error: Missing required Supabase environment variables"
    echo "Required: SUPABASE_URL, SUPABASE_ANON_KEY, SUPABASE_SERVICE_ROLE_KEY"
    exit 1
fi

echo "✅ All Supabase environment variables are set"
echo "🚀 Starting KYB Platform server with Supabase integration..."

# Start the server with environment variables
env SUPABASE_URL="$SUPABASE_URL" \
    SUPABASE_ANON_KEY="$SUPABASE_ANON_KEY" \
    SUPABASE_SERVICE_ROLE_KEY="$SUPABASE_SERVICE_ROLE_KEY" \
    SUPABASE_ENABLED="$SUPABASE_ENABLED" \
    ./kyb-platform
