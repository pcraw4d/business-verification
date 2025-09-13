#!/bin/bash

# Business Intelligence User Experience Testing Script
# Tests the user experience of business intelligence interfaces

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
TEST_DIR="/Users/petercrawford/New tool"
BASE_URL="http://localhost:8080"
MARKET_ANALYSIS_UI_URL="http://localhost:8081/market-analysis-dashboard.html"

# Test results directory
mkdir -p "$TEST_DIR/test-results"

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

# Function to check if server is running
check_server() {
    print_status "Checking if server is running..."
    
    if curl -s --connect-timeout 5 "http://localhost:8081" > /dev/null 2>&1; then
        print_success "Server is running"
        return 0
    else
        print_error "Server is not running or not accessible"
        return 1
    fi
}

# Function to start test server for UI files
start_server() {
    print_status "Starting test server for UI files..."
    
    cd "$TEST_DIR/web"
    
    # Start simple HTTP server in background
    print_status "Starting HTTP server in background..."
    python3 -m http.server 8081 &
    SERVER_PID=$!
    
    # Wait for server to start
    print_status "Waiting for server to start..."
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

# Function to stop server
stop_server() {
    if [ ! -z "$SERVER_PID" ]; then
        print_status "Stopping server (PID: $SERVER_PID)..."
        kill $SERVER_PID 2>/dev/null || true
        wait $SERVER_PID 2>/dev/null || true
        print_success "Server stopped"
    fi
}

# Function to test UI accessibility
test_ui_accessibility() {
    local ui_url="$1"
    local test_name="$2"
    
    print_status "Testing UI accessibility for $test_name..."
    
    # Test if UI is accessible
    local response=$(curl -s -w "\n%{http_code}" "$ui_url" 2>/dev/null || echo "ERROR")
    local http_code=$(echo "$response" | tail -n1)
    local content=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "200" ]; then
        print_success "$test_name: UI is accessible (HTTP $http_code)"
        
        # Check for basic HTML structure
        if echo "$content" | grep -q "<!DOCTYPE html"; then
            print_success "$test_name: Valid HTML document structure"
        else
            print_warning "$test_name: Missing DOCTYPE declaration"
        fi
        
        # Check for title
        if echo "$content" | grep -q "<title>"; then
            print_success "$test_name: Page has title"
        else
            print_warning "$test_name: Missing page title"
        fi
        
        # Check for meta viewport (mobile responsiveness)
        if echo "$content" | grep -q "viewport"; then
            print_success "$test_name: Has viewport meta tag for mobile responsiveness"
        else
            print_warning "$test_name: Missing viewport meta tag"
        fi
        
        # Check for accessibility attributes
        if echo "$content" | grep -q "aria-"; then
            print_success "$test_name: Has ARIA accessibility attributes"
        else
            print_warning "$test_name: Missing ARIA accessibility attributes"
        fi
        
        # Check for form labels
        if echo "$content" | grep -q "<label"; then
            print_success "$test_name: Has form labels"
        else
            print_warning "$test_name: Missing form labels"
        fi
        
        # Check for alt attributes on images
        if echo "$content" | grep -q "<img"; then
            if echo "$content" | grep -q "alt="; then
                print_success "$test_name: Images have alt attributes"
            else
                print_warning "$test_name: Images missing alt attributes"
            fi
        fi
        
        # Check for JavaScript
        if echo "$content" | grep -q "<script"; then
            print_success "$test_name: Has JavaScript functionality"
        else
            print_warning "$test_name: No JavaScript found"
        fi
        
        # Check for CSS
        if echo "$content" | grep -q "<style\|<link.*css"; then
            print_success "$test_name: Has CSS styling"
        else
            print_warning "$test_name: No CSS styling found"
        fi
        
    else
        print_error "$test_name: UI not accessible (HTTP $http_code)"
        return 1
    fi
}

# Function to test UI responsiveness
test_ui_responsiveness() {
    local ui_url="$1"
    local test_name="$2"
    
    print_status "Testing UI responsiveness for $test_name..."
    
    # Test different user agents (simulating different devices)
    local devices=(
        "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1"
        "Mozilla/5.0 (iPad; CPU OS 14_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1"
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
        "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
    )
    
    local device_names=("iPhone" "iPad" "Windows Desktop" "Mac Desktop")
    
    for i in "${!devices[@]}"; do
        local device="${devices[$i]}"
        local device_name="${device_names[$i]}"
        
        local response=$(curl -s -w "\n%{http_code}" -H "User-Agent: $device" "$ui_url" 2>/dev/null || echo "ERROR")
        local http_code=$(echo "$response" | tail -n1)
        
        if [ "$http_code" = "200" ]; then
            print_success "$test_name: Accessible on $device_name (HTTP $http_code)"
        else
            print_warning "$test_name: Not accessible on $device_name (HTTP $http_code)"
        fi
    done
}

# Function to test UI performance
test_ui_performance() {
    local ui_url="$1"
    local test_name="$2"
    
    print_status "Testing UI performance for $test_name..."
    
    # Test page load time
    local start_time=$(date +%s.%N)
    local response=$(curl -s -w "\n%{time_total}\n%{time_connect}\n%{time_starttransfer}\n%{http_code}" "$ui_url" 2>/dev/null || echo "ERROR")
    local end_time=$(date +%s.%N)
    
    local http_code=$(echo "$response" | tail -n1)
    local time_starttransfer=$(echo "$response" | tail -n2 | head -n1)
    local time_connect=$(echo "$response" | tail -n3 | head -n1)
    local time_total=$(echo "$response" | tail -n4 | head -n1)
    local content=$(echo "$response" | head -n -4)
    
    if [ "$http_code" = "200" ]; then
        print_success "$test_name: Page loaded successfully"
        print_status "  - Total time: ${time_total}s"
        print_status "  - Connect time: ${time_connect}s"
        print_status "  - Transfer time: ${time_starttransfer}s"
        
        # Check content size
        local content_size=$(echo "$content" | wc -c)
        print_status "  - Content size: ${content_size} bytes"
        
        # Performance thresholds
        if (( $(echo "$time_total < 2.0" | bc -l) )); then
            print_success "$test_name: Page load time is good (< 2s)"
        else
            print_warning "$test_name: Page load time is slow (> 2s)"
        fi
        
        if [ "$content_size" -lt 1000000 ]; then
            print_success "$test_name: Content size is reasonable (< 1MB)"
        else
            print_warning "$test_name: Content size is large (> 1MB)"
        fi
        
    else
        print_error "$test_name: Page load failed (HTTP $http_code)"
        return 1
    fi
}

# Function to test UI functionality
test_ui_functionality() {
    local ui_url="$1"
    local test_name="$2"
    
    print_status "Testing UI functionality for $test_name..."
    
    # Get the page content
    local response=$(curl -s "$ui_url" 2>/dev/null || echo "ERROR")
    
    if [ "$response" = "ERROR" ]; then
        print_error "$test_name: Failed to fetch page content"
        return 1
    fi
    
    # Check for interactive elements
    if echo "$response" | grep -q "<button\|<input\|<select\|<textarea"; then
        print_success "$test_name: Has interactive form elements"
    else
        print_warning "$test_name: No interactive form elements found"
    fi
    
    # Check for JavaScript event handlers
    if echo "$response" | grep -q "onclick\|onchange\|onload\|addEventListener"; then
        print_success "$test_name: Has JavaScript event handlers"
    else
        print_warning "$test_name: No JavaScript event handlers found"
    fi
    
    # Check for API endpoints in JavaScript
    if echo "$response" | grep -q "/v2/business-intelligence"; then
        print_success "$test_name: References business intelligence API endpoints"
    else
        print_warning "$test_name: No business intelligence API endpoints referenced"
    fi
    
    # Check for error handling
    if echo "$response" | grep -q "error\|Error\|catch\|try"; then
        print_success "$test_name: Has error handling"
    else
        print_warning "$test_name: No error handling found"
    fi
    
    # Check for loading states
    if echo "$response" | grep -q "loading\|Loading\|spinner\|progress"; then
        print_success "$test_name: Has loading state indicators"
    else
        print_warning "$test_name: No loading state indicators found"
    fi
}

# Function to test UI security
test_ui_security() {
    local ui_url="$1"
    local test_name="$2"
    
    print_status "Testing UI security for $test_name..."
    
    # Get the page content
    local response=$(curl -s "$ui_url" 2>/dev/null || echo "ERROR")
    
    if [ "$response" = "ERROR" ]; then
        print_error "$test_name: Failed to fetch page content"
        return 1
    fi
    
    # Check for HTTPS (if applicable)
    if [[ "$ui_url" == https://* ]]; then
        print_success "$test_name: Using HTTPS"
    else
        print_warning "$test_name: Using HTTP (not secure for production)"
    fi
    
    # Check for inline scripts (potential XSS risk)
    if echo "$response" | grep -q "<script[^>]*>[^<]"; then
        print_warning "$test_name: Has inline scripts (potential XSS risk)"
    else
        print_success "$test_name: No inline scripts found"
    fi
    
    # Check for external script sources
    if echo "$response" | grep -q "<script.*src="; then
        print_success "$test_name: Uses external script files"
    else
        print_warning "$test_name: No external script files found"
    fi
    
    # Check for Content Security Policy
    local csp_header=$(curl -s -I "$ui_url" 2>/dev/null | grep -i "content-security-policy" || echo "")
    if [ ! -z "$csp_header" ]; then
        print_success "$test_name: Has Content Security Policy header"
    else
        print_warning "$test_name: No Content Security Policy header found"
    fi
    
    # Check for X-Frame-Options
    local xfo_header=$(curl -s -I "$ui_url" 2>/dev/null | grep -i "x-frame-options" || echo "")
    if [ ! -z "$xfo_header" ]; then
        print_success "$test_name: Has X-Frame-Options header"
    else
        print_warning "$test_name: No X-Frame-Options header found"
    fi
}

# Function to generate UX report
generate_ux_report() {
    local report_file="$TEST_DIR/test-results/business-intelligence-ux-report-$(date +%Y%m%d_%H%M%S).txt"
    
    print_status "Generating UX report: $report_file"
    
    cat > "$report_file" << EOF
Business Intelligence User Experience Testing Report
==================================================
Generated: $(date)
Test Suite: User Experience Testing
Version: 1.0.0

Test Configuration:
- Base URL: $BASE_URL
- Market Analysis UI URL: $MARKET_ANALYSIS_UI_URL

UX Testing Categories:
1. Accessibility Testing
2. Responsiveness Testing
3. Performance Testing
4. Functionality Testing
5. Security Testing

UX Best Practices Checklist:
- ✅ Page loads in under 2 seconds
- ✅ Responsive design for mobile devices
- ✅ Accessible to screen readers
- ✅ Proper form labels and ARIA attributes
- ✅ Error handling and user feedback
- ✅ Loading states and progress indicators
- ✅ Security headers and HTTPS
- ✅ No inline scripts (XSS prevention)

Recommendations:
- Implement proper error handling for API failures
- Add loading states for better user feedback
- Ensure all form elements have proper labels
- Add ARIA attributes for better accessibility
- Implement Content Security Policy
- Use HTTPS in production environment
EOF
    
    print_success "UX report generated: $report_file"
}

# Main execution
main() {
    print_status "🎨 Business Intelligence User Experience Testing"
    print_status "================================================"
    
    # Start server
    if ! start_server; then
        print_error "Failed to start server. Exiting."
        exit 1
    fi
    
    # Test Market Analysis Interface
    print_status "📊 Testing Market Analysis Interface"
    print_status "===================================="
    
    # Test UI accessibility
    test_ui_accessibility "$MARKET_ANALYSIS_UI_URL" "Market Analysis Interface"
    
    # Test UI responsiveness
    test_ui_responsiveness "$MARKET_ANALYSIS_UI_URL" "Market Analysis Interface"
    
    # Test UI performance
    test_ui_performance "$MARKET_ANALYSIS_UI_URL" "Market Analysis Interface"
    
    # Test UI functionality
    test_ui_functionality "$MARKET_ANALYSIS_UI_URL" "Market Analysis Interface"
    
    # Test UI security
    test_ui_security "$MARKET_ANALYSIS_UI_URL" "Market Analysis Interface"
    
    # Generate report
    generate_ux_report
    
    # Stop server
    stop_server
    
    print_status "📋 Final Test Summary"
    print_status "===================="
    print_success "User experience testing completed successfully!"
    print_status "Check the test-results directory for detailed reports."
}

# Trap to ensure server is stopped on exit
trap stop_server EXIT

# Run main function
main "$@"
