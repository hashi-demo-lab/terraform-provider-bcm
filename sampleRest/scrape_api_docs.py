#!/usr/bin/env python3
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

"""
Script to scrape API documentation from BCM HTML pages
Extracts service definitions, methods, parameters, and examples
"""

import requests
import json
import sys
from datetime import datetime
import urllib3
from bs4 import BeautifulSoup
import re

# Disable SSL warnings for self-signed certificates
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)


class BCMDocsScraper:
    """Scraper for BCM API Documentation HTML pages"""

    def __init__(self, base_url="https://172.21.15.254:8081"):
        self.base_url = base_url
        self.session = requests.Session()
        self.session.verify = False
        self.session.headers.update({
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
        })

    def fetch_page(self, url):
        """Fetch HTML page content"""
        print(f"Fetching: {url}")

        try:
            response = self.session.get(url, timeout=30)
            response.raise_for_status()

            print(f"Status: {response.status_code}")
            print(f"Content-Type: {response.headers.get('Content-Type', 'unknown')}")
            print(f"Content-Length: {len(response.content)} bytes")

            return response.text

        except Exception as e:
            print(f"Error fetching page: {e}")
            return None

    def parse_api_docs(self, html_content):
        """Parse API documentation from HTML"""
        soup = BeautifulSoup(html_content, 'html.parser')

        docs_data = {
            'title': None,
            'service': None,
            'description': None,
            'methods': [],
            'examples': [],
            'raw_text': None
        }

        # Try to extract title
        title = soup.find('title')
        if title:
            docs_data['title'] = title.get_text(strip=True)
            print(f"Page title: {docs_data['title']}")

        # Extract all text content
        # Remove script and style elements
        for script in soup(["script", "style"]):
            script.decompose()

        # Get text
        text = soup.get_text()

        # Clean up text
        lines = (line.strip() for line in text.splitlines())
        chunks = (phrase.strip() for line in lines for phrase in line.split("  "))
        text = '\n'.join(chunk for chunk in chunks if chunk)

        docs_data['raw_text'] = text

        # Look for specific patterns in the HTML
        # Check for React app divs
        root_div = soup.find('div', id='root')
        if root_div:
            print("Detected React app structure")
            docs_data['app_type'] = 'react'

        # Try to find any JSON data embedded in script tags
        scripts = soup.find_all('script')
        for script in scripts:
            if script.string:
                # Look for JSON-like structures
                json_matches = re.findall(r'\{[^{}]*(?:\{[^{}]*\}[^{}]*)*\}', script.string)
                if json_matches:
                    docs_data['embedded_json'] = json_matches[:5]  # First 5 matches

        # Look for specific API documentation patterns
        # Search for method definitions
        method_patterns = [
            r'(\w+)\s*\([^)]*\)',  # Function calls
            r'service["\s:]+(\w+)',  # Service names
            r'call["\s:]+(\w+)',  # API calls
        ]

        for pattern in method_patterns:
            matches = re.findall(pattern, text, re.IGNORECASE)
            if matches:
                docs_data['methods'].extend(matches)

        return docs_data

    def save_html(self, html_content, filename):
        """Save raw HTML to file"""
        with open(filename, 'w', encoding='utf-8') as f:
            f.write(html_content)
        print(f"HTML saved to: {filename}")

    def save_parsed_data(self, data, filename):
        """Save parsed data to JSON"""
        with open(filename, 'w', encoding='utf-8') as f:
            json.dump(data, f, indent=2, ensure_ascii=False)
        print(f"Parsed data saved to: {filename}")

    def extract_react_app_data(self, html_content):
        """
        For React apps, try to extract the initial state or data
        by examining the HTML and any inline scripts
        """
        soup = BeautifulSoup(html_content, 'html.parser')

        data = {
            'meta_tags': [],
            'links': [],
            'scripts': [],
            'manifest': None
        }

        # Extract meta tags
        for meta in soup.find_all('meta'):
            data['meta_tags'].append({
                'name': meta.get('name'),
                'content': meta.get('content'),
                'property': meta.get('property')
            })

        # Extract script sources
        for script in soup.find_all('script', src=True):
            data['scripts'].append(script['src'])

        # Extract links (CSS, manifest, etc.)
        for link in soup.find_all('link'):
            data['links'].append({
                'rel': link.get('rel'),
                'href': link.get('href'),
                'type': link.get('type')
            })

        # Get manifest if exists
        manifest_link = soup.find('link', rel='manifest')
        if manifest_link:
            data['manifest'] = manifest_link.get('href')

        return data

    def fetch_json_endpoint(self, path):
        """Try to fetch JSON data from a path"""
        url = f"{self.base_url}{path}"
        print(f"\nAttempting to fetch JSON from: {url}")

        try:
            response = self.session.get(url, timeout=10)
            response.raise_for_status()

            if 'application/json' in response.headers.get('Content-Type', ''):
                return response.json()
            else:
                print(f"Not JSON: {response.headers.get('Content-Type')}")
                return None

        except Exception as e:
            print(f"Failed to fetch: {e}")
            return None


def main():
    """Main execution function"""

    # URL to scrape
    target_url = "https://172.21.15.254:8081/api/search?type=service&name=CMDevice"

    print("="*80)
    print("BCM API Documentation Scraper")
    print("="*80)
    print(f"Target URL: {target_url}\n")

    # Initialize scraper
    scraper = BCMDocsScraper()

    # Generate timestamp for filenames
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")

    # Fetch the HTML page
    html_content = scraper.fetch_page(target_url)

    if not html_content:
        print("Failed to fetch page content")
        sys.exit(1)

    # Save raw HTML
    print("\n" + "="*80)
    print("Saving raw HTML...")
    print("="*80)
    html_filename = f"cmdevice_docs_raw_{timestamp}.html"
    scraper.save_html(html_content, html_filename)

    # Parse the HTML
    print("\n" + "="*80)
    print("Parsing HTML content...")
    print("="*80)
    parsed_data = scraper.parse_api_docs(html_content)

    # Save parsed data
    parsed_filename = f"cmdevice_docs_parsed_{timestamp}.json"
    scraper.save_parsed_data(parsed_data, parsed_filename)

    # Extract React app data
    print("\n" + "="*80)
    print("Extracting React app metadata...")
    print("="*80)
    react_data = scraper.extract_react_app_data(html_content)

    react_filename = f"cmdevice_react_metadata_{timestamp}.json"
    scraper.save_parsed_data(react_data, react_filename)

    # Try to fetch any JSON endpoints discovered
    print("\n" + "="*80)
    print("Attempting to discover JSON endpoints...")
    print("="*80)

    potential_endpoints = [
        "/api/docs",
        "/api/swagger.json",
        "/api/openapi.json",
        "/api/spec",
        "/api/services",
        "/api/methods",
        "/json",
    ]

    json_data = {}
    for endpoint in potential_endpoints:
        data = scraper.fetch_json_endpoint(endpoint)
        if data:
            json_data[endpoint] = data

    if json_data:
        json_filename = f"cmdevice_discovered_json_{timestamp}.json"
        scraper.save_parsed_data(json_data, json_filename)

    # Print summary
    print("\n" + "="*80)
    print("Summary")
    print("="*80)
    print(f"Files created:")
    print(f"  1. {html_filename} - Raw HTML")
    print(f"  2. {parsed_filename} - Parsed documentation data")
    print(f"  3. {react_filename} - React app metadata")
    if json_data:
        print(f"  4. {json_filename} - Discovered JSON endpoints")

    print(f"\nParsed data summary:")
    print(f"  - Title: {parsed_data.get('title', 'N/A')}")
    print(f"  - Methods found: {len(set(parsed_data.get('methods', [])))}")
    print(f"  - React scripts: {len(react_data.get('scripts', []))}")

    # Print sample of raw text
    print(f"\nFirst 500 characters of page text:")
    print("-" * 80)
    print(parsed_data.get('raw_text', '')[:500])
    print("-" * 80)

    print("\n✓ Scraping completed successfully")


if __name__ == "__main__":
    main()
