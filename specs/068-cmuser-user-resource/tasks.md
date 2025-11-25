# Tasks: BCM User Resource (bcm_cmuser_user)

**Input**: Design documents from `/workspace/specs/068-cmuser-user-resource/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, quickstart.md
**Branch**: `068-cmuser-user-resource`
**GitHub Issue**: #68

**Tests**: Tests are REQUIRED for this feature - TDD methodology mandates tests first.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4, US5)
- Include exact file paths in descriptions

## Path Conventions

- **Resource implementation**: `internal/provider/resource_cmuser_user.go`
- **Test file**: `internal/provider/resource_cmuser_user_test.go`
- **Examples**: `examples/resources/bcm_cmuser_user/`
- **Documentation**: `docs/resources/cmuser_user.md` (auto-generated)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and API verification

- [ ] T001 Verify API methods exist by testing `cmuser.getUser`, `cmuser.addUser`, `cmuser.updateUser`, `cmuser.removeUser` against live BCM API
- [ ] T002 [P] Document API method signatures and response formats in `/workspace/specs/068-cmuser-user-resource/research.md`
- [ ] T003 [P] Verify user entity structure for Create/Update operations matches `/workspace/specs/068-cmuser-user-resource/data-model.md`
- [ ] T004 Create feature branch `068-cmuser-user-resource` if not exists

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core resource scaffold that MUST be complete before user story tests can be written

**CRITICAL**: No TDD test work can begin until this phase is complete

- [ ] T005 Create resource scaffold in `/workspace/internal/provider/resource_cmuser_user.go` with struct definitions (CMUserUserResource, CMUserUserResourceModel)
- [ ] T006 Implement Schema() method with all attributes (required, optional, computed) per `/workspace/specs/068-cmuser-user-resource/data-model.md`
- [ ] T007 Implement Metadata() method returning resource type name `bcm_cmuser_user`
- [ ] T008 Implement Configure() method to inject BCMClient dependency
- [ ] T009 Add resource interface compile-time checks (var _ resource.Resource = &CMUserUserResource{})
- [ ] T010 Register resource in `/workspace/internal/provider/provider.go` Resources() method
- [ ] T011 [P] Create test file scaffold in `/workspace/internal/provider/resource_cmuser_user_test.go` with testAccCMUserUserConfig helper function
- [ ] T012 [P] Create example directory at `/workspace/examples/resources/bcm_cmuser_user/`

**Checkpoint**: Foundation ready - TDD test writing can now begin

---

## Phase 3: User Story 1 - Create BCM User for Kubernetes Administration (Priority: P1) - MVP

**Goal**: Create BCM user account with username and password for subsequent Kubernetes cluster user setup

**Independent Test**: Create user with minimal attributes, verify exists via `data.bcm_cmuser_users` data source

### RED Phase: Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T013 [US1] Write TestAccCMUserUser_Basic acceptance test (create/read/destroy) in `/workspace/internal/provider/resource_cmuser_user_test.go`
- [ ] T014 [P] [US1] Write TestAccCMUserUser_Complete acceptance test (all attributes) in `/workspace/internal/provider/resource_cmuser_user_test.go`
- [ ] T015 [P] [US1] Write TestAccCMUserUser_PasswordSensitive test (verify password not logged) in `/workspace/internal/provider/resource_cmuser_user_test.go`
- [ ] T016 [US1] Run tests to verify they FAIL (no implementation yet)

### GREEN Phase: Implementation for User Story 1

- [ ] T017 [US1] Implement buildAPIEntity() helper method to construct BCM API entity from Terraform model in `/workspace/internal/provider/resource_cmuser_user.go`
- [ ] T018 [US1] Implement Create() method with addUser API call in `/workspace/internal/provider/resource_cmuser_user.go`
- [ ] T019 [US1] Implement Read() method with getUser(username) direct lookup in `/workspace/internal/provider/resource_cmuser_user.go`
- [ ] T020 [US1] Implement Delete() method with removeUser API call in `/workspace/internal/provider/resource_cmuser_user.go`
- [ ] T021 [US1] Implement password write-only handling (preserve from state, never read from API) in `/workspace/internal/provider/resource_cmuser_user.go`
- [ ] T022 [US1] Run TestAccCMUserUser_Basic - verify it PASSES

### REFACTOR Phase: Production Quality for User Story 1

- [ ] T023 [US1] Add pre-flight validation via validateUser API call before Create in `/workspace/internal/provider/resource_cmuser_user.go`
- [ ] T024 [US1] Add comprehensive error handling for API failures in `/workspace/internal/provider/resource_cmuser_user.go`
- [ ] T025 [P] [US1] Create basic example in `/workspace/examples/resources/bcm_cmuser_user/resource.tf`
- [ ] T026 [US1] Run all US1 tests - verify they PASS

**Checkpoint**: User Story 1 complete - basic user creation works independently

---

## Phase 4: User Story 2 - Update User Attributes (Priority: P2)

**Goal**: Modify user attributes (shell, home_directory, notes, etc.) via Terraform

**Independent Test**: Create user, change shell attribute, verify change applied

### RED Phase: Tests for User Story 2

- [ ] T027 [US2] Write TestAccCMUserUser_Update acceptance test (modify mutable fields) in `/workspace/internal/provider/resource_cmuser_user_test.go`
- [ ] T028 [P] [US2] Write TestAccCMUserUser_Idempotent test (verify no changes on reapply) in `/workspace/internal/provider/resource_cmuser_user_test.go`
- [ ] T029 [US2] Run tests to verify they FAIL

### GREEN Phase: Implementation for User Story 2

- [ ] T030 [US2] Implement Update() method with updateUser API call in `/workspace/internal/provider/resource_cmuser_user.go`
- [ ] T031 [US2] Handle UUID lookup for existing user in Update() in `/workspace/internal/provider/resource_cmuser_user.go`
- [ ] T032 [US2] Implement conditional password update (only send if changed) in `/workspace/internal/provider/resource_cmuser_user.go`
- [ ] T033 [US2] Run TestAccCMUserUser_Update - verify it PASSES

### REFACTOR Phase: Production Quality for User Story 2

- [ ] T034 [US2] Add pre-flight validation via validateUser before Update in `/workspace/internal/provider/resource_cmuser_user.go`
- [ ] T035 [P] [US2] Create update example in `/workspace/examples/resources/bcm_cmuser_user/resource.tf` showing attribute changes
- [ ] T036 [US2] Run all US2 tests - verify they PASS

**Checkpoint**: User Story 2 complete - user attribute updates work independently

---

## Phase 5: User Story 3 - Import Existing User (Priority: P2)

**Goal**: Import existing BCM users into Terraform state using username as import identifier

**Independent Test**: Create user manually, import into Terraform, verify state matches

### RED Phase: Tests for User Story 3

- [ ] T037 [US3] Write TestAccCMUserUser_Import acceptance test (import existing user by username) in `/workspace/internal/provider/resource_cmuser_user_test.go`
- [ ] T038 [P] [US3] Write TestAccCMUserUser_ImportNonExistent test (verify error on non-existent user) in `/workspace/internal/provider/resource_cmuser_user_test.go`
- [ ] T039 [US3] Run tests to verify they FAIL

### GREEN Phase: Implementation for User Story 3

- [ ] T040 [US3] Implement ImportState() method using username as import identifier in `/workspace/internal/provider/resource_cmuser_user.go`
- [ ] T041 [US3] Add ResourceWithImportState interface to resource struct in `/workspace/internal/provider/resource_cmuser_user.go`
- [ ] T042 [US3] Handle import state verification in Read() (populate all fields from API) in `/workspace/internal/provider/resource_cmuser_user.go`
- [ ] T043 [US3] Run TestAccCMUserUser_Import - verify it PASSES

### REFACTOR Phase: Production Quality for User Story 3

- [ ] T044 [P] [US3] Add import example to `/workspace/examples/resources/bcm_cmuser_user/import.sh`
- [ ] T045 [US3] Run all US3 tests - verify they PASS

**Checkpoint**: User Story 3 complete - import functionality works independently

---

## Phase 6: User Story 4 - Delete User (Priority: P2)

**Goal**: Remove users via Terraform destroy

**Independent Test**: Create user, destroy, verify not found in BCM

### RED Phase: Tests for User Story 4

- [ ] T046 [US4] Write testAccCheckCMUserUserDestroy helper function in `/workspace/internal/provider/resource_cmuser_user_test.go`
- [ ] T047 [P] [US4] Enhance TestAccCMUserUser_Basic with CheckDestroy verification in `/workspace/internal/provider/resource_cmuser_user_test.go`
- [ ] T048 [US4] Run tests to verify CheckDestroy works

### GREEN Phase: Implementation for User Story 4

- [ ] T049 [US4] Enhance Delete() with idempotent handling (already deleted = success) in `/workspace/internal/provider/resource_cmuser_user.go`
- [ ] T050 [US4] Add retry logic with exponential backoff for eventual consistency in `/workspace/internal/provider/resource_cmuser_user.go`
- [ ] T051 [US4] Run TestAccCMUserUser_Basic with CheckDestroy - verify it PASSES

### REFACTOR Phase: Production Quality for User Story 4

- [ ] T052 [US4] Add detailed logging for delete operations in `/workspace/internal/provider/resource_cmuser_user.go`
- [ ] T053 [US4] Run all US4 tests - verify they PASS

**Checkpoint**: User Story 4 complete - delete functionality works independently

---

## Phase 7: User Story 5 - Drift Detection (Priority: P3)

**Goal**: Detect external modifications to user attributes and reconcile via Terraform

**Independent Test**: Create user, modify shell via BCM API directly, verify terraform plan detects drift

### RED Phase: Tests for User Story 5

- [ ] T054 [US5] Write TestAccCMUserUser_DriftShell test (external shell modification) in `/workspace/internal/provider/resource_cmuser_user_test.go`
- [ ] T055 [P] [US5] Write TestAccCMUserUser_DriftNotes test (external notes modification) in `/workspace/internal/provider/resource_cmuser_user_test.go`
- [ ] T056 [US5] Run tests to verify they FAIL (drift not detected yet)

### GREEN Phase: Implementation for User Story 5

- [ ] T057 [US5] Verify Read() properly maps all BCM API fields to Terraform model in `/workspace/internal/provider/resource_cmuser_user.go`
- [ ] T058 [US5] Ensure field mapping uses correct camelCase (loginShell, homeDirectory, commonName) in `/workspace/internal/provider/resource_cmuser_user.go`
- [ ] T059 [US5] Run TestAccCMUserUser_DriftShell - verify it PASSES

### REFACTOR Phase: Production Quality for User Story 5

- [ ] T060 [US5] Add drift detection documentation to resource description in `/workspace/internal/provider/resource_cmuser_user.go`
- [ ] T061 [US5] Run all US5 tests - verify they PASS

**Checkpoint**: User Story 5 complete - drift detection works independently

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Documentation generation and final quality improvements

- [ ] T062 [P] Create full example with all attributes in `/workspace/examples/resources/bcm_cmuser_user/resource.tf`
- [ ] T063 Run `make fmt` to format all Go code
- [ ] T064 Run `make lint` to verify code quality
- [ ] T065 Run `make generate` to generate documentation in `/workspace/docs/resources/cmuser_user.md`
- [ ] T066 [P] Update `/workspace/CLAUDE.md` with bcm_cmuser_user field mappings for drift detection tests
- [ ] T067 Run all acceptance tests: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMUserUser`
- [ ] T068 [P] Validate quickstart.md test scenarios work per `/workspace/specs/068-cmuser-user-resource/quickstart.md`
- [ ] T069 Update research.md status to VERIFIED per `/workspace/specs/068-cmuser-user-resource/research.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately (API verification)
- **Foundational (Phase 2)**: Depends on Setup - creates resource scaffold BEFORE tests
- **User Story 1 (Phase 3)**: Depends on Foundational - basic CRUD (MVP)
- **User Story 2 (Phase 4)**: Depends on US1 - Update requires Create/Read
- **User Story 3 (Phase 5)**: Depends on US1 - Import requires Read
- **User Story 4 (Phase 6)**: Depends on US1 - Enhanced Delete verification
- **User Story 5 (Phase 7)**: Depends on US1 - Drift requires complete Read
- **Polish (Phase 8)**: Depends on all user stories complete

### TDD Flow Within Each User Story

1. **RED**: Write failing acceptance tests
2. **Verify RED**: Run tests - confirm they FAIL
3. **GREEN**: Write minimal implementation to pass tests
4. **Verify GREEN**: Run tests - confirm they PASS
5. **REFACTOR**: Improve code quality while keeping tests green

### Parallel Opportunities

**Within Phase 1 (Setup)**:
- T002 and T003 can run in parallel (documentation tasks)

**Within Phase 2 (Foundational)**:
- T011 and T012 can run in parallel (test scaffold and examples)

**Within User Story Phases**:
- RED phase tests marked [P] can run in parallel
- REFACTOR phase documentation tasks marked [P] can run in parallel

**Across User Stories** (with multiple developers):
- Once US1 GREEN phase complete:
  - Developer A: US2 (Update)
  - Developer B: US3 (Import)
  - Developer C: US4 (Delete)
- US5 (Drift) can start after US1 but benefits from complete implementation

---

## Parallel Example: User Story 1 RED Phase

```bash
# After scaffold (T005-T012), launch RED tests in parallel:
Task T014: "TestAccCMUserUser_Complete acceptance test"
Task T015: "TestAccCMUserUser_PasswordSensitive test"

# After T013 TestAccCMUserUser_Basic is written
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (API verification)
2. Complete Phase 2: Foundational (resource scaffold)
3. Complete Phase 3: User Story 1 (basic Create/Read/Delete)
4. **STOP and VALIDATE**: Run `TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMUserUser_Basic`
5. Deploy/demo if ready - users can create basic BCM accounts

### Incremental Delivery

1. Setup + Foundational -> Scaffold ready
2. Add User Story 1 -> MVP: Basic user creation works
3. Add User Story 2 -> Attribute updates work
4. Add User Story 3 -> Import existing users works
5. Add User Story 4 -> Enhanced delete verification
6. Add User Story 5 -> Drift detection complete
7. Polish -> Documentation and final quality

### TDD Verification Commands

```bash
# Run specific user story tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMUserUser_Basic      # US1
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMUserUser_Update     # US2
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMUserUser_Import     # US3
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMUserUser_Drift      # US5

# Run all resource tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMUserUser

# Run with debug logging
TF_LOG=DEBUG TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMUserUser_Basic
```

---

## Task Summary

| Phase | Description | Task Count | Parallel Tasks |
|-------|-------------|------------|----------------|
| Phase 1 | Setup | 4 | 2 |
| Phase 2 | Foundational | 8 | 2 |
| Phase 3 | US1 - Create (MVP) | 14 | 3 |
| Phase 4 | US2 - Update | 10 | 2 |
| Phase 5 | US3 - Import | 9 | 2 |
| Phase 6 | US4 - Delete | 8 | 1 |
| Phase 7 | US5 - Drift | 8 | 1 |
| Phase 8 | Polish | 8 | 3 |
| **Total** | | **69** | **16** |

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label (US1, US2, etc.) maps task to specific user story for traceability
- Each user story is independently completable and testable
- TDD mandates: Write test -> Verify FAIL -> Implement -> Verify PASS
- Commit after each task or logical group
- Password is write-only - BCM API never returns password values
- Import identifier is username (not UUID) for user-friendliness
- Use modern terraform-plugin-testing patterns (statecheck, plancheck) per CLAUDE.md
