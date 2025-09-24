#!/bin/bash

# Deploy Web Frontend to Railway
echo "🚀 Deploying Web Frontend to Railway..."

# Check if Railway CLI is installed
if ! command -v railway &> /dev/null; then
    echo "❌ Railway CLI not found. Please install it first."
    echo "   Visit: https://docs.railway.app/develop/cli"
    exit 1
fi

# Login to Railway (if not already logged in)
echo "🔐 Checking Railway authentication..."
if ! railway whoami &> /dev/null; then
    echo "🔑 Logging into Railway..."
    railway login --browserless
fi

# Create a new Railway project for the web frontend
echo "📦 Creating Railway project for web frontend..."
railway project create "kyb-platform-web-frontend"

# Link to the project
echo "🔗 Linking to Railway project..."
railway link

# Deploy the web frontend
echo "🚀 Deploying web frontend..."
railway up --detach

# Get the deployment URL
echo "🌐 Getting deployment URL..."
WEB_URL=$(railway domain)

echo "✅ Web Frontend deployed successfully!"
echo "🌐 Web Frontend URL: $WEB_URL"
echo "🔗 API Server URL: https://shimmering-comfort-production.up.railway.app"
echo ""
echo "📋 Next Steps:"
echo "1. Test the web frontend at: $WEB_URL"
echo "2. Verify it connects to the API server"
echo "3. Test the enhanced features (website scraping, classification)"
echo ""
echo "🎯 The web UI should now be accessible and calling the enhanced API features!"
