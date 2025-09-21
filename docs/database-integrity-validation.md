# Database Integrity Validation System

## Overview

The Database Integrity Validation System is a comprehensive, modular solution for validating the integrity of the KYB Platform Supabase database. It provides automated checks for data consistency, referential integrity, and schema validation to ensure the database maintains high quality and reliability.

## Features

### Core Validation Checks

1. **Foreign Key Constraints** - Validates referential integrity across all tables
2. **Data Types and Formats** - Ensures data conforms to expected types and formats
3. **Orphaned Records** - Identifies records that violate referential integrity
4. **Data Consistency** - Validates business rules and data consistency across tables
5. **Table Structure** - Verifies table schemas and column definitions
6. **Indexes** - Validates index structure and identifies missing or duplicate indexes
7. **Constraints** - Validates database constraints and identifies violations

### Key Benefits

- **Comprehensive Coverage**: Validates all aspects of database integrity
- **Modular Design**: Each validation check is independent and can be run separately
- **Detailed Reporting**: Provides comprehensive reports with actionable recommendations
- **Performance Optimized**: Uses efficient queries and batch processing
- **Extensible**: Easy to add new validation checks
- **Professional Quality**: Follows Go best practices and clean architecture principles

## Architecture

### Package Structure

```
internal/database/integrity/
├── validator.go                    # Main validator and orchestration
├── foreign_key_validator.go        # Foreign key constraint validation
├── data_type_validator.go          # Data type and format validation
├── orphaned_records_validator.go   # Orphaned records validation
├── data_consistency_validator.go   # Data consistency validation
├── table_structure_validator.go    # Table structure validation
├── index_validator.go              # Index validation
├── constraint_validator.go         # Constraint validation
└── validator_test.go               # Comprehensive test suite
```

### Design Principles

1. **Interface-Driven**: All validators implement the `ValidationCheck` interface
2. **Dependency Injection**: Validators receive dependencies through constructor
3. **Error Handling**: Comprehensive error handling with detailed error messages
4. **Logging**: Structured logging for debugging and monitoring
5. **Configuration**: Flexible configuration for different validation scenarios
6. **Testing**: Comprehensive test coverage with integration tests

## Usage

### Command Line Tool

The system includes a command-line tool for running validation checks:

```bash
# Basic usage
./validate-db -db-url "postgres://user:pass@host:port/dbname"

# With custom output file
./validate-db -db-url "postgres://user:pass@host:port/dbname" -output "validation-report.json"

# With verbose logging
./validate-db -db-url "postgres://user:pass@host:port/dbname" -verbose

# With custom timeout
./validate-db -db-url "postgres://user:pass@host:port/dbname" -timeout 1h

# Run specific checks only
./validate-db -db-url "postgres://user:pass@host:port/dbname" -checks "foreign_keys,data_types"
```

### Programmatic Usage

```go
package main

import (
    "context"
    "database/sql"
    "log"
    
    _ "github.com/lib/pq"
    "github.com/company/kyb-platform/internal/database/integrity"
)

func main() {
    // Connect to database
    db, err := sql.Open("postgres", "your-connection-string")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Create validator
    logger := log.New(os.Stdout, "[VALIDATOR] ", log.LstdFlags)
    config := &integrity.ValidationConfig{
        CheckForeignKeys:     true,
        CheckDataTypes:       true,
        CheckOrphanedRecords: true,
        CheckDataConsistency: true,
        BatchSize:           1000,
        Timeout:             30 * time.Minute,
        ParallelValidation:  true,
        DetailedReporting:   true,
        IncludeStatistics:   true,
    }
    
    validator := integrity.NewValidator(db, logger, config)
    
    // Run validation
    ctx := context.Background()
    report, err := validator.ValidateAll(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    // Process results
    if report.Summary.FailedChecks > 0 {
        log.Printf("Validation failed with %d failed checks", report.Summary.FailedChecks)
    } else {
        log.Printf("Validation passed successfully")
    }
}
```

## Validation Checks

### 1. Foreign Key Constraints

**Purpose**: Validates referential integrity across all tables

**What it checks**:
- All foreign key constraints are properly defined
- No orphaned records exist in referencing tables
- Referenced records exist in target tables

**Key relationships validated**:
- `api_keys.user_id` → `users.id`
- `businesses.user_id` → `users.id`
- `business_classifications.business_id` → `businesses.id`
- `industry_keywords.industry_id` → `industries.id`
- `risk_assessments.business_id` → `businesses.id`
- And many more...

**Example violations**:
- Business record referencing non-existent user
- Classification record referencing non-existent business
- Risk assessment referencing non-existent business

### 2. Data Types and Formats

**Purpose**: Ensures data conforms to expected types and formats

**What it checks**:
- UUID format validation
- String length constraints
- Numeric range validation
- Boolean value validation
- Timestamp format validation
- JSON structure validation
- Email format validation
- URL format validation
- Phone number format validation

**Business-specific validations**:
- Confidence scores must be between 0 and 1
- Risk scores must be between 0 and 1
- MCC codes must be 4 digits
- NAICS codes must be 2-6 digits
- SIC codes must be 2-4 digits

### 3. Orphaned Records

**Purpose**: Identifies records that violate referential integrity

**What it checks**:
- Records in referencing tables that don't have corresponding records in referenced tables
- Specific business-critical relationships
- Data integrity across the entire database

**Example scenarios**:
- API keys for deleted users
- Business classifications for deleted businesses
- Risk assessments for deleted businesses
- Industry keywords for deleted industries

### 4. Data Consistency

**Purpose**: Validates business rules and data consistency across tables

**What it checks**:
- Duplicate email addresses
- Valid user roles
- Business name consistency
- Valid website URLs
- Classification confidence scores
- Risk level consistency
- Code format validation
- Timestamp logical relationships
- JSON data structure validation

**Business rules validated**:
- User roles must be valid enum values
- Risk levels must be valid enum values
- Confidence scores must be in valid range
- Updated timestamps must be after created timestamps

### 5. Table Structure

**Purpose**: Verifies table schemas and column definitions

**What it checks**:
- All expected tables exist
- No unexpected tables exist
- Required columns are present
- Column data types are correct
- Nullable constraints are correct

**Expected tables**:
- Core tables: `users`, `businesses`, `api_keys`
- Classification tables: `industries`, `industry_keywords`, `classification_codes`
- Risk tables: `risk_keywords`, `business_risk_assessments`
- Performance tables: `business_classifications`, `risk_assessments`

### 6. Indexes

**Purpose**: Validates index structure and identifies missing or duplicate indexes

**What it checks**:
- Missing critical indexes for performance
- Duplicate indexes that waste space
- Unused indexes that slow down writes

**Critical indexes validated**:
- Unique indexes on email addresses
- Foreign key indexes for joins
- Search indexes on keywords
- Performance indexes on timestamps

### 7. Constraints

**Purpose**: Validates database constraints and identifies violations

**What it checks**:
- Missing critical constraints
- Constraint violations in data
- Check constraint violations

**Constraints validated**:
- Unique constraints on email addresses
- Check constraints on enum values
- Foreign key constraints
- Not null constraints

## Configuration

### ValidationConfig

```go
type ValidationConfig struct {
    // Validation settings
    CheckForeignKeys     bool          // Enable foreign key validation
    CheckDataTypes       bool          // Enable data type validation
    CheckOrphanedRecords bool          // Enable orphaned records validation
    CheckDataConsistency bool          // Enable data consistency validation
    
    // Performance settings
    BatchSize           int           // Batch size for processing
    Timeout             time.Duration // Overall validation timeout
    ParallelValidation  bool          // Enable parallel validation
    
    // Reporting settings
    DetailedReporting   bool          // Include detailed results
    IncludeStatistics   bool          // Include performance statistics
}
```

### Default Configuration

```go
config := &ValidationConfig{
    CheckForeignKeys:     true,
    CheckDataTypes:       true,
    CheckOrphanedRecords: true,
    CheckDataConsistency: true,
    BatchSize:           1000,
    Timeout:             30 * time.Minute,
    ParallelValidation:  true,
    DetailedReporting:   true,
    IncludeStatistics:   true,
}
```

## Output and Reporting

### Validation Report Structure

```go
type IntegrityReport struct {
    Summary           ValidationSummary     // High-level summary
    Results           []ValidationResult    // Detailed results for each check
    Recommendations   []string              // Actionable recommendations
    GeneratedAt       time.Time             // Report generation timestamp
    DatabaseVersion   string                // Database version information
    ValidationVersion string                // Validation system version
}
```

### Validation Summary

```go
type ValidationSummary struct {
    TotalChecks       int           // Total number of checks executed
    PassedChecks      int           // Number of checks that passed
    FailedChecks      int           // Number of checks that failed
    WarningChecks     int           // Number of checks with warnings
    SkippedChecks     int           // Number of checks that were skipped
    TotalErrors       int           // Total number of errors found
    TotalWarnings     int           // Total number of warnings found
    ExecutionTime     time.Duration // Total execution time
}
```

### Validation Result

```go
type ValidationResult struct {
    CheckName        string                 // Name of the validation check
    Status           ValidationStatus       // Status: passed, failed, warning, skipped
    Message          string                 // Human-readable message
    Details          map[string]interface{} // Detailed information
    ErrorCount       int                    // Number of errors found
    WarningCount     int                    // Number of warnings found
    ExecutionTime    time.Duration          // Time taken to execute this check
    Timestamp        time.Time              // When this check was executed
}
```

### Example Report Output

```json
{
  "summary": {
    "total_checks": 7,
    "passed_checks": 5,
    "failed_checks": 1,
    "warning_checks": 1,
    "skipped_checks": 0,
    "total_errors": 3,
    "total_warnings": 2,
    "execution_time": "2m30s"
  },
  "results": [
    {
      "check_name": "foreign_key_constraints",
      "status": "passed",
      "message": "All foreign key constraints are valid",
      "error_count": 0,
      "warning_count": 0,
      "execution_time": "45s"
    },
    {
      "check_name": "data_types",
      "status": "failed",
      "message": "Found 3 data type violations",
      "error_count": 3,
      "warning_count": 0,
      "execution_time": "1m15s"
    }
  ],
  "recommendations": [
    "Review and fix foreign key constraint violations to ensure referential integrity",
    "Validate and correct data type mismatches to prevent runtime errors"
  ],
  "generated_at": "2025-01-19T10:30:00Z",
  "database_version": "PostgreSQL 15.4",
  "validation_version": "1.0.0"
}
```

## Testing

### Running Tests

```bash
# Run all tests
go test ./internal/database/integrity/...

# Run with verbose output
go test -v ./internal/database/integrity/...

# Run integration tests (requires TEST_DATABASE_URL)
TEST_DATABASE_URL="postgres://user:pass@host:port/dbname" go test -v ./internal/database/integrity/...

# Run benchmarks
go test -bench=. ./internal/database/integrity/...
```

### Test Coverage

The test suite includes:
- Unit tests for all validators
- Integration tests with real database
- Benchmark tests for performance
- Configuration tests
- Error handling tests

## Performance Considerations

### Optimization Strategies

1. **Batch Processing**: Processes records in configurable batches
2. **Parallel Validation**: Runs independent checks in parallel
3. **Efficient Queries**: Uses optimized SQL queries
4. **Index Usage**: Leverages database indexes for fast lookups
5. **Timeout Management**: Prevents long-running validations

### Performance Benchmarks

Typical performance on a medium-sized database:
- **Small database** (< 10K records): 30-60 seconds
- **Medium database** (10K-100K records): 2-5 minutes
- **Large database** (100K+ records): 5-15 minutes

### Memory Usage

- **Peak memory usage**: ~50-100MB for large databases
- **Memory efficiency**: Processes data in batches to minimize memory usage
- **Garbage collection**: Optimized to reduce GC pressure

## Troubleshooting

### Common Issues

1. **Connection Timeout**
   - Increase timeout in configuration
   - Check database connection limits
   - Verify network connectivity

2. **Memory Issues**
   - Reduce batch size
   - Disable parallel validation
   - Check available system memory

3. **Permission Errors**
   - Ensure database user has read permissions
   - Check access to information_schema tables
   - Verify table-level permissions

### Debug Mode

Enable verbose logging for debugging:

```go
config := &ValidationConfig{
    // ... other settings
    DetailedReporting: true,
}

logger := log.New(os.Stdout, "[DEBUG] ", log.LstdFlags|log.Lshortfile)
validator := integrity.NewValidator(db, logger, config)
```

## Integration with CI/CD

### GitHub Actions Example

```yaml
name: Database Integrity Validation

on:
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM
  workflow_dispatch:

jobs:
  validate-db:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v3
        with:
          go-version: '1.22'
          
      - name: Build validator
        run: go build -o validate-db ./cmd/validate-db
        
      - name: Run database validation
        env:
          DATABASE_URL: ${{ secrets.DATABASE_URL }}
        run: |
          ./validate-db -db-url "$DATABASE_URL" -output "validation-report.json"
          
      - name: Upload validation report
        uses: actions/upload-artifact@v3
        with:
          name: validation-report
          path: validation-report.json
```

### Alerting Integration

```go
// Example integration with monitoring system
func (v *Validator) ValidateWithAlerting(ctx context.Context) (*IntegrityReport, error) {
    report, err := v.ValidateAll(ctx)
    if err != nil {
        return nil, err
    }
    
    // Send alerts for critical issues
    if report.Summary.FailedChecks > 0 {
        v.sendAlert("Database integrity validation failed", report)
    }
    
    return report, nil
}
```

## Future Enhancements

### Planned Features

1. **Real-time Monitoring**: Continuous validation with real-time alerts
2. **Automated Fixes**: Automatic correction of common issues
3. **Performance Analytics**: Historical performance tracking
4. **Custom Rules**: User-defined validation rules
5. **API Integration**: REST API for programmatic access
6. **Dashboard**: Web-based dashboard for monitoring

### Extension Points

The system is designed to be easily extensible:

1. **New Validators**: Implement the `ValidationCheck` interface
2. **Custom Rules**: Add business-specific validation logic
3. **Output Formats**: Support for different report formats
4. **Integration**: Easy integration with monitoring systems

## Contributing

### Adding New Validators

1. Create a new validator file (e.g., `custom_validator.go`)
2. Implement the `ValidationCheck` interface
3. Add the validator to the main validator's check list
4. Write comprehensive tests
5. Update documentation

### Code Standards

- Follow Go best practices and idioms
- Write comprehensive tests
- Document all public functions
- Use meaningful variable names
- Handle errors appropriately
- Write clear commit messages

## License

This project is part of the KYB Platform and follows the same licensing terms.
