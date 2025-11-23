# Tasks: BCM CMUser Users Data Source

**Branch**: `002-cmuser-users-datasource`
**Input**: Design documents from `/workspace/specs/002-cmuser-users-datasource/`
**Prerequisites**: plan.md, spec.md (user stories), data-model.md, contracts/

**Organization**: Tasks follow TDD RED-GREEN-REFACTOR cycle with modern terraform-plugin-testing patterns. Tasks are grouped by implementation phase with clear validation checkpoints.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Project Structure)

**Purpose**: Prepare development environment and branch

- [ ] T001 Create feature branch `002-cmuser-users-datasource` from main
- [ ] T002 [P] Verify BCM cluster access at https://172.21.15.254:8081
- [ ] T003 [P] Verify test environment variables (BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD)

**Validation**: Branch exists, BCM cluster accessible, environment ready

---

## Phase 2: Foundational (API Research & Validation)

**Purpose**: Validate BCM CMUser API structure before implementation

**⚠️ CRITICAL**: Must complete before any data source implementation

- [ ] T004 Create BCM API exploration script `sampleRest/explore_cmuser_users.py`
- [ ] T005 Run exploration script to validate CMUser getUsers API response structure
- [ ] T006 Document field name mappings in `/workspace/specs/002-cmuser-users-datasource/research.md` (BCM `name` → TF `username`, BCM `ID` → TF `user_id`, BCM `groupID` → TF `group_id`)
- [ ] T007 Validate shadow password field types (shadowExpire, shadowLastChange, shadowMax, shadowMin, shadowWarning, shadowInactive)
- [ ] T008 Document epoch day calculation formula for account_active computation in research.md
- [ ] T009 Test username pattern matching strategy (filepath.Match vs strings.Contains) in research.md

**Validation**: Research.md complete with confirmed API structure, field mappings, and computation formulas

**Checkpoint**: Foundation ready - TDD implementation can now begin

---

## Phase 3: TDD RED Phase - Write Failing Tests (User Story 1)

**Goal**: Query all users without filters (P1 - MVP)

**Independent Test**: Data source retrieves all users and populates Unix attributes

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T010 Create test file `internal/provider/data_source_cmuser_users_test.go`
- [ ] T011 [P] [US1] Write TestAccCMUserUsersDataSource_Basic test - verify data source retrieves users without errors
- [ ] T012 [P] [US1] Write test helper function testAccCMUserUsersDataSourceConfigBasic() with provider config
- [ ] T013 [P] [US1] Add statecheck.ExpectKnownValue for id and users list (NotNull validation)
- [ ] T014 [US1] Run TF_ACC=1 go test to verify test FAILS (no implementation exists)

**Validation**: Test T011 compiles and FAILS with "unknown data source" error

---

## Phase 4: TDD RED Phase - Additional Test Scenarios (User Stories 2-4)

**Goal**: Write all failing tests for filtering and Unix attributes

### Tests for User Story 2 - Filter by Username Pattern (P2)

- [ ] T015 [P] [US2] Write TestAccCMUserUsersDataSource_FilterUsername test
- [ ] T016 [P] [US2] Write test helper testAccCMUserUsersDataSourceConfigFilterUsername(pattern string)
- [ ] T017 [P] [US2] Add statecheck validation for filtered results

### Tests for User Story 3 - Filter by Group ID (P3)

- [ ] T018 [P] [US3] Write TestAccCMUserUsersDataSource_FilterGroupID test
- [ ] T019 [P] [US3] Write test helper testAccCMUserUsersDataSourceConfigFilterGroupID(groupID string)
- [ ] T020 [P] [US3] Add statecheck validation for group filtering

### Tests for User Story 4 - Filter by User ID (P3)

- [ ] T021 [P] [US4] Write TestAccCMUserUsersDataSource_FilterUserID test
- [ ] T022 [P] [US4] Write test helper testAccCMUserUsersDataSourceConfigFilterUserID(userID string)
- [ ] T023 [P] [US4] Add statecheck validation for user ID filtering

### Tests for Nested Attributes Validation

- [ ] T024 [P] Write TestAccCMUserUsersDataSource_NestedAttributes test
- [ ] T025 [P] Add statecheck validations for Unix attributes (uuid, username, user_id, group_id, home_directory, login_shell)

### Tests for Account Active Computation

- [ ] T026 [P] Write TestAccCMUserUsersDataSource_AccountActive test
- [ ] T027 [P] Add statecheck validation for account_active computed field

### Validation Checkpoint

- [ ] T028 Run all tests with TF_ACC=1 go test -run TestAccCMUserUsersDataSource to verify ALL tests FAIL
- [ ] T029 Verify test output shows "unknown data source bcm_cmuser_users" errors

**Validation**: All 6 test scenarios compile and FAIL with expected errors

**Checkpoint**: RED phase complete - all tests written and failing

---

## Phase 5: TDD GREEN Phase - Minimal Implementation

**Goal**: Create minimal data source that makes tests pass (hardcoded empty list)

### Data Source Skeleton

- [ ] T030 Create data source file `internal/provider/data_source_cmuser_users.go`
- [ ] T031 [P] Define CMUserUsersDataSource struct with client field
- [ ] T032 [P] Define CMUserUsersDataSourceModel struct with ID, filters (UsernamePattern, GroupID, UserID), and Users list
- [ ] T033 [P] Define UserModel struct with all 22 Unix attributes (UUID, Username, UserID, GroupID, Email, CommonName, Surname, HomeDirectory, LoginShell, Notes, Information, AuthorizedSSHKeys, Shadow fields, AccountActive)

### Data Source Methods

- [ ] T034 [P] Implement Metadata() method - set TypeName to "bcm_cmuser_users"
- [ ] T035 Implement Schema() method - define all filter attributes (optional) and users list (computed)
- [ ] T036 [P] Add schema definitions for 22 user attributes with MarkdownDescription
- [ ] T037 [P] Implement Configure() method - receive BCMClient from provider
- [ ] T038 Implement minimal Read() method - return hardcoded empty users list and set ID

### Provider Registration

- [ ] T039 Register NewCMUserUsersDataSource in `internal/provider/provider.go` DataSources() method

### Validation Checkpoint

- [ ] T040 Run TF_ACC=1 go test -run TestAccCMUserUsersDataSource_Basic to verify test now RUNS but fails validation checks
- [ ] T041 Verify test error changed from "unknown data source" to "expected users list to be populated"

**Validation**: Tests run but fail validation (empty users list)

**Checkpoint**: GREEN phase started - data source exists but not functional

---

## Phase 6: TDD REFACTOR Phase - Full Implementation

**Goal**: Complete BCM API integration with filtering and null-safe field mapping

### Helper Functions

- [ ] T042 [P] Implement computeAccountActive(shadowExpire types.Int64) helper function with epoch day calculation
- [ ] T043 [P] Implement matchesFilters(user UserModel, filters) helper function for client-side filtering
- [ ] T044 [P] Implement mapUserAPIResponseToModel(userData map[string]interface{}) helper function using null-safe helpers

### Full Read Implementation

- [ ] T045 Update Read() method - call client.CallJSONRPC(ctx, "cmuser", "getUsers")
- [ ] T046 [P] Add error handling for BCM API call failures
- [ ] T047 [P] Add JSON unmarshaling with error handling
- [ ] T048 Implement client-side filtering loop - apply username_pattern, group_id, user_id filters
- [ ] T049 [P] Map BCM API fields to UserModel using helper functions (name→username, ID→user_id, groupID→group_id)
- [ ] T050 [P] Compute account_active from shadow_expire for each user
- [ ] T051 Set filteredUsers to data.Users and generate deterministic ID

### Validation Checkpoint

- [ ] T052 Run TF_ACC=1 go test -run TestAccCMUserUsersDataSource_Basic to verify basic test PASSES
- [ ] T053 Run TF_ACC=1 go test -run TestAccCMUserUsersDataSource_FilterUsername to verify username filter test PASSES
- [ ] T054 Run TF_ACC=1 go test -run TestAccCMUserUsersDataSource_FilterGroupID to verify group_id filter test PASSES
- [ ] T055 Run TF_ACC=1 go test -run TestAccCMUserUsersDataSource_FilterUserID to verify user_id filter test PASSES
- [ ] T056 Run TF_ACC=1 go test -run TestAccCMUserUsersDataSource_NestedAttributes to verify Unix attributes test PASSES
- [ ] T057 Run TF_ACC=1 go test -run TestAccCMUserUsersDataSource_AccountActive to verify account_active computation test PASSES
- [ ] T058 Run TF_ACC=1 go test -run TestAccCMUserUsersDataSource to verify ALL 6 tests PASS

**Validation**: All acceptance tests pass with 100% success rate

**Checkpoint**: REFACTOR phase complete - full implementation working

---

## Phase 7: Examples & Documentation

**Goal**: Create example configurations and auto-generate documentation

### Example Configurations

- [ ] T059 Create directory `examples/data-sources/bcm_cmuser_users/`
- [ ] T060 [P] Create `examples/data-sources/bcm_cmuser_users/basic.tf` - query all users example
- [ ] T061 [P] Create `examples/data-sources/bcm_cmuser_users/filter_username.tf` - username pattern filter example
- [ ] T062 [P] Create `examples/data-sources/bcm_cmuser_users/filter_group.tf` - group_id filter example
- [ ] T063 [P] Create `examples/data-sources/bcm_cmuser_users/filter_user.tf` - user_id filter example
- [ ] T064 [P] Create `examples/data-sources/bcm_cmuser_users/combined_filters.tf` - multiple filters example

### Documentation Generation

- [ ] T065 Run `make generate` to auto-generate documentation from schema and examples
- [ ] T066 Verify `docs/data-sources/cmuser_users.md` created with correct schema documentation
- [ ] T067 Verify examples formatted correctly with terraform fmt
- [ ] T068 Verify copyright headers added via copywrite

**Validation**: Documentation exists and accurately reflects implementation

---

## Phase 8: Integration Testing & Validation

**Goal**: Manual testing and final validation

### Code Quality

- [ ] T069 [P] Run `make fmt` to format Go code
- [ ] T070 [P] Run `make lint` to check for linting issues
- [ ] T071 [P] Run `pre-commit run --all-files` to validate all pre-commit hooks

### Manual Testing Scenarios

- [ ] T072 Test Scenario 1 - Query all users: cd examples/data-sources/bcm_cmuser_users && terraform init && terraform apply
- [ ] T073 Test Scenario 2 - Filter by username: terraform apply -target=data.bcm_cmuser_users.admins
- [ ] T074 Test Scenario 3 - Combined filters: terraform apply -target=data.bcm_cmuser_users.filtered
- [ ] T075 Test Scenario 4 - Verify account_active computation: terraform console → data.bcm_cmuser_users.all.users[*].account_active

### Success Criteria Validation

- [ ] T076 Verify SC-001: Operators can retrieve all BCM users without errors
- [ ] T077 Verify SC-002: Username pattern filtering returns only matching results
- [ ] T078 Verify SC-003: Group ID filtering returns only users in specified group
- [ ] T079 Verify SC-004: User ID filtering returns specific user by UID
- [ ] T080 Verify SC-005: BCM API errors provide clear error messages
- [ ] T081 Verify SC-006: Client-side filtering uses single API call
- [ ] T082 Verify SC-007: All Unix attributes accurately mapped from BCM API
- [ ] T083 Verify SC-008: Account active status correctly computed from shadow_expire
- [ ] T084 Verify SC-009: Tests work on any BCM cluster (no hardcoded assumptions)
- [ ] T085 Verify SC-010: All 6 acceptance tests pass with 100% success rate
- [ ] T086 Verify SC-011: Documentation auto-generated successfully

**Validation**: All success criteria met, manual tests pass

---

## Phase 9: Polish & Final Validation

**Goal**: Final cleanup and validation before PR

- [ ] T087 Review all code for consistency with existing data source patterns
- [ ] T088 Verify null-safe helper functions used correctly (getStringValue, getInt64Value)
- [ ] T089 Verify field name mappings correct (name→username, ID→user_id, groupID→group_id)
- [ ] T090 Verify account_active computation logic handles shadowExpire=-1 and positive values
- [ ] T091 Run quickstart.md validation steps from `/workspace/specs/002-cmuser-users-datasource/quickstart.md`
- [ ] T092 Final acceptance test run: TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMUserUsersDataSource

**Validation**: All tests pass, code quality checks pass, documentation complete

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all TDD implementation
- **TDD RED (Phases 3-4)**: Depends on Foundational - Write all failing tests
- **TDD GREEN (Phase 5)**: Depends on RED phase - Minimal implementation
- **TDD REFACTOR (Phase 6)**: Depends on GREEN phase - Full implementation
- **Examples & Docs (Phase 7)**: Depends on REFACTOR phase - Tests passing
- **Integration Testing (Phase 8)**: Depends on Examples - Manual validation
- **Polish (Phase 9)**: Depends on Integration - Final cleanup

### User Story Dependencies

- **User Story 1 (P1)**: Query all users - No dependencies on other stories (MVP)
- **User Story 2 (P2)**: Filter by username - Independent, builds on US1 infrastructure
- **User Story 3 (P3)**: Filter by group_id - Independent, builds on US1 infrastructure
- **User Story 4 (P3)**: Filter by user_id - Independent, builds on US1 infrastructure

### Within Each Phase

**RED Phase**:
- All test files can be written in parallel (T015-T027 marked [P])
- Tests must FAIL before proceeding to GREEN phase

**GREEN Phase**:
- Struct definitions can be created in parallel (T031-T033 marked [P])
- Methods can be created in parallel (T034, T036-T037 marked [P])
- Schema must be complete before Read implementation

**REFACTOR Phase**:
- Helper functions can be created in parallel (T042-T044 marked [P])
- Read implementation tasks are sequential (T045-T051)
- All validation tests run in parallel (T052-T057)

**Examples Phase**:
- All example files can be created in parallel (T060-T064 marked [P])

**Polish Phase**:
- Code quality tasks can run in parallel (T069-T071 marked [P])

### Parallel Opportunities

**Phase 2 (Foundational)**: T004-T009 can run in sequence (API exploration)

**Phase 3 (RED - US1)**: T011-T013 can run in parallel (test writing)

**Phase 4 (RED - US2-4)**: All test scenarios can run in parallel
- T015-T017 (US2 tests)
- T018-T020 (US3 tests)
- T021-T023 (US4 tests)
- T024-T025 (nested attributes tests)
- T026-T027 (account_active tests)

**Phase 5 (GREEN)**: Struct definitions and methods can run in parallel
- T031-T033 (model structs)
- T034, T036-T037 (data source methods)

**Phase 6 (REFACTOR)**: Helper functions can run in parallel
- T042-T044 (helpers)
- T046-T047, T049-T050 (implementation tasks)

**Phase 7 (Examples)**: All example files can run in parallel
- T060-T064 (5 example files)

**Phase 8 (Integration)**: Code quality checks can run in parallel
- T069-T071 (fmt, lint, pre-commit)

---

## Parallel Example: RED Phase

```bash
# Launch all test scenarios in parallel (Phase 4):
Task T015: "Write TestAccCMUserUsersDataSource_FilterUsername test"
Task T018: "Write TestAccCMUserUsersDataSource_FilterGroupID test"
Task T021: "Write TestAccCMUserUsersDataSource_FilterUserID test"
Task T024: "Write TestAccCMUserUsersDataSource_NestedAttributes test"
Task T026: "Write TestAccCMUserUsersDataSource_AccountActive test"
```

---

## Parallel Example: REFACTOR Phase

```bash
# Launch all helper functions in parallel (Phase 6):
Task T042: "Implement computeAccountActive helper function"
Task T043: "Implement matchesFilters helper function"
Task T044: "Implement mapUserAPIResponseToModel helper function"
```

---

## Implementation Strategy

### TDD RED-GREEN-REFACTOR Approach

1. **Complete Phase 1**: Setup → Branch ready
2. **Complete Phase 2**: Foundational → API validated
3. **Complete Phase 3**: RED (US1) → Basic test fails
4. **Complete Phase 4**: RED (US2-4) → All tests fail
5. **VALIDATE RED**: All 6 tests compile and fail with expected errors
6. **Complete Phase 5**: GREEN → Minimal implementation, tests run but fail validation
7. **Complete Phase 6**: REFACTOR → Full implementation, all tests PASS
8. **VALIDATE REFACTOR**: All 6 acceptance tests pass with 100% success rate
9. **Complete Phase 7**: Examples & Docs → Documentation generated
10. **Complete Phase 8**: Integration Testing → Manual validation
11. **Complete Phase 9**: Polish → Final cleanup

### MVP First (User Story 1 Only)

1. Complete Phase 1-2: Setup + Foundational
2. Complete Phase 3: RED (US1 only)
3. Complete Phase 5-6: GREEN + REFACTOR (basic query only, no filters)
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Add remaining user stories (filters) in priority order

### Incremental Delivery

Each checkpoint represents a deployable increment:
1. **Checkpoint 1**: Foundation ready (API validated)
2. **Checkpoint 2**: RED phase complete (all tests written and failing)
3. **Checkpoint 3**: GREEN phase complete (skeleton implementation)
4. **Checkpoint 4**: REFACTOR phase complete (full implementation, all tests pass) → **MVP READY**
5. **Checkpoint 5**: Examples & Docs complete → **PRODUCTION READY**
6. **Checkpoint 6**: Integration testing complete → **VALIDATED**
7. **Checkpoint 7**: Polish complete → **READY FOR PR**

---

## Notes

- **TDD Discipline**: NEVER implement before tests are written and failing
- **Modern Patterns**: Use statecheck.ExpectKnownValue with knownvalue matchers (StringExact, Bool, Int64Exact, NotNull)
- **Null Safety**: Use getStringValue, getInt64Value helpers for all BCM API field extraction
- **Field Mapping**: BCM `name` → TF `username`, BCM `ID` → TF `user_id`, BCM `groupID` → TF `group_id`
- **Account Active**: Compute from shadow_expire (active = shadowExpire == -1 OR shadowExpire > currentEpochDay)
- **Client-Side Filtering**: All filtering done in Go after single BCM API call
- **Environment Portability**: No hardcoded user counts, usernames, or assumptions about cluster state
- **[P] tasks**: Different files, no dependencies, can run in parallel
- **[Story] label**: Maps task to specific user story for traceability
- **Commit Strategy**: Commit after each phase checkpoint or logical group
- **Stop Points**: Each checkpoint is a validation opportunity - stop and verify before proceeding
