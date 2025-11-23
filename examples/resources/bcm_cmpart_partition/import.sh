#!/bin/bash
# Example: Import an existing BCM partition into Terraform state

# Step 1: Get the partition UUID from BCM
# You can find this via the BCM UI or by using the bcm_cmpart_partitions data source

# Step 2: Import the partition by UUID
terraform import bcm_cmpart_partition.engineering "550e8400-e29b-41d4-a716-446655440000"

# Step 3: Verify the import succeeded
terraform show

# Step 4: Generate Terraform configuration from imported state (Terraform 1.5+)
# terraform show -no-color > partition.tf
