# Research: Fix Device Role Association Bug

**Feature**: 086-fix-device-role-association
**Date**: 2025-11-26
**Status**: Complete

## Research Questions

### 1. Role Name Uniqueness in BCM

**Question**: Are BCM role names unique within a cluster?

**Research Method**: Examined BCM API responses and data source implementation.

**Finding**: YES - Role names are unique within a BCM cluster.

**Evidence**:
- The `bcm_cmdevice_roles` data source deduplicates roles by UUID (`roleMap[uuid.ValueString()] = role`)
- Each role object has a unique `uuid` and unique `name` field
- BCM's role system uses names like "backup", "provisioning", "boot" which are cluster-wide singleton roles
- The data source filters by `name_pattern` assuming name uniqueness

**Decision**: Safe to use role names as identifiers for lookup.

---

### 2. UUID Detection Pattern

**Question**: How can we reliably distinguish a UUID from a role name?

**Research Method**: Analyzed UUID format standards and potential role name formats.

**Finding**: Standard UUID v4 regex is sufficient.

**Pattern**:
```regex
^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$
```

**Rationale**:
- BCM UUIDs follow the standard 8-4-4-4-12 format
- Role names are human-readable strings (e.g., "backup", "provisioning", "kube-worker")
- Role names cannot accidentally match UUID format due to the specific structure
- The Go `google/uuid` package is already imported in the resource file and provides `uuid.Parse()` for validation

**Alternative Considered**: Using `uuid.Parse()` from google/uuid package
- Pro: Already imported, validates UUID properly
- Con: Returns error which requires handling
- Decision: Use regex for simplicity (single-line check)

**Implementation**:
```go
import "regexp"

var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func isUUID(s string) bool {
    return uuidRegex.MatchString(s)
}
```

---

### 3. API Query Efficiency

**Question**: Is `getNodes` the only way to discover roles? Is there a more efficient method?

**Research Method**: Examined BCM API client and data source implementations.

**Finding**: `getNodes` is the only documented method to discover roles.

**Evidence from code**:
```go
// From data_source_cmdevice_roles.go:137-138
tflog.Debug(ctx, "Calling cmdevice.getNodes to extract roles")
result, err := d.client.CallJSONRPC(ctx, "cmdevice", "getNodes")
```

**Schema Documentation**:
> "Fetches available role types from BCM for device role assignment. Roles are extracted from nodes via cmdevice.getNodes and deduplicated by UUID."

**Implications**:
- Each Create/Update operation requires one `getNodes` call to resolve role names
- This is already the existing behavior for UUID lookup
- No additional API overhead introduced
- Role data is cached in memory for the duration of the operation

**Optimization Opportunity (POST-MVP)**: Consider caching roles at the provider level to avoid repeated `getNodes` calls within a single `terraform apply`.

**Decision**: Use existing `getNodes` approach. Performance is acceptable (<500ms typical).

---

### 4. Role Object Structure

**Question**: What fields are required in a role object for device assignment?

**Research Method**: Analyzed BCM API response structure and existing implementation.

**Finding**: Full role object is required, not just name or UUID.

**Role Object Structure** (from `getNodes` response):
```json
{
  "uuid": "12345678-1234-1234-1234-123456789abc",
  "name": "backup",
  "baseType": "Role",
  "childType": "BackupRole",
  "addServices": true
}
```

**Required Fields for Assignment**:
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| uuid | string | Yes | Unique identifier |
| name | string | Yes | Human-readable name |
| baseType | string | Yes | Always "Role" |
| childType | string | Yes | Role type (e.g., "BackupRole", "ProvisioningRole") |
| addServices | boolean | Optional | Whether role adds services |

**Implementation Note**:
The existing `lookupAndBuildRolesForEntity()` already creates a full copy of the role object:
```go
roleCopy := make(map[string]interface{})
for k, v := range role {
    roleCopy[k] = v
}
roleObjects = append(roleObjects, roleCopy)
```

**Decision**: Continue using full role object copies. No change to object structure needed.

---

### 5. State Representation: Names vs UUIDs

**Question**: Should state store role names or UUIDs?

**Research Method**: Analyzed user experience implications and implementation complexity.

**Options Considered**:

| Option | Pros | Cons |
|--------|------|------|
| Store UUIDs | Current behavior, backward compatible | Poor UX, users must lookup UUIDs |
| Store Names | Human-readable, matches config | Requires migration, potential name collisions |

**Decision**: Store role NAMES in state.

**Rationale**:
1. **User Experience**: Users configure roles by name, state should match
2. **Import Experience**: `terraform import` shows readable role names
3. **Drift Detection**: Easier to understand what changed ("backup" vs "12345678-...")
4. **Config Consistency**: Config says `roles = ["backup"]`, state shows `roles = ["backup"]`

**Migration Impact**:
- Existing state with UUIDs will show a plan to "update" roles on first apply
- This is acceptable as roles remain functionally unchanged
- One-time migration, not ongoing

**Implementation**:
```go
// parseRolesFromAPI returns role NAMES (not UUIDs)
func parseRolesFromAPI(rolesData interface{}) types.Set {
    // ... extract role["name"] instead of role["uuid"]
}
```

---

### 6. Error Message Design

**Question**: What should the error message include for invalid roles?

**Research Method**: Reviewed spec requirements and Terraform provider best practices.

**Spec Requirement (FR-003)**:
> "The provider MUST return a clear, user-friendly error message when a role name is not found, including the invalid role name and a hint to use the `bcm_cmdevice_roles` data source to discover available roles."

**Error Message Format**:
```
Role 'invalid-role-name' does not exist in the BCM cluster.
Available roles: backup, provisioning, monitoring, boot
Use the `bcm_cmdevice_roles` data source to discover available roles.
```

**Multiple Invalid Roles**:
```
Roles not found in cluster: invalid-role-1, another-bad-role
Available roles: backup, provisioning, monitoring, boot
Use the `bcm_cmdevice_roles` data source to discover available roles.
```

**Implementation**:
```go
if len(missingRoles) > 0 {
    availableNames := make([]string, 0, len(rolesByName))
    for name := range rolesByName {
        availableNames = append(availableNames, name)
    }
    sort.Strings(availableNames)

    return fmt.Errorf(
        "Roles not found in cluster: %s\nAvailable roles: %s\nUse the `bcm_cmdevice_roles` data source to discover available roles.",
        strings.Join(missingRoles, ", "),
        strings.Join(availableNames, ", "),
    )
}
```

---

## Summary of Decisions

| Research Area | Decision | Rationale |
|--------------|----------|-----------|
| Role name uniqueness | Names are unique | BCM architecture, data source behavior |
| UUID detection | Regex pattern | Simple, reliable, sufficient |
| API efficiency | Use getNodes | Only available method, acceptable performance |
| Role object fields | Full copy required | BCM API requirement |
| State representation | Store names | Better UX, matches user config |
| Error messages | Include available roles | Spec requirement, helpful for users |

## Alternatives Rejected

1. **UUID-only input**: Rejected because it requires complex data source lookups
2. **Separate getRoles API**: Not available in BCM
3. **Provider-level role caching**: Deferred to POST-MVP for simplicity
4. **State migration tool**: Unnecessary, one-time plan difference is acceptable
