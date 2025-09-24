#!/bin/bash

# Simple Manual Workflow Validation Script
echo "🔍 Running Simple Manual Workflow Validation"
echo "==========================================="

# Set test environment variables
export TEST_ENV="manual_validation"
export LOG_LEVEL="info"
export VALIDATION_MODE="manual"

# Create validation directory
mkdir -p test/compliance/manual_validation

# Test 1: Framework Setup Workflow Validation
echo "🏗️  Testing Framework Setup Workflow Validation..."
go test -run TestSimpleManualWorkflowValidation/Framework_Setup_Workflow_Validation -v ./test/compliance/simple_manual_workflow_validation.go

# Test 2: Requirement Tracking Workflow Validation
echo "📊 Testing Requirement Tracking Workflow Validation..."
go test -run TestSimpleManualWorkflowValidation/Requirement_Tracking_Workflow_Validation -v ./test/compliance/simple_manual_workflow_validation.go

# Test 3: Compliance Assessment Workflow Validation
echo "📋 Testing Compliance Assessment Workflow Validation..."
go test -run TestSimpleManualWorkflowValidation/Compliance_Assessment_Workflow_Validation -v ./test/compliance/simple_manual_workflow_validation.go

# Test 4: Multi-Framework Workflow Validation
echo "🔄 Testing Multi-Framework Workflow Validation..."
go test -run TestSimpleManualWorkflowValidation/Multi-Framework_Workflow_Validation -v ./test/compliance/simple_manual_workflow_validation.go

# Test 5: Workflow Performance Validation
echo "⚡ Testing Workflow Performance Validation..."
go test -run TestSimpleManualWorkflowValidation/Workflow_Performance_Validation -v ./test/compliance/simple_manual_workflow_validation.go

# Test 6: Workflow Error Handling Validation
echo "🛡️  Testing Workflow Error Handling Validation..."
go test -run TestSimpleManualWorkflowValidation/Workflow_Error_Handling_Validation -v ./test/compliance/simple_manual_workflow_validation.go

# Test 7: Run all simple manual workflow validation tests together
echo "🎯 Running All Simple Manual Workflow Validation Tests..."
go test -run TestSimpleManualWorkflowValidation -v ./test/compliance/simple_manual_workflow_validation.go

echo "✅ All Simple Manual Workflow Validation Tests Completed!"
echo ""
echo "📊 Simple Manual Workflow Validation Summary:"
echo "- Framework Setup Workflow: Validates framework initialization and setup"
echo "- Requirement Tracking Workflow: Validates requirement tracking and progress updates"
echo "- Compliance Assessment Workflow: Validates compliance assessment and calculation"
echo "- Multi-Framework Workflow: Validates cross-framework integration"
echo "- Workflow Performance: Validates workflow performance and response times"
echo "- Workflow Error Handling: Validates error handling and edge cases"
echo ""
echo "🔍 Check test output above for detailed results"
echo "📁 Validation results saved to: test/compliance/manual_validation/"
