#!/usr/bin/env bash
# Cleanup script for basic.tf test resources

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
    jq -r '.[] | select(.name | startswith("citest-")) | "  - \(.name) [\(.uuid)]"'

echo ""
read -p "Delete these software images? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    IMAGES=$(curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
        -H "Content-Type: application/json" \
        -d '{"service":"CMPart","call":"getSoftwareImages"}' | \
        jq -r '[.[] | select(.name | startswith("citest-")) | .name] | @json')

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
    jq -r '.[] | select(.name | startswith("citest-")) | "  - \(.name) [\(.uuid)]"'

echo ""
read -p "Delete these categories? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    CATEGORIES=$(curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
        -H "Content-Type: application/json" \
        -d '{"service":"CMDevice","call":"getCategories"}' | \
        jq -r '[.[] | select(.name | startswith("citest-")) | .uuid] | @json')

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
    jq -r '.[] | select(.hostname | startswith("citest-")) | "  - \(.hostname) [\(.uuid)]"'

echo ""
read -p "Delete these devices? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    DEVICES=$(curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
        -H "Content-Type: application/json" \
        -d '{"service":"CMDevice","call":"getNodes"}' | \
        jq -r '[.[] | select(.hostname | startswith("citest-")) | .uuid] | @json')

    if [ "$DEVICES" != "[]" ]; then
        curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
            -H "Content-Type: application/json" \
            -d "{\"service\":\"CMDevice\",\"call\":\"removeNodes\",\"args\":[$DEVICES,false]}"
        echo "Devices deleted"
    fi
fi

echo ""
echo "Cleanup complete!"
