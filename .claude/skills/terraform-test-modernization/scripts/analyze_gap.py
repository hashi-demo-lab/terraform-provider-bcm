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
        self.optional_fields = []  # List of optional field names
        self.content = ""

    def load(self):
        """Load file content."""
        with open(self.path, 'r') as f:
            self.content = f.read()

    def parse_schema(self):
        """Extract required fields and their validators from schema."""
        # Simple approach: find all field definitions, check for Required: true within next ~15 lines
        # Skip nested fields (those inside NestedObject/NestedAttribute)
        lines = self.content.split('\n')

        in_nested_object = False
        nested_depth = 0

        for i, line in enumerate(lines):
            # Track when we enter/exit nested objects
            if 'NestedObject:' in line or 'NestedAttributeObject' in line:
                in_nested_object = True
                nested_depth = 0

            # Count braces to track nesting depth
            if in_nested_object:
                nested_depth += line.count('{') - line.count('}')
                # Exit nested object when depth returns to 0
                if nested_depth <= 0:
                    in_nested_object = False
                    continue

            # Skip fields inside nested objects
            if in_nested_object:
                continue

            # Look for field definitions at top level only
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

            # Check if field is optional
            has_optional = False
            for j in range(i, min(i + 15, len(lines))):
                if 'Optional:' in lines[j] and 'true' in lines[j]:
                    has_optional = True
                    break

            # If field is required AND has validators, track it
            if has_required and has_validator:
                self.required_fields.append((field_name, True))
            # Track optional fields (for coverage analysis)
            elif has_optional and not has_required:
                self.optional_fields.append(field_name)


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

        # CRUD operation coverage
        self.has_create_test = False
        self.has_update_test = False
        self.has_delete_test = False

        # Test quality metrics
        self.uses_unique_names = False
        self.uses_env_vars = False
        self.has_precheck = False
        self.has_check_destroy = False
        self.quality_score = 0

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
        self._analyze_crud_coverage()
        self._analyze_quality_metrics()
        self._calculate_quality_score()

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

    def _analyze_crud_coverage(self):
        """Detect CRUD operation test coverage."""
        # Create: Tests typically have initial Config step
        self.has_create_test = bool(re.search(r'Steps:\s*\[\]resource\.TestStep\{', self.content))

        # Update: Tests with multiple Config steps with different values
        config_steps = len(re.findall(r'Config:\s+\w+', self.content))
        self.has_update_test = config_steps >= 2

        # Delete: Tests with CheckDestroy
        self.has_delete_test = 'CheckDestroy:' in self.content

    def _analyze_quality_metrics(self):
        """Analyze test quality indicators."""
        # Uses generateUniqueTestName for unique resource names
        self.uses_unique_names = 'generateUniqueTestName' in self.content

        # Uses environment variables instead of hardcoded values
        self.uses_env_vars = 'os.Getenv' in self.content

        # Has PreCheck function
        self.has_precheck = 'PreCheck:' in self.content

        # Has CheckDestroy for cleanup verification
        self.has_check_destroy = 'CheckDestroy:' in self.content

    def _calculate_quality_score(self):
        """Calculate overall quality score (0-100)."""
        score = 0

        # Modern patterns (40 points)
        if self.modern_state_checks > 0:
            score += 20
        if self.modern_plan_checks > 0:
            score += 10
        if len(self.legacy_checks) == 0:
            score += 10

        # CRUD coverage (20 points)
        if self.has_create_test:
            score += 5
        if self.has_update_test:
            score += 5
        if self.has_delete_test:
            score += 10

        # Test completeness (20 points)
        if self.has_import_test:
            score += 10
        if self.has_drift_test or self.is_mock_test:
            score += 10

        # Quality practices (20 points)
        if self.uses_unique_names:
            score += 5
        if self.uses_env_vars:
            score += 5
        if self.has_precheck:
            score += 5
        if self.has_check_destroy:
            score += 5

        self.quality_score = score


class GapAnalyzer:
    """Analyzes test directory for modernization gaps."""

    def __init__(self, test_dir: str):
        self.test_dir = Path(test_dir)
        self.resource_tests: List[TestFile] = []
        self.data_source_tests: List[TestFile] = []
        self.other_tests: List[TestFile] = []
        self.resource_schemas: Dict[str, ResourceSchema] = {}  # resource_name -> schema
        self.validation_coverage: Dict[str, List[str]] = {}  # resource_name -> missing required fields
        self.optional_field_coverage: Dict[str, List[str]] = {}  # resource_name -> untested optional fields

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
        self._analyze_optional_field_coverage()

    def _analyze_optional_field_coverage(self):
        """Check which optional fields are never tested."""
        for schema_name, schema in self.resource_schemas.items():
            if not schema.optional_fields:
                continue

            # Find matching test files
            matching_tests = [t for t in self.resource_tests
                            if schema_name in t.name.replace('_test.go', '')]

            if not matching_tests:
                # No tests for this resource
                self.optional_field_coverage[schema_name] = schema.optional_fields
                continue

            # Collect all test content
            all_test_content = '\n'.join(t.content for t in matching_tests)

            # Check which optional fields never appear in any test config
            # Improved detection: look for field names in test configs (more precise than global search)
            untested_fields = []
            for field_name in schema.optional_fields:
                # Look for field in test content
                # Strategy: Check if field appears in typical config patterns
                # 1. Direct TF config: "field_name = "
                # 2. JSON encode patterns: field_name:
                # 3. State check paths: New("field_name")

                field_lower = field_name.lower()
                field_no_underscore = field_lower.replace('_', '')
                test_content_lower = all_test_content.lower()
                test_content_no_underscore = test_content_lower.replace('_', '')

                # Check multiple patterns
                config_pattern_match = (
                    f'{field_lower} =' in test_content_lower or
                    f'{field_lower}:' in test_content_lower or
                    f'"{field_lower}"' in test_content_lower or
                    f'({field_lower})' in test_content_lower
                )

                # Also check without underscores (camelCase variants)
                no_underscore_match = (
                    field_no_underscore in test_content_no_underscore
                )

                field_tested = config_pattern_match or no_underscore_match

                if not field_tested:
                    untested_fields.append(field_name)

            if untested_fields:
                self.optional_field_coverage[schema_name] = untested_fields

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

        # CRUD coverage statistics
        resources_with_create = sum(1 for t in real_resources if t.has_create_test)
        resources_with_update = sum(1 for t in real_resources if t.has_update_test)
        resources_with_delete = sum(1 for t in real_resources if t.has_delete_test)

        # Quality metrics
        avg_quality_score = sum(t.quality_score for t in real_resources) / len(real_resources) if real_resources else 0

        # Optional field coverage
        total_optional_fields = sum(len(schema.optional_fields) for schema in self.resource_schemas.values())
        untested_optional_count = sum(len(fields) for fields in self.optional_field_coverage.values())
        tested_optional_count = total_optional_fields - untested_optional_count

        report.append(f"- **{total_modern_state}** modern state checks (`statecheck.ExpectKnownValue`)\n")
        report.append(f"- **{total_modern_plan}** modern plan checks (`plancheck.Expect*`)\n")
        report.append(f"- **{total_legacy}** legacy check calls (needs cleanup)\n")
        report.append(f"- **{resources_with_drift}/{len(real_resources)}** acceptance test resources have drift detection tests\n")
        report.append(f"- **{resources_with_import}/{len(real_resources)}** acceptance test resources have import tests\n")
        if total_required_fields > 0:
            report.append(f"- **{covered_validation_count}/{total_required_fields}** required fields have validation tests\n")
        if total_optional_fields > 0:
            report.append(f"- **{tested_optional_count}/{total_optional_fields}** optional fields are tested\n")
        report.append(f"\n**CRUD Coverage:**\n")
        report.append(f"- Create: {resources_with_create}/{len(real_resources)}\n")
        report.append(f"- Update: {resources_with_update}/{len(real_resources)}\n")
        report.append(f"- Delete: {resources_with_delete}/{len(real_resources)}\n")
        report.append(f"\n**Average Quality Score:** {avg_quality_score:.0f}/100\n")
        if mock_resources:
            report.append(f"\n**{len(mock_resources)}** mock/unit test files (import/drift N/A)\n")

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

                # CRUD coverage
                crud_ops = []
                if test.has_create_test:
                    crud_ops.append("Create")
                if test.has_update_test:
                    crud_ops.append("Update")
                if test.has_delete_test:
                    crud_ops.append("Delete")
                lines.append(f"- **CRUD coverage:** {', '.join(crud_ops) if crud_ops else 'None'}\n")

                # Quality score
                lines.append(f"- **Quality score:** {test.quality_score}/100\n")

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

        has_high_priority = False

        # Missing validation tests for required fields
        if self.validation_coverage:
            has_high_priority = True
            lines.append("\n**Missing Validation Tests for Required Fields:**\n")
            for resource_name, missing_fields in sorted(self.validation_coverage.items()):
                lines.append(f"- **`{resource_name}`**: {', '.join(missing_fields)}\n")

        if not has_high_priority:
            lines.append("\nNo high priority issues found! ✅\n")

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

        # Untested optional fields
        if self.optional_field_coverage:
            lines.append("\n**Untested Optional Fields:**\n")
            for resource_name, untested_fields in sorted(self.optional_field_coverage.items()):
                if len(untested_fields) > 5:
                    # Show first 5 and count
                    lines.append(f"- **`{resource_name}`**: {', '.join(untested_fields[:5])} (and {len(untested_fields) - 5} more)\n")
                else:
                    lines.append(f"- **`{resource_name}`**: {', '.join(untested_fields)}\n")

        # Low quality scores (below 70)
        low_quality = [t for t in self.resource_tests if t.quality_score < 70 and not t.is_mock_test]
        if low_quality:
            lines.append("\n**Tests with Low Quality Scores (<70/100):**\n")
            for test in sorted(low_quality, key=lambda x: x.quality_score):
                improvements = []
                if not test.uses_unique_names:
                    improvements.append("add unique names")
                if not test.uses_env_vars:
                    improvements.append("use env vars")
                if not test.has_precheck:
                    improvements.append("add PreCheck")
                if not test.has_check_destroy:
                    improvements.append("add CheckDestroy")
                lines.append(f"- `{test.name}` ({test.quality_score}/100) - Improve: {', '.join(improvements)}\n")

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
