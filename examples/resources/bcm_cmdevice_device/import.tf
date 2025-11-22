

provider "bcm" {
  # Configuration provided via environment variables:
  # export BCM_ENDPOINT="https://172.21.15.254:8081"
  # export BCM_USERNAME="root"
  # export BCM_PASSWORD="your-password"
  insecure_skip_verify = true
}

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
  category           = data.bcm_cmdevice_categories.default.categories[0].id
  management_network = data.bcm_cmnet_networks.management.networks[0].id
}
