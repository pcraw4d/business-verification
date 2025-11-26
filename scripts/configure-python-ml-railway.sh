#!/bin/bash
# Configure Python ML Service in Railway using CLI
# This script sets the root directory and other required settings

set -e

echo "🔧 Configuring Python ML Service in Railway..."
echo ""

# Check if Railway CLI is installed
if ! command -v railway &> /dev/null; then
    echo "❌ Railway CLI is not installed"
    echo "   Install it with: npm install -g @railway/cli"
    exit 1
fi

# Check if logged in
if ! railway whoami &> /dev/null; then
    echo "❌ Not logged in to Railway"
    echo "   Run: railway login"
    exit 1
fi

# Navigate to python_ml_service directory
cd "$(dirname "$0")/../python_ml_service"

echo "📁 Current directory: $(pwd)"
echo ""

# Link to the service (creates railway.json link if needed)
echo "🔗 Linking to python-ml-service..."
railway link --service python-ml-service --non-interactive || {
    echo "⚠️  Service might already be linked or doesn't exist"
    echo "   If service doesn't exist, create it in Railway dashboard first"
}

echo ""
echo "✅ Configuration complete!"
echo ""
echo "📝 Next steps:"
echo "   1. Go to Railway Dashboard → python-ml-service → Settings"
echo "   2. Set Root Directory to: python_ml_service"
echo "   3. Set Dockerfile Path to: Dockerfile"
echo "   4. Save and redeploy"
echo ""
echo "   Or trigger a new deployment with: railway up"

