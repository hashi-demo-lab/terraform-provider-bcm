# Phase 0 Research: CMKube API Alignment

**Date**: 2026-01-09
**Feature**: CMKube API Alignment (102-cmkube-api-alignment)

## Executive Summary

Research validates that the current `bcm_cmkube_cluster` resource schema does not match the actual BCM API. Key findings:
1. BCM does NOT persist `master_nodes`, `worker_nodes`, `etcd_nodes` fields on KubeCluster entity
2. Node assignment is done via **KubeletRole** and **EtcdHostRole** embedded in Device entities
3. EtcdCluster is a **separate entity** with its own CRUD operations (CMEtcd service)
4. KubeCluster has network UUID references (`internalNetwork`, `serviceNetwork`, `podNetwork`)

## Research Tasks

### 1. BCM KubeCluster Entity Structure

**Decision**: Align resource to actual BCM API entity fields

**Rationale**: Current tests use `ImportStateVerifyIgnore: []string{"master_nodes", "worker_nodes", "etcd_nodes"}` which indicates these fields are NOT returned by `getKubeCluster`. The BCM API documentation and codebase analysis confirm the actual entity structure differs significantly from the current schema.

**Actual BCM KubeCluster Entity Fields** (verified via API documentation):

| BCM Field | Type | Description |
|-----------|------|-------------|
| `uuid` | string | BCM-assigned UUID (client-generated for cmkube) |
| `name` | string | Cluster name |
| `internalNetwork` | string | UUID reference to Network entity |
| `serviceNetwork` | string | UUID reference to Network entity |
| `podNetwork` | string | UUID reference to Network entity |
| `etcdCluster` | string | UUID reference to EtcdCluster entity |
| `podNetworkNodeMask` | string | Pod CIDR mask (e.g., "/24") |
| `kubeDnsIp` | string | Cluster DNS IP address |
| `kubernetesApiServer` | string | API server URL |
| `kubernetesApiServerProxyPort` | int | API server proxy port (default: 6444) |
| `version` | string | Kubernetes version (semver) |
| `trustedDomains` | []string | Certificate SANs |
| `appGroups` | []KubeAppGroup | Application groups with KubeApp entities |
| `labelSets` | []KubeLabelSet | Node label definitions |
| `users` | []KubeUser | Kubeconfig user definitions |
| `ingressProxyEnable` | bool | Enable ingress proxy |
| `ingressProxyListenPort` | int | Ingress proxy listen port |
| `ingressProxyBackendPort` | int | Ingress proxy backend port |
| `options` | object | Extensible JSON configuration |

**Fields to REMOVE from current schema** (not persisted by BCM):
- `master_nodes` - Write-only, not returned by getKubeCluster
- `worker_nodes` - Write-only, not returned by getKubeCluster
- `etcd_nodes` - Write-only, not returned by getKubeCluster
- `dns_servers` - Not a BCM field
- `cni_plugin` - Not a BCM field (CNI is part of app_groups)
- `storage_classes` - Not a BCM field (managed via app_groups)
- `load_balancer_mode` - Not a BCM field
- `addons` - Incorrect structure (should be app_groups)
- `ingress_controller` - Incorrect structure (should be app_groups)
- `overlay_network` - Renamed to proper network UUIDs
- `management_network` - Renamed to internalNetwork

**Alternatives Considered**:
- Keep deprecated fields with warnings: Rejected - too confusing, fields never worked
- Add both old and new fields: Rejected - maintenance burden, ambiguity

---

### 2. BCM EtcdCluster Entity Structure

**Decision**: Create new `bcm_cmetcd_cluster` resource

**Rationale**: EtcdCluster is a first-class entity in BCM with its own CRUD operations via the CMEtcd service. KubeCluster references EtcdCluster by UUID.

**BCM EtcdCluster Entity Fields**:

| BCM Field | Type | Description |
|-----------|------|-------------|
| `uuid` | string | BCM-assigned UUID |
| `name` | string | Cluster name |
| `heartbeatInterval` | int | Etcd heartbeat interval (ms, default: 100) |
| `electionTimeout` | int | Etcd election timeout (ms, default: 1000) |
| `options` | object | Extensible JSON configuration |
| `creationTime` | int | Unix timestamp |
| `revisionID` | int | Version for optimistic locking |

**CMEtcd API Methods**:
- `addEtcdCluster(entity)` - Create
- `getEtcdCluster(uuid)` - Read
- `updateEtcdCluster(entity)` - Update
- `removeEtcdCluster(uuid)` - Delete

**Alternatives Considered**:
- Embed etcd config in KubeCluster: Rejected - BCM has separate entity
- Skip EtcdCluster resource: Rejected - Required for complete cluster setup

---

### 3. Role-Based Node Assignment

**Decision**: Add `kubelet_role` and `etcd_host_role` nested blocks to `bcm_cmdevice_device`

**Rationale**: BCM assigns cluster membership via roles embedded in the Device entity. The existing `roles` attribute uses role names/UUIDs for simple roles. Kubernetes roles require nested configuration (cluster reference, control_plane flag, etc.).

**KubeletRole Structure** (embedded in Device.roles):

| Field | Type | Description |
|-------|------|-------------|
| `baseType` | string | "Role" |
| `childType` | string | "KubeletRole" |
| `uuid` | string | Role UUID |
| `name` | string | "kubelet" |
| `kubeCluster` | string | UUID reference to KubeCluster |
| `controlPlane` | bool | Is control plane node (default: true) |
| `worker` | bool | Is worker node (default: true) |
| `containerRuntimeService` | string | Runtime service (default: "docker.service") |
| `maxPods` | int | Max pods per node (default: 110) |
| `options` | object | Kubelet flags |
| `customYaml` | string | /var/lib/kubelet/config.yaml content |

**EtcdHostRole Structure** (embedded in Device.roles):

| Field | Type | Description |
|-------|------|-------------|
| `baseType` | string | "Role" |
| `childType` | string | "EtcdHostRole" |
| `uuid` | string | Role UUID |
| `name` | string | "etcdhost" |
| `etcdCluster` | string | UUID reference to EtcdCluster |
| `memberName` | string | Etcd member name (default: "$hostname") |
| `spool` | string | Data directory (default: "/var/lib/etcd") |
| `listenClientUrls` | []string | Client URLs (default: ["https://0.0.0.0:2379"]) |
| `listenPeerUrls` | []string | Peer URLs (default: ["https://0.0.0.0:2380"]) |
| `advertiseClientUrls` | []string | Advertise client URLs |
| `advertisePeerUrls` | []string | Advertise peer URLs |
| `snapshotCount` | int | Snapshot threshold (default: 100000) |
| `maxSnapshots` | int | Max snapshots retained (default: 5) |

**Implementation Pattern**: Follow existing `interfaces` block pattern in device resource
- Nested block with typed configuration
- UUID preservation across updates
- Role merging with existing roles

**Alternatives Considered**:
- Separate role resources: Rejected - BCM embeds roles in Device entity
- Simple role name assignment: Rejected - Kubernetes roles need configuration

---

### 4. Validation API Integration

**Decision**: Use existing ValidateEntity pattern with cmkube/cmetcd services

**Rationale**: The codebase already has pre-flight validation via `ValidateEntity`. Extend to new resources.

**Validation Methods**:

| Resource | Service | Method |
|----------|---------|--------|
| KubeCluster | cmkube (lowercase) | validateKubeCluster |
| EtcdCluster | cmetcd (lowercase) | validateEtcdCluster |
| Device | CMDevice | validateDevice |

**Note**: Service names for cmkube/cmetcd are **lowercase** (exception to CamelCase pattern).

---

### 5. Nested Block Patterns (app_groups, label_sets, users)

**Decision**: Implement app_groups as first nested block, others in future iteration

**Rationale**: app_groups is most commonly used (contains cluster addons). label_sets and users can be added later as P2/P3 features.

**KubeAppGroup Structure**:

```json
{
  "baseType": "KubeAppGroup",
  "name": "monitoring",
  "enabled": true,
  "applications": [
    {
      "baseType": "KubeApp",
      "name": "prometheus",
      "enabled": true,
      "manifest": "apiVersion: v1..."
    }
  ]
}
```

**Terraform Schema Pattern**:

```hcl
resource "bcm_cmkube_cluster" "example" {
  name = "production"

  app_groups {
    name    = "monitoring"
    enabled = true

    applications {
      name     = "prometheus"
      enabled  = true
      manifest = file("manifests/prometheus.yaml")
    }
  }
}
```

---

### 6. Breaking Change Strategy

**Decision**: Clean break with version bump, no deprecation path

**Rationale**:
- Current fields (`master_nodes`, etc.) NEVER worked - BCM silently ignored them
- Deprecation would require making non-functional code pretend to work
- Provider is pre-1.0, breaking changes are expected

**Migration Guide** (to be included in documentation):

1. Remove `master_nodes`, `worker_nodes`, `etcd_nodes` from cluster resource
2. Create `bcm_cmetcd_cluster` resource for etcd
3. Add `kubelet_role` blocks to device resources for cluster membership
4. Add `etcd_host_role` blocks to device resources for etcd membership
5. Update network references to use proper UUID fields

---

### 7. Test Infrastructure

**Decision**: Follow existing test patterns with new test helpers

**Test Helpers Needed**:
- `getTestEtcdClusterUUID(t)` - Get/create test etcd cluster
- `testAccCheckCMEtcdClusterDestroy(s)` - Verify etcd cluster deletion
- Role verification helpers for device tests

**Test Categories**:
1. **Unit Tests**: Schema validation, entity building, API response parsing
2. **Mock Tests**: BCM API interaction simulation
3. **Acceptance Tests**: Full CRUD with real BCM API

---

## Open Questions (Resolved)

| Question | Resolution |
|----------|------------|
| Does BCM return app_groups in getKubeCluster? | Yes, per API analysis |
| Is etcdCluster UUID required for KubeCluster? | Yes, cluster won't function without etcd |
| Can a device have multiple KubeletRoles? | No, one per device per cluster |
| How does BCM handle role UUID generation? | Provider generates UUID for new roles |

## Dependencies

1. **Existing Code**: `resource_cmkube_cluster.go`, `resource_cmdevice_device.go` - templates for patterns
2. **BCM API**: CMKube, CMEtcd, CMDevice services must be available
3. **Testing**: BCM test environment with available nodes and networks

## Next Steps

1. Generate data-model.md with complete entity definitions
2. Generate contracts/ with JSON schemas for API validation
3. Generate quickstart.md with migration examples
4. Proceed to Phase 2 task generation (/speckit.tasks)
