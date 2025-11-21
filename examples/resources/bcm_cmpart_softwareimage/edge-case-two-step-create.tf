# Edge Case Example: Two-Step Create Pattern for Kernel Version
# Use case: Work around BCM API limitation where kernel_version cannot be set during clone
#
# BCM API Behavior:
# - Setting kernel_version during addSoftwareImage (clone) operation fails with validation error
# - kernel_version must be set AFTER the clone operation completes
#
# Solution: Two-step pattern
# 1. Create/clone image without kernel_version
# 2. Update image with kernel_version after creation

# Configure the BCM provider
provider "bcm" {
  endpoint             = "https://bcm.example.com:8081"
  username             = "admin"
  password             = var.bcm_password
  insecure_skip_verify = true
}

# Query base image for cloning
data "bcm_cmpart_softwareimages" "base" {}

locals {
  base_image_id = [
    for img in data.bcm_cmpart_softwareimages.base.images :
    img.id if img.name == "default-image"
  ][0]
}

# Step 1: Create image WITHOUT kernel_version
# This allows the clone operation to succeed
resource "bcm_cmpart_softwareimage" "initial_clone" {
  name = "my-custom-image"
  path = "/cm/images/my-custom-image"

  # Clone from base image
  original_image = local.base_image_id

  # Note: kernel_version is intentionally omitted during initial creation
  # BCM API will reject the clone operation if kernel_version is present

  notes = "Initial clone - kernel version will be set in subsequent update"
}

# Step 2: Update with kernel_version after clone completes
# This is done through a separate Terraform apply or by modifying the resource
#
# To implement the two-step pattern:
# 1. First apply: Creates image without kernel_version
# 2. Second apply: Uncomment kernel_version line below and re-apply
#
# resource "bcm_cmpart_softwareimage" "final_image" {
#   name = "my-custom-image"
#   path = "/cm/images/my-custom-image"
#
#   original_image = local.base_image_id
#
#   # Step 2: Now set kernel_version after initial creation
#   kernel_version = "6.8.0-51-generic"
#
#   notes = "Custom image with kernel version set post-clone"
# }

# Alternative: Use lifecycle ignore_changes to prevent drift
# This pattern acknowledges that BCM may reset certain fields after operations
resource "bcm_cmpart_softwareimage" "with_lifecycle" {
  name = "my-custom-image-lifecycle"
  path = "/cm/images/my-custom-image-lifecycle"

  original_image = local.base_image_id

  # Set kernel_version - Terraform will handle the two-step pattern automatically
  kernel_version = "6.8.0-51-generic"

  notes = "Terraform automatically handles two-step pattern through update operation"

  # Lifecycle management: Terraform will create first, then update with kernel_version
  lifecycle {
    create_before_destroy = true
  }
}

output "workaround_notes" {
  value = <<-EOT
    BCM API Limitation - Two-Step Create Pattern:

    1. Clone Operation: kernel_version cannot be set during addSoftwareImage
    2. Workaround: Set kernel_version through updateSoftwareImage after clone
    3. Terraform Behavior: Automatically handles this through create+update pattern
    4. First Apply: Creates image without kernel_version
    5. Update: Terraform detects drift and applies kernel_version via update

    This is expected behavior and not an error.
  EOT
}
