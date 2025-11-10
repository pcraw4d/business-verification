#!/bin/bash

# Script to authenticate with Railway CLI and fetch logs for failing services
# Run this script in your terminal to get the logs

set -e

echo "🔐 Railway CLI Authentication and Log Retrieval"
echo "================================================"
echo ""

# Check if Railway CLI is installed
if ! command -v railway &> /dev/null; then
    echo "❌ Railway CLI not found. Installing..."
    npm install -g @railway/cli || brew install railway
fi

# Step 1: Authenticate
echo "📝 Step 1: Authenticating with Railway..."
echo "   Run this command and follow the prompts:"
echo "   railway login --browserless"
echo ""
read -p "Press Enter after you've authenticated..."

# Step 2: Link to project
echo "📝 Step 2: Linking to project..."
cd "$(dirname "$0")"
railway link || echo "⚠️  Project may already be linked"

# Step 3: Get logs for risk-assessment-service
echo ""
echo "📋 Fetching logs for risk-assessment-service..."
echo "================================================"
railway logs --service risk-assessment-service --tail 200 > risk-assessment-service-logs.txt 2>&1 || echo "⚠️  Could not fetch logs for risk-assessment-service"
cat risk-assessment-service-logs.txt

# Step 4: Get logs for classification-service
echo ""
echo "📋 Fetching logs for classification-service..."
echo "=============================================="
railway logs --service classification-service --tail 200 > classification-service-logs.txt 2>&1 || echo "⚠️  Could not fetch logs for classification-service"
cat classification-service-logs.txt

# Step 5: Get deployment status
echo ""
echo "📊 Deployment Status..."
echo "======================"
railway status

echo ""
echo "✅ Logs saved to:"
echo "   - risk-assessment-service-logs.txt"
echo "   - classification-service-logs.txt"
echo ""
echo "📝 Next steps:"
echo "   1. Review the logs above"
echo "   2. Check for error messages"
echo "   3. Share the error messages if you need help fixing them"

