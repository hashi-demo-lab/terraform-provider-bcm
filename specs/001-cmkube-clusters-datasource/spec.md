# Feature Specification: BCM Kubernetes Clusters Data Source

**Feature Branch**: `001-cmkube-clusters-datasource`
**Created**: 2025-11-23
**Status**: Draft
**Input**: User description: "Create a feature specification for implementing the data.bcm_cmkube_clusters data source to list and filter Kubernetes clusters managed by BCM."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Cluster Discovery for Import (Priority: P1)

Platform engineers need to discover existing Kubernetes clusters in BCM to import them into Terraform state for infrastructure management. This enables transitioning from manually-managed clusters to infrastructure-as-code without recreating clusters.

**Why this priority**: This is the foundational use case that enables Terraform adoption. Without the ability to discover and import existing clusters, users cannot begin managing BCM Kubernetes infrastructure with Terraform. This is a prerequisite for all other cluster management workflows.

**Independent Test**: Can be fully tested by querying the data source without filters and verifying all existing clusters are returned with their UUIDs and names, then using those UUIDs for terraform import commands.

**Acceptance Scenarios**:

1. **Given** BCM contains 3 Kubernetes clusters, **When** engineer queries `data.bcm_cmkube_clusters.all` without filters, **Then** data source returns all 3 clusters with uuid, name, master_nodes, worker_nodes, and version attributes populated
2. **Given** engineer has cluster UUID from data source, **When** they run `terraform import bcm_cmkube_cluster.prod <uuid>`, **Then** cluster is successfully imported into Terraform state
3. **Given** BCM contains no clusters, **When** engineer queries data source, **Then** data source returns empty clusters list without error

---

### User Story 2 - Filter Clusters by Name Pattern (Priority: P2)

DevOps engineers managing multiple environments need to filter clusters by naming convention (e.g., "prod-*", "dev-*", "staging-*") to organize cluster references and build environment-specific dependencies.

**Why this priority**: This enables organized multi-environment management and is commonly needed for production Terraform configurations. While not absolutely required for basic import, it significantly improves usability for teams with multiple clusters.

**Independent Test**: Can be tested independently by creating test clusters with specific name patterns, then verifying the filter returns only matching clusters while excluding others.

**Acceptance Scenarios**:

1. **Given** BCM has clusters "prod-us-east", "prod-eu-west", "dev-us-east", **When** engineer applies filter `name_pattern = "prod-*"`, **Then** data source returns only the 2 production clusters
2. **Given** BCM has cluster "k8s-monitoring-001", **When** engineer applies filter `name_pattern = "*monitoring*"`, **Then** data source returns the monitoring cluster
3. **Given** BCM has 10 clusters but none match "nonexistent-*", **When** engineer applies this filter, **Then** data source returns empty list without error

---

### User Story 3 - Filter Clusters by Version (Priority: P3)

Platform teams performing Kubernetes version upgrades need to identify clusters running specific Kubernetes versions to plan upgrade sequences and validate version consistency across environments.

**Why this priority**: This is valuable for operational workflows but not critical for basic Terraform adoption. Most teams can discover versions through the unfiltered data source, making this a convenience feature rather than a requirement.

**Independent Test**: Can be tested by creating clusters with different Kubernetes versions and verifying the filter returns only clusters matching the specified version.

**Acceptance Scenarios**:

1. **Given** BCM has clusters running v1.27.3, v1.28.1, v1.28.1, **When** engineer applies filter `version = "v1.28.1"`, **Then** data source returns the 2 clusters running v1.28.1
2. **Given** all clusters have been upgraded to v1.29.0, **When** engineer checks for legacy versions with `version = "v1.27.3"`, **Then** data source returns empty list confirming upgrade completion
3. **Given** version filter is not specified, **When** engineer queries clusters, **Then** all clusters are returned regardless of version

---

### User Story 4 - Filter Clusters by Master Node (Priority: P3)

Infrastructure operators troubleshooting node assignments or planning node replacements need to find which cluster(s) a specific node belongs to based on master node UUID.

**Why this priority**: This is a specialized troubleshooting scenario. While helpful for operations, it's not frequently needed and can be accomplished through other means (examining cluster details manually). It's a quality-of-life improvement rather than a core requirement.

**Independent Test**: Can be tested by creating a cluster with a known master node UUID, then verifying the filter returns that cluster when queried by the master node ID.

**Acceptance Scenarios**:

1. **Given** cluster "prod-main" has master node with UUID "node-abc-123", **When** engineer applies filter `master_node_id = "node-abc-123"`, **Then** data source returns only the "prod-main" cluster
2. **Given** a node UUID exists but is not a master node in any cluster, **When** engineer filters by this UUID, **Then** data source returns empty list
3. **Given** multiple clusters share a master node in HA configuration, **When** engineer filters by shared master node UUID, **Then** data source returns all clusters using that master node

---

### Edge Cases

- What happens when BCM API is unreachable or returns authentication errors? (System should return clear Terraform error message with API failure details)
- How does the data source handle clusters with missing or null attributes (e.g., no worker nodes, no version specified)? (Use null-safe field extraction to return types.String/types.List null values for missing fields)
- What happens when name_pattern filter contains invalid glob syntax? (Terraform validates pattern syntax and returns validation error before API call)
- How are multiple filters combined (name_pattern + version)? (Use AND logic - clusters must match all specified filters)
- What happens when BCM returns clusters with unexpected field types or structures? (Log warning and skip invalid clusters, continue processing valid ones)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Data source MUST retrieve all Kubernetes clusters from BCM via cmkube.getKubeClusters API
- **FR-002**: Data source MUST support optional name_pattern filter using glob pattern matching (case-sensitive, supports * wildcard)
- **FR-003**: Data source MUST support optional version filter for exact Kubernetes version match (semver string)
- **FR-004**: Data source MUST support optional master_node_id filter to find clusters containing specific master node UUID
- **FR-005**: Data source MUST apply multiple filters with AND logic (cluster must match all specified filters)
- **FR-006**: Data source MUST populate all cluster attributes including uuid, name, master_nodes, worker_nodes, etcd_nodes, management_network, version, cni_plugin, creation_time
- **FR-007**: Data source MUST use null-safe field extraction for optional API fields (overlayNetwork, dnsServers, storageClasses, etc.)
- **FR-008**: Data source MUST set ID to placeholder value (e.g., "cmkube-clusters") since this is a list data source
- **FR-009**: Data source MUST handle BCM API errors gracefully with clear error messages for authentication failures, network errors, and invalid API responses
- **FR-010**: Data source MUST return empty clusters list (not error) when no clusters match filters or BCM has no clusters
- **FR-011**: Data source MUST map BCM API camelCase field names to Terraform snake_case attribute names (masterNodes → master_nodes, workerNodes → worker_nodes, managementNetwork → management_network)
- **FR-012**: Data source MUST perform client-side filtering after retrieving all clusters from BCM (not API-side filtering)

### Key Entities

- **Cluster**: Represents a Kubernetes cluster managed by BCM with attributes:
  - Identity: uuid (unique identifier), name (user-friendly name)
  - Node Configuration: master_nodes (list of master node UUIDs), worker_nodes (list of worker node UUIDs), etcd_nodes (list of etcd node UUIDs)
  - Network Configuration: management_network (UUID), overlay_network (pod network config), dns_servers (list of IPs)
  - Kubernetes Configuration: version (semver string), cni_plugin (CNI plugin name)
  - Storage: storage_classes (JSON config)
  - Load Balancer: load_balancer_mode (strategy)
  - Addons: addons (JSON config), ingress_controller (JSON config)
  - Metadata: creation_time (Unix timestamp), revision_id (version number), base_type, child_type

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Engineers can discover all BCM Kubernetes clusters by querying the data source without filters, retrieving uuid, name, and node configuration for each cluster in under 5 seconds
- **SC-002**: Engineers can filter clusters using name_pattern glob syntax with 100% accuracy (only matching clusters returned)
- **SC-003**: Data source handles environments with 0-100+ clusters without performance degradation or timeout errors
- **SC-004**: Data source successfully processes clusters with missing optional attributes (null worker_nodes, null version) without errors, returning appropriate null values
- **SC-005**: Engineers can use discovered cluster UUIDs for terraform import commands with 100% success rate
- **SC-006**: All 12 functional requirements are verified by passing acceptance tests covering list all, filter by name, filter by version, filter by master node, and edge cases
- **SC-007**: Data source documentation auto-generates via `make generate` with complete attribute descriptions and filter usage examples
