#!/bin/bash

# Priority 5: Classification Accuracy Analysis Script
# Analyzes classification accuracy and identifies misclassification patterns

set -e

API_URL="${API_URL:-https://classification-service-production.up.railway.app}"
OUTPUT_DIR="${OUTPUT_DIR:-test/results}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
OUTPUT_FILE="${OUTPUT_DIR}/CLASSIFICATION_ACCURACY_ANALYSIS_${TIMESTAMP}.json"

# Test cases with expected industries
declare -a TEST_CASES=(
    '{"business_name":"Microsoft Corporation","description":"Software development and cloud services","expected_industry":"Technology"}'
    '{"business_name":"JPMorgan Chase","description":"Banking and financial services","expected_industry":"Financial Services"}'
    '{"business_name":"Mayo Clinic","description":"Healthcare and medical services","expected_industry":"Healthcare"}'
    '{"business_name":"Starbucks","description":"Coffee retail and food service","expected_industry":"Food & Beverage"}'
    '{"business_name":"Amazon","description":"E-commerce and retail","expected_industry":"Retail & Commerce"}'
    '{"business_name":"Tesla","description":"Electric vehicle manufacturing","expected_industry":"Manufacturing"}'
    '{"business_name":"Netflix","description":"Streaming entertainment services","expected_industry":"Entertainment"}'
    '{"business_name":"Harvard University","description":"Higher education","expected_industry":"Education"}'
    '{"business_name":"Goldman Sachs","description":"Investment banking","expected_industry":"Financial Services"}'
    '{"business_name":"Walmart","description":"Retail stores","expected_industry":"Retail & Commerce"}'
    '{"business_name":"Apple Inc","description":"Technology and consumer electronics","expected_industry":"Technology"}'
    '{"business_name":"McDonalds","description":"Fast food restaurant chain","expected_industry":"Food & Beverage"}'
    '{"business_name":"Disney","description":"Entertainment and media","expected_industry":"Entertainment"}'
    '{"business_name":"Coca-Cola","description":"Beverage manufacturing","expected_industry":"Food & Beverage"}'
    '{"business_name":"Ford Motor Company","description":"Automotive manufacturing","expected_industry":"Manufacturing"}'
    '{"business_name":"Home Depot","description":"Home improvement retail","expected_industry":"Retail & Commerce"}'
    '{"business_name":"UnitedHealth Group","description":"Healthcare insurance","expected_industry":"Healthcare"}'
    '{"business_name":"Verizon","description":"Telecommunications services","expected_industry":"Technology"}'
    '{"business_name":"Bank of America","description":"Banking services","expected_industry":"Financial Services"}'
    '{"business_name":"CVS Health","description":"Pharmacy and healthcare retail","expected_industry":"Healthcare"}'
)

echo "=========================================="
echo "Classification Accuracy Analysis"
echo "=========================================="
echo ""
echo "API URL: $API_URL"
echo "Test Cases: ${#TEST_CASES[@]}"
echo "Output File: $OUTPUT_FILE"
echo ""

# Initialize results array
RESULTS="[]"
TOTAL=0
CORRECT=0
INDUSTRY_STATS="{}"

for test_case in "${TEST_CASES[@]}"; do
    TOTAL=$((TOTAL + 1))
    
    # Extract expected industry
    EXPECTED=$(echo "$test_case" | python3 -c "import sys, json; print(json.load(sys.stdin)['expected_industry'])" 2>/dev/null)
    
    # Extract test data
    TEST_DATA=$(echo "$test_case" | python3 -c "import sys, json; d=json.load(sys.stdin); del d['expected_industry']; print(json.dumps(d))" 2>/dev/null)
    
    echo "Test $TOTAL: Expected '$EXPECTED'"
    
    # Make API request
    RESPONSE=$(curl -s -X POST "$API_URL/v1/classify" \
        -H "Content-Type: application/json" \
        -d "$TEST_DATA" \
        --max-time 60 2>/dev/null)
    
    if [ $? -ne 0 ] || [ -z "$RESPONSE" ]; then
        echo "  ❌ Request failed"
        continue
    fi
    
    # Extract predicted industry and confidence
    PREDICTED=$(echo "$RESPONSE" | python3 -c "import sys, json; d=json.load(sys.stdin); print(d.get('primary_industry', 'Unknown'))" 2>/dev/null)
    CONFIDENCE=$(echo "$RESPONSE" | python3 -c "import sys, json; d=json.load(sys.stdin); print(d.get('confidence_score', 0))" 2>/dev/null)
    SUCCESS=$(echo "$RESPONSE" | python3 -c "import sys, json; d=json.load(sys.stdin); print(d.get('success', False))" 2>/dev/null)
    
    # Check if correct using industry name normalization
    IS_CORRECT=false
    if [ "$SUCCESS" = "True" ] && [ -n "$PREDICTED" ] && [ "$PREDICTED" != "Unknown" ]; then
        # Use Python to normalize industry names and check equivalence
        # This matches the logic in industry_name_normalizer.go
        IS_CORRECT=$(python3 <<PYTHON_SCRIPT
expected = "$EXPECTED"
predicted = "$PREDICTED"

# Industry name mappings (from industry_name_normalizer.go)
# Maps variations to canonical names
mappings = {
    'banking': 'Financial Services',
    'finance': 'Financial Services',
    'financial': 'Financial Services',
    'financial services': 'Financial Services',
    'restaurants': 'Food & Beverage',
    'restaurant': 'Food & Beverage',
    'cafes & coffee shops': 'Food & Beverage',
    'cafes': 'Food & Beverage',
    'cafe': 'Food & Beverage',
    'coffee shops': 'Food & Beverage',
    'coffee shop': 'Food & Beverage',
    'food & beverage': 'Food & Beverage',
    'food and beverage': 'Food & Beverage',
    'food': 'Food & Beverage',
    'beverage': 'Food & Beverage',
    'retail': 'Retail',
    'retail & commerce': 'Retail',
    'retail and commerce': 'Retail',
    'commerce': 'Retail',
    'industrial manufacturing': 'Manufacturing',
    'manufacturing': 'Manufacturing',
    'technology': 'Technology',
    'tech': 'Technology',
    'healthcare': 'Healthcare',
    'health': 'Healthcare',
    'medical': 'Healthcare',
    'entertainment': 'Entertainment',
    'media': 'Entertainment',
    'streaming': 'Entertainment',
    'education': 'Education',
}

def normalize(name):
    if not name:
        return 'General Business'
    normalized = name.lower().strip()
    # First check exact match
    if normalized in mappings:
        return mappings[normalized]
    # Check if any mapping key is contained in the name
    for key, value in mappings.items():
        if key in normalized:
            return value
    # Return original if no mapping found
    return name

expected_norm = normalize(expected)
predicted_norm = normalize(predicted)

# Check if normalized names match (case-insensitive)
if expected_norm.lower() == predicted_norm.lower():
    print('true')
elif expected.lower() == predicted.lower():
    # Direct match
    print('true')
else:
    print('false')
PYTHON_SCRIPT
)
        
        if [ "$IS_CORRECT" = "true" ]; then
            CORRECT=$((CORRECT + 1))
        fi
    fi
    
    # Update industry stats
    INDUSTRY_STATS=$(echo "$INDUSTRY_STATS" | python3 -c "
import sys, json
stats = json.load(sys.stdin)
expected = '$EXPECTED'
if expected not in stats:
    stats[expected] = {'total': 0, 'correct': 0, 'predictions': {}}
stats[expected]['total'] += 1
if '$IS_CORRECT' == 'true':
    stats[expected]['correct'] += 1
predicted = '$PREDICTED'
if predicted not in stats[expected]['predictions']:
    stats[expected]['predictions'][predicted] = 0
stats[expected]['predictions'][predicted] += 1
print(json.dumps(stats))
" 2>/dev/null)
    
    # Add result
    RESULT=$(echo "$RESPONSE" | python3 -c "
import sys, json
resp = json.load(sys.stdin)
result = {
    'test_number': $TOTAL,
    'expected_industry': '$EXPECTED',
    'predicted_industry': '$PREDICTED',
    'confidence_score': float('$CONFIDENCE') if '$CONFIDENCE' != 'None' else 0.0,
    'is_correct': '$IS_CORRECT' == 'true',
    'success': '$SUCCESS' == 'True',
    'business_name': resp.get('business_name', ''),
    'description': resp.get('description', ''),
    'method': resp.get('metadata', {}).get('method', 'unknown'),
    'processing_path': resp.get('processing_path', ''),
    'early_exit': resp.get('metadata', {}).get('early_exit', False)
}
print(json.dumps(result))
" 2>/dev/null)
    
    RESULTS=$(echo "$RESULTS" | python3 -c "
import sys, json
results = json.load(sys.stdin)
result = json.loads('$RESULT')
results.append(result)
print(json.dumps(results))
" 2>/dev/null)
    
    if [ "$IS_CORRECT" = "true" ]; then
        echo "  ✅ Correct: '$PREDICTED' (confidence: ${CONFIDENCE})"
    else
        echo "  ❌ Incorrect: Expected '$EXPECTED', Got '$PREDICTED' (confidence: ${CONFIDENCE})"
    fi
    
    sleep 1
done

# Calculate overall accuracy
ACCURACY=$(echo "scale=2; $CORRECT * 100 / $TOTAL" | bc)

# Generate summary
SUMMARY=$(python3 <<EOF
import json

results = json.loads('$RESULTS')
stats = json.loads('$INDUSTRY_STATS')

summary = {
    'total_tests': $TOTAL,
    'correct': $CORRECT,
    'incorrect': $TOTAL - $CORRECT,
    'overall_accuracy': $ACCURACY,
    'industry_accuracy': {},
    'misclassification_patterns': {},
    'confidence_analysis': {
        'correct_mean': 0.0,
        'incorrect_mean': 0.0,
        'correct_min': 1.0,
        'incorrect_min': 1.0,
        'correct_max': 0.0,
        'incorrect_max': 0.0
    },
    'method_analysis': {},
    'early_exit_analysis': {
        'total': 0,
        'correct': 0,
        'incorrect': 0
    }
}

# Calculate industry accuracy
for industry, data in stats.items():
    if data['total'] > 0:
        accuracy = (data['correct'] / data['total']) * 100
        summary['industry_accuracy'][industry] = {
            'accuracy': accuracy,
            'correct': data['correct'],
            'total': data['total'],
            'common_misclassifications': dict(sorted(data['predictions'].items(), key=lambda x: x[1], reverse=True)[:3])
        }

# Analyze misclassification patterns
for result in results:
    if not result['is_correct']:
        pattern = f"{result['expected_industry']} -> {result['predicted_industry']}"
        if pattern not in summary['misclassification_patterns']:
            summary['misclassification_patterns'][pattern] = 0
        summary['misclassification_patterns'][pattern] += 1

# Confidence analysis
correct_confidences = [r['confidence_score'] for r in results if r['is_correct']]
incorrect_confidences = [r['confidence_score'] for r in results if not r['is_correct']]

if correct_confidences:
    summary['confidence_analysis']['correct_mean'] = sum(correct_confidences) / len(correct_confidences)
    summary['confidence_analysis']['correct_min'] = min(correct_confidences)
    summary['confidence_analysis']['correct_max'] = max(correct_confidences)

if incorrect_confidences:
    summary['confidence_analysis']['incorrect_mean'] = sum(incorrect_confidences) / len(incorrect_confidences)
    summary['confidence_analysis']['incorrect_min'] = min(incorrect_confidences)
    summary['confidence_analysis']['incorrect_max'] = max(incorrect_confidences)

# Method analysis
for result in results:
    method = result.get('method', 'unknown')
    if method not in summary['method_analysis']:
        summary['method_analysis'][method] = {'total': 0, 'correct': 0}
    summary['method_analysis'][method]['total'] += 1
    if result['is_correct']:
        summary['method_analysis'][method]['correct'] += 1

# Early exit analysis
for result in results:
    if result.get('early_exit', False):
        summary['early_exit_analysis']['total'] += 1
        if result['is_correct']:
            summary['early_exit_analysis']['correct'] += 1
        else:
            summary['early_exit_analysis']['incorrect'] += 1

print(json.dumps(summary, indent=2))
EOF
)

# Save results
FINAL_OUTPUT=$(python3 <<EOF
import json

output = {
    'timestamp': '$TIMESTAMP',
    'api_url': '$API_URL',
    'summary': json.loads('''$SUMMARY'''),
    'detailed_results': json.loads('$RESULTS'),
    'industry_stats': json.loads('$INDUSTRY_STATS')
}

print(json.dumps(output, indent=2))
EOF
)

echo "$FINAL_OUTPUT" > "$OUTPUT_FILE"

echo ""
echo "=========================================="
echo "Analysis Complete"
echo "=========================================="
echo ""
echo "Overall Accuracy: ${ACCURACY}%"
echo "Correct: $CORRECT / $TOTAL"
echo ""
echo "Results saved to: $OUTPUT_FILE"
echo ""

# Print summary
echo "$SUMMARY" | python3 -c "
import sys, json
data = json.load(sys.stdin)
print('Industry Accuracy:')
for industry, stats in sorted(data['industry_accuracy'].items(), key=lambda x: x[1]['accuracy'], reverse=True):
    print(f\"  {industry}: {stats['accuracy']:.1f}% ({stats['correct']}/{stats['total']})\")
    if stats['common_misclassifications']:
        print(f\"    Common misclassifications: {', '.join(stats['common_misclassifications'].keys())}\")
print('')
print('Top Misclassification Patterns:')
for pattern, count in sorted(data['misclassification_patterns'].items(), key=lambda x: x[1], reverse=True)[:5]:
    print(f\"  {pattern}: {count} times\")
print('')
print('Confidence Analysis:')
print(f\"  Correct predictions: mean={data['confidence_analysis']['correct_mean']:.2f}, min={data['confidence_analysis']['correct_min']:.2f}, max={data['confidence_analysis']['correct_max']:.2f}\")
print(f\"  Incorrect predictions: mean={data['confidence_analysis']['incorrect_mean']:.2f}, min={data['confidence_analysis']['incorrect_min']:.2f}, max={data['confidence_analysis']['incorrect_max']:.2f}\")
print('')
print('Method Analysis:')
for method, stats in data['method_analysis'].items():
    accuracy = (stats['correct'] / stats['total']) * 100 if stats['total'] > 0 else 0
    print(f\"  {method}: {accuracy:.1f}% ({stats['correct']}/{stats['total']})\")
"

