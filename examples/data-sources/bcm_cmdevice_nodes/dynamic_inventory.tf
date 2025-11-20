# Complex usage: Build dynamic inventory with network details
data "bcm_cmdevice_nodes" "all" {}

# Extract primary IP addresses from first interface
locals {
  node_ips = {
    for node in data.bcm_cmdevice_nodes.all.nodes :
    node.hostname => length(node.interfaces) > 0 ? node.interfaces[0].ip : null
  }
}

# Group nodes by type
locals {
  nodes_by_type = {
    for node_type in distinct([for node in data.bcm_cmdevice_nodes.all.nodes : node.child_type]) :
    node_type => [
      for node in data.bcm_cmdevice_nodes.all.nodes :
      node if node.child_type == node_type
    ]
  }
}

# Build ansible-style inventory
locals {
  ansible_inventory = {
    for node in data.bcm_cmdevice_nodes.all.nodes :
    node.hostname => {
      ansible_host = length(node.interfaces) > 0 ? node.interfaces[0].ip : null
      node_type    = node.child_type
      node_uuid    = node.uuid
      node_mac     = node.mac
      roles        = [for role in node.roles : role.name]
      interfaces = [
        for iface in node.interfaces : {
          name = iface.name
          ip   = iface.ip
          mac  = iface.mac
        }
      ]
    }
  }
}

# Outputs
output "node_ips" {
  description = "Map of hostnames to primary IP addresses"
  value       = local.node_ips
}

output "nodes_by_type" {
  description = "Nodes grouped by type"
  value       = { for k, v in local.nodes_by_type : k => [for node in v : node.hostname] }
}

output "ansible_inventory" {
  description = "Ansible-compatible inventory"
  value       = local.ansible_inventory
}

# Filter head nodes and extract management IPs
data "bcm_cmdevice_nodes" "head_nodes" {
  filter {
    node_type = "HeadNode"
  }
}

output "head_node_management_ips" {
  value = [
    for node in data.bcm_cmdevice_nodes.head_nodes.nodes :
    length(node.interfaces) > 0 ? node.interfaces[0].ip : "no-ip"
  ]
}
