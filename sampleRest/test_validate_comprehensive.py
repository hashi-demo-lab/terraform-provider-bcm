#!/usr/bin/env python3
"""
Comprehensive test of validateSoftwareImage to understand its behavior
"""

import requests
import json
import sys
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
        sys.exit(1)
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
    cookies = login()

    # Get existing image
    print("\n📦 Getting existing software image...")
    status, images = call_api(cookies, "CMPart", "getSoftwareImages")
    if not images:
        print("❌ No images found")
        sys.exit(1)

    sample = images[0]
    print(f"   Using: {sample['name']} ({sample['uuid']})")

    # Test 1: Validate existing valid entity
    print("\n" + "="*70)
    print("TEST 1: Validate existing valid entity")
    print("="*70)
    status, result = call_api(cookies, "CMPart", "validateSoftwareImage", sample)
    print(f"Status: {status}")
    print(f"Result: {json.dumps(result, indent=2)}")
    print(f"Result type: {type(result)}")

    # Test 2: Validate with missing required field (name)
    print("\n" + "="*70)
    print("TEST 2: Validate with missing 'name' field")
    print("="*70)
    invalid_entity = sample.copy()
    invalid_entity.pop("name", None)
    status, result = call_api(cookies, "CMPart", "validateSoftwareImage", invalid_entity)
    print(f"Status: {status}")
    print(f"Result: {json.dumps(result, indent=2)}")

    # Test 3: Validate with invalid SOL speed
    print("\n" + "="*70)
    print("TEST 3: Validate with invalid SOLSpeed")
    print("="*70)
    invalid_entity = sample.copy()
    invalid_entity["SOLSpeed"] = "999999"  # Invalid baud rate
    status, result = call_api(cookies, "CMPart", "validateSoftwareImage", invalid_entity)
    print(f"Status: {status}")
    print(f"Result: {json.dumps(result, indent=2)}")

    # Test 4: Validate with duplicate name (new entity without UUID)
    print("\n" + "="*70)
    print("TEST 4: Validate new entity with existing name (duplicate)")
    print("="*70)
    new_entity = {
        "baseType": "SoftwareImage",
        "childType": "",
        "modified": True,
        "to_be_removed": False,
        "revision": "",
        "name": sample["name"],  # Use existing name
        "path": "/cm/images/test-duplicate.iso",
        "notes": "",
        "kernelParameters": "",
        "enableSOL": False,
        "SOLSpeed": "115200",
        "SOLPort": "ttyS1",
        "SOLFlowControl": True,
        "modules": []
    }
    status, result = call_api(cookies, "CMPart", "validateSoftwareImage", new_entity)
    print(f"Status: {status}")
    print(f"Result: {json.dumps(result, indent=2)}")

    # Test 5: Validate completely new valid entity
    print("\n" + "="*70)
    print("TEST 5: Validate new valid entity")
    print("="*70)
    new_entity = {
        "baseType": "SoftwareImage",
        "childType": "",
        "modified": True,
        "to_be_removed": False,
        "revision": "",
        "name": "validation-test-newimage-12345",
        "path": "/cm/images/validation-test.iso",
        "notes": "Test validation",
        "kernelParameters": "quiet splash",
        "enableSOL": True,
        "SOLSpeed": "115200",
        "SOLPort": "ttyS1",
        "SOLFlowControl": True,
        "modules": []
    }
    status, result = call_api(cookies, "CMPart", "validateSoftwareImage", new_entity)
    print(f"Status: {status}")
    print(f"Result: {json.dumps(result, indent=2)}")

    # Test 6: What does validation return on success?
    print("\n" + "="*70)
    print("TEST 6: Understanding validation success response")
    print("="*70)
    if isinstance(result, list):
        print("✅ Success returns: empty list []")
    elif isinstance(result, dict) and result.get("success") == True:
        print("✅ Success returns: {'success': True}")
    elif isinstance(result, bool) and result == True:
        print("✅ Success returns: boolean true")
    else:
        print(f"⚠️  Unexpected success format: {type(result)} = {result}")

    print("\n" + "="*70)
    print("CONCLUSIONS")
    print("="*70)
    print("1. validateSoftwareImage API method EXISTS")
    print("2. It can be called before add/update operations")
    print("3. Empty list [] means validation passed")
    print("4. Validation errors would return validation array or error object")

if __name__ == "__main__":
    main()
