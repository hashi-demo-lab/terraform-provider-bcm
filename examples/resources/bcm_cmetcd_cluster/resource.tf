# Example: Minimal EtcdCluster configuration
resource "bcm_cmetcd_cluster" "example" {
  name = "production-etcd"
}

# Example: EtcdCluster with custom timing parameters
resource "bcm_cmetcd_cluster" "custom_timing" {
  name               = "ha-etcd"
  heartbeat_interval = 150  # milliseconds (default: 100)
  election_timeout   = 1500 # milliseconds (default: 1000)
}

# Example: EtcdCluster with custom options
resource "bcm_cmetcd_cluster" "with_options" {
  name = "custom-etcd"
  options = jsonencode({
    "quota-backend-bytes"       = "8589934592"
    "auto-compaction-mode"      = "periodic"
    "auto-compaction-retention" = "1h"
  })
}

# Example: Production-grade EtcdCluster for high availability
# Recommendation: election_timeout >= 5x heartbeat_interval
resource "bcm_cmetcd_cluster" "production" {
  name               = "prod-ha-etcd"
  heartbeat_interval = 100  # 100ms heartbeat
  election_timeout   = 1000 # 1000ms = 10x heartbeat for stable elections

  options = jsonencode({
    "quota-backend-bytes"       = "8589934592" # 8GB backend quota
    "auto-compaction-mode"      = "periodic"
    "auto-compaction-retention" = "1h"
    "snapshot-count"            = "10000" # Snapshot after 10k transactions
  })
}
