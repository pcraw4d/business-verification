#!/bin/bash

# Start Classification Service Script
# This script helps start the classification service with proper environment variables

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "🚀 Starting Classification Service"
echo ""

# Check if .env file exists in project root or service directory
ENV_FILE=""
if [ -f "$PROJECT_ROOT/.env" ]; then
    ENV_FILE="$PROJECT_ROOT/.env"
    echo "📄 Found .env file at project root"
elif [ -f "$SCRIPT_DIR/.env" ]; then
    ENV_FILE="$SCRIPT_DIR/.env"
    echo "📄 Found .env file in service directory"
elif [ -f "$PROJECT_ROOT/configs/development.env" ]; then
    ENV_FILE="$PROJECT_ROOT/configs/development.env"
    echo "📄 Found development.env config file"
    echo "⚠️  Note: This file has placeholder values. You'll need to set real Supabase credentials."
fi

# Load environment variables if file exists
if [ -n "$ENV_FILE" ]; then
    echo "Loading environment variables from $ENV_FILE"
    set -a
    source "$ENV_FILE"
    set +a
    echo ""
fi

# Check required environment variables
echo "Checking required environment variables..."

# Handle SUPABASE_API_KEY as alias for SUPABASE_ANON_KEY
if [ -n "$SUPABASE_API_KEY" ] && [ -z "$SUPABASE_ANON_KEY" ]; then
    export SUPABASE_ANON_KEY="$SUPABASE_API_KEY"
    echo "✅ Using SUPABASE_API_KEY as SUPABASE_ANON_KEY"
fi

MISSING_VARS=()

if [ -z "$SUPABASE_URL" ] || [ "$SUPABASE_URL" = "https://your-project.supabase.co" ]; then
    MISSING_VARS+=("SUPABASE_URL")
fi

if [ -z "$SUPABASE_ANON_KEY" ] || [ "$SUPABASE_ANON_KEY" = "your_supabase_anon_key" ] || [ "$SUPABASE_ANON_KEY" = "your-supabase-anon-key" ]; then
    MISSING_VARS+=("SUPABASE_ANON_KEY")
fi

if [ -z "$SUPABASE_SERVICE_ROLE_KEY" ] || [ "$SUPABASE_SERVICE_ROLE_KEY" = "your_supabase_service_role_key" ] || [ "$SUPABASE_SERVICE_ROLE_KEY" = "your-supabase-service-role-key" ]; then
    MISSING_VARS+=("SUPABASE_SERVICE_ROLE_KEY")
fi

if [ ${#MISSING_VARS[@]} -gt 0 ]; then
    echo "❌ Missing or placeholder environment variables:"
    for var in "${MISSING_VARS[@]}"; do
        echo "   - $var"
    done
    echo ""
    echo "Please set these environment variables:"
    echo ""
    echo "Option 1: Export in your shell:"
    echo "  export SUPABASE_URL='https://your-project.supabase.co'"
    echo "  export SUPABASE_ANON_KEY='your_anon_key'"
    echo "  export SUPABASE_SERVICE_ROLE_KEY='your_service_role_key'"
    echo ""
    echo "Option 2: Create a .env file in the project root:"
    echo "  SUPABASE_URL=https://your-project.supabase.co"
    echo "  SUPABASE_ANON_KEY=your_anon_key"
    echo "  SUPABASE_SERVICE_ROLE_KEY=your_service_role_key"
    echo ""
    echo "Option 3: Use Railway environment variables (if deployed):"
    echo "  railway variables"
    echo ""
    exit 1
fi

echo "✅ All required environment variables are set"
echo ""

# Set default port if not set
export PORT="${PORT:-8081}"
export HOST="${HOST:-0.0.0.0}"

echo "Configuration:"
echo "  Port: $PORT"
echo "  Host: $HOST"
echo "  Supabase URL: ${SUPABASE_URL:0:30}..."
echo ""

# Change to service directory
cd "$SCRIPT_DIR"

echo "Starting service..."
echo "Press Ctrl+C to stop"
echo ""

# Run the service
go run cmd/main.go
