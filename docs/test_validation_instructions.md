# Test Validation Instructions

## Re-running Accuracy Tests with Fixes

The accuracy tests have been rebuilt with the following fixes:
1. Enhanced keyword extraction with fallback from business names/descriptions
2. Adaptive confidence thresholds for code generation
3. Industry name normalization for better matching

## Monitoring Test Progress

### Check if test is running:
```bash
ps aux | grep comprehensive_accuracy_test
```

### Monitor test output in real-time:
```bash
tail -f accuracy_test_v2_output.log
```

### Check test completion:
```bash
# Check if report file exists
ls -lh accuracy_report_v2.json

# View quick summary
python3 << 'PYEOF'
import json
with open('accuracy_report_v2.json', 'r') as f:
    data = json.load(f)
metrics = data['metrics']
print(f"Industry Accuracy: {metrics['industry_accuracy']*100:.2f}%")
print(f"Code Accuracy: {metrics['code_accuracy']*100:.2f}%")
print(f"Overall Accuracy: {metrics['overall_accuracy']*100:.2f}%")
print(f"Cases with codes: {sum(1 for r in metrics['test_results'] if len(r['actual_mcc_codes']) > 0 or len(r['actual_naics_codes']) > 0 or len(r['actual_sic_codes']) > 0)}/{metrics['total_test_cases']}")
PYEOF
```

## Expected Test Duration

- **184 test cases** × **~30 seconds per case** = **~10-15 minutes total**
- Each case involves:
  - Website scraping (if URL provided)
  - Keyword extraction
  - Industry detection
  - Code generation

## Comparing Results

### Baseline (Before Fixes):
- Industry Accuracy: **9.24%**
- Code Accuracy: **0.00%**
- Overall Accuracy: **3.70%**
- Cases with codes: **14/184 (7.6%)**

### Expected Improvements (After Fixes):
- Industry Accuracy: **30-50%+** (3-5x improvement)
- Code Accuracy: **20-40%+** (from 0% to meaningful)
- Overall Accuracy: **20-35%+** (5-10x improvement)
- Cases with codes: **50-80%+** (6-10x improvement)

## Analyzing Results

Once the test completes, compare:

1. **Industry Accuracy Improvement**:
   - Check if "General Business" fallback decreased
   - Verify industry name normalization is working
   - Look for better keyword extraction

2. **Code Generation Improvement**:
   - Count cases with codes generated
   - Check if codes are generated for low-confidence industries
   - Verify code quality (not just quantity)

3. **Overall Accuracy**:
   - Compare against baseline
   - Identify remaining failure patterns
   - Plan next improvements

## Next Steps After Validation

1. **If improvements are significant**:
   - Proceed with dataset expansion (184 → 1000+ cases)
   - Add more test cases based on successful patterns
   - Focus on categories with good accuracy

2. **If improvements are minimal**:
   - Investigate remaining issues
   - Check logs for error patterns
   - Refine fixes based on actual results

3. **If new issues appear**:
   - Document new failure patterns
   - Adjust fixes as needed
   - Re-test after adjustments

## Manual Test Run (if needed)

If you need to manually run the test:

```bash
cd "/Users/petercrawford/New tool"

# Set environment variables
export SUPABASE_URL="https://qpqhuqqmkjxsltzshfam.supabase.co"
export SUPABASE_ANON_KEY="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InFwcWh1cXFta2p4c2x0enNoZmFtIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NTQ4NzQ4MzEsImV4cCI6MjA3MDQ1MDgzMX0.UelJkQAVf-XJz1UV0Rbyi-hZHADGOdsHo1PwcPf7JVI"
export SUPABASE_SERVICE_ROLE_KEY="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InFwcWh1cXFta2p4c2x0enNoZmFtIiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTc1NDg3NDgzMSwiZXhwIjoyMDcwNDUwODMxfQ.sIm3w7Ad2kLv08whNBrzdP42nz0s4dsLpvUiYDSwArw"
export DATABASE_URL="postgresql://postgres:Geaux44tigers%21@db.qpqhuqqmkjxsltzshfam.supabase.co:5432/postgres"

# Run test
./bin/comprehensive_accuracy_test -verbose -output accuracy_report_v2.json 2>&1 | tee accuracy_test_v2_output.log
```

## Report Files

- **Baseline Report**: `accuracy_report.json` (before fixes)
- **New Report**: `accuracy_report_v2.json` (after fixes)
- **Test Output**: `accuracy_test_v2_output.log` (detailed logs)

