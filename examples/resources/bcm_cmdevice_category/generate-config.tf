# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# =============================================================================
# Example: Generate Terraform configuration from existing BCM category
# =============================================================================
# Terraform 1.5+ supports generating HCL configuration from existing resources.
# This is useful for adopting existing infrastructure into Terraform management.
#
# Step 1: Add the import block below to your configuration
# Step 2: Run: terraform plan -generate-config-out=generated_category.tf
# Step 3: Review and adjust the generated configuration
# Step 4: Run: terraform plan (verify no unexpected changes)
# Step 5: Remove the import block and move the resource to your main config
#
# Equivalent CLI command:
#   terraform import bcm_cmdevice_category.example "0ae6d733-3015-4479-bfab-ce2d237a2809"

import {
  to = bcm_cmdevice_category.example
  id = "0ae6d733-3015-4479-bfab-ce2d237a2809"
}
