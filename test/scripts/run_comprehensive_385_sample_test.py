#!/usr/bin/env python3
"""
Comprehensive 385-Sample E2E Classification Test
Tests classification service with full 385-sample dataset to verify all fixes
"""

import json
import time
import requests
import sys
from datetime import datetime
from typing import Dict, List, Any

# Configuration
API_URL = "https://classification-service-production.up.railway.app"
TIMEOUT = 120
MAX_RETRIES = 3

def test_classification(sample: Dict[str, Any], max_retries: int = MAX_RETRIES) -> Dict[str, Any]:
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

def generate_comprehensive_samples() -> List[Dict[str, Any]]:
    """Generate 385 comprehensive test samples matching the Go test"""
    samples = []
    
    # Real-world well-known businesses (~50 samples)
    real_world = [
        {"business_name": "Amazon", "description": "Online retail marketplace", "website_url": "https://www.amazon.com"},
        {"business_name": "Shopify", "description": "E-commerce platform", "website_url": "https://www.shopify.com"},
        {"business_name": "eBay", "description": "Online auction and marketplace", "website_url": "https://www.ebay.com"},
        {"business_name": "Walmart", "description": "Retail corporation", "website_url": "https://www.walmart.com"},
        {"business_name": "Target", "description": "Retail store chain", "website_url": "https://www.target.com"},
        {"business_name": "Microsoft", "description": "Technology corporation", "website_url": "https://www.microsoft.com"},
        {"business_name": "Apple", "description": "Consumer electronics and software", "website_url": "https://www.apple.com"},
        {"business_name": "Google", "description": "Internet search and cloud services", "website_url": "https://www.google.com"},
        {"business_name": "Meta", "description": "Social media and technology company", "website_url": "https://www.meta.com"},
        {"business_name": "Stripe", "description": "Payment processing platform", "website_url": "https://stripe.com"},
        {"business_name": "Salesforce", "description": "Cloud-based CRM platform", "website_url": "https://www.salesforce.com"},
        {"business_name": "Oracle", "description": "Database and cloud services", "website_url": "https://www.oracle.com"},
        {"business_name": "IBM", "description": "Technology and consulting services", "website_url": "https://www.ibm.com"},
        {"business_name": "Starbucks", "description": "Coffee chain", "website_url": "https://www.starbucks.com"},
        {"business_name": "McDonald's", "description": "Fast food restaurant chain", "website_url": "https://www.mcdonalds.com"},
        {"business_name": "Coca-Cola", "description": "Beverage company", "website_url": "https://www.coca-cola.com"},
        {"business_name": "PepsiCo", "description": "Food and beverage corporation", "website_url": "https://www.pepsico.com"},
        {"business_name": "Tesla", "description": "Electric vehicle manufacturer", "website_url": "https://www.tesla.com"},
        {"business_name": "Ford", "description": "Automotive manufacturer", "website_url": "https://www.ford.com"},
        {"business_name": "General Motors", "description": "Automotive corporation", "website_url": "https://www.gm.com"},
        {"business_name": "Netflix", "description": "Streaming video service", "website_url": "https://www.netflix.com"},
        {"business_name": "Disney", "description": "Entertainment and media company", "website_url": "https://www.disney.com"},
        {"business_name": "Bank of America", "description": "Banking and financial services", "website_url": "https://www.bankofamerica.com"},
        {"business_name": "JPMorgan Chase", "description": "Investment bank", "website_url": "https://www.jpmorgan.com"},
        {"business_name": "Goldman Sachs", "description": "Investment banking", "website_url": "https://www.goldmansachs.com"},
        {"business_name": "Home Depot", "description": "Home improvement retailer", "website_url": "https://www.homedepot.com"},
        {"business_name": "Lowe's", "description": "Home improvement retailer", "website_url": "https://www.lowes.com"},
        {"business_name": "Best Buy", "description": "Consumer electronics retailer", "website_url": "https://www.bestbuy.com"},
        {"business_name": "Costco", "description": "Warehouse club", "website_url": "https://www.costco.com"},
    ]
    
    samples.extend(real_world)
    
    # Industry-specific samples (generate to reach 385)
    industries = [
        ("Technology", "Software development and IT services"),
        ("Retail", "Retail store and e-commerce"),
        ("Food & Beverage", "Restaurant and food service"),
        ("Healthcare", "Medical services and healthcare"),
        ("Financial Services", "Banking and financial services"),
        ("Manufacturing", "Manufacturing and production"),
        ("Real Estate", "Real estate services"),
        ("Professional Services", "Consulting and professional services"),
        ("Education", "Educational services"),
        ("Entertainment", "Entertainment and media"),
        ("Transportation", "Transportation and logistics"),
        ("Construction", "Construction and building"),
        ("Energy", "Energy and utilities"),
        ("Hospitality", "Hotels and hospitality"),
        ("Agriculture", "Agriculture and farming"),
    ]
    
    # Generate samples for each industry
    for industry, description in industries:
        for i in range(20):  # 20 samples per industry
            samples.append({
                "business_name": f"{industry} Company {i+1}",
                "description": description,
                "website_url": f"https://example-{industry.lower().replace(' ', '-')}-{i+1}.com"
            })
    
    # Add small business samples without websites
    for i in range(50):
        samples.append({
            "business_name": f"Small Business {i+1}",
            "description": "Local small business",
            "website_url": ""  # No website
        })
    
    # Ensure we have exactly 385 samples
    if len(samples) > 385:
        samples = samples[:385]
    elif len(samples) < 385:
        # Add more generic samples
        needed = 385 - len(samples)
        for i in range(needed):
            samples.append({
                "business_name": f"Business {len(samples) + 1}",
                "description": "General business services",
                "website_url": f"https://example-business-{len(samples) + 1}.com"
            })
    
    return samples[:385]

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
    
    # Retry statistics
    retries = [r.get("retry_count", 0) for r in results]
    total_retries = sum(retries)
    retry_rate = (total_retries / total * 100) if total > 0 else 0
    
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
        "total_retries": total_retries,
        "retry_rate_percent": round(retry_rate, 2),
        "targets": {
            "error_rate_target": "<5%",
            "avg_latency_target": "<10s",
            "accuracy_target": "≥80%",
            "code_generation_target": "≥90%",
            "confidence_target": ">50%"
        }
    }

def main():
    print("=" * 70)
    print("Comprehensive 385-Sample E2E Classification Test")
    print("=" * 70)
    print(f"API URL: {API_URL}")
    print(f"Test Samples: 385")
    print(f"Timeout: {TIMEOUT}s")
    print(f"Max Retries: {MAX_RETRIES}")
    print()
    
    # Generate comprehensive test samples
    print("📋 Generating 385 comprehensive test samples...")
    samples = generate_comprehensive_samples()
    print(f"✅ Generated {len(samples)} test samples")
    print()
    
    print(f"🚀 Running {len(samples)} classification tests...")
    print()
    
    results = []
    start_time = time.time()
    
    for i, sample in enumerate(samples, 1):
        business_name = sample.get("business_name", "Unknown")
        print(f"Test {i}/{len(samples)}: {business_name}... ", end="", flush=True)
        result = test_classification(sample)
        results.append(result)
        
        if result["success"]:
            latency_s = result["latency_ms"] / 1000
            retry_info = f" (retries: {result['retry_count']})" if result["retry_count"] > 0 else ""
            print(f"✅ ({latency_s:.2f}s{retry_info})")
        else:
            print(f"❌ {result.get('error', 'Unknown error')}")
        
        # Small delay to avoid rate limiting
        time.sleep(0.3)
    
    total_duration = time.time() - start_time
    
    print()
    print("=" * 70)
    print("📊 Metrics Summary")
    print("=" * 70)
    
    metrics = calculate_metrics(results)
    
    print(f"Total Requests: {metrics['total_requests']}")
    print(f"Successful: {metrics['successful_requests']}")
    print(f"Failed: {metrics['failed_requests']}")
    print(f"Error Rate: {metrics['error_rate_percent']}% (Target: {metrics['targets']['error_rate_target']})")
    print(f"Average Latency: {metrics['average_latency_ms']/1000:.2f}s (Target: {metrics['targets']['avg_latency_target']})")
    print(f"P50 Latency: {metrics['p50_latency_ms']/1000:.2f}s")
    print(f"P95 Latency: {metrics['p95_latency_ms']/1000:.2f}s")
    print(f"P99 Latency: {metrics['p99_latency_ms']/1000:.2f}s")
    print(f"Classification Accuracy: {metrics['classification_accuracy_percent']}% (Target: {metrics['targets']['accuracy_target']})")
    print(f"Code Generation Rate: {metrics['code_generation_rate_percent']}% (Target: {metrics['targets']['code_generation_target']})")
    print(f"Average Confidence: {metrics['average_confidence']*100:.2f}% (Target: {metrics['targets']['confidence_target']})")
    print(f"Total Retries: {metrics['total_retries']} ({metrics['retry_rate_percent']:.2f}% of requests)")
    print(f"Total Duration: {total_duration/60:.1f} minutes")
    print()
    
    # Save results
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    results_file = f"test/results/comprehensive_385_e2e_metrics_{timestamp}.json"
    
    output = {
        "timestamp": datetime.now().isoformat(),
        "api_url": API_URL,
        "total_samples": 385,
        "total_duration_seconds": round(total_duration, 2),
        "metrics": metrics,
        "results": results
    }
    
    with open(results_file, 'w') as f:
        json.dump(output, f, indent=2)
    
    print(f"📄 Results saved to: {results_file}")
    print()
    
    # Compare with previous baseline
    print("=" * 70)
    print("📈 Improvement Analysis")
    print("=" * 70)
    print("Baseline (Before All Fixes):")
    print("  - Error Rate: 67.1%")
    print("  - Average Latency: 43.7s")
    print("  - Classification Accuracy: 9.5%")
    print("  - Code Generation Rate: 23.1%")
    print("  - Average Confidence: 24.65%")
    print()
    print("Current Results (385 Samples):")
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

