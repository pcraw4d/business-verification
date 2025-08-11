#!/bin/bash

# KYB Tool - Security Scanning Script
# Comprehensive security scanning for the KYB platform

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCAN_DIR="${SCAN_DIR:-.}"
OUTPUT_DIR="${OUTPUT_DIR:-./security-reports}"
LOG_FILE="${LOG_FILE:-./security-scan.log}"
FAIL_ON_CRITICAL="${FAIL_ON_CRITICAL:-true}"
SCAN_TYPE="${SCAN_TYPE:-all}" # all, code, deps, container, infra

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Logging function
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1" | tee -a "$LOG_FILE"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$LOG_FILE"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$LOG_FILE"
}

# Check if required tools are installed
check_dependencies() {
    log "Checking security scanning dependencies..."
    
    local missing_tools=()
    
    # Check for Go security tools
    if ! command -v gosec &> /dev/null; then
        missing_tools+=("gosec")
    fi
    
    if ! command -v govulncheck &> /dev/null; then
        missing_tools+=("govulncheck")
    fi
    
    # Check for general security tools
    if ! command -v trivy &> /dev/null; then
        missing_tools+=("trivy")
    fi
    
    if ! command -v bandit &> /dev/null; then
        missing_tools+=("bandit")
    fi
    
    if ! command -v safety &> /dev/null; then
        missing_tools+=("safety")
    fi
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        error "Missing required security tools: ${missing_tools[*]}"
        error "Please install missing tools:"
        for tool in "${missing_tools[@]}"; do
            case $tool in
                "gosec")
                    echo "  go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"
                    ;;
                "govulncheck")
                    echo "  go install golang.org/x/vuln/cmd/govulncheck@latest"
                    ;;
                "trivy")
                    echo "  See: https://aquasecurity.github.io/trivy/latest/getting-started/installation/"
                    ;;
                "bandit")
                    echo "  pip install bandit"
                    ;;
                "safety")
                    echo "  pip install safety"
                    ;;
            esac
        done
        exit 1
    fi
    
    success "All security tools are available"
}

# Static code analysis with security focus
scan_code_security() {
    log "Starting static code security analysis..."
    
    local scan_results="$OUTPUT_DIR/code-security-scan.json"
    local scan_report="$OUTPUT_DIR/code-security-report.txt"
    
    # Run gosec for Go security scanning
    log "Running gosec security scanner..."
    if gosec -fmt=json -out="$scan_results" ./... 2>/dev/null; then
        success "gosec scan completed"
    else
        warning "gosec scan completed with issues"
    fi
    
    # Generate human-readable report
    echo "=== Go Security Scan Results ===" > "$scan_report"
    echo "Generated: $(date)" >> "$scan_report"
    echo "" >> "$scan_report"
    
    if [ -f "$scan_results" ]; then
        # Parse JSON and create readable report
        if command -v jq &> /dev/null; then
            echo "Issues found:" >> "$scan_report"
            jq -r '.Issues[] | "\(.severity | ascii_upcase): \(.details) [\(.file):\(.line)]"' "$scan_results" >> "$scan_report" 2>/dev/null || echo "No issues found" >> "$scan_report"
        else
            echo "Raw scan results available in: $scan_results" >> "$scan_report"
        fi
    fi
    
    # Check for critical issues
    if [ "$FAIL_ON_CRITICAL" = "true" ] && [ -f "$scan_results" ]; then
        if command -v jq &> /dev/null; then
            local critical_count=$(jq '[.Issues[] | select(.severity == "HIGH")] | length' "$scan_results" 2>/dev/null || echo "0")
            if [ "$critical_count" -gt 0 ]; then
                error "Found $critical_count critical security issues"
                return 1
            fi
        fi
    fi
    
    success "Code security scan completed"
}

# Dependency vulnerability scanning
scan_dependencies() {
    log "Starting dependency vulnerability scanning..."
    
    local deps_report="$OUTPUT_DIR/dependency-scan-report.txt"
    
    echo "=== Dependency Vulnerability Scan Results ===" > "$deps_report"
    echo "Generated: $(date)" >> "$deps_report"
    echo "" >> "$deps_report"
    
    # Go module vulnerability scanning
    log "Running govulncheck for Go dependencies..."
    echo "Go Dependencies:" >> "$deps_report"
    if govulncheck ./... 2>&1 | tee -a "$deps_report"; then
        success "govulncheck completed"
    else
        warning "govulncheck found vulnerabilities"
    fi
    
    # Check for known vulnerabilities in go.mod
    log "Checking go.mod for known vulnerabilities..."
    echo "" >> "$deps_report"
    echo "Go Module Analysis:" >> "$deps_report"
    if [ -f "go.mod" ]; then
        go list -m all 2>/dev/null | head -20 >> "$deps_report" || echo "Unable to list Go modules" >> "$deps_report"
    fi
    
    # Python dependencies (if any)
    if [ -f "requirements.txt" ] || [ -f "pyproject.toml" ]; then
        log "Scanning Python dependencies..."
        echo "" >> "$deps_report"
        echo "Python Dependencies:" >> "$deps_report"
        if command -v safety &> /dev/null; then
            safety check 2>&1 | tee -a "$deps_report" || echo "Safety check failed" >> "$deps_report"
        fi
    fi
    
    # Node.js dependencies (if any)
    if [ -f "package.json" ]; then
        log "Scanning Node.js dependencies..."
        echo "" >> "$deps_report"
        echo "Node.js Dependencies:" >> "$deps_report"
        if command -v npm &> /dev/null; then
            npm audit --audit-level=moderate 2>&1 | tee -a "$deps_report" || echo "npm audit failed" >> "$deps_report"
        fi
    fi
    
    success "Dependency vulnerability scan completed"
}

# Container security scanning
scan_containers() {
    log "Starting container security scanning..."
    
    local container_report="$OUTPUT_DIR/container-scan-report.txt"
    
    echo "=== Container Security Scan Results ===" > "$container_report"
    echo "Generated: $(date)" >> "$container_report"
    echo "" >> "$container_report"
    
    # Scan Dockerfile
    if [ -f "Dockerfile" ]; then
        log "Scanning Dockerfile..."
        echo "Dockerfile Analysis:" >> "$container_report"
        if command -v trivy &> /dev/null; then
            trivy config . 2>&1 | tee -a "$container_report" || echo "Trivy config scan failed" >> "$container_report"
        fi
    fi
    
    # Scan docker-compose files
    for compose_file in docker-compose*.yml; do
        if [ -f "$compose_file" ]; then
            log "Scanning $compose_file..."
            echo "" >> "$container_report"
            echo "$compose_file Analysis:" >> "$container_report"
            if command -v trivy &> /dev/null; then
                trivy config "$compose_file" 2>&1 | tee -a "$container_report" || echo "Trivy scan failed for $compose_file" >> "$container_report"
            fi
        fi
    done
    
    # Scan built images (if available)
    if command -v docker &> /dev/null; then
        log "Scanning built Docker images..."
        echo "" >> "$container_report"
        echo "Docker Images:" >> "$container_report"
        
        # Get list of local images
        docker images --format "table {{.Repository}}:{{.Tag}}" | grep -v "REPOSITORY" | while read -r image; do
            if [[ "$image" == *"kyb"* ]] || [[ "$image" == *"business-verification"* ]]; then
                log "Scanning image: $image"
                echo "Scanning: $image" >> "$container_report"
                if command -v trivy &> /dev/null; then
                    trivy image --severity HIGH,CRITICAL "$image" 2>&1 | tee -a "$container_report" || echo "Trivy image scan failed for $image" >> "$container_report"
                fi
            fi
        done
    fi
    
    success "Container security scan completed"
}

# Infrastructure security scanning
scan_infrastructure() {
    log "Starting infrastructure security scanning..."
    
    local infra_report="$OUTPUT_DIR/infrastructure-scan-report.txt"
    
    echo "=== Infrastructure Security Scan Results ===" > "$infra_report"
    echo "Generated: $(date)" >> "$infra_report"
    echo "" >> "$infra_report"
    
    # Scan Terraform configurations
    if [ -d "deployments/terraform" ]; then
        log "Scanning Terraform configurations..."
        echo "Terraform Security Analysis:" >> "$infra_report"
        
        if command -v trivy &> /dev/null; then
            trivy config deployments/terraform 2>&1 | tee -a "$infra_report" || echo "Terraform security scan failed" >> "$infra_report"
        fi
        
        # Check for hardcoded secrets
        echo "" >> "$infra_report"
        echo "Secret Detection:" >> "$infra_report"
        if grep -r -i "password\|secret\|key\|token" deployments/terraform/ 2>/dev/null | grep -v ".git" | head -10; then
            warning "Potential secrets found in Terraform files"
        else
            echo "No obvious secrets found in Terraform files" >> "$infra_report"
        fi
    fi
    
    # Scan Kubernetes configurations
    if [ -d "deployments/kubernetes" ]; then
        log "Scanning Kubernetes configurations..."
        echo "" >> "$infra_report"
        echo "Kubernetes Security Analysis:" >> "$infra_report"
        
        if command -v trivy &> /dev/null; then
            trivy config deployments/kubernetes 2>&1 | tee -a "$infra_report" || echo "Kubernetes security scan failed" >> "$infra_report"
        fi
    fi
    
    # Check for exposed ports and services
    echo "" >> "$infra_report"
    echo "Service Exposure Analysis:" >> "$infra_report"
    if [ -f "docker-compose.yml" ]; then
        echo "Docker Compose Services:" >> "$infra_report"
        grep -A 5 -B 5 "ports:" docker-compose.yml 2>/dev/null | tee -a "$infra_report" || echo "No port configurations found" >> "$infra_report"
    fi
    
    success "Infrastructure security scan completed"
}

# Generate summary report
generate_summary() {
    log "Generating security scan summary..."
    
    local summary_report="$OUTPUT_DIR/security-scan-summary.txt"
    
    echo "=== KYB Tool Security Scan Summary ===" > "$summary_report"
    echo "Generated: $(date)" >> "$summary_report"
    echo "Scan Type: $SCAN_TYPE" >> "$summary_report"
    echo "" >> "$summary_report"
    
    # Count issues by severity
    local critical_count=0
    local high_count=0
    local medium_count=0
    local low_count=0
    
    # Parse gosec results
    if [ -f "$OUTPUT_DIR/code-security-scan.json" ] && command -v jq &> /dev/null; then
        critical_count=$(jq '[.Issues[] | select(.severity == "HIGH")] | length' "$OUTPUT_DIR/code-security-scan.json" 2>/dev/null || echo "0")
        high_count=$(jq '[.Issues[] | select(.severity == "MEDIUM")] | length' "$OUTPUT_DIR/code-security-scan.json" 2>/dev/null || echo "0")
        medium_count=$(jq '[.Issues[] | select(.severity == "LOW")] | length' "$OUTPUT_DIR/code-security-scan.json" 2>/dev/null || echo "0")
    fi
    
    echo "Security Issues Summary:" >> "$summary_report"
    echo "  Critical: $critical_count" >> "$summary_report"
    echo "  High: $high_count" >> "$summary_report"
    echo "  Medium: $medium_count" >> "$summary_report"
    echo "  Low: $low_count" >> "$summary_report"
    echo "" >> "$summary_report"
    
    # List all generated reports
    echo "Generated Reports:" >> "$summary_report"
    for report in "$OUTPUT_DIR"/*.txt "$OUTPUT_DIR"/*.json; do
        if [ -f "$report" ]; then
            echo "  - $(basename "$report")" >> "$summary_report"
        fi
    done
    
    echo "" >> "$summary_report"
    echo "Next Steps:" >> "$summary_report"
    if [ "$critical_count" -gt 0 ]; then
        echo "  ⚠️  CRITICAL: Address $critical_count critical security issues immediately" >> "$summary_report"
    fi
    if [ "$high_count" -gt 0 ]; then
        echo "  ⚠️  HIGH: Review and fix $high_count high severity issues" >> "$summary_report"
    fi
    if [ "$medium_count" -gt 0 ]; then
        echo "  ℹ️  MEDIUM: Consider addressing $medium_count medium severity issues" >> "$summary_report"
    fi
    if [ "$low_count" -gt 0 ]; then
        echo "  ℹ️  LOW: Monitor $low_count low severity issues" >> "$summary_report"
    fi
    
    if [ "$critical_count" -eq 0 ] && [ "$high_count" -eq 0 ]; then
        echo "  ✅ No critical or high severity issues found" >> "$summary_report"
    fi
    
    success "Security scan summary generated: $summary_report"
}

# Main execution
main() {
    log "Starting KYB Tool security scanning..."
    log "Scan directory: $SCAN_DIR"
    log "Output directory: $OUTPUT_DIR"
    log "Scan type: $SCAN_TYPE"
    
    # Check dependencies
    check_dependencies
    
    # Run scans based on type
    case $SCAN_TYPE in
        "all")
            scan_code_security
            scan_dependencies
            scan_containers
            scan_infrastructure
            ;;
        "code")
            scan_code_security
            ;;
        "deps")
            scan_dependencies
            ;;
        "container")
            scan_containers
            ;;
        "infra")
            scan_infrastructure
            ;;
        *)
            error "Invalid scan type: $SCAN_TYPE"
            echo "Valid types: all, code, deps, container, infra"
            exit 1
            ;;
    esac
    
    # Generate summary
    generate_summary
    
    log "Security scanning completed successfully"
    log "Reports available in: $OUTPUT_DIR"
    
    # Exit with error if critical issues found and fail_on_critical is true
    if [ "$FAIL_ON_CRITICAL" = "true" ]; then
        if [ -f "$OUTPUT_DIR/code-security-scan.json" ] && command -v jq &> /dev/null; then
            local critical_count=$(jq '[.Issues[] | select(.severity == "HIGH")] | length' "$OUTPUT_DIR/code-security-scan.json" 2>/dev/null || echo "0")
            if [ "$critical_count" -gt 0 ]; then
                error "Security scan failed: $critical_count critical issues found"
                exit 1
            fi
        fi
    fi
    
    success "Security scanning completed without critical issues"
}

# Handle script arguments
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
        --help)
            echo "KYB Tool Security Scanner"
            echo ""
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --scan-dir DIR        Directory to scan (default: .)"
            echo "  --output-dir DIR      Output directory for reports (default: ./security-reports)"
            echo "  --scan-type TYPE      Type of scan: all, code, deps, container, infra (default: all)"
            echo "  --fail-on-critical    Exit with error on critical issues (default: true)"
            echo "  --help                Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                                    # Full security scan"
            echo "  $0 --scan-type code                  # Code security only"
            echo "  $0 --scan-type deps                  # Dependency scan only"
            echo "  $0 --fail-on-critical false         # Don't fail on critical issues"
            exit 0
            ;;
        *)
            error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Run main function
main "$@"
