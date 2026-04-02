// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlattenCategoryObjectList(t *testing.T) {
	ctx := context.Background()

	t.Run("extracts preferred string field", func(t *testing.T) {
		result := flattenCategoryObjectList(ctx, map[string]interface{}{
			"roles": []interface{}{
				map[string]interface{}{"name": "backup", "uuid": "role-1"},
				map[string]interface{}{"name": "provisioning", "uuid": "role-2"},
			},
		}, "roles", "name")

		assert.False(t, result.IsNull())

		var values []string
		diags := result.ElementsAs(ctx, &values, false)
		require.False(t, diags.HasError())
		assert.Equal(t, []string{"backup", "provisioning"}, values)
	})

	t.Run("falls back to secondary field", func(t *testing.T) {
		result := flattenCategoryObjectList(ctx, map[string]interface{}{
			"fsmounts": []interface{}{
				map[string]interface{}{"path": "/cm/shared"},
				map[string]interface{}{"mountpoint": "/home"},
			},
		}, "fsmounts", "mountpoint", "path")

		assert.False(t, result.IsNull())

		var values []string
		diags := result.ElementsAs(ctx, &values, false)
		require.False(t, diags.HasError())
		assert.Equal(t, []string{"/cm/shared", "/home"}, values)
	})

	t.Run("preserves empty arrays as empty lists", func(t *testing.T) {
		result := flattenCategoryObjectList(ctx, map[string]interface{}{
			"services": []interface{}{},
		}, "services", "name")

		assert.False(t, result.IsNull())

		var values []string
		diags := result.ElementsAs(ctx, &values, false)
		require.False(t, diags.HasError())
		assert.Empty(t, values)
	})

	t.Run("returns null when field is absent", func(t *testing.T) {
		result := flattenCategoryObjectList(ctx, map[string]interface{}{}, "modules", "name")
		assert.True(t, result.IsNull())
	})
}
