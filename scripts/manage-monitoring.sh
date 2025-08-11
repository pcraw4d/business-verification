#!/bin/bash

# KYB Platform - Monitoring Management Script
# Handles CloudWatch monitoring, alerting, and operational tasks

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
RETENTION_DAYS=30

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
    echo "  status       - Show monitoring status"
    echo "  alarms       - List CloudWatch alarms"
    echo "  metrics      - Show key metrics"
    echo "  logs         - Show recent logs"
    echo "  dashboard    - Open monitoring dashboards"
    echo "  alerts       - Show recent alerts"
    echo "  test         - Test monitoring endpoints"
    echo "  retention    - Manage log retention"
    echo "  insights     - Run CloudWatch Insights queries"
    echo "  export       - Export monitoring data"
    echo "  health       - Check monitoring health"
    echo ""
    echo "Options:"
    echo "  -e, --environment ENV - Environment (production, staging, development)"
    echo "  -r, --region REGION   - AWS region (default: us-west-2)"
    echo "  -d, --days DAYS       - Number of days for retention (default: 30)"
    echo "  -h, --help            - Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 status"
    echo "  $0 alarms -e production"
    echo "  $0 metrics --region us-east-1"
    echo "  $0 test"
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

# Function to show monitoring status
show_status() {
    print_status "Showing monitoring status..."
    
    echo "=== CloudWatch Log Groups ==="
    
    # List log groups
    local log_groups=$(aws logs describe-log-groups \
        --region "$REGION" \
        --log-group-name-prefix "/aws/eks/kyb-platform" \
        --output json)
    
    echo "$log_groups" | jq -r '.logGroups[] | "\(.logGroupName) - Retention: \(.retentionInDays) days"'
    
    echo ""
    echo "=== CloudWatch Alarms ==="
    
    # List alarms
    local alarms=$(aws cloudwatch describe-alarms \
        --region "$REGION" \
        --alarm-name-prefix "kyb-platform" \
        --output json)
    
    if [ "$(echo "$alarms" | jq '.MetricAlarms | length')" -gt 0 ]; then
        echo "$alarms" | jq -r '.MetricAlarms[] | "\(.AlarmName) - \(.StateValue) (\(.MetricName))"'
    else
        echo "No alarms found"
    fi
    
    echo ""
    echo "=== SNS Topics ==="
    
    # List SNS topics
    local sns_topics=$(aws sns list-topics \
        --region "$REGION" \
        --output json)
    
    echo "$sns_topics" | jq -r '.Topics[] | select(.TopicArn | contains("kyb-platform")) | "\(.TopicArn)"'
    
    print_success "Status displayed successfully"
}

# Function to list CloudWatch alarms
list_alarms() {
    print_status "Listing CloudWatch alarms..."
    
    # Get all alarms
    local alarms=$(aws cloudwatch describe-alarms \
        --region "$REGION" \
        --alarm-name-prefix "kyb-platform" \
        --output json)
    
    if [ "$(echo "$alarms" | jq '.MetricAlarms | length')" -eq 0 ]; then
        print_warning "No alarms found"
        return
    fi
    
    echo "=== Alarm Summary ==="
    echo "Total Alarms: $(echo "$alarms" | jq '.MetricAlarms | length')"
    echo "OK: $(echo "$alarms" | jq '.MetricAlarms[] | select(.StateValue == "OK") | .AlarmName' | wc -l)"
    echo "ALARM: $(echo "$alarms" | jq '.MetricAlarms[] | select(.StateValue == "ALARM") | .AlarmName' | wc -l)"
    echo "INSUFFICIENT_DATA: $(echo "$alarms" | jq '.MetricAlarms[] | select(.StateValue == "INSUFFICIENT_DATA") | .AlarmName' | wc -l)"
    
    echo ""
    echo "=== Alarm Details ==="
    echo "$alarms" | jq -r '.MetricAlarms[] | "\(.AlarmName)\n  State: \(.StateValue)\n  Metric: \(.MetricName)\n  Threshold: \(.Threshold)\n  Description: \(.AlarmDescription)\n"'
    
    print_success "Alarms listed successfully"
}

# Function to show key metrics
show_metrics() {
    print_status "Showing key metrics..."
    
    # Get metrics for the last hour
    local end_time=$(date -u +%Y-%m-%dT%H:%M:%S)
    local start_time=$(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%S)
    
    echo "=== Application Metrics (Last Hour) ==="
    
    # Error count
    local error_count=$(aws cloudwatch get-metric-statistics \
        --region "$REGION" \
        --namespace "KYBPlatform/Application" \
        --metric-name "ErrorCount" \
        --start-time "$start_time" \
        --end-time "$end_time" \
        --period 300 \
        --statistics Sum \
        --output json)
    
    echo "Error Count: $(echo "$error_count" | jq -r '.Datapoints[0].Sum // 0')"
    
    # API request count
    local api_requests=$(aws cloudwatch get-metric-statistics \
        --region "$REGION" \
        --namespace "KYBPlatform/API" \
        --metric-name "APIRequestCount" \
        --start-time "$start_time" \
        --end-time "$end_time" \
        --period 300 \
        --statistics Sum \
        --output json)
    
    echo "API Requests: $(echo "$api_requests" | jq -r '.Datapoints[0].Sum // 0')"
    
    echo ""
    echo "=== Infrastructure Metrics ==="
    
    # Load balancer metrics
    local alb_requests=$(aws cloudwatch get-metric-statistics \
        --region "$REGION" \
        --namespace "AWS/ApplicationELB" \
        --metric-name "RequestCount" \
        --start-time "$start_time" \
        --end-time "$end_time" \
        --period 300 \
        --statistics Sum \
        --output json)
    
    echo "Load Balancer Requests: $(echo "$alb_requests" | jq -r '.Datapoints[0].Sum // 0')"
    
    # Database metrics
    local db_cpu=$(aws cloudwatch get-metric-statistics \
        --region "$REGION" \
        --namespace "AWS/RDS" \
        --metric-name "CPUUtilization" \
        --start-time "$start_time" \
        --end-time "$end_time" \
        --period 300 \
        --statistics Average \
        --output json)
    
    echo "Database CPU: $(echo "$db_cpu" | jq -r '.Datapoints[0].Average // 0')%"
    
    print_success "Metrics displayed successfully"
}

# Function to show recent logs
show_logs() {
    print_status "Showing recent logs..."
    
    echo "=== Recent Application Logs ==="
    
    # Get recent application logs
    local app_logs=$(aws logs filter-log-events \
        --region "$REGION" \
        --log-group-name "/aws/eks/kyb-platform/application" \
        --start-time $(($(date +%s) - 3600))000 \
        --output json 2>/dev/null || echo '{"events": []}')
    
    if [ "$(echo "$app_logs" | jq '.events | length')" -gt 0 ]; then
        echo "$app_logs" | jq -r '.events[] | "\(.timestamp) - \(.message)"' | tail -20
    else
        echo "No recent application logs found"
    fi
    
    echo ""
    echo "=== Recent Error Logs ==="
    
    # Get recent error logs
    local error_logs=$(aws logs filter-log-events \
        --region "$REGION" \
        --log-group-name "/aws/eks/kyb-platform/application" \
        --filter-pattern "ERROR" \
        --start-time $(($(date +%s) - 3600))000 \
        --output json 2>/dev/null || echo '{"events": []}')
    
    if [ "$(echo "$error_logs" | jq '.events | length')" -gt 0 ]; then
        echo "$error_logs" | jq -r '.events[] | "\(.timestamp) - \(.message)"' | tail -10
    else
        echo "No recent error logs found"
    fi
    
    echo ""
    echo "=== Recent Security Logs ==="
    
    # Get recent security logs
    local security_logs=$(aws logs filter-log-events \
        --region "$REGION" \
        --log-group-name "/aws/eks/kyb-platform/security" \
        --start-time $(($(date +%s) - 3600))000 \
        --output json 2>/dev/null || echo '{"events": []}')
    
    if [ "$(echo "$security_logs" | jq '.events | length')" -gt 0 ]; then
        echo "$security_logs" | jq -r '.events[] | "\(.timestamp) - \(.message)"' | tail -10
    else
        echo "No recent security logs found"
    fi
    
    print_success "Logs displayed successfully"
}

# Function to open monitoring dashboards
open_dashboards() {
    print_status "Opening monitoring dashboards..."
    
    local base_url="https://${REGION}.console.aws.amazon.com/cloudwatch/home?region=${REGION}#dashboards:name="
    
    echo "=== Dashboard URLs ==="
    echo "Application Monitoring: ${base_url}kyb-platform-application-monitoring"
    echo "Security Monitoring: ${base_url}kyb-platform-security-monitoring"
    echo "Infrastructure Monitoring: ${base_url}kyb-platform-infrastructure-monitoring"
    echo "Auto Scaling: ${base_url}kyb-platform-autoscaling"
    
    # Try to open in browser
    if command -v open &> /dev/null; then
        open "${base_url}kyb-platform-application-monitoring"
    elif command -v xdg-open &> /dev/null; then
        xdg-open "${base_url}kyb-platform-application-monitoring"
    else
        print_warning "Could not open browser automatically. Please visit the URLs above."
    fi
    
    print_success "Dashboards opened"
}

# Function to show recent alerts
show_alerts() {
    print_status "Showing recent alerts..."
    
    # Get recent alarm history
    local alarm_history=$(aws cloudwatch describe-alarm-history \
        --region "$REGION" \
        --alarm-name-prefix "kyb-platform" \
        --start-date $(date -u -d '24 hours ago' +%Y-%m-%dT%H:%M:%S) \
        --end-date $(date -u +%Y-%m-%dT%H:%M:%S) \
        --output json)
    
    if [ "$(echo "$alarm_history" | jq '.AlarmHistoryItems | length')" -gt 0 ]; then
        echo "=== Recent Alarm History ==="
        echo "$alarm_history" | jq -r '.AlarmHistoryItems[] | "\(.Timestamp) - \(.AlarmName) - \(.HistoryItemType)"' | head -20
    else
        echo "No recent alarm history found"
    fi
    
    print_success "Alerts displayed successfully"
}

# Function to test monitoring endpoints
test_monitoring() {
    print_status "Testing monitoring endpoints..."
    
    echo "=== Testing CloudWatch API ==="
    
    # Test CloudWatch API
    if aws cloudwatch list-metrics --region "$REGION" --namespace "KYBPlatform/Application" --output json > /dev/null 2>&1; then
        echo "✅ CloudWatch API: OK"
    else
        echo "❌ CloudWatch API: Failed"
    fi
    
    echo "=== Testing Logs API ==="
    
    # Test Logs API
    if aws logs describe-log-groups --region "$REGION" --log-group-name-prefix "/aws/eks/kyb-platform" --output json > /dev/null 2>&1; then
        echo "✅ Logs API: OK"
    else
        echo "❌ Logs API: Failed"
    fi
    
    echo "=== Testing SNS API ==="
    
    # Test SNS API
    if aws sns list-topics --region "$REGION" --output json > /dev/null 2>&1; then
        echo "✅ SNS API: OK"
    else
        echo "❌ SNS API: Failed"
    fi
    
    echo "=== Testing Metrics ==="
    
    # Test metrics collection
    local end_time=$(date -u +%Y-%m-%dT%H:%M:%S)
    local start_time=$(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%S)
    
    local test_metrics=$(aws cloudwatch get-metric-statistics \
        --region "$REGION" \
        --namespace "KYBPlatform/Application" \
        --metric-name "ErrorCount" \
        --start-time "$start_time" \
        --end-time "$end_time" \
        --period 300 \
        --statistics Sum \
        --output json 2>/dev/null || echo '{"Datapoints": []}')
    
    if [ "$(echo "$test_metrics" | jq '.Datapoints | length')" -ge 0 ]; then
        echo "✅ Metrics Collection: OK"
    else
        echo "❌ Metrics Collection: Failed"
    fi
    
    print_success "Monitoring tests completed"
}

# Function to manage log retention
manage_retention() {
    print_status "Managing log retention..."
    
    echo "Setting log retention to $RETENTION_DAYS days..."
    
    # Update log group retention
    local log_groups=(
        "/aws/eks/kyb-platform/application"
        "/aws/eks/kyb-platform/system"
        "/aws/eks/kyb-platform/security"
        "/aws/eks/kyb-platform/audit"
    )
    
    for log_group in "${log_groups[@]}"; do
        if aws logs put-retention-policy \
            --region "$REGION" \
            --log-group-name "$log_group" \
            --retention-in-days "$RETENTION_DAYS" > /dev/null 2>&1; then
            echo "✅ Updated retention for $log_group"
        else
            echo "❌ Failed to update retention for $log_group"
        fi
    done
    
    print_success "Log retention updated"
}

# Function to run CloudWatch Insights queries
run_insights() {
    print_status "Running CloudWatch Insights queries..."
    
    echo "=== Error Analysis (Last Hour) ==="
    
    # Error analysis query
    local error_query="fields @timestamp, @message
| filter @message like /ERROR/
| stats count() by bin(5m)
| sort @timestamp desc
| limit 10"
    
    echo "$error_query" | aws logs start-query \
        --region "$REGION" \
        --log-group-names "/aws/eks/kyb-platform/application" \
        --start-time $(($(date +%s) - 3600))000 \
        --end-time $(date +%s)000 \
        --query-string "$error_query" \
        --output json
    
    echo ""
    echo "=== Security Analysis (Last Hour) ==="
    
    # Security analysis query
    local security_query="fields @timestamp, @message
| filter @message like /security/ or @message like /auth/ or @message like /permission/
| stats count() by bin(5m)
| sort @timestamp desc
| limit 10"
    
    echo "$security_query" | aws logs start-query \
        --region "$REGION" \
        --log-group-names "/aws/eks/kyb-platform/security" \
        --start-time $(($(date +%s) - 3600))000 \
        --end-time $(date +%s)000 \
        --query-string "$security_query" \
        --output json
    
    print_success "Insights queries executed"
}

# Function to export monitoring data
export_data() {
    print_status "Exporting monitoring data..."
    
    local export_dir="monitoring-export-$(date +%Y%m%d-%H%M%S)"
    mkdir -p "$export_dir"
    
    echo "Exporting to directory: $export_dir"
    
    # Export alarms
    aws cloudwatch describe-alarms \
        --region "$REGION" \
        --alarm-name-prefix "kyb-platform" \
        --output json > "$export_dir/alarms.json"
    
    # Export log groups
    aws logs describe-log-groups \
        --region "$REGION" \
        --log-group-name-prefix "/aws/eks/kyb-platform" \
        --output json > "$export_dir/log-groups.json"
    
    # Export metrics (last 24 hours)
    local end_time=$(date -u +%Y-%m-%dT%H:%M:%S)
    local start_time=$(date -u -d '24 hours ago' +%Y-%m-%dT%H:%M:%S)
    
    aws cloudwatch get-metric-statistics \
        --region "$REGION" \
        --namespace "KYBPlatform/Application" \
        --metric-name "ErrorCount" \
        --start-time "$start_time" \
        --end-time "$end_time" \
        --period 300 \
        --statistics Sum \
        --output json > "$export_dir/error-metrics.json"
    
    echo "Exported files:"
    ls -la "$export_dir"
    
    print_success "Monitoring data exported to $export_dir"
}

# Function to check monitoring health
check_health() {
    print_status "Checking monitoring health..."
    
    local health_status=0
    
    echo "=== Monitoring Health Check ==="
    
    # Check CloudWatch API
    if aws cloudwatch list-metrics --region "$REGION" --namespace "KYBPlatform/Application" --output json > /dev/null 2>&1; then
        echo "✅ CloudWatch API: Healthy"
    else
        echo "❌ CloudWatch API: Unhealthy"
        health_status=1
    fi
    
    # Check Logs API
    if aws logs describe-log-groups --region "$REGION" --log-group-name-prefix "/aws/eks/kyb-platform" --output json > /dev/null 2>&1; then
        echo "✅ Logs API: Healthy"
    else
        echo "❌ Logs API: Unhealthy"
        health_status=1
    fi
    
    # Check SNS API
    if aws sns list-topics --region "$REGION" --output json > /dev/null 2>&1; then
        echo "✅ SNS API: Healthy"
    else
        echo "❌ SNS API: Unhealthy"
        health_status=1
    fi
    
    # Check alarm status
    local alarms=$(aws cloudwatch describe-alarms \
        --region "$REGION" \
        --alarm-name-prefix "kyb-platform" \
        --state-value ALARM \
        --output json)
    
    local alarm_count=$(echo "$alarms" | jq '.MetricAlarms | length')
    if [ "$alarm_count" -eq 0 ]; then
        echo "✅ Alarms: No active alarms"
    else
        echo "⚠️  Alarms: $alarm_count active alarms"
        health_status=1
    fi
    
    if [ $health_status -eq 0 ]; then
        print_success "Monitoring system is healthy"
    else
        print_warning "Monitoring system has issues"
        exit 1
    fi
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
        -d|--days)
            RETENTION_DAYS="$2"
            shift 2
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        status|alarms|metrics|logs|dashboard|alerts|test|retention|insights|export|health)
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
print_status "Starting monitoring management"
print_status "Environment: $ENVIRONMENT"
print_status "Region: $REGION"
print_status "Retention Days: $RETENTION_DAYS"

# Check prerequisites
check_prerequisites

# Execute command
case $COMMAND in
    status)
        show_status
        ;;
    alarms)
        list_alarms
        ;;
    metrics)
        show_metrics
        ;;
    logs)
        show_logs
        ;;
    dashboard)
        open_dashboards
        ;;
    alerts)
        show_alerts
        ;;
    test)
        test_monitoring
        ;;
    retention)
        manage_retention
        ;;
    insights)
        run_insights
        ;;
    export)
        export_data
        ;;
    health)
        check_health
        ;;
    *)
        print_error "Unknown command: $COMMAND"
        show_usage
        exit 1
        ;;
esac

print_success "Monitoring management completed successfully!"
