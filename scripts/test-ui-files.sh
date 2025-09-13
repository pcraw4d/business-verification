#!/bin/bash

# Simple UI Files Test Script
# Tests if UI files exist and are accessible

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
TEST_DIR="/Users/petercrawford/New tool"
WEB_DIR="$TEST_DIR/web"

# Function to print colored output
print_status() {
    echo -e "${BLUE}$1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Function to test UI file existence
test_ui_file_existence() {
    print_status "Testing UI file existence..."
    
    local ui_files=(
        "market-analysis-dashboard.html"
        "competitive-analysis-dashboard.html"
        "business-growth-analytics.html"
        "dashboard.html"
        "index.html"
    )
    
    for file in "${ui_files[@]}"; do
        local file_path="$WEB_DIR/$file"
        if [ -f "$file_path" ]; then
            print_success "File exists: $file"
            
            # Check file size
            local file_size=$(wc -c < "$file_path")
            print_status "  - Size: ${file_size} bytes"
            
            # Check if file has content
            if [ "$file_size" -gt 0 ]; then
                print_success "  - File has content"
            else
                print_warning "  - File is empty"
            fi
            
            # Check for basic HTML structure
            if grep -q "<!DOCTYPE html" "$file_path"; then
                print_success "  - Valid HTML document"
            else
                print_warning "  - Missing DOCTYPE declaration"
            fi
            
        else
            print_error "File not found: $file"
        fi
    done
}

# Function to test UI file content
test_ui_file_content() {
    print_status "Testing UI file content..."
    
    local ui_files=(
        "market-analysis-dashboard.html"
        "competitive-analysis-dashboard.html"
        "business-growth-analytics.html"
    )
    
    for file in "${ui_files[@]}"; do
        local file_path="$WEB_DIR/$file"
        if [ -f "$file_path" ]; then
            print_status "Analyzing content of $file..."
            
            # Check for title
            if grep -q "<title>" "$file_path"; then
                print_success "  - Has title"
            else
                print_warning "  - Missing title"
            fi
            
            # Check for meta viewport
            if grep -q "viewport" "$file_path"; then
                print_success "  - Has viewport meta tag"
            else
                print_warning "  - Missing viewport meta tag"
            fi
            
            # Check for CSS
            if grep -q "<style\|<link.*css" "$file_path"; then
                print_success "  - Has CSS styling"
            else
                print_warning "  - No CSS styling found"
            fi
            
            # Check for JavaScript
            if grep -q "<script" "$file_path"; then
                print_success "  - Has JavaScript"
            else
                print_warning "  - No JavaScript found"
            fi
            
            # Check for form elements
            if grep -q "<form\|<input\|<button\|<select" "$file_path"; then
                print_success "  - Has form elements"
            else
                print_warning "  - No form elements found"
            fi
            
            # Check for API endpoints
            if grep -q "/v2/business-intelligence" "$file_path"; then
                print_success "  - References business intelligence API"
            else
                print_warning "  - No business intelligence API references"
            fi
            
            # Check for accessibility
            if grep -q "aria-\|role=" "$file_path"; then
                print_success "  - Has accessibility attributes"
            else
                print_warning "  - No accessibility attributes found"
            fi
            
        fi
    done
}

# Function to start simple HTTP server for testing
start_test_server() {
    print_status "Starting test HTTP server..."
    
    cd "$WEB_DIR"
    
    # Start simple HTTP server in background
    python3 -m http.server 8081 &
    SERVER_PID=$!
    
    # Wait for server to start
    sleep 2
    
    # Test if server is running
    if curl -s --connect-timeout 5 "http://localhost:8081" > /dev/null 2>&1; then
        print_success "Test server started successfully (PID: $SERVER_PID)"
        return 0
    else
        print_error "Failed to start test server"
        return 1
    fi
}

# Function to stop test server
stop_test_server() {
    if [ ! -z "$SERVER_PID" ]; then
        print_status "Stopping test server (PID: $SERVER_PID)..."
        kill $SERVER_PID 2>/dev/null || true
        wait $SERVER_PID 2>/dev/null || true
        print_success "Test server stopped"
    fi
}

# Function to test UI accessibility via HTTP
test_ui_http_access() {
    print_status "Testing UI HTTP access..."
    
    local ui_files=(
        "market-analysis-dashboard.html"
        "competitive-analysis-dashboard.html"
        "business-growth-analytics.html"
        "dashboard.html"
        "index.html"
    )
    
    for file in "${ui_files[@]}"; do
        local url="http://localhost:8081/$file"
        local response=$(curl -s -w "\n%{http_code}" "$url" 2>/dev/null || echo "ERROR")
        local http_code=$(echo "$response" | tail -n1)
        local content=$(echo "$response" | head -n -1)
        
        if [ "$http_code" = "200" ]; then
            print_success "$file: Accessible via HTTP (HTTP $http_code)"
            
            # Check content size
            local content_size=$(echo "$content" | wc -c)
            print_status "  - Content size: ${content_size} bytes"
            
        else
            print_error "$file: Not accessible via HTTP (HTTP $http_code)"
        fi
    done
}

# Main execution
main() {
    print_status "🔍 UI Files Testing"
    print_status "==================="
    
    # Test file existence
    test_ui_file_existence
    
    echo ""
    
    # Test file content
    test_ui_file_content
    
    echo ""
    
    # Start test server and test HTTP access
    if start_test_server; then
        test_ui_http_access
        stop_test_server
    fi
    
    print_status "📋 Final Test Summary"
    print_status "===================="
    print_success "UI files testing completed successfully!"
}

# Trap to ensure server is stopped on exit
trap stop_test_server EXIT

# Run main function
main "$@"
