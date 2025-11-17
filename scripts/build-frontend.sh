#!/bin/bash

# Build script for Next.js frontend
# This script builds the Next.js application and prepares it for deployment

set -e

FRONTEND_DIR="frontend"
OUTPUT_DIR="cmd/frontend-service/static/.next"
STATIC_DIR="cmd/frontend-service/static"

echo "Building Next.js frontend..."

# Navigate to frontend directory
cd "$FRONTEND_DIR" || exit 1

# Install dependencies if needed
if [ ! -d "node_modules" ]; then
  echo "Installing dependencies..."
  npm install
fi

# Build Next.js application
echo "Running Next.js build..."
npm run build

# Create output directory if it doesn't exist
mkdir -p "../$OUTPUT_DIR"

# Copy build output based on output mode
if [ -d ".next" ]; then
  echo "Copying build output to $OUTPUT_DIR..."
  
  # Copy .next directory for standalone/server mode
  if [ -d ".next/standalone" ]; then
    echo "Copying standalone build..."
    cp -r .next/standalone/* "../$OUTPUT_DIR/"
  else
    # Copy standard .next directory
    cp -r .next/* "../$OUTPUT_DIR/"
  fi
  
  # Copy public files if they exist
  if [ -d "public" ]; then
    echo "Copying public files..."
    mkdir -p "../$STATIC_DIR/public"
    cp -r public/* "../$STATIC_DIR/public/" || true
  fi
fi

# Copy static export if using static export mode
if [ -d "out" ]; then
  echo "Copying static export..."
  cp -r out/* "../$STATIC_DIR/" || true
fi

echo "Frontend build completed successfully!"
echo "Build output is in: $OUTPUT_DIR"

