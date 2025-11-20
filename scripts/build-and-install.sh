#!/bin/bash
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Build and install Terraform BCM provider locally
# This script builds the provider and installs it to ~/.terraform.d/plugins/

set -e  # Exit on error

echo "=========================================="
echo "Building Terraform BCM Provider"
echo "=========================================="

# Set Go environment
export GOMODCACHE=/workspace/.go/pkg/mod
export GOCACHE=/workspace/.go/cache
export GOPATH=/workspace/.go

# Get OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Map architecture names
case "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
esac

echo "Target platform: ${OS}_${ARCH}"

# Provider details
PROVIDER_NAME="bcm"
PROVIDER_VERSION="0.1.0"
PROVIDER_NAMESPACE="hashicorp"
PROVIDER_DIR="$HOME/.terraform.d/plugins/registry.terraform.io/${PROVIDER_NAMESPACE}/${PROVIDER_NAME}/${PROVIDER_VERSION}/${OS}_${ARCH}"
BINARY_NAME="terraform-provider-${PROVIDER_NAME}_v${PROVIDER_VERSION}"

echo "Provider: ${PROVIDER_NAMESPACE}/${PROVIDER_NAME} v${PROVIDER_VERSION}"
echo "Install directory: ${PROVIDER_DIR}"

# Create plugin directory
mkdir -p "${PROVIDER_DIR}"

# Build the provider
echo ""
echo "Building provider binary..."
cd /workspace
go build -o "${PROVIDER_DIR}/${BINARY_NAME}" .

# Check if build was successful
if [ -f "${PROVIDER_DIR}/${BINARY_NAME}" ]; then
    BINARY_SIZE=$(ls -lh "${PROVIDER_DIR}/${BINARY_NAME}" | awk '{print $5}')
    echo ""
    echo "✅ Build successful!"
    echo "Binary: ${PROVIDER_DIR}/${BINARY_NAME}"
    echo "Size: ${BINARY_SIZE}"

    # Make binary executable
    chmod +x "${PROVIDER_DIR}/${BINARY_NAME}"

    # Show installed location
    echo ""
    echo "Provider installed to:"
    echo "  ${PROVIDER_DIR}/"
    ls -lh "${PROVIDER_DIR}/"

    echo ""
    echo "=========================================="
    echo "✅ Provider ready for use!"
    echo "=========================================="
else
    echo ""
    echo "❌ Build failed!"
    exit 1
fi
