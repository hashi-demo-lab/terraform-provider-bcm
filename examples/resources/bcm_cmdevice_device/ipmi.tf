


# Lookup existing category by name
data "bcm_cmdevice_categories" "default" {
  name = "default"
}

# Lookup management network
data "bcm_cmnet_networks" "management" {
  filter {
    name_pattern = "DefaultEthernet"
  }
}

# Example: IPMI-enabled device with power control and network configuration
resource "bcm_cmdevice_device" "ipmi" {
  hostname           = "citest-compute-ipmi"
  mac                = "00:11:22:33:44:BB"
  category           = try(data.bcm_cmdevice_categories.default.categories[0].id, null)
  management_network = try(data.bcm_cmnet_networks.management.networks[0].id, null)

  # Power control via IPMI
  power_control = "ipmi"

  # Network gateway configuration
  default_gateway        = "192.168.1.1"
  default_gateway_metric = 100

  # Boot configuration
  boot_loader       = "pxelinux"
  kernel_parameters = "console=ttyS0,115200 ipmi_si.type=kcs"

  # Hardware identifiers (typically auto-discovered from BMC)
  serial_number = "SN123456789"
  part_number   = "PN-COMPUTE-001"

  notes = "IPMI-enabled compute node with power management"
}
