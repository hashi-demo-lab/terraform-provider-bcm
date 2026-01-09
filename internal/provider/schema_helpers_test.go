// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// =============================================================================
// UUID Validator Tests
// =============================================================================

func TestUUIDValidator_ValidUUIDs(t *testing.T) {
	t.Parallel()

	validUUIDs := []string{
		"12345678-1234-1234-1234-123456789abc",
		"00000000-0000-0000-0000-000000000000",
		"ffffffff-ffff-ffff-ffff-ffffffffffff",
		"a1b2c3d4-e5f6-7890-abcd-ef1234567890",
	}

	for _, uuid := range validUUIDs {
		t.Run(uuid, func(t *testing.T) {
			if !UUIDRegex.MatchString(uuid) {
				t.Errorf("UUIDRegex should match valid UUID %q", uuid)
			}
		})
	}
}

func TestUUIDValidator_InvalidUUIDs(t *testing.T) {
	t.Parallel()

	invalidUUIDs := []string{
		"",                                      // empty
		"not-a-uuid",                            // wrong format
		"12345678123412341234123456789abc",      // no dashes
		"12345678-1234-1234-1234-123456789ab",   // too short
		"12345678-1234-1234-1234-123456789abcd", // too long
		"AAAAAAAA-AAAA-AAAA-AAAA-AAAAAAAAAAAA",  // uppercase (RFC 4122 lowercase only)
		"12345678-1234-1234-1234-123456789ABC",  // mixed case
		"g2345678-1234-1234-1234-123456789abc",  // invalid hex character
		"12345678_1234_1234_1234_123456789abc",  // underscores instead of dashes
		"123456781234-1234-1234-123456789abc",   // wrong dash positions
	}

	for _, uuid := range invalidUUIDs {
		t.Run(uuid, func(t *testing.T) {
			if UUIDRegex.MatchString(uuid) {
				t.Errorf("UUIDRegex should NOT match invalid UUID %q", uuid)
			}
		})
	}
}

func TestUUIDValidator_Function(t *testing.T) {
	t.Parallel()

	v := UUIDValidator()
	if v == nil {
		t.Fatal("UUIDValidator() returned nil")
	}

}

// =============================================================================
// Schema Attribute Tests
// =============================================================================

func TestComputedIDAttribute(t *testing.T) {
	t.Parallel()

	attr := ComputedIDAttribute()

	if !attr.Computed {
		t.Error("ComputedIDAttribute should be computed")
	}
	if attr.Optional {
		t.Error("ComputedIDAttribute should not be optional")
	}
	if attr.Required {
		t.Error("ComputedIDAttribute should not be required")
	}
	if attr.MarkdownDescription == "" {
		t.Error("ComputedIDAttribute should have a description")
	}
	if len(attr.PlanModifiers) == 0 {
		t.Error("ComputedIDAttribute should have plan modifiers")
	}
}

func TestComputedUUIDAttribute(t *testing.T) {
	t.Parallel()

	t.Run("with_custom_description", func(t *testing.T) {
		attr := ComputedUUIDAttribute("Custom UUID description")
		if attr.MarkdownDescription != "Custom UUID description" {
			t.Errorf("Expected custom description, got %q", attr.MarkdownDescription)
		}
	})

	t.Run("with_empty_description", func(t *testing.T) {
		attr := ComputedUUIDAttribute("")
		if attr.MarkdownDescription == "" {
			t.Error("Should have default description when empty string provided")
		}
	})

	t.Run("attributes", func(t *testing.T) {
		attr := ComputedUUIDAttribute("")
		if !attr.Computed {
			t.Error("ComputedUUIDAttribute should be computed")
		}
		if len(attr.PlanModifiers) == 0 {
			t.Error("ComputedUUIDAttribute should have plan modifiers")
		}
	})
}

func TestRequiredUUIDAttribute(t *testing.T) {
	t.Parallel()

	attr := RequiredUUIDAttribute("Test description")

	if !attr.Required {
		t.Error("RequiredUUIDAttribute should be required")
	}
	if attr.Optional {
		t.Error("RequiredUUIDAttribute should not be optional")
	}
	if attr.Computed {
		t.Error("RequiredUUIDAttribute should not be computed")
	}
	if len(attr.Validators) == 0 {
		t.Error("RequiredUUIDAttribute should have validators")
	}
}

func TestOptionalUUIDAttribute(t *testing.T) {
	t.Parallel()

	attr := OptionalUUIDAttribute("Test description")

	if !attr.Optional {
		t.Error("OptionalUUIDAttribute should be optional")
	}
	if attr.Required {
		t.Error("OptionalUUIDAttribute should not be required")
	}
	if attr.Computed {
		t.Error("OptionalUUIDAttribute should not be computed")
	}
	if len(attr.Validators) == 0 {
		t.Error("OptionalUUIDAttribute should have validators")
	}
}

func TestForceAttribute(t *testing.T) {
	t.Parallel()

	t.Run("with_custom_description", func(t *testing.T) {
		attr := ForceAttribute("Custom force description")
		if attr.MarkdownDescription != "Custom force description" {
			t.Errorf("Expected custom description, got %q", attr.MarkdownDescription)
		}
	})

	t.Run("with_empty_description", func(t *testing.T) {
		attr := ForceAttribute("")
		if attr.MarkdownDescription == "" {
			t.Error("Should have default description when empty string provided")
		}
	})

	t.Run("attributes", func(t *testing.T) {
		attr := ForceAttribute("")
		if !attr.Optional {
			t.Error("ForceAttribute should be optional")
		}
		if !attr.Computed {
			t.Error("ForceAttribute should be computed (for default value)")
		}
		if attr.Default == nil {
			t.Error("ForceAttribute should have a default value")
		}
	})
}

func TestNotesAttribute(t *testing.T) {
	t.Parallel()

	t.Run("computed_true_for_data_source", func(t *testing.T) {
		attr := NotesAttribute(true, "Notes for data source")
		if !attr.Computed {
			t.Error("NotesAttribute(true, ...) should be computed")
		}
		if attr.Optional {
			t.Error("NotesAttribute(true, ...) should not be optional")
		}
	})

	t.Run("computed_false_for_resource", func(t *testing.T) {
		attr := NotesAttribute(false, "Notes for resource")
		if attr.Computed {
			t.Error("NotesAttribute(false, ...) should not be computed")
		}
		if !attr.Optional {
			t.Error("NotesAttribute(false, ...) should be optional")
		}
	})

	t.Run("with_empty_description", func(t *testing.T) {
		attr := NotesAttribute(false, "")
		if attr.MarkdownDescription == "" {
			t.Error("Should have default description when empty string provided")
		}
	})
}

func TestBaseTypeAttribute(t *testing.T) {
	t.Parallel()

	t.Run("with_custom_description", func(t *testing.T) {
		attr := BaseTypeAttribute("Custom base type")
		if attr.MarkdownDescription != "Custom base type" {
			t.Errorf("Expected custom description, got %q", attr.MarkdownDescription)
		}
	})

	t.Run("with_empty_description", func(t *testing.T) {
		attr := BaseTypeAttribute("")
		if attr.MarkdownDescription == "" {
			t.Error("Should have default description when empty string provided")
		}
	})

	t.Run("attributes", func(t *testing.T) {
		attr := BaseTypeAttribute("")
		if !attr.Computed {
			t.Error("BaseTypeAttribute should be computed")
		}
	})
}

func TestChildTypeAttribute(t *testing.T) {
	t.Parallel()

	t.Run("with_custom_description", func(t *testing.T) {
		attr := ChildTypeAttribute("Custom child type")
		if attr.MarkdownDescription != "Custom child type" {
			t.Errorf("Expected custom description, got %q", attr.MarkdownDescription)
		}
	})

	t.Run("with_empty_description", func(t *testing.T) {
		attr := ChildTypeAttribute("")
		if attr.MarkdownDescription == "" {
			t.Error("Should have default description when empty string provided")
		}
	})

	t.Run("attributes", func(t *testing.T) {
		attr := ChildTypeAttribute("")
		if !attr.Computed {
			t.Error("ChildTypeAttribute should be computed")
		}
	})
}

func TestRevisionAttribute(t *testing.T) {
	t.Parallel()

	attr := RevisionAttribute()
	if !attr.Computed {
		t.Error("RevisionAttribute should be computed")
	}
	if attr.MarkdownDescription == "" {
		t.Error("RevisionAttribute should have a description")
	}
}

func TestModifiedAttribute(t *testing.T) {
	t.Parallel()

	attr := ModifiedAttribute()
	if !attr.Computed {
		t.Error("ModifiedAttribute should be computed")
	}
	if attr.MarkdownDescription == "" {
		t.Error("ModifiedAttribute should have a description")
	}
}

func TestToBeRemovedAttribute(t *testing.T) {
	t.Parallel()

	attr := ToBeRemovedAttribute()
	if !attr.Computed {
		t.Error("ToBeRemovedAttribute should be computed")
	}
	if attr.MarkdownDescription == "" {
		t.Error("ToBeRemovedAttribute should have a description")
	}
}

func TestCreationTimeAttribute(t *testing.T) {
	t.Parallel()

	t.Run("with_custom_description", func(t *testing.T) {
		attr := CreationTimeAttribute("Custom creation time")
		if attr.MarkdownDescription != "Custom creation time" {
			t.Errorf("Expected custom description, got %q", attr.MarkdownDescription)
		}
	})

	t.Run("with_empty_description", func(t *testing.T) {
		attr := CreationTimeAttribute("")
		if attr.MarkdownDescription == "" {
			t.Error("Should have default description when empty string provided")
		}
	})

	t.Run("attributes", func(t *testing.T) {
		attr := CreationTimeAttribute("")
		if !attr.Computed {
			t.Error("CreationTimeAttribute should be computed")
		}
	})
}

// =============================================================================
// Data Source Schema Helper Tests
// =============================================================================

func TestDataSourceComputedStringAttribute(t *testing.T) {
	t.Parallel()

	attr := DataSourceComputedStringAttribute("Test description")
	if !attr.Computed {
		t.Error("DataSourceComputedStringAttribute should be computed")
	}
	if attr.MarkdownDescription != "Test description" {
		t.Errorf("Expected description %q, got %q", "Test description", attr.MarkdownDescription)
	}
}

func TestDataSourceComputedBoolAttribute(t *testing.T) {
	t.Parallel()

	attr := DataSourceComputedBoolAttribute("Test description")
	if !attr.Computed {
		t.Error("DataSourceComputedBoolAttribute should be computed")
	}
	if attr.MarkdownDescription != "Test description" {
		t.Errorf("Expected description %q, got %q", "Test description", attr.MarkdownDescription)
	}
}

func TestDataSourceComputedInt64Attribute(t *testing.T) {
	t.Parallel()

	attr := DataSourceComputedInt64Attribute("Test description")
	if !attr.Computed {
		t.Error("DataSourceComputedInt64Attribute should be computed")
	}
	if attr.MarkdownDescription != "Test description" {
		t.Errorf("Expected description %q, got %q", "Test description", attr.MarkdownDescription)
	}
}

// =============================================================================
// UUID Validator Integration Tests
// =============================================================================

func TestUUIDValidator_ValidateString(t *testing.T) {
	t.Parallel()

	v := UUIDValidator()
	ctx := context.Background()

	testCases := []struct {
		name        string
		value       string
		expectError bool
	}{
		{
			name:        "valid_uuid",
			value:       "12345678-1234-1234-1234-123456789abc",
			expectError: false,
		},
		{
			name:        "invalid_uuid_uppercase",
			value:       "12345678-1234-1234-1234-123456789ABC",
			expectError: true,
		},
		{
			name:        "invalid_uuid_format",
			value:       "not-a-uuid",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := validator.StringRequest{
				ConfigValue: types.StringValue(tc.value),
			}
			resp := &validator.StringResponse{}

			v.ValidateString(ctx, req, resp)

			if tc.expectError && !resp.Diagnostics.HasError() {
				t.Errorf("Expected validation error for %q but got none", tc.value)
			}
			if !tc.expectError && resp.Diagnostics.HasError() {
				t.Errorf("Did not expect validation error for %q but got: %v", tc.value, resp.Diagnostics)
			}
		})
	}
}
