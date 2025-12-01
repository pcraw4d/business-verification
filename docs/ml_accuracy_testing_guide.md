# ML-Enabled Accuracy Testing Guide

**Date**: 2025-11-30  
**Status**: ✅ **Ready to Use**

---

## Quick Start

### Option 1: Automated Script (Recommended)

```bash
# Run the automated script that starts ML service and runs tests
./scripts/run_ml_accuracy_tests.sh
```

This script will:

1. ✅ Check if Python ML service is running
2. ✅ Start it if needed (in background)
3. ✅ Wait for it to be ready
4. ✅ Set `PYTHON_ML_SERVICE_URL` environment variable
5. ✅ Run accuracy tests with ML support
6. ✅ Save results to timestamped JSON file

### Option 2: Manual Steps

#### Step 1: Start Python ML Service

**Option A: With Docker (if Docker is running)**

```bash
cd python_ml_service
docker-compose up -d python-ml-service
```

**Option B: Direct Python (if Docker is not available)**

```bash
cd python_ml_service

# Create virtual environment if needed
python3 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt

# Start service
python app.py
# Or: uvicorn app:app --host 0.0.0.0 --port 8000
```

The service will start on `http://localhost:8000`

#### Step 2: Verify Service is Running

```bash
# Check health
curl http://localhost:8000/health

# Check model info
curl http://localhost:8000/model-info | jq
```

#### Step 3: Set Environment Variable

```bash
export PYTHON_ML_SERVICE_URL="http://localhost:8000"
```

#### Step 4: Run Accuracy Tests

```bash
# Build if needed
go build -o bin/comprehensive_accuracy_test ./cmd/comprehensive_accuracy_test

# Run tests
./bin/comprehensive_accuracy_test -verbose -output accuracy_report_ml.json
```

---

## Comparing ML vs Keyword-Based Results

### Run with ML

```bash
export PYTHON_ML_SERVICE_URL="http://localhost:8000"
./bin/comprehensive_accuracy_test -verbose -output accuracy_report_ml.json
```

### Run without ML (keyword-based only)

```bash
unset PYTHON_ML_SERVICE_URL
./bin/comprehensive_accuracy_test -verbose -output accuracy_report_keyword.json
```

### Compare Results

```bash
# Compare overall accuracy
jq '.metrics.overall_accuracy' accuracy_report_ml.json
jq '.metrics.overall_accuracy' accuracy_report_keyword.json

# Compare industry accuracy
jq '.metrics.industry_accuracy' accuracy_report_ml.json
jq '.metrics.industry_accuracy' accuracy_report_keyword.json

# Compare code accuracy
jq '.metrics.code_accuracy' accuracy_report_ml.json
jq '.metrics.code_accuracy' accuracy_report_keyword.json
```

---

## Expected Improvements with ML

Based on the implementation, ML (DistilBART) should provide:

1. **Better Industry Detection**:

   - More accurate industry classification from business descriptions
   - Better handling of ambiguous business names
   - Improved confidence scores

2. **Enhanced Context Understanding**:

   - Better extraction of industry-relevant information from website content
   - Improved summarization of business descriptions
   - More accurate keyword extraction

3. **Reduced "General Business" Fallback**:
   - ML should reduce cases where system falls back to "General Business"
   - Better industry-specific classification

---

## Troubleshooting

### Python ML Service Won't Start

**Check logs:**

```bash
# If using Docker
docker-compose logs python-ml-service

# If running directly
tail -f python_ml_service.log
```

**Common issues:**

- Port 8000 already in use: Change port in `app.py` or stop other service
- Missing dependencies: Run `pip install -r requirements.txt`
- Out of memory: ML models need 2-4GB RAM

### Service Not Responding

**Check if service is running:**

```bash
curl http://localhost:8000/health
```

**Check if port is accessible:**

```bash
lsof -i :8000
```

### Tests Still Using Keyword-Based Classification

**Verify environment variable is set:**

```bash
echo $PYTHON_ML_SERVICE_URL
```

**Check test logs for ML initialization:**

```bash
./bin/comprehensive_accuracy_test -verbose 2>&1 | grep -i "ml\|python\|distilbart"
```

You should see:

```
🐍 Initializing Python ML Service: http://localhost:8000
✅ Python ML Service initialized successfully
🤖 Creating IndustryDetectionService with ML support
```

---

## Performance Considerations

### ML Service Resources

- **Memory**: 2-4GB RAM recommended
- **CPU**: 1-2 cores recommended
- **Startup Time**: 30-60 seconds (model loading)
- **Response Time**: 100-500ms per classification

### Test Execution Time

- **With ML**: ~15-20 minutes for 184 test cases
- **Without ML**: ~10-15 minutes for 184 test cases

The ML service adds overhead but should provide better accuracy.

---

## Next Steps After Testing

1. **Analyze Results**:

   - Compare ML vs keyword-based accuracy
   - Identify industries where ML performs better
   - Review cases where ML fails

2. **Optimize ML Service**:

   - Fine-tune DistilBART on your test dataset
   - Adjust confidence thresholds
   - Add industry-specific models

3. **Production Deployment**:
   - Deploy Python ML service to Railway/cloud
   - Update `PYTHON_ML_SERVICE_URL` to production URL
   - Monitor ML service performance

---

## Files Created

- `scripts/run_ml_accuracy_tests.sh` - Automated test runner
- `docs/ml_accuracy_testing_guide.md` - This guide
- `accuracy_report_ml_*.json` - Test results with ML
- `python_ml_service.log` - ML service logs (if run directly)

---

## Support

If you encounter issues:

1. Check service logs: `tail -f python_ml_service.log`
2. Verify service health: `curl http://localhost:8000/health`
3. Review test output for ML initialization messages
4. Check environment variables are set correctly
