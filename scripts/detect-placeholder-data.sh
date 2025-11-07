#!/bin/bash

# Placeholder Data Detection Script
# Scans codebase for mock/placeholder data usage

set -e

# Colors
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

# Patterns to detect
PATTERNS=(
    "Sample Merchant"
    "Mock"
    "mock"
    "TODO.*return"
    "placeholder"
    "test-"
    "dummy"
    "fake"
    "example"
    "For now"
    "temporary"
    "fallback"
)

# Directories to scan
SCAN_DIRS=(
    "services"
    "internal"
    "web"
    "cmd"
)

# Results
declare -a FOUND_ISSUES
ISSUE_COUNT=0

echo -e "${BLUE}🔍 Scanning for Placeholder/Mock Data...${NC}\n"

for dir in "${SCAN_DIRS[@]}"; do
    if [ ! -d "$dir" ]; then
        continue
    fi
    
    echo -e "${YELLOW}Scanning $dir...${NC}"
    
    # Find Go files
    while IFS= read -r file; do
        if [ -z "$file" ]; then
            continue
        fi
        
        for pattern in "${PATTERNS[@]}"; do
            # Skip test files for some patterns
            if [[ "$file" == *_test.go ]] && [[ "$pattern" == "test-" ]]; then
                continue
            fi
            
            # Search for pattern
            matches=$(grep -n -i "$pattern" "$file" 2>/dev/null || true)
            
            if [ -n "$matches" ]; then
                while IFS= read -r line; do
                    if [ -n "$line" ]; then
                        ISSUE_COUNT=$((ISSUE_COUNT + 1))
                        FOUND_ISSUES+=("$file: $line")
                        echo -e "${RED}  ⚠️  $file: $line${NC}"
                    fi
                done <<< "$matches"
            fi
        done
    done < <(find "$dir" -type f \( -name "*.go" -o -name "*.js" \) ! -name "*_test.go" ! -name "*.test.js" ! -path "*/node_modules/*" ! -path "*/vendor/*" ! -path "*/.git/*" 2>/dev/null)
done

echo ""
echo -e "${BLUE}📊 Summary${NC}"
echo "Total issues found: $ISSUE_COUNT"

if [ $ISSUE_COUNT -eq 0 ]; then
    echo -e "${GREEN}✅ No placeholder data detected!${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠️  Found $ISSUE_COUNT potential placeholder data usages${NC}"
    echo ""
    echo "Review the above matches to ensure they are not used in production code paths."
    exit 1
fi

