# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# =============================================================================
# Example: Import existing device and move to a new category
# =============================================================================
# A common workflow when adopting Terraform for an existing BCM cluster:
# import a device that was created outside Terraform, then change its
# category to bring it under Terraform-managed configuration.
#
# Step 1: terraform apply -target=data.bcm_cmdevice_nodes.server
# Step 2: terraform import bcm_cmdevice_device.server <device-uuid>
# Step 3: terraform plan (shows category change)
# Step 4: terraform apply (moves device to new category)

# Lookup the existing device to discover its current MAC
data "bcm_cmdevice_nodes" "server" {
  filter {
    hostname_pattern = "node001"
  }
}

# Lookup the network for interfaces
data "bcm_cmnet_networks" "management" {
  filter {
    name_pattern = "managementnet"
  }
}

# Lookup the software image for the new category
data "bcm_cmpart_softwareimages" "all" {}

# Create the target category that the device will be moved into
resource "bcm_cmdevice_category" "gpu_compute" {
  name               = "gpu-compute"
  management_network = data.bcm_cmnet_networks.management.networks[0].id

  software_image_proxy = {
    parent_software_image = data.bcm_cmpart_softwareimages.all.images[0].uuid
  }
}

# The device — after import, the category change triggers an update in BCM
resource "bcm_cmdevice_device" "server" {
  hostname = data.bcm_cmdevice_nodes.server.nodes[0].hostname
  category = bcm_cmdevice_category.gpu_compute.id

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = data.bcm_cmdevice_nodes.server.nodes[0].mac
    network  = data.bcm_cmnet_networks.management.networks[0].id
    bootable = true
    dhcp     = true
  }
}
