# Example 1: List all Kubernetes clusters
data "bcm_cmkube_clusters" "all" {}

# Output all cluster names
output "all_cluster_names" {
  value = [for cluster in data.bcm_cmkube_clusters.all.clusters : cluster.name]
}

# Example 2: Filter clusters by name pattern
data "bcm_cmkube_clusters" "prod_clusters" {
  filter {
    name_pattern = "prod"
  }
}

# Output production cluster UUIDs
output "prod_cluster_uuids" {
  value = [for cluster in data.bcm_cmkube_clusters.prod_clusters.clusters : cluster.uuid]
}

# Example 3: Filter clusters by Kubernetes version
data "bcm_cmkube_clusters" "k8s_1_28" {
  filter {
    version = "1.28.0"
  }
}

# Output clusters running Kubernetes 1.28.0
output "k8s_1_28_clusters" {
  value = [for cluster in data.bcm_cmkube_clusters.k8s_1_28.clusters : {
    name    = cluster.name
    version = cluster.version
  }]
}

# Example 4: Filter clusters by master node UUID
data "bcm_cmkube_clusters" "clusters_with_node" {
  filter {
    master_node_id = "node-uuid-123"
  }
}

# Output clusters containing specific master node
output "clusters_with_specific_node" {
  value = [for cluster in data.bcm_cmkube_clusters.clusters_with_node.clusters : cluster.name]
}

# Example 5: Combine multiple filters (AND logic)
data "bcm_cmkube_clusters" "filtered" {
  filter {
    name_pattern = "prod"
    version      = "1.28.0"
  }
}

# Output clusters matching all filters
output "filtered_clusters" {
  value = [for cluster in data.bcm_cmkube_clusters.filtered.clusters : {
    name    = cluster.name
    version = cluster.version
    masters = cluster.master_nodes
  }]
}

# Example 6: Use cluster UUID for terraform import
# First, discover the cluster UUID:
data "bcm_cmkube_clusters" "import_lookup" {
  filter {
    name_pattern = "my-cluster"
  }
}

# Then use it in a resource import:
# terraform import bcm_cmkube_cluster.imported <uuid-from-data-source>
# UUID from: data.bcm_cmkube_clusters.import_lookup.clusters[0].uuid

# Example 7: Access cluster network details
data "bcm_cmkube_clusters" "network_info" {}

# Output cluster network configurations
output "cluster_networks" {
  value = [for cluster in data.bcm_cmkube_clusters.network_info.clusters : {
    name               = cluster.name
    management_network = cluster.management_network
    overlay_network    = cluster.overlay_network
    dns_servers        = cluster.dns_servers
  }]
}
