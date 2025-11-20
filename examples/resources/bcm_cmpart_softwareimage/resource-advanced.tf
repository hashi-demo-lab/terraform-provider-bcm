# Advanced production example: GPU compute node image with full configuration
# Use case: Create enterprise-grade images for HPC/AI workloads with GPU support

# Query all available software images
data "bcm_cmpart_softwareimages" "all" {}

# Production pattern: Dynamic base image lookup for portability
locals {
  base_image_name = "default-image"
  base_image_uuid = [
    for img in data.bcm_cmpart_softwareimages.all.images :
    img.id if img.name == local.base_image_name
  ][0]

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
    - Cloned from: ${local.base_image_name}
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
    cloned_from    = local.base_image_name
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
