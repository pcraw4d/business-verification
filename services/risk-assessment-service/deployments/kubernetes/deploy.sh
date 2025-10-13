#!/bin/bash

# Risk Assessment Service - Kubernetes Deployment Script
# This script deploys the Risk Assessment Service to Kubernetes

set -e

# Configuration
NAMESPACE="kyb-platform"
SERVICE_NAME="risk-assessment-service"
IMAGE_TAG="${IMAGE_TAG:-latest}"
ENVIRONMENT="${ENVIRONMENT:-production}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if kubectl is available
check_kubectl() {
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed or not in PATH"
        exit 1
    fi
    log_success "kubectl is available"
}

# Check if namespace exists, create if not
check_namespace() {
    if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
        log_info "Creating namespace: $NAMESPACE"
        kubectl create namespace "$NAMESPACE"
        log_success "Namespace $NAMESPACE created"
    else
        log_info "Namespace $NAMESPACE already exists"
    fi
}

# Apply Kubernetes manifests
apply_manifests() {
    local manifest_dir="$(dirname "$0")"
    
    log_info "Applying Kubernetes manifests..."
    
    # Apply in order
    kubectl apply -f "$manifest_dir/namespace.yaml" -n "$NAMESPACE" || true
    kubectl apply -f "$manifest_dir/rbac.yaml" -n "$NAMESPACE"
    kubectl apply -f "$manifest_dir/configmap.yaml" -n "$NAMESPACE"
    kubectl apply -f "$manifest_dir/secrets.yaml" -n "$NAMESPACE"
    kubectl apply -f "$manifest_dir/network-policies.yaml" -n "$NAMESPACE"
    kubectl apply -f "$manifest_dir/pdb.yaml" -n "$NAMESPACE"
    kubectl apply -f "$manifest_dir/deployment.yaml" -n "$NAMESPACE"
    kubectl apply -f "$manifest_dir/service.yaml" -n "$NAMESPACE"
    kubectl apply -f "$manifest_dir/hpa.yaml" -n "$NAMESPACE"
    kubectl apply -f "$manifest_dir/ingress.yaml" -n "$NAMESPACE"
    
    log_success "All manifests applied successfully"
}

# Wait for deployment to be ready
wait_for_deployment() {
    log_info "Waiting for deployment to be ready..."
    
    kubectl wait --for=condition=available --timeout=300s deployment/"$SERVICE_NAME" -n "$NAMESPACE"
    
    log_success "Deployment is ready"
}

# Check deployment status
check_deployment_status() {
    log_info "Checking deployment status..."
    
    kubectl get deployment "$SERVICE_NAME" -n "$NAMESPACE"
    kubectl get pods -l app="$SERVICE_NAME" -n "$NAMESPACE"
    kubectl get services -l app="$SERVICE_NAME" -n "$NAMESPACE"
    kubectl get hpa "$SERVICE_NAME-hpa" -n "$NAMESPACE"
}

# Get service endpoints
get_endpoints() {
    log_info "Getting service endpoints..."
    
    local service_ip=$(kubectl get service "$SERVICE_NAME" -n "$NAMESPACE" -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    local service_hostname=$(kubectl get service "$SERVICE_NAME" -n "$NAMESPACE" -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
    
    if [ -n "$service_ip" ]; then
        log_success "Service IP: $service_ip"
    elif [ -n "$service_hostname" ]; then
        log_success "Service Hostname: $service_hostname"
    else
        log_warning "LoadBalancer IP/Hostname not available yet"
    fi
    
    # Get ingress endpoints
    local ingress_host=$(kubectl get ingress "$SERVICE_NAME-ingress" -n "$NAMESPACE" -o jsonpath='{.spec.rules[0].host}')
    if [ -n "$ingress_host" ]; then
        log_success "Ingress Host: $ingress_host"
    fi
}

# Health check
health_check() {
    log_info "Performing health check..."
    
    local pod_name=$(kubectl get pods -l app="$SERVICE_NAME" -n "$NAMESPACE" -o jsonpath='{.items[0].metadata.name}')
    
    if [ -n "$pod_name" ]; then
        kubectl exec "$pod_name" -n "$NAMESPACE" -- curl -f http://localhost:8080/health || {
            log_error "Health check failed"
            return 1
        }
        log_success "Health check passed"
    else
        log_error "No pods found for health check"
        return 1
    fi
}

# Cleanup function
cleanup() {
    log_info "Cleaning up resources..."
    
    kubectl delete -f "$(dirname "$0")/ingress.yaml" -n "$NAMESPACE" || true
    kubectl delete -f "$(dirname "$0")/hpa.yaml" -n "$NAMESPACE" || true
    kubectl delete -f "$(dirname "$0")/service.yaml" -n "$NAMESPACE" || true
    kubectl delete -f "$(dirname "$0")/deployment.yaml" -n "$NAMESPACE" || true
    kubectl delete -f "$(dirname "$0")/pdb.yaml" -n "$NAMESPACE" || true
    kubectl delete -f "$(dirname "$0")/network-policies.yaml" -n "$NAMESPACE" || true
    kubectl delete -f "$(dirname "$0")/secrets.yaml" -n "$NAMESPACE" || true
    kubectl delete -f "$(dirname "$0")/configmap.yaml" -n "$NAMESPACE" || true
    kubectl delete -f "$(dirname "$0")/rbac.yaml" -n "$NAMESPACE" || true
    
    log_success "Cleanup completed"
}

# Main deployment function
deploy() {
    log_info "Starting deployment of $SERVICE_NAME to Kubernetes"
    log_info "Environment: $ENVIRONMENT"
    log_info "Image Tag: $IMAGE_TAG"
    log_info "Namespace: $NAMESPACE"
    
    check_kubectl
    check_namespace
    apply_manifests
    wait_for_deployment
    check_deployment_status
    get_endpoints
    
    # Wait a bit for services to be ready
    sleep 10
    
    health_check
    
    log_success "Deployment completed successfully!"
    log_info "You can now access the service at the endpoints shown above"
}

# Script usage
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help              Show this help message"
    echo "  -d, --deploy            Deploy the service"
    echo "  -c, --cleanup           Clean up resources"
    echo "  -s, --status            Check deployment status"
    echo "  -t, --tag TAG           Set image tag (default: latest)"
    echo "  -e, --env ENVIRONMENT   Set environment (default: production)"
    echo "  -n, --namespace NS      Set namespace (default: kyb-platform)"
    echo ""
    echo "Examples:"
    echo "  $0 --deploy --tag v1.0.0"
    echo "  $0 --status"
    echo "  $0 --cleanup"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            exit 0
            ;;
        -d|--deploy)
            DEPLOY=true
            shift
            ;;
        -c|--cleanup)
            CLEANUP=true
            shift
            ;;
        -s|--status)
            STATUS=true
            shift
            ;;
        -t|--tag)
            IMAGE_TAG="$2"
            shift 2
            ;;
        -e|--env)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -n|--namespace)
            NAMESPACE="$2"
            shift 2
            ;;
        *)
            log_error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Execute based on options
if [ "$CLEANUP" = true ]; then
    cleanup
elif [ "$STATUS" = true ]; then
    check_deployment_status
elif [ "$DEPLOY" = true ]; then
    deploy
else
    log_error "No action specified"
    usage
    exit 1
fi
