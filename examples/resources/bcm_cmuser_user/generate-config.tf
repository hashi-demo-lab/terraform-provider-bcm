# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# =============================================================================
# Example: Generate Terraform configuration from existing BCM user
# =============================================================================
# Terraform 1.5+ supports generating HCL configuration from existing resources.
# This is useful for adopting existing infrastructure into Terraform management.
#
# Step 1: Add the import block below to your configuration
# Step 2: Run: terraform plan -generate-config-out=generated_user.tf
# Step 3: Review and adjust the generated configuration
# Step 4: Run: terraform plan (verify no unexpected changes)
# Step 5: Remove the import block and move the resource to your main config
#
# Note: The generated configuration will not include the password field.
# You must manually add the password before running terraform apply.
#
# Equivalent CLI command:
#   terraform import bcm_cmuser_user.example "c2d3e4f5-a6b7-8901-2345-6789abcdef01"

import {
  to = bcm_cmuser_user.example
  id = "c2d3e4f5-a6b7-8901-2345-6789abcdef01"
}
