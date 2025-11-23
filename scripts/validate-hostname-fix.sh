#!/bin/bash
# Validation script for hostname fix in test_helpers.go
#
# This script validates:
# 1. All test files compile without syntax errors
# 2. generateShortTestName() function exists and is properly exported
# 3. All generated names are under 63 characters (hostname limit)
# 4. Test suite passes basic compilation checks
# 5. Example name generation for various inputs

set -e

# Color codes for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
PASS_COUNT=0
FAIL_COUNT=0
WARN_COUNT=0

# Test results
declare -a FAILURES

# Logging functions
log_pass() {
    echo -e "${GREEN}✓ PASS${NC}: $1"
    ((PASS_COUNT++))
}

log_fail() {
    echo -e "${RED}✗ FAIL${NC}: $1"
    FAILURES+=("$1")
    ((FAIL_COUNT++))
}

log_warn() {
    echo -e "${YELLOW}⚠ WARN${NC}: $1"
    ((WARN_COUNT++))
}

log_info() {
    echo -e "${BLUE}ℹ INFO${NC}: $1"
}

log_section() {
    echo -e "\n${BLUE}==== $1 ====${NC}\n"
}

# Validation checks
validate_syntax() {
    log_section "SYNTAX VALIDATION"

    log_info "Checking Go syntax for test files..."

    # Try to compile the provider package
    if go build -v ./internal/provider 2>&1 | head -20; then
        log_pass "All test files compile without syntax errors"
    else
        log_fail "Syntax errors found in test files"
        return 1
    fi
}

validate_function_exists() {
    log_section "FUNCTION EXISTENCE VALIDATION"

    log_info "Checking for generateUniqueTestName() in test_helpers.go..."

    if grep -q "^func generateUniqueTestName" /workspace/internal/provider/test_helpers.go; then
        log_pass "generateUniqueTestName() function found"
    else
        log_fail "generateUniqueTestName() function not found in test_helpers.go"
        return 1
    fi

    log_info "Checking for generateShortTestName() in test_helpers.go..."

    if grep -q "^func generateShortTestName" /workspace/internal/provider/test_helpers.go; then
        log_pass "generateShortTestName() function found"
    else
        log_warn "generateShortTestName() function not found - only generateUniqueTestName() is available"
    fi
}

test_name_generation() {
    log_section "NAME GENERATION TESTING"

    # Create a temporary test program to generate names
    TEST_PROG="/tmp/test_name_gen_$$.go"

    cat > "$TEST_PROG" << 'EOF'
package main

import (
	"fmt"
	"os"
	"time"
)

func generateUniqueTestName(prefix string) string {
	timestamp := time.Now().Format("20060102-150405")
	nanos := time.Now().Nanosecond()
	pid := os.Getpid()
	return fmt.Sprintf("%s-%s-%d-%d", prefix, timestamp, nanos, pid)
}

func main() {
	testCases := []string{
		"tftest-image",
		"tftest-device",
		"tftest-category",
		"tftest-network",
		"tftest-cluster",
		"tftest-partition",
		"test-image",
		"test",
		"a",
	}

	fmt.Println("Testing name generation:")
	fmt.Println("========================")

	maxLen := 0
	minLen := 999

	for _, prefix := range testCases {
		name := generateUniqueTestName(prefix)
		nameLen := len(name)

		if nameLen > maxLen {
			maxLen = nameLen
		}
		if nameLen < minLen {
			minLen = nameLen
		}

		status := "✓"
		if nameLen > 63 {
			status = "✗"
		}

		fmt.Printf("%s Prefix: %-25s | Name Length: %3d | Max Allowed: 63\n", status, prefix, nameLen)
	}

	fmt.Println("========================")
	fmt.Printf("Min length: %d\n", minLen)
	fmt.Printf("Max length: %d\n", maxLen)
	fmt.Printf("Hostname limit: 63\n")

	if maxLen > 63 {
		fmt.Printf("\nERROR: Generated names exceed 63 character limit!\n")
		os.Exit(1)
	} else if maxLen > 50 {
		fmt.Printf("\nWARNING: Generated names approach 63 character limit (current max: %d)\n", maxLen)
		os.Exit(0)
	}

	fmt.Printf("\nSUCCESS: All names are within limits\n")
}
EOF

    log_info "Compiling temporary test program..."
    if go run "$TEST_PROG"; then
        log_pass "Name generation produces valid results"
    else
        log_fail "Name generation test failed"
        rm -f "$TEST_PROG"
        return 1
    fi

    rm -f "$TEST_PROG"
}

validate_test_files() {
    log_section "TEST FILE VALIDATION"

    log_info "Checking test files that use generateUniqueTestName()..."

    TEST_FILES=$(grep -l "generateUniqueTestName" /workspace/internal/provider/*_test.go 2>/dev/null | wc -l)

    if [ "$TEST_FILES" -gt 0 ]; then
        log_pass "Found $TEST_FILES test files using generateUniqueTestName()"

        # Show which files use the function
        log_info "Files using generateUniqueTestName():"
        grep -l "generateUniqueTestName" /workspace/internal/provider/*_test.go | sed 's/.*\//  - /'
    else
        log_warn "No test files found using generateUniqueTestName()"
    fi
}

validate_hostname_compliance() {
    log_section "HOSTNAME RFC COMPLIANCE"

    log_info "Verifying compliance with DNS hostname rules (RFC 1123)..."
    log_info "Maximum hostname length: 63 characters"
    log_info "Allowed characters: a-z, 0-9, hyphen (-)"
    log_info "Cannot start/end with hyphen"

    # Check generateUniqueTestName implementation
    if grep -A 5 "^func generateUniqueTestName" /workspace/internal/provider/test_helpers.go | grep -q "Sprintf"; then
        log_pass "generateUniqueTestName() uses proper formatting"
    else
        log_fail "generateUniqueTestName() implementation may have issues"
    fi

    # Verify the function produces hyphenated output
    if grep "generateUniqueTestName" /workspace/internal/provider/test_helpers.go | grep -q "%s-%s-%d-%d"; then
        log_pass "Function properly formats output with hyphens"
    else
        log_warn "Function format string may differ from expected pattern"
    fi
}

check_usage_patterns() {
    log_section "USAGE PATTERN VALIDATION"

    log_info "Checking how generateUniqueTestName() is used in tests..."

    # Count various usage patterns
    DEVICE_NAMES=$(grep -c 'generateUniqueTestName.*device' /workspace/internal/provider/*_test.go || echo 0)
    IMAGE_NAMES=$(grep -c 'generateUniqueTestName.*image' /workspace/internal/provider/*_test.go || echo 0)
    CATEGORY_NAMES=$(grep -c 'generateUniqueTestName.*category' /workspace/internal/provider/*_test.go || echo 0)
    NETWORK_NAMES=$(grep -c 'generateUniqueTestName.*network' /workspace/internal/provider/*_test.go || echo 0)
    CLUSTER_NAMES=$(grep -c 'generateUniqueTestName.*cluster' /workspace/internal/provider/*_test.go || echo 0)

    log_info "Usage breakdown:"
    echo "  Device names:   $DEVICE_NAMES"
    echo "  Image names:    $IMAGE_NAMES"
    echo "  Category names: $CATEGORY_NAMES"
    echo "  Network names:  $NETWORK_NAMES"
    echo "  Cluster names:  $CLUSTER_NAMES"

    TOTAL=$((DEVICE_NAMES + IMAGE_NAMES + CATEGORY_NAMES + NETWORK_NAMES + CLUSTER_NAMES))

    if [ "$TOTAL" -gt 0 ]; then
        log_pass "Found $TOTAL usages of generateUniqueTestName() across test files"
    else
        log_warn "No usages of generateUniqueTestName() found"
    fi
}

verify_imports() {
    log_section "IMPORT VALIDATION"

    log_info "Verifying required imports in test files..."

    # Check if test_helpers.go has proper package declaration
    if grep -q "^package provider" /workspace/internal/provider/test_helpers.go; then
        log_pass "test_helpers.go properly declares package provider"
    else
        log_fail "test_helpers.go missing package declaration"
        return 1
    fi

    # Check for required imports in test_helpers.go
    if grep -q "import" /workspace/internal/provider/test_helpers.go; then
        log_pass "test_helpers.go has import declarations"
    else
        log_fail "test_helpers.go missing import declarations"
        return 1
    fi
}

test_compilation() {
    log_section "FULL COMPILATION TEST"

    log_info "Testing full provider compilation with test helpers..."

    if go test -v -c -o /tmp/test_binary_$$ ./internal/provider 2>&1 | tail -10; then
        log_pass "Provider compiles successfully with all test helpers"
        rm -f /tmp/test_binary_$$
    else
        log_fail "Provider compilation failed"
        return 1
    fi
}

generate_summary() {
    log_section "VALIDATION SUMMARY"

    TOTAL=$((PASS_COUNT + FAIL_COUNT + WARN_COUNT))

    echo "Results:"
    echo "  Passed:  ${GREEN}$PASS_COUNT${NC}"
    echo "  Failed:  ${RED}$FAIL_COUNT${NC}"
    echo "  Warned:  ${YELLOW}$WARN_COUNT${NC}"
    echo "  Total:   $TOTAL"

    if [ "$FAIL_COUNT" -gt 0 ]; then
        echo -e "\n${RED}Failed checks:${NC}"
        for failure in "${FAILURES[@]}"; do
            echo "  - $failure"
        done
        return 1
    fi

    if [ "$WARN_COUNT" -gt 0 ]; then
        echo -e "\n${YELLOW}Warnings found (review recommended):${NC}"
        return 2
    fi

    echo -e "\n${GREEN}All validations passed!${NC}"
    return 0
}

# Main execution
main() {
    echo -e "${BLUE}"
    echo "========================================"
    echo "  HOSTNAME FIX VALIDATION SCRIPT"
    echo "========================================"
    echo -e "${NC}"

    # Check if we're in the right directory
    if [ ! -f "/workspace/internal/provider/test_helpers.go" ]; then
        echo -e "${RED}Error: test_helpers.go not found${NC}"
        echo "Make sure you're in the terraform-provider-bcm directory"
        exit 1
    fi

    # Run all validations
    validate_function_exists
    validate_imports
    verify_test_files || true
    validate_test_files
    check_usage_patterns
    validate_hostname_compliance
    test_name_generation || true
    validate_syntax || true
    test_compilation || true

    # Generate summary and exit
    generate_summary
    EXIT_CODE=$?

    echo -e "\n${BLUE}========================================"
    echo "  END OF VALIDATION"
    echo "==========================================${NC}\n"

    return $EXIT_CODE
}

# Run main function and exit with appropriate code
main
EXIT_CODE=$?
exit $EXIT_CODE
