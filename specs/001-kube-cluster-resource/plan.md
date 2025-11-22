# Implementation Plan: BCM Kubernetes Cluster Resource

**Branch**: `001-kube-cluster-resource` | **Date**: 2025-11-22 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/workspace/specs/001-kube-cluster-resource/spec.md`

## Summary

This feature implements a Terraform resource `bcm_cmkube_cluster` for managing Kubernetes clusters in BCM (Bright Cluster Manager). The resource follows TDD principles with acceptance-test-first development, supporting full CRUD operations (Create, Read, Update, Delete) plus Import and drift detection. The implementation leverages existing BCM client infrastructure and follows established patterns from `resource_cmpart_softwareimage.go` and `resource_cmdevice_category.go`.

**Primary Requirement**: Enable infrastructure-as-code management of Kubernetes clusters with declarative configuration for cluster topology (master/worker nodes), networking, and Kubernetes version.

**Technical Approach**: Use BCM JSON-RPC API (`cmkube` service) for cluster lifecycle management with efficient direct UUID lookups for Read operations, full entity structure for Create/Update operations, and polling mechanisms for eventual consistency. All development follows RED-GREEN-REFACTOR cycles with modern terraform-plugin-testing patterns (statecheck, knownvalue, plancheck).

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**:
- terraform-plugin-framework v1.16.1 (resource schema, CRUD operations)
- terraform-plugin-testing v1.13.3 (modern test patterns)
- terraform-plugin-log (structured logging)

**Storage**: BCM cluster state managed via JSON-RPC API
**Testing**:
- Acceptance tests with TF_ACC=1 (terraform-plugin-testing)
- Modern patterns: statecheck.ExpectKnownValue, plancheck.ExpectEmptyPlan, compare.ValuesSame

**Target Platform**: Linux server (BCM cluster endpoint)
**Project Type**: Single (Terraform provider resource)

**Performance Goals**:
- Cluster creation: <30 minutes (BCM provisioning time)
- Read operations: <5 seconds (direct UUID lookup)
- State refresh: <10 seconds
- Test suite execution: <15 minutes (full acceptance tests)

**Constraints**:
- BCM API eventual consistency (cluster provisioning is asynchronous)
- Minimum 2 available nodes in test environment (1 master + 1 worker)
- Cookie-based authentication with self-signed certs (insecure_skip_verify)
- No partial failures allowed (atomic state updates)

**Scale/Scope**:
- MVP: Basic cluster with name + master nodes + optional workers
- 10+ concurrent cluster operations without state corruption
- All CRUD operations following existing BCM resource patterns
- Zero hardcoded test values (portable across environments)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Constitution Gates** (from `.specify/memory/constitution.md`):

✅ **TDD Compliance**:
- RED-GREEN-REFACTOR workflow enforced
- Acceptance tests written BEFORE implementation
- All CRUD operations tested (Create, Read, Update, Delete, Import, Drift)

✅ **Test-First Development**:
- Tests define behavior specifications
- Minimal implementation to pass tests (GREEN phase)
- Refactoring only after tests pass

✅ **Modern Testing Patterns**:
- statecheck.ExpectKnownValue for type-safe assertions
- plancheck.ExpectEmptyPlan for idempotency verification
- compare.ValuesSame for ID consistency tracking
- No hardcoded test values (generateUniqueTestName pattern)

✅ **Architecture Simplicity**:
- Single resource implementation (no additional abstractions)
- Reuse existing BCM client and test helpers
- Follow established resource patterns (no new patterns invented)

✅ **No Premature Optimization**:
- Direct API calls (no caching layer in MVP)
- Simple polling for eventual consistency (no complex state machines)
- Minimal schema in MVP (advanced features marked P2/P3 in spec)

**Re-evaluation After Phase 1**: Will verify schema design remains minimal and follows existing patterns without introducing unnecessary complexity.

## Project Structure

### Documentation (this feature)

```text
specs/001-kube-cluster-resource/
├── plan.md              # This file (/speckit.plan output)
├── research.md          # Phase 0 API exploration findings
├── data-model.md        # Phase 1 schema and entity mappings
├── quickstart.md        # Phase 1 developer quick start guide
└── contracts/           # Phase 1 API contracts and examples
    ├── create-cluster.json
    ├── read-cluster.json
    ├── update-cluster.json
    └── delete-cluster.json
```

### Source Code (repository root)

```text
internal/provider/
├── resource_cmkube_cluster.go          # Resource implementation (CRUD)
├── resource_cmkube_cluster_test.go     # Acceptance tests (TDD-first)
├── bcm_client.go                        # Existing - BCM JSON-RPC client
├── test_helpers.go                      # Existing - shared test utilities
└── provider.go                          # Update to register new resource

examples/
├── resources/
│   └── bcm_cmkube_cluster/
│       ├── resource.tf                  # Basic cluster example
│       ├── import.sh                    # Import existing cluster example
│       └── advanced.tf                  # Advanced configuration example
└── provider/
    └── provider.tf                      # Provider configuration

sampleRest/
├── cmkube-get-clusters.py              # Phase 0 - List clusters exploration
├── cmkube-crud-test.py                 # Phase 0 - Full CRUD test script
└── cmkube-cluster-example.json         # Phase 0 - Sample cluster entity

docs/
└── resources/
    └── bcm_cmkube_cluster.md           # Auto-generated (make generate)
```

**Structure Decision**: This is a single resource addition to an existing Terraform provider. We follow the standard provider structure with resource implementation in `internal/provider/`, acceptance tests co-located, examples in `examples/resources/`, and auto-generated documentation in `docs/`. All Phase 0 exploration scripts go in `sampleRest/` following existing patterns.

## Complexity Tracking

No constitution violations requiring justification. This feature:
- Adds one resource to existing provider (within normal scope)
- Reuses all existing patterns and infrastructure
- Follows established TDD workflow
- Maintains architectural simplicity

## Phase 0: API Exploration & Research

**Objective**: Answer all Research Questions from spec.md by exploring BCM cmkube API and documenting actual behavior.

**NEEDS CLARIFICATION Items from Technical Context**: None - all technical context is known from existing provider infrastructure.

### Research Tasks

#### RQ-001 to RQ-005: API Contract Verification

**Task**: Explore cmkube API methods and document exact entity structure

**Scripts to Create**:

1. **`sampleRest/cmkube-get-clusters.py`** - List all clusters
   ```python
   # Call cmkube.getKubeClusters (plural) to list all clusters
   # Document response structure, field names, data types
   # Identify if response is array of cluster objects or different format
   ```

2. **`sampleRest/cmkube-crud-test.py`** - Full lifecycle test
   ```python
   # Test sequence:
   # 1. Call addKubeCluster with minimal entity
   #    - Document required fields vs optional fields
   #    - Capture UUID returned
   #    - Note if operation is synchronous or async
   # 2. Call getKubeCluster with UUID (test args pattern)
   #    - Verify direct lookup works (not list+filter)
   #    - Document full response structure
   # 3. Call validateKubeCluster with cluster entity
   #    - Document validation behavior
   #    - Determine if useful for plan phase
   # 4. Call updateKubeCluster with modified entity
   #    - Test force parameter behavior
   #    - Document required fields in update
   # 5. Call removeKubeCluster with UUID
   #    - Test force parameter
   #    - Note sync vs async behavior
   # 6. Clean up test cluster
   ```

**Deliverables**:
- `research.md` with findings for RQ-001 to RQ-005
- Sample JSON files in `contracts/` showing actual API requests/responses
- Documentation of exact field names (masterNodes vs master_nodes vs masters)

#### RQ-006 to RQ-010: Field Mappings and Data Types

**Task**: Document exact BCM API field names and Terraform mappings

**Script**: Extend `cmkube-get-clusters.py` to inspect existing cluster if available

**Questions to Answer**:
- Master node field: `masterNodes`, `master_nodes`, or `masters`?
- Worker node field: `workerNodes`, `worker_nodes`, or `workers`?
- Node reference format: Array of UUID strings? Nested objects?
- Kubernetes version format: Semver string, enum, or code?
- Network fields: `managementNetwork`, `overlayNetwork`, CNI config?
- Optional fields: DNS servers, storage, addons, load balancer?

**Deliverables**:
- Field mapping table in `data-model.md`
- JSON schema for KubeCluster entity
- Terraform schema attribute definitions

#### RQ-011 to RQ-014: Operational Behavior

**Task**: Test async operations and timing

**Script**: `cmkube-crud-test.py` with timing instrumentation

**Questions to Answer**:
- Is creation synchronous? If async, how to poll for status?
- Typical creation time in test environment?
- Can updates occur during PROVISIONING state?
- Force parameter actual behavior vs assumptions?
- Full entity replacement (PUT) vs partial update (PATCH)?

**Deliverables**:
- Polling strategy documented in `research.md`
- Timeout values for Create/Update/Delete operations
- State transition diagram if async

#### RQ-015 to RQ-018: Error Handling

**Task**: Test failure scenarios and error responses

**Script**: `cmkube-crud-test.py` with invalid inputs

**Test Cases**:
- Invalid UUID references (nodes, networks)
- Malformed cluster name
- Missing required fields
- Conflicting configuration
- Eventual consistency vs hard failures

**Deliverables**:
- Error code catalog in `research.md`
- Error handling strategy for implementation
- Retry logic recommendations

#### RQ-019 to RQ-021: Test Environment

**Task**: Survey test BCM cluster for available resources

**Script**: Query existing data sources

```python
# Call cmdevice.getNodes to list available nodes
# Call cmnet.getNetworks to list networks
# Call cmkube.getKubeClusters to check current clusters
# Document supported Kubernetes versions
```

**Deliverables**:
- Test environment inventory in `research.md`
- Available node UUIDs for test fixtures
- Network UUIDs for test configurations
- Supported Kubernetes versions list

### Phase 0 Outputs

All findings consolidated into:

1. **`research.md`** - Comprehensive research findings
   - Decision: Field mappings, API patterns, operational behavior
   - Rationale: Why chosen based on BCM API exploration
   - Alternatives: What else was evaluated

2. **`contracts/`** - API contract examples
   - `create-cluster.json` - addKubeCluster request/response
   - `read-cluster.json` - getKubeCluster request/response
   - `update-cluster.json` - updateKubeCluster request/response
   - `delete-cluster.json` - removeKubeCluster request/response

3. **Update `spec.md`** - Populate "BCM API Contract" section with actual examples

**Success Criteria**: All NEEDS CLARIFICATION items resolved, all Research Questions answered with evidence from API exploration.

## Phase 1: Design & Contracts

**Prerequisites**: `research.md` complete with all API findings

### 1.1 Data Model Design

**Task**: Extract entities from feature spec and Phase 0 research

**Output**: `data-model.md`

**Content Structure**:

```markdown
# Data Model: BCM Kubernetes Cluster Resource

## Entities

### KubeCluster (Primary Entity)

**Terraform Resource**: `bcm_cmkube_cluster`

**Schema Attributes**:

| Terraform Attribute | Type | BCM API Field | Required | Computed | Description |
|---------------------|------|---------------|----------|----------|-------------|
| id | string | (same as uuid) | - | ✓ | Resource identifier |
| uuid | string | uuid | - | ✓ | BCM-assigned cluster UUID |
| name | string | name | ✓ | - | Cluster name (validated) |
| master_nodes | list(string) | masterNodes | ✓ | - | Master node UUIDs |
| worker_nodes | list(string) | workerNodes | - | - | Worker node UUIDs |
| management_network | string | managementNetwork | - | - | Network UUID |
| version | string | version | - | - | Kubernetes version |
| force | bool | (args param) | - | - | Bypass validation warnings |

**BCM Entity Structure** (from Phase 0 research):
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
  "version": "1.28.0"
}
```

**Validation Rules**:
- name: Alphanumeric, hyphens, underscores only (regex: `^[a-zA-Z0-9_-]+$`)
- version: Semver format if specified (regex: `^\d+\.\d+\.\d+$`)
- master_nodes: Minimum 1 UUID required
- All UUIDs: Valid UUID v4 format

**State Transitions**: [Document if async operations discovered in Phase 0]
```

### 1.2 API Contract Design

**Task**: Generate API contracts from functional requirements using Phase 0 findings

**Output**: `contracts/` directory with JSON examples

**Files**:

1. **`contracts/create-cluster.json`**
   ```json
   {
     "request": {
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
           "masterNodes": ["<node-uuid>"],
           "workerNodes": [],
           "managementNetwork": "<network-uuid>",
           "version": "1.28.0"
         },
         false
       ]
     },
     "response": {
       "success": true,
       "result": "<cluster-uuid>"
     }
   }
   ```

2. **`contracts/read-cluster.json`** - getKubeCluster with args pattern
3. **`contracts/update-cluster.json`** - updateKubeCluster with full entity
4. **`contracts/delete-cluster.json`** - removeKubeCluster with UUID

### 1.3 Quick Start Guide

**Task**: Create developer onboarding documentation

**Output**: `quickstart.md`

**Content**:
```markdown
# Quick Start: bcm_cmkube_cluster Resource

## Prerequisites
- BCM endpoint accessible at https://172.21.15.254:8081
- Valid credentials (username/password)
- Minimum 2 available nodes for cluster (1 master + 1 worker)

## Development Setup

1. Set environment variables:
   ```bash
   export TF_ACC=1
   export BCM_ENDPOINT="https://172.21.15.254:8081"
   export BCM_USERNAME="root"
   export BCM_PASSWORD="Hashicorp123!"
   ```

2. Run acceptance tests:
   ```bash
   TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMKubeCluster
   ```

## TDD Workflow

1. RED: Write failing acceptance test in `resource_cmkube_cluster_test.go`
2. GREEN: Implement minimal CRUD in `resource_cmkube_cluster.go`
3. REFACTOR: Improve code quality while keeping tests green

## Example Usage

```hcl
resource "bcm_cmkube_cluster" "example" {
  name         = "my-cluster"
  master_nodes = ["<node-uuid>"]
  worker_nodes = ["<worker-uuid>"]
  version      = "1.28.0"
}
```

## File Locations
- Implementation: `internal/provider/resource_cmkube_cluster.go`
- Tests: `internal/provider/resource_cmkube_cluster_test.go`
- Examples: `examples/resources/bcm_cmkube_cluster/`
```

### 1.4 Agent Context Update

**Task**: Run agent context update script

**Command**:
```bash
.specify/scripts/bash/update-agent-context.sh copilot
```

**Updates**:
- Add cmkube API methods to agent context
- Document KubeCluster entity structure
- Add test patterns for async operations (if applicable)
- Preserve manual additions between markers

### 1.5 Constitution Re-check

**Gate**: Verify Phase 1 design maintains simplicity

**Checklist**:
- ✅ Schema follows existing resource patterns (no new abstractions)
- ✅ No premature optimization (direct API calls)
- ✅ Minimal MVP scope (only required fields in Phase 1)
- ✅ Test patterns match existing resources
- ✅ No additional dependencies introduced

**If violations**: Document in Complexity Tracking table with justification

### Phase 1 Outputs

1. ✅ `data-model.md` - Complete schema and entity mappings
2. ✅ `contracts/` - API request/response examples
3. ✅ `quickstart.md` - Developer onboarding guide
4. ✅ Agent context updated with new technology
5. ✅ Constitution re-check passed

## Phase 2: Acceptance Tests (TDD - RED Phase)

**Prerequisites**: Phase 1 complete, data model and contracts defined

**Objective**: Write comprehensive failing acceptance tests BEFORE implementation

**Test File**: `internal/provider/resource_cmkube_cluster_test.go`

### 2.1 Test Infrastructure Setup

**Code**:
```go
package provider

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "testing"
    "time"

    "github.com/hashicorp/terraform-plugin-testing/helper/resource"
    "github.com/hashicorp/terraform-plugin-testing/knownvalue"
    "github.com/hashicorp/terraform-plugin-testing/plancheck"
    "github.com/hashicorp/terraform-plugin-testing/statecheck"
    "github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
    "github.com/hashicorp/terraform-plugin-testing/compare"
)

// testAccProtoV6ProviderFactories - reuse existing from provider_test.go

func testAccPreCheckCMKubeCluster(t *testing.T) {
    // Verify environment variables
    if v := os.Getenv("BCM_ENDPOINT"); v == "" {
        t.Fatal("BCM_ENDPOINT must be set for acceptance tests")
    }
    if v := os.Getenv("BCM_USERNAME"); v == "" {
        t.Fatal("BCM_USERNAME must be set for acceptance tests")
    }
    if v := os.Getenv("BCM_PASSWORD"); v == "" {
        t.Fatal("BCM_PASSWORD must be set for acceptance tests")
    }
}
```

### 2.2 Basic CRUD Test (Priority: P1)

**Test**: `TestAccCMKubeClusterResource_Basic`

**Coverage**: Create, Read, Import, Update, Delete

**Pattern**: Modern terraform-plugin-testing with statecheck, plancheck, compare

```go
func TestAccCMKubeClusterResource_Basic(t *testing.T) {
    clusterName := generateUniqueTestName("test-cluster")
    clusterNameUpdated := generateUniqueTestName("test-cluster-updated")

    // Get available node UUIDs from test environment
    masterNodeUUID := getTestMasterNodeUUID(t)
    workerNodeUUID := getTestWorkerNodeUUID(t)

    // ID consistency tracking across Create/Import/Update
    compareID := statecheck.CompareValue(compare.ValuesSame())

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheckCMKubeCluster(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:             testAccCheckCMKubeClusterDestroy,
        Steps: []resource.TestStep{
            // Create with minimal config
            {
                Config: testAccCMKubeClusterResourceConfig(clusterName, masterNodeUUID),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "bcm_cmkube_cluster.test",
                        tfjsonpath.New("name"),
                        knownvalue.StringExact(clusterName),
                    ),
                    statecheck.ExpectKnownValue(
                        "bcm_cmkube_cluster.test",
                        tfjsonpath.New("uuid"),
                        knownvalue.NotNull(),
                    ),
                    statecheck.ExpectKnownValue(
                        "bcm_cmkube_cluster.test",
                        tfjsonpath.New("master_nodes"),
                        knownvalue.ListSizeExact(1),
                    ),
                    compareID.AddStateValue(
                        "bcm_cmkube_cluster.test",
                        tfjsonpath.New("id"),
                    ),
                },
            },
            // Idempotency check after Create
            {
                Config: testAccCMKubeClusterResourceConfig(clusterName, masterNodeUUID),
                ConfigPlanChecks: resource.ConfigPlanChecks{
                    PreApply: []plancheck.PlanCheck{
                        plancheck.ExpectEmptyPlan(),
                    },
                },
            },
            // Import
            {
                ResourceName:      "bcm_cmkube_cluster.test",
                ImportState:       true,
                ImportStateVerify: true,
                ConfigStateChecks: []statecheck.StateCheck{
                    compareID.AddStateValue(
                        "bcm_cmkube_cluster.test",
                        tfjsonpath.New("id"),
                    ),
                },
            },
            // Update name
            {
                Config: testAccCMKubeClusterResourceConfig(clusterNameUpdated, masterNodeUUID),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "bcm_cmkube_cluster.test",
                        tfjsonpath.New("name"),
                        knownvalue.StringExact(clusterNameUpdated),
                    ),
                    compareID.AddStateValue(
                        "bcm_cmkube_cluster.test",
                        tfjsonpath.New("id"),
                    ),
                },
            },
            // Idempotency check after Update
            {
                Config: testAccCMKubeClusterResourceConfig(clusterNameUpdated, masterNodeUUID),
                ConfigPlanChecks: resource.ConfigPlanChecks{
                    PreApply: []plancheck.PlanCheck{
                        plancheck.ExpectEmptyPlan(),
                    },
                },
            },
            // Delete tested automatically by TestCase
        },
    })
}

func testAccCMKubeClusterResourceConfig(name, masterNodeUUID string) string {
    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmkube_cluster" "test" {
  name         = %[4]q
  master_nodes = [%[5]q]
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        name,
        masterNodeUUID,
    )
}
```

### 2.3 Drift Detection Test (Priority: P1)

**Test**: `TestAccCMKubeClusterResource_DriftDetection`

**Coverage**: External changes detected, plan proposes correction, apply restores state

```go
func TestAccCMKubeClusterResource_DriftDetection(t *testing.T) {
    clusterName := generateUniqueTestName("test-cluster-drift")
    masterNodeUUID := getTestMasterNodeUUID(t)

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheckCMKubeCluster(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:             testAccCheckCMKubeClusterDestroy,
        Steps: []resource.TestStep{
            // Create cluster
            {
                Config: testAccCMKubeClusterResourceConfigWithVersion(clusterName, masterNodeUUID, "1.28.0"),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "bcm_cmkube_cluster.test",
                        tfjsonpath.New("version"),
                        knownvalue.StringExact("1.28.0"),
                    ),
                },
            },
            // Modify cluster externally via BCM API
            {
                PreConfig: func() {
                    client := createTestBCMClient(t)
                    ctx := context.Background()

                    // Get cluster UUID
                    uuid := getResourceUUIDByName(t, "cmkube", "getKubeCluster", clusterName)

                    // Fetch cluster entity
                    body, _ := client.CallJSONRPC(ctx, "cmkube", "getKubeCluster", uuid)
                    var clusterData map[string]interface{}
                    json.Unmarshal(body, &clusterData)

                    // Modify version externally
                    clusterData["version"] = "1.29.0"

                    // Build entity structure
                    entity := map[string]interface{}{
                        "baseType":      "KubeCluster",
                        "childType":     "",
                        "modified":      true,
                        "to_be_removed": false,
                        "revision":      "",
                        "uuid":          uuid,
                    }
                    for k, v := range clusterData {
                        if k != "uuid" {
                            entity[k] = v
                        }
                    }

                    // Update via API
                    client.CallJSONRPC(ctx, "cmkube", "updateKubeCluster", entity, false)
                    time.Sleep(2 * time.Second) // Eventual consistency

                    t.Logf("[DEBUG] Modified version externally to: 1.29.0")
                },
                Config: testAccCMKubeClusterResourceConfigWithVersion(clusterName, masterNodeUUID, "1.28.0"),
                ConfigPlanChecks: resource.ConfigPlanChecks{
                    PreApply: []plancheck.PlanCheck{
                        plancheck.ExpectNonEmptyPlan(), // Drift detected
                    },
                },
            },
            // Terraform restores desired state
            {
                Config: testAccCMKubeClusterResourceConfigWithVersion(clusterName, masterNodeUUID, "1.28.0"),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "bcm_cmkube_cluster.test",
                        tfjsonpath.New("version"),
                        knownvalue.StringExact("1.28.0"),
                    ),
                },
            },
        },
    })
}
```

### 2.4 Worker Node Management Test (Priority: P2)

**Test**: `TestAccCMKubeClusterResource_WorkerNodes`

**Coverage**: Add/remove worker nodes, verify cluster scales correctly

```go
func TestAccCMKubeClusterResource_WorkerNodes(t *testing.T) {
    clusterName := generateUniqueTestName("test-cluster-workers")
    masterNodeUUID := getTestMasterNodeUUID(t)
    workerNodeUUID1 := getTestWorkerNodeUUID(t, 0)
    workerNodeUUID2 := getTestWorkerNodeUUID(t, 1)

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheckCMKubeCluster(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:             testAccCheckCMKubeClusterDestroy,
        Steps: []resource.TestStep{
            // Create with 1 worker
            {
                Config: testAccCMKubeClusterResourceConfigWithWorkers(
                    clusterName,
                    masterNodeUUID,
                    []string{workerNodeUUID1},
                ),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "bcm_cmkube_cluster.test",
                        tfjsonpath.New("worker_nodes"),
                        knownvalue.ListSizeExact(1),
                    ),
                },
            },
            // Scale to 2 workers
            {
                Config: testAccCMKubeClusterResourceConfigWithWorkers(
                    clusterName,
                    masterNodeUUID,
                    []string{workerNodeUUID1, workerNodeUUID2},
                ),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "bcm_cmkube_cluster.test",
                        tfjsonpath.New("worker_nodes"),
                        knownvalue.ListSizeExact(2),
                    ),
                },
            },
            // Scale down to 0 workers
            {
                Config: testAccCMKubeClusterResourceConfigWithWorkers(
                    clusterName,
                    masterNodeUUID,
                    []string{},
                ),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "bcm_cmkube_cluster.test",
                        tfjsonpath.New("worker_nodes"),
                        knownvalue.ListSizeExact(0),
                    ),
                },
            },
        },
    })
}
```

### 2.5 Validation Tests (Priority: P2)

**Test**: `TestAccCMKubeClusterResource_Validation`

**Coverage**: Invalid inputs rejected during plan phase

```go
func TestAccCMKubeClusterResource_ValidationInvalidName(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheckCMKubeCluster(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config:      testAccCMKubeClusterResourceConfig("invalid name!", "uuid"),
                ExpectError: regexp.MustCompile(`Attribute.*name.*must contain only alphanumeric`),
            },
        },
    })
}

func TestAccCMKubeClusterResource_ValidationInvalidVersion(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheckCMKubeCluster(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config:      testAccCMKubeClusterResourceConfigWithVersion("test", "uuid", "invalid"),
                ExpectError: regexp.MustCompile(`Attribute.*version.*must be valid semver`),
            },
        },
    })
}
```

### 2.6 CheckDestroy Helper

**Code**:
```go
func testAccCheckCMKubeClusterDestroy(s *terraform.State) error {
    client := createTestBCMClient(&testing.T{})

    var errors []string
    resourceCount := 0

    for _, rs := range s.RootModule().Resources {
        if rs.Type != "bcm_cmkube_cluster" {
            continue
        }

        resourceCount++
        id := rs.Primary.ID

        // Verify cluster deleted with exponential backoff
        deleted, err := verifyResourceDeleted(
            context.Background(),
            client,
            "cmkube",
            "removeKubeCluster",
            id,
            4, // retry count
        )

        if err != nil {
            errors = append(errors, fmt.Sprintf(
                "Resource type: %s, ID: %s, Error: %v",
                rs.Type,
                id,
                err,
            ))
        }

        if !deleted {
            errors = append(errors, fmt.Sprintf(
                "Cluster still exists after destroy. Type: %s, ID: %s, Retries: 4",
                rs.Type,
                id,
            ))
        }
    }

    if len(errors) > 0 {
        return fmt.Errorf("CheckDestroy failures:\n  - %s", strings.Join(errors, "\n  - "))
    }

    t.Logf("[DEBUG] CheckDestroy verified %d clusters were deleted", resourceCount)
    return nil
}
```

### 2.7 Test Helper Functions

**Code**:
```go
// getTestMasterNodeUUID queries BCM for available master node
func getTestMasterNodeUUID(t *testing.T) string {
    client := createTestBCMClient(t)
    ctx := context.Background()

    // Query available nodes
    body, err := client.CallJSONRPC(ctx, "cmdevice", "getNodes")
    if err != nil {
        t.Fatalf("Failed to get nodes: %v", err)
    }

    var nodes []map[string]interface{}
    if err := json.Unmarshal(body, &nodes); err != nil {
        t.Fatalf("Failed to parse nodes: %v", err)
    }

    // Return first available node UUID
    if len(nodes) > 0 {
        return nodes[0]["uuid"].(string)
    }

    t.Fatal("No available nodes for test cluster")
    return ""
}

// getTestWorkerNodeUUID queries BCM for available worker node
func getTestWorkerNodeUUID(t *testing.T, index int) string {
    client := createTestBCMClient(t)
    ctx := context.Background()

    body, err := client.CallJSONRPC(ctx, "cmdevice", "getNodes")
    if err != nil {
        t.Fatalf("Failed to get nodes: %v", err)
    }

    var nodes []map[string]interface{}
    if err := json.Unmarshal(body, &nodes); err != nil {
        t.Fatalf("Failed to parse nodes: %v", err)
    }

    if len(nodes) <= index+1 { // +1 because first node is master
        t.Fatalf("Not enough nodes for worker index %d", index)
    }

    return nodes[index+1]["uuid"].(string)
}
```

### Phase 2 Outputs

1. ✅ Complete test file with all test cases (failing)
2. ✅ Test helper functions for node UUID retrieval
3. ✅ CheckDestroy implementation
4. ✅ All tests use modern patterns (statecheck, plancheck, compare)
5. ✅ Zero hardcoded values (portable tests)

**Expected Result**: All tests FAIL (RED phase complete)

## Phase 3: Resource Implementation (TDD - GREEN Phase)

**Prerequisites**: Phase 2 complete, all tests failing

**Objective**: Write minimal implementation to pass all acceptance tests

**File**: `internal/provider/resource_cmkube_cluster.go`

### 3.1 Resource Structure and Boilerplate

**Code**:
```go
// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
    "context"
    "encoding/json"
    "fmt"
    "regexp"
    "time"

    "github.com/google/uuid"
    "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
    "github.com/hashicorp/terraform-plugin-framework/path"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
    "github.com/hashicorp/terraform-plugin-framework/schema/validator"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
    _ resource.Resource                = &CMKubeClusterResource{}
    _ resource.ResourceWithImportState = &CMKubeClusterResource{}
)

// CMKubeClusterResource defines the resource implementation.
type CMKubeClusterResource struct {
    client *BCMClient
}

// CMKubeClusterResourceModel describes the resource data model.
type CMKubeClusterResourceModel struct {
    // Identity fields
    ID   types.String `tfsdk:"id"`   // Computed, same as UUID
    UUID types.String `tfsdk:"uuid"` // Computed, BCM-assigned
    Name types.String `tfsdk:"name"` // Required

    // Node configuration
    MasterNodes types.List `tfsdk:"master_nodes"` // Required, list of UUIDs
    WorkerNodes types.List `tfsdk:"worker_nodes"` // Optional, list of UUIDs

    // Network configuration
    ManagementNetwork types.String `tfsdk:"management_network"` // Optional, UUID

    // Kubernetes configuration
    Version types.String `tfsdk:"version"` // Optional, semver string

    // Operations
    Force types.Bool `tfsdk:"force"` // Optional, default false

    // Computed metadata
    CreationTime types.Int64 `tfsdk:"creation_time"` // Computed
    RevisionID   types.Int64 `tfsdk:"revision_id"`   // Computed
}

// NewCMKubeClusterResource creates a new resource instance.
func NewCMKubeClusterResource() resource.Resource {
    return &CMKubeClusterResource{}
}

// Metadata returns the resource type name.
func (r *CMKubeClusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_cmkube_cluster"
}

// Configure adds the provider configured client to the resource.
func (r *CMKubeClusterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }

    client, ok := req.ProviderData.(*BCMClient)
    if !ok {
        resp.Diagnostics.AddError(
            "Unexpected Resource Configure Type",
            fmt.Sprintf("Expected *BCMClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
        )
        return
    }

    r.client = client
}
```

### 3.2 Schema Definition

**Code**:
```go
// Schema defines the resource schema.
func (r *CMKubeClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "Manages a BCM Kubernetes cluster.\n\n" +
            "Kubernetes clusters in BCM define the cluster topology (master and worker nodes), " +
            "networking configuration, and Kubernetes version for container orchestration workloads.",

        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Computed:            true,
                MarkdownDescription: "Cluster identifier (same as uuid)",
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.UseStateForUnknown(),
                },
            },
            "uuid": schema.StringAttribute{
                Computed:            true,
                MarkdownDescription: "BCM-assigned cluster UUID",
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.UseStateForUnknown(),
                },
            },
            "name": schema.StringAttribute{
                Required:            true,
                MarkdownDescription: "Cluster name (alphanumeric, hyphens, underscores only)",
                Validators: []validator.String{
                    stringvalidator.RegexMatches(
                        regexp.MustCompile(`^[a-zA-Z0-9_-]+$`),
                        "must contain only alphanumeric characters, hyphens, and underscores",
                    ),
                },
            },
            "master_nodes": schema.ListAttribute{
                ElementType:         types.StringType,
                Required:            true,
                MarkdownDescription: "List of master node UUIDs (minimum 1 required)",
            },
            "worker_nodes": schema.ListAttribute{
                ElementType:         types.StringType,
                Optional:            true,
                MarkdownDescription: "List of worker node UUIDs",
            },
            "management_network": schema.StringAttribute{
                Optional:            true,
                MarkdownDescription: "Management network UUID",
            },
            "version": schema.StringAttribute{
                Optional:            true,
                MarkdownDescription: "Kubernetes version (semver format, e.g., '1.28.0')",
                Validators: []validator.String{
                    stringvalidator.RegexMatches(
                        regexp.MustCompile(`^\d+\.\d+\.\d+$`),
                        "must be valid semver format (e.g., '1.28.0')",
                    ),
                },
            },
            "force": schema.BoolAttribute{
                Optional:            true,
                Computed:            true,
                Default:             booldefault.StaticBool(false),
                MarkdownDescription: "Bypass validation warnings during operations (default: false)",
            },
            "creation_time": schema.Int64Attribute{
                Computed:            true,
                MarkdownDescription: "Cluster creation timestamp",
            },
            "revision_id": schema.Int64Attribute{
                Computed:            true,
                MarkdownDescription: "BCM revision ID for optimistic locking",
            },
        },
    }
}
```

### 3.3 Create Method

**Code**:
```go
// Create creates a new Kubernetes cluster.
func (r *CMKubeClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan CMKubeClusterResourceModel

    // Read Terraform plan data
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Build cluster entity for BCM API
    entity, diags := buildClusterEntity(ctx, plan, "")
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Call BCM API to create cluster
    tflog.Debug(ctx, "Creating Kubernetes cluster via BCM API", map[string]interface{}{
        "name": plan.Name.ValueString(),
    })

    body, err := r.client.CallJSONRPC(ctx, "cmkube", "addKubeCluster", entity, plan.Force.ValueBool())
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Creating Kubernetes Cluster",
            fmt.Sprintf("Could not create cluster, unexpected error: %s", err.Error()),
        )
        return
    }

    // Parse response to get cluster UUID
    var result string
    if err := json.Unmarshal(body, &result); err != nil {
        resp.Diagnostics.AddError(
            "Error Parsing Cluster Creation Response",
            fmt.Sprintf("Could not parse UUID from response: %s", err.Error()),
        )
        return
    }

    // Set UUID in state
    plan.UUID = types.StringValue(result)
    plan.ID = types.StringValue(result)

    tflog.Info(ctx, "Created Kubernetes cluster", map[string]interface{}{
        "uuid": result,
        "name": plan.Name.ValueString(),
    })

    // Read back full cluster state from BCM
    // (This populates computed fields like creation_time, revision_id)
    readDiags := r.readCluster(ctx, &plan)
    resp.Diagnostics.Append(readDiags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Save state
    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
```

### 3.4 Read Method

**Code**:
```go
// Read reads the current cluster state from BCM.
func (r *CMKubeClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state CMKubeClusterResourceModel

    // Read current state
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Read cluster from BCM API
    diags := r.readCluster(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Save updated state
    resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// readCluster is a helper function to read cluster state from BCM API
func (r *CMKubeClusterResource) readCluster(ctx context.Context, model *CMKubeClusterResourceModel) diag.Diagnostics {
    var diags diag.Diagnostics

    clusterUUID := model.UUID.ValueString()

    tflog.Debug(ctx, "Reading Kubernetes cluster from BCM API", map[string]interface{}{
        "uuid": clusterUUID,
    })

    // Call BCM API with direct UUID lookup (args pattern)
    body, err := r.client.CallJSONRPC(ctx, "cmkube", "getKubeCluster", clusterUUID)
    if err != nil {
        diags.AddError(
            "Error Reading Kubernetes Cluster",
            fmt.Sprintf("Could not read cluster UUID %s: %s", clusterUUID, err.Error()),
        )
        return diags
    }

    // Parse cluster data
    var clusterData map[string]interface{}
    if err := json.Unmarshal(body, &clusterData); err != nil {
        diags.AddError(
            "Error Parsing Cluster Data",
            fmt.Sprintf("Could not parse cluster response: %s", err.Error()),
        )
        return diags
    }

    // Map BCM API fields to Terraform model
    model.Name = getStringValue(clusterData, "name")

    // Master nodes
    if masterNodes, ok := clusterData["masterNodes"].([]interface{}); ok {
        elements := make([]attr.Value, len(masterNodes))
        for i, node := range masterNodes {
            elements[i] = types.StringValue(node.(string))
        }
        model.MasterNodes, _ = types.ListValue(types.StringType, elements)
    }

    // Worker nodes
    if workerNodes, ok := clusterData["workerNodes"].([]interface{}); ok {
        elements := make([]attr.Value, len(workerNodes))
        for i, node := range workerNodes {
            elements[i] = types.StringValue(node.(string))
        }
        model.WorkerNodes, _ = types.ListValue(types.StringType, elements)
    } else {
        model.WorkerNodes = types.ListNull(types.StringType)
    }

    // Optional fields
    model.ManagementNetwork = getStringValue(clusterData, "managementNetwork")
    model.Version = getStringValue(clusterData, "version")

    // Computed fields
    model.CreationTime = getInt64Value(clusterData, "creationTime")
    model.RevisionID = getInt64Value(clusterData, "revisionID")

    return diags
}
```

### 3.5 Update Method

**Code**:
```go
// Update updates an existing Kubernetes cluster.
func (r *CMKubeClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var plan CMKubeClusterResourceModel
    var state CMKubeClusterResourceModel

    // Read plan and current state
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Build cluster entity with UUID for update
    entity, diags := buildClusterEntity(ctx, plan, state.UUID.ValueString())
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Call BCM API to update cluster
    tflog.Debug(ctx, "Updating Kubernetes cluster via BCM API", map[string]interface{}{
        "uuid": state.UUID.ValueString(),
        "name": plan.Name.ValueString(),
    })

    _, err := r.client.CallJSONRPC(ctx, "cmkube", "updateKubeCluster", entity, plan.Force.ValueBool())
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Updating Kubernetes Cluster",
            fmt.Sprintf("Could not update cluster UUID %s: %s", state.UUID.ValueString(), err.Error()),
        )
        return
    }

    tflog.Info(ctx, "Updated Kubernetes cluster", map[string]interface{}{
        "uuid": state.UUID.ValueString(),
    })

    // Read back updated state
    readDiags := r.readCluster(ctx, &plan)
    resp.Diagnostics.Append(readDiags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Save state
    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
```

### 3.6 Delete Method

**Code**:
```go
// Delete removes a Kubernetes cluster.
func (r *CMKubeClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state CMKubeClusterResourceModel

    // Read current state
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    clusterUUID := state.UUID.ValueString()

    tflog.Debug(ctx, "Deleting Kubernetes cluster via BCM API", map[string]interface{}{
        "uuid": clusterUUID,
    })

    // Call BCM API to delete cluster
    _, err := r.client.CallJSONRPC(ctx, "cmkube", "removeKubeCluster", clusterUUID, state.Force.ValueBool())
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Deleting Kubernetes Cluster",
            fmt.Sprintf("Could not delete cluster UUID %s: %s", clusterUUID, err.Error()),
        )
        return
    }

    tflog.Info(ctx, "Deleted Kubernetes cluster", map[string]interface{}{
        "uuid": clusterUUID,
    })

    // State is automatically cleared by framework after successful Delete
}
```

### 3.7 ImportState Method

**Code**:
```go
// ImportState imports an existing cluster by UUID.
func (r *CMKubeClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    // Import by UUID
    resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)

    // Also set ID to same value
    resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
```

### 3.8 Helper Function - Build Cluster Entity

**Code**:
```go
// buildClusterEntity constructs a BCM KubeCluster entity from Terraform model
func buildClusterEntity(ctx context.Context, model CMKubeClusterResourceModel, uuid string) (map[string]interface{}, diag.Diagnostics) {
    var diags diag.Diagnostics

    entity := map[string]interface{}{
        "baseType":      "KubeCluster",
        "childType":     "",
        "modified":      true,
        "to_be_removed": false,
        "revision":      "",
    }

    // Add UUID if updating
    if uuid != "" {
        entity["uuid"] = uuid
    }

    // Required fields
    entity["name"] = model.Name.ValueString()

    // Master nodes
    var masterNodes []string
    diags.Append(model.MasterNodes.ElementsAs(ctx, &masterNodes, false)...)
    entity["masterNodes"] = masterNodes

    // Optional fields
    if !model.WorkerNodes.IsNull() {
        var workerNodes []string
        diags.Append(model.WorkerNodes.ElementsAs(ctx, &workerNodes, false)...)
        entity["workerNodes"] = workerNodes
    }

    if !model.ManagementNetwork.IsNull() {
        entity["managementNetwork"] = model.ManagementNetwork.ValueString()
    }

    if !model.Version.IsNull() {
        entity["version"] = model.Version.ValueString()
    }

    return entity, diags
}
```

### Phase 3 Outputs

1. ✅ Complete resource implementation (minimal GREEN phase)
2. ✅ All CRUD methods implemented
3. ✅ Helper functions for entity building and state mapping
4. ✅ ImportState support
5. ✅ All acceptance tests PASS (GREEN phase complete)

**Expected Result**: Run `TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMKubeCluster` → ALL TESTS PASS

## Phase 4: Integration & Examples

**Prerequisites**: Phase 3 complete, all tests passing

### 4.1 Provider Registration

**File**: `internal/provider/provider.go`

**Edit**: Add resource to Resources() method

```go
// Resources defines the resources implemented in the provider.
func (p *BCMProvider) Resources(ctx context.Context) []func() resource.Resource {
    return []func() resource.Resource{
        NewCMPartSoftwareImageResource,
        NewCMDeviceCategoryResource,
        NewCMKubeClusterResource,  // ADD THIS LINE
    }
}
```

### 4.2 Example Configurations

**File**: `examples/resources/bcm_cmkube_cluster/resource.tf`

```hcl
# Basic Kubernetes cluster with minimal configuration
resource "bcm_cmkube_cluster" "example" {
  name         = "my-k8s-cluster"
  master_nodes = ["<master-node-uuid>"]

  # Optional: Add worker nodes
  worker_nodes = [
    "<worker-node-uuid-1>",
    "<worker-node-uuid-2>",
  ]

  # Optional: Specify Kubernetes version
  version = "1.28.0"

  # Optional: Management network
  management_network = "<network-uuid>"
}

# Output cluster UUID for reference
output "cluster_uuid" {
  value = bcm_cmkube_cluster.example.uuid
}
```

**File**: `examples/resources/bcm_cmkube_cluster/import.sh`

```bash
#!/bin/bash
# Import an existing Kubernetes cluster by UUID

terraform import bcm_cmkube_cluster.example <cluster-uuid>
```

**File**: `examples/resources/bcm_cmkube_cluster/advanced.tf`

```hcl
# Advanced cluster configuration with multiple workers
resource "bcm_cmkube_cluster" "production" {
  name = "prod-k8s-cluster"

  # High-availability master nodes
  master_nodes = [
    "<master-1-uuid>",
    "<master-2-uuid>",
    "<master-3-uuid>",
  ]

  # Worker nodes for workloads
  worker_nodes = [
    "<worker-1-uuid>",
    "<worker-2-uuid>",
    "<worker-3-uuid>",
    "<worker-4-uuid>",
  ]

  # Kubernetes version
  version = "1.29.0"

  # Management network
  management_network = "<prod-network-uuid>"

  # Force operations (use with caution)
  force = false
}
```

### 4.3 Run Acceptance Tests

**Command**:
```bash
TF_ACC=1 \
  BCM_ENDPOINT="https://172.21.15.254:8081" \
  BCM_USERNAME="root" \
  BCM_PASSWORD="Hashicorp123!" \
  go test -v -timeout 120m ./internal/provider/ -run TestAccCMKubeCluster
```

**Expected Output**:
```
=== RUN   TestAccCMKubeClusterResource_Basic
--- PASS: TestAccCMKubeClusterResource_Basic (45.23s)
=== RUN   TestAccCMKubeClusterResource_DriftDetection
--- PASS: TestAccCMKubeClusterResource_DriftDetection (52.11s)
=== RUN   TestAccCMKubeClusterResource_WorkerNodes
--- PASS: TestAccCMKubeClusterResource_WorkerNodes (38.67s)
=== RUN   TestAccCMKubeClusterResource_ValidationInvalidName
--- PASS: TestAccCMKubeClusterResource_ValidationInvalidName (0.45s)
=== RUN   TestAccCMKubeClusterResource_ValidationInvalidVersion
--- PASS: TestAccCMKubeClusterResource_ValidationInvalidVersion (0.38s)
PASS
ok      github.com/hashicorp/terraform-provider-bcm/internal/provider   136.84s
```

### 4.4 Fix Issues (If Any)

If tests fail:
1. Review error messages in test output
2. Check BCM API responses with `TF_LOG=TRACE`
3. Verify entity structure matches Phase 0 research
4. Fix implementation, re-run tests
5. Repeat until all tests pass

### Phase 4 Outputs

1. ✅ Resource registered in provider
2. ✅ Example configurations created
3. ✅ All acceptance tests pass in real environment
4. ✅ Import functionality verified
5. ✅ Documentation examples tested

## Phase 5: Documentation

**Prerequisites**: Phase 4 complete, all tests passing, examples working

### 5.1 Generate Documentation

**Command**:
```bash
make generate
```

This runs:
1. `copywrite` - Add copyright headers
2. `tfplugindocs generate` - Generate provider documentation
3. `terraform fmt` - Format example files

**Generated File**: `docs/resources/bcm_cmkube_cluster.md`

**Expected Content**:
- Resource description (from schema MarkdownDescription)
- Argument reference (all schema attributes)
- Attribute reference (computed fields)
- Import instructions
- Example usage (from examples/resources/bcm_cmkube_cluster/)

### 5.2 Verify Documentation

**Checklist**:
- ✅ Resource description is clear and accurate
- ✅ All arguments documented with types and descriptions
- ✅ Computed attributes listed
- ✅ Examples render correctly
- ✅ Import instructions included
- ✅ No placeholder text or TODOs

### 5.3 Update CHANGELOG (If Applicable)

**File**: `CHANGELOG.md` (if exists)

```markdown
## [Unreleased]

### Added
- New resource: `bcm_cmkube_cluster` for managing Kubernetes clusters in BCM
  - Supports full CRUD operations
  - Import existing clusters by UUID
  - Drift detection for external changes
  - Master and worker node configuration
  - Kubernetes version management
```

### Phase 5 Outputs

1. ✅ Documentation auto-generated via `make generate`
2. ✅ All examples formatted and validated
3. ✅ Documentation reviewed for completeness
4. ✅ CHANGELOG updated (if applicable)
5. ✅ No manual edits to docs/ directory (all auto-generated)

## Success Criteria Verification

**From spec.md Success Criteria section:**

- ✅ **SC-001**: Infrastructure engineers can create a basic Kubernetes cluster with name and master nodes in under 10 minutes
  - *Verification*: Basic CRUD test completes in <2 minutes (cluster provisioning time varies by BCM)

- ✅ **SC-002**: Resource detects 100% of configuration drift for all managed cluster attributes
  - *Verification*: `TestAccCMKubeClusterResource_DriftDetection` passes

- ✅ **SC-003**: All acceptance tests pass using modern terraform-plugin-testing patterns
  - *Verification*: All tests use statecheck, plancheck, compare patterns

- ✅ **SC-004**: Resource import successfully adopts existing clusters with complete state population
  - *Verification*: Import test step in Basic test passes

- ✅ **SC-005**: Cluster create/update/delete operations complete successfully
  - *Verification*: All CRUD test steps pass

- ✅ **SC-006**: Documentation auto-generated via `make generate` with working examples
  - *Verification*: Phase 5 complete, docs generated

- ✅ **SC-007**: Resource supports minimum 10 concurrent cluster operations
  - *Verification*: Tests use unique names, no hardcoded values

- ✅ **SC-008**: Test suite executes in under 15 minutes
  - *Verification*: Measured during Phase 4 (target: <15min)

- ✅ **SC-009**: Zero hardcoded cluster names, UUIDs, or node references in test suite
  - *Verification*: All tests use generateUniqueTestName and getTestNodeUUID helpers

- ✅ **SC-010**: 100% of CRUD operations follow existing resource patterns
  - *Verification*: Code review against resource_cmpart_softwareimage.go

- ✅ **SC-011**: All optional fields have sensible defaults
  - *Verification*: force defaults to false, workers defaults to empty list

- ✅ **SC-012**: State updates occur atomically
  - *Verification*: No partial state updates in implementation

## Implementation Timeline (Autonomous Execution)

**Total Estimated Time**: 8-12 hours (autonomous, no user input)

| Phase | Tasks | Duration | Dependencies |
|-------|-------|----------|--------------|
| **Phase 0** | API exploration scripts, research.md, contracts/ | 2-3 hours | None |
| **Phase 1** | data-model.md, quickstart.md, agent context | 1-2 hours | Phase 0 complete |
| **Phase 2** | Acceptance tests (RED phase) | 2-3 hours | Phase 1 complete |
| **Phase 3** | Resource implementation (GREEN phase) | 2-3 hours | Phase 2 complete |
| **Phase 4** | Integration, examples, test execution | 1-2 hours | Phase 3 complete |
| **Phase 5** | Documentation generation, verification | 0.5-1 hour | Phase 4 complete |

**Parallelization Opportunities**:
- Phase 0: Multiple API exploration scripts can run concurrently
- Phase 2: Multiple test functions can be written in parallel
- Phase 3: CRUD methods can be stubbed in parallel, then filled in
- Phase 4: Examples can be written while tests run

## Risk Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| BCM API differs from assumptions | High | Phase 0 API exploration validates all assumptions before implementation |
| Async cluster provisioning complexity | Medium | Research polling patterns in Phase 0, implement simple exponential backoff |
| Test environment resource constraints | Medium | Test helpers query available resources dynamically, no hardcoded assumptions |
| Drift detection false positives | Low | Follow existing patterns from resource_cmpart_softwareimage.go |
| Documentation generation issues | Low | Use established `make generate` workflow, verify in Phase 5 |

## Appendix A: Key Decisions

**Decision Log**:

1. **Read Strategy**: Direct UUID lookup using args pattern
   - *Rationale*: Efficient, follows pattern from resource_cmpart_softwareimage.go
   - *Alternative Rejected*: List+filter (slower, unnecessary network overhead)

2. **Test Pattern**: Modern terraform-plugin-testing (statecheck, plancheck)
   - *Rationale*: Type-safe, better error messages, current best practice
   - *Alternative Rejected*: Legacy TestCheckResourceAttr only (less type safety)

3. **Schema Scope**: Minimal MVP (name, nodes, network, version)
   - *Rationale*: Satisfies P1/P2 user stories, reduces initial complexity
   - *Future Work*: P3 features (DNS, storage, addons) in separate iteration

4. **Force Parameter**: Optional boolean (default: false)
   - *Rationale*: Safety-first, explicit opt-in for force operations
   - *Alternative Rejected*: Always force (unsafe), no force option (inflexible)

5. **Validation Strategy**: Schema validators + plan-time checks
   - *Rationale*: Early error detection, clear user feedback
   - *Alternative Rejected*: API-only validation (slower feedback loop)

## Appendix B: BCM API Reference

**To be populated in Phase 0 with actual API findings**

Expected structure:
- Service: `cmkube`
- Methods: `addKubeCluster`, `getKubeCluster`, `updateKubeCluster`, `removeKubeCluster`, `validateKubeCluster`
- Entity: `KubeCluster` with baseType, childType, modified, to_be_removed, revision fields
- Args pattern support: Confirmed in Phase 0

## Appendix C: Test Coverage Matrix

| Test Case | Create | Read | Update | Delete | Import | Drift |
|-----------|--------|------|--------|--------|--------|-------|
| Basic CRUD | ✓ | ✓ | ✓ | ✓ | ✓ | - |
| Drift Detection | ✓ | ✓ | - | ✓ | - | ✓ |
| Worker Nodes | ✓ | ✓ | ✓ | ✓ | - | - |
| Validation (name) | - | - | - | - | - | - |
| Validation (version) | - | - | - | - | - | - |
| Idempotency | - | ✓ | ✓ | - | - | - |
| ID Consistency | ✓ | - | ✓ | - | ✓ | - |

**Coverage**: 100% of CRUD operations, 100% of acceptance criteria from spec.md

---

**Plan Status**: READY FOR AUTONOMOUS EXECUTION
**Next Command**: Begin Phase 0 - Create API exploration scripts
