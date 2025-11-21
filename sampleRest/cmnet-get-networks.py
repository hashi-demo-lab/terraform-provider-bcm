#!/usr/bin/env python3
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

"""
API Exploration Script: CMNet getNetworks

Purpose: Call BCM API endpoint {"service": "cmnet", "call": "getNetworks"}
         and document the response structure for data source implementation.

Usage:
    BCM_ENDPOINT="https://172.21.15.254:8081" \
    BCM_USERNAME="root" \
    BCM_PASSWORD="Hashicorp123!" \
    python3 sampleRest/cmnet-get-networks.py
"""

import os
import sys
import json
import requests
from urllib3.exceptions import InsecureRequestWarning

# Suppress SSL warnings for self-signed certificates
requests.packages.urllib3.disable_warnings(category=InsecureRequestWarning)

def main():
    # Get credentials from environment
    endpoint = os.getenv("BCM_ENDPOINT", "https://172.21.15.254:8081")
    username = os.getenv("BCM_USERNAME", "root")
    password = os.getenv("BCM_PASSWORD")

    if not password:
        print("ERROR: BCM_PASSWORD environment variable not set", file=sys.stderr)
        sys.exit(1)

    # Create session
    session = requests.Session()

    # Step 1: Login
    print(f"Authenticating to {endpoint}...")
    login_url = f"{endpoint}/json"
    login_payload = {
        "service": "login",
        "username": username,
        "password": password
    }

    try:
        login_response = session.post(
            login_url,
            json=login_payload,
            verify=False,
            timeout=10
        )
        login_response.raise_for_status()

        # Check for login errors
        login_data = login_response.json()
        if isinstance(login_data, dict) and login_data.get("error"):
            print(f"ERROR: Login failed: {login_data.get('error')}", file=sys.stderr)
            sys.exit(1)

        print("✓ Authentication successful")

    except requests.exceptions.RequestException as e:
        print(f"ERROR: Failed to authenticate: {e}", file=sys.stderr)
        sys.exit(1)

    # Step 2: Call cmnet.getNetworks
    print("\nCalling cmnet.getNetworks API...")
    api_url = f"{endpoint}/json"
    api_payload = {
        "service": "cmnet",
        "call": "getNetworks"
    }

    try:
        api_response = session.post(
            api_url,
            json=api_payload,
            verify=False,
            timeout=10
        )
        api_response.raise_for_status()

        # Parse response
        response_data = api_response.json()

        # Pretty print response
        print("\n" + "=" * 80)
        print("API RESPONSE: cmnet.getNetworks")
        print("=" * 80)
        print(json.dumps(response_data, indent=2))
        print("=" * 80)

        # Analyze response structure
        if isinstance(response_data, list):
            print(f"\n✓ Response is a JSON array with {len(response_data)} network(s)")

            if response_data:
                print("\nFirst network object keys:")
                for key in sorted(response_data[0].keys()):
                    value = response_data[0][key]
                    value_type = type(value).__name__
                    print(f"  - {key}: {value_type}")
            else:
                print("\n⚠ WARNING: No networks found in BCM cluster")
                print("Acceptance tests require at least one network to exist")
        else:
            print(f"\n⚠ WARNING: Unexpected response type: {type(response_data).__name__}")

        print("\n✓ API exploration complete")

    except requests.exceptions.RequestException as e:
        print(f"\nERROR: API call failed: {e}", file=sys.stderr)
        sys.exit(1)
    except json.JSONDecodeError as e:
        print(f"\nERROR: Failed to parse JSON response: {e}", file=sys.stderr)
        print(f"Response text: {api_response.text}")
        sys.exit(1)

if __name__ == "__main__":
    main()
