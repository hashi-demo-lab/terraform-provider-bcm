#!/usr/bin/env python3
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

"""
Detailed test of validateKubeCluster with minimal entity
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

def call_api(cookies, service, call, *args):
    url = f"{BCM_ENDPOINT}/json"
    payload = {"service": service, "call": call}
    if args:
        payload["args"] = list(args)
    response = requests.post(url, json=payload, cookies=cookies, verify=False)
    return response.status_code, response.json()

def main():
    print("=" * 70)
    print("Testing validateKubeCluster with minimal entity")
    print("=" * 70)

    cookies = login()
    print("✅ Logged in\n")

    # Create a minimal test entity
    minimal_entity = {
        "baseType": "KubeCluster",
        "childType": "",
        "modified": True,
        "to_be_removed": False,
        "revision": "",
        "name": "test-validation-cluster",
        "uuid": "00000000-0000-0000-0000-000000000000"
    }

    print("🔍 Test 1: validateKubeCluster with minimal entity")
    print(f"   Entity: {json.dumps(minimal_entity, indent=2)}")

    status, result = call_api(cookies, "cmkube", "validateKubeCluster", minimal_entity)
    print(f"\n   Status: {status}")
    print(f"   Result: {json.dumps(result, indent=2)}")

    if status == 200:
        print("\n   ✅ validateKubeCluster METHOD EXISTS")

        if isinstance(result, list):
            if len(result) == 0:
                print("   Validation passed (empty array)")
            else:
                print(f"   Validation errors found: {len(result)}")
                for i, error in enumerate(result[:3]):
                    print(f"\n   Error {i+1}:")
                    print(f"     Field: {error.get('field')}")
                    print(f"     Message: {error.get('message')}")
                    print(f"     Severity: {error.get('severity')}")
    else:
        print("\n   ❌ validateKubeCluster failed or doesn't exist")

    print("\n" + "=" * 70)

if __name__ == "__main__":
    main()
