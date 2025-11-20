# Filter categories by name
data "bcm_cmdevice_categories" "default" {
  name = "default"
}

# Output the default category details
output "default_category_uuid" {
  description = "UUID of the default category"
  value       = data.bcm_cmdevice_categories.default.categories[0].uuid
}

output "default_category_software_image" {
  description = "Software image assigned to default category"
  value       = data.bcm_cmdevice_categories.default.categories[0].software_image_id
}

output "default_category_network" {
  description = "Management network for default category"
  value       = data.bcm_cmdevice_categories.default.categories[0].management_network_id
}

# Use in local variables for cleaner references
locals {
  default_category = data.bcm_cmdevice_categories.default.categories[0]
}

output "category_summary" {
  description = "Summary of default category configuration"
  value = {
    name           = local.default_category.name
    uuid           = local.default_category.uuid
    software_image = local.default_category.software_image_id
    boot_loader    = local.default_category.boot_loader
    install_mode   = local.default_category.install_mode
    kernel_params  = local.default_category.kernel_parameters
    notes          = local.default_category.notes
  }
}
