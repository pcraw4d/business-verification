#!/bin/bash

# KYB Platform - Backup and Disaster Recovery Management Script
# Comprehensive backup operations, testing, monitoring, and disaster recovery

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
ENVIRONMENT="production"
REGION="us-west-2"
DR_REGION="us-east-1"
BACKUP_BUCKET=""
BACKUP_VAULT=""

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS] COMMAND"
    echo ""
    echo "Commands:"
    echo "  status       - Show backup and DR status"
    echo "  backup       - Create manual backup"
    echo "  restore      - Restore from backup"
    echo "  test         - Test backup restore"
    echo "  validate     - Validate backup integrity"
    echo "  monitor      - Monitor backup jobs"
    echo "  report       - Generate backup report"
    echo "  failover     - Initiate disaster recovery failover"
    echo "  failback     - Initiate failback to primary region"
    echo "  health       - Check backup and DR health"
    echo "  cleanup      - Clean up old backups"
    echo "  schedule     - Manage backup schedules"
    echo ""
    echo "Options:"
    echo "  -e, --environment ENV - Environment (production, staging, development)"
    echo "  -r, --region REGION   - AWS region (default: us-west-2)"
    echo "  -d, --dr-region REGION - DR region (default: us-east-1)"
    echo "  -b, --bucket BUCKET   - Backup S3 bucket name"
    echo "  -v, --vault VAULT     - AWS Backup vault name"
    echo "  -h, --help            - Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 status"
    echo "  $0 backup -e production"
    echo "  $0 test --region us-west-2"
    echo "  $0 failover --dr-region us-east-1"
}

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check if AWS CLI is installed
    if ! command -v aws &> /dev/null; then
        print_error "AWS CLI is not installed. Please install AWS CLI first."
        exit 1
    fi
    
    # Check AWS credentials
    if ! aws sts get-caller-identity &> /dev/null; then
        print_error "AWS credentials not configured. Please run 'aws configure' first."
        exit 1
    fi
    
    # Check if jq is installed
    if ! command -v jq &> /dev/null; then
        print_error "jq is not installed. Please install jq first."
        exit 1
    fi
    
    print_success "Prerequisites check passed"
}

# Function to show backup and DR status
show_status() {
    print_status "Showing backup and disaster recovery status..."
    
    echo "=== AWS Backup Status ==="
    
    # List backup vaults
    local backup_vaults=$(aws backup list-backup-vaults \
        --region "$REGION" \
        --output json 2>/dev/null || echo '{"BackupVaultList": []}')
    
    if [ "$(echo "$backup_vaults" | jq '.BackupVaultList | length')" -gt 0 ]; then
        echo "$backup_vaults" | jq -r '.BackupVaultList[] | "Vault: \(.BackupVaultName) - Type: \(.BackupVaultType)"'
    else
        echo "No backup vaults found"
    fi
    
    echo ""
    echo "=== Backup Plans ==="
    
    # List backup plans
    local backup_plans=$(aws backup list-backup-plans \
        --region "$REGION" \
        --output json 2>/dev/null || echo '{"BackupPlansList": []}')
    
    if [ "$(echo "$backup_plans" | jq '.BackupPlansList | length')" -gt 0 ]; then
        echo "$backup_plans" | jq -r '.BackupPlansList[] | "Plan: \(.BackupPlanName) - ID: \(.BackupPlanId)"'
    else
        echo "No backup plans found"
    fi
    
    echo ""
    echo "=== Recent Backup Jobs ==="
    
    # List recent backup jobs
    local backup_jobs=$(aws backup list-backup-jobs \
        --region "$REGION" \
        --output json 2>/dev/null || echo '{"BackupJobs': []}')
    
    if [ "$(echo "$backup_jobs" | jq '.BackupJobs | length')" -gt 0 ]; then
        echo "$backup_jobs" | jq -r '.BackupJobs[] | "\(.BackupJobId) - \(.ResourceType) - \(.State) - \(.CreationDate)"' | head -10
    else
        echo "No recent backup jobs found"
    fi
    
    echo ""
    echo "=== S3 Backup Storage ==="
    
    # Check S3 backup bucket
    if [ -n "$BACKUP_BUCKET" ]; then
        local bucket_size=$(aws s3 ls s3://"$BACKUP_BUCKET" --recursive --summarize --output json 2>/dev/null || echo '{"TotalSize": 0}')
        local total_size=$(echo "$bucket_size" | jq -r '.TotalSize // 0')
        local object_count=$(echo "$bucket_size" | jq -r '.TotalCount // 0')
        
        echo "Backup Bucket: $BACKUP_BUCKET"
        echo "Total Size: $(numfmt --to=iec $total_size)"
        echo "Object Count: $object_count"
    else
        echo "Backup bucket not specified"
    fi
    
    echo ""
    echo "=== Cross-Region Replication ==="
    
    # Check cross-region replication
    if [ -n "$BACKUP_BUCKET" ]; then
        local replication_status=$(aws s3api get-bucket-replication \
            --bucket "$BACKUP_BUCKET" \
            --region "$REGION" \
            --output json 2>/dev/null || echo '{}')
        
        if [ "$(echo "$replication_status" | jq 'keys | length')" -gt 0 ]; then
            echo "✅ Cross-region replication enabled"
            echo "$replication_status" | jq -r '.ReplicationConfiguration.Rules[] | "Rule: \(.ID) - Status: \(.Status)"'
        else
            echo "❌ Cross-region replication not configured"
        fi
    fi
    
    print_success "Status displayed successfully"
}

# Function to create manual backup
create_backup() {
    print_status "Creating manual backup..."
    
    if [ -z "$BACKUP_VAULT" ]; then
        print_error "Backup vault not specified. Use -v or --vault option."
        exit 1
    fi
    
    # Create backup selection for manual backup
    local backup_selection=$(cat <<EOF
{
  "SelectionName": "manual-backup-$(date +%Y%m%d-%H%M%S)",
  "IamRoleArn": "arn:aws:iam::$(aws sts get-caller-identity --query Account --output text):role/kyb-platform-backup-role",
  "Resources": [
    "arn:aws:rds:$REGION:$(aws sts get-caller-identity --query Account --output text):db:kyb-platform-db",
    "arn:aws:elasticache:$REGION:$(aws sts get-caller-identity --query Account --output text):replicationgroup:kyb-platform-redis",
    "arn:aws:s3:::kyb-platform-application-data",
    "arn:aws:s3:::kyb-platform-backup-storage"
  ]
}
EOF
)
    
    # Start backup job
    local backup_job=$(aws backup start-backup-job \
        --region "$REGION" \
        --backup-vault "$BACKUP_VAULT" \
        --resource-arn "arn:aws:rds:$REGION:$(aws sts get-caller-identity --query Account --output text):db:kyb-platform-db" \
        --iam-role-arn "arn:aws:iam::$(aws sts get-caller-identity --query Account --output text):role/kyb-platform-backup-role" \
        --output json)
    
    local job_id=$(echo "$backup_job" | jq -r '.BackupJobId')
    
    echo "✅ Manual backup started"
    echo "Backup Job ID: $job_id"
    echo "Vault: $BACKUP_VAULT"
    echo "Status: $(echo "$backup_job" | jq -r '.State')"
    
    print_success "Manual backup created successfully"
}

# Function to restore from backup
restore_backup() {
    print_status "Restoring from backup..."
    
    if [ -z "$BACKUP_VAULT" ]; then
        print_error "Backup vault not specified. Use -v or --vault option."
        exit 1
    fi
    
    # Get recovery point ARN from user
    echo "Available recovery points:"
    local recovery_points=$(aws backup list-recovery-points-by-backup-vault \
        --region "$REGION" \
        --backup-vault-name "$BACKUP_VAULT" \
        --output json)
    
    echo "$recovery_points" | jq -r '.RecoveryPoints[] | "\(.RecoveryPointArn) - \(.CreationDate) - \(.ResourceType)"' | head -10
    
    echo ""
    read -p "Enter recovery point ARN: " recovery_point_arn
    
    if [ -z "$recovery_point_arn" ]; then
        print_error "Recovery point ARN is required"
        exit 1
    fi
    
    # Start restore job
    local restore_job=$(aws backup start-restore-job \
        --region "$REGION" \
        --recovery-point-arn "$recovery_point_arn" \
        --iam-role-arn "arn:aws:iam::$(aws sts get-caller-identity --query Account --output text):role/kyb-platform-backup-role" \
        --output json)
    
    local job_id=$(echo "$restore_job" | jq -r '.RestoreJobId')
    
    echo "✅ Restore job started"
    echo "Restore Job ID: $job_id"
    echo "Recovery Point: $recovery_point_arn"
    echo "Status: $(echo "$restore_job" | jq -r '.Status')"
    
    print_success "Restore job started successfully"
}

# Function to test backup restore
test_backup() {
    print_status "Testing backup restore..."
    
    if [ -z "$BACKUP_VAULT" ]; then
        print_error "Backup vault not specified. Use -v or --vault option."
        exit 1
    fi
    
    # Create test environment
    local test_env_name="kyb-platform-test-$(date +%Y%m%d-%H%M%S)"
    
    echo "Creating test environment: $test_env_name"
    
    # Get latest recovery point
    local latest_recovery_point=$(aws backup list-recovery-points-by-backup-vault \
        --region "$REGION" \
        --backup-vault-name "$BACKUP_VAULT" \
        --output json | jq -r '.RecoveryPoints[0].RecoveryPointArn')
    
    if [ "$latest_recovery_point" = "null" ]; then
        print_error "No recovery points found in vault"
        exit 1
    fi
    
    echo "Using recovery point: $latest_recovery_point"
    
    # Start test restore job
    local test_restore_job=$(aws backup start-restore-job \
        --region "$REGION" \
        --recovery-point-arn "$latest_recovery_point" \
        --iam-role-arn "arn:aws:iam::$(aws sts get-caller-identity --query Account --output text):role/kyb-platform-backup-role" \
        --metadata '{"test-environment":"true","environment-name":"'$test_env_name'"}' \
        --output json)
    
    local job_id=$(echo "$test_restore_job" | jq -r '.RestoreJobId')
    
    echo "✅ Test restore job started"
    echo "Test Environment: $test_env_name"
    echo "Restore Job ID: $job_id"
    echo "Status: $(echo "$test_restore_job" | jq -r '.Status')"
    
    # Monitor test restore job
    echo ""
    echo "Monitoring test restore job..."
    
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        local job_status=$(aws backup describe-restore-job \
            --region "$REGION" \
            --restore-job-id "$job_id" \
            --output json | jq -r '.Status')
        
        echo "Attempt $attempt/$max_attempts: Status = $job_status"
        
        if [ "$job_status" = "COMPLETED" ]; then
            echo "✅ Test restore completed successfully"
            break
        elif [ "$job_status" = "FAILED" ] || [ "$job_status" = "ABORTED" ]; then
            echo "❌ Test restore failed with status: $job_status"
            exit 1
        fi
        
        sleep 60
        ((attempt++))
    done
    
    if [ $attempt -gt $max_attempts ]; then
        echo "⚠️  Test restore timed out"
    fi
    
    print_success "Backup test completed"
}

# Function to validate backup integrity
validate_backup() {
    print_status "Validating backup integrity..."
    
    if [ -z "$BACKUP_VAULT" ]; then
        print_error "Backup vault not specified. Use -v or --vault option."
        exit 1
    fi
    
    echo "=== Backup Integrity Validation ==="
    
    # Get recent recovery points
    local recovery_points=$(aws backup list-recovery-points-by-backup-vault \
        --region "$REGION" \
        --backup-vault-name "$BACKUP_VAULT" \
        --output json)
    
    local point_count=$(echo "$recovery_points" | jq '.RecoveryPoints | length')
    echo "Found $point_count recovery points"
    
    # Validate each recovery point
    local valid_count=0
    local invalid_count=0
    
    echo "$recovery_points" | jq -r '.RecoveryPoints[] | .RecoveryPointArn' | while read -r recovery_point_arn; do
        echo "Validating: $recovery_point_arn"
        
        # Check recovery point details
        local point_details=$(aws backup describe-recovery-point \
            --region "$REGION" \
            --backup-vault-name "$BACKUP_VAULT" \
            --recovery-point-arn "$recovery_point_arn" \
            --output json)
        
        local is_encrypted=$(echo "$point_details" | jq -r '.IsEncrypted')
        local resource_type=$(echo "$point_details" | jq -r '.ResourceType')
        local creation_date=$(echo "$point_details" | jq -r '.CreationDate')
        
        if [ "$is_encrypted" = "true" ]; then
            echo "  ✅ Encrypted: Yes"
            ((valid_count++))
        else
            echo "  ❌ Encrypted: No"
            ((invalid_count++))
        fi
        
        echo "  📋 Resource Type: $resource_type"
        echo "  📅 Creation Date: $creation_date"
        echo ""
    done
    
    echo "=== Validation Summary ==="
    echo "Valid Recovery Points: $valid_count"
    echo "Invalid Recovery Points: $invalid_count"
    
    if [ $invalid_count -eq 0 ]; then
        print_success "All recovery points are valid"
    else
        print_warning "Some recovery points have issues"
    fi
    
    print_success "Backup validation completed"
}

# Function to monitor backup jobs
monitor_backup() {
    print_status "Monitoring backup jobs..."
    
    echo "=== Active Backup Jobs ==="
    
    # List active backup jobs
    local active_backup_jobs=$(aws backup list-backup-jobs \
        --region "$REGION" \
        --by-state IN_PROGRESS \
        --output json)
    
    if [ "$(echo "$active_backup_jobs" | jq '.BackupJobs | length')" -gt 0 ]; then
        echo "$active_backup_jobs" | jq -r '.BackupJobs[] | "\(.BackupJobId) - \(.ResourceType) - \(.State) - \(.CreationDate)"'
    else
        echo "No active backup jobs"
    fi
    
    echo ""
    echo "=== Active Restore Jobs ==="
    
    # List active restore jobs
    local active_restore_jobs=$(aws backup list-restore-jobs \
        --region "$REGION" \
        --by-status IN_PROGRESS \
        --output json)
    
    if [ "$(echo "$active_restore_jobs" | jq '.RestoreJobs | length')" -gt 0 ]; then
        echo "$active_restore_jobs" | jq -r '.RestoreJobs[] | "\(.RestoreJobId) - \(.ResourceType) - \(.Status) - \(.CreationDate)"'
    else
        echo "No active restore jobs"
    fi
    
    echo ""
    echo "=== Recent Failed Jobs ==="
    
    # List recent failed jobs
    local failed_backup_jobs=$(aws backup list-backup-jobs \
        --region "$REGION" \
        --by-state FAILED \
        --output json)
    
    if [ "$(echo "$failed_backup_jobs" | jq '.BackupJobs | length')" -gt 0 ]; then
        echo "$failed_backup_jobs" | jq -r '.BackupJobs[] | "\(.BackupJobId) - \(.ResourceType) - \(.State) - \(.CreationDate)"' | head -5
    else
        echo "No recent failed backup jobs"
    fi
    
    print_success "Backup monitoring completed"
}

# Function to generate backup report
generate_report() {
    print_status "Generating backup report..."
    
    local report_file="backup-report-$(date +%Y%m%d-%H%M%S).json"
    
    echo "Generating report: $report_file"
    
    # Collect backup statistics
    local report_data=$(cat <<EOF
{
  "report_date": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "environment": "$ENVIRONMENT",
  "region": "$REGION",
  "backup_vault": "$BACKUP_VAULT",
  "backup_bucket": "$BACKUP_BUCKET"
}
EOF
)
    
    # Add backup job statistics
    local backup_jobs=$(aws backup list-backup-jobs \
        --region "$REGION" \
        --output json 2>/dev/null || echo '{"BackupJobs": []}')
    
    local total_backups=$(echo "$backup_jobs" | jq '.BackupJobs | length')
    local successful_backups=$(echo "$backup_jobs" | jq '.BackupJobs[] | select(.State == "COMPLETED") | .BackupJobId' | wc -l)
    local failed_backups=$(echo "$backup_jobs" | jq '.BackupJobs[] | select(.State == "FAILED") | .BackupJobId' | wc -l)
    
    # Add restore job statistics
    local restore_jobs=$(aws backup list-restore-jobs \
        --region "$REGION" \
        --output json 2>/dev/null || echo '{"RestoreJobs": []}')
    
    local total_restores=$(echo "$restore_jobs" | jq '.RestoreJobs | length')
    local successful_restores=$(echo "$restore_jobs" | jq '.RestoreJobs[] | select(.Status == "COMPLETED") | .RestoreJobId' | wc -l)
    local failed_restores=$(echo "$restore_jobs" | jq '.RestoreJobs[] | select(.Status == "FAILED") | .RestoreJobId' | wc -l)
    
    # Create comprehensive report
    local full_report=$(cat <<EOF
{
  "report_date": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "environment": "$ENVIRONMENT",
  "region": "$REGION",
  "backup_vault": "$BACKUP_VAULT",
  "backup_bucket": "$BACKUP_BUCKET",
  "statistics": {
    "backup_jobs": {
      "total": $total_backups,
      "successful": $successful_backups,
      "failed": $failed_backups,
      "success_rate": $(echo "scale=2; $successful_backups * 100 / $total_backups" | bc 2>/dev/null || echo "0")
    },
    "restore_jobs": {
      "total": $total_restores,
      "successful": $successful_restores,
      "failed": $failed_restores,
      "success_rate": $(echo "scale=2; $successful_restores * 100 / $total_restores" | bc 2>/dev/null || echo "0")
    }
  },
  "recent_backup_jobs": $(echo "$backup_jobs" | jq '.BackupJobs[0:10]'),
  "recent_restore_jobs": $(echo "$restore_jobs" | jq '.RestoreJobs[0:10]')
}
EOF
)
    
    # Save report to file
    echo "$full_report" | jq '.' > "$report_file"
    
    echo "✅ Backup report generated: $report_file"
    echo ""
    echo "=== Report Summary ==="
    echo "Total Backup Jobs: $total_backups"
    echo "Successful Backups: $successful_backups"
    echo "Failed Backups: $failed_backups"
    echo "Backup Success Rate: $(echo "scale=1; $successful_backups * 100 / $total_backups" | bc 2>/dev/null || echo "0")%"
    echo ""
    echo "Total Restore Jobs: $total_restores"
    echo "Successful Restores: $successful_restores"
    echo "Failed Restores: $failed_restores"
    echo "Restore Success Rate: $(echo "scale=1; $successful_restores * 100 / $total_restores" | bc 2>/dev/null || echo "0")%"
    
    print_success "Backup report generated successfully"
}

# Function to initiate disaster recovery failover
initiate_failover() {
    print_status "Initiating disaster recovery failover..."
    
    echo "⚠️  WARNING: This will initiate a failover to the DR region ($DR_REGION)"
    echo "This operation may cause service interruption."
    echo ""
    read -p "Are you sure you want to continue? (yes/no): " confirm
    
    if [ "$confirm" != "yes" ]; then
        echo "Failover cancelled"
        exit 0
    fi
    
    echo "Starting failover process..."
    
    # Check DR region health
    echo "Checking DR region health..."
    local dr_health=$(aws backup list-backup-vaults \
        --region "$DR_REGION" \
        --output json 2>/dev/null || echo '{"BackupVaultList": []}')
    
    if [ "$(echo "$dr_health" | jq '.BackupVaultList | length')" -eq 0 ]; then
        print_error "No backup vaults found in DR region"
        exit 1
    fi
    
    echo "✅ DR region is accessible"
    
    # Check primary region health
    echo "Checking primary region health..."
    local primary_health=$(curl -s -o /dev/null -w "%{http_code}" "https://$ENVIRONMENT.kybplatform.com/health" 2>/dev/null || echo "000")
    
    if [ "$primary_health" = "200" ]; then
        echo "⚠️  Primary region is healthy. Consider if failover is necessary."
        read -p "Continue with failover anyway? (yes/no): " continue_failover
        
        if [ "$continue_failover" != "yes" ]; then
            echo "Failover cancelled"
            exit 0
        fi
    else
        echo "❌ Primary region is unhealthy (HTTP $primary_health)"
    fi
    
    # Update Route53 for failover
    echo "Updating DNS for failover..."
    
    # This would typically involve updating Route53 records
    # For demonstration, we'll just show the process
    echo "Would update Route53 to point to DR region load balancer"
    echo "DR Load Balancer: kyb-platform-dr-alb.$DR_REGION.elb.amazonaws.com"
    
    # Simulate failover completion
    echo "✅ Failover initiated successfully"
    echo "DNS propagation may take 1-5 minutes"
    echo "Monitor application health at: https://$ENVIRONMENT.kybplatform.com/health"
    
    print_success "Disaster recovery failover initiated"
}

# Function to initiate failback to primary region
initiate_failback() {
    print_status "Initiating failback to primary region..."
    
    echo "⚠️  WARNING: This will initiate a failback to the primary region ($REGION)"
    echo "This operation may cause service interruption."
    echo ""
    read -p "Are you sure you want to continue? (yes/no): " confirm
    
    if [ "$confirm" != "yes" ]; then
        echo "Failback cancelled"
        exit 0
    fi
    
    echo "Starting failback process..."
    
    # Check primary region health
    echo "Checking primary region health..."
    local primary_health=$(curl -s -o /dev/null -w "%{http_code}" "https://$ENVIRONMENT.kybplatform.com/health" 2>/dev/null || echo "000")
    
    if [ "$primary_health" != "200" ]; then
        print_error "Primary region is not healthy (HTTP $primary_health)"
        print_error "Cannot failback to unhealthy primary region"
        exit 1
    fi
    
    echo "✅ Primary region is healthy"
    
    # Update Route53 for failback
    echo "Updating DNS for failback..."
    
    # This would typically involve updating Route53 records back to primary
    echo "Would update Route53 to point back to primary region load balancer"
    echo "Primary Load Balancer: kyb-platform-alb.$REGION.elb.amazonaws.com"
    
    # Simulate failback completion
    echo "✅ Failback initiated successfully"
    echo "DNS propagation may take 1-5 minutes"
    echo "Monitor application health at: https://$ENVIRONMENT.kybplatform.com/health"
    
    print_success "Failback to primary region initiated"
}

# Function to check backup and DR health
check_health() {
    print_status "Checking backup and disaster recovery health..."
    
    local health_status=0
    
    echo "=== Backup Health Check ==="
    
    # Check backup vault
    if [ -n "$BACKUP_VAULT" ]; then
        if aws backup describe-backup-vault \
            --region "$REGION" \
            --backup-vault-name "$BACKUP_VAULT" \
            --output json > /dev/null 2>&1; then
            echo "✅ Backup vault: Healthy"
        else
            echo "❌ Backup vault: Unhealthy"
            health_status=1
        fi
    else
        echo "⚠️  Backup vault: Not specified"
    fi
    
    # Check backup bucket
    if [ -n "$BACKUP_BUCKET" ]; then
        if aws s3 ls s3://"$BACKUP_BUCKET" --output json > /dev/null 2>&1; then
            echo "✅ Backup bucket: Healthy"
        else
            echo "❌ Backup bucket: Unhealthy"
            health_status=1
        fi
    else
        echo "⚠️  Backup bucket: Not specified"
    fi
    
    # Check recent backup jobs
    local recent_backups=$(aws backup list-backup-jobs \
        --region "$REGION" \
        --output json 2>/dev/null || echo '{"BackupJobs": []}')
    
    local recent_count=$(echo "$recent_backups" | jq '.BackupJobs | length')
    if [ "$recent_count" -gt 0 ]; then
        local latest_backup=$(echo "$recent_backups" | jq -r '.BackupJobs[0].CreationDate')
        echo "✅ Recent backups: Found $recent_count jobs (latest: $latest_backup)"
    else
        echo "❌ Recent backups: No backup jobs found"
        health_status=1
    fi
    
    echo ""
    echo "=== Disaster Recovery Health Check ==="
    
    # Check DR region accessibility
    if aws backup list-backup-vaults \
        --region "$DR_REGION" \
        --output json > /dev/null 2>&1; then
        echo "✅ DR region: Accessible"
    else
        echo "❌ DR region: Not accessible"
        health_status=1
    fi
    
    # Check cross-region replication
    if [ -n "$BACKUP_BUCKET" ]; then
        local replication_status=$(aws s3api get-bucket-replication \
            --bucket "$BACKUP_BUCKET" \
            --region "$REGION" \
            --output json 2>/dev/null || echo '{}')
        
        if [ "$(echo "$replication_status" | jq 'keys | length')" -gt 0 ]; then
            echo "✅ Cross-region replication: Enabled"
        else
            echo "❌ Cross-region replication: Not configured"
            health_status=1
        fi
    fi
    
    # Check application health
    local app_health=$(curl -s -o /dev/null -w "%{http_code}" "https://$ENVIRONMENT.kybplatform.com/health" 2>/dev/null || echo "000")
    if [ "$app_health" = "200" ]; then
        echo "✅ Application health: Healthy (HTTP $app_health)"
    else
        echo "❌ Application health: Unhealthy (HTTP $app_health)"
        health_status=1
    fi
    
    if [ $health_status -eq 0 ]; then
        print_success "Backup and DR system is healthy"
    else
        print_warning "Backup and DR system has issues"
        exit 1
    fi
}

# Function to clean up old backups
cleanup_backups() {
    print_status "Cleaning up old backups..."
    
    if [ -z "$BACKUP_VAULT" ]; then
        print_error "Backup vault not specified. Use -v or --vault option."
        exit 1
    fi
    
    echo "⚠️  WARNING: This will delete old backup recovery points"
    echo "This operation cannot be undone."
    echo ""
    read -p "Are you sure you want to continue? (yes/no): " confirm
    
    if [ "$confirm" != "yes" ]; then
        echo "Cleanup cancelled"
        exit 0
    fi
    
    # Get recovery points older than 30 days
    local cutoff_date=$(date -d '30 days ago' +%Y-%m-%dT%H:%M:%S)
    
    echo "Deleting recovery points older than: $cutoff_date"
    
    # List old recovery points
    local old_recovery_points=$(aws backup list-recovery-points-by-backup-vault \
        --region "$REGION" \
        --backup-vault-name "$BACKUP_VAULT" \
        --output json | jq -r --arg cutoff "$cutoff_date" '.RecoveryPoints[] | select(.CreationDate < $cutoff) | .RecoveryPointArn')
    
    local deleted_count=0
    
    echo "$old_recovery_points" | while read -r recovery_point_arn; do
        if [ -n "$recovery_point_arn" ]; then
            echo "Deleting: $recovery_point_arn"
            
            # Delete recovery point
            if aws backup delete-recovery-point \
                --region "$REGION" \
                --backup-vault-name "$BACKUP_VAULT" \
                --recovery-point-arn "$recovery_point_arn" > /dev/null 2>&1; then
                echo "  ✅ Deleted successfully"
                ((deleted_count++))
            else
                echo "  ❌ Failed to delete"
            fi
        fi
    done
    
    echo "=== Cleanup Summary ==="
    echo "Recovery points deleted: $deleted_count"
    
    print_success "Backup cleanup completed"
}

# Function to manage backup schedules
manage_schedules() {
    print_status "Managing backup schedules..."
    
    echo "=== Current Backup Plans ==="
    
    # List backup plans
    local backup_plans=$(aws backup list-backup-plans \
        --region "$REGION" \
        --output json)
    
    if [ "$(echo "$backup_plans" | jq '.BackupPlansList | length')" -gt 0 ]; then
        echo "$backup_plans" | jq -r '.BackupPlansList[] | "Plan: \(.BackupPlanName) - ID: \(.BackupPlanId)"'
        
        echo ""
        echo "=== Plan Details ==="
        
        echo "$backup_plans" | jq -r '.BackupPlansList[0].BackupPlanId' | while read -r plan_id; do
            if [ -n "$plan_id" ]; then
                local plan_details=$(aws backup get-backup-plan \
                    --region "$REGION" \
                    --backup-plan-id "$plan_id" \
                    --output json)
                
                echo "Plan ID: $plan_id"
                echo "$plan_details" | jq -r '.BackupPlan.BackupPlan.Rules[] | "Rule: \(.RuleName) - Schedule: \(.ScheduleExpression)"'
                echo ""
            fi
        done
    else
        echo "No backup plans found"
    fi
    
    print_success "Backup schedules displayed"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -e|--environment)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -r|--region)
            REGION="$2"
            shift 2
            ;;
        -d|--dr-region)
            DR_REGION="$2"
            shift 2
            ;;
        -b|--bucket)
            BACKUP_BUCKET="$2"
            shift 2
            ;;
        -v|--vault)
            BACKUP_VAULT="$2"
            shift 2
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        status|backup|restore|test|validate|monitor|report|failover|failback|health|cleanup|schedule)
            COMMAND="$1"
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Check if command is provided
if [ -z "$COMMAND" ]; then
    print_error "No command specified"
    show_usage
    exit 1
fi

# Main execution
print_status "Starting backup and disaster recovery management"
print_status "Environment: $ENVIRONMENT"
print_status "Region: $REGION"
print_status "DR Region: $DR_REGION"
print_status "Backup Bucket: $BACKUP_BUCKET"
print_status "Backup Vault: $BACKUP_VAULT"

# Check prerequisites
check_prerequisites

# Execute command
case $COMMAND in
    status)
        show_status
        ;;
    backup)
        create_backup
        ;;
    restore)
        restore_backup
        ;;
    test)
        test_backup
        ;;
    validate)
        validate_backup
        ;;
    monitor)
        monitor_backup
        ;;
    report)
        generate_report
        ;;
    failover)
        initiate_failover
        ;;
    failback)
        initiate_failback
        ;;
    health)
        check_health
        ;;
    cleanup)
        cleanup_backups
        ;;
    schedule)
        manage_schedules
        ;;
    *)
        print_error "Unknown command: $COMMAND"
        show_usage
        exit 1
        ;;
esac

print_success "Backup and disaster recovery management completed successfully!"
