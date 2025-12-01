# Phase 5: Accuracy Testing - Execution Guide

## ✅ Implementation Complete

All components for Phase 5 comprehensive accuracy testing have been successfully implemented:

1. ✅ **Test Dataset** (`scripts/populate_accuracy_test_dataset.sql`) - 184 test cases
2. ✅ **Dataset Manager** (`internal/testing/accuracy_test_dataset.go`)
3. ✅ **Comprehensive Accuracy Tester** (`internal/testing/comprehensive_accuracy_tester.go`)
4. ✅ **Accuracy Report Generator** (`internal/testing/accuracy_report.go`)
5. ✅ **Standalone Test Runner** (`cmd/comprehensive_accuracy_test/main.go`)
6. ✅ **Integration Tests** (`test/integration/comprehensive_accuracy_test.go`)

## 🚀 Running the Accuracy Tests

### Option 1: Standalone Command (Recommended)

A standalone command has been created that can run the accuracy tests without package structure conflicts:

```bash
# Build the command
go build -o bin/comprehensive_accuracy_test ./cmd/comprehensive_accuracy_test

# Set environment variables
export SUPABASE_URL="https://<project-ref>.supabase.co"
export SUPABASE_ANON_KEY="<anon-key>"
export SUPABASE_SERVICE_ROLE_KEY="<service-role-key>"
export TEST_DATABASE_URL="postgres://postgres:[PASSWORD]@db.[PROJECT-REF].supabase.co:5432/postgres"
# OR
export DATABASE_URL="postgres://postgres:[PASSWORD]@db.[PROJECT-REF].supabase.co:5432/postgres"

# Run all tests
./bin/comprehensive_accuracy_test -verbose

# Run tests for a specific category
./bin/comprehensive_accuracy_test -category "Technology" -verbose

# Save JSON report
./bin/comprehensive_accuracy_test -output accuracy_report.json -verbose
```

### Option 2: Command Line Flags

```bash
./bin/comprehensive_accuracy_test \
  -supabase-url "https://<project-ref>.supabase.co" \
  -supabase-key "<anon-key>" \
  -supabase-service-key "<service-role-key>" \
  -database-url "postgres://postgres:[PASSWORD]@db.[PROJECT-REF].supabase.co:5432/postgres" \
  -verbose \
  -output accuracy_report.json
```

### Command Options

- `-supabase-url`: Supabase project URL (or set `SUPABASE_URL` env var)
- `-supabase-key`: Supabase anon key (or set `SUPABASE_ANON_KEY` env var)
- `-supabase-service-key`: Supabase service role key (or set `SUPABASE_SERVICE_ROLE_KEY` env var)
- `-database-url`: Direct PostgreSQL connection string for test dataset (or set `TEST_DATABASE_URL` or `DATABASE_URL` env var)
- `-category`: Run tests for specific category only (optional)
- `-output`: Save JSON report to file (optional)
- `-verbose`: Enable verbose logging

## 📊 Expected Output

The test runner will:

1. **Connect to databases** (Supabase API and direct PostgreSQL)
2. **Load test cases** from `accuracy_test_dataset` table
3. **Run classification** for each test case
4. **Calculate accuracy metrics**:
   - Overall accuracy
   - Industry classification accuracy (target: 95%+)
   - Code accuracy (target: 90%+)
     - MCC code accuracy
     - NAICS code accuracy
     - SIC code accuracy
5. **Generate reports**:
   - Console output with summary
   - Text report with detailed breakdowns
   - JSON report (if `-output` specified)

### Sample Output

```
🚀 Starting Comprehensive Accuracy Test Suite
   Supabase URL: https://xxx.supabase.co
   Database URL: postgres://***@db.xxx.supabase.co:5432/postgres
✅ Database connections established
📊 Loaded 184 test cases
🚀 Starting comprehensive accuracy tests...
✅ Accuracy tests completed: Overall Accuracy: 87.50%, Industry: 92.39%, Codes: 85.22%

================================================================================
COMPREHENSIVE ACCURACY TEST RESULTS
================================================================================
Total Test Cases: 184
Passed: 161
Failed: 23
Overall Accuracy: 87.50%
Industry Accuracy: 92.39% (target: 95%)
Code Accuracy: 85.22% (target: 90%)
  - MCC Accuracy: 88.04%
  - NAICS Accuracy: 84.24%
  - SIC Accuracy: 83.15%
Average Processing Time: 245ms
Total Processing Time: 45s

Accuracy by Category:
  Technology: 91.23%
  Healthcare: 89.36%
  Financial Services: 86.67%
  Retail: 88.00%
  Edge Cases: 70.00%

Accuracy by Industry:
  Technology: 90.91%
  Healthcare: 91.49%
  Financial Services: 87.88%
  Retail: 88.00%
================================================================================
```

## 🎯 Accuracy Targets

The tests validate against these targets:

1. **Industry Classification Accuracy**: ≥ 95%
   - Measures exact industry matches
   - Target: 95%+

2. **Code Generation Accuracy**: ≥ 90%
   - Measures if at least one expected code appears in top 3
   - Calculated separately for MCC, NAICS, SIC
   - Overall target: 90%+

3. **Overall Accuracy**: ≥ 85%
   - Weighted combination of industry and code accuracy
   - Minimum acceptable threshold

## 📈 Interpreting Results

### Success Indicators

✅ **All Targets Met**: 
- Industry accuracy ≥ 95%
- Code accuracy ≥ 90%
- Overall accuracy ≥ 85%

⚠️ **Partial Success**:
- One or more targets close but not met
- Review failed test cases for patterns
- Consider expanding dataset or improving classification logic

❌ **Targets Not Met**:
- Significant gaps in accuracy
- Review failed cases for common patterns
- Consider:
  - Expanding keyword coverage
  - Improving classification algorithms
  - Adding more training data
  - Refining crosswalk mappings

### Common Failure Patterns

1. **Industry Misclassification**:
   - Similar industries confused (e.g., "Technology" vs "Professional Services")
   - Edge cases with ambiguous descriptions
   - Multi-industry businesses

2. **Code Generation Issues**:
   - Missing keywords for specific codes
   - Crosswalk mappings incomplete
   - Code metadata incomplete

3. **Edge Cases**:
   - Unusual business descriptions
   - Non-standard industry classifications
   - International businesses

## 🔍 Analyzing Failed Test Cases

The JSON report includes detailed information about failed test cases:

```json
{
  "test_results": [
    {
      "test_case_id": 42,
      "business_name": "Example Business",
      "expected_industry": "Technology",
      "actual_industry": "Professional Services",
      "industry_match": false,
      "expected_mcc_codes": ["5734", "5045"],
      "actual_mcc_codes": ["7372", "5045"],
      "mcc_match": true,
      "error": ""
    }
  ]
}
```

Use this data to:
1. Identify common failure patterns
2. Determine which industries/categories need improvement
3. Find missing keywords or mappings
4. Prioritize improvements

## 📝 Next Steps After Running Tests

### 1. Review Results

- Check overall accuracy metrics
- Identify categories/industries with low accuracy
- Review failed test cases

### 2. Expand Test Dataset

Current dataset: **184 test cases** (target: 1000+)

**Areas to expand**:
- Add more test cases to underrepresented industries
- Add more edge cases
- Add boundary conditions
- Add international businesses
- Add multi-industry businesses

**How to expand**:
- Run `scripts/populate_accuracy_test_dataset.sql` and add more INSERT statements
- Or use the dataset manager programmatically to add cases

### 3. Improve Classification

Based on test results, consider:

- **Keyword Expansion**: Add more keywords for low-accuracy codes
- **Crosswalk Enhancement**: Improve crosswalk mappings
- **Algorithm Tuning**: Adjust confidence thresholds
- **Metadata Completion**: Ensure all codes have complete metadata

### 4. Iterate

- Run tests regularly
- Track accuracy trends over time
- Set up automated testing in CI/CD
- Create dashboards for monitoring

## 🛠️ Troubleshooting

### Database Connection Issues

**Error**: `Failed to ping Supabase` or `Failed to ping database`

**Solutions**:
- Verify environment variables are set correctly
- Check database URL format
- Ensure database is accessible
- Check firewall/network settings

### No Test Cases Found

**Error**: `no test cases found in dataset`

**Solutions**:
- Verify `accuracy_test_dataset` table exists
- Run `scripts/populate_accuracy_test_dataset.sql` to populate data
- Check database connection is correct

### Low Accuracy

**If accuracy is below targets**:

1. **Review failed cases**: Check JSON report for patterns
2. **Check keyword coverage**: Run `scripts/verify_code_keywords_enhanced.sql`
3. **Verify metadata**: Run `scripts/verify_code_metadata_complete.sql`
4. **Check crosswalks**: Run `scripts/verify_crosswalk_coverage.sql`

## 📚 Related Documentation

- `docs/phase5_accuracy_testing_summary.md` - Implementation summary
- `Accuracy Plan Enhancements.plan.md` - Full plan details
- `scripts/populate_accuracy_test_dataset.sql` - Test dataset script
- `internal/testing/comprehensive_accuracy_tester.go` - Tester implementation

## ✅ Success Criteria

Phase 5 is considered complete when:

1. ✅ Test dataset created (184 cases, expandable to 1000+)
2. ✅ Accuracy test suite implemented
3. ✅ Test runner created and working
4. ✅ Reports generated successfully
5. ⏳ Tests run and results analyzed (pending execution)
6. ⏳ Accuracy targets validated (pending execution)
7. ⏳ Dataset expanded based on results (pending execution)

---

**Status**: Ready for Execution  
**Last Updated**: [Current Date]  
**Next Action**: Run tests with database credentials

