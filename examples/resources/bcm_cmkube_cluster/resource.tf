# Basic Kubernetes cluster resource example

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

# Create a Kubernetes cluster with minimal configuration
# Note: Requires at least 1 master node in the environment
resource "bcm_cmkube_cluster" "example" {
  name         = "my-k8s-cluster"
  master_nodes = length(data.bcm_cmdevice_nodes.masters.nodes) > 0 ? [data.bcm_cmdevice_nodes.masters.nodes[0].id] : []
}

# Create a Kubernetes cluster with worker nodes
# Note: Requires at least 1 master node and 1+ worker nodes
resource "bcm_cmkube_cluster" "with_workers" {
  name         = "prod-cluster"
  master_nodes = length(data.bcm_cmdevice_nodes.masters.nodes) > 0 ? [data.bcm_cmdevice_nodes.masters.nodes[0].id] : []
  worker_nodes = slice(data.bcm_cmdevice_nodes.workers.nodes[*].id, 0, min(3, length(data.bcm_cmdevice_nodes.workers.nodes)))

  version = "1.28.0"
}

# Create a Kubernetes cluster with full configuration
# Note: Requires at least 1 master node, 1+ worker nodes, and 1+ networks
resource "bcm_cmkube_cluster" "advanced" {
  name         = "advanced-cluster"
  master_nodes = slice(data.bcm_cmdevice_nodes.masters.nodes[*].id, 0, min(3, length(data.bcm_cmdevice_nodes.masters.nodes)))
  worker_nodes = slice(data.bcm_cmdevice_nodes.workers.nodes[*].id, 0, min(5, length(data.bcm_cmdevice_nodes.workers.nodes)))

  version            = "1.29.0"
  management_network = length(data.bcm_cmnet_networks.all.networks) > 0 ? data.bcm_cmnet_networks.all.networks[0].id : null

  force = false # Set to true to bypass validation
}

# Query etcd nodes for HA configuration
data "bcm_cmdevice_nodes" "etcd" {
  filter {
    hostname_pattern = "etcd"
  }
}

# Create a production HA Kubernetes cluster with dedicated etcd nodes
# NVIDIA DGX BasePOD deployments require 3 dedicated etcd nodes for quorum
# Note: Requires at least 1 master, 3 etcd nodes, and 1+ worker nodes
resource "bcm_cmkube_cluster" "production_ha" {
  name         = "production-ha-cluster"
  master_nodes = slice(data.bcm_cmdevice_nodes.masters.nodes[*].id, 0, min(3, length(data.bcm_cmdevice_nodes.masters.nodes)))
  worker_nodes = slice(data.bcm_cmdevice_nodes.workers.nodes[*].id, 0, min(5, length(data.bcm_cmdevice_nodes.workers.nodes)))
  etcd_nodes   = slice(data.bcm_cmdevice_nodes.etcd.nodes[*].id, 0, min(3, length(data.bcm_cmdevice_nodes.etcd.nodes)))

  version            = "1.29.0"
  management_network = length(data.bcm_cmnet_networks.all.networks) > 0 ? data.bcm_cmnet_networks.all.networks[0].id : null

  force = false
}

# Output cluster information
output "cluster_id" {
  value = bcm_cmkube_cluster.example.id
}

output "cluster_uuid" {
  value = bcm_cmkube_cluster.example.uuid
}

output "cluster_creation_time" {
  value = bcm_cmkube_cluster.example.creation_time
}
