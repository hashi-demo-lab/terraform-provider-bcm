# Feature Specification: BCM CMPart Entity Info Data Source

**Feature Branch**: `095-cmpart-entity-info`
**Created**: 2025-11-27
**Status**: Draft
**Input**: User description: "Add a new data source bcm_cmpart_entity_info that queries the BCM API getBasicEntityInformation endpoint to retrieve entity metadata by type and/or name."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - List Entities by Type (Priority: P1)

As a Terraform user, I want to retrieve all BCM entities of a specific type (e.g., all SoftwareImages, all Categories) so that I can discover what resources exist in my cluster and reference them in my Terraform configurations.

**Why this priority**: This is the primary use case - users most commonly need to look up entities by their type to find UUIDs for use in other resources. This provides immediate value for cross-resource references.

**Independent Test**: Can be fully tested by querying for entities with `type = "SoftwareImage"` and verifying the returned list contains only SoftwareImage entities with valid name, type, and uuid fields.

**Acceptance Scenarios**:

1. **Given** a BCM cluster with multiple entity types (SoftwareImages, Categories, Networks), **When** I query with `type = "SoftwareImage"`, **Then** I receive a list containing only SoftwareImage entities, each with name, type, and uuid populated.

2. **Given** a BCM cluster with 8 SoftwareImage entities, **When** I query with `type = "SoftwareImage"`, **Then** I receive exactly 8 entities in the result list.

3. **Given** a BCM cluster, **When** I query with `type = "NonExistentType"`, **Then** I receive an empty list (no error).

---

### User Story 2 - Filter Entities by Name Pattern (Priority: P2)

As a Terraform user, I want to filter entities by name using glob patterns so that I can find specific entities when I know part of their name but not the exact full name.

**Why this priority**: Name filtering provides flexibility when users know entity naming conventions but need to discover exact names. This complements type filtering for more precise lookups.

**Independent Test**: Can be tested by querying with `name_pattern = "default*"` and verifying only entities whose names match the glob pattern are returned.

**Acceptance Scenarios**:

1. **Given** entities named "default-image", "default-category", and "custom-image", **When** I query with `name_pattern = "default*"`, **Then** I receive only "default-image" and "default-category".

2. **Given** entities in the cluster, **When** I query with `name_pattern = "*node*"`, **Then** I receive all entities with "node" anywhere in their name.

3. **Given** entities in the cluster, **When** I query with `name_pattern = "exact-name"`, **Then** I receive only entities with exactly that name (pattern without wildcards matches literally).

---

### User Story 3 - Combined Type and Name Filtering (Priority: P2)

As a Terraform user, I want to combine type and name filters so that I can precisely locate specific entities by both characteristics.

**Why this priority**: Combined filtering allows users to narrow down results when multiple entity types share similar naming conventions, providing precise lookups.

**Independent Test**: Can be tested by querying with both `type = "SoftwareImage"` and `name_pattern = "default*"` and verifying only SoftwareImages matching the name pattern are returned.

**Acceptance Scenarios**:

1. **Given** a "default-image" SoftwareImage and a "default" Category, **When** I query with `type = "SoftwareImage"` and `name_pattern = "default*"`, **Then** I receive only the "default-image" SoftwareImage.

2. **Given** multiple entity types, **When** I provide both filters, **Then** both filters are applied using AND logic (entities must match both criteria).

---

### User Story 4 - Retrieve All Entities (Priority: P3)

As a Terraform user, I want to retrieve all entities without any filters so that I can explore what exists in the BCM cluster for discovery and documentation purposes.

**Why this priority**: Discovery mode is useful for initial exploration but returns large result sets (500+ entities). Less common than filtered queries.

**Independent Test**: Can be tested by querying without any filters and verifying all entity types are returned.

**Acceptance Scenarios**:

1. **Given** a BCM cluster with 500+ entities across multiple types, **When** I query without type or name_pattern filters, **Then** I receive all entities in the cluster.

2. **Given** the full entity list, **When** I examine the results, **Then** each entity has name, type, and uuid fields populated (no null values for required fields).

---

### User Story 5 - Lookup Entity UUID by Known Name (Priority: P3)

As a Terraform user, I want to find the UUID of an entity when I know its exact name so that I can reference it in other Terraform resources.

**Why this priority**: Direct UUID lookup is a common workflow, enabled by combining type and name_pattern filters to return a single result.

**Independent Test**: Can be tested by querying with exact name and type, then using the returned UUID in another resource configuration.

**Acceptance Scenarios**:

1. **Given** a SoftwareImage named "default-image", **When** I query with `type = "SoftwareImage"` and `name_pattern = "default-image"`, **Then** I receive exactly one entity with the correct UUID.

2. **Given** the returned entity, **When** I reference its uuid in another Terraform resource, **Then** the reference resolves correctly.

---

### Edge Cases

- What happens when the BCM API returns an empty response? The data source returns an empty entities list.
- What happens when the API returns 500+ entities without filters? All entities are returned; users should apply filters for performance.
- How does the system handle special characters in name patterns? Glob patterns support `*` (any characters) and `?` (single character); literal special characters are matched as-is.
- What happens when an entity has no resolveName in the API response? The entity is included with an empty string for name.
- How are duplicate entity names handled? Duplicates are allowed - the API may return multiple entities with the same name but different UUIDs or types.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Data source MUST query the BCM API endpoint `cmpart.getBasicEntityInformation` to retrieve entity metadata.
- **FR-002**: Data source MUST support optional filtering by entity `type` (case-sensitive exact match against the API's `type` field).
- **FR-003**: Data source MUST support optional filtering by `name_pattern` using glob syntax (`*` for any characters, `?` for single character).
- **FR-004**: Data source MUST return a list of entities, where each entity contains `name`, `type`, and `uuid` fields.
- **FR-005**: The `name` field MUST be mapped from the API's `resolveName` field for clarity in Terraform configurations.
- **FR-006**: Data source MUST apply filters client-side after retrieving the full entity list from the API.
- **FR-007**: When both `type` and `name_pattern` filters are provided, data source MUST apply AND logic (entities must match both criteria).
- **FR-008**: When no filters are provided, data source MUST return all entities from the API response.
- **FR-009**: Data source MUST set a unique `id` attribute for state management (using a combination of filter values or a static placeholder).
- **FR-010**: Data source MUST handle API errors gracefully with descriptive error messages including endpoint and error details.
- **FR-011**: The `name_pattern` filter MUST be case-insensitive to match user expectations for glob patterns.

### Key Entities

- **Entity Info**: Represents basic metadata about a BCM entity. Contains:
  - `name` (string): The human-readable entity name (mapped from API's `resolveName`)
  - `type` (string): The entity type classification (e.g., "SoftwareImage", "Category", "HeadNode")
  - `uuid` (string): The unique identifier used to reference the entity in other API calls

- **Entity Types**: BCM supports 60+ entity types including but not limited to:
  - Category, SoftwareImage, Network, PhysicalNode, HeadNode
  - Partition, Rack, ConfigurationOverlay, FSPart, Profile
  - SlurmWlmCluster, MonitoringMeasurableMetric, PrometheusQuery

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can retrieve entity UUIDs by type in a single Terraform plan/apply cycle without manual API queries.
- **SC-002**: Data source queries complete within acceptable time limits even with 500+ entities (under 10 seconds for full retrieval).
- **SC-003**: All acceptance tests pass for type filtering, name pattern filtering, and combined filtering scenarios.
- **SC-004**: Documentation is generated and accurately describes the data source schema, filters, and example usage.
- **SC-005**: Data source correctly handles the API response structure with zero data loss or corruption during field mapping.
- **SC-006**: 100% of returned entities have non-null uuid and type fields (name may be empty string if API returns empty resolveName).

## Assumptions

- The BCM API `cmpart.getBasicEntityInformation` endpoint is stable and returns consistent response structure.
- The API does not support server-side filtering, requiring client-side filtering implementation.
- Entity types are case-sensitive as returned by the API (e.g., "SoftwareImage" not "softwareimage").
- The glob pattern implementation uses standard glob semantics (`*` matches any characters including none, `?` matches exactly one character).
- Performance is acceptable for clusters with up to 1000 entities; larger clusters may benefit from type filtering.

## Out of Scope

- Server-side pagination (API returns all entities in single response)
- Caching of entity information between Terraform runs
- Write operations (this is a read-only data source)
- Filtering by UUID (users with UUIDs can use direct resource lookups)
- Regular expression support for name filtering (glob patterns only)
