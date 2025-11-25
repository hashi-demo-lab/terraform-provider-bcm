# Implementation Plan: BCM Device Roles Block

**Branch**: `039-device-roles-block` | **Date**: 2025-11-26 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/039-device-roles-block/spec.md`

## Summary

Enhance the `bcm_cmdevice_device` resource to support role assignment via a new `roles` attribute. This enables Terraform users to define Kubernetes cluster topology by assigning control-plane, worker, etcd, and other service roles to BCM devices. The implementation follows the existing interfaces block pattern, embedding role assignments in the device entity sent to BCM's addDevice/updateDevice operations.

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**: terraform-plugin-framework v1.16.1, terraform-plugin-testing v1.13.3
**Storage**: N/A (state managed by Terraform, data stored in BCM)
**Testing**: TF_ACC=1 acceptance tests with BCM cluster at 172.21.15.254:8081
**Target Platform**: Linux/macOS/Windows (Terraform provider binary)
**Project Type**: Single project (Terraform provider enhancement)
**Performance Goals**: Role assignment within single API call (<30 seconds)
**Constraints**: BCM API eventual consistency, atomic role assignment
**Scale/Scope**: Enhancement to existing resource (add ~200 LOC)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Notes |
|------|--------|-------|
| TDD-first development | PASS | Acceptance tests defined in spec, will write failing tests first |
| Single responsibility | PASS | Enhances existing resource with related functionality |
| No new abstractions | PASS | Uses existing patterns from interfaces block |
| API pattern compliance | PASS | Follows established BCM JSON-RPC patterns |
| Documentation required | PASS | Examples and docs will be generated via tfplugindocs |

## Project Structure

### Documentation (this feature)

```text
specs/039-device-roles-block/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
internal/provider/
├── resource_cmdevice_device.go       # MODIFY: Add roles attribute and handling
├── resource_cmdevice_device_test.go  # MODIFY: Add roles acceptance tests
└── data_source_cmdevice_roles.go     # REFERENCE: Role parsing patterns

examples/
└── resources/bcm_cmdevice_device/
    └── resource_with_roles.tf        # NEW: Example showing roles usage
```

**Structure Decision**: Enhancement to existing resource files. No new files required except example configuration.

## Complexity Tracking

No constitution violations requiring justification.

---

## Phase 0: Research & Unknowns Resolution

### Research Tasks

1. **BCM API Role Assignment Pattern** - How roles are embedded in device entity for addDevice/updateDevice
2. **Role Object Structure** - Exact fields required when sending roles to BCM
3. **Role Name Resolution** - Whether BCM accepts role names or requires UUIDs
4. **Existing Implementation Reference** - How interfaces block is implemented for pattern consistency

### Research Findings

#### 1. BCM API Role Assignment Pattern

**Decision**: Roles are embedded in the device entity as an array, sent with addDevice/updateDevice operations.

**Rationale**: Analysis of existing BCM API documentation (DeviceEntity.md, CMDevice_Complete_Documentation.md) and data_source_cmdevice_roles.go confirms:
- Roles are a direct property of the Device entity
- BCM returns roles in getDevice/getNode responses
- The pattern matches how interfaces are handled - embedded arrays in the entity

**Evidence**:
```json
// BCM getDevice response showing roles structure
{
  "baseType": "Device",
  "childType": "PhysicalNode",
  "hostname": "master",
  "roles": [
    {
      "baseType": "Role",
      "childType": "HeadNodeRole",
      "name": "headnode",
      "uuid": "a458760f-3898-4870-bd7d-8127a5ea44b8",
      "addServices": true
    }
  ]
}
```

**Alternatives Considered**:
- Separate addRole/removeRole API calls: Not available in BCM API
- Role UUID assignment only: BCM supports name-based assignment

#### 2. Role Object Structure for Create/Update

**Decision**: When creating/updating devices, include roles array with minimal role objects:

```json
{
  "baseType": "Role",
  "name": "worker",
  "modified": true,
  "to_be_removed": false,
  "revision": ""
}
```

**Rationale**: This matches the pattern used for interfaces and other embedded entities. BCM resolves role name to UUID and fills in childType server-side.

**Evidence**: From spec.md API Contract section and existing resource patterns in resource_cmdevice_device.go (interfaces handling at line 1180-1200).

#### 3. Role Name Resolution

**Decision**: Provider sends role names (strings), BCM resolves to UUIDs server-side.

**Rationale**:
- Simplifies user experience (users specify "worker" not UUIDs)
- Matches data source behavior (bcm_cmdevice_roles returns names)
- BCM API validates role names and returns errors for invalid names

**Evidence**: From spec.md assumption ASSUME-002 and role data source implementation.

#### 4. Implementation Pattern Reference

**Decision**: Follow interfaces block pattern from resource_cmdevice_device.go:

1. Add `Roles types.List` to model (line 72 area)
2. Add schema definition in Schema() method (line 86 area)
3. Extend buildDeviceAPIEntity() to include roles array (line 1147 area)
4. Extend parseDeviceFromAPI() to extract roles from response (line 1274 area)

**Rationale**: Consistency with existing code, proven pattern that handles BCM's eventual consistency and state management correctly.

### Unknowns Resolved

| Unknown | Resolution |
|---------|------------|
| API pattern for roles | Embedded in device entity, sent with addDevice/updateDevice |
| Role object structure | Minimal: baseType, name, modified, to_be_removed, revision |
| Name vs UUID | Provider uses names, BCM resolves server-side |
| Implementation pattern | Follow interfaces block pattern |
| Validation | BCM validates role names, provider calls validateDevice |
| Drift detection | Standard Read() extracts roles, Terraform detects differences |

---

## Phase 1: Design & Contracts

### Data Model

See [data-model.md](./data-model.md) for complete entity definitions.

**Key Model Changes**:

```go
// CMDeviceDeviceResourceModel - Add after Interfaces field (line 72)
type CMDeviceDeviceResourceModel struct {
    // ... existing fields ...

    // Roles - list of role names assigned to this device
    Roles types.List `tfsdk:"roles"` // List of StringType
}
```

### API Contracts

See [contracts/](./contracts/) directory for OpenAPI specifications.

**Key API Changes**:

1. **buildDeviceAPIEntity** - Add roles array construction
2. **parseDeviceFromAPI** - Add roles extraction from response
3. **Schema** - Add roles attribute definition

### Implementation Approach

#### Schema Definition

```go
"roles": schema.ListAttribute{
    ElementType:         types.StringType,
    Optional:            true,
    MarkdownDescription: "List of role names to assign to the device (e.g., 'control-plane', 'worker', 'etcd'). " +
        "Use data.bcm_cmdevice_roles to discover available role names. " +
        "Roles are resolved by name - BCM maps names to role UUIDs server-side.",
},
```

#### Build Entity Logic

```go
// In buildDeviceAPIEntityWithExisting, after interfaces handling:

// Add roles if specified
if !plan.Roles.IsNull() && !plan.Roles.IsUnknown() {
    var roleNames []string
    plan.Roles.ElementsAs(ctx, &roleNames, false)

    // Deduplicate role names
    seen := make(map[string]bool)
    uniqueRoles := make([]string, 0, len(roleNames))
    for _, name := range roleNames {
        if !seen[name] && name != "" {
            seen[name] = true
            uniqueRoles = append(uniqueRoles, name)
        }
    }

    rolesArray := make([]map[string]interface{}, 0, len(uniqueRoles))
    for _, roleName := range uniqueRoles {
        role := map[string]interface{}{
            "baseType":      "Role",
            "name":          roleName,
            "modified":      true,
            "to_be_removed": false,
            "revision":      "",
        }
        rolesArray = append(rolesArray, role)
    }
    entity["roles"] = rolesArray
}
```

#### Parse Response Logic

```go
// In parseDeviceFromAPI, after interfaces parsing:

// Parse roles from BCM response
if rolesData, ok := data["roles"].([]interface{}); ok && len(rolesData) > 0 {
    roleNames := make([]string, 0, len(rolesData))
    for _, roleData := range rolesData {
        if role, ok := roleData.(map[string]interface{}); ok {
            if name, ok := role["name"].(string); ok && name != "" {
                roleNames = append(roleNames, name)
            }
        }
    }
    // Sort for consistent state comparison
    sort.Strings(roleNames)
    model.Roles, _ = types.ListValueFrom(ctx, types.StringType, roleNames)
} else {
    model.Roles = types.ListNull(types.StringType)
}
```

### Test Strategy

Seven acceptance tests as defined in spec.md:

1. **TestAccCMDeviceDevice_RolesCreate** - Create with roles
2. **TestAccCMDeviceDevice_RolesUpdate** - Update roles
3. **TestAccCMDeviceDevice_RolesRemove** - Remove all roles
4. **TestAccCMDeviceDevice_RolesImport** - Import with roles
5. **TestAccCMDeviceDevice_RolesDrift** - Drift detection
6. **TestAccCMDeviceDevice_RolesMultiple** - Multiple roles
7. **TestAccCMDeviceDevice_RolesIdempotent** - Idempotency verification

### Quick Start Guide

See [quickstart.md](./quickstart.md) for developer setup and testing instructions.

---

## Post-Design Constitution Re-Check

| Gate | Status | Notes |
|------|--------|-------|
| TDD-first development | PASS | Test definitions complete, ready for RED phase |
| Single responsibility | PASS | Roles enhancement is cohesive with device resource |
| No new abstractions | PASS | Reuses existing List attribute pattern |
| API pattern compliance | PASS | Follows established BCM JSON-RPC entity patterns |
| Documentation required | PASS | Schema includes MarkdownDescription, examples planned |

---

## Next Steps

1. Run `/speckit.tasks` to generate task breakdown
2. Execute RED phase: Write failing acceptance tests
3. Execute GREEN phase: Implement minimal code to pass tests
4. Execute REFACTOR phase: Improve code quality
5. Run `make generate` to update documentation
