// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Package provider contains entity builder helpers for BCM Kubernetes role types.
// These builders construct BCM-compatible entity structures for KubeletRole and EtcdHostRole.
package provider

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// =============================================================================
// KubeletRole Entity Builder
// =============================================================================

// BuildKubeletRoleEntity constructs a BCM KubeletRole entity from Terraform model.
// KubeletRole is embedded in Device.roles[] array with childType="KubeletRole".
//
// BCM API Entity Structure:
//
//	{
//	  "baseType": "Role",
//	  "childType": "KubeletRole",
//	  "uuid": "...", // Provider-generated
//	  "name": "kubelet",
//	  "kubeCluster": "...", // UUID reference
//	  "controlPlane": true,
//	  "worker": true,
//	  "containerRuntimeService": "docker.service",
//	  "maxPods": 110,
//	  "options": {},
//	  "customYaml": "",
//	  "modified": true,
//	  "to_be_removed": false
//	}
func BuildKubeletRoleEntity(ctx context.Context, model *KubeletRoleModel, diags *diag.Diagnostics) map[string]interface{} {
	// Start with base role structure
	entity := map[string]interface{}{
		"baseType":      "Role",
		"childType":     "KubeletRole",
		"name":          "kubelet",
		"modified":      true,
		"to_be_removed": false,
		"revision":      "",
	}

	// Set UUID - generate if not provided (new role)
	if model.UUID.IsNull() || model.UUID.IsUnknown() || model.UUID.ValueString() == "" {
		entity["uuid"] = generateUUID()
	} else {
		entity["uuid"] = model.UUID.ValueString()
	}

	// Required field: kubeCluster reference
	if !model.KubeCluster.IsNull() && !model.KubeCluster.IsUnknown() {
		entity["kubeCluster"] = model.KubeCluster.ValueString()
	} else {
		diags.AddError("Missing Required Field", "kube_cluster is required for kubelet_role")
		return nil
	}

	// Optional fields with defaults
	if !model.ControlPlane.IsNull() && !model.ControlPlane.IsUnknown() {
		entity["controlPlane"] = model.ControlPlane.ValueBool()
	} else {
		entity["controlPlane"] = true // Default
	}

	if !model.Worker.IsNull() && !model.Worker.IsUnknown() {
		entity["worker"] = model.Worker.ValueBool()
	} else {
		entity["worker"] = true // Default
	}

	if !model.ContainerRuntimeService.IsNull() && !model.ContainerRuntimeService.IsUnknown() {
		entity["containerRuntimeService"] = model.ContainerRuntimeService.ValueString()
	} else {
		entity["containerRuntimeService"] = "docker.service" // Default
	}

	if !model.MaxPods.IsNull() && !model.MaxPods.IsUnknown() {
		entity["maxPods"] = model.MaxPods.ValueInt64()
	} else {
		entity["maxPods"] = int64(110) // Default
	}

	// Options field - JSON-encoded string to object
	if !model.Options.IsNull() && !model.Options.IsUnknown() && model.Options.ValueString() != "" {
		var options map[string]interface{}
		if err := json.Unmarshal([]byte(model.Options.ValueString()), &options); err != nil {
			diags.AddWarning("Options Parse Warning", "Failed to parse options JSON, using empty object")
			entity["options"] = map[string]interface{}{}
		} else {
			entity["options"] = options
		}
	} else {
		entity["options"] = map[string]interface{}{}
	}

	// Custom YAML
	if !model.CustomYAML.IsNull() && !model.CustomYAML.IsUnknown() {
		entity["customYaml"] = model.CustomYAML.ValueString()
	} else {
		entity["customYaml"] = ""
	}

	return entity
}

// ParseKubeletRoleFromBCM parses a BCM KubeletRole entity into a Terraform model.
// Used when reading device roles from BCM API response.
func ParseKubeletRoleFromBCM(ctx context.Context, data map[string]interface{}) *KubeletRoleModel {
	model := &KubeletRoleModel{}

	model.UUID = getStringValue(data, "uuid")
	model.KubeCluster = getStringValue(data, "kubeCluster")
	model.ControlPlane = getBoolValue(data, "controlPlane")
	model.Worker = getBoolValue(data, "worker")
	model.ContainerRuntimeService = getStringValue(data, "containerRuntimeService")
	model.MaxPods = getInt64Value(data, "maxPods")
	model.CustomYAML = getStringValue(data, "customYaml")

	// Options - convert object to JSON string
	if options, ok := data["options"]; ok && options != nil {
		if optMap, ok := options.(map[string]interface{}); ok && len(optMap) > 0 {
			optBytes, err := json.Marshal(optMap)
			if err == nil {
				model.Options = types.StringValue(string(optBytes))
			} else {
				model.Options = types.StringNull()
			}
		} else {
			model.Options = types.StringNull()
		}
	} else {
		model.Options = types.StringNull()
	}

	return model
}

// =============================================================================
// EtcdHostRole Entity Builder
// =============================================================================

// BuildEtcdHostRoleEntity constructs a BCM EtcdHostRole entity from Terraform model.
// EtcdHostRole is embedded in Device.roles[] array with childType="EtcdHostRole".
//
// BCM API Entity Structure:
//
//	{
//	  "baseType": "Role",
//	  "childType": "EtcdHostRole",
//	  "uuid": "...", // Provider-generated
//	  "name": "etcdhost",
//	  "etcdCluster": "...", // UUID reference
//	  "memberName": "$hostname",
//	  "spool": "/var/lib/etcd",
//	  "listenClientUrls": ["https://0.0.0.0:2379"],
//	  "listenPeerUrls": ["https://0.0.0.0:2380"],
//	  "advertiseClientUrls": ["https://$ip:2379"],
//	  "advertisePeerUrls": ["https://$ip:2380"],
//	  "snapshotCount": 100000,
//	  "maxSnapshots": 5,
//	  "modified": true,
//	  "to_be_removed": false
//	}
func BuildEtcdHostRoleEntity(ctx context.Context, model *EtcdHostRoleModel, diags *diag.Diagnostics) map[string]interface{} {
	// Start with base role structure
	entity := map[string]interface{}{
		"baseType":      "Role",
		"childType":     "EtcdHostRole",
		"name":          "etcdhost",
		"modified":      true,
		"to_be_removed": false,
		"revision":      "",
	}

	// Set UUID - generate if not provided (new role)
	if model.UUID.IsNull() || model.UUID.IsUnknown() || model.UUID.ValueString() == "" {
		entity["uuid"] = generateUUID()
	} else {
		entity["uuid"] = model.UUID.ValueString()
	}

	// Required field: etcdCluster reference
	if !model.EtcdCluster.IsNull() && !model.EtcdCluster.IsUnknown() {
		entity["etcdCluster"] = model.EtcdCluster.ValueString()
	} else {
		diags.AddError("Missing Required Field", "etcd_cluster is required for etcd_host_role")
		return nil
	}

	// Optional fields with defaults
	if !model.MemberName.IsNull() && !model.MemberName.IsUnknown() {
		entity["memberName"] = model.MemberName.ValueString()
	} else {
		entity["memberName"] = "$hostname" // Default - uses BCM variable
	}

	if !model.Spool.IsNull() && !model.Spool.IsUnknown() {
		entity["spool"] = model.Spool.ValueString()
	} else {
		entity["spool"] = "/var/lib/etcd" // Default
	}

	// Listen/Advertise URLs - convert Terraform list to []string
	entity["listenClientUrls"] = extractStringList(ctx, model.ListenClientURLs, diags, []string{"https://0.0.0.0:2379"})
	entity["listenPeerUrls"] = extractStringList(ctx, model.ListenPeerURLs, diags, []string{"https://0.0.0.0:2380"})
	entity["advertiseClientUrls"] = extractStringList(ctx, model.AdvertiseClientURLs, diags, []string{"https://$ip:2379"})
	entity["advertisePeerUrls"] = extractStringList(ctx, model.AdvertisePeerURLs, diags, []string{"https://$ip:2380"})

	// Snapshot settings
	if !model.SnapshotCount.IsNull() && !model.SnapshotCount.IsUnknown() {
		entity["snapshotCount"] = model.SnapshotCount.ValueInt64()
	} else {
		entity["snapshotCount"] = int64(100000) // Default
	}

	if !model.MaxSnapshots.IsNull() && !model.MaxSnapshots.IsUnknown() {
		entity["maxSnapshots"] = model.MaxSnapshots.ValueInt64()
	} else {
		entity["maxSnapshots"] = int64(5) // Default
	}

	return entity
}

// ParseEtcdHostRoleFromBCM parses a BCM EtcdHostRole entity into a Terraform model.
// Used when reading device roles from BCM API response.
func ParseEtcdHostRoleFromBCM(ctx context.Context, data map[string]interface{}) *EtcdHostRoleModel {
	model := &EtcdHostRoleModel{}

	model.UUID = getStringValue(data, "uuid")
	model.EtcdCluster = getStringValue(data, "etcdCluster")
	model.MemberName = getStringValue(data, "memberName")
	model.Spool = getStringValue(data, "spool")
	model.SnapshotCount = getInt64Value(data, "snapshotCount")
	model.MaxSnapshots = getInt64Value(data, "maxSnapshots")

	// URL lists
	model.ListenClientURLs = GetStringListValue(data, "listenClientUrls")
	model.ListenPeerURLs = GetStringListValue(data, "listenPeerUrls")
	model.AdvertiseClientURLs = GetStringListValue(data, "advertiseClientUrls")
	model.AdvertisePeerURLs = GetStringListValue(data, "advertisePeerUrls")

	return model
}

// =============================================================================
// Device Role Merging Logic
// =============================================================================

// MergeKubernetesRoles merges Kubernetes roles with existing device roles.
// Preserves non-Kubernetes roles (any role where childType is not KubeletRole or EtcdHostRole).
// Replaces Kubernetes roles for the clusters referenced in the Terraform config.
//
// Parameters:
//   - existingRoles: Current roles array from BCM device entity
//   - kubeletRole: KubeletRole entity to add (nil if not configured)
//   - etcdHostRole: EtcdHostRole entity to add (nil if not configured)
//   - kubeClusterUUID: KubeCluster UUID being managed (for role replacement)
//   - etcdClusterUUID: EtcdCluster UUID being managed (for role replacement)
//
// Returns:
//   - Merged roles array ready for device update
func MergeKubernetesRoles(
	existingRoles []interface{},
	kubeletRole map[string]interface{},
	etcdHostRole map[string]interface{},
	kubeClusterUUID string,
	etcdClusterUUID string,
) []interface{} {
	var mergedRoles []interface{}

	// Preserve non-Kubernetes roles and roles for different clusters
	for _, role := range existingRoles {
		roleMap, ok := role.(map[string]interface{})
		if !ok {
			continue
		}

		childType, _ := roleMap["childType"].(string)

		// Check if this is a KubeletRole for the managed cluster
		if childType == "KubeletRole" {
			roleCluster, _ := roleMap["kubeCluster"].(string)
			if roleCluster == kubeClusterUUID && kubeClusterUUID != "" {
				// Skip - will be replaced by new kubeletRole
				continue
			}
		}

		// Check if this is an EtcdHostRole for the managed cluster
		if childType == "EtcdHostRole" {
			roleCluster, _ := roleMap["etcdCluster"].(string)
			if roleCluster == etcdClusterUUID && etcdClusterUUID != "" {
				// Skip - will be replaced by new etcdHostRole
				continue
			}
		}

		// Preserve this role
		mergedRoles = append(mergedRoles, roleMap)
	}

	// Add new Kubernetes roles
	if kubeletRole != nil {
		mergedRoles = append(mergedRoles, kubeletRole)
	}

	if etcdHostRole != nil {
		mergedRoles = append(mergedRoles, etcdHostRole)
	}

	return mergedRoles
}

// ExtractKubernetesRolesFromDevice extracts Kubernetes role entities from a device's roles array.
// Returns the first matching KubeletRole and EtcdHostRole found.
//
// Parameters:
//   - roles: roles array from BCM device entity
//
// Returns:
//   - kubeletRoles: All KubeletRole entities found (may be multiple for different clusters)
//   - etcdHostRoles: All EtcdHostRole entities found (may be multiple for different clusters)
func ExtractKubernetesRolesFromDevice(roles []interface{}) ([]map[string]interface{}, []map[string]interface{}) {
	var kubeletRoles []map[string]interface{}
	var etcdHostRoles []map[string]interface{}

	for _, role := range roles {
		roleMap, ok := role.(map[string]interface{})
		if !ok {
			continue
		}

		childType, _ := roleMap["childType"].(string)

		switch childType {
		case "KubeletRole":
			kubeletRoles = append(kubeletRoles, roleMap)
		case "EtcdHostRole":
			etcdHostRoles = append(etcdHostRoles, roleMap)
		}
	}

	return kubeletRoles, etcdHostRoles
}

// =============================================================================
// Helper Functions
// =============================================================================

// extractStringList extracts a string slice from Terraform types.List.
// Returns defaultValue if list is null/unknown/empty.
func extractStringList(ctx context.Context, list types.List, diags *diag.Diagnostics, defaultValue []string) []string {
	if list.IsNull() || list.IsUnknown() {
		return defaultValue
	}

	var result []string
	d := list.ElementsAs(ctx, &result, false)
	diags.Append(d...)

	if d.HasError() || len(result) == 0 {
		return defaultValue
	}

	return result
}
