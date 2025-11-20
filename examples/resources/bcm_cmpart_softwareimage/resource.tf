# Query existing software images to find base image for cloning
data "bcm_cmpart_softwareimages" "available" {}

# Find the default image UUID using a data source lookup
# This is more maintainable than hardcoding UUIDs
locals {
  default_image = [
    for img in data.bcm_cmpart_softwareimages.available.images :
    img.id if img.name == "default-image"
  ][0]
}

# Basic example: Clone from an existing image
resource "bcm_cmpart_softwareimage" "example" {
  name = "ubuntu-22.04-dpu"
  path = "/cm/images/ubuntu-22.04-dpu"

  # Clone from the default image using data source lookup (best practice)
  # This ensures the UUID is always current and works across environments
  original_image = local.default_image

  kernel_version    = "6.8.0-51-generic"
  kernel_parameters = "quiet splash"

  enable_sol = true
  sol_speed  = "115200"

  notes = "Ubuntu 22.04 LTS image for DPU nodes (cloned from default-image)"
}
