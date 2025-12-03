#!/bin/bash

# Analyze Markdown Files Script
# Categorizes markdown files to identify obsolete ones

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

OUTPUT_FILE="$REPO_ROOT/docs/markdown-files-analysis.md"

log() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Main analysis function
main() {
    log "Analyzing markdown files in repository..."
    echo ""
    
    # Count total
    TOTAL=$(find "$REPO_ROOT" -name "*.md" -type f ! -path "*/node_modules/*" ! -path "*/.git/*" ! -path "*/archive/*" ! -path "*/.cursor/*" ! -path "*/venv/*" 2>/dev/null | wc -l | tr -d ' ')
    log "Total markdown files found: $TOTAL"
    echo ""
    
    # Count root level
    ROOT_COUNT=$(find "$REPO_ROOT" -maxdepth 1 -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')
    log "Root level markdown files: $ROOT_COUNT"
    echo ""
    
    # Start output file
    > "$OUTPUT_FILE"
    echo "# Markdown Files Analysis" >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    echo "**Generated**: $(date)" >> "$OUTPUT_FILE"
    echo "**Total Files**: $TOTAL" >> "$OUTPUT_FILE"
    echo "**Root Level Files**: $ROOT_COUNT" >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    
    log "=== Category Analysis ==="
    echo ""
    
    TOTAL_OBSOLETE=0
    
    # Analyze each category
    analyze_category "status_reports" "*STATUS*.md *STATUS_REPORT*.md *COMPLETION_STATUS*.md"
    analyze_category "test_results" "*TEST_RESULTS*.md *TEST_REPORT*.md PHASE_*_TEST*.md *TESTING_RESULTS*.md"
    analyze_category "investigations" "*INVESTIGATION*.md *ROOT_CAUSE*.md *ANALYSIS*.md *FINDINGS*.md"
    analyze_category "fix_plans" "*FIX_PLAN*.md *FIX*.md *FIXES*.md *REMEDIATION*.md"
    analyze_category "deployment_reports" "DEPLOYMENT_*.md *DEPLOYMENT*.md"
    analyze_category "phase_docs" "PHASE_*.md phase_*.md"
    analyze_category "error_reports" "ERROR_*.md *ERROR*.md"
    analyze_category "verification" "*VERIFICATION*.md *VERIFICATION_REPORT*.md"
    analyze_category "test_execution" "*TEST_EXECUTION*.md *EXECUTION_RESULTS*.md"
    analyze_category "completion" "*COMPLETION*.md *COMPLETE*.md"
    analyze_category "reflection" "*reflection*.md *REFLECTION*.md"
    analyze_category "old_plans" "*PLAN*.md *PLANNING*.md *ROADMAP*.md"
    analyze_category "analysis" "*ANALYSIS*.md *ANALYSIS_REPORT*.md"
    
    echo "" >> "$OUTPUT_FILE"
    echo "## Summary" >> "$OUTPUT_FILE"
    echo "**Total potentially obsolete files (root level)**: $TOTAL_OBSOLETE" >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    echo "## Files to Preserve" >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    echo "The following files should be preserved:" >> "$OUTPUT_FILE"
    echo "- README.md" >> "$OUTPUT_FILE"
    echo "- CONTRIBUTING.md" >> "$OUTPUT_FILE"
    echo "- DATABASE_SETUP_GUIDE.md" >> "$OUTPUT_FILE"
    echo "- docs/repository-cleanup-plan.md" >> "$OUTPUT_FILE"
    echo "- docs/repository-cleanup-test-results.md" >> "$OUTPUT_FILE"
    echo "- docs/phase1-*.md (active guides)" >> "$OUTPUT_FILE"
    echo "- docs/railway-deployment-*.md (active guides)" >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    
    echo ""
    log "Analysis complete. Results saved to: $OUTPUT_FILE"
    log "Total potentially obsolete files (root level): $TOTAL_OBSOLETE"
}

analyze_category() {
    local category="$1"
    local patterns="$2"
    
    local count=0
    local files=""
    
    # Convert patterns to find-compatible format
    while IFS= read -r file; do
        local basename_file=$(basename "$file")
        local matched=false
        
        # Check each pattern
        for pattern in $patterns; do
            # Convert glob to regex
            regex=$(echo "$pattern" | sed 's/\*/.*/g' | sed 's/\.md$//')
            if echo "$basename_file" | grep -qiE "$regex"; then
                matched=true
                break
            fi
        done
        
        if [ "$matched" = true ]; then
            files="$files$basename_file"$'\n'
            ((count++))
        fi
    done < <(find "$REPO_ROOT" -maxdepth 1 -name "*.md" -type f 2>/dev/null)
    
    if [ "$count" -gt 0 ]; then
        echo "## $category" >> "$OUTPUT_FILE"
        echo "**Patterns**: \`$patterns\`" >> "$OUTPUT_FILE"
        echo "**Count**: $count files" >> "$OUTPUT_FILE"
        echo "" >> "$OUTPUT_FILE"
        echo "$files" | sort >> "$OUTPUT_FILE"
        echo "" >> "$OUTPUT_FILE"
        
        TOTAL_OBSOLETE=$((TOTAL_OBSOLETE + count))
        log "  $category: $count files"
    fi
}

# Run main function
main
