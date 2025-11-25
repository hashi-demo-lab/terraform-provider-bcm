# API Contract: bcm_cmdevice_roles Data Source

**Feature**: BCM Device Roles Data Source
**Type**: Terraform Data Source (read-only)
**Provider**: bcm
**Date**: 2025-11-25

## Terraform Schema

### Resource Name
`data.bcm_cmdevice_roles`

### Schema Definition

```hcl
data "bcm_cmdevice_roles" "example" {
  # Optional filter attributes
  name_pattern = string  # Glob pattern for role name (e.g., "kube-*", "*-prod")
  child_type   = string  # Exact match for role type (e.g., "ComputeRole")

  # Computed attributes
  id = string  # Data source identifier

  roles = list(object({
    id           = string  # Role identifier (same as uuid)
    uuid         = string  # Unique role identifier
    name         = string  # Role name
    child_type   = string  # Role type (HeadNodeRole, ComputeRole, etc.)
    base_type    = string  # Always "Role"
    add_services = bool    # Whether role adds services to node (nullable)
  }))
}
```

### Attribute Details

**Optional Attributes (Filters)**:

| Attribute | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `name_pattern` | string | No | Glob pattern to filter roles by name. Supports wildcards: `*` (zero or more chars), `?` (single char), `[abc]` (character class). Case-sensitive. | `"kube-*"`, `"*-prod"`, `"node-?"` |
| `child_type` | string | No | Exact match filter for role type. Case-sensitive. Must match BCM childType exactly. | `"HeadNodeRole"`, `"ComputeRole"` |

**Computed Attributes**:

| Attribute | Type | Description | Source |
|-----------|------|-------------|--------|
| `id` | string | Data source identifier. Static value `"roles"` for this data source. | Computed |
| `roles` | list(object) | List of roles matching filter criteria. Empty list if no matches. | Aggregated from all nodes |

**Role Object Attributes**:

| Attribute | Type | Nullable | Description | BCM Field |
|-----------|------|----------|-------------|-----------|
| `id` | string | No | Role identifier (same as uuid for consistency) | `uuid` |
| `uuid` | string | No | Unique role identifier | `uuid` |
| `name` | string | No | Role name (e.g., "headnode", "compute") | `name` |
| `child_type` | string | No | Role type (e.g., "HeadNodeRole") | `childType` |
| `base_type` | string | No | Always "Role" | `baseType` |
| `add_services` | bool | Yes | Whether role adds services to node. Null if not specified by BCM. | `addServices` |

## BCM API Contract

### API Method

**Service**: `cmdevice`
**Method**: `getNodes`
**Pattern**: List all nodes, extract embedded roles, deduplicate

### Request

```json
{
  "service": "cmdevice",
  "call": "getNodes"
}
```

**Parameters**: None (retrieves all nodes)

### Response

```json
[
  {
    "uuid": "c8b3a3e1-5c77-4da2-bd8d-8127a5ea44b8",
    "name": "node01",
    "hostname": "node01.example.com",
    "ipAddress": "192.168.1.10",
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
      },
      {
        "baseType": "Role",
        "childType": "MonitoringRole",
        "name": "monitoring",
        "uuid": "b12c8910-4a66-4f80-ae3d-9238b6fb55c9",
        "addServices": false,
        "modified": false,
        "to_be_removed": false,
        "revision": ""
      }
    ],
    "category": "default",
    "softwareImage": "default-image"
  },
  {
    "uuid": "d9c4b4f2-6d88-5eb3-ce9e-9238b6fb55c9",
    "name": "node02",
    "hostname": "node02.example.com",
    "ipAddress": "192.168.1.11",
    "roles": [
      {
        "baseType": "Role",
        "childType": "ComputeRole",
        "name": "compute",
        "uuid": "c569a821-5b77-5fa4-df0f-a349c7fc66d0",
        "addServices": false,
        "modified": false,
        "to_be_removed": false,
        "revision": ""
      }
    ],
    "category": "compute",
    "softwareImage": "compute-image"
  }
]
```

### Role Extraction Logic

1. **Query Nodes**: Call `cmdevice.getNodes` to retrieve all nodes
2. **Extract Roles**: Iterate through each node's `roles` array
3. **Deduplicate**: Use map[uuid]Role to store unique roles only
4. **Filter**: Apply client-side filters (name_pattern glob, child_type exact match)
5. **Return**: Convert filtered map to sorted slice for Terraform state

### Field Mapping

| BCM API Field | Terraform Attribute | Transformation | Null Handling |
|---------------|---------------------|----------------|---------------|
| `uuid` | `uuid`, `id` | Direct copy | getStringValue |
| `name` | `name` | Direct copy | getStringValue |
| `childType` | `child_type` | Snake case | getStringValue |
| `baseType` | `base_type` | Snake case | getStringValue |
| `addServices` | `add_services` | Snake case | getBoolValue (nullable) |

## Usage Examples

### Example 1: Query All Roles

```hcl
# Discover all available roles in BCM
data "bcm_cmdevice_roles" "all" {}

# Output role names for reference
output "available_roles" {
  value = [for role in data.bcm_cmdevice_roles.all.roles : role.name]
}

# Output: ["headnode", "compute", "storage", "monitoring"]
```

### Example 2: Filter by Role Type

```hcl
# Find all compute roles
data "bcm_cmdevice_roles" "compute_roles" {
  child_type = "ComputeRole"
}

# Verify at least one compute role exists
locals {
  has_compute_role = length(data.bcm_cmdevice_roles.compute_roles.roles) > 0
}

# Use first compute role UUID for device assignment
locals {
  compute_role_uuid = length(data.bcm_cmdevice_roles.compute_roles.roles) > 0 ? data.bcm_cmdevice_roles.compute_roles.roles[0].uuid : null
}
```

### Example 3: Filter by Name Pattern

```hcl
# Find all Kubernetes-related roles (kube-master, kube-worker, etc.)
data "bcm_cmdevice_roles" "kube_roles" {
  name_pattern = "kube-*"
}

# Create a map of role names to UUIDs
locals {
  kube_role_map = {
    for role in data.bcm_cmdevice_roles.kube_roles.roles :
    role.name => role.uuid
  }
}

# Output: { "kube-master" = "uuid-123", "kube-worker" = "uuid-456" }
```

### Example 4: Combined Filters

```hcl
# Find compute roles matching naming pattern
data "bcm_cmdevice_roles" "custom_compute" {
  name_pattern = "custom-*"
  child_type   = "ComputeRole"
}

# Use in device resource
resource "bcm_cmdevice_device" "compute_node" {
  name     = "compute01"
  category = "default"

  roles = [
    for role in data.bcm_cmdevice_roles.custom_compute.roles : role.uuid
  ]
}
```

### Example 5: Validate Role Existence

```hcl
# Check if specific role type exists before using it
data "bcm_cmdevice_roles" "required_roles" {
  child_type = "HeadNodeRole"
}

# Fail fast if role doesn't exist
locals {
  headnode_role_exists = length(data.bcm_cmdevice_roles.required_roles.roles) > 0
}

# Use in validation
resource "terraform_data" "validate_roles" {
  lifecycle {
    precondition {
      condition     = local.headnode_role_exists
      error_message = "HeadNodeRole not found in BCM cluster. Please create headnode role first."
    }
  }
}
```

## Error Handling

### API Errors

**Scenario**: BCM API call fails (network, authentication, server error)

```hcl
# User sees diagnostic error
Error: Error Reading Roles
Could not query nodes: API call failed: connection refused
```

**Implementation**:
```go
result, err := client.CallJSONRPC(ctx, "cmdevice", "getNodes")
if err != nil {
    resp.Diagnostics.AddError(
        "Error Reading Roles",
        fmt.Sprintf("Could not query nodes: %s", err.Error()),
    )
    return
}
```

### Invalid Filter Pattern

**Scenario**: User provides malformed glob pattern

```hcl
data "bcm_cmdevice_roles" "invalid" {
  name_pattern = "invalid["  # Unclosed bracket
}

# User sees diagnostic error
Error: Invalid name_pattern
Glob pattern syntax error: syntax error in pattern
```

**Implementation**:
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

### Empty Results

**Scenario**: No roles match filter criteria

```hcl
data "bcm_cmdevice_roles" "no_match" {
  child_type = "NonExistentRole"
}

# Output: roles = []
# This is NOT an error - empty list is valid
```

**Implementation**:
```go
// Empty list is valid response
data.Roles = []RoleModel{}
data.ID = types.StringValue("roles")
resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
```

## Filter Behavior

### AND Logic

Multiple filters use AND logic - all specified filters must match:

```hcl
data "bcm_cmdevice_roles" "filtered" {
  name_pattern = "kube-*"
  child_type   = "ComputeRole"
}

# Returns: roles where (name matches "kube-*") AND (childType == "ComputeRole")
# Does NOT return: roles matching only one condition
```

### Null/Omitted Filters

Omitted filters are ignored (do not restrict results):

```hcl
# No filters - return all roles
data "bcm_cmdevice_roles" "all" {}

# Only name_pattern - ignore child_type
data "bcm_cmdevice_roles" "pattern_only" {
  name_pattern = "kube-*"
}

# Only child_type - ignore name_pattern
data "bcm_cmdevice_roles" "type_only" {
  child_type = "ComputeRole"
}
```

### Pattern Matching

Glob patterns support standard shell wildcards:

| Pattern | Matches | Does Not Match |
|---------|---------|----------------|
| `*` | All roles | None (matches everything) |
| `kube-*` | "kube-master", "kube-worker" | "master-kube", "compute" |
| `*-prod` | "api-prod", "web-prod" | "prod-api", "staging" |
| `node-?` | "node-1", "node-a" | "node-10", "node" |
| `[abc]*` | "alpha", "beta", "cluster" | "delta", "echo" |

**Case Sensitivity**: Patterns are case-sensitive.

## Performance Characteristics

### Time Complexity

- **API Call**: O(1) - single request to BCM
- **Role Extraction**: O(n × r) where n = nodes, r = avg roles per node
- **Deduplication**: O(u) where u = unique roles
- **Filtering**: O(u) - linear scan of unique roles
- **Total**: O(n × r + u) ≈ O(n × r) for typical clusters

### Typical Performance

**Small Cluster** (10 nodes, 3 roles/node):
- Role references: 30
- Unique roles: ~5
- Query time: <500ms

**Medium Cluster** (100 nodes, 3 roles/node):
- Role references: 300
- Unique roles: ~10
- Query time: 1-2 seconds

**Large Cluster** (1000 nodes, 5 roles/node):
- Role references: 5000
- Unique roles: ~50
- Query time: 2-5 seconds

### Memory Usage

- Node data: O(n) - all nodes stored temporarily
- Role map: O(u) - unique roles only
- Filtered results: O(u) - subset of unique roles
- **Peak Memory**: O(n + u) ≈ O(n) for typical clusters

## Testing Requirements

### Acceptance Tests

1. **TestAccCMDeviceRolesDataSource_All**: Query without filters
   - Verify `id` is computed
   - Verify `roles` list is populated
   - Verify role attributes are present (uuid, name, child_type, base_type)

2. **TestAccCMDeviceRolesDataSource_FilterByChildType**: Filter by type
   - Create test config with `child_type = "ComputeRole"`
   - Verify only ComputeRole instances returned
   - Verify other role types excluded

3. **TestAccCMDeviceRolesDataSource_FilterByNamePattern**: Filter by pattern
   - Create test config with `name_pattern = "kube-*"`
   - Verify only matching role names returned
   - Verify non-matching names excluded

4. **TestAccCMDeviceRolesDataSource_CombinedFilters**: Both filters
   - Create test config with both filters
   - Verify AND logic (both conditions must match)
   - Verify empty result if no roles match both

5. **TestAccCMDeviceRolesDataSource_EmptyResults**: No matches
   - Create test config with non-existent filter
   - Verify empty `roles` list returned
   - Verify no errors (empty result is valid)

### Environment Portability

- No hardcoded role names or UUIDs
- No assumptions about cluster configuration
- Use `knownvalue.NotNull()` for presence checks only
- Generate unique test names with timestamps

## Security Considerations

- **Read-Only**: Data source cannot modify BCM state
- **Authentication**: Requires valid BCM credentials (provider config)
- **Authorization**: BCM API enforces role-based access control
- **Data Exposure**: Role metadata is not sensitive (names, types, UUIDs)

## Compatibility

- **Terraform**: >= 1.0
- **Provider Framework**: >= 1.16.1
- **BCM API**: All versions supporting `cmdevice.getNodes`
- **Go**: >= 1.24

## References

- **Spec**: `/workspace/specs/001-cmdevice-roles/spec.md`
- **Research**: `/workspace/specs/001-cmdevice-roles/research.md`
- **Data Model**: `/workspace/specs/001-cmdevice-roles/data-model.md`
- **Reference Implementation**: `internal/provider/data_source_cmdevice_categories.go`
- **Filter Pattern**: `internal/provider/data_source_cmpart_softwareimages.go`
