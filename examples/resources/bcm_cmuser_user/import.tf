# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# =============================================================================
# Example: Import an existing BCM user into Terraform
# =============================================================================
# Import by username. After importing, you must add the password to your
# configuration as it cannot be recovered from the BCM API.
#
# Step 1: terraform import bcm_cmuser_user.existing <username>
# Step 2: Add password to configuration (required for subsequent applies)
# Step 3: terraform plan (verify no unexpected changes)

# Import by username
# terraform import bcm_cmuser_user.existing "cmsupport"

# Write a matching resource configuration
# Note: password is required but cannot be read back from BCM after import
resource "bcm_cmuser_user" "existing" {
  username       = "cmsupport"
  password       = var.imported_user_password
  full_name      = "CM Support"
  home_directory = "/home/cmsupport"
  shell          = "/bin/bash"
}

variable "imported_user_password" {
  description = "Password for the imported user (required, cannot be recovered from BCM)"
  type        = string
  sensitive   = true
}
