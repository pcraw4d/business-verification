#!/bin/bash

# KYB Platform Enhanced Deployment Script
# Supports multiple deployment platforms

set -e

echo "🚀 KYB Platform Enhanced v4.0.0 Deployment Script"
echo "=================================================="

# Build the enhanced railway-server
echo "📦 Building enhanced railway-server..."
cd cmd/railway-server
go build -o railway-server main.go
echo "✅ Build completed successfully"

# Check if we're in a Railway environment
if [ ! -z "$RAILWAY_ENVIRONMENT" ]; then
    echo "🚂 Detected Railway environment"
    echo "Starting enhanced railway-server..."
    ./railway-server
else
    echo "🔧 Local development mode"
    echo "Starting enhanced railway-server on port 8080..."
    ./railway-server
fi
