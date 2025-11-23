# Example: Kubernetes Cluster with Advanced P3 Features
# Demonstrates full API coverage including networking, storage, and addons

# Query available nodes for the cluster
data "bcm_cmdevice_nodes" "masters" {
  filter {
    hostname_pattern = "master"
  }
}

data "bcm_cmdevice_nodes" "workers" {
  filter {
    hostname_pattern = "worker"
  }
}

# Query available networks for cluster management
data "bcm_cmnet_networks" "all" {}

# Example 1: Production cluster with advanced networking
# Note: Requires at least 1 master node, 1+ worker nodes, and 1+ networks
resource "bcm_cmkube_cluster" "production_advanced" {
  name         = "prod-k8s-advanced"
  master_nodes = slice(data.bcm_cmdevice_nodes.masters.nodes[*].id, 0, min(3, length(data.bcm_cmdevice_nodes.masters.nodes)))
  worker_nodes = slice(data.bcm_cmdevice_nodes.workers.nodes[*].id, 0, min(5, length(data.bcm_cmdevice_nodes.workers.nodes)))

  # Kubernetes configuration
  version    = "1.29.0"
  cni_plugin = "calico"

  # Network configuration
  management_network = length(data.bcm_cmnet_networks.all.networks) > 0 ? data.bcm_cmnet_networks.all.networks[0].id : null
  overlay_network    = "10.244.0.0/16"        # Pod network CIDR
  dns_servers        = ["8.8.8.8", "8.8.4.4"] # Custom DNS servers

  # Load balancer configuration
  load_balancer_mode = "metallb"

  # Storage classes (JSON-encoded array)
  storage_classes = jsonencode([
    {
      name              = "fast-ssd"
      provisioner       = "kubernetes.io/csi-driver"
      volumeBindingMode = "Immediate"
      parameters = {
        type = "ssd"
        iops = "3000"
      }
    },
    {
      name              = "standard"
      provisioner       = "kubernetes.io/csi-driver"
      volumeBindingMode = "WaitForFirstConsumer"
      parameters = {
        type = "standard"
      }
    }
  ])

  # Cluster addons (JSON-encoded array)
  addons = jsonencode([
    {
      name    = "prometheus"
      enabled = true
      version = "2.45.0"
      config = {
        retention = "30d"
        storage   = "100Gi"
      }
    },
    {
      name    = "grafana"
      enabled = true
      version = "10.0.0"
      config = {
        adminPassword = "changeme"
      }
    },
    {
      name    = "elasticsearch"
      enabled = true
      version = "8.9.0"
      config = {
        replicas = 3
        storage  = "200Gi"
      }
    }
  ])

  # Ingress controller (JSON-encoded object)
  ingress_controller = jsonencode({
    type    = "nginx"
    enabled = true
    version = "1.8.0"
    config = {
      replicaCount = 3
      service = {
        type = "LoadBalancer"
      }
      resources = {
        requests = {
          cpu    = "100m"
          memory = "128Mi"
        }
        limits = {
          cpu    = "500m"
          memory = "512Mi"
        }
      }
    }
  })

  force = false
}

# Example 2: Development cluster with minimal P3 features
# Note: Requires at least 1 master node
resource "bcm_cmkube_cluster" "dev_with_addons" {
  name         = "dev-k8s-addons"
  master_nodes = length(data.bcm_cmdevice_nodes.masters.nodes) > 0 ? [data.bcm_cmdevice_nodes.masters.nodes[0].id] : []

  # Kubernetes configuration
  version    = "1.28.0"
  cni_plugin = "flannel" # Simpler CNI for dev

  # Custom DNS for development
  dns_servers = ["192.168.1.1"]

  # Basic monitoring addon only
  addons = jsonencode([
    {
      name    = "prometheus"
      enabled = true
      version = "2.45.0"
      config = {
        retention = "7d" # Shorter retention for dev
        storage   = "10Gi"
      }
    }
  ])

  force = false
}

# Example 3: High-availability cluster with Weave CNI and full stack
# Note: Requires at least 1 master node, 1+ worker nodes, and 1+ networks
resource "bcm_cmkube_cluster" "ha_full_stack" {
  name         = "ha-k8s-full"
  master_nodes = slice(data.bcm_cmdevice_nodes.masters.nodes[*].id, 0, min(3, length(data.bcm_cmdevice_nodes.masters.nodes)))
  worker_nodes = slice(data.bcm_cmdevice_nodes.workers.nodes[*].id, 0, min(10, length(data.bcm_cmdevice_nodes.workers.nodes)))

  # Kubernetes configuration
  version    = "1.29.0"
  cni_plugin = "weave"

  # Network configuration
  management_network = length(data.bcm_cmnet_networks.all.networks) > 0 ? data.bcm_cmnet_networks.all.networks[0].id : null
  overlay_network    = "10.32.0.0/12"             # Weave network CIDR
  dns_servers        = ["10.0.0.10", "10.0.0.11"] # Internal DNS

  # Load balancer
  load_balancer_mode = "haproxy"

  # Multiple storage classes for different workloads
  storage_classes = jsonencode([
    {
      name        = "fast"
      provisioner = "kubernetes.io/csi-driver"
      parameters = {
        type = "nvme"
      }
    },
    {
      name        = "slow"
      provisioner = "kubernetes.io/csi-driver"
      parameters = {
        type = "hdd"
      }
    }
  ])

  # Full monitoring and logging stack
  addons = jsonencode([
    {
      name    = "prometheus"
      enabled = true
    },
    {
      name    = "grafana"
      enabled = true
    },
    {
      name    = "elasticsearch"
      enabled = true
    },
    {
      name    = "kibana"
      enabled = true
    },
    {
      name    = "fluentd"
      enabled = true
    }
  ])

  # Traefik ingress controller
  ingress_controller = jsonencode({
    type    = "traefik"
    enabled = true
    version = "2.10.0"
    config = {
      replicaCount = 3
      ports = {
        web = {
          port = 80
        }
        websecure = {
          port = 443
        }
      }
    }
  })
}

# Outputs demonstrating P3 fields
output "production_cluster_details" {
  value = {
    uuid               = bcm_cmkube_cluster.production_advanced.uuid
    name               = bcm_cmkube_cluster.production_advanced.name
    version            = bcm_cmkube_cluster.production_advanced.version
    cni_plugin         = bcm_cmkube_cluster.production_advanced.cni_plugin
    load_balancer_mode = bcm_cmkube_cluster.production_advanced.load_balancer_mode
    dns_servers        = bcm_cmkube_cluster.production_advanced.dns_servers
  }
}

output "dev_cluster_uuid" {
  value       = bcm_cmkube_cluster.dev_with_addons.uuid
  description = "Development cluster UUID"
}

output "ha_cluster_uuid" {
  value       = bcm_cmkube_cluster.ha_full_stack.uuid
  description = "High-availability cluster UUID"
}
