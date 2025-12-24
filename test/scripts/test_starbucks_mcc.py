#!/usr/bin/env python3
"""
Quick test to verify MCC codes are generated for Starbucks Coffee
"""

import json
import os
import requests
import sys
from typing import Dict, Any
import urllib3
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

# Configuration
API_URL = os.getenv("CLASSIFICATION_SERVICE_URL", "https://classification-service-production.up.railway.app")
TIMEOUT = 120

def test_starbucks_classification():
    """Test Starbucks Coffee classification and verify MCC codes"""
    
    print("=" * 70)
    print("🧪 Testing Starbucks Coffee Classification - MCC Code Generation")
    print("=" * 70)
    print(f"API URL: {API_URL}\n")
    
    # Starbucks Coffee test sample
    test_data = {
        "business_name": "Starbucks Coffee",
        "description": "Coffee shop chain and retail coffee products",
        "website_url": "https://starbucks.com"
    }
    
    print(f"Test Data:")
    print(f"  Business Name: {test_data['business_name']}")
    print(f"  Description: {test_data['description']}")
    print(f"  Website: {test_data['website_url']}\n")
    
    print("Sending classification request...")
    
    try:
        # Add nocache parameter to bypass cache and get fresh results
        response = requests.post(
            f"{API_URL}/classify?nocache=true",
            json=test_data,
            timeout=TIMEOUT,
            verify=False
        )
        
        response.raise_for_status()
        result = response.json()
        
        print(f"✅ Request successful (HTTP {response.status_code})\n")
        
        # Extract classification data
        classification = result.get("classification", {})
        industry = classification.get("industry", result.get("primary_industry", "Unknown"))
        mcc_codes = classification.get("mcc_codes", [])
        naics_codes = classification.get("naics_codes", [])
        sic_codes = classification.get("sic_codes", [])
        explanation = classification.get("explanation", {})
        
        print("=" * 70)
        print("📊 Classification Results")
        print("=" * 70)
        print(f"Industry: {industry}")
        print(f"\nMCC Codes: {len(mcc_codes)}")
        if mcc_codes:
            for i, code in enumerate(mcc_codes[:3], 1):
                print(f"  {i}. {code.get('code', 'N/A')} - {code.get('description', 'N/A')} (confidence: {code.get('confidence', 0):.2f})")
        else:
            print("  ❌ No MCC codes generated")
        
        print(f"\nNAICS Codes: {len(naics_codes)}")
        if naics_codes:
            for i, code in enumerate(naics_codes[:3], 1):
                print(f"  {i}. {code.get('code', 'N/A')} - {code.get('description', 'N/A')} (confidence: {code.get('confidence', 0):.2f})")
        
        print(f"\nSIC Codes: {len(sic_codes)}")
        if sic_codes:
            for i, code in enumerate(sic_codes[:3], 1):
                print(f"  {i}. {code.get('code', 'N/A')} - {code.get('description', 'N/A')} (confidence: {code.get('confidence', 0):.2f})")
        
        # Check explanation
        if explanation:
            primary_reason = explanation.get("primary_reason", "N/A")
            print(f"\nExplanation: {primary_reason[:100]}...")
        
        print("\n" + "=" * 70)
        print("✅ Test Results")
        print("=" * 70)
        
        # Validation
        has_mcc = len(mcc_codes) > 0
        has_naics = len(naics_codes) > 0
        has_sic = len(sic_codes) > 0
        has_explanation = explanation is not None and explanation != {}
        
        print(f"MCC Codes Generated: {'✅ YES' if has_mcc else '❌ NO'} ({len(mcc_codes)} codes)")
        print(f"NAICS Codes Generated: {'✅ YES' if has_naics else '❌ NO'} ({len(naics_codes)} codes)")
        print(f"SIC Codes Generated: {'✅ YES' if has_sic else '❌ NO'} ({len(sic_codes)} codes)")
        print(f"Explanation Generated: {'✅ YES' if has_explanation else '❌ NO'}")
        
        # Check if all frontend requirements are met
        all_requirements = has_mcc and has_naics and has_sic and has_explanation
        print(f"\nAll Frontend Requirements Met: {'✅ YES' if all_requirements else '❌ NO'}")
        
        if has_mcc:
            print("\n🎉 SUCCESS: MCC codes are now being generated!")
            print(f"   Found {len(mcc_codes)} MCC code(s)")
        else:
            print("\n⚠️  WARNING: MCC codes are still missing")
            print("   This indicates the fix may need further investigation")
        
        # Save full response for analysis
        output_file = "test/results/starbucks_test_result.json"
        os.makedirs(os.path.dirname(output_file), exist_ok=True)
        with open(output_file, 'w') as f:
            json.dump(result, f, indent=2)
        print(f"\nFull response saved to: {output_file}")
        
        return has_mcc, result
        
    except requests.exceptions.RequestException as e:
        print(f"❌ Request failed: {e}")
        if hasattr(e, 'response') and e.response is not None:
            print(f"Response status: {e.response.status_code}")
            print(f"Response body: {e.response.text[:500]}")
        return False, None
    except Exception as e:
        print(f"❌ Error: {e}")
        import traceback
        traceback.print_exc()
        return False, None

if __name__ == "__main__":
    success, result = test_starbucks_classification()
    sys.exit(0 if success else 1)

