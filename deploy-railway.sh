#!/bin/bash

# Railway CLI Deployment Script
# Run this after logging in with: railway login

set -e

echo "🚀 Deploying KYB Platform to Railway..."

# Check if logged in
if ! railway whoami > /dev/null 2>&1; then
    echo "❌ Not logged in to Railway. Please run: railway login"
    exit 1
fi

# Deploy
echo "📦 Deploying application..."
railway up

# Get deployment URL
echo "🌐 Getting deployment URL..."
DEPLOYMENT_URL=$(railway domain)

echo "✅ Deployment complete!"
echo "🌐 Your beta testing URL: $DEPLOYMENT_URL"
echo ""
echo "📋 Next steps:"
echo "1. Configure environment variables in Railway dashboard"
echo "2. Add PostgreSQL database if needed"
echo "3. Test the deployment"
echo "4. Share URL with beta testers"
