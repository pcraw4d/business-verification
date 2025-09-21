# Data Type and Format Validation Guide

## Overview

This guide provides comprehensive instructions for validating data types and formats in the KYB Platform database. The validation system ensures that all data stored in the database conforms to expected types and formats, maintaining data integrity and consistency.

## 🎯 Purpose

Data type and format validation ensures:
- **Data Integrity**: All data conforms to expected types and formats
- **Format Consistency**: Standardized formats for emails, UUIDs, phone numbers, URLs, etc.
- **Constraint Compliance**: Data respects length limits and other constraints
- **Quality Assurance**: High-quality data for business operations and analytics

## 🛠️ Validation Tools

### 1. Go-Based Validation Tool (`validate-data-types.go`)

**Features:**
- Comprehensive column discovery and analysis
- Automated format validation using regex patterns
- Length constraint validation for varchar columns
- Performance timing for each validation
- Detailed reporting with pass/fail status
- Sample invalid values for debugging

**Usage:**
```bash
# Using environment variable
export DATABASE_URL="postgresql://user:pass@localhost:5432/dbname"
go run scripts/validate-data-types.go

# Using command line argument
go run scripts/validate-data-types.go "postgresql://user:pass@localhost:5432/dbname"
```

### 2. SQL-Based Validation Tool (`validate-data-types.sql`)

**Features:**
- Static SQL queries for comprehensive data type analysis
- Format validation for specific data types
- Length constraint checks
- NULL constraint validation
- Data type consistency analysis
- Comprehensive reporting

**Usage:**
```bash
psql "postgresql://user:pass@localhost:5432/dbname" -f scripts/validate-data-types.sql
```

### 3. Automated Test Runner (`run-data-type-tests.sh`)

**Features:**
- Runs both Go and SQL tests automatically
- Generates comprehensive reports
- Timestamped output files
- Colored console output
- Error handling and validation

**Usage:**
```bash
# Using environment variable
export DATABASE_URL="postgresql://user:pass@localhost:5432/dbname"
./scripts/run-data-type-tests.sh

# Using command line argument
./scripts/run-data-type-tests.sh "postgresql://user:pass@localhost:5432/dbname"
```

## 📊 Validation Coverage

### Data Types Validated

1. **Email Addresses**
   - Pattern: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
   - Columns: `email`, `contact_email`, `notification_email`

2. **UUIDs**
   - Pattern: `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
   - Columns: `id`, `uuid`, `user_id`, `merchant_id`

3. **Phone Numbers**
   - Pattern: `^\+?[1-9]\d{1,14}$` (E.164 format)
   - Columns: `phone`, `mobile`, `contact_phone`

4. **URLs**
   - Pattern: `^https?://[^\s/$.?#].[^\s]*$`
   - Columns: `website`, `url`, `homepage`

5. **Dates and Timestamps**
   - Date Pattern: `^\d{4}-\d{2}-\d{2}$`
   - Timestamp Pattern: `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`
   - Columns: `created_at`, `updated_at`, `date`, `timestamp`

6. **String Length Constraints**
   - Validates varchar columns against maximum length limits
   - Identifies oversized strings

7. **Numeric Ranges**
   - Validates integer, decimal, and numeric columns
   - Checks for valid numeric values

8. **Boolean Values**
   - Validates boolean columns for valid true/false values

9. **JSON Format**
   - Validates JSON and JSONB columns for valid JSON syntax

10. **NULL Constraints**
    - Checks for NULL values in non-nullable columns

### Test Results

The validation system provides clear, actionable results:
- ✅ **PASS**: All values are valid
- ❌ **FAIL**: Invalid values detected with counts and samples
- 🚨 **ERROR**: Validation execution errors

### Sample Output

```
🔍 Starting Data Type and Format Validation...
============================================================
Found 45 columns to validate

[1/45] Validating users.email (character varying)
  ✅ PASS - All values are valid
  ⏱️  Execution time: 23ms

[2/45] Validating merchants.phone (character varying)
  ❌ FAIL - Found 3 invalid values out of 1,247 total
  📝 Sample invalid values: "123-456-7890", "555.123.4567", "invalid-phone"
  ⏱️  Execution time: 45ms

============================================================
📊 DATA TYPE VALIDATION SUMMARY
============================================================
Total Tests: 45
✅ Passed: 42
❌ Failed: 3
🚨 Errors: 0
Success Rate: 93.3%
```

## 🔍 Validation Patterns

### Email Validation
```regex
^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$
```
- Allows alphanumeric characters, dots, underscores, percent signs, plus signs, and hyphens
- Requires @ symbol and valid domain format
- Ensures proper TLD format

### UUID Validation
```regex
^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$
```
- Validates standard UUID v4 format
- Ensures proper hyphen placement
- Case-insensitive hexadecimal characters

### Phone Number Validation (E.164)
```regex
^\+?[1-9]\d{1,14}$
```
- Optional + prefix
- First digit cannot be 0
- Maximum 15 digits total
- International format compliance

### URL Validation
```regex
^https?://[^\s/$.?#].[^\s]*$
```
- Requires http:// or https:// protocol
- Validates domain format
- Prevents common URL injection patterns

## 🚨 Common Issues and Solutions

### Issue 1: Invalid Email Formats

**Problem:** Email addresses don't match expected format

**Example:**
```sql
-- Find invalid email addresses
SELECT email FROM users 
WHERE email !~ '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$';
```

**Solution:**
1. Identify invalid email addresses
2. Update with valid formats or remove invalid entries
3. Implement email validation in application layer
4. Re-run validation to verify fixes

### Issue 2: Invalid Phone Number Formats

**Problem:** Phone numbers don't follow E.164 format

**Example:**
```sql
-- Find invalid phone numbers
SELECT phone FROM merchants 
WHERE phone !~ '^\+?[1-9]\d{1,14}$';
```

**Solution:**
```sql
-- Update phone numbers to E.164 format
UPDATE merchants 
SET phone = '+1' || REGEXP_REPLACE(phone, '[^0-9]', '', 'g')
WHERE phone !~ '^\+?[1-9]\d{1,14}$';
```

### Issue 3: Oversized String Values

**Problem:** String values exceed column length limits

**Example:**
```sql
-- Find oversized strings
SELECT name, LENGTH(name) as name_length 
FROM merchants 
WHERE LENGTH(name) > 255;
```

**Solution:**
```sql
-- Truncate oversized strings
UPDATE merchants 
SET name = LEFT(name, 255)
WHERE LENGTH(name) > 255;
```

### Issue 4: NULL Values in Non-Nullable Columns

**Problem:** NULL values found in columns marked as NOT NULL

**Example:**
```sql
-- Find NULL values in non-nullable columns
SELECT COUNT(*) FROM users WHERE email IS NULL;
```

**Solution:**
1. Identify the source of NULL values
2. Either update with valid values or alter column to allow NULLs
3. Implement application-level validation to prevent future NULLs

## 🔧 Configuration

### Environment Variables

- **`DATABASE_URL`** - PostgreSQL connection string
  - Format: `postgresql://user:password@host:port/database`
  - Example: `postgresql://kyb_user:password@localhost:5432/kyb_platform`

### Database Permissions

The validation tools require the following permissions:
- `SELECT` on all tables being validated
- `SELECT` on `information_schema` tables

### Performance Considerations

- **Large Tables**: Validation may take longer on tables with millions of records
- **Pattern Matching**: Regex validation can be CPU-intensive on large datasets
- **Index Usage**: Ensure columns being validated are indexed for optimal performance

## 📈 Monitoring and Alerting

### Integration with Monitoring Systems

The validation results can be integrated with monitoring systems:

```bash
# Run validation and check exit code
./scripts/run-data-type-tests.sh "$DATABASE_URL"
if [ $? -ne 0 ]; then
    # Send alert to monitoring system
    curl -X POST "https://monitoring.example.com/alerts" \
        -d '{"type": "data_type_violation", "severity": "medium"}'
fi
```

### Automated Validation

Set up automated validation in CI/CD pipelines:

```yaml
# .github/workflows/data-type-validation.yml
name: Data Type Validation
on:
  schedule:
    - cron: '0 3 * * *'  # Daily at 3 AM
  push:
    branches: [main]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Run Data Type Validation
        run: ./scripts/run-data-type-tests.sh "${{ secrets.DATABASE_URL }}"
```

## 🎯 Best Practices

### 1. Regular Validation

- Run validation daily in production
- Run validation after any data imports or migrations
- Run validation before major deployments

### 2. Application-Level Validation

- Implement client-side validation for user inputs
- Add server-side validation in API endpoints
- Use database constraints as the final validation layer

### 3. Data Quality Monitoring

- Monitor validation results over time
- Set up alerts for validation failures
- Track data quality metrics

### 4. Performance Optimization

- Index columns that are frequently validated
- Use batch processing for large datasets
- Consider validation during off-peak hours

## 🔗 Related Documentation

- [Foreign Key Constraint Testing Guide](../docs/foreign-key-testing-guide.md)
- [Database Schema Documentation](../docs/database-schema.md)
- [Data Integrity Validation Guide](../docs/data-integrity-validation.md)
- [Performance Optimization Guide](../docs/performance-optimization.md)

## 📞 Support

For issues or questions regarding data type validation:

1. Check the validation output logs for specific error messages
2. Review this guide for common solutions
3. Consult the database schema documentation
4. Contact the development team for complex issues

---

**Last Updated:** January 19, 2025  
**Version:** 1.0  
**Next Review:** February 19, 2025
