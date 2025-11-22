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
  base_image_id = [
    for img in data.bcm_cmpart_softwareimages.base.images :
    img.id if img.name == "default-image"
  ][0]
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
