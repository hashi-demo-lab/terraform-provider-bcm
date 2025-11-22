#!/bin/bash
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Generate Terraform provider documentation using Docker container
# This ensures consistent documentation generation across different environments

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Generating Terraform Provider Documentation in Docker ===${NC}"

# Configuration
GOLANG_IMAGE="${GOLANG_IMAGE:-golang:1.24}"
TERRAFORM_VERSION="${TERRAFORM_VERSION:-1.13.5}"
PROVIDER_NAME="${PROVIDER_NAME:-bcm}"

echo -e "${YELLOW}Configuration:${NC}"
echo "  Go Image: $GOLANG_IMAGE"
echo "  Terraform Version: $TERRAFORM_VERSION"
echo "  Provider Name: $PROVIDER_NAME"
echo "  Project Root: $PROJECT_ROOT"
echo ""

# Step 1: Format examples with Terraform in container
echo -e "${BLUE}Step 1: Formatting Terraform examples...${NC}"
docker run --rm \
  -v "$PROJECT_ROOT:/workspace" \
  -w /workspace \
  --user "$(id -u):$(id -g)" \
  hashicorp/terraform:${TERRAFORM_VERSION} \
  fmt -recursive examples/ || {
    echo -e "${YELLOW}Warning: terraform fmt failed, continuing...${NC}"
  }

# Step 2: Generate copyright headers
echo -e "${BLUE}Step 2: Generating copyright headers...${NC}"
docker run --rm \
  -v "$PROJECT_ROOT:/workspace" \
  -w /workspace/tools \
  --user "$(id -u):$(id -g)" \
  -e GOCACHE=/workspace/.go/cache \
  -e GOMODCACHE=/workspace/.go/pkg/mod \
  -e GOPATH=/workspace/.go \
  "$GOLANG_IMAGE" \
  go run github.com/hashicorp/copywrite headers -d .. --config ../.copywrite.hcl

# Step 3: Install dependencies and generate provider documentation
echo -e "${BLUE}Step 3: Installing dependencies...${NC}"
docker run --rm \
  -v "$PROJECT_ROOT:/workspace" \
  -w /workspace/tools \
  --user "$(id -u):$(id -g)" \
  -e GOCACHE=/workspace/.go/cache \
  -e GOMODCACHE=/workspace/.go/pkg/mod \
  -e GOPATH=/workspace/.go \
  "$GOLANG_IMAGE" \
  sh -c "go mod download && go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest"

echo -e "${BLUE}Step 4: Generating provider documentation...${NC}"
docker run --rm \
  -v "$PROJECT_ROOT:/workspace" \
  -w /workspace \
  --user "$(id -u):$(id -g)" \
  -e GOCACHE=/workspace/.go/cache \
  -e GOMODCACHE=/workspace/.go/pkg/mod \
  -e GOPATH=/workspace/.go \
  -e PATH="/workspace/.go/bin:$PATH" \
  "$GOLANG_IMAGE" \
  /workspace/.go/bin/tfplugindocs generate \
    --provider-name "$PROVIDER_NAME" \
    --tf-version "$TERRAFORM_VERSION"

echo ""
echo -e "${GREEN}✅ Documentation generation complete!${NC}"
echo -e "${BLUE}Generated files:${NC}"
echo "  - docs/data-sources/"
echo "  - docs/resources/"
echo "  - docs/index.md"
