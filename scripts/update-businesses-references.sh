#!/bin/bash

# Update Businesses References Script
# Subtask 2.2.4: Update all references to businesses table
# Date: January 19, 2025
# Purpose: Update all code references from businesses table to merchants table

set -e  # Exit on any error

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
UPDATE_LOG="$PROJECT_ROOT/logs/businesses-references-update-$(date +%Y%m%d-%H%M%S).log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
FILES_UPDATED=0
REFERENCES_UPDATED=0

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$UPDATE_LOG"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$UPDATE_LOG"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$UPDATE_LOG"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$UPDATE_LOG"
}

# Create logs directory if it doesn't exist
mkdir -p "$(dirname "$UPDATE_LOG")"

log_info "Starting businesses table references update"
log_info "Update log: $UPDATE_LOG"

# Function to update references in a file
update_file_references() {
    local file="$1"
    local temp_file="${file}.tmp"
    local changes_made=0
    
    # Create backup
    cp "$file" "${file}.backup"
    
    # Update references
    sed -e 's/table.*businesses/table merchants/g' \
        -e 's/from businesses/from merchants/g' \
        -e 's/into businesses/into merchants/g' \
        -e 's/update businesses/update merchants/g' \
        -e 's/delete from businesses/delete from merchants/g' \
        -e 's/businesses\./merchants\./g' \
        -e 's/businesses table/merchants table/g' \
        -e 's/businesses_table/merchants_table/g' \
        -e 's/businessesTable/merchantsTable/g' \
        -e 's/businessesTableName/merchantsTableName/g' \
        -e 's/businesses\.id/merchants.id/g' \
        -e 's/businesses\.name/merchants.name/g' \
        -e 's/businesses\.created_at/merchants.created_at/g' \
        -e 's/businesses\.updated_at/merchants.updated_at/g' \
        "$file" > "$temp_file"
    
    # Check if changes were made
    if ! cmp -s "$file" "$temp_file"; then
        mv "$temp_file" "$file"
        changes_made=1
        ((FILES_UPDATED++))
        
        # Count the number of changes
        local change_count=$(diff "$file.backup" "$file" | grep "^<" | wc -l)
        ((REFERENCES_UPDATED += change_count))
        
        log_success "Updated $file ($change_count changes)"
    else
        rm "$temp_file"
    fi
    
    return $changes_made
}

# Function to update Go struct references
update_go_struct_references() {
    local file="$1"
    local temp_file="${file}.tmp"
    local changes_made=0
    
    # Create backup
    cp "$file" "${file}.backup"
    
    # Update Go-specific references
    sed -e 's/type Business struct/type Merchant struct/g' \
        -e 's/\*Business/\*Merchant/g' \
        -e 's/\[\]Business/\[\]Merchant/g' \
        -e 's/map\[string\]Business/map[string]Merchant/g' \
        -e 's/business\./merchant\./g' \
        -e 's/businesss\./merchants\./g' \
        -e 's/businessID/merchantID/g' \
        -e 's/businessId/merchantId/g' \
        -e 's/business_id/merchant_id/g' \
        -e 's/businessName/merchantName/g' \
        -e 's/businessName/merchantName/g' \
        -e 's/business_name/merchant_name/g' \
        -e 's/businessData/merchantData/g' \
        -e 's/businessData/merchantData/g' \
        -e 's/business_data/merchant_data/g' \
        -e 's/businessInfo/merchantInfo/g' \
        -e 's/businessInfo/merchantInfo/g' \
        -e 's/business_info/merchant_info/g' \
        -e 's/businessService/merchantService/g' \
        -e 's/businessService/merchantService/g' \
        -e 's/business_service/merchant_service/g' \
        -e 's/businessHandler/merchantHandler/g' \
        -e 's/businessHandler/merchantHandler/g' \
        -e 's/business_handler/merchant_handler/g' \
        -e 's/businessRepository/merchantRepository/g' \
        -e 's/businessRepository/merchantRepository/g' \
        -e 's/business_repository/merchant_repository/g' \
        -e 's/businessModel/merchantModel/g' \
        -e 's/businessModel/merchantModel/g' \
        -e 's/business_model/merchant_model/g' \
        -e 's/businessQuery/merchantQuery/g' \
        -e 's/businessQuery/merchantQuery/g' \
        -e 's/business_query/merchant_query/g' \
        -e 's/businessResult/merchantResult/g' \
        -e 's/businessResult/merchantResult/g' \
        -e 's/business_result/merchant_result/g' \
        -e 's/businessResponse/merchantResponse/g' \
        -e 's/businessResponse/merchantResponse/g' \
        -e 's/business_response/merchant_response/g' \
        -e 's/businessRequest/merchantRequest/g' \
        -e 's/businessRequest/merchantRequest/g' \
        -e 's/business_request/merchant_request/g' \
        -e 's/businessList/merchantList/g' \
        -e 's/businessList/merchantList/g' \
        -e 's/business_list/merchant_list/g' \
        -e 's/businessCount/merchantCount/g' \
        -e 's/businessCount/merchantCount/g' \
        -e 's/business_count/merchant_count/g' \
        -e 's/businessSearch/merchantSearch/g' \
        -e 's/businessSearch/merchantSearch/g' \
        -e 's/business_search/merchant_search/g' \
        -e 's/businessFilter/merchantFilter/g' \
        -e 's/businessFilter/merchantFilter/g' \
        -e 's/business_filter/merchant_filter/g' \
        -e 's/businessSort/merchantSort/g' \
        -e 's/businessSort/merchantSort/g' \
        -e 's/business_sort/merchant_sort/g' \
        -e 's/businessPagination/merchantPagination/g' \
        -e 's/businessPagination/merchantPagination/g' \
        -e 's/business_pagination/merchant_pagination/g' \
        -e 's/businessValidation/merchantValidation/g' \
        -e 's/businessValidation/merchantValidation/g' \
        -e 's/business_validation/merchant_validation/g' \
        -e 's/businessError/merchantError/g' \
        -e 's/businessError/merchantError/g' \
        -e 's/business_error/merchant_error/g' \
        -e 's/businessSuccess/merchantSuccess/g' \
        -e 's/businessSuccess/merchantSuccess/g' \
        -e 's/business_success/merchant_success/g' \
        -e 's/businessStatus/merchantStatus/g' \
        -e 's/businessStatus/merchantStatus/g' \
        -e 's/business_status/merchant_status/g' \
        -e 's/businessType/merchantType/g' \
        -e 's/businessType/merchantType/g' \
        -e 's/business_type/merchant_type/g' \
        -e 's/businessCategory/merchantCategory/g' \
        -e 's/businessCategory/merchantCategory/g' \
        -e 's/business_category/merchant_category/g' \
        -e 's/businessIndustry/merchantIndustry/g' \
        -e 's/businessIndustry/merchantIndustry/g' \
        -e 's/business_industry/merchant_industry/g' \
        -e 's/businessAddress/merchantAddress/g' \
        -e 's/businessAddress/merchantAddress/g' \
        -e 's/business_address/merchant_address/g' \
        -e 's/businessContact/merchantContact/g' \
        -e 's/businessContact/merchantContact/g' \
        -e 's/business_contact/merchant_contact/g' \
        -e 's/businessPhone/merchantPhone/g' \
        -e 's/businessPhone/merchantPhone/g' \
        -e 's/business_phone/merchant_phone/g' \
        -e 's/businessEmail/merchantEmail/g' \
        -e 's/businessEmail/merchantEmail/g' \
        -e 's/business_email/merchant_email/g' \
        -e 's/businessWebsite/merchantWebsite/g' \
        -e 's/businessWebsite/merchantWebsite/g' \
        -e 's/business_website/merchant_website/g' \
        -e 's/businessDescription/merchantDescription/g' \
        -e 's/businessDescription/merchantDescription/g' \
        -e 's/business_description/merchant_description/g' \
        -e 's/businessMetadata/merchantMetadata/g' \
        -e 's/businessMetadata/merchantMetadata/g' \
        -e 's/business_metadata/merchant_metadata/g' \
        -e 's/businessCreatedAt/merchantCreatedAt/g' \
        -e 's/businessCreatedAt/merchantCreatedAt/g' \
        -e 's/business_created_at/merchant_created_at/g' \
        -e 's/businessUpdatedAt/merchantUpdatedAt/g' \
        -e 's/businessUpdatedAt/merchantUpdatedAt/g' \
        -e 's/business_updated_at/merchant_updated_at/g' \
        "$file" > "$temp_file"
    
    # Check if changes were made
    if ! cmp -s "$file" "$temp_file"; then
        mv "$temp_file" "$file"
        changes_made=1
        ((FILES_UPDATED++))
        
        # Count the number of changes
        local change_count=$(diff "$file.backup" "$file" | grep "^<" | wc -l)
        ((REFERENCES_UPDATED += change_count))
        
        log_success "Updated Go struct references in $file ($change_count changes)"
    else
        rm "$temp_file"
    fi
    
    return $changes_made
}

# Function to update SQL references
update_sql_references() {
    local file="$1"
    local temp_file="${file}.tmp"
    local changes_made=0
    
    # Create backup
    cp "$file" "${file}.backup"
    
    # Update SQL-specific references
    sed -e 's/CREATE TABLE businesses/CREATE TABLE merchants/g' \
        -e 's/DROP TABLE businesses/DROP TABLE merchants/g' \
        -e 's/ALTER TABLE businesses/ALTER TABLE merchants/g' \
        -e 's/INSERT INTO businesses/INSERT INTO merchants/g' \
        -e 's/UPDATE businesses/UPDATE merchants/g' \
        -s/DELETE FROM businesses/DELETE FROM merchants/g' \
        -e 's/FROM businesses/FROM merchants/g' \
        -e 's/JOIN businesses/JOIN merchants/g' \
        -e 's/LEFT JOIN businesses/LEFT JOIN merchants/g' \
        -e 's/RIGHT JOIN businesses/RIGHT JOIN merchants/g' \
        -e 's/INNER JOIN businesses/INNER JOIN merchants/g' \
        -e 's/businesses\./merchants\./g' \
        -e 's/businesses table/merchants table/g' \
        -e 's/businesses_table/merchants_table/g' \
        -e 's/businessesTable/merchantsTable/g' \
        -e 's/businessesTableName/merchantsTableName/g' \
        -e 's/businesses\.id/merchants.id/g' \
        -e 's/businesses\.name/merchants.name/g' \
        -e 's/businesses\.created_at/merchants.created_at/g' \
        -e 's/businesses\.updated_at/merchants.updated_at/g' \
        -e 's/businesses\.user_id/merchants.user_id/g' \
        -e 's/businesses\.industry/merchants.industry/g' \
        -e 's/businesses\.industry_code/merchants.industry_code/g' \
        -e 's/businesses\.website_url/merchants.website_url/g' \
        -e 's/businesses\.description/merchants.description/g' \
        -e 's/businesses\.metadata/merchants.metadata/g' \
        -e 's/businesses\.address/merchants.address/g' \
        -e 's/businesses\.contact_info/merchants.contact_info/g' \
        -e 's/businesses\.founded_date/merchants.founded_date/g' \
        -e 's/businesses\.employee_count/merchants.employee_count/g' \
        -e 's/businesses\.annual_revenue/merchants.annual_revenue/g' \
        -e 's/businesses\.country_code/merchants.country_code/g' \
        -e 's/businesses\.registration_number/merchants.registration_number/g' \
        -e 's/businesses\.portfolio_type_id/merchants.portfolio_type_id/g' \
        -e 's/businesses\.risk_level_id/merchants.risk_level_id/g' \
        -e 's/businesses\.compliance_status/merchants.compliance_status/g' \
        -e 's/businesses\.status/merchants.status/g' \
        -e 's/businesses\.created_by/merchants.created_by/g' \
        -e 's/businesses\.legal_name/merchants.legal_name/g' \
        -e 's/businesses\.tax_id/merchants.tax_id/g' \
        -e 's/businesses\.business_type/merchants.business_type/g' \
        -e 's/businesses\.address_street1/merchants.address_street1/g' \
        -e 's/businesses\.address_street2/merchants.address_street2/g' \
        -e 's/businesses\.address_city/merchants.address_city/g' \
        -e 's/businesses\.address_state/merchants.address_state/g' \
        -e 's/businesses\.address_postal_code/merchants.address_postal_code/g' \
        -e 's/businesses\.address_country/merchants.address_country/g' \
        -e 's/businesses\.address_country_code/merchants.address_country_code/g' \
        -e 's/businesses\.contact_phone/merchants.contact_phone/g' \
        -e 's/businesses\.contact_email/merchants.contact_email/g' \
        -e 's/businesses\.contact_website/merchants.contact_website/g' \
        -e 's/businesses\.contact_primary_contact/merchants.contact_primary_contact/g' \
        "$file" > "$temp_file"
    
    # Check if changes were made
    if ! cmp -s "$file" "$temp_file"; then
        mv "$temp_file" "$file"
        changes_made=1
        ((FILES_UPDATED++))
        
        # Count the number of changes
        local change_count=$(diff "$file.backup" "$file" | grep "^<" | wc -l)
        ((REFERENCES_UPDATED += change_count))
        
        log_success "Updated SQL references in $file ($change_count changes)"
    else
        rm "$temp_file"
    fi
    
    return $changes_made
}

# Function to update documentation references
update_doc_references() {
    local file="$1"
    local temp_file="${file}.tmp"
    local changes_made=0
    
    # Create backup
    cp "$file" "${file}.backup"
    
    # Update documentation-specific references
    sed -e 's/businesses table/merchants table/g' \
        -e 's/businesses_table/merchants_table/g' \
        -e 's/businessesTable/merchantsTable/g' \
        -e 's/businesses\./merchants\./g' \
        -e 's/businesses\.id/merchants.id/g' \
        -e 's/businesses\.name/merchants.name/g' \
        -e 's/businesses\.created_at/merchants.created_at/g' \
        -e 's/businesses\.updated_at/merchants.updated_at/g' \
        -e 's/businesses\.user_id/merchants.user_id/g' \
        -e 's/businesses\.industry/merchants.industry/g' \
        -e 's/businesses\.industry_code/merchants.industry_code/g' \
        -e 's/businesses\.website_url/merchants.website_url/g' \
        -e 's/businesses\.description/merchants.description/g' \
        -e 's/businesses\.metadata/merchants.metadata/g' \
        -e 's/businesses\.address/merchants.address/g' \
        -e 's/businesses\.contact_info/merchants.contact_info/g' \
        -e 's/businesses\.founded_date/merchants.founded_date/g' \
        -e 's/businesses\.employee_count/merchants.employee_count/g' \
        -e 's/businesses\.annual_revenue/merchants.annual_revenue/g' \
        -e 's/businesses\.country_code/merchants.country_code/g' \
        -e 's/businesses\.registration_number/merchants.registration_number/g' \
        -e 's/businesses\.portfolio_type_id/merchants.portfolio_type_id/g' \
        -e 's/businesses\.risk_level_id/merchants.risk_level_id/g' \
        -e 's/businesses\.compliance_status/merchants.compliance_status/g' \
        -e 's/businesses\.status/merchants.status/g' \
        -e 's/businesses\.created_by/merchants.created_by/g' \
        -e 's/businesses\.legal_name/merchants.legal_name/g' \
        -e 's/businesses\.tax_id/merchants.tax_id/g' \
        -e 's/businesses\.business_type/merchants.business_type/g' \
        -e 's/businesses\.address_street1/merchants.address_street1/g' \
        -e 's/businesses\.address_street2/merchants.address_street2/g' \
        -e 's/businesses\.address_city/merchants.address_city/g' \
        -e 's/businesses\.address_state/merchants.address_state/g' \
        -e 's/businesses\.address_postal_code/merchants.address_postal_code/g' \
        -e 's/businesses\.address_country/merchants.address_country/g' \
        -e 's/businesses\.address_country_code/merchants.address_country_code/g' \
        -e 's/businesses\.contact_phone/merchants.contact_phone/g' \
        -e 's/businesses\.contact_email/merchants.contact_email/g' \
        -e 's/businesses\.contact_website/merchants.contact_website/g' \
        -e 's/businesses\.contact_primary_contact/merchants.contact_primary_contact/g' \
        "$file" > "$temp_file"
    
    # Check if changes were made
    if ! cmp -s "$file" "$temp_file"; then
        mv "$temp_file" "$file"
        changes_made=1
        ((FILES_UPDATED++))
        
        # Count the number of changes
        local change_count=$(diff "$file.backup" "$file" | grep "^<" | wc -l)
        ((REFERENCES_UPDATED += change_count))
        
        log_success "Updated documentation references in $file ($change_count changes)"
    else
        rm "$temp_file"
    fi
    
    return $changes_made
}

# Main update process
log_info "=== STARTING BUSINESSES REFERENCES UPDATE ==="

# Update Go files
log_info "Updating Go files..."
find "$PROJECT_ROOT" -name "*.go" -type f -exec grep -l "businesses" {} \; 2>/dev/null | while read -r file; do
    if update_go_struct_references "$file"; then
        log_info "Updated Go file: $file"
    fi
done

# Update SQL files
log_info "Updating SQL files..."
find "$PROJECT_ROOT" -name "*.sql" -type f -exec grep -l "businesses" {} \; 2>/dev/null | while read -r file; do
    if update_sql_references "$file"; then
        log_info "Updated SQL file: $file"
    fi
done

# Update documentation files
log_info "Updating documentation files..."
find "$PROJECT_ROOT" -name "*.md" -type f -exec grep -l "businesses" {} \; 2>/dev/null | while read -r file; do
    if update_doc_references "$file"; then
        log_info "Updated documentation file: $file"
    fi
done

# Update shell scripts
log_info "Updating shell scripts..."
find "$PROJECT_ROOT" -name "*.sh" -type f -exec grep -l "businesses" {} \; 2>/dev/null | while read -r file; do
    if update_file_references "$file"; then
        log_info "Updated shell script: $file"
    fi
done

# Update YAML files
log_info "Updating YAML files..."
find "$PROJECT_ROOT" -name "*.yaml" -o -name "*.yml" -type f -exec grep -l "businesses" {} \; 2>/dev/null | while read -r file; do
    if update_file_references "$file"; then
        log_info "Updated YAML file: $file"
    fi
done

# Update JSON files
log_info "Updating JSON files..."
find "$PROJECT_ROOT" -name "*.json" -type f -exec grep -l "businesses" {} \; 2>/dev/null | while read -r file; do
    if update_file_references "$file"; then
        log_info "Updated JSON file: $file"
    fi
done

# Final summary
log_info "=== UPDATE SUMMARY ==="
log_success "References update completed!"
log_info "Files updated: $FILES_UPDATED"
log_info "Total references updated: $REFERENCES_UPDATED"
log_info "Update log saved to: $UPDATE_LOG"

# Create summary report
cat >> "$UPDATE_LOG" << EOF

=== BUSINESSES REFERENCES UPDATE SUMMARY ===
Date: $(date)
Action: Updated all references from businesses table to merchants table
Files updated: $FILES_UPDATED
Total references updated: $REFERENCES_UPDATED

=== FILES UPDATED ===
$(find "$PROJECT_ROOT" -name "*.backup" -type f | sed 's/\.backup$//' | sort)

=== NEXT STEPS ===
1. Review all updated files for accuracy
2. Test the application to ensure functionality is preserved
3. Remove backup files after verification
4. Update any remaining manual references

EOF

log_info "Next steps:"
log_info "1. Review all updated files for accuracy"
log_info "2. Test the application to ensure functionality is preserved"
log_info "3. Remove backup files after verification"
log_info "4. Update any remaining manual references"

log_success "Businesses references update completed successfully!"

exit 0
