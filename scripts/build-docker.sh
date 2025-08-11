#!/bin/bash

# KYB Platform - Multi-Stage Docker Build Script
# Supports building for different environments and purposes

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
IMAGE_NAME="kyb-platform"
TAG="latest"
PLATFORM="linux/amd64"
PUSH=false
CACHE_FROM=""

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
    echo "Usage: $0 [OPTIONS] STAGE"
    echo ""
    echo "Stages:"
    echo "  development    - Build development image with hot reload"
    echo "  testing        - Build testing image and run tests"
    echo "  production     - Build production image (default)"
    echo "  debug          - Build debug image with debugging tools"
    echo "  minimal        - Build minimal production image"
    echo "  security       - Build security-scanning image"
    echo "  all            - Build all stages"
    echo ""
    echo "Options:"
    echo "  -t, --tag TAG      - Image tag (default: latest)"
    echo "  -p, --platform     - Target platform (default: linux/amd64)"
    echo "  --push             - Push image to registry after build"
    echo "  --cache-from IMAGE - Use image as cache source"
    echo "  -h, --help         - Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 production"
    echo "  $0 development -t dev"
    echo "  $0 all --push"
    echo "  $0 production -t v1.0.0 --push"
}

# Function to build a specific stage
build_stage() {
    local stage=$1
    local full_tag="${IMAGE_NAME}:${TAG}-${stage}"
    
    print_status "Building ${stage} stage..."
    
    case $stage in
        "development")
            docker build \
                --target development \
                --platform $PLATFORM \
                --cache-from $CACHE_FROM \
                -t $full_tag \
                -f Dockerfile .
            ;;
        "testing")
            docker build \
                --target testing \
                --platform $PLATFORM \
                --cache-from $CACHE_FROM \
                -t $full_tag \
                -f Dockerfile .
            ;;
        "production")
            docker build \
                --target production \
                --platform $PLATFORM \
                --cache-from $CACHE_FROM \
                -t $full_tag \
                -f Dockerfile .
            ;;
        "debug")
            docker build \
                --target debug \
                --platform $PLATFORM \
                --cache-from $CACHE_FROM \
                -t $full_tag \
                -f Dockerfile .
            ;;
        "minimal")
            docker build \
                --target production-minimal \
                --platform $PLATFORM \
                --cache-from $CACHE_FROM \
                -t $full_tag \
                -f Dockerfile .
            ;;
        "security")
            docker build \
                --target security-scan \
                --platform $PLATFORM \
                --cache-from $CACHE_FROM \
                -t $full_tag \
                -f Dockerfile .
            ;;
        *)
            print_error "Unknown stage: $stage"
            exit 1
            ;;
    esac
    
    print_success "Built ${stage} stage: $full_tag"
    
    if [ "$PUSH" = true ]; then
        print_status "Pushing $full_tag..."
        docker push $full_tag
        print_success "Pushed $full_tag"
    fi
}

# Function to build all stages
build_all() {
    print_status "Building all stages..."
    
    local stages=("development" "testing" "production" "debug" "minimal" "security")
    
    for stage in "${stages[@]}"; do
        build_stage $stage
    done
    
    print_success "Built all stages"
}

# Function to show image sizes
show_sizes() {
    print_status "Image sizes:"
    docker images | grep $IMAGE_NAME | sort
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--tag)
            TAG="$2"
            shift 2
            ;;
        -p|--platform)
            PLATFORM="$2"
            shift 2
            ;;
        --push)
            PUSH=true
            shift
            ;;
        --cache-from)
            CACHE_FROM="$2"
            shift 2
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        -*)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
        *)
            STAGE="$1"
            shift
            ;;
    esac
done

# Check if stage is provided
if [ -z "$STAGE" ]; then
    print_error "No stage specified"
    show_usage
    exit 1
fi

# Main execution
print_status "Starting build for stage: $STAGE"
print_status "Image name: $IMAGE_NAME"
print_status "Tag: $TAG"
print_status "Platform: $PLATFORM"
print_status "Push: $PUSH"

# Build based on stage
case $STAGE in
    "all")
        build_all
        ;;
    "development"|"testing"|"production"|"debug"|"minimal"|"security")
        build_stage $STAGE
        ;;
    *)
        print_error "Unknown stage: $STAGE"
        show_usage
        exit 1
        ;;
esac

# Show final image sizes
show_sizes

print_success "Build completed successfully!"
