#!/bin/bash

# Script to configure PYTHON_ML_SERVICE_URL for classification-service in Railway
# This enables enhanced classification with DistilBART

set -e

echo "🔧 Configuring Classification Service for Python ML Service Integration"
echo ""

# Check if Railway CLI is installed
if ! command -v railway &> /dev/null; then
    echo "❌ Railway CLI is not installed"
    echo "   Install it from: https://docs.railway.app/develop/cli"
    exit 1
fi

# Check if logged in
if ! railway whoami &> /dev/null; then
    echo "❌ Not logged in to Railway"
    echo "   Run: railway login"
    exit 1
fi

echo "📋 Current Railway project:"
railway status

echo ""
echo "🔍 Finding Python ML Service URL..."

# Try to get Python ML service URL from Railway
PYTHON_ML_SERVICE_URL=$(railway service --service python-ml-service --json 2>/dev/null | jq -r '.url // empty' || echo "")

if [ -z "$PYTHON_ML_SERVICE_URL" ]; then
    echo "⚠️  Could not automatically detect Python ML Service URL"
    echo ""
    echo "Please provide the Python ML Service URL:"
    echo "  1. Go to Railway Dashboard"
    echo "  2. Click on 'python-ml-service'"
    echo "  3. Copy the public URL (e.g., https://python-ml-service-production-xxx.up.railway.app)"
    echo ""
    read -p "Enter Python ML Service URL: " PYTHON_ML_SERVICE_URL
    
    if [ -z "$PYTHON_ML_SERVICE_URL" ]; then
        echo "❌ URL is required"
        exit 1
    fi
else
    echo "✅ Found Python ML Service URL: $PYTHON_ML_SERVICE_URL"
fi

echo ""
echo "🔧 Setting PYTHON_ML_SERVICE_URL for classification-service..."

# Set the environment variable
railway variables set PYTHON_ML_SERVICE_URL="$PYTHON_ML_SERVICE_URL" --service classification-service

if [ $? -eq 0 ]; then
    echo "✅ Successfully set PYTHON_ML_SERVICE_URL=$PYTHON_ML_SERVICE_URL"
    echo ""
    echo "📝 Next steps:"
    echo "   1. The classification service will automatically redeploy"
    echo "   2. Check logs for: '✅ Python ML Service initialized successfully'"
    echo "   3. Test enhanced classification with a request that includes a website URL"
else
    echo "❌ Failed to set environment variable"
    echo ""
    echo "💡 Manual setup:"
    echo "   1. Go to Railway Dashboard"
    echo "   2. Select 'classification-service'"
    echo "   3. Go to 'Variables' tab"
    echo "   4. Add: PYTHON_ML_SERVICE_URL = $PYTHON_ML_SERVICE_URL"
    exit 1
fi

echo ""
echo "✅ Configuration complete!"

