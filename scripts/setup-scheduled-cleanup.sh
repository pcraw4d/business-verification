#!/bin/bash

# Setup Scheduled Cleanup Workflow
# This script configures automated cleanup jobs for regular maintenance

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
CRON_LOG_FILE="$PROJECT_ROOT/logs/cleanup-cron.log"
CRON_ERROR_LOG="$PROJECT_ROOT/logs/cleanup-cron-error.log"

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

Setup scheduled cleanup workflows for the KYB Platform.

OPTIONS:
    --install-cron           Install cron jobs for automated cleanup
    --remove-cron            Remove cron jobs
    --list-cron              List current cron jobs
    --test-schedule          Test the cleanup schedule
    --setup-monitoring       Setup monitoring for cleanup jobs
    --help                   Show this help message

EXAMPLES:
    $0 --install-cron        # Install all scheduled cleanup jobs
    $0 --remove-cron         # Remove all cleanup cron jobs
    $0 --list-cron           # Show current cleanup cron jobs
    $0 --test-schedule       # Test the cleanup schedule

EOF
}

# Create necessary directories
setup_directories() {
    print_info "Setting up directories..."
    
    mkdir -p "$PROJECT_ROOT/logs"
    mkdir -p "$PROJECT_ROOT/reports"
    mkdir -p "$PROJECT_ROOT/tmp"
    
    # Create log files if they don't exist
    touch "$CRON_LOG_FILE"
    touch "$CRON_ERROR_LOG"
    
    print_success "Directories and log files created"
}

# Install cron jobs
install_cron_jobs() {
    print_header "Installing Scheduled Cleanup Jobs"
    
    # Get current user
    local current_user=$(whoami)
    print_info "Installing cron jobs for user: $current_user"
    
    # Create temporary cron file
    local temp_cron=$(mktemp)
    
    # Export current cron jobs
    crontab -l 2>/dev/null > "$temp_cron" || true
    
    # Add cleanup jobs
    cat >> "$temp_cron" << EOF

# KYB Platform - Automated Cleanup Schedule
# Daily cleanup (2 AM) - Safe auto-fixable items only
0 2 * * * cd "$PROJECT_ROOT" && ./scripts/run-cleanup.sh --go-only --pattern backup_files --auto-fix --non-interactive >> "$CRON_LOG_FILE" 2>> "$CRON_ERROR_LOG"

# Weekly cleanup (Sunday 1 AM) - Medium priority items
0 1 * * 0 cd "$PROJECT_ROOT" && ./scripts/run-cleanup.sh --go-only --pattern deprecated_comment --non-interactive >> "$CRON_LOG_FILE" 2>> "$CRON_ERROR_LOG"

# Monthly cleanup (1st of month, midnight) - Full scan with report
0 0 1 * * cd "$PROJECT_ROOT" && ./scripts/run-cleanup.sh --go-only --non-interactive --format html >> "$CRON_LOG_FILE" 2>> "$CRON_ERROR_LOG"

# Weekly cleanup report generation (Saturday 3 AM)
0 3 * * 6 cd "$PROJECT_ROOT" && ./scripts/generate-cleanup-report.sh >> "$CRON_LOG_FILE" 2>> "$CRON_ERROR_LOG"

# Daily cleanup health check (6 AM)
0 6 * * * cd "$PROJECT_ROOT" && ./scripts/check-cleanup-health.sh >> "$CRON_LOG_FILE" 2>> "$CRON_ERROR_LOG"
EOF
    
    # Install the new cron jobs
    crontab "$temp_cron"
    rm "$temp_cron"
    
    print_success "Cron jobs installed successfully"
    print_info "Log files: $CRON_LOG_FILE"
    print_info "Error log: $CRON_ERROR_LOG"
}

# Remove cron jobs
remove_cron_jobs() {
    print_header "Removing Scheduled Cleanup Jobs"
    
    # Create temporary cron file
    local temp_cron=$(mktemp)
    
    # Export current cron jobs and filter out cleanup jobs
    crontab -l 2>/dev/null | grep -v "KYB Platform" | grep -v "run-cleanup.sh" | grep -v "generate-cleanup-report.sh" | grep -v "check-cleanup-health.sh" > "$temp_cron" || true
    
    # Install the filtered cron jobs
    crontab "$temp_cron"
    rm "$temp_cron"
    
    print_success "Cleanup cron jobs removed successfully"
}

# List current cron jobs
list_cron_jobs() {
    print_header "Current Cleanup Cron Jobs"
    
    local cleanup_jobs=$(crontab -l 2>/dev/null | grep -E "(KYB Platform|run-cleanup|generate-cleanup|check-cleanup)" || echo "No cleanup jobs found")
    
    if [[ "$cleanup_jobs" == "No cleanup jobs found" ]]; then
        print_warning "No cleanup cron jobs are currently installed"
    else
        echo "$cleanup_jobs"
    fi
}

# Test the cleanup schedule
test_schedule() {
    print_header "Testing Cleanup Schedule"
    
    print_info "Testing daily cleanup (backup files)..."
    if cd "$PROJECT_ROOT" && ./scripts/run-cleanup.sh --go-only --pattern backup_files --auto-fix --non-interactive --dry-run; then
        print_success "Daily cleanup test passed"
    else
        print_error "Daily cleanup test failed"
        return 1
    fi
    
    print_info "Testing weekly cleanup (deprecated comments)..."
    if cd "$PROJECT_ROOT" && ./scripts/run-cleanup.sh --go-only --pattern deprecated_comment --non-interactive --dry-run; then
        print_success "Weekly cleanup test passed"
    else
        print_error "Weekly cleanup test failed"
        return 1
    fi
    
    print_success "All cleanup schedule tests passed"
}

# Setup monitoring
setup_monitoring() {
    print_header "Setting up Cleanup Monitoring"
    
    # Create monitoring script
    cat > "$PROJECT_ROOT/scripts/check-cleanup-health.sh" << 'EOF'
#!/bin/bash

# Cleanup Health Check Script
# Monitors the health of cleanup jobs and sends alerts if needed

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOG_FILE="$PROJECT_ROOT/logs/cleanup-cron.log"
ERROR_LOG="$PROJECT_ROOT/logs/cleanup-cron-error.log"
ALERT_THRESHOLD=5

# Check if cleanup jobs are running
check_cleanup_jobs() {
    local running_jobs=$(pgrep -f "run-cleanup.sh" | wc -l)
    if [[ $running_jobs -gt 0 ]]; then
        echo "WARNING: $running_jobs cleanup jobs are currently running"
        return 1
    fi
    return 0
}

# Check for recent errors
check_recent_errors() {
    local error_count=0
    if [[ -f "$ERROR_LOG" ]]; then
        error_count=$(tail -n 100 "$ERROR_LOG" | grep -c "ERROR\|FAILED\|FAIL" || echo 0)
    fi
    
    if [[ $error_count -gt $ALERT_THRESHOLD ]]; then
        echo "ALERT: $error_count errors detected in recent cleanup jobs"
        return 1
    fi
    return 0
}

# Check disk space
check_disk_space() {
    local disk_usage=$(df "$PROJECT_ROOT" | tail -1 | awk '{print $5}' | sed 's/%//')
    if [[ $disk_usage -gt 90 ]]; then
        echo "ALERT: Disk usage is ${disk_usage}% - cleanup may be needed"
        return 1
    fi
    return 0
}

# Check log file size
check_log_size() {
    if [[ -f "$LOG_FILE" ]]; then
        local log_size=$(stat -f%z "$LOG_FILE" 2>/dev/null || stat -c%s "$LOG_FILE" 2>/dev/null || echo 0)
        local max_size=$((100 * 1024 * 1024)) # 100MB
        
        if [[ $log_size -gt $max_size ]]; then
            echo "WARNING: Log file size is $(($log_size / 1024 / 1024))MB - consider rotation"
            return 1
        fi
    fi
    return 0
}

# Main health check
main() {
    local exit_code=0
    
    echo "$(date): Starting cleanup health check"
    
    check_cleanup_jobs || exit_code=1
    check_recent_errors || exit_code=1
    check_disk_space || exit_code=1
    check_log_size || exit_code=1
    
    if [[ $exit_code -eq 0 ]]; then
        echo "$(date): Cleanup health check passed"
    else
        echo "$(date): Cleanup health check failed"
    fi
    
    exit $exit_code
}

main "$@"
EOF
    
    # Create report generation script
    cat > "$PROJECT_ROOT/scripts/generate-cleanup-report.sh" << 'EOF'
#!/bin/bash

# Generate Cleanup Report Script
# Creates weekly cleanup reports for review

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPORTS_DIR="$PROJECT_ROOT/reports"
WEEKLY_REPORT="$REPORTS_DIR/weekly-cleanup-report-$(date +%Y%m%d).html"

# Generate weekly summary
generate_weekly_summary() {
    echo "Generating weekly cleanup summary..."
    
    # Run cleanup with HTML report
    cd "$PROJECT_ROOT"
    ./scripts/run-cleanup.sh --go-only --non-interactive --format html --output "$WEEKLY_REPORT"
    
    echo "Weekly report generated: $WEEKLY_REPORT"
}

# Send notification (placeholder for integration)
send_notification() {
    echo "Weekly cleanup report generated: $WEEKLY_REPORT"
    # TODO: Integrate with Slack, email, or other notification systems
}

main() {
    echo "$(date): Starting weekly cleanup report generation"
    
    generate_weekly_summary
    send_notification
    
    echo "$(date): Weekly cleanup report generation completed"
}

main "$@"
EOF
    
    # Make scripts executable
    chmod +x "$PROJECT_ROOT/scripts/check-cleanup-health.sh"
    chmod +x "$PROJECT_ROOT/scripts/generate-cleanup-report.sh"
    
    print_success "Monitoring scripts created"
    print_info "Health check script: scripts/check-cleanup-health.sh"
    print_info "Report generation script: scripts/generate-cleanup-report.sh"
}

# Create log rotation configuration
setup_log_rotation() {
    print_header "Setting up Log Rotation"
    
    # Create logrotate configuration
    local logrotate_conf="/etc/logrotate.d/kyb-cleanup"
    
    if [[ -w "/etc/logrotate.d" ]]; then
        cat > "$logrotate_conf" << EOF
$CRON_LOG_FILE {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 $(whoami) $(whoami)
    postrotate
        echo "\$(date): Log rotated" >> $CRON_LOG_FILE
    endscript
}

$CRON_ERROR_LOG {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 $(whoami) $(whoami)
    postrotate
        echo "\$(date): Error log rotated" >> $CRON_ERROR_LOG
    endscript
}
EOF
        
        print_success "Log rotation configured: $logrotate_conf"
    else
        print_warning "Cannot write to /etc/logrotate.d - log rotation not configured"
        print_info "Consider running with sudo or manually configure log rotation"
    fi
}

# Main execution function
main() {
    local install_cron=false
    local remove_cron=false
    local list_cron=false
    local test_schedule=false
    local setup_monitoring=false
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --install-cron)
                install_cron=true
                shift
                ;;
            --remove-cron)
                remove_cron=true
                shift
                ;;
            --list-cron)
                list_cron=true
                shift
                ;;
            --test-schedule)
                test_schedule=true
                shift
                ;;
            --setup-monitoring)
                setup_monitoring=true
                shift
                ;;
            --help)
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
    
    # Default action if no options specified
    if [[ "$install_cron" == false && "$remove_cron" == false && "$list_cron" == false && "$test_schedule" == false && "$setup_monitoring" == false ]]; then
        print_info "No options specified, running full setup..."
        install_cron=true
        setup_monitoring=true
        test_schedule=true
    fi
    
    # Setup directories
    setup_directories
    
    # Execute requested actions
    if [[ "$setup_monitoring" == true ]]; then
        setup_monitoring
        setup_log_rotation
    fi
    
    if [[ "$install_cron" == true ]]; then
        install_cron_jobs
    fi
    
    if [[ "$remove_cron" == true ]]; then
        remove_cron_jobs
    fi
    
    if [[ "$list_cron" == true ]]; then
        list_cron_jobs
    fi
    
    if [[ "$test_schedule" == true ]]; then
        test_schedule
    fi
    
    # Final status
    echo
    print_header "Setup Summary"
    echo "Project root: $PROJECT_ROOT"
    echo "Log files: $CRON_LOG_FILE"
    echo "Error log: $CRON_ERROR_LOG"
    echo "Reports directory: $PROJECT_ROOT/reports"
    
    if [[ "$install_cron" == true ]]; then
        echo
        print_success "Scheduled cleanup workflow setup completed!"
        echo "Cron jobs will run automatically according to the schedule:"
        echo "  - Daily (2 AM): Backup file cleanup"
        echo "  - Weekly (Sunday 1 AM): Deprecated comment review"
        echo "  - Monthly (1st, midnight): Full scan with report"
        echo "  - Weekly (Saturday 3 AM): Report generation"
        echo "  - Daily (6 AM): Health check"
    fi
    
    return 0
}

# Run main function
main "$@"
