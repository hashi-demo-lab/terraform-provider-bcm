# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# =============================================================================
# Example: Import an existing software image into Terraform
# =============================================================================
# Use the softwareimages data source to discover the UUID, then write a
# matching resource configuration before running terraform import.
#
# Step 1: terraform apply -target=data.bcm_cmpart_softwareimages.all
# Step 2: terraform import bcm_cmpart_softwareimage.existing <image-uuid>
# Step 3: terraform plan (verify no unexpected changes)

# Lookup existing images to find UUIDs
data "bcm_cmpart_softwareimages" "all" {}

# Find the target image by name
locals {
  target_image = [
    for img in data.bcm_cmpart_softwareimages.all.images :
    img if img.name == "ubuntu-22.04-dpu"
  ][0]
}

# Write a matching resource configuration
resource "bcm_cmpart_softwareimage" "existing" {
  name = local.target_image.name
  path = local.target_image.path
}
