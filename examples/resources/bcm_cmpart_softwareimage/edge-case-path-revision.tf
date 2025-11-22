# Edge Case Example: Path with Revision Syntax
# Use case: Version software images using BCM's @revision syntax
#
# BCM API Feature:
# - Paths can include @revision suffix for versioning
# - Format: /path/to/image@revision_number
# - Allows multiple versions of the same logical image
#
# Solution: Use @revision syntax for image versioning strategy

# Configure the BCM provider
# Authentication can be provided via environment variables:
# export BCM_ENDPOINT="https://bcm.example.com:8081"
# export BCM_USERNAME="admin"
# export BCM_PASSWORD="your-password"


# Query base image
data "bcm_cmpart_softwareimages" "base" {}

locals {
  base_image_id = [
    for img in data.bcm_cmpart_softwareimages.base.images :
    img.id if img.name == "default-image"
  ][0]

  # Version management locals
  image_version = "123" # Could be from variable or data source
}

# Example 1: Simple revision syntax
resource "bcm_cmpart_softwareimage" "versioned_simple" {
  name = "ubuntu-22.04-v1"
  path = "/cm/images/ubuntu-22.04@1" # @1 indicates revision 1

  original_image = local.base_image_id
  kernel_version = "6.8.0-51-generic"

  notes = "Ubuntu 22.04 - Revision 1"
}

# Example 2: Dynamic revision from variable
resource "bcm_cmpart_softwareimage" "versioned_dynamic" {
  name = "ubuntu-22.04-v${local.image_version}"
  path = "/cm/images/ubuntu-22.04@${local.image_version}"

  original_image = local.base_image_id
  kernel_version = "6.8.0-51-generic"

  notes = "Ubuntu 22.04 - Revision ${local.image_version} (dynamic)"
}

# Example 3: Multi-level path with revision
resource "bcm_cmpart_softwareimage" "versioned_multilevel" {
  name = "prod-rhel9-gpu-v5"
  path = "/cm/images/production/rhel-9/gpu-compute@5"

  original_image = local.base_image_id
  kernel_version = "6.8.0-51-generic"

  notes = "Production RHEL 9 GPU image - Revision 5"
}

# Example 4: Production versioning strategy with timestamp
resource "bcm_cmpart_softwareimage" "versioned_timestamp" {
  name = "app-image-${formatdate("YYYYMMDD", timestamp())}"
  path = "/cm/images/application/${formatdate("YYYYMMDD", timestamp())}@${formatdate("hhmm", timestamp())}"

  original_image = local.base_image_id
  kernel_version = "6.8.0-51-generic"

  notes = "Application image versioned by date and time: ${timestamp()}"

  lifecycle {
    ignore_changes = [
      notes, # Ignore timestamp changes in notes
    ]
  }
}

# Example 5: Semantic versioning pattern
variable "major_version" {
  description = "Major version number"
  type        = number
  default     = 2
}

variable "minor_version" {
  description = "Minor version number"
  type        = number
  default     = 1
}

variable "patch_version" {
  description = "Patch version number"
  type        = number
  default     = 0
}

locals {
  semver         = "${var.major_version}.${var.minor_version}.${var.patch_version}"
  semver_encoded = replace(local.semver, ".", "-") # BCM may not support dots in all contexts
}

resource "bcm_cmpart_softwareimage" "versioned_semver" {
  name = "myapp-${local.semver}"
  path = "/cm/images/myapp/v${local.semver_encoded}@${var.patch_version}"

  original_image = local.base_image_id
  kernel_version = "6.8.0-51-generic"

  notes = "Application image version ${local.semver} - Semantic versioning pattern"
}

# Example 6: Blue/Green deployment pattern with revisions
resource "bcm_cmpart_softwareimage" "blue_deployment" {
  name = "webapp-blue"
  path = "/cm/images/webapp/blue@${var.blue_revision}"

  original_image = local.base_image_id
  kernel_version = "6.8.0-51-generic"

  notes = "Blue deployment - Revision ${var.blue_revision}"
}

resource "bcm_cmpart_softwareimage" "green_deployment" {
  name = "webapp-green"
  path = "/cm/images/webapp/green@${var.green_revision}"

  original_image = local.base_image_id
  kernel_version = "6.8.0-51-generic"

  notes = "Green deployment - Revision ${var.green_revision}"
}

variable "blue_revision" {
  description = "Revision number for blue deployment"
  type        = number
  default     = 10
}

variable "green_revision" {
  description = "Revision number for green deployment"
  type        = number
  default     = 11
}

# Outputs
output "revision_syntax_examples" {
  value = {
    simple           = bcm_cmpart_softwareimage.versioned_simple.path
    dynamic          = bcm_cmpart_softwareimage.versioned_dynamic.path
    multilevel       = bcm_cmpart_softwareimage.versioned_multilevel.path
    timestamp        = bcm_cmpart_softwareimage.versioned_timestamp.path
    semver           = bcm_cmpart_softwareimage.versioned_semver.path
    blue_deployment  = bcm_cmpart_softwareimage.blue_deployment.path
    green_deployment = bcm_cmpart_softwareimage.green_deployment.path
  }
}

output "revision_syntax_rules" {
  value = <<-EOT
    BCM Path Revision Syntax Rules:

    Format: /path/to/image@revision

    Valid Examples:
    ✅ /cm/images/ubuntu@1
    ✅ /cm/images/prod/rhel@123
    ✅ /cm/images/app/version-2-1@45

    Use Cases:
    - Version control for images
    - Blue/Green deployments
    - A/B testing configurations
    - Rollback capabilities
    - Timestamp-based versions

    Benefits:
    - Multiple versions of same logical image
    - Clear version identification in path
    - Easy rollback by changing revision number
    - Compatible with BCM's internal versioning
  EOT
}
