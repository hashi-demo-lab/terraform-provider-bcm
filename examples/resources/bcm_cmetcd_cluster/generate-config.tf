# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# =============================================================================
# Example: Generate Terraform configuration from existing BCM etcd cluster
# =============================================================================
# Terraform 1.5+ supports generating HCL configuration from existing resources.
# This is useful for adopting existing infrastructure into Terraform management.
#
# Step 1: Add the import block below to your configuration
# Step 2: Run: terraform plan -generate-config-out=generated_etcd_cluster.tf
# Step 3: Review and adjust the generated configuration
# Step 4: Run: terraform plan (verify no unexpected changes)
# Step 5: Remove the import block and move the resource to your main config
#
# Equivalent CLI command:
#   terraform import bcm_cmetcd_cluster.example "e1e2e3e4-f5f6-7890-abcd-ef1234567890"

import {
  to = bcm_cmetcd_cluster.example
  id = "e1e2e3e4-f5f6-7890-abcd-ef1234567890"
}
