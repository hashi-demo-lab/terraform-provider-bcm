#!/usr/bin/env python3
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

"""
BCM Category List Fields Persistence Investigation

This script tests whether BCM API persists category list fields after create/update operations.
Tested fields: staticRoutes, fsexports, roles, gpuSettings, services

Evidence output: /workspace/specs/073-category-list-fields/evidence/
"""

import requests
import json
import os
import sys
import argparse
from datetime import datetime
from urllib3.exceptions import InsecureRequestWarning
import uuid
import time

# Suppress SSL warnings for self-signed certs
requests.packages.urllib3.disable_warnings(category=InsecureRequestWarning)

EVIDENCE_DIR = "/workspace/specs/073-category-list-fields/evidence"

class BCMInvestigator:
    """Investigates BCM Category List Fields Persistence"""

    def __init__(self, base_url, username, password):
        self.base_url = base_url.rstrip('/')
        self.username = username
        self.password = password
        self.session = requests.Session()
        self.session.verify = False
        self.results = {
            "test_date": datetime.now().isoformat(),
            "bcm_endpoint": base_url,
            "bcm_version": None,
            "fields_tested": {},
            "alternative_apis": {},
            "summary": {}
        }

    def login(self):
        """Authenticate with BCM API"""
        url = f"{self.base_url}/json"
        payload = {
            "service": "login",
            "username": self.username,
            "password": self.password
        }

        print(f"[INFO] Logging in to {url}...")
        response = self.session.post(url, json=payload, timeout=30)
        response.raise_for_status()

        if 'cm-login-token' in self.session.cookies:
            print("[OK] Login successful")
            return True
        else:
            print("[FAIL] Login failed: No session cookie received")
            return False

    def call_api(self, service, method, *args):
        """Make BCM API call and return response"""
        url = f"{self.base_url}/json"
        payload = {
            "service": service,
            "call": method
        }

        if args:
            payload["args"] = list(args)

        try:
            response = self.session.post(url, json=payload, timeout=60)
            response.raise_for_status()
            return response.json(), None
        except Exception as e:
            return None, str(e)

    def generate_test_name(self):
        """Generate unique test category name"""
        timestamp = datetime.now().strftime("%Y%m%d%H%M%S")
        return f"citest-listfields-{timestamp}"

    def create_category(self, name, fields_to_set):
        """Create a category with specified fields"""
        entity = {
            "baseType": "Category",
            "childType": "",
            "modified": True,
            "to_be_removed": False,
            "revision": "",
            "uuid": "",
            "name": name,
            "notes": f"Test category for list fields investigation - {datetime.now().isoformat()}"
        }
        entity.update(fields_to_set)

        print(f"\n[INFO] Creating category '{name}' with fields: {list(fields_to_set.keys())}")
        result, error = self.call_api("cmdevice", "addCategory", entity)

        if error:
            print(f"[FAIL] Create failed: {error}")
            return None

        if result:
            print(f"[OK] Category created")
            return result
        return None

    def get_category(self, name):
        """Get category by name"""
        result, error = self.call_api("cmdevice", "getCategory", name)
        if error:
            print(f"[FAIL] Get category failed: {error}")
            return None
        return result

    def update_category(self, category_data, fields_to_update):
        """Update a category with new field values"""
        entity = {
            "baseType": "Category",
            "childType": "",
            "modified": True,
            "to_be_removed": False,
            "revision": "",
            "uuid": category_data.get("uuid", ""),
            "name": category_data.get("name", "")
        }

        # Copy existing fields
        for key, value in category_data.items():
            if key not in entity:
                entity[key] = value

        # Apply updates
        entity.update(fields_to_update)

        print(f"\n[INFO] Updating category with fields: {list(fields_to_update.keys())}")
        result, error = self.call_api("cmdevice", "updateCategory", entity, False)

        if error:
            print(f"[FAIL] Update failed: {error}")
            return None

        return result

    def delete_category(self, name):
        """Delete a category by name"""
        category = self.get_category(name)
        if not category:
            return False

        result, error = self.call_api("cmdevice", "removeCategory", category.get("uuid", name))
        if error:
            print(f"[WARN] Delete may have failed: {error}")
            return False
        return True

    def test_static_routes(self, category_name):
        """Test staticRoutes field persistence"""
        test_result = {
            "field": "staticRoutes",
            "terraform_attr": "static_routes",
            "create_value": None,
            "read_value": None,
            "persisted": None,
            "update_value": None,
            "update_read_value": None,
            "update_persisted": None,
            "evidence": {}
        }

        # Test value
        static_routes = [
            {
                "destination": "10.0.0.0/8",
                "gateway": "192.168.1.1",
                "metric": 100
            },
            {
                "destination": "172.16.0.0/12",
                "gateway": "192.168.1.1",
                "metric": 200
            }
        ]
        test_result["create_value"] = static_routes

        # Create category with static routes
        create_result = self.create_category(category_name, {"staticRoutes": static_routes})
        test_result["evidence"]["create_response"] = create_result

        # Read back
        time.sleep(1)  # Allow BCM to process
        read_result = self.get_category(category_name)
        test_result["evidence"]["read_response"] = read_result

        if read_result:
            read_routes = read_result.get("staticRoutes", [])
            test_result["read_value"] = read_routes
            test_result["persisted"] = len(read_routes) > 0 and read_routes == static_routes

            print(f"[INFO] staticRoutes - Sent: {len(static_routes)} routes")
            print(f"[INFO] staticRoutes - Read: {len(read_routes)} routes")
            print(f"[{'OK' if test_result['persisted'] else 'FAIL'}] staticRoutes persisted: {test_result['persisted']}")

            # Test update
            new_routes = [
                {
                    "destination": "192.168.100.0/24",
                    "gateway": "192.168.1.254",
                    "metric": 50
                }
            ]
            test_result["update_value"] = new_routes

            update_result = self.update_category(read_result, {"staticRoutes": new_routes})
            test_result["evidence"]["update_response"] = update_result

            time.sleep(1)
            update_read = self.get_category(category_name)
            test_result["evidence"]["update_read_response"] = update_read

            if update_read:
                update_routes = update_read.get("staticRoutes", [])
                test_result["update_read_value"] = update_routes
                test_result["update_persisted"] = len(update_routes) > 0

        return test_result

    def test_roles(self, category_name):
        """Test roles field persistence"""
        test_result = {
            "field": "roles",
            "terraform_attr": "roles",
            "create_value": None,
            "read_value": None,
            "persisted": None,
            "evidence": {}
        }

        # Test value - simple role structure
        roles = [
            {
                "name": "test-compute-role",
                "childType": "ComputeRole",
                "addServices": False
            }
        ]
        test_result["create_value"] = roles

        # Create category with roles
        create_result = self.create_category(category_name, {"roles": roles})
        test_result["evidence"]["create_response"] = create_result

        # Read back
        time.sleep(1)
        read_result = self.get_category(category_name)
        test_result["evidence"]["read_response"] = read_result

        if read_result:
            read_roles = read_result.get("roles", [])
            test_result["read_value"] = read_roles
            test_result["persisted"] = len(read_roles) > 0

            print(f"[INFO] roles - Sent: {len(roles)} roles")
            print(f"[INFO] roles - Read: {len(read_roles)} roles")
            print(f"[{'OK' if test_result['persisted'] else 'FAIL'}] roles persisted: {test_result['persisted']}")

        return test_result

    def test_fsexports(self, category_name):
        """Test fsexports field persistence"""
        test_result = {
            "field": "fsexports",
            "terraform_attr": "fsexports",
            "create_value": None,
            "read_value": None,
            "persisted": None,
            "evidence": {}
        }

        # Test value
        fsexports = [
            {
                "path": "/shared/data",
                "network": "10.0.0.0/8",
                "allowWrite": True,
                "async": True,
                "rootSquash": False
            }
        ]
        test_result["create_value"] = fsexports

        # Create category with fsexports
        create_result = self.create_category(category_name, {"fsexports": fsexports})
        test_result["evidence"]["create_response"] = create_result

        # Read back
        time.sleep(1)
        read_result = self.get_category(category_name)
        test_result["evidence"]["read_response"] = read_result

        if read_result:
            read_exports = read_result.get("fsexports", [])
            test_result["read_value"] = read_exports
            test_result["persisted"] = len(read_exports) > 0

            print(f"[INFO] fsexports - Sent: {len(fsexports)} exports")
            print(f"[INFO] fsexports - Read: {len(read_exports)} exports")
            print(f"[{'OK' if test_result['persisted'] else 'FAIL'}] fsexports persisted: {test_result['persisted']}")

        return test_result

    def test_gpu_settings(self, category_name):
        """Test gpuSettings field persistence"""
        test_result = {
            "field": "gpuSettings",
            "terraform_attr": "gpu_settings",
            "create_value": None,
            "read_value": None,
            "persisted": None,
            "evidence": {}
        }

        # Test value
        gpu_settings = [
            {
                "deviceId": "0",
                "model": "NVIDIA Tesla V100",
                "computeMode": "default"
            }
        ]
        test_result["create_value"] = gpu_settings

        # Create category with gpuSettings
        create_result = self.create_category(category_name, {"gpuSettings": gpu_settings})
        test_result["evidence"]["create_response"] = create_result

        # Read back
        time.sleep(1)
        read_result = self.get_category(category_name)
        test_result["evidence"]["read_response"] = read_result

        if read_result:
            read_gpu = read_result.get("gpuSettings", [])
            test_result["read_value"] = read_gpu
            test_result["persisted"] = len(read_gpu) > 0

            print(f"[INFO] gpuSettings - Sent: {len(gpu_settings)} settings")
            print(f"[INFO] gpuSettings - Read: {len(read_gpu)} settings")
            print(f"[{'OK' if test_result['persisted'] else 'FAIL'}] gpuSettings persisted: {test_result['persisted']}")

        return test_result

    def test_services(self, category_name):
        """Test services field persistence"""
        test_result = {
            "field": "services",
            "terraform_attr": "services",
            "create_value": None,
            "read_value": None,
            "persisted": None,
            "evidence": {}
        }

        # Test value - services structure may vary
        services = [
            {
                "name": "test-service",
                "enabled": True
            }
        ]
        test_result["create_value"] = services

        # Create category with services
        create_result = self.create_category(category_name, {"services": services})
        test_result["evidence"]["create_response"] = create_result

        # Read back
        time.sleep(1)
        read_result = self.get_category(category_name)
        test_result["evidence"]["read_response"] = read_result

        if read_result:
            read_services = read_result.get("services", [])
            test_result["read_value"] = read_services
            test_result["persisted"] = len(read_services) > 0

            print(f"[INFO] services - Sent: {len(services)} services")
            print(f"[INFO] services - Read: {len(read_services)} services")
            print(f"[{'OK' if test_result['persisted'] else 'FAIL'}] services persisted: {test_result['persisted']}")

        return test_result

    def discover_alternative_apis(self):
        """Probe for alternative APIs to manage category list fields"""
        print("\n" + "="*80)
        print("ALTERNATIVE API DISCOVERY")
        print("="*80)

        methods_to_test = [
            # Category-specific role methods
            ("cmdevice", "addCategoryRole", "Adds role to category"),
            ("cmdevice", "removeCategoryRole", "Removes role from category"),
            ("cmdevice", "setCategoryRoles", "Sets roles on category"),
            ("cmdevice", "getCategoryRoles", "Gets roles for category"),

            # Category-specific static route methods
            ("cmdevice", "addCategoryStaticRoute", "Adds static route to category"),
            ("cmdevice", "removeCategoryStaticRoute", "Removes static route from category"),
            ("cmdevice", "setCategoryStaticRoutes", "Sets static routes on category"),
            ("cmdevice", "getCategoryStaticRoutes", "Gets static routes for category"),

            # Category-specific fsexport methods
            ("cmdevice", "addCategoryFSExport", "Adds FS export to category"),
            ("cmdevice", "removeCategoryFSExport", "Removes FS export from category"),
            ("cmdevice", "setCategoryFSExports", "Sets FS exports on category"),
            ("cmdevice", "getCategoryFSExports", "Gets FS exports for category"),

            # Category-specific GPU methods
            ("cmdevice", "addCategoryGPUSetting", "Adds GPU setting to category"),
            ("cmdevice", "setCategoryGPUSettings", "Sets GPU settings on category"),
            ("cmdevice", "getCategoryGPUSettings", "Gets GPU settings for category"),

            # Category-specific service methods
            ("cmdevice", "addCategoryService", "Adds service to category"),
            ("cmdevice", "setCategoryServices", "Sets services on category"),
            ("cmdevice", "getCategoryServices", "Gets services for category"),

            # Node-level role methods (may be where roles are actually managed)
            ("cmdevice", "addNodeRole", "Adds role to node"),
            ("cmdevice", "removeNodeRole", "Removes role from node"),
            ("cmdevice", "getNodeRoles", "Gets roles for node"),
            ("cmdevice", "getRoles", "Gets all roles"),
            ("cmdevice", "getRole", "Gets single role"),

            # Generic role methods
            ("cmdevice", "listRoles", "Lists available roles"),
            ("cmdevice", "getRoleTypes", "Gets role types"),
            ("cmdevice", "getAvailableRoles", "Gets available roles"),
        ]

        for service, method, description in methods_to_test:
            print(f"\n[TEST] {service}.{method}")
            result, error = self.call_api(service, method)

            status = "ERROR"
            response_summary = None

            if error:
                if "not found" in error.lower() or "unknown" in error.lower():
                    status = "NOT_FOUND"
                else:
                    status = "ERROR"
                    response_summary = error[:200]
            else:
                status = "EXISTS"
                if isinstance(result, list):
                    response_summary = f"Returns list of {len(result)} items"
                elif isinstance(result, dict):
                    response_summary = f"Returns object with keys: {list(result.keys())[:5]}"
                else:
                    response_summary = str(type(result))

            self.results["alternative_apis"][f"{service}.{method}"] = {
                "description": description,
                "status": status,
                "response_summary": response_summary
            }

            print(f"[{status}] {service}.{method} - {response_summary or 'No response'}")

    def run_all_field_tests(self):
        """Run tests for all category list fields"""
        print("\n" + "="*80)
        print("FIELD PERSISTENCE TESTS")
        print("="*80)

        # Test each field individually with unique category names
        test_configs = [
            ("staticRoutes", self.test_static_routes),
            ("roles", self.test_roles),
            ("fsexports", self.test_fsexports),
            ("gpuSettings", self.test_gpu_settings),
            ("services", self.test_services),
        ]

        for field_name, test_func in test_configs:
            print(f"\n{'='*40}")
            print(f"TESTING: {field_name}")
            print(f"{'='*40}")

            category_name = f"{self.generate_test_name()}-{field_name[:6]}"

            try:
                result = test_func(category_name)
                self.results["fields_tested"][field_name] = result
            except Exception as e:
                print(f"[ERROR] Test failed with exception: {e}")
                self.results["fields_tested"][field_name] = {
                    "field": field_name,
                    "error": str(e),
                    "persisted": None
                }
            finally:
                # Cleanup test category
                print(f"[INFO] Cleaning up test category: {category_name}")
                self.delete_category(category_name)

    def generate_summary(self):
        """Generate summary of findings"""
        print("\n" + "="*80)
        print("SUMMARY OF FINDINGS")
        print("="*80)

        persisted_count = 0
        not_persisted_count = 0
        error_count = 0

        summary_table = []

        for field_name, result in self.results["fields_tested"].items():
            if "error" in result:
                status = "ERROR"
                error_count += 1
            elif result.get("persisted"):
                status = "PERSISTED"
                persisted_count += 1
            else:
                status = "NOT_PERSISTED"
                not_persisted_count += 1

            summary_table.append({
                "field": field_name,
                "terraform_attr": result.get("terraform_attr", "N/A"),
                "status": status,
                "sent_count": len(result.get("create_value", [])) if result.get("create_value") else 0,
                "read_count": len(result.get("read_value", [])) if result.get("read_value") else 0
            })

        # Print table
        print(f"\n{'Field':<20} {'Terraform Attr':<20} {'Sent':<8} {'Read':<8} {'Status':<15}")
        print("-"*75)
        for row in summary_table:
            print(f"{row['field']:<20} {row['terraform_attr']:<20} {row['sent_count']:<8} {row['read_count']:<8} {row['status']:<15}")

        print(f"\nPersisted: {persisted_count}, Not Persisted: {not_persisted_count}, Errors: {error_count}")

        # Alternative APIs summary
        working_apis = [k for k, v in self.results["alternative_apis"].items() if v["status"] == "EXISTS"]
        print(f"\nAlternative APIs found: {len(working_apis)}")
        for api in working_apis:
            print(f"  - {api}")

        self.results["summary"] = {
            "fields_persisted": persisted_count,
            "fields_not_persisted": not_persisted_count,
            "fields_error": error_count,
            "alternative_apis_found": len(working_apis),
            "conclusion": "BCM does NOT persist category list fields" if not_persisted_count > 0 else "BCM persists category list fields"
        }

    def save_results(self):
        """Save results to evidence file"""
        os.makedirs(EVIDENCE_DIR, exist_ok=True)

        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")

        # Main results file
        results_file = os.path.join(EVIDENCE_DIR, f"category_list_fields_test_results.json")
        with open(results_file, 'w') as f:
            json.dump(self.results, f, indent=2, default=str)
        print(f"\n[OK] Results saved to: {results_file}")

        # API discovery results
        api_file = os.path.join(EVIDENCE_DIR, "api_discovery_results.md")
        with open(api_file, 'w') as f:
            f.write("# BCM Alternative API Discovery Results\n\n")
            f.write(f"**Test Date**: {self.results['test_date']}\n")
            f.write(f"**BCM Endpoint**: {self.results['bcm_endpoint']}\n\n")

            f.write("## Method Discovery Results\n\n")
            f.write("| Service | Method | Status | Response |\n")
            f.write("|---------|--------|--------|----------|\n")

            for method, info in self.results["alternative_apis"].items():
                service, method_name = method.split(".", 1)
                f.write(f"| {service} | {method_name} | {info['status']} | {info.get('response_summary', 'N/A')} |\n")

            f.write("\n## Conclusion\n\n")
            working = [k for k, v in self.results["alternative_apis"].items() if v["status"] == "EXISTS"]
            if working:
                f.write("The following alternative APIs were found to exist:\n\n")
                for api in working:
                    f.write(f"- `{api}`\n")
            else:
                f.write("No alternative APIs were found for managing category list fields.\n")

        print(f"[OK] API discovery saved to: {api_file}")

        return results_file


def main():
    parser = argparse.ArgumentParser(description="Investigate BCM Category List Fields Persistence")
    parser.add_argument("--discover-apis", action="store_true", help="Only run alternative API discovery")
    parser.add_argument("--field", type=str, help="Test only a specific field (staticRoutes, roles, fsexports, gpuSettings, services)")
    args = parser.parse_args()

    # Get credentials from environment
    endpoint = os.getenv("BCM_ENDPOINT", "https://172.21.15.254:8081")
    username = os.getenv("BCM_USERNAME", "root")
    password = os.getenv("BCM_PASSWORD", "Hashicorp123!")

    print("="*80)
    print("BCM CATEGORY LIST FIELDS PERSISTENCE INVESTIGATION")
    print("="*80)
    print(f"Endpoint: {endpoint}")
    print(f"Username: {username}")
    print(f"Evidence Dir: {EVIDENCE_DIR}")
    print("="*80)

    investigator = BCMInvestigator(endpoint, username, password)

    # Login
    if not investigator.login():
        print("[FAIL] Could not authenticate. Exiting.")
        return 1

    # Run investigations
    if args.discover_apis:
        investigator.discover_alternative_apis()
    else:
        investigator.run_all_field_tests()
        investigator.discover_alternative_apis()

    # Generate summary and save
    investigator.generate_summary()
    results_file = investigator.save_results()

    print("\n" + "="*80)
    print("INVESTIGATION COMPLETE")
    print("="*80)
    print(f"Results saved to: {results_file}")

    return 0


if __name__ == "__main__":
    sys.exit(main())
