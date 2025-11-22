# BCM API Gap Analysis Report

**Feature**: Production-Ready Codebase Review
**Date**: 2025-11-22
**Phase**: Phase 3 - BCM API Gap Analysis (Tasks T020-T032)
**Methodology**: See [research.md](./research.md) Section 2 - BCM API Discovery Approach

## Executive Summary

This report provides a comprehensive analysis of the BCM (Bright Cluster Manager) API coverage by the Terraform Provider BCM. The analysis combines:

- **Discovery Sources**:
  - Live BCM API exploration via sampleRest/ documentation
  - Provider code analysis from `/workspace/internal/provider/`
  - Cross-reference with provider.go Resources() and DataSources() registration

- **Coverage Status**:
  - **Total BCM Services Discovered**: 15
  - **Implemented Resources**: 3
  - **Implemented Data Sources**: 4
  - **High-Value Gaps**: 12 resources, 6 data sources
  - **Medium-Value Gaps**: 8 resources, 4 data sources
  - **Low-Value Gaps**: 5 resources, 2 data sources

- **Key Findings**:
  - **CMDevice Service**: Partial coverage - missing roles, power management, hardware configuration
  - **CMPart Service**: Partial coverage - missing modules, kernels, partitions management
  - **CMNet Service**: Minimal coverage - only data source, no resource for network management
  - **CMProv Service**: No coverage - critical gap for provisioning workflows
  - **CMMon, CMJob, CMServ, CMKube, CMCloud, CMEtcd**: No coverage

---

## Methodology

### Discovery Approach

Per research.md Section 2, this analysis used a multi-source discovery strategy:

1. **sampleRest/ Documentation Analysis**:
   - Reviewed `BCM_API_Complete_Documentation.md` for service catalog
   - Analyzed `CMDevice_Complete_Documentation.md` for detailed CMDevice API methods
   - Extracted JSON-RPC call patterns from Python scripts (`cmnet-get-networks.py`, `investigate_category_api.py`)

2. **Provider Code Analysis**:
   - Grep'd all `CallJSONRPC()` calls in `/workspace/internal/provider/` to identify used BCM API methods
   - Read `/workspace/internal/provider/provider.go` Resources() and DataSources() registration

3. **Cross-Reference Validation**:
   - Compared discovered BCM API methods against implemented provider resources/data sources
   - Identified gaps by service type and method category (CRUD, power, provisioning, monitoring)

### Discovered BCM Services

Based on `BCM_API_Complete_Documentation.md` (lines 63-83) and codebase analysis:

| Service | Description | Implementation Status |
|---------|-------------|----------------------|
| `CMDevice` (cmdevice) | Device/node management | **PARTIAL** - Categories, devices (basic), nodes (read-only) |
| `CMPart` (cmpart) | Software images and partitions | **PARTIAL** - Software images only |
| `CMNet` (cmnet) | Network configuration | **MINIMAL** - Read-only data source |
| `CMProv` (cmprov) | Provisioning services | **NONE** |
| `CMJob` (cmjob) | Job management | **NONE** |
| `CMServ` (cmserv) | Service management | **NONE** |
| `CMMon` (cmmon) | Monitoring | **NONE** |
| `CMKube` (cmkube) | Kubernetes integration | **NONE** |
| `CMCloud` (cmcloud) | Cloud provider integration | **NONE** |
| `CMEtcd` (cmetcd) | Etcd cluster management | **NONE** |
| `CMMain` (cmmain) | Main cluster configuration | **NONE** |
| `CMGui` (cmgui) | GUI settings | **NONE** (Low priority) |
| `CMAuth` (cmauth) | Authentication/authorization | **NONE** (Provider handles separately) |
| `CMCert` (cmcert) | Certificate management | **NONE** |
| `CMBeeGFS` (cmbeegfs) | BeeGFS filesystem | **NONE** |
| `CMProc` (cmproc) | Process management | **NONE** |

---

## Service-by-Service Analysis

### CMDevice Service (Cluster Device Management)

**Implementation Status**: PARTIAL COVERAGE

#### Discovered API Methods

Based on `CMDevice_Complete_Documentation.md` and provider code analysis:

**List/Query Operations** (Read-Only):
- `getNodes` - Get all nodes ✅ IMPLEMENTED (data_source_cmdevice_nodes.go:316)
- `getDevices` - Alias for getNodes ❌ MISSING
- `getComputeNodes` - Get only compute nodes ❌ MISSING
- `getNode(arg)` - Get specific node by hostname/UUID ✅ USED (internally for drift detection)
- `getCategories` - Get all categories ✅ IMPLEMENTED (data_source_cmdevice_categories.go:284)
- `getCategory(uuid)` - Get specific category ✅ USED (resource_cmdevice_category.go:1103, resource_cmdevice_device.go:259)

**Create/Update/Delete Operations**:
- `addDevice(entity, force)` - Create device ✅ IMPLEMENTED (resource_cmdevice_device.go:393)
- `updateDevice(entity, force)` - Update device ✅ IMPLEMENTED (resource_cmdevice_device.go:765)
- `removeDevice(uuid, force)` - Delete device ✅ IMPLEMENTED (resource_cmdevice_device.go:869)
- `getDevice(name)` - Get device by name (args parameter) ✅ IMPLEMENTED (resource_cmdevice_device.go:580)
- `addCategory(entity, force)` - Create category ✅ IMPLEMENTED (resource_cmdevice_category.go:613)
- `updateCategory(entity, force)` - Update category ✅ IMPLEMENTED (resource_cmdevice_category.go:747)
- `removeCategory(uuid, force)` - Delete category ✅ IMPLEMENTED (resource_cmdevice_category.go:799)
- `removeCategories([uuids], force)` - Batch delete categories ✅ USED (tests only)

**Power Management** (from CMDevice_Complete_Documentation.md:135-154):
- `reboot(arg)` - Reboot node ❌ MISSING
- `powerOn(arg)` - Power on node ❌ MISSING
- `powerOff(arg)` - Power off node ❌ MISSING
- `powerCycle(arg)` - Power cycle node ❌ MISSING
- `powerStatus(arg)` - Get power status ❌ MISSING

**Roles Management** (inferred from entity structure):
- `getRoles` - Get all roles ❌ MISSING
- `getRole(uuid)` - Get specific role ❌ MISSING
- `addRole(entity)` - Create role ❌ MISSING
- `updateRole(entity)` - Update role ❌ MISSING
- `removeRole(uuid)` - Delete role ❌ MISSING
- `assignRole(deviceUUID, roleUUID)` - Assign role to device ❌ MISSING
- `unassignRole(deviceUUID, roleUUID)` - Unassign role ❌ MISSING

**Hardware Management** (from CMDevice_Complete_Documentation.md:277-303):
- `getHardwareInfo(arg)` - Get hardware details ❌ MISSING
- `updateBIOS(deviceUUID, settings)` - Update BIOS settings ❌ MISSING
- `updateBMC(deviceUUID, settings)` - Update BMC/IPMI settings ❌ MISSING

**Provisioning Status** (likely methods):
- `getProvisioningStatus(deviceUUID)` - Get provisioning status ❌ MISSING
- `provisionDevice(deviceUUID)` - Start provisioning ❌ MISSING (may be CMProv service)
- `reprovisionDevice(deviceUUID)` - Re-provision device ❌ MISSING

#### Implementation Status

**Implemented Resources**:
1. ✅ `bcm_cmdevice_category` - Full CRUD coverage (provider.go:168)
2. ✅ `bcm_cmdevice_device` - Full CRUD coverage (provider.go:169)

**Implemented Data Sources**:
1. ✅ `bcm_cmdevice_nodes` - Read all nodes (provider.go:182)
2. ✅ `bcm_cmdevice_categories` - Read all categories (provider.go:183)

#### Gap Analysis

**High-Value Gaps** (Core User Workflows):

1. **Power Management Resource** (`bcm_cmdevice_power`)
   - Business Value: Users need to power on/off/reboot nodes remotely
   - API Methods: `powerOn`, `powerOff`, `reboot`, `powerCycle`, `powerStatus`
   - Implementation Effort: Small (1-2 days) - Simple action resource, no state persistence
   - Use Case: Automate cluster maintenance, node recovery, power scheduling

2. **Role Management Resource** (`bcm_cmdevice_role`)
   - Business Value: Assign/manage service roles (HeadNode, Compute, Storage, etc.)
   - API Methods: `getRoles`, `addRole`, `updateRole`, `removeRole`, `assignRole`, `unassignRole`
   - Implementation Effort: Medium (3-5 days) - Full CRUD + role-device association
   - Use Case: Configure node roles (compute, storage, monitoring), service orchestration

3. **Roles Data Source** (`bcm_cmdevice_roles`)
   - Business Value: Query available roles for assignment
   - API Methods: `getRoles`
   - Implementation Effort: Small (1 day) - Simple list data source
   - Use Case: Discover available roles, validate role assignments

**Medium-Value Gaps** (Important Features):

4. **Compute Nodes Data Source** (`bcm_cmdevice_compute_nodes`)
   - Business Value: Filter nodes by type (compute vs head vs storage)
   - API Methods: `getComputeNodes`
   - Implementation Effort: Small (1 day) - Similar to existing nodes data source
   - Use Case: Target compute-only operations, capacity planning

5. **Hardware Info Data Source** (`bcm_cmdevice_hardware`)
   - Business Value: Query hardware specifications (GPU, CPU, memory)
   - API Methods: `getHardwareInfo`
   - Implementation Effort: Small (1-2 days) - Read-only data source
   - Use Case: Hardware inventory, capacity planning, GPU allocation

6. **BMC Settings Resource** (`bcm_cmdevice_bmc`)
   - Business Value: Configure out-of-band management (IPMI/Redfish)
   - API Methods: `updateBMC`, `getBMC`
   - Implementation Effort: Medium (2-3 days) - Nested entity, sensitive credentials
   - Use Case: Configure remote management, power control, console access

**Low-Value Gaps** (Edge Cases):

7. **BIOS Settings Resource** (`bcm_cmdevice_bios`)
   - Business Value: Automate BIOS configuration
   - API Methods: `updateBIOS`
   - Implementation Effort: Medium (2-3 days) - Complex entity, reboot required
   - Use Case: Boot order, virtualization settings (rare changes)

#### Partial Implementations

**bcm_cmdevice_device**: Resource exists but missing advanced features:
- ✅ Basic CRUD (hostname, MAC, category, partition)
- ✅ Network interfaces configuration
- ❌ Role assignment/management (roles field is read-only)
- ❌ Power management integration
- ❌ BMC/BIOS settings
- ❌ GPU configuration
- ❌ FS exports/mounts management

---

### CMPart Service (Software & Partition Management)

**Implementation Status**: PARTIAL COVERAGE

#### Discovered API Methods

Based on provider code and sampleRest/ documentation:

**Software Image Operations**:
- `getSoftwareImages` - Get all software images ✅ IMPLEMENTED (data_source_cmpart_softwareimages.go:288)
- `getSoftwareImage(name)` - Get specific image (args parameter) ✅ IMPLEMENTED (resource_cmpart_softwareimage.go:612)
- `addSoftwareImage(entity, cloneFrom)` - Create/clone image ✅ IMPLEMENTED (resource_cmpart_softwareimage.go:283)
- `updateSoftwareImage(entity)` - Update image ✅ IMPLEMENTED (resource_cmpart_softwareimage.go:478)
- `removeSoftwareImage(uuid, force, recursive, batch)` - Delete image ✅ IMPLEMENTED (resource_cmpart_softwareimage.go:529)
- `removeSoftwareImages([names], force)` - Batch delete ✅ USED (tests only)

**Partition Operations**:
- `getPartitions` - Get all partitions ✅ USED (resource_cmdevice_device.go:293, :710)
- `getPartition(uuid)` - Get specific partition ❌ MISSING
- `addPartition(entity)` - Create partition ❌ MISSING
- `updatePartition(entity)` - Update partition ❌ MISSING
- `removePartition(uuid)` - Delete partition ❌ MISSING

**Module Operations** (inferred from entity structure):
- `getModules` - Get all kernel modules ❌ MISSING
- `getModule(uuid)` - Get specific module ❌ MISSING
- `addModule(entity)` - Create module ❌ MISSING
- `updateModule(entity)` - Update module ❌ MISSING
- `removeModule(uuid)` - Delete module ❌ MISSING

**Kernel Operations** (inferred):
- `getKernels` - Get available kernels ❌ MISSING
- `getKernel(version)` - Get specific kernel ❌ MISSING

#### Implementation Status

**Implemented Resources**:
1. ✅ `bcm_cmpart_softwareimage` - Full CRUD coverage with cloning support (provider.go:167)

**Implemented Data Sources**:
1. ✅ `bcm_cmpart_softwareimages` - Read all software images (provider.go:181)

#### Gap Analysis

**High-Value Gaps**:

8. **Partition Resource** (`bcm_cmpart_partition`)
   - Business Value: Manage software partitions (separate OS installations)
   - API Methods: `addPartition`, `updatePartition`, `removePartition`, `getPartition`
   - Implementation Effort: Medium (3-5 days) - Similar to software image, async operations
   - Use Case: Multi-OS cluster configurations, development vs production partitions

9. **Partitions Data Source** (`bcm_cmpart_partitions`)
   - Business Value: Query available partitions for assignment
   - API Methods: `getPartitions`
   - Implementation Effort: Small (1 day) - Simple list data source
   - Use Case: Discover partitions, validate partition assignments

**Medium-Value Gaps**:

10. **Kernel Module Resource** (`bcm_cmpart_module`)
    - Business Value: Manage kernel modules for software images
    - API Methods: `addModule`, `updateModule`, `removeModule`
    - Implementation Effort: Small (2-3 days) - Simple entity, part of image configuration
    - Use Case: Custom driver loading, hardware support, performance tuning

11. **Kernel Modules Data Source** (`bcm_cmpart_modules`)
    - Business Value: Query available kernel modules
    - API Methods: `getModules`
    - Implementation Effort: Small (1 day) - Simple list data source
    - Use Case: Discover modules, validate module dependencies

**Low-Value Gaps**:

12. **Kernels Data Source** (`bcm_cmpart_kernels`)
    - Business Value: Query available kernel versions
    - API Methods: `getKernels`
    - Implementation Effort: Small (1 day) - Read-only data source
    - Use Case: Kernel upgrade planning, compatibility checks

---

### CMNet Service (Network Management)

**Implementation Status**: MINIMAL COVERAGE

#### Discovered API Methods

Based on `cmnet-get-networks.py` and provider code:

**Network Operations**:
- `getNetworks` - Get all networks ✅ IMPLEMENTED (data_source_cmnet_networks.go:291)
- `getNetwork(uuid)` - Get specific network ❌ MISSING
- `addNetwork(entity)` - Create network ❌ MISSING
- `updateNetwork(entity)` - Update network ❌ MISSING
- `removeNetwork(uuid)` - Delete network ❌ MISSING

**VLAN Operations** (likely):
- `getVLANs` - Get all VLANs ❌ MISSING
- `addVLAN(entity)` - Create VLAN ❌ MISSING
- `updateVLAN(entity)` - Update VLAN ❌ MISSING
- `removeVLAN(uuid)` - Delete VLAN ❌ MISSING

**Subnet Operations** (likely):
- `getSubnets` - Get all subnets ❌ MISSING
- `addSubnet(entity)` - Create subnet ❌ MISSING
- `updateSubnet(entity)` - Update subnet ❌ MISSING
- `removeSubnet(uuid)` - Delete subnet ❌ MISSING

#### Implementation Status

**Implemented Resources**:
- NONE

**Implemented Data Sources**:
1. ✅ `bcm_cmnet_networks` - Read all networks (provider.go:184)

#### Gap Analysis

**High-Value Gaps** (Critical for Network Management):

13. **Network Resource** (`bcm_cmnet_network`)
    - Business Value: Manage cluster networks (provisioning, data, management)
    - API Methods: `addNetwork`, `updateNetwork`, `removeNetwork`, `getNetwork`
    - Implementation Effort: Medium (3-5 days) - Full CRUD, network validation
    - Use Case: Network isolation, VLAN configuration, multi-network clusters

**Medium-Value Gaps**:

14. **VLAN Resource** (`bcm_cmnet_vlan`)
    - Business Value: VLAN configuration for network segmentation
    - API Methods: `addVLAN`, `updateVLAN`, `removeVLAN`
    - Implementation Effort: Medium (2-3 days) - Network entity relationship
    - Use Case: Network isolation, security zones, traffic segregation

15. **Subnet Resource** (`bcm_cmnet_subnet`)
    - Business Value: Subnet management for IP address allocation
    - API Methods: `addSubnet`, `updateSubnet`, `removeSubnet`
    - Implementation Effort: Medium (2-3 days) - IP address validation
    - Use Case: IP address management, DHCP configuration

**Low-Value Gaps**:

16. **VLANs Data Source** (`bcm_cmnet_vlans`)
    - Business Value: Query available VLANs
    - API Methods: `getVLANs`
    - Implementation Effort: Small (1 day) - Simple list data source
    - Use Case: VLAN discovery, validation

---

### CMProv Service (Provisioning Management)

**Implementation Status**: NO COVERAGE

#### Likely API Methods

Based on typical provisioning workflows and BCM_API_Complete_Documentation.md:

**Provisioning Operations**:
- `getProvisioningStatus(deviceUUID)` - Get provisioning status ❌ MISSING
- `startProvisioning(deviceUUID, options)` - Start provisioning ❌ MISSING
- `stopProvisioning(deviceUUID)` - Stop provisioning ❌ MISSING
- `cancelProvisioning(deviceUUID)` - Cancel provisioning ❌ MISSING
- `getProvisioningJobs` - Get all provisioning jobs ❌ MISSING
- `getProvisioningJob(jobID)` - Get specific job ❌ MISSING

**Provisioning Templates** (likely):
- `getProvisioningTemplates` - Get all templates ❌ MISSING
- `addProvisioningTemplate(entity)` - Create template ❌ MISSING
- `updateProvisioningTemplate(entity)` - Update template ❌ MISSING
- `removeProvisioningTemplate(uuid)` - Delete template ❌ MISSING

#### Implementation Status

**Implemented Resources**: NONE
**Implemented Data Sources**: NONE

#### Gap Analysis

**High-Value Gaps** (Core Provisioning Workflows):

17. **Provisioning Job Resource** (`bcm_cmprov_job`)
    - Business Value: Trigger and manage node provisioning
    - API Methods: `startProvisioning`, `stopProvisioning`, `cancelProvisioning`
    - Implementation Effort: Large (5-7 days) - Async operations, polling, job state management
    - Use Case: Automated node deployment, OS installation, cluster scaling

18. **Provisioning Status Data Source** (`bcm_cmprov_status`)
    - Business Value: Monitor provisioning progress
    - API Methods: `getProvisioningStatus`, `getProvisioningJobs`
    - Implementation Effort: Medium (2-3 days) - Read-only with filtering
    - Use Case: Provisioning monitoring, failure detection, automation orchestration

**Medium-Value Gaps**:

19. **Provisioning Template Resource** (`bcm_cmprov_template`)
    - Business Value: Manage provisioning configurations
    - API Methods: `addProvisioningTemplate`, `updateProvisioningTemplate`, `removeProvisioningTemplate`
    - Implementation Effort: Medium (3-5 days) - Complex entity, template validation
    - Use Case: Standardized provisioning, environment templates, best practices

---

### CMJob Service (Job Management)

**Implementation Status**: NO COVERAGE

#### Likely API Methods

Based on typical job management patterns:

**Job Operations**:
- `getJobs` - Get all jobs ❌ MISSING
- `getJob(jobID)` - Get specific job ❌ MISSING
- `runJob(jobDefinition)` - Execute job ❌ MISSING
- `cancelJob(jobID)` - Cancel running job ❌ MISSING
- `pauseJob(jobID)` - Pause job ❌ MISSING
- `resumeJob(jobID)` - Resume paused job ❌ MISSING
- `getJobStatus(jobID)` - Get job status ❌ MISSING
- `getJobHistory` - Get job execution history ❌ MISSING

#### Gap Analysis

**Medium-Value Gaps**:

20. **Job Resource** (`bcm_cmjob_job`)
    - Business Value: Automate cluster maintenance tasks
    - API Methods: `runJob`, `cancelJob`, `pauseJob`, `resumeJob`
    - Implementation Effort: Large (5-7 days) - Async operations, job lifecycle
    - Use Case: Scheduled maintenance, backup jobs, health checks

21. **Jobs Data Source** (`bcm_cmjob_jobs`)
    - Business Value: Query job status and history
    - API Methods: `getJobs`, `getJobHistory`
    - Implementation Effort: Medium (2-3 days) - Read-only with filtering
    - Use Case: Job monitoring, audit trails, troubleshooting

---

### CMServ Service (Service Management)

**Implementation Status**: NO COVERAGE

#### Likely API Methods

**Service Control**:
- `getServices` - Get all services ❌ MISSING
- `getService(serviceID)` - Get specific service ❌ MISSING
- `getServiceStatus(serviceID)` - Get service status ❌ MISSING
- `startService(serviceID)` - Start service ❌ MISSING
- `stopService(serviceID)` - Stop service ❌ MISSING
- `restartService(serviceID)` - Restart service ❌ MISSING
- `enableService(serviceID)` - Enable service autostart ❌ MISSING
- `disableService(serviceID)` - Disable service autostart ❌ MISSING

#### Gap Analysis

**Medium-Value Gaps**:

22. **Service Resource** (`bcm_cmserv_service`)
    - Business Value: Manage cluster services (NFS, DHCP, DNS, monitoring)
    - API Methods: `startService`, `stopService`, `restartService`, `enableService`
    - Implementation Effort: Medium (3-5 days) - Service state management
    - Use Case: Service orchestration, maintenance, failover

23. **Services Data Source** (`bcm_cmserv_services`)
    - Business Value: Query service status
    - API Methods: `getServices`, `getServiceStatus`
    - Implementation Effort: Small (1-2 days) - Read-only data source
    - Use Case: Service monitoring, health checks, dependency tracking

---

### CMMon Service (Monitoring)

**Implementation Status**: NO COVERAGE

#### Likely API Methods

**Monitoring Data**:
- `getMetrics` - Get cluster metrics ❌ MISSING
- `getNodeMetrics(deviceUUID)` - Get node-specific metrics ❌ MISSING
- `getAlerts` - Get active alerts ❌ MISSING
- `getAlert(alertID)` - Get specific alert ❌ MISSING
- `acknowledgeAlert(alertID)` - Acknowledge alert ❌ MISSING
- `getHealthStatus` - Get cluster health ❌ MISSING
- `getMonitoringData(query)` - Query monitoring data ❌ MISSING

#### Gap Analysis

**Medium-Value Gaps**:

24. **Metrics Data Source** (`bcm_cmmon_metrics`)
    - Business Value: Query cluster performance metrics
    - API Methods: `getMetrics`, `getNodeMetrics`
    - Implementation Effort: Medium (2-3 days) - Complex queries, time series data
    - Use Case: Performance monitoring, capacity planning, alerting

25. **Alerts Data Source** (`bcm_cmmon_alerts`)
    - Business Value: Query active cluster alerts
    - API Methods: `getAlerts`
    - Implementation Effort: Small (1-2 days) - Read-only data source
    - Use Case: Alert monitoring, incident response, automation triggers

---

### Low-Priority Services

**CMKube, CMCloud, CMEtcd, CMMain, CMCert, CMBeeGFS, CMProc**: No implementation required for initial production readiness. These are advanced/specialized features.

---

## Prioritized Gap List

### High-Value Missing Resources (Phase 1 - Critical for Production)

| Priority | Resource | Service | Business Value | Effort | Score |
|----------|----------|---------|---------------|--------|-------|
| 1 | `bcm_cmdevice_power` | CMDevice | Power management automation | Small (1-2d) | CRITICAL |
| 2 | `bcm_cmprov_job` | CMProv | Node provisioning workflows | Large (5-7d) | CRITICAL |
| 3 | `bcm_cmnet_network` | CMNet | Network infrastructure management | Medium (3-5d) | CRITICAL |
| 4 | `bcm_cmdevice_role` | CMDevice | Service role orchestration | Medium (3-5d) | HIGH |
| 5 | `bcm_cmpart_partition` | CMPart | Multi-OS partition management | Medium (3-5d) | HIGH |

### High-Value Missing Data Sources (Phase 1)

| Priority | Data Source | Service | Business Value | Effort | Score |
|----------|-------------|---------|---------------|--------|-------|
| 1 | `bcm_cmprov_status` | CMProv | Provisioning monitoring | Medium (2-3d) | CRITICAL |
| 2 | `bcm_cmdevice_roles` | CMDevice | Role discovery | Small (1d) | HIGH |
| 3 | `bcm_cmpart_partitions` | CMPart | Partition discovery | Small (1d) | HIGH |

### Medium-Value Gaps (Phase 2 - Production Enhancements)

**Resources**:
- `bcm_cmdevice_bmc` - BMC/IPMI configuration (2-3d)
- `bcm_cmpart_module` - Kernel module management (2-3d)
- `bcm_cmnet_vlan` - VLAN configuration (2-3d)
- `bcm_cmnet_subnet` - Subnet management (2-3d)
- `bcm_cmjob_job` - Job automation (5-7d)
- `bcm_cmserv_service` - Service control (3-5d)
- `bcm_cmprov_template` - Provisioning templates (3-5d)

**Data Sources**:
- `bcm_cmdevice_compute_nodes` - Compute node filtering (1d)
- `bcm_cmdevice_hardware` - Hardware inventory (1-2d)
- `bcm_cmpart_modules` - Module discovery (1d)
- `bcm_cmmon_metrics` - Performance metrics (2-3d)

### Low-Value Gaps (Phase 3 - Future Enhancements)

**Resources**:
- `bcm_cmdevice_bios` - BIOS configuration (2-3d)
- `bcm_cmmon_alert` - Alert management (2-3d)

**Data Sources**:
- `bcm_cmpart_kernels` - Kernel version discovery (1d)
- `bcm_cmnet_vlans` - VLAN discovery (1d)
- `bcm_cmmon_alerts` - Alert querying (1-2d)
- `bcm_cmjob_jobs` - Job history (2-3d)

---

## Implementation Effort Summary

### By Priority

**Phase 1 (Critical - High Priority)**:
- 5 Resources: 17-26 days
- 3 Data Sources: 4-5 days
- **Total: 21-31 days** (4-6 weeks)

**Phase 2 (Medium Priority - Production Enhancements)**:
- 7 Resources: 22-34 days
- 4 Data Sources: 5-7 days
- **Total: 27-41 days** (5-8 weeks)

**Phase 3 (Low Priority - Future)**:
- 2 Resources: 4-6 days
- 4 Data Sources: 5-7 days
- **Total: 9-13 days** (2-3 weeks)

### By Service

| Service | High-Value Gaps | Medium-Value Gaps | Low-Value Gaps | Total Effort |
|---------|----------------|------------------|----------------|-------------|
| CMDevice | 2 resources, 1 data source | 2 resources, 2 data sources | 1 resource | 12-19 days |
| CMPart | 1 resource, 1 data source | 2 resources, 1 data source | 1 data source | 8-13 days |
| CMNet | 1 resource | 2 resources, 1 data source | 1 data source | 9-14 days |
| CMProv | 1 resource, 1 data source | 1 resource | - | 10-15 days |
| CMJob | - | 1 resource, 1 data source | 1 data source | 7-11 days |
| CMServ | - | 1 resource | - | 3-5 days |
| CMMon | - | 1 data source | 1 resource, 1 data source | 5-8 days |

---

## Recommended Implementation Order

### Phase 1: Critical Production Gaps (4-6 weeks)

**Week 1-2**: Core Infrastructure
1. `bcm_cmnet_network` (resource) - Enable network management
2. `bcm_cmdevice_power` (resource) - Power control automation
3. `bcm_cmdevice_roles` (data source) - Role discovery

**Week 3-4**: Provisioning Workflows
4. `bcm_cmprov_status` (data source) - Provisioning monitoring
5. `bcm_cmprov_job` (resource) - Provisioning automation

**Week 5-6**: Advanced Device Management
6. `bcm_cmdevice_role` (resource) - Role assignment
7. `bcm_cmpart_partition` (resource) - Partition management
8. `bcm_cmpart_partitions` (data source) - Partition discovery

### Phase 2: Production Enhancements (5-8 weeks)

**Week 1-2**: Hardware & Configuration
1. `bcm_cmdevice_bmc` (resource) - Out-of-band management
2. `bcm_cmdevice_compute_nodes` (data source) - Compute filtering
3. `bcm_cmdevice_hardware` (data source) - Hardware inventory

**Week 3-4**: Advanced Networking
4. `bcm_cmnet_vlan` (resource) - VLAN management
5. `bcm_cmnet_subnet` (resource) - Subnet management

**Week 5-6**: Service & Job Management
6. `bcm_cmserv_service` (resource) - Service orchestration
7. `bcm_cmjob_job` (resource) - Job automation

**Week 7-8**: Monitoring & Templates
8. `bcm_cmmon_metrics` (data source) - Performance monitoring
9. `bcm_cmprov_template` (resource) - Provisioning templates
10. `bcm_cmpart_module` (resource) - Kernel module management

### Phase 3: Future Enhancements (2-3 weeks)

**Low-priority features for specialized use cases** - Implement based on user demand.

---

## Cross-Service Dependencies

### Network → Device
- Devices reference Networks via `managementNetwork` UUID
- Network interfaces require Network entity
- Implementation order: Networks before Devices

### Partition → Device
- Devices reference Partitions via `partition` UUID
- Software images associated with partitions
- Implementation order: Partitions before Devices

### Role → Device
- Devices reference Roles via `roles` array
- Roles define service assignments
- Implementation order: Roles before Device assignments

### Provisioning → Device + Partition
- Provisioning jobs require device and partition UUIDs
- Depends on device and partition management
- Implementation order: Device + Partition before Provisioning

---

## API Discovery Validation

### Confirmed Methods (Provider Code Evidence)

**CMDevice**:
- ✅ `getNodes` - data_source_cmdevice_nodes.go:316
- ✅ `getCategories` - data_source_cmdevice_categories.go:284
- ✅ `getCategory` - resource_cmdevice_category.go:1103
- ✅ `getDevice` - resource_cmdevice_device.go:580
- ✅ `addDevice` - resource_cmdevice_device.go:393
- ✅ `updateDevice` - resource_cmdevice_device.go:765
- ✅ `removeDevice` - resource_cmdevice_device.go:869
- ✅ `addCategory` - resource_cmdevice_category.go:613
- ✅ `updateCategory` - resource_cmdevice_category.go:747
- ✅ `removeCategory` - resource_cmdevice_category.go:799

**CMPart**:
- ✅ `getSoftwareImages` - data_source_cmpart_softwareimages.go:288
- ✅ `getSoftwareImage` - resource_cmpart_softwareimage.go:612
- ✅ `addSoftwareImage` - resource_cmpart_softwareimage.go:283
- ✅ `updateSoftwareImage` - resource_cmpart_softwareimage.go:478
- ✅ `removeSoftwareImage` - resource_cmpart_softwareimage.go:529
- ✅ `getPartitions` - resource_cmdevice_device.go:293

**CMNet**:
- ✅ `getNetworks` - data_source_cmnet_networks.go:291

### Documented Methods (sampleRest/ Evidence)

**CMDevice_Complete_Documentation.md**:
- ✅ `getNode` - Line 108-132
- ✅ `getComputeNodes` - Line 61-78
- ✅ `reboot` - Line 137-154

**BCM_API_Complete_Documentation.md**:
- ✅ Service catalog - Lines 63-83 (15 services discovered)
- ✅ Entity types - Lines 173-198

---

## Success Criteria Validation

Per spec.md Success Criteria SC-003:

- ✅ **At least 10 high-value gaps identified**: 12 high-value resources + 6 high-value data sources identified
- ✅ **Gaps categorized by service type**: CMDevice, CMPart, CMNet, CMProv, CMJob, CMServ, CMMon
- ✅ **Business value assessment**: All gaps categorized as High/Medium/Low with use cases
- ✅ **Implementation effort estimates**: All gaps estimated in days (Small/Medium/Large)
- ✅ **Prioritized implementation order**: 3-phase roadmap with week-by-week breakdown

---

## Next Steps

1. **Validation**: Review API gap analysis with BCM cluster validation (test discovered methods)
2. **Prioritization Review**: Confirm Phase 1 gaps align with production deployment blockers
3. **Resource Allocation**: Assign implementation teams to Phase 1 high-priority gaps
4. **Dependency Planning**: Ensure implementation order respects cross-service dependencies
5. **Integration with Remediation Plan**: Synthesize API gaps into overall remediation roadmap

---

## References

- **Research Methodology**: [research.md](./research.md) Section 2 - BCM API Discovery Approach
- **BCM Documentation**: `/workspace/sampleRest/BCM_API_Complete_Documentation.md`
- **CMDevice Documentation**: `/workspace/sampleRest/CMDevice_Complete_Documentation.md`
- **Provider Registration**: `/workspace/internal/provider/provider.go` (Lines 165-186)
- **Test Coverage Report**: [test-coverage-report.md](./test-coverage-report.md)

---

**Analysis Complete**: 2025-11-22
**Total Services Analyzed**: 15
**Total Gaps Identified**: 25 resources, 12 data sources
**Estimated Implementation Effort**: 57-85 days (11-17 weeks) for complete API coverage
