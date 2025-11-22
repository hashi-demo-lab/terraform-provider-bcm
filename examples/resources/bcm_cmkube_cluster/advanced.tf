# Query available nodes for the cluster
data "bcm_cmdevice_nodes" "masters" {
  filter {
    hostname_pattern = "master"
  }
}

data "bcm_cmdevice_nodes" "workers" {
  filter {
    hostname_pattern = "worker"
  }
}

# Query available networks for cluster management
data "bcm_cmnet_networks" "all" {}

# Advanced cluster configuration with multiple workers
resource "bcm_cmkube_cluster" "production" {
  name = "prod-k8s-cluster"

  # High-availability master nodes (use first 3 masters)
  master_nodes = slice(data.bcm_cmdevice_nodes.masters.nodes[*].id, 0, 3)

  # Worker nodes for workloads (use first 4 workers)
  worker_nodes = slice(data.bcm_cmdevice_nodes.workers.nodes[*].id, 0, 4)

  # Kubernetes version
  version = "1.29.0"

  # Management network
  management_network = data.bcm_cmnet_networks.all.networks[0].id

  # Force operations (use with caution)
  force = false
}

# Example: Minimal cluster for development
resource "bcm_cmkube_cluster" "dev" {
  name         = "dev-k8s-cluster"
  master_nodes = [data.bcm_cmdevice_nodes.masters.nodes[0].id]
  version      = "1.28.0"
}

# Output both cluster UUIDs
output "production_cluster_uuid" {
  value       = bcm_cmkube_cluster.production.uuid
  description = "Production cluster UUID"
}

output "dev_cluster_uuid" {
  value       = bcm_cmkube_cluster.dev.uuid
  description = "Development cluster UUID"
}
