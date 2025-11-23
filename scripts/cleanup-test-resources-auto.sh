#!/usr/bin/env bash
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Automated cleanup script for test resources (no confirmation prompts)
# Cleans up all resources with citest-* or tftest-* prefixes
# Use with caution - this will delete resources without asking!

set -euo pipefail

# Check environment variables
if [ -z "${BCM_ENDPOINT:-}" ] || [ -z "${BCM_USERNAME:-}" ] || [ -z "${BCM_PASSWORD:-}" ]; then
    echo "Error: BCM_ENDPOINT, BCM_USERNAME, and BCM_PASSWORD must be set"
    exit 1
fi

# Optional: Set AUTO_CLEANUP=1 to enable (safety check)
if [ "${AUTO_CLEANUP:-}" != "1" ]; then
    echo "Error: AUTO_CLEANUP environment variable must be set to 1 to run automated cleanup"
    echo "Usage: AUTO_CLEANUP=1 $0"
    exit 1
fi

echo "=========================================="
echo "Automated BCM Test Resource Cleanup"
echo "Prefixes: citest-*, tftest-*"
echo "=========================================="
echo ""

# Create cookie file
COOKIE_FILE=$(mktemp)
trap "rm -f $COOKIE_FILE" EXIT

# Login to BCM
echo "→ Logging in to BCM..."
LOGIN_RESPONSE=$(curl -k -s -c "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
    -H "Content-Type: application/json" \
    -d "{\"service\":\"login\",\"username\":\"$BCM_USERNAME\",\"password\":\"$BCM_PASSWORD\"}")

if [ -z "$LOGIN_RESPONSE" ]; then
    echo "✗ Login failed"
    exit 1
fi
echo "✓ Logged in successfully"
echo ""

# Function to delete resources
delete_resources() {
    local service=$1
    local resource_type=$2
    local get_method=$3
    local remove_method=$4
    local name_field=$5

    echo "→ Checking ${resource_type}..."

    # Get resources
    RESOURCES=$(curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
        -H "Content-Type: application/json" \
        -d "{\"service\":\"${service}\",\"call\":\"${get_method}\"}")

    # Extract matching resources
    if [ "$name_field" == "name" ]; then
        MATCH=$(echo "$RESOURCES" | jq -r "[.[] | select(.name | startswith(\"citest-\") or startswith(\"tftest-\"))]")
    else
        MATCH=$(echo "$RESOURCES" | jq -r "[.[] | select(.${name_field} | startswith(\"citest-\") or startswith(\"tftest-\"))]")
    fi

    COUNT=$(echo "$MATCH" | jq -r 'length')

    if [ "$COUNT" -eq 0 ]; then
        echo "  No ${resource_type} to delete"
        return 0
    fi

    echo "  Found $COUNT ${resource_type} to delete:"
    echo "$MATCH" | jq -r --arg field "$name_field" '.[] | "    - \(.[$field]) [\(.uuid)]"'

    # Extract names or UUIDs for deletion
    if [ "$remove_method" == "removeSoftwareImages" ]; then
        # Software images use names
        TO_DELETE=$(echo "$MATCH" | jq -r '[.[].name] | @json')
    else
        # Others use UUIDs
        TO_DELETE=$(echo "$MATCH" | jq -r '[.[].uuid] | @json')
    fi

    if [ "$TO_DELETE" != "[]" ]; then
        DELETE_RESPONSE=$(curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
            -H "Content-Type: application/json" \
            -d "{\"service\":\"${service}\",\"call\":\"${remove_method}\",\"args\":[${TO_DELETE},false]}")

        echo "  ✓ ${resource_type} deleted"
    fi
}

# Delete in dependency order (reverse of creation)
# 1. Devices first (depend on categories)
delete_resources "CMDevice" "Devices" "getNodes" "removeNodes" "hostname"
echo ""

# 2. Kubernetes clusters (independent)
delete_resources "CMKube" "Kubernetes Clusters" "getClusters" "removeClusters" "name"
echo ""

# 3. Networks (independent)
delete_resources "CMNet" "Networks" "getNetworks" "removeNetworks" "name"
echo ""

# 4. Categories (depend on software images)
delete_resources "CMDevice" "Categories" "getCategories" "removeCategories" "name"
echo ""

# 5. Software images last (no dependencies)
delete_resources "CMPart" "Software Images" "getSoftwareImages" "removeSoftwareImages" "name"
echo ""

echo "=========================================="
echo "✓ Automated cleanup complete!"
echo "=========================================="
