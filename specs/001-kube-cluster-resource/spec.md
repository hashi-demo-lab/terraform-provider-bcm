# Feature Specification: BCM Kubernetes Cluster Resource

**Feature Branch**: `001-kube-cluster-resource`
**Created**: 2025-11-22
**Status**: Draft
**Input**: User description: "Create a comprehensive feature specification for implementing a Terraform resource bcm_cmkube_cluster that manages Kubernetes clusters in BCM (Bright Cluster Manager)"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Basic Cluster Lifecycle Management (Priority: P1)

Infrastructure engineers need to create, update, and destroy Kubernetes clusters in BCM through declarative Terraform configurations, allowing them to version control cluster definitions and automate cluster provisioning.

**Why this priority**: This is the core functionality - without basic CRUD operations, the resource provides no value. This enables infrastructure-as-code workflows for Kubernetes cluster management.

**Independent Test**: Can be fully tested by creating a minimal cluster with required fields only (name, master nodes), verifying creation via BCM API, updating cluster configuration, and destroying it. Delivers immediate value by enabling automated cluster provisioning.

**Acceptance Scenarios**:

1. **Given** no existing Kubernetes cluster, **When** user applies Terraform config with cluster name and master nodes, **Then** BCM creates the cluster and returns UUID
2. **Given** an existing cluster managed by Terraform, **When** user modifies cluster configuration in Terraform, **Then** BCM updates the cluster with new configuration
3. **Given** an existing cluster managed by Terraform, **When** user runs terraform destroy, **Then** BCM removes the cluster and all associated resources
4. **Given** an existing cluster not managed by Terraform, **When** user imports the cluster by UUID, **Then** Terraform adopts the cluster into state with full configuration

---

### User Story 2 - Drift Detection and State Reconciliation (Priority: P1)

When infrastructure engineers make manual changes to Kubernetes clusters via BCM UI or API, Terraform must detect the configuration drift and offer to restore the desired state, ensuring infrastructure matches the declared configuration.

**Why this priority**: Drift detection is critical for maintaining infrastructure consistency and is a fundamental expectation of Terraform resources. Without it, manual changes can cause unexpected behavior.

**Independent Test**: Can be fully tested by creating a cluster via Terraform, modifying cluster attributes directly via BCM API (e.g., changing master count, worker nodes, network settings), running terraform plan to verify drift detected, and applying to restore desired state.

**Acceptance Scenarios**:

1. **Given** a cluster managed by Terraform, **When** cluster configuration is changed directly in BCM, **Then** terraform plan shows the drift and proposes corrective changes
2. **Given** detected configuration drift, **When** user runs terraform apply, **Then** cluster configuration is restored to match Terraform config
3. **Given** a cluster with external changes to master nodes, **When** terraform refresh is run, **Then** state is updated to reflect actual BCM configuration

---

### User Story 3 - Network and Node Configuration (Priority: P2)

Platform engineers need to configure Kubernetes cluster networking (management network, overlay network) and node assignments (master nodes, worker nodes) through Terraform, enabling standardized cluster topology definitions.

**Why this priority**: Network and node configuration are essential for production clusters but can be handled with reasonable defaults in MVP. Engineers need control over cluster topology for different environments.

**Independent Test**: Can be fully tested by creating a cluster with custom network settings (specific management network UUID, overlay network type), assigning specific master and worker nodes, and verifying the cluster operates with the specified configuration.

**Acceptance Scenarios**:

1. **Given** cluster configuration with management network UUID, **When** cluster is created, **Then** cluster uses the specified network for management traffic
2. **Given** cluster configuration with master node UUIDs, **When** cluster is created, **Then** specified nodes are configured as Kubernetes masters
3. **Given** cluster configuration with worker node UUIDs, **When** cluster is created, **Then** specified nodes join cluster as workers
4. **Given** an existing cluster, **When** worker nodes are added/removed in Terraform config, **Then** cluster scales to match desired node count

---

### User Story 4 - Kubernetes Version Management (Priority: P2)

Infrastructure teams need to specify and update Kubernetes versions for clusters, enabling controlled upgrades and version consistency across environments.

**Why this priority**: Version management is important for production workloads but clusters can start with default versions. This becomes critical for upgrades and compliance.

**Independent Test**: Can be fully tested by creating a cluster with a specific Kubernetes version, verifying the version via BCM API, updating the version in Terraform config, and validating the cluster upgrades successfully.

**Acceptance Scenarios**:

1. **Given** cluster configuration with specific k8s version, **When** cluster is created, **Then** cluster runs the specified Kubernetes version
2. **Given** an existing cluster with older k8s version, **When** version is updated in Terraform config, **Then** cluster upgrades to new version
3. **Given** invalid k8s version in config, **When** terraform plan is run, **Then** validation error is returned before API call

---

### User Story 5 - Advanced Cluster Configuration (Priority: P3)

DevOps teams need to configure advanced cluster settings like DNS, load balancer mode, storage classes, and add-ons through Terraform for specialized cluster requirements.

**Why this priority**: Advanced features are needed for specific use cases but not required for basic cluster operation. Teams can start with defaults and customize later.

**Independent Test**: Can be fully tested by creating a cluster with custom DNS servers, specific load balancer configuration, storage class definitions, and optional add-ons, then verifying each setting is applied correctly via BCM API.

**Acceptance Scenarios**:

1. **Given** cluster config with custom DNS servers, **When** cluster is created, **Then** cluster DNS resolves using specified servers
2. **Given** cluster config with load balancer mode, **When** cluster is created, **Then** services use specified load balancing strategy
3. **Given** cluster config with storage classes, **When** cluster is created, **Then** persistent volumes use defined storage classes
4. **Given** cluster config with optional add-ons (monitoring, logging), **When** cluster is created, **Then** add-ons are installed and operational

---

### User Story 6 - Force Operations and Error Recovery (Priority: P3)

Operations teams need the ability to force cluster operations (creation, updates, deletion) even when validation warnings exist, enabling recovery from edge cases and emergency operations.

**Why this priority**: Force operations are safety-critical edge cases needed only when normal operations fail. Most workflows should never use force.

**Independent Test**: Can be fully tested by attempting operations that would normally fail validation (e.g., deleting cluster with active workloads), using force parameter to override, and verifying operation completes successfully.

**Acceptance Scenarios**:

1. **Given** cluster deletion would fail due to active resources, **When** force parameter is set to true, **Then** cluster is deleted despite warnings
2. **Given** cluster update has validation warnings, **When** force parameter is enabled, **Then** update proceeds bypassing validation
3. **Given** force parameter is used, **When** operation completes, **Then** warning is logged about force usage

---

### Edge Cases

- **What happens when cluster creation times out?** BCM cluster provisioning may take several minutes; implementation must poll for completion with exponential backoff (similar to software image cloning pattern) or handle eventual consistency gracefully
- **How does system handle concurrent modifications?** BCM uses revision IDs for optimistic locking; conflicting updates should return clear error messages directing users to refresh state
- **What happens when referenced resources (nodes, networks) don't exist?** Validation should occur during plan phase using BCM API calls; missing UUIDs should fail with actionable error messages listing valid options
- **How are partial failures handled during cluster creation?** If cluster creation starts but fails midway (e.g., master nodes provision but workers fail), resource should track cluster UUID and allow retry/recovery rather than orphaning resources
- **What happens when cluster is deleted outside Terraform?** Next terraform refresh/plan should detect missing cluster and update state accordingly; terraform apply should recreate if still in config
- **How does import handle clusters with incomplete metadata?** Import should fetch full cluster configuration from BCM and populate all computed fields; missing optional fields should be set to null rather than failing import

## Requirements *(mandatory)*

### Functional Requirements

#### Core CRUD Operations

- **FR-001**: Resource MUST support creating Kubernetes clusters via BCM `cmkube.addKubeCluster` API call with cluster entity and optional force parameter
- **FR-002**: Resource MUST support reading cluster configuration via BCM `cmkube.getKubeCluster` API call using cluster UUID for efficient direct lookup
- **FR-003**: Resource MUST support updating cluster configuration via BCM `cmkube.updateKubeCluster` API call with modified cluster entity and optional force parameter
- **FR-004**: Resource MUST support deleting clusters via BCM `cmkube.removeKubeCluster` API call using cluster UUID and optional force parameter
- **FR-005**: Resource MUST implement ImportState using resource.ImportStatePassthroughID pattern to adopt existing clusters by UUID

#### Schema and Validation

- **FR-006**: Resource MUST define Terraform schema with all cluster attributes following BCM KubeCluster entity structure (baseType, childType, uuid, name, master nodes, worker nodes, networks, etc.)
- **FR-007**: Resource MUST mark cluster name as required and validate it matches BCM naming conventions (alphanumeric, hyphens, underscores only)
- **FR-008**: Resource MUST mark UUID and ID as computed attributes, with ID equal to UUID for consistency with other BCM resources
- **FR-009**: Resource MUST provide optional force parameter (boolean, default: false) to bypass validation warnings during create/update/delete operations
- **FR-010**: Resource MUST validate Kubernetes version format if specified (semver pattern like "1.28.0")
- **FR-011**: Resource MUST validate node UUIDs reference valid BCM nodes via API calls during plan phase when possible
- **FR-012**: Resource MUST validate network UUIDs reference valid BCM networks via API calls during plan phase when possible

#### State Management

- **FR-013**: Resource MUST preserve BCM entity structure in API calls (baseType="KubeCluster", childType="", modified=true, to_be_removed=false, revision="")
- **FR-014**: Resource MUST handle computed fields that BCM populates automatically (uuid, creation_time, revision_id)
- **FR-015**: Resource MUST properly map Terraform snake_case attribute names to BCM API camelCase field names (e.g., master_nodes → masterNodes)
- **FR-016**: Resource MUST handle null values for optional fields using types.String/Bool/Int64/List with null semantics
- **FR-017**: Resource MUST NOT propagate Unknown values to state; all fields must resolve to known values (actual value or explicit null) before setting state

#### Error Handling

- **FR-018**: Resource MUST use multi-layer error detection pattern from bcm_client.go (HTTP errors, JSON parsing, BCM success field, BCM error field)
- **FR-019**: Resource MUST provide clear error messages when cluster operations fail, including BCM API error details
- **FR-020**: Resource MUST handle eventual consistency for long-running operations (cluster creation/deletion) with polling and timeout mechanisms
- **FR-021**: Resource MUST return validation errors during plan phase before making API calls when possible (e.g., invalid UUIDs, malformed names)

#### Drift Detection

- **FR-022**: Resource Read operation MUST fetch current cluster state from BCM and update Terraform state to reflect actual configuration
- **FR-023**: Resource MUST detect when cluster configuration has changed outside Terraform and mark affected attributes for update
- **FR-024**: Resource MUST compare cluster attributes between plan and BCM state to identify drift for all managed fields

### Key Entities *(mandatory)*

- **KubeCluster**: Represents a Kubernetes cluster in BCM with attributes including:
  - Identity: uuid (BCM-assigned), name (user-provided)
  - Nodes: masterNodes (list of node UUIDs), workerNodes (list of node UUIDs)
  - Networks: managementNetwork (UUID), overlayNetwork (UUID or configuration)
  - Kubernetes: version, addons, configuration options
  - Metadata: baseType (always "KubeCluster"), childType (empty), revision, modified, to_be_removed
  - Operations: force flag for bypassing validation

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Infrastructure engineers can create a basic Kubernetes cluster with name and master nodes in under 10 minutes (including BCM provisioning time)
- **SC-002**: Resource detects 100% of configuration drift for all managed cluster attributes when changes occur outside Terraform
- **SC-003**: All acceptance tests pass using modern terraform-plugin-testing patterns (statecheck, knownvalue, plancheck)
- **SC-004**: Resource import successfully adopts existing clusters with complete state population (no Unknown or null values for fields BCM provides)
- **SC-005**: Cluster create/update/delete operations complete successfully with clear progress indication and error messages on failure
- **SC-006**: Documentation is auto-generated via `make generate` and includes working examples for common cluster configurations
- **SC-007**: Resource supports minimum 10 concurrent cluster operations without state corruption or race conditions
- **SC-008**: Test suite executes in under 15 minutes for full acceptance test run including create/import/update/delete/drift scenarios

### Quality Metrics

- **SC-009**: Zero hardcoded cluster names, UUIDs, or node references in test suite (all tests portable across BCM environments)
- **SC-010**: 100% of CRUD operations follow existing resource patterns (bcm_cmpart_softwareimage, bcm_cmdevice_category)
- **SC-011**: All optional fields have sensible defaults that allow minimal cluster configurations (name + master nodes only)
- **SC-012**: State updates occur atomically - partial failures leave Terraform state in recoverable condition

## Assumptions *(mandatory)*

### API Behavior Assumptions

- **A-001**: BCM `cmkube.addKubeCluster` returns cluster UUID immediately even if provisioning is asynchronous (similar to software image cloning)
- **A-002**: BCM `cmkube.getKubeCluster` accepts UUID parameter using args pattern for efficient direct lookup (not list+filter)
- **A-003**: BCM `cmkube.updateKubeCluster` accepts full cluster entity with baseType/childType/modified fields like other BCM resources
- **A-004**: BCM `cmkube.removeKubeCluster` accepts UUID and force parameter, returning success when cluster deletion begins
- **A-005**: BCM `cmkube.validateKubeCluster` provides pre-flight validation and can be called during plan phase for early error detection

### State Management Assumptions

- **A-006**: Cluster UUID is stable and unique across cluster lifecycle (create → update → delete)
- **A-007**: BCM preserves cluster UUID during updates, allowing Terraform to track cluster identity reliably
- **A-008**: BCM returns full cluster configuration in getKubeCluster response, including all computed and optional fields
- **A-009**: Empty/null optional fields are returned as null (not omitted) in BCM API responses for consistent state handling

### Operational Assumptions

- **A-010**: Cluster creation completes within 30 minutes in normal conditions (used for polling timeout)
- **A-011**: Cluster deletion completes within 15 minutes in normal conditions (used for polling timeout)
- **A-012**: Force parameter bypasses validation warnings but not critical errors (e.g., invalid UUIDs still fail)
- **A-013**: Test environment has minimum 2 available nodes for creating test clusters (1 master + 1 worker minimum)
- **A-014**: Test environment has BCM network configuration that supports Kubernetes cluster networking

### Implementation Assumptions

- **A-015**: Resource follows TDD workflow: write failing acceptance tests first, then implement minimal CRUD to pass
- **A-016**: Documentation examples use environment variables for credentials (BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD)
- **A-017**: Test helpers from test_helpers.go (createTestBCMClient, generateUniqueTestName, verifyResourceDeleted) are reusable for cluster resource
- **A-018**: Cluster resource uses same BCM entity structure pattern as existing resources (no special-case handling needed)

## Out of Scope *(mandatory)*

The following items are explicitly excluded from this feature to maintain focused scope:

### Advanced Kubernetes Features

- **OS-001**: Kubernetes cluster monitoring dashboards (Grafana/Prometheus integration) - users can configure monitoring via Kubernetes addons after cluster creation
- **OS-002**: Automated cluster backup and restore functionality - handled by separate BCM backup mechanisms
- **OS-003**: Multi-cluster federation or cluster mesh configurations - requires separate resource type
- **OS-004**: Custom CNI (Container Network Interface) plugin installation - assumes BCM default CNI or manual post-provisioning configuration

### Complex Validation

- **OS-005**: Pre-flight validation of node capacity/resources (CPU, memory, disk) - BCM API handles capacity validation
- **OS-006**: Kubernetes version compatibility matrix validation - assumes BCM enforces supported version combinations
- **OS-007**: Node affinity/anti-affinity rules for master/worker placement - uses BCM default scheduling
- **OS-008**: Network policy validation or RBAC (Role-Based Access Control) pre-configuration - handled via Kubernetes API post-creation

### Operational Automation

- **OS-009**: Automated cluster upgrades with rolling updates - users trigger version changes via Terraform config update
- **OS-010**: Automatic node replacement on failure - handled by BCM cluster management, not Terraform resource
- **OS-011**: Load balancer provisioning external to BCM - assumes BCM-provided load balancing or manual external LB setup
- **OS-012**: Certificate rotation automation - assumes BCM handles internal certificate lifecycle

### Data Source Companion

- **OS-013**: Data source `bcm_cmkube_cluster` for reading cluster information - can be added as separate feature if needed
- **OS-014**: Data source `bcm_cmkube_clusters` for listing all clusters - separate feature scope

### Edge Case Handling

- **OS-015**: Cluster migration between BCM servers - assumes single BCM endpoint for resource lifecycle
- **OS-016**: Cluster state recovery from catastrophic BCM failure - out of Terraform scope, handled by BCM disaster recovery
- **OS-017**: Handling of clusters in ERROR state - assumes BCM API provides error details, Terraform surfaces to user

## Dependencies *(mandatory)*

### External Dependencies

- **D-001**: BCM API endpoint must be accessible and authenticated (bcm_client.go handles authentication)
- **D-002**: BCM API must provide cmkube service with addKubeCluster, getKubeCluster, updateKubeCluster, removeKubeCluster, validateKubeCluster methods
- **D-003**: BCM cluster must have available nodes for assignment to Kubernetes cluster (minimum 1 for master, additional for workers)
- **D-004**: BCM network configuration must support Kubernetes cluster networking requirements

### Internal Dependencies

- **D-005**: BCM client implementation in internal/provider/bcm_client.go must support CallJSONRPC with args parameter for cmkube service
- **D-006**: Test helpers in internal/provider/test_helpers.go must be available for test client creation and resource verification
- **D-007**: Provider configuration must include endpoint, username, password, and insecure_skip_verify for BCM authentication

### Terraform Framework Dependencies

- **D-008**: terraform-plugin-framework v1.16.1+ for resource schema and CRUD operations
- **D-009**: terraform-plugin-testing v1.13.3+ for modern acceptance test patterns (statecheck, knownvalue, plancheck, tfjsonpath)
- **D-010**: terraform-plugin-log for structured logging during resource operations

### Build Dependencies

- **D-011**: Go 1.24+ for resource implementation
- **D-012**: make generate (tfplugindocs) for documentation generation
- **D-013**: pre-commit hooks for code quality enforcement

## Research Questions for Phase 0 *(mandatory)*

The following questions MUST be answered during Phase 0 implementation planning by exploring BCM API:

### API Contract Verification

- **RQ-001**: What is the exact structure of the KubeCluster entity expected by addKubeCluster/updateKubeCluster? (Confirm baseType, required fields, optional fields, nested objects)
- **RQ-002**: Does getKubeCluster accept UUID as args parameter for direct lookup, or does it require list+filter pattern?
- **RQ-003**: What does the response from getKubeCluster look like? (Identify all returned fields, data types, null vs omitted handling)
- **RQ-004**: What validation does validateKubeCluster perform? Is it worth calling during plan phase or is it redundant with add/update validation?
- **RQ-005**: What does removeKubeCluster return on success/failure? Does it wait for deletion completion or return immediately?

### Field Mappings and Data Types

- **RQ-006**: What are the exact field names in BCM API for master nodes and worker nodes? (masterNodes vs master_nodes vs masters?)
- **RQ-007**: What format does BCM expect for node references? (UUID strings, array of UUIDs, nested objects with UUIDs?)
- **RQ-008**: What are valid values for Kubernetes version field? (Semver strings, version codes, enum values?)
- **RQ-009**: What network configuration fields exist in KubeCluster entity? (Management network UUID, overlay network type, CNI configuration?)
- **RQ-010**: What optional fields exist for advanced cluster configuration? (DNS, storage, addons, load balancer settings?)

### Operational Behavior

- **RQ-011**: Is cluster creation synchronous or asynchronous? If async, how do we poll for completion status?
- **RQ-012**: How long does typical cluster creation take in test environment? (Needed to set realistic polling timeout)
- **RQ-013**: Can clusters be updated while in PROVISIONING state, or must updates wait for ready state?
- **RQ-014**: What happens if force parameter is set to true vs false during operations? (Document actual behavior vs assumptions)
- **RQ-015**: Does BCM support partial updates (PATCH semantics) or require full entity replacement (PUT semantics)?

### Error Handling

- **RQ-016**: What error codes/messages does BCM return for common failure scenarios? (Invalid UUID, missing nodes, network conflicts, etc.)
- **RQ-017**: How does BCM indicate eventual consistency vs hard failures during long-running operations?
- **RQ-018**: Can we safely retry failed operations, or do partial failures leave orphaned resources requiring manual cleanup?

### Test Environment

- **RQ-019**: How many available nodes exist in test BCM cluster for creating test Kubernetes clusters?
- **RQ-020**: What network UUIDs are available in test environment for cluster network configuration?
- **RQ-021**: What Kubernetes versions are supported by test BCM installation?

## Phase 0 Exploration Strategy *(mandatory)*

To answer research questions and validate API contract:

1. **API Exploration Script**: Create `sampleRest/cmkube-get-clusters.py` to call `cmkube.getKubeClusters` (list) and inspect response structure
2. **Entity Structure Analysis**: If clusters exist, inspect full entity structure to understand all fields and data types
3. **CRUD Operation Testing**: Create `sampleRest/cmkube-crud-test.py` to test full lifecycle:
   - Call `addKubeCluster` with minimal entity and capture response
   - Call `getKubeCluster` with returned UUID to verify read
   - Call `updateKubeCluster` with modified entity
   - Call `validateKubeCluster` to understand validation behavior
   - Call `removeKubeCluster` to clean up test cluster
4. **Documentation Search**: Check if `sampleRest/BCM_API_Complete_Documentation.md` or similar files contain cmkube service documentation
5. **Field Mapping Reference**: Document exact JSON field names and types in spec for implementation reference
6. **Update Spec**: Add findings to "BCM API Contract" section in spec.md with actual JSON examples

## BCM API Contract *(to be populated in Phase 0)*

This section will be updated during Phase 0 with actual API exploration findings.

### Expected API Methods

```json
// Create cluster
{
  "service": "cmkube",
  "call": "addKubeCluster",
  "args": [
    {
      "baseType": "KubeCluster",
      "childType": "",
      "modified": true,
      "to_be_removed": false,
      "revision": "",
      "name": "test-cluster",
      "masterNodes": ["node-uuid-1"],
      "workerNodes": ["node-uuid-2", "node-uuid-3"],
      "managementNetwork": "network-uuid",
      "version": "1.28.0"
    },
    false  // force parameter
  ]
}

// Read cluster (verify args pattern)
{
  "service": "cmkube",
  "call": "getKubeCluster",
  "args": ["cluster-uuid"]
}

// Update cluster
{
  "service": "cmkube",
  "call": "updateKubeCluster",
  "args": [
    {
      "baseType": "KubeCluster",
      "uuid": "cluster-uuid",
      // ... full entity with modifications
    },
    false  // force parameter
  ]
}

// Delete cluster
{
  "service": "cmkube",
  "call": "removeKubeCluster",
  "args": ["cluster-uuid", false]  // uuid, force
}
```

**Note**: Actual field names, structure, and behavior will be confirmed via API exploration in Phase 0.
