#!/bin/bash

# User Experience Testing Script
echo "👤 Running Compliance User Experience Tests"
echo "==========================================="

# Set test environment variables
export TEST_ENV="user_experience"
export LOG_LEVEL="info"

# Test 1: Dashboard User Experience
echo "🖥️  Testing Dashboard User Experience..."
go test -run TestComplianceDashboardUserExperience -v ./test/compliance/user_experience_test.go

# Test 2: Workflow User Experience
echo "🔄 Testing Workflow User Experience..."
go test -run TestComplianceWorkflowUserExperience -v ./test/compliance/user_experience_test.go

# Test 3: Dashboard Accessibility
echo "♿ Testing Dashboard Accessibility..."
go test -run TestComplianceDashboardAccessibility -v ./test/compliance/user_experience_test.go

# Test 4: Dashboard Performance
echo "⚡ Testing Dashboard Performance..."
go test -run TestComplianceDashboardPerformance -v ./test/compliance/user_experience_test.go

# Test 5: Run all user experience tests together
echo "🎯 Running All User Experience Tests..."
go test -v ./test/compliance/user_experience_test.go

echo "✅ All User Experience Tests Completed!"
echo ""
echo "📊 Test Summary:"
echo "- Dashboard data loading: Validates data loading performance and user experience"
echo "- Dashboard navigation: Validates navigation performance and user flow"
echo "- Dashboard responsiveness: Validates real-time updates and responsiveness"
echo "- Dashboard error handling: Validates error handling and user feedback"
echo "- Workflow initialization: Validates workflow setup and initialization"
echo "- Workflow progress: Validates progress tracking and user experience"
echo "- Workflow completion: Validates completion and finalization experience"
echo "- Workflow error recovery: Validates error recovery and user guidance"
echo "- Framework accessibility: Validates framework accessibility and availability"
echo "- Requirement accessibility: Validates requirement accessibility and availability"
echo "- Data consistency accessibility: Validates data consistency and accessibility"
echo "- Dashboard load performance: Validates dashboard load performance"
echo "- Dashboard update performance: Validates dashboard update performance"
echo "- Dashboard query performance: Validates dashboard query performance"
echo ""
echo "🔍 Check test output above for detailed results"
