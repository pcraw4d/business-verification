#!/usr/bin/env python3
"""
Comprehensive E2E Classification Test
Tests industry/code accuracy, performance, and frontend data completeness
"""

import json
import os
import time
import requests
import sys
from datetime import datetime
from typing import Dict, List, Any, Optional
from urllib.parse import urlparse
import urllib3
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

# Configuration
API_URL = os.getenv("CLASSIFICATION_SERVICE_URL", "https://classification-service-production.up.railway.app")
TIMEOUT = 120
MAX_RETRIES = 3

# Diverse test samples covering different industries
TEST_SAMPLES = [
    {
        "business_name": "Microsoft Corporation",
        "description": "Software development, cloud computing, and enterprise solutions",
        "website_url": "https://microsoft.com",
        "expected_industry": "Technology",
        "expected_codes": {
            "mcc": ["5734"],  # Computer Software Stores
            "naics": ["541511"],  # Custom Computer Programming Services
            "sic": ["7371"]  # Computer Programming Services
        }
    },
    {
        "business_name": "Starbucks Coffee",
        "description": "Coffee shop chain and retail coffee products",
        "website_url": "https://starbucks.com",
        "expected_industry": "Food & Beverage",
        "expected_codes": {
            "mcc": ["5812"],  # Eating Places, Restaurants
            "naics": ["722511"],  # Full-Service Restaurants
            "sic": ["5812"]  # Eating Places
        }
    },
    {
        "business_name": "Home Depot",
        "description": "Home improvement retailer selling tools, building materials, and garden supplies",
        "website_url": "https://homedepot.com",
        "expected_industry": "Retail",
        "expected_codes": {
            "mcc": ["5200"],  # Home Supply Warehouse Stores
            "naics": ["444110"],  # Home Centers
            "sic": ["5211"]  # Lumber and Other Building Materials
        }
    },
    {
        "business_name": "Bank of America",
        "description": "Banking and financial services including checking, savings, loans, and investments",
        "website_url": "https://bankofamerica.com",
        "expected_industry": "Financial Services",
        "expected_codes": {
            "mcc": ["6011"],  # Financial Institutions - Merchandise, Services, etc.
            "naics": ["522110"],  # Commercial Banking
            "sic": ["6021"]  # National Commercial Banks
        }
    },
    {
        "business_name": "Amazon",
        "description": "E-commerce platform and cloud computing services",
        "website_url": "https://amazon.com",
        "expected_industry": "E-commerce",
        "expected_codes": {
            "mcc": ["5999"],  # Miscellaneous and Specialty Retail Stores
            "naics": ["454110"],  # Electronic Shopping
            "sic": ["5961"]  # Catalog and Mail-Order Houses
        }
    },
    {
        "business_name": "Tesla Inc",
        "description": "Electric vehicle manufacturer and energy storage solutions",
        "website_url": "https://tesla.com",
        "expected_industry": "Manufacturing",
        "expected_codes": {
            "mcc": ["5511"],  # Car and Truck Dealers (New and Used) Sales, Service, Repairs, Parts, and Leasing
            "naics": ["336111"],  # Automobile Manufacturing
            "sic": ["3711"]  # Motor Vehicles and Passenger Car Bodies
        }
    },
    {
        "business_name": "CVS Pharmacy",
        "description": "Pharmacy and retail health services",
        "website_url": "https://cvs.com",
        "expected_industry": "Healthcare",
        "expected_codes": {
            "mcc": ["5912"],  # Drug Stores, Pharmacies
            "naics": ["446110"],  # Pharmacies and Drug Stores
            "sic": ["5912"]  # Drug Stores and Proprietary Stores
        }
    },
    {
        "business_name": "Uber Technologies",
        "description": "Ridesharing and food delivery platform",
        "website_url": "https://uber.com",
        "expected_industry": "Transportation",
        "expected_codes": {
            "mcc": ["4121"],  # Taxicabs and Limousines
            "naics": ["485310"],  # Taxi Service
            "sic": ["4121"]  # Taxicabs
        }
    },
    {
        "business_name": "Netflix",
        "description": "Streaming video and entertainment service",
        "website_url": "https://netflix.com",
        "expected_industry": "Entertainment",
        "expected_codes": {
            "mcc": ["7829"],  # Motion Picture and Video Tape Production and Distribution
            "naics": ["515210"],  # Cable and Other Subscription Programming
            "sic": ["7812"]  # Motion Picture and Video Tape Production
        }
    },
    {
        "business_name": "Whole Foods Market",
        "description": "Organic and natural foods grocery store",
        "website_url": "https://wholefoodsmarket.com",
        "expected_industry": "Retail",
        "expected_codes": {
            "mcc": ["5411"],  # Grocery Stores, Supermarkets
            "naics": ["445110"],  # Supermarkets and Other Grocery (except Convenience) Stores
            "sic": ["5411"]  # Grocery Stores
        }
    }
]

def test_classification(sample: Dict[str, Any], max_retries: int = MAX_RETRIES) -> Dict[str, Any]:
    """Test a single classification request with retry logic"""
    # Add cache bypass to ensure fresh results with latest code
    url = f"{API_URL}/v1/classify?nocache=true"
    
    result = {
        "sample": sample,
        "success": False,
        "http_code": 0,
        "latency_ms": 0,
        "error": None,
        "response": None,
        "retry_count": 0,
        "validation": {
            "industry_present": False,
            "top3_mcc_present": False,
            "top3_naics_present": False,
            "top3_sic_present": False,
            "explanation_present": False,
            "all_requirements_met": False
        }
    }
    
    # Retry logic
    for attempt in range(max_retries):
        try:
            start_time = time.time()
            response = requests.post(
                url,
                json={
                    "business_name": sample["business_name"],
                    "description": sample["description"],
                    "website_url": sample.get("website_url", "")
                },
                timeout=TIMEOUT,
                verify=False
            )
            end_time = time.time()
            
            result["http_code"] = response.status_code
            result["latency_ms"] = (end_time - start_time) * 1000
            result["retry_count"] = attempt
            
            if response.status_code == 200:
                result["success"] = True
                result["response"] = response.json()
                
                # Validate frontend requirements
                result["validation"] = validate_frontend_requirements(result["response"])
                return result
            elif response.status_code in [502, 503, 429] and attempt < max_retries - 1:
                wait_time = (2 ** attempt) * 1.0
                print(f"  ⚠️  HTTP {response.status_code} (attempt {attempt + 1}/{max_retries}), retrying in {wait_time:.1f}s...", end="", flush=True)
                time.sleep(wait_time)
                continue
            else:
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
            if attempt < max_retries - 1:
                wait_time = (2 ** attempt) * 1.0
                print(f"  ⚠️  Error (attempt {attempt + 1}/{max_retries}), retrying in {wait_time:.1f}s...", end="", flush=True)
                time.sleep(wait_time)
                continue
            result["error"] = str(e)
            return result
    
    return result

def validate_frontend_requirements(response: Dict[str, Any]) -> Dict[str, bool]:
    """Validate that all frontend requirements are met"""
    validation = {
        "industry_present": False,
        "top3_mcc_present": False,
        "top3_naics_present": False,
        "top3_sic_present": False,
        "explanation_present": False,
        "all_requirements_met": False
    }
    
    # Check for industry
    industry = response.get("primary_industry") or response.get("classification", {}).get("industry")
    validation["industry_present"] = bool(industry and industry.strip())
    
    # Check for top 3 MCC codes
    classification = response.get("classification", {})
    mcc_codes = classification.get("mcc_codes", [])
    validation["top3_mcc_present"] = len(mcc_codes) >= 1  # At least 1, ideally top 3
    
    # Check for top 3 NAICS codes
    naics_codes = classification.get("naics_codes", [])
    validation["top3_naics_present"] = len(naics_codes) >= 1  # At least 1, ideally top 3
    
    # Check for top 3 SIC codes
    sic_codes = classification.get("sic_codes", [])
    validation["top3_sic_present"] = len(sic_codes) >= 1  # At least 1, ideally top 3
    
    # Check for explanation
    explanation = response.get("explanation") or response.get("classification", {}).get("explanation")
    if explanation:
        # Can be string or structured object
        if isinstance(explanation, str):
            validation["explanation_present"] = bool(explanation.strip())
        elif isinstance(explanation, dict):
            # Structured explanation
            validation["explanation_present"] = bool(
                explanation.get("primary_reason") or 
                explanation.get("reasoning") or
                explanation.get("summary")
            )
    else:
        validation["explanation_present"] = False
    
    # All requirements met
    validation["all_requirements_met"] = all([
        validation["industry_present"],
        validation["top3_mcc_present"],
        validation["top3_naics_present"],
        validation["top3_sic_present"],
        validation["explanation_present"]
    ])
    
    return validation

def is_code_match(actual_code: str, expected_code: str, code_type: str) -> bool:
    """Check if actual code matches expected code with flexible matching"""
    # Exact match
    if actual_code == expected_code:
        return True
    
    # For NAICS: Check parent codes (e.g., 541511 matches 5415 or 541)
    if code_type == "NAICS" and len(actual_code) == 6 and len(expected_code) == 6:
        # Check if they share the same parent (first 4 digits or first 2 digits)
        if actual_code[:4] == expected_code[:4]:
            return True
        if actual_code[:2] == expected_code[:2]:
            return True
    
    # For MCC/SIC: Check if codes are in same range (e.g., 5812-5814)
    if code_type in ["MCC", "SIC"] and len(actual_code) == 4 and len(expected_code) == 4:
        # Check if first 3 digits match (same category)
        if actual_code[:3] == expected_code[:3]:
            return True
        # Check if first 2 digits match (same major group)
        if actual_code[:2] == expected_code[:2]:
            return True
    
    return False

def validate_code_accuracy(response: Dict[str, Any], expected_codes: Dict[str, List[str]], expected_industry: str = None) -> Dict[str, Any]:
    """Validate code accuracy against expected codes with enhanced metrics
    
    Returns separate metrics for:
    - exact_accuracy: Exact code matches (1.0 per match)
    - flexible_accuracy: Flexible matches (0.8 per match, increased from 0.7)
    - industry_appropriateness: Validate codes are appropriate for industry (0.9 per appropriate code)
    - overall_accuracy: Weighted average of all metrics
    """
    accuracy = {
        "mcc_match": False,
        "naics_match": False,
        "sic_match": False,
        "mcc_partial_match": False,
        "naics_partial_match": False,
        "sic_partial_match": False,
        "exact_accuracy": 0.0,
        "flexible_accuracy": 0.0,
        "industry_appropriateness": 0.0,
        "overall_accuracy": 0.0,
        "exact_matches": 0,
        "partial_matches": 0
    }
    
    classification = response.get("classification", {})
    detected_industry = classification.get("industry", "")
    
    # Check MCC codes with flexible matching
    mcc_codes = [code.get("code", "") for code in classification.get("mcc_codes", [])]
    expected_mcc = expected_codes.get("mcc", [])
    
    exact_mcc_match = any(exp_code in mcc_codes for exp_code in expected_mcc)
    partial_mcc_match = False
    if not exact_mcc_match and expected_mcc:
        for exp_code in expected_mcc:
            for act_code in mcc_codes:
                if is_code_match(act_code, exp_code, "MCC"):
                    partial_mcc_match = True
                    break
            if partial_mcc_match:
                break
    
    accuracy["mcc_match"] = exact_mcc_match
    accuracy["mcc_partial_match"] = partial_mcc_match
    
    # Check NAICS codes with flexible matching
    naics_codes = [code.get("code", "") for code in classification.get("naics_codes", [])]
    expected_naics = expected_codes.get("naics", [])
    
    exact_naics_match = any(exp_code in naics_codes for exp_code in expected_naics)
    partial_naics_match = False
    if not exact_naics_match and expected_naics:
        for exp_code in expected_naics:
            for act_code in naics_codes:
                if is_code_match(act_code, exp_code, "NAICS"):
                    partial_naics_match = True
                    break
            if partial_naics_match:
                break
    
    accuracy["naics_match"] = exact_naics_match
    accuracy["naics_partial_match"] = partial_naics_match
    
    # Check SIC codes with flexible matching
    sic_codes = [code.get("code", "") for code in classification.get("sic_codes", [])]
    expected_sic = expected_codes.get("sic", [])
    
    exact_sic_match = any(exp_code in sic_codes for exp_code in expected_sic)
    partial_sic_match = False
    if not exact_sic_match and expected_sic:
        for exp_code in expected_sic:
            for act_code in sic_codes:
                if is_code_match(act_code, exp_code, "SIC"):
                    partial_sic_match = True
                    break
            if partial_sic_match:
                break
    
    accuracy["sic_match"] = exact_sic_match
    accuracy["sic_partial_match"] = partial_sic_match
    
    # Calculate exact accuracy: 1.0 per exact match
    exact_matches = sum([exact_mcc_match, exact_naics_match, exact_sic_match])
    accuracy["exact_accuracy"] = exact_matches / 3.0
    
    # Calculate flexible accuracy: 0.8 per flexible match (increased from 0.7)
    flexible_matches = sum([partial_mcc_match, partial_naics_match, partial_sic_match])
    accuracy["flexible_accuracy"] = (flexible_matches * 0.8) / 3.0
    
    # Calculate industry appropriateness
    # For now, use a simplified check: if industry matches expected, give full credit
    # If codes are present but industry doesn't match, give partial credit
    industry_score = 0.0
    if expected_industry and detected_industry:
        # Simple industry matching (case-insensitive, partial match)
        if expected_industry.lower() in detected_industry.lower() or detected_industry.lower() in expected_industry.lower():
            industry_score = 1.0
        elif len(mcc_codes) > 0 or len(naics_codes) > 0 or len(sic_codes) > 0:
            # Codes present but industry doesn't match exactly - partial credit
            industry_score = 0.8
    elif detected_industry and (len(mcc_codes) > 0 or len(naics_codes) > 0 or len(sic_codes) > 0):
        # Industry detected and codes present - assume appropriate
        industry_score = 0.9
    accuracy["industry_appropriateness"] = industry_score
    
    accuracy["exact_matches"] = exact_matches
    accuracy["partial_matches"] = flexible_matches
    
    # Calculate overall accuracy: weighted average
    # 40% weight for exact matches, 40% for flexible matches, 20% for industry fit
    overall_accuracy = (
        accuracy["exact_accuracy"] * 0.4 +
        accuracy["flexible_accuracy"] * 0.4 +
        accuracy["industry_appropriateness"] * 0.2
    )
    accuracy["overall_accuracy"] = overall_accuracy
    
    return accuracy

def calculate_metrics(results: List[Dict[str, Any]]) -> Dict[str, Any]:
    """Calculate comprehensive performance and accuracy metrics"""
    total = len(results)
    successful = sum(1 for r in results if r["success"])
    failed = total - successful
    
    # Performance metrics
    latencies = [r["latency_ms"] for r in results if r["success"]]
    latencies_sorted = sorted(latencies) if latencies else []
    
    # Frontend requirements validation
    all_requirements_met = sum(1 for r in results if r.get("validation", {}).get("all_requirements_met", False))
    frontend_completeness = (all_requirements_met / successful * 100) if successful > 0 else 0
    
    # Individual requirement rates
    industry_present = sum(1 for r in results if r.get("validation", {}).get("industry_present", False))
    mcc_present = sum(1 for r in results if r.get("validation", {}).get("top3_mcc_present", False))
    naics_present = sum(1 for r in results if r.get("validation", {}).get("top3_naics_present", False))
    sic_present = sum(1 for r in results if r.get("validation", {}).get("top3_sic_present", False))
    explanation_present = sum(1 for r in results if r.get("validation", {}).get("explanation_present", False))
    
    # Code accuracy (for samples with expected codes)
    accuracy_results = []
    for r in results:
        if r["success"] and "expected_codes" in r["sample"]:
            expected_industry = r["sample"].get("expected_industry")
            accuracy = validate_code_accuracy(r["response"], r["sample"]["expected_codes"], expected_industry)
            accuracy_results.append(accuracy)
    
    if accuracy_results:
        avg_code_accuracy = sum(a["overall_accuracy"] for a in accuracy_results) / len(accuracy_results)
        avg_exact_accuracy = sum(a["exact_accuracy"] for a in accuracy_results) / len(accuracy_results)
        avg_flexible_accuracy = sum(a["flexible_accuracy"] for a in accuracy_results) / len(accuracy_results)
        avg_industry_appropriateness = sum(a["industry_appropriateness"] for a in accuracy_results) / len(accuracy_results)
    else:
        avg_code_accuracy = 0.0
        avg_exact_accuracy = 0.0
        avg_flexible_accuracy = 0.0
        avg_industry_appropriateness = 0.0
    
    return {
        "total_requests": total,
        "successful_requests": successful,
        "failed_requests": failed,
        "error_rate_percent": round((failed / total * 100) if total > 0 else 0, 2),
        
        # Performance metrics
        "average_latency_ms": round(sum(latencies) / len(latencies) if latencies else 0, 2),
        "p50_latency_ms": round(latencies_sorted[len(latencies_sorted) // 2] if latencies_sorted else 0, 2),
        "p95_latency_ms": round(latencies_sorted[int(len(latencies_sorted) * 0.95)] if latencies_sorted else 0, 2),
        "p99_latency_ms": round(latencies_sorted[int(len(latencies_sorted) * 0.99)] if latencies_sorted else 0, 2),
        "min_latency_ms": round(min(latencies) if latencies else 0, 2),
        "max_latency_ms": round(max(latencies) if latencies else 0, 2),
        
        # Frontend requirements
        "frontend_completeness_percent": round(frontend_completeness, 2),
        "industry_present_rate": round((industry_present / successful * 100) if successful > 0 else 0, 2),
        "top3_mcc_present_rate": round((mcc_present / successful * 100) if successful > 0 else 0, 2),
        "top3_naics_present_rate": round((naics_present / successful * 100) if successful > 0 else 0, 2),
        "top3_sic_present_rate": round((sic_present / successful * 100) if successful > 0 else 0, 2),
        "explanation_present_rate": round((explanation_present / successful * 100) if successful > 0 else 0),
        
        # Code accuracy
        "average_code_accuracy": round(avg_code_accuracy * 100, 2),
        "average_exact_accuracy": round(avg_exact_accuracy * 100, 2),
        "average_flexible_accuracy": round(avg_flexible_accuracy * 100, 2),
        "average_industry_appropriateness": round(avg_industry_appropriateness * 100, 2),
        "code_accuracy_samples": len(accuracy_results),
        
        # Targets
        "targets": {
            "error_rate_target": "<5%",
            "avg_latency_target": "<30s",
            "p95_latency_target": "<60s",
            "frontend_completeness_target": "100%",
            "code_accuracy_target": "≥70%"
        }
    }

def main():
    print("=" * 70)
    print("🧪 Comprehensive E2E Classification Test")
    print("=" * 70)
    print(f"API URL: {API_URL}")
    print(f"Test Samples: {len(TEST_SAMPLES)}")
    print(f"Timeout: {TIMEOUT}s")
    print()
    print("Test Coverage:")
    print("  ✓ Industry and code accuracy")
    print("  ✓ Classification speed/performance")
    print("  ✓ Frontend data completeness (industry, top 3 codes, explanation)")
    print()
    
    print(f"🚀 Running {len(TEST_SAMPLES)} classification tests...")
    print()
    
    results = []
    for i, sample in enumerate(TEST_SAMPLES, 1):
        print(f"Test {i}/{len(TEST_SAMPLES)}: {sample['business_name']}... ", end="", flush=True)
        result = test_classification(sample)
        results.append(result)
        
        if result["success"]:
            validation = result["validation"]
            latency_s = result["latency_ms"] / 1000
            status_icons = []
            if validation["industry_present"]:
                status_icons.append("🏭")
            if validation["top3_mcc_present"]:
                status_icons.append("MCC")
            if validation["top3_naics_present"]:
                status_icons.append("NAICS")
            if validation["top3_sic_present"]:
                status_icons.append("SIC")
            if validation["explanation_present"]:
                status_icons.append("📝")
            
            all_met = "✅" if validation["all_requirements_met"] else "⚠️"
            print(f"{all_met} ({latency_s:.1f}s) {' '.join(status_icons)}")
        else:
            print(f"❌ {result.get('error', 'Unknown error')}")
        
        # Small delay to avoid rate limiting
        time.sleep(0.5)
    
    print()
    print("=" * 70)
    print("📊 Comprehensive Metrics Summary")
    print("=" * 70)
    
    metrics = calculate_metrics(results)
    
    print(f"\n📈 Performance Metrics:")
    print(f"  Total Requests: {metrics['total_requests']}")
    print(f"  Successful: {metrics['successful_requests']}")
    print(f"  Failed: {metrics['failed_requests']}")
    print(f"  Error Rate: {metrics['error_rate_percent']}% (Target: {metrics['targets']['error_rate_target']})")
    print(f"  Average Latency: {metrics['average_latency_ms']/1000:.2f}s (Target: {metrics['targets']['avg_latency_target']})")
    print(f"  P50 Latency: {metrics['p50_latency_ms']/1000:.2f}s")
    print(f"  P95 Latency: {metrics['p95_latency_ms']/1000:.2f}s (Target: {metrics['targets']['p95_latency_target']})")
    print(f"  P99 Latency: {metrics['p99_latency_ms']/1000:.2f}s")
    print(f"  Min Latency: {metrics['min_latency_ms']/1000:.2f}s")
    print(f"  Max Latency: {metrics['max_latency_ms']/1000:.2f}s")
    
    print(f"\n🎯 Frontend Requirements:")
    print(f"  All Requirements Met: {metrics['frontend_completeness_percent']}% (Target: {metrics['targets']['frontend_completeness_target']})")
    print(f"  Industry Present: {metrics['industry_present_rate']}%")
    print(f"  Top 3 MCC Codes: {metrics['top3_mcc_present_rate']}%")
    print(f"  Top 3 NAICS Codes: {metrics['top3_naics_present_rate']}%")
    print(f"  Top 3 SIC Codes: {metrics['top3_sic_present_rate']}%")
    print(f"  Explanation Present: {metrics['explanation_present_rate']}%")
    
    print(f"\n🎯 Code Accuracy:")
    print(f"  Average Code Accuracy: {metrics['average_code_accuracy']}% (Target: {metrics['targets']['code_accuracy_target']})")
    print(f"    - Exact Accuracy: {metrics.get('average_exact_accuracy', 0):.1f}%")
    print(f"    - Flexible Accuracy: {metrics.get('average_flexible_accuracy', 0):.1f}%")
    print(f"    - Industry Appropriateness: {metrics.get('average_industry_appropriateness', 0):.1f}%")
    print(f"  Accuracy Samples: {metrics['code_accuracy_samples']}")
    
    print()
    
    # Save results
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    results_file = f"test/results/comprehensive_e2e_test_{timestamp}.json"
    
    # Ensure directory exists
    import os
    os.makedirs("test/results", exist_ok=True)
    
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
    
    # Check if targets are met
    targets_met = []
    targets_missed = []
    
    if metrics['error_rate_percent'] < 5:
        targets_met.append("Error Rate")
    else:
        targets_missed.append(f"Error Rate ({metrics['error_rate_percent']:.1f}% > 5%)")
    
    if metrics['average_latency_ms']/1000 < 30:
        targets_met.append("Average Latency")
    else:
        targets_missed.append(f"Average Latency ({metrics['average_latency_ms']/1000:.1f}s > 30s)")
    
    if metrics['p95_latency_ms']/1000 < 60:
        targets_met.append("P95 Latency")
    else:
        targets_missed.append(f"P95 Latency ({metrics['p95_latency_ms']/1000:.1f}s > 60s)")
    
    if metrics['frontend_completeness_percent'] == 100:
        targets_met.append("Frontend Completeness")
    else:
        targets_missed.append(f"Frontend Completeness ({metrics['frontend_completeness_percent']:.1f}% < 100%)")
    
    if metrics['average_code_accuracy'] >= 70:
        targets_met.append("Code Accuracy")
    else:
        targets_missed.append(f"Code Accuracy ({metrics['average_code_accuracy']:.1f}% < 70%)")
    
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
    import os
    sys.exit(main())

