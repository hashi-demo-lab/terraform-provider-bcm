# BCM Pre-flight Validation - Task Generation Summary

**Generated**: 2025-11-24
**Feature**: BCM Pre-flight Validation
**Branch**: `001-bcm-preflight-validation`
**Location**: `/workspace/specs/001-bcm-preflight-validation/tasks.md`

## Overview

Successfully generated comprehensive, dependency-ordered task list for implementing BCM pre-flight validation feature across all 5 resource types (software images, categories, devices, networks, kubernetes clusters).

## Task Statistics

### Total Tasks: 90

**Distribution by Phase**:
- Phase 1 (Setup): 3 tasks
- Phase 2 (Foundational): 14 tasks (RED-GREEN pattern)
- Phase 3 (US1 - Software Images): 9 tasks
- Phase 4 (US2 - Categories): 8 tasks
- Phase 5 (US4 - Devices): 8 tasks
- Phase 6 (US5 - Networks + Kube Clusters): 16 tasks
- Phase 7 (US3 - Warning Verification): 5 tasks
- Phase 8 (Refactor): 5 tasks
- Phase 9 (Performance): 6 tasks
- Phase 10 (Documentation): 8 tasks
- Phase 11 (Final Validation): 8 tasks

### Tasks by User Story

- **Setup/Foundational**: 24 tasks (no story label)
- **US1** (Invalid Field Values): 9 tasks
- **US2** (Duplicate Names): 8 tasks
- **US3** (Advisory Warnings): 5 tasks
- **US4** (Update Validation): 8 tasks
- **US5** (Consistent Validation): 16 tasks

### Parallel Execution Opportunities

**Tasks marked [P]**: 41 tasks can run in parallel
- Foundational RED phase: 4 parallel unit tests
- US1 RED phase: 2 parallel acceptance tests
- US2 RED phase: 2 parallel acceptance tests
- US4 RED phase: 2 parallel acceptance tests
- US5 RED phase: 4 parallel acceptance tests
- Refactor phase: 4 parallel improvements
- Performance phase: 5 parallel tests
- Documentation phase: 6 parallel example files

## TDD Workflow Structure

All implementation phases follow strict **RED-GREEN-REFACTOR** cycle:

1. **RED Phase**: Write failing tests first
   - Unit tests for ValidateEntity() helper
   - Acceptance tests for each resource type
   - Verify tests FAIL before proceeding

2. **GREEN Phase**: Minimal implementation to pass tests
   - Implement ValidateEntity() helper in bcm_client.go
   - Integrate validation in each resource's Create() and Update() methods
   - Verify tests PASS

3. **REFACTOR Phase**: Improve code quality
   - Extract common patterns
   - Add comprehensive logging
   - Enhance error messages
   - Keep tests passing

## Test Coverage

### New Tests Created: ~24 tests

**Unit Tests (5)**:
- TestValidateEntity_Success
- TestValidateEntity_ErrorResponse
- TestValidateEntity_WarningResponse
- TestValidateEntity_ZeroUUIDFiltering
- TestValidateEntity_MalformedResponse

**Acceptance Tests (19)**:
- Software Images: 3 tests (invalid field, duplicate name, warning)
- Categories: 3 tests (invalid field, duplicate name, warning)
- Devices: 2 tests (invalid hostname, update validation)
- Networks: 2 tests (invalid field, duplicate name)
- Kubernetes Clusters: 2 tests (invalid field, duplicate name)
- Edge cases: 7 additional tests (unknown severity, empty response, multiple errors)

## File Impact

### New Files (2)
1. `/workspace/internal/provider/bcm_client_test.go` - Unit tests for validation
2. 6 validation example files in `/workspace/examples/resources/`

### Modified Files (11)
1. `/workspace/internal/provider/bcm_client.go` - ValidateEntity() helper
2. `/workspace/internal/provider/resource_cmpart_softwareimage.go` - Add validation
3. `/workspace/internal/provider/resource_cmdevice_category.go` - Add validation
4. `/workspace/internal/provider/resource_cmdevice_device.go` - Add validation
5. `/workspace/internal/provider/resource_cmnet_network.go` - Add validation
6. `/workspace/internal/provider/resource_cmkube_cluster.go` - Add validation
7. `/workspace/internal/provider/resource_cmpart_softwareimage_test.go` - Validation tests
8. `/workspace/internal/provider/resource_cmdevice_category_test.go` - Validation tests
9. `/workspace/internal/provider/resource_cmdevice_device_test.go` - Validation tests
10. `/workspace/internal/provider/resource_cmnet_network_test.go` - Validation tests
11. `/workspace/internal/provider/resource_cmkube_cluster_test.go` - Validation tests

### Documentation Updates
- CLAUDE.md - Add validation pattern section
- examples/ - 6 new validation error example files
- docs/ - Auto-generated via `make generate`

## Key Design Decisions

### Service Name Casing Convention
| Resource Type | Service Name | Note |
|---------------|--------------|------|
| Software Images | CMPart | CamelCase |
| Categories | CMDevice | CamelCase |
| Devices | CMDevice | CamelCase |
| Networks | CMNet | CamelCase |
| Kubernetes Clusters | **cmkube** | **LOWERCASE (exception)** |

### Integration Points

All resources follow identical pattern:
1. Build API entity with buildAPIEntity()
2. Call ValidateEntity(ctx, service, method, entity, isCreate)
3. Process validation errors (AddError for ERROR, AddWarning for WARNING)
4. Halt if hasErrors==true
5. Continue with CREATE/UPDATE operation

### Zero UUID Filtering

- **CREATE operations** (isCreate=true): Filter out "Zero UUID" errors (expected for new entities)
- **UPDATE operations** (isCreate=false): Preserve all validation errors including UUID errors

### Severity Handling

- **ERROR severity**: Halt operation with resp.Diagnostics.AddError()
- **WARNING severity**: Display advisory with resp.Diagnostics.AddWarning() but continue
- **Unknown severity**: Treat as ERROR (conservative approach)

## Implementation Strategy

### MVP First (Recommended)

1. Complete Phase 1: Setup (3 tasks)
2. Complete Phase 2: Foundational (14 tasks) - BLOCKS all user stories
3. Complete Phase 3: User Story 1 - Software Images (9 tasks)
4. **VALIDATE MVP**: Test software image validation independently
5. Deploy/demo if ready

**MVP Deliverable**: Single resource (software images) with full validation capability

### Incremental Delivery

1. Setup + Foundational → Foundation ready (17 tasks)
2. User Story 1 → Software Image validation (9 tasks) → MVP!
3. User Story 2 → Category validation (8 tasks)
4. User Story 4 → Device validation (8 tasks)
5. User Story 5 → Network + Kube Cluster validation (16 tasks)
6. User Story 3 → Warning verification (5 tasks)
7. Refactor → Code quality (5 tasks)
8. Performance + Documentation + Final (22 tasks)

Each user story adds validation coverage without breaking previous stories.

### Parallel Team Strategy

**Phase 2 (Foundational)**: Team works together on core validation infrastructure

**After Foundational Complete**:
- Developer A: User Story 1 (Software Images) - Phase 3
- Developer B: User Story 2 (Categories) - Phase 4
- Developer C: User Story 4 (Devices) - Phase 5

**Final Phase**:
- Developer A: Network validation (Phase 6 first half)
- Developer B: Kube Cluster validation (Phase 6 second half)
- Developer C: Documentation (Phase 10)

## Success Criteria

### Functional Requirements (100% Coverage)
- ValidateEntity() helper function: DESIGNED
- ValidationError struct with helper methods: DESIGNED
- All 5 resource types integrate validation: PLANNED (46 integration tasks)
- Zero UUID filtering for CREATE: DESIGNED
- ERROR/WARNING severity handling: DESIGNED
- Correct service name casing: DOCUMENTED (cmkube lowercase exception)

### Testing Requirements
- Unit tests for validation logic: PLANNED (5 tests)
- Acceptance tests for all resources: PLANNED (19 tests)
- Edge case coverage: PLANNED (5 additional unit tests)
- No regressions in existing tests: VALIDATED (make testacc in final phase)

### Performance Requirements
- <200ms validation overhead: TESTABLE (Phase 9 performance measurement)
- No timeouts: VERIFIABLE (acceptance test execution)

### Documentation Requirements
- CLAUDE.md validation section: PLANNED (T075)
- Example validation scenarios: PLANNED (6 example files, T076-T080)
- Generated terraform docs: AUTOMATED (make generate, T081)

## Risk Mitigation

### Critical Risks Addressed

**Risk: Service Name Casing Errors**
- Mitigation: Service name reference table in tasks.md
- Verification: T088 reviews TF_LOG output for correct casing
- Special attention to cmkube lowercase exception

**Risk: Zero UUID Filtering Too Aggressive**
- Mitigation: Specific filtering logic (Field=="uuid" AND Message contains "Zero UUID")
- Testing: Dedicated unit test (T007, T074)

**Risk: Breaking Changes to Existing Resources**
- Mitigation: Validation only adds pre-flight checks, no changes to CRUD logic
- Verification: Full acceptance test suite in T083
- Rollback: Clean revert possible (validation is additive)

**Risk: Test Environment Instability**
- Mitigation: Use generateUniqueTestName() for all test resources
- Cleanup: CheckDestroy functions verify deletion
- Retry: Exponential backoff in verifyResourceDeleted helper

## Estimated Effort

**Total Time**: 2-3 days with parallel execution
- Phase 1-2 (Foundation): 4-6 hours
- Phase 3-7 (User Stories): 8-12 hours (can parallelize)
- Phase 8-11 (Polish): 4-6 hours

**With MVP Strategy**: 1 day to working software image validation

## Dependencies

### External (Already Available)
- BCM API validate* methods: VERIFIED in research
- Terraform Plugin Framework: v1.16.1 installed
- terraform-plugin-testing: v1.13.3 installed

### Internal (Already Available)
- BCMClient.CallJSONRPC(): EXISTS
- buildAPIEntity() helpers: EXISTS in all 5 resources
- Test helper infrastructure: EXISTS (createTestBCMClient, generateUniqueTestName)

**No Blockers**: All infrastructure exists, ready to start implementation

## Next Steps

1. Execute Phase 1 (Setup): Tasks T001-T003
2. Execute Phase 2 (Foundational): Tasks T004-T017
   - RED: Write failing unit tests (T004-T009)
   - GREEN: Implement ValidateEntity() (T010-T017)
3. Execute Phase 3 (US1): Tasks T018-T026
   - RED: Write failing acceptance tests (T018-T020)
   - GREEN: Integrate validation in software images (T021-T026)
4. VALIDATE MVP: Run software image validation tests
5. Continue with remaining user stories or deploy MVP

## Conclusion

Successfully generated comprehensive task list with:
- 90 detailed, actionable tasks
- Clear TDD RED-GREEN-REFACTOR workflow
- User story organization for independent testing
- 41 parallel execution opportunities
- Complete integration coverage across 5 resource types
- ~24 new tests for validation logic
- Clear success criteria and acceptance tests

**Task file location**: `/workspace/specs/001-bcm-preflight-validation/tasks.md`

**Status**: READY FOR EXECUTION
