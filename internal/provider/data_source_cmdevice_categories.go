// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &CMDeviceCategoriesDataSource{}

func NewCMDeviceCategoriesDataSource() datasource.DataSource {
	return &CMDeviceCategoriesDataSource{}
}

// CMDeviceCategoriesDataSource defines the data source implementation.
type CMDeviceCategoriesDataSource struct {
	client *BCMClient
}

// CMDeviceCategoriesDataSourceModel describes the data source data model.
type CMDeviceCategoriesDataSourceModel struct {
	ID         types.String        `tfsdk:"id"`
	Name       types.String        `tfsdk:"name"`
	Categories []CategoryDataModel `tfsdk:"categories"`
}

// CategoryDataModel describes a single category.
type CategoryDataModel struct {
	ID                     types.String `tfsdk:"id"`
	UUID                   types.String `tfsdk:"uuid"`
	Name                   types.String `tfsdk:"name"`
	BaseType               types.String `tfsdk:"base_type"`
	ChildType              types.String `tfsdk:"child_type"`
	SoftwareImageID        types.String `tfsdk:"software_image_id"`
	ManagementNetworkID    types.String `tfsdk:"management_network_id"`
	DiskSetup              types.String `tfsdk:"disksetup"`
	BootLoader             types.String `tfsdk:"boot_loader"`
	BootLoaderProtocol     types.String `tfsdk:"boot_loader_protocol"`
	InstallMode            types.String `tfsdk:"install_mode"`
	KernelVersion          types.String `tfsdk:"kernel_version"`
	KernelParameters       types.String `tfsdk:"kernel_parameters"`
	KernelOutputConsole    types.String `tfsdk:"kernel_output_console"`
	DefaultGateway         types.String `tfsdk:"default_gateway"`
	DefaultGatewayMetric   types.Int64  `tfsdk:"default_gateway_metric"`
	NameServers            types.List   `tfsdk:"name_servers"`
	SearchDomains          types.List   `tfsdk:"search_domains"`
	TimeServers            types.List   `tfsdk:"time_servers"`
	AuthenticationService  types.String `tfsdk:"authentication_service"`
	IOScheduler            types.String `tfsdk:"io_scheduler"`
	FIPS                   types.String `tfsdk:"fips"`
	InteractiveUser        types.String `tfsdk:"interactive_user"`
	AllowNetworkingRestart types.Bool   `tfsdk:"allow_networking_restart"`
	InstallBootRecord      types.Bool   `tfsdk:"install_boot_record"`
	DataNode               types.Bool   `tfsdk:"data_node"`
	VersionConfigFiles     types.Bool   `tfsdk:"version_config_files"`
	Modified               types.Bool   `tfsdk:"modified"`
	ToBeRemoved            types.Bool   `tfsdk:"to_be_removed"`
	Notes                  types.String `tfsdk:"notes"`
	ParentUUID             types.String `tfsdk:"parent_uuid"`
	Modules                types.List   `tfsdk:"modules"`
	FSMounts               types.List   `tfsdk:"fsmounts"`
	Roles                  types.List   `tfsdk:"roles"`
	Services               types.List   `tfsdk:"services"`
}

func (d *CMDeviceCategoriesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cmdevice_categories"
}

func (d *CMDeviceCategoriesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a list of node categories from BCM CMDevice API. Categories define configuration templates for nodes including software images, disk setup, boot configuration, and network settings.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Optional filter to return only categories matching this name exactly",
				Optional:            true,
			},
			"categories": schema.ListNestedAttribute{
				MarkdownDescription: "List of categories matching the filter criteria",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Category identifier (same as UUID)",
							Computed:            true,
						},
						"uuid": schema.StringAttribute{
							MarkdownDescription: "Unique identifier for the category",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Category name",
							Computed:            true,
						},
						"base_type": schema.StringAttribute{
							MarkdownDescription: "Base entity type (always 'Category')",
							Computed:            true,
						},
						"child_type": schema.StringAttribute{
							MarkdownDescription: "Specific category subtype",
							Computed:            true,
						},
						"software_image_id": schema.StringAttribute{
							MarkdownDescription: "UUID of the associated software image",
							Computed:            true,
						},
						"management_network_id": schema.StringAttribute{
							MarkdownDescription: "UUID of the management network",
							Computed:            true,
						},
						"disksetup": schema.StringAttribute{
							MarkdownDescription: "Disk partitioning configuration (XML format)",
							Computed:            true,
						},
						"boot_loader": schema.StringAttribute{
							MarkdownDescription: "Boot loader type (SYSLINUX, GRUB, etc.)",
							Computed:            true,
						},
						"boot_loader_protocol": schema.StringAttribute{
							MarkdownDescription: "Boot loader protocol (HTTP, TFTP, etc.)",
							Computed:            true,
						},
						"install_mode": schema.StringAttribute{
							MarkdownDescription: "Installation mode (AUTO, FULL, etc.)",
							Computed:            true,
						},
						"kernel_version": schema.StringAttribute{
							MarkdownDescription: "Kernel version",
							Computed:            true,
						},
						"kernel_parameters": schema.StringAttribute{
							MarkdownDescription: "Kernel boot parameters",
							Computed:            true,
						},
						"kernel_output_console": schema.StringAttribute{
							MarkdownDescription: "Kernel console output device",
							Computed:            true,
						},
						"default_gateway": schema.StringAttribute{
							MarkdownDescription: "Default gateway IP address",
							Computed:            true,
						},
						"default_gateway_metric": schema.Int64Attribute{
							MarkdownDescription: "Default gateway metric",
							Computed:            true,
						},
						"name_servers": schema.ListAttribute{
							MarkdownDescription: "List of DNS servers",
							Computed:            true,
							ElementType:         types.StringType,
						},
						"search_domains": schema.ListAttribute{
							MarkdownDescription: "List of DNS search domains",
							Computed:            true,
							ElementType:         types.StringType,
						},
						"time_servers": schema.ListAttribute{
							MarkdownDescription: "List of NTP time servers",
							Computed:            true,
							ElementType:         types.StringType,
						},
						"authentication_service": schema.StringAttribute{
							MarkdownDescription: "Authentication service type",
							Computed:            true,
						},
						"io_scheduler": schema.StringAttribute{
							MarkdownDescription: "I/O scheduler",
							Computed:            true,
						},
						"fips": schema.StringAttribute{
							MarkdownDescription: "FIPS mode setting",
							Computed:            true,
						},
						"interactive_user": schema.StringAttribute{
							MarkdownDescription: "Interactive user mode",
							Computed:            true,
						},
						"allow_networking_restart": schema.BoolAttribute{
							MarkdownDescription: "Allow network restart",
							Computed:            true,
						},
						"install_boot_record": schema.BoolAttribute{
							MarkdownDescription: "Install boot record",
							Computed:            true,
						},
						"data_node": schema.BoolAttribute{
							MarkdownDescription: "Is data node",
							Computed:            true,
						},
						"version_config_files": schema.BoolAttribute{
							MarkdownDescription: "Version configuration files",
							Computed:            true,
						},
						"modified": schema.BoolAttribute{
							MarkdownDescription: "Has unsaved changes",
							Computed:            true,
						},
						"to_be_removed": schema.BoolAttribute{
							MarkdownDescription: "Scheduled for removal",
							Computed:            true,
						},
						"notes": schema.StringAttribute{
							MarkdownDescription: "Administrative notes",
							Computed:            true,
						},
						"parent_uuid": schema.StringAttribute{
							MarkdownDescription: "Parent category UUID (for cloned categories)",
							Computed:            true,
						},
						"modules": schema.ListAttribute{
							MarkdownDescription: "Kernel modules",
							Computed:            true,
							ElementType:         types.StringType,
						},
						"fsmounts": schema.ListAttribute{
							MarkdownDescription: "Filesystem mounts",
							Computed:            true,
							ElementType:         types.StringType,
						},
						"roles": schema.ListAttribute{
							MarkdownDescription: "Assigned roles",
							Computed:            true,
							ElementType:         types.StringType,
						},
						"services": schema.ListAttribute{
							MarkdownDescription: "Assigned services",
							Computed:            true,
							ElementType:         types.StringType,
						},
					},
				},
			},
		},
	}
}

func (d *CMDeviceCategoriesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*BCMClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *BCMClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *CMDeviceCategoriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CMDeviceCategoriesDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Call BCM API to get categories
	tflog.Debug(ctx, "Calling cmdevice.getCategories()")
	result, err := d.client.CallJSONRPC(ctx, "cmdevice", "getCategories")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Categories",
			fmt.Sprintf("Could not read categories: %s", err.Error()),
		)
		return
	}

	// Parse JSON response
	var categories []map[string]interface{}
	if err := json.Unmarshal(result, &categories); err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Categories",
			fmt.Sprintf("Could not parse categories response: %s", err.Error()),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Retrieved %d categories from BCM API", len(categories)))

	// Filter by name if specified
	filteredCategories := categories
	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		filterName := data.Name.ValueString()
		filteredCategories = make([]map[string]interface{}, 0)
		for _, cat := range categories {
			if name, ok := cat["name"].(string); ok && name == filterName {
				filteredCategories = append(filteredCategories, cat)
			}
		}
		tflog.Debug(ctx, fmt.Sprintf("Filtered to %d categories matching name '%s'", len(filteredCategories), filterName))
	}

	// Map categories to Terraform state
	data.Categories = make([]CategoryDataModel, 0, len(filteredCategories))
	for _, catData := range filteredCategories {
		uuid := getStringValue(catData, "uuid")
		category := CategoryDataModel{
			ID:                     uuid, // ID is same as UUID for consistency
			UUID:                   uuid,
			Name:                   getStringValue(catData, "name"),
			BaseType:               getStringValue(catData, "baseType"),
			ChildType:              getStringValue(catData, "childType"),
			ManagementNetworkID:    getStringValue(catData, "managementNetwork"),
			DiskSetup:              getStringValue(catData, "disksetup"),
			BootLoader:             getStringValue(catData, "bootLoader"),
			BootLoaderProtocol:     getStringValue(catData, "bootLoaderProtocol"),
			InstallMode:            getStringValue(catData, "installMode"),
			KernelVersion:          getStringValue(catData, "kernelVersion"),
			KernelParameters:       getStringValue(catData, "kernelParameters"),
			KernelOutputConsole:    getStringValue(catData, "kernelOutputConsole"),
			DefaultGateway:         getStringValue(catData, "defaultGateway"),
			DefaultGatewayMetric:   getInt64Value(catData, "defaultGatewayMetric"),
			AuthenticationService:  getStringValue(catData, "authenticationService"),
			IOScheduler:            getStringValue(catData, "ioScheduler"),
			FIPS:                   getStringValue(catData, "fips"),
			InteractiveUser:        getStringValue(catData, "interactiveUser"),
			AllowNetworkingRestart: getBoolValue(catData, "allowNetworkingRestart"),
			InstallBootRecord:      getBoolValue(catData, "installBootRecord"),
			DataNode:               getBoolValue(catData, "dataNode"),
			VersionConfigFiles:     getBoolValue(catData, "versionConfigFiles"),
			Modified:               getBoolValue(catData, "modified"),
			ToBeRemoved:            getBoolValue(catData, "to_be_removed"),
			Notes:                  getStringValue(catData, "notes"),
			ParentUUID:             getStringValue(catData, "parent_uuid"),
		}

		// Extract software image ID from softwareImageProxy
		if proxy, ok := catData["softwareImageProxy"].(map[string]interface{}); ok {
			category.SoftwareImageID = getStringValue(proxy, "parentSoftwareImage")
		} else {
			category.SoftwareImageID = types.StringNull()
		}

		// Handle list attributes - for now, create empty lists
		// Will implement proper nested object mapping in refactor phase
		category.NameServers, _ = types.ListValue(types.StringType, []attr.Value{})
		category.SearchDomains, _ = types.ListValue(types.StringType, []attr.Value{})
		category.TimeServers, _ = types.ListValue(types.StringType, []attr.Value{})
		category.Modules, _ = types.ListValue(types.StringType, []attr.Value{})
		category.FSMounts, _ = types.ListValue(types.StringType, []attr.Value{})
		category.Roles, _ = types.ListValue(types.StringType, []attr.Value{})
		category.Services, _ = types.ListValue(types.StringType, []attr.Value{})

		data.Categories = append(data.Categories, category)
	}

	tflog.Trace(ctx, "Mapped categories to Terraform state", map[string]interface{}{
		"count": len(data.Categories),
	})

	// Set data source ID (use timestamp as unique identifier)
	data.ID = types.StringValue("categories")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
