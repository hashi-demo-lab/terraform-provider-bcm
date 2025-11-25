# Query all available roles in BCM
data "bcm_cmdevice_roles" "all" {}

# Output all available roles
output "all_roles" {
  description = "All roles available in BCM"
  value       = data.bcm_cmdevice_roles.all.roles
}

# Output role names for easy reference
output "role_names" {
  description = "Names of all available roles"
  value       = [for role in data.bcm_cmdevice_roles.all.roles : role.name]
}

# Output role types for categorization
output "role_types" {
  description = "Types of all available roles"
  value       = distinct([for role in data.bcm_cmdevice_roles.all.roles : role.child_type])
}
