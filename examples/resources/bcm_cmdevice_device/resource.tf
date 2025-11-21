terraform {
  required_providers {
    bcm = {
      source = "hashicorp/bcm"
    }
  }
}

provider "bcm" {
  endpoint             = "https://172.21.15.254:8081"
  username             = "root"
  password             = "Hashicorp123!"
  insecure_skip_verify = true
}

# Example: Create a basic compute device
resource "bcm_cmdevice_device" "compute_node" {
  hostname           = "compute-node-01"
  mac                = "00:11:22:33:44:55"
  category           = "your-category-uuid-here"
  management_network = "your-network-uuid-here"

  notes             = "Compute node managed by Terraform"
  kernel_parameters = "console=ttyS0,115200"
  boot_loader       = "pxelinux"
}

# Example: Create a device with power control and network configuration
resource "bcm_cmdevice_device" "ipmi_node" {
  hostname           = "ipmi-node-01"
  mac                = "00:11:22:33:44:66"
  category           = "your-category-uuid-here"
  management_network = "your-network-uuid-here"

  # Power control configuration
  power_control = "ipmi"

  # Network gateway configuration
  default_gateway        = "192.168.1.1"
  default_gateway_metric = 100

  # Hardware identifiers (optional, may be auto-discovered)
  serial_number = "SN123456789"
  part_number   = "PN-ABC-123"

  notes = "IPMI-enabled node with custom network configuration"
}
