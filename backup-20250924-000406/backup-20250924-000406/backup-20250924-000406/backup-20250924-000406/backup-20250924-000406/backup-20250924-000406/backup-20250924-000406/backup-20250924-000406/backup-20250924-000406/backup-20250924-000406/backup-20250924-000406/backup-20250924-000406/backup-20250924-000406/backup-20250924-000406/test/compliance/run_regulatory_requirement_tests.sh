#!/bin/bash

# Regulatory Requirement Testing Script
echo "📋 Running Regulatory Requirement Tests"
echo "======================================="

# Set test environment variables
export TEST_ENV="regulatory_requirements"
export LOG_LEVEL="info"

# Test 1: Regulatory Requirement Validation
echo "🔍 Testing Regulatory Requirement Validation..."
go test -run TestRegulatoryRequirementValidation -v ./test/compliance/regulatory_requirement_test.go

# Test 2: Regulatory Requirement Tracking
echo "📊 Testing Regulatory Requirement Tracking..."
go test -run TestRegulatoryRequirementTracking -v ./test/compliance/regulatory_requirement_test.go

# Test 3: Regulatory Requirement Validation (Detailed)
echo "✅ Testing Detailed Regulatory Requirement Validation..."
go test -run TestRegulatoryRequirementValidation -v ./test/compliance/regulatory_requirement_test.go

# Test 4: Regulatory Requirement Integration
echo "🔗 Testing Regulatory Requirement Integration..."
go test -run TestRegulatoryRequirementIntegration -v ./test/compliance/regulatory_requirement_test.go

# Test 5: Run all regulatory requirement tests together
echo "🎯 Running All Regulatory Requirement Tests..."
go test -v ./test/compliance/regulatory_requirement_test.go

echo "✅ All Regulatory Requirement Tests Completed!"
echo ""
echo "📊 Test Summary:"
echo "- Framework validation: Validates compliance framework structure and metadata"
echo "- Requirement validation: Validates individual requirement properties and relationships"
echo "- Requirement tracking: Validates progress tracking and status management"
echo "- Requirement validation (detailed): Validates priority, category, type, and assessment methods"
echo "- Requirement integration: Validates multi-framework integration and cross-references"
echo ""
echo "🔍 Check test output above for detailed results"
