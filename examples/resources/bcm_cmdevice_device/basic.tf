terraform {
  required_providers {
    bcm = {
      source = "hashicorp/bcm"
    }
  }
}

provider "bcm" {
  insecure_skip_verify = true
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
resource "bcm_cmdevice_category" "basic_category" {
  name               = "citest-category-${local.test_suffix}"
  management_network = data.bcm_cmnet_networks.management.networks[0].id

  software_image_proxy = {
    parent_software_image = bcm_cmpart_softwareimage.test_image.id
  }

  notes = "Test category (run: ${local.test_suffix})"

  depends_on = [bcm_cmpart_softwareimage.test_image]
}

# Example: Basic compute device with minimal configuration
# Note: partition is NOT specified because the category has a software_image_proxy
# which provides the partition automatically
resource "bcm_cmdevice_device" "basic" {
  hostname           = "citest-device-${local.test_suffix}"
  mac                = "00:11:22:33:44:AA"
  category           = bcm_cmdevice_category.basic_category.id
  management_network = data.bcm_cmnet_networks.management.networks[0].id

  notes = "Basic test device (run: ${local.test_suffix})"

  depends_on = [bcm_cmdevice_category.basic_category]
}

# Outputs
output "device_id" {
  value       = bcm_cmdevice_device.basic.id
  description = "UUID of the created device"
}

output "device_hostname" {
  value       = bcm_cmdevice_device.basic.hostname
  description = "Hostname of the created device"
}

output "test_suffix" {
  value       = local.test_suffix
  description = "Unique suffix for this test run"
}
