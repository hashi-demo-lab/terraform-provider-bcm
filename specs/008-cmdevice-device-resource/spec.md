# Feature Specification: BCM Device Resource Management

**Feature Branch**: `008-cmdevice-device-resource`
**Created**: 2025-11-22
**Status**: Draft
**Input**: User description: "Add a new resource for managing BCM devices with the following API contract: Service: cmdevice, Create method: addDevice with args: [device, force]"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create and Manage Individual Cluster Nodes (Priority: P1)

As a cluster administrator, I want to create and manage individual devices (nodes) in my BCM cluster so that I can programmatically provision compute infrastructure using Terraform.

**Why this priority**: This is the core value proposition - allowing users to define cluster topology as Infrastructure as Code. Without this, users cannot create devices via Terraform.

**Independent Test**: Can be fully tested by creating a single device with minimal required fields (hostname, MAC address, category, management network), verifying it appears in BCM, and cleaning it up. Delivers immediate value by enabling programmatic device creation.

**Acceptance Scenarios**:

1. **Given** a valid category UUID and management network UUID exist in BCM, **When** I create a device resource with hostname "test-node001" and MAC address "00:11:22:33:44:55", **Then** the device is created in BCM with a UUID and appears in the cluster inventory
2. **Given** a device resource exists in Terraform state, **When** I modify the device's notes field, **Then** Terraform updates the device in BCM and the change is reflected
3. **Given** a device resource exists in Terraform state, **When** the device is modified externally in BCM (e.g., hostname changed), **Then** Terraform detects drift and can restore the desired state
4. **Given** a device resource exists in Terraform state, **When** I run terraform destroy, **Then** the device is removed from BCM cluster

---

### User Story 2 - Import Existing Devices into Terraform Management (Priority: P2)

As a cluster administrator, I want to import existing devices from my BCM cluster into Terraform state so that I can bring brownfield infrastructure under Infrastructure as Code management.

**Why this priority**: Essential for adoption in existing clusters. Most users have pre-existing devices that need to be imported before Terraform can manage them.

**Independent Test**: Can be tested by manually creating a device in BCM, running `terraform import bcm_cmdevice_device.test <uuid>`, and verifying all fields are correctly populated in state.

**Acceptance Scenarios**:

1. **Given** a device exists in BCM with UUID "abc-123", **When** I run `terraform import bcm_cmdevice_device.test abc-123`, **Then** the device is imported into Terraform state with all attributes populated correctly
2. **Given** an imported device in Terraform state, **When** I run terraform plan with no changes to the configuration, **Then** Terraform reports no changes are needed (idempotency verification)

---

### User Story 3 - Configure Network Interfaces and Roles (Priority: P3)

As a cluster administrator, I want to configure network interfaces and role assignments for devices so that nodes are properly integrated into the cluster networking and service topology.

**Why this priority**: Advanced configuration that builds on basic device creation. Required for production clusters but can be added incrementally after basic CRUD operations work.

**Independent Test**: Can be tested by creating a device with multiple network interfaces configured (management + data network) and verifying the interface configurations are applied correctly in BCM.

**Acceptance Scenarios**:

1. **Given** a device resource, **When** I configure multiple network interfaces with different IP addresses and network references, **Then** the device is created with all interfaces properly configured
2. **Given** a device resource, **When** I assign roles (e.g., ComputeRole, MonitoringRole), **Then** the device is created with those roles assigned and associated services are configured

---

### Edge Cases

- What happens when attempting to create a device with a duplicate hostname? (BCM should reject, Terraform should report error)
- What happens when attempting to create a device with an invalid MAC address format? (Schema validation should catch this before API call)
- What happens when a device is deleted externally in BCM while in Terraform state? (Read operation should detect deletion and remove from state)
- What happens when updating a device that has pending provisioning operations? (BCM may require force=true parameter)
- What happens when attempting to delete a device that has active workloads or dependencies? (BCM may reject, Terraform should provide clear error message)
- What happens when network interface references point to non-existent network UUIDs? (BCM API validation should reject, Terraform should report error)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST allow creating devices with required fields: hostname, MAC address, category UUID, management network UUID
- **FR-002**: System MUST support optional device configuration including: network interfaces, roles, boot settings, kernel parameters, power control settings, BMC configuration
- **FR-003**: System MUST validate hostname format (RFC 1123 DNS label: lowercase alphanumeric and hyphens, start/end with alphanumeric)
- **FR-004**: System MUST validate MAC address format (six groups of two hexadecimal digits separated by colons)
- **FR-005**: System MUST validate UUID references (RFC 4122 format) for category, management network, and other UUID-based references
- **FR-006**: System MUST support import functionality using device UUID as identifier
- **FR-007**: System MUST detect external modifications to devices (drift detection) during Read operations
- **FR-008**: System MUST support force parameter for operations that may conflict with BCM constraints
- **FR-009**: System MUST handle device type variants (HeadNode, PhysicalNode, ComputeNode, CloudNode, StorageNode) as computed childType field
- **FR-010**: System MUST preserve BCM-assigned computed values: UUID, creation_time, modified, to_be_removed, revision
- **FR-011**: System MUST support updating device configuration fields while preserving UUID and identity
- **FR-012**: System MUST support deletion with proper cleanup (force parameter may be required if device has dependencies)

### API Contract

**Service**: `cmdevice`

**CRUD Operations**:

| Operation | API Method | Args Parameter | Description |
|-----------|------------|----------------|-------------|
| Create | `addDevice` | `[device_entity, force]` | Creates new device in BCM cluster |
| Read | `getDevice` | `[identifier]` | Retrieves single device by hostname or UUID (efficient direct lookup) |
| Update | `updateDevice` | `[device_entity, force]` | Updates existing device configuration |
| Delete | `removeDevice` | `[uuid, force]` | Removes device from BCM cluster |
| Validate | `validateDevice` | `[device_entity]` | Validates device configuration before create/update (optional pre-flight check) |
| List | `getNodes` | `[]` | Lists all devices (used for data source, not resource Read) |

**Force Parameter**:
- Type: boolean
- Default: false
- Purpose: Override BCM validation warnings and constraints (e.g., updating devices with active provisioning, deleting devices with dependencies)
- Usage: Required for certain operations that conflict with BCM's safety checks

### Key Entities

- **Device**: Represents a physical or virtual node in the BCM cluster
  - **Core Identity**: hostname (string, unique), uuid (string, BCM-assigned), mac (string, primary MAC address)
  - **Type Classification**: baseType (always "Device"), childType (HeadNode | PhysicalNode | ComputeNode | CloudNode | StorageNode)
  - **Relationships**: category (UUID reference), managementNetwork (UUID reference), partition (UUID reference, optional)
  - **Network Configuration**: interfaces (array of NetworkInterface objects), provisioningInterface (UUID reference)
  - **Role Assignments**: roles (array of Role objects defining node function in cluster)
  - **Boot Configuration**: bootLoader (string), bootLoaderProtocol (string), installMode (string), kernelParameters (string)
  - **Hardware Management**: bmcSettings (BMC/IPMI configuration), powerControl (power management method), gpuSettings (GPU configuration)
  - **Storage Configuration**: fsmounts (filesystem mounts), fsexports (NFS exports)
  - **Metadata**: creationTime (timestamp), modified (boolean), revision (string)

- **NetworkInterface**: Represents a network interface on a device
  - **Core Attributes**: name (string, e.g., "ens33"), mac (string), ip (IPv4 address), ipv6Ip (IPv6 address)
  - **Network Reference**: network (UUID reference to Network entity)
  - **Configuration**: dhcp (boolean), bootable (boolean), startIf (when to activate), cardtype (Ethernet | InfiniBand)
  - **Types**: NetworkPhysicalInterface, NetworkBondInterface, NetworkBridgeInterface, NetworkVlanInterface

- **Role**: Defines services and functions assigned to a device
  - **Core Attributes**: name (string), uuid (string), addServices (boolean)
  - **Common Types**: HeadNodeRole, ComputeRole, StorageRole, MonitoringRole, ProvisioningRole, BootRole

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can create a basic device resource with minimal configuration (hostname, MAC, category, management network) in a single terraform apply operation
- **SC-002**: Users can import existing BCM devices into Terraform state using device UUID, with all fields correctly populated on first import
- **SC-003**: Terraform detects configuration drift when devices are modified externally in BCM (e.g., hostname, notes, kernel parameters changed)
- **SC-004**: Resource supports full CRUD lifecycle: Create → Read → Update → Read → Delete with proper state management
- **SC-005**: Acceptance tests pass at 100% success rate covering: basic create, import, update, delete, drift detection scenarios
- **SC-006**: Documentation is auto-generated and includes working examples for common device configurations (compute node, head node, storage node)
- **SC-007**: Schema validation prevents invalid configurations (malformed MAC addresses, invalid UUIDs, etc.) before API calls

## Assumptions *(optional)*

- BCM cluster is accessible at configured endpoint with valid credentials
- Category and management network resources already exist in BCM (or are created separately)
- Device types (childType) are determined by BCM based on role assignments and configuration, not directly settable by users
- Network interface configuration follows BCM conventions (physical interfaces discovered, logical interfaces explicitly configured)
- Power control methods (IPMI, Redfish, PDU) are configured out-of-band and referenced by UUID
- Provisioning operations (PXE boot, OS installation) are handled by separate BCM workflows, not directly by this resource

## Dependencies *(optional)*

**Upstream Dependencies**:
- BCM cluster running and accessible
- Valid authentication credentials (username/password)
- Existing category resource (bcm_cmdevice_category) for device categorization
- Existing network resource (bcm_cmnet_network) for management network reference

**Downstream Dependencies**:
- Device resources may be referenced by other resources (e.g., monitoring configurations, workload scheduling)
- Changes to devices may trigger BCM provisioning workflows

## Open Questions *(optional)*

1. **Device Type Selection**: How does BCM determine childType (HeadNode vs ComputeNode vs PhysicalNode)? Is it based on roles assigned, or is there an explicit field?
   - *Research needed*: Examine existing devices in BCM to understand childType assignment logic
   - *Assumption for now*: childType is computed by BCM based on role assignments and configuration

2. **Network Interface Ordering**: Does BCM require interfaces to be specified in a particular order, or does it handle ordering automatically?
   - *Assumption for now*: Interfaces can be specified in any order; BCM handles internal ordering

3. **Async Operations**: Are device creation and updates synchronous, or do some operations (like OS provisioning) require polling for completion?
   - *Assumption for now*: Device creation/update is synchronous; provisioning is a separate workflow
   - *Fallback*: Implement exponential backoff polling if needed (similar to software image cloning pattern)

4. **Force Parameter Scenarios**: What specific scenarios require force=true for device operations?
   - *Research needed*: Test device updates and deletions to identify scenarios requiring force
   - *Assumption for now*: Force may be needed for deleting devices with active workloads or updating provisioning-locked devices

## Test Pattern Requirements *(mandatory)*

All acceptance tests MUST follow the established project patterns:

### Required Test Structure

```go
func TestAccCMDeviceDeviceResource_Basic(t *testing.T) {
    deviceName := generateUniqueTestName("test-device")

    // ID consistency tracking across all CRUD operations
    compareID := statecheck.CompareValue(compare.ValuesSame())

    resource.Test(t, resource.TestCase{
        PreCheck: func() {
            testAccCMDeviceDevicePreCheck(t, deviceName)
        },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
        Steps: []resource.TestStep{
            // Create and Read testing
            {
                Config: testAccCMDeviceDeviceResourceConfig_Basic(deviceName),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "hostname", deviceName),
                    resource.TestCheckResourceAttrSet("bcm_cmdevice_device.test", "uuid"),
                ),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "bcm_cmdevice_device.test",
                        tfjsonpath.New("hostname"),
                        knownvalue.StringExact(deviceName),
                    ),
                    statecheck.ExpectKnownValue(
                        "bcm_cmdevice_device.test",
                        tfjsonpath.New("uuid"),
                        knownvalue.NotNull(),
                    ),
                    compareID.AddStateValue(
                        "bcm_cmdevice_device.test",
                        tfjsonpath.New("id"),
                    ),
                },
            },
            // Idempotency check after Create
            {
                Config: testAccCMDeviceDeviceResourceConfig_Basic(deviceName),
                ConfigPlanChecks: resource.ConfigPlanChecks{
                    PreApply: []plancheck.PlanCheck{
                        plancheck.ExpectEmptyPlan(),
                    },
                },
            },
            // ImportState testing
            {
                ResourceName:      "bcm_cmdevice_device.test",
                ImportState:       true,
                ImportStateVerify: true,
                ConfigStateChecks: []statecheck.StateCheck{
                    compareID.AddStateValue(
                        "bcm_cmdevice_device.test",
                        tfjsonpath.New("id"),
                    ),
                },
            },
            // Update and Read testing
            {
                Config: testAccCMDeviceDeviceResourceConfig_Updated(deviceName),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "hostname", deviceName),
                ),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "bcm_cmdevice_device.test",
                        tfjsonpath.New("notes"),
                        knownvalue.StringExact("Updated notes"),
                    ),
                    compareID.AddStateValue(
                        "bcm_cmdevice_device.test",
                        tfjsonpath.New("id"),
                    ),
                },
            },
            // Idempotency check after Update
            {
                Config: testAccCMDeviceDeviceResourceConfig_Updated(deviceName),
                ConfigPlanChecks: resource.ConfigPlanChecks{
                    PreApply: []plancheck.PlanCheck{
                        plancheck.ExpectEmptyPlan(),
                    },
                },
            },
        },
    })
}
```

### Required Test Helpers

1. **testAccCMDeviceDevicePreCheck(t, names...)** - Clean up leftover test devices using `verifyResourceDeleted()` helper with 5 retries
2. **testAccCheckCMDeviceDeviceDestroy(s)** - Verify all devices deleted after test, use `createTestBCMClient()` and check all resources
3. **testAccCMDeviceDeviceResourceConfig_Basic(name)** - Return Terraform config with provider block using environment variables
4. **testAccCMDeviceDeviceResourceConfig_Updated(name)** - Return updated config (e.g., modified notes)

### Required Drift Detection Test

```go
func TestAccCMDeviceDevice_DriftHostname(t *testing.T) {
    deviceName := generateUniqueTestName("test-device")

    resource.Test(t, resource.TestCase{
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
        Steps: []resource.TestStep{
            // Create with initial value
            {
                Config: testAccCMDeviceDeviceResourceConfig_Drift(deviceName, "initial-value"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "notes", "initial-value"),
                ),
            },
            // Modify externally via BCM API
            {
                PreConfig: func() {
                    client := createTestBCMClient(t)
                    ctx := context.Background()

                    uuid := getResourceUUIDByName(t, "cmdevice", "getDevice", deviceName)

                    body, _ := client.CallJSONRPC(ctx, "cmdevice", "getDevice", deviceName)
                    var deviceData map[string]interface{}
                    json.Unmarshal(body, &deviceData)

                    // Modify field externally
                    deviceData["notes"] = "externally-modified"

                    // Build BCM entity structure
                    entity := map[string]interface{}{
                        "baseType":      "Device",
                        "childType":     deviceData["childType"],
                        "modified":      true,
                        "to_be_removed": false,
                        "uuid":          uuid,
                    }
                    for k, v := range deviceData {
                        if k != "uuid" {
                            entity[k] = v
                        }
                    }

                    client.CallJSONRPC(ctx, "cmdevice", "updateDevice", entity, false)
                    time.Sleep(2 * time.Second)
                },
                Config: testAccCMDeviceDeviceResourceConfig_Drift(deviceName, "initial-value"),
                ConfigPlanChecks: resource.ConfigPlanChecks{
                    PreApply: []plancheck.PlanCheck{
                        plancheck.ExpectNonEmptyPlan(),
                    },
                },
            },
            // Terraform restores desired state
            {
                Config: testAccCMDeviceDeviceResourceConfig_Drift(deviceName, "initial-value"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "notes", "initial-value"),
                ),
            },
        },
    })
}
```

### Required Test Imports

```go
import (
    "github.com/hashicorp/terraform-plugin-testing/helper/resource"
    "github.com/hashicorp/terraform-plugin-testing/plancheck"
    "github.com/hashicorp/terraform-plugin-testing/statecheck"
    "github.com/hashicorp/terraform-plugin-testing/knownvalue"
    "github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
    "github.com/hashicorp/terraform-plugin-testing/compare"
)
```

### Required Test Coverage

- ✅ Basic Create/Read with state checks
- ✅ Idempotency after Create (empty plan check)
- ✅ ImportState with ID consistency tracking
- ✅ Update/Read with state checks
- ✅ Idempotency after Update (empty plan check)
- ✅ Drift detection with PreConfig external modification
- ✅ CheckDestroy verification
- ✅ Validation error scenarios (invalid MAC, invalid hostname, etc.)

## Out of Scope *(optional)*

**Phase 1 (MVP)** - Not included in initial implementation:
- Advanced network interface types (bonding, bridging, VLANs) - only physical interfaces initially
- GPU configuration (gpuSettings) - deferred to Phase 2
- Custom power control scripts - only standard IPMI/Redfish methods
- Detailed BIOS/UEFI configuration - minimal BMC settings only
- Prometheus metric forwarders - monitoring configuration deferred
- Disk setup XML configuration - basic defaults only
- Provisioning script customization (initialize, finalize) - deferred

**Future Enhancements**:
- Bulk device creation with template-based configuration
- Device cloning from existing configurations
- Power control operations (power on/off, reboot) as separate resource or data source
- Integration with BCM provisioning workflows (trigger OS installation)
- Health monitoring and status tracking
- Advanced role-based configuration templates
