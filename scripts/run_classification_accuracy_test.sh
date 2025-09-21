#!/bin/bash

# Classification Accuracy Test Runner Script
# This script runs comprehensive classification accuracy testing

set -e  # Exit on any error

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CONFIG_FILE="${PROJECT_ROOT}/configs/dev/config.yaml"
LOG_FILE="${PROJECT_ROOT}/logs/classification_accuracy_test.log"
RESULTS_DIR="${PROJECT_ROOT}/test_results/classification_accuracy"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$LOG_FILE"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE"
}

# Create necessary directories
create_directories() {
    log_info "Creating necessary directories..."
    mkdir -p "$(dirname "$LOG_FILE")"
    mkdir -p "$RESULTS_DIR"
    mkdir -p "${PROJECT_ROOT}/logs"
    log_success "Directories created successfully"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    # Check if PostgreSQL client is available
    if ! command -v psql &> /dev/null; then
        log_warning "PostgreSQL client (psql) not found. Some database operations may fail."
    fi
    
    # Check if configuration file exists
    if [ ! -f "$CONFIG_FILE" ]; then
        log_error "Configuration file not found: $CONFIG_FILE"
        exit 1
    fi
    
    # Check if database migration has been run
    if [ ! -f "${PROJECT_ROOT}/migrations/classification_accuracy_testing_schema.sql" ]; then
        log_error "Classification accuracy testing schema not found. Please run database migrations first."
        exit 1
    fi
    
    log_success "Prerequisites check completed"
}

# Run database migrations
run_migrations() {
    log_info "Running database migrations for classification accuracy testing..."
    
    # Check if we can connect to the database
    if ! psql -h localhost -U postgres -d kyb_platform -c "SELECT 1;" &> /dev/null; then
        log_warning "Cannot connect to database. Please ensure PostgreSQL is running and accessible."
        log_info "Attempting to run migrations anyway..."
    fi
    
    # Run the classification accuracy testing schema migration
    if psql -h localhost -U postgres -d kyb_platform -f "${PROJECT_ROOT}/migrations/classification_accuracy_testing_schema.sql" &>> "$LOG_FILE"; then
        log_success "Database migrations completed successfully"
    else
        log_warning "Database migration failed or database not accessible. Continuing with test..."
    fi
}

# Generate test data
generate_test_data() {
    log_info "Generating comprehensive test data..."
    
    cd "$PROJECT_ROOT"
    
    # Build the test data generator
    if go build -o bin/test_data_generator ./cmd/test_runner/; then
        log_success "Test data generator built successfully"
    else
        log_error "Failed to build test data generator"
        exit 1
    fi
    
    # Run test data generation
    if ./bin/test_data_generator -config "$CONFIG_FILE" -generate-data &>> "$LOG_FILE"; then
        log_success "Test data generated successfully"
    else
        log_warning "Test data generation failed or database not accessible. Using existing test data..."
    fi
}

# Run classification accuracy test
run_accuracy_test() {
    log_info "Running classification accuracy test..."
    
    cd "$PROJECT_ROOT"
    
    # Build the test runner
    if go build -o bin/classification_accuracy_test ./cmd/test_runner/; then
        log_success "Classification accuracy test runner built successfully"
    else
        log_error "Failed to build classification accuracy test runner"
        exit 1
    fi
    
    # Run the accuracy test
    log_info "Starting comprehensive classification accuracy test..."
    if ./bin/classification_accuracy_test -config "$CONFIG_FILE" -verbose &>> "$LOG_FILE"; then
        log_success "Classification accuracy test completed successfully"
    else
        log_error "Classification accuracy test failed"
        exit 1
    fi
}

# Generate test report
generate_report() {
    log_info "Generating test report..."
    
    # Create a simple HTML report
    REPORT_FILE="${RESULTS_DIR}/classification_accuracy_report.html"
    
    cat > "$REPORT_FILE" << EOF
<!DOCTYPE html>
<html>
<head>
    <title>Classification Accuracy Test Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background-color: #f0f0f0; padding: 20px; border-radius: 5px; }
        .section { margin: 20px 0; }
        .metric { display: inline-block; margin: 10px; padding: 10px; background-color: #e8f4f8; border-radius: 5px; }
        .success { color: green; }
        .warning { color: orange; }
        .error { color: red; }
        .log { background-color: #f5f5f5; padding: 10px; border-radius: 5px; font-family: monospace; white-space: pre-wrap; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Classification Accuracy Test Report</h1>
        <p>Generated on: $(date)</p>
        <p>Test Duration: $(grep "Accuracy test completed in" "$LOG_FILE" | tail -1 | cut -d' ' -f6- || echo "Unknown")</p>
    </div>
    
    <div class="section">
        <h2>Test Summary</h2>
        <p>This report contains the results of comprehensive classification accuracy testing for the KYB Platform.</p>
        <p>For detailed results, please check the log file: <code>$LOG_FILE</code></p>
    </div>
    
    <div class="section">
        <h2>Key Metrics</h2>
        <div class="metric">
            <strong>Overall Accuracy:</strong> <span class="success">See log for details</span>
        </div>
        <div class="metric">
            <strong>Processing Time:</strong> <span class="success">See log for details</span>
        </div>
        <div class="metric">
            <strong>Error Rate:</strong> <span class="success">See log for details</span>
        </div>
    </div>
    
    <div class="section">
        <h2>Test Log</h2>
        <div class="log">$(tail -50 "$LOG_FILE")</div>
    </div>
    
    <div class="section">
        <h2>Next Steps</h2>
        <ul>
            <li>Review the accuracy metrics and recommendations in the log file</li>
            <li>Address any issues identified in the error analysis</li>
            <li>Implement recommended improvements</li>
            <li>Re-run tests to validate improvements</li>
        </ul>
    </div>
</body>
</html>
EOF
    
    log_success "Test report generated: $REPORT_FILE"
}

# Cleanup function
cleanup() {
    log_info "Cleaning up temporary files..."
    rm -f "${PROJECT_ROOT}/bin/test_data_generator"
    rm -f "${PROJECT_ROOT}/bin/classification_accuracy_test"
    log_success "Cleanup completed"
}

# Main execution
main() {
    log_info "Starting Classification Accuracy Test Suite"
    log_info "Project Root: $PROJECT_ROOT"
    log_info "Log File: $LOG_FILE"
    log_info "Results Directory: $RESULTS_DIR"
    
    # Create directories
    create_directories
    
    # Check prerequisites
    check_prerequisites
    
    # Run database migrations
    run_migrations
    
    # Generate test data
    generate_test_data
    
    # Run accuracy test
    run_accuracy_test
    
    # Generate report
    generate_report
    
    # Cleanup
    cleanup
    
    log_success "Classification Accuracy Test Suite completed successfully!"
    log_info "Check the log file for detailed results: $LOG_FILE"
    log_info "Check the HTML report: ${RESULTS_DIR}/classification_accuracy_report.html"
    
    # Display summary
    echo ""
    echo "=========================================="
    echo "CLASSIFICATION ACCURACY TEST SUMMARY"
    echo "=========================================="
    echo "Log File: $LOG_FILE"
    echo "Report: ${RESULTS_DIR}/classification_accuracy_report.html"
    echo "Results Directory: $RESULTS_DIR"
    echo ""
    echo "To view the latest test results:"
    echo "  tail -f $LOG_FILE"
    echo ""
    echo "To view the HTML report:"
    echo "  open ${RESULTS_DIR}/classification_accuracy_report.html"
    echo "=========================================="
}

# Handle script interruption
trap cleanup EXIT

# Run main function
main "$@"
