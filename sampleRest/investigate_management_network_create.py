#!/usr/bin/env python3
"""
Investigation: Create fresh resources to test managementNetwork behavior.

Creates category + devices from scratch to confirm:
1. Does BCM store managementNetwork on a device when explicitly sent?
2. What does BCM return when zero UUID is sent?
3. What does BCM return when the field is omitted?
4. Can we update a device's managementNetwork after creation?
5. Can we clear a device's managementNetwork (set back to zero)?
"""

import requests
import json
import os
import uuid
import time
import random
from urllib3.exceptions import InsecureRequestWarning

requests.packages.urllib3.disable_warnings(category=InsecureRequestWarning)

BASE_URL = os.getenv("BCM_ENDPOINT", "https://172.21.15.254:8081")
USERNAME = os.getenv("BCM_USERNAME", "root")
PASSWORD = os.getenv("BCM_PASSWORD", "Hashicorp123!")

ZERO_UUID = "00000000-0000-0000-0000-000000000000"

# Known networks from previous investigation
INTERNALNET_UUID = "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"
MANAGEMENTNET_UUID = "21b20743-d055-42c6-b03c-583c0c061e2e"

# Get partition UUID from default category (needed for device creation)
DEFAULT_PARTITION = "ddd19eb5-f04a-48dc-9cc6-f160b704a7dd"

session = requests.Session()
session.verify = False

# Track resources for cleanup
created_resources = []


def random_mac():
    """Generate a random locally-administered unicast MAC."""
    return "02:00:00:{:02X}:{:02X}:{:02X}".format(
        random.randint(0, 255), random.randint(0, 255), random.randint(0, 255)
    )


def call_api(service, method, *args):
    payload = {
        "service": service,
        "call": method,
        "args": list(args) if args else []
    }
    resp = session.post(f"{BASE_URL}/json", json=payload, timeout=30)
    resp.raise_for_status()
    return resp.json()


def login():
    payload = {"service": "login", "username": USERNAME, "password": PASSWORD}
    resp = session.post(f"{BASE_URL}/json", json=payload, timeout=30)
    resp.raise_for_status()
    print(f"Login: {resp.json()}")


def get_device(hostname):
    """Read back a device by hostname."""
    result = call_api("cmdevice", "getDevice", hostname)
    if isinstance(result, list) and len(result) > 0:
        return result[0]
    elif isinstance(result, dict):
        return result
    return None


def cleanup():
    """Remove all created resources in reverse order."""
    print("\n" + "=" * 80)
    print("CLEANUP")
    print("=" * 80)
    for resource_type, name_or_uuid, label in reversed(created_resources):
        try:
            if resource_type == "device":
                result = call_api("cmdevice", "removeDevice", name_or_uuid)
                print(f"  Removed device '{label}': {result}")
            elif resource_type == "category":
                result = call_api("cmdevice", "removeCategory", name_or_uuid)
                print(f"  Removed category '{label}': {result}")
        except Exception as e:
            print(f"  WARN: Failed to remove {resource_type} '{label}': {e}")


def main():
    login()

    tag = f"mgmtnet-inv-{int(time.time())}"

    # =========================================================================
    # STEP 1: Create a category with managementNetwork = internalnet
    # =========================================================================
    print("\n" + "=" * 80)
    print("STEP 1: Create category with managementNetwork = internalnet")
    print("=" * 80)

    cat_uuid = str(uuid.uuid4())
    cat_name = f"{tag}-cat"
    cat_entity = {
        "baseType": "Category",
        "childType": "",
        "name": cat_name,
        "uuid": cat_uuid,
        "managementNetwork": INTERNALNET_UUID,
        "modified": True,
        "to_be_removed": False,
        "revision": "",
        "softwareImageProxy": {
            "baseType": "SoftwareImageProxy",
            "childType": "",
            "modified": True,
            "to_be_removed": False,
            "revision": "",
            "parentSoftwareImage": "8482c4e9-383c-43de-873f-8c54ee77ee74",
            "uuid": str(uuid.uuid4()),
        },
    }

    result = call_api("cmdevice", "addCategory", cat_entity, True)
    print(f"  addCategory result: {json.dumps(result, indent=2)[:500]}")
    created_resources.append(("category", cat_name, cat_name))

    # Read it back
    time.sleep(1)
    cats = call_api("cmdevice", "getCategories")
    our_cat = None
    for c in cats:
        if c.get("name") == cat_name:
            our_cat = c
            break

    if our_cat:
        print(f"  READ BACK category:")
        print(f"    name: {our_cat.get('name')}")
        print(f"    uuid: {our_cat.get('uuid')}")
        print(f"    managementNetwork: {our_cat.get('managementNetwork')}")
        cat_uuid = our_cat.get("uuid")  # Use BCM's UUID in case it changed
    else:
        print("  ERROR: Category not found after creation!")
        cleanup()
        return

    # =========================================================================
    # STEP 2: Create device A - WITH explicit managementNetwork (managementnet)
    # =========================================================================
    print("\n" + "=" * 80)
    print("STEP 2: Create device A - explicit managementNetwork = managementnet")
    print("=" * 80)

    dev_a_uuid = str(uuid.uuid4())
    dev_a_hostname = f"{tag}-dev-a"
    iface_a_uuid = str(uuid.uuid4())

    dev_a_entity = {
        "baseType": "Device",
        "childType": "PhysicalNode",
        "hostname": dev_a_hostname,
        "mac": random_mac(),
        "category": cat_uuid,
        "managementNetwork": MANAGEMENTNET_UUID,  # Explicit: managementnet
        "modified": True,
        "to_be_removed": False,
        "revision": "",
        "uuid": dev_a_uuid,
        "partition": DEFAULT_PARTITION,
        "provisioningInterface": iface_a_uuid,
        "interfaces": [
            {
                "baseType": "NetworkInterface",
                "childType": "NetworkPhysicalInterface",
                "name": "eth0",
                "network": MANAGEMENTNET_UUID,
                "uuid": iface_a_uuid,
                "modified": True,
                "to_be_removed": False,
                "revision": "",
                "startIf": "ALWAYS",
                "dhcp": False,
            }
        ],
    }

    result = call_api("cmdevice", "addDevice", dev_a_entity, True)
    print(f"  addDevice result: {json.dumps(result, indent=2)[:500]}")
    created_resources.append(("device", dev_a_hostname, dev_a_hostname))

    # Read back
    time.sleep(1)
    dev_a = get_device(dev_a_hostname)
    if dev_a:
        print(f"  READ BACK device A:")
        print(f"    hostname: {dev_a.get('hostname')}")
        print(f"    managementNetwork: {dev_a.get('managementNetwork')}")
        print(f"    category: {dev_a.get('category')}")
        print(f"    SENT managementnet ({MANAGEMENTNET_UUID}), GOT: {dev_a.get('managementNetwork')}")
        print(f"    MATCH: {dev_a.get('managementNetwork') == MANAGEMENTNET_UUID}")
    else:
        print("  ERROR: Device A not found!")

    # =========================================================================
    # STEP 3: Create device B - WITH zero UUID managementNetwork
    # =========================================================================
    print("\n" + "=" * 80)
    print("STEP 3: Create device B - managementNetwork = ZERO UUID")
    print("=" * 80)

    dev_b_uuid = str(uuid.uuid4())
    dev_b_hostname = f"{tag}-dev-b"
    iface_b_uuid = str(uuid.uuid4())

    dev_b_entity = {
        "baseType": "Device",
        "childType": "PhysicalNode",
        "hostname": dev_b_hostname,
        "mac": random_mac(),
        "category": cat_uuid,
        "managementNetwork": ZERO_UUID,  # Zero UUID
        "modified": True,
        "to_be_removed": False,
        "revision": "",
        "uuid": dev_b_uuid,
        "partition": DEFAULT_PARTITION,
        "provisioningInterface": iface_b_uuid,
        "interfaces": [
            {
                "baseType": "NetworkInterface",
                "childType": "NetworkPhysicalInterface",
                "name": "eth0",
                "network": MANAGEMENTNET_UUID,
                "uuid": iface_b_uuid,
                "modified": True,
                "to_be_removed": False,
                "revision": "",
                "startIf": "ALWAYS",
                "dhcp": False,
            }
        ],
    }

    result = call_api("cmdevice", "addDevice", dev_b_entity, True)
    print(f"  addDevice result: {json.dumps(result, indent=2)[:500]}")
    created_resources.append(("device", dev_b_hostname, dev_b_hostname))

    time.sleep(1)
    dev_b = get_device(dev_b_hostname)
    if dev_b:
        print(f"  READ BACK device B:")
        print(f"    hostname: {dev_b.get('hostname')}")
        print(f"    managementNetwork: {dev_b.get('managementNetwork')}")
        print(f"    SENT ZERO UUID, GOT: {dev_b.get('managementNetwork')}")
        print(f"    IS ZERO: {dev_b.get('managementNetwork') == ZERO_UUID}")
    else:
        print("  ERROR: Device B not found!")

    # =========================================================================
    # STEP 4: Create device C - managementNetwork field OMITTED
    # =========================================================================
    print("\n" + "=" * 80)
    print("STEP 4: Create device C - managementNetwork field OMITTED entirely")
    print("=" * 80)

    dev_c_uuid = str(uuid.uuid4())
    dev_c_hostname = f"{tag}-dev-c"
    iface_c_uuid = str(uuid.uuid4())

    dev_c_entity = {
        "baseType": "Device",
        "childType": "PhysicalNode",
        "hostname": dev_c_hostname,
        "mac": random_mac(),
        "category": cat_uuid,
        # NO managementNetwork field
        "modified": True,
        "to_be_removed": False,
        "revision": "",
        "uuid": dev_c_uuid,
        "partition": DEFAULT_PARTITION,
        "provisioningInterface": iface_c_uuid,
        "interfaces": [
            {
                "baseType": "NetworkInterface",
                "childType": "NetworkPhysicalInterface",
                "name": "eth0",
                "network": MANAGEMENTNET_UUID,
                "uuid": iface_c_uuid,
                "modified": True,
                "to_be_removed": False,
                "revision": "",
                "startIf": "ALWAYS",
                "dhcp": False,
            }
        ],
    }

    result = call_api("cmdevice", "addDevice", dev_c_entity, True)
    print(f"  addDevice result: {json.dumps(result, indent=2)[:500]}")
    created_resources.append(("device", dev_c_hostname, dev_c_hostname))

    time.sleep(1)
    dev_c = get_device(dev_c_hostname)
    if dev_c:
        print(f"  READ BACK device C:")
        print(f"    hostname: {dev_c.get('hostname')}")
        print(f"    managementNetwork: {dev_c.get('managementNetwork')}")
        print(f"    OMITTED field, GOT: {dev_c.get('managementNetwork')}")
        print(f"    IS ZERO: {dev_c.get('managementNetwork') == ZERO_UUID}")
    else:
        print("  ERROR: Device C not found!")

    # =========================================================================
    # STEP 5: Update device B (zero) -> set managementNetwork to internalnet
    # =========================================================================
    print("\n" + "=" * 80)
    print("STEP 5: Update device B - change managementNetwork from ZERO to internalnet")
    print("=" * 80)

    if dev_b:
        update_entity = dev_b.copy()
        update_entity["managementNetwork"] = INTERNALNET_UUID
        update_entity["modified"] = True

        result = call_api("cmdevice", "updateDevice", update_entity)
        print(f"  updateDevice result: {json.dumps(result, indent=2)[:500]}")

        time.sleep(1)
        dev_b_updated = get_device(dev_b_hostname)
        if dev_b_updated:
            print(f"  READ BACK device B after update:")
            print(f"    managementNetwork: {dev_b_updated.get('managementNetwork')}")
            print(f"    SENT internalnet ({INTERNALNET_UUID}), GOT: {dev_b_updated.get('managementNetwork')}")
            print(f"    MATCH: {dev_b_updated.get('managementNetwork') == INTERNALNET_UUID}")

    # =========================================================================
    # STEP 6: Update device A (managementnet) -> clear to ZERO UUID
    # =========================================================================
    print("\n" + "=" * 80)
    print("STEP 6: Update device A - clear managementNetwork to ZERO UUID")
    print("=" * 80)

    dev_a_fresh = get_device(dev_a_hostname)
    if dev_a_fresh:
        update_entity = dev_a_fresh.copy()
        update_entity["managementNetwork"] = ZERO_UUID
        update_entity["modified"] = True

        result = call_api("cmdevice", "updateDevice", update_entity)
        print(f"  updateDevice result: {json.dumps(result, indent=2)[:500]}")

        time.sleep(1)
        dev_a_updated = get_device(dev_a_hostname)
        if dev_a_updated:
            print(f"  READ BACK device A after clearing:")
            print(f"    managementNetwork: {dev_a_updated.get('managementNetwork')}")
            print(f"    SENT ZERO UUID, GOT: {dev_a_updated.get('managementNetwork')}")
            print(f"    IS ZERO: {dev_a_updated.get('managementNetwork') == ZERO_UUID}")

    # =========================================================================
    # STEP 7: Summary
    # =========================================================================
    print("\n" + "=" * 80)
    print("SUMMARY")
    print("=" * 80)

    dev_a_final = get_device(dev_a_hostname)
    dev_b_final = get_device(dev_b_hostname)
    dev_c_final = get_device(dev_c_hostname)

    print(f"  Device A (created with managementnet, then cleared to zero):")
    print(f"    managementNetwork = {dev_a_final.get('managementNetwork') if dev_a_final else 'NOT_FOUND'}")

    print(f"  Device B (created with zero, then updated to internalnet):")
    print(f"    managementNetwork = {dev_b_final.get('managementNetwork') if dev_b_final else 'NOT_FOUND'}")

    print(f"  Device C (created with field omitted):")
    print(f"    managementNetwork = {dev_c_final.get('managementNetwork') if dev_c_final else 'NOT_FOUND'}")

    # =========================================================================
    # CLEANUP
    # =========================================================================
    cleanup()

    print("\nDONE")


if __name__ == "__main__":
    main()
