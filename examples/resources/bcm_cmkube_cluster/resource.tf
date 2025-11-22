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
resource "bcm_cmkube_cluster" "example" {
  name         = "my-k8s-cluster"
  master_nodes = [data.bcm_cmdevice_nodes.masters.nodes[0].id]
}

# Create a Kubernetes cluster with worker nodes
resource "bcm_cmkube_cluster" "with_workers" {
  name         = "prod-cluster"
  master_nodes = [data.bcm_cmdevice_nodes.masters.nodes[0].id]
  worker_nodes = slice(data.bcm_cmdevice_nodes.workers.nodes[*].id, 0, 3)

  version = "1.28.0"
}

# Create a Kubernetes cluster with full configuration
resource "bcm_cmkube_cluster" "advanced" {
  name         = "advanced-cluster"
  master_nodes = slice(data.bcm_cmdevice_nodes.masters.nodes[*].id, 0, 3)
  worker_nodes = slice(data.bcm_cmdevice_nodes.workers.nodes[*].id, 0, 5)

  version            = "1.29.0"
  management_network = data.bcm_cmnet_networks.all.networks[0].id

  force = false # Set to true to bypass validation
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
