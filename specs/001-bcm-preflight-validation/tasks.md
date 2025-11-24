---
description: "Task list for BCM Pre-flight Validation Feature"
---

# Tasks: BCM Pre-flight Validation

**Input**: Design documents from `/workspace/specs/001-bcm-preflight-validation/`
**Prerequisites**: plan.md (complete), spec.md (complete)
**Branch**: `001-bcm-preflight-validation`
**Tests**: TDD approach - all acceptance tests written BEFORE implementation

**Organization**: Tasks organized by user story to enable independent implementation and testing. Following RED-GREEN-REFACTOR TDD cycle.

## Format: `- [ ] [ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US5)
- All file paths are absolute
- All tasks follow TDD RED-GREEN-REFACTOR pattern

---

## Phase 1: Setup (Project Initialization)

**Purpose**: Ensure branch is ready and baseline tests pass

- [ ] T001 Verify checkout of branch `001-bcm-preflight-validation` from main
- [ ] T002 Run baseline acceptance tests to confirm no regressions: `make testacc`
- [ ] T003 [P] Export test environment variables (BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD)
- [ ] T003a Verify integration point line numbers by searching for "addSoftwareImage", "updateSoftwareImage", "addCategory", "updateCategory", "addDevice", "updateDevice", "addKubeCluster", "updateKubeCluster" in resource files and update Service Name Reference table with actual line numbers

**Checkpoint**: Branch ready, baseline established, integration points verified

---

## Phase 2: Foundational (Core Validation Infrastructure)

**Purpose**: Core validation helper and data structures that ALL user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### RED Phase - Failing Unit Tests for ValidateEntity()

- [ ] T004 Create `/workspace/internal/provider/bcm_client_test.go` with failing TestValidateEntity_Success unit test
- [ ] T005 [P] Add failing TestValidateEntity_ErrorResponse unit test to bcm_client_test.go
- [ ] T006 [P] Add failing TestValidateEntity_WarningResponse unit test to bcm_client_test.go
- [ ] T007 [P] Add failing TestValidateEntity_ZeroUUIDFiltering unit test to bcm_client_test.go
- [ ] T008 [P] Add failing TestValidateEntity_MalformedResponse unit test to bcm_client_test.go
- [ ] T009 Run unit tests to confirm all FAIL: `go test -v ./internal/provider/ -run TestValidateEntity`

### GREEN Phase - Implement ValidateEntity() Helper

- [ ] T010 Add ValidationError struct to `/workspace/internal/provider/bcm_client.go` with fields: Field, Message, ErrorCode, Severity, EntityUUID
- [ ] T011 Add IsError() method to ValidationError struct (returns true for ERROR severity or unknown)
- [ ] T012 Add IsWarning() method to ValidationError struct (returns true for WARNING severity)
- [ ] T013 Implement ValidateEntity() helper function in bcm_client.go with signature: `func (c *BCMClient) ValidateEntity(ctx context.Context, service, validateMethod string, entity map[string]interface{}, isCreate bool) ([]ValidationError, error)`
- [ ] T014 Add validation response parsing logic using null-safe getString() helpers (reference: getStringValue pattern from data_source_cmpart_softwareimages.go:399-431)
- [ ] T015 Implement Zero UUID filtering logic (filter when isCreate==true AND Field=="uuid" AND Message contains "Zero UUID")
- [ ] T016 Add debug logging for validation calls and filtering in ValidateEntity()
- [ ] T017 Run unit tests to confirm all PASS: `go test -v ./internal/provider/ -run TestValidateEntity`

**Checkpoint**: Foundation ready - validation helper fully tested and working. User story implementation can now begin in parallel.

---

## Phase 3: User Story 1 - Catch Invalid Field Values Before Create (Priority: P1) 🎯 MVP

**Goal**: Validate field values (SOL speed, hostname format, required fields) before CREATE operations and return specific error messages

**Independent Test**: Create resource with invalid field value (e.g., SOL speed = 999999) and verify provider returns validation error before API call

### RED Phase - Failing Acceptance Tests for Software Image Validation

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T018 [P] [US1] Add failing TestAccCMPartSoftwareImage_ValidationErrorInvalidSOLSpeed to `/workspace/internal/provider/resource_cmpart_softwareimage_test.go`
- [ ] T019 [P] [US1] Add failing TestAccCMPartSoftwareImage_ValidationErrorDuplicateName to resource_cmpart_softwareimage_test.go
- [ ] T020 [P] [US1] Run software image validation tests to confirm FAIL: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMPartSoftwareImage_Validation"`

### GREEN Phase - Integrate Validation in Software Image Resource

- [ ] T021 [US1] Add ValidateEntity() call before addSoftwareImage in `/workspace/internal/provider/resource_cmpart_softwareimage.go` Create() method (line ~288)
- [ ] T022 [US1] Add validation error processing loop with AddError for ERROR severity and AddWarning for WARNING severity in software image Create()
- [ ] T023 [US1] Add early return if hasErrors==true after validation in software image Create()
- [ ] T024 [US1] Add ValidateEntity() call before updateSoftwareImage in resource_cmpart_softwareimage.go Update() method (line ~483) with isCreate=false
- [ ] T025 [US1] Add validation error processing in software image Update() (same pattern as Create)
- [ ] T026 [US1] Run software image validation tests to confirm PASS: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMPartSoftwareImage_Validation"`

**Checkpoint**: User Story 1 complete for software images - field validation working before CREATE/UPDATE operations

---

## Phase 4: User Story 2 - Detect Duplicate Names Before Create (Priority: P2)

**Goal**: Detect duplicate name conflicts during validation before CREATE operations

**Independent Test**: Attempt to create resource with existing name and verify "already exists" error

**Note**: This capability is tested as part of US1 tests (T019) and works automatically through BCM validateSoftwareImage API. This phase extends validation to remaining 4 resource types.

### RED Phase - Failing Acceptance Tests for Category Validation

- [ ] T027 [P] [US2] Add failing TestAccCMDeviceCategory_ValidationErrorInvalidField to `/workspace/internal/provider/resource_cmdevice_category_test.go`
- [ ] T028 [P] [US2] Add failing TestAccCMDeviceCategory_ValidationErrorDuplicateName to resource_cmdevice_category_test.go
- [ ] T029 [P] [US2] Run category validation tests to confirm FAIL: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceCategory_Validation"`

### GREEN Phase - Integrate Validation in Category Resource

- [ ] T030 [US2] Add ValidateEntity() call before addCategory in `/workspace/internal/provider/resource_cmdevice_category.go` Create() method with service="CMDevice" and validateMethod="validateCategory"
- [ ] T031 [US2] Add validation error processing loop in category Create() (same pattern as software image)
- [ ] T032 [US2] Add ValidateEntity() call before updateCategory in resource_cmdevice_category.go Update() method with isCreate=false
- [ ] T033 [US2] Add validation error processing in category Update()
- [ ] T034 [US2] Run category validation tests to confirm PASS: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceCategory_Validation"`

**Checkpoint**: User Story 2 complete for categories - duplicate name detection working

---

## Phase 5: User Story 4 - Validate Updates Before Modification (Priority: P2)

**Goal**: Validate updated configurations before UPDATE operations

**Independent Test**: Update resource with invalid field values and verify validation errors before update API call

**Note**: UPDATE validation was implemented alongside CREATE in previous phases. This phase extends to remaining 3 resource types.

### RED Phase - Failing Acceptance Tests for Device Validation

- [ ] T035 [P] [US4] Add failing TestAccCMDeviceDevice_ValidationErrorInvalidHostname to `/workspace/internal/provider/resource_cmdevice_device_test.go`
- [ ] T036 [P] [US4] Add failing TestAccCMDeviceDevice_ValidationUpdate to resource_cmdevice_device_test.go (test UPDATE validation)
- [ ] T037 [P] [US4] Run device validation tests to confirm FAIL: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceDevice_Validation"`

### GREEN Phase - Integrate Validation in Device Resource

- [ ] T038 [US4] Add ValidateEntity() call before addDevice in `/workspace/internal/provider/resource_cmdevice_device.go` Create() method with service="CMDevice" and validateMethod="validateDevice"
- [ ] T039 [US4] Add validation error processing loop in device Create()
- [ ] T040 [US4] Add ValidateEntity() call before updateDevice in resource_cmdevice_device.go Update() method (line ~783) with isCreate=false
- [ ] T041 [US4] Add validation error processing in device Update()
- [ ] T042 [US4] Run device validation tests to confirm PASS: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceDevice_Validation"`

**Checkpoint**: User Story 4 complete for devices - CREATE and UPDATE validation working

---

## Phase 6: User Story 5 - Consistent Validation Across All Resource Types (Priority: P2)

**Goal**: All 5 resource types provide consistent validation behavior with same error format and severity handling

**Independent Test**: Trigger validation errors on each resource type and verify identical error format

### RED Phase - Failing Acceptance Tests for Network Validation

- [ ] T043 [P] [US5] Add failing TestAccCMNetNetwork_ValidationErrorInvalidField to `/workspace/internal/provider/resource_cmnet_network_test.go`
- [ ] T044 [P] [US5] Add failing TestAccCMNetNetwork_ValidationErrorDuplicateName to resource_cmnet_network_test.go
- [ ] T045 [P] [US5] Run network validation tests to confirm FAIL: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMNetNetwork_Validation"`

### GREEN Phase - Integrate Validation in Network Resource

- [ ] T046 [US5] Add ValidateEntity() call before network create in `/workspace/internal/provider/resource_cmnet_network.go` Create() method with service="CMNet" and validateMethod="validateNetwork"
- [ ] T047 [US5] Add validation error processing loop in network Create()
- [ ] T048 [US5] Add ValidateEntity() call before network update in resource_cmnet_network.go Update() method with isCreate=false
- [ ] T049 [US5] Add validation error processing in network Update()
- [ ] T050 [US5] Run network validation tests to confirm PASS: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMNetNetwork_Validation"`

### RED Phase - Failing Acceptance Tests for Kubernetes Cluster Validation

- [ ] T051 [P] [US5] Add failing TestAccCMKubeCluster_ValidationErrorInvalidField to `/workspace/internal/provider/resource_cmkube_cluster_test.go`
- [ ] T052 [P] [US5] Add failing TestAccCMKubeCluster_ValidationErrorDuplicateName to resource_cmkube_cluster_test.go
- [ ] T053 [P] [US5] Run kube cluster validation tests to confirm FAIL: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMKubeCluster_Validation"`

### GREEN Phase - Integrate Validation in Kubernetes Cluster Resource

- [ ] T054 [US5] Add ValidateEntity() call before addKubeCluster in `/workspace/internal/provider/resource_cmkube_cluster.go` Create() method (line ~272) with service="cmkube" (LOWERCASE) and validateMethod="validateKubeCluster"
- [ ] T055 [US5] Add validation error processing loop in kube cluster Create()
- [ ] T056 [US5] Add ValidateEntity() call before updateKubeCluster in resource_cmkube_cluster.go Update() method (line ~569) with isCreate=false
- [ ] T057 [US5] Add validation error processing in kube cluster Update()
- [ ] T058 [US5] Run kube cluster validation tests to confirm PASS: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMKubeCluster_Validation"`

**Checkpoint**: User Story 5 complete - all 5 resource types have consistent validation

---

## Phase 7: User Story 3 - Receive Advisory Warnings for Suspicious Values (Priority: P3)

**Goal**: Display WARNING severity validation messages but allow operations to proceed

**Independent Test**: Create resource with non-existent path and verify warning displayed but operation completes

### RED Phase - Failing Acceptance Tests for Warning Handling

- [ ] T059 [P] [US3] Add failing TestAccCMPartSoftwareImage_ValidationWarning to `/workspace/internal/provider/resource_cmpart_softwareimage_test.go` (test WARNING severity allows operation)
- [ ] T060 [P] [US3] Add failing TestAccCMDeviceCategory_ValidationWarning to `/workspace/internal/provider/resource_cmdevice_category_test.go`
- [ ] T061 [P] [US3] Run warning validation tests to confirm FAIL: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "ValidationWarning"`

### GREEN Phase - Verify Warning Handling Works

**Note**: WARNING handling is already implemented in validation error processing loops. This phase adds acceptance tests to verify behavior.

- [ ] T062 [US3] Verify WARNING severity tests PASS (no code changes needed if previous implementation correct): `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "ValidationWarning"`
- [ ] T063 [US3] If tests fail, debug validation error processing to ensure WARNING severity uses AddWarning() not AddError()

**Checkpoint**: User Story 3 complete - WARNING severity handling verified across resources

---

## Phase 8: Refactor & Code Quality

**Purpose**: Improve code quality while keeping all tests green

- [ ] T064 [P] Extract validation error processing into helper function validateAndProcessErrors() in bcm_client.go to reduce duplication across 5 resources
- [ ] T065 [P] Add comprehensive logging at DEBUG level for all validation calls (service, method, entity name, result count)
- [ ] T066 [P] Review error messages for consistency and clarity (ensure all follow "Validation Error: [field] - [message]" format)
- [ ] T067 [P] Add code comments documenting ValidateEntity() parameters and Zero UUID filtering logic
- [ ] T068 Run all acceptance tests to confirm no regressions after refactoring: `make testacc`

**Checkpoint**: Code quality improved, all tests still passing

---

## Phase 9: Performance & Edge Case Validation

**Purpose**: Verify performance targets and edge case handling

- [ ] T069 [P] Add performance logging to measure validation overhead (log duration of ValidateEntity calls)
- [ ] T070 [P] Run acceptance tests with TF_LOG=TRACE, manually review logs for validation call duration, verify each ValidateEntity call completes within 200ms threshold (manual verification - BCM API response times vary by environment)
- [ ] T071 [P] Add TestValidateEntity_UnknownSeverity unit test to verify unknown severity treated as ERROR
- [ ] T072 [P] Add TestValidateEntity_EmptyResponse unit test to verify empty array handled correctly
- [ ] T073 [P] Add TestValidateEntity_MultipleErrors unit test to verify multiple validation errors displayed
- [ ] T074 Run all unit tests to confirm edge cases handled: `go test -v ./internal/provider/ -run TestValidateEntity`

**Checkpoint**: Performance validated, edge cases covered

---

## Phase 10: Documentation & Examples

**Purpose**: Update provider documentation and examples

- [ ] T075 [P] Update `/workspace/CLAUDE.md` with validation pattern section (ValidateEntity usage, integration points, service name casing)
- [ ] T076 [P] Create example validation error scenario in `/workspace/examples/resources/bcm_cmpart_softwareimage/softwareimage_validation_error.tf`
- [ ] T077 [P] Create example validation error scenario in `/workspace/examples/resources/bcm_cmdevice_category/category_validation_error.tf`
- [ ] T078 [P] Create example validation error scenario in `/workspace/examples/resources/bcm_cmdevice_device/device_validation_error.tf`
- [ ] T079 [P] Create example validation error scenario in `/workspace/examples/resources/bcm_cmnet_network/network_validation_error.tf`
- [ ] T080 [P] Create example validation error scenario in `/workspace/examples/resources/bcm_cmkube_cluster/kubecluster_validation_error.tf`
- [ ] T081 Run `make generate` to update terraform provider documentation in `/workspace/docs/`
- [ ] T082 Verify generated documentation includes validation error examples and behavior

**Checkpoint**: Documentation complete and up-to-date

---

## Phase 11: Final Validation & Cleanup

**Purpose**: Comprehensive testing and verification before merge

- [ ] T083 Run full acceptance test suite to confirm all tests pass: `TF_ACC=1 BCM_ENDPOINT="https://172.21.15.254:8081" BCM_USERNAME="root" BCM_PASSWORD="Hashicorp123!" go test -v -timeout 120m ./internal/provider/`
- [ ] T084 Run unit test suite: `go test -v -cover ./internal/provider/`
- [ ] T085 Run code formatting: `make fmt`
- [ ] T086 Run linting: `make lint`
- [ ] T087 Verify no hardcoded test values or credentials in code
- [ ] T088 Review TF_LOG=TRACE output for all 5 resource types to verify correct service name casing (CMPart, CMDevice, CMNet, cmkube)
- [ ] T089 Test example validation error scenarios manually to confirm user-facing error messages are clear
- [ ] T090 Create summary report of validation integration: resource count, test count, performance measurements

**Checkpoint**: Feature complete, all tests passing, ready for code review

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phases 3-7)**: All depend on Foundational phase completion
  - Phase 3 (US1): Software Image validation
  - Phase 4 (US2): Category validation (extends duplicate detection)
  - Phase 5 (US4): Device validation (extends UPDATE validation)
  - Phase 6 (US5): Network + Kube Cluster validation (completes coverage)
  - Phase 7 (US3): Warning handling verification
- **Refactor (Phase 8)**: Depends on all user stories complete
- **Performance (Phase 9)**: Can run parallel with documentation
- **Documentation (Phase 10)**: Depends on implementation complete
- **Final (Phase 11)**: Depends on all phases complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after US1 - Extends duplicate detection pattern
- **User Story 4 (P2)**: Can start after US1 - Extends UPDATE validation pattern
- **User Story 5 (P2)**: Can start after US1, US2, US4 - Completes consistency across all resources
- **User Story 3 (P3)**: Can start after implementation complete - Verifies WARNING handling

### Within Each User Story Phase

- RED tests MUST be written and FAIL before GREEN implementation
- GREEN implementation proceeds sequentially:
  1. Add ValidateEntity() call in Create()
  2. Add validation error processing in Create()
  3. Add ValidateEntity() call in Update()
  4. Add validation error processing in Update()
  5. Verify tests PASS
- Each resource type completes independently before moving to next

### Parallel Opportunities

**Phase 2 (Foundational) - RED phase**:
- T005, T006, T007, T008 (unit tests) can run in parallel

**Phase 3 (US1) - RED phase**:
- T018, T019 (acceptance tests) can run in parallel

**Phase 4 (US2) - RED phase**:
- T027, T028 (acceptance tests) can run in parallel

**Phase 5 (US4) - RED phase**:
- T035, T036 (acceptance tests) can run in parallel

**Phase 6 (US5) - RED phases**:
- T043, T044 (network tests) can run in parallel
- T051, T052 (kube cluster tests) can run in parallel

**Phase 7 (US3) - RED phase**:
- T059, T060 (warning tests) can run in parallel

**Phase 8 (Refactor)**:
- T064, T065, T066, T067 can run in parallel (different concerns)

**Phase 9 (Performance)**:
- T069, T070, T071, T072, T073 can run in parallel

**Phase 10 (Documentation)**:
- T075, T076, T077, T078, T079, T080 can run in parallel (different files)

---

## Parallel Execution Example: Phase 2 RED

```bash
# Launch all unit tests together for Foundational phase:
# Terminal 1:
go test -v ./internal/provider/ -run TestValidateEntity_Success

# Terminal 2:
go test -v ./internal/provider/ -run TestValidateEntity_ErrorResponse

# Terminal 3:
go test -v ./internal/provider/ -run TestValidateEntity_WarningResponse

# Terminal 4:
go test -v ./internal/provider/ -run TestValidateEntity_ZeroUUIDFiltering
```

---

## Parallel Execution Example: Phase 6 GREEN

```bash
# After Foundational phase complete, integrate validation in parallel:
# Developer A: Network resource (T046-T050)
# Developer B: Kubernetes cluster resource (T054-T058)
# Both can work simultaneously on different files
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (Software Image validation)
4. **STOP and VALIDATE**: Test software image validation independently
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 (Software Images) → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 (Categories) → Test independently → Deploy/Demo
4. Add User Story 4 (Devices) → Test independently → Deploy/Demo
5. Add User Story 5 (Networks + Kube Clusters) → Test independently → Deploy/Demo
6. Add User Story 3 (Warning verification) → Test independently → Deploy/Demo
7. Each story adds validation coverage without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (Software Images) - Phase 3
   - Developer B: User Story 2 (Categories) - Phase 4
   - Developer C: User Story 4 (Devices) - Phase 5
3. After US1-US4 complete:
   - Developer A: Network validation (Phase 6 first half)
   - Developer B: Kube Cluster validation (Phase 6 second half)
4. Stories complete and integrate independently

---

## Service Name and Method Reference

**CRITICAL**: Use correct service name casing to avoid API errors

| Resource Type | Resource File | Service Name | Validation Method | Create Line | Update Line |
|---------------|---------------|--------------|-------------------|-------------|-------------|
| Software Images | resource_cmpart_softwareimage.go | CMPart | validateSoftwareImage | ~288 | ~483 |
| Categories | resource_cmdevice_category.go | CMDevice | validateCategory | N/A | N/A |
| Devices | resource_cmdevice_device.go | CMDevice | validateDevice | N/A | ~783 |
| Networks | resource_cmnet_network.go | CMNet | validateNetwork | N/A | N/A |
| Kubernetes Clusters | resource_cmkube_cluster.go | **cmkube** (lowercase!) | validateKubeCluster | ~272 | ~569 |

**Note**: cmkube uses lowercase service name (exception to CamelCase pattern)

---

## Validation Pattern Reference

### ValidateEntity() Call Pattern

```go
// Build API entity first
entity := buildResourceAPIEntity(ctx, data)

// For UPDATE: Add UUID
if isUpdate {
    entity["uuid"] = data.UUID.ValueString()
}

// Call validation
validationErrors, err := r.client.ValidateEntity(
    ctx,
    "ServiceName",      // CMPart, CMDevice, CMNet, or cmkube
    "validateMethod",   // e.g., validateSoftwareImage
    entity,
    isCreate,          // true for CREATE, false for UPDATE
)

// Handle API errors
if err != nil {
    resp.Diagnostics.AddError(
        "Validation API Error",
        fmt.Sprintf("Could not validate resource: %s", err.Error()),
    )
    return
}

// Process validation results
hasErrors := false
for _, valErr := range validationErrors {
    if valErr.IsError() {
        resp.Diagnostics.AddError(
            fmt.Sprintf("Validation Error: %s", valErr.Field),
            valErr.Message,
        )
        hasErrors = true
    } else if valErr.IsWarning() {
        resp.Diagnostics.AddWarning(
            fmt.Sprintf("Validation Warning: %s", valErr.Field),
            valErr.Message,
        )
    }
}

// Halt if errors found
if hasErrors {
    return
}

// Continue with CREATE/UPDATE operation...
```

---

## Test Environment Setup

```bash
# Required environment variables for acceptance tests
export TF_ACC=1
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"

# Optional: Enable trace logging for debugging
export TF_LOG=TRACE

# Run specific resource validation tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMPartSoftwareImage_Validation"

# Run all validation tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "Validation"

# Run specific unit tests
go test -v ./internal/provider/ -run TestValidateEntity
```

---

## Success Criteria Checklist

### Functional Requirements

- [ ] ValidateEntity() helper function implemented with correct signature
- [ ] ValidationError struct with IsError() and IsWarning() methods
- [ ] All 5 resource types call validation before CREATE operations
- [ ] All 5 resource types call validation before UPDATE operations
- [ ] Zero UUID filtering works correctly for CREATE (isCreate=true)
- [ ] ERROR severity validation halts operations
- [ ] WARNING severity validation displays advisory but continues
- [ ] Correct service name casing for all resources (cmkube lowercase)

### Testing Requirements

- [ ] Unit tests for ValidateEntity() pass (5 test scenarios)
- [ ] Acceptance tests for invalid field values pass (5 resources × 1 test = 5 tests)
- [ ] Acceptance tests for duplicate name detection pass (5 resources × 1 test = 5 tests)
- [ ] Acceptance tests for WARNING severity pass (2 resources × 1 test = 2 tests)
- [ ] Acceptance tests for UPDATE validation pass (2 resources × 1 test = 2 tests)
- [ ] All existing acceptance tests still pass (no regressions)
- [ ] Total new tests: ~19 acceptance tests + 5 unit tests = 24 new tests

### Performance Requirements

- [ ] Validation adds <200ms overhead per operation (measured with TF_LOG=TRACE)
- [ ] No timeouts or performance degradation observed

### Documentation Requirements

- [ ] CLAUDE.md updated with validation pattern section
- [ ] examples/ directories updated with validation error examples (5 resources)
- [ ] `make generate` runs successfully and updates docs/
- [ ] Generated documentation includes validation behavior

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Follow TDD RED-GREEN-REFACTOR cycle strictly
- Verify tests fail (RED) before implementing (GREEN)
- Run refactoring (REFACTOR) only after tests pass
- Commit after each logical task group
- Stop at any checkpoint to validate story independently
- All file paths are absolute for clarity
- Service name casing is critical: CMPart, CMDevice, CMNet, **cmkube (lowercase)**

---

## Estimated Effort

**Total Tasks**: 90 tasks
**Estimated Time**: 2-3 days with parallel execution
**Test Count**: ~24 new tests (19 acceptance + 5 unit)
**Files Modified**: 11 files (1 new bcm_client_test.go + 1 modified bcm_client.go + 5 resource files + 5 test files)
**Files Created**: 6 example files + generated documentation

---

**Status**: ✅ READY FOR EXECUTION
**Next Step**: Execute Phase 1 (Setup) tasks T001-T003
