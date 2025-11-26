# Example: Device with role assignments
#
# This example demonstrates how to create a device with specific roles assigned.
# Roles define the device's function in the cluster (monitoring, storage, etc.).
# Role UUIDs are obtained from the bcm_cmdevice_roles data source.

# Lookup available roles in the cluster
data "bcm_cmdevice_roles" "all" {}

# Extract specific role UUIDs by name
# Available roles vary by BCM cluster - common ones include: backup, provisioning, boot
locals {
  backup_role       = [for r in data.bcm_cmdevice_roles.all.roles : r.uuid if r.name == "backup"][0]
  provisioning_role = [for r in data.bcm_cmdevice_roles.all.roles : r.uuid if r.name == "provisioning"][0]
}

# Generate unique suffix for this test run
locals {
  test_suffix = formatdate("YYYYMMDDhhmmss", timestamp())
}

# Lookup management network
data "bcm_cmnet_networks" "management" {
  filter {
    name_pattern = "managementnet"
  }
}

# Create a unique software image for each test run
resource "bcm_cmpart_softwareimage" "test_image" {
  name = "citest-image-${local.test_suffix}"
  path = "/cm/images/ubuntu-22.04-server-amd64.iso"
}

# Create a unique category for each test run
resource "bcm_cmdevice_category" "roles_category" {
  name               = "citest-category-${local.test_suffix}"
  management_network = data.bcm_cmnet_networks.management.networks[0].id

  software_image_proxy = {
    parent_software_image = bcm_cmpart_softwareimage.test_image.id
  }

  notes = "Test category for device with roles (run: ${local.test_suffix})"

  depends_on = [bcm_cmpart_softwareimage.test_image]
}

# Create device with multiple roles assigned
resource "bcm_cmdevice_device" "with_roles" {
  hostname           = "citest-roles-${local.test_suffix}"
  mac                = "00:11:22:33:44:BB"
  category           = bcm_cmdevice_category.roles_category.id
  management_network = data.bcm_cmnet_networks.management.networks[0].id

  # Assign multiple roles using UUIDs from the data source
  roles = [
    local.backup_role,
    local.provisioning_role,
  ]

  notes = "Device with backup and provisioning roles (run: ${local.test_suffix})"

  depends_on = [bcm_cmdevice_category.roles_category, data.bcm_cmdevice_roles.all]
}

# Outputs
output "device_id" {
  value       = bcm_cmdevice_device.with_roles.id
  description = "UUID of the created device"
}

output "device_hostname" {
  value       = bcm_cmdevice_device.with_roles.hostname
  description = "Hostname of the created device"
}

output "device_roles" {
  value       = bcm_cmdevice_device.with_roles.roles
  description = "Role UUIDs assigned to the device"
}

output "available_roles" {
  value       = [for r in data.bcm_cmdevice_roles.all.roles : { name = r.name, uuid = r.uuid }]
  description = "All available roles in the cluster"
}
