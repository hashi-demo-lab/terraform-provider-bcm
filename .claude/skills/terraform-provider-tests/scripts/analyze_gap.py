#!/usr/bin/env python3
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

"""
Terraform Provider Test Coverage Analyzer

Scans test files and generates comprehensive coverage analysis reports.
Detects:
- Legacy Check blocks (resource.TestCheckResourceAttr)
- Missing drift detection tests
- Missing import tests
- Missing idempotency checks
- Modern pattern usage statistics
- Optional field coverage
- Test cleanup issues and verification
- ID consistency tracking issues
- Hardcoded credentials and secrets

Codebase Statistics:
- Total Go files and line counts
- Code, comment, and blank line breakdown
- File categorization (resource, data_source, test, other)
- Test-to-implementation ratio

ID Consistency Analysis:
- CompareValue(ValuesSame()) usage for ID tracking across steps
- Partial ID tracking (some steps tracked, not all)
- Legacy ID checks (TestCheckResourceAttr for "id")
- Missing ID verification in resource tests

Cleanup Analysis:
- Robust CheckDestroy verification
- Cleanup-friendly naming patterns (citest prefix, unique names)
- External resource cleanup
- Hardcoded vs dynamic resource names

Security Analysis:
- Hardcoded passwords and secrets
- Hardcoded API keys and tokens
- Hardcoded IP addresses and endpoints
- Credentials not using environment variables

Usage:
    python3 analyze_gap.py <test_directory> [--output report.md]

Example:
    python3 analyze_gap.py ./internal/provider/ --output ./ai_reports/tf_provider_tests_gap_$(date +%Y%m%d_%H%M%S).md
"""

import os
import re
import sys
import argparse
from collections import defaultdict
from pathlib import Path
from typing import Dict, List, Tuple, Set
from datetime import datetime


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
        self.is_error_only_test = False  # Tests that only test error paths
        self.uses_httptest = False  # Uses httptest.Server for mocking
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

        # Cleanup metrics
        self.has_robust_check_destroy = False  # CheckDestroy with actual verification logic
        self.uses_cleanup_names = False  # Uses cleanup-friendly naming (e.g., "citest")
        self.cleanup_issues = []  # List of cleanup concerns

        # Naming issues
        self.hardcoded_names = []  # List of (line_number, hardcoded_name) tuples
        self.non_unique_patterns = []  # List of non-unique naming patterns found

        # ID consistency tracking
        self.has_compare_value_for_id = False  # Uses CompareValue(ValuesSame()) for ID
        self.id_tracking_steps = 0  # Number of steps with AddStateValue for ID
        self.total_test_steps = 0  # Total number of test steps
        self.legacy_id_checks = []  # Legacy TestCheckResourceAttr for "id"
        self.modern_id_checks = 0  # Modern tfjsonpath.New("id") assertions
        self.id_consistency_issues = []  # List of ID consistency problems

        # Security: Hardcoded credentials
        self.hardcoded_credentials = []  # List of (line_num, credential_type, value, severity)

    def load(self):
        """Load file content."""
        with open(self.path, 'r') as f:
            self.content = f.read()

    def analyze(self):
        """Analyze file for patterns."""
        self._find_test_functions()
        self._detect_mock_and_error_tests()
        self._find_legacy_checks()
        self._find_modern_patterns()
        self._find_import_tests()
        self._find_drift_tests()
        self._find_idempotency_checks()
        self._find_validation_tests()
        self._analyze_crud_coverage()
        self._analyze_quality_metrics()
        self._analyze_cleanup_patterns()
        self._detect_hardcoded_names()
        self._analyze_id_consistency()
        self._detect_hardcoded_credentials()
        self._calculate_quality_score()

    def _find_test_functions(self):
        """Extract test function names."""
        pattern = r'func\s+(TestAcc\w+)\s*\('
        self.test_functions = re.findall(pattern, self.content)

    def _detect_mock_and_error_tests(self):
        """Detect if this is a mock server or error-only test file."""
        # Check for httptest.Server usage (mock server tests)
        self.uses_httptest = bool(re.search(r'httptest\.(?:New)?Server', self.content))

        # Check if ALL test steps use ExpectError (error-only tests)
        # Count total test steps
        test_steps = re.findall(r'Steps:\s*\[\]resource\.TestStep\{', self.content)

        if test_steps:
            # Count ExpectError usage
            expect_errors = len(re.findall(r'ExpectError:\s*regexp\.MustCompile', self.content))

            # Count non-error test steps (Check blocks or ConfigStateChecks without ExpectError)
            # This is approximate - we look for Config: without nearby ExpectError
            config_steps = len(re.findall(r'^\s*Config:\s+\w+', self.content, re.MULTILINE))

            # If we have ExpectError calls and they roughly match the number of test functions,
            # and we use httptest, this is likely an error-only test file
            if expect_errors >= len(self.test_functions) * 0.8 and self.uses_httptest:
                self.is_error_only_test = True

            # Alternative detection: Check for error scenario constants
            if re.search(r'type\s+\w+Scenario\s+string', self.content):
                if re.search(r'scenario\w+Error', self.content) or re.search(r'scenario\w+Failed', self.content):
                    self.is_error_only_test = True

        # Update is_mock_test flag only for error-only tests
        # Note: Just using httptest doesn't make it a mock - some real tests use httptest for specific validation
        if self.is_error_only_test:
            self.is_mock_test = True

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

    def _analyze_cleanup_patterns(self):
        """Analyze test cleanup patterns and identify issues."""
        # 1. Check for robust CheckDestroy (not just empty/nil)
        if self.has_check_destroy:
            # Look for CheckDestroy function with actual verification logic
            check_destroy_patterns = [
                r'CheckDestroy:\s*testAccCheck\w+Destroy',  # Named function
                r'verifyResourceDeleted\(',  # Uses helper
                r'client\.CallJSONRPC.*remove',  # API cleanup call
                r'if\s+.*still\s+exists',  # Verification logic
            ]

            self.has_robust_check_destroy = any(
                re.search(pattern, self.content)
                for pattern in check_destroy_patterns
            )

            # Check if CheckDestroy is just nil or empty
            if re.search(r'CheckDestroy:\s*nil', self.content):
                self.cleanup_issues.append("CheckDestroy is nil (no cleanup verification)")
            elif not self.has_robust_check_destroy and self.has_check_destroy:
                self.cleanup_issues.append("CheckDestroy present but may lack verification logic")

        # 2. Check for cleanup-friendly naming patterns
        cleanup_name_patterns = [
            r'generateUniqueTestName\(',  # Uses unique name generator
            r'citest[-_]',  # Uses citest prefix
            r'tftest[-_]',  # Uses tftest prefix
            r'test-\w+-\d+',  # Uses timestamped names
            r'fmt\.Sprintf.*%d.*time\.',  # Uses timestamp in name
        ]

        self.uses_cleanup_names = any(
            re.search(pattern, self.content)
            for pattern in cleanup_name_patterns
        )

        # 3. Detect potential cleanup issues
        # Resources created but no CheckDestroy (skip for error-only tests)
        if self.has_create_test and not self.has_check_destroy and not self.is_error_only_test:
            self.cleanup_issues.append("Creates resources but missing CheckDestroy")

        # Uses hardcoded names (harder to clean up)
        if not self.uses_cleanup_names and not self.uses_unique_names:
            # Check for hardcoded test names (not generated)
            if re.search(r'name.*=.*"test-.*"', self.content, re.IGNORECASE):
                self.cleanup_issues.append("Uses hardcoded resource names (harder to clean up)")

        # Creates external resources via API but may not clean up
        if 'createTestBCMClient' in self.content:
            # Test uses BCM client directly (creates external resources)
            if not self.has_robust_check_destroy:
                self.cleanup_issues.append("Creates external resources via API but cleanup verification unclear")

    def _detect_hardcoded_names(self):
        """Detect hardcoded test resource names that may not be unique."""
        lines = self.content.split('\n')

        # Skip error-only tests (they use mock servers, hardcoded names are OK)
        if self.is_error_only_test:
            return

        # Track if we're inside a block comment
        in_block_comment = False

        for i, line in enumerate(lines, 1):
            stripped = line.strip()

            # Track block comments
            if '/*' in line:
                in_block_comment = True
            if '*/' in line:
                in_block_comment = False
                continue

            # Skip if inside block comment or line comment
            if in_block_comment or stripped.startswith('//') or stripped.startswith('*'):
                continue

            # Skip lines that are clearly documentation (indented comments in test output)
            # Example: lines explaining scenarios, errors, solutions
            if any(keyword in stripped for keyword in [
                'Scenario', 'Solution:', 'Before:', 'After:', 'Expected:',
                'Trigger Conditions:', 'User Experience:', 'Recovery Strategy:'
            ]):
                continue

            # Look for hardcoded resource names in actual test configs
            # Only check top-level resource name/hostname attributes
            # Exclude nested attributes (modules, storage_classes, addons)

            # Check if we're inside a nested JSON structure (modules, storage_classes, addons)
            context_start = max(0, i - 10)
            context_lines = lines[context_start:i]
            context = '\n'.join(context_lines)

            # Skip if we're inside jsonencode, modules, or other nested blocks
            if any(pattern in context for pattern in [
                'modules = [', 'storage_classes = jsonencode', 'addons = jsonencode',
                'ingress_controller = jsonencode', 'parameters = {'
            ]):
                continue

            # Check for direct string assignments to top-level resource name/hostname fields
            # Only in actual Terraform config strings (within backticks or quotes)
            hardcoded_patterns = [
                (r'\bname\s*=\s*"([^"]+)"', 'name'),
                (r'\bhostname\s*=\s*"([^"]+)"', 'hostname'),
                (r'\bimage_name\s*=\s*"([^"]+)"', 'image_name'),
            ]

            # Skip lines that use string concatenation with variables
            # Pattern: `"` + varName + `"` indicates parameterized value from function argument
            if re.search(r'"\s*\+\s*\w+\s*\+\s*"', line):
                continue

            # Skip lines that are clearly Go string concatenation in config helpers
            # These pass parameters from test functions that use generateUniqueTestName
            if '`' in line and '+' in line:
                continue

            for pattern, field in hardcoded_patterns:
                matches = re.finditer(pattern, line)
                for match in matches:
                    value = match.group(1)

                    # Ignore if it's a variable placeholder or template
                    if '%' in value or '$' in value or '{' in value:
                        continue

                    # Ignore if it's clearly a reference (contains dots, underscores at start)
                    if value.startswith('.') or value.startswith('_'):
                        continue

                    # Ignore environment-specific values and documentation examples
                    if any(keyword in value for keyword in ['localhost', 'example', 'base', 'default', 'node', 'proxy', 'direct', 'standard', 'fast', 'prometheus']):
                        continue

                    # Check if the line contains generateUniqueTestName or variable reference
                    # Look at surrounding context (few lines before)
                    extended_context_start = max(0, i - 5)
                    extended_context = '\n'.join(lines[extended_context_start:i])

                    # If this value was generated by generateUniqueTestName, it's OK
                    if f'"{value}"' in extended_context and 'generateUniqueTestName' in extended_context:
                        continue

                    # If it's assigned from a variable, it's OK
                    # Look for: varName := "value" followed by name = varName
                    if re.search(rf'\w+\s*:=.*"{re.escape(value)}".*generateUniqueTestName', extended_context):
                        continue

                    # Check if it uses test-friendly prefixes
                    if not (value.startswith('tftest-') or value.startswith('citest-')):
                        # This is a hardcoded name without unique prefix
                        self.hardcoded_names.append((i, field, value))
                        self.non_unique_patterns.append(f"Line {i}: {field} = \"{value}\" (missing unique prefix)")

            # Also check for Sprintf without generateUniqueTestName
            sprintf_match = re.search(r'fmt\.Sprintf\s*\(\s*"([^"]*)"', line)
            if sprintf_match and 'name' in line.lower():
                template = sprintf_match.group(1)
                # If it doesn't include timestamp or random component, it might not be unique
                if '%d' not in template and 'time' not in line.lower():
                    # Check if this line also uses generateUniqueTestName
                    if 'generateUniqueTestName' not in line:
                        # This might be a non-unique pattern
                        if not any(prefix in template for prefix in ['tftest-', 'citest-', '%s']):
                            self.non_unique_patterns.append(f"Line {i}: fmt.Sprintf without unique component")

    def _analyze_id_consistency(self):
        """Analyze ID property usage consistency across test steps."""
        # Skip for data source tests (they don't typically manage ID lifecycle)
        if 'data_source_' in self.name:
            return

        # Skip for error-only tests
        if self.is_error_only_test:
            return

        # 1. Check for CompareValue usage for ID tracking
        compare_value_id_pattern = r'compareID\s*:=\s*statecheck\.CompareValue\(compare\.ValuesSame\(\)\)'
        self.has_compare_value_for_id = bool(re.search(compare_value_id_pattern, self.content))

        # 2. Count test steps with ConfigStateChecks
        # Look for { Config: patterns that start a TestStep
        test_step_pattern = r'\{\s*Config:\s*\w+'
        self.total_test_steps = len(re.findall(test_step_pattern, self.content))

        # 3. Count steps with ID tracking via AddStateValue
        add_state_id_pattern = r'compareID\.AddStateValue\s*\(\s*[^)]+,\s*tfjsonpath\.New\s*\(\s*"id"\s*\)'
        self.id_tracking_steps = len(re.findall(add_state_id_pattern, self.content))

        # 4. Find legacy ID checks (TestCheckResourceAttr for "id")
        legacy_id_patterns = [
            r'resource\.TestCheckResourceAttr\s*\([^,]+,\s*"id"',
            r'resource\.TestCheckResourceAttrSet\s*\([^,]+,\s*"id"',
        ]
        for pattern in legacy_id_patterns:
            matches = re.finditer(pattern, self.content)
            for match in matches:
                line_num = self.content[:match.start()].count('\n') + 1
                self.legacy_id_checks.append((line_num, match.group()))

        # 5. Count modern ID checks (tfjsonpath.New("id"))
        modern_id_pattern = r'tfjsonpath\.New\s*\(\s*"id"\s*\)'
        self.modern_id_checks = len(re.findall(modern_id_pattern, self.content))

        # 6. Identify ID consistency issues
        self._identify_id_consistency_issues()

    def _identify_id_consistency_issues(self):
        """Identify specific ID consistency problems."""
        # Issue 1: Uses legacy ID checks instead of modern patterns
        if self.legacy_id_checks:
            self.id_consistency_issues.append(
                f"Uses legacy ID checks ({len(self.legacy_id_checks)} occurrences) - migrate to statecheck.ExpectKnownValue"
            )

        # Issue 2: Has multiple test steps but no ID tracking
        if self.total_test_steps >= 2 and not self.has_compare_value_for_id:
            self.id_consistency_issues.append(
                f"Multiple test steps ({self.total_test_steps}) without ID consistency tracking - add CompareValue(ValuesSame())"
            )

        # Issue 3: Partial ID tracking (some steps tracked, not all)
        if self.has_compare_value_for_id and self.id_tracking_steps > 0:
            # Calculate expected steps (Create, Import if present, Update steps)
            expected_tracked_steps = self.total_test_steps
            # Import steps are tracked via ImportStateVerify, not AddStateValue, so adjust
            if self.has_import_test:
                # Import steps may use ConfigStateChecks in modern tests
                pass

            # If we have CompareValue but fewer tracked steps than expected, that's partial
            if self.id_tracking_steps < self.total_test_steps - 1:  # -1 for some tolerance (import step)
                self.id_consistency_issues.append(
                    f"Partial ID tracking: {self.id_tracking_steps}/{self.total_test_steps} steps track ID"
                )

        # Issue 4: Modern ID checks without consistency tracking
        if self.modern_id_checks > 0 and not self.has_compare_value_for_id and self.total_test_steps >= 2:
            self.id_consistency_issues.append(
                f"Uses {self.modern_id_checks} ExpectKnownValue for ID but no CompareValue consistency tracking"
            )

        # Issue 5: No ID verification at all in resource tests
        if (self.total_test_steps > 0 and
            self.modern_id_checks == 0 and
            len(self.legacy_id_checks) == 0 and
            'resource_' in self.name):
            self.id_consistency_issues.append(
                "No ID verification in any test step - resources should verify ID persistence"
            )

    def _detect_hardcoded_credentials(self):
        """Detect hardcoded credentials, secrets, and sensitive values."""
        lines = self.content.split('\n')

        # Track if we're inside a block comment
        in_block_comment = False

        # Patterns to detect (pattern, credential_type, severity)
        # severity: 'critical' = definite credential, 'warning' = potential issue
        credential_patterns = [
            # Passwords
            (r'password\s*[=:]\s*["\']([^"\']{4,})["\']', 'password', 'critical'),
            (r'passwd\s*[=:]\s*["\']([^"\']{4,})["\']', 'password', 'critical'),
            (r'pwd\s*[=:]\s*["\']([^"\']{4,})["\']', 'password', 'critical'),

            # API keys and tokens
            (r'api[_-]?key\s*[=:]\s*["\']([^"\']{8,})["\']', 'api_key', 'critical'),
            (r'apikey\s*[=:]\s*["\']([^"\']{8,})["\']', 'api_key', 'critical'),
            (r'secret[_-]?key\s*[=:]\s*["\']([^"\']{8,})["\']', 'secret_key', 'critical'),
            (r'access[_-]?key\s*[=:]\s*["\']([^"\']{8,})["\']', 'access_key', 'critical'),
            (r'auth[_-]?token\s*[=:]\s*["\']([^"\']{8,})["\']', 'auth_token', 'critical'),
            (r'bearer\s+([A-Za-z0-9_-]{20,})', 'bearer_token', 'critical'),

            # AWS-style keys
            (r'AKIA[0-9A-Z]{16}', 'aws_access_key', 'critical'),
            (r'aws[_-]?secret[_-]?access[_-]?key\s*[=:]\s*["\']([^"\']{20,})["\']', 'aws_secret_key', 'critical'),

            # Private keys
            (r'-----BEGIN\s+(?:RSA\s+)?PRIVATE\s+KEY-----', 'private_key', 'critical'),
            (r'-----BEGIN\s+OPENSSH\s+PRIVATE\s+KEY-----', 'ssh_private_key', 'critical'),

            # Connection strings with credentials
            (r'://[^:]+:([^@]{4,})@', 'connection_string_password', 'critical'),

            # Hardcoded IPs (warning level - may be intentional for tests)
            (r'endpoint\s*[=:]\s*["\']https?://(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})[:\d]*["\']', 'hardcoded_ip', 'warning'),

            # Generic secrets
            (r'secret\s*[=:]\s*["\']([^"\']{8,})["\']', 'secret', 'warning'),
            (r'credential\s*[=:]\s*["\']([^"\']{8,})["\']', 'credential', 'warning'),
        ]

        # Safe patterns to exclude (using env vars correctly)
        safe_patterns = [
            r'os\.Getenv\s*\(\s*["\']',  # Go env var
            r'\$\{?\w+\}?',  # Shell variable
            r'%\[\d+\]',  # Go format placeholder
            r'%[sqvd]',  # Go format verbs
        ]

        for i, line in enumerate(lines, 1):
            stripped = line.strip()

            # Track block comments
            if '/*' in line:
                in_block_comment = True
            if '*/' in line:
                in_block_comment = False
                continue

            # Skip comments
            if in_block_comment or stripped.startswith('//'):
                continue

            # Skip lines with safe patterns (env vars)
            if any(re.search(pattern, line) for pattern in safe_patterns):
                continue

            # Check each credential pattern
            for pattern, cred_type, severity in credential_patterns:
                matches = re.finditer(pattern, line, re.IGNORECASE)
                for match in matches:
                    value = match.group(1) if match.lastindex else match.group(0)

                    # Skip placeholder values
                    placeholder_indicators = [
                        'example', 'placeholder', 'changeme', 'xxx', 'yyy', 'zzz',
                        'your-', 'my-', 'test', 'dummy', 'fake', 'mock', 'sample',
                        '...', '***', '${', '%[', '%s', '%q', '<', '>'
                    ]
                    if any(ind in value.lower() for ind in placeholder_indicators):
                        continue

                    # Skip if it's clearly a variable reference
                    if re.match(r'^[a-z_]+$', value) and len(value) < 20:
                        continue

                    # Skip empty or very short values (likely false positives)
                    if len(value.strip()) < 4:
                        continue

                    # For IPs, check if it's a test/local IP
                    if cred_type == 'hardcoded_ip':
                        # 127.x.x.x, 10.x.x.x, 192.168.x.x, 172.16-31.x.x are often OK for tests
                        if re.match(r'^(127\.|10\.|192\.168\.|172\.(1[6-9]|2[0-9]|3[01])\.)', value):
                            severity = 'info'  # Downgrade to info for private IPs

                    self.hardcoded_credentials.append((i, cred_type, value[:50], severity))

    def _calculate_quality_score(self):
        """Calculate overall quality score (0-100)."""
        # Error-only tests have different scoring criteria
        if self.is_error_only_test:
            score = 0
            # Award points for error path coverage
            if len(self.test_functions) > 0:
                score += 40  # Good error coverage
            if self.uses_httptest:
                score += 30  # Uses proper mocking
            if len(self.legacy_checks) == 0:
                score += 20  # No legacy patterns
            if self.has_precheck:
                score += 10  # Has PreCheck
            self.quality_score = score
            return

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
        if self.uses_unique_names or self.uses_cleanup_names:
            score += 5
        if self.uses_env_vars:
            score += 5
        if self.has_precheck:
            score += 5
        if self.has_robust_check_destroy:
            score += 5
        elif self.has_check_destroy:
            score += 2  # Partial credit for having CheckDestroy without robust verification

        # Penalty for cleanup issues (not applied to mock/error tests)
        if not self.is_mock_test:
            cleanup_penalty = min(len(self.cleanup_issues) * 3, 15)  # Max -15 points
            score = max(0, score - cleanup_penalty)

        # Penalty for hardcoded names (not unique)
        if self.hardcoded_names:
            naming_penalty = min(len(self.hardcoded_names) * 5, 20)  # Max -20 points
            score = max(0, score - naming_penalty)

        # Penalty for ID consistency issues (resources only)
        if 'resource_' in self.name and self.id_consistency_issues:
            id_penalty = min(len(self.id_consistency_issues) * 5, 15)  # Max -15 points
            score = max(0, score - id_penalty)

        # Bonus for proper ID consistency tracking
        if self.has_compare_value_for_id and self.id_tracking_steps > 0:
            score = min(100, score + 5)  # Bonus for good ID tracking

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
        self.codebase_stats: Dict[str, Dict] = {}  # Codebase line count statistics

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

        # Analyze codebase line counts
        self._analyze_codebase_lines()

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

    def _analyze_codebase_lines(self):
        """Analyze line counts for Go files in the internal directory."""
        # Find the internal directory (parent of test_dir or test_dir itself)
        internal_dir = self.test_dir.parent
        if internal_dir.name != 'internal':
            # Check if test_dir is internal/provider, so parent is internal
            if self.test_dir.name == 'provider' and internal_dir.name == 'internal':
                pass  # internal_dir is correct
            else:
                # Try to find internal directory
                for parent in self.test_dir.parents:
                    if parent.name == 'internal':
                        internal_dir = parent
                        break
                    internal_internal = parent / 'internal'
                    if internal_internal.exists():
                        internal_dir = internal_internal
                        break

        if not internal_dir.exists():
            return

        # Initialize statistics
        stats = {
            'total': {'files': 0, 'lines': 0, 'code_lines': 0, 'comment_lines': 0, 'blank_lines': 0},
            'by_type': {
                'resource': {'files': 0, 'lines': 0, 'code_lines': 0},
                'data_source': {'files': 0, 'lines': 0, 'code_lines': 0},
                'test': {'files': 0, 'lines': 0, 'code_lines': 0},
                'other': {'files': 0, 'lines': 0, 'code_lines': 0},
            },
            'by_extension': {},
        }

        # Walk through all files in internal directory
        for root, dirs, files in os.walk(internal_dir):
            # Skip hidden directories and vendor
            dirs[:] = [d for d in dirs if not d.startswith('.') and d != 'vendor']

            for filename in files:
                file_path = Path(root) / filename
                ext = file_path.suffix

                # Only count Go files for detailed analysis
                if ext != '.go':
                    # Track other extensions briefly
                    if ext not in stats['by_extension']:
                        stats['by_extension'][ext] = {'files': 0, 'lines': 0}
                    try:
                        with open(file_path, 'r', errors='ignore') as f:
                            line_count = sum(1 for _ in f)
                        stats['by_extension'][ext]['files'] += 1
                        stats['by_extension'][ext]['lines'] += line_count
                    except:
                        pass
                    continue

                # Analyze Go file
                try:
                    with open(file_path, 'r', errors='ignore') as f:
                        content = f.read()
                        lines = content.split('\n')
                except:
                    continue

                total_lines = len(lines)
                blank_lines = sum(1 for line in lines if not line.strip())
                comment_lines = 0
                in_block_comment = False

                for line in lines:
                    stripped = line.strip()
                    if in_block_comment:
                        comment_lines += 1
                        if '*/' in stripped:
                            in_block_comment = False
                    elif stripped.startswith('//'):
                        comment_lines += 1
                    elif stripped.startswith('/*'):
                        comment_lines += 1
                        if '*/' not in stripped:
                            in_block_comment = True

                code_lines = total_lines - blank_lines - comment_lines

                # Update totals
                stats['total']['files'] += 1
                stats['total']['lines'] += total_lines
                stats['total']['code_lines'] += code_lines
                stats['total']['comment_lines'] += comment_lines
                stats['total']['blank_lines'] += blank_lines

                # Track by extension
                if ext not in stats['by_extension']:
                    stats['by_extension'][ext] = {'files': 0, 'lines': 0}
                stats['by_extension'][ext]['files'] += 1
                stats['by_extension'][ext]['lines'] += total_lines

                # Categorize by type
                fname = filename.lower()
                if '_test.go' in fname:
                    category = 'test'
                elif fname.startswith('resource_'):
                    category = 'resource'
                elif fname.startswith('data_source_'):
                    category = 'data_source'
                else:
                    category = 'other'

                stats['by_type'][category]['files'] += 1
                stats['by_type'][category]['lines'] += total_lines
                stats['by_type'][category]['code_lines'] += code_lines

        self.codebase_stats = stats

    def generate_report(self) -> str:
        """Generate markdown gap analysis report."""
        report = []
        report.append("# Terraform Provider Test Modernization Gap Analysis\n")
        report.append(f"**Analysis Date:** {self._get_date()}\n")
        report.append(f"**Test Directory:** `{self.test_dir}`\n")

        # Add codebase statistics section
        if self.codebase_stats and self.codebase_stats.get('total', {}).get('files', 0) > 0:
            report.append("\n## Codebase Statistics\n")
            stats = self.codebase_stats
            total = stats['total']
            by_type = stats['by_type']

            report.append(f"**Total Go Files:** {total['files']:,} files | {total['lines']:,} lines\n")
            report.append(f"- Code: {total['code_lines']:,} lines ({total['code_lines']*100//total['lines'] if total['lines'] else 0}%)\n")
            report.append(f"- Comments: {total['comment_lines']:,} lines ({total['comment_lines']*100//total['lines'] if total['lines'] else 0}%)\n")
            report.append(f"- Blank: {total['blank_lines']:,} lines ({total['blank_lines']*100//total['lines'] if total['lines'] else 0}%)\n")

            report.append("\n**By File Type:**\n")
            report.append("| Type | Files | Lines | Code Lines |\n")
            report.append("|------|------:|------:|-----------:|\n")

            for type_name, type_stats in sorted(by_type.items(), key=lambda x: x[1]['lines'], reverse=True):
                if type_stats['files'] > 0:
                    display_name = type_name.replace('_', ' ').title()
                    report.append(f"| {display_name} | {type_stats['files']:,} | {type_stats['lines']:,} | {type_stats['code_lines']:,} |\n")

            # Add total row
            report.append(f"| **Total** | **{total['files']:,}** | **{total['lines']:,}** | **{total['code_lines']:,}** |\n")

            # Calculate test-to-code ratio
            impl_lines = by_type['resource']['code_lines'] + by_type['data_source']['code_lines'] + by_type['other']['code_lines']
            test_lines = by_type['test']['code_lines']
            if impl_lines > 0:
                ratio = test_lines / impl_lines
                report.append(f"\n**Test-to-Implementation Ratio:** {ratio:.2f}:1 ({test_lines:,} test lines / {impl_lines:,} impl lines)\n")

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

        # Cleanup metrics (exclude error-only tests)
        real_resources_non_error = [t for t in real_resources if not t.is_error_only_test]
        resources_with_robust_cleanup = sum(1 for t in real_resources_non_error if t.has_robust_check_destroy)
        resources_with_cleanup_issues = sum(1 for t in real_resources_non_error if t.cleanup_issues)
        total_cleanup_issues = sum(len(t.cleanup_issues) for t in real_resources_non_error)

        # Naming metrics (non-unique names)
        all_tests_non_error = [t for t in self.resource_tests + self.data_source_tests if not t.is_error_only_test]
        tests_with_hardcoded_names = sum(1 for t in all_tests_non_error if t.hardcoded_names)
        total_hardcoded_names = sum(len(t.hardcoded_names) for t in all_tests_non_error)

        # Optional field coverage
        total_optional_fields = sum(len(schema.optional_fields) for schema in self.resource_schemas.values())
        untested_optional_count = sum(len(fields) for fields in self.optional_field_coverage.values())
        tested_optional_count = total_optional_fields - untested_optional_count

        # ID consistency metrics
        resources_with_id_tracking = sum(1 for t in real_resources if t.has_compare_value_for_id)
        resources_with_id_issues = sum(1 for t in real_resources if t.id_consistency_issues)
        total_id_issues = sum(len(t.id_consistency_issues) for t in real_resources)

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
        report.append(f"\n**Cleanup Analysis:**\n")
        if len(real_resources_non_error) > 0:
            report.append(f"- **{resources_with_robust_cleanup}/{len(real_resources_non_error)}** resources have robust cleanup verification\n")
            if resources_with_cleanup_issues > 0:
                report.append(f"- **{resources_with_cleanup_issues}/{len(real_resources_non_error)}** resources have cleanup issues ({total_cleanup_issues} total issues) ⚠️\n")
            else:
                report.append(f"- **No cleanup issues detected** ✅\n")
        else:
            report.append(f"- **All tests are error-only/mock tests** (cleanup verification N/A)\n")

        report.append(f"\n**Naming Uniqueness:**\n")
        if total_hardcoded_names > 0:
            report.append(f"- **{tests_with_hardcoded_names}/{len(all_tests_non_error)}** tests use hardcoded names ({total_hardcoded_names} total) ⚠️\n")
            report.append(f"- **Risk**: Name conflicts in parallel tests or after failed test runs\n")
        else:
            report.append(f"- **All tests use unique name generation** ✅\n")

        report.append(f"\n**ID Consistency Tracking:**\n")
        if len(real_resources) > 0:
            report.append(f"- **{resources_with_id_tracking}/{len(real_resources)}** resources use `CompareValue(ValuesSame())` for ID tracking\n")
            if resources_with_id_issues > 0:
                report.append(f"- **{resources_with_id_issues}/{len(real_resources)}** resources have ID consistency issues ({total_id_issues} total) ⚠️\n")
            else:
                report.append(f"- **All resources have consistent ID tracking** ✅\n")
        else:
            report.append(f"- **No resource tests found** (N/A)\n")

        report.append(f"\n**Average Quality Score:** {avg_quality_score:.0f}/100\n")
        if mock_resources:
            error_only_count = sum(1 for t in mock_resources if t.is_error_only_test)
            if error_only_count > 0:
                report.append(f"\n**{len(mock_resources)}** mock/unit test files ({error_only_count} error-only, import/drift N/A)\n")
            else:
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
                if test.is_error_only_test:
                    lines.append(f"- **Test type:** Error-only tests (uses httptest mock servers)\n")
                elif test.is_mock_test:
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

                # Cleanup status
                if test.cleanup_issues:
                    lines.append(f"- **Cleanup issues:** {len(test.cleanup_issues)} ⚠️\n")
                    for issue in test.cleanup_issues:
                        lines.append(f"  - {issue}\n")
                else:
                    cleanup_status = "✅ Robust" if test.has_robust_check_destroy else "✅ Basic"
                    if test.has_check_destroy:
                        lines.append(f"- **Cleanup:** {cleanup_status}\n")

                # Naming issues (hardcoded names)
                if test.hardcoded_names:
                    lines.append(f"- **Hardcoded names:** {len(test.hardcoded_names)} ⚠️ (may cause conflicts)\n")
                    for line_num, field, value in test.hardcoded_names[:5]:  # Show first 5
                        lines.append(f"  - Line {line_num}: {field} = \"{value}\"\n")
                    if len(test.hardcoded_names) > 5:
                        lines.append(f"  - (and {len(test.hardcoded_names) - 5} more)\n")

                # ID consistency tracking
                if test.has_compare_value_for_id:
                    lines.append(f"- **ID tracking:** ✅ Uses CompareValue ({test.id_tracking_steps}/{test.total_test_steps} steps)\n")
                elif test.total_test_steps >= 2:
                    lines.append(f"- **ID tracking:** ❌ Missing CompareValue for ID consistency\n")

                if test.id_consistency_issues:
                    lines.append(f"- **ID consistency issues:** {len(test.id_consistency_issues)} ⚠️\n")
                    for issue in test.id_consistency_issues:
                        lines.append(f"  - {issue}\n")

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

        # Error-only tests have different requirements
        if test.is_error_only_test:
            return "✅ Error path coverage (httptest mocks)"

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

        # Hardcoded names (non-unique naming patterns)
        tests_with_hardcoded_names = [t for t in self.resource_tests + self.data_source_tests
                                       if t.hardcoded_names and not t.is_error_only_test]
        if tests_with_hardcoded_names:
            has_high_priority = True
            lines.append("\n**Non-Unique Test Names (Hardcoded):**\n")
            lines.append("These tests use hardcoded resource names that may cause conflicts:\n\n")
            for test in sorted(tests_with_hardcoded_names, key=lambda x: len(x.hardcoded_names), reverse=True):
                lines.append(f"- **`{test.name}`** ({len(test.hardcoded_names)} hardcoded names):\n")
                for line_num, field, value in test.hardcoded_names[:3]:  # Show first 3
                    lines.append(f"  - Line {line_num}: `{field} = \"{value}\"` → Should use `generateUniqueTestName(\"tftest-{value}\")`\n")
                if len(test.hardcoded_names) > 3:
                    lines.append(f"  - (and {len(test.hardcoded_names) - 3} more)\n")
            lines.append("\n")

        # Cleanup issues across all tests (exclude error-only tests)
        tests_with_cleanup_issues = [t for t in self.resource_tests if t.cleanup_issues and not t.is_error_only_test]
        if tests_with_cleanup_issues:
            has_high_priority = True
            lines.append("\n**Test Cleanup Issues:**\n")
            for test in sorted(tests_with_cleanup_issues, key=lambda x: len(x.cleanup_issues), reverse=True):
                lines.append(f"- **`{test.name}`**:\n")
                for issue in test.cleanup_issues:
                    lines.append(f"  - {issue}\n")

        # Missing validation tests for required fields
        if self.validation_coverage:
            has_high_priority = True
            lines.append("\n**Missing Validation Tests for Required Fields:**\n")
            for resource_name, missing_fields in sorted(self.validation_coverage.items()):
                lines.append(f"- **`{resource_name}`**: {', '.join(missing_fields)}\n")

        # ID consistency issues
        tests_with_id_issues = [t for t in self.resource_tests
                                 if t.id_consistency_issues and not t.is_error_only_test]
        if tests_with_id_issues:
            has_high_priority = True
            lines.append("\n**ID Consistency Issues:**\n")
            lines.append("These tests have inconsistent ID property usage:\n\n")
            for test in sorted(tests_with_id_issues, key=lambda x: len(x.id_consistency_issues), reverse=True):
                lines.append(f"- **`{test.name}`** ({len(test.id_consistency_issues)} issues):\n")
                for issue in test.id_consistency_issues:
                    lines.append(f"  - {issue}\n")
            lines.append("\n")

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
                if not test.uses_unique_names and not test.uses_cleanup_names:
                    improvements.append("add unique/cleanup-friendly names")
                if not test.uses_env_vars:
                    improvements.append("use env vars")
                if not test.has_precheck:
                    improvements.append("add PreCheck")
                if not test.has_robust_check_destroy:
                    if test.has_check_destroy:
                        improvements.append("strengthen CheckDestroy verification")
                    else:
                        improvements.append("add CheckDestroy")
                if test.cleanup_issues:
                    improvements.append(f"fix {len(test.cleanup_issues)} cleanup issue(s)")
                if test.id_consistency_issues:
                    improvements.append(f"fix {len(test.id_consistency_issues)} ID consistency issue(s)")
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
        "example_resource.test",
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
// Initialize ID tracker at test start (before Steps)
compareID := statecheck.CompareValue(compare.ValuesSame())

// Step 1: Create - track ID
{
    Config: testAccResourceConfig(name),
    ConfigStateChecks: []statecheck.StateCheck{
        compareID.AddStateValue("example_resource.test", tfjsonpath.New("id")),
    },
}

// Step 2: Import - track ID to verify consistency
{
    ResourceName:      "example_resource.test",
    ImportState:       true,
    ImportStateVerify: true,
    ConfigStateChecks: []statecheck.StateCheck{
        compareID.AddStateValue("example_resource.test", tfjsonpath.New("id")),
    },
}

// Step 3: Update - track ID to ensure stability
{
    Config: testAccResourceConfig(name, "updated"),
    ConfigStateChecks: []statecheck.StateCheck{
        compareID.AddStateValue("example_resource.test", tfjsonpath.New("id")),
    },
}
// CompareValue(ValuesSame()) ensures ID remains identical across all steps
```

### Robust CheckDestroy Pattern
```go
func testAccCheckResourceDestroy(s *terraform.State) error {
    client := createTestBCMClient(&testing.T{})

    for _, rs := range s.RootModule().Resources {
        if rs.Type != "example_resource" {
            continue
        }

        // Verify resource deleted with retry logic
        deleted, err := verifyResourceDeleted(
            context.Background(),
            client,
            "Service",
            "getMethod",
            rs.Primary.ID,
            4, // retry count
        )

        if !deleted || err != nil {
            return fmt.Errorf("resource still exists after destroy: %s", rs.Primary.ID)
        }
    }

    return nil
}
```

### Cleanup-Friendly Naming
```go
// Use unique names for easy cleanup (recommended: tftest- prefix)
resourceName := generateUniqueTestName("tftest-resource")

// Alternative: citest prefix for CI/CD examples
resourceName := generateUniqueTestName("citest-resource")

// Manual timestamp-based naming
resourceName := fmt.Sprintf("tftest-%s-%d", "resource", time.Now().Unix())
```
"""

    def _get_date(self) -> str:
        """Get current date string."""
        from datetime import datetime
        return datetime.now().strftime("%Y-%m-%d")


def main():
    parser = argparse.ArgumentParser(
        description="Analyze Terraform provider test coverage and quality"
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
        output_path = Path(args.output)

        with open(output_path, 'w') as f:
            f.write(report)
        print(f"✅ Gap analysis report written to: {output_path}")
    else:
        print(report)


if __name__ == "__main__":
    main()
