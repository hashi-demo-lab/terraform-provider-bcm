# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# =============================================================================
# Example: Generate Terraform configuration from existing BCM software image
# =============================================================================
# Terraform 1.5+ supports generating HCL configuration from existing resources.
# This is useful for adopting existing infrastructure into Terraform management.
#
# Step 1: Add the import block below to your configuration
# Step 2: Run: terraform plan -generate-config-out=generated_softwareimage.tf
# Step 3: Review and adjust the generated configuration
# Step 4: Run: terraform plan (verify no unexpected changes)
# Step 5: Remove the import block and move the resource to your main config
#
# Equivalent CLI command:
#   terraform import bcm_cmpart_softwareimage.example "d4e5f6a7-b8c9-0d1e-2f3a-4b5c6d7e8f9a"

import {
  to = bcm_cmpart_softwareimage.example
  id = "d4e5f6a7-b8c9-0d1e-2f3a-4b5c6d7e8f9a"
}
