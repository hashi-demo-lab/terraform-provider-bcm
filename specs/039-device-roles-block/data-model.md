# Data Model: BCM Device Roles Block

**Feature**: 039-device-roles-block
**Date**: 2025-11-26
**Status**: Complete

## Entity Overview

This feature adds role assignment capability to the existing `bcm_cmdevice_device` resource. The data model extends the current device model with a roles list attribute.

## Terraform Schema Model

### CMDeviceDeviceResourceModel Changes

**Existing Model** (abbreviated):
```go
type CMDeviceDeviceResourceModel struct {
    ID                   types.String            `tfsdk:"id"`
    UUID                 types.String            `tfsdk:"uuid"`
    Hostname             types.String            `tfsdk:"hostname"`
    MAC                  types.String            `tfsdk:"mac"`
    Category             types.String            `tfsdk:"category"`
    ManagementNetwork    types.String            `tfsdk:"management_network"`
    Partition            types.String            `tfsdk:"partition"`
    Notes                types.String            `tfsdk:"notes"`
    KernelParameters     types.String            `tfsdk:"kernel_parameters"`
    // ... other fields ...
    Interfaces           []DeviceInterfaceModel  `tfsdk:"interfaces"`
}
```

**New Field**:
```go
type CMDeviceDeviceResourceModel struct {
    // ... existing fields ...

    // Roles - list of role names assigned to this device
    // Examples: ["control-plane", "master", "etcd"], ["worker"], ["headnode", "storage"]
    Roles types.List `tfsdk:"roles"` // Element type: types.StringType
}
```

### Schema Attribute Definition

```go
"roles": schema.ListAttribute{
    ElementType:         types.StringType,
    Optional:            true,
    MarkdownDescription: "List of role names to assign to the device (e.g., 'control-plane', 'worker', 'etcd'). " +
        "Use data.bcm_cmdevice_roles to discover available role names. " +
        "Roles are resolved by name - BCM maps names to role UUIDs server-side.",
},
```

## BCM API Entity Model

### Role Object (within Device entity)

**For Create/Update Operations**:
```json
{
  "baseType": "Role",
  "name": "<role_name>",
  "modified": true,
  "to_be_removed": false,
  "revision": ""
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| baseType | string | Yes | Always "Role" |
| name | string | Yes | Role name (e.g., "worker", "headnode") |
| modified | boolean | Yes | Set to true for new/changed roles |
| to_be_removed | boolean | Yes | Set to false for add, true for remove |
| revision | string | Yes | Empty string for new roles |

**From Read Operations** (BCM response):
```json
{
  "baseType": "Role",
  "childType": "HeadNodeRole",
  "name": "headnode",
  "uuid": "a458760f-3898-4870-bd7d-8127a5ea44b8",
  "addServices": true,
  "modified": false,
  "to_be_removed": false,
  "revision": ""
}
```

| Field | Type | Description |
|-------|------|-------------|
| baseType | string | Always "Role" |
| childType | string | Role type (HeadNodeRole, ComputeRole, etc.) |
| name | string | Role name |
| uuid | string | BCM-assigned role UUID |
| addServices | boolean | Whether role auto-adds services |
| modified | boolean | Change flag |
| to_be_removed | boolean | Deletion flag |
| revision | string | Version tracking |

### Device Entity with Roles

**Create/Update Request**:
```json
{
  "baseType": "Device",
  "childType": "PhysicalNode",
  "hostname": "control-01",
  "mac": "00:11:22:33:44:55",
  "category": "<category_uuid>",
  "uuid": "<device_uuid>",
  "partition": "<partition_uuid>",
  "interfaces": [...],
  "roles": [
    {
      "baseType": "Role",
      "name": "control-plane",
      "modified": true,
      "to_be_removed": false,
      "revision": ""
    },
    {
      "baseType": "Role",
      "name": "master",
      "modified": true,
      "to_be_removed": false,
      "revision": ""
    },
    {
      "baseType": "Role",
      "name": "etcd",
      "modified": true,
      "to_be_removed": false,
      "revision": ""
    }
  ],
  "modified": true,
  "to_be_removed": false,
  "revision": ""
}
```

## State Transitions

### Role Assignment States

```
[No roles field] ---> [Empty list] ---> [Populated list]
      |                    |                   |
      v                    v                   v
   BCM: no roles       BCM: remove all     BCM: assign roles
   State: null         State: []           State: ["role1", "role2"]
```

### Terraform Attribute Semantics

| Configuration | State Value | BCM Behavior |
|--------------|-------------|--------------|
| (omitted) | `null` | Device created without roles |
| `roles = []` | `[]` | Explicitly remove all roles |
| `roles = ["worker"]` | `["worker"]` | Assign worker role |
| `roles = ["a", "b"]` | `["a", "b"]` (sorted) | Assign both roles |

## Validation Rules

### Provider-Side Validation

1. **Non-empty strings**: Each role name must be non-empty
2. **Deduplication**: Duplicate role names are removed before API call
3. **Sorting**: Role names are sorted alphabetically in state for consistent comparison

### BCM Server-Side Validation

1. **Role existence**: BCM validates role names exist
2. **Role compatibility**: BCM checks role can be assigned to device type
3. **Conflict detection**: BCM detects incompatible role combinations

## Relationships

```
Device (bcm_cmdevice_device)
    |
    +-- Category (required reference)
    |
    +-- Partition (optional reference)
    |
    +-- Management Network (required reference)
    |
    +-- Interfaces[] (nested block)
    |
    +-- Roles[] (NEW - list of role names)
            |
            +-- resolved to Role entities in BCM
```

## Known Role Types

| Role Name | BCM childType | Use Case |
|-----------|---------------|----------|
| headnode | HeadNodeRole | Cluster management |
| storage | StorageRole | NFS storage |
| backup | BackupRole | Backup services |
| monitoring | MonitoringRole | Monitoring agent |
| provisioning | ProvisioningRole | Node provisioning |
| boot | BootRole | PXE boot services |
| compute | ComputeRole | Compute workloads |
| control-plane | KubeControlPlaneRole | K8s control plane |
| master | KubeMasterRole | K8s master |
| worker | KubeWorkerRole | K8s worker |
| etcd | KubeEtcdRole | K8s etcd |

Note: Role availability depends on BCM cluster configuration. Use `data.bcm_cmdevice_roles` to discover available roles.

## Example Terraform Configurations

### Kubernetes Control Plane Node
```hcl
resource "bcm_cmdevice_device" "control" {
  hostname           = "control-01"
  mac                = "00:11:22:33:44:55"
  category           = bcm_cmdevice_category.kube_control.uuid
  management_network = data.bcm_cmnet_networks.mgmt.networks[0].uuid

  roles = ["control-plane", "master", "etcd"]
}
```

### Kubernetes Worker Node
```hcl
resource "bcm_cmdevice_device" "worker" {
  hostname           = "worker-01"
  mac                = "00:11:22:33:44:66"
  category           = bcm_cmdevice_category.kube_worker.uuid
  management_network = data.bcm_cmnet_networks.mgmt.networks[0].uuid

  roles = ["worker"]
}
```

### Multi-Role Head Node
```hcl
resource "bcm_cmdevice_device" "head" {
  hostname           = "head-01"
  mac                = "00:11:22:33:44:77"
  category           = bcm_cmdevice_category.head.uuid
  management_network = data.bcm_cmnet_networks.mgmt.networks[0].uuid

  roles = ["headnode", "storage", "monitoring"]
}
```

### Device Without Roles (Legacy)
```hcl
resource "bcm_cmdevice_device" "basic" {
  hostname           = "node-01"
  mac                = "00:11:22:33:44:88"
  category           = bcm_cmdevice_category.compute.uuid
  management_network = data.bcm_cmnet_networks.mgmt.networks[0].uuid

  # No roles field - device created without role assignments
}
```
