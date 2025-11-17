#!/bin/bash

# Script to archive legacy UI files
# Usage: ./scripts/archive-legacy-ui.sh

STATIC_DIR="cmd/frontend-service/static"
ARCHIVE_DIR="archive/legacy-ui"

echo "Archiving legacy UI files..."

# Archive HTML files
if [ -d "$STATIC_DIR" ]; then
    echo "Archiving HTML files..."
    find "$STATIC_DIR" -maxdepth 1 -name "*.html" -type f -exec cp {} "$ARCHIVE_DIR/html/" \;
    echo "HTML files archived to $ARCHIVE_DIR/html/"
fi

# Archive JavaScript files
if [ -d "$STATIC_DIR/js" ]; then
    echo "Archiving JavaScript files..."
    cp -r "$STATIC_DIR/js" "$ARCHIVE_DIR/"
    echo "JavaScript files archived to $ARCHIVE_DIR/js/"
fi

# Archive CSS files
if [ -d "$STATIC_DIR/css" ]; then
    echo "Archiving CSS files..."
    cp -r "$STATIC_DIR/css" "$ARCHIVE_DIR/"
    echo "CSS files archived to $ARCHIVE_DIR/css/"
fi

# Archive components
if [ -d "$STATIC_DIR/components" ]; then
    echo "Archiving component files..."
    cp -r "$STATIC_DIR/components" "$ARCHIVE_DIR/"
    echo "Component files archived to $ARCHIVE_DIR/components/"
fi

echo "Legacy UI files archived successfully!"
echo "Archive location: $ARCHIVE_DIR"

