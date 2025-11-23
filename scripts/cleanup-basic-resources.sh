#!/usr/bin/env bash
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Cleanup script for test resources (citest-* and tftest-* prefixes)

set -euo pipefail

# Check environment variables
if [ -z "${BCM_ENDPOINT:-}" ] || [ -z "${BCM_USERNAME:-}" ] || [ -z "${BCM_PASSWORD:-}" ]; then
    echo "Error: BCM_ENDPOINT, BCM_USERNAME, and BCM_PASSWORD must be set"
    exit 1
fi

echo "Cleaning up BCM resources for basic.tf test..."

# Create cookie file
COOKIE_FILE=$(mktemp)
trap "rm -f $COOKIE_FILE" EXIT

# Login to BCM
echo "Logging in to BCM..."
curl -k -s -c "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
    -H "Content-Type: application/json" \
    -d "{\"service\":\"login\",\"username\":\"$BCM_USERNAME\",\"password\":\"$BCM_PASSWORD\"}" > /dev/null

# List and delete software images
echo ""
echo "Software Images to delete:"
curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
    -H "Content-Type: application/json" \
    -d '{"service":"CMPart","call":"getSoftwareImages"}' | \
    jq -r '.[] | select(.name | startswith("citest-") or startswith("tftest-")) | "  - \(.name) [\(.uuid)]"'

echo ""
read -p "Delete these software images? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    IMAGES=$(curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
        -H "Content-Type: application/json" \
        -d '{"service":"CMPart","call":"getSoftwareImages"}' | \
        jq -r '[.[] | select(.name | startswith("citest-") or startswith("tftest-")) | .name] | @json')

    if [ "$IMAGES" != "[]" ]; then
        curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
            -H "Content-Type: application/json" \
            -d "{\"service\":\"CMPart\",\"call\":\"removeSoftwareImages\",\"args\":[$IMAGES,false]}"
        echo "Software images deleted"
    fi
fi

# List and delete categories
echo ""
echo "Categories to delete:"
curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
    -H "Content-Type: application/json" \
    -d '{"service":"CMDevice","call":"getCategories"}' | \
    jq -r '.[] | select(.name | startswith("citest-") or startswith("tftest-")) | "  - \(.name) [\(.uuid)]"'

echo ""
read -p "Delete these categories? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    CATEGORIES=$(curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
        -H "Content-Type: application/json" \
        -d '{"service":"CMDevice","call":"getCategories"}' | \
        jq -r '[.[] | select(.name | startswith("citest-") or startswith("tftest-")) | .uuid] | @json')

    if [ "$CATEGORIES" != "[]" ]; then
        curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
            -H "Content-Type: application/json" \
            -d "{\"service\":\"CMDevice\",\"call\":\"removeCategories\",\"args\":[$CATEGORIES,false]}"
        echo "Categories deleted"
    fi
fi

# List and delete devices
echo ""
echo "Devices to delete:"
curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
    -H "Content-Type: application/json" \
    -d '{"service":"CMDevice","call":"getNodes"}' | \
    jq -r '.[] | select(.hostname | startswith("citest-") or startswith("tftest-")) | "  - \(.hostname) [\(.uuid)]"'

echo ""
read -p "Delete these devices? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    DEVICES=$(curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
        -H "Content-Type: application/json" \
        -d '{"service":"CMDevice","call":"getNodes"}' | \
        jq -r '[.[] | select(.hostname | startswith("citest-") or startswith("tftest-")) | .uuid] | @json')

    if [ "$DEVICES" != "[]" ]; then
        curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
            -H "Content-Type: application/json" \
            -d "{\"service\":\"CMDevice\",\"call\":\"removeNodes\",\"args\":[$DEVICES,false]}"
        echo "Devices deleted"
    fi
fi

# List and delete Kubernetes clusters
echo ""
echo "Kubernetes Clusters to delete:"
curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
    -H "Content-Type: application/json" \
    -d '{"service":"CMKube","call":"getClusters"}' | \
    jq -r '.[] | select(.name | startswith("citest-") or startswith("tftest-")) | "  - \(.name) [\(.uuid)]"'

echo ""
read -p "Delete these clusters? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    CLUSTERS=$(curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
        -H "Content-Type: application/json" \
        -d '{"service":"CMKube","call":"getClusters"}' | \
        jq -r '[.[] | select(.name | startswith("citest-") or startswith("tftest-")) | .uuid] | @json')

    if [ "$CLUSTERS" != "[]" ]; then
        curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
            -H "Content-Type: application/json" \
            -d "{\"service\":\"CMKube\",\"call\":\"removeClusters\",\"args\":[$CLUSTERS,false]}"
        echo "Kubernetes clusters deleted"
    fi
fi

# List and delete networks
echo ""
echo "Networks to delete:"
curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
    -H "Content-Type: application/json" \
    -d '{"service":"CMNet","call":"getNetworks"}' | \
    jq -r '.[] | select(.name | startswith("citest-") or startswith("tftest-")) | "  - \(.name) [\(.uuid)]"'

echo ""
read -p "Delete these networks? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    NETWORKS=$(curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
        -H "Content-Type: application/json" \
        -d '{"service":"CMNet","call":"getNetworks"}' | \
        jq -r '[.[] | select(.name | startswith("citest-") or startswith("tftest-")) | .uuid] | @json')

    if [ "$NETWORKS" != "[]" ]; then
        curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
            -H "Content-Type: application/json" \
            -d "{\"service\":\"CMNet\",\"call\":\"removeNetworks\",\"args\":[$NETWORKS,false]}"
        echo "Networks deleted"
    fi
fi

echo ""
echo "Cleanup complete!"
