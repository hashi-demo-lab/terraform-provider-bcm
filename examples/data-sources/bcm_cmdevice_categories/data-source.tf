# Query all categories
data "bcm_cmdevice_categories" "all" {}

# Output all category names and UUIDs
output "all_categories" {
  description = "Map of category names to UUIDs"
  value = {
    for cat in data.bcm_cmdevice_categories.all.categories :
    cat.name => cat.uuid
  }
}

# Output software image assignments
output "category_images" {
  description = "Map of categories to their software images"
  value = {
    for cat in data.bcm_cmdevice_categories.all.categories :
    cat.name => cat.software_image_id
  }
}

# Output boot configuration
output "boot_configuration" {
  description = "Boot configuration per category"
  value = {
    for cat in data.bcm_cmdevice_categories.all.categories :
    cat.name => {
      boot_loader  = cat.boot_loader
      protocol     = cat.boot_loader_protocol
      install_mode = cat.install_mode
    }
  }
}
