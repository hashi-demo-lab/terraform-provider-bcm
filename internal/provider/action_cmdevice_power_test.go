// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/action"
)

// TestCMDevicePowerAction_Metadata verifies the action returns the correct type name.
func TestCMDevicePowerAction_Metadata(t *testing.T) {
	t.Parallel()

	// Create action instance
	a := NewCMDevicePowerAction()

	// Create request with provider type name
	req := action.MetadataRequest{
		ProviderTypeName: "bcm",
	}
	resp := &action.MetadataResponse{}

	// Call Metadata method
	a.Metadata(context.Background(), req, resp)

	// Verify type name
	expected := "bcm_cmdevice_power"
	if resp.TypeName != expected {
		t.Errorf("Expected TypeName %q, got %q", expected, resp.TypeName)
	}
}

// TestCMDevicePowerAction_Schema verifies the action schema contains required attributes.
func TestCMDevicePowerAction_Schema(t *testing.T) {
	t.Parallel()

	// Create action instance
	a := NewCMDevicePowerAction()

	// Create request and response
	req := action.SchemaRequest{}
	resp := &action.SchemaResponse{}

	// Call Schema method
	a.Schema(context.Background(), req, resp)

	// Verify no diagnostics errors
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema returned errors: %v", resp.Diagnostics)
	}

	// Verify required attributes exist
	requiredAttrs := []string{"device_id", "power_action"}
	for _, attr := range requiredAttrs {
		if _, exists := resp.Schema.Attributes[attr]; !exists {
			t.Errorf("Expected required attribute %q not found in schema", attr)
		}
	}

	// Verify optional attributes exist
	optionalAttrs := []string{"wait_for_completion", "timeout"}
	for _, attr := range optionalAttrs {
		if _, exists := resp.Schema.Attributes[attr]; !exists {
			t.Errorf("Expected optional attribute %q not found in schema", attr)
		}
	}

	// Verify markdown description is set
	if resp.Schema.MarkdownDescription == "" {
		t.Error("Expected MarkdownDescription to be set")
	}
}

// TestCMDevicePowerAction_SchemaAttributeTypes verifies attribute types are correct.
func TestCMDevicePowerAction_SchemaAttributeTypes(t *testing.T) {
	t.Parallel()

	a := NewCMDevicePowerAction()

	req := action.SchemaRequest{}
	resp := &action.SchemaResponse{}

	a.Schema(context.Background(), req, resp)

	// Test cases for attribute verification
	testCases := []struct {
		name       string
		attrName   string
		isRequired bool
	}{
		{"device_id is required", "device_id", true},
		{"power_action is required", "power_action", true},
		{"wait_for_completion is optional", "wait_for_completion", false},
		{"timeout is optional", "timeout", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			attr, exists := resp.Schema.Attributes[tc.attrName]
			if !exists {
				t.Fatalf("Attribute %q not found", tc.attrName)
			}

			// Check if attribute is required/optional based on IsRequired() method
			if tc.isRequired != attr.IsRequired() {
				t.Errorf("Attribute %q: expected Required=%v, got Required=%v", tc.attrName, tc.isRequired, attr.IsRequired())
			}
		})
	}
}

// TestPowerMethodMapping verifies the power action to BCM method mapping.
func TestPowerMethodMapping(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input    string
		expected string
	}{
		{"power_on", "powerOn"},
		{"power_off", "powerOff"},
		{"reboot", "reboot"},
		{"power_cycle", "powerCycle"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result, exists := powerMethodMapping[tc.input]
			if !exists {
				t.Fatalf("Expected mapping for %q to exist", tc.input)
			}
			if result != tc.expected {
				t.Errorf("Expected %q to map to %q, got %q", tc.input, tc.expected, result)
			}
		})
	}
}

// TestPowerMethodMapping_InvalidValues verifies invalid power actions are not mapped.
func TestPowerMethodMapping_InvalidValues(t *testing.T) {
	t.Parallel()

	invalidValues := []string{"invalid", "POWER_ON", "PowerOn", "shutdown", "restart", ""}

	for _, val := range invalidValues {
		t.Run(val, func(t *testing.T) {
			_, exists := powerMethodMapping[val]
			if exists {
				t.Errorf("Expected no mapping for invalid value %q", val)
			}
		})
	}
}

// TestCMDevicePowerAction_InterfaceCompliance verifies interface implementation.
func TestCMDevicePowerAction_InterfaceCompliance(t *testing.T) {
	t.Parallel()

	// Verify NewCMDevicePowerAction returns an action.Action
	var _ action.Action = NewCMDevicePowerAction()

	// Verify the concrete type implements ActionWithConfigure
	a := NewCMDevicePowerAction()
	_, ok := a.(action.ActionWithConfigure)
	if !ok {
		t.Error("CMDevicePowerAction should implement ActionWithConfigure interface")
	}
}

// TestCMDevicePowerAction_Configure_NilProviderData verifies Configure handles nil provider data.
func TestCMDevicePowerAction_Configure_NilProviderData(t *testing.T) {
	t.Parallel()

	a := &CMDevicePowerAction{}

	req := action.ConfigureRequest{
		ProviderData: nil,
	}
	resp := &action.ConfigureResponse{}

	// Should not panic and should not set errors
	a.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Expected no errors for nil ProviderData, got: %v", resp.Diagnostics)
	}
}

// TestCMDevicePowerAction_Configure_WrongType verifies Configure handles wrong type.
func TestCMDevicePowerAction_Configure_WrongType(t *testing.T) {
	t.Parallel()

	a := &CMDevicePowerAction{}

	// Pass wrong type (string instead of *BCMClient)
	req := action.ConfigureRequest{
		ProviderData: "wrong-type",
	}
	resp := &action.ConfigureResponse{}

	a.Configure(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("Expected error for wrong ProviderData type")
	}

	// Verify error message mentions expected type
	found := false
	for _, diag := range resp.Diagnostics.Errors() {
		if diag.Summary() == "Unexpected Action Configure Type" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'Unexpected Action Configure Type' error")
	}
}

// TestCMDevicePowerAction_PowerActionValidValues verifies all valid power_action values.
func TestCMDevicePowerAction_PowerActionValidValues(t *testing.T) {
	t.Parallel()

	validValues := []string{"power_on", "power_off", "reboot", "power_cycle"}

	for _, val := range validValues {
		_, exists := powerMethodMapping[val]
		if !exists {
			t.Errorf("Expected %q to be a valid power_action value", val)
		}
	}

	// Verify we have exactly 4 valid values
	if len(powerMethodMapping) != 4 {
		t.Errorf("Expected 4 power_action values, got %d", len(powerMethodMapping))
	}
}
