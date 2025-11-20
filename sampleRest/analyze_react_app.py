#!/usr/bin/env python3
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

"""
Analyze React app JavaScript bundle to discover API endpoints and structure
"""

import requests
import json
import re
import sys
from datetime import datetime
import urllib3

urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)


class ReactAppAnalyzer:
    """Analyzer for React app JavaScript bundles"""

    def __init__(self, base_url="https://172.21.15.254:8081"):
        self.base_url = base_url
        self.session = requests.Session()
        self.session.verify = False

    def fetch_resource(self, path):
        """Fetch a resource from the server"""
        url = f"{self.base_url}{path}"
        print(f"Fetching: {url}")

        try:
            response = self.session.get(url, timeout=30)
            response.raise_for_status()
            print(f"✓ Success ({len(response.content)} bytes)")
            return response.text
        except Exception as e:
            print(f"✗ Failed: {e}")
            return None

    def analyze_js_bundle(self, js_content):
        """Analyze JavaScript bundle for API endpoints and patterns"""
        print("\nAnalyzing JavaScript bundle...")

        analysis = {
            'size': len(js_content),
            'api_endpoints': [],
            'services': [],
            'methods': [],
            'url_patterns': [],
            'json_objects': [],
            'function_names': []
        }

        # Extract API endpoint patterns
        # Look for URL patterns
        url_patterns = [
            r'["\']/(api|json)/[^"\']*["\']',
            r'https?://[^"\']+',
            r'["\']/(cm|bcm)/[^"\']*["\']',
        ]

        for pattern in url_patterns:
            matches = re.findall(pattern, js_content)
            analysis['url_patterns'].extend(matches)

        # Look for service names
        service_pattern = r'service["\s:]+["\'](\w+)["\']'
        services = re.findall(service_pattern, js_content, re.IGNORECASE)
        analysis['services'] = list(set(services))

        # Look for API call methods
        call_pattern = r'call["\s:]+["\'](\w+)["\']'
        calls = re.findall(call_pattern, js_content, re.IGNORECASE)
        analysis['methods'] = list(set(calls))

        # Look for endpoint definitions
        endpoint_patterns = [
            r'endpoint["\s:]+["\']([^"\']+)["\']',
            r'path["\s:]+["\']([^"\']+)["\']',
            r'route["\s:]+["\']([^"\']+)["\']',
        ]

        for pattern in endpoint_patterns:
            endpoints = re.findall(pattern, js_content, re.IGNORECASE)
            analysis['api_endpoints'].extend(endpoints)

        # Look for JSON-like objects (limited to prevent huge output)
        # Simple JSON object pattern
        json_pattern = r'\{[^{}]{20,200}\}'
        json_matches = re.findall(json_pattern, js_content)
        analysis['json_objects'] = json_matches[:20]  # First 20

        # Look for function names that might be API-related
        function_patterns = [
            r'function\s+(\w+)',
            r'const\s+(\w+)\s*=\s*(?:async\s+)?(?:function|\()',
            r'(\w+)\s*:\s*(?:async\s+)?function',
        ]

        for pattern in function_patterns:
            funcs = re.findall(pattern, js_content)
            analysis['function_names'].extend(funcs)

        # Filter for API-related function names
        api_related = [f for f in analysis['function_names']
                      if any(keyword in f.lower() for keyword in
                             ['api', 'fetch', 'get', 'post', 'call', 'request', 'service'])]
        analysis['api_related_functions'] = list(set(api_related))[:50]

        # Deduplicate
        analysis['url_patterns'] = list(set(analysis['url_patterns']))
        analysis['services'] = list(set(analysis['services']))
        analysis['methods'] = list(set(analysis['methods']))
        analysis['api_endpoints'] = list(set(analysis['api_endpoints']))

        return analysis

    def fetch_manifest(self):
        """Fetch and parse manifest.json"""
        return self.fetch_resource('/api/manifest.json')


def save_analysis(data, filename):
    """Save analysis to JSON file"""
    with open(filename, 'w', encoding='utf-8') as f:
        json.dump(data, f, indent=2, ensure_ascii=False)
    print(f"Analysis saved to: {filename}")


def save_content(content, filename):
    """Save content to file"""
    with open(filename, 'w', encoding='utf-8') as f:
        f.write(content)
    print(f"Content saved to: {filename}")


def main():
    """Main execution"""

    print("="*80)
    print("React App Analyzer - BCM API Documentation")
    print("="*80)

    analyzer = ReactAppAnalyzer()
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")

    # Fetch and analyze JavaScript bundle
    print("\n1. Fetching JavaScript bundle...")
    js_bundle = analyzer.fetch_resource('/api/static/js/main.9875e9e5.js')

    if js_bundle:
        # Save raw bundle
        bundle_filename = f"react_bundle_{timestamp}.js"
        save_content(js_bundle, bundle_filename)

        # Analyze bundle
        analysis = analyzer.analyze_js_bundle(js_bundle)

        # Save analysis
        analysis_filename = f"react_bundle_analysis_{timestamp}.json"
        save_analysis(analysis, analysis_filename)

        # Print summary
        print("\n" + "="*80)
        print("JavaScript Bundle Analysis")
        print("="*80)
        print(f"Bundle size: {analysis['size']:,} bytes")
        print(f"\nDiscovered patterns:")
        print(f"  - URL patterns: {len(analysis['url_patterns'])}")
        print(f"  - Services: {len(analysis['services'])}")
        print(f"  - Methods/Calls: {len(analysis['methods'])}")
        print(f"  - API endpoints: {len(analysis['api_endpoints'])}")
        print(f"  - API-related functions: {len(analysis.get('api_related_functions', []))}")

        if analysis['services']:
            print(f"\nServices found:")
            for service in sorted(analysis['services'])[:20]:
                print(f"  - {service}")

        if analysis['methods']:
            print(f"\nAPI Methods/Calls found:")
            for method in sorted(analysis['methods'])[:20]:
                print(f"  - {method}")

        if analysis['url_patterns']:
            print(f"\nURL Patterns found:")
            for url in sorted(analysis['url_patterns'])[:20]:
                print(f"  - {url}")

        if analysis.get('api_related_functions'):
            print(f"\nAPI-related functions:")
            for func in sorted(analysis['api_related_functions'])[:20]:
                print(f"  - {func}")

    # Fetch manifest
    print("\n" + "="*80)
    print("2. Fetching manifest.json...")
    print("="*80)
    manifest = analyzer.fetch_manifest()

    if manifest:
        manifest_filename = f"manifest_{timestamp}.json"
        save_content(manifest, manifest_filename)

        try:
            manifest_data = json.loads(manifest)
            print("\nManifest data:")
            print(json.dumps(manifest_data, indent=2))
        except:
            print("Manifest is not valid JSON")

    # Try to fetch CSS
    print("\n" + "="*80)
    print("3. Fetching CSS...")
    print("="*80)
    css = analyzer.fetch_resource('/api/static/css/main.9e6e7576.css')

    if css:
        css_filename = f"styles_{timestamp}.css"
        save_content(css, css_filename)

    print("\n" + "="*80)
    print("✓ Analysis completed successfully")
    print("="*80)


if __name__ == "__main__":
    main()
