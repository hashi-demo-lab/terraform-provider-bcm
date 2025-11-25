# Filter roles by type (exact match on child_type)
# Common role types: HeadNodeRole, ComputeRole, StorageRole, etc.

# Get all compute roles
data "bcm_cmdevice_roles" "compute_roles" {
  child_type = "ComputeRole"
}

# Get all head node roles
data "bcm_cmdevice_roles" "head_roles" {
  child_type = "HeadNodeRole"
}

# Get all storage roles
data "bcm_cmdevice_roles" "storage_roles" {
  child_type = "StorageRole"
}

# Output filtered compute roles
output "compute_roles" {
  description = "Compute roles available in BCM"
  value       = data.bcm_cmdevice_roles.compute_roles.roles
}

# Output filtered head node roles
output "head_roles" {
  description = "Head node roles available in BCM"
  value       = data.bcm_cmdevice_roles.head_roles.roles
}
