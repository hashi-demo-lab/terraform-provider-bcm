#!/usr/bin/env bash
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Pre-test cleanup: Ensure no orphaned test resources exist before running tests
# This guarantees test name uniqueness by removing any conflicting resources

set -euo pipefail

# Check environment variables
if [ -z "${BCM_ENDPOINT:-}" ] || [ -z "${BCM_USERNAME:-}" ] || [ -z "${BCM_PASSWORD:-}" ]; then
    echo "Error: BCM_ENDPOINT, BCM_USERNAME, and BCM_PASSWORD must be set"
    exit 1
fi

echo "=========================================="
echo "Pre-Test Cleanup: Removing orphaned test resources"
echo "=========================================="
echo ""

# Create cookie file
COOKIE_FILE=$(mktemp)
trap "rm -f $COOKIE_FILE" EXIT

# Login to BCM
echo "→ Logging in to BCM..."
curl -k -s -c "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
    -H "Content-Type: application/json" \
    -d "{\"service\":\"login\",\"username\":\"$BCM_USERNAME\",\"password\":\"$BCM_PASSWORD\"}" > /dev/null
echo "✓ Logged in"
echo ""

# Function to count and delete resources silently
cleanup_resource_type() {
    local service=$1
    local resource_type=$2
    local get_method=$3
    local remove_method=$4
    local name_field=$5

    # Get resources
    RESOURCES=$(curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
        -H "Content-Type: application/json" \
        -d "{\"service\":\"${service}\",\"call\":\"${get_method}\"}")

    # Count matching resources
    if [ "$name_field" == "name" ]; then
        COUNT=$(echo "$RESOURCES" | jq -r "[.[] | select(.name | startswith(\"citest-\") or startswith(\"tftest-\"))] | length")
    else
        COUNT=$(echo "$RESOURCES" | jq -r "[.[] | select(.${name_field} | startswith(\"citest-\") or startswith(\"tftest-\"))] | length")
    fi

    if [ "$COUNT" -eq 0 ]; then
        echo "  ✓ No orphaned ${resource_type}"
        return 0
    fi

    echo "  ⚠ Found $COUNT orphaned ${resource_type} - cleaning up..."

    # Extract IDs for deletion
    if [ "$remove_method" == "removeSoftwareImages" ]; then
        TO_DELETE=$(echo "$RESOURCES" | jq -r "[.[] | select(.${name_field} | startswith(\"citest-\") or startswith(\"tftest-\")) | .name] | @json")
    else
        TO_DELETE=$(echo "$RESOURCES" | jq -r "[.[] | select(.${name_field} | startswith(\"citest-\") or startswith(\"tftest-\")) | .uuid] | @json")
    fi

    if [ "$TO_DELETE" != "[]" ]; then
        curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
            -H "Content-Type: application/json" \
            -d "{\"service\":\"${service}\",\"call\":\"${remove_method}\",\"args\":[${TO_DELETE},false]}" > /dev/null
        echo "  ✓ Deleted $COUNT ${resource_type}"
    fi
}

# Clean up in dependency order
echo "→ Checking for orphaned resources..."
cleanup_resource_type "CMDevice" "devices" "getNodes" "removeNodes" "hostname"
cleanup_resource_type "CMKube" "Kubernetes clusters" "getClusters" "removeClusters" "name"
cleanup_resource_type "CMNet" "networks" "getNetworks" "removeNetworks" "name"
cleanup_resource_type "CMDevice" "categories" "getCategories" "removeCategories" "name"
cleanup_resource_type "CMPart" "software images" "getSoftwareImages" "removeSoftwareImages" "name"

echo ""
echo "=========================================="
echo "✓ Pre-test cleanup complete"
echo "  Tests can now run with guaranteed name uniqueness"
echo "=========================================="
