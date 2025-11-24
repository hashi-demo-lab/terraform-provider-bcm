#!/usr/bin/env bash
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Cleanup script for test resources (citest-* and tftest-* prefixes)
# FIXED: Deletion order now follows dependency graph to prevent orphaned references

set -euo pipefail

# Support dry-run mode
DRY_RUN=${DRY_RUN:-false}

# Check environment variables
if [ -z "${BCM_ENDPOINT:-}" ] || [ -z "${BCM_USERNAME:-}" ] || [ -z "${BCM_PASSWORD:-}" ]; then
    echo "Error: BCM_ENDPOINT, BCM_USERNAME, and BCM_PASSWORD must be set"
    exit 1
fi

echo "========================================="
echo "BCM CLEANUP SCRIPT"
echo "========================================="
echo "Endpoint: $BCM_ENDPOINT"
echo "Dry-run mode: $DRY_RUN"
echo ""
echo "DELETION ORDER (Dependency-Safe):"
echo "  1. Devices (highest level - delete first)"
echo "  2. Kubernetes Clusters (independent)"
echo "  3. Networks (independent)"
echo "  4. Categories (mid-level)"
echo "  5. Software Images (lowest level - delete last)"
echo "========================================="
echo ""

# Create cookie file
COOKIE_FILE=$(mktemp)
trap "rm -f $COOKIE_FILE" EXIT

# Login to BCM
echo "Logging in to BCM..."
curl -k -s -c "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
    -H "Content-Type: application/json" \
    -d "{\"service\":\"login\",\"username\":\"$BCM_USERNAME\",\"password\":\"$BCM_PASSWORD\"}" > /dev/null

if [ $? -ne 0 ]; then
    echo "ERROR: BCM login failed"
    exit 1
fi

echo "Login successful"
echo ""

# Health check function
check_bcm_health() {
    curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
        -H "Content-Type: application/json" \
        -d '{"service":"cmgui","call":"getSystemStatus"}' > /dev/null 2>&1

    if [ $? -ne 0 ]; then
        echo "ERROR: BCM health check failed"
        return 1
    fi
    return 0
}

# Resource deletion function
delete_resources() {
    local resource_type=$1
    local service=$2
    local get_method=$3
    local remove_method=$4
    local name_field=$5

    echo "Querying $resource_type..."

    # Query resources
    RESOURCES=$(curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
        -H "Content-Type: application/json" \
        -d "{\"service\":\"$service\",\"call\":\"$get_method\"}" | \
        jq -r ".[] | select(.$name_field | startswith(\"citest-\") or startswith(\"tftest-\")) | \"  - \(.$name_field) [\\(.uuid)]\"")

    if [ -z "$RESOURCES" ]; then
        echo "  No $resource_type found to delete"
        echo ""
        return 0
    fi

    echo "Found $resource_type to delete:"
    echo "$RESOURCES"
    echo ""

    # Extract resource identifiers (UUIDs or names depending on remove method)
    if [[ "$remove_method" == "removeSoftwareImages" ]]; then
        # Software images use names, not UUIDs
        RESOURCE_IDS=$(curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
            -H "Content-Type: application/json" \
            -d "{\"service\":\"$service\",\"call\":\"$get_method\"}" | \
            jq -r "[.[] | select(.$name_field | startswith(\"citest-\") or startswith(\"tftest-\")) | .name] | @json")
    else
        # Other resources use UUIDs
        RESOURCE_IDS=$(curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
            -H "Content-Type: application/json" \
            -d "{\"service\":\"$service\",\"call\":\"$get_method\"}" | \
            jq -r "[.[] | select(.$name_field | startswith(\"citest-\") or startswith(\"tftest-\")) | .uuid] | @json")
    fi

    if [ "$RESOURCE_IDS" = "[]" ] || [ -z "$RESOURCE_IDS" ]; then
        echo "  No $resource_type to delete"
        echo ""
        return 0
    fi

    local count=$(echo "$RESOURCE_IDS" | jq -r 'length')

    if [ "$DRY_RUN" = "true" ]; then
        echo "  [DRY-RUN] Would delete $count $resource_type"
        echo ""
        return 0
    fi

    # Confirm deletion (unless in non-interactive mode)
    if [ -t 0 ]; then
        read -p "Delete these $count $resource_type? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "  Skipped deletion of $resource_type"
            echo ""
            return 0
        fi
    fi

    # Delete resources
    echo "  Deleting $count $resource_type..."
    curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
        -H "Content-Type: application/json" \
        -d "{\"service\":\"$service\",\"call\":\"$remove_method\",\"args\":[$RESOURCE_IDS,false]}" > /dev/null

    if [ $? -eq 0 ]; then
        echo "  Successfully deleted $count $resource_type"
    else
        echo "  ERROR: Failed to delete $resource_type"
        return 1
    fi

    echo ""
    return 0
}

# CORRECT DELETION ORDER: Devices → Clusters → Networks → Categories → Images

echo "[1/5] Deleting Devices..."
echo "  (Devices depend on Categories - must delete first)"
delete_resources "devices" "CMDevice" "getNodes" "removeNodes" "hostname"
sleep 2
check_bcm_health || exit 1

echo "[2/5] Deleting Kubernetes Clusters..."
echo "  (Clusters are independent - safe to delete)"
delete_resources "clusters" "CMKube" "getClusters" "removeClusters" "name"
sleep 2
check_bcm_health || exit 1

echo "[3/5] Deleting Networks..."
echo "  (Networks are independent - safe to delete)"
delete_resources "networks" "CMNet" "getNetworks" "removeNetworks" "name"
sleep 2
check_bcm_health || exit 1

echo "[4/5] Deleting Categories..."
echo "  (Categories depend on Software Images - delete before images)"
delete_resources "categories" "CMDevice" "getCategories" "removeCategories" "name"
sleep 2
check_bcm_health || exit 1

echo "[5/5] Deleting Software Images..."
echo "  (Software Images have no dependencies - delete last)"
delete_resources "images" "CMPart" "getSoftwareImages" "removeSoftwareImages" "name"
sleep 2
check_bcm_health || exit 1

echo "========================================="
echo "CLEANUP COMPLETE!"
echo "========================================="
echo "All test resources have been processed in dependency-safe order."
echo "No orphaned references should remain in the BCM database."
echo ""
