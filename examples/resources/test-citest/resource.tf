# Test resource example for CI testing
# This example is used by test-examples.sh to validate resource CRUD operations
# Uses bcm_cmdevice_category with unique citest naming pattern

terraform {
  required_version = ">= 1.5.0"

  required_providers {
    bcm = {
      source  = "hashicorp/bcm"
      version = "~> 0.1"
    }
  }
}

# Provider configuration using environment variables
# Set BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD before running
provider "bcm" {
  insecure_skip_verify = true
}

# Query existing software images to find base image for cloning
data "bcm_cmpart_softwareimages" "available" {}

# Dynamically select first available image as base for cloning
locals {
  base_image = length(data.bcm_cmpart_softwareimages.available.images) > 0 ? data.bcm_cmpart_softwareimages.available.images[0].id : null

  # Generate unique name with timestamp pattern: citest-YYYYMMDD-HHMMSS-softwareimage
  timestamp  = formatdate("YYYYMMDD-HHmmss", timestamp())
  image_name = "citest-${local.timestamp}-softwareimage"
  image_path = "/cm/images/${local.image_name}"
}

# Test resource: Create a minimal software image by cloning
# This is the only resource type currently available in the provider
resource "bcm_cmpart_softwareimage" "citest_minimal" {
  name = local.image_name
  path = "/cm/images/citest-minimal"  # Use simple static path for testing

  # Clone from first available image
  original_image = local.base_image

  # Minimal configuration for testing
  # Note: kernel_version causes validation errors during clone - BCM API limitation
  # kernel_version    = "5.15.0-generic"
  kernel_parameters = "quiet"

  notes = "Test software image for CI validation - created by test-examples.sh"
}

# Output resource ID for validation
output "citest_image_id" {
  description = "UUID of the created test software image"
  value       = bcm_cmpart_softwareimage.citest_minimal.id
}

# Output resource name for validation
output "citest_image_name" {
  description = "Name of the created test software image"
  value       = bcm_cmpart_softwareimage.citest_minimal.name
}

# Output image path for validation
output "citest_image_path" {
  description = "Path of the created test software image"
  value       = bcm_cmpart_softwareimage.citest_minimal.path
}
