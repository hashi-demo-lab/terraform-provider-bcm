# Tasks: Category Test Coverage Enhancement

**Input**: Design documents from `/specs/001-category-test-coverage/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, quickstart.md
**GitHub Issue**: #66

**Two-Track Approach**:
- **Track A**: Test fields with existing implementation (install_mode, new_node_install_mode, kernel_version)
- **Track B**: Implement + test fields missing from buildAPIEntity/readCategory

**Organization**: Tasks are grouped by user story from spec.md to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Prepare test environment and verify existing patterns

- [ ] T001 Review existing test patterns in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T002 [P] Verify required imports for modern testing patterns are available (statecheck, plancheck, knownvalue, tfjsonpath, compare)
- [ ] T003 [P] Verify test helpers exist in /workspace/internal/provider/test_helpers.go (generateUniqueTestName, createTestBCMClient)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Ensure Track B fields have implementation support before tests can be written

**CRITICAL**: Track B tests (US2-US8) depend on resource implementation being complete

### Implementation for Track B Fields

- [X] T004 Implement name_servers list field in buildAPIEntity() in /workspace/internal/provider/resource_cmdevice_category.go
- [X] T005 [P] Implement search_domains list field in buildAPIEntity() in /workspace/internal/provider/resource_cmdevice_category.go
- [X] T006 [P] Implement time_servers list field in buildAPIEntity() in /workspace/internal/provider/resource_cmdevice_category.go
- [X] T007 Implement name_servers, search_domains, time_servers list parsing in readCategory() in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T008 Implement io_scheduler field in buildAPIEntity() and readCategory() in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T009 Implement initialize script field in buildAPIEntity() and readCategory() in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T010 [P] Implement finalize script field in buildAPIEntity() and readCategory() in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T011 Implement bmc_settings nested object in buildAPIEntity() and readCategory() in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T012 Implement modules nested list in buildAPIEntity() and readCategory() in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T013 Implement exclude_list_full field in buildAPIEntity() and readCategory() in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T014 [P] Implement exclude_list_grab field in buildAPIEntity() and readCategory() in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T015 [P] Implement exclude_list_sync field in buildAPIEntity() and readCategory() in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T016 [P] Implement exclude_list_update field in buildAPIEntity() and readCategory() in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T017 [P] Implement exclude_list_grabnew field in buildAPIEntity() and readCategory() in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T018 [P] Implement exclude_list_manipulate_script field in buildAPIEntity() and readCategory() in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T019 Implement fips field in buildAPIEntity() and readCategory() in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T020 [P] Implement data_node field in buildAPIEntity() and readCategory() in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T021 [P] Implement interactive_user field in buildAPIEntity() and readCategory() in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T022 [P] Implement use_exclusively_for field in buildAPIEntity() and readCategory() in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T023 [P] Implement node_installer_disk field in buildAPIEntity() and readCategory() in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T024 [P] Implement version_config_files field in buildAPIEntity() and readCategory() in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T025 [P] Implement authentication_service field in buildAPIEntity() and readCategory() in /workspace/internal/provider/resource_cmdevice_category.go

**Checkpoint**: All optional fields now have CRUD support - test implementation can proceed

---

## Phase 3: User Story 1 - Installation Mode Field Testing (Priority: P1) - MVP

**Goal**: Test install_mode and new_node_install_mode fields function correctly across CRUD operations

**Independent Test**: Create category with install_mode="AUTO", new_node_install_mode="FULL", update values, verify idempotency

**Note**: These fields ARE implemented in buildAPIEntity/readCategory - can test immediately (Track A)

### Implementation for User Story 1

- [X] T026 [US1] Create testAccCMDeviceCategoryResourceConfig_InstallationModes config helper in /workspace/internal/provider/resource_cmdevice_category_test.go
- [X] T027 [US1] Create TestAccCMDeviceCategoryResource_InstallationModes test function in /workspace/internal/provider/resource_cmdevice_category_test.go
- [X] T028 [US1] Add Step 1: Create with install_mode="AUTO", new_node_install_mode="FULL" with statecheck.ExpectKnownValue assertions
- [X] T029 [US1] Add Step 2: Idempotency check with plancheck.ExpectEmptyPlan()
- [X] T030 [US1] Add Step 3: Update to install_mode="FULL", new_node_install_mode="FULL" with state checks (Note: BCM only accepts FULL for newNodeInstallMode)
- [X] T031 [US1] Add Step 4: Idempotency check after update
- [X] T032 [US1] Add Step 5: Import state verification with ImportStateVerify
- [X] T033 [US1] Add ID consistency tracking with statecheck.CompareValue(compare.ValuesSame())
- [X] T034 [US1] Run test and verify passes: TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategoryResource_InstallationModes

**Checkpoint**: User Story 1 complete - installation mode fields have acceptance test coverage

---

## Phase 4: User Story 2 - Network Settings Field Testing (Priority: P1)

**Goal**: Test name_servers, search_domains, and time_servers list fields handle list operations correctly

**Independent Test**: Create category with network lists, update lists, verify list handling behavior

**Depends on**: T004-T007 (network list field implementation)

### Implementation for User Story 2

- [X] T035 [US2] Create testAccCMDeviceCategoryResourceConfig_NetworkListFields config helper in /workspace/internal/provider/resource_cmdevice_category_test.go
- [X] T036 [US2] Create TestAccCMDeviceCategoryResource_NetworkListFields test function in /workspace/internal/provider/resource_cmdevice_category_test.go
- [X] T037 [US2] Add Step 1: Create with name_servers=["8.8.8.8", "8.8.4.4"], search_domains=["example.com"], time_servers=["ntp.example.com"]
- [X] T038 [US2] Add state checks using knownvalue.ListSizeExact() for list sizes
- [X] T039 [US2] Add Step 2: Idempotency check with plancheck.ExpectEmptyPlan()
- [X] T040 [US2] Add Step 3: Update name_servers list (add/remove items)
- [X] T041 [US2] Add Step 4: Idempotency check after update
- [X] T042 [US2] Add Step 5: Test empty list handling (set to [])
- [ ] T043 [US2] Run test and verify passes: TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategoryResource_NetworkListFields (BLOCKED: requires T004-T007 implementation)

**Checkpoint**: User Story 2 complete - network list fields have acceptance test coverage

---

## Phase 5: User Story 3 - I/O Scheduler and Kernel Fields Testing (Priority: P2)

**Goal**: Test io_scheduler and kernel_version fields for kernel-level configuration

**Independent Test**: Create category with io_scheduler="mq-deadline", update to "none", verify persistence

**Depends on**: T008 (io_scheduler implementation) - kernel_version already implemented

### Implementation for User Story 3

- [ ] T044 [US3] Create testAccCMDeviceCategoryResourceConfig_IOSchedulerKernel config helper in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T045 [US3] Create TestAccCMDeviceCategoryResource_IOSchedulerKernel test function in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T046 [US3] Add Step 1: Create with io_scheduler="mq-deadline" and kernel_version (if environment supports)
- [ ] T047 [US3] Add state checks using knownvalue.StringExact() for io_scheduler
- [ ] T048 [US3] Add Step 2: Idempotency check with plancheck.ExpectEmptyPlan()
- [ ] T049 [US3] Add Step 3: Update io_scheduler to "none"
- [ ] T050 [US3] Add Step 4: Idempotency check after update
- [ ] T051 [US3] Add Step 5: Import state verification
- [ ] T052 [US3] Run test and verify passes: TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategoryResource_IOSchedulerKernel

**Checkpoint**: User Story 3 complete - I/O scheduler and kernel fields have acceptance test coverage

---

## Phase 6: User Story 4 - Provisioning Scripts Field Testing (Priority: P2)

**Goal**: Test initialize and finalize provisioning script fields preserve multi-line content

**Independent Test**: Create category with multi-line scripts, update scripts, verify exact string preservation

**Depends on**: T009-T010 (script field implementation)

### Implementation for User Story 4

- [X] T053 [US4] Create testAccCMDeviceCategoryResourceConfig_ProvisioningScripts config helper in /workspace/internal/provider/resource_cmdevice_category_test.go
- [X] T054 [US4] Create TestAccCMDeviceCategoryResource_ProvisioningScripts test function in /workspace/internal/provider/resource_cmdevice_category_test.go
- [X] T055 [US4] Add Step 1: Create with initialize="#!/bin/bash\necho 'init'" and finalize="#!/bin/bash\necho 'done'"
- [X] T056 [US4] Add state checks using knownvalue.StringExact() with escaped newlines
- [X] T057 [US4] Add Step 2: Idempotency check with plancheck.ExpectEmptyPlan()
- [X] T058 [US4] Add Step 3: Update initialize script content
- [X] T059 [US4] Add Step 4: Idempotency check after update
- [X] T060 [US4] Add Step 5: Verify finalize script unchanged after initialize update
- [X] T061 [US4] Run test and verify passes: TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategoryResource_ProvisioningScripts

**Checkpoint**: User Story 4 complete - provisioning script fields have acceptance test coverage

---

## Phase 7: User Story 5 - BMC Settings Nested Object Testing (Priority: P3)

**Goal**: Test bmc_settings nested object handles all nested fields correctly

**Independent Test**: Create category with bmc_settings containing user_name, privilege, user_id; update individual fields

**Depends on**: T011 (bmc_settings implementation)

### Implementation for User Story 5

- [ ] T062 [US5] Create testAccCMDeviceCategoryResourceConfig_BMCSettings config helper in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T063 [US5] Create TestAccCMDeviceCategoryResource_BMCSettings test function in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T064 [US5] Add Step 1: Create with bmc_settings={user_name="admin", privilege="ADMINISTRATOR", user_id=2}
- [ ] T065 [US5] Add state checks using tfjsonpath.New("bmc_settings").AtMapKey() for nested fields
- [ ] T066 [US5] Add Step 2: Idempotency check with plancheck.ExpectEmptyPlan()
- [ ] T067 [US5] Add Step 3: Update individual bmc_settings field (user_id)
- [ ] T068 [US5] Add Step 4: Verify other bmc_settings fields unchanged
- [ ] T069 [US5] Add Step 5: Import state verification
- [ ] T070 [US5] Run test and verify passes: TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategoryResource_BMCSettings

**Checkpoint**: User Story 5 complete - BMC settings nested object has acceptance test coverage

---

## Phase 8: User Story 6 - Kernel Modules Nested List Testing (Priority: P3)

**Goal**: Test modules nested list handles add/remove/update of module entries

**Independent Test**: Create category with modules=[{name="mlx5_core", parameters="debug=1"}], add/remove modules

**Depends on**: T012 (modules implementation)

### Implementation for User Story 6

- [ ] T071 [US6] Create testAccCMDeviceCategoryResourceConfig_KernelModules config helper in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T072 [US6] Create TestAccCMDeviceCategoryResource_KernelModules test function in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T073 [US6] Add Step 1: Create with modules=[{name="mlx5_core", parameters="debug=1"}]
- [ ] T074 [US6] Add state checks using tfjsonpath.New("modules").AtSliceIndex(0).AtMapKey() for nested list elements
- [ ] T075 [US6] Add Step 2: Idempotency check with plancheck.ExpectEmptyPlan()
- [ ] T076 [US6] Add Step 3: Add second module to list
- [ ] T077 [US6] Add Step 4: Verify list size with knownvalue.ListSizeExact(2)
- [ ] T078 [US6] Add Step 5: Remove module from list
- [ ] T079 [US6] Run test and verify passes: TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategoryResource_KernelModules

**Checkpoint**: User Story 6 complete - kernel modules nested list has acceptance test coverage

---

## Phase 9: User Story 7 - Exclude Lists Field Testing (Priority: P3)

**Goal**: Test exclude list fields handle large text content without truncation

**Independent Test**: Create category with exclude_list_full containing multiple file patterns, update list

**Depends on**: T013-T018 (exclude list field implementations)

### Implementation for User Story 7

- [ ] T080 [US7] Create testAccCMDeviceCategoryResourceConfig_ExcludeLists config helper in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T081 [US7] Create TestAccCMDeviceCategoryResource_ExcludeLists test function in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T082 [US7] Add Step 1: Create with exclude_list_full="/var/log/*\n/tmp/*", exclude_list_sync="/var/cache/*"
- [ ] T083 [US7] Add state checks using knownvalue.StringExact() for multiline content
- [ ] T084 [US7] Add Step 2: Idempotency check with plancheck.ExpectEmptyPlan()
- [ ] T085 [US7] Add Step 3: Update exclude_list_sync content
- [ ] T086 [US7] Add Step 4: Verify exclude_list_full unchanged
- [ ] T087 [US7] Add Step 5: Test exclude_list_manipulate_script field
- [ ] T088 [US7] Run test and verify passes: TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategoryResource_ExcludeLists

**Checkpoint**: User Story 7 complete - exclude list fields have acceptance test coverage

---

## Phase 10: User Story 8 - Miscellaneous Boolean and String Fields Testing (Priority: P3)

**Goal**: Test remaining fields: fips, data_node, interactive_user, use_exclusively_for, node_installer_disk, version_config_files

**Independent Test**: Create category with miscellaneous fields set, update values, verify boolean/string handling

**Depends on**: T019-T025 (miscellaneous field implementations)

### Implementation for User Story 8

- [ ] T089 [US8] Create testAccCMDeviceCategoryResourceConfig_MiscellaneousFields config helper in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T090 [US8] Create TestAccCMDeviceCategoryResource_MiscellaneousFields test function in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T091 [US8] Add Step 1: Create with fips="YES", data_node=true, interactive_user="testuser"
- [ ] T092 [US8] Add state checks using knownvalue.StringExact() for strings and knownvalue.Bool() for booleans
- [ ] T093 [US8] Add Step 2: Idempotency check with plancheck.ExpectEmptyPlan()
- [ ] T094 [US8] Add Step 3: Update fips from "YES" to "NO"
- [ ] T095 [US8] Add Step 4: Idempotency check after update
- [ ] T096 [US8] Add Step 5: Test additional fields (use_exclusively_for, node_installer_disk, version_config_files)
- [ ] T097 [US8] Run test and verify passes: TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategoryResource_MiscellaneousFields

**Checkpoint**: User Story 8 complete - all miscellaneous fields have acceptance test coverage

---

## Phase 11: Polish and Cross-Cutting Concerns

**Purpose**: Final validation and documentation

- [ ] T098 Run full acceptance test suite: TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceCategory"
- [ ] T099 [P] Generate updated documentation: make generate
- [ ] T100 [P] Update field coverage matrix in /workspace/specs/001-category-test-coverage/research.md
- [ ] T101 Verify test execution time is under 15 minutes total
- [ ] T102 Run tests 3 times consecutively to verify zero flakiness

---

## Dependencies and Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS Track B user stories (US2-US8)
- **User Story 1 (Phase 3)**: Can start after Setup - NO dependency on Phase 2 (Track A)
- **User Stories 2-8 (Phases 4-10)**: ALL depend on Phase 2 (Foundational) completion
- **Polish (Phase 11)**: Depends on all user stories being complete

### User Story Dependencies

| User Story | Phase | Track | Blocker |
|------------|-------|-------|---------|
| US1 - Installation Modes | 3 | A (implemented) | None - can start immediately |
| US2 - Network Lists | 4 | B | T004-T007 (list implementation) |
| US3 - I/O Scheduler + Kernel | 5 | B | T008 (io_scheduler) - kernel_version ready |
| US4 - Provisioning Scripts | 6 | B | T009-T010 (script fields) |
| US5 - BMC Settings | 7 | B | T011 (nested object) |
| US6 - Kernel Modules | 8 | B | T012 (nested list) |
| US7 - Exclude Lists | 9 | B | T013-T018 (exclude fields) |
| US8 - Miscellaneous | 10 | B | T019-T025 (misc fields) |

### Within Each User Story

- Config helper before test function
- Test steps in order: Create -> Idempotency -> Update -> Idempotency -> Import
- Run test to verify before marking complete

### Parallel Opportunities

**Phase 2 (Foundational)**: T005, T006 can run parallel with T004; T010 parallel with T009; T014-T018 all parallel; T020-T025 all parallel

**After Phase 2 completes**: All Track B user stories (US2-US8) can be developed in parallel by different team members

**Within each phase**: Tasks marked [P] can run in parallel

---

## Parallel Example: User Story 1

```bash
# User Story 1 can start immediately (Track A - no blockers)

# Step 1: Create config helper
Task T026: Create testAccCMDeviceCategoryResourceConfig_InstallationModes

# Step 2: Create test function and steps (sequential)
Task T027: Create TestAccCMDeviceCategoryResource_InstallationModes
Task T028-T033: Add test steps (sequential - each builds on previous)

# Step 3: Run and verify
Task T034: TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategoryResource_InstallationModes
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T003)
2. Skip Phase 2: Foundational (not needed for Track A)
3. Complete Phase 3: User Story 1 (T026-T034)
4. **STOP and VALIDATE**: Test installation mode fields work
5. Deploy test coverage improvement

### Incremental Delivery

1. Complete Setup -> Ready
2. Add User Story 1 (Track A) -> Test independently -> Deploy (MVP!)
3. Complete Phase 2: Foundational (implement missing fields)
4. Add User Story 2 -> Test independently -> Deploy
5. Add User Story 3-8 -> Test independently -> Deploy
6. Each story adds coverage without breaking previous tests

### Parallel Team Strategy

With multiple developers:

1. Developer A: User Story 1 (Track A - no blockers)
2. Developer B: Phase 2 Foundational (T004-T025)
3. After Phase 2 completes:
   - Developer A: User Story 2
   - Developer B: User Story 3
   - Developer C: User Story 4
   - (etc.)

---

## Task Summary

| Phase | Tasks | Parallel | User Story |
|-------|-------|----------|------------|
| Setup | T001-T003 | 2 | - |
| Foundational | T004-T025 | 14 | - |
| US1 | T026-T034 | 0 | Installation Modes |
| US2 | T035-T043 | 0 | Network Lists |
| US3 | T044-T052 | 0 | I/O Scheduler + Kernel |
| US4 | T053-T061 | 0 | Provisioning Scripts |
| US5 | T062-T070 | 0 | BMC Settings |
| US6 | T071-T079 | 0 | Kernel Modules |
| US7 | T080-T088 | 0 | Exclude Lists |
| US8 | T089-T097 | 0 | Miscellaneous |
| Polish | T098-T102 | 2 | - |

**Total Tasks**: 102
**Parallel Opportunities**: 18 tasks can run in parallel
**MVP Scope**: Setup (3 tasks) + User Story 1 (9 tasks) = 12 tasks

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Track A (US1) can be completed independently of Track B
- Track B requires Phase 2 (Foundational) implementation work first
- Verify tests fail before implementing (TDD)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- All tests must use modern patterns: statecheck, plancheck, knownvalue matchers
