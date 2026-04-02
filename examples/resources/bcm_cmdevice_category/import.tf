# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# =============================================================================
# Example: Import an existing BCM category into Terraform
# =============================================================================
# Use the categories data source to discover the UUID, then write a matching
# resource configuration before running terraform import.
#
# Step 1: terraform apply -target=data.bcm_cmdevice_categories.existing
# Step 2: terraform import bcm_cmdevice_category.existing <category-uuid>
# Step 3: terraform plan (verify no unexpected changes)

# Lookup the existing category to discover its current configuration
data "bcm_cmdevice_categories" "existing" {
  name = "gpu-compute-nodes"
}

# Lookup the management network for the category
data "bcm_cmnet_networks" "management" {
  filter {
    name_pattern = "managementnet"
  }
}

# Write a resource configuration that matches the existing category
resource "bcm_cmdevice_category" "existing" {
  name               = data.bcm_cmdevice_categories.existing.categories[0].name
  management_network = data.bcm_cmnet_networks.management.networks[0].id
}
