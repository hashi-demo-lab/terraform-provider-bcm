# BCM CMDevice Power Action Example

This example demonstrates how to use the `bcm_cmdevice_power` action to execute power operations on BCM-managed devices.

## Requirements

- Terraform 1.14 or later (actions are a new Terraform 1.14 feature)
- BCM API endpoint and credentials
- Target device UUID or hostname

## Usage

### 1. Set Environment Variables

```bash
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="your-password"
export TF_VAR_device_uuid="2870c0b0-6fda-4026-9b8f-28be4c372fee"
export TF_VAR_device_hostname="node001"
```

### 2. Initialize Terraform

```bash
terraform init
```

### 3. Plan and Validate

```bash
terraform plan
```

### 4. Invoke Action Directly

Actions can be invoked directly using the `-invoke` flag:

```bash
# Power on a device
terraform apply -invoke="action.bcm_cmdevice_power.power_on_by_uuid"

# Reboot a device
terraform apply -invoke="action.bcm_cmdevice_power.reboot_by_hostname"

# Power cycle a device
terraform apply -invoke="action.bcm_cmdevice_power.power_cycle"
```

## Power Actions

| Action | Description |
|--------|-------------|
| `power_on` | Power on device via BMC/IPMI |
| `power_off` | Power off device via BMC/IPMI |
| `reboot` | Graceful device reboot |
| `power_cycle` | Hard power cycle (off/on) |

## Device Identification

The `device_id` attribute accepts either:
- **UUID**: `"2870c0b0-6fda-4026-9b8f-28be4c372fee"`
- **Hostname**: `"node001"`

## Lifecycle Triggers (Future)

Actions can be triggered by resource lifecycle events:

```hcl
resource "bcm_cmdevice_device" "worker" {
  hostname = "worker-01"
  # ... other configuration

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.bcm_cmdevice_power.boot_worker]
    }
  }
}

action "bcm_cmdevice_power" "boot_worker" {
  device_id    = bcm_cmdevice_device.worker.uuid
  power_action = "power_on"
}
```

## Wait for Completion (Future)

The `wait_for_completion` attribute is reserved for future functionality:

```hcl
action "bcm_cmdevice_power" "shutdown" {
  device_id           = var.device_uuid
  power_action        = "power_off"
  wait_for_completion = true  # Wait for device to power off
  timeout             = "2m"  # Timeout after 2 minutes
}
```

## Notes

- Actions do not maintain state - they execute side effects
- Each invocation is independent
- BMC/IPMI must be configured on the target device for power operations to succeed
