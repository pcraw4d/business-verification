#!/bin/bash

# Script to add deprecation banner to all legacy HTML files
# Usage: ./scripts/add-deprecation-banner.sh

DEPRECATION_BANNER='<!-- DEPRECATION NOTICE -->
<div style="background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%); color: white; padding: 16px; margin-bottom: 20px; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.1);">
    <div style="display: flex; align-items: center; gap: 12px;">
        <span style="font-size: 24px;">⚠️</span>
        <div style="flex: 1;">
            <strong style="font-size: 18px; display: block; margin-bottom: 4px;">This page is deprecated</strong>
            <p style="margin: 0; font-size: 14px; opacity: 0.95;">
                This legacy UI page has been replaced with a new shadcn UI implementation. 
                Please use the new UI for the best experience. 
                <a href="/" style="color: white; text-decoration: underline; font-weight: bold;">Go to new UI →</a>
            </p>
        </div>
    </div>
</div>'

# Directories containing legacy HTML files
STATIC_DIRS=(
    "cmd/frontend-service/static"
    "services/frontend/public"
)

TOTAL_FILES=0
PROCESSED_FILES=0
SKIPPED_FILES=0
ERROR_FILES=0

# Process each directory
for STATIC_DIR in "${STATIC_DIRS[@]}"; do
    if [ ! -d "$STATIC_DIR" ]; then
        echo "Warning: Directory $STATIC_DIR does not exist, skipping..."
        continue
    fi
    
    echo "Processing directory: $STATIC_DIR"
    
    # Find all HTML files (excluding node_modules and dist directories)
    find "$STATIC_DIR" -name "*.html" -type f ! -path "*/node_modules/*" ! -path "*/dist/*" | while read -r file; do
        TOTAL_FILES=$((TOTAL_FILES + 1))
        
        # Skip if already has deprecation notice
        if grep -q "DEPRECATION NOTICE" "$file"; then
            echo "Skipping $file (already has deprecation notice)"
            SKIPPED_FILES=$((SKIPPED_FILES + 1))
            continue
        fi
        
        # Add banner after <body> tag
        if grep -q "<body" "$file"; then
            # Create a temporary file with the banner inserted
            TEMP_FILE=$(mktemp)
            
            # Use Python to insert banner after <body> tag (more reliable for multi-line strings)
            python3 << EOF > "$TEMP_FILE"
import sys

banner = """<!-- DEPRECATION NOTICE -->
<div style="background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%); color: white; padding: 16px; margin-bottom: 20px; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.1);">
    <div style="display: flex; align-items: center; gap: 12px;">
        <span style="font-size: 24px;">⚠️</span>
        <div style="flex: 1;">
            <strong style="font-size: 18px; display: block; margin-bottom: 4px;">This page is deprecated</strong>
            <p style="margin: 0; font-size: 14px; opacity: 0.95;">
                This legacy UI page has been replaced with a new shadcn UI implementation. 
                Please use the new UI for the best experience. 
                <a href="/" style="color: white; text-decoration: underline; font-weight: bold;">Go to new UI →</a>
            </p>
        </div>
    </div>
</div>"""

with open("$file", 'r') as f:
    lines = f.readlines()
    
banner_added = False
for i, line in enumerate(lines):
    print(line, end='')
    if '<body' in line and not banner_added:
        print(banner)
        banner_added = True
EOF
            
            # Replace original file
            if [ -f "$TEMP_FILE" ] && [ -s "$TEMP_FILE" ]; then
                mv "$TEMP_FILE" "$file" 2>/dev/null
                echo "✓ Added deprecation banner to $file"
                PROCESSED_FILES=$((PROCESSED_FILES + 1))
            else
                echo "✗ Error processing $file"
                ERROR_FILES=$((ERROR_FILES + 1))
                rm -f "$TEMP_FILE"
            fi
        else
            echo "Warning: $file does not have a <body> tag, skipping..."
            SKIPPED_FILES=$((SKIPPED_FILES + 1))
        fi
    done
done

echo ""
echo "=========================================="
echo "Deprecation Banner Addition Summary"
echo "=========================================="
echo "Total HTML files found: $TOTAL_FILES"
echo "Files processed: $PROCESSED_FILES"
echo "Files skipped (already have banner): $SKIPPED_FILES"
echo "Files with errors: $ERROR_FILES"
echo "=========================================="

