# Tasks: BCM Partitions Data Source

**Feature**: `bcm_cmpart_partitions` - Retrieve partition information from BCM cluster
**Branch**: `001-partitions-data-source`
**Input**: Design documents from `/workspace/specs/001-partitions-data-source/`

**Prerequisites**:
- ✅ spec.md (feature specification with user stories)
- ✅ plan.md (implementation plan with TDD workflow)
- ⏳ research.md (API exploration - to be created in Phase 0)
- ⏳ data-model.md (schema design - to be created in Phase 1)

**TDD Workflow**: This feature follows strict RED-GREEN-REFACTOR cycles:
1. **RED**: Write failing acceptance tests FIRST
2. **GREEN**: Write minimal implementation to pass tests
3. **REFACTOR**: Improve code quality while keeping tests green

**Organization**: Tasks organized by user story (P1, P2, P3) to enable independent implementation and testing.

---

## Format: `- [ ] [ID] [P?] [Story?] Description`

- **Checkbox**: `- [ ]` - Required for all tasks
- **[ID]**: Sequential task number (T001, T002, etc.)
- **[P]**: Marks tasks that can run in parallel (different files, no dependencies)
- **[Story]**: User story label (US1, US2, US3) - maps to priorities from spec.md
- **File paths**: All tasks include exact file paths for implementation

---

## Phase 0: Research & API Exploration

**Purpose**: Validate BCM API response structure before implementation

**Gate**: Must confirm API response matches expectations before proceeding to Phase 1

### API Validation

- [ ] T001 Run BCM API exploration script `/workspace/sampleRest/cmpart-get-partitions.py` to retrieve raw partition data
- [ ] T002 Create `/workspace/specs/001-partitions-data-source/research.md` with API response analysis
- [ ] T003 Document field mapping table (camelCase API → snake_case Terraform) in research.md
- [ ] T004 Analyze nested object structures (bmcSettings, provisioningSettings, timeZoneSettings, etc.) and decide flatten vs. nested strategy
- [ ] T005 Identify which fields may be null/missing and require null-safe extraction
- [ ] T006 Verify existing helper functions (getStringValue, getBoolValue, getInt64Value) are sufficient or identify gaps
- [ ] T007 Document client-side filtering strategy (name_pattern with case-insensitive substring match)
- [ ] T008 Update spec.md if API response structure differs significantly from assumptions

**Checkpoint**: API structure validated - proceed to Phase 1 design

---

## Phase 1: Design & Schema Definition

**Purpose**: Define complete schema and data structures before implementation

**Gate**: Re-run Constitution Check after this phase to ensure no architectural complexity introduced

### Schema Design

- [ ] T009 Create `/workspace/specs/001-partitions-data-source/data-model.md` with complete Terraform schema definition
- [ ] T010 [P] Define Go data structures (CMPartPartitionsDataSourceModel, PartitionFilterModel, PartitionModel) in data-model.md
- [ ] T011 [P] Create API contract documentation `/workspace/specs/001-partitions-data-source/contracts/cmpart_getPartitions.json`
- [ ] T012 [P] Generate quickstart guide `/workspace/specs/001-partitions-data-source/quickstart.md` for developers

### Helper Function Planning

- [ ] T013 Determine if getStringListValue() helper is needed for array fields (adminEmail, timeServers, searchDomains, nameServers)
- [ ] T014 Document helper function implementation strategy in data-model.md (inline vs. reusable)

**Checkpoint**: Complete schema design validated against Phase 0 research - ready for TDD implementation

---

## Phase 2: TDD Implementation - RED Phase (Write Failing Tests)

**Purpose**: Define expected behavior through failing acceptance tests BEFORE writing any implementation

**Critical**: All tests MUST fail initially - if a test passes before implementation, it's not testing the right thing

### Test File Structure

- [X] T015 Create test file `/workspace/internal/provider/data_source_cmpart_partitions_test.go` with copyright header and imports
- [X] T016 Add test helper function `testAccCMPartPartitionsDataSourceConfig()` returning provider config + basic data source
- [X] T017 [P] Add test helper function `testAccCMPartPartitionsDataSourceConfigFilter(namePattern string)` for filter tests

### Acceptance Tests (RED - Must Fail)

- [X] T018 [P] [US1] Write `TestAccCMPartPartitionsDataSource_Basic` - verify data source retrieves all partitions and sets computed ID
- [X] T019 [P] [US2] Write `TestAccCMPartPartitionsDataSource_FilterByNamePattern` - verify client-side filtering by name works
- [X] T020 [P] [US2] Write `TestAccCMPartPartitionsDataSource_NoMatches` - verify filter with no matches returns empty list (not error)
- [X] T021 [P] [US1] Write `TestAccCMPartPartitionsDataSource_ComputedFields` - verify all partition attributes exposed with correct types

### Verify Tests Fail

- [X] T022 Run acceptance tests with `TF_ACC=1 go test -v -timeout 30m ./internal/provider/ -run "TestAccCMPartPartitionsDataSource"` and confirm all 4+ tests FAIL with "data source not found" error

**Checkpoint**: RED phase complete - all tests fail as expected (this is GOOD!)

---

## Phase 2: TDD Implementation - GREEN Phase (Minimal Implementation)

**Purpose**: Write minimal code to make tests pass - focus on correctness, not elegance

**Critical**: Implement just enough to make tests GREEN - no extra features, no premature optimization

### Data Source File Structure

- [X] T023 Create implementation file `/workspace/internal/provider/data_source_cmpart_partitions.go` with copyright header and imports
- [X] T024 [P] Define CMPartPartitionsDataSource struct with client field
- [X] T025 [P] Define CMPartPartitionsDataSourceModel, PartitionFilterModel, PartitionModel structs matching Phase 1 design
- [X] T026 [P] Implement NewCMPartPartitionsDataSource() factory function

### Boilerplate Methods

- [X] T027 [P] [US1] Implement Metadata() method returning `req.ProviderTypeName + "_cmpart_partitions"`
- [X] T028 [P] [US1] Implement Configure() method to receive BCMClient from provider
- [X] T029 [US1] Implement Schema() method with complete attribute definitions (30+ fields based on Phase 1 design)

### Minimal Read Implementation

- [X] T030 [US1] Implement Read() method skeleton: read config, call BCM API `cmpart.getPartitions`, set hardcoded ID placeholder
- [X] T031 [US1] Add JSON unmarshaling of API response into `[]map[string]interface{}`
- [X] T032 [US1] Register data source in `/workspace/internal/provider/provider.go` DataSources() method

### Helper Functions (If Needed)

- [X] T033 [P] [US1] Implement getStringListValue() helper function if not already exists (for adminEmail, timeServers arrays)
- [X] T034 [P] [US1] Copy existing helper functions (getStringValue, getBoolValue, getInt64Value) if not shared across files

### Full Data Mapping

- [X] T035 [US1] Implement full partition field mapping in Read() method using helper functions (uuid, name, clusterName, etc.)
- [X] T036 [US1] Map all string fields (clusterName, slaveName, relayHost, notes, primaryHeadNode, etc.)
- [X] T037 [US1] Map all boolean fields (noZeroConf, modified, to_be_removed)
- [X] T038 [US1] Map all int64 fields (slaveDigits, if any)
- [X] T039 [US1] Map all array fields (adminEmail, timeServers, searchDomains, nameServers) using getStringListValue()
- [X] T040 [US1] Handle nested objects based on Phase 1 decisions (flatten or keep nested)

### Client-Side Filtering

- [X] T041 [US2] Implement applyPartitionFilters() function with name_pattern case-insensitive substring matching
- [X] T042 [US2] Integrate filter application in Read() method before setting state

### Verify Tests Pass

- [X] T043 Run acceptance tests with `TF_ACC=1 go test -v -timeout 30m ./internal/provider/ -run "TestAccCMPartPartitionsDataSource"` and confirm all tests PASS

**Checkpoint**: GREEN phase complete - all tests pass with minimal implementation

---

## Phase 2: TDD Implementation - REFACTOR Phase (Improve Quality)

**Purpose**: Improve code quality while keeping tests GREEN - refactor for readability, maintainability, performance

**Critical**: Run tests after each refactoring to ensure they stay GREEN

### Code Quality Improvements

- [X] T044 [P] Extract magic strings to constants (bcmPartitionService = "cmpart", bcmPartitionMethod = "getPartitions")
- [X] T045 [P] Add comprehensive debug logging with tflog.Debug() at key points (API call, filtering, response parsing)
- [X] T046 [P] Improve error messages with context-specific details (service name, method name)
- [X] T047 [P] Add field mapping comments documenting camelCase → snake_case conversions

### Performance Optimization

- [X] T048 Optimize filter performance: pre-lowercase filter strings once, short-circuit empty filters

### Final Validation

- [X] T049 Run `make fmt` to format code
- [X] T050 Run `make lint` to check code quality
- [X] T051 Run `TF_ACC=1 go test -v -timeout 30m ./internal/provider/ -run "TestAccCMPartPartitionsDataSource"` to verify tests still pass
- [X] T052 Run `make test` to ensure no regressions in unit tests

**Checkpoint**: REFACTOR phase complete - code is clean, tests are GREEN

---

## Phase 3: Documentation & Examples

**Purpose**: Create working examples and auto-generate documentation

**Gate**: All acceptance tests must be GREEN before proceeding to documentation

### Example Configuration

- [X] T053 Create directory `/workspace/examples/data-sources/bcm_cmpart_partitions/`
- [X] T054 [US1] Create `/workspace/examples/data-sources/bcm_cmpart_partitions/data-source.tf` with basic retrieval example
- [X] T055 [US2] Add filter by name_pattern example to data-source.tf
- [X] T056 [US3] Add example using partition UUID in software image resource reference
- [X] T057 [P] Add output blocks showing partition names, UUIDs, and detailed information

### Documentation Generation

- [X] T058 Run `make generate` to auto-generate documentation via tfplugindocs
- [X] T059 Verify `/workspace/docs/data-sources/cmpart_partitions.md` was created and is complete
- [X] T060 Review generated documentation for accuracy (schema table, attribute types, descriptions)

### Example Validation

- [X] T061 Test example configuration against real BCM cluster: `cd examples/data-sources/bcm_cmpart_partitions && terraform init && terraform validate`
- [X] T062 Run `terraform plan` on example to verify no errors
- [X] T063 Optionally run `terraform apply` to verify outputs display partition information

**Checkpoint**: Documentation complete and validated

---

## Phase 4: Final Validation & Commit

**Purpose**: Comprehensive validation and feature branch commit

**Gate**: All quality checks must pass before committing

### Full Test Suite

- [X] T064 Run complete acceptance test suite: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/`
- [X] T065 Verify no regressions in existing tests (software images, nodes, networks data sources)

### Code Quality Final Checks

- [X] T066 Run `make fmt` - verify no formatting changes needed
- [X] T067 Run `make lint` - verify no linting errors
- [X] T068 Run `make test` - verify all unit tests pass

### Specification Compliance Review

- [X] T069 Verify FR-001: Data source calls `cmpart.getPartitions` API ✓
- [X] T070 Verify FR-002: All partition attributes exposed ✓
- [X] T071 Verify FR-003: Supports name_pattern filter ✓
- [X] T072 Verify FR-004: Handles null fields gracefully ✓
- [X] T073 Verify FR-009: Uses modern testing patterns (statecheck.ExpectKnownValue) ✓
- [X] T074 Verify FR-010: Tests are environment-portable ✓

### Git Commit

- [X] T075 Stage all implementation files for commit
- [X] T076 Commit with message: "feat: add bcm_cmpart_partitions data source\n\nImplement read-only data source for retrieving partition information from BCM cluster. Supports client-side filtering by name pattern.\n\n- Add data source implementation with 30+ computed attributes\n- Add 4 acceptance tests using modern terraform-plugin-testing patterns\n- Add example configuration and auto-generated documentation\n- Follow TDD RED-GREEN-REFACTOR workflow\n\nCloses: 001-partitions-data-source\n\n🤖 Generated with [Claude Code](https://claude.com/claude-code)\n\nCo-Authored-By: Claude <noreply@anthropic.com>"

**Checkpoint**: Feature complete and committed to branch

---

## Dependencies & Execution Order

### Phase Dependencies (Must Run Sequentially)

1. **Phase 0 (Research)** → Must complete before Phase 1
2. **Phase 1 (Design)** → Must complete before Phase 2
3. **Phase 2 RED** → Must complete before Phase 2 GREEN
4. **Phase 2 GREEN** → Must complete before Phase 2 REFACTOR
5. **Phase 2 REFACTOR** → Must complete before Phase 3
6. **Phase 3 (Documentation)** → Must complete before Phase 4

### User Story Dependencies

- **User Story 1 (P1)**: Query All Partitions - NO dependencies, foundational capability
- **User Story 2 (P2)**: Filter by Name Pattern - Depends on US1 (builds on basic retrieval)
- **User Story 3 (P3)**: Reference UUIDs in Resources - Depends on US1 and US2 (uses filtered partition lookups)

### Task-Level Parallel Opportunities

**Phase 0 - Research (can run some tasks in parallel)**:
- T003, T004, T005, T006, T007 (analysis tasks) can run concurrently after T001-T002 complete

**Phase 1 - Design (can run in parallel)**:
- T010, T011, T012 can all run in parallel after T009 completes

**Phase 2 RED - Test Writing (HIGH parallelism)**:
- T018, T019, T020, T021 can ALL run in parallel (4 concurrent test writes)

**Phase 2 GREEN - Implementation (some parallelism)**:
- T024, T025, T026 can run in parallel (struct definitions)
- T027, T028 can run in parallel (boilerplate methods)
- T033, T034 can run in parallel (helper functions)
- T036, T037, T038, T039 can run in parallel (field mapping groups)

**Phase 2 REFACTOR - Quality (HIGH parallelism)**:
- T044, T045, T046, T047 can ALL run in parallel (independent improvements)

**Phase 3 - Documentation (some parallelism)**:
- T054, T055, T056, T057 can run in parallel (example sections)

**Optimal Concurrent Execution**: 5 agents in parallel for test writing (RED) and refactoring phases

---

## Parallel Execution Examples

### Phase 2 RED - Write All Tests in Parallel (4 concurrent agents)

```bash
# Agent 1: Basic test
Task T018: "Write TestAccCMPartPartitionsDataSource_Basic in data_source_cmpart_partitions_test.go"

# Agent 2: Filter test
Task T019: "Write TestAccCMPartPartitionsDataSource_FilterByNamePattern in data_source_cmpart_partitions_test.go"

# Agent 3: No matches test
Task T020: "Write TestAccCMPartPartitionsDataSource_NoMatches in data_source_cmpart_partitions_test.go"

# Agent 4: Computed fields test
Task T021: "Write TestAccCMPartPartitionsDataSource_ComputedFields in data_source_cmpart_partitions_test.go"
```

### Phase 2 GREEN - Parallel Struct Definitions (3 concurrent agents)

```bash
# Agent 1: Data source struct
Task T024: "Define CMPartPartitionsDataSource struct in data_source_cmpart_partitions.go"

# Agent 2: Model structs
Task T025: "Define CMPartPartitionsDataSourceModel, PartitionFilterModel, PartitionModel in data_source_cmpart_partitions.go"

# Agent 3: Factory function
Task T026: "Implement NewCMPartPartitionsDataSource() in data_source_cmpart_partitions.go"
```

### Phase 2 REFACTOR - All Quality Improvements in Parallel (4 concurrent agents)

```bash
# Agent 1: Constants
Task T044: "Extract magic strings to constants in data_source_cmpart_partitions.go"

# Agent 2: Logging
Task T045: "Add comprehensive debug logging in data_source_cmpart_partitions.go"

# Agent 3: Error messages
Task T046: "Improve error messages with context in data_source_cmpart_partitions.go"

# Agent 4: Comments
Task T047: "Add field mapping comments in data_source_cmpart_partitions.go"
```

---

## Implementation Strategy

### MVP-First Approach (User Story 1 Only)

1. Complete Phase 0: Research (validate API)
2. Complete Phase 1: Design (schema definition)
3. Complete Phase 2 RED: Write basic retrieval test (T018, T021)
4. Complete Phase 2 GREEN: Implement basic retrieval (T023-T040, skip T041-T042)
5. Complete Phase 2 REFACTOR: Code quality (T044-T052)
6. Complete Phase 3: Documentation (basic example only)
7. **STOP and VALIDATE**: Test data source retrieves all partitions
8. **MVP COMPLETE**: Users can now discover partitions in their cluster

### Incremental Delivery (P1 → P2 → P3)

1. **MVP (US1)**: Basic partition retrieval - delivers partition discovery capability
2. **Enhancement (US2)**: Add filtering - delivers improved usability for large partition sets
3. **Integration (US3)**: UUID references - delivers practical workflow for software image configuration

### Full Feature Implementation (All User Stories)

1. Phase 0: Research (8 tasks, ~2-3 hours)
2. Phase 1: Design (6 tasks, ~2-3 hours)
3. Phase 2 RED: Tests (8 tasks, ~1 hour with parallel execution)
4. Phase 2 GREEN: Implementation (21 tasks, ~3-4 hours)
5. Phase 2 REFACTOR: Quality (9 tasks, ~1 hour with parallel execution)
6. Phase 3: Documentation (11 tasks, ~1-2 hours)
7. Phase 4: Validation (13 tasks, ~1-2 hours)

**Total Estimated Time**: 11-16 hours (realistic with some parallelism)

---

## Success Criteria Validation

After completing all tasks, verify:

- ✅ **SC-001**: Data source retrieves all partitions in single declaration
- ✅ **SC-002**: Filtered queries complete in <5 seconds for 100 partitions
- ✅ **SC-003**: Handles empty partition lists without errors
- ✅ **SC-004**: All acceptance tests pass with modern patterns
- ✅ **SC-005**: Generated documentation is clear and complete
- ✅ **SC-006**: Partition UUIDs usable in software image resources

---

## Deliverables Checklist

- [ ] `/workspace/specs/001-partitions-data-source/research.md` - API exploration results
- [ ] `/workspace/specs/001-partitions-data-source/data-model.md` - Schema and struct definitions
- [ ] `/workspace/specs/001-partitions-data-source/quickstart.md` - Developer guide
- [ ] `/workspace/specs/001-partitions-data-source/contracts/cmpart_getPartitions.json` - API contract
- [ ] `/workspace/internal/provider/data_source_cmpart_partitions.go` - Implementation (~500-600 lines)
- [ ] `/workspace/internal/provider/data_source_cmpart_partitions_test.go` - Acceptance tests (~200-300 lines)
- [ ] `/workspace/examples/data-sources/bcm_cmpart_partitions/data-source.tf` - Example configuration
- [ ] `/workspace/docs/data-sources/cmpart_partitions.md` - Auto-generated documentation
- [ ] Modified `/workspace/internal/provider/provider.go` - Data source registration

---

## Notes for Implementation

1. **TDD Discipline**: Never skip RED phase - tests must fail before implementation
2. **Environment Portability**: No hardcoded partition names, counts, or UUIDs in tests
3. **Modern Testing**: Use `statecheck.ExpectKnownValue()` with type-safe matchers (knownvalue.StringExact, NotNull, etc.)
4. **Null Safety**: Use helper functions (getStringValue, getBoolValue, getInt64Value) for all field extraction
5. **Client-Side Filtering**: BCM API does not support server-side filters - all filtering in Go code
6. **Parallel Execution**: Mark tasks with [P] for concurrent execution with `/speckit.implement` (5 agents)
7. **Reference Pattern**: Follow `data_source_cmpart_softwareimages.go` exactly for consistency
8. **Commit Early**: Commit after GREEN phase passes, commit again after REFACTOR phase
9. **Documentation**: Never edit `/workspace/docs/` manually - always use `make generate`
10. **API Contract**: Exact field names and types from Phase 0 research - update spec.md if deviations found

---

## Risk Mitigation

### Technical Risks

| Risk | Mitigation | Tasks |
|------|------------|-------|
| API response structure differs from spec | Phase 0 validates API before implementation | T001-T008 |
| Complex nested objects require extensive mapping | Phase 1 decides flatten vs. nested strategy | T004, T009, T040 |
| Null fields cause panics | Use null-safe helpers for all extractions | T006, T034-T039 |
| Filter performance degrades with 100+ partitions | REFACTOR phase optimizes filtering | T048 |
| Tests fail in different environments | Environment-portable test design (no hardcoded values) | T018-T021 |

### Process Risks

| Risk | Mitigation | Tasks |
|------|------------|-------|
| Skipping RED phase (tests not first) | Plan enforces RED before GREEN explicitly | T015-T022 before T023+ |
| Incomplete API research | Phase 0 required before Phase 1 starts | T001-T008 gate |
| Tests pass before implementation | T022 verifies all tests fail initially | T022 checkpoint |
| Premature optimization | GREEN focuses on correctness, REFACTOR on quality | Separate phases |

---

**Ready for execution with `/speckit.implement` or manual task-by-task implementation following TDD principles.**
