# Tasks: BCM CMPart Entity Info Data Source

**Input**: Design documents from `/workspace/specs/095-cmpart-entity-info/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, quickstart.md

**Tests**: This implementation follows TDD approach (RED-GREEN-REFACTOR) as requested.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Source**: `internal/provider/`
- **Examples**: `examples/data-sources/bcm_cmpart_entity_info/`
- **Docs**: `docs/data-sources/` (generated via `make generate`)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and test file scaffolding

- [X] T001 Create data source skeleton file at internal/provider/data_source_cmpart_entity_info.go with struct definition and interface assertions
- [X] T002 Create test file skeleton at internal/provider/data_source_cmpart_entity_info_test.go with package declaration and imports
- [X] T003 Register data source in internal/provider/provider.go DataSources() function by adding NewCMPartEntityInfoDataSource

**Acceptance Criteria**:
- Files compile without errors
- Data source is registered in provider (visible in provider schema)
- Test file compiles with placeholder test

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Define schema and models that all user stories depend on

**CRITICAL**: No user story implementation can begin until this phase is complete

- [X] T004 Define CMPartEntityInfoDataSourceModel struct in internal/provider/data_source_cmpart_entity_info.go with ID, Type, NamePattern, Entities fields
- [X] T005 Define EntityInfoModel struct in internal/provider/data_source_cmpart_entity_info.go with Name, Type, UUID fields
- [X] T006 Implement Schema() method in internal/provider/data_source_cmpart_entity_info.go with type (optional string), name_pattern (optional string), id (computed string), entities (computed list)
- [X] T007 Implement Metadata() method in internal/provider/data_source_cmpart_entity_info.go returning TypeName "bcm_cmpart_entity_info"
- [X] T008 Implement Configure() method in internal/provider/data_source_cmpart_entity_info.go to receive BCMClient from provider

**Acceptance Criteria**:
- Schema defines all input/output attributes per data-model.md
- `terraform providers schema` shows bcm_cmpart_entity_info data source
- Configure properly stores BCMClient reference

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - List Entities by Type (Priority: P1) MVP

**Goal**: Enable filtering BCM entities by type (e.g., "SoftwareImage") to discover resources and obtain UUIDs

**Independent Test**: Query with `type = "SoftwareImage"` returns only SoftwareImage entities with valid name, type, uuid fields

### Tests for User Story 1 (TDD RED Phase)

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T009 [P] [US1] Write TestAccCMPartEntityInfoDataSource_Basic acceptance test in internal/provider/data_source_cmpart_entity_info_test.go - verify data source returns entities with id computed
- [X] T010 [P] [US1] Write TestAccCMPartEntityInfoDataSource_FilterByType acceptance test in internal/provider/data_source_cmpart_entity_info_test.go - query with type="SoftwareImage", verify id computed
- [X] T011 [P] [US1] Write TestAccCMPartEntityInfoDataSource_EmptyResult acceptance test in internal/provider/data_source_cmpart_entity_info_test.go - query with type="NonExistentType123", verify empty entities list returned (not error)
- [X] T012 [P] [US1] Write TestAccCMPartEntityInfoDataSource_InvalidCredentials acceptance test in internal/provider/data_source_cmpart_entity_info_test.go - verify authentication error message

**Checkpoint**: All US1 tests should FAIL (RED phase complete)

### Implementation for User Story 1 (TDD GREEN Phase)

- [X] T013 [US1] Implement Read() method skeleton in internal/provider/data_source_cmpart_entity_info.go with API call to cmpart.getBasicEntityInformation
- [X] T014 [US1] Implement API response parsing in internal/provider/data_source_cmpart_entity_info.go - parse JSON array, extract resolveName->name, type, uuid fields
- [X] T015 [US1] Implement type filter logic in internal/provider/data_source_cmpart_entity_info.go - case-sensitive exact match when type attribute is set
- [X] T016 [US1] Implement ID generation in internal/provider/data_source_cmpart_entity_info.go - format: "cmpart-entity-info:{type}:{name_pattern}" or "cmpart-entity-info:all"
- [X] T017 [US1] Implement error handling in internal/provider/data_source_cmpart_entity_info.go - API errors with descriptive messages per error_messages.go patterns

**Acceptance Criteria**:
- All US1 tests pass (GREEN phase complete)
- `type = "SoftwareImage"` returns only SoftwareImage entities
- `type = "NonExistentType"` returns empty list (no error)
- Authentication errors show helpful message

**Checkpoint**: User Story 1 fully functional and independently testable

---

## Phase 4: User Story 2 - Filter Entities by Name Pattern (Priority: P2)

**Goal**: Enable filtering entities by name using glob patterns (*,?) for discovery when exact name is unknown

**Independent Test**: Query with `name_pattern = "default*"` returns only entities whose names match the glob pattern (case-insensitive)

### Tests for User Story 2 (TDD RED Phase)

- [X] T018 [P] [US2] Write TestAccCMPartEntityInfoDataSource_FilterByNamePattern acceptance test in internal/provider/data_source_cmpart_entity_info_test.go - query with name_pattern="default*", verify id computed
- [X] T019 [P] [US2] Write TestAccCMPartEntityInfoDataSource_FilterByNamePatternMiddle acceptance test in internal/provider/data_source_cmpart_entity_info_test.go - query with name_pattern="*node*" for middle match
- [X] T020 [P] [US2] Write TestAccCMPartEntityInfoDataSource_FilterByExactName acceptance test in internal/provider/data_source_cmpart_entity_info_test.go - query without wildcards for literal match

**Checkpoint**: All US2 tests should FAIL (RED phase complete)

### Implementation for User Story 2 (TDD GREEN Phase)

- [X] T021 [US2] Implement matchesNamePattern helper function in internal/provider/data_source_cmpart_entity_info.go using filepath.Match with case normalization (strings.ToLower)
- [X] T022 [US2] Integrate name_pattern filter into Read() method in internal/provider/data_source_cmpart_entity_info.go - apply filter after type filter

**Acceptance Criteria**:
- All US2 tests pass (GREEN phase complete)
- `name_pattern = "default*"` returns entities starting with "default" (case-insensitive)
- `name_pattern = "*node*"` returns entities containing "node"
- Invalid glob patterns silently return no matches

**Checkpoint**: User Story 2 fully functional and independently testable

---

## Phase 5: User Story 3 - Combined Type and Name Filtering (Priority: P2)

**Goal**: Enable combined filtering with AND logic for precise entity lookups

**Independent Test**: Query with both `type = "SoftwareImage"` and `name_pattern = "default*"` returns only SoftwareImages matching the name pattern

### Tests for User Story 3 (TDD RED Phase)

- [X] T023 [P] [US3] Write TestAccCMPartEntityInfoDataSource_CombinedFilters acceptance test in internal/provider/data_source_cmpart_entity_info_test.go - query with type="SoftwareImage" and name_pattern="default*"

**Checkpoint**: All US3 tests should FAIL (RED phase complete)

### Implementation for User Story 3 (TDD GREEN Phase)

- [X] T024 [US3] Implement matchesEntityFilter helper function in internal/provider/data_source_cmpart_entity_info.go combining type and name_pattern checks with AND logic
- [X] T025 [US3] Refactor Read() method in internal/provider/data_source_cmpart_entity_info.go to use matchesEntityFilter for unified filtering

**Acceptance Criteria**:
- All US3 tests pass (GREEN phase complete)
- Combined filters use AND logic (must match both)
- ID includes both filter values when both are set

**Checkpoint**: User Story 3 fully functional and independently testable

---

## Phase 6: User Story 4 - Retrieve All Entities (Priority: P3)

**Goal**: Enable discovery of all BCM entities without filters

**Independent Test**: Query without any filters returns all entities from all types

### Tests for User Story 4 (TDD RED Phase)

- [X] T026 [P] [US4] Write TestAccCMPartEntityInfoDataSource_NoFilters acceptance test in internal/provider/data_source_cmpart_entity_info_test.go - query with no type or name_pattern, verify multiple entity types returned

**Checkpoint**: All US4 tests should FAIL (RED phase complete)

### Implementation for User Story 4 (TDD GREEN Phase)

- [X] T027 [US4] Verify null filter handling in internal/provider/data_source_cmpart_entity_info.go - ensure null/unset filters pass all entities through

**Acceptance Criteria**:
- All US4 tests pass (GREEN phase complete)
- No filters returns all entities from API
- ID is "cmpart-entity-info:all" when no filters

**Checkpoint**: User Story 4 fully functional

---

## Phase 7: User Story 5 - Lookup Entity UUID by Known Name (Priority: P3)

**Goal**: Enable direct UUID lookup for known entities to use in other resources

**Independent Test**: Query with exact type and name returns single entity with correct UUID

### Tests for User Story 5 (TDD RED Phase)

- [X] T028 [P] [US5] Write TestAccCMPartEntityInfoDataSource_UUIDLookup acceptance test in internal/provider/data_source_cmpart_entity_info_test.go - query known entity, verify UUID format (36 chars with dashes)

**Checkpoint**: All US5 tests should FAIL (RED phase complete)

### Implementation for User Story 5 (TDD GREEN Phase)

- [X] T029 [US5] No additional implementation needed - covered by combined filters. Add validation test to confirm single-entity lookup works

**Acceptance Criteria**:
- All US5 tests pass (GREEN phase complete)
- Known entity lookup returns valid UUID
- UUID can be referenced in other resources

**Checkpoint**: User Story 5 fully functional

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, examples, and code quality improvements

- [X] T030 [P] Create example Terraform configuration at examples/data-sources/bcm_cmpart_entity_info/data-source.tf with all filter combinations per quickstart.md
- [X] T031 [P] Add tflog debug/trace logging to Read() method in internal/provider/data_source_cmpart_entity_info.go for API calls and filter operations
- [X] T032 Run make generate to create docs/data-sources/cmpart_entity_info.md
- [X] T033 Run make lint to verify code passes linting checks
- [X] T034 Run full acceptance test suite: TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMPartEntityInfo
- [X] T035 Validate quickstart.md examples work against live BCM cluster

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-7)**: All depend on Foundational phase completion
  - US1 (type filter) provides foundation for US3 (combined)
  - US2 (name filter) provides foundation for US3 (combined)
  - US4 and US5 can proceed independently
- **Polish (Phase 8)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 3 (P2)**: Depends on US1 and US2 (uses both filter types)
- **User Story 4 (P3)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 5 (P3)**: Depends on US3 (uses combined filters for lookup)

### Within Each User Story

- Tests MUST be written and FAIL before implementation (RED phase)
- Implementation makes tests pass (GREEN phase)
- Refactoring improves code quality without breaking tests (REFACTOR phase)

### Parallel Opportunities

- T009-T012 (US1 tests) can run in parallel
- T018-T020 (US2 tests) can run in parallel
- T030-T031 (Polish) can run in parallel
- US1 and US2 can be developed in parallel after Foundational phase

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together (RED phase):
Task T009: "Write TestAccCMPartEntityInfoDataSource_Basic acceptance test"
Task T010: "Write TestAccCMPartEntityInfoDataSource_FilterByType acceptance test"
Task T011: "Write TestAccCMPartEntityInfoDataSource_EmptyResult acceptance test"
Task T012: "Write TestAccCMPartEntityInfoDataSource_InvalidCredentials acceptance test"

# Verify tests fail before proceeding to implementation
TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMPartEntityInfo
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: Foundational (T004-T008)
3. Complete Phase 3: User Story 1 (T009-T017)
4. **STOP and VALIDATE**: Run `TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMPartEntityInfo`
5. Deploy/demo if ready - basic type filtering works!

### Incremental Delivery

1. Complete Setup + Foundational -> Foundation ready
2. Add User Story 1 (type filter) -> Test independently -> MVP!
3. Add User Story 2 (name filter) -> Test independently -> Enhanced filtering
4. Add User Story 3 (combined) -> Test independently -> Full filtering
5. Add User Stories 4-5 -> Complete feature set
6. Each story adds value without breaking previous stories

### TDD Workflow Per Story

```
RED Phase:
  1. Write failing acceptance tests
  2. Run tests - verify they fail
  3. Commit tests

GREEN Phase:
  1. Write minimal implementation
  2. Run tests - verify they pass
  3. Commit implementation

REFACTOR Phase:
  1. Improve code quality (logging, error messages, comments)
  2. Run tests - verify they still pass
  3. Commit refactoring
```

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing (TDD RED phase)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Test environment: `export TF_ACC=1 BCM_ENDPOINT="https://172.21.15.254:8081" BCM_USERNAME="root" BCM_PASSWORD="Hashicorp123!"`

---

## Summary

| Metric | Count |
|--------|-------|
| Total Tasks | 35 |
| Phase 1 (Setup) | 3 tasks |
| Phase 2 (Foundational) | 5 tasks |
| Phase 3 (US1 - Type Filter) | 9 tasks |
| Phase 4 (US2 - Name Filter) | 5 tasks |
| Phase 5 (US3 - Combined) | 3 tasks |
| Phase 6 (US4 - All Entities) | 2 tasks |
| Phase 7 (US5 - UUID Lookup) | 2 tasks |
| Phase 8 (Polish) | 6 tasks |
| Parallel Opportunities | 12 tasks marked [P] |
| MVP Scope | Phases 1-3 (17 tasks) |
