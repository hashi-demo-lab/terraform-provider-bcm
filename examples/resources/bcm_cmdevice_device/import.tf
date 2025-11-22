


# Lookup existing category and network
data "bcm_cmdevice_categories" "default" {
  name = "default"
}

data "bcm_cmnet_networks" "management" {
  filter {
    name_pattern = "DefaultEthernet"
  }
}

# Import an existing device using its UUID
# terraform import bcm_cmdevice_device.example <device-uuid>

resource "bcm_cmdevice_device" "example" {
  hostname           = "citest-import-example"
  mac                = "00:11:22:33:44:CC"
  category           = one(data.bcm_cmdevice_categories.default.categories[*].id)
  management_network = one(data.bcm_cmnet_networks.management.networks[*].id)
}
