# Research: BCM Device Roles Block

**Feature**: 039-device-roles-block
**Date**: 2025-11-26
**Status**: Complete

## Research Questions

### 1. How does BCM API handle role assignment in device entities?

**Finding**: Roles are embedded in the Device entity as an array field. BCM's addDevice and updateDevice operations accept the roles array as part of the device object.

**Evidence Sources**:
- `/workspace/sampleRest/DeviceEntity.md` - Role Assignments section
- `/workspace/sampleRest/CMDevice_Complete_Documentation.md` - roles field documentation
- `/workspace/internal/provider/data_source_cmdevice_roles.go` - Existing role extraction logic

**BCM Response Structure**:
```json
{
  "baseType": "Device",
  "childType": "PhysicalNode",
  "hostname": "control-01",
  "uuid": "...",
  "roles": [
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
  ]
}
```

**Conclusion**: Follow the same pattern as interfaces - embed roles array in device entity for CRUD operations.

---

### 2. What role object structure is required for create/update?

**Finding**: Minimal role object with name-based resolution. BCM accepts role names and resolves to UUIDs server-side.

**Create/Update Request Structure**:
```json
{
  "baseType": "Role",
  "name": "worker",
  "modified": true,
  "to_be_removed": false,
  "revision": ""
}
```

**Evidence**:
- Spec.md API Contract section (lines 156-195)
- Existing interface entity handling in resource_cmdevice_device.go (lines 1180-1200)
- BCM validation uses validateDevice which checks role names

**Fields Analysis**:
| Field | Required | Purpose |
|-------|----------|---------|
| baseType | Yes | Always "Role" |
| name | Yes | Role name for BCM resolution |
| modified | Yes | Indicates new/changed entity |
| to_be_removed | Yes | Set to false for add, true for remove |
| revision | Yes | Empty string for new roles |
| childType | No | BCM fills this from role definition |
| uuid | No | BCM assigns or resolves |
| addServices | No | BCM uses default from role definition |

**Conclusion**: Use minimal object with name-based assignment for simplicity.

---

### 3. How are existing roles handled on update?

**Finding**: Send complete desired state. BCM compares against current state and applies changes.

**Update Semantics**:
- Include all desired roles in the array
- Omitted roles are removed
- Empty array `[]` removes all roles
- BCM handles add/remove logic internally

**Evidence**:
- Interface block pattern in resource_cmdevice_device.go (buildDeviceAPIEntityWithExisting)
- Spec.md FR-004: "System MUST update role assignments during UPDATE operation"

**Conclusion**: Build roles array from plan state, let BCM handle delta calculation.

---

### 4. How should the provider detect drift?

**Finding**: Standard Read operation extracts roles from getDevice response. Terraform compares against state.

**Drift Detection Flow**:
1. Read() calls `cmdevice.getDevice(uuid)`
2. parseDeviceFromAPI() extracts role names from response
3. Sort role names for consistent comparison
4. Terraform detects differences between state and refreshed values
5. Plan shows changes if roles differ

**Evidence**:
- Existing drift detection for other fields (e.g., kernelParameters)
- Spec.md User Story 5 - Drift Detection requirements

**Conclusion**: No special handling needed - standard Read() refresh enables drift detection.

---

### 5. What known role types exist in BCM?

**Finding**: BCM has standard role types plus Kubernetes-specific roles when configured.

**Standard BCM Roles**:
| Name | childType | Description |
|------|-----------|-------------|
| headnode | HeadNodeRole | Cluster management node |
| storage | StorageRole | NFS storage services |
| backup | BackupRole | Backup services |
| monitoring | MonitoringRole | Cluster monitoring agent |
| provisioning | ProvisioningRole | Node provisioning services |
| boot | BootRole | PXE boot services |
| compute | ComputeRole | Compute workload execution |

**Kubernetes Roles** (when K8s configured):
| Name | childType | Description |
|------|-----------|-------------|
| control-plane | KubeControlPlaneRole | Kubernetes control plane |
| master | KubeMasterRole | Kubernetes master node |
| worker | KubeWorkerRole | Kubernetes worker node |
| etcd | KubeEtcdRole | Kubernetes etcd node |

**Evidence**:
- Spec.md Known Role Types table (lines 202-216)
- CMDevice_Quick_Reference.md Common Role Types section

**Conclusion**: Document common roles in schema description, but don't hard-code validation - let BCM validate.

---

### 6. How to handle import of devices with existing roles?

**Finding**: Import uses standard ImportStatePassthroughID. Read() populates roles from BCM response.

**Import Flow**:
1. User runs `terraform import bcm_cmdevice_device.name <uuid>`
2. ImportState sets ID from import argument
3. Read() calls getDevice(uuid) and parses full response
4. Roles extracted into state
5. Subsequent plan shows no changes if config matches

**Evidence**:
- Existing ImportState implementation (line 1137-1144)
- Read() already handles import case with isImport detection (lines 772-830)

**Conclusion**: No special import handling needed - existing pattern handles roles.

---

## Summary

| Research Item | Decision | Confidence |
|---------------|----------|------------|
| API Pattern | Embedded in device entity | High |
| Role Structure | Minimal with name-based resolution | High |
| Update Semantics | Send complete desired state | High |
| Drift Detection | Standard Read() refresh | High |
| Role Types | Document but don't validate | Medium |
| Import | Existing pattern sufficient | High |

## Implementation Recommendations

1. **Schema**: Use `types.List` with `StringType` for role names
2. **Build**: Deduplicate role names before sending to BCM
3. **Parse**: Sort role names for consistent state comparison
4. **Validation**: Rely on BCM's validateDevice for role name validation
5. **State**: Handle null vs empty list semantics (null = no roles block, empty = explicit no roles)

## References

- `/workspace/specs/039-device-roles-block/spec.md`
- `/workspace/internal/provider/resource_cmdevice_device.go`
- `/workspace/internal/provider/data_source_cmdevice_roles.go`
- `/workspace/sampleRest/DeviceEntity.md`
- `/workspace/sampleRest/CMDevice_Complete_Documentation.md`
- `/workspace/sampleRest/CMDevice_Quick_Reference.md`
