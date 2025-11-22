---
description: "Task list for bcm_cmkube_cluster resource implementation"
---

# Tasks: BCM Kubernetes Cluster Resource

**Feature**: `bcm_cmkube_cluster` - Terraform resource for managing Kubernetes clusters in BCM
**Branch**: `001-kube-cluster-resource`
**Input**: Design documents from `/workspace/specs/001-kube-cluster-resource/`

**Prerequisites**:
- spec.md (user stories and requirements) ✅
- plan.md (implementation design) ✅

**TDD Workflow**: RED-GREEN-REFACTOR
- Phase 2: Write failing tests (RED)
- Phase 3: Minimal implementation to pass (GREEN)
- Phase 4: Refactor and polish

**Organization**: Tasks grouped by phase to support autonomous execution without user intervention.

## Format: `- [ ] [ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: User story label (US1, US2, US3) - only for user story implementation tasks
- File paths are absolute

---

## Phase 0: API Exploration & Research

**Purpose**: Validate all BCM API assumptions BEFORE implementation

**⚠️ CRITICAL**: This phase MUST complete before any code is written. All research questions from spec.md must be answered with evidence.

### API Exploration Scripts

- [X] T001 [P] Create /workspace/sampleRest/cmkube-get-clusters.py to list all Kubernetes clusters via cmkube.getKubeClusters API
- [X] T002 [P] Create /workspace/sampleRest/cmkube-crud-test.py to test full cluster lifecycle (add, get, update, validate, remove)
- [X] T003 Execute cmkube-get-clusters.py script to inspect existing cluster structure and document response format
- [X] T004 Execute cmkube-crud-test.py script to verify CRUD operations and document API behavior

### Research Documentation

- [X] T005 Create /workspace/specs/001-kube-cluster-resource/research.md with findings from API exploration
- [X] T006 Document KubeCluster entity structure (baseType, required fields, optional fields, field names)
- [X] T007 Document API method signatures (addKubeCluster, getKubeCluster, updateKubeCluster, removeKubeCluster args)
- [X] T008 Document field mappings (masterNodes vs master_nodes, camelCase vs snake_case)
- [X] T009 Document operational behavior (sync vs async, polling requirements, timeouts)
- [X] T010 Document error codes and failure scenarios
- [X] T011 Query test environment resources (available nodes, networks, supported k8s versions)
- [X] T012 Update /workspace/specs/001-kube-cluster-resource/spec.md BCM API Contract section with actual findings

**Checkpoint**: All Research Questions (RQ-001 through RQ-021) from spec.md must be answered before proceeding.

---

## Phase 1: Design Artifacts

**Purpose**: Create design documentation based on Phase 0 research findings

**Prerequisites**: Phase 0 complete, research.md populated

### Data Model Design

- [X] T013 Create /workspace/specs/001-kube-cluster-resource/data-model.md with KubeCluster entity schema
- [X] T014 Document Terraform attribute to BCM API field mappings in data-model.md
- [X] T015 Document validation rules (name format, version semver, UUID formats) in data-model.md
- [X] T016 Document state transitions if cluster provisioning is asynchronous (based on Phase 0 findings)

### API Contract Examples

- [X] T017 [P] Create /workspace/specs/001-kube-cluster-resource/contracts/ directory
- [X] T018 [P] Create /workspace/specs/001-kube-cluster-resource/contracts/create-cluster.json with addKubeCluster request/response
- [X] T019 [P] Create /workspace/specs/001-kube-cluster-resource/contracts/read-cluster.json with getKubeCluster request/response
- [X] T020 [P] Create /workspace/specs/001-kube-cluster-resource/contracts/update-cluster.json with updateKubeCluster request/response
- [X] T021 [P] Create /workspace/specs/001-kube-cluster-resource/contracts/delete-cluster.json with removeKubeCluster request/response

### Developer Quick Start

- [X] T022 Create /workspace/specs/001-kube-cluster-resource/quickstart.md with developer onboarding guide
- [X] T023 Document TDD workflow (RED-GREEN-REFACTOR) in quickstart.md
- [X] T024 Document environment setup (TF_ACC, BCM_ENDPOINT, credentials) in quickstart.md
- [X] T025 Document example usage and test execution commands in quickstart.md

**Checkpoint**: Design artifacts complete - ready for test implementation (TDD RED phase)

---

## Phase 2: Acceptance Tests (TDD - RED Phase)

**Purpose**: Write comprehensive failing acceptance tests BEFORE implementation

**Prerequisites**: Phase 1 complete, data-model.md and contracts/ exist

**⚠️ CRITICAL TDD REQUIREMENT**: ALL tests in this phase MUST FAIL initially. Do NOT implement resource code yet.

### Test Infrastructure

- [X] T026 Create /workspace/internal/provider/resource_cmkube_cluster_test.go with test package and imports
- [X] T027 Add testAccPreCheckCMKubeCluster function to verify environment variables (BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD)
- [X] T028 Add testAccCheckCMKubeClusterDestroy function with enhanced error messages and exponential backoff verification
- [X] T029 [P] Add getTestMasterNodeUUID helper function to query available master nodes from BCM
- [X] T030 [P] Add getTestWorkerNodeUUID helper function to query available worker nodes from BCM

### Basic CRUD Test (User Story 1 - Priority P1)

- [X] T031 Implement TestAccCMKubeClusterResource_Basic test function with Create/Read/Import/Update/Delete steps
- [X] T032 Add ID consistency tracking using statecheck.CompareValue(compare.ValuesSame())
- [X] T033 Add Create step with minimal config (name + master_nodes) using ConfigStateChecks
- [X] T034 Add idempotency check after Create using plancheck.ExpectEmptyPlan()
- [X] T035 Add Import step with ImportStateVerify and ID tracking
- [X] T036 Add Update step (change cluster name) with ConfigStateChecks
- [X] T037 Add idempotency check after Update using plancheck.ExpectEmptyPlan()
- [X] T038 Create testAccCMKubeClusterResourceConfig helper function returning provider + resource config

### Drift Detection Test (User Story 2 - Priority P1)

- [X] T039 Implement TestAccCMKubeClusterResource_DriftDetection test function
- [X] T040 Add Create step with version attribute
- [X] T041 Add drift injection step using PreConfig to modify cluster via BCM API
- [X] T042 Implement PreConfig function to query cluster UUID, fetch entity, modify version field, update via API
- [X] T043 Add drift verification step using plancheck.ExpectNonEmptyPlan()
- [X] T044 Add restoration step to verify Terraform restores desired state
- [X] T045 Create testAccCMKubeClusterResourceConfigWithVersion helper function

### Worker Node Management Test (User Story 3 - Priority P2)

- [X] T046 Implement TestAccCMKubeClusterResource_WorkerNodes test function
- [X] T047 Add Create step with 1 worker node, verify worker_nodes list size using knownvalue.ListSizeExact(1)
- [X] T048 Add scale-up step to 2 worker nodes, verify list size 2
- [X] T049 Add scale-down step to 0 worker nodes, verify empty list
- [X] T050 Create testAccCMKubeClusterResourceConfigWithWorkers helper function accepting worker node array

### Validation Tests (User Story 4 - Priority P2)

- [X] T051 [P] Implement TestAccCMKubeClusterResource_ValidationInvalidName test with invalid cluster name (special chars)
- [X] T052 [P] Implement TestAccCMKubeClusterResource_ValidationInvalidVersion test with invalid semver format
- [X] T053 [P] Add ExpectError checks for both validation tests using regexp.MustCompile

### Test Verification

- [X] T054 Run all acceptance tests with TF_ACC=1 and verify ALL tests FAIL (RED phase complete)
- [X] T055 Verify error messages indicate missing resource implementation (not test failures)

**Checkpoint**: RED phase complete - All tests written and failing. Ready for implementation (GREEN phase).

---

## Phase 3: Resource Implementation (TDD - GREEN Phase)

**Purpose**: Write minimal implementation to make all acceptance tests PASS

**Prerequisites**: Phase 2 complete, all tests failing

**⚠️ TDD REQUIREMENT**: Write minimal code to pass tests. No premature optimization.

### Resource Structure

- [X] T056 Create /workspace/internal/provider/resource_cmkube_cluster.go with package and imports
- [X] T057 Define CMKubeClusterResource struct with client field
- [X] T058 Define CMKubeClusterResourceModel struct with all schema fields (ID, UUID, Name, MasterNodes, WorkerNodes, etc.)
- [X] T059 Implement NewCMKubeClusterResource factory function
- [X] T060 [P] Implement Metadata method returning resource type name "bcm_cmkube_cluster"
- [X] T061 [P] Implement Configure method to receive BCM client from provider

### Schema Definition

- [X] T062 Implement Schema method with complete resource schema
- [X] T063 Add id attribute (computed, StringType)
- [X] T064 Add uuid attribute (computed, StringType, UseStateForUnknown plan modifier)
- [X] T065 Add name attribute (required, StringType, regex validator for alphanumeric/hyphens/underscores)
- [X] T066 Add master_nodes attribute (required, ListType of strings)
- [X] T067 Add worker_nodes attribute (optional, ListType of strings)
- [X] T068 Add management_network attribute (optional, StringType)
- [X] T069 Add version attribute (optional, StringType, semver regex validator)
- [X] T070 Add force attribute (optional, BoolType, default false)
- [X] T071 [P] Add creation_time attribute (computed, Int64Type)
- [X] T072 [P] Add revision_id attribute (computed, Int64Type)

### Create Method

- [X] T073 Implement Create method signature and plan extraction
- [X] T074 Call buildClusterEntity helper to construct BCM entity from plan
- [X] T075 Make CallJSONRPC to cmkube.addKubeCluster with entity and force parameter
- [X] T076 Parse response to extract cluster UUID
- [X] T077 Set UUID and ID in state
- [X] T078 Call readCluster helper to populate computed fields
- [X] T079 Save final state
- [X] T080 Add error handling and tflog debug/info messages

### Read Method

- [X] T081 Implement Read method signature and state extraction
- [X] T082 Call readCluster helper function
- [X] T083 Save updated state
- [X] T084 Implement readCluster helper function (reusable by Create and Read)
- [X] T085 Make CallJSONRPC to cmkube.getKubeCluster with UUID args parameter (direct lookup)
- [X] T086 Parse cluster data JSON response
- [X] T087 Map BCM fields to Terraform model (name, masterNodes, workerNodes, networks, version)
- [X] T088 Use getStringValue helper for null-safe string fields
- [X] T089 Use getInt64Value helper for null-safe computed fields
- [X] T090 Handle list fields (master_nodes, worker_nodes) with proper type conversion

### Update Method

- [X] T091 Implement Update method signature and extract plan + state
- [X] T092 Call buildClusterEntity helper with UUID from state
- [X] T093 Make CallJSONRPC to cmkube.updateKubeCluster with entity and force parameter
- [X] T094 Call readCluster helper to get updated state
- [X] T095 Save final state
- [X] T096 Add error handling and tflog messages

### Delete Method

- [X] T097 Implement Delete method signature and state extraction
- [X] T098 Make CallJSONRPC to cmkube.removeKubeCluster with UUID and force parameter
- [X] T099 Add error handling and tflog messages
- [X] T100 Note: Framework automatically clears state after successful Delete

### ImportState Method

- [X] T101 Implement ImportState method using resource.ImportStatePassthroughID for uuid attribute
- [X] T102 Set id attribute to same value as uuid for consistency

### Helper Functions

- [X] T103 Implement buildClusterEntity helper function
- [X] T104 Construct entity map with baseType, childType, modified, to_be_removed, revision fields
- [X] T105 Add UUID field if updating (empty string if creating)
- [X] T106 Map Terraform snake_case fields to BCM camelCase (master_nodes → masterNodes)
- [X] T107 Handle optional fields with null checks
- [X] T108 Return entity map and diagnostics

### Test Execution (GREEN Phase Verification)

- [X] T109 Run TestAccCMKubeClusterResource_Basic and verify it PASSES
- [X] T110 Run TestAccCMKubeClusterResource_DriftDetection and verify it PASSES
- [X] T111 Run TestAccCMKubeClusterResource_WorkerNodes and verify it PASSES
- [X] T112 Run TestAccCMKubeClusterResource_ValidationInvalidName and verify it PASSES
- [X] T113 Run TestAccCMKubeClusterResource_ValidationInvalidVersion and verify it PASSES
- [X] T114 Run all acceptance tests together with TF_ACC=1 and verify 100% pass rate

**Checkpoint**: GREEN phase complete - All tests passing with minimal implementation.

---

## Phase 4: Integration & Examples

**Purpose**: Register resource, create examples, verify end-to-end functionality

**Prerequisites**: Phase 3 complete, all tests passing

### Provider Registration

- [X] T115 Edit /workspace/internal/provider/provider.go Resources() method to register NewCMKubeClusterResource
- [X] T116 Verify provider compiles with new resource registered

### Example Configurations

- [X] T117 Create /workspace/examples/resources/bcm_cmkube_cluster/ directory
- [X] T118 Create /workspace/examples/resources/bcm_cmkube_cluster/resource.tf with basic cluster example
- [X] T119 Add example with master_nodes, worker_nodes, and version attributes
- [X] T120 Add output block for cluster_uuid
- [X] T121 Create /workspace/examples/resources/bcm_cmkube_cluster/import.sh with import command example
- [X] T122 Create /workspace/examples/resources/bcm_cmkube_cluster/advanced.tf with multi-master HA cluster example

### Test Execution & Validation

- [X] T123 Run full acceptance test suite with all environment variables set
- [X] T124 Verify test execution completes in under 15 minutes (success criterion SC-008)
- [X] T125 Verify no hardcoded values in test output (success criterion SC-009)
- [X] T126 Verify import functionality works correctly
- [X] T127 Verify drift detection accurately identifies external changes
- [X] T128 Run make build to verify provider compiles successfully

**Checkpoint**: Integration complete - Resource fully functional with examples.

---

## Phase 5: Documentation & Polish

**Purpose**: Generate documentation, validate quality, prepare for merge

**Prerequisites**: Phase 4 complete, all tests passing, examples working

### Documentation Generation

- [X] T129 Run make generate to auto-generate provider documentation
- [X] T130 Verify /workspace/docs/resources/bcm_cmkube_cluster.md was created
- [X] T131 Verify documentation includes resource description, argument reference, attribute reference
- [X] T132 Verify examples from examples/resources/bcm_cmkube_cluster/ are included in docs
- [X] T133 Verify import instructions are documented

### Code Quality

- [X] T134 [P] Run make fmt to format all Go code
- [X] T135 [P] Run make lint to check code quality with golangci-lint
- [X] T136 Fix any linting issues identified
- [X] T137 Run pre-commit hooks to verify code quality gates pass
- [X] T138 Review code against existing patterns (resource_cmpart_softwareimage.go, resource_cmdevice_category.go)

### Final Validation

- [X] T139 Verify all success criteria from spec.md are met (SC-001 through SC-012)
- [X] T140 Verify all functional requirements from spec.md are implemented (FR-001 through FR-024)
- [X] T141 Run quickstart.md validation (verify developer can follow guide successfully)
- [X] T142 Verify zero TODOs or placeholder comments in implementation
- [X] T143 Verify error messages are clear and actionable

**Checkpoint**: Documentation complete - Feature ready for review/merge.

---

## Dependencies & Execution Order

### Phase Dependencies

1. **Phase 0 (API Exploration)**: No dependencies - START HERE
   - MUST complete before any implementation work
   - All research questions must be answered

2. **Phase 1 (Design Artifacts)**: Depends on Phase 0 complete
   - Cannot proceed without API contract knowledge

3. **Phase 2 (Tests - RED)**: Depends on Phase 1 complete
   - TDD requirement: Tests FIRST
   - All tests MUST fail before implementation

4. **Phase 3 (Implementation - GREEN)**: Depends on Phase 2 complete
   - TDD requirement: Minimal implementation to pass tests
   - Tests guide the implementation

5. **Phase 4 (Integration)**: Depends on Phase 3 complete
   - All tests must be passing

6. **Phase 5 (Documentation)**: Depends on Phase 4 complete
   - Final polish and quality checks

### Task Dependencies Within Phases

**Phase 0**:
- T001, T002 can run in parallel (P)
- T003 depends on T001
- T004 depends on T002
- T005-T012 sequential (documentation tasks)

**Phase 1**:
- T013-T016 sequential (data model)
- T017-T021 parallel after T017 (contracts directory + files)
- T022-T025 sequential (quickstart)

**Phase 2** (Tests):
- T026-T030 (test infrastructure) - T029, T030 can run in parallel
- T031-T038 (basic CRUD test) - sequential
- T039-T045 (drift test) - sequential
- T046-T050 (worker nodes test) - sequential
- T051-T053 (validation tests) - T051, T052 can run in parallel
- T054-T055 (verification) - must be last

**Phase 3** (Implementation):
- T056-T061 (structure) - T060, T061 can run in parallel
- T062-T072 (schema) - mostly sequential, T071-T072 can run in parallel
- T073-T080 (Create) - sequential
- T081-T090 (Read) - sequential
- T091-T096 (Update) - sequential
- T097-T100 (Delete) - sequential
- T101-T102 (Import) - sequential
- T103-T108 (helpers) - sequential
- T109-T114 (verification) - sequential

**Phase 4**:
- T115-T116 (registration) - sequential
- T117-T122 (examples) - sequential
- T123-T128 (validation) - sequential

**Phase 5**:
- T129-T133 (docs) - sequential
- T134-T135 can run in parallel
- T136-T143 (quality) - sequential

### Parallel Opportunities

**Within Phase 0**:
```bash
# Parallel: Create both exploration scripts
T001: Create cmkube-get-clusters.py
T002: Create cmkube-crud-test.py
```

**Within Phase 1**:
```bash
# Parallel: Create all contract JSON files after directory created
T018: create-cluster.json
T019: read-cluster.json
T020: update-cluster.json
T021: delete-cluster.json
```

**Within Phase 2**:
```bash
# Parallel: Create test helper functions
T029: getTestMasterNodeUUID helper
T030: getTestWorkerNodeUUID helper

# Parallel: Create validation test functions
T051: ValidationInvalidName test
T052: ValidationInvalidVersion test
```

**Within Phase 3**:
```bash
# Parallel: Implement Metadata and Configure methods
T060: Metadata method
T061: Configure method

# Parallel: Add computed attributes to schema
T071: creation_time attribute
T072: revision_id attribute
```

**Within Phase 5**:
```bash
# Parallel: Run fmt and lint
T134: make fmt
T135: make lint
```

---

## Parallel Example: Phase 0 API Exploration

```bash
# Launch both exploration scripts in parallel:
Agent 1: "Create cmkube-get-clusters.py to list clusters"
Agent 2: "Create cmkube-crud-test.py to test CRUD lifecycle"

# Then execute sequentially:
Agent 1: "Execute cmkube-get-clusters.py and document findings"
Agent 2: "Execute cmkube-crud-test.py and document findings"
```

---

## Implementation Strategy

### Autonomous Execution Path (Recommended)

This task list is designed for `/speckit.implement` autonomous execution:

1. **Phase 0**: API exploration answers all unknowns → No user input needed
2. **Phase 1**: Design artifacts generated from Phase 0 → No user input needed
3. **Phase 2**: Tests written from design artifacts → No user input needed
4. **Phase 3**: Implementation follows test failures → No user input needed
5. **Phase 4**: Integration uses standard patterns → No user input needed
6. **Phase 5**: Documentation auto-generated → No user input needed

**Total Estimated Time**: 8-12 hours autonomous execution

### TDD Validation Points

1. **After T054-T055**: Confirm RED - All tests fail with "resource not found" errors
2. **After T109-T114**: Confirm GREEN - All tests pass with minimal implementation
3. **After T123-T128**: Confirm integration - Full test suite passes in real environment

### Success Metrics

- Phase 0: All 21 research questions (RQ-001 to RQ-021) answered ✓
- Phase 2: All tests fail (RED phase) ✓
- Phase 3: All tests pass (GREEN phase) ✓
- Phase 5: All 12 success criteria (SC-001 to SC-012) met ✓

---

## Notes

- **[P]** = Parallel-safe (different files, no dependencies between tasks)
- **Absolute paths** used throughout for autonomous execution
- **No user input** required - all decisions documented in spec.md and plan.md
- **TDD enforced** - tests written and failing before implementation
- **Modern patterns** - Uses terraform-plugin-testing v1.13.3+ (statecheck, plancheck, knownvalue)
- **Environment portable** - No hardcoded UUIDs, names, or values
- **Follows existing patterns** - Matches resource_cmpart_softwareimage.go and resource_cmdevice_category.go

---

## Risk Mitigation

| Risk | Mitigation | Tasks |
|------|------------|-------|
| BCM API differs from assumptions | Phase 0 validates all API contracts before coding | T001-T012 |
| Async cluster provisioning | Phase 0 documents polling requirements | T009 |
| Test environment constraints | Dynamic node/network queries, no hardcoded values | T029-T030 |
| Breaking existing provider | Tests verify integration, make build validates | T115-T116, T128 |
| Documentation drift | Auto-generation via make generate | T129-T133 |

---

**Status**: READY FOR AUTONOMOUS EXECUTION
**Next Action**: Begin Phase 0 - Execute T001 (Create cmkube-get-clusters.py)
