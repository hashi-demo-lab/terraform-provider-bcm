#!/usr/bin/env python3
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

"""
UUID Requirement Testing: Software Images

Purpose: Test whether BCM API requires UUID for Software Image creation.
This script compares behavior with and without UUID to understand API requirements.

Usage:
    BCM_ENDPOINT="https://172.21.15.254:8081" \
    BCM_USERNAME="root" \
    BCM_PASSWORD="Hashicorp123!" \
    python3 sampleRest/test_softwareimage_uuid_requirement.py

Context:
    Discovered during issue #26 that Networks require UUID for creation.
    Direct API testing suggests Software Images also require UUID,
    but existing Terraform tests pass without sending UUID.
    This script investigates the discrepancy.
"""

import os
import sys
import requests
import json
import uuid as py_uuid
from urllib3.exceptions import InsecureRequestWarning

# Suppress SSL warnings for self-signed certificates
requests.packages.urllib3.disable_warnings(category=InsecureRequestWarning)

def main():
    endpoint = os.getenv("BCM_ENDPOINT", "https://172.21.15.254:8081")
    username = os.getenv("BCM_USERNAME", "root")
    password = os.getenv("BCM_PASSWORD")

    if not password:
        print("ERROR: BCM_PASSWORD environment variable not set", file=sys.stderr)
        sys.exit(1)

    # Create session and login
    session = requests.Session()
    login_payload = {"service": "login", "username": username, "password": password}

    try:
        login_response = session.post(f"{endpoint}/json", json=login_payload, verify=False, timeout=10)
        login_response.raise_for_status()
        print("✓ Authenticated to BCM\n")
    except requests.exceptions.RequestException as e:
        print(f"ERROR: Failed to authenticate: {e}", file=sys.stderr)
        sys.exit(1)

    # Test 1: Software Image WITHOUT UUID
    print("=" * 80)
    print("TEST 1: Software Image WITHOUT UUID")
    print("=" * 80)

    create_entity = {
        "name": "uuid-test-no-uuid",
        "path": "/cm/images/uuid-test-no-uuid",
        "baseType": "SoftwareImage",
        "childType": "",
        "modified": True,
        "to_be_removed": False,
        "revision": ""
        # NOTE: No UUID field
    }

    create_payload = {"service": "CMPart", "call": "addSoftwareImage", "args": [create_entity, False]}

    print("Request Entity:")
    print(json.dumps(create_entity, indent=2))
    print()

    try:
        response = session.post(f"{endpoint}/json", json=create_payload, verify=False, timeout=10)
        response.raise_for_status()
        data = response.json()

        print("Response:")
        print(json.dumps(data, indent=2))
        print()

        if isinstance(data, dict) and not data.get("success", True):
            has_uuid_error = any("uuid" in str(v.get("field", "")).lower() for v in data.get("validation", []))
            if has_uuid_error:
                print("❌ FAILED - Has 'Zero UUID' validation error")
                for v in data.get("validation", []):
                    if "uuid" in str(v.get("field", "")).lower():
                        print(f"   Error: {v.get('message')}")
            else:
                print("❌ FAILED - But NOT due to UUID issue")
        elif isinstance(data, str):
            print(f"✅ SUCCESS - BCM returned UUID: {data}")
        else:
            print("⚠️ UNEXPECTED - Response format unknown")
    except requests.exceptions.RequestException as e:
        print(f"ERROR: API call failed: {e}", file=sys.stderr)

    # Test 2: Software Image WITH UUID
    print("\n" + "=" * 80)
    print("TEST 2: Software Image WITH UUID")
    print("=" * 80)

    generated_uuid = str(py_uuid.uuid4())

    create_entity_with_uuid = {
        "uuid": generated_uuid,
        "name": "uuid-test-with-uuid",
        "path": "/cm/images/uuid-test-with-uuid",
        "baseType": "SoftwareImage",
        "childType": "",
        "modified": True,
        "to_be_removed": False,
        "revision": ""
    }

    create_payload2 = {"service": "CMPart", "call": "addSoftwareImage", "args": [create_entity_with_uuid, False]}

    print("Request Entity (with UUID):")
    print(json.dumps(create_entity_with_uuid, indent=2))
    print()

    try:
        response2 = session.post(f"{endpoint}/json", json=create_payload2, verify=False, timeout=10)
        response2.raise_for_status()
        data2 = response2.json()

        print("Response:")
        print(json.dumps(data2, indent=2))
        print()

        if isinstance(data2, dict) and data2.get("success", False):
            print("✅ SUCCESS - Created with UUID")
            # Cleanup
            delete_payload = {"service": "CMPart", "call": "removeSoftwareImage", "args": [generated_uuid, False]}
            session.post(f"{endpoint}/json", json=delete_payload, verify=False, timeout=10)
            print("Cleaned up test image")
        else:
            print("❌ FAILED - Even with UUID")
    except requests.exceptions.RequestException as e:
        print(f"ERROR: API call failed: {e}", file=sys.stderr)

    # Summary
    print("\n" + "=" * 80)
    print("SUMMARY")
    print("=" * 80)
    print("Compare results above to determine if UUID is required.")
    print("Expected if UUID IS required: Test 1 fails, Test 2 succeeds")
    print("Expected if UUID NOT required: Both tests succeed")
    print()

if __name__ == "__main__":
    main()
