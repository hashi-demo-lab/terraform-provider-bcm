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

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &CMDeviceDeviceResource{}
	_ resource.ResourceWithImportState = &CMDeviceDeviceResource{}
)

// CMDeviceDeviceResource defines the resource implementation.
type CMDeviceDeviceResource struct {
	client *BCMClient
}

// CMDeviceDeviceResourceModel describes the resource data model.
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

	// Power control configuration
	PowerControl types.String `tfsdk:"power_control"` // Optional, e.g., "none", "ipmi", "ipdu"

	// Network gateway configuration
	DefaultGateway       types.String `tfsdk:"default_gateway"`        // Optional, IP address
	DefaultGatewayMetric types.Int64  `tfsdk:"default_gateway_metric"` // Optional+Computed, gateway metric/priority

	// Hardware identifiers
	SerialNumber types.String `tfsdk:"serial_number"` // Optional+Computed, hardware serial number
	PartNumber   types.String `tfsdk:"part_number"`   // Optional+Computed, hardware part number

	// Computed fields
	CreationTime types.Int64  `tfsdk:"creation_time"` // Computed
	BaseType     types.String `tfsdk:"base_type"`     // Computed, always "Device"
	ChildType    types.String `tfsdk:"child_type"`    // Computed, BCM-determined

	// Interfaces block - NEW for multi-interface support
	// When specified, takes precedence over legacy mac/management_network
	Interfaces []DeviceInterfaceModel `tfsdk:"interfaces"`
}

// NewCMDeviceDeviceResource creates a new resource instance.
func NewCMDeviceDeviceResource() resource.Resource {
	return &CMDeviceDeviceResource{}
}

// Metadata returns the resource type name.
func (r *CMDeviceDeviceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cmdevice_device"
}

// Schema defines the resource schema.
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
			"power_control": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Power control method (e.g., 'none', 'ipmi', 'ipdu', 'redfish')",
			},
			"default_gateway": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Default gateway IP address for the device",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`),
						"must be a valid IPv4 address",
					),
				},
			},
			"default_gateway_metric": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Default gateway metric/priority (lower is preferred)",
			},
			"serial_number": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Hardware serial number",
			},
			"part_number": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Hardware part number",
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
		Blocks: map[string]schema.Block{
			"interfaces": schema.ListNestedBlock{
				MarkdownDescription: "Network interface configurations for the device. " +
					"When specified, provides full control over interface setup. " +
					"Each interface can be physical, bond, or BMC type.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Interface name (e.g., 'eth0', 'bond0', 'ipmi'). Must be unique within the device.",
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 63),
								stringvalidator.RegexMatches(
									regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`),
									"must start with letter, contain only alphanumeric, underscore, or hyphen",
								),
							},
						},
						"type": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Interface type: 'physical', 'bond', or 'bmc'.",
							Validators: []validator.String{
								stringvalidator.OneOf("physical", "bond", "bmc"),
							},
						},
						"network": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Network UUID reference for interface assignment.",
							Validators: []validator.String{
								stringvalidator.RegexMatches(
									regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`),
									"must be valid UUID (RFC 4122)",
								),
							},
						},
						"mac": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "MAC address (format: 00:11:22:33:44:55). Required for physical interfaces on create.",
							Validators: []validator.String{
								stringvalidator.RegexMatches(
									regexp.MustCompile(`^([0-9A-Fa-f]{2}:){5}[0-9A-Fa-f]{2}$`),
									"must be six groups of two hexadecimal digits separated by colons",
								),
							},
						},
						"ip": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Static IPv4 address.",
							Validators: []validator.String{
								stringvalidator.RegexMatches(
									regexp.MustCompile(`^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`),
									"must be a valid IPv4 address",
								),
							},
						},
						"ipv6_ip": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Static IPv6 address.",
						},
						"dhcp": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							MarkdownDescription: "Enable DHCP for IP assignment. Default: true.",
						},
						"bootable": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							MarkdownDescription: "Enable PXE boot capability. Default: false. First bootable interface becomes provisioning interface.",
						},
						"start_if": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							MarkdownDescription: "Interface startup condition: 'ALWAYS', 'NEVER', 'HOTPLUG'. Default: 'ALWAYS'.",
							Validators: []validator.String{
								stringvalidator.OneOf("ALWAYS", "NEVER", "HOTPLUG"),
							},
						},
						"members": schema.ListAttribute{
							Optional:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "Member interface names for bond type. Required when type is 'bond'.",
						},
						"bond_mode": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Bond mode (e.g., '802.3ad', 'active-backup', 'balance-rr'). Only applicable when type is 'bond'.",
							Validators: []validator.String{
								stringvalidator.OneOf(
									"802.3ad",
									"active-backup",
									"balance-rr",
									"balance-xor",
									"broadcast",
									"balance-tlb",
									"balance-alb",
								),
							},
						},
						"uuid": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "BCM-assigned interface UUID.",
						},
						"base_type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Entity base type (always 'NetworkInterface').",
						},
						"child_type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Interface type (NetworkPhysicalInterface, NetworkBondInterface, NetworkBMCInterface).",
						},
						"cardtype": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Hardware card type (Ethernet, InfiniBand, BMC).",
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
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

// Create creates the device resource.
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
	usesProxy := false // Track if partition comes from softwareImageProxy

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

		// Try direct partition field first
		if partition, ok := categoryData["partition"].(string); ok && partition != "" {
			partitionUUID = partition
			tflog.Debug(ctx, "Using category's direct partition", map[string]interface{}{
				"partition": partitionUUID,
			})
		} else if proxyData, ok := categoryData["softwareImageProxy"].(map[string]interface{}); ok && proxyData != nil {
			// Check if category uses softwareImageProxy instead
			if parentImage, ok := proxyData["parentSoftwareImage"].(string); ok && parentImage != "" {
				usesProxy = true // Mark that this category uses softwareImageProxy
				tflog.Debug(ctx, "Category uses softwareImageProxy - will use cluster's base partition", map[string]interface{}{
					"parent_software_image": parentImage,
				})

				// When softwareImageProxy is used, devices must reference the cluster's default partition
				// Query for the base partition
				partitionsBody, err := r.client.CallJSONRPC(ctx, "CMPart", "getPartitions")
				if err != nil {
					resp.Diagnostics.AddError(
						"Error Querying Partitions",
						fmt.Sprintf("Could not query partitions: %s", err.Error()),
					)
					return
				}

				var partitions []map[string]interface{}
				if err := json.Unmarshal(partitionsBody, &partitions); err != nil {
					resp.Diagnostics.AddError(
						"Error Parsing Partitions",
						fmt.Sprintf("Could not parse partitions response: %s", err.Error()),
					)
					return
				}

				// Find the base partition (or first available partition)
				basePartitionFound := false
				for _, part := range partitions {
					if name, ok := part["name"].(string); ok && name == "base" {
						if uuid, ok := part["uuid"].(string); ok && uuid != "" {
							partitionUUID = uuid
							basePartitionFound = true
							tflog.Debug(ctx, "Found base partition for softwareImageProxy", map[string]interface{}{
								"partition_uuid": partitionUUID,
							})
							break
						}
					}
				}

				if !basePartitionFound {
					resp.Diagnostics.AddError(
						"Missing Base Partition",
						"Category uses softwareImageProxy but no 'base' partition found in cluster",
					)
					return
				}

				// Add delay for BCM to process the proxy configuration (eventual consistency)
				// BCM needs time to propagate the software image proxy configuration
				tflog.Debug(ctx, "Waiting for category proxy to stabilize", nil)
				time.Sleep(5 * time.Second)
			} else {
				resp.Diagnostics.AddError(
					"Missing Partition",
					"Category has softwareImageProxy but no parentSoftwareImage. Please specify partition explicitly.",
				)
				return
			}
		} else {
			resp.Diagnostics.AddError(
				"Missing Partition",
				"Category does not have a default partition or software image proxy. Please specify partition explicitly.",
			)
			return
		}
	}

	// Only wait for partition commit if NOT using softwareImageProxy
	// When softwareImageProxy is used, the partition is derived from category's proxy configuration
	// and we only need to ensure the software image is committed (not the partition)
	if !usesProxy {
		tflog.Debug(ctx, "Waiting for partition to be committed", map[string]interface{}{
			"partition_uuid": partitionUUID,
			"uses_proxy":     usesProxy,
		})
		if err := r.waitForPartitionCommit(ctx, r.client, partitionUUID); err != nil {
			resp.Diagnostics.AddError(
				"Partition Not Ready",
				fmt.Sprintf("Partition %s is not committed/available after waiting: %s\n\n"+
					"This typically happens when a newly created software image hasn't finished "+
					"its async commit process. The provider will automatically retry, but if this "+
					"persists, the software image may need to be created separately and given time "+
					"to commit before creating devices.", partitionUUID, err.Error()),
			)
			return
		}
	} else {
		// When using softwareImageProxy, no need to wait for partition
		// The partition is the cluster's existing base partition, not a newly created one
		tflog.Debug(ctx, "Category uses softwareImageProxy with base partition", map[string]interface{}{
			"partition_uuid": partitionUUID,
			"uses_proxy":     usesProxy,
		})
	}

	// Build device entity for BCM API (with generated UUID and resolved partition)
	// Partition field is always included as BCM requires it
	deviceEntity := r.buildDeviceAPIEntity(plan, newUUID, partitionUUID)

	// Pre-flight validation: Call validateDevice before CREATE
	validationErrors, err := r.client.ValidateEntity(ctx, "CMDevice", "validateDevice", deviceEntity, true)
	if err != nil {
		resp.Diagnostics.AddError(
			"Validation API Error",
			fmt.Sprintf("Could not validate device '%s': %s", plan.Hostname.ValueString(), err.Error()),
		)
		return
	}

	// Process validation results
	hasErrors := false
	for _, valErr := range validationErrors {
		if valErr.IsError() {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Validation Error: %s", valErr.Field),
				valErr.Message,
			)
			hasErrors = true
		} else if valErr.IsWarning() {
			resp.Diagnostics.AddWarning(
				fmt.Sprintf("Validation Warning: %s", valErr.Field),
				valErr.Message,
			)
		}
	}

	// Halt if validation errors found
	if hasErrors {
		return
	}

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

	// BCM returns default values for Optional+Computed fields when not explicitly set
	// Preserve null from plan to avoid drift
	if plan.PowerControl.IsNull() && !state.PowerControl.IsNull() {
		state.PowerControl = types.StringNull()
	}

	if plan.DefaultGateway.IsNull() && !state.DefaultGateway.IsNull() {
		state.DefaultGateway = types.StringNull()
	}

	if plan.DefaultGatewayMetric.IsNull() && !state.DefaultGatewayMetric.IsNull() {
		state.DefaultGatewayMetric = types.Int64Null()
	}

	if plan.SerialNumber.IsNull() && !state.SerialNumber.IsNull() {
		state.SerialNumber = types.StringNull()
	}

	if plan.PartNumber.IsNull() && !state.PartNumber.IsNull() {
		state.PartNumber = types.StringNull()
	}

	// Handle interfaces in state based on mode
	if len(plan.Interfaces) > 0 {
		// Interfaces mode: Normalize interface order to match plan order (prevents spurious diffs)
		state.Interfaces = normalizeInterfaceOrder(state.Interfaces, plan.Interfaces)
	} else {
		// Legacy mode: Don't populate interfaces in state (user didn't define interfaces block)
		// BCM creates interfaces automatically, but we don't expose them in legacy mode
		state.Interfaces = nil
	}

	// Set state - use what BCM returns for all other fields
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// waitForPartitionCommit polls BCM to verify a software image partition is committed and available
// This is necessary because BCM software image creation is asynchronous - the API returns success
// before the image is fully committed and ready to be used as a device partition.
func (r *CMDeviceDeviceResource) waitForPartitionCommit(ctx context.Context, client *BCMClient, partitionUUID string) error {
	// Increased from 10 to 20 retries to handle BCM's async partition commit timing
	// BCM can take 60-120 seconds for partition commits in some environments
	maxRetries := 20
	baseDelay := 2 * time.Second
	maxDelay := 10 * time.Second

	totalWait := 0 * time.Second

	for i := 0; i < maxRetries; i++ {
		tflog.Debug(ctx, "Checking partition commit status", map[string]interface{}{
			"partition_uuid":  partitionUUID,
			"attempt":         i + 1,
			"max_retries":     maxRetries,
			"total_wait_secs": totalWait.Seconds(),
		})

		// Try to query the partition - if it's committed, this will succeed
		_, err := client.CallJSONRPC(ctx, "CMPart", "getSoftwareImage", partitionUUID)
		if err == nil {
			tflog.Info(ctx, "Partition is committed and available", map[string]interface{}{
				"partition_uuid":  partitionUUID,
				"attempts":        i + 1,
				"total_wait_secs": totalWait.Seconds(),
			})
			return nil // Partition is accessible
		}

		// If not the last retry, wait with exponential backoff (capped at maxDelay)
		if i < maxRetries-1 {
			// Exponential backoff: 2s, 4s, 6s, 8s, 10s, 10s, ...
			delay := baseDelay * time.Duration(i+1)
			if delay > maxDelay {
				delay = maxDelay
			}
			totalWait += delay

			tflog.Debug(ctx, "Partition not ready, waiting before retry", map[string]interface{}{
				"partition_uuid":  partitionUUID,
				"delay_seconds":   delay.Seconds(),
				"attempt":         i + 1,
				"total_wait_secs": totalWait.Seconds(),
				"error":           err.Error(),
			})
			time.Sleep(delay)
		}
	}

	return fmt.Errorf("partition not committed after %d retries (waited up to %.0f seconds)",
		maxRetries, totalWait.Seconds())
}

// Read reads the device resource.
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

	// CRITICAL FIX FOR IMPORT: Handle optional+computed fields differently for import vs normal Read
	// During import (state is empty), we should NOT preserve null values - instead use what BCM returns
	// During normal Read (state has values), preserve null if user didn't explicitly set the field

	isImport := state.ManagementNetwork.IsNull() && state.BootLoader.IsNull() && state.BootLoaderProtocol.IsNull()

	if !isImport {
		// Normal Read path: Preserve null values from state to avoid drift

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

		// BCM returns default values for Optional+Computed fields when not explicitly set
		// Preserve null from plan to avoid drift
		if state.PowerControl.IsNull() && !newState.PowerControl.IsNull() {
			newState.PowerControl = types.StringNull()
		}

		if state.DefaultGateway.IsNull() && !newState.DefaultGateway.IsNull() {
			newState.DefaultGateway = types.StringNull()
		}

		if state.DefaultGatewayMetric.IsNull() && !newState.DefaultGatewayMetric.IsNull() {
			newState.DefaultGatewayMetric = types.Int64Null()
		}

		if state.SerialNumber.IsNull() && !newState.SerialNumber.IsNull() {
			newState.SerialNumber = types.StringNull()
		}

		if state.PartNumber.IsNull() && !newState.PartNumber.IsNull() {
			newState.PartNumber = types.StringNull()
		}
	} else {
		// Import path: Use all values from BCM, don't set to null
		// BCM returns "CATEGORY" for fields inheriting from category - this is valid during import
		tflog.Debug(ctx, "Import detected - using all BCM values", map[string]interface{}{
			"management_network":     newState.ManagementNetwork.ValueString(),
			"boot_loader":            newState.BootLoader.ValueString(),
			"boot_loader_protocol":   newState.BootLoaderProtocol.ValueString(),
			"partition":              newState.Partition.ValueString(),
			"default_gateway_metric": newState.DefaultGatewayMetric.ValueInt64(),
		})
	}

	// Handle interfaces based on mode
	// During import (isImport=true), always populate interfaces from BCM
	// During normal read, use state to determine if user used interfaces block
	if isImport {
		// Import: Keep interfaces from BCM API response (already parsed in newState)
		tflog.Debug(ctx, "Import path - keeping interfaces from BCM API", map[string]interface{}{
			"interface_count": len(newState.Interfaces),
		})
	} else if len(state.Interfaces) > 0 {
		// Interfaces mode: normalize order and merge bond-specific fields from state
		// BCM doesn't return members/bond_mode fields, so we need to preserve them from state
		if len(newState.Interfaces) > 0 {
			newState.Interfaces = normalizeInterfaceOrder(newState.Interfaces, state.Interfaces)
		}
	} else {
		// Legacy mode: Don't populate interfaces in state (user didn't define interfaces block)
		// This maintains backward compatibility with existing configs
		newState.Interfaces = nil
	}

	// Set state - use what BCM returns (with preserved fields)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

// Update updates the device resource.
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

		// Try direct partition field first
		if partition, ok := categoryData["partition"].(string); ok && partition != "" {
			partitionUUID = partition
			tflog.Debug(ctx, "Using category's direct partition", map[string]interface{}{
				"partition": partitionUUID,
			})
		} else if proxyData, ok := categoryData["softwareImageProxy"].(map[string]interface{}); ok && proxyData != nil {
			// Check if category uses softwareImageProxy instead
			if parentImage, ok := proxyData["parentSoftwareImage"].(string); ok && parentImage != "" {
				tflog.Debug(ctx, "Category uses softwareImageProxy - will use cluster's base partition", map[string]interface{}{
					"parent_software_image": parentImage,
				})

				// Query for the base partition
				partitionsBody, err := r.client.CallJSONRPC(ctx, "CMPart", "getPartitions")
				if err != nil {
					resp.Diagnostics.AddError(
						"Error Querying Partitions",
						fmt.Sprintf("Could not query partitions: %s", err.Error()),
					)
					return
				}

				var partitions []map[string]interface{}
				if err := json.Unmarshal(partitionsBody, &partitions); err != nil {
					resp.Diagnostics.AddError(
						"Error Parsing Partitions",
						fmt.Sprintf("Could not parse partitions response: %s", err.Error()),
					)
					return
				}

				// Find the base partition
				basePartitionFound := false
				for _, part := range partitions {
					if name, ok := part["name"].(string); ok && name == "base" {
						if uuid, ok := part["uuid"].(string); ok && uuid != "" {
							partitionUUID = uuid
							basePartitionFound = true
							tflog.Debug(ctx, "Found base partition for softwareImageProxy", map[string]interface{}{
								"partition_uuid": partitionUUID,
							})
							break
						}
					}
				}

				if !basePartitionFound {
					resp.Diagnostics.AddError(
						"Missing Base Partition",
						"Category uses softwareImageProxy but no 'base' partition found in cluster",
					)
					return
				}
			}
		}
	}

	// Build device entity for BCM API (include UUID for update)
	// Partition field is always included as BCM requires it
	// Pass existing interfaces from state to preserve UUIDs
	deviceEntity := r.buildDeviceAPIEntityWithExisting(plan, state.UUID.ValueString(), partitionUUID, state.Interfaces)

	// Pre-flight validation: Call validateDevice before UPDATE
	validationErrors, err := r.client.ValidateEntity(ctx, "CMDevice", "validateDevice", deviceEntity, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Validation API Error",
			fmt.Sprintf("Could not validate device '%s': %s", plan.Hostname.ValueString(), err.Error()),
		)
		return
	}

	// Process validation results
	hasErrors := false
	for _, valErr := range validationErrors {
		if valErr.IsError() {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Validation Error: %s", valErr.Field),
				valErr.Message,
			)
			hasErrors = true
		} else if valErr.IsWarning() {
			resp.Diagnostics.AddWarning(
				fmt.Sprintf("Validation Warning: %s", valErr.Field),
				valErr.Message,
			)
		}
	}

	// Halt if validation errors found
	if hasErrors {
		return
	}

	// Get force parameter value
	forceValue := false
	if !plan.Force.IsNull() {
		forceValue = plan.Force.ValueBool()
	}

	// Call BCM API to update device
	_, err = r.client.CallJSONRPC(ctx, "cmdevice", "updateDevice", deviceEntity, forceValue)
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

	// BCM sets management_network to nil UUID - preserve plan value if set, otherwise use state or null
	if !plan.ManagementNetwork.IsNull() && !plan.ManagementNetwork.IsUnknown() {
		newState.ManagementNetwork = plan.ManagementNetwork
	} else if !state.ManagementNetwork.IsNull() && !state.ManagementNetwork.IsUnknown() {
		newState.ManagementNetwork = state.ManagementNetwork
	} else {
		newState.ManagementNetwork = types.StringNull()
	}

	// Set partition to the resolved value (either from plan or resolved from category)
	if partitionUUID != "" {
		newState.Partition = types.StringValue(partitionUUID)
	} else if !plan.Partition.IsNull() {
		newState.Partition = plan.Partition
	} else {
		newState.Partition = types.StringNull()
	}

	// Preserve null values for optional fields that BCM populates from category
	if plan.BootLoader.IsNull() {
		newState.BootLoader = types.StringNull() // Keep null if not explicitly set
	}
	if plan.BootLoaderProtocol.IsNull() {
		newState.BootLoaderProtocol = types.StringNull() // Keep null if not explicitly set
	}

	// BCM returns default values for Optional+Computed fields when not explicitly set
	// Preserve null from plan to avoid drift
	if plan.PowerControl.IsNull() && !newState.PowerControl.IsNull() {
		newState.PowerControl = types.StringNull()
	}

	if plan.DefaultGateway.IsNull() && !newState.DefaultGateway.IsNull() {
		newState.DefaultGateway = types.StringNull()
	}

	if plan.DefaultGatewayMetric.IsNull() && !newState.DefaultGatewayMetric.IsNull() {
		newState.DefaultGatewayMetric = types.Int64Null()
	}

	if plan.SerialNumber.IsNull() && !newState.SerialNumber.IsNull() {
		newState.SerialNumber = types.StringNull()
	}

	if plan.PartNumber.IsNull() && !newState.PartNumber.IsNull() {
		newState.PartNumber = types.StringNull()
	}

	// Handle interfaces in state based on mode
	if len(plan.Interfaces) > 0 {
		// Interfaces mode: Normalize interface order to match plan order (prevents spurious diffs)
		newState.Interfaces = normalizeInterfaceOrder(newState.Interfaces, plan.Interfaces)
	} else {
		// Legacy mode: Don't populate interfaces in state (user didn't define interfaces block)
		// BCM creates/maintains interfaces, but we don't expose them in legacy mode
		newState.Interfaces = nil
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

// Delete deletes the device resource.
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

// ImportState imports an existing device by UUID.
func (r *CMDeviceDeviceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	tflog.Debug(ctx, "Importing BCM device", map[string]interface{}{
		"id": req.ID,
	})
}

// buildDeviceAPIEntity constructs BCM API entity from Terraform model.
func (r *CMDeviceDeviceResource) buildDeviceAPIEntity(plan CMDeviceDeviceResourceModel, uuid string, partitionUUID string) map[string]interface{} {
	return r.buildDeviceAPIEntityWithExisting(plan, uuid, partitionUUID, nil)
}

// buildDeviceAPIEntityWithExisting constructs BCM API entity, preserving interface UUIDs from existing state.
func (r *CMDeviceDeviceResource) buildDeviceAPIEntityWithExisting(plan CMDeviceDeviceResourceModel, deviceUUID string, partitionUUID string, existingInterfaces []DeviceInterfaceModel) map[string]interface{} {
	var interfaces []interface{}
	var provisioningInterfaceUUID string

	// Check if using interfaces block or legacy mode
	if len(plan.Interfaces) > 0 {
		// NEW: Build interfaces from the interfaces block
		interfaces = buildInterfacesAPIArray(plan.Interfaces, existingInterfaces)
		provisioningInterfaceUUID = getProvisioningInterfaceUUID(plan.Interfaces)

		// If no provisioning interface found from plan, get from built interfaces
		if provisioningInterfaceUUID == "" && len(interfaces) > 0 {
			if firstIface, ok := interfaces[0].(map[string]interface{}); ok {
				if ifaceUUID, ok := firstIface["uuid"].(string); ok {
					provisioningInterfaceUUID = ifaceUUID
				}
			}
		}
	} else {
		// LEGACY: Create a basic network interface from mac/management_network
		interfaceUUID := uuid.New().String()

		// Use management_network from plan if specified, otherwise nil UUID
		networkUUID := "00000000-0000-0000-0000-000000000000"
		if !plan.ManagementNetwork.IsNull() && !plan.ManagementNetwork.IsUnknown() {
			networkUUID = plan.ManagementNetwork.ValueString()
		}

		networkInterface := map[string]interface{}{
			"baseType":             "NetworkInterface",
			"childType":            "NetworkPhysicalInterface",
			"mac":                  plan.MAC.ValueString(),
			"network":              networkUUID,
			"name":                 "eth0",
			"dhcp":                 true,
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

		interfaces = []interface{}{networkInterface}
		provisioningInterfaceUUID = interfaceUUID
	}

	// Get MAC for device entity - use first interface MAC if interfaces block, else legacy mac field
	deviceMAC := ""
	if len(plan.Interfaces) > 0 && !plan.Interfaces[0].MAC.IsNull() {
		deviceMAC = plan.Interfaces[0].MAC.ValueString()
	} else if !plan.MAC.IsNull() && !plan.MAC.IsUnknown() {
		deviceMAC = plan.MAC.ValueString()
	}

	entity := map[string]interface{}{
		"baseType":              "Device",
		"childType":             "PhysicalNode",
		"hostname":              plan.Hostname.ValueString(),
		"mac":                   deviceMAC,
		"category":              plan.Category.ValueString(),
		"managementNetwork":     "00000000-0000-0000-0000-000000000000",
		"modified":              true,
		"to_be_removed":         false,
		"revision":              "",
		"uuid":                  deviceUUID,
		"provisioningInterface": provisioningInterfaceUUID,
		"interfaces":            interfaces,
	}

	// Always include partition field as BCM requires it
	// When usesProxy is true, partitionUUID contains the software image UUID from parentSoftwareImage
	// BCM will use this in combination with the category's softwareImageProxy configuration
	entity["partition"] = partitionUUID

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

	// Power control configuration
	if !plan.PowerControl.IsNull() && !plan.PowerControl.IsUnknown() {
		entity["powerControl"] = plan.PowerControl.ValueString()
	}

	// Network gateway configuration
	if !plan.DefaultGateway.IsNull() && !plan.DefaultGateway.IsUnknown() {
		entity["defaultGateway"] = plan.DefaultGateway.ValueString()
	}

	if !plan.DefaultGatewayMetric.IsNull() && !plan.DefaultGatewayMetric.IsUnknown() {
		entity["defaultGatewayMetric"] = plan.DefaultGatewayMetric.ValueInt64()
	}

	// Hardware identifiers
	if !plan.SerialNumber.IsNull() && !plan.SerialNumber.IsUnknown() {
		entity["serialNumber"] = plan.SerialNumber.ValueString()
	}

	if !plan.PartNumber.IsNull() && !plan.PartNumber.IsUnknown() {
		entity["partNumber"] = plan.PartNumber.ValueString()
	}

	return entity
}

// parseDeviceFromAPI parses BCM API response into Terraform model.
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

	// Power control configuration
	if powerControl, ok := data["powerControl"].(string); ok && powerControl != "" {
		model.PowerControl = types.StringValue(powerControl)
	} else {
		model.PowerControl = types.StringNull()
	}

	// Network gateway configuration
	if defaultGateway, ok := data["defaultGateway"].(string); ok && defaultGateway != "" {
		model.DefaultGateway = types.StringValue(defaultGateway)
	} else {
		model.DefaultGateway = types.StringNull()
	}

	if defaultGatewayMetric, ok := data["defaultGatewayMetric"].(float64); ok {
		model.DefaultGatewayMetric = types.Int64Value(int64(defaultGatewayMetric))
	} else if defaultGatewayMetricInt, ok := data["defaultGatewayMetric"].(int64); ok {
		model.DefaultGatewayMetric = types.Int64Value(defaultGatewayMetricInt)
	} else {
		model.DefaultGatewayMetric = types.Int64Null()
	}

	// Hardware identifiers
	if serialNumber, ok := data["serialNumber"].(string); ok && serialNumber != "" {
		model.SerialNumber = types.StringValue(serialNumber)
	} else {
		model.SerialNumber = types.StringNull()
	}

	if partNumber, ok := data["partNumber"].(string); ok && partNumber != "" {
		model.PartNumber = types.StringValue(partNumber)
	} else {
		model.PartNumber = types.StringNull()
	}

	// Force is not persisted by BCM, will be preserved from plan/state

	// Parse interfaces from BCM response
	model.Interfaces = parseInterfacesFromAPI(data["interfaces"])

	return model
}
