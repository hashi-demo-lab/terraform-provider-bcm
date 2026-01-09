# Feature Specification: CMKube API Alignment

**Feature Branch**: `102-cmkube-api-alignment`
**Created**: 2026-01-09
**Status**: Draft
**Input**: Align the `bcm_cmkube_cluster` Terraform resource with the actual BCM KubeCluster API entity structure

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create Kubernetes Cluster with Correct BCM Entity Mapping (Priority: P1)

As a platform engineer, I want to create a Kubernetes cluster definition in BCM using Terraform so that the cluster configuration is accurately stored and retrievable from the BCM API.

**Why this priority**: This is the core functionality - without correct field mapping, clusters cannot be properly created or managed. Currently, fields like `master_nodes`, `worker_nodes`, and `etcd_nodes` are silently ignored by BCM, leading to configuration drift and management issues.

**Independent Test**: Can be fully tested by creating a `bcm_cmkube_cluster` resource and verifying via BCM API that all configured fields are persisted correctly.

**Acceptance Scenarios**:

1. **Given** a Terraform configuration with `bcm_cmkube_cluster` resource specifying `internal_network`, `service_network`, and `pod_network`, **When** the resource is applied, **Then** the BCM KubeCluster entity contains the correct network UUID references in `internalNetwork`, `serviceNetwork`, and `podNetwork` fields.

2. **Given** a Terraform configuration with `bcm_cmkube_cluster` specifying `etcd_cluster` reference, **When** the resource is applied, **Then** the BCM KubeCluster entity correctly references the EtcdCluster UUID.

3. **Given** an existing `bcm_cmkube_cluster` resource, **When** `terraform refresh` is executed, **Then** all persisted fields are accurately reflected in Terraform state without drift.

4. **Given** a Terraform configuration with `app_groups` nested blocks containing applications, **When** the resource is applied, **Then** BCM stores the applications in the correct `appGroups` structure with `KubeAppGroup` and `KubeApp` entities.

---

### User Story 2 - Manage Etcd Clusters Independently (Priority: P1)

As a platform engineer, I want to create and manage EtcdCluster entities separately from KubeClusters so that I can share etcd infrastructure across multiple Kubernetes clusters or manage etcd independently.

**Why this priority**: KubeCluster requires an EtcdCluster reference. Without the ability to manage EtcdCluster as a first-class resource, users cannot properly set up Kubernetes clusters.

**Independent Test**: Can be fully tested by creating a `bcm_cmetcd_cluster` resource and verifying the entity exists in BCM with correct configuration.

**Acceptance Scenarios**:

1. **Given** a Terraform configuration with `bcm_cmetcd_cluster` resource, **When** the resource is applied, **Then** an EtcdCluster entity is created in BCM with the specified configuration.

2. **Given** an existing `bcm_cmetcd_cluster` resource, **When** configuration parameters like `heartbeat_interval` are modified, **Then** the BCM entity is updated accordingly.

3. **Given** an existing `bcm_cmetcd_cluster` with no dependent KubeClusters, **When** `terraform destroy` is executed, **Then** the EtcdCluster entity is removed from BCM.

4. **Given** a `bcm_cmetcd_cluster` UUID, **When** importing the resource via `terraform import`, **Then** all etcd cluster configuration is accurately reflected in Terraform state.

---

### User Story 3 - Assign Kubernetes Roles to Devices (Priority: P1)

As a platform engineer, I want to assign KubeletRole and EtcdHostRole to devices so that nodes participate in the Kubernetes and etcd clusters with the correct configuration.

**Why this priority**: BCM does not have `master_nodes` or `worker_nodes` fields on KubeCluster. Nodes are assigned via roles embedded in Device entities. This is the actual mechanism for building cluster topology.

**Independent Test**: Can be fully tested by adding a role to the `bcm_cmdevice_device` resource and verifying the device has the role in BCM.

**Acceptance Scenarios**:

1. **Given** a device and a KubeCluster, **When** a `kubelet_role` block is added to the device resource referencing the cluster, **Then** the device is assigned a KubeletRole in BCM.

2. **Given** a device with a `kubelet_role`, **When** `control_plane = true` is specified, **Then** the device is configured as a Kubernetes control plane node.

3. **Given** a device with a `kubelet_role`, **When** `worker = true` is specified, **Then** the device is configured as a Kubernetes worker node.

4. **Given** a device and an EtcdCluster, **When** an `etcd_host_role` block is added to the device resource, **Then** the device is assigned an EtcdHostRole in BCM.

5. **Given** a device with both `kubelet_role` and `etcd_host_role`, **When** the resource is applied, **Then** both roles are present in the device's roles array in BCM.

---

### User Story 4 - Import Existing Kubernetes Infrastructure (Priority: P2)

As a platform engineer, I want to import existing BCM Kubernetes infrastructure into Terraform so that I can manage pre-existing clusters with Infrastructure as Code.

**Why this priority**: Organizations with existing BCM deployments need a migration path to Terraform management without recreating infrastructure.

**Independent Test**: Can be fully tested by importing an existing KubeCluster or EtcdCluster by UUID and verifying state accuracy.

**Acceptance Scenarios**:

1. **Given** an existing KubeCluster in BCM, **When** `terraform import bcm_cmkube_cluster.example <uuid>` is executed, **Then** all cluster configuration is accurately reflected in Terraform state.

2. **Given** an existing EtcdCluster in BCM, **When** `terraform import bcm_cmetcd_cluster.example <uuid>` is executed, **Then** all etcd configuration is accurately reflected in Terraform state.

3. **Given** an imported resource, **When** `terraform plan` is executed without configuration changes, **Then** no changes are detected (clean import).

---

### User Story 5 - Configure Kubernetes API Server Settings (Priority: P2)

As a platform engineer, I want to configure Kubernetes API server settings so that I can customize the cluster's API endpoint and proxy configuration.

**Why this priority**: API server configuration is essential for production deployments but can use reasonable defaults for basic clusters.

**Independent Test**: Can be fully tested by configuring API server settings and verifying they are persisted in BCM.

**Acceptance Scenarios**:

1. **Given** a `bcm_cmkube_cluster` with `kubernetes_api_server` specified, **When** the resource is applied, **Then** BCM stores the API server URL.

2. **Given** a `bcm_cmkube_cluster` with `kubernetes_api_server_proxy_port` specified, **When** the resource is applied, **Then** BCM stores the proxy port (default: 6444).

3. **Given** a `bcm_cmkube_cluster` with `trusted_domains` list specified, **When** the resource is applied, **Then** BCM stores the certificate SANs.

---

### User Story 6 - Configure Ingress Proxy Settings (Priority: P3)

As a platform engineer, I want to configure ingress proxy settings so that external traffic can be routed to cluster services.

**Why this priority**: Ingress configuration is an advanced feature that not all clusters require.

**Independent Test**: Can be fully tested by enabling ingress proxy and verifying settings in BCM.

**Acceptance Scenarios**:

1. **Given** a `bcm_cmkube_cluster` with `ingress_proxy_enable = true`, **When** the resource is applied, **Then** BCM enables the ingress proxy.

2. **Given** a `bcm_cmkube_cluster` with ingress proxy ports configured, **When** the resource is applied, **Then** BCM stores the listen and backend port settings.

---

### Edge Cases

- What happens when a KubeCluster references a non-existent EtcdCluster UUID? Provider should fail with a clear validation error during plan/apply.
- What happens when deleting an EtcdCluster that is referenced by a KubeCluster? Provider should detect the dependency and fail with an error explaining the constraint.
- How does the system handle a device with a KubeletRole referencing a deleted KubeCluster? Provider should detect orphaned roles during refresh and report the inconsistency.
- What happens when both `control_plane` and `worker` are set to false on a KubeletRole? Provider should allow this (valid in BCM for label-only nodes) but warn the user.
- What happens when updating a KubeCluster's network references while nodes are running? BCM may reject or require force - provider should surface the BCM validation error clearly.

## Requirements *(mandatory)*

### Functional Requirements

#### bcm_cmkube_cluster Resource

- **FR-001**: Resource MUST map `internal_network` attribute to BCM's `internalNetwork` field (UUID reference to Network entity)
- **FR-002**: Resource MUST map `service_network` attribute to BCM's `serviceNetwork` field (UUID reference to Network entity)
- **FR-003**: Resource MUST map `pod_network` attribute to BCM's `podNetwork` field (UUID reference to Network entity)
- **FR-004**: Resource MUST support `etcd_cluster` attribute to reference an EtcdCluster entity by UUID
- **FR-005**: Resource MUST support `pod_network_node_mask` attribute (e.g., "/24") for pod CIDR allocation
- **FR-006**: Resource MUST support `kube_dns_ip` attribute for cluster DNS IP address
- **FR-007**: Resource MUST support `kubernetes_api_server` attribute for API server URL
- **FR-008**: Resource MUST support `kubernetes_api_server_proxy_port` attribute with default value 6444
- **FR-009**: Resource MUST support `version` attribute for Kubernetes version (semver format)
- **FR-010**: Resource MUST support `trusted_domains` list attribute for certificate SANs
- **FR-011**: Resource MUST support `app_groups` nested block with proper `KubeAppGroup` structure containing:
  - `name` (required string)
  - `enabled` (optional boolean, default true)
  - `applications` nested block for `KubeApp` entities
- **FR-012**: Resource MUST support `label_sets` nested block with proper `KubeLabelSet` structure
- **FR-013**: Resource MUST support `users` nested block with proper `KubeUser` structure
- **FR-014**: Resource MUST support `ingress_proxy_enable`, `ingress_proxy_listen_port`, and `ingress_proxy_backend_port` attributes
- **FR-015**: Resource MUST support `options` attribute for extensible JSON configuration
- **FR-016**: Resource MUST remove non-functional fields: `master_nodes`, `worker_nodes`, `etcd_nodes`, `dns_servers`, `cni_plugin`, `storage_classes`, `load_balancer_mode`, `addons`, `overlay_network`, `ingress_controller`
- **FR-017**: Resource MUST support import via UUID using `terraform import`
- **FR-018**: Resource MUST call `validateKubeCluster` API before create/update operations
- **FR-019**: Resource MUST generate client-side UUID for new clusters (BCM cmkube API requirement)

#### bcm_cmetcd_cluster Resource (NEW)

- **FR-020**: Resource MUST manage BCM EtcdCluster entities with CRUD operations via CMEtcd service
- **FR-021**: Resource MUST support `name` attribute (required, unique identifier)
- **FR-022**: Resource MUST support `heartbeat_interval` attribute (default: 100ms)
- **FR-023**: Resource MUST support `election_timeout` attribute (default: 1000ms)
- **FR-024**: Resource MUST support `options` attribute for extensible JSON configuration
- **FR-025**: Resource MUST support import via UUID using `terraform import`

#### bcm_cmdevice_device Resource Updates

- **FR-026**: Resource MUST support `kubelet_role` nested block within `roles` containing:
  - `kube_cluster` (required, UUID reference to KubeCluster)
  - `control_plane` (optional boolean, default true)
  - `worker` (optional boolean, default true)
  - `container_runtime_service` (optional string, default "docker.service")
  - `max_pods` (optional integer, default 110)
  - `options` (optional JSON string for kubelet flags)
  - `custom_yaml` (optional string for /var/lib/kubelet/config.yaml)
- **FR-027**: Resource MUST support `etcd_host_role` nested block within `roles` containing:
  - `etcd_cluster` (required, UUID reference to EtcdCluster)
  - `member_name` (optional string, default "$hostname")
  - `spool` (optional string, default "/var/lib/etcd")
  - `listen_client_urls` (optional list, default ["https://0.0.0.0:2379"])
  - `listen_peer_urls` (optional list, default ["https://0.0.0.0:2380"])
  - `advertise_client_urls` (optional list, default ["https://$ip:2379"])
  - `advertise_peer_urls` (optional list, default ["https://$ip:2380"])
  - `snapshot_count` (optional integer, default 100000)
  - `max_snapshots` (optional integer, default 5)
- **FR-028**: Resource MUST embed roles in the Device entity when calling `updateDevice` API (not separate API calls)
- **FR-029**: Resource MUST preserve existing roles when updating device (BCM replaces entire roles array)

### Key Entities

- **KubeCluster**: Represents a Kubernetes cluster definition in BCM. Contains network references (internal, service, pod networks), API server configuration, application groups for cluster addons, and metadata. Does NOT contain node lists - nodes are assigned via roles.

- **EtcdCluster**: Represents an etcd cluster that provides distributed key-value storage for Kubernetes. Can be shared across multiple KubeClusters or dedicated to one.

- **KubeAppGroup**: A named group of Kubernetes applications (manifests) that can be enabled/disabled together. Used for cluster addons like monitoring, ingress controllers, etc.

- **KubeApp**: A single Kubernetes application within an AppGroup. Contains YAML/JSON manifest content and enabled state.

- **KubeLabelSet**: Defines node labels that can be applied to nodes, categories, or overlays.

- **KubeUser**: Represents a Kubernetes user for kubeconfig management.

- **KubeletRole**: A role assigned to devices that makes them Kubernetes cluster members. References a KubeCluster and specifies control plane/worker status.

- **EtcdHostRole**: A role assigned to devices that makes them etcd cluster members. References an EtcdCluster and configures etcd member settings.

- **Device**: Represents a physical or virtual node. Roles are embedded in the Device entity, not managed separately.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All fields configured in `bcm_cmkube_cluster` resource are persisted in BCM and accurately retrieved on subsequent reads (0% silent data loss)
- **SC-002**: Users can create a complete Kubernetes cluster topology (etcd + control plane + workers) using Terraform in under 10 minutes of configuration time
- **SC-003**: `terraform plan` shows no unexpected changes after `terraform apply` completes successfully (idempotent operations)
- **SC-004**: Existing KubeCluster and EtcdCluster entities can be imported into Terraform with 100% configuration fidelity
- **SC-005**: Validation errors from BCM API are surfaced as clear, actionable Terraform error messages
- **SC-006**: Resource deletion order is automatically validated - dependent resources cannot be deleted while references exist
- **SC-007**: Documentation examples demonstrate complete cluster setup patterns that work on first apply

## Assumptions

- BCM API behavior verified against actual BCM 9.2+ deployment
- The `options` JSON field in BCM entities can be used to store provider-specific metadata if needed
- Users accept that this is a breaking change from the previous resource schema
- Device role assignment via embedded roles in `updateDevice` is the only supported mechanism in BCM
- KubeletRole's `controlPlane` and `worker` flags are not mutually exclusive (a node can be both)

## Out of Scope

- Migration tooling for existing Terraform configurations (separate task)
- Data source for listing KubeletRoles or EtcdHostRoles (roles are accessed via Device)
- Support for BCM versions prior to 9.2
- Kubernetes workload management (this is cluster infrastructure only)
- Certificate management automation (BCM handles internally)
