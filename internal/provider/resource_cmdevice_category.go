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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
	DefaultGateway         types.String `tfsdk:"default_gateway"`          // Optional, IP address
	DefaultGatewayMetric   types.Int64  `tfsdk:"default_gateway_metric"`   // Optional
	NameServers            types.List   `tfsdk:"name_servers"`             // Optional, list of strings
	SearchDomains          types.List   `tfsdk:"search_domains"`           // Optional, list of strings
	TimeServers            types.List   `tfsdk:"time_servers"`             // Optional, list of strings
	StaticRoutes           types.List   `tfsdk:"static_routes"`            // Optional, list of StaticRouteModel
	AllowNetworkingRestart types.Bool   `tfsdk:"allow_networking_restart"` // Optional

	// Filesystem configuration
	FSMounts  types.List `tfsdk:"fsmounts"`  // Optional, list of FSMountModel
	FSExports types.List `tfsdk:"fsexports"` // Optional, list of FSExportModel

	// Role and service assignments
	Roles    types.List `tfsdk:"roles"`    // Optional, list of RoleModel
	Services types.List `tfsdk:"services"` // Optional, list of ServiceModel

	// Hardware-specific settings
	BMCSettings types.Object `tfsdk:"bmc_settings"` // Optional, nested BMCSettingsModel
	BiosSetup   types.Object `tfsdk:"bios_setup"`   // Optional, nested object
	DPUSettings types.Object `tfsdk:"dpu_settings"` // Optional, nested object
	GPUSettings types.List   `tfsdk:"gpu_settings"` // Optional, list of GPUSettingModel

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

// StaticRouteModel describes a static route nested object.
type StaticRouteModel struct {
	Destination types.String `tfsdk:"destination"` // Required, CIDR notation
	Gateway     types.String `tfsdk:"gateway"`     // Required, IPv4 address
	Metric      types.Int64  `tfsdk:"metric"`      // Optional, route priority
}

// FSExportModel describes an NFS filesystem export nested object.
type FSExportModel struct {
	Path       types.String `tfsdk:"path"`        // Required, export path
	Network    types.String `tfsdk:"network"`     // Required, network UUID reference
	AllowWrite types.Bool   `tfsdk:"allow_write"` // Optional, write access
	RootSquash types.Bool   `tfsdk:"root_squash"` // Optional, root squash security
	Async      types.Bool   `tfsdk:"async"`       // Optional, async mode
}

// CategoryRoleModel describes a service role assignment nested object for categories.
// Named differently from data_source_cmdevice_nodes.go RoleModel to avoid conflict.
type CategoryRoleModel struct {
	Name        types.String `tfsdk:"name"`         // Required, role name
	ChildType   types.String `tfsdk:"child_type"`   // Required, role type
	UUID        types.String `tfsdk:"uuid"`         // Computed, BCM-assigned
	AddServices types.Bool   `tfsdk:"add_services"` // Optional, add role services
}

// GPUSettingModel describes a GPU hardware configuration nested object.
type GPUSettingModel struct {
	DeviceID    types.String `tfsdk:"device_id"`    // Required, GPU device ID
	Model       types.String `tfsdk:"model"`        // Optional, GPU model name
	ComputeMode types.String `tfsdk:"compute_mode"` // Optional, compute mode
}

// ServiceModel describes a service configuration nested object.
// Structure based on BCM API - currently empty array in default category.
type ServiceModel struct {
	// TODO: Define fields based on actual BCM API usage
	// Marked as POST-MVP until actual service structure is documented
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uuid": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier assigned by BCM",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"parent_software_image": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Parent software image UUID reference",
					},
					"revision_id": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Revision identifier",
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"boot_loader_file": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Boot loader file path. If not specified, BCM uses defaults based on boot_loader.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"boot_loader_protocol": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Boot loader protocol (HTTP, TFTP, NFS). If not specified, BCM assigns a default.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"new_node_install_mode": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "New node installation mode (FULL, MINIMAL, SKIP). If not specified, BCM assigns a default.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"install_boot_record": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Install boot record flag. If not specified, BCM assigns a default.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"io_scheduler": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "I/O scheduler",
			},
			"node_installer_disk": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Node installer disk flag. If not specified, BCM assigns a default.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"version_config_files": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Version config files flag. If not specified, BCM assigns a default.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"authentication_service": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Authentication service (AUTO, LDAP, SSSD, LOCAL). If not specified, BCM assigns AUTO.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"default_gateway": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Default gateway IP address. If not specified, BCM may assign a default.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"default_gateway_metric": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Default gateway metric. If not specified, BCM may assign a default.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
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
			"static_routes": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Static network routes for nodes in this category. **Known Limitation**: BCM API does not persist this field - values are stored in Terraform state only. After import, re-apply configuration to restore values.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"destination": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Destination network in CIDR notation (e.g., 192.168.1.0/24)",
							Validators: []validator.String{
								stringvalidator.RegexMatches(
									regexp.MustCompile(`^([0-9]{1,3}\.){3}[0-9]{1,3}/[0-9]{1,2}$`),
									"must be valid CIDR notation (e.g., 192.168.1.0/24)",
								),
							},
						},
						"gateway": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Gateway IP address (e.g., 10.0.0.1)",
							Validators: []validator.String{
								stringvalidator.RegexMatches(
									regexp.MustCompile(`^([0-9]{1,3}\.){3}[0-9]{1,3}$`),
									"must be valid IPv4 address",
								),
							},
						},
						"metric": schema.Int64Attribute{
							Optional:            true,
							MarkdownDescription: "Route metric (priority, lower is preferred)",
						},
					},
				},
			},
			"allow_networking_restart": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Allow networking restart flag. If not specified, BCM assigns a default.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
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
			"fsexports": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "NFS filesystem exports for nodes in this category. **Known Limitation**: BCM API does not persist this field - values are stored in Terraform state only. After import, re-apply configuration to restore values.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"path": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Export path (e.g., /home, /shared)",
						},
						"network": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Network UUID reference for export access",
							Validators: []validator.String{
								stringvalidator.RegexMatches(
									regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`),
									"must be valid RFC 4122 UUID",
								),
							},
						},
						"allow_write": schema.BoolAttribute{
							Optional:            true,
							MarkdownDescription: "Allow write access (default: false)",
						},
						"root_squash": schema.BoolAttribute{
							Optional:            true,
							MarkdownDescription: "Enable root squash security (default: false)",
						},
						"async": schema.BoolAttribute{
							Optional:            true,
							MarkdownDescription: "Use async mode for writes (default: false)",
						},
					},
				},
			},
			"roles": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Service role assignments for nodes in this category. **Known Limitation**: BCM API does not persist this field - values are stored in Terraform state only. Role UUIDs are generated locally by the provider. After import, re-apply configuration to restore values.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Role name (e.g., headnode, storage, compute)",
						},
						"child_type": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Role type (e.g., HeadNodeRole, StorageRole, BackupRole)",
						},
						"uuid": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Role UUID (assigned by BCM)",
						},
						"add_services": schema.BoolAttribute{
							Optional:            true,
							MarkdownDescription: "Automatically add role services (default: false)",
						},
					},
				},
			},
			"services": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Service configurations for nodes in this category (structure TBD - marked as POST-MVP). **Known Limitation**: BCM API does not persist this field - values are stored in Terraform state only. After import, re-apply configuration to restore values.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						// TODO: Define service fields based on actual BCM API usage
						// Placeholder for now - services field structure needs investigation
					},
				},
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
			"gpu_settings": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "GPU hardware configuration for nodes in this category. **Known Limitation**: BCM API does not persist this field - values are stored in Terraform state only. After import, re-apply configuration to restore values.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"device_id": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "GPU device ID (e.g., 0, 1, 2)",
						},
						"model": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "GPU model name (e.g., Tesla V100, A100)",
						},
						"compute_mode": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Compute mode (default, exclusive, prohibited)",
						},
					},
				},
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
				Computed:            true,
				MarkdownDescription: "FIPS mode (YES or NO). If not specified, BCM assigns NO.",
				Validators: []validator.String{
					stringvalidator.OneOf("YES", "NO"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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
				Computed:            true,
				MarkdownDescription: "Data node flag. If not specified, BCM assigns a default.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"interactive_user": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Interactive user. If not specified, BCM assigns ALWAYS.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"revision": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Revision",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"modified": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Modified flag",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"to_be_removed": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "To be removed flag",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"base_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Base type (always 'Category')",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"child_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Child type",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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

	// Preserve plan values for optional list fields that BCM API may not return/store
	// These will be restored after readCategory to avoid "inconsistent result after apply" errors
	// BCM returns empty arrays for these fields even when we send data, so we preserve plan values
	planStaticRoutes := plan.StaticRoutes
	planFSExports := plan.FSExports
	planFSMounts := plan.FSMounts // Issue #84: Preserve fsmounts
	planRoles := plan.Roles
	planGPUSettings := plan.GPUSettings
	planServices := plan.Services
	// Issue #82: Preserve BMC settings (password is sensitive and not returned by API)
	planBMCSettings := plan.BMCSettings

	// Build API entity from plan
	entity := r.buildAPIEntity(ctx, &plan, "")

	// Pre-flight validation: Call validateCategory before CREATE
	validationErrors, err := r.client.ValidateEntity(ctx, "CMDevice", "validateCategory", entity, true)
	if err != nil {
		resp.Diagnostics.AddError(
			"Validation API Error",
			fmt.Sprintf("Could not validate category '%s': %s", plan.Name.ValueString(), err.Error()),
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

	// Preserve plan's software_image_proxy before reading (BCM API may return different values)
	planSoftwareImageProxy := plan.SoftwareImageProxy

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

	// CRITICAL FIX: Restore parent_software_image from plan while keeping computed fields
	// BCM API may return different parent_software_image UUID on reads
	// But we must preserve the user's configured value
	if !planSoftwareImageProxy.IsNull() && !planSoftwareImageProxy.IsUnknown() {
		var planProxy SoftwareImageProxyModel
		planSoftwareImageProxy.As(ctx, &planProxy, basetypes.ObjectAsOptions{})

		// Only restore if we have a valid parent_software_image from plan
		if !planProxy.ParentSoftwareImage.IsNull() && !planProxy.ParentSoftwareImage.IsUnknown() {
			// Get the computed values from the API response (uuid, revision_id)
			var apiProxy SoftwareImageProxyModel
			if !plan.SoftwareImageProxy.IsNull() {
				plan.SoftwareImageProxy.As(ctx, &apiProxy, basetypes.ObjectAsOptions{})
			}

			// Build merged proxy: user's parent_software_image + API's computed fields
			mergedProxy := SoftwareImageProxyModel{
				UUID:                apiProxy.UUID,
				ParentSoftwareImage: planProxy.ParentSoftwareImage, // Preserve user config
				RevisionID:          apiProxy.RevisionID,
			}

			// Convert back to types.Object
			proxyObj, diagsProxy := types.ObjectValueFrom(ctx, map[string]attr.Type{
				"uuid":                  types.StringType,
				"parent_software_image": types.StringType,
				"revision_id":           types.Int64Type,
			}, mergedProxy)
			if !diagsProxy.HasError() {
				plan.SoftwareImageProxy = proxyObj
				tflog.Debug(ctx, "Merged software_image_proxy: preserved parent_software_image from plan", map[string]interface{}{
					"parent_software_image": planProxy.ParentSoftwareImage.ValueString(),
				})
			}
		}
	}

	tflog.Info(ctx, "Created category resource", map[string]interface{}{
		"name": plan.Name.ValueString(),
		"uuid": createdUUID,
	})

	// Restore plan values for optional list fields that BCM doesn't persist
	// BCM returns empty arrays for these fields, so we preserve what the user configured
	// This ensures Terraform state matches plan when BCM doesn't store these values
	plan.StaticRoutes = planStaticRoutes
	plan.FSExports = planFSExports
	// Issue #84 FIX: Merge fsmounts instead of unconditional overwrite
	// This preserves user config (device, mountpoint, filesystem, etc.) while populating
	// computed values (uuid) from BCM API response
	plan.FSMounts = mergeFSMountsWithAPIResponse(ctx, planFSMounts, plan.FSMounts)
	// Issue #83 FIX: Merge roles instead of unconditional overwrite
	// This preserves user config (name, child_type, add_services) while populating
	// computed values (uuid) from BCM API response
	plan.Roles = mergeRolesWithAPIResponse(ctx, planRoles, plan.Roles)
	plan.GPUSettings = planGPUSettings
	plan.Services = planServices

	// Issue #82 FIX: Preserve bmc_settings from plan for sensitive object consistency
	// For objects containing sensitive attributes, Terraform requires the returned state
	// to match the plan exactly (except for computed fields marked UseStateForUnknown).
	// Since we can't mark individual nested attributes as UseStateForUnknown, we need to:
	// 1. Use plan values for ALL user-configured fields (including password)
	// 2. Use API values ONLY for computed fields (uuid)
	// 3. Keep null values as null (don't replace with API defaults)
	if !planBMCSettings.IsNull() && !planBMCSettings.IsUnknown() {
		var planBMC BMCSettingsModel
		planBMCSettings.As(ctx, &planBMC, basetypes.ObjectAsOptions{})

		// Get API-returned values for computed fields only
		var apiBMC BMCSettingsModel
		if !plan.BMCSettings.IsNull() {
			plan.BMCSettings.As(ctx, &apiBMC, basetypes.ObjectAsOptions{})
		}

		// Build merged model:
		// - UUID: From API (computed)
		// - ALL other fields: From plan (to ensure consistency with Terraform's expected values)
		mergedBMC := BMCSettingsModel{
			UUID:               apiBMC.UUID,                // Only computed field from API
			UserName:           planBMC.UserName,           // From plan
			Password:           planBMC.Password,           // From plan (sensitive)
			Privilege:          planBMC.Privilege,          // From plan
			UserID:             planBMC.UserID,             // From plan
			FirmwareManageMode: planBMC.FirmwareManageMode, // From plan (null if not set)
			LeakPolicy:         planBMC.LeakPolicy,         // From plan (null if not set)
			LeakReactionDelay:  planBMC.LeakReactionDelay,  // From plan (null if not set)
			PowerResetDelay:    planBMC.PowerResetDelay,    // From plan (null if not set)
		}

		// Convert back to types.Object
		bmcSettingsObjectType := map[string]attr.Type{
			"uuid":                 types.StringType,
			"user_name":            types.StringType,
			"password":             types.StringType,
			"privilege":            types.StringType,
			"user_id":              types.Int64Type,
			"firmware_manage_mode": types.StringType,
			"leak_policy":          types.StringType,
			"leak_reaction_delay":  types.Float64Type,
			"power_reset_delay":    types.Int64Type,
		}

		bmcObj, diagsBMC := types.ObjectValueFrom(ctx, bmcSettingsObjectType, mergedBMC)
		if !diagsBMC.HasError() {
			plan.BMCSettings = bmcObj
			tflog.Debug(ctx, "Preserved bmc_settings from plan in Create", map[string]interface{}{
				"has_password": !planBMC.Password.IsNull(),
				"user_name":    planBMC.UserName.ValueString(),
			})
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read implements resource.Resource (REFACTOR phase - Real API integration).
func (r *CMDeviceCategoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CMDeviceCategoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve original values from state for fields BCM API doesn't persist/return correctly
	// These will be restored after readCategory to avoid false drift detection
	originalManagementNetwork := state.ManagementNetwork
	originalForce := state.Force
	originalSoftwareImageProxy := state.SoftwareImageProxy
	originalStaticRoutes := state.StaticRoutes
	originalFSExports := state.FSExports
	originalFSMounts := state.FSMounts // Issue #84: Preserve fsmounts
	originalRoles := state.Roles
	originalGPUSettings := state.GPUSettings
	originalServices := state.Services
	// Issue #82: Preserve BMC settings (password is sensitive and not returned by API)
	originalBMCSettings := state.BMCSettings

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
				"original":     originalManagementNetwork.ValueString(),
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

	// CRITICAL FIX: Preserve optional list fields from state
	// BCM API doesn't persist these fields for categories - preserve user's configured values
	state.StaticRoutes = originalStaticRoutes
	state.FSExports = originalFSExports
	// Issue #84 FIX: Merge fsmounts instead of unconditional overwrite
	// This preserves user config (device, mountpoint, filesystem, etc.) while populating
	// computed values (uuid) from BCM API response
	state.FSMounts = mergeFSMountsWithAPIResponse(ctx, originalFSMounts, state.FSMounts)
	// Issue #83 FIX: Merge roles instead of unconditional overwrite
	// This preserves user config (name, child_type, add_services) while populating
	// computed values (uuid) from BCM API response
	state.Roles = mergeRolesWithAPIResponse(ctx, originalRoles, state.Roles)
	state.GPUSettings = originalGPUSettings
	state.Services = originalServices

	// CRITICAL FIX: Preserve parent_software_image from state while keeping computed fields
	// BCM API may return different parent_software_image UUID on subsequent reads,
	// causing false drift detection. Preserve the user's configured value.
	if !originalSoftwareImageProxy.IsNull() && !originalSoftwareImageProxy.IsUnknown() {
		var stateProxy SoftwareImageProxyModel
		originalSoftwareImageProxy.As(ctx, &stateProxy, basetypes.ObjectAsOptions{})

		// Only restore if we have a valid parent_software_image from state
		if !stateProxy.ParentSoftwareImage.IsNull() && !stateProxy.ParentSoftwareImage.IsUnknown() {
			// Get the computed values from the API response (uuid, revision_id)
			var apiProxy SoftwareImageProxyModel
			if !state.SoftwareImageProxy.IsNull() {
				state.SoftwareImageProxy.As(ctx, &apiProxy, basetypes.ObjectAsOptions{})
			}

			// Build merged proxy: user's parent_software_image + API's computed fields
			mergedProxy := SoftwareImageProxyModel{
				UUID:                apiProxy.UUID,
				ParentSoftwareImage: stateProxy.ParentSoftwareImage, // Preserve user config
				RevisionID:          apiProxy.RevisionID,
			}

			// Convert back to types.Object
			proxyObj, diagsProxy := types.ObjectValueFrom(ctx, map[string]attr.Type{
				"uuid":                  types.StringType,
				"parent_software_image": types.StringType,
				"revision_id":           types.Int64Type,
			}, mergedProxy)
			if !diagsProxy.HasError() {
				state.SoftwareImageProxy = proxyObj
				tflog.Debug(ctx, "Merged software_image_proxy: preserved parent_software_image from state", map[string]interface{}{
					"parent_software_image": stateProxy.ParentSoftwareImage.ValueString(),
				})
			}
		}
	}

	// Issue #82 FIX: Preserve bmc_settings from state for consistency
	// For objects containing sensitive attributes, we must preserve the state values
	// to avoid drift detection. BCM API does not return password (sensitive).
	if !originalBMCSettings.IsNull() && !originalBMCSettings.IsUnknown() {
		var originalBMC BMCSettingsModel
		originalBMCSettings.As(ctx, &originalBMC, basetypes.ObjectAsOptions{})

		// Get API-returned values for computed fields only
		var apiBMC BMCSettingsModel
		if !state.BMCSettings.IsNull() {
			state.BMCSettings.As(ctx, &apiBMC, basetypes.ObjectAsOptions{})
		}

		// Build merged model:
		// - UUID: From API (computed)
		// - ALL other fields: From prior state (to maintain consistency and avoid drift)
		mergedBMC := BMCSettingsModel{
			UUID:               apiBMC.UUID,                    // Only computed field from API
			UserName:           originalBMC.UserName,           // From state
			Password:           originalBMC.Password,           // From state (sensitive)
			Privilege:          originalBMC.Privilege,          // From state
			UserID:             originalBMC.UserID,             // From state
			FirmwareManageMode: originalBMC.FirmwareManageMode, // From state
			LeakPolicy:         originalBMC.LeakPolicy,         // From state
			LeakReactionDelay:  originalBMC.LeakReactionDelay,  // From state
			PowerResetDelay:    originalBMC.PowerResetDelay,    // From state
		}

		// Convert back to types.Object
		bmcSettingsObjectType := map[string]attr.Type{
			"uuid":                 types.StringType,
			"user_name":            types.StringType,
			"password":             types.StringType,
			"privilege":            types.StringType,
			"user_id":              types.Int64Type,
			"firmware_manage_mode": types.StringType,
			"leak_policy":          types.StringType,
			"leak_reaction_delay":  types.Float64Type,
			"power_reset_delay":    types.Int64Type,
		}

		bmcObj, diagsBMC := types.ObjectValueFrom(ctx, bmcSettingsObjectType, mergedBMC)
		if !diagsBMC.HasError() {
			state.BMCSettings = bmcObj
			tflog.Debug(ctx, "Preserved bmc_settings from state in Read", map[string]interface{}{
				"has_password": !originalBMC.Password.IsNull(),
			})
		}
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

	// Pre-flight validation: Call validateCategory before UPDATE
	validationErrors, err := r.client.ValidateEntity(ctx, "CMDevice", "validateCategory", entity, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Validation API Error",
			fmt.Sprintf("Could not validate category '%s': %s", plan.Name.ValueString(), err.Error()),
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
	_, err = r.client.CallJSONRPC(ctx, "cmdevice", "updateCategory", entity, force)
	if err != nil {
		resp.Diagnostics.AddError(
			"Category Update Failed",
			fmt.Sprintf("Failed to update category '%s': %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	// Preserve plan values for fields BCM may reset or not persist after update
	planBootLoader := plan.BootLoader
	planSoftwareImageProxy := plan.SoftwareImageProxy
	planStaticRoutes := plan.StaticRoutes
	planFSExports := plan.FSExports
	planFSMounts := plan.FSMounts // Issue #84: Preserve fsmounts
	planRoles := plan.Roles
	planGPUSettings := plan.GPUSettings
	planServices := plan.Services
	// Issue #82: Preserve BMC settings (password is sensitive and not returned by API)
	planBMCSettings := plan.BMCSettings

	// Read back updated category with retry for eventual consistency
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

	// CRITICAL FIX: Restore optional list fields from plan
	// BCM API doesn't persist these fields for categories - preserve user's configured values
	plan.StaticRoutes = planStaticRoutes
	plan.FSExports = planFSExports
	// Issue #84 FIX: Merge fsmounts instead of unconditional overwrite
	// This preserves user config (device, mountpoint, filesystem, etc.) while populating
	// computed values (uuid) from BCM API response
	plan.FSMounts = mergeFSMountsWithAPIResponse(ctx, planFSMounts, plan.FSMounts)
	// Issue #83 FIX: Merge roles instead of unconditional overwrite
	// This preserves user config (name, child_type, add_services) while populating
	// computed values (uuid) from BCM API response
	plan.Roles = mergeRolesWithAPIResponse(ctx, planRoles, plan.Roles)
	plan.GPUSettings = planGPUSettings
	plan.Services = planServices

	// CRITICAL FIX: Preserve parent_software_image from plan while keeping computed fields
	// BCM API may return different parent_software_image UUID on reads
	if !planSoftwareImageProxy.IsNull() && !planSoftwareImageProxy.IsUnknown() {
		var planProxy SoftwareImageProxyModel
		planSoftwareImageProxy.As(ctx, &planProxy, basetypes.ObjectAsOptions{})

		// Only restore if we have a valid parent_software_image from plan
		if !planProxy.ParentSoftwareImage.IsNull() && !planProxy.ParentSoftwareImage.IsUnknown() {
			// Get the computed values from the API response (uuid, revision_id)
			var apiProxy SoftwareImageProxyModel
			if !plan.SoftwareImageProxy.IsNull() {
				plan.SoftwareImageProxy.As(ctx, &apiProxy, basetypes.ObjectAsOptions{})
			}

			// Build merged proxy: user's parent_software_image + API's computed fields
			mergedProxy := SoftwareImageProxyModel{
				UUID:                apiProxy.UUID,
				ParentSoftwareImage: planProxy.ParentSoftwareImage, // Preserve user config
				RevisionID:          apiProxy.RevisionID,
			}

			// Convert back to types.Object
			proxyObj, diagsProxy := types.ObjectValueFrom(ctx, map[string]attr.Type{
				"uuid":                  types.StringType,
				"parent_software_image": types.StringType,
				"revision_id":           types.Int64Type,
			}, mergedProxy)
			if !diagsProxy.HasError() {
				plan.SoftwareImageProxy = proxyObj
				tflog.Debug(ctx, "Merged software_image_proxy: preserved parent_software_image from plan in Update", map[string]interface{}{
					"parent_software_image": planProxy.ParentSoftwareImage.ValueString(),
				})
			}
		}
	}

	// Issue #82 FIX: Preserve bmc_settings from plan for consistency
	// For objects containing sensitive attributes, we must preserve the plan values
	// to avoid "inconsistent values for sensitive attribute" errors.
	if !planBMCSettings.IsNull() && !planBMCSettings.IsUnknown() {
		var planBMC BMCSettingsModel
		planBMCSettings.As(ctx, &planBMC, basetypes.ObjectAsOptions{})

		// Get API-returned values for computed fields only
		var apiBMC BMCSettingsModel
		if !plan.BMCSettings.IsNull() {
			plan.BMCSettings.As(ctx, &apiBMC, basetypes.ObjectAsOptions{})
		}

		// Build merged model:
		// - UUID: From API (computed)
		// - ALL other fields: From plan (to ensure consistency)
		mergedBMC := BMCSettingsModel{
			UUID:               apiBMC.UUID,                // Only computed field from API
			UserName:           planBMC.UserName,           // From plan
			Password:           planBMC.Password,           // From plan (sensitive)
			Privilege:          planBMC.Privilege,          // From plan
			UserID:             planBMC.UserID,             // From plan
			FirmwareManageMode: planBMC.FirmwareManageMode, // From plan
			LeakPolicy:         planBMC.LeakPolicy,         // From plan
			LeakReactionDelay:  planBMC.LeakReactionDelay,  // From plan
			PowerResetDelay:    planBMC.PowerResetDelay,    // From plan
		}

		// Convert back to types.Object
		bmcSettingsObjectType := map[string]attr.Type{
			"uuid":                 types.StringType,
			"user_name":            types.StringType,
			"password":             types.StringType,
			"privilege":            types.StringType,
			"user_id":              types.Int64Type,
			"firmware_manage_mode": types.StringType,
			"leak_policy":          types.StringType,
			"leak_reaction_delay":  types.Float64Type,
			"power_reset_delay":    types.Int64Type,
		}

		bmcObj, diagsBMC := types.ObjectValueFrom(ctx, bmcSettingsObjectType, mergedBMC)
		if !diagsBMC.HasError() {
			plan.BMCSettings = bmcObj
			tflog.Debug(ctx, "Preserved bmc_settings from plan in Update", map[string]interface{}{
				"has_password": !planBMC.Password.IsNull(),
			})
		}
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

	// PROACTIVE DEPENDENCY CHECK: Check for dependent devices before deletion (unless force=true)
	if !force {
		tflog.Debug(ctx, "Performing dependency check for category deletion", map[string]interface{}{
			"category_uuid": state.UUID.ValueString(),
			"category_name": state.Name.ValueString(),
		})

		result, err := CheckDevicesInCategory(ctx, r.client, state.UUID.ValueString())
		if err != nil {
			// Dependency check failed - log warning but allow user to proceed with force
			resp.Diagnostics.AddWarning(
				"Dependency Check Failed",
				fmt.Sprintf(
					"Unable to verify dependencies for category '%s': %s\n\n"+
						"You can proceed with deletion by setting 'force = true', "+
						"but this may create orphaned references if dependencies exist.",
					state.Name.ValueString(),
					err.Error(),
				),
			)
			return
		}

		if result.HasDependencies {
			// Dependencies exist - block deletion with detailed error message
			tflog.Info(ctx, "Category deletion blocked due to dependencies", map[string]interface{}{
				"category_uuid":   state.UUID.ValueString(),
				"category_name":   state.Name.ValueString(),
				"dependent_count": result.DependentCount,
				"dependent_type":  result.DependentType,
			})

			resp.Diagnostics.AddError(
				"Category In Use - Cannot Delete",
				BuildDependencyError(
					"Category",
					state.Name.ValueString(),
					"device",
					result.Identifiers,
				),
			)
			return
		}

		tflog.Debug(ctx, "No dependencies found - proceeding with deletion", map[string]interface{}{
			"category_uuid": state.UUID.ValueString(),
			"category_name": state.Name.ValueString(),
		})
	} else {
		// Force deletion - log warning about potential orphaned references
		tflog.Warn(ctx, BuildForceDeleteionWarning("Category", state.Name.ValueString()), map[string]interface{}{
			"category_uuid": state.UUID.ValueString(),
			"category_name": state.Name.ValueString(),
			"force":         true,
		})
	}

	// Call removeCategory API
	_, err := r.client.CallJSONRPC(ctx, "cmdevice", "removeCategory", state.UUID.ValueString(), force)
	if err != nil {
		// Enhanced error handling - Parse error to check if already deleted (idempotent)
		errStr := err.Error()

		// Check if category was already deleted externally (idempotent delete)
		if containsAny(errStr, []string{"No such category", "not found", "does not exist"}) {
			tflog.Info(ctx, "Category already deleted (idempotent)", map[string]interface{}{
				"name": state.Name.ValueString(),
				"uuid": state.UUID.ValueString(),
			})
			return
		}

		// Other deletion errors
		resp.Diagnostics.AddError(
			"Category Deletion Failed",
			fmt.Sprintf("Could not delete category '%s' (UUID: %s): %s",
				state.Name.ValueString(), state.UUID.ValueString(), err.Error()),
		)
		return
	}

	tflog.Info(ctx, "Deleted category resource", map[string]interface{}{
		"name":  state.Name.ValueString(),
		"uuid":  state.UUID.ValueString(),
		"force": force,
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

	// Network list fields
	if !model.NameServers.IsNull() && !model.NameServers.IsUnknown() {
		var servers []string
		model.NameServers.ElementsAs(ctx, &servers, false)
		entity["nameServers"] = servers
	}
	if !model.SearchDomains.IsNull() && !model.SearchDomains.IsUnknown() {
		var domains []string
		model.SearchDomains.ElementsAs(ctx, &domains, false)
		entity["searchDomains"] = domains
	}
	if !model.TimeServers.IsNull() && !model.TimeServers.IsUnknown() {
		var servers []string
		model.TimeServers.ElementsAs(ctx, &servers, false)
		entity["timeServers"] = servers
	}

	// Disk and storage
	if !model.Disksetup.IsNull() {
		entity["disksetup"] = model.Disksetup.ValueString()
	}
	if !model.Raidconf.IsNull() {
		entity["raidconf"] = model.Raidconf.ValueString()
	}

	// I/O scheduler
	if !model.IOScheduler.IsNull() && !model.IOScheduler.IsUnknown() {
		entity["ioScheduler"] = model.IOScheduler.ValueString()
	}

	// FIPS setting (T019)
	if !model.FIPS.IsNull() && !model.FIPS.IsUnknown() {
		entity["fips"] = model.FIPS.ValueString()
	}

	// Behavioral flags (T020-T022)
	if !model.DataNode.IsNull() && !model.DataNode.IsUnknown() {
		entity["dataNode"] = model.DataNode.ValueBool()
	}
	if !model.InteractiveUser.IsNull() && !model.InteractiveUser.IsUnknown() {
		entity["interactiveUser"] = model.InteractiveUser.ValueString()
	}
	if !model.UseExclusivelyFor.IsNull() && !model.UseExclusivelyFor.IsUnknown() {
		entity["useExclusivelyFor"] = model.UseExclusivelyFor.ValueString()
	}

	// Installation additional settings (T023-T025)
	if !model.NodeInstallerDisk.IsNull() && !model.NodeInstallerDisk.IsUnknown() {
		entity["nodeInstallerDisk"] = model.NodeInstallerDisk.ValueBool()
	}
	if !model.VersionConfigFiles.IsNull() && !model.VersionConfigFiles.IsUnknown() {
		entity["versionConfigFiles"] = model.VersionConfigFiles.ValueBool()
	}
	if !model.AuthenticationService.IsNull() && !model.AuthenticationService.IsUnknown() {
		entity["authenticationService"] = model.AuthenticationService.ValueString()
	}

	// Allow networking restart
	if !model.AllowNetworkingRestart.IsNull() && !model.AllowNetworkingRestart.IsUnknown() {
		entity["allowNetworkingRestart"] = model.AllowNetworkingRestart.ValueBool()
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

	// Serialize static_routes (Terraform snake_case → BCM camelCase)
	if !model.StaticRoutes.IsNull() && !model.StaticRoutes.IsUnknown() {
		var routes []StaticRouteModel
		diags := model.StaticRoutes.ElementsAs(ctx, &routes, false)
		if !diags.HasError() {
			routesList := make([]map[string]interface{}, 0, len(routes))
			for _, route := range routes {
				routeMap := map[string]interface{}{
					"destination": route.Destination.ValueString(),
					"gateway":     route.Gateway.ValueString(),
				}
				if !route.Metric.IsNull() {
					routeMap["metric"] = route.Metric.ValueInt64()
				}
				routesList = append(routesList, routeMap)
			}
			entity["staticRoutes"] = routesList
		}
	}

	// Serialize fsexports (snake_case → camelCase for BCM API)
	if !model.FSExports.IsNull() && !model.FSExports.IsUnknown() {
		var exports []FSExportModel
		diags := model.FSExports.ElementsAs(ctx, &exports, false)
		if !diags.HasError() {
			exportsList := make([]map[string]interface{}, 0, len(exports))
			for _, export := range exports {
				exportMap := map[string]interface{}{
					"baseType": "FSExport",
					"path":     export.Path.ValueString(),
					"network":  export.Network.ValueString(),
				}
				if !export.AllowWrite.IsNull() {
					exportMap["allowWrite"] = export.AllowWrite.ValueBool()
				}
				if !export.RootSquash.IsNull() {
					exportMap["rootSquash"] = export.RootSquash.ValueBool()
				}
				if !export.Async.IsNull() {
					exportMap["async"] = export.Async.ValueBool()
				}
				exportsList = append(exportsList, exportMap)
			}
			entity["fsexports"] = exportsList
		}
	}

	// Serialize roles (snake_case → camelCase for BCM API)
	if !model.Roles.IsNull() && !model.Roles.IsUnknown() {
		var roles []CategoryRoleModel
		diags := model.Roles.ElementsAs(ctx, &roles, false)
		if !diags.HasError() {
			rolesList := make([]map[string]interface{}, 0, len(roles))
			for _, role := range roles {
				roleMap := map[string]interface{}{
					"baseType":  "Role",
					"name":      role.Name.ValueString(),
					"childType": role.ChildType.ValueString(),
				}
				// Include UUID if present (for updates), BCM assigns on create
				if !role.UUID.IsNull() && role.UUID.ValueString() != "" {
					roleMap["uuid"] = role.UUID.ValueString()
				}
				if !role.AddServices.IsNull() {
					roleMap["addServices"] = role.AddServices.ValueBool()
				}
				rolesList = append(rolesList, roleMap)
			}
			entity["roles"] = rolesList
		}
	}

	// Serialize gpu_settings (Terraform snake_case → BCM camelCase)
	if !model.GPUSettings.IsNull() && !model.GPUSettings.IsUnknown() {
		var gpuSettings []GPUSettingModel
		diags := model.GPUSettings.ElementsAs(ctx, &gpuSettings, false)
		if !diags.HasError() {
			gpuList := make([]map[string]interface{}, 0, len(gpuSettings))
			for _, gpu := range gpuSettings {
				gpuMap := map[string]interface{}{
					"baseType": "GPUSetting",
					"deviceId": gpu.DeviceID.ValueString(),
				}
				if !gpu.Model.IsNull() {
					gpuMap["model"] = gpu.Model.ValueString()
				}
				if !gpu.ComputeMode.IsNull() {
					gpuMap["computeMode"] = gpu.ComputeMode.ValueString()
				}
				gpuList = append(gpuList, gpuMap)
			}
			entity["gpuSettings"] = gpuList
		}
	}

	// Serialize services (POST-MVP - currently empty list or null)
	if !model.Services.IsNull() && !model.Services.IsUnknown() {
		// Services field structure is TBD - send empty array if set
		entity["services"] = []map[string]interface{}{}
	}

	// Provisioning scripts
	if !model.Initialize.IsNull() && !model.Initialize.IsUnknown() {
		entity["initialize"] = model.Initialize.ValueString()
	}
	if !model.Finalize.IsNull() && !model.Finalize.IsUnknown() {
		entity["finalize"] = model.Finalize.ValueString()
	}

	// Exclude lists (large text fields)
	if !model.ExcludeListFull.IsNull() && !model.ExcludeListFull.IsUnknown() {
		entity["excludeListFull"] = model.ExcludeListFull.ValueString()
	}
	if !model.ExcludeListGrab.IsNull() && !model.ExcludeListGrab.IsUnknown() {
		entity["excludeListGrab"] = model.ExcludeListGrab.ValueString()
	}
	if !model.ExcludeListGrabnew.IsNull() && !model.ExcludeListGrabnew.IsUnknown() {
		entity["excludeListGrabnew"] = model.ExcludeListGrabnew.ValueString()
	}
	if !model.ExcludeListSync.IsNull() && !model.ExcludeListSync.IsUnknown() {
		entity["excludeListSync"] = model.ExcludeListSync.ValueString()
	}
	if !model.ExcludeListUpdate.IsNull() && !model.ExcludeListUpdate.IsUnknown() {
		entity["excludeListUpdate"] = model.ExcludeListUpdate.ValueString()
	}
	if !model.ExcludeListManipulateScript.IsNull() && !model.ExcludeListManipulateScript.IsUnknown() {
		entity["excludeListManipulateScript"] = model.ExcludeListManipulateScript.ValueString()
	}

	// BMC Settings nested object
	if !model.BMCSettings.IsNull() && !model.BMCSettings.IsUnknown() {
		var bmcModel BMCSettingsModel
		model.BMCSettings.As(ctx, &bmcModel, basetypes.ObjectAsOptions{})

		bmcEntity := map[string]interface{}{
			"baseType":      "BMCSettings",
			"childType":     "",
			"modified":      true,
			"to_be_removed": false,
		}

		if !bmcModel.UUID.IsNull() && bmcModel.UUID.ValueString() != "" {
			bmcEntity["uuid"] = bmcModel.UUID.ValueString()
		} else {
			bmcEntity["uuid"] = generateUUID()
		}
		if !bmcModel.UserName.IsNull() {
			bmcEntity["userName"] = bmcModel.UserName.ValueString()
		}
		if !bmcModel.Password.IsNull() {
			bmcEntity["password"] = bmcModel.Password.ValueString()
		}
		if !bmcModel.Privilege.IsNull() {
			bmcEntity["privilege"] = bmcModel.Privilege.ValueString()
		}
		if !bmcModel.UserID.IsNull() {
			bmcEntity["userID"] = bmcModel.UserID.ValueInt64()
		}
		if !bmcModel.FirmwareManageMode.IsNull() {
			bmcEntity["firmwareManageMode"] = bmcModel.FirmwareManageMode.ValueString()
		}
		if !bmcModel.LeakPolicy.IsNull() {
			bmcEntity["leakPolicy"] = bmcModel.LeakPolicy.ValueString()
		}
		if !bmcModel.LeakReactionDelay.IsNull() {
			bmcEntity["leakReactionDelay"] = bmcModel.LeakReactionDelay.ValueFloat64()
		}
		if !bmcModel.PowerResetDelay.IsNull() {
			bmcEntity["powerResetDelay"] = bmcModel.PowerResetDelay.ValueInt64()
		}

		entity["bmcSettings"] = bmcEntity
	}

	// Kernel modules list
	if !model.Modules.IsNull() && !model.Modules.IsUnknown() {
		var modules []KernelModuleCategoryModel
		model.Modules.ElementsAs(ctx, &modules, false)

		var moduleEntities []map[string]interface{}
		for _, mod := range modules {
			moduleEntity := map[string]interface{}{
				"baseType":      "KernelModule",
				"childType":     "",
				"modified":      true,
				"to_be_removed": false,
			}
			if !mod.Name.IsNull() {
				moduleEntity["name"] = mod.Name.ValueString()
			}
			if !mod.Parameters.IsNull() {
				moduleEntity["parameters"] = mod.Parameters.ValueString()
			}
			moduleEntities = append(moduleEntities, moduleEntity)
		}
		entity["modules"] = moduleEntities
	}

	// Serialize fsmounts (snake_case → camelCase for BCM API)
	// Issue #84: Implement fsmounts field that was previously marked as Phase 6 TODO
	if !model.FSMounts.IsNull() && !model.FSMounts.IsUnknown() {
		var mounts []FSMountModel
		diags := model.FSMounts.ElementsAs(ctx, &mounts, false)
		if !diags.HasError() {
			mountsList := make([]map[string]interface{}, 0, len(mounts))
			for _, mount := range mounts {
				mountMap := map[string]interface{}{
					"baseType": "FSMount",
					"device":   mount.Device.ValueString(),
					"path":     mount.Mountpoint.ValueString(), // mountpoint -> path
					"type":     mount.Filesystem.ValueString(), // filesystem -> type
				}
				// Include UUID if present (for updates), BCM assigns on create
				if !mount.UUID.IsNull() && mount.UUID.ValueString() != "" {
					mountMap["uuid"] = mount.UUID.ValueString()
				}
				// Handle optional fields
				if !mount.MountOptions.IsNull() {
					mountMap["options"] = mount.MountOptions.ValueString() // mountoptions -> options
				}
				if !mount.Fsck.IsNull() {
					mountMap["fsck"] = mount.Fsck.ValueString()
				}
				if !mount.Dump.IsNull() {
					mountMap["dump"] = mount.Dump.ValueBool()
				}
				if !mount.RDMA.IsNull() {
					mountMap["rdma"] = mount.RDMA.ValueBool()
				}
				mountsList = append(mountsList, mountMap)
			}
			entity["fsmounts"] = mountsList
		}
	}

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

	// I/O Scheduler
	model.IOScheduler = getStringValue(categoryData, "ioScheduler")

	// Network configuration (Optional+Computed - Terraform handles plan/state automatically)
	model.DefaultGateway = getStringValue(categoryData, "defaultGateway")
	model.DefaultGatewayMetric = getInt64Value(categoryData, "defaultGatewayMetric")

	// Network lists - parse from API
	model.NameServers = parseStringListValue(ctx, categoryData, "nameServers")
	model.SearchDomains = parseStringListValue(ctx, categoryData, "searchDomains")
	model.TimeServers = parseStringListValue(ctx, categoryData, "timeServers")

	// Parse static_routes from BCM API (camelCase → snake_case)
	staticRouteObjectType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"destination": types.StringType,
		"gateway":     types.StringType,
		"metric":      types.Int64Type,
	}}
	if routesData, ok := categoryData["staticRoutes"].([]interface{}); ok {
		// BCM returns array (empty or with data) - convert to Terraform list
		routeValues := make([]attr.Value, 0, len(routesData))
		for _, routeRaw := range routesData {
			if routeMap, ok := routeRaw.(map[string]interface{}); ok {
				routeObj, objDiags := types.ObjectValue(staticRouteObjectType.AttrTypes, map[string]attr.Value{
					"destination": getStringValue(routeMap, "destination"),
					"gateway":     getStringValue(routeMap, "gateway"),
					"metric":      getInt64Value(routeMap, "metric"),
				})
				if !objDiags.HasError() {
					routeValues = append(routeValues, routeObj)
				}
			}
		}
		model.StaticRoutes, _ = types.ListValue(staticRouteObjectType, routeValues)
	} else {
		// Field not present in response - set to null
		model.StaticRoutes = types.ListNull(staticRouteObjectType)
	}

	// Disk and storage
	model.Disksetup = getStringValue(categoryData, "disksetup")
	model.Raidconf = getStringValue(categoryData, "raidconf")

	// FIPS setting (T019)
	model.FIPS = getStringValue(categoryData, "fips")

	// Behavioral flags (T020-T022)
	model.DataNode = getBoolValue(categoryData, "dataNode")
	model.InteractiveUser = getStringValue(categoryData, "interactiveUser")
	model.UseExclusivelyFor = getStringValue(categoryData, "useExclusivelyFor")

	// Installation additional settings (T023-T025)
	model.NodeInstallerDisk = getBoolValue(categoryData, "nodeInstallerDisk")
	model.VersionConfigFiles = getBoolValue(categoryData, "versionConfigFiles")
	model.AuthenticationService = getStringValue(categoryData, "authenticationService")

	// Allow networking restart
	model.AllowNetworkingRestart = getBoolValue(categoryData, "allowNetworkingRestart")

	// Provisioning scripts
	model.Initialize = getStringValue(categoryData, "initialize")
	model.Finalize = getStringValue(categoryData, "finalize")

	// Exclude lists
	model.ExcludeListFull = getStringValue(categoryData, "excludeListFull")
	model.ExcludeListGrab = getStringValue(categoryData, "excludeListGrab")
	model.ExcludeListGrabnew = getStringValue(categoryData, "excludeListGrabnew")
	model.ExcludeListSync = getStringValue(categoryData, "excludeListSync")
	model.ExcludeListUpdate = getStringValue(categoryData, "excludeListUpdate")
	model.ExcludeListManipulateScript = getStringValue(categoryData, "excludeListManipulateScript")

	// Parse fsmounts from BCM API (camelCase → snake_case)
	// Issue #84: Implement fsmounts parsing that was previously set to null
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
	if mountsData, ok := categoryData["fsmounts"].([]interface{}); ok && len(mountsData) > 0 {
		// BCM returns array with data - convert to Terraform list
		mountValues := make([]attr.Value, 0, len(mountsData))
		for _, mountRaw := range mountsData {
			if mountMap, ok := mountRaw.(map[string]interface{}); ok {
				mountObj, objDiags := types.ObjectValue(fsMountObjectType.AttrTypes, map[string]attr.Value{
					"uuid":         getStringValue(mountMap, "uuid"),
					"device":       getStringValue(mountMap, "device"),
					"mountpoint":   getStringValue(mountMap, "path"),    // path -> mountpoint
					"filesystem":   getStringValue(mountMap, "type"),    // type -> filesystem
					"mountoptions": getStringValue(mountMap, "options"), // options -> mountoptions
					"fsck":         getStringValue(mountMap, "fsck"),
					"dump":         getBoolValue(mountMap, "dump"),
					"rdma":         getBoolValue(mountMap, "rdma"),
				})
				if !objDiags.HasError() {
					mountValues = append(mountValues, mountObj)
				}
			}
		}
		model.FSMounts, _ = types.ListValue(fsMountObjectType, mountValues)
	} else {
		// Field not present in response or empty - set to null
		model.FSMounts = types.ListNull(fsMountObjectType)
	}

	// Parse fsexports from BCM API (camelCase → snake_case)
	fsExportObjectType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"path":        types.StringType,
		"network":     types.StringType,
		"allow_write": types.BoolType,
		"root_squash": types.BoolType,
		"async":       types.BoolType,
	}}
	if exportsData, ok := categoryData["fsexports"].([]interface{}); ok {
		// BCM returns array (empty or with data) - convert to Terraform list
		exportValues := make([]attr.Value, 0, len(exportsData))
		for _, exportRaw := range exportsData {
			if exportMap, ok := exportRaw.(map[string]interface{}); ok {
				exportObj, objDiags := types.ObjectValue(fsExportObjectType.AttrTypes, map[string]attr.Value{
					"path":        getStringValue(exportMap, "path"),
					"network":     getStringValue(exportMap, "network"),
					"allow_write": getBoolValue(exportMap, "allowWrite"),
					"root_squash": getBoolValue(exportMap, "rootSquash"),
					"async":       getBoolValue(exportMap, "async"),
				})
				if !objDiags.HasError() {
					exportValues = append(exportValues, exportObj)
				}
			}
		}
		model.FSExports, _ = types.ListValue(fsExportObjectType, exportValues)
	} else {
		// Field not present in response - set to null
		model.FSExports = types.ListNull(fsExportObjectType)
	}

	// Parse kernel modules from API response
	moduleObjectType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"name":       types.StringType,
		"parameters": types.StringType,
	}}
	if modulesData, ok := categoryData["modules"].([]interface{}); ok && modulesData != nil {
		var moduleModels []KernelModuleCategoryModel
		for _, modData := range modulesData {
			if mod, ok := modData.(map[string]interface{}); ok {
				moduleModels = append(moduleModels, KernelModuleCategoryModel{
					Name:       getStringValue(mod, "name"),
					Parameters: getStringValue(mod, "parameters"),
				})
			}
		}
		if len(moduleModels) > 0 {
			moduleList, diags := types.ListValueFrom(ctx, moduleObjectType, moduleModels)
			if !diags.HasError() {
				model.Modules = moduleList
			} else {
				model.Modules = types.ListNull(moduleObjectType)
			}
		} else {
			model.Modules = types.ListNull(moduleObjectType)
		}
	} else {
		model.Modules = types.ListNull(moduleObjectType)
	}

	// Parse roles from BCM API (camelCase → snake_case)
	roleObjectType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"name":         types.StringType,
		"child_type":   types.StringType,
		"uuid":         types.StringType,
		"add_services": types.BoolType,
	}}
	if rolesData, ok := categoryData["roles"].([]interface{}); ok {
		// BCM returns array (empty or with data) - convert to Terraform list
		roleValues := make([]attr.Value, 0, len(rolesData))
		for _, roleRaw := range rolesData {
			if roleMap, ok := roleRaw.(map[string]interface{}); ok {
				roleObj, objDiags := types.ObjectValue(roleObjectType.AttrTypes, map[string]attr.Value{
					"name":         getStringValue(roleMap, "name"),
					"child_type":   getStringValue(roleMap, "childType"),
					"uuid":         getStringValue(roleMap, "uuid"),
					"add_services": getBoolValue(roleMap, "addServices"),
				})
				if !objDiags.HasError() {
					roleValues = append(roleValues, roleObj)
				}
			}
		}
		model.Roles, _ = types.ListValue(roleObjectType, roleValues)
	} else {
		// Field not present in response - set to null
		model.Roles = types.ListNull(roleObjectType)
	}

	// Parse services from BCM API (POST-MVP - structure TBD)
	serviceObjectType := types.ObjectType{AttrTypes: map[string]attr.Type{
		// Empty for now - services field structure is TBD
	}}
	if _, ok := categoryData["services"].([]interface{}); ok {
		// BCM returns array (empty or with data) - set to empty list
		model.Services, _ = types.ListValue(serviceObjectType, []attr.Value{})
	} else {
		// Field not present in response - set to null
		model.Services = types.ListNull(serviceObjectType)
	}

	// Parse gpu_settings from BCM API (camelCase → snake_case)
	gpuSettingObjectType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"device_id":    types.StringType,
		"model":        types.StringType,
		"compute_mode": types.StringType,
	}}
	if gpuData, ok := categoryData["gpuSettings"].([]interface{}); ok {
		// BCM returns array (empty or with data) - convert to Terraform list
		gpuValues := make([]attr.Value, 0, len(gpuData))
		for _, gpuRaw := range gpuData {
			if gpuMap, ok := gpuRaw.(map[string]interface{}); ok {
				gpuObj, objDiags := types.ObjectValue(gpuSettingObjectType.AttrTypes, map[string]attr.Value{
					"device_id":    getStringValue(gpuMap, "deviceId"),
					"model":        getStringValue(gpuMap, "model"),
					"compute_mode": getStringValue(gpuMap, "computeMode"),
				})
				if !objDiags.HasError() {
					gpuValues = append(gpuValues, gpuObj)
				}
			}
		}
		model.GPUSettings, _ = types.ListValue(gpuSettingObjectType, gpuValues)
	} else {
		// Field not present in response - set to null
		model.GPUSettings = types.ListNull(gpuSettingObjectType)
	}

	// BMC Settings nested object - parse from API response
	bmcSettingsObjectType := map[string]attr.Type{
		"uuid":                 types.StringType,
		"user_name":            types.StringType,
		"password":             types.StringType,
		"privilege":            types.StringType,
		"user_id":              types.Int64Type,
		"firmware_manage_mode": types.StringType,
		"leak_policy":          types.StringType,
		"leak_reaction_delay":  types.Float64Type,
		"power_reset_delay":    types.Int64Type,
	}

	if bmcData, ok := categoryData["bmcSettings"].(map[string]interface{}); ok && bmcData != nil {
		bmcModel := BMCSettingsModel{
			UUID:               getStringValue(bmcData, "uuid"),
			UserName:           getStringValue(bmcData, "userName"),
			Password:           types.StringNull(), // Don't read back password (sensitive)
			Privilege:          getStringValue(bmcData, "privilege"),
			UserID:             getInt64Value(bmcData, "userID"),
			FirmwareManageMode: getStringValue(bmcData, "firmwareManageMode"),
			LeakPolicy:         getStringValue(bmcData, "leakPolicy"),
			LeakReactionDelay:  getFloat64Value(bmcData, "leakReactionDelay"),
			PowerResetDelay:    getInt64Value(bmcData, "powerResetDelay"),
		}

		bmcObj, bmcDiags := types.ObjectValueFrom(ctx, bmcSettingsObjectType, bmcModel)
		if !bmcDiags.HasError() {
			model.BMCSettings = bmcObj
		} else {
			tflog.Error(ctx, "Failed to convert bmc_settings to object", map[string]interface{}{
				"errors": bmcDiags.Errors(),
			})
			model.BMCSettings = types.ObjectNull(bmcSettingsObjectType)
		}
	} else {
		model.BMCSettings = types.ObjectNull(bmcSettingsObjectType)
	}

	// Other security and access objects (set to null for now, Phase 6 will parse these)
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

// mergeRolesWithAPIResponse merges user-configured role attributes with BCM API-computed values.
// It matches roles by name and:
// - Preserves user-specified: name, child_type, add_services
// - Populates computed: uuid from BCM API response
//
// This fixes issue #83 where roles[].uuid was never populated because
// the Read operation unconditionally overwrote API response with original state.
//
// Parameters:
//   - ctx: Context for logging
//   - originalRoles: Roles from Terraform state (before API call)
//   - apiRoles: Roles parsed from BCM API response (contains computed UUIDs)
//
// Returns:
//   - Merged list with user config + computed UUIDs
func mergeRolesWithAPIResponse(ctx context.Context, originalRoles types.List, apiRoles types.List) types.List {
	// If no original roles in plan (null), preserve null to avoid plan->state inconsistency
	// BCM returns empty list [] for roles even when user didn't specify any
	if originalRoles.IsNull() {
		tflog.Debug(ctx, "Original roles is null, preserving null")
		roleObjectType := types.ObjectType{AttrTypes: map[string]attr.Type{
			"name":         types.StringType,
			"child_type":   types.StringType,
			"uuid":         types.StringType,
			"add_services": types.BoolType,
		}}
		return types.ListNull(roleObjectType)
	}

	// If original is unknown (during plan), return API response if available
	if originalRoles.IsUnknown() {
		tflog.Debug(ctx, "Original roles is unknown, using API response")
		return apiRoles
	}

	// If API returned null/unknown, preserve original (handles API quirks)
	if apiRoles.IsNull() || apiRoles.IsUnknown() {
		tflog.Debug(ctx, "API returned null/unknown roles, preserving original")
		return originalRoles
	}

	// Extract role models from both lists
	var origRoles []CategoryRoleModel
	var apiRolesList []CategoryRoleModel

	if diags := originalRoles.ElementsAs(ctx, &origRoles, false); diags.HasError() {
		tflog.Warn(ctx, "Failed to parse original roles, preserving as-is", map[string]interface{}{
			"errors": diags.Errors(),
		})
		return originalRoles
	}
	if diags := apiRoles.ElementsAs(ctx, &apiRolesList, false); diags.HasError() {
		tflog.Warn(ctx, "Failed to parse API roles, preserving original", map[string]interface{}{
			"errors": diags.Errors(),
		})
		return originalRoles
	}

	// Build lookup map: name -> API role (for efficient matching)
	apiRolesByName := make(map[string]CategoryRoleModel)
	for _, role := range apiRolesList {
		if !role.Name.IsNull() && !role.Name.IsUnknown() {
			apiRolesByName[role.Name.ValueString()] = role
		}
	}

	// Merge: preserve user config fields + populate computed UUID from API
	mergedRoles := make([]CategoryRoleModel, 0, len(origRoles))
	for _, origRole := range origRoles {
		roleName := origRole.Name.ValueString()
		if apiRole, found := apiRolesByName[roleName]; found {
			// Match found - preserve user config, populate computed UUID from API
			mergedRole := CategoryRoleModel{
				Name:        origRole.Name,        // Preserve user value
				ChildType:   origRole.ChildType,   // Preserve user value
				AddServices: origRole.AddServices, // Preserve user value
				UUID:        apiRole.UUID,         // Populate from API
			}
			mergedRoles = append(mergedRoles, mergedRole)
			tflog.Debug(ctx, "Merged role with API UUID", map[string]interface{}{
				"name": roleName,
				"uuid": apiRole.UUID.ValueString(),
			})
		} else {
			// Role not found in API response - BCM doesn't persist category roles
			// Generate a UUID if one doesn't exist (handles Unknown/null UUIDs from plan)
			mergedRole := CategoryRoleModel{
				Name:        origRole.Name,
				ChildType:   origRole.ChildType,
				AddServices: origRole.AddServices,
			}

			// Generate UUID if original doesn't have a known one
			if origRole.UUID.IsNull() || origRole.UUID.IsUnknown() || origRole.UUID.ValueString() == "" {
				newUUID := generateUUID()
				mergedRole.UUID = types.StringValue(newUUID)
				tflog.Debug(ctx, "Generated UUID for role (BCM doesn't persist category roles)", map[string]interface{}{
					"name": roleName,
					"uuid": newUUID,
				})
			} else {
				// Preserve existing UUID (from previous state)
				mergedRole.UUID = origRole.UUID
				tflog.Debug(ctx, "Preserved existing UUID for role", map[string]interface{}{
					"name": roleName,
					"uuid": origRole.UUID.ValueString(),
				})
			}
			mergedRoles = append(mergedRoles, mergedRole)
		}
	}

	// Convert back to types.List
	roleObjectType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"name":         types.StringType,
		"child_type":   types.StringType,
		"uuid":         types.StringType,
		"add_services": types.BoolType,
	}}

	roleValues := make([]attr.Value, 0, len(mergedRoles))
	for _, role := range mergedRoles {
		roleObj, diags := types.ObjectValue(roleObjectType.AttrTypes, map[string]attr.Value{
			"name":         role.Name,
			"child_type":   role.ChildType,
			"uuid":         role.UUID,
			"add_services": role.AddServices,
		})
		if !diags.HasError() {
			roleValues = append(roleValues, roleObj)
		}
	}

	result, _ := types.ListValue(roleObjectType, roleValues)
	tflog.Debug(ctx, "Role merge complete", map[string]interface{}{
		"original_count": len(origRoles),
		"api_count":      len(apiRolesList),
		"merged_count":   len(mergedRoles),
	})
	return result
}

// generateUUID creates a new UUID v4 string.
func generateUUID() string {
	return uuid.New().String()
}

// mergeFSMountsWithAPIResponse merges user-configured fsmount attributes with BCM API-computed values.
// It matches mounts by device+mountpoint combination and:
// - Preserves user-specified: device, mountpoint, filesystem, mountoptions, fsck, dump, rdma
// - Populates computed: uuid from BCM API response
//
// This fixes issue #84 where fsmounts[].uuid was never populated because
// the fsmounts field was always set to null.
//
// Parameters:
//   - ctx: Context for logging
//   - originalMounts: Mounts from Terraform plan/state (before API call)
//   - apiMounts: Mounts parsed from BCM API response (contains computed UUIDs)
//
// Returns:
//   - Merged list with user config + computed UUIDs
func mergeFSMountsWithAPIResponse(ctx context.Context, originalMounts types.List, apiMounts types.List) types.List {
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

	// If no original mounts in plan (null), preserve null to avoid plan->state inconsistency
	// BCM returns empty array [] for fsmounts even when user didn't specify any
	if originalMounts.IsNull() {
		tflog.Debug(ctx, "Original fsmounts is null, preserving null")
		return types.ListNull(fsMountObjectType)
	}

	// If original is unknown (during plan), return API response if available
	if originalMounts.IsUnknown() {
		tflog.Debug(ctx, "Original fsmounts is unknown, using API response")
		return apiMounts
	}

	// If API returned null/unknown, preserve original (handles API quirks)
	if apiMounts.IsNull() || apiMounts.IsUnknown() {
		tflog.Debug(ctx, "API returned null/unknown fsmounts, preserving original")
		// Need to ensure UUIDs are populated for original mounts
		var origMounts []FSMountModel
		if diags := originalMounts.ElementsAs(ctx, &origMounts, false); diags.HasError() {
			return originalMounts
		}

		// Generate UUIDs for mounts that don't have them
		mountValues := make([]attr.Value, 0, len(origMounts))
		for _, mount := range origMounts {
			mountUUID := mount.UUID
			if mount.UUID.IsNull() || mount.UUID.IsUnknown() || mount.UUID.ValueString() == "" {
				mountUUID = types.StringValue(generateUUID())
				tflog.Debug(ctx, "Generated UUID for fsmount (BCM didn't persist)", map[string]interface{}{
					"device":     mount.Device.ValueString(),
					"mountpoint": mount.Mountpoint.ValueString(),
					"uuid":       mountUUID.ValueString(),
				})
			}

			mountObj, diags := types.ObjectValue(fsMountObjectType.AttrTypes, map[string]attr.Value{
				"uuid":         mountUUID,
				"device":       mount.Device,
				"mountpoint":   mount.Mountpoint,
				"filesystem":   mount.Filesystem,
				"mountoptions": mount.MountOptions,
				"fsck":         mount.Fsck,
				"dump":         mount.Dump,
				"rdma":         mount.RDMA,
			})
			if !diags.HasError() {
				mountValues = append(mountValues, mountObj)
			}
		}
		result, _ := types.ListValue(fsMountObjectType, mountValues)
		return result
	}

	// Extract mount models from both lists
	var origMounts []FSMountModel
	var apiMountsList []FSMountModel

	if diags := originalMounts.ElementsAs(ctx, &origMounts, false); diags.HasError() {
		tflog.Warn(ctx, "Failed to parse original fsmounts, preserving as-is", map[string]interface{}{
			"errors": diags.Errors(),
		})
		return originalMounts
	}
	if diags := apiMounts.ElementsAs(ctx, &apiMountsList, false); diags.HasError() {
		tflog.Warn(ctx, "Failed to parse API fsmounts, preserving original", map[string]interface{}{
			"errors": diags.Errors(),
		})
		return originalMounts
	}

	// Build lookup map: device+mountpoint -> API mount (for efficient matching)
	apiMountsByKey := make(map[string]FSMountModel)
	for _, mount := range apiMountsList {
		if !mount.Device.IsNull() && !mount.Device.IsUnknown() &&
			!mount.Mountpoint.IsNull() && !mount.Mountpoint.IsUnknown() {
			key := fmt.Sprintf("%s:%s", mount.Device.ValueString(), mount.Mountpoint.ValueString())
			apiMountsByKey[key] = mount
		}
	}

	// Merge: preserve user config fields + populate computed UUID from API
	mergedMounts := make([]FSMountModel, 0, len(origMounts))
	for _, origMount := range origMounts {
		key := fmt.Sprintf("%s:%s", origMount.Device.ValueString(), origMount.Mountpoint.ValueString())
		if apiMount, found := apiMountsByKey[key]; found {
			// Match found - preserve user config, populate computed UUID from API
			mergedMount := FSMountModel{
				UUID:         apiMount.UUID,          // Populate from API
				Device:       origMount.Device,       // Preserve user value
				Mountpoint:   origMount.Mountpoint,   // Preserve user value
				Filesystem:   origMount.Filesystem,   // Preserve user value
				MountOptions: origMount.MountOptions, // Preserve user value
				Fsck:         origMount.Fsck,         // Preserve user value
				Dump:         origMount.Dump,         // Preserve user value
				RDMA:         origMount.RDMA,         // Preserve user value
			}
			mergedMounts = append(mergedMounts, mergedMount)
			tflog.Debug(ctx, "Merged fsmount with API UUID", map[string]interface{}{
				"device":     origMount.Device.ValueString(),
				"mountpoint": origMount.Mountpoint.ValueString(),
				"uuid":       apiMount.UUID.ValueString(),
			})
		} else {
			// Mount not found in API response - BCM doesn't persist category fsmounts
			// Generate a UUID if one doesn't exist (handles Unknown/null UUIDs from plan)
			mergedMount := FSMountModel{
				Device:       origMount.Device,
				Mountpoint:   origMount.Mountpoint,
				Filesystem:   origMount.Filesystem,
				MountOptions: origMount.MountOptions,
				Fsck:         origMount.Fsck,
				Dump:         origMount.Dump,
				RDMA:         origMount.RDMA,
			}

			// Generate UUID if original doesn't have a known one
			if origMount.UUID.IsNull() || origMount.UUID.IsUnknown() || origMount.UUID.ValueString() == "" {
				newUUID := generateUUID()
				mergedMount.UUID = types.StringValue(newUUID)
				tflog.Debug(ctx, "Generated UUID for fsmount (BCM doesn't persist category fsmounts)", map[string]interface{}{
					"device":     origMount.Device.ValueString(),
					"mountpoint": origMount.Mountpoint.ValueString(),
					"uuid":       newUUID,
				})
			} else {
				// Preserve existing UUID (from previous state)
				mergedMount.UUID = origMount.UUID
				tflog.Debug(ctx, "Preserved existing UUID for fsmount", map[string]interface{}{
					"device":     origMount.Device.ValueString(),
					"mountpoint": origMount.Mountpoint.ValueString(),
					"uuid":       origMount.UUID.ValueString(),
				})
			}
			mergedMounts = append(mergedMounts, mergedMount)
		}
	}

	// Convert back to types.List
	mountValues := make([]attr.Value, 0, len(mergedMounts))
	for _, mount := range mergedMounts {
		mountObj, diags := types.ObjectValue(fsMountObjectType.AttrTypes, map[string]attr.Value{
			"uuid":         mount.UUID,
			"device":       mount.Device,
			"mountpoint":   mount.Mountpoint,
			"filesystem":   mount.Filesystem,
			"mountoptions": mount.MountOptions,
			"fsck":         mount.Fsck,
			"dump":         mount.Dump,
			"rdma":         mount.RDMA,
		})
		if !diags.HasError() {
			mountValues = append(mountValues, mountObj)
		}
	}

	result, _ := types.ListValue(fsMountObjectType, mountValues)
	tflog.Debug(ctx, "FSMount merge complete", map[string]interface{}{
		"original_count": len(origMounts),
		"api_count":      len(apiMountsList),
		"merged_count":   len(mergedMounts),
	})
	return result
}

// parseStringListValue parses a string array from API response into types.List.
// Returns a null list if the key doesn't exist or the value is empty.
func parseStringListValue(ctx context.Context, data map[string]interface{}, key string) types.List {
	if val, ok := data[key]; ok && val != nil {
		if arr, ok := val.([]interface{}); ok {
			var items []string
			for _, item := range arr {
				if s, ok := item.(string); ok {
					items = append(items, s)
				}
			}
			// Return list even if empty - maintains consistency with Terraform plan
			list, _ := types.ListValueFrom(ctx, types.StringType, items)
			return list
		}
	}
	return types.ListNull(types.StringType)
}

// Issue #82: Coalesce helper functions for preserving plan values over API values
// These return the first non-null, non-unknown value, preferring plan values

// coalesceString returns the first non-null, non-unknown string value
func coalesceString(values ...types.String) types.String {
	for _, v := range values {
		if !v.IsNull() && !v.IsUnknown() {
			return v
		}
	}
	// All values are null or unknown, return null
	return types.StringNull()
}

// coalesceInt64 returns the first non-null, non-unknown int64 value
func coalesceInt64(values ...types.Int64) types.Int64 {
	for _, v := range values {
		if !v.IsNull() && !v.IsUnknown() {
			return v
		}
	}
	// All values are null or unknown, return null
	return types.Int64Null()
}

// coalesceFloat64 returns the first non-null, non-unknown float64 value
func coalesceFloat64(values ...types.Float64) types.Float64 {
	for _, v := range values {
		if !v.IsNull() && !v.IsUnknown() {
			return v
		}
	}
	// All values are null or unknown, return null
	return types.Float64Null()
}
