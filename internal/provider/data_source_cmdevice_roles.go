// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &CMDeviceRolesDataSource{}

func NewCMDeviceRolesDataSource() datasource.DataSource {
	return &CMDeviceRolesDataSource{}
}

// CMDeviceRolesDataSource defines the data source implementation.
type CMDeviceRolesDataSource struct {
	BCMDataSourceBase
}

// CMDeviceRolesDataSourceModel describes the data source data model.
type CMDeviceRolesDataSourceModel struct {
	ID          types.String               `tfsdk:"id"`
	NamePattern types.String               `tfsdk:"name_pattern"`
	ChildType   types.String               `tfsdk:"child_type"`
	Roles       []RolesDataSourceRoleModel `tfsdk:"roles"`
}

// RolesDataSourceRoleModel describes a single role in the roles data source.
// This is separate from RoleModel in data_source_cmdevice_nodes.go which doesn't have an id field.
type RolesDataSourceRoleModel struct {
	ID          types.String `tfsdk:"id"`
	UUID        types.String `tfsdk:"uuid"`
	Name        types.String `tfsdk:"name"`
	ChildType   types.String `tfsdk:"child_type"`
	BaseType    types.String `tfsdk:"base_type"`
	AddServices types.Bool   `tfsdk:"add_services"`
}

func (d *CMDeviceRolesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cmdevice_roles"
}

func (d *CMDeviceRolesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches available role types from BCM for device role assignment. Roles are extracted from nodes via cmdevice.getNodes and deduplicated by UUID.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier",
				Computed:            true,
			},
			"name_pattern": schema.StringAttribute{
				MarkdownDescription: "Glob pattern to filter roles by name (e.g., 'kube-*', 'head*'). Supports wildcards: * (any chars), ? (single char), [abc] (char class).",
				Optional:            true,
			},
			"child_type": schema.StringAttribute{
				MarkdownDescription: "Exact match filter for role type (e.g., 'HeadNodeRole', 'ComputeRole', 'StorageRole').",
				Optional:            true,
			},
			"roles": schema.ListNestedAttribute{
				MarkdownDescription: "List of roles matching filter criteria",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Role identifier (same as UUID)",
							Computed:            true,
						},
						"uuid": schema.StringAttribute{
							MarkdownDescription: "Unique role identifier",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Role name (e.g., 'headnode', 'compute', 'storage')",
							Computed:            true,
						},
						"child_type": schema.StringAttribute{
							MarkdownDescription: "Role type (e.g., 'HeadNodeRole', 'ComputeRole')",
							Computed:            true,
						},
						"base_type": schema.StringAttribute{
							MarkdownDescription: "Base entity type (always 'Role')",
							Computed:            true,
						},
						"add_services": schema.BoolAttribute{
							MarkdownDescription: "Whether role adds services to node",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *CMDeviceRolesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.ConfigureDataSource(req, resp)
}

func (d *CMDeviceRolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CMDeviceRolesDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Call BCM API to get nodes (roles are embedded in node objects)
	tflog.Debug(ctx, "Calling cmdevice.getNodes to extract roles")
	result, err := d.Client.CallJSONRPC(ctx, "cmdevice", "getNodes")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Roles",
			fmt.Sprintf("Could not query nodes to extract roles: %s", err.Error()),
		)
		return
	}

	// Parse JSON response
	var nodes []map[string]interface{}
	if err := json.Unmarshal(result, &nodes); err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Nodes",
			fmt.Sprintf("Could not parse nodes response: %s", err.Error()),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Retrieved %d nodes from BCM API", len(nodes)))

	// Extract and deduplicate roles by UUID using map
	roleMap := make(map[string]map[string]interface{})
	for _, node := range nodes {
		if rolesData, ok := node["roles"].([]interface{}); ok {
			for _, roleData := range rolesData {
				if role, ok := roleData.(map[string]interface{}); ok {
					uuid := getStringValue(role, "uuid")
					if !uuid.IsNull() && uuid.ValueString() != "" {
						roleMap[uuid.ValueString()] = role
					}
				}
			}
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("Found %d unique roles across %d nodes", len(roleMap), len(nodes)))

	// Validate name_pattern if provided
	if !data.NamePattern.IsNull() && !data.NamePattern.IsUnknown() {
		pattern := data.NamePattern.ValueString()
		// Validate glob pattern syntax by testing with empty string
		_, err := filepath.Match(pattern, "")
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid name_pattern",
				fmt.Sprintf("Glob pattern syntax error: %s", err.Error()),
			)
			return
		}
	}

	// Convert to slice and apply filters
	data.Roles = make([]RolesDataSourceRoleModel, 0, len(roleMap))
	for _, roleData := range roleMap {
		role := RolesDataSourceRoleModel{
			UUID:        getStringValue(roleData, "uuid"),
			Name:        getStringValue(roleData, "name"),
			ChildType:   getStringValue(roleData, "childType"),
			BaseType:    getStringValue(roleData, "baseType"),
			AddServices: getBoolValue(roleData, "addServices"),
		}
		role.ID = role.UUID

		// Apply filters (AND logic - all specified filters must match)
		if matchesRolesDataSourceFilter(role, data.NamePattern, data.ChildType) {
			data.Roles = append(data.Roles, role)
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("Returning %d roles after filtering", len(data.Roles)))

	// Set data source ID
	data.ID = types.StringValue("roles")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// matchesRolesDataSourceFilter applies AND logic for multiple filters.
// All specified filters must match for the role to be included.
//
// Filter behavior:
//   - name_pattern: Glob pattern matching using filepath.Match (supports *, ?, [abc])
//   - child_type: Exact case-sensitive matching (e.g., "HeadNodeRole" matches only roles with that exact type)
//   - Omitted filters are ignored (do not restrict results)
//
// Returns true if the role matches all specified filters, false otherwise.
func matchesRolesDataSourceFilter(role RolesDataSourceRoleModel, namePattern, childType types.String) bool {
	// name_pattern check (glob matching)
	if !namePattern.IsNull() && !namePattern.IsUnknown() {
		pattern := namePattern.ValueString()
		roleName := role.Name.ValueString()

		matched, err := filepath.Match(pattern, roleName)
		if err != nil {
			// Invalid glob pattern - should have been validated earlier
			return false
		}
		if !matched {
			return false
		}
	}

	// child_type check (exact match)
	if !childType.IsNull() && !childType.IsUnknown() {
		if role.ChildType.ValueString() != childType.ValueString() {
			return false
		}
	}

	return true
}
