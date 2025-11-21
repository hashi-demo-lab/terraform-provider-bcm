# BCM Device Resource Research

**Date**: 2025-11-22
**Feature**: bcm_cmdevice_device resource

## API Method Signatures

Based on existing BCM client implementation and DeviceEntity.md documentation:

### Confirmed API Methods

| Method | Service | Args | Description |
|--------|---------|------|-------------|
| `addDevice` | cmdevice | `[device_entity, force]` | Create new device |
| `getDevice` | cmdevice | `[identifier]` | Get single device by hostname or UUID (efficient direct lookup) |
| `getNode` | cmdevice | `[identifier]` | Alias for getDevice (legacy naming) |
| `updateDevice` | cmdevice | `[device_entity, force]` | Update existing device |
| `removeDevice` | cmdevice | `[uuid, force]` | Delete device |
| `getNodes` | cmdevice | `[]` | List all devices (for data source, not resource Read) |

**Args Parameter Support**: CONFIRMED - bcm_client.go supports variadic args parameter (line 147-150)

## Device Type (childType) Assignment

Based on DeviceEntity.md example response:

**Device Types (childType values)**:
- `HeadNode` - Cluster management node
- `ComputeNode` - Compute worker node
- `PhysicalNode` - Generic physical node
- `CloudNode` - Cloud/virtual node
- `StorageNode` - Storage server

**Assignment Logic**: BCM determines childType based on role assignments. When creating a device:
- Device initially created with empty childType
- BCM assigns childType after role assignment
- HeadNodeRole → HeadNode
- ComputeRole → ComputeNode
- StorageRole → StorageNode (if primary role)

**Implementation**: childType is Computed field, not user-settable

## Force Parameter Scenarios

Based on existing resource patterns (cmpart_softwareimage.go, cmdevice_category.go):

### Create Operations
- `force=false` (default): BCM performs validation checks
- `force=true`: Override validation warnings

### Update Operations
- Device has active provisioning → requires `force=true`
- Device configuration locked → requires `force=true`
- Breaking changes (e.g., management network change) → may require `force=true`

### Delete Operations
- Device has active workloads → requires `force=true`
- Device has dependencies (monitoring, services) → requires `force=true`
- Device is head node → requires `force=true`

**Implementation**: Add optional `force` boolean field (default: false) with clear error messages when force is required

## Network Interface Configuration Patterns

From DeviceEntity.md NetworkInterface schema:

### Interface Types
- `NetworkPhysicalInterface` - Physical NIC
- `NetworkBondInterface` - Bonded interfaces
- `NetworkBridgeInterface` - Bridge interface
- `NetworkVlanInterface` - VLAN interface

### Required Fields (Physical Interface)
- `name` - Interface name (e.g., "ens33")
- `mac` - MAC address
- `network` - Network UUID reference

### Optional Fields
- `ip` - Static IPv4 address
- `ipv6Ip` - Static IPv6 address
- `dhcp` - Enable DHCP (boolean)
- `bootable` - PXE bootable (boolean)
- `cardtype` - "Ethernet" or "InfiniBand"

### Ordering
- No specific ordering required - BCM handles internally
- Primary interface typically listed first by convention

**MVP Scope**: Support physical interfaces only (defer bonding/bridging/VLANs to Phase 2)

## Validation Implementations

### RFC 1123 Hostname Validation

Pattern: `^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`

Rules:
- Lowercase letters and digits only
- Hyphens allowed in middle positions
- Must start and end with alphanumeric
- Maximum 63 characters
- Minimum 1 character

**Implementation**:
```go
stringvalidator.RegexMatches(
    regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`),
    "hostname must be RFC 1123 DNS label (lowercase alphanumeric and hyphens, 1-63 chars)",
)
```

### MAC Address Validation

Pattern: `^([0-9A-Fa-f]{2}:){5}[0-9A-Fa-f]{2}$`

Rules:
- Six groups of two hexadecimal digits
- Separated by colons
- Case insensitive

**Implementation**:
```go
stringvalidator.RegexMatches(
    regexp.MustCompile(`^([0-9A-Fa-f]{2}:){5}[0-9A-Fa-f]{2}$`),
    "mac must be six groups of two hexadecimal digits separated by colons (e.g., 00:11:22:33:44:55)",
)
```

### UUID Validation (RFC 4122)

Pattern: `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`

Rules:
- Lowercase hexadecimal
- Five groups separated by hyphens
- Group lengths: 8-4-4-4-12

**Implementation**:
```go
stringvalidator.RegexMatches(
    regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`),
    "must be valid UUID (RFC 4122)",
)
```

## Field Name Mappings (Terraform → BCM API)

Following established snake_case → camelCase pattern:

| Terraform Schema Field | BCM API Field |
|------------------------|---------------|
| `hostname` | `hostname` |
| `mac` | `mac` |
| `category` | `category` |
| `management_network` | `managementNetwork` |
| `kernel_parameters` | `kernelParameters` |
| `boot_loader` | `bootLoader` |
| `boot_loader_protocol` | `bootLoaderProtocol` |
| `power_control` | `powerControl` |
| `bmc_settings` | `bmcSettings` |
| `creation_time` | `creationTime` |
| `default_gateway` | `defaultGateway` |
| `install_boot_record` | `installBootRecord` |

## Error Handling Patterns

Based on existing resource implementations:

### Common BCM API Errors
1. **Duplicate Hostname**: `"Device with hostname 'X' already exists"`
2. **Invalid UUID Reference**: `"Category/Network UUID not found"`
3. **Force Required**: `"Device has active provisioning operations"`
4. **Validation Failure**: `"Invalid MAC address format"`

### Terraform Diagnostic Messages
- Use clear, actionable error messages
- Suggest remediation (e.g., "Set force = true to override")
- Include context (resource name, field name, value)

## Async Operation Behavior

Based on cmpart_softwareimage.go pattern:

**Device Creation**: Likely synchronous (no evidence of async operations)
**Device Updates**: Synchronous
**Device Deletion**: Synchronous (with force parameter if needed)

**Implementation**: No polling required initially. If async behavior discovered during testing, implement exponential backoff polling pattern (similar to software image cloning)

## Test Patterns

Based on existing tests (resource_cmpart_softwareimage_test.go, resource_cmdevice_category_test.go):

### Required Test Helpers
1. `createTestBCMClient(t)` - Already exists in test_helpers.go
2. `verifyResourceDeleted(ctx, client, "cmdevice", "getDevice", id, retries)` - Already exists
3. `generateUniqueTestName(prefix)` - Already exists
4. `getResourceUUIDByName(t, "cmdevice", "getDevice", name)` - Already exists

### Modern Testing Patterns (v1.13.3+)
- `statecheck.ExpectKnownValue()` for type-safe state checks
- `plancheck.ExpectEmptyPlan()` for idempotency verification
- `compare.ValuesSame()` for ID consistency tracking
- `plancheck.ExpectNonEmptyPlan()` for drift detection

### Required Test Coverage
- ✅ Basic CRUD (Create, Read, Update, Delete)
- ✅ Import functionality
- ✅ Idempotency checks (after Create and Update)
- ✅ Drift detection (external BCM API modifications)
- ✅ Validation errors (invalid hostname, MAC, UUID)
- ✅ CheckDestroy verification

## Implementation Notes

1. **Use existing BCM client** - No modifications needed to bcm_client.go
2. **Follow established patterns** - CMDeviceCategory and CMPartSoftwareImage serve as excellent templates
3. **MVP Focus** - Start with required fields only, add optional fields in REFACTOR phase
4. **Test-first approach** - Write failing tests before implementation
5. **Validation early** - Schema validators catch errors before API calls
6. **Clear error messages** - Help users understand BCM API constraints

## Phase 0 Complete

All research tasks complete. Ready to proceed to Phase 1 (Design Artifacts).
