#!/bin/bash

# Script to verify frontend files are synced between source and deployment directories
# Use this before committing to ensure files are in sync

set -e

SOURCE_DIR="services/frontend/public"
TARGET_DIR="cmd/frontend-service/static"

echo "🔍 Verifying frontend file sync..."
echo ""

if [ ! -d "$SOURCE_DIR" ]; then
    echo "❌ Source directory not found: $SOURCE_DIR"
    exit 1
fi

if [ ! -d "$TARGET_DIR" ]; then
    echo "❌ Target directory not found: $TARGET_DIR"
    exit 1
fi

# Check critical files
CRITICAL_FILES=("add-merchant.html" "merchant-details.html")
UNSYNCED=0
MISSING=0

echo "📋 Checking critical files:"
for file in "${CRITICAL_FILES[@]}"; do
    if [ -f "$SOURCE_DIR/$file" ]; then
        if [ ! -f "$TARGET_DIR/$file" ]; then
            echo "  ❌ $file: Missing in deployment directory"
            MISSING=$((MISSING + 1))
        elif ! diff -q "$SOURCE_DIR/$file" "$TARGET_DIR/$file" > /dev/null 2>&1; then
            echo "  ⚠️  $file: Files differ"
            UNSYNCED=$((UNSYNCED + 1))
        else
            echo "  ✅ $file: Synced"
        fi
    else
        echo "  ⚠️  $file: Not found in source directory"
    fi
done

echo ""
echo "📊 Summary:"
echo "   Unsynced files: $UNSYNCED"
echo "   Missing files: $MISSING"

if [ $UNSYNCED -gt 0 ] || [ $MISSING -gt 0 ]; then
    echo ""
    echo "❌ Sync required!"
    echo "   Run: ./scripts/sync-frontend-files.sh"
    exit 1
else
    echo ""
    echo "✅ All critical files are synced"
    exit 0
fi

