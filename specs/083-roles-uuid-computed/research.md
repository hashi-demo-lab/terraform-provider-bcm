# Research: Fix roles[].uuid Computed Value Population

**Issue**: #83 - bcm_cmdevice_category: roles[].uuid computed value never populated from BCM API
**Date**: 2025-11-26

## Research Tasks

### 1. BCM API Role Response Structure

**Task**: Verify BCM API returns role UUIDs in category response

**Finding**: Confirmed via code analysis at lines 2249-2276 of `resource_cmdevice_category.go`

**Decision**: BCM API DOES return role UUIDs - the parsing code already exists and works correctly

**Evidence**:
```go
// Lines 2256-2272 in resource_cmdevice_category.go
if rolesData, ok := categoryData["roles"].([]interface{}); ok {
    for _, roleRaw := range rolesData {
        if roleMap, ok := roleRaw.(map[string]interface{}); ok {
            roleObj, objDiags := types.ObjectValue(roleObjectType.AttrTypes, map[string]attr.Value{
                "name":         getStringValue(roleMap, "name"),
                "child_type":   getStringValue(roleMap, "childType"),
                "uuid":         getStringValue(roleMap, "uuid"),  // <-- UUID is parsed
                "add_services": getBoolValue(roleMap, "addServices"),
            })
            // ...
        }
    }
}
```

**Alternatives Considered**: None - the existing code correctly parses UUIDs

---

### 2. Root Cause of UUID Loss

**Task**: Identify why correctly-parsed UUIDs are lost

**Finding**: Line 1193 unconditionally overwrites API response with original state

**Decision**: Replace unconditional overwrite with merge logic

**Evidence**:
```go
// Line 1075: Original state captured
originalRoles := state.Roles

// After readCategory() correctly populates state.Roles with UUIDs...

// Line 1193: BUG - Overwrites API data
state.Roles = originalRoles  // Discards BCM-assigned UUIDs!
```

**Rationale for Original Code**: The comment states BCM doesn't persist these fields. This is true for some fields (static_routes, fsexports) but NOT for roles.

**Alternatives Considered**:
1. Remove preservation entirely - Rejected (may cause drift for add_services)
2. Conditional preservation based on null check - Rejected (doesn't merge computed fields)

---

### 3. Role Matching Strategy

**Task**: Determine how to match config roles with API roles

**Finding**: Role `name` is unique within a category

**Decision**: Match roles by `name` attribute

**Evidence**:
```go
// Line 1874: Role name used as primary identifier
roleMap := map[string]interface{}{
    "baseType":  "Role",
    "name":      role.Name.ValueString(),  // <-- Primary identifier
    "childType": role.ChildType.ValueString(),
}
```

**Alternatives Considered**:
1. Match by index - Rejected (order may change)
2. Match by UUID - Not possible (UUID is computed, not known at config time)

---

### 4. CategoryRoleModel Field Analysis

**Task**: Identify which fields are user-specified vs computed

**Finding**: Field classification from schema analysis

| Field | Type | Source |
|-------|------|--------|
| `name` | Required | User config |
| `child_type` | Required | User config |
| `add_services` | Optional | User config |
| `uuid` | Computed | BCM API |

**Decision**: Merge preserves user fields, populates computed fields

**Evidence** (from line 194-199):
```go
type CategoryRoleModel struct {
    Name        types.String `tfsdk:"name"`         // Required, role name
    ChildType   types.String `tfsdk:"child_type"`   // Required, role type
    UUID        types.String `tfsdk:"uuid"`         // Computed, BCM-assigned
    AddServices types.Bool   `tfsdk:"add_services"` // Optional, add role services
}
```

---

### 5. Existing Test Coverage

**Task**: Understand current test status for roles

**Finding**: Tests exist but avoid roles due to known bug

**Evidence** (from test file lines 3857-3861):
```go
// TestAccCMDeviceCategoryResource_RolesConfiguration tests the roles list field.
// Note: BCM does not persist roles after category creation (returns null).
// The provider has a bug where roles[0].uuid remains Unknown after apply.
// This test creates a category without roles to verify basic functionality.
```

**Decision**: Update existing test and add new tests that verify UUID population

---

## Summary of Decisions

| Decision | Rationale |
|----------|-----------|
| Use merge-by-name strategy | Role name is unique, reliable identifier |
| Preserve user config fields | Avoid false drift detection |
| Populate UUID from API | Fix the actual bug |
| Add new tests first (TDD) | Following project constitution |

## No NEEDS CLARIFICATION Items

All technical questions resolved through code analysis. No external research or user clarification needed.
