# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# =============================================================================
# Example: Import an existing Kubernetes cluster into Terraform
# =============================================================================
# Use data sources to look up the required network and etcd references,
# then write a matching resource configuration before importing.
#
# Step 1: terraform apply -target=data.bcm_cmnet_networks.all -target=data.bcm_cmkube_clusters.existing
# Step 2: terraform import bcm_cmkube_cluster.existing <cluster-uuid>
# Step 3: terraform plan (verify no unexpected changes)

# Lookup existing clusters to discover UUIDs
data "bcm_cmkube_clusters" "existing" {}

# Lookup networks referenced by the cluster
data "bcm_cmnet_networks" "all" {}

# Lookup the etcd cluster backing Kubernetes
resource "bcm_cmetcd_cluster" "backing" {
  name = "production-etcd"
}

# Write a matching resource configuration
resource "bcm_cmkube_cluster" "existing" {
  name = "production-cluster"

  etcd_cluster     = bcm_cmetcd_cluster.backing.uuid
  internal_network = data.bcm_cmnet_networks.all.networks[0].uuid
  service_network  = data.bcm_cmnet_networks.all.networks[1].uuid
  pod_network      = data.bcm_cmnet_networks.all.networks[2].uuid

  version = "1.29.0"
}
