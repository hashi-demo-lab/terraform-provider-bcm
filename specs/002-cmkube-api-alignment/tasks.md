# Tasks: CMKube API Alignment

**Input**: Design documents from `/specs/002-cmkube-api-alignment/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: TDD approach requested - write failing tests first (RED), implement minimal code to pass (GREEN), refactor while tests pass (REFACTOR).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Provider code**: `internal/provider/`
- **Tests**: `internal/provider/*_test.go`
- **Examples**: `examples/resources/bcm_*/`
- **Generated docs**: `docs/` (via `make generate`)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and test infrastructure

- [ ] T001 Create test helper functions for etcd cluster operations in internal/provider/test_helpers.go
- [ ] T002 [P] Add EtcdCluster model types to internal/provider/models.go
- [ ] T003 [P] Add KubeCluster aligned model types to internal/provider/models.go
- [ ] T004 [P] Add KubeletRole and EtcdHostRole model types to internal/provider/models.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**CRITICAL**: No user story work can begin until this phase is complete

- [ ] T005 Add cmetcd service CRUD methods to bcm_client.go (addEtcdCluster, getEtcdCluster, updateEtcdCluster, removeEtcdCluster) in internal/provider/bcm_client.go
- [ ] T006 [P] Add cmetcd validation method (validateEtcdCluster) to bcm_client.go in internal/provider/bcm_client.go
- [ ] T007 [P] Add cmkube validation update for lowercase service name in internal/provider/bcm_client.go
- [ ] T008 [P] Create entity builder helpers for KubeletRole and EtcdHostRole in internal/provider/role_builders.go

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 2 - Manage Etcd Clusters Independently (Priority: P1)

**Goal**: Create and manage EtcdCluster entities separately from KubeClusters

**Independent Test**: Can be fully tested by creating a `bcm_cmetcd_cluster` resource and verifying the entity exists in BCM with correct configuration.

**Why First**: KubeCluster requires an EtcdCluster reference. This must be implemented before User Story 1.

### RED Phase - Write Failing Tests for US2

> **TDD**: Write these tests FIRST, ensure they FAIL before implementation

- [ ] T009 [P] [US2] Write acceptance test TestAccCMEtcdCluster_basic for create/read in internal/provider/resource_cmetcd_cluster_test.go
- [ ] T010 [P] [US2] Write acceptance test TestAccCMEtcdCluster_update for update operations in internal/provider/resource_cmetcd_cluster_test.go
- [ ] T011 [P] [US2] Write acceptance test TestAccCMEtcdCluster_import for import functionality in internal/provider/resource_cmetcd_cluster_test.go
- [ ] T012 [P] [US2] Write unit test TestCMEtcdClusterEntityBuilder for entity construction in internal/provider/resource_cmetcd_cluster_test.go
- [ ] T013 [P] [US2] Write CheckDestroy function testAccCheckCMEtcdClusterDestroy in internal/provider/resource_cmetcd_cluster_test.go

### GREEN Phase - Implement to Pass Tests for US2

- [ ] T014 [US2] Create resource schema for bcm_cmetcd_cluster in internal/provider/resource_cmetcd_cluster.go
- [ ] T015 [US2] Implement Create method for bcm_cmetcd_cluster in internal/provider/resource_cmetcd_cluster.go
- [ ] T016 [US2] Implement Read method for bcm_cmetcd_cluster in internal/provider/resource_cmetcd_cluster.go
- [ ] T017 [US2] Implement Update method for bcm_cmetcd_cluster in internal/provider/resource_cmetcd_cluster.go
- [ ] T018 [US2] Implement Delete method for bcm_cmetcd_cluster in internal/provider/resource_cmetcd_cluster.go
- [ ] T019 [US2] Implement ImportState method for bcm_cmetcd_cluster in internal/provider/resource_cmetcd_cluster.go
- [ ] T020 [US2] Register bcm_cmetcd_cluster resource in provider.go Resources() in internal/provider/provider.go
- [ ] T021 [US2] Run tests to verify GREEN status for all US2 tests

### REFACTOR Phase for US2

- [ ] T022 [P] [US2] Create example configuration in examples/resources/bcm_cmetcd_cluster/resource.tf
- [ ] T023 [P] [US2] Add validation for heartbeat_interval and election_timeout relationship in internal/provider/resource_cmetcd_cluster.go

**Checkpoint**: User Story 2 complete - EtcdCluster can be managed independently

---

## Phase 4: User Story 1 - Create Kubernetes Cluster with Correct BCM Entity Mapping (Priority: P1)

**Goal**: Create a Kubernetes cluster definition in BCM using Terraform with correct field mapping

**Independent Test**: Can be fully tested by creating a `bcm_cmkube_cluster` resource and verifying via BCM API that all configured fields are persisted correctly.

**Dependency**: Requires US2 (EtcdCluster) to be complete

### RED Phase - Write Failing Tests for US1

> **TDD**: Write these tests FIRST, ensure they FAIL before implementation

- [ ] T024 [P] [US1] Write acceptance test TestAccCMKubeCluster_aligned_basic for create/read with new schema in internal/provider/resource_cmkube_cluster_test.go
- [ ] T025 [P] [US1] Write acceptance test TestAccCMKubeCluster_aligned_update for update operations in internal/provider/resource_cmkube_cluster_test.go
- [ ] T026 [P] [US1] Write acceptance test TestAccCMKubeCluster_aligned_networks for network UUID persistence in internal/provider/resource_cmkube_cluster_test.go
- [ ] T027 [P] [US1] Write acceptance test TestAccCMKubeCluster_aligned_appGroups for app_groups nested block in internal/provider/resource_cmkube_cluster_test.go
- [ ] T028 [P] [US1] Write drift detection test TestAccCMKubeCluster_aligned_drift in internal/provider/resource_cmkube_cluster_test.go
- [ ] T029 [P] [US1] Write unit test TestCMKubeClusterEntityBuilder_aligned for entity construction in internal/provider/resource_cmkube_cluster_test.go

### GREEN Phase - Implement to Pass Tests for US1

- [ ] T030 [US1] Refactor bcm_cmkube_cluster schema to align with BCM API in internal/provider/resource_cmkube_cluster.go
- [ ] T031 [US1] Remove deprecated fields (master_nodes, worker_nodes, etcd_nodes, dns_servers, cni_plugin) in internal/provider/resource_cmkube_cluster.go
- [ ] T032 [US1] Add new required fields (internal_network, service_network, pod_network, etcd_cluster) in internal/provider/resource_cmkube_cluster.go
- [ ] T033 [US1] Add app_groups nested block schema with KubeAppGroup and KubeApp structure in internal/provider/resource_cmkube_cluster.go
- [ ] T034 [US1] Update Create method to build aligned BCM entity in internal/provider/resource_cmkube_cluster.go
- [ ] T035 [US1] Update Read method to parse aligned BCM entity response in internal/provider/resource_cmkube_cluster.go
- [ ] T036 [US1] Update Update method to handle aligned entity structure in internal/provider/resource_cmkube_cluster.go
- [ ] T037 [US1] Add validation using validateKubeCluster (cmkube service, lowercase) in internal/provider/resource_cmkube_cluster.go
- [ ] T038 [US1] Run tests to verify GREEN status for all US1 tests

### REFACTOR Phase for US1

- [ ] T039 [P] [US1] Update example configuration in examples/resources/bcm_cmkube_cluster/resource.tf
- [ ] T040 [P] [US1] Add UUID validation for network and etcd_cluster references in internal/provider/resource_cmkube_cluster.go

**Checkpoint**: User Story 1 complete - KubeCluster with correct BCM entity mapping

---

## Phase 5: User Story 3 - Assign Kubernetes Roles to Devices (Priority: P1)

**Goal**: Assign KubeletRole and EtcdHostRole to devices for cluster membership

**Independent Test**: Can be fully tested by adding a role to the `bcm_cmdevice_device` resource and verifying the device has the role in BCM.

**Dependency**: Requires US1 and US2 (clusters must exist to reference)

### RED Phase - Write Failing Tests for US3

> **TDD**: Write these tests FIRST, ensure they FAIL before implementation

- [ ] T041 [P] [US3] Write acceptance test TestAccCMDeviceDevice_kubeletRole for kubelet_role block in internal/provider/resource_cmdevice_device_test.go
- [ ] T042 [P] [US3] Write acceptance test TestAccCMDeviceDevice_etcdHostRole for etcd_host_role block in internal/provider/resource_cmdevice_device_test.go
- [ ] T043 [P] [US3] Write acceptance test TestAccCMDeviceDevice_bothRoles for combined kubelet + etcd roles in internal/provider/resource_cmdevice_device_test.go
- [ ] T044 [P] [US3] Write acceptance test TestAccCMDeviceDevice_roleUpdate for role modification in internal/provider/resource_cmdevice_device_test.go
- [ ] T045 [P] [US3] Write unit test TestKubeletRoleEntityBuilder for role construction in internal/provider/resource_cmdevice_device_roles_test.go
- [ ] T046 [P] [US3] Write unit test TestEtcdHostRoleEntityBuilder for role construction in internal/provider/resource_cmdevice_device_roles_test.go
- [ ] T047 [P] [US3] Write unit test TestDeviceRoleMerging for role array handling in internal/provider/resource_cmdevice_device_roles_test.go

### GREEN Phase - Implement to Pass Tests for US3

- [ ] T048 [US3] Add kubelet_role nested block schema to bcm_cmdevice_device in internal/provider/resource_cmdevice_device.go
- [ ] T049 [US3] Add etcd_host_role nested block schema to bcm_cmdevice_device in internal/provider/resource_cmdevice_device.go
- [ ] T050 [US3] Create KubeletRole entity builder in internal/provider/resource_cmdevice_device_roles.go
- [ ] T051 [US3] Create EtcdHostRole entity builder in internal/provider/resource_cmdevice_device_roles.go
- [ ] T052 [US3] Implement role merging logic (preserve non-Kubernetes roles, replace managed roles) in internal/provider/resource_cmdevice_device_roles.go
- [ ] T053 [US3] Update device Create to include Kubernetes roles in roles array in internal/provider/resource_cmdevice_device.go
- [ ] T054 [US3] Update device Read to parse Kubernetes roles from response in internal/provider/resource_cmdevice_device.go
- [ ] T055 [US3] Update device Update to handle role array replacement in internal/provider/resource_cmdevice_device.go
- [ ] T056 [US3] Run tests to verify GREEN status for all US3 tests

### REFACTOR Phase for US3

- [ ] T057 [P] [US3] Update device example with role blocks in examples/resources/bcm_cmdevice_device/resource.tf
- [ ] T058 [P] [US3] Add UUID preservation for existing roles during updates in internal/provider/resource_cmdevice_device_roles.go

**Checkpoint**: User Story 3 complete - Devices can be assigned to Kubernetes/Etcd clusters via roles

---

## Phase 6: User Story 4 - Import Existing Kubernetes Infrastructure (Priority: P2)

**Goal**: Import existing BCM Kubernetes infrastructure into Terraform

**Independent Test**: Can be fully tested by importing an existing KubeCluster or EtcdCluster by UUID and verifying state accuracy.

### RED Phase - Write Failing Tests for US4

> **TDD**: Write these tests FIRST, ensure they FAIL before implementation

- [ ] T059 [P] [US4] Write acceptance test TestAccCMKubeCluster_import_existing for import and plan verification in internal/provider/resource_cmkube_cluster_test.go
- [ ] T060 [P] [US4] Write acceptance test TestAccCMEtcdCluster_import_existing for import and plan verification in internal/provider/resource_cmetcd_cluster_test.go
- [ ] T061 [P] [US4] Write acceptance test TestAccCMDeviceDevice_import_withRoles for device import with Kubernetes roles in internal/provider/resource_cmdevice_device_test.go

### GREEN Phase - Implement to Pass Tests for US4

- [ ] T062 [US4] Verify ImportState preserves all fields for bcm_cmkube_cluster in internal/provider/resource_cmkube_cluster.go
- [ ] T063 [US4] Verify ImportState preserves all fields for bcm_cmetcd_cluster in internal/provider/resource_cmetcd_cluster.go
- [ ] T064 [US4] Update device ImportState to parse Kubernetes roles from BCM response in internal/provider/resource_cmdevice_device.go
- [ ] T065 [US4] Run tests to verify GREEN status for all US4 tests

### REFACTOR Phase for US4

- [ ] T066 [P] [US4] Add import examples to documentation in examples/resources/bcm_cmkube_cluster/import.sh
- [ ] T067 [P] [US4] Add import examples for etcd cluster in examples/resources/bcm_cmetcd_cluster/import.sh

**Checkpoint**: User Story 4 complete - Existing infrastructure can be imported

---

## Phase 7: User Story 5 - Configure Kubernetes API Server Settings (Priority: P2)

**Goal**: Configure Kubernetes API server settings for the cluster

**Independent Test**: Can be fully tested by configuring API server settings and verifying they are persisted in BCM.

### RED Phase - Write Failing Tests for US5

> **TDD**: Write these tests FIRST, ensure they FAIL before implementation

- [ ] T068 [P] [US5] Write acceptance test TestAccCMKubeCluster_apiServer for API server URL and proxy port in internal/provider/resource_cmkube_cluster_test.go
- [ ] T069 [P] [US5] Write acceptance test TestAccCMKubeCluster_trustedDomains for certificate SANs in internal/provider/resource_cmkube_cluster_test.go

### GREEN Phase - Implement to Pass Tests for US5

- [ ] T070 [US5] Ensure kubernetes_api_server attribute is properly mapped in schema in internal/provider/resource_cmkube_cluster.go
- [ ] T071 [US5] Ensure kubernetes_api_server_proxy_port has default value 6444 in internal/provider/resource_cmkube_cluster.go
- [ ] T072 [US5] Ensure trusted_domains list attribute is properly handled in internal/provider/resource_cmkube_cluster.go
- [ ] T073 [US5] Run tests to verify GREEN status for all US5 tests

**Checkpoint**: User Story 5 complete - API server settings configurable

---

## Phase 8: User Story 6 - Configure Ingress Proxy Settings (Priority: P3)

**Goal**: Configure ingress proxy settings for external traffic routing

**Independent Test**: Can be fully tested by enabling ingress proxy and verifying settings in BCM.

### RED Phase - Write Failing Tests for US6

> **TDD**: Write these tests FIRST, ensure they FAIL before implementation

- [ ] T074 [P] [US6] Write acceptance test TestAccCMKubeCluster_ingressProxy for ingress configuration in internal/provider/resource_cmkube_cluster_test.go

### GREEN Phase - Implement to Pass Tests for US6

- [ ] T075 [US6] Ensure ingress_proxy_enable attribute defaults to false in internal/provider/resource_cmkube_cluster.go
- [ ] T076 [US6] Ensure ingress_proxy_listen_port and ingress_proxy_backend_port are mapped correctly in internal/provider/resource_cmkube_cluster.go
- [ ] T077 [US6] Run tests to verify GREEN status for all US6 tests

**Checkpoint**: User Story 6 complete - Ingress proxy configurable

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T078 [P] Update data source bcm_cmkube_clusters model to align with new schema in internal/provider/data_source_cmkube_clusters.go
- [ ] T079 [P] Create data source bcm_cmetcd_clusters for listing etcd clusters in internal/provider/data_source_cmetcd_clusters.go
- [ ] T080 [P] Write test for bcm_cmetcd_clusters data source in internal/provider/data_source_cmetcd_clusters_test.go
- [ ] T081 Register bcm_cmetcd_clusters data source in provider.go DataSources() in internal/provider/provider.go
- [ ] T082 Run make generate to update documentation in docs/
- [ ] T083 Run make lint to verify code quality
- [ ] T084 Run full test suite (make test && TF_ACC=1 go test ./internal/provider/ -v)
- [ ] T085 Validate quickstart.md examples work end-to-end

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Story 2 (Phase 3)**: Must complete before US1 (KubeCluster references EtcdCluster)
- **User Story 1 (Phase 4)**: Depends on US2 completion
- **User Story 3 (Phase 5)**: Depends on US1 + US2 (needs clusters to reference)
- **User Stories 4-6 (Phases 6-8)**: Depend on core functionality (US1-3)
- **Polish (Phase 9)**: Depends on all user stories being complete

### User Story Dependencies

```
Setup → Foundational → US2 (EtcdCluster) → US1 (KubeCluster) → US3 (Device Roles)
                                                              ↓
                                               US4 (Import) → US5 (API Server) → US6 (Ingress)
```

### Within Each User Story (TDD Cycle)

1. **RED**: Write failing tests FIRST
2. **GREEN**: Implement minimal code to pass
3. **REFACTOR**: Clean up while keeping tests green
4. Tests marked [P] within same phase can run in parallel
5. GREEN phase depends on RED phase completion
6. REFACTOR phase depends on GREEN phase completion

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel
- Within each user story: All RED phase tests [P] can run in parallel
- Within each user story: All REFACTOR tasks [P] can run in parallel
- Different team members could work on different user stories after dependencies are met

---

## Parallel Example: User Story 2 (EtcdCluster)

```bash
# RED Phase - Launch all tests in parallel:
Task: "Write acceptance test TestAccCMEtcdCluster_basic in resource_cmetcd_cluster_test.go"
Task: "Write acceptance test TestAccCMEtcdCluster_update in resource_cmetcd_cluster_test.go"
Task: "Write acceptance test TestAccCMEtcdCluster_import in resource_cmetcd_cluster_test.go"
Task: "Write unit test TestCMEtcdClusterEntityBuilder in resource_cmetcd_cluster_test.go"
Task: "Write CheckDestroy function in resource_cmetcd_cluster_test.go"

# GREEN Phase - Sequential (depends on all RED tests existing):
Task: "Create resource schema for bcm_cmetcd_cluster"
Task: "Implement Create method"
Task: "Implement Read method"
# ... etc

# REFACTOR Phase - Launch cleanup tasks in parallel:
Task: "Create example configuration"
Task: "Add validation for heartbeat/election timeout"
```

---

## Implementation Strategy

### MVP First (User Stories 2, 1, 3 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 2 (EtcdCluster) - **TDD: RED → GREEN → REFACTOR**
4. Complete Phase 4: User Story 1 (KubeCluster alignment) - **TDD: RED → GREEN → REFACTOR**
5. Complete Phase 5: User Story 3 (Device roles) - **TDD: RED → GREEN → REFACTOR**
6. **STOP and VALIDATE**: Test complete cluster topology creation
7. Deploy/demo if ready - this is the MVP!

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add US2 (EtcdCluster) → Test independently → usable for etcd management
3. Add US1 (KubeCluster) → Test independently → cluster definitions work
4. Add US3 (Device roles) → Test independently → **MVP complete!**
5. Add US4 (Import) → Migrate existing infrastructure
6. Add US5 (API Server) → Production configuration
7. Add US6 (Ingress) → Full feature set

### TDD Discipline

For each user story:
1. **RED**: Run tests - they should FAIL (no implementation yet)
2. **GREEN**: Write MINIMAL code to pass - tests should PASS
3. **REFACTOR**: Improve code quality - tests must STAY GREEN
4. Commit after each TDD cycle phase

---

## Task Summary

| Phase | Description | Task Count | Parallel Tasks |
|-------|-------------|------------|----------------|
| 1 | Setup | 4 | 3 |
| 2 | Foundational | 4 | 3 |
| 3 | US2 - EtcdCluster | 15 | 7 |
| 4 | US1 - KubeCluster | 17 | 8 |
| 5 | US3 - Device Roles | 18 | 9 |
| 6 | US4 - Import | 9 | 5 |
| 7 | US5 - API Server | 6 | 2 |
| 8 | US6 - Ingress | 4 | 1 |
| 9 | Polish | 8 | 4 |
| **Total** | | **85** | **42** |

### Per User Story

| User Story | Priority | Task Count | Description |
|------------|----------|------------|-------------|
| US2 | P1 | 15 | Manage Etcd Clusters Independently |
| US1 | P1 | 17 | Create Kubernetes Cluster with Correct BCM Entity Mapping |
| US3 | P1 | 18 | Assign Kubernetes Roles to Devices |
| US4 | P2 | 9 | Import Existing Kubernetes Infrastructure |
| US5 | P2 | 6 | Configure Kubernetes API Server Settings |
| US6 | P3 | 4 | Configure Ingress Proxy Settings |

### MVP Scope

**MVP = US2 + US1 + US3** (50 tasks total, ~30 parallelizable)

This delivers:
- EtcdCluster management (bcm_cmetcd_cluster resource)
- KubeCluster management aligned with BCM API (bcm_cmkube_cluster refactored)
- Device role assignment for cluster membership (kubelet_role, etcd_host_role blocks)

---

## Notes

- [P] tasks = different files, no dependencies within the same phase
- [Story] label maps task to specific user story for traceability
- TDD: RED → GREEN → REFACTOR cycle for each user story
- Service names: `cmetcd` and `cmkube` are **lowercase** (exception to CamelCase pattern)
- BCM cmkube API requires **client-generated UUIDs** for new clusters
- Roles are **embedded** in Device entity, not managed as separate API calls
- Commit after each TDD phase (RED tests written, GREEN tests passing, REFACTOR complete)
