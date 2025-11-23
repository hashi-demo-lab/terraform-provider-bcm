# Feature Specification: BCM Network Resource Management

**Feature Branch**: `002-cmnet-network-resource`
**Created**: 2025-11-23
**Status**: Draft
**Input**: User description: "Implement resource.bcm_cmnet_network to create and manage networks in BCM"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Basic Network Creation (Priority: P1)

Infrastructure administrators need to create basic network configurations in BCM for cluster networking using Terraform declarative infrastructure-as-code.

**Why this priority**: This is the foundational capability - without the ability to create basic networks, no other network management features are possible. This delivers immediate value by enabling network provisioning via Terraform.

**Independent Test**: Can be fully tested by creating a network with only required attributes (name) and verifying it exists in BCM with auto-assigned defaults for optional fields.

**Acceptance Scenarios**:

1. **Given** no existing network with name "test-network", **When** user applies Terraform config with `bcm_cmnet_network` resource specifying only `name = "test-network"`, **Then** BCM creates the network with BCM-assigned defaults and Terraform state captures the UUID
2. **Given** an existing network created via Terraform, **When** user runs `terraform plan` without changes, **Then** Terraform reports no changes needed (idempotency verification)
3. **Given** a network created via Terraform, **When** user runs `terraform destroy`, **Then** the network is removed from BCM and Terraform state is cleaned

---

### User Story 2 - Network Update Management (Priority: P2)

Infrastructure administrators need to modify network configurations (MTU, gateway, notes) after initial creation to adapt to changing cluster requirements.

**Why this priority**: After basic creation works, administrators need the ability to update configurations without destroying and recreating networks, which would disrupt cluster operations.

**Independent Test**: Can be tested independently by creating a network, then updating specific attributes (e.g., changing MTU from default to 9000) and verifying the change persists in both BCM and Terraform state.

**Acceptance Scenarios**:

1. **Given** an existing network with default MTU, **When** user updates `mtu = 9000` in Terraform config and applies, **Then** BCM updates the network MTU and Terraform state reflects the change
2. **Given** a network with specific gateway configured, **When** user removes the gateway attribute from config and applies, **Then** BCM resets gateway to null/default and Terraform state updates accordingly

---

### User Story 3 - DHCP Configuration (Priority: P2)

Network administrators need to enable DHCP services on networks with configured IP ranges to automatically provision cluster nodes without manual IP assignment.

**Why this priority**: DHCP automation is critical for scalable cluster management, but the core network resource must exist first (depends on P1).

**Independent Test**: Can be tested by creating a network with `dhcp_enabled = true`, `dhcp_range_start`, and `dhcp_range_end` attributes, then verifying DHCP service is active in BCM.

**Acceptance Scenarios**:

1. **Given** a network with base_address and netmask_bits configured, **When** user enables DHCP with valid IP range, **Then** BCM activates DHCP service and Terraform state shows dhcp_enabled = true
2. **Given** a network with DHCP enabled, **When** user disables DHCP by setting `dhcp_enabled = false`, **Then** BCM deactivates DHCP service while preserving the network configuration

---

### User Story 4 - VLAN Segmentation (Priority: P3)

Network architects need to configure VLAN IDs for network isolation to implement multi-tenant or security-segmented cluster architectures.

**Why this priority**: VLAN support is important for advanced networking scenarios but not required for basic cluster networking (depends on P1).

**Independent Test**: Can be tested by creating a network with specific VLAN ID (e.g., vlan_id = 100) and verifying the VLAN tag is applied in BCM.

**Acceptance Scenarios**:

1. **Given** no VLAN configured, **When** user creates network with `vlan_id = 100`, **Then** BCM applies VLAN tag 100 to the network
2. **Given** an existing network without VLAN, **When** user adds `vlan_id = 200` via update, **Then** BCM applies the VLAN tag without recreating the network

---

### User Story 5 - Import Existing Networks (Priority: P2)

Infrastructure teams need to import manually-created or pre-existing BCM networks into Terraform state to establish infrastructure-as-code management over existing resources.

**Why this priority**: Essential for brownfield deployments where networks already exist but need to be managed by Terraform going forward.

**Independent Test**: Can be tested by manually creating a network via BCM UI/API, then importing it into Terraform using `terraform import bcm_cmnet_network.test <uuid>` and verifying state matches actual BCM configuration.

**Acceptance Scenarios**:

1. **Given** a network created outside Terraform with UUID "abc-123", **When** user runs `terraform import bcm_cmnet_network.imported abc-123`, **Then** Terraform state accurately reflects all network attributes from BCM
2. **Given** an imported network, **When** user runs `terraform plan`, **Then** Terraform shows no changes if config matches actual BCM state

---

### User Story 6 - Drift Detection (Priority: P3)

Operations teams need Terraform to detect when network configurations are modified outside Terraform (via BCM UI or API) to maintain infrastructure-as-code consistency and prevent configuration drift.

**Why this priority**: Important for operational safety and consistency but depends on basic CRUD operations working first.

**Independent Test**: Can be tested by creating a network via Terraform, modifying it externally (e.g., changing MTU via BCM API), then running `terraform plan` and verifying Terraform detects the drift and proposes corrective changes.

**Acceptance Scenarios**:

1. **Given** a network managed by Terraform with MTU 1500, **When** an administrator changes MTU to 9000 via BCM UI, **Then** next `terraform plan` detects the drift and proposes to restore MTU to 1500
2. **Given** drift detected, **When** user runs `terraform apply`, **Then** Terraform restores the network configuration to match the declared Terraform config

---

### Edge Cases

- **Duplicate network names**: What happens when attempting to create a network with a name that already exists in BCM? (Expected: BCM API error, Terraform reports error to user)
- **Invalid CIDR notation**: How does the system handle malformed subnet values like "10.0.0.0/33" or "not-a-cidr"? (Expected: Client-side validation fails before API call)
- **DHCP range outside subnet**: What happens when dhcp_range_start/end are configured outside the network's subnet? (Expected: BCM API validation error)
- **VLAN ID out of range**: How does BCM handle VLAN IDs outside valid range (1-4094)? (Expected: Client-side validation or BCM API error)
- **Network in use during deletion**: What happens when trying to delete a network that has active node assignments? (Expected: BCM API error with actionable message; may require force=true parameter)
- **Concurrent modifications**: How does system handle simultaneous updates to the same network from multiple Terraform processes? (Expected: BCM API handles via revision/locking mechanism)
- **IPv6 configuration**: Does the resource support IPv6-only or dual-stack configurations? (Assumption: IPv6 attributes are optional/future enhancement based on data source schema)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST allow users to create networks with unique names via Terraform resource `bcm_cmnet_network`
- **FR-002**: System MUST support updating network configurations (MTU, gateway, DHCP settings, notes) without destroying and recreating the resource
- **FR-003**: System MUST persist network UUID as computed `id` and `uuid` attributes in Terraform state
- **FR-004**: Users MUST be able to import existing BCM networks into Terraform state using UUID identifier
- **FR-005**: System MUST delete networks from BCM when Terraform resource is destroyed
- **FR-006**: System MUST support configuring DHCP services with IP range specification (dhcp_range_start, dhcp_range_end)
- **FR-007**: System MUST allow optional VLAN ID assignment for network segmentation
- **FR-008**: System MUST support subnet specification in CIDR notation (e.g., "10.0.0.0/24")
- **FR-009**: System MUST validate network configuration inputs (CIDR format, IP addresses, VLAN range) before submitting to BCM API
- **FR-010**: System MUST detect configuration drift when networks are modified outside Terraform
- **FR-011**: System MUST preserve Terraform plan values for attributes that BCM may reset during operations
- **FR-012**: System MUST support force parameter for deletion of networks with active dependencies
- **FR-013**: System MUST map Terraform attribute names (snake_case) to BCM API field names (camelCase) correctly
- **FR-014**: System MUST use BCM API service "cmnet" with appropriate methods: addNetwork (create), updateNetwork (update), removeNetwork (delete), getNetwork (read)
- **FR-015**: System MUST handle BCM entity structure requirements (baseType, childType, modified, to_be_removed, revision, uuid fields)

### Key Entities

- **Network Entity**: Represents a BCM network configuration including identity (name, UUID), IP addressing (base_address, netmask_bits, gateway), DHCP settings (dynamic_range_start/end), VLAN configuration, MTU, and metadata (notes, revision, modified flags). Maps to BCM CMNet service Network entity with baseType="Network".

- **DHCP Configuration**: Logical grouping of DHCP-related attributes (dhcp_enabled derived from dynamic range presence, dhcp_range_start/end IP addresses). DHCP is enabled when both range start and end are non-null and not "0.0.0.0".

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Infrastructure administrators can create a basic network (name only) in under 30 seconds via Terraform apply
- **SC-002**: All CRUD operations (Create, Read, Update, Delete) complete successfully with 100% pass rate in acceptance tests (TF_ACC=1)
- **SC-003**: Network import functionality successfully imports 100% of existing BCM networks with complete attribute fidelity
- **SC-004**: Drift detection identifies external configuration changes with 100% accuracy within one Terraform plan cycle
- **SC-005**: Resource demonstrates idempotency - repeated `terraform apply` with no config changes produces empty plan 100% of the time
- **SC-006**: Terraform documentation is auto-generated successfully via `make generate` with zero manual edits required
- **SC-007**: DHCP configuration changes (enable/disable/range updates) apply within 60 seconds with zero downtime to existing network traffic
- **SC-008**: Resource supports at least 7 distinct test scenarios (basic create, DHCP config, VLAN config, update, delete, import, drift detection) with 100% pass rate

## Assumptions

- **ASSUM-001**: BCM API supports network entity CRUD operations via cmnet service (confirmed via existing data source implementation)
- **ASSUM-002**: Network names must be unique within BCM cluster (standard BCM constraint)
- **ASSUM-003**: BCM API uses camelCase field names while Terraform uses snake_case (confirmed via data_source_cmnet_networks.go mapping)
- **ASSUM-004**: DHCP enabled status is derived from presence of non-zero dynamic_range_start and dynamic_range_end values (confirmed in data source logic)
- **ASSUM-005**: BCM network entity follows standard entity structure with baseType/childType/revision/modified/to_be_removed fields (confirmed via resource_cmdevice_category.go pattern)
- **ASSUM-006**: Network deletion may require force=true parameter when network has active node assignments (standard BCM deletion pattern)
- **ASSUM-007**: IPv6 configuration attributes exist in BCM schema but are deferred as optional/future enhancement (present in data source but marked as stretch goal)
- **ASSUM-008**: MTU default value is 1500 if not specified (industry standard, BCM likely assigns this default)
- **ASSUM-009**: BCM API method "getNetwork" accepts UUID argument for efficient single-network lookup (standard args parameter pattern)
- **ASSUM-010**: Network UUID is stable and does not change during update operations (standard BCM resource behavior)

## Dependencies

- **DEP-001**: Existing data source `bcm_cmnet_networks` provides schema reference and field mapping guidance
- **DEP-002**: BCM API client (`internal/provider/bcm_client.go`) with authenticated session and CallJSONRPC method
- **DEP-003**: Terraform Plugin Framework v1.16+ for resource schema definitions and lifecycle management
- **DEP-004**: terraform-plugin-testing v1.13.3+ for modern test patterns (statecheck, plancheck, knownvalue)
- **DEP-005**: Test helper functions (createTestBCMClient, getResourceUUIDByName, verifyResourceDeleted, generateUniqueTestName)
- **DEP-006**: BCM CMNet service availability at configured endpoint for API operations

## Out of Scope

- **OOS-001**: IPv6-specific configuration (ipv6_enabled, ipv6_base_address, ipv6_gateway, ipv6_netmask_bits) - deferred to future enhancement
- **OOS-002**: Advanced network type configuration (network_type field) - assumes BCM assigns appropriate default
- **OOS-003**: DNS zone generation settings (generate_dns_zone, search_domain_index) - optional BCM-managed settings
- **OOS-004**: Cloud-specific attributes (cloud_subnet_id, ec2_availability_zone) - not applicable for on-premises BCM deployments
- **OOS-005**: Layer 3 routing configuration (layer3, layer3_route, gateway_metric) - advanced networking feature deferred
- **OOS-006**: Bulk network operations or network provisioning workflows - this resource manages individual networks only
- **OOS-007**: Network monitoring, metrics, or health checks - outside provider scope
- **OOS-008**: Integration with external network management systems - BCM is the source of truth

## API Contract Research

Based on data_source_cmnet_networks.go analysis:

### BCM CMNet Service Methods

- `getNetworks()` - Returns array of all network entities (used by data source)
- `getNetwork(uuid)` - Returns single network entity by UUID (for efficient Read operation)
- `addNetwork(entity, force)` - Creates new network, returns created entity with UUID
- `updateNetwork(entity, force)` - Updates existing network, requires full entity with UUID
- `removeNetwork(uuid, force)` - Deletes network by UUID, force parameter for dependencies

### Network Entity Schema (BCM API → Terraform Mapping)

| BCM API Field (camelCase) | Terraform Attribute (snake_case) | Type | Required/Optional/Computed |
|---------------------------|----------------------------------|------|----------------------------|
| uuid | uuid, id | string | Computed |
| name | name | string | Required |
| baseAddress | base_address | string | Optional (derived from subnet) |
| netmaskBits | netmask_bits | int64 | Optional (derived from subnet) |
| gateway | gateway | string | Optional |
| type | network_type | string | Computed (BCM-assigned) |
| mtu | mtu | int64 | Optional (default: 1500) |
| domainName | domain_name | string | Optional |
| dynamicRangeStart | dhcp_range_start | string | Optional (DHCP config) |
| dynamicRangeEnd | dhcp_range_end | string | Optional (DHCP config) |
| management | management | bool | Computed (BCM-assigned) |
| bootable | bootable | bool | Computed (BCM-assigned) |
| notes | notes | string | Optional |
| baseType | base_type | string | Computed (always "Network") |
| childType | child_type | string | Computed |
| revision | revision | string | Computed |
| modified | modified | bool | Computed |
| to_be_removed | to_be_removed | bool | Computed |

### DHCP Enabled Logic

```
dhcp_enabled (computed bool) = (dhcp_range_start != null AND dhcp_range_start != "0.0.0.0" AND
                                 dhcp_range_end != null AND dhcp_range_end != "0.0.0.0")
```

### Subnet Attribute Design Decision

**User Input**: `subnet = "10.0.1.0/24"` (CIDR notation - user-friendly)

**BCM API Requirements**: Separate fields `baseAddress = "10.0.1.0"` and `netmaskBits = 24`

**Implementation Strategy**:
1. Resource schema accepts `subnet` attribute (string, CIDR format)
2. Client-side validation ensures valid CIDR notation (regex: `^(\d{1,3}\.){3}\d{1,3}/\d{1,2}$`)
3. Before API submission, parse CIDR into baseAddress and netmaskBits
4. On Read operation, reconstruct `subnet` attribute from baseAddress + netmaskBits for Terraform state
5. State stores both `subnet` (user-facing) and computed `base_address`/`netmask_bits` for transparency

### VLAN Configuration

- VLAN ID attribute likely maps to a BCM field (needs Phase 0 API exploration to confirm exact field name)
- Standard VLAN range: 1-4094 (client-side validation)
- If BCM API doesn't expose VLAN in getNetworks response, may need separate API call or deferred to future enhancement

### Force Parameter Pattern

- `force` attribute (optional bool, default: false) passed to BCM API operations
- Used during deletion to override dependency checks (e.g., network has assigned nodes)
- Not stored in state (operation-time parameter only)

## Implementation Notes

### TDD Test Coverage Requirements

1. **TestAccCMNetNetwork_Basic** - Create network with name only, verify BCM-assigned defaults
2. **TestAccCMNetNetwork_Subnet** - Create network with CIDR subnet configuration
3. **TestAccCMNetNetwork_DHCP** - Create network with DHCP enabled and IP range
4. **TestAccCMNetNetwork_VLAN** - Create network with VLAN ID (if supported)
5. **TestAccCMNetNetwork_CompleteConfig** - Create network with all optional attributes
6. **TestAccCMNetNetwork_Update** - Update network attributes (MTU, gateway, notes)
7. **TestAccCMNetNetwork_UpdateDHCP** - Enable/disable DHCP dynamically
8. **TestAccCMNetNetwork_Import** - Import existing network by UUID
9. **TestAccCMNetNetwork_DriftDetection** - Detect external modifications to network
10. **TestAccCMNetNetwork_IdempotencyAfterCreate** - Verify empty plan after create
11. **TestAccCMNetNetwork_IdempotencyAfterUpdate** - Verify empty plan after update

### Modern Testing Pattern Usage

- Use `statecheck.ExpectKnownValue()` with `knownvalue.StringExact()`, `knownvalue.Int64Exact()`, `knownvalue.Bool()`
- Use `plancheck.ExpectEmptyPlan()` for idempotency verification
- Use `statecheck.CompareValue(compare.ValuesSame())` for ID consistency tracking across Create/Import/Update
- Generate unique test names: `generateUniqueTestName("test-network")`
- Enhanced CheckDestroy with detailed error messages

### Example Configuration Template

```hcl
resource "bcm_cmnet_network" "example" {
  name    = "compute-network"
  subnet  = "10.0.1.0/24"
  gateway = "10.0.1.1"
  mtu     = 9000

  dhcp_enabled     = true
  dhcp_range_start = "10.0.1.100"
  dhcp_range_end   = "10.0.1.200"

  notes = "High-performance compute network with jumbo frames"
}
```

## Next Steps

After specification approval:

1. **Phase 0 - API Research**: Run BCM API exploration script to confirm network CRUD operations and VLAN field mapping
2. **Phase 1 - Test Development**: Write failing acceptance tests following TDD Red phase
3. **Phase 2 - Implementation**: Implement minimal CRUD logic to pass tests (Green phase)
4. **Phase 3 - Refinement**: Refactor code, add validation, improve error handling (Refactor phase)
5. **Phase 4 - Documentation**: Run `make generate` to create provider documentation
6. **Phase 5 - Integration**: Test with real BCM cluster, validate examples
