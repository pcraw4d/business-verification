#!/bin/bash

# Regulatory Accuracy Testing Script
echo "📋 Running Regulatory Accuracy Tests"
echo "==================================="

# Set test environment variables
export TEST_ENV="regulatory_accuracy"
export LOG_LEVEL="info"
export ACCURACY_MODE="regulatory"

# Create validation directory
mkdir -p test/compliance/regulatory_accuracy

# Test 1: Framework Accuracy Validation
echo "🏛️  Testing Framework Accuracy Validation..."
go test -run TestRegulatoryAccuracy/Framework_Accuracy_Validation -v ./test/compliance/regulatory_accuracy_test.go

# Test 2: Requirement Accuracy Validation
echo "📝 Testing Requirement Accuracy Validation..."
go test -run TestRegulatoryAccuracy/Requirement_Accuracy_Validation -v ./test/compliance/regulatory_accuracy_test.go

# Test 3: Compliance Calculation Accuracy
echo "🧮 Testing Compliance Calculation Accuracy..."
go test -run TestRegulatoryAccuracy/Compliance_Calculation_Accuracy -v ./test/compliance/regulatory_accuracy_test.go

# Test 4: Multi-Framework Accuracy Validation
echo "🔄 Testing Multi-Framework Accuracy Validation..."
go test -run TestRegulatoryAccuracy/Multi-Framework_Accuracy_Validation -v ./test/compliance/regulatory_accuracy_test.go

# Test 5: Regulatory Mapping Accuracy
echo "🗺️  Testing Regulatory Mapping Accuracy..."
go test -run TestRegulatoryAccuracy/Regulatory_Mapping_Accuracy -v ./test/compliance/regulatory_accuracy_test.go

# Test 6: Jurisdiction and Scope Accuracy
echo "🌍 Testing Jurisdiction and Scope Accuracy..."
go test -run TestRegulatoryAccuracy/Jurisdiction_and_Scope_Accuracy -v ./test/compliance/regulatory_accuracy_test.go

# Test 7: Authority and Documentation Accuracy
echo "📚 Testing Authority and Documentation Accuracy..."
go test -run TestRegulatoryAccuracy/Authority_and_Documentation_Accuracy -v ./test/compliance/regulatory_accuracy_test.go

# Test 8: Run all regulatory accuracy tests together
echo "🎯 Running All Regulatory Accuracy Tests..."
go test -run TestRegulatoryAccuracy -v ./test/compliance/regulatory_accuracy_test.go

echo "✅ All Regulatory Accuracy Tests Completed!"
echo ""
echo "📊 Regulatory Accuracy Test Summary:"
echo "- Framework Accuracy: Validates framework data accuracy and consistency"
echo "- Requirement Accuracy: Validates requirement data accuracy and consistency"
echo "- Compliance Calculation: Validates compliance calculation accuracy"
echo "- Multi-Framework Accuracy: Validates cross-framework accuracy"
echo "- Regulatory Mapping: Validates framework-requirement mapping accuracy"
echo "- Jurisdiction and Scope: Validates jurisdiction and scope accuracy"
echo "- Authority and Documentation: Validates authority and documentation accuracy"
echo ""
echo "🔍 Check test output above for detailed results"
echo "📁 Validation results saved to: test/compliance/regulatory_accuracy/"
