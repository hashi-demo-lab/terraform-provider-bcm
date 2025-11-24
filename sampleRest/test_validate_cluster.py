#!/usr/bin/env python3
"""
Test if BCM API supports Kubernetes cluster validation methods
"""

import requests
import json
import os
from urllib3.exceptions import InsecureRequestWarning

requests.packages.urllib3.disable_warnings(category=InsecureRequestWarning)

BCM_ENDPOINT = os.getenv("BCM_ENDPOINT", "https://172.21.15.254:8081")
BCM_USERNAME = os.getenv("BCM_USERNAME", "root")
BCM_PASSWORD = os.getenv("BCM_PASSWORD", "Hashicorp123!")

def login():
    url = f"{BCM_ENDPOINT}/json"
    payload = {"service": "login", "username": BCM_USERNAME, "password": BCM_PASSWORD}
    response = requests.post(url, json=payload, verify=False)
    if response.status_code != 200 or response.json() != True:
        print(f"❌ Login failed")
        exit(1)
    print("✅ Login successful")
    return response.cookies

def call_api(cookies, service, call, *args):
    url = f"{BCM_ENDPOINT}/json"
    payload = {"service": service, "call": call}
    if args:
        payload["args"] = list(args)
    response = requests.post(url, json=payload, cookies=cookies, verify=False)
    return response.status_code, response.json()

def main():
    print("=" * 70)
    print("BCM API Kubernetes Cluster Validation Method Discovery")
    print("=" * 70)

    cookies = login()

    # Get clusters
    print("\n📦 Fetching Kubernetes clusters...")
    status, clusters = call_api(cookies, "cmkube", "getKubeClusters")

    if status != 200:
        print(f"❌ Failed to get clusters: {status}")
        exit(1)

    if not clusters or len(clusters) == 0:
        print("⚠️  No Kubernetes clusters found in system")
        print("   Will test validation method availability anyway\n")
        sample_cluster = None
    else:
        sample_cluster = clusters[0]
        print(f"   Found cluster: {sample_cluster.get('name', 'unknown')}")
        print(f"   UUID: {sample_cluster.get('uuid', 'none')}")

    # Test validation methods
    validation_methods = [
        "validateCluster",
        "validateKubeCluster",
        "validateKubernetesCluster",
    ]

    print("\n" + "=" * 70)
    print("TESTING CLUSTER VALIDATION METHODS")
    print("=" * 70)

    found_methods = []

    for method in validation_methods:
        print(f"\n🔍 Testing CMKube.{method}")

        if sample_cluster:
            status, result = call_api(cookies, "cmkube", method, sample_cluster)
        else:
            # Test without entity to see if method exists
            status, result = call_api(cookies, "cmkube", method)

        print(f"   Status: {status}")

        if status == 200:
            print(f"   ✅ METHOD EXISTS")
            found_methods.append(method)

            if isinstance(result, list):
                if len(result) == 0:
                    print(f"   Result: [] (validation passed)")
                else:
                    print(f"   Result: {len(result)} validation messages")
                    if len(result) > 0:
                        print(f"   First message: {json.dumps(result[0], indent=2)[:300]}")
            elif isinstance(result, dict):
                print(f"   Result: {json.dumps(result, indent=2)[:200]}")
            else:
                print(f"   Result type: {type(result)}")

        elif status == 400:
            result_json = result if isinstance(result, dict) else {}
            error_msg = result_json.get("errormessage", str(result))

            if "does not exist" in error_msg.lower():
                print(f"   ❌ Method does not exist")
            else:
                print(f"   ⚠️  Status 400: {error_msg[:100]}")
        else:
            print(f"   ❌ Unexpected status: {result}")

    # Test with invalid modifications (if we have a cluster)
    if found_methods and sample_cluster:
        method = found_methods[0]
        print("\n" + "=" * 70)
        print(f"TEST: {method} with invalid modifications")
        print("=" * 70)

        modified_cluster = sample_cluster.copy()
        modified_cluster["modified"] = True
        modified_cluster["name"] = ""  # Empty name (likely invalid)

        status, result = call_api(cookies, "cmkube", method, modified_cluster)
        print(f"Status: {status}")
        print(f"Result: {json.dumps(result, indent=2)[:500]}")

    # Summary
    print("\n" + "=" * 70)
    print("SUMMARY")
    print("=" * 70)

    if found_methods:
        print(f"\n✅ Found {len(found_methods)} validation method(s):")
        for method in found_methods:
            print(f"   - CMKube.{method}")

        print("\n💡 Recommendation:")
        print(f"   Use CMKube.{found_methods[0]} for pre-flight validation")
        print("   in resource_cmkube_cluster.go Update() operation")
    else:
        print("\n❌ No cluster validation methods found")
        print("   Continue relying on server-side validation during update")

if __name__ == "__main__":
    main()
