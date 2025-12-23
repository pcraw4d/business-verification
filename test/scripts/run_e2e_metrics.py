#!/usr/bin/env python3
"""
E2E Metrics Test Script
Tests classification service and calculates performance metrics
"""

import json
import time
import requests
import sys
from datetime import datetime
from typing import Dict, List, Any

# Configuration
API_URL = "https://classification-service-production.up.railway.app"
NUM_SAMPLES = 50
TIMEOUT = 120

# Test samples (subset of comprehensive test data)
TEST_SAMPLES = [
    {
        "business_name": "Microsoft Corporation",
        "description": "Software development and cloud computing services",
        "website_url": "https://microsoft.com"
    },
    {
        "business_name": "Apple Inc",
        "description": "Consumer electronics and software",
        "website_url": "https://apple.com"
    },
    {
        "business_name": "Amazon",
        "description": "E-commerce and cloud services",
        "website_url": "https://amazon.com"
    },
    {
        "business_name": "Starbucks Coffee",
        "description": "Coffee shop chain",
        "website_url": "https://starbucks.com"
    },
    {
        "business_name": "Walmart",
        "description": "Retail store chain",
        "website_url": "https://walmart.com"
    },
    {
        "business_name": "McDonalds",
        "description": "Fast food restaurant chain",
        "website_url": "https://mcdonalds.com"
    },
    {
        "business_name": "Tesla Inc",
        "description": "Electric vehicle manufacturer",
        "website_url": "https://tesla.com"
    },
    {
        "business_name": "Netflix",
        "description": "Streaming video service",
        "website_url": "https://netflix.com"
    },
    {
        "business_name": "Bank of America",
        "description": "Banking and financial services",
        "website_url": "https://bankofamerica.com"
    },
    {
        "business_name": "Home Depot",
        "description": "Home improvement retailer",
        "website_url": "https://homedepot.com"
    }
]

def test_classification(sample: Dict[str, Any], max_retries: int = 3) -> Dict[str, Any]:
    """Test a single classification request with retry logic for 502 errors"""
    url = f"{API_URL}/v1/classify"
    
    result = {
        "sample": sample,
        "success": False,
        "http_code": 0,
        "latency_ms": 0,
        "error": None,
        "response": None,
        "retry_count": 0
    }
    
    # Retry logic for 502 errors (transient cold start failures)
    for attempt in range(max_retries):
        try:
            start_time = time.time()
            response = requests.post(
                url,
                json=sample,
                timeout=TIMEOUT,
                verify=False  # Skip SSL verification for Railway
            )
            end_time = time.time()
            
            result["http_code"] = response.status_code
            result["latency_ms"] = (end_time - start_time) * 1000
            result["retry_count"] = attempt
            
            if response.status_code == 200:
                result["success"] = True
                result["response"] = response.json()
                return result  # Success, no retry needed
            elif response.status_code == 502 and attempt < max_retries - 1:
                # 502 error - retry with exponential backoff
                wait_time = (2 ** attempt) * 1.0  # 1s, 2s, 4s
                print(f"  ⚠️  502 error (attempt {attempt + 1}/{max_retries}), retrying in {wait_time:.1f}s...", end="", flush=True)
                time.sleep(wait_time)
                continue
            else:
                # Non-retryable error or last attempt
                result["error"] = f"HTTP {response.status_code}: {response.text[:200]}"
                return result
                
        except requests.exceptions.Timeout:
            if attempt < max_retries - 1:
                wait_time = (2 ** attempt) * 1.0
                print(f"  ⚠️  Timeout (attempt {attempt + 1}/{max_retries}), retrying in {wait_time:.1f}s...", end="", flush=True)
                time.sleep(wait_time)
                continue
            result["error"] = "Request timeout"
            result["latency_ms"] = TIMEOUT * 1000
            return result
        except Exception as e:
            if attempt < max_retries - 1 and "502" in str(e):
                wait_time = (2 ** attempt) * 1.0
                print(f"  ⚠️  Error (attempt {attempt + 1}/{max_retries}), retrying in {wait_time:.1f}s...", end="", flush=True)
                time.sleep(wait_time)
                continue
            result["error"] = str(e)
            return result
    
    return result

def calculate_metrics(results: List[Dict[str, Any]]) -> Dict[str, Any]:
    """Calculate performance metrics from test results"""
    total = len(results)
    successful = sum(1 for r in results if r["success"])
    failed = total - successful
    
    error_rate = (failed / total * 100) if total > 0 else 0
    
    latencies = [r["latency_ms"] for r in results if r["success"]]
    avg_latency = sum(latencies) / len(latencies) if latencies else 0
    
    # Calculate percentiles
    latencies_sorted = sorted(latencies)
    p50 = latencies_sorted[len(latencies_sorted) // 2] if latencies_sorted else 0
    p95 = latencies_sorted[int(len(latencies_sorted) * 0.95)] if latencies_sorted else 0
    p99 = latencies_sorted[int(len(latencies_sorted) * 0.99)] if latencies_sorted else 0
    
    # Classification accuracy (simplified - check if industry is present)
    classifications = sum(1 for r in results if r.get("success") and r.get("response", {}).get("primary_industry"))
    accuracy = (classifications / successful * 100) if successful > 0 else 0
    
    # Code generation rate
    codes_generated = sum(1 for r in results if r.get("success") and (
        r.get("response", {}).get("classification", {}).get("mcc_codes") or
        r.get("response", {}).get("classification", {}).get("naics_codes")
    ))
    code_generation_rate = (codes_generated / successful * 100) if successful > 0 else 0
    
    # Confidence scores
    confidences = [
        r.get("response", {}).get("confidence_score", 0) or 
        r.get("response", {}).get("confidence", 0)
        for r in results if r.get("success")
    ]
    avg_confidence = sum(confidences) / len(confidences) if confidences else 0
    
    return {
        "total_requests": total,
        "successful_requests": successful,
        "failed_requests": failed,
        "error_rate_percent": round(error_rate, 2),
        "average_latency_ms": round(avg_latency, 2),
        "p50_latency_ms": round(p50, 2),
        "p95_latency_ms": round(p95, 2),
        "p99_latency_ms": round(p99, 2),
        "classification_accuracy_percent": round(accuracy, 2),
        "code_generation_rate_percent": round(code_generation_rate, 2),
        "average_confidence": round(avg_confidence, 4),
        "targets": {
            "error_rate_target": "<5%",
            "avg_latency_target": "<10s",
            "accuracy_target": "≥80%",
            "code_generation_target": "≥90%",
            "confidence_target": ">50%"
        }
    }

def main():
    print("=" * 50)
    print("E2E Metrics Test - Classification Service")
    print("=" * 50)
    print(f"API URL: {API_URL}")
    print(f"Test Samples: {NUM_SAMPLES}")
    print(f"Timeout: {TIMEOUT}s")
    print()
    
    # Extend test samples to NUM_SAMPLES
    samples = TEST_SAMPLES * ((NUM_SAMPLES // len(TEST_SAMPLES)) + 1)
    samples = samples[:NUM_SAMPLES]
    
    print(f"🚀 Running {len(samples)} classification tests...")
    print()
    
    results = []
    for i, sample in enumerate(samples, 1):
        print(f"Test {i}/{len(samples)}: {sample['business_name']}... ", end="", flush=True)
        result = test_classification(sample)
        results.append(result)
        
        if result["success"]:
            print(f"✅ ({result['latency_ms']:.0f}ms)")
        else:
            print(f"❌ {result.get('error', 'Unknown error')}")
        
        # Small delay to avoid rate limiting
        time.sleep(0.5)
    
    print()
    print("=" * 50)
    print("📊 Metrics Summary")
    print("=" * 50)
    
    metrics = calculate_metrics(results)
    
    print(f"Total Requests: {metrics['total_requests']}")
    print(f"Successful: {metrics['successful_requests']}")
    print(f"Failed: {metrics['failed_requests']}")
    print(f"Error Rate: {metrics['error_rate_percent']}% (Target: {metrics['targets']['error_rate_target']})")
    print(f"Average Latency: {metrics['average_latency_ms']/1000:.2f}s (Target: {metrics['targets']['avg_latency_target']})")
    print(f"P95 Latency: {metrics['p95_latency_ms']/1000:.2f}s")
    print(f"Classification Accuracy: {metrics['classification_accuracy_percent']}% (Target: {metrics['targets']['accuracy_target']})")
    print(f"Code Generation Rate: {metrics['code_generation_rate_percent']}% (Target: {metrics['targets']['code_generation_target']})")
    print(f"Average Confidence: {metrics['average_confidence']*100:.2f}% (Target: {metrics['targets']['confidence_target']})")
    print()
    
    # Save results
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    results_file = f"test/results/e2e_metrics_{timestamp}.json"
    
    output = {
        "timestamp": datetime.now().isoformat(),
        "api_url": API_URL,
        "metrics": metrics,
        "results": results
    }
    
    with open(results_file, 'w') as f:
        json.dump(output, f, indent=2)
    
    print(f"📄 Results saved to: {results_file}")
    print()
    
    # Compare with previous baseline
    print("=" * 50)
    print("📈 Improvement Analysis")
    print("=" * 50)
    print("Baseline (Before Fixes):")
    print("  - Error Rate: 67.1%")
    print("  - Average Latency: 43.7s")
    print("  - Classification Accuracy: 9.5%")
    print("  - Code Generation Rate: 23.1%")
    print("  - Average Confidence: 24.65%")
    print()
    print("Current Results:")
    print(f"  - Error Rate: {metrics['error_rate_percent']}%")
    print(f"  - Average Latency: {metrics['average_latency_ms']/1000:.2f}s")
    print(f"  - Classification Accuracy: {metrics['classification_accuracy_percent']}%")
    print(f"  - Code Generation Rate: {metrics['code_generation_rate_percent']}%")
    print(f"  - Average Confidence: {metrics['average_confidence']*100:.2f}%")
    print()
    
    # Calculate improvements
    error_improvement = 67.1 - metrics['error_rate_percent']
    latency_improvement = 43.7 - (metrics['average_latency_ms']/1000)
    accuracy_improvement = metrics['classification_accuracy_percent'] - 9.5
    code_gen_improvement = metrics['code_generation_rate_percent'] - 23.1
    confidence_improvement = (metrics['average_confidence']*100) - 24.65
    
    print("Improvements:")
    print(f"  - Error Rate: {error_improvement:+.1f}% improvement")
    print(f"  - Average Latency: {latency_improvement:+.1f}s improvement")
    print(f"  - Classification Accuracy: {accuracy_improvement:+.1f}% improvement")
    print(f"  - Code Generation Rate: {code_gen_improvement:+.1f}% improvement")
    print(f"  - Average Confidence: {confidence_improvement:+.1f}% improvement")
    print()
    
    # Check if targets are met
    targets_met = []
    targets_missed = []
    
    if metrics['error_rate_percent'] < 5:
        targets_met.append("Error Rate")
    else:
        targets_missed.append(f"Error Rate ({metrics['error_rate_percent']:.1f}% > 5%)")
    
    if metrics['average_latency_ms']/1000 < 10:
        targets_met.append("Average Latency")
    else:
        targets_missed.append(f"Average Latency ({metrics['average_latency_ms']/1000:.1f}s > 10s)")
    
    if metrics['classification_accuracy_percent'] >= 80:
        targets_met.append("Classification Accuracy")
    else:
        targets_missed.append(f"Classification Accuracy ({metrics['classification_accuracy_percent']:.1f}% < 80%)")
    
    if metrics['code_generation_rate_percent'] >= 90:
        targets_met.append("Code Generation Rate")
    else:
        targets_missed.append(f"Code Generation Rate ({metrics['code_generation_rate_percent']:.1f}% < 90%)")
    
    if metrics['average_confidence']*100 > 50:
        targets_met.append("Average Confidence")
    else:
        targets_missed.append(f"Average Confidence ({metrics['average_confidence']*100:.1f}% <= 50%)")
    
    if targets_met:
        print("✅ Targets Met:")
        for target in targets_met:
            print(f"   - {target}")
        print()
    
    if targets_missed:
        print("⚠️  Targets Not Yet Met:")
        for target in targets_missed:
            print(f"   - {target}")
        print()
    
    return 0 if len(targets_missed) == 0 else 1

if __name__ == "__main__":
    sys.exit(main())

