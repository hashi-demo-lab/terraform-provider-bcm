# Contract: Role Validation

**Feature**: 086-fix-device-role-association
**Date**: 2025-11-26
**Type**: Internal Provider Contract

## Overview

This document defines the contract for role validation in the `bcm_cmdevice_device` resource. The provider performs client-side validation of role identifiers before sending requests to the BCM API.

---

## Validation Function Contract

### Function Signature

```go
func (r *CMDeviceDeviceResource) lookupAndBuildRolesForEntity(
    ctx context.Context,
    plan CMDeviceDeviceResourceModel,
    entity map[string]interface{},
) error
```

### Input Contract

| Parameter | Type | Description |
|-----------|------|-------------|
| ctx | context.Context | Request context for logging and cancellation |
| plan | CMDeviceDeviceResourceModel | Terraform plan containing roles set |
| entity | map[string]interface{} | Device entity being built for API |

### plan.Roles Field

| Condition | Handling |
|-----------|----------|
| IsNull() | Return nil (no roles to assign) |
| IsUnknown() | Return nil (deferred evaluation) |
| Empty set | Set entity["roles"] = [] (explicit removal) |
| Non-empty set | Validate and resolve each identifier |

### Output Contract

| Condition | Return Value | Side Effect |
|-----------|--------------|-------------|
| Success | nil | entity["roles"] populated with role objects |
| Invalid identifier | error with message | entity unchanged |
| API failure | error with cause | entity unchanged |

---

## Identifier Resolution Contract

### UUID Detection

```go
func isUUID(s string) bool
```

**Pattern**: `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`

| Input | isUUID() | Lookup Method |
|-------|----------|---------------|
| "backup" | false | By name |
| "kube-worker" | false | By name |
| "12345678-1234-1234-1234-123456789abc" | true | By UUID |
| "ABCDEF12-3456-7890-ABCD-EF1234567890" | true | By UUID (case-insensitive) |
| "" | false | Error (empty string) |
| "backup-12345" | false | By name |

### Resolution Priority

1. Check for empty string -> Error
2. Check isUUID() -> If true, lookup by UUID
3. Otherwise -> Lookup by name

---

## Error Message Contract

### Empty String Error

**Condition**: One or more role identifiers are empty strings.

**Format**:
```
Invalid role identifiers found: (empty string) - role identifiers must be non-empty strings
```

### Role Not Found Error

**Condition**: One or more role identifiers (name or UUID) do not exist in the cluster.

**Format** (single):
```
Role '<identifier>' does not exist in the BCM cluster.
Available roles: <comma-separated sorted list>
Use the `bcm_cmdevice_roles` data source to discover available roles.
```

**Format** (multiple):
```
Roles not found in cluster: <comma-separated list>
Available roles: <comma-separated sorted list>
Use the `bcm_cmdevice_roles` data source to discover available roles.
```

### API Error

**Condition**: BCM API call fails.

**Format**:
```
Failed to query nodes for role lookup: <underlying error>
```

---

## Role Object Contract

### Required Fields for BCM API

When assigning roles to a device, each role object must include:

```go
type RoleObject struct {
    UUID        string `json:"uuid"`
    Name        string `json:"name"`
    BaseType    string `json:"baseType"`
    ChildType   string `json:"childType"`
    AddServices bool   `json:"addServices,omitempty"`
}
```

### Copy Semantics

Role objects are copied (not referenced) when added to the device entity:

```go
roleCopy := make(map[string]interface{})
for k, v := range role {
    roleCopy[k] = v
}
roleObjects = append(roleObjects, roleCopy)
```

This prevents mutation of the cached role data.

---

## State Representation Contract

### parseRolesFromAPI Function

```go
func parseRolesFromAPI(rolesData interface{}) types.Set
```

### Contract

| Input | Output |
|-------|--------|
| nil | types.SetNull(types.StringType) |
| Empty array | types.SetNull(types.StringType) |
| Array of role objects | types.Set containing role names |

### Role Name Extraction

```go
for _, roleItem := range rolesArray {
    if role, ok := roleItem.(map[string]interface{}); ok {
        if name, ok := role["name"].(string); ok && name != "" {
            roleNames = append(roleNames, name)
        }
    }
}
```

---

## Backward Compatibility Contract

### UUID Input Support

| Input Format | Support Level | Notes |
|--------------|---------------|-------|
| Role name | Primary | Recommended approach |
| Role UUID | Supported | Backward compatibility |
| Mixed | Supported | Both resolved correctly |

### State Migration

| Current State | After First Apply | Notes |
|---------------|-------------------|-------|
| `["<uuid>"]` | `["<role-name>"]` | One-time migration |
| `["<name>"]` | `["<name>"]` | No change |

---

## Performance Contract

| Operation | Target | Notes |
|-----------|--------|-------|
| Role resolution | < 500ms | Single getNodes call |
| Validation | < 100ms | In-memory map lookup |
| Total overhead | < 600ms | Per Create/Update |

---

## Testing Contract

### Required Test Cases

1. **Valid name input**: `roles = ["backup"]` -> Success
2. **Valid UUID input**: `roles = ["12345678-..."]` -> Success
3. **Mixed input**: `roles = ["backup", "12345678-..."]` -> Success
4. **Invalid name**: `roles = ["nonexistent"]` -> Error with message
5. **Invalid UUID**: `roles = ["99999999-..."]` -> Error with message
6. **Empty string**: `roles = [""]` -> Error with message
7. **Empty set**: `roles = []` -> Success, removes all roles
8. **Null set**: (omitted) -> Success, no role change

### State Verification

After successful role assignment:
- State contains role **names** (not UUIDs)
- State matches the role names in BCM
- Import produces same state representation
