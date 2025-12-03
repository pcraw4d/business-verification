#!/bin/bash

# Consolidate Railway Configuration Files
# Removes obsolete railway.json variant files while preserving active configurations

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

# Main cleanup function
main() {
    log "Starting Railway configuration consolidation"
    if [ "$DRY_RUN" = true ]; then
        warn "DRY RUN MODE - No files will be deleted"
    fi
    echo ""
    
    # Files to remove (obsolete variants in root)
    OBSOLETE_FILES=(
        "railway.complete.json"
        "railway.docker.json"
        "railway.clean.json"
        "railway.minimal.json"
        "railway.nixpacks.json"
        "railway.simple.json"
    )
    
    log "=== Removing Obsolete Railway Variant Files ==="
    local removed=0
    local not_found=0
    
    for file in "${OBSOLETE_FILES[@]}"; do
        local full_path="$REPO_ROOT/$file"
        if [ -f "$full_path" ]; then
            if [ "$VERBOSE" = true ]; then
                echo "  - $file"
            fi
            
            if [ "$DRY_RUN" = false ]; then
                rm -f "$full_path"
            fi
            ((removed++))
        else
            if [ "$VERBOSE" = true ]; then
                warn "  File not found: $file (may have been already removed)"
            fi
            ((not_found++))
        fi
    done
    
    echo ""
    log "=== Active Railway Configuration Files (Preserved) ==="
    
    # List active files
    ACTIVE_FILES=(
        "railway.json (main multi-service configuration)"
        "services/classification-service/railway.json"
        "services/risk-assessment-service/railway.json"
        "services/merchant-service/railway.json"
        "services/api-gateway/railway.json"
        "services/frontend-service/railway.json"
        "services/frontend/railway.json"
        "services/redis-cache/railway.json"
        "cmd/frontend-service/railway.json"
        "cmd/business-intelligence-gateway/railway.json"
        "cmd/pipeline-service/railway.json"
        "cmd/service-discovery/railway.json"
        "python_ml_service/railway.json"
    )
    
    for file_desc in "${ACTIVE_FILES[@]}"; do
        if [ "$VERBOSE" = true ]; then
            echo "  ✓ $file_desc"
        fi
    done
    
    echo ""
    if [ "$DRY_RUN" = true ]; then
        log "Would remove $removed obsolete variant file(s)"
        if [ $not_found -gt 0 ]; then
            warn "$not_found file(s) not found (may have been already removed)"
        fi
        warn "This was a DRY RUN. No files were actually deleted."
        warn "Run without --dry-run to perform the actual cleanup."
    else
        log "Removed $removed obsolete variant file(s)"
        if [ $not_found -gt 0 ]; then
            warn "$not_found file(s) were not found (may have been already removed)"
        fi
        log "Railway configuration consolidation completed successfully!"
        echo ""
        log "Active configuration:"
        log "  - Main: railway.json (multi-service configuration)"
        log "  - Service-level: Individual railway.json files in service directories"
    fi
}

# Run main function
main

