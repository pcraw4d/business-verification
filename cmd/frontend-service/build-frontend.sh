#!/bin/bash

# Frontend Build Script
# Minifies, bundles, and optimizes JavaScript and CSS files

set -e

echo "🚀 Starting frontend build process..."
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Directories
STATIC_DIR="static"
BUILD_DIR="static/dist"
TEMP_DIR=".build-temp"

# Create build directories
mkdir -p "$BUILD_DIR"
mkdir -p "$TEMP_DIR"

# Check if terser is available (for JS minification)
if ! command -v terser &> /dev/null; then
    echo -e "${YELLOW}⚠️  terser not found. Installing...${NC}"
    npm install -g terser 2>/dev/null || {
        echo -e "${YELLOW}⚠️  Global install failed. Using npx...${NC}"
        USE_NPX=true
    }
fi

# Check if csso is available (for CSS minification)
if ! command -v csso &> /dev/null; then
    echo -e "${YELLOW}⚠️  csso not found. Will use npx...${NC}"
    USE_NPX=true
fi

# Function to minify JavaScript
minify_js() {
    local file=$1
    local output=$2
    
    if [ "$USE_NPX" = true ]; then
        npx --yes terser "$file" -c -m -o "$output" 2>/dev/null || cp "$file" "$output"
    else
        terser "$file" -c -m -o "$output" 2>/dev/null || cp "$file" "$output"
    fi
}

# Function to minify CSS
minify_css() {
    local file=$1
    local output=$2
    
    npx --yes csso "$file" -o "$output" 2>/dev/null || cp "$file" "$output"
}

# Count files
JS_COUNT=$(find "$STATIC_DIR" -name "*.js" ! -path "*/node_modules/*" ! -path "*/dist/*" ! -name "*.min.js" -type f | wc -l | tr -d ' ')
CSS_COUNT=$(find "$STATIC_DIR" -name "*.css" ! -path "*/node_modules/*" ! -path "*/dist/*" ! -name "*.min.css" -type f | wc -l | tr -d ' ')

echo "📊 Files to process:"
echo "   JavaScript files: $JS_COUNT"
echo "   CSS files: $CSS_COUNT"
echo ""

# Minify JavaScript files
echo "📜 Minifying JavaScript files..."
JS_PROCESSED=0
JS_FAILED=0

while IFS= read -r -d '' file; do
    rel_path="${file#$STATIC_DIR/}"
    output_file="$BUILD_DIR/${rel_path%.js}.min.js"
    output_dir=$(dirname "$output_file")
    mkdir -p "$output_dir"
    
    if minify_js "$file" "$output_file"; then
        JS_PROCESSED=$((JS_PROCESSED + 1))
        # Calculate size reduction
        orig_size=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null || echo 0)
        new_size=$(stat -f%z "$output_file" 2>/dev/null || stat -c%s "$output_file" 2>/dev/null || echo 0)
        if [ "$orig_size" -gt 0 ] && [ "$new_size" -gt 0 ]; then
            reduction=$((100 - (new_size * 100 / orig_size)))
            echo "   ✓ $(basename "$file") (${reduction}% reduction)"
        fi
    else
        JS_FAILED=$((JS_FAILED + 1))
        echo "   ✗ $(basename "$file")"
    fi
done < <(find "$STATIC_DIR" -name "*.js" ! -path "*/node_modules/*" ! -path "*/dist/*" ! -name "*.min.js" -type f -print0)

echo ""

# Minify CSS files
echo "🎨 Minifying CSS files..."
CSS_PROCESSED=0
CSS_FAILED=0

while IFS= read -r -d '' file; do
    rel_path="${file#$STATIC_DIR/}"
    output_file="$BUILD_DIR/${rel_path%.css}.min.css"
    output_dir=$(dirname "$output_file")
    mkdir -p "$output_dir"
    
    if minify_css "$file" "$output_file"; then
        CSS_PROCESSED=$((CSS_PROCESSED + 1))
        orig_size=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null || echo 0)
        new_size=$(stat -f%z "$output_file" 2>/dev/null || stat -c%s "$output_file" 2>/dev/null || echo 0)
        if [ "$orig_size" -gt 0 ] && [ "$new_size" -gt 0 ]; then
            reduction=$((100 - (new_size * 100 / orig_size)))
            echo "   ✓ $(basename "$file") (${reduction}% reduction)"
        fi
    else
        CSS_FAILED=$((CSS_FAILED + 1))
        echo "   ✗ $(basename "$file")"
    fi
done < <(find "$STATIC_DIR" -name "*.css" ! -path "*/node_modules/*" ! -path "*/dist/*" ! -name "*.min.css" -type f -type f -print0)

echo ""

# Summary
echo "=========================================="
echo "Build Summary"
echo "=========================================="
echo -e "${GREEN}✓ JavaScript: $JS_PROCESSED processed${NC}"
[ $JS_FAILED -gt 0 ] && echo -e "${RED}✗ JavaScript: $JS_FAILED failed${NC}"
echo -e "${GREEN}✓ CSS: $CSS_PROCESSED processed${NC}"
[ $CSS_FAILED -gt 0 ] && echo -e "${RED}✗ CSS: $CSS_FAILED failed${NC}"
echo ""
echo "📦 Minified files saved to: $BUILD_DIR"
echo ""

# Cleanup
rm -rf "$TEMP_DIR"

echo -e "${GREEN}✅ Build complete!${NC}"

