#!/usr/bin/env python3
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

"""
UUID Requirement Testing: All BCM Resource Types

Purpose: Compare UUID requirements across different BCM resource types
(Networks, Categories, Software Images) to understand API consistency.

Usage:
    BCM_ENDPOINT="https://172.21.15.254:8081" \
    BCM_USERNAME="root" \
    BCM_PASSWORD="Hashicorp123!" \
    python3 sampleRest/test_uuid_requirements_all_types.py

Context:
    During issue #26 implementation, discovered that:
    - Networks: REQUIRE UUID (confirmed)
    - Categories: REQUIRE UUID (confirmed in code)
    - Software Images: Unknown (this script tests it)

    This script systematically tests all three to document the pattern.
"""

import os
import sys
import requests
import json
from urllib3.exceptions import InsecureRequestWarning

# Suppress SSL warnings for self-signed certificates
requests.packages.urllib3.disable_warnings(category=InsecureRequestWarning)

def test_resource_without_uuid(session, endpoint, service, call, entity_type, name, **fields):
    """
    Test creating a BCM resource WITHOUT UUID.

    Returns:
        bool: True if "Zero UUID" error found, False otherwise
    """
    entity = {
        "name": name,
        "baseType": entity_type,
        "childType": "",
        "modified": True,
        "to_be_removed": False,
        "revision": "",
        **fields
    }

    payload = {"service": service, "call": call, "args": [entity]}

    try:
        response = session.post(f"{endpoint}/json", json=payload, verify=False, timeout=10)
        response.raise_for_status()
        data = response.json()

        # Check for UUID validation error
        if isinstance(data, dict) and not data.get("success", True):
            validations = data.get("validation", [])
            for v in validations:
                field = str(v.get("field", "")).lower()
                severity = str(v.get("severity", "")).upper()
                if "uuid" in field and severity == "ERROR":
                    return True, v.get("message", "Zero UUID error")

        return False, None

    except requests.exceptions.RequestException as e:
        return None, f"API Error: {e}"

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

    print("=" * 80)
    print("UUID REQUIREMENT TESTING - ALL RESOURCE TYPES")
    print("=" * 80)
    print()

    # Test Network
    print("Testing Network...")
    network_needs_uuid, network_msg = test_resource_without_uuid(
        session, endpoint,
        "cmnet", "addNetwork", "Network", "uuid-test-network",
        domainName="test.cluster"
    )

    if network_needs_uuid is None:
        print(f"  Network:       ⚠️  ERROR - {network_msg}")
    elif network_needs_uuid:
        print(f"  Network:       ✅ REQUIRES UUID")
        print(f"                 └─ Error: {network_msg}")
    else:
        print(f"  Network:       ❌ No UUID required")

    # Test Category
    print("\nTesting Category...")
    category_needs_uuid, category_msg = test_resource_without_uuid(
        session, endpoint,
        "cmdevice", "addCategory", "Category", "uuid-test-category",
        managementNetwork="default"
    )

    if category_needs_uuid is None:
        print(f"  Category:      ⚠️  ERROR - {category_msg}")
    elif category_needs_uuid:
        print(f"  Category:      ✅ REQUIRES UUID")
        print(f"                 └─ Error: {category_msg}")
    else:
        print(f"  Category:      ❌ No UUID required")

    # Test Software Image
    print("\nTesting Software Image...")
    image_needs_uuid, image_msg = test_resource_without_uuid(
        session, endpoint,
        "CMPart", "addSoftwareImage", "SoftwareImage", "uuid-test-image",
        path="/cm/images/uuid-test-image"
    )

    if image_needs_uuid is None:
        print(f"  SoftwareImage: ⚠️  ERROR - {image_msg}")
    elif image_needs_uuid:
        print(f"  SoftwareImage: ✅ REQUIRES UUID")
        print(f"                 └─ Error: {image_msg}")
    else:
        print(f"  SoftwareImage: ❌ No UUID required")

    # Summary
    print()
    print("=" * 80)
    print("SUMMARY")
    print("=" * 80)

    results = {
        "Network": network_needs_uuid,
        "Category": category_needs_uuid,
        "SoftwareImage": image_needs_uuid
    }

    all_require = all(v == True for v in results.values() if v is not None)
    none_require = all(v == False for v in results.values() if v is not None)

    if all_require:
        print("✅ ALL resource types require UUID (consistent API)")
    elif none_require:
        print("❌ NO resource types require UUID (unexpected)")
    else:
        print("⚠️  INCONSISTENT - Some require UUID, some don't")
        print()
        print("Terraform Provider Implementation Implications:")
        for resource_type, needs_uuid in results.items():
            if needs_uuid is None:
                continue
            status = "MUST generate UUID" if needs_uuid else "UUID optional"
            print(f"  - {resource_type}: {status}")

    print()

if __name__ == "__main__":
    main()
