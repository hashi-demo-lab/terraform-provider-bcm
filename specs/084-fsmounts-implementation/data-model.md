# Data Model: fsmounts Field Implementation

**Date**: 2025-11-26
**Feature**: 084-fsmounts-implementation
**Issue**: #84

## Entity Overview

This document describes the data model for the `fsmounts` field within the `bcm_cmdevice_category` resource. The field represents filesystem mount configurations for device categories.

---

## Entities

### FSMountModel (Existing)

**Location**: `internal/provider/resource_cmdevice_category.go:158-168`

**Purpose**: Represents a filesystem mount configuration within a category.

```go
// FSMountModel describes a filesystem mount nested object.
type FSMountModel struct {
    UUID         types.String `tfsdk:"uuid"`         // Computed, BCM-assigned
    Device       types.String `tfsdk:"device"`       // Required
    Mountpoint   types.String `tfsdk:"mountpoint"`   // Required
    Filesystem   types.String `tfsdk:"filesystem"`   // Required
    MountOptions types.String `tfsdk:"mountoptions"` // Optional
    Fsck         types.String `tfsdk:"fsck"`         // Optional
    Dump         types.Bool   `tfsdk:"dump"`         // Optional
    RDMA         types.Bool   `tfsdk:"rdma"`         // Optional
}
```

### Field Specifications

| Field | Type | Required | Computed | Description | Validation |
|-------|------|----------|----------|-------------|------------|
| uuid | string | No | Yes | BCM-assigned unique identifier | UUID format |
| device | string | Yes | No | Device path or NFS export | Non-empty string |
| mountpoint | string | Yes | No | Mount point path | Absolute path starting with / |
| filesystem | string | Yes | No | Filesystem type | Valid FS type (xfs, ext4, nfs, etc.) |
| mountoptions | string | No | No | Mount options string | Valid mount options |
| fsck | string | No | No | Filesystem check mode | Valid fsck mode |
| dump | bool | No | No | Dump backup flag | Boolean |
| rdma | bool | No | No | Use RDMA for NFS | Boolean |

### Relationships

```
CMDeviceCategoryResourceModel
    |
    +-- FSMounts (types.List)
           |
           +-- FSMountModel[0]
           +-- FSMountModel[1]
           +-- ...
```

**Cardinality**: One Category has zero or more FSMounts (1:N)

---

## BCM API Entity Mapping

### Terraform to BCM Mapping

| Terraform (snake_case) | BCM API (camelCase) | Transform |
|------------------------|---------------------|-----------|
| uuid | uuid | Direct |
| device | device | Direct |
| mountpoint | path | Name change |
| filesystem | type | Name change |
| mountoptions | options | Name change |
| fsck | fsck | Direct |
| dump | dump | Direct |
| rdma | rdma | Direct |

### BCM API Structure

**Serialization** (Terraform to BCM):
```json
{
  "fsmounts": [
    {
      "baseType": "FSMount",
      "uuid": "existing-uuid-if-update",
      "device": "/dev/sdb1",
      "path": "/data",
      "type": "xfs",
      "options": "defaults,noatime",
      "fsck": "0",
      "dump": false,
      "rdma": false
    }
  ]
}
```

**Parsing** (BCM to Terraform):
```json
{
  "fsmounts": [
    {
      "baseType": "FSMount",
      "uuid": "550e8400-e29b-41d4-a716-446655440000",
      "device": "/dev/sdb1",
      "path": "/data",
      "type": "xfs",
      "options": "defaults,noatime"
    }
  ]
}
```

---

## State Transitions

### Create Flow

```
1. User Config (plan)
   - uuid: unknown
   - device: "/dev/sdb1"
   - mountpoint: "/data"
   - filesystem: "xfs"

2. BCM API Response
   - uuid: "bcm-assigned-uuid"
   - device: "/dev/sdb1"
   - path: "/data"
   - type: "xfs"

3. Terraform State
   - uuid: "bcm-assigned-uuid"
   - device: "/dev/sdb1"
   - mountpoint: "/data"
   - filesystem: "xfs"
```

### Update Flow

```
1. User Config (plan)
   - uuid: "existing-uuid"
   - device: "/dev/sdb1"
   - mountpoint: "/data"
   - filesystem: "xfs"
   - mountoptions: "defaults,noatime" (added)

2. BCM API Update
   - Send: device, path, type, options, uuid
   - Receive: updated entity

3. Terraform State
   - Merge: preserve user config + populate uuid from API
```

### Delete Flow

```
1. User removes fsmount from config
2. buildAPIEntity excludes removed mount
3. BCM API receives category without mount
4. Mount removed from category
```

---

## Validation Rules

### Device Field
- **Required**: Must not be empty
- **Format**: String representing device path or NFS export
- **Examples**:
  - Local device: `/dev/sdb1`, `/dev/vda2`
  - NFS export: `nfs-server:/export/data`
  - LVM: `/dev/mapper/vg0-data`

### Mountpoint Field
- **Required**: Must not be empty
- **Format**: Absolute filesystem path starting with `/`
- **Examples**: `/data`, `/mnt/storage`, `/var/lib/data`
- **Validation**: BCM validates against system paths

### Filesystem Field
- **Required**: Must not be empty
- **Format**: Valid filesystem type string
- **Examples**: `xfs`, `ext4`, `nfs`, `tmpfs`, `btrfs`

### Mountoptions Field
- **Optional**: May be empty or null
- **Format**: Comma-separated mount options
- **Examples**: `defaults`, `defaults,noatime`, `ro,noexec`

---

## Uniqueness Constraints

**Within Category**: The combination of `device` + `mountpoint` must be unique within a category's fsmounts list.

**Match Key**: For merge operations, mounts are matched by `device + mountpoint` combination:
```go
matchKey := fmt.Sprintf("%s:%s", mount.Device.ValueString(), mount.Mountpoint.ValueString())
```

This ensures:
1. Same device can be mounted at different points
2. Same mount point cannot have multiple devices
3. Exact match required for UUID population during merge

---

## Schema Definition Reference

**Location**: `internal/provider/resource_cmdevice_category.go:478-517`

```go
"fsmounts": schema.ListNestedAttribute{
    Description:         "Filesystem mounts for category nodes.",
    MarkdownDescription: "Filesystem mounts for category nodes. Each mount defines a device, mount point, and filesystem type.",
    Optional:            true,
    NestedObject: schema.NestedAttributeObject{
        Attributes: map[string]schema.Attribute{
            "uuid": schema.StringAttribute{
                Description:         "BCM-assigned unique identifier for this mount.",
                MarkdownDescription: "BCM-assigned unique identifier for this mount.",
                Computed:            true,
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.UseStateForUnknown(),
                },
            },
            "device": schema.StringAttribute{
                Description:         "Device path or NFS export to mount.",
                MarkdownDescription: "Device path or NFS export to mount (e.g., `/dev/sdb1`, `nfs-server:/export`).",
                Required:            true,
            },
            "mountpoint": schema.StringAttribute{
                Description:         "Mount point path.",
                MarkdownDescription: "Mount point path (e.g., `/data`, `/mnt/storage`).",
                Required:            true,
            },
            "filesystem": schema.StringAttribute{
                Description:         "Filesystem type.",
                MarkdownDescription: "Filesystem type (e.g., `xfs`, `ext4`, `nfs`).",
                Required:            true,
            },
            "mountoptions": schema.StringAttribute{
                Description:         "Mount options.",
                MarkdownDescription: "Mount options (e.g., `defaults,noatime`).",
                Optional:            true,
            },
            "fsck": schema.StringAttribute{
                Description:         "Filesystem check mode.",
                MarkdownDescription: "Filesystem check mode.",
                Optional:            true,
            },
            "dump": schema.BoolAttribute{
                Description:         "Enable dump backup.",
                MarkdownDescription: "Enable dump backup for this mount.",
                Optional:            true,
            },
            "rdma": schema.BoolAttribute{
                Description:         "Use RDMA for NFS mounts.",
                MarkdownDescription: "Use RDMA for NFS mounts.",
                Optional:            true,
            },
        },
    },
},
```
