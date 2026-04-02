// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &CMNetNetworksDataSource{}
	_ datasource.DataSourceWithConfigure = &CMNetNetworksDataSource{}
)

// NewCMNetNetworksDataSource is a helper function to simplify the provider implementation.
// It returns a new instance of the CMNetNetworksDataSource data source.
func NewCMNetNetworksDataSource() datasource.DataSource {
	return &CMNetNetworksDataSource{}
}

// CMNetNetworksDataSource is the data source implementation for querying BCM network configurations.
// It retrieves network information from the BCM CMNet service via the getNetworks API call.
type CMNetNetworksDataSource struct {
	BCMDataSourceBase
}

// CMNetNetworksDataSourceModel describes the data source data model.
// It represents the top-level structure returned by the bcm_cmnet_networks data source.
type CMNetNetworksDataSourceModel struct {
	ID       types.String        `tfsdk:"id"`       // Placeholder identifier for the data source
	Filter   *NetworkFilterModel `tfsdk:"filter"`   // Optional filter criteria for narrowing results
	Networks []NetworkModel      `tfsdk:"networks"` // List of network objects matching the filter
}

// NetworkFilterModel describes the filter block for client-side filtering.
// Multiple filters use AND logic (all filters must match for a network to be included).
type NetworkFilterModel struct {
	NamePattern types.String `tfsdk:"name_pattern"` // Case-insensitive substring match for network name
	DHCPEnabled types.Bool   `tfsdk:"dhcp_enabled"` // Exact boolean match for DHCP enabled status
}

// NetworkModel describes a single network entity from the BCM CMNet service.
// It maps API response fields to Terraform-friendly attribute names.
type NetworkModel struct {
	ID                  types.String `tfsdk:"id"`
	UUID                types.String `tfsdk:"uuid"`
	Name                types.String `tfsdk:"name"`
	BaseAddress         types.String `tfsdk:"base_address"`
	NetmaskBits         types.Int64  `tfsdk:"netmask_bits"`
	Gateway             types.String `tfsdk:"gateway"`
	NetworkType         types.String `tfsdk:"network_type"`
	MTU                 types.Int64  `tfsdk:"mtu"`
	DomainName          types.String `tfsdk:"domain_name"`
	GenerateDNSZone     types.String `tfsdk:"generate_dns_zone"`
	SearchDomainIndex   types.Int64  `tfsdk:"search_domain_index"`
	DHCPEnabled         types.Bool   `tfsdk:"dhcp_enabled"`
	DynamicRangeStart   types.String `tfsdk:"dynamic_range_start"`
	DynamicRangeEnd     types.String `tfsdk:"dynamic_range_end"`
	Management          types.Bool   `tfsdk:"management"`
	Bootable            types.Bool   `tfsdk:"bootable"`
	Layer3              types.Bool   `tfsdk:"layer3"`
	IPv6Enabled         types.Bool   `tfsdk:"ipv6_enabled"`
	IPv6BaseAddress     types.String `tfsdk:"ipv6_base_address"`
	IPv6Gateway         types.String `tfsdk:"ipv6_gateway"`
	IPv6NetmaskBits     types.Int64  `tfsdk:"ipv6_netmask_bits"`
	BaseType            types.String `tfsdk:"base_type"`
	ChildType           types.String `tfsdk:"child_type"`
	Revision            types.String `tfsdk:"revision"`
	Modified            types.Bool   `tfsdk:"modified"`
	ToBeRemoved         types.Bool   `tfsdk:"to_be_removed"`
	Layer3Route         types.String `tfsdk:"layer3_route"`
	GatewayMetric       types.Int64  `tfsdk:"gateway_metric"`
	AllowAutosign       types.String `tfsdk:"allow_autosign"`
	CloudSubnetID       types.String `tfsdk:"cloud_subnet_id"`
	EC2AvailabilityZone types.String `tfsdk:"ec2_availability_zone"`
	Notes               types.String `tfsdk:"notes"`
}

// Metadata returns the data source type name.
func (d *CMNetNetworksDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cmnet_networks"
}

// Schema defines the schema for the data source.
func (d *CMNetNetworksDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves network configurations from the BCM CMNet service.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Placeholder identifier for the data source.",
			},
			"networks": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of network objects matching the filter criteria.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Network UUID (same as uuid).",
						},
						"uuid": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Unique network identifier.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Network name.",
						},
						"base_address": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Network base IP address.",
						},
						"netmask_bits": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "CIDR netmask bits.",
						},
						"gateway": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Default gateway IP address.",
						},
						"network_type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Network type (INTERNAL, GLOBAL, etc.).",
						},
						"mtu": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Maximum transmission unit.",
						},
						"domain_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "DNS domain name.",
						},
						"generate_dns_zone": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "DNS zone generation policy.",
						},
						"search_domain_index": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Search domain priority index.",
						},
						"dhcp_enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "DHCP enabled flag (computed from dynamic range).",
						},
						"dynamic_range_start": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "DHCP range start IP address.",
						},
						"dynamic_range_end": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "DHCP range end IP address.",
						},
						"management": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Is management network flag.",
						},
						"bootable": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Network supports PXE boot.",
						},
						"layer3": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Layer 3 networking enabled.",
						},
						"ipv6_enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "IPv6 protocol enabled flag.",
						},
						"ipv6_base_address": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "IPv6 network base address.",
						},
						"ipv6_gateway": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "IPv6 gateway address.",
						},
						"ipv6_netmask_bits": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "IPv6 CIDR netmask bits.",
						},
						"base_type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Entity base type.",
						},
						"child_type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Network subtype.",
						},
						"revision": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Revision identifier.",
						},
						"modified": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Has unsaved changes pending.",
						},
						"to_be_removed": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Scheduled for deletion flag.",
						},
						"layer3_route": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Layer 3 routing mode.",
						},
						"gateway_metric": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Gateway routing metric.",
						},
						"allow_autosign": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Autosign policy.",
						},
						"cloud_subnet_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Cloud provider subnet identifier.",
						},
						"ec2_availability_zone": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "AWS EC2 availability zone.",
						},
						"notes": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "User-provided notes.",
						},
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"filter": schema.SingleNestedBlock{
				MarkdownDescription: "Optional filters to narrow down the list of networks.",
				Attributes: map[string]schema.Attribute{
					"name_pattern": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Case-insensitive substring match for network name.",
					},
					"dhcp_enabled": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Filter by DHCP enabled status.",
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *CMNetNetworksDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.ConfigureDataSource(req, resp)
}

// Read refreshes the Terraform state with the latest data.
func (d *CMNetNetworksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config CMNetNetworksDataSourceModel

	// Read configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading BCM networks from API", map[string]interface{}{
		"service": "cmnet",
		"call":    "getNetworks",
	})

	// Call BCM API
	body, err := d.Client.CallJSONRPC(ctx, "cmnet", "getNetworks")
	if err != nil {
		// Enhanced error handling with actionable guidance
		resp.Diagnostics.AddError(
			"Unable to Read BCM Networks",
			fmt.Sprintf("Failed to call BCM API: %s\n\n"+
				"Possible causes:\n"+
				"  - Authentication failure: Verify BCM credentials are correct\n"+
				"  - Network connectivity: Ensure BCM endpoint is reachable\n"+
				"  - API endpoint: Confirm BCM API is available at the configured endpoint\n\n"+
				"Error details: %s", d.Client.Endpoint, err.Error()),
		)
		return
	}

	tflog.Trace(ctx, "Successfully retrieved BCM API response", map[string]interface{}{
		"response_size": len(body),
	})

	// Parse response
	var apiResponse []map[string]interface{}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Parse BCM API Response",
			fmt.Sprintf("The BCM API returned malformed JSON that could not be parsed.\n\n"+
				"This may indicate an API version mismatch or internal BCM error.\n\n"+
				"Error details: %s\n"+
				"Response preview: %s", err.Error(), string(body[:minInt(len(body), 200)])),
		)
		return
	}

	tflog.Debug(ctx, "Parsed BCM API response", map[string]interface{}{
		"total_networks": len(apiResponse),
	})

	// Map and filter networks
	state := CMNetNetworksDataSourceModel{
		ID:       types.StringValue("cmnet-networks"),
		Filter:   config.Filter,
		Networks: []NetworkModel{},
	}

	// Log filter criteria if present
	if config.Filter != nil {
		filterCriteria := map[string]interface{}{}
		if !config.Filter.NamePattern.IsNull() {
			filterCriteria["name_pattern"] = config.Filter.NamePattern.ValueString()
		}
		if !config.Filter.DHCPEnabled.IsNull() {
			filterCriteria["dhcp_enabled"] = config.Filter.DHCPEnabled.ValueBool()
		}
		tflog.Debug(ctx, "Applying client-side filters", filterCriteria)
	}

	// Map and filter networks
	for _, netData := range apiResponse {
		network := mapAPIToNetwork(netData)
		if matchesNetworkFilter(network, config.Filter) {
			state.Networks = append(state.Networks, network)
		}
	}

	sort.Slice(state.Networks, func(i, j int) bool {
		if state.Networks[i].Name.ValueString() != state.Networks[j].Name.ValueString() {
			return state.Networks[i].Name.ValueString() < state.Networks[j].Name.ValueString()
		}
		if state.Networks[i].NetworkType.ValueString() != state.Networks[j].NetworkType.ValueString() {
			return state.Networks[i].NetworkType.ValueString() < state.Networks[j].NetworkType.ValueString()
		}
		return state.Networks[i].UUID.ValueString() < state.Networks[j].UUID.ValueString()
	})

	tflog.Info(ctx, "Network data source read complete", map[string]interface{}{
		"total_networks":    len(apiResponse),
		"filtered_networks": len(state.Networks),
	})

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// mapAPIToNetwork converts an API response map to a NetworkModel.
func mapAPIToNetwork(data map[string]interface{}) NetworkModel {
	// Compute DHCP enabled based on dynamic range
	dynamicStart := getStringValue(data, "dynamicRangeStart")
	dynamicEnd := getStringValue(data, "dynamicRangeEnd")
	dhcpEnabled := types.BoolValue(false)
	if !dynamicStart.IsNull() && !dynamicEnd.IsNull() &&
		dynamicStart.ValueString() != "0.0.0.0" && dynamicEnd.ValueString() != "0.0.0.0" {
		dhcpEnabled = types.BoolValue(true)
	}

	uuid := getStringValue(data, "uuid")

	return NetworkModel{
		ID:                  uuid, // ID is same as UUID
		UUID:                uuid,
		Name:                getStringValue(data, "name"),
		BaseAddress:         getStringValue(data, "baseAddress"),
		NetmaskBits:         getInt64Value(data, "netmaskBits"),
		Gateway:             getStringValue(data, "gateway"),
		NetworkType:         getStringValue(data, "type"),
		MTU:                 getInt64Value(data, "mtu"),
		DomainName:          getStringValue(data, "domainName"),
		GenerateDNSZone:     getStringValue(data, "generateDNSZone"),
		SearchDomainIndex:   getInt64Value(data, "searchDomainIndex"),
		DHCPEnabled:         dhcpEnabled,
		DynamicRangeStart:   dynamicStart,
		DynamicRangeEnd:     dynamicEnd,
		Management:          getBoolValue(data, "management"),
		Bootable:            getBoolValue(data, "bootable"),
		Layer3:              getBoolValue(data, "layer3"),
		IPv6Enabled:         getBoolValue(data, "IPv6"),
		IPv6BaseAddress:     getStringValue(data, "ipv6BaseAddress"),
		IPv6Gateway:         getStringValue(data, "ipv6Gateway"),
		IPv6NetmaskBits:     getInt64Value(data, "ipv6NetmaskBits"),
		BaseType:            getStringValue(data, "baseType"),
		ChildType:           getStringValue(data, "childType"),
		Revision:            getStringValue(data, "revision"),
		Modified:            getBoolValue(data, "modified"),
		ToBeRemoved:         getBoolValue(data, "to_be_removed"),
		Layer3Route:         getStringValue(data, "layer3route"),
		GatewayMetric:       getInt64Value(data, "gatewayMetric"),
		AllowAutosign:       getStringValue(data, "allowAutosign"),
		CloudSubnetID:       getStringValue(data, "cloudSubnetID"),
		EC2AvailabilityZone: getStringValue(data, "EC2AvailabilityZone"),
		Notes:               getStringValue(data, "notes"),
	}
}

// matchesNetworkFilter checks if a network matches the filter criteria.
// Multiple filters use AND logic - the network must match all specified filters.
//
// Filter behavior:
//   - name_pattern: Case-insensitive substring matching (e.g., "management" matches "managementnet", "Management-Net")
//   - dhcp_enabled: Exact boolean matching (true matches only DHCP-enabled networks, false matches only non-DHCP networks)
//   - Omitted filters are ignored (do not restrict results)
//
// Returns true if the network matches all specified filters, false otherwise.
func matchesNetworkFilter(network NetworkModel, filter *NetworkFilterModel) bool {
	// No filter means all networks match
	if filter == nil {
		return true
	}

	// Name pattern filter: case-insensitive substring matching
	if !filter.NamePattern.IsNull() && !filter.NamePattern.IsUnknown() {
		pattern := strings.ToLower(filter.NamePattern.ValueString())
		name := strings.ToLower(network.Name.ValueString())
		// Network name must contain the pattern (case-insensitive)
		if !strings.Contains(name, pattern) {
			return false
		}
	}

	// DHCP enabled filter: exact boolean matching
	if !filter.DHCPEnabled.IsNull() && !filter.DHCPEnabled.IsUnknown() {
		// Network DHCP status must exactly match the filter value
		if filter.DHCPEnabled.ValueBool() != network.DHCPEnabled.ValueBool() {
			return false
		}
	}

	// Network matches all specified filters
	return true
}
