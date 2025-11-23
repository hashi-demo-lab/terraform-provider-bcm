#!/usr/bin/env python3
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

"""Test BCM partition naming constraints"""

import os
import sys
import json
import requests
from urllib3.exceptions import InsecureRequestWarning

requests.packages.urllib3.disable_warnings(category=InsecureRequestWarning)

def main():
    endpoint = os.getenv("BCM_ENDPOINT", "https://172.21.15.254:8081")
    username = os.getenv("BCM_USERNAME", "root")
    password = os.getenv("BCM_PASSWORD")

    if not password:
        print("ERROR: BCM_PASSWORD not set", file=sys.stderr)
        sys.exit(1)

    session = requests.Session()

    # Login
    print(f"Authenticating to {endpoint}...")
    login_response = session.post(
        f"{endpoint}/json",
        json={"service": "login", "username": username, "password": password},
        verify=False,
        timeout=10
    )
    login_response.raise_for_status()
    print("✓ Authenticated")

    # Get existing partition
    print("\nFetching existing partitions...")
    partitions_response = session.post(
        f"{endpoint}/json",
        json={"service": "cmpart", "call": "getPartitions"},
        verify=False,
        timeout=10
    )
    partitions = partitions_response.json()

    if partitions:
        base_partition = partitions[0]
        print(f"✓ Found partition: {base_partition['name']} (UUID: {base_partition['uuid']})")
        print(f"  Cluster Name: {base_partition['clusterName']}")
        print(f"  Slave Name: {base_partition['slaveName']}")
        print(f"  Slave Digits: {base_partition['slaveDigits']}")

    # Test 1: Try to create partition with name "test-partition-unique"
    print("\n" + "="*80)
    print("TEST 1: Create partition with name 'test-partition-unique'")
    print("="*80)

    test_partition = {
        "baseType": "Partition",
        "childType": "",
        "modified": True,
        "to_be_removed": False,
        "revision": "",
        "name": "test-partition-unique",
        "clusterName": "Test Cluster",
        "slaveName": "node",
        "slaveDigits": 3,
        "timeZoneSettings": "America/Los_Angeles",
        "primaryHeadNode": base_partition["primaryHeadNode"],
        "externalNetwork": base_partition["externalNetwork"],
        "defaultCategory": base_partition["defaultCategory"],
        "managementNetwork": base_partition["managementNetwork"]
    }

    try:
        create_response = session.post(
            f"{endpoint}/json",
            json={"service": "cmpart", "call": "addPartition", "args": [test_partition, False]},
            verify=False,
            timeout=10
        )
        result = create_response.json()
        print(f"Response: {json.dumps(result, indent=2)}")

        if isinstance(result, dict) and result.get("error"):
            print(f"❌ FAILED: {result['error']}")
        else:
            print(f"✅ SUCCESS: Partition created")
    except Exception as e:
        print(f"❌ EXCEPTION: {e}")

    # Test 2: Try to create partition with very long name
    print("\n" + "="*80)
    print("TEST 2: Create partition with long name (256 chars)")
    print("="*80)

    long_name = "a" * 256
    test_partition["name"] = long_name

    try:
        create_response = session.post(
            f"{endpoint}/json",
            json={"service": "cmpart", "call": "addPartition", "args": [test_partition, False]},
            verify=False,
            timeout=10
        )
        result = create_response.json()

        if isinstance(result, dict) and result.get("error"):
            print(f"❌ FAILED: {result['error']}")
        else:
            print(f"✅ Unexpectedly succeeded with 256 char name")
    except Exception as e:
        print(f"❌ EXCEPTION: {e}")

    # Test 3: Try with 255 char name
    print("\n" + "="*80)
    print("TEST 3: Create partition with name at 255 char limit")
    print("="*80)

    limit_name = "a" * 255
    test_partition["name"] = limit_name

    try:
        create_response = session.post(
            f"{endpoint}/json",
            json={"service": "cmpart", "call": "addPartition", "args": [test_partition, False]},
            verify=False,
            timeout=10
        )
        result = create_response.json()

        if isinstance(result, dict) and result.get("error"):
            print(f"❌ FAILED: {result['error']}")
        else:
            print(f"✅ Unexpectedly succeeded with 255 char name")
    except Exception as e:
        print(f"❌ EXCEPTION: {e}")

    # Test 4: Try to create duplicate "base"
    print("\n" + "="*80)
    print("TEST 4: Create duplicate partition named 'base'")
    print("="*80)

    test_partition["name"] = "base"

    try:
        create_response = session.post(
            f"{endpoint}/json",
            json={"service": "cmpart", "call": "addPartition", "args": [test_partition, False]},
            verify=False,
            timeout=10
        )
        result = create_response.json()

        if isinstance(result, dict) and result.get("error"):
            print(f"❌ EXPECTED FAILURE: {result['error']}")
        else:
            print(f"⚠️ Unexpectedly succeeded creating duplicate 'base'")
    except Exception as e:
        print(f"❌ EXCEPTION: {e}")

if __name__ == "__main__":
    main()
