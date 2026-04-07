---
page_title: "bcm_cmpart_softwareimage Resource - bcm"
subcategory: ""
description: |-
  Manages a BCM software image (OS image for DPU node provisioning).
  Software images define the operating system kernel, kernel parameters, modules, and boot configuration used to provision compute nodes in the BCM cluster.
---

# bcm_cmpart_softwareimage (Resource)

Manages a BCM software image (OS image for DPU node provisioning).

Software images define the operating system kernel, kernel parameters, modules, and boot configuration used to provision compute nodes in the BCM cluster.

## Example Usage

```terraform
# Production example: Create a custom software image by cloning a base image
# Use case: Deploy standardized OS images with custom kernel configuration

# Configure the BCM provider
# Authentication can be provided via environment variables:
# export BCM_ENDPOINT="https://bcm.example.com:8081"
# export BCM_USERNAME="admin"
# export BCM_PASSWORD="your-password"

# Query existing images to find the base image for cloning
data "bcm_cmpart_softwareimages" "available" {}

# Production pattern: Use data source lookup instead of hardcoded UUIDs
# This makes configurations portable across environments
locals {
  base_image_id   = data.bcm_cmpart_softwareimages.available.images[0].id
  base_image_name = data.bcm_cmpart_softwareimages.available.images[0].name
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
```

### Advanced Production Image

```terraform
# Advanced production example: GPU compute node image with full configuration
# Use case: Create enterprise-grade images for HPC/AI workloads with GPU support

# Configure the BCM provider
# Authentication can be provided via environment variables:
# export BCM_ENDPOINT="https://bcm.example.com:8081"
# export BCM_USERNAME="admin"
# export BCM_PASSWORD="your-password"


# Query all available software images
data "bcm_cmpart_softwareimages" "all" {}

# Production pattern: Dynamic base image lookup for portability
locals {
  base_image_uuid = data.bcm_cmpart_softwareimages.all.images[0].id

  # Environment-specific configuration
  environment  = "production"
  cluster_name = "hpc-cluster-01"
}

# Advanced example: Full-featured GPU compute image
resource "bcm_cmpart_softwareimage" "gpu_compute" {
  name = "${local.environment}-rhel9-gpu-${local.cluster_name}"
  path = "/cm/images/${local.environment}/rhel-9-gpu-compute"

  # Clone from verified base image (dynamic lookup pattern)
  original_image = local.base_image_uuid

  # Production kernel configuration for GPU nodes
  kernel_version        = "6.8.0-51-generic"
  kernel_parameters     = "rd.driver.blacklist=nouveau quiet splash iommu=pt intel_iommu=on"
  kernel_output_console = "tty0"

  # Serial Over LAN configuration for lights-out management
  enable_sol       = true
  sol_port         = "ttyS1"
  sol_speed        = "115200"
  sol_flow_control = true

  # Production kernel modules: GPU drivers + high-speed networking
  modules = [
    {
      name       = "nvidia-drm"
      parameters = "modeset=1"
    },
    {
      name       = "nvidia-uvm"
      parameters = ""
    },
    {
      name       = "e1000e"
      parameters = ""
    },
    {
      name       = "mlx5_core"
      parameters = "enable_roce=1"
    },
    {
      name       = "ib_core"
      parameters = ""
    }
  ]

  notes = <<-EOT
    Production GPU compute image for ${local.cluster_name}
    - RHEL 9.3 base
    - NVIDIA GPU drivers with DRM/UVM support
    - Mellanox ConnectX networking with RoCE
    - InfiniBand support
    - Cloned from: base image
    - Environment: ${local.environment}
  EOT

  # Production lifecycle management
  lifecycle {
    prevent_destroy       = false # Set to true in production
    create_before_destroy = true

    # Ignore changes to original_image after initial creation
    # This prevents recreation if the base image is updated
    ignore_changes = [original_image]
  }
}

# Create a second image variant for CPU-only nodes
resource "bcm_cmpart_softwareimage" "cpu_compute" {
  name = "${local.environment}-rhel9-cpu-${local.cluster_name}"
  path = "/cm/images/${local.environment}/rhel-9-cpu-compute"

  original_image = local.base_image_uuid

  kernel_version    = "6.8.0-51-generic"
  kernel_parameters = "quiet splash"

  enable_sol = true
  sol_speed  = "115200"

  # Minimal modules for CPU-only nodes
  modules = [
    {
      name       = "e1000e"
      parameters = ""
    }
  ]

  notes = "CPU-only compute image for ${local.cluster_name} (${local.environment})"

  lifecycle {
    prevent_destroy       = false
    create_before_destroy = true
  }
}

# Outputs for cluster management

output "available_images_map" {
  description = "Reference map of all available base images"
  value       = { for img in data.bcm_cmpart_softwareimages.all.images : img.name => img.id }
}

output "gpu_image_details" {
  description = "GPU compute image configuration details"
  value = {
    id             = bcm_cmpart_softwareimage.gpu_compute.id
    name           = bcm_cmpart_softwareimage.gpu_compute.name
    path           = bcm_cmpart_softwareimage.gpu_compute.path
    cloned_from    = data.bcm_cmpart_softwareimages.all.images[0].name
    kernel_version = bcm_cmpart_softwareimage.gpu_compute.kernel_version
    module_count   = length(bcm_cmpart_softwareimage.gpu_compute.modules)
    modules        = [for mod in bcm_cmpart_softwareimage.gpu_compute.modules : mod.name]
  }
}

output "cpu_image_details" {
  description = "CPU compute image configuration details"
  value = {
    id             = bcm_cmpart_softwareimage.cpu_compute.id
    name           = bcm_cmpart_softwareimage.cpu_compute.name
    path           = bcm_cmpart_softwareimage.cpu_compute.path
    kernel_version = bcm_cmpart_softwareimage.cpu_compute.kernel_version
  }
}

# Generate image assignment map for node provisioning
output "image_assignment_map" {
  description = "Map of node types to image IDs for provisioning workflows"
  value = {
    "gpu_nodes" = bcm_cmpart_softwareimage.gpu_compute.id
    "cpu_nodes" = bcm_cmpart_softwareimage.cpu_compute.id
  }
}
```

### Edge Case: Empty Module Parameters

```terraform
# Edge Case Example: Kernel Modules with Empty Parameters
# Use case: Load kernel modules that don't require parameters
#
# BCM API Behavior:
# - Module "parameters" field must be empty string "", NOT null or omitted
# - Setting parameters = null causes API errors
# - Omitting parameters field causes API errors
#
# Solution: Always set parameters to empty string when no parameters are needed

# Configure the BCM provider
# Authentication can be provided via environment variables:
# export BCM_ENDPOINT="https://bcm.example.com:8081"
# export BCM_USERNAME="admin"
# export BCM_PASSWORD="your-password"


# Query base image
data "bcm_cmpart_softwareimages" "base" {}

locals {
  base_image_id = data.bcm_cmpart_softwareimages.base.images[0].id
}

# Correct: Modules with empty parameters field
resource "bcm_cmpart_softwareimage" "correct_modules" {
  name = "image-with-modules-correct"
  path = "/cm/images/modules-correct"

  original_image = local.base_image_id
  kernel_version = "6.8.0-51-generic"

  # Correct pattern: Always provide parameters field
  modules = [
    {
      name       = "nvidia-drm"
      parameters = "modeset=1" # Module WITH parameters
    },
    {
      name       = "nvidia-uvm"
      parameters = "" # Module WITHOUT parameters - use empty string
    },
    {
      name       = "e1000e"
      parameters = "" # Another module without parameters
    },
  ]

  notes = "Correct: All modules have parameters field (empty string when no params needed)"
}

# Common Mistake #1: Omitting parameters field (WILL FAIL)
# resource "bcm_cmpart_softwareimage" "wrong_omit_params" {
#   name = "image-wrong-omit"
#   path = "/cm/images/wrong-omit"
#
#   original_image = local.base_image_id
#   kernel_version = "6.8.0-51-generic"
#
#   modules = [
#     {
#       name = "nvidia-uvm"
#       # ERROR: parameters field is omitted - BCM API will reject
#     },
#   ]
# }

# Common Mistake #2: Setting parameters = null (WILL FAIL)
# resource "bcm_cmpart_softwareimage" "wrong_null_params" {
#   name = "image-wrong-null"
#   path = "/cm/images/wrong-null"
#
#   original_image = local.base_image_id
#   kernel_version = "6.8.0-51-generic"
#
#   modules = [
#     {
#       name       = "nvidia-uvm"
#       parameters = null # ERROR: null is not allowed - must be empty string
#     },
#   ]
# }

# Production Example: Mixed modules with and without parameters
resource "bcm_cmpart_softwareimage" "production_modules" {
  name = "production-gpu-image"
  path = "/cm/images/production-gpu"

  original_image = local.base_image_id
  kernel_version = "6.8.0-51-generic"

  modules = [
    # GPU driver with specific modeset parameter
    {
      name       = "nvidia-drm"
      parameters = "modeset=1"
    },

    # GPU UVM module - no parameters needed (empty string)
    {
      name       = "nvidia-uvm"
      parameters = ""
    },

    # Network driver with specific configuration
    {
      name       = "mlx5_core"
      parameters = "debug_mask=0x8"
    },

    # Standard network driver - no parameters (empty string)
    {
      name       = "e1000e"
      parameters = ""
    },

    # InfiniBand module - no parameters (empty string)
    {
      name       = "ib_uverbs"
      parameters = ""
    },
  ]

  notes = "Production GPU image with multiple kernel modules - correct empty parameter handling"
}

output "module_parameter_rules" {
  value = <<-EOT
    BCM API Requirements for Kernel Module Parameters:

    CORRECT:
    ✅ parameters = ""           (empty string for no parameters)
    ✅ parameters = "key=value"  (actual parameters)

    INCORRECT:
    ❌ parameters field omitted  (BCM API error)
    ❌ parameters = null         (BCM API error)

    Always provide the parameters field, even if empty!
  EOT
}
```

### Edge Case: Path Revision Syntax

```terraform
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
  base_image_id = data.bcm_cmpart_softwareimages.base.images[0].id

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
```

### Edge Case: Two-Step Create Pattern

```terraform
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
# Authentication can be provided via environment variables:
# export BCM_ENDPOINT="https://bcm.example.com:8081"
# export BCM_USERNAME="admin"
# export BCM_PASSWORD="your-password"


# Query base image for cloning
data "bcm_cmpart_softwareimages" "base" {}

locals {
  base_image_id = data.bcm_cmpart_softwareimages.base.images[0].id
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Software image name (must be unique)
- `path` (String) Image path in BCM filesystem (e.g., `/cm/images/ubuntu-22.04`). Must be unique.

### Optional

- `enable_sol` (Boolean) Enable Serial Over LAN for remote console access. Defaults to `false`.
- `force` (Boolean) Force deletion even if categories reference this software image. **WARNING**: Force deletion may create orphaned references in the BCM database. Use with caution.
- `kernel_output_console` (String) Kernel output console device. Defaults to `tty0`.
- `kernel_parameters` (String) Kernel command-line parameters (e.g., `quiet splash`)
- `kernel_version` (String) Kernel version string (e.g., `5.15.0-58-generic`). When cloning an image, this value is inherited from the source image and becomes known after the clone completes.
- `modules` (Attributes List) List of kernel modules to load at boot (see [below for nested schema](#nestedatt--modules))
- `notes` (String) User notes or description for the software image
- `original_image` (String) UUID of the original image to clone from. When set, BCM will copy the filesystem from the specified image. This is only used during resource creation.
- `sol_flow_control` (Boolean) Enable SOL hardware flow control. Defaults to `true`.
- `sol_port` (String) SOL serial port device. Defaults to `ttyS1`.
- `sol_speed` (String) SOL baud rate. Valid values: `115200`, `57600`, `38400`, `19200`, `9600`, `4800`, `2400`, `1200`. Defaults to `115200`.

### Read-Only

- `bootfspart` (String) Boot filesystem partition UUID reference (auto-generated when cloning)
- `creation_time` (Number) Unix timestamp of image creation (seconds since epoch)
- `file_operation_in_progress` (Boolean) Indicates if a file operation is currently in progress for this image
- `fspart` (String) Filesystem partition UUID reference (auto-generated when cloning)
- `id` (String) Resource identifier (same as UUID)
- `parent_software_image` (String) UUID of the parent image if this is a revision
- `revision_id` (Number) Image revision number
- `uuid` (String) Unique identifier assigned by BCM

<a id="nestedatt--modules"></a>
### Nested Schema for `modules`

Required:

- `name` (String) Kernel module name (e.g., `nvidia-drm`, `e1000e`)

Optional:

- `parameters` (String) Module parameters (e.g., `modeset=1`)

## Import

Import is supported using the following syntax:

```shell
#!/bin/bash
# Import an existing software image using its UUID

# Example: Import a software image with UUID
terraform import bcm_cmpart_softwareimage.example "eaad50d3-432a-4703-a9f8-66551c255a69"

# You can find the UUID of existing images using the data source:
# data "bcm_cmpart_softwareimages" "all" {}
# output "image_uuids" {
#   value = { for img in data.bcm_cmpart_softwareimages.all.images : img.name => img.uuid }
# }
```

### Import with Data Source Lookup

```terraform
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
```

### Generate Configuration (Terraform 1.5+)

```terraform
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# =============================================================================
# Example: Generate Terraform configuration from existing BCM software image
# =============================================================================
# Terraform 1.5+ supports generating HCL configuration from existing resources.
# This is useful for adopting existing infrastructure into Terraform management.
#
# Step 1: Add the import block below to your configuration
# Step 2: Run: terraform plan -generate-config-out=generated_softwareimage.tf
# Step 3: Review and adjust the generated configuration
# Step 4: Run: terraform plan (verify no unexpected changes)
# Step 5: Remove the import block and move the resource to your main config
#
# Equivalent CLI command:
#   terraform import bcm_cmpart_softwareimage.example "d4e5f6a7-b8c9-0d1e-2f3a-4b5c6d7e8f9a"

import {
  to = bcm_cmpart_softwareimage.example
  id = "d4e5f6a7-b8c9-0d1e-2f3a-4b5c6d7e8f9a"
}
```
