#!/bin/bash

# Security Scan Script for KYB Platform
# This script runs comprehensive security scans on the codebase

set -e

# Default values
SCAN_DIR="."
OUTPUT_DIR="./security-reports"
SCAN_TYPE="all"
FAIL_ON_CRITICAL="false"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --scan-dir)
            SCAN_DIR="$2"
            shift 2
            ;;
        --output-dir)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        --scan-type)
            SCAN_TYPE="$2"
            shift 2
            ;;
        --fail-on-critical)
            FAIL_ON_CRITICAL="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  --scan-dir DIR        Directory to scan (default: .)"
            echo "  --output-dir DIR      Output directory for reports (default: ./security-reports)"
            echo "  --scan-type TYPE      Type of scan: all, code, dependencies, secrets (default: all)"
            echo "  --fail-on-critical    Fail on critical issues: true/false (default: false)"
            echo "  -h, --help           Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Log file
LOG_FILE="$OUTPUT_DIR/security-scan.log"

# Function to log messages
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Function to run code security scan
run_code_security_scan() {
    log "Running code security scan with gosec..."
    
    if command -v gosec &> /dev/null; then
        gosec -fmt json -out "$OUTPUT_DIR/code-security-scan.json" \
              -severity medium \
              -confidence medium \
              -no-fail \
              "$SCAN_DIR" 2>&1 | tee -a "$LOG_FILE"
        
        # Also generate a text report
        gosec -fmt text -out "$OUTPUT_DIR/code-security-scan.txt" \
              -severity medium \
              -confidence medium \
              -no-fail \
              "$SCAN_DIR" 2>&1 | tee -a "$LOG_FILE"
    else
        log "ERROR: gosec not found. Please install gosec first."
        return 1
    fi
}

# Function to run dependency vulnerability scan
run_dependency_scan() {
    log "Running dependency vulnerability scan with govulncheck..."
    
    if command -v govulncheck &> /dev/null; then
        govulncheck ./... > "$OUTPUT_DIR/dependency-scan.txt" 2>&1 || true
        log "Dependency scan completed. Results saved to $OUTPUT_DIR/dependency-scan.txt"
    else
        log "ERROR: govulncheck not found. Please install govulncheck first."
        return 1
    fi
}

# Function to run secret scanning
run_secret_scan() {
    log "Running secret scan..."
    
    # Basic secret patterns
    SECRET_PATTERNS=(
        "password\s*=\s*['\"][^'\"]{8,}['\"]"
        "secret\s*=\s*['\"][^'\"]{8,}['\"]"
        "api[_-]?key\s*=\s*['\"][^'\"]{8,}['\"]"
        "token\s*=\s*['\"][^'\"]{8,}['\"]"
        "private[_-]?key\s*=\s*['\"][^'\"]{8,}['\"]"
    )
    
    SECRET_FOUND=false
    
    for pattern in "${SECRET_PATTERNS[@]}"; do
        if grep -r -i -E "$pattern" "$SCAN_DIR" \
           --exclude-dir=.git \
           --exclude-dir=vendor \
           --exclude-dir=node_modules \
           --exclude-dir=test \
           --exclude="*_test.go" \
           --exclude="*example*" \
           --exclude="*mock*" \
           --exclude="*.md" \
           --exclude="*.yml" \
           --exclude="*.yaml" \
           --exclude="security-scan.sh" > "$OUTPUT_DIR/secret-scan.txt" 2>/dev/null; then
            SECRET_FOUND=true
        fi
    done
    
    if [ "$SECRET_FOUND" = true ]; then
        log "WARNING: Potential secrets found. Check $OUTPUT_DIR/secret-scan.txt"
    else
        log "No obvious secrets found in code"
        echo "No secrets found" > "$OUTPUT_DIR/secret-scan.txt"
    fi
}

# Function to run Trivy filesystem scan
run_trivy_scan() {
    log "Running Trivy filesystem scan..."
    
    if command -v trivy &> /dev/null; then
        trivy fs --format json --output "$OUTPUT_DIR/trivy-fs-scan.json" \
                 --severity CRITICAL,HIGH,MEDIUM \
                 "$SCAN_DIR" 2>&1 | tee -a "$LOG_FILE"
        
        # Also generate a text report
        trivy fs --format table --output "$OUTPUT_DIR/trivy-fs-scan.txt" \
                 --severity CRITICAL,HIGH,MEDIUM \
                 "$SCAN_DIR" 2>&1 | tee -a "$LOG_FILE"
    else
        log "ERROR: trivy not found. Please install trivy first."
        return 1
    fi
}

# Function to check for critical issues
check_critical_issues() {
    log "Checking for critical issues..."
    
    CRITICAL_ISSUES=0
    
    # Check gosec results
    if [ -f "$OUTPUT_DIR/code-security-scan.json" ]; then
        CRITICAL_COUNT=$(jq '[.Issues[] | select(.severity == "HIGH")] | length' "$OUTPUT_DIR/code-security-scan.json" 2>/dev/null || echo "0")
        CRITICAL_ISSUES=$((CRITICAL_ISSUES + CRITICAL_COUNT))
    fi
    
    # Check Trivy results
    if [ -f "$OUTPUT_DIR/trivy-fs-scan.json" ]; then
        TRIVY_CRITICAL=$(jq '[.Results[]?.Vulnerabilities[]? | select(.Severity == "CRITICAL")] | length' "$OUTPUT_DIR/trivy-fs-scan.json" 2>/dev/null || echo "0")
        CRITICAL_ISSUES=$((CRITICAL_ISSUES + TRIVY_CRITICAL))
    fi
    
    # Check for secrets
    if [ -f "$OUTPUT_DIR/secret-scan.txt" ] && grep -q -v "No secrets found" "$OUTPUT_DIR/secret-scan.txt"; then
        CRITICAL_ISSUES=$((CRITICAL_ISSUES + 1))
    fi
    
    log "Total critical issues found: $CRITICAL_ISSUES"
    
    if [ "$CRITICAL_ISSUES" -gt 0 ] && [ "$FAIL_ON_CRITICAL" = "true" ]; then
        log "ERROR: Critical security issues detected!"
        exit 1
    fi
}

# Main execution
main() {
    log "Starting security scan..."
    log "Scan directory: $SCAN_DIR"
    log "Output directory: $OUTPUT_DIR"
    log "Scan type: $SCAN_TYPE"
    log "Fail on critical: $FAIL_ON_CRITICAL"
    
    case "$SCAN_TYPE" in
        "all")
            run_code_security_scan
            run_dependency_scan
            run_secret_scan
            run_trivy_scan
            ;;
        "code")
            run_code_security_scan
            ;;
        "dependencies")
            run_dependency_scan
            run_trivy_scan
            ;;
        "secrets")
            run_secret_scan
            ;;
        *)
            log "ERROR: Unknown scan type: $SCAN_TYPE"
            exit 1
            ;;
    esac
    
    check_critical_issues
    
    log "Security scan completed successfully!"
    log "Reports saved to: $OUTPUT_DIR"
}

# Run main function
main "$@"