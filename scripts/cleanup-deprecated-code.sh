#!/bin/bash

# Automated Cleanup Script for Deprecated Code
# This script identifies and helps clean up deprecated code patterns
# in the KYB Platform codebase.

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
LOG_FILE="$PROJECT_ROOT/logs/deprecated-cleanup.log"
REPORT_FILE="$PROJECT_ROOT/reports/deprecated-code-report-$(date +%Y%m%d-%H%M%S).json"
DRY_RUN=false
INTERACTIVE=true
BACKUP_DIR="$PROJECT_ROOT/.cleanup-backups"

# Ensure required directories exist
mkdir -p "$(dirname "$LOG_FILE")" "$(dirname "$REPORT_FILE")" "$BACKUP_DIR"

# Patterns for deprecated code detection
declare -A DEPRECATED_PATTERNS=(
    ["deprecated_comments"]="// DEPRECATED|// TODO: Remove|// FIXME: Legacy|// Legacy code"
    ["deprecated_imports"]="webanalysis.problematic|github.com/deprecated|legacy/|old/"
    ["deprecated_functions"]="func.*Deprecated|func.*Legacy|func.*Old"
    ["deprecated_variables"]="deprecated_|legacy_|old_"
    ["deprecated_types"]="type.*Deprecated|type.*Legacy|type.*Old"
    ["deprecated_files"]="_deprecated\.|\.deprecated\.|_legacy\.|\.legacy\.|_old\.|\.old\."
    ["test_files"]="_test\.go$"
    ["backup_files"]="\.bak$|\.backup$|\.orig$"
    ["temp_files"]="\.tmp$|\.temp$|~$"
)

# File exclusion patterns
EXCLUDE_PATTERNS=(
    "vendor/"
    "node_modules/"
    ".git/"
    "*.log"
    "*.bin"
    "*.exe"
    ".cleanup-backups/"
    "reports/"
    "logs/"
)

# Print functions
print_header() {
    echo -e "${BLUE}=== $1 ===${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
    echo "$(date '+%Y-%m-%d %H:%M:%S') [INFO] $1" >> "$LOG_FILE"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
    echo "$(date '+%Y-%m-%d %H:%M:%S') [WARNING] $1" >> "$LOG_FILE"
}

print_error() {
    echo -e "${RED}✗${NC} $1" >&2
    echo "$(date '+%Y-%m-%d %H:%M:%S') [ERROR] $1" >> "$LOG_FILE"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
    echo "$(date '+%Y-%m-%d %H:%M:%S') [SUCCESS] $1" >> "$LOG_FILE"
}

# Usage function
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Automated cleanup script for deprecated code in the KYB Platform.

OPTIONS:
    -d, --dry-run           Perform a dry run without making changes
    -n, --non-interactive   Run without user interaction
    -p, --pattern PATTERN   Only check specific pattern (see list below)
    -f, --file FILE         Only check specific file
    -r, --report-only       Generate report only, no cleanup
    -b, --backup            Create backup before cleanup (default: true)
    --no-backup             Skip backup creation
    -h, --help              Show this help message

PATTERNS:
    deprecated_comments     Comments marking deprecated code
    deprecated_imports      Deprecated import statements
    deprecated_functions    Deprecated function definitions
    deprecated_variables    Deprecated variable names
    deprecated_types        Deprecated type definitions
    deprecated_files        Files marked as deprecated
    test_files             Test files that can be cleaned up
    backup_files           Backup files (.bak, .backup, etc.)
    temp_files             Temporary files (.tmp, .temp, etc.)

EXAMPLES:
    $0                      # Interactive cleanup with backup
    $0 --dry-run            # Show what would be cleaned without changes
    $0 --pattern backup_files --non-interactive
                           # Clean backup files without interaction
    $0 --report-only       # Generate report only

EOF
}

# Parse command line arguments
PATTERN_FILTER=""
FILE_FILTER=""
REPORT_ONLY=false
CREATE_BACKUP=true

while [[ $# -gt 0 ]]; do
    case $1 in
        -d|--dry-run)
            DRY_RUN=true
            shift
            ;;
        -n|--non-interactive)
            INTERACTIVE=false
            shift
            ;;
        -p|--pattern)
            PATTERN_FILTER="$2"
            shift 2
            ;;
        -f|--file)
            FILE_FILTER="$2"
            shift 2
            ;;
        -r|--report-only)
            REPORT_ONLY=true
            shift
            ;;
        -b|--backup)
            CREATE_BACKUP=true
            shift
            ;;
        --no-backup)
            CREATE_BACKUP=false
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Validate pattern filter
if [[ -n "$PATTERN_FILTER" && ! -v "DEPRECATED_PATTERNS[$PATTERN_FILTER]" ]]; then
    print_error "Invalid pattern: $PATTERN_FILTER"
    echo "Available patterns: ${!DEPRECATED_PATTERNS[*]}"
    exit 1
fi

# Check if file filter exists
if [[ -n "$FILE_FILTER" && ! -f "$FILE_FILTER" ]]; then
    print_error "File not found: $FILE_FILTER"
    exit 1
fi

# Function to check if file should be excluded
should_exclude_file() {
    local file="$1"
    for pattern in "${EXCLUDE_PATTERNS[@]}"; do
        if [[ "$file" == *"$pattern"* ]]; then
            return 0
        fi
    done
    return 1
}

# Function to create backup
create_backup() {
    local file="$1"
    if [[ "$CREATE_BACKUP" == true && ! "$DRY_RUN" == true ]]; then
        local backup_path="$BACKUP_DIR/$(date +%Y%m%d-%H%M%S)-$(basename "$file").bak"
        cp "$file" "$backup_path"
        print_info "Created backup: $backup_path"
    fi
}

# Function to scan for deprecated patterns
scan_deprecated_patterns() {
    local pattern_name="$1"
    local pattern="$2"
    local results=()
    
    print_info "Scanning for pattern: $pattern_name"
    
    # Find files to scan
    local files_to_scan=()
    if [[ -n "$FILE_FILTER" ]]; then
        files_to_scan=("$FILE_FILTER")
    else
        while IFS= read -r -d '' file; do
            if ! should_exclude_file "$file"; then
                files_to_scan+=("$file")
            fi
        done < <(find "$PROJECT_ROOT" -type f -print0)
    fi
    
    # Scan each file
    for file in "${files_to_scan[@]}"; do
        if [[ -f "$file" ]]; then
            local matches
            if matches=$(grep -n -E "$pattern" "$file" 2>/dev/null); then
                while IFS= read -r match; do
                    local line_num=$(echo "$match" | cut -d: -f1)
                    local content=$(echo "$match" | cut -d: -f2-)
                    results+=("{\"file\":\"$file\",\"line\":$line_num,\"content\":\"$(echo "$content" | sed 's/"/\\"/g')\",\"pattern\":\"$pattern_name\"}")
                done <<< "$matches"
            fi
        fi
    done
    
    printf '%s\n' "${results[@]}"
}

# Function to process deprecated comments
process_deprecated_comments() {
    local file="$1"
    local line_num="$2"
    local content="$3"
    
    if [[ "$INTERACTIVE" == true && "$DRY_RUN" == false ]]; then
        echo -e "\n${YELLOW}Deprecated comment found:${NC}"
        echo -e "File: $file:$line_num"
        echo -e "Content: $content"
        echo
        read -p "Remove this line? (y/n/skip all): " choice
        case $choice in
            y|Y) return 0 ;;
            s|skip*) INTERACTIVE=false; return 1 ;;
            *) return 1 ;;
        esac
    fi
    
    return 0
}

# Function to process deprecated files
process_deprecated_files() {
    local file="$1"
    
    if [[ "$INTERACTIVE" == true && "$DRY_RUN" == false ]]; then
        echo -e "\n${YELLOW}Deprecated file found:${NC}"
        echo -e "File: $file"
        echo
        read -p "Delete this file? (y/n/skip all): " choice
        case $choice in
            y|Y) return 0 ;;
            s|skip*) INTERACTIVE=false; return 1 ;;
            *) return 1 ;;
        esac
    fi
    
    return 0
}

# Function to clean up deprecated items
cleanup_deprecated_items() {
    local results_json="$1"
    local cleanup_count=0
    local skipped_count=0
    
    print_header "Processing Deprecated Items"
    
    # Group results by file and pattern
    local current_file=""
    local items_to_remove=()
    
    while IFS= read -r result; do
        if [[ -z "$result" ]]; then continue; fi
        
        local file pattern line content
        file=$(echo "$result" | jq -r '.file')
        pattern=$(echo "$result" | jq -r '.pattern')
        line=$(echo "$result" | jq -r '.line')
        content=$(echo "$result" | jq -r '.content')
        
        case "$pattern" in
            "deprecated_comments")
                if process_deprecated_comments "$file" "$line" "$content"; then
                    if [[ "$DRY_RUN" == false ]]; then
                        create_backup "$file"
                        sed -i "${line}d" "$file"
                        print_success "Removed deprecated comment from $file:$line"
                        ((cleanup_count++))
                    else
                        print_info "Would remove deprecated comment from $file:$line"
                        ((cleanup_count++))
                    fi
                else
                    ((skipped_count++))
                fi
                ;;
            "deprecated_files"|"backup_files"|"temp_files")
                if process_deprecated_files "$file"; then
                    if [[ "$DRY_RUN" == false ]]; then
                        rm -f "$file"
                        print_success "Removed deprecated file: $file"
                        ((cleanup_count++))
                    else
                        print_info "Would remove deprecated file: $file"
                        ((cleanup_count++))
                    fi
                else
                    ((skipped_count++))
                fi
                ;;
            *)
                print_warning "Manual review needed for $pattern in $file:$line"
                ((skipped_count++))
                ;;
        esac
    done <<< "$results_json"
    
    echo
    print_success "Cleanup completed: $cleanup_count items processed, $skipped_count skipped"
}

# Function to generate comprehensive report
generate_report() {
    local all_results="$1"
    
    print_header "Generating Comprehensive Report"
    
    # Count results by pattern
    local pattern_counts=""
    for pattern_name in "${!DEPRECATED_PATTERNS[@]}"; do
        local count
        count=$(echo "$all_results" | grep "\"pattern\":\"$pattern_name\"" | wc -l)
        if [[ $count -gt 0 ]]; then
            pattern_counts="${pattern_counts}\"$pattern_name\":$count,"
        fi
    done
    pattern_counts="${pattern_counts%,}" # Remove trailing comma
    
    # Generate summary statistics
    local total_items
    total_items=$(echo "$all_results" | grep -c "\"file\":" || echo 0)
    
    local unique_files
    unique_files=$(echo "$all_results" | grep -o '"file":"[^"]*"' | sort -u | wc -l)
    
    # Create JSON report
    cat > "$REPORT_FILE" << EOF
{
    "scan_date": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "project_root": "$PROJECT_ROOT",
    "scan_options": {
        "dry_run": $DRY_RUN,
        "interactive": $INTERACTIVE,
        "pattern_filter": "${PATTERN_FILTER:-"all"}",
        "file_filter": "${FILE_FILTER:-"all"}"
    },
    "summary": {
        "total_deprecated_items": $total_items,
        "files_affected": $unique_files,
        "patterns_found": {$pattern_counts}
    },
    "deprecated_items": [
$(echo "$all_results" | sed 's/$/,/' | sed '$s/,$//')
    ],
    "recommendations": [
        "Review all deprecated comments and remove obsolete code",
        "Remove backup and temporary files to reduce clutter",
        "Update imports to use current packages",
        "Consider refactoring deprecated functions",
        "Update documentation to reflect current code structure"
    ]
}
EOF
    
    print_success "Report generated: $REPORT_FILE"
    
    # Print summary to console
    echo
    print_header "Scan Summary"
    echo "Total deprecated items found: $total_items"
    echo "Files affected: $unique_files"
    echo "Report saved to: $REPORT_FILE"
    
    if [[ $total_items -gt 0 ]]; then
        echo
        print_header "Items by Pattern"
        for pattern_name in "${!DEPRECATED_PATTERNS[@]}"; do
            local count
            count=$(echo "$all_results" | grep "\"pattern\":\"$pattern_name\"" | wc -l)
            if [[ $count -gt 0 ]]; then
                printf "  %-20s: %d items\n" "$pattern_name" "$count"
            fi
        done
    fi
}

# Function to integrate with technical debt monitor
integrate_with_debt_monitor() {
    local report_file="$1"
    
    if [[ -f "$PROJECT_ROOT/internal/observability/technical_debt_monitor.go" ]]; then
        print_info "Integrating with technical debt monitoring system..."
        
        # Calculate metrics for integration
        local total_items
        total_items=$(jq '.summary.total_deprecated_items' "$report_file")
        
        local files_affected
        files_affected=$(jq '.summary.files_affected' "$report_file")
        
        # Create integration data file
        local integration_file="$PROJECT_ROOT/tmp/cleanup-integration-$(date +%Y%m%d-%H%M%S).json"
        mkdir -p "$(dirname "$integration_file")"
        
        cat > "$integration_file" << EOF
{
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "deprecated_items_found": $total_items,
    "files_with_deprecated_code": $files_affected,
    "cleanup_run": true,
    "cleanup_report": "$report_file"
}
EOF
        
        print_success "Integration data created: $integration_file"
        print_info "This data can be consumed by the technical debt monitoring system"
    else
        print_warning "Technical debt monitoring system not found, skipping integration"
    fi
}

# Main execution function
main() {
    print_header "KYB Platform - Automated Deprecated Code Cleanup"
    print_info "Starting cleanup scan at $(date)"
    print_info "Project root: $PROJECT_ROOT"
    print_info "Mode: $([ "$DRY_RUN" == true ] && echo "DRY RUN" || echo "LIVE")"
    print_info "Interactive: $INTERACTIVE"
    
    # Initialize results
    local all_results=""
    
    # Scan for deprecated patterns
    if [[ -n "$PATTERN_FILTER" ]]; then
        print_info "Scanning for specific pattern: $PATTERN_FILTER"
        local pattern="${DEPRECATED_PATTERNS[$PATTERN_FILTER]}"
        local results
        results=$(scan_deprecated_patterns "$PATTERN_FILTER" "$pattern")
        all_results="$results"
    else
        print_info "Scanning for all deprecated patterns..."
        for pattern_name in "${!DEPRECATED_PATTERNS[@]}"; do
            local pattern="${DEPRECATED_PATTERNS[$pattern_name]}"
            local results
            results=$(scan_deprecated_patterns "$pattern_name" "$pattern")
            if [[ -n "$results" ]]; then
                all_results="$all_results"$'\n'"$results"
            fi
        done
    fi
    
    # Remove empty lines
    all_results=$(echo "$all_results" | grep -v '^$' || true)
    
    # Generate report
    generate_report "$all_results"
    
    # Process cleanup if not report-only
    if [[ "$REPORT_ONLY" == false && -n "$all_results" ]]; then
        cleanup_deprecated_items "$all_results"
    fi
    
    # Integrate with technical debt monitor
    integrate_with_debt_monitor "$REPORT_FILE"
    
    print_success "Cleanup process completed successfully"
    
    # Show next steps
    echo
    print_header "Next Steps"
    echo "1. Review the generated report: $REPORT_FILE"
    echo "2. Run with --dry-run first to preview changes"
    echo "3. Create backups before running live cleanup"
    echo "4. Consider running technical debt analysis: scripts/analyze-technical-debt.sh"
    echo "5. Update documentation after cleanup"
}

# Trap to ensure cleanup on exit
cleanup_on_exit() {
    print_info "Cleanup script terminated"
}
trap cleanup_on_exit EXIT

# Check dependencies
command -v jq >/dev/null 2>&1 || {
    print_error "jq is required but not installed. Please install jq first."
    exit 1
}

# Run main function
main "$@"
