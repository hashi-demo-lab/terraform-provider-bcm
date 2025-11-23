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
	_ datasource.DataSource              = &CMKubeClustersDataSource{}
	_ datasource.DataSourceWithConfigure = &CMKubeClustersDataSource{}
)

// NewCMKubeClustersDataSource is a helper function to simplify the provider implementation.
func NewCMKubeClustersDataSource() datasource.DataSource {
	return &CMKubeClustersDataSource{}
}

// CMKubeClustersDataSource is the data source implementation.
type CMKubeClustersDataSource struct {
	client *BCMClient
}

// CMKubeClustersDataSourceModel describes the data source data model.
type CMKubeClustersDataSourceModel struct {
	ID       types.String        `tfsdk:"id"`
	Filter   *ClusterFilterModel `tfsdk:"filter"`
	Clusters []KubeClusterModel  `tfsdk:"clusters"`
}

// ClusterFilterModel describes the filter block for client-side filtering.
// Multiple filters use AND logic (all filters must match for a cluster to be included).
type ClusterFilterModel struct {
	NamePattern  types.String `tfsdk:"name_pattern"`   // Case-insensitive substring match for cluster name
	Version      types.String `tfsdk:"version"`        // Exact match for Kubernetes version (semver)
	MasterNodeID types.String `tfsdk:"master_node_id"` // Find clusters containing this master node UUID
}

// KubeClusterModel represents a BCM Kubernetes cluster with all fields.
type KubeClusterModel struct {
	// Identity fields
	ID   types.String `tfsdk:"id"`
	UUID types.String `tfsdk:"uuid"`
	Name types.String `tfsdk:"name"`

	// Node configuration
	MasterNodes types.List `tfsdk:"master_nodes"`
	WorkerNodes types.List `tfsdk:"worker_nodes"`

	// Network configuration
	ManagementNetwork types.String `tfsdk:"management_network"`
	OverlayNetwork    types.String `tfsdk:"overlay_network"`
	DNSServers        types.List   `tfsdk:"dns_servers"`

	// Kubernetes configuration
	Version   types.String `tfsdk:"version"`
	CNIPlugin types.String `tfsdk:"cni_plugin"`

	// Storage configuration
	StorageClasses types.String `tfsdk:"storage_classes"`

	// Load balancer configuration
	LoadBalancerMode types.String `tfsdk:"load_balancer_mode"`

	// Cluster addons
	Addons            types.String `tfsdk:"addons"`
	IngressController types.String `tfsdk:"ingress_controller"`

	// Computed metadata
	CreationTime types.Int64 `tfsdk:"creation_time"`
	RevisionID   types.Int64 `tfsdk:"revision_id"`
}

// Metadata returns the data source type name.
func (d *CMKubeClustersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cmkube_clusters"
}

// Schema defines the schema for the data source.
func (d *CMKubeClustersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Data source to discover and filter BCM Kubernetes clusters.\n\n" +
			"Use this data source to list all Kubernetes clusters managed by BCM or filter by name pattern, " +
			"Kubernetes version, or master node UUID. This is useful for discovering clusters for import, " +
			"building dependencies, or operational queries.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Data source identifier (always 'cmkube-clusters').",
			},
			"clusters": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of Kubernetes clusters matching the filter criteria.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Cluster identifier (same as uuid).",
						},
						"uuid": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "BCM-assigned cluster UUID.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Cluster name.",
						},
						"master_nodes": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "List of master node UUIDs.",
						},
						"worker_nodes": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "List of worker node UUIDs.",
						},
						"management_network": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Management network UUID for cluster management traffic.",
						},
						"overlay_network": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Pod network overlay configuration (CIDR).",
						},
						"dns_servers": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "List of DNS server IP addresses.",
						},
						"version": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Kubernetes version (semver format).",
						},
						"cni_plugin": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Container Network Interface (CNI) plugin name.",
						},
						"storage_classes": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Storage class definitions (JSON-encoded).",
						},
						"load_balancer_mode": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Load balancer mode/strategy.",
						},
						"addons": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Cluster addons configuration (JSON-encoded).",
						},
						"ingress_controller": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Ingress controller configuration (JSON-encoded).",
						},
						"creation_time": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Cluster creation timestamp (Unix epoch).",
						},
						"revision_id": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "BCM revision number for change tracking.",
						},
					},
				},
			},
		},

		Blocks: map[string]schema.Block{
			"filter": schema.SingleNestedBlock{
				MarkdownDescription: "Optional filter criteria to limit returned clusters. " +
					"Multiple filters use AND logic (all filters must match).",
				Attributes: map[string]schema.Attribute{
					"name_pattern": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Case-insensitive substring match for cluster name (e.g., 'prod' matches 'Prod-Cluster-01').",
					},
					"version": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Exact match for Kubernetes version (semver format, e.g., '1.28.0').",
					},
					"master_node_id": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Find clusters containing this master node UUID in their master_nodes list.",
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *CMKubeClustersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*BCMClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			"Expected *BCMClient, got something else. Please report this issue to the provider developers.",
		)
		return
	}

	d.client = client
}

// Read refreshes the Terraform state with the latest data.
func (d *CMKubeClustersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CMKubeClustersDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading Kubernetes clusters from BCM API")

	// Call BCM API to retrieve all Kubernetes clusters
	body, err := d.client.CallJSONRPC(ctx, "cmkube", "getKubeClusters")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Kubernetes Clusters",
			fmt.Sprintf("Could not read Kubernetes clusters from BCM API: %s", err.Error()),
		)
		return
	}

	tflog.Trace(ctx, "BCM API response received", map[string]interface{}{
		"responseLength": len(body),
	})

	// Parse API response
	var apiResponse []map[string]interface{}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Kubernetes Clusters Response",
			fmt.Sprintf("Could not parse BCM API response: %s", err.Error()),
		)
		return
	}

	tflog.Debug(ctx, "Parsed BCM API response", map[string]interface{}{
		"clusterCount": len(apiResponse),
	})

	// Apply client-side filters (AND logic - all filters must match)
	var filteredClusters []KubeClusterModel
	for _, clusterData := range apiResponse {
		include := true

		// Filter 1: name_pattern (case-insensitive substring match)
		if data.Filter != nil && !data.Filter.NamePattern.IsNull() {
			pattern := strings.ToLower(data.Filter.NamePattern.ValueString())
			clusterName := strings.ToLower(getStringValue(clusterData, "name").ValueString())
			if !strings.Contains(clusterName, pattern) {
				include = false
				tflog.Trace(ctx, "Cluster excluded by name_pattern filter", map[string]interface{}{
					"clusterName": clusterName,
					"pattern":     pattern,
				})
			}
		}

		// Filter 2: version (exact match)
		if include && data.Filter != nil && !data.Filter.Version.IsNull() {
			clusterVersion := getStringValue(clusterData, "version").ValueString()
			filterVersion := data.Filter.Version.ValueString()
			if clusterVersion != filterVersion {
				include = false
				tflog.Trace(ctx, "Cluster excluded by version filter", map[string]interface{}{
					"clusterVersion": clusterVersion,
					"filterVersion":  filterVersion,
				})
			}
		}

		// Filter 3: master_node_id (ANY match in master_nodes list)
		if include && data.Filter != nil && !data.Filter.MasterNodeID.IsNull() {
			targetNodeID := data.Filter.MasterNodeID.ValueString()
			found := false

			// Extract master_nodes list from API response
			if masterNodesRaw, ok := clusterData["masterNodes"]; ok && masterNodesRaw != nil {
				if masterNodesSlice, ok := masterNodesRaw.([]interface{}); ok {
					for _, nodeUUID := range masterNodesSlice {
						if nodeStr, ok := nodeUUID.(string); ok && nodeStr == targetNodeID {
							found = true
							tflog.Trace(ctx, "Found matching master node", map[string]interface{}{
								"nodeUUID": nodeStr,
							})
							break
						}
					}
				}
			}

			if !found {
				include = false
				tflog.Trace(ctx, "Cluster excluded by master_node_id filter", map[string]interface{}{
					"targetNodeID": targetNodeID,
				})
			}
		}

		// Include cluster if it passes all filters
		if include {
			model := mapClusterDataToModel(clusterData)
			filteredClusters = append(filteredClusters, model)
		}
	}

	tflog.Debug(ctx, "Filtered clusters", map[string]interface{}{
		"totalClusters":    len(apiResponse),
		"filteredClusters": len(filteredClusters),
	})

	// Set state
	data.Clusters = filteredClusters
	data.ID = types.StringValue("cmkube-clusters")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Helper functions for data extraction (will be used in REFACTOR phase)

// getListValue extracts a list of strings from BCM API response with null handling.
// Returns types.List with StringType elements, or types.ListNull if field missing/null.
func getListValue(data map[string]interface{}, key string) types.List {
	if val, ok := data[key]; ok && val != nil {
		if slice, ok := val.([]interface{}); ok && len(slice) > 0 {
			// Convert []interface{} to []attr.Value (StringValue elements)
			elements := make([]attr.Value, 0, len(slice))
			for _, item := range slice {
				if str, ok := item.(string); ok && str != "" {
					elements = append(elements, types.StringValue(str))
				}
			}

			if len(elements) > 0 {
				listValue, _ := types.ListValue(types.StringType, elements)
				return listValue
			}
		}
	}

	// Return null list if field missing, null, or empty
	return types.ListNull(types.StringType)
}

// mapClusterDataToModel maps BCM API fields (camelCase) to Terraform attributes (snake_case).
// This will be used in REFACTOR phase when integrating with real BCM API.
func mapClusterDataToModel(apiData map[string]interface{}) KubeClusterModel {
	model := KubeClusterModel{}

	// Identity fields
	uuid := getStringValue(apiData, "uuid")
	model.ID = uuid
	model.UUID = uuid
	model.Name = getStringValue(apiData, "name")

	// Node configuration
	model.MasterNodes = getListValue(apiData, "masterNodes")
	model.WorkerNodes = getListValue(apiData, "workerNodes")

	// Network configuration
	model.ManagementNetwork = getStringValue(apiData, "managementNetwork")
	model.OverlayNetwork = getStringValue(apiData, "overlayNetwork")
	model.DNSServers = getListValue(apiData, "dnsServers")

	// Kubernetes configuration
	model.Version = getStringValue(apiData, "version")
	model.CNIPlugin = getStringValue(apiData, "cniPlugin")

	// Storage configuration
	model.StorageClasses = getStringValue(apiData, "storageClasses")

	// Load balancer configuration
	model.LoadBalancerMode = getStringValue(apiData, "loadBalancerMode")

	// Cluster addons
	model.Addons = getStringValue(apiData, "addons")
	model.IngressController = getStringValue(apiData, "ingressController")

	// Computed metadata
	model.CreationTime = getInt64Value(apiData, "creationTime")
	model.RevisionID = getInt64Value(apiData, "revisionID")

	return model
}

// Note: getStringValue and getInt64Value are already defined in data_source_cmpart_softwareimages.go
// They will be reused for null-safe extraction.
