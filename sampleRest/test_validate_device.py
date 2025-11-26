#!/usr/bin/env python3
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

"""
Test if BCM API supports device validation methods
"""

import requests
import json
import os
from urllib3.exceptions import InsecureRequestWarning

requests.packages.urllib3.disable_warnings(category=InsecureRequestWarning)

BCM_ENDPOINT = os.getenv("BCM_ENDPOINT", "https://172.21.15.254:8081")
BCM_USERNAME = os.getenv("BCM_USERNAME", "root")
BCM_PASSWORD = os.getenv("BCM_PASSWORD", "Hashicorp123!")

def login():
    url = f"{BCM_ENDPOINT}/json"
    payload = {"service": "login", "username": BCM_USERNAME, "password": BCM_PASSWORD}
    response = requests.post(url, json=payload, verify=False)
    if response.status_code != 200 or response.json() != True:
        print(f"❌ Login failed")
        exit(1)
    print("✅ Login successful")
    return response.cookies

def call_api(cookies, service, call, *args):
    url = f"{BCM_ENDPOINT}/json"
    payload = {"service": service, "call": call}
    if args:
        payload["args"] = list(args)
    response = requests.post(url, json=payload, cookies=cookies, verify=False)
    return response.status_code, response.json()

def main():
    print("=" * 70)
    print("BCM API Device Validation Method Discovery")
    print("=" * 70)

    cookies = login()

    # Get a sample device/node
    print("\n📦 Fetching sample device/node...")
    status, nodes = call_api(cookies, "CMDevice", "getNodes")

    if not nodes or len(nodes) == 0:
        print("❌ No devices/nodes found in system")
        exit(1)

    sample_device = nodes[0]
    print(f"   Found device: {sample_device.get('name', 'unknown')}")
    print(f"   UUID: {sample_device.get('uuid', 'none')}")

    # Test validation methods
    validation_methods = [
        "validateDevice",
        "validateNode",
        "validateNodeDevice",
        "validateDeviceNode",
    ]

    print("\n" + "=" * 70)
    print("TESTING DEVICE VALIDATION METHODS")
    print("=" * 70)

    found_methods = []

    for method in validation_methods:
        print(f"\n🔍 Testing CMDevice.{method}")
        status, result = call_api(cookies, "CMDevice", method, sample_device)

        print(f"   Status: {status}")

        if status == 200:
            print(f"   ✅ METHOD EXISTS")
            found_methods.append(method)

            if isinstance(result, list):
                if len(result) == 0:
                    print(f"   Result: [] (validation passed)")
                else:
                    print(f"   Result: {len(result)} validation messages")
                    print(f"   First message: {json.dumps(result[0], indent=2)[:300]}")
            elif isinstance(result, dict):
                print(f"   Result: {json.dumps(result, indent=2)[:200]}")
            else:
                print(f"   Result type: {type(result)}")

        elif status == 400:
            result_json = result if isinstance(result, dict) else {}
            error_msg = result_json.get("errormessage", str(result))

            if "does not exist" in error_msg.lower():
                print(f"   ❌ Method does not exist")
            else:
                print(f"   ⚠️  Status 400: {error_msg[:100]}")
        else:
            print(f"   ❌ Unexpected status: {result}")

    # Test with invalid modifications
    if found_methods:
        method = found_methods[0]
        print("\n" + "=" * 70)
        print(f"TEST: {method} with invalid modifications")
        print("=" * 70)

        modified_device = sample_device.copy()
        # Try to set an invalid field value
        modified_device["modified"] = True
        modified_device["hostname"] = ""  # Empty hostname (likely invalid)

        status, result = call_api(cookies, "CMDevice", method, modified_device)
        print(f"Status: {status}")
        print(f"Result: {json.dumps(result, indent=2)[:500]}")

    # Summary
    print("\n" + "=" * 70)
    print("SUMMARY")
    print("=" * 70)

    if found_methods:
        print(f"\n✅ Found {len(found_methods)} validation method(s):")
        for method in found_methods:
            print(f"   - CMDevice.{method}")

        print("\n💡 Recommendation:")
        print(f"   Use CMDevice.{found_methods[0]} for pre-flight validation")
        print("   in resource_cmdevice_device.go Update() operation")
    else:
        print("\n❌ No device validation methods found")
        print("   Continue relying on server-side validation during update")

if __name__ == "__main__":
    main()
