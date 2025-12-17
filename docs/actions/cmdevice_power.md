---
page_title: "bcm_cmdevice_power Action - bcm"
subcategory: ""
description: |-
  Execute power operations on BCM-managed devices.
---

# bcm_cmdevice_power (Action)

~> **Note:** Actions require Terraform 1.14 or later.

Execute power operations on BCM-managed devices. This action allows you to power on, power off, reboot, or power cycle devices managed by BCM.

## Example Usage

```terraform
# Power on a device by UUID
action "bcm_cmdevice_power" "power_on" {
  device_id    = "device-uuid-here"
  power_action = "power_on"
}

# Reboot a device
action "bcm_cmdevice_power" "reboot" {
  device_id    = var.device_hostname
  power_action = "reboot"
}

# Power cycle a device
action "bcm_cmdevice_power" "power_cycle" {
  device_id    = bcm_cmdevice_device.worker.uuid
  power_action = "power_cycle"
}
```

## Argument Reference

The following arguments are supported:

* `device_id` - (Required) The UUID or hostname of the device to perform the power action on.
* `power_action` - (Required) The power action to perform. Valid values are:
  * `power_on` - Power on the device
  * `power_off` - Power off the device
  * `reboot` - Reboot the device
  * `power_cycle` - Power cycle the device (off then on)
* `wait_for_completion` - (Optional) Whether to wait for the power action to complete. Defaults to `false`.
* `timeout` - (Optional) Timeout for the power action. Defaults to `2m`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `status` - The status of the power action execution.
