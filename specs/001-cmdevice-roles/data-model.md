# Data Model: BCM Device Roles

**Feature**: bcm_cmdevice_roles data source
**Date**: 2025-11-25
**Phase**: Phase 1 - Design & Contracts

## Entity: Role

Represents a BCM role that can be assigned to devices. Roles define service configurations and capabilities for cluster nodes.

### Attributes

| Attribute | Type | Required | Description | BCM API Field | Null-Safe Helper |
|-----------|------|----------|-------------|---------------|------------------|
| `id` | string | Computed | Role identifier (same as uuid) | `uuid` | getStringValue |
| `uuid` | string | Computed | Unique role identifier | `uuid` | getStringValue |
| `name` | string | Computed | Role name (e.g., "headnode", "kube-master") | `name` | getStringValue |
| `child_type` | string | Computed | Role type (e.g., HeadNodeRole, ComputeRole) | `childType` | getStringValue |
| `base_type` | string | Computed | Always "Role" | `baseType` | getStringValue |
| `add_services` | bool | Computed | Whether role adds services to node | `addServices` | getBoolValue |

### Relationships

**Embedded in Device Objects**:
- Roles are stored in the `roles` array field of Device objects
- Accessed via `cmdevice.getNodes` API call
- Each node can have zero or more roles assigned

**Deduplication**:
- Roles are deduplicated by UUID across all nodes
- Same role UUID represents the same role definition
- Multiple nodes can share the same role

**Discovery Pattern**:
```
cmdevice.getNodes → nodes[] → node.roles[] → unique roles by UUID
```

### State Management

**Read-Only Data Source**:
- No persistence required (data source only reads, never writes)
- Terraform state stores filtered role list with all attributes
- State refreshes on each `terraform plan` or `terraform apply`

**ID Computation**:
```go
// Static ID for data source (roles don't change frequently)
data.ID = types.StringValue("roles")

// Alternative: Hash of all role UUIDs for change detection
// hasher := sha256.New()
// for _, role := range roles {
//     hasher.Write([]byte(role.UUID.ValueString()))
// }
// data.ID = types.StringValue(hex.EncodeToString(hasher.Sum(nil)))
```

### Known Role Types (childType)

Based on BCM architecture and API exploration:

| childType | Description | Common Use Cases |
|-----------|-------------|------------------|
| HeadNodeRole | Cluster management node | Primary controller, cluster services |
| StorageRole | NFS storage services | Shared storage for compute nodes |
| BackupRole | Backup services | Data backup and recovery |
| MonitoringRole | Cluster monitoring agent | Metrics collection, alerting |
| ProvisioningRole | Node provisioning services | PXE boot, image deployment |
| BootRole | PXE boot services | Network boot infrastructure |
| ComputeRole | Compute workload execution | Job execution, workload scheduling |

**Note**: Custom roles may exist with user-defined childType values.

## Filtering Model

### Filter Attributes

| Attribute | Type | Optional | Description | Matching Logic |
|-----------|------|----------|-------------|----------------|
| `name_pattern` | string | Yes | Glob pattern for role name | filepath.Match (wildcards: *, ?, [abc]) |
| `child_type` | string | Yes | Exact match for role type | Exact string comparison (case-sensitive) |

### Filter Semantics

**AND Logic** (all specified filters must match):
- No filters → return all roles
- `name_pattern` only → return roles matching pattern
- `child_type` only → return roles with exact type match
- Both filters → return roles matching BOTH pattern AND type

**Null/Empty Handling**:
```go
// Ignore null/unknown filter values
if !namePattern.IsNull() && !namePattern.IsUnknown() {
    // Apply filter
}
```

### Example Filter Queries

```hcl
# Query all roles
data "bcm_cmdevice_roles" "all" {}

# Filter by type (exact match)
data "bcm_cmdevice_roles" "compute" {
  child_type = "ComputeRole"
}

# Filter by name pattern (glob)
data "bcm_cmdevice_roles" "kube_roles" {
  name_pattern = "kube-*"
}

# Combined filters (AND logic)
data "bcm_cmdevice_roles" "kube_compute" {
  name_pattern = "kube-*"
  child_type   = "ComputeRole"
}
```

## API Response Structure

### BCM API Call

```json
{
  "service": "cmdevice",
  "call": "getNodes"
}
```

### Response Format

```json
[
  {
    "uuid": "node-uuid-123",
    "name": "node01",
    "hostname": "node01.example.com",
    "roles": [
      {
        "baseType": "Role",
        "childType": "HeadNodeRole",
        "name": "headnode",
        "uuid": "role-uuid-456",
        "addServices": true,
        "modified": false,
        "to_be_removed": false,
        "revision": ""
      }
    ]
  },
  {
    "uuid": "node-uuid-789",
    "name": "node02",
    "roles": [
      {
        "baseType": "Role",
        "childType": "ComputeRole",
        "name": "compute",
        "uuid": "role-uuid-abc",
        "addServices": false,
        "modified": false,
        "to_be_removed": false,
        "revision": ""
      }
    ]
  }
]
```

### Field Mapping

| BCM API Field | Terraform Attribute | Type Conversion | Null-Safe |
|---------------|---------------------|-----------------|-----------|
| `uuid` | `uuid`, `id` | string | getStringValue |
| `name` | `name` | string | getStringValue |
| `childType` | `child_type` | string | getStringValue |
| `baseType` | `base_type` | string | getStringValue |
| `addServices` | `add_services` | bool | getBoolValue |

### Null-Safe Helpers

```go
// Located in data_source_cmpart_softwareimages.go
func getStringValue(data map[string]interface{}, key string) types.String
func getBoolValue(data map[string]interface{}, key string) types.Bool
```

**Reuse Strategy**: Use existing helpers for consistent null handling across data sources.

## Implementation Notes

### Deduplication Algorithm

```go
// Step 1: Query all nodes
result, err := client.CallJSONRPC(ctx, "cmdevice", "getNodes")

// Step 2: Parse nodes
var nodes []map[string]interface{}
json.Unmarshal(result, &nodes)

// Step 3: Extract and deduplicate roles
roleMap := make(map[string]map[string]interface{})
for _, node := range nodes {
    if rolesData, ok := node["roles"].([]interface{}); ok {
        for _, roleData := range rolesData {
            if role, ok := roleData.(map[string]interface{}); ok {
                uuid := getStringValue(role, "uuid").ValueString()
                if uuid != "" {
                    roleMap[uuid] = role
                }
            }
        }
    }
}

// Step 4: Convert to slice and apply filters
roles := make([]RoleModel, 0, len(roleMap))
for _, roleData := range roleMap {
    role := mapRoleToModel(roleData)
    if matchesRoleFilter(role, namePattern, childType) {
        roles = append(roles, role)
    }
}
```

### Performance Characteristics

**Time Complexity**:
- API call: O(1) - single request
- Role extraction: O(n × r) where n=nodes, r=avg roles per node
- Deduplication: O(u) where u=unique roles
- Filtering: O(u) - linear scan of unique roles
- **Total**: O(n × r) dominated by extraction

**Space Complexity**:
- Node data: O(n) - all nodes in memory
- Role map: O(u) - unique roles only
- **Total**: O(n + u) ≈ O(n) for typical clusters

**Typical Clusters**:
- 100 nodes × 3 roles/node = 300 role references → ~10 unique roles
- Extraction + deduplication + filtering: <1 second

**Large Clusters**:
- 1000 nodes × 5 roles/node = 5000 references → ~50 unique roles
- Processing time: 2-5 seconds (acceptable for data source)

### Error Handling

**API Failures**:
```go
if err != nil {
    resp.Diagnostics.AddError(
        "Error Reading Roles",
        fmt.Sprintf("Could not query nodes: %s", err.Error()),
    )
    return
}
```

**Invalid Filter Patterns**:
```go
matched, err := filepath.Match(pattern, roleName)
if err != nil {
    resp.Diagnostics.AddError(
        "Invalid name_pattern",
        fmt.Sprintf("Glob pattern syntax error: %s", err.Error()),
    )
    return
}
```

**Empty Results**:
```go
// Return empty list (not an error)
data.Roles = []RoleModel{}
```

## Validation Rules

**No validation required** (data source is read-only):
- All attributes are computed (user cannot provide invalid values)
- Filter patterns validated at runtime with clear error messages
- Empty result sets are valid (no roles match filter)

## Testing Strategy

### Test Cases

1. **All Roles**: Query without filters, verify all unique roles returned
2. **Filter by Type**: Query with `child_type`, verify exact match
3. **Filter by Pattern**: Query with `name_pattern`, verify glob matching
4. **Combined Filters**: Query with both filters, verify AND logic
5. **Empty Results**: Query with no matches, verify empty list returned
6. **Null Fields**: Verify null-safe handling for missing role attributes

### Environment Portability

- No hardcoded role names or counts
- Works on any BCM cluster configuration
- Tests verify structure, not specific role data
- Use `knownvalue.NotNull()` for presence checks only

## References

- BCM API Documentation: `sampleRest/CMDevice_Complete_Documentation.md`
- Reference Data Source: `internal/provider/data_source_cmdevice_categories.go`
- Filter Pattern: `internal/provider/data_source_cmpart_softwareimages.go`
- Null-Safe Helpers: `data_source_cmpart_softwareimages.go:463-493`
