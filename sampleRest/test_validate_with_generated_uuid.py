#!/usr/bin/env python3
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

"""
Test validation with GENERATED UUID (like the provider does)
vs without UUID (like my previous test)
"""

import requests
import json
import os
import uuid
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
    print("Testing Validation: Generated UUID vs No UUID vs Zero UUID")
    print("=" * 70)

    cookies = login()
    print("✅ Logged in\n")

    base_entity = {
        "baseType": "SoftwareImage",
        "childType": "",
        "modified": True,
        "to_be_removed": False,
        "revision": "",
        "name": "test-validation-uuid-comparison",
        "path": "/cm/images/test.iso",
        "notes": "Test",
        "kernelParameters": "",
        "enableSOL": True,
        "SOLSpeed": "115200",
        "SOLPort": "ttyS1",
        "SOLFlowControl": True,
        "modules": []
    }

    # Test 1: No UUID field (my previous test)
    print("TEST 1: Validate without UUID field")
    print("-" * 70)
    entity_no_uuid = base_entity.copy()
    status, result = call_api(cookies, "CMPart", "validateSoftwareImage", entity_no_uuid)
    print(f"Status: {status}")

    if isinstance(result, list):
        print(f"Validation results: {len(result)} messages")
        for r in result:
            if r.get('field') == 'uuid':
                print(f"  - UUID error: {r.get('message')}")
    else:
        print(f"Result: {json.dumps(result, indent=2)[:200]}")

    # Test 2: With GENERATED UUID (what the provider actually does)
    print("\n" + "=" * 70)
    print("TEST 2: Validate with GENERATED UUID (provider pattern)")
    print("-" * 70)
    entity_with_uuid = base_entity.copy()
    entity_with_uuid["uuid"] = str(uuid.uuid4())
    print(f"Generated UUID: {entity_with_uuid['uuid']}")

    status, result = call_api(cookies, "CMPart", "validateSoftwareImage", entity_with_uuid)
    print(f"Status: {status}")

    if isinstance(result, list):
        print(f"Validation results: {len(result)} messages")
        if len(result) == 0:
            print("  ✅ NO VALIDATION ERRORS!")
        else:
            for r in result:
                print(f"  - {r.get('severity')}: {r.get('field')} - {r.get('message')}")
    else:
        print(f"Result: {json.dumps(result, indent=2)[:200]}")

    # Test 3: With ZERO UUID (00000000-0000-0000-0000-000000000000)
    print("\n" + "=" * 70)
    print("TEST 3: Validate with ZERO UUID")
    print("-" * 70)
    entity_zero_uuid = base_entity.copy()
    entity_zero_uuid["uuid"] = "00000000-0000-0000-0000-000000000000"

    status, result = call_api(cookies, "CMPart", "validateSoftwareImage", entity_zero_uuid)
    print(f"Status: {status}")

    if isinstance(result, list):
        print(f"Validation results: {len(result)} messages")
        for r in result:
            if r.get('field') == 'uuid':
                print(f"  - UUID error: {r.get('message')}")
            else:
                print(f"  - {r.get('severity')}: {r.get('field')} - {r.get('message')}")
    else:
        print(f"Result: {json.dumps(result, indent=2)[:200]}")

    # Test 4: Invalid field with GENERATED UUID
    print("\n" + "=" * 70)
    print("TEST 4: Invalid SOL speed with GENERATED UUID")
    print("-" * 70)
    entity_invalid = base_entity.copy()
    entity_invalid["uuid"] = str(uuid.uuid4())
    entity_invalid["SOLSpeed"] = "999999"  # Invalid
    print(f"Generated UUID: {entity_invalid['uuid']}")

    status, result = call_api(cookies, "CMPart", "validateSoftwareImage", entity_invalid)
    print(f"Status: {status}")

    if isinstance(result, list):
        print(f"Validation results: {len(result)} messages")
        if len(result) == 0:
            print("  ⚠️  NO VALIDATION ERRORS (unexpected!)")
        else:
            for r in result:
                severity = "✅" if r.get('severity') == "ERROR" else "⚠️"
                print(f"  {severity} {r.get('severity')}: {r.get('field')} - {r.get('message')}")
    else:
        print(f"Result: {json.dumps(result, indent=2)[:200]}")

    # Summary
    print("\n" + "=" * 70)
    print("SUMMARY")
    print("=" * 70)
    print("\n🔍 Key Finding:")
    print("   The provider GENERATES a UUID for CREATE operations!")
    print("   It does NOT send zero UUID or omit the UUID field.")
    print("\n💡 Implication:")
    print("   If provider sends a GENERATED UUID, validation should work")
    print("   without 'Zero UUID' errors!")
    print("\n📋 Recommendation:")
    print("   Re-test validation using the ACTUAL provider flow:")
    print("   1. Generate UUID with uuid.New().String()")
    print("   2. Include UUID in entity")
    print("   3. Call validateSoftwareImage")
    print("   4. Check if 'Zero UUID' error still appears")

if __name__ == "__main__":
    main()
