# Quickstart Guide: fsmounts Field Implementation

**Feature**: 084-fsmounts-implementation
**Date**: 2025-11-26

## Overview

This guide provides step-by-step instructions for implementing the `fsmounts` field in `bcm_cmdevice_category` resource following TDD principles.

---

## Prerequisites

- Go 1.24.0+
- BCM cluster access (172.21.15.254)
- Environment variables set:
  ```bash
  export TF_ACC=1
  export BCM_ENDPOINT="https://172.21.15.254:8081"
  export BCM_USERNAME="root"
  export BCM_PASSWORD="Hashicorp123!"
  ```

---

## Implementation Steps

### Step 1: RED Phase - Write Failing Tests

Create acceptance tests in `internal/provider/resource_cmdevice_category_test.go`:

```bash
# Run specific test to verify it fails (fsmounts returns null)
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceCategory_FSMountsBasic"
```

**Expected**: Test fails because fsmounts is null in state.

### Step 2: GREEN Phase - Implement Serialization

**File**: `internal/provider/resource_cmdevice_category.go`

**Location**: After line 2023 (replace the TODO comment at line 2025-2026)

Add fsmounts serialization in `buildAPIEntity`:

```go
// Serialize fsmounts (snake_case -> camelCase for BCM API)
if !model.FSMounts.IsNull() && !model.FSMounts.IsUnknown() {
    var mounts []FSMountModel
    diags := model.FSMounts.ElementsAs(ctx, &mounts, false)
    if !diags.HasError() {
        mountsList := make([]map[string]interface{}, 0, len(mounts))
        for _, mount := range mounts {
            mountMap := map[string]interface{}{
                "baseType": "FSMount",
                "device":   mount.Device.ValueString(),
                "path":     mount.Mountpoint.ValueString(),
                "type":     mount.Filesystem.ValueString(),
            }
            if !mount.UUID.IsNull() && mount.UUID.ValueString() != "" {
                mountMap["uuid"] = mount.UUID.ValueString()
            }
            if !mount.MountOptions.IsNull() {
                mountMap["options"] = mount.MountOptions.ValueString()
            }
            if !mount.Fsck.IsNull() {
                mountMap["fsck"] = mount.Fsck.ValueString()
            }
            if !mount.Dump.IsNull() {
                mountMap["dump"] = mount.Dump.ValueBool()
            }
            if !mount.RDMA.IsNull() {
                mountMap["rdma"] = mount.RDMA.ValueBool()
            }
            mountsList = append(mountsList, mountMap)
        }
        entity["fsmounts"] = mountsList
    }
}
```

### Step 3: GREEN Phase - Implement Parsing

**File**: `internal/provider/resource_cmdevice_category.go`

**Location**: Replace lines 2184-2196 (the TODO and `types.ListNull` call)

Replace the fsmounts parsing section in `readCategory`:

```go
// Parse fsmounts from BCM API (camelCase -> snake_case)
fsMountObjectType := types.ObjectType{AttrTypes: map[string]attr.Type{
    "uuid":         types.StringType,
    "device":       types.StringType,
    "mountpoint":   types.StringType,
    "filesystem":   types.StringType,
    "mountoptions": types.StringType,
    "fsck":         types.StringType,
    "dump":         types.BoolType,
    "rdma":         types.BoolType,
}}
if mountsData, ok := categoryData["fsmounts"].([]interface{}); ok && len(mountsData) > 0 {
    mountValues := make([]attr.Value, 0, len(mountsData))
    for _, mountRaw := range mountsData {
        if mountMap, ok := mountRaw.(map[string]interface{}); ok {
            mountObj, objDiags := types.ObjectValue(fsMountObjectType.AttrTypes, map[string]attr.Value{
                "uuid":         getStringValue(mountMap, "uuid"),
                "device":       getStringValue(mountMap, "device"),
                "mountpoint":   getStringValue(mountMap, "path"),
                "filesystem":   getStringValue(mountMap, "type"),
                "mountoptions": getStringValue(mountMap, "options"),
                "fsck":         getStringValue(mountMap, "fsck"),
                "dump":         getBoolValue(mountMap, "dump"),
                "rdma":         getBoolValue(mountMap, "rdma"),
            })
            if !objDiags.HasError() {
                mountValues = append(mountValues, mountObj)
            }
        }
    }
    model.FSMounts, _ = types.ListValue(fsMountObjectType, mountValues)
} else {
    model.FSMounts = types.ListNull(fsMountObjectType)
}
```

### Step 4: GREEN Phase - Add Preservation/Merge Logic

**Location**: Create function, Read function, Update function

**Create function** (around line 828):
```go
planFSMounts := plan.FSMounts
```

**Create function** (around line 1052):
```go
plan.FSMounts = mergeFSMountsWithAPIResponse(ctx, planFSMounts, plan.FSMounts)
```

**Read function** (around line 1077):
```go
originalFSMounts := state.FSMounts
```

**Read function** (around line 1195):
```go
state.FSMounts = mergeFSMountsWithAPIResponse(ctx, originalFSMounts, state.FSMounts)
```

**Update function** (around line 1321):
```go
planFSMounts := plan.FSMounts
```

**Update function** (around line 1390):
```go
plan.FSMounts = mergeFSMountsWithAPIResponse(ctx, planFSMounts, plan.FSMounts)
```

### Step 5: GREEN Phase - Add Merge Function

**Location**: After `mergeRolesWithAPIResponse` function (around line 2560)

```go
// mergeFSMountsWithAPIResponse merges user-configured fsmount attributes with BCM API-computed values.
func mergeFSMountsWithAPIResponse(ctx context.Context, originalMounts types.List, apiMounts types.List) types.List {
    fsMountObjectType := types.ObjectType{AttrTypes: map[string]attr.Type{
        "uuid":         types.StringType,
        "device":       types.StringType,
        "mountpoint":   types.StringType,
        "filesystem":   types.StringType,
        "mountoptions": types.StringType,
        "fsck":         types.StringType,
        "dump":         types.BoolType,
        "rdma":         types.BoolType,
    }}

    if originalMounts.IsNull() {
        tflog.Debug(ctx, "Original fsmounts is null, preserving null")
        return types.ListNull(fsMountObjectType)
    }

    if originalMounts.IsUnknown() {
        tflog.Debug(ctx, "Original fsmounts is unknown, using API response")
        return apiMounts
    }

    if apiMounts.IsNull() || apiMounts.IsUnknown() {
        tflog.Debug(ctx, "API returned null/unknown fsmounts, preserving original")
        return originalMounts
    }

    var origMounts []FSMountModel
    var apiMountsList []FSMountModel

    if diags := originalMounts.ElementsAs(ctx, &origMounts, false); diags.HasError() {
        return originalMounts
    }
    if diags := apiMounts.ElementsAs(ctx, &apiMountsList, false); diags.HasError() {
        return originalMounts
    }

    // Build lookup map: device+mountpoint -> API mount
    apiMountsByKey := make(map[string]FSMountModel)
    for _, mount := range apiMountsList {
        key := fmt.Sprintf("%s:%s", mount.Device.ValueString(), mount.Mountpoint.ValueString())
        apiMountsByKey[key] = mount
    }

    // Merge: preserve user config + populate UUID from API
    mergedMounts := make([]FSMountModel, 0, len(origMounts))
    for _, origMount := range origMounts {
        key := fmt.Sprintf("%s:%s", origMount.Device.ValueString(), origMount.Mountpoint.ValueString())
        if apiMount, found := apiMountsByKey[key]; found {
            mergedMount := FSMountModel{
                UUID:         apiMount.UUID,
                Device:       origMount.Device,
                Mountpoint:   origMount.Mountpoint,
                Filesystem:   origMount.Filesystem,
                MountOptions: origMount.MountOptions,
                Fsck:         origMount.Fsck,
                Dump:         origMount.Dump,
                RDMA:         origMount.RDMA,
            }
            mergedMounts = append(mergedMounts, mergedMount)
        } else {
            // Not in API - generate UUID if needed
            mergedMount := origMount
            if origMount.UUID.IsNull() || origMount.UUID.IsUnknown() || origMount.UUID.ValueString() == "" {
                mergedMount.UUID = types.StringValue(generateUUID())
            }
            mergedMounts = append(mergedMounts, mergedMount)
        }
    }

    // Convert back to types.List
    mountValues := make([]attr.Value, 0, len(mergedMounts))
    for _, mount := range mergedMounts {
        mountObj, diags := types.ObjectValue(fsMountObjectType.AttrTypes, map[string]attr.Value{
            "uuid":         mount.UUID,
            "device":       mount.Device,
            "mountpoint":   mount.Mountpoint,
            "filesystem":   mount.Filesystem,
            "mountoptions": mount.MountOptions,
            "fsck":         mount.Fsck,
            "dump":         mount.Dump,
            "rdma":         mount.RDMA,
        })
        if !diags.HasError() {
            mountValues = append(mountValues, mountObj)
        }
    }

    result, _ := types.ListValue(fsMountObjectType, mountValues)
    return result
}
```

### Step 6: Verify Tests Pass

```bash
# Run all fsmounts tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "FSMount"

# Run all category tests to ensure no regressions
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "CMDeviceCategory"
```

### Step 7: REFACTOR Phase - Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Generate documentation
make generate
```

---

## Test Configurations

### Basic FSMount Test

```hcl
resource "bcm_cmdevice_category" "test" {
  name               = "tftest-fsmounts-basic"
  management_network = data.bcm_cmnet_networks.management.networks[0].id

  fsmounts = [
    {
      device     = "/dev/sdb1"
      mountpoint = "/data"
      filesystem = "xfs"
    }
  ]
}
```

### Multiple FSMounts Test

```hcl
resource "bcm_cmdevice_category" "test" {
  name               = "tftest-fsmounts-multi"
  management_network = data.bcm_cmnet_networks.management.networks[0].id

  fsmounts = [
    {
      device       = "/dev/sdb1"
      mountpoint   = "/data"
      filesystem   = "xfs"
      mountoptions = "defaults,noatime"
    },
    {
      device     = "nfs-server:/export"
      mountpoint = "/shared"
      filesystem = "nfs"
      rdma       = true
    }
  ]
}
```

---

## Troubleshooting

### fsmounts is null in state
- Verify serialization code is reached (add tflog.Debug)
- Check BCM API response contains fsmounts data
- Verify parsing code handles the response format

### UUID not populated
- Verify merge function is called after readCategory
- Check apiMountsByKey lookup is matching correctly
- Verify BCM returns UUID in fsmounts array

### Drift detection on unchanged config
- Verify merge preserves user config values
- Check field mapping (mountpoint vs path, filesystem vs type)
- Ensure optional fields maintain null state when not configured

### Import doesn't include fsmounts
- Verify parsing handles import case (no original state)
- Check readCategory is called during import
- Verify BCM returns fsmounts for existing categories

---

## Related Files

| File | Purpose |
|------|---------|
| `internal/provider/resource_cmdevice_category.go` | Main implementation |
| `internal/provider/resource_cmdevice_category_test.go` | Acceptance tests |
| `specs/084-fsmounts-implementation/spec.md` | Feature specification |
| `specs/084-fsmounts-implementation/plan.md` | Implementation plan |
