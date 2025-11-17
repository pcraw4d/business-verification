#!/bin/bash

# Script to remove legacy UI files after archiving
# Usage: ./scripts/remove-legacy-files.sh [--dry-run]

DRY_RUN=false
if [ "$1" == "--dry-run" ]; then
    DRY_RUN=true
    echo "🔍 DRY RUN MODE - No files will be deleted"
fi

echo "🗑️  Removing Legacy UI Files"
echo ""

REMOVED_COUNT=0
SKIPPED_COUNT=0

# Directories to clean
LEGACY_DIRS=(
    "cmd/frontend-service/static"
    "services/frontend/public"
)

# Remove HTML files (excluding .next directory)
for dir in "${LEGACY_DIRS[@]}"; do
    if [ ! -d "$dir" ]; then
        echo "⚠️  Directory not found: $dir"
        continue
    fi
    
    echo "Processing: $dir"
    
    # Remove HTML files (excluding .next and node_modules)
    find "$dir" -name "*.html" -type f ! -path "*/.next/*" ! -path "*/node_modules/*" ! -path "*/dist/*" | while read -r file; do
        if [ "$DRY_RUN" = true ]; then
            echo "  [DRY RUN] Would remove: $file"
        else
            if rm "$file" 2>/dev/null; then
                echo "  ✅ Removed: ${file#$dir/}"
                REMOVED_COUNT=$((REMOVED_COUNT + 1))
            else
                echo "  ❌ Failed: ${file#$dir/}"
                SKIPPED_COUNT=$((SKIPPED_COUNT + 1))
            fi
        fi
    done
    
    # Remove JS directory (if not used by new UI)
    if [ -d "$dir/js" ] && [ ! -d "$dir/.next" ]; then
        if [ "$DRY_RUN" = true ]; then
            echo "  [DRY RUN] Would remove: $dir/js/"
        else
            if rm -rf "$dir/js" 2>/dev/null; then
                echo "  ✅ Removed: js/"
            fi
        fi
    fi
    
    # Remove CSS directory (if not used by new UI)
    if [ -d "$dir/css" ] && [ ! -d "$dir/.next" ]; then
        if [ "$DRY_RUN" = true ]; then
            echo "  [DRY RUN] Would remove: $dir/css/"
        else
            if rm -rf "$dir/css" 2>/dev/null; then
                echo "  ✅ Removed: css/"
            fi
        fi
    fi
    
    # Remove components directory (if not used by new UI)
    if [ -d "$dir/components" ] && [ ! -d "$dir/.next" ]; then
        if [ "$DRY_RUN" = true ]; then
            echo "  [DRY RUN] Would remove: $dir/components/"
        else
            if rm -rf "$dir/components" 2>/dev/null; then
                echo "  ✅ Removed: components/"
            fi
        fi
    fi
done

echo ""
echo "=========================================="
echo "Removal Summary"
echo "=========================================="
if [ "$DRY_RUN" = true ]; then
    echo "Mode: DRY RUN (no files deleted)"
    echo ""
    echo "⚠️  WARNING: This will permanently delete legacy files!"
    echo "Make sure you have:"
    echo "1. Created archive using: ./scripts/archive-legacy-ui.sh"
    echo "2. Verified archive contents"
    echo "3. Tested new UI thoroughly"
else
    echo "Files removed: $REMOVED_COUNT"
    echo "Files skipped: $SKIPPED_COUNT"
    echo ""
    echo "✅ Legacy files removed"
    echo "📋 Next steps:"
    echo "1. Update documentation"
    echo "2. Run final verification"
    echo "3. Test all pages"
fi
echo "=========================================="

