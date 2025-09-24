#!/bin/bash

# User Acceptance Testing Script
echo "👥 Running User Acceptance Tests"
echo "==============================="

# Set test environment variables
export TEST_ENV="user_acceptance"
export LOG_LEVEL="info"
export USER_MODE="acceptance"

# Create validation directory
mkdir -p test/compliance/user_acceptance

# Test 1: User Dashboard Access
echo "🏠 Testing User Dashboard Access..."
go test -run TestUserAcceptance/User_Dashboard_Access -v ./test/compliance/user_acceptance_test.go

# Test 2: User Compliance Tracking
echo "📊 Testing User Compliance Tracking..."
go test -run TestUserAcceptance/User_Compliance_Tracking -v ./test/compliance/user_acceptance_test.go

# Test 3: User Multi-Framework Management
echo "🔄 Testing User Multi-Framework Management..."
go test -run TestUserAcceptance/User_Multi-Framework_Management -v ./test/compliance/user_acceptance_test.go

# Test 4: User Requirement Management
echo "📝 Testing User Requirement Management..."
go test -run TestUserAcceptance/User_Requirement_Management -v ./test/compliance/user_acceptance_test.go

# Test 5: User Compliance Reporting
echo "📋 Testing User Compliance Reporting..."
go test -run TestUserAcceptance/User_Compliance_Reporting -v ./test/compliance/user_acceptance_test.go

# Test 6: User Error Handling
echo "🛡️  Testing User Error Handling..."
go test -run TestUserAcceptance/User_Error_Handling -v ./test/compliance/user_acceptance_test.go

# Test 7: User Performance Expectations
echo "⚡ Testing User Performance Expectations..."
go test -run TestUserAcceptance/User_Performance_Expectations -v ./test/compliance/user_acceptance_test.go

# Test 8: User Workflow Completion
echo "✅ Testing User Workflow Completion..."
go test -run TestUserAcceptance/User_Workflow_Completion -v ./test/compliance/user_acceptance_test.go

# Test 9: Run all user acceptance tests together
echo "🎯 Running All User Acceptance Tests..."
go test -run TestUserAcceptance -v ./test/compliance/user_acceptance_test.go

echo "✅ All User Acceptance Tests Completed!"
echo ""
echo "📊 User Acceptance Test Summary:"
echo "- Dashboard Access: Validates user access to compliance dashboard and frameworks"
echo "- Compliance Tracking: Validates user compliance tracking and progress updates"
echo "- Multi-Framework Management: Validates user management of multiple frameworks"
echo "- Requirement Management: Validates user requirement tracking and updates"
echo "- Compliance Reporting: Validates user compliance reporting and analysis"
echo "- Error Handling: Validates user error handling and edge cases"
echo "- Performance Expectations: Validates user performance expectations"
echo "- Workflow Completion: Validates complete user workflow from start to finish"
echo ""
echo "🔍 Check test output above for detailed results"
echo "📁 Validation results saved to: test/compliance/user_acceptance/"
