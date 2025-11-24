# Tasks: Deletion Order Validation

**Input**: Design documents from `/workspace/specs/002-deletion-order-validation/`
**Prerequisites**: plan.md (required), spec.md (required for user stories)

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure validation

- [ ] T001 Create feature documentation structure at /workspace/specs/002-deletion-order-validation/
- [ ] T002 [P] Verify BCM client connection and authentication via /workspace/internal/provider/bcm_client.go
- [ ] T003 [P] Review existing cleanup scripts structure in /workspace/scripts/

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core dependency checking infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Research & API Pattern Discovery

- [ ] T004 Research BCM API dependency check patterns - test getNodes filtering by category UUID
- [ ] T005 Research BCM API dependency check patterns - test getCategories filtering by softwareimage name
- [ ] T006 [P] Research BCM force parameter behavior in removeCategory and removeSoftwareImage methods
- [ ] T007 [P] Research BCM error response formats for dependency violations
- [ ] T008 Determine optimal retry timing for deletion verification with exponential backoff
- [ ] T009 Document findings in /workspace/specs/002-deletion-order-validation/research.md

### Design Artifacts

- [ ] T010 Create dependency graph data model in /workspace/specs/002-deletion-order-validation/data-model.md
- [ ] T011 [P] Create API contract for check-devices-in-category in /workspace/specs/002-deletion-order-validation/contracts/check-devices-in-category.json
- [ ] T012 [P] Create API contract for check-categories-using-image in /workspace/specs/002-deletion-order-validation/contracts/check-categories-using-image.json
- [ ] T013 [P] Create API contract for dependency-error-response in /workspace/specs/002-deletion-order-validation/contracts/dependency-error-response.json
- [ ] T014 Create developer quick start guide in /workspace/specs/002-deletion-order-validation/quickstart.md

### Core Infrastructure Implementation

- [ ] T015 Create dependency check helper functions in /workspace/internal/provider/dependency_helpers.go (CheckDevicesInCategory, CheckCategoriesUsingImage)
- [ ] T016 Create error message formatting functions in /workspace/internal/provider/error_messages.go (BuildDependencyError, BuildForceDeleteionWarning)
- [ ] T017 Add unit tests for dependency helpers in /workspace/internal/provider/dependency_helpers_test.go
- [ ] T018 Add unit tests for error message formatting in /workspace/internal/provider/error_messages_test.go

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Safe Cleanup Script Execution (Priority: P1) 🎯 MVP

**Goal**: Fix cleanup scripts to delete resources in correct dependency order (Devices → Clusters → Networks → Categories → Images) to prevent BCM database corruption

**Independent Test**: Run each cleanup script against BCM cluster with test resources and verify: (1) resources deleted in correct order, (2) no BCM errors, (3) no orphaned references remain

### Tests for User Story 1 (Script Validation) ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T019 [P] [US1] Create deletion order validation test in /workspace/scripts/test-cleanup-deletion-order.sh
- [ ] T020 [P] [US1] Create dry-run mode validation test in /workspace/scripts/test-cleanup-dry-run.sh

### Implementation for User Story 1

- [ ] T021 [US1] Fix deletion order in /workspace/scripts/cleanup-basic-resources.sh (Devices → Clusters → Networks → Categories → Images)
- [ ] T022 [US1] Add dry-run mode support to /workspace/scripts/cleanup-basic-resources.sh (DRY_RUN env var)
- [ ] T023 [US1] Add health check function between deletion batches in /workspace/scripts/cleanup-basic-resources.sh
- [ ] T024 [US1] Fix deletion order in /workspace/scripts/cleanup-before-tests.sh
- [ ] T025 [US1] Add dry-run mode support to /workspace/scripts/cleanup-before-tests.sh
- [ ] T026 [US1] Add health check function to /workspace/scripts/cleanup-before-tests.sh
- [ ] T027 [US1] Fix deletion order in /workspace/scripts/cleanup-test-resources-auto.sh
- [ ] T028 [US1] Add dry-run mode support to /workspace/scripts/cleanup-test-resources-auto.sh
- [ ] T029 [US1] Add health check function to /workspace/scripts/cleanup-test-resources-auto.sh
- [ ] T030 [US1] Fix deletion order in /workspace/scripts/cleanup-test-resources-safe.sh
- [ ] T031 [US1] Add dry-run mode support to /workspace/scripts/cleanup-test-resources-safe.sh
- [ ] T032 [US1] Add health check function to /workspace/scripts/cleanup-test-resources-safe.sh
- [ ] T033 [US1] Add detailed logging showing deletion order to all cleanup scripts
- [ ] T034 [US1] Run deletion order validation test against live BCM cluster
- [ ] T035 [US1] Run dry-run mode validation test for all scripts
- [ ] T036 [US1] Verify no orphaned references remain after cleanup script execution

**Checkpoint**: At this point, User Story 1 should be fully functional - cleanup scripts delete in correct order without database corruption

---

## Phase 4: User Story 2 - Provider Delete Method Protection (Priority: P2)

**Goal**: Add pre-deletion dependency validation to Category and SoftwareImage resources to prevent orphaned references during Terraform operations

**Independent Test**: Attempt to delete resources with dependencies via Terraform and verify: (1) deletion blocked when dependencies exist, (2) clear error message with resolution options, (3) force=true bypasses check

### Tests for User Story 2 (RED Phase - Write Failing Tests First) ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T037 [P] [US2] RED: Write failing test TestAccCMDeviceCategory_DeleteWithDependencies in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T038 [P] [US2] RED: Write failing test TestAccCMDeviceCategory_DeleteWithForce in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T039 [P] [US2] RED: Write failing test TestAccCMDeviceCategory_DeleteNoDependencies in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T040 [P] [US2] RED: Write failing test TestAccCMPartSoftwareImage_DeleteWithDependencies in /workspace/internal/provider/resource_cmpart_softwareimage_test.go
- [ ] T041 [P] [US2] RED: Write failing test TestAccCMPartSoftwareImage_DeleteWithForce in /workspace/internal/provider/resource_cmpart_softwareimage_test.go
- [ ] T042 [P] [US2] RED: Write failing test TestAccCMPartSoftwareImage_DeleteNoDependencies in /workspace/internal/provider/resource_cmpart_softwareimage_test.go
- [ ] T043 [US2] Run acceptance tests to verify they FAIL (TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "DeleteWith")

### Implementation for User Story 2 (GREEN Phase - Minimal Implementation)

- [ ] T044 [US2] GREEN: Add force parameter to Category schema in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T045 [US2] GREEN: Add dependency check to Category Delete method in /workspace/internal/provider/resource_cmdevice_category.go (calls CheckDevicesInCategory unless force=true)
- [ ] T046 [US2] GREEN: Add force deletion warning logging to Category Delete method in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T047 [US2] GREEN: Add force parameter to SoftwareImage schema in /workspace/internal/provider/resource_cmpart_softwareimage.go
- [ ] T048 [US2] GREEN: Add dependency check to SoftwareImage Delete method in /workspace/internal/provider/resource_cmpart_softwareimage.go (calls CheckCategoriesUsingImage unless force=true)
- [ ] T049 [US2] GREEN: Add force deletion warning logging to SoftwareImage Delete method in /workspace/internal/provider/resource_cmpart_softwareimage.go
- [ ] T050 [US2] Run acceptance tests to verify they PASS (TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "DeleteWith")

### Refactor Phase for User Story 2

- [ ] T051 [US2] REFACTOR: Extract common deletion validation pattern to helper function
- [ ] T052 [US2] REFACTOR: Improve error message formatting for readability
- [ ] T053 [US2] REFACTOR: Add comprehensive logging for dependency check operations
- [ ] T054 [US2] REFACTOR: Optimize BCM API query performance with timeout configuration
- [ ] T055 [US2] Run acceptance tests to verify refactoring didn't break functionality
- [ ] T056 [US2] Update examples in /workspace/examples/resources/bcm_cmdevice_category/resource.tf to show force parameter usage
- [ ] T057 [US2] Update examples in /workspace/examples/resources/bcm_cmpart_softwareimage/resource.tf to show force parameter usage

**Checkpoint**: At this point, User Stories 1 AND 2 should both work - cleanup scripts work correctly AND provider blocks unsafe deletions

---

## Phase 5: User Story 3 - Test Infrastructure Reliability (Priority: P3)

**Goal**: Enhance test CheckDestroy functions to clean up resources in correct dependency order to prevent test flakiness and orphaned test resources

**Independent Test**: Run acceptance tests and verify: (1) CheckDestroy deletes in correct order, (2) all test resources cleaned up, (3) no orphaned references after tests

### Tests for User Story 3 (RED Phase) ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T058 [P] [US3] RED: Write failing test TestAccMultipleResources_CheckDestroyOrder in /workspace/internal/provider/resource_integration_test.go
- [ ] T059 [US3] Run test to verify it FAILS with current CheckDestroy implementation

### Implementation for User Story 3 (GREEN Phase)

- [ ] T060 [US3] GREEN: Add GroupResourcesByType helper function to /workspace/internal/provider/test_helpers.go
- [ ] T061 [US3] GREEN: Add VerifyResourcesDeleted helper with ordered deletion to /workspace/internal/provider/test_helpers.go
- [ ] T062 [US3] GREEN: Create TestAccCheckResourcesDestroyOrdered function in /workspace/internal/provider/test_helpers.go (deletes: Devices → Clusters → Networks → Categories → Images)
- [ ] T063 [US3] GREEN: Add exponential backoff retry logic to deletion verification in /workspace/internal/provider/test_helpers.go
- [ ] T064 [US3] GREEN: Enhance CheckDestroy functions to use ordered cleanup pattern (update existing *_test.go files as needed)
- [ ] T065 [US3] Run acceptance test to verify TestAccMultipleResources_CheckDestroyOrder PASSES
- [ ] T066 [US3] Run full acceptance test suite to verify CheckDestroy improvements work (TF_ACC=1 go test -v -timeout 120m ./internal/provider/)

### Refactor Phase for User Story 3

- [ ] T067 [US3] REFACTOR: Add detailed logging to CheckDestroy functions showing deletion order
- [ ] T068 [US3] REFACTOR: Standardize error messages across all CheckDestroy functions
- [ ] T069 [US3] REFACTOR: Extract deletion order constants to shared location
- [ ] T070 [US3] Run full test suite to verify refactoring didn't break functionality

**Checkpoint**: All user stories should now be independently functional - scripts, provider, and tests all respect deletion order

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories and final validation

### Documentation

- [ ] T071 [P] Update /workspace/CLAUDE.md with deletion order requirements and dependency graph documentation
- [ ] T072 [P] Update /workspace/README.md with dependency management section and troubleshooting guide
- [ ] T073 [P] Create troubleshooting guide for deletion errors in /workspace/specs/002-deletion-order-validation/TROUBLESHOOTING.md
- [ ] T074 [P] Document force parameter implications and best practices in /workspace/specs/002-deletion-order-validation/quickstart.md

### Code Quality & Validation

- [ ] T075 Run make fmt to format all Go code
- [ ] T076 Run make lint to check for code quality issues
- [ ] T077 Fix any linting errors identified
- [ ] T078 Run make test to execute unit tests
- [ ] T079 Run make testacc to execute full acceptance test suite
- [ ] T080 Fix any failing tests identified

### Final Integration Testing

- [ ] T081 Test cleanup scripts against live BCM cluster with realistic test data
- [ ] T082 Validate examples via /workspace/scripts/test-examples.sh --verbose
- [ ] T083 Verify BCM cluster health after cleanup script execution
- [ ] T084 Test dependency validation with real user workflows (create category with devices, attempt delete)
- [ ] T085 Verify force deletion behavior against live BCM cluster (check for orphaned references)

### Documentation Generation

- [ ] T086 Run make generate to regenerate provider documentation
- [ ] T087 Review generated documentation in /workspace/docs/ for accuracy
- [ ] T088 Verify all force parameter documentation is clear and includes warnings

### Security & Performance Review

- [ ] T089 Review dependency check query performance with large datasets (>100 resources)
- [ ] T090 Verify timeout handling for slow BCM API responses
- [ ] T091 Review error messages for information disclosure concerns
- [ ] T092 Validate that force parameter cannot be accidentally enabled

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
  - Research tasks (T004-T009) must complete before design artifacts
  - Design artifacts (T010-T014) must complete before core infrastructure
  - Core infrastructure (T015-T018) must complete before any user story
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion (T018)
  - User Story 1 (Cleanup Scripts): Can start after T018
  - User Story 2 (Provider Delete): Can start after T018
  - User Story 3 (Test Infrastructure): Can start after T018 (independent of US1/US2)
- **Polish (Phase 6)**: Depends on completion of all user stories (T070)

### User Story Dependencies

- **User Story 1 (P1 - Cleanup Scripts)**: Can start after Foundational - No dependencies on other stories
- **User Story 2 (P2 - Provider Delete)**: Can start after Foundational - No dependencies on other stories
- **User Story 3 (P3 - Test Infrastructure)**: Can start after Foundational - No dependencies on other stories

**KEY INSIGHT**: All three user stories are independent and can be implemented in parallel once Foundational phase completes

### Within Each User Story

**User Story 1 (Cleanup Scripts)**:
- Tests (T019-T020) written first, should FAIL
- Script fixes happen in order (T021-T036)
- Validation tests run at end

**User Story 2 (Provider Delete)**:
- RED: Tests (T037-T043) written first, MUST FAIL
- GREEN: Minimal implementation (T044-T050) to make tests pass
- REFACTOR: Code improvements (T051-T057) while keeping tests green

**User Story 3 (Test Infrastructure)**:
- RED: Tests (T058-T059) written first, MUST FAIL
- GREEN: Implementation (T060-T066) to make tests pass
- REFACTOR: Improvements (T067-T070) while keeping tests green

### Parallel Opportunities

**Within Phase 2 (Foundational)**:
- Research tasks T006-T007 can run in parallel (after T004-T005)
- Design artifact tasks T011-T013 can run in parallel (after T010)

**Across User Stories (after Phase 2 complete)**:
- User Story 1, 2, and 3 can be worked on in parallel by different team members
- Within each user story:
  - Test file creation (T037-T042 for US2, for example) can run in parallel
  - Script updates (T021-T032 for US1) can run in parallel if different files

**Within Phase 6 (Polish)**:
- Documentation tasks T071-T074 can run in parallel
- Code quality tasks T075-T080 run sequentially (formatting before linting before testing)

---

## Parallel Example: User Story 2 (Provider Delete Methods)

```bash
# RED Phase - Launch all failing tests together:
Task: "Write failing test TestAccCMDeviceCategory_DeleteWithDependencies"
Task: "Write failing test TestAccCMDeviceCategory_DeleteWithForce"
Task: "Write failing test TestAccCMDeviceCategory_DeleteNoDependencies"
Task: "Write failing test TestAccCMPartSoftwareImage_DeleteWithDependencies"
Task: "Write failing test TestAccCMPartSoftwareImage_DeleteWithForce"
Task: "Write failing test TestAccCMPartSoftwareImage_DeleteNoDependencies"

# GREEN Phase - Implement both resources in parallel:
Task: "Add dependency check to Category Delete method"
Task: "Add dependency check to SoftwareImage Delete method"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only - Cleanup Scripts)

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: Foundational (T004-T018) - CRITICAL, blocks all stories
3. Complete Phase 3: User Story 1 (T019-T036)
4. **STOP and VALIDATE**: Test cleanup scripts against live BCM cluster
5. **Deploy/Integrate**: Updated cleanup scripts prevent database corruption immediately

**Rationale**: User Story 1 is highest priority (P1) and provides immediate value by fixing the most common source of database corruption (cleanup scripts run daily in CI/CD).

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready (dependency helpers, error formatting)
2. Add User Story 1 → Test cleanup scripts → Deploy (MVP! Scripts now safe)
3. Add User Story 2 → Test provider delete validation → Deploy (Terraform now safe)
4. Add User Story 3 → Test CheckDestroy improvements → Deploy (Tests now reliable)
5. Add Polish → Final validation → Deploy (Complete feature with documentation)

Each increment adds value without breaking previous work.

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together (T001-T018)
2. Once T018 complete:
   - Developer A: User Story 1 (Cleanup Scripts) - T019-T036
   - Developer B: User Story 2 (Provider Delete) - T037-T057
   - Developer C: User Story 3 (Test Infrastructure) - T058-T070
3. Stories complete independently and integrate without conflicts
4. Team reconvenes for Polish phase (T071-T092)

---

## Validation Checklist

### User Story 1 Success Criteria
- [ ] SC-001: All cleanup scripts delete resources in correct order (verified by test-cleanup-deletion-order.sh)
- [ ] SC-002: Zero database corruption incidents from cleanup script execution (manual verification)
- [ ] SC-008: All cleanup scripts support dry-run mode (verified by test-cleanup-dry-run.sh)

### User Story 2 Success Criteria
- [ ] SC-003: Provider Delete methods block 100% of deletions that would create orphaned references when force=false
- [ ] SC-004: Error messages contain specific resource identifiers and resolution steps in 100% of cases
- [ ] SC-006: Dependency check failures provide warnings rather than blocking when user can proceed with force=true

### User Story 3 Success Criteria
- [ ] SC-005: Test CheckDestroy functions successfully clean up resources in correct order with 100% success rate
- [ ] SC-007: Documentation clearly explains dependency graph and deletion order requirements

---

## Notes

- **[P] tasks** = different files, no dependencies, can run in parallel
- **[Story] label** (US1, US2, US3) maps task to specific user story for traceability
- **RED-GREEN-REFACTOR**: User Stories 2 and 3 strictly follow TDD cycles
- Each user story should be independently completable and testable
- Verify tests FAIL before implementing (RED phase)
- Verify tests PASS after implementation (GREEN phase)
- Verify tests still PASS after refactoring (REFACTOR phase)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently

**Key Risk Mitigations**:
- Performance: Dependency checks have 5-second timeout (configurable)
- Eventual consistency: Use exponential backoff (1s, 2s, 4s, 8s) up to 15s total
- Backward compatibility: Force parameter optional, defaults to false
- API failures: Dependency check failures produce warnings, not hard blocks
