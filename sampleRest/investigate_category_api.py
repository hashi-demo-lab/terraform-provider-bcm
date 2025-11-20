#!/usr/bin/env python3
"""
Category API Investigation Script
Explores the CMDevice Category API to document full schema and methods.
"""

import requests
import json
import os
from datetime import datetime
from urllib3.exceptions import InsecureRequestWarning

# Suppress SSL warnings for self-signed certs
requests.packages.urllib3.disable_warnings(category=InsecureRequestWarning)

class BCMCategoryInvestigator:
    """Investigates BCM Category API"""

    def __init__(self, base_url, username, password):
        self.base_url = base_url.rstrip('/')
        self.username = username
        self.password = password
        self.session = requests.Session()
        self.session.verify = False

    def login(self):
        """Authenticate with BCM API"""
        url = f"{self.base_url}/json"
        payload = {
            "service": "login",
            "username": self.username,
            "password": self.password
        }

        print(f"Logging in to {url}...")
        response = self.session.post(url, json=payload, timeout=30)
        response.raise_for_status()

        # Check if we got a session cookie
        if 'cm-login-token' in self.session.cookies:
            print("✅ Login successful")
            return True
        else:
            print(f"❌ Login failed: No session cookie received")
            return False

    def call_api(self, service, method, *args):
        """Make BCM API call"""
        url = f"{self.base_url}/json"
        payload = {
            "service": service,
            "call": method
        }

        if args:
            payload["args"] = args

        print(f"\n📡 Calling {service}.{method}({', '.join(map(str, args))})")
        response = self.session.post(url, json=payload)

        try:
            response.raise_for_status()
            return response.json()
        except Exception as e:
            print(f"❌ Error: {e}")
            print(f"Response: {response.text[:500]}")
            return None

    def investigate_categories(self):
        """Get all categories and analyze structure"""
        print("\n" + "="*80)
        print("INVESTIGATING CATEGORY API")
        print("="*80)

        # Test 1: Get all categories
        print("\n--- Test 1: getCategories() ---")
        categories = self.call_api("cmdevice", "getCategories")

        if not categories:
            print("❌ Failed to retrieve categories")
            return None

        print(f"\n✅ Retrieved {len(categories)} categories")

        # Save full response
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        filename = f"category_full_schema_{timestamp}.json"
        with open(filename, 'w') as f:
            json.dump(categories, f, indent=2)
        print(f"📁 Saved full response to: {filename}")

        # Analyze first category in detail
        if categories:
            print("\n--- Category Schema Analysis ---")
            category = categories[0]
            print(f"\nCategory: {category.get('name', 'N/A')}")
            print(f"UUID: {category.get('uuid', 'N/A')}")
            print(f"\nAll attributes found:")

            # Collect all unique keys across all categories
            all_keys = set()
            for cat in categories:
                all_keys.update(cat.keys())

            # Print sorted keys with types
            for key in sorted(all_keys):
                # Get value from first category that has this key
                value = None
                value_type = "N/A"
                for cat in categories:
                    if key in cat:
                        value = cat[key]
                        value_type = type(value).__name__
                        break

                # Truncate long values
                value_str = str(value)
                if len(value_str) > 100:
                    value_str = value_str[:97] + "..."

                print(f"  {key:30} ({value_type:15}) = {value_str}")

        return categories

    def test_get_category(self, categories):
        """Test getting a single category"""
        if not categories:
            return

        print("\n--- Test 2: getCategory(name) ---")

        # Try getting by name
        category_name = categories[0].get('name')
        result = self.call_api("cmdevice", "getCategory", category_name)

        if result:
            print(f"✅ Successfully retrieved category by name: {category_name}")
        else:
            print(f"❌ Failed to retrieve category by name")

        # Try getting by UUID
        category_uuid = categories[0].get('uuid')
        result = self.call_api("cmdevice", "getCategory", category_uuid)

        if result:
            print(f"✅ Successfully retrieved category by UUID: {category_uuid}")
        else:
            print(f"❌ Failed to retrieve category by UUID")

    def analyze_category_relationships(self, categories):
        """Analyze relationships to other entities"""
        print("\n--- Test 3: Category Relationships ---")

        if not categories:
            return

        category = categories[0]

        # Check for software image reference
        if 'softwareImageProxy' in category:
            proxy = category['softwareImageProxy']
            print(f"\nSoftware Image Proxy:")
            print(f"  Type: {type(proxy).__name__}")
            if isinstance(proxy, dict):
                for key, value in proxy.items():
                    print(f"  {key}: {value}")

        # Check for partition reference
        if 'partition' in category:
            print(f"\nPartition: {category['partition']}")

        # Check for management network
        if 'managementNetwork' in category:
            print(f"\nManagement Network: {category['managementNetwork']}")

        # Check for roles
        if 'roles' in category:
            roles = category['roles']
            print(f"\nRoles: {len(roles)} roles defined")
            if roles:
                print(f"  First role: {roles[0]}")

    def test_category_methods(self):
        """Test various category-related methods"""
        print("\n--- Test 4: Testing Related Methods ---")

        # These are likely methods based on common API patterns
        methods_to_test = [
            ("cmdevice", "getCategories"),
            ("cmdevice", "getCategoriesWithDetails"),  # Maybe?
            ("cmpart", "getSoftwareImages"),  # Related: get images for category
        ]

        for service, method in methods_to_test:
            result = self.call_api(service, method)
            if result:
                print(f"✅ {service}.{method}() works - returned {len(result) if isinstance(result, list) else 'object'}")

    def analyze_disksetup(self, categories):
        """Analyze disk setup configuration"""
        print("\n--- Test 5: Disk Setup Analysis ---")

        if not categories:
            return

        for cat in categories:
            if 'disksetup' in cat and cat['disksetup']:
                disksetup = cat['disksetup']
                print(f"\nCategory: {cat['name']}")
                print(f"Disksetup type: {type(disksetup).__name__}")

                if isinstance(disksetup, str):
                    # Check if it's XML
                    if disksetup.strip().startswith('<?xml'):
                        print(f"  Format: XML content")
                        print(f"  Length: {len(disksetup)} bytes")
                        # Extract root element
                        if '<disksetup' in disksetup:
                            print(f"  Root element: <disksetup>")
                    else:
                        print(f"  Format: Path/String")
                        print(f"  Value: {disksetup}")
                elif disksetup is None:
                    print(f"  Value: null")

                break  # Only show first non-null example

    def generate_schema_documentation(self, categories):
        """Generate comprehensive schema documentation"""
        print("\n--- Generating Schema Documentation ---")

        if not categories:
            return

        # Collect all fields across all categories with their types
        field_analysis = {}

        for cat in categories:
            for key, value in cat.items():
                if key not in field_analysis:
                    field_analysis[key] = {
                        'type': type(value).__name__,
                        'examples': [],
                        'null_count': 0,
                        'non_null_count': 0
                    }

                if value is None:
                    field_analysis[key]['null_count'] += 1
                else:
                    field_analysis[key]['non_null_count'] += 1

                    # Store example values (limit to 3)
                    if len(field_analysis[key]['examples']) < 3:
                        # Truncate long values
                        example = str(value)
                        if len(example) > 50:
                            example = example[:47] + "..."
                        field_analysis[key]['examples'].append(example)

        # Generate markdown documentation
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        doc_filename = f"category_schema_documentation_{timestamp}.md"

        with open(doc_filename, 'w') as f:
            f.write("# BCM Category API Schema Documentation\n\n")
            f.write(f"Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n\n")
            f.write(f"Total categories analyzed: {len(categories)}\n\n")

            f.write("## Category Attributes\n\n")
            f.write("| Attribute | Type | Nullable | Examples |\n")
            f.write("|-----------|------|----------|----------|\n")

            for key in sorted(field_analysis.keys()):
                info = field_analysis[key]
                nullable = "Yes" if info['null_count'] > 0 else "No"
                examples = " | ".join(info['examples'][:2])
                f.write(f"| `{key}` | {info['type']} | {nullable} | {examples} |\n")

            f.write("\n## Full Category Example\n\n")
            f.write("```json\n")
            f.write(json.dumps(categories[0], indent=2))
            f.write("\n```\n")

        print(f"📁 Generated schema documentation: {doc_filename}")

        return doc_filename

def main():
    # Get credentials from environment
    endpoint = os.getenv("BCM_ENDPOINT", "https://172.21.15.254:8081")
    username = os.getenv("BCM_USERNAME", "root")
    password = os.getenv("BCM_PASSWORD", "Hashicorp123!")

    print("BCM Category API Investigation")
    print("="*80)
    print(f"Endpoint: {endpoint}")
    print(f"Username: {username}")
    print("="*80)

    # Create investigator
    investigator = BCMCategoryInvestigator(endpoint, username, password)

    # Login
    if not investigator.login():
        print("Failed to login. Exiting.")
        return 1

    # Run investigation
    categories = investigator.investigate_categories()

    if categories:
        investigator.test_get_category(categories)
        investigator.analyze_category_relationships(categories)
        investigator.test_category_methods()
        investigator.analyze_disksetup(categories)
        investigator.generate_schema_documentation(categories)

    print("\n" + "="*80)
    print("✅ Investigation Complete")
    print("="*80)

    return 0

if __name__ == "__main__":
    exit(main())
