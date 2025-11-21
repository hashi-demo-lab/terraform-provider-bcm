# Find management network by name pattern
data "bcm_cmnet_networks" "management" {
  filter {
    name_pattern = "management"
  }
}

# Find all DHCP-enabled networks
data "bcm_cmnet_networks" "dhcp_networks" {
  filter {
    dhcp_enabled = true
  }
}

# Find internal network with combined filters (AND logic)
data "bcm_cmnet_networks" "internal_dhcp" {
  filter {
    name_pattern = "internal"
    dhcp_enabled = true
  }
}

# Create map of network names to UUIDs for DHCP-enabled networks
locals {
  dhcp_network_map = {
    for net in data.bcm_cmnet_networks.dhcp_networks.networks :
    net.name => net.uuid
  }

  # Extract management network UUID if found
  management_network_uuid = length(data.bcm_cmnet_networks.management.networks) > 0 ? data.bcm_cmnet_networks.management.networks[0].uuid : null
}

# Output filtered network information
output "management_network_uuid" {
  description = "UUID of the management network"
  value       = local.management_network_uuid
}

output "dhcp_network_names" {
  description = "List of all DHCP-enabled network names"
  value       = [for net in data.bcm_cmnet_networks.dhcp_networks.networks : net.name]
}

output "dhcp_network_map" {
  description = "Map of DHCP network names to their UUIDs"
  value       = local.dhcp_network_map
}

output "internal_dhcp_count" {
  description = "Number of internal networks with DHCP enabled"
  value       = length(data.bcm_cmnet_networks.internal_dhcp.networks)
}

# Example: Use network UUID in a resource configuration (commented out - for reference)
# resource "bcm_cmdevice_node" "example" {
#   name = "example-node"
#
#   network_interface {
#     network_uuid = local.management_network_uuid
#     # other interface configuration...
#   }
# }
