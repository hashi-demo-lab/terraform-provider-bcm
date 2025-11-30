// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"strings"

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
	BCMDataSourceBase
}

// CMPartSoftwareImagesDataSourceModel describes the data source data model.
type CMPartSoftwareImagesDataSourceModel struct {
	ID     types.String              `tfsdk:"id"`
	Filter *SoftwareImageFilterModel `tfsdk:"filter"`
	Images []SoftwareImageModel      `tfsdk:"images"`
}

// SoftwareImageFilterModel describes the filter block for client-side filtering.
// Multiple filters use AND logic (all filters must match for an image to be included).
type SoftwareImageFilterModel struct {
	NamePattern types.String `tfsdk:"name_pattern"` // Case-insensitive substring match for image name
	Category    types.String `tfsdk:"category"`     // Exact match for child_type (category)
}

// SoftwareImageModel represents a BCM software image with all fields.
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

// KernelModuleModel represents a kernel module configured for a software image.
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
		Blocks: map[string]schema.Block{
			"filter": schema.SingleNestedBlock{
				MarkdownDescription: "Optional filters to narrow down the list of software images.",
				Attributes: map[string]schema.Attribute{
					"name_pattern": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Case-insensitive substring match for image name.",
					},
					"category": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Exact match for image category (child_type).",
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *CMPartSoftwareImagesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.ConfigureDataSource(req, resp)
}

// Read refreshes the Terraform state with the latest data.
func (d *CMPartSoftwareImagesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config CMPartSoftwareImagesDataSourceModel

	// Read configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading BCM software images from API", map[string]interface{}{
		"service": "CMPart",
		"call":    "getSoftwareImages",
	})

	// Call BCM API
	body, err := d.Client.CallJSONRPC(ctx, "CMPart", "getSoftwareImages")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read BCM Software Images",
			"An unexpected error occurred when calling the BCM API. "+
				"Check the endpoint, credentials, and network connectivity.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "Successfully retrieved BCM API response", map[string]interface{}{
		"response_size": len(body),
	})

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

	tflog.Debug(ctx, "Parsed BCM API response", map[string]interface{}{
		"total_images": len(apiResponse),
	})

	// Map and filter images
	state := CMPartSoftwareImagesDataSourceModel{
		ID:     types.StringValue("placeholder"),
		Filter: config.Filter,
		Images: []SoftwareImageModel{},
	}

	// Log filter criteria if present
	if config.Filter != nil {
		filterCriteria := map[string]interface{}{}
		if !config.Filter.NamePattern.IsNull() {
			filterCriteria["name_pattern"] = config.Filter.NamePattern.ValueString()
		}
		if !config.Filter.Category.IsNull() {
			filterCriteria["category"] = config.Filter.Category.ValueString()
		}
		tflog.Debug(ctx, "Applying client-side filters", filterCriteria)
	}

	// Map and filter images
	for _, imgData := range apiResponse {
		image := mapAPIResponseToModel(imgData)
		if matchesSoftwareImageFilter(image, config.Filter) {
			state.Images = append(state.Images, image)
		}
	}

	tflog.Info(ctx, "Software images data source read complete", map[string]interface{}{
		"total_images":    len(apiResponse),
		"filtered_images": len(state.Images),
	})

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// mapAPIResponseToModel converts API JSON to SoftwareImageModel with null-safe field extraction.
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

// matchesSoftwareImageFilter checks if an image matches the filter criteria.
// Multiple filters use AND logic - the image must match all specified filters.
//
// Filter behavior:
//   - name_pattern: Case-insensitive substring matching (e.g., "ubuntu" matches "Ubuntu-20.04", "ubuntu-base")
//   - category: Exact case-sensitive matching for child_type (e.g., "Ubuntu" matches only images with child_type="Ubuntu")
//   - Omitted filters are ignored (do not restrict results)
//
// Returns true if the image matches all specified filters, false otherwise.
func matchesSoftwareImageFilter(image SoftwareImageModel, filter *SoftwareImageFilterModel) bool {
	// No filter means all images match
	if filter == nil {
		return true
	}

	// Name pattern filter: case-insensitive substring matching
	if !filter.NamePattern.IsNull() && !filter.NamePattern.IsUnknown() {
		pattern := strings.ToLower(filter.NamePattern.ValueString())
		name := strings.ToLower(image.Name.ValueString())
		// Image name must contain the pattern (case-insensitive)
		if !strings.Contains(name, pattern) {
			return false
		}
	}

	// Category filter: exact case-sensitive matching for child_type
	if !filter.Category.IsNull() && !filter.Category.IsUnknown() {
		// Image child_type must exactly match the filter category
		if filter.Category.ValueString() != image.ChildType.ValueString() {
			return false
		}
	}

	// Image matches all specified filters
	return true
}
