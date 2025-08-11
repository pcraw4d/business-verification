#!/bin/bash

# KYB Platform - Load Balancer Management Script
# Handles load balancer monitoring, health checks, and operational tasks

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
ALB_NAME="kyb-platform-alb"
API_TG_NAME="kyb-platform-api-tg"
WEB_TG_NAME="kyb-platform-web-tg"

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
    echo "  status       - Show load balancer status"
    echo "  health       - Check target health"
    echo "  metrics      - Show load balancer metrics"
    echo "  logs         - Show access logs"
    echo "  targets      - List target groups and targets"
    echo "  waf-status   - Show WAF status and rules"
    echo "  alarms       - Show CloudWatch alarms"
    echo "  test-health  - Test health check endpoints"
    echo "  drain        - Drain targets from load balancer"
    echo "  register     - Register targets with load balancer"
    echo "  deregister   - Deregister targets from load balancer"
    echo ""
    echo "Options:"
    echo "  -e, --environment ENV - Environment (production, staging, development)"
    echo "  -r, --region REGION   - AWS region (default: us-west-2)"
    echo "  -a, --alb-name NAME   - Load balancer name"
    echo "  -h, --help            - Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 status"
    echo "  $0 health -e production"
    echo "  $0 metrics --region us-east-1"
    echo "  $0 test-health"
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

# Function to get load balancer ARN
get_alb_arn() {
    local alb_arn=$(aws elbv2 describe-load-balancers \
        --region "$REGION" \
        --names "$ALB_NAME" \
        --query 'LoadBalancers[0].LoadBalancerArn' \
        --output text 2>/dev/null)
    
    if [ "$alb_arn" = "None" ] || [ -z "$alb_arn" ]; then
        print_error "Load balancer '$ALB_NAME' not found in region '$REGION'"
        exit 1
    fi
    
    echo "$alb_arn"
}

# Function to get target group ARN
get_target_group_arn() {
    local tg_name=$1
    local tg_arn=$(aws elbv2 describe-target-groups \
        --region "$REGION" \
        --names "$tg_name" \
        --query 'TargetGroups[0].TargetGroupArn' \
        --output text 2>/dev/null)
    
    if [ "$tg_arn" = "None" ] || [ -z "$tg_arn" ]; then
        print_error "Target group '$tg_name' not found in region '$REGION'"
        exit 1
    fi
    
    echo "$tg_arn"
}

# Function to show load balancer status
show_status() {
    print_status "Showing load balancer status..."
    
    local alb_arn=$(get_alb_arn)
    
    # Get load balancer details
    local alb_info=$(aws elbv2 describe-load-balancers \
        --region "$REGION" \
        --load-balancer-arns "$alb_arn" \
        --output json)
    
    echo "=== Load Balancer Status ==="
    echo "Name: $(echo "$alb_info" | jq -r '.LoadBalancers[0].LoadBalancerName')"
    echo "DNS Name: $(echo "$alb_info" | jq -r '.LoadBalancers[0].DNSName')"
    echo "State: $(echo "$alb_info" | jq -r '.LoadBalancers[0].State.Code')"
    echo "Type: $(echo "$alb_info" | jq -r '.LoadBalancers[0].Type')"
    echo "Scheme: $(echo "$alb_info" | jq -r '.LoadBalancers[0].Scheme')"
    echo "VPC: $(echo "$alb_info" | jq -r '.LoadBalancers[0].VpcId')"
    
    # Get listeners
    local listeners=$(aws elbv2 describe-listeners \
        --region "$REGION" \
        --load-balancer-arn "$alb_arn" \
        --output json)
    
    echo ""
    echo "=== Listeners ==="
    echo "$listeners" | jq -r '.Listeners[] | "Port \(.Port) (\(.Protocol)) -> \(.DefaultActions[0].Type)"'
    
    print_success "Status displayed successfully"
}

# Function to check target health
check_health() {
    print_status "Checking target health..."
    
    local api_tg_arn=$(get_target_group_arn "$API_TG_NAME")
    local web_tg_arn=$(get_target_group_arn "$WEB_TG_NAME")
    
    echo "=== API Target Group Health ==="
    aws elbv2 describe-target-health \
        --region "$REGION" \
        --target-group-arn "$api_tg_arn" \
        --query 'TargetHealthDescriptions[].[Target.Id,TargetHealth.State,TargetHealth.Description]' \
        --output table
    
    echo ""
    echo "=== Web Target Group Health ==="
    aws elbv2 describe-target-health \
        --region "$REGION" \
        --target-group-arn "$web_tg_arn" \
        --query 'TargetHealthDescriptions[].[Target.Id,TargetHealth.State,TargetHealth.Description]' \
        --output table
    
    print_success "Health check completed"
}

# Function to show metrics
show_metrics() {
    print_status "Showing load balancer metrics..."
    
    local alb_arn=$(get_alb_arn)
    local alb_suffix=$(echo "$alb_arn" | cut -d'/' -f2)
    
    # Get metrics for the last hour
    local end_time=$(date -u +%Y-%m-%dT%H:%M:%S)
    local start_time=$(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%S)
    
    echo "=== Load Balancer Metrics (Last Hour) ==="
    
    # Request count
    local request_count=$(aws cloudwatch get-metric-statistics \
        --region "$REGION" \
        --namespace "AWS/ApplicationELB" \
        --metric-name "RequestCount" \
        --dimensions "Name=LoadBalancer,Value=$alb_suffix" \
        --start-time "$start_time" \
        --end-time "$end_time" \
        --period 300 \
        --statistics Sum \
        --output json)
    
    echo "Total Requests: $(echo "$request_count" | jq -r '.Datapoints[0].Sum // 0')"
    
    # Target response time
    local response_time=$(aws cloudwatch get-metric-statistics \
        --region "$REGION" \
        --namespace "AWS/ApplicationELB" \
        --metric-name "TargetResponseTime" \
        --dimensions "Name=LoadBalancer,Value=$alb_suffix" \
        --start-time "$start_time" \
        --end-time "$end_time" \
        --period 300 \
        --statistics Average \
        --output json)
    
    echo "Average Response Time: $(echo "$response_time" | jq -r '.Datapoints[0].Average // 0') seconds"
    
    # Error rates
    local error_4xx=$(aws cloudwatch get-metric-statistics \
        --region "$REGION" \
        --namespace "AWS/ApplicationELB" \
        --metric-name "HTTPCode_ELB_4XX_Count" \
        --dimensions "Name=LoadBalancer,Value=$alb_suffix" \
        --start-time "$start_time" \
        --end-time "$end_time" \
        --period 300 \
        --statistics Sum \
        --output json)
    
    local error_5xx=$(aws cloudwatch get-metric-statistics \
        --region "$REGION" \
        --namespace "AWS/ApplicationELB" \
        --metric-name "HTTPCode_ELB_5XX_Count" \
        --dimensions "Name=LoadBalancer,Value=$alb_suffix" \
        --start-time "$start_time" \
        --end-time "$end_time" \
        --period 300 \
        --statistics Sum \
        --output json)
    
    echo "4XX Errors: $(echo "$error_4xx" | jq -r '.Datapoints[0].Sum // 0')"
    echo "5XX Errors: $(echo "$error_5xx" | jq -r '.Datapoints[0].Sum // 0')"
    
    print_success "Metrics displayed successfully"
}

# Function to show access logs
show_logs() {
    print_status "Showing recent access logs..."
    
    # Get S3 bucket for ALB logs
    local log_bucket=$(aws elbv2 describe-load-balancers \
        --region "$REGION" \
        --names "$ALB_NAME" \
        --query 'LoadBalancers[0].LoadBalancerAttributes[?Key==`access_logs.s3.bucket`].Value' \
        --output text 2>/dev/null)
    
    if [ -z "$log_bucket" ] || [ "$log_bucket" = "None" ]; then
        print_warning "Access logs not configured for load balancer"
        return
    fi
    
    # List recent log files
    local recent_logs=$(aws s3 ls "s3://$log_bucket/alb-logs/" --recursive | tail -10)
    
    if [ -z "$recent_logs" ]; then
        print_warning "No recent log files found"
        return
    fi
    
    echo "=== Recent Access Logs ==="
    echo "$recent_logs"
    
    print_success "Logs displayed successfully"
}

# Function to list targets
list_targets() {
    print_status "Listing target groups and targets..."
    
    local api_tg_arn=$(get_target_group_arn "$API_TG_NAME")
    local web_tg_arn=$(get_target_group_arn "$WEB_TG_NAME")
    
    echo "=== Target Groups ==="
    aws elbv2 describe-target-groups \
        --region "$REGION" \
        --target-group-arns "$api_tg_arn" "$web_tg_arn" \
        --query 'TargetGroups[].[TargetGroupName,Port,Protocol,TargetType]' \
        --output table
    
    echo ""
    echo "=== API Targets ==="
    aws elbv2 describe-target-health \
        --region "$REGION" \
        --target-group-arn "$api_tg_arn" \
        --query 'TargetHealthDescriptions[].[Target.Id,Target.Port,TargetHealth.State]' \
        --output table
    
    echo ""
    echo "=== Web Targets ==="
    aws elbv2 describe-target-health \
        --region "$REGION" \
        --target-group-arn "$web_tg_arn" \
        --query 'TargetHealthDescriptions[].[Target.Id,Target.Port,TargetHealth.State]' \
        --output table
    
    print_success "Targets listed successfully"
}

# Function to show WAF status
show_waf_status() {
    print_status "Showing WAF status and rules..."
    
    # Get WAF Web ACL
    local waf_acl=$(aws wafv2 list-web-acls \
        --region "$REGION" \
        --scope REGIONAL \
        --query 'WebACLs[?contains(Name, `kyb-platform`)].{Name:Name,ARN:ARN}' \
        --output json)
    
    if [ "$(echo "$waf_acl" | jq '. | length')" -eq 0 ]; then
        print_warning "No WAF Web ACL found for KYB Platform"
        return
    fi
    
    local acl_arn=$(echo "$waf_acl" | jq -r '.[0].ARN')
    local acl_name=$(echo "$waf_acl" | jq -r '.[0].Name')
    
    echo "=== WAF Web ACL ==="
    echo "Name: $acl_name"
    echo "ARN: $acl_arn"
    
    # Get WAF rules
    local waf_rules=$(aws wafv2 get-web-acl \
        --region "$REGION" \
        --name "$acl_name" \
        --scope REGIONAL \
        --output json)
    
    echo ""
    echo "=== WAF Rules ==="
    echo "$waf_rules" | jq -r '.WebACL.Rules[] | "\(.Priority) - \(.Name) (\(.Action.Block // .Action.Allow // .Action.Override))"'
    
    # Get WAF metrics
    echo ""
    echo "=== WAF Metrics (Last Hour) ==="
    local end_time=$(date -u +%Y-%m-%dT%H:%M:%S)
    local start_time=$(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%S)
    
    local waf_requests=$(aws cloudwatch get-metric-statistics \
        --region "$REGION" \
        --namespace "AWS/WAFV2" \
        --metric-name "AllowedRequests" \
        --dimensions "Name=WebACL,Value=$acl_name" "Name=Region,Value=$REGION" \
        --start-time "$start_time" \
        --end-time "$end_time" \
        --period 300 \
        --statistics Sum \
        --output json)
    
    local waf_blocked=$(aws cloudwatch get-metric-statistics \
        --region "$REGION" \
        --namespace "AWS/WAFV2" \
        --metric-name "BlockedRequests" \
        --dimensions "Name=WebACL,Value=$acl_name" "Name=Region,Value=$REGION" \
        --start-time "$start_time" \
        --end-time "$end_time" \
        --period 300 \
        --statistics Sum \
        --output json)
    
    echo "Allowed Requests: $(echo "$waf_requests" | jq -r '.Datapoints[0].Sum // 0')"
    echo "Blocked Requests: $(echo "$waf_blocked" | jq -r '.Datapoints[0].Sum // 0')"
    
    print_success "WAF status displayed successfully"
}

# Function to show CloudWatch alarms
show_alarms() {
    print_status "Showing CloudWatch alarms..."
    
    # Get ALB-related alarms
    local alarms=$(aws cloudwatch describe-alarms \
        --region "$REGION" \
        --alarm-name-prefix "kyb-platform-alb" \
        --output json)
    
    if [ "$(echo "$alarms" | jq '.MetricAlarms | length')" -eq 0 ]; then
        print_warning "No ALB alarms found"
        return
    fi
    
    echo "=== Load Balancer Alarms ==="
    echo "$alarms" | jq -r '.MetricAlarms[] | "\(.AlarmName) - \(.StateValue) (\(.MetricName))"'
    
    print_success "Alarms displayed successfully"
}

# Function to test health check endpoints
test_health() {
    print_status "Testing health check endpoints..."
    
    local alb_arn=$(get_alb_arn)
    local dns_name=$(aws elbv2 describe-load-balancers \
        --region "$REGION" \
        --load-balancer-arns "$alb_arn" \
        --query 'LoadBalancers[0].DNSName' \
        --output text)
    
    echo "=== Testing Health Endpoints ==="
    echo "Load Balancer DNS: $dns_name"
    echo ""
    
    # Test health endpoint
    echo "Testing /health endpoint..."
    local health_response=$(curl -s -w "\nHTTP Status: %{http_code}\nResponse Time: %{time_total}s\n" \
        "https://$dns_name/health" || echo "Failed to connect")
    echo "$health_response"
    echo ""
    
    # Test metrics endpoint
    echo "Testing /metrics endpoint..."
    local metrics_response=$(curl -s -w "\nHTTP Status: %{http_code}\nResponse Time: %{time_total}s\n" \
        "https://$dns_name/metrics" || echo "Failed to connect")
    echo "$metrics_response"
    echo ""
    
    # Test API endpoint
    echo "Testing /v1/status endpoint..."
    local api_response=$(curl -s -w "\nHTTP Status: %{http_code}\nResponse Time: %{time_total}s\n" \
        "https://$dns_name/v1/status" || echo "Failed to connect")
    echo "$api_response"
    
    print_success "Health tests completed"
}

# Function to drain targets
drain_targets() {
    print_status "Draining targets from load balancer..."
    
    local target_group=$1
    local target_ip=$2
    
    if [ -z "$target_group" ] || [ -z "$target_ip" ]; then
        print_error "Usage: $0 drain <target-group> <target-ip>"
        print_error "Example: $0 drain api-tg 10.0.1.100"
        exit 1
    fi
    
    local tg_arn=$(get_target_group_arn "$target_group")
    
    # Deregister target
    aws elbv2 deregister-targets \
        --region "$REGION" \
        --target-group-arn "$tg_arn" \
        --targets "Id=$target_ip"
    
    print_success "Target $target_ip drained from $target_group"
}

# Function to register targets
register_targets() {
    print_status "Registering targets with load balancer..."
    
    local target_group=$1
    local target_ip=$2
    local target_port=${3:-8080}
    
    if [ -z "$target_group" ] || [ -z "$target_ip" ]; then
        print_error "Usage: $0 register <target-group> <target-ip> [port]"
        print_error "Example: $0 register api-tg 10.0.1.100 8080"
        exit 1
    fi
    
    local tg_arn=$(get_target_group_arn "$target_group")
    
    # Register target
    aws elbv2 register-targets \
        --region "$REGION" \
        --target-group-arn "$tg_arn" \
        --targets "Id=$target_ip,Port=$target_port"
    
    print_success "Target $target_ip:$target_port registered with $target_group"
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
        -a|--alb-name)
            ALB_NAME="$2"
            shift 2
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        status|health|metrics|logs|targets|waf-status|alarms|test-health|drain|register|deregister)
            COMMAND="$1"
            shift
            ;;
        *)
            ARGS+=("$1")
            shift
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
print_status "Starting load balancer management"
print_status "Environment: $ENVIRONMENT"
print_status "Region: $REGION"
print_status "Load Balancer: $ALB_NAME"

# Check prerequisites
check_prerequisites

# Execute command
case $COMMAND in
    status)
        show_status
        ;;
    health)
        check_health
        ;;
    metrics)
        show_metrics
        ;;
    logs)
        show_logs
        ;;
    targets)
        list_targets
        ;;
    waf-status)
        show_waf_status
        ;;
    alarms)
        show_alarms
        ;;
    test-health)
        test_health
        ;;
    drain)
        drain_targets "${ARGS[0]}" "${ARGS[1]}"
        ;;
    register)
        register_targets "${ARGS[0]}" "${ARGS[1]}" "${ARGS[2]}"
        ;;
    deregister)
        drain_targets "${ARGS[0]}" "${ARGS[1]}"
        ;;
    *)
        print_error "Unknown command: $COMMAND"
        show_usage
        exit 1
        ;;
esac

print_success "Load balancer management completed successfully!"
