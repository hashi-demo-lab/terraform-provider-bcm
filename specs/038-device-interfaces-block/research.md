# Phase 0 Research: Device Interfaces Block

**Feature**: 038-device-interfaces-block
**Date**: 2025-11-25
**Status**: COMPLETE

## Research Summary

This document consolidates findings from API exploration, codebase analysis, and BCM documentation review to resolve all NEEDS CLARIFICATION items from the implementation plan.

---

## Research Task 1: BCM Interface API Structure

### Decision
Interface entities are embedded in the device `interfaces` array. Bond members are specified by interface **names** (not UUIDs). BMC interfaces use `childType: "NetworkBMCInterface"` with no special additional fields beyond standard interface attributes.

### Rationale
Analysis of existing code in `data_source_cmdevice_nodes.go` (lines 394-425) and `resource_cmdevice_device.go` (lines 977-1004) shows:
1. Interfaces are mapped using the `mapInterfaces()` function
2. BCM returns interfaces as `[]interface{}` that maps to `[]map[string]interface{}`
3. The existing `buildDeviceAPIEntity` creates a hardcoded single interface - we need to extend this

### Evidence

**From `data_source_cmdevice_nodes.go` (NetworkInterfaceModel):**
```go
type NetworkInterfaceModel struct {
    Name      types.String `tfsdk:"name"`
    MAC       types.String `tfsdk:"mac"`
    IP        types.String `tfsdk:"ip"`
    IPv6IP    types.String `tfsdk:"ipv6_ip"`
    DHCP      types.Bool   `tfsdk:"dhcp"`
    Network   types.String `tfsdk:"network"`
    BaseType  types.String `tfsdk:"base_type"`
    ChildType types.String `tfsdk:"child_type"`
    CardType  types.String `tfsdk:"cardtype"`
    Bootable  types.Bool   `tfsdk:"bootable"`
    StartIf   types.String `tfsdk:"start_if"`
}
```

**From `resource_cmdevice_device.go` (current hardcoded interface):**
```go
networkInterface := map[string]interface{}{
    "baseType":             "NetworkInterface",
    "childType":            "NetworkPhysicalInterface",
    "mac":                  plan.MAC.ValueString(),
    "network":              networkUUID,
    "name":                 "eth0",
    "dhcp":                 true,
    "bootable":             true,
    "startIf":              "ALWAYS",
    "modified":             true,
    "to_be_removed":        false,
    "revision":             "",
    "uuid":                 interfaceUUID,
    "ipv6Ip":               "::0",
    "ipv6Dhcp":             false,
    "bringupduringinstall": "NO",
    "cardtype":             "Ethernet",
}
```

**From BCM API Documentation (`CMDevice_Complete_Documentation.md`):**
```json
{
  "baseType": "NetworkInterface",
  "childType": "NetworkPhysicalInterface",
  "name": "ens33",
  "mac": "00:50:56:9B:E4:6D",
  "ip": "172.21.15.254",
  "ipv6Ip": "::0",
  "dhcp": false,
  "network": "<network_uuid>",
  "cardtype": "Ethernet",
  "bootable": false,
  "startIf": "ALWAYS"
}
```

**From Spec (`spec.md` lines 239-242) - Bond Interface Structure:**
```json
{
  "childType": "NetworkBondInterface",
  "bondMode": "802.3ad",
  "members": ["eth0", "eth1"]
}
```

### Alternatives Considered
1. **UUIDs for bond members**: Rejected - BCM uses interface names which is more user-friendly and consistent with Linux bonding
2. **Separate BMC resource**: Rejected - BMC is just another interface type, keeping unified simplifies the API

---

## Research Task 2: Interface Update Semantics

### Decision
BCM uses **full replacement** semantics for the interfaces array. When updating a device, the entire `interfaces` array must be sent, not just modified interfaces. Interface UUIDs are **client-generated** before create and must be preserved on update.

### Rationale
The existing `buildDeviceAPIEntity` function generates a UUID for the interface before the API call:
```go
interfaceUUID := "00000000-0000-0000-0000-000000000001" // Generate a temporary UUID
```

This pattern suggests:
1. BCM expects UUID in the request (not generated server-side)
2. The provider must track interface UUIDs in state
3. Updates require sending the same UUID to preserve identity

### Evidence

**From `resource_cmdevice_device.go` (lines 979, 999):**
```go
interfaceUUID := "00000000-0000-0000-0000-000000000001"
// ...
"uuid": interfaceUUID,
```

**From test patterns (`resource_cmdevice_device_idempotency_test.go` lines 303-324):**
```go
// Modify notes field externally (snake_case -> camelCase)
deviceData["notes"] = "Modified externally"

// Wrap in BCM API entity structure
entity := map[string]interface{}{
    "baseType":      "Node",
    "childType":     "",
    "modified":      true,
    "to_be_removed": false,
    "revision":      "",
    "uuid":          uuid,
}
```

### Implementation Implications
1. Generate UUIDs for new interfaces using `github.com/google/uuid`
2. Store interface UUIDs in state (computed attribute)
3. On update, preserve existing interface UUIDs from state
4. Match interfaces by name for update correlation

---

## Research Task 3: Validation Patterns

### Decision
Use BCM's `validateDevice` API for server-side validation. Interface-specific errors appear in the `validation` response array with `field` indicating the problematic interface attribute. Bond member validation is handled server-side by BCM.

### Rationale
The existing validation pattern in `resource_cmdevice_device.go` shows:
1. Pre-flight validation before Create/Update
2. Error messages include field names
3. Warnings are surfaced but don't block operations

### Evidence

**From `resource_cmdevice_device.go` (lines 386-416):**
```go
// Pre-flight validation: Call validateDevice before CREATE
validationErrors, err := r.client.ValidateEntity(ctx, "CMDevice", "validateDevice", deviceEntity, true)
if err != nil {
    resp.Diagnostics.AddError(
        "Validation API Error",
        fmt.Sprintf("Could not validate device '%s': %s", plan.Hostname.ValueString(), err.Error()),
    )
    return
}

// Process validation results
hasErrors := false
for _, valErr := range validationErrors {
    if valErr.IsError() {
        resp.Diagnostics.AddError(
            fmt.Sprintf("Validation Error: %s", valErr.Field),
            valErr.Message,
        )
        hasErrors = true
    } else if valErr.IsWarning() {
        resp.Diagnostics.AddWarning(
            fmt.Sprintf("Validation Warning: %s", valErr.Field),
            valErr.Message,
        )
    }
}
```

### Terraform-Side Validators Needed

Based on the spec requirements (FR-012, FR-013):

1. **Unique Interface Names** (FR-012):
```go
// Validator to ensure interface names are unique within device
stringvalidator.NoneOf(existingNames...)
```

2. **Bond Members Required** (FR-013):
```go
// Validator to ensure members specified when type is "bond"
objectvalidator.AtLeastOneOf(
    path.MatchRelative().AtParent().AtName("members"),
)
```

---

## Research Task 4: Interface childType Mapping

### Decision
The mapping between Terraform `type` attribute and BCM `childType` is:

| Terraform `type` | BCM `childType` | BCM `cardtype` |
|------------------|-----------------|----------------|
| `"physical"` | `"NetworkPhysicalInterface"` | `"Ethernet"` (or hardware-detected) |
| `"bond"` | `"NetworkBondInterface"` | `"Ethernet"` |
| `"bmc"` | `"NetworkBMCInterface"` | `"BMC"` |

### Rationale
This mapping provides a user-friendly abstraction over BCM's internal type system while maintaining full compatibility.

### Evidence

**From `CMDevice_Complete_Documentation.md` (line 219):**
```
| childType | Description |
|-----------|-------------|
| NetworkPhysicalInterface | Physical NIC |
| NetworkBondInterface | Bonded interfaces |
| NetworkBridgeInterface | Bridge interfaces |
| NetworkVlanInterface | VLAN interfaces |
```

**Note**: VLAN and Bridge types are out of scope for this feature (per spec "Out of Scope" section).

---

## Research Task 5: Interface Field Defaults

### Decision
BCM applies these defaults for interface fields:

| Field | Default Value | Notes |
|-------|---------------|-------|
| `dhcp` | `true` | DHCP enabled by default |
| `bootable` | `false` | Not PXE bootable by default |
| `startIf` | `"ALWAYS"` | Always bring up interface |
| `ipv6Dhcp` | `false` | IPv6 DHCP disabled |
| `ipv6Ip` | `"::0"` | Empty IPv6 address |
| `bringupduringinstall` | `"NO"` | Don't bring up during install |

### Rationale
These defaults match the existing implementation in `buildDeviceAPIEntity` and align with typical cluster configuration patterns.

### Implementation Note
For the `interfaces` block, we should:
1. Set `dhcp = true` as Terraform default
2. Set `bootable = false` as Terraform default
3. Set `start_if = "ALWAYS"` as Terraform default
4. Handle other fields as computed from BCM response

---

## Research Task 6: Provisioning Interface Handling

### Decision
The first bootable interface in the `interfaces` list becomes the provisioning interface. The `provisioningInterface` field in the device entity stores the UUID of this interface.

### Rationale
From the existing implementation:
```go
entity["provisioningInterface"] = interfaceUUID
```

The provisioning interface is referenced by UUID, and the current code sets it to the single hardcoded interface.

### Implementation
1. Iterate through interfaces in order
2. Find first interface with `bootable = true`
3. Set `provisioningInterface` to that interface's UUID
4. If no bootable interface, leave as zero UUID (BCM handles default)

---

## Consolidated API Contract

Based on research, here is the complete interface entity structure for BCM API:

```json
{
  "baseType": "NetworkInterface",
  "childType": "NetworkPhysicalInterface | NetworkBondInterface | NetworkBMCInterface",
  "uuid": "client-generated-uuid",
  "name": "eth0",
  "mac": "00:11:22:33:44:55",
  "network": "network-uuid",
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
  "revision": "",

  // Bond-specific fields (only for NetworkBondInterface)
  "bondMode": "802.3ad",
  "members": ["eth0", "eth1"]
}
```

---

## Research Gaps Remaining

None. All NEEDS CLARIFICATION items have been resolved through codebase analysis and documentation review.

---

## Next Steps

1. Proceed to Phase 1: Generate `data-model.md` with complete entity definitions
2. Generate `contracts/interfaces.json` with JSON Schema
3. Generate `quickstart.md` for developers
4. Update agent context files

---

## References

- `/workspace/internal/provider/resource_cmdevice_device.go` - Current device resource
- `/workspace/internal/provider/data_source_cmdevice_nodes.go` - NetworkInterfaceModel reference
- `/workspace/sampleRest/CMDevice_Complete_Documentation.md` - BCM API documentation
- `/workspace/specs/038-device-interfaces-block/spec.md` - Feature specification
