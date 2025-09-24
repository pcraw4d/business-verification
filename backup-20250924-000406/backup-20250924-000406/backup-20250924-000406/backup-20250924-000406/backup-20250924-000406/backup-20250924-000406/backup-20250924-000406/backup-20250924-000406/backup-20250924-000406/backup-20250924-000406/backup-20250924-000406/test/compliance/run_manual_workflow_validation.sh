#!/bin/bash

# Manual Workflow Validation Script
echo "🔍 Running Manual Workflow Validation"
echo "====================================="

# Set test environment variables
export TEST_ENV="manual_validation"
export LOG_LEVEL="info"
export VALIDATION_MODE="manual"

# Create validation directory
mkdir -p test/compliance/manual_validation

# Test 1: Framework Setup Workflow Validation
echo "🏗️  Testing Framework Setup Workflow Validation..."
go test -run TestManualWorkflowValidation/Framework_Setup_Workflow_Validation -v ./test/compliance/manual_workflow_validation.go

# Test 2: Requirement Tracking Workflow Validation
echo "📊 Testing Requirement Tracking Workflow Validation..."
go test -run TestManualWorkflowValidation/Requirement_Tracking_Workflow_Validation -v ./test/compliance/manual_workflow_validation.go

# Test 3: Compliance Assessment Workflow Validation
echo "📋 Testing Compliance Assessment Workflow Validation..."
go test -run TestManualWorkflowValidation/Compliance_Assessment_Workflow_Validation -v ./test/compliance/manual_workflow_validation.go

# Test 4: Multi-Framework Workflow Validation
echo "🔄 Testing Multi-Framework Workflow Validation..."
go test -run TestManualWorkflowValidation/Multi-Framework_Workflow_Validation -v ./test/compliance/manual_workflow_validation.go

# Test 5: Workflow Performance Validation
echo "⚡ Testing Workflow Performance Validation..."
go test -run TestManualWorkflowValidation/Workflow_Performance_Validation -v ./test/compliance/manual_workflow_validation.go

# Test 6: Workflow Error Handling Validation
echo "🛡️  Testing Workflow Error Handling Validation..."
go test -run TestManualWorkflowValidation/Workflow_Error_Handling_Validation -v ./test/compliance/manual_workflow_validation.go

# Test 7: Run all manual workflow validation tests together
echo "🎯 Running All Manual Workflow Validation Tests..."
go test -run TestManualWorkflowValidation -v ./test/compliance/manual_workflow_validation.go

echo "✅ All Manual Workflow Validation Tests Completed!"
echo ""
echo "📊 Manual Workflow Validation Summary:"
echo "- Framework Setup Workflow: Validates framework initialization and setup"
echo "- Requirement Tracking Workflow: Validates requirement tracking and progress updates"
echo "- Compliance Assessment Workflow: Validates compliance assessment and calculation"
echo "- Multi-Framework Workflow: Validates cross-framework integration"
echo "- Workflow Performance: Validates workflow performance and response times"
echo "- Workflow Error Handling: Validates error handling and edge cases"
echo ""
echo "🔍 Check test output above for detailed results"
echo "📁 Validation results saved to: test/compliance/manual_validation/"
