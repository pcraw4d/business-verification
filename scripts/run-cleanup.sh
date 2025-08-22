#!/bin/bash

# Comprehensive Cleanup Runner Script
# This script orchestrates both the shell-based and Go-based cleanup tools

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
CONFIG_FILE="$PROJECT_ROOT/configs/cleanup-config.yaml"
REPORTS_DIR="$PROJECT_ROOT/reports"
LOGS_DIR="$PROJECT_ROOT/logs"

# Ensure required directories exist
mkdir -p "$REPORTS_DIR" "$LOGS_DIR"

# Default options
RUN_SHELL_CLEANUP=true
RUN_GO_CLEANUP=true
DRY_RUN=false
INTERACTIVE=true
AUTO_FIX=false
PATTERN_FILTER=""
SEVERITY_FILTER=""
OUTPUT_FORMAT="json"
GENERATE_SUMMARY=true

# Print functions
print_header() {
    echo -e "${BLUE}=== $1 ===${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1" >&2
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

# Usage function
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Comprehensive cleanup runner for the KYB Platform.

OPTIONS:
    -d, --dry-run           Perform a dry run without making changes
    -n, --non-interactive   Run without user interaction
    -a, --auto-fix          Automatically fix simple issues
    -p, --pattern PATTERN   Only process specific pattern
    -s, --severity LEVEL    Only process specific severity (critical,high,medium,low)
    -f, --format FORMAT     Output format (json,yaml,text,html)
    --shell-only            Run only shell-based cleanup
    --go-only               Run only Go-based cleanup
    --no-summary            Skip summary generation
    -h, --help              Show this help message

EXAMPLES:
    $0                      # Run full cleanup suite
    $0 --dry-run            # Preview what would be cleaned
    $0 --auto-fix --pattern backup_files
                           # Auto-fix only backup files
    $0 --severity critical --non-interactive
                           # Process only critical issues without interaction

EOF
}

# Parse command line arguments
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
        -a|--auto-fix)
            AUTO_FIX=true
            shift
            ;;
        -p|--pattern)
            PATTERN_FILTER="$2"
            shift 2
            ;;
        -s|--severity)
            SEVERITY_FILTER="$2"
            shift 2
            ;;
        -f|--format)
            OUTPUT_FORMAT="$2"
            shift 2
            ;;
        --shell-only)
            RUN_GO_CLEANUP=false
            shift
            ;;
        --go-only)
            RUN_SHELL_CLEANUP=false
            shift
            ;;
        --no-summary)
            GENERATE_SUMMARY=false
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

# Validate severity filter
if [[ -n "$SEVERITY_FILTER" ]]; then
    case "$SEVERITY_FILTER" in
        critical|high|medium|low) ;;
        *)
            print_error "Invalid severity: $SEVERITY_FILTER"
            echo "Valid severities: critical, high, medium, low"
            exit 1
            ;;
    esac
fi

# Check dependencies
check_dependencies() {
    print_info "Checking dependencies..."
    
    # Check for required tools
    local missing_deps=()
    
    if ! command -v jq >/dev/null 2>&1; then
        missing_deps+=("jq")
    fi
    
    if ! command -v go >/dev/null 2>&1 && [[ "$RUN_GO_CLEANUP" == true ]]; then
        missing_deps+=("go")
    fi
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        print_error "Missing dependencies: ${missing_deps[*]}"
        echo "Please install the missing dependencies before running."
        exit 1
    fi
    
    print_success "All dependencies are available"
}

# Build Go cleanup tool if needed
build_go_tool() {
    if [[ "$RUN_GO_CLEANUP" == false ]]; then
        return 0
    fi
    
    print_info "Building Go cleanup tool..."
    
    local go_tool_path="$PROJECT_ROOT/bin/cleanup"
    mkdir -p "$(dirname "$go_tool_path")"
    
    if ! go build -o "$go_tool_path" "$PROJECT_ROOT/cmd/cleanup"; then
        print_error "Failed to build Go cleanup tool"
        return 1
    fi
    
    print_success "Go cleanup tool built successfully"
    return 0
}

# Run shell-based cleanup
run_shell_cleanup() {
    if [[ "$RUN_SHELL_CLEANUP" == false ]]; then
        return 0
    fi
    
    print_header "Running Shell-based Cleanup"
    
    local shell_script="$SCRIPT_DIR/cleanup-deprecated-code.sh"
    if [[ ! -x "$shell_script" ]]; then
        print_error "Shell cleanup script not found or not executable: $shell_script"
        return 1
    fi
    
    local shell_args=()
    
    if [[ "$DRY_RUN" == true ]]; then
        shell_args+=("--dry-run")
    fi
    
    if [[ "$INTERACTIVE" == false ]]; then
        shell_args+=("--non-interactive")
    fi
    
    if [[ -n "$PATTERN_FILTER" ]]; then
        shell_args+=("--pattern" "$PATTERN_FILTER")
    fi
    
    local shell_report="$REPORTS_DIR/shell-cleanup-$(date +%Y%m%d-%H%M%S).json"
    
    print_info "Running shell cleanup with args: ${shell_args[*]}"
    
    if ! "$shell_script" "${shell_args[@]}" > "$shell_report.log" 2>&1; then
        print_warning "Shell cleanup encountered issues (check log: $shell_report.log)"
    else
        print_success "Shell cleanup completed successfully"
    fi
    
    return 0
}

# Run Go-based cleanup
run_go_cleanup() {
    if [[ "$RUN_GO_CLEANUP" == false ]]; then
        return 0
    fi
    
    print_header "Running Go-based Cleanup"
    
    local go_tool="$PROJECT_ROOT/bin/cleanup"
    if [[ ! -x "$go_tool" ]]; then
        print_error "Go cleanup tool not found: $go_tool"
        return 1
    fi
    
    local go_args=()
    go_args+=("--project" "$PROJECT_ROOT")
    
    if [[ "$DRY_RUN" == true ]]; then
        go_args+=("--dry-run")
    fi
    
    if [[ "$INTERACTIVE" == true ]]; then
        go_args+=("--interactive")
    fi
    
    if [[ "$AUTO_FIX" == true ]]; then
        go_args+=("--auto-fix")
    fi
    
    if [[ -n "$PATTERN_FILTER" ]]; then
        go_args+=("--pattern" "$PATTERN_FILTER")
    fi
    
    if [[ -n "$SEVERITY_FILTER" ]]; then
        go_args+=("--severity" "$SEVERITY_FILTER")
    fi
    
    go_args+=("--format" "$OUTPUT_FORMAT")
    
    local go_report="$REPORTS_DIR/go-cleanup-$(date +%Y%m%d-%H%M%S).json"
    go_args+=("--output" "$go_report")
    
    print_info "Running Go cleanup with args: ${go_args[*]}"
    
    if ! "$go_tool" "${go_args[@]}"; then
        print_error "Go cleanup failed"
        return 1
    fi
    
    print_success "Go cleanup completed successfully"
    echo "Report saved to: $go_report"
    
    return 0
}

# Generate combined summary
generate_summary() {
    if [[ "$GENERATE_SUMMARY" == false ]]; then
        return 0
    fi
    
    print_header "Generating Combined Summary"
    
    local summary_file="$REPORTS_DIR/cleanup-summary-$(date +%Y%m%d-%H%M%S).json"
    local total_items=0
    local total_files=0
    local critical_items=0
    local high_items=0
    local medium_items=0
    local low_items=0
    
    # Aggregate data from all reports
    for report_file in "$REPORTS_DIR"/*cleanup*.json; do
        if [[ -f "$report_file" && "$report_file" != "$summary_file" ]]; then
            if command -v jq >/dev/null 2>&1; then
                local items
                items=$(jq -r '.summary.total_items // .summary.total_deprecated_items // 0' "$report_file" 2>/dev/null || echo 0)
                total_items=$((total_items + items))
                
                local files
                files=$(jq -r '.summary.files_affected // 0' "$report_file" 2>/dev/null || echo 0)
                total_files=$((total_files + files))
                
                # Count by severity if available
                local critical
                critical=$(jq -r '.summary.items_by_severity.critical // 0' "$report_file" 2>/dev/null || echo 0)
                critical_items=$((critical_items + critical))
                
                local high
                high=$(jq -r '.summary.items_by_severity.high // 0' "$report_file" 2>/dev/null || echo 0)
                high_items=$((high_items + high))
                
                local medium
                medium=$(jq -r '.summary.items_by_severity.medium // 0' "$report_file" 2>/dev/null || echo 0)
                medium_items=$((medium_items + medium))
                
                local low
                low=$(jq -r '.summary.items_by_severity.low // 0' "$report_file" 2>/dev/null || echo 0)
                low_items=$((low_items + low))
            fi
        fi
    done
    
    # Generate combined summary
    cat > "$summary_file" << EOF
{
    "summary_date": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "project_root": "$PROJECT_ROOT",
    "cleanup_options": {
        "dry_run": $DRY_RUN,
        "interactive": $INTERACTIVE,
        "auto_fix": $AUTO_FIX,
        "pattern_filter": "${PATTERN_FILTER:-"all"}",
        "severity_filter": "${SEVERITY_FILTER:-"all"}",
        "shell_cleanup": $RUN_SHELL_CLEANUP,
        "go_cleanup": $RUN_GO_CLEANUP
    },
    "combined_summary": {
        "total_deprecated_items": $total_items,
        "total_files_affected": $total_files,
        "items_by_severity": {
            "critical": $critical_items,
            "high": $high_items,
            "medium": $medium_items,
            "low": $low_items
        }
    },
    "recommendations": [
        "Review individual tool reports for detailed findings",
        "Run with --dry-run first to preview changes",
        "Focus on critical and high severity items first",
        "Schedule regular cleanup sessions",
        "Consider setting up automated cleanup for low-risk items",
        "Integrate cleanup checks into CI/CD pipeline"
    ],
    "reports_generated": [
$(find "$REPORTS_DIR" -name "*cleanup*.json" -newer "$summary_file" 2>/dev/null | sed 's/.*/"&"/' | paste -sd, - || echo '""')
    ]
}
EOF
    
    print_success "Combined summary generated: $summary_file"
    
    # Print summary to console
    echo
    print_header "Cleanup Summary"
    echo "Total deprecated items found: $total_items"
    echo "Files affected: $total_files"
    echo "Critical items: $critical_items"
    echo "High severity items: $high_items"
    echo "Medium severity items: $medium_items"
    echo "Low severity items: $low_items"
    
    return 0
}

# Integrate with technical debt monitoring
integrate_with_monitoring() {
    print_header "Integrating with Technical Debt Monitoring"
    
    # Check if technical debt monitor is available
    if [[ -f "$PROJECT_ROOT/internal/observability/technical_debt_monitor.go" ]]; then
        print_info "Technical debt monitoring system detected"
        
        # Create integration trigger file
        local integration_file="$PROJECT_ROOT/tmp/cleanup-integration-$(date +%Y%m%d-%H%M%S).json"
        mkdir -p "$(dirname "$integration_file")"
        
        cat > "$integration_file" << EOF
{
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "cleanup_completed": true,
    "tools_used": {
        "shell_cleanup": $RUN_SHELL_CLEANUP,
        "go_cleanup": $RUN_GO_CLEANUP
    },
    "reports_directory": "$REPORTS_DIR"
}
EOF
        
        print_success "Integration trigger created: $integration_file"
        print_info "Technical debt monitor can process this data for metrics updates"
    else
        print_warning "Technical debt monitoring system not found"
    fi
}

# Main execution function
main() {
    print_header "KYB Platform - Comprehensive Cleanup Runner"
    print_info "Starting cleanup process at $(date)"
    print_info "Project root: $PROJECT_ROOT"
    print_info "Mode: $([ "$DRY_RUN" == true ] && echo "DRY RUN" || echo "LIVE")"
    print_info "Shell cleanup: $RUN_SHELL_CLEANUP"
    print_info "Go cleanup: $RUN_GO_CLEANUP"
    
    # Check dependencies
    check_dependencies
    
    # Build Go tool if needed
    if ! build_go_tool; then
        print_error "Failed to build Go cleanup tool"
        exit 1
    fi
    
    # Run cleanup tools
    local shell_success=true
    local go_success=true
    
    if ! run_shell_cleanup; then
        shell_success=false
        print_warning "Shell cleanup completed with issues"
    fi
    
    if ! run_go_cleanup; then
        go_success=false
        print_warning "Go cleanup failed"
    fi
    
    # Generate summary
    generate_summary
    
    # Integrate with monitoring
    integrate_with_monitoring
    
    # Final status
    echo
    if [[ "$shell_success" == true && "$go_success" == true ]]; then
        print_success "All cleanup tools completed successfully"
    elif [[ "$shell_success" == true || "$go_success" == true ]]; then
        print_warning "Cleanup completed with some issues"
    else
        print_error "Cleanup encountered significant issues"
        exit 1
    fi
    
    # Show next steps
    echo
    print_header "Next Steps"
    echo "1. Review cleanup reports in: $REPORTS_DIR"
    echo "2. Check logs for any issues in: $LOGS_DIR"
    echo "3. Run tests to ensure functionality after cleanup"
    echo "4. Consider scheduling regular cleanup runs"
    echo "5. Update documentation if significant changes were made"
    
    return 0
}

# Trap to ensure cleanup on exit
cleanup_on_exit() {
    print_info "Cleanup runner terminated"
}
trap cleanup_on_exit EXIT

# Run main function
main "$@"
