#!/usr/bin/env python3
"""
Terraform Provider Test Modernization Gap Analyzer

Scans test files for legacy patterns and generates a comprehensive gap analysis report.
Detects:
- Legacy Check blocks (resource.TestCheckResourceAttr)
- Missing drift detection tests
- Missing import tests
- Missing idempotency checks
- Modern pattern usage statistics

Usage:
    python analyze_gap.py <test_directory> [--output report.md]
"""

import os
import re
import sys
import argparse
from collections import defaultdict
from pathlib import Path
from typing import Dict, List, Tuple, Set


class ResourceSchema:
    """Represents parsed resource schema."""

    def __init__(self, path: str):
        self.path = path
        self.name = os.path.basename(path).replace('resource_', '').replace('.go', '')
        self.required_fields = []  # List of (field_name, has_validator)
        self.content = ""

    def load(self):
        """Load file content."""
        with open(self.path, 'r') as f:
            self.content = f.read()

    def parse_schema(self):
        """Extract required fields and their validators from schema."""
        # Simple approach: find all field definitions, check for Required: true within next ~15 lines
        lines = self.content.split('\n')

        for i, line in enumerate(lines):
            # Look for field definitions
            field_match = re.match(r'\s*"(\w+)":\s*schema\.\w+Attribute\{', line)
            if not field_match:
                continue

            field_name = field_match.group(1)

            # Check next 15 lines for Required: true and Validators
            has_required = False
            has_validator = False

            for j in range(i, min(i + 15, len(lines))):
                check_line = lines[j]

                # Check for Required: true
                if 'Required:' in check_line and 'true' in check_line:
                    has_required = True

                # Check for validators
                if 'Validators:' in check_line:
                    has_validator = True

                # Stop at the closing brace of this attribute (rough heuristic)
                if j > i and check_line.strip() == '},':
                    break

            # If field is required AND has validators, track it
            if has_required and has_validator:
                self.required_fields.append((field_name, True))


class TestFile:
    """Represents a single test file with analysis results."""

    def __init__(self, path: str):
        self.path = path
        self.name = os.path.basename(path)
        self.content = ""
        self.legacy_checks = []
        self.modern_state_checks = 0
        self.modern_plan_checks = 0
        self.has_import_test = False
        self.has_drift_test = False
        self.idempotency_checks = 0
        self.compare_value_usage = 0
        self.test_functions = []
        self.is_mock_test = '_mock_test.go' in self.name
        self.validation_tests = []  # List of validation test names

    def load(self):
        """Load file content."""
        with open(self.path, 'r') as f:
            self.content = f.read()

    def analyze(self):
        """Analyze file for patterns."""
        self._find_test_functions()
        self._find_legacy_checks()
        self._find_modern_patterns()
        self._find_import_tests()
        self._find_drift_tests()
        self._find_idempotency_checks()
        self._find_validation_tests()

    def _find_test_functions(self):
        """Extract test function names."""
        pattern = r'func\s+(TestAcc\w+)\s*\('
        self.test_functions = re.findall(pattern, self.content)

    def _find_legacy_checks(self):
        """Find legacy Check blocks."""
        # Pattern: resource.TestCheckResourceAttr or resource.TestCheckResourceAttrSet
        legacy_patterns = [
            r'resource\.TestCheckResourceAttr\(',
            r'resource\.TestCheckResourceAttrSet\(',
            r'resource\.TestCheckResourceAttrPair\(',
            r'resource\.TestCheckNoResourceAttr\(',
        ]

        for pattern in legacy_patterns:
            matches = re.finditer(pattern, self.content)
            for match in matches:
                line_num = self.content[:match.start()].count('\n') + 1
                self.legacy_checks.append((line_num, match.group()))

    def _find_modern_patterns(self):
        """Count modern pattern usage."""
        # Count statecheck.ExpectKnownValue
        self.modern_state_checks = len(re.findall(r'statecheck\.ExpectKnownValue\(', self.content))

        # Count plancheck patterns
        self.modern_plan_checks = len(re.findall(r'plancheck\.ExpectEmptyPlan\(', self.content))
        self.modern_plan_checks += len(re.findall(r'plancheck\.ExpectNonEmptyPlan\(', self.content))

        # Count CompareValue usage
        self.compare_value_usage = len(re.findall(r'statecheck\.CompareValue\(', self.content))

    def _find_import_tests(self):
        """Check for import test patterns."""
        self.has_import_test = 'ImportState:' in self.content and 'ImportStateVerify:' in self.content

    def _find_drift_tests(self):
        """Check for drift detection test patterns."""
        # Look for test functions with "Drift" in the name
        drift_pattern = r'func\s+TestAcc\w*Drift\w*\s*\('
        self.has_drift_test = bool(re.search(drift_pattern, self.content))

        # Also check for ExpectNonEmptyPlan usage (common in drift tests)
        if not self.has_drift_test:
            self.has_drift_test = 'ExpectNonEmptyPlan' in self.content

    def _find_idempotency_checks(self):
        """Count idempotency verification patterns."""
        self.idempotency_checks = len(re.findall(r'plancheck\.ExpectEmptyPlan\(', self.content))

    def _find_validation_tests(self):
        """Find validation tests using ExpectError."""
        # Pattern: func TestAcc.*Validation.*
        validation_pattern = r'func\s+(TestAcc\w*(?:Validation|Error|Invalid)\w*)\s*\('
        self.validation_tests = re.findall(validation_pattern, self.content)


class GapAnalyzer:
    """Analyzes test directory for modernization gaps."""

    def __init__(self, test_dir: str):
        self.test_dir = Path(test_dir)
        self.resource_tests: List[TestFile] = []
        self.data_source_tests: List[TestFile] = []
        self.other_tests: List[TestFile] = []
        self.resource_schemas: Dict[str, ResourceSchema] = {}  # resource_name -> schema
        self.validation_coverage: Dict[str, List[str]] = {}  # resource_name -> missing required fields

    def scan(self):
        """Scan directory for test files and resource schemas."""
        if not self.test_dir.exists():
            print(f"Error: Directory {self.test_dir} does not exist")
            sys.exit(1)

        # Scan test files
        for file_path in self.test_dir.glob('*_test.go'):
            test_file = TestFile(str(file_path))
            test_file.load()
            test_file.analyze()

            # Categorize
            if 'resource_' in test_file.name:
                self.resource_tests.append(test_file)
            elif 'data_source_' in test_file.name:
                self.data_source_tests.append(test_file)
            else:
                self.other_tests.append(test_file)

        # Scan resource schema files
        for file_path in self.test_dir.glob('resource_*.go'):
            # Skip test files
            if '_test.go' in str(file_path):
                continue

            schema = ResourceSchema(str(file_path))
            schema.load()
            schema.parse_schema()

            if schema.required_fields:  # Only track if has required fields
                self.resource_schemas[schema.name] = schema

        # Analyze validation coverage
        self._analyze_validation_coverage()

    def _analyze_validation_coverage(self):
        """Cross-reference required fields with validation tests."""
        for schema_name, schema in self.resource_schemas.items():
            # Find matching test files for this resource
            matching_tests = [t for t in self.resource_tests
                            if schema_name in t.name.replace('_test.go', '')]

            if not matching_tests:
                # No test file found for this resource
                self.validation_coverage[schema_name] = [f[0] for f in schema.required_fields]
                continue

            # Collect all validation test names across matching test files
            all_validation_tests = []
            for test_file in matching_tests:
                all_validation_tests.extend(test_file.validation_tests)

            # Check which required fields are NOT covered by validation tests
            missing_validations = []
            for field_name, has_validator in schema.required_fields:
                # Try multiple matching strategies for field name
                # 1. Direct match: "management_network" in test name
                # 2. CamelCase match: "managementnetwork" (remove underscores)
                # 3. Remove all non-alphanumeric: handle any naming convention

                field_lower = field_name.lower()
                field_no_underscore = field_lower.replace('_', '')

                field_tested = any(
                    field_lower in test_name.lower() or
                    field_no_underscore in test_name.lower().replace('_', '')
                    for test_name in all_validation_tests
                )

                # If field has validator but no validation test, mark as missing
                if has_validator and not field_tested:
                    missing_validations.append(field_name)

            if missing_validations:
                self.validation_coverage[schema_name] = missing_validations

    def generate_report(self) -> str:
        """Generate markdown gap analysis report."""
        report = []
        report.append("# Terraform Provider Test Modernization Gap Analysis\n")
        report.append(f"**Analysis Date:** {self._get_date()}\n")
        report.append(f"**Test Directory:** `{self.test_dir}`\n")
        report.append("\n## Executive Summary\n")

        # Calculate statistics
        total_legacy = sum(len(t.legacy_checks) for t in self.resource_tests + self.data_source_tests)
        total_modern_state = sum(t.modern_state_checks for t in self.resource_tests + self.data_source_tests)
        total_modern_plan = sum(t.modern_plan_checks for t in self.resource_tests + self.data_source_tests)

        # Count resources excluding mock tests for import/drift requirements
        real_resources = [t for t in self.resource_tests if not t.is_mock_test]
        mock_resources = [t for t in self.resource_tests if t.is_mock_test]

        resources_with_drift = sum(1 for t in real_resources if t.has_drift_test)
        resources_with_import = sum(1 for t in real_resources if t.has_import_test)

        # Validation coverage statistics
        total_required_fields = sum(len(schema.required_fields) for schema in self.resource_schemas.values())
        missing_validation_count = sum(len(fields) for fields in self.validation_coverage.values())
        covered_validation_count = total_required_fields - missing_validation_count

        report.append(f"- **{total_modern_state}** modern state checks (`statecheck.ExpectKnownValue`)\n")
        report.append(f"- **{total_modern_plan}** modern plan checks (`plancheck.Expect*`)\n")
        report.append(f"- **{total_legacy}** legacy check calls (needs cleanup)\n")
        report.append(f"- **{resources_with_drift}/{len(real_resources)}** acceptance test resources have drift detection tests\n")
        report.append(f"- **{resources_with_import}/{len(real_resources)}** acceptance test resources have import tests\n")
        if total_required_fields > 0:
            report.append(f"- **{covered_validation_count}/{total_required_fields}** required fields have validation tests\n")
        if mock_resources:
            report.append(f"- **{len(mock_resources)}** mock/unit test files (import/drift N/A)\n")

        # Overall grade
        if total_legacy == 0 and resources_with_drift == len(real_resources):
            grade = "A"
        elif total_legacy <= 50 and resources_with_drift >= len(real_resources) * 0.8:
            grade = "B"
        else:
            grade = "C"

        report.append(f"\n**Overall Grade: {grade}**\n")

        # Detailed analysis
        report.append("\n## Resource Tests Analysis\n")
        report.append(self._analyze_category(self.resource_tests, "Resource"))

        report.append("\n## Data Source Tests Analysis\n")
        report.append(self._analyze_category(self.data_source_tests, "Data Source"))

        # Gaps and recommendations
        report.append("\n## Gaps and Recommendations\n")
        report.append(self._generate_recommendations())

        # Pattern reference
        report.append("\n## Modern Testing Patterns Quick Reference\n")
        report.append(self._pattern_reference())

        return ''.join(report)

    def _analyze_category(self, tests: List[TestFile], category: str) -> str:
        """Analyze a category of tests."""
        if not tests:
            return f"No {category.lower()} tests found.\n"

        lines = []
        for test in tests:
            lines.append(f"\n### {test.name}\n")
            lines.append(f"- **Test functions:** {len(test.test_functions)}\n")
            lines.append(f"- **Modern state checks:** {test.modern_state_checks}\n")
            lines.append(f"- **Modern plan checks:** {test.modern_plan_checks}\n")

            if category == "Resource":
                if test.is_mock_test:
                    lines.append(f"- **Test type:** Mock/Unit tests (import/drift N/A)\n")
                else:
                    lines.append(f"- **Has import test:** {'✅' if test.has_import_test else '❌'}\n")
                    lines.append(f"- **Has drift test:** {'✅' if test.has_drift_test else '❌'}\n")
                lines.append(f"- **Idempotency checks:** {test.idempotency_checks}\n")
                lines.append(f"- **Validation tests:** {len(test.validation_tests)}\n")

            if test.legacy_checks:
                lines.append(f"- **Legacy checks:** {len(test.legacy_checks)} ⚠️\n")
                lines.append(f"  - Lines: {', '.join(str(line) for line, _ in test.legacy_checks[:5])}")
                if len(test.legacy_checks) > 5:
                    lines.append(f" (and {len(test.legacy_checks) - 5} more)")
                lines.append("\n")
            else:
                lines.append(f"- **Legacy checks:** None ✅\n")

            # Status
            status = self._file_status(test, category == "Resource")
            lines.append(f"- **Status:** {status}\n")

        return ''.join(lines)

    def _file_status(self, test: TestFile, is_resource: bool) -> str:
        """Determine file modernization status."""
        if test.legacy_checks:
            return f"🟡 Needs cleanup ({len(test.legacy_checks)} legacy checks)"

        # Mock/unit tests don't need import/drift tests
        if is_resource and not test.is_mock_test:
            if not test.has_import_test:
                return "⚠️ Missing import test"
            if not test.has_drift_test:
                return "⚠️ Missing drift detection test"
            if test.idempotency_checks == 0:
                return "⚠️ Missing idempotency checks"

        # Mock tests are OK if they test error paths
        if test.is_mock_test:
            return "✅ Mock/unit tests (error validation)"

        if test.modern_state_checks > 0:
            return "✅ Fully modernized"

        return "❓ Needs review"

    def _generate_recommendations(self) -> str:
        """Generate prioritized recommendations."""
        lines = []

        # High priority
        lines.append("\n### HIGH PRIORITY ⚠️\n")

        # Missing validation tests for required fields
        if self.validation_coverage:
            lines.append("\n**Missing Validation Tests for Required Fields:**\n")
            for resource_name, missing_fields in sorted(self.validation_coverage.items()):
                lines.append(f"- **`{resource_name}`**: {', '.join(missing_fields)}\n")

        # Resources missing drift tests (exclude mock tests)
        missing_drift = [t for t in self.resource_tests if not t.has_drift_test and not t.is_mock_test]
        if missing_drift:
            lines.append("\n**Missing Drift Detection Tests:**\n")
            for test in missing_drift:
                lines.append(f"- `{test.name}`\n")

        # Files with many legacy checks
        legacy_heavy = [(t, len(t.legacy_checks)) for t in self.resource_tests + self.data_source_tests
                        if len(t.legacy_checks) > 10]
        if legacy_heavy:
            legacy_heavy.sort(key=lambda x: x[1], reverse=True)
            lines.append("\n**Files with Heavy Legacy Usage:**\n")
            for test, count in legacy_heavy:
                lines.append(f"- `{test.name}` ({count} legacy checks)\n")

        # Medium priority
        lines.append("\n### MEDIUM PRIORITY 📋\n")

        # Resources missing import tests (exclude mock tests)
        missing_import = [t for t in self.resource_tests if not t.has_import_test and not t.is_mock_test]
        if missing_import:
            lines.append("\n**Missing Import Tests:**\n")
            for test in missing_import:
                lines.append(f"- `{test.name}`\n")

        # Resources with no idempotency checks (exclude mock tests)
        missing_idempotency = [t for t in self.resource_tests if t.idempotency_checks == 0 and not t.is_mock_test]
        if missing_idempotency:
            lines.append("\n**Missing Idempotency Checks:**\n")
            for test in missing_idempotency:
                lines.append(f"- `{test.name}`\n")

        return ''.join(lines)

    def _pattern_reference(self) -> str:
        """Generate pattern reference section."""
        return """
### Required Imports
```go
import (
    "github.com/hashicorp/terraform-plugin-testing/helper/resource"
    "github.com/hashicorp/terraform-plugin-testing/plancheck"
    "github.com/hashicorp/terraform-plugin-testing/statecheck"
    "github.com/hashicorp/terraform-plugin-testing/knownvalue"
    "github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
    "github.com/hashicorp/terraform-plugin-testing/compare"
)
```

### Modern State Check
```go
ConfigStateChecks: []statecheck.StateCheck{
    statecheck.ExpectKnownValue(
        "bcm_resource.test",
        tfjsonpath.New("name"),
        knownvalue.StringExact("expected-value"),
    ),
}
```

### Idempotency Check
```go
ConfigPlanChecks: resource.ConfigPlanChecks{
    PreApply: []plancheck.PlanCheck{
        plancheck.ExpectEmptyPlan(),
    },
}
```

### ID Consistency Tracking
```go
compareID := statecheck.CompareValue(compare.ValuesSame())

ConfigStateChecks: []statecheck.StateCheck{
    compareID.AddStateValue("bcm_resource.test", tfjsonpath.New("id")),
}
```
"""

    def _get_date(self) -> str:
        """Get current date string."""
        from datetime import datetime
        return datetime.now().strftime("%Y-%m-%d")


def main():
    parser = argparse.ArgumentParser(
        description="Analyze Terraform provider tests for modernization gaps"
    )
    parser.add_argument(
        "test_dir",
        help="Directory containing test files (e.g., ./internal/provider/)"
    )
    parser.add_argument(
        "--output", "-o",
        help="Output file path (default: stdout)",
        default=None
    )

    args = parser.parse_args()

    # Run analysis
    analyzer = GapAnalyzer(args.test_dir)
    analyzer.scan()
    report = analyzer.generate_report()

    # Output
    if args.output:
        with open(args.output, 'w') as f:
            f.write(report)
        print(f"✅ Gap analysis report written to: {args.output}")
    else:
        print(report)


if __name__ == "__main__":
    main()
