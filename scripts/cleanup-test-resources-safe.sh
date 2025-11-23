#!/usr/bin/env bash
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Safe cleanup script with rate limiting to prevent BCM overload
# Includes delays between deletions, health checks, and dry-run mode

set -euo pipefail

# Configuration
: ${DELAY_MS:=500}  # Delay between deletions (milliseconds)
: ${DRY_RUN:=0}     # Set to 1 for dry-run mode
: ${DELETE_ONLY:=""}  # Optional: "images", "categories", "devices", "networks", "clusters"

# Check environment variables
if [ -z "${BCM_ENDPOINT:-}" ] || [ -z "${BCM_USERNAME:-}" ] || [ -z "${BCM_PASSWORD:-}" ]; then
    echo "Error: BCM_ENDPOINT, BCM_USERNAME, and BCM_PASSWORD must be set"
    exit 1
fi

# Safety check (unless dry run)
if [ "$DRY_RUN" != "1" ] && [ "${AUTO_CLEANUP:-}" != "1" ]; then
    echo "Error: AUTO_CLEANUP environment variable must be set to 1 to run cleanup"
    echo "Or use DRY_RUN=1 to see what would be deleted"
    echo "Usage: AUTO_CLEANUP=1 $0"
    exit 1
fi

if [ "$DRY_RUN" == "1" ]; then
    echo "=========================================="
    echo "DRY RUN MODE - No deletions will occur"
    echo "=========================================="
fi

echo "=========================================="
echo "Safe BCM Test Resource Cleanup"
echo "Prefixes: citest-*, tftest-*"
echo "Rate limit: ${DELAY_MS}ms between deletions"
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

if [ -z "$LOGIN_RESPONSE" ] || [ "$LOGIN_RESPONSE" != "true" ]; then
    echo "✗ Login failed"
    exit 1
fi
echo "✓ Logged in successfully"
echo ""

# Function to check BCM health
check_bcm_health() {
    local test_resp=$(curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
        -H "Content-Type: application/json" \
        -d '{"service":"CMPart","call":"getPartitions"}' 2>&1)

    if echo "$test_resp" | jq . >/dev/null 2>&1; then
        return 0
    else
        echo "⚠ BCM health check failed, waiting..."
        sleep 2
        return 1
    fi
}

# Function to delete resources with rate limiting
delete_resources_safe() {
    local service=$1
    local resource_type=$2
    local get_method=$3
    local remove_method=$4
    local name_field=$5

    # Skip if DELETE_ONLY is set and doesn't match
    if [ -n "$DELETE_ONLY" ]; then
        case "$resource_type" in
            *"Images"*) [[ "$DELETE_ONLY" != "images" ]] && return 0 ;;
            *"Categories"*) [[ "$DELETE_ONLY" != "categories" ]] && return 0 ;;
            *"Devices"*) [[ "$DELETE_ONLY" != "devices" ]] && return 0 ;;
            *"Networks"*) [[ "$DELETE_ONLY" != "networks" ]] && return 0 ;;
            *"Clusters"*) [[ "$DELETE_ONLY" != "clusters" ]] && return 0 ;;
        esac
    fi

    echo "→ Checking ${resource_type}..."

    # Get resources
    RESOURCES=$(curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
        -H "Content-Type: application/json" \
        -d "{\"service\":\"${service}\",\"call\":\"${get_method}\"}")

    # Check if BCM is healthy
    if ! echo "$RESOURCES" | jq . >/dev/null 2>&1; then
        echo "  ⚠ BCM returned invalid response, skipping ${resource_type}"
        return 1
    fi

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
    echo "$MATCH" | jq -r --arg field "$name_field" '.[] | "    - \(.[$field]) [\(.uuid[0:8])]"'

    if [ "$DRY_RUN" == "1" ]; then
        echo "  [DRY RUN] Would delete $COUNT resources"
        return 0
    fi

    # Extract IDs for deletion
    if [ "$remove_method" == "removeSoftwareImages" ]; then
        # Software images use names
        TO_DELETE=$(echo "$MATCH" | jq -r '[.[].name]')
    else
        # Others use UUIDs
        TO_DELETE=$(echo "$MATCH" | jq -r '[.[].uuid]')
    fi

    # Delete one at a time with delays
    local deleted=0
    echo "$TO_DELETE" | jq -r '.[]' | while read -r id; do
        if [ -n "$id" ]; then
            # Delete single resource
            local id_array
            if [ "$remove_method" == "removeSoftwareImages" ]; then
                id_array="[\"$id\"]"
            else
                id_array="[\"$id\"]"
            fi

            DELETE_RESPONSE=$(curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
                -H "Content-Type: application/json" \
                -d "{\"service\":\"${service}\",\"call\":\"${remove_method}\",\"args\":[${id_array},false]}" 2>&1)

            if echo "$DELETE_RESPONSE" | jq . >/dev/null 2>&1; then
                deleted=$((deleted + 1))
                echo "  ✓ Deleted $deleted/$COUNT"
            else
                echo "  ⚠ Failed to delete: $id"
            fi

            # Rate limiting delay
            if [ "$DELAY_MS" -gt 0 ]; then
                sleep $(echo "scale=3; $DELAY_MS / 1000" | bc)
            fi
        fi
    done

    # Health check after batch
    if ! check_bcm_health; then
        echo "  ⚠ BCM health check failed after deletions"
        return 1
    fi

    echo "  ✓ ${resource_type} cleanup complete"
}

# Delete in dependency order (reverse of creation)
echo "Starting cleanup with ${DELAY_MS}ms delays..."
echo ""

# 1. Devices first (depend on categories)
delete_resources_safe "CMDevice" "Devices" "getNodes" "removeNodes" "hostname"
echo ""

# 2. Kubernetes clusters (independent)
delete_resources_safe "CMKube" "Kubernetes Clusters" "getClusters" "removeClusters" "name"
echo ""

# 3. Networks (independent)
delete_resources_safe "CMNet" "Networks" "getNetworks" "removeNetworks" "name"
echo ""

# 4. Categories (depend on software images)
delete_resources_safe "CMDevice" "Categories" "getCategories" "removeCategories" "name"
echo ""

# 5. Software images last (no dependencies)
delete_resources_safe "CMPart" "Software Images" "getSoftwareImages" "removeSoftwareImages" "name"
echo ""

# Final health check
echo "→ Final BCM health check..."
if check_bcm_health; then
    echo "✓ BCM is healthy"
else
    echo "⚠ BCM health check failed - cluster may need time to stabilize"
    exit 1
fi

echo ""
echo "=========================================="
if [ "$DRY_RUN" == "1" ]; then
    echo "✓ Dry run complete - no changes made"
else
    echo "✓ Safe cleanup complete!"
fi
echo "=========================================="
