#!/bin/bash

# Manual Workflow Testing Script
# Provides guided manual testing procedures for business intelligence workflows

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
TEST_DIR="/Users/petercrawford/New tool"
BASE_URL="http://localhost:8080"
UI_URL="http://localhost:8081"
TEST_RESULTS_DIR="$TEST_DIR/test-results"

# Test results directory
mkdir -p "$TEST_RESULTS_DIR"

# Function to print colored output
print_header() {
    echo -e "${PURPLE}$1${NC}"
}

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

print_info() {
    echo -e "${CYAN}ℹ️  $1${NC}"
}

print_instruction() {
    echo -e "${YELLOW}📋 $1${NC}"
}

# Function to start servers for manual testing
start_servers() {
    print_status "Starting servers for manual testing..."
    
    # Start API server
    cd "$TEST_DIR"
    if ! curl -s --connect-timeout 5 "$BASE_URL/health" > /dev/null 2>&1; then
        print_status "Starting API server..."
        go run cmd/api-enhanced/main-enhanced-with-database-classification.go &
        API_SERVER_PID=$!
        sleep 5
    else
        print_success "API server is already running"
    fi
    
    # Start UI server
    cd "$TEST_DIR/web"
    if ! curl -s --connect-timeout 5 "$UI_URL" > /dev/null 2>&1; then
        print_status "Starting UI server..."
        python3 -m http.server 8081 &
        UI_SERVER_PID=$!
        sleep 2
    else
        print_success "UI server is already running"
    fi
}

# Function to stop servers
stop_servers() {
    if [ ! -z "$API_SERVER_PID" ]; then
        print_status "Stopping API server (PID: $API_SERVER_PID)..."
        kill $API_SERVER_PID 2>/dev/null || true
        wait $API_SERVER_PID 2>/dev/null || true
    fi
    
    if [ ! -z "$UI_SERVER_PID" ]; then
        print_status "Stopping UI server (PID: $UI_SERVER_PID)..."
        kill $UI_SERVER_PID 2>/dev/null || true
        wait $UI_SERVER_PID 2>/dev/null || true
    fi
}

# Function to display manual testing instructions
display_manual_testing_instructions() {
    print_header "📋 Manual Workflow Testing Instructions"
    print_header "======================================"
    
    echo ""
    print_info "This script will guide you through manual testing procedures for the business intelligence system."
    echo ""
    
    print_instruction "Manual Testing Workflows:"
    echo "1. Market Analysis Workflow Testing"
    echo "2. Competitive Analysis Workflow Testing"
    echo "3. Growth Analytics Workflow Testing"
    echo "4. Business Intelligence Aggregation Testing"
    echo "5. Error Handling and Edge Cases Testing"
    echo "6. User Interface Navigation Testing"
    echo ""
    
    print_instruction "Prerequisites:"
    echo "- API server running on $BASE_URL"
    echo "- UI server running on $UI_URL"
    echo "- Web browser for UI testing"
    echo "- API testing tool (curl, Postman, or similar)"
    echo ""
    
    print_instruction "Testing Approach:"
    echo "- Follow each workflow step by step"
    echo "- Document any issues or observations"
    echo "- Test both positive and negative scenarios"
    echo "- Verify data accuracy and system behavior"
    echo ""
}

# Function to test market analysis workflow
test_market_analysis_workflow() {
    print_header "🔍 Market Analysis Workflow Testing"
    print_status "=================================="
    
    print_instruction "Step 1: Test Market Analysis API Endpoint"
    echo "Test the market analysis API endpoint with the following curl command:"
    echo ""
    echo "curl -X POST $BASE_URL/v2/business-intelligence/market-analysis \\"
    echo "  -H 'Content-Type: application/json' \\"
    echo "  -d '{"
    echo "    \"business_id\": \"test-business-123\","
    echo "    \"time_range\": {"
    echo "      \"start_date\": \"2024-01-01T00:00:00Z\","
    echo "      \"end_date\": \"2024-12-31T23:59:59Z\""
    echo "    },"
    echo "    \"options\": {"
    echo "      \"include_competitors\": true,"
    echo "      \"include_trends\": true,"
    echo "      \"include_forecasts\": true"
    echo "    }"
    echo "  }'"
    echo ""
    
    print_instruction "Expected Results:"
    echo "- HTTP status code: 200 (if implemented) or 501 (not implemented)"
    echo "- Response should contain market analysis data"
    echo "- Response time should be under 5 seconds"
    echo ""
    
    print_instruction "Step 2: Test Market Analysis UI"
    echo "Open your web browser and navigate to:"
    echo "$UI_URL/market-analysis-dashboard.html"
    echo ""
    echo "Manual UI Testing Checklist:"
    echo "- [ ] Page loads successfully"
    echo "- [ ] All form fields are visible and functional"
    echo "- [ ] Date pickers work correctly"
    echo "- [ ] Submit button is clickable"
    echo "- [ ] Error messages display appropriately"
    echo "- [ ] Loading states are shown during processing"
    echo "- [ ] Results are displayed in a readable format"
    echo ""
    
    print_instruction "Step 3: Test Market Analysis Job Creation"
    echo "Test the job creation endpoint:"
    echo ""
    echo "curl -X POST $BASE_URL/v2/business-intelligence/market-analysis/jobs \\"
    echo "  -H 'Content-Type: application/json' \\"
    echo "  -d '{"
    echo "    \"business_id\": \"test-business-123\","
    echo "    \"time_range\": {"
    echo "      \"start_date\": \"2024-01-01T00:00:00Z\","
    echo "      \"end_date\": \"2024-12-31T23:59:59Z\""
    echo "    }"
    echo "  }'"
    echo ""
    
    print_instruction "Step 4: Test Market Analysis Job Status"
    echo "Check job status using the job ID from the previous response:"
    echo ""
    echo "curl -X GET $BASE_URL/v2/business-intelligence/market-analysis/jobs/{job_id}"
    echo ""
    
    print_instruction "Step 5: Test Market Analysis Results Retrieval"
    echo "Retrieve analysis results:"
    echo ""
    echo "curl -X GET $BASE_URL/v2/business-intelligence/market-analysis/{analysis_id}"
    echo ""
    
    echo ""
    print_instruction "Document your findings:"
    echo "- Record any errors or unexpected behavior"
    echo "- Note response times and data accuracy"
    echo "- Document UI usability issues"
    echo ""
}

# Function to test competitive analysis workflow
test_competitive_analysis_workflow() {
    print_header "🏆 Competitive Analysis Workflow Testing"
    print_status "======================================"
    
    print_instruction "Step 1: Test Competitive Analysis API Endpoint"
    echo "Test the competitive analysis API endpoint:"
    echo ""
    echo "curl -X POST $BASE_URL/v2/business-intelligence/competitive-analysis \\"
    echo "  -H 'Content-Type: application/json' \\"
    echo "  -d '{"
    echo "    \"business_id\": \"test-business-123\","
    echo "    \"competitors\": [\"competitor1\", \"competitor2\", \"competitor3\"],"
    echo "    \"time_range\": {"
    echo "      \"start_date\": \"2024-01-01T00:00:00Z\","
    echo "      \"end_date\": \"2024-12-31T23:59:59Z\""
    echo "    },"
    echo "    \"options\": {"
    echo "      \"include_market_share\": true,"
    echo "      \"include_pricing\": true,"
    echo "      \"include_features\": true"
    echo "    }"
    echo "  }'"
    echo ""
    
    print_instruction "Step 2: Test Competitive Analysis UI"
    echo "Open your web browser and navigate to:"
    echo "$UI_URL/competitive-analysis-dashboard.html"
    echo ""
    echo "Manual UI Testing Checklist:"
    echo "- [ ] Page loads successfully"
    echo "- [ ] Competitor selection interface works"
    echo "- [ ] Analysis options are configurable"
    echo "- [ ] Results display competitor comparisons"
    echo "- [ ] Charts and graphs render correctly"
    echo "- [ ] Export functionality works (if available)"
    echo ""
    
    print_instruction "Step 3: Test Competitive Analysis Job Workflow"
    echo "Follow the same job creation and status checking pattern as market analysis."
    echo ""
    
    echo ""
    print_instruction "Document your findings:"
    echo "- Record competitor data accuracy"
    echo "- Note analysis depth and insights"
    echo "- Document UI responsiveness and usability"
    echo ""
}

# Function to test growth analytics workflow
test_growth_analytics_workflow() {
    print_header "📈 Growth Analytics Workflow Testing"
    print_status "=================================="
    
    print_instruction "Step 1: Test Growth Analytics API Endpoint"
    echo "Test the growth analytics API endpoint:"
    echo ""
    echo "curl -X POST $BASE_URL/v2/business-intelligence/growth-analytics \\"
    echo "  -H 'Content-Type: application/json' \\"
    echo "  -d '{"
    echo "    \"business_id\": \"test-business-123\","
    echo "    \"time_range\": {"
    echo "      \"start_date\": \"2024-01-01T00:00:00Z\","
    echo "      \"end_date\": \"2024-12-31T23:59:59Z\""
    echo "    },"
    echo "    \"options\": {"
    echo "      \"include_revenue\": true,"
    echo "      \"include_customers\": true,"
    echo "      \"include_metrics\": true"
    echo "    }"
    echo "  }'"
    echo ""
    
    print_instruction "Step 2: Test Growth Analytics UI"
    echo "Open your web browser and navigate to:"
    echo "$UI_URL/business-growth-analytics.html"
    echo ""
    echo "Manual UI Testing Checklist:"
    echo "- [ ] Page loads successfully"
    echo "- [ ] Growth metrics are displayed clearly"
    echo "- [ ] Time series charts render correctly"
    echo "- [ ] Trend analysis is accurate"
    echo "- [ ] Forecasting features work"
    echo "- [ ] Data export functionality works"
    echo ""
    
    print_instruction "Step 3: Test Growth Analytics Job Workflow"
    echo "Follow the same job creation and status checking pattern."
    echo ""
    
    echo ""
    print_instruction "Document your findings:"
    echo "- Record growth metric accuracy"
    echo "- Note trend analysis quality"
    echo "- Document forecasting reliability"
    echo ""
}

# Function to test error handling and edge cases
test_error_handling() {
    print_header "⚠️ Error Handling and Edge Cases Testing"
    print_status "======================================"
    
    print_instruction "Step 1: Test Invalid Input Handling"
    echo "Test with invalid JSON:"
    echo ""
    echo "curl -X POST $BASE_URL/v2/business-intelligence/market-analysis \\"
    echo "  -H 'Content-Type: application/json' \\"
    echo "  -d 'invalid json'"
    echo ""
    echo "Expected: HTTP 400 Bad Request"
    echo ""
    
    print_instruction "Step 2: Test Missing Required Fields"
    echo "Test with missing business_id:"
    echo ""
    echo "curl -X POST $BASE_URL/v2/business-intelligence/market-analysis \\"
    echo "  -H 'Content-Type: application/json' \\"
    echo "  -d '{"
    echo "    \"time_range\": {"
    echo "      \"start_date\": \"2024-01-01T00:00:00Z\","
    echo "      \"end_date\": \"2024-12-31T23:59:59Z\""
    echo "    }"
    echo "  }'"
    echo ""
    echo "Expected: HTTP 400 Bad Request with validation error"
    echo ""
    
    print_instruction "Step 3: Test Invalid Date Ranges"
    echo "Test with invalid date range:"
    echo ""
    echo "curl -X POST $BASE_URL/v2/business-intelligence/market-analysis \\"
    echo "  -H 'Content-Type: application/json' \\"
    echo "  -d '{"
    echo "    \"business_id\": \"test-business-123\","
    echo "    \"time_range\": {"
    echo "      \"start_date\": \"2024-12-31T23:59:59Z\","
    echo "      \"end_date\": \"2024-01-01T00:00:00Z\""
    echo "    }"
    echo "  }'"
    echo ""
    echo "Expected: HTTP 400 Bad Request with date validation error"
    echo ""
    
    print_instruction "Step 4: Test Non-existent Resource Access"
    echo "Test accessing non-existent analysis:"
    echo ""
    echo "curl -X GET $BASE_URL/v2/business-intelligence/market-analysis/non-existent-id"
    echo ""
    echo "Expected: HTTP 404 Not Found"
    echo ""
    
    print_instruction "Step 5: Test Rate Limiting"
    echo "Send multiple rapid requests to test rate limiting:"
    echo ""
    echo "for i in {1..10}; do"
    echo "  curl -X POST $BASE_URL/v2/business-intelligence/market-analysis \\"
    echo "    -H 'Content-Type: application/json' \\"
    echo "    -d '{\"business_id\": \"test-$i\"}' &"
    echo "done"
    echo "wait"
    echo ""
    echo "Expected: Some requests should be rate limited (HTTP 429)"
    echo ""
    
    echo ""
    print_instruction "Document your findings:"
    echo "- Record error message quality and helpfulness"
    echo "- Note error handling consistency across endpoints"
    echo "- Document any security vulnerabilities found"
    echo ""
}

# Function to test UI navigation and usability
test_ui_navigation() {
    print_header "🖥️ User Interface Navigation Testing"
    print_status "=================================="
    
    print_instruction "Step 1: Test Main Dashboard Navigation"
    echo "Open your web browser and navigate to:"
    echo "$UI_URL/dashboard.html"
    echo ""
    echo "Navigation Testing Checklist:"
    echo "- [ ] Main dashboard loads successfully"
    echo "- [ ] Navigation menu is visible and functional"
    echo "- [ ] All dashboard sections are accessible"
    echo "- [ ] Links to business intelligence modules work"
    echo "- [ ] Responsive design works on different screen sizes"
    echo ""
    
    print_instruction "Step 2: Test Business Intelligence Module Navigation"
    echo "Test navigation between BI modules:"
    echo ""
    echo "1. Market Analysis Dashboard: $UI_URL/market-analysis-dashboard.html"
    echo "2. Competitive Analysis Dashboard: $UI_URL/competitive-analysis-dashboard.html"
    echo "3. Growth Analytics Dashboard: $UI_URL/business-growth-analytics.html"
    echo ""
    echo "Navigation Testing Checklist:"
    echo "- [ ] All modules load without errors"
    echo "- [ ] Navigation between modules is smooth"
    echo "- [ ] Breadcrumbs or back navigation works"
    echo "- [ ] Page titles and headers are correct"
    echo "- [ ] Loading states are appropriate"
    echo ""
    
    print_instruction "Step 3: Test Form Interactions"
    echo "Test form functionality across all modules:"
    echo ""
    echo "Form Testing Checklist:"
    echo "- [ ] All input fields accept data correctly"
    echo "- [ ] Date pickers work and validate dates"
    echo "- [ ] Dropdown selections work properly"
    echo "- [ ] Checkboxes and radio buttons function"
    echo "- [ ] Form validation provides clear feedback"
    echo "- [ ] Submit buttons trigger appropriate actions"
    echo "- [ ] Reset/clear functionality works"
    echo ""
    
    print_instruction "Step 4: Test Data Display and Visualization"
    echo "Test how data is presented to users:"
    echo ""
    echo "Data Display Testing Checklist:"
    echo "- [ ] Tables display data clearly and are sortable"
    echo "- [ ] Charts and graphs render correctly"
    echo "- [ ] Data is formatted appropriately (dates, numbers, currency)"
    echo "- [ ] Empty states are handled gracefully"
    echo "- [ ] Loading states are shown during data fetching"
    echo "- [ ] Error states display helpful messages"
    echo ""
    
    print_instruction "Step 5: Test Accessibility"
    echo "Test accessibility features:"
    echo ""
    echo "Accessibility Testing Checklist:"
    echo "- [ ] Page can be navigated using keyboard only"
    echo "- [ ] Screen reader compatibility (if available)"
    echo "- [ ] Color contrast is sufficient"
    echo "- [ ] Text is readable at different zoom levels"
    echo "- [ ] Form labels are properly associated"
    echo "- [ ] Error messages are accessible"
    echo ""
    
    echo ""
    print_instruction "Document your findings:"
    echo "- Record any navigation issues or confusion"
    echo "- Note usability problems or improvements needed"
    echo "- Document accessibility barriers"
    echo "- Record performance issues or slow loading"
    echo ""
}

# Function to generate manual testing report
generate_manual_testing_report() {
    local report_file="$TEST_RESULTS_DIR/manual-workflow-testing-report-$(date +%Y%m%d_%H%M%S).txt"
    
    print_status "Generating manual testing report: $report_file"
    
    cat > "$report_file" << EOF
Manual Workflow Testing Report
=============================
Generated: $(date)
Test Suite: Manual Workflow Testing
Version: 1.0.0

Manual Testing Procedures Executed:
1. Market Analysis Workflow Testing
2. Competitive Analysis Workflow Testing
3. Growth Analytics Workflow Testing
4. Error Handling and Edge Cases Testing
5. User Interface Navigation Testing

Testing Environment:
- API Base URL: $BASE_URL
- UI Base URL: $UI_URL
- Test Date: $(date)

Manual Testing Checklist:
=======================

Market Analysis Workflow:
- [ ] API endpoint responds correctly
- [ ] UI loads and functions properly
- [ ] Job creation and status checking works
- [ ] Results retrieval functions correctly
- [ ] Error handling is appropriate

Competitive Analysis Workflow:
- [ ] API endpoint responds correctly
- [ ] UI loads and functions properly
- [ ] Competitor selection works
- [ ] Analysis results are accurate
- [ ] Export functionality works (if available)

Growth Analytics Workflow:
- [ ] API endpoint responds correctly
- [ ] UI loads and functions properly
- [ ] Growth metrics are accurate
- [ ] Charts and visualizations work
- [ ] Forecasting features function

Error Handling and Edge Cases:
- [ ] Invalid JSON handling
- [ ] Missing required fields handling
- [ ] Invalid date ranges handling
- [ ] Non-existent resource access
- [ ] Rate limiting functionality

User Interface Navigation:
- [ ] Main dashboard navigation
- [ ] Business intelligence module navigation
- [ ] Form interactions
- [ ] Data display and visualization
- [ ] Accessibility features

Issues Found:
============
(To be filled in during manual testing)

Recommendations:
===============
(To be filled in during manual testing)

Overall Assessment:
==================
(To be filled in after completing all manual tests)

Next Steps:
===========
1. Complete all manual testing procedures
2. Document all findings and issues
3. Prioritize issues for resolution
4. Plan improvements based on findings
5. Schedule follow-up testing after fixes
EOF
    
    print_success "Manual testing report generated: $report_file"
    print_info "Please complete the manual testing procedures and update the report with your findings."
}

# Function to display testing summary
display_testing_summary() {
    print_header "📊 Manual Testing Summary"
    print_status "========================"
    
    print_info "Manual testing procedures have been prepared and documented."
    print_info "Please follow the instructions above to complete manual testing."
    echo ""
    
    print_instruction "Testing Steps Completed:"
    echo "✅ Testing instructions displayed"
    echo "✅ API endpoints documented"
    echo "✅ UI testing procedures outlined"
    echo "✅ Error handling scenarios defined"
    echo "✅ Navigation testing checklist provided"
    echo "✅ Testing report template generated"
    echo ""
    
    print_instruction "Next Steps:"
    echo "1. Follow each testing workflow step by step"
    echo "2. Document your findings in the generated report"
    echo "3. Record any issues or observations"
    echo "4. Complete the testing checklist"
    echo "5. Update the report with your assessment"
    echo ""
    
    print_info "Testing report location: $TEST_RESULTS_DIR/manual-workflow-testing-report-*.txt"
}

# Main execution
main() {
    print_header "🧪 Manual Workflow Testing"
    print_header "========================="
    
    # Start servers
    start_servers
    
    # Display instructions
    display_manual_testing_instructions
    
    echo ""
    read -p "Press Enter to continue with testing procedures..."
    echo ""
    
    # Run testing procedures
    test_market_analysis_workflow
    
    echo ""
    read -p "Press Enter to continue with competitive analysis testing..."
    echo ""
    
    test_competitive_analysis_workflow
    
    echo ""
    read -p "Press Enter to continue with growth analytics testing..."
    echo ""
    
    test_growth_analytics_workflow
    
    echo ""
    read -p "Press Enter to continue with error handling testing..."
    echo ""
    
    test_error_handling
    
    echo ""
    read -p "Press Enter to continue with UI navigation testing..."
    echo ""
    
    test_ui_navigation
    
    # Generate report
    generate_manual_testing_report
    
    # Display summary
    display_testing_summary
    
    # Keep servers running for manual testing
    print_info "Servers are still running for your manual testing."
    print_info "API Server: $BASE_URL"
    print_info "UI Server: $UI_URL"
    print_info "Press Ctrl+C to stop servers when testing is complete."
    
    # Wait for user to finish testing
    print_status "Waiting for manual testing to complete..."
    print_status "Press Ctrl+C when finished to stop servers and exit."
    
    # Trap to stop servers on exit
    trap stop_servers EXIT
    
    # Keep script running
    while true; do
        sleep 10
    done
}

# Run main function
main "$@"
