resource "bcm_cmnet_network" "example" {
  name             = "compute-network"
  subnet           = "10.0.1.0/24"
  gateway          = "10.0.1.1"
  mtu              = 9000
  domain_name      = "cluster.local"
  dhcp_range_start = "10.0.1.100"
  dhcp_range_end   = "10.0.1.200"
  notes            = "High-performance compute network managed by Terraform"
}
