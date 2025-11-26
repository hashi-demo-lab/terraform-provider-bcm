# Tasks: Fix roles[].uuid Computed Value Population

**Input**: Design documents from `/workspace/specs/083-roles-uuid-computed/`
**Prerequisites**: plan.md (required), spec.md (required for user stories)

**Tests**: Acceptance tests are REQUIRED - this is a TDD bug fix following RED-GREEN-REFACTOR workflow.

**Organization**: Tasks follow TDD phases (RED -> GREEN -> REFACTOR) aligned with user stories from spec.md.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

- **Project Type**: Terraform Provider (single project)
- **Primary File**: `internal/provider/resource_cmdevice_category.go`
- **Test File**: `internal/provider/resource_cmdevice_category_test.go`

---

## Phase 1: Setup (Analysis & Preparation)

**Purpose**: Understand current behavior and prepare for TDD workflow

- [ ] T001 Review root cause: preservation logic at lines 1068-1077, 1189-1195 in internal/provider/resource_cmdevice_category.go
- [ ] T002 Verify role UUID parsing works correctly at lines 2249-2276 in internal/provider/resource_cmdevice_category.go
- [ ] T003 [P] Identify test helper dependencies in internal/provider/test_helpers.go

---

## Phase 2: RED - Write Failing Tests First

**Purpose**: Create acceptance tests that demonstrate the bug (tests MUST fail before implementation)

**CRITICAL**: All tests in this phase MUST fail initially - this confirms the bug exists

### Tests for User Story 1 - Role UUID Available After Create (P1)

- [ ] T004 [US1] Add test config helper `testAccCMDeviceCategoryResourceConfig_WithRole` in internal/provider/resource_cmdevice_category_test.go
- [ ] T005 [US1] Write `TestAccCMDeviceCategory_RolesUUIDPopulated` test in internal/provider/resource_cmdevice_category_test.go
- [ ] T006 [US1] Run test and verify it FAILS (uuid is null) with `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategory_RolesUUIDPopulated`

### Tests for User Story 2 - Role UUID Preserved on Refresh (P1)

- [ ] T007 [US2] Write `TestAccCMDeviceCategory_RolesUUIDPreservedOnRefresh` test in internal/provider/resource_cmdevice_category_test.go
- [ ] T008 [US2] Write `TestAccCMDeviceCategory_RolesIdempotency` test in internal/provider/resource_cmdevice_category_test.go
- [ ] T009 [US2] Run tests and verify they FAIL with `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "RolesUUID|RolesIdempotency"`

### Tests for User Story 3 - Role UUID Available After Import (P2)

- [ ] T010 [US3] Write import state verification test within `TestAccCMDeviceCategory_RolesUUIDPopulated` test steps
- [ ] T011 [US3] Verify import test FAILS (imported state has null uuid)

### Tests for User Story 4 - Merge User Config with API Values (P1)

- [ ] T012 [P] [US4] Add test config helper `testAccCMDeviceCategoryResourceConfig_MultipleRoles` in internal/provider/resource_cmdevice_category_test.go
- [ ] T013 [US4] Write `TestAccCMDeviceCategory_MultipleRolesUUID` test in internal/provider/resource_cmdevice_category_test.go
- [ ] T014 [US4] Run test and verify it FAILS with `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategory_MultipleRolesUUID`

**Checkpoint**: All new tests FAIL - confirms bug exists and tests are valid

---

## Phase 3: GREEN - Implement Minimal Fix

**Purpose**: Write the minimum code needed to make all failing tests pass

### Implementation for Core Fix (All User Stories)

- [ ] T015 Add `mergeRolesWithAPIResponse` helper function in internal/provider/resource_cmdevice_category.go (after line 2276)
- [ ] T016 Replace line 1193 (`state.Roles = originalRoles`) with merge function call in internal/provider/resource_cmdevice_category.go
- [ ] T017 Run all new tests to verify they PASS with `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "RolesUUID|RolesIdempotency|MultipleRoles"`

**Checkpoint**: All new tests PASS - bug is fixed

---

## Phase 4: REFACTOR - Verify & Polish

**Purpose**: Ensure quality, verify no regressions, update documentation

### Verification Tasks

- [ ] T018 Run ALL existing category tests to verify no regression with `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "CMDeviceCategory"`
- [ ] T019 [P] Run linter to verify code quality with `make lint`
- [ ] T020 [P] Run `make generate` to update documentation

### Documentation Tasks

- [ ] T021 [P] Add debug logging to `mergeRolesWithAPIResponse` function in internal/provider/resource_cmdevice_category.go
- [ ] T022 [P] Add code comment explaining the fix (reference issue #83) in internal/provider/resource_cmdevice_category.go

---

## Phase 5: Polish & Final Validation

**Purpose**: Final verification and cleanup

- [ ] T023 Run full test suite for category resource with `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run CMDeviceCategory`
- [ ] T024 Verify example configurations still work in examples/resources/bcm_cmdevice_category/
- [ ] T025 Update CHANGELOG.md with bug fix entry (if project has changelog)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies - analysis and preparation
- **Phase 2 (RED)**: Depends on Phase 1 - write failing tests
- **Phase 3 (GREEN)**: Depends on Phase 2 - tests must fail first before implementing fix
- **Phase 4 (REFACTOR)**: Depends on Phase 3 - tests must pass before refactoring
- **Phase 5 (Polish)**: Depends on Phase 4 - all tests must be stable

### Task Dependencies Within Phases

**Phase 2 (RED)**:
- T004 -> T005 -> T006 (config helper before test, test before run)
- T007 and T008 can run in parallel (different tests)
- T012 -> T013 -> T014 (config helper before test, test before run)

**Phase 3 (GREEN)**:
- T015 -> T016 -> T017 (helper function, then integration, then verify)

**Phase 4 (REFACTOR)**:
- T018 blocks T019, T020, T021, T022 (regression tests first)
- T19, T20, T21, T22 marked [P] can run in parallel after T018

### User Story to Task Mapping

| User Story | Priority | Tasks |
|------------|----------|-------|
| US1 - Role UUID After Create | P1 | T004, T005, T006 |
| US2 - Role UUID on Refresh | P1 | T007, T008, T009 |
| US3 - Role UUID After Import | P2 | T010, T011 |
| US4 - Merge Config with API | P1 | T012, T013, T014 |
| All Stories (Core Fix) | - | T015, T016, T017 |

### Parallel Opportunities

**Phase 2 (RED) - Parallel Test Writing**:
```bash
# Can run these test file edits in parallel:
T007 [US2] Write TestAccCMDeviceCategory_RolesUUIDPreservedOnRefresh
T008 [US2] Write TestAccCMDeviceCategory_RolesIdempotency
T012 [US4] Add testAccCMDeviceCategoryResourceConfig_MultipleRoles helper
```

**Phase 4 (REFACTOR) - Parallel Verification**:
```bash
# After T018 completes, these can run in parallel:
T019 [P] make lint
T020 [P] make generate
T021 [P] Add debug logging
T022 [P] Add code comments
```

---

## Implementation Strategy

### TDD Workflow (RED-GREEN-REFACTOR)

1. **RED Phase (Phase 2)**: Write all tests first, verify they fail
   - Creates confidence that tests actually validate the bug
   - Documents expected behavior before implementation

2. **GREEN Phase (Phase 3)**: Implement minimal fix
   - Add `mergeRolesWithAPIResponse` helper function
   - Replace unconditional role preservation with merge call
   - Focus on making tests pass, not perfect code

3. **REFACTOR Phase (Phase 4)**: Polish and verify
   - Run ALL existing tests to catch regressions
   - Add logging, comments, documentation
   - Code quality checks

### MVP Completion Criteria

After Phase 3 (GREEN) completes:
- All four user stories are addressed by the fix
- Role UUIDs populated after create (US1)
- Role UUIDs preserved on refresh (US2)
- Role UUIDs populated after import (US3)
- User config merged with API values (US4)

### Success Criteria Validation

| Criteria | Validation Task |
|----------|-----------------|
| SC-001: 100% roles have non-empty uuid | T017 (tests pass) |
| SC-002: Idempotency verified | T008, T017 |
| SC-003: Refresh preserves UUIDs | T007, T017 |
| SC-004: Existing tests pass | T018 |
| SC-005: Drift detection works | T007, T008, T017 |

---

## Summary

| Metric | Value |
|--------|-------|
| Total Tasks | 25 |
| Phase 1 (Setup) | 3 tasks |
| Phase 2 (RED) | 11 tasks |
| Phase 3 (GREEN) | 3 tasks |
| Phase 4 (REFACTOR) | 5 tasks |
| Phase 5 (Polish) | 3 tasks |
| Parallelizable Tasks | 8 tasks marked [P] |
| User Stories Covered | 4 (US1, US2, US3, US4) |
| Files Modified | 2 (resource_cmdevice_category.go, resource_cmdevice_category_test.go) |

---

## Notes

- [P] tasks = different files/operations, no dependencies
- [Story] label maps task to specific user story for traceability
- TDD is MANDATORY - tests must fail before implementation (Phase 2 before Phase 3)
- All existing tests must continue to pass (verified in T018)
- Commit after each phase completion for clean git history
- Reference issue #83 in commit messages
