#!/bin/bash
# Script to analyze 50-sample validation test results

LATEST_RESULT=$(ls -t test/integration/test/results/railway_e2e_classification_*.json 2>/dev/null | head -1)
LATEST_ANALYSIS=$(ls -t test/integration/test/results/railway_e2e_analysis_*.json 2>/dev/null | head -1)

if [ -z "$LATEST_RESULT" ]; then
    echo "❌ No test results found. Waiting for test to complete..."
    exit 1
fi

echo "📊 Analyzing validation test results..."
echo "Result file: $LATEST_RESULT"
echo "Analysis file: $LATEST_ANALYSIS"
echo ""

python3 << 'EOF'
import json
import sys
from datetime import datetime

# Find latest result file
import glob
result_files = sorted(glob.glob('test/integration/test/results/railway_e2e_classification_*.json'), reverse=True)
analysis_files = sorted(glob.glob('test/integration/test/results/railway_e2e_analysis_*.json'), reverse=True)

if not result_files:
    print("❌ No result files found")
    sys.exit(1)

result_file = result_files[0]
analysis_file = analysis_files[0] if analysis_files else None

print(f"📊 Analyzing: {result_file}")
if analysis_file:
    print(f"📊 Analysis: {analysis_file}")
print("")

with open(result_file) as f:
    data = json.load(f)

# Extract key metrics
summary = data.get('summary', {})
metrics = data.get('metrics', {})

print("=" * 80)
print("VALIDATION TEST RESULTS - Priority 1 Fixes")
print("=" * 80)
print("")

# Overall metrics
print("📈 OVERALL METRICS")
print("-" * 80)
print(f"Total Samples: {summary.get('total_samples', 0)}")
print(f"Successful Requests: {summary.get('successful_requests', 0)} ({summary.get('success_rate', 0)*100:.1f}%)")
print(f"Failed Requests: {summary.get('failed_requests', 0)} ({summary.get('failure_rate', 0)*100:.1f}%)")
print(f"Average Latency: {summary.get('average_latency_ms', 0):.1f}ms")
print("")

# Scraping metrics
print("🌐 SCRAPING METRICS (Track 5.1)")
print("-" * 80)
scraping_metrics = metrics.get('scraping', {})
print(f"Scraping Success Rate: {scraping_metrics.get('success_rate', 0)*100:.1f}%")
print(f"  Target: ≥70%")
print(f"  Status: {'✅ PASS' if scraping_metrics.get('success_rate', 0) >= 0.70 else '❌ FAIL'}")
print(f"Average Scraping Time: {scraping_metrics.get('average_time_ms', 0):.1f}ms")
print(f"Early Exit Rate: {scraping_metrics.get('early_exit_rate', 0)*100:.1f}%")
print("")

# Code accuracy metrics
print("📊 CODE ACCURACY METRICS (Track 4.2)")
print("-" * 80)
code_metrics = metrics.get('code_accuracy', {})
print(f"Overall Code Accuracy: {code_metrics.get('overall_accuracy', 0)*100:.1f}%")
print(f"  Target: 25-35%")
print(f"  Status: {'✅ PASS' if 0.25 <= code_metrics.get('overall_accuracy', 0) <= 0.35 else '❌ FAIL'}")
print("")
print(f"MCC Top 1 Accuracy: {code_metrics.get('mcc_top1_accuracy', 0)*100:.1f}%")
print(f"  Target: 10-20%")
print(f"  Status: {'✅ PASS' if 0.10 <= code_metrics.get('mcc_top1_accuracy', 0) <= 0.20 else '❌ FAIL'}")
print("")
print(f"MCC Top 3 Accuracy: {code_metrics.get('mcc_top3_accuracy', 0)*100:.1f}%")
print(f"  Target: 25-35%")
print(f"  Status: {'✅ PASS' if 0.25 <= code_metrics.get('mcc_top3_accuracy', 0) <= 0.35 else '❌ FAIL'}")
print("")
print(f"NAICS Accuracy: {code_metrics.get('naics_accuracy', 0)*100:.1f}%")
print(f"  Target: 20-40%")
print(f"  Status: {'✅ PASS' if 0.20 <= code_metrics.get('naics_accuracy', 0) <= 0.40 else '❌ FAIL'}")
print("")
print(f"SIC Accuracy: {code_metrics.get('sic_accuracy', 0)*100:.1f}%")
print(f"  Target: 20-40%")
print(f"  Status: {'✅ PASS' if 0.20 <= code_metrics.get('sic_accuracy', 0) <= 0.40 else '❌ FAIL'}")
print("")

# Code generation metrics
print("🔧 CODE GENERATION METRICS")
print("-" * 80)
code_gen = metrics.get('code_generation', {})
print(f"Code Generation Rate: {code_gen.get('generation_rate', 0)*100:.1f}%")
print(f"Average Codes Generated: {code_gen.get('average_codes', 0):.1f}")
print("")

# Comparison with baseline
print("=" * 80)
print("COMPARISON WITH BASELINE (Before Fixes)")
print("=" * 80)
print("")
print("Metric                    | Before  | After   | Change")
print("-" * 80)
print(f"Scraping Success Rate    | 0.0%    | {scraping_metrics.get('success_rate', 0)*100:.1f}%  | {scraping_metrics.get('success_rate', 0)*100:.1f}%")
print(f"Overall Code Accuracy    | 10.8%   | {code_metrics.get('overall_accuracy', 0)*100:.1f}%  | {code_metrics.get('overall_accuracy', 0)*100 - 10.8:+.1f}%")
print(f"MCC Top 1 Accuracy       | 0.0%    | {code_metrics.get('mcc_top1_accuracy', 0)*100:.1f}%  | {code_metrics.get('mcc_top1_accuracy', 0)*100:.1f}%")
print(f"MCC Top 3 Accuracy       | 12.5%   | {code_metrics.get('mcc_top3_accuracy', 0)*100:.1f}%  | {code_metrics.get('mcc_top3_accuracy', 0)*100 - 12.5:+.1f}%")
print(f"NAICS Accuracy           | 0.0%    | {code_metrics.get('naics_accuracy', 0)*100:.1f}%  | {code_metrics.get('naics_accuracy', 0)*100:.1f}%")
print(f"SIC Accuracy             | 0.0%    | {code_metrics.get('sic_accuracy', 0)*100:.1f}%  | {code_metrics.get('sic_accuracy', 0)*100:.1f}%")
print("")

# Overall status
print("=" * 80)
print("OVERALL STATUS")
print("=" * 80)
all_passed = (
    scraping_metrics.get('success_rate', 0) >= 0.70 and
    0.25 <= code_metrics.get('overall_accuracy', 0) <= 0.35 and
    0.10 <= code_metrics.get('mcc_top1_accuracy', 0) <= 0.20
)

if all_passed:
    print("✅ ALL PRIORITY 1 FIXES SUCCESSFUL!")
    print("   - Scraping success rate improved")
    print("   - Code accuracy improved")
    print("   - Ready to proceed with remaining tracks")
else:
    print("⚠️  SOME METRICS NEED ATTENTION")
    if scraping_metrics.get('success_rate', 0) < 0.70:
        print("   - Scraping success rate below target")
    if code_metrics.get('overall_accuracy', 0) < 0.25:
        print("   - Code accuracy below target")
    if code_metrics.get('mcc_top1_accuracy', 0) < 0.10:
        print("   - MCC Top 1 accuracy below target")

print("")
EOF

echo "✅ Analysis complete!"

