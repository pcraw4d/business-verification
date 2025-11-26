#!/bin/bash
# Quick Quantization Validation Script
# Run this before deploying to production

set -e

echo "🚀 Starting Quantization Validation..."
echo ""

cd "$(dirname "$0")"

# Test 1: Quantization enabled
echo "Test 1: Verify quantization is enabled"
USE_QUANTIZATION=true python3 -c "
from distilbart_classifier import DistilBARTBusinessClassifier
import torch
classifier = DistilBARTBusinessClassifier({
    'model_save_path': 'models/distilbart',
    'quantized_models_path': 'models/quantized',
    'use_quantization': True,
    'quantization_dtype': torch.qint8,
    'industry_labels': ['Technology', 'Healthcare']
})
assert classifier.use_quantization == True, 'Quantization not enabled!'
print('✅ Quantization enabled')
" || { echo "❌ Test 1 failed"; exit 1; }

# Test 2: Classification works
echo "Test 2: Verify classification works"
USE_QUANTIZATION=true python3 -c "
from distilbart_classifier import DistilBARTBusinessClassifier
import torch
classifier = DistilBARTBusinessClassifier({
    'model_save_path': 'models/distilbart',
    'quantized_models_path': 'models/quantized',
    'use_quantization': True,
    'quantization_dtype': torch.qint8,
    'industry_labels': ['Technology', 'Healthcare']
})
result = classifier.classify_only(content='Technology company', max_length=512)
assert result.get('industry') is not None, 'Classification failed!'
assert result.get('confidence', 0) > 0.85, 'Confidence too low!'
print('✅ Classification works (confidence: {:.2%})'.format(result.get('confidence', 0)))
" || { echo "❌ Test 2 failed"; exit 1; }

# Test 3: Performance
echo "Test 3: Verify performance"
USE_QUANTIZATION=true python3 -c "
from distilbart_classifier import DistilBARTBusinessClassifier
import torch
import time
classifier = DistilBARTBusinessClassifier({
    'model_save_path': 'models/distilbart',
    'quantized_models_path': 'models/quantized',
    'use_quantization': True,
    'quantization_dtype': torch.qint8,
    'industry_labels': ['Technology', 'Healthcare']
})
start = time.time()
result = classifier.classify_only(content='Technology company', max_length=512)
elapsed = time.time() - start
assert elapsed < 0.300, f'Too slow: {elapsed:.3f}s'
print(f'✅ Performance acceptable ({elapsed:.3f}s)')
" || { echo "❌ Test 3 failed"; exit 1; }

# Test 4: Enhanced classification
echo "Test 4: Verify enhanced classification"
USE_QUANTIZATION=true python3 -c "
from distilbart_classifier import DistilBARTBusinessClassifier
import torch
classifier = DistilBARTBusinessClassifier({
    'model_save_path': 'models/distilbart',
    'quantized_models_path': 'models/quantized',
    'use_quantization': True,
    'quantization_dtype': torch.qint8,
    'industry_labels': ['Technology', 'Healthcare']
})
result = classifier.classify_with_enhancement(
    content='Technology company providing software services',
    business_name='TechCorp',
    max_length=1024
)
assert result.get('industry') is not None, 'Enhanced classification failed!'
assert result.get('summary') is not None, 'Summary missing!'
assert result.get('explanation') is not None, 'Explanation missing!'
print('✅ Enhanced classification works')
" || { echo "❌ Test 4 failed"; exit 1; }

echo ""
echo "✅ All validation tests passed!"
echo "Ready for production deployment."

