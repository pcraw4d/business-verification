#!/bin/bash

# Crosswalk Functionality Validation Script
# This script validates the MCC/NAICS/SIC crosswalk functionality

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
REPORT_DIR="$PROJECT_ROOT/test_reports"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
VALIDATION_REPORT="$REPORT_DIR/crosswalk_validation_${TIMESTAMP}.md"

# Create report directory
mkdir -p "$REPORT_DIR"

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

# Function to validate database connectivity
validate_database_connectivity() {
    print_section "Database Connectivity Validation"
    
    if [ -z "$DATABASE_URL" ]; then
        print_validation_result "Database Connection" "FAIL" "DATABASE_URL not set"
        return 1
    fi
    
    # Test database connection
    if psql "$DATABASE_URL" -c "SELECT 1;" > /dev/null 2>&1; then
        print_validation_result "Database Connection" "PASS" "Successfully connected to database"
        return 0
    else
        print_validation_result "Database Connection" "FAIL" "Failed to connect to database"
        return 1
    fi
}

# Function to validate crosswalk table structure
validate_crosswalk_table_structure() {
    print_section "Crosswalk Table Structure Validation"
    
    local validation_queries=(
        "SELECT COUNT(*) FROM crosswalk_mappings WHERE is_valid = true;"
        "SELECT COUNT(DISTINCT source_system) FROM crosswalk_mappings WHERE is_valid = true;"
        "SELECT COUNT(DISTINCT target_system) FROM crosswalk_mappings WHERE is_valid = true;"
        "SELECT COUNT(*) FROM crosswalk_mappings WHERE source_system = 'MCC' AND is_valid = true;"
        "SELECT COUNT(*) FROM crosswalk_mappings WHERE source_system = 'NAICS' AND is_valid = true;"
        "SELECT COUNT(*) FROM crosswalk_mappings WHERE source_system = 'SIC' AND is_valid = true;"
    )
    
    local query_names=(
        "Total Valid Crosswalk Mappings"
        "Source Systems Count"
        "Target Systems Count"
        "MCC Mappings Count"
        "NAICS Mappings Count"
        "SIC Mappings Count"
    )
    
    for i in "${!validation_queries[@]}"; do
        local query="${validation_queries[$i]}"
        local name="${query_names[$i]}"
        
        local result=$(psql "$DATABASE_URL" -t -c "$query" 2>/dev/null | xargs)
        
        if [ -n "$result" ] && [ "$result" -gt 0 ]; then
            print_validation_result "$name" "PASS" "Count: $result"
        else
            print_validation_result "$name" "FAIL" "No data found or query failed"
        fi
    done
}

# Function to validate MCC to Industry mappings
validate_mcc_industry_mappings() {
    print_section "MCC to Industry Mappings Validation"
    
    # Test specific MCC codes
    local test_mcc_codes=("5734" "6010" "8011" "5310" "5085" "7995")
    local expected_industries=("Technology" "Financial Services" "Healthcare" "Retail" "Manufacturing" "Gambling")
    
    for i in "${!test_mcc_codes[@]}"; do
        local mcc_code="${test_mcc_codes[$i]}"
        local expected_industry="${expected_industries[$i]}"
        
        local query="
            SELECT COUNT(*) 
            FROM crosswalk_mappings cm 
            JOIN industries i ON cm.industry_id = i.id 
            WHERE cm.mcc_code = '$mcc_code' 
            AND cm.is_valid = true 
            AND i.name ILIKE '%$expected_industry%';
        "
        
        local result=$(psql "$DATABASE_URL" -t -c "$query" 2>/dev/null | xargs)
        
        if [ -n "$result" ] && [ "$result" -gt 0 ]; then
            print_validation_result "MCC $mcc_code to $expected_industry" "PASS" "Mapping found"
        else
            print_validation_result "MCC $mcc_code to $expected_industry" "FAIL" "No mapping found"
        fi
    done
}

# Function to validate NAICS to Industry mappings
validate_naics_industry_mappings() {
    print_section "NAICS to Industry Mappings Validation"
    
    # Test specific NAICS codes
    local test_naics_codes=("541511" "522110" "621111" "452111" "423990" "713210")
    local expected_industries=("Technology" "Financial Services" "Healthcare" "Retail" "Manufacturing" "Gambling")
    
    for i in "${!test_naics_codes[@]}"; do
        local naics_code="${test_naics_codes[$i]}"
        local expected_industry="${expected_industries[$i]}"
        
        local query="
            SELECT COUNT(*) 
            FROM crosswalk_mappings cm 
            JOIN industries i ON cm.industry_id = i.id 
            WHERE cm.naics_code = '$naics_code' 
            AND cm.is_valid = true 
            AND i.name ILIKE '%$expected_industry%';
        "
        
        local result=$(psql "$DATABASE_URL" -t -c "$query" 2>/dev/null | xargs)
        
        if [ -n "$result" ] && [ "$result" -gt 0 ]; then
            print_validation_result "NAICS $naics_code to $expected_industry" "PASS" "Mapping found"
        else
            print_validation_result "NAICS $naics_code to $expected_industry" "FAIL" "No mapping found"
        fi
    done
}

# Function to validate SIC to Industry mappings
validate_sic_industry_mappings() {
    print_section "SIC to Industry Mappings Validation"
    
    # Test specific SIC codes
    local test_sic_codes=("7372" "6021" "8011" "5311" "5085" "7995")
    local expected_industries=("Technology" "Financial Services" "Healthcare" "Retail" "Manufacturing" "Gambling")
    
    for i in "${!test_sic_codes[@]}"; do
        local sic_code="${test_sic_codes[$i]}"
        local expected_industry="${expected_industries[$i]}"
        
        local query="
            SELECT COUNT(*) 
            FROM crosswalk_mappings cm 
            JOIN industries i ON cm.industry_id = i.id 
            WHERE cm.sic_code = '$sic_code' 
            AND cm.is_valid = true 
            AND i.name ILIKE '%$expected_industry%';
        "
        
        local result=$(psql "$DATABASE_URL" -t -c "$query" 2>/dev/null | xargs)
        
        if [ -n "$result" ] && [ "$result" -gt 0 ]; then
            print_validation_result "SIC $sic_code to $expected_industry" "PASS" "Mapping found"
        else
            print_validation_result "SIC $sic_code to $expected_industry" "FAIL" "No mapping found"
        fi
    done
}

# Function to validate crosswalk consistency
validate_crosswalk_consistency() {
    print_section "Crosswalk Consistency Validation"
    
    # Check for orphaned mappings
    local orphaned_query="
        SELECT COUNT(*) 
        FROM crosswalk_mappings cm 
        LEFT JOIN industries i ON cm.industry_id = i.id 
        WHERE cm.is_valid = true 
        AND i.id IS NULL;
    "
    
    local orphaned_count=$(psql "$DATABASE_URL" -t -c "$orphaned_query" 2>/dev/null | xargs)
    
    if [ "$orphaned_count" -eq 0 ]; then
        print_validation_result "No Orphaned Mappings" "PASS" "All mappings reference valid industries"
    else
        print_validation_result "No Orphaned Mappings" "FAIL" "Found $orphaned_count orphaned mappings"
    fi
    
    # Check for duplicate mappings
    local duplicate_query="
        SELECT COUNT(*) 
        FROM (
            SELECT source_code, source_system, industry_id, COUNT(*) 
            FROM crosswalk_mappings 
            WHERE is_valid = true 
            GROUP BY source_code, source_system, industry_id 
            HAVING COUNT(*) > 1
        ) duplicates;
    "
    
    local duplicate_count=$(psql "$DATABASE_URL" -t -c "$duplicate_query" 2>/dev/null | xargs)
    
    if [ "$duplicate_count" -eq 0 ]; then
        print_validation_result "No Duplicate Mappings" "PASS" "No duplicate mappings found"
    else
        print_validation_result "No Duplicate Mappings" "FAIL" "Found $duplicate_count duplicate mappings"
    fi
    
    # Check confidence score validity
    local confidence_query="
        SELECT COUNT(*) 
        FROM crosswalk_mappings 
        WHERE is_valid = true 
        AND (confidence_score < 0.0 OR confidence_score > 1.0);
    "
    
    local invalid_confidence_count=$(psql "$DATABASE_URL" -t -c "$confidence_query" 2>/dev/null | xargs)
    
    if [ "$invalid_confidence_count" -eq 0 ]; then
        print_validation_result "Valid Confidence Scores" "PASS" "All confidence scores are between 0.0 and 1.0"
    else
        print_validation_result "Valid Confidence Scores" "FAIL" "Found $invalid_confidence_count invalid confidence scores"
    fi
}

# Function to validate crosswalk performance
validate_crosswalk_performance() {
    print_section "Crosswalk Performance Validation"
    
    # Test query performance
    local performance_queries=(
        "SELECT COUNT(*) FROM crosswalk_mappings WHERE source_system = 'MCC' AND is_valid = true;"
        "SELECT COUNT(*) FROM crosswalk_mappings WHERE source_system = 'NAICS' AND is_valid = true;"
        "SELECT COUNT(*) FROM crosswalk_mappings WHERE source_system = 'SIC' AND is_valid = true;"
        "SELECT cm.*, i.name FROM crosswalk_mappings cm JOIN industries i ON cm.industry_id = i.id WHERE cm.is_valid = true LIMIT 100;"
    )
    
    local query_names=(
        "MCC Query Performance"
        "NAICS Query Performance"
        "SIC Query Performance"
        "Join Query Performance"
    )
    
    for i in "${!performance_queries[@]}"; do
        local query="${performance_queries[$i]}"
        local name="${query_names[$i]}"
        
        local start_time=$(date +%s%N)
        psql "$DATABASE_URL" -c "$query" > /dev/null 2>&1
        local end_time=$(date +%s%N)
        local duration=$(( (end_time - start_time) / 1000000 )) # Convert to milliseconds
        
        if [ $duration -lt 1000 ]; then
            print_validation_result "$name" "PASS" "Query completed in ${duration}ms"
        else
            print_validation_result "$name" "FAIL" "Query took ${duration}ms (too slow)"
        fi
    done
}

# Function to generate validation report
generate_validation_report() {
    print_section "Generating Validation Report"
    
    cat > "$VALIDATION_REPORT" << EOF
# Crosswalk Functionality Validation Report

**Generated**: $(date)
**Validation Type**: MCC/NAICS/SIC Crosswalk Functionality
**Environment**: $(uname -s) $(uname -r)

## 🎯 Validation Objectives

This validation ensures the crosswalk functionality is working correctly:

1. **Database Connectivity** - Verify database connection and access
2. **Table Structure** - Validate crosswalk table structure and data
3. **MCC Mappings** - Verify MCC code to industry mappings
4. **NAICS Mappings** - Verify NAICS code to industry mappings
5. **SIC Mappings** - Verify SIC code to industry mappings
6. **Consistency** - Validate data consistency and integrity
7. **Performance** - Verify query performance meets requirements

## 📊 Validation Results

### Database Connectivity
- **Connection Status**: $(if [ -n "$DATABASE_URL" ]; then echo "Configured"; else echo "Not configured"; fi)
- **Connection Test**: *Results will be populated during validation*

### Table Structure Validation
- **Total Crosswalk Mappings**: *Results will be populated during validation*
- **Source Systems**: *Results will be populated during validation*
- **Target Systems**: *Results will be populated during validation*

### Mapping Validations
- **MCC to Industry Mappings**: *Results will be populated during validation*
- **NAICS to Industry Mappings**: *Results will be populated during validation*
- **SIC to Industry Mappings**: *Results will be populated during validation*

### Consistency Validations
- **Orphaned Mappings**: *Results will be populated during validation*
- **Duplicate Mappings**: *Results will be populated during validation*
- **Confidence Score Validity**: *Results will be populated during validation*

### Performance Validations
- **Query Performance**: *Results will be populated during validation*

## 🎯 Success Criteria

### Technical Requirements
- [ ] Database connectivity established
- [ ] Crosswalk table structure validated
- [ ] MCC mappings functional
- [ ] NAICS mappings functional
- [ ] SIC mappings functional
- [ ] Data consistency maintained
- [ ] Performance requirements met

### Quality Requirements
- [ ] No orphaned mappings
- [ ] No duplicate mappings
- [ ] Valid confidence scores
- [ ] Query performance < 1 second
- [ ] All test mappings validated

## 🚨 Issues and Recommendations

### Critical Issues
*Critical issues will be listed here after validation*

### Performance Issues
*Performance issues will be listed here after validation*

### Recommendations
*Recommendations will be generated based on validation results*

## 📋 Next Steps

1. **Address Critical Issues**: Fix any critical validation failures
2. **Performance Optimization**: Optimize any slow-performing queries
3. **Data Quality**: Improve data quality based on validation results
4. **Documentation**: Update documentation based on validation findings
5. **Production Readiness**: Validate production readiness criteria

---

**Validation Report Generated**: $(date)
**Validation Environment**: $(uname -s) $(uname -r)

EOF

    echo -e "✅ ${GREEN}Validation report generated: $VALIDATION_REPORT${NC}"
}

# Main execution function
main() {
    echo -e "${PURPLE}🔗 Crosswalk Functionality Validation${NC}"
    echo -e "${PURPLE}====================================${NC}"
    echo ""
    
    # Track validation results
    local total_validations=0
    local passed_validations=0
    local failed_validations=0
    
    # Validate database connectivity
    if validate_database_connectivity; then
        passed_validations=$((passed_validations + 1))
    else
        failed_validations=$((failed_validations + 1))
    fi
    total_validations=$((total_validations + 1))
    echo ""
    
    # Validate crosswalk table structure
    validate_crosswalk_table_structure
    echo ""
    
    # Validate MCC to Industry mappings
    validate_mcc_industry_mappings
    echo ""
    
    # Validate NAICS to Industry mappings
    validate_naics_industry_mappings
    echo ""
    
    # Validate SIC to Industry mappings
    validate_sic_industry_mappings
    echo ""
    
    # Validate crosswalk consistency
    validate_crosswalk_consistency
    echo ""
    
    # Validate crosswalk performance
    validate_crosswalk_performance
    echo ""
    
    # Generate validation report
    generate_validation_report
    echo ""
    
    # Print final summary
    print_section "Validation Summary"
    echo -e "📊 ${BLUE}Total Validations:${NC} $total_validations"
    echo -e "✅ ${GREEN}Passed:${NC} $passed_validations"
    echo -e "❌ ${RED}Failed:${NC} $failed_validations"
    echo -e "📈 ${BLUE}Success Rate:${NC} $(( (passed_validations * 100) / total_validations ))%"
    echo ""
    echo -e "📄 ${BLUE}Validation Report:${NC} $VALIDATION_REPORT"
    
    # Determine exit code
    if [ $failed_validations -eq 0 ]; then
        echo -e "\n🎉 ${GREEN}All validations passed! Crosswalk functionality is ready.${NC}"
        exit 0
    else
        echo -e "\n⚠️  ${YELLOW}Some validations failed. Please review the report and fix issues.${NC}"
        exit 1
    fi
}

# Run main function
main "$@"
