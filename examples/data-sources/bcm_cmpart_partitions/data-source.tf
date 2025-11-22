# Retrieve all partitions from BCM cluster
data "bcm_cmpart_partitions" "all" {}

# Filter partitions by name pattern (case-insensitive)
data "bcm_cmpart_partitions" "base_partitions" {
  filter {
    name_pattern = "base"
  }
}

# Example: Use partition UUID in software image resource
# resource "bcm_cmpart_softwareimage" "example" {
#   name       = "my-custom-image"
#   path       = "/cm/images/my-custom-image"
#   bootfspart = data.bcm_cmpart_partitions.base_partitions.partitions[0].uuid
#   # ... other configuration
# }

# Output partition information
output "all_partition_names" {
  description = "Names of all partitions in the cluster"
  value       = [for p in data.bcm_cmpart_partitions.all.partitions : p.name]
}

output "base_partition_uuid" {
  description = "UUID of the first partition matching 'base'"
  value = length(data.bcm_cmpart_partitions.base_partitions.partitions) > 0 ? data.bcm_cmpart_partitions.base_partitions.partitions[0].uuid : null
}

output "partition_details" {
  description = "Detailed information about all partitions"
  value = {
    for p in data.bcm_cmpart_partitions.all.partitions : p.name => {
      uuid         = p.uuid
      cluster_name = p.cluster_name
      slave_name   = p.slave_name
      time_servers = p.time_servers
      name_servers = p.name_servers
    }
  }
}
