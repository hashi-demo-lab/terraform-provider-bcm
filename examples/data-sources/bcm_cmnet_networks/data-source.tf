# Retrieve all networks from BCM cluster
data "bcm_cmnet_networks" "all" {}

# Output network information
output "all_networks" {
  description = "All networks configured in the BCM cluster"
  value       = data.bcm_cmnet_networks.all.networks
}

output "network_count" {
  description = "Total number of networks in the BCM cluster"
  value       = length(data.bcm_cmnet_networks.all.networks)
}

# Output specific network details (first network as example)
output "first_network_details" {
  description = "Details of the first network"
  value = length(data.bcm_cmnet_networks.all.networks) > 0 ? {
    name         = data.bcm_cmnet_networks.all.networks[0].name
    uuid         = data.bcm_cmnet_networks.all.networks[0].uuid
    base_address = data.bcm_cmnet_networks.all.networks[0].base_address
    netmask_bits = data.bcm_cmnet_networks.all.networks[0].netmask_bits
    dhcp_enabled = data.bcm_cmnet_networks.all.networks[0].dhcp_enabled
    management   = data.bcm_cmnet_networks.all.networks[0].management
  } : null
}
