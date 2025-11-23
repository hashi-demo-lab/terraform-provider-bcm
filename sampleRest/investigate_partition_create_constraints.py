#!/usr/bin/env python3
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

"""
Partition Creation Constraints Investigation

This script investigates BCM partition creation constraints:
1. What timezone values are accepted?
2. Can we create partitions with names other than "base"?
3. What are the name length limits?
4. What fields are required for creation?

Usage:
    BCM_ENDPOINT="https://172.21.15.254:8081" \
    BCM_USERNAME="root" \
    BCM_PASSWORD="Hashicorp123!" \
    python3 sampleRest/investigate_partition_create_constraints.py
"""

import os
import sys
import json
import requests
from urllib3.exceptions import InsecureRequestWarning

# Suppress SSL warnings for self-signed certificates
requests.packages.urllib3.disable_warnings(category=InsecureRequestWarning)

def call_api(session, endpoint, service, method, *args):
    """Make BCM API call"""
    url = f"{endpoint}/json"
    payload = {
        "service": service,
        "call": method
    }

    if args:
        payload["args"] = args

    response = session.post(url, json=payload, verify=False, timeout=30)
    response.raise_for_status()
    return response.json()

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

    # Login
    print(f"Authenticating to {endpoint}...")
    login_payload = {
        "service": "login",
        "username": username,
        "password": password
    }

    login_response = session.post(
        f"{endpoint}/json",
        json=login_payload,
        verify=False,
        timeout=10
    )
    login_response.raise_for_status()
    print("✓ Authentication successful\n")

    # Investigation 1: Get existing partitions
    print("="*80)
    print("INVESTIGATION 1: Existing Partitions")
    print("="*80)

    partitions = call_api(session, endpoint, "cmpart", "getPartitions")
    print(f"\nFound {len(partitions)} partition(s):\n")

    for partition in partitions:
        print(f"Name: {partition.get('name')}")
        print(f"UUID: {partition.get('uuid')}")
        print(f"Timezone: {partition.get('timeZoneSettings')}")
        print(f"Cluster Name: {partition.get('clusterName')}")
        print(f"Creation Time: {partition.get('creationTime')}")
        print(f"Base Type: {partition.get('baseType')}")
        print(f"Child Type: {partition.get('childType')}")
        print()

    # Investigation 2: Analyze timezone field from existing partition
    if partitions:
        base_partition = partitions[0]
        timezone = base_partition.get('timeZoneSettings')

        print("="*80)
        print("INVESTIGATION 2: Timezone Analysis")
        print("="*80)
        print(f"\nExisting partition uses timezone: {repr(timezone)}")
        print(f"Timezone type: {type(timezone).__name__}")

        # Common timezone variations to document
        print("\nCommon IANA timezone formats:")
        print("  - America/Los_Angeles  (Pacific Time)")
        print("  - America/New_York     (Eastern Time)")
        print("  - UTC                  (Coordinated Universal Time)")
        print("  - Europe/London        (GMT)")
        print("  - US/Pacific           (Deprecated, use America/Los_Angeles)")
        print()

    # Investigation 3: Analyze all fields from existing partition
    print("="*80)
    print("INVESTIGATION 3: Full Partition Schema")
    print("="*80)

    if partitions:
        partition = partitions[0]
        print("\nAll fields in existing partition:")
        for key in sorted(partition.keys()):
            value = partition[key]
            value_type = type(value).__name__

            # Show preview of value
            if isinstance(value, str):
                if len(value) > 100:
                    value_preview = value[:100] + "..."
                else:
                    value_preview = repr(value)
            elif isinstance(value, list):
                value_preview = f"[{len(value)} items]"
            else:
                value_preview = repr(value)

            print(f"  {key:25s} ({value_type:10s}): {value_preview}")

        # Save full partition data
        with open("partition_full_data.json", "w") as f:
            json.dump(partition, f, indent=2)
        print("\n✓ Saved full partition data to: partition_full_data.json")

    # Investigation 4: Required fields for partition creation
    print("\n" + "="*80)
    print("INVESTIGATION 4: Required Fields Analysis")
    print("="*80)

    if partitions:
        partition = partitions[0]

        print("\nKey fields for partition creation:")
        print(f"  name:                {partition.get('name')}")
        print(f"  clusterName:         {partition.get('clusterName')}")
        print(f"  timeZoneSettings:    {partition.get('timeZoneSettings')}")
        print(f"  primaryHeadNode:     {partition.get('primaryHeadNode')}")
        print(f"  externalNetwork:     {partition.get('externalNetwork')}")
        print(f"  defaultCategory:     {partition.get('defaultCategory')}")
        print(f"  managementNetwork:   {partition.get('managementNetwork')}")
        print(f"  slaveName:           {partition.get('slaveName')}")
        print(f"  slaveDigits:         {partition.get('slaveDigits')}")

    # Investigation 5: Test timezone variants (read-only, no modification)
    print("\n" + "="*80)
    print("INVESTIGATION 5: Timezone Recommendations")
    print("="*80)

    if partitions:
        current_tz = partitions[0].get('timeZoneSettings')
        print(f"\nCurrent timezone in use: {repr(current_tz)}")
        print("\nRecommendation for tests:")
        print(f"  Use: timezone_settings = {repr(current_tz)}")
        print("\nAlternatives (if current doesn't work):")
        print("  - 'UTC'")
        print("  - 'America/New_York'")
        print("  - 'Europe/London'")

    # Investigation 6: Name constraints
    print("\n" + "="*80)
    print("INVESTIGATION 6: Partition Name Constraints")
    print("="*80)

    print(f"\nExisting partition names:")
    for p in partitions:
        name = p.get('name')
        print(f"  - {repr(name)} (length: {len(name)})")

    print("\nObservations:")
    if len(partitions) == 1 and partitions[0].get('name') == 'base':
        print("  ⚠️  Only one partition exists named 'base'")
        print("  ⚠️  BCM may restrict partition creation to single 'base' partition")
        print("  ⚠️  Tests should use Import/Update pattern, not Create")
    else:
        print("  ✓ Multiple partitions exist - unique names allowed")
        print("  ✓ Tests can create partitions with unique names")

    print("\n" + "="*80)
    print("SUMMARY: Test Strategy Recommendations")
    print("="*80)

    if partitions:
        tz = partitions[0].get('timeZoneSettings')
        print(f"\n1. Timezone: Use {repr(tz)} instead of 'America/Los_Angeles'")

    if len(partitions) == 1 and partitions[0].get('name') == 'base':
        print("\n2. Name Strategy: Use 'base' (existing partition)")
        print("   - Don't test Create (partition already exists)")
        print("   - Test Import → Update → Read")
        print("   - Skip CheckDestroy (don't delete base partition)")
    else:
        print("\n2. Name Strategy: Use unique names")
        print("   - generateUniqueTestName('test-partition') should work")
        print("   - Full CRUD testing possible")

    print("\n3. Test Pattern:")
    print("   - Use Import-based tests for existing 'base' partition")
    print("   - Update non-critical fields (clusterName, notes, etc.)")
    print("   - Avoid deleting 'base' partition in CheckDestroy")

    print("\n✓ Investigation complete")

if __name__ == "__main__":
    main()
