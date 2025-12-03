#!/bin/bash

# Repository Cleanup Script
# This script safely removes obsolete files from the repository
# Run with --dry-run first to see what would be deleted

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

# Function to remove files
remove_files() {
    local pattern="$1"
    local description="$2"
    local count=0
    
    log "Processing: $description"
    
    while IFS= read -r -d '' file; do
        if [ "$VERBOSE" = true ]; then
            echo "  - $file"
        fi
        
        if [ "$DRY_RUN" = false ]; then
            rm -f "$file"
        fi
        ((count++))
    done < <(find "$REPO_ROOT" -type f -name "$pattern" -print0 2>/dev/null)
    
    if [ "$DRY_RUN" = true ]; then
        log "Would remove $count file(s) matching '$pattern'"
    else
        log "Removed $count file(s) matching '$pattern'"
    fi
}

# Function to remove files by exact path
remove_file_list() {
    local description="$1"
    shift
    local files=("$@")
    local count=0
    
    log "Processing: $description"
    
    for file in "${files[@]}"; do
        local full_path="$REPO_ROOT/$file"
        if [ -f "$full_path" ]; then
            if [ "$VERBOSE" = true ]; then
                echo "  - $file"
            fi
            
            if [ "$DRY_RUN" = false ]; then
                rm -f "$full_path"
            fi
            ((count++))
        fi
    done
    
    if [ "$DRY_RUN" = true ]; then
        log "Would remove $count file(s)"
    else
        log "Removed $count file(s)"
    fi
}

# Main cleanup function
main() {
    log "Starting repository cleanup"
    if [ "$DRY_RUN" = true ]; then
        warn "DRY RUN MODE - No files will be deleted"
    fi
    echo ""
    
    # 1. Remove task completion summary files
    log "=== Category 1: Task Completion Summaries ==="
    remove_files "*task*completion*.md" "Task completion summary files"
    remove_files "*completion_summary*.md" "Completion summary files"
    remove_files "*_completion_summary.md" "Completion summary files (underscore)"
    remove_files "subtask_*_completion_summary.md" "Subtask completion summaries"
    remove_files "TASK_*_COMPLETION_SUMMARY.md" "Task completion summaries (uppercase)"
    echo ""
    
    # 2. Remove old log files
    log "=== Category 2: Old Log Files ==="
    remove_files "*.log" "Log files"
    echo ""
    
    # 3. Remove old test output files
    log "=== Category 3: Old Test Output Files ==="
    remove_files "*test-output*.txt" "Test output text files"
    remove_files "*test-results*.txt" "Test results text files"
    remove_files "beta-test*.txt" "Beta test output files"
    remove_files "performance-test-report.txt" "Performance test reports"
    remove_files "uat-test-report.txt" "UAT test reports"
    remove_files "review-test-output.txt" "Review test output"
    echo ""
    
    # 4. Remove backup files
    log "=== Category 4: Backup Files ==="
    remove_files "*.backup" "Backup files"
    echo ""
    
    # 5. Remove old coverage files
    log "=== Category 5: Old Coverage Files ==="
    remove_files "coverage*.out" "Coverage output files"
    remove_files "coverage.html" "Coverage HTML files"
    echo ""
    
    # 6. Remove old accuracy report JSON files (keep baseline if exists)
    log "=== Category 6: Old Accuracy Report JSON Files ==="
    remove_files "accuracy_report_railway_production_*.json" "Old Railway production accuracy reports"
    remove_files "accuracy_report_v*.json" "Old versioned accuracy reports"
    # Keep accuracy_report_baseline.json and accuracy_report_integration_phases.json for reference
    echo ""
    
    # 7. Remove old JSON analysis files (dated ones)
    log "=== Category 7: Old JSON Analysis Files ==="
    remove_files "*_2025-09-19.json" "Old dated JSON analysis files"
    echo ""
    
    # 8. Remove specific obsolete files
    log "=== Category 8: Specific Obsolete Files ==="
    local obsolete_files=(
        "compliance-verification-report.txt"
        "database.test"
        "execute-task-1-1-validation.go"
        "execute-task-3-2-testing.go"
        "fix_complex_logger_calls.py"
        "fix_imports.sh"
        "fix_logger_calls.py"
        "fix_performance_alerting.sh"
        "get_railway_logs.sh"
        "git_automation.py"
        "main"  # Compiled binary
    )
    remove_file_list "Specific obsolete files" "${obsolete_files[@]}"
    echo ""
    
    # 9. Clean up empty completion-summaries directory
    if [ -d "$REPO_ROOT/completion-summaries" ]; then
        if [ -z "$(ls -A "$REPO_ROOT/completion-summaries")" ]; then
            log "Removing empty completion-summaries directory"
            if [ "$DRY_RUN" = false ]; then
                rmdir "$REPO_ROOT/completion-summaries"
            fi
        fi
    fi
    echo ""
    
    # Summary
    log "=== Cleanup Complete ==="
    if [ "$DRY_RUN" = true ]; then
        warn "This was a DRY RUN. No files were actually deleted."
        warn "Run without --dry-run to perform the actual cleanup."
    else
        log "Repository cleanup completed successfully!"
    fi
}

# Run main function
main

