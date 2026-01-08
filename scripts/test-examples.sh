#!/usr/bin/env bash
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

#
# Terraform Provider BCM - Example Test Suite
#
# This script is based on the generic template from:
# .claude/skills/terraform-provider-tests/templates/test-examples-template.sh
#
# IMPORTANT: The cleanup_resources() function (lines 890-1041) is BCM-specific
# and must remain in this project. The generic test framework does not include
# provider-specific cleanup logic.
#
# See: .claude/skills/terraform-provider-tests/docs/example-testing-guide.md
#

set -euo pipefail

# T031: Trap handlers for cleanup on exit/interrupt
INTERRUPTED=false

cleanup_on_exit() {
    if [ "$INTERRUPTED" = true ]; then
        log_info ""
        log_info "Interrupted by user - cleaning up before exit..."
        cleanup_resources || true
        exit "$EXIT_INTERRUPTED"
    fi
}

handle_interrupt() {
    INTERRUPTED=true
    log_info ""
    log_info "Received interrupt signal (Ctrl+C)..."
    cleanup_on_exit
}

trap cleanup_on_exit EXIT
trap handle_interrupt SIGINT SIGTERM

# Exit codes
readonly EXIT_SUCCESS=0
readonly EXIT_TEST_FAILURE=1
readonly EXIT_CONFIG_ERROR=2
readonly EXIT_BUILD_FAILURE=3
readonly EXIT_INTERRUPTED=130

# Color codes for output
readonly COLOR_INFO='\033[0;36m'
readonly COLOR_PASS='\033[0;32m'
readonly COLOR_FAIL='\033[0;31m'
readonly COLOR_ERROR='\033[0;31m'
readonly COLOR_RESET='\033[0m'

# Default configuration
PROVIDER_VERSION="${PROVIDER_VERSION:-0.1.0}"
SKIP_BUILD="${SKIP_BUILD:-false}"
PARALLEL_LIMIT="${PARALLEL_LIMIT:-4}"
CLEANUP_RETRIES="${CLEANUP_RETRIES:-4}"
CLEANUP_ONLY="${CLEANUP_ONLY:-false}"
VERBOSE="${VERBOSE:-false}"
NO_CLEANUP="${NO_CLEANUP:-false}"
DATA_SOURCES_ONLY="${DATA_SOURCES_ONLY:-false}"
RESOURCES_ONLY="${RESOURCES_ONLY:-false}"

# Dev overrides detection
DEV_OVERRIDES_PATH=""
USE_DEV_OVERRIDES=false

# Test results tracking
PASSED_COUNT=0
FAILED_COUNT=0
FAILED_EXAMPLES=()

# Timing tracking
START_TIME=0
BUILD_TIME=0
DATA_SOURCES_TIME=0
RESOURCES_TIME=0
CLEANUP_TIME=0
CLEANUP_RESOURCE_COUNT=0

#############################################################################
# T014: CLI Parsing and Help
#############################################################################

show_help() {
    cat <<EOF
Terraform Example Test Suite

Validates all Terraform examples in the examples/ directory by building the
provider, executing examples, and cleaning up test resources.

USAGE:
  ./scripts/test-examples.sh [OPTIONS]

OPTIONS:
  --help                  Display this help message
  --cleanup-only          Skip tests, only cleanup existing test resources
  --verbose               Show detailed output including terraform logs
  --no-cleanup            Skip cleanup phase (useful for debugging)
  --data-sources-only     Only test data source examples
  --resources-only        Only test resource examples

ENVIRONMENT VARIABLES (Required):
  BCM_ENDPOINT        BCM API endpoint (e.g., https://172.21.15.254:8081)
  BCM_USERNAME        BCM authentication username (e.g., root)
  BCM_PASSWORD        BCM authentication password

ENVIRONMENT VARIABLES (Optional):
  PROVIDER_VERSION    Provider version for binary (default: 0.1.0)
  SKIP_BUILD          Skip provider build phase (default: false)
  PARALLEL_LIMIT      Max parallel data source tests (default: 4)
  CLEANUP_RETRIES     Max cleanup retry attempts (default: 4)
  VERBOSE             Enable verbose logging (default: false)

EXIT CODES:
  0   All tests passed, cleanup successful
  1   One or more tests failed
  2   Configuration error (missing env vars)
  3   Provider build failed
  130 Interrupted by user (Ctrl+C)

EXAMPLES:
  # Run all tests
  ./scripts/test-examples.sh

  # Debug a failing resource test
  ./scripts/test-examples.sh --resources-only --no-cleanup --verbose

  # Cleanup orphaned resources
  ./scripts/test-examples.sh --cleanup-only

  # Quick validation (data sources only)
  ./scripts/test-examples.sh --data-sources-only

For more information, see:
  /workspace/specs/001-example-test-infrastructure/quickstart.md

EOF
    exit 0
}

# Parse command-line arguments
for arg in "$@"; do
    case "$arg" in
        --help)
            show_help
            ;;
        --cleanup-only)
            CLEANUP_ONLY=true
            ;;
        --verbose)
            VERBOSE=true
            ;;
        --no-cleanup)
            NO_CLEANUP=true
            ;;
        --data-sources-only)
            DATA_SOURCES_ONLY=true
            ;;
        --resources-only)
            RESOURCES_ONLY=true
            ;;
        *)
            echo -e "${COLOR_ERROR}[ERROR] Unknown argument: $arg${COLOR_RESET}"
            echo "Use --help for usage information"
            exit "$EXIT_CONFIG_ERROR"
            ;;
    esac
done

#############################################################################
# T015: Environment Variable Validation
#############################################################################

log_info() {
    if [ "$VERBOSE" = true ]; then
        echo -e "${COLOR_INFO}[INFO]${COLOR_RESET} $(date '+%Y-%m-%d %H:%M:%S') | $*"
    else
        echo -e "${COLOR_INFO}[INFO]${COLOR_RESET} $*"
    fi
}

log_pass() {
    if [ "$VERBOSE" = true ]; then
        echo -e "${COLOR_PASS}[PASS]${COLOR_RESET} $(date '+%Y-%m-%d %H:%M:%S') | $*"
    else
        echo -e "${COLOR_PASS}[PASS]${COLOR_RESET} $*"
    fi
}

log_fail() {
    if [ "$VERBOSE" = true ]; then
        echo -e "${COLOR_FAIL}[FAIL]${COLOR_RESET} $(date '+%Y-%m-%d %H:%M:%S') | $*"
    else
        echo -e "${COLOR_FAIL}[FAIL]${COLOR_RESET} $*"
    fi
}

log_error() {
    if [ "$VERBOSE" = true ]; then
        echo -e "${COLOR_ERROR}[ERROR]${COLOR_RESET} $(date '+%Y-%m-%d %H:%M:%S') | $*"
    else
        echo -e "${COLOR_ERROR}[ERROR]${COLOR_RESET} $*"
    fi
}

log_debug() {
    if [ "$VERBOSE" = true ]; then
        echo -e "${COLOR_INFO}[DEBUG]${COLOR_RESET} $(date '+%Y-%m-%d %H:%M:%S') | $*"
    fi
}

#############################################################################
# Detect dev_overrides in .terraformrc
#############################################################################

detect_dev_overrides() {
    local terraformrc="$HOME/.terraformrc"

    if [ ! -f "$terraformrc" ]; then
        log_debug "No .terraformrc found, using standard plugin directory"
        return
    fi

    # Check for hashicorp/bcm dev_overrides
    if grep -q '"hashicorp/bcm"' "$terraformrc" 2>/dev/null; then
        # Extract the dev_overrides path
        DEV_OVERRIDES_PATH=$(grep -A1 '"hashicorp/bcm"' "$terraformrc" | grep -oE '= ".*"' | sed 's/= "//;s/"$//' | head -1)

        if [ -z "$DEV_OVERRIDES_PATH" ]; then
            # Try alternate format: hashicorp/bcm = "/path"
            DEV_OVERRIDES_PATH=$(grep '"hashicorp/bcm"' "$terraformrc" | grep -oE '= ".*"' | sed 's/= "//;s/"$//')
        fi

        if [ -n "$DEV_OVERRIDES_PATH" ]; then
            USE_DEV_OVERRIDES=true
            log_debug "Detected dev_overrides for hashicorp/bcm: $DEV_OVERRIDES_PATH"
        fi
    fi
}

validate_environment() {
    # Validate flag combinations first
    if [ "$DATA_SOURCES_ONLY" = true ] && [ "$RESOURCES_ONLY" = true ]; then
        log_error "Cannot use --data-sources-only and --resources-only together"
        exit "$EXIT_CONFIG_ERROR"
    fi

    if [ "$CLEANUP_ONLY" = true ] && [ "$NO_CLEANUP" = true ]; then
        log_error "Cannot use --cleanup-only and --no-cleanup together"
        exit "$EXIT_CONFIG_ERROR"
    fi

    if [ "$CLEANUP_ONLY" = true ] && [ "$DATA_SOURCES_ONLY" = true ]; then
        log_error "Cannot use --cleanup-only with --data-sources-only"
        exit "$EXIT_CONFIG_ERROR"
    fi

    if [ "$CLEANUP_ONLY" = true ] && [ "$RESOURCES_ONLY" = true ]; then
        log_error "Cannot use --cleanup-only with --resources-only"
        exit "$EXIT_CONFIG_ERROR"
    fi

    log_info "========================================"
    log_info "Terraform Example Test Suite"
    log_info "========================================"
    log_debug "BCM Endpoint: ${BCM_ENDPOINT:-<not set>}"
    log_debug "Provider Version: $PROVIDER_VERSION"

    # Determine execution mode
    local exec_mode="all"
    if [ "$CLEANUP_ONLY" = true ]; then
        exec_mode="cleanup-only"
    elif [ "$DATA_SOURCES_ONLY" = true ]; then
        exec_mode="data-sources-only"
    elif [ "$RESOURCES_ONLY" = true ]; then
        exec_mode="resources-only"
    fi
    log_debug "Execution Mode: $exec_mode"
    log_info ""
    log_info "Phase 1: Validating environment..."

    local missing_vars=()

    if [ -z "${BCM_ENDPOINT:-}" ]; then
        missing_vars+=("BCM_ENDPOINT")
    else
        log_info "✓ BCM_ENDPOINT set"
    fi

    if [ -z "${BCM_USERNAME:-}" ]; then
        missing_vars+=("BCM_USERNAME")
    else
        log_info "✓ BCM_USERNAME set"
    fi

    if [ -z "${BCM_PASSWORD:-}" ]; then
        missing_vars+=("BCM_PASSWORD")
    else
        log_info "✓ BCM_PASSWORD set"
    fi

    if [ ${#missing_vars[@]} -gt 0 ]; then
        log_error "Missing required environment variables: ${missing_vars[*]}"
        log_error ""
        log_error "Context: Environment validation phase"
        log_error "Required: BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD"
        log_error ""
        log_error "Example:"
        log_error "  export BCM_ENDPOINT=\"https://172.21.15.254:8081\""
        log_error "  export BCM_USERNAME=\"root\""
        log_error "  export BCM_PASSWORD=\"your-password\""
        log_error "  ./scripts/test-examples.sh"
        log_error ""
        log_error "Suggestion: Set all required environment variables and retry"
        log_error "For more information: ./scripts/test-examples.sh --help"
        exit "$EXIT_CONFIG_ERROR"
    fi

    # Detect dev_overrides configuration
    detect_dev_overrides
    if [ "$USE_DEV_OVERRIDES" = true ]; then
        log_info "✓ dev_overrides detected: $DEV_OVERRIDES_PATH"
        log_info "  (terraform init will be skipped)"
    fi

    log_info ""
}

#############################################################################
# T016: Provider Binary Build
#############################################################################

build_provider() {
    if [ "$SKIP_BUILD" = "true" ]; then
        log_info "Phase 2: Skipping provider build (SKIP_BUILD=true)..."
        log_info ""
        return
    fi

    local build_start
    build_start=$(date +%s)

    log_info "Phase 2: Building provider binary..."

    # Detect platform - use terraform's reported platform to handle emulation scenarios
    local os arch tf_platform
    tf_platform=$(terraform version | grep -oE 'linux_[a-z0-9]+' | head -1)
    if [ -n "$tf_platform" ]; then
        os="${tf_platform%%_*}"
        arch="${tf_platform#*_}"
    else
        # Fallback to uname detection
        os=$(uname -s | tr '[:upper:]' '[:lower:]')
        case "$(uname -m)" in
            x86_64) arch="amd64" ;;
            arm64|aarch64) arch="arm64" ;;
            *) arch="amd64" ;;
        esac
    fi

    log_info "Detected platform: ${os}_${arch}"
    log_debug "Build directory: /workspace"

    # Build binary
    local binary_name="terraform-provider-bcm_v${PROVIDER_VERSION}"
    log_info "Building: $binary_name"

    # Set GOOS/GOARCH for cross-compilation when host != target
    export GOOS="$os"
    export GOARCH="$arch"
    log_debug "Executing: GOOS=$os GOARCH=$arch go build -o $binary_name ."

    cd /workspace
    local build_output
    if ! build_output=$(go build -o "$binary_name" . 2>&1); then
        log_error "Build failed: go build command failed"
        log_error ""
        log_error "Context: Phase 2 - Provider build"
        log_error "Command: go build -o $binary_name ."
        log_error ""
        if [ "$VERBOSE" = true ]; then
            log_error "Build output:"
            echo "$build_output" | while IFS= read -r line; do
                log_error "  $line"
            done
        else
            log_error "First error:"
            echo "$build_output" | head -5 | while IFS= read -r line; do
                log_error "  $line"
            done
        fi
        log_error ""
        log_error "Suggestion: Fix compilation errors and rerun script"
        log_error "Run with --verbose for full build output"
        exit "$EXIT_BUILD_FAILURE"
    fi

    # Check binary was created
    if [ ! -f "$binary_name" ]; then
        log_error "Build failed: binary not found after build"
        log_error "Context: Phase 2 - Provider build verification"
        log_error "Expected: $binary_name in /workspace"
        log_error "Suggestion: Check disk space and permissions"
        exit "$EXIT_BUILD_FAILURE"
    fi

    local binary_size
    binary_size=$(du -h "$binary_name" | cut -f1)
    log_info "✓ Provider built successfully ($binary_size)"
    log_debug "Binary path: /workspace/$binary_name"

    # Install to appropriate directory based on dev_overrides configuration
    if [ "$USE_DEV_OVERRIDES" = true ] && [ -n "$DEV_OVERRIDES_PATH" ]; then
        # Install to dev_overrides path (no version suffix needed)
        local install_path="$DEV_OVERRIDES_PATH"
        mkdir -p "$install_path"
        cp "$binary_name" "$install_path/terraform-provider-bcm"
        log_info "✓ Installed to dev_overrides: $install_path/terraform-provider-bcm"
    else
        # Install to plugin directory (standard path)
        local plugin_dir="$HOME/.terraform.d/plugins/registry.terraform.io/hashicorp/bcm/${PROVIDER_VERSION}/${os}_${arch}"
        log_debug "Plugin directory: $plugin_dir"
        mkdir -p "$plugin_dir"
        cp "$binary_name" "$plugin_dir/"
        log_info "✓ Installed to: $plugin_dir"
    fi

    local build_end
    build_end=$(date +%s)
    BUILD_TIME=$((build_end - build_start))
    log_debug "Build completed in ${BUILD_TIME}s"
    log_info ""
}

#############################################################################
# T017: Example Discovery
# T022: Categorize examples into data sources vs resources
#############################################################################

discover_examples() {
    local examples_dir="/workspace/examples"
    local data_source_examples=()
    local resource_examples=()

    # Discover data source examples (parallel-safe)
    # Each .tf file is tested individually as a separate example
    local data_sources_dir="$examples_dir/data-sources"
    if [ -d "$data_sources_dir" ]; then
        while IFS= read -r -d '' tf_file; do
            data_source_examples+=("$tf_file")
        done < <(find "$data_sources_dir" -mindepth 2 -maxdepth 2 -name "*.tf" -print0)
    fi

    # Discover resource examples (sequential only)
    # Each .tf file is tested individually as a separate example
    local resources_dir="$examples_dir/resources"
    if [ -d "$resources_dir" ]; then
        while IFS= read -r -d '' tf_file; do
            resource_examples+=("$tf_file")
        done < <(find "$resources_dir" -mindepth 2 -maxdepth 2 -name "*.tf" -print0)
    fi

    local total_examples=$((${#data_source_examples[@]} + ${#resource_examples[@]}))
    if [ $total_examples -eq 0 ]; then
        log_error "No test examples found in $examples_dir"
        log_error "Expected to find .tf files in: $examples_dir/data-sources/*/ and $examples_dir/resources/*/"
        exit "$EXIT_CONFIG_ERROR"
    fi

    # Return as two space-separated lists with a delimiter
    # Format: "data_source1 data_source2 ... | resource1 resource2 ..."
    echo "${data_source_examples[*]} | ${resource_examples[*]}"
}

#############################################################################
# T018: Sequential Execution
#############################################################################

# Test result files for parallel execution
declare -A PARALLEL_RESULTS

test_example() {
    local example_file="$1"
    local example_index="$2"
    local total_examples="$3"
    local example_name
    # Display as "category/filename"
    local dir_name=$(basename "$(dirname "$example_file")")
    local file_name=$(basename "$example_file")
    example_name="$dir_name/$file_name"

    local test_start
    test_start=$(date +%s)

    log_info "[$example_index/$total_examples] Testing $example_name..."
    log_debug "Example file: $example_file"

    # Create temporary working directory
    local temp_dir
    temp_dir=$(mktemp -d -t "bcm-test-${dir_name}-XXXXX")
    log_debug "Temp directory: $temp_dir"

    # Copy single example file to temp directory
    cp "$example_file" "$temp_dir/"
    log_debug "Copied $file_name to $temp_dir"

    # Inject provider configuration for examples that don't include provider blocks
    # Provider reads authentication from environment variables (BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD)
    if ! grep -q "provider \"bcm\"" "$temp_dir"/*.tf 2>/dev/null; then
        cat > "$temp_dir/_provider.tf" <<EOF
terraform {
  required_version = ">= 1.5.0"
  required_providers {
    bcm = {
      source  = "hashicorp/bcm"
      version = "~> 0.1"
    }
  }
}

provider "bcm" {
  insecure_skip_verify = true
}
EOF
    fi

    cd "$temp_dir"

    local test_passed=true
    local error_output=""
    local failed_phase=""

    # Run terraform init (skip if using dev_overrides)
    if [ "$USE_DEV_OVERRIDES" = true ]; then
        log_debug "  ├─ terraform init (skipped - using dev_overrides)"
    else
        log_debug "  ├─ terraform init..."
        if ! error_output=$(terraform init -backend=false 2>&1); then
            failed_phase="terraform init"
            test_passed=false
        fi
    fi

    # Run terraform validate
    if [ "$test_passed" = true ]; then
        log_debug "  ├─ terraform validate..."
        if ! error_output=$(terraform validate 2>&1); then
            failed_phase="terraform validate"
            test_passed=false
        fi
    fi

    # Run terraform plan
    if [ "$test_passed" = true ]; then
        log_debug "  ├─ terraform plan..."
        if ! error_output=$(terraform plan -out=tfplan 2>&1); then
            failed_phase="terraform plan"
            test_passed=false
        fi
    fi

    # Determine if we should apply/destroy based on example type
    # Apply/destroy for "test-citest" examples only (full integration testing)
    # Skip apply/destroy for documentation examples (plan-only validation)
    local should_apply=false
    if echo "$example_name" | grep -q "test-citest"; then
        should_apply=true
    fi

    # Run terraform apply (for test examples) - creates actual infrastructure
    if [ "$test_passed" = true ] && [ "$should_apply" = true ]; then
        log_debug "  ├─ terraform apply..."
        if ! error_output=$(terraform apply -auto-approve tfplan 2>&1); then
            failed_phase="terraform apply"
            test_passed=false
        fi
    fi

    # Run terraform destroy - cleanup created resources
    if [ "$test_passed" = true ] && [ "$should_apply" = true ]; then
        log_debug "  └─ terraform destroy..."
        if ! error_output=$(terraform destroy -auto-approve 2>&1); then
            failed_phase="terraform destroy"
            test_passed=false
        fi
    fi

    # Cleanup temp directory
    cd /workspace
    rm -rf "$temp_dir"

    local test_end
    test_end=$(date +%s)
    local test_time=$((test_end - test_start))

    if [ "$test_passed" = true ]; then
        log_pass "[PASS] ✓ $example_name (${test_time}s)"
        PASSED_COUNT=$((PASSED_COUNT + 1))
        return 0
    else
        log_fail "[FAIL] ✗ $example_name (${test_time}s)"
        log_error "       Context: Phase 3 - Sequential test execution"
        log_error "       Failed at: $failed_phase"
        log_error "       Example: $example_file"
        if [ "$VERBOSE" = true ]; then
            log_error "       Full output:"
            echo "$error_output" | while IFS= read -r line; do
                log_error "         $line"
            done
        else
            log_error "       Error:"
            echo "$error_output" | head -10 | while IFS= read -r line; do
                log_error "         $line"
            done
            log_error "       Run with --verbose for full output"
        fi
        log_error "       Suggestion: Review example configuration and fix errors"
        FAILED_COUNT=$((FAILED_COUNT + 1))
        FAILED_EXAMPLES+=("$example_name")
        return 1
    fi
}

#############################################################################
# T023: Parallel Execution for Data Sources
#############################################################################

test_example_parallel() {
    local example_file="$1"
    local result_file="$2"
    # Display as "category/filename"
    local dir_name=$(basename "$(dirname "$example_file")")
    local file_name=$(basename "$example_file")
    local example_name="$dir_name/$file_name"

    # Create temporary working directory
    local temp_dir
    temp_dir=$(mktemp -d -t "bcm-test-${dir_name}-XXXXX")

    # Copy single example file to temp directory
    cp "$example_file" "$temp_dir/"

    # Inject provider configuration for examples that don't include provider blocks
    # Provider reads authentication from environment variables (BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD)
    if ! grep -q "provider \"bcm\"" "$temp_dir"/*.tf 2>/dev/null; then
        cat > "$temp_dir/_provider.tf" <<EOF
terraform {
  required_version = ">= 1.5.0"
  required_providers {
    bcm = {
      source  = "hashicorp/bcm"
      version = "~> 0.1"
    }
  }
}

provider "bcm" {
  insecure_skip_verify = true
}
EOF
    fi

    cd "$temp_dir"

    local test_passed=true
    local error_output=""

    # Run terraform init (skip if using dev_overrides)
    if [ "$USE_DEV_OVERRIDES" != true ]; then
        if ! error_output=$(terraform init -backend=false 2>&1); then
            echo "FAIL|$example_name|terraform init failed: $error_output" > "$result_file"
            test_passed=false
        fi
    fi

    # Run terraform validate
    if [ "$test_passed" = true ]; then
        if ! error_output=$(terraform validate 2>&1); then
            echo "FAIL|$example_name|terraform validate failed: $error_output" > "$result_file"
            test_passed=false
        fi
    fi

    # Run terraform plan
    if [ "$test_passed" = true ]; then
        if ! error_output=$(terraform plan 2>&1); then
            if echo "$error_output" | grep -q "context deadline exceeded\|connection refused\|Unable to Create BCM Client"; then
                echo "SKIP|$example_name|BCM connectivity required for plan" > "$result_file"
                test_passed=true
            else
                echo "FAIL|$example_name|terraform plan failed: $error_output" > "$result_file"
                test_passed=false
            fi
        fi
    fi

    # Cleanup temp directory
    cd /workspace
    rm -rf "$temp_dir"

    if [ "$test_passed" = true ]; then
        echo "PASS|$example_name" > "$result_file"
    fi
}

run_tests() {
    log_info "Phase 3: Testing examples..."

    # T022: Discover and categorize examples
    local discovery_result
    discovery_result=$(discover_examples)

    # Parse data sources and resources
    local data_sources_list="${discovery_result%% | *}"
    local resources_list="${discovery_result##* | }"

    local data_source_examples=()
    local resource_examples=()

    if [ -n "$data_sources_list" ] && [ "$data_sources_list" != "" ]; then
        read -ra data_source_examples <<< "$data_sources_list"
    fi

    if [ -n "$resources_list" ] && [ "$resources_list" != "" ]; then
        read -ra resource_examples <<< "$resources_list"
    fi

    # Apply filters based on flags
    if [ "$RESOURCES_ONLY" = true ]; then
        data_source_examples=()
        log_debug "Execution mode: resources-only (skipping data sources)"
    fi

    if [ "$DATA_SOURCES_ONLY" = true ]; then
        resource_examples=()
        log_debug "Execution mode: data-sources-only (skipping resources)"
    fi

    local total_examples=$((${#data_source_examples[@]} + ${#resource_examples[@]}))
    log_info "Found $total_examples example(s) to test"
    log_info "  Data sources: ${#data_source_examples[@]} (parallel)"
    log_info "  Resources: ${#resource_examples[@]} (sequential)"
    log_info ""

    # T023 & T024: Execute data sources in parallel with PARALLEL_LIMIT
    if [ ${#data_source_examples[@]} -gt 0 ]; then
        local ds_start
        ds_start=$(date +%s)

        log_info "Testing data sources (parallel limit: $PARALLEL_LIMIT)..."

        local pids=()
        local result_files=()
        local active_jobs=0
        local ds_index=0

        for example_dir in "${data_source_examples[@]}"; do
            ds_index=$((ds_index + 1))

            # Wait if we've reached the parallel limit
            while [ $active_jobs -ge "$PARALLEL_LIMIT" ]; do
                # Check if any job has completed
                for i in "${!pids[@]}"; do
                    if ! kill -0 "${pids[$i]}" 2>/dev/null; then
                        # Job completed, remove from tracking
                        unset "pids[$i]"
                        active_jobs=$((active_jobs - 1))
                    fi
                done
                sleep 0.1
            done

            # Start new parallel test
            local result_file
            result_file=$(mktemp -t "bcm-result-XXXXX")
            result_files+=("$result_file")

            test_example_parallel "$example_dir" "$result_file" &
            pids+=($!)
            active_jobs=$((active_jobs + 1))

            local example_name
            example_name=$(basename "$(dirname "$example_dir")")/$(basename "$example_dir")
            log_info "[$ds_index/${#data_source_examples[@]}] → $example_name [PID: $!]"
        done

        # T024: Wait for all parallel tests to complete and aggregate results
        log_info "Waiting for parallel tests to complete..."
        for pid in "${pids[@]}"; do
            wait "$pid" 2>/dev/null || true
        done

        # Process results
        for result_file in "${result_files[@]}"; do
            if [ -f "$result_file" ]; then
                local result
                result=$(cat "$result_file")
                local status="${result%%|*}"
                local message="${result#*|}"

                case "$status" in
                    PASS)
                        log_pass "[PASS] ✓ $message"
                        PASSED_COUNT=$((PASSED_COUNT + 1))
                        ;;
                    SKIP)
                        log_info "  ~ $message"
                        PASSED_COUNT=$((PASSED_COUNT + 1))
                        ;;
                    FAIL)
                        local example_name="${message%%|*}"
                        log_fail "[FAIL] ✗ $example_name"
                        FAILED_COUNT=$((FAILED_COUNT + 1))
                        FAILED_EXAMPLES+=("$example_name")
                        ;;
                esac
                rm -f "$result_file"
            fi
        done

        local ds_end
        ds_end=$(date +%s)
        DATA_SOURCES_TIME=$((ds_end - ds_start))
        log_debug "Data sources completed in ${DATA_SOURCES_TIME}s"
        log_info ""
    fi

    # Execute resources sequentially (unchanged)
    if [ ${#resource_examples[@]} -gt 0 ]; then
        local res_start
        res_start=$(date +%s)

        log_info "Testing resources (sequential)..."
        local res_index=0
        for example_dir in "${resource_examples[@]}"; do
            res_index=$((res_index + 1))
            test_example "$example_dir" "$res_index" "${#resource_examples[@]}" || true
        done

        local res_end
        res_end=$(date +%s)
        RESOURCES_TIME=$((res_end - res_start))
        log_debug "Resources completed in ${RESOURCES_TIME}s"
        log_info ""
    fi
}

#############################################################################
# T019 & T027-T033: Comprehensive Cleanup with Retry Logic
#############################################################################

# T027: Exponential backoff retry logic
cleanup_with_retry() {
    local operation="$1"
    local description="$2"
    local retry_count=0
    local wait_time=1

    while [ $retry_count -lt "$CLEANUP_RETRIES" ]; do
        if eval "$operation" 2>/dev/null; then
            return 0
        fi

        retry_count=$((retry_count + 1))
        if [ $retry_count -lt "$CLEANUP_RETRIES" ]; then
            log_info "  Retry $retry_count/$CLEANUP_RETRIES for $description after ${wait_time}s..."
            sleep $wait_time
            wait_time=$((wait_time * 2))
        fi
    done

    log_fail "  Failed to cleanup $description after $CLEANUP_RETRIES attempts"
    return 1
}

# T028: BCM API authentication and client setup
bcm_login() {
    local cookie_file="$1"
    local login_data
    login_data=$(cat <<EOF
{
  "service": "login",
  "username": "$BCM_USERNAME",
  "password": "$BCM_PASSWORD"
}
EOF
)

    if curl -k -s -c "$cookie_file" -X POST "${BCM_ENDPOINT}/json" \
        -H "Content-Type: application/json" \
        -d "$login_data" > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# T028: Query BCM API for test resources
query_test_resources() {
    local cookie_file="$1"
    local service="$2"
    local method="$3"
    local query_data
    query_data=$(cat <<EOF
{
  "service": "$service",
  "call": "$method"
}
EOF
)

    curl -k -s -b "$cookie_file" -X POST "${BCM_ENDPOINT}/json" \
        -H "Content-Type: application/json" \
        -d "$query_data"
}

# T028: Delete resources via BCM API
delete_bcm_resources() {
    local cookie_file="$1"
    local service="$2"
    local method="$3"
    local resources_json="$4"
    local delete_data
    delete_data=$(cat <<EOF
{
  "service": "$service",
  "call": "$method",
  "args": [$resources_json, false]
}
EOF
)

    if curl -k -s -b "$cookie_file" -X POST "${BCM_ENDPOINT}/json" \
        -H "Content-Type: application/json" \
        -d "$delete_data" > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# T029: Verify resources are deleted
verify_deletion() {
    local cookie_file="$1"
    local service="$2"
    local method="$3"
    local prefix="$4"

    local response
    response=$(query_test_resources "$cookie_file" "$service" "$method")

    # Check if any resources with the prefix still exist
    if echo "$response" | grep -q "\"name\":\"$prefix"; then
        return 1  # Resources still exist
    else
        return 0  # Resources deleted
    fi
}

# T028-T029: Main cleanup implementation with BCM API integration
cleanup_resources() {
    local cleanup_start
    cleanup_start=$(date +%s)

    log_info "Phase 4: Cleanup..."
    log_debug "Cleanup retries: $CLEANUP_RETRIES"

    # Create temporary cookie file for BCM authentication
    local cookie_file
    cookie_file=$(mktemp -t "bcm-cookies-XXXXX")
    log_debug "Cookie file: $cookie_file"

    # Authenticate with BCM
    log_info "Authenticating with BCM API..."
    log_debug "Endpoint: $BCM_ENDPOINT"
    if ! cleanup_with_retry "bcm_login '$cookie_file'" "BCM authentication"; then
        log_error "Failed to authenticate with BCM API"
        log_error "Context: Phase 4 - Cleanup authentication"
        log_error "Endpoint: $BCM_ENDPOINT"
        log_error "Suggestion: Check BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD"
        log_error "Verify BCM API is accessible and credentials are valid"
        rm -f "$cookie_file"
        return 1
    fi
    log_info "✓ Authenticated with BCM API"

    local cleanup_success=0
    local cleanup_failed=0

    # Cleanup software images (from resource examples)
    log_info "Cleaning up test software images..."
    local images_response
    images_response=$(query_test_resources "$cookie_file" "CMPart" "getSoftwareImages")

    # Extract test images (prefix: citest-)
    local test_images=()
    while IFS= read -r line; do
        if [[ "$line" =~ \"name\":\"(citest-[^\"]+)\" ]]; then
            test_images+=("${BASH_REMATCH[1]}")
        fi
    done <<< "$images_response"

    if [ ${#test_images[@]} -gt 0 ]; then
        log_info "Found ${#test_images[@]} test image(s) to cleanup"

        # Build JSON array of image names
        local images_json=""
        for img in "${test_images[@]}"; do
            if [ -z "$images_json" ]; then
                images_json="\"$img\""
            else
                images_json="$images_json, \"$img\""
            fi
        done
        images_json="[$images_json]"

        # Delete with retry
        if cleanup_with_retry "delete_bcm_resources '$cookie_file' 'CMPart' 'removeSoftwareImages' '$images_json'" "software images"; then
            # Verify deletion
            sleep 2  # Wait for eventual consistency
            if verify_deletion "$cookie_file" "CMPart" "getSoftwareImages" "citest-"; then
                log_pass "  ✓ Deleted ${#test_images[@]} software image(s)"
                cleanup_success=$((cleanup_success + ${#test_images[@]}))
            else
                log_fail "  ✗ Deletion verification failed for software images"
                cleanup_failed=$((cleanup_failed + ${#test_images[@]}))
            fi
        else
            log_fail "  ✗ Failed to delete software images"
            cleanup_failed=$((cleanup_failed + ${#test_images[@]}))
        fi
    else
        log_info "No test software images found"
    fi

    # Cleanup categories (from resource examples)
    log_info "Cleaning up test categories..."
    local categories_response
    categories_response=$(query_test_resources "$cookie_file" "CMDevice" "getCategories")

    # Extract test categories (prefix: citest-)
    local test_categories=()
    while IFS= read -r line; do
        if [[ "$line" =~ \"name\":\"(citest-[^\"]+)\" ]]; then
            # Extract UUID for categories (needed for deletion)
            if [[ "$line" =~ \"uuid\":\"([^\"]+)\" ]]; then
                test_categories+=("${BASH_REMATCH[1]}")
            fi
        fi
    done <<< "$categories_response"

    if [ ${#test_categories[@]} -gt 0 ]; then
        log_info "Found ${#test_categories[@]} test categor(ies) to cleanup"

        # Build JSON array of category UUIDs
        local categories_json=""
        for cat in "${test_categories[@]}"; do
            if [ -z "$categories_json" ]; then
                categories_json="\"$cat\""
            else
                categories_json="$categories_json, \"$cat\""
            fi
        done
        categories_json="[$categories_json]"

        # Delete with retry
        if cleanup_with_retry "delete_bcm_resources '$cookie_file' 'CMDevice' 'removeCategories' '$categories_json'" "categories"; then
            # Verify deletion
            sleep 2  # Wait for eventual consistency
            if verify_deletion "$cookie_file" "CMDevice" "getCategories" "citest-"; then
                log_pass "  ✓ Deleted ${#test_categories[@]} categor(ies)"
                cleanup_success=$((cleanup_success + ${#test_categories[@]}))
            else
                log_fail "  ✗ Deletion verification failed for categories"
                cleanup_failed=$((cleanup_failed + ${#test_categories[@]}))
            fi
        else
            log_fail "  ✗ Failed to delete categories"
            cleanup_failed=$((cleanup_failed + ${#test_categories[@]}))
        fi
    else
        log_info "No test categories found"
    fi

    # Cleanup cookie file
    rm -f "$cookie_file"

    local cleanup_end
    cleanup_end=$(date +%s)
    CLEANUP_TIME=$((cleanup_end - cleanup_start))
    CLEANUP_RESOURCE_COUNT=$cleanup_success

    # Report cleanup summary
    log_info ""
    if [ $cleanup_failed -eq 0 ]; then
        log_pass "Cleanup complete: $cleanup_success resource(s) removed, 0 failed (${CLEANUP_TIME}s)"
    else
        log_fail "Cleanup incomplete: $cleanup_success succeeded, $cleanup_failed failed (${CLEANUP_TIME}s)"
        log_error "Context: Phase 4 - Resource cleanup"
        log_error "Suggestion: Run './scripts/test-examples.sh --cleanup-only' to retry cleanup"
    fi
    log_debug "Cleanup completed in ${CLEANUP_TIME}s"
    log_info ""

    # Return success if cleanup succeeded or if there were no resources to clean
    if [ $cleanup_failed -eq 0 ]; then
        return 0
    else
        return 1
    fi
}

#############################################################################
# T020: Exit Code Handling
#############################################################################

print_summary() {
    local end_time
    end_time=$(date +%s)
    local total_time=$((end_time - START_TIME))

    log_info "========================================"
    log_info "TEST SUMMARY"
    log_info "========================================"

    # Test results
    local total_tests=$((PASSED_COUNT + FAILED_COUNT))
    log_info "Examples tested: $total_tests"

    if [ $FAILED_COUNT -eq 0 ]; then
        log_pass "  Passed: $PASSED_COUNT"
        log_pass "  Failed: 0"
    else
        log_info "  Passed: $PASSED_COUNT"
        log_fail "  Failed: $FAILED_COUNT"
    fi
    log_info ""

    # Failed examples detail
    if [ $FAILED_COUNT -gt 0 ]; then
        log_info "Failed examples:"
        for example in "${FAILED_EXAMPLES[@]}"; do
            log_fail "  ✗ $example"
        done
        log_info ""
    fi

    # Timing breakdown
    log_info "Timing:"
    if [ $BUILD_TIME -gt 0 ]; then
        log_info "  Build: ${BUILD_TIME}s"
    fi
    if [ $DATA_SOURCES_TIME -gt 0 ]; then
        log_info "  Data sources: ${DATA_SOURCES_TIME}s (parallel)"
    fi
    if [ $RESOURCES_TIME -gt 0 ]; then
        log_info "  Resources: ${RESOURCES_TIME}s (sequential)"
    fi
    if [ $CLEANUP_TIME -gt 0 ]; then
        log_info "  Cleanup: ${CLEANUP_TIME}s"
    fi
    log_info "  Total: ${total_time}s"
    log_info ""

    # Cleanup summary
    if [ $CLEANUP_RESOURCE_COUNT -gt 0 ]; then
        log_info "Cleanup:"
        log_info "  Resources removed: $CLEANUP_RESOURCE_COUNT"
        log_info ""
    fi

    log_info "========================================"

    # Final verdict
    if [ $FAILED_COUNT -eq 0 ]; then
        log_pass "RESULT: All tests passed ✓"
    else
        log_fail "RESULT: $FAILED_COUNT test(s) failed ✗"
        log_info ""
        log_info "To debug failures:"
        log_info "  ./scripts/test-examples.sh --verbose --no-cleanup"
    fi

    log_info "========================================"
}

#############################################################################
# Main Execution
#############################################################################

main() {
    # Track start time for total execution time
    START_TIME=$(date +%s)

    # T015: Validate environment
    validate_environment

    # T030: Check if cleanup-only mode
    if [ "$CLEANUP_ONLY" = true ]; then
        log_info "Running in cleanup-only mode (skipping build and tests)"
        log_info ""

        # Only run cleanup phase
        if cleanup_resources; then
            log_pass "Cleanup-only mode completed successfully"
            exit "$EXIT_SUCCESS"
        else
            log_fail "Cleanup-only mode failed"
            exit "$EXIT_TEST_FAILURE"
        fi
    fi

    # T016: Build provider
    build_provider

    # T017 & T018: Discover and test examples
    run_tests

    # T019 & T027-T033: Comprehensive cleanup with retry
    if [ "$NO_CLEANUP" = true ]; then
        log_info "Phase 4: Skipping cleanup (--no-cleanup flag set)"
        log_info ""
        log_info "Note: Test resources remain in BCM for debugging"
        log_info "To cleanup later: ./scripts/test-examples.sh --cleanup-only"
        log_info ""
    else
        cleanup_resources
    fi

    # T020: Print summary and exit with appropriate code
    print_summary

    if [ $FAILED_COUNT -eq 0 ]; then
        exit "$EXIT_SUCCESS"
    else
        exit "$EXIT_TEST_FAILURE"
    fi
}

#############################################################################
# T021: Execute main function
#############################################################################

main "$@"
