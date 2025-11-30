// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &CMPartPartitionsDataSource{}
	_ datasource.DataSourceWithConfigure = &CMPartPartitionsDataSource{}
)

// Constants for BCM API calls.
const (
	bcmPartitionService    = "cmpart"
	bcmPartitionMethod     = "getPartitions"
	partitionIDPlaceholder = "placeholder"
)

// NewCMPartPartitionsDataSource is a helper function to simplify the provider implementation.
func NewCMPartPartitionsDataSource() datasource.DataSource {
	return &CMPartPartitionsDataSource{}
}

// CMPartPartitionsDataSource is the data source implementation.
type CMPartPartitionsDataSource struct {
	BCMDataSourceBase
}

// CMPartPartitionsDataSourceModel describes the data source data model.
type CMPartPartitionsDataSourceModel struct {
	ID         types.String          `tfsdk:"id"`
	Filter     *PartitionFilterModel `tfsdk:"filter"`
	Partitions []PartitionModel      `tfsdk:"partitions"`
}

// PartitionFilterModel describes the filter block for client-side filtering.
type PartitionFilterModel struct {
	NamePattern types.String `tfsdk:"name_pattern"`
}

// PartitionModel represents a BCM partition with all fields.
type PartitionModel struct {
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

	// Email and network configuration (arrays)
	AdminEmail    types.List `tfsdk:"admin_email"`    // List of types.String
	TimeServers   types.List `tfsdk:"time_servers"`   // List of types.String
	SearchDomains types.List `tfsdk:"search_domains"` // List of types.String
	NameServers   types.List `tfsdk:"name_servers"`   // List of types.String

	// Metadata and state fields
	CreationTime types.Int64  `tfsdk:"creation_time"`
	Revision     types.String `tfsdk:"revision"`
	Modified     types.Bool   `tfsdk:"modified"`
	ToBeRemoved  types.Bool   `tfsdk:"to_be_removed"`
	Notes        types.String `tfsdk:"notes"`
}

// Metadata returns the data source type name.
func (d *CMPartPartitionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cmpart_partitions"
}

// Schema defines the schema for the data source.
func (d *CMPartPartitionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches partition information from BCM CMPart service. Partitions define filesystem configurations referenced by software images via bootfs_part and fs_part UUID fields.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Placeholder identifier for this data source",
			},
			"partitions": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of partitions retrieved from BCM cluster",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Partition identifier (same as uuid)",
						},
						"uuid": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Partition UUID (primary identifier)",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Human-readable partition name",
						},
						"base_type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "BCM entity base type (typically Partition)",
						},
						"child_type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "BCM polymorphic type discriminator",
						},
						"cluster_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Display name for the cluster",
						},
						"slave_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Node naming prefix (e.g., node for node001, node002)",
						},
						"slave_digits": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Number of digits in node numbering (e.g., 3 for node001)",
						},
						"relay_host": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "SMTP relay hostname for email notifications",
						},
						"no_zero_conf": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Disable Zeroconf networking",
						},
						"admin_email": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "Administrator email addresses",
						},
						"time_servers": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "NTP time server addresses",
						},
						"search_domains": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "DNS search domains",
						},
						"name_servers": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "DNS name server addresses",
						},
						"creation_time": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Unix timestamp of partition creation",
						},
						"revision": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "BCM revision tracking field",
						},
						"modified": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether partition has uncommitted modifications",
						},
						"to_be_removed": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether partition is marked for deletion",
						},
						"notes": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "User notes or description",
						},
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"filter": schema.SingleNestedBlock{
				MarkdownDescription: "Client-side filtering criteria for partitions",
				Attributes: map[string]schema.Attribute{
					"name_pattern": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Case-insensitive substring match for partition name (e.g., base matches base-partition)",
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *CMPartPartitionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.ConfigureDataSource(req, resp)
}

// Read refreshes the Terraform state with the latest data.
func (d *CMPartPartitionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CMPartPartitionsDataSourceModel

	// Read configuration
	tflog.Debug(ctx, "Reading partition data source configuration")
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Log filter configuration if present
	if data.Filter != nil && !data.Filter.NamePattern.IsNull() {
		tflog.Debug(ctx, "Applying name pattern filter", map[string]interface{}{
			"pattern": data.Filter.NamePattern.ValueString(),
		})
	}

	tflog.Debug(ctx, "Calling BCM API for partitions", map[string]interface{}{
		"service": bcmPartitionService,
		"method":  bcmPartitionMethod,
	})

	// Call BCM API
	body, err := d.Client.CallJSONRPC(ctx, bcmPartitionService, bcmPartitionMethod)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read BCM Partitions",
			fmt.Sprintf("Failed to call BCM API (service: %s, method: %s): %v",
				bcmPartitionService, bcmPartitionMethod, err),
		)
		return
	}

	// Parse JSON response
	var partitionsData []map[string]interface{}
	if err := json.Unmarshal(body, &partitionsData); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Parse BCM Partitions Response",
			fmt.Sprintf("Failed to unmarshal JSON response from BCM API (service: %s, method: %s): %v",
				bcmPartitionService, bcmPartitionMethod, err),
		)
		return
	}

	tflog.Debug(ctx, "Successfully retrieved partitions from API", map[string]interface{}{
		"count": len(partitionsData),
	})

	// Apply client-side filtering
	filteredPartitions := partitionsData
	if data.Filter != nil {
		filteredPartitions = applyPartitionFilters(partitionsData, data.Filter)
		tflog.Debug(ctx, "Applied client-side filters", map[string]interface{}{
			"original_count": len(partitionsData),
			"filtered_count": len(filteredPartitions),
		})
	}

	// Map to Terraform state
	data.Partitions = make([]PartitionModel, 0, len(filteredPartitions))
	for _, partitionData := range filteredPartitions {
		partition := mapPartitionAPIResponseToModel(ctx, partitionData)
		data.Partitions = append(data.Partitions, partition)
	}

	// Set computed ID
	data.ID = types.StringValue(partitionIDPlaceholder)

	tflog.Debug(ctx, "Returning partitions to Terraform state", map[string]interface{}{
		"partition_count": len(data.Partitions),
	})

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// mapPartitionAPIResponseToModel converts API JSON to PartitionModel with null-safe field extraction.
func mapPartitionAPIResponseToModel(ctx context.Context, apiData map[string]interface{}) PartitionModel {
	model := PartitionModel{}

	// Identity fields - uuid → uuid, uuid → id
	model.UUID = getStringValue(apiData, "uuid")
	model.ID = model.UUID // Use UUID as ID
	model.Name = getStringValue(apiData, "name")
	model.BaseType = getStringValue(apiData, "baseType")
	model.ChildType = getStringValue(apiData, "childType")

	// Configuration fields - camelCase → snake_case mapping
	model.ClusterName = getStringValue(apiData, "clusterName") // clusterName → cluster_name
	model.SlaveName = getStringValue(apiData, "slaveName")     // slaveName → slave_name
	model.SlaveDigits = getInt64Value(apiData, "slaveDigits")  // slaveDigits → slave_digits
	model.RelayHost = getStringValue(apiData, "relayHost")     // relayHost → relay_host
	model.NoZeroConf = getBoolValue(apiData, "noZeroConf")     // noZeroConf → no_zero_conf

	// Email and network configuration arrays
	model.AdminEmail = getStringListValue(ctx, apiData, "adminEmail")       // adminEmail → admin_email
	model.TimeServers = getStringListValue(ctx, apiData, "timeServers")     // timeServers → time_servers
	model.SearchDomains = getStringListValue(ctx, apiData, "searchDomains") // searchDomains → search_domains
	model.NameServers = getStringListValue(ctx, apiData, "nameServers")     // nameServers → name_servers

	// Metadata and state fields
	model.CreationTime = getInt64Value(apiData, "creationTime") // creationTime → creation_time
	model.Revision = getStringValue(apiData, "revision")
	model.Modified = getBoolValue(apiData, "modified")
	model.ToBeRemoved = getBoolValue(apiData, "to_be_removed") // to_be_removed → to_be_removed
	model.Notes = getStringValue(apiData, "notes")

	return model
}

// getStringListValue extracts a string array from API response map with null safety.
// Returns a types.List of types.String elements, or null List if missing/invalid.
func getStringListValue(ctx context.Context, data map[string]interface{}, key string) types.List {
	if val, ok := data[key]; ok && val != nil {
		if arr, ok := val.([]interface{}); ok {
			// Convert []interface{} to []attr.Value
			elements := make([]attr.Value, 0, len(arr))
			for _, item := range arr {
				if str, ok := item.(string); ok {
					elements = append(elements, types.StringValue(str))
				} else {
					// Skip non-string elements (defensive)
					tflog.Warn(ctx, fmt.Sprintf("Skipping non-string element in array field %s", key))
				}
			}

			// Create List with StringType element type
			listValue, diags := types.ListValue(types.StringType, elements)
			if diags.HasError() {
				tflog.Error(ctx, fmt.Sprintf("Failed to create List for field %s: %v", key, diags))
				return types.ListNull(types.StringType)
			}
			return listValue
		}
	}
	return types.ListNull(types.StringType)
}

// applyPartitionFilters applies client-side filtering to partition list.
// Returns filtered slice of partitions matching all specified filter criteria.
//
// Filter behavior:
//   - name_pattern: Case-insensitive substring matching (e.g., "base" matches "base-partition", "BASE-PROD")
//   - Omitted filters are ignored (do not restrict results)
//
// Returns the original list if filter is nil or no filters are active.
func applyPartitionFilters(partitions []map[string]interface{}, filter *PartitionFilterModel) []map[string]interface{} {
	if filter == nil {
		return partitions
	}

	// Pre-process filter values (optimize by converting to lowercase once)
	var namePatternLower string
	hasNameFilter := false
	if !filter.NamePattern.IsNull() && !filter.NamePattern.IsUnknown() {
		namePatternLower = strings.ToLower(filter.NamePattern.ValueString())
		hasNameFilter = true
	}

	// Early return if no filters active
	if !hasNameFilter {
		return partitions
	}

	filtered := make([]map[string]interface{}, 0, len(partitions))

	for _, partition := range partitions {
		// Name pattern filter (case-insensitive substring match)
		if hasNameFilter {
			name := strings.ToLower(getStringValue(partition, "name").ValueString())
			if !strings.Contains(name, namePatternLower) {
				continue // Skip partition that doesn't match
			}
		}

		// Partition matches all filters (AND logic)
		filtered = append(filtered, partition)
	}

	return filtered
}
