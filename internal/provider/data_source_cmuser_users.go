// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &CMUserUsersDataSource{}

func NewCMUserUsersDataSource() datasource.DataSource {
	return &CMUserUsersDataSource{}
}

// CMUserUsersDataSource defines the data source implementation.
type CMUserUsersDataSource struct {
	BCMDataSourceBase
}

// CMUserUsersDataSourceModel describes the data source data model.
type CMUserUsersDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	UsernamePattern types.String `tfsdk:"username_pattern"`
	GroupID         types.String `tfsdk:"group_id"`
	UserID          types.String `tfsdk:"user_id"`
	Users           []UserModel  `tfsdk:"users"`
}

// UserModel describes a single user.
type UserModel struct {
	UUID              types.String `tfsdk:"uuid"`
	Username          types.String `tfsdk:"username"`
	UserID            types.String `tfsdk:"user_id"`
	GroupID           types.String `tfsdk:"group_id"`
	Email             types.String `tfsdk:"email"`
	CommonName        types.String `tfsdk:"common_name"`
	Surname           types.String `tfsdk:"surname"`
	HomeDirectory     types.String `tfsdk:"home_directory"`
	LoginShell        types.String `tfsdk:"login_shell"`
	Notes             types.String `tfsdk:"notes"`
	Information       types.String `tfsdk:"information"`
	AuthorizedSSHKeys types.String `tfsdk:"authorized_ssh_keys"`
	ShadowExpire      types.Int64  `tfsdk:"shadow_expire"`
	ShadowLastChange  types.Int64  `tfsdk:"shadow_last_change"`
	ShadowMax         types.Int64  `tfsdk:"shadow_max"`
	ShadowMin         types.Int64  `tfsdk:"shadow_min"`
	ShadowWarning     types.Int64  `tfsdk:"shadow_warning"`
	ShadowInactive    types.Int64  `tfsdk:"shadow_inactive"`
	AccountActive     types.Bool   `tfsdk:"account_active"`
}

func (d *CMUserUsersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cmuser_users"
}

func (d *CMUserUsersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Queries BCM Unix user accounts via the CMUser service. Provides access to user identity, Unix environment attributes, and shadow password fields.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier",
				Computed:            true,
			},
			"username_pattern": schema.StringAttribute{
				MarkdownDescription: "Optional pattern to match username field (supports wildcards like 'admin*', '*backup', '*cms*'). Uses filepath.Match glob-style matching.",
				Optional:            true,
			},
			"group_id": schema.StringAttribute{
				MarkdownDescription: "Optional Unix group ID (GID) to filter users by (exact match)",
				Optional:            true,
			},
			"user_id": schema.StringAttribute{
				MarkdownDescription: "Optional Unix user ID (UID) to filter users by (exact match)",
				Optional:            true,
			},
			"users": schema.ListNestedAttribute{
				MarkdownDescription: "List of users matching filter criteria",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"uuid": schema.StringAttribute{
							MarkdownDescription: "Unique identifier for the user",
							Computed:            true,
						},
						"username": schema.StringAttribute{
							MarkdownDescription: "User login name",
							Computed:            true,
						},
						"user_id": schema.StringAttribute{
							MarkdownDescription: "Unix user ID (UID)",
							Computed:            true,
						},
						"group_id": schema.StringAttribute{
							MarkdownDescription: "Unix group ID (GID)",
							Computed:            true,
						},
						"email": schema.StringAttribute{
							MarkdownDescription: "User email address",
							Computed:            true,
						},
						"common_name": schema.StringAttribute{
							MarkdownDescription: "User's common/display name",
							Computed:            true,
						},
						"surname": schema.StringAttribute{
							MarkdownDescription: "User's surname",
							Computed:            true,
						},
						"home_directory": schema.StringAttribute{
							MarkdownDescription: "Unix home directory path",
							Computed:            true,
						},
						"login_shell": schema.StringAttribute{
							MarkdownDescription: "Unix login shell (e.g., /bin/bash)",
							Computed:            true,
						},
						"notes": schema.StringAttribute{
							MarkdownDescription: "User notes",
							Computed:            true,
						},
						"information": schema.StringAttribute{
							MarkdownDescription: "Additional user information",
							Computed:            true,
						},
						"authorized_ssh_keys": schema.StringAttribute{
							MarkdownDescription: "SSH authorized public keys (may be multi-line)",
							Computed:            true,
						},
						"shadow_expire": schema.Int64Attribute{
							MarkdownDescription: "Account expiration date in days since Unix epoch (1970-01-01). Value of -1 means account never expires.",
							Computed:            true,
						},
						"shadow_last_change": schema.Int64Attribute{
							MarkdownDescription: "Last password change date in days since Unix epoch",
							Computed:            true,
						},
						"shadow_max": schema.Int64Attribute{
							MarkdownDescription: "Maximum password age in days",
							Computed:            true,
						},
						"shadow_min": schema.Int64Attribute{
							MarkdownDescription: "Minimum password age in days",
							Computed:            true,
						},
						"shadow_warning": schema.Int64Attribute{
							MarkdownDescription: "Password expiration warning period in days",
							Computed:            true,
						},
						"shadow_inactive": schema.Int64Attribute{
							MarkdownDescription: "Account inactive grace period in days after password expiration",
							Computed:            true,
						},
						"account_active": schema.BoolAttribute{
							MarkdownDescription: "Whether the account is active (not expired). Computed from shadow_expire field: active if shadow_expire is -1 (never expires) or shadow_expire > current_epoch_day.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *CMUserUsersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.ConfigureDataSource(req, resp)
}

func (d *CMUserUsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CMUserUsersDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Call BCM API to get all users
	tflog.Debug(ctx, "Calling cmuser.getUsers()")
	body, err := d.Client.CallJSONRPC(ctx, "cmuser", "getUsers")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Users",
			fmt.Sprintf("Could not query BCM users via CMUser API: %s", err.Error()),
		)
		return
	}

	// Parse JSON response
	var usersData []map[string]interface{}
	if err := json.Unmarshal(body, &usersData); err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Users Response",
			fmt.Sprintf("Could not parse BCM API response: %s", err.Error()),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Retrieved %d users from BCM API", len(usersData)))

	// Map and filter users
	var filteredUsers []UserModel
	for _, userData := range usersData {
		user := mapUserAPIResponseToModel(userData)

		// Apply client-side filters (AND logic)
		if matchesFilters(user, data.UsernamePattern, data.GroupID, data.UserID) {
			filteredUsers = append(filteredUsers, user)
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("Filtered to %d users matching criteria", len(filteredUsers)))

	data.Users = filteredUsers

	// Generate deterministic ID
	data.ID = types.StringValue(fmt.Sprintf("users-%d", time.Now().Unix()))

	tflog.Trace(ctx, "read cmuser users data source", map[string]interface{}{
		"total_users":    len(usersData),
		"filtered_users": len(filteredUsers),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Helper Functions

// computeAccountActive determines if user account is active based on shadowExpire.
// Account is active if:
//   - shadowExpire is -1 (never expires), OR
//   - shadowExpire > current_epoch_day
func computeAccountActive(shadowExpire types.Int64) types.Bool {
	if shadowExpire.IsNull() {
		return types.BoolValue(false)
	}

	expireDays := shadowExpire.ValueInt64()
	if expireDays == -1 {
		return types.BoolValue(true) // Never expires
	}

	currentEpochDay := time.Now().Unix() / 86400
	return types.BoolValue(expireDays > currentEpochDay)
}

// matchesFilters checks if user matches all filter criteria (AND logic).
// Returns true if user matches all specified filters, false otherwise.
func matchesFilters(user UserModel, usernamePattern, groupIDFilter, userIDFilter types.String) bool {
	// Filter by username_pattern (glob-style using filepath.Match)
	if !usernamePattern.IsNull() {
		pattern := usernamePattern.ValueString()
		username := user.Username.ValueString()
		matched, err := filepath.Match(pattern, username)
		if err != nil || !matched {
			return false
		}
	}

	// Filter by group_id (exact match)
	if !groupIDFilter.IsNull() {
		if user.GroupID.ValueString() != groupIDFilter.ValueString() {
			return false
		}
	}

	// Filter by user_id (exact match)
	if !userIDFilter.IsNull() {
		if user.UserID.ValueString() != userIDFilter.ValueString() {
			return false
		}
	}

	return true // All filters passed
}

// mapUserAPIResponseToModel converts BCM API user data to UserModel.
// Maps BCM API field names (camelCase) to Terraform attributes (snake_case):
//   - name → username
//   - ID → user_id
//   - groupID → group_id
func mapUserAPIResponseToModel(userData map[string]interface{}) UserModel {
	user := UserModel{
		UUID:              getStringValue(userData, "uuid"),
		Username:          getStringValue(userData, "name"), // BCM field is "name"
		UserID:            getStringValue(userData, "ID"),   // BCM field is "ID"
		GroupID:           getStringValue(userData, "groupID"),
		Email:             getStringValue(userData, "email"),
		CommonName:        getStringValue(userData, "commonName"),
		Surname:           getStringValue(userData, "surname"),
		HomeDirectory:     getStringValue(userData, "homeDirectory"),
		LoginShell:        getStringValue(userData, "loginShell"),
		Notes:             getStringValue(userData, "notes"),
		Information:       getStringValue(userData, "information"),
		AuthorizedSSHKeys: getStringValue(userData, "authorizedSshKeys"),
		ShadowExpire:      getInt64Value(userData, "shadowExpire"),
		ShadowLastChange:  getInt64Value(userData, "shadowLastChange"),
		ShadowMax:         getInt64Value(userData, "shadowMax"),
		ShadowMin:         getInt64Value(userData, "shadowMin"),
		ShadowWarning:     getInt64Value(userData, "shadowWarning"),
		ShadowInactive:    getInt64Value(userData, "shadowInactive"),
	}

	// Compute account_active from shadow_expire
	user.AccountActive = computeAccountActive(user.ShadowExpire)

	return user
}
