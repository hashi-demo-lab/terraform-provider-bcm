# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# =============================================================================
# Example: Import an existing EtcdCluster into Terraform
# =============================================================================
# Write a resource configuration that matches the existing cluster,
# then run terraform import with the cluster UUID.
#
# Step 1: terraform import bcm_cmetcd_cluster.existing <cluster-uuid>
# Step 2: terraform plan (verify no unexpected changes)

# Import by UUID
# terraform import bcm_cmetcd_cluster.existing "e1e2e3e4-f5f6-7890-abcd-ef1234567890"

# Write a matching resource configuration
resource "bcm_cmetcd_cluster" "existing" {
  name               = "production-etcd"
  heartbeat_interval = 100
  election_timeout   = 1000
}
