#!/bin/bash
# Build script for Next.js frontend

set -e

echo "Building Next.js frontend..."

# Navigate to frontend directory
cd "$(dirname "$0")/../../frontend" || exit 1

# Install dependencies if node_modules doesn't exist
if [ ! -d "node_modules" ]; then
    echo "Installing dependencies..."
    npm install
fi

# Build Next.js app
echo "Building Next.js application..."
npm run build

# Copy build output to static directory
echo "Copying build output..."
mkdir -p ../frontend-service/static/.next
cp -r .next/* ../frontend-service/static/.next/ 2>/dev/null || true
cp -r public/* ../frontend-service/static/ 2>/dev/null || true

echo "Frontend build complete!"
