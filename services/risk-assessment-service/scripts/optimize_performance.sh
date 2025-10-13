#!/bin/bash

# Performance Optimization Script
# Risk Assessment Service - Phase 4.8 Implementation

set -e

# Configuration
SERVICE_NAME="risk-assessment-service"
NAMESPACE="kyb-platform"
OPTIMIZATION_TARGET="sub-1-second"
LOG_DIR="logs/optimization"
mkdir -p "$LOG_DIR"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
LOG_FILE="$LOG_DIR/performance_optimization_$TIMESTAMP.log"

# Function to log with timestamp
log() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1" | tee -a "$LOG_FILE"
}

# Function to log success
log_success() {
    echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')] ✅ $1${NC}" | tee -a "$LOG_FILE"
}

# Function to log warning
log_warning() {
    echo -e "${YELLOW}[$(date '+%Y-%m-%d %H:%M:%S')] ⚠️  $1${NC}" | tee -a "$LOG_FILE"
}

# Function to log error
log_error() {
    echo -e "${RED}[$(date '+%Y-%m-%d %H:%M:%S')] ❌ $1${NC}" | tee -a "$LOG_FILE"
}

# Function to check prerequisites
check_prerequisites() {
    log "Checking optimization prerequisites..."
    
    # Check if service is running
    if ! curl -s http://localhost:8080/health > /dev/null; then
        log_error "Service is not running. Please start the service first."
        exit 1
    fi
    
    # Check if optimization endpoints are available
    if ! curl -s http://localhost:8080/api/v1/performance/optimization/status > /dev/null; then
        log_error "Performance optimization endpoints are not available."
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Function to get current performance metrics
get_current_metrics() {
    log "Getting current performance metrics..."
    
    # Get optimization status
    curl -s http://localhost:8080/api/v1/performance/optimization/status > "$LOG_DIR/current_metrics.json"
    
    # Extract key metrics
    P95_LATENCY=$(jq -r '.response_times.p95_latency' "$LOG_DIR/current_metrics.json" 2>/dev/null || echo "unknown")
    P99_LATENCY=$(jq -r '.response_times.p99_latency' "$LOG_DIR/current_metrics.json" 2>/dev/null || echo "unknown")
    AVG_LATENCY=$(jq -r '.response_times.avg_latency' "$LOG_DIR/current_metrics.json" 2>/dev/null || echo "unknown")
    OPTIMIZATION_SCORE=$(jq -r '.overall.optimization_score' "$LOG_DIR/current_metrics.json" 2>/dev/null || echo "unknown")
    
    log "Current Performance Metrics:"
    log "  P95 Latency: $P95_LATENCY"
    log "  P99 Latency: $P99_LATENCY"
    log "  Average Latency: $AVG_LATENCY"
    log "  Optimization Score: $OPTIMIZATION_SCORE"
    
    # Check if targets are met
    if [[ "$P95_LATENCY" != "unknown" && "$P95_LATENCY" < "1s" ]]; then
        log_success "P95 latency target (1s) is met: $P95_LATENCY"
        P95_TARGET_MET=true
    else
        log_warning "P95 latency target (1s) not met: $P95_LATENCY"
        P95_TARGET_MET=false
    fi
    
    if [[ "$P99_LATENCY" != "unknown" && "$P99_LATENCY" < "2s" ]]; then
        log_success "P99 latency target (2s) is met: $P99_LATENCY"
        P99_TARGET_MET=true
    else
        log_warning "P99 latency target (2s) not met: $P99_LATENCY"
        P99_TARGET_MET=false
    fi
    
    if [[ "$AVG_LATENCY" != "unknown" && "$AVG_LATENCY" < "500ms" ]]; then
        log_success "Average latency target (500ms) is met: $AVG_LATENCY"
        AVG_TARGET_MET=true
    else
        log_warning "Average latency target (500ms) not met: $AVG_LATENCY"
        AVG_TARGET_MET=false
    fi
}

# Function to run database optimization
optimize_database() {
    log "Running database optimization..."
    
    # Trigger database optimization
    curl -s -X POST http://localhost:8080/api/v1/performance/optimization/database > "$LOG_DIR/database_optimization.json"
    
    # Check if optimization was successful
    if jq -e '.message' "$LOG_DIR/database_optimization.json" > /dev/null; then
        log_success "Database optimization completed successfully"
    else
        log_error "Database optimization failed"
        return 1
    fi
    
    # Get database optimization status
    curl -s http://localhost:8080/api/v1/performance/optimization/database > "$LOG_DIR/database_status.json"
    
    # Extract database metrics
    DB_OPTIMIZATION_SCORE=$(jq -r '.stats.optimization_score' "$LOG_DIR/database_status.json" 2>/dev/null || echo "unknown")
    DB_IS_OPTIMIZED=$(jq -r '.stats.is_optimized' "$LOG_DIR/database_status.json" 2>/dev/null || echo "unknown")
    
    log "Database Optimization Results:"
    log "  Optimization Score: $DB_OPTIMIZATION_SCORE"
    log "  Is Optimized: $DB_IS_OPTIMIZED"
    
    # Show recommendations
    jq -r '.recommendations[]?' "$LOG_DIR/database_status.json" 2>/dev/null | while read -r recommendation; do
        if [[ -n "$recommendation" ]]; then
            log "  Recommendation: $recommendation"
        fi
    done
}

# Function to run response time optimization
optimize_response_time() {
    log "Running response time optimization..."
    
    # Trigger response time optimization
    curl -s -X POST http://localhost:8080/api/v1/performance/optimization/response-time > "$LOG_DIR/response_time_optimization.json"
    
    # Check if optimization was successful
    if jq -e '.message' "$LOG_DIR/response_time_optimization.json" > /dev/null; then
        log_success "Response time optimization completed successfully"
    else
        log_error "Response time optimization failed"
        return 1
    fi
    
    # Get response time optimization status
    curl -s http://localhost:8080/api/v1/performance/optimization/response-time > "$LOG_DIR/response_time_status.json"
    
    # Extract response time metrics
    RT_OPTIMIZATION_SCORE=$(jq -r '.stats.optimization_score' "$LOG_DIR/response_time_status.json" 2>/dev/null || echo "unknown")
    RT_IS_OPTIMIZED=$(jq -r '.stats.is_optimized' "$LOG_DIR/response_time_status.json" 2>/dev/null || echo "unknown")
    
    log "Response Time Optimization Results:"
    log "  Optimization Score: $RT_OPTIMIZATION_SCORE"
    log "  Is Optimized: $RT_IS_OPTIMIZED"
    
    # Show recommendations
    jq -r '.recommendations[]?' "$LOG_DIR/response_time_status.json" 2>/dev/null | while read -r recommendation; do
        if [[ -n "$recommendation" ]]; then
            log "  Recommendation: $recommendation"
        fi
    done
}

# Function to run comprehensive optimization
run_comprehensive_optimization() {
    log "Running comprehensive performance optimization..."
    
    # Trigger comprehensive optimization
    curl -s -X POST http://localhost:8080/api/v1/performance/optimization/optimize > "$LOG_DIR/comprehensive_optimization.json"
    
    # Check if optimization was successful
    if jq -e '.message' "$LOG_DIR/comprehensive_optimization.json" > /dev/null; then
        log_success "Comprehensive optimization completed successfully"
    else
        log_error "Comprehensive optimization failed"
        return 1
    fi
    
    # Extract optimization results
    OPTIMIZATION_SCORE=$(jq -r '.optimization_score' "$LOG_DIR/comprehensive_optimization.json" 2>/dev/null || echo "unknown")
    IS_OPTIMIZED=$(jq -r '.is_optimized' "$LOG_DIR/comprehensive_optimization.json" 2>/dev/null || echo "unknown")
    
    log "Comprehensive Optimization Results:"
    log "  Optimization Score: $OPTIMIZATION_SCORE"
    log "  Is Optimized: $IS_OPTIMIZED"
}

# Function to set performance targets
set_performance_targets() {
    log "Setting performance targets for sub-1-second response times..."
    
    # Set response time targets
    curl -s -X PUT http://localhost:8080/api/v1/performance/optimization/targets \
        -H "Content-Type: application/json" \
        -d '{
            "p95_target": "1s",
            "p99_target": "2s",
            "avg_target": "500ms",
            "max_target": "5s"
        }' > "$LOG_DIR/set_targets.json"
    
    # Check if targets were set successfully
    if jq -e '.message' "$LOG_DIR/set_targets.json" > /dev/null; then
        log_success "Performance targets set successfully"
    else
        log_error "Failed to set performance targets"
        return 1
    fi
    
    # Show set targets
    jq -r '.p95_target, .p99_target, .avg_target, .max_target' "$LOG_DIR/set_targets.json" | while read -r target; do
        if [[ -n "$target" ]]; then
            log "  Target: $target"
        fi
    done
}

# Function to run load test
run_load_test() {
    log "Running load test to validate performance optimization..."
    
    # Check if load test script exists
    if [[ -f "scripts/load_test_10k.sh" ]]; then
        # Run quick load test
        ./scripts/load_test_10k.sh --quick
        log_success "Load test completed"
    else
        log_warning "Load test script not found, skipping load test"
    fi
}

# Function to generate optimization report
generate_optimization_report() {
    log "Generating performance optimization report..."
    
    REPORT_FILE="$LOG_DIR/optimization_report_$TIMESTAMP.md"
    
    cat > "$REPORT_FILE" << EOF
# Performance Optimization Report
**Optimization Date**: $(date)
**Service**: $SERVICE_NAME
**Target**: $OPTIMIZATION_TARGET
**Namespace**: $NAMESPACE

## Optimization Summary

### Performance Targets
- **P95 Latency**: < 1 second
- **P99 Latency**: < 2 seconds
- **Average Latency**: < 500ms
- **Maximum Latency**: < 5 seconds

### Current Performance Metrics
- **P95 Latency**: $P95_LATENCY
- **P99 Latency**: $P99_LATENCY
- **Average Latency**: $AVG_LATENCY
- **Optimization Score**: $OPTIMIZATION_SCORE

### Target Achievement
- **P95 Target Met**: $P95_TARGET_MET
- **P99 Target Met**: $P99_TARGET_MET
- **Average Target Met**: $AVG_TARGET_MET

## Optimization Results

### Database Optimization
EOF

    if [[ -f "$LOG_DIR/database_optimization.json" ]]; then
        DB_SCORE=$(jq -r '.optimization_score' "$LOG_DIR/database_optimization.json" 2>/dev/null || echo "unknown")
        DB_OPTIMIZED=$(jq -r '.is_optimized' "$LOG_DIR/database_optimization.json" 2>/dev/null || echo "unknown")
        cat >> "$REPORT_FILE" << EOF
- **Optimization Score**: $DB_SCORE
- **Is Optimized**: $DB_OPTIMIZED
EOF
    fi

    cat >> "$REPORT_FILE" << EOF

### Response Time Optimization
EOF

    if [[ -f "$LOG_DIR/response_time_optimization.json" ]]; then
        RT_SCORE=$(jq -r '.optimization_score' "$LOG_DIR/response_time_optimization.json" 2>/dev/null || echo "unknown")
        RT_OPTIMIZED=$(jq -r '.is_optimized' "$LOG_DIR/response_time_optimization.json" 2>/dev/null || echo "unknown")
        cat >> "$REPORT_FILE" << EOF
- **Optimization Score**: $RT_SCORE
- **Is Optimized**: $RT_OPTIMIZED
EOF
    fi

    cat >> "$REPORT_FILE" << EOF

## Recommendations

### Database Recommendations
EOF

    if [[ -f "$LOG_DIR/database_status.json" ]]; then
        jq -r '.recommendations[]?' "$LOG_DIR/database_status.json" 2>/dev/null | while read -r recommendation; do
            if [[ -n "$recommendation" ]]; then
                echo "- $recommendation" >> "$REPORT_FILE"
            fi
        done
    fi

    cat >> "$REPORT_FILE" << EOF

### Response Time Recommendations
EOF

    if [[ -f "$LOG_DIR/response_time_status.json" ]]; then
        jq -r '.recommendations[]?' "$LOG_DIR/response_time_status.json" 2>/dev/null | while read -r recommendation; do
            if [[ -n "$recommendation" ]]; then
                echo "- $recommendation" >> "$REPORT_FILE"
            fi
        done
    fi

    cat >> "$REPORT_FILE" << EOF

## Next Steps
1. Monitor performance metrics continuously
2. Implement recommended optimizations
3. Run regular load tests
4. Adjust targets based on results
5. Scale infrastructure as needed

## Files Generated
- Current Metrics: \`$LOG_DIR/current_metrics.json\`
- Database Optimization: \`$LOG_DIR/database_optimization.json\`
- Response Time Optimization: \`$LOG_DIR/response_time_optimization.json\`
- Comprehensive Optimization: \`$LOG_DIR/comprehensive_optimization.json\`
- Performance Targets: \`$LOG_DIR/set_targets.json\`

EOF
    
    log_success "Optimization report generated: $REPORT_FILE"
}

# Main execution
main() {
    log "Starting performance optimization for sub-1-second response times..."
    log "Configuration:"
    log "  Service: $SERVICE_NAME"
    log "  Target: $OPTIMIZATION_TARGET"
    log "  Namespace: $NAMESPACE"
    log "  Log File: $LOG_FILE"
    
    # Execute optimization phases
    check_prerequisites
    get_current_metrics
    set_performance_targets
    optimize_database
    optimize_response_time
    run_comprehensive_optimization
    run_load_test
    generate_optimization_report
    
    log_success "Performance optimization completed successfully!"
    log "Optimization report: $LOG_DIR/optimization_report_$TIMESTAMP.md"
    
    # Final status check
    log "Final performance status:"
    get_current_metrics
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [OPTIONS]"
        echo ""
        echo "Options:"
        echo "  --help, -h          Show this help message"
        echo "  --database-only     Run database optimization only"
        echo "  --response-time-only Run response time optimization only"
        echo "  --comprehensive     Run comprehensive optimization only"
        echo "  --status-only       Get current performance status only"
        echo "  --set-targets       Set performance targets only"
        echo ""
        echo "Examples:"
        echo "  $0                  # Run full optimization"
        echo "  $0 --database-only  # Run database optimization only"
        echo "  $0 --status-only    # Get current status only"
        echo ""
        exit 0
        ;;
    --database-only)
        check_prerequisites
        get_current_metrics
        optimize_database
        ;;
    --response-time-only)
        check_prerequisites
        get_current_metrics
        optimize_response_time
        ;;
    --comprehensive)
        check_prerequisites
        get_current_metrics
        run_comprehensive_optimization
        ;;
    --status-only)
        check_prerequisites
        get_current_metrics
        ;;
    --set-targets)
        check_prerequisites
        set_performance_targets
        ;;
    *)
        main
        ;;
esac
