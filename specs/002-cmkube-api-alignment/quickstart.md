# Quickstart Guide: CMKube API Alignment

**Feature**: CMKube API Alignment (102-cmkube-api-alignment)
**Breaking Change**: Yes - schema changes are NOT backward compatible

## Migration Overview

This update aligns the BCM Kubernetes resources with the actual BCM API structure. The previous implementation had fields that were never persisted by BCM.

### What Changed

| Previous (Non-Functional) | New (Working) |
|---------------------------|---------------|
| `master_nodes` in cluster | `kubelet_role` block in device |
| `worker_nodes` in cluster | `kubelet_role` block in device |
| `etcd_nodes` in cluster | `etcd_host_role` block in device |
| `management_network` in cluster | `internal_network` in cluster |
| `overlay_network` in cluster | `pod_network` in cluster |
| `dns_servers` in cluster | NOT a BCM field (removed) |
| `cni_plugin` in cluster | Configured via `app_groups` |
| `storage_classes` in cluster | Configured via `app_groups` |
| `addons` in cluster | `app_groups` block |
| N/A | NEW: `bcm_cmetcd_cluster` resource |

## Complete Example: Kubernetes Cluster Setup

### Before (Old Schema - Never Worked Properly)

```hcl
# OLD SCHEMA - DO NOT USE
resource "bcm_cmkube_cluster" "example" {
  name               = "production"
  master_nodes       = [bcm_cmdevice_device.master.uuid]  # IGNORED BY BCM
  worker_nodes       = [bcm_cmdevice_device.worker.uuid]  # IGNORED BY BCM
  etcd_nodes         = [bcm_cmdevice_device.etcd.uuid]    # IGNORED BY BCM
  management_network = bcm_cmnet_network.internal.uuid    # WRONG FIELD NAME
  version            = "1.28.0"
  dns_servers        = ["8.8.8.8"]                        # NOT A BCM FIELD
  cni_plugin         = "calico"                           # NOT A BCM FIELD
}
```

### After (New Schema - Works With BCM API)

```hcl
# Step 1: Create Networks (prerequisite)
data "bcm_cmnet_networks" "all" {}

locals {
  internal_network = [for n in data.bcm_cmnet_networks.all.networks : n.uuid if n.name == "internal"][0]
  service_network  = [for n in data.bcm_cmnet_networks.all.networks : n.uuid if n.name == "service"][0]
  pod_network      = [for n in data.bcm_cmnet_networks.all.networks : n.uuid if n.name == "pod"][0]
}

# Step 2: Create EtcdCluster
resource "bcm_cmetcd_cluster" "production" {
  name               = "production-etcd"
  heartbeat_interval = 100
  election_timeout   = 1000
}

# Step 3: Create KubeCluster (references networks and etcd)
resource "bcm_cmkube_cluster" "production" {
  name             = "production"
  internal_network = local.internal_network
  service_network  = local.service_network
  pod_network      = local.pod_network
  etcd_cluster     = bcm_cmetcd_cluster.production.uuid
  version          = "1.28.0"

  pod_network_node_mask            = "/24"
  kube_dns_ip                      = "10.96.0.10"
  kubernetes_api_server_proxy_port = 6444

  trusted_domains = [
    "kubernetes.local",
    "api.production.cluster.local"
  ]

  # Optional: Application groups for addons
  app_groups {
    name    = "monitoring"
    enabled = true

    applications {
      name    = "prometheus"
      enabled = true
    }
  }
}

# Step 4: Assign nodes to cluster using role blocks in devices

# Etcd host node
resource "bcm_cmdevice_device" "etcd01" {
  hostname           = "etcd01"
  mac                = "00:11:22:33:44:01"
  category           = data.bcm_cmdevice_categories.all.categories[0].uuid
  management_network = local.internal_network

  etcd_host_role {
    etcd_cluster = bcm_cmetcd_cluster.production.uuid
    member_name  = "etcd01"
    spool        = "/var/lib/etcd"
  }
}

# Control plane node (master)
resource "bcm_cmdevice_device" "master01" {
  hostname           = "master01"
  mac                = "00:11:22:33:44:02"
  category           = data.bcm_cmdevice_categories.all.categories[0].uuid
  management_network = local.internal_network

  kubelet_role {
    kube_cluster  = bcm_cmkube_cluster.production.uuid
    control_plane = true
    worker        = false
    max_pods      = 110
  }
}

# Worker node
resource "bcm_cmdevice_device" "worker01" {
  hostname           = "worker01"
  mac                = "00:11:22:33:44:03"
  category           = data.bcm_cmdevice_categories.all.categories[0].uuid
  management_network = local.internal_network

  kubelet_role {
    kube_cluster  = bcm_cmkube_cluster.production.uuid
    control_plane = false
    worker        = true
    max_pods      = 250
  }
}
```

## Resource Examples

### bcm_cmetcd_cluster

```hcl
# Minimal configuration
resource "bcm_cmetcd_cluster" "minimal" {
  name = "etcd-cluster"
}

# Full configuration
resource "bcm_cmetcd_cluster" "full" {
  name               = "production-etcd"
  heartbeat_interval = 100   # milliseconds
  election_timeout   = 1000  # milliseconds (>= 5x heartbeat)

  options = jsonencode({
    "snapshot-count" = "10000"
  })
}
```

### bcm_cmkube_cluster

```hcl
# Minimal configuration
resource "bcm_cmkube_cluster" "minimal" {
  name             = "cluster"
  internal_network = "network-uuid"
  service_network  = "network-uuid"
  pod_network      = "network-uuid"
  etcd_cluster     = bcm_cmetcd_cluster.example.uuid
}

# Full configuration
resource "bcm_cmkube_cluster" "full" {
  name             = "production"
  internal_network = bcm_cmnet_network.internal.uuid
  service_network  = bcm_cmnet_network.service.uuid
  pod_network      = bcm_cmnet_network.pod.uuid
  etcd_cluster     = bcm_cmetcd_cluster.production.uuid

  version                          = "1.28.0"
  pod_network_node_mask            = "/24"
  kube_dns_ip                      = "10.96.0.10"
  kubernetes_api_server            = "https://api.cluster.local:6443"
  kubernetes_api_server_proxy_port = 6444

  trusted_domains = [
    "kubernetes.local",
    "api.cluster.local",
    "*.cluster.local"
  ]

  ingress_proxy_enable       = true
  ingress_proxy_listen_port  = 80
  ingress_proxy_backend_port = 30080

  app_groups {
    name    = "networking"
    enabled = true

    applications {
      name    = "calico"
      enabled = true
    }
  }

  app_groups {
    name    = "monitoring"
    enabled = true

    applications {
      name     = "prometheus"
      enabled  = true
      manifest = file("${path.module}/manifests/prometheus.yaml")
    }

    applications {
      name    = "grafana"
      enabled = true
    }
  }

  options = jsonencode({
    "featureGates" = {
      "NodeLease" = true
    }
  })
}
```

### Device with Kubernetes Roles

```hcl
# Combined etcd + control plane node
resource "bcm_cmdevice_device" "combined" {
  hostname           = "k8s-node01"
  mac                = "00:11:22:33:44:55"
  category           = data.bcm_cmdevice_categories.all.categories[0].uuid
  management_network = local.internal_network

  # Existing roles still work
  roles = ["backup", "provisioning"]

  # Etcd membership
  etcd_host_role {
    etcd_cluster           = bcm_cmetcd_cluster.production.uuid
    member_name            = "k8s-node01"
    spool                  = "/var/lib/etcd"
    listen_client_urls     = ["https://0.0.0.0:2379"]
    listen_peer_urls       = ["https://0.0.0.0:2380"]
    advertise_client_urls  = ["https://$ip:2379"]
    advertise_peer_urls    = ["https://$ip:2380"]
    snapshot_count         = 100000
    max_snapshots          = 5
  }

  # Kubernetes membership
  kubelet_role {
    kube_cluster              = bcm_cmkube_cluster.production.uuid
    control_plane             = true
    worker                    = true  # Combined control-plane + worker
    container_runtime_service = "containerd.service"
    max_pods                  = 110

    options = jsonencode({
      "image-gc-high-threshold" = "85"
      "image-gc-low-threshold"  = "80"
    })

    custom_yaml = <<-EOF
      apiVersion: kubelet.config.k8s.io/v1beta1
      kind: KubeletConfiguration
      evictionHard:
        memory.available: "100Mi"
        nodefs.available: "10%"
    EOF
  }
}
```

## Migration Steps

### Step 1: Import Existing Resources

```bash
# Import existing etcd cluster (if managed outside Terraform)
terraform import bcm_cmetcd_cluster.production <etcd-cluster-uuid>

# Import existing kube cluster
terraform import bcm_cmkube_cluster.production <kube-cluster-uuid>
```

### Step 2: Update Terraform Configuration

1. Remove deprecated fields from `bcm_cmkube_cluster`:
   - `master_nodes`
   - `worker_nodes`
   - `etcd_nodes`
   - `dns_servers`
   - `cni_plugin`
   - `storage_classes`
   - `load_balancer_mode`

2. Add required new fields:
   - `internal_network` (was `management_network`)
   - `service_network` (NEW)
   - `pod_network` (was `overlay_network`)
   - `etcd_cluster` (NEW)

3. Create `bcm_cmetcd_cluster` resource if not exists

4. Add role blocks to device resources:
   - `kubelet_role` for Kubernetes nodes
   - `etcd_host_role` for etcd nodes

### Step 3: Apply Changes

```bash
# Review planned changes
terraform plan

# Apply changes (will update cluster configuration)
terraform apply
```

## Dependency Graph

```
Networks (prerequisite)
    │
    └──► EtcdCluster
            │
            └──► KubeCluster
                    │
                    └──► Devices with kubelet_role
                            │
                            └──► (etcd_host_role references EtcdCluster)
```

## Troubleshooting

### Error: etcd_cluster is required

```
Error: Missing required argument
  etcd_cluster is required for bcm_cmkube_cluster
```

**Solution**: Create a `bcm_cmetcd_cluster` resource first and reference it:

```hcl
resource "bcm_cmetcd_cluster" "main" {
  name = "main-etcd"
}

resource "bcm_cmkube_cluster" "main" {
  # ...
  etcd_cluster = bcm_cmetcd_cluster.main.uuid
}
```

### Error: Invalid network UUID

```
Error: Invalid attribute value
  internal_network must be a valid UUID
```

**Solution**: Use a network data source or resource to get the UUID:

```hcl
data "bcm_cmnet_networks" "all" {}

resource "bcm_cmkube_cluster" "main" {
  internal_network = data.bcm_cmnet_networks.all.networks[0].uuid
  # ...
}
```

### Error: Node not joining cluster

**Cause**: Device missing `kubelet_role` block

**Solution**: Add the role block to the device:

```hcl
resource "bcm_cmdevice_device" "worker" {
  # ...

  kubelet_role {
    kube_cluster = bcm_cmkube_cluster.main.uuid
  }
}
```

## Data Sources

### bcm_cmkube_clusters

```hcl
# List all clusters
data "bcm_cmkube_clusters" "all" {}

# Filter by name pattern
data "bcm_cmkube_clusters" "prod" {
  filter {
    name_pattern = "prod"
  }
}

# Filter by version
data "bcm_cmkube_clusters" "v128" {
  filter {
    version = "1.28.0"
  }
}
```

### bcm_cmetcd_clusters (NEW)

```hcl
# List all etcd clusters
data "bcm_cmetcd_clusters" "all" {}

# Filter by name
data "bcm_cmetcd_clusters" "production" {
  filter {
    name_pattern = "production"
  }
}
```
