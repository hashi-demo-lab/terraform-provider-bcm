// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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

// buildInterfaceAPIEntity constructs a BCM API interface entity from Terraform model.
func buildInterfaceAPIEntity(iface DeviceInterfaceModel, existingUUID string) map[string]interface{} {
	// Generate or use existing UUID
	ifaceUUID := existingUUID
	if ifaceUUID == "" {
		ifaceUUID = uuid.New().String()
	}

	entity := map[string]interface{}{
		"baseType":      "NetworkInterface",
		"uuid":          ifaceUUID,
		"modified":      true,
		"to_be_removed": false,
		"revision":      "",
	}

	// Required fields - set via helpers for defensive null/unknown guards
	if !iface.Type.IsNull() && !iface.Type.IsUnknown() {
		entity["childType"] = interfaceTypeToBCMChildType(iface.Type.ValueString())
	}
	SetStringField(entity, "name", iface.Name)

	// Network assignment
	SetStringField(entity, "network", iface.Network)
	SetStringField(entity, "mac", iface.MAC)
	SetStringField(entity, "ip", iface.IP)

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

	// MAC address - BCM returns "00:00:00:00:00:00" as default for bonds, treat as null
	if mac := getStringValue(data, "mac"); !mac.IsNull() && mac.ValueString() != "" && mac.ValueString() != "00:00:00:00:00:00" {
		model.MAC = mac
	} else {
		model.MAC = types.StringNull()
	}

	// IP addresses - BCM returns "0.0.0.0" and "::0" as defaults, treat as null
	if ip := getStringValue(data, "ip"); !ip.IsNull() && ip.ValueString() != "" && ip.ValueString() != "0.0.0.0" {
		model.IP = ip
	} else {
		model.IP = types.StringNull()
	}

	if ipv6 := getStringValue(data, "ipv6Ip"); !ipv6.IsNull() && ipv6.ValueString() != "" && ipv6.ValueString() != "::0" {
		model.IPv6IP = ipv6
	} else {
		model.IPv6IP = types.StringNull()
	}

	// Configuration flags
	model.DHCP = getBoolValue(data, "dhcp")
	model.Bootable = getBoolValue(data, "bootable")
	model.StartIf = getStringValue(data, "startIf")
	model.CardType = getStringValue(data, "cardtype")

	// Bond-specific
	if members, ok := data["members"].([]interface{}); ok && len(members) > 0 {
		memberStrings := make([]string, len(members))
		for i, m := range members {
			if str, ok := m.(string); ok {
				memberStrings[i] = str
			}
		}
		memberList, _ := types.ListValueFrom(context.Background(), types.StringType, memberStrings)
		model.Members = memberList
	} else {
		model.Members = types.ListNull(types.StringType)
	}

	model.BondMode = getStringValue(data, "bondMode")

	return model
}

// findInterfaceByName searches for an interface by name in a list of interfaces.
// Returns the interface and its index, or nil and -1 if not found.
func findInterfaceByName(interfaces []DeviceInterfaceModel, name string) (*DeviceInterfaceModel, int) {
	for i, iface := range interfaces {
		if iface.Name.ValueString() == name {
			return &interfaces[i], i
		}
	}
	return nil, -1
}

// buildInterfacesAPIArray constructs BCM API interfaces array from Terraform model.
// existingInterfaces is used to preserve UUIDs during updates.
func buildInterfacesAPIArray(interfaces []DeviceInterfaceModel, existingInterfaces []DeviceInterfaceModel) []interface{} {
	result := make([]interface{}, 0, len(interfaces))

	for _, iface := range interfaces {
		// Look for existing interface with same name to preserve UUID
		existingUUID := ""
		if existing, _ := findInterfaceByName(existingInterfaces, iface.Name.ValueString()); existing != nil {
			existingUUID = existing.UUID.ValueString()
		}

		entity := buildInterfaceAPIEntity(iface, existingUUID)
		result = append(result, entity)
	}

	return result
}

// parseInterfacesFromAPI parses BCM API interfaces array into Terraform model.
func parseInterfacesFromAPI(data interface{}) []DeviceInterfaceModel {
	interfaces, ok := data.([]interface{})
	if !ok || len(interfaces) == 0 {
		return nil
	}

	result := make([]DeviceInterfaceModel, 0, len(interfaces))
	for _, ifaceData := range interfaces {
		if ifaceMap, ok := ifaceData.(map[string]interface{}); ok {
			iface := parseInterfaceFromAPI(ifaceMap)
			result = append(result, iface)
		}
	}

	return result
}

// normalizeInterfaceOrder reorders parsed interfaces to match plan order and
// preserves plan values for fields that BCM doesn't return (like members, bond_mode).
// This prevents spurious diffs when BCM returns interfaces in a different order.
func normalizeInterfaceOrder(parsedInterfaces []DeviceInterfaceModel, planInterfaces []DeviceInterfaceModel) []DeviceInterfaceModel {
	if len(planInterfaces) == 0 {
		return parsedInterfaces
	}

	// Create lookup map for parsed interfaces by name
	parsedByName := make(map[string]DeviceInterfaceModel)
	for _, iface := range parsedInterfaces {
		parsedByName[iface.Name.ValueString()] = iface
	}

	// Reorder to match plan order and merge plan values
	result := make([]DeviceInterfaceModel, 0, len(planInterfaces))
	for _, planIface := range planInterfaces {
		if parsed, ok := parsedByName[planIface.Name.ValueString()]; ok {
			// Merge plan values for fields BCM doesn't return
			merged := mergeInterfaceWithPlan(parsed, planIface)
			result = append(result, merged)
			delete(parsedByName, planIface.Name.ValueString())
		}
	}

	// Append any interfaces not in plan (e.g., added externally)
	for _, iface := range parsedByName {
		result = append(result, iface)
	}

	return result
}

// mergeInterfaceWithPlan merges parsed interface data with plan values for fields
// that BCM API doesn't return. This is essential for bond-specific fields.
func mergeInterfaceWithPlan(parsed DeviceInterfaceModel, plan DeviceInterfaceModel) DeviceInterfaceModel {
	result := parsed

	// Preserve bond-specific fields from plan if not returned by BCM
	if parsed.Type.ValueString() == "bond" {
		// Members - BCM may not return this field
		if parsed.Members.IsNull() && !plan.Members.IsNull() && !plan.Members.IsUnknown() {
			result.Members = plan.Members
		}

		// Bond mode - BCM may not return this field
		if parsed.BondMode.IsNull() && !plan.BondMode.IsNull() && !plan.BondMode.IsUnknown() {
			result.BondMode = plan.BondMode
		}
	}

	return result
}
