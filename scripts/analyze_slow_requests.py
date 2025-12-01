#!/usr/bin/env python3
"""Analyze slow requests from accuracy test results"""

import json
import sys

def analyze_slow_requests(json_file):
    with open(json_file, 'r') as f:
        data = json.load(f)
    
    results = [r for r in data.get('test_results', []) if r.get('processing_time') is not None]
    if not results:
        print('No results with processing_time found')
        print('Sample result keys:', list(data.get('test_results', [{}])[0].keys()) if data.get('test_results') else 'No results')
        return
    results.sort(key=lambda x: x.get('processing_time', 0), reverse=True)
    
    print('Top 15 Slowest Requests:')
    print('=' * 100)
    print(f'{"#":<4} {"Time (s)":<10} {"Business Name":<40} {"Website":<30} {"Industry":<20}')
    print('-' * 100)
    
    for i, r in enumerate(results[:15], 1):
        time_s = r.get('processing_time', 0) / 1_000_000_000
        name = r.get('business_name', 'N/A')[:38]
        url = (r.get('website_url') or 'N/A')[:28]
        industry = (r.get('actual_industry') or 'N/A')[:18]
        print(f'{i:<4} {time_s:<10.2f} {name:<40} {url:<30} {industry:<20}')
    
    print('\n' + '=' * 100)
    print(f'Total requests analyzed: {len(results)}')
    print(f'Average processing time: {sum(r.get("processing_time", 0) for r in results) / len(results) / 1_000_000_000:.2f}s')
    print(f'Median processing time: {sorted([r.get("processing_time", 0) for r in results])[len(results)//2] / 1_000_000_000:.2f}s')
    
    # Analyze by website URL patterns
    print('\nSlow Requests by Website Pattern:')
    print('-' * 100)
    url_patterns = {}
    for r in results[:20]:  # Top 20 slowest
        url = r.get('website_url', 'N/A')
        if url != 'N/A':
            domain = url.split('/')[2] if '/' in url else url
            if domain not in url_patterns:
                url_patterns[domain] = []
            url_patterns[domain].append(r.get('processing_time', 0) / 1_000_000_000)
    
    for domain, times in sorted(url_patterns.items(), key=lambda x: sum(x[1])/len(x[1]), reverse=True)[:10]:
        avg_time = sum(times) / len(times)
        print(f'{domain[:50]:<50} Avg: {avg_time:.2f}s (count: {len(times)})')

if __name__ == '__main__':
    json_file = sys.argv[1] if len(sys.argv) > 1 else 'accuracy_report_railway_production_20251201_132726.json'
    analyze_slow_requests(json_file)

