# Tasks: Fix Device Role Association Bug

**Input**: Design documents from `/specs/086-fix-device-role-association/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/role-validation.md
**Branch**: `086-fix-device-role-association`

**Tests**: This is a TDD project. Acceptance tests are REQUIRED per CLAUDE.md and the feature specification.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

- **Terraform Provider**: `internal/provider/` for source, `examples/` for documentation
- **Tests**: In same directory as implementation (`*_test.go` files)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Foundation changes that support all user stories

- [X] T001 Add `regexp` and `sort` imports to internal/provider/resource_cmdevice_device.go (if not present)
- [X] T002 Add `isUUID()` helper function for UUID format detection in internal/provider/resource_cmdevice_device.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core changes that MUST be complete before acceptance tests can pass

**CRITICAL**: No user story tests can pass until this phase is complete

- [X] T003 Modify `lookupAndBuildRolesForEntity()` to build both rolesByName and rolesByUUID maps in internal/provider/resource_cmdevice_device.go
- [X] T004 Modify `lookupAndBuildRolesForEntity()` to resolve identifiers by name or UUID based on format in internal/provider/resource_cmdevice_device.go
- [X] T005 Modify `lookupAndBuildRolesForEntity()` to return clear error messages with available roles list in internal/provider/resource_cmdevice_device.go
- [X] T006 Modify `parseRolesFromAPI()` to return role names instead of UUIDs in internal/provider/resource_cmdevice_device.go

**Checkpoint**: Foundation ready - acceptance tests can now be written and should pass after this phase

---

## Phase 3: User Story 1 - Device Role Assignment by Name (Priority: P1) MVP

**Goal**: Allow users to assign roles using human-readable names like `roles = ["backup", "provisioning"]`

**Independent Test**: Create a device with roles specified by name and verify roles are correctly associated in BCM

### Tests for User Story 1

> **NOTE: Write these tests FIRST (TDD RED phase), then verify implementation passes them**

- [X] T007 [P] [US1] Write acceptance test `TestAccCMDeviceDevice_RolesByName` in internal/provider/resource_cmdevice_device_test.go
- [X] T008 [P] [US1] Write test helper `testAccCMDeviceDeviceConfigWithRoles()` in internal/provider/resource_cmdevice_device_test.go
- [X] T009 [P] [US1] Write acceptance test `TestAccCMDeviceDevice_RolesImport` to verify imported devices show role names in internal/provider/resource_cmdevice_device_test.go
- [X] T010 [P] [US1] Write acceptance test `TestAccCMDeviceDevice_RolesUpdate` to verify role changes work correctly in internal/provider/resource_cmdevice_device_test.go
- [X] T011 [P] [US1] Write acceptance test `TestAccCMDeviceDevice_RolesDrift` to verify drift detection for externally changed roles in internal/provider/resource_cmdevice_device_test.go

### Verification for User Story 1

- [ ] T012 [US1] Run `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceDevice_RolesByName"` and verify PASS
- [ ] T013 [US1] Run `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceDevice_RolesImport"` and verify PASS
- [ ] T014 [US1] Run `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceDevice_RolesUpdate"` and verify PASS
- [ ] T015 [US1] Run `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceDevice_RolesDrift"` and verify PASS

**Checkpoint**: User Story 1 complete - devices can be created with roles by name

---

## Phase 4: User Story 2 - Client-Side Role Validation (Priority: P1)

**Goal**: Provide clear error messages when invalid role names are specified

**Independent Test**: Attempt to create a device with `roles = ["nonexistent-role"]` and verify clear error is returned

### Tests for User Story 2

- [X] T016 [P] [US2] Write acceptance test `TestAccCMDeviceDevice_InvalidRoleName` in internal/provider/resource_cmdevice_device_test.go
- [X] T017 [P] [US2] Write acceptance test `TestAccCMDeviceDevice_InvalidRoleUUID` in internal/provider/resource_cmdevice_device_test.go
- [X] T018 [P] [US2] Write acceptance test `TestAccCMDeviceDevice_EmptyRoleString` in internal/provider/resource_cmdevice_device_test.go
- [X] T019 [P] [US2] Write acceptance test `TestAccCMDeviceDevice_MultipleInvalidRoles` in internal/provider/resource_cmdevice_device_test.go

### Verification for User Story 2

- [ ] T020 [US2] Run `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceDevice_InvalidRoleName"` and verify PASS
- [ ] T021 [US2] Run `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceDevice_InvalidRoleUUID"` and verify PASS
- [ ] T022 [US2] Run `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceDevice_EmptyRoleString"` and verify PASS
- [ ] T023 [US2] Run `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceDevice_MultipleInvalidRoles"` and verify PASS

**Checkpoint**: User Story 2 complete - invalid roles produce clear error messages

---

## Phase 5: User Story 3 - Updated Example Documentation (Priority: P2)

**Goal**: Simplify example code to demonstrate role assignment by name

**Independent Test**: Run `terraform validate` and `terraform plan` on updated example

### Implementation for User Story 3

- [X] T024 [P] [US3] Update example file examples/resources/bcm_cmdevice_device/with_roles.tf to use role names directly
- [X] T025 [US3] Update `roles` attribute MarkdownDescription in Schema() method in internal/provider/resource_cmdevice_device.go
- [X] T026 [US3] Run `make generate` to regenerate documentation in docs/

### Verification for User Story 3

- [ ] T027 [US3] Run `terraform validate` on examples/resources/bcm_cmdevice_device/with_roles.tf and verify success
- [ ] T028 [US3] Verify docs/resources/cmdevice_device.md reflects updated roles documentation

**Checkpoint**: User Story 3 complete - example and documentation updated

---

## Phase 6: User Story 4 - Backward Compatibility with UUID Input (Priority: P3)

**Goal**: Ensure existing configurations using role UUIDs continue to work

**Independent Test**: Create a device with roles specified by UUID and verify it still works

### Tests for User Story 4

- [X] T029 [P] [US4] Write acceptance test `TestAccCMDeviceDevice_RolesByUUID` in internal/provider/resource_cmdevice_device_test.go
- [X] T030 [P] [US4] Write acceptance test `TestAccCMDeviceDevice_RolesMixedInput` for combined name and UUID input in internal/provider/resource_cmdevice_device_test.go

### Verification for User Story 4

- [ ] T031 [US4] Run `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceDevice_RolesByUUID"` and verify PASS
- [ ] T032 [US4] Run `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceDevice_RolesMixedInput"` and verify PASS

**Checkpoint**: User Story 4 complete - backward compatibility verified

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and quality checks

- [X] T033 Run `make lint` and fix any linting errors
- [X] T034 Run `pre-commit run --all-files` and ensure all checks pass
- [ ] T035 Run full acceptance test suite: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceDevice_Roles"`
- [ ] T036 Verify example file works end-to-end: `./scripts/test-examples.sh --resources-only`
- [ ] T037 Run quickstart.md verification checklist

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user story tests from passing
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion
  - US1 (P1) and US2 (P1): Same priority, can proceed in parallel
  - US3 (P2): Can proceed after US1/US2 or in parallel
  - US4 (P3): Can proceed after US1 or in parallel
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Depends on T001-T006 (Setup + Foundational)
- **User Story 2 (P1)**: Depends on T005 specifically (error message handling)
- **User Story 3 (P2)**: Depends on T006 (parseRolesFromAPI returns names)
- **User Story 4 (P3)**: Depends on T003-T004 (UUID lookup support)

### Within Each User Story

- Tests MUST be written and FAIL before implementation (TDD RED phase)
- Implementation in Foundational phase makes tests GREEN
- Verification confirms tests pass

### Parallel Opportunities

- T001, T002 (Setup) can run in parallel
- T003-T006 (Foundational) are sequential modifications to same function
- All test tasks within a phase marked [P] can run in parallel
- US3 (documentation) can proceed in parallel with US1/US2 testing
- US4 (backward compatibility) tests can proceed in parallel with other testing

---

## Parallel Example: User Story 1 Tests

```bash
# Launch all tests for User Story 1 together:
Task: "Write acceptance test TestAccCMDeviceDevice_RolesByName in internal/provider/resource_cmdevice_device_test.go"
Task: "Write test helper testAccCMDeviceDeviceConfigWithRoles() in internal/provider/resource_cmdevice_device_test.go"
Task: "Write acceptance test TestAccCMDeviceDevice_RolesImport in internal/provider/resource_cmdevice_device_test.go"
Task: "Write acceptance test TestAccCMDeviceDevice_RolesUpdate in internal/provider/resource_cmdevice_device_test.go"
Task: "Write acceptance test TestAccCMDeviceDevice_RolesDrift in internal/provider/resource_cmdevice_device_test.go"
```

---

## Implementation Strategy

### MVP First (User Stories 1 + 2 Only)

1. Complete Phase 1: Setup (T001-T002)
2. Complete Phase 2: Foundational (T003-T006)
3. Complete Phase 3: User Story 1 - Role assignment by name
4. Complete Phase 4: User Story 2 - Client-side validation
5. **STOP and VALIDATE**: Both P1 stories are complete and tested
6. Can merge/deploy as MVP

### Incremental Delivery

1. Setup + Foundational -> Core functionality ready
2. Add User Story 1 -> Test -> Role names work (MVP!)
3. Add User Story 2 -> Test -> Error messages clear
4. Add User Story 3 -> Test -> Documentation updated
5. Add User Story 4 -> Test -> Backward compatibility verified
6. Polish -> All quality checks pass

### TDD Workflow Per Story

1. **RED**: Write failing acceptance tests (T007-T011 for US1)
2. **GREEN**: Implementation is in Foundational phase (T003-T006)
3. **REFACTOR**: Polish phase handles cleanup (T033-T037)
4. **VERIFY**: Run tests to confirm GREEN (T012-T015 for US1)

---

## Files Modified Summary

| File | Changes |
|------|---------|
| `internal/provider/resource_cmdevice_device.go` | Add isUUID(), modify lookupAndBuildRolesForEntity(), modify parseRolesFromAPI(), update schema docs |
| `internal/provider/resource_cmdevice_device_test.go` | Add 11 new acceptance tests for role functionality |
| `examples/resources/bcm_cmdevice_device/with_roles.tf` | Simplify to use role names directly |
| `docs/resources/cmdevice_device.md` | Auto-generated via make generate |

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- TDD: Tests are written first (RED), implementation makes them pass (GREEN)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Success metrics from spec.md: <500ms role resolution, clear error messages, 100% backward compatibility
