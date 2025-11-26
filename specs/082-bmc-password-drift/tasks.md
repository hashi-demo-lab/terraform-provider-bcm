# Tasks: Fix BMC Settings Password Perpetual Drift

**Input**: Design documents from `/workspace/specs/082-bmc-password-drift/`
**Prerequisites**: plan.md (complete), spec.md (complete)
**GitHub Issue**: #82
**Branch**: `082-bmc-password-drift`

**TDD Workflow**: RED-GREEN-REFACTOR with acceptance tests first

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Source**: `internal/provider/`
- **Tests**: `internal/provider/*_test.go`
- **Specs**: `/workspace/specs/082-bmc-password-drift/`

---

## Phase 1: Setup (No Setup Required)

**Purpose**: This is a bug fix in an existing provider - no project initialization needed

**Status**: SKIP - Existing project structure is already in place

---

## Phase 2: Foundational (Test Infrastructure)

**Purpose**: Add test configuration helper that both user stories depend on

- [X] T001 Add `testAccCMDeviceCategoryResourceConfig_BMCPassword` helper function in `internal/provider/resource_cmdevice_category_test.go`

**Checkpoint**: Test helper ready - acceptance tests can now be written

---

## Phase 3: User Story 1 - Password Stability on Refresh (Priority: P1)

**Goal**: `bmc_settings.password` remains stable across plan/apply cycles - no false drift

**Independent Test**: Create category with BMC password, run `terraform plan` after apply, verify no changes detected

### RED Phase: Write Failing Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T002 [US1] Write `TestAccCMDeviceCategory_BMCPasswordNoDrift` acceptance test in `internal/provider/resource_cmdevice_category_test.go`
  - Step 1: Create category with BMC password `"secret123"`
  - Step 2: Idempotency check with `plancheck.ExpectEmptyPlan()` - THIS SHOULD FAIL initially
  - Step 3: Import and verify (ignore `bmc_settings` as password cannot be imported)
- [X] T003 [US1] Run acceptance test to confirm it FAILS (RED phase verification) using `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategory_BMCPasswordNoDrift`

### GREEN Phase: Implement Fix for User Story 1

- [X] T004 [US1] Add `originalBMCSettings` capture before `readCategory()` call at line ~1077 in `internal/provider/resource_cmdevice_category.go`
- [X] T005 [US1] Add BMC password preservation logic after `readCategory()` call at line ~1196 in `internal/provider/resource_cmdevice_category.go`
  - Extract original password from `originalBMCSettings` using `basetypes.ObjectAsOptions{}`
  - Build merged BMC model with API values + preserved password
  - Convert back to `types.Object` and set on state
  - Add `tflog.Debug` for observability
- [X] T006 [US1] Run acceptance test to confirm it PASSES (GREEN phase verification) using `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategory_BMCPasswordNoDrift`

**Checkpoint**: User Story 1 complete - no drift on `terraform plan` after apply with BMC password

---

## Phase 4: User Story 2 - Password Update Detection (Priority: P1)

**Goal**: Password changes in configuration are detected and applied correctly

**Independent Test**: Modify password value in configuration, verify change is detected and applied

### RED Phase: Write Failing Tests for User Story 2

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**
> **NOTE: If implementation from US1 is already in place, this test should PASS immediately (same fix covers both stories)**

- [X] T007 [US2] Write `TestAccCMDeviceCategory_BMCPasswordUpdate` acceptance test in `internal/provider/resource_cmdevice_category_test.go`
  - Step 1: Create category with initial password `"oldpass123"`
  - Step 2: Update password to `"newpass456"` - verify change applied
  - Step 3: Idempotency check with `plancheck.ExpectEmptyPlan()`
- [X] T008 [US2] Run acceptance test to verify it PASSES (since fix from US1 covers this case) using `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategory_BMCPasswordUpdate`

**Checkpoint**: User Story 2 complete - password changes detected and applied correctly

---

## Phase 5: User Story 3 - Import with Password (Priority: P2)

**Goal**: Users can import existing categories and add BMC password without unexpected drift

**Independent Test**: Import a category, add BMC password to config, verify stable state after apply

**NOTE**: This is covered by the import step in `TestAccCMDeviceCategory_BMCPasswordNoDrift` (T002). No additional implementation needed - import naturally works with the state preservation pattern.

- [X] T009 [US3] Verify import behavior in existing test by reviewing Step 3 of `TestAccCMDeviceCategory_BMCPasswordNoDrift` output

**Checkpoint**: User Story 3 verified - import works correctly with password handling

---

## Phase 6: Polish & Validation

**Purpose**: Regression testing, documentation, and code quality

### Regression Testing

- [X] T010 [P] Run full acceptance test suite for `bcm_cmdevice_category` resource using `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategory`
- [X] T011 [P] Run unit tests to ensure no regressions using `make test`
- [X] T012 [P] Run linting and formatting using `make fmt && make lint`

### Documentation Updates

- [X] T013 Generate provider documentation using `make generate`
- [X] T014 Update or remove skip comment on existing `TestAccCMDeviceCategoryResource_BMCSettings` test if present in `internal/provider/resource_cmdevice_category_test.go`

### Final Validation

- [X] T015 Review code changes against spec.md success criteria (SC-001 through SC-004)
- [X] T016 Verify all edge cases from spec.md are handled:
  - `bmc_settings` null in state - skip preservation
  - `bmc_settings` added first time - password from plan, preserved on subsequent reads
  - Password empty string `""` - preserved as empty string (not converted to null)
  - Only non-password BMC fields updated - password preserved

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: SKIPPED - existing project
- **Foundational (Phase 2)**: No dependencies - T001 can start immediately
- **User Story 1 (Phase 3)**: Depends on T001 completion
- **User Story 2 (Phase 4)**: Depends on Phase 3 completion (uses same fix)
- **User Story 3 (Phase 5)**: Depends on Phase 3 completion (uses same fix)
- **Polish (Phase 6)**: Depends on all user stories complete

### Task Dependencies Graph

```
T001 (test helper)
  |
  v
T002 (write NoDrift test) --> T003 (run - expect FAIL)
  |
  v
T004 (capture originalBMCSettings) --> T005 (preservation logic)
  |
  v
T006 (run NoDrift - expect PASS)
  |
  v
T007 (write Update test) --> T008 (run Update - expect PASS)
  |
  v
T009 (verify import behavior)
  |
  v
T010, T011, T012 (parallel regression tests)
  |
  v
T013 (generate docs) --> T014 (update test comments)
  |
  v
T015 (review criteria) --> T016 (verify edge cases)
```

### Parallel Opportunities

- T010, T011, T012 can run in parallel (different test scopes)
- All tasks within a user story must be sequential (TDD: RED-GREEN)

---

## Parallel Example: Regression Testing

```bash
# Launch all regression tests together (Phase 6):
Task T010: "Run full acceptance test suite for bcm_cmdevice_category"
Task T011: "Run unit tests to ensure no regressions"
Task T012: "Run linting and formatting"
```

---

## Implementation Strategy

### TDD Execution Order

1. **T001**: Add test config helper function
2. **T002-T003**: Write failing acceptance test (RED) - verify failure
3. **T004-T005**: Implement password preservation (GREEN)
4. **T006**: Verify test passes (GREEN complete)
5. **T007-T008**: Write and verify update test (already GREEN due to same fix)
6. **T009**: Verify import behavior
7. **T010-T016**: Polish phase (parallel regression + docs)

### MVP Scope

The MVP is **User Story 1 only** (T001-T006):
- Creates test infrastructure
- Validates fix with idempotency test
- Implements password preservation

User Stories 2 and 3 are validation tests that confirm the same fix handles additional scenarios.

### Success Criteria Mapping

| Success Criteria | Task(s) | Verification |
|-----------------|---------|--------------|
| SC-001: No drift on plan after apply | T002, T006 | `plancheck.ExpectEmptyPlan()` passes |
| SC-002: Password changes detected | T007, T008 | Update test passes |
| SC-003: No regression | T010, T011 | All existing tests pass |
| SC-004: Idempotency | T002, T006 | Empty plan check passes |

---

## Notes

- [P] tasks = different files/scopes, can run in parallel
- [Story] label maps task to specific user story from spec.md
- Tests MUST fail before implementation (TDD RED-GREEN pattern)
- Total tasks: 16
- Tasks per user story: US1=5, US2=2, US3=1, Setup=0, Foundational=1, Polish=7
- Parallel opportunities: T010/T011/T012 (regression testing)
- Estimated time: 1-2 hours for implementation, 30+ minutes for acceptance tests
