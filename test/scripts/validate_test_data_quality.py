#!/usr/bin/env python3
"""
Test Data Quality Validation Script
Validates test data files for:
- Malformed URLs
- Missing expected results
- Invalid code formats (MCC, NAICS, SIC)
- Required fields
"""

import json
import sys
import re
from typing import Dict, List, Any, Tuple
from pathlib import Path

# Import URL validation from existing script
sys.path.insert(0, str(Path(__file__).parent))
from validate_test_urls import validate_url, validate_urls

# Validation patterns
MCC_PATTERN = re.compile(r'^\d{4}$')  # 4-digit MCC codes
NAICS_PATTERN = re.compile(r'^\d{2,6}$')  # 2-6 digit NAICS codes
SIC_PATTERN = re.compile(r'^\d{4}$')  # 4-digit SIC codes

def validate_mcc_code(code: Any) -> Tuple[bool, str]:
    """Validate MCC code format"""
    if code is None:
        return True, "No MCC code (optional)"
    
    code_str = str(code).strip()
    if not code_str:
        return True, "Empty MCC code (optional)"
    
    if MCC_PATTERN.match(code_str):
        return True, "Valid MCC code"
    return False, f"Invalid MCC format: {code_str} (expected 4 digits)"

def validate_naics_code(code: Any) -> Tuple[bool, str]:
    """Validate NAICS code format"""
    if code is None:
        return True, "No NAICS code (optional)"
    
    code_str = str(code).strip()
    if not code_str:
        return True, "Empty NAICS code (optional)"
    
    if NAICS_PATTERN.match(code_str):
        return True, "Valid NAICS code"
    return False, f"Invalid NAICS format: {code_str} (expected 2-6 digits)"

def validate_sic_code(code: Any) -> Tuple[bool, str]:
    """Validate SIC code format"""
    if code is None:
        return True, "No SIC code (optional)"
    
    code_str = str(code).strip()
    if not code_str:
        return True, "Empty SIC code (optional)"
    
    if SIC_PATTERN.match(code_str):
        return True, "Valid SIC code"
    return False, f"Invalid SIC format: {code_str} (expected 4 digits)"

def validate_sample(sample: Dict[str, Any], index: int, validate_urls: bool = False) -> Dict[str, Any]:
    """Validate a single test sample"""
    errors = []
    warnings = []
    
    # Required fields
    if "business_name" not in sample or not sample["business_name"]:
        errors.append("Missing or empty 'business_name' field")
    
    # URL validation (optional, can be slow)
    website_url = sample.get("website_url", "")
    if website_url:
        if validate_urls:
            url_validation = validate_url(website_url)
            if not url_validation["valid"]:
                warnings.append(f"Invalid URL: {url_validation.get('error', 'Unknown error')}")
        else:
            # Basic URL format check without network validation
            if not website_url.startswith(("http://", "https://")):
                warnings.append(f"URL missing protocol: {website_url}")
    else:
        warnings.append("No website_url provided (will use description-only classification)")
    
    # Expected results validation (optional but recommended)
    if "expected_results" in sample:
        expected = sample["expected_results"]
        
        # Validate MCC codes
        if "mcc_codes" in expected:
            for mcc in expected["mcc_codes"]:
                if isinstance(mcc, dict):
                    code = mcc.get("code", "")
                else:
                    code = mcc
                valid, msg = validate_mcc_code(code)
                if not valid:
                    errors.append(f"MCC code validation: {msg}")
        
        # Validate NAICS codes
        if "naics_codes" in expected:
            for naics in expected["naics_codes"]:
                if isinstance(naics, dict):
                    code = naics.get("code", "")
                else:
                    code = naics
                valid, msg = validate_naics_code(code)
                if not valid:
                    errors.append(f"NAICS code validation: {msg}")
        
        # Validate SIC codes
        if "sic_codes" in expected:
            for sic in expected["sic_codes"]:
                if isinstance(sic, dict):
                    code = sic.get("code", "")
                else:
                    code = sic
                valid, msg = validate_sic_code(code)
                if not valid:
                    errors.append(f"SIC code validation: {msg}")
    else:
        warnings.append("No 'expected_results' field (optional but recommended for validation)")
    
    return {
        "index": index,
        "sample": sample.get("business_name", f"Sample {index}"),
        "valid": len(errors) == 0,
        "errors": errors,
        "warnings": warnings
    }

def validate_test_data_file(file_path: str, validate_urls: bool = False) -> Dict[str, Any]:
    """Validate a test data JSON file"""
    print(f"📋 Validating test data file: {file_path}")
    if validate_urls:
        print("⚠️  URL validation enabled (may be slow)")
    print("=" * 70)
    
    try:
        with open(file_path, 'r') as f:
            data = json.load(f)
    except FileNotFoundError:
        return {
            "valid": False,
            "error": f"File not found: {file_path}",
            "samples": [],
            "summary": {}
        }
    except json.JSONDecodeError as e:
        return {
            "valid": False,
            "error": f"Invalid JSON: {e}",
            "samples": [],
            "summary": {}
        }
    
    # Extract samples from different possible structures
    samples = []
    if isinstance(data, list):
        samples = data
    elif isinstance(data, dict):
        if "samples" in data:
            samples = data["samples"]
        elif "test_businesses" in data:
            samples = data["test_businesses"]
        elif "test_samples" in data:
            samples = data["test_samples"]
        else:
            # Try to find any list in the dict
            for key, value in data.items():
                if isinstance(value, list):
                    samples = value
                    break
    
    if not samples:
        return {
            "valid": False,
            "error": "No samples found in file",
            "samples": [],
            "summary": {}
        }
    
    print(f"Found {len(samples)} samples\n")
    
    # Validate each sample
    validation_results = []
    for i, sample in enumerate(samples):
        result = validate_sample(sample, i, validate_urls=validate_urls)
        validation_results.append(result)
    
    # Calculate summary
    valid_count = sum(1 for r in validation_results if r["valid"])
    invalid_count = len(validation_results) - valid_count
    total_errors = sum(len(r["errors"]) for r in validation_results)
    total_warnings = sum(len(r["warnings"]) for r in validation_results)
    
    summary = {
        "total_samples": len(samples),
        "valid_samples": valid_count,
        "invalid_samples": invalid_count,
        "total_errors": total_errors,
        "total_warnings": total_warnings,
        "error_rate": (invalid_count / len(samples) * 100) if samples else 0
    }
    
    # Print results
    print(f"✅ Valid samples: {valid_count}")
    print(f"❌ Invalid samples: {invalid_count}")
    print(f"⚠️  Total warnings: {total_warnings}")
    print(f"🚨 Total errors: {total_errors}")
    print()
    
    if invalid_count > 0 or total_errors > 0:
        print("Invalid samples:")
        for result in validation_results:
            if not result["valid"] or result["errors"]:
                print(f"\n  Sample {result['index']}: {result['sample']}")
                for error in result["errors"]:
                    print(f"    ❌ {error}")
                for warning in result["warnings"]:
                    print(f"    ⚠️  {warning}")
    
    return {
        "valid": invalid_count == 0 and total_errors == 0,
        "file_path": file_path,
        "samples": validation_results,
        "summary": summary
    }

def main():
    """Main function"""
    if len(sys.argv) < 2:
        print("Usage: python3 validate_test_data_quality.py [--validate-urls] <test_data.json> [test_data2.json ...]")
        print("\nValidates test data files for:")
        print("  - Malformed URLs (optional, use --validate-urls)")
        print("  - Missing required fields")
        print("  - Invalid code formats (MCC, NAICS, SIC)")
        print("  - Missing expected results")
        sys.exit(1)
    
    validate_urls = "--validate-urls" in sys.argv
    files = [f for f in sys.argv[1:] if f != "--validate-urls"]
    all_results = []
    
    for file_path in files:
        result = validate_test_data_file(file_path, validate_urls=validate_urls)
        all_results.append(result)
        print()
    
    # Overall summary
    total_files = len(files)
    valid_files = sum(1 for r in all_results if r.get("valid", False))
    total_samples = sum(r["summary"].get("total_samples", 0) for r in all_results)
    total_valid_samples = sum(r["summary"].get("valid_samples", 0) for r in all_results)
    total_errors = sum(r["summary"].get("total_errors", 0) for r in all_results)
    
    print("=" * 70)
    print("OVERALL SUMMARY")
    print("=" * 70)
    print(f"Files validated: {total_files}")
    print(f"Valid files: {valid_files}")
    print(f"Total samples: {total_samples}")
    print(f"Valid samples: {total_valid_samples}")
    print(f"Total errors: {total_errors}")
    
    if total_errors == 0:
        print("\n✅ All test data is valid!")
        return 0
    else:
        print(f"\n❌ Found {total_errors} errors in test data")
        return 1

if __name__ == "__main__":
    sys.exit(main())

