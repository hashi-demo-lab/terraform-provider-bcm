// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Package provider contains shared utility functions for the BCM Terraform provider.
// These helpers provide null-safe field extraction from BCM API responses (map[string]interface{})
// and common operations used across multiple resources and data sources.
package provider

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
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

