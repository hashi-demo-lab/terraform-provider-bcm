# Tasks: BCM Device Roles Block

**Feature Branch**: `039-device-roles-block`
**Input**: Design documents from `/specs/039-device-roles-block/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Tests**: This feature follows TDD approach. Acceptance tests are written first and must fail before implementation.

**Organization**: Tasks grouped by user story for independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US5)
- Include exact file paths in descriptions

## File Paths

- **Resource**: `internal/provider/resource_cmdevice_device.go`
- **Tests**: `internal/provider/resource_cmdevice_device_test.go`
- **Example**: `examples/resources/bcm_cmdevice_device/resource_with_roles.tf`

---

## Phase 1: Setup

**Purpose**: Prepare development environment and understand existing implementation

- [ ] T001 Review existing device resource implementation in internal/provider/resource_cmdevice_device.go
- [ ] T002 [P] Review existing device tests in internal/provider/resource_cmdevice_device_test.go
- [ ] T003 [P] Review roles data source for parsing patterns in internal/provider/data_source_cmdevice_roles.go
- [ ] T004 Create feature branch `039-device-roles-block` from main

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before user story implementation

**CRITICAL**: These changes enable all role-related functionality

- [ ] T005 Add Roles field to CMDeviceDeviceResourceModel struct in internal/provider/resource_cmdevice_device.go
- [ ] T006 Add roles schema attribute definition to Schema() method in internal/provider/resource_cmdevice_device.go
- [ ] T007 Add roles array building logic to buildDeviceAPIEntityWithExisting() in internal/provider/resource_cmdevice_device.go
- [ ] T008 Add roles parsing logic to parseDeviceFromAPI() in internal/provider/resource_cmdevice_device.go
- [ ] T009 Add sort import if not present for role name sorting in internal/provider/resource_cmdevice_device.go

**Checkpoint**: Foundation ready - acceptance tests can now be written and will fail appropriately

---

## Phase 3: User Story 1 - Assign Kubernetes Roles to Control Plane Nodes (Priority: P1) MVP

**Goal**: Infrastructure engineers can assign control-plane, master, and etcd roles to designated nodes

**Independent Test**: Create device with `roles = ["control-plane", "master", "etcd"]` and verify BCM assigns roles correctly

### Acceptance Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation passes**

- [ ] T010 [US1] Write TestAccCMDeviceDevice_RolesCreate test in internal/provider/resource_cmdevice_device_test.go
- [ ] T011 [US1] Write TestAccCMDeviceDevice_RolesMultiple test for multiple roles in internal/provider/resource_cmdevice_device_test.go
- [ ] T012 [US1] Write testAccCMDeviceDeviceConfigWithRoles helper function in internal/provider/resource_cmdevice_device_test.go

### Verification for User Story 1

- [ ] T013 [US1] Run TestAccCMDeviceDevice_RolesCreate to verify device created with roles
- [ ] T014 [US1] Run TestAccCMDeviceDevice_RolesMultiple to verify multiple role assignment

**Checkpoint**: User Story 1 complete - devices can be created with Kubernetes control plane roles

---

## Phase 4: User Story 2 - Assign Worker Roles to Compute Nodes (Priority: P1)

**Goal**: Infrastructure engineers can designate nodes as Kubernetes workers

**Independent Test**: Create device with `roles = ["worker"]` and verify BCM configures node as worker

### Acceptance Tests for User Story 2

> **NOTE: Tests should pass with existing implementation from Phase 2/3**

- [ ] T015 [US2] Write TestAccCMDeviceDevice_RolesIdempotent test in internal/provider/resource_cmdevice_device_test.go

### Verification for User Story 2

- [ ] T016 [US2] Run TestAccCMDeviceDevice_RolesIdempotent to verify idempotent role assignment

**Checkpoint**: User Story 2 complete - worker nodes can be configured independently

---

## Phase 5: User Story 3 - Update Device Roles (Priority: P2)

**Goal**: Infrastructure engineers can modify role assignments during cluster lifecycle

**Independent Test**: Create device with one role set, update to different role set, verify change

### Acceptance Tests for User Story 3

- [ ] T017 [US3] Write TestAccCMDeviceDevice_RolesUpdate test in internal/provider/resource_cmdevice_device_test.go
- [ ] T018 [US3] Write TestAccCMDeviceDevice_RolesRemove test for removing all roles in internal/provider/resource_cmdevice_device_test.go

### Verification for User Story 3

- [ ] T019 [US3] Run TestAccCMDeviceDevice_RolesUpdate to verify role updates work
- [ ] T020 [US3] Run TestAccCMDeviceDevice_RolesRemove to verify roles can be removed

**Checkpoint**: User Story 3 complete - role updates and removal work correctly

---

## Phase 6: User Story 4 - Import Device with Existing Roles (Priority: P2)

**Goal**: Engineers migrating to Terraform can import existing devices with roles preserved

**Independent Test**: Import existing device with roles and verify roles appear in state

### Acceptance Tests for User Story 4

- [ ] T021 [US4] Write TestAccCMDeviceDevice_RolesImport test in internal/provider/resource_cmdevice_device_test.go

### Verification for User Story 4

- [ ] T022 [US4] Run TestAccCMDeviceDevice_RolesImport to verify import preserves roles

**Checkpoint**: User Story 4 complete - existing infrastructure can be imported

---

## Phase 7: User Story 5 - Drift Detection for Role Changes (Priority: P3)

**Goal**: Terraform detects when roles are modified outside of Terraform

**Independent Test**: Modify roles via BCM API, run terraform plan, verify drift detected

### Acceptance Tests for User Story 5

- [ ] T023 [US5] Write TestAccCMDeviceDevice_RolesDrift test in internal/provider/resource_cmdevice_device_test.go

### Verification for User Story 5

- [ ] T024 [US5] Run TestAccCMDeviceDevice_RolesDrift to verify drift detection works

**Checkpoint**: User Story 5 complete - drift detection ensures configuration consistency

---

## Phase 8: Polish and Cross-Cutting Concerns

**Purpose**: Documentation, examples, and final validation

- [ ] T025 [P] Create example configuration in examples/resources/bcm_cmdevice_device/resource_with_roles.tf
- [ ] T026 [P] Run make generate to update documentation in docs/
- [ ] T027 Run make lint to verify code quality
- [ ] T028 Run make test to verify unit tests pass
- [ ] T029 Run full acceptance test suite with TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "CMDeviceDevice.*Roles"
- [ ] T030 Validate quickstart.md scenarios work as documented

---

## Dependencies and Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories
- **User Stories (Phases 3-7)**: All depend on Foundational completion
  - US1 and US2 can run in parallel (both P1 priority)
  - US3 and US4 can run in parallel (both P2 priority)
  - US5 runs after US3/US4 (P3 priority)
- **Polish (Phase 8)**: Depends on all user stories being complete

### Within Each User Story (TDD Pattern)

1. Write acceptance test FIRST
2. Run test - verify it FAILS (RED)
3. Implementation already done in Phase 2 (GREEN)
4. Run test - verify it PASSES
5. Refactor if needed

### Parallel Opportunities

**Phase 1 (Setup)**:
```bash
# Parallel review tasks
Task: T002 "Review existing device tests"
Task: T003 "Review roles data source for parsing patterns"
```

**Phases 3-4 (US1 + US2 - both P1)**:
```bash
# Can work on both user stories in parallel after Phase 2
Task: T010-T014 "User Story 1 tests and verification"
Task: T015-T016 "User Story 2 tests and verification"
```

**Phases 5-6 (US3 + US4 - both P2)**:
```bash
# Can work on both user stories in parallel
Task: T017-T020 "User Story 3 tests and verification"
Task: T021-T022 "User Story 4 tests and verification"
```

**Phase 8 (Polish)**:
```bash
# Parallel documentation tasks
Task: T025 "Create example configuration"
Task: T026 "Run make generate"
```

---

## Implementation Strategy

### MVP First (User Stories 1-2)

1. Complete Phase 1: Setup (review existing code)
2. Complete Phase 2: Foundational (add roles support)
3. Complete Phase 3: User Story 1 (control plane roles)
4. Complete Phase 4: User Story 2 (worker roles)
5. **STOP and VALIDATE**: Test Kubernetes topology definition
6. Deploy/demo if ready for immediate Kubernetes use cases

### Incremental Delivery

1. **Foundation Ready**: Setup + Foundational complete
2. **MVP (P1)**: Add US1 + US2 -> Test -> Deploy (Kubernetes topology works!)
3. **Enhanced (P2)**: Add US3 + US4 -> Test -> Deploy (Updates + Import work!)
4. **Complete (P3)**: Add US5 -> Test -> Deploy (Full drift detection!)
5. **Polish**: Documentation and final validation

### Test Execution Summary

| Test Name | User Story | Priority | Validates |
|-----------|------------|----------|-----------|
| TestAccCMDeviceDevice_RolesCreate | US1 | P1 | Create with roles |
| TestAccCMDeviceDevice_RolesMultiple | US1 | P1 | Multiple roles |
| TestAccCMDeviceDevice_RolesIdempotent | US2 | P1 | Idempotency |
| TestAccCMDeviceDevice_RolesUpdate | US3 | P2 | Role updates |
| TestAccCMDeviceDevice_RolesRemove | US3 | P2 | Role removal |
| TestAccCMDeviceDevice_RolesImport | US4 | P2 | Import with roles |
| TestAccCMDeviceDevice_RolesDrift | US5 | P3 | Drift detection |

---

## Notes

- **TDD**: All acceptance tests must be written before verifying implementation
- **7 tests total** as defined in plan.md (aligned with spec.md Testing Strategy)
- **Existing patterns**: Follow interfaces block implementation for consistency
- **BCM validation**: Rely on server-side validation for role names
- **Sorting**: Role names sorted alphabetically in state for consistent comparison
- Commit after each task or logical group
- Stop at any checkpoint to validate user story independently
