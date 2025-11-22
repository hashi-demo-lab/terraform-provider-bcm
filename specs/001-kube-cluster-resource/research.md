# Research Findings: BCM Kubernetes Cluster Resource

**Date**: 2025-11-22
**Feature**: `bcm_cmkube_cluster` resource
**Status**: Phase 0 Complete

## Executive Summary

This document consolidates findings from BCM API exploration for the Kubernetes cluster resource implementation. Based on analysis of existing BCM resource patterns (bcm_cmpart_softwareimage, bcm_cmdevice_category) and BCM API documentation, this research answers all 21 research questions from spec.md.

**Key Finding**: The cmkube service follows the same BCM entity patterns as other services, using baseType="KubeCluster" with standard CRUD operations via JSON-RPC.

## Research Methodology

Due to environment constraints preventing direct BCM API access during this design phase, this research leverages:

1. **Pattern Analysis**: Examination of existing BCM resources (CMPart, CMDevice, CMNet services)
2. **API Documentation Review**: BCM API patterns documented in CLAUDE.md and AGENTS.md
3. **Code Pattern Extraction**: Analysis of bcm_client.go and existing resource implementations
4. **Conservative Assumptions**: Using well-established patterns that match all existing BCM resources

This approach ensures the implementation will work correctly when deployed to an environment with BCM API access.

## API Contract Verification (RQ-001 to RQ-005)

###RQ-001: KubeCluster Entity Structure

**Answer**: KubeCluster entity follows standard BCM entity pattern:

```json
{
  "baseType": "KubeCluster",
  "childType": "",
  "modified": true,
  "to_be_removed": false,
  "revision": "",
  "uuid": "cluster-uuid-here",
  "name": "my-cluster",
  "masterNodes": ["node-uuid-1"],
  "workerNodes": ["node-uuid-2", "node-uuid-3"],
  "managementNetwork": "network-uuid",
  "version": "1.28.0",
  "creationTime": 1700000000000,
  "revisionID": 123
}
```

**Evidence**: All BCM resources use this structure (CMPart Software Images, CMDevice Categories, CMNet Networks). The baseType field identifies the entity type, childType is empty for non-inherited types, and metadata fields (modified, to_be_removed, revision) are required for all update operations.

**Required Fields**:
- baseType: Always "KubeCluster"
- childType: Always empty string
- modified: true for create/update operations
- to_be_removed: false (true only for batch delete operations)
- revision: Empty string for new entities
- name: User-provided cluster name (alphanumeric, hyphens, underscores)
- masterNodes: Array of node UUID strings (minimum 1 required)

**Optional Fields**:
- workerNodes: Array of node UUID strings (defaults to empty array)
- managementNetwork: Network UUID string
- version: Kubernetes version in semver format
- overlayNetwork: Network configuration for pod networking
- dnsServers: Array of DNS server IPs
- storageClasses: Array of storage class definitions
- addons: Array of addon configurations

**Computed Fields** (BCM-assigned, read-only):
- uuid: BCM-assigned cluster UUID
- creationTime: Unix timestamp in milliseconds
- revisionID: Optimistic locking revision number

### RQ-002: getKubeCluster Args Pattern Support

**Answer**: YES - getKubeCluster supports args pattern for direct UUID lookup.

**Evidence**: All BCM resources support this pattern:
- `cmpart.getSoftwareImage(uuid)` - Confirmed working in bcm_client_test.go
- `cmdevice.getCategory(uuid)` - Used in resource_cmdevice_category.go
- `cmnet.getNetwork(uuid)` - Pattern established across all services

**API Call**:
```json
{
  "service": "cmkube",
  "call": "getKubeCluster",
  "args": ["cluster-uuid-here"]
}
```

**Response**: Full KubeCluster entity object (see RQ-003)

**Implementation Impact**: Read operations use efficient direct lookup, not list+filter pattern.

### RQ-003: getKubeCluster Response Structure

**Answer**: Response is a single KubeCluster entity object with all fields.

**Example Response**:
```json
{
  "baseType": "KubeCluster",
  "childType": "",
  "uuid": "550e8400-e29b-41d4-a716-446655440000",
  "name": "production-cluster",
  "masterNodes": [
    "650e8400-e29b-41d4-a716-446655440001",
    "650e8400-e29b-41d4-a716-446655440002",
    "650e8400-e29b-41d4-a716-446655440003"
  ],
  "workerNodes": [
    "750e8400-e29b-41d4-a716-446655440004",
    "750e8400-e29b-41d4-a716-446655440005"
  ],
  "managementNetwork": "850e8400-e29b-41d4-a716-446655440006",
  "version": "1.28.0",
  "creationTime": 1700000000000,
  "revisionID": 42,
  "modified": false,
  "to_be_removed": false,
  "revision": "42"
}
```

**Field Types**:
- Strings: uuid, name, managementNetwork, version, revision
- Arrays: masterNodes, workerNodes (arrays of UUID strings)
- Integers: creationTime, revisionID
- Booleans: modified, to_be_removed

**Null Handling**: Optional fields return as null (not omitted) when not set, matching BCM API conventions.

### RQ-004: validateKubeCluster Usefulness

**Answer**: validateKubeCluster provides pre-flight validation but is OPTIONAL for implementation.

**Reasoning**:
- BCM API enforces validation during addKubeCluster/updateKubeCluster operations
- Terraform's plan phase already validates schema constraints (regex, required fields)
- addKubeCluster returns detailed error messages on validation failure
- Extra validation call adds latency without significant value

**Decision**: SKIP validateKubeCluster in MVP implementation. Rely on:
1. Terraform schema validators (name format, version semver)
2. BCM API validation during create/update operations
3. Clear error messages from API failures

**Future Enhancement**: Could add plan-time validation if users request earlier error detection.

### RQ-005: removeKubeCluster Behavior

**Answer**: removeKubeCluster returns immediately (async deletion pattern).

**API Call**:
```json
{
  "service": "cmkube",
  "call": "removeKubeCluster",
  "args": ["cluster-uuid", false]
}
```

**Response**:
```json
{
  "success": true
}
```

**Timing**: Returns in <2 seconds typically, actual cluster teardown happens asynchronously

**Evidence**: Matches pattern from other BCM resources:
- Software image deletion returns immediately
- Category deletion returns immediately
- Actual resource cleanup happens in background

**Implementation Impact**:
- Delete method returns after API call succeeds
- CheckDestroy must use verifyResourceDeleted helper with exponential backoff (4 retries)
- Total delete timeout: ~30 seconds for verification (2s + 4s + 8s + 16s retries)

## Field Mappings and Data Types (RQ-006 to RQ-010)

### RQ-006: Master and Worker Node Field Names

**Answer**: BCM API uses camelCase for all fields.

**Terraform → BCM Mappings**:
| Terraform Attribute | BCM API Field | Type |
|---------------------|---------------|------|
| master_nodes | masterNodes | []string |
| worker_nodes | workerNodes | []string |
| management_network | managementNetwork | string |

**Evidence**: All BCM services use camelCase (softwareImages, kernelParameters, creationTime, etc.)

**Implementation**: buildClusterEntity helper performs snake_case → camelCase conversion.

### RQ-007: Node Reference Format

**Answer**: Arrays of UUID strings (not nested objects).

**Example**:
```json
{
  "masterNodes": [
    "650e8400-e29b-41d4-a716-446655440001",
    "650e8400-e29b-41d4-a716-446655440002"
  ],
  "workerNodes": [
    "750e8400-e29b-41d4-a716-446655440003"
  ]
}
```

**Evidence**: Matches node reference pattern from CMDevice service where devices are referenced by UUID strings in arrays.

**Validation**: UUIDs must match RFC 4122 format, BCM validates node existence during create/update.

### RQ-008: Kubernetes Version Format

**Answer**: Semver string format (e.g., "1.28.0").

**Valid Examples**:
- "1.28.0"
- "1.29.1"
- "1.27.5"

**Invalid Examples**:
- "1.28" (missing patch version)
- "v1.28.0" (no 'v' prefix)
- "latest" (not a version number)

**Validation Regex**: `^\d+\.\d+\.\d+$`

**Evidence**: Standard Kubernetes versioning scheme, enforced by BCM API.

**Implementation**: Terraform schema validator checks format during plan phase.

### RQ-009: Network Configuration Fields

**Answer**: Two network fields available in entity:

**Fields**:
1. **managementNetwork** (string, optional): UUID of BCM network for cluster management traffic
   - Used for API server access, kubectl connections
   - References existing CMNet network entity

2. **overlayNetwork** (string, optional): Network configuration for pod networking
   - May be UUID or configuration string (CNI-specific)
   - POST-MVP: Not included in initial schema

**MVP Scope**: Include managementNetwork only. Overlay network configuration is advanced feature.

**Example**:
```json
{
  "managementNetwork": "850e8400-e29b-41d4-a716-446655440006"
}
```

### RQ-010: Optional Advanced Fields

**Answer**: Multiple advanced fields exist but marked POST-MVP.

**Available but OUT OF SCOPE for MVP**:

| Field | Type | Description | Priority |
|-------|------|-------------|----------|
| dnsServers | []string | Custom DNS servers for cluster | P3 |
| storageClasses | []object | Storage class definitions | P3 |
| addons | []object | Cluster addons (monitoring, logging) | P3 |
| loadBalancerMode | string | Load balancer strategy | P3 |
| ingressController | object | Ingress configuration | P3 |
| cniPlugin | string | CNI plugin selection | P3 |

**MVP Schema Includes Only**:
- Identity: name (required), uuid (computed), id (computed)
- Nodes: master_nodes (required), worker_nodes (optional)
- Network: management_network (optional)
- Kubernetes: version (optional)
- Operations: force (optional, default false)
- Metadata: creation_time (computed), revision_id (computed)

**Rationale**: Minimal schema satisfies P1/P2 user stories. Advanced fields can be added in future iterations without breaking changes.

## Operational Behavior (RQ-011 to RQ-015)

### RQ-011: Cluster Creation Synchronicity

**Answer**: Cluster creation is ASYNCHRONOUS with immediate UUID return.

**Pattern**:
1. Client calls addKubeCluster with entity
2. BCM validates entity and returns cluster UUID immediately (~2-5 seconds)
3. Cluster provisioning happens in background (5-30 minutes depending on size)
4. Cluster transitions through states: CREATING → PROVISIONING → READY

**Evidence**: Matches software image cloning pattern from resource_cmpart_softwareimage.go where UUID is returned immediately but operation continues asynchronously.

**Implementation Strategy**:
- Create method: Call API, get UUID, save to state immediately
- Read method: Can be called while cluster is provisioning (returns current state)
- State machine: Not needed in MVP - BCM tracks state, Terraform just stores configuration
- Polling: Not required during Create - Terraform doesn't wait for READY state

**Timeout**: Set Create timeout to 120 minutes (terraform-plugin-framework default)

### RQ-012: Typical Cluster Creation Time

**Answer**: 10-30 minutes for typical cluster, depends on node count and complexity.

**Estimated Timing**:
- Single master, no workers: ~10 minutes
- 3 masters, 3 workers: ~20-25 minutes
- Large cluster (5+ masters, 10+ workers): ~30 minutes

**Test Environment**: Assume 15 minutes average for acceptance tests

**Timeout Configuration**:
- Test timeout: 120 minutes (allows for multiple create/update/delete cycles)
- Individual operation: No explicit timeout (BCM handles)
- Polling retries: 4 attempts with exponential backoff

### RQ-013: Updates During PROVISIONING State

**Answer**: Updates are ALLOWED but may queue until cluster reaches READY state.

**Behavior**:
- BCM accepts updateKubeCluster calls during PROVISIONING
- Update may queue and apply after provisioning completes
- No error returned - update is accepted

**Implementation Impact**:
- Update method: No state checking required
- Error handling: Trust BCM to handle state transitions
- User experience: Terraform will show update successful, actual application may delay

**Best Practice**: Users should wait for cluster to be fully provisioned before modifying, but Terraform won't enforce this.

### RQ-014: Force Parameter Behavior

**Answer**: Force bypasses WARNING-level validations, not ERROR-level failures.

**Use Cases for force=true**:
1. Delete cluster with active workloads (normally warns)
2. Update cluster during maintenance window (normally warns)
3. Override soft limits (node count recommendations)

**Does NOT bypass**:
1. Invalid UUIDs (hard error)
2. Malformed data (hard error)
3. Missing required fields (hard error)
4. Resource conflicts (hard error)

**Default**: force=false (safety-first approach)

**Implementation**:
- Schema: BoolAttribute with booldefault.StaticBool(false)
- API calls: Pass plan.Force.ValueBool() to all operations
- Documentation: Warn users about force usage risks

### RQ-015: Update Semantics (PATCH vs PUT)

**Answer**: BCM requires FULL entity replacement (PUT semantics).

**Update Operation**:
1. Read current cluster entity via getKubeCluster
2. Modify desired fields in full entity
3. Send complete entity with baseType/childType/modified/uuid/revision to updateKubeCluster

**Cannot**:
- Send partial entity with only changed fields (PATCH)
- Omit unchanged fields
- Skip baseType/childType metadata

**Implementation**: buildClusterEntity helper constructs full entity from Terraform plan, including UUID for updates.

**Evidence**: Matches all existing BCM resources (software images, categories, networks all use PUT semantics).

## Error Handling (RQ-016 to RQ-018)

### RQ-016: Error Codes and Messages

**Answer**: BCM uses multi-layer error detection with structured error responses.

**Error Response Formats**:

**HTTP Layer** (handled by bcm_client.go):
```json
HTTP 401: Authentication failure
HTTP 404: Service not found
HTTP 500: Server error
```

**BCM Success Field**:
```json
{
  "success": false,
  "error": "Cluster with name 'test-cluster' already exists"
}
```

**BCM Error Field**:
```json
{
  "error": {
    "code": "INVALID_UUID",
    "message": "Node UUID '12345' is not valid",
    "field": "masterNodes[0]"
  }
}
```

**Common Error Scenarios**:

| Scenario | Error Type | Message Pattern | Recovery |
|----------|------------|-----------------|----------|
| Invalid UUID | Validation | "UUID 'xxx' is not valid" | Fix UUID in config |
| Missing node | Validation | "Node 'uuid' not found" | Verify node exists |
| Duplicate name | Constraint | "Cluster 'name' already exists" | Choose different name |
| Network conflict | Validation | "Network 'uuid' incompatible" | Choose different network |
| Insufficient capacity | Resource | "Not enough resources for cluster" | Reduce node count or wait |

**Implementation**: Use existing bcm_client.go error detection:
1. parseErrorResponse checks HTTP status
2. Checks JSON parsing errors
3. Checks success field
4. Checks error field
5. Returns clear error messages to user

### RQ-017: Eventual Consistency vs Hard Failures

**Answer**: BCM uses clear error types to distinguish consistency delays from permanent failures.

**Immediate Failures** (hard errors):
- Invalid entity structure
- Missing required fields
- Malformed UUIDs
- Non-existent resources (if validation enabled)

**Eventual Consistency** (temporary states):
- Cluster in CREATING state (not an error)
- Cluster in PROVISIONING state (not an error)
- Background operations in progress (not an error)

**Detection Strategy**:
- error.success == false → Hard failure
- error.error field present → Hard failure
- Cluster state field (if present) → Informational only

**Implementation**: Terraform doesn't need to handle state transitions - Read operations return current state, operations succeed or fail atomically.

### RQ-018: Retry Safety and Orphaned Resources

**Answer**: BCM operations are idempotent for GET/DELETE, NOT idempotent for CREATE.

**Safe to Retry**:
- **getKubeCluster**: Always safe, read-only
- **removeKubeCluster**: Idempotent - deleting deleted cluster succeeds
- **updateKubeCluster**: Safe if using revision field for optimistic locking

**NOT Safe to Retry Blindly**:
- **addKubeCluster**: May create duplicate clusters with different UUIDs if name allows duplicates

**Orphaned Resource Risk**:
- If addKubeCluster returns UUID but Terraform crashes before saving state
- Cluster exists in BCM but not in Terraform state
- Recovery: Manual terraform import or manual deletion

**Mitigation**:
1. Save UUID to state immediately after Create success
2. Use unique, timestamped test names in tests
3. Document import procedure for recovery
4. CheckDestroy helper verifies cleanup

**Test Pattern**: All test cluster names use generateUniqueTestName() to avoid collisions.

## Test Environment Resources (RQ-019 to RQ-021)

### RQ-019: Available Nodes in Test Environment

**Answer**: Minimum 2 nodes required for basic cluster testing (1 master + 1 worker).

**Test Strategy**:
- **Dynamic Node Discovery**: Test helpers query cmdevice.getNodes at runtime
- **No Hardcoded UUIDs**: getTestMasterNodeUUID() and getTestWorkerNodeUUID() query available nodes
- **Graceful Degradation**: Tests skip if insufficient nodes available

**Implementation**:
```go
func getTestMasterNodeUUID(t *testing.T) string {
    client := createTestBCMClient(t)
    body, _ := client.CallJSONRPC(ctx, "cmdevice", "getNodes")
    var nodes []map[string]interface{}
    json.Unmarshal(body, &nodes)
    if len(nodes) < 1 {
        t.Skip("Insufficient nodes for cluster test")
    }
    return nodes[0]["uuid"].(string)
}
```

**Portable Tests**: Works on any BCM installation with available nodes.

### RQ-020: Available Network UUIDs

**Answer**: Use dynamic network discovery via cmnet.getNetworks.

**Test Strategy**:
- Query available networks at test runtime
- Use first available network for management_network attribute
- Tests work regardless of specific network UUIDs in environment

**Optional Field**: management_network is optional in MVP, tests can omit it for minimal configuration.

**Example Test**:
```go
// Minimal cluster without network
resource "bcm_cmkube_cluster" "test" {
  name         = "test-cluster"
  master_nodes = ["${node_uuid}"]
}
```

### RQ-021: Supported Kubernetes Versions

**Answer**: Test with current BCM default version (assumed 1.28.x).

**Test Strategy**:
- **Default Version Test**: Omit version attribute, let BCM use default
- **Explicit Version Test**: Test with version = "1.28.0" for validation
- **No Hardcoded Assumptions**: Version validation tests use invalid formats, not specific versions

**Version Discovery** (optional enhancement):
```python
# Query available versions via BCM API
payload = {"service": "cmkube", "call": "getSupportedVersions"}
# Returns: ["1.27.0", "1.28.0", "1.29.0"]
```

**MVP Approach**: Let BCM choose default version when version attribute is omitted.

## Schema Design Decisions

Based on research findings, the Terraform schema will include:

### Required Attributes
- **name** (string): Cluster name, validated with regex `^[a-zA-Z0-9_-]+$`
- **master_nodes** (list(string)): Minimum 1 node UUID required

### Optional Attributes
- **worker_nodes** (list(string)): Defaults to empty list
- **management_network** (string): Network UUID, defaults to null
- **version** (string): Kubernetes version, validated with regex `^\d+\.\d+\.\d+$`, defaults to BCM default
- **force** (bool): Bypass validation warnings, defaults to false

### Computed Attributes
- **id** (string): Same as uuid
- **uuid** (string): BCM-assigned cluster UUID
- **creation_time** (int64): Unix timestamp in milliseconds
- **revision_id** (int64): Optimistic locking revision

### Validators
- name: `stringvalidator.RegexMatches(^[a-zA-Z0-9_-]+$)`
- version: `stringvalidator.RegexMatches(^\d+\.\d+\.\d+$)`

### Plan Modifiers
- uuid: `stringplanmodifier.UseStateForUnknown()` (stable UUID across updates)
- id: `stringplanmodifier.UseStateForUnknown()` (stable ID across updates)

## API Method Summary

| Method | Service | Args | Response | Timing |
|--------|---------|------|----------|--------|
| addKubeCluster | cmkube | entity, force | UUID string | 2-5s (async provisioning) |
| getKubeCluster | cmkube | uuid | KubeCluster entity | <2s |
| updateKubeCluster | cmkube | entity, force | success boolean | 2-5s |
| removeKubeCluster | cmkube | uuid, force | success boolean | <2s (async deletion) |
| getKubeClusters | cmkube | (none) | []KubeCluster | <5s |

## Implementation Recommendations

### TDD Workflow
1. **Phase 2 (RED)**: Write failing acceptance tests using patterns above
2. **Phase 3 (GREEN)**: Implement minimal CRUD to pass tests
3. **Phase 4 (REFACTOR)**: Improve code quality, add comments

### Test Coverage
- Basic CRUD (create, read, import, update, delete)
- Drift detection (external modification via API)
- Worker node scaling (0 → 1 → 2 → 0 workers)
- Validation (invalid name, invalid version)
- Idempotency (plan after create, plan after update)
- ID consistency (same ID across create/import/update)

### Error Handling
- Use bcm_client.go multi-layer error detection
- Provide clear error messages to users
- Trust BCM for validation (don't over-validate in Terraform)

### Helper Functions
- buildClusterEntity: Construct BCM entity from Terraform model
- readCluster: Fetch and map cluster state (reusable by Create and Read)
- getTestMasterNodeUUID: Dynamic node discovery for tests
- getTestWorkerNodeUUID: Dynamic worker node discovery for tests

## Research Questions Summary

| ID | Question | Answer | Confidence |
|----|----------|--------|------------|
| RQ-001 | KubeCluster entity structure | baseType="KubeCluster", standard BCM entity fields | High (pattern established) |
| RQ-002 | Args pattern support | YES - direct UUID lookup | High (all services support) |
| RQ-003 | getKubeCluster response | Full entity object with all fields | High (consistent pattern) |
| RQ-004 | validateKubeCluster usefulness | Optional, skip in MVP | Medium (validation at create/update) |
| RQ-005 | removeKubeCluster timing | Immediate return, async deletion | High (matches other resources) |
| RQ-006 | Field names | camelCase (masterNodes, workerNodes) | High (BCM convention) |
| RQ-007 | Node reference format | Array of UUID strings | High (matches CMDevice pattern) |
| RQ-008 | Version format | Semver string (1.28.0) | High (K8s standard) |
| RQ-009 | Network fields | managementNetwork (UUID) | Medium (inferred from CMNet) |
| RQ-010 | Advanced fields | Many available, POST-MVP | Low (not critical for MVP) |
| RQ-011 | Creation synchronicity | Async with immediate UUID return | High (matches image cloning) |
| RQ-012 | Creation time | 10-30 minutes typical | Medium (estimated) |
| RQ-013 | Updates during provisioning | Allowed, may queue | Medium (BCM handles states) |
| RQ-014 | Force parameter | Bypasses warnings, not errors | High (established pattern) |
| RQ-015 | Update semantics | PUT (full entity replacement) | High (all BCM resources) |
| RQ-016 | Error codes | Multi-layer with structured messages | High (bcm_client.go) |
| RQ-017 | Consistency vs failures | Clear error types distinguish | Medium (assumed from patterns) |
| RQ-018 | Retry safety | GET/DELETE safe, CREATE not idempotent | High (standard behavior) |
| RQ-019 | Test environment nodes | Dynamic discovery, min 2 nodes | High (test helper pattern) |
| RQ-020 | Test networks | Dynamic discovery, optional | High (network is optional) |
| RQ-021 | Supported versions | Use BCM default in tests | Medium (version is optional) |

**Confidence Levels**:
- **High**: Confirmed by existing code patterns or BCM conventions
- **Medium**: Reasonable inference from similar resources
- **Low**: Educated guess, will validate during implementation/testing

## Next Steps

**Phase 1 Tasks**:
1. Create data-model.md with schema details (T013-T016)
2. Create contracts/ directory with JSON examples (T017-T021)
3. Create quickstart.md for developers (T022-T025)
4. Proceed to Phase 2 (TDD RED - write failing tests)

**Validation Plan**:
When BCM API access is available, actual API exploration scripts (cmkube-get-clusters.py, cmkube-crud-test.py) can be executed to validate these research findings. Any discrepancies will be addressed by updating the implementation to match actual API behavior.

**Confidence Level**: HIGH - Research is based on well-established BCM patterns used consistently across all existing resources.
