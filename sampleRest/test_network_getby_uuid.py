#!/usr/bin/env python3
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

"""
Test BCM CMNet getNetwork with args parameter (direct lookup by UUID)

Purpose: Validate that getNetwork(uuid) supports args parameter for efficient direct lookup

Usage:
    BCM_ENDPOINT="https://172.21.15.254:8081" \
    BCM_USERNAME="root" \
    BCM_PASSWORD="Hashicorp123!" \
    python3 sampleRest/test_network_getby_uuid.py
"""

import os
import sys
import json
import requests
from urllib3.exceptions import InsecureRequestWarning

# Suppress SSL warnings
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
    print("✓ Authentication successful\n")

    # Step 1: Get all networks to find one to test with
    print("=" * 80)
    print("Step 1: Getting all networks via getNetworks()")
    print("=" * 80)

    networks_response = session.post(
        f"{endpoint}/json",
        json={"service": "cmnet", "call": "getNetworks"},
        verify=False,
        timeout=10
    )
    networks_response.raise_for_status()
    networks = networks_response.json()

    if not networks:
        print("ERROR: No networks found in BCM cluster", file=sys.stderr)
        sys.exit(1)

    print(f"✓ Found {len(networks)} network(s)\n")

    # Pick first network for testing
    test_network = networks[0]
    network_uuid = test_network.get("uuid")
    network_name = test_network.get("name")

    print(f"Test network: {network_name}")
    print(f"UUID: {network_uuid}\n")

    # Step 2: Test getNetwork with args parameter (direct UUID lookup)
    print("=" * 80)
    print("Step 2: Testing getNetwork(uuid) with args parameter")
    print("=" * 80)

    get_network_payload = {
        "service": "cmnet",
        "call": "getNetwork",
        "args": [network_uuid]
    }

    print(f"Request: {json.dumps(get_network_payload, indent=2)}\n")

    try:
        get_response = session.post(
            f"{endpoint}/json",
            json=get_network_payload,
            verify=False,
            timeout=10
        )
        get_response.raise_for_status()

        network_detail = get_response.json()

        print("✓ SUCCESS: getNetwork(uuid) with args parameter works!\n")
        print(f"Retrieved network: {network_detail.get('name')}")
        print(f"UUID match: {network_detail.get('uuid') == network_uuid}\n")

        # Analyze for VLAN fields
        print("=" * 80)
        print("Field Analysis for VLAN Support")
        print("=" * 80)

        vlan_fields = []
        for key in sorted(network_detail.keys()):
            if 'vlan' in key.lower():
                vlan_fields.append(key)
                print(f"  ✓ VLAN field found: {key} = {network_detail[key]}")

        if not vlan_fields:
            print("  ✗ No VLAN-related fields found in network object\n")
        else:
            print(f"\n✓ VLAN support: {len(vlan_fields)} field(s) found\n")

        # Print full network structure
        print("=" * 80)
        print("Full Network Object Structure")
        print("=" * 80)
        print(json.dumps(network_detail, indent=2))

    except requests.exceptions.RequestException as e:
        print(f"✗ FAILED: getNetwork with args parameter failed: {e}", file=sys.stderr)
        if hasattr(e.response, 'text'):
            print(f"Response: {e.response.text}")
        sys.exit(1)

    # Summary
    print("\n" + "=" * 80)
    print("CONCLUSION")
    print("=" * 80)
    print("✓ getNetwork(uuid) args parameter: SUPPORTED")
    print(f"✓ VLAN support: {'YES (' + ', '.join(vlan_fields) + ')' if vlan_fields else 'NO'}")
    print("\nRecommendation for resource implementation:")
    print("  - Use getNetwork(uuid) for Read operations (efficient direct lookup)")
    print("  - Use getNetworks() for data source list operations")
    if vlan_fields:
        print(f"  - Include VLAN attribute in schema: {vlan_fields[0]}")
    else:
        print("  - Omit VLAN from schema (not supported by BCM API)")
    print()

if __name__ == "__main__":
    main()
