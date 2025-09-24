#!/bin/bash

# Implementation Validation Script
# This script validates that the feature functionality testing implementation is complete and correct

set -e  # Exit on any error

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check file existence and content
check_file() {
    local file_path="$1"
    local description="$2"
    local min_size="$3"
    
    if [[ -f "$file_path" ]]; then
        local file_size=$(wc -c < "$file_path")
        if [[ $file_size -ge $min_size ]]; then
            log_success "$description: $file_path ($file_size bytes)"
            return 0
        else
            log_warning "$description: $file_path (too small: $file_size bytes, expected >= $min_size)"
            return 1
        fi
    else
        log_error "$description: $file_path (not found)"
        return 1
    fi
}

# Function to check Go syntax
check_go_syntax() {
    local file_path="$1"
    local description="$2"
    
    if [[ -f "$file_path" ]]; then
        if go fmt "$file_path" > /dev/null 2>&1; then
            log_success "$description: Go syntax valid"
            return 0
        else
            log_error "$description: Go syntax invalid"
            return 1
        fi
    else
        log_error "$description: File not found"
        return 1
    fi
}

# Function to check YAML syntax
check_yaml_syntax() {
    local file_path="$1"
    local description="$2"
    
    if [[ -f "$file_path" ]]; then
        if command -v yq &> /dev/null; then
            if yq eval '.' "$file_path" > /dev/null 2>&1; then
                log_success "$description: YAML syntax valid"
                return 0
            else
                log_error "$description: YAML syntax invalid"
                return 1
            fi
        else
            log_warning "$description: YAML syntax check skipped (yq not installed)"
            return 0
        fi
    else
        log_error "$description: File not found"
        return 1
    fi
}

# Function to check script permissions
check_script_permissions() {
    local file_path="$1"
    local description="$2"
    
    if [[ -f "$file_path" ]]; then
        if [[ -x "$file_path" ]]; then
            log_success "$description: Executable permissions set"
            return 0
        else
            log_warning "$description: Not executable, setting permissions..."
            chmod +x "$file_path"
            log_success "$description: Executable permissions set"
            return 0
        fi
    else
        log_error "$description: File not found"
        return 1
    fi
}

# Main validation function
validate_implementation() {
    log_info "Validating Feature Functionality Testing Implementation"
    log_info "======================================================"
    
    local errors=0
    local warnings=0
    
    # Check main test files
    log_info "Checking main test files..."
    
    if ! check_file "$SCRIPT_DIR/feature_functionality_test.go" "Main test suite" 1000; then
        ((errors++))
    fi
    
    if ! check_file "$SCRIPT_DIR/business_classification_test.go" "Business classification tests" 2000; then
        ((errors++))
    fi
    
    if ! check_file "$SCRIPT_DIR/risk_assessment_test.go" "Risk assessment tests" 2000; then
        ((errors++))
    fi
    
    if ! check_file "$SCRIPT_DIR/compliance_checking_test.go" "Compliance checking tests" 2000; then
        ((errors++))
    fi
    
    if ! check_file "$SCRIPT_DIR/merchant_management_test.go" "Merchant management tests" 2000; then
        ((errors++))
    fi
    
    if ! check_file "$SCRIPT_DIR/test_runner.go" "Test runner framework" 2000; then
        ((errors++))
    fi
    
    # Check configuration and scripts
    log_info "Checking configuration and scripts..."
    
    if ! check_file "$SCRIPT_DIR/test_config.yaml" "Test configuration" 1000; then
        ((errors++))
    fi
    
    if ! check_file "$SCRIPT_DIR/run_feature_tests.sh" "Test execution script" 1000; then
        ((errors++))
    fi
    
    if ! check_file "$SCRIPT_DIR/README.md" "Test documentation" 2000; then
        ((errors++))
    fi
    
    # Check Go syntax
    log_info "Checking Go syntax..."
    
    if ! check_go_syntax "$SCRIPT_DIR/feature_functionality_test.go" "Main test suite"; then
        ((errors++))
    fi
    
    if ! check_go_syntax "$SCRIPT_DIR/business_classification_test.go" "Business classification tests"; then
        ((errors++))
    fi
    
    if ! check_go_syntax "$SCRIPT_DIR/risk_assessment_test.go" "Risk assessment tests"; then
        ((errors++))
    fi
    
    if ! check_go_syntax "$SCRIPT_DIR/compliance_checking_test.go" "Compliance checking tests"; then
        ((errors++))
    fi
    
    if ! check_go_syntax "$SCRIPT_DIR/merchant_management_test.go" "Merchant management tests"; then
        ((errors++))
    fi
    
    if ! check_go_syntax "$SCRIPT_DIR/test_runner.go" "Test runner framework"; then
        ((errors++))
    fi
    
    # Check YAML syntax
    log_info "Checking YAML syntax..."
    
    if ! check_yaml_syntax "$SCRIPT_DIR/test_config.yaml" "Test configuration"; then
        ((errors++))
    fi
    
    # Check script permissions
    log_info "Checking script permissions..."
    
    if ! check_script_permissions "$SCRIPT_DIR/run_feature_tests.sh" "Test execution script"; then
        ((errors++))
    fi
    
    if ! check_script_permissions "$SCRIPT_DIR/validate_implementation.sh" "Validation script"; then
        ((errors++))
    fi
    
    # Check test structure
    log_info "Checking test structure..."
    
    # Check if test functions exist in main test file
    if grep -q "TestBusinessClassificationFeatures" "$SCRIPT_DIR/feature_functionality_test.go"; then
        log_success "Business classification test function found"
    else
        log_error "Business classification test function not found"
        ((errors++))
    fi
    
    if grep -q "TestRiskAssessmentFeatures" "$SCRIPT_DIR/feature_functionality_test.go"; then
        log_success "Risk assessment test function found"
    else
        log_error "Risk assessment test function not found"
        ((errors++))
    fi
    
    if grep -q "TestComplianceCheckingFeatures" "$SCRIPT_DIR/feature_functionality_test.go"; then
        log_success "Compliance checking test function found"
    else
        log_error "Compliance checking test function not found"
        ((errors++))
    fi
    
    if grep -q "TestMerchantManagementFeatures" "$SCRIPT_DIR/feature_functionality_test.go"; then
        log_success "Merchant management test function found"
    else
        log_error "Merchant management test function not found"
        ((errors++))
    fi
    
    # Check if test methods exist in individual test files
    log_info "Checking individual test methods..."
    
    # Business classification tests
    if grep -q "testMultiMethodClassification" "$SCRIPT_DIR/business_classification_test.go"; then
        log_success "Multi-method classification test found"
    else
        log_error "Multi-method classification test not found"
        ((errors++))
    fi
    
    if grep -q "testKeywordBasedClassification" "$SCRIPT_DIR/business_classification_test.go"; then
        log_success "Keyword-based classification test found"
    else
        log_error "Keyword-based classification test not found"
        ((errors++))
    fi
    
    # Risk assessment tests
    if grep -q "testComprehensiveRiskAssessment" "$SCRIPT_DIR/risk_assessment_test.go"; then
        log_success "Comprehensive risk assessment test found"
    else
        log_error "Comprehensive risk assessment test not found"
        ((errors++))
    fi
    
    if grep -q "testSecurityAnalysis" "$SCRIPT_DIR/risk_assessment_test.go"; then
        log_success "Security analysis test found"
    else
        log_error "Security analysis test not found"
        ((errors++))
    fi
    
    # Compliance checking tests
    if grep -q "testAMLCompliance" "$SCRIPT_DIR/compliance_checking_test.go"; then
        log_success "AML compliance test found"
    else
        log_error "AML compliance test not found"
        ((errors++))
    fi
    
    if grep -q "testKYCCompliance" "$SCRIPT_DIR/compliance_checking_test.go"; then
        log_success "KYC compliance test found"
    else
        log_error "KYC compliance test not found"
        ((errors++))
    fi
    
    # Merchant management tests
    if grep -q "testCreateMerchant" "$SCRIPT_DIR/merchant_management_test.go"; then
        log_success "Create merchant test found"
    else
        log_error "Create merchant test not found"
        ((errors++))
    fi
    
    if grep -q "testGetMerchant" "$SCRIPT_DIR/merchant_management_test.go"; then
        log_success "Get merchant test found"
    else
        log_error "Get merchant test not found"
        ((errors++))
    fi
    
    # Check configuration completeness
    log_info "Checking configuration completeness..."
    
    if grep -q "business_classification:" "$SCRIPT_DIR/test_config.yaml"; then
        log_success "Business classification configuration found"
    else
        log_error "Business classification configuration not found"
        ((errors++))
    fi
    
    if grep -q "risk_assessment:" "$SCRIPT_DIR/test_config.yaml"; then
        log_success "Risk assessment configuration found"
    else
        log_error "Risk assessment configuration not found"
        ((errors++))
    fi
    
    if grep -q "compliance_checking:" "$SCRIPT_DIR/test_config.yaml"; then
        log_success "Compliance checking configuration found"
    else
        log_error "Compliance checking configuration not found"
        ((errors++))
    fi
    
    if grep -q "merchant_management:" "$SCRIPT_DIR/test_config.yaml"; then
        log_success "Merchant management configuration found"
    else
        log_error "Merchant management configuration not found"
        ((errors++))
    fi
    
    # Check script functionality
    log_info "Checking script functionality..."
    
    if grep -q "run_tests()" "$SCRIPT_DIR/run_feature_tests.sh"; then
        log_success "Test execution function found in script"
    else
        log_error "Test execution function not found in script"
        ((errors++))
    fi
    
    if grep -q "run_benchmark_tests()" "$SCRIPT_DIR/run_feature_tests.sh"; then
        log_success "Benchmark test function found in script"
    else
        log_error "Benchmark test function not found in script"
        ((errors++))
    fi
    
    # Summary
    log_info "Validation Summary"
    log_info "================="
    
    if [[ $errors -eq 0 ]]; then
        log_success "Implementation validation PASSED"
        log_success "All required components are present and correctly implemented"
        return 0
    else
        log_error "Implementation validation FAILED"
        log_error "Found $errors errors that need to be fixed"
        if [[ $warnings -gt 0 ]]; then
            log_warning "Found $warnings warnings that should be reviewed"
        fi
        return 1
    fi
}

# Run validation
main() {
    validate_implementation
}

# Run main function
main "$@"
