# End-to-End Classification Test Guide

This guide explains how to run comprehensive end-to-end tests for the 3-layer classification system.

## Prerequisites

- `curl` installed
- `jq` installed (optional, for better output formatting)
- Access to the deployed classification service

## Quick Start

### Option 1: Test Against Railway Deployment

1. **Get your Railway service URL:**
   - Go to Railway dashboard
   - Find your classification service
   - Copy the public URL (e.g., `https://classification-service-production.up.railway.app`)

2. **Run the test:**
   ```bash
   export CLASSIFICATION_SERVICE_URL="https://your-classification-service.up.railway.app"
   ./test_e2e_classification.sh
   ```

### Option 2: Test Locally

1. **Start the classification service locally:**
   ```bash
   cd services/classification-service
   go run cmd/main.go
   ```

2. **Run the test:**
   ```bash
   export CLASSIFICATION_SERVICE_URL="http://localhost:8080"
   ./test_e2e_classification.sh
   ```

## Test Coverage

The test suite covers:

### 1. Layer 1 (Multi-Strategy) Tests
- **Clear Technology Company**: Should use keyword-based classification
- **Restaurant Business**: Should use keyword-based classification

### 2. Layer 3 (LLM) Tests - Ambiguous Cases
- **Ambiguous Multi-Industry Business**: Complex, multi-sector business that should trigger LLM reasoning
- **Novel Business Model**: Unique business model that requires advanced reasoning
- **Vague Description**: Ambiguous description that needs LLM interpretation

### 3. Website Content Tests
- **E-commerce with Website**: Tests website scraping and embedding-based classification
- **Healthcare Provider with Website**: Tests full pipeline with website content

### 4. Edge Cases
- **Minimal Information**: Tests classification with minimal data
- **Long Description**: Tests handling of long descriptions

## Expected Results

### Layer 1 Tests
- **Method**: Should contain "multi_strategy" or "keyword"
- **Confidence**: Typically > 0.85
- **Processing Time**: < 2 seconds

### Layer 3 (LLM) Tests
- **Method**: Should contain "llm" or "reasoning"
- **Confidence**: Should be > 0.85 (LLM provides high confidence)
- **Processing Time**: 5-15 seconds (LLM inference takes longer)
- **Explanation**: Should include detailed reasoning

### Response Structure
All responses should include:
- `request_id`: Unique request identifier
- `primary_industry`: Industry classification
- `confidence_score`: Confidence level (0.0 - 1.0)
- `classification`: Detailed classification results
  - `industry`: Primary industry
  - `mcc_codes`: Merchant category codes
  - `naics_codes`: NAICS codes
  - `sic_codes`: SIC codes
  - `explanation`: Detailed explanation (Phase 2+)
- `metadata`: Additional metadata including `method` field

## Troubleshooting

### Service Not Responding
- Check that the service URL is correct
- Verify the service is running (check `/health` endpoint)
- Check Railway logs for errors

### LLM Layer Not Triggering
- Verify `LLM_SERVICE_URL` environment variable is set in Railway
- Check that LLM service is deployed and healthy
- Review classification service logs for LLM routing decisions
- Low confidence cases (< 0.88) should trigger LLM

### Timeout Errors
- LLM inference can take 10-30 seconds
- Increase timeout: `--max-time 120` (already set in script)
- Check Railway service resource limits

### Missing Industry Codes
- Verify database is populated with classification codes
- Check Supabase connection
- Review code generation service logs

## Manual Testing

You can also test manually using curl:

```bash
# Simple test
curl -X POST https://your-service.up.railway.app/v1/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "TechCorp Solutions",
    "description": "Software development and cloud computing services"
  }' | jq .

# Test with website (may trigger LLM)
curl -X POST https://your-service.up.railway.app/v1/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Global Innovations Group",
    "description": "A diversified company providing technology consulting, healthcare data analytics, and financial advisory services",
    "website_url": "https://example.com"
  }' | jq .
```

## Success Criteria

✅ **All tests should:**
- Return HTTP 200 status
- Include valid `primary_industry`
- Have `confidence_score` between 0.0 and 1.0
- Include industry codes (MCC, NAICS, SIC)
- Complete within timeout limits

✅ **Layer 3 tests should:**
- Use LLM method (check `metadata.method`)
- Include detailed reasoning in explanation
- Have confidence > 0.85

## Next Steps

After successful testing:
1. Monitor production metrics
2. Review classification accuracy
3. Adjust confidence thresholds if needed
4. Optimize LLM routing logic based on results

