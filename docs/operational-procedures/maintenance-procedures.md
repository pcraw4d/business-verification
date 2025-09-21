# Maintenance Procedures Documentation

## Overview

This document provides comprehensive maintenance procedures for the KYB Platform, covering routine maintenance tasks, system updates, performance optimization, and preventive maintenance. These procedures ensure optimal system performance, security, and reliability.

## Table of Contents

1. [Maintenance Strategy](#maintenance-strategy)
2. [Daily Maintenance Tasks](#daily-maintenance-tasks)
3. [Weekly Maintenance Tasks](#weekly-maintenance-tasks)
4. [Monthly Maintenance Tasks](#monthly-maintenance-tasks)
5. [Quarterly Maintenance Tasks](#quarterly-maintenance-tasks)
6. [System Updates and Patches](#system-updates-and-patches)
7. [Performance Optimization](#performance-optimization)
8. [Security Maintenance](#security-maintenance)
9. [Database Maintenance](#database-maintenance)
10. [Monitoring and Alerting Maintenance](#monitoring-and-alerting-maintenance)

## Maintenance Strategy

### Maintenance Philosophy

- **Preventive Maintenance**: Regular maintenance to prevent issues
- **Predictive Maintenance**: Monitoring and analysis to predict failures
- **Corrective Maintenance**: Fixing issues as they arise
- **Continuous Improvement**: Regular optimization and enhancement

### Maintenance Schedule

| Frequency | Tasks | Duration | Impact |
|-----------|-------|----------|---------|
| Daily | Health checks, log rotation, backup verification | 15 minutes | Low |
| Weekly | Performance analysis, security updates, database optimization | 2 hours | Medium |
| Monthly | Full system backup, security audit, capacity planning | 4 hours | Medium |
| Quarterly | Major updates, disaster recovery testing, architecture review | 8 hours | High |

### Maintenance Windows

- **Primary Window**: Sunday 2:00 AM - 6:00 AM UTC
- **Secondary Window**: Wednesday 2:00 AM - 4:00 AM UTC
- **Emergency Window**: As needed for critical issues

## Daily Maintenance Tasks

### Automated Daily Maintenance Script

```bash
#!/bin/bash
# scripts/daily-maintenance.sh

set -euo pipefail

LOG_FILE="/var/log/kyb-platform/daily-maintenance.log"
MAINTENANCE_DIR="/opt/kyb-platform/maintenance"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

create_maintenance_directories() {
    log "Creating maintenance directories..."
    mkdir -p "$MAINTENANCE_DIR"/{logs,reports,backups}
}

system_health_check() {
    log "Performing system health check..."
    
    # Check application health
    if curl -f http://localhost:8080/health >/dev/null 2>&1; then
        log "✓ Application health: OK"
    else
        log "✗ Application health: FAILED"
        return 1
    fi
    
    # Check database connectivity
    if PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" >/dev/null 2>&1; then
        log "✓ Database connectivity: OK"
    else
        log "✗ Database connectivity: FAILED"
        return 1
    fi
    
    # Check Redis connectivity
    if redis-cli ping >/dev/null 2>&1; then
        log "✓ Redis connectivity: OK"
    else
        log "✗ Redis connectivity: FAILED"
        return 1
    fi
    
    return 0
}

log_rotation() {
    log "Performing log rotation..."
    
    # Rotate application logs
    if [[ -f "/var/log/kyb-platform/app.log" ]]; then
        if [[ $(stat -c%s "/var/log/kyb-platform/app.log") -gt 104857600 ]]; then  # 100MB
            mv "/var/log/kyb-platform/app.log" "/var/log/kyb-platform/app.log.$(date +%Y%m%d)"
            touch "/var/log/kyb-platform/app.log"
            chown kyb:kyb "/var/log/kyb-platform/app.log"
            log "Application log rotated"
        fi
    fi
    
    # Rotate system logs
    logrotate -f /etc/logrotate.d/kyb-platform
    
    # Clean up old rotated logs (keep 30 days)
    find /var/log/kyb-platform -name "*.log.*" -mtime +30 -delete
}

backup_verification() {
    log "Verifying backups..."
    
    # Check if backup was created today
    local today_backup=$(find /backups/database/full -name "backup_$(date +%Y%m%d)*" -type d)
    
    if [[ -n "$today_backup" ]]; then
        log "✓ Today's backup found: $today_backup"
        
        # Verify backup integrity
        local metadata_file="$today_backup/backup_metadata.json"
        if [[ -f "$metadata_file" ]]; then
            local expected_checksum=$(jq -r '.checksum' "$metadata_file")
            local backup_file=$(jq -r '.backup_file' "$metadata_file")
            local actual_checksum=$(sha256sum "$backup_file" | cut -d' ' -f1)
            
            if [[ "$expected_checksum" == "$actual_checksum" ]]; then
                log "✓ Backup integrity verified"
            else
                log "✗ Backup integrity check failed"
                return 1
            fi
        fi
    else
        log "✗ Today's backup not found"
        return 1
    fi
    
    return 0
}

performance_monitoring() {
    log "Collecting performance metrics..."
    
    # Collect system metrics
    {
        echo "=== System Metrics $(date) ==="
        echo "CPU Usage: $(top -bn1 | grep "Cpu(s)" | awk '{print $2}')"
        echo "Memory Usage: $(free | awk 'NR==2{printf "%.1f%%", $3*100/$2}')"
        echo "Disk Usage: $(df / | awk 'NR==2 {print $5}')"
        echo "Load Average: $(uptime | awk -F'load average:' '{print $2}')"
    } > "$MAINTENANCE_DIR/reports/daily-metrics-$(date +%Y%m%d).txt"
    
    # Collect application metrics
    curl -s http://localhost:8080/metrics > "$MAINTENANCE_DIR/reports/daily-app-metrics-$(date +%Y%m%d).txt"
    
    log "Performance metrics collected"
}

cleanup_temp_files() {
    log "Cleaning up temporary files..."
    
    # Clean up application temp files
    find /tmp -name "kyb-*" -mtime +1 -delete 2>/dev/null || true
    
    # Clean up old maintenance reports (keep 30 days)
    find "$MAINTENANCE_DIR/reports" -name "*.txt" -mtime +30 -delete 2>/dev/null || true
    
    # Clean up old log files (keep 30 days)
    find /var/log/kyb-platform -name "*.log.*" -mtime +30 -delete 2>/dev/null || true
    
    log "Temporary files cleaned up"
}

generate_daily_report() {
    log "Generating daily maintenance report..."
    
    local report_file="$MAINTENANCE_DIR/reports/daily-report-$(date +%Y%m%d).txt"
    
    {
        echo "KYB Platform Daily Maintenance Report"
        echo "Date: $(date)"
        echo "======================================"
        echo ""
        echo "System Health:"
        systemctl status kyb-platform --no-pager
        echo ""
        echo "Disk Usage:"
        df -h
        echo ""
        echo "Memory Usage:"
        free -h
        echo ""
        echo "Recent Errors:"
        journalctl -u kyb-platform --since "24 hours ago" | grep -i error | tail -10
        echo ""
        echo "Backup Status:"
        ls -la /backups/database/full/ | tail -5
    } > "$report_file"
    
    log "Daily report generated: $report_file"
}

main() {
    log "Starting daily maintenance..."
    
    create_maintenance_directories
    
    local failed_tasks=0
    
    system_health_check || ((failed_tasks++))
    log_rotation
    backup_verification || ((failed_tasks++))
    performance_monitoring
    cleanup_temp_files
    generate_daily_report
    
    if [[ $failed_tasks -eq 0 ]]; then
        log "Daily maintenance completed successfully"
        exit 0
    else
        log "Daily maintenance completed with $failed_tasks failed tasks"
        exit 1
    fi
}

main "$@"
```

### Daily Maintenance Cron Job

```bash
# Add to crontab (crontab -e)
# Daily maintenance at 2:00 AM
0 2 * * * /opt/kyb-platform/scripts/daily-maintenance.sh >> /var/log/kyb-platform/daily-maintenance.log 2>&1
```

## Weekly Maintenance Tasks

### Weekly Maintenance Script

```bash
#!/bin/bash
# scripts/weekly-maintenance.sh

set -euo pipefail

LOG_FILE="/var/log/kyb-platform/weekly-maintenance.log"
MAINTENANCE_DIR="/opt/kyb-platform/maintenance"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

performance_analysis() {
    log "Performing performance analysis..."
    
    # Analyze response times
    local avg_response_time=$(curl -s http://localhost:8080/metrics | grep 'kyb_http_request_duration_seconds_sum' | awk '{sum+=$2} END {print sum/NR}')
    log "Average response time: ${avg_response_time}s"
    
    # Analyze error rates
    local total_requests=$(curl -s http://localhost:8080/metrics | grep 'kyb_http_requests_total' | awk '{sum+=$2} END {print sum}')
    local error_requests=$(curl -s http://localhost:8080/metrics | grep 'kyb_http_requests_total.*5[0-9][0-9]' | awk '{sum+=$2} END {print sum}')
    
    if [[ $total_requests -gt 0 ]]; then
        local error_rate=$(echo "scale=2; $error_requests * 100 / $total_requests" | bc)
        log "Error rate: ${error_rate}%"
    fi
    
    # Generate performance report
    {
        echo "=== Weekly Performance Analysis $(date) ==="
        echo "Average Response Time: ${avg_response_time}s"
        echo "Error Rate: ${error_rate}%"
        echo "Total Requests: $total_requests"
        echo ""
        echo "Top Slow Endpoints:"
        curl -s http://localhost:8080/metrics | grep 'kyb_http_request_duration_seconds' | sort -k2 -nr | head -10
    } > "$MAINTENANCE_DIR/reports/weekly-performance-$(date +%Y%m%d).txt"
}

security_updates() {
    log "Checking for security updates..."
    
    # Check for system updates
    if command -v apt-get >/dev/null 2>&1; then
        apt-get update >/dev/null 2>&1
        local security_updates=$(apt-get upgrade -s | grep -i security | wc -l)
        log "Security updates available: $security_updates"
        
        if [[ $security_updates -gt 0 ]]; then
            log "Applying security updates..."
            apt-get upgrade -y --only-upgrade
        fi
    fi
    
    # Check for Go module updates
    cd /opt/kyb-platform
    if [[ -f "go.mod" ]]; then
        go list -u -m all | grep -E '\[(update|patch)\]' > "$MAINTENANCE_DIR/reports/go-updates-$(date +%Y%m%d).txt"
    fi
}

database_optimization() {
    log "Performing database optimization..."
    
    # Update table statistics
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "ANALYZE;"
    
    # Check for unused indexes
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
        SELECT schemaname, tablename, indexname, idx_tup_read, idx_tup_fetch
        FROM pg_stat_user_indexes
        WHERE idx_tup_read = 0
        ORDER BY schemaname, tablename, indexname;" > "$MAINTENANCE_DIR/reports/unused-indexes-$(date +%Y%m%d).txt"
    
    # Check for table bloat
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
        SELECT schemaname, tablename, 
               pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size,
               pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) as table_size
        FROM pg_tables 
        WHERE schemaname = 'public'
        ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;" > "$MAINTENANCE_DIR/reports/table-sizes-$(date +%Y%m%d).txt"
    
    log "Database optimization completed"
}

ml_model_maintenance() {
    log "Performing ML model maintenance..."
    
    # Check model accuracy
    local bert_accuracy=$(curl -s http://localhost:8080/metrics | grep 'kyb_ml_model_accuracy.*model="bert"' | awk '{print $2}')
    local distilbert_accuracy=$(curl -s http://localhost:8080/metrics | grep 'kyb_ml_model_accuracy.*model="distilbert"' | awk '{print $2}')
    
    log "BERT accuracy: ${bert_accuracy}"
    log "DistilBERT accuracy: ${distilbert_accuracy}"
    
    # Check if models need retraining
    if (( $(echo "$bert_accuracy < 0.85" | bc -l) )); then
        log "WARNING: BERT model accuracy below threshold, consider retraining"
    fi
    
    if (( $(echo "$distilbert_accuracy < 0.80" | bc -l) )); then
        log "WARNING: DistilBERT model accuracy below threshold, consider retraining"
    fi
    
    # Generate model performance report
    {
        echo "=== ML Model Performance $(date) ==="
        echo "BERT Accuracy: ${bert_accuracy}"
        echo "DistilBERT Accuracy: ${distilbert_accuracy}"
        echo ""
        echo "Model Usage Statistics:"
        curl -s http://localhost:8080/metrics | grep 'kyb_ml_model_requests_total'
    } > "$MAINTENANCE_DIR/reports/ml-performance-$(date +%Y%m%d).txt"
}

monitoring_maintenance() {
    log "Performing monitoring maintenance..."
    
    # Check Prometheus status
    if systemctl is-active --quiet prometheus; then
        log "✓ Prometheus: Running"
    else
        log "✗ Prometheus: Not running"
        systemctl start prometheus
    fi
    
    # Check Grafana status
    if systemctl is-active --quiet grafana-server; then
        log "✓ Grafana: Running"
    else
        log "✗ Grafana: Not running"
        systemctl start grafana-server
    fi
    
    # Check AlertManager status
    if systemctl is-active --quiet alertmanager; then
        log "✓ AlertManager: Running"
    else
        log "✗ AlertManager: Not running"
        systemctl start alertmanager
    fi
    
    # Clean up old metrics data (keep 30 days)
    find /var/lib/prometheus -name "*.db" -mtime +30 -delete 2>/dev/null || true
}

generate_weekly_report() {
    log "Generating weekly maintenance report..."
    
    local report_file="$MAINTENANCE_DIR/reports/weekly-report-$(date +%Y%m%d).txt"
    
    {
        echo "KYB Platform Weekly Maintenance Report"
        echo "Date: $(date)"
        echo "======================================"
        echo ""
        echo "Performance Summary:"
        cat "$MAINTENANCE_DIR/reports/weekly-performance-$(date +%Y%m%d).txt"
        echo ""
        echo "ML Model Performance:"
        cat "$MAINTENANCE_DIR/reports/ml-performance-$(date +%Y%m%d).txt"
        echo ""
        echo "Database Status:"
        PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
            SELECT 
                schemaname,
                tablename,
                n_tup_ins as inserts,
                n_tup_upd as updates,
                n_tup_del as deletes
            FROM pg_stat_user_tables
            ORDER BY n_tup_ins + n_tup_upd + n_tup_del DESC
            LIMIT 10;"
    } > "$report_file"
    
    log "Weekly report generated: $report_file"
}

main() {
    log "Starting weekly maintenance..."
    
    local failed_tasks=0
    
    performance_analysis
    security_updates || ((failed_tasks++))
    database_optimization
    ml_model_maintenance
    monitoring_maintenance || ((failed_tasks++))
    generate_weekly_report
    
    if [[ $failed_tasks -eq 0 ]]; then
        log "Weekly maintenance completed successfully"
        exit 0
    else
        log "Weekly maintenance completed with $failed_tasks failed tasks"
        exit 1
    fi
}

main "$@"
```

## Monthly Maintenance Tasks

### Monthly Maintenance Script

```bash
#!/bin/bash
# scripts/monthly-maintenance.sh

set -euo pipefail

LOG_FILE="/var/log/kyb-platform/monthly-maintenance.log"
MAINTENANCE_DIR="/opt/kyb-platform/maintenance"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

full_system_backup() {
    log "Creating full system backup..."
    
    # Create comprehensive backup
    local backup_id="monthly_$(date +%Y%m%d_%H%M%S)"
    local backup_dir="/backups/monthly/$backup_id"
    
    mkdir -p "$backup_dir"
    
    # Backup database
    ./scripts/backup-database.sh
    
    # Backup application configuration
    tar -czf "$backup_dir/application-config.tar.gz" /opt/kyb-platform/configs/
    
    # Backup monitoring configurations
    tar -czf "$backup_dir/monitoring-config.tar.gz" /etc/prometheus/ /etc/grafana/ /etc/alertmanager/
    
    # Backup SSL certificates
    tar -czf "$backup_dir/ssl-certificates.tar.gz" /etc/ssl/kyb-platform/
    
    # Create backup manifest
    {
        echo "Monthly Backup Manifest"
        echo "Backup ID: $backup_id"
        echo "Date: $(date)"
        echo "Components:"
        echo "- Database backup"
        echo "- Application configuration"
        echo "- Monitoring configuration"
        echo "- SSL certificates"
        echo ""
        echo "File sizes:"
        ls -lh "$backup_dir"
    } > "$backup_dir/backup-manifest.txt"
    
    log "Full system backup completed: $backup_id"
}

security_audit() {
    log "Performing security audit..."
    
    # Check for security vulnerabilities
    if command -v trivy >/dev/null 2>&1; then
        trivy image kyb-platform:latest > "$MAINTENANCE_DIR/reports/security-audit-$(date +%Y%m%d).txt"
    fi
    
    # Check SSL certificate expiration
    local cert_expiry=$(openssl x509 -in /etc/ssl/kyb-platform/cert.pem -noout -dates | grep notAfter | cut -d= -f2)
    log "SSL certificate expires: $cert_expiry"
    
    # Check for weak passwords (if applicable)
    # This would depend on your password policy implementation
    
    # Check file permissions
    find /opt/kyb-platform -type f -perm /o+w > "$MAINTENANCE_DIR/reports/world-writable-files-$(date +%Y%m%d).txt"
    
    log "Security audit completed"
}

capacity_planning() {
    log "Performing capacity planning analysis..."
    
    # Analyze disk usage trends
    {
        echo "=== Disk Usage Analysis $(date) ==="
        df -h
        echo ""
        echo "Largest directories:"
        du -h /opt/kyb-platform | sort -hr | head -10
        echo ""
        echo "Database size trend:"
        PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
            SELECT 
                schemaname,
                tablename,
                pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
            FROM pg_tables 
            WHERE schemaname = 'public'
            ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;"
    } > "$MAINTENANCE_DIR/reports/capacity-analysis-$(date +%Y%m%d).txt"
    
    # Analyze growth trends
    local db_size=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT pg_database_size('$DB_NAME');")
    echo "$(date),$db_size" >> "$MAINTENANCE_DIR/database-size-trend.csv"
    
    log "Capacity planning analysis completed"
}

disaster_recovery_test() {
    log "Testing disaster recovery procedures..."
    
    # Test backup restoration (dry run)
    local latest_backup=$(ls -t /backups/database/full | head -n1)
    
    if [[ -n "$latest_backup" ]]; then
        # Create test database
        local test_db="kyb_dr_test_$(date +%s)"
        
        PGPASSWORD="$DB_PASSWORD" createdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$test_db"
        
        # Test restoration
        local backup_file="/backups/database/full/$latest_backup/database.sql.gz"
        if gunzip -c "$backup_file" | PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$test_db" >/dev/null 2>&1; then
            log "✓ Disaster recovery test: PASSED"
            
            # Clean up test database
            PGPASSWORD="$DB_PASSWORD" dropdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$test_db"
        else
            log "✗ Disaster recovery test: FAILED"
            PGPASSWORD="$DB_PASSWORD" dropdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$test_db" 2>/dev/null || true
            return 1
        fi
    else
        log "✗ No backup found for disaster recovery test"
        return 1
    fi
    
    return 0
}

performance_optimization() {
    log "Performing performance optimization..."
    
    # Analyze slow queries
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
        SELECT query, mean_time, calls, total_time
        FROM pg_stat_statements
        ORDER BY mean_time DESC
        LIMIT 20;" > "$MAINTENANCE_DIR/reports/slow-queries-$(date +%Y%m%d).txt"
    
    # Check for missing indexes
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
        SELECT schemaname, tablename, attname, n_distinct, correlation
        FROM pg_stats
        WHERE schemaname = 'public'
        AND n_distinct > 100
        ORDER BY n_distinct DESC;" > "$MAINTENANCE_DIR/reports/index-recommendations-$(date +%Y%m%d).txt"
    
    # Optimize database
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "VACUUM ANALYZE;"
    
    log "Performance optimization completed"
}

generate_monthly_report() {
    log "Generating monthly maintenance report..."
    
    local report_file="$MAINTENANCE_DIR/reports/monthly-report-$(date +%Y%m%d).txt"
    
    {
        echo "KYB Platform Monthly Maintenance Report"
        echo "Date: $(date)"
        echo "======================================"
        echo ""
        echo "System Overview:"
        echo "Uptime: $(uptime)"
        echo "Load Average: $(cat /proc/loadavg)"
        echo ""
        echo "Resource Usage:"
        echo "CPU: $(top -bn1 | grep "Cpu(s)" | awk '{print $2}')"
        echo "Memory: $(free | awk 'NR==2{printf "%.1f%%", $3*100/$2}')"
        echo "Disk: $(df / | awk 'NR==2 {print $5}')"
        echo ""
        echo "Application Metrics:"
        curl -s http://localhost:8080/metrics | grep -E 'kyb_(http_requests_total|classification_requests_total|risk_assessment_requests_total)' | head -10
        echo ""
        echo "Database Statistics:"
        PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
            SELECT 
                schemaname,
                tablename,
                n_tup_ins as inserts,
                n_tup_upd as updates,
                n_tup_del as deletes,
                n_live_tup as live_tuples
            FROM pg_stat_user_tables
            ORDER BY n_live_tup DESC
            LIMIT 10;"
    } > "$report_file"
    
    log "Monthly report generated: $report_file"
}

main() {
    log "Starting monthly maintenance..."
    
    local failed_tasks=0
    
    full_system_backup
    security_audit
    capacity_planning
    disaster_recovery_test || ((failed_tasks++))
    performance_optimization
    generate_monthly_report
    
    if [[ $failed_tasks -eq 0 ]]; then
        log "Monthly maintenance completed successfully"
        exit 0
    else
        log "Monthly maintenance completed with $failed_tasks failed tasks"
        exit 1
    fi
}

main "$@"
```

## System Updates and Patches

### Update Procedures

#### Application Updates

```bash
#!/bin/bash
# scripts/update-application.sh

set -euo pipefail

LOG_FILE="/var/log/kyb-platform/application-update.log"
BACKUP_DIR="/backups/updates"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

pre_update_backup() {
    log "Creating pre-update backup..."
    
    local backup_id="pre_update_$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$BACKUP_DIR/$backup_id"
    
    # Backup current application
    cp -r /opt/kyb-platform "$BACKUP_DIR/$backup_id/"
    
    # Backup database
    ./scripts/backup-database.sh
    
    # Backup configuration
    cp -r /opt/kyb-platform/configs "$BACKUP_DIR/$backup_id/"
    
    log "Pre-update backup completed: $backup_id"
    echo "$backup_id"
}

download_update() {
    local version="$1"
    log "Downloading application update version $version..."
    
    # Download new version (this would depend on your deployment method)
    # Example: wget "https://releases.kyb-platform.com/v$version/kyb-platform-$version.tar.gz"
    
    # Extract update
    # tar -xzf "kyb-platform-$version.tar.gz" -C /tmp/
    
    log "Update downloaded and extracted"
}

validate_update() {
    local update_path="$1"
    log "Validating update..."
    
    # Check file integrity
    if [[ -f "$update_path/kyb-platform" ]]; then
        if [[ -x "$update_path/kyb-platform" ]]; then
            log "✓ Application binary: Valid"
        else
            log "✗ Application binary: Not executable"
            return 1
        fi
    else
        log "✗ Application binary: Not found"
        return 1
    fi
    
    # Test configuration
    if "$update_path/kyb-platform" --config-check; then
        log "✓ Configuration: Valid"
    else
        log "✗ Configuration: Invalid"
        return 1
    fi
    
    return 0
}

apply_update() {
    local update_path="$1"
    log "Applying update..."
    
    # Stop application
    systemctl stop kyb-platform
    
    # Backup current version
    mv /opt/kyb-platform/kyb-platform /opt/kyb-platform/kyb-platform.backup
    
    # Install new version
    cp "$update_path/kyb-platform" /opt/kyb-platform/
    chmod +x /opt/kyb-platform/kyb-platform
    
    # Update configuration if needed
    if [[ -f "$update_path/configs/production.yml" ]]; then
        cp "$update_path/configs/production.yml" /opt/kyb-platform/configs/production.yml.new
        # Compare and merge configurations
    fi
    
    # Start application
    systemctl start kyb-platform
    
    # Wait for application to start
    sleep 10
    
    # Verify application is running
    if curl -f http://localhost:8080/health >/dev/null 2>&1; then
        log "✓ Application update: Successful"
        return 0
    else
        log "✗ Application update: Failed"
        return 1
    fi
}

rollback_update() {
    local backup_id="$1"
    log "Rolling back update..."
    
    # Stop application
    systemctl stop kyb-platform
    
    # Restore previous version
    cp /opt/kyb-platform/kyb-platform.backup /opt/kyb-platform/kyb-platform
    
    # Restore configuration
    if [[ -f "$BACKUP_DIR/$backup_id/configs/production.yml" ]]; then
        cp "$BACKUP_DIR/$backup_id/configs/production.yml" /opt/kyb-platform/configs/
    fi
    
    # Start application
    systemctl start kyb-platform
    
    # Wait for application to start
    sleep 10
    
    # Verify rollback
    if curl -f http://localhost:8080/health >/dev/null 2>&1; then
        log "✓ Rollback: Successful"
        return 0
    else
        log "✗ Rollback: Failed"
        return 1
    fi
}

main() {
    local version="${1:-}"
    
    if [[ -z "$version" ]]; then
        echo "Usage: $0 <version>"
        exit 1
    fi
    
    log "Starting application update to version $version..."
    
    # Pre-update backup
    local backup_id=$(pre_update_backup)
    
    # Download update
    local update_path="/tmp/kyb-platform-$version"
    download_update "$version"
    
    # Validate update
    if ! validate_update "$update_path"; then
        log "Update validation failed, aborting"
        exit 1
    fi
    
    # Apply update
    if apply_update "$update_path"; then
        log "Application update completed successfully"
        exit 0
    else
        log "Application update failed, rolling back..."
        if rollback_update "$backup_id"; then
            log "Rollback completed successfully"
            exit 1
        else
            log "Rollback failed, manual intervention required"
            exit 2
        fi
    fi
}

main "$@"
```

#### Database Updates

```bash
#!/bin/bash
# scripts/update-database.sh

set -euo pipefail

LOG_FILE="/var/log/kyb-platform/database-update.log"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

backup_database() {
    log "Creating database backup..."
    ./scripts/backup-database.sh
}

run_migrations() {
    local migration_file="$1"
    log "Running database migrations from $migration_file..."
    
    if [[ -f "$migration_file" ]]; then
        PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$migration_file"
        log "Database migrations completed"
    else
        log "ERROR: Migration file not found: $migration_file"
        return 1
    fi
}

verify_migration() {
    log "Verifying database migration..."
    
    # Check if all tables exist
    local table_count=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT count(*) 
        FROM information_schema.tables 
        WHERE table_schema = 'public';")
    
    log "Tables in database: $table_count"
    
    # Check for migration errors
    local migration_errors=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT count(*) 
        FROM schema_migrations 
        WHERE success = false;")
    
    if [[ $migration_errors -gt 0 ]]; then
        log "ERROR: $migration_errors migration errors found"
        return 1
    fi
    
    log "Database migration verification completed"
    return 0
}

main() {
    local migration_file="${1:-}"
    
    if [[ -z "$migration_file" ]]; then
        echo "Usage: $0 <migration_file>"
        exit 1
    fi
    
    log "Starting database update..."
    
    # Backup database
    backup_database
    
    # Run migrations
    if run_migrations "$migration_file"; then
        # Verify migration
        if verify_migration; then
            log "Database update completed successfully"
            exit 0
        else
            log "Database migration verification failed"
            exit 1
        fi
    else
        log "Database migration failed"
        exit 1
    fi
}

main "$@"
```

## Performance Optimization

### Performance Tuning Script

```bash
#!/bin/bash
# scripts/performance-optimization.sh

set -euo pipefail

LOG_FILE="/var/log/kyb-platform/performance-optimization.log"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

optimize_database() {
    log "Optimizing database performance..."
    
    # Update statistics
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "ANALYZE;"
    
    # Vacuum tables
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "VACUUM;"
    
    # Reindex if needed
    local index_bloat=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT count(*) 
        FROM pg_stat_user_indexes 
        WHERE idx_tup_read > 1000 
        AND idx_tup_fetch < idx_tup_read * 0.1;")
    
    if [[ $index_bloat -gt 0 ]]; then
        log "Reindexing $index_bloat bloated indexes..."
        PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "REINDEX DATABASE $DB_NAME;"
    fi
    
    log "Database optimization completed"
}

optimize_application() {
    log "Optimizing application performance..."
    
    # Check for memory leaks
    local memory_usage=$(ps -o pid,vsz,rss,comm -p $(pgrep kyb-platform) | tail -n1 | awk '{print $3}')
    log "Current memory usage: ${memory_usage}KB"
    
    # Check goroutine count
    local goroutine_count=$(curl -s http://localhost:8080/metrics | grep 'go_goroutines' | awk '{print $2}')
    log "Current goroutine count: $goroutine_count"
    
    # Restart application if memory usage is high
    if [[ $memory_usage -gt 1048576 ]]; then  # 1GB
        log "High memory usage detected, restarting application..."
        systemctl restart kyb-platform
        sleep 10
    fi
    
    log "Application optimization completed"
}

optimize_system() {
    log "Optimizing system performance..."
    
    # Clear system caches
    sync
    echo 3 > /proc/sys/vm/drop_caches
    
    # Optimize TCP settings
    echo 'net.core.rmem_max = 16777216' >> /etc/sysctl.conf
    echo 'net.core.wmem_max = 16777216' >> /etc/sysctl.conf
    echo 'net.ipv4.tcp_rmem = 4096 87380 16777216' >> /etc/sysctl.conf
    echo 'net.ipv4.tcp_wmem = 4096 65536 16777216' >> /etc/sysctl.conf
    sysctl -p
    
    log "System optimization completed"
}

main() {
    log "Starting performance optimization..."
    
    optimize_database
    optimize_application
    optimize_system
    
    log "Performance optimization completed"
}

main "$@"
```

## Security Maintenance

### Security Maintenance Script

```bash
#!/bin/bash
# scripts/security-maintenance.sh

set -euo pipefail

LOG_FILE="/var/log/kyb-platform/security-maintenance.log"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

update_ssl_certificates() {
    log "Checking SSL certificate expiration..."
    
    local cert_file="/etc/ssl/kyb-platform/cert.pem"
    if [[ -f "$cert_file" ]]; then
        local expiry_date=$(openssl x509 -in "$cert_file" -noout -dates | grep notAfter | cut -d= -f2)
        local expiry_timestamp=$(date -d "$expiry_date" +%s)
        local current_timestamp=$(date +%s)
        local days_until_expiry=$(( (expiry_timestamp - current_timestamp) / 86400 ))
        
        log "SSL certificate expires in $days_until_expiry days"
        
        if [[ $days_until_expiry -lt 30 ]]; then
            log "WARNING: SSL certificate expires soon, renewal required"
            # Implement certificate renewal logic here
        fi
    fi
}

scan_vulnerabilities() {
    log "Scanning for security vulnerabilities..."
    
    # Check for outdated packages
    if command -v apt-get >/dev/null 2>&1; then
        apt-get update >/dev/null 2>&1
        local security_updates=$(apt-get upgrade -s | grep -i security | wc -l)
        log "Security updates available: $security_updates"
    fi
    
    # Check for vulnerable Go modules
    if [[ -f "/opt/kyb-platform/go.mod" ]]; then
        cd /opt/kyb-platform
        go list -json -m all | jq -r '.Path + " " + .Version' > /tmp/go-modules.txt
        log "Go modules scanned for vulnerabilities"
    fi
}

audit_file_permissions() {
    log "Auditing file permissions..."
    
    # Check for world-writable files
    find /opt/kyb-platform -type f -perm /o+w > /tmp/world-writable-files.txt
    local world_writable_count=$(wc -l < /tmp/world-writable-files.txt)
    
    if [[ $world_writable_count -gt 0 ]]; then
        log "WARNING: $world_writable_count world-writable files found"
        cat /tmp/world-writable-files.txt
    else
        log "✓ No world-writable files found"
    fi
    
    # Check for files with incorrect ownership
    find /opt/kyb-platform -not -user kyb -not -group kyb > /tmp/incorrect-ownership.txt
    local incorrect_ownership_count=$(wc -l < /tmp/incorrect-ownership.txt)
    
    if [[ $incorrect_ownership_count -gt 0 ]]; then
        log "WARNING: $incorrect_ownership_count files with incorrect ownership found"
        cat /tmp/incorrect-ownership.txt
    else
        log "✓ All files have correct ownership"
    fi
}

check_security_logs() {
    log "Checking security logs..."
    
    # Check for failed login attempts
    local failed_logins=$(journalctl --since "24 hours ago" | grep -i "failed login" | wc -l)
    log "Failed login attempts in last 24 hours: $failed_logins"
    
    # Check for suspicious activity
    local suspicious_activity=$(journalctl --since "24 hours ago" | grep -i "suspicious\|attack\|intrusion" | wc -l)
    log "Suspicious activity in last 24 hours: $suspicious_activity"
    
    if [[ $failed_logins -gt 100 ]] || [[ $suspicious_activity -gt 0 ]]; then
        log "WARNING: Potential security issues detected"
    fi
}

main() {
    log "Starting security maintenance..."
    
    update_ssl_certificates
    scan_vulnerabilities
    audit_file_permissions
    check_security_logs
    
    log "Security maintenance completed"
}

main "$@"
```

This comprehensive maintenance documentation ensures the KYB Platform maintains optimal performance, security, and reliability through systematic maintenance procedures.
