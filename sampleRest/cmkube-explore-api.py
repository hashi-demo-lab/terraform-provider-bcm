#!/usr/bin/env python3
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

"""
API Exploration Script: CMKube API

Purpose: Explore BCM cmkube service API endpoints to understand KubeCluster entity structure
         and available operations.

Usage:
    BCM_ENDPOINT="https://172.21.15.254:8081" \
    BCM_USERNAME="root" \
    BCM_PASSWORD="Hashicorp123!" \
    python3 sampleRest/cmkube-explore-api.py
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

        print("✓ Authentication successful\n")

    except requests.exceptions.RequestException as e:
        print(f"ERROR: Failed to authenticate: {e}", file=sys.stderr)
        sys.exit(1)

    # Step 2: Explore cmkube API methods
    methods_to_test = [
        "getKubeClusters",
        "getKubeCluster",
        "listKubeClusters",
        "getKubeClusterNames",
        "validateKubeCluster",
    ]

    print("=" * 80)
    print("EXPLORING CMKUBE API METHODS")
    print("=" * 80)

    api_url = f"{endpoint}/json"
    successful_methods = []

    for method in methods_to_test:
        print(f"\n[Testing] {method}...")
        api_payload = {
            "service": "cmkube",
            "call": method
        }

        try:
            api_response = session.post(
                api_url,
                json=api_payload,
                verify=False,
                timeout=10
            )

            if api_response.status_code == 200:
                try:
                    response_data = api_response.json()

                    # Check for API-level errors
                    if isinstance(response_data, dict) and response_data.get("error"):
                        print(f"  ✗ Error: {response_data.get('error')}")
                    else:
                        print(f"  ✓ Success!")
                        successful_methods.append(method)

                        # Show response preview
                        if isinstance(response_data, list):
                            print(f"    Response type: Array with {len(response_data)} items")
                            if response_data:
                                print(f"    First item keys: {list(response_data[0].keys())[:10]}")
                        elif isinstance(response_data, dict):
                            print(f"    Response type: Object")
                            print(f"    Keys: {list(response_data.keys())[:10]}")
                        else:
                            print(f"    Response type: {type(response_data).__name__}")
                            print(f"    Value: {str(response_data)[:100]}")
                except json.JSONDecodeError:
                    print(f"  ✗ Invalid JSON response")
            else:
                print(f"  ✗ HTTP {api_response.status_code}")

        except requests.exceptions.RequestException as e:
            print(f"  ✗ Request failed: {e}")

    # Step 3: Get detailed response from successful method
    print("\n" + "=" * 80)
    print("DETAILED API RESPONSE")
    print("=" * 80)

    if successful_methods:
        # Use the first successful list method
        method = successful_methods[0]
        print(f"\nUsing method: {method}\n")

        api_payload = {
            "service": "cmkube",
            "call": method
        }

        try:
            api_response = session.post(
                api_url,
                json=api_payload,
                verify=False,
                timeout=10
            )
            response_data = api_response.json()

            print(json.dumps(response_data, indent=2))

            # Save to file
            output_file = f"sampleRest/cmkube_response_{method}.json"
            with open(output_file, 'w') as f:
                json.dump(response_data, f, indent=2)
            print(f"\n✓ Response saved to: {output_file}")

        except Exception as e:
            print(f"ERROR: {e}", file=sys.stderr)
    else:
        print("\n⚠ No successful methods found. The cmkube service may not be available")
        print("  or may require different method names.")
        print("\n  This could mean:")
        print("  - Kubernetes integration is not configured in this BCM cluster")
        print("  - The service name is different (try 'CMKube' vs 'cmkube')")
        print("  - Different method naming convention is used")

    print("\n" + "=" * 80)
    print("EXPLORATION COMPLETE")
    print("=" * 80)
    print(f"\nSuccessful methods: {len(successful_methods)}")
    if successful_methods:
        for m in successful_methods:
            print(f"  - {m}")

if __name__ == "__main__":
    main()
