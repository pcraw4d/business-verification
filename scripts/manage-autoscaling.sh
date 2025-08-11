#!/bin/bash

# KYB Platform - Auto Scaling Management Script
# Handles auto-scaling policies, monitoring, and operational tasks

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
CLUSTER_NAME="kyb-platform-cluster"
NAMESPACE="default"

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
    echo "  status       - Show auto-scaling status"
    echo "  policies     - List scaling policies"
    echo "  metrics      - Show scaling metrics"
    echo "  events       - Show scaling events"
    echo "  scale        - Manually scale resources"
    echo "  suspend      - Suspend auto-scaling"
    echo "  resume       - Resume auto-scaling"
    echo "  test         - Test scaling policies"
    echo "  dashboard    - Open CloudWatch dashboard"
    echo "  alarms       - Show scaling alarms"
    echo "  logs         - Show autoscaler logs"
    echo ""
    echo "Options:"
    echo "  -e, --environment ENV - Environment (production, staging, development)"
    echo "  -r, --region REGION   - AWS region (default: us-west-2)"
    echo "  -c, --cluster NAME    - EKS cluster name"
    echo "  -n, --namespace NS    - Kubernetes namespace (default: default)"
    echo "  -h, --help            - Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 status"
    echo "  $0 scale api 5"
    echo "  $0 suspend -e production"
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
    
    # Check if kubectl is installed
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl is not installed. Please install kubectl first."
        exit 1
    fi
    
    # Check if jq is installed
    if ! command -v jq &> /dev/null; then
        print_error "jq is not installed. Please install jq first."
        exit 1
    fi
    
    print_success "Prerequisites check passed"
}

# Function to configure kubectl
configure_kubectl() {
    print_status "Configuring kubectl for EKS cluster..."
    
    aws eks update-kubeconfig \
        --region "$REGION" \
        --name "$CLUSTER_NAME"
    
    print_success "kubectl configured for cluster: $CLUSTER_NAME"
}

# Function to show auto-scaling status
show_status() {
    print_status "Showing auto-scaling status..."
    
    echo "=== EKS Cluster Auto Scaling Status ==="
    
    # Get EKS node groups
    local node_groups=$(aws eks list-nodegroups \
        --region "$REGION" \
        --cluster-name "$CLUSTER_NAME" \
        --output json)
    
    echo "$node_groups" | jq -r '.nodegroups[]' | while read -r nodegroup; do
        echo "Node Group: $nodegroup"
        
        # Get node group scaling configuration
        local scaling_config=$(aws eks describe-nodegroup \
            --region "$REGION" \
            --cluster-name "$CLUSTER_NAME" \
            --nodegroup-name "$nodegroup" \
            --output json)
        
        local desired=$(echo "$scaling_config" | jq -r '.nodegroup.scalingConfig.desiredSize')
        local min=$(echo "$scaling_config" | jq -r '.nodegroup.scalingConfig.minSize')
        local max=$(echo "$scaling_config" | jq -r '.nodegroup.scalingConfig.maxSize')
        
        echo "  Desired: $desired, Min: $min, Max: $max"
    done
    
    echo ""
    echo "=== Kubernetes HPA Status ==="
    
    # Get Horizontal Pod Autoscalers
    kubectl get hpa -n "$NAMESPACE" -o wide
    
    echo ""
    echo "=== Kubernetes VPA Status ==="
    
    # Get Vertical Pod Autoscalers
    kubectl get vpa -n "$NAMESPACE" -o wide
    
    echo ""
    echo "=== Application Auto Scaling Status ==="
    
    # Get RDS auto scaling
    local rds_scaling=$(aws application-autoscaling describe-scalable-targets \
        --region "$REGION" \
        --service-namespace rds \
        --output json 2>/dev/null || echo '{"ScalableTargets": []}')
    
    if [ "$(echo "$rds_scaling" | jq '.ScalableTargets | length')" -gt 0 ]; then
        echo "RDS Auto Scaling:"
        echo "$rds_scaling" | jq -r '.ScalableTargets[] | "  \(.ResourceId) - Min: \(.MinCapacity), Max: \(.MaxCapacity)"'
    else
        echo "RDS Auto Scaling: Not configured"
    fi
    
    # Get Redis auto scaling
    local redis_scaling=$(aws application-autoscaling describe-scalable-targets \
        --region "$REGION" \
        --service-namespace elasticache \
        --output json 2>/dev/null || echo '{"ScalableTargets": []}')
    
    if [ "$(echo "$redis_scaling" | jq '.ScalableTargets | length')" -gt 0 ]; then
        echo "Redis Auto Scaling:"
        echo "$redis_scaling" | jq -r '.ScalableTargets[] | "  \(.ResourceId) - Min: \(.MinCapacity), Max: \(.MaxCapacity)"'
    else
        echo "Redis Auto Scaling: Not configured"
    fi
    
    print_success "Status displayed successfully"
}

# Function to list scaling policies
list_policies() {
    print_status "Listing scaling policies..."
    
    echo "=== Kubernetes HPA Policies ==="
    kubectl describe hpa -n "$NAMESPACE"
    
    echo ""
    echo "=== Kubernetes VPA Policies ==="
    kubectl describe vpa -n "$NAMESPACE"
    
    echo ""
    echo "=== Application Auto Scaling Policies ==="
    
    # RDS policies
    local rds_policies=$(aws application-autoscaling describe-scaling-policies \
        --region "$REGION" \
        --service-namespace rds \
        --output json 2>/dev/null || echo '{"ScalingPolicies": []}')
    
    if [ "$(echo "$rds_policies" | jq '.ScalingPolicies | length')" -gt 0 ]; then
        echo "RDS Scaling Policies:"
        echo "$rds_policies" | jq -r '.ScalingPolicies[] | "  \(.PolicyName) - Type: \(.PolicyType)"'
    fi
    
    # Redis policies
    local redis_policies=$(aws application-autoscaling describe-scaling-policies \
        --region "$REGION" \
        --service-namespace elasticache \
        --output json 2>/dev/null || echo '{"ScalingPolicies": []}')
    
    if [ "$(echo "$redis_policies" | jq '.ScalingPolicies | length')" -gt 0 ]; then
        echo "Redis Scaling Policies:"
        echo "$redis_policies" | jq -r '.ScalingPolicies[] | "  \(.PolicyName) - Type: \(.PolicyType)"'
    fi
    
    print_success "Policies listed successfully"
}

# Function to show scaling metrics
show_metrics() {
    print_status "Showing scaling metrics..."
    
    # Get metrics for the last hour
    local end_time=$(date -u +%Y-%m-%dT%H:%M:%S)
    local start_time=$(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%S)
    
    echo "=== Cluster Metrics (Last Hour) ==="
    
    # EKS cluster metrics
    local cluster_cpu=$(aws cloudwatch get-metric-statistics \
        --region "$REGION" \
        --namespace "AWS/ECS" \
        --metric-name "CPUUtilization" \
        --dimensions "Name=ClusterName,Value=$CLUSTER_NAME" \
        --start-time "$start_time" \
        --end-time "$end_time" \
        --period 300 \
        --statistics Average \
        --output json)
    
    local cluster_memory=$(aws cloudwatch get-metric-statistics \
        --region "$REGION" \
        --namespace "AWS/ECS" \
        --metric-name "MemoryUtilization" \
        --dimensions "Name=ClusterName,Value=$CLUSTER_NAME" \
        --start-time "$start_time" \
        --end-time "$end_time" \
        --period 300 \
        --statistics Average \
        --output json)
    
    echo "Cluster CPU: $(echo "$cluster_cpu" | jq -r '.Datapoints[0].Average // 0')%"
    echo "Cluster Memory: $(echo "$cluster_memory" | jq -r '.Datapoints[0].Average // 0')%"
    
    echo ""
    echo "=== Pod Metrics ==="
    
    # Get pod metrics
    kubectl top pods -n "$NAMESPACE"
    
    echo ""
    echo "=== Node Metrics ==="
    
    # Get node metrics
    kubectl top nodes
    
    print_success "Metrics displayed successfully"
}

# Function to show scaling events
show_events() {
    print_status "Showing scaling events..."
    
    echo "=== Kubernetes Scaling Events ==="
    
    # Get HPA events
    kubectl get events -n "$NAMESPACE" --field-selector involvedObject.kind=HorizontalPodAutoscaler --sort-by='.lastTimestamp'
    
    echo ""
    echo "=== EKS Scaling Events ==="
    
    # Get EKS node group events
    local node_groups=$(aws eks list-nodegroups \
        --region "$REGION" \
        --cluster-name "$CLUSTER_NAME" \
        --output json)
    
    echo "$node_groups" | jq -r '.nodegroups[]' | while read -r nodegroup; do
        echo "Node Group: $nodegroup"
        
        # Get recent activities
        aws autoscaling describe-scaling-activities \
            --region "$REGION" \
            --auto-scaling-group-name "$nodegroup" \
            --max-items 5 \
            --output table
    done
    
    print_success "Events displayed successfully"
}

# Function to manually scale resources
scale_resource() {
    local resource_type=$1
    local target_value=$2
    
    if [ -z "$resource_type" ] || [ -z "$target_value" ]; then
        print_error "Usage: $0 scale <resource> <value>"
        print_error "Examples:"
        print_error "  $0 scale api 5"
        print_error "  $0 scale web 3"
        exit 1
    fi
    
    print_status "Scaling $resource_type to $target_value replicas..."
    
    case $resource_type in
        api)
            kubectl scale deployment kyb-platform-api -n "$NAMESPACE" --replicas="$target_value"
            ;;
        web)
            kubectl scale deployment kyb-platform-web -n "$NAMESPACE" --replicas="$target_value"
            ;;
        *)
            print_error "Unknown resource type: $resource_type"
            print_error "Supported types: api, web"
            exit 1
            ;;
    esac
    
    print_success "$resource_type scaled to $target_value replicas"
}

# Function to suspend auto-scaling
suspend_autoscaling() {
    print_status "Suspending auto-scaling..."
    
    # Suspend HPA
    kubectl patch hpa kyb-platform-api-hpa -n "$NAMESPACE" -p '{"spec":{"minReplicas":0,"maxReplicas":0}}'
    
    # Suspend EKS node group scaling
    local node_groups=$(aws eks list-nodegroups \
        --region "$REGION" \
        --cluster-name "$CLUSTER_NAME" \
        --output json)
    
    echo "$node_groups" | jq -r '.nodegroups[]' | while read -r nodegroup; do
        aws autoscaling suspend-processes \
            --region "$REGION" \
            --auto-scaling-group-name "$nodegroup" \
            --scaling-processes ReplaceUnhealthy
    done
    
    print_success "Auto-scaling suspended"
}

# Function to resume auto-scaling
resume_autoscaling() {
    print_status "Resuming auto-scaling..."
    
    # Resume HPA
    kubectl patch hpa kyb-platform-api-hpa -n "$NAMESPACE" -p '{"spec":{"minReplicas":3,"maxReplicas":30}}'
    
    # Resume EKS node group scaling
    local node_groups=$(aws eks list-nodegroups \
        --region "$REGION" \
        --cluster-name "$CLUSTER_NAME" \
        --output json)
    
    echo "$node_groups" | jq -r '.nodegroups[]' | while read -r nodegroup; do
        aws autoscaling resume-processes \
            --region "$REGION" \
            --auto-scaling-group-name "$nodegroup" \
            --scaling-processes ReplaceUnhealthy
    done
    
    print_success "Auto-scaling resumed"
}

# Function to test scaling policies
test_scaling() {
    print_status "Testing scaling policies..."
    
    echo "=== Current State ==="
    show_status
    
    echo ""
    echo "=== Generating Load Test ==="
    
    # Get load balancer DNS
    local alb_dns=$(aws elbv2 describe-load-balancers \
        --region "$REGION" \
        --names "kyb-platform-alb" \
        --query 'LoadBalancers[0].DNSName' \
        --output text 2>/dev/null)
    
    if [ -n "$alb_dns" ] && [ "$alb_dns" != "None" ]; then
        echo "Load testing endpoint: https://$alb_dns/health"
        
        # Simple load test
        for i in {1..10}; do
            echo "Request $i/10"
            curl -s "https://$alb_dns/health" > /dev/null &
        done
        wait
        
        echo "Load test completed"
    else
        print_warning "Load balancer not found, skipping load test"
    fi
    
    echo ""
    echo "=== Monitoring Scaling ==="
    
    # Monitor for 2 minutes
    for i in {1..12}; do
        echo "Check $i/12"
        kubectl get hpa -n "$NAMESPACE" -o wide
        sleep 10
    done
    
    print_success "Scaling test completed"
}

# Function to open CloudWatch dashboard
open_dashboard() {
    print_status "Opening CloudWatch dashboard..."
    
    local dashboard_url="https://${REGION}.console.aws.amazon.com/cloudwatch/home?region=${REGION}#dashboards:name=kyb-platform-autoscaling"
    
    echo "Dashboard URL: $dashboard_url"
    
    # Try to open in browser
    if command -v open &> /dev/null; then
        open "$dashboard_url"
    elif command -v xdg-open &> /dev/null; then
        xdg-open "$dashboard_url"
    else
        print_warning "Could not open browser automatically. Please visit: $dashboard_url"
    fi
    
    print_success "Dashboard opened"
}

# Function to show scaling alarms
show_alarms() {
    print_status "Showing scaling alarms..."
    
    # Get auto-scaling related alarms
    local alarms=$(aws cloudwatch describe-alarms \
        --region "$REGION" \
        --alarm-name-prefix "kyb-platform" \
        --output json)
    
    if [ "$(echo "$alarms" | jq '.MetricAlarms | length')" -gt 0 ]; then
        echo "=== Auto Scaling Alarms ==="
        echo "$alarms" | jq -r '.MetricAlarms[] | "\(.AlarmName) - \(.StateValue) (\(.MetricName))"'
    else
        print_warning "No auto-scaling alarms found"
    fi
    
    print_success "Alarms displayed successfully"
}

# Function to show autoscaler logs
show_logs() {
    print_status "Showing autoscaler logs..."
    
    echo "=== Cluster Autoscaler Logs ==="
    kubectl logs -n kube-system deployment/cluster-autoscaler --tail=50
    
    echo ""
    echo "=== HPA Controller Logs ==="
    kubectl logs -n kube-system deployment/horizontal-pod-autoscaler --tail=50 2>/dev/null || echo "HPA controller not found"
    
    echo ""
    echo "=== VPA Admission Controller Logs ==="
    kubectl logs -n kube-system deployment/vpa-admission-controller --tail=50 2>/dev/null || echo "VPA admission controller not found"
    
    print_success "Logs displayed successfully"
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
        -c|--cluster)
            CLUSTER_NAME="$2"
            shift 2
            ;;
        -n|--namespace)
            NAMESPACE="$2"
            shift 2
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        status|policies|metrics|events|scale|suspend|resume|test|dashboard|alarms|logs)
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
print_status "Starting auto-scaling management"
print_status "Environment: $ENVIRONMENT"
print_status "Region: $REGION"
print_status "Cluster: $CLUSTER_NAME"
print_status "Namespace: $NAMESPACE"

# Check prerequisites
check_prerequisites

# Configure kubectl
configure_kubectl

# Execute command
case $COMMAND in
    status)
        show_status
        ;;
    policies)
        list_policies
        ;;
    metrics)
        show_metrics
        ;;
    events)
        show_events
        ;;
    scale)
        scale_resource "${ARGS[0]}" "${ARGS[1]}"
        ;;
    suspend)
        suspend_autoscaling
        ;;
    resume)
        resume_autoscaling
        ;;
    test)
        test_scaling
        ;;
    dashboard)
        open_dashboard
        ;;
    alarms)
        show_alarms
        ;;
    logs)
        show_logs
        ;;
    *)
        print_error "Unknown command: $COMMAND"
        show_usage
        exit 1
        ;;
esac

print_success "Auto-scaling management completed successfully!"
