# Filter roles by name pattern using glob wildcards
# Supports: * (any chars), ? (single char), [abc] (char class)

# Get all roles starting with "kube-"
data "bcm_cmdevice_roles" "kube_roles" {
  name_pattern = "kube-*"
}

# Get all roles ending with "node"
data "bcm_cmdevice_roles" "node_roles" {
  name_pattern = "*node"
}

# Get all roles containing "storage"
data "bcm_cmdevice_roles" "storage_pattern" {
  name_pattern = "*storage*"
}

# Combine pattern with type filter (AND logic)
data "bcm_cmdevice_roles" "kube_compute" {
  name_pattern = "kube-*"
  child_type   = "ComputeRole"
}

# Output filtered roles
output "kube_roles" {
  description = "Roles matching kube-* pattern"
  value       = data.bcm_cmdevice_roles.kube_roles.roles
}

output "node_roles" {
  description = "Roles matching *node pattern"
  value       = data.bcm_cmdevice_roles.node_roles.roles
}

# Use role data in other resources
# Example: Lookup a specific role to assign to a device
data "bcm_cmdevice_roles" "headnode" {
  name_pattern = "headnode"
}

output "headnode_role_uuid" {
  description = "UUID of the headnode role for device assignment"
  value       = length(data.bcm_cmdevice_roles.headnode.roles) > 0 ? data.bcm_cmdevice_roles.headnode.roles[0].uuid : null
}
