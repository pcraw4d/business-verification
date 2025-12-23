#!/usr/bin/env python3
"""
URL Validation Script for Test Data
Validates URLs before running tests to ensure they are accessible and DNS-resolvable.
This helps reduce test failures due to invalid URLs.
"""

import socket
import requests
import sys
import time
from typing import Dict, List, Tuple
from urllib.parse import urlparse
import concurrent.futures
from datetime import datetime

# Configuration
DNS_TIMEOUT = 2  # seconds
HTTP_TIMEOUT = 5  # seconds
MAX_WORKERS = 10  # concurrent validations

def validate_dns(url: str) -> Tuple[bool, str]:
    """Validate DNS resolution for a URL"""
    try:
        parsed = urlparse(url)
        hostname = parsed.hostname
        if not hostname:
            return False, "No hostname in URL"
        
        # Try DNS resolution with timeout
        socket.setdefaulttimeout(DNS_TIMEOUT)
        socket.gethostbyname(hostname)
        return True, "DNS resolved"
    except socket.timeout:
        return False, "DNS timeout"
    except socket.gaierror as e:
        return False, f"DNS error: {e}"
    except Exception as e:
        return False, f"DNS validation error: {e}"

def validate_http(url: str) -> Tuple[bool, str, int]:
    """Validate HTTP connectivity for a URL"""
    try:
        response = requests.get(
            url,
            timeout=HTTP_TIMEOUT,
            allow_redirects=True,
            verify=False,  # Skip SSL verification for speed
            headers={'User-Agent': 'Mozilla/5.0 (compatible; KYB-Platform-Test/1.0)'}
        )
        return True, f"HTTP {response.status_code}", response.status_code
    except requests.exceptions.Timeout:
        return False, "HTTP timeout", 0
    except requests.exceptions.ConnectionError as e:
        return False, f"Connection error: {str(e)[:100]}", 0
    except requests.exceptions.RequestException as e:
        return False, f"HTTP error: {str(e)[:100]}", 0
    except Exception as e:
        return False, f"HTTP validation error: {e}", 0

def validate_url(url: str) -> Dict:
    """Validate a single URL (DNS + HTTP)"""
    if not url or url.strip() == "":
        return {
            "url": url,
            "valid": False,
            "dns_valid": False,
            "http_valid": False,
            "error": "Empty URL",
            "status_code": 0
        }
    
    # Normalize URL
    url = url.strip()
    if not url.startswith(("http://", "https://")):
        url = "https://" + url
    
    # Validate DNS
    dns_valid, dns_message = validate_dns(url)
    
    if not dns_valid:
        return {
            "url": url,
            "valid": False,
            "dns_valid": False,
            "http_valid": False,
            "error": dns_message,
            "status_code": 0
        }
    
    # Validate HTTP
    http_valid, http_message, status_code = validate_http(url)
    
    return {
        "url": url,
        "valid": dns_valid and http_valid,
        "dns_valid": dns_valid,
        "http_valid": http_valid,
        "error": None if http_valid else http_message,
        "status_code": status_code,
        "dns_message": dns_message,
        "http_message": http_message
    }

def validate_urls(urls: List[str], parallel: bool = True) -> List[Dict]:
    """Validate multiple URLs"""
    results = []
    
    if parallel:
        with concurrent.futures.ThreadPoolExecutor(max_workers=MAX_WORKERS) as executor:
            future_to_url = {executor.submit(validate_url, url): url for url in urls}
            for future in concurrent.futures.as_completed(future_to_url):
                result = future.result()
                results.append(result)
    else:
        for url in urls:
            results.append(validate_url(url))
    
    return results

def filter_valid_urls(samples: List[Dict]) -> Tuple[List[Dict], List[Dict]]:
    """Filter samples into valid and invalid based on URL validation"""
    urls = [sample.get("website_url", "") for sample in samples]
    
    print(f"🔍 Validating {len(urls)} URLs...")
    validation_results = validate_urls(urls, parallel=True)
    
    valid_samples = []
    invalid_samples = []
    
    for sample, validation in zip(samples, validation_results):
        if validation["valid"]:
            valid_samples.append(sample)
        else:
            invalid_samples.append({
                "sample": sample,
                "validation": validation
            })
    
    return valid_samples, invalid_samples

def main():
    """Main function for CLI usage"""
    if len(sys.argv) < 2:
        print("Usage: python3 validate_test_urls.py <url1> [url2] ...")
        print("   or: python3 validate_test_urls.py --file <test_samples.json>")
        sys.exit(1)
    
    urls = []
    
    if sys.argv[1] == "--file":
        import json
        with open(sys.argv[2], 'r') as f:
            data = json.load(f)
            if isinstance(data, list):
                urls = [item.get("website_url", "") for item in data]
            elif isinstance(data, dict) and "samples" in data:
                urls = [item.get("website_url", "") for item in data["samples"]]
    else:
        urls = sys.argv[1:]
    
    print(f"Validating {len(urls)} URLs...")
    print()
    
    results = validate_urls(urls, parallel=True)
    
    valid_count = sum(1 for r in results if r["valid"])
    invalid_count = len(results) - valid_count
    
    print(f"✅ Valid URLs: {valid_count}")
    print(f"❌ Invalid URLs: {invalid_count}")
    print()
    
    if invalid_count > 0:
        print("Invalid URLs:")
        for r in results:
            if not r["valid"]:
                print(f"  ❌ {r['url']}: {r.get('error', 'Unknown error')}")
    
    return 0 if invalid_count == 0 else 1

if __name__ == "__main__":
    sys.exit(main())

