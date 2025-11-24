#!/usr/bin/env python3
"""
Test if BCM API supports validateSoftwareImage method
"""

import requests
import json
import sys
import os
from urllib3.exceptions import InsecureRequestWarning

# Suppress SSL warnings for self-signed certs
requests.packages.urllib3.disable_warnings(category=InsecureRequestWarning)

# Configuration
BCM_ENDPOINT = os.getenv("BCM_ENDPOINT", "https://172.21.15.254:8081")
BCM_USERNAME = os.getenv("BCM_USERNAME", "root")
BCM_PASSWORD = os.getenv("BCM_PASSWORD", "Hashicorp123!")

def login():
    """Authenticate with BCM API"""
    url = f"{BCM_ENDPOINT}/json"
    payload = {
        "service": "login",
        "username": BCM_USERNAME,
        "password": BCM_PASSWORD
    }

    response = requests.post(url, json=payload, verify=False)

    if response.status_code != 200 or response.json() != True:
        print(f"❌ Login failed: {response.text}")
        sys.exit(1)

    print("✅ Login successful")
    return response.cookies

def get_sample_image(cookies):
    """Get a sample software image to use for validation testing"""
    url = f"{BCM_ENDPOINT}/json"
    payload = {
        "service": "CMPart",
        "call": "getSoftwareImages"
    }

    response = requests.post(url, json=payload, cookies=cookies, verify=False)

    if response.status_code != 200:
        print(f"❌ Failed to get software images: {response.text}")
        return None

    images = response.json()
    if images and len(images) > 0:
        return images[0]

    return None

def test_validate_softwareimage(cookies, image_entity):
    """Test if validateSoftwareImage API method exists"""
    url = f"{BCM_ENDPOINT}/json"

    print("\n🔍 Testing validateSoftwareImage API method...")
    print(f"   Using image: {image_entity.get('name', 'unknown')}")
    print(f"   UUID: {image_entity.get('uuid', 'none')}")

    # Test 1: Validate with full entity
    print("\n📋 Test 1: validateSoftwareImage with full entity")
    payload = {
        "service": "CMPart",
        "call": "validateSoftwareImage",
        "args": [image_entity]
    }

    response = requests.post(url, json=payload, cookies=cookies, verify=False)
    print(f"   Status Code: {response.status_code}")
    print(f"   Response: {json.dumps(response.json(), indent=2)}")

    if response.status_code == 200:
        result = response.json()
        if isinstance(result, dict) and result.get("success") == True:
            print("   ✅ validateSoftwareImage exists and returned success")
            return True
        elif isinstance(result, dict) and "error" in result:
            print(f"   ⚠️  validateSoftwareImage exists but returned error: {result.get('error')}")
            return True
        elif isinstance(result, bool):
            print(f"   ✅ validateSoftwareImage exists and returned: {result}")
            return True
        else:
            print(f"   ⚠️  validateSoftwareImage returned unexpected format")
            return True
    else:
        print(f"   ❌ validateSoftwareImage may not exist or returned error")
        return False

def test_validate_modified_entity(cookies, image_entity):
    """Test validation with a modified entity (should trigger validation errors)"""
    url = f"{BCM_ENDPOINT}/json"

    print("\n📋 Test 2: validateSoftwareImage with invalid modifications")

    # Create a modified copy with invalid data
    modified_entity = image_entity.copy()
    modified_entity["kernelParameters"] = "invalid parameter @@@ test"
    modified_entity["modified"] = True

    payload = {
        "service": "CMPart",
        "call": "validateSoftwareImage",
        "args": [modified_entity]
    }

    response = requests.post(url, json=payload, cookies=cookies, verify=False)
    print(f"   Status Code: {response.status_code}")

    try:
        result = response.json()
        print(f"   Response: {json.dumps(result, indent=2)}")

        if isinstance(result, dict):
            if result.get("success") == False and "validation" in result:
                print("   ✅ Validation correctly caught invalid modifications")
                print(f"   Validation errors: {result['validation']}")
                return True
            elif result.get("success") == True:
                print("   ⚠️  Validation passed (may not detect this specific issue)")
                return True

    except Exception as e:
        print(f"   ⚠️  Response parsing error: {e}")
        print(f"   Raw response: {response.text}")

    return False

def main():
    print("=" * 70)
    print("BCM API validateSoftwareImage Test")
    print("=" * 70)

    # Login
    cookies = login()

    # Get a sample image
    print("\n📦 Fetching sample software image...")
    sample_image = get_sample_image(cookies)

    if not sample_image:
        print("❌ No software images found in system")
        sys.exit(1)

    print(f"   Found: {sample_image.get('name', 'unknown')}")

    # Test validateSoftwareImage
    test1_result = test_validate_softwareimage(cookies, sample_image)
    test2_result = test_validate_modified_entity(cookies, sample_image)

    # Summary
    print("\n" + "=" * 70)
    print("SUMMARY")
    print("=" * 70)

    if test1_result:
        print("✅ validateSoftwareImage API method EXISTS")
        print("   Can be used for pre-flight validation")
    else:
        print("❌ validateSoftwareImage API method NOT FOUND or not working")
        print("   Must rely on server-side validation during add/update operations")

    print("\n💡 Recommendation:")
    if test1_result:
        print("   Implement pre-flight validation using validateSoftwareImage")
        print("   This will provide better error messages before actual CRUD operations")
    else:
        print("   Continue relying on server-side validation in add/update calls")

if __name__ == "__main__":
    main()
