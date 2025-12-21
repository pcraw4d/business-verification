# 50-Sample Validation Test - Monitoring

**Date**: December 21, 2025  
**Status**: ⏳ **Running**  
**Test File**: `test/results/railway_e2e_validation_50_sample_20251221_173403.txt`

---

## Test Progress

**Latest Status**: Test is running successfully
- ✅ Service health check passed
- ✅ Test started with 50 samples
- ⏳ Processing samples (currently at sample 30/50)

---

## How to Monitor

### Check Test Progress
```bash
tail -f test/results/railway_e2e_validation_50_sample_20251221_173403.txt
```

### Check Latest Status
```bash
tail -50 test/results/railway_e2e_validation_50_sample_20251221_173403.txt | grep -E "(Running test|✅|❌|Completed)"
```

### Wait for Completion
The test will complete when you see:
- `✅ Completed all tests in ...`
- `📊 Test report saved to ...`
- `📊 Analysis report saved to ...`

---

## Expected Results Files

Once complete, the following files will be generated:

1. **Test Report**: `test/integration/test/results/railway_e2e_classification_YYYYMMDD_HHMMSS.json`
2. **Analysis Report**: `test/integration/test/results/railway_e2e_analysis_YYYYMMDD_HHMMSS.json`

---

## Analyze Results

Once the test completes, run:
```bash
./test/results/analyze_validation_results.sh
```

This will:
- Extract key metrics
- Compare with baseline
- Show pass/fail status for each target
- Provide overall status

---

## Key Metrics to Watch

### Track 5.1: Scraping Success Rate
- **Target**: ≥70%
- **Baseline**: 0.0%

### Track 4.2: Code Accuracy
- **Overall Accuracy Target**: 25-35% (baseline: 10.8%)
- **MCC Top 1 Target**: 10-20% (baseline: 0.0%)
- **MCC Top 3 Target**: 25-35% (baseline: 12.5%)
- **NAICS Accuracy Target**: 20-40% (baseline: 0.0%)
- **SIC Accuracy Target**: 20-40% (baseline: 0.0%)

---

**Last Checked**: $(date)  
**Status**: ⏳ Test in progress

