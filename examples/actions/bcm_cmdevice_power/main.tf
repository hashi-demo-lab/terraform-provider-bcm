# Example: BCM CMDevice Power Action
#
# This example demonstrates how to use the bcm_cmdevice_power action
# to execute power operations on BCM-managed devices.
#
# Requirements:
# - Terraform 1.14 or later
# - BCM API endpoint and credentials
# - Target device UUID or hostname

terraform {
  required_providers {
    bcm = {
      source = "hashicorp/bcm"
    }
  }
  required_version = ">= 1.14.0"
}

provider "bcm" {
  # Configuration via environment variables:
  # BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD
  # Or set directly:
  # endpoint             = "https://172.21.15.254:8081"
  # username             = "root"
  # password             = var.bcm_password
  insecure_skip_verify = true
}

# Example 1: Power on a device by UUID
action "bcm_cmdevice_power" "power_on_by_uuid" {
  device_id    = var.device_uuid
  power_action = "power_on"
}

# Example 2: Reboot a device by hostname
action "bcm_cmdevice_power" "reboot_by_hostname" {
  device_id    = var.device_hostname
  power_action = "reboot"
}

# Example 3: Power off with wait for completion (future feature)
action "bcm_cmdevice_power" "shutdown" {
  device_id           = var.device_uuid
  power_action        = "power_off"
  wait_for_completion = true
  timeout             = "2m"
}

# Example 4: Power cycle a device
action "bcm_cmdevice_power" "power_cycle" {
  device_id    = var.device_uuid
  power_action = "power_cycle"
}
