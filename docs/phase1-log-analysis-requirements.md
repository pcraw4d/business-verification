# Phase 1 Log Analysis Requirements

## Overview

To extract the remaining Phase 1 metrics (quality scores, word counts, accuracy improvement), we need to analyze logs from multiple sources with different formats.

---

## Log Sources and Formats

### 1. Structured JSON Logs (zap.Logger)
**Location:** Docker logs from classification service  
**Format:** JSON with structured fields  
**Contains:**
- Strategy attempts and results
- Quality scores (`quality_score` field)
- Word counts (`word_count` field)
- Strategy names (`strategy` field)

**Example:**
```json
{"level":"info","ts":1234567890,"caller":"website_scraper.go:1254","msg":"✅ [Phase1] Strategy succeeded","strategy":"simple_http","quality_score":0.85,"word_count":342}
```

### 2. Plain Text Logs (log.Printf)
**Location:** stdout/stderr from classification service  
**Format:** Plain text  
**Contains:**
- Phase 1 success messages with quality scores and word counts
- Strategy information

**Example:**
```
✅ [Phase1] Strategy succeeded - Quality: 0.85, Words: 342
```

**Issue:** These may not be visible in Docker logs if they're written to stdout differently.

### 3. Playwright Service Logs
**Location:** Docker logs from playwright-scraper  
**Format:** Plain text console logs  
**Contains:**
- Scraping requests
- Success/failure messages
- HTML size information

---

## Required Tools and Setup

### 1. Log Extraction Tools

#### A. Docker Logs Access
```bash
# Get all classification service logs
docker compose -f docker-compose.local.yml logs classification-service > classification-logs.json

# Get Playwright service logs
docker compose -f docker-compose.local.yml logs playwright-scraper > playwright-logs.txt
```

#### B. JSON Log Parser (jq)
```bash
# Extract quality scores from structured logs
jq 'select(.msg | contains("Strategy succeeded")) | .quality_score' classification-logs.json

# Extract word counts
jq 'select(.msg | contains("Strategy succeeded")) | .word_count' classification-logs.json

# Extract strategy names
jq 'select(.msg | contains("Strategy")) | .strategy' classification-logs.json
```

#### C. Text Log Parser (grep/awk)
```bash
# Extract quality scores from plain text logs
grep -oP 'Quality: \K[0-9.]+' classification-logs.txt

# Extract word counts
grep -oP 'Words: \K[0-9]+' classification-logs.txt
```

### 2. Metrics Calculation Script

A comprehensive script is needed to:
1. Extract metrics from both log formats
2. Calculate averages and distributions
3. Generate reports

---

## What's Needed for Each Metric

### Quality Scores (Target: ≥0.7 for 90%+ of scrapes)

**Required:**
1. Extract all quality scores from logs
2. Filter for successful scrapes only
3. Calculate:
   - Average quality score
   - Percentage with score ≥0.7
   - Distribution histogram

**Log Sources:**
- Structured JSON: `quality_score` field in "Strategy succeeded" messages
- Plain text: "Quality: X.XX" patterns

**Script Needed:**
```bash
# Extract and analyze quality scores
./scripts/extract-quality-scores.sh
```

### Word Counts (Target: ≥200 average)

**Required:**
1. Extract all word counts from logs
2. Filter for successful scrapes only
3. Calculate:
   - Average word count
   - Median word count
   - Distribution

**Log Sources:**
- Structured JSON: `word_count` field in "Strategy succeeded" messages
- Plain text: "Words: XXX" patterns

**Script Needed:**
```bash
# Extract and analyze word counts
./scripts/extract-word-counts.sh
```

### Classification Accuracy Improvement (Target: 50-60% improvement)

**Required:**
1. Baseline accuracy data (pre-Phase 1)
   - Need historical classification results
   - Or run same test set with old scraper
2. Post-Phase 1 accuracy data
   - Current test results
3. Compare:
   - Calculate accuracy percentage for both
   - Measure improvement
   - Validate 50-60% target

**Data Sources:**
- Test results JSON: `/tmp/phase1-test-results.json`
- Historical baseline (if available)
- Supabase database (if accuracy is tracked)

**Script Needed:**
```bash
# Compare accuracy
./scripts/compare-accuracy.sh baseline.json current-results.json
```

---

## Log Visibility Issues

### Problem
- `log.Printf` output may not be captured in Docker logs
- Structured JSON logs are visible, but plain text may be lost

### Solutions

#### Option 1: Redirect log.Printf to Structured Logger
Modify code to use structured logger instead of `log.Printf`:
```go
// Instead of:
ews.logger.Printf("✅ [Phase1] Strategy succeeded - Quality: %.2f, Words: %d", score, words)

// Use:
ews.zapLogger.Info("✅ [Phase1] Strategy succeeded",
    zap.Float64("quality_score", score),
    zap.Int("word_count", words))
```

#### Option 2: Capture stdout/stderr
Ensure Docker captures all output:
```bash
# Check if logs are going to stdout
docker compose -f docker-compose.local.yml logs classification-service --follow

# Or check container logs directly
docker logs kyb-classification-local 2>&1
```

#### Option 3: Add Explicit Structured Logging
Add structured logging calls alongside `log.Printf`:
```go
// Keep log.Printf for compatibility
ews.logger.Printf("✅ [Phase1] Strategy succeeded - Quality: %.2f, Words: %d", score, words)

// Add structured logging
ews.zapLogger.Info("Phase1 strategy succeeded",
    zap.Float64("quality_score", score),
    zap.Int("word_count", words),
    zap.String("url", url))
```

---

## Recommended Approach

### Step 1: Extract from Structured Logs (Immediate)
Use existing structured JSON logs to extract metrics:
- Quality scores from `quality_score` field
- Word counts from `word_count` field
- Strategy names from `strategy` field

### Step 2: Fix Log Visibility (Short-term)
Ensure all logs are captured:
- Redirect `log.Printf` to structured logger
- Or ensure Docker captures stdout/stderr

### Step 3: Create Analysis Scripts (Immediate)
Build scripts to:
- Extract metrics from logs
- Calculate statistics
- Generate reports

### Step 4: Baseline Comparison (If Available)
Compare with pre-Phase 1 results:
- Use historical data if available
- Or re-run test set with old scraper

---

## Tools to Create

1. **`scripts/extract-phase1-metrics.sh`**
   - Extracts quality scores, word counts, strategies from logs
   - Calculates averages and distributions
   - Generates metrics report

2. **`scripts/analyze-quality-scores.sh`**
   - Focuses on quality score analysis
   - Validates ≥0.7 for 90%+ target

3. **`scripts/analyze-word-counts.sh`**
   - Focuses on word count analysis
   - Validates ≥200 average target

4. **`scripts/compare-accuracy.sh`**
   - Compares baseline vs current accuracy
   - Calculates improvement percentage

---

## Next Steps

1. ✅ Create log extraction scripts
2. ✅ Extract metrics from structured logs
3. ⚠️ Fix log visibility for plain text logs
4. ⚠️ Get baseline accuracy data (if available)
5. ✅ Generate comprehensive metrics report

