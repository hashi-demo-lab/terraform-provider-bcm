# Advanced cluster configuration with multiple workers
resource "bcm_cmkube_cluster" "production" {
  name = "prod-k8s-cluster"

  # High-availability master nodes
  master_nodes = [
    "<master-1-uuid>",
    "<master-2-uuid>",
    "<master-3-uuid>",
  ]

  # Worker nodes for workloads
  worker_nodes = [
    "<worker-1-uuid>",
    "<worker-2-uuid>",
    "<worker-3-uuid>",
    "<worker-4-uuid>",
  ]

  # Kubernetes version
  version = "1.29.0"

  # Management network
  management_network = "<prod-network-uuid>"

  # Force operations (use with caution)
  force = false
}

# Example: Minimal cluster for development
resource "bcm_cmkube_cluster" "dev" {
  name         = "dev-k8s-cluster"
  master_nodes = ["<dev-master-uuid>"]
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
