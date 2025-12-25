#!/usr/bin/env python3
"""
Direct Tesla Classification Test
Tests Tesla request directly to verify service behavior
"""

import json
import time
import requests
import sys
from datetime import datetime

# Configuration
API_URL = "https://classification-service-production.up.railway.app"
TIMEOUT = 150  # 150 seconds to allow for processing

# Tesla test data
TESLA_DATA = {
    "business_name": "Tesla Inc",
    "description": "Electric vehicle manufacturer and energy storage solutions",
    "website_url": "https://tesla.com"
}

def test_tesla_direct():
    """Test Tesla classification directly"""
    print("=" * 70)
    print("🧪 Direct Tesla Classification Test")
    print("=" * 70)
    print(f"API URL: {API_URL}")
    print(f"Timeout: {TIMEOUT}s")
    print(f"Test Data: {json.dumps(TESLA_DATA, indent=2)}")
    print()
    
    # Add cache bypass
    url = f"{API_URL}/v1/classify?nocache=true"
    
    print("🚀 Sending Tesla classification request...")
    print(f"URL: {url}")
    print()
    
    start_time = time.time()
    
    try:
        response = requests.post(
            url,
            json=TESLA_DATA,
            timeout=TIMEOUT,
            verify=False,  # Skip SSL verification for testing
            headers={"Content-Type": "application/json"}
        )
        
        elapsed_time = time.time() - start_time
        
        print(f"⏱️  Response Time: {elapsed_time:.2f}s")
        print(f"📊 HTTP Status: {response.status_code}")
        print()
        
        if response.status_code == 200:
            print("✅ SUCCESS: Request completed successfully")
            result = response.json()
            print(f"📝 Request ID: {result.get('request_id', 'N/A')}")
            print(f"🏭 Industry: {result.get('classification', {}).get('industry', 'N/A')}")
            
            # Check codes
            classification = result.get('classification', {})
            mcc_count = len(classification.get('mcc_codes', []))
            naics_count = len(classification.get('naics_codes', []))
            sic_count = len(classification.get('sic_codes', []))
            
            print(f"📊 Codes Generated:")
            print(f"   - MCC: {mcc_count}")
            print(f"   - NAICS: {naics_count}")
            print(f"   - SIC: {sic_count}")
            
            return True, elapsed_time, response.status_code, result
            
        else:
            print(f"❌ FAILED: HTTP {response.status_code}")
            try:
                error_data = response.json()
                print(f"📝 Error Message: {error_data.get('message', 'N/A')}")
                print(f"📝 Request ID: {error_data.get('request_id', 'N/A')}")
                print(f"📝 Full Response: {json.dumps(error_data, indent=2)}")
            except:
                print(f"📝 Response Body: {response.text[:500]}")
            
            return False, elapsed_time, response.status_code, None
            
    except requests.exceptions.Timeout:
        elapsed_time = time.time() - start_time
        print(f"⏱️  Timeout after: {elapsed_time:.2f}s")
        print("❌ FAILED: Request timed out")
        return False, elapsed_time, 0, None
        
    except requests.exceptions.RequestException as e:
        elapsed_time = time.time() - start_time
        print(f"⏱️  Error after: {elapsed_time:.2f}s")
        print(f"❌ FAILED: Request error - {str(e)}")
        return False, elapsed_time, 0, None

if __name__ == "__main__":
    print(f"🕐 Test started at: {datetime.now().isoformat()}")
    print()
    
    success, elapsed, status_code, result = test_tesla_direct()
    
    print()
    print("=" * 70)
    print("📊 Test Summary")
    print("=" * 70)
    print(f"Status: {'✅ SUCCESS' if success else '❌ FAILED'}")
    print(f"Response Time: {elapsed:.2f}s")
    print(f"HTTP Status: {status_code}")
    print(f"Test completed at: {datetime.now().isoformat()}")
    print()
    
    if not success:
        print("🔍 Analysis:")
        if elapsed >= 30 and elapsed < 35:
            print("   - Failed around 30s mark (Railway platform timeout threshold)")
        elif elapsed >= 40 and elapsed < 45:
            print("   - Failed around 40s mark (service processing delay)")
        elif elapsed >= 60:
            print("   - Failed after 60s (extended processing time)")
        else:
            print(f"   - Failed at {elapsed:.2f}s (unusual timing)")
        
        print()
        print("💡 Recommendations:")
        print("   1. Check Railway logs for request processing details")
        print("   2. Verify service health and resource usage")
        print("   3. Check if Tesla website scraping is causing delays")
        print("   4. Review service timeout configurations")
    
    sys.exit(0 if success else 1)

