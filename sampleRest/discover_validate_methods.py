#!/usr/bin/env python3
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

"""
Discover all validate* methods available in BCM API
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
    return response.cookies

def test_validate_method(cookies, service, method, test_entity=None):
    """Test if a validate method exists"""
    url = f"{BCM_ENDPOINT}/json"

    payload = {"service": service, "call": method}
    if test_entity:
        payload["args"] = [test_entity]

    try:
        response = requests.post(url, json=payload, cookies=cookies, verify=False, timeout=5)

        # Method doesn't exist usually returns 400 or error message
        if response.status_code == 400:
            result = response.json()
            if "errormessage" in result and "does not exist" in result.get("errormessage", "").lower():
                return False, "Method does not exist"

        # Method exists if we get 200
        if response.status_code == 200:
            return True, response.json()

        return False, f"Status {response.status_code}"
    except Exception as e:
        return False, str(e)

def main():
    print("=" * 70)
    print("BCM API Validate Method Discovery")
    print("=" * 70)

    cookies = login()
    print("✅ Logged in\n")

    # Test services and their potential validate methods
    test_cases = [
        # CMPart (Software Images, Partitions)
        ("CMPart", "validateSoftwareImage"),
        ("CMPart", "validatePartition"),

        # CMDevice (Nodes, Categories)
        ("CMDevice", "validateNode"),
        ("CMDevice", "validateCategory"),
        ("CMDevice", "validateDevice"),

        # CMNet (Networks)
        ("CMNet", "validateNetwork"),

        # CMKube (Kubernetes clusters)
        ("CMKube", "validateCluster"),

        # CMUser (Users, roles)
        ("CMUser", "validateUser"),
        ("CMUser", "validateRole"),
    ]

    print("Testing for validate methods...\n")
    found_methods = []
    not_found = []

    for service, method in test_cases:
        exists, result = test_validate_method(cookies, service, method)

        if exists:
            print(f"✅ {service}.{method} EXISTS")
            found_methods.append((service, method))
            if isinstance(result, list) and len(result) == 0:
                print(f"   Returns: [] (empty list - needs entity)")
            elif isinstance(result, dict):
                print(f"   Returns: {json.dumps(result, indent=2)[:100]}...")
        else:
            print(f"❌ {service}.{method} NOT FOUND ({result})")
            not_found.append((service, method))

    print("\n" + "=" * 70)
    print("SUMMARY")
    print("=" * 70)

    if found_methods:
        print(f"\n✅ Found {len(found_methods)} validate methods:")
        for service, method in found_methods:
            print(f"   - {service}.{method}")

    if not_found:
        print(f"\n❌ Not found ({len(not_found)} methods):")
        for service, method in not_found:
            print(f"   - {service}.{method}")

    print("\n💡 Recommendation:")
    if len(found_methods) > 1:
        print("   Multiple validate methods exist - create a generic helper function")
        print("   that can be reused across resources")
    elif len(found_methods) == 1:
        print("   Only one validate method found - keep implementation specific")
    else:
        print("   No validate methods found - rely on server-side validation")

if __name__ == "__main__":
    main()
