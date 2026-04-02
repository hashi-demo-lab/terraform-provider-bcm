# Example: Device with role assignments
#
# This example demonstrates how to create a device with specific roles assigned.
# Roles define the device's function in the cluster (boot, headnode, etc.).
# Use the bcm_cmdevice_roles data source to discover available roles, then assign by name.

# Generate unique suffix for this test run
locals {
  test_suffix = formatdate("YYYYMMDDhhmmss", timestamp())
}

# Discover available roles in the cluster
data "bcm_cmdevice_roles" "all" {}

# Transform roles data into a lookup map by name
# This enables easy reference: local.roles["boot"] returns "boot"
locals {
  # Map of role name -> role name for validation via data source
  roles = { for r in data.bcm_cmdevice_roles.all.roles : r.name => r.name }

  # Define which roles to assign (validated against the data source)
  device_roles = [
    local.roles["boot"],
    local.roles["headnode"],
  ]
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

# Create device with roles assigned BY NAME
resource "bcm_cmdevice_device" "with_roles" {
  hostname = "citest-roles-${local.test_suffix}"
  category = bcm_cmdevice_category.roles_category.id

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = "00:11:22:33:44:BB"
    network  = data.bcm_cmnet_networks.management.networks[0].id
    bootable = true
    dhcp     = true
  }

  # Assign roles by name - provider validates against BCM cluster
  roles = local.device_roles

  notes = "Device with roles assigned by name (run: ${local.test_suffix})"

  depends_on = [bcm_cmdevice_category.roles_category]
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
  description = "Role names assigned to the device"
}

output "available_roles" {
  value       = keys(local.roles)
  description = "All available role names in the cluster"
}
