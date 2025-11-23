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
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &CMDeviceCategoryResource{}
	_ resource.ResourceWithImportState = &CMDeviceCategoryResource{}
)

// CMDeviceCategoryResource defines the resource implementation.
type CMDeviceCategoryResource struct {
	client *BCMClient
}

// CMDeviceCategoryResourceModel describes the resource data model.
type CMDeviceCategoryResourceModel struct {
	// Identity fields
	ID   types.String `tfsdk:"id"`   // Computed, same as UUID
	UUID types.String `tfsdk:"uuid"` // Computed, BCM-assigned
	Name types.String `tfsdk:"name"` // Required, unique

	// Core configuration
	Notes             types.String `tfsdk:"notes"`              // Optional
	ManagementNetwork types.String `tfsdk:"management_network"` // Required, UUID reference

	// Software image reference
	SoftwareImageProxy types.Object `tfsdk:"software_image_proxy"` // Optional, nested SoftwareImageProxyModel

	// Boot configuration
	BootLoader         types.String `tfsdk:"boot_loader"`          // Optional, default: "SYSLINUX"
	BootLoaderFile     types.String `tfsdk:"boot_loader_file"`     // Optional
	BootLoaderProtocol types.String `tfsdk:"boot_loader_protocol"` // Optional, default: "HTTP"

	// Kernel configuration
	KernelVersion       types.String `tfsdk:"kernel_version"`        // Optional
	KernelParameters    types.String `tfsdk:"kernel_parameters"`     // Optional
	KernelOutputConsole types.String `tfsdk:"kernel_output_console"` // Optional
	Modules             types.List   `tfsdk:"modules"`               // Optional, list of KernelModuleModel

	// Disk and storage
	Disksetup types.String `tfsdk:"disksetup"` // Optional, XML string up to 10KB
	Raidconf  types.String `tfsdk:"raidconf"`  // Optional

	// Installation settings
	InstallMode           types.String `tfsdk:"install_mode"`           // Optional, default: "AUTO"
	NewNodeInstallMode    types.String `tfsdk:"new_node_install_mode"`  // Optional, default: "FULL"
	InstallBootRecord     types.Bool   `tfsdk:"install_boot_record"`    // Optional, default: false
	IOScheduler           types.String `tfsdk:"io_scheduler"`           // Optional
	NodeInstallerDisk     types.Bool   `tfsdk:"node_installer_disk"`    // Optional
	VersionConfigFiles    types.Bool   `tfsdk:"version_config_files"`   // Optional
	AuthenticationService types.String `tfsdk:"authentication_service"` // Optional

	// Network configuration
	DefaultGateway         types.String  `tfsdk:"default_gateway"`          // Optional, IP address
	DefaultGatewayMetric   types.Int64   `tfsdk:"default_gateway_metric"`   // Optional
	NameServers            types.List    `tfsdk:"name_servers"`             // Optional, list of strings
	SearchDomains          types.List    `tfsdk:"search_domains"`           // Optional, list of strings
	TimeServers            types.List    `tfsdk:"time_servers"`             // Optional, list of strings
	StaticRoutes           types.Dynamic `tfsdk:"static_routes"`            // Optional, dynamic type (TODO: define proper schema)
	AllowNetworkingRestart types.Bool    `tfsdk:"allow_networking_restart"` // Optional

	// Filesystem configuration
	FSMounts  types.List    `tfsdk:"fsmounts"`  // Optional, list of FSMountModel
	FSExports types.Dynamic `tfsdk:"fsexports"` // Optional, dynamic type (TODO: define proper schema)

	// Role and service assignments
	Roles    types.Dynamic `tfsdk:"roles"`    // Optional, dynamic type (TODO: define proper schema)
	Services types.Dynamic `tfsdk:"services"` // Optional, dynamic type (TODO: define proper schema)

	// Hardware-specific settings
	BMCSettings types.Object  `tfsdk:"bmc_settings"` // Optional, nested BMCSettingsModel
	BiosSetup   types.Object  `tfsdk:"bios_setup"`   // Optional, nested object
	DPUSettings types.Object  `tfsdk:"dpu_settings"` // Optional, nested object
	GPUSettings types.Dynamic `tfsdk:"gpu_settings"` // Optional, dynamic type (TODO: define proper schema)

	// Security and access
	AccessSettings   types.Object `tfsdk:"access_settings"`   // Optional, nested object
	SELinuxSettings  types.Object `tfsdk:"selinux_settings"`  // Optional, nested object
	ProxySettings    types.Object `tfsdk:"proxy_settings"`    // Optional, nested object
	TimeZoneSettings types.Object `tfsdk:"timezone_settings"` // Optional, nested object
	ZTPSettings      types.Object `tfsdk:"ztp_settings"`      // Optional, nested object
	FIPS             types.String `tfsdk:"fips"`              // Optional, "YES" or "NO"

	// Provisioning scripts
	Initialize types.String `tfsdk:"initialize"` // Optional, initialization script
	Finalize   types.String `tfsdk:"finalize"`   // Optional, finalization script

	// Exclude lists (large text fields)
	ExcludeListFull             types.String `tfsdk:"exclude_list_full"`              // Optional, up to 50KB
	ExcludeListGrab             types.String `tfsdk:"exclude_list_grab"`              // Optional, up to 50KB
	ExcludeListGrabnew          types.String `tfsdk:"exclude_list_grabnew"`           // Optional, up to 50KB
	ExcludeListSync             types.String `tfsdk:"exclude_list_sync"`              // Optional, up to 50KB
	ExcludeListUpdate           types.String `tfsdk:"exclude_list_update"`            // Optional, up to 50KB
	ExcludeListManipulateScript types.String `tfsdk:"exclude_list_manipulate_script"` // Optional

	// Behavioral flags
	DataNode          types.Bool   `tfsdk:"data_node"`           // Optional
	InteractiveUser   types.String `tfsdk:"interactive_user"`    // Optional
	UseExclusivelyFor types.String `tfsdk:"use_exclusively_for"` // Optional

	// Force parameter
	Force types.Bool `tfsdk:"force"` // Optional, default: false

	// Computed metadata fields
	ParentUUID  types.String `tfsdk:"parent_uuid"`   // Computed
	Revision    types.String `tfsdk:"revision"`      // Computed
	Modified    types.Bool   `tfsdk:"modified"`      // Computed
	ToBeRemoved types.Bool   `tfsdk:"to_be_removed"` // Computed
	BaseType    types.String `tfsdk:"base_type"`     // Computed, always "Category"
	ChildType   types.String `tfsdk:"child_type"`    // Computed, empty for base type
}

// SoftwareImageProxyModel describes a software image proxy nested object.
type SoftwareImageProxyModel struct {
	UUID                types.String `tfsdk:"uuid"`                  // Computed
	ParentSoftwareImage types.String `tfsdk:"parent_software_image"` // Required, UUID reference
	RevisionID          types.Int64  `tfsdk:"revision_id"`           // Computed
}

// BMCSettingsModel describes BMC configuration nested object.
type BMCSettingsModel struct {
	UUID               types.String  `tfsdk:"uuid"`                 // Computed
	UserName           types.String  `tfsdk:"user_name"`            // Optional
	Password           types.String  `tfsdk:"password"`             // Optional, sensitive
	Privilege          types.String  `tfsdk:"privilege"`            // Optional
	UserID             types.Int64   `tfsdk:"user_id"`              // Optional
	FirmwareManageMode types.String  `tfsdk:"firmware_manage_mode"` // Optional
	LeakPolicy         types.String  `tfsdk:"leak_policy"`          // Optional
	LeakReactionDelay  types.Float64 `tfsdk:"leak_reaction_delay"`  // Optional, seconds
	PowerResetDelay    types.Int64   `tfsdk:"power_reset_delay"`    // Optional, seconds
}

// FSMountModel describes a filesystem mount nested object.
type FSMountModel struct {
	UUID         types.String `tfsdk:"uuid"`         // Computed
	Device       types.String `tfsdk:"device"`       // Required
	Mountpoint   types.String `tfsdk:"mountpoint"`   // Required
	Filesystem   types.String `tfsdk:"filesystem"`   // Required
	MountOptions types.String `tfsdk:"mountoptions"` // Optional
	Fsck         types.String `tfsdk:"fsck"`         // Optional
	Dump         types.Bool   `tfsdk:"dump"`         // Optional
	RDMA         types.Bool   `tfsdk:"rdma"`         // Optional
}

// KernelModuleCategoryModel describes a kernel module nested object.
type KernelModuleCategoryModel struct {
	Name       types.String `tfsdk:"name"`       // Required
	Parameters types.String `tfsdk:"parameters"` // Optional
}

// NewCMDeviceCategoryResource creates a new resource instance.
func NewCMDeviceCategoryResource() resource.Resource {
	return &CMDeviceCategoryResource{}
}

// Metadata returns the resource type name.
func (r *CMDeviceCategoryResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cmdevice_category"
}

// Schema defines the resource schema.
func (r *CMDeviceCategoryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a BCM device category.\n\n" +
			"Device categories define node configuration templates including boot configuration, kernel parameters, " +
			"disk layouts, network settings, and filesystem mounts used to provision compute nodes in the BCM cluster.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Resource identifier (same as UUID)",
			},
			"uuid": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier assigned by BCM",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Category name (must be unique, 1-255 characters)",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"notes": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "User notes or description for the category",
			},
			"management_network": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Management network UUID reference (must be valid RFC 4122 UUID)",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`),
						"must be a valid RFC 4122 UUID format",
					),
				},
			},
			"software_image_proxy": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Software image proxy configuration",
				Attributes: map[string]schema.Attribute{
					"uuid": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Unique identifier",
					},
					"parent_software_image": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Parent software image UUID reference",
					},
					"revision_id": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Revision identifier",
					},
				},
			},
			"boot_loader": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Boot loader type (SYSLINUX, GRUB, GRUB2, PXELINUX). If not specified, BCM assigns a default.",
				Validators: []validator.String{
					stringvalidator.OneOf("SYSLINUX", "GRUB", "GRUB2", "PXELINUX"),
				},
			},
			"boot_loader_file": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Boot loader file path. If not specified, BCM uses defaults based on boot_loader.",
			},
			"boot_loader_protocol": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Boot loader protocol (HTTP, TFTP, NFS). If not specified, BCM assigns a default.",
			},
			"kernel_version": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Kernel version string",
			},
			"kernel_parameters": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Kernel command-line parameters",
			},
			"kernel_output_console": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Kernel output console device",
			},
			"modules": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Kernel modules to load",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Module name",
						},
						"parameters": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Module parameters",
						},
					},
				},
			},
			"disksetup": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Disk setup XML configuration (max 10KB)",
			},
			"raidconf": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "RAID configuration",
			},
			"install_mode": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Installation mode (AUTO, FULL, MINIMAL, CUSTOM). If not specified, BCM assigns a default.",
			},
			"new_node_install_mode": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "New node installation mode (FULL, MINIMAL, SKIP). If not specified, BCM assigns a default.",
			},
			"install_boot_record": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Install boot record flag. If not specified, BCM assigns a default.",
			},
			"io_scheduler": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "I/O scheduler",
			},
			"node_installer_disk": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Node installer disk flag",
			},
			"version_config_files": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Version config files flag",
			},
			"authentication_service": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Authentication service (AUTO, LDAP, SSSD, LOCAL)",
			},
			"default_gateway": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Default gateway IP address. If not specified, BCM may assign a default.",
			},
			"default_gateway_metric": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Default gateway metric. If not specified, BCM may assign a default.",
			},
			"name_servers": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "DNS name servers",
			},
			"search_domains": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "DNS search domains",
			},
			"time_servers": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "NTP time servers",
			},
			"static_routes": schema.DynamicAttribute{
				Optional:            true,
				MarkdownDescription: "Static network routes (dynamic type - TODO: define proper schema)",
			},
			"allow_networking_restart": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Allow networking restart flag",
			},
			"fsmounts": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Filesystem mounts",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"uuid": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Unique identifier",
						},
						"device": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Device path or NFS export",
						},
						"mountpoint": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Mount point path",
						},
						"filesystem": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Filesystem type",
						},
						"mountoptions": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Mount options",
						},
						"fsck": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Filesystem check mode",
						},
						"dump": schema.BoolAttribute{
							Optional:            true,
							MarkdownDescription: "Dump backup flag",
						},
						"rdma": schema.BoolAttribute{
							Optional:            true,
							MarkdownDescription: "Use RDMA for NFS",
						},
					},
				},
			},
			"fsexports": schema.DynamicAttribute{
				Optional:            true,
				MarkdownDescription: "Filesystem exports (dynamic type - TODO: define proper schema)",
			},
			"roles": schema.DynamicAttribute{
				Optional:            true,
				MarkdownDescription: "Role assignments (dynamic type - TODO: define proper schema)",
			},
			"services": schema.DynamicAttribute{
				Optional:            true,
				MarkdownDescription: "Service assignments (dynamic type - TODO: define proper schema)",
			},
			"bmc_settings": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "BMC configuration settings",
				Attributes: map[string]schema.Attribute{
					"uuid": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Unique identifier",
					},
					"user_name": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "BMC username",
					},
					"password": schema.StringAttribute{
						Optional:            true,
						Sensitive:           true,
						MarkdownDescription: "BMC password (sensitive)",
					},
					"privilege": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "BMC privilege level (USER, OPERATOR, ADMINISTRATOR)",
					},
					"user_id": schema.Int64Attribute{
						Optional:            true,
						MarkdownDescription: "BMC user ID",
					},
					"firmware_manage_mode": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Firmware management mode (AUTO, MANUAL, DISABLED)",
					},
					"leak_policy": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Leak policy",
					},
					"leak_reaction_delay": schema.Float64Attribute{
						Optional:            true,
						MarkdownDescription: "Leak reaction delay in seconds",
					},
					"power_reset_delay": schema.Int64Attribute{
						Optional:            true,
						MarkdownDescription: "Power reset delay in seconds",
					},
				},
			},
			"bios_setup": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "BIOS setup configuration",
				Attributes:          map[string]schema.Attribute{},
			},
			"dpu_settings": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "DPU settings",
				Attributes:          map[string]schema.Attribute{},
			},
			"gpu_settings": schema.DynamicAttribute{
				Optional:            true,
				MarkdownDescription: "GPU settings (dynamic type - TODO: define proper schema)",
			},
			"access_settings": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Access settings",
				Attributes:          map[string]schema.Attribute{},
			},
			"selinux_settings": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "SELinux settings",
				Attributes:          map[string]schema.Attribute{},
			},
			"proxy_settings": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Proxy settings",
				Attributes:          map[string]schema.Attribute{},
			},
			"timezone_settings": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Timezone settings",
				Attributes:          map[string]schema.Attribute{},
			},
			"ztp_settings": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "ZTP settings",
				Attributes:          map[string]schema.Attribute{},
			},
			"fips": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "FIPS mode (YES or NO)",
				Validators: []validator.String{
					stringvalidator.OneOf("YES", "NO"),
				},
			},
			"initialize": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Initialization script",
			},
			"finalize": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Finalization script",
			},
			"exclude_list_full": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Exclude list for full operations (max 50KB)",
			},
			"exclude_list_grab": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Exclude list for grab operations (max 50KB)",
			},
			"exclude_list_grabnew": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Exclude list for grabnew operations (max 50KB)",
			},
			"exclude_list_sync": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Exclude list for sync operations (max 50KB)",
			},
			"exclude_list_update": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Exclude list for update operations (max 50KB)",
			},
			"exclude_list_manipulate_script": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Exclude list manipulate script",
			},
			"data_node": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Data node flag",
			},
			"interactive_user": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Interactive user",
			},
			"use_exclusively_for": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Use exclusively for",
			},
			"force": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Force parameter to override warnings and constraints",
			},
			"parent_uuid": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Parent UUID",
			},
			"revision": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Revision",
			},
			"modified": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Modified flag",
			},
			"to_be_removed": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "To be removed flag",
			},
			"base_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Base type (always 'Category')",
			},
			"child_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Child type",
			},
		},
	}
}

// Configure stores the provider client for later use.
func (r *CMDeviceCategoryResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*BCMClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			"Expected *BCMClient, got something else. Please report this issue to the provider developers.",
		)
		return
	}

	r.client = client
}

// Create implements resource.Resource (REFACTOR phase - Real API integration).
func (r *CMDeviceCategoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CMDeviceCategoryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build API entity from plan
	entity := r.buildAPIEntity(ctx, &plan, "")

	// Get force parameter from plan (default to false)
	force := false
	if !plan.Force.IsNull() {
		force = plan.Force.ValueBool()
	}

	tflog.Debug(ctx, "Creating category via BCM API", map[string]interface{}{
		"name":  plan.Name.ValueString(),
		"force": force,
	})

	// Call addCategory API
	body, err := r.client.CallJSONRPC(ctx, "cmdevice", "addCategory", entity, force)
	if err != nil {
		resp.Diagnostics.AddError(
			"Category Creation Failed",
			fmt.Sprintf("Failed to create category '%s': %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	// Parse response - BCM returns validation response when UUID is provided
	// Response format: {"success": true/false, "validation": [...], ...}
	var validationResp struct {
		Success    bool                     `json:"success"`
		Validation []map[string]interface{} `json:"validation"`
	}

	if err := json.Unmarshal(body, &validationResp); err != nil {
		resp.Diagnostics.AddError(
			"Response Parse Error",
			fmt.Sprintf("Failed to parse category creation response: %s", err.Error()),
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
			"Category Creation Failed",
			fmt.Sprintf("Failed to create category '%s': validation errors: %v", plan.Name.ValueString(), errorMsgs),
		)
		return
	}

	// When validation succeeds, BCM returns the UUID in the entity we provided
	createdUUID, ok := entity["uuid"].(string)
	if !ok {
		resp.Diagnostics.AddError(
			"Failed to extract UUID from created category",
			"The UUID returned from BCM API was not a string",
		)
		return
	}

	tflog.Debug(ctx, "Category created successfully", map[string]interface{}{
		"name": plan.Name.ValueString(),
		"uuid": createdUUID,
	})

	// Set UUID in plan model for readCategory
	plan.ID = types.StringValue(createdUUID)
	plan.UUID = types.StringValue(createdUUID)

	// Read back created category to populate all fields
	// BCM has eventual consistency - retry if computed fields are not populated
	// Pattern from terraform-provider-design skill: api_client_patterns.md lines 294-326
	maxRetries := 5
	var lastReadErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Clear diagnostics from previous attempt
		resp.Diagnostics = diag.Diagnostics{}

		r.readCategory(ctx, &plan, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			lastReadErr = fmt.Errorf("read attempt %d failed", attempt+1)
			if attempt < maxRetries-1 {
				// Wait with exponential backoff before retry
				sleepDuration := time.Duration(1<<attempt) * time.Second
				tflog.Warn(ctx, "Category read failed, retrying due to eventual consistency", map[string]interface{}{
					"attempt":       attempt + 1,
					"sleep_seconds": sleepDuration.Seconds(),
				})
				time.Sleep(sleepDuration)
				continue
			}
			// Last attempt failed - return the error
			return
		}

		// Check if computed metadata fields are properly populated
		// BCM should return baseType="Category" for all categories
		fieldsPopulated := !plan.BaseType.IsNull() &&
			!plan.BaseType.IsUnknown() &&
			plan.BaseType.ValueString() == "Category"

		if fieldsPopulated {
			// Fields populated successfully
			tflog.Debug(ctx, "Computed fields populated successfully", map[string]interface{}{
				"attempt":   attempt + 1,
				"base_type": plan.BaseType.ValueString(),
			})
			break
		}

		if attempt < maxRetries-1 {
			// Wait with exponential backoff before retry: 1s, 2s, 4s, 8s, 16s
			sleepDuration := time.Duration(1<<attempt) * time.Second
			tflog.Debug(ctx, "Retrying category read due to eventual consistency (computed fields not populated)", map[string]interface{}{
				"attempt":           attempt + 1,
				"sleep_seconds":     sleepDuration.Seconds(),
				"base_type_is_null": plan.BaseType.IsNull(),
				"base_type_unknown": plan.BaseType.IsUnknown(),
			})
			time.Sleep(sleepDuration)
		} else {
			// Max retries reached - log warning but continue
			tflog.Warn(ctx, "Computed fields not fully populated after retries", map[string]interface{}{
				"attempts":          maxRetries,
				"base_type_is_null": plan.BaseType.IsNull(),
				"base_type_unknown": plan.BaseType.IsUnknown(),
				"last_error":        lastReadErr,
			})
		}
	}

	tflog.Info(ctx, "Created category resource", map[string]interface{}{
		"name": plan.Name.ValueString(),
		"uuid": createdUUID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read implements resource.Resource (REFACTOR phase - Real API integration).
func (r *CMDeviceCategoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CMDeviceCategoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve original management_network and force from state for later comparison
	// These client-side parameters are not returned by BCM API
	originalManagementNetwork := state.ManagementNetwork
	originalForce := state.Force

	// Fetch current state from BCM API with retry for eventual consistency
	// BCM may not return all computed fields immediately after create/update
	maxRetries := 5
	var lastReadErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Clear diagnostics from previous attempt
		resp.Diagnostics = diag.Diagnostics{}

		r.readCategory(ctx, &state, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			// Check if resource was externally deleted
			resourceDeleted := false
			for _, diagnostic := range resp.Diagnostics.Errors() {
				summary := diagnostic.Summary()
				detail := diagnostic.Detail()
				// Check for various "not found" error patterns
				if summary == "Category Not Found" ||
					(summary == "Category Read Failed" && (containsAny(detail, []string{"not found", "unexpected JSON type in response: null"}))) {
					// Resource no longer exists - remove from state
					tflog.Info(ctx, "Category not found during refresh - removing from state", map[string]interface{}{
						"name": state.Name.ValueString(),
					})
					resp.State.RemoveResource(ctx)
					resp.Diagnostics = nil // Clear the diagnostics
					resourceDeleted = true
					break
				}
			}
			if resourceDeleted {
				return
			}

			// Other read error - retry with backoff
			lastReadErr = fmt.Errorf("read attempt %d failed", attempt+1)
			if attempt < maxRetries-1 {
				sleepDuration := time.Duration(1<<attempt) * time.Second
				tflog.Warn(ctx, "Category read failed, retrying due to eventual consistency", map[string]interface{}{
					"attempt":       attempt + 1,
					"sleep_seconds": sleepDuration.Seconds(),
				})
				time.Sleep(sleepDuration)
				continue
			}
			// Last attempt failed - return the error
			return
		}

		// Check if computed metadata fields are properly populated
		// BCM should return baseType="Category" for all categories
		fieldsPopulated := !state.BaseType.IsNull() &&
			!state.BaseType.IsUnknown() &&
			state.BaseType.ValueString() == "Category"

		if fieldsPopulated {
			// Fields populated successfully
			tflog.Debug(ctx, "Computed fields populated successfully in Read", map[string]interface{}{
				"attempt":   attempt + 1,
				"base_type": state.BaseType.ValueString(),
			})
			break
		}

		if attempt < maxRetries-1 {
			// Wait with exponential backoff before retry: 1s, 2s, 4s, 8s, 16s
			sleepDuration := time.Duration(1<<attempt) * time.Second
			tflog.Debug(ctx, "Retrying category read due to eventual consistency (computed fields not populated)", map[string]interface{}{
				"attempt":           attempt + 1,
				"sleep_seconds":     sleepDuration.Seconds(),
				"base_type_is_null": state.BaseType.IsNull(),
				"base_type_unknown": state.BaseType.IsUnknown(),
			})
			time.Sleep(sleepDuration)
		} else {
			// Max retries reached - log warning but continue
			tflog.Warn(ctx, "Computed fields not fully populated after retries in Read", map[string]interface{}{
				"attempts":          maxRetries,
				"base_type_is_null": state.BaseType.IsNull(),
				"base_type_unknown": state.BaseType.IsUnknown(),
				"last_error":        lastReadErr,
			})
		}
	}

	// CRITICAL FIX: BCM may return different management_network UUID during refresh
	// This happens when BCM reassigns the network internally. Preserve the original
	// configured value to avoid drift detection during destroy operations.
	if !originalManagementNetwork.IsNull() && !originalManagementNetwork.IsUnknown() {
		// If BCM returned a different UUID, preserve the original from state
		if !state.ManagementNetwork.Equal(originalManagementNetwork) {
			tflog.Debug(ctx, "Management network changed in BCM, preserving original value", map[string]interface{}{
				"original": originalManagementNetwork.ValueString(),
				"bcm_returned": state.ManagementNetwork.ValueString(),
			})
			state.ManagementNetwork = originalManagementNetwork
		}
	}

	// CRITICAL FIX: Preserve force parameter from state
	// BCM API does not return force (client-side parameter only), but we need to preserve
	// the user's configured value to avoid false drift detection.
	if !originalForce.IsNull() && !originalForce.IsUnknown() {
		state.Force = originalForce
		tflog.Debug(ctx, "Preserved force parameter from state", map[string]interface{}{
			"force": originalForce.ValueBool(),
		})
	} else {
		// If force was never set, keep it as null (not false - that would be a config change)
		state.Force = types.BoolNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update implements resource.Resource (REFACTOR phase - Real API integration).
func (r *CMDeviceCategoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CMDeviceCategoryResourceModel
	var state CMDeviceCategoryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build updated entity from plan (include UUID from state, as UUID is computed)
	// UUID must come from state, not plan, since it's a computed attribute
	entity := r.buildAPIEntity(ctx, &plan, state.UUID.ValueString())

	// Get force parameter from plan
	// IMPORTANT: BCM's updateCategory sometimes requires force=true to bypass validation quirks
	// when updating certain fields even if the name hasn't changed. For safety, we always use force=true
	// for updates unless explicitly set to false by the user.
	force := true // Default to true for updates
	if !plan.Force.IsNull() {
		force = plan.Force.ValueBool()
	}

	// Log the complete update entity for debugging
	entityJSON, _ := json.MarshalIndent(entity, "", "  ")
	tflog.Info(ctx, "Updating category via BCM API", map[string]interface{}{
		"name":       plan.Name.ValueString(),
		"uuid":       state.UUID.ValueString(),
		"force":      force,
		"entity_len": len(string(entityJSON)),
	})

	// Call updateCategory API
	_, err := r.client.CallJSONRPC(ctx, "cmdevice", "updateCategory", entity, force)
	if err != nil {
		resp.Diagnostics.AddError(
			"Category Update Failed",
			fmt.Sprintf("Failed to update category '%s': %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	// Read back updated category with retry for eventual consistency
	// Preserve boot_loader: BCM may reset it to default after update
	planBootLoader := plan.BootLoader
	maxRetries := 5
	var lastReadErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Clear diagnostics from previous attempt
		resp.Diagnostics = diag.Diagnostics{}

		r.readCategory(ctx, &plan, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			lastReadErr = fmt.Errorf("read attempt %d failed", attempt+1)
			if attempt < maxRetries-1 {
				sleepDuration := time.Duration(1<<attempt) * time.Second
				tflog.Warn(ctx, "Category read failed after update, retrying due to eventual consistency", map[string]interface{}{
					"attempt":       attempt + 1,
					"sleep_seconds": sleepDuration.Seconds(),
				})
				time.Sleep(sleepDuration)
				continue
			}
			// Last attempt failed - return the error
			return
		}

		// Check if computed metadata fields are properly populated
		fieldsPopulated := !plan.BaseType.IsNull() &&
			!plan.BaseType.IsUnknown() &&
			plan.BaseType.ValueString() == "Category"

		if fieldsPopulated {
			tflog.Debug(ctx, "Computed fields populated successfully after Update", map[string]interface{}{
				"attempt":   attempt + 1,
				"base_type": plan.BaseType.ValueString(),
			})
			break
		}

		if attempt < maxRetries-1 {
			sleepDuration := time.Duration(1<<attempt) * time.Second
			tflog.Debug(ctx, "Retrying category read after update due to eventual consistency", map[string]interface{}{
				"attempt":           attempt + 1,
				"sleep_seconds":     sleepDuration.Seconds(),
				"base_type_is_null": plan.BaseType.IsNull(),
				"base_type_unknown": plan.BaseType.IsUnknown(),
			})
			time.Sleep(sleepDuration)
		} else {
			tflog.Warn(ctx, "Computed fields not fully populated after retries in Update", map[string]interface{}{
				"attempts":          maxRetries,
				"base_type_is_null": plan.BaseType.IsNull(),
				"base_type_unknown": plan.BaseType.IsUnknown(),
				"last_error":        lastReadErr,
			})
		}
	}

	// CRITICAL FIX: Restore boot_loader from plan if explicitly set, never use Unknown values
	// BCM API may reset boot_loader to default, but we want to preserve the user's setting
	if !planBootLoader.IsNull() && !planBootLoader.IsUnknown() {
		plan.BootLoader = planBootLoader
	}

	tflog.Info(ctx, "Updated category resource", map[string]interface{}{
		"name": plan.Name.ValueString(),
		"uuid": plan.UUID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete implements resource.Resource (REFACTOR phase - Real API integration).
func (r *CMDeviceCategoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CMDeviceCategoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get force parameter from state (default to false)
	force := false
	if !state.Force.IsNull() {
		force = state.Force.ValueBool()
	}

	tflog.Debug(ctx, "Deleting category via BCM API", map[string]interface{}{
		"name":  state.Name.ValueString(),
		"uuid":  state.UUID.ValueString(),
		"force": force,
	})

	// Call removeCategory API
	_, err := r.client.CallJSONRPC(ctx, "cmdevice", "removeCategory", state.UUID.ValueString(), force)
	if err != nil {
		// T042: Enhanced error handling - Parse error to check if it's a "category in use" error
		errStr := err.Error()

		// Check if category was already deleted externally (idempotent delete)
		if containsAny(errStr, []string{"No such category", "not found", "does not exist"}) {
			tflog.Info(ctx, "Category already deleted (idempotent)", map[string]interface{}{
				"name": state.Name.ValueString(),
				"uuid": state.UUID.ValueString(),
			})
			return
		}

		if !force && (containsAny(errStr, []string{"in use", "assigned", "nodes", "cannot be deleted"})) {
			// Category has nodes assigned - provide clear guidance
			resp.Diagnostics.AddError(
				"Category Deletion Failed: Category In Use",
				fmt.Sprintf("Category '%s' cannot be deleted because it has nodes assigned.\n\n"+
					"To delete the category anyway (nodes will remain but lose category reference), "+
					"set force=true in the resource configuration:\n\n"+
					"resource \"bcm_cmdevice_category\" \"%s\" {\n"+
					"  name  = \"%s\"\n"+
					"  force = true\n"+
					"  ...\n"+
					"}\n\n"+
					"Then run 'terraform apply' again to delete with force.\n\n"+
					"Original error: %s",
					state.Name.ValueString(),
					state.Name.ValueString(),
					state.Name.ValueString(),
					err.Error(),
				),
			)
			return
		}

		// T043: Other deletion errors (not related to nodes in use)
		resp.Diagnostics.AddError(
			"Category Deletion Failed",
			fmt.Sprintf("Could not delete category '%s' (UUID: %s): %s",
				state.Name.ValueString(), state.UUID.ValueString(), err.Error()),
		)
		return
	}

	tflog.Info(ctx, "Deleted category resource", map[string]interface{}{
		"name": state.Name.ValueString(),
		"uuid": state.UUID.ValueString(),
	})
}

// containsAny checks if a string contains any of the specified substrings (case-insensitive).
func containsAny(s string, substrings []string) bool {
	// Import strings package functionality inline for case-insensitive check
	lowerStr := ""
	for _, c := range s {
		if c >= 'A' && c <= 'Z' {
			lowerStr += string(c + 32)
		} else {
			lowerStr += string(c)
		}
	}

	for _, substring := range substrings {
		lowerSub := ""
		for _, c := range substring {
			if c >= 'A' && c <= 'Z' {
				lowerSub += string(c + 32)
			} else {
				lowerSub += string(c)
			}
		}

		// Check if lowerStr contains lowerSub
		for i := 0; i <= len(lowerStr)-len(lowerSub); i++ {
			if lowerStr[i:i+len(lowerSub)] == lowerSub {
				return true
			}
		}
	}
	return false
}

// ImportState implements resource.ResourceWithImportState.
func (r *CMDeviceCategoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// T034-T035: Two-phase import implementation
	// Phase 1: Call getCategories to list all categories and find the one with matching UUID
	// Phase 2: Extract the name, then call getCategory(name) to fetch full data

	importedUUID := req.ID
	tflog.Debug(ctx, "Starting two-phase import for category", map[string]interface{}{
		"uuid": importedUUID,
	})

	// Phase 1: List all categories to find the name for this UUID
	body, err := r.client.CallJSONRPC(ctx, "cmdevice", "getCategories")
	if err != nil {
		resp.Diagnostics.AddError(
			"Import Failed",
			fmt.Sprintf("Failed to list categories for import: %s", err.Error()),
		)
		return
	}

	var categories []map[string]interface{}
	if err := json.Unmarshal(body, &categories); err != nil {
		resp.Diagnostics.AddError(
			"Import Failed",
			fmt.Sprintf("Failed to parse category list: %s", err.Error()),
		)
		return
	}

	// Find the category with matching UUID
	var categoryName string
	for _, cat := range categories {
		if uuid, ok := cat["uuid"].(string); ok && uuid == importedUUID {
			if name, ok := cat["name"].(string); ok {
				categoryName = name
				break
			}
		}
	}

	if categoryName == "" {
		resp.Diagnostics.AddError(
			"Category Not Found",
			fmt.Sprintf("No category found with UUID: %s", importedUUID),
		)
		return
	}

	tflog.Debug(ctx, "Found category name for UUID", map[string]interface{}{
		"uuid": importedUUID,
		"name": categoryName,
	})

	// Phase 2: Use readCategory helper to fetch full category data
	// Set name in model so readCategory can use it for lookup
	var model CMDeviceCategoryResourceModel
	model.ID = types.StringValue(importedUUID)
	model.UUID = types.StringValue(importedUUID)
	model.Name = types.StringValue(categoryName)

	// T036: Populate all fields from API response
	r.readCategory(ctx, &model, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Force parameter is not persisted in BCM, set to null
	model.Force = types.BoolNull()

	// Set the populated model in state
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)

	tflog.Trace(ctx, "successfully imported cmdevice category resource", map[string]interface{}{
		"uuid": importedUUID,
		"name": categoryName,
	})
}

// ========================================
// REFACTOR PHASE: Helper Functions
// ========================================

// buildAPIEntity constructs a BCM API Category entity from Terraform model.
func (r *CMDeviceCategoryResource) buildAPIEntity(ctx context.Context, model *CMDeviceCategoryResourceModel, uuid string) map[string]interface{} {
	entity := map[string]interface{}{
		"baseType":      "Category",
		"childType":     "",
		"modified":      true,
		"to_be_removed": false,
		"revision":      "",
	}

	// BCM REQUIRES a UUID for categories (unlike software images)
	// For create: generate new UUID, for update: use existing UUID
	if uuid != "" {
		entity["uuid"] = uuid
	} else {
		// Generate UUID for new category
		entity["uuid"] = generateUUID()
	}

	// Required fields
	if !model.Name.IsNull() {
		entity["name"] = model.Name.ValueString()
	}
	if !model.ManagementNetwork.IsNull() {
		entity["managementNetwork"] = model.ManagementNetwork.ValueString()
	}

	// Optional core fields
	if !model.Notes.IsNull() {
		entity["notes"] = model.Notes.ValueString()
	}

	// Boot configuration (only include if explicitly set)
	if !model.BootLoader.IsNull() && !model.BootLoader.IsUnknown() {
		entity["bootLoader"] = model.BootLoader.ValueString()
	}
	if !model.BootLoaderFile.IsNull() && !model.BootLoaderFile.IsUnknown() {
		entity["bootLoaderFile"] = model.BootLoaderFile.ValueString()
	}
	if !model.BootLoaderProtocol.IsNull() && !model.BootLoaderProtocol.IsUnknown() {
		entity["bootLoaderProtocol"] = model.BootLoaderProtocol.ValueString()
	}

	// Kernel configuration
	if !model.KernelVersion.IsNull() {
		entity["kernelVersion"] = model.KernelVersion.ValueString()
	}
	if !model.KernelParameters.IsNull() {
		entity["kernelParameters"] = model.KernelParameters.ValueString()
	}
	if !model.KernelOutputConsole.IsNull() {
		entity["kernelOutputConsole"] = model.KernelOutputConsole.ValueString()
	}

	// Installation settings (only include if explicitly set)
	if !model.InstallMode.IsNull() && !model.InstallMode.IsUnknown() {
		entity["installMode"] = model.InstallMode.ValueString()
	}
	if !model.NewNodeInstallMode.IsNull() && !model.NewNodeInstallMode.IsUnknown() {
		entity["newNodeInstallMode"] = model.NewNodeInstallMode.ValueString()
	}
	if !model.InstallBootRecord.IsNull() && !model.InstallBootRecord.IsUnknown() {
		entity["installBootRecord"] = model.InstallBootRecord.ValueBool()
	}

	// Network configuration (only include if explicitly set)
	if !model.DefaultGateway.IsNull() && !model.DefaultGateway.IsUnknown() {
		entity["defaultGateway"] = model.DefaultGateway.ValueString()
	}
	if !model.DefaultGatewayMetric.IsNull() && !model.DefaultGatewayMetric.IsUnknown() {
		entity["defaultGatewayMetric"] = model.DefaultGatewayMetric.ValueInt64()
	}

	// Disk and storage
	if !model.Disksetup.IsNull() {
		entity["disksetup"] = model.Disksetup.ValueString()
	}
	if !model.Raidconf.IsNull() {
		entity["raidconf"] = model.Raidconf.ValueString()
	}

	// Nested object: software_image_proxy (minimal support for Phase 4)
	if !model.SoftwareImageProxy.IsNull() {
		var proxyModel SoftwareImageProxyModel
		model.SoftwareImageProxy.As(ctx, &proxyModel, basetypes.ObjectAsOptions{})

		proxyEntity := map[string]interface{}{
			"baseType":      "SoftwareImageProxy",
			"childType":     "",
			"modified":      true,
			"to_be_removed": false,
			"revision":      "",
		}

		// Generate UUID for proxy if not set (create), or use existing (update)
		if !proxyModel.UUID.IsNull() && proxyModel.UUID.ValueString() != "" {
			proxyEntity["uuid"] = proxyModel.UUID.ValueString()
		} else {
			proxyEntity["uuid"] = generateUUID()
		}

		if !proxyModel.ParentSoftwareImage.IsNull() {
			proxyEntity["parentSoftwareImage"] = proxyModel.ParentSoftwareImage.ValueString()
		}

		entity["softwareImageProxy"] = proxyEntity
	}

	// TODO: Add remaining nested objects and arrays in Phase 6 (Comprehensive Schema)
	// - modules (array of KernelModule)
	// - fsmounts (array of FSMount)
	// - bmc_settings (nested BMCSettings)

	return entity
}

// readCategory fetches category data from BCM API using efficient getCategory(name).
func (r *CMDeviceCategoryResource) readCategory(ctx context.Context, model *CMDeviceCategoryResourceModel, diags *diag.Diagnostics) {

	// Determine which identifier to use for lookup
	// Read operations always use name (ImportState handles UUID -> name lookup)
	var lookupName string
	if !model.Name.IsNull() && model.Name.ValueString() != "" {
		lookupName = model.Name.ValueString()
	} else {
		diags.AddError(
			"Invalid State",
			"Cannot read category: name is required for Read operation (ImportState handles UUID lookup)",
		)
		return
	}

	tflog.Debug(ctx, "Reading category via BCM API", map[string]interface{}{
		"name": lookupName,
	})

	// Use efficient getCategory(name) API for direct lookup
	body, err := r.client.CallJSONRPC(ctx, "cmdevice", "getCategory", lookupName)
	if err != nil {
		diags.AddError(
			"Category Read Failed",
			fmt.Sprintf("Failed to read category '%s': %s", lookupName, err.Error()),
		)
		return
	}

	// Check for null response (resource was externally deleted)
	if string(body) == "null" {
		diags.AddError(
			"Category Not Found",
			fmt.Sprintf("Category '%s' not found in BCM (may have been deleted externally)", lookupName),
		)
		return
	}

	// Parse response as single category entity
	var categoryData map[string]interface{}
	if err := json.Unmarshal(body, &categoryData); err != nil {
		diags.AddError(
			"Category Read Failed",
			fmt.Sprintf("Failed to read category '%s': unexpected JSON type in response: %s", lookupName, err.Error()),
		)
		return
	}

	// Check if category not found (empty response)
	if len(categoryData) == 0 {
		diags.AddError(
			"Category Not Found",
			fmt.Sprintf("Category '%s' not found in BCM", lookupName),
		)
		return
	}

	// Map API fields to model using helper functions
	model.ID = getStringValue(categoryData, "uuid")
	model.UUID = getStringValue(categoryData, "uuid")
	model.Name = getStringValue(categoryData, "name")
	model.ManagementNetwork = getStringValue(categoryData, "managementNetwork")

	// Optional core fields
	model.Notes = getStringValue(categoryData, "notes")

	// Boot configuration (Optional+Computed - Terraform handles plan/state automatically)
	model.BootLoader = getStringValue(categoryData, "bootLoader")
	model.BootLoaderFile = getStringValue(categoryData, "bootLoaderFile")
	model.BootLoaderProtocol = getStringValue(categoryData, "bootLoaderProtocol")

	// Kernel configuration
	model.KernelVersion = getStringValue(categoryData, "kernelVersion")
	model.KernelParameters = getStringValue(categoryData, "kernelParameters")
	model.KernelOutputConsole = getStringValue(categoryData, "kernelOutputConsole")

	// Installation settings (Optional+Computed - Terraform handles plan/state automatically)
	model.InstallMode = getStringValue(categoryData, "installMode")
	model.NewNodeInstallMode = getStringValue(categoryData, "newNodeInstallMode")
	model.InstallBootRecord = getBoolValue(categoryData, "installBootRecord")

	// Network configuration (Optional+Computed - Terraform handles plan/state automatically)
	model.DefaultGateway = getStringValue(categoryData, "defaultGateway")
	model.DefaultGatewayMetric = getInt64Value(categoryData, "defaultGatewayMetric")

	// Network lists (set to null for now, Phase 6 will parse these)
	model.NameServers = types.ListNull(types.StringType)
	model.SearchDomains = types.ListNull(types.StringType)
	model.TimeServers = types.ListNull(types.StringType)
	// TODO Phase 6: Define proper schema for static_routes
	model.StaticRoutes = types.DynamicNull()

	// Disk and storage
	model.Disksetup = getStringValue(categoryData, "disksetup")
	model.Raidconf = getStringValue(categoryData, "raidconf")

	// Filesystem lists (set to null for now, Phase 6 will parse these)
	// TODO Phase 6: Parse actual fsmounts from API
	fsMountObjectType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"uuid":         types.StringType,
		"device":       types.StringType,
		"mountpoint":   types.StringType,
		"filesystem":   types.StringType,
		"mountoptions": types.StringType,
		"fsck":         types.StringType,
		"dump":         types.BoolType,
		"rdma":         types.BoolType,
	}}
	model.FSMounts = types.ListNull(fsMountObjectType)
	// TODO Phase 6: Define proper schema for fsexports
	model.FSExports = types.DynamicNull()

	// Kernel modules (use proper KernelModule object type)
	moduleObjectType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"name":       types.StringType,
		"parameters": types.StringType,
	}}
	model.Modules = types.ListNull(moduleObjectType)

	// Role and service lists (set to null for now, Phase 6 will parse these)
	// TODO Phase 6: Define proper schema for roles
	model.Roles = types.DynamicNull()
	// TODO Phase 6: Define proper schema for services
	model.Services = types.DynamicNull()

	// Hardware settings lists (set to null for now, Phase 6 will parse these)
	// TODO Phase 6: Define proper schema for gpu_settings
	model.GPUSettings = types.DynamicNull()

	// Security and access objects (set to null for now, Phase 6 will parse these)
	// TODO Phase 6: Parse actual BMC settings from API
	bmcSettingsObjectType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"uuid":                 types.StringType,
		"user_name":            types.StringType,
		"password":             types.StringType,
		"privilege":            types.StringType,
		"user_id":              types.Int64Type,
		"firmware_manage_mode": types.StringType,
		"leak_policy":          types.StringType,
		"leak_reaction_delay":  types.Float64Type,
		"power_reset_delay":    types.Int64Type,
	}}
	model.BMCSettings = types.ObjectNull(bmcSettingsObjectType.AttrTypes)
	model.BiosSetup = types.ObjectNull(map[string]attr.Type{})
	model.DPUSettings = types.ObjectNull(map[string]attr.Type{})
	model.AccessSettings = types.ObjectNull(map[string]attr.Type{})
	model.SELinuxSettings = types.ObjectNull(map[string]attr.Type{})
	model.ProxySettings = types.ObjectNull(map[string]attr.Type{})
	model.TimeZoneSettings = types.ObjectNull(map[string]attr.Type{})
	model.ZTPSettings = types.ObjectNull(map[string]attr.Type{})

	// Computed metadata fields
	model.BaseType = getStringValue(categoryData, "baseType")
	model.ChildType = getStringValue(categoryData, "childType")
	model.Modified = getBoolValue(categoryData, "modified")
	model.ToBeRemoved = getBoolValue(categoryData, "to_be_removed")
	model.Revision = getStringValue(categoryData, "revision")
	model.ParentUUID = getStringValue(categoryData, "parent_uuid")

	// Parse software_image_proxy (Phase 4 minimal support)
	if proxyData, ok := categoryData["softwareImageProxy"].(map[string]interface{}); ok && proxyData != nil {
		proxyModel := SoftwareImageProxyModel{
			UUID:                getStringValue(proxyData, "uuid"),
			ParentSoftwareImage: getStringValue(proxyData, "parentSoftwareImage"),
			RevisionID:          getInt64Value(proxyData, "revisionID"),
		}

		// Convert to types.Object
		proxyObj, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
			"uuid":                  types.StringType,
			"parent_software_image": types.StringType,
			"revision_id":           types.Int64Type,
		}, proxyModel)
		if diags.HasError() {
			tflog.Error(ctx, "Failed to convert software_image_proxy to object", map[string]interface{}{
				"errors": diags.Errors(),
			})
		} else {
			model.SoftwareImageProxy = proxyObj
		}
	} else {
		// No proxy in response, set to null
		model.SoftwareImageProxy = types.ObjectNull(map[string]attr.Type{
			"uuid":                  types.StringType,
			"parent_software_image": types.StringType,
			"revision_id":           types.Int64Type,
		})
	}

	// TODO: Parse remaining nested objects and arrays in Phase 6 (Comprehensive Schema)
	// - modules (array of KernelModule)
	// - fsmounts (array of FSMount)
	// - bmc_settings (nested BMCSettings)
}

// generateUUID creates a new UUID v4 string.
func generateUUID() string {
	return uuid.New().String()
}
