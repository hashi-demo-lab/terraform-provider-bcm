# K8s Control Plane Workflow Integration Example
# This example shows how to use categories with software images

# Step 1: Get all software images to find the base image
data "bcm_cmpart_softwareimages" "all" {}

# Step 2: Create a custom software image with kernel modules
resource "bcm_cmpart_softwareimage" "k8s_control_plane" {
  name = "k8s-control-plane-image"

  # Reference the base image by finding its UUID
  original_image = data.bcm_cmpart_softwareimages.all.images[
    index(data.bcm_cmpart_softwareimages.all.images.*.name, "default-image")
  ].id

  path = "/cm/images/k8s-control-plane-image"

  # Add required kernel modules
  modules = [
    {
      name       = "mlx5_core"
      parameters = ""
    },
    {
      name       = "bonding"
      parameters = ""
    }
  ]

  # Optional kernel parameters
  kernel_parameters = "quiet splash console=ttyS0,115200"
}

# Step 3: Query the default category to use as a template
data "bcm_cmdevice_categories" "default" {
  name = "default"
}

# Output the category UUID that will be used for node provisioning
output "default_category_id" {
  description = "Category UUID for node provisioning"
  value       = data.bcm_cmdevice_categories.default.categories[0].uuid
}

output "k8s_image_id" {
  description = "K8s control plane software image UUID"
  value       = bcm_cmpart_softwareimage.k8s_control_plane.id
}

# Future: Create custom category (when resource is implemented)
# resource "bcm_cmdevice_category" "k8s_control_plane" {
#   name              = "k8s-control-plane"
#   original_category = data.bcm_cmdevice_categories.default.categories[0].uuid
#   software_image_id = bcm_cmpart_softwareimage.k8s_control_plane.id
#   disksetup         = "/cm/local/apps/cmd/etc/htdocs/disk-setup/x86_64-slave-one-big-partition-ext4.xml"
# }

# Display category configuration for verification
output "category_configuration" {
  description = "Current category configuration details"
  value = {
    name              = data.bcm_cmdevice_categories.default.categories[0].name
    uuid              = data.bcm_cmdevice_categories.default.categories[0].uuid
    software_image_id = data.bcm_cmdevice_categories.default.categories[0].software_image_id
    boot_loader       = data.bcm_cmdevice_categories.default.categories[0].boot_loader
    install_mode      = data.bcm_cmdevice_categories.default.categories[0].install_mode
    disksetup_length  = length(data.bcm_cmdevice_categories.default.categories[0].disksetup)
  }
}
