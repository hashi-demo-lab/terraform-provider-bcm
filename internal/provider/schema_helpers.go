// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Package provider contains reusable schema attribute helpers for the BCM Terraform provider.
// These helpers reduce code duplication by providing pre-configured schema attributes
// that are commonly used across multiple resources and data sources.
package provider

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// =============================================================================
// Common Validators
// =============================================================================

// UUIDRegex is the compiled regex for RFC 4122 UUID validation.
var UUIDRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// UUIDValidator returns a string validator that ensures the value is a valid RFC 4122 UUID.
func UUIDValidator() validator.String {
	return stringvalidator.RegexMatches(
		UUIDRegex,
		"must be a valid RFC 4122 UUID format (e.g., 12345678-1234-1234-1234-123456789abc)",
	)
}

// =============================================================================
// Identity Attributes - ID and UUID fields
// =============================================================================

// ComputedIDAttribute returns a computed ID attribute with UseStateForUnknown.
// This is the standard pattern for resource identifiers.
func ComputedIDAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "Resource identifier (same as UUID).",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}

// ComputedUUIDAttribute returns a computed UUID attribute with UseStateForUnknown.
// Use this for BCM-assigned unique identifiers.
func ComputedUUIDAttribute(description string) schema.StringAttribute {
	if description == "" {
		description = "Unique identifier assigned by BCM."
	}
	return schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: description,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}

// RequiredUUIDAttribute returns a required UUID attribute with validation.
// Use this for UUID references to other resources.
func RequiredUUIDAttribute(description string) schema.StringAttribute {
	return schema.StringAttribute{
		Required:            true,
		MarkdownDescription: description,
		Validators: []validator.String{
			UUIDValidator(),
		},
	}
}

// OptionalUUIDAttribute returns an optional UUID attribute with validation.
// Use this for optional UUID references.
func OptionalUUIDAttribute(description string) schema.StringAttribute {
	return schema.StringAttribute{
		Optional:            true,
		MarkdownDescription: description,
		Validators: []validator.String{
			UUIDValidator(),
		},
	}
}

// =============================================================================
// Common Resource Attributes
// =============================================================================

// ForceAttribute returns the standard force attribute with default false.
// This attribute controls whether operations proceed despite warnings.
func ForceAttribute(description string) schema.BoolAttribute {
	if description == "" {
		description = "Force operation even with validation warnings. Defaults to `false`."
	}
	return schema.BoolAttribute{
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(false),
		MarkdownDescription: description,
	}
}

// NotesAttribute returns an optional notes/description attribute.
// Use computed=true for data sources, computed=false for resources.
func NotesAttribute(computed bool, description string) schema.StringAttribute {
	if description == "" {
		description = "User notes or description."
	}
	return schema.StringAttribute{
		Optional:            !computed,
		Computed:            computed,
		MarkdownDescription: description,
	}
}

// =============================================================================
// BCM Metadata Attributes - Common to all BCM entities
// =============================================================================

// BaseTypeAttribute returns the computed base_type attribute.
func BaseTypeAttribute(description string) schema.StringAttribute {
	if description == "" {
		description = "BCM entity base type."
	}
	return schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: description,
	}
}

// ChildTypeAttribute returns the computed child_type attribute.
func ChildTypeAttribute(description string) schema.StringAttribute {
	if description == "" {
		description = "BCM entity child type."
	}
	return schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: description,
	}
}

// RevisionAttribute returns the computed revision attribute.
func RevisionAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "BCM revision identifier for concurrency control.",
	}
}

// ModifiedAttribute returns the computed modified flag attribute.
func ModifiedAttribute() schema.BoolAttribute {
	return schema.BoolAttribute{
		Computed:            true,
		MarkdownDescription: "BCM modification flag.",
	}
}

// ToBeRemovedAttribute returns the computed to_be_removed flag attribute.
func ToBeRemovedAttribute() schema.BoolAttribute {
	return schema.BoolAttribute{
		Computed:            true,
		MarkdownDescription: "BCM removal flag.",
	}
}

// CreationTimeAttribute returns a computed creation_time attribute.
func CreationTimeAttribute(description string) schema.Int64Attribute {
	if description == "" {
		description = "Creation timestamp (Unix epoch)."
	}
	return schema.Int64Attribute{
		Computed:            true,
		MarkdownDescription: description,
	}
}

// =============================================================================
// Data Source Schema Helpers
// =============================================================================

// DataSourceComputedStringAttribute returns a computed string for data sources.
func DataSourceComputedStringAttribute(description string) schema.StringAttribute {
	return schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: description,
	}
}

// DataSourceComputedBoolAttribute returns a computed bool for data sources.
func DataSourceComputedBoolAttribute(description string) schema.BoolAttribute {
	return schema.BoolAttribute{
		Computed:            true,
		MarkdownDescription: description,
	}
}

// DataSourceComputedInt64Attribute returns a computed int64 for data sources.
func DataSourceComputedInt64Attribute(description string) schema.Int64Attribute {
	return schema.Int64Attribute{
		Computed:            true,
		MarkdownDescription: description,
	}
}
