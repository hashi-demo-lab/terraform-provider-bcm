terraform {
  required_providers {
    bcm = {
      source = "hashicorp.com/nvidia/bcm"
    }
  }
}

provider "bcm" {
  endpoint             = "https://172.21.15.254:8081"
  username             = "root"
  password             = "your-password-here"
  insecure_skip_verify = true
}

# Basic partition with minimal configuration
resource "bcm_cmpart_partition" "basic" {
  name         = "basic-partition"
  cluster_name = "Test Cluster"
}

# Complete partition with all optional fields
resource "bcm_cmpart_partition" "engineering" {
  name         = "engineering"
  cluster_name = "HPC Production Cluster"
  slave_name   = "compute"
  slave_digits = 4

  admin_email    = ["admin@example.com", "ops@example.com"]
  time_servers   = ["ntp1.example.com", "ntp2.example.com"]
  name_servers   = ["8.8.8.8", "8.8.4.4"]
  search_domains = ["example.com", "corp.example.com"]

  relay_host   = "smtp.example.com"
  no_zero_conf = false

  notes = "Engineering team partition for GPU workloads"
}

output "engineering_partition_uuid" {
  description = "UUID of the engineering partition"
  value       = bcm_cmpart_partition.engineering.uuid
}

output "engineering_partition_id" {
  description = "ID of the engineering partition (same as UUID)"
  value       = bcm_cmpart_partition.engineering.id
}
