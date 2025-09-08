#!/bin/bash

# Test Security Scan Script for KYB Platform
# This script tests the security scanning setup and tools

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    
    case $status in
        "SUCCESS")
            echo -e "${GREEN}✅ $message${NC}"
            ;;
        "ERROR")
            echo -e "${RED}❌ $message${NC}"
            ;;
        "WARNING")
            echo -e "${YELLOW}⚠️  $message${NC}"
            ;;
        "INFO")
            echo -e "${BLUE}ℹ️  $message${NC}"
            ;;
    esac
}

# Function to test tool availability
test_tool() {
    local tool_name=$1
    local tool_command=$2
    local install_command=$3
    
    print_status "INFO" "Testing $tool_name..."
    
    if command -v "$tool_command" &> /dev/null; then
        print_status "SUCCESS" "$tool_name is available"
        
        # Test basic functionality
        case $tool_name in
            "gosec")
                if gosec --version &> /dev/null; then
                    print_status "SUCCESS" "$tool_name version check passed"
                else
                    print_status "WARNING" "$tool_name version check failed"
                fi
                ;;
            "govulncheck")
                if govulncheck --version &> /dev/null; then
                    print_status "SUCCESS" "$tool_name version check passed"
                else
                    print_status "WARNING" "$tool_name version check failed"
                fi
                ;;
            "trivy")
                if trivy --version &> /dev/null; then
                    print_status "SUCCESS" "$tool_name version check passed"
                else
                    print_status "WARNING" "$tool_name version check failed"
                fi
                ;;
            "semgrep")
                if semgrep --version &> /dev/null; then
                    print_status "SUCCESS" "$tool_name version check passed"
                else
                    print_status "WARNING" "$tool_name version check failed"
                fi
                ;;
        esac
    else
        print_status "ERROR" "$tool_name is not available"
        if [ -n "$install_command" ]; then
            print_status "INFO" "Install command: $install_command"
        fi
        return 1
    fi
}

# Function to test security scan script
test_security_script() {
    print_status "INFO" "Testing enhanced security scan script..."
    
    if [ -f "scripts/enhanced-security-scan.sh" ]; then
        if [ -x "scripts/enhanced-security-scan.sh" ]; then
            print_status "SUCCESS" "Enhanced security scan script is executable"
            
            # Test help functionality
            if ./scripts/enhanced-security-scan.sh --help &> /dev/null; then
                print_status "SUCCESS" "Enhanced security scan script help works"
            else
                print_status "WARNING" "Enhanced security scan script help failed"
            fi
        else
            print_status "ERROR" "Enhanced security scan script is not executable"
            print_status "INFO" "Run: chmod +x scripts/enhanced-security-scan.sh"
        fi
    else
        print_status "ERROR" "Enhanced security scan script not found"
    fi
}

# Function to test basic security scan
test_basic_scan() {
    print_status "INFO" "Running basic security scan test..."
    
    # Create a temporary directory for test output
    local test_dir="/tmp/security-scan-test-$$"
    mkdir -p "$test_dir"
    
    # Run a basic scan on the current directory
    if ./scripts/enhanced-security-scan.sh \
        --scan-dir . \
        --output-dir "$test_dir" \
        --scan-type code \
        --fail-on-critical false \
        --enable-gosec true \
        --enable-semgrep true \
        --enable-trivy false \
        --enable-govulncheck false \
        --enable-secret-scan false \
        --enable-zap false 2>/dev/null; then
        
        print_status "SUCCESS" "Basic security scan completed"
        
        # Check if output files were created
        if [ -f "$test_dir/security-scan.log" ]; then
            print_status "SUCCESS" "Security scan log created"
        else
            print_status "WARNING" "Security scan log not found"
        fi
        
        if [ -f "$test_dir/security-summary.txt" ]; then
            print_status "SUCCESS" "Security summary created"
            print_status "INFO" "Summary content:"
            cat "$test_dir/security-summary.txt"
        else
            print_status "WARNING" "Security summary not found"
        fi
    else
        print_status "ERROR" "Basic security scan failed"
    fi
    
    # Cleanup
    rm -rf "$test_dir"
}

# Function to test GitHub Actions workflow
test_workflow() {
    print_status "INFO" "Testing GitHub Actions workflow files..."
    
    local workflow_files=(
        ".github/workflows/security-scan.yml"
        ".github/workflows/ci-cd.yml"
    )
    
    for workflow_file in "${workflow_files[@]}"; do
        if [ -f "$workflow_file" ]; then
            print_status "SUCCESS" "Found $workflow_file"
            
            # Basic YAML syntax check
            if command -v yamllint &> /dev/null; then
                if yamllint "$workflow_file" &> /dev/null; then
                    print_status "SUCCESS" "$workflow_file YAML syntax is valid"
                else
                    print_status "WARNING" "$workflow_file YAML syntax issues detected"
                fi
            else
                print_status "INFO" "yamllint not available, skipping YAML validation"
            fi
        else
            print_status "ERROR" "$workflow_file not found"
        fi
    done
}

# Function to test configuration files
test_config() {
    print_status "INFO" "Testing configuration files..."
    
    if [ -f "configs/security-scan-config.yaml" ]; then
        print_status "SUCCESS" "Security scan configuration found"
        
        # Basic YAML syntax check
        if command -v yamllint &> /dev/null; then
            if yamllint "configs/security-scan-config.yaml" &> /dev/null; then
                print_status "SUCCESS" "Security scan configuration YAML syntax is valid"
            else
                print_status "WARNING" "Security scan configuration YAML syntax issues detected"
            fi
        else
            print_status "INFO" "yamllint not available, skipping YAML validation"
        fi
    else
        print_status "ERROR" "Security scan configuration not found"
    fi
}

# Function to test documentation
test_documentation() {
    print_status "INFO" "Testing documentation..."
    
    if [ -f "docs/SECURITY_SCANNING_GUIDE.md" ]; then
        print_status "SUCCESS" "Security scanning guide found"
        
        # Check if documentation has required sections
        local required_sections=(
            "Overview"
            "Available Security Tools"
            "Setup and Installation"
            "Configuration"
            "Running Security Scans"
        )
        
        for section in "${required_sections[@]}"; do
            if grep -q "## $section" "docs/SECURITY_SCANNING_GUIDE.md"; then
                print_status "SUCCESS" "Documentation section '$section' found"
            else
                print_status "WARNING" "Documentation section '$section' not found"
            fi
        done
    else
        print_status "ERROR" "Security scanning guide not found"
    fi
}

# Main test function
main() {
    echo "🔍 KYB Platform Security Scan Test Suite"
    echo "========================================"
    echo ""
    
    # Test tool availability
    echo "📋 Testing Security Tools Availability"
    echo "--------------------------------------"
    test_tool "gosec" "gosec" "go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"
    test_tool "govulncheck" "govulncheck" "go install golang.org/x/vuln/cmd/govulncheck@latest"
    test_tool "trivy" "trivy" "curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin v0.50.0"
    test_tool "semgrep" "semgrep" "python3 -m pip install semgrep"
    echo ""
    
    # Test security scan script
    echo "📋 Testing Security Scan Script"
    echo "-------------------------------"
    test_security_script
    echo ""
    
    # Test basic security scan
    echo "📋 Testing Basic Security Scan"
    echo "-----------------------------"
    test_basic_scan
    echo ""
    
    # Test GitHub Actions workflow
    echo "📋 Testing GitHub Actions Workflow"
    echo "----------------------------------"
    test_workflow
    echo ""
    
    # Test configuration
    echo "📋 Testing Configuration Files"
    echo "-----------------------------"
    test_config
    echo ""
    
    # Test documentation
    echo "📋 Testing Documentation"
    echo "-----------------------"
    test_documentation
    echo ""
    
    echo "🎉 Security scan test suite completed!"
    echo ""
    echo "📝 Next Steps:"
    echo "1. Install any missing security tools"
    echo "2. Run a full security scan: ./scripts/enhanced-security-scan.sh"
    echo "3. Review security scan results"
    echo "4. Configure security scanning for your specific needs"
    echo "5. Set up pre-commit hooks for local development"
}

# Run main function
main "$@"
