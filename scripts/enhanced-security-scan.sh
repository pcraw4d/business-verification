#!/bin/bash

# Enhanced Security Scan Script for KYB Platform
# This script runs comprehensive security scans using free and open-source tools

set -e

# Default values
SCAN_DIR="."
OUTPUT_DIR="./security-reports"
SCAN_TYPE="all"
FAIL_ON_CRITICAL="false"
ENABLE_SEMGREP="true"
ENABLE_TRIVY="true"
ENABLE_GOSEC="true"
ENABLE_GOVULNCHECK="true"
ENABLE_SECRET_SCAN="true"
ENABLE_ZAP="false"  # Disabled by default as it requires a running application

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
        --enable-semgrep)
            ENABLE_SEMGREP="$2"
            shift 2
            ;;
        --enable-trivy)
            ENABLE_TRIVY="$2"
            shift 2
            ;;
        --enable-gosec)
            ENABLE_GOSEC="$2"
            shift 2
            ;;
        --enable-govulncheck)
            ENABLE_GOVULNCHECK="$2"
            shift 2
            ;;
        --enable-secret-scan)
            ENABLE_SECRET_SCAN="$2"
            shift 2
            ;;
        --enable-zap)
            ENABLE_ZAP="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  --scan-dir DIR        Directory to scan (default: .)"
            echo "  --output-dir DIR      Output directory for reports (default: ./security-reports)"
            echo "  --scan-type TYPE      Type of scan: all, code, dependencies, secrets, containers (default: all)"
            echo "  --fail-on-critical    Fail on critical issues: true/false (default: false)"
            echo "  --enable-semgrep      Enable Semgrep: true/false (default: true)"
            echo "  --enable-trivy        Enable Trivy: true/false (default: true)"
            echo "  --enable-gosec        Enable gosec: true/false (default: true)"
            echo "  --enable-govulncheck  Enable govulncheck: true/false (default: true)"
            echo "  --enable-secret-scan  Enable secret scanning: true/false (default: true)"
            echo "  --enable-zap          Enable OWASP ZAP: true/false (default: false)"
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

# Function to check if tool is available
check_tool() {
    local tool_name="$1"
    local tool_command="$2"
    
    if command -v "$tool_command" &> /dev/null; then
        log "✅ $tool_name is available"
        return 0
    else
        log "❌ $tool_name is not available"
        return 1
    fi
}

# Function to run gosec security scan
run_gosec_scan() {
    if [ "$ENABLE_GOSEC" != "true" ]; then
        log "⏭️ Skipping gosec scan (disabled)"
        return 0
    fi
    
    log "🔍 Running gosec security scan..."
    
    if check_tool "gosec" "gosec"; then
        # Run gosec with comprehensive rules
        gosec -fmt json -out "$OUTPUT_DIR/gosec-results.json" \
              -severity medium \
              -confidence medium \
              -no-fail \
              -rules all \
              "$SCAN_DIR" 2>&1 | tee -a "$LOG_FILE"
        
        # Also generate a text report
        gosec -fmt text -out "$OUTPUT_DIR/gosec-results.txt" \
              -severity medium \
              -confidence medium \
              -no-fail \
              -rules all \
              "$SCAN_DIR" 2>&1 | tee -a "$LOG_FILE"
        
        log "✅ gosec scan completed"
    else
        log "❌ gosec not available, skipping scan"
        return 1
    fi
}

# Function to run govulncheck
run_govulncheck_scan() {
    if [ "$ENABLE_GOVULNCHECK" != "true" ]; then
        log "⏭️ Skipping govulncheck scan (disabled)"
        return 0
    fi
    
    log "🔍 Running govulncheck vulnerability scan..."
    
    if check_tool "govulncheck" "govulncheck"; then
        govulncheck ./... > "$OUTPUT_DIR/govulncheck-results.txt" 2>&1 || true
        log "✅ govulncheck scan completed"
    else
        log "❌ govulncheck not available, skipping scan"
        return 1
    fi
}

# Function to run Trivy filesystem scan
run_trivy_scan() {
    if [ "$ENABLE_TRIVY" != "true" ]; then
        log "⏭️ Skipping Trivy scan (disabled)"
        return 0
    fi
    
    log "🔍 Running Trivy filesystem scan..."
    
    if check_tool "Trivy" "trivy"; then
        # Run Trivy filesystem scan
        trivy fs --format json --output "$OUTPUT_DIR/trivy-fs-results.json" \
                 --severity CRITICAL,HIGH,MEDIUM \
                 --security-checks vuln,secret,config \
                 "$SCAN_DIR" 2>&1 | tee -a "$LOG_FILE"
        
        # Also generate a text report
        trivy fs --format table --output "$OUTPUT_DIR/trivy-fs-results.txt" \
                 --severity CRITICAL,HIGH,MEDIUM \
                 --security-checks vuln,secret,config \
                 "$SCAN_DIR" 2>&1 | tee -a "$LOG_FILE"
        
        log "✅ Trivy filesystem scan completed"
    else
        log "❌ Trivy not available, skipping scan"
        return 1
    fi
}

# Function to run Semgrep static analysis
run_semgrep_scan() {
    if [ "$ENABLE_SEMGREP" != "true" ]; then
        log "⏭️ Skipping Semgrep scan (disabled)"
        return 0
    fi
    
    log "🔍 Running Semgrep static analysis..."
    
    if check_tool "Semgrep" "semgrep"; then
        # Run Semgrep with auto-config (includes security rules)
        semgrep --config=auto \
                --json \
                --output="$OUTPUT_DIR/semgrep-results.json" \
                --severity=ERROR \
                --severity=WARNING \
                --severity=INFO \
                "$SCAN_DIR" 2>&1 | tee -a "$LOG_FILE"
        
        # Also generate a text report
        semgrep --config=auto \
                --output="$OUTPUT_DIR/semgrep-results.txt" \
                --severity=ERROR \
                --severity=WARNING \
                --severity=INFO \
                "$SCAN_DIR" 2>&1 | tee -a "$LOG_FILE"
        
        log "✅ Semgrep scan completed"
    else
        log "❌ Semgrep not available, skipping scan"
        return 1
    fi
}

# Function to run secret scanning
run_secret_scan() {
    if [ "$ENABLE_SECRET_SCAN" != "true" ]; then
        log "⏭️ Skipping secret scan (disabled)"
        return 0
    fi
    
    log "🔍 Running secret scan..."
    
    # Enhanced secret patterns
    SECRET_PATTERNS=(
        "password\s*=\s*['\"][^'\"]{8,}['\"]"
        "secret\s*=\s*['\"][^'\"]{8,}['\"]"
        "api[_-]?key\s*=\s*['\"][^'\"]{8,}['\"]"
        "token\s*=\s*['\"][^'\"]{8,}['\"]"
        "private[_-]?key\s*=\s*['\"][^'\"]{8,}['\"]"
        "access[_-]?key\s*=\s*['\"][^'\"]{8,}['\"]"
        "auth[_-]?token\s*=\s*['\"][^'\"]{8,}['\"]"
        "jwt[_-]?secret\s*=\s*['\"][^'\"]{8,}['\"]"
        "database[_-]?password\s*=\s*['\"][^'\"]{8,}['\"]"
        "redis[_-]?password\s*=\s*['\"][^'\"]{8,}['\"]"
    )
    
    SECRET_FOUND=false
    SECRET_COUNT=0
    
    echo "=== Secret Scan Results ===" > "$OUTPUT_DIR/secret-scan-results.txt"
    echo "Scan started: $(date)" >> "$OUTPUT_DIR/secret-scan-results.txt"
    echo "" >> "$OUTPUT_DIR/secret-scan-results.txt"
    
    for pattern in "${SECRET_PATTERNS[@]}"; do
        if grep -r -i -E "$pattern" "$SCAN_DIR" \
           --exclude-dir=.git \
           --exclude-dir=vendor \
           --exclude-dir=node_modules \
           --exclude-dir=test \
           --exclude-dir=.github \
           --exclude="*_test.go" \
           --exclude="*example*" \
           --exclude="*mock*" \
           --exclude="*.md" \
           --exclude="*.yml" \
           --exclude="*.yaml" \
           --exclude="security-scan.sh" \
           --exclude="enhanced-security-scan.sh" \
           --exclude="*.log" \
           --exclude="*.csv" \
           --exclude="*.html" \
           --exclude="*.json" > "$OUTPUT_DIR/temp-secrets.txt" 2>/dev/null; then
            SECRET_FOUND=true
            SECRET_COUNT=$((SECRET_COUNT + $(wc -l < "$OUTPUT_DIR/temp-secrets.txt")))
            echo "Pattern: $pattern" >> "$OUTPUT_DIR/secret-scan-results.txt"
            cat "$OUTPUT_DIR/temp-secrets.txt" >> "$OUTPUT_DIR/secret-scan-results.txt"
            echo "" >> "$OUTPUT_DIR/secret-scan-results.txt"
        fi
    done
    
    rm -f "$OUTPUT_DIR/temp-secrets.txt"
    
    if [ "$SECRET_FOUND" = true ]; then
        log "⚠️ Found $SECRET_COUNT potential secrets. Check $OUTPUT_DIR/secret-scan-results.txt"
    else
        log "✅ No obvious secrets found in code"
        echo "No secrets found" >> "$OUTPUT_DIR/secret-scan-results.txt"
    fi
}

# Function to run OWASP ZAP baseline scan (if enabled)
run_zap_scan() {
    if [ "$ENABLE_ZAP" != "true" ]; then
        log "⏭️ Skipping OWASP ZAP scan (disabled)"
        return 0
    fi
    
    log "🔍 Running OWASP ZAP baseline scan..."
    
    if [ -f "/opt/zap_2.14.0/zap.sh" ]; then
        # Note: This requires a running application to scan
        # For now, we'll just log that it's available
        log "✅ OWASP ZAP is available at /opt/zap_2.14.0/zap.sh"
        log "ℹ️ To use ZAP, start your application and run:"
        log "   /opt/zap_2.14.0/zap.sh -cmd -quickurl http://localhost:8080 -quickprogress -quickout $OUTPUT_DIR/zap-results.json"
    else
        log "❌ OWASP ZAP not available, skipping scan"
        return 1
    fi
}

# Function to run container security scan (if Dockerfile exists)
run_container_scan() {
    if [ ! -f "Dockerfile" ] && [ ! -f "Dockerfile.enhanced" ]; then
        log "⏭️ No Dockerfile found, skipping container scan"
        return 0
    fi
    
    log "🔍 Running container security scan..."
    
    if check_tool "Trivy" "trivy"; then
        # Build a temporary image for scanning
        local image_name="kyb-platform-security-scan"
        
        log "Building temporary image for scanning..."
        docker build -t "$image_name" . 2>&1 | tee -a "$LOG_FILE"
        
        # Scan the container image
        trivy image --format json --output "$OUTPUT_DIR/trivy-container-results.json" \
                    --severity CRITICAL,HIGH,MEDIUM \
                    --security-checks vuln,secret,config \
                    "$image_name" 2>&1 | tee -a "$LOG_FILE"
        
        # Also generate a text report
        trivy image --format table --output "$OUTPUT_DIR/trivy-container-results.txt" \
                    --severity CRITICAL,HIGH,MEDIUM \
                    --security-checks vuln,secret,config \
                    "$image_name" 2>&1 | tee -a "$LOG_FILE"
        
        # Clean up temporary image
        docker rmi "$image_name" 2>/dev/null || true
        
        log "✅ Container security scan completed"
    else
        log "❌ Trivy not available, skipping container scan"
        return 1
    fi
}

# Function to check for critical issues
check_critical_issues() {
    log "🔍 Checking for critical issues..."
    
    local CRITICAL_ISSUES=0
    local HIGH_ISSUES=0
    local MEDIUM_ISSUES=0
    
    # Check gosec results
    if [ -f "$OUTPUT_DIR/gosec-results.json" ]; then
        local gosec_high=$(jq '[.Issues[] | select(.severity == "HIGH")] | length' "$OUTPUT_DIR/gosec-results.json" 2>/dev/null || echo "0")
        local gosec_medium=$(jq '[.Issues[] | select(.severity == "MEDIUM")] | length' "$OUTPUT_DIR/gosec-results.json" 2>/dev/null || echo "0")
        CRITICAL_ISSUES=$((CRITICAL_ISSUES + gosec_high))
        HIGH_ISSUES=$((HIGH_ISSUES + gosec_high))
        MEDIUM_ISSUES=$((MEDIUM_ISSUES + gosec_medium))
    fi
    
    # Check Trivy filesystem results
    if [ -f "$OUTPUT_DIR/trivy-fs-results.json" ]; then
        local trivy_critical=$(jq '[.Results[]?.Vulnerabilities[]? | select(.Severity == "CRITICAL")] | length' "$OUTPUT_DIR/trivy-fs-results.json" 2>/dev/null || echo "0")
        local trivy_high=$(jq '[.Results[]?.Vulnerabilities[]? | select(.Severity == "HIGH")] | length' "$OUTPUT_DIR/trivy-fs-results.json" 2>/dev/null || echo "0")
        local trivy_medium=$(jq '[.Results[]?.Vulnerabilities[]? | select(.Severity == "MEDIUM")] | length' "$OUTPUT_DIR/trivy-fs-results.json" 2>/dev/null || echo "0")
        CRITICAL_ISSUES=$((CRITICAL_ISSUES + trivy_critical))
        HIGH_ISSUES=$((HIGH_ISSUES + trivy_high))
        MEDIUM_ISSUES=$((MEDIUM_ISSUES + trivy_medium))
    fi
    
    # Check Trivy container results
    if [ -f "$OUTPUT_DIR/trivy-container-results.json" ]; then
        local container_critical=$(jq '[.Results[]?.Vulnerabilities[]? | select(.Severity == "CRITICAL")] | length' "$OUTPUT_DIR/trivy-container-results.json" 2>/dev/null || echo "0")
        local container_high=$(jq '[.Results[]?.Vulnerabilities[]? | select(.Severity == "HIGH")] | length' "$OUTPUT_DIR/trivy-container-results.json" 2>/dev/null || echo "0")
        local container_medium=$(jq '[.Results[]?.Vulnerabilities[]? | select(.Severity == "MEDIUM")] | length' "$OUTPUT_DIR/trivy-container-results.json" 2>/dev/null || echo "0")
        CRITICAL_ISSUES=$((CRITICAL_ISSUES + container_critical))
        HIGH_ISSUES=$((HIGH_ISSUES + container_high))
        MEDIUM_ISSUES=$((MEDIUM_ISSUES + container_medium))
    fi
    
    # Check for secrets
    if [ -f "$OUTPUT_DIR/secret-scan-results.txt" ] && grep -q -v "No secrets found" "$OUTPUT_DIR/secret-scan-results.txt"; then
        CRITICAL_ISSUES=$((CRITICAL_ISSUES + 1))
    fi
    
    # Generate summary
    echo "=== Security Scan Summary ===" > "$OUTPUT_DIR/security-summary.txt"
    echo "Scan completed: $(date)" >> "$OUTPUT_DIR/security-summary.txt"
    echo "" >> "$OUTPUT_DIR/security-summary.txt"
    echo "Issues found:" >> "$OUTPUT_DIR/security-summary.txt"
    echo "- Critical: $CRITICAL_ISSUES" >> "$OUTPUT_DIR/security-summary.txt"
    echo "- High: $HIGH_ISSUES" >> "$OUTPUT_DIR/security-summary.txt"
    echo "- Medium: $MEDIUM_ISSUES" >> "$OUTPUT_DIR/security-summary.txt"
    echo "" >> "$OUTPUT_DIR/security-summary.txt"
    
    if [ "$CRITICAL_ISSUES" -gt 0 ]; then
        echo "⚠️ CRITICAL ISSUES DETECTED!" >> "$OUTPUT_DIR/security-summary.txt"
        echo "Please review and address critical security issues immediately." >> "$OUTPUT_DIR/security-summary.txt"
    else
        echo "✅ No critical security issues detected" >> "$OUTPUT_DIR/security-summary.txt"
    fi
    
    log "📊 Security scan summary:"
    log "   Critical issues: $CRITICAL_ISSUES"
    log "   High issues: $HIGH_ISSUES"
    log "   Medium issues: $MEDIUM_ISSUES"
    
    if [ "$CRITICAL_ISSUES" -gt 0 ] && [ "$FAIL_ON_CRITICAL" = "true" ]; then
        log "❌ Critical security issues detected! Build failed."
        exit 1
    fi
}

# Main execution
main() {
    log "🚀 Starting enhanced security scan..."
    log "Scan directory: $SCAN_DIR"
    log "Output directory: $OUTPUT_DIR"
    log "Scan type: $SCAN_TYPE"
    log "Fail on critical: $FAIL_ON_CRITICAL"
    log "Tools enabled:"
    log "  - gosec: $ENABLE_GOSEC"
    log "  - govulncheck: $ENABLE_GOVULNCHECK"
    log "  - Trivy: $ENABLE_TRIVY"
    log "  - Semgrep: $ENABLE_SEMGREP"
    log "  - Secret scan: $ENABLE_SECRET_SCAN"
    log "  - OWASP ZAP: $ENABLE_ZAP"
    
    case "$SCAN_TYPE" in
        "all")
            run_gosec_scan
            run_govulncheck_scan
            run_trivy_scan
            run_semgrep_scan
            run_secret_scan
            run_container_scan
            run_zap_scan
            ;;
        "code")
            run_gosec_scan
            run_semgrep_scan
            ;;
        "dependencies")
            run_govulncheck_scan
            run_trivy_scan
            ;;
        "secrets")
            run_secret_scan
            ;;
        "containers")
            run_container_scan
            ;;
        *)
            log "❌ Unknown scan type: $SCAN_TYPE"
            exit 1
            ;;
    esac
    
    check_critical_issues
    
    log "✅ Enhanced security scan completed successfully!"
    log "📋 Reports saved to: $OUTPUT_DIR"
    log "📄 Summary available at: $OUTPUT_DIR/security-summary.txt"
}

# Run main function
main "$@"
