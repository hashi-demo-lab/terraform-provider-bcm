# Advanced usage: Accessing List attributes and network configuration
# Demonstrates time_servers, name_servers, search_domains, admin_email

data "bcm_cmpart_partitions" "all" {}

# Extract all unique DNS servers across partitions
locals {
  all_dns_servers = distinct(flatten([
    for p in data.bcm_cmpart_partitions.all.partitions :
    p.name_servers if p.name_servers != null
  ]))
}

# Extract all unique NTP servers across partitions
locals {
  all_ntp_servers = distinct(flatten([
    for p in data.bcm_cmpart_partitions.all.partitions :
    p.time_servers if p.time_servers != null
  ]))
}

# Build network configuration summary per partition
locals {
  network_summary = {
    for p in data.bcm_cmpart_partitions.all.partitions :
    p.name => {
      dns_servers    = p.name_servers
      search_domains = p.search_domains
      ntp_servers    = p.time_servers
      admin_emails   = p.admin_email
    }
  }
}

# Outputs
output "unique_dns_servers" {
  description = "All unique DNS servers configured across partitions"
  value       = local.all_dns_servers
}

output "unique_ntp_servers" {
  description = "All unique NTP time servers configured across partitions"
  value       = local.all_ntp_servers
}

output "network_summary" {
  description = "Network configuration summary for all partitions"
  value       = local.network_summary
}

# Find partitions with specific NTP server
output "partitions_using_google_ntp" {
  description = "Partitions configured with Google's NTP server"
  value = [
    for p in data.bcm_cmpart_partitions.all.partitions :
    p.name if p.time_servers != null && contains(p.time_servers, "time.google.com")
  ]
}
