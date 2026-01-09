# BCM Kubernetes Cluster Resource Examples
#
# The bcm_cmkube_cluster resource aligns with BCM's KubeCluster API entity.
# Cluster node membership is managed via roles on bcm_cmdevice_device resources,
# not directly on the cluster resource.

# Query available networks for cluster networking
data "bcm_cmnet_networks" "all" {}

# Create an EtcdCluster first (required for KubeCluster)
resource "bcm_cmetcd_cluster" "example" {
  name               = "example-etcd"
  heartbeat_interval = 100
  election_timeout   = 1000
}

# Basic Kubernetes cluster with minimal configuration
resource "bcm_cmkube_cluster" "basic" {
  name = "basic-cluster"

  # Required network references (must be valid Network UUIDs)
  etcd_cluster     = bcm_cmetcd_cluster.example.uuid
  internal_network = data.bcm_cmnet_networks.all.networks[0].uuid
  service_network  = data.bcm_cmnet_networks.all.networks[1].uuid
  pod_network      = data.bcm_cmnet_networks.all.networks[2].uuid
}

# Production Kubernetes cluster with full configuration
resource "bcm_cmkube_cluster" "production" {
  name = "production-cluster"

  # Required network references
  etcd_cluster     = bcm_cmetcd_cluster.example.uuid
  internal_network = data.bcm_cmnet_networks.all.networks[0].uuid
  service_network  = data.bcm_cmnet_networks.all.networks[1].uuid
  pod_network      = data.bcm_cmnet_networks.all.networks[2].uuid

  # Kubernetes version
  version = "1.29.0"

  # Pod networking configuration
  pod_network_node_mask = "/24"

  # DNS configuration
  kube_dns_ip = "10.96.0.10"

  # API server configuration
  kubernetes_api_server            = "https://api.cluster.local:6443"
  kubernetes_api_server_proxy_port = 6444

  # Certificate SANs
  trusted_domains = [
    "kubernetes.local",
    "api.cluster.local",
    "192.168.1.100"
  ]

  # Ingress proxy settings
  ingress_proxy_enable       = true
  ingress_proxy_listen_port  = 443
  ingress_proxy_backend_port = 8443

  # Application groups for cluster addons
  app_groups {
    name    = "monitoring"
    enabled = true
  }

  app_groups {
    name    = "networking"
    enabled = true
  }

  # Node label sets
  label_sets {
    name = "gpu-nodes"
    labels = {
      "nvidia.com/gpu" = "true"
      "gpu-type"       = "a100"
    }
  }

  # Kubernetes users for kubeconfig
  users {
    name   = "admin"
    groups = ["system:masters"]
  }

  users {
    name   = "developer"
    groups = ["developers", "viewers"]
  }

  # Extensible options (JSON)
  options = jsonencode({
    "custom-setting" = "value"
  })
}

# Output cluster information
output "basic_cluster_id" {
  value = bcm_cmkube_cluster.basic.id
}

output "basic_cluster_uuid" {
  value = bcm_cmkube_cluster.basic.uuid
}

output "production_cluster_creation_time" {
  value = bcm_cmkube_cluster.production.creation_time
}

output "production_cluster_revision" {
  value = bcm_cmkube_cluster.production.revision_id
}
