#!/bin/bash

# Script to archive legacy UI files for Phase 4 removal
# Usage: ./scripts/archive-legacy-ui.sh [--dry-run]

DRY_RUN=false
if [ "$1" == "--dry-run" ]; then
    DRY_RUN=true
    echo "🔍 DRY RUN MODE - No files will be moved"
fi

ARCHIVE_DIR="archive/legacy-ui"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
ARCHIVE_SUBDIR="$ARCHIVE_DIR/$TIMESTAMP"

# Directories to archive
LEGACY_DIRS=(
    "cmd/frontend-service/static"
    "services/frontend/public"
)

# File patterns to archive
FILE_PATTERNS=(
    "*.html"
    "js/"
    "css/"
    "components/"
)

echo "📦 Archiving Legacy UI Files"
echo "Archive location: $ARCHIVE_SUBDIR"
echo ""

# Create archive directory structure
if [ "$DRY_RUN" = false ]; then
    mkdir -p "$ARCHIVE_SUBDIR/html"
    mkdir -p "$ARCHIVE_SUBDIR/js"
    mkdir -p "$ARCHIVE_SUBDIR/css"
    mkdir -p "$ARCHIVE_SUBDIR/components"
fi

ARCHIVED_COUNT=0
SKIPPED_COUNT=0

# Archive HTML files
for dir in "${LEGACY_DIRS[@]}"; do
    if [ ! -d "$dir" ]; then
        echo "⚠️  Directory not found: $dir"
        continue
    fi
    
    echo "Processing: $dir"
    
    # Archive HTML files (excluding .next directory)
    find "$dir" -name "*.html" -type f ! -path "*/.next/*" ! -path "*/node_modules/*" ! -path "*/dist/*" | while read -r file; do
        relative_path="${file#$dir/}"
        archive_path="$ARCHIVE_SUBDIR/html/$relative_path"
        archive_dir=$(dirname "$archive_path")
        
        if [ "$DRY_RUN" = true ]; then
            echo "  [DRY RUN] Would archive: $file → $archive_path"
        else
            mkdir -p "$archive_dir"
            if cp "$file" "$archive_path" 2>/dev/null; then
                echo "  ✅ Archived: $relative_path"
                ARCHIVED_COUNT=$((ARCHIVED_COUNT + 1))
            else
                echo "  ❌ Failed: $relative_path"
                SKIPPED_COUNT=$((SKIPPED_COUNT + 1))
            fi
        fi
    done
    
    # Archive JS directory
    if [ -d "$dir/js" ]; then
        if [ "$DRY_RUN" = true ]; then
            echo "  [DRY RUN] Would archive: $dir/js → $ARCHIVE_SUBDIR/js/"
        else
            if cp -r "$dir/js" "$ARCHIVE_SUBDIR/js/" 2>/dev/null; then
                echo "  ✅ Archived: js/"
            fi
        fi
    fi
    
    # Archive CSS directory
    if [ -d "$dir/css" ]; then
        if [ "$DRY_RUN" = true ]; then
            echo "  [DRY RUN] Would archive: $dir/css → $ARCHIVE_SUBDIR/css/"
        else
            if cp -r "$dir/css" "$ARCHIVE_SUBDIR/css/" 2>/dev/null; then
                echo "  ✅ Archived: css/"
            fi
        fi
    fi
    
    # Archive components directory
    if [ -d "$dir/components" ]; then
        if [ "$DRY_RUN" = true ]; then
            echo "  [DRY RUN] Would archive: $dir/components → $ARCHIVE_SUBDIR/components/"
        else
            if cp -r "$dir/components" "$ARCHIVE_SUBDIR/components/" 2>/dev/null; then
                echo "  ✅ Archived: components/"
            fi
        fi
    fi
done

echo ""
echo "=========================================="
echo "Archive Summary"
echo "=========================================="
if [ "$DRY_RUN" = true ]; then
    echo "Mode: DRY RUN (no files moved)"
else
    echo "Files archived: $ARCHIVED_COUNT"
    echo "Files skipped: $SKIPPED_COUNT"
    echo "Archive location: $ARCHIVE_SUBDIR"
    echo ""
    echo "📋 Next steps:"
    echo "1. Verify archive contents"
    echo "2. Test new UI thoroughly"
    echo "3. Remove legacy files after verification"
fi
echo "=========================================="
