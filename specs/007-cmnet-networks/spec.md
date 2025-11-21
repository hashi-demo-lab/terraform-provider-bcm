# Feature Specification: BCM CMNet Networks Data Source

**Feature Branch**: `007-cmnet-networks`
**Created**: 2025-11-21
**Status**: Draft
**Input**: User description: "Create a feature specification for implementing a new Terraform data source bcm_cmnet_networks"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Query All Networks (Priority: P1)

As a Terraform user managing BCM infrastructure, I want to retrieve a list of all networks configured in the BCM cluster so that I can reference network UUIDs in my resource configurations (e.g., assigning nodes to networks).

**Why this priority**: This is the fundamental use case for the data source. Without the ability to list networks, users cannot discover available networks or obtain their UUIDs for configuration.

**Independent Test**: Can be fully tested by querying the data source without filters and verifying that all networks are returned with complete attributes. Delivers immediate value by enabling network discovery.

**Acceptance Scenarios**:

1. **Given** a BCM cluster with multiple networks configured, **When** I query the bcm_cmnet_networks data source without filters, **Then** all networks are returned with their full metadata (name, UUID, IP configuration, DHCP settings).
2. **Given** the BCM API is accessible, **When** I read the data source, **Then** the Terraform state contains accurate network information matching the BCM cluster configuration.

---

### User Story 2 - Filter Networks by Attributes (Priority: P2)

As a Terraform user, I want to filter networks by specific attributes (such as name pattern or network type) so that I can quickly locate the specific network I need without manually searching through all networks.

**Why this priority**: While listing all networks is essential, filtering improves usability for large BCM deployments with many networks. This is a quality-of-life improvement that doesn't block core functionality.

**Independent Test**: Can be tested independently by applying various filter criteria and verifying that only matching networks are returned. The filtering logic is client-side and doesn't depend on API features.

**Acceptance Scenarios**:

1. **Given** a BCM cluster with networks named "management-net", "compute-net", and "storage-net", **When** I filter by name pattern "compute", **Then** only "compute-net" is returned.
2. **Given** multiple networks with different DHCP settings, **When** I filter by DHCP enabled status, **Then** only networks matching that criterion are returned.

---

### User Story 3 - Reference Networks in Other Resources (Priority: P3)

As a Terraform user, I want to use network data in conjunction with other BCM resources so that I can build complete infrastructure configurations (e.g., assign specific network UUIDs to node interfaces).

**Why this priority**: This represents the integration scenario where the data source is used with other Terraform resources. It depends on both the data source working (P1) and potentially other resources being implemented.

**Independent Test**: Can be tested by using the data source output as an input to another resource configuration and verifying the reference resolves correctly.

**Acceptance Scenarios**:

1. **Given** I have queried the bcm_cmnet_networks data source, **When** I reference a network's UUID in a bcm_cmdevice_node resource configuration, **Then** the network UUID is correctly passed to the node configuration.
2. **Given** network information from the data source, **When** I use conditional logic based on network attributes (e.g., DHCP enabled), **Then** Terraform correctly evaluates the conditions.

---

### Edge Cases

- What happens when the BCM cluster has no networks configured? The data source should return an empty list without errors.
- How does the system handle authentication failures? The data source should fail immediately with a clear diagnostic error message indicating authentication issues. No retry logic. Error message must include the HTTP status code and indicate the likely cause (e.g., "Authentication failed: HTTP 401 - verify BCM credentials").
- What happens if the BCM API returns malformed JSON? The provider should detect the parse error and fail immediately with a diagnostic error including context about the malformed response. No retry logic.
- How are network connectivity issues handled? Fail fast with a clear error message indicating the connection failure (e.g., "Failed to connect to BCM API at https://172.21.15.254:8081: connection refused"). No automatic retry.
- How are null or missing fields in the API response handled? The null-safe helper functions convert both null values and empty strings to Terraform null (normalized to single null representation, no distinction).
- What happens when filtering produces no matches? The data source should return an empty networks list (not an error).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Data source MUST call the BCM JSON-RPC API with service "cmnet" and method "getNetworks"
- **FR-002**: Data source MUST use the existing BCMClient.CallJSONRPC() method for API communication
- **FR-003**: Data source MUST parse the JSON array response and map each network object to a Terraform model
- **FR-004**: Data source MUST handle all network attributes including name, UUID, network address, netmask, gateway, domain, and DHCP settings
- **FR-005**: Data source MUST implement client-side filtering (not API-side) for network attributes
- **FR-006**: Data source MUST use null-safe helper functions (getStringValue, getBoolValue, getInt64Value) for field extraction, normalizing all empty strings and null values to Terraform null (no distinction between empty string and null)
- **FR-007**: Data source MUST return an empty list (not an error) when no networks match filter criteria
- **FR-008**: Data source MUST set a placeholder ID for the data source instance
- **FR-009**: Acceptance tests MUST be written before implementation (TDD RED phase)
- **FR-010**: Acceptance tests MUST verify data source can be read without errors
- **FR-011**: Acceptance tests MUST verify returned network attributes match expected values
- **FR-012**: Acceptance tests MUST include provider configuration block with BCM credentials
- **FR-018**: Acceptance tests MUST use existing networks in the BCM cluster (read-only operations, no create/destroy of test networks)
- **FR-013**: Implementation MUST follow terraform-plugin-framework patterns for data sources
- **FR-014**: Implementation MUST include comprehensive schema documentation with MarkdownDescription for all attributes
- **FR-015**: Examples directory MUST contain a working Terraform configuration demonstrating data source usage
- **FR-016**: Documentation MUST be auto-generated using tfplugindocs tool
- **FR-017**: Error handling MUST follow fail-fast pattern with clear diagnostic messages (no retry logic, include HTTP status codes and actionable guidance for authentication, network, and JSON parsing errors)

### Key Entities *(include if feature involves data)*

- **Network**: Represents a network configuration in the BCM cluster with attributes:
  - Identity: UUID (unique identifier), name (human-readable network name)
  - Network Configuration: network address (base IP), netmask (subnet mask), gateway (default gateway IP)
  - DNS/Domain: domain name for the network
  - DHCP Settings: boolean flag indicating if DHCP is enabled
  - Metadata: baseType (always "Network"), childType (network subtype if applicable)
  - State flags: modified (unsaved changes), to_be_removed (scheduled for deletion)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can retrieve all networks from BCM cluster in under 5 seconds (for typical clusters with up to 100 networks)
- **SC-002**: Data source accurately reflects 100% of network attributes available in the BCM API response
- **SC-003**: Acceptance tests achieve 100% pass rate in CI/CD pipeline
- **SC-004**: Documentation is generated automatically and includes all attributes with clear descriptions
- **SC-005**: Data source follows the same code structure and patterns as existing data sources (verified through code review)
- **SC-006**: Filtering logic correctly processes all defined filter criteria with zero false positives or negatives
- **SC-007**: Error messages provide actionable guidance for authentication failures, network issues, and API errors

## API Contract

### Request Format

The data source will make the following JSON-RPC API call:

```json
{
  "service": "cmnet",
  "call": "getNetworks"
}
```

**HTTP Details**:
- Method: POST
- Endpoint: `https://<bcm-endpoint>/json`
- Headers: `Content-Type: application/json`, `Cookie: cm-login-token=<token>`
- Authentication: Cookie-based session (handled by BCMClient)

### Response Format

The BCM API returns a JSON array of network objects. Based on the BCM API documentation pattern and similar services, the expected response structure is:

```json
[
  {
    "baseType": "Network",
    "childType": "Network",
    "uuid": "550e8400-e29b-41d4-a716-446655440000",
    "name": "management-net",
    "network": "192.168.1.0",
    "netmask": "255.255.255.0",
    "gateway": "192.168.1.1",
    "domain": "cluster.local",
    "dhcp": true,
    "modified": false,
    "to_be_removed": false,
    "creationTime": 1637856000,
    "revision": "1.0",
    "revisionID": 1
  },
  {
    "baseType": "Network",
    "childType": "Network",
    "uuid": "660e8400-e29b-41d4-a716-446655440001",
    "name": "compute-net",
    "network": "10.0.0.0",
    "netmask": "255.255.255.0",
    "gateway": "10.0.0.1",
    "domain": "compute.local",
    "dhcp": false,
    "modified": false,
    "to_be_removed": false,
    "creationTime": 1637856100,
    "revision": "1.0",
    "revisionID": 1
  }
]
```

**Field Descriptions**:
- `baseType`: Entity category (always "Network")
- `childType`: Specific network subtype (typically "Network", may have specialized types)
- `uuid`: Unique network identifier (RFC 4122 UUID format)
- `name`: Human-readable network name
- `network`: Network base address (IPv4 format)
- `netmask`: Subnet mask (IPv4 format)
- `gateway`: Default gateway IP address (IPv4 format, may be null)
- `domain`: DNS domain name for the network (may be null)
- `dhcp`: Boolean indicating if DHCP is enabled for this network
- `modified`: Boolean indicating if network has unsaved changes
- `to_be_removed`: Boolean indicating if network is scheduled for deletion
- `creationTime`: Unix timestamp (seconds) when network was created
- `revision`: Revision string (may be null)
- `revisionID`: Numeric revision identifier (may be 0 or null)

## Schema Definition

### Data Source Schema

```hcl
data "bcm_cmnet_networks" "example" {
  # Optional filter block
  filter {
    name_pattern = "compute"  # Case-insensitive substring match
    dhcp_enabled = true       # Exact boolean match
  }
}

# Output attributes
output "networks" {
  value = data.bcm_cmnet_networks.example.networks
}
```

### Terraform Schema Attributes

**Root Attributes**:
- `id` (String, Computed): Placeholder identifier for the data source
- `networks` (List of Object, Computed): List of network objects matching filter criteria

**Filter Block** (Optional, SingleNestedBlock):
- `name_pattern` (String, Optional): Filter networks by name using case-insensitive substring match (e.g., "compute" matches "Compute-Net", "my-compute", "COMPUTE-STORAGE")
- `dhcp_enabled` (Boolean, Optional): Filter by DHCP enabled status (exact boolean match). When not specified, returns networks with both dhcp=true and dhcp=false. When set to true, returns only DHCP-enabled networks. When set to false, returns only non-DHCP networks.

**Network Object Attributes** (within `networks` list):
- `id` (String, Computed): Network UUID (same as uuid)
- `uuid` (String, Computed): Unique network identifier
- `name` (String, Computed): Network name
- `network` (String, Computed): Network base address (e.g., "192.168.1.0")
- `netmask` (String, Computed): Subnet mask (e.g., "255.255.255.0")
- `gateway` (String, Computed): Default gateway IP address
- `domain` (String, Computed): DNS domain name
- `dhcp` (Boolean, Computed): DHCP enabled flag
- `base_type` (String, Computed): Entity base type (always "Network")
- `child_type` (String, Computed): Network subtype
- `creation_time` (Int64, Computed): Unix timestamp when network was created
- `revision` (String, Computed): Revision string
- `revision_id` (Int64, Computed): Numeric revision identifier
- `modified` (Boolean, Computed): Indicates unsaved changes
- `to_be_removed` (Boolean, Computed): Indicates scheduled for deletion

## Implementation Patterns

### Code Structure

The implementation must follow the established patterns from existing data sources:

1. **File Location**: `internal/provider/data_source_cmnet_networks.go`
2. **Test Location**: `internal/provider/data_source_cmnet_networks_test.go`
3. **Example Location**: `examples/data-sources/bcm_cmnet_networks/data-source.tf`

### Required Components

1. **Data Source Struct**: Implements `datasource.DataSource` and `datasource.DataSourceWithConfigure` interfaces
2. **Model Structs**:
   - `CMNetNetworksDataSourceModel` (root model with ID, Filter, Networks)
   - `NetworkFilterModel` (filter criteria)
   - `NetworkModel` (individual network attributes)
3. **Interface Methods**:
   - `Metadata()`: Returns data source type name
   - `Schema()`: Defines Terraform schema
   - `Configure()`: Receives BCMClient from provider
   - `Read()`: Executes API call and populates state
4. **Helper Functions**:
   - `mapAPIToNetwork()`: Converts API response to NetworkModel
   - `matchesFilter()`: Client-side filtering logic
   - Reuse existing `getStringValue()`, `getBoolValue()`, `getInt64Value()` from `data_source_cmpart_softwareimages.go`

### TDD Workflow

**Phase 1 - RED (Write Failing Tests)**:
1. Create `data_source_cmnet_networks_test.go`
2. Implement `TestAccCMNetNetworksDataSource_Basic` acceptance test (read-only, uses existing networks)
3. Test configuration includes provider block with BCM credentials
4. Tests assume at least one network exists in the BCM cluster
5. Run test with `TF_ACC=1` - verify it fails (data source not registered)

**Phase 2 - GREEN (Minimal Implementation)**:
1. Create `data_source_cmnet_networks.go` with minimal implementation
2. Register data source in `provider.go` DataSources() method
3. Implement Schema() with all attributes
4. Implement Read() with API call and basic mapping
5. Run test with `TF_ACC=1` - verify it passes

**Phase 3 - REFACTOR (Improve Code)**:
1. Add comprehensive error handling
2. Implement client-side filtering logic
3. Add additional test cases (filtering, edge cases)
4. Add debug logging with tflog
5. Verify all tests still pass

**Phase 4 - DOCUMENT**:
1. Create example configuration in `examples/data-sources/bcm_cmnet_networks/`
2. Run `make generate` to generate documentation
3. Verify generated docs are accurate and complete

## Test Scenarios

**Test Data Strategy**: All acceptance tests MUST use existing networks in the BCM cluster (read-only operations). Tests MUST NOT create, modify, or destroy networks. This follows the data source pattern where tests verify read operations against live infrastructure.

### Acceptance Test 1: Basic Read

**Objective**: Verify data source can retrieve networks from BCM API using existing networks

**Test Configuration**:
```hcl
provider "bcm" {
  endpoint             = "<BCM_ENDPOINT>"
  username             = "<BCM_USERNAME>"
  password             = "<BCM_PASSWORD>"
  insecure_skip_verify = true
}

data "bcm_cmnet_networks" "test" {}
```

**Assertions**:
- `data.bcm_cmnet_networks.test.id` is set
- `data.bcm_cmnet_networks.test.networks` is a non-empty list (assumes at least one network exists in BCM cluster)
- At least one network has `name` attribute set
- At least one network has `uuid` attribute set
- At least one network has `network` attribute set

### Acceptance Test 2: Filtering by Name Pattern

**Objective**: Verify client-side filtering by name works correctly using case-insensitive substring matching

**Test Configuration**:
```hcl
data "bcm_cmnet_networks" "filtered" {
  filter {
    name_pattern = "management"  # Matches "management", "Management-Net", "MY-MANAGEMENT", etc.
  }
}
```

**Assertions**:
- All returned networks have names containing "management" (case-insensitive substring match)
- Networks not matching the pattern are excluded
- Matching is case-insensitive (e.g., "management" matches "Management-Net" and "MANAGEMENT")

### Acceptance Test 3: Filtering by DHCP Status

**Objective**: Verify filtering by boolean attributes

**Test Configuration**:
```hcl
data "bcm_cmnet_networks" "dhcp_enabled" {
  filter {
    dhcp_enabled = true
  }
}
```

**Assertions**:
- All returned networks have `dhcp = true`
- Networks with `dhcp = false` are excluded

### Acceptance Test 4: No Matches (Empty Result)

**Objective**: Verify graceful handling when no networks match filter

**Test Configuration**:
```hcl
data "bcm_cmnet_networks" "no_match" {
  filter {
    name_pattern = "nonexistent-network-pattern-xyz"
  }
}
```

**Assertions**:
- No error is returned
- `networks` attribute is an empty list

## Example Terraform Configuration

### Basic Usage

```hcl
# Provider configuration
provider "bcm" {
  endpoint             = "https://172.21.15.254:8081"
  username             = "root"
  password             = "Hashicorp123!"
  insecure_skip_verify = true
}

# Retrieve all networks
data "bcm_cmnet_networks" "all" {}

# Output network information
output "all_networks" {
  value = data.bcm_cmnet_networks.all.networks
}

output "network_count" {
  value = length(data.bcm_cmnet_networks.all.networks)
}
```

### Filtered Usage

```hcl
# Find management network
data "bcm_cmnet_networks" "management" {
  filter {
    name_pattern = "management"
  }
}

# Use network UUID in other resources
output "management_network_uuid" {
  value = length(data.bcm_cmnet_networks.management.networks) > 0 ? data.bcm_cmnet_networks.management.networks[0].uuid : null
}
```

### Advanced Usage

```hcl
# Find all DHCP-enabled networks
data "bcm_cmnet_networks" "dhcp_networks" {
  filter {
    dhcp_enabled = true
  }
}

# Create map of network names to UUIDs
locals {
  network_map = {
    for net in data.bcm_cmnet_networks.dhcp_networks.networks :
    net.name => net.uuid
  }
}

output "dhcp_network_map" {
  value = local.network_map
}
```

## Documentation Requirements

### Auto-Generated Documentation

The `make generate` command must produce documentation in `docs/data-sources/cmnet_networks.md` with:

1. **Description**: Clear explanation of the data source purpose
2. **Example Usage**: Working code examples from `examples/` directory
3. **Schema Reference**: Complete attribute listing with types and descriptions
4. **Filter Options**: Detailed explanation of available filters
5. **Notes**: Important considerations (authentication, API limitations, etc.)

### Example Directory Structure

```
examples/
└── data-sources/
    └── bcm_cmnet_networks/
        ├── data-source.tf       # Basic usage example
        └── filtered.tf          # Filtering examples
```

## Assumptions

1. **API Endpoint Availability**: The BCM API endpoint `{"service": "cmnet", "call": "getNetworks"}` exists and returns a JSON array of network objects (based on API documentation pattern).

2. **Network Attribute Naming**: Field names in the API response follow the same camelCase convention as other BCM services (e.g., `creationTime`, `revisionID`).

3. **Authentication**: Cookie-based authentication is already handled by the existing BCMClient implementation and doesn't require changes.

4. **Helper Function Reuse**: The null-safe helper functions (`getStringValue`, `getBoolValue`, `getInt64Value`) currently in `data_source_cmpart_softwareimages.go` will remain available for reuse (they may need to be moved to a shared location if they aren't already).

5. **Test Environment**: Acceptance tests assume a live BCM environment is available at the endpoint specified in environment variables, with at least one network configured. Tests are read-only and use existing networks (do not create or destroy test networks).

6. **IPv4 Only**: Network addresses, netmasks, and gateways are assumed to be IPv4 format based on typical BCM deployments. IPv6 support may exist but is not prioritized for initial implementation.

7. **No Nested Entities**: Unlike nodes (which have interfaces and roles) or software images (which have modules), networks are assumed to be flat entities without nested sub-entities.

8. **Standard BCM Response Format**: The API response follows the standard BCM pattern of returning an array of objects with `baseType`, `childType`, `uuid`, and common metadata fields.

## Dependencies

- **BCMClient**: Existing client in `internal/provider/bcm_client.go` with `CallJSONRPC()` method
- **Helper Functions**: `getStringValue()`, `getBoolValue()`, `getInt64Value()` from existing data sources
- **Provider Registration**: Modification to `provider.go` to register new data source
- **Test Framework**: `terraform-plugin-testing` for acceptance tests
- **Documentation Tool**: `tfplugindocs` for generating provider documentation

## Clarifications

### Session 2025-11-21

- Q: How should the name_pattern filter match network names? Should it be case-sensitive or case-insensitive? Exact match or substring? → A: Case-insensitive substring contains
- Q: How should the data source handle null or empty string values from the API (e.g., gateway, domain fields that may be null)? → A: Normalize all to null
- Q: When the user specifies no dhcp_enabled filter (tri-state: true/false/unspecified), what networks should be returned? → A: Return both when not specified
- Q: When the BCM API returns errors (authentication failures, network issues, malformed JSON), should the data source retry, partially succeed, or fail immediately? → A: Fail fast with clear error messages
- Q: Should acceptance tests create/destroy test networks, or use existing networks in the BCM cluster? → A: Use existing networks only

## Out of Scope

- **Resource Implementation**: This specification covers only the data source (read-only). Creating, updating, or deleting networks (resource implementation) is out of scope.
- **Advanced Filtering**: Complex filtering logic (regex patterns, IP range matching, CIDR notation parsing) is deferred to future iterations.
- **Data Source Arguments**: Querying a single network by name or UUID (similar to resource Read) is out of scope for this data source, which focuses on listing all networks.
- **IPv6 Attribute Mapping**: While the API may return IPv6 fields, they are not prioritized for initial implementation.
- **Network Interface Relationships**: Mapping which node interfaces use which networks is out of scope (belongs in node data source/resource).
- **Performance Optimization**: Caching API responses or implementing pagination is out of scope for the initial implementation.
