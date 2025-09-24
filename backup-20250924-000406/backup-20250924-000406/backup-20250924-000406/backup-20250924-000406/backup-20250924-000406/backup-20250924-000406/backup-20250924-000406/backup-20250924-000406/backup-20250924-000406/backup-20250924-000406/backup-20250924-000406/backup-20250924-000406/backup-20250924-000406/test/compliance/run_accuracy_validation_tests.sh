#!/bin/bash

# Compliance Accuracy Validation Testing Script
echo "🎯 Running Compliance Accuracy Validation Tests"
echo "================================================"

# Set test environment variables
export TEST_ENV="compliance_accuracy"
export LOG_LEVEL="info"

# Test 1: Compliance Calculation Accuracy
echo "🧮 Testing Compliance Calculation Accuracy..."
go test -run TestComplianceCalculationAccuracy -v ./test/compliance/accuracy_validation_test.go

# Test 2: Risk Level Calculation Accuracy
echo "⚠️  Testing Risk Level Calculation Accuracy..."
go test -run TestRiskLevelCalculationAccuracy -v ./test/compliance/accuracy_validation_test.go

# Test 3: Velocity Calculation Accuracy
echo "📈 Testing Velocity Calculation Accuracy..."
go test -run TestVelocityCalculationAccuracy -v ./test/compliance/accuracy_validation_test.go

# Test 4: Trend Calculation Accuracy
echo "📊 Testing Trend Calculation Accuracy..."
go test -run TestTrendCalculationAccuracy -v ./test/compliance/accuracy_validation_test.go

# Test 5: Compliance Score Accuracy
echo "🎯 Testing Compliance Score Accuracy..."
go test -run TestComplianceScoreAccuracy -v ./test/compliance/accuracy_validation_test.go

# Test 6: Requirement Status Accuracy
echo "📋 Testing Requirement Status Accuracy..."
go test -run TestRequirementStatusAccuracy -v ./test/compliance/accuracy_validation_test.go

# Test 7: Metrics Calculation Accuracy
echo "📊 Testing Metrics Calculation Accuracy..."
go test -run TestMetricsCalculationAccuracy -v ./test/compliance/accuracy_validation_test.go

# Test 8: Integrated Compliance Accuracy
echo "🔗 Testing Integrated Compliance Accuracy..."
go test -run TestComplianceAccuracyIntegration -v ./test/compliance/accuracy_validation_test.go

# Test 9: Run all accuracy validation tests together
echo "🎯 Running All Accuracy Validation Tests..."
go test -v ./test/compliance/accuracy_validation_test.go

echo "✅ All Compliance Accuracy Validation Tests Completed!"
echo ""
echo "📊 Test Summary:"
echo "- Compliance calculation accuracy: Validates progress and level calculations"
echo "- Risk level calculation accuracy: Validates risk assessment algorithms"
echo "- Velocity calculation accuracy: Validates progress velocity calculations"
echo "- Trend calculation accuracy: Validates trend analysis algorithms"
echo "- Compliance score accuracy: Validates scoring across frameworks"
echo "- Requirement status accuracy: Validates requirement status logic"
echo "- Metrics calculation accuracy: Validates metrics computation"
echo "- Integrated accuracy: Validates end-to-end calculation integration"
echo ""
echo "🔍 Check test output above for detailed results"
