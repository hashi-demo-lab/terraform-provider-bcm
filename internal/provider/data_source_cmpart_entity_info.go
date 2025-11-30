// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &CMPartEntityInfoDataSource{}
	_ datasource.DataSourceWithConfigure = &CMPartEntityInfoDataSource{}
)

// NewCMPartEntityInfoDataSource is a helper function to simplify the provider implementation.
func NewCMPartEntityInfoDataSource() datasource.DataSource {
	return &CMPartEntityInfoDataSource{}
}

// CMPartEntityInfoDataSource is the data source implementation.
type CMPartEntityInfoDataSource struct {
	BCMDataSourceBase
}

// CMPartEntityInfoDataSourceModel describes the data source data model.
type CMPartEntityInfoDataSourceModel struct {
	ID          types.String      `tfsdk:"id"`
	Type        types.String      `tfsdk:"type"`
	NamePattern types.String      `tfsdk:"name_pattern"`
	Entities    []EntityInfoModel `tfsdk:"entities"`
}

// EntityInfoModel represents a BCM entity's basic information.
type EntityInfoModel struct {
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
	UUID types.String `tfsdk:"uuid"`
}

// Metadata returns the data source type name.
func (d *CMPartEntityInfoDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cmpart_entity_info"
}

// Schema defines the schema for the data source.
func (d *CMPartEntityInfoDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves basic entity information from BCM CMPart service. Use this data source to discover BCM entities and obtain their UUIDs for reference in other Terraform resources.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Data source identifier for state management. Format: `cmpart-entity-info:{type}:{name_pattern}` or `cmpart-entity-info:all` when no filters.",
			},
			"type": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter by entity type (case-sensitive exact match). Examples: `SoftwareImage`, `Category`, `Network`, `HeadNode`.",
			},
			"name_pattern": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter by name using glob pattern (case-insensitive). Supports `*` (any characters) and `?` (single character) wildcards. Examples: `default*`, `*node*`, `ubuntu-?`.",
			},
			"entities": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of entities matching the filter criteria.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Entity name (mapped from API's `resolveName` field).",
						},
						"type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Entity type classification (e.g., `SoftwareImage`, `Category`).",
						},
						"uuid": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Entity unique identifier (UUID format).",
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *CMPartEntityInfoDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.ConfigureDataSource(req, resp)
}

// Read refreshes the Terraform state with the latest data.
func (d *CMPartEntityInfoDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config CMPartEntityInfoDataSourceModel

	// Read configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Log the read operation
	tflog.Debug(ctx, "Reading BCM entity information from API", map[string]interface{}{
		"service": "cmpart",
		"call":    "getBasicEntityInformation",
	})

	// Log filter criteria if present
	filterCriteria := map[string]interface{}{}
	if !config.Type.IsNull() && !config.Type.IsUnknown() {
		filterCriteria["type"] = config.Type.ValueString()
	}
	if !config.NamePattern.IsNull() && !config.NamePattern.IsUnknown() {
		filterCriteria["name_pattern"] = config.NamePattern.ValueString()
	}
	if len(filterCriteria) > 0 {
		tflog.Debug(ctx, "Applying client-side filters", filterCriteria)
	}

	// Call BCM API - service name is lowercase "cmpart"
	body, err := d.Client.CallJSONRPC(ctx, "cmpart", "getBasicEntityInformation")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read BCM Entity Information",
			"An unexpected error occurred when calling the BCM API. "+
				"Check the endpoint, credentials, and network connectivity.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "Successfully retrieved BCM API response", map[string]interface{}{
		"response_size": len(body),
	})

	// Parse JSON response - expect array of entity objects
	var apiResponse []map[string]interface{}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Parse BCM API Response",
			"The provider received an unexpected response format from the BCM API. "+
				"This may indicate an API compatibility issue.\n\n"+
				"Parse Error: "+err.Error()+"\n"+
				"Response: "+limitString(string(body), 500),
		)
		return
	}

	tflog.Debug(ctx, "Parsed BCM API response", map[string]interface{}{
		"total_entities": len(apiResponse),
	})

	// Generate data source ID based on filter values
	id := generateEntityInfoDataSourceID(config.Type, config.NamePattern)

	// Initialize state with computed ID and filter values
	state := CMPartEntityInfoDataSourceModel{
		ID:          types.StringValue(id),
		Type:        config.Type,
		NamePattern: config.NamePattern,
		Entities:    []EntityInfoModel{},
	}

	// Map and filter entities
	for _, entityData := range apiResponse {
		entity := mapAPIResponseToEntityInfo(entityData)

		// Apply filters using matchesEntityFilter helper
		if matchesEntityFilter(entity, config.Type, config.NamePattern) {
			state.Entities = append(state.Entities, entity)
		}
	}

	tflog.Info(ctx, "Entity information data source read complete", map[string]interface{}{
		"total_entities":    len(apiResponse),
		"filtered_entities": len(state.Entities),
		"id":                id,
	})

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// mapAPIResponseToEntityInfo converts API JSON to EntityInfoModel.
// Maps API "resolveName" field to "name" attribute for clarity.
func mapAPIResponseToEntityInfo(apiData map[string]interface{}) EntityInfoModel {
	return EntityInfoModel{
		// Map resolveName to name (FR-005)
		Name: getStringValue(apiData, "resolveName"),
		Type: getStringValue(apiData, "type"),
		UUID: getStringValue(apiData, "uuid"),
	}
}

// matchesEntityFilter checks if an entity matches the filter criteria.
// Both filters use AND logic - the entity must match all specified filters.
//
// Filter behavior:
//   - type: Case-sensitive exact match (FR-002)
//   - name_pattern: Case-insensitive glob pattern matching (FR-003, FR-011)
//   - Omitted filters (null/unknown) are ignored (do not restrict results)
//
// Returns true if the entity matches all specified filters, false otherwise.
func matchesEntityFilter(entity EntityInfoModel, typeFilter, namePattern types.String) bool {
	// Type filter: case-sensitive exact match (FR-002)
	if !typeFilter.IsNull() && !typeFilter.IsUnknown() {
		if entity.Type.ValueString() != typeFilter.ValueString() {
			return false
		}
	}

	// Name pattern filter: case-insensitive glob match (FR-003, FR-011)
	if !namePattern.IsNull() && !namePattern.IsUnknown() {
		if !matchesNamePattern(entity.Name.ValueString(), namePattern.ValueString()) {
			return false
		}
	}

	// Entity matches all specified filters (FR-007, FR-008)
	return true
}

// matchesNamePattern checks if a name matches a glob pattern (case-insensitive).
// Uses filepath.Match for glob pattern support (* and ? wildcards).
// Returns false for invalid patterns (defensive behavior).
func matchesNamePattern(name, pattern string) bool {
	// Case-insensitive matching per FR-011
	lowerName := strings.ToLower(name)
	lowerPattern := strings.ToLower(pattern)

	matched, err := filepath.Match(lowerPattern, lowerName)
	if err != nil {
		// Invalid pattern - treat as no match (defensive behavior)
		return false
	}
	return matched
}

// generateEntityInfoDataSourceID generates a deterministic ID from filter parameters.
// Format: "cmpart-entity-info:{type}:{name_pattern}" or "cmpart-entity-info:all" when no filters.
// This ensures consistent IDs across terraform plan/apply cycles with same filters.
func generateEntityInfoDataSourceID(typeFilter, namePattern types.String) string {
	typeVal := ""
	patternVal := ""

	if !typeFilter.IsNull() && !typeFilter.IsUnknown() {
		typeVal = typeFilter.ValueString()
	}

	if !namePattern.IsNull() && !namePattern.IsUnknown() {
		patternVal = namePattern.ValueString()
	}

	// No filters = return "all" ID (FR-008)
	if typeVal == "" && patternVal == "" {
		return "cmpart-entity-info:all"
	}

	// Return formatted ID with filter values (FR-009)
	return fmt.Sprintf("cmpart-entity-info:%s:%s", typeVal, patternVal)
}
