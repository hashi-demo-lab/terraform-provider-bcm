# Data Model: Device Interfaces Block

**Feature**: 038-device-interfaces-block
**Date**: 2025-11-25
**Status**: COMPLETE

## Overview

This document defines the data model for the `interfaces` block enhancement to the `bcm_cmdevice_device` resource. It includes Go struct definitions, Terraform schema specifications, and BCM API entity mappings.

---

## Entity Definitions

### DeviceInterfaceModel (New Terraform Model)

```go
// DeviceInterfaceModel represents a network interface configuration within a device resource.
// This model supports physical interfaces, bonded interfaces, and BMC interfaces.
type DeviceInterfaceModel struct {
    // ===== Identity Fields (User-Provided) =====

    // Name is the interface name (e.g., "eth0", "bond0", "ipmi").
    // Required. Must be unique within the device.
    Name types.String `tfsdk:"name"`

    // Type specifies the interface type: "physical", "bond", or "bmc".
    // Required. Maps to BCM childType.
    Type types.String `tfsdk:"type"`

    // ===== Network Assignment =====

    // Network is the UUID reference to a bcm_cmnet_network resource.
    // Optional. When not specified, interface is not assigned to a network.
    Network types.String `tfsdk:"network"`

    // MAC is the interface MAC address (format: 00:11:22:33:44:55).
    // Optional for create (BCM may auto-detect), but recommended for physical interfaces.
    MAC types.String `tfsdk:"mac"`

    // IP is the static IPv4 address for the interface.
    // Optional. When DHCP is enabled, this is ignored.
    IP types.String `tfsdk:"ip"`

    // IPv6IP is the static IPv6 address for the interface.
    // Optional.
    IPv6IP types.String `tfsdk:"ipv6_ip"`

    // ===== Configuration Flags =====

    // DHCP enables DHCP for IP address assignment.
    // Optional. Default: true.
    DHCP types.Bool `tfsdk:"dhcp"`

    // Bootable indicates if this interface supports PXE boot.
    // Optional. Default: false. The first bootable interface becomes provisioningInterface.
    Bootable types.Bool `tfsdk:"bootable"`

    // StartIf specifies when to bring up the interface.
    // Optional. Default: "ALWAYS". Valid values: "ALWAYS", "NEVER", "HOTPLUG".
    StartIf types.String `tfsdk:"start_if"`

    // ===== Bond-Specific Fields =====

    // Members lists the member interface names for bond type.
    // Required when type is "bond". List of strings (interface names, not UUIDs).
    Members types.List `tfsdk:"members"` // Element type: types.StringType

    // BondMode specifies the bonding mode.
    // Optional. Valid values: "802.3ad", "active-backup", "balance-rr", "balance-xor", etc.
    BondMode types.String `tfsdk:"bond_mode"`

    // ===== Computed Fields (From BCM API) =====

    // UUID is the BCM-assigned interface identifier.
    // Computed. Client-generated before create, preserved on update.
    UUID types.String `tfsdk:"uuid"`

    // BaseType is the BCM entity base type.
    // Computed. Always "NetworkInterface".
    BaseType types.String `tfsdk:"base_type"`

    // ChildType is the BCM-determined interface type.
    // Computed. Maps from Type: physical->NetworkPhysicalInterface, bond->NetworkBondInterface, bmc->NetworkBMCInterface.
    ChildType types.String `tfsdk:"child_type"`

    // CardType is the hardware card type.
    // Computed. Values: "Ethernet", "InfiniBand", "BMC".
    CardType types.String `tfsdk:"cardtype"`
}
```

### Updated CMDeviceDeviceResourceModel

```go
// CMDeviceDeviceResourceModel describes the device resource data model.
// Enhanced with interfaces block support while maintaining backward compatibility.
type CMDeviceDeviceResourceModel struct {
    // ===== Existing Fields (Unchanged) =====

    // Identity fields (required/computed)
    ID       types.String `tfsdk:"id"`       // Computed, same as UUID
    UUID     types.String `tfsdk:"uuid"`     // Computed, BCM-assigned
    Hostname types.String `tfsdk:"hostname"` // Required, RFC 1123 validation
    MAC      types.String `tfsdk:"mac"`      // Required (or interfaces block), MAC address validation

    // References (required)
    Category          types.String `tfsdk:"category"`           // Required, UUID reference
    ManagementNetwork types.String `tfsdk:"management_network"` // Optional+Computed, UUID reference
    Partition         types.String `tfsdk:"partition"`          // Optional+Computed, UUID reference

    // Optional configuration
    Notes              types.String `tfsdk:"notes"`
    KernelParameters   types.String `tfsdk:"kernel_parameters"`
    BootLoader         types.String `tfsdk:"boot_loader"`
    BootLoaderProtocol types.String `tfsdk:"boot_loader_protocol"`
    Force              types.Bool   `tfsdk:"force"`

    // Power control configuration
    PowerControl types.String `tfsdk:"power_control"`

    // Network gateway configuration
    DefaultGateway       types.String `tfsdk:"default_gateway"`
    DefaultGatewayMetric types.Int64  `tfsdk:"default_gateway_metric"`

    // Hardware identifiers
    SerialNumber types.String `tfsdk:"serial_number"`
    PartNumber   types.String `tfsdk:"part_number"`

    // Computed fields
    CreationTime types.Int64  `tfsdk:"creation_time"`
    BaseType     types.String `tfsdk:"base_type"`
    ChildType    types.String `tfsdk:"child_type"`

    // ===== NEW: Interfaces Block =====

    // Interfaces is a list of network interface configurations.
    // Optional. When specified, takes precedence over legacy mac/management_network.
    Interfaces []DeviceInterfaceModel `tfsdk:"interfaces"`
}
```

---

## Terraform Schema Definition

### Interfaces Block Schema

```go
// InterfacesBlockSchema returns the schema for the interfaces nested block.
func InterfacesBlockSchema() schema.ListNestedBlock {
    return schema.ListNestedBlock{
        MarkdownDescription: "Network interface configurations for the device. " +
            "When specified, provides full control over interface setup. " +
            "Each interface can be physical, bond, or BMC type.",
        NestedObject: schema.NestedBlockObject{
            Attributes: map[string]schema.Attribute{
                // Required attributes
                "name": schema.StringAttribute{
                    Required:            true,
                    MarkdownDescription: "Interface name (e.g., 'eth0', 'bond0', 'ipmi'). Must be unique within the device.",
                    Validators: []validator.String{
                        stringvalidator.LengthBetween(1, 63),
                        stringvalidator.RegexMatches(
                            regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`),
                            "must start with letter, contain only alphanumeric, underscore, or hyphen",
                        ),
                    },
                },
                "type": schema.StringAttribute{
                    Required:            true,
                    MarkdownDescription: "Interface type: 'physical', 'bond', or 'bmc'.",
                    Validators: []validator.String{
                        stringvalidator.OneOf("physical", "bond", "bmc"),
                    },
                },

                // Optional network attributes
                "network": schema.StringAttribute{
                    Optional:            true,
                    MarkdownDescription: "Network UUID reference for interface assignment.",
                    Validators: []validator.String{
                        stringvalidator.RegexMatches(
                            regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`),
                            "must be valid UUID (RFC 4122)",
                        ),
                    },
                },
                "mac": schema.StringAttribute{
                    Optional:            true,
                    MarkdownDescription: "MAC address (format: 00:11:22:33:44:55). Required for physical interfaces on create.",
                    Validators: []validator.String{
                        stringvalidator.RegexMatches(
                            regexp.MustCompile(`^([0-9A-Fa-f]{2}:){5}[0-9A-Fa-f]{2}$`),
                            "must be six groups of two hexadecimal digits separated by colons",
                        ),
                    },
                },
                "ip": schema.StringAttribute{
                    Optional:            true,
                    MarkdownDescription: "Static IPv4 address.",
                    Validators: []validator.String{
                        stringvalidator.RegexMatches(
                            regexp.MustCompile(`^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`),
                            "must be a valid IPv4 address",
                        ),
                    },
                },
                "ipv6_ip": schema.StringAttribute{
                    Optional:            true,
                    MarkdownDescription: "Static IPv6 address.",
                },

                // Configuration flags
                "dhcp": schema.BoolAttribute{
                    Optional:            true,
                    Computed:            true,
                    MarkdownDescription: "Enable DHCP for IP assignment. Default: true.",
                },
                "bootable": schema.BoolAttribute{
                    Optional:            true,
                    Computed:            true,
                    MarkdownDescription: "Enable PXE boot capability. Default: false. First bootable interface becomes provisioning interface.",
                },
                "start_if": schema.StringAttribute{
                    Optional:            true,
                    Computed:            true,
                    MarkdownDescription: "Interface startup condition: 'ALWAYS', 'NEVER', 'HOTPLUG'. Default: 'ALWAYS'.",
                    Validators: []validator.String{
                        stringvalidator.OneOf("ALWAYS", "NEVER", "HOTPLUG"),
                    },
                },

                // Bond-specific
                "members": schema.ListAttribute{
                    Optional:            true,
                    ElementType:         types.StringType,
                    MarkdownDescription: "Member interface names for bond type. Required when type is 'bond'.",
                },
                "bond_mode": schema.StringAttribute{
                    Optional:            true,
                    MarkdownDescription: "Bond mode (e.g., '802.3ad', 'active-backup', 'balance-rr'). Only applicable when type is 'bond'.",
                    Validators: []validator.String{
                        stringvalidator.OneOf(
                            "802.3ad",
                            "active-backup",
                            "balance-rr",
                            "balance-xor",
                            "broadcast",
                            "balance-tlb",
                            "balance-alb",
                        ),
                    },
                },

                // Computed attributes
                "uuid": schema.StringAttribute{
                    Computed:            true,
                    MarkdownDescription: "BCM-assigned interface UUID.",
                },
                "base_type": schema.StringAttribute{
                    Computed:            true,
                    MarkdownDescription: "Entity base type (always 'NetworkInterface').",
                },
                "child_type": schema.StringAttribute{
                    Computed:            true,
                    MarkdownDescription: "Interface type (NetworkPhysicalInterface, NetworkBondInterface, NetworkBMCInterface).",
                },
                "cardtype": schema.StringAttribute{
                    Computed:            true,
                    MarkdownDescription: "Hardware card type (Ethernet, InfiniBand, BMC).",
                },
            },
        },
    }
}
```

---

## Type Mapping Functions

### Terraform Type to BCM childType

```go
// interfaceTypeToBCMChildType maps Terraform interface type to BCM childType.
func interfaceTypeToBCMChildType(tfType string) string {
    switch tfType {
    case "physical":
        return "NetworkPhysicalInterface"
    case "bond":
        return "NetworkBondInterface"
    case "bmc":
        return "NetworkBMCInterface"
    default:
        return "NetworkPhysicalInterface" // Default fallback
    }
}

// bcmChildTypeToInterfaceType maps BCM childType to Terraform interface type.
func bcmChildTypeToInterfaceType(childType string) string {
    switch childType {
    case "NetworkPhysicalInterface":
        return "physical"
    case "NetworkBondInterface":
        return "bond"
    case "NetworkBMCInterface":
        return "bmc"
    default:
        return "physical" // Default fallback
    }
}
```

---

## API Entity Mapping

### buildInterfaceAPIEntity

```go
// buildInterfaceAPIEntity constructs a BCM API interface entity from Terraform model.
func buildInterfaceAPIEntity(iface DeviceInterfaceModel, existingUUID string) map[string]interface{} {
    // Generate or use existing UUID
    uuid := existingUUID
    if uuid == "" {
        uuid = uuid.New().String()
    }

    entity := map[string]interface{}{
        "baseType":             "NetworkInterface",
        "childType":            interfaceTypeToBCMChildType(iface.Type.ValueString()),
        "uuid":                 uuid,
        "name":                 iface.Name.ValueString(),
        "modified":             true,
        "to_be_removed":        false,
        "revision":             "",
    }

    // Network assignment
    if !iface.Network.IsNull() && !iface.Network.IsUnknown() {
        entity["network"] = iface.Network.ValueString()
    }

    if !iface.MAC.IsNull() && !iface.MAC.IsUnknown() {
        entity["mac"] = iface.MAC.ValueString()
    }

    if !iface.IP.IsNull() && !iface.IP.IsUnknown() {
        entity["ip"] = iface.IP.ValueString()
    }

    if !iface.IPv6IP.IsNull() && !iface.IPv6IP.IsUnknown() {
        entity["ipv6Ip"] = iface.IPv6IP.ValueString()
    } else {
        entity["ipv6Ip"] = "::0"
    }

    // Configuration flags with defaults
    if !iface.DHCP.IsNull() && !iface.DHCP.IsUnknown() {
        entity["dhcp"] = iface.DHCP.ValueBool()
    } else {
        entity["dhcp"] = true // Default
    }

    if !iface.Bootable.IsNull() && !iface.Bootable.IsUnknown() {
        entity["bootable"] = iface.Bootable.ValueBool()
    } else {
        entity["bootable"] = false // Default
    }

    if !iface.StartIf.IsNull() && !iface.StartIf.IsUnknown() {
        entity["startIf"] = iface.StartIf.ValueString()
    } else {
        entity["startIf"] = "ALWAYS" // Default
    }

    // Standard BCM fields
    entity["ipv6Dhcp"] = false
    entity["bringupduringinstall"] = "NO"

    // Card type based on interface type
    switch iface.Type.ValueString() {
    case "bmc":
        entity["cardtype"] = "BMC"
    default:
        entity["cardtype"] = "Ethernet"
    }

    // Bond-specific fields
    if iface.Type.ValueString() == "bond" {
        if !iface.Members.IsNull() && !iface.Members.IsUnknown() {
            var members []string
            iface.Members.ElementsAs(context.Background(), &members, false)
            entity["members"] = members
        }

        if !iface.BondMode.IsNull() && !iface.BondMode.IsUnknown() {
            entity["bondMode"] = iface.BondMode.ValueString()
        }
    }

    return entity
}
```

### parseInterfaceFromAPI

```go
// parseInterfaceFromAPI parses a BCM API interface response into Terraform model.
func parseInterfaceFromAPI(data map[string]interface{}) DeviceInterfaceModel {
    model := DeviceInterfaceModel{}

    // Identity
    model.Name = getStringValue(data, "name")
    model.UUID = getStringValue(data, "uuid")
    model.BaseType = getStringValue(data, "baseType")
    model.ChildType = getStringValue(data, "childType")

    // Derive type from childType
    if childType := model.ChildType.ValueString(); childType != "" {
        model.Type = types.StringValue(bcmChildTypeToInterfaceType(childType))
    }

    // Network assignment
    model.Network = getStringValue(data, "network")
    model.MAC = getStringValue(data, "mac")
    model.IP = getStringValue(data, "ip")
    model.IPv6IP = getStringValue(data, "ipv6Ip")

    // Configuration flags
    model.DHCP = getBoolValue(data, "dhcp")
    model.Bootable = getBoolValue(data, "bootable")
    model.StartIf = getStringValue(data, "startIf")
    model.CardType = getStringValue(data, "cardtype")

    // Bond-specific
    if members, ok := data["members"].([]interface{}); ok && len(members) > 0 {
        memberStrings := make([]string, len(members))
        for i, m := range members {
            memberStrings[i] = m.(string)
        }
        memberList, _ := types.ListValueFrom(context.Background(), types.StringType, memberStrings)
        model.Members = memberList
    } else {
        model.Members = types.ListNull(types.StringType)
    }

    model.BondMode = getStringValue(data, "bondMode")

    return model
}
```

---

## Validation Rules

### Terraform-Side Validators

| Rule | Validator | Error Message |
|------|-----------|---------------|
| Interface name unique per device | Custom plan modifier | "interface name '{name}' is duplicated" |
| Bond requires members | Custom validator | "type 'bond' requires at least one member in 'members'" |
| Members only for bond | Custom validator | "'members' can only be specified when type is 'bond'" |
| Bond mode only for bond | Custom validator | "'bond_mode' can only be specified when type is 'bond'" |

### BCM-Side Validation (via validateDevice)

| Validation | Field | Example Error |
|------------|-------|---------------|
| Network exists | `interfaces[0].network` | "Network with UUID 'xxx' not found" |
| MAC format valid | `interfaces[0].mac` | "Invalid MAC address format" |
| Bond members exist | `interfaces[0].members` | "Member interface 'eth99' not found" |
| Duplicate names | `interfaces` | "Duplicate interface name 'eth0'" |

---

## State Management

### Interface UUID Tracking

1. **Create**: Generate UUID before API call, store in state
2. **Read**: Extract UUID from BCM response, update state
3. **Update**: Match interfaces by name, preserve UUIDs
4. **Import**: Populate UUIDs from BCM response

### Interface Ordering

The order of interfaces in the `interfaces` block is significant:
- Terraform maintains order for drift detection
- First bootable interface becomes `provisioningInterface`
- BCM may return interfaces in different order - normalize on read

### Null vs Empty Handling

| Field | Null Meaning | Empty Meaning |
|-------|--------------|---------------|
| `network` | Not assigned | N/A (use null) |
| `ip` | DHCP or no IP | N/A (use null) |
| `members` | Not a bond | Empty bond (invalid) |
| `bond_mode` | BCM default | N/A (use null) |

---

## Backward Compatibility

### Legacy Mode Detection

```go
// isLegacyMode returns true if device uses mac/management_network instead of interfaces block.
func isLegacyMode(plan CMDeviceDeviceResourceModel) bool {
    return len(plan.Interfaces) == 0 && !plan.MAC.IsNull()
}
```

### Legacy to Interfaces Migration

When importing a device created with legacy mode:
1. Create single interface entry from device-level `mac` and `management_network`
2. Name the interface "eth0" (BCM default)
3. Set `type = "physical"`, `bootable = true`, `dhcp = true`
4. Populate `interfaces` block in state

---

## Example Configurations

### Single Physical Interface

```hcl
resource "bcm_cmdevice_device" "compute" {
  hostname = "compute-01"
  category = data.bcm_cmdevice_categories.default.categories[0].uuid

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = "00:11:22:33:44:55"
    network  = data.bcm_cmnet_networks.mgmt.networks[0].uuid
    bootable = true
    dhcp     = true
  }
}
```

### Bond with Two Physical Members

```hcl
resource "bcm_cmdevice_device" "dgx" {
  hostname = "dgx-01"
  category = data.bcm_cmdevice_categories.dgx.categories[0].uuid

  interfaces {
    name      = "bond0"
    type      = "bond"
    members   = ["eth0", "eth1"]
    bond_mode = "802.3ad"
    network   = data.bcm_cmnet_networks.data.networks[0].uuid
    bootable  = true
    dhcp      = true
  }

  interfaces {
    name    = "eth2"
    type    = "physical"
    mac     = "00:11:22:33:44:57"
    network = data.bcm_cmnet_networks.storage.networks[0].uuid
  }
}
```

### BMC Interface

```hcl
resource "bcm_cmdevice_device" "server" {
  hostname = "server-01"
  category = data.bcm_cmdevice_categories.default.categories[0].uuid

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = "00:11:22:33:44:55"
    network  = data.bcm_cmnet_networks.mgmt.networks[0].uuid
    bootable = true
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
