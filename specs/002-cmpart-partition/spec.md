# Feature Specification: BCM Partition Resource

**Feature Branch**: `002-cmpart-partition`
**Created**: 2025-11-23
**Status**: Draft
**Input**: User description: "Create a CORRECTED feature specification for implementing the bcm_cmpart_partition resource based on the ACTUAL BCM API structure for managing BCM cluster partition (organizational unit) lifecycle via Terraform"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create and Manage Cluster Partition (Priority: P1)

A cluster administrator wants to create a new BCM partition (organizational unit) to logically separate different groups of nodes within their HPC cluster. They need to specify essential configuration like cluster name, node naming conventions, and network settings through Terraform infrastructure-as-code.

**Why this priority**: This is the foundational use case - without the ability to create and manage partitions, the resource provides no value. Partitions are organizational units that group nodes and resources in BCM clusters.

**Independent Test**: Can be fully tested by creating a Terraform configuration with `resource "bcm_cmpart_partition" "test"` specifying name and cluster_name, applying it, and verifying the partition exists in BCM via API or data source lookup. Delivers immediate value by enabling partition lifecycle management.

**Acceptance Scenarios**:

1. **Given** no partition named "engineering" exists, **When** administrator applies Terraform config with `name = "engineering"` and `cluster_name = "hpc-prod"`, **Then** partition is created in BCM with specified attributes
2. **Given** partition "engineering" exists with `slave_name = "node"`, **When** administrator updates config to `slave_name = "compute"` and applies, **Then** BCM partition is updated with new node naming prefix
3. **Given** partition "engineering" exists, **When** administrator removes the resource from Terraform config and applies, **Then** partition is deleted from BCM
4. **Given** partition managed by Terraform exists in BCM, **When** administrator runs `terraform import bcm_cmpart_partition.test <uuid>`, **Then** partition state is imported into Terraform

---

### User Story 2 - Configure Network and Email Settings (Priority: P2)

A cluster administrator wants to configure network infrastructure settings (DNS servers, NTP time servers, search domains) and administrative contact emails for a partition to ensure proper cluster communication and notifications.

**Why this priority**: Network configuration is essential for cluster operations but can be configured after initial partition creation. This represents an enhancement to the basic partition management.

**Independent Test**: Can be tested by creating a partition with optional network fields (`time_servers = ["ntp1.example.com"]`, `name_servers = ["8.8.8.8"]`, `admin_email = ["admin@example.com"]`) and verifying these values are set in BCM and reflected in Terraform state.

**Acceptance Scenarios**:

1. **Given** partition exists with no time servers configured, **When** administrator adds `time_servers = ["ntp1.corp.com", "ntp2.corp.com"]` and applies, **Then** partition is updated with NTP server configuration
2. **Given** partition exists with one admin email, **When** administrator updates `admin_email` list to include multiple addresses and applies, **Then** all email addresses are stored in partition configuration
3. **Given** partition exists with custom DNS configuration, **When** administrator removes `name_servers` attribute from config and applies, **Then** partition reverts to default DNS settings

---

### User Story 3 - Detect and Correct Configuration Drift (Priority: P2)

A cluster administrator wants Terraform to detect when partition configuration has been modified outside of Terraform (via BCM GUI or API) and restore the desired state defined in their Terraform configuration.

**Why this priority**: Drift detection ensures configuration compliance and prevents out-of-band changes from causing state inconsistencies. Critical for production environments with multiple administrators.

**Independent Test**: Can be tested by creating a partition via Terraform, manually modifying it via BCM API (e.g., changing notes field), running `terraform plan`, and verifying Terraform detects the drift and proposes corrective changes.

**Acceptance Scenarios**:

1. **Given** partition managed by Terraform with `notes = "Engineering cluster"`, **When** administrator modifies notes directly in BCM to "Modified externally" and runs `terraform plan`, **Then** Terraform detects drift and shows plan to restore original notes
2. **Given** partition with `cluster_name = "prod-hpc"` in Terraform, **When** cluster_name is changed via BCM API and `terraform apply` runs, **Then** Terraform restores the cluster_name to "prod-hpc"

---

### User Story 4 - Advanced Node Naming Configuration (Priority: P3)

A cluster administrator wants to configure detailed node naming conventions including prefix (`slave_name`) and zero-padding digit count (`slave_digits`) to match organizational naming standards.

**Why this priority**: Node naming is important for consistency but is optional configuration that can default to reasonable values. Lower priority than core CRUD operations.

**Independent Test**: Can be tested by creating a partition with `slave_name = "gpu"` and `slave_digits = 4` and verifying nodes would be named gpu0001, gpu0002, etc. according to BCM's naming convention.

**Acceptance Scenarios**:

1. **Given** new partition being created, **When** administrator specifies `slave_name = "compute"` and `slave_digits = 3`, **Then** partition is configured for node names like compute001, compute002
2. **Given** partition with default slave_digits, **When** administrator updates to `slave_digits = 4` and applies, **Then** future node naming uses 4-digit zero-padding

---

### Edge Cases

- What happens when creating a partition with a name that already exists in BCM? (Should fail with clear error message)
- How does the system handle attempting to delete a partition that has active nodes assigned to it? (BCM API may reject deletion - should surface error to user)
- What happens when required field `cluster_name` is empty or contains invalid characters? (Schema validation should catch before API call)
- How does the system handle concurrent updates to the same partition from multiple Terraform workspaces? (BCM API versioning via `revision` field should detect conflicts)
- What happens when importing a partition that doesn't exist by UUID? (Import should fail with "not found" error)
- How does the system handle very long lists (100+ entries) in `admin_email` or `name_servers`? (Should handle gracefully, BCM API may have limits)
- What happens when attempting to update the `name` field of an existing partition? (Terraform should force replacement as name may be a logical identifier)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST support creating BCM partitions via Terraform with required fields `name` and `cluster_name`
- **FR-002**: System MUST support updating partition configuration fields (cluster_name, slave_name, slave_digits, network settings, notes) without recreating the partition
- **FR-003**: System MUST support deleting partitions via Terraform destroy operation
- **FR-004**: System MUST support importing existing BCM partitions into Terraform state using their UUID
- **FR-005**: System MUST persist all partition configuration in BCM via JSON-RPC API calls to cmpart service
- **FR-006**: System MUST retrieve current partition state from BCM during Read operations to detect drift
- **FR-007**: System MUST validate that `name` attribute is non-empty and follows BCM naming conventions
- **FR-008**: System MUST handle optional list attributes (admin_email, time_servers, search_domains, name_servers) as empty lists when not specified
- **FR-009**: System MUST set computed fields (uuid, id, base_type, child_type, creation_time, revision, modified, to_be_removed) from BCM API responses
- **FR-010**: System MUST map Terraform snake_case attribute names to BCM API camelCase field names (e.g., `cluster_name` → `clusterName`)
- **FR-011**: System MUST use the `getPartition(uuid)` API method with args parameter for efficient direct lookup during Read operations
- **FR-012**: System MUST support concurrent partition management operations across different Terraform resources
- **FR-013**: System MUST provide clear error messages when BCM API operations fail (creation conflicts, validation errors, deletion failures)
- **FR-014**: System MUST implement acceptance tests covering all CRUD operations, import, drift detection, and idempotency verification
- **FR-015**: System MUST follow modern Terraform testing patterns using statecheck.ExpectKnownValue and plancheck.ExpectEmptyPlan

### Key Entities

- **Partition**: BCM cluster organizational unit that groups nodes and defines cluster-wide configuration
  - Identity: uuid (unique identifier), name (human-readable label)
  - Cluster Configuration: cluster_name (display name), slave_name (node prefix), slave_digits (numbering format)
  - Network Arrays: admin_email (admin contacts), time_servers (NTP servers), search_domains (DNS search), name_servers (DNS resolvers)
  - Settings: relay_host (SMTP relay), no_zero_conf (disable Zeroconf)
  - Metadata: creation_time (timestamp), revision (version), modified (dirty flag), to_be_removed (deletion flag), notes (description)

- **BCM API Entity Structure**: Wrapper for partition data in API calls
  - Contains: baseType, childType, modified, to_be_removed, revision, uuid fields
  - Plus all partition-specific fields in camelCase format

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Cluster administrators can create a new BCM partition with required fields in under 30 seconds (single Terraform apply)
- **SC-002**: Partition configuration updates (changing network settings, notes, naming conventions) complete without recreation (in-place update)
- **SC-003**: System detects configuration drift within one Terraform plan operation when partition is modified externally
- **SC-004**: Acceptance test suite achieves 100% pass rate covering Create, Read, Update, Delete, Import, and Drift Detection scenarios
- **SC-005**: Import operation successfully brings existing partition under Terraform management with no attribute mismatches
- **SC-006**: Resource handles concurrent operations (multiple partitions created/updated in parallel) without conflicts or state corruption
- **SC-007**: All error scenarios (duplicate name, invalid fields, deletion failures) provide actionable error messages to users
- **SC-008**: Documentation is auto-generated via `make generate` and includes all attributes with descriptions and examples

## Assumptions

- BCM API endpoint is accessible and authenticated via provider configuration (endpoint, username, password)
- Partition names are unique within a BCM cluster (enforced by BCM API)
- The `uuid` field is the stable identifier for partitions (used for Read, Update, Delete, Import operations)
- BCM API supports direct partition lookup via `getPartition(uuid)` method with args parameter (following pattern from resource_cmpart_softwareimage.go)
- Empty lists for array attributes (admin_email, time_servers, etc.) are valid and represent "not configured"
- The `revision` field in BCM API is used for optimistic concurrency control
- Changing the `name` attribute may require partition recreation (force replacement) as it could be used as a logical identifier elsewhere in BCM
- The `base_type` field is always "Partition" for partition entities
- The `child_type` field distinguishes partition subtypes (if any) via polymorphism
- SMTP relay configuration (`relay_host`) is optional and only needed if email notifications are required
- The `no_zero_conf` boolean defaults to false (Zeroconf enabled) when not specified
- Slave naming defaults (slave_name="node", slave_digits=3) are reasonable if not specified
- Standard CRUD pattern from bcm_cmpart_softwareimage resource applies to partition management
- Modern testing patterns (statecheck, plancheck) are used per project TDD guidelines
- Test environment uses unique names (via generateUniqueTestName) to avoid conflicts across test runs

## Dependencies

- Terraform Plugin Framework v1.16.1+
- BCM JSON-RPC API (cmpart service) with methods: addPartition, getPartition, updatePartition, removePartition
- Existing BCM client implementation (internal/provider/bcm_client.go) with CallJSONRPC method
- Reference data source implementation (internal/provider/data_source_cmpart_partitions.go) for schema structure
- Reference resource implementation (internal/provider/resource_cmpart_softwareimage.go) for CRUD patterns
- Testing infrastructure: terraform-plugin-testing v1.13.3+, test helpers (createTestBCMClient, generateUniqueTestName, etc.)
- Documentation generation tools: tfplugindocs

## Out of Scope

- Managing nodes assigned to partitions (separate bcm_cmdevice_node resource)
- Managing software images associated with partitions (separate bcm_cmpart_softwareimage resource)
- Advanced partition types or hierarchical partition structures (if BCM supports them)
- Real-time validation of network settings (DNS reachability, NTP sync) - BCM handles this
- Migration of nodes between partitions (complex operation beyond resource scope)
- Bulk operations on multiple partitions simultaneously (use count/for_each in Terraform config)
- Custom validation rules beyond BCM API requirements (rely on BCM API errors)
- Backup and restore of partition configuration (infrastructure concern, not resource concern)
