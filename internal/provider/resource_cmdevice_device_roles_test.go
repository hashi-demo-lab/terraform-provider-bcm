// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Package provider contains unit tests for device role builders (KubeletRole, EtcdHostRole).
// These tests verify the entity construction for Kubernetes roles embedded in Device.roles[].
package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Unit Tests for KubeletRole Entity Builder (T045)
// =============================================================================

// TestKubeletRoleEntityBuilder tests the construction of KubeletRole entities
// for embedding in Device.roles[] array.
func TestKubeletRoleEntityBuilder(t *testing.T) {
	t.Run("build minimal kubelet role", func(t *testing.T) {
		// Given a minimal KubeletRoleModel with only required field
		model := KubeletRoleModel{
			KubeCluster: types.StringValue("kube-cluster-uuid-1234"),
		}

		// When building the entity
		entity, err := buildKubeletRoleEntity(context.Background(), model)

		// Then it should succeed with correct structure
		require.NoError(t, err)
		assert.Equal(t, "Role", entity["baseType"])
		assert.Equal(t, "KubeletRole", entity["childType"])
		assert.Equal(t, "kubelet", entity["name"])
		assert.Equal(t, "kube-cluster-uuid-1234", entity["kubeCluster"])
		assert.NotEmpty(t, entity["uuid"], "UUID should be generated")

		// Default values should be applied
		assert.Equal(t, true, entity["controlPlane"])
		assert.Equal(t, true, entity["worker"])
		assert.Equal(t, "docker.service", entity["containerRuntimeService"])
		assert.Equal(t, int64(110), entity["maxPods"])
	})

	t.Run("build kubelet role with all fields", func(t *testing.T) {
		// Given a complete KubeletRoleModel
		model := KubeletRoleModel{
			UUID:                    types.StringValue("existing-role-uuid"),
			KubeCluster:             types.StringValue("kube-cluster-uuid"),
			ControlPlane:            types.BoolValue(true),
			Worker:                  types.BoolValue(false),
			ContainerRuntimeService: types.StringValue("containerd.service"),
			MaxPods:                 types.Int64Value(200),
			Options:                 types.StringValue(`{"image-gc-high-threshold": "85"}`),
			CustomYAML:              types.StringValue("apiVersion: kubelet.config.k8s.io/v1beta1"),
		}

		// When building the entity
		entity, err := buildKubeletRoleEntity(context.Background(), model)

		// Then it should use all provided values
		require.NoError(t, err)
		assert.Equal(t, "existing-role-uuid", entity["uuid"])
		assert.Equal(t, "kube-cluster-uuid", entity["kubeCluster"])
		assert.Equal(t, true, entity["controlPlane"])
		assert.Equal(t, false, entity["worker"])
		assert.Equal(t, "containerd.service", entity["containerRuntimeService"])
		assert.Equal(t, int64(200), entity["maxPods"])
		assert.Equal(t, `{"image-gc-high-threshold": "85"}`, entity["options"])
		assert.Equal(t, "apiVersion: kubelet.config.k8s.io/v1beta1", entity["customYaml"])
	})

	t.Run("build kubelet role control plane only", func(t *testing.T) {
		// Given a control plane only node
		model := KubeletRoleModel{
			KubeCluster:  types.StringValue("kube-cluster-uuid"),
			ControlPlane: types.BoolValue(true),
			Worker:       types.BoolValue(false),
		}

		// When building the entity
		entity, err := buildKubeletRoleEntity(context.Background(), model)

		// Then control_plane should be true, worker should be false
		require.NoError(t, err)
		assert.Equal(t, true, entity["controlPlane"])
		assert.Equal(t, false, entity["worker"])
	})

	t.Run("build kubelet role worker only", func(t *testing.T) {
		// Given a worker only node
		model := KubeletRoleModel{
			KubeCluster:  types.StringValue("kube-cluster-uuid"),
			ControlPlane: types.BoolValue(false),
			Worker:       types.BoolValue(true),
		}

		// When building the entity
		entity, err := buildKubeletRoleEntity(context.Background(), model)

		// Then control_plane should be false, worker should be true
		require.NoError(t, err)
		assert.Equal(t, false, entity["controlPlane"])
		assert.Equal(t, true, entity["worker"])
	})

	t.Run("kubelet role generates new UUID when not provided", func(t *testing.T) {
		// Given a model without UUID
		model := KubeletRoleModel{
			KubeCluster: types.StringValue("kube-cluster-uuid"),
		}

		// When building the entity twice
		entity1, _ := buildKubeletRoleEntity(context.Background(), model)
		entity2, _ := buildKubeletRoleEntity(context.Background(), model)

		// Then UUIDs should be different (new UUIDs generated)
		assert.NotEmpty(t, entity1["uuid"])
		assert.NotEmpty(t, entity2["uuid"])
		assert.NotEqual(t, entity1["uuid"], entity2["uuid"], "Each call should generate new UUID")
	})

	t.Run("kubelet role preserves existing UUID", func(t *testing.T) {
		// Given a model with existing UUID
		existingUUID := "preserve-this-uuid"
		model := KubeletRoleModel{
			UUID:        types.StringValue(existingUUID),
			KubeCluster: types.StringValue("kube-cluster-uuid"),
		}

		// When building the entity
		entity, _ := buildKubeletRoleEntity(context.Background(), model)

		// Then the existing UUID should be preserved
		assert.Equal(t, existingUUID, entity["uuid"])
	})
}

// =============================================================================
// Unit Tests for EtcdHostRole Entity Builder (T046)
// =============================================================================

// TestEtcdHostRoleEntityBuilder tests the construction of EtcdHostRole entities
// for embedding in Device.roles[] array.
func TestEtcdHostRoleEntityBuilder(t *testing.T) {
	t.Run("build minimal etcd host role", func(t *testing.T) {
		// Given a minimal EtcdHostRoleModel with only required field
		model := EtcdHostRoleModel{
			EtcdCluster: types.StringValue("etcd-cluster-uuid-1234"),
		}

		// When building the entity
		entity, err := buildEtcdHostRoleEntity(context.Background(), model)

		// Then it should succeed with correct structure
		require.NoError(t, err)
		assert.Equal(t, "Role", entity["baseType"])
		assert.Equal(t, "EtcdHostRole", entity["childType"])
		assert.Equal(t, "etcdhost", entity["name"])
		assert.Equal(t, "etcd-cluster-uuid-1234", entity["etcdCluster"])
		assert.NotEmpty(t, entity["uuid"], "UUID should be generated")

		// Default values should be applied
		assert.Equal(t, "$hostname", entity["memberName"])
		assert.Equal(t, "/var/lib/etcd", entity["spool"])
		assert.Equal(t, int64(100000), entity["snapshotCount"])
		assert.Equal(t, int64(5), entity["maxSnapshots"])
	})

	t.Run("build etcd host role with all fields", func(t *testing.T) {
		// Given a complete EtcdHostRoleModel
		model := EtcdHostRoleModel{
			UUID:                types.StringValue("existing-etcd-role-uuid"),
			EtcdCluster:         types.StringValue("etcd-cluster-uuid"),
			MemberName:          types.StringValue("etcd-node-1"),
			Spool:               types.StringValue("/data/etcd"),
			ListenClientURLs:    createStringList([]string{"https://0.0.0.0:2379"}),
			ListenPeerURLs:      createStringList([]string{"https://0.0.0.0:2380"}),
			AdvertiseClientURLs: createStringList([]string{"https://192.168.1.10:2379"}),
			AdvertisePeerURLs:   createStringList([]string{"https://192.168.1.10:2380"}),
			SnapshotCount:       types.Int64Value(50000),
			MaxSnapshots:        types.Int64Value(10),
		}

		// When building the entity
		entity, err := buildEtcdHostRoleEntity(context.Background(), model)

		// Then it should use all provided values
		require.NoError(t, err)
		assert.Equal(t, "existing-etcd-role-uuid", entity["uuid"])
		assert.Equal(t, "etcd-cluster-uuid", entity["etcdCluster"])
		assert.Equal(t, "etcd-node-1", entity["memberName"])
		assert.Equal(t, "/data/etcd", entity["spool"])
		assert.Equal(t, int64(50000), entity["snapshotCount"])
		assert.Equal(t, int64(10), entity["maxSnapshots"])

		// Check list fields
		listenClientURLs, ok := entity["listenClientUrls"].([]string)
		require.True(t, ok, "listenClientUrls should be []string")
		assert.Equal(t, []string{"https://0.0.0.0:2379"}, listenClientURLs)
	})

	t.Run("etcd host role generates new UUID when not provided", func(t *testing.T) {
		// Given a model without UUID
		model := EtcdHostRoleModel{
			EtcdCluster: types.StringValue("etcd-cluster-uuid"),
		}

		// When building the entity twice
		entity1, _ := buildEtcdHostRoleEntity(context.Background(), model)
		entity2, _ := buildEtcdHostRoleEntity(context.Background(), model)

		// Then UUIDs should be different (new UUIDs generated)
		assert.NotEmpty(t, entity1["uuid"])
		assert.NotEmpty(t, entity2["uuid"])
		assert.NotEqual(t, entity1["uuid"], entity2["uuid"], "Each call should generate new UUID")
	})

	t.Run("etcd host role preserves existing UUID", func(t *testing.T) {
		// Given a model with existing UUID
		existingUUID := "preserve-etcd-uuid"
		model := EtcdHostRoleModel{
			UUID:        types.StringValue(existingUUID),
			EtcdCluster: types.StringValue("etcd-cluster-uuid"),
		}

		// When building the entity
		entity, _ := buildEtcdHostRoleEntity(context.Background(), model)

		// Then the existing UUID should be preserved
		assert.Equal(t, existingUUID, entity["uuid"])
	})
}

// =============================================================================
// Unit Tests for Device Role Merging (T047)
// =============================================================================

// TestDeviceRoleMerging tests the role merging logic that combines Kubernetes roles
// with existing non-Kubernetes roles in the Device.roles[] array.
func TestDeviceRoleMerging(t *testing.T) {
	t.Run("merge kubelet role with empty existing roles", func(t *testing.T) {
		// Given no existing roles and a new kubelet role
		existingRoles := []map[string]interface{}{}
		newKubeletRole := map[string]interface{}{
			"baseType":    "Role",
			"childType":   "KubeletRole",
			"uuid":        "kubelet-uuid",
			"kubeCluster": "kube-cluster-uuid",
		}

		// When merging
		result := mergeDeviceRoles(existingRoles, []map[string]interface{}{newKubeletRole}, nil)

		// Then result should contain only the new role
		require.Len(t, result, 1)
		assert.Equal(t, "KubeletRole", result[0]["childType"])
	})

	t.Run("merge kubelet role preserves non-kubernetes roles", func(t *testing.T) {
		// Given existing non-Kubernetes roles
		existingRoles := []map[string]interface{}{
			{"baseType": "Role", "childType": "BackupRole", "uuid": "backup-uuid", "name": "backup"},
			{"baseType": "Role", "childType": "ProvisioningRole", "uuid": "prov-uuid", "name": "provisioning"},
		}
		newKubeletRole := map[string]interface{}{
			"baseType":    "Role",
			"childType":   "KubeletRole",
			"uuid":        "kubelet-uuid",
			"kubeCluster": "kube-cluster-uuid",
		}

		// When merging
		result := mergeDeviceRoles(existingRoles, []map[string]interface{}{newKubeletRole}, nil)

		// Then result should contain both non-Kubernetes roles and new Kubernetes role
		require.Len(t, result, 3)

		// Verify non-Kubernetes roles are preserved
		childTypes := make([]string, len(result))
		for i, r := range result {
			if ct, ok := r["childType"].(string); ok {
				childTypes[i] = ct
			}
		}
		assert.Contains(t, childTypes, "BackupRole")
		assert.Contains(t, childTypes, "ProvisioningRole")
		assert.Contains(t, childTypes, "KubeletRole")
	})

	t.Run("merge replaces existing kubelet role for same cluster preserving UUID", func(t *testing.T) {
		// Given existing KubeletRole for the same cluster
		existingRoles := []map[string]interface{}{
			{
				"baseType":    "Role",
				"childType":   "KubeletRole",
				"uuid":        "old-kubelet-uuid",
				"kubeCluster": "kube-cluster-uuid",
			},
		}
		newKubeletRole := map[string]interface{}{
			"baseType":    "Role",
			"childType":   "KubeletRole",
			"uuid":        "new-kubelet-uuid",
			"kubeCluster": "kube-cluster-uuid",
		}

		// When merging
		result := mergeDeviceRoles(existingRoles, []map[string]interface{}{newKubeletRole}, nil)

		// Then result should contain only one role (old one replaced)
		// But UUID should be preserved from the existing role (T058)
		require.Len(t, result, 1)
		assert.Equal(t, "old-kubelet-uuid", result[0]["uuid"], "UUID should be preserved from existing role")
	})

	t.Run("merge preserves kubelet role for different cluster", func(t *testing.T) {
		// Given existing KubeletRole for a different cluster
		existingRoles := []map[string]interface{}{
			{
				"baseType":    "Role",
				"childType":   "KubeletRole",
				"uuid":        "other-kubelet-uuid",
				"kubeCluster": "other-kube-cluster-uuid",
			},
		}
		newKubeletRole := map[string]interface{}{
			"baseType":    "Role",
			"childType":   "KubeletRole",
			"uuid":        "new-kubelet-uuid",
			"kubeCluster": "new-kube-cluster-uuid",
		}

		// When merging
		result := mergeDeviceRoles(existingRoles, []map[string]interface{}{newKubeletRole}, nil)

		// Then result should contain both roles (different clusters)
		require.Len(t, result, 2)
	})

	t.Run("merge both kubelet and etcd roles", func(t *testing.T) {
		// Given no existing roles
		existingRoles := []map[string]interface{}{}
		kubeletRole := map[string]interface{}{
			"baseType":    "Role",
			"childType":   "KubeletRole",
			"uuid":        "kubelet-uuid",
			"kubeCluster": "kube-cluster-uuid",
		}
		etcdRole := map[string]interface{}{
			"baseType":    "Role",
			"childType":   "EtcdHostRole",
			"uuid":        "etcd-uuid",
			"etcdCluster": "etcd-cluster-uuid",
		}

		// When merging both
		result := mergeDeviceRoles(existingRoles, []map[string]interface{}{kubeletRole}, []map[string]interface{}{etcdRole})

		// Then result should contain both roles
		require.Len(t, result, 2)
	})

	t.Run("remove kubelet role when nil in terraform config", func(t *testing.T) {
		// Given existing KubeletRole but nil in new config
		existingRoles := []map[string]interface{}{
			{
				"baseType":    "Role",
				"childType":   "KubeletRole",
				"uuid":        "kubelet-uuid",
				"kubeCluster": "kube-cluster-uuid",
			},
			{
				"baseType":  "Role",
				"childType": "BackupRole",
				"uuid":      "backup-uuid",
			},
		}

		// When merging with explicit removal (empty slice for kubelet roles)
		// Note: nil means "don't touch", empty slice means "remove all"
		result := mergeDeviceRoles(existingRoles, []map[string]interface{}{}, nil)

		// Then KubeletRole should be removed, BackupRole should remain
		require.Len(t, result, 1)
		assert.Equal(t, "BackupRole", result[0]["childType"])
	})
}

// =============================================================================
// Unit Tests for UUID Preservation During Role Updates (T058)
// =============================================================================

// TestDeviceRoleUUIDPreservation tests that UUIDs are preserved when updating
// roles for the same cluster reference.
func TestDeviceRoleUUIDPreservation(t *testing.T) {
	t.Run("preserve kubelet role UUID on update for same cluster", func(t *testing.T) {
		// Given an existing KubeletRole with a UUID
		existingRoles := []map[string]any{
			{
				"baseType":     "Role",
				"childType":    "KubeletRole",
				"uuid":         "existing-kubelet-uuid-12345",
				"kubeCluster":  "kube-cluster-uuid",
				"controlPlane": true,
				"worker":       false,
			},
		}

		// And a new KubeletRole for the same cluster with different settings but new UUID
		newKubeletRole := map[string]any{
			"baseType":     "Role",
			"childType":    "KubeletRole",
			"uuid":         "newly-generated-uuid", // This would be from buildKubeletRoleEntity
			"kubeCluster":  "kube-cluster-uuid",
			"controlPlane": false, // Changed setting
			"worker":       true,  // Changed setting
		}

		// When merging
		result := mergeDeviceRoles(existingRoles, []map[string]any{newKubeletRole}, nil)

		// Then the existing UUID should be preserved (not replaced with new one)
		require.Len(t, result, 1)
		assert.Equal(t, "existing-kubelet-uuid-12345", result[0]["uuid"],
			"UUID should be preserved from existing role for same cluster")

		// And the new settings should be applied
		assert.Equal(t, false, result[0]["controlPlane"])
		assert.Equal(t, true, result[0]["worker"])
	})

	t.Run("preserve etcd host role UUID on update for same cluster", func(t *testing.T) {
		// Given an existing EtcdHostRole with a UUID
		existingRoles := []map[string]any{
			{
				"baseType":    "Role",
				"childType":   "EtcdHostRole",
				"uuid":        "existing-etcd-uuid-67890",
				"etcdCluster": "etcd-cluster-uuid",
				"memberName":  "etcd-node-1",
			},
		}

		// And a new EtcdHostRole for the same cluster with different settings
		newEtcdRole := map[string]any{
			"baseType":    "Role",
			"childType":   "EtcdHostRole",
			"uuid":        "newly-generated-uuid",
			"etcdCluster": "etcd-cluster-uuid",
			"memberName":  "etcd-node-renamed", // Changed setting
		}

		// When merging
		result := mergeDeviceRoles(existingRoles, nil, []map[string]any{newEtcdRole})

		// Then the existing UUID should be preserved
		require.Len(t, result, 1)
		assert.Equal(t, "existing-etcd-uuid-67890", result[0]["uuid"],
			"UUID should be preserved from existing role for same cluster")

		// And the new settings should be applied
		assert.Equal(t, "etcd-node-renamed", result[0]["memberName"])
	})

	t.Run("preserve UUIDs for multiple roles on update", func(t *testing.T) {
		// Given existing roles for both types
		existingRoles := []map[string]any{
			{
				"baseType":    "Role",
				"childType":   "KubeletRole",
				"uuid":        "kubelet-uuid-preserve",
				"kubeCluster": "kube-cluster-1",
			},
			{
				"baseType":    "Role",
				"childType":   "EtcdHostRole",
				"uuid":        "etcd-uuid-preserve",
				"etcdCluster": "etcd-cluster-1",
			},
			{
				"baseType":  "Role",
				"childType": "BackupRole",
				"uuid":      "backup-uuid",
			},
		}

		// And new roles for the same clusters
		newKubeletRole := map[string]any{
			"baseType":    "Role",
			"childType":   "KubeletRole",
			"uuid":        "new-gen-kubelet",
			"kubeCluster": "kube-cluster-1",
		}
		newEtcdRole := map[string]any{
			"baseType":    "Role",
			"childType":   "EtcdHostRole",
			"uuid":        "new-gen-etcd",
			"etcdCluster": "etcd-cluster-1",
		}

		// When merging
		result := mergeDeviceRoles(existingRoles, []map[string]any{newKubeletRole}, []map[string]any{newEtcdRole})

		// Then all UUIDs should be preserved appropriately
		require.Len(t, result, 3)

		// Find each role in result
		var foundKubelet, foundEtcd, foundBackup bool
		for _, r := range result {
			switch r["childType"] {
			case "KubeletRole":
				assert.Equal(t, "kubelet-uuid-preserve", r["uuid"], "KubeletRole UUID should be preserved")
				foundKubelet = true
			case "EtcdHostRole":
				assert.Equal(t, "etcd-uuid-preserve", r["uuid"], "EtcdHostRole UUID should be preserved")
				foundEtcd = true
			case "BackupRole":
				assert.Equal(t, "backup-uuid", r["uuid"], "Non-K8s role UUID unchanged")
				foundBackup = true
			}
		}
		assert.True(t, foundKubelet && foundEtcd && foundBackup, "All roles should be present")
	})

	t.Run("new cluster gets new UUID not preserved", func(t *testing.T) {
		// Given existing KubeletRole for cluster-1
		existingRoles := []map[string]any{
			{
				"baseType":    "Role",
				"childType":   "KubeletRole",
				"uuid":        "cluster1-kubelet-uuid",
				"kubeCluster": "kube-cluster-1",
			},
		}

		// And a new KubeletRole for cluster-2 (different cluster)
		newKubeletRole := map[string]any{
			"baseType":    "Role",
			"childType":   "KubeletRole",
			"uuid":        "cluster2-new-uuid",
			"kubeCluster": "kube-cluster-2", // Different cluster
		}

		// When merging
		result := mergeDeviceRoles(existingRoles, []map[string]any{newKubeletRole}, nil)

		// Then both roles should exist
		require.Len(t, result, 2)

		// cluster-1 role should have its original UUID
		// cluster-2 role should keep its new UUID (no existing to preserve from)
		for _, r := range result {
			switch r["kubeCluster"] {
			case "kube-cluster-1":
				assert.Equal(t, "cluster1-kubelet-uuid", r["uuid"])
			case "kube-cluster-2":
				assert.Equal(t, "cluster2-new-uuid", r["uuid"],
					"New cluster should keep its generated UUID")
			}
		}
	})
}

// =============================================================================
// Test Helpers
// =============================================================================

// createStringList creates a types.List from a slice of strings for testing.
func createStringList(values []string) types.List {
	if len(values) == 0 {
		return types.ListNull(types.StringType)
	}
	listValues := make([]types.String, len(values))
	for i, v := range values {
		listValues[i] = types.StringValue(v)
	}
	list, _ := types.ListValueFrom(context.Background(), types.StringType, listValues)
	return list
}
