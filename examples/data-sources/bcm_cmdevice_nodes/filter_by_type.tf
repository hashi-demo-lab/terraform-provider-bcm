# Filter nodes by type (PhysicalNode)
data "bcm_cmdevice_nodes" "physical" {
  filter {
    node_type = "PhysicalNode"
  }
}

# Output physical nodes
output "physical_nodes" {
  value = [for node in data.bcm_cmdevice_nodes.physical.nodes : {
    hostname = node.hostname
    uuid     = node.uuid
    mac      = node.mac
  }]
}

# Filter for compute nodes
data "bcm_cmdevice_nodes" "compute" {
  filter {
    node_type = "ComputeNode"
  }
}

output "compute_count" {
  value = length(data.bcm_cmdevice_nodes.compute.nodes)
}
