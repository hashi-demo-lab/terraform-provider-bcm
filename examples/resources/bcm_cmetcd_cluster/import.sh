# Import an existing EtcdCluster by UUID
terraform import bcm_cmetcd_cluster.example "e1e2e3e4-f5f6-7890-abcd-ef1234567890"

# Import by name (if BCM supports name-based lookup)
terraform import bcm_cmetcd_cluster.example "production-etcd"
