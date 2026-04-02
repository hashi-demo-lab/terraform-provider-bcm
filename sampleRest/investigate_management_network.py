#!/usr/bin/env python3
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

"""
Investigation: How does BCM handle managementNetwork on devices vs categories?

Questions to answer:
1. What does the category's managementNetwork UUID resolve to (which network)?
2. What do existing devices return for managementNetwork?
3. If we create a device with an explicit managementNetwork, does BCM store it?
4. If we create a device WITHOUT managementNetwork (or zero UUID), what does BCM return?
5. Is managementNetwork required in addDevice payload?
"""

import requests
import json
import os
import uuid
from urllib3.exceptions import InsecureRequestWarning

requests.packages.urllib3.disable_warnings(category=InsecureRequestWarning)

BASE_URL = os.getenv("BCM_ENDPOINT", "https://172.21.15.254:8081")
USERNAME = os.getenv("BCM_USERNAME", "root")
PASSWORD = os.getenv("BCM_PASSWORD", "Hashicorp123!")

ZERO_UUID = "00000000-0000-0000-0000-000000000000"

session = requests.Session()
session.verify = False


def call_api(service, method, *args):
    """Make a JSON-RPC call to BCM."""
    payload = {
        "service": service,
        "call": method,
        "args": list(args) if args else []
    }
    resp = session.post(f"{BASE_URL}/json", json=payload, timeout=30)
    resp.raise_for_status()
    data = resp.json()
    return data


def login():
    payload = {"service": "login", "username": USERNAME, "password": PASSWORD}
    resp = session.post(f"{BASE_URL}/json", json=payload, timeout=30)
    resp.raise_for_status()
    result = resp.json()
    print(f"Login: {result}")
    return result


def main():
    login()

    # =========================================================================
    # 1. Get all networks - map UUIDs to names
    # =========================================================================
    print("\n" + "=" * 80)
    print("STEP 1: List all networks (CMNet.getNetworks)")
    print("=" * 80)

    networks_result = call_api("cmnet", "getNetworks")
    networks = {}
    if isinstance(networks_result, list):
        for net in networks_result:
            net_uuid = net.get("uuid", "")
            net_name = net.get("name", "")
            net_type = net.get("childType", net.get("baseType", ""))
            networks[net_uuid] = net_name
            print(f"  Network: {net_name:30s} UUID: {net_uuid}  Type: {net_type}")
    else:
        print(f"  Raw result: {json.dumps(networks_result, indent=2)[:500]}")

    # =========================================================================
    # 2. Get all categories and their managementNetwork
    # =========================================================================
    print("\n" + "=" * 80)
    print("STEP 2: Get categories and their managementNetwork")
    print("=" * 80)

    categories_result = call_api("cmdevice", "getCategories")
    categories = {}
    default_category_uuid = None
    default_mgmt_network = None

    if isinstance(categories_result, list):
        for cat in categories_result:
            cat_uuid = cat.get("uuid", "")
            cat_name = cat.get("name", "")
            mgmt_net = cat.get("managementNetwork", "NOT_PRESENT")
            categories[cat_uuid] = cat

            net_name = networks.get(mgmt_net, "UNKNOWN/ZERO")
            print(f"  Category: {cat_name:20s} UUID: {cat_uuid}")
            print(f"    managementNetwork: {mgmt_net} ({net_name})")

            if cat_name == "default":
                default_category_uuid = cat_uuid
                default_mgmt_network = mgmt_net

    # =========================================================================
    # 3. Get all devices and their managementNetwork
    # =========================================================================
    print("\n" + "=" * 80)
    print("STEP 3: Get all devices and their managementNetwork")
    print("=" * 80)

    devices_result = call_api("cmdevice", "getNodes")
    if isinstance(devices_result, list):
        for dev in devices_result:
            dev_name = dev.get("hostname", dev.get("name", "unknown"))
            dev_uuid = dev.get("uuid", "")
            mgmt_net = dev.get("managementNetwork", "NOT_PRESENT")
            cat_uuid = dev.get("category", "")
            cat_name = "unknown"
            if cat_uuid in categories:
                cat_name = categories[cat_uuid].get("name", "unknown")

            net_name = networks.get(mgmt_net, "ZERO/UNKNOWN")
            print(f"  Device: {dev_name:25s} UUID: {dev_uuid}")
            print(f"    managementNetwork: {mgmt_net} ({net_name})")
            print(f"    category: {cat_name} ({cat_uuid})")

            # Check interfaces for network references
            interfaces = dev.get("interfaces", [])
            for iface in interfaces:
                iface_name = iface.get("name", "")
                iface_net = iface.get("network", "")
                iface_net_name = networks.get(iface_net, "UNKNOWN")
                print(f"    interface: {iface_name} -> network: {iface_net_name} ({iface_net})")
    else:
        print(f"  Raw result: {json.dumps(devices_result, indent=2)[:500]}")

    # =========================================================================
    # 4. Get head node specifically (it always exists)
    # =========================================================================
    print("\n" + "=" * 80)
    print("STEP 4: Get head node details")
    print("=" * 80)

    head_result = call_api("cmdevice", "getDevice", "brght92")
    if isinstance(head_result, dict):
        print(f"  hostname: {head_result.get('hostname')}")
        print(f"  managementNetwork: {head_result.get('managementNetwork')}")
        print(f"  category: {head_result.get('category')}")
        mgmt = head_result.get("managementNetwork", "")
        print(f"  managementNetwork resolves to: {networks.get(mgmt, 'ZERO/NOT_FOUND')}")
    elif isinstance(head_result, list) and len(head_result) > 0:
        dev = head_result[0]
        print(f"  hostname: {dev.get('hostname')}")
        print(f"  managementNetwork: {dev.get('managementNetwork')}")
        mgmt = dev.get("managementNetwork", "")
        print(f"  managementNetwork resolves to: {networks.get(mgmt, 'ZERO/NOT_FOUND')}")

    # =========================================================================
    # 5. Check: does the category's managementNetwork match any device interface network?
    # =========================================================================
    print("\n" + "=" * 80)
    print("STEP 5: Cross-reference category managementNetwork with device interface networks")
    print("=" * 80)

    if default_mgmt_network and default_mgmt_network != ZERO_UUID:
        cat_net_name = networks.get(default_mgmt_network, "UNKNOWN")
        print(f"  Default category managementNetwork: {default_mgmt_network} ({cat_net_name})")

        if isinstance(devices_result, list):
            for dev in devices_result:
                dev_name = dev.get("hostname", "unknown")
                for iface in dev.get("interfaces", []):
                    if iface.get("network") == default_mgmt_network:
                        print(f"  -> Device '{dev_name}' interface '{iface.get('name')}' is on this network")

    # =========================================================================
    # 6. Try getDevice with explicit args to see full schema
    # =========================================================================
    print("\n" + "=" * 80)
    print("STEP 6: Examine a device entity template (getDeviceEntityTemplate or similar)")
    print("=" * 80)

    # Try to get a new device template
    try:
        template_result = call_api("cmdevice", "getNewDeviceEntity")
        print(f"  getNewDeviceEntity: {json.dumps(template_result, indent=2)[:1000]}")
    except Exception as e:
        print(f"  getNewDeviceEntity failed: {e}")

    try:
        template_result = call_api("cmdevice", "getDeviceTemplate")
        print(f"  getDeviceTemplate: {json.dumps(template_result, indent=2)[:1000]}")
    except Exception as e:
        print(f"  getDeviceTemplate failed: {e}")

    # =========================================================================
    # 7. KEY TEST: Validate a device entity WITH and WITHOUT managementNetwork
    # =========================================================================
    print("\n" + "=" * 80)
    print("STEP 7: Validate device entity with different managementNetwork values")
    print("=" * 80)

    test_uuid = str(uuid.uuid4())

    # Base entity for validation
    base_entity = {
        "baseType": "Device",
        "childType": "PhysicalNode",
        "hostname": "mgmtnet-test-node",
        "mac": "00:00:00:00:00:00",
        "category": default_category_uuid,
        "modified": True,
        "to_be_removed": False,
        "revision": "",
        "uuid": test_uuid,
        "partition": "",
        "interfaces": [],
    }

    # Test A: With zero UUID
    entity_a = base_entity.copy()
    entity_a["managementNetwork"] = ZERO_UUID
    print("\n  Test A: managementNetwork = ZERO UUID")
    try:
        result = call_api("cmdevice", "validateDevice", entity_a, True)
        result = call_api("cmdevice", "validateDevice", entity_a, True)
        print(f"    Result: {json.dumps(result, indent=2)}")
    except Exception as e:
        print(f"    Error: {e}")

    # Test B: With category's management network (internalnet)
    entity_b = base_entity.copy()
    entity_b["managementNetwork"] = default_mgmt_network
    print(f"\n  Test B: managementNetwork = category's network ({default_mgmt_network}) = internalnet")
    try:
        result = call_api("cmdevice", "validateDevice", entity_b, True)
        print(f"    Result: {json.dumps(result, indent=2)}")
    except Exception as e:
        print(f"    Error: {e}")

    # Test C: Without managementNetwork field at all
    entity_c = base_entity.copy()
    print("\n  Test C: managementNetwork field OMITTED")
    try:
        result = call_api("cmdevice", "validateDevice", entity_c, True)
        print(f"    Result: {json.dumps(result, indent=2)}")
    except Exception as e:
        print(f"    Error: {e}")

    # Test D: With a random (non-existent) UUID
    random_uuid = str(uuid.uuid4())
    entity_d = base_entity.copy()
    entity_d["managementNetwork"] = random_uuid
    print(f"\n  Test D: managementNetwork = random non-existent UUID ({random_uuid})")
    try:
        result = call_api("cmdevice", "validateDevice", entity_d, True)
        print(f"    Result: {json.dumps(result, indent=2)}")
    except Exception as e:
        print(f"    Error: {e}")

    # Test E: With managementnet UUID (different from category default)
    mgmtnet_uuid = "21b20743-d055-42c6-b03c-583c0c061e2e"
    entity_e = base_entity.copy()
    entity_e["managementNetwork"] = mgmtnet_uuid
    print(f"\n  Test E: managementNetwork = managementnet UUID ({mgmtnet_uuid})")
    try:
        result = call_api("cmdevice", "validateDevice", entity_e, True)
        print(f"    Result: {json.dumps(result, indent=2)}")
    except Exception as e:
        print(f"    Error: {e}")

    # =========================================================================
    # 8. KEY: Examine the device that HAS a non-zero managementNetwork
    # =========================================================================
    print("\n" + "=" * 80)
    print("STEP 8: Examine device with non-zero managementNetwork")
    print("=" * 80)

    # This device was created by the management network passthrough test
    mgmt_device_uuid = "0d80ebd9-c19a-488a-bf27-3abba65e0ef9"
    try:
        result = call_api("cmdevice", "getDevice", "tftest-device-mgmtdrift-20260402-151058-413069000-88223")
        if isinstance(result, list) and len(result) > 0:
            dev = result[0]
        elif isinstance(result, dict):
            dev = result
        else:
            dev = result

        print(f"  Full device entity:")
        # Print key fields
        for key in ["hostname", "uuid", "managementNetwork", "category", "partition",
                     "provisioningInterface", "mac"]:
            print(f"    {key}: {dev.get(key, 'NOT_PRESENT')}")

        mgmt = dev.get("managementNetwork", "")
        print(f"    managementNetwork resolves to: {networks.get(mgmt, 'ZERO/NOT_FOUND')}")

        # Compare with its category
        cat_uuid = dev.get("category", "")
        if cat_uuid in categories:
            cat = categories[cat_uuid]
            cat_mgmt = cat.get("managementNetwork", "")
            print(f"    category managementNetwork: {cat_mgmt} ({networks.get(cat_mgmt, 'UNKNOWN')})")
            print(f"    MATCH: {mgmt == cat_mgmt}")
    except Exception as e:
        print(f"  Error: {e}")

    # =========================================================================
    # 9. Compare: device with zero UUID vs device with real UUID
    # =========================================================================
    print("\n" + "=" * 80)
    print("STEP 9: Compare devices - zero UUID vs real UUID")
    print("=" * 80)

    if isinstance(devices_result, list):
        zero_devices = [d for d in devices_result if d.get("managementNetwork") == ZERO_UUID]
        real_devices = [d for d in devices_result if d.get("managementNetwork") != ZERO_UUID]

        print(f"  Devices with ZERO managementNetwork: {len(zero_devices)}")
        for d in zero_devices:
            cat_uuid = d.get("category", "")
            cat_mgmt = categories.get(cat_uuid, {}).get("managementNetwork", "N/A")
            iface_nets = [networks.get(i.get("network", ""), "?") for i in d.get("interfaces", [])]
            print(f"    {d.get('hostname'):50s} iface_networks={iface_nets}  cat_mgmt={networks.get(cat_mgmt, cat_mgmt)}")

        print(f"\n  Devices with REAL managementNetwork: {len(real_devices)}")
        for d in real_devices:
            mgmt = d.get("managementNetwork", "")
            cat_uuid = d.get("category", "")
            cat_mgmt = categories.get(cat_uuid, {}).get("managementNetwork", "N/A")
            iface_nets = [networks.get(i.get("network", ""), "?") for i in d.get("interfaces", [])]
            print(f"    {d.get('hostname'):50s} mgmt={networks.get(mgmt, mgmt)}  iface_networks={iface_nets}  cat_mgmt={networks.get(cat_mgmt, cat_mgmt)}")

    print("\n" + "=" * 80)
    print("INVESTIGATION COMPLETE")
    print("=" * 80)


if __name__ == "__main__":
    main()
