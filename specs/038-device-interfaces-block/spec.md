# Feature Specification: Add Interfaces Block to bcm_cmdevice_device

**Feature Branch**: `038-device-interfaces-block`
**Created**: 2025-11-25
**Status**: Draft
**GitHub Issue**: #38
**Priority**: P0 - CRITICAL (Blocking DGX Deployment)
**Input**: User description: "Enhance bcm_cmdevice_device - Add Interfaces Block"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Configure Multiple Physical Interfaces (Priority: P1)

As a cluster administrator deploying DGX nodes, I need to configure multiple network interfaces on a device so that I can assign different interfaces to different networks (management, data, storage) for proper network segmentation.

**Why this priority**: This is the fundamental capability. Without multiple interface support, administrators cannot deploy DGX systems which require at minimum separate management and data networks. This is blocking production deployments.

**Independent Test**: Can be fully tested by creating a device with two physical interfaces on different networks and verifying both interfaces are created with correct network assignments.

**Acceptance Scenarios**:

1. **Given** a valid device configuration with two interfaces blocks, **When** I apply the configuration, **Then** the device is created with both interfaces assigned to their respective networks
2. **Given** an existing device with one interface, **When** I add a second interfaces block and apply, **Then** the new interface is added without disrupting the existing interface
3. **Given** an interfaces block with name, type, and network, **When** I apply the configuration, **Then** the interface is created with DHCP enabled by default and bootable set to false unless specified

---

### User Story 2 - Configure Bonded Interfaces (Priority: P1)

As a cluster administrator, I need to create bonded network interfaces so that I can achieve network redundancy and increased bandwidth for critical workloads.

**Why this priority**: Bonded interfaces are essential for production DGX deployments requiring high availability. Network bond configuration is a core requirement for enterprise clusters.

**Independent Test**: Can be fully tested by creating a device with a bond interface specifying member interfaces and verifying the bond is created with correct member assignment.

**Acceptance Scenarios**:

1. **Given** an interfaces block with type "bond" and members array, **When** I apply the configuration, **Then** a bonded interface is created with the specified member interfaces
2. **Given** a bond interface configuration without specifying bond_mode, **When** I apply the configuration, **Then** the bond is created with BCM's default bond mode
3. **Given** a bond interface with invalid member interface names, **When** I apply the configuration, **Then** a validation error is returned indicating the invalid members

---

### User Story 3 - Configure BMC/IPMI Interface (Priority: P2)

As a cluster administrator, I need to configure a BMC (Baseboard Management Controller) interface so that I can enable out-of-band management for hardware power control and console access.

**Why this priority**: BMC interface configuration is required for remote management capabilities but can be configured after initial device deployment if needed.

**Independent Test**: Can be fully tested by creating a device with a BMC interface block and verifying the interface is created with the correct type and network assignment.

**Acceptance Scenarios**:

1. **Given** an interfaces block with type "bmc", **When** I apply the configuration, **Then** a BMC interface is created with child_type "NetworkBMCInterface"
2. **Given** a BMC interface with an IP address specified, **When** I apply the configuration, **Then** the BMC interface has the configured static IP
3. **Given** a device without existing BMC interface, **When** I add a BMC interface block, **Then** the BMC interface is added to the device

---

### User Story 4 - Import Device with Interfaces (Priority: P2)

As a cluster administrator, I need to import existing devices that have interfaces configured so that I can manage them through Terraform without recreating them.

**Why this priority**: Import functionality is essential for adopting existing infrastructure into Terraform management. Devices deployed before this enhancement need to be importable.

**Independent Test**: Can be fully tested by importing an existing device with multiple interfaces and verifying all interface configurations are populated in state.

**Acceptance Scenarios**:

1. **Given** an existing device with multiple interfaces in BCM, **When** I import it into Terraform, **Then** all interfaces are populated in the state with correct attributes
2. **Given** an imported device with interfaces, **When** I run terraform plan with no changes, **Then** no drift is detected (plan is empty)
3. **Given** an imported device, **When** I modify an interface network assignment and apply, **Then** only the changed interface is updated

---

### User Story 5 - Remove and Replace Interfaces (Priority: P3)

As a cluster administrator, I need to remove interfaces from a device and replace them with different configurations so that I can reconfigure network topology without recreating devices.

**Why this priority**: Interface lifecycle management is important for ongoing operations but is less critical than initial configuration capabilities.

**Independent Test**: Can be fully tested by creating a device with interfaces, removing one interface block, and verifying the interface is removed from the device.

**Acceptance Scenarios**:

1. **Given** a device with two interfaces, **When** I remove one interface block and apply, **Then** the removed interface is deleted from the device
2. **Given** a device with a physical interface, **When** I change it to a bond interface in config, **Then** the old interface is replaced with the new bond interface
3. **Given** a device with the provisioning interface, **When** I attempt to remove it, **Then** a validation error prevents removal (provisioning interface required)

---

### Edge Cases

- What happens when creating a bond interface with only one member?
  - Expected: BCM validation error indicating bonds require at least two members
- What happens when two interfaces have the same name?
  - Expected: Terraform validation error before API call - interface names must be unique per device
- What happens when the provisioning interface MAC address doesn't match any interface?
  - Expected: BCM validation error indicating provisioning interface not found
- What happens when an interface references a non-existent network UUID?
  - Expected: BCM validation error indicating network not found
- What happens when removing all interfaces from a device?
  - Expected: BCM validation error - at least one interface with provisioning capability required
- What happens when a bond member interface is also configured as a standalone interface?
  - Expected: BCM validation error - interface cannot be both standalone and bond member

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Resource MUST support a nested `interfaces` block (list of objects) in the schema allowing zero or more interface configurations
- **FR-002**: Each interfaces block MUST support the following attributes:
  - `name` (Required, String): Interface name (e.g., "eth0", "bond0", "ipmi")
  - `type` (Required, String): Interface type - "physical", "bond", or "bmc"
  - `network` (Optional, String): Network UUID reference for interface assignment
  - `mac` (Optional, String): MAC address (required for physical interfaces on create)
  - `ip` (Optional, String): Static IPv4 address
  - `ipv6_ip` (Optional, String): Static IPv6 address
  - `dhcp` (Optional, Bool, Default: true): Enable DHCP for IP assignment
  - `bootable` (Optional, Bool, Default: false): Enable PXE boot capability
  - `start_if` (Optional, String, Default: "ALWAYS"): Interface startup condition
  - `members` (Optional, List of String): Member interface names (required for bond type)
  - `bond_mode` (Optional, String): Bond mode (e.g., "802.3ad", "active-backup")
- **FR-003**: Resource MUST preserve the existing `mac` field at the device level for backward compatibility (primary MAC for provisioning interface)
- **FR-004**: Resource MUST preserve the existing `management_network` field for backward compatibility (used when no interfaces block is specified)
- **FR-005**: When interfaces block is specified, resource MUST use interfaces block for network configuration instead of legacy mac/management_network fields
- **FR-006**: Resource MUST map interface type to BCM childType:
  - "physical" maps to "NetworkPhysicalInterface"
  - "bond" maps to "NetworkBondInterface"
  - "bmc" maps to "NetworkBMCInterface"
- **FR-007**: Resource MUST call BCM validateDevice API before create/update to validate interface configuration
- **FR-008**: Resource MUST handle interface ordering - the first bootable interface becomes the provisioning interface
- **FR-009**: Resource MUST support interface import - Read operation MUST populate interfaces block from BCM API response
- **FR-010**: Resource MUST detect drift in interface configurations during Read operation
- **FR-011**: Resource MUST support updating individual interface attributes without recreating other interfaces
- **FR-012**: Resource MUST validate interface names are unique within a device (Terraform-side validation)
- **FR-013**: Resource MUST validate bond members are specified when type is "bond" (Terraform-side validation)
- **FR-014**: Resource MUST include computed attributes in interfaces block:
  - `uuid` (Computed, String): BCM-assigned interface UUID
  - `base_type` (Computed, String): Always "NetworkInterface"
  - `child_type` (Computed, String): BCM-determined interface type
  - `cardtype` (Computed, String): Hardware card type (e.g., "Ethernet", "InfiniBand")
- **FR-015**: Implementation MUST include comprehensive acceptance tests covering all interface types using modern terraform-plugin-testing patterns
- **FR-016**: Implementation MUST include example Terraform configurations in examples/resources/bcm_cmdevice_device/ directory showing interface usage
- **FR-017**: Resource MUST generate documentation automatically via tfplugindocs (make generate)

### Key Entities

- **Device Interface**: Represents a network interface configuration within a device
  - Primary identifier: `uuid` (computed, BCM-assigned)
  - Required for create: `name`, `type`
  - Type classification via `child_type`: NetworkPhysicalInterface, NetworkBondInterface, NetworkBMCInterface
  - Network assignment: `network` (UUID reference to bcm_cmnet_network)
  - Physical attributes: `mac`, `ip`, `ipv6_ip`, `cardtype`
  - Configuration flags: `dhcp`, `bootable`, `start_if`
  - Bond-specific: `members` (list of interface names), `bond_mode`

- **Device (enhanced)**: Existing device resource with new interfaces capability
  - Maintains existing fields: `hostname`, `mac`, `category`, `management_network`, etc.
  - New field: `interfaces` (list of interface configurations)
  - Backward compatible: Legacy single-interface behavior preserved when interfaces block absent

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Administrators can configure devices with multiple network interfaces in a single Terraform resource
- **SC-002**: Administrators can create bonded interfaces with member assignment and mode configuration
- **SC-003**: Administrators can configure BMC/IPMI interfaces for out-of-band management
- **SC-004**: Existing device configurations without interfaces block continue to work unchanged (backward compatibility)
- **SC-005**: Devices with interfaces can be imported into Terraform with full interface state population
- **SC-006**: All acceptance tests pass with 100% success rate including interface CRUD operations
- **SC-007**: Example configurations validate successfully with `terraform validate`
- **SC-008**: Documentation is auto-generated and accurately describes the interfaces block and all attributes
- **SC-009**: Implementation follows TDD workflow (RED-GREEN-REFACTOR) with tests written before implementation
- **SC-010**: Tests are environment-portable with no hardcoded values or assumptions about cluster state

## Assumptions

- BCM API addDevice/updateDevice accepts nested interfaces array in device entity
- BCM API getDevice returns interfaces array with all interface attributes
- BCM validates interface configurations server-side (duplicate names, bond member validity, network existence)
- Interface UUIDs are assigned by BCM during create, not generated by provider
- Bond mode values follow standard Linux bonding modes supported by BCM
- The existing provisioning_interface field in BCM API can reference any bootable interface UUID
- BCM supports atomic interface updates - updating one interface doesn't affect others

## Dependencies

- Existing BCM API client (internal/provider/bcm_client.go) with CallJSONRPC support
- Existing CMDeviceDeviceResource implementation (internal/provider/resource_cmdevice_device.go)
- Existing NetworkInterfaceModel struct in data_source_cmdevice_nodes.go as reference
- Terraform Plugin Framework v1.16.1+ nested block schema support
- Terraform Plugin Testing v1.13.3+ for modern test patterns
- BCM cluster accessible at configured endpoint for acceptance testing
- Environment variables (BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD) for test configuration

## Out of Scope

- Network resource creation (networks must exist before interface assignment)
- Category or partition management (handled by existing resources)
- Interface performance monitoring or traffic statistics
- Real-time interface status (up/down) reporting
- VLAN tagging configuration (may be added in future enhancement)
- MTU configuration (may be added in future enhancement)
- Interface hardware discovery or auto-configuration

## API Reference

### BCM API Interface Entity Structure

Based on existing code analysis and BCM API patterns:

```json
{
  "baseType": "NetworkInterface",
  "childType": "NetworkPhysicalInterface",
  "uuid": "00000000-0000-0000-0000-000000000001",
  "name": "eth0",
  "mac": "00:11:22:33:44:55",
  "network": "uuid-of-network",
  "ip": "10.0.0.10",
  "ipv6Ip": "::0",
  "dhcp": true,
  "ipv6Dhcp": false,
  "bootable": true,
  "startIf": "ALWAYS",
  "cardtype": "Ethernet",
  "bringupduringinstall": "NO",
  "modified": true,
  "to_be_removed": false,
  "revision": ""
}
```

### Bond Interface Additional Fields

```json
{
  "childType": "NetworkBondInterface",
  "bondMode": "802.3ad",
  "members": ["eth0", "eth1"]
}
```

### Terraform Schema Example

```hcl
resource "bcm_cmdevice_device" "node" {
  hostname = "knode-01"
  category = data.bcm_cmdevice_categories.default.categories[0].uuid

  interfaces {
    name      = "bond0"
    type      = "bond"
    members   = ["eth0", "eth1"]
    network   = data.bcm_cmnet_networks.mgmt.networks[0].uuid
    bootable  = true
  }

  interfaces {
    name    = "eth2"
    type    = "physical"
    mac     = "00:11:22:33:44:57"
    network = data.bcm_cmnet_networks.data.networks[0].uuid
  }

  interfaces {
    name    = "ipmi"
    type    = "bmc"
    network = data.bcm_cmnet_networks.ipmi.networks[0].uuid
    ip      = "10.0.100.10"
    dhcp    = false
  }
}
```
