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
	BCMDataSourceBase
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
	NamePattern types.String `tfsdk:"name_pattern"` // Case-insensitive substring match for cluster name
	Version     types.String `tfsdk:"version"`      // Exact match for Kubernetes version (semver)
	EtcdCluster types.String `tfsdk:"etcd_cluster"` // Filter by EtcdCluster UUID reference
}

// KubeClusterModel represents a BCM Kubernetes cluster aligned with BCM CMKube API entity.
// This model matches KubeClusterAlignedResourceModel for consistency between resource and data source.
type KubeClusterModel struct {
	// Identity fields
	ID   types.String `tfsdk:"id"`
	UUID types.String `tfsdk:"uuid"`
	Name types.String `tfsdk:"name"`

	// Network references (UUIDs)
	InternalNetwork types.String `tfsdk:"internal_network"`
	ServiceNetwork  types.String `tfsdk:"service_network"`
	PodNetwork      types.String `tfsdk:"pod_network"`

	// EtcdCluster reference (UUID)
	EtcdCluster types.String `tfsdk:"etcd_cluster"`

	// Kubernetes configuration
	Version            types.String `tfsdk:"version"`
	PodNetworkNodeMask types.String `tfsdk:"pod_network_node_mask"`
	KubeDnsIP          types.String `tfsdk:"kube_dns_ip"`

	// API server configuration
	KubernetesAPIServer          types.String `tfsdk:"kubernetes_api_server"`
	KubernetesAPIServerProxyPort types.Int64  `tfsdk:"kubernetes_api_server_proxy_port"`
	TrustedDomains               types.List   `tfsdk:"trusted_domains"`

	// Ingress proxy configuration
	IngressProxyEnable      types.Bool  `tfsdk:"ingress_proxy_enable"`
	IngressProxyListenPort  types.Int64 `tfsdk:"ingress_proxy_listen_port"`
	IngressProxyBackendPort types.Int64 `tfsdk:"ingress_proxy_backend_port"`

	// Extensible options
	Options types.String `tfsdk:"options"`

	// Nested blocks
	AppGroups types.List `tfsdk:"app_groups"`
	LabelSets types.List `tfsdk:"label_sets"`
	Users     types.List `tfsdk:"users"`

	// Computed metadata
	CreationTime types.Int64 `tfsdk:"creation_time"`
	RevisionID   types.Int64 `tfsdk:"revision_id"`
}

// Metadata returns the data source type name.
func (d *CMKubeClustersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cmkube_clusters"
}

// Schema defines the schema for the data source - aligned with BCM CMKube API entity.
func (d *CMKubeClustersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Data source to discover and filter BCM Kubernetes clusters.\n\n" +
			"Use this data source to list all Kubernetes clusters managed by BCM or filter by name pattern, " +
			"Kubernetes version, or etcd cluster UUID. This is useful for discovering clusters for import, " +
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
						// Identity fields
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

						// Network references
						"internal_network": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "UUID of the internal network for cluster node communication.",
						},
						"service_network": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "UUID of the service network for Kubernetes service IPs.",
						},
						"pod_network": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "UUID of the pod network for container IPs.",
						},

						// EtcdCluster reference
						"etcd_cluster": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "UUID of the EtcdCluster entity that backs this Kubernetes cluster.",
						},

						// Kubernetes configuration
						"version": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Kubernetes version (semver format).",
						},
						"pod_network_node_mask": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Pod network node mask for CIDR allocation (e.g., '/24').",
						},
						"kube_dns_ip": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Cluster DNS IP address.",
						},

						// API server configuration
						"kubernetes_api_server": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Kubernetes API server URL.",
						},
						"kubernetes_api_server_proxy_port": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Kubernetes API server proxy port.",
						},
						"trusted_domains": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "List of trusted domains for certificate SANs.",
						},

						// Ingress proxy configuration
						"ingress_proxy_enable": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Enable ingress proxy for external traffic routing.",
						},
						"ingress_proxy_listen_port": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Ingress proxy listen port.",
						},
						"ingress_proxy_backend_port": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Ingress proxy backend port.",
						},

						// Extensible options
						"options": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Extensible configuration options as JSON string.",
						},

						// Nested blocks as nested attributes for data source
						"app_groups": schema.ListNestedAttribute{
							Computed:            true,
							MarkdownDescription: "Application groups for cluster addons.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Application group name.",
									},
									"enabled": schema.BoolAttribute{
										Computed:            true,
										MarkdownDescription: "Whether the application group is enabled.",
									},
									"applications": schema.ListNestedAttribute{
										Computed:            true,
										MarkdownDescription: "Applications within this group.",
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"name": schema.StringAttribute{
													Computed:            true,
													MarkdownDescription: "Application name.",
												},
												"enabled": schema.BoolAttribute{
													Computed:            true,
													MarkdownDescription: "Whether the application is enabled.",
												},
												"manifest": schema.StringAttribute{
													Computed:            true,
													MarkdownDescription: "Kubernetes manifest YAML/JSON content.",
												},
											},
										},
									},
								},
							},
						},
						"label_sets": schema.ListNestedAttribute{
							Computed:            true,
							MarkdownDescription: "Label sets that can be applied to nodes, categories, or overlays.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Label set name.",
									},
									"labels": schema.MapAttribute{
										ElementType:         types.StringType,
										Computed:            true,
										MarkdownDescription: "Map of label key-value pairs.",
									},
								},
							},
						},
						"users": schema.ListNestedAttribute{
							Computed:            true,
							MarkdownDescription: "Kubernetes users for kubeconfig management.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "User name.",
									},
									"groups": schema.ListAttribute{
										ElementType:         types.StringType,
										Computed:            true,
										MarkdownDescription: "List of groups the user belongs to.",
									},
								},
							},
						},

						// Computed metadata
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
						MarkdownDescription: "Case-insensitive substring match for cluster name (e.g., 'prod' matches 'prod-cluster-01').",
					},
					"version": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Exact match for Kubernetes version (semver format, e.g., '1.28.0').",
					},
					"etcd_cluster": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Filter by EtcdCluster UUID reference.",
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *CMKubeClustersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.ConfigureDataSource(req, resp)
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
	body, err := d.Client.CallJSONRPC(ctx, "cmkube", "getKubeClusters")
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

		// Filter 3: etcd_cluster (exact match on etcdCluster UUID reference)
		if include && data.Filter != nil && !data.Filter.EtcdCluster.IsNull() {
			filterEtcdCluster := data.Filter.EtcdCluster.ValueString()
			clusterEtcdCluster := getStringValue(clusterData, "etcdCluster").ValueString()
			if clusterEtcdCluster != filterEtcdCluster {
				include = false
				tflog.Trace(ctx, "Cluster excluded by etcd_cluster filter", map[string]interface{}{
					"clusterEtcdCluster": clusterEtcdCluster,
					"filterEtcdCluster":  filterEtcdCluster,
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

// mapClusterDataToModel maps BCM API fields (camelCase) to Terraform attributes (snake_case).
// Aligned with BCM CMKube API entity structure.
func mapClusterDataToModel(apiData map[string]interface{}) KubeClusterModel {
	model := KubeClusterModel{}

	// Identity fields
	uuid := getStringValue(apiData, "uuid")
	model.ID = uuid
	model.UUID = uuid
	model.Name = getStringValue(apiData, "name")

	// Network references
	model.InternalNetwork = getStringValue(apiData, "internalNetwork")
	model.ServiceNetwork = getStringValue(apiData, "serviceNetwork")
	model.PodNetwork = getStringValue(apiData, "podNetwork")

	// EtcdCluster reference
	model.EtcdCluster = getStringValue(apiData, "etcdCluster")

	// Kubernetes configuration
	model.Version = getStringValue(apiData, "version")
	model.PodNetworkNodeMask = getStringValue(apiData, "podNetworkNodeMask")
	model.KubeDnsIP = getStringValue(apiData, "kubeDnsIp")

	// API server configuration
	model.KubernetesAPIServer = getStringValue(apiData, "kubernetesApiServer")
	model.KubernetesAPIServerProxyPort = getInt64Value(apiData, "kubernetesApiServerProxyPort")

	// Trusted domains
	if trustedDomains, ok := apiData["trustedDomains"].([]interface{}); ok && len(trustedDomains) > 0 {
		elements := make([]attr.Value, 0, len(trustedDomains))
		for _, domain := range trustedDomains {
			if domainStr, ok := domain.(string); ok {
				elements = append(elements, types.StringValue(domainStr))
			}
		}
		model.TrustedDomains, _ = types.ListValue(types.StringType, elements)
	} else {
		model.TrustedDomains, _ = types.ListValue(types.StringType, []attr.Value{})
	}

	// Ingress proxy configuration
	model.IngressProxyEnable = getBoolValue(apiData, "ingressProxyEnable")
	model.IngressProxyListenPort = getInt64Value(apiData, "ingressProxyListenPort")
	model.IngressProxyBackendPort = getInt64Value(apiData, "ingressProxyBackendPort")

	// Options as JSON string
	if options, ok := apiData["options"]; ok && options != nil {
		if optMap, ok := options.(map[string]interface{}); ok {
			optBytes, err := json.Marshal(optMap)
			if err == nil {
				model.Options = types.StringValue(string(optBytes))
			} else {
				model.Options = types.StringValue("{}")
			}
		} else {
			model.Options = types.StringValue("{}")
		}
	} else {
		model.Options = types.StringValue("{}")
	}

	// Nested blocks - AppGroups
	if appGroupsData, ok := apiData["appGroups"].([]interface{}); ok && len(appGroupsData) > 0 {
		model.AppGroups = mapAppGroupsToModel(appGroupsData)
	} else {
		model.AppGroups = types.ListNull(kubeAppGroupDataSourceType())
	}

	// Nested blocks - LabelSets
	if labelSetsData, ok := apiData["labelSets"].([]interface{}); ok && len(labelSetsData) > 0 {
		model.LabelSets = mapLabelSetsToModel(labelSetsData)
	} else {
		model.LabelSets = types.ListNull(kubeLabelSetDataSourceType())
	}

	// Nested blocks - Users
	if usersData, ok := apiData["users"].([]interface{}); ok && len(usersData) > 0 {
		model.Users = mapUsersToModel(usersData)
	} else {
		model.Users = types.ListNull(kubeUserDataSourceType())
	}

	// Computed metadata
	model.CreationTime = getInt64Value(apiData, "creationTime")
	model.RevisionID = getInt64Value(apiData, "revisionID")

	return model
}

// kubeAppGroupDataSourceType returns the object type for KubeAppGroup in data source context.
func kubeAppGroupDataSourceType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":    types.StringType,
			"enabled": types.BoolType,
			"applications": types.ListType{
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name":     types.StringType,
						"enabled":  types.BoolType,
						"manifest": types.StringType,
					},
				},
			},
		},
	}
}

// kubeLabelSetDataSourceType returns the object type for KubeLabelSet in data source context.
func kubeLabelSetDataSourceType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name": types.StringType,
			"labels": types.MapType{
				ElemType: types.StringType,
			},
		},
	}
}

// kubeUserDataSourceType returns the object type for KubeUser in data source context.
func kubeUserDataSourceType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name": types.StringType,
			"groups": types.ListType{
				ElemType: types.StringType,
			},
		},
	}
}

// mapAppGroupsToModel maps BCM appGroups to Terraform types.List.
func mapAppGroupsToModel(appGroupsData []interface{}) types.List {
	appGroupElements := make([]attr.Value, 0, len(appGroupsData))

	for _, agData := range appGroupsData {
		agMap, ok := agData.(map[string]interface{})
		if !ok {
			continue
		}

		// Parse applications within the group
		var applicationsValue types.List
		if appsData, ok := agMap["applications"].([]interface{}); ok && len(appsData) > 0 {
			appElements := make([]attr.Value, 0, len(appsData))
			for _, appData := range appsData {
				appMap, ok := appData.(map[string]interface{})
				if !ok {
					continue
				}
				appObj, _ := types.ObjectValue(
					map[string]attr.Type{
						"name":     types.StringType,
						"enabled":  types.BoolType,
						"manifest": types.StringType,
					},
					map[string]attr.Value{
						"name":     getStringValueForTFDS(appMap, "name"),
						"enabled":  getBoolValueForTFDS(appMap, "enabled"),
						"manifest": getStringValueForTFDS(appMap, "manifest"),
					},
				)
				appElements = append(appElements, appObj)
			}
			applicationsValue, _ = types.ListValue(
				types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name":     types.StringType,
						"enabled":  types.BoolType,
						"manifest": types.StringType,
					},
				},
				appElements,
			)
		} else {
			applicationsValue, _ = types.ListValue(
				types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name":     types.StringType,
						"enabled":  types.BoolType,
						"manifest": types.StringType,
					},
				},
				[]attr.Value{},
			)
		}

		agObj, _ := types.ObjectValue(
			map[string]attr.Type{
				"name":    types.StringType,
				"enabled": types.BoolType,
				"applications": types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"name":     types.StringType,
							"enabled":  types.BoolType,
							"manifest": types.StringType,
						},
					},
				},
			},
			map[string]attr.Value{
				"name":         getStringValueForTFDS(agMap, "name"),
				"enabled":      getBoolValueForTFDS(agMap, "enabled"),
				"applications": applicationsValue,
			},
		)
		appGroupElements = append(appGroupElements, agObj)
	}

	result, _ := types.ListValue(kubeAppGroupDataSourceType(), appGroupElements)
	return result
}

// mapLabelSetsToModel maps BCM labelSets to Terraform types.List.
func mapLabelSetsToModel(labelSetsData []interface{}) types.List {
	labelSetElements := make([]attr.Value, 0, len(labelSetsData))

	for _, lsData := range labelSetsData {
		lsMap, ok := lsData.(map[string]interface{})
		if !ok {
			continue
		}

		// Parse labels map
		var labelsValue types.Map
		if labelsData, ok := lsMap["labels"].(map[string]interface{}); ok && len(labelsData) > 0 {
			labelElements := make(map[string]attr.Value)
			for k, v := range labelsData {
				if vStr, ok := v.(string); ok {
					labelElements[k] = types.StringValue(vStr)
				}
			}
			labelsValue, _ = types.MapValue(types.StringType, labelElements)
		} else {
			labelsValue, _ = types.MapValue(types.StringType, map[string]attr.Value{})
		}

		lsObj, _ := types.ObjectValue(
			map[string]attr.Type{
				"name": types.StringType,
				"labels": types.MapType{
					ElemType: types.StringType,
				},
			},
			map[string]attr.Value{
				"name":   getStringValueForTFDS(lsMap, "name"),
				"labels": labelsValue,
			},
		)
		labelSetElements = append(labelSetElements, lsObj)
	}

	result, _ := types.ListValue(kubeLabelSetDataSourceType(), labelSetElements)
	return result
}

// mapUsersToModel maps BCM users to Terraform types.List.
func mapUsersToModel(usersData []interface{}) types.List {
	userElements := make([]attr.Value, 0, len(usersData))

	for _, uData := range usersData {
		uMap, ok := uData.(map[string]interface{})
		if !ok {
			continue
		}

		// Parse groups list
		var groupsValue types.List
		if groupsData, ok := uMap["groups"].([]interface{}); ok && len(groupsData) > 0 {
			groupElements := make([]attr.Value, 0, len(groupsData))
			for _, g := range groupsData {
				if gStr, ok := g.(string); ok {
					groupElements = append(groupElements, types.StringValue(gStr))
				}
			}
			groupsValue, _ = types.ListValue(types.StringType, groupElements)
		} else {
			groupsValue, _ = types.ListValue(types.StringType, []attr.Value{})
		}

		uObj, _ := types.ObjectValue(
			map[string]attr.Type{
				"name": types.StringType,
				"groups": types.ListType{
					ElemType: types.StringType,
				},
			},
			map[string]attr.Value{
				"name":   getStringValueForTFDS(uMap, "name"),
				"groups": groupsValue,
			},
		)
		userElements = append(userElements, uObj)
	}

	result, _ := types.ListValue(kubeUserDataSourceType(), userElements)
	return result
}

// getStringValueForTFDS extracts a string value from a map and returns it as types.String.
// DS suffix to avoid collision with resource helpers.
func getStringValueForTFDS(data map[string]interface{}, key string) types.String {
	if v, ok := data[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return types.StringValue(s)
		}
	}
	return types.StringValue("")
}

// getBoolValueForTFDS extracts a bool value from a map and returns it as types.Bool.
// DS suffix to avoid collision with resource helpers.
func getBoolValueForTFDS(data map[string]interface{}, key string) types.Bool {
	if v, ok := data[key]; ok && v != nil {
		if b, ok := v.(bool); ok {
			return types.BoolValue(b)
		}
	}
	return types.BoolValue(false)
}

// Note: getStringValue, getInt64Value, getBoolValue are already defined elsewhere.
// They will be reused for null-safe extraction.
