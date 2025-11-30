// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"strings"
	"testing"
)

func TestBuildDependencyError_Category_SingleDevice(t *testing.T) {
	identifiers := []ResourceIdentifier{
		{UUID: "device-uuid-1", Name: "node01", Type: "device"},
	}

	msg := BuildDependencyError("Category", "default", "device", identifiers)

	// Verify message structure
	if !strings.Contains(msg, "Category 'default' cannot be deleted") {
		t.Error("Message should contain resource type and name")
	}

	if !strings.Contains(msg, "1 device assigned") {
		t.Error("Message should contain correct count with singular form")
	}

	if !strings.Contains(msg, "node01 (uuid: device-uuid-1)") {
		t.Error("Message should list dependent device")
	}

	if !strings.Contains(msg, "Resolution options:") {
		t.Error("Message should include resolution options")
	}

	if !strings.Contains(msg, "Reassign devices to another category") {
		t.Error("Message should include category-specific resolution option")
	}

	if !strings.Contains(msg, "force = true") {
		t.Error("Message should mention force parameter option")
	}
}

func TestBuildDependencyError_Category_MultipleDevices(t *testing.T) {
	identifiers := []ResourceIdentifier{
		{UUID: "device-uuid-1", Name: "node01", Type: "device"},
		{UUID: "device-uuid-2", Name: "node02", Type: "device"},
		{UUID: "device-uuid-3", Name: "node03", Type: "device"},
	}

	msg := BuildDependencyError("Category", "default", "device", identifiers)

	// Verify plural form
	if !strings.Contains(msg, "3 device(s) assigned") {
		t.Error("Message should contain correct count with plural form")
	}

	// Verify all devices listed
	if !strings.Contains(msg, "node01") {
		t.Error("Message should list first device")
	}
	if !strings.Contains(msg, "node02") {
		t.Error("Message should list second device")
	}
	if !strings.Contains(msg, "node03") {
		t.Error("Message should list third device")
	}
}

func TestBuildDependencyError_SoftwareImage_Categories(t *testing.T) {
	identifiers := []ResourceIdentifier{
		{UUID: "category-uuid-1", Name: "default", Type: "category"},
		{UUID: "category-uuid-2", Name: "compute", Type: "category"},
	}

	msg := BuildDependencyError("Software Image", "Rocky-8.10", "category", identifiers)

	// Verify message structure
	if !strings.Contains(msg, "Software Image 'Rocky-8.10' cannot be deleted") {
		t.Error("Message should contain resource type and name")
	}

	if !strings.Contains(msg, "2 categor(ies) assigned") {
		t.Error("Message should contain correct count with proper plural form for 'category'")
	}

	if !strings.Contains(msg, "default (uuid: category-uuid-1)") {
		t.Error("Message should list dependent category")
	}

	if !strings.Contains(msg, "Update categories to use a different software image") {
		t.Error("Message should include software image-specific resolution option")
	}
}

func TestBuildDependencyError_Truncation(t *testing.T) {
	// Create 15 identifiers (should truncate to 10)
	identifiers := make([]ResourceIdentifier, 15)
	for i := 0; i < 15; i++ {
		identifiers[i] = ResourceIdentifier{
			UUID: "device-uuid-" + string(rune(i)),
			Name: "node" + string(rune('0'+i)),
			Type: "device",
		}
	}

	msg := BuildDependencyError("Category", "test", "device", identifiers)

	// Verify truncation message
	if !strings.Contains(msg, "showing first 10 of 15") {
		t.Error("Message should indicate truncation when more than 10 dependents")
	}

	if !strings.Contains(msg, "15 device(s) assigned") {
		t.Error("Message should show total count, not truncated count")
	}
}

func TestBuildForceDeletionWarning_Category(t *testing.T) {
	msg := BuildForceDeletionWarning("Category", "default")

	// Verify message structure
	if !strings.Contains(msg, "Category 'default' is being deleted with force=true") {
		t.Error("Message should contain resource type and name")
	}

	if !strings.Contains(msg, "orphaned references") {
		t.Error("Message should warn about orphaned references")
	}

	if !strings.Contains(msg, "Devices assigned to this category will have invalid category references") {
		t.Error("Message should include category-specific impact")
	}

	if !strings.Contains(msg, "This operation cannot be undone") {
		t.Error("Message should include irreversibility warning")
	}
}

func TestBuildForceDeletionWarning_SoftwareImage(t *testing.T) {
	msg := BuildForceDeletionWarning("Software Image", "Rocky-8.10")

	// Verify message structure
	if !strings.Contains(msg, "Software Image 'Rocky-8.10' is being deleted with force=true") {
		t.Error("Message should contain resource type and name")
	}

	if !strings.Contains(msg, "Categories using this image will have invalid software image references") {
		t.Error("Message should include software image-specific impact")
	}

	if !strings.Contains(msg, "Device provisioning may fail") {
		t.Error("Message should warn about potential provisioning failures")
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		a        int
		b        int
		expected int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{5, 5, 5},
		{0, 10, 0},
		{-1, 5, -1},
	}

	for _, tt := range tests {
		result := min(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("min(%d, %d) = %d, expected %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestBuildDependencyError_DefaultResourceType(t *testing.T) {
	// Test with a resource type that doesn't have specific handling
	identifiers := []ResourceIdentifier{
		{UUID: "network-uuid-1", Name: "network01", Type: "network"},
	}

	msg := BuildDependencyError("Network", "test-network", "connection", identifiers)

	// Verify generic message is used
	if !strings.Contains(msg, "Network 'test-network' cannot be deleted") {
		t.Error("Message should contain resource type and name")
	}

	if !strings.Contains(msg, "Remove dependencies on this Network") {
		t.Error("Message should include generic resolution option")
	}
}
