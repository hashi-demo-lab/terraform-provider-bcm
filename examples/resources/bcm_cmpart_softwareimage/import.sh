#!/bin/bash
# Import an existing software image using its UUID

# Example: Import a software image with UUID
terraform import bcm_cmpart_softwareimage.example "eaad50d3-432a-4703-a9f8-66551c255a69"

# You can find the UUID of existing images using the data source:
# data "bcm_cmpart_softwareimages" "all" {}
# output "image_uuids" {
#   value = { for img in data.bcm_cmpart_softwareimages.all.images : img.name => img.uuid }
# }
