# Complete Phase 1 Log Analysis Guide

## What's Needed for Remaining Metrics

### Summary

To extract the remaining Phase 1 metrics, you need:

1. **Log Access** - Docker logs from classification service
2. **Log Parsing Tools** - jq, grep, awk for extraction
3. **Analysis Scripts** - Automated metric extraction
4. **Baseline Data** - Pre-Phase 1 accuracy (for comparison)

---

## 1. Quality Scores (Target: ≥0.7 for 90%+)

### What's Needed:

**A. Log Source:**
- Structured JSON logs from `zap.Logger` in `internal/external/website_scraper.go`
- Log entries with `"quality_score"` field
- Format: `{"level":"info","msg":"✅ [Phase1] Strategy succeeded","quality_score":0.85}`

**B. Extraction Method:**
```bash
# Method 1: From Docker logs
docker compose -f docker-compose.local.yml logs classification-service | \
  jq -r 'select(.quality_score != null) | .quality_score'

# Method 2: Using extraction script
./scripts/extract-phase1-metrics.sh
```

**C. Analysis:**
- Count total quality scores
- Calculate average
- Count scores ≥0.7
- Calculate percentage ≥0.7
- Validate ≥90% target

**D. Current Status:**
- ✅ Logs are being generated (structured JSON)
- ⚠️ May need to extract from Docker log format
- ✅ Script available: `scripts/extract-phase1-metrics.sh`

---

## 2. Word Counts (Target: ≥200 average)

### What's Needed:

**A. Log Source:**
- Structured JSON logs from `zap.Logger`
- Log entries with `"word_count"` field
- Format: `{"level":"info","msg":"✅ [Phase1] Strategy succeeded","word_count":342}`

**B. Extraction Method:**
```bash
# Method 1: From Docker logs
docker compose -f docker-compose.local.yml logs classification-service | \
  jq -r 'select(.word_count != null) | .word_count'

# Method 2: Using extraction script
./scripts/extract-phase1-metrics.sh
```

**C. Analysis:**
- Extract all word counts
- Calculate average
- Calculate median
- Validate ≥200 average target

**D. Current Status:**
- ✅ Logs are being generated (structured JSON)
- ⚠️ May need to extract from Docker log format
- ✅ Script available: `scripts/extract-phase1-metrics.sh`

---

## 3. Classification Accuracy Improvement (Target: 50-60%)

### What's Needed:

**A. Baseline Data (Pre-Phase 1):**
- Historical classification results
- Or re-run test set with old scraper
- Accuracy percentage before Phase 1

**B. Current Data (Post-Phase 1):**
- Test results from comprehensive test suite
- File: `/tmp/phase1-test-results.json`
- 44/44 successful classifications

**C. Comparison Method:**
```bash
# If baseline data exists
./scripts/compare-accuracy.sh baseline-results.json /tmp/phase1-test-results.json

# Or calculate from test results
jq '[.[] | select(.success == true)] | length' /tmp/phase1-test-results.json
```

**D. Analysis:**
- Calculate baseline accuracy
- Calculate current accuracy
- Calculate improvement percentage
- Validate 50-60% improvement target

**E. Current Status:**
- ✅ Current test results available
- ⚠️ Baseline data needed for comparison
- ⚠️ Need to define "accuracy" metric (success rate vs correct classification)

---

## Log Format Issues

### Problem

The classification service uses **two logging systems**:

1. **Structured JSON (zap.Logger)** - Used by external scraper
   - ✅ Visible in Docker logs
   - ✅ Contains quality_score, word_count fields
   - ✅ Easy to parse with jq

2. **Plain Text (log.Printf)** - Used by enhanced scraper
   - ⚠️ May not be visible in Docker logs
   - ⚠️ Format: "Quality: 0.85, Words: 342"
   - ⚠️ Requires grep/awk parsing

### Solution Options

**Option 1: Use Structured Logs Only (Recommended)**
- Extract from zap.Logger structured logs
- These are already visible and contain all metrics
- Use `scripts/extract-phase1-metrics.sh`

**Option 2: Fix Plain Text Log Visibility**
- Ensure Docker captures stdout/stderr
- Or redirect `log.Printf` to structured logger

**Option 3: Add Explicit Structured Logging**
- Add zap.Logger calls alongside log.Printf
- Ensures metrics are in structured format

---

## Required Tools

### 1. Command Line Tools
- `jq` - JSON parsing (already installed)
- `grep` - Text pattern matching (already installed)
- `awk` - Text processing (already installed)
- `bc` - Calculator (already installed)

### 2. Scripts
- ✅ `scripts/extract-phase1-metrics.sh` - Extract all metrics
- ⚠️ `scripts/compare-accuracy.sh` - Compare baseline (needs baseline data)

### 3. Docker Access
- ✅ Docker Compose access
- ✅ Service logs access

---

## Step-by-Step Process

### Step 1: Extract Quality Scores and Word Counts

```bash
# Run extraction script
./scripts/extract-phase1-metrics.sh

# Or manually extract
docker compose -f docker-compose.local.yml logs classification-service --tail=10000 > /tmp/logs.json
jq -r 'select(.quality_score != null) | .quality_score' /tmp/logs.json
jq -r 'select(.word_count != null) | .word_count' /tmp/logs.json
```

### Step 2: Calculate Statistics

```bash
# Quality scores
quality_scores=$(jq -r 'select(.quality_score != null) | .quality_score' /tmp/logs.json)
avg=$(echo "$quality_scores" | awk '{sum+=$1; count++} END {print sum/count}')
high=$(echo "$quality_scores" | awk '$1 >= 0.7 {count++} END {print count}')

# Word counts
word_counts=$(jq -r 'select(.word_count != null) | .word_count' /tmp/logs.json)
avg_words=$(echo "$word_counts" | awk '{sum+=$1; count++} END {print sum/count}')
```

### Step 3: Validate Targets

- Quality scores: Check if ≥90% are ≥0.7
- Word counts: Check if average ≥200

### Step 4: Accuracy Comparison (If Baseline Available)

```bash
# Compare baseline vs current
./scripts/compare-accuracy.sh baseline.json current.json
```

---

## Current Status

### ✅ Available Now:
- Log extraction script
- Docker log access
- Test results (44/44 successful)
- Structured JSON logs with metrics

### ⚠️ Needs Work:
- Log parsing (Docker log format may need adjustment)
- Baseline accuracy data (for comparison)
- Accuracy metric definition (what counts as "accurate"?)

### 📝 Next Actions:
1. Run `./scripts/extract-phase1-metrics.sh` to extract metrics
2. Analyze results and validate targets
3. Get baseline accuracy data (if available)
4. Compare and calculate improvement

---

## Quick Start

```bash
# 1. Extract all metrics
./scripts/extract-phase1-metrics.sh

# 2. View results
cat /tmp/phase1-metrics.json | jq .

# 3. Check if targets are met
# (Script will show PASS/FAIL for each metric)
```

