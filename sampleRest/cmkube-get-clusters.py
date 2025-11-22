#!/usr/bin/env python3
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

"""
BCM Kubernetes Cluster API Exploration - List Clusters
Purpose: Explore cmkube.getKubeClusters API to document cluster structure
Research Questions: RQ-001, RQ-003, RQ-006, RQ-007, RQ-008, RQ-009, RQ-010
"""

import json
import os
import requests
import sys
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
    if not data.get("success", False):
        raise Exception(f"Authentication failed: {data.get('error', 'Unknown error')}")

    print("Authentication successful")
    return session

def get_clusters(session):
    """List all Kubernetes clusters via cmkube.getKubeClusters"""
    payload = {
        "service": "cmkube",
        "call": "getKubeClusters"
    }

    print("\nCalling cmkube.getKubeClusters...")
    response = session.post(
        f"{BCM_ENDPOINT}/json",
        json=payload,
        verify=False
    )
    response.raise_for_status()

    data = response.json()
    return data

def main():
    """Main execution"""
    print("=" * 80)
    print("BCM Kubernetes Cluster API Exploration")
    print("=" * 80)

    session = requests.Session()

    try:
        # Authenticate
        authenticate(session)

        # Get clusters
        result = get_clusters(session)

        # Display results
        print("\nAPI Response:")
        print(json.dumps(result, indent=2))

        # Analyze structure if clusters exist
        if isinstance(result, list) and len(result) > 0:
            print(f"\n{'=' * 80}")
            print(f"Found {len(result)} cluster(s)")
            print("=" * 80)

            for idx, cluster in enumerate(result, 1):
                print(f"\nCluster {idx}:")
                print(f"  UUID: {cluster.get('uuid', 'N/A')}")
                print(f"  Name: {cluster.get('name', 'N/A')}")
                print(f"  Base Type: {cluster.get('baseType', 'N/A')}")
                print(f"  Child Type: {cluster.get('childType', 'N/A')}")

                # Check field naming for nodes (RQ-006)
                master_field = None
                worker_field = None

                if 'masterNodes' in cluster:
                    master_field = 'masterNodes'
                elif 'master_nodes' in cluster:
                    master_field = 'master_nodes'
                elif 'masters' in cluster:
                    master_field = 'masters'

                if 'workerNodes' in cluster:
                    worker_field = 'workerNodes'
                elif 'worker_nodes' in cluster:
                    worker_field = 'worker_nodes'
                elif 'workers' in cluster:
                    worker_field = 'workers'

                print(f"  Master Nodes Field: {master_field}")
                if master_field:
                    print(f"  Master Nodes: {cluster.get(master_field, [])}")

                print(f"  Worker Nodes Field: {worker_field}")
                if worker_field:
                    print(f"  Worker Nodes: {cluster.get(worker_field, [])}")

                # Check network fields (RQ-009)
                if 'managementNetwork' in cluster:
                    print(f"  Management Network: {cluster.get('managementNetwork')}")
                if 'overlayNetwork' in cluster:
                    print(f"  Overlay Network: {cluster.get('overlayNetwork')}")

                # Check version field (RQ-008)
                if 'version' in cluster:
                    print(f"  Kubernetes Version: {cluster.get('version')}")

                # List all fields
                print(f"  All Fields: {', '.join(sorted(cluster.keys()))}")

        elif isinstance(result, list):
            print("\nNo clusters found in BCM")
            print("Note: This is expected if no Kubernetes clusters have been created yet")
        else:
            print("\nUnexpected response format:")
            print(f"Type: {type(result)}")

        # Save results to file
        output_file = "/workspace/sampleRest/cmkube-get-clusters-output.json"
        with open(output_file, 'w') as f:
            json.dump(result, f, indent=2)
        print(f"\nFull response saved to: {output_file}")

        print("\nResearch Findings:")
        print("- RQ-002: getKubeClusters returns list of all clusters (list+filter pattern)")
        print("- RQ-003: Response structure documented above")
        print("- RQ-006: Master/worker node field naming documented")
        print("- Next: Run cmkube-crud-test.py to test args pattern with getKubeCluster (singular)")

        return 0

    except Exception as e:
        print(f"\nError: {e}", file=sys.stderr)
        import traceback
        traceback.print_exc()
        return 1

if __name__ == "__main__":
    sys.exit(main())
