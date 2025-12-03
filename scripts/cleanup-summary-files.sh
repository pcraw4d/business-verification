#!/bin/bash

# Cleanup Summary Files Script
# Removes old summary/status report files while preserving important documentation

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

DRY_RUN=false
VERBOSE=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--dry-run] [--verbose]"
            exit 1
            ;;
    esac
done

# Function to log messages
log() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Files to preserve (important documentation)
PRESERVE_FILES=(
    "repository-cleanup-plan.md"
    "phase1-implementation-summary.md"
    "phase1-testing-guide.md"
    "phase1-log-analysis-guide.md"
    "railway-deployment-solution.md"
    "railway-deployment-issues.md"
)

# Directories to exclude
EXCLUDE_DIRS=(
    ".cursor/plans"
    "node_modules"
    ".git"
    "archive"
)

# Function to check if file should be preserved
should_preserve() {
    local file="$1"
    local basename=$(basename "$file")
    
    # Check preserve list
    for preserve in "${PRESERVE_FILES[@]}"; do
        if [[ "$basename" == "$preserve" ]] || [[ "$file" == *"$preserve" ]]; then
            return 0
        fi
    done
    
    # Check exclude directories
    for exclude in "${EXCLUDE_DIRS[@]}"; do
        if [[ "$file" == *"$exclude"* ]]; then
            return 0
        fi
    done
    
    return 1
}

# Function to remove summary files
remove_summary_files() {
    local count=0
    local preserved=0
    
    log "Searching for summary files to remove..."
    echo ""
    
    while IFS= read -r -d '' file; do
        # Skip if should be preserved
        if should_preserve "$file"; then
            if [ "$VERBOSE" = true ]; then
                warn "Preserving: $file"
            fi
            ((preserved++))
            continue
        fi
        
        if [ "$VERBOSE" = true ]; then
            echo "  - $file"
        fi
        
        if [ "$DRY_RUN" = false ]; then
            rm -f "$file"
        fi
        ((count++))
    done < <(find "$REPO_ROOT" -type f \( -name "*summary*.md" -o -name "*SUMMARY*.md" -o -name "*_SUMMARY.md" \) ! -path "*/node_modules/*" ! -path "*/.git/*" -print0 2>/dev/null)
    
    if [ "$DRY_RUN" = true ]; then
        log "Would remove $count file(s)"
        log "Would preserve $preserved file(s)"
    else
        log "Removed $count file(s)"
        log "Preserved $preserved file(s)"
    fi
}

# Main cleanup function
main() {
    log "Starting summary files cleanup"
    if [ "$DRY_RUN" = true ]; then
        warn "DRY RUN MODE - No files will be deleted"
    fi
    echo ""
    
    remove_summary_files
    echo ""
    
    log "=== Cleanup Complete ==="
    if [ "$DRY_RUN" = true ]; then
        warn "This was a DRY RUN. No files were actually deleted."
        warn "Run without --dry-run to perform the actual cleanup."
    else
        log "Summary files cleanup completed successfully!"
    fi
}

# Run main function
main

