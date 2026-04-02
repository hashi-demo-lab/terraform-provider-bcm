

# Lookup all networks (use first available as management network)
data "bcm_cmnet_networks" "all" {}

# Create a category for compute nodes
resource "bcm_cmdevice_category" "compute" {
  name               = "citest-compute-nodes"
  management_network = data.bcm_cmnet_networks.all.networks[0].id
  notes              = "Category for compute cluster nodes"
}

# Create a software image for compute nodes
resource "bcm_cmpart_softwareimage" "ubuntu_compute" {
  name              = "citest-ubuntu-22.04-compute"
  path              = "/cm/images/ubuntu-22.04-server-amd64.iso"
  kernel_parameters = "console=ttyS0,115200 net.ifnames=0"
  enable_sol        = true
  sol_speed         = "115200"
}

# Example 1: Basic compute device with minimal configuration
resource "bcm_cmdevice_device" "compute_basic" {
  hostname = "citest-compute-node-01"
  category = bcm_cmdevice_category.compute.id

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = "00:11:22:33:44:55"
    network  = data.bcm_cmnet_networks.all.networks[0].id
    bootable = true
    dhcp     = true
  }

  notes = "Basic compute node managed by Terraform"
}

# Example 2: Compute device with custom kernel parameters and boot configuration
resource "bcm_cmdevice_device" "compute_custom" {
  hostname = "citest-compute-node-02"
  category = bcm_cmdevice_category.compute.id

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = "00:11:22:33:44:56"
    network  = data.bcm_cmnet_networks.all.networks[0].id
    bootable = true
    dhcp     = true
  }

  # Boot configuration
  boot_loader       = "pxelinux"
  kernel_parameters = "console=ttyS0,115200 net.ifnames=0 biosdevname=0"

  notes = "Compute node with custom kernel parameters"
}

# Example 3: IPMI-enabled device with power control and network configuration
resource "bcm_cmdevice_device" "compute_ipmi" {
  hostname = "citest-compute-node-03"
  category = bcm_cmdevice_category.compute.id

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = "00:11:22:33:44:57"
    network  = data.bcm_cmnet_networks.all.networks[0].id
    bootable = true
    dhcp     = true
  }

  # Power control via IPMI
  power_control = "ipmi"

  # Network gateway configuration
  default_gateway        = "192.168.1.1"
  default_gateway_metric = 100

  # Boot configuration
  boot_loader       = "pxelinux"
  kernel_parameters = "console=ttyS0,115200 ipmi_si.type=kcs"

  # Hardware identifiers (typically auto-discovered from BMC)
  serial_number = "SN123456789"
  part_number   = "PN-COMPUTE-001"

  notes = "IPMI-enabled compute node with power management"
}

# Example 4: GPU compute node with multiple interfaces
# Management NIC for provisioning, high-speed data NIC for GPU traffic,
# and a BMC interface for out-of-band management.
resource "bcm_cmdevice_device" "gpu_node" {
  hostname = "citest-gpu-node-01"
  category = bcm_cmdevice_category.compute.id

  # Management interface - used for PXE boot and provisioning
  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = "00:11:22:33:44:58"
    network  = data.bcm_cmnet_networks.all.networks[0].id
    bootable = true
    dhcp     = true
  }

  # High-speed data interface - for GPU-to-GPU and storage traffic
  interfaces {
    name = "eth1"
    type = "physical"
    mac  = "00:11:22:33:44:68"
    dhcp = true
  }

  # BMC interface - out-of-band IPMI management
  interfaces {
    name = "ipmi"
    type = "bmc"
    ip   = "192.168.100.11"
    dhcp = false
  }

  # Power control via IPMI (uses BMC interface)
  power_control = "ipmi"

  # Network configuration for high-performance networking
  default_gateway        = "192.168.1.1"
  default_gateway_metric = 50

  # Boot configuration with GPU-specific parameters
  boot_loader       = "pxelinux"
  kernel_parameters = "console=ttyS0,115200 nouveau.modeset=0 nvidia-drm.modeset=1"

  # Hardware identifiers
  serial_number = "SN-GPU-001"
  part_number   = "PN-GPU-NODE-001"

  notes = "GPU compute node with management, data, and BMC interfaces"
}

# Example 5: Storage node with bonded interfaces
# Uses a bond for redundancy on the storage network.
resource "bcm_cmdevice_device" "storage_node" {
  hostname = "citest-storage-node-01"
  category = bcm_cmdevice_category.compute.id

  # Management interface for provisioning
  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = "00:11:22:33:44:59"
    network  = data.bcm_cmnet_networks.all.networks[0].id
    bootable = true
    dhcp     = true
  }

  # Member interfaces for bond (no network assignment - managed by bond)
  interfaces {
    name = "eth1"
    type = "physical"
    mac  = "00:11:22:33:44:69"
  }

  interfaces {
    name = "eth2"
    type = "physical"
    mac  = "00:11:22:33:44:79"
  }

  # Bonded interface for storage traffic (uses eth1 + eth2)
  interfaces {
    name      = "bond0"
    type      = "bond"
    dhcp      = true
    members   = ["eth1", "eth2"]
    bond_mode = "802.3ad"
  }

  # Power control
  power_control = "ipmi"

  # Network configuration
  default_gateway        = "192.168.1.1"
  default_gateway_metric = 100

  # Boot and partition configuration
  boot_loader       = "pxelinux"
  kernel_parameters = "console=ttyS0,115200"

  # Hardware identifiers
  serial_number = "SN-STORAGE-001"
  part_number   = "PN-STORAGE-NODE-001"

  notes = "Storage node with bonded interfaces for Ceph cluster"
}

# =============================================================================
# Kubernetes Cluster Integration Examples
# =============================================================================
# The following examples demonstrate how to assign Kubernetes and etcd roles
# to devices. Node membership is managed via role blocks on device resources,
# not directly on cluster resources.

# Create an EtcdCluster for Kubernetes backing store
resource "bcm_cmetcd_cluster" "k8s_etcd" {
  name               = "citest-k8s-etcd"
  heartbeat_interval = 100
  election_timeout   = 1000
}

# Create a KubeCluster for Kubernetes workloads
resource "bcm_cmkube_cluster" "production" {
  name = "citest-production"

  # Required network references
  etcd_cluster     = bcm_cmetcd_cluster.k8s_etcd.uuid
  internal_network = data.bcm_cmnet_networks.all.networks[0].uuid
  service_network  = data.bcm_cmnet_networks.all.networks[0].uuid
  pod_network      = data.bcm_cmnet_networks.all.networks[0].uuid

  # Kubernetes version
  version = "1.29.0"
}

# Example 6: Kubernetes control plane node with kubelet_role
# This device runs the Kubernetes control plane components (API server, etc.)
resource "bcm_cmdevice_device" "k8s_control_plane" {
  hostname = "citest-k8s-cp-01"
  category = bcm_cmdevice_category.compute.id

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = "00:11:22:33:44:60"
    network  = data.bcm_cmnet_networks.all.networks[0].id
    bootable = true
    dhcp     = true
  }

  # Power control
  power_control = "ipmi"

  # Network configuration
  default_gateway        = "192.168.1.1"
  default_gateway_metric = 100

  # Kubernetes control plane role
  kubelet_role {
    kube_cluster  = bcm_cmkube_cluster.production.uuid
    control_plane = true
    worker        = false # Control plane only, no workloads
  }

  notes = "Kubernetes control plane node"
}

# Example 7: Kubernetes worker node with kubelet_role
# This device runs application workloads scheduled by Kubernetes
resource "bcm_cmdevice_device" "k8s_worker" {
  hostname = "citest-k8s-worker-01"
  category = bcm_cmdevice_category.compute.id

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = "00:11:22:33:44:61"
    network  = data.bcm_cmnet_networks.all.networks[0].id
    bootable = true
    dhcp     = true
  }

  # Power control
  power_control = "ipmi"

  # Network configuration
  default_gateway        = "192.168.1.1"
  default_gateway_metric = 100

  # Kubernetes worker role
  kubelet_role {
    kube_cluster  = bcm_cmkube_cluster.production.uuid
    control_plane = false
    worker        = true
  }

  notes = "Kubernetes worker node for application workloads"
}

# Example 8: Etcd host node with etcd_host_role
# This device hosts an etcd cluster member for distributed key-value storage
resource "bcm_cmdevice_device" "etcd_host" {
  hostname = "citest-etcd-host-01"
  category = bcm_cmdevice_category.compute.id

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = "00:11:22:33:44:62"
    network  = data.bcm_cmnet_networks.all.networks[0].id
    bootable = true
    dhcp     = true
  }

  # Power control
  power_control = "ipmi"

  # Network configuration
  default_gateway        = "192.168.1.1"
  default_gateway_metric = 100

  # Etcd host role
  etcd_host_role {
    etcd_cluster = bcm_cmetcd_cluster.k8s_etcd.uuid
  }

  notes = "Etcd cluster member node"
}

# Example 9: Combined control plane node with both roles
# This device runs both etcd and Kubernetes control plane (common in small clusters)
resource "bcm_cmdevice_device" "k8s_combined_cp" {
  hostname = "citest-k8s-combined-01"
  category = bcm_cmdevice_category.compute.id

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = "00:11:22:33:44:63"
    network  = data.bcm_cmnet_networks.all.networks[0].id
    bootable = true
    dhcp     = true
  }

  # Power control
  power_control = "ipmi"

  # Network configuration
  default_gateway        = "192.168.1.1"
  default_gateway_metric = 100

  # Etcd host role - runs etcd for cluster state
  etcd_host_role {
    etcd_cluster = bcm_cmetcd_cluster.k8s_etcd.uuid
  }

  # Kubernetes control plane role - runs API server, scheduler, etc.
  kubelet_role {
    kube_cluster  = bcm_cmkube_cluster.production.uuid
    control_plane = true
    worker        = true # Also accepts workloads (useful for small clusters)
  }

  notes = "Combined etcd + Kubernetes control plane node"
}

# =============================================================================
# Provisioning Kubernetes on Existing Devices
# =============================================================================
# When you have existing devices in BCM and want to provision Kubernetes:
#
# 1. Import existing devices into Terraform state:
#    terraform import bcm_cmdevice_device.existing_cp <device-uuid>
#    terraform import bcm_cmdevice_device.existing_worker <device-uuid>
#
# 2. Create EtcdCluster and KubeCluster resources (see above)
#
# 3. Add role blocks to existing device resources
#
# The following examples show configurations for imported devices:

# Example 10: Existing device converted to Kubernetes control plane + etcd
# Import first: terraform import bcm_cmdevice_device.existing_cp <uuid>
resource "bcm_cmdevice_device" "existing_cp" {
  # These values must match the existing device after import
  hostname = "existing-server-01"
  category = bcm_cmdevice_category.compute.id

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = "00:11:22:33:44:70"
    network  = data.bcm_cmnet_networks.all.networks[0].id
    bootable = true
    dhcp     = true
  }

  # Existing device settings are preserved
  power_control = "ipmi"

  # Add etcd role to run etcd cluster member
  etcd_host_role {
    etcd_cluster = bcm_cmetcd_cluster.k8s_etcd.uuid
  }

  # Add kubelet role for Kubernetes control plane
  kubelet_role {
    kube_cluster  = bcm_cmkube_cluster.production.uuid
    control_plane = true
    worker        = false # Dedicated control plane, no workloads
  }

  notes = "Existing device converted to K8s control plane"
}

# Example 11: Existing device converted to Kubernetes worker
# Import first: terraform import bcm_cmdevice_device.existing_worker <uuid>
resource "bcm_cmdevice_device" "existing_worker" {
  hostname = "existing-server-02"
  category = bcm_cmdevice_category.compute.id

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = "00:11:22:33:44:71"
    network  = data.bcm_cmnet_networks.all.networks[0].id
    bootable = true
    dhcp     = true
  }

  power_control = "ipmi"

  # Workers only need kubelet_role (no etcd)
  kubelet_role {
    kube_cluster  = bcm_cmkube_cluster.production.uuid
    control_plane = false
    worker        = true
  }

  notes = "Existing device converted to K8s worker"
}

# =============================================================================
# Typical Kubernetes Node Configurations Reference
# =============================================================================
#
# | Node Type              | etcd_host_role | kubelet_role                    |
# |------------------------|----------------|---------------------------------|
# | Control plane + etcd   | Yes            | control_plane=true, worker=false|
# | Control plane only     | No             | control_plane=true, worker=false|
# | Worker                 | No             | control_plane=false, worker=true|
# | Combined (small cluster)| Yes           | control_plane=true, worker=true |
#
# For production clusters:
# - Use 3+ control plane nodes with etcd for high availability
# - Separate etcd from control plane for large clusters (1000+ nodes)
# - Workers should not run etcd or control plane components

# Output device information for reference
output "compute_basic_id" {
  description = "ID of the basic compute device"
  value       = bcm_cmdevice_device.compute_basic.id
}

output "ipmi_devices" {
  description = "IPMI-enabled devices with their serial numbers"
  value = {
    compute_ipmi = {
      id            = bcm_cmdevice_device.compute_ipmi.id
      hostname      = bcm_cmdevice_device.compute_ipmi.hostname
      serial_number = bcm_cmdevice_device.compute_ipmi.serial_number
      power_control = bcm_cmdevice_device.compute_ipmi.power_control
    }
    gpu_node = {
      id            = bcm_cmdevice_device.gpu_node.id
      hostname      = bcm_cmdevice_device.gpu_node.hostname
      serial_number = bcm_cmdevice_device.gpu_node.serial_number
      power_control = bcm_cmdevice_device.gpu_node.power_control
    }
    storage_node = {
      id            = bcm_cmdevice_device.storage_node.id
      hostname      = bcm_cmdevice_device.storage_node.hostname
      serial_number = bcm_cmdevice_device.storage_node.serial_number
      power_control = bcm_cmdevice_device.storage_node.power_control
    }
  }
}

# Kubernetes cluster outputs
output "kubernetes_cluster" {
  description = "Kubernetes cluster information"
  value = {
    uuid    = bcm_cmkube_cluster.production.uuid
    name    = bcm_cmkube_cluster.production.name
    version = bcm_cmkube_cluster.production.version
  }
}

output "kubernetes_nodes" {
  description = "Kubernetes cluster node information"
  value = {
    control_plane = {
      id       = bcm_cmdevice_device.k8s_control_plane.id
      hostname = bcm_cmdevice_device.k8s_control_plane.hostname
    }
    worker = {
      id       = bcm_cmdevice_device.k8s_worker.id
      hostname = bcm_cmdevice_device.k8s_worker.hostname
    }
    combined = {
      id       = bcm_cmdevice_device.k8s_combined_cp.id
      hostname = bcm_cmdevice_device.k8s_combined_cp.hostname
    }
  }
}

output "etcd_cluster" {
  description = "Etcd cluster information"
  value = {
    uuid = bcm_cmetcd_cluster.k8s_etcd.uuid
    name = bcm_cmetcd_cluster.k8s_etcd.name
  }
}
