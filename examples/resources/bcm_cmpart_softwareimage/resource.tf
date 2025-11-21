# Production example: Create a custom software image by cloning a base image
# Use case: Deploy standardized OS images with custom kernel configuration

# Configure the BCM provider
# Authentication can be provided via environment variables:
# export BCM_ENDPOINT="https://bcm.example.com:8081"
# export BCM_USERNAME="admin"
# export BCM_PASSWORD="your-password"
provider "bcm" {
  insecure_skip_verify = true # Only for self-signed certificates
}

# Query existing images to find the base image for cloning
data "bcm_cmpart_softwareimages" "available" {}

# Production pattern: Use data source lookup instead of hardcoded UUIDs
# This makes configurations portable across environments
locals {
  base_image_name = "default-image"
  base_image_id = [
    for img in data.bcm_cmpart_softwareimages.available.images :
    img.id if img.name == local.base_image_name
  ][0]
}

# Basic example: Clone and customize an image
resource "bcm_cmpart_softwareimage" "dpu_image" {
  name = "ubuntu-22.04-dpu"
  path = "/cm/images/ubuntu-22.04-dpu"

  # Clone from base image using dynamic lookup (best practice)
  original_image = local.base_image_id

  # Kernel configuration for DPU nodes
  kernel_version    = "6.8.0-51-generic"
  kernel_parameters = "quiet splash console=ttyS0,115200"

  # Enable Serial Over LAN for remote console access
  enable_sol = true
  sol_speed  = "115200"
  sol_port   = "ttyS0"

  notes = "Ubuntu 22.04 LTS customized for DPU nodes - Cloned from ${local.base_image_name}"

  # Production pattern: Add lifecycle rules to prevent accidental destruction
  lifecycle {
    prevent_destroy       = false # Set to true in production
    create_before_destroy = true
  }
}

# Output the created image details for reference
output "dpu_image_id" {
  description = "UUID of the created DPU image"
  value       = bcm_cmpart_softwareimage.dpu_image.id
}

output "dpu_image_path" {
  description = "Path to the DPU image on the BCM server"
  value       = bcm_cmpart_softwareimage.dpu_image.path
}
