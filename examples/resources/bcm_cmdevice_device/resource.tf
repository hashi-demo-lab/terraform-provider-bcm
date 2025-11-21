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
