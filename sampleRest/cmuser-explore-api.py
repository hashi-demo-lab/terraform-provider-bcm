#!/usr/bin/env python3
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

"""
BCM CMUser API Exploration - Discover User Management API
Purpose: Explore cmuser service to document available methods and user structure
"""

import json
import os
import requests
import sys
from urllib3.exceptions import InsecureRequestWarning

# Suppress SSL warnings for self-signed certs
requests.packages.urllib3.disable_warnings(category=InsecureRequestWarning)

# Configuration from environment variables
BCM_ENDPOINT = os.getenv("BCM_ENDPOINT", "https://172.21.15.254:8081")
BCM_USERNAME = os.getenv("BCM_USERNAME", "root")
BCM_PASSWORD = os.getenv("BCM_PASSWORD", "Hashicorp123!")

def authenticate(session):
    """Authenticate with BCM and get session cookie"""
    login_payload = {
        "service": "login",
        "username": BCM_USERNAME,
        "password": BCM_PASSWORD
    }

    print(f"Authenticating to {BCM_ENDPOINT}...")
    response = session.post(
        f"{BCM_ENDPOINT}/json",
        json=login_payload,
        verify=False
    )
    response.raise_for_status()

    data = response.json()
    # BCM login returns True on success, or error object on failure
    if data is True:
        print("Authentication successful")
        return session
    elif isinstance(data, dict) and data.get("success", False):
        print("Authentication successful")
        return session
    else:
        raise Exception(f"Authentication failed: {data}")


def try_api_call(session, service, method, args=None):
    """Try a BCM API call and return result"""
    payload = {
        "service": service,
        "call": method
    }
    if args is not None:
        payload["args"] = args if isinstance(args, list) else [args]

    print(f"\nTrying {service}.{method}{'(' + str(args) + ')' if args else ''}...")
    try:
        response = session.post(
            f"{BCM_ENDPOINT}/json",
            json=payload,
            verify=False,
            timeout=10
        )
        response.raise_for_status()
        data = response.json()
        return {"success": True, "data": data}
    except Exception as e:
        return {"success": False, "error": str(e)}

def main():
    """Main execution"""
    print("=" * 80)
    print("BCM CMUser API Exploration")
    print("=" * 80)

    session = requests.Session()

    try:
        # Authenticate
        authenticate(session)

        # Try various CMUser API methods
        methods_to_try = [
            ("cmuser", "getUsers"),
            ("cmuser", "listUsers"),
            ("cmuser", "getUserList"),
            ("cmuser", "getAllUsers"),
            ("cmuser", "getUser", ["root"]),  # Try with username arg
            ("cmuser", "getUserByName", ["root"]),
        ]

        results = {}
        for call_info in methods_to_try:
            service = call_info[0]
            method = call_info[1]
            args = call_info[2] if len(call_info) > 2 else None

            key = f"{service}.{method}"
            if args:
                key += f"({args})"

            result = try_api_call(session, service, method, args)
            results[key] = result

        # Display summary
        print("\n" + "=" * 80)
        print("API Call Summary")
        print("=" * 80)

        successful_calls = []
        failed_calls = []

        for key, result in results.items():
            if result["success"]:
                successful_calls.append(key)
                print(f"✓ {key}")

                # If this is a successful data call, show sample
                data = result["data"]
                if isinstance(data, list) and len(data) > 0:
                    print(f"  Found {len(data)} item(s)")
                    print(f"  Sample fields: {', '.join(sorted(data[0].keys()) if isinstance(data[0], dict) else [])}")
            else:
                failed_calls.append((key, result["error"]))
                print(f"✗ {key}: {result['error']}")

        # Show detailed output for successful calls
        if successful_calls:
            print("\n" + "=" * 80)
            print("Detailed Results for Successful Calls")
            print("=" * 80)

            for key in successful_calls:
                result = results[key]
                print(f"\n{key}:")
                print(json.dumps(result["data"], indent=2))

        # Save all results to file
        output_file = "/workspace/sampleRest/cmuser-explore-output.json"
        with open(output_file, 'w') as f:
            json.dump(results, f, indent=2)
        print(f"\n\nFull results saved to: {output_file}")

        return 0

    except Exception as e:
        print(f"\nError: {e}", file=sys.stderr)
        import traceback
        traceback.print_exc()
        return 1

if __name__ == "__main__":
    sys.exit(main())
