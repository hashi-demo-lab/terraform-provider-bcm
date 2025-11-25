# Feature Specification: bcm_cmkube_cluster - Add etcd_nodes and Validate Attributes

**Feature Branch**: `025-cmkube-etcd-nodes`
**Created**: 2025-11-25
**Status**: Draft
**Issue**: [#25 - P0 - resource.bcm_cmkube_cluster - Add etcd_nodes and validate attributes](https://github.com/hashi-demo-lab/terraform-provider-bcm/issues/25)
**Priority**: P0 - Critical (required by NVIDIA deployment guide)
**Input**: GitHub Issue #25 - Add etcd_nodes attribute and validate existing attributes against BCM API

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Etcd Node Designation for Production K8s (Priority: P1)

Infrastructure engineers deploying NVIDIA DGX BasePOD need to designate specific nodes as etcd nodes for Kubernetes cluster high availability. The NVIDIA deployment guide specifies 3 etcd nodes for production clusters.

**Why this priority**: This is the primary requirement from the issue - NVIDIA BasePOD deployments require explicit etcd node designation for production-grade Kubernetes clusters. Without this, clusters cannot meet NVIDIA's HA requirements.

**Independent Test**: Can be fully tested by creating a cluster with `etcd_nodes` attribute specifying 3 node UUIDs, verifying the etcd configuration via BCM API, and validating etcd cluster health.

**Acceptance Scenarios**:

1. **Given** a Terraform config with `etcd_nodes` list containing 3 node UUIDs, **When** cluster is created, **Then** BCM configures those nodes as etcd members
2. **Given** no `etcd_nodes` specified, **When** cluster is created, **Then** etcd runs on master nodes (default behavior)
3. **Given** an existing cluster, **When** `etcd_nodes` is updated, **Then** etcd membership changes to new node set
4. **Given** imported cluster with etcd nodes, **When** import completes, **Then** `etcd_nodes` is populated in state

---

### User Story 2 - Attribute Validation Against BCM API (Priority: P1)

Infrastructure engineers need assurance that all Terraform attributes actually work with the BCM API. Unsupported attributes should be clearly documented to avoid configuration surprises.

**Why this priority**: Without API validation, users may configure attributes that BCM ignores, leading to deployment failures and wasted debugging time. This is critical for production reliability.

**Independent Test**: Can be tested by creating clusters with each optional attribute and verifying via BCM API that the attribute value is persisted and returned.

**Acceptance Scenarios**:

1. **Given** `addons` JSON configuration, **When** cluster is created, **Then** BCM API accepts and returns the addons configuration
2. **Given** `storage_classes` JSON configuration, **When** cluster is created, **Then** storage classes are configured in cluster
3. **Given** `ingress_controller` JSON configuration, **When** cluster is created, **Then** ingress controller is deployed
4. **Given** unsupported attribute in config, **When** user applies config, **Then** documentation clearly indicates BCM API limitation

---

### User Story 3 - Version and CNI Configuration (Priority: P2)

Platform engineers need to specify Kubernetes version and CNI plugin for consistent cluster deployments across environments.

**Why this priority**: Version and CNI are commonly configured but have reasonable defaults. Testing validates these work correctly with BCM API.

**Independent Test**: Can be tested by creating cluster with specific `version` and `cni_plugin`, then verifying configuration via BCM API response.

**Acceptance Scenarios**:

1. **Given** `version = "1.28.0"`, **When** cluster is created, **Then** BCM deploys specified K8s version
2. **Given** `cni_plugin = "calico"`, **When** cluster is created, **Then** BCM configures Calico CNI
3. **Given** invalid version format, **When** plan is run, **Then** validation error is returned

---

### User Story 4 - Networking Configuration (Priority: P2)

Network engineers need to configure DNS servers, overlay network, and load balancer mode for cluster networking requirements.

**Why this priority**: Network configuration is important for production but has defaults that work for basic clusters.

**Independent Test**: Can be tested by creating cluster with custom `dns_servers`, `overlay_network`, and `load_balancer_mode`, then verifying via BCM API.

**Acceptance Scenarios**:

1. **Given** `dns_servers = ["8.8.8.8", "8.8.4.4"]`, **When** cluster is created, **Then** cluster DNS uses specified servers
2. **Given** `overlay_network` UUID, **When** cluster is created, **Then** pods use specified overlay network
3. **Given** `load_balancer_mode = "metallb"`, **When** cluster is created, **Then** MetalLB is configured for services

---

### Edge Cases

- **What happens when etcd_nodes overlap with master_nodes?** BCM should handle this gracefully - etcd can run on master nodes
- **What happens when etcd_nodes contains invalid UUIDs?** Pre-flight validation should catch invalid node references
- **What happens when etcd node count is not 1, 3, or 5?** BCM may require odd numbers for quorum - document behavior
- **What happens when JSON attributes contain invalid schema?** BCM validation should return descriptive errors
- **How does BCM handle empty lists vs null for optional attributes?** Document and handle consistently

## Requirements *(mandatory)*

### Functional Requirements

#### New etcd_nodes Attribute

- **FR-001**: Resource MUST add `etcd_nodes` attribute as optional list of node UUID strings
- **FR-002**: Resource MUST default `etcd_nodes` to master nodes when not specified (BCM default behavior)
- **FR-003**: Resource MUST include `etcd_nodes` in BCM API entity when creating/updating clusters
- **FR-004**: Resource MUST read `etcd_nodes` from BCM API response if returned
- **FR-005**: Resource MUST handle `etcd_nodes` in Import to populate from existing clusters
- **FR-006**: Resource schema description MUST document etcd node requirements (odd numbers, minimum 1, recommended 3 for HA)

#### Attribute Validation

- **FR-007**: Resource MUST test `addons` attribute against BCM API and document support status
- **FR-008**: Resource MUST test `storage_classes` attribute against BCM API and document support status
- **FR-009**: Resource MUST test `ingress_controller` attribute against BCM API and document support status
- **FR-010**: Resource MUST test `version` attribute against BCM API and document support status
- **FR-011**: Resource MUST test `cni_plugin` attribute against BCM API and document support status
- **FR-012**: Resource MUST test `overlay_network` attribute against BCM API and document support status
- **FR-013**: Resource MUST test `dns_servers` attribute against BCM API and document support status
- **FR-014**: Resource MUST test `load_balancer_mode` attribute against BCM API and document support status

#### Schema Updates

- **FR-015**: Schema descriptions MUST indicate BCM API support status for each attribute
- **FR-016**: Unsupported attributes MUST include "Note: Not currently supported by BCM API" in description
- **FR-017**: Resource MUST use proper validators for JSON-encoded fields (valid JSON syntax)

### Key Entities *(mandatory)*

- **etcd_nodes**: List of node UUIDs designated as etcd cluster members. Etcd stores Kubernetes cluster state and requires odd number of members (1, 3, 5) for quorum consensus. Production clusters should use 3 etcd nodes for high availability.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `etcd_nodes` attribute is added to schema and functions correctly in CRUD operations
- **SC-002**: All 8 optional attributes are validated against BCM API with documented support status
- **SC-003**: Acceptance test verifies etcd node designation with 3 nodes
- **SC-004**: Schema descriptions updated to reflect actual BCM API support
- **SC-005**: All existing tests continue to pass: `TF_ACC=1 go test -v ./internal/provider/ -run CMKubeCluster`
- **SC-006**: Documentation regenerated with `make generate` showing new attribute
- **SC-007**: Example configuration updated to demonstrate etcd_nodes usage

### Quality Metrics

- **SC-008**: Zero regression in existing CMKubeCluster tests
- **SC-009**: New test follows modern terraform-plugin-testing patterns (statecheck, knownvalue)
- **SC-010**: Test uses dynamically discovered node UUIDs (not hardcoded)

## Assumptions *(mandatory)*

### API Behavior Assumptions

- **A-001**: BCM cmkube API supports `etcdNodes` field in KubeCluster entity (requires verification)
- **A-002**: BCM API returns `etcdNodes` in getKubeCluster response for state population
- **A-003**: When `etcdNodes` is not specified, BCM defaults to running etcd on master nodes
- **A-004**: BCM validation enforces odd number of etcd nodes for quorum

### Implementation Assumptions

- **A-005**: Field name mapping: Terraform `etcd_nodes` maps to BCM API `etcdNodes`
- **A-006**: Test environment has at least 4 available nodes (1 master + 3 etcd minimum)
- **A-007**: Existing helper functions (getTestMasterNodeUUID, etc.) can be extended for etcd nodes

## Out of Scope *(mandatory)*

- **OS-001**: Etcd backup/restore functionality
- **OS-002**: Etcd cluster health monitoring
- **OS-003**: Automatic etcd node replacement on failure
- **OS-004**: Etcd encryption configuration
- **OS-005**: Adding new attributes not in current schema

## Dependencies *(mandatory)*

### External Dependencies

- **D-001**: BCM API must support `etcdNodes` field in KubeCluster entity
- **D-002**: Test environment must have minimum 4 available nodes

### Internal Dependencies

- **D-003**: Existing bcm_cmkube_cluster resource implementation
- **D-004**: Test helper functions for node UUID discovery
- **D-005**: BCM client CallJSONRPC with args parameter

## Research Questions for Phase 0 *(mandatory)*

### API Contract Verification

- **RQ-001**: Does BCM cmkube API support `etcdNodes` field in KubeCluster entity?
- **RQ-002**: What is the exact field name in BCM API? (etcdNodes, etcd_nodes, EtcdNodes?)
- **RQ-003**: Does BCM return `etcdNodes` in getKubeCluster response?
- **RQ-004**: Does BCM enforce odd number requirement for etcd nodes?
- **RQ-005**: What happens when `etcdNodes` is omitted vs explicitly empty?

### Attribute Support Validation

- **RQ-006**: Does BCM API accept and persist `addons` JSON field?
- **RQ-007**: Does BCM API accept and persist `storageClasses` JSON field?
- **RQ-008**: Does BCM API accept and persist `ingressController` JSON field?
- **RQ-009**: Does BCM API accept and persist `version` field?
- **RQ-010**: Does BCM API accept and persist `cniPlugin` field?
- **RQ-011**: Does BCM API accept and persist `overlayNetwork` field?
- **RQ-012**: Does BCM API accept and persist `dnsServers` field?
- **RQ-013**: Does BCM API accept and persist `loadBalancerMode` field?

### Test Environment

- **RQ-014**: How many available nodes exist in test BCM for etcd testing?
- **RQ-015**: What node UUIDs can be used for etcd designation tests?

## Phase 0 Exploration Strategy *(mandatory)*

To answer research questions:

1. **API Exploration**: Run `sampleRest/cmkube-crud-test.py` with `etcdNodes` field added to test entity
2. **Field Discovery**: Inspect BCM API response to identify all supported fields
3. **Attribute Testing**: Test each optional attribute individually to verify BCM support
4. **Document Findings**: Update spec with actual API behavior
5. **Update Implementation**: Modify resource based on findings

## BCM API Contract *(VERIFIED - Phase 0 Complete)*

### API Research Findings (2025-11-25)

**Research Script**: `sampleRest/cmkube-etcd-test.py`

#### RQ-001: etcdNodes Field Support

**CONFIRMED**: BCM API accepts `etcdNodes` field in addKubeCluster.

```json
// Request with etcdNodes (SUCCESS)
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
      "uuid": "client-generated-uuid",
      "name": "cluster-name",
      "masterNodes": ["master-node-uuid"],
      "etcdNodes": ["etcd-node-uuid-1", "etcd-node-uuid-2", "etcd-node-uuid-3"]
    },
    false
  ]
}

// Response
{
  "success": true,
  "task_uuid": "00000000-0000-0000-0000-000000000000",
  "updated_entity": null,
  "validation": []
}
```

#### RQ-003: etcdNodes in getKubeCluster Response

**FINDING**: etcdNodes is **NOT returned** in getKubeCluster response.
This matches the behavior of masterNodes and workerNodes - they are write-only fields.

**Implementation Impact**: etcd_nodes must be preserved from plan/state like master_nodes and worker_nodes.
Cannot rely on API to return this field during Read operations.

#### Attribute Support Matrix

| Terraform Attribute | BCM Field | API Accept | API Return | Support Level |
|---------------------|-----------|------------|------------|---------------|
| etcd_nodes | etcdNodes | YES | NO | WRITE_ONLY |
| version | version | YES | YES | FULL_SUPPORT |
| cni_plugin | cniPlugin | YES | NO | WRITE_ONLY |
| dns_servers | dnsServers | YES | NO | WRITE_ONLY |
| overlay_network | overlayNetwork | YES | NO | WRITE_ONLY |
| load_balancer_mode | loadBalancerMode | YES | NO | WRITE_ONLY |
| storage_classes | storageClasses | YES | NO | WRITE_ONLY |
| addons | addons | YES | NO | WRITE_ONLY |
| ingress_controller | ingressController | YES | NO | WRITE_ONLY |
| master_nodes | masterNodes | YES | NO | WRITE_ONLY |
| worker_nodes | workerNodes | YES | NO | WRITE_ONLY |

### getKubeCluster Response Fields

Fields returned by BCM API:

```json
{
  "appGroups": [],
  "baseType": "KubeCluster",
  "capiNamespace": "default",
  "capiTemplate": false,
  "childType": "",
  "etcdCluster": "00000000-0000-0000-0000-000000000000",
  "external": false,
  "externalIngressServer": "",
  "externalPort": 0,
  "ingressProxyBackendPort": 0,
  "ingressProxyEnable": false,
  "ingressProxyListenPort": 443,
  "internalNetwork": "00000000-0000-0000-0000-000000000000",
  "kubeCluster": "00000000-0000-0000-0000-000000000000",
  "kubeDnsIp": "0.0.0.0",
  "kubernetesApiServer": "",
  "kubernetesApiServerProxyPort": 6444,
  "labelSets": [],
  "modified": false,
  "moduleFileTemplate": "",
  "name": "cluster-name",
  "notes": "",
  "podNetwork": "00000000-0000-0000-0000-000000000000",
  "podNetworkNodeMask": "",
  "revision": "",
  "serviceNetwork": "00000000-0000-0000-0000-000000000000",
  "to_be_removed": false,
  "trustedDomains": [],
  "users": [],
  "uuid": "cluster-uuid",
  "version": "1.28.0"
}
```

**Note**: masterNodes, workerNodes, etcdNodes, cniPlugin, dnsServers, overlayNetwork, loadBalancerMode, storageClasses, addons, ingressController are NOT in the response.

## Implementation Notes

### Existing Resource Location
- **File**: `internal/provider/resource_cmkube_cluster.go`
- **Test File**: `internal/provider/resource_cmkube_cluster_test.go`
- **Example**: `examples/resources/bcm_cmkube_cluster/resource.tf`

### Key Implementation Areas

1. **Schema**: Add `etcd_nodes` attribute to schema
2. **Model**: Add `EtcdNodes` field to `CMKubeClusterResourceModel`
3. **Build Entity**: Include `etcdNodes` in `buildClusterEntity` function
4. **Read**: Parse `etcdNodes` from BCM response in `readCluster`
5. **Tests**: Add acceptance test for etcd node designation
6. **Docs**: Update example and regenerate documentation
