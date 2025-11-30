// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &CMDeviceNodesDataSource{}
	_ datasource.DataSourceWithConfigure = &CMDeviceNodesDataSource{}
)

// NewCMDeviceNodesDataSource is a helper function to simplify the provider implementation.
func NewCMDeviceNodesDataSource() datasource.DataSource {
	return &CMDeviceNodesDataSource{}
}

// CMDeviceNodesDataSource is the data source implementation.
type CMDeviceNodesDataSource struct {
	BCMDataSourceBase
}

// CMDeviceNodesDataSourceModel describes the data source data model.
type CMDeviceNodesDataSourceModel struct {
	ID     types.String `tfsdk:"id"`
	Filter *FilterModel `tfsdk:"filter"`
	Nodes  []NodeModel  `tfsdk:"nodes"`
}

// FilterModel represents client-side filtering configuration.
type FilterModel struct {
	ChildType       types.String `tfsdk:"child_type"`
	CategoryUUID    types.String `tfsdk:"category_uuid"`
	HostnamePattern types.String `tfsdk:"hostname_pattern"`
}

// NodeModel represents a BCM cluster node.
type NodeModel struct {
	// Identity
	ID           types.String `tfsdk:"id"`
	UUID         types.String `tfsdk:"uuid"`
	Hostname     types.String `tfsdk:"hostname"`
	BaseType     types.String `tfsdk:"base_type"`
	ChildType    types.String `tfsdk:"child_type"`
	MAC          types.String `tfsdk:"mac"`
	CreationTime types.Int64  `tfsdk:"creation_time"`

	// Network
	Interfaces []NetworkInterfaceModel `tfsdk:"interfaces"`

	// Roles
	Roles []RoleModel `tfsdk:"roles"`

	// Categorization
	Category  types.String `tfsdk:"category"`
	Partition types.String `tfsdk:"partition"`

	// Management
	PowerControl          types.String `tfsdk:"power_control"`
	AuthenticationService types.String `tfsdk:"authentication_service"`
	ProvisioningTransport types.String `tfsdk:"provisioning_transport"`

	// State
	Modified    types.Bool `tfsdk:"modified"`
	ToBeRemoved types.Bool `tfsdk:"to_be_removed"`
}

// NetworkInterfaceModel represents a network interface.
type NetworkInterfaceModel struct {
	Name      types.String `tfsdk:"name"`
	MAC       types.String `tfsdk:"mac"`
	IP        types.String `tfsdk:"ip"`
	IPv6IP    types.String `tfsdk:"ipv6_ip"`
	DHCP      types.Bool   `tfsdk:"dhcp"`
	Network   types.String `tfsdk:"network"`
	BaseType  types.String `tfsdk:"base_type"`
	ChildType types.String `tfsdk:"child_type"`
	CardType  types.String `tfsdk:"cardtype"`
	Bootable  types.Bool   `tfsdk:"bootable"`
	StartIf   types.String `tfsdk:"start_if"`
}

// RoleModel represents a service role.
type RoleModel struct {
	UUID        types.String `tfsdk:"uuid"`
	Name        types.String `tfsdk:"name"`
	BaseType    types.String `tfsdk:"base_type"`
	ChildType   types.String `tfsdk:"child_type"`
	AddServices types.Bool   `tfsdk:"add_services"`
}

// Metadata returns the data source type name.
func (d *CMDeviceNodesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cmdevice_nodes"
}

// Schema defines the schema for the data source.
func (d *CMDeviceNodesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches cluster nodes from BCM CMDevice service with comprehensive metadata including network interfaces and roles.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Placeholder identifier for the data source.",
			},
			"nodes": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of cluster nodes matching the filter criteria.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Node UUID (same as uuid).",
						},
						"uuid": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Unique node identifier (RFC 4122 UUID format).",
						},
						"hostname": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Node hostname.",
						},
						"base_type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Entity base type (always 'Device').",
						},
						"child_type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Specific node type (PhysicalNode, HeadNode, ComputeNode, CloudNode, StorageNode).",
						},
						"mac": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Primary MAC address.",
						},
						"creation_time": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Node creation time (Unix timestamp in seconds).",
						},
						"category": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Category UUID.",
						},
						"partition": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Partition UUID.",
						},
						"power_control": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Power control method (none, ipmi, redfish, pdu).",
						},
						"authentication_service": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Authentication service type.",
						},
						"provisioning_transport": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Provisioning transport method.",
						},
						"modified": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Indicates if node has unsaved changes.",
						},
						"to_be_removed": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Indicates if node is scheduled for deletion.",
						},
						"interfaces": schema.ListNestedAttribute{
							Computed:            true,
							MarkdownDescription: "Network interfaces configured on the node.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Interface name (e.g., ens33).",
									},
									"mac": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Interface MAC address.",
									},
									"ip": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "IPv4 address.",
									},
									"ipv6_ip": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "IPv6 address.",
									},
									"dhcp": schema.BoolAttribute{
										Computed:            true,
										MarkdownDescription: "DHCP enabled flag.",
									},
									"network": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Network UUID reference.",
									},
									"base_type": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Entity base type (always 'NetworkInterface').",
									},
									"child_type": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Interface type (NetworkPhysicalInterface, NetworkBondInterface, etc.).",
									},
									"cardtype": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Network card type (Ethernet, InfiniBand, etc.).",
									},
									"bootable": schema.BoolAttribute{
										Computed:            true,
										MarkdownDescription: "PXE boot capable flag.",
									},
									"start_if": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Startup condition (ALWAYS, NEVER, HOTPLUG).",
									},
								},
							},
						},
						"roles": schema.ListNestedAttribute{
							Computed:            true,
							MarkdownDescription: "Service roles assigned to the node.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"uuid": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Role UUID.",
									},
									"name": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Role name.",
									},
									"base_type": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Entity base type (always 'Role').",
									},
									"child_type": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "Role type (HeadNodeRole, ComputeRole, etc.).",
									},
									"add_services": schema.BoolAttribute{
										Computed:            true,
										MarkdownDescription: "Auto-add related services flag.",
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
				MarkdownDescription: "Optional filter criteria to limit returned nodes. Multiple filters are AND-ed together.",
				Attributes: map[string]schema.Attribute{
					"child_type": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Filter by node childType (exact match, case-sensitive). Example: 'PhysicalNode'.",
					},
					"category_uuid": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Filter by category UUID (exact match).",
					},
					"hostname_pattern": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Filter by hostname substring (case-insensitive). Example: 'compute' matches 'compute-node-01'.",
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *CMDeviceNodesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.ConfigureDataSource(req, resp)
}

// Read refreshes the Terraform state with the latest data.
func (d *CMDeviceNodesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config CMDeviceNodesDataSourceModel

	// Read configuration from request
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call BCM API to get nodes
	tflog.Debug(ctx, "Calling BCM API: cmdevice.getNodes")
	body, err := d.Client.CallJSONRPC(ctx, "cmdevice", "getNodes")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read BCM Nodes",
			fmt.Sprintf("Error calling cmdevice.getNodes API: %s\n\nPlease verify:\n- BCM endpoint is accessible\n- Credentials are correct\n- BCM service is running", err.Error()),
		)
		return
	}

	// Parse API response
	var apiResponse []map[string]interface{}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Parse API Response",
			fmt.Sprintf("Error unmarshaling JSON response from cmdevice.getNodes: %s\n\nResponse body: %s", err.Error(), string(body)),
		)
		return
	}

	// Build state model
	state := CMDeviceNodesDataSourceModel{
		ID:     types.StringValue("placeholder"),
		Filter: config.Filter,
		Nodes:  make([]NodeModel, 0, len(apiResponse)),
	}

	// Map API response to models with filtering
	for _, nodeData := range apiResponse {
		node := mapAPIToNode(nodeData)
		if matchesFilter(node, config.Filter) {
			state.Nodes = append(state.Nodes, node)
		}
	}

	tflog.Debug(ctx, "Successfully retrieved and filtered nodes", map[string]interface{}{
		"total_nodes":    len(apiResponse),
		"filtered_nodes": len(state.Nodes),
	})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// mapAPIToNode converts API response data to NodeModel.
func mapAPIToNode(apiData map[string]interface{}) NodeModel {
	model := NodeModel{
		// Identity
		UUID:         getStringValue(apiData, "uuid"),
		Hostname:     getStringValue(apiData, "hostname"),
		BaseType:     getStringValue(apiData, "baseType"),
		ChildType:    getStringValue(apiData, "childType"),
		MAC:          getStringValue(apiData, "mac"),
		CreationTime: getInt64Value(apiData, "creationTime"),

		// Categorization
		Category:  getStringValue(apiData, "category"),
		Partition: getStringValue(apiData, "partition"),

		// Management
		PowerControl:          getStringValue(apiData, "powerControl"),
		AuthenticationService: getStringValue(apiData, "authenticationService"),
		ProvisioningTransport: getStringValue(apiData, "provisioningTransport"),

		// State
		Modified:    getBoolValue(apiData, "modified"),
		ToBeRemoved: getBoolValue(apiData, "to_be_removed"),

		// Nested attributes
		Interfaces: mapInterfaces(apiData["interfaces"]),
		Roles:      mapRoles(apiData["roles"]),
	}

	// ID is same as UUID
	model.ID = model.UUID

	return model
}

// mapInterfaces converts API interfaces array to model slice.
func mapInterfaces(data interface{}) []NetworkInterfaceModel {
	interfaceArray, ok := data.([]interface{})
	if !ok || interfaceArray == nil {
		return []NetworkInterfaceModel{}
	}

	models := make([]NetworkInterfaceModel, 0, len(interfaceArray))
	for _, iface := range interfaceArray {
		ifaceMap, ok := iface.(map[string]interface{})
		if !ok {
			continue
		}

		model := NetworkInterfaceModel{
			Name:      getStringValue(ifaceMap, "name"),
			MAC:       getStringValue(ifaceMap, "mac"),
			IP:        getStringValue(ifaceMap, "ip"),
			IPv6IP:    getStringValue(ifaceMap, "ipv6Ip"),
			DHCP:      getBoolValue(ifaceMap, "dhcp"),
			Network:   getStringValue(ifaceMap, "network"),
			BaseType:  getStringValue(ifaceMap, "baseType"),
			ChildType: getStringValue(ifaceMap, "childType"),
			CardType:  getStringValue(ifaceMap, "cardtype"),
			Bootable:  getBoolValue(ifaceMap, "bootable"),
			StartIf:   getStringValue(ifaceMap, "startIf"),
		}
		models = append(models, model)
	}

	return models
}

// mapRoles converts API roles array to model slice.
func mapRoles(data interface{}) []RoleModel {
	roleArray, ok := data.([]interface{})
	if !ok || roleArray == nil {
		return []RoleModel{}
	}

	models := make([]RoleModel, 0, len(roleArray))
	for _, role := range roleArray {
		roleMap, ok := role.(map[string]interface{})
		if !ok {
			continue
		}

		model := RoleModel{
			UUID:        getStringValue(roleMap, "uuid"),
			Name:        getStringValue(roleMap, "name"),
			BaseType:    getStringValue(roleMap, "baseType"),
			ChildType:   getStringValue(roleMap, "childType"),
			AddServices: getBoolValue(roleMap, "addServices"),
		}
		models = append(models, model)
	}

	return models
}

// matchesFilter checks if a node matches the filter criteria.
func matchesFilter(node NodeModel, filter *FilterModel) bool {
	if filter == nil {
		return true
	}

	// Filter by child_type (exact match)
	if !filter.ChildType.IsNull() && !filter.ChildType.IsUnknown() {
		if node.ChildType.ValueString() != filter.ChildType.ValueString() {
			return false
		}
	}

	// Filter by category_uuid (exact match)
	if !filter.CategoryUUID.IsNull() && !filter.CategoryUUID.IsUnknown() {
		if node.Category.ValueString() != filter.CategoryUUID.ValueString() {
			return false
		}
	}

	// Filter by hostname_pattern (case-insensitive substring)
	if !filter.HostnamePattern.IsNull() && !filter.HostnamePattern.IsUnknown() {
		pattern := strings.ToLower(filter.HostnamePattern.ValueString())
		hostname := strings.ToLower(node.Hostname.ValueString())
		if !strings.Contains(hostname, pattern) {
			return false
		}
	}

	return true
}

// Note: Helper functions getStringValue, getBoolValue, getInt64Value are defined in data_source_cmpart_softwareimages.go
