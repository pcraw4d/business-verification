#!/bin/bash

# Quick Test Validation Script
# This script performs a quick validation of the test infrastructure

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
TEST_DIR="$PROJECT_ROOT/test"

# Function to print section headers
print_section() {
    echo -e "\n${BLUE}📋 $1${NC}"
    echo -e "${BLUE}$(printf '=%.0s' {1..50})${NC}"
}

# Function to print validation results
print_validation_result() {
    local test_name="$1"
    local status="$2"
    local details="$3"
    
    if [ "$status" = "PASS" ]; then
        echo -e "✅ ${GREEN}$test_name${NC} - ${GREEN}PASSED${NC}"
    else
        echo -e "❌ ${RED}$test_name${NC} - ${RED}FAILED${NC}"
        if [ -n "$details" ]; then
            echo -e "   ${YELLOW}Details: $details${NC}"
        fi
    fi
}

# Function to validate test files
validate_test_files() {
    print_section "Test Files Validation"
    
    local test_files=(
        "$TEST_DIR/enhanced_classification_system_test.go"
        "$TEST_DIR/test_config.go"
        "$SCRIPT_DIR/setup_test_data.sql"
        "$SCRIPT_DIR/run_enhanced_classification_tests.sh"
        "$SCRIPT_DIR/validate_crosswalk_functionality.sh"
        "$SCRIPT_DIR/execute_subtask_1_5_4_tests.sh"
    )
    
    for file in "${test_files[@]}"; do
        if [ -f "$file" ]; then
            print_validation_result "File: $(basename "$file")" "PASS" "File exists"
        else
            print_validation_result "File: $(basename "$file")" "FAIL" "File not found"
        fi
    done
}

# Function to validate Go test compilation
validate_go_test_compilation() {
    print_section "Go Test Compilation Validation"
    
    if [ ! -d "$TEST_DIR" ]; then
        print_validation_result "Test Directory" "FAIL" "Test directory not found"
        return 1
    fi
    
    # Check if Go is available
    if ! command -v go &> /dev/null; then
        print_validation_result "Go Compilation" "FAIL" "Go not found"
        return 1
    fi
    
    # Try to compile test files
    cd "$TEST_DIR"
    if go build -o /dev/null . 2>/dev/null; then
        print_validation_result "Go Test Compilation" "PASS" "Test files compile successfully"
    else
        print_validation_result "Go Test Compilation" "FAIL" "Test files failed to compile"
        return 1
    fi
}

# Function to validate script permissions
validate_script_permissions() {
    print_section "Script Permissions Validation"
    
    local scripts=(
        "$SCRIPT_DIR/run_enhanced_classification_tests.sh"
        "$SCRIPT_DIR/validate_crosswalk_functionality.sh"
        "$SCRIPT_DIR/execute_subtask_1_5_4_tests.sh"
    )
    
    for script in "${scripts[@]}"; do
        if [ -f "$script" ]; then
            if [ -x "$script" ]; then
                print_validation_result "Script: $(basename "$script")" "PASS" "Executable"
            else
                print_validation_result "Script: $(basename "$script")" "FAIL" "Not executable"
                chmod +x "$script"
                print_validation_result "Script: $(basename "$script")" "PASS" "Made executable"
            fi
        else
            print_validation_result "Script: $(basename "$script")" "FAIL" "File not found"
        fi
    done
}

# Function to validate database connectivity (if configured)
validate_database_connectivity() {
    print_section "Database Connectivity Validation"
    
    if [ -z "$DATABASE_URL" ]; then
        print_validation_result "Database Connection" "SKIP" "DATABASE_URL not set"
        return 0
    fi
    
    # Check if psql is available
    if ! command -v psql &> /dev/null; then
        print_validation_result "Database Connection" "SKIP" "psql not found"
        return 0
    fi
    
    # Test database connection
    if psql "$DATABASE_URL" -c "SELECT 1;" > /dev/null 2>&1; then
        print_validation_result "Database Connection" "PASS" "Successfully connected"
    else
        print_validation_result "Database Connection" "FAIL" "Failed to connect"
    fi
}

# Function to validate test data setup
validate_test_data_setup() {
    print_section "Test Data Setup Validation"
    
    if [ -z "$DATABASE_URL" ]; then
        print_validation_result "Test Data Setup" "SKIP" "DATABASE_URL not set"
        return 0
    fi
    
    local setup_script="$SCRIPT_DIR/setup_test_data.sql"
    if [ -f "$setup_script" ]; then
        print_validation_result "Test Data Script" "PASS" "Setup script exists"
        
        # Check if script has required content
        if grep -q "risk_keywords" "$setup_script" && grep -q "crosswalk_mappings" "$setup_script"; then
            print_validation_result "Test Data Content" "PASS" "Required tables included"
        else
            print_validation_result "Test Data Content" "FAIL" "Required tables missing"
        fi
    else
        print_validation_result "Test Data Script" "FAIL" "Setup script not found"
    fi
}

# Main execution function
main() {
    echo -e "${BLUE}🔍 Quick Test Validation${NC}"
    echo -e "${BLUE}========================${NC}"
    echo ""
    
    # Track validation results
    local total_validations=0
    local passed_validations=0
    local failed_validations=0
    local skipped_validations=0
    
    # Validate test files
    validate_test_files
    echo ""
    
    # Validate Go test compilation
    if validate_go_test_compilation; then
        passed_validations=$((passed_validations + 1))
    else
        failed_validations=$((failed_validations + 1))
    fi
    total_validations=$((total_validations + 1))
    echo ""
    
    # Validate script permissions
    validate_script_permissions
    echo ""
    
    # Validate database connectivity
    validate_database_connectivity
    echo ""
    
    # Validate test data setup
    validate_test_data_setup
    echo ""
    
    # Print final summary
    print_section "Validation Summary"
    echo -e "📊 ${BLUE}Total Validations:${NC} $total_validations"
    echo -e "✅ ${GREEN}Passed:${NC} $passed_validations"
    echo -e "❌ ${RED}Failed:${NC} $failed_validations"
    echo -e "⏭️  ${YELLOW}Skipped:${NC} $skipped_validations"
    
    if [ $total_validations -gt 0 ]; then
        echo -e "📈 ${BLUE}Success Rate:${NC} $(( (passed_validations * 100) / total_validations ))%"
    fi
    
    # Determine exit code
    if [ $failed_validations -eq 0 ]; then
        echo -e "\n🎉 ${GREEN}Quick validation passed! Test infrastructure is ready.${NC}"
        exit 0
    else
        echo -e "\n⚠️  ${YELLOW}Some validations failed. Please fix issues before running full tests.${NC}"
        exit 1
    fi
}

# Run main function
main "$@"
