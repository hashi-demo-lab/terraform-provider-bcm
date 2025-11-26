# Feature Specification: Implement fsmounts Field in bcm_cmdevice_category Resource

**Feature Branch**: `084-fsmounts-implementation`
**Created**: 2025-11-26
**Status**: Draft
**Issue**: [#84](https://github.com/hashi-demo-lab/terraform-provider-bcm/issues/84)
**Input**: Implement fsmounts field in bcm_cmdevice_category resource

## Problem Statement

The `fsmounts` field in `bcm_cmdevice_category` resource is not implemented - it is never serialized to the BCM API during Create/Update operations and always returns null during Read operations. Users cannot configure filesystem mounts for device categories even though the schema defines the field.

### Root Cause Analysis

In `resource_cmdevice_category.go`:

1. **buildAPIEntity (line 2026)**: Contains a TODO comment indicating fsmounts serialization is not implemented:
   ```go
   // TODO: Add remaining nested objects and arrays in Phase 6 (Comprehensive Schema)
   // - fsmounts (array of FSMount)
   ```

2. **readCategory (lines 2184-2196)**: fsmounts is always set to null, ignoring BCM API response:
   ```go
   // Filesystem lists (set to null for now, Phase 6 will parse these)
   // TODO Phase 6: Parse actual fsmounts from API
   fsMountObjectType := types.ObjectType{...}
   model.FSMounts = types.ListNull(fsMountObjectType)
   ```

3. **No preservation logic**: Unlike fsexports, roles, and static_routes, fsmounts is not preserved from plan/state during Create/Read/Update operations.

### Current Behavior

```hcl
resource "bcm_cmdevice_category" "test" {
  name               = "test-fsmounts"
  management_network = data.bcm_cmnet_networks.management.networks[0].id

  fsmounts = [
    {
      device     = "/dev/sdb1"
      mountpoint = "/data"
      filesystem = "xfs"
    }
  ]
}

# After apply:
# - fsmounts is NOT sent to BCM API (TODO in buildAPIEntity)
# - fsmounts is always null in state (explicit types.ListNull())
# - Configuration drift is immediately detected on next plan
```

### Expected Behavior

```hcl
# After apply:
# - fsmounts is sent to BCM API with proper structure
# - fsmounts is populated in state with BCM-assigned UUIDs
# - Subsequent plans show no drift when configuration unchanged
```

## User Scenarios & Testing

### User Story 1 - Configure Filesystem Mounts on Category (Priority: P1)

As a Terraform practitioner managing BCM infrastructure, I want to define filesystem mount configurations for device categories, so that nodes in the category automatically have the correct mount points configured during provisioning.

**Why this priority**: This is the core functionality - without it, the fsmounts field is unusable and users cannot configure mount points through Terraform.

**Independent Test**: Create a category with fsmounts configuration, verify the mount configuration is sent to BCM API and appears in Terraform state after apply.

**Acceptance Scenarios**:

1. **Given** a category configuration with one fsmount specifying device, mountpoint, and filesystem, **When** terraform apply completes successfully, **Then** the mount configuration is sent to BCM API with proper FSMount structure
2. **Given** a category with fsmounts in state, **When** terraform plan is run with no configuration changes, **Then** no drift is detected (empty plan)
3. **Given** a category configuration with multiple fsmounts, **When** terraform apply completes, **Then** all mount configurations are persisted and visible in state

---

### User Story 2 - Update Filesystem Mounts (Priority: P1)

As a Terraform practitioner, I want to add, modify, or remove filesystem mounts on an existing category, so that I can manage mount configurations through the standard Terraform workflow.

**Why this priority**: Equal priority to Story 1 - users need full CRUD operations on mounts, not just initial creation.

**Independent Test**: Update an existing category to add/modify/remove mounts and verify changes are applied.

**Acceptance Scenarios**:

1. **Given** an existing category with one fsmount, **When** a second fsmount is added to configuration and terraform apply runs, **Then** both mounts are present in BCM and state
2. **Given** an existing category with fsmounts, **When** a mount's mountoptions is modified and terraform apply runs, **Then** the updated options are sent to BCM API
3. **Given** an existing category with two fsmounts, **When** one mount is removed from configuration and terraform apply runs, **Then** only the remaining mount exists in BCM and state

---

### User Story 3 - Populate Computed UUID on Mounts (Priority: P2)

As a Terraform practitioner, I want BCM-assigned UUIDs for filesystem mounts to be populated in Terraform state, so that I can reference mount identifiers in other resources or outputs.

**Why this priority**: Secondary to basic functionality - UUID population is useful for reference but not blocking for core use case.

**Independent Test**: Create a category with fsmounts and verify each mount's `uuid` attribute is populated after apply.

**Acceptance Scenarios**:

1. **Given** a category configuration with fsmounts, **When** terraform apply completes, **Then** each fsmount's `uuid` attribute contains a BCM-assigned UUID string
2. **Given** a category with fsmounts that have populated UUIDs, **When** terraform refresh runs, **Then** the UUIDs remain populated with the same values

---

### User Story 4 - Import Category with Existing Mounts (Priority: P2)

As a Terraform practitioner importing existing BCM categories, I want the imported state to include filesystem mount configurations, so that I can manage existing infrastructure through Terraform.

**Why this priority**: Import is a secondary workflow but important for adopting existing infrastructure.

**Independent Test**: Import an existing category that has fsmounts configured in BCM and verify mounts appear in Terraform state.

**Acceptance Scenarios**:

1. **Given** an existing BCM category with fsmounts, **When** terraform import is executed, **Then** the imported state includes the fsmount configurations with UUIDs

---

### Edge Cases

- What happens when BCM returns an empty fsmounts array but user configured mounts? (Mounts may not persist in BCM - preserve from plan like fsexports)
- How does the system handle invalid filesystem types? (BCM validation should reject - provider passes error to user)
- What happens if mount points conflict with system paths? (BCM validation handles this)
- How are NFS mounts (with remote device paths) handled vs local device mounts? (Same structure, device field contains NFS path like "nfs-server:/export")

## Requirements

### Functional Requirements

- **FR-001**: Provider MUST serialize `fsmounts` field to BCM API during Create operations with proper FSMount structure
- **FR-002**: Provider MUST serialize `fsmounts` field to BCM API during Update operations
- **FR-003**: Provider MUST parse `fsmounts` from BCM API response during Read operations (following fsexports pattern at lines 2198-2227)
- **FR-004**: Provider MUST preserve fsmounts from plan/state when BCM API doesn't persist the data (following fsexports preservation pattern)
- **FR-005**: Provider MUST populate computed `uuid` field for each fsmount from BCM API response
- **FR-006**: Provider MUST support all FSMountModel fields: device (required), mountpoint (required), filesystem (required), mountoptions (optional), fsck (optional), dump (optional), rdma (optional)
- **FR-007**: Provider MUST NOT cause drift detection when fsmounts configuration is unchanged

### Key Entities

- **FSMountModel**: Represents a filesystem mount configuration within a category
  - `uuid` (Computed): BCM-assigned unique identifier
  - `device` (Required): Device path or NFS export (e.g., "/dev/sdb1", "nfs-server:/export")
  - `mountpoint` (Required): Mount point path (e.g., "/data", "/shared")
  - `filesystem` (Required): Filesystem type (e.g., "xfs", "ext4", "nfs")
  - `mountoptions` (Optional): Mount options string (e.g., "defaults,noatime")
  - `fsck` (Optional): Filesystem check mode
  - `dump` (Optional): Dump backup flag (boolean)
  - `rdma` (Optional): Use RDMA for NFS (boolean)

- **BCM API FSMount Structure** (based on API documentation):
  ```json
  {
    "baseType": "FSMount",
    "uuid": "...",
    "path": "/shared",          // corresponds to mountpoint
    "device": "nfs-server:/export",
    "type": "nfs",              // corresponds to filesystem
    "options": "defaults"       // corresponds to mountoptions
  }
  ```

## Success Criteria

### Measurable Outcomes

- **SC-001**: After terraform apply on a category with fsmounts, 100% of configured mounts appear in Terraform state
- **SC-002**: Terraform plan shows no changes immediately after terraform apply (idempotency verified)
- **SC-003**: Existing acceptance tests for bcm_cmdevice_category continue to pass (no regressions)
- **SC-004**: New acceptance tests for fsmounts CRUD operations pass
- **SC-005**: Users can successfully configure NFS mounts and local device mounts through the fsmounts field

## Technical Context

### Files Requiring Modification

**Primary file**: `internal/provider/resource_cmdevice_category.go`

1. **buildAPIEntity function** (around line 2026): Add fsmounts serialization following fsexports pattern (lines 1847-1871)
2. **readCategory function** (lines 2184-2196): Replace `types.ListNull()` with proper parsing following fsexports pattern (lines 2198-2227)
3. **Create function** (around line 828): Add `planFSMounts := plan.FSMounts` preservation
4. **Create function** (around line 1052): Add `plan.FSMounts = planFSMounts` restoration
5. **Read function** (around line 1077): Add `originalFSMounts := state.FSMounts` preservation
6. **Read function** (around line 1195): Add `state.FSMounts = originalFSMounts` restoration (or merge if BCM persists data)
7. **Update function** (around line 1321): Add `planFSMounts := plan.FSMounts` preservation
8. **Update function** (around line 1390): Add `plan.FSMounts = planFSMounts` restoration

### Implementation Pattern Reference

Follow the existing fsexports serialization pattern in buildAPIEntity (lines 1847-1871):

```go
// Serialize fsexports (snake_case -> camelCase for BCM API)
if !model.FSExports.IsNull() && !model.FSExports.IsUnknown() {
    var exports []FSExportModel
    diags := model.FSExports.ElementsAs(ctx, &exports, false)
    if !diags.HasError() {
        exportsList := make([]map[string]interface{}, 0, len(exports))
        for _, export := range exports {
            exportMap := map[string]interface{}{
                "baseType": "FSExport",
                "path":     export.Path.ValueString(),
                ...
            }
            exportsList = append(exportsList, exportMap)
        }
        entity["fsexports"] = exportsList
    }
}
```

### BCM API Field Mapping (Terraform to BCM)

| Terraform Field | BCM API Field | Notes |
|-----------------|---------------|-------|
| device | device | Direct mapping |
| mountpoint | path | Name difference |
| filesystem | type | Name difference |
| mountoptions | options | Name difference |
| fsck | fsck | Direct mapping (if supported) |
| dump | dump | Direct mapping (if supported) |
| rdma | rdma | Direct mapping (if supported) |
| uuid | uuid | Computed, BCM-assigned |

### Test Verification

The fix should be verified with:

1. **Acceptance tests**:
   - `TestAccCMDeviceCategory_FSMountsBasic` - verify basic fsmount configuration
   - `TestAccCMDeviceCategory_FSMountsMultiple` - verify multiple mounts
   - `TestAccCMDeviceCategory_FSMountsUpdate` - verify add/modify/remove mounts
   - `TestAccCMDeviceCategory_FSMountsIdempotency` - verify no drift after apply
   - `TestAccCMDeviceCategory_FSMountsImport` - verify import includes mounts

## Assumptions

- BCM API accepts and returns `fsmounts` array in category entities (structure documented in API docs shows FSMount with path, device, type, options)
- The FSMount baseType is "FSMount" (consistent with other BCM entity patterns)
- BCM may or may not persist fsmounts data long-term (if not, we follow fsexports preservation pattern)
- Field mapping between Terraform schema and BCM API follows documented pattern (mountpoint -> path, filesystem -> type, mountoptions -> options)
- Existing FSMountModel struct and schema definition are correct and complete
