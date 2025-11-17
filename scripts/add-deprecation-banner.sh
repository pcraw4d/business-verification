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

STATIC_DIR="cmd/frontend-service/static"

# Find all HTML files and add deprecation banner after <body> tag
find "$STATIC_DIR" -name "*.html" -type f | while read -r file; do
    # Skip if already has deprecation notice
    if grep -q "DEPRECATION NOTICE" "$file"; then
        echo "Skipping $file (already has deprecation notice)"
        continue
    fi
    
    # Add banner after <body> tag
    if grep -q "<body" "$file"; then
        # Use sed to insert banner after <body> tag
        sed -i.bak "/<body[^>]*>/a\\
$DEPRECATION_BANNER
" "$file"
        echo "Added deprecation banner to $file"
        # Remove backup file
        rm -f "${file}.bak"
    fi
done

echo "Deprecation banners added to all legacy HTML files"

