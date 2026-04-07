---
page_title: "bcm_cmdevice_device Resource - bcm"
subcategory: ""
description: |-
  Manages a BCM device (compute node) in the cluster.
  Devices represent physical or virtual nodes that can be provisioned and managed by BCM. Each device requires a unique hostname, category assignment, and at least one network interface.
---

# bcm_cmdevice_device (Resource)

Manages a BCM device (compute node) in the cluster.

Devices represent physical or virtual nodes that can be provisioned and managed by BCM. Each device requires a unique hostname, category assignment, and at least one network interface.

## Example Usage

```terraform
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
  boot_loader       = "PXELINUX"
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
  boot_loader       = "PXELINUX"
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
  boot_loader       = "PXELINUX"
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
  boot_loader       = "PXELINUX"
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
```

### Basic Device

```terraform
# Generate unique suffix for this test run
locals {
  test_suffix = formatdate("YYYYMMDDhhmmss", timestamp())
}

# Lookup management network
data "bcm_cmnet_networks" "management" {
  filter {
    name_pattern = "managementnet"
  }
}

# Create a unique software image for each test run
resource "bcm_cmpart_softwareimage" "test_image" {
  name = "citest-image-${local.test_suffix}"
  path = "/cm/images/ubuntu-22.04-server-amd64.iso"
}

# Create a unique category for each test run
resource "bcm_cmdevice_category" "basic_category" {
  name               = "citest-category-${local.test_suffix}"
  management_network = data.bcm_cmnet_networks.management.networks[0].id

  software_image_proxy = {
    parent_software_image = bcm_cmpart_softwareimage.test_image.id
  }

  notes = "Test category (run: ${local.test_suffix})"

  depends_on = [bcm_cmpart_softwareimage.test_image]
}

# Example: Basic compute device with minimal configuration
# Note: partition is NOT specified because the category has a software_image_proxy
# which provides the partition automatically
resource "bcm_cmdevice_device" "basic" {
  hostname = "citest-device-${local.test_suffix}"
  category = bcm_cmdevice_category.basic_category.id

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = "00:11:22:33:44:AA"
    network  = data.bcm_cmnet_networks.management.networks[0].id
    bootable = true
    dhcp     = true
  }

  notes = "Basic test device (run: ${local.test_suffix})"

  depends_on = [bcm_cmdevice_category.basic_category]
}

# Outputs
output "device_id" {
  value       = bcm_cmdevice_device.basic.id
  description = "UUID of the created device"
}

output "device_hostname" {
  value       = bcm_cmdevice_device.basic.hostname
  description = "Hostname of the created device"
}

output "test_suffix" {
  value       = local.test_suffix
  description = "Unique suffix for this test run"
}
```

### IPMI-Enabled Device

```terraform
# Lookup existing category by name
data "bcm_cmdevice_categories" "default" {
  name = "default"
}

# Lookup management network
data "bcm_cmnet_networks" "management" {
  filter {
    name_pattern = "DefaultEthernet"
  }
}

# Example: IPMI-enabled device with power control and network configuration
resource "bcm_cmdevice_device" "ipmi" {
  hostname = "citest-compute-ipmi"
  category = try(data.bcm_cmdevice_categories.default.categories[0].id, null)

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = "00:11:22:33:44:BB"
    network  = try(data.bcm_cmnet_networks.management.networks[0].id, null)
    bootable = true
    dhcp     = true
  }

  # Power control via IPMI
  power_control = "ipmi"

  # Network gateway configuration
  default_gateway        = "192.168.1.1"
  default_gateway_metric = 100

  # Boot configuration
  boot_loader       = "PXELINUX"
  kernel_parameters = "console=ttyS0,115200 ipmi_si.type=kcs"

  # Hardware identifiers (typically auto-discovered from BMC)
  serial_number = "SN123456789"
  part_number   = "PN-COMPUTE-001"

  notes = "IPMI-enabled compute node with power management"
}
```

### Device with Roles

```terraform
# Example: Device with role assignments
#
# This example demonstrates how to create a device with specific roles assigned.
# Roles define the device's function in the cluster (boot, headnode, etc.).
# Use the bcm_cmdevice_roles data source to discover available roles, then assign by name.

# Generate unique suffix for this test run
locals {
  test_suffix = formatdate("YYYYMMDDhhmmss", timestamp())
}

# Discover available roles in the cluster
data "bcm_cmdevice_roles" "all" {}

# Transform roles data into a lookup map by name
# This enables easy reference: local.roles["boot"] returns "boot"
locals {
  # Map of role name -> role name for validation via data source
  roles = { for r in data.bcm_cmdevice_roles.all.roles : r.name => r.name }

  # Define which roles to assign (validated against the data source)
  device_roles = [
    local.roles["boot"],
    local.roles["headnode"],
  ]
}

# Lookup management network
data "bcm_cmnet_networks" "management" {
  filter {
    name_pattern = "managementnet"
  }
}

# Create a unique software image for each test run
resource "bcm_cmpart_softwareimage" "test_image" {
  name = "citest-image-${local.test_suffix}"
  path = "/cm/images/ubuntu-22.04-server-amd64.iso"
}

# Create a unique category for each test run
resource "bcm_cmdevice_category" "roles_category" {
  name               = "citest-category-${local.test_suffix}"
  management_network = data.bcm_cmnet_networks.management.networks[0].id

  software_image_proxy = {
    parent_software_image = bcm_cmpart_softwareimage.test_image.id
  }

  notes = "Test category for device with roles (run: ${local.test_suffix})"

  depends_on = [bcm_cmpart_softwareimage.test_image]
}

# Create device with roles assigned BY NAME
resource "bcm_cmdevice_device" "with_roles" {
  hostname = "citest-roles-${local.test_suffix}"
  category = bcm_cmdevice_category.roles_category.id

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = "00:11:22:33:44:BB"
    network  = data.bcm_cmnet_networks.management.networks[0].id
    bootable = true
    dhcp     = true
  }

  # Assign roles by name - provider validates against BCM cluster
  roles = local.device_roles

  notes = "Device with roles assigned by name (run: ${local.test_suffix})"

  depends_on = [bcm_cmdevice_category.roles_category]
}

# Outputs
output "device_id" {
  value       = bcm_cmdevice_device.with_roles.id
  description = "UUID of the created device"
}

output "device_hostname" {
  value       = bcm_cmdevice_device.with_roles.hostname
  description = "Hostname of the created device"
}

output "device_roles" {
  value       = bcm_cmdevice_device.with_roles.roles
  description = "Role names assigned to the device"
}

output "available_roles" {
  value       = keys(local.roles)
  description = "All available role names in the cluster"
}
```

### Import and Recategorize

```terraform
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# =============================================================================
# Example: Import existing device and move to a new category
# =============================================================================
# A common workflow when adopting Terraform for an existing BCM cluster:
# import a device that was created outside Terraform, then change its
# category to bring it under Terraform-managed configuration.
#
# Step 1: terraform apply -target=data.bcm_cmdevice_nodes.server
# Step 2: terraform import bcm_cmdevice_device.server <device-uuid>
# Step 3: terraform plan (shows category change)
# Step 4: terraform apply (moves device to new category)

# Lookup the existing device to discover its current MAC
data "bcm_cmdevice_nodes" "server" {
  filter {
    hostname_pattern = "node001"
  }
}

# Lookup the network for interfaces
data "bcm_cmnet_networks" "management" {
  filter {
    name_pattern = "managementnet"
  }
}

# Lookup the software image for the new category
data "bcm_cmpart_softwareimages" "all" {}

# Create the target category that the device will be moved into
resource "bcm_cmdevice_category" "gpu_compute" {
  name               = "gpu-compute"
  management_network = data.bcm_cmnet_networks.management.networks[0].id

  software_image_proxy = {
    parent_software_image = data.bcm_cmpart_softwareimages.all.images[0].uuid
  }
}

# The device — after import, the category change triggers an update in BCM
resource "bcm_cmdevice_device" "server" {
  hostname = data.bcm_cmdevice_nodes.server.nodes[0].hostname
  category = bcm_cmdevice_category.gpu_compute.id

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = data.bcm_cmdevice_nodes.server.nodes[0].mac
    network  = data.bcm_cmnet_networks.management.networks[0].id
    bootable = true
    dhcp     = true
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `category` (String) Category UUID reference
- `hostname` (String) Device hostname (RFC 1123 DNS label: lowercase alphanumeric and hyphens, 1-63 chars)

### Optional

- `boot_loader` (String) Boot loader type (e.g., SYSLINUX, GRUB) - defaults to category value
- `boot_loader_protocol` (String) Boot loader protocol (e.g., HTTP, TFTP) - defaults to category value
- `default_gateway` (String) Default gateway IP address for the device
- `default_gateway_metric` (Number) Default gateway metric/priority (lower is preferred)
- `etcd_host_role` (Block List) Etcd host role configuration. Defines this device as a member of an EtcdCluster. Each etcd_host_role block associates the device with one EtcdCluster as an etcd node. (see [below for nested schema](#nestedblock--etcd_host_role))
- `force` (Boolean) Force operation (override BCM validation warnings)
- `interfaces` (Block List) Network interface configurations for the device. At least one interface is required. Each interface can be physical, bond, or BMC type. (see [below for nested schema](#nestedblock--interfaces))
- `kernel_parameters` (String) Kernel boot parameters
- `kubelet_role` (Block List) Kubernetes kubelet role configuration. Defines this device as a member of a KubeCluster. Each kubelet_role block associates the device with one KubeCluster as a control plane node, worker node, or both. (see [below for nested schema](#nestedblock--kubelet_role))
- `mac` (String) Device MAC address, computed from the first interface. Can be set explicitly to override.
- `management_network` (String) Management network UUID reference. Optional — if not specified, the device has no management network set.
- `notes` (String) Device notes/description
- `part_number` (String) Hardware part number
- `partition` (String) Partition UUID reference (uses category default if not specified)
- `power_control` (String) Power control method (e.g., 'none', 'ipmi', 'ipdu', 'redfish')
- `roles` (Set of String) Set of role names assigned to this device. Roles define the device's function in the cluster (e.g., "backup", "provisioning", "boot"). Use the `bcm_cmdevice_roles` data source to discover available roles. **Only role names are accepted** (not UUIDs). Role names are case-sensitive.

Example usage:

```hcl
# Discover available roles
data "bcm_cmdevice_roles" "all" {}

resource "bcm_cmdevice_device" "node" {
  # ... other configuration ...
  roles = [data.bcm_cmdevice_roles.all.roles[0].name]
}
```
- `serial_number` (String) Hardware serial number

### Read-Only

- `base_type` (String) Entity base type (always 'Device')
- `child_type` (String) Device type (HeadNode, ComputeNode, PhysicalNode, etc.)
- `creation_time` (Number) Device creation timestamp (Unix epoch)
- `id` (String) Device identifier (same as UUID)
- `uuid` (String) Device UUID assigned by BCM

<a id="nestedblock--etcd_host_role"></a>
### Nested Schema for `etcd_host_role`

Required:

- `etcd_cluster` (String) UUID of the EtcdCluster this device belongs to.

Optional:

- `advertise_client_urls` (List of String) URLs to advertise to clients for connecting to this member.
- `advertise_peer_urls` (List of String) URLs to advertise to peers for connecting to this member.
- `listen_client_urls` (List of String) URLs etcd listens on for client traffic.
- `listen_peer_urls` (List of String) URLs etcd listens on for peer traffic.
- `max_snapshots` (Number) Maximum number of snapshot files to retain. Default: 5.
- `member_name` (String) Etcd member name. Default: '$hostname' (uses device hostname).
- `snapshot_count` (Number) Number of committed transactions to trigger a snapshot. Default: 100000.
- `spool` (String) Etcd data directory path. Default: '/var/lib/etcd'.

Read-Only:

- `uuid` (String) BCM-assigned role UUID.


<a id="nestedblock--interfaces"></a>
### Nested Schema for `interfaces`

Required:

- `name` (String) Interface name (e.g., 'eth0', 'bond0', 'ipmi'). Must be unique within the device.
- `type` (String) Interface type: 'physical', 'bond', or 'bmc'.

Optional:

- `bond_mode` (String) Bond mode (e.g., '802.3ad', 'active-backup', 'balance-rr'). Only applicable when type is 'bond'.
- `bootable` (Boolean) Enable PXE boot capability. Default: false. First bootable interface becomes provisioning interface.
- `dhcp` (Boolean) Enable DHCP for IP assignment. Default: true.
- `ip` (String) Static IPv4 address.
- `ipv6_ip` (String) Static IPv6 address.
- `mac` (String) MAC address (format: 00:11:22:33:44:55). Required for physical interfaces on create.
- `members` (List of String) Member interface names for bond type. Required when type is 'bond'.
- `network` (String) Network UUID reference for interface assignment.
- `start_if` (String) Interface startup condition: 'ALWAYS', 'NEVER', 'HOTPLUG'. Default: 'ALWAYS'.

Read-Only:

- `base_type` (String) Entity base type (always 'NetworkInterface').
- `cardtype` (String) Hardware card type (Ethernet, InfiniBand, BMC).
- `child_type` (String) Interface type (NetworkPhysicalInterface, NetworkBondInterface, NetworkBMCInterface).
- `uuid` (String) BCM-assigned interface UUID.


<a id="nestedblock--kubelet_role"></a>
### Nested Schema for `kubelet_role`

Required:

- `kube_cluster` (String) UUID of the KubeCluster this device belongs to.

Optional:

- `container_runtime_service` (String) Container runtime service name (e.g., 'docker.service', 'containerd.service'). Default: 'docker.service'.
- `control_plane` (Boolean) Whether this node runs control plane components (API server, controller manager, scheduler). Default: true.
- `custom_yaml` (String) Custom kubelet configuration YAML.
- `max_pods` (Number) Maximum number of pods that can run on this node. Default: 110.
- `options` (String) Additional kubelet options as JSON string.
- `worker` (Boolean) Whether this node can schedule workload pods. Default: true.

Read-Only:

- `uuid` (String) BCM-assigned role UUID.

## Import

Import is supported using the following syntax:

```terraform
# Lookup existing category and network
data "bcm_cmdevice_categories" "default" {
  name = "default"
}

data "bcm_cmnet_networks" "management" {
  filter {
    name_pattern = "DefaultEthernet"
  }
}

# =============================================================================
# Example 1: Import with known values
# =============================================================================
# When you know the device's MAC and network, specify them directly.
#
# Step 1: terraform import bcm_cmdevice_device.known <device-uuid>
# Step 2: terraform plan (verify no unexpected changes)

resource "bcm_cmdevice_device" "known" {
  hostname = "citest-import-known"
  category = one(data.bcm_cmdevice_categories.default.categories[*].id)

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = "00:11:22:33:44:CC"
    network  = one(data.bcm_cmnet_networks.management.networks[*].id)
    bootable = true
    dhcp     = true
  }
}

# =============================================================================
# Example 2: Import using data source lookup
# =============================================================================
# When importing an existing device, use the nodes data source to discover
# the device's current MAC and interface configuration. This avoids
# hardcoding values that may differ from what BCM has.
#
# Step 1: terraform apply -target=data.bcm_cmdevice_nodes.existing
# Step 2: terraform import bcm_cmdevice_device.discovered <device-uuid>
# Step 3: terraform plan (verify no unexpected changes)

data "bcm_cmdevice_nodes" "existing" {
  filter {
    hostname_pattern = "existing-server-01"
  }
}

resource "bcm_cmdevice_device" "discovered" {
  hostname = data.bcm_cmdevice_nodes.existing.nodes[0].hostname
  category = one(data.bcm_cmdevice_categories.default.categories[*].id)

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = data.bcm_cmdevice_nodes.existing.nodes[0].mac
    network  = one(data.bcm_cmnet_networks.management.networks[*].id)
    bootable = true
    dhcp     = true
  }
}
```

### Generate Configuration (Terraform 1.5+)

```terraform
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# =============================================================================
# Example: Generate Terraform configuration from existing BCM device
# =============================================================================
# Terraform 1.5+ supports generating HCL configuration from existing resources.
# This is useful for adopting existing infrastructure into Terraform management.
#
# Step 1: Add the import block below to your configuration
# Step 2: Run: terraform plan -generate-config-out=generated_device.tf
# Step 3: Review and adjust the generated configuration
# Step 4: Run: terraform plan (verify no unexpected changes)
# Step 5: Remove the import block and move the resource to your main config
#
# Equivalent CLI command:
#   terraform import bcm_cmdevice_device.example "1b4e8f2a-6c3d-4e7f-9a1b-2c3d4e5f6a7b"

import {
  to = bcm_cmdevice_device.example
  id = "1b4e8f2a-6c3d-4e7f-9a1b-2c3d4e5f6a7b"
}
```
