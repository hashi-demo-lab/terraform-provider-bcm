// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource                = &CMDeviceDeviceResource{}
	_ resource.ResourceWithImportState = &CMDeviceDeviceResource{}
)

// CMDeviceDeviceResource defines the resource implementation
type CMDeviceDeviceResource struct {
	client *BCMClient
}

// CMDeviceDeviceResourceModel describes the resource data model
type CMDeviceDeviceResourceModel struct {
	// Identity fields (required/computed)
	ID       types.String `tfsdk:"id"`       // Computed, same as UUID
	UUID     types.String `tfsdk:"uuid"`     // Computed, BCM-assigned
	Hostname types.String `tfsdk:"hostname"` // Required, RFC 1123 validation
	MAC      types.String `tfsdk:"mac"`      // Required, MAC address validation

	// References (required)
	Category          types.String `tfsdk:"category"`           // Required, UUID reference
	ManagementNetwork types.String `tfsdk:"management_network"` // Required, UUID reference
	Partition         types.String `tfsdk:"partition"`          // Optional, UUID reference (uses default if not set)

	// Optional configuration
	Notes              types.String `tfsdk:"notes"`                // Optional
	KernelParameters   types.String `tfsdk:"kernel_parameters"`    // Optional
	BootLoader         types.String `tfsdk:"boot_loader"`          // Optional
	BootLoaderProtocol types.String `tfsdk:"boot_loader_protocol"` // Optional
	Force              types.Bool   `tfsdk:"force"`                // Optional, default: false

	// Computed fields
	CreationTime types.Int64  `tfsdk:"creation_time"` // Computed
	BaseType     types.String `tfsdk:"base_type"`     // Computed, always "Device"
	ChildType    types.String `tfsdk:"child_type"`    // Computed, BCM-determined
}

// NewCMDeviceDeviceResource creates a new resource instance
func NewCMDeviceDeviceResource() resource.Resource {
	return &CMDeviceDeviceResource{}
}

// Metadata returns the resource type name
func (r *CMDeviceDeviceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cmdevice_device"
}

// Schema defines the resource schema
func (r *CMDeviceDeviceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a BCM device (compute node) in the cluster.\n\n" +
			"Devices represent physical or virtual nodes that can be provisioned and managed by BCM. " +
			"Each device requires a unique hostname, MAC address, category assignment, and management network configuration.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Device identifier (same as UUID)",
			},
			"uuid": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Device UUID assigned by BCM",
			},
			"hostname": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Device hostname (RFC 1123 DNS label: lowercase alphanumeric and hyphens, 1-63 chars)",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`),
						"hostname must be RFC 1123 DNS label (lowercase alphanumeric and hyphens, 1-63 chars)",
					),
				},
			},
			"mac": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Primary MAC address (format: 00:11:22:33:44:55)",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^([0-9A-Fa-f]{2}:){5}[0-9A-Fa-f]{2}$`),
						"mac must be six groups of two hexadecimal digits separated by colons (e.g., 00:11:22:33:44:55)",
					),
				},
			},
			"category": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Category UUID reference",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`),
						"must be valid UUID (RFC 4122)",
					),
				},
			},
			"management_network": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Management network UUID reference (may be reset by BCM, required for device creation)",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`),
						"must be valid UUID (RFC 4122)",
					),
				},
			},
			"partition": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Partition UUID reference (uses category default if not specified)",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`),
						"must be valid UUID (RFC 4122)",
					),
				},
			},
			"notes": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Device notes/description",
			},
			"kernel_parameters": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Kernel boot parameters",
			},
			"boot_loader": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Boot loader type (e.g., SYSLINUX, GRUB) - defaults to category value",
			},
			"boot_loader_protocol": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Boot loader protocol (e.g., HTTP, TFTP) - defaults to category value",
			},
			"force": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Force operation (override BCM validation warnings)",
			},
			"creation_time": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Device creation timestamp (Unix epoch)",
			},
			"base_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Entity base type (always 'Device')",
			},
			"child_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Device type (HeadNode, ComputeNode, PhysicalNode, etc.)",
			},
		},
	}
}

// Configure adds the provider configured client to the resource
func (r *CMDeviceDeviceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*BCMClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *BCMClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates the device resource
func (r *CMDeviceDeviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CMDeviceDeviceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate UUID for new device (BCM requires UUID before creation)
	newUUID := uuid.New().String()

	tflog.Debug(ctx, "Creating BCM device", map[string]interface{}{
		"hostname": plan.Hostname.ValueString(),
		"mac":      plan.MAC.ValueString(),
		"uuid":     newUUID,
	})

	// If partition not specified, query category to get default partition
	partitionUUID := ""
	if !plan.Partition.IsNull() && !plan.Partition.IsUnknown() {
		partitionUUID = plan.Partition.ValueString()
	} else {
		// Query category to get its default partition
		categoryBody, err := r.client.CallJSONRPC(ctx, "cmdevice", "getCategory", plan.Category.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Querying Category",
				fmt.Sprintf("Could not query category '%s' to get default partition: %s", plan.Category.ValueString(), err.Error()),
			)
			return
		}

		var categoryData map[string]interface{}
		if err := json.Unmarshal(categoryBody, &categoryData); err != nil {
			resp.Diagnostics.AddError(
				"Error Parsing Category",
				fmt.Sprintf("Could not parse category data: %s", err.Error()),
			)
			return
		}

		if partition, ok := categoryData["partition"].(string); ok && partition != "" {
			partitionUUID = partition
			tflog.Debug(ctx, "Using category's default partition", map[string]interface{}{
				"partition": partitionUUID,
			})
		} else {
			resp.Diagnostics.AddError(
				"Missing Partition",
				"Category does not have a default partition. Please specify partition explicitly.",
			)
			return
		}
	}

	// Build device entity for BCM API (with generated UUID and resolved partition)
	deviceEntity := r.buildDeviceAPIEntity(plan, newUUID, partitionUUID)

	// Get force parameter value
	forceValue := false
	if !plan.Force.IsNull() {
		forceValue = plan.Force.ValueBool()
	}

	// Call BCM API to create device
	body, err := r.client.CallJSONRPC(ctx, "cmdevice", "addDevice", deviceEntity, forceValue)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Device",
			fmt.Sprintf("Could not create device '%s': %s", plan.Hostname.ValueString(), err.Error()),
		)
		return
	}

	// Parse validation response
	var validationResp struct {
		Success    bool                     `json:"success"`
		Validation []map[string]interface{} `json:"validation"`
	}

	if err := json.Unmarshal(body, &validationResp); err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Create Response",
			fmt.Sprintf("Could not parse device creation response: %s", err.Error()),
		)
		return
	}

	// Check validation response
	if !validationResp.Success {
		// Collect validation errors
		var errorMsgs []string
		for _, v := range validationResp.Validation {
			if field, ok := v["field"].(string); ok {
				if msg, ok := v["message"].(string); ok {
					errorMsgs = append(errorMsgs, fmt.Sprintf("%s: %s", field, msg))
				}
			}
		}
		resp.Diagnostics.AddError(
			"Error Creating Device",
			fmt.Sprintf("Failed to create device '%s': validation errors: %v", plan.Hostname.ValueString(), errorMsgs),
		)
		return
	}

	tflog.Debug(ctx, "Device created successfully", map[string]interface{}{
		"uuid":     newUUID,
		"hostname": plan.Hostname.ValueString(),
	})

	// Read back the created device to populate computed fields
	// Wait a moment for BCM to process the device creation
	time.Sleep(2 * time.Second)

	readBody, err := r.client.CallJSONRPC(ctx, "cmdevice", "getDevice", newUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Created Device",
			fmt.Sprintf("Device created but could not read back: %s", err.Error()),
		)
		return
	}

	// Parse device data
	var deviceData map[string]interface{}
	if err := json.Unmarshal(readBody, &deviceData); err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Device Data",
			fmt.Sprintf("Could not parse device data: %s", err.Error()),
		)
		return
	}

	// Map device data to state
	state := r.parseDeviceFromAPI(deviceData)

	// Preserve write-only fields that BCM doesn't return
	state.Force = plan.Force

	// Handle partition - BCM may not return it, use what was resolved during create
	if state.Partition.IsNull() || state.Partition.ValueString() == "" {
		state.Partition = types.StringValue(partitionUUID)
	}

	// BCM returns nil UUID for management_network - preserve the configured value
	if !plan.ManagementNetwork.IsNull() && !plan.ManagementNetwork.IsUnknown() {
		if state.ManagementNetwork.IsNull() || state.ManagementNetwork.ValueString() == "00000000-0000-0000-0000-000000000000" {
			state.ManagementNetwork = plan.ManagementNetwork
		}
	}

	// Set state - use what BCM returns for all other fields
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Read reads the device resource
func (r *CMDeviceDeviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CMDeviceDeviceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use ID if UUID is not set (happens during import)
	deviceID := state.UUID.ValueString()
	if deviceID == "" {
		deviceID = state.ID.ValueString()
	}

	tflog.Debug(ctx, "Reading BCM device", map[string]interface{}{
		"id": deviceID,
	})

	// Call BCM API to get device (efficient direct lookup)
	body, err := r.client.CallJSONRPC(ctx, "cmdevice", "getDevice", deviceID)
	if err != nil || len(body) == 0 {
		tflog.Warn(ctx, "Device not found in BCM, removing from state", map[string]interface{}{
			"uuid": state.UUID.ValueString(),
		})
		resp.State.RemoveResource(ctx)
		return
	}

	// Parse device data
	var deviceData map[string]interface{}
	if err := json.Unmarshal(body, &deviceData); err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Device Data",
			fmt.Sprintf("Could not parse device data: %s", err.Error()),
		)
		return
	}

	// Map device data to state
	newState := r.parseDeviceFromAPI(deviceData)

	// Preserve write-only fields that BCM doesn't return
	newState.Force = state.Force

	// BCM returns nil UUID for management_network - preserve the configured value
	// when BCM returns "00000000-0000-0000-0000-000000000000"
	if !state.ManagementNetwork.IsNull() && !state.ManagementNetwork.IsUnknown() {
		if newState.ManagementNetwork.IsNull() || newState.ManagementNetwork.ValueString() == "00000000-0000-0000-0000-000000000000" {
			newState.ManagementNetwork = state.ManagementNetwork
		}
	}

	// BCM returns "CATEGORY" for boot_loader/boot_loader_protocol when inheriting from category
	// Preserve null/empty plan values to avoid drift
	if state.BootLoader.IsNull() && newState.BootLoader.ValueString() == "CATEGORY" {
		newState.BootLoader = types.StringNull()
	}
	if state.BootLoaderProtocol.IsNull() && newState.BootLoaderProtocol.ValueString() == "CATEGORY" {
		newState.BootLoaderProtocol = types.StringNull()
	}

	// BCM may add partition if not explicitly set - preserve null from plan to avoid drift
	if state.Partition.IsNull() && !newState.Partition.IsNull() {
		newState.Partition = types.StringNull()
	}

	// Set state - use what BCM returns (with preserved fields)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

// Update updates the device resource
func (r *CMDeviceDeviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CMDeviceDeviceResourceModel
	var state CMDeviceDeviceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating BCM device", map[string]interface{}{
		"uuid":     state.UUID.ValueString(),
		"hostname": plan.Hostname.ValueString(),
	})

	// Resolve partition UUID (either from plan or from category's default)
	partitionUUID := ""
	if !plan.Partition.IsNull() && !plan.Partition.IsUnknown() {
		partitionUUID = plan.Partition.ValueString()
	} else {
		// Query category to get default partition
		categoryBody, err := r.client.CallJSONRPC(ctx, "cmdevice", "getCategory", plan.Category.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Querying Category",
				fmt.Sprintf("Could not query category '%s' to get partition: %s", plan.Category.ValueString(), err.Error()),
			)
			return
		}

		var categoryData map[string]interface{}
		if err := json.Unmarshal(categoryBody, &categoryData); err != nil {
			resp.Diagnostics.AddError(
				"Error Parsing Category Data",
				fmt.Sprintf("Could not parse category data: %s", err.Error()),
			)
			return
		}

		if partition, ok := categoryData["partition"].(string); ok && partition != "" {
			partitionUUID = partition
			tflog.Debug(ctx, "Using category's default partition", map[string]interface{}{
				"partition": partitionUUID,
			})
		}
	}

	// Build device entity for BCM API (include UUID for update)
	deviceEntity := r.buildDeviceAPIEntity(plan, state.UUID.ValueString(), partitionUUID)

	// Get force parameter value
	forceValue := false
	if !plan.Force.IsNull() {
		forceValue = plan.Force.ValueBool()
	}

	// Call BCM API to update device
	_, err := r.client.CallJSONRPC(ctx, "cmdevice", "updateDevice", deviceEntity, forceValue)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Device",
			fmt.Sprintf("Could not update device '%s': %s", plan.Hostname.ValueString(), err.Error()),
		)
		return
	}

	tflog.Debug(ctx, "Device updated successfully", map[string]interface{}{
		"uuid":     state.UUID.ValueString(),
		"hostname": plan.Hostname.ValueString(),
	})

	// Read back the updated device
	readBody, err := r.client.CallJSONRPC(ctx, "cmdevice", "getDevice", state.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Updated Device",
			fmt.Sprintf("Device updated but could not read back: %s", err.Error()),
		)
		return
	}

	// Parse device data
	var deviceData map[string]interface{}
	if err := json.Unmarshal(readBody, &deviceData); err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Device Data",
			fmt.Sprintf("Could not parse device data: %s", err.Error()),
		)
		return
	}

	// Map device data to state
	newState := r.parseDeviceFromAPI(deviceData)

	// Preserve plan values for fields not persisted by BCM or modified during updates
	newState.Force = plan.Force
	newState.ManagementNetwork = plan.ManagementNetwork // BCM sets to nil UUID, preserve plan value
	newState.Partition = plan.Partition                 // BCM doesn't return partition, preserve plan value

	// Preserve null values for optional fields that BCM populates from category
	if plan.BootLoader.IsNull() {
		newState.BootLoader = types.StringNull() // Keep null if not explicitly set
	}
	if plan.BootLoaderProtocol.IsNull() {
		newState.BootLoaderProtocol = types.StringNull() // Keep null if not explicitly set
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

// Delete deletes the device resource
func (r *CMDeviceDeviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CMDeviceDeviceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting BCM device", map[string]interface{}{
		"uuid":     state.UUID.ValueString(),
		"hostname": state.Hostname.ValueString(),
	})

	// Get force parameter value
	forceValue := false
	if !state.Force.IsNull() {
		forceValue = state.Force.ValueBool()
	}

	// Call BCM API to delete device
	_, err := r.client.CallJSONRPC(ctx, "cmdevice", "removeDevice", state.UUID.ValueString(), forceValue)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Device",
			fmt.Sprintf("Could not delete device '%s' (UUID: %s): %s",
				state.Hostname.ValueString(), state.UUID.ValueString(), err.Error()),
		)
		return
	}

	tflog.Debug(ctx, "Device deleted successfully", map[string]interface{}{
		"uuid":     state.UUID.ValueString(),
		"hostname": state.Hostname.ValueString(),
	})
}

// ImportState imports an existing device by UUID
func (r *CMDeviceDeviceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	tflog.Debug(ctx, "Importing BCM device", map[string]interface{}{
		"id": req.ID,
	})
}

// buildDeviceAPIEntity constructs BCM API entity from Terraform model
func (r *CMDeviceDeviceResource) buildDeviceAPIEntity(plan CMDeviceDeviceResourceModel, uuid string, partitionUUID string) map[string]interface{} {
	// Create a basic network interface for the device
	interfaceUUID := "00000000-0000-0000-0000-000000000001" // Generate a temporary UUID for the interface

	// Use management_network from plan if specified, otherwise nil UUID
	networkUUID := "00000000-0000-0000-0000-000000000000" // Nil UUID by default
	if !plan.ManagementNetwork.IsNull() && !plan.ManagementNetwork.IsUnknown() {
		networkUUID = plan.ManagementNetwork.ValueString()
	}

	networkInterface := map[string]interface{}{
		"baseType":             "NetworkInterface",
		"childType":            "NetworkPhysicalInterface",
		"mac":                  plan.MAC.ValueString(),
		"network":              networkUUID,
		"name":                 "eth0", // Default interface name
		"dhcp":                 true,   // Use DHCP by default
		"bootable":             true,
		"startIf":              "ALWAYS",
		"modified":             true,
		"to_be_removed":        false,
		"revision":             "",
		"uuid":                 interfaceUUID,
		"ipv6Ip":               "::0",
		"ipv6Dhcp":             false,
		"bringupduringinstall": "NO",
		"cardtype":             "Ethernet",
	}

	entity := map[string]interface{}{
		"baseType":              "Device",
		"childType":             "PhysicalNode", // Default to PhysicalNode, BCM may change based on roles
		"hostname":              plan.Hostname.ValueString(),
		"mac":                   plan.MAC.ValueString(),
		"category":              plan.Category.ValueString(),
		"managementNetwork":     "00000000-0000-0000-0000-000000000000", // Nil UUID as BCM will assign
		"partition":             partitionUUID,                           // Required field, resolved from category if not specified
		"modified":              true,
		"to_be_removed":         false,
		"revision":              "",
		"uuid":                  uuid, // Always include UUID (generated for create, from state for update)
		"provisioningInterface": interfaceUUID,
		"interfaces":            []interface{}{networkInterface}, // Include the interface in the device
	}

	// Add optional fields if present
	if !plan.Notes.IsNull() && !plan.Notes.IsUnknown() {
		entity["notes"] = plan.Notes.ValueString()
	}

	if !plan.KernelParameters.IsNull() && !plan.KernelParameters.IsUnknown() {
		entity["kernelParameters"] = plan.KernelParameters.ValueString()
	}

	if !plan.BootLoader.IsNull() && !plan.BootLoader.IsUnknown() {
		entity["bootLoader"] = plan.BootLoader.ValueString()
	}

	if !plan.BootLoaderProtocol.IsNull() && !plan.BootLoaderProtocol.IsUnknown() {
		entity["bootLoaderProtocol"] = plan.BootLoaderProtocol.ValueString()
	}

	return entity
}

// parseDeviceFromAPI parses BCM API response into Terraform model
func (r *CMDeviceDeviceResource) parseDeviceFromAPI(data map[string]interface{}) CMDeviceDeviceResourceModel {
	model := CMDeviceDeviceResourceModel{}

	// Required fields
	if uuid, ok := data["uuid"].(string); ok {
		model.UUID = types.StringValue(uuid)
		model.ID = types.StringValue(uuid) // ID same as UUID
	}

	if hostname, ok := data["hostname"].(string); ok {
		model.Hostname = types.StringValue(hostname)
	}

	if mac, ok := data["mac"].(string); ok {
		model.MAC = types.StringValue(mac)
	}

	if category, ok := data["category"].(string); ok {
		model.Category = types.StringValue(category)
	}

	// BCM often returns nil UUID for managementNetwork - handle gracefully
	if managementNetwork, ok := data["managementNetwork"].(string); ok && managementNetwork != "" && managementNetwork != "00000000-0000-0000-0000-000000000000" {
		model.ManagementNetwork = types.StringValue(managementNetwork)
	} else {
		model.ManagementNetwork = types.StringNull()
	}

	// Partition (optional, may not be returned)
	if partition, ok := data["partition"].(string); ok && partition != "" {
		model.Partition = types.StringValue(partition)
	} else {
		model.Partition = types.StringNull()
	}

	// Optional fields with null safety
	if notes, ok := data["notes"].(string); ok && notes != "" {
		model.Notes = types.StringValue(notes)
	} else {
		model.Notes = types.StringNull()
	}

	if kernelParams, ok := data["kernelParameters"].(string); ok && kernelParams != "" {
		model.KernelParameters = types.StringValue(kernelParams)
	} else {
		model.KernelParameters = types.StringNull()
	}

	if bootLoader, ok := data["bootLoader"].(string); ok && bootLoader != "" {
		model.BootLoader = types.StringValue(bootLoader)
	} else {
		model.BootLoader = types.StringNull()
	}

	if bootLoaderProtocol, ok := data["bootLoaderProtocol"].(string); ok && bootLoaderProtocol != "" {
		model.BootLoaderProtocol = types.StringValue(bootLoaderProtocol)
	} else {
		model.BootLoaderProtocol = types.StringNull()
	}

	// Computed fields
	if creationTime, ok := data["creationTime"].(float64); ok {
		model.CreationTime = types.Int64Value(int64(creationTime))
	}

	if baseType, ok := data["baseType"].(string); ok {
		model.BaseType = types.StringValue(baseType)
	}

	if childType, ok := data["childType"].(string); ok && childType != "" {
		model.ChildType = types.StringValue(childType)
	} else {
		model.ChildType = types.StringNull()
	}

	// Force is not persisted by BCM, will be preserved from plan/state

	return model
}
