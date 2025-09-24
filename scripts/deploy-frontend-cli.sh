#!/bin/bash

# Deploy Frontend Service using Railway CLI
echo "🚀 Deploying Frontend Service using Railway CLI..."

# Check if Railway CLI is installed
if ! command -v railway &> /dev/null; then
    echo "❌ Railway CLI not found. Please install it first."
    exit 1
fi

# Check current service
echo "📋 Current service status:"
railway status

# Try to deploy with a smaller timeout
echo "🚀 Attempting deployment with timeout handling..."
timeout 300 railway up --detach

# Check if deployment was successful
if [ $? -eq 0 ]; then
    echo "✅ Frontend deployment successful!"
    
    # Get the deployment URL
    echo "🌐 Getting frontend deployment URL..."
    FRONTEND_URL=$(railway domain)
    
    echo "🎉 Frontend deployed successfully!"
    echo "🌐 Frontend URL: $FRONTEND_URL"
    echo "🔗 API Server URL: https://shimmering-comfort-production.up.railway.app"
    echo ""
    echo "📋 Architecture:"
    echo "   Frontend: $FRONTEND_URL (serves web UI)"
    echo "   Backend:  https://shimmering-comfort-production.up.railway.app (API server)"
    echo ""
    echo "🎯 Both services are now deployed separately on Railway!"
else
    echo "❌ Deployment timed out or failed."
    echo "💡 Alternative: Use Railway web interface at https://railway.app/dashboard"
    echo "   1. Select project 'zooming-celebration'"
    echo "   2. Select service 'frontend-UI'"
    echo "   3. Set Dockerfile to 'Dockerfile.frontend-simple'"
    echo "   4. Deploy"
fi
