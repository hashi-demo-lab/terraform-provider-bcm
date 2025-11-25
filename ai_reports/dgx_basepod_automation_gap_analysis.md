# DGX BasePOD Kubernetes Deployment - Terraform Provider Gap Analysis

**Analysis Date:** 2025-11-25
**Reference:** [NVIDIA DGX BasePOD Deployment Guide - Kubernetes](https://docs.nvidia.com/dgx-basepod/deployment-guide-dgx-basepod/latest/kube-deploy.html)
**Provider:** terraform-provider-bcm

## Executive Summary

The current provider covers **~60%** of the resources needed for full DGX BasePOD Kubernetes deployment automation. Key gaps exist in:
- Device interface management (bonding, BMC) - **In Progress**
- Kubernetes cluster operators configuration
- User/credential management resource
- Node provisioning operations

## Deployment Workflow Mapping

### Phase 1: Software Image Preparation

| Step | BCM Operation | Provider Resource | Status |
|------|--------------|-------------------|--------|
| Clone default image | `cmpart.addSoftwareImage` | `bcm_cmpart_softwareimage` | ✅ Implemented |
| Add kernel modules | `cmpart.updateSoftwareImage` | `bcm_cmpart_softwareimage.modules` | ✅ Implemented |
| Set kernel parameters | `cmpart.updateSoftwareImage` | `bcm_cmpart_softwareimage.kernel_parameters` | ✅ Implemented |
| Configure disk layout | `cmpart.updateSoftwareImage` | `bcm_cmpart_softwareimage.disksetup` | ✅ Implemented |
| Query existing images | `cmpart.getSoftwareImages` | `data.bcm_cmpart_softwareimages` | ✅ Implemented |

### Phase 2: Category Configuration

| Step | BCM Operation | Provider Resource | Status |
|------|--------------|-------------------|--------|
| Create k8s-control-plane category | `cmdevice.addCategory` | `bcm_cmdevice_category` | ✅ Implemented |
| Link software image to category | `cmdevice.updateCategory` | `bcm_cmdevice_category.software_image` | ✅ Implemented |
| Set disk layout | `cmdevice.updateCategory` | `bcm_cmdevice_category.disksetup` | ✅ Implemented |
| Configure category roles | `cmdevice.updateCategory` | `bcm_cmdevice_category.roles` | ✅ Implemented |
| Query categories | `cmdevice.getCategories` | `data.bcm_cmdevice_categories` | ✅ Implemented |

### Phase 3: Node/Device Management

| Step | BCM Operation | Provider Resource | Status |
|------|--------------|-------------------|--------|
| Create/clone nodes | `cmdevice.addDevice` | `bcm_cmdevice_device` | ✅ Implemented |
| Assign category | `cmdevice.updateDevice` | `bcm_cmdevice_device.category` | ✅ Implemented |
| Set software image | `cmdevice.updateDevice` | `bcm_cmdevice_device.software_image` | ✅ Implemented |
| Set provisioning interface | `cmdevice.updateDevice` | `bcm_cmdevice_device.provisioning_interface` | ✅ Implemented |
| Set MAC address | `cmdevice.updateDevice` | `bcm_cmdevice_device.mac` | ✅ Implemented |
| Query nodes | `cmdevice.getNodes` | `data.bcm_cmdevice_nodes` | ✅ Implemented |
| Query roles | `cmdevice.getRoles` | `data.bcm_cmdevice_roles` | ✅ Implemented |

### Phase 4: Network Interface Configuration

| Step | BCM Operation | Provider Resource | Status |
|------|--------------|-------------------|--------|
| Add BMC/IPMI interface | `cmdevice.updateDevice` | `bcm_cmdevice_device.interfaces` | ⚠️ **In Progress** |
| Add physical interfaces | `cmdevice.updateDevice` | `bcm_cmdevice_device.interfaces` | ⚠️ **In Progress** |
| Create bond interface | `cmdevice.updateDevice` | `bcm_cmdevice_device.interfaces[type=bond]` | ⚠️ **In Progress** |
| Set bond mode (802.3ad) | `cmdevice.updateDevice` | `bcm_cmdevice_device.interfaces[].bond_mode` | ⚠️ **In Progress** |
| Assign to network | `cmdevice.updateDevice` | `bcm_cmdevice_device.interfaces[].network` | ⚠️ **In Progress** |
| Query interfaces | `cmdevice.getNodes` | `data.bcm_cmdevice_interfaces` | ⚠️ **In Progress** |

**Current Progress:** Device interfaces block is specified in `specs/001-cmdevice-interfaces/spec.md`

### Phase 5: Network Management

| Step | BCM Operation | Provider Resource | Status |
|------|--------------|-------------------|--------|
| Create networks | `cmnet.addNetwork` | `bcm_cmnet_network` | ✅ Implemented |
| Configure DHCP | `cmnet.updateNetwork` | `bcm_cmnet_network.dhcp_enabled` | ✅ Implemented |
| Set network range | `cmnet.updateNetwork` | `bcm_cmnet_network.base_address/netmask_bits` | ✅ Implemented |
| Configure gateway | `cmnet.updateNetwork` | `bcm_cmnet_network.gateway_ip` | ✅ Implemented |
| Query networks | `cmnet.getNetworks` | `data.bcm_cmnet_networks` | ✅ Implemented |

### Phase 6: Kubernetes Cluster Creation

| Step | BCM Operation | Provider Resource | Status |
|------|--------------|-------------------|--------|
| Create K8s cluster | `cmkube.addKubeCluster` | `bcm_cmkube_cluster` | ✅ Implemented |
| Set master nodes | `cmkube.updateKubeCluster` | `bcm_cmkube_cluster.master_nodes` | ✅ Implemented |
| Set worker nodes | `cmkube.updateKubeCluster` | `bcm_cmkube_cluster.worker_nodes` | ✅ Implemented |
| Set etcd nodes | `cmkube.updateKubeCluster` | `bcm_cmkube_cluster.etcd_nodes` | ✅ Implemented |
| Select CNI plugin | `cmkube.updateKubeCluster` | `bcm_cmkube_cluster.cni_plugin` | ✅ Implemented |
| Set management network | `cmkube.updateKubeCluster` | `bcm_cmkube_cluster.management_network` | ✅ Implemented |
| Query clusters | `cmkube.getKubeClusters` | `data.bcm_cmkube_clusters` | ✅ Implemented |
| Configure addons | `cmkube.updateKubeCluster` | `bcm_cmkube_cluster.addons` | ✅ Implemented |

### Phase 7: Operator Installation (GAP)

| Step | BCM Operation | Provider Resource | Status |
|------|--------------|-------------------|--------|
| Install GPU Operator | `cmkube.*` (TBD) | N/A | ❌ **NOT IMPLEMENTED** |
| Install Network Operator | `cmkube.*` (TBD) | N/A | ❌ **NOT IMPLEMENTED** |
| Install MPI Operator | `cmkube.*` (TBD) | N/A | ❌ **NOT IMPLEMENTED** |
| Install Kubeflow Training | `cmkube.*` (TBD) | N/A | ❌ **NOT IMPLEMENTED** |
| Install Prometheus | `cmkube.*` (TBD) | N/A | ❌ **NOT IMPLEMENTED** |
| Configure MetalLB | `cmkube.*` (TBD) | N/A | ❌ **NOT IMPLEMENTED** |

### Phase 8: User Management (GAP)

| Step | BCM Operation | Provider Resource | Status |
|------|--------------|-------------------|--------|
| Create BCM user | `cmuser.addUser` | N/A | ❌ **NOT IMPLEMENTED** |
| Set user password | `cmuser.updateUser` | N/A | ❌ **NOT IMPLEMENTED** |
| Set user groups | `cmuser.updateUser` | N/A | ❌ **NOT IMPLEMENTED** |
| Add K8s user | `cm-kubernetes-setup --add-user` | N/A | ❌ **NOT IMPLEMENTED** |
| Query users | `cmuser.getUsers` | `data.bcm_cmuser_users` | ✅ Implemented |

### Phase 9: Node Provisioning Operations (GAP)

| Step | BCM Operation | Provider Resource | Status |
|------|--------------|-------------------|--------|
| Power on node | `cmdevice.powerOn` | N/A | ❌ **NOT IMPLEMENTED** |
| Power off node | `cmdevice.powerOff` | N/A | ❌ **NOT IMPLEMENTED** |
| Reboot node | `cmdevice.reboot` | N/A | ❌ **NOT IMPLEMENTED** |
| Provision node | `cmprov.*` | N/A | ❌ **NOT IMPLEMENTED** |
| Check provision status | `cmprov.*` | N/A | ❌ **NOT IMPLEMENTED** |

---

## Gap Summary

### Currently Implemented (13 Resources/Data Sources)

| Type | Resource | Coverage |
|------|----------|----------|
| Resource | `bcm_cmpart_softwareimage` | Full CRUD + cloning |
| Resource | `bcm_cmdevice_category` | Full CRUD |
| Resource | `bcm_cmdevice_device` | Full CRUD (interfaces pending) |
| Resource | `bcm_cmnet_network` | Full CRUD |
| Resource | `bcm_cmkube_cluster` | Full CRUD |
| Data Source | `bcm_cmpart_softwareimages` | Full query + filtering |
| Data Source | `bcm_cmpart_partitions` | Full query + filtering |
| Data Source | `bcm_cmdevice_categories` | Full query + filtering |
| Data Source | `bcm_cmdevice_nodes` | Full query + filtering |
| Data Source | `bcm_cmdevice_roles` | Full query + filtering |
| Data Source | `bcm_cmnet_networks` | Full query + filtering |
| Data Source | `bcm_cmkube_clusters` | Full query + filtering |
| Data Source | `bcm_cmuser_users` | Full query + filtering |

### In Progress (2 Items)

| Type | Resource | Status | Spec Location |
|------|----------|--------|---------------|
| Data Source | `bcm_cmdevice_interfaces` | Spec Complete | `specs/001-cmdevice-interfaces/spec.md` |
| Resource Block | `bcm_cmdevice_device.interfaces` | Model Defined | `resource_cmdevice_device_interfaces.go` |

### Critical Gaps for DGX BasePOD Automation

#### Priority 1: Required for Basic Deployment

| Gap | Description | BCM Service | Effort |
|-----|-------------|-------------|--------|
| **Device Interfaces** | Bond interfaces, BMC/IPMI, physical NICs | CMDevice | 🟡 Medium (In Progress) |
| **User Resource** | Create/manage BCM users | CMUser | 🟢 Low |
| **Kubernetes User** | Add users to K8s cluster | CMKube | 🟢 Low |

#### Priority 2: Required for Full Automation

| Gap | Description | BCM Service | Effort |
|-----|-------------|-------------|--------|
| **Node Power Operations** | Power on/off, reboot | CMDevice | 🟢 Low |
| **Provisioning Status** | Check node provisioning state | CMProv | 🟡 Medium |
| **GPU Operator Config** | Enable/configure GPU operator | CMKube | 🟡 Medium |
| **Network Operator Config** | SR-IOV, IPoIB, NFD | CMKube | 🔴 High |

#### Priority 3: Enhanced Automation

| Gap | Description | BCM Service | Effort |
|-----|-------------|-------------|--------|
| **Storage Classes** | K8s storage configuration | CMKube | 🟡 Medium |
| **MetalLB Config** | Load balancer IP pools | CMKube | 🟡 Medium |
| **Certificate Management** | TLS certificates | CMCert | 🟡 Medium |
| **BeeGFS Integration** | Parallel storage | CMBeeGFS | 🔴 High |

---

## Recommended Implementation Roadmap

### Sprint 1: Complete Device Interfaces (Current)

```
[x] DeviceInterfaceModel defined
[ ] Implement interfaces block in bcm_cmdevice_device resource
[ ] Implement data.bcm_cmdevice_interfaces data source
[ ] Tests: Bond creation, BMC interface, physical NICs
```

### Sprint 2: User Management

```
[ ] Implement bcm_cmuser_user resource (CRUD)
[ ] Add Kubernetes user integration
[ ] Tests: User creation, password, groups
```

### Sprint 3: Provisioning Operations

```
[ ] Implement ephemeral resource for power operations
[ ] Add provisioning status data source
[ ] Tests: Power cycle, provision verification
```

### Sprint 4: Kubernetes Operators

```
[ ] Research CMKube operator APIs
[ ] Implement operator configuration attributes
[ ] Tests: GPU Operator, Network Operator enablement
```

---

## Example: Full DGX BasePOD Deployment with Current + Planned Resources

```hcl
# Phase 1: Software Image
resource "bcm_cmpart_softwareimage" "k8s_control_plane" {
  name             = "k8s-control-plane-image"
  original_image   = "default-image"
  kernel_parameters = "console=ttyS0,115200n8"
  modules          = ["mlx5_core", "bonding"]
}

# Phase 2: Category
resource "bcm_cmdevice_category" "k8s_control_plane" {
  name           = "k8s-control-plane"
  software_image = bcm_cmpart_softwareimage.k8s_control_plane.uuid
}

# Phase 3: Networks
resource "bcm_cmnet_network" "management" {
  name         = "managementnet"
  base_address = "10.0.0.0"
  netmask_bits = 24
  dhcp_enabled = true
}

resource "bcm_cmnet_network" "ipmi" {
  name         = "ipminet"
  base_address = "192.168.100.0"
  netmask_bits = 24
  dhcp_enabled = false
}

# Phase 4: Control Plane Nodes (REQUIRES INTERFACES BLOCK)
resource "bcm_cmdevice_device" "knode" {
  count    = 3
  hostname = "knode-0${count.index + 1}"
  category = bcm_cmdevice_category.k8s_control_plane.uuid

  # PLANNED: Device interfaces block
  interfaces {
    name    = "ipmi0"
    type    = "bmc"
    network = bcm_cmnet_network.ipmi.uuid
  }

  interfaces {
    name = "ens2f1np1"
    type = "physical"
  }

  interfaces {
    name = "ens1f1np1"
    type = "physical"
  }

  interfaces {
    name      = "bond0"
    type      = "bond"
    members   = ["ens2f1np1", "ens1f1np1"]
    bond_mode = "802.3ad"
    network   = bcm_cmnet_network.management.uuid
    bootable  = true
  }

  provisioning_interface = "bond0"
}

# Phase 5: Kubernetes Cluster
resource "bcm_cmkube_cluster" "basepod" {
  name               = "dgx-basepod"
  master_nodes       = [for node in bcm_cmdevice_device.knode : node.uuid]
  etcd_nodes         = [for node in bcm_cmdevice_device.knode : node.uuid]
  management_network = bcm_cmnet_network.management.uuid
  cni_plugin         = "calico"

  # Get DGX worker nodes from existing category
  worker_nodes = data.bcm_cmdevice_nodes.dgx_workers.nodes[*].uuid
}

# PLANNED: User Management
resource "bcm_cmuser_user" "k8s_admin" {
  username = "k8sadmin"
  password = var.k8s_admin_password
  groups   = ["wheel", "docker"]
}

# PLANNED: Kubernetes user
resource "bcm_cmkube_user" "k8s_admin" {
  cluster  = bcm_cmkube_cluster.basepod.uuid
  username = bcm_cmuser_user.k8s_admin.username
}
```

---

## BCM API Research Needed

| Service | Methods to Investigate | Purpose |
|---------|----------------------|---------|
| CMUser | `addUser`, `updateUser`, `removeUser` | User management resource |
| CMKube | `getOperators`, `enableOperator`, `configureOperator` | Operator management |
| CMProv | `provisionNode`, `getProvisioningStatus` | Provisioning operations |
| CMDevice | `powerOn`, `powerOff`, `powerCycle`, `powerStatus` | Power management |

---

## Conclusion

The provider has strong foundational coverage for BCM resources. The **device interfaces block** (currently in progress) is the most critical gap blocking DGX BasePOD automation. After completing that, user management would unlock the remaining deployment steps.

**Automation Coverage:**
- **Current:** ~60% (core infrastructure)
- **After Interfaces:** ~75% (full node configuration)
- **After Users + Power:** ~90% (end-to-end deployment)
- **After Operators:** ~100% (full production automation)
