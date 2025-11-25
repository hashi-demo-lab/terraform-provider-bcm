# Terraform Actions Research: Node Power Operations

**Research Date:** 2025-11-25
**Purpose:** Evaluate Terraform Actions for BCM node power operations (power on/off/cycle/reboot)

## Executive Summary

Terraform Actions (v1.14+) are the **ideal solution** for node power operations. They're designed specifically for imperative, non-CRUD operations like "power off this server" without managing state.

**Key Finding:** The `terraform-plugin-framework` v1.16.1 (already installed) includes full action support.

## Terraform Actions Overview

### What Are Actions?

Actions are a new Terraform block type for operations that don't fit CRUD patterns:
- Triggering webhooks
- Power management operations
- Cache invalidation
- Service restarts
- Disaster recovery tasks

### Version Requirements

| Component | Minimum Version | Current |
|-----------|-----------------|---------|
| Terraform CLI | 1.14.0 | Beta |
| terraform-plugin-framework | 1.16.0 | **1.16.1** ✅ |

**Status:** Technical preview - no compatibility guarantees until Terraform 1.14 GA

## Implementation Architecture

### Provider Interface

```go
// provider.go - implement ProviderWithActions
type ProviderWithActions interface {
    Provider

    // Actions returns a slice of functions to instantiate each Action
    Actions(context.Context) []func() action.Action
}
```

### Action Interface

```go
// action.Action interface (3 required methods)
type Action interface {
    // Schema returns the action's configuration schema
    Schema(context.Context, SchemaRequest, *SchemaResponse)

    // Metadata returns the full action name (e.g., bcm_device_power)
    Metadata(context.Context, MetadataRequest, *MetadataResponse)

    // Invoke executes the action logic
    Invoke(context.Context, InvokeRequest, *InvokeResponse)
}
```

### Optional Interfaces

```go
// Add provider configuration (BCM client)
type ActionWithConfigure interface {
    Action
    Configure(context.Context, ConfigureRequest, *ConfigureResponse)
}

// Add validation
type ActionWithValidateConfig interface {
    Action
    ValidateConfig(context.Context, ValidateConfigRequest, *ValidateConfigResponse)
}
```

## Proposed BCM Power Action Design

### Action: `bcm_cmdevice_power`

```hcl
# Terraform configuration syntax
action "bcm_cmdevice_power" "worker_shutdown" {
  device_id    = bcm_cmdevice_device.worker[0].uuid
  power_action = "power_off"  # power_on, power_off, reboot, power_cycle
}
```

### Invocation Methods

**1. CLI Direct Invocation:**
```bash
terraform apply -invoke="action.bcm_cmdevice_power.worker_shutdown"
```

**2. Lifecycle Trigger (after resource changes):**
```hcl
resource "bcm_cmdevice_device" "worker" {
  hostname = "worker-01"
  category = bcm_cmdevice_category.compute.uuid

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

### Implementation Skeleton

```go
// internal/provider/action_cmdevice_power.go

package provider

import (
    "context"
    "github.com/hashicorp/terraform-plugin-framework/action"
    "github.com/hashicorp/terraform-plugin-framework/action/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

var (
    _ action.Action              = &CMDevicePowerAction{}
    _ action.ActionWithConfigure = &CMDevicePowerAction{}
)

type CMDevicePowerAction struct {
    client *BCMClient
}

type CMDevicePowerActionModel struct {
    DeviceID    types.String `tfsdk:"device_id"`
    PowerAction types.String `tfsdk:"power_action"`
}

func NewCMDevicePowerAction() action.Action {
    return &CMDevicePowerAction{}
}

func (a *CMDevicePowerAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_cmdevice_power"
}

func (a *CMDevicePowerAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "Execute power operations on BCM devices",
        Attributes: map[string]schema.Attribute{
            "device_id": schema.StringAttribute{
                Required:            true,
                MarkdownDescription: "UUID of the device to control",
            },
            "power_action": schema.StringAttribute{
                Required:            true,
                MarkdownDescription: "Power action: power_on, power_off, reboot, power_cycle",
            },
        },
    }
}

func (a *CMDevicePowerAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }
    client, ok := req.ProviderData.(*BCMClient)
    if !ok {
        resp.Diagnostics.AddError("Unexpected Provider Data", "...")
        return
    }
    a.client = client
}

func (a *CMDevicePowerAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
    var config CMDevicePowerActionModel
    resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Progress reporting
    resp.SendProgress(action.InvokeProgressEvent{
        Message: fmt.Sprintf("Executing %s on device %s",
            config.PowerAction.ValueString(),
            config.DeviceID.ValueString()),
    })

    // Call BCM API
    _, err := a.client.CallJSONRPC(ctx, "cmdevice", config.PowerAction.ValueString(),
        config.DeviceID.ValueString())
    if err != nil {
        resp.Diagnostics.AddError("Power Operation Failed", err.Error())
        return
    }

    resp.SendProgress(action.InvokeProgressEvent{
        Message: "Power operation completed successfully",
    })
}
```

### Provider Registration

```go
// internal/provider/provider.go

var _ provider.ProviderWithActions = &BCMProvider{}

func (p *BCMProvider) Actions(ctx context.Context) []func() action.Action {
    return []func() action.Action{
        NewCMDevicePowerAction,
    }
}
```

## BCM API Methods (to verify)

| Power Action | Expected BCM Method | Service |
|--------------|---------------------|---------|
| `power_on` | `powerOn` | cmdevice |
| `power_off` | `powerOff` | cmdevice |
| `reboot` | `reboot` | cmdevice |
| `power_cycle` | `powerCycle` | cmdevice |
| `power_status` | `powerStatus` | cmdevice |

**Note:** Need to verify exact method names via BCM API exploration.

## DGX BasePOD Use Case

```hcl
# After creating control plane nodes, power them on for PXE boot
resource "bcm_cmdevice_device" "knode" {
  count    = 3
  hostname = "knode-0${count.index + 1}"
  category = bcm_cmdevice_category.k8s_control_plane.uuid

  interfaces { ... }

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.bcm_cmdevice_power.boot_knodes[count.index]]
    }
  }
}

action "bcm_cmdevice_power" "boot_knodes" {
  count        = 3
  device_id    = bcm_cmdevice_device.knode[count.index].uuid
  power_action = "power_on"
}
```

## Alternative: Ephemeral Resources

If actions don't fit, ephemeral resources could work for power status checks:

```hcl
ephemeral "bcm_cmdevice_power_status" "worker" {
  device_id = bcm_cmdevice_device.worker.uuid
}

# Use ephemeral output in other resources
# Status not stored in state (sensitive operation)
```

## Comparison: Actions vs Ephemeral vs Resource

| Feature | Action | Ephemeral Resource | Resource |
|---------|--------|-------------------|----------|
| State storage | No | No | Yes |
| CRUD lifecycle | No | Open/Renew/Close | Yes |
| Imperative operations | **Yes** | Limited | No |
| Terraform version | 1.14+ | 1.10+ | All |
| Idempotency required | No | No | Yes |
| Power operations | **Best fit** | Possible | Poor fit |

## Recommendations

1. **Use Actions for power operations** - Perfect semantic fit
2. **Phase API research first** - Verify cmdevice power methods
3. **Wait for Terraform 1.14 GA** - Currently beta, no compatibility guarantees
4. **Consider fallback** - null_resource + local-exec as interim solution

## Sources

- [Ephemeral Resources - HashiCorp Developer](https://developer.hashicorp.com/terraform/plugin/framework/ephemeral-resources)
- [Plugin Framework Benefits](https://developer.hashicorp.com/terraform/plugin/framework-benefits)
- [Terraform Actions Introduction - DanielMSchmidt.de](https://danielmschmidt.de/posts/2025-09-26-terraform-actions-introduction/)
- [Terraform Actions - Mike Guy](https://mikeguy.co.uk/posts/terraform-actions/)
- [terraform-plugin-framework GitHub](https://github.com/hashicorp/terraform-plugin-framework)
- [HashiDays 2025 Announcements](https://www.hashicorp.com/en/blog/terraform-ephemeral-resources-waypoint-actions-and-more-at-hashidays-2025)
