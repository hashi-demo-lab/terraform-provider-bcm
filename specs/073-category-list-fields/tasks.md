# Tasks: BCM Category List Fields Persistence Investigation

**Input**: Design documents from `/workspace/specs/073-category-list-fields/`
**Prerequisites**: plan.md (complete), spec.md (complete), research.md (pending investigation)

**Tests**: This investigation includes validation tasks but no new test development. Existing tests will be verified to work correctly with the current workarounds.

**Organization**: Tasks are grouped by user story to enable independent implementation and verification of each investigation phase.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Investigation Infrastructure)

**Purpose**: Prepare investigation environment and scripts

- [ ] T001 Create evidence directory structure at `/workspace/specs/073-category-list-fields/evidence/`
- [ ] T002 [P] Verify BCM API access with test credentials (BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD)
- [ ] T003 [P] Review existing investigation scripts in `/workspace/sampleRest/` for reusable patterns

---

## Phase 2: User Story 1 - Verify API Behavior (Priority: P1)

**Goal**: Verify actual BCM API behavior for category list fields through direct API testing

**Independent Test**: Run investigation script against BCM API and capture JSON evidence showing whether fields persist after create/read cycle

### Implementation for User Story 1

- [ ] T004 [US1] Create investigation script at `/workspace/sampleRest/investigate_category_list_fields.py`
  - **Acceptance**: Script authenticates with BCM API successfully
  - **File**: `/workspace/sampleRest/investigate_category_list_fields.py`

- [ ] T005 [US1] Implement test for `staticRoutes` field persistence
  - **Acceptance**: Script creates category with `staticRoutes: [{destination: "10.0.0.0/8", gateway: "192.168.1.1", metric: 100}]`, reads back, captures persistence status
  - **File**: `/workspace/sampleRest/investigate_category_list_fields.py`

- [ ] T006 [US1] Implement test for `roles` field persistence
  - **Acceptance**: Script creates category with `roles: [{name: "test-role", childType: "ComputeRole", addServices: false}]`, reads back, captures persistence status
  - **File**: `/workspace/sampleRest/investigate_category_list_fields.py`

- [ ] T007 [US1] Implement test for `fsexports` field persistence
  - **Acceptance**: Script creates category with `fsexports` populated, reads back, captures persistence status
  - **File**: `/workspace/sampleRest/investigate_category_list_fields.py`

- [ ] T008 [US1] Implement test for `gpuSettings` field persistence
  - **Acceptance**: Script creates category with `gpuSettings: [{deviceId: "0", model: "Test GPU", computeMode: "default"}]`, reads back, captures persistence status
  - **File**: `/workspace/sampleRest/investigate_category_list_fields.py`

- [ ] T009 [US1] Implement test for `services` field persistence
  - **Acceptance**: Script creates category with `services` populated (structure TBD), reads back, captures persistence status
  - **File**: `/workspace/sampleRest/investigate_category_list_fields.py`

- [ ] T010 [US1] Implement update test for list fields
  - **Acceptance**: Script updates existing category to add items to list fields, reads back to verify update persistence
  - **File**: `/workspace/sampleRest/investigate_category_list_fields.py`

- [ ] T011 [US1] Implement cleanup and evidence generation
  - **Acceptance**: Script removes test category and generates JSON evidence file with all findings
  - **Files**: `/workspace/sampleRest/investigate_category_list_fields.py`, `/workspace/specs/073-category-list-fields/evidence/category_list_fields_test_results.json`

- [ ] T012 [US1] Execute investigation script and capture results
  - **Acceptance**: Script runs successfully, evidence JSON file created
  - **Command**: `cd /workspace/sampleRest && python3 investigate_category_list_fields.py`
  - **Output**: `/workspace/specs/073-category-list-fields/evidence/category_list_fields_test_results.json`

- [ ] T013 [US1] Update research.md with actual findings
  - **Acceptance**: All "PENDING VERIFICATION" sections updated with actual results from evidence
  - **File**: `/workspace/specs/073-category-list-fields/research.md`

**Checkpoint**: API behavior verified and documented with evidence

---

## Phase 3: User Story 2 - Document BCM Limitations (Priority: P2)

**Goal**: Update provider documentation with clear warnings about BCM persistence limitations

**Independent Test**: Verify documentation contains warnings for all 5 affected fields by checking `docs/resources/cmdevice_category.md`

### Implementation for User Story 2

- [ ] T014 [P] [US2] Add "Known Limitations" section to resource documentation
  - **Acceptance**: New section added with overview of BCM limitation for all 5 fields
  - **File**: `/workspace/docs/resources/cmdevice_category.md`

- [ ] T015 [P] [US2] Add inline warning for `static_routes` attribute
  - **Acceptance**: Attribute description includes note about BCM not persisting this field
  - **File**: `/workspace/docs/resources/cmdevice_category.md`

- [ ] T016 [P] [US2] Add inline warning for `fsexports` attribute
  - **Acceptance**: Attribute description includes note about BCM not persisting this field
  - **File**: `/workspace/docs/resources/cmdevice_category.md`

- [ ] T017 [P] [US2] Add inline warning for `roles` attribute
  - **Acceptance**: Attribute description includes note about BCM not persisting this field
  - **File**: `/workspace/docs/resources/cmdevice_category.md`

- [ ] T018 [P] [US2] Add inline warning for `gpu_settings` attribute
  - **Acceptance**: Attribute description includes note about BCM not persisting this field
  - **File**: `/workspace/docs/resources/cmdevice_category.md`

- [ ] T019 [P] [US2] Add inline warning for `services` attribute
  - **Acceptance**: Attribute description includes note about BCM not persisting this field
  - **File**: `/workspace/docs/resources/cmdevice_category.md`

- [ ] T020 [US2] Add import guidance documentation
  - **Acceptance**: Documentation explains that imported categories need re-apply for list fields
  - **File**: `/workspace/docs/resources/cmdevice_category.md`

- [ ] T021 [US2] Update CLAUDE.md with BCM category list field notes
  - **Acceptance**: "BCM-Specific Notes" section includes summary of category list field behavior with code references
  - **File**: `/workspace/CLAUDE.md`

- [ ] T022 [US2] Run documentation generation to verify changes
  - **Acceptance**: `make generate` completes successfully without errors
  - **Command**: `cd /workspace && make generate`

**Checkpoint**: All documentation updated with BCM limitation warnings

---

## Phase 4: User Story 3 - Investigate Alternative APIs (Priority: P3)

**Goal**: Probe BCM API for alternative methods to persist category list fields

**Independent Test**: Document results of API discovery in research.md with method-by-method status

### Implementation for User Story 3

- [ ] T023 [US3] Implement alternative API discovery in investigation script
  - **Acceptance**: Script probes for methods: addCategoryRole, setCategoryStaticRoutes, addCategoryFSExport, etc.
  - **File**: `/workspace/sampleRest/investigate_category_list_fields.py`

- [ ] T024 [US3] Test CMDevice service for category-specific role methods
  - **Acceptance**: Test addCategoryRole, removeCategoryRole, setCategoryRoles - document results
  - **File**: `/workspace/sampleRest/investigate_category_list_fields.py`

- [ ] T025 [US3] Test CMDevice service for category-specific route methods
  - **Acceptance**: Test addCategoryStaticRoute, removeCategoryStaticRoute, setCategoryStaticRoutes - document results
  - **File**: `/workspace/sampleRest/investigate_category_list_fields.py`

- [ ] T026 [US3] Test CMDevice service for category-specific export methods
  - **Acceptance**: Test addCategoryFSExport, removeCategoryFSExport, setCategoryFSExports - document results
  - **File**: `/workspace/sampleRest/investigate_category_list_fields.py`

- [ ] T027 [US3] Test node-level role methods as alternative
  - **Acceptance**: Test addNodeRole, removeNodeRole, getNodeRoles - document if these are the correct API for roles
  - **File**: `/workspace/sampleRest/investigate_category_list_fields.py`

- [ ] T028 [US3] Execute alternative API discovery and capture results
  - **Acceptance**: All probe results captured in evidence file
  - **Command**: `cd /workspace/sampleRest && python3 investigate_category_list_fields.py --discover-apis`
  - **Output**: `/workspace/specs/073-category-list-fields/evidence/api_discovery_results.md`

- [ ] T029 [US3] Update research.md with alternative API findings
  - **Acceptance**: Alternative API Discovery Results table populated with actual findings
  - **File**: `/workspace/specs/073-category-list-fields/research.md`

**Checkpoint**: Alternative API investigation complete with documented findings

---

## Phase 5: Validation & Verification

**Purpose**: Verify existing provider workarounds function correctly

### Idempotency Validation

- [ ] T030 [P] Verify TestAccCMDeviceCategoryResource_StaticRoutes passes with idempotency
  - **Acceptance**: Test passes with `plancheck.ExpectEmptyPlan()` after create
  - **Command**: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceCategoryResource_StaticRoutes"`

- [ ] T031 [P] Verify TestAccCMDeviceCategoryResource_FilesystemExports passes with idempotency
  - **Acceptance**: Test passes with `plancheck.ExpectEmptyPlan()` after create
  - **Command**: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceCategoryResource_FilesystemExports"`

- [ ] T032 [P] Verify TestAccCMDeviceCategoryResource_Roles passes with idempotency
  - **Acceptance**: Test passes with `plancheck.ExpectEmptyPlan()` after create
  - **Command**: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceCategoryResource_Roles"`

### Import Verification

- [ ] T033 [P] Verify ImportStateVerifyIgnore includes all 5 non-persisted fields
  - **Acceptance**: Code review confirms `static_routes`, `fsexports`, `roles`, `gpu_settings`, `services` in ignore list
  - **File**: `/workspace/internal/provider/resource_cmdevice_category_test.go`

- [ ] T034 [P] Run import test to verify no false verification failures
  - **Acceptance**: Import test passes without failures for non-persisted fields
  - **Command**: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceCategoryResource_Import"`

### Debug Logging Verification

- [ ] T035 Verify debug logging exists for value preservation
  - **Acceptance**: Code contains tflog.Debug calls when preserving plan values over BCM empty responses
  - **File**: `/workspace/internal/provider/resource_cmdevice_category.go`

- [ ] T036 Verify debug logging for role UUID generation
  - **Acceptance**: Code contains tflog.Debug calls when generating UUIDs for roles
  - **File**: `/workspace/internal/provider/resource_cmdevice_category.go`

**Checkpoint**: All validation tasks complete - workarounds verified functional

---

## Phase 6: Polish & Finalization

**Purpose**: Complete investigation and close issue

- [ ] T037 Update research.md with final conclusions and recommendations
  - **Acceptance**: Executive summary updated, all sections marked VERIFIED or updated with findings
  - **File**: `/workspace/specs/073-category-list-fields/research.md`

- [ ] T038 Create investigation summary for GitHub issue #73
  - **Acceptance**: Summary includes: findings, evidence links, documentation updates, test verification results
  - **Output**: Comment content for GitHub issue

- [ ] T039 Review all test comments for accuracy
  - **Acceptance**: Test file comments accurately reflect BCM behavior (Lines 2590, 2785, 3599, 3696)
  - **File**: `/workspace/internal/provider/resource_cmdevice_category_test.go`

- [ ] T040 Run full category resource test suite
  - **Acceptance**: All TestAccCMDeviceCategoryResource_* tests pass
  - **Command**: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceCategoryResource"`

- [ ] T041 Verify quickstart.md is complete and accurate
  - **Acceptance**: All code references valid, commands work, documentation links active
  - **File**: `/workspace/specs/073-category-list-fields/quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **User Story 1 (Phase 2)**: Depends on Setup completion - API verification is foundational
- **User Story 2 (Phase 3)**: Depends on User Story 1 - documentation should reflect verified findings
- **User Story 3 (Phase 4)**: Can run in parallel with User Story 2 after User Story 1 completes
- **Validation (Phase 5)**: Can run in parallel with User Stories 2 and 3
- **Polish (Phase 6)**: Depends on all user stories and validation completing

### User Story Dependencies

- **User Story 1 (P1)**: Setup complete - no other dependencies
- **User Story 2 (P2)**: Depends on US1 findings to document correct behavior
- **User Story 3 (P3)**: Depends on US1 completion; can run parallel to US2

### Within Each User Story

- Script creation before test implementation
- Tests for individual fields can run in parallel
- Execute script after all field tests implemented
- Update research.md after evidence captured

### Parallel Opportunities

**Phase 2 - Within US1**:
```bash
# After T004 (script creation), these can run in parallel:
T005 [US1] staticRoutes test
T006 [US1] roles test
T007 [US1] fsexports test
T008 [US1] gpuSettings test
T009 [US1] services test
```

**Phase 3 - Within US2**:
```bash
# All documentation updates can run in parallel:
T014 [P] [US2] Known Limitations section
T015 [P] [US2] static_routes warning
T016 [P] [US2] fsexports warning
T017 [P] [US2] roles warning
T018 [P] [US2] gpu_settings warning
T019 [P] [US2] services warning
```

**Phase 5 - Validation**:
```bash
# All validation tasks can run in parallel:
T030 [P] StaticRoutes idempotency
T031 [P] FilesystemExports idempotency
T032 [P] Roles idempotency
T033 [P] ImportStateVerifyIgnore review
T034 [P] Import test execution
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: User Story 1 - API Verification
3. **STOP and VALIDATE**: Review evidence, confirm BCM behavior
4. Proceed with US2/US3 based on findings

### Recommended Order

1. Setup (T001-T003)
2. User Story 1: API Verification (T004-T013)
3. In parallel:
   - User Story 2: Documentation (T014-T022)
   - User Story 3: Alternative API Discovery (T023-T029)
   - Validation (T030-T036)
4. Polish & Finalization (T037-T041)

### Success Criteria Mapping

| Criteria | Task(s) |
|----------|---------|
| SC-001: Investigation script runs successfully | T012 |
| SC-002: Acceptance tests pass with idempotency | T030, T031, T032 |
| SC-003: Documentation includes warnings | T014-T020 |
| SC-004: Import tests have correct ignore lists | T033, T034 |
| SC-005: Debug logging present | T035, T036 |

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Investigation is evidence-based - findings may change recommendations
- Existing workarounds assumed correct unless evidence contradicts
- Commit after each logical group of tasks
- Stop at any checkpoint to validate findings independently
