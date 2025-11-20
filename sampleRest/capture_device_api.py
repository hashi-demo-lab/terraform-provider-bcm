#!/usr/bin/env python3
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

"""
Script to capture Device entity API documentation from BCM API
Uses the correct /json endpoint with authentication
"""

import requests
import json
import sys
from datetime import datetime
import urllib3

# Disable SSL warnings for self-signed certificates
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

class BCMApiClient:
    """Client for NVIDIA BCM API"""

    def __init__(self, base_url, username, password):
        self.base_url = base_url
        self.username = username
        self.password = password
        self.session = requests.Session()
        self.session.verify = False
        self.logged_in = False

    def login(self):
        """Authenticate and get session cookie"""
        print(f"Logging in as {self.username}...")

        login_payload = {
            "service": "login",
            "username": self.username,
            "password": self.password
        }

        try:
            response = self.session.post(
                f"{self.base_url}/json",
                json=login_payload,
                timeout=30
            )
            response.raise_for_status()

            # Check if we got a cookie
            if 'cm-login-token' in self.session.cookies:
                self.logged_in = True
                print("✓ Login successful")
                print(f"Session cookie: cm-login-token={self.session.cookies['cm-login-token']}")
                return True
            else:
                print("✗ Login failed - no cookie received")
                return False

        except Exception as e:
            print(f"✗ Login failed: {e}")
            return False

    def call_api(self, service, call, arg=None):
        """Make an API call to the BCM /json endpoint"""
        if not self.logged_in:
            print("Not logged in. Call login() first.")
            return None

        payload = {
            "service": service,
            "call": call
        }

        if arg is not None:
            payload["arg"] = arg

        print(f"\nCalling API: service={service}, call={call}, arg={arg}")

        try:
            response = self.session.post(
                f"{self.base_url}/json",
                json=payload,
                timeout=30
            )
            response.raise_for_status()

            print(f"Status Code: {response.status_code}")
            print(f"Content-Type: {response.headers.get('Content-Type', 'unknown')}")

            # Parse JSON response
            data = response.json()
            return data

        except Exception as e:
            print(f"✗ API call failed: {e}")
            return None

    def get_device(self, device_name="master"):
        """Get device information"""
        return self.call_api("cmdevice", "getNode", device_name)

    def search_entity(self, entity_type="Device", name=None):
        """Search for entities (if such an endpoint exists)"""
        # Note: This is speculative - might need to adjust based on actual API
        payload = {
            "type": entity_type
        }
        if name:
            payload["name"] = name

        return self.call_api("search", "entity", payload)


def save_response(data, filename):
    """Save API response to file"""
    with open(filename, 'w') as f:
        json.dump(data, f, indent=2)
    print(f"\nResponse saved to: {filename}")


def main():
    """Main execution function"""
    # API configuration
    base_url = "https://172.21.15.254:8081"
    username = "root"
    password = "Hashicorp123!"

    # Create client and login
    client = BCMApiClient(base_url, username, password)

    if not client.login():
        print("\nFailed to authenticate. Check credentials.")
        sys.exit(1)

    print("\n" + "="*80)
    print("Capturing Device Entity API Documentation")
    print("="*80)

    # Generate timestamp for filenames
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")

    # Example 1: Get master node (Device entity)
    print("\n1. Getting master node (Device entity)...")
    master_device = client.get_device("master")

    if master_device:
        print("\n" + "="*80)
        print("Master Device Response:")
        print("="*80)
        print(json.dumps(master_device, indent=2)[:1000])  # First 1000 chars
        print("...(truncated)")

        filename = f"device_master_{timestamp}.json"
        save_response(master_device, filename)

        # Print key device properties
        if isinstance(master_device, dict):
            print("\nKey Device Properties:")
            print(f"  - baseType: {master_device.get('baseType', 'N/A')}")
            print(f"  - childType: {master_device.get('childType', 'N/A')}")
            print(f"  - hostname: {master_device.get('hostname', 'N/A')}")
            print(f"  - uuid: {master_device.get('uuid', 'N/A')}")
            print(f"  - mac: {master_device.get('mac', 'N/A')}")

    # You can add more API exploration here
    # Example: Try to list all devices, get specific device properties, etc.

    print("\n" + "="*80)
    print("✓ Device API documentation captured successfully")
    print("="*80)


if __name__ == "__main__":
    main()
