---
page_title: "bcm_cmkube_cluster Resource - bcm"
subcategory: ""
description: |-
  Manages a BCM Kubernetes cluster definition.
  KubeCluster entities define the cluster configuration including network references, API server settings, and application groups. Node membership is managed via KubeletRole on device resources, not on the cluster itself.
---

# bcm_cmkube_cluster (Resource)

Manages a BCM Kubernetes cluster definition.

KubeCluster entities define the cluster configuration including network references, API server settings, and application groups. Node membership is managed via KubeletRole on device resources, not on the cluster itself.

## Example Usage

```terraform
# BCM Kubernetes Cluster Resource Examples
#
# The bcm_cmkube_cluster resource aligns with BCM's KubeCluster API entity.
# Cluster node membership is managed via roles on bcm_cmdevice_device resources,
# not directly on the cluster resource.

# Query available networks for cluster networking
data "bcm_cmnet_networks" "all" {}

# Create an EtcdCluster first (required for KubeCluster)
resource "bcm_cmetcd_cluster" "example" {
  name               = "example-etcd"
  heartbeat_interval = 100
  election_timeout   = 1000
}

# Basic Kubernetes cluster with minimal configuration
resource "bcm_cmkube_cluster" "basic" {
  name = "basic-cluster"

  # Required network references (must be valid Network UUIDs)
  etcd_cluster     = bcm_cmetcd_cluster.example.uuid
  internal_network = data.bcm_cmnet_networks.all.networks[0].uuid
  service_network  = data.bcm_cmnet_networks.all.networks[1].uuid
  pod_network      = data.bcm_cmnet_networks.all.networks[2].uuid
}

# Production Kubernetes cluster with full configuration
resource "bcm_cmkube_cluster" "production" {
  name = "production-cluster"

  # Required network references
  etcd_cluster     = bcm_cmetcd_cluster.example.uuid
  internal_network = data.bcm_cmnet_networks.all.networks[0].uuid
  service_network  = data.bcm_cmnet_networks.all.networks[1].uuid
  pod_network      = data.bcm_cmnet_networks.all.networks[2].uuid

  # Kubernetes version
  version = "1.29.0"

  # Pod networking configuration
  pod_network_node_mask = "/24"

  # DNS configuration
  kube_dns_ip = "10.96.0.10"

  # API server configuration
  kubernetes_api_server            = "https://api.cluster.local:6443"
  kubernetes_api_server_proxy_port = 6444

  # Certificate SANs
  trusted_domains = [
    "kubernetes.local",
    "api.cluster.local",
    "192.168.1.100"
  ]

  # Ingress proxy settings
  ingress_proxy_enable       = true
  ingress_proxy_listen_port  = 443
  ingress_proxy_backend_port = 8443

  # Application groups for cluster addons
  app_groups {
    name    = "monitoring"
    enabled = true
  }

  app_groups {
    name    = "networking"
    enabled = true
  }

  # Node label sets
  label_sets {
    name = "gpu-nodes"
    labels = {
      "nvidia.com/gpu" = "true"
      "gpu-type"       = "a100"
    }
  }

  # Kubernetes users for kubeconfig
  users {
    name   = "admin"
    groups = ["system:masters"]
  }

  users {
    name   = "developer"
    groups = ["developers", "viewers"]
  }

  # Extensible options (JSON)
  options = jsonencode({
    "custom-setting" = "value"
  })
}

# Output cluster information
output "basic_cluster_id" {
  value = bcm_cmkube_cluster.basic.id
}

output "basic_cluster_uuid" {
  value = bcm_cmkube_cluster.basic.uuid
}

output "production_cluster_creation_time" {
  value = bcm_cmkube_cluster.production.creation_time
}

output "production_cluster_revision" {
  value = bcm_cmkube_cluster.production.revision_id
}
```

### Advanced Configuration

```terraform
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

# Advanced cluster configuration with multiple workers
# Note: Requires at least 1 master node, 1+ worker nodes, and 1+ networks
resource "bcm_cmkube_cluster" "production" {
  name = "prod-k8s-cluster"

  # High-availability master nodes (use up to first 3 masters)
  master_nodes = slice(data.bcm_cmdevice_nodes.masters.nodes[*].id, 0, min(3, length(data.bcm_cmdevice_nodes.masters.nodes)))

  # Worker nodes for workloads (use up to first 4 workers)
  worker_nodes = slice(data.bcm_cmdevice_nodes.workers.nodes[*].id, 0, min(4, length(data.bcm_cmdevice_nodes.workers.nodes)))

  # Kubernetes version
  version = "1.29.0"

  # Management network
  management_network = length(data.bcm_cmnet_networks.all.networks) > 0 ? data.bcm_cmnet_networks.all.networks[0].id : null

  # Force operations (use with caution)
  force = false
}

# Example: Minimal cluster for development
# Note: Requires at least 1 master node
resource "bcm_cmkube_cluster" "dev" {
  name         = "dev-k8s-cluster"
  master_nodes = length(data.bcm_cmdevice_nodes.masters.nodes) > 0 ? [data.bcm_cmdevice_nodes.masters.nodes[0].id] : []
  version      = "1.28.0"
}

# Output both cluster UUIDs
output "production_cluster_uuid" {
  value       = bcm_cmkube_cluster.production.uuid
  description = "Production cluster UUID"
}

output "dev_cluster_uuid" {
  value       = bcm_cmkube_cluster.dev.uuid
  description = "Development cluster UUID"
}
```

### Advanced Networking, Storage, and Addons

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `etcd_cluster` (String) UUID of the EtcdCluster entity that backs this Kubernetes cluster.
- `internal_network` (String) UUID of the internal network for cluster node communication.
- `name` (String) Cluster name (RFC 1123 DNS label: lowercase alphanumeric and hyphens, 1-63 characters).
- `pod_network` (String) UUID of the pod network for container IPs.
- `service_network` (String) UUID of the service network for Kubernetes service IPs.

### Optional

- `app_groups` (Block List) Application groups for cluster addons. Each group contains applications (Kubernetes manifests) that can be enabled/disabled together. (see [below for nested schema](#nestedblock--app_groups))
- `ingress_proxy_backend_port` (Number) Ingress proxy backend port.
- `ingress_proxy_enable` (Boolean) Enable ingress proxy for external traffic routing.
- `ingress_proxy_listen_port` (Number) Ingress proxy listen port. BCM may set a server-side default.
- `kube_dns_ip` (String) Cluster DNS IP address. BCM sets a default value if not specified.
- `kubernetes_api_server` (String) Kubernetes API server URL.
- `kubernetes_api_server_proxy_port` (Number) Kubernetes API server proxy port. BCM defaults to 6444.
- `label_sets` (Block List) Label sets that can be applied to nodes, categories, or overlays. (see [below for nested schema](#nestedblock--label_sets))
- `options` (String) Extensible configuration options as JSON string.
- `pod_network_node_mask` (String) Pod network node mask for CIDR allocation (e.g., '/24').
- `trusted_domains` (List of String) List of trusted domains for certificate SANs.
- `users` (Block List) Kubernetes users for kubeconfig management. (see [below for nested schema](#nestedblock--users))
- `version` (String) Kubernetes version (semver format, e.g., '1.28.0').

### Read-Only

- `creation_time` (Number) Unix timestamp of when the cluster was created.
- `id` (String) Cluster identifier (same as uuid)
- `revision_id` (Number) Revision number for change tracking.
- `uuid` (String) BCM-assigned cluster UUID

<a id="nestedblock--app_groups"></a>
### Nested Schema for `app_groups`

Required:

- `name` (String) Application group name.

Optional:

- `applications` (Block List) Applications within this group. (see [below for nested schema](#nestedblock--app_groups--applications))
- `enabled` (Boolean) Whether the application group is enabled.

<a id="nestedblock--app_groups--applications"></a>
### Nested Schema for `app_groups.applications`

Required:

- `name` (String) Application name.

Optional:

- `enabled` (Boolean) Whether the application is enabled.
- `manifest` (String) Kubernetes manifest YAML/JSON content.



<a id="nestedblock--label_sets"></a>
### Nested Schema for `label_sets`

Required:

- `name` (String) Label set name.

Optional:

- `labels` (Map of String) Map of label key-value pairs.


<a id="nestedblock--users"></a>
### Nested Schema for `users`

Required:

- `name` (String) User name.

Optional:

- `groups` (List of String) List of groups the user belongs to.

## Import

Import is supported using the following syntax:

```shell
#!/bin/bash
# Import an existing Kubernetes cluster into Terraform state

# Get the cluster UUID from BCM
# You can find this via the BCM API or web interface

CLUSTER_UUID="your-cluster-uuid-here"

# Import the cluster
terraform import bcm_cmkube_cluster.example "${CLUSTER_UUID}"

# After import, create a matching configuration:
# resource "bcm_cmkube_cluster" "example" {
#   name         = "existing-cluster"
#   master_nodes = ["master-uuid"]
# }
```

### Import with Data Source Lookup

```terraform
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# =============================================================================
# Example: Import an existing Kubernetes cluster into Terraform
# =============================================================================
# Use data sources to look up the required network and etcd references,
# then write a matching resource configuration before importing.
#
# Step 1: terraform apply -target=data.bcm_cmnet_networks.all -target=data.bcm_cmkube_clusters.existing
# Step 2: terraform import bcm_cmkube_cluster.existing <cluster-uuid>
# Step 3: terraform plan (verify no unexpected changes)

# Lookup existing clusters to discover UUIDs
data "bcm_cmkube_clusters" "existing" {}

# Lookup networks referenced by the cluster
data "bcm_cmnet_networks" "all" {}

# Lookup the etcd cluster backing Kubernetes
resource "bcm_cmetcd_cluster" "backing" {
  name = "production-etcd"
}

# Write a matching resource configuration
resource "bcm_cmkube_cluster" "existing" {
  name = "production-cluster"

  etcd_cluster     = bcm_cmetcd_cluster.backing.uuid
  internal_network = data.bcm_cmnet_networks.all.networks[0].uuid
  service_network  = data.bcm_cmnet_networks.all.networks[1].uuid
  pod_network      = data.bcm_cmnet_networks.all.networks[2].uuid

  version = "1.29.0"
}
```
