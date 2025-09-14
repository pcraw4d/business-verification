#!/bin/bash

# Railway Fixed Deployment Script
echo "🚀 Starting Railway Fixed Deployment..."

# Check if we're in the right directory
if [ ! -f "Dockerfile.production" ]; then
    echo "❌ Error: Dockerfile.production not found"
    exit 1
fi

# Check if railway CLI is available
if ! command -v railway &> /dev/null; then
    echo "❌ Error: Railway CLI not found"
    exit 1
fi

echo "📦 Building Docker image locally to test..."
docker build -f Dockerfile.production -t kyb-platform-fixed .

if [ $? -ne 0 ]; then
    echo "❌ Docker build failed"
    exit 1
fi

echo "✅ Docker build successful"

echo "🚀 Deploying to Railway..."
railway up --detach

if [ $? -ne 0 ]; then
    echo "❌ Railway deployment failed"
    exit 1
fi

echo "✅ Deployment initiated"

echo "⏳ Waiting for deployment to complete..."
sleep 30

echo "📋 Checking deployment logs..."
railway logs

echo "🔍 Testing endpoints..."
echo "Health endpoint:"
curl -s https://shimmering-comfort-production.up.railway.app/health | jq .

echo "Classification endpoint test:"
curl -s -X POST https://shimmering-comfort-production.up.railway.app/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name":"Test Company","description":"A test company"}' | jq .

echo "Merchant API test:"
curl -s https://shimmering-comfort-production.up.railway.app/api/v1/merchants | jq .

echo "🎉 Deployment verification complete!"
