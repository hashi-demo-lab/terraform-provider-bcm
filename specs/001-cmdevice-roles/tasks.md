# Tasks: BCM Device Roles Data Source

**Feature**: bcm_cmdevice_roles data source
**Input**: Design documents from `/workspace/specs/001-cmdevice-roles/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Organization**: Tasks follow TDD RED-GREEN-REFACTOR cycle, organized by user story priority (P1, P2, P3) to enable independent implementation and testing.

## Format: `- [ ] [ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [ ] T001 Verify BCM API credentials configured in environment (BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD)
- [ ] T002 Verify test dependencies available in /workspace/internal/provider/test_helpers.go
- [ ] T003 [P] Review reference implementation patterns in /workspace/internal/provider/data_source_cmdevice_categories.go
- [ ] T004 [P] Review null-safe helper functions in /workspace/internal/provider/data_source_cmpart_softwareimages.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**Critical**: No user story work can begin until this phase is complete

- [ ] T005 Create data source skeleton in /workspace/internal/provider/data_source_cmdevice_roles.go
- [ ] T006 Define schema with optional filter attributes (name_pattern, child_type) and computed roles list
- [ ] T007 Register data source in /workspace/internal/provider/provider.go DataSources() method
- [ ] T008 Create test file skeleton in /workspace/internal/provider/data_source_cmdevice_roles_test.go
- [ ] T009 Implement testAccPreCheck function if not already present in test file

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Query All Available Roles (Priority: P1) - MVP

**Goal**: Enable DevOps engineers to discover all available role types in BCM without filters

**Independent Test**: Query data source without filters and verify all roles are returned with complete metadata (uuid, name, childType)

### RED: Write Failing Tests for User Story 1

**Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T010 [P] [US1] Write TestAccCMDeviceRolesDataSource_All in /workspace/internal/provider/data_source_cmdevice_roles_test.go
- [ ] T011 [P] [US1] Write testAccCMDeviceRolesDataSourceConfig helper function with provider config
- [ ] T012 [US1] Run test to verify it fails with "data source not found" error

### GREEN: Minimal Implementation for User Story 1

- [ ] T013 [US1] Implement Metadata method returning "bcm_cmdevice_roles" type name in /workspace/internal/provider/data_source_cmdevice_roles.go
- [ ] T014 [US1] Implement Configure method to receive BCMClient from provider in /workspace/internal/provider/data_source_cmdevice_roles.go
- [ ] T015 [US1] Implement Schema method defining id, name_pattern, child_type, roles attributes in /workspace/internal/provider/data_source_cmdevice_roles.go
- [ ] T016 [US1] Implement Read method: Call cmdevice.getNodes API in /workspace/internal/provider/data_source_cmdevice_roles.go
- [ ] T017 [US1] Implement Read method: Parse nodes response and extract roles array in /workspace/internal/provider/data_source_cmdevice_roles.go
- [ ] T018 [US1] Implement Read method: Deduplicate roles by UUID using map[uuid]Role in /workspace/internal/provider/data_source_cmdevice_roles.go
- [ ] T019 [US1] Implement Read method: Convert deduplicated roles to RoleModel slice in /workspace/internal/provider/data_source_cmdevice_roles.go
- [ ] T020 [US1] Implement Read method: Set id and roles in Terraform state in /workspace/internal/provider/data_source_cmdevice_roles.go
- [ ] T021 [US1] Run TestAccCMDeviceRolesDataSource_All and verify it passes
- [ ] T022 [US1] Add ConfigStateChecks using statecheck.ExpectKnownValue for id attribute in test

### REFACTOR: Improve User Story 1

- [ ] T023 [US1] Add null-safe field extraction using getStringValue and getBoolValue helpers in /workspace/internal/provider/data_source_cmdevice_roles.go
- [ ] T024 [US1] Add error handling for API call failures in /workspace/internal/provider/data_source_cmdevice_roles.go
- [ ] T025 [US1] Add error handling for JSON parsing failures in /workspace/internal/provider/data_source_cmdevice_roles.go
- [ ] T026 [US1] Add debug logging for role extraction in /workspace/internal/provider/data_source_cmdevice_roles.go
- [ ] T027 [US1] Add ConfigStateChecks for role attributes (uuid, name, child_type, base_type) in test
- [ ] T028 [US1] Run all User Story 1 tests and verify they pass

**Checkpoint**: At this point, User Story 1 should be fully functional - can query all roles without filters

---

## Phase 4: User Story 2 - Filter Roles by Type (Priority: P2)

**Goal**: Enable automation scripts to filter roles by childType for targeted queries

**Independent Test**: Apply child_type filter and verify only matching roles are returned

### RED: Write Failing Tests for User Story 2

- [ ] T029 [P] [US2] Write TestAccCMDeviceRolesDataSource_FilterByChildType in /workspace/internal/provider/data_source_cmdevice_roles_test.go
- [ ] T030 [P] [US2] Write testAccCMDeviceRolesDataSourceConfigFilterByChildType helper function in /workspace/internal/provider/data_source_cmdevice_roles_test.go
- [ ] T031 [US2] Run test to verify it fails (returns unfiltered results)

### GREEN: Minimal Implementation for User Story 2

- [ ] T032 [US2] Implement matchesRoleFilter function with child_type exact match logic in /workspace/internal/provider/data_source_cmdevice_roles.go
- [ ] T033 [US2] Update Read method to apply matchesRoleFilter before adding roles to results in /workspace/internal/provider/data_source_cmdevice_roles.go
- [ ] T034 [US2] Handle null/unknown child_type values (skip filter if not specified) in matchesRoleFilter function
- [ ] T035 [US2] Run TestAccCMDeviceRolesDataSource_FilterByChildType and verify it passes

### REFACTOR: Improve User Story 2

- [ ] T036 [US2] Add test cases for different childType values (HeadNodeRole, ComputeRole, StorageRole) in /workspace/internal/provider/data_source_cmdevice_roles_test.go
- [ ] T037 [US2] Add test case for non-existent childType returning empty results in /workspace/internal/provider/data_source_cmdevice_roles_test.go
- [ ] T038 [US2] Add debug logging for filter application in /workspace/internal/provider/data_source_cmdevice_roles.go
- [ ] T039 [US2] Run all User Story 2 tests and verify they pass

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - Filter Roles by Name Pattern (Priority: P3)

**Goal**: Enable advanced users to filter roles by name patterns for complex deployments

**Independent Test**: Apply name_pattern filter with glob pattern and verify pattern matching works

### RED: Write Failing Tests for User Story 3

- [ ] T040 [P] [US3] Write TestAccCMDeviceRolesDataSource_FilterByNamePattern in /workspace/internal/provider/data_source_cmdevice_roles_test.go
- [ ] T041 [P] [US3] Write testAccCMDeviceRolesDataSourceConfigFilterByNamePattern helper function in /workspace/internal/provider/data_source_cmdevice_roles_test.go
- [ ] T042 [US3] Run test to verify it fails (glob matching not implemented)

### GREEN: Minimal Implementation for User Story 3

- [ ] T043 [US3] Add filepath.Match import for glob pattern matching in /workspace/internal/provider/data_source_cmdevice_roles.go
- [ ] T044 [US3] Update matchesRoleFilter to add name_pattern glob matching using filepath.Match in /workspace/internal/provider/data_source_cmdevice_roles.go
- [ ] T045 [US3] Handle null/unknown name_pattern values (skip filter if not specified) in matchesRoleFilter function
- [ ] T046 [US3] Add error handling for invalid glob patterns in matchesRoleFilter function
- [ ] T047 [US3] Run TestAccCMDeviceRolesDataSource_FilterByNamePattern and verify it passes

### REFACTOR: Improve User Story 3

- [ ] T048 [US3] Add test cases for different glob patterns (prefix*, *suffix, node-?, [abc]*) in /workspace/internal/provider/data_source_cmdevice_roles_test.go
- [ ] T049 [US3] Write TestAccCMDeviceRolesDataSource_CombinedFilters testing both filters together (AND logic) in /workspace/internal/provider/data_source_cmdevice_roles_test.go
- [ ] T050 [US3] Write testAccCMDeviceRolesDataSourceConfigCombinedFilters helper function in /workspace/internal/provider/data_source_cmdevice_roles_test.go
- [ ] T051 [US3] Write TestAccCMDeviceRolesDataSource_EmptyResults testing filter with no matches in /workspace/internal/provider/data_source_cmdevice_roles_test.go
- [ ] T052 [US3] Run all User Story 3 tests and verify they pass

**Checkpoint**: All user stories should now be independently functional

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

### Examples & Documentation

- [ ] T053 [P] Create /workspace/examples/data-sources/bcm_cmdevice_roles/data-source.tf (query all roles)
- [ ] T054 [P] Create /workspace/examples/data-sources/bcm_cmdevice_roles/filter-by-type.tf (filter by childType)
- [ ] T055 [P] Create /workspace/examples/data-sources/bcm_cmdevice_roles/filter-by-pattern.tf (filter by name pattern)
- [ ] T056 Generate provider documentation using make generate command
- [ ] T057 Verify generated documentation in /workspace/docs/data-sources/bcm_cmdevice_roles.md

### Code Quality & Final Validation

- [ ] T058 [P] Run gofmt on data source implementation file
- [ ] T059 [P] Run golangci-lint on data source implementation file
- [ ] T060 Run all acceptance tests with TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMDeviceRoles
- [ ] T061 Verify all tests pass and no regression in existing data sources
- [ ] T062 Manual test on live BCM cluster using examples
- [ ] T063 Verify quickstart.md instructions work end-to-end
- [ ] T064 Update CHANGELOG.md with new data source entry (if project has one)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Independent, but builds on US1 Read implementation
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - Independent, but uses same matchesRoleFilter as US2

### Within Each User Story

**RED Phase** (write failing tests):
- Tests can be written in parallel [P] marked tasks
- Must run and FAIL before GREEN phase

**GREEN Phase** (minimal implementation):
- Sequential implementation steps
- Must make tests pass before REFACTOR phase

**REFACTOR Phase** (improve quality):
- Additional tests can be written in parallel [P]
- Code improvements are sequential
- All tests must pass before moving to next story

### Parallel Opportunities

**Phase 1 (Setup)**:
- T003 and T004 can run in parallel (reviewing different reference files)

**Phase 2 (Foundational)**:
- All tasks are sequential (building data source skeleton)

**Phase 3 (User Story 1 - RED)**:
- T010 and T011 can be written in parallel (test function and helper)

**Phase 4 (User Story 2 - RED)**:
- T029 and T030 can be written in parallel (test function and helper)

**Phase 5 (User Story 3 - RED)**:
- T040 and T041 can be written in parallel (test function and helper)

**Phase 6 (Polish)**:
- T053, T054, T055 can be written in parallel (independent example files)
- T058 and T059 can run in parallel (formatting and linting)

**Entire User Stories** (if multiple developers):
- After Phase 2 completes, US1, US2, US3 can be worked on in parallel by different team members
- Each story's RED-GREEN-REFACTOR cycle is independent

---

## Parallel Example: User Story 1 RED Phase

```bash
# Launch test files in parallel:
Task T010: Write TestAccCMDeviceRolesDataSource_All
Task T011: Write testAccCMDeviceRolesDataSourceConfig helper
# Then run T012 to verify both fail
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (review patterns)
2. Complete Phase 2: Foundational (data source skeleton)
3. Complete Phase 3: User Story 1 (query all roles without filters)
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Deploy/demo if ready

**Estimated Time**: 2-3 hours following quickstart.md

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP - basic role discovery)
3. Add User Story 2 → Test independently → Deploy/Demo (+ childType filtering)
4. Add User Story 3 → Test independently → Deploy/Demo (+ pattern matching)
5. Each story adds value without breaking previous stories

**Estimated Time**: 4-5 hours total

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together (1 hour)
2. Once Foundational is done:
   - Developer A: User Story 1 (1.5 hours)
   - Developer B: User Story 2 (1 hour) - can start after US1 GREEN phase
   - Developer C: User Story 3 (1 hour) - can start after US2 GREEN phase
3. Team completes Polish together (30 minutes)

**Estimated Time**: 2-3 hours with 3 developers

---

## Testing Strategy

### Test Environment Setup

```bash
export TF_ACC=1
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"
export TF_LOG=DEBUG  # Optional for debugging
```

### Running Tests

```bash
# Run all role data source tests
TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMDeviceRoles

# Run specific test
TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMDeviceRolesDataSource_All

# Run with timeout
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceRoles
```

### Expected Test Coverage

1. **TestAccCMDeviceRolesDataSource_All**: Query all roles (US1)
2. **TestAccCMDeviceRolesDataSource_FilterByChildType**: Filter by exact type (US2)
3. **TestAccCMDeviceRolesDataSource_FilterByNamePattern**: Filter by glob pattern (US3)
4. **TestAccCMDeviceRolesDataSource_CombinedFilters**: Both filters with AND logic (US3)
5. **TestAccCMDeviceRolesDataSource_EmptyResults**: No matches returns empty list (US3)

All tests use modern terraform-plugin-testing patterns:
- `statecheck.ExpectKnownValue` for type-safe state verification
- `knownvalue.NotNull()` for presence checks
- Environment-portable assertions (no hardcoded role names/counts)
- Provider configuration injected via helper functions

---

## Notes

- **[P] tasks**: Different files, no dependencies, can run in parallel
- **[Story] label**: Maps task to specific user story (US1, US2, US3) for traceability
- **TDD Cycle**: Each user story follows RED (failing tests) → GREEN (minimal implementation) → REFACTOR (improve quality)
- **File paths**: All paths are absolute starting from /workspace/
- **Test-first**: Write and verify tests FAIL before implementing functionality
- **Independent stories**: Each story delivers value independently and can be tested in isolation
- **Commit strategy**: Commit after each GREEN phase (tests passing) and REFACTOR phase (quality improved)
- **Stop points**: After each user story checkpoint, validate independently before continuing

---

## Success Criteria

- [ ] All 5 acceptance tests pass
- [ ] Data source registered in provider.go
- [ ] Examples created for all use cases
- [ ] Documentation auto-generated
- [ ] Code follows existing patterns (null-safe helpers, filter logic)
- [ ] Manual testing on live BCM cluster successful
- [ ] No regression in existing data sources
- [ ] Follows TDD RED-GREEN-REFACTOR cycle for each user story
