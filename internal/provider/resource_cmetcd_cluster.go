// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Package provider implements the bcm_cmetcd_cluster resource for managing BCM EtcdCluster entities.
// This resource enables Terraform to manage etcd cluster definitions in BCM.
package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider-defined types implement resource interfaces.
var (
	_ resource.Resource                = &CMEtcdClusterResource{}
	_ resource.ResourceWithImportState = &CMEtcdClusterResource{}
)

// CMEtcdClusterResource defines the resource implementation.
type CMEtcdClusterResource struct {
	client *BCMClient
}

// NewCMEtcdClusterResource creates a new CMEtcdClusterResource.
func NewCMEtcdClusterResource() resource.Resource {
	return &CMEtcdClusterResource{}
}

// Metadata returns the resource type name.
func (r *CMEtcdClusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cmetcd_cluster"
}

// Schema defines the resource schema (T014).
func (r *CMEtcdClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages a BCM EtcdCluster entity. EtcdCluster defines the etcd cluster configuration for Kubernetes deployments.",
		MarkdownDescription: "Manages a BCM EtcdCluster entity.\n\nEtcdCluster defines the etcd cluster configuration that backs Kubernetes clusters in BCM.",

		Attributes: map[string]schema.Attribute{
			// Identifier attributes
			"id": schema.StringAttribute{
				Description: "Resource identifier (UUID).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uuid": schema.StringAttribute{
				Description: "BCM-assigned UUID for the EtcdCluster entity.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Required attributes
			"name": schema.StringAttribute{
				Description:         "Etcd cluster name (RFC 1123 DNS label: lowercase alphanumeric and hyphens, 1-63 characters).",
				MarkdownDescription: "Etcd cluster name (RFC 1123 DNS label: lowercase alphanumeric and hyphens, 1-63 characters).",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`),
						"must contain only lowercase alphanumeric characters and hyphens, must start and end with alphanumeric",
					),
				},
			},

			// Optional attributes with defaults
			"heartbeat_interval": schema.Int64Attribute{
				Description:         "Etcd heartbeat interval in milliseconds (default: 100, range: 50-500).",
				MarkdownDescription: "Etcd heartbeat interval in milliseconds (default: 100, range: 50-500).",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(100),
				Validators: []validator.Int64{
					int64validator.Between(50, 500),
				},
			},
			"election_timeout": schema.Int64Attribute{
				Description:         "Etcd election timeout in milliseconds (default: 1000, range: 500-5000). Should be at least 5x heartbeat_interval.",
				MarkdownDescription: "Etcd election timeout in milliseconds (default: 1000, range: 500-5000). Should be at least 5x heartbeat_interval.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1000),
				Validators: []validator.Int64{
					int64validator.Between(500, 5000),
				},
			},
			"options": schema.StringAttribute{
				Description:         "Extensible configuration options as JSON string.",
				MarkdownDescription: "Extensible configuration options as JSON string. Example: `{\"custom_key\": \"value\"}`",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("{}"),
			},

			// Computed attributes
			"creation_time": schema.Int64Attribute{
				Description: "Unix timestamp of when the EtcdCluster was created.",
				Computed:    true,
			},
			"revision_id": schema.Int64Attribute{
				Description: "Revision number for change tracking.",
				Computed:    true,
			},
		},
	}
}

// Configure sets up the resource with the provider client.
func (r *CMEtcdClusterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if provider is not configured
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

// Create implements resource.Resource (T015).
func (r *CMEtcdClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EtcdClusterResourceModel

	// Read Terraform plan data into model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Nil client check
	if r.client == nil {
		resp.Diagnostics.AddError("Client Not Configured", "The BCM client is not configured. Please configure the provider.")
		return
	}

	// Build entity
	entity := r.buildEntity(ctx, &data)

	// Pre-flight validation
	validationErrors, err := r.client.ValidateEtcdCluster(ctx, entity, true)
	if err != nil {
		resp.Diagnostics.AddError("Validation API Failed", fmt.Sprintf("Failed to validate EtcdCluster: %s", err))
		return
	}
	if ProcessValidationErrors(validationErrors, &resp.Diagnostics) {
		return
	}

	// Create via BCM API
	tflog.Debug(ctx, "Creating EtcdCluster", map[string]interface{}{
		"name": data.Name.ValueString(),
	})

	body, err := r.client.AddEtcdCluster(ctx, entity)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", fmt.Sprintf("Failed to create EtcdCluster: %s", err))
		return
	}

	// Parse response - BCM returns validation response: {"success": true/false, "validation": [...]}
	var validationResp struct {
		Success    bool                     `json:"success"`
		Validation []map[string]interface{} `json:"validation"`
	}

	if err := json.Unmarshal(body, &validationResp); err != nil {
		resp.Diagnostics.AddError(
			"Response Parse Error",
			fmt.Sprintf("Failed to parse EtcdCluster creation response: %s", err),
		)
		return
	}

	// Check validation response
	if !validationResp.Success {
		var errorMsgs []string
		for _, v := range validationResp.Validation {
			if field, ok := v["field"].(string); ok {
				if msg, ok := v["message"].(string); ok {
					errorMsgs = append(errorMsgs, fmt.Sprintf("%s: %s", field, msg))
				}
			}
		}
		resp.Diagnostics.AddError(
			"EtcdCluster Creation Failed",
			fmt.Sprintf("Failed to create EtcdCluster '%s': validation errors: %v", data.Name.ValueString(), errorMsgs),
		)
		return
	}

	// BCM doesn't return the entity on create - use UUID from entity we sent
	createdUUID, ok := entity["uuid"].(string)
	if !ok {
		resp.Diagnostics.AddError(
			"Failed to extract UUID from created EtcdCluster",
			"The UUID was not found in the created entity",
		)
		return
	}

	// Set UUID in model for read
	data.ID = types.StringValue(createdUUID)
	data.UUID = types.StringValue(createdUUID)

	tflog.Debug(ctx, "EtcdCluster created successfully", map[string]interface{}{
		"name": data.Name.ValueString(),
		"uuid": createdUUID,
	})

	// Preserve plan value for options (BCM doesn't persist options field)
	planOptions := data.Options

	// Read back created entity to populate all fields with eventual consistency handling
	maxRetries := 5
	var lastReadErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		readBody, err := r.client.GetEtcdCluster(ctx, createdUUID)
		if err != nil {
			lastReadErr = err
			if attempt < maxRetries-1 {
				sleepDuration := time.Duration(1<<attempt) * time.Second
				tflog.Warn(ctx, "EtcdCluster read after create failed, retrying", map[string]interface{}{
					"attempt":       attempt + 1,
					"sleep_seconds": sleepDuration.Seconds(),
					"error":         err.Error(),
				})
				time.Sleep(sleepDuration)
				continue
			}
			resp.Diagnostics.AddError(
				"Read After Create Failed",
				fmt.Sprintf("Failed to read EtcdCluster after create: %s", lastReadErr),
			)
			return
		}

		// Parse the read response
		var responseData map[string]interface{}
		if err := json.Unmarshal(readBody, &responseData); err != nil {
			if attempt < maxRetries-1 {
				sleepDuration := time.Duration(1<<attempt) * time.Second
				tflog.Warn(ctx, "EtcdCluster response parse failed, retrying", map[string]interface{}{
					"attempt":       attempt + 1,
					"sleep_seconds": sleepDuration.Seconds(),
				})
				time.Sleep(sleepDuration)
				continue
			}
			resp.Diagnostics.AddError(
				"Response Parse Failed",
				fmt.Sprintf("Failed to parse EtcdCluster response: %s", err),
			)
			return
		}

		// Check if response has expected fields
		if responseData["name"] != nil && responseData["uuid"] != nil {
			r.parseResponseIntoModel(ctx, responseData, &data)
			tflog.Debug(ctx, "Successfully read EtcdCluster after create", map[string]interface{}{
				"uuid":    data.UUID.ValueString(),
				"name":    data.Name.ValueString(),
				"attempt": attempt + 1,
			})
			break
		}

		if attempt < maxRetries-1 {
			sleepDuration := time.Duration(1<<attempt) * time.Second
			tflog.Warn(ctx, "EtcdCluster fields not populated, retrying", map[string]interface{}{
				"attempt":       attempt + 1,
				"sleep_seconds": sleepDuration.Seconds(),
			})
			time.Sleep(sleepDuration)
		} else {
			// Final attempt failed - error out instead of saving incomplete state
			resp.Diagnostics.AddError(
				"Read After Create Incomplete",
				"EtcdCluster was created but read-back returned incomplete data after all retries",
			)
			return
		}
	}

	// Restore plan value for options (BCM quirk: accepts but doesn't persist options field)
	if !planOptions.IsNull() && !planOptions.IsUnknown() {
		data.Options = planOptions
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read implements resource.Resource (T016).
func (r *CMEtcdClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EtcdClusterResourceModel

	// Read Terraform state data into model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Nil client check
	if r.client == nil {
		resp.Diagnostics.AddError("Client Not Configured", "The BCM client is not configured. Please configure the provider.")
		return
	}

	// Preserve state value for options (BCM quirk: accepts but doesn't persist options field)
	stateOptions := data.Options

	// Get from BCM API using UUID
	identifier := data.UUID.ValueString()
	if identifier == "" {
		identifier = data.ID.ValueString()
	}

	tflog.Debug(ctx, "Reading EtcdCluster", map[string]interface{}{
		"id": identifier,
	})

	body, err := r.client.GetEtcdCluster(ctx, identifier)
	if err != nil {
		// Check if resource no longer exists
		if containsAny(err.Error(), []string{"not found", "does not exist", "404", "null"}) {
			tflog.Info(ctx, "EtcdCluster not found, removing from state", map[string]interface{}{
				"id": identifier,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("Failed to read EtcdCluster: %s", err))
		return
	}

	// Parse response
	var responseData map[string]interface{}
	if err := json.Unmarshal(body, &responseData); err != nil {
		resp.Diagnostics.AddError("Parse Failed", fmt.Sprintf("Failed to parse EtcdCluster response: %s", err))
		return
	}

	// Check if response is empty (deleted)
	if len(responseData) == 0 {
		tflog.Info(ctx, "EtcdCluster returned empty response, removing from state", map[string]interface{}{
			"id": identifier,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	// Update model from response
	r.parseResponseIntoModel(ctx, responseData, &data)

	// Restore state value for options (BCM quirk: accepts but doesn't persist options field)
	if !stateOptions.IsNull() && !stateOptions.IsUnknown() {
		data.Options = stateOptions
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update implements resource.Resource (T017).
func (r *CMEtcdClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data EtcdClusterResourceModel

	// Read Terraform plan data into model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Nil client check
	if r.client == nil {
		resp.Diagnostics.AddError("Client Not Configured", "The BCM client is not configured. Please configure the provider.")
		return
	}

	// Preserve plan value for options (BCM quirk: accepts but doesn't persist options field)
	planOptions := data.Options

	// Get current state to preserve UUID
	var stateData EtcdClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve UUID from state
	data.UUID = stateData.UUID
	data.ID = stateData.ID

	// Build entity with existing UUID
	entity := r.buildEntity(ctx, &data)
	entity["uuid"] = data.UUID.ValueString()

	// Pre-flight validation
	validationErrors, err := r.client.ValidateEtcdCluster(ctx, entity, false)
	if err != nil {
		resp.Diagnostics.AddError("Validation API Failed", fmt.Sprintf("Failed to validate EtcdCluster: %s", err))
		return
	}
	if ProcessValidationErrors(validationErrors, &resp.Diagnostics) {
		return
	}

	tflog.Debug(ctx, "Updating EtcdCluster", map[string]interface{}{
		"id":   data.UUID.ValueString(),
		"name": data.Name.ValueString(),
	})

	// Update via BCM API
	body, err := r.client.UpdateEtcdCluster(ctx, entity)
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", fmt.Sprintf("Failed to update EtcdCluster: %s", err))
		return
	}

	// Parse response - BCM returns validation response: {"success": true/false, "validation": [...]}
	var validationResp struct {
		Success    bool                     `json:"success"`
		Validation []map[string]interface{} `json:"validation"`
	}

	if err := json.Unmarshal(body, &validationResp); err != nil {
		resp.Diagnostics.AddError(
			"Response Parse Error",
			fmt.Sprintf("Failed to parse EtcdCluster update response: %s", err),
		)
		return
	}

	// Check validation response
	if !validationResp.Success {
		var errorMsgs []string
		for _, v := range validationResp.Validation {
			if field, ok := v["field"].(string); ok {
				if msg, ok := v["message"].(string); ok {
					errorMsgs = append(errorMsgs, fmt.Sprintf("%s: %s", field, msg))
				}
			}
		}
		resp.Diagnostics.AddError(
			"EtcdCluster Update Failed",
			fmt.Sprintf("Failed to update EtcdCluster '%s': validation errors: %v", data.Name.ValueString(), errorMsgs),
		)
		return
	}

	// Read back updated entity to populate all fields with eventual consistency handling
	maxRetries := 5
	var lastReadErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		readBody, err := r.client.GetEtcdCluster(ctx, data.UUID.ValueString())
		if err != nil {
			lastReadErr = err
			if attempt < maxRetries-1 {
				sleepDuration := time.Duration(1<<attempt) * time.Second
				tflog.Warn(ctx, "EtcdCluster read after update failed, retrying", map[string]interface{}{
					"attempt":       attempt + 1,
					"sleep_seconds": sleepDuration.Seconds(),
					"error":         err.Error(),
				})
				time.Sleep(sleepDuration)
				continue
			}
			resp.Diagnostics.AddError(
				"Read After Update Failed",
				fmt.Sprintf("Failed to read EtcdCluster after update: %s", lastReadErr),
			)
			return
		}

		var responseData map[string]interface{}
		if err := json.Unmarshal(readBody, &responseData); err != nil {
			if attempt < maxRetries-1 {
				sleepDuration := time.Duration(1<<attempt) * time.Second
				tflog.Warn(ctx, "EtcdCluster response parse after update failed, retrying", map[string]interface{}{
					"attempt":       attempt + 1,
					"sleep_seconds": sleepDuration.Seconds(),
				})
				time.Sleep(sleepDuration)
				continue
			}
			resp.Diagnostics.AddError(
				"Response Parse Failed",
				fmt.Sprintf("Failed to parse EtcdCluster response: %s", err),
			)
			return
		}

		r.parseResponseIntoModel(ctx, responseData, &data)

		tflog.Debug(ctx, "Successfully read EtcdCluster after update", map[string]interface{}{
			"uuid":    data.UUID.ValueString(),
			"name":    data.Name.ValueString(),
			"attempt": attempt + 1,
		})
		break
	}

	// Restore plan value for options (BCM quirk: accepts but doesn't persist options field)
	if !planOptions.IsNull() && !planOptions.IsUnknown() {
		data.Options = planOptions
	}

	tflog.Debug(ctx, "Updated EtcdCluster", map[string]interface{}{
		"id":   data.ID.ValueString(),
		"name": data.Name.ValueString(),
	})

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete implements resource.Resource (T018).
func (r *CMEtcdClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EtcdClusterResourceModel

	// Read Terraform state data into model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Nil client check
	if r.client == nil {
		resp.Diagnostics.AddError("Client Not Configured", "The BCM client is not configured. Please configure the provider.")
		return
	}

	uuid := data.UUID.ValueString()
	if uuid == "" {
		uuid = data.ID.ValueString()
	}

	tflog.Debug(ctx, "Deleting EtcdCluster", map[string]interface{}{
		"id":   uuid,
		"name": data.Name.ValueString(),
	})

	// Delete via BCM API
	_, err := r.client.RemoveEtcdCluster(ctx, uuid)
	if err != nil {
		// BCM can report missing entities as a failed task payload instead of a transport error.
		if !containsAny(err.Error(), []string{"not found", "does not exist", "404", "No such", "no such"}) {
			resp.Diagnostics.AddError("Delete Failed", fmt.Sprintf("Failed to delete EtcdCluster: %s", err))
			return
		}
		tflog.Info(ctx, "EtcdCluster already deleted", map[string]interface{}{
			"id": uuid,
		})
	}

	tflog.Debug(ctx, "Deleted EtcdCluster", map[string]interface{}{
		"id": uuid,
	})
}

// ImportState implements resource.ResourceWithImportState (T019).
func (r *CMEtcdClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by UUID
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// =============================================================================
// Helper Methods
// =============================================================================

// buildEntity constructs a BCM EtcdCluster entity from Terraform model.
func (r *CMEtcdClusterResource) buildEntity(_ context.Context, data *EtcdClusterResourceModel) map[string]interface{} {
	entity := map[string]interface{}{
		"baseType":      "EtcdCluster",
		"childType":     "",
		"modified":      true,
		"to_be_removed": false,
		"revision":      "",
	}

	// Set name
	entity["name"] = data.Name.ValueString()

	// Set or generate UUID
	if !data.UUID.IsNull() && !data.UUID.IsUnknown() && data.UUID.ValueString() != "" {
		entity["uuid"] = data.UUID.ValueString()
	} else {
		entity["uuid"] = generateUUID()
	}

	// Set timing parameters (BCM uses camelCase: heartBeatInterval, electionTimeout)
	if !data.HeartbeatInterval.IsNull() && !data.HeartbeatInterval.IsUnknown() {
		entity["heartBeatInterval"] = data.HeartbeatInterval.ValueInt64()
	} else {
		entity["heartBeatInterval"] = int64(100)
	}

	if !data.ElectionTimeout.IsNull() && !data.ElectionTimeout.IsUnknown() {
		entity["electionTimeout"] = data.ElectionTimeout.ValueInt64()
	} else {
		entity["electionTimeout"] = int64(1000)
	}

	// Set options
	if !data.Options.IsNull() && !data.Options.IsUnknown() && data.Options.ValueString() != "" {
		var options map[string]interface{}
		if err := json.Unmarshal([]byte(data.Options.ValueString()), &options); err == nil {
			entity["options"] = options
		} else {
			entity["options"] = map[string]interface{}{}
		}
	} else {
		entity["options"] = map[string]interface{}{}
	}

	return entity
}

// parseResponseIntoModel updates the Terraform model from BCM API response.
func (r *CMEtcdClusterResource) parseResponseIntoModel(_ context.Context, data map[string]interface{}, model *EtcdClusterResourceModel) {
	// Parse identifiers
	model.UUID = getStringValue(data, "uuid")
	model.ID = model.UUID
	model.Name = getStringValue(data, "name")

	// Parse timing parameters (BCM uses camelCase: heartBeatInterval, electionTimeout)
	model.HeartbeatInterval = getInt64Value(data, "heartBeatInterval")
	model.ElectionTimeout = getInt64Value(data, "electionTimeout")

	// Parse options as JSON string
	if options, ok := data["options"]; ok && options != nil {
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

	// Parse computed fields
	model.CreationTime = getInt64Value(data, "creationTime")
	model.RevisionID = getInt64Value(data, "revisionID")
}
