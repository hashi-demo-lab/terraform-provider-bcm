// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Package provider contains shared utility functions for the BCM Terraform provider.
// These helpers provide null-safe field extraction from BCM API responses (map[string]interface{})
// and common operations used across multiple resources and data sources.
package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// =============================================================================
// Type Conversion Helpers - Null-safe extraction from map[string]interface{}
// =============================================================================

// getStringValue safely extracts a string value from a map, returning types.StringNull()
// if the key is missing, nil, empty string, or wrong type.
//
// This is the primary helper for converting BCM API JSON responses to Terraform types.
// Empty strings are treated as null since BCM often returns "" for unset fields.
func getStringValue(data map[string]interface{}, key string) types.String {
	if data == nil {
		return types.StringNull()
	}
	if val, ok := data[key]; ok && val != nil {
		if str, ok := val.(string); ok && str != "" {
			return types.StringValue(str)
		}
	}
	return types.StringNull()
}

// getBoolValue safely extracts a boolean value from a map, returning types.BoolNull()
// if the key is missing, nil, or wrong type.
//
// Note: Only actual bool values are accepted. Strings like "true"/"false" or
// integers like 0/1 will return null - use explicit conversion if needed.
func getBoolValue(data map[string]interface{}, key string) types.Bool {
	if data == nil {
		return types.BoolNull()
	}
	if val, ok := data[key]; ok && val != nil {
		if b, ok := val.(bool); ok {
			return types.BoolValue(b)
		}
	}
	return types.BoolNull()
}

// getInt64Value safely extracts an integer value from a map, returning types.Int64Null()
// if the key is missing, nil, or wrong type.
//
// Handles multiple numeric types since JSON unmarshal typically produces float64:
//   - float64: Truncated to int64 (JSON default for numbers)
//   - int64: Used directly
//   - int: Converted to int64
func getInt64Value(data map[string]interface{}, key string) types.Int64 {
	if data == nil {
		return types.Int64Null()
	}
	if val, ok := data[key]; ok && val != nil {
		switch v := val.(type) {
		case float64:
			return types.Int64Value(int64(v))
		case int64:
			return types.Int64Value(v)
		case int:
			return types.Int64Value(int64(v))
		}
	}
	return types.Int64Null()
}

// getFloat64Value safely extracts a float64 value from a map, returning types.Float64Null()
// if the key is missing, nil, or wrong type.
//
// Handles conversion from integer types for flexibility.
func getFloat64Value(data map[string]interface{}, key string) types.Float64 {
	if data == nil {
		return types.Float64Null()
	}
	if val, ok := data[key]; ok && val != nil {
		switch v := val.(type) {
		case float64:
			return types.Float64Value(v)
		case int64:
			return types.Float64Value(float64(v))
		case int:
			return types.Float64Value(float64(v))
		}
	}
	return types.Float64Null()
}

// =============================================================================
// String Utilities
// =============================================================================

// containsAny checks if a string contains any of the specified substrings (case-insensitive).
// Returns true if s contains at least one of the substrings, false otherwise.
//
// Example:
//
//	containsAny("Hello World", []string{"world", "foo"}) // true
//	containsAny("Hello", []string{"world", "foo"})       // false
func containsAny(s string, substrings []string) bool {
	lowerStr := strings.ToLower(s)
	for _, substring := range substrings {
		if strings.Contains(lowerStr, strings.ToLower(substring)) {
			return true
		}
	}
	return false
}

// =============================================================================
// UUID Utilities
// =============================================================================

// generateUUID creates a new UUID v4 string.
// Used for resources that require client-generated UUIDs (e.g., roles, static routes).
func generateUUID() string {
	return uuid.New().String()
}

// =============================================================================
// Validation Helpers
// =============================================================================

// ProcessValidationErrors converts BCM validation results to Terraform diagnostics.
// Returns true if any validation errors were found (caller should halt the operation).
//
// This centralizes the common pattern of processing BCM API validation responses
// and converting them to Terraform diagnostics with appropriate severity levels.
//
// Usage:
//
//	validationErrors, err := client.ValidateEntity(ctx, service, method, entity, isCreate)
//	if err != nil {
//	    resp.Diagnostics.AddError("Validation Failed", err.Error())
//	    return
//	}
//	if ProcessValidationErrors(validationErrors, &resp.Diagnostics) {
//	    return // Validation errors prevent operation
//	}
func ProcessValidationErrors(validationErrors []ValidationError, diags *diag.Diagnostics) bool {
	hasErrors := false
	for _, valErr := range validationErrors {
		if valErr.IsError() {
			diags.AddError(
				fmt.Sprintf("Validation Error: %s", valErr.Field),
				valErr.Message,
			)
			hasErrors = true
		} else if valErr.IsWarning() {
			diags.AddWarning(
				fmt.Sprintf("Validation Warning: %s", valErr.Field),
				valErr.Message,
			)
		}
	}
	return hasErrors
}

// =============================================================================
// Terraform Type Helpers
// =============================================================================

// resolveUnknownToNull converts Unknown values to Null for state consistency.
// This is critical because Terraform Plugin Framework does not allow Unknown values
// in state after Create/Update operations.
//
// Use this for computed fields that might be Unknown during planning but must be
// resolved to either a concrete value or Null after the operation.
func resolveUnknownStringToNull(val types.String) types.String {
	if val.IsUnknown() {
		return types.StringNull()
	}
	return val
}

// resolveUnknownInt64ToNull converts Unknown Int64 values to Null.
func resolveUnknownInt64ToNull(val types.Int64) types.Int64 {
	if val.IsUnknown() {
		return types.Int64Null()
	}
	return val
}

// resolveUnknownBoolToNull converts Unknown Bool values to Null.
func resolveUnknownBoolToNull(val types.Bool) types.Bool {
	if val.IsUnknown() {
		return types.BoolNull()
	}
	return val
}

// =============================================================================
// Entity Building Helpers - For constructing BCM API request entities
// =============================================================================

// NewBCMEntity creates a new BCM entity map with standard base fields.
// All BCM entities require these common fields for API operations.
//
// Parameters:
//   - baseType: The BCM entity type (e.g., "Category", "Device", "SoftwareImage")
//
// Example:
//
//	entity := NewBCMEntity("Category")
//	entity["name"] = "my-category"
func NewBCMEntity(baseType string) map[string]interface{} {
	return map[string]interface{}{
		"baseType":      baseType,
		"childType":     "",
		"modified":      true,
		"to_be_removed": false,
		"revision":      "",
	}
}

// NewBCMEntityWithUUID creates a new BCM entity with UUID handling.
// If uuid is empty, a new UUID v4 is generated.
//
// Parameters:
//   - baseType: The BCM entity type
//   - uuid: Existing UUID or empty string for new entity
//
// Returns the entity and the UUID that was set.
func NewBCMEntityWithUUID(baseType, uuid string) (map[string]interface{}, string) {
	entity := NewBCMEntity(baseType)
	if uuid == "" {
		uuid = generateUUID()
	}
	entity["uuid"] = uuid
	return entity, uuid
}

// SetEntityUUID sets or generates a UUID for an entity.
// Returns the UUID that was set (either the provided one or a new generated one).
func SetEntityUUID(entity map[string]interface{}, uuid string) string {
	if uuid == "" {
		uuid = generateUUID()
	}
	entity["uuid"] = uuid
	return uuid
}

// =============================================================================
// Entity Field Setters - Type-safe setters for BCM entity fields
// =============================================================================

// SetStringField sets a string field on an entity if the value is not null/unknown.
// This is the primary helper for converting Terraform types.String to entity fields.
func SetStringField(entity map[string]interface{}, key string, value types.String) {
	if !value.IsNull() && !value.IsUnknown() {
		entity[key] = value.ValueString()
	}
}

// SetBoolField sets a bool field on an entity if the value is not null/unknown.
func SetBoolField(entity map[string]interface{}, key string, value types.Bool) {
	if !value.IsNull() && !value.IsUnknown() {
		entity[key] = value.ValueBool()
	}
}

// SetInt64Field sets an int64 field on an entity if the value is not null/unknown.
func SetInt64Field(entity map[string]interface{}, key string, value types.Int64) {
	if !value.IsNull() && !value.IsUnknown() {
		entity[key] = value.ValueInt64()
	}
}

// SetFloat64Field sets a float64 field on an entity if the value is not null/unknown.
func SetFloat64Field(entity map[string]interface{}, key string, value types.Float64) {
	if !value.IsNull() && !value.IsUnknown() {
		entity[key] = value.ValueFloat64()
	}
}

// SetStringListField sets a string list field on an entity if the value is not null/unknown.
// Returns any diagnostics from the conversion.
func SetStringListField(ctx context.Context, entity map[string]interface{}, key string, value types.List, diags *diag.Diagnostics) {
	if !value.IsNull() && !value.IsUnknown() {
		var list []string
		d := value.ElementsAs(ctx, &list, false)
		diags.Append(d...)
		if !d.HasError() {
			entity[key] = list
		}
	}
}

// =============================================================================
// Force Parameter Helper
// =============================================================================

// GetForceValue extracts the force boolean value, defaulting to false if null/unknown.
// This standardizes the common pattern of extracting force parameters.
func GetForceValue(force types.Bool) bool {
	if force.IsNull() || force.IsUnknown() {
		return false
	}
	return force.ValueBool()
}

// =============================================================================
// List Value Helpers - For extracting list types from API responses
// =============================================================================

// GetStringListValue safely extracts a string list from a map, returning types.ListNull()
// if the key is missing, nil, empty, or wrong type.
func GetStringListValue(data map[string]interface{}, key string) types.List {
	if data == nil {
		return types.ListNull(types.StringType)
	}
	if val, ok := data[key]; ok && val != nil {
		if slice, ok := val.([]interface{}); ok && len(slice) > 0 {
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
	return types.ListNull(types.StringType)
}

// =============================================================================
// Filter Helpers - For data source filtering
// =============================================================================

// MatchesSubstringFilter performs case-insensitive substring matching.
// Returns true if value contains pattern (case-insensitive).
// Returns true if pattern is null/unknown (no filter applied).
func MatchesSubstringFilter(value types.String, pattern types.String) bool {
	if pattern.IsNull() || pattern.IsUnknown() {
		return true // No filter applied
	}
	if value.IsNull() || value.IsUnknown() {
		return false // Can't match against null/unknown value
	}
	return strings.Contains(
		strings.ToLower(value.ValueString()),
		strings.ToLower(pattern.ValueString()),
	)
}

// MatchesExactFilter performs exact string matching.
// Returns true if value equals expected exactly.
// Returns true if expected is null/unknown (no filter applied).
func MatchesExactFilter(value types.String, expected types.String) bool {
	if expected.IsNull() || expected.IsUnknown() {
		return true // No filter applied
	}
	if value.IsNull() || value.IsUnknown() {
		return false // Can't match against null/unknown value
	}
	return value.ValueString() == expected.ValueString()
}

// MatchesExactFilterString performs exact string matching with a raw string pattern.
// Returns true if value equals expected exactly.
// Returns true if expected is empty (no filter applied).
func MatchesExactFilterString(value types.String, expected string) bool {
	if expected == "" {
		return true // No filter applied
	}
	if value.IsNull() || value.IsUnknown() {
		return false // Can't match against null/unknown value
	}
	return value.ValueString() == expected
}

