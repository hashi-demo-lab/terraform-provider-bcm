// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &CMPartSoftwareImagesDataSource{}
	_ datasource.DataSourceWithConfigure = &CMPartSoftwareImagesDataSource{}
)

// NewCMPartSoftwareImagesDataSource is a helper function to simplify the provider implementation.
func NewCMPartSoftwareImagesDataSource() datasource.DataSource {
	return &CMPartSoftwareImagesDataSource{}
}

// CMPartSoftwareImagesDataSource is the data source implementation.
type CMPartSoftwareImagesDataSource struct {
	client *BCMClient
}

// CMPartSoftwareImagesDataSourceModel describes the data source data model.
type CMPartSoftwareImagesDataSourceModel struct {
	ID     types.String         `tfsdk:"id"`
	Images []SoftwareImageModel `tfsdk:"images"`
}

// SoftwareImageModel represents a BCM software image with all fields
type SoftwareImageModel struct {
	// Identity fields
	ID   types.String `tfsdk:"id"`
	UUID types.String `tfsdk:"uuid"`
	Name types.String `tfsdk:"name"`
	Path types.String `tfsdk:"path"`

	// Kernel configuration
	KernelVersion       types.String `tfsdk:"kernel_version"`
	KernelParameters    types.String `tfsdk:"kernel_parameters"`
	KernelOutputConsole types.String `tfsdk:"kernel_output_console"`

	// Partitions
	BootfsPart types.String `tfsdk:"bootfs_part"`
	FsPart     types.String `tfsdk:"fs_part"`

	// Serial Over LAN
	EnableSOL      types.Bool   `tfsdk:"enable_sol"`
	SOLPort        types.String `tfsdk:"sol_port"`
	SOLSpeed       types.String `tfsdk:"sol_speed"`
	SOLFlowControl types.Bool   `tfsdk:"sol_flow_control"`

	// Metadata
	BaseType     types.String `tfsdk:"base_type"`
	ChildType    types.String `tfsdk:"child_type"`
	CreationTime types.Int64  `tfsdk:"creation_time"`
	Revision     types.String `tfsdk:"revision"`
	RevisionID   types.Int64  `tfsdk:"revision_id"`

	// Relationships
	OriginalImage       types.String `tfsdk:"original_image"`
	ParentSoftwareImage types.String `tfsdk:"parent_software_image"`
	ParentUUID          types.String `tfsdk:"parent_uuid"`

	// State flags
	FileOperationInProgress types.Bool `tfsdk:"file_operation_in_progress"`
	Modified                types.Bool `tfsdk:"modified"`
	ToBeRemoved             types.Bool `tfsdk:"to_be_removed"`

	// Notes
	Notes types.String `tfsdk:"notes"`

	// Nested modules
	Modules []KernelModuleModel `tfsdk:"modules"`
}

// KernelModuleModel represents a kernel module configured for a software image
type KernelModuleModel struct {
	UUID        types.String `tfsdk:"uuid"`
	Name        types.String `tfsdk:"name"`
	Parameters  types.String `tfsdk:"parameters"`
	BaseType    types.String `tfsdk:"base_type"`
	ChildType   types.String `tfsdk:"child_type"`
	Revision    types.String `tfsdk:"revision"`
	Modified    types.Bool   `tfsdk:"modified"`
	ToBeRemoved types.Bool   `tfsdk:"to_be_removed"`
}

// Metadata returns the data source type name.
func (d *CMPartSoftwareImagesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cmpart_softwareimages"
}

// Schema defines the schema for the data source.
func (d *CMPartSoftwareImagesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches software images from BCM CMPart service. Software images are operating system images used to provision DPU nodes with kernel configuration and modules.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Placeholder identifier for this data source",
			},
			"images": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of software images retrieved from BCM",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Software image identifier (same as uuid)",
						},
						"uuid": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Software image UUID",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Software image name",
						},
						"path": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "File system path to image on BCM server",
						},
						"kernel_version": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Linux kernel version (e.g., 5.15.0-58-generic)",
						},
						"kernel_parameters": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Kernel boot parameters",
						},
						"kernel_output_console": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Kernel output console configuration",
						},
						"bootfs_part": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Boot filesystem partition",
						},
						"fs_part": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Root filesystem partition",
						},
						"enable_sol": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Serial Over LAN enabled flag",
						},
						"sol_port": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "SOL serial port (e.g., ttyS0)",
						},
						"sol_speed": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "SOL baud rate (e.g., 115200)",
						},
						"sol_flow_control": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "SOL hardware flow control enabled",
						},
						"base_type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Base operating system type (e.g., Linux)",
						},
						"child_type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "OS distribution type (e.g., Ubuntu)",
						},
						"creation_time": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Unix timestamp (milliseconds) when image was created",
						},
						"revision": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Image revision string",
						},
						"revision_id": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Numeric revision identifier",
						},
						"original_image": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Name of original base image",
						},
						"parent_software_image": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Name of parent image (for derived images)",
						},
						"parent_uuid": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "UUID of parent image",
						},
						"file_operation_in_progress": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "File operation currently executing on image",
						},
						"modified": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Image has been modified from original",
						},
						"to_be_removed": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Image scheduled for deletion",
						},
						"notes": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Free-form notes about the image",
						},
						"modules": schema.ListNestedAttribute{
							Computed:            true,
							MarkdownDescription: "Kernel modules configured for this image",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"uuid": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Module UUID",
									},
									"name": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Module name (e.g., nvidia-drm)",
									},
									"parameters": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Module load parameters",
									},
									"base_type": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Module category",
									},
									"child_type": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Module subcategory",
									},
									"revision": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Module version",
									},
									"modified": schema.BoolAttribute{
										Computed:            true,
										MarkdownDescription: "Module configuration has been modified",
									},
									"to_be_removed": schema.BoolAttribute{
										Computed:            true,
										MarkdownDescription: "Module scheduled for removal",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *CMPartSoftwareImagesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read refreshes the Terraform state with the latest data.
func (d *CMPartSoftwareImagesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state CMPartSoftwareImagesDataSourceModel

	// Call BCM API
	body, err := d.client.CallJSONRPC(ctx, "CMPart", "getSoftwareImages")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read BCM Software Images",
			"An unexpected error occurred when calling the BCM API. "+
				"Check the endpoint, credentials, and network connectivity.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	// Parse JSON response
	var apiResponse []map[string]interface{}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Parse BCM API Response",
			"The provider received an unexpected response format from the BCM API. "+
				"This may indicate an API compatibility issue.\n\n"+
				"Parse Error: "+err.Error()+"\n"+
				"Response: "+limitString(string(body), 500),
		)
		return
	}

	// Map API response to Terraform models
	state.Images = make([]SoftwareImageModel, 0, len(apiResponse))
	for _, imgData := range apiResponse {
		state.Images = append(state.Images, mapAPIResponseToModel(imgData))
	}

	// Set placeholder ID
	state.ID = types.StringValue("placeholder")

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

	tflog.Debug(ctx, "Successfully read software images", map[string]interface{}{
		"count": len(state.Images),
	})
}

// mapAPIResponseToModel converts API JSON to SoftwareImageModel with null-safe field extraction
func mapAPIResponseToModel(apiData map[string]interface{}) SoftwareImageModel {
	model := SoftwareImageModel{}

	// Identity fields
	model.UUID = getStringValue(apiData, "uuid")
	model.ID = model.UUID // Use UUID as ID
	model.Name = getStringValue(apiData, "name")
	model.Path = getStringValue(apiData, "path")

	// Kernel configuration
	model.KernelVersion = getStringValue(apiData, "kernelVersion")
	model.KernelParameters = getStringValue(apiData, "kernelParameters")
	model.KernelOutputConsole = getStringValue(apiData, "kernelOutputConsole")

	// Partitions
	model.BootfsPart = getStringValue(apiData, "bootfspart")
	model.FsPart = getStringValue(apiData, "fspart")

	// Serial Over LAN
	model.EnableSOL = getBoolValue(apiData, "enableSOL")
	model.SOLPort = getStringValue(apiData, "SOLPort")
	model.SOLSpeed = getStringValue(apiData, "SOLSpeed")
	model.SOLFlowControl = getBoolValue(apiData, "SOLFlowControl")

	// Metadata
	model.BaseType = getStringValue(apiData, "baseType")
	model.ChildType = getStringValue(apiData, "childType")
	model.CreationTime = getInt64Value(apiData, "creationTime")
	model.Revision = getStringValue(apiData, "revision")
	model.RevisionID = getInt64Value(apiData, "revisionID")

	// Relationships
	model.OriginalImage = getStringValue(apiData, "originalImage")
	model.ParentSoftwareImage = getStringValue(apiData, "parentSoftwareImage")
	model.ParentUUID = getStringValue(apiData, "parent_uuid")

	// State flags
	model.FileOperationInProgress = getBoolValue(apiData, "fileOperationInProgress")
	model.Modified = getBoolValue(apiData, "modified")
	model.ToBeRemoved = getBoolValue(apiData, "to_be_removed")

	// Notes
	model.Notes = getStringValue(apiData, "notes")

	// Nested modules array
	if modulesData, ok := apiData["modules"].([]interface{}); ok {
		model.Modules = make([]KernelModuleModel, 0, len(modulesData))
		for _, modData := range modulesData {
			if modMap, ok := modData.(map[string]interface{}); ok {
				module := KernelModuleModel{
					UUID:        getStringValue(modMap, "uuid"),
					Name:        getStringValue(modMap, "name"),
					Parameters:  getStringValue(modMap, "parameters"),
					BaseType:    getStringValue(modMap, "baseType"),
					ChildType:   getStringValue(modMap, "childType"),
					Revision:    getStringValue(modMap, "revision"),
					Modified:    getBoolValue(modMap, "modified"),
					ToBeRemoved: getBoolValue(modMap, "to_be_removed"),
				}
				model.Modules = append(model.Modules, module)
			}
		}
	} else {
		model.Modules = []KernelModuleModel{} // Empty slice if modules null/missing
	}

	return model
}

// Helper functions for null-safe field extraction

func getStringValue(data map[string]interface{}, key string) types.String {
	if val, ok := data[key]; ok && val != nil {
		if str, ok := val.(string); ok && str != "" {
			return types.StringValue(str)
		}
	}
	return types.StringNull()
}

func getBoolValue(data map[string]interface{}, key string) types.Bool {
	if val, ok := data[key]; ok && val != nil {
		if b, ok := val.(bool); ok {
			return types.BoolValue(b)
		}
	}
	return types.BoolNull()
}

func getInt64Value(data map[string]interface{}, key string) types.Int64 {
	if val, ok := data[key]; ok && val != nil {
		switch v := val.(type) {
		case float64:
			return types.Int64Value(int64(v))
		case int64:
			return types.Int64Value(v)
		case int:
			return types.Int64Value(int64(v))
		}
	}
	return types.Int64Null()
}
