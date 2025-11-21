# Tasks: Modernize Terraform BCM Provider Test Suite

**Feature Branch**: `001-modernize-test-suite`
**Input**: Design documents from `/workspace/specs/001-modernize-test-suite/`
**Prerequisites**: plan.md, spec.md (user stories with priorities), research.md (modern patterns)

**Organization**: Tasks are organized by priority phases following the plan.md structure. Tests are OPTIONAL and only included if TDD approach is explicitly requested in the specification.

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US6)
- All file paths are absolute

---

## Phase 1: HIGH PRIORITY - Resource State Checks + Idempotency

**Purpose**: Implement modern state verification and idempotency checking for resource tests (18 tests across 2 files)

**Target**: resource_cmpart_softwareimage_test.go (85% → 95%+), resource_cmdevice_category_test.go (80% → 95%+)

**TDD Workflow**: RED (add modern checks, verify some fail) → GREEN (fix issues) → REFACTOR (cleanup)

### Resource: Software Image (12 tests) - User Story 1 & 2

- [X] T001 [P] [US1] Add statecheck.ExpectKnownValue() for all string attributes in TestAccCMPartSoftwareImageResource_Basic in /workspace/internal/provider/resource_cmpart_softwareimage_test.go
- [X] T002 [P] [US1] Add statecheck.ExpectKnownValue() for all boolean attributes (enable_sol, install_boot_record) in TestAccCMPartSoftwareImageResource_Basic in /workspace/internal/provider/resource_cmpart_softwareimage_test.go
- [X] T003 [P] [US1] Add statecheck.ExpectKnownValue() with knownvalue.Int64Exact() for sol_speed in TestAccCMPartSoftwareImageResource_Basic in /workspace/internal/provider/resource_cmpart_softwareimage_test.go
- [X] T004 [P] [US1] Add statecheck.ExpectKnownValue() with knownvalue.NotNull() for computed fields (uuid, id) in TestAccCMPartSoftwareImageResource_Basic in /workspace/internal/provider/resource_cmpart_softwareimage_test.go
- [X] T005 [US1] Add statecheck.CompareValue() ID tracking across Create/Import/Update steps in TestAccCMPartSoftwareImageResource_Basic in /workspace/internal/provider/resource_cmpart_softwareimage_test.go
- [X] T006 [P] [US2] Add plancheck.ExpectEmptyPlan() step after Create in TestAccCMPartSoftwareImageResource_Basic in /workspace/internal/provider/resource_cmpart_softwareimage_test.go
- [X] T007 [P] [US2] Add plancheck.ExpectEmptyPlan() step after Update in TestAccCMPartSoftwareImageResource_Basic in /workspace/internal/provider/resource_cmpart_softwareimage_test.go
- [ ] T008 [P] [US1] Add state checks to TestAccCMPartSoftwareImageResource_FullConfig in /workspace/internal/provider/resource_cmpart_softwareimage_test.go
- [ ] T009 [P] [US2] Add idempotency checks to TestAccCMPartSoftwareImageResource_FullConfig in /workspace/internal/provider/resource_cmpart_softwareimage_test.go
- [ ] T010 [P] [US1] Add state checks to TestAccCMPartSoftwareImageResource_UpdateKernelConfig in /workspace/internal/provider/resource_cmpart_softwareimage_test.go
- [ ] T011 [P] [US2] Add idempotency checks to TestAccCMPartSoftwareImageResource_UpdateKernelConfig in /workspace/internal/provider/resource_cmpart_softwareimage_test.go
- [ ] T012 [P] [US1] Add state checks to TestAccCMPartSoftwareImageResource_UpdateBootRecord in /workspace/internal/provider/resource_cmpart_softwareimage_test.go
- [ ] T013 [P] [US2] Add idempotency checks to TestAccCMPartSoftwareImageResource_UpdateBootRecord in /workspace/internal/provider/resource_cmpart_softwareimage_test.go
- [ ] T014 [P] [US1] Add state checks to TestAccCMPartSoftwareImageResource_UpdateSOL in /workspace/internal/provider/resource_cmpart_softwareimage_test.go
- [ ] T015 [P] [US2] Add idempotency checks to TestAccCMPartSoftwareImageResource_UpdateSOL in /workspace/internal/provider/resource_cmpart_softwareimage_test.go
- [ ] T016 [P] [US1] Add state checks to TestAccCMPartSoftwareImageResource_UpdateModules in /workspace/internal/provider/resource_cmpart_softwareimage_test.go
- [ ] T017 [P] [US2] Add idempotency checks to TestAccCMPartSoftwareImageResource_UpdateModules in /workspace/internal/provider/resource_cmpart_softwareimage_test.go
- [ ] T018 [P] [US1] Add state checks to TestAccCMPartSoftwareImageResource_Clone in /workspace/internal/provider/resource_cmpart_softwareimage_test.go
- [ ] T019 [P] [US2] Add idempotency checks to TestAccCMPartSoftwareImageResource_Clone (accounting for eventual consistency) in /workspace/internal/provider/resource_cmpart_softwareimage_test.go
- [ ] T020 [P] [US1] Add state checks to TestAccCMPartSoftwareImageResource_DriftKernelParameters in /workspace/internal/provider/resource_cmpart_softwareimage_test.go
- [ ] T021 [P] [US1] Add state checks to TestAccCMPartSoftwareImageResource_ConcurrentCreate in /workspace/internal/provider/resource_cmpart_softwareimage_test.go
- [ ] T022 [P] [US1] Add state checks to TestAccCMPartSoftwareImageResource_UpdatePath in /workspace/internal/provider/resource_cmpart_softwareimage_test.go
- [ ] T023 [P] [US2] Add idempotency checks to TestAccCMPartSoftwareImageResource_UpdatePath in /workspace/internal/provider/resource_cmpart_softwareimage_test.go
- [ ] T024 [P] [US1] Add state checks to TestAccCMPartSoftwareImageResource_Import in /workspace/internal/provider/resource_cmpart_softwareimage_test.go
- [ ] T025 [US1] Verify all 12 tests in resource_cmpart_softwareimage_test.go pass with modern patterns by running TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMPartSoftwareImage

### Resource: Category (6 tests) - User Story 1 & 2

- [ ] T026 [P] [US1] Add statecheck.ExpectKnownValue() for all string attributes (name, notes) in TestAccCMDeviceCategoryResource_Basic in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T027 [P] [US1] Add statecheck.ExpectKnownValue() with knownvalue.NotNull() for computed fields (uuid, id) in TestAccCMDeviceCategoryResource_Basic in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T028 [US1] Add statecheck.CompareValue() ID tracking across Create/Import/Update steps in TestAccCMDeviceCategoryResource_Basic in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T029 [P] [US2] Add plancheck.ExpectEmptyPlan() step after Create in TestAccCMDeviceCategoryResource_Basic in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T030 [P] [US2] Add plancheck.ExpectEmptyPlan() step after Update in TestAccCMDeviceCategoryResource_Basic in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T031 [P] [US1] Add state checks to TestAccCMDeviceCategoryResource_UpdateNotes in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T032 [P] [US2] Add idempotency checks to TestAccCMDeviceCategoryResource_UpdateNotes in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T033 [P] [US1] Add state checks to TestAccCMDeviceCategoryResource_DriftNotes in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T034 [P] [US1] Add state checks to TestAccCMDeviceCategoryResource_ConcurrentCreate in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T035 [P] [US1] Add state checks to TestAccCMDeviceCategoryResource_Import in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T036 [P] [US1] Add state checks to TestAccCMDeviceCategoryResource_UpdateManagementNetwork in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T037 [P] [US2] Add idempotency checks to TestAccCMDeviceCategoryResource_UpdateManagementNetwork in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T038 [US1] Verify all 6 tests in resource_cmdevice_category_test.go pass with modern patterns by running TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategory

**Phase 1 Checkpoint**: All resource tests (18 total) now use modern state checks, ID tracking, and idempotency verification. Target quality scores: 95%+

---

## Phase 2: HIGH PRIORITY - Data Source Filter Verification

**Purpose**: Remove hardcoded environment assumptions and add filter verification to data source tests (13 tests across 3 files)

**Target**: data_source_cmnet_networks_test.go (60% → 90%+), data_source_cmdevice_nodes_test.go (40% → 90%+), data_source_cmpart_softwareimages_test.go (50% → 90%+)

**TDD Workflow**: RED (add filter verification, expect initial pass) → GREEN (confirm) → REFACTOR (remove legacy)

### Data Source: Networks (4 tests) - User Story 3 & 4

- [ ] T039 [P] [US4] Remove hardcoded networks.# = "3" assertion from TestAccCMNetNetworksDataSource_Basic in /workspace/internal/provider/data_source_cmnet_networks_test.go
- [ ] T040 [P] [US4] Replace with dynamic assertion checking networks.# > 0 in TestAccCMNetNetworksDataSource_Basic in /workspace/internal/provider/data_source_cmnet_networks_test.go
- [ ] T041 [P] [US1] Add statecheck.ExpectKnownValue() to verify network attributes (name, uuid, dhcp_enabled) in TestAccCMNetNetworksDataSource_Basic in /workspace/internal/provider/data_source_cmnet_networks_test.go
- [ ] T042 [P] [US4] Remove hardcoded networks.0.name = "managementnet" assertion from TestAccCMNetNetworksDataSource_FilterByName in /workspace/internal/provider/data_source_cmnet_networks_test.go
- [ ] T043 [P] [US3] Add statecheck.ExpectKnownValue() with knownvalue.StringRegexp() to verify filtered networks match name_pattern in TestAccCMNetNetworksDataSource_FilterByName in /workspace/internal/provider/data_source_cmnet_networks_test.go
- [ ] T044 [P] [US4] Remove hardcoded networks.# = "2" from TestAccCMNetNetworksDataSource_FilterByDHCP in /workspace/internal/provider/data_source_cmnet_networks_test.go
- [ ] T045 [P] [US3] Add statecheck.ExpectKnownValue() with knownvalue.Bool() to verify filtered networks have dhcp_enabled = filter value in TestAccCMNetNetworksDataSource_FilterByDHCP in /workspace/internal/provider/data_source_cmnet_networks_test.go
- [ ] T046 [P] [US1] Add statecheck.ExpectKnownValue() for network attributes in TestAccCMNetNetworksDataSource_FilterMultiple in /workspace/internal/provider/data_source_cmnet_networks_test.go
- [ ] T047 [P] [US3] Add filter verification for combined name_pattern + dhcp_enabled filters in TestAccCMNetNetworksDataSource_FilterMultiple in /workspace/internal/provider/data_source_cmnet_networks_test.go
- [ ] T048 [US4] Verify all 4 tests in data_source_cmnet_networks_test.go pass on different cluster configurations by running TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMNetNetworksDataSource

### Data Source: Nodes (4 tests) - User Story 3 & 4

- [ ] T049 [P] [US1] Add statecheck.ExpectKnownValue() for node attributes (hostname, uuid, node_type) in TestAccCMDeviceNodesDataSource_Basic in /workspace/internal/provider/data_source_cmdevice_nodes_test.go
- [ ] T050 [P] [US4] Replace any hardcoded node counts with dynamic assertions checking nodes.# > 0 in TestAccCMDeviceNodesDataSource_Basic in /workspace/internal/provider/data_source_cmdevice_nodes_test.go
- [ ] T051 [P] [US3] Add statecheck.ExpectKnownValue() with knownvalue.StringExact() to verify filtered nodes match node_type filter in TestAccCMDeviceNodesDataSource_FilterByType in /workspace/internal/provider/data_source_cmdevice_nodes_test.go
- [ ] T052 [P] [US3] Add statecheck.ExpectKnownValue() with knownvalue.StringRegexp() to verify filtered nodes match hostname_pattern filter in TestAccCMDeviceNodesDataSource_FilterByHostname in /workspace/internal/provider/data_source_cmdevice_nodes_test.go
- [ ] T053 [P] [US1] Add statecheck.ExpectKnownValue() for node attributes in TestAccCMDeviceNodesDataSource_FilterMultiple in /workspace/internal/provider/data_source_cmdevice_nodes_test.go
- [ ] T054 [P] [US3] Add filter verification for combined node_type + hostname_pattern filters in TestAccCMDeviceNodesDataSource_FilterMultiple in /workspace/internal/provider/data_source_cmdevice_nodes_test.go
- [ ] T055 [US4] Verify all 4 tests in data_source_cmdevice_nodes_test.go pass on different cluster configurations by running TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceNodesDataSource

### Data Source: Software Images (5 tests) - User Story 3 & 4

- [ ] T056 [P] [US1] Add statecheck.ExpectKnownValue() for software image attributes (name, path, uuid, category) in TestAccCMPartSoftwareImagesDataSource_Basic in /workspace/internal/provider/data_source_cmpart_softwareimages_test.go
- [ ] T057 [P] [US4] Replace any hardcoded image counts with dynamic assertions checking software_images.# > 0 in TestAccCMPartSoftwareImagesDataSource_Basic in /workspace/internal/provider/data_source_cmpart_softwareimages_test.go
- [ ] T058 [P] [US3] Add statecheck.ExpectKnownValue() to verify filtered images match category filter in TestAccCMPartSoftwareImagesDataSource_FilterByCategory in /workspace/internal/provider/data_source_cmpart_softwareimages_test.go
- [ ] T059 [P] [US1] Add statecheck.ExpectKnownValue() for image attributes in TestAccCMPartSoftwareImagesDataSource_FilterByName in /workspace/internal/provider/data_source_cmpart_softwareimages_test.go
- [ ] T060 [P] [US3] Add statecheck.ExpectKnownValue() with knownvalue.StringRegexp() to verify filtered images match name pattern in TestAccCMPartSoftwareImagesDataSource_FilterByName in /workspace/internal/provider/data_source_cmpart_softwareimages_test.go
- [ ] T061 [P] [US1] Add statecheck.ExpectKnownValue() for image attributes in TestAccCMPartSoftwareImagesDataSource_AllAttributes in /workspace/internal/provider/data_source_cmpart_softwareimages_test.go
- [ ] T062 [P] [US1] Add statecheck.ExpectKnownValue() for image attributes in TestAccCMPartSoftwareImagesDataSource_EmptyFilter in /workspace/internal/provider/data_source_cmpart_softwareimages_test.go
- [ ] T063 [US4] Verify all 5 tests in data_source_cmpart_softwareimages_test.go pass on different cluster configurations by running TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMPartSoftwareImagesDataSource

**Phase 2 Checkpoint**: All data source filter tests (13 total) now use modern state checks, verify filter correctness, and work on any cluster. Target quality scores: 90%+

---

## Phase 3: MEDIUM PRIORITY - Validation + CheckDestroy Enhancement

**Purpose**: Add validation testing and enhance CheckDestroy error reporting

**Target**: Validation tests for resources, enhanced CheckDestroy in 2 resource files

### Validation Tests - User Story 6

- [ ] T064 [P] [US6] Add TestAccCMPartSoftwareImageResource_ValidationInvalidProxy with ExpectError for invalid software_image_proxy URL in /workspace/internal/provider/resource_cmpart_softwareimage_test.go
- [ ] T065 [P] [US6] Add TestAccCMPartSoftwareImageResource_ValidationInvalidPath with ExpectError for invalid path format in /workspace/internal/provider/resource_cmpart_softwareimage_test.go
- [ ] T066 [P] [US6] Add TestAccCMDeviceCategoryResource_ValidationInvalidManagementNetwork with ExpectError for invalid management_network name in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T067 [US6] Verify validation tests catch invalid input by running TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "Validation"

### Enhanced CheckDestroy - User Story 5

- [ ] T068 [P] [US5] Enhance testAccCheckCMPartSoftwareImageDestroy in /workspace/internal/provider/resource_cmpart_softwareimage_test.go with detailed error messages including resource type, ID, and failure reason
- [ ] T069 [P] [US5] Update testAccCheckCMPartSoftwareImageDestroy to accumulate all deletion failures and report them together in /workspace/internal/provider/resource_cmpart_softwareimage_test.go
- [ ] T070 [P] [US5] Add exponential backoff verification logging to testAccCheckCMPartSoftwareImageDestroy in /workspace/internal/provider/resource_cmpart_softwareimage_test.go
- [ ] T071 [P] [US5] Enhance testAccCheckCMDeviceCategoryDestroy in /workspace/internal/provider/resource_cmdevice_category_test.go with detailed error messages including resource type, ID, and failure reason
- [ ] T072 [P] [US5] Update testAccCheckCMDeviceCategoryDestroy to accumulate all deletion failures and report them together in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T073 [P] [US5] Add exponential backoff verification logging to testAccCheckCMDeviceCategoryDestroy in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T074 [US5] Verify enhanced CheckDestroy provides clear error messages by running TF_ACC=1 go test -v -timeout 120m ./internal/provider/

**Phase 3 Checkpoint**: Validation tests ensure schema validators work correctly. Enhanced CheckDestroy provides detailed debugging information for test failures.

---

## Phase 4: LOW PRIORITY - Documentation + Final Quality Verification

**Purpose**: Enhance remaining data source tests, update documentation, verify overall quality targets

**Target**: data_source_cmdevice_categories_test.go (75% → 90%+), CLAUDE.md updates, quality verification

### Data Source: Categories (4 tests)

- [ ] T075 [P] [US1] Add statecheck.ExpectKnownValue() for category attributes (name, uuid, notes) in TestAccCMDeviceCategoriesDataSource_Basic in /workspace/internal/provider/data_source_cmdevice_categories_test.go
- [ ] T076 [P] [US4] Replace any hardcoded category counts with dynamic assertions in TestAccCMDeviceCategoriesDataSource_Basic in /workspace/internal/provider/data_source_cmdevice_categories_test.go
- [ ] T077 [P] [US1] Add enhanced state checks to TestAccCMDeviceCategoriesDataSource_FilterByName in /workspace/internal/provider/data_source_cmdevice_categories_test.go
- [ ] T078 [P] [US3] Add filter verification for name filter in TestAccCMDeviceCategoriesDataSource_FilterByName in /workspace/internal/provider/data_source_cmdevice_categories_test.go
- [ ] T079 [P] [US1] Add enhanced state checks to TestAccCMDeviceCategoriesDataSource_EmptyResult in /workspace/internal/provider/data_source_cmdevice_categories_test.go
- [ ] T080 [P] [US1] Add enhanced state checks to TestAccCMDeviceCategoriesDataSource_AllAttributes in /workspace/internal/provider/data_source_cmdevice_categories_test.go
- [ ] T081 [US4] Verify all 4 tests in data_source_cmdevice_categories_test.go pass by running TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategoriesDataSource

### Documentation Updates

- [ ] T082 [P] Update /workspace/CLAUDE.md with modern testing pattern examples (statecheck, plancheck, knownvalue imports and usage)
- [ ] T083 [P] Add BCM provider attribute type mapping table to /workspace/CLAUDE.md showing string/bool/int64 → knownvalue matcher mappings
- [ ] T084 [P] Document filter verification pattern in /workspace/CLAUDE.md with examples for node_type, hostname_pattern, name_pattern, dhcp_enabled filters
- [ ] T085 [P] Document idempotency verification pattern in /workspace/CLAUDE.md with ExpectEmptyPlan examples
- [ ] T086 [P] Document ID consistency tracking pattern in /workspace/CLAUDE.md with CompareValue examples

### Final Quality Verification

- [ ] T087 Calculate quality scores for all 6 test files using scoring methodology from plan.md (modern patterns 40%, verification completeness 30%, environment portability 20%, best practices 10%)
- [ ] T088 Verify overall test suite quality score is 90%+ by averaging individual file scores
- [ ] T089 Run full acceptance test suite to confirm 100% pass rate by running TF_ACC=1 go test -v -timeout 120m ./internal/provider/
- [ ] T090 Verify test execution time is within 10% of original baseline by comparing test durations
- [ ] T091 Create summary report documenting quality improvements, test coverage, and environment portability enhancements in /workspace/specs/001-modernize-test-suite/RESULTS.md

**Phase 4 Checkpoint**: Test suite modernization complete. All tests use modern patterns, verify correctness thoroughly, and work on any BCM cluster. Documentation updated with modern patterns.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (HIGH PRIORITY)**: Resource state checks + idempotency - Can start immediately, highest impact
- **Phase 2 (HIGH PRIORITY)**: Data source filter verification - Can run in parallel with Phase 1 (different files)
- **Phase 3 (MEDIUM PRIORITY)**: Validation + CheckDestroy - Depends on Phase 1 completion (same files)
- **Phase 4 (LOW PRIORITY)**: Documentation + final verification - Depends on Phase 1-3 completion

### User Story Dependencies

- **US1 (Modern State Verification)**: No dependencies - foundational pattern
- **US2 (Idempotency)**: Depends on US1 state checks being added to same test steps
- **US3 (Filter Verification)**: Depends on US1 state checks for data sources
- **US4 (Environment Portability)**: Independent - can run in parallel with US1/US2/US3
- **US5 (CheckDestroy)**: Independent - different functions than US1/US2
- **US6 (Validation)**: Independent - new test functions

### Within Each Phase

- **Phase 1**: All tasks marked [P] can run in parallel (different test functions in same or different files)
- **Phase 2**: All tasks marked [P] can run in parallel (different test files)
- **Phase 3**: All tasks marked [P] can run in parallel (validation tests are new functions, CheckDestroy enhancements are in different functions)
- **Phase 4**: All tasks marked [P] can run in parallel (categories tests, documentation updates)

### Parallel Opportunities

**Maximum Parallelism (All Phases)**:
```bash
# Phase 1: Modernize all resource tests in parallel
Task T001-T025: Software image tests (25 tasks)
Task T026-T038: Category tests (13 tasks)

# Phase 2: Modernize all data source tests in parallel
Task T039-T048: Networks tests (10 tasks)
Task T049-T055: Nodes tests (7 tasks)
Task T056-T063: Software images tests (8 tasks)

# Phase 3: All validation and CheckDestroy tasks in parallel
Task T064-T067: Validation tests (4 tasks)
Task T068-T074: CheckDestroy enhancements (7 tasks)

# Phase 4: Documentation and quality verification
Task T075-T081: Categories tests (7 tasks)
Task T082-T086: Documentation updates (5 tasks)
Task T087-T091: Quality verification (5 tasks - sequential)
```

**Recommended Strategy** (TDD RED-GREEN-REFACTOR):

1. **RED Phase** - Add all modern checks to one test file, verify some expectations correct:
   - Complete T001-T025 (software image) or T026-T038 (category)
   - Run tests, expect most to pass (patterns already good)
   - Fix any revealed issues

2. **GREEN Phase** - Verify all tests pass with modern checks:
   - Run verification tasks (T025, T038, T048, T055, T063, etc.)
   - Confirm 100% pass rate maintained

3. **REFACTOR Phase** - Optional cleanup:
   - Remove redundant legacy TestCheckResourceAttr calls
   - Add comments explaining modern pattern usage
   - Re-verify tests still pass

---

## Implementation Strategy

### MVP Approach (Phases 1-2 Only)

If time-constrained, prioritize:
1. Complete Phase 1: Resource state checks + idempotency (highest reliability impact)
2. Complete Phase 2: Data source filter verification + portability (highest correctness impact)
3. Stop and validate: Run full test suite, verify quality score improvement
4. Phase 3-4 can be deferred to future iteration

### Incremental Delivery

1. **Phase 1 Complete** → Resource tests: 85-80% → 95%+ quality
2. **Phase 2 Complete** → Data source tests: 40-60% → 90%+ quality
3. **Phase 3 Complete** → Validation coverage + better debugging
4. **Phase 4 Complete** → All tests 90%+, documentation complete

### Parallel Team Strategy

With multiple developers:
1. **Developer A**: Phase 1 (resource modernization)
2. **Developer B**: Phase 2 (data source modernization)
3. **Developer C**: Phase 3 (validation + CheckDestroy)
4. **All**: Phase 4 (documentation + verification) after Phases 1-3 done

---

## HashiCorp Testing Patterns Reference

All tasks implement patterns from https://developer.hashicorp.com/terraform/plugin/testing/testing-patterns:

1. **Basic Attribute Verification**: Using `statecheck.ExpectKnownValue()` with type-safe matchers
2. **Configuration Updates**: Maintaining update tests with idempotency verification via `plancheck.ExpectEmptyPlan()`
3. **Import Mode Testing**: Preserving ImportState tests with ID consistency tracking via `statecheck.CompareValue()`
4. **Error Expectations**: Adding validation tests with `ExpectError` for schema validators

**Required Imports** (all test files):
```go
import (
    "github.com/hashicorp/terraform-plugin-testing/helper/resource"
    "github.com/hashicorp/terraform-plugin-testing/plancheck"
    "github.com/hashicorp/terraform-plugin-testing/statecheck"
    "github.com/hashicorp/terraform-plugin-testing/knownvalue"
    "github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
    "github.com/hashicorp/terraform-plugin-testing/compare"
)
```

---

## Success Criteria Summary

- ✅ **SC-001**: Test suite quality score 69% → 90%+ (measured after T087-T088)
- ✅ **SC-002**: All 18 resource tests include idempotency verification (T006-T007, T009, T011, T013, T015, T017, T019, T023, T029-T030, T032, T037)
- ✅ **SC-003**: All resource tests use ExpectKnownValue for 80%+ of attributes (T001-T005, T008, T010, T012, T014, T016, T018, T020-T022, T024, T026-T027, T031, T033-T036)
- ✅ **SC-004**: All resource tests use CompareValue for ID tracking (T005, T028)
- ✅ **SC-005**: All data source filter tests verify filter correctness (T043, T045, T047, T051-T052, T054, T058, T060, T078)
- ✅ **SC-006**: Zero hardcoded environment values (T039, T042, T044, T050, T057, T076)
- ✅ **SC-007**: All tests pass on different BCM clusters (T048, T055, T063, T081)
- ✅ **SC-008**: Enhanced CheckDestroy with detailed errors (T068-T073)
- ✅ **SC-009**: Test execution time within 10% baseline (T090)
- ✅ **SC-010**: All existing scenarios pass (T089)

---

## Notes

- **[P] tasks**: Different files or different test functions - can run in parallel
- **[Story] labels**: Map tasks to user stories for traceability (US1-US6)
- **TDD approach**: Add modern checks first (RED), verify pass (GREEN), cleanup (REFACTOR)
- **Backward compatibility**: Keep legacy TestCheckResourceAttr initially, remove after modern checks proven
- **Test isolation**: Each test uses generateUniqueTestName() for concurrent execution safety
- **Eventual consistency**: Software image clone tests account for async operations with polling
- **Filter edge cases**: Filter verification handles empty results gracefully (check list size before accessing elements)
- **Quality scoring**: Use plan.md methodology (40% modern patterns + 30% verification + 20% portability + 10% best practices)
