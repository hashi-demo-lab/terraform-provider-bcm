# Example 1: List all Kubernetes clusters
data "bcm_cmkube_clusters" "all" {}

# Output all cluster names
output "all_cluster_names" {
  value = data.bcm_cmkube_clusters.all.clusters != null ? [for cluster in data.bcm_cmkube_clusters.all.clusters : cluster.name] : []
}

# Example 2: Filter clusters by name pattern
data "bcm_cmkube_clusters" "prod_clusters" {
  filter {
    name_pattern = "prod"
  }
}

# Output production cluster UUIDs
output "prod_cluster_uuids" {
  value = data.bcm_cmkube_clusters.prod_clusters.clusters != null ? [for cluster in data.bcm_cmkube_clusters.prod_clusters.clusters : cluster.uuid] : []
}

# Example 3: Filter clusters by Kubernetes version
data "bcm_cmkube_clusters" "k8s_1_28" {
  filter {
    version = "1.28.0"
  }
}

# Output clusters running Kubernetes 1.28.0
output "k8s_1_28_clusters" {
  value = data.bcm_cmkube_clusters.k8s_1_28.clusters != null ? [for cluster in data.bcm_cmkube_clusters.k8s_1_28.clusters : {
    name    = cluster.name
    version = cluster.version
  }] : []
}

# Example 4: Filter clusters by etcd cluster UUID
data "bcm_cmkube_clusters" "clusters_with_etcd" {
  filter {
    etcd_cluster = "12345678-1234-1234-1234-123456789abc"
  }
}

# Output clusters referencing specific etcd cluster
output "clusters_with_specific_etcd" {
  value = data.bcm_cmkube_clusters.clusters_with_etcd.clusters != null ? [for cluster in data.bcm_cmkube_clusters.clusters_with_etcd.clusters : cluster.name] : []
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
  value = data.bcm_cmkube_clusters.filtered.clusters != null ? [for cluster in data.bcm_cmkube_clusters.filtered.clusters : {
    name         = cluster.name
    version      = cluster.version
    etcd_cluster = cluster.etcd_cluster
  }] : []
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

# Example 7: Access cluster network and API server details
data "bcm_cmkube_clusters" "network_info" {}

# Output cluster network configurations - aligned with BCM CMKube API entity
output "cluster_networks" {
  value = data.bcm_cmkube_clusters.network_info.clusters != null ? [for cluster in data.bcm_cmkube_clusters.network_info.clusters : {
    name             = cluster.name
    internal_network = cluster.internal_network
    service_network  = cluster.service_network
    pod_network      = cluster.pod_network
    etcd_cluster     = cluster.etcd_cluster
  }] : []
}

# Example 8: Access API server and ingress proxy configuration
output "cluster_api_config" {
  value = data.bcm_cmkube_clusters.network_info.clusters != null ? [for cluster in data.bcm_cmkube_clusters.network_info.clusters : {
    name                             = cluster.name
    kubernetes_api_server            = cluster.kubernetes_api_server
    kubernetes_api_server_proxy_port = cluster.kubernetes_api_server_proxy_port
    ingress_proxy_enable             = cluster.ingress_proxy_enable
    ingress_proxy_listen_port        = cluster.ingress_proxy_listen_port
  }] : []
}

# Example 9: Access nested application groups
output "cluster_app_groups" {
  value = data.bcm_cmkube_clusters.network_info.clusters != null ? [for cluster in data.bcm_cmkube_clusters.network_info.clusters : {
    name       = cluster.name
    app_groups = cluster.app_groups
  }] : []
}

# Example 10: Access nested label sets and users
output "cluster_label_sets_and_users" {
  value = data.bcm_cmkube_clusters.network_info.clusters != null ? [for cluster in data.bcm_cmkube_clusters.network_info.clusters : {
    name       = cluster.name
    label_sets = cluster.label_sets
    users      = cluster.users
  }] : []
}
