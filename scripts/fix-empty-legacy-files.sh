#!/bin/bash

# Script to fix empty legacy HTML files by copying from services/frontend/public/
# Usage: ./scripts/fix-empty-legacy-files.sh

SOURCE_DIR="services/frontend/public"
TARGET_DIR="cmd/frontend-service/static"

# List of empty files that need to be fixed
EMPTY_FILES=(
    "admin-dashboard.html"
    "api-test.html"
    "business-growth-analytics-testing.html"
    "compliance-alert-system.html"
    "compliance-dashboard.html"
    "compliance-gap-analysis.html"
    "compliance-summary-reports.html"
    "enhanced-risk-indicators.html"
    "index.html"
    "merchant-comparison.html"
    "merchant-hub.html"
    "monitoring-dashboard.html"
    "register.html"
)

FIXED_COUNT=0
NOT_FOUND_COUNT=0
ALREADY_HAS_CONTENT=0

echo "🔧 Fixing empty legacy HTML files..."
echo ""

for file in "${EMPTY_FILES[@]}"; do
    source_file="$SOURCE_DIR/$file"
    target_file="$TARGET_DIR/$file"
    
    # Check if target is actually empty
    if [ -s "$target_file" ]; then
        echo "⏭️  Skipping $file (already has content)"
        ALREADY_HAS_CONTENT=$((ALREADY_HAS_CONTENT + 1))
        continue
    fi
    
    # Check if source file exists
    if [ ! -f "$source_file" ]; then
        echo "⚠️  Source not found: $file"
        NOT_FOUND_COUNT=$((NOT_FOUND_COUNT + 1))
        continue
    fi
    
    # Check if source has content
    if [ ! -s "$source_file" ]; then
        echo "⚠️  Source is also empty: $file"
        NOT_FOUND_COUNT=$((NOT_FOUND_COUNT + 1))
        continue
    fi
    
    # Copy file
    if cp "$source_file" "$target_file" 2>/dev/null; then
        echo "✅ Fixed: $file"
        FIXED_COUNT=$((FIXED_COUNT + 1))
    else
        echo "❌ Failed to copy: $file"
        NOT_FOUND_COUNT=$((NOT_FOUND_COUNT + 1))
    fi
done

echo ""
echo "=========================================="
echo "Empty Files Fix Summary"
echo "=========================================="
echo "Files fixed: $FIXED_COUNT"
echo "Files not found/empty: $NOT_FOUND_COUNT"
echo "Files already have content: $ALREADY_HAS_CONTENT"
echo "Total processed: ${#EMPTY_FILES[@]}"
echo "=========================================="

if [ $FIXED_COUNT -gt 0 ]; then
    echo ""
    echo "✅ Successfully fixed $FIXED_COUNT empty files!"
    echo "📋 Next step: Run deprecation banner script again:"
    echo "   ./scripts/add-deprecation-banner.sh"
fi

