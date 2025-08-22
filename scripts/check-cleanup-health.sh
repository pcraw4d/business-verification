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
