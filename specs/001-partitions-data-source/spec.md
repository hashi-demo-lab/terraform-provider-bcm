# Feature Specification: BCM Partitions Data Source

**Feature Branch**: `001-partitions-data-source`
**Created**: 2025-11-22
**Status**: Draft
**Input**: User description: "Implement a new data source bcm_cmpart_partitions that retrieves partition information from the BCM API using the JSON-RPC call to cmpart service getPartitions method. Partitions are filesystem partitions referenced by software images via bootfs_part and fs_part UUID fields."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Query All Available Partitions (Priority: P1)

As a Terraform user managing BCM infrastructure, I need to retrieve a list of all filesystem partitions available in the cluster so that I can reference them when configuring software images or understanding cluster storage configuration.

**Why this priority**: This is the foundational capability - without the ability to list partitions, users cannot discover what partitions exist to reference in their software image configurations. This is the MVP.

**Independent Test**: Can be fully tested by declaring a data source with no filters and verifying it returns all partitions from the BCM API. Delivers immediate value by enabling partition discovery.

**Acceptance Scenarios**:

1. **Given** a BCM cluster with configured partitions, **When** I declare `data "bcm_cmpart_partitions" "all" {}` in my Terraform configuration, **Then** the data source retrieves all partitions and exposes their attributes (UUID, name, path, size, type, etc.)
2. **Given** the data source has retrieved partitions, **When** I reference `data.bcm_cmpart_partitions.all.partitions`, **Then** I receive a list of partition objects with all computed attributes populated
3. **Given** a BCM cluster with no partitions, **When** I read the data source, **Then** it returns an empty list without error

---

### User Story 2 - Filter Partitions by Name Pattern (Priority: P2)

As a Terraform user, I need to filter partitions by name pattern so that I can quickly find specific partitions without retrieving the entire list, making my configurations more efficient and readable.

**Why this priority**: Filtering improves usability in production environments with many partitions. While not essential for MVP, it significantly improves user experience and configuration maintainability.

**Independent Test**: Can be tested by creating a filter block with a name pattern and verifying only matching partitions are returned. Delivers value for users managing large partition sets.

**Acceptance Scenarios**:

1. **Given** partitions named "boot-partition", "root-partition", and "data-partition" exist, **When** I specify `filter { name_pattern = "boot" }`, **Then** only partitions containing "boot" in their name are returned
2. **Given** I specify a name pattern that matches no partitions, **When** the data source executes, **Then** it returns an empty list without error
3. **Given** I specify `name_pattern = ""` (empty string), **When** the filter executes, **Then** all partitions are returned (empty pattern matches everything)

---

### User Story 3 - Reference Partition UUIDs in Software Image Resources (Priority: P3)

As a Terraform user defining software image resources, I need to look up partition UUIDs by name so that I can reference the correct partitions without hardcoding UUID values, making my configurations more maintainable.

**Why this priority**: This builds on P1 and P2 to enable practical workflows. While important, users can manually look up UUIDs if needed, so this is an enhancement rather than core functionality.

**Independent Test**: Can be tested by using the data source output to populate a software image resource's bootfspart or fspart fields, then verifying the software image is created with the correct partition references.

**Acceptance Scenarios**:

1. **Given** I query partitions and filter by name, **When** I reference `data.bcm_cmpart_partitions.boot.partitions[0].uuid` in a software image resource, **Then** the software image is created with the correct partition UUID
2. **Given** multiple partitions match my filter, **When** I attempt to reference `partitions[0].uuid`, **Then** I get the first matching partition's UUID consistently
3. **Given** I need both boot and root partitions, **When** I create two data sources with different filters, **Then** I can reference both partition UUIDs in my software image configuration

---

### Edge Cases

- What happens when the BCM API is unavailable or returns an error during partition retrieval?
- How does the system handle partitions with missing or null field values?
- What occurs when a partition name contains special characters or very long strings?
- How does client-side filtering behave with case sensitivity for partition names?
- What happens if a user references a partition UUID that no longer exists between plan and apply?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Data source MUST call BCM JSON-RPC API with service "cmpart" and call "getPartitions" to retrieve partition list
- **FR-002**: Data source MUST expose all partition attributes as computed fields including UUID, name, path, size, filesystem type, and metadata
- **FR-003**: Data source MUST support client-side filtering by name pattern using case-insensitive substring matching
- **FR-004**: Data source MUST handle null or missing API response fields gracefully using null-safe helper functions
- **FR-005**: Data source MUST follow the established project pattern for data sources (schema definition, API call, JSON unmarshaling, data mapping, client-side filtering, state management)
- **FR-006**: Data source MUST use terraform-plugin-framework types (types.String, types.Int64, types.Bool) for all attributes
- **FR-007**: Data source MUST set a computed ID attribute for Terraform state tracking
- **FR-008**: Data source MUST return an empty list when no partitions exist or no partitions match the filter, not an error
- **FR-009**: Acceptance tests MUST use modern terraform-plugin-testing patterns with statecheck.ExpectKnownValue() and type-safe matchers
- **FR-010**: Acceptance tests MUST be environment-portable with no hardcoded partition counts or names
- **FR-011**: Implementation MUST include three files: data source implementation, acceptance tests, and example Terraform configuration
- **FR-012**: Documentation MUST be auto-generated using tfplugindocs via `make generate` command

### API Contract

**API Request**:
```json
{
  "service": "cmpart",
  "call": "getPartitions"
}
```

**Expected API Response Structure** (to be validated in Phase 0):
```json
[
  {
    "uuid": "f3652da2-2efe-414b-a650-10c8feea5d8f",
    "name": "boot-partition-name",
    "path": "/dev/sda1",
    "size": 1073741824,
    "fsType": "ext4",
    "mountPoint": "/boot",
    "baseType": "Partition",
    "childType": "",
    "creationTime": 1763006515,
    "modified": false,
    "to_be_removed": false,
    "revision": "",
    "notes": "Boot partition for cluster nodes"
  }
]
```

Note: Exact field names and types will be validated during Phase 0 (API exploration). The above structure is based on patterns observed in other BCM API responses.

### Key Entities

- **Partition**: A filesystem partition in the BCM cluster that can be referenced by software images. Key attributes include unique identifier (UUID), human-readable name, filesystem path, size in bytes, filesystem type (ext4, xfs, etc.), mount point, and creation metadata. Partitions are referenced by software images via UUID in the bootfs_part and fs_part fields.

- **Filter**: An optional configuration block for client-side filtering. Contains name_pattern (case-insensitive substring match). Multiple filter criteria use AND logic (all must match).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can retrieve all partitions from BCM API in a single Terraform data source declaration
- **SC-002**: Filtered queries return results in under 5 seconds for clusters with up to 100 partitions
- **SC-003**: Data source handles empty partition lists and null field values without errors
- **SC-004**: All acceptance tests pass using modern testing patterns with zero hardcoded environment assumptions
- **SC-005**: Generated documentation clearly explains partition attributes and filter usage with working examples
- **SC-006**: Users can successfully reference partition UUIDs from the data source output in software image resource configurations

## Implementation Details *(non-normative)*

### File Structure

```
terraform-provider-bcm/
├── internal/provider/
│   ├── data_source_cmpart_partitions.go          # Data source implementation
│   └── data_source_cmpart_partitions_test.go     # Acceptance tests
├── examples/
│   └── data-sources/
│       └── bcm_cmpart_partitions/
│           └── data-source.tf                     # Example configuration
└── docs/
    └── data-sources/
        └── cmpart_partitions.md                   # Auto-generated docs
```

### Schema Design

**Data Source**: `bcm_cmpart_partitions`

**Attributes**:
- `id` (string, computed): Placeholder identifier for Terraform state
- `filter` (block, optional): Client-side filtering configuration
  - `name_pattern` (string, optional): Case-insensitive substring match for partition name
- `partitions` (list of objects, computed): List of partition objects with attributes:
  - `id` (string, computed): Partition identifier (same as uuid)
  - `uuid` (string, computed): Partition UUID (primary identifier)
  - `name` (string, computed): Human-readable partition name
  - `path` (string, computed): Filesystem device path (e.g., /dev/sda1)
  - `size` (int64, computed): Partition size in bytes
  - `fs_type` (string, computed): Filesystem type (ext4, xfs, btrfs, etc.)
  - `mount_point` (string, computed): Mount point path
  - `base_type` (string, computed): BCM base type (typically "Partition")
  - `child_type` (string, computed): BCM polymorphic type discriminator
  - `creation_time` (int64, computed): Unix timestamp of partition creation
  - `revision` (string, computed): BCM revision tracking field
  - `modified` (bool, computed): Whether partition has uncommitted modifications
  - `to_be_removed` (bool, computed): Whether partition is marked for deletion
  - `notes` (string, computed): User notes or description

Note: Field names use Terraform snake_case convention. BCM API uses camelCase (e.g., "creationTime" becomes "creation_time").

### Test Coverage Requirements

1. **Basic Test**: Retrieve all partitions without filters, verify data source ID and partition count > 0 (environment portable)
2. **Filter Test - Name Pattern**: Filter by name substring, verify only matching partitions returned
3. **Filter Test - No Matches**: Filter with pattern matching no partitions, verify empty list returned
4. **Computed Fields Test**: Verify all partition attributes are populated with correct types using statecheck.ExpectKnownValue()
5. **Empty Cluster Test**: Handle BCM cluster with no partitions, verify empty list without error

All tests must:
- Use modern terraform-plugin-testing patterns (statecheck, knownvalue, tfjsonpath)
- Be environment-portable (no hardcoded counts, names, or UUIDs)
- Include provider configuration with BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD environment variables
- Use generateUniqueTestName() for any created test resources
- Follow the three-file pattern: implementation, tests, examples

### Example Terraform Configuration

```hcl
# Query all partitions
data "bcm_cmpart_partitions" "all" {}

# Filter partitions by name pattern
data "bcm_cmpart_partitions" "boot_partitions" {
  filter {
    name_pattern = "boot"
  }
}

# Reference partition UUID in software image resource
resource "bcm_cmpart_softwareimage" "example" {
  name             = "my-image"
  bootfspart       = data.bcm_cmpart_partitions.boot_partitions.partitions[0].uuid
  # ... other configuration
}

# Output partition information
output "partition_names" {
  value = [for p in data.bcm_cmpart_partitions.all.partitions : p.name]
}

output "boot_partition_uuid" {
  value = data.bcm_cmpart_partitions.boot_partitions.partitions[0].uuid
}
```

## Assumptions

1. **API Response Format**: The getPartitions API call returns a JSON array of partition objects, consistent with other BCM API endpoints (getSoftwareImages, getNetworks). This will be validated in Phase 0 (API exploration).

2. **Field Naming Convention**: BCM API uses camelCase field names which are converted to snake_case for Terraform compatibility (e.g., "creationTime" to "creation_time"), consistent with existing data sources.

3. **UUID Uniqueness**: Partition UUIDs are unique and stable identifiers that do not change across API calls, making them suitable for cross-references in Terraform configurations.

4. **Null Handling**: BCM API may return null values for optional fields. The implementation will use null-safe helper functions (getStringValue, getBoolValue, getInt64Value) as established in data_source_cmpart_softwareimages.go.

5. **Filter Logic**: Multiple filter criteria (if extended in the future) use AND logic, consistent with the pattern in existing data sources.

6. **Client-Side Filtering**: All filtering is performed client-side in the provider after retrieving the full partition list from the API. The BCM API does not support server-side filtering for getPartitions.

7. **Case Sensitivity**: Name pattern filtering is case-insensitive to improve user experience, using strings.Contains with strings.ToLower() conversion.

8. **Test Environment**: Acceptance tests assume a BCM cluster with at least one partition exists. Tests are designed to be environment-portable by not assuming specific partition names or counts.

## Dependencies

- Terraform Plugin Framework v1.16.1
- terraform-plugin-testing v1.13.3 (for modern test patterns)
- Existing BCMClient with CallJSONRPC() method
- Existing null-safe helper functions (getStringValue, getBoolValue, getInt64Value)

## Out of Scope

- Creating or modifying partitions (this is a read-only data source)
- Server-side filtering via BCM API (all filtering is client-side)
- Partition health monitoring or status checking
- Multi-filter support beyond name pattern (can be added later if needed)
- Validation that referenced partition UUIDs still exist (Terraform handles this via lifecycle)

## TDD Workflow

This feature follows the RED-GREEN-REFACTOR cycle:

1. **RED Phase**: Write failing acceptance tests first
   - Test basic partition retrieval
   - Test filter functionality
   - Test edge cases (empty lists, null fields)

2. **GREEN Phase**: Implement minimal working code
   - Define schema with all attributes
   - Implement Read() method with API call
   - Implement client-side filtering logic
   - Map API response to Terraform state

3. **REFACTOR Phase**: Improve code quality
   - Extract common patterns to helper functions
   - Add comprehensive error handling
   - Improve logging and debugging output
   - Optimize filter performance if needed

4. **DOCUMENT Phase**: Generate documentation
   - Run `make generate` to create docs from examples
   - Verify generated documentation is clear and accurate

## Next Steps

1. **Phase 0 - API Exploration**: Run the cmpart-get-partitions.py script to validate API response structure and confirm field names/types
2. **Phase 1 - Specification Review**: Review and finalize this specification with stakeholders
3. **Phase 2 - Implementation Planning**: Use `/speckit.plan` to generate detailed implementation plan
4. **Phase 3 - Task Breakdown**: Use `/speckit.tasks` to create actionable task list
5. **Phase 4 - Implementation**: Follow TDD workflow to implement feature
6. **Phase 5 - Testing**: Run acceptance tests and validate against real BCM cluster
7. **Phase 6 - Documentation**: Generate and review documentation
