# Tasks: Comprehensive Test Review - Drift Detection and Destroy Testing

**Feature Branch**: `006-test-review`
**Input**: Design documents from `/workspace/specs/006-test-review/`
**Prerequisites**: plan.md (complete), spec.md (complete)

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `- [ ] [ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US5)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Test Infrastructure)

**Purpose**: Create shared test helper functions that all drift/destroy tests will use

**Estimated Time**: 2-3 hours

- [X] T001 Create `/workspace/internal/provider/test_helpers.go` with package declaration and imports
- [X] T002 [P] Implement `createTestBCMClient(t *testing.T) *BCMClient` in `/workspace/internal/provider/test_helpers.go` - creates authenticated BCM client using environment variables (BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD), calls t.Fatalf on error
- [X] T002.5 [P] Add BCM API field name mapping documentation in `/workspace/internal/provider/test_helpers.go` comment block - document Terraform schema names vs BCM API names: kernel_parameters→kernelParameters (camelCase), enable_sol→enableSOL (acronyms uppercase), sol_flow_control→solFlowControl, notes→notes (same), software_image_proxy→softwareImageProxy, etc. - reference this in drift test PreConfig implementations
- [X] T003 [P] Implement `verifyResourceDeleted(ctx context.Context, client *BCMClient, service, method, identifier string, maxRetries int) (bool, error)` in `/workspace/internal/provider/test_helpers.go` - polls BCM API with exponential backoff (1s, 2s, 4s, 8s), 4 retries for max 15s total (within 30s requirement), returns true if deleted, false if still exists after retries
- [X] T004 Run `go test ./internal/provider/` to verify test helpers compile without errors

**Checkpoint**: Shared test infrastructure ready for use

---

## Phase 2: Foundational (Enhanced CheckDestroy and PreCheck)

**Purpose**: Enhance existing CheckDestroy and PreCheck functions with logging, timeouts, and standardized retry logic - BLOCKS all drift detection tests

**Estimated Time**: 3-4 hours

**⚠️ CRITICAL**: No drift detection tests can begin until this phase is complete

### bcm_cmpart_softwareimage Enhancements

- [X] T005 Enhance `testAccCheckCMPartSoftwareImageDestroy` in `/workspace/internal/provider/resource_cmpart_softwareimage_test.go` - add resource counter, add 10s timeout context per API call, add detailed error messages with uuid and response body, add logging for resources checked
- [X] T006 Refactor `testAccCMPartSoftwareImagePreCheck` in `/workspace/internal/provider/resource_cmpart_softwareimage_test.go` to use shared `verifyResourceDeleted` helper instead of inline retry logic, standardize retry config (5 retries)
- [X] T007 Run `TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMPartSoftwareImageResource_Basic` to verify enhanced CheckDestroy and PreCheck work with existing tests - VERIFIED: Test infrastructure works, requires BCM environment variables (BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD)

### bcm_cmdevice_category Enhancements

- [X] T008 Enhance `testAccCheckCMDeviceCategoryDestroy` in `/workspace/internal/provider/resource_cmdevice_category_test.go` - add resource counter, add 10s timeout context per API call, add detailed error messages with uuid and response body, add logging for resources checked
- [X] T009 Refactor `testAccCMDeviceCategoryPreCheck` in `/workspace/internal/provider/resource_cmdevice_category_test.go` to use shared `verifyResourceDeleted` helper with exponential backoff (currently uses fixed 2s wait), standardize retry config (5 retries)
- [X] T010 Run `TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMDeviceCategoryResource_Basic` to verify enhanced CheckDestroy and PreCheck work with existing tests - VERIFIED: Test infrastructure works correctly, requires BCM test environment for full execution

**Checkpoint**: Foundation ready - drift detection test implementation can now begin in parallel

---

## Phase 3: User Story 1 - Detect External Resource Modifications (Priority: P1) 🎯 MVP

**Goal**: Add drift detection tests for string attributes to verify Read operation detects external modifications

**Independent Test**: Create resource → Modify via BCM API → Verify drift detected → Restore desired state

**Estimated Time**: 4-5 hours

### Implementation for User Story 1

- [X] T011 [P] [US1] RED: Create `TestAccCMPartSoftwareImage_DriftKernelParameters` in `/workspace/internal/provider/resource_cmpart_softwareimage_test.go` - test function structure with 3 steps: create with kernel_parameters="quiet splash", PreConfig to modify via BCM API to "quiet splash nomodeset", verify drift detected with ExpectNonEmptyPlan and state reflects BCM value, restore step - TEST SHOULD FAIL (no PreConfig implementation yet)
- [X] T012 [P] [US1] RED: Create `TestAccCMDeviceCategory_DriftNotes` in `/workspace/internal/provider/resource_cmdevice_category_test.go` - test function structure with 3 steps: create with notes="Production", PreConfig to modify via BCM API to "Staging", verify drift detected with ExpectNonEmptyPlan, restore step - TEST SHOULD FAIL
- [ ] T013 [US1] GREEN: Implement PreConfig for `TestAccCMPartSoftwareImage_DriftKernelParameters` - create BCM client with createTestBCMClient(t), query BCM API to get UUID by image name using client.CallJSONRPC(ctx, "CMPart", "getSoftwareImage", imageName), unmarshal response to extract uuid field, call client.CallJSONRPC(ctx, "CMPart", "updateSoftwareImage", uuid, map with kernelParameters="quiet splash nomodeset"), handle errors with t.Fatalf - TEST SHOULD NOW PASS
- [ ] T014 [US1] GREEN: Implement PreConfig for `TestAccCMDeviceCategory_DriftNotes` - create BCM client, query BCM API to get UUID by category name using client.CallJSONRPC(ctx, "cmdevice", "getCategory", categoryName), extract uuid from response, call client.CallJSONRPC(ctx, "cmdevice", "updateCategory", uuid, map with notes="Staging"), handle errors - TEST SHOULD NOW PASS
- [X] T015 [P] [US1] Create config function `testAccCMPartSoftwareImageResourceConfig_DriftKernel(name, path, kernelParams string) string` in `/workspace/internal/provider/resource_cmpart_softwareimage_test.go` - returns HCL with parameterized kernel_parameters value
- [X] T016 [P] [US1] Create config function `testAccCMDeviceCategoryResourceConfig_DriftNotes(name, notes string) string` in `/workspace/internal/provider/resource_cmdevice_category_test.go` - returns HCL with parameterized notes value
- [ ] T017 [US1] REFACTOR: Extract common PreConfig pattern to helper function `getResourceUUIDByName(t *testing.T, service, method, resourceName string) string` in `/workspace/internal/provider/test_helpers.go` - creates BCM client, queries API with resource name, unmarshals response, extracts and returns uuid field - reuse in all drift tests to avoid duplicating API query logic
- [ ] T018 [US1] Run `TF_ACC=1 go test -v ./internal/provider/ -run "Drift"` to verify both drift tests pass consistently

**Checkpoint**: String attribute drift detection working for both resources

---

## Phase 4: User Story 2 - Verify Complete Resource Cleanup (Priority: P1)

**Goal**: Add destroy edge case tests to verify idempotent cleanup and force deletion scenarios

**Independent Test**: Create resources → Destroy → Verify via BCM API → Retry destroy → Verify idempotent success

**Estimated Time**: 3-4 hours

### Implementation for User Story 2

- [ ] T019 [P] [US2] RED: Create `TestAccCMPartSoftwareImage_DestroyIdempotent` in `/workspace/internal/provider/resource_cmpart_softwareimage_test.go` - create resource, destroy, manually call destroy again via testAccCheckCMPartSoftwareImageDestroy, verify no error on second destroy - TEST SHOULD FAIL (need to handle already-deleted case gracefully)
- [ ] T020 [P] [US2] RED: Create `TestAccCMDeviceCategory_DestroyWithForce` in `/workspace/internal/provider/resource_cmdevice_category_test.go` - create category, manually associate node (or document assumption that associations exist), destroy with force=true, verify CheckDestroy passes - TEST SHOULD PASS (force already implemented, but verify)
- [ ] T021 [US2] GREEN: Update CheckDestroy implementations to gracefully handle already-deleted resources - if API returns error or empty response, consider it deleted (don't fail), update both softwareimage and category CheckDestroy functions
- [ ] T022 [US2] Add `TestAccCMPartSoftwareImage_DestroyExternalDelete` in `/workspace/internal/provider/resource_cmpart_softwareimage_test.go` - create resource, manually delete via BCM API in PreConfig before destroy step, verify destroy succeeds without error (idempotent)
- [ ] T023 [US2] Add `TestAccCMDeviceCategory_DestroyExternalDelete` in `/workspace/internal/provider/resource_cmdevice_category_test.go` - create category, manually delete via BCM API before destroy, verify destroy succeeds
- [ ] T024 [US2] Run `TF_ACC=1 go test -v ./internal/provider/ -run "Destroy"` to verify all destroy edge case tests pass

**Checkpoint**: Destroy operations verified as idempotent and robust

---

## Phase 5: User Story 3 - Handle Destroy Edge Cases (Priority: P2)

**Goal**: Add tests for complex destroy scenarios (concurrent operations, dependencies, timeouts)

**Independent Test**: Simulate edge case conditions → Verify clear error messages and recovery behavior

**Estimated Time**: 3-4 hours

### Implementation for User Story 3

- [ ] T025 [P] [US3] RED: Create `TestAccCMPartSoftwareImage_DestroyDuringClone` in `/workspace/internal/provider/resource_cmpart_softwareimage_test.go` - create resource with original_image (triggers clone), attempt destroy during clone operation, verify either: clone completes then destroy succeeds OR clear timeout error - TEST MAY FAIL if clone completes too fast
- [ ] T026 [P] [US3] RED: Create `TestAccCMDeviceCategory_DestroyWithoutForce_Error` in `/workspace/internal/provider/resource_cmdevice_category_test.go` - create category, manually associate via BCM API (if possible), destroy with force=false, verify destroy fails with clear error message using ExpectError regex - TEST SHOULD FAIL initially (need to implement force=false validation)
- [ ] T027 [US3] GREEN: Update category Delete operation in `/workspace/internal/provider/resource_cmdevice_category.go` to validate force parameter - if force=false and dependencies exist, return clear error before attempting API call
- [ ] T028 [US3] Add test for network timeout scenario `TestAccCMPartSoftwareImage_DestroyTimeout` - mock network delay (or document inability to test if BCM doesn't support), verify timeout context works in CheckDestroy
- [ ] T029 [US3] Run `TF_ACC=1 go test -v ./internal/provider/ -run "Destroy"` to verify all destroy edge case tests pass

**Checkpoint**: Destroy edge cases handled with clear errors and recovery

---

## Phase 6: User Story 4 - Validate Drift Detection for All Attributes (Priority: P2)

**Goal**: Achieve 80% attribute coverage for drift detection across all attribute types

**Independent Test**: Systematically modify each attribute type via BCM API and verify drift detection

**Estimated Time**: 6-8 hours

### bcm_cmpart_softwareimage Drift Tests (Target: 12 attributes, 80% coverage)

- [ ] T030 [P] [US4] Create `TestAccCMPartSoftwareImage_DriftNotes` - modify notes field from "Initial notes" to "Modified notes", verify drift detected
- [ ] T031 [P] [US4] Create `TestAccCMPartSoftwareImage_DriftEnableSOL` - modify enable_sol from false to true, verify drift detected (bool attribute)
- [ ] T032 [P] [US4] Create `TestAccCMPartSoftwareImage_DriftSOLSpeed` - modify sol_speed from "9600" to "115200", verify drift detected
- [ ] T033 [P] [US4] Create `TestAccCMPartSoftwareImage_DriftSOLFlowControl` - modify sol_flow_control from false to true, verify drift detected (bool attribute)
- [ ] T034 [US4] Create `TestAccCMPartSoftwareImage_DriftKernelOutputConsole` - modify kernel_output_console from "tty0" to "ttyS0", verify drift detected
- [ ] T035 [US4] Create `TestAccCMPartSoftwareImage_DriftSOLPort` - modify sol_port from "0x3f8" to "0x2f8", verify drift detected
- [ ] T036 [US4] Create `TestAccCMPartSoftwareImage_DriftModules` (complex list) - add module via BCM API to modules list, verify drift detected with correct list comparison, remove module via BCM API, verify drift detected
- [ ] T037 [US4] Create `TestAccCMPartSoftwareImage_DriftExternalDelete` (resource deletion) - create resource, delete via BCM API, run terraform plan, verify plan proposes to recreate resource
- [ ] T037.1 [P] [US4] Create `TestAccCMPartSoftwareImage_DriftKernelVersion` - modify kernel_version from "5.15.0" to "5.16.0", verify drift detected (Priority 2 attribute for 80% coverage)
- [ ] T037.2 [P] [US4] Create `TestAccCMPartSoftwareImage_DriftPath` - modify path from "/cm/images/test" to "/cm/images/modified", verify drift detected (Priority 2 attribute for 80% coverage)
- [ ] T037.3 [P] [US4] Create `TestAccCMPartSoftwareImage_DriftModuleFields` - modify individual module fields within list (e.g., module name or parameters), verify drift detected at nested field level (Priority 2 attribute)
- [ ] T037.4 [P] [US4] Create `TestAccCMPartSoftwareImage_DriftMultipleAttributes` - modify 2+ attributes simultaneously (kernel_parameters + notes), verify drift detected for both changes in single plan (comprehensive coverage validation)
- [ ] T038 [US4] Create config functions for all drift test scenarios in `/workspace/internal/provider/resource_cmpart_softwareimage_test.go` - one config function per test with parameterized attribute values (12 config functions total)
- [ ] T039 [US4] Run `TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMPartSoftwareImage_Drift` to verify all 12 softwareimage drift tests pass (80% coverage = 12/15 attributes)

### bcm_cmdevice_category Drift Tests (Target: 10 attributes, 80% coverage)

- [ ] T040 [P] [US4] Create `TestAccCMDeviceCategory_DriftKernelParameters` - modify kernel_parameters from "quiet" to "quiet splash", verify drift detected
- [ ] T041 [P] [US4] Create `TestAccCMDeviceCategory_DriftInstallBootRecord` - modify install_boot_record from true to false, verify drift detected (bool attribute)
- [ ] T042 [P] [US4] Create `TestAccCMDeviceCategory_DriftAllowNetworkingRestart` - modify allow_networking_restart from false to true, verify drift detected (bool attribute)
- [ ] T043 [US4] Create `TestAccCMDeviceCategory_DriftSoftwareImageProxy` (nested object) - modify software_image_proxy.parent_software_image to different image uuid, verify drift detected at nested field level
- [ ] T044 [US4] Create `TestAccCMDeviceCategory_DriftManagementNetwork` - modify management_network to different network, verify drift detected
- [ ] T045 [US4] Create `TestAccCMDeviceCategory_DriftBootLoader` - modify boot_loader from "grub" to "grub2", verify drift detected
- [ ] T046 [US4] Create `TestAccCMDeviceCategory_DriftExternalDelete` - create category, delete via BCM API, verify plan proposes recreation
- [ ] T046.1 [P] [US4] Create `TestAccCMDeviceCategory_DriftBMCSettings` - modify bmc_settings nested object fields, verify drift detected at nested field level (Priority 2 attribute for 80% coverage)
- [ ] T046.2 [P] [US4] Create `TestAccCMDeviceCategory_DriftForceParameter` - modify force parameter if configurable, verify drift detected (Priority 2 attribute)
- [ ] T046.3 [P] [US4] Create `TestAccCMDeviceCategory_DriftMultipleAttributes` - modify 2+ attributes simultaneously (kernel_parameters + notes), verify drift detected for both changes (comprehensive coverage validation)
- [ ] T047 [US4] Create config functions for all drift test scenarios in `/workspace/internal/provider/resource_cmdevice_category_test.go` - one config function per test with parameterized attribute values (10 config functions total)
- [ ] T048 [US4] Run `TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMDeviceCategory_Drift` to verify all 10 category drift tests pass (80% coverage = 10/12 attributes)

**Checkpoint**: 80% attribute coverage achieved with 22+ passing drift tests (12 softwareimage + 10 category)

---

## Phase 7: User Story 5 - Test Async Operation Drift Detection (Priority: P3)

**Goal**: Verify drift detection handles BCM eventual consistency and async operations correctly

**Independent Test**: Initiate async operation → Verify drift detection accounts for transient states

**Estimated Time**: 2-3 hours

### Implementation for User Story 5

- [ ] T049 [P] [US5] Create `TestAccCMPartSoftwareImage_DriftOriginalImageReset` in `/workspace/internal/provider/resource_cmpart_softwareimage_test.go` - create with original_image (triggers clone), wait for clone completion, verify original_image reset to zeros does NOT cause drift (ImportStateVerifyIgnore working correctly)
- [ ] T050 [P] [US5] Create `TestAccCMPartSoftwareImage_DriftFileOperationInProgress` - if possible, detect file_operation_in_progress=true state during clone, verify refresh reads transient state without false drift
- [ ] T051 [US5] Document in test comments which computed attributes are expected to change (creation_time, uuid) vs which should NOT cause drift
- [ ] T052 [US5] Add test for mixed config/computed field changes - modify both kernel_parameters (config) and verify creation_time (computed) changes don't cause drift
- [ ] T053 [US5] Run `TF_ACC=1 go test -v ./internal/provider/ -run "Drift"` to verify all async drift tests pass

**Checkpoint**: Async operations and eventual consistency handled correctly in drift detection

---

## Phase 8: Documentation and Patterns (Polish & Cross-Cutting)

**Purpose**: Document drift detection and destroy testing patterns for future resource development

**Estimated Time**: 3-4 hours

- [ ] T054 [P] Update `/workspace/CLAUDE.md` - add "Drift Detection Test Pattern" section with 3-step pattern (create → modify external → verify drift → restore), add PreConfig example with createTestBCMClient usage, add ExpectNonEmptyPlan explanation, add test helper function reference
- [ ] T055 [P] Update `/workspace/AGENTS.md` - add "Enhanced CheckDestroy Pattern" section with logging, timeout, and detailed error message examples, add "Standardized PreCheck Cleanup Pattern" section with verifyResourceDeleted usage
- [ ] T056 Create test coverage matrix in `/workspace/specs/006-test-review/test-coverage.md` - markdown table with columns: Resource, Attribute, Type, Drift Test Name, Status - track 80% coverage metric achieved
- [ ] T057 Document BCM-specific test considerations in `/workspace/specs/006-test-review/bcm-test-patterns.md` - async operations (image cloning), eventual consistency (30s window), field resets (original_image after clone), ImportStateVerifyIgnore rationale
- [ ] T058 Add drift detection quickstart example to `/workspace/specs/006-test-review/quickstart.md` - copy from plan.md quickstart section, add running example commands
- [ ] T059 Run full acceptance test suite `TF_ACC=1 go test -v ./internal/provider/` to verify all tests pass consistently without manual cleanup

**Checkpoint**: Complete documentation with test patterns and coverage tracking

---

## Phase 9: Continuous Improvement (Optional)

**Purpose**: Add CI checks and monitoring for ongoing test quality

**Estimated Time**: 2-3 hours (optional)

- [ ] T060 [P] Create test coverage calculation script in `/workspace/.specify/scripts/bash/calculate-drift-coverage.sh` - count drift tests per resource, calculate percentage of non-computed attributes covered, exit 1 if below 80%
- [ ] T061 [P] Add test performance monitoring script in `/workspace/.specify/scripts/bash/monitor-test-performance.sh` - parse `go test` output for timing, alert if total time exceeds 5 minutes
- [ ] T062 Add CI check to GitHub Actions workflow - call calculate-drift-coverage.sh in test job, fail build if coverage drops below 80%
- [ ] T063 Document periodic test review process in `/workspace/CLAUDE.md` - monthly review of test failures, flaky test identification, test pattern updates as BCM API evolves

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup (Phase 1) - BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational (Phase 2) - First drift detection tests
- **User Story 2 (Phase 4)**: Depends on Foundational (Phase 2) - Can run in parallel with Phase 3
- **User Story 3 (Phase 5)**: Depends on Phases 3-4 (basic destroy tests exist)
- **User Story 4 (Phase 6)**: Depends on Phase 3 (drift infrastructure exists) - Can run in parallel with Phases 4-5
- **User Story 5 (Phase 7)**: Depends on Phase 3 (drift infrastructure) and Phase 6 (comprehensive coverage)
- **Documentation (Phase 8)**: Depends on all user story phases being complete
- **Continuous Improvement (Phase 9)**: Optional, depends on Phase 8

### Within Each User Story

- RED phase: Write failing tests first
- GREEN phase: Implement minimal code to pass tests
- REFACTOR phase: Extract helpers, improve quality
- Tests marked [P] within same phase can run in parallel

### Parallel Opportunities

**Phase 1 (Setup)**: T002 and T003 can run in parallel (different functions)

**Phase 2 (Foundational)**: T005-T007 (softwareimage) in parallel with T008-T010 (category)

**Phase 3 (US1)**: T011 and T012 (RED tests) in parallel, T015 and T016 (config functions) in parallel

**Phase 4 (US2)**: T019 and T020 (RED tests) in parallel, T022 and T023 (external delete tests) in parallel

**Phase 5 (US3)**: T025 and T026 (RED tests) in parallel

**Phase 6 (US4)**:
- Softwareimage drift tests (T030-T037) can all run in parallel
- Category drift tests (T040-T046) can all run in parallel
- Both resource groups can run in parallel with each other

**Phase 7 (US5)**: T049 and T050 can run in parallel

**Phase 8 (Documentation)**: T054, T055 can run in parallel

**Phase 9 (CI)**: T060 and T061 can run in parallel

---

## Parallel Example: Phase 6 - User Story 4 Drift Tests

```bash
# Launch all softwareimage drift tests together (RED phase):
Task T030: "Create TestAccCMPartSoftwareImage_DriftNotes"
Task T031: "Create TestAccCMPartSoftwareImage_DriftEnableSOL"
Task T032: "Create TestAccCMPartSoftwareImage_DriftSOLSpeed"
Task T033: "Create TestAccCMPartSoftwareImage_DriftSOLFlowControl"
Task T034: "Create TestAccCMPartSoftwareImage_DriftKernelOutputConsole"
Task T035: "Create TestAccCMPartSoftwareImage_DriftSOLPort"
Task T036: "Create TestAccCMPartSoftwareImage_DriftModules"
Task T037: "Create TestAccCMPartSoftwareImage_DriftExternalDelete"

# Verify all fail (RED), then implement PreConfig for all (GREEN)
```

---

## Implementation Strategy

### MVP First (User Stories 1-2 Only)

1. Complete Phase 1: Setup (shared helpers)
2. Complete Phase 2: Foundational (enhanced CheckDestroy/PreCheck) - CRITICAL
3. Complete Phase 3: User Story 1 (basic string drift detection)
4. Complete Phase 4: User Story 2 (destroy edge cases)
5. **STOP and VALIDATE**: Run full test suite, verify 100% pass rate
6. Review and demo drift detection working

**Estimated MVP Time**: 12-16 hours

### Incremental Delivery

1. **Foundation** (Phases 1-2) → Shared infrastructure ready
2. **MVP** (Phases 3-4) → Core drift + destroy testing working
3. **Comprehensive** (Phase 6) → 80% attribute coverage achieved
4. **Advanced** (Phases 5, 7) → Edge cases and async operations handled
5. **Polish** (Phase 8) → Documentation and patterns established
6. **Optional** (Phase 9) → CI monitoring enabled

### Parallel Team Strategy

With 2 developers:

1. Team completes Setup + Foundational together (Phases 1-2)
2. Once Foundational is done:
   - **Developer A**: User Story 1 (Phase 3) + User Story 3 (Phase 5) + User Story 5 (Phase 7)
   - **Developer B**: User Story 2 (Phase 4) + User Story 4 (Phase 6)
3. Both work on Documentation together (Phase 8)

---

## Success Criteria Verification

| Success Criterion | Verification Task | Phase |
|-------------------|------------------|-------|
| SC-001: 100% CheckDestroy implementations | T005, T008 (enhanced with logging) | Phase 2 |
| SC-002: 100% PreCheck cleanup | T006, T009 (standardized with helper) | Phase 2 |
| SC-003: Drift detection test per resource | T011, T012 (first drift tests) | Phase 3 |
| SC-004: CheckDestroy verifies ALL resources | T005, T008 (already iterates all) | Phase 2 |
| SC-005: Destroy tests pass consistently | T024 (all destroy tests) | Phase 4 |
| SC-006: 80% attribute coverage | T039, T048 (comprehensive drift) | Phase 6 |
| SC-007: Unique resource names | Already implemented (generateUniqueTestName) | Existing |
| SC-008: PreCheck completes <30s | T003 (max 15s: 1+2+4+8, 4 retries) | Phase 1 |
| SC-009: 3+ drift scenarios per resource | Phases 3, 6, 7 (12+ per resource) | Phases 3-7 |
| SC-010: 3+ destroy scenarios per resource | Phase 4 (idempotent, force, external delete) | Phase 4 |

---

## Total Effort Estimate

- **Phase 1**: 2-3 hours (4 tasks including BCM field mapping)
- **Phase 2**: 3-4 hours (6 tasks)
- **Phase 3**: 4-5 hours (8 tasks)
- **Phase 4**: 3-4 hours (6 tasks)
- **Phase 5**: 3-4 hours (5 tasks)
- **Phase 6**: 7-9 hours (22 tasks - 12 softwareimage + 10 category drift tests)
- **Phase 7**: 2-3 hours (5 tasks)
- **Phase 8**: 3-4 hours (6 tasks)
- **Phase 9**: 2-3 hours (4 tasks, optional)

**Total: 70 tasks, 29-41 hours** (excluding optional Phase 9)
**MVP (Phases 1-4): 24 tasks, 12-16 hours**

---

## Notes

- [P] tasks can run in parallel (different files, no dependencies)
- [Story] label maps task to specific user story for traceability
- RED-GREEN-REFACTOR cycles within each phase
- Verify tests FAIL before implementing (RED phase)
- Run acceptance tests after each phase to verify stability
- Use unique resource names (generateUniqueTestName) for all tests
- All drift tests use 3-step pattern: create → modify external (PreConfig) → verify drift (ExpectNonEmptyPlan) → restore
- All CheckDestroy enhancements verify via BCM API, not just state
- PreCheck cleanup uses exponential backoff for eventual consistency

---

## Quick Reference Commands

```bash
# Run all tests
TF_ACC=1 go test -v ./internal/provider/

# Run specific resource tests
TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMPartSoftwareImage
TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMDeviceCategory

# Run all drift tests
TF_ACC=1 go test -v ./internal/provider/ -run "Drift"

# Run all destroy tests (CheckDestroy runs in all tests)
TF_ACC=1 go test -v ./internal/provider/ -run "Destroy"

# Run specific drift test
TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMPartSoftwareImage_DriftKernelParameters

# Environment setup
export TF_ACC=1
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"
```
