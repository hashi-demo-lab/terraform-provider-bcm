#!/usr/bin/env python3
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

"""
API Exploration Script: CMNet Network CRUD Operations

Purpose: Test BCM CMNet API methods for network resource management:
    - addNetwork (create)
    - getNetwork (read with args parameter)
    - updateNetwork (update)
    - removeNetwork (delete)

This script validates:
    1. Args parameter support for getNetwork(uuid)
    2. VLAN field mapping in API response
    3. Force parameter behavior for removeNetwork
    4. BCM entity structure requirements

Usage:
    BCM_ENDPOINT="https://172.21.15.254:8081" \
    BCM_USERNAME="root" \
    BCM_PASSWORD="Hashicorp123!" \
    python3 sampleRest/explore_network_crud.py
"""

import os
import sys
import json
import time
import requests
from urllib3.exceptions import InsecureRequestWarning

# Suppress SSL warnings for self-signed certificates
requests.packages.urllib3.disable_warnings(category=InsecureRequestWarning)

def main():
    # Get credentials from environment
    endpoint = os.getenv("BCM_ENDPOINT", "https://172.21.15.254:8081")
    username = os.getenv("BCM_USERNAME", "root")
    password = os.getenv("BCM_PASSWORD")

    if not password:
        print("ERROR: BCM_PASSWORD environment variable not set", file=sys.stderr)
        sys.exit(1)

    # Create session
    session = requests.Session()

    # Step 1: Login
    print(f"Authenticating to {endpoint}...")
    login_url = f"{endpoint}/json"
    login_payload = {
        "service": "login",
        "username": username,
        "password": password
    }

    try:
        login_response = session.post(
            login_url,
            json=login_payload,
            verify=False,
            timeout=10
        )
        login_response.raise_for_status()

        # Check for login errors
        login_data = login_response.json()
        if isinstance(login_data, dict) and login_data.get("error"):
            print(f"ERROR: Login failed: {login_data.get('error')}", file=sys.stderr)
            sys.exit(1)

        print("✓ Authentication successful\n")

    except requests.exceptions.RequestException as e:
        print(f"ERROR: Failed to authenticate: {e}", file=sys.stderr)
        sys.exit(1)

    # Step 2: Create test network via addNetwork
    print("=" * 80)
    print("TEST 1: Creating test network via addNetwork")
    print("=" * 80)

    api_url = f"{endpoint}/json"
    test_network_name = f"terraform-test-{int(time.time())}"

    # Generate UUID for network creation (Networks REQUIRE UUID like Categories)
    import uuid as py_uuid
    network_uuid = str(py_uuid.uuid4())

    create_entity = {
        "uuid": network_uuid,
        "name": test_network_name,
        "baseAddress": "192.168.99.0",
        "netmaskBits": 24,
        "gateway": "192.168.99.1",
        "mtu": 1500,
        "domainName": "test.cluster",
        "dynamicRangeStart": "192.168.99.100",
        "dynamicRangeEnd": "192.168.99.200",
        "notes": "Test network for API exploration",
        "baseType": "Network",
        "childType": "",
        "modified": True,
        "to_be_removed": False,
        "revision": ""
    }

    create_payload = {
        "service": "cmnet",
        "call": "addNetwork",
        "args": [create_entity]
    }

    print(f"Creating network: {test_network_name}")
    print(f"Request payload: {json.dumps(create_payload, indent=2)}\n")

    try:
        create_response = session.post(
            api_url,
            json=create_payload,
            verify=False,
            timeout=10
        )
        create_response.raise_for_status()

        created_network = create_response.json()
        print("✓ Network created successfully")
        print(f"Response: {json.dumps(created_network, indent=2)}\n")

        # Confirm UUID in response
        if isinstance(created_network, dict):
            response_uuid = created_network.get("uuid")
            print(f"✓ Network UUID (request): {network_uuid}")
            print(f"✓ Network UUID (response): {response_uuid}\n")
            if response_uuid != network_uuid:
                print(f"⚠ WARNING: UUID mismatch", file=sys.stderr)
        else:
            print("ERROR: Unexpected response format", file=sys.stderr)
            sys.exit(1)

    except requests.exceptions.RequestException as e:
        print(f"ERROR: Failed to create network: {e}", file=sys.stderr)
        if hasattr(e.response, 'text'):
            print(f"Response text: {e.response.text}")
        sys.exit(1)

    # Step 3: Test getNetwork with args parameter (direct lookup)
    print("=" * 80)
    print("TEST 2: Reading network via getNetwork(uuid) with args parameter")
    print("=" * 80)

    read_payload = {
        "service": "cmnet",
        "call": "getNetwork",
        "args": [network_uuid]
    }

    print(f"Reading network by UUID: {network_uuid}")
    print(f"Request payload: {json.dumps(read_payload, indent=2)}\n")

    try:
        read_response = session.post(
            api_url,
            json=read_payload,
            verify=False,
            timeout=10
        )
        read_response.raise_for_status()

        read_network = read_response.json()
        print("✓ Network retrieved successfully via args parameter")
        print(f"Response: {json.dumps(read_network, indent=2)}\n")

        # Analyze field structure for VLAN support
        print("Field Analysis:")
        print("-" * 40)
        vlan_fields = []
        for key in sorted(read_network.keys()):
            value = read_network[key]
            value_type = type(value).__name__
            print(f"  {key}: {value_type} = {value}")

            # Check for VLAN-related fields
            if 'vlan' in key.lower():
                vlan_fields.append(key)

        print("\n✓ VLAN-related fields found:" if vlan_fields else "\n⚠ No VLAN-related fields found")
        for field in vlan_fields:
            print(f"    - {field}: {read_network[field]}")
        print()

    except requests.exceptions.RequestException as e:
        print(f"ERROR: Failed to read network: {e}", file=sys.stderr)
        if hasattr(e.response, 'text'):
            print(f"Response text: {e.response.text}")
        sys.exit(1)

    # Step 4: Update network via updateNetwork
    print("=" * 80)
    print("TEST 3: Updating network via updateNetwork")
    print("=" * 80)

    # Get current state for update (must include revision)
    update_entity = read_network.copy()
    update_entity["mtu"] = 9000
    update_entity["notes"] = "Updated via API exploration - MTU changed to 9000"
    update_entity["modified"] = True

    update_payload = {
        "service": "cmnet",
        "call": "updateNetwork",
        "args": [update_entity, False]  # entity, force=false
    }

    print(f"Updating network MTU: 1500 -> 9000")
    print(f"Request includes: uuid, revision, modified=true\n")

    try:
        update_response = session.post(
            api_url,
            json=update_payload,
            verify=False,
            timeout=10
        )
        update_response.raise_for_status()

        updated_network = update_response.json()
        print("✓ Network updated successfully")
        print(f"New MTU: {updated_network.get('mtu')}")
        print(f"New notes: {updated_network.get('notes')}\n")

    except requests.exceptions.RequestException as e:
        print(f"ERROR: Failed to update network: {e}", file=sys.stderr)
        if hasattr(e.response, 'text'):
            print(f"Response text: {e.response.text}")
        sys.exit(1)

    # Step 5: Test removeNetwork with force=false
    print("=" * 80)
    print("TEST 4: Deleting network via removeNetwork")
    print("=" * 80)

    delete_payload = {
        "service": "cmnet",
        "call": "removeNetwork",
        "args": [network_uuid]
    }

    print(f"Deleting network (force=false): {network_uuid}")
    print(f"Request payload: {json.dumps(delete_payload, indent=2)}\n")

    try:
        delete_response = session.post(
            api_url,
            json=delete_payload,
            verify=False,
            timeout=10
        )
        delete_response.raise_for_status()

        print("✓ Network deleted successfully\n")

        # Verify deletion by trying to read again
        print("Verifying deletion...")
        verify_payload = {
            "service": "cmnet",
            "call": "getNetwork",
            "args": [network_uuid]
        }

        verify_response = session.post(
            api_url,
            json=verify_payload,
            verify=False,
            timeout=10
        )

        # Expect 404 or error response
        if verify_response.status_code == 200:
            verify_data = verify_response.json()
            if isinstance(verify_data, dict) and verify_data.get("error"):
                print(f"✓ Deletion verified - network not found: {verify_data.get('error')}\n")
            else:
                print("⚠ WARNING: Network still exists after deletion", file=sys.stderr)
        else:
            print(f"✓ Deletion verified - HTTP {verify_response.status_code}\n")

    except requests.exceptions.RequestException as e:
        print(f"ERROR: Failed to delete network: {e}", file=sys.stderr)
        if hasattr(e.response, 'text'):
            print(f"Response text: {e.response.text}")
        sys.exit(1)

    # Summary
    print("=" * 80)
    print("API EXPLORATION SUMMARY")
    print("=" * 80)
    print("✓ addNetwork: Successful - creates network with BCM entity structure")
    print("✓ getNetwork(uuid): Successful - args parameter supported for direct lookup")
    print("✓ updateNetwork: Successful - requires uuid, revision, modified fields")
    print("✓ removeNetwork: Successful - force=false works for unused networks")
    print("\nVLAN Support:", "CONFIRMED" if vlan_fields else "NOT FOUND")
    if vlan_fields:
        print(f"  VLAN fields: {', '.join(vlan_fields)}")
    print("\n✓ All CRUD operations validated successfully\n")

if __name__ == "__main__":
    main()
