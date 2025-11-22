// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestGetStringValue tests the getStringValue utility function with various input scenarios.
func TestGetStringValue(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		key      string
		expected types.String
	}{
		{
			name:     "valid string value",
			data:     map[string]interface{}{"name": "test-value"},
			key:      "name",
			expected: types.StringValue("test-value"),
		},
		{
			name:     "empty string returns null",
			data:     map[string]interface{}{"name": ""},
			key:      "name",
			expected: types.StringNull(),
		},
		{
			name:     "missing key returns null",
			data:     map[string]interface{}{"other": "value"},
			key:      "name",
			expected: types.StringNull(),
		},
		{
			name:     "nil value returns null",
			data:     map[string]interface{}{"name": nil},
			key:      "name",
			expected: types.StringNull(),
		},
		{
			name:     "wrong type (int) returns null",
			data:     map[string]interface{}{"name": 123},
			key:      "name",
			expected: types.StringNull(),
		},
		{
			name:     "wrong type (bool) returns null",
			data:     map[string]interface{}{"name": true},
			key:      "name",
			expected: types.StringNull(),
		},
		{
			name:     "wrong type (float64) returns null",
			data:     map[string]interface{}{"name": 123.45},
			key:      "name",
			expected: types.StringNull(),
		},
		{
			name:     "empty map returns null",
			data:     map[string]interface{}{},
			key:      "name",
			expected: types.StringNull(),
		},
		{
			name:     "nil map returns null",
			data:     nil,
			key:      "name",
			expected: types.StringNull(),
		},
		{
			name:     "string with spaces",
			data:     map[string]interface{}{"description": "test description with spaces"},
			key:      "description",
			expected: types.StringValue("test description with spaces"),
		},
		{
			name:     "string with special characters",
			data:     map[string]interface{}{"path": "/path/to/file-123_test.img"},
			key:      "path",
			expected: types.StringValue("/path/to/file-123_test.img"),
		},
		{
			name:     "whitespace-only string preserved",
			data:     map[string]interface{}{"notes": "   "},
			key:      "notes",
			expected: types.StringValue("   "),
		},
		{
			name:     "uuid zero value is valid string",
			data:     map[string]interface{}{"uuid": "00000000-0000-0000-0000-000000000000"},
			key:      "uuid",
			expected: types.StringValue("00000000-0000-0000-0000-000000000000"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStringValue(tt.data, tt.key)
			if !result.Equal(tt.expected) {
				t.Errorf("getStringValue() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestGetBoolValue tests the getBoolValue utility function with various input scenarios.
func TestGetBoolValue(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		key      string
		expected types.Bool
	}{
		{
			name:     "valid true boolean",
			data:     map[string]interface{}{"enabled": true},
			key:      "enabled",
			expected: types.BoolValue(true),
		},
		{
			name:     "valid false boolean",
			data:     map[string]interface{}{"enabled": false},
			key:      "enabled",
			expected: types.BoolValue(false),
		},
		{
			name:     "missing key returns null",
			data:     map[string]interface{}{"other": true},
			key:      "enabled",
			expected: types.BoolNull(),
		},
		{
			name:     "nil value returns null",
			data:     map[string]interface{}{"enabled": nil},
			key:      "enabled",
			expected: types.BoolNull(),
		},
		{
			name:     "wrong type (string 'true') returns null",
			data:     map[string]interface{}{"enabled": "true"},
			key:      "enabled",
			expected: types.BoolNull(),
		},
		{
			name:     "wrong type (string 'false') returns null",
			data:     map[string]interface{}{"enabled": "false"},
			key:      "enabled",
			expected: types.BoolNull(),
		},
		{
			name:     "wrong type (int 1) returns null",
			data:     map[string]interface{}{"enabled": 1},
			key:      "enabled",
			expected: types.BoolNull(),
		},
		{
			name:     "wrong type (int 0) returns null",
			data:     map[string]interface{}{"enabled": 0},
			key:      "enabled",
			expected: types.BoolNull(),
		},
		{
			name:     "wrong type (float64) returns null",
			data:     map[string]interface{}{"enabled": 1.0},
			key:      "enabled",
			expected: types.BoolNull(),
		},
		{
			name:     "empty map returns null",
			data:     map[string]interface{}{},
			key:      "enabled",
			expected: types.BoolNull(),
		},
		{
			name:     "nil map returns null",
			data:     nil,
			key:      "enabled",
			expected: types.BoolNull(),
		},
		{
			name:     "false is not null",
			data:     map[string]interface{}{"modified": false},
			key:      "modified",
			expected: types.BoolValue(false),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBoolValue(tt.data, tt.key)
			if !result.Equal(tt.expected) {
				t.Errorf("getBoolValue() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestGetInt64Value tests the getInt64Value utility function with various input scenarios.
func TestGetInt64Value(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		key      string
		expected types.Int64
	}{
		{
			name:     "valid int64 value",
			data:     map[string]interface{}{"port": int64(8081)},
			key:      "port",
			expected: types.Int64Value(8081),
		},
		{
			name:     "valid int value",
			data:     map[string]interface{}{"port": 8081},
			key:      "port",
			expected: types.Int64Value(8081),
		},
		{
			name:     "valid float64 value (JSON unmarshal default)",
			data:     map[string]interface{}{"port": 8081.0},
			key:      "port",
			expected: types.Int64Value(8081),
		},
		{
			name:     "float64 with decimal gets truncated",
			data:     map[string]interface{}{"value": 123.99},
			key:      "value",
			expected: types.Int64Value(123),
		},
		{
			name:     "zero value int is valid not null",
			data:     map[string]interface{}{"count": 0},
			key:      "count",
			expected: types.Int64Value(0),
		},
		{
			name:     "zero value float64 is valid not null",
			data:     map[string]interface{}{"count": 0.0},
			key:      "count",
			expected: types.Int64Value(0),
		},
		{
			name:     "negative int value",
			data:     map[string]interface{}{"offset": -10},
			key:      "offset",
			expected: types.Int64Value(-10),
		},
		{
			name:     "negative float64 value",
			data:     map[string]interface{}{"offset": -10.0},
			key:      "offset",
			expected: types.Int64Value(-10),
		},
		{
			name:     "large int64 value (unix timestamp)",
			data:     map[string]interface{}{"creationTime": int64(1732234567)},
			key:      "creationTime",
			expected: types.Int64Value(1732234567),
		},
		{
			name:     "large float64 value (unix timestamp from JSON)",
			data:     map[string]interface{}{"creationTime": float64(1732234567)},
			key:      "creationTime",
			expected: types.Int64Value(1732234567),
		},
		{
			name:     "missing key returns null",
			data:     map[string]interface{}{"other": 123},
			key:      "port",
			expected: types.Int64Null(),
		},
		{
			name:     "nil value returns null",
			data:     map[string]interface{}{"port": nil},
			key:      "port",
			expected: types.Int64Null(),
		},
		{
			name:     "wrong type (string number) returns null",
			data:     map[string]interface{}{"port": "8081"},
			key:      "port",
			expected: types.Int64Null(),
		},
		{
			name:     "wrong type (bool) returns null",
			data:     map[string]interface{}{"port": true},
			key:      "port",
			expected: types.Int64Null(),
		},
		{
			name:     "empty map returns null",
			data:     map[string]interface{}{},
			key:      "port",
			expected: types.Int64Null(),
		},
		{
			name:     "nil map returns null",
			data:     nil,
			key:      "port",
			expected: types.Int64Null(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getInt64Value(tt.data, tt.key)
			if !result.Equal(tt.expected) {
				t.Errorf("getInt64Value() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
