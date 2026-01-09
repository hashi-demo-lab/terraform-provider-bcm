# Data Model: CMKube API Alignment

**Date**: 2026-01-09
**Feature**: CMKube API Alignment (102-cmkube-api-alignment)

## Entity Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          BCM Kubernetes Architecture                         │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ┌─────────────────┐         ┌─────────────────┐                          │
│   │  EtcdCluster    │◄────────│  KubeCluster    │                          │
│   │  (cmetcd)       │         │  (cmkube)       │                          │
│   └────────┬────────┘         └────────┬────────┘                          │
│            │                           │                                    │
│            │ references                │ references                         │
│            ▼                           ▼                                    │
│   ┌─────────────────┐         ┌─────────────────┐                          │
│   │  EtcdHostRole   │         │  KubeletRole    │                          │
│   │  (in Device)    │         │  (in Device)    │                          │
│   └────────┬────────┘         └────────┬────────┘                          │
│            │                           │                                    │
│            └───────────┬───────────────┘                                   │
│                        │ embedded in                                        │
│                        ▼                                                    │
│               ┌─────────────────┐                                          │
│               │     Device      │                                          │
│               │   (cmdevice)    │                                          │
│               └─────────────────┘                                          │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Entity: KubeCluster

**BCM Service**: `cmkube` (lowercase)
**Terraform Resource**: `bcm_cmkube_cluster`

### Schema Definition

```go
// CMKubeClusterResourceModel - Aligned to BCM API
type CMKubeClusterResourceModel struct {
    // Identity
    ID   types.String `tfsdk:"id"`   // Computed, same as UUID
    UUID types.String `tfsdk:"uuid"` // Computed (client-generated for cmkube)
    Name types.String `tfsdk:"name"` // Required

    // Network Configuration (UUID references)
    InternalNetwork types.String `tfsdk:"internal_network"` // Required, Network UUID
    ServiceNetwork  types.String `tfsdk:"service_network"`  // Required, Network UUID
    PodNetwork      types.String `tfsdk:"pod_network"`      // Required, Network UUID

    // Etcd Configuration
    EtcdCluster types.String `tfsdk:"etcd_cluster"` // Required, EtcdCluster UUID

    // Kubernetes Configuration
    Version               types.String `tfsdk:"version"`                  // Optional, semver
    PodNetworkNodeMask    types.String `tfsdk:"pod_network_node_mask"`    // Optional, e.g., "/24"
    KubeDnsIp             types.String `tfsdk:"kube_dns_ip"`              // Optional, IP address
    KubernetesApiServer   types.String `tfsdk:"kubernetes_api_server"`    // Optional, URL
    KubernetesApiServerProxyPort types.Int64 `tfsdk:"kubernetes_api_server_proxy_port"` // Optional, default 6444

    // Certificate Configuration
    TrustedDomains types.List `tfsdk:"trusted_domains"` // Optional, []string SANs

    // Ingress Proxy Configuration
    IngressProxyEnable      types.Bool  `tfsdk:"ingress_proxy_enable"`       // Optional, default false
    IngressProxyListenPort  types.Int64 `tfsdk:"ingress_proxy_listen_port"`  // Optional
    IngressProxyBackendPort types.Int64 `tfsdk:"ingress_proxy_backend_port"` // Optional

    // Extensible Configuration
    Options types.String `tfsdk:"options"` // Optional, JSON string

    // Nested Blocks
    AppGroups []KubeAppGroupModel `tfsdk:"app_groups"` // Optional

    // Operations
    Force types.Bool `tfsdk:"force"` // Optional, default false

    // Computed Metadata
    CreationTime types.Int64 `tfsdk:"creation_time"` // Computed
    RevisionID   types.Int64 `tfsdk:"revision_id"`   // Computed
}
```

### Nested: KubeAppGroup

```go
type KubeAppGroupModel struct {
    Name         types.String      `tfsdk:"name"`         // Required
    Enabled      types.Bool        `tfsdk:"enabled"`      // Optional, default true
    Applications []KubeAppModel    `tfsdk:"applications"` // Optional
}

type KubeAppModel struct {
    Name     types.String `tfsdk:"name"`     // Required
    Enabled  types.Bool   `tfsdk:"enabled"`  // Optional, default true
    Manifest types.String `tfsdk:"manifest"` // Optional, YAML content
}
```

### Field Mappings (Terraform -> BCM API)

| Terraform Attribute | BCM API Field | Type | Notes |
|---------------------|---------------|------|-------|
| `id` | `uuid` | string | Computed, same as uuid |
| `uuid` | `uuid` | string | Client-generated for cmkube |
| `name` | `name` | string | Required |
| `internal_network` | `internalNetwork` | string | Network UUID |
| `service_network` | `serviceNetwork` | string | Network UUID |
| `pod_network` | `podNetwork` | string | Network UUID |
| `etcd_cluster` | `etcdCluster` | string | EtcdCluster UUID |
| `version` | `version` | string | Semver format |
| `pod_network_node_mask` | `podNetworkNodeMask` | string | CIDR mask |
| `kube_dns_ip` | `kubeDnsIp` | string | IP address |
| `kubernetes_api_server` | `kubernetesApiServer` | string | URL |
| `kubernetes_api_server_proxy_port` | `kubernetesApiServerProxyPort` | int | Default 6444 |
| `trusted_domains` | `trustedDomains` | []string | Certificate SANs |
| `ingress_proxy_enable` | `ingressProxyEnable` | bool | Default false |
| `ingress_proxy_listen_port` | `ingressProxyListenPort` | int | |
| `ingress_proxy_backend_port` | `ingressProxyBackendPort` | int | |
| `options` | `options` | object | JSON-encoded |
| `app_groups` | `appGroups` | []object | Nested |
| `creation_time` | `creationTime` | int | Unix timestamp |
| `revision_id` | `revisionID` | int | Version |

### Validation Rules

| Field | Validation |
|-------|------------|
| `name` | RFC 1123 DNS label, alphanumeric + hyphens |
| `internal_network` | Valid UUID (RFC 4122) |
| `service_network` | Valid UUID (RFC 4122) |
| `pod_network` | Valid UUID (RFC 4122) |
| `etcd_cluster` | Valid UUID (RFC 4122) |
| `version` | Semver format (e.g., "1.28.0") |
| `pod_network_node_mask` | CIDR notation (e.g., "/24") |
| `kube_dns_ip` | Valid IPv4 address |
| `kubernetes_api_server` | Valid URL |

---

## Entity: EtcdCluster

**BCM Service**: `cmetcd` (lowercase)
**Terraform Resource**: `bcm_cmetcd_cluster`

### Schema Definition

```go
type CMEtcdClusterResourceModel struct {
    // Identity
    ID   types.String `tfsdk:"id"`   // Computed, same as UUID
    UUID types.String `tfsdk:"uuid"` // Computed
    Name types.String `tfsdk:"name"` // Required

    // Etcd Configuration
    HeartbeatInterval types.Int64 `tfsdk:"heartbeat_interval"` // Optional, ms, default 100
    ElectionTimeout   types.Int64 `tfsdk:"election_timeout"`   // Optional, ms, default 1000

    // Extensible Configuration
    Options types.String `tfsdk:"options"` // Optional, JSON string

    // Operations
    Force types.Bool `tfsdk:"force"` // Optional, default false

    // Computed Metadata
    CreationTime types.Int64 `tfsdk:"creation_time"` // Computed
    RevisionID   types.Int64 `tfsdk:"revision_id"`   // Computed
}
```

### Field Mappings (Terraform -> BCM API)

| Terraform Attribute | BCM API Field | Type | Notes |
|---------------------|---------------|------|-------|
| `id` | `uuid` | string | Computed |
| `uuid` | `uuid` | string | Computed |
| `name` | `name` | string | Required |
| `heartbeat_interval` | `heartbeatInterval` | int | Default 100ms |
| `election_timeout` | `electionTimeout` | int | Default 1000ms |
| `options` | `options` | object | JSON-encoded |
| `creation_time` | `creationTime` | int | Unix timestamp |
| `revision_id` | `revisionID` | int | Version |

### Validation Rules

| Field | Validation |
|-------|------------|
| `name` | RFC 1123 DNS label |
| `heartbeat_interval` | Positive integer, typically 50-500 |
| `election_timeout` | Positive integer, typically 500-5000, >= 5x heartbeat |

---

## Entity: KubeletRole (Embedded in Device)

**BCM Service**: `cmdevice`
**Terraform Resource**: `bcm_cmdevice_device` (nested block)

### Schema Definition

```go
type KubeletRoleModel struct {
    // Identity (Computed)
    UUID types.String `tfsdk:"uuid"` // Computed, role UUID

    // Required Configuration
    KubeCluster types.String `tfsdk:"kube_cluster"` // Required, KubeCluster UUID

    // Node Type Configuration
    ControlPlane types.Bool `tfsdk:"control_plane"` // Optional, default true
    Worker       types.Bool `tfsdk:"worker"`        // Optional, default true

    // Runtime Configuration
    ContainerRuntimeService types.String `tfsdk:"container_runtime_service"` // Optional, default "docker.service"
    MaxPods                 types.Int64  `tfsdk:"max_pods"`                  // Optional, default 110

    // Advanced Configuration
    Options    types.String `tfsdk:"options"`     // Optional, JSON kubelet flags
    CustomYaml types.String `tfsdk:"custom_yaml"` // Optional, kubelet config.yaml content
}
```

### Field Mappings (Terraform -> BCM API)

| Terraform Attribute | BCM API Field | Type | Notes |
|---------------------|---------------|------|-------|
| `uuid` | `uuid` | string | Generated by provider |
| `kube_cluster` | `kubeCluster` | string | KubeCluster UUID |
| `control_plane` | `controlPlane` | bool | Default true |
| `worker` | `worker` | bool | Default true |
| `container_runtime_service` | `containerRuntimeService` | string | Default "docker.service" |
| `max_pods` | `maxPods` | int | Default 110 |
| `options` | `options` | object | JSON-encoded |
| `custom_yaml` | `customYaml` | string | Raw YAML |

### BCM API Entity Structure

```json
{
  "baseType": "Role",
  "childType": "KubeletRole",
  "uuid": "generated-uuid",
  "name": "kubelet",
  "kubeCluster": "kube-cluster-uuid",
  "controlPlane": true,
  "worker": true,
  "containerRuntimeService": "docker.service",
  "maxPods": 110,
  "options": {},
  "customYaml": "",
  "modified": true,
  "to_be_removed": false,
  "revision": ""
}
```

---

## Entity: EtcdHostRole (Embedded in Device)

**BCM Service**: `cmdevice`
**Terraform Resource**: `bcm_cmdevice_device` (nested block)

### Schema Definition

```go
type EtcdHostRoleModel struct {
    // Identity (Computed)
    UUID types.String `tfsdk:"uuid"` // Computed, role UUID

    // Required Configuration
    EtcdCluster types.String `tfsdk:"etcd_cluster"` // Required, EtcdCluster UUID

    // Member Configuration
    MemberName types.String `tfsdk:"member_name"` // Optional, default "$hostname"
    Spool      types.String `tfsdk:"spool"`       // Optional, default "/var/lib/etcd"

    // URL Configuration
    ListenClientUrls    types.List `tfsdk:"listen_client_urls"`    // Optional, default ["https://0.0.0.0:2379"]
    ListenPeerUrls      types.List `tfsdk:"listen_peer_urls"`      // Optional, default ["https://0.0.0.0:2380"]
    AdvertiseClientUrls types.List `tfsdk:"advertise_client_urls"` // Optional, default ["https://$ip:2379"]
    AdvertisePeerUrls   types.List `tfsdk:"advertise_peer_urls"`   // Optional, default ["https://$ip:2380"]

    // Snapshot Configuration
    SnapshotCount types.Int64 `tfsdk:"snapshot_count"` // Optional, default 100000
    MaxSnapshots  types.Int64 `tfsdk:"max_snapshots"`  // Optional, default 5
}
```

### Field Mappings (Terraform -> BCM API)

| Terraform Attribute | BCM API Field | Type | Notes |
|---------------------|---------------|------|-------|
| `uuid` | `uuid` | string | Generated by provider |
| `etcd_cluster` | `etcdCluster` | string | EtcdCluster UUID |
| `member_name` | `memberName` | string | Default "$hostname" |
| `spool` | `spool` | string | Default "/var/lib/etcd" |
| `listen_client_urls` | `listenClientUrls` | []string | Default ["https://0.0.0.0:2379"] |
| `listen_peer_urls` | `listenPeerUrls` | []string | Default ["https://0.0.0.0:2380"] |
| `advertise_client_urls` | `advertiseClientUrls` | []string | Default ["https://$ip:2379"] |
| `advertise_peer_urls` | `advertisePeerUrls` | []string | Default ["https://$ip:2380"] |
| `snapshot_count` | `snapshotCount` | int | Default 100000 |
| `max_snapshots` | `maxSnapshots` | int | Default 5 |

### BCM API Entity Structure

```json
{
  "baseType": "Role",
  "childType": "EtcdHostRole",
  "uuid": "generated-uuid",
  "name": "etcdhost",
  "etcdCluster": "etcd-cluster-uuid",
  "memberName": "$hostname",
  "spool": "/var/lib/etcd",
  "listenClientUrls": ["https://0.0.0.0:2379"],
  "listenPeerUrls": ["https://0.0.0.0:2380"],
  "advertiseClientUrls": ["https://$ip:2379"],
  "advertisePeerUrls": ["https://$ip:2380"],
  "snapshotCount": 100000,
  "maxSnapshots": 5,
  "modified": true,
  "to_be_removed": false,
  "revision": ""
}
```

---

## State Transitions

### KubeCluster Lifecycle

```
┌─────────────┐     addKubeCluster      ┌─────────────┐
│   Planned   │ ──────────────────────► │   Created   │
└─────────────┘                         └──────┬──────┘
                                               │
                                               │ updateKubeCluster
                                               ▼
                                        ┌─────────────┐
                                        │   Updated   │
                                        └──────┬──────┘
                                               │
                                               │ removeKubeCluster
                                               ▼
                                        ┌─────────────┐
                                        │   Deleted   │
                                        └─────────────┘
```

### Device Role Assignment

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│ Device without  │     │ Build new role  │     │ Device with     │
│ KubeletRole     │ ──► │ array including │ ──► │ KubeletRole     │
└─────────────────┘     │ KubeletRole     │     └─────────────────┘
                        └─────────────────┘
                               │
                               │ updateDevice (replaces entire roles array)
                               ▼
                        BCM updates Device entity
```

---

## Relationships

### Entity Relationship Diagram

```
┌──────────────────┐
│    Network       │ ◄───────────────────┐
│   (cmnet)        │                     │
└──────────────────┘                     │
        ▲                                │
        │ references (3x)                │
        │                                │
┌──────────────────┐            ┌────────┴─────────┐
│   EtcdCluster    │◄───────────│   KubeCluster    │
│   (cmetcd)       │ references │   (cmkube)       │
└────────┬─────────┘            └────────┬─────────┘
         │                               │
         │ referenced by                 │ referenced by
         ▼                               ▼
┌──────────────────┐            ┌──────────────────┐
│  EtcdHostRole    │            │   KubeletRole    │
│  (embedded)      │            │   (embedded)     │
└────────┬─────────┘            └────────┬─────────┘
         │                               │
         └───────────┬───────────────────┘
                     │ embedded in Device.roles[]
                     ▼
              ┌──────────────────┐
              │     Device       │
              │   (cmdevice)     │
              └──────────────────┘
```

### Dependency Order

**Creation Order**:
1. Networks (prerequisite)
2. EtcdCluster
3. KubeCluster (references Networks + EtcdCluster)
4. Devices with roles (references KubeCluster + EtcdCluster)

**Deletion Order** (reverse):
1. Devices (remove roles first)
2. KubeCluster
3. EtcdCluster
4. Networks (if not used elsewhere)
