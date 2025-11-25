# Tasks: BCM Device Roles Block

**Input**: Design documents from `/workspace/specs/039-device-roles-block/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Tests**: This feature requires 7 acceptance tests as defined in plan.md and spec.md. Tests are written FIRST following TDD (RED-GREEN-REFACTOR).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

## Path Conventions

- **Provider source**: `internal/provider/`
- **Tests**: `internal/provider/*_test.go`
- **Examples**: `examples/resources/bcm_cmdevice_device/`

---

## Phase 1: Setup

**Purpose**: Prepare the codebase for role assignment feature

- [ ] T001 Create feature branch `039-device-roles-block` from main
- [ ] T002 [P] Review existing device resource implementation in `/workspace/internal/provider/resource_cmdevice_device.go`
- [ ] T003 [P] Review role data source patterns in `/workspace/internal/provider/data_source_cmdevice_roles.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core schema and model changes that MUST be complete before ANY user story can be implemented

**CRITICAL**: No user story work can begin until this phase is complete

- [ ] T004 Add `Roles types.List` field to `CMDeviceDeviceResourceModel` struct in `/workspace/internal/provider/resource_cmdevice_device.go`
- [ ] T005 Add `roles` schema attribute definition in `Schema()` method in `/workspace/internal/provider/resource_cmdevice_device.go`
- [ ] T006 Implement roles array building in `buildDeviceAPIEntityWithExisting()` function in `/workspace/internal/provider/resource_cmdevice_device.go`
- [ ] T007 Implement roles parsing in `parseDeviceFromAPI()` function in `/workspace/internal/provider/resource_cmdevice_device.go`
- [ ] T008 Add `sort` import and role name sorting for consistent state comparison in `/workspace/internal/provider/resource_cmdevice_device.go`
- [ ] T009 Handle null vs empty list semantics for roles in Create/Update methods in `/workspace/internal/provider/resource_cmdevice_device.go`

**Checkpoint**: Schema and core logic ready - acceptance tests can now be written and run

---

## Phase 3: User Story 1 - Assign Kubernetes Roles to Control Plane Nodes (Priority: P1)

**Goal**: Enable infrastructure engineers to assign control-plane, master, and etcd roles to designated nodes for Kubernetes cluster topology definition.

**Independent Test**: Create a device with `roles = ["control-plane", "master", "etcd"]` and verify BCM assigns the roles correctly.

### Tests for User Story 1 (TDD - Write FIRST, must FAIL before implementation)

- [ ] T010 [P] [US1] Write acceptance test `TestAccCMDeviceDevice_RolesCreate` for creating device with multiple roles in `/workspace/internal/provider/resource_cmdevice_device_test.go`
- [ ] T011 [P] [US1] Write test config helper `testAccCMDeviceDeviceConfigWithRoles()` in `/workspace/internal/provider/resource_cmdevice_device_test.go`

### Implementation for User Story 1

- [ ] T012 [US1] Run `TestAccCMDeviceDevice_RolesCreate` to verify RED state (test fails)
- [ ] T013 [US1] Verify Create() correctly includes roles in device entity sent to BCM addDevice in `/workspace/internal/provider/resource_cmdevice_device.go`
- [ ] T014 [US1] Verify Read() correctly extracts roles from BCM getDevice response in `/workspace/internal/provider/resource_cmdevice_device.go`
- [ ] T015 [US1] Run `TestAccCMDeviceDevice_RolesCreate` to verify GREEN state (test passes)

**Checkpoint**: User Story 1 complete - can create devices with control-plane roles

---

## Phase 4: User Story 2 - Assign Worker Roles to Compute Nodes (Priority: P1)

**Goal**: Enable infrastructure engineers to designate nodes as Kubernetes workers by assigning the worker role.

**Independent Test**: Create a device with `roles = ["worker"]` and verify BCM configures the node as a Kubernetes worker.

### Tests for User Story 2

- [ ] T016 [P] [US2] Write acceptance test `TestAccCMDeviceDevice_RolesMultiple` for multiple independent worker nodes in `/workspace/internal/provider/resource_cmdevice_device_test.go`
- [ ] T017 [P] [US2] Write acceptance test `TestAccCMDeviceDevice_RolesIdempotent` for idempotency verification in `/workspace/internal/provider/resource_cmdevice_device_test.go`

### Implementation for User Story 2

- [ ] T018 [US2] Run `TestAccCMDeviceDevice_RolesMultiple` and `TestAccCMDeviceDevice_RolesIdempotent` to verify RED state
- [ ] T019 [US2] Implement role name deduplication logic before sending to BCM API in `/workspace/internal/provider/resource_cmdevice_device.go`
- [ ] T020 [US2] Run tests to verify GREEN state (tests pass)

**Checkpoint**: User Stories 1 AND 2 complete - can create control-plane and worker nodes with roles

---

## Phase 5: User Story 3 - Update Device Roles (Priority: P2)

**Goal**: Enable infrastructure engineers to modify role assignments during cluster lifecycle events such as promoting a worker to control-plane or removing roles during maintenance.

**Independent Test**: Create a device with one role set, then update to a different role set and verify the change.

### Tests for User Story 3

- [ ] T021 [P] [US3] Write acceptance test `TestAccCMDeviceDevice_RolesUpdate` for updating roles in `/workspace/internal/provider/resource_cmdevice_device_test.go`
- [ ] T022 [P] [US3] Write acceptance test `TestAccCMDeviceDevice_RolesRemove` for removing all roles in `/workspace/internal/provider/resource_cmdevice_device_test.go`

### Implementation for User Story 3

- [ ] T023 [US3] Run `TestAccCMDeviceDevice_RolesUpdate` and `TestAccCMDeviceDevice_RolesRemove` to verify RED state
- [ ] T024 [US3] Verify Update() correctly sends complete desired role state to BCM updateDevice in `/workspace/internal/provider/resource_cmdevice_device.go`
- [ ] T025 [US3] Handle empty roles list (`roles = []`) as explicit removal of all roles in `/workspace/internal/provider/resource_cmdevice_device.go`
- [ ] T026 [US3] Run tests to verify GREEN state (tests pass)

**Checkpoint**: User Story 3 complete - can update and remove roles from devices

---

## Phase 6: User Story 4 - Import Device with Existing Roles (Priority: P2)

**Goal**: Enable infrastructure engineers migrating to Terraform to import existing devices with their current role assignments preserved.

**Independent Test**: Import an existing device with roles and verify the roles appear in state.

### Tests for User Story 4

- [ ] T027 [US4] Write acceptance test `TestAccCMDeviceDevice_RolesImport` for import with roles in `/workspace/internal/provider/resource_cmdevice_device_test.go`

### Implementation for User Story 4

- [ ] T028 [US4] Run `TestAccCMDeviceDevice_RolesImport` to verify RED state
- [ ] T029 [US4] Verify ImportState with ImportStatePassthroughID handles roles correctly (existing implementation should work) in `/workspace/internal/provider/resource_cmdevice_device.go`
- [ ] T030 [US4] Verify Read() populates roles from BCM response during import in `/workspace/internal/provider/resource_cmdevice_device.go`
- [ ] T031 [US4] Run test to verify GREEN state (test passes)

**Checkpoint**: User Story 4 complete - can import devices with existing roles

---

## Phase 7: User Story 5 - Drift Detection for Role Changes (Priority: P3)

**Goal**: When roles are modified outside of Terraform (via BCM UI or API), Terraform should detect the drift and propose corrections.

**Independent Test**: Modify roles via BCM API, then run `terraform plan` to verify drift is detected.

### Tests for User Story 5

- [ ] T032 [US5] Write acceptance test `TestAccCMDeviceDevice_RolesDrift` for drift detection in `/workspace/internal/provider/resource_cmdevice_device_test.go`

### Implementation for User Story 5

- [ ] T033 [US5] Run `TestAccCMDeviceDevice_RolesDrift` to verify RED state
- [ ] T034 [US5] Verify drift detection works via standard Read() operation in `/workspace/internal/provider/resource_cmdevice_device.go`
- [ ] T035 [US5] Implement drift test helper to modify roles via BCM API directly in `/workspace/internal/provider/resource_cmdevice_device_test.go`
- [ ] T036 [US5] Use `plancheck.ExpectNonEmptyPlan()` to verify drift is detected in `/workspace/internal/provider/resource_cmdevice_device_test.go`
- [ ] T037 [US5] Run test to verify GREEN state (test passes)

**Checkpoint**: All user stories complete - full CRUD + Import + Drift detection working

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, examples, and final quality improvements

- [ ] T038 [P] Create example configuration in `/workspace/examples/resources/bcm_cmdevice_device/resource_with_roles.tf`
- [ ] T039 [P] Update MarkdownDescription for roles attribute with common role names
- [ ] T040 Run `make generate` to update provider documentation
- [ ] T041 Run `make lint` to verify code quality
- [ ] T042 Run full acceptance test suite for device resource: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceDevice`
- [ ] T043 Update `/workspace/specs/039-device-roles-block/quickstart.md` with final test results

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phases 3-7)**: All depend on Foundational phase completion
  - US1 and US2 are both P1 priority and can proceed in parallel
  - US3 and US4 are both P2 priority and can proceed after US1/US2 or in parallel with them
  - US5 is P3 priority and can proceed after earlier stories or in parallel
- **Polish (Phase 8)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P1)**: Can start after Foundational (Phase 2) - Independent of US1
- **User Story 3 (P2)**: Can start after Foundational (Phase 2) - Verifies Create/Update flow works together
- **User Story 4 (P2)**: Can start after Foundational (Phase 2) - Verifies Import works
- **User Story 5 (P3)**: Can start after Foundational (Phase 2) - Verifies Drift detection

### Within Each User Story

- Tests MUST be written and FAIL before implementation (TDD)
- Schema changes before entity building logic
- Entity building before response parsing
- Core implementation before edge cases
- Story complete before moving to next priority

### Parallel Opportunities

- T002, T003: Review existing code in parallel
- T010, T011: Write US1 tests in parallel
- T016, T017: Write US2 tests in parallel
- T021, T022: Write US3 tests in parallel
- T038, T039: Documentation tasks in parallel

---

## Parallel Example: User Story 1 + User Story 2 (Both P1)

```bash
# After Foundational phase completes, launch US1 and US2 tests together:

# Developer A: User Story 1
Task T010: "Write acceptance test TestAccCMDeviceDevice_RolesCreate"
Task T011: "Write test config helper testAccCMDeviceDeviceConfigWithRoles()"
# Then T012-T015 sequentially

# Developer B: User Story 2 (can start simultaneously)
Task T016: "Write acceptance test TestAccCMDeviceDevice_RolesMultiple"
Task T017: "Write acceptance test TestAccCMDeviceDevice_RolesIdempotent"
# Then T018-T020 sequentially
```

---

## Implementation Strategy

### MVP First (User Stories 1 + 2 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (control-plane roles)
4. Complete Phase 4: User Story 2 (worker roles)
5. **STOP and VALIDATE**: Test Kubernetes cluster topology definition
6. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational -> Foundation ready
2. Add User Story 1 -> Test independently -> Demo (control-plane nodes!)
3. Add User Story 2 -> Test independently -> Demo (full K8s topology!)
4. Add User Story 3 -> Test independently -> Demo (role updates!)
5. Add User Story 4 -> Test independently -> Demo (import capability!)
6. Add User Story 5 -> Test independently -> Demo (drift detection!)
7. Each story adds value without breaking previous stories

### Suggested MVP Scope

For immediate DGX deployment needs, complete:
- Phase 1: Setup
- Phase 2: Foundational
- Phase 3: User Story 1 (control-plane roles)
- Phase 4: User Story 2 (worker roles)

This delivers the core capability to define Kubernetes cluster topology in Terraform.

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing (TDD RED phase)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- All tests use modern terraform-plugin-testing v1.13.3+ patterns
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence

---

## Summary

| Metric | Value |
|--------|-------|
| Total Tasks | 43 |
| Setup Phase | 3 tasks |
| Foundational Phase | 6 tasks |
| User Story 1 (P1) | 6 tasks |
| User Story 2 (P1) | 5 tasks |
| User Story 3 (P2) | 6 tasks |
| User Story 4 (P2) | 5 tasks |
| User Story 5 (P3) | 6 tasks |
| Polish Phase | 6 tasks |
| Acceptance Tests | 7 (as specified in plan.md) |
| Parallel Opportunities | 12 tasks marked [P] |
| Suggested MVP | Phases 1-4 (20 tasks) |
