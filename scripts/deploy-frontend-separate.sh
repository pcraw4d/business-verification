#!/bin/bash

# Deploy Frontend as Separate Railway Service
echo "🚀 Deploying Frontend as Separate Railway Service..."

# Check if Railway CLI is installed
if ! command -v railway &> /dev/null; then
    echo "❌ Railway CLI not found. Please install it first."
    exit 1
fi

# Create a new service in the current project for the frontend
echo "📦 Adding frontend service to current project..."
railway add

# This will create a new service. Let's wait for it to be created
echo "⏳ Waiting for service to be created..."
sleep 5

# Deploy the frontend
echo "🚀 Deploying frontend service..."
railway up --detach

# Get the deployment URL
echo "🌐 Getting frontend deployment URL..."
FRONTEND_URL=$(railway domain)

echo "✅ Frontend deployed successfully!"
echo "🌐 Frontend URL: $FRONTEND_URL"
echo "🔗 API Server URL: https://shimmering-comfort-production.up.railway.app"
echo ""
echo "📋 Architecture:"
echo "   Frontend: $FRONTEND_URL (serves web UI)"
echo "   Backend:  https://shimmering-comfort-production.up.railway.app (API server)"
echo ""
echo "🎯 Both services are now deployed separately on Railway!"
