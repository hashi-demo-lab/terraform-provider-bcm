#!/usr/bin/env python3
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

"""
BCM Kubernetes Cluster API - etcd_nodes Research
Purpose: Validate etcdNodes field support and test all optional attributes
GitHub Issue: #25 - Add etcd_nodes and validate attributes
"""

import json
import os
import requests
import sys
import time
import uuid
from urllib3.exceptions import InsecureRequestWarning

# Suppress SSL warnings for self-signed certs
requests.packages.urllib3.disable_warnings(category=InsecureRequestWarning)

# Configuration from environment variables
BCM_ENDPOINT = os.getenv("BCM_ENDPOINT", "https://172.21.15.254:8081")
BCM_USERNAME = os.getenv("BCM_USERNAME", "root")
BCM_PASSWORD = os.getenv("BCM_PASSWORD", "Hashicorp123!")

def authenticate(session):
    """Authenticate with BCM and get session cookie"""
    login_payload = {
        "service": "login",
        "username": BCM_USERNAME,
        "password": BCM_PASSWORD
    }

    print(f"Authenticating to {BCM_ENDPOINT}...")
    response = session.post(
        f"{BCM_ENDPOINT}/json",
        json=login_payload,
        verify=False
    )
    response.raise_for_status()

    data = response.json()
    # BCM login returns True on success
    if data is True:
        pass  # Success
    elif isinstance(data, dict) and not data.get("success", False):
        raise Exception(f"Authentication failed: {data.get('error', 'Unknown error')}")
    elif data is False:
        raise Exception("Authentication failed: Login returned false")

    print("Authentication successful\n")
    return session

def get_available_nodes(session):
    """Get available nodes for cluster creation"""
    payload = {
        "service": "cmdevice",
        "call": "getNodes"
    }

    print("Querying available nodes...")
    response = session.post(
        f"{BCM_ENDPOINT}/json",
        json=payload,
        verify=False
    )
    response.raise_for_status()

    nodes = response.json()
    print(f"Found {len(nodes)} node(s)")

    for i, node in enumerate(nodes[:5]):  # Show first 5 nodes
        print(f"  [{i}] {node.get('name', 'unknown')}: {node.get('uuid', 'no-uuid')}")

    if len(nodes) < 1:
        raise Exception("Need at least 1 node for test cluster")

    return nodes

def test_etcd_nodes_support(session, master_node_uuid, etcd_node_uuids):
    """Test if BCM API supports etcdNodes field"""

    cluster_name = f"etcd-test-{int(time.time())}"
    cluster_uuid = str(uuid.uuid4())

    # Test cluster entity WITH etcdNodes field
    cluster_entity = {
        "baseType": "KubeCluster",
        "childType": "",
        "modified": True,
        "to_be_removed": False,
        "revision": "",
        "uuid": cluster_uuid,  # BCM cmkube requires client-generated UUID
        "name": cluster_name,
        "masterNodes": [master_node_uuid],
        "etcdNodes": etcd_node_uuids  # TEST: Does BCM accept this field?
    }

    payload = {
        "service": "cmkube",
        "call": "addKubeCluster",
        "args": [cluster_entity, False]
    }

    print("\n" + "=" * 80)
    print("RQ-001: Testing etcdNodes field support in addKubeCluster")
    print("=" * 80)
    print("Request entity (with etcdNodes):")
    print(json.dumps(cluster_entity, indent=2))

    try:
        response = session.post(
            f"{BCM_ENDPOINT}/json",
            json=payload,
            verify=False
        )
        response.raise_for_status()
        result = response.json()

        print("\nResponse:")
        print(json.dumps(result, indent=2))

        # Check if operation succeeded
        if isinstance(result, str):
            # BCM returned UUID string directly
            print(f"\n SUCCESS: Cluster created with UUID: {result}")
            print("   etcdNodes field was ACCEPTED by BCM API")

            # Now read back to see if etcdNodes is returned
            time.sleep(2)  # Wait for consistency
            read_cluster_with_etcd_check(session, cluster_uuid)

            # Cleanup
            delete_cluster(session, cluster_uuid)
            return True
        elif isinstance(result, dict):
            if result.get("success") is True:
                print(f"\n SUCCESS: Cluster created with UUID: {cluster_uuid}")
                print("   etcdNodes field was ACCEPTED by BCM API")
                time.sleep(2)
                read_cluster_with_etcd_check(session, cluster_uuid)
                delete_cluster(session, cluster_uuid)
                return True
            elif result.get("error"):
                print(f"\n FAILED: {result.get('error')}")
                print("   etcdNodes field may NOT be supported")
                return False
            elif result.get("success") is False:
                validation = result.get("validation", [])
                print(f"\n FAILED: success=false")
                for v in validation:
                    print(f"   - {v.get('field')}: {v.get('message')}")
                return False
            elif result.get("uuid"):
                returned_uuid = result["uuid"]
                print(f"\n SUCCESS: Cluster created with UUID: {returned_uuid}")
                time.sleep(2)
                read_cluster_with_etcd_check(session, returned_uuid)
                delete_cluster(session, returned_uuid)
                return True

        print(f"\n UNKNOWN RESULT: {result}")
        return False

    except Exception as e:
        print(f"\n ERROR: {e}")
        return False

def read_cluster_with_etcd_check(session, cluster_uuid):
    """Read cluster and check if etcdNodes is returned"""
    payload = {
        "service": "cmkube",
        "call": "getKubeCluster",
        "args": [cluster_uuid]
    }

    print("\n" + "=" * 80)
    print("RQ-003: Checking if getKubeCluster returns etcdNodes")
    print("=" * 80)

    try:
        response = session.post(
            f"{BCM_ENDPOINT}/json",
            json=payload,
            verify=False
        )
        response.raise_for_status()
        result = response.json()

        print("Response:")
        print(json.dumps(result, indent=2))

        if isinstance(result, dict):
            print(f"\nFields returned by getKubeCluster:")
            for key in sorted(result.keys()):
                value = result[key]
                print(f"  {key}: {type(value).__name__}")

            # Check specifically for etcdNodes
            if "etcdNodes" in result:
                print(f"\n etcdNodes IS returned: {result['etcdNodes']}")
            else:
                print(f"\n etcdNodes NOT returned in response")
                print("   (May need to check alternate field names)")

            # Check for alternate field names
            for key in ["etcd_nodes", "EtcdNodes", "etcd", "etcdMembers"]:
                if key in result:
                    print(f"   Found alternate field: {key} = {result[key]}")

    except Exception as e:
        print(f"Error: {e}")

def test_optional_attributes(session, master_node_uuid):
    """Test all optional attributes to verify BCM API support"""

    test_results = {}

    # Attributes to test
    attributes_to_test = [
        ("version", "1.28.0", "version"),
        ("cniPlugin", "calico", "cni_plugin"),
        ("dnsServers", ["8.8.8.8", "8.8.4.4"], "dns_servers"),
        ("overlayNetwork", "test-overlay-uuid", "overlay_network"),
        ("loadBalancerMode", "metallb", "load_balancer_mode"),
        ("storageClasses", json.dumps([{"name": "fast-ssd", "provisioner": "kubernetes.io/csi-driver"}]), "storage_classes"),
        ("addons", json.dumps([{"name": "prometheus", "enabled": True}]), "addons"),
        ("ingressController", json.dumps({"type": "nginx", "enabled": True}), "ingress_controller"),
    ]

    print("\n" + "=" * 80)
    print("RQ-006 to RQ-013: Testing optional attribute support")
    print("=" * 80)

    for bcm_field, test_value, tf_field in attributes_to_test:
        cluster_name = f"attr-test-{bcm_field}-{int(time.time())}"
        cluster_uuid = str(uuid.uuid4())

        cluster_entity = {
            "baseType": "KubeCluster",
            "childType": "",
            "modified": True,
            "to_be_removed": False,
            "revision": "",
            "uuid": cluster_uuid,  # BCM cmkube requires client-generated UUID
            "name": cluster_name,
            "masterNodes": [master_node_uuid],
            bcm_field: test_value
        }

        payload = {
            "service": "cmkube",
            "call": "addKubeCluster",
            "args": [cluster_entity, False]
        }

        print(f"\nTesting {bcm_field} = {test_value}")

        try:
            response = session.post(
                f"{BCM_ENDPOINT}/json",
                json=payload,
                verify=False
            )
            result = response.json()

            created = False
            if isinstance(result, str):
                created = True
            elif isinstance(result, dict):
                if result.get("success") is True:
                    created = True
                elif result.get("uuid"):
                    created = True

            if created:
                print(f"   ACCEPTED: Cluster created")

                # Check if attribute is returned in read
                time.sleep(1)
                read_payload = {
                    "service": "cmkube",
                    "call": "getKubeCluster",
                    "args": [cluster_uuid]
                }
                read_response = session.post(
                    f"{BCM_ENDPOINT}/json",
                    json=read_payload,
                    verify=False
                )
                read_data = read_response.json()

                if isinstance(read_data, dict):
                    if bcm_field in read_data:
                        returned_value = read_data[bcm_field]
                        print(f"   RETURNED: {bcm_field} = {returned_value}")
                        test_results[tf_field] = "FULL_SUPPORT"
                    else:
                        print(f"   NOT RETURNED: {bcm_field} not in response (write-only?)")
                        test_results[tf_field] = "WRITE_ONLY"

                # Cleanup
                delete_cluster(session, cluster_uuid)
            elif isinstance(result, dict) and result.get("success") is False:
                validation = result.get("validation", [])
                errors = [v.get("message", "Unknown") for v in validation if v.get("severity") == "ERROR"]
                if errors:
                    print(f"   VALIDATION ERRORS: {errors}")
                    # Check if error is specific to the field we're testing or just UUID validation
                    if any(bcm_field.lower() in e.lower() for e in errors):
                        test_results[tf_field] = "NOT_SUPPORTED"
                    else:
                        test_results[tf_field] = "UNKNOWN_VALIDATION_ERROR"
                else:
                    test_results[tf_field] = "UNKNOWN"
            elif isinstance(result, dict) and result.get("error"):
                print(f"   REJECTED: {result.get('error')}")
                test_results[tf_field] = "NOT_SUPPORTED"
            else:
                print(f"   UNKNOWN: {type(result)}")
                test_results[tf_field] = "UNKNOWN"

        except Exception as e:
            print(f"   ERROR: {e}")
            test_results[tf_field] = "ERROR"

    # Summary
    print("\n" + "=" * 80)
    print("Attribute Support Summary")
    print("=" * 80)
    for tf_field, status in test_results.items():
        emoji = {"FULL_SUPPORT": "", "WRITE_ONLY": "", "NOT_SUPPORTED": "", "UNKNOWN": "", "ERROR": ""}
        print(f"  {emoji.get(status, '')} {tf_field}: {status}")

    return test_results

def delete_cluster(session, cluster_uuid):
    """Delete test cluster"""
    payload = {
        "service": "cmkube",
        "call": "removeKubeCluster",
        "args": [cluster_uuid, True]  # force=True
    }

    try:
        response = session.post(
            f"{BCM_ENDPOINT}/json",
            json=payload,
            verify=False
        )
        print(f"   Deleted cluster {cluster_uuid[:8]}...")
    except Exception as e:
        print(f"   Cleanup warning: {e}")

def main():
    """Main execution"""
    print("=" * 80)
    print("BCM Kubernetes API - etcd_nodes Research (Issue #25)")
    print("=" * 80)

    session = requests.Session()

    try:
        # Authenticate
        authenticate(session)

        # Get available nodes
        nodes = get_available_nodes(session)

        if len(nodes) < 4:
            print(f"\nWARNING: Only {len(nodes)} nodes available.")
            print("Full etcd test requires 4+ nodes (1 master + 3 etcd)")

        master_node_uuid = nodes[0]['uuid']

        # Use different nodes for etcd if available, otherwise use master node
        if len(nodes) >= 4:
            etcd_node_uuids = [nodes[1]['uuid'], nodes[2]['uuid'], nodes[3]['uuid']]
        elif len(nodes) >= 2:
            etcd_node_uuids = [nodes[1]['uuid']]
        else:
            etcd_node_uuids = [nodes[0]['uuid']]

        print(f"\nUsing master node: {master_node_uuid}")
        print(f"Using etcd nodes: {etcd_node_uuids}")

        # Test etcdNodes field support
        etcd_supported = test_etcd_nodes_support(session, master_node_uuid, etcd_node_uuids)

        # Test all optional attributes
        attribute_results = test_optional_attributes(session, master_node_uuid)

        # Final summary
        print("\n" + "=" * 80)
        print("RESEARCH SUMMARY")
        print("=" * 80)
        print(f"RQ-001 (etcdNodes support): {'YES' if etcd_supported else 'NEEDS VERIFICATION'}")
        print("\nAttribute Support:")
        for attr, status in attribute_results.items():
            print(f"  - {attr}: {status}")

        # Save results
        output_file = "/workspace/sampleRest/cmkube-etcd-test-output.json"
        with open(output_file, 'w') as f:
            json.dump({
                "etcd_nodes_supported": etcd_supported,
                "attribute_results": attribute_results,
                "research_complete": True
            }, f, indent=2)
        print(f"\nResults saved to: {output_file}")

        return 0

    except Exception as e:
        print(f"\nError: {e}", file=sys.stderr)
        import traceback
        traceback.print_exc()
        return 1

if __name__ == "__main__":
    sys.exit(main())
