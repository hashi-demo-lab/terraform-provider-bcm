# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# =============================================================================
# Example: Generate Terraform configuration from existing BCM network
# =============================================================================
# Terraform 1.5+ supports generating HCL configuration from existing resources.
# This is useful for adopting existing infrastructure into Terraform management.
#
# Step 1: Add the import block below to your configuration
# Step 2: Run: terraform plan -generate-config-out=generated_network.tf
# Step 3: Review and adjust the generated configuration
# Step 4: Run: terraform plan (verify no unexpected changes)
# Step 5: Remove the import block and move the resource to your main config
#
# Equivalent CLI command:
#   terraform import bcm_cmnet_network.example "7f8a9b0c-1d2e-3f4a-5b6c-7d8e9f0a1b2c"

import {
  to = bcm_cmnet_network.example
  id = "7f8a9b0c-1d2e-3f4a-5b6c-7d8e9f0a1b2c"
}
