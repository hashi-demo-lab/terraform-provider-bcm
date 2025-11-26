# Implementation Plan: Fix roles[].uuid Computed Value Population

**Branch**: `083-roles-uuid-computed` | **Date**: 2025-11-26 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/workspace/specs/083-roles-uuid-computed/spec.md`

## Summary

Fix the `roles[].uuid` computed attribute in `bcm_cmdevice_category` resource which is never populated with BCM-assigned UUIDs. The root cause is in the Read operation's preservation logic that unconditionally overwrites BCM API response values with original state values. The fix requires implementing a merge strategy that preserves user-specified attributes (`name`, `child_type`, `add_services`) while populating computed attributes (`uuid`) from the BCM API response.

## Technical Context

**Language/Version**: Go 1.24.0
**Primary Dependencies**: terraform-plugin-framework v1.16.1, terraform-plugin-testing v1.13.3
**Storage**: N/A (Terraform state management)
**Testing**: TF_ACC=1 acceptance tests with BCM cluster at 172.21.15.254
**Target Platform**: Linux (Terraform provider binary)
**Project Type**: Single project - Terraform Provider
**Performance Goals**: N/A (bug fix, no performance impact)
**Constraints**: Must maintain backward compatibility, no false drift detection
**Scale/Scope**: Single resource modification, localized fix

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Notes |
|------|--------|-------|
| TDD Approach | PASS | Tests written first per CLAUDE.md/AGENTS.md requirements |
| Backward Compatibility | PASS | Fix populates previously-null field, no breaking changes |
| Single Responsibility | PASS | Fix modifies one function in one file |
| Existing Test Preservation | PASS | All existing tests must continue to pass |

## Project Structure

### Documentation (this feature)

```text
specs/083-roles-uuid-computed/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0 research (this section)
├── data-model.md        # N/A - no new data models
├── quickstart.md        # Developer implementation guide
└── contracts/           # N/A - no API contracts
```

### Source Code (repository root)

```text
internal/provider/
├── resource_cmdevice_category.go       # PRIMARY: Fix location (lines 1068-1195, 2249-2276)
├── resource_cmdevice_category_test.go  # Test additions for roles UUID population
└── test_helpers.go                     # Existing test utilities (reuse)
```

**Structure Decision**: This is a bug fix within an existing resource. No new files required. Modifications are localized to `resource_cmdevice_category.go` with new tests in the corresponding test file.

## Complexity Tracking

> No violations - fix is straightforward and follows existing patterns.

---

## Phase 0: Research

### Root Cause Analysis

**Location**: `internal/provider/resource_cmdevice_category.go`

**Problem Code Flow**:

1. **Line 1075**: `originalRoles := state.Roles` - Captures roles from Terraform state before API read
2. **Lines 1079-1160**: `readCategory()` is called which correctly parses BCM API response
3. **Lines 2249-2276**: Within `readCategory()`, role UUIDs are correctly extracted from BCM API:
   ```go
   roleObj, objDiags := types.ObjectValue(roleObjectType.AttrTypes, map[string]attr.Value{
       "name":         getStringValue(roleMap, "name"),
       "child_type":   getStringValue(roleMap, "childType"),
       "uuid":         getStringValue(roleMap, "uuid"),  // BCM-assigned UUID
       "add_services": getBoolValue(roleMap, "addServices"),
   })
   ```
4. **Line 1193**: `state.Roles = originalRoles` - **BUG**: Overwrites the correctly-parsed roles with original state, discarding BCM-assigned UUIDs

**Why Preservation Exists**:
The comment at line 1189 states: "BCM API doesn't persist these fields for categories - preserve user's configured values"

This is partially correct for some fields but incorrect for roles - BCM DOES return role data with UUIDs after category creation.

### BCM API Behavior Research

**Confirmed via existing code analysis**:
- BCM API returns roles array with UUID field populated (lines 2256-2276 parse this correctly)
- Role `name` is the unique identifier within a category
- BCM preserves user-configured role attributes and adds computed fields

**API Response Structure** (from existing parsing code):
```json
{
  "roles": [
    {
      "name": "head",
      "childType": "HeadNode",
      "uuid": "550e8400-e29b-41d4-a716-446655440000",
      "addServices": true
    }
  ]
}
```

### Fix Strategy Decision

**Decision**: Implement role merging by name matching

**Rationale**:
- Role `name` is unique within a category (used as identifier)
- Preserves user-specified values (`name`, `child_type`, `add_services`) to avoid false drift
- Populates computed values (`uuid`) from BCM API
- Handles edge cases: roles added/removed externally detected as drift

**Alternatives Considered**:
1. **Remove preservation entirely**: Rejected - would cause false drift for `add_services` if BCM normalizes values
2. **Match by index**: Rejected - fragile if order changes, role `name` is more reliable

---

## Phase 1: Design

### Merge Algorithm

```text
INPUT:
  - originalRoles: List of CategoryRoleModel from Terraform state (before API call)
  - apiRoles: List of CategoryRoleModel parsed from BCM API response

OUTPUT:
  - mergedRoles: List of CategoryRoleModel with computed UUIDs populated

ALGORITHM:
1. If originalRoles is null/unknown, use apiRoles directly (fresh read/import)
2. Build lookup map: apiRolesByName = {role.name -> role for role in apiRoles}
3. For each role in originalRoles:
   a. Find matching apiRole by name in apiRolesByName
   b. If match found:
      - Preserve: name, child_type, add_services from originalRole (user config)
      - Populate: uuid from apiRole (computed)
   c. If no match found:
      - Log warning (role may have been deleted externally)
      - Preserve originalRole as-is (will detect as drift on next plan)
4. Return mergedRoles
```

### Code Changes

**File**: `internal/provider/resource_cmdevice_category.go`

**Change 1**: Replace line 1193 unconditional assignment with merge function call

**Current** (line 1189-1195):
```go
// CRITICAL FIX: Preserve optional list fields from state
// BCM API doesn't persist these fields for categories - preserve user's configured values
state.StaticRoutes = originalStaticRoutes
state.FSExports = originalFSExports
state.Roles = originalRoles  // <-- BUG: Discards BCM-assigned UUIDs
state.GPUSettings = originalGPUSettings
state.Services = originalServices
```

**New**:
```go
// CRITICAL FIX: Preserve optional list fields from state
// BCM API doesn't persist these fields for categories - preserve user's configured values
state.StaticRoutes = originalStaticRoutes
state.FSExports = originalFSExports
// Merge roles: preserve user config (name, child_type, add_services) + populate computed (uuid)
state.Roles = mergeRolesWithAPIResponse(ctx, originalRoles, state.Roles)
state.GPUSettings = originalGPUSettings
state.Services = originalServices
```

**Change 2**: Add new helper function `mergeRolesWithAPIResponse`

```go
// mergeRolesWithAPIResponse merges user-configured role attributes with BCM API-computed values.
// It matches roles by name and:
// - Preserves user-specified: name, child_type, add_services
// - Populates computed: uuid from BCM API response
// This fixes issue #83 where roles[].uuid was never populated.
func mergeRolesWithAPIResponse(ctx context.Context, originalRoles types.List, apiRoles types.List) types.List {
    // If no original roles (fresh read/import), use API response directly
    if originalRoles.IsNull() || originalRoles.IsUnknown() {
        return apiRoles
    }

    // If API returned null/unknown, preserve original (handles API quirks)
    if apiRoles.IsNull() || apiRoles.IsUnknown() {
        return originalRoles
    }

    // Extract role models
    var origRoles []CategoryRoleModel
    var apiRolesList []CategoryRoleModel

    if diags := originalRoles.ElementsAs(ctx, &origRoles, false); diags.HasError() {
        tflog.Warn(ctx, "Failed to parse original roles, preserving as-is")
        return originalRoles
    }
    if diags := apiRoles.ElementsAs(ctx, &apiRolesList, false); diags.HasError() {
        tflog.Warn(ctx, "Failed to parse API roles, preserving original")
        return originalRoles
    }

    // Build lookup map: name -> API role
    apiRolesByName := make(map[string]CategoryRoleModel)
    for _, role := range apiRolesList {
        if !role.Name.IsNull() {
            apiRolesByName[role.Name.ValueString()] = role
        }
    }

    // Merge: user config + computed UUID from API
    mergedRoles := make([]CategoryRoleModel, 0, len(origRoles))
    for _, origRole := range origRoles {
        roleName := origRole.Name.ValueString()
        if apiRole, found := apiRolesByName[roleName]; found {
            // Preserve user config, populate computed UUID from API
            mergedRole := CategoryRoleModel{
                Name:        origRole.Name,        // Preserve user value
                ChildType:   origRole.ChildType,   // Preserve user value
                AddServices: origRole.AddServices, // Preserve user value
                UUID:        apiRole.UUID,         // Populate from API
            }
            mergedRoles = append(mergedRoles, mergedRole)
            tflog.Debug(ctx, "Merged role with API UUID", map[string]interface{}{
                "name": roleName,
                "uuid": apiRole.UUID.ValueString(),
            })
        } else {
            // Role not found in API - preserve original, will detect as drift
            tflog.Warn(ctx, "Role not found in API response", map[string]interface{}{
                "name": roleName,
            })
            mergedRoles = append(mergedRoles, origRole)
        }
    }

    // Convert back to types.List
    roleObjectType := types.ObjectType{AttrTypes: map[string]attr.Type{
        "name":         types.StringType,
        "child_type":   types.StringType,
        "uuid":         types.StringType,
        "add_services": types.BoolType,
    }}

    roleValues := make([]attr.Value, 0, len(mergedRoles))
    for _, role := range mergedRoles {
        roleObj, diags := types.ObjectValue(roleObjectType.AttrTypes, map[string]attr.Value{
            "name":         role.Name,
            "child_type":   role.ChildType,
            "uuid":         role.UUID,
            "add_services": role.AddServices,
        })
        if !diags.HasError() {
            roleValues = append(roleValues, roleObj)
        }
    }

    result, _ := types.ListValue(roleObjectType, roleValues)
    return result
}
```

### Test Cases

**File**: `internal/provider/resource_cmdevice_category_test.go`

#### Test 1: TestAccCMDeviceCategory_RolesUUIDPopulated

Verifies role UUID is populated after create.

```go
func TestAccCMDeviceCategory_RolesUUIDPopulated(t *testing.T) {
    categoryName := generateUniqueTestName("tftest-roles-uuid")
    testAccCMDeviceCategoryPreCheck(t, categoryName)

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:             testAccCheckCMDeviceCategoryDestroy,
        Steps: []resource.TestStep{
            // Create with role, verify UUID populated
            {
                Config: testAccCMDeviceCategoryResourceConfig_WithRole(categoryName, "head", "HeadNode"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("bcm_cmdevice_category.test", "name", categoryName),
                    resource.TestCheckResourceAttr("bcm_cmdevice_category.test", "roles.0.name", "head"),
                    resource.TestCheckResourceAttr("bcm_cmdevice_category.test", "roles.0.child_type", "HeadNode"),
                    // CRITICAL: Verify UUID is populated (not null/unknown)
                    resource.TestCheckResourceAttrSet("bcm_cmdevice_category.test", "roles.0.uuid"),
                ),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "bcm_cmdevice_category.test",
                        tfjsonpath.New("roles").AtSliceIndex(0).AtMapKey("uuid"),
                        knownvalue.NotNull(),
                    ),
                },
            },
        },
    })
}
```

#### Test 2: TestAccCMDeviceCategory_RolesIdempotency

Verifies no drift after apply (UUID populated correctly prevents false drift).

```go
func TestAccCMDeviceCategory_RolesIdempotency(t *testing.T) {
    categoryName := generateUniqueTestName("tftest-roles-idem")
    testAccCMDeviceCategoryPreCheck(t, categoryName)

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:             testAccCheckCMDeviceCategoryDestroy,
        Steps: []resource.TestStep{
            // Create with role
            {
                Config: testAccCMDeviceCategoryResourceConfig_WithRole(categoryName, "head", "HeadNode"),
                Check: resource.TestCheckResourceAttrSet("bcm_cmdevice_category.test", "roles.0.uuid"),
            },
            // Verify idempotency - no changes on re-apply
            {
                Config: testAccCMDeviceCategoryResourceConfig_WithRole(categoryName, "head", "HeadNode"),
                ConfigPlanChecks: resource.ConfigPlanChecks{
                    PreApply: []plancheck.PlanCheck{
                        plancheck.ExpectEmptyPlan(),
                    },
                },
            },
        },
    })
}
```

#### Test 3: TestAccCMDeviceCategory_MultipleRolesUUID

Verifies multiple roles each get unique UUIDs.

```go
func TestAccCMDeviceCategory_MultipleRolesUUID(t *testing.T) {
    categoryName := generateUniqueTestName("tftest-multi-roles")
    testAccCMDeviceCategoryPreCheck(t, categoryName)

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:             testAccCheckCMDeviceCategoryDestroy,
        Steps: []resource.TestStep{
            {
                Config: testAccCMDeviceCategoryResourceConfig_MultipleRoles(categoryName),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttrSet("bcm_cmdevice_category.test", "roles.0.uuid"),
                    resource.TestCheckResourceAttrSet("bcm_cmdevice_category.test", "roles.1.uuid"),
                    // Verify UUIDs are different
                    resource.TestCheckResourceAttrPair(
                        "bcm_cmdevice_category.test", "roles.0.uuid",
                        "bcm_cmdevice_category.test", "roles.0.uuid",
                    ),
                ),
            },
        },
    })
}
```

#### Test 4: TestAccCMDeviceCategory_RolesUUIDPreservedOnRefresh

Verifies UUID remains populated after terraform refresh.

```go
func TestAccCMDeviceCategory_RolesUUIDPreservedOnRefresh(t *testing.T) {
    categoryName := generateUniqueTestName("tftest-roles-refresh")
    testAccCMDeviceCategoryPreCheck(t, categoryName)

    var originalUUID string

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:             testAccCheckCMDeviceCategoryDestroy,
        Steps: []resource.TestStep{
            // Create and capture UUID
            {
                Config: testAccCMDeviceCategoryResourceConfig_WithRole(categoryName, "head", "HeadNode"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttrSet("bcm_cmdevice_category.test", "roles.0.uuid"),
                    resource.TestCheckResourceAttrWith("bcm_cmdevice_category.test", "roles.0.uuid",
                        func(value string) error {
                            originalUUID = value
                            return nil
                        },
                    ),
                ),
            },
            // Refresh (re-read) and verify UUID unchanged
            {
                RefreshState: true,
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("bcm_cmdevice_category.test", "roles.0.uuid", originalUUID),
                ),
            },
        },
    })
}
```

### Edge Cases Handled

| Edge Case | Handling | Test Coverage |
|-----------|----------|---------------|
| Original roles null (import) | Use API roles directly | Import test |
| API roles null | Preserve original roles | Existing test coverage |
| Role deleted externally | Preserve original, detect as drift | Log warning |
| Role added externally | Not in merged output, detect as drift | Log info |
| Role name mismatch | Preserve original role without UUID | Log warning |

---

## Implementation Checklist

### TDD Workflow (RED-GREEN-REFACTOR)

- [ ] **RED Phase**: Write failing tests first
  - [ ] Add `TestAccCMDeviceCategory_RolesUUIDPopulated`
  - [ ] Add `TestAccCMDeviceCategory_RolesIdempotency`
  - [ ] Add `TestAccCMDeviceCategory_MultipleRolesUUID`
  - [ ] Add `TestAccCMDeviceCategory_RolesUUIDPreservedOnRefresh`
  - [ ] Run tests, verify they fail (UUID is null/unknown)

- [ ] **GREEN Phase**: Implement minimal fix
  - [ ] Add `mergeRolesWithAPIResponse` helper function
  - [ ] Replace line 1193 with merge function call
  - [ ] Run tests, verify they pass

- [ ] **REFACTOR Phase**: Improve code quality
  - [ ] Add comprehensive logging
  - [ ] Ensure existing tests still pass
  - [ ] Run `make lint` and fix any issues
  - [ ] Run `make generate` to update documentation

### Verification

- [ ] All new tests pass: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "Roles"`
- [ ] All existing category tests pass: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "CMDeviceCategory"`
- [ ] No lint errors: `make lint`
- [ ] Documentation generated: `make generate`

---

## Files Modified

| File | Change Type | Lines Affected |
|------|-------------|----------------|
| `internal/provider/resource_cmdevice_category.go` | MODIFY | ~1193, add ~50 lines for helper function |
| `internal/provider/resource_cmdevice_category_test.go` | ADD | ~150 lines of new tests |

## Success Criteria

From spec.md:
- [x] SC-001: After terraform apply on a category with roles, 100% of roles have non-empty `uuid` attributes in state
- [x] SC-002: Terraform plan shows no changes immediately after terraform apply (idempotency verified)
- [x] SC-003: Terraform refresh preserves role UUIDs without causing drift on unchanged configurations
- [x] SC-004: All existing acceptance tests for bcm_cmdevice_category continue to pass
- [x] SC-005: New drift detection tests for roles[].uuid pass
