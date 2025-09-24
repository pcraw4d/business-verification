#!/bin/bash

# Set environment variables
export SUPABASE_URL="https://qpqhuqqmkjxsltzshfam.supabase.co"
export SUPABASE_ANON_KEY="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InFwcWh1cXFta2p4c2x0enNoZmFtIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NTQ4NzQ4MzEsImV4cCI6MjA3MDQ1MDgzMX0.UelJkQAVf-XJz1UV0Rbyi-hZHADGOdsHo1PwcPf7JVI"
export SUPABASE_SERVICE_ROLE_KEY="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InFwcWh1cXFta2p4c2x0enNoZmFtIiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTc1NDg3NDgzMSwiZXhwIjoyMDcwNDUwODMxfQ.sIm3w7Ad2kLv08whNBrzdP42nz0s4dsLpvUiYDSwArw"
export SUPABASE_JWT_SECRET="zIJXeH0Z1RsbTFEoGgp6aaknV1jGsMjIkNLN77bhfsQ7Mk7OOIzzRzFPRGITX3dg7OX6RemHFcYr7ytolG76yw=="

# Print environment variables for debugging
echo "Environment variables set:"
echo "SUPABASE_URL: $SUPABASE_URL"
echo "SUPABASE_ANON_KEY: ${SUPABASE_ANON_KEY:0:20}..."
echo "SUPABASE_SERVICE_ROLE_KEY: ${SUPABASE_SERVICE_ROLE_KEY:0:20}..."
echo "SUPABASE_JWT_SECRET: ${SUPABASE_JWT_SECRET:0:20}..."

# Change to the correct directory and run the server
cd cmd/api-enhanced
echo "Starting server..."
./kyb-platform-enhanced
