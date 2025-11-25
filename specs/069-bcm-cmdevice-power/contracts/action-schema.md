# bcm_cmdevice_power Action Schema Contract

**Action Type**: bcm_cmdevice_power
**Provider**: bcm
**Terraform Version**: 1.14+
**Framework Version**: terraform-plugin-framework v1.16.1

## Overview

This contract defines the Terraform configuration schema for the `bcm_cmdevice_power` action.

## HCL Configuration

```hcl
action "bcm_cmdevice_power" "example" {
  device_id           = "uuid-or-hostname"
  power_action        = "power_on"
  wait_for_completion = false
  timeout             = "5m"
}
```

## Schema Definition

### Attributes

| Attribute | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `device_id` | string | Yes | - | BCM device identifier (UUID or hostname) |
| `power_action` | string | Yes | - | Power operation to perform |
| `wait_for_completion` | bool | No | `false` | Wait for power state change |
| `timeout` | string | No | `"5m"` | Timeout when waiting for completion |

### device_id

**Type**: `string`
**Required**: Yes

The BCM device identifier. Accepts either:
- Device UUID (RFC 4122 format): `"2870c0b0-6fda-4026-9b8f-28be4c372fee"`
- Device hostname: `"node001"`

**Validation**:
- Must not be empty
- If UUID format, must be valid RFC 4122

**Examples**:
```hcl
# Using UUID
device_id = "2870c0b0-6fda-4026-9b8f-28be4c372fee"

# Using hostname
device_id = "node001"

# Using reference from resource
device_id = bcm_cmdevice_device.worker.uuid
```

### power_action

**Type**: `string`
**Required**: Yes

The power operation to perform on the device.

**Allowed Values**:
| Value | Description |
|-------|-------------|
| `power_on` | Power on the device via BMC/IPMI |
| `power_off` | Power off the device via BMC/IPMI |
| `reboot` | Graceful reboot of the device |
| `power_cycle` | Hard power cycle (off then on) |

**Validation**:
- Schema-level validator enforces allowed values
- Invalid values rejected at plan time

**Examples**:
```hcl
power_action = "power_on"
power_action = "power_off"
power_action = "reboot"
power_action = "power_cycle"
```

### wait_for_completion

**Type**: `bool`
**Required**: No
**Default**: `false`

When `true`, the action blocks until the power state change is confirmed or the timeout is reached.

When `false`, the action returns immediately after sending the power command to BCM.

**Examples**:
```hcl
# Default: fire and forget
wait_for_completion = false

# Wait for confirmation
wait_for_completion = true
```

### timeout

**Type**: `string`
**Required**: No
**Default**: `"5m"`

Maximum duration to wait when `wait_for_completion = true`. Ignored when `wait_for_completion = false`.

**Format**: Go duration string (e.g., "30s", "5m", "1h")

**Validation**:
- Must be valid Go duration format
- Minimum: 10s
- Maximum: 30m

**Examples**:
```hcl
timeout = "30s"   # 30 seconds
timeout = "5m"    # 5 minutes (default)
timeout = "10m"   # 10 minutes
```

---

## Go Schema Implementation

```go
import (
    "github.com/hashicorp/terraform-plugin-framework/action/schema"
    "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
    "github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (a *CMDevicePowerAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "Execute power operations on BCM devices.\n\n" +
            "Supports power on, power off, reboot, and power cycle operations. " +
            "Can be invoked directly or triggered by resource lifecycle events.",

        Attributes: map[string]schema.Attribute{
            "device_id": schema.StringAttribute{
                Required:            true,
                MarkdownDescription: "BCM device identifier (UUID or hostname)",
                Validators: []validator.String{
                    stringvalidator.LengthAtLeast(1),
                },
            },
            "power_action": schema.StringAttribute{
                Required:            true,
                MarkdownDescription: "Power operation: `power_on`, `power_off`, `reboot`, `power_cycle`",
                Validators: []validator.String{
                    stringvalidator.OneOf("power_on", "power_off", "reboot", "power_cycle"),
                },
            },
            "wait_for_completion": schema.BoolAttribute{
                Optional:            true,
                MarkdownDescription: "Wait for power state change confirmation (default: false)",
            },
            "timeout": schema.StringAttribute{
                Optional:            true,
                MarkdownDescription: "Timeout duration when waiting (default: 5m). Format: Go duration (30s, 5m, 1h)",
                Validators: []validator.String{
                    // Custom duration validator
                },
            },
        },
    }
}
```

---

## Configuration Model

```go
// CMDevicePowerActionModel describes the action configuration.
type CMDevicePowerActionModel struct {
    DeviceID          types.String `tfsdk:"device_id"`
    PowerAction       types.String `tfsdk:"power_action"`
    WaitForCompletion types.Bool   `tfsdk:"wait_for_completion"`
    Timeout           types.String `tfsdk:"timeout"`
}
```

---

## Usage Examples

### Direct Invocation

```hcl
action "bcm_cmdevice_power" "shutdown_worker" {
  device_id    = bcm_cmdevice_device.worker.uuid
  power_action = "power_off"
}
```

```bash
# Invoke the action directly
terraform apply -invoke="action.bcm_cmdevice_power.shutdown_worker"
```

### Lifecycle Trigger (Auto-boot after creation)

```hcl
resource "bcm_cmdevice_device" "knode" {
  hostname           = "knode-01"
  mac                = "00:11:22:33:44:55"
  category           = bcm_cmdevice_category.k8s.uuid
  management_network = bcm_cmnet_network.mgmt.uuid

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.bcm_cmdevice_power.boot_node]
    }
  }
}

action "bcm_cmdevice_power" "boot_node" {
  device_id    = bcm_cmdevice_device.knode.uuid
  power_action = "power_on"
}
```

### Wait for Power State

```hcl
action "bcm_cmdevice_power" "boot_with_wait" {
  device_id           = bcm_cmdevice_device.worker.uuid
  power_action        = "power_on"
  wait_for_completion = true
  timeout             = "10m"
}
```

### Power Cycle Before Maintenance

```hcl
action "bcm_cmdevice_power" "maintenance_cycle" {
  device_id    = var.maintenance_node_uuid
  power_action = "power_cycle"
}
```

---

## Validation Behavior

### Plan-Time Validation

These errors are caught at `terraform plan`:

1. **Invalid power_action**:
   ```
   Error: Invalid Attribute Value Match

   Attribute power_action value must be one of: ["power_on" "power_off" "reboot"
   "power_cycle"], got: "shutdown"
   ```

2. **Empty device_id**:
   ```
   Error: Invalid Attribute Value Length

   Attribute device_id string length must be at least 1
   ```

### Apply-Time Validation

These errors occur during `terraform apply -invoke`:

1. **Device not found**:
   ```
   Error: Power Operation Failed

   Device 'nonexistent-node' not found in BCM
   ```

2. **BMC unreachable**:
   ```
   Error: Power Operation Failed

   Failed to contact BMC for device 'node001': Connection timed out
   ```

3. **Timeout exceeded**:
   ```
   Warning: Wait Timeout

   Power command sent successfully but confirmation timed out after 5m0s
   ```
