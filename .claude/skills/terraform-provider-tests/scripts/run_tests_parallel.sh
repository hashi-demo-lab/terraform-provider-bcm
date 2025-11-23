#!/bin/bash
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

#
# Terraform Provider Parallel Test Execution Script
#
# Runs acceptance tests concurrently per test file to speed up execution.
# Supports filtering, concurrency control, and detailed reporting.
#
# Usage:
#   ./run_tests_parallel.sh [OPTIONS]
#
# Options:
#   -d, --dir DIR           Test directory (default: ./internal/provider)
#   -p, --pattern PATTERN   Test pattern to match (default: TestAcc)
#   -c, --concurrency N     Max concurrent test files (default: 4)
#   -t, --timeout DURATION  Timeout per test file (default: 30m)
#   -f, --file FILE         Run only tests from specific file
#   --resources-only        Run only resource tests
#   --data-sources-only     Run only data source tests
#   --verbose               Show detailed test output
#   --no-color              Disable colored output
#   -h, --help              Show this help message
#
# Examples:
#   # Run all acceptance tests with 4 concurrent files
#   ./run_tests_parallel.sh
#
#   # Run only resource tests with higher concurrency
#   ./run_tests_parallel.sh --resources-only -c 8
#
#   # Run tests matching specific pattern
#   ./run_tests_parallel.sh -p "TestAccCMPartSoftwareImage"
#
#   # Run tests from specific file
#   ./run_tests_parallel.sh -f resource_cmpart_softwareimage_test.go
#

set -euo pipefail

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Default values
TEST_DIR="./internal/provider"
TEST_PATTERN="TestAcc"
CONCURRENCY=4
TIMEOUT="30m"
SPECIFIC_FILE=""
RESOURCES_ONLY=false
DATA_SOURCES_ONLY=false
VERBOSE=false
NO_COLOR=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -d|--dir)
            TEST_DIR="$2"
            shift 2
            ;;
        -p|--pattern)
            TEST_PATTERN="$2"
            shift 2
            ;;
        -c|--concurrency)
            CONCURRENCY="$2"
            shift 2
            ;;
        -t|--timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        -f|--file)
            SPECIFIC_FILE="$2"
            shift 2
            ;;
        --resources-only)
            RESOURCES_ONLY=true
            shift
            ;;
        --data-sources-only)
            DATA_SOURCES_ONLY=true
            shift
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        --no-color)
            NO_COLOR=true
            shift
            ;;
        -h|--help)
            sed -n '3,39p' "$0" | sed 's/^# \?//'
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Disable colors if requested
if [ "$NO_COLOR" = true ]; then
    GREEN=''
    RED=''
    YELLOW=''
    BLUE=''
    CYAN=''
    NC=''
fi

# Validate directory
if [ ! -d "$TEST_DIR" ]; then
    echo -e "${RED}Error: Directory $TEST_DIR does not exist${NC}"
    exit 1
fi

# Check for required environment variables
if [ -z "${TF_ACC:-}" ]; then
    echo -e "${YELLOW}Warning: TF_ACC is not set. Set TF_ACC=1 to enable acceptance tests.${NC}"
    echo ""
fi

# Set Go environment variables
export GOMODCACHE="${GOMODCACHE:-/workspace/.go/pkg/mod}"
export GOCACHE="${GOCACHE:-/workspace/.go/cache}"
export GOPATH="${GOPATH:-/workspace/.go}"

# Create results directory
RESULTS_DIR=$(mktemp -d)
trap "rm -rf $RESULTS_DIR" EXIT

# Create output log file
OUTPUT_LOG="${RESULTS_DIR}/parallel_test_output.log"

echo "===== Terraform Provider Parallel Test Execution =====" | tee "$OUTPUT_LOG"
echo "Test Directory: $TEST_DIR" | tee -a "$OUTPUT_LOG"
echo "Test Pattern: $TEST_PATTERN" | tee -a "$OUTPUT_LOG"
echo "Concurrency: $CONCURRENCY" | tee -a "$OUTPUT_LOG"
echo "Timeout: $TIMEOUT" | tee -a "$OUTPUT_LOG"
echo "Output Log: $OUTPUT_LOG" | tee -a "$OUTPUT_LOG"
echo "" | tee -a "$OUTPUT_LOG"

# Find test files
TEST_FILES=()
if [ -n "$SPECIFIC_FILE" ]; then
    if [ -f "$TEST_DIR/$SPECIFIC_FILE" ]; then
        TEST_FILES=("$TEST_DIR/$SPECIFIC_FILE")
    else
        echo -e "${RED}Error: Test file $SPECIFIC_FILE not found in $TEST_DIR${NC}"
        exit 1
    fi
elif [ "$RESOURCES_ONLY" = true ]; then
    mapfile -t TEST_FILES < <(find "$TEST_DIR" -name "resource_*_test.go" | sort)
elif [ "$DATA_SOURCES_ONLY" = true ]; then
    mapfile -t TEST_FILES < <(find "$TEST_DIR" -name "data_source_*_test.go" | sort)
else
    mapfile -t TEST_FILES < <(find "$TEST_DIR" -name "*_test.go" | sort)
fi

if [ ${#TEST_FILES[@]} -eq 0 ]; then
    echo -e "${RED}Error: No test files found${NC}"
    exit 1
fi

echo "Found ${#TEST_FILES[@]} test file(s)" | tee -a "$OUTPUT_LOG"
echo "" | tee -a "$OUTPUT_LOG"

# Function to run tests from a single file
run_test_file() {
    local file=$1
    local file_base=$(basename "$file")
    local result_file="$RESULTS_DIR/${file_base}.result"
    local output_file="$RESULTS_DIR/${file_base}.output"
    local start_time=$(date +%s)

    echo -e "${CYAN}[START]${NC} $file_base" | tee -a "$OUTPUT_LOG"

    # Extract test function names from the file
    # This allows us to run only tests from this specific file
    local test_functions=$(grep -o "^func Test[A-Za-z0-9_]*" "$file" | sed 's/^func //' | paste -sd "|" -)

    if [ -z "$test_functions" ]; then
        echo -e "${YELLOW}[SKIP]${NC} $file_base (no test functions found)" | tee -a "$OUTPUT_LOG"
        echo "0|0|0|0|0" > "$result_file"
        return 0
    fi

    # Use file-specific test names as pattern
    # If TEST_PATTERN is "TestAcc", just use the extracted function names directly
    # since they already start with Test/TestAcc
    local combined_pattern="^(${test_functions})$"

    # Run the test
    if [ "$VERBOSE" = true ]; then
        TF_ACC=1 go test -v -timeout "$TIMEOUT" "$TEST_DIR" -run "$combined_pattern" \
            2>&1 | tee "$output_file"
        exit_code=${PIPESTATUS[0]}
    else
        TF_ACC=1 go test -v -timeout "$TIMEOUT" "$TEST_DIR" -run "$combined_pattern" \
            > "$output_file" 2>&1
        exit_code=$?
    fi

    local end_time=$(date +%s)
    local duration=$((end_time - start_time))

    # Parse test results
    local passed=$(grep -c "^--- PASS:" "$output_file" 2>/dev/null || echo 0)
    local failed=$(grep -c "^--- FAIL:" "$output_file" 2>/dev/null || echo 0)
    local skipped=$(grep -c "^--- SKIP:" "$output_file" 2>/dev/null || echo 0)

    # Save result summary
    echo "$exit_code|$passed|$failed|$skipped|$duration" > "$result_file"

    if [ $exit_code -eq 0 ]; then
        echo -e "${GREEN}[PASS]${NC} $file_base (${duration}s, passed: $passed)" | tee -a "$OUTPUT_LOG"
    else
        echo -e "${RED}[FAIL]${NC} $file_base (${duration}s, passed: $passed, failed: $failed)" | tee -a "$OUTPUT_LOG"

        # Show failures if not verbose
        if [ "$VERBOSE" = false ]; then
            echo -e "${YELLOW}Failures in $file_base:${NC}" | tee -a "$OUTPUT_LOG"
            grep -A 5 "^--- FAIL:" "$output_file" | head -n 30 | tee -a "$OUTPUT_LOG"
            echo "" | tee -a "$OUTPUT_LOG"
        fi
    fi

    return $exit_code
}

export -f run_test_file
export RESULTS_DIR TEST_DIR TEST_PATTERN TIMEOUT VERBOSE OUTPUT_LOG
export GREEN RED YELLOW BLUE CYAN NC

# Run tests in parallel using GNU parallel or xargs
if command -v parallel &> /dev/null; then
    # Use GNU parallel for better progress tracking
    printf '%s\n' "${TEST_FILES[@]}" | \
        parallel -j "$CONCURRENCY" --line-buffer run_test_file {}
    parallel_exit=$?
else
    # Fallback to xargs
    printf '%s\n' "${TEST_FILES[@]}" | \
        xargs -I {} -P "$CONCURRENCY" bash -c 'run_test_file "$@"' _ {}
    parallel_exit=$?
fi

echo ""
echo "===== Test Summary ====="
echo ""

# Aggregate results
total_passed=0
total_failed=0
total_skipped=0
total_duration=0
failed_files=()

for file in "${TEST_FILES[@]}"; do
    file_base=$(basename "$file")
    result_file="$RESULTS_DIR/${file_base}.result"

    if [ -f "$result_file" ]; then
        IFS='|' read -r exit_code passed failed skipped duration < "$result_file"
        total_passed=$((total_passed + passed))
        total_failed=$((total_failed + failed))
        total_skipped=$((total_skipped + skipped))
        total_duration=$((total_duration + duration))

        if [ "$exit_code" -ne 0 ]; then
            failed_files+=("$file_base")
        fi
    fi
done

# Display summary
echo "Test Files: ${#TEST_FILES[@]}"
echo "Total Passed: ${total_passed}"
echo "Total Failed: ${total_failed}"
echo "Total Skipped: ${total_skipped}"
echo "Total Duration: ${total_duration}s"
echo ""

if [ ${#failed_files[@]} -gt 0 ]; then
    echo -e "${RED}Failed Files (${#failed_files[@]}):${NC}"
    for file in "${failed_files[@]}"; do
        echo "  - $file"
    done
    echo ""
    echo -e "${YELLOW}Detailed output saved in: $RESULTS_DIR${NC}"
    echo ""
    exit 1
else
    echo -e "${GREEN}✅ All tests passed!${NC}"
    echo ""
    exit 0
fi
