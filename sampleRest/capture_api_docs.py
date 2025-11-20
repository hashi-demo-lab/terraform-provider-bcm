#!/usr/bin/env python3
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

"""
Script to capture API documentation from BCM API endpoint
"""

import requests
import json
import sys
from datetime import datetime
import urllib3

# Disable SSL warnings for self-signed certificates
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

def capture_api_docs(url, output_file=None, headers=None):
    """
    Capture API documentation from the specified endpoint

    Args:
        url: The API endpoint URL
        output_file: Optional file path to save the response
        headers: Optional dict of HTTP headers to include

    Returns:
        dict: Parsed JSON response or None on failure
    """
    print(f"Fetching API docs from: {url}")

    try:
        # Make the request with SSL verification disabled (common for internal dev APIs)
        response = requests.get(
            url,
            headers=headers,
            verify=False,  # Disable SSL verification for self-signed certs
            timeout=30
        )

        # Check if request was successful
        response.raise_for_status()

        # Print summary first
        print(f"\nStatus Code: {response.status_code}")
        print(f"Response Size: {len(response.content)} bytes")
        print(f"Content-Type: {response.headers.get('Content-Type', 'unknown')}")

        # Check if response is JSON
        content_type = response.headers.get('Content-Type', '').lower()

        if 'application/json' in content_type or url.endswith('.json'):
            # Parse JSON response
            api_docs = response.json()

            # Pretty print to console
            print("\n" + "="*80)
            print("API Response (JSON):")
            print("="*80)
            print(json.dumps(api_docs, indent=2))
            print("="*80)

            # Save to file if specified
            if output_file:
                with open(output_file, 'w') as f:
                    json.dump(api_docs, f, indent=2)
                print(f"\nAPI documentation saved to: {output_file}")

            return api_docs

        elif 'text/html' in content_type:
            print("\n" + "="*80)
            print("Response is HTML (Web UI detected)")
            print("="*80)
            print("This appears to be a web interface, not a REST API endpoint.")
            print("\nFirst 500 characters:")
            print(response.text[:500])
            print("...")

            # Save HTML to file
            if output_file:
                html_file = output_file.replace('.json', '.html')
                with open(html_file, 'w') as f:
                    f.write(response.text)
                print(f"\nHTML content saved to: {html_file}")

            print("\nSuggestions:")
            print("1. Open the URL in a browser and use Developer Tools (F12) to inspect network requests")
            print("2. Look for the actual API endpoint being called by the web UI")
            print("3. Check if authentication/API keys are required")
            print("4. Try accessing the base API path: https://172.21.15.254:8081/api/")

            return None

        else:
            # Unknown content type, save raw response
            print(f"\nUnexpected content type: {content_type}")
            print("Raw response (first 500 chars):")
            print(response.text[:500])

            if output_file:
                raw_file = output_file.replace('.json', '.txt')
                with open(raw_file, 'w') as f:
                    f.write(response.text)
                print(f"\nRaw content saved to: {raw_file}")

            return None

    except requests.exceptions.SSLError as e:
        print(f"SSL Error: {e}", file=sys.stderr)
        print("Tip: The script already disables SSL verification. Check if the host is reachable.", file=sys.stderr)
        return None

    except requests.exceptions.ConnectionError as e:
        print(f"Connection Error: {e}", file=sys.stderr)
        print("Tip: Verify the IP address and port are correct and the server is running.", file=sys.stderr)
        return None

    except requests.exceptions.Timeout as e:
        print(f"Timeout Error: {e}", file=sys.stderr)
        return None

    except requests.exceptions.HTTPError as e:
        print(f"HTTP Error: {e}", file=sys.stderr)
        print(f"Response: {response.text}", file=sys.stderr)
        return None

    except json.JSONDecodeError as e:
        print(f"JSON Decode Error: {e}", file=sys.stderr)
        print(f"Raw Response: {response.text[:500]}", file=sys.stderr)
        return None

    except Exception as e:
        print(f"Unexpected Error: {e}", file=sys.stderr)
        return None


def main():
    """Main execution function"""
    # API endpoint
    api_url = "https://172.21.15.254:8081/api/search?type=entity&name=Device"

    # Generate timestamped output filename
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    output_file = f"device_entity_docs_{timestamp}.json"

    # Try with Accept header for JSON
    headers = {
        'Accept': 'application/json',
        'User-Agent': 'API Documentation Capture Script'
    }

    # Capture API docs
    print("Attempting to fetch API documentation...")
    print("=" * 80)
    result = capture_api_docs(api_url, output_file, headers=headers)

    if result:
        print("\n✓ API documentation captured successfully")
        sys.exit(0)
    else:
        print("\n✗ Response was not JSON (see suggestions above)")
        sys.exit(1)


if __name__ == "__main__":
    main()
