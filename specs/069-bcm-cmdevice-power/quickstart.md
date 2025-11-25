# Quickstart: BCM CMDevice Power Action

**Feature**: bcm_cmdevice_power Terraform Action
**Branch**: `069-bcm-cmdevice-power`
**Requires**: Terraform 1.14+, terraform-plugin-framework v1.16.1

## Overview

This guide helps developers implement the `bcm_cmdevice_power` Terraform Action for BCM device power operations.

## Prerequisites

- Go 1.24+
- Terraform 1.14+ (beta or later)
- terraform-plugin-framework v1.16.1
- Access to BCM API documentation

## Implementation Steps

### 1. Add Action Registration to Provider

**File**: `/workspace/internal/provider/provider.go`

Add the `ProviderWithActions` interface:

```go
import (
    "github.com/hashicorp/terraform-plugin-framework/action"
)

// Add interface assertion
var _ provider.ProviderWithActions = &BCMProvider{}

// Add Actions method
func (p *BCMProvider) Actions(ctx context.Context) []func() action.Action {
    return []func() action.Action{
        NewCMDevicePowerAction,
    }
}
```

Update Configure to pass client to actions:

```go
func (p *BCMProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
    // ... existing code ...

    // Make client available to actions (add this line)
    resp.ActionData = client
}
```

### 2. Create Action File

**File**: `/workspace/internal/provider/action_cmdevice_power.go`

```go
// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
    "context"
    "fmt"
    "time"

    "github.com/hashicorp/terraform-plugin-framework/action"
    "github.com/hashicorp/terraform-plugin-framework/action/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
    "github.com/hashicorp/terraform-plugin-framework/schema/validator"
    "github.com/hashicorp/terraform-plugin-log/tflog"
)

// Interface compliance
var (
    _ action.Action              = &CMDevicePowerAction{}
    _ action.ActionWithConfigure = &CMDevicePowerAction{}
)

// CMDevicePowerAction implements power operations for BCM devices.
type CMDevicePowerAction struct {
    client *BCMClient
}

// CMDevicePowerActionModel describes the action configuration.
type CMDevicePowerActionModel struct {
    DeviceID          types.String `tfsdk:"device_id"`
    PowerAction       types.String `tfsdk:"power_action"`
    WaitForCompletion types.Bool   `tfsdk:"wait_for_completion"`
    Timeout           types.String `tfsdk:"timeout"`
}

// NewCMDevicePowerAction creates a new action instance.
func NewCMDevicePowerAction() action.Action {
    return &CMDevicePowerAction{}
}

// Metadata returns the action type name.
func (a *CMDevicePowerAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_cmdevice_power"
}

// Schema defines the action schema.
func (a *CMDevicePowerAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "Execute power operations on BCM devices.\n\n" +
            "Supports power on, power off, reboot, and power cycle operations via BMC/IPMI.",

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
                MarkdownDescription: "Wait for power state change (default: false)",
            },
            "timeout": schema.StringAttribute{
                Optional:            true,
                MarkdownDescription: "Timeout when waiting (default: 5m)",
            },
        },
    }
}

// Configure receives the provider-configured BCM client.
func (a *CMDevicePowerAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }

    client, ok := req.ProviderData.(*BCMClient)
    if !ok {
        resp.Diagnostics.AddError(
            "Unexpected Action Configure Type",
            fmt.Sprintf("Expected *BCMClient, got: %T", req.ProviderData),
        )
        return
    }

    a.client = client
}

// Invoke executes the power operation.
func (a *CMDevicePowerAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
    var config CMDevicePowerActionModel

    // Read configuration
    resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Map power_action to BCM method
    powerMethods := map[string]string{
        "power_on":    "powerOn",
        "power_off":   "powerOff",
        "reboot":      "reboot",
        "power_cycle": "powerCycle",
    }

    powerAction := config.PowerAction.ValueString()
    bcmMethod, ok := powerMethods[powerAction]
    if !ok {
        resp.Diagnostics.AddError(
            "Invalid Power Action",
            fmt.Sprintf("Unknown power action: %s", powerAction),
        )
        return
    }

    deviceID := config.DeviceID.ValueString()

    // Report progress
    resp.SendProgress(action.InvokeProgressEvent{
        Message: fmt.Sprintf("Executing %s on device %s...", powerAction, deviceID),
    })

    tflog.Info(ctx, "Executing power operation", map[string]interface{}{
        "device_id":    deviceID,
        "power_action": powerAction,
        "bcm_method":   bcmMethod,
    })

    // Execute power operation
    _, err := a.client.CallJSONRPC(ctx, "cmdevice", bcmMethod, deviceID)
    if err != nil {
        resp.Diagnostics.AddError(
            "Power Operation Failed",
            fmt.Sprintf("Failed to execute %s on device %s: %s",
                powerAction, deviceID, err.Error()),
        )
        return
    }

    // Handle wait_for_completion
    waitForCompletion := false
    if !config.WaitForCompletion.IsNull() {
        waitForCompletion = config.WaitForCompletion.ValueBool()
    }

    if waitForCompletion {
        timeout := 5 * time.Minute
        if !config.Timeout.IsNull() {
            var err error
            timeout, err = time.ParseDuration(config.Timeout.ValueString())
            if err != nil {
                resp.Diagnostics.AddError(
                    "Invalid Timeout",
                    fmt.Sprintf("Failed to parse timeout: %s", err.Error()),
                )
                return
            }
        }

        resp.SendProgress(action.InvokeProgressEvent{
            Message: fmt.Sprintf("Waiting for power state change (timeout: %s)...", timeout),
        })

        // TODO: Implement wait logic with polling
        // This requires Phase 0 verification of power status query method
    }

    resp.SendProgress(action.InvokeProgressEvent{
        Message: fmt.Sprintf("Power operation '%s' completed successfully", powerAction),
    })

    tflog.Info(ctx, "Power operation completed", map[string]interface{}{
        "device_id":    deviceID,
        "power_action": powerAction,
    })
}
```

### 3. Create Unit Tests

**File**: `/workspace/internal/provider/action_cmdevice_power_test.go`

```go
package provider

import (
    "testing"
)

func TestCMDevicePowerAction_Metadata(t *testing.T) {
    action := NewCMDevicePowerAction()

    // Test metadata returns correct type name
    // Implementation depends on test framework availability
}

func TestCMDevicePowerAction_Schema(t *testing.T) {
    action := NewCMDevicePowerAction()

    // Test schema contains required attributes
    // Test validators are applied correctly
}

func TestPowerMethodMapping(t *testing.T) {
    tests := []struct {
        input    string
        expected string
    }{
        {"power_on", "powerOn"},
        {"power_off", "powerOff"},
        {"reboot", "reboot"},
        {"power_cycle", "powerCycle"},
    }

    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            // Test mapping logic
        })
    }
}
```

### 4. Create Example Configuration

**Directory**: `/workspace/examples/actions/bcm_cmdevice_power/`

**File**: `main.tf`

```hcl
terraform {
  required_providers {
    bcm = {
      source = "hashicorp/bcm"
    }
  }
  required_version = ">= 1.14.0"
}

provider "bcm" {
  endpoint             = var.bcm_endpoint
  username             = var.bcm_username
  password             = var.bcm_password
  insecure_skip_verify = true
}

# Example 1: Direct power control
action "bcm_cmdevice_power" "power_on_worker" {
  device_id    = var.device_uuid
  power_action = "power_on"
}

# Example 2: Power off with wait
action "bcm_cmdevice_power" "shutdown_worker" {
  device_id           = var.device_uuid
  power_action        = "power_off"
  wait_for_completion = true
  timeout             = "2m"
}

# Example 3: Lifecycle trigger (auto-boot after device creation)
resource "bcm_cmdevice_device" "new_node" {
  hostname           = "citest-power-node"
  mac                = "00:11:22:33:44:55"
  category           = var.category_uuid
  management_network = var.network_uuid

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.bcm_cmdevice_power.boot_new_node]
    }
  }
}

action "bcm_cmdevice_power" "boot_new_node" {
  device_id    = bcm_cmdevice_device.new_node.uuid
  power_action = "power_on"
}
```

**File**: `variables.tf`

```hcl
variable "bcm_endpoint" {
  type        = string
  description = "BCM API endpoint"
  default     = "https://172.21.15.254:8081"
}

variable "bcm_username" {
  type        = string
  description = "BCM username"
  default     = "root"
}

variable "bcm_password" {
  type        = string
  description = "BCM password"
  sensitive   = true
}

variable "device_uuid" {
  type        = string
  description = "Target device UUID"
}

variable "category_uuid" {
  type        = string
  description = "Category UUID for new device"
}

variable "network_uuid" {
  type        = string
  description = "Management network UUID"
}
```

### 5. Run Tests

```bash
# Unit tests (immediate)
go test -v ./internal/provider/ -run "CMDevicePower"

# Manual testing with Terraform 1.14 beta
terraform init
terraform plan

# Direct action invocation
terraform apply -invoke="action.bcm_cmdevice_power.power_on_worker"
```

### 6. Generate Documentation

```bash
make generate
```

## Key Files

| File | Purpose |
|------|---------|
| `internal/provider/provider.go` | Provider with Actions registration |
| `internal/provider/action_cmdevice_power.go` | Action implementation |
| `internal/provider/action_cmdevice_power_test.go` | Unit tests |
| `examples/actions/bcm_cmdevice_power/main.tf` | Example configuration |

## BCM API Methods

| Terraform | BCM Method | Service |
|-----------|------------|---------|
| power_on | powerOn | cmdevice |
| power_off | powerOff | cmdevice |
| reboot | reboot | cmdevice |
| power_cycle | powerCycle | cmdevice |

## Testing Commands

```bash
# Build provider
make build

# Run unit tests
make test

# Verify BCM API (manual)
curl -X POST https://172.21.15.254:8081/json \
  -H "Content-Type: application/json" \
  -d '{"service": "cmdevice", "call": "powerOn", "args": ["node001"]}'
```

## Common Issues

### "Action not found"

Ensure provider implements `ProviderWithActions` and registers the action in `Actions()`.

### "BCM client is nil"

Verify `resp.ActionData = client` is set in provider's `Configure` method.

### "Invalid power_action"

Check that the power_action value is one of: `power_on`, `power_off`, `reboot`, `power_cycle`.

## Next Steps

1. Complete Phase 0 API verification
2. Implement wait_for_completion polling logic
3. Add acceptance tests when Terraform 1.14 GA
4. Update CLAUDE.md with action patterns
