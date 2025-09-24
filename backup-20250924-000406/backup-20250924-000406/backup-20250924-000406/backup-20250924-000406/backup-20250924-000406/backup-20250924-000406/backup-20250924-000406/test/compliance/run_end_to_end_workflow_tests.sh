#!/bin/bash

# End-to-End Compliance Workflow Testing Script
echo "🧪 Running End-to-End Compliance Workflow Tests"
echo "================================================"

# Set test environment variables
export TEST_ENV="compliance_workflow"
export LOG_LEVEL="info"

# Test 1: Complete Compliance Workflow
echo "📋 Testing Complete Compliance Workflow..."
go test -run TestEndToEndComplianceWorkflow -v ./test/compliance/end_to_end_workflow_test.go

# Test 2: Multi-Framework Compliance Workflow
echo "🔄 Testing Multi-Framework Compliance Workflow..."
go test -run TestComplianceWorkflowWithMultipleFrameworks -v ./test/compliance/end_to_end_workflow_test.go

# Test 3: Error Scenario Testing
echo "⚠️  Testing Error Scenarios..."
go test -run TestComplianceWorkflowErrorScenarios -v ./test/compliance/end_to_end_workflow_test.go

# Test 4: Performance Testing
echo "⚡ Testing Workflow Performance..."
go test -run TestComplianceWorkflowPerformance -v ./test/compliance/end_to_end_workflow_test.go

# Test 5: Run all workflow tests together
echo "🎯 Running All Workflow Tests..."
go test -v ./test/compliance/end_to_end_workflow_test.go

echo "✅ All End-to-End Compliance Workflow Tests Completed!"
echo ""
echo "📊 Test Summary:"
echo "- Complete workflow test: Validates full compliance process"
echo "- Multi-framework test: Tests multiple compliance frameworks"
echo "- Error scenario test: Validates error handling"
echo "- Performance test: Ensures workflow performance"
echo ""
echo "🔍 Check test output above for detailed results"
