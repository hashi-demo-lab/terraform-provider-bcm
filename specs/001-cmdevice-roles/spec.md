# Feature Specification: BCM Device Roles Data Source

**Feature Branch**: `001-cmdevice-roles`
**Created**: 2025-11-25
**Status**: Draft
**Input**: User description: "Implement a Terraform data source (data.bcm_cmdevice_roles) to query available role types in BCM for device role assignment."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Query All Available Roles (Priority: P1)

DevOps engineers need to discover what role types are available in BCM before assigning them to devices. This enables dynamic infrastructure automation without hardcoding role names.

**Why this priority**: This is the foundation for role-based device management. Without the ability to query available roles, users cannot automate role assignments or validate role names.

**Independent Test**: Can be fully tested by querying the data source and verifying it returns all available role types. Delivers immediate value by enabling role discovery for automation workflows.

**Acceptance Scenarios**:

1. **Given** BCM has multiple roles configured (headnode, storage, backup, etc.), **When** user queries `data.bcm_cmdevice_roles` without filters, **Then** all available roles are returned with complete metadata (uuid, name, childType)
2. **Given** a fresh BCM installation, **When** user queries all roles, **Then** at least the default system roles (HeadNodeRole, StorageRole, etc.) are returned

---

### User Story 2 - Filter Roles by Type (Priority: P2)

Infrastructure automation scripts need to filter roles by their childType (HeadNodeRole, ComputeRole, etc.) to programmatically assign specific role types to devices.

**Why this priority**: Enables targeted role queries for automation scripts that need specific role types (e.g., "find all compute roles"). Builds on P1 by adding filtering capability.

**Independent Test**: Can be tested independently by applying childType filters and verifying only matching roles are returned. Useful for automation scripts that target specific role categories.

**Acceptance Scenarios**:

1. **Given** BCM has roles of different types, **When** user filters by `child_type = "ComputeRole"`, **Then** only compute roles are returned
2. **Given** multiple HeadNodeRole instances exist, **When** user filters by `child_type = "HeadNodeRole"`, **Then** all head node roles are returned

---

### User Story 3 - Filter Roles by Name Pattern (Priority: P3)

Advanced users need to filter roles by name patterns (wildcards) to find roles matching specific naming conventions (e.g., "kube-*", "*-prod").

**Why this priority**: Provides convenience for environments with many roles following naming conventions. Not critical for basic functionality but valuable for complex deployments.

**Independent Test**: Can be tested by creating roles with specific name patterns and verifying pattern-based filtering works correctly. Useful for large-scale environments with standardized naming.

**Acceptance Scenarios**:

1. **Given** roles named "kube-master", "kube-worker", "storage", **When** user filters by `name_pattern = "kube-*"`, **Then** only "kube-master" and "kube-worker" are returned
2. **Given** roles with various names, **When** user filters by `name_pattern = "*-prod"`, **Then** only roles ending with "-prod" are returned

---

### Edge Cases

- What happens when no roles match the filter criteria?
  - Return empty list with no errors
- What happens when BCM API is unavailable?
  - Return appropriate error message indicating connectivity/authentication issue
- What happens with invalid filter patterns?
  - Terraform validates filter syntax before API call
- How does the system handle roles with missing/null fields?
  - Use null-safe helper functions to return types.String/types.Bool with proper null handling

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST extract all available roles from BCM by querying all nodes and aggregating their role assignments
- **FR-002**: System MUST deduplicate roles by UUID to ensure each unique role appears only once in results
- **FR-003**: System MUST support optional filtering by `child_type` (exact match) to return only roles of specific types
- **FR-004**: System MUST support optional filtering by `name_pattern` (glob pattern matching) to enable pattern-based role discovery
- **FR-005**: System MUST return role attributes: id (UUID), uuid, name, child_type, base_type for each role
- **FR-006**: System MUST handle empty result sets gracefully when no roles match filter criteria
- **FR-007**: System MUST use null-safe field extraction helpers for optional role attributes
- **FR-008**: Data source MUST be read-only and not modify any BCM state

### Key Entities *(include if feature involves data)*

- **Role**: Represents a BCM role that can be assigned to devices
  - **Attributes**: uuid (unique identifier), name (role name), childType (role type like HeadNodeRole, ComputeRole), baseType (always "Role"), addServices (whether role adds services)
  - **Relationship**: Roles are embedded in Device objects, accessed via node.roles array
  - **Discovery**: Roles are discovered by querying all nodes via `cmdevice.getNodes` and extracting their roles array, then deduplicating by UUID

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can discover all available roles in BCM within 5 seconds (typical cluster with <100 nodes)
- **SC-002**: Filtered role queries return results in under 2 seconds for typical filter patterns
- **SC-003**: Data source correctly identifies and deduplicates 100% of unique roles across all nodes
- **SC-004**: Zero false positives or false negatives in filter matching (childType and name_pattern)
- **SC-005**: Users can successfully validate role names before device assignment in 95% of automation scenarios
- **SC-006**: System handles edge cases (empty results, null fields, API errors) without Terraform crashes

## Assumptions

- **ASSUME-001**: Roles in BCM are consistent across nodes (same UUID = same role definition)
- **ASSUME-002**: The `cmdevice.getNodes` API returns complete role information for all nodes
- **ASSUME-003**: Role UUIDs are globally unique within a BCM cluster
- **ASSUME-004**: The childType field is always present for roles (never null)
- **ASSUME-005**: Role name patterns follow standard glob syntax (*, ?) for filtering
- **ASSUME-006**: BCM API performance scales linearly with node count (no pagination needed for typical clusters <1000 nodes)

## API Contract

### BCM API Method

**Service**: `cmdevice`
**Method**: `getNodes`
**Pattern**: List all nodes, extract roles from each node, deduplicate by UUID

**Rationale**: Based on API exploration, BCM does not provide a dedicated `getRoles` or `listRoles` method. Roles are embedded in Device objects and must be extracted by:
1. Calling `cmdevice.getNodes` to get all nodes
2. Extracting the `roles` array from each node
3. Deduplicating roles by UUID across all nodes
4. Applying client-side filters (childType, name_pattern)

### Role Object Structure

```json
{
  "baseType": "Role",
  "childType": "HeadNodeRole",
  "name": "headnode",
  "uuid": "a458760f-3898-4870-bd7d-8127a5ea44b8",
  "addServices": true,
  "modified": false,
  "to_be_removed": false,
  "revision": ""
}
```

**Key Fields**:
- `uuid` (string, required): Unique role identifier
- `name` (string, required): Role name
- `childType` (string, required): Role type (HeadNodeRole, StorageRole, BackupRole, MonitoringRole, ProvisioningRole, BootRole, ComputeRole)
- `baseType` (string, required): Always "Role"
- `addServices` (bool, optional): Whether role adds services to the node

### Known Role Types (childType)

| childType | Description |
|-----------|-------------|
| HeadNodeRole | Cluster management node |
| StorageRole | NFS storage services |
| BackupRole | Backup services |
| MonitoringRole | Cluster monitoring agent |
| ProvisioningRole | Node provisioning services |
| BootRole | PXE boot services |
| ComputeRole | Compute workload execution |

## Terraform Schema

### Data Source: `bcm_cmdevice_roles`

**Optional Attributes (Filters)**:
- `name_pattern` (String) - Filter roles by name using glob pattern (e.g., "kube-*", "*-prod")
- `child_type` (String) - Filter roles by exact childType match (e.g., "ComputeRole", "HeadNodeRole")

**Computed Attributes**:
- `id` (String) - Terraform identifier (computed from all role UUIDs)
- `roles` (List of Object) - List of matching roles with attributes:
  - `id` (String) - Role UUID (same as uuid)
  - `uuid` (String) - Unique role identifier
  - `name` (String) - Role name
  - `child_type` (String) - Role type (HeadNodeRole, ComputeRole, etc.)
  - `base_type` (String) - Always "Role"
  - `add_services` (Bool) - Whether role adds services (optional, may be null)

## Example Usage

### Query All Roles

```hcl
# Discover all available roles in BCM
data "bcm_cmdevice_roles" "all" {}

# Use role names in device configuration
output "available_roles" {
  value = data.bcm_cmdevice_roles.all.roles[*].name
}
```

### Filter by Role Type

```hcl
# Find all compute roles
data "bcm_cmdevice_roles" "compute_roles" {
  child_type = "ComputeRole"
}

# Verify compute role exists before assignment
locals {
  compute_role_exists = length(data.bcm_cmdevice_roles.compute_roles.roles) > 0
}
```

### Filter by Name Pattern

```hcl
# Find Kubernetes-related roles
data "bcm_cmdevice_roles" "kube_roles" {
  name_pattern = "kube-*"
}

# Use first matching role
locals {
  kube_master_role = data.bcm_cmdevice_roles.kube_roles.roles[0].uuid
}
```

## Implementation Notes

### Data Extraction Strategy

1. Call `cmdevice.getNodes` to retrieve all nodes
2. Iterate through each node and extract its `roles` array
3. Build a map[uuid]Role to deduplicate roles across nodes
4. Apply filters (childType, name_pattern) in Go code
5. Convert filtered map to list for Terraform state

### Client-Side Filtering

- **childType**: Exact string match comparison
- **name_pattern**: Use Go's `filepath.Match` for glob pattern matching

### Null Safety

Use helper functions from existing data sources:
- `getStringValue(data, "name")` - Returns types.String with null handling
- `getBoolValue(data, "addServices")` - Returns types.Bool with null handling

### Performance Considerations

- For clusters with <100 nodes: ~1-2 second query time
- For clusters with 100-1000 nodes: ~2-5 second query time
- Role deduplication happens in-memory (minimal overhead)
- Client-side filtering adds negligible latency

## Testing Strategy

### Acceptance Tests

1. **TestAccCMDeviceRolesDataSource_All**: Query all roles without filters, verify known role types present
2. **TestAccCMDeviceRolesDataSource_FilterByChildType**: Filter by childType, verify only matching roles returned
3. **TestAccCMDeviceRolesDataSource_FilterByNamePattern**: Filter by name pattern, verify pattern matching works
4. **TestAccCMDeviceRolesDataSource_EmptyResults**: Apply filter with no matches, verify empty list returned
5. **TestAccCMDeviceRolesDataSource_CombinedFilters**: Apply both filters simultaneously, verify AND logic

### Test Data Requirements

- Use existing BCM cluster roles (no test data creation needed)
- Verify at least default system roles exist (HeadNodeRole, StorageRole, etc.)
- Test with environment-portable assertions (no hardcoded role counts or names)

## Dependencies

- **API Method**: `cmdevice.getNodes` (existing, verified working)
- **Reference Implementations**:
  - `internal/provider/data_source_cmdevice_categories.go` - Data source pattern
  - `internal/provider/data_source_cmpart_softwareimages.go` - Filtering pattern
  - `internal/provider/test_helpers.go` - Test utilities
