# Feature Specification: BCM CMDevice Interfaces Data Source

**Feature Branch**: `001-cmdevice-interfaces`
**Created**: 2025-11-23
**Status**: Draft
**Input**: User description: "Implement data.bcm_cmdevice_interfaces data source to list and filter network interfaces on compute nodes."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Discover All Network Interfaces (Priority: P1)

As a cluster administrator, I need to discover all network interfaces across all compute nodes in my BCM cluster so that I can understand my network topology and plan network configurations.

**Why this priority**: This is the fundamental capability - being able to list all interfaces. Without this, no other filtering or querying is possible. This delivers immediate value by providing visibility into the cluster's network configuration.

**Independent Test**: Can be fully tested by executing a Terraform plan with `data "bcm_cmdevice_interfaces" "all" {}` and verifying that it returns a list of interface objects with all required attributes populated.

**Acceptance Scenarios**:

1. **Given** a BCM cluster with multiple compute nodes, **When** I query all interfaces without filters, **Then** I receive a list containing all network interfaces from all nodes with complete attribute information (uuid, name, type, child_type, device_id, network_id, mac_address, ip_address, ipv6_address, dhcp_enabled, card_type, bootable, start_if, base_type)
2. **Given** a BCM cluster with no compute nodes, **When** I query all interfaces, **Then** I receive an empty list without errors
3. **Given** interfaces with missing optional fields (e.g., no network assignment, no IP address), **When** I query all interfaces, **Then** those fields are returned as null without causing errors

---

### User Story 2 - Filter by Parent Node (Priority: P2)

As a cluster administrator, I need to filter network interfaces by their parent compute node so that I can understand the network configuration of specific nodes when troubleshooting or planning maintenance.

**Why this priority**: This is the most common filtering use case - operators typically need to see "what interfaces does this specific node have?" when working with individual nodes.

**Independent Test**: Can be fully tested by creating a test node, querying interfaces with `device_id` filter matching that node's UUID, and verifying only interfaces from that node are returned.

**Acceptance Scenarios**:

1. **Given** a cluster with multiple nodes each having multiple interfaces, **When** I filter by a specific device_id, **Then** I receive only interfaces belonging to that node
2. **Given** a device_id that doesn't exist, **When** I filter by that device_id, **Then** I receive an empty list without errors
3. **Given** a device_id that exists but has no interfaces, **When** I filter by that device_id, **Then** I receive an empty list without errors

---

### User Story 3 - Filter by Interface Type (Priority: P2)

As a cluster administrator, I need to filter interfaces by type (physical, bmc, bond) so that I can audit specific categories of interfaces or configure network policies based on interface types.

**Why this priority**: Interface type filtering is essential for targeted network management - administrators often need to see only bonded interfaces for redundancy verification, or only BMC interfaces for out-of-band management configuration.

**Independent Test**: Can be fully tested by querying interfaces with `type = "bond"` filter and verifying all returned interfaces have type "bond" and child_type "NetworkBondInterface".

**Acceptance Scenarios**:

1. **Given** a cluster with mixed interface types, **When** I filter by type "physical", **Then** I receive only interfaces with child_type "NetworkPhysicalInterface" and type "physical"
2. **Given** a cluster with mixed interface types, **When** I filter by type "bmc", **Then** I receive only interfaces with child_type "NetworkBMCInterface" and type "bmc"
3. **Given** a cluster with mixed interface types, **When** I filter by type "bond", **Then** I receive only interfaces with child_type "NetworkBondInterface" and type "bond"
4. **Given** a filter with an invalid type value, **When** I query interfaces, **Then** I receive an empty list (no matches)

---

### User Story 4 - Filter by Assigned Network (Priority: P3)

As a cluster administrator, I need to filter interfaces by their assigned network so that I can identify which nodes are connected to specific networks for network planning and troubleshooting.

**Why this priority**: Network-based filtering is useful for understanding network membership but is less frequently needed than node-based or type-based filtering. It's valuable for network planning scenarios.

**Independent Test**: Can be fully tested by creating a test network, assigning it to interfaces, and querying with `network_id` filter to verify only interfaces on that network are returned.

**Acceptance Scenarios**:

1. **Given** multiple networks with interfaces assigned, **When** I filter by a specific network_id, **Then** I receive only interfaces assigned to that network
2. **Given** a network_id with no assigned interfaces, **When** I filter by that network_id, **Then** I receive an empty list without errors
3. **Given** interfaces not assigned to any network (null network_id), **When** I filter by a network_id, **Then** those interfaces are excluded from results

---

### User Story 5 - Combine Multiple Filters (Priority: P3)

As a cluster administrator, I need to combine multiple filters (device_id + type, or type + network_id) so that I can perform targeted queries like "show me all bonded interfaces on this specific node" or "show me all physical interfaces on this network".

**Why this priority**: Combined filtering is a power-user feature that enhances query precision but is not essential for basic operations. The individual filters already provide substantial value independently.

**Independent Test**: Can be fully tested by querying with multiple filters simultaneously (e.g., `device_id = "uuid-1"` and `type = "bond"`) and verifying results match ALL filter criteria.

**Acceptance Scenarios**:

1. **Given** multiple filters specified, **When** I query interfaces, **Then** I receive only interfaces matching ALL filter criteria (AND logic)
2. **Given** filters that produce no matching results, **When** I query interfaces, **Then** I receive an empty list without errors
3. **Given** filters with null values, **When** I query interfaces, **Then** null filters are ignored and other filters apply normally

---

### Edge Cases

- What happens when the BCM API is unreachable during a Terraform plan/apply?
  - Expected: Terraform returns a clear error indicating API connectivity issues, plan/apply fails gracefully
- What happens when a node exists but has an empty interfaces array?
  - Expected: No interfaces extracted for that node, contributes zero items to flattened interface list
- What happens when BCM API returns malformed JSON for node or interface data?
  - Expected: Provider returns error with details about parsing failure, including which phase failed (node parsing vs interface parsing)
- What happens when an interface has no MAC address, IP address, or network assignment?
  - Expected: Those fields returned as null (types.StringNull()) without causing errors, interface still included in results
- What happens when querying a very large cluster with thousands of interfaces across hundreds of nodes?
  - Expected: Query completes successfully with reasonable performance (< 30 seconds for 1000+ interfaces from 100+ nodes)
- What happens when an interface has a child_type value that doesn't map to known types (physical/bond/bmc)?
  - Expected: child_type preserved as-is, type field set to normalized lowercase version of child_type for unknown types

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Data source MUST retrieve all nodes using cmdevice.getNodes API method and extract interfaces from nested "interfaces" arrays
- **FR-002**: Data source MUST flatten interface arrays and attach parent node UUID as device_id for each interface
- **FR-003**: Data source MUST support optional client-side filtering by device_id (parent node UUID)
- **FR-004**: Data source MUST support optional client-side filtering by interface type (physical, bmc, bond) - accepting simplified values and mapping to BCM childType
- **FR-005**: Data source MUST support optional client-side filtering by network_id (assigned network UUID from "network" field)
- **FR-006**: Data source MUST return interface attributes: uuid, name, child_type, device_id, network_id, mac_address, ip_address, ipv6_address, dhcp_enabled, card_type, bootable, start_if, base_type
- **FR-007**: Data source MUST expose simplified "type" attribute derived from child_type (NetworkPhysicalInterface → "physical", NetworkBondInterface → "bond", NetworkBMCInterface → "bmc")
- **FR-008**: Data source MUST handle null/missing values for optional fields (network_id, ip_address, ipv6_address) without errors using helper functions
- **FR-009**: Data source MUST support combining multiple filters with AND logic (all filters must match)
- **FR-010**: Data source MUST return an empty list when no interfaces match filter criteria or no nodes exist
- **FR-011**: Data source MUST use null-safe helper functions (getStringValue, getBoolValue, getInt64Value, getStringListValue) for field extraction
- **FR-012**: Data source MUST follow existing provider patterns from data_source_cmdevice_nodes.go (schema structure, filtering logic, API call pattern)
- **FR-013**: Data source MUST be registered in provider.go DataSources() method with factory function NewCMDeviceInterfacesDataSource
- **FR-014**: Implementation MUST include comprehensive acceptance tests covering all user scenarios using modern terraform-plugin-testing patterns
- **FR-015**: Implementation MUST include example Terraform configurations in examples/data-sources/bcm_cmdevice_interfaces/ directory
- **FR-016**: Implementation MUST generate documentation automatically via tfplugindocs (make generate)

### Key Entities

- **Network Interface**: Represents a network interface card (NIC) on a compute node
  - Primary identifier: uuid (string, unique per interface from BCM API)
  - Parent relationship: device_id (derived from parent node's uuid field when flattening)
  - Network assignment: network_id (mapped from BCM API "network" field)
  - Type classification:
    - child_type (BCM API raw value: NetworkPhysicalInterface, NetworkBondInterface, NetworkBMCInterface)
    - type (simplified computed value: physical, bond, bmc)
    - base_type (always "NetworkInterface")
  - Physical attributes:
    - name (interface name like "eth0", "ens33")
    - mac_address (mapped from BCM API "mac" field)
    - ip_address (mapped from BCM API "ip" field)
    - ipv6_address (mapped from BCM API "ipv6Ip" field)
    - card_type (mapped from BCM API "cardtype" field, e.g., "Ethernet", "InfiniBand")
  - Configuration flags:
    - dhcp_enabled (mapped from BCM API "dhcp" boolean field)
    - bootable (PXE boot capability)
    - start_if (startup condition: ALWAYS, NEVER, HOTPLUG)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Administrators can discover all cluster network interfaces in a single Terraform data source query
- **SC-002**: Administrators can filter interfaces by node, type, or network with accurate results
- **SC-003**: Data source handles clusters with 1000+ interfaces without errors or timeouts
- **SC-004**: All acceptance tests pass with 100% success rate
- **SC-005**: Example configurations validate successfully with `terraform validate`
- **SC-006**: Documentation is auto-generated and accurately describes all attributes and filters
- **SC-007**: Data source follows TDD workflow (RED-GREEN-REFACTOR) with tests written before implementation
- **SC-008**: Implementation uses modern terraform-plugin-testing patterns (statecheck, plancheck, knownvalue)
- **SC-009**: Tests are environment-portable with no hardcoded values or assumptions about cluster state

## Clarifications

### Session 2025-11-23

- Q: What is the actual BCM API method for retrieving interfaces? → A: Use cmdevice.getNodes (existing verified method) and extract interfaces from the nested "interfaces" array within each node response. The cmdevice.getInterfaces method does not exist (confirmed by failed methods log).
- Q: What are the BCM API field names for interface data? → A: BCM API uses camelCase: "uuid", "name", "childType" (for interface type), "mac", "ip", "network" (for network_id), "dhcp", "baseType", "cardtype", "bootable", "startIf", "ipv6Ip". Reference: data_source_cmdevice_nodes.go lines 79-92.
- Q: What are the valid interface type values from BCM API? → A: Interface types are indicated by childType field with values: "NetworkPhysicalInterface", "NetworkBondInterface", "NetworkBMCInterface" (based on existing code patterns in data_source_cmdevice_nodes.go and resource_cmdevice_device.go).
- Q: How should bond configuration data be modeled? → A: Bond configuration is available but requires investigating actual BCM API response structure. For now, derive from childType (if "NetworkBondInterface", then it's a bond). Bond mode and members require Phase 0 API exploration.
- Q: What helper function is needed for list/array fields like bond_members? → A: Use getStringListValue helper function (defined in data_source_cmpart_partitions.go lines 337-361) which safely converts []interface{} to types.List with proper null handling.
- Q: How should interface type filtering work? → A: Filter by childType field with exact match on "NetworkPhysicalInterface", "NetworkBondInterface", or "NetworkBMCInterface". Expose as simplified "type" filter accepting "physical", "bond", "bmc" values and map internally to childType values.
- Q: Should the data source read strategy use direct lookup or list-and-filter? → A: Use list-and-filter pattern: call cmdevice.getNodes (no args), extract all interfaces from all nodes, then apply client-side filtering. This follows the standard data source pattern for collections without dedicated BCM API endpoints.
- Q: How should the ID field be populated for the data source? → A: Set ID to types.StringValue("placeholder") following the pattern from data_source_cmdevice_nodes.go line 337. The ID is a framework requirement but has no semantic meaning for collection data sources.
- Q: What BCM API structure contains the device_id (parent node UUID)? → A: Interfaces are nested within node objects, so the device_id must be derived from the parent node's "uuid" field when flattening the interfaces array.
- Q: How should null/missing optional fields be handled? → A: Use existing null-safe helper functions: getStringValue, getBoolValue, getInt64Value (from data_source_cmpart_softwareimages.go lines 463-493). These return types.StringNull(), types.BoolNull(), types.Int64Null() for missing/null values.

## Assumptions

- BCM API method cmdevice.getNodes returns nodes with nested "interfaces" arrays containing interface data
- All interfaces within node responses have uuid, name, childType, and parent node UUID available
- Optional fields (network, ip, ipv6Ip, bond-specific fields) may be null/missing in BCM API responses
- Client-side filtering is acceptable performance-wise (no server-side filtering needed)
- Interface childType values follow BCM convention: NetworkPhysicalInterface, NetworkBondInterface, NetworkBMCInterface
- Standard authentication mechanism (cookie-based via BCM client) applies to cmdevice.getNodes
- Bond configuration details (bond mode, members) require Phase 0 API investigation to determine exact field names

## Dependencies

- Existing BCM API client (internal/provider/bcm_client.go) with CallJSONRPC support for cmdevice.getNodes
- Terraform Plugin Framework v1.16.1+
- Terraform Plugin Testing v1.13.3+ for modern test patterns (statecheck, plancheck, knownvalue)
- Existing helper functions for null-safe field extraction:
  - getStringValue (internal/provider/data_source_cmpart_softwareimages.go)
  - getBoolValue (internal/provider/data_source_cmpart_softwareimages.go)
  - getInt64Value (internal/provider/data_source_cmpart_softwareimages.go)
  - getStringListValue (internal/provider/data_source_cmpart_partitions.go)
- Test helper functions (internal/provider/test_helpers.go):
  - createTestBCMClient
  - generateUniqueTestName
- BCM cluster accessible at configured endpoint for acceptance testing
- Environment variables (BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD) for test configuration

## Out of Scope

- Server-side filtering (all filtering performed client-side in provider)
- Interface creation, modification, or deletion (this is a data source, not a resource)
- Real-time interface status monitoring or change detection
- Interface performance metrics or traffic statistics
- Validation of interface configurations or network assignments
- Integration with external network management systems
- Custom interface attribute extraction beyond BCM API response
