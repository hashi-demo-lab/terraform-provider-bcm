# Query all cluster nodes
data "bcm_cmdevice_nodes" "all" {}

# Output all node information
output "all_nodes" {
  value = data.bcm_cmdevice_nodes.all.nodes
}

# Output node count
output "node_count" {
  value = length(data.bcm_cmdevice_nodes.all.nodes)
}

# Output just hostnames
output "hostnames" {
  value = [for node in data.bcm_cmdevice_nodes.all.nodes : node.hostname]
}

# Create inventory map
output "node_inventory" {
  value = {
    for node in data.bcm_cmdevice_nodes.all.nodes :
    node.hostname => {
      uuid       = node.uuid
      type       = node.child_type
      mac        = node.mac
      interfaces = length(node.interfaces)
      roles = [
        for role in node.roles :
        role.name if role.name != null
      ]
    }
  }
}
