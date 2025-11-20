# Filter nodes by hostname pattern (case-insensitive substring match)
data "bcm_cmdevice_nodes" "matching" {
  filter {
    hostname_pattern = "node"
  }
}

# Output matching nodes
output "matching_nodes" {
  value = [for node in data.bcm_cmdevice_nodes.matching.nodes : node.hostname]
}

# Filter for specific node name prefix
data "bcm_cmdevice_nodes" "compute_nodes" {
  filter {
    hostname_pattern = "compute"
  }
}

output "compute_hostnames" {
  value = [for node in data.bcm_cmdevice_nodes.compute_nodes.nodes : node.hostname]
}
