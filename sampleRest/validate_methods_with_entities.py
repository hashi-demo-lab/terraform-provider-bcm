#!/usr/bin/env python3
"""
Test validate methods with actual entities
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
    cookies = login()
    print("✅ Logged in\n")

    # Get sample entities
    print("📦 Fetching sample entities...\n")

    _, images = call_api(cookies, "CMPart", "getSoftwareImages")
    _, categories = call_api(cookies, "CMDevice", "getCategories")
    _, networks = call_api(cookies, "CMNet", "getNetworks")

    sample_image = images[0] if images else None
    sample_category = categories[0] if categories else None
    sample_network = networks[0] if networks else None

    print("=" * 70)
    print("TESTING VALIDATE METHODS WITH ENTITIES")
    print("=" * 70)

    # Test 1: validateSoftwareImage
    if sample_image:
        print("\n1️⃣  CMPart.validateSoftwareImage")
        print(f"   Image: {sample_image.get('name')}")
        status, result = call_api(cookies, "CMPart", "validateSoftwareImage", sample_image)
        print(f"   Status: {status}")
        if status == 200:
            print(f"   ✅ METHOD EXISTS")
            print(f"   Result: {json.dumps(result, indent=2)[:200]}")
        else:
            print(f"   ❌ Method not found or error")
            print(f"   Response: {result}")

    # Test 2: validateCategory
    if sample_category:
        print("\n2️⃣  CMDevice.validateCategory")
        print(f"   Category: {sample_category.get('name')}")
        status, result = call_api(cookies, "CMDevice", "validateCategory", sample_category)
        print(f"   Status: {status}")
        if status == 200:
            print(f"   ✅ METHOD EXISTS")
            print(f"   Result: {json.dumps(result, indent=2)[:200]}")
        else:
            print(f"   ❌ Method not found or error")
            print(f"   Response: {result}")

    # Test 3: validateNetwork
    if sample_network:
        print("\n3️⃣  CMNet.validateNetwork")
        print(f"   Network: {sample_network.get('name')}")
        status, result = call_api(cookies, "CMNet", "validateNetwork", sample_network)
        print(f"   Status: {status}")
        if status == 200:
            print(f"   ✅ METHOD EXISTS")
            print(f"   Result: {json.dumps(result, indent=2)[:200]}")
        else:
            print(f"   ❌ Method not found or error")
            print(f"   Response: {result}")

    print("\n" + "=" * 70)

if __name__ == "__main__":
    main()
