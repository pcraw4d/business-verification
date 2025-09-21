# Backup Procedures Documentation

## Overview

This document provides comprehensive backup procedures for the KYB Platform, ensuring data safety and business continuity. These procedures cover database backups, application state backups, configuration backups, and disaster recovery preparation.

## Table of Contents

1. [Backup Strategy](#backup-strategy)
2. [Database Backup Procedures](#database-backup-procedures)
3. [Application Backup Procedures](#application-backup-procedures)
4. [Configuration Backup Procedures](#configuration-backup-procedures)
5. [Backup Verification and Testing](#backup-verification-and-testing)
6. [Recovery Procedures](#recovery-procedures)
7. [Backup Monitoring and Alerting](#backup-monitoring-and-alerting)
8. [Maintenance and Cleanup](#maintenance-and-cleanup)

## Backup Strategy

### Backup Types

#### 1. Full Database Backup
- **Frequency**: Daily at 2:00 AM UTC
- **Retention**: 30 days
- **Scope**: Complete database including all tables, indexes, and constraints
- **Format**: Compressed SQL dump

#### 2. Incremental Database Backup
- **Frequency**: Every 6 hours
- **Retention**: 7 days
- **Scope**: Changes since last full backup
- **Format**: WAL (Write-Ahead Log) files

#### 3. Application State Backup
- **Frequency**: Before deployments
- **Retention**: 10 deployments
- **Scope**: Application configuration, feature flags, user sessions
- **Format**: JSON configuration files

#### 4. Configuration Backup
- **Frequency**: Daily
- **Retention**: 90 days
- **Scope**: Environment variables, configuration files, secrets
- **Format**: Encrypted configuration archives

### Backup Locations

```
/backups/
├── database/
│   ├── full/           # Full database backups
│   ├── incremental/    # Incremental backups
│   └── verification/   # Backup verification logs
├── application/
│   ├── config/         # Application configurations
│   ├── state/          # Application state snapshots
│   └── logs/           # Application logs
└── monitoring/
    ├── metrics/        # Monitoring data
    └── alerts/         # Alert configurations
```

## Database Backup Procedures

### Automated Backup Script

```bash
#!/bin/bash
# scripts/backup-database.sh

set -euo pipefail

# Configuration
BACKUP_DIR="/backups/database"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_ID="backup_${TIMESTAMP}"
LOG_FILE="/var/log/kyb-platform/backup.log"

# Environment variables
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-kyb_platform}"
DB_USER="${DB_USER:-kyb_user}"

# Create backup directory
mkdir -p "${BACKUP_DIR}/full/${BACKUP_ID}"

# Log function
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Pre-backup validation
validate_environment() {
    log "Starting backup validation..."
    
    # Check required environment variables
    if [[ -z "${DB_PASSWORD:-}" ]]; then
        log "ERROR: DB_PASSWORD environment variable not set"
        exit 1
    fi
    
    # Test database connection
    if ! PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" >/dev/null 2>&1; then
        log "ERROR: Cannot connect to database"
        exit 1
    fi
    
    # Check disk space (require at least 2GB free)
    AVAILABLE_SPACE=$(df "$BACKUP_DIR" | awk 'NR==2 {print $4}')
    if [[ $AVAILABLE_SPACE -lt 2097152 ]]; then  # 2GB in KB
        log "ERROR: Insufficient disk space for backup"
        exit 1
    fi
    
    log "Environment validation completed successfully"
}

# Create full database backup
create_full_backup() {
    log "Creating full database backup: $BACKUP_ID"
    
    local backup_file="${BACKUP_DIR}/full/${BACKUP_ID}/database.sql.gz"
    local metadata_file="${BACKUP_DIR}/full/${BACKUP_ID}/backup_metadata.json"
    
    # Create database dump
    PGPASSWORD="$DB_PASSWORD" pg_dump \
        -h "$DB_HOST" \
        -p "$DB_PORT" \
        -U "$DB_USER" \
        -d "$DB_NAME" \
        --verbose \
        --no-password \
        --format=plain \
        --compress=9 \
        --file="$backup_file"
    
    # Create metadata file
    cat > "$metadata_file" << EOF
{
    "backup_id": "$BACKUP_ID",
    "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
    "type": "full",
    "database": {
        "host": "$DB_HOST",
        "port": "$DB_PORT",
        "name": "$DB_NAME",
        "user": "$DB_USER"
    },
    "backup_file": "$backup_file",
    "file_size": "$(stat -c%s "$backup_file")",
    "checksum": "$(sha256sum "$backup_file" | cut -d' ' -f1)"
}
EOF
    
    log "Full backup completed: $backup_file"
}

# Verify backup integrity
verify_backup() {
    log "Verifying backup integrity..."
    
    local backup_file="${BACKUP_DIR}/full/${BACKUP_ID}/database.sql.gz"
    local metadata_file="${BACKUP_DIR}/full/${BACKUP_ID}/backup_metadata.json"
    
    # Check if files exist
    if [[ ! -f "$backup_file" ]] || [[ ! -f "$metadata_file" ]]; then
        log "ERROR: Backup files not found"
        return 1
    fi
    
    # Verify file size
    local expected_size=$(jq -r '.file_size' "$metadata_file")
    local actual_size=$(stat -c%s "$backup_file")
    
    if [[ "$expected_size" != "$actual_size" ]]; then
        log "ERROR: Backup file size mismatch"
        return 1
    fi
    
    # Verify checksum
    local expected_checksum=$(jq -r '.checksum' "$metadata_file")
    local actual_checksum=$(sha256sum "$backup_file" | cut -d' ' -f1)
    
    if [[ "$expected_checksum" != "$actual_checksum" ]]; then
        log "ERROR: Backup checksum mismatch"
        return 1
    fi
    
    # Test backup restoration (dry run)
    if ! gunzip -t "$backup_file"; then
        log "ERROR: Backup file is corrupted"
        return 1
    fi
    
    log "Backup verification completed successfully"
}

# Main backup process
main() {
    log "Starting database backup process"
    
    validate_environment
    create_full_backup
    
    if verify_backup; then
        log "Backup process completed successfully"
        exit 0
    else
        log "Backup verification failed"
        exit 1
    fi
}

# Run main function
main "$@"
```

### Manual Backup Commands

```bash
# Create immediate backup
./scripts/backup-database.sh

# Create backup with custom name
BACKUP_ID="manual_$(date +%Y%m%d_%H%M%S)" ./scripts/backup-database.sh

# Create backup to specific location
BACKUP_DIR="/custom/backup/location" ./scripts/backup-database.sh

# Create backup with compression
pg_dump -h localhost -U kyb_user -d kyb_platform | gzip > backup_$(date +%Y%m%d).sql.gz
```

### Incremental Backup Setup

```bash
# Enable WAL archiving in PostgreSQL
# Add to postgresql.conf:
wal_level = replica
archive_mode = on
archive_command = 'cp %p /backups/database/incremental/%f'

# Create incremental backup script
#!/bin/bash
# scripts/backup-incremental.sh

BACKUP_DIR="/backups/database/incremental"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# Create base backup
pg_basebackup -D "${BACKUP_DIR}/base_${TIMESTAMP}" -Ft -z -P

# Archive WAL files
pg_archivecleanup "${BACKUP_DIR}" 000000010000000000000001
```

## Application Backup Procedures

### Application State Backup

```bash
#!/bin/bash
# scripts/backup-application.sh

set -euo pipefail

BACKUP_DIR="/backups/application"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_ID="app_backup_${TIMESTAMP}"

# Create backup directory
mkdir -p "${BACKUP_DIR}/state/${BACKUP_ID}"

# Backup application configuration
cp -r /opt/kyb-platform/configs "${BACKUP_DIR}/state/${BACKUP_ID}/"

# Backup feature flags
curl -s "http://localhost:8080/api/v1/admin/feature-flags" > "${BACKUP_DIR}/state/${BACKUP_ID}/feature_flags.json"

# Backup user sessions (if applicable)
# This would depend on your session storage implementation

# Backup application logs
tar -czf "${BACKUP_DIR}/state/${BACKUP_ID}/logs.tar.gz" /var/log/kyb-platform/

# Create metadata
cat > "${BACKUP_DIR}/state/${BACKUP_ID}/metadata.json" << EOF
{
    "backup_id": "$BACKUP_ID",
    "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
    "type": "application_state",
    "version": "$(cat /opt/kyb-platform/VERSION)",
    "components": {
        "configs": true,
        "feature_flags": true,
        "logs": true
    }
}
EOF

echo "Application backup completed: $BACKUP_ID"
```

### Configuration Backup

```bash
#!/bin/bash
# scripts/backup-configuration.sh

set -euo pipefail

BACKUP_DIR="/backups/application/config"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_ID="config_backup_${TIMESTAMP}"

# Create backup directory
mkdir -p "${BACKUP_DIR}/${BACKUP_ID}"

# Backup environment files (without secrets)
find /opt/kyb-platform -name "*.env" -not -name "*secret*" -exec cp {} "${BACKUP_DIR}/${BACKUP_ID}/" \;

# Backup configuration files
cp -r /opt/kyb-platform/configs "${BACKUP_DIR}/${BACKUP_ID}/"

# Backup deployment configurations
cp -r /opt/kyb-platform/deployments "${BACKUP_DIR}/${BACKUP_ID}/"

# Create encrypted archive
tar -czf - "${BACKUP_DIR}/${BACKUP_ID}" | gpg --symmetric --cipher-algo AES256 --output "${BACKUP_DIR}/${BACKUP_ID}.tar.gz.gpg"

# Remove unencrypted files
rm -rf "${BACKUP_DIR}/${BACKUP_ID}"

echo "Configuration backup completed: ${BACKUP_ID}.tar.gz.gpg"
```

## Backup Verification and Testing

### Automated Verification Script

```bash
#!/bin/bash
# scripts/verify-backups.sh

set -euo pipefail

BACKUP_DIR="/backups"
LOG_FILE="/var/log/kyb-platform/backup-verification.log"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Verify database backups
verify_database_backups() {
    log "Verifying database backups..."
    
    local full_backups=("${BACKUP_DIR}/database/full"/*)
    local failed_backups=()
    
    for backup_dir in "${full_backups[@]}"; do
        if [[ -d "$backup_dir" ]]; then
            local backup_id=$(basename "$backup_dir")
            local metadata_file="${backup_dir}/backup_metadata.json"
            local backup_file="${backup_dir}/database.sql.gz"
            
            if [[ -f "$metadata_file" ]] && [[ -f "$backup_file" ]]; then
                # Verify checksum
                local expected_checksum=$(jq -r '.checksum' "$metadata_file")
                local actual_checksum=$(sha256sum "$backup_file" | cut -d' ' -f1)
                
                if [[ "$expected_checksum" == "$actual_checksum" ]]; then
                    log "✓ Backup $backup_id verified successfully"
                else
                    log "✗ Backup $backup_id checksum mismatch"
                    failed_backups+=("$backup_id")
                fi
            else
                log "✗ Backup $backup_id missing files"
                failed_backups+=("$backup_id")
            fi
        fi
    done
    
    if [[ ${#failed_backups[@]} -gt 0 ]]; then
        log "ERROR: ${#failed_backups[@]} backups failed verification"
        return 1
    else
        log "All database backups verified successfully"
        return 0
    fi
}

# Test backup restoration
test_backup_restoration() {
    log "Testing backup restoration..."
    
    # Find the most recent backup
    local latest_backup=$(ls -t "${BACKUP_DIR}/database/full" | head -n1)
    
    if [[ -z "$latest_backup" ]]; then
        log "ERROR: No backups found for testing"
        return 1
    fi
    
    local backup_file="${BACKUP_DIR}/database/full/${latest_backup}/database.sql.gz"
    
    # Create test database
    local test_db="kyb_backup_test_$(date +%s)"
    
    PGPASSWORD="$DB_PASSWORD" createdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$test_db"
    
    # Restore backup to test database
    if gunzip -c "$backup_file" | PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$test_db"; then
        log "✓ Backup restoration test successful"
        
        # Clean up test database
        PGPASSWORD="$DB_PASSWORD" dropdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$test_db"
        
        return 0
    else
        log "✗ Backup restoration test failed"
        
        # Clean up test database
        PGPASSWORD="$DB_PASSWORD" dropdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$test_db" 2>/dev/null || true
        
        return 1
    fi
}

# Main verification process
main() {
    log "Starting backup verification process"
    
    if verify_database_backups && test_backup_restoration; then
        log "All backup verifications passed"
        exit 0
    else
        log "Backup verification failed"
        exit 1
    fi
}

main "$@"
```

## Recovery Procedures

### Database Recovery

```bash
#!/bin/bash
# scripts/restore-database.sh

set -euo pipefail

BACKUP_ID="${1:-}"
BACKUP_DIR="/backups/database/full"

if [[ -z "$BACKUP_ID" ]]; then
    echo "Usage: $0 <backup_id>"
    echo "Available backups:"
    ls -la "$BACKUP_DIR"
    exit 1
fi

BACKUP_PATH="${BACKUP_DIR}/${BACKUP_ID}"
BACKUP_FILE="${BACKUP_PATH}/database.sql.gz"
METADATA_FILE="${BACKUP_PATH}/backup_metadata.json"

# Validate backup exists
if [[ ! -f "$BACKUP_FILE" ]] || [[ ! -f "$METADATA_FILE" ]]; then
    echo "ERROR: Backup $BACKUP_ID not found or incomplete"
    exit 1
fi

# Confirm restoration
echo "WARNING: This will restore the database from backup $BACKUP_ID"
echo "This operation will overwrite all current data."
read -p "Are you sure you want to continue? (yes/no): " confirm

if [[ "$confirm" != "yes" ]]; then
    echo "Restoration cancelled"
    exit 0
fi

# Create pre-restoration backup
echo "Creating pre-restoration backup..."
./scripts/backup-database.sh

# Stop application services
echo "Stopping application services..."
systemctl stop kyb-platform

# Restore database
echo "Restoring database from backup..."
gunzip -c "$BACKUP_FILE" | PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME"

# Verify restoration
echo "Verifying restoration..."
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT COUNT(*) FROM users;"

# Start application services
echo "Starting application services..."
systemctl start kyb-platform

echo "Database restoration completed successfully"
```

### Application Recovery

```bash
#!/bin/bash
# scripts/restore-application.sh

set -euo pipefail

BACKUP_ID="${1:-}"
BACKUP_DIR="/backups/application/state"

if [[ -z "$BACKUP_ID" ]]; then
    echo "Usage: $0 <backup_id>"
    echo "Available backups:"
    ls -la "$BACKUP_DIR"
    exit 1
fi

BACKUP_PATH="${BACKUP_DIR}/${BACKUP_ID}"

# Validate backup exists
if [[ ! -d "$BACKUP_PATH" ]]; then
    echo "ERROR: Application backup $BACKUP_ID not found"
    exit 1
fi

# Stop application
echo "Stopping application..."
systemctl stop kyb-platform

# Restore configurations
echo "Restoring application configurations..."
cp -r "${BACKUP_PATH}/configs" /opt/kyb-platform/

# Restore feature flags
echo "Restoring feature flags..."
if [[ -f "${BACKUP_PATH}/feature_flags.json" ]]; then
    curl -X POST "http://localhost:8080/api/v1/admin/feature-flags/restore" \
         -H "Content-Type: application/json" \
         -d @"${BACKUP_PATH}/feature_flags.json"
fi

# Restore logs (optional)
if [[ -f "${BACKUP_PATH}/logs.tar.gz" ]]; then
    echo "Restoring application logs..."
    tar -xzf "${BACKUP_PATH}/logs.tar.gz" -C /
fi

# Start application
echo "Starting application..."
systemctl start kyb-platform

echo "Application restoration completed successfully"
```

## Backup Monitoring and Alerting

### Backup Status Monitoring

```bash
#!/bin/bash
# scripts/monitor-backups.sh

BACKUP_DIR="/backups"
ALERT_EMAIL="admin@kyb-platform.com"

# Check backup age
check_backup_age() {
    local backup_type="$1"
    local max_age_hours="$2"
    local backup_path="${BACKUP_DIR}/${backup_type}"
    
    if [[ -d "$backup_path" ]]; then
        local latest_backup=$(ls -t "$backup_path" | head -n1)
        if [[ -n "$latest_backup" ]]; then
            local backup_time=$(stat -c %Y "${backup_path}/${latest_backup}")
            local current_time=$(date +%s)
            local age_hours=$(( (current_time - backup_time) / 3600 ))
            
            if [[ $age_hours -gt $max_age_hours ]]; then
                echo "ALERT: ${backup_type} backup is ${age_hours} hours old (max: ${max_age_hours})"
                return 1
            fi
        else
            echo "ALERT: No ${backup_type} backups found"
            return 1
        fi
    else
        echo "ALERT: ${backup_type} backup directory not found"
        return 1
    fi
    
    return 0
}

# Check backup integrity
check_backup_integrity() {
    local backup_type="$1"
    local backup_path="${BACKUP_DIR}/${backup_type}"
    
    if [[ -d "$backup_path" ]]; then
        local latest_backup=$(ls -t "$backup_path" | head -n1)
        if [[ -n "$latest_backup" ]]; then
            local metadata_file="${backup_path}/${latest_backup}/backup_metadata.json"
            if [[ -f "$metadata_file" ]]; then
                # Verify checksum
                local expected_checksum=$(jq -r '.checksum' "$metadata_file")
                local backup_file=$(jq -r '.backup_file' "$metadata_file")
                
                if [[ -f "$backup_file" ]]; then
                    local actual_checksum=$(sha256sum "$backup_file" | cut -d' ' -f1)
                    if [[ "$expected_checksum" != "$actual_checksum" ]]; then
                        echo "ALERT: ${backup_type} backup integrity check failed"
                        return 1
                    fi
                fi
            fi
        fi
    fi
    
    return 0
}

# Main monitoring
main() {
    local alerts=()
    
    # Check database backups (max 25 hours old)
    if ! check_backup_age "database/full" 25; then
        alerts+=("Database backup age check failed")
    fi
    
    if ! check_backup_integrity "database/full"; then
        alerts+=("Database backup integrity check failed")
    fi
    
    # Check application backups (max 7 days old)
    if ! check_backup_age "application/state" 168; then
        alerts+=("Application backup age check failed")
    fi
    
    # Send alerts if any issues found
    if [[ ${#alerts[@]} -gt 0 ]]; then
        local alert_message="Backup monitoring alerts:\n\n"
        for alert in "${alerts[@]}"; do
            alert_message+="- $alert\n"
        done
        
        echo -e "$alert_message" | mail -s "KYB Platform Backup Alert" "$ALERT_EMAIL"
        echo "Alerts sent: ${#alerts[@]} issues found"
        exit 1
    else
        echo "All backup checks passed"
        exit 0
    fi
}

main "$@"
```

## Maintenance and Cleanup

### Backup Cleanup Script

```bash
#!/bin/bash
# scripts/cleanup-backups.sh

set -euo pipefail

BACKUP_DIR="/backups"
RETENTION_DAYS="${1:-30}"

cleanup_old_backups() {
    local backup_path="$1"
    local retention_days="$2"
    
    if [[ -d "$backup_path" ]]; then
        echo "Cleaning up backups older than $retention_days days in $backup_path"
        
        # Find and remove old backups
        find "$backup_path" -type d -mtime +$retention_days -exec rm -rf {} \;
        
        echo "Cleanup completed for $backup_path"
    fi
}

# Cleanup different backup types
cleanup_old_backups "${BACKUP_DIR}/database/full" $RETENTION_DAYS
cleanup_old_backups "${BACKUP_DIR}/database/incremental" 7  # Keep incremental for 7 days
cleanup_old_backups "${BACKUP_DIR}/application/state" 10    # Keep app backups for 10 days
cleanup_old_backups "${BACKUP_DIR}/application/config" 90   # Keep config backups for 90 days

echo "Backup cleanup completed"
```

### Backup Health Report

```bash
#!/bin/bash
# scripts/backup-health-report.sh

BACKUP_DIR="/backups"
REPORT_FILE="/var/log/kyb-platform/backup-health-report.txt"

generate_report() {
    cat > "$REPORT_FILE" << EOF
KYB Platform Backup Health Report
Generated: $(date)
=====================================

Database Backups:
$(ls -la "${BACKUP_DIR}/database/full" | tail -n +2)

Application Backups:
$(ls -la "${BACKUP_DIR}/application/state" | tail -n +2)

Configuration Backups:
$(ls -la "${BACKUP_DIR}/application/config" | tail -n +2)

Disk Usage:
$(df -h "$BACKUP_DIR")

Backup Verification Status:
$(./scripts/verify-backups.sh 2>&1)

=====================================
EOF

    echo "Backup health report generated: $REPORT_FILE"
}

generate_report
```

## Best Practices

### 1. Backup Scheduling

- **Automated Backups**: Use cron jobs for regular automated backups
- **Off-Peak Hours**: Schedule backups during low-traffic periods
- **Staggered Backups**: Don't run all backups simultaneously

### 2. Backup Security

- **Encryption**: Encrypt sensitive backup data
- **Access Control**: Restrict backup file access permissions
- **Secure Storage**: Store backups in secure, off-site locations

### 3. Backup Testing

- **Regular Testing**: Test backup restoration monthly
- **Documentation**: Document all restoration procedures
- **Training**: Train staff on backup and recovery procedures

### 4. Monitoring

- **Automated Monitoring**: Set up automated backup monitoring
- **Alerting**: Configure alerts for backup failures
- **Reporting**: Generate regular backup health reports

This comprehensive backup documentation ensures the KYB Platform maintains data integrity and business continuity through robust backup and recovery procedures.
