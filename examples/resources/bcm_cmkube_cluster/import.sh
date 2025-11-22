#!/bin/bash
# Import an existing Kubernetes cluster into Terraform state

# Get the cluster UUID from BCM
# You can find this via the BCM API or web interface

CLUSTER_UUID="your-cluster-uuid-here"

# Import the cluster
terraform import bcm_cmkube_cluster.example "${CLUSTER_UUID}"

# After import, create a matching configuration:
# resource "bcm_cmkube_cluster" "example" {
#   name         = "existing-cluster"
#   master_nodes = ["master-uuid"]
# }
