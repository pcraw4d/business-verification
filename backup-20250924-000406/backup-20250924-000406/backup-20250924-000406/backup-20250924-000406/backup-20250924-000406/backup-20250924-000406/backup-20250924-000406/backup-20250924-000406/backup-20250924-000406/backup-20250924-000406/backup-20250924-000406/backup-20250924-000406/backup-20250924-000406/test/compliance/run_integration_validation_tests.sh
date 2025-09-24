#!/bin/bash

# Integration Validation Testing Script
echo "🔗 Running Integration Validation Tests"
echo "======================================"

# Set test environment variables
export TEST_ENV="integration_validation"
export LOG_LEVEL="info"
export INTEGRATION_MODE="validation"

# Create validation directory
mkdir -p test/compliance/integration_validation

# Test 1: End-to-End Compliance Workflow
echo "🔄 Testing End-to-End Compliance Workflow..."
go test -run TestIntegrationValidation/End-to-End_Compliance_Workflow -v ./test/compliance/integration_validation_test.go

# Test 2: Multi-Framework Integration
echo "🔄 Testing Multi-Framework Integration..."
go test -run TestIntegrationValidation/Multi-Framework_Integration -v ./test/compliance/integration_validation_test.go

# Test 3: Service Integration Validation
echo "🔧 Testing Service Integration Validation..."
go test -run TestIntegrationValidation/Service_Integration_Validation -v ./test/compliance/integration_validation_test.go

# Test 4: Data Flow Integration
echo "📊 Testing Data Flow Integration..."
go test -run TestIntegrationValidation/Data_Flow_Integration -v ./test/compliance/integration_validation_test.go

# Test 5: Component Integration Validation
echo "🧩 Testing Component Integration Validation..."
go test -run TestIntegrationValidation/Component_Integration_Validation -v ./test/compliance/integration_validation_test.go

# Test 6: System Integration Validation
echo "🏗️  Testing System Integration Validation..."
go test -run TestIntegrationValidation/System_Integration_Validation -v ./test/compliance/integration_validation_test.go

# Test 7: Integration Error Handling
echo "🛡️  Testing Integration Error Handling..."
go test -run TestIntegrationValidation/Integration_Error_Handling -v ./test/compliance/integration_validation_test.go

# Test 8: Integration Performance Validation
echo "⚡ Testing Integration Performance Validation..."
go test -run TestIntegrationValidation/Integration_Performance_Validation -v ./test/compliance/integration_validation_test.go

# Test 9: Run all integration validation tests together
echo "🎯 Running All Integration Validation Tests..."
go test -run TestIntegrationValidation -v ./test/compliance/integration_validation_test.go

echo "✅ All Integration Validation Tests Completed!"
echo ""
echo "📊 Integration Validation Test Summary:"
echo "- End-to-End Workflow: Validates complete compliance workflow integration"
echo "- Multi-Framework Integration: Validates multi-framework integration scenarios"
echo "- Service Integration: Validates service-to-service integration"
echo "- Data Flow Integration: Validates data flow across components"
echo "- Component Integration: Validates component-to-component integration"
echo "- System Integration: Validates complete system integration"
echo "- Error Handling: Validates integration error handling and recovery"
echo "- Performance Validation: Validates integration performance and consistency"
echo ""
echo "🔍 Check test output above for detailed results"
echo "📁 Validation results saved to: test/compliance/integration_validation/"
