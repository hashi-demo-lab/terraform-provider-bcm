# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# =============================================================================
# Example: Import an existing BCM network into Terraform
# =============================================================================
# Use the networks data source to discover the UUID, then write a matching
# resource configuration before running terraform import.
#
# Step 1: terraform apply -target=data.bcm_cmnet_networks.existing
# Step 2: terraform import bcm_cmnet_network.existing <network-uuid>
# Step 3: terraform plan (verify no unexpected changes)

# Lookup the existing network
data "bcm_cmnet_networks" "existing" {
  filter {
    name_pattern = "compute-network"
  }
}

# Write a matching resource configuration
resource "bcm_cmnet_network" "existing" {
  name        = data.bcm_cmnet_networks.existing.networks[0].name
  subnet      = "10.0.1.0/24"
  gateway     = "10.0.1.1"
  domain_name = "cluster.local"
}
