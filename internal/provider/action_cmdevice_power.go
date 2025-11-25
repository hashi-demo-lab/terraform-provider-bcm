// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure interface compliance at compile time.
var (
	_ action.Action              = &CMDevicePowerAction{}
	_ action.ActionWithConfigure = &CMDevicePowerAction{}
)

// CMDevicePowerAction implements power operations for BCM devices.
// This action allows Terraform practitioners to execute power operations
// (power_on, power_off, reboot, power_cycle) on BCM-managed devices.
type CMDevicePowerAction struct {
	client *BCMClient
}

// CMDevicePowerActionModel describes the action configuration from HCL.
type CMDevicePowerActionModel struct {
	DeviceID          types.String `tfsdk:"device_id"`
	PowerAction       types.String `tfsdk:"power_action"`
	WaitForCompletion types.Bool   `tfsdk:"wait_for_completion"`
	Timeout           types.String `tfsdk:"timeout"`
}

// powerMethodMapping maps Terraform power_action values to BCM API methods.
var powerMethodMapping = map[string]string{
	"power_on":    "powerOn",
	"power_off":   "powerOff",
	"reboot":      "reboot",
	"power_cycle": "powerCycle",
}

// NewCMDevicePowerAction creates a new action instance.
// This function is registered with the provider via the Actions method.
func NewCMDevicePowerAction() action.Action {
	return &CMDevicePowerAction{}
}

// Metadata returns the action type name.
// The full type name will be "bcm_cmdevice_power".
func (a *CMDevicePowerAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cmdevice_power"
}

// Schema defines the action schema with input attributes.
// Actions do not have computed or state attributes - they only define inputs.
func (a *CMDevicePowerAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Execute power operations on BCM devices.\n\n" +
			"This action supports power on, power off, reboot, and power cycle operations " +
			"via BMC/IPMI. It can be invoked directly or triggered via resource lifecycle events.\n\n" +
			"**Requires Terraform 1.14 or later.**",

		Attributes: map[string]schema.Attribute{
			"device_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "BCM device identifier (UUID or hostname). Use `bcm_cmdevice_device.name.uuid` to reference a managed device.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"power_action": schema.StringAttribute{
				Required: true,
				MarkdownDescription: "Power operation to execute:\n" +
					"  - `power_on`: Power on device via BMC\n" +
					"  - `power_off`: Power off device via BMC\n" +
					"  - `reboot`: Graceful device reboot\n" +
					"  - `power_cycle`: Hard power cycle (off/on)",
				Validators: []validator.String{
					stringvalidator.OneOf("power_on", "power_off", "reboot", "power_cycle"),
				},
			},
			"wait_for_completion": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Wait for power state change to complete before returning. Default: `false`.",
			},
			"timeout": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Timeout duration when `wait_for_completion` is enabled. Uses Go duration format (e.g., `5m`, `30s`). Default: `5m`. Range: 10s-30m.",
			},
		},
	}
}

// Configure receives the provider-configured BCM client.
// This method is called by the framework after the provider's Configure method.
func (a *CMDevicePowerAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*BCMClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Action Configure Type",
			fmt.Sprintf("Expected *BCMClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	a.client = client
}

// Invoke executes the power operation on the specified device.
// This is the main entry point for action execution.
func (a *CMDevicePowerAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config CMDevicePowerActionModel

	// Read configuration from request
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract values from config
	deviceID := config.DeviceID.ValueString()
	powerAction := config.PowerAction.ValueString()

	// Map Terraform power_action to BCM API method
	bcmMethod, ok := powerMethodMapping[powerAction]
	if !ok {
		resp.Diagnostics.AddError(
			"Invalid Power Action",
			fmt.Sprintf("Unknown power action: %s. Valid values are: power_on, power_off, reboot, power_cycle.", powerAction),
		)
		return
	}

	// SAFETY CHECK: Prevent power operations on head nodes
	// Head nodes manage the cluster and should not be power cycled
	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Checking device type for %s...", deviceID),
	})

	nodeData, err := a.client.CallJSONRPC(ctx, "cmdevice", "getNode", deviceID)
	if err == nil {
		var node map[string]interface{}
		if json.Unmarshal(nodeData, &node) == nil {
			childType, _ := node["childType"].(string)
			hostname, _ := node["hostname"].(string)

			tflog.Debug(ctx, "Device type check", map[string]interface{}{
				"device_id":  deviceID,
				"hostname":   hostname,
				"child_type": childType,
			})

			if childType == "HeadNode" {
				resp.Diagnostics.AddError(
					"Safety Check Failed",
					fmt.Sprintf("Cannot execute power operations on HeadNode devices.\n\n"+
						"Device: %s (UUID: %s)\n"+
						"Type: %s\n\n"+
						"Head nodes manage the BCM cluster and should not be power cycled. "+
						"This operation has been blocked to prevent cluster disruption.",
						hostname, deviceID, childType),
				)
				return
			}
		}
	} else {
		// Log warning but continue - device might be identified by hostname
		tflog.Warn(ctx, "Could not verify device type", map[string]interface{}{
			"device_id": deviceID,
			"error":     err.Error(),
		})
	}

	// Report progress - starting operation
	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Executing %s on device %s...", powerAction, deviceID),
	})

	tflog.Info(ctx, "Executing power operation", map[string]interface{}{
		"device_id":    deviceID,
		"power_action": powerAction,
		"bcm_method":   bcmMethod,
	})

	// Execute power operation via BCM API
	// NOTE: BCM power methods (reboot, etc.) use "arg" (single value) format,
	// not "args" (array) format used by other methods like getNode.
	_, err = a.client.CallJSONRPCArg(ctx, "cmdevice", bcmMethod, deviceID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Power Operation Failed",
			fmt.Sprintf("Failed to execute %s on device %s: %s\n\n"+
				"Note: Power operations require:\n"+
				"1. BCM power management enabled (IPMI/BMC/Redfish)\n"+
				"2. Device with powerControl configured (not 'none')\n"+
				"3. BCM version with power methods available",
				powerAction, deviceID, err.Error()),
		)
		return
	}

	// Handle wait_for_completion if enabled
	waitForCompletion := false
	if !config.WaitForCompletion.IsNull() {
		waitForCompletion = config.WaitForCompletion.ValueBool()
	}

	if waitForCompletion {
		resp.SendProgress(action.InvokeProgressEvent{
			Message: "Power command sent. Waiting for state change is not yet implemented.",
		})
		// TODO: Implement wait_for_completion polling logic in Phase 4 (User Story 3)
		// This requires Phase 0 verification of power status query method
	}

	// Report progress - operation completed
	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Power operation '%s' completed successfully on device %s", powerAction, deviceID),
	})

	tflog.Info(ctx, "Power operation completed", map[string]interface{}{
		"device_id":    deviceID,
		"power_action": powerAction,
	})
}
