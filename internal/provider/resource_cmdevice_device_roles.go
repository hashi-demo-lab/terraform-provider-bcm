// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Package provider contains role builders for Kubernetes roles (KubeletRole, EtcdHostRole)
// embedded in Device.roles[] arrays.
package provider

import (
	"context"
)

// =============================================================================
// Role Entity Builders (T050, T051)
// =============================================================================

// buildKubeletRoleEntity builds a KubeletRole entity for embedding in Device.roles[].
// The entity structure must match BCM's Role entity with childType="KubeletRole".
//
// BCM API Field Mappings:
//
//	Terraform Schema             BCM API Field
//	-----------------            ---------------
//	uuid                      → uuid (computed)
//	kube_cluster              → kubeCluster (UUID reference)
//	control_plane             → controlPlane
//	worker                    → worker
//	container_runtime_service → containerRuntimeService
//	max_pods                  → maxPods
//	options                   → options (JSON-encoded string)
//	custom_yaml               → customYaml
func buildKubeletRoleEntity(_ context.Context, model KubeletRoleModel) (map[string]any, error) {
	entity := map[string]any{
		"baseType":    "Role",
		"childType":   "KubeletRole",
		"name":        "kubelet",
		"modified":    true,
		"to_be_removed": false,
		"revision":    "",
	}

	// UUID - use existing or generate new
	if !model.UUID.IsNull() && !model.UUID.IsUnknown() && model.UUID.ValueString() != "" {
		entity["uuid"] = model.UUID.ValueString()
	} else {
		entity["uuid"] = generateUUID()
	}

	// Required: KubeCluster reference
	entity["kubeCluster"] = model.KubeCluster.ValueString()

	// ControlPlane - default true
	if !model.ControlPlane.IsNull() && !model.ControlPlane.IsUnknown() {
		entity["controlPlane"] = model.ControlPlane.ValueBool()
	} else {
		entity["controlPlane"] = true
	}

	// Worker - default true
	if !model.Worker.IsNull() && !model.Worker.IsUnknown() {
		entity["worker"] = model.Worker.ValueBool()
	} else {
		entity["worker"] = true
	}

	// ContainerRuntimeService - default "docker.service"
	if !model.ContainerRuntimeService.IsNull() && !model.ContainerRuntimeService.IsUnknown() {
		entity["containerRuntimeService"] = model.ContainerRuntimeService.ValueString()
	} else {
		entity["containerRuntimeService"] = "docker.service"
	}

	// MaxPods - default 110
	if !model.MaxPods.IsNull() && !model.MaxPods.IsUnknown() {
		entity["maxPods"] = model.MaxPods.ValueInt64()
	} else {
		entity["maxPods"] = int64(110)
	}

	// Options - optional JSON string
	if !model.Options.IsNull() && !model.Options.IsUnknown() {
		entity["options"] = model.Options.ValueString()
	}

	// CustomYAML - optional
	if !model.CustomYAML.IsNull() && !model.CustomYAML.IsUnknown() {
		entity["customYaml"] = model.CustomYAML.ValueString()
	}

	return entity, nil
}

// buildEtcdHostRoleEntity builds an EtcdHostRole entity for embedding in Device.roles[].
// The entity structure must match BCM's Role entity with childType="EtcdHostRole".
//
// BCM API Field Mappings:
//
//	Terraform Schema           BCM API Field
//	-----------------          ---------------
//	uuid                    → uuid (computed)
//	etcd_cluster            → etcdCluster (UUID reference)
//	member_name             → memberName
//	spool                   → spool
//	listen_client_urls      → listenClientUrls (list)
//	listen_peer_urls        → listenPeerUrls (list)
//	advertise_client_urls   → advertiseClientUrls (list)
//	advertise_peer_urls     → advertisePeerUrls (list)
//	snapshot_count          → snapshotCount
//	max_snapshots           → maxSnapshots
func buildEtcdHostRoleEntity(ctx context.Context, model EtcdHostRoleModel) (map[string]any, error) {
	entity := map[string]any{
		"baseType":    "Role",
		"childType":   "EtcdHostRole",
		"name":        "etcdhost",
		"modified":    true,
		"to_be_removed": false,
		"revision":    "",
	}

	// UUID - use existing or generate new
	if !model.UUID.IsNull() && !model.UUID.IsUnknown() && model.UUID.ValueString() != "" {
		entity["uuid"] = model.UUID.ValueString()
	} else {
		entity["uuid"] = generateUUID()
	}

	// Required: EtcdCluster reference
	entity["etcdCluster"] = model.EtcdCluster.ValueString()

	// MemberName - default "$hostname"
	if !model.MemberName.IsNull() && !model.MemberName.IsUnknown() {
		entity["memberName"] = model.MemberName.ValueString()
	} else {
		entity["memberName"] = "$hostname"
	}

	// Spool - default "/var/lib/etcd"
	if !model.Spool.IsNull() && !model.Spool.IsUnknown() {
		entity["spool"] = model.Spool.ValueString()
	} else {
		entity["spool"] = "/var/lib/etcd"
	}

	// ListenClientURLs - optional list
	if !model.ListenClientURLs.IsNull() && !model.ListenClientURLs.IsUnknown() {
		var urls []string
		model.ListenClientURLs.ElementsAs(ctx, &urls, false)
		entity["listenClientUrls"] = urls
	}

	// ListenPeerURLs - optional list
	if !model.ListenPeerURLs.IsNull() && !model.ListenPeerURLs.IsUnknown() {
		var urls []string
		model.ListenPeerURLs.ElementsAs(ctx, &urls, false)
		entity["listenPeerUrls"] = urls
	}

	// AdvertiseClientURLs - optional list
	if !model.AdvertiseClientURLs.IsNull() && !model.AdvertiseClientURLs.IsUnknown() {
		var urls []string
		model.AdvertiseClientURLs.ElementsAs(ctx, &urls, false)
		entity["advertiseClientUrls"] = urls
	}

	// AdvertisePeerURLs - optional list
	if !model.AdvertisePeerURLs.IsNull() && !model.AdvertisePeerURLs.IsUnknown() {
		var urls []string
		model.AdvertisePeerURLs.ElementsAs(ctx, &urls, false)
		entity["advertisePeerUrls"] = urls
	}

	// SnapshotCount - default 100000
	if !model.SnapshotCount.IsNull() && !model.SnapshotCount.IsUnknown() {
		entity["snapshotCount"] = model.SnapshotCount.ValueInt64()
	} else {
		entity["snapshotCount"] = int64(100000)
	}

	// MaxSnapshots - default 5
	if !model.MaxSnapshots.IsNull() && !model.MaxSnapshots.IsUnknown() {
		entity["maxSnapshots"] = model.MaxSnapshots.ValueInt64()
	} else {
		entity["maxSnapshots"] = int64(5)
	}

	return entity, nil
}

// =============================================================================
// Role Merging Logic (T052)
// =============================================================================

// mergeDeviceRoles merges new Kubernetes roles with existing device roles.
// It replaces KubeletRole/EtcdHostRole for the same cluster references while
// preserving all other roles (non-Kubernetes roles and roles for different clusters).
//
// Parameters:
//   - existingRoles: Current roles from BCM device entity
//   - newKubeletRoles: KubeletRole entities from Terraform config (nil = don't change, empty = remove all)
//   - newEtcdHostRoles: EtcdHostRole entities from Terraform config (nil = don't change, empty = remove all)
//
// Returns merged roles array for updateDevice API call.
func mergeDeviceRoles(
	existingRoles []map[string]any,
	newKubeletRoles []map[string]any,
	newEtcdHostRoles []map[string]any,
) []map[string]any {
	result := make([]map[string]any, 0)

	// Build a set of cluster UUIDs that will be replaced by new roles
	kubeletClusters := make(map[string]bool)
	etcdClusters := make(map[string]bool)

	// Track which clusters are being managed by Terraform
	if newKubeletRoles != nil {
		for _, role := range newKubeletRoles {
			if cluster, ok := role["kubeCluster"].(string); ok {
				kubeletClusters[cluster] = true
			}
		}
	}

	if newEtcdHostRoles != nil {
		for _, role := range newEtcdHostRoles {
			if cluster, ok := role["etcdCluster"].(string); ok {
				etcdClusters[cluster] = true
			}
		}
	}

	// First pass: preserve existing roles that are NOT being replaced
	for _, role := range existingRoles {
		childType, _ := role["childType"].(string)

		switch childType {
		case "KubeletRole":
			// If newKubeletRoles is nil, preserve all existing KubeletRoles
			if newKubeletRoles == nil {
				result = append(result, role)
				continue
			}
			// If newKubeletRoles is empty slice, remove all KubeletRoles
			if len(newKubeletRoles) == 0 {
				continue
			}
			// Otherwise, check if this cluster is being replaced
			cluster, _ := role["kubeCluster"].(string)
			if !kubeletClusters[cluster] {
				// This cluster is not in Terraform config, preserve the role
				result = append(result, role)
			}
			// If cluster IS in kubeletClusters, it will be replaced by new role

		case "EtcdHostRole":
			// If newEtcdHostRoles is nil, preserve all existing EtcdHostRoles
			if newEtcdHostRoles == nil {
				result = append(result, role)
				continue
			}
			// If newEtcdHostRoles is empty slice, remove all EtcdHostRoles
			if len(newEtcdHostRoles) == 0 {
				continue
			}
			// Otherwise, check if this cluster is being replaced
			cluster, _ := role["etcdCluster"].(string)
			if !etcdClusters[cluster] {
				// This cluster is not in Terraform config, preserve the role
				result = append(result, role)
			}
			// If cluster IS in etcdClusters, it will be replaced by new role

		default:
			// Non-Kubernetes roles are always preserved
			result = append(result, role)
		}
	}

	// Second pass: add new Kubernetes roles
	if newKubeletRoles != nil {
		result = append(result, newKubeletRoles...)
	}

	if newEtcdHostRoles != nil {
		result = append(result, newEtcdHostRoles...)
	}

	return result
}

// =============================================================================
// Role Parsing from BCM API Response
// =============================================================================

// parseKubeletRoleFromAPI parses a KubeletRole from BCM API response into a Terraform model.
// Uses shared helpers from utils.go for type-safe extraction.
func parseKubeletRoleFromAPI(roleData map[string]interface{}) KubeletRoleModel {
	model := KubeletRoleModel{}

	model.UUID = getStringValue(roleData, "uuid")
	model.KubeCluster = getStringValue(roleData, "kubeCluster")
	model.ControlPlane = getBoolValue(roleData, "controlPlane")
	model.Worker = getBoolValue(roleData, "worker")
	model.ContainerRuntimeService = getStringValue(roleData, "containerRuntimeService")
	model.MaxPods = getInt64Value(roleData, "maxPods")
	model.Options = getStringValue(roleData, "options")
	model.CustomYAML = getStringValue(roleData, "customYaml")

	return model
}

// parseEtcdHostRoleFromAPI parses an EtcdHostRole from BCM API response into a Terraform model.
// Uses shared helpers from utils.go for type-safe extraction.
func parseEtcdHostRoleFromAPI(roleData map[string]interface{}) EtcdHostRoleModel {
	model := EtcdHostRoleModel{}

	model.UUID = getStringValue(roleData, "uuid")
	model.EtcdCluster = getStringValue(roleData, "etcdCluster")
	model.MemberName = getStringValue(roleData, "memberName")
	model.Spool = getStringValue(roleData, "spool")
	model.SnapshotCount = getInt64Value(roleData, "snapshotCount")
	model.MaxSnapshots = getInt64Value(roleData, "maxSnapshots")

	// Parse URL lists using shared helper
	model.ListenClientURLs = GetStringListValue(roleData, "listenClientUrls")
	model.ListenPeerURLs = GetStringListValue(roleData, "listenPeerUrls")
	model.AdvertiseClientURLs = GetStringListValue(roleData, "advertiseClientUrls")
	model.AdvertisePeerURLs = GetStringListValue(roleData, "advertisePeerUrls")

	return model
}

// =============================================================================
// Role State Merging Helpers
// =============================================================================

// mergeKubeletRolesWithDefaults merges BCM API response values with plan values,
// preserving user-specified values and using computed defaults where appropriate.
// This ensures that optional computed fields like control_plane get proper defaults in state.
func mergeKubeletRolesWithDefaults(apiRoles []KubeletRoleModel, planRoles []KubeletRoleModel) []KubeletRoleModel {
	if len(planRoles) == 0 {
		return nil
	}

	// Build a map of API roles by cluster UUID for lookup
	apiRolesByCluster := make(map[string]KubeletRoleModel)
	for _, r := range apiRoles {
		if !r.KubeCluster.IsNull() && !r.KubeCluster.IsUnknown() {
			apiRolesByCluster[r.KubeCluster.ValueString()] = r
		}
	}

	result := make([]KubeletRoleModel, 0, len(planRoles))
	for _, planRole := range planRoles {
		clusterUUID := planRole.KubeCluster.ValueString()
		merged := planRole

		// If we have API data for this cluster, use computed values from BCM
		if apiRole, ok := apiRolesByCluster[clusterUUID]; ok {
			// UUID is always from BCM
			merged.UUID = apiRole.UUID

			// Use BCM values for optional+computed fields if plan didn't specify them
			if planRole.ControlPlane.IsNull() || planRole.ControlPlane.IsUnknown() {
				merged.ControlPlane = apiRole.ControlPlane
			}
			if planRole.Worker.IsNull() || planRole.Worker.IsUnknown() {
				merged.Worker = apiRole.Worker
			}
			if planRole.ContainerRuntimeService.IsNull() || planRole.ContainerRuntimeService.IsUnknown() {
				merged.ContainerRuntimeService = apiRole.ContainerRuntimeService
			}
			if planRole.MaxPods.IsNull() || planRole.MaxPods.IsUnknown() {
				merged.MaxPods = apiRole.MaxPods
			}
		}

		result = append(result, merged)
	}

	return result
}

// mergeEtcdHostRolesWithDefaults merges BCM API response values with plan values,
// preserving user-specified values and using computed defaults where appropriate.
func mergeEtcdHostRolesWithDefaults(apiRoles []EtcdHostRoleModel, planRoles []EtcdHostRoleModel) []EtcdHostRoleModel {
	if len(planRoles) == 0 {
		return nil
	}

	// Build a map of API roles by cluster UUID for lookup
	apiRolesByCluster := make(map[string]EtcdHostRoleModel)
	for _, r := range apiRoles {
		if !r.EtcdCluster.IsNull() && !r.EtcdCluster.IsUnknown() {
			apiRolesByCluster[r.EtcdCluster.ValueString()] = r
		}
	}

	result := make([]EtcdHostRoleModel, 0, len(planRoles))
	for _, planRole := range planRoles {
		clusterUUID := planRole.EtcdCluster.ValueString()
		merged := planRole

		// If we have API data for this cluster, use computed values from BCM
		if apiRole, ok := apiRolesByCluster[clusterUUID]; ok {
			// UUID is always from BCM
			merged.UUID = apiRole.UUID

			// Use BCM values for optional+computed fields if plan didn't specify them
			if planRole.MemberName.IsNull() || planRole.MemberName.IsUnknown() {
				merged.MemberName = apiRole.MemberName
			}
			if planRole.Spool.IsNull() || planRole.Spool.IsUnknown() {
				merged.Spool = apiRole.Spool
			}
			if planRole.SnapshotCount.IsNull() || planRole.SnapshotCount.IsUnknown() {
				merged.SnapshotCount = apiRole.SnapshotCount
			}
			if planRole.MaxSnapshots.IsNull() || planRole.MaxSnapshots.IsUnknown() {
				merged.MaxSnapshots = apiRole.MaxSnapshots
			}

			// URL lists - use API values if plan didn't specify
			if planRole.ListenClientURLs.IsNull() || planRole.ListenClientURLs.IsUnknown() {
				merged.ListenClientURLs = apiRole.ListenClientURLs
			}
			if planRole.ListenPeerURLs.IsNull() || planRole.ListenPeerURLs.IsUnknown() {
				merged.ListenPeerURLs = apiRole.ListenPeerURLs
			}
			if planRole.AdvertiseClientURLs.IsNull() || planRole.AdvertiseClientURLs.IsUnknown() {
				merged.AdvertiseClientURLs = apiRole.AdvertiseClientURLs
			}
			if planRole.AdvertisePeerURLs.IsNull() || planRole.AdvertisePeerURLs.IsUnknown() {
				merged.AdvertisePeerURLs = apiRole.AdvertisePeerURLs
			}
		}

		result = append(result, merged)
	}

	return result
}
