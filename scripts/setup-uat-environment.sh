#!/bin/bash

# KYB Platform - UAT Environment Setup Script
# This script sets up the UAT environment with test data and scenarios

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if application is running
check_application() {
    if curl -f -s "http://localhost:8080/health" > /dev/null 2>&1; then
        print_success "Application is running and healthy"
        return 0
    else
        print_error "Application is not accessible"
        return 1
    fi
}

# Function to create test data directory
create_test_data_directory() {
    print_status "Creating test data directory..."
    
    mkdir -p test/uat/data
    mkdir -p test/uat/scenarios
    mkdir -p test/uat/results
    
    print_success "Test data directory structure created"
}

# Function to create business classification test data
create_business_classification_test_data() {
    print_status "Creating business classification test data..."
    
    cat > test/uat/data/business_classification_test_cases.json << EOF
{
  "test_cases": [
    {
      "id": "BC-001",
      "name": "Technology Corporation",
      "input": {
        "business_name": "Acme Technology Corporation",
        "business_type": "Corporation",
        "industry": "Technology"
      },
      "expected_output": {
        "naics_code": "511210",
        "confidence_score": 0.9,
        "business_category": "Software Publishers"
      }
    },
    {
      "id": "BC-002",
      "name": "Financial Services LLC",
      "input": {
        "business_name": "A & B Financial Services LLC",
        "business_type": "Limited Liability Company",
        "industry": "Financial Services"
      },
      "expected_output": {
        "naics_code": "522320",
        "confidence_score": 0.85,
        "business_category": "Financial Transaction Processing"
      }
    },
    {
      "id": "BC-003",
      "name": "International Consulting",
      "input": {
        "business_name": "Global Solutions Ltd",
        "business_type": "Limited Company",
        "industry": "Consulting"
      },
      "expected_output": {
        "naics_code": "541611",
        "confidence_score": 0.8,
        "business_category": "Administrative Management"
      }
    },
    {
      "id": "BC-004",
      "name": "Healthcare Provider",
      "input": {
        "business_name": "Metro Medical Center",
        "business_type": "Corporation",
        "industry": "Healthcare"
      },
      "expected_output": {
        "naics_code": "622110",
        "confidence_score": 0.95,
        "business_category": "General Medical and Surgical Hospitals"
      }
    },
    {
      "id": "BC-005",
      "name": "Manufacturing Company",
      "input": {
        "business_name": "Precision Manufacturing Inc",
        "business_type": "Corporation",
        "industry": "Manufacturing"
      },
      "expected_output": {
        "naics_code": "332996",
        "confidence_score": 0.88,
        "business_category": "Fabricated Pipe and Pipe Fitting Manufacturing"
      }
    }
  ]
}
EOF
    
    print_success "Business classification test data created"
}

# Function to create risk assessment test data
create_risk_assessment_test_data() {
    print_status "Creating risk assessment test data..."
    
    cat > test/uat/data/risk_assessment_test_cases.json << EOF
{
  "test_cases": [
    {
      "id": "RA-001",
      "name": "Low Risk - Established Bank",
      "input": {
        "business_id": "test-low-risk-001",
        "business_name": "Established Bank Corp",
        "business_type": "Corporation",
        "industry": "Financial Services",
        "years_in_business": 15,
        "annual_revenue": 1000000000,
        "employee_count": 5000,
        "credit_score": 750,
        "regulatory_compliance": "Compliant"
      },
      "expected_output": {
        "risk_score": 0.2,
        "risk_level": "Low",
        "risk_factors": ["Established business", "High revenue", "Good compliance"]
      }
    },
    {
      "id": "RA-002",
      "name": "High Risk - New Startup",
      "input": {
        "business_id": "test-high-risk-001",
        "business_name": "New Startup LLC",
        "business_type": "Limited Liability Company",
        "industry": "Technology",
        "years_in_business": 1,
        "annual_revenue": 100000,
        "employee_count": 5,
        "credit_score": 650,
        "regulatory_compliance": "Unknown"
      },
      "expected_output": {
        "risk_score": 0.8,
        "risk_level": "High",
        "risk_factors": ["New business", "Low revenue", "Unknown compliance"]
      }
    },
    {
      "id": "RA-003",
      "name": "Medium Risk - Growing Company",
      "input": {
        "business_id": "test-medium-risk-001",
        "business_name": "Growing Company Inc",
        "business_type": "Corporation",
        "industry": "Retail",
        "years_in_business": 5,
        "annual_revenue": 5000000,
        "employee_count": 100,
        "credit_score": 700,
        "regulatory_compliance": "Compliant"
      },
      "expected_output": {
        "risk_score": 0.5,
        "risk_level": "Medium",
        "risk_factors": ["Established business", "Moderate revenue", "Good compliance"]
      }
    }
  ]
}
EOF
    
    print_success "Risk assessment test data created"
}

# Function to create compliance test data
create_compliance_test_data() {
    print_status "Creating compliance test data..."
    
    cat > test/uat/data/compliance_test_cases.json << EOF
{
  "test_cases": [
    {
      "id": "COMP-001",
      "name": "SOC2 Compliance Check",
      "input": {
        "business_id": "test-soc2-001",
        "business_name": "Financial Services Corp",
        "industry": "Financial Services",
        "frameworks": ["SOC2"],
        "data_handling": "Customer Data",
        "security_controls": "Implemented"
      },
      "expected_output": {
        "compliance_status": "Compliant",
        "requirements": [
          "Access Control",
          "Data Encryption",
          "Audit Logging",
          "Incident Response"
        ],
        "score": 0.9
      }
    },
    {
      "id": "COMP-002",
      "name": "PCI-DSS Compliance Check",
      "input": {
        "business_id": "test-pci-001",
        "business_name": "E-commerce Solutions Inc",
        "industry": "Retail",
        "frameworks": ["PCI-DSS"],
        "payment_processing": "Credit Cards",
        "security_controls": "Implemented"
      },
      "expected_output": {
        "compliance_status": "Compliant",
        "requirements": [
          "Build and Maintain a Secure Network",
          "Protect Cardholder Data",
          "Maintain Vulnerability Management",
          "Implement Strong Access Controls"
        ],
        "score": 0.85
      }
    },
    {
      "id": "COMP-003",
      "name": "GDPR Compliance Check",
      "input": {
        "business_id": "test-gdpr-001",
        "business_name": "European Data Corp",
        "industry": "Technology",
        "frameworks": ["GDPR"],
        "data_subjects": "EU Citizens",
        "data_processing": "Personal Data"
      },
      "expected_output": {
        "compliance_status": "Compliant",
        "requirements": [
          "Data Minimization",
          "Consent Management",
          "Data Subject Rights",
          "Data Retention"
        ],
        "score": 0.8
      }
    }
  ]
}
EOF
    
    print_success "Compliance test data created"
}

# Function to create UAT scenarios
create_uat_scenarios() {
    print_status "Creating UAT scenarios..."
    
    cat > test/uat/scenarios/uat_scenarios.md << EOF
# KYB Platform - UAT Scenarios

## Scenario 1: Business Classification Workflow

### Objective
Test the complete business classification workflow with various business types.

### Steps
1. Submit business classification request
2. Verify NAICS code assignment
3. Check confidence score
4. Validate business category

### Test Cases
- Technology Corporation
- Financial Services LLC
- International Consulting
- Healthcare Provider
- Manufacturing Company

### Success Criteria
- All test cases return valid NAICS codes
- Confidence scores > 0.8 for standard businesses
- Response time < 500ms

## Scenario 2: Risk Assessment Workflow

### Objective
Test risk assessment functionality for different business profiles.

### Steps
1. Submit business information
2. Calculate risk score
3. Determine risk level
4. Identify risk factors

### Test Cases
- Low Risk: Established Bank
- High Risk: New Startup
- Medium Risk: Growing Company

### Success Criteria
- Risk scores accurately reflect business profile
- Risk levels properly categorized
- Risk factors identified correctly

## Scenario 3: Compliance Checking Workflow

### Objective
Test compliance framework checking for different industries.

### Steps
1. Select compliance framework
2. Submit business information
3. Check compliance status
4. List requirements

### Test Cases
- SOC2: Financial Services
- PCI-DSS: E-commerce
- GDPR: European Data

### Success Criteria
- Compliance status accurately determined
- Requirements properly listed
- Scores calculated correctly

## Scenario 4: End-to-End User Journey

### Objective
Test complete user journey from business input to compliance report.

### Steps
1. Business classification
2. Risk assessment
3. Compliance checking
4. Generate comprehensive report

### Success Criteria
- All steps complete successfully
- Data consistency maintained
- Report generation works
- Performance meets targets

## Scenario 5: Error Handling and Edge Cases

### Objective
Test system behavior with invalid inputs and edge cases.

### Steps
1. Submit malformed data
2. Test missing required fields
3. Test invalid business types
4. Test boundary conditions

### Success Criteria
- Proper error messages returned
- System remains stable
- No data corruption
- Graceful degradation

## Scenario 6: Performance and Load Testing

### Objective
Test system performance under various load conditions.

### Steps
1. Baseline performance test
2. Load test with multiple users
3. Stress test with high load
4. Endurance test

### Success Criteria
- Response times < 500ms
- System handles concurrent users
- No memory leaks
- Stable performance

## Scenario 7: Integration Testing

### Objective
Test integration between different system components.

### Steps
1. Database connectivity
2. Cache functionality
3. External API calls
4. Monitoring integration

### Success Criteria
- All integrations working
- Data consistency maintained
- Error handling works
- Monitoring active

EOF
    
    print_success "UAT scenarios created"
}

# Function to create test execution script
create_test_execution_script() {
    print_status "Creating test execution script..."
    
    cat > test/uat/run_uat_tests.sh << 'EOF'
#!/bin/bash

# KYB Platform - UAT Test Execution Script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to run business classification tests
run_business_classification_tests() {
    print_status "Running Business Classification Tests..."
    
    local test_file="data/business_classification_test_cases.json"
    local results_file="results/business_classification_results.json"
    
    # Create results directory
    mkdir -p results
    
    # Run tests and collect results
    echo "[]" > "$results_file"
    
    # Extract test cases and run them
    jq -c '.test_cases[]' "$test_file" | while read -r test_case; do
        local test_id=$(echo "$test_case" | jq -r '.id')
        local test_name=$(echo "$test_case" | jq -r '.name')
        local input=$(echo "$test_case" | jq -r '.input')
        
        print_status "Running test: $test_name ($test_id)"
        
        # Make API call
        local response=$(curl -s -X POST \
            -H "Content-Type: application/json" \
            -d "$input" \
            "http://localhost:8080/v1/classify" 2>/dev/null || echo "{}")
        
        # Check if response contains expected fields
        if echo "$response" | jq -e '.naics_code' > /dev/null 2>&1; then
            print_success "Test $test_id: PASSED"
            local status="PASSED"
        else
            print_error "Test $test_id: FAILED"
            local status="FAILED"
        fi
        
        # Add result to results file
        local result=$(jq -n \
            --arg id "$test_id" \
            --arg name "$test_name" \
            --arg status "$status" \
            --arg response "$response" \
            '{id: $id, name: $name, status: $status, response: $response}')
        
        jq --argjson result "$result" '. += [$result]' "$results_file" > "${results_file}.tmp" && mv "${results_file}.tmp" "$results_file"
    done
    
    print_success "Business Classification Tests completed"
}

# Function to run performance tests
run_performance_tests() {
    print_status "Running Performance Tests..."
    
    local results_file="results/performance_results.json"
    
    # Test response times
    local total_time=0
    local count=0
    
    for i in {1..20}; do
        local start_time=$(date +%s%N)
        curl -s "http://localhost:8080/health" > /dev/null
        local end_time=$(date +%s%N)
        local duration=$(( (end_time - start_time) / 1000000 ))
        total_time=$((total_time + duration))
        count=$((count + 1))
        echo "Request $i: ${duration}ms"
    done
    
    local avg_time=$((total_time / count))
    
    # Create performance result
    local result=$(jq -n \
        --arg avg_time "$avg_time" \
        --arg total_requests "$count" \
        '{avg_response_time_ms: $avg_time, total_requests: $count}')
    
    echo "$result" > "$results_file"
    
    if [ $avg_time -lt 200 ]; then
        print_success "Performance Test: PASSED (avg: ${avg_time}ms)"
    else
        print_warning "Performance Test: SLOW (avg: ${avg_time}ms)"
    fi
}

# Function to generate test report
generate_test_report() {
    print_status "Generating UAT Test Report..."
    
    local report_file="results/uat_test_report.md"
    
    cat > "$report_file" << EOF
# KYB Platform - UAT Test Report
Generated: $(date)

## Test Results Summary

### Business Classification Tests
EOF
    
    # Add business classification results
    if [ -f "results/business_classification_results.json" ]; then
        local passed_count=$(jq '[.[] | select(.status == "PASSED")] | length' results/business_classification_results.json)
        local total_count=$(jq 'length' results/business_classification_results.json)
        echo "- Passed: $passed_count/$total_count" >> "$report_file"
    fi
    
    cat >> "$report_file" << EOF

### Performance Tests
EOF
    
    # Add performance results
    if [ -f "results/performance_results.json" ]; then
        local avg_time=$(jq -r '.avg_response_time_ms' results/performance_results.json)
        echo "- Average Response Time: ${avg_time}ms" >> "$report_file"
    fi
    
    cat >> "$report_file" << EOF

## Recommendations
1. Review failed test cases
2. Address performance issues if any
3. Implement missing functionality
4. Prepare for beta testing

EOF
    
    print_success "UAT Test Report generated: $report_file"
}

# Main execution
main() {
    echo "🧪 KYB Platform - UAT Test Execution"
    echo "===================================="
    echo
    
    # Check if application is running
    if ! curl -f -s "http://localhost:8080/health" > /dev/null 2>&1; then
        print_error "Application is not running. Please start the application first."
        exit 1
    fi
    
    # Run tests
    run_business_classification_tests
    echo
    run_performance_tests
    echo
    generate_test_report
    
    echo
    print_success "UAT Test Execution completed!"
}

# Run main function
main "$@"
EOF
    
    chmod +x test/uat/run_uat_tests.sh
    print_success "Test execution script created"
}

# Function to create UAT environment configuration
create_uat_configuration() {
    print_status "Creating UAT environment configuration..."
    
    cat > test/uat/uat_config.json << EOF
{
  "environment": "uat",
  "application_url": "http://localhost:8080",
  "timeout": 30,
  "retry_attempts": 3,
  "performance_thresholds": {
    "max_response_time_ms": 500,
    "max_concurrent_users": 100,
    "error_rate_threshold": 0.01
  },
  "test_data": {
    "business_classification": "data/business_classification_test_cases.json",
    "risk_assessment": "data/risk_assessment_test_cases.json",
    "compliance": "data/compliance_test_cases.json"
  },
  "scenarios": {
    "business_classification": "scenarios/uat_scenarios.md"
  },
  "results": {
    "output_directory": "results",
    "report_format": "markdown"
  }
}
EOF
    
    print_success "UAT configuration created"
}

# Main setup function
main_setup() {
    echo "🔧 KYB Platform - UAT Environment Setup"
    echo "======================================="
    echo
    
    # Check if application is running
    if ! check_application; then
        print_error "Cannot setup UAT environment - application is not running"
        exit 1
    fi
    
    echo
    print_status "Setting up UAT environment..."
    echo
    
    # Create UAT environment
    create_test_data_directory
    create_business_classification_test_data
    create_risk_assessment_test_data
    create_compliance_test_data
    create_uat_scenarios
    create_test_execution_script
    create_uat_configuration
    
    echo
    print_success "UAT environment setup completed!"
    echo
    print_status "UAT environment is ready for testing."
    print_status "Run './test/uat/run_uat_tests.sh' to execute UAT tests."
}

# Function to show usage
show_usage() {
    echo "KYB Platform - UAT Environment Setup Tool"
    echo "========================================="
    echo
    echo "Usage: $0 [COMMAND]"
    echo
    echo "Commands:"
    echo "  setup     - Set up complete UAT environment"
    echo "  data      - Create test data only"
    echo "  scenarios - Create UAT scenarios only"
    echo "  config    - Create UAT configuration only"
    echo "  help      - Show this help message"
    echo
}

# Main execution
main() {
    case "${1:-help}" in
        setup)
            main_setup
            ;;
        data)
            create_test_data_directory
            create_business_classification_test_data
            create_risk_assessment_test_data
            create_compliance_test_data
            ;;
        scenarios)
            create_uat_scenarios
            ;;
        config)
            create_uat_configuration
            ;;
        help|*)
            show_usage
            ;;
    esac
}

# Run main function
main "$@"
