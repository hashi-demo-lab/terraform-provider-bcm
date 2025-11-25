// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &CMUserUserResource{}
	_ resource.ResourceWithImportState = &CMUserUserResource{}
)

// CMUserUserResource defines the resource implementation.
type CMUserUserResource struct {
	client *BCMClient
}

// CMUserUserResourceModel describes the resource data model.
type CMUserUserResourceModel struct {
	// Identity fields
	ID       types.String `tfsdk:"id"`
	UUID     types.String `tfsdk:"uuid"`
	Username types.String `tfsdk:"username"`

	// Authentication (write-only)
	Password types.String `tfsdk:"password"`

	// Unix identity
	UID           types.Int64  `tfsdk:"uid"`
	GID           types.Int64  `tfsdk:"gid"`
	HomeDirectory types.String `tfsdk:"home_directory"`
	Shell         types.String `tfsdk:"shell"`

	// Profile information
	FullName          types.String `tfsdk:"full_name"`
	Surname           types.String `tfsdk:"surname"`
	Email             types.String `tfsdk:"email"`
	Notes             types.String `tfsdk:"notes"`
	AuthorizedSSHKeys types.String `tfsdk:"authorized_ssh_keys"`

	// Shadow password attributes
	ShadowMax        types.Int64 `tfsdk:"shadow_max"`
	ShadowMin        types.Int64 `tfsdk:"shadow_min"`
	ShadowWarning    types.Int64 `tfsdk:"shadow_warning"`
	ShadowInactive   types.Int64 `tfsdk:"shadow_inactive"`
	ShadowExpire     types.Int64 `tfsdk:"shadow_expire"`
	ShadowLastChange types.Int64 `tfsdk:"shadow_last_change"`

	// Computed derived field
	AccountActive types.Bool `tfsdk:"account_active"`

	// Operation parameter
	Force types.Bool `tfsdk:"force"`
}

// NewCMUserUserResource creates a new resource instance.
func NewCMUserUserResource() resource.Resource {
	return &CMUserUserResource{}
}

// Metadata returns the resource type name.
func (r *CMUserUserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cmuser_user"
}

// Schema defines the resource schema.
func (r *CMUserUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a BCM Unix user account.\n\n" +
			"This resource enables full lifecycle management of BCM users including creation, " +
			"updating attributes, importing existing users, and deletion. Users are required " +
			"for DGX BasePOD automation workflows before Kubernetes cluster user setup.\n\n" +
			"**Note**: The `password` attribute is write-only and will never be returned by the BCM API. " +
			"The password is preserved in Terraform state but marked as sensitive.",

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
			"username": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "User login name (1-32 characters, must start with letter, alphanumeric and underscore only)",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 32),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`),
						"must start with a letter and contain only alphanumeric characters and underscores",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "User password (write-only, never returned by BCM API). Required for user creation.",
			},
			"uid": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Unix user ID (UID). If not specified, BCM auto-assigns the next available UID.",
				Validators: []validator.Int64{
					int64validator.Between(0, 65535),
				},
				PlanModifiers: []planmodifier.Int64{
					// UID should not change after creation
				},
			},
			"gid": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Primary group ID (GID). If not specified, BCM auto-assigns (typically creates a group with same name as user).",
				Validators: []validator.Int64{
					int64validator.Between(0, 65535),
				},
			},
			"home_directory": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "User home directory path. Defaults to `/home/{username}`.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^/.*`),
						"must be an absolute path starting with /",
					),
				},
			},
			"shell": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("/bin/bash"),
				MarkdownDescription: "Login shell path. Defaults to `/bin/bash`.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^/.*`),
						"must be an absolute path starting with /",
					),
				},
			},
			"full_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "User's full display name (maps to commonName in BCM). Defaults to username if not specified.",
			},
			"surname": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "User's surname (last name). Defaults to username if not specified.",
			},
			"email": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "User's email address",
			},
			"notes": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Administrative notes about the user",
			},
			"authorized_ssh_keys": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "SSH authorized public keys (may be multi-line for multiple keys)",
			},
			"shadow_max": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(99999),
				MarkdownDescription: "Maximum password age in days. Defaults to `99999` (effectively no expiration).",
			},
			"shadow_min": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				MarkdownDescription: "Minimum password age in days (time before password can be changed). Defaults to `0`.",
			},
			"shadow_warning": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(7),
				MarkdownDescription: "Password expiration warning period in days. Defaults to `7`.",
			},
			"shadow_inactive": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				MarkdownDescription: "Account inactive grace period in days after password expiration. Defaults to `0`.",
			},
			"shadow_expire": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Account expiration date in days since Unix epoch (1970-01-01). Value of -1 means account never expires.",
			},
			"shadow_last_change": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Last password change date in days since Unix epoch",
			},
			"account_active": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the account is active (not expired). Computed from shadow_expire field.",
			},
			"force": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Force operation even with validation warnings. Defaults to `false`.",
			},
		},
	}
}

// Configure stores the BCM client in the resource instance.
func (r *CMUserUserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// getNextAvailableUID queries existing users and returns the next available UID.
// Starts from 60000 to avoid conflicts with system users (0-999) and regular users (1000-59999).
func (r *CMUserUserResource) getNextAvailableUID(ctx context.Context) (int64, error) {
	body, err := r.client.CallJSONRPC(ctx, "cmuser", "getUsers")
	if err != nil {
		return 0, fmt.Errorf("failed to query existing users: %w", err)
	}

	var users []map[string]interface{}
	if err := json.Unmarshal(body, &users); err != nil {
		return 0, fmt.Errorf("failed to parse users response: %w", err)
	}

	// Find the maximum UID currently in use
	maxUID := int64(59999) // Start just below our range
	for _, user := range users {
		if idStr, ok := user["ID"].(string); ok && idStr != "" {
			var uid int64
			fmt.Sscanf(idStr, "%d", &uid)
			if uid > maxUID && uid < 65535 {
				maxUID = uid
			}
		}
	}

	// Return next available UID
	return maxUID + 1, nil
}

// Create implements the resource Create operation.
func (r *CMUserUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CMUserUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If UID not specified, auto-assign from 60000+ range
	if plan.UID.IsNull() || plan.UID.IsUnknown() {
		nextUID, err := r.getNextAvailableUID(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"UID Auto-Assignment Failed",
				fmt.Sprintf("Could not determine next available UID: %s", err.Error()),
			)
			return
		}
		plan.UID = types.Int64Value(nextUID)
		tflog.Debug(ctx, "Auto-assigned UID for new user", map[string]interface{}{
			"username": plan.Username.ValueString(),
			"uid":      nextUID,
		})
	}

	// If GID not specified, use the same as UID (create user's personal group)
	// Note: BCM requires the group to already exist. Using 1000 (cmsupport group) as fallback.
	if plan.GID.IsNull() || plan.GID.IsUnknown() {
		// Use group 1000 (cmsupport) as default since BCM requires existing group
		plan.GID = types.Int64Value(1000)
		tflog.Debug(ctx, "Using default GID for new user", map[string]interface{}{
			"username": plan.Username.ValueString(),
			"gid":      1000,
		})
	}

	// Build API entity from plan
	entity := r.buildAPIEntity(ctx, &plan, "")

	// Log the request entity for debugging
	entityJSON, _ := json.MarshalIndent(entity, "", "  ")
	tflog.Debug(ctx, "Creating user via BCM API with entity", map[string]interface{}{
		"username": plan.Username.ValueString(),
		"entity":   string(entityJSON),
	})

	// Get force parameter
	force := false
	if !plan.Force.IsNull() {
		force = plan.Force.ValueBool()
	}

	// Call addUser API
	createBody, err := r.client.CallJSONRPC(ctx, "cmuser", "addUser", entity, force)
	if err != nil {
		resp.Diagnostics.AddError(
			"User Creation Failed",
			fmt.Sprintf("Failed to create user '%s': %s", plan.Username.ValueString(), err.Error()),
		)
		return
	}

	// Parse UUID from response
	// BCM API returns: {"success": true, "updated_entity": {"uuid": "...", ...}, ...}
	var createdUUID string
	var respObj map[string]interface{}
	if err := json.Unmarshal(createBody, &respObj); err == nil {
		// Check for success response with updated_entity
		if success, ok := respObj["success"].(bool); ok && success {
			if updatedEntity, ok := respObj["updated_entity"].(map[string]interface{}); ok {
				if uuid, ok := updatedEntity["uuid"].(string); ok && uuid != "" {
					createdUUID = uuid
				}
			}
		} else if !success {
			// Check for validation errors
			var errorMsg string
			if validation, ok := respObj["validation"].([]interface{}); ok && len(validation) > 0 {
				for _, v := range validation {
					if valObj, ok := v.(map[string]interface{}); ok {
						field := valObj["field"]
						message := valObj["message"]
						severity := valObj["severity"]
						errorMsg += fmt.Sprintf("\n  - %s: %s (severity: %s)", field, message, severity)
					}
				}
			}
			resp.Diagnostics.AddError(
				"User Creation Failed",
				fmt.Sprintf("BCM API returned validation errors:%s", errorMsg),
			)
			return
		}
		// Try direct uuid field
		if createdUUID == "" {
			if uuid, ok := respObj["uuid"].(string); ok && uuid != "" {
				createdUUID = uuid
			}
		}
	} else {
		// Try parsing as simple string (UUID only)
		var uuidStr string
		if err := json.Unmarshal(createBody, &uuidStr); err == nil && uuidStr != "" {
			createdUUID = uuidStr
		}
	}

	if createdUUID == "" {
		resp.Diagnostics.AddError(
			"UUID Extraction Failed",
			fmt.Sprintf("Failed to extract UUID from create response: %s", string(createBody)),
		)
		return
	}

	// Set ID and UUID
	plan.ID = types.StringValue(createdUUID)
	plan.UUID = types.StringValue(createdUUID)

	// Wait for eventual consistency
	time.Sleep(1 * time.Second)

	// Read back created resource to populate computed fields
	// Preserve password from plan (BCM API returns empty string)
	planPassword := plan.Password
	r.readUser(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Restore password from plan (write-only field)
	plan.Password = planPassword

	tflog.Info(ctx, "Created user resource", map[string]interface{}{
		"username": plan.Username.ValueString(),
		"uuid":     createdUUID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read implements the resource Read operation.
func (r *CMUserUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CMUserUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve password from state (BCM API returns empty string)
	statePassword := state.Password

	// Fetch current state from BCM API
	r.readUser(ctx, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Restore password from prior state (write-only field)
	state.Password = statePassword

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update implements the resource Update operation.
func (r *CMUserUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CMUserUserResourceModel
	var state CMUserUserResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve UID and GID from state (they are immutable after creation)
	if plan.UID.IsNull() || plan.UID.IsUnknown() {
		plan.UID = state.UID
	}
	if plan.GID.IsNull() || plan.GID.IsUnknown() {
		plan.GID = state.GID
	}

	// Build updated entity from plan (include UUID for update)
	entity := r.buildAPIEntity(ctx, &plan, state.UUID.ValueString())

	// Handle password update - only include if changed
	// Compare plan password with state password
	if !plan.Password.Equal(state.Password) {
		entity["password"] = plan.Password.ValueString()
		tflog.Debug(ctx, "Password change detected, including in update", map[string]interface{}{
			"username": plan.Username.ValueString(),
		})
	}

	// Log the update entity for debugging
	entityJSON, _ := json.MarshalIndent(entity, "", "  ")
	tflog.Debug(ctx, "Updating user via BCM API with entity", map[string]interface{}{
		"username": plan.Username.ValueString(),
		"uuid":     state.UUID.ValueString(),
		"entity":   string(entityJSON),
	})

	// Get force parameter
	force := false
	if !plan.Force.IsNull() {
		force = plan.Force.ValueBool()
	}

	// Call updateUser API
	_, err := r.client.CallJSONRPC(ctx, "cmuser", "updateUser", entity, force)
	if err != nil {
		resp.Diagnostics.AddError(
			"User Update Failed",
			fmt.Sprintf("Failed to update user '%s': %s", plan.Username.ValueString(), err.Error()),
		)
		return
	}

	// Wait for eventual consistency
	time.Sleep(1 * time.Second)

	// Read back updated resource to verify changes
	// Preserve password from plan
	planPassword := plan.Password
	r.readUser(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Restore password from plan (write-only field)
	plan.Password = planPassword

	tflog.Info(ctx, "Updated user resource", map[string]interface{}{
		"username": plan.Username.ValueString(),
		"uuid":     plan.UUID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete implements the resource Delete operation.
func (r *CMUserUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CMUserUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting user via BCM API", map[string]interface{}{
		"username": state.Username.ValueString(),
		"uuid":     state.UUID.ValueString(),
	})

	// Call removeUser with UUID (BCM requires UUID, not username)
	_, err := r.client.CallJSONRPC(ctx, "cmuser", "removeUser", state.UUID.ValueString())
	if err != nil {
		// Enhanced error handling - check if already deleted (idempotent)
		errStr := err.Error()

		// Check if user was already deleted externally (idempotent delete)
		if containsAny(errStr, []string{"not found", "does not exist", "No such", "unknown user"}) {
			tflog.Info(ctx, "User already deleted (idempotent)", map[string]interface{}{
				"username": state.Username.ValueString(),
				"uuid":     state.UUID.ValueString(),
			})
			return
		}

		// Other deletion errors
		resp.Diagnostics.AddError(
			"User Deletion Failed",
			fmt.Sprintf("Failed to delete user '%s': %s", state.Username.ValueString(), err.Error()),
		)
		return
	}

	tflog.Info(ctx, "Deleted user resource", map[string]interface{}{
		"username": state.Username.ValueString(),
		"uuid":     state.UUID.ValueString(),
	})
}

// ImportState implements resource import using username as the import identifier.
func (r *CMUserUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID is the username
	tflog.Debug(ctx, "Importing user by username", map[string]interface{}{
		"username": req.ID,
	})

	// Set username attribute - Read() will populate the rest
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("username"), req.ID)...)

	// Also need to set a placeholder password since it's required but we can't recover it
	// After import, user should update their config with the actual password
	resp.Diagnostics.AddWarning(
		"Password Required After Import",
		"The user password cannot be recovered from BCM API. Please update your Terraform configuration "+
			"with the correct password value to prevent drift on subsequent applies.",
	)
}

// Helper Methods

// readUser fetches user from API and populates model.
func (r *CMUserUserResource) readUser(ctx context.Context, model *CMUserUserResourceModel, diags *diag.Diagnostics) {
	// Determine lookup identifier
	var lookupName string
	if !model.Username.IsNull() && model.Username.ValueString() != "" {
		lookupName = model.Username.ValueString()
	} else if !model.ID.IsNull() && model.ID.ValueString() != "" {
		// Import case: need to look up by UUID to find username
		allUsersBody, err := r.client.CallJSONRPC(ctx, "cmuser", "getUsers")
		if err != nil {
			diags.AddError(
				"User Read Failed",
				fmt.Sprintf("Failed to list users during import: %s", err.Error()),
			)
			return
		}

		var usersList []map[string]interface{}
		if err := json.Unmarshal(allUsersBody, &usersList); err != nil {
			diags.AddError(
				"Response Parse Error",
				fmt.Sprintf("Failed to parse users list: %s", err.Error()),
			)
			return
		}

		// Find user with matching UUID
		targetUUID := model.ID.ValueString()
		for _, user := range usersList {
			if uuid, ok := user["uuid"].(string); ok && uuid == targetUUID {
				if name, ok := user["name"].(string); ok {
					lookupName = name
					break
				}
			}
		}

		if lookupName == "" {
			diags.AddError(
				"User Not Found",
				fmt.Sprintf("User with ID '%s' not found in BCM", targetUUID),
			)
			return
		}
	} else {
		diags.AddError(
			"Invalid State",
			"Cannot read user: both username and ID are empty",
		)
		return
	}

	tflog.Debug(ctx, "Reading user via BCM API", map[string]interface{}{
		"username": lookupName,
	})

	// Use getUser(username) for direct lookup
	body, err := r.client.CallJSONRPC(ctx, "cmuser", "getUser", lookupName)
	if err != nil {
		diags.AddError(
			"User Read Failed",
			fmt.Sprintf("Failed to read user '%s': %s", lookupName, err.Error()),
		)
		return
	}

	// Parse response
	var userData map[string]interface{}
	if err := json.Unmarshal(body, &userData); err != nil {
		diags.AddError(
			"Response Parse Error",
			fmt.Sprintf("Failed to parse user response: %s", err.Error()),
		)
		return
	}

	// Check if user not found
	if len(userData) == 0 {
		diags.AddError(
			"User Not Found",
			fmt.Sprintf("User '%s' not found in BCM", lookupName),
		)
		return
	}

	// Map API fields to model
	model.UUID = getStringValue(userData, "uuid")
	model.Username = getStringValue(userData, "name")

	// Unix identity - BCM returns ID/groupID as strings
	if idStr, ok := userData["ID"].(string); ok && idStr != "" {
		var uid int64
		fmt.Sscanf(idStr, "%d", &uid)
		model.UID = types.Int64Value(uid)
	} else {
		model.UID = types.Int64Null()
	}

	if gidStr, ok := userData["groupID"].(string); ok && gidStr != "" {
		var gid int64
		fmt.Sscanf(gidStr, "%d", &gid)
		model.GID = types.Int64Value(gid)
	} else {
		model.GID = types.Int64Null()
	}

	model.HomeDirectory = getStringValue(userData, "homeDirectory")
	model.Shell = getStringValue(userData, "loginShell")

	// Profile information
	model.FullName = getStringValue(userData, "commonName")
	model.Surname = getStringValue(userData, "surname")
	model.Email = getStringValue(userData, "email")
	model.Notes = getStringValue(userData, "notes")
	model.AuthorizedSSHKeys = getStringValue(userData, "authorizedSshKeys")

	// Shadow password attributes
	model.ShadowMax = getInt64Value(userData, "shadowMax")
	model.ShadowMin = getInt64Value(userData, "shadowMin")
	model.ShadowWarning = getInt64Value(userData, "shadowWarning")
	model.ShadowInactive = getInt64Value(userData, "shadowInactive")
	model.ShadowExpire = getInt64Value(userData, "shadowExpire")
	model.ShadowLastChange = getInt64Value(userData, "shadowLastChange")

	// Compute account_active from shadow_expire
	model.AccountActive = computeAccountActive(model.ShadowExpire)

	// Ensure ID is set
	if model.ID.IsNull() && !model.UUID.IsNull() {
		model.ID = model.UUID
	}

	// Note: Password is NOT read from API (always empty)
	// It will be preserved from state/plan by the calling function

	tflog.Trace(ctx, "Successfully read user", map[string]interface{}{
		"username": model.Username.ValueString(),
		"uuid":     model.UUID.ValueString(),
	})
}

// buildAPIEntity constructs BCM API entity from Terraform model.
// If uuid is provided, this is an update operation, otherwise it's a create.
func (r *CMUserUserResource) buildAPIEntity(ctx context.Context, model *CMUserUserResourceModel, uuid string) map[string]interface{} {
	entity := map[string]interface{}{
		"baseType":      "User",
		"childType":     "",
		"modified":      true,
		"to_be_removed": false,
		"revision":      "",
	}

	// UUID handling
	if uuid != "" {
		entity["uuid"] = uuid
	} else {
		// Generate new UUID for create
		entity["uuid"] = generateUUID()
	}

	// Username (required)
	if !model.Username.IsNull() {
		entity["name"] = model.Username.ValueString()
	}

	// Password - only include for create (update handles separately)
	if uuid == "" && !model.Password.IsNull() {
		entity["password"] = model.Password.ValueString()
	}

	// Unix identity - BCM requires ID and groupID fields
	// UID and GID should be set by Create() before calling this method
	if !model.UID.IsNull() && !model.UID.IsUnknown() {
		entity["ID"] = fmt.Sprintf("%d", model.UID.ValueInt64())
	}

	if !model.GID.IsNull() && !model.GID.IsUnknown() {
		entity["groupID"] = fmt.Sprintf("%d", model.GID.ValueInt64())
	}

	// Home directory
	if !model.HomeDirectory.IsNull() && model.HomeDirectory.ValueString() != "" {
		entity["homeDirectory"] = model.HomeDirectory.ValueString()
	} else if !model.Username.IsNull() {
		// Default to /home/{username}
		entity["homeDirectory"] = fmt.Sprintf("/home/%s", model.Username.ValueString())
	}

	// Shell
	if !model.Shell.IsNull() {
		entity["loginShell"] = model.Shell.ValueString()
	}

	// Profile information - BCM requires commonName and surname
	if !model.FullName.IsNull() && model.FullName.ValueString() != "" {
		entity["commonName"] = model.FullName.ValueString()
	} else if !model.Username.IsNull() {
		// Default to username if not specified
		entity["commonName"] = model.Username.ValueString()
	}
	if !model.Surname.IsNull() && model.Surname.ValueString() != "" {
		entity["surname"] = model.Surname.ValueString()
	} else if !model.Username.IsNull() {
		// Default to username if not specified
		entity["surname"] = model.Username.ValueString()
	}
	if !model.Email.IsNull() {
		entity["email"] = model.Email.ValueString()
	}
	if !model.Notes.IsNull() {
		entity["notes"] = model.Notes.ValueString()
	}
	if !model.AuthorizedSSHKeys.IsNull() {
		entity["authorizedSshKeys"] = model.AuthorizedSSHKeys.ValueString()
	}

	// Shadow password attributes
	if !model.ShadowMax.IsNull() {
		entity["shadowMax"] = model.ShadowMax.ValueInt64()
	}
	if !model.ShadowMin.IsNull() {
		entity["shadowMin"] = model.ShadowMin.ValueInt64()
	}
	if !model.ShadowWarning.IsNull() {
		entity["shadowWarning"] = model.ShadowWarning.ValueInt64()
	}
	if !model.ShadowInactive.IsNull() {
		entity["shadowInactive"] = model.ShadowInactive.ValueInt64()
	}

	// BCM operational flags
	entity["homeDirOperation"] = true
	entity["createSshKey"] = false

	return entity
}
