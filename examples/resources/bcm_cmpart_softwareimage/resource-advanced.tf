# Query all available software images to find the base image for cloning
data "bcm_cmpart_softwareimages" "all" {}

# Lookup the default-image UUID dynamically instead of hardcoding
locals {
  # Find the default image to clone from (more maintainable than hardcoded UUIDs)
  base_image_name = "default-image"
  base_image_uuid = [
    for img in data.bcm_cmpart_softwareimages.all.images :
    img.id if img.name == local.base_image_name
  ][0]
}

# Advanced example: Clone with full configuration including kernel modules
resource "bcm_cmpart_softwareimage" "advanced" {
  name = "rhel-9-gpu-node"
  path = "/cm/images/rhel-9-gpu-node"

  # Clone from an existing image using data source lookup (real-world pattern)
  # This is more maintainable than hardcoding UUIDs which vary per environment
  original_image = local.base_image_uuid

  # Kernel configuration
  kernel_version        = "6.8.0-51-generic"
  kernel_parameters     = "rd.driver.blacklist=nouveau quiet splash"
  kernel_output_console = "tty0"

  # Serial Over LAN (SOL) configuration for remote console access
  enable_sol       = true
  sol_port         = "ttyS1"
  sol_speed        = "115200"
  sol_flow_control = true

  # Kernel modules to load at boot
  modules = [
    {
      name       = "nvidia-drm"
      parameters = "modeset=1"
    },
    {
      name       = "e1000e"
      parameters = ""
    },
    {
      name       = "mlx5_core"
      parameters = "enable_roce=1"
    }
  ]

  notes = "RHEL 9.3 with NVIDIA GPU drivers and Mellanox networking for compute nodes"
}

# Output all available images for reference
output "available_images" {
  description = "Map of image names to UUIDs (useful for selecting base images)"
  value       = { for img in data.bcm_cmpart_softwareimages.all.images : img.name => img.id }
}

# Output the created image details
output "created_image" {
  description = "Details of the newly created software image"
  value = {
    id           = bcm_cmpart_softwareimage.advanced.id
    name         = bcm_cmpart_softwareimage.advanced.name
    path         = bcm_cmpart_softwareimage.advanced.path
    cloned_from  = local.base_image_name
    kernel       = bcm_cmpart_softwareimage.advanced.kernel_version
    module_count = length(bcm_cmpart_softwareimage.advanced.modules)
  }
}
