// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &CMPartPartitionResource{}
	_ resource.ResourceWithImportState = &CMPartPartitionResource{}
)

// CMPartPartitionResource defines the resource implementation.
type CMPartPartitionResource struct {
	client *BCMClient
}

// PartitionResourceModel describes the resource data model.
type PartitionResourceModel struct {
	// Identity fields
	ID        types.String `tfsdk:"id"`
	UUID      types.String `tfsdk:"uuid"`
	Name      types.String `tfsdk:"name"`
	BaseType  types.String `tfsdk:"base_type"`
	ChildType types.String `tfsdk:"child_type"`

	// Configuration fields
	ClusterName types.String `tfsdk:"cluster_name"`
	SlaveName   types.String `tfsdk:"slave_name"`
	SlaveDigits types.Int64  `tfsdk:"slave_digits"`
	RelayHost   types.String `tfsdk:"relay_host"`
	NoZeroConf  types.Bool   `tfsdk:"no_zero_conf"`

	// Network configuration (list attributes)
	AdminEmail    types.List `tfsdk:"admin_email"`
	TimeServers   types.List `tfsdk:"time_servers"`
	SearchDomains types.List `tfsdk:"search_domains"`
	NameServers   types.List `tfsdk:"name_servers"`

	// Metadata fields
	CreationTime types.Int64  `tfsdk:"creation_time"`
	Revision     types.String `tfsdk:"revision"`
	Modified     types.Bool   `tfsdk:"modified"`
	ToBeRemoved  types.Bool   `tfsdk:"to_be_removed"`
	Notes        types.String `tfsdk:"notes"`
}

// NewCMPartPartitionResource creates a new resource instance.
func NewCMPartPartitionResource() resource.Resource {
	return &CMPartPartitionResource{}
}

// Metadata returns the resource type name.
func (r *CMPartPartitionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cmpart_partition"
}

// Schema defines the resource schema.
func (r *CMPartPartitionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a BCM cluster partition (organizational unit).\n\n" +
			"Partitions define logical groupings of nodes within a BCM cluster, including cluster-wide " +
			"configuration such as naming conventions, network settings, and administrative contacts.",

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
				MarkdownDescription: "Partition name (must be unique within BCM cluster)",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"base_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "BCM entity base type (always 'Partition')",
			},
			"child_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "BCM entity polymorphic child type",
			},
			"cluster_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Cluster display name",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"slave_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Node naming prefix (default: 'node')",
			},
			"slave_digits": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Node numbering zero-padding digits (default: 3, range: 1-10)",
				Validators: []validator.Int64{
					int64validator.Between(1, 10),
				},
			},
			"relay_host": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "SMTP relay hostname for email notifications",
			},
			"no_zero_conf": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Disable Zeroconf networking (default: false)",
			},
			"admin_email": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Administrative contact email addresses",
			},
			"time_servers": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "NTP time servers for clock synchronization",
			},
			"search_domains": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "DNS search domains",
			},
			"name_servers": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "DNS resolver servers",
			},
			"creation_time": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Unix timestamp of partition creation",
			},
			"revision": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Concurrency control version string",
			},
			"modified": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "BCM dirty flag indicating unsaved changes",
			},
			"to_be_removed": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "BCM deletion pending flag",
			},
			"notes": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Partition description or notes",
			},
		},
	}
}

// Configure receives provider configuration.
func (r *CMPartPartitionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if provider has not been configured
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

// Create creates a new partition via BCM API.
func (r *CMPartPartitionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PartitionResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating partition", map[string]interface{}{
		"name":         plan.Name.ValueString(),
		"cluster_name": plan.ClusterName.ValueString(),
	})

	// Build BCM API entity
	entity := r.buildAPIEntity(&plan)

	// Call BCM API to create partition
	body, err := r.client.CallJSONRPC(ctx, "cmpart", "addPartition", entity, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Partition",
			fmt.Sprintf("Could not create partition '%s': %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	// Parse response to extract UUID
	var partitionData map[string]interface{}
	if err := json.Unmarshal(body, &partitionData); err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Partition Response",
			fmt.Sprintf("Could not parse partition creation response: %s", err.Error()),
		)
		return
	}

	// Extract UUID
	uuid, ok := partitionData["uuid"].(string)
	if !ok || uuid == "" {
		resp.Diagnostics.AddError(
			"Error Extracting UUID",
			"Partition creation response did not contain a valid UUID",
		)
		return
	}

	plan.UUID = types.StringValue(uuid)
	plan.ID = types.StringValue(uuid)

	// Fetch full partition data to populate computed fields
	r.readPartition(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Partition created successfully", map[string]interface{}{
		"uuid": uuid,
		"name": plan.Name.ValueString(),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read fetches partition state from BCM API (drift detection).
func (r *CMPartPartitionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PartitionResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	uuid := state.UUID.ValueString()
	tflog.Debug(ctx, "Reading partition", map[string]interface{}{"uuid": uuid})

	// Fetch partition data from BCM API
	r.readPartition(ctx, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		// Check if partition was deleted externally
		if len(resp.Diagnostics.Errors()) > 0 {
			for _, diag := range resp.Diagnostics.Errors() {
				if diag.Summary() == "Partition Not Found" {
					// Resource was deleted outside Terraform - remove from state
					resp.State.RemoveResource(ctx)
					return
				}
			}
		}
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update modifies partition configuration via BCM API.
func (r *CMPartPartitionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PartitionResourceModel
	var state PartitionResourceModel

	// Read Terraform plan and current state
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve UUID from state
	plan.UUID = state.UUID
	plan.ID = state.ID

	uuid := plan.UUID.ValueString()
	tflog.Debug(ctx, "Updating partition", map[string]interface{}{
		"uuid": uuid,
		"name": plan.Name.ValueString(),
	})

	// Build updated BCM API entity
	entity := r.buildAPIEntity(&plan)

	// Call BCM API to update partition
	_, err := r.client.CallJSONRPC(ctx, "cmpart", "updatePartition", entity, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Partition",
			fmt.Sprintf("Could not update partition '%s' (uuid: %s): %s",
				plan.Name.ValueString(), uuid, err.Error()),
		)
		return
	}

	// Fetch updated partition data
	r.readPartition(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Partition updated successfully", map[string]interface{}{"uuid": uuid})

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes partition from BCM cluster.
func (r *CMPartPartitionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PartitionResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	uuid := state.UUID.ValueString()
	tflog.Debug(ctx, "Deleting partition", map[string]interface{}{"uuid": uuid})

	// Call BCM API to remove partition
	_, err := r.client.CallJSONRPC(ctx, "cmpart", "removePartition", uuid, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Partition",
			fmt.Sprintf("Could not delete partition '%s' (uuid: %s): %s\n\n"+
				"If this partition has active nodes assigned, remove them first.",
				state.Name.ValueString(), uuid, err.Error()),
		)
		return
	}

	tflog.Debug(ctx, "Partition deleted successfully", map[string]interface{}{"uuid": uuid})
}

// ImportState imports existing partition by UUID.
func (r *CMPartPartitionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Use UUID as the import identifier
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)

	// Set ID to match UUID
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// readPartition fetches partition data from BCM API and populates model.
func (r *CMPartPartitionResource) readPartition(ctx context.Context, model *PartitionResourceModel, diags *diag.Diagnostics) {
	uuid := model.UUID.ValueString()

	// Fetch partition from BCM API using direct UUID lookup
	body, err := r.client.CallJSONRPC(ctx, "cmpart", "getPartition", uuid)
	if err != nil {
		diags.AddError(
			"Partition Not Found",
			fmt.Sprintf("Could not fetch partition (uuid: %s): %s", uuid, err.Error()),
		)
		return
	}

	// Parse response
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		diags.AddError(
			"Error Parsing Partition Data",
			fmt.Sprintf("Could not parse partition response: %s", err.Error()),
		)
		return
	}

	// Map BCM API fields (camelCase) to Terraform model (snake_case)
	model.UUID = getStringValue(data, "uuid")
	model.ID = model.UUID
	model.Name = getStringValue(data, "name")
	model.BaseType = getStringValue(data, "baseType")
	model.ChildType = getStringValue(data, "childType")

	// Configuration fields
	model.ClusterName = getStringValue(data, "clusterName")
	model.SlaveName = getStringValue(data, "slaveName")
	model.SlaveDigits = getInt64Value(data, "slaveDigits")
	model.RelayHost = getStringValue(data, "relayHost")
	model.NoZeroConf = getBoolValue(data, "noZeroConf")

	// List attributes - unmarshal JSON arrays to types.List
	model.AdminEmail = unmarshalListAttribute(data["adminEmail"])
	model.TimeServers = unmarshalListAttribute(data["timeServers"])
	model.SearchDomains = unmarshalListAttribute(data["searchDomains"])
	model.NameServers = unmarshalListAttribute(data["nameServers"])

	// Metadata fields
	model.CreationTime = getInt64Value(data, "creationTime")
	model.Revision = getStringValue(data, "revision")
	model.Modified = getBoolValue(data, "modified")
	model.ToBeRemoved = getBoolValue(data, "to_be_removed")
	model.Notes = getStringValue(data, "notes")
}

// buildAPIEntity constructs BCM entity structure from Terraform model.
func (r *CMPartPartitionResource) buildAPIEntity(model *PartitionResourceModel) map[string]interface{} {
	entity := map[string]interface{}{
		"baseType":      "Partition",
		"childType":     "",
		"modified":      true,
		"to_be_removed": false,
	}

	// Add UUID and revision if updating (preserve for optimistic locking)
	if !model.UUID.IsNull() && !model.UUID.IsUnknown() {
		entity["uuid"] = model.UUID.ValueString()
	}
	if !model.Revision.IsNull() && !model.Revision.IsUnknown() {
		entity["revision"] = model.Revision.ValueString()
	}

	// Identity fields
	if !model.Name.IsNull() && !model.Name.IsUnknown() {
		entity["name"] = model.Name.ValueString()
	}

	// Configuration fields (snake_case → camelCase)
	if !model.ClusterName.IsNull() && !model.ClusterName.IsUnknown() {
		entity["clusterName"] = model.ClusterName.ValueString()
	}
	if !model.SlaveName.IsNull() && !model.SlaveName.IsUnknown() {
		entity["slaveName"] = model.SlaveName.ValueString()
	}
	if !model.SlaveDigits.IsNull() && !model.SlaveDigits.IsUnknown() {
		entity["slaveDigits"] = model.SlaveDigits.ValueInt64()
	}
	if !model.RelayHost.IsNull() && !model.RelayHost.IsUnknown() {
		entity["relayHost"] = model.RelayHost.ValueString()
	}
	if !model.NoZeroConf.IsNull() && !model.NoZeroConf.IsUnknown() {
		entity["noZeroConf"] = model.NoZeroConf.ValueBool()
	}

	// List attributes - convert types.List to []string
	if !model.AdminEmail.IsNull() && !model.AdminEmail.IsUnknown() {
		entity["adminEmail"] = mapListAttribute(model.AdminEmail)
	}
	if !model.TimeServers.IsNull() && !model.TimeServers.IsUnknown() {
		entity["timeServers"] = mapListAttribute(model.TimeServers)
	}
	if !model.SearchDomains.IsNull() && !model.SearchDomains.IsUnknown() {
		entity["searchDomains"] = mapListAttribute(model.SearchDomains)
	}
	if !model.NameServers.IsNull() && !model.NameServers.IsUnknown() {
		entity["nameServers"] = mapListAttribute(model.NameServers)
	}

	// Notes field
	if !model.Notes.IsNull() && !model.Notes.IsUnknown() {
		entity["notes"] = model.Notes.ValueString()
	}

	return entity
}

// mapListAttribute converts types.List to []string for BCM API.
func mapListAttribute(list types.List) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	elements := list.Elements()
	result := make([]string, 0, len(elements))

	for _, elem := range elements {
		if strVal, ok := elem.(types.String); ok && !strVal.IsNull() {
			result = append(result, strVal.ValueString())
		}
	}

	return result
}

// unmarshalListAttribute converts BCM API array to types.List.
func unmarshalListAttribute(data interface{}) types.List {
	if data == nil {
		return types.ListNull(types.StringType)
	}

	// Type assert to []interface{}
	arr, ok := data.([]interface{})
	if !ok {
		return types.ListNull(types.StringType)
	}

	if len(arr) == 0 {
		// Return empty list, not null
		listVal, _ := types.ListValue(types.StringType, []attr.Value{})
		return listVal
	}

	// Convert each element to types.String
	elements := make([]attr.Value, 0, len(arr))
	for _, item := range arr {
		if str, ok := item.(string); ok {
			elements = append(elements, types.StringValue(str))
		}
	}

	listVal, _ := types.ListValue(types.StringType, elements)
	return listVal
}

// Helper functions getStringValue, getBoolValue, and getInt64Value are defined in
// data_source_cmpart_softwareimages.go and are reused here for consistency.
