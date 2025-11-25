#!/usr/bin/env python3
"""
Explore BCM API for role management methods
Tests various method patterns to discover the correct API for roles
"""

import requests
import urllib3
import sys

# Disable SSL warnings for self-signed certificates
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

class BCMApiClient:
    def __init__(self, base_url, username, password):
        self.base_url = base_url
        self.username = username
        self.password = password
        self.session = requests.Session()
        self.session.verify = False

    def login(self):
        """Login to BCM API and establish session"""
        url = f"{self.base_url}/json"
        payload = {
            "service": "login",
            "username": self.username,
            "password": self.password
        }

        try:
            response = self.session.post(url, json=payload, verify=False)
            response.raise_for_status()
            print("✅ Login successful")
            return True
        except Exception as e:
            print(f"❌ Login failed: {e}")
            return False

    def call_api(self, service, call, *args):
        """Make a JSON-RPC API call"""
        url = f"{self.base_url}/json"
        payload = {
            "service": service,
            "call": call
        }

        if args:
            payload["args"] = list(args)

        try:
            response = self.session.post(url, json=payload, verify=False)
            response.raise_for_status()
            return response.json()
        except Exception as e:
            return None

def main():
    # Initialize client
    client = BCMApiClient(
        base_url="https://172.21.15.254:8081",
        username="root",
        password="Hashicorp123!"
    )

    if not client.login():
        sys.exit(1)

    print("\n" + "="*80)
    print("EXPLORING ROLES API")
    print("="*80)

    # Test methods to discover role API
    test_methods = [
        # List all roles
        ("cmdevice", "getRoles", []),
        ("cmdevice", "getRoles", None),
        ("cmdevice", "getAllRoles", []),
        ("cmdevice", "listRoles", []),

        # Get single role
        ("cmdevice", "getRole", ["headnode"]),
        ("cmdevice", "getRole", ["ComputeRole"]),

        # Get role types
        ("cmdevice", "getRoleTypes", []),
        ("cmdevice", "getAvailableRoles", []),
    ]

    results = []

    for service, method, args in test_methods:
        print(f"\n🔍 Testing: {service}.{method}({args if args else ''})")

        if args is None:
            result = client.call_api(service, method)
        elif args:
            result = client.call_api(service, method, *args)
        else:
            result = client.call_api(service, method)

        if result is not None:
            print(f"✅ SUCCESS - {service}.{method}")
            print(f"   Response type: {type(result)}")
            if isinstance(result, list):
                print(f"   Items returned: {len(result)}")
                if result:
                    print(f"   First item keys: {list(result[0].keys()) if isinstance(result[0], dict) else 'N/A'}")
                    print(f"   Sample item: {result[0]}")
            elif isinstance(result, dict):
                print(f"   Keys: {list(result.keys())}")
                print(f"   Response: {result}")

            results.append({
                "service": service,
                "method": method,
                "args": args,
                "success": True,
                "result": result
            })
        else:
            print(f"❌ FAILED - {service}.{method}")

    # Try to get roles from a node
    print("\n" + "="*80)
    print("CHECKING ROLES IN NODE DATA")
    print("="*80)

    node = client.call_api("cmdevice", "getNode", "master")
    if node and "roles" in node:
        print(f"\n✅ Found roles in node data")
        print(f"   Node: {node.get('hostname', 'N/A')}")
        print(f"   Roles count: {len(node['roles'])}")
        print(f"\n   Roles structure:")
        for role in node['roles']:
            print(f"   - {role}")

    # Summary
    print("\n" + "="*80)
    print("SUMMARY")
    print("="*80)

    successful = [r for r in results if r['success']]

    if successful:
        print(f"\n✅ Found {len(successful)} working method(s):")
        for r in successful:
            print(f"   - {r['service']}.{r['method']}")
    else:
        print("\n❌ No working methods found")
        print("\n💡 Roles may be:")
        print("   1. Embedded in node/device objects only")
        print("   2. Accessed through a different service")
        print("   3. Using a method pattern not yet tested")

if __name__ == "__main__":
    main()
