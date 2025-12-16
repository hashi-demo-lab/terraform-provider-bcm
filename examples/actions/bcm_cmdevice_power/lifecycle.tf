# Example: Lifecycle-Triggered Power Actions
#
# This example demonstrates how to use lifecycle triggers to automatically
# power on a device after it is created in BCM.
#
# NOTE: Lifecycle triggers require Terraform 1.14 or later
# and may have additional framework requirements.

terraform {
  required_providers {
    bcm = {
      source = "hashi-demo-lab/bcm"
    }
  }
  required_version = ">= 1.14.0"
}

provider "bcm" {
  insecure_skip_verify = true
}

# Data source to get existing category and network
data "bcm_cmdevice_categories" "compute" {
  name_pattern = "default"
}

data "bcm_cmnet_networks" "mgmt" {
  name_pattern = "internalnet"
}

# Create a new device and automatically power it on
resource "bcm_cmdevice_device" "new_worker" {
  hostname           = "citest-lifecycle-node"
  mac                = "00:11:22:33:44:55"
  category           = data.bcm_cmdevice_categories.compute.categories[0].uuid
  management_network = data.bcm_cmnet_networks.mgmt.networks[0].uuid

  # Lifecycle trigger to power on after device creation
  # NOTE: This syntax is subject to change in Terraform 1.14+
  # lifecycle {
  #   action_trigger {
  #     events  = [after_create]
  #     actions = [action.bcm_cmdevice_power.boot_new_worker]
  #   }
  # }
}

# Action to power on the newly created device
action "bcm_cmdevice_power" "boot_new_worker" {
  device_id    = bcm_cmdevice_device.new_worker.uuid
  power_action = "power_on"
}

# Output the device UUID for reference
output "device_uuid" {
  description = "UUID of the created device"
  value       = bcm_cmdevice_device.new_worker.uuid
}
