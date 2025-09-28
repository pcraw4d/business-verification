#!/bin/bash

# KYB Platform Microservices Fixes Testing Script
# This script tests the fixes locally and verifies they work

set -e

echo "🧪 Testing KYB Platform Microservices Fixes"
echo "==========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Function to test service compilation
test_service_compilation() {
    local service_name=$1
    local service_path=$2
    local binary_name=$3
    
    print_status "Testing compilation of $service_name..."
    
    if [ ! -d "$service_path" ]; then
        print_error "Service directory not found: $service_path"
        return 1
    fi
    
    cd "$service_path"
    
    if go build -o "$binary_name" main.go 2>/dev/null; then
        print_success "$service_name compiles successfully"
        rm -f "$binary_name"  # Clean up
        cd - > /dev/null
        return 0
    else
        print_error "$service_name compilation failed"
        cd - > /dev/null
        return 1
    fi
}

# Function to test service locally
test_service_locally() {
    local service_name=$1
    local service_path=$2
    local binary_name=$3
    local port=$4
    local health_endpoint=$5
    
    print_status "Testing $service_name locally..."
    
    cd "$service_path"
    
    # Build the service
    if ! go build -o "$binary_name" main.go 2>/dev/null; then
        print_error "$service_name build failed"
        cd - > /dev/null
        return 1
    fi
    
    # Start the service in background
    PORT="$port" ./"$binary_name" &
    local service_pid=$!
    
    # Wait for service to start
    sleep 3
    
    # Test health endpoint
    if curl -s -f "http://localhost:$port$health_endpoint" > /dev/null; then
        print_success "$service_name is running and healthy on port $port"
        kill $service_pid 2>/dev/null || true
        rm -f "$binary_name"  # Clean up
        cd - > /dev/null
        return 0
    else
        print_error "$service_name health check failed"
        kill $service_pid 2>/dev/null || true
        rm -f "$binary_name"  # Clean up
        cd - > /dev/null
        return 1
    fi
}

# Main testing process
main() {
    print_status "Starting microservices fixes testing..."
    
    # Test compilation of all services
    print_status "=== 1. Testing Service Compilation ==="
    
    compilation_tests=(
        "Frontend Service:services/frontend:frontend-server"
        "Business Intelligence Gateway:cmd/business-intelligence-gateway:kyb-business-intelligence-gateway"
        "Service Discovery:cmd/service-discovery:kyb-service-discovery"
    )
    
    compilation_success=0
    total_compilation_tests=${#compilation_tests[@]}
    
    for test_info in "${compilation_tests[@]}"; do
        IFS=':' read -r name path binary <<< "$test_info"
        if test_service_compilation "$name" "$path" "$binary"; then
            ((compilation_success++))
        fi
    done
    
    # Test services locally
    print_status "=== 2. Testing Services Locally ==="
    
    local_tests=(
        "Service Discovery:cmd/service-discovery:kyb-service-discovery:8086:/health"
        "Business Intelligence Gateway:cmd/business-intelligence-gateway:kyb-business-intelligence-gateway:8087:/health"
    )
    
    local_success=0
    total_local_tests=${#local_tests[@]}
    
    for test_info in "${local_tests[@]}"; do
        IFS=':' read -r name path binary port health <<< "$test_info"
        if test_service_locally "$name" "$path" "$binary" "$port" "$health"; then
            ((local_success++))
        fi
    done
    
    # Test frontend fix
    print_status "=== 3. Testing Frontend Fix ==="
    
    cd services/frontend
    
    if [ -d "public" ]; then
        print_success "Frontend public directory exists"
        
        # Check if main.go serves from public directory
        if grep -q 'http.Dir("./public/")' cmd/main.go; then
            print_success "Frontend serves from correct public directory"
        else
            print_error "Frontend still serves from wrong directory"
        fi
        
        # Check if health endpoint is added
        if grep -q 'handleFunc("/health"' cmd/main.go; then
            print_success "Frontend has health check endpoint"
        else
            print_error "Frontend missing health check endpoint"
        fi
    else
        print_error "Frontend public directory not found"
    fi
    
    cd - > /dev/null
    
    # Summary
    print_status "=== Testing Summary ==="
    print_status "Compilation Tests: $compilation_success/$total_compilation_tests passed"
    print_status "Local Tests: $local_success/$total_local_tests passed"
    
    if [ $compilation_success -eq $total_compilation_tests ] && [ $local_success -eq $total_local_tests ]; then
        print_success "All tests passed! 🎉"
        print_status "The microservices fixes are ready for deployment."
    else
        print_warning "Some tests failed. Please check the issues above."
    fi
    
    # Service URLs for testing
    print_status "=== Service URLs for Testing ==="
    echo "Service Discovery Dashboard: http://localhost:8086/dashboard"
    echo "Business Intelligence Gateway: http://localhost:8087/dashboard/executive"
    echo "Frontend Service: http://localhost:8080 (after deployment)"
    
    print_success "Microservices fixes testing completed!"
}

# Run main function
main "$@"
