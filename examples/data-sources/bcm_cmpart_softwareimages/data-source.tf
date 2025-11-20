data "bcm_cmpart_softwareimages" "all" {}

output "all_images" {
  description = "All software images from BCM"
  value       = data.bcm_cmpart_softwareimages.all.images
}

output "image_count" {
  description = "Number of software images"
  value       = length(data.bcm_cmpart_softwareimages.all.images)
}

output "image_names" {
  description = "List of software image names"
  value = [
    for img in data.bcm_cmpart_softwareimages.all.images :
    img.name
  ]
}

output "ubuntu_images" {
  description = "Filter images by name pattern (ubuntu)"
  value = [
    for img in data.bcm_cmpart_softwareimages.all.images :
    img if can(regex("ubuntu", lower(img.name)))
  ]
}

output "images_with_modules" {
  description = "Images that have kernel modules configured"
  value = [
    for img in data.bcm_cmpart_softwareimages.all.images :
    {
      name         = img.name
      module_count = length(img.modules)
      module_names = [for mod in img.modules : mod.name]
    }
    if length(img.modules) > 0
  ]
}

output "sol_enabled_images" {
  description = "Images with Serial Over LAN enabled"
  value = [
    for img in data.bcm_cmpart_softwareimages.all.images :
    {
      name      = img.name
      sol_port  = img.sol_port
      sol_speed = img.sol_speed
    }
    if img.enable_sol
  ]
}
