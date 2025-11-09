#!/bin/bash

# Script to sync frontend files from services/frontend/public to cmd/frontend-service/static
# This ensures all fixes are in the deployment directory

set -e

echo "🔄 Syncing frontend files for deployment..."
echo ""

SOURCE_DIR="services/frontend/public"
TARGET_DIR="cmd/frontend-service/static"

if [ ! -d "$SOURCE_DIR" ]; then
    echo "❌ Source directory not found: $SOURCE_DIR"
    exit 1
fi

if [ ! -d "$TARGET_DIR" ]; then
    echo "❌ Target directory not found: $TARGET_DIR"
    exit 1
fi

# Count files to sync
HTML_FILES=$(find "$SOURCE_DIR" -name "*.html" -type f | wc -l | tr -d ' ')
JS_FILES=$(find "$SOURCE_DIR" -name "*.js" -type f | wc -l | tr -d ' ')
CSS_FILES=$(find "$SOURCE_DIR" -name "*.css" -type f | wc -l | tr -d ' ')

echo "📊 Files to sync:"
echo "   HTML files: $HTML_FILES"
echo "   JS files: $JS_FILES"
echo "   CSS files: $CSS_FILES"
echo ""

# Sync HTML files
echo "📄 Syncing HTML files..."
rsync -av --delete "$SOURCE_DIR/"*.html "$TARGET_DIR/" 2>/dev/null || {
    echo "⚠️  rsync not available, using cp..."
    cp -v "$SOURCE_DIR"/*.html "$TARGET_DIR/" 2>/dev/null || true
}

# Sync JS files (preserve directory structure)
echo "📜 Syncing JS files..."
if [ -d "$SOURCE_DIR/js" ]; then
    rsync -av --delete "$SOURCE_DIR/js/" "$TARGET_DIR/js/" 2>/dev/null || {
        echo "⚠️  rsync not available, using cp..."
        mkdir -p "$TARGET_DIR/js"
        cp -rv "$SOURCE_DIR/js/"* "$TARGET_DIR/js/" 2>/dev/null || true
    }
fi

# Sync components
echo "🧩 Syncing components..."
if [ -d "$SOURCE_DIR/components" ]; then
    rsync -av --delete "$SOURCE_DIR/components/" "$TARGET_DIR/components/" 2>/dev/null || {
        echo "⚠️  rsync not available, using cp..."
        mkdir -p "$TARGET_DIR/components"
        cp -rv "$SOURCE_DIR/components/"* "$TARGET_DIR/components/" 2>/dev/null || true
    }
fi

# Sync CSS files
echo "🎨 Syncing CSS files..."
if [ -d "$SOURCE_DIR/css" ]; then
    rsync -av --delete "$SOURCE_DIR/css/" "$TARGET_DIR/css/" 2>/dev/null || {
        echo "⚠️  rsync not available, using cp..."
        mkdir -p "$TARGET_DIR/css"
        cp -rv "$SOURCE_DIR/css/"* "$TARGET_DIR/css/" 2>/dev/null || true
    }
fi

echo ""
echo "✅ Sync complete!"
echo ""
echo "📋 Summary:"
echo "   Source: $SOURCE_DIR"
echo "   Target: $TARGET_DIR"
echo ""
echo "⚠️  Remember to commit and push changes to trigger Railway deployment"

