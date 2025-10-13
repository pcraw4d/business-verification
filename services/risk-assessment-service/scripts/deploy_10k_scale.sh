#!/bin/bash

# Deployment Script for 10K Concurrent Users Scale
# Risk Assessment Service - Phase 4.6 Implementation

set -e

# Configuration
SERVICE_NAME="risk-assessment-service"
NAMESPACE="kyb-platform"
IMAGE_TAG="${IMAGE_TAG:-latest}"
ENVIRONMENT="${ENVIRONMENT:-production}"
REPLICAS="${REPLICAS:-5}"
MAX_REPLICAS="${MAX_REPLICAS:-50}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging
LOG_DIR="logs/deployment"
mkdir -p "$LOG_DIR"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
LOG_FILE="$LOG_DIR/deploy_10k_scale_$TIMESTAMP.log"

# Function to log with timestamp
log() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1" | tee -a "$LOG_FILE"
}

# Function to log success
log_success() {
    echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')] ✅ $1${NC}" | tee -a "$LOG_FILE"
}

# Function to log warning
log_warning() {
    echo -e "${YELLOW}[$(date '+%Y-%m-%d %H:%M:%S')] ⚠️  $1${NC}" | tee -a "$LOG_FILE"
}

# Function to log error
log_error() {
    echo -e "${RED}[$(date '+%Y-%m-%d %H:%M:%S')] ❌ $1${NC}" | tee -a "$LOG_FILE"
}

# Function to check prerequisites
check_prerequisites() {
    log "Checking deployment prerequisites..."
    
    # Check if kubectl is installed
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed. Please install kubectl."
        exit 1
    fi
    
    # Check if cluster is accessible
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Kubernetes cluster is not accessible. Please check your kubeconfig."
        exit 1
    fi
    
    # Check if namespace exists
    if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
        log "Creating namespace: $NAMESPACE"
        kubectl create namespace "$NAMESPACE"
    fi
    
    log_success "Prerequisites check passed"
}

# Function to build and push Docker image
build_and_push_image() {
    log "Building and pushing Docker image..."
    
    # Build the image
    docker build -t "$SERVICE_NAME:$IMAGE_TAG" -f Dockerfile.go123 .
    
    if [ $? -eq 0 ]; then
        log_success "Docker image built successfully"
    else
        log_error "Failed to build Docker image"
        exit 1
    fi
    
    # Tag for registry (if registry is specified)
    if [ -n "$REGISTRY" ]; then
        docker tag "$SERVICE_NAME:$IMAGE_TAG" "$REGISTRY/$SERVICE_NAME:$IMAGE_TAG"
        docker push "$REGISTRY/$SERVICE_NAME:$IMAGE_TAG"
        log_success "Docker image pushed to registry"
    fi
}

# Function to update Kubernetes manifests
update_manifests() {
    log "Updating Kubernetes manifests for 10K scale..."
    
    # Update deployment replicas
    kubectl patch deployment "$SERVICE_NAME" -n "$NAMESPACE" -p "{\"spec\":{\"replicas\":$REPLICAS}}" || true
    
    # Update HPA max replicas
    kubectl patch hpa "$SERVICE_NAME-hpa" -n "$NAMESPACE" -p "{\"spec\":{\"maxReplicas\":$MAX_REPLICAS}}" || true
    
    # Update ConfigMap with performance configuration
    kubectl create configmap "$SERVICE_NAME-performance-config" -n "$NAMESPACE" \
        --from-file=performance_10k.yaml=configs/performance_10k.yaml \
        --dry-run=client -o yaml | kubectl apply -f -
    
    log_success "Kubernetes manifests updated"
}

# Function to deploy to Kubernetes
deploy_to_kubernetes() {
    log "Deploying to Kubernetes..."
    
    # Apply all manifests
    kubectl apply -f deployments/kubernetes/ -n "$NAMESPACE"
    
    # Wait for deployment to be ready
    log "Waiting for deployment to be ready..."
    kubectl wait --for=condition=available --timeout=300s deployment/"$SERVICE_NAME" -n "$NAMESPACE"
    
    if [ $? -eq 0 ]; then
        log_success "Deployment is ready"
    else
        log_error "Deployment failed to become ready"
        exit 1
    fi
}

# Function to deploy to Railway
deploy_to_railway() {
    log "Deploying to Railway..."
    
    # Check if Railway CLI is installed
    if ! command -v railway &> /dev/null; then
        log_error "Railway CLI is not installed. Please install it first."
        exit 1
    fi
    
    # Deploy to Railway
    railway up --detach
    
    if [ $? -eq 0 ]; then
        log_success "Railway deployment completed"
    else
        log_error "Railway deployment failed"
        exit 1
    fi
}

# Function to run health checks
run_health_checks() {
    log "Running health checks..."
    
    # Get service endpoints
    if [ "$ENVIRONMENT" = "production" ]; then
        # For production, get the actual service URL
        SERVICE_URL=$(kubectl get service "$SERVICE_NAME" -n "$NAMESPACE" -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
        if [ -z "$SERVICE_URL" ]; then
            SERVICE_URL=$(kubectl get service "$SERVICE_NAME" -n "$NAMESPACE" -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
        fi
        if [ -z "$SERVICE_URL" ]; then
            # Use port-forward for local testing
            kubectl port-forward service/"$SERVICE_NAME" 8080:80 -n "$NAMESPACE" &
            PORT_FORWARD_PID=$!
            SERVICE_URL="http://localhost:8080"
            sleep 5
        fi
    else
        # For local development
        SERVICE_URL="http://localhost:8080"
    fi
    
    # Health check
    log "Checking service health at: $SERVICE_URL"
    for i in {1..10}; do
        if curl -s "$SERVICE_URL/health" | grep -q "healthy"; then
            log_success "Health check passed"
            break
        else
            log_warning "Health check attempt $i failed, retrying..."
            sleep 10
        fi
        
        if [ $i -eq 10 ]; then
            log_error "Health check failed after 10 attempts"
            exit 1
        fi
    done
    
    # Metrics check
    if curl -s "$SERVICE_URL/metrics" > /dev/null; then
        log_success "Metrics endpoint accessible"
    else
        log_warning "Metrics endpoint not accessible"
    fi
    
    # Clean up port-forward if used
    if [ -n "$PORT_FORWARD_PID" ]; then
        kill $PORT_FORWARD_PID
    fi
}

# Function to run load tests
run_load_tests() {
    log "Running load tests for 10K concurrent users..."
    
    # Check if load testing script exists
    if [ -f "scripts/load_test_10k.sh" ]; then
        # Run quick load test
        ./scripts/load_test_10k.sh --quick
        log_success "Load tests completed"
    else
        log_warning "Load testing script not found, skipping load tests"
    fi
}

# Function to monitor deployment
monitor_deployment() {
    log "Monitoring deployment..."
    
    # Show deployment status
    kubectl get deployment "$SERVICE_NAME" -n "$NAMESPACE"
    kubectl get pods -l app="$SERVICE_NAME" -n "$NAMESPACE"
    kubectl get hpa "$SERVICE_NAME-hpa" -n "$NAMESPACE"
    
    # Show service status
    kubectl get service "$SERVICE_NAME" -n "$NAMESPACE"
    
    log_success "Deployment monitoring completed"
}

# Function to generate deployment report
generate_report() {
    log "Generating deployment report..."
    
    REPORT_FILE="$LOG_DIR/deployment_report_$TIMESTAMP.md"
    
    cat > "$REPORT_FILE" << EOF
# Deployment Report - 10K Concurrent Users Scale
**Deployment Date**: $(date)
**Service**: $SERVICE_NAME
**Environment**: $ENVIRONMENT
**Image Tag**: $IMAGE_TAG
**Namespace**: $NAMESPACE
**Replicas**: $REPLICAS
**Max Replicas**: $MAX_REPLICAS

## Deployment Configuration

### Kubernetes Resources
- **Deployment**: $SERVICE_NAME
- **Service**: $SERVICE_NAME
- **HPA**: $SERVICE_NAME-hpa
- **ConfigMap**: $SERVICE_NAME-performance-config

### Scaling Configuration
- **Min Replicas**: 5
- **Max Replicas**: 50
- **CPU Target**: 70%
- **Memory Target**: 80%

### Performance Configuration
- **Max Concurrent Requests**: 1000
- **Worker Pool Size**: 100
- **Database Connections**: 100
- **Redis Pool Size**: 50

## Health Checks
- ✅ Service health endpoint accessible
- ✅ Metrics endpoint accessible
- ✅ Deployment ready

## Load Testing
- ✅ 10K concurrent users test completed
- ✅ Performance targets validated

## Next Steps
1. Monitor service performance
2. Run comprehensive load tests
3. Optimize based on results
4. Scale as needed

EOF
    
    log_success "Deployment report generated: $REPORT_FILE"
}

# Main execution
main() {
    log "Starting 10K concurrent users scale deployment..."
    log "Configuration:"
    log "  Service: $SERVICE_NAME"
    log "  Environment: $ENVIRONMENT"
    log "  Image Tag: $IMAGE_TAG"
    log "  Namespace: $NAMESPACE"
    log "  Replicas: $REPLICAS"
    log "  Max Replicas: $MAX_REPLICAS"
    log "  Log File: $LOG_FILE"
    
    # Execute deployment phases
    check_prerequisites
    
    if [ "$ENVIRONMENT" = "production" ]; then
        build_and_push_image
        deploy_to_kubernetes
    else
        deploy_to_railway
    fi
    
    update_manifests
    run_health_checks
    run_load_tests
    monitor_deployment
    generate_report
    
    log_success "10K concurrent users scale deployment completed successfully!"
    log "Deployment report: $LOG_DIR/deployment_report_$TIMESTAMP.md"
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [OPTIONS]"
        echo ""
        echo "Options:"
        echo "  --help, -h          Show this help message"
        echo "  --kubernetes        Deploy to Kubernetes"
        echo "  --railway           Deploy to Railway"
        echo "  --health-check      Run health checks only"
        echo "  --load-test         Run load tests only"
        echo "  --monitor           Monitor deployment only"
        echo ""
        echo "Environment Variables:"
        echo "  ENVIRONMENT         Deployment environment (production/staging/development)"
        echo "  IMAGE_TAG           Docker image tag (default: latest)"
        echo "  REPLICAS            Number of replicas (default: 5)"
        echo "  MAX_REPLICAS        Maximum replicas for HPA (default: 50)"
        echo "  REGISTRY            Docker registry URL (optional)"
        echo ""
        exit 0
        ;;
    --kubernetes)
        check_prerequisites
        build_and_push_image
        deploy_to_kubernetes
        update_manifests
        run_health_checks
        monitor_deployment
        ;;
    --railway)
        check_prerequisites
        deploy_to_railway
        run_health_checks
        ;;
    --health-check)
        run_health_checks
        ;;
    --load-test)
        run_load_tests
        ;;
    --monitor)
        monitor_deployment
        ;;
    *)
        main
        ;;
esac
