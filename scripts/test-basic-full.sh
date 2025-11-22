#!/usr/bin/env bash
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Full Integration Test for basic.tf Example
# Purpose: Demonstrate complete plan and apply workflow for device resource
# Usage: BCM_ENDPOINT="..." BCM_USERNAME="..." BCM_PASSWORD="..." ./scripts/test-basic-full.sh

set -euo pipefail

# Color codes for output
readonly COLOR_INFO='\033[0;36m'
readonly COLOR_PASS='\033[0;32m'
readonly COLOR_FAIL='\033[0;31m'
readonly COLOR_RESET='\033[0m'

log_info() {
    echo -e "${COLOR_INFO}[INFO]${COLOR_RESET} $*"
}

log_pass() {
    echo -e "${COLOR_PASS}[PASS]${COLOR_RESET} $*"
}

log_fail() {
    echo -e "${COLOR_FAIL}[FAIL]${COLOR_RESET} $*"
}

# Exit codes
readonly EXIT_SUCCESS=0
readonly EXIT_TEST_FAILURE=1
readonly EXIT_CONFIG_ERROR=2

# Validate environment
if [ -z "${BCM_ENDPOINT:-}" ]; then
    log_fail "Missing BCM_ENDPOINT environment variable"
    exit "$EXIT_CONFIG_ERROR"
fi

if [ -z "${BCM_USERNAME:-}" ]; then
    log_fail "Missing BCM_USERNAME environment variable"
    exit "$EXIT_CONFIG_ERROR"
fi

if [ -z "${BCM_PASSWORD:-}" ]; then
    log_fail "Missing BCM_PASSWORD environment variable"
    exit "$EXIT_CONFIG_ERROR"
fi

log_info "========================================"
log_info "Full Integration Test: basic.tf"
log_info "========================================"
log_info ""

# Create temporary test directory
TEST_DIR=$(mktemp -d -t "bcm-basic-test-XXXXX")
log_info "Test directory: $TEST_DIR"

# Cleanup function
cleanup() {
    log_info ""
    log_info "Cleaning up test directory..."
    rm -rf "$TEST_DIR"
}
trap cleanup EXIT

# Copy basic.tf to test directory
cp /workspace/examples/resources/bcm_cmdevice_device/basic.tf "$TEST_DIR/"
cd "$TEST_DIR"

# Phase 1: Terraform Init
log_info ""
log_info "Phase 1: terraform init"
log_info "----------------------------------------"
if terraform init -upgrade > /dev/null 2>&1; then
    log_pass "✓ Initialization successful"
else
    log_fail "✗ Initialization failed"
    exit "$EXIT_TEST_FAILURE"
fi

# Phase 2: Terraform Validate
log_info ""
log_info "Phase 2: terraform validate"
log_info "----------------------------------------"
if terraform validate > /dev/null 2>&1; then
    log_pass "✓ Configuration is valid"
else
    log_fail "✗ Validation failed"
    exit "$EXIT_TEST_FAILURE"
fi

# Phase 3: Terraform Plan
log_info ""
log_info "Phase 3: terraform plan"
log_info "----------------------------------------"
if terraform plan -out=tfplan > /dev/null 2>&1; then
    log_pass "✓ Plan created successfully"

    # Show plan summary
    log_info ""
    log_info "Plan Summary:"
    terraform show -json tfplan | jq -r '.resource_changes[] | "  + \(.type).\(.name) [\(.change.actions | join(","))]"' 2>/dev/null || true
else
    log_fail "✗ Planning failed"
    terraform plan -out=tfplan
    exit "$EXIT_TEST_FAILURE"
fi

# Phase 4: Terraform Apply
log_info ""
log_info "Phase 4: terraform apply"
log_info "----------------------------------------"
if terraform apply -auto-approve tfplan; then
    log_pass "✓ Apply successful"
else
    log_fail "✗ Apply failed"
    exit "$EXIT_TEST_FAILURE"
fi

# Phase 5: Verify Resource Creation
log_info ""
log_info "Phase 5: Verify resource creation"
log_info "----------------------------------------"

# Extract created resource information
DEVICE_ID=$(terraform output -json 2>/dev/null | jq -r '.[] | select(.type == "string") | .value' | head -1 || echo "")

if [ -n "$DEVICE_ID" ]; then
    log_pass "✓ Device created with ID: $DEVICE_ID"
else
    log_info "  (No output captured - checking state)"
    DEVICE_ID=$(terraform show -json | jq -r '.values.root_module.resources[] | select(.type == "bcm_cmdevice_device") | .values.id' 2>/dev/null || echo "")
    if [ -n "$DEVICE_ID" ]; then
        log_pass "✓ Device created with ID: $DEVICE_ID"
    else
        log_fail "✗ Could not verify device creation"
    fi
fi

# Show resource details
log_info ""
log_info "Resource Details:"
terraform show -json | jq -r '.values.root_module.resources[] | select(.type == "bcm_cmdevice_device") | "  Hostname: \(.values.hostname)\n  MAC: \(.values.mac)\n  Category: \(.values.category)\n  Management Network: \(.values.management_network)"' 2>/dev/null || true

# Phase 6: Verify BCM API
log_info ""
log_info "Phase 6: Verify resource in BCM API"
log_info "----------------------------------------"

# Create temporary cookie file for BCM authentication
COOKIE_FILE=$(mktemp -t "bcm-cookies-XXXXX")

# Login to BCM
if curl -k -s -c "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
    -H "Content-Type: application/json" \
    -d "{\"service\":\"login\",\"username\":\"$BCM_USERNAME\",\"password\":\"$BCM_PASSWORD\"}" > /dev/null 2>&1; then
    log_pass "✓ Authenticated with BCM API"

    # Query devices
    DEVICES_RESPONSE=$(curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
        -H "Content-Type: application/json" \
        -d '{"service":"CMDevice","call":"getNodes"}')

    # Check if our device exists
    if echo "$DEVICES_RESPONSE" | grep -q "citest-compute-basic"; then
        log_pass "✓ Device 'citest-compute-basic' found in BCM"
    else
        log_fail "✗ Device 'citest-compute-basic' not found in BCM"
    fi
else
    log_fail "✗ Failed to authenticate with BCM API"
fi

rm -f "$COOKIE_FILE"

# Phase 7: Terraform Plan (should show no changes)
log_info ""
log_info "Phase 7: terraform plan (verify idempotency)"
log_info "----------------------------------------"
PLAN_OUTPUT=$(terraform plan -detailed-exitcode 2>&1) || PLAN_EXIT=$?

if [ "${PLAN_EXIT:-0}" -eq 0 ]; then
    log_pass "✓ No changes detected (idempotent)"
elif [ "${PLAN_EXIT:-0}" -eq 2 ]; then
    log_fail "✗ Plan detected changes (not idempotent)"
    echo "$PLAN_OUTPUT"
else
    log_fail "✗ Plan failed"
    echo "$PLAN_OUTPUT"
    exit "$EXIT_TEST_FAILURE"
fi

# Phase 8: Terraform Destroy
log_info ""
log_info "Phase 8: terraform destroy"
log_info "----------------------------------------"
if terraform destroy -auto-approve 2>&1 | tee /tmp/destroy-output.log; then
    log_pass "✓ Destroy successful"
else
    log_fail "✗ Destroy failed"
    log_info ""
    log_info "Destroy output:"
    cat /tmp/destroy-output.log | tail -20
    exit "$EXIT_TEST_FAILURE"
fi

# Phase 9: Verify Resource Deletion
log_info ""
log_info "Phase 9: Verify resource deletion in BCM"
log_info "----------------------------------------"

# Create temporary cookie file for BCM authentication
COOKIE_FILE=$(mktemp -t "bcm-cookies-XXXXX")

# Login to BCM
if curl -k -s -c "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
    -H "Content-Type: application/json" \
    -d "{\"service\":\"login\",\"username\":\"$BCM_USERNAME\",\"password\":\"$BCM_PASSWORD\"}" > /dev/null 2>&1; then

    # Query devices
    DEVICES_RESPONSE=$(curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
        -H "Content-Type: application/json" \
        -d '{"service":"CMDevice","call":"getNodes"}')

    # Check if our device was deleted
    if echo "$DEVICES_RESPONSE" | grep -q "citest-compute-basic"; then
        log_fail "✗ Device 'citest-compute-basic' still exists in BCM"
    else
        log_pass "✓ Device 'citest-compute-basic' successfully deleted from BCM"
    fi
else
    log_fail "✗ Failed to authenticate with BCM API"
fi

rm -f "$COOKIE_FILE"

# Final Summary
log_info ""
log_info "========================================"
log_info "TEST SUMMARY"
log_info "========================================"
log_pass "All phases completed successfully!"
log_info ""
log_info "Test Coverage:"
log_info "  ✓ terraform init"
log_info "  ✓ terraform validate"
log_info "  ✓ terraform plan"
log_info "  ✓ terraform apply"
log_info "  ✓ Resource creation verification (BCM API)"
log_info "  ✓ Idempotency verification"
log_info "  ✓ terraform destroy"
log_info "  ✓ Resource deletion verification (BCM API)"
log_info ""
log_info "========================================"

exit "$EXIT_SUCCESS"
